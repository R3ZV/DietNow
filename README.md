# DietNow 🥗

> **Intelligent Nutrition, Calorie & Weight Tracker with an AI Recipe Assistant**

DietNow is a modern web application designed to simplify personal
nutrition and help users reach their physique goals (bulking, cutting, or maintenance)
through streamlined tracking and real-time AI recipe generation, while also
educating in regards to longevity foods that are missing from his / her diet.

## Tech stack

| Part      | Technology                                         | Folder      |
|-----------|----------------------------------------------------|-------------|
| Frontend  | React 18 + TypeScript + Vite, React Router         | `frontend/` |
| Backend   | Go 1.22+ (`net/http`), GORM                        | `backend/`  |
| Database  | PostgreSQL 16 (Docker Compose)                     | `backend/docker-compose.yml` |

Why this stack, the API conventions (`/api` prefix, JSON error envelope, status
codes) and the configuration conventions are documented in
👉 **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)**.

## Getting started

### Prerequisites

| Tool | Version | Install (Windows) | Check |
|------|---------|-------------------|-------|
| Git | any | https://git-scm.com/download/win | `git --version` |
| Go | 1.22 or newer | https://go.dev/dl/ (MSI installer) or `winget install GoLang.Go` | `go version` |
| Node.js (includes npm) | 20 or newer (22 LTS recommended) | https://nodejs.org or `winget install OpenJS.NodeJS.LTS` | `node -v` |
| Docker Desktop | any recent | https://www.docker.com/products/docker-desktop/ or `winget install Docker.DockerDesktop` | `docker info` |

Open a **new** terminal after installing so the tools are on your `PATH`.
Docker Desktop must be **running** (whale icon in the tray) before `docker compose up`.

### 1. Configure

```bash
cp backend/.env.example backend/.env       # PowerShell: Copy-Item backend/.env.example backend/.env
cp frontend/.env.example frontend/.env     # PowerShell: Copy-Item frontend/.env.example frontend/.env
```

Edit `backend/.env` and set your own `DB_PASSWORD`. The `.env` files are
git-ignored; the `.env.example` files contain placeholders only.

### 2. Start the database, then the API (terminal 1)

```bash
cd backend
docker compose up -d           # PostgreSQL on localhost:${DB_PORT}
go run ./cmd/seed              # optional: create tables + test data
go run ./cmd/api               # API on http://localhost:8080
```

### 3. Start the web app (terminal 2)

```bash
cd frontend
npm ci                         # exact versions from package-lock.json
npm run dev                    # http://localhost:5173
```

Open http://localhost:5173 — the home page shows **"API: online"** when the
frontend can reach the backend. The dev server forwards every `/api/*` request
to the Go API, so the browser only ever talks to one origin.

Health endpoints:

```bash
curl http://localhost:8080/api/health          # liveness  -> 200 {"status":"ok",...}
curl http://localhost:8080/api/health/ready    # readiness -> 200 if PostgreSQL is reachable, else 503
```

The API also starts when the database is down (liveness stays `200`,
readiness returns `503`), so the frontend can be developed without Docker.

### Commands

| Task | Backend (`cd backend`) | Frontend (`cd frontend`) |
|------|------------------------|--------------------------|
| Install dependencies | `go mod download` | `npm ci` |
| Run in development | `go run ./cmd/api` | `npm run dev` |
| Build | `go build -o bin/api ./cmd/api` | `npm run build` (output in `dist/`) |
| Lint / static checks | `go vet ./...` and `gofmt -l .` (must print nothing) | `npm run lint` |
| Type check | (part of `go build` / `go vet`) | `npm run typecheck` |
| Tests | `go test ./...` (integration tests need the DB; see [backend/README.md](backend/README.md#tests)) | `npm test` |
| DB migrate / seed | `go run ./cmd/migrate` / `go run ./cmd/seed` | – |
| Smoke test a running instance | `SMOKE_EMAIL=… SMOKE_PASSWORD=… go run ./cmd/smoke -base-url http://localhost:8080 -register` | – |
| Release image (API + frontend) | `docker build -t dietnow .` (repo root) | – |

The same checks, plus PostgreSQL integration tests and a smoke test of the
built Docker image, run in CI on every pull request and push to `master`
(`.github/workflows/ci.yml`). Green `master` builds deploy to **staging** on
Render; hosting, secrets, migrations, deploy and rollback are documented in
👉 **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)**.

### Configuration

**`backend/.env`** (read by `docker compose` and all Go commands; real environment variables win):

| Variable | Default | Purpose |
|----------|---------|---------|
| `APP_ENV` | `development` | `development`, `test`, `staging` or `production` (the last two enable strict startup checks) |
| `API_PORT` | `8080` | Port the API listens on (`PORT`, set by hosting platforms, wins) |
| `API_ALLOWED_ORIGINS` | *(empty)* | Comma-separated origins allowed for CORS; leave empty when using the Vite proxy |
| `STATIC_DIR` | *(empty)* | Serve the built frontend from this directory (set in the Docker image) |
| `DB_AUTO_MIGRATE` | `true` | Apply migrations + food catalog at startup |
| `SESSION_*`, `COOKIE_*`, `BCRYPT_COST` | see `.env.example` | Sessions & password hashing ([backend/AUTH.md](backend/AUTH.md)) |
| `DB_HOST`, `DB_PORT` | `localhost`, `5432` | PostgreSQL address (Docker publishes it on `DB_PORT`) |
| `DB_USER`, `DB_PASSWORD`, `DB_NAME` | `dietnow`, –, `dietnow` | Credentials; Docker creates the database with these |
| `DB_SSLMODE` | `disable` | Use `require` for hosted databases |
| `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME_MIN` | `20`, `10`, `60` | Connection pool |
| `DB_LOG_LEVEL` | `warn` | GORM logging: `silent`, `error`, `warn`, `info` |

**`frontend/.env`** (only `VITE_*` variables; bundled into public JS, so never secrets):

| Variable | Default | Purpose |
|----------|---------|---------|
| `VITE_API_PROXY_TARGET` | `http://localhost:8080` | Where the dev server forwards `/api` |

### Troubleshooting

- **`go`/`npm`/`docker` is not recognized** — install it (see Prerequisites) and open a new terminal.
- **`docker compose` fails with "cannot find the file specified … dockerDesktopLinuxEngine"** — Docker Desktop is not running; start it and wait until it says "Engine running".
- **Readiness returns 503 / `role "dietnow" does not exist`** — another PostgreSQL (e.g. a locally installed one) is already using port 5432, so the API connects to it instead of the Docker one. Set `DB_PORT=5433` in `backend/.env`, then `docker compose up -d` again and restart the API.
- **Port 8080 or 5173 already in use** — change `API_PORT` (and `VITE_API_PROXY_TARGET` to match), or stop the other program.

## Team

- Maracine Constantin-Razvan - Product Owner
- Antonie Chirilus - Scrum Master
- Stancu Rares - Developer
- Istrate Irina - Developer
- Petrache Alexia - Developer

## 📋 Product Backlog

The complete, official Product Backlog is documented in:
👉 **[docs/PRODUCT_BACKLOG.md](docs/PRODUCT_BACKLOG.md)**

- `EPIC 1`: Onboarding & Biometric Profiling
- `EPIC 2`: Caloric Calculation & Nutrition Plans
- `EPIC 3`: Daily Tracking (4 Meals & Weight Progress)
- `EPIC 4`: Conversational AI Nutritionist (AI Chat)
- `EPIC 5`: Account Management & Enhancements
- `EPIC 6`: Longevity Recommendation Engine & Scientific Insights (Bryan Johnson Research)

## 💡 Project Goal

In a fast-paced society where fast food is booming, more people are neglecting
their nutrition to save time—ultimately trading years of their lives for short-term
convenience. Our application draws on Bryan Johnson’s longevity research to help
users understand the true impact of diet on their bodies.

By cutting through the noise of online influencers pushing cookbooks and trendy
recipes, we deliver strictly science-backed insights into the foods proven to
extend your lifespan.

DietNow removes the guesswork and decision fatigue around daily eating through
two interconnected pillars:

### 1. Guided Onboarding & Goal Calculation
Users complete an initial profiling questionnaire covering:
- **Biometrics:** Gender (male/female), age, height (cm), and current weight (kg).
- **Activity Level:** Daily physical exertion and training frequency.
- **Goal & Timeline:** Target weight (kg) for bulk/cut and the desired timeframe (weeks/months).
- **Automated Engine:** Automatically computes BMR, TDEE, required daily caloric
surplus/deficit, and optimal macronutrient distribution (protein, carbohydrates, fats).

### 2. Core Feature 1: Calorie & Weight Tracking + Nutrition Plans
- **4-Meal Daily Tracker:** Dedicated sections for **Breakfast**, **Lunch**, **Dinner**, and **Snacks**.
- **Real-Time Dashboard:** Instant visual feedback of consumed vs. remaining
calories and progress towards macro targets.
- **Weight Progress:** Historical weigh-in log and trajectory toward the user's target weight.
- **Structured Nutrition Plans:** Choose from pre-configured meal regimes
(e.g., High-Protein Bulk, Balanced, Vegetarian, Low-Carb).

### 3. Core Feature 2: Conversational AI Nutritionist (AI Chat)
- **Natural Language Chat:** Ask anything about food, nutrition, or cravings.
- **Fridge-to-Plate Recipe Generator:** Tell the AI what ingredients you have at
home and what meal you want to make; the AI outputs a complete step-by-step recipe
tailored to your remaining macros.
- **Allergies & Preferences:** Strict compliance with user preferences (lactose-free,
gluten-free, vegan, sweet vs. savory cravings).
- **Save to Favorites:** Instant 1-click saving of generated recipes.

### 4. Core feature 3: Longevity Recommendation Engine
- **Insights:** Show the user what crucial vitamins / antioxidants it misses
- **Engine:** Offer suggestions of what foods can be eaten to solve previously
  mentioned insights.
- **Educate:** Educate the user on the effect of what he is eating / not eating
  has on its body.

## Users & Usecases

The main groups of people that would benefit from this application are people who:
- Go to the gym and want to dial down their diet
- Want to lose fat
- Want to prioritize longevity and slow down biological aging
- Don't know what to eat and want science-backed advice instead of influencer trends
