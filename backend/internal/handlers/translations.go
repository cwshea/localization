package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/cwshea/localization/internal/models"
	"github.com/cwshea/localization/internal/service"
)

type Handler struct {
	svc *service.TranslationService
}

func NewHandler(svc *service.TranslationService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListSourceStrings(w http.ResponseWriter, r *http.Request) {
	sources, err := h.svc.ListSourceStrings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sources)
}

func (h *Handler) GetSourceString(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	source, err := h.svc.GetSourceString(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "source string not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, source)
}

func (h *Handler) CreateSourceString(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}
	if len(req.Locales) == 0 {
		writeError(w, http.StatusBadRequest, "at least one locale is required")
		return
	}
	for _, locale := range req.Locales {
		if _, ok := models.ValidLocales[locale]; !ok {
			writeError(w, http.StatusBadRequest, "invalid locale: "+locale)
			return
		}
	}
	if len(req.LLMProviders) == 0 {
		writeError(w, http.StatusBadRequest, "at least one llm_provider is required")
		return
	}
	for _, p := range req.LLMProviders {
		if !models.ValidProviders[p] {
			writeError(w, http.StatusBadRequest, "invalid llm_provider: "+p)
			return
		}
	}

	source, err := h.svc.CreateSourceString(r.Context(), req)
	if err != nil {
		if source != nil {
			// Partial success: source created but some translations failed
			log.Printf("Translation error for source %s: %v", source.ID, err)
			writeJSON(w, http.StatusCreated, struct {
				*models.SourceString
				Warning string `json:"warning"`
			}{source, err.Error()})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, source)
}

func (h *Handler) UpdateSourceString(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req models.UpdateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	source, err := h.svc.UpdateSourceString(r.Context(), id, req.Text)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "source string not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, source)
}

func (h *Handler) DeleteSourceString(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.DeleteSourceString(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "source string not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateTranslation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req models.UpdateTranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.TranslatedText == "" {
		writeError(w, http.StatusBadRequest, "translated_text is required")
		return
	}

	t, err := h.svc.UpdateTranslation(r.Context(), id, req.TranslatedText)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "translation not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) DeleteTranslation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.DeleteTranslation(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "translation not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Retranslate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req models.RetranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Locales) == 0 {
		writeError(w, http.StatusBadRequest, "at least one locale is required")
		return
	}
	for _, locale := range req.Locales {
		if _, ok := models.ValidLocales[locale]; !ok {
			writeError(w, http.StatusBadRequest, "invalid locale: "+locale)
			return
		}
	}
	if len(req.LLMProviders) == 0 {
		writeError(w, http.StatusBadRequest, "at least one llm_provider is required")
		return
	}
	for _, p := range req.LLMProviders {
		if !models.ValidProviders[p] {
			writeError(w, http.StatusBadRequest, "invalid llm_provider: "+p)
			return
		}
	}

	source, err := h.svc.Retranslate(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "source string not found")
			return
		}
		// Translation may have partially succeeded -- return what we have with the error
		if source != nil {
			log.Printf("Retranslation error for source %s: %v", id, err)
			writeJSON(w, http.StatusOK, struct {
				*models.SourceString
				Warning string `json:"warning"`
			}{source, err.Error()})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, source)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
