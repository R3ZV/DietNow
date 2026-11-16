# Authentication & Sessions

DietNow authenticates users **in-house** — no third-party identity provider.

| Concern            | Choice                                                                 |
|--------------------|------------------------------------------------------------------------|
| Password hashing   | **bcrypt** (`golang.org/x/crypto/bcrypt`), cost from `BCRYPT_COST` (default 12) |
| Sessions           | **Server-side sessions** in PostgreSQL (`sessions` table)              |
| Session transport  | Opaque 256-bit random token in an **HttpOnly, SameSite=Lax** cookie    |
| HTTP               | Go standard library `net/http` (`cmd/api`)                             |

**Why not JWT or an external provider?** The acceptance criteria require logout
to invalidate a session. With server-side sessions, logout deletes one row and
the cookie is dead immediately — a stateless JWT stays valid until it expires.
An external provider (Auth0, Clerk, Supabase…) would add an account, network
dependency and callback configuration that a single-app first release does not
need. Swapping later is contained: only `internal/auth` touches identity.

Because there is no OAuth/OIDC provider, there are **no callback URLs**. The
only origin setting is `ALLOWED_ORIGINS` (below).

## Code map

```
internal/models/session.go     Session table (token hash, expiry, last seen)
internal/auth/config.go        env vars + startup validation
internal/auth/password.go      bcrypt hashing, password policy
internal/auth/service.go       Register / Authenticate / Create|Validate|Revoke session
internal/auth/middleware.go    RequireAuth, cookie helpers, UserFrom(ctx)
internal/httpapi/              routes, CORS/origin guard, handlers, integration tests
cmd/api/main.go                server entry point
```

## Identity ↔ application data

There is one identity record: `users`. Its `id` is the owner key for every
user-owned table (`user_profiles`, `weight_logs`, `meal_entries`, private
`foods`, and `sessions`), all with `ON DELETE CASCADE`.

`auth.RequireAuth` resolves the cookie to a `models.User` and stores it on the
request context. Handlers read the owner id **only** via
`auth.UserFrom(r.Context())` and pass it to the repositories, whose `OwnedBy`
scopes make every query `WHERE user_id = <session user>` (see README →
*User ownership & data isolation*). A request for another user's row returns
`404`, never revealing that it exists.

## Passwords

- Only the bcrypt hash is stored (`users.password_hash`, never serialized to
  JSON). bcrypt generates a random salt per hash and embeds it in the hash.
- Policy: 8–72 bytes (US-19 minimum; 72 is bcrypt's input limit — longer input
  is rejected rather than silently truncated).
- Emails are trimmed and lower-cased before storage and lookup.
- Wrong password and unknown email return the same `401 invalid email or
  password`, and both run a bcrypt comparison so timing does not reveal which
  accounts exist.

## Sessions

| Event      | Behaviour |
|------------|-----------|
| Register / login | 32 random bytes → base64url token in the cookie; `sha256(token)` stored in `sessions.token_hash`. Login first revokes any session the browser already carried (no session fixation). |
| Each request | `RequireAuth` looks up the hash, then checks **absolute expiry** (`SESSION_TTL_HOURS`, default 7 days) and **idle timeout** (`SESSION_IDLE_MINUTES`, default 24 h). Expired/idle sessions are deleted. `last_seen_at` is refreshed at most once a minute. |
| Logout | `POST /api/auth/logout` deletes the row and clears the cookie. Idempotent. |
| Cleanup | `cmd/api` purges expired rows hourly. |

Unauthenticated access to a private route returns:

```
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Cache-Control: no-store

{"error":"unauthenticated"}
```

A frontend should treat any `401` as "redirect to login".

## Cross-origin & CSRF

- Cookie is `HttpOnly` (JavaScript cannot read it), `SameSite=Lax` (not sent on
  cross-site POSTs), and `Secure` outside local dev.
- `ALLOWED_ORIGINS` lists the exact frontend origins. Those get CORS headers
  with `Access-Control-Allow-Credentials: true`; any **other** `Origin` is
  refused with `403` on state-changing methods.
- Request bodies must be `application/json` (an HTML form cannot send that
  cross-site), capped at 1 MiB.
- Frontend calls must send credentials, e.g.
  `fetch(url, { credentials: "include", headers: {"Content-Type": "application/json"}, ... })`.
- In staging/production put the frontend and API on the **same site**
  (e.g. `staging.dietnow.app` + `api.staging.dietnow.app`, or one host behind a
  reverse proxy); a `SameSite=Lax` cookie is not sent to a different site.

## API

| Method & path            | Auth | Success | Errors |
|--------------------------|------|---------|--------|
| `POST /api/auth/register` `{email, password, displayName?}` | – | `201 {user, expiresAt}` + cookie | `400` invalid email / password policy, `409` email taken |
| `POST /api/auth/login` `{email, password}` | – | `200 {user, expiresAt}` + cookie | `401` invalid credentials |
| `POST /api/auth/logout`  | – | `204`, cookie cleared | – |
| `GET /api/auth/me`       | ✔ | `200 {user, expiresAt}` | `401` |
| `GET /api/meals?date=YYYY-MM-DD` | ✔ | `200 {date, entries}` | `400`, `401` |
| `DELETE /api/meals/{id}` | ✔ | `204` | `401`, `404` (missing **or not yours**) |
| `GET /api/weights?from=&to=` | ✔ | `200 {entries}` | `400`, `401` |

## Environment variables

Templates: [`.env.example`](.env.example) (local) and
[`.env.staging.example`](.env.staging.example) (staging).

| Variable | Local | Staging | Notes |
|----------|-------|---------|-------|
| `APP_ENV` | `local` | `staging` | non-`local` turns on the strict checks below |
| `HTTP_ADDR` | `:8080` | `:8080` | |
| `ALLOWED_ORIGINS` | `http://localhost:5173,http://localhost:3000` | `https://<staging-frontend>` | comma-separated, exact origins |
| `SESSION_TTL_HOURS` | `168` | `168` | absolute lifetime |
| `SESSION_IDLE_MINUTES` | `1440` | `1440` | inactivity timeout |
| `SESSION_COOKIE_NAME` | `dietnow_session` | `__Host-dietnow_session` | `__Host-` makes the browser enforce Secure + host-only |
| `COOKIE_SECURE` | `false` | `true` | must be `true` over HTTPS |
| `COOKIE_DOMAIN` | *(empty)* | *(empty)* | keep empty = host-only cookie |
| `BCRYPT_COST` | `12` | `12` | 10–31 outside local |
| `DB_*` | see `.env.example` | from secret manager | `DB_SSLMODE=require` on staging |

With `APP_ENV` ≠ `local`, the server **refuses to start** if `COOKIE_SECURE` is
not `true`, `BCRYPT_COST` < 10, or `DB_PASSWORD` is empty/the dev default.


## Running & verifying

```bash
docker compose up -d
go run ./cmd/seed          # optional: test@dietnow.app / Password123!
go run ./cmd/api           # http://localhost:8080
go test ./...              # integration tests need the database running
```

`internal/httpapi/auth_integration_test.go` covers: registration (hash and
hashed session token stored, cookie flags), validation (`400`/`409`/`415`),
valid and invalid login with identical error messages, `401` on every private
route and for forged cookies, server-side logout invalidation (a copied cookie
stops working), absolute and idle expiry, two-user isolation (Bob cannot list
or delete Alice's meal entry; gets `404`), and the cross-origin guard.

Manual check with curl:

```bash
curl -i -c jar -H 'Content-Type: application/json' \
     -d '{"email":"test@dietnow.app","password":"Password123!"}' localhost:8080/api/auth/login
curl -i -b jar localhost:8080/api/auth/me          # 200
curl -i -b jar -X POST localhost:8080/api/auth/logout
curl -i -b jar localhost:8080/api/auth/me          # 401
```

## Not yet implemented

- Login rate limiting / lockout (recommended before public launch; put it in
  front of `POST /api/auth/login`).
- Password reset and email verification.
- "Log out of all devices" (one `DELETE FROM sessions WHERE user_id = ?`).
