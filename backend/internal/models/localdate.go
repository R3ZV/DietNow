package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// dateLayout is the only textual representation we ever use for a calendar day.
const dateLayout = "2006-01-02"

// LocalDate is a timezone-free calendar date (year / month / day) with NO time
// and NO location component. It maps to a PostgreSQL `DATE` column.
//
// Why a dedicated type instead of time.Time?
//
// The classic "my log jumped a day" bug happens when a wall-clock date is stored
// as a `timestamptz`/`time.Time`: e.g. a user in Bucharest (UTC+2) logs breakfast
// on 2026-11-06. If we naively store time.Now() = 2026-11-06T01:00+02:00 and the
// driver converts it to UTC, it becomes 2026-11-05T23:00Z — the day the user
// "ate" silently rolls back to the 5th. Reading it back in UTC (or any other
// zone) gives the wrong calendar day.
//
// LocalDate sidesteps this entirely: it carries only Y-M-D, serializes as the
// bare string "2006-01-02", and is stored in a `DATE` column that has no zone.
// The only place a timezone is ever applied is when we DECIDE which calendar day
// "now" falls on for a given user — done explicitly via TodayIn(loc), using the
// timezone saved on the user's profile. After that decision, the date is inert.
type LocalDate struct {
	Year  int
	Month time.Month
	Day   int
}

// NewLocalDate constructs a LocalDate from its parts.
func NewLocalDate(year int, month time.Month, day int) LocalDate {
	return LocalDate{Year: year, Month: month, Day: day}
}

// TodayIn returns the calendar date that "now" falls on in the given location.
// This is the ONE sanctioned way to turn the current instant into a user-facing
// day. Pass the location parsed from the user's profile timezone, e.g.:
//
//	loc, _ := time.LoadLocation(profile.Timezone) // "Europe/Bucharest"
//	today := models.TodayIn(loc)
func TodayIn(loc *time.Location) LocalDate {
	if loc == nil {
		loc = time.UTC
	}
	n := time.Now().In(loc)
	return LocalDate{Year: n.Year(), Month: n.Month(), Day: n.Day()}
}

// DateIn converts an arbitrary instant into the calendar day it lands on in loc.
func DateIn(t time.Time, loc *time.Location) LocalDate {
	if loc == nil {
		loc = time.UTC
	}
	t = t.In(loc)
	return LocalDate{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}

// ParseLocalDate parses a "2006-01-02" string (the format the React client sends).
func ParseLocalDate(s string) (LocalDate, error) {
	t, err := time.Parse(dateLayout, strings.TrimSpace(s))
	if err != nil {
		return LocalDate{}, fmt.Errorf("invalid date %q, want YYYY-MM-DD: %w", s, err)
	}
	return LocalDate{Year: t.Year(), Month: t.Month(), Day: t.Day()}, nil
}

// String renders the date as "2006-01-02".
func (d LocalDate) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, int(d.Month), d.Day)
}

// IsZero reports whether the date is the zero value.
func (d LocalDate) IsZero() bool { return d == LocalDate{} }

// AddDays returns the date shifted by n days (n may be negative). It normalizes
// overflow (e.g. Jan 32 -> Feb 1) using time arithmetic in UTC, which is safe
// because no zone boundary is ever crossed for a pure date.
func (d LocalDate) AddDays(n int) LocalDate {
	t := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, n)
	return LocalDate{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}

// --- JSON (de)serialization: always the bare "YYYY-MM-DD" string ---

func (d LocalDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *LocalDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*d = LocalDate{}
		return nil
	}
	parsed, err := ParseLocalDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// --- database/sql integration: map to a PostgreSQL DATE column ---

// Value implements driver.Valuer. We hand the driver a bare "YYYY-MM-DD" string,
// which PostgreSQL stores in a zone-less DATE column verbatim — no conversion.
func (d LocalDate) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.String(), nil
}

// Scan implements sql.Scanner. PostgreSQL's DATE comes back from the pgx/lib-pq
// drivers as a time.Time (midnight) or as a string/[]byte depending on config;
// we handle all three and keep only the Y-M-D fields, discarding any zone.
func (d *LocalDate) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*d = LocalDate{}
		return nil
	case time.Time:
		*d = LocalDate{Year: v.Year(), Month: v.Month(), Day: v.Day()}
		return nil
	case string:
		return d.scanString(v)
	case []byte:
		return d.scanString(string(v))
	default:
		return fmt.Errorf("cannot scan %T into LocalDate", src)
	}
}

func (d *LocalDate) scanString(s string) error {
	// Some drivers append a time part (e.g. "2026-11-06T00:00:00Z"); keep the day.
	if len(s) >= 10 {
		s = s[:10]
	}
	parsed, err := ParseLocalDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// --- GORM integration: force the column type to DATE on AutoMigrate ---

// GormDBDataType tells GORM to create a `date` column regardless of dialect.
func (LocalDate) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "date"
}
