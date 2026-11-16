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
│   ├── cmd/migrate, cmd/seed DB tooling; cmd/smoke = MVP smoke test
│   └── internal/
│       ├── config/           server configuration from env
│       ├── httpapi/          router, middleware, JSON helpers, handlers, SPA serving
│       ├── auth/             bcrypt passwords, server-side sessions
│       ├── nutrition/        calorie engine (BMR/TDEE, surplus, macros)
│       ├── smoke/            MVP smoke flow
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
  (`/api/meals/{id}`, `/api/daily-summary`); JSON fields in `camelCase`.
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
- **Authentication** (see `backend/AUTH.md`): the authenticated user id comes
  **only** from the session cookie, never from the request body or URL, and is
  passed as the trusted `userID` to the repositories.
- **Validation errors** are `400` with code `validation_failed` and a `fields`
  map naming each invalid JSON field, so forms can highlight them:
  ```json
  { "error": { "code": "validation_failed", "message": "some fields are invalid",
               "fields": { "currentWeightKg": "must be between 30 and 300 kg" } } }
  ```
- **Health**:
  - `GET /api/health` — liveness. Always `200` while the process is serving;
    reports the deployed `version` (git commit).
  - `GET /api/health/ready` — readiness. `200` when PostgreSQL answers a ping,
    `503` otherwise. Used as the hosting health check.

## Endpoints

All endpoints except health and register/login/logout require the session
cookie (`401 unauthenticated` otherwise). Another user's data is always
reported as `404`.

| Method & path | Body / query | Success | Story |
|---------------|--------------|---------|-------|
| `POST /api/auth/register` | `{email, password, displayName?}` | `201 {user, expiresAt}` + cookie | US-19 |
| `POST /api/auth/login` | `{email, password}` | `200 {user, expiresAt}` + cookie | US-19 |
| `POST /api/auth/logout` | – | `204` | US-19 |
| `GET /api/auth/me` | – | `200 {user, expiresAt}` | US-19 |
| `PUT /api/profile` | `{gender, birthDate, heightCm, currentWeightKg, activityLevel, goal, targetWeightKg?, timeframeWeeks?, timezone?}` | `200 {profile, warnings}` — profile includes `bmr`, `tdee`, `weeklyRateKg`, `dailyAdjustmentKcal`, `targetCalories`, `targetProteinG/CarbsG/FatG` | US-01…06, US-20 |
| `GET /api/profile` | – | `200 {profile, warnings: []}`, `404 profile_not_found` before onboarding | US-01…06 |
| `GET /api/foods` | `?q=<2-100 chars>&limit=1..100` | `200 {foods}` — global + own private foods, nutrition per 100 g | US-09 |
| `POST /api/meals` | `{date, mealType, foodId, grams}` (`0 < grams ≤ 2000`) | `201` entry with scaled, frozen `nutrition` | US-10 |
| `GET /api/meals` | `?date=YYYY-MM-DD` | `200 {date, entries}` | US-08 |
| `DELETE /api/meals/{id}` | – | `204` | US-10 |
| `GET /api/daily-summary` | `?date=YYYY-MM-DD` (default: today in the profile time zone) | `200 {date, targets, consumed, remaining, meals[4]}`; `targets`/`remaining` are `null` before onboarding, `remaining` goes negative when over budget | US-08, US-11 |
| `GET /api/weights` | `?from=&to=` | `200 {entries}` | US-12 |

**Calorie engine** (`backend/internal/nutrition`): Mifflin-St Jeor BMR, TDEE
multipliers 1.2 / 1.375 / 1.55 / 1.725 / 1.9, a bulk surplus of 5600 kcal per kg
of planned gain and a cut deficit of 7700 kcal per kg of planned loss (spread
over the timeframe), a floor of 1200 kcal (female) or 1500 kcal (male), and
protein 2.0 g/kg + fat 1.0 g/kg with the rest as carbohydrates. A pace faster
than 0.5 kg/week (bulk) or 1 kg/week (cut) is saved but returned as a warning.
Inputs are bounded: age 13–100, height 100–250 cm, weight 30–300 kg, 1–104
weeks.

## Configuration conventions

- Configuration comes from **environment variables only**; nothing secret is
  committed. Each app ships a `.env.example` with placeholders; copy it to
  `.env` (git-ignored) for local development.
- Backend variables are prefixed by concern: `API_*` (HTTP server), `DB_*`
  (database), `SESSION_*`/`COOKIE_*`/`BCRYPT_COST` (auth), future `AI_*` (LLM provider).
- Frontend variables must start with `VITE_` to be exposed to the browser —
  therefore **never** put secrets in frontend env files.
- Real environment variables always override values in `.env`.
- Staging/production configuration lives in the hosting platform and the CI
  secret store, never in the repository; see [DEPLOYMENT.md](DEPLOYMENT.md).

## Deployment shape

One Docker image (`Dockerfile` at the repo root) holds the Go API, the
`migrate` tool and the built frontend. The API serves the SPA for every
non-`/api` path (`STATIC_DIR`), so browser, frontend and API share one origin.
Staging runs this image on Render with a managed PostgreSQL; see
[DEPLOYMENT.md](DEPLOYMENT.md).
