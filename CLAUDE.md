# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Backend (Go)
```bash
make run           # Run backend dev server (serves frontend/dist at :9000)
make build         # Build the Go binary
make test          # Run Go tests (go test ./...)
```

### Frontend (Vue 3)
```bash
make run-frontend  # Start Vite dev server at :8080 (proxies API to :9000)
cd frontend && yarn dev    # Same as above
cd frontend && yarn build  # Build production bundle to frontend/dist
cd frontend && yarn lint   # ESLint on .js and .vue files
```

### Both together
Run `make run` (backend on :9000) and `make run-frontend` (frontend on :8080) in separate terminals. The Vite dev server proxies all API/subscription/webhook routes to `:9000`.

### E2E Tests (Cypress)
```bash
cd frontend && npx cypress open  # Interactive mode
cd frontend && npx cypress run   # Headless (requires running server at :9000)
```
Cypress config: `frontend/cypress.config.js`. Tests use admin/listmonk credentials against a fresh listmonk instance.

### Full distribution build
```bash
make dist          # Builds backend + frontend + packs with stuffbin into single binary
```

## Architecture

### Backend (Go)

Entry point: `cmd/main.go`. The `App` struct wires together all dependencies:
- **`internal/core/`** — all database CRUD (lists, subscribers, campaigns, templates, bounces, users, roles, media, settings). Uses `sqlx` with pre-compiled queries from `queries/*.sql` loaded via `knadh/goyesql`.
- **`internal/manager/`** — campaign scheduler and message dispatcher. Runs workers that pull queued campaigns, render templates, and dispatch via messengers.
- **`internal/auth/`** — JWT sessions, TOTP 2FA, role-based permission checks.
- **`internal/bounce/`** — bounce processing via webhooks (Sendgrid, Postmark, etc.) and POP3 mailbox polling.
- **`internal/messengers/email/`** — SMTP pool via `smtppool`.
- **`internal/media/`** — media storage abstraction (S3-compatible or local filesystem).
- **`internal/events/`** — SSE event stream for real-time UI updates.

HTTP server: `labstack/echo/v4`. Routes and middleware are in `cmd/handlers.go`.

Database: PostgreSQL. Schema in `schema.sql`. Queries are raw SQL files, not an ORM.

Config: TOML via `knadh/koanf`. Environment variables override config using `LISTMONK_` prefix with `__` as dot separator (e.g. `LISTMONK_db__host`).

Static assets (frontend dist, schema, i18n files) are embedded into the binary at build time using `stuffbin`. In dev mode (`make run`), they're served from disk.

### Frontend (Vue 3 + PrimeVue)

**Stack**: Vue 3 Options API · PrimeVue 4 (Aura theme, blue preset) · Pinia · Vue Router 4 · PrimeFlex · Axios

**Key files**:
- `frontend/src/main.js` — app bootstrap: creates Vue app, registers PrimeVue components globally with `Pv*` prefix (e.g. `PvButton`, `PvDataTable`), loads server config and user profile before mount.
- `frontend/src/store/index.js` — single Pinia store (`useMainStore`). Holds all model data (lists, campaigns, subscribers, etc.), per-model `loading` flags, and a `refreshTick` counter used to trigger view re-fetches without a full page reload.
- `frontend/src/api/index.js` — all Axios API calls. Interceptors auto-convert snake_case responses to camelCase, set loading flags, and update the store. Each call can pass `{ loading: 'modelName', store: 'modelName' }` config to wire these automatically.
- `frontend/src/router/index.js` — 35 lazily-loaded routes.
- `frontend/src/assets/style.scss` — global SCSS. CSS custom properties for the design system are defined here (`--lm-primary`, `--lm-surface`, `--lm-border`, `--lm-text`, `--lm-text-muted`, `--lm-bg`, `--lm-bg-subtle`).
- `frontend/src/constants.js` — model names, URI prefixes, regex helpers (e.g. `regDuration`).

**PrimeVue component API reference**: `Primevue.md` in the repo root documents the props, variants, design tokens, and PassThrough (PT) options for PrimeVue components used in this project. Check it before guessing at component APIs — it covers the exact prop names, severity values, variant strings, and CSS class names.

**Component conventions (PrimeVue migration)**:
- All PrimeVue components are registered globally and used with the `Pv` prefix.
- Dialog/modal layout uses `.lm-form` / `.lm-form-header` / `.lm-form-body` / `.lm-form-footer` classes. `.lm-form-body` uses `flex-direction: column; gap: 1.1rem` — set `.field { margin-bottom: 0 }` inside it to avoid double spacing.
- Settings sub-views use `.settings-card` (bordered content sections), `.settings-section-label` (uppercase muted headers), `.items` (flex-column wrapper), and `.quick-links` (inline link row).
- `PvPassword` requires `:deep(.p-password) { width: 100% }` and `:deep(.p-password-input) { width: 100% }` for full-width layout.
- State is read from the store via `mapState(useMainStore, [...])` in the Options API `computed` block.
- `refreshTick` from the store replaces the old `$root` event bus for triggering data reloads.

**Vite dev proxy** (`frontend/vite.config.js`): Paths `/`, `/api/*`, `/webhooks/*`, `/subscription/*`, `/public/*`, `/health`, `/admin/login` all proxy to `http://127.0.0.1:9000`.

### Email Builder

Separate Vite/TypeScript project in `frontend/email-builder/`. Built independently with `make build-email-builder` and output to `frontend/public/static/email-builder/`. Embedded in the template editor as a widget.
