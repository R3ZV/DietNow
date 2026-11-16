package nutrition

import (
	"errors"
	"testing"
	"time"

	"github.com/dietnow/backend/internal/models"
)

func maleBulk() Input {
	return Input{
		Gender: models.GenderMale, Age: 22, HeightCM: 180, CurrentWeightKG: 70,
		ActivityLevel: models.ActivityModerate, Goal: models.GoalBulk,
		TargetWeightKG: 75, TimeframeWeeks: 10,
	}
}

// US-04: male, 70 kg, 180 cm, 22 y, moderately active -> BMR 1720, TDEE ~2666.
// US-05: TDEE 2666 + 0.5 kg/week -> +400 kcal -> 3066 kcal/day.
func TestBacklogExampleBulk(t *testing.T) {
	got, err := Calculate(maleBulk())
	if err != nil {
		t.Fatal(err)
	}
	if got.BMR != 1720 || got.TDEE != 2666 {
		t.Errorf("BMR/TDEE = %v/%v, want 1720/2666", got.BMR, got.TDEE)
	}
	if got.WeeklyRateKG != 0.5 || got.DailyAdjustmentKcal != 400 || got.TargetCalories != 3066 {
		t.Errorf("rate/adjust/target = %v/%v/%v, want 0.5/400/3066", got.WeeklyRateKG, got.DailyAdjustmentKcal, got.TargetCalories)
	}
	if len(got.Warnings) != 0 {
		t.Errorf("unexpected warnings %v", got.Warnings)
	}
	// Macros add back up to the calorie target (within rounding).
	kcal := got.TargetProteinG*4 + got.TargetCarbsG*4 + got.TargetFatG*9
	if d := kcal - got.TargetCalories; d > 1 || d < -1 {
		t.Errorf("macros sum to %v kcal, target %v", kcal, got.TargetCalories)
	}
}

// US-06: 70 kg, 3000 kcal -> protein 140 g, fat 70 g, carbs ~450 g.
func TestBacklogExampleMacroSplit(t *testing.T) {
	p, c, f := MacroSplit(3000, 70)
	if p != 140 || f != 70 || c != 452.5 {
		t.Errorf("split = P%v C%v F%v, want P140 C452.5 F70", p, c, f)
	}
}

func TestFemaleBMRAndActivityLevels(t *testing.T) {
	// 10*60 + 6.25*165 - 5*30 - 161 = 1320.25
	if got := BMR(models.GenderFemale, 60, 165, 30); got != 1320.25 {
		t.Errorf("female BMR = %v, want 1320.25", got)
	}
	in := Input{Gender: models.GenderFemale, Age: 30, HeightCM: 165, CurrentWeightKG: 60, Goal: models.GoalMaintain}
	want := map[models.ActivityLevel]float64{
		models.ActivitySedentary: 1584, models.ActivityLight: 1815, models.ActivityModerate: 2046,
		models.ActivityVery: 2277, models.ActivityExtra: 2508,
	}
	for lvl, tdee := range want {
		in.ActivityLevel = lvl
		got, err := Calculate(in)
		if err != nil {
			t.Fatal(err)
		}
		if got.TDEE != tdee || got.TargetCalories != tdee || got.DailyAdjustmentKcal != 0 {
			t.Errorf("%s: TDEE %v target %v, want %v (maintain = TDEE)", lvl, got.TDEE, got.TargetCalories, tdee)
		}
	}
}

func TestCutDeficitAndFloor(t *testing.T) {
	in := Input{
		Gender: models.GenderFemale, Age: 28, HeightCM: 168, CurrentWeightKG: 62,
		ActivityLevel: models.ActivityModerate, Goal: models.GoalCut, TargetWeightKG: 58, TimeframeWeeks: 8,
	}
	got, err := Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	// 0.5 kg/week loss -> -550 kcal/day.
	if got.WeeklyRateKG != -0.5 || got.DailyAdjustmentKcal != -550 || got.TargetCalories != got.TDEE-550 {
		t.Errorf("cut: rate %v adjust %v target %v (TDEE %v)", got.WeeklyRateKG, got.DailyAdjustmentKcal, got.TargetCalories, got.TDEE)
	}

	// An aggressive cut is clamped to the safety floor and flagged.
	in.TargetWeightKG, in.TimeframeWeeks, in.ActivityLevel = 50, 4, models.ActivitySedentary
	got, err = Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	if got.TargetCalories != MinCaloriesFemale || len(got.Warnings) != 2 {
		t.Errorf("aggressive cut: target %v warnings %v, want floor %v + pace and floor warnings", got.TargetCalories, got.Warnings, MinCaloriesFemale)
	}
}

// US-03 scenario 2: 10 kg in 2 weeks is accepted with an advisory warning;
// exactly 0.5 kg/week is the boundary and is not flagged.
func TestPaceWarningBoundary(t *testing.T) {
	in := maleBulk()
	in.TargetWeightKG, in.TimeframeWeeks = 80, 2
	got, err := Calculate(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Warnings) != 1 {
		t.Fatalf("expected one pace warning, got %v", got.Warnings)
	}
	in.TargetWeightKG, in.TimeframeWeeks = 72, 4 // exactly 0.5 kg/week
	if got, _ := Calculate(in); len(got.Warnings) != 0 {
		t.Errorf("0.5 kg/week should not warn: %v", got.Warnings)
	}
}

func TestValidationBoundaries(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
		field  string // "" = valid
	}{
		{"valid", func(*Input) {}, ""},
		{"negative weight", func(i *Input) { i.CurrentWeightKG = -10 }, "currentWeightKg"},
		{"600 kg", func(i *Input) { i.CurrentWeightKG = 600 }, "currentWeightKg"},
		{"weight at min", func(i *Input) { i.CurrentWeightKG, i.TargetWeightKG = MinWeightKG, 31 }, ""},
		{"weight just below min", func(i *Input) { i.CurrentWeightKG = MinWeightKG - 0.1 }, "currentWeightKg"},
		{"weight at max target", func(i *Input) { i.CurrentWeightKG, i.TargetWeightKG = 299, MaxWeightKG }, ""},
		{"height at max", func(i *Input) { i.HeightCM = MaxHeightCM }, ""},
		{"height above max", func(i *Input) { i.HeightCM = MaxHeightCM + 0.1 }, "heightCm"},
		{"height missing", func(i *Input) { i.HeightCM = 0 }, "heightCm"},
		{"age at min", func(i *Input) { i.Age = MinAge }, ""},
		{"age below min", func(i *Input) { i.Age = MinAge - 1 }, "birthDate"},
		{"age above max", func(i *Input) { i.Age = MaxAge + 1 }, "birthDate"},
		{"blank gender", func(i *Input) { i.Gender = "" }, "gender"},
		{"bad activity", func(i *Input) { i.ActivityLevel = "couch" }, "activityLevel"},
		{"bad goal", func(i *Input) { i.Goal = "shred" }, "goal"},
		{"bulk target below current", func(i *Input) { i.TargetWeightKG = 65 }, "targetWeightKg"},
		{"bulk target equal current", func(i *Input) { i.TargetWeightKG = 70 }, "targetWeightKg"},
		{"cut target above current", func(i *Input) { i.Goal, i.TargetWeightKG = models.GoalCut, 75 }, "targetWeightKg"},
		{"zero weeks", func(i *Input) { i.TimeframeWeeks = 0 }, "timeframeWeeks"},
		{"too many weeks", func(i *Input) { i.TimeframeWeeks = MaxWeeks + 1 }, "timeframeWeeks"},
		{"maintain ignores target", func(i *Input) { i.Goal, i.TargetWeightKG, i.TimeframeWeeks = models.GoalMaintain, 0, 0 }, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := maleBulk()
			tc.mutate(&in)
			err := in.Validate()
			if tc.field == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			var fe FieldErrors
			if !errors.As(err, &fe) {
				t.Fatalf("want FieldErrors on %q, got %v", tc.field, err)
			}
			if _, ok := fe[tc.field]; !ok {
				t.Fatalf("want error on %q, got %v", tc.field, fe)
			}
		})
	}
}

func TestBulkErrorMessageMatchesBacklog(t *testing.T) {
	in := maleBulk()
	in.TargetWeightKG = 65
	var fe FieldErrors
	errors.As(in.Validate(), &fe)
	if fe["targetWeightKg"] != "for a bulk goal, your target weight must be greater than your current weight" {
		t.Errorf("message = %q", fe["targetWeightKg"])
	}
}

func TestMacroSplitNeverExceedsCalories(t *testing.T) {
	// 150 kg at the 1500 kcal floor: protein+fat alone would be 2550 kcal.
	p, c, f := MacroSplit(1500, 150)
	if c != 0 {
		t.Errorf("carbs = %v, want 0", c)
	}
	if kcal := p*4 + f*9; kcal > 1500.5 || kcal < 1499 {
		t.Errorf("scaled macros = %v kcal, want ~1500", kcal)
	}
}

func TestAgeOn(t *testing.T) {
	birth := models.NewLocalDate(2004, time.March, 14)
	for _, tc := range []struct {
		today models.LocalDate
		want  int
	}{
		{models.NewLocalDate(2026, time.March, 13), 21},
		{models.NewLocalDate(2026, time.March, 14), 22},
		{models.NewLocalDate(2026, time.December, 1), 22},
	} {
		if got := AgeOn(birth, tc.today); got != tc.want {
			t.Errorf("AgeOn(%v) = %d, want %d", tc.today, got, tc.want)
		}
	}
}

// US-10: oatmeal at 375 kcal/100g, 80 g -> 300 kcal.
func TestPortionScalingAndBounds(t *testing.T) {
	got := Scale(models.Macros{CaloriesKcal: 375, ProteinG: 13, CarbsG: 68, FatG: 6.5}, 80)
	want := models.NutritionSnapshot{CaloriesKcal: 300, ProteinG: 10.4, CarbsG: 54.4, FatG: 5.2}
	if got != want {
		t.Errorf("Scale = %+v, want %+v", got, want)
	}
	// Floating-point noise is rounded away: 389 * 0.8 = 311.20000000000005.
	if got := Scale(models.Macros{CaloriesKcal: 389}, 80); got.CaloriesKcal != 311.2 {
		t.Errorf("rounding: %v", got.CaloriesKcal)
	}
	for grams, ok := range map[float64]bool{-5: false, 0: false, 0.1: true, 2000: true, 2000.1: false} {
		if ValidPortion(grams) != ok {
			t.Errorf("ValidPortion(%v) = %v, want %v", grams, !ok, ok)
		}
	}
}
