# LOCALIZATION Localization

A full-stack localization application that translates American English text into multiple languages using LLM APIs, with a Go backend, PostgreSQL database, and React frontend.

## Requirements

1. **Core translation service in Go** -- Translate American English strings into:
   - British English (`en-GB`)
   - Spanish (`es`)
   - Traditional Chinese (`zh-Hant`)
   - Simplified Chinese (`zh-Hans`)
   - Hindi (`hi`)

2. **PostgreSQL persistence** -- Store the source string, translated strings, translation timestamp, locale, and which LLM was used.

3. **React UI** -- Create, edit, and delete translations. On creation, the user selects target language(s) via checkboxes. The main list page displays all translation details (locale, translated text, LLM provider, timestamp) inline without requiring navigation to a detail page.

## Architecture

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│   React UI   │──────▶│  Go Backend  │──────▶│  PostgreSQL  │
│  (Vite/TS)   │ REST  │   (Chi)      │  SQL  │    (v17)     │
│  port 5173   │◀──────│  port 8080   │◀──────│  port 5432   │
└──────────────┘       └──────┬───────┘       └──────────────┘
                              │
                    ┌─────────┴─────────┐
                    ▼                   ▼
              ┌──────────┐       ┌──────────┐
              │ ChatGPT 5│       │Gemini 2.5│
              │ (OpenAI) │       │  Pro     │
              └──────────┘       └──────────┘
```

- **Frontend** proxies `/api` requests to the backend via Vite dev server
- **Backend** translates concurrently across selected locales using `errgroup`
- **LLM selection** is per-request -- the user can select one or both of ChatGPT 5 and Gemini 2.5 Pro via checkboxes. Selecting both runs translations concurrently across all providers
- **LLM calls** have a configurable timeout per locale (default 600s from YAML config) to prevent hangs
- **Partial failures** are surfaced as warnings -- if some locales or providers fail, the successful ones are saved and the error is shown to the user

### LLM Configuration

| Provider   | Model           | Auth Method                      | Temperature | Config File |
|------------|-----------------|----------------------------------|-------------|-------------|
| ChatGPT 5  | `gpt-5`         | `OPENAI_API_KEY` env var or GCP Secret Manager | 1.0 | `gpt-5.yaml` |
| Gemini 2.5 | `gemini-2.5-pro` | Application Default Credentials (ADC) via Vertex AI | 1.0 | `gemini.yaml` |

**OpenAI key resolution order** (from `gpt-5.yaml`):
1. `OPENAI_API_KEY` environment variable
2. GCP Secret Manager (`prj-dm-ds-v0-pdigi-dev-5gf` / `open-ai-key` / version 3)

**Gemini** requires no API key -- it uses Google Cloud ADC via the Vertex AI backend (from `gemini.yaml`). Run `gcloud auth application-default login` to authenticate. The GCP project and location can be overridden with `GCP_PROJECT` and `GCP_LOCATION` environment variables (defaults: `prj-dm-ds-v0-pdigi-dev-5gf` / `us-central1`).

**Multi-provider comparison**: Users can select multiple LLM providers simultaneously. All selected providers run concurrently, and results are stored independently per `(source, locale, provider)`. The detail page groups translations by locale with provider results displayed side-by-side for easy comparison.

## Project Structure

```
localization/
├── startup.sh                          # Start all services (Colima, Postgres, backend, frontend)
├── docker-compose.yml                  # PostgreSQL, backend, frontend
├── .env.example                        # Environment variable template
├── gpt-5.yaml                          # OpenAI LLM config (model, temperature, timeout, secrets)
├── gemini.yaml                         # Gemini LLM config (model, temperature, timeout, ADC)
│
├── backend/
│   ├── main.go                         # Entrypoint: Chi router, CORS, graceful shutdown
│   ├── Dockerfile
│   ├── go.mod
│   └── internal/
│       ├── config/
│       │   ├── config.go               # Env vars + GCP Secret Manager
│       │   └── llmconfig.go            # YAML config loader for LLM providers
│       ├── database/
│       │   ├── database.go             # pgxpool connection
│       │   └── migrations/001_init.sql # Schema DDL
│       ├── models/models.go            # Data structs, valid locales/providers
│       ├── handlers/translations.go    # HTTP handlers (CRUD + error/warning responses)
│       ├── service/translation.go      # Business logic, concurrent LLM calls
│       └── llm/
│           ├── client.go               # Translator interface + factory
│           ├── openai.go               # ChatGPT 5 implementation
│           └── gemini.go               # Gemini 2.5 Pro implementation
│
└── frontend/
    ├── Dockerfile
    ├── vite.config.ts                  # Dev server proxy to backend
    └── src/
        ├── App.tsx                     # Router: /, /new, /source/:id
        ├── api/translations.ts         # REST API client (90s fetch timeout, null-safe)
        ├── types/index.ts              # TypeScript interfaces
        └── components/
            ├── TranslationList.tsx      # Main page: source strings with inline translation details
            ├── TranslationForm.tsx      # Create page: text + locale checkboxes + LLM provider checkboxes
            ├── TranslationDetail.tsx    # View/edit/delete translations, retranslate with multi-provider
            ├── LocaleCheckboxes.tsx     # Locale multi-select checkboxes
            ├── LlmSelector.tsx          # LLM provider multi-select checkboxes
            └── ConfirmDelete.tsx        # Delete confirmation modal
```

## Database Schema

Two normalized tables -- one row per locale per provider:

```sql
source_strings
├── id          UUID (PK)
├── text        TEXT (unique)
├── created_at  TIMESTAMPTZ
└── updated_at  TIMESTAMPTZ

translations
├── id              UUID (PK)
├── source_id       UUID (FK → source_strings, CASCADE delete)
├── locale          VARCHAR(10)       -- en-GB, es, zh-Hant, zh-Hans, hi
├── translated_text TEXT
├── llm_provider    VARCHAR(20)       -- chatgpt5, gemini25
├── translated_at   TIMESTAMPTZ
├── updated_at      TIMESTAMPTZ
└── UNIQUE(source_id, locale, llm_provider)
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/health` | Health check |
| `GET` | `/api/source-strings` | List all source strings with translations |
| `GET` | `/api/source-strings/{id}` | Get one source string with translations |
| `POST` | `/api/source-strings` | Create source string and translate |
| `PUT` | `/api/source-strings/{id}` | Update source text |
| `DELETE` | `/api/source-strings/{id}` | Delete source string (cascades translations) |
| `PUT` | `/api/translations/{id}` | Edit a single translation |
| `DELETE` | `/api/translations/{id}` | Delete a single translation |
| `POST` | `/api/source-strings/{id}/retranslate` | Re-translate selected locales |

Responses for `POST /api/source-strings` and `POST /api/source-strings/{id}/retranslate` may include a `"warning"` field if some translations failed while others succeeded.

### Example: Create a translation

```bash
curl -X POST http://localhost:8080/api/source-strings \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Hello, how are you?",
    "locales": ["en-GB", "es", "zh-Hant", "zh-Hans", "hi"],
    "llm_providers": ["chatgpt5", "gemini25"]
  }'
```

## Prerequisites

- **Go** 1.25+
- **Node.js** 22+
- **Colima** (container runtime for macOS -- provides Docker CLI)
- **Docker Buildx** plugin (`brew install docker-buildx`, then `mkdir -p ~/.docker/cli-plugins && ln -sfn $(which docker-buildx) ~/.docker/cli-plugins/docker-buildx`)
- **Google Cloud CLI** (`gcloud auth application-default login` for Gemini ADC and GCP Secret Manager)

## Getting Started

### Quick start

```bash
# 1. Clone and enter the project
cd localization

# 2. Set up environment
cp .env.example .env
# Edit .env with your OpenAI API key (or rely on GCP Secret Manager)

# 3. Authenticate for Gemini (one-time)
gcloud auth application-default login

# 4. Install frontend dependencies (one-time)
cd frontend && npm install && cd ..

# 5. Start everything
./startup.sh
```

The startup script will:
1. Load environment variables from `.env`
2. Start Colima if not already running (auto-recovers from stale state)
3. Start PostgreSQL in a Docker container with a named volume (`localization-pgdata`) for data persistence
   - Schema is applied automatically on first run via the migration file
   - Data (translations, source strings) persists across `Ctrl+C` / restart cycles
4. Start the Go backend and wait for it to be healthy
5. Start the Vite dev server (only after the backend is ready)

Open http://localhost:5173 in your browser. Press `Ctrl+C` to stop all processes.

> **Note:** To fully reset the database, run `docker volume rm localization-pgdata` when the container is stopped.

### Manual start

If you prefer to start services individually:

```bash
# Terminal 1: Start Colima + PostgreSQL
colima start
docker volume create localization-pgdata
docker run -d --name localization-postgres \
  -e POSTGRES_USER=localization -e POSTGRES_PASSWORD=localization -e POSTGRES_DB=localization \
  -p 5432:5432 \
  -v localization-pgdata:/var/lib/postgresql/data \
  -v $(pwd)/backend/internal/database/migrations:/docker-entrypoint-initdb.d \
  postgres:17

# Terminal 2: Go backend
cd backend
source ../.env
go run .

# Terminal 3: React frontend (start after backend is ready)
cd frontend
npm run dev
```

### Docker Compose (full stack)

```bash
docker compose up --build
```

This builds and runs all three services. Docker Compose also:
- Mounts `gpt-5.yaml` and `gemini.yaml` into the backend container (via `CONFIG_DIR=/config`)
- Mounts `~/.config/gcloud/application_default_credentials.json` for ADC (Gemini Vertex AI + GCP Secret Manager)
- Applies the database schema automatically on first run via the migration file mounted into PostgreSQL's init directory

To fully reset the database: `docker compose down -v` removes the data volume, then `docker compose up --build` recreates it.

## Error Handling

- **LLM timeouts**: Each translation call has a configurable timeout (set in `gpt-5.yaml` / `gemini.yaml`, default 600s). If the LLM API is slow or unresponsive, the request will fail gracefully instead of hanging.
- **Partial failures**: If translating with some locales or providers succeeds but others fail, the successful translations are saved and a warning is shown in the UI.
- **Frontend timeouts**: API calls from the browser have a 90-second timeout with an `AbortController`. If a request times out, the UI shows an error message instead of hanging.
- **Null safety**: The API client normalizes all responses to ensure `translations` is always an array, preventing rendering crashes.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, Chi router, pgx (raw SQL) |
| Database | PostgreSQL 17 |
| Frontend | React 19, TypeScript, Vite, React Router |
| LLM | OpenAI API (gpt-5), Google Vertex AI (gemini-2.5-pro) |
| LLM Config | YAML files (`gpt-5.yaml`, `gemini.yaml`) |
| Secrets | GCP Secret Manager, ADC, environment variables |
| Container Runtime | Colima (macOS) |
