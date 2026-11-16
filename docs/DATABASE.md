# DietNow – Database (ERD)

PostgreSQL 16. The schema is defined only in Go (`backend/internal/models/`) and created by GORM AutoMigrate (`go run ./cmd/migrate`, or at API startup). There are no SQL scripts.

The diagram below was checked against the live schema that `go run ./cmd/migrate` creates (`information_schema` and `pg_indexes`), not only against the Go structs.

## Diagram

Keys: **PK** primary key, **FK** foreign key, **UK** unique. A `?` in the note means the column may be NULL.

```mermaid
erDiagram
    users ||--o| user_profiles : "has (0 or 1)"
    users ||--o{ sessions : "is logged in by"
    users ||--o{ weight_logs : "weighs in"
    users ||--o{ meal_entries : "logs"
    users |o--o{ foods : "owns private"
    foods |o--o{ meal_entries : "is the source of"

    users {
        bigint id PK
        varchar email UK "lower-cased"
        varchar password_hash "bcrypt, never sent to clients"
        varchar display_name "?"
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at "? soft delete"
    }

    user_profiles {
        bigint id PK
        bigint user_id FK, UK "one profile per user"
        varchar gender "male | female"
        date birth_date "? age is derived"
        numeric height_cm
        numeric current_weight_kg
        numeric target_weight_kg "?"
        varchar activity_level "sedentary | light | moderate | very | extra"
        varchar goal "bulk | cut | maintain"
        bigint timeframe_weeks
        varchar timezone "IANA, e.g. Europe/Bucharest"
        numeric bmr "? calculated"
        numeric tdee "? calculated"
        numeric weekly_rate_kg "calculated, + gain or - loss per week"
        numeric daily_adjustment_kcal "calculated, + surplus or - deficit on TDEE"
        numeric target_calories "? calculated"
        numeric target_protein_g "? calculated"
        numeric target_carbs_g "? calculated"
        numeric target_fat_g "? calculated"
        timestamptz created_at
        timestamptz updated_at
    }

    sessions {
        bigint id PK
        bigint user_id FK
        varchar token_hash UK "SHA-256 of the cookie token"
        timestamptz expires_at
        timestamptz last_seen_at
        varchar user_agent "?"
        varchar ip "?"
        timestamptz created_at
    }

    weight_logs {
        bigint id PK
        bigint user_id FK, UK "UK together with logged_on"
        date logged_on UK "one weigh-in per user per day"
        numeric weight_kg "CHECK > 0"
        varchar note "?"
        timestamptz created_at
        timestamptz updated_at
    }

    foods {
        bigint id PK
        varchar name
        varchar brand "?"
        bigint owner_user_id FK "? NULL means global food"
        boolean is_verified
        numeric per100g_calories_kcal
        numeric per100g_protein_g
        numeric per100g_carbs_g
        numeric per100g_fat_g
        numeric per100g_fiber_g
        numeric per100g_sugar_g
        numeric per100g_sodium_mg
        numeric per100g_potassium_mg
        numeric per100g_calcium_mg
        numeric per100g_iron_mg
        numeric per100g_vitamin_aug
        numeric per100g_vitamin_cmg
        numeric per100g_vitamin_dug
        numeric per100g_vitamin_emg
        numeric per100g_vitamin_kug
        numeric per100g_omega3_g
        jsonb antioxidants "? longevity data"
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at "? soft delete"
    }

    meal_entries {
        bigint id PK
        bigint user_id FK
        date consumed_on
        varchar meal_type "breakfast | lunch | dinner | snack"
        bigint food_id FK "? source food"
        numeric portion_grams "CHECK > 0"
        varchar snap_food_name "frozen at logging time"
        numeric snap_calories_kcal "frozen, for portion_grams"
        numeric snap_protein_g "frozen"
        numeric snap_carbs_g "frozen"
        numeric snap_fat_g "frozen"
        timestamptz created_at
        timestamptz updated_at
    }
```

## Relationships

| From | To | Cardinality | Foreign key | When the parent row is deleted |
|---|---|---|---|---|
| users | user_profiles | 1 to 0..1 | `user_profiles.user_id` (unique) | profile deleted (CASCADE) |
| users | sessions | 1 to many | `sessions.user_id` | sessions deleted (CASCADE) |
| users | weight_logs | 1 to many | `weight_logs.user_id` | weigh-ins deleted (CASCADE) |
| users | meal_entries | 1 to many | `meal_entries.user_id` | entries deleted (CASCADE) |
| users | foods | 0..1 to many | `foods.owner_user_id` (nullable) | the user's private foods deleted (CASCADE); global foods have no owner and stay |
| foods | meal_entries | 0..1 to many | `meal_entries.food_id` (nullable) | entries kept, `food_id` set to NULL (SET NULL) |

## Design decisions

**Every user-owned row carries `user_id`.** The repository layer adds `WHERE user_id = <logged-in user>` to every query, so one user can never read or change another user's rows. A request for someone else's row gets 404.

**Global and private foods share one table.** `owner_user_id IS NULL` means a global food everyone can see (the seeded catalog); otherwise the food is private to its owner. One table lets a meal entry point to either kind through a single `food_id`.

**Meal entries freeze their nutrition.** When a portion is logged, the food's name and its nutrition scaled to `portion_grams` are copied into the `snap_*` columns.

Example: 80 g of Rolled Oats (389 kcal / 100 g) is stored with `snap_calories_kcal = 311.2`. If the food is later edited to 400 kcal / 100 g, or deleted, that day's log still says 311.2. The daily totals are sums of the `snap_*` columns, so past days never change.

**Calendar days are `date`, not timestamps.** `consumed_on` and `logged_on` have no time and no time zone.

Example: a user in Bucharest logs breakfast at 01:00 on 6 November. Stored as a UTC timestamp it would be 23:00 on 5 November, and the entry would show up on the wrong day. As a `date` it stays 6 November. The user's `timezone` is used once, to decide which day "today" is.

**Calculated targets are stored on the profile.** BMR, TDEE and the calorie and macro targets are saved when the profile is saved, so the daily dashboard reads them instead of recalculating on every request.

**Soft delete on `users` and `foods`.** Deleting through GORM only sets `deleted_at`; the row stays in the table and normal queries skip it. The CASCADE rules above fire only when a row is really removed from the table.

## Where each table is used

| Table | User Stories |
|---|---|
| users, sessions | US-19 registration and login |
| user_profiles | US-01 to US-06 onboarding and targets, US-20 recalibration |
| foods | US-09 food search, US-21 to US-23 longevity (micronutrients, `antioxidants`) |
| meal_entries | US-08 daily tracker, US-10 portion logging, US-11 dashboard |
| weight_logs | US-12 weigh-ins, US-13 progress chart |
