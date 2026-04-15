# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Full-stack localization app: Go backend + React frontend + PostgreSQL. Translates American English into 5 locales using OpenAI GPT-5 and Google Gemini 2.5 Pro concurrently. See README.md for full architecture diagram, API endpoints, database schema, and setup instructions.

## Common Commands

### Full stack (recommended)
```bash
./startup.sh          # Starts Colima, PostgreSQL, backend (:8080), frontend (:5173)
# Ctrl+C to stop all
```

### Backend (Go)
```bash
cd backend
go run .              # Run server (requires DATABASE_URL, see .env.example)
go build -o server    # Build binary
go fmt ./...          # Format
go vet ./...          # Static analysis
go mod tidy           # Clean dependencies
```

### Frontend (React/TypeScript)
```bash
cd frontend
npm install           # Install dependencies (first time)
npm run dev           # Dev server at :5173 (proxies /api to :8080)
npm run build         # Production build
npm run lint          # ESLint
npx tsc -b            # Type check
```

### Docker Compose
```bash
docker compose up --build    # Full stack
docker compose down -v       # Reset (destroys database volume)
```

### Database
```bash
# Connect directly
psql postgres://localization:localization@localhost:5432/localization

# Full reset (container must be stopped)
docker volume rm localization-pgdata
```

## Architecture

**Backend request flow:** `main.go` (Chi router) -> `handlers/` (validation, HTTP concerns) -> `service/` (business logic) -> `llm/` (LLM API calls) + `database/` (pgx raw SQL).

**Key patterns:**

- **`Translator` interface** (`internal/llm/client.go`): `Translate(ctx, text, targetLanguage) (string, error)`. Implementations in `openai.go` and `gemini.go`. `ClientFactory.NewTranslator(provider)` dispatches by provider name (`"chatgpt5"`, `"gemini25"`).

- **Two-level concurrency** (`internal/service/translation.go`): `translateAllProviders()` fans out across providers via `errgroup`, then `translateAndStore()` fans out across locales within each provider. Errors are collected (not short-circuited) so partial results are saved.

- **Upsert pattern**: Translations use `ON CONFLICT (source_id, locale, llm_provider) DO UPDATE` so retranslation overwrites in-place.

- **Handlers are thin**: Validation and JSON response formatting only. Business logic lives in the service layer.

- **No ORM**: All SQL is raw via `pgx`. No test suite exists currently.

## Configuration

- **YAML configs** (`gpt-5.yaml`, `gemini.yaml` at repo root): model name, temperature, timeout per provider. Loaded from `CONFIG_DIR` env var (defaults to `..` relative to backend binary, i.e., repo root).
- **OpenAI key**: `OPENAI_API_KEY` env var takes precedence; falls back to GCP Secret Manager if configured in YAML.
- **Gemini auth**: Uses Application Default Credentials (ADC) via Vertex AI. No API key needed. Run `gcloud auth application-default login` once.
- **Valid locales** are defined in `internal/models/models.go` (`ValidLocales` map) and mirrored in `frontend/src/types/index.ts` (`LOCALES`).
- **Valid providers** are defined in `internal/models/models.go` (`ValidProviders` map) and mirrored in `frontend/src/types/index.ts` (`LLM_PROVIDERS`).

## Adding a New LLM Provider

1. Implement `Translator` interface in `internal/llm/`
2. Add case to `ClientFactory.NewTranslator()` in `internal/llm/client.go`
3. Add to `ValidProviders` in `internal/models/models.go`
4. Add to `LLM_PROVIDERS` in `frontend/src/types/index.ts`
5. Create YAML config file at repo root
