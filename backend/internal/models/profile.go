package models

import (
	"time"
)

// UserProfile holds biometrics and the persisted calorie/macro targets produced
// by the onboarding engine (BMR/TDEE -> surplus/deficit -> macro split).
//
// One-to-one with User: UserID carries a UNIQUE index so a user can have at most
// one profile. Targets are stored (not recomputed on every request) so the
// dashboard is cheap and historical reasoning stays stable even if the formula
// changes later — a fresh calculation overwrites them explicitly.
type UserProfile struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"uniqueIndex;not null" json:"userId"` // one-to-one owner

	// --- Biometrics ---
	Gender          Gender        `gorm:"size:16;not null" json:"gender"`
	BirthDate       LocalDate     `gorm:"type:date" json:"birthDate"` // age is derived, never stored
	HeightCM        float64       `gorm:"not null" json:"heightCm"`
	CurrentWeightKG float64       `gorm:"not null" json:"currentWeightKg"`
	TargetWeightKG  float64       `json:"targetWeightKg"`
	ActivityLevel   ActivityLevel `gorm:"size:16;not null" json:"activityLevel"`
	Goal            Goal          `gorm:"size:16;not null" json:"goal"`
	// TimeframeWeeks is how long the user gives themselves to reach
	// TargetWeightKG (US-03). Zero for a maintenance goal.
	TimeframeWeeks int `gorm:"not null;default:0" json:"timeframeWeeks"`

	// Timezone is the user's IANA zone (e.g. "Europe/Bucharest"). It is the
	// single source of truth for turning "now" into the user's calendar day.
	// See LocalDate / TodayIn. Defaults to UTC so the column is never ambiguous.
	Timezone string `gorm:"size:64;not null;default:UTC" json:"timezone"`

	// --- Persisted targets (output of the calorie engine) ---
	BMR  float64 `json:"bmr"`
	TDEE float64 `json:"tdee"`
	// WeeklyRateKG is the planned weight change per week (+ gain, - loss) and
	// DailyAdjustmentKcal the matching surplus (+) or deficit (-) on TDEE.
	WeeklyRateKG        float64 `gorm:"not null;default:0" json:"weeklyRateKg"`
	DailyAdjustmentKcal float64 `gorm:"not null;default:0" json:"dailyAdjustmentKcal"`
	TargetCalories      float64 `json:"targetCalories"`
	TargetProteinG      float64 `json:"targetProteinG"`
	TargetCarbsG        float64 `json:"targetCarbsG"`
	TargetFatG          float64 `json:"targetFatG"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Location parses the profile's timezone, falling back to UTC on any error so
// callers never have to nil-check. Use it with TodayIn:
//
//	today := models.TodayIn(profile.Location())
func (p UserProfile) Location() *time.Location {
	if p.Timezone == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}
