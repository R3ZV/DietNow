# DietNow – Architecture & Stack Decision

Status: **Accepted** · Scope: MVP application foundation

This document records the selected frontend/backend stack, why it was chosen,
and the conventions every feature (onboarding, tracking, AI chat, auth) must
follow. Read it before adding new endpoints, pages, or configuration.

## Selected stack

| Layer        | Choice                                             | Version  |
|--------------|----------------------------------------------------|----------|
| Frontend     | React + TypeScript, built with Vite                | React 18, Vite 5, TS 5 |
| Routing (UI) | React Router (`createBrowserRouter`)               | v6       |
| Frontend QA  | ESLint 9 (flat config, typescript-eslint), `tsc`, Vitest + Testing Library | – |
| Backend      | Go, standard library `net/http`                    | Go 1.22+ |
| Persistence  | PostgreSQL 16 via GORM (already in `backend/`)     | –        |
| Backend QA   | `go vet`, `gofmt`, `go test`                       | –        |
| Local infra  | Docker Compose (PostgreSQL only)                   | –        |
| CI           | GitHub Actions (`.github/workflows/ci.yml`)        | –        |

## Rationale

**Go backend (stdlib `net/http`).** The persistence layer (models, migrations,
ownership-enforcing repositories) is already written in Go + GORM, so the API
lives in the same module and calls those repositories directly. Go 1.22 added
method- and wildcard-aware routing to `http.ServeMux`
(`mux.HandleFunc("GET /api/meals/{id}", ...)`), which covers everything the MVP
needs without a third-party router — fewer dependencies, one less thing to learn.
Go compiles to a single static binary, which keeps deployment trivial.

**React + TypeScript + Vite frontend.** The MVP is UI-heavy (multi-step
onboarding forms, a live dashboard with progress bars, a weight chart, a chat
interface). React has the largest ecosystem for these (charting, forms, testing)
and is the stack the team is most familiar with. TypeScript catches API contract
mistakes at compile time. Vite gives a near-instant dev server and a simple
production build to static files.

**Single-page app + JSON API (not server-rendered).** The frontend and backend
are deployed and tested independently and talk only through the documented
`/api` contract. In development the Vite dev server proxies `/api/*` to the Go
server, so the browser sees a single origin (no CORS setup needed locally,
and cookies for auth will "just work").

**Alternatives considered.**
- *Next.js / full-stack JS*: would duplicate or replace the existing Go data layer.
- *Gin / Echo / chi*: not needed now that the stdlib mux supports methods and
  path parameters; can be revisited if middleware needs grow.
- *Server-side templates (Go `html/template`)*: poor fit for the interactive
  dashboard and chat UI.

## Repository layout

```
DietNow/
├── backend/                  Go module (API + persistence)
│   ├── cmd/api/              HTTP server entry point   → go run ./cmd/api
│   ├── cmd/migrate, cmd/seed DB tooling
│   └── internal/
│       ├── config/           server configuration from env
│       ├── httpapi/          router, middleware, JSON helpers, handlers
│       ├── database/         connection, migrations, seed
│       ├── models/           GORM models
│       └── repository/       ownership-enforcing data access
├── frontend/                 React + Vite SPA
│   └── src/
│       ├── api/              typed fetch client for /api
│       ├── components/       shared layout & UI
│       └── pages/            one component per route
└── docs/                     backlog, architecture
```

## API conventions

These conventions apply to every endpoint, including those added by the
database and authentication tasks.

- **Prefix**: all endpoints live under `/api`. Anything else belongs to the SPA.
- **Format**: JSON request and response bodies, `Content-Type: application/json`.
- **Naming**: plural resource nouns, lower-case, kebab-case
  (`/api/meal-entries`, `/api/weight-logs/{id}`); JSON fields in `camelCase`.
- **Dates**: calendar days as `"YYYY-MM-DD"` strings (see `models.LocalDate`);
  instants as RFC 3339 UTC timestamps.
- **Errors**: non-2xx responses always have the shape
  ```json
  { "error": { "code": "not_found", "message": "resource not found" } }
  ```
  `code` is a stable machine-readable snake_case string; `message` is for humans.
- **Status codes**: `200` read/update, `201` create, `204` delete, `400`
  validation, `401` unauthenticated, `403` forbidden, `404` not found (also used
  for resources owned by another user — never reveal existence), `503` dependency
  unavailable.
- **Authentication** (to be implemented by the auth task): the authenticated
  user id comes **only** from the session/token, never from the request body
  or URL, and is passed as the trusted `userID` to the repositories.
- **Health**:
  - `GET /api/health` — liveness. Always `200` while the process is serving.
  - `GET /api/health/ready` — readiness. `200` when PostgreSQL answers a ping,
    `503` otherwise.

## Configuration conventions

- Configuration comes from **environment variables only**; nothing secret is
  committed. Each app ships a `.env.example` with placeholders; copy it to
  `.env` (git-ignored) for local development.
- Backend variables are prefixed by concern: `API_*` (HTTP server), `DB_*`
  (database), and future `AUTH_*` (sessions/JWT), `AI_*` (LLM provider).
- Frontend variables must start with `VITE_` to be exposed to the browser —
  therefore **never** put secrets in frontend env files.
- Real environment variables always override values in `.env`.
