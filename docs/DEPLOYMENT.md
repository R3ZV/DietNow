# DietNow – CI, Staging & Release Procedure

This document covers how a commit becomes a verified staging release: what
CI checks, where staging is hosted, how it is configured, how database
migrations are applied, and how to deploy and roll back.

```
 pull request ──► CI (lint, types, unit + integration tests, image + smoke) ──► review ──► merge
                                                                                              │
 master push ──► CI ──green──► Deploy staging: build image ghcr.io/r3zv/dietnow:<sha>         │
                                 ├─ Render deploy hook (exact image)                          ◄┘
                                 ├─ wait until /api/health reports <sha>
                                 ├─ /api/health/ready (database up)
                                 ├─ smoke test with the non-production account
                                 └─ on failure: redeploy the previous <sha> (automatic rollback)
```

## 1. Continuous integration

Workflow: [`.github/workflows/ci.yml`](../.github/workflows/ci.yml). Runs on
every pull request and on every push to `master`.

| Job | What it does |
|-----|--------------|
| **Backend (Go)** | `go mod download && go mod verify` (locked by `go.sum`, `-mod=readonly`), `gofmt`, `go vet`, build, `go test -race` with a **PostgreSQL 16 service**. `DIETNOW_REQUIRE_DB=1` makes the integration tests fail rather than skip if the database is missing. |
| **Frontend (React)** | `npm ci` (locked by `package-lock.json`), ESLint, `tsc --noEmit`, Vitest, production build. |
| **Release image** | Builds the same `Dockerfile` that staging runs, starts it against PostgreSQL and runs the **smoke test** (`cmd/smoke`) against it, then checks the SPA is served. |

The MVP is covered at three levels:

- **Unit** (`backend/internal/nutrition`): the backlog examples (US-04 BMR 1720 /
  TDEE 2666, US-05 +400 kcal → 3066, US-06 140 g P / 70 g F / ~450 g C), each
  activity level, cut deficit and calorie floor, pace warnings at the 0.5 kg/week
  boundary, every validation bound (just inside / just outside), and portion
  scaling (US-10: 80 g of 375 kcal/100 g = 300 kcal).
- **Integration** (`backend/internal/httpapi/mvp_integration_test.go`, real
  PostgreSQL over HTTP):
  - the full flow: register → intake → targets → food search → log portion →
    totals → delete → logout;
  - **user isolation**: a second user cannot read the first user's profile,
    entries or dashboard, cannot find or log their private food, and cannot
    delete their entries (`404`), and the second user's activity never changes
    the first user's totals or targets;
  - **boundaries**: -10 kg / 600 kg / missing fields / future birth date /
    under-age / bad time zone / conflicting bulk/cut targets are rejected
    with per-field errors and nothing is stored, too-fast pace returns a
    warning, maintenance removes the surplus; portions of 0, -5 and 2000.1 g,
    unknown meal types, bad dates and unknown foods are rejected; 0.1 g and
    2000 g are accepted; subtotals, daily totals, negative "remaining" when over
    budget, other days unaffected, delete restores totals; LIKE wildcards in
    search are literal.
- **Smoke** (`backend/internal/smoke`): the same journey against any URL. It
  runs in the integration tests, against the built image in CI, and against
  staging after every deploy.

**Branch protection (recommended, GitHub → Settings → Branches → `master`):**
require the `Backend (Go)`, `Frontend (React)` and `Release image` checks and
one review before merging.

## 2. Staging hosting

| Piece | Where | Notes |
|-------|-------|-------|
| Web service `dietnow-staging` | [Render](https://render.com), Docker image runtime, region Frankfurt | URL `https://dietnow-staging.onrender.com` (Render may add a suffix if the name is taken; the real URL is in the dashboard and in the `STAGING_URL` variable). |
| Database `dietnow-staging-db` | Render PostgreSQL 16, same region | Private network only (`ipAllowList: []`). |
| Release images | GitHub Container Registry `ghcr.io/r3zv/dietnow:<git sha>` | One immutable image per commit; `:master` = latest master build. |
| Definition | [`render.yaml`](../render.yaml) (Render Blueprint) | Infrastructure as code; contains no secrets. |

The image ([`Dockerfile`](../Dockerfile)) contains the Go API, the `migrate`
tool and the built frontend. The API serves the frontend for every non-`/api`
path, so the browser, the SPA and the API share one HTTPS origin. That is what
lets the `__Host-`, `Secure`, `SameSite=Lax` session cookie work with no CORS.

Render routes traffic to a new deploy only after `GET /api/health/ready`
returns `200`. A release that cannot start, migrate or reach the database
therefore never receives traffic, and the previous one keeps serving.

**Free-tier limits.** The free web service sleeps after ~15 minutes idle (the
first request then takes ~1 minute; the deploy workflow waits for it). The free
database expires 30 days after creation and has no backups. Upgrade the
database plan, or recreate it from the Blueprint, before then.

## 3. Runtime configuration

All configuration is environment variables. On staging they come from
`render.yaml` and Render; locally from `backend/.env` (see
[`backend/.env.example`](../backend/.env.example)).
[`backend/.env.staging.example`](../backend/.env.staging.example) documents the
effective staging values.

| Variable | Staging value | Source |
|----------|---------------|--------|
| `APP_ENV` | `staging` (enables strict startup checks) | render.yaml |
| `PORT` | `10000` | injected by Render (`API_PORT` is the local equivalent) |
| `STATIC_DIR` | `/app/web` | baked into the image |
| `DB_AUTO_MIGRATE` | `true` | render.yaml |
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER` | managed database | `fromDatabase` in render.yaml |
| `DB_PASSWORD` | generated by Render | `fromDatabase` (**secret**, never leaves Render) |
| `DB_SSLMODE` | `prefer` | render.yaml |
| `DB_MAX_OPEN_CONNS` | `10` (free database allows ~97 connections) | render.yaml |
| `API_ALLOWED_ORIGINS` | *(empty: same origin)* | render.yaml |
| `COOKIE_SECURE` / `SESSION_COOKIE_NAME` / `COOKIE_DOMAIN` | `true` / `__Host-dietnow_session` / *(empty)* | render.yaml |
| `SESSION_TTL_HOURS` / `SESSION_IDLE_MINUTES` / `BCRYPT_COST` | `168` / `1440` / `12` | render.yaml |

With `APP_ENV=staging` the API **refuses to start** when cookies are not
secure, `BCRYPT_COST < 10`, `DB_PASSWORD` is empty or a dev placeholder, or the
database is unreachable. A misconfiguration therefore fails the deploy instead
of producing an insecure release.

To change a value: edit `render.yaml` in a PR (non-secret values) and apply it
in Render → Blueprints → *Sync*. Or set it in Render → `dietnow-staging` →
Environment, which triggers a restart. Keep `render.yaml` in sync so the
change is not lost.

## 4. Secrets

No secret is ever committed. `.gitignore` excludes every `.env*` file except
the `*.example` templates, and `.dockerignore` keeps them out of images.

| Secret | Stored in | Used by |
|--------|-----------|---------|
| Database password / connection | **Render** (generated with the database) | the web service, via `fromDatabase` |
| `RENDER_DEPLOY_HOOK_URL` | **GitHub → Settings → Environments → `staging` → secrets** | deploy workflow (triggers deploys and rollbacks) |
| `SMOKE_EMAIL`, `SMOKE_PASSWORD` | **GitHub environment `staging` secrets** | deploy workflow's smoke test |
| GHCR push | built-in `GITHUB_TOKEN` (`packages: write`) | deploy workflow |
| GHCR pull (only if the package is private) | **Render → Account → Registry credentials** `ghcr-dietnow` (a GitHub PAT with `read:packages`) | Render pulling the image |

Non-secret settings for the workflow are **environment variables** on the same
GitHub environment: `STAGING_URL` (for example
`https://dietnow-staging.onrender.com`, no trailing slash). The repository
variable `STAGING_DEPLOY_ENABLED=true` turns on automatic deploys.

The smoke account is a **dedicated non-production account** (for example
`smoke+staging@<team domain>`) with a long random password. It only ever owns
smoke-test data: each run overwrites its profile and deletes the meal entry it
logs. Never reuse a personal account.

**Rotation.**
- Deploy hook: Render → service → Settings → *Deploy Hook* → *Regenerate*,
  then update `RENDER_DEPLOY_HOOK_URL`.
- Smoke account: set a new `SMOKE_EMAIL` and `SMOKE_PASSWORD`. The next run
  registers the new account; the old one can be deleted from the database.
- Database credentials: they are wired in by `fromDatabase`, so after any
  credential change on Render's side, trigger a deploy (Actions → *Deploy
  staging* with the current SHA) so the service restarts with the new values.

## 5. Database migrations

**How they are applied.** The schema is defined by the GORM models
(`backend/internal/models`) and applied by `database.Migrate` (GORM
AutoMigrate). Reference data (the verified global food catalog) is applied by
`database.SeedFoodCatalog`. Both are idempotent.

On staging they run **automatically at startup of every release**, before the
HTTP server starts listening (`DB_AUTO_MIGRATE=true`):

1. Render starts the new container.
2. `cmd/api` connects, runs the migration and the catalog seed.
3. On failure the process exits with an error → `/api/health/ready` never
   turns green → Render aborts the deploy and the old release keeps serving.
4. On success the server starts and Render switches traffic.

The same code is available as a standalone command, `/app/migrate` in the
image or `go run ./cmd/migrate` locally. Use it to migrate without starting the
API, e.g. as a Render *pre-deploy command* on a paid plan (then set
`DB_AUTO_MIGRATE=false`).

Demo data (`go run ./cmd/seed`, user `test@dietnow.app`) is for local
development only and is **never** run on staging.

**Rules for schema changes** (because a rollback deploys old code but does
not undo the schema):

- AutoMigrate only *adds* tables, columns and indexes. It never drops or
  renames anything, so the previous release keeps working on the new schema.
- Make every change backwards compatible: add nullable columns or columns with
  defaults, and keep reading old columns.
- To remove or rename something, use two releases: first stop using it, then
  delete it in a later release with an explicit, reviewed SQL step.
- Before any destructive change, take a backup:
  `pg_dump "<external connection string from Render>" > staging-$(date +%F).sql`
  (temporarily allow your IP on the database's *Access Control* page).

## 6. First-time setup (once)

1. **Build the first image.** GitHub → Actions → *Deploy staging* → *Run
   workflow* on `master` with **build_only = true**. This pushes
   `ghcr.io/r3zv/dietnow:<sha>` and `:master`.
2. **Package visibility.** GitHub → your profile/org → Packages → `dietnow`.
   Either set it to *Public*, or add the Render registry credential
   `ghcr-dietnow` and uncomment `creds` in `render.yaml`.
3. **Create the environment on Render.** Dashboard → *New* → *Blueprint* →
   connect `R3ZV/DietNow` → it reads `render.yaml` and creates
   `dietnow-staging` and `dietnow-staging-db`. Wait until the service shows
   *Live*.
4. **Check it:** `curl https://<service>.onrender.com/api/health/ready` →
   `{"status":"ok","database":"up",...}`.
5. **Deploy hook.** Render → `dietnow-staging` → Settings → *Deploy Hook* →
   copy the URL.
6. **GitHub environment.** Settings → Environments → *New environment*
   `staging`:
   - secrets: `RENDER_DEPLOY_HOOK_URL`, `SMOKE_EMAIL`, `SMOKE_PASSWORD`;
   - variable: `STAGING_URL` = `https://<service>.onrender.com`;
   - optional: *Deployment branches* = `master` only, and required reviewers
     if deploys should be approved.
7. **Enable automatic deploys.** Settings → Secrets and variables → Actions →
   *Variables* → `STAGING_DEPLOY_ENABLED` = `true`.
8. **Run the full pipeline once.** Actions → *Deploy staging* → *Run workflow*
   (leave the inputs empty). The smoke account is registered on this first
   run.

## 7. Deploying

**Automatic:** merge a PR into `master`. When CI is green, *Deploy staging*
runs and does the following:

1. Builds `ghcr.io/r3zv/dietnow:<sha>` from that exact commit (`npm ci`,
   locked Go modules, version stamped into the binary), or reuses the image if
   it already exists.
2. Records the version currently serving (`GET /api/health` → `version`) as
   the rollback target.
3. Calls the Render deploy hook with `imgURL=ghcr.io/r3zv/dietnow:<sha>`.
4. Waits until `/api/health` reports the new `<sha>`.
5. Checks `/api/health/ready` (`database: up`).
6. Runs the smoke test with the non-production account and
   `-expect-version <sha>`.
7. Writes a summary (deployed SHA, URL, previous SHA) on the run page.

Deployments are serialized (`concurrency: deploy-staging`), so two merges
never deploy at the same time.

**Manual:** Actions → *Deploy staging* → *Run workflow*, with `sha` empty
(latest master) or a specific commit SHA.

**Verify by hand** (useful after any manual change on Render):

```bash
curl -s https://<service>.onrender.com/api/health        # {"status":"ok","version":"<sha>",...}
curl -s https://<service>.onrender.com/api/health/ready  # {"status":"ok","database":"up",...}
cd backend
SMOKE_EMAIL=... SMOKE_PASSWORD=... go run ./cmd/smoke -base-url https://<service>.onrender.com
```

The smoke test prints one ✓ line per step and exits non-zero on the first
failure.

## 8. Rolling back

| Situation | What happens / what to do |
|-----------|---------------------------|
| New release fails to start, migrate or become ready | Render never switches traffic; the old release keeps serving. The workflow times out waiting for the new version and redeploys the previous SHA as well, so Render's target is the known-good image again. |
| New release is live but the smoke test fails | The workflow **automatically redeploys the previously verified SHA** and waits until it serves again. The run is marked failed and its summary says *Rolled back to `<sha>`*. |
| A problem is found later | Actions → *Deploy staging* → *Run workflow* with `sha` = the last good commit (shown in the previous run's summary, or `git log`). That exact image is redeployed and verified, with no rebuild. |
| GitHub Actions unavailable | Render → `dietnow-staging` → *Events* → pick the last good deploy → *Rollback*. |

Rollback restores the **application**, not the database: migrations are
additive and the previous release runs fine on the newer schema (see §5). If
data itself must be restored, use the `pg_dump` backup from before the
change (`psql "<connection string>" < backup.sql`), or Render's point-in-time
recovery on paid plans.

After any rollback, fix forward on `master` through a normal PR. The next
green merge deploys it.

## 9. Troubleshooting

| Symptom | Likely cause / fix |
|---------|--------------------|
| *Configure the 'staging' environment in GitHub first* | A secret or `STAGING_URL` is missing (§6 step 6). |
| Deploy hook returns 4xx | Hook URL regenerated or wrong; update `RENDER_DEPLOY_HOOK_URL`. |
| Render: *failed to pull image* | Package private and no registry credential (§6 step 2), or the SHA was never built (run the workflow, which builds it). |
| Service logs `config: COOKIE_SECURE must be true …` | An env var was changed on Render; restore the `render.yaml` values. |
| Service logs `connect database: …` | Database suspended or expired (free plan, 30 days) or credentials rotated; check the database in Render. |
| Smoke: `login … status 401` after a password change | The smoke account exists with the old password; delete it on staging or use a new `SMOKE_EMAIL`. |
| Waits forever on the first request | Free instance waking up; the workflow tolerates up to 15 minutes. |
