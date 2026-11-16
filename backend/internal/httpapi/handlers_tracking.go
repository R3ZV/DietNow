package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dietnow/backend/internal/auth"
	"github.com/dietnow/backend/internal/models"
	"github.com/dietnow/backend/internal/nutrition"
	"github.com/dietnow/backend/internal/repository"
)

// GET /api/foods?q=chicken&limit=20 — global foods plus the caller's private
// ones, nutrition per 100 g (US-09).
func (s *server) searchFoods(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if n := utf8.RuneCountInString(q); n < 2 || n > 100 {
		WriteValidationError(w, map[string]string{"q": "search text must be 2-100 characters"})
		return
	}
	limit := 25
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			WriteValidationError(w, map[string]string{"limit": "must be between 1 and 100"})
			return
		}
		limit = n
	}
	foods, err := s.foods.Search(r.Context(), user.ID, q, limit)
	if err != nil {
		s.internalError(w, "search foods", err)
		return
	}
	if foods == nil {
		foods = []models.Food{}
	}
	WriteJSON(w, http.StatusOK, map[string]any{"foods": foods})
}

type createMealRequest struct {
	Date     string          `json:"date"` // YYYY-MM-DD
	MealType models.MealType `json:"mealType"`
	FoodID   uint            `json:"foodId"`
	Grams    float64         `json:"grams"`
}

// POST /api/meals logs a portion of a food into one of the four meals
// (US-10). The nutrition is scaled from the per-100 g values and frozen on
// the entry.
func (s *server) createMeal(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	var in createMealRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	fields := map[string]string{}
	day, err := models.ParseLocalDate(in.Date)
	if err != nil {
		fields["date"] = "required, format YYYY-MM-DD"
	}
	if !in.MealType.Valid() {
		fields["mealType"] = "must be breakfast, lunch, dinner or snack"
	}
	if !nutrition.ValidPortion(in.Grams) {
		fields["grams"] = "must be greater than 0 and at most 2000"
	}
	if in.FoodID == 0 {
		fields["foodId"] = "required"
	}
	if len(fields) > 0 {
		WriteValidationError(w, fields)
		return
	}

	food, err := s.foods.GetVisible(r.Context(), user.ID, in.FoodID)
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "food_not_found", "food not found")
		return
	}
	if err != nil {
		s.internalError(w, "load food", err)
		return
	}

	entry := models.NewMealEntry(user.ID, day, in.MealType, food, in.Grams)
	entry.Nutrition = nutrition.Scale(food.Nutrition, in.Grams)
	if err := s.meals.Create(r.Context(), &entry); err != nil {
		s.internalError(w, "create meal entry", err)
		return
	}
	WriteJSON(w, http.StatusCreated, entry)
}

// GET /api/meals?date=YYYY-MM-DD
func (s *server) listMeals(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	day, err := models.ParseLocalDate(r.URL.Query().Get("date"))
	if err != nil {
		WriteValidationError(w, map[string]string{"date": "query parameter date=YYYY-MM-DD is required"})
		return
	}
	entries, err := s.meals.ListForDay(r.Context(), user.ID, day)
	if err != nil {
		s.internalError(w, "list meals", err)
		return
	}
	if entries == nil {
		entries = []models.MealEntry{}
	}
	WriteJSON(w, http.StatusOK, map[string]any{"date": day, "entries": entries})
}

// DELETE /api/meals/{id}
func (s *server) deleteMeal(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not_found", "meal entry not found")
		return
	}
	err = s.meals.Delete(r.Context(), user.ID, uint(id))
	if errors.Is(err, repository.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "not_found", "meal entry not found")
		return
	}
	if err != nil {
		s.internalError(w, "delete meal", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Totals is a calorie + macro amount.
type Totals struct {
	CaloriesKcal float64 `json:"caloriesKcal"`
	ProteinG     float64 `json:"proteinG"`
	CarbsG       float64 `json:"carbsG"`
	FatG         float64 `json:"fatG"`
}

func (t *Totals) add(n models.NutritionSnapshot) {
	t.CaloriesKcal += n.CaloriesKcal
	t.ProteinG += n.ProteinG
	t.CarbsG += n.CarbsG
	t.FatG += n.FatG
}

func (t Totals) rounded() Totals {
	return Totals{
		CaloriesKcal: nutrition.Round(t.CaloriesKcal, 1),
		ProteinG:     nutrition.Round(t.ProteinG, 1),
		CarbsG:       nutrition.Round(t.CarbsG, 1),
		FatG:         nutrition.Round(t.FatG, 1),
	}
}

type mealSection struct {
	MealType models.MealType    `json:"mealType"`
	Entries  []models.MealEntry `json:"entries"`
	Subtotal Totals             `json:"subtotal"`
}

type dailySummary struct {
	Date      models.LocalDate `json:"date"`
	Targets   *Totals          `json:"targets"`   // null until onboarding is done
	Consumed  Totals           `json:"consumed"`  // sum of all four meals
	Remaining *Totals          `json:"remaining"` // targets - consumed; negative = over budget
	Meals     []mealSection    `json:"meals"`     // always breakfast, lunch, dinner, snack
}

// GET /api/daily-summary?date=YYYY-MM-DD — the dashboard: the four meals with
// subtotals, the day's total, and what is left of the targets (US-08, US-11).
// Without date, "today" is taken in the user's profile time zone.
func (s *server) dailySummary(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())

	profile, err := s.profiles.Get(r.Context(), user.ID)
	hasProfile := err == nil
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		s.internalError(w, "load profile", err)
		return
	}

	var day models.LocalDate
	if raw := r.URL.Query().Get("date"); raw != "" {
		if day, err = models.ParseLocalDate(raw); err != nil {
			WriteValidationError(w, map[string]string{"date": "format YYYY-MM-DD"})
			return
		}
	} else {
		day = models.TodayIn(profile.Location())
	}

	entries, err := s.meals.ListForDay(r.Context(), user.ID, day)
	if err != nil {
		s.internalError(w, "list meals", err)
		return
	}

	order := []models.MealType{models.MealBreakfast, models.MealLunch, models.MealDinner, models.MealSnack}
	sections := make([]mealSection, len(order))
	index := map[models.MealType]int{}
	for i, mt := range order {
		sections[i] = mealSection{MealType: mt, Entries: []models.MealEntry{}}
		index[mt] = i
	}
	var consumed Totals
	for _, e := range entries {
		i, ok := index[e.MealType]
		if !ok {
			continue
		}
		sections[i].Entries = append(sections[i].Entries, e)
		sections[i].Subtotal.add(e.Nutrition)
		consumed.add(e.Nutrition)
	}
	for i := range sections {
		sections[i].Subtotal = sections[i].Subtotal.rounded()
	}

	out := dailySummary{Date: day, Consumed: consumed.rounded(), Meals: sections}
	if hasProfile {
		targets := Totals{
			CaloriesKcal: profile.TargetCalories,
			ProteinG:     profile.TargetProteinG,
			CarbsG:       profile.TargetCarbsG,
			FatG:         profile.TargetFatG,
		}
		remaining := Totals{
			CaloriesKcal: targets.CaloriesKcal - consumed.CaloriesKcal,
			ProteinG:     targets.ProteinG - consumed.ProteinG,
			CarbsG:       targets.CarbsG - consumed.CarbsG,
			FatG:         targets.FatG - consumed.FatG,
		}.rounded()
		out.Targets, out.Remaining = &targets, &remaining
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/weights?from=YYYY-MM-DD&to=YYYY-MM-DD
func (s *server) listWeights(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	q := r.URL.Query()
	from, err1 := models.ParseLocalDate(q.Get("from"))
	to, err2 := models.ParseLocalDate(q.Get("to"))
	if err1 != nil || err2 != nil {
		WriteValidationError(w, map[string]string{"from": "YYYY-MM-DD required", "to": "YYYY-MM-DD required"})
		return
	}
	logs, err := s.weights.History(r.Context(), user.ID, from, to)
	if err != nil {
		s.internalError(w, "list weights", err)
		return
	}
	if logs == nil {
		logs = []models.WeightLog{}
	}
	WriteJSON(w, http.StatusOK, map[string]any{"entries": logs})
}
