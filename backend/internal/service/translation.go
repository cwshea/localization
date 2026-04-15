package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/cwshea/localization/internal/llm"
	"github.com/cwshea/localization/internal/models"
	"golang.org/x/sync/errgroup"
)

type TranslationService struct {
	db  *pgxpool.Pool
	llm *llm.ClientFactory
}

func NewTranslationService(db *pgxpool.Pool, llmFactory *llm.ClientFactory) *TranslationService {
	return &TranslationService{db: db, llm: llmFactory}
}

func (s *TranslationService) ListSourceStrings(ctx context.Context) ([]models.SourceString, error) {
	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.text, s.created_at, s.updated_at,
		       t.id, t.locale, t.translated_text, t.llm_provider, t.translated_at, t.updated_at
		FROM source_strings s
		LEFT JOIN translations t ON t.source_id = s.id
		ORDER BY s.created_at DESC, t.locale, t.llm_provider
	`)
	if err != nil {
		return nil, fmt.Errorf("querying source strings: %w", err)
	}
	defer rows.Close()

	sourceMap := make(map[uuid.UUID]*models.SourceString)
	var order []uuid.UUID

	for rows.Next() {
		var s models.SourceString
		var tID *uuid.UUID
		var tLocale, tText, tProvider *string
		var tTranslatedAt, tUpdatedAt *time.Time

		err := rows.Scan(
			&s.ID, &s.Text, &s.CreatedAt, &s.UpdatedAt,
			&tID, &tLocale, &tText, &tProvider, &tTranslatedAt, &tUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		if _, exists := sourceMap[s.ID]; !exists {
			s.Translations = []models.Translation{}
			sourceMap[s.ID] = &s
			order = append(order, s.ID)
		}

		if tID != nil {
			sourceMap[s.ID].Translations = append(sourceMap[s.ID].Translations, models.Translation{
				ID:             *tID,
				SourceID:       s.ID,
				Locale:         *tLocale,
				TranslatedText: *tText,
				LLMProvider:    *tProvider,
				TranslatedAt:   *tTranslatedAt,
				UpdatedAt:      *tUpdatedAt,
			})
		}
	}

	result := make([]models.SourceString, 0, len(order))
	for _, id := range order {
		result = append(result, *sourceMap[id])
	}
	return result, nil
}

func (s *TranslationService) GetSourceString(ctx context.Context, id uuid.UUID) (*models.SourceString, error) {
	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.text, s.created_at, s.updated_at,
		       t.id, t.locale, t.translated_text, t.llm_provider, t.translated_at, t.updated_at
		FROM source_strings s
		LEFT JOIN translations t ON t.source_id = s.id
		WHERE s.id = $1
		ORDER BY t.locale, t.llm_provider
	`, id)
	if err != nil {
		return nil, fmt.Errorf("querying source string: %w", err)
	}
	defer rows.Close()

	var source *models.SourceString
	for rows.Next() {
		var ss models.SourceString
		var tID *uuid.UUID
		var tLocale, tText, tProvider *string
		var tTranslatedAt, tUpdatedAt *time.Time

		err := rows.Scan(
			&ss.ID, &ss.Text, &ss.CreatedAt, &ss.UpdatedAt,
			&tID, &tLocale, &tText, &tProvider, &tTranslatedAt, &tUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		if source == nil {
			ss.Translations = []models.Translation{}
			source = &ss
		}

		if tID != nil {
			source.Translations = append(source.Translations, models.Translation{
				ID:             *tID,
				SourceID:       ss.ID,
				Locale:         *tLocale,
				TranslatedText: *tText,
				LLMProvider:    *tProvider,
				TranslatedAt:   *tTranslatedAt,
				UpdatedAt:      *tUpdatedAt,
			})
		}
	}

	if source == nil {
		return nil, pgx.ErrNoRows
	}
	return source, nil
}

func (s *TranslationService) CreateSourceString(ctx context.Context, req models.CreateRequest) (*models.SourceString, error) {
	var source models.SourceString
	err := s.db.QueryRow(ctx,
		`INSERT INTO source_strings (text) VALUES ($1) RETURNING id, text, created_at, updated_at`,
		req.Text,
	).Scan(&source.ID, &source.Text, &source.CreatedAt, &source.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting source string: %w", err)
	}

	translations, err := s.translateAllProviders(ctx, source.ID, req.Text, req.Locales, req.LLMProviders)
	if translations == nil {
		translations = []models.Translation{}
	}
	source.Translations = translations
	if err != nil {
		return &source, fmt.Errorf("translation completed with errors: %w", err)
	}
	return &source, nil
}

func (s *TranslationService) UpdateSourceString(ctx context.Context, id uuid.UUID, text string) (*models.SourceString, error) {
	var source models.SourceString
	err := s.db.QueryRow(ctx,
		`UPDATE source_strings SET text = $1, updated_at = NOW() WHERE id = $2 RETURNING id, text, created_at, updated_at`,
		text, id,
	).Scan(&source.ID, &source.Text, &source.CreatedAt, &source.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("updating source string: %w", err)
	}
	return &source, nil
}

func (s *TranslationService) DeleteSourceString(ctx context.Context, id uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM source_strings WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting source string: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *TranslationService) UpdateTranslation(ctx context.Context, id uuid.UUID, translatedText string) (*models.Translation, error) {
	var t models.Translation
	err := s.db.QueryRow(ctx,
		`UPDATE translations SET translated_text = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, source_id, locale, translated_text, llm_provider, translated_at, updated_at`,
		translatedText, id,
	).Scan(&t.ID, &t.SourceID, &t.Locale, &t.TranslatedText, &t.LLMProvider, &t.TranslatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("updating translation: %w", err)
	}
	return &t, nil
}

func (s *TranslationService) DeleteTranslation(ctx context.Context, id uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM translations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting translation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *TranslationService) Retranslate(ctx context.Context, id uuid.UUID, req models.RetranslateRequest) (*models.SourceString, error) {
	var text string
	err := s.db.QueryRow(ctx, `SELECT text FROM source_strings WHERE id = $1`, id).Scan(&text)
	if err != nil {
		return nil, fmt.Errorf("fetching source string: %w", err)
	}

	_, translateErr := s.translateAllProviders(ctx, id, text, req.Locales, req.LLMProviders)

	source, err := s.GetSourceString(ctx, id)
	if err != nil {
		return nil, err
	}

	if translateErr != nil {
		return source, fmt.Errorf("translation completed with errors: %w", translateErr)
	}
	return source, nil
}

func (s *TranslationService) translateAllProviders(ctx context.Context, sourceID uuid.UUID, text string, locales []string, providers []string) ([]models.Translation, error) {
	type result struct {
		translations []models.Translation
		err          error
	}

	results := make([]result, len(providers))
	g, gctx := errgroup.WithContext(ctx)

	for i, provider := range providers {
		g.Go(func() error {
			t, err := s.translateAndStore(gctx, sourceID, text, locales, provider)
			results[i] = result{translations: t, err: err}
			// Don't return err — we want all providers to run even if one fails
			return nil
		})
	}

	g.Wait()

	var all []models.Translation
	var errs []error
	for _, r := range results {
		if r.translations != nil {
			all = append(all, r.translations...)
		}
		if r.err != nil {
			errs = append(errs, r.err)
		}
	}

	if len(errs) > 0 {
		return all, fmt.Errorf("%w", errors.Join(errs...))
	}
	return all, nil
}

func (s *TranslationService) translateAndStore(ctx context.Context, sourceID uuid.UUID, text string, locales []string, provider string) ([]models.Translation, error) {
	translator, err := s.llm.NewTranslator(provider)
	if err != nil {
		return nil, err
	}

	translations := make([]models.Translation, len(locales))
	g, gctx := errgroup.WithContext(ctx)

	for i, locale := range locales {
		g.Go(func() error {
			langName := models.ValidLocales[locale]
			translated, err := translator.Translate(gctx, text, langName)
			if err != nil {
				return fmt.Errorf("translating to %s: %w", locale, err)
			}

			var t models.Translation
			err = s.db.QueryRow(gctx,
				`INSERT INTO translations (source_id, locale, translated_text, llm_provider)
				 VALUES ($1, $2, $3, $4)
				 ON CONFLICT (source_id, locale, llm_provider)
				 DO UPDATE SET translated_text = EXCLUDED.translated_text,
				              translated_at = NOW(),
				              updated_at = NOW()
				 RETURNING id, source_id, locale, translated_text, llm_provider, translated_at, updated_at`,
				sourceID, locale, translated, provider,
			).Scan(&t.ID, &t.SourceID, &t.Locale, &t.TranslatedText, &t.LLMProvider, &t.TranslatedAt, &t.UpdatedAt)
			if err != nil {
				return fmt.Errorf("storing translation for %s: %w", locale, err)
			}

			translations[i] = t
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		// Return whatever translations succeeded along with the error
		completed := make([]models.Translation, 0)
		for _, t := range translations {
			if t.ID != uuid.Nil {
				completed = append(completed, t)
			}
		}
		return completed, err
	}

	return translations, nil
}
