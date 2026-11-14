package models

// Small, string-backed enums. We store them as short text columns instead of
// integers so the database rows stay human-readable and new values can be
// added without a numeric-mapping migration.

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

// Goal is the user's physique objective; it drives the calorie surplus/deficit.
type Goal string

const (
	GoalBulk     Goal = "bulk"
	GoalCut      Goal = "cut"
	GoalMaintain Goal = "maintain"
)

// ActivityLevel maps to the standard TDEE activity multipliers.
type ActivityLevel string

const (
	ActivitySedentary   ActivityLevel = "sedentary"    // little/no exercise
	ActivityLight        ActivityLevel = "light"        // 1-3 days/week
	ActivityModerate     ActivityLevel = "moderate"     // 3-5 days/week
	ActivityVery         ActivityLevel = "very"         // 6-7 days/week
	ActivityExtra        ActivityLevel = "extra"        // hard training / physical job
)

// MealType corresponds to the four daily tracker sections.
type MealType string

const (
	MealBreakfast MealType = "breakfast"
	MealLunch     MealType = "lunch"
	MealDinner    MealType = "dinner"
	MealSnack     MealType = "snack"
)

// Valid reports whether m is one of the four known meal sections. Use it at the
// service/handler layer before persisting user input.
func (m MealType) Valid() bool {
	switch m {
	case MealBreakfast, MealLunch, MealDinner, MealSnack:
		return true
	default:
		return false
	}
}
