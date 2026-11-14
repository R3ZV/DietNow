# DietNow — Backend Persistence Layer

Go + GORM + PostgreSQL persistent storage for DietNow's first release. No raw SQL
scripts: all schema, migrations, constraints, and seeding are defined in Go via
GORM.

## Layout

```
backend/
├── docker-compose.yml          # local PostgreSQL with a persistent volume
├── .env.example                # copy to .env
├── go.mod
├── cmd/
│   ├── migrate/main.go         # `go run ./cmd/migrate`  -> AutoMigrate
│   └── seed/main.go            # `go run ./cmd/seed`      -> migrate + seed
└── internal/
    ├── models/                 # GORM structs & relations
    │   ├── enums.go            # Gender, Goal, ActivityLevel, MealType
    │   ├── localdate.go        # timezone-safe calendar DATE type
    │   ├── user.go             # identity
    │   ├── profile.go          # biometrics + persisted calorie/macro targets
    │   ├── weightlog.go        # historical weigh-ins
    │   ├── food.go             # per-100g macro/micro profiles (global + private)
    │   └── mealentry.go        # 4-meal tracker with nutrition snapshots
    ├── database/
    │   ├── database.go         # pooled connection from env
    │   ├── migrate.go          # AutoMigrate
    │   └── seed.go             # SeedDatabase(db) — idempotent
    └── repository/             # ownership-enforcing data access
        ├── scopes.go           # OwnedBy / VisibleFoodsFor GORM scopes
        ├── meal_entry_repository.go
        ├── food_repository.go
        └── weight_log_repository.go
```

## Quick start

```bash
cd backend
cp .env.example .env            # Windows PowerShell: Copy-Item .env.example .env

docker compose up -d            # start PostgreSQL (data persists in volume dietnow_pgdata)

go mod tidy                     # resolve dependencies
go run ./cmd/seed               # migrate the schema + insert the test dataset
```

Seed login: `test@dietnow.app` / `Password123!`.

Stop without losing data: `docker compose down`.
Wipe everything and start fresh: `docker compose down -v`.

## Data model & relations

| Entity        | Owns / relates to                                   | Key points |
|---------------|-----------------------------------------------------|------------|
| `User`        | 1–1 `UserProfile`, 1–N `WeightLog`/`MealEntry`/private `Food` | identity; deleting cascades to owned rows |
| `UserProfile` | belongs to `User` (unique `user_id`)                | height (cm), weights (kg), **IANA timezone**, persisted BMR/TDEE/calorie & macro targets |
| `WeightLog`   | belongs to `User`                                   | unique `(user_id, logged_on)` → one weigh-in/day |
| `Food`        | optional owner `User`                               | per-100g macros+micros; `owner_user_id IS NULL` ⇒ global, else private; JSONB `antioxidants` for the longevity engine |
| `MealEntry`   | belongs to `User`, references `Food` (nullable)     | `consumed_on` DATE, `meal_type`, `portion_grams`, **frozen nutrition snapshot** |

### Historical integrity (nutrition snapshots)

A `MealEntry` keeps a reference to its source `Food` (`food_id`, nullable with
`ON DELETE SET NULL`) **and** a frozen copy of the food name and scaled macros
(`snap_food_name`, `snap_calories_kcal`, …). The dashboard sums the snapshot
columns, so a past log never changes if the food is later edited, and it survives
the food being deleted. `models.NewMealEntry(...)` builds the reference and the
snapshot together so they can't disagree.

## Date handling — how `consumed_on` / `logged_on` avoid timezone bugs

**Rule: calendar days are stored as a zone-less `DATE`, never as a timestamp.**

The classic "my log jumped a day" bug comes from storing a wall-clock date in a
`timestamptz`/`time.Time`: a user in `Europe/Bucharest` (UTC+2) logging at
`2026-11-06 01:00` has that value converted to `2026-11-05 23:00Z`, and the day
silently rolls back. We avoid this with `models.LocalDate` (`internal/models/localdate.go`):

- It carries **only** year/month/day — no clock, no zone.
- It maps to a PostgreSQL `DATE` column (`GormDBDataType` returns `"date"`), so
  the DB stores it verbatim with no conversion.
- It serializes to/from the bare string `"2006-01-02"` (JSON and SQL both).

A timezone is applied in exactly **one** place: deciding which day "now" is for a
given user, using the IANA zone saved on their profile.

**Define today for a user**

```go
profile, _ := profileRepo.Get(ctx, userID)      // profile.Timezone = "Europe/Bucharest"
today := models.TodayIn(profile.Location())      // LocalDate{2026, 11, 6}
```

`UserProfile.Location()` parses `Timezone`, falling back to UTC on error.
For an arbitrary instant use `models.DateIn(t, loc)`; parse client input with
`models.ParseLocalDate("2026-11-06")`.

**Query a single day / a range** — plain, zone-free DATE comparisons:

```go
// All of today's entries for a user (hits the (user_id, consumed_on) index):
entries, _ := mealRepo.ListForDay(ctx, userID, today)

// Last 7 days of weigh-ins:
history, _ := weightRepo.History(ctx, userID, today.AddDays(-6), today)
```

Because both sides of the comparison are bare dates, there is no offset to get
wrong — no `AT TIME ZONE`, no half-open `[00:00, next-00:00)` timestamp windows.

## User ownership & data isolation

Isolation is enforced in the **repository layer**, not left to callers. Two
reusable GORM scopes live in `repository/scopes.go`:

```go
func OwnedBy(userID uint) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB { return db.Where("user_id = ?", userID) }
}

func VisibleFoodsFor(userID uint) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("owner_user_id IS NULL OR owner_user_id = ?", userID)
    }
}
```

Every read/update/delete threads the authenticated `userID`:

- **Read**: `db.Scopes(OwnedBy(userID)).Where("consumed_on = ?", day).Find(&entries)`
- **Update / Delete**: the `WHERE` carries **both** `id` and `user_id`. A request
  for another user's row updates/deletes **zero** rows → the repo returns
  `ErrNotFound` (we never reveal whether the row exists but belongs to someone
  else). See `MealEntryRepository.UpdatePortion` / `.Delete`.
- **Foods**: `GetVisible`/`Search` use `VisibleFoodsFor` so private foods never
  leak; `CreatePrivate` forces `owner_user_id` server-side so a client can't
  forge a global food, and `UpdatePrivate`/`DeletePrivate` require
  `owner_user_id = userID` so the shared catalog is untouchable.

> The authenticated user id must come from the session/JWT, **never** from the
> request body. The repositories assume the `userID` argument is trusted.

## Seeding

`database.SeedDatabase(db)` (run via `go run ./cmd/seed`) inserts a test user, a
biometric profile, a global food catalog, one private food, and today's meal
entries. It is **idempotent and restart-safe**: users/foods match on natural keys
with `FirstOrCreate`, the profile upserts on its unique `user_id`, and the day's
meal entries are reset-then-recreated. Running it repeatedly never duplicates
data, and the data persists across container restarts via the Docker volume.

## Dependencies

```
gorm.io/gorm, gorm.io/driver/postgres, gorm.io/datatypes, golang.org/x/crypto/bcrypt
```

Run `go mod tidy` once to populate `go.sum` and indirect requirements.
