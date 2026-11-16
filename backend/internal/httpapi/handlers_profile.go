package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dietnow/backend/internal/auth"
	"github.com/dietnow/backend/internal/models"
	"github.com/dietnow/backend/internal/nutrition"
	"github.com/dietnow/backend/internal/repository"
)

// profileRequest is the biometric intake + goal form (US-01..US-03).
type profileRequest struct {
	Gender          models.Gender        `json:"gender"`
	BirthDate       string               `json:"birthDate"` // YYYY-MM-DD
	HeightCM        float64              `json:"heightCm"`
	CurrentWeightKG float64              `json:"currentWeightKg"`
	ActivityLevel   models.ActivityLevel `json:"activityLevel"`
	Goal            models.Goal          `json:"goal"`
	TargetWeightKG  float64              `json:"targetWeightKg"`
	TimeframeWeeks  int                  `json:"timeframeWeeks"`
	Timezone        string               `json:"timezone"` // IANA zone, default UTC
}

type profileResponse struct {
	Profile models.UserProfile `json:"profile"`
	// Warnings are advisory (e.g. a too-fast pace); the profile is still saved.
	Warnings []string `json:"warnings"`
}

// GET /api/profile
func (s *server) getProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	p, err := s.profiles.Get(r.Context(), user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "profile_not_found", "complete onboarding first")
		return
	}
	if err != nil {
		s.internalError(w, "get profile", err)
		return
	}
	WriteJSON(w, http.StatusOK, profileResponse{Profile: p, Warnings: []string{}})
}

// PUT /api/profile validates the intake, runs the calorie engine and stores
// both the biometrics and the resulting targets.
func (s *server) putProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	var in profileRequest
	if !decodeJSON(w, r, &in) {
		return
	}

	fields := map[string]string{}
	tz := strings.TrimSpace(in.Timezone)
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "Local" {
		fields["timezone"] = "must be an IANA time zone such as Europe/Bucharest"
		loc = time.UTC
	}
	birth, err := models.ParseLocalDate(in.BirthDate)
	if err != nil {
		fields["birthDate"] = "required, format YYYY-MM-DD"
	}

	calc := nutrition.Input{
		Gender:          in.Gender,
		HeightCM:        in.HeightCM,
		CurrentWeightKG: in.CurrentWeightKG,
		ActivityLevel:   in.ActivityLevel,
		Goal:            in.Goal,
		TargetWeightKG:  in.TargetWeightKG,
		TimeframeWeeks:  in.TimeframeWeeks,
	}
	if in.Goal == models.GoalMaintain {
		calc.TargetWeightKG, calc.TimeframeWeeks = in.CurrentWeightKG, 0
	}
	if _, bad := fields["birthDate"]; !bad {
		calc.Age = nutrition.AgeOn(birth, models.TodayIn(loc))
	} else {
		calc.Age = nutrition.MinAge // birthDate already reported; avoid a duplicate message
	}

	targets, err := nutrition.Calculate(calc)
	var fe nutrition.FieldErrors
	if errors.As(err, &fe) {
		for k, v := range fe {
			if _, exists := fields[k]; !exists {
				fields[k] = v
			}
		}
	} else if err != nil {
		s.internalError(w, "calculate targets", err)
		return
	}
	if len(fields) > 0 {
		WriteValidationError(w, fields)
		return
	}

	p := models.UserProfile{
		Gender:              calc.Gender,
		BirthDate:           birth,
		HeightCM:            calc.HeightCM,
		CurrentWeightKG:     calc.CurrentWeightKG,
		TargetWeightKG:      calc.TargetWeightKG,
		ActivityLevel:       calc.ActivityLevel,
		Goal:                calc.Goal,
		TimeframeWeeks:      calc.TimeframeWeeks,
		Timezone:            tz,
		BMR:                 targets.BMR,
		TDEE:                targets.TDEE,
		WeeklyRateKG:        targets.WeeklyRateKG,
		DailyAdjustmentKcal: targets.DailyAdjustmentKcal,
		TargetCalories:      targets.TargetCalories,
		TargetProteinG:      targets.TargetProteinG,
		TargetCarbsG:        targets.TargetCarbsG,
		TargetFatG:          targets.TargetFatG,
	}
	if err := s.profiles.Upsert(r.Context(), user.ID, &p); err != nil {
		s.internalError(w, "save profile", err)
		return
	}
	saved, err := s.profiles.Get(r.Context(), user.ID)
	if err != nil {
		s.internalError(w, "reload profile", err)
		return
	}
	WriteJSON(w, http.StatusOK, profileResponse{Profile: saved, Warnings: targets.Warnings})
}
