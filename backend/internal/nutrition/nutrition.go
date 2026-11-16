// Package nutrition is the calorie engine: biometric validation, BMR/TDEE,
// goal-driven calorie adjustment and the macronutrient split (US-01..US-06).
//
// Everything here is a pure function of its inputs, so it is unit-tested
// without a database and reused unchanged by the HTTP handlers.
package nutrition

import (
	"fmt"
	"math"

	"github.com/dietnow/backend/internal/models"
)

// Validation bounds for biometric intake. Values outside are rejected as
// unrealistic (US-01 scenario 2: -10 kg and 600 kg must fail).
const (
	MinAge, MaxAge           = 13, 100
	MinHeightCM, MaxHeightCM = 100.0, 250.0
	MinWeightKG, MaxWeightKG = 30.0, 300.0
	MinWeeks, MaxWeeks       = 1, 104
)

// Energy cost of a change in body weight, used to turn a weekly rate into a
// daily calorie adjustment.
//
//   - Gaining (bulk): ~5600 kcal per kg of lean-biased gain (muscle + some fat),
//     which gives the +400 kcal/day for 0.5 kg/week required by US-05.
//   - Losing (cut): ~7700 kcal per kg of body fat.
const (
	BulkKcalPerKG = 5600.0
	CutKcalPerKG  = 7700.0
)

// Pace guidance (US-03 scenario 2). Faster rates are accepted but flagged.
const (
	MaxHealthyBulkKGPerWeek = 0.5
	MaxHealthyCutKGPerWeek  = 1.0
)

// Daily calorie floors: a cut never targets less than this.
const (
	MinCaloriesFemale = 1200.0
	MinCaloriesMale   = 1500.0
)

// Macro split rules (US-06): grams per kg of current body weight; carbs take
// the remaining calories.
const (
	ProteinGPerKG = 2.0
	FatGPerKG     = 1.0
	KcalPerGProt  = 4.0
	KcalPerGCarb  = 4.0
	KcalPerGFat   = 9.0
)

var activityMultipliers = map[models.ActivityLevel]float64{
	models.ActivitySedentary: 1.2,
	models.ActivityLight:     1.375,
	models.ActivityModerate:  1.55,
	models.ActivityVery:      1.725,
	models.ActivityExtra:     1.9,
}

// ActivityMultiplier returns the TDEE multiplier for a level.
func ActivityMultiplier(a models.ActivityLevel) (float64, bool) {
	m, ok := activityMultipliers[a]
	return m, ok
}

// Input is the biometric intake plus goal configuration.
type Input struct {
	Gender          models.Gender
	Age             int
	HeightCM        float64
	CurrentWeightKG float64
	ActivityLevel   models.ActivityLevel
	Goal            models.Goal
	TargetWeightKG  float64 // required for bulk/cut; ignored for maintain
	TimeframeWeeks  int     // required for bulk/cut; ignored for maintain
}

// FieldErrors maps a JSON field name to a human-readable problem.
type FieldErrors map[string]string

func (f FieldErrors) Error() string { return fmt.Sprintf("invalid fields: %v", map[string]string(f)) }

// Validate checks the intake against realistic bounds and goal consistency.
// It returns nil or a FieldErrors keyed by the API's JSON field names.
func (in Input) Validate() error {
	errs := FieldErrors{}
	if in.Gender != models.GenderMale && in.Gender != models.GenderFemale {
		errs["gender"] = "must be male or female"
	}
	if in.Age < MinAge || in.Age > MaxAge {
		errs["birthDate"] = fmt.Sprintf("age must be between %d and %d years", MinAge, MaxAge)
	}
	if !between(in.HeightCM, MinHeightCM, MaxHeightCM) {
		errs["heightCm"] = fmt.Sprintf("must be between %g and %g cm", MinHeightCM, MaxHeightCM)
	}
	if !between(in.CurrentWeightKG, MinWeightKG, MaxWeightKG) {
		errs["currentWeightKg"] = fmt.Sprintf("must be between %g and %g kg", MinWeightKG, MaxWeightKG)
	}
	if _, ok := activityMultipliers[in.ActivityLevel]; !ok {
		errs["activityLevel"] = "must be one of sedentary, light, moderate, very, extra"
	}

	switch in.Goal {
	case models.GoalMaintain:
		// Target and timeframe are not used.
	case models.GoalBulk, models.GoalCut:
		if !between(in.TargetWeightKG, MinWeightKG, MaxWeightKG) {
			errs["targetWeightKg"] = fmt.Sprintf("must be between %g and %g kg", MinWeightKG, MaxWeightKG)
		} else if _, bad := errs["currentWeightKg"]; !bad {
			if in.Goal == models.GoalBulk && in.TargetWeightKG <= in.CurrentWeightKG {
				errs["targetWeightKg"] = "for a bulk goal, your target weight must be greater than your current weight"
			}
			if in.Goal == models.GoalCut && in.TargetWeightKG >= in.CurrentWeightKG {
				errs["targetWeightKg"] = "for a cut goal, your target weight must be lower than your current weight"
			}
		}
		if in.TimeframeWeeks < MinWeeks || in.TimeframeWeeks > MaxWeeks {
			errs["timeframeWeeks"] = fmt.Sprintf("must be between %d and %d weeks", MinWeeks, MaxWeeks)
		}
	default:
		errs["goal"] = "must be bulk, cut or maintain"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Targets is the engine output persisted on the profile.
type Targets struct {
	BMR                 float64  `json:"bmr"`
	TDEE                float64  `json:"tdee"`
	WeeklyRateKG        float64  `json:"weeklyRateKg"`        // signed: + gain, - loss
	DailyAdjustmentKcal float64  `json:"dailyAdjustmentKcal"` // signed surplus/deficit
	TargetCalories      float64  `json:"targetCalories"`
	TargetProteinG      float64  `json:"targetProteinG"`
	TargetCarbsG        float64  `json:"targetCarbsG"`
	TargetFatG          float64  `json:"targetFatG"`
	Warnings            []string `json:"warnings"`
}

// BMR is the Mifflin-St Jeor basal metabolic rate in kcal/day.
func BMR(g models.Gender, weightKG, heightCM float64, age int) float64 {
	base := 10*weightKG + 6.25*heightCM - 5*float64(age)
	if g == models.GenderMale {
		return base + 5
	}
	return base - 161
}

// Calculate validates the input and computes all targets. Calories are rounded
// to whole kcal and grams to one decimal.
func Calculate(in Input) (Targets, error) {
	if err := in.Validate(); err != nil {
		return Targets{}, err
	}
	t := Targets{Warnings: []string{}}

	bmr := BMR(in.Gender, in.CurrentWeightKG, in.HeightCM, in.Age)
	mult := activityMultipliers[in.ActivityLevel]
	t.BMR = math.Round(bmr)
	t.TDEE = math.Round(bmr * mult)

	switch in.Goal {
	case models.GoalBulk:
		rate := (in.TargetWeightKG - in.CurrentWeightKG) / float64(in.TimeframeWeeks)
		t.WeeklyRateKG = round(rate, 2)
		t.DailyAdjustmentKcal = math.Round(rate * BulkKcalPerKG / 7)
		if rate > MaxHealthyBulkKGPerWeek {
			t.Warnings = append(t.Warnings, fmt.Sprintf(
				"A gain of %.2f kg/week is faster than recommended; aim for 0.25-0.5 kg/week to prioritise lean muscle over fat.", rate))
		}
	case models.GoalCut:
		rate := (in.CurrentWeightKG - in.TargetWeightKG) / float64(in.TimeframeWeeks)
		t.WeeklyRateKG = -round(rate, 2)
		t.DailyAdjustmentKcal = -math.Round(rate * CutKcalPerKG / 7)
		if rate > MaxHealthyCutKGPerWeek {
			t.Warnings = append(t.Warnings, fmt.Sprintf(
				"A loss of %.2f kg/week is faster than recommended; aim for at most 1 kg/week to preserve muscle.", rate))
		}
	}

	t.TargetCalories = t.TDEE + t.DailyAdjustmentKcal
	floor := MinCaloriesFemale
	if in.Gender == models.GenderMale {
		floor = MinCaloriesMale
	}
	if t.TargetCalories < floor {
		t.TargetCalories = floor
		t.Warnings = append(t.Warnings, fmt.Sprintf(
			"Your daily target was raised to the %.0f kcal safety minimum; consider a longer timeframe.", floor))
	}

	t.TargetProteinG, t.TargetCarbsG, t.TargetFatG = MacroSplit(t.TargetCalories, in.CurrentWeightKG)
	return t, nil
}

// MacroSplit allocates protein and fat per kg of body weight and gives the
// remaining calories to carbohydrates. If protein and fat alone would exceed
// the calorie budget, both are scaled down proportionally and carbs are 0, so
// the macros always add up to the calorie target.
func MacroSplit(calories, weightKG float64) (proteinG, carbsG, fatG float64) {
	proteinG = ProteinGPerKG * weightKG
	fatG = FatGPerKG * weightKG
	pfKcal := proteinG*KcalPerGProt + fatG*KcalPerGFat
	if pfKcal > calories {
		scale := calories / pfKcal
		return round(proteinG*scale, 1), 0, round(fatG*scale, 1)
	}
	carbsG = (calories - pfKcal) / KcalPerGCarb
	return round(proteinG, 1), round(carbsG, 1), round(fatG, 1)
}

// AgeOn returns the age in whole years on a given day.
func AgeOn(birth, today models.LocalDate) int {
	age := today.Year - birth.Year
	if today.Month < birth.Month || (today.Month == birth.Month && today.Day < birth.Day) {
		age--
	}
	return age
}

// Portion bounds for logging a food (US-10).
const (
	MinPortionG = 0.0 // exclusive
	MaxPortionG = 2000.0
)

// ValidPortion reports whether grams is a loggable portion: > 0 and <= 2000.
func ValidPortion(grams float64) bool {
	return !math.IsNaN(grams) && grams > MinPortionG && grams <= MaxPortionG
}

// Scale returns per-100g macros scaled to a portion, rounded to one decimal.
func Scale(per100 models.Macros, grams float64) models.NutritionSnapshot {
	k := grams / 100
	return models.NutritionSnapshot{
		CaloriesKcal: round(per100.CaloriesKcal*k, 1),
		ProteinG:     round(per100.ProteinG*k, 1),
		CarbsG:       round(per100.CarbsG*k, 1),
		FatG:         round(per100.FatG*k, 1),
	}
}

// Round rounds v to the given number of decimals.
func Round(v float64, decimals int) float64 { return round(v, decimals) }

func round(v float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals))
	return math.Round(v*p) / p
}

func between(v, lo, hi float64) bool {
	return !math.IsNaN(v) && v >= lo && v <= hi
}
