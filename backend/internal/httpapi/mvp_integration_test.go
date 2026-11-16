package httpapi_test

// End-to-end MVP tests against a real PostgreSQL (see setup in
// auth_integration_test.go): the full user journey, user data isolation, and
// calculation / portion boundaries through the HTTP API.

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/dietnow/backend/internal/models"
	"github.com/dietnow/backend/internal/smoke"
)

// TestMVPFlowSmoke runs the exact smoke flow used against staging:
// auth -> biometric intake -> targets -> food lookup -> portion logging ->
// updated totals -> delete -> logout.
func TestMVPFlowSmoke(t *testing.T) {
	e := setup(t)
	email := e.uniqueEmail("smoke")
	err := smoke.Run(context.Background(), smoke.Options{
		BaseURL:       e.srv.URL,
		Email:         email,
		Password:      testPassword,
		Register:      true,
		ExpectVersion: "test",
		Logf:          t.Logf,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Re-running with the same (now existing) account logs in instead of
	// registering and leaves no entries behind, as on staging.
	if err := smoke.Run(context.Background(), smoke.Options{BaseURL: e.srv.URL, Email: email, Password: testPassword}); err != nil {
		t.Fatalf("second run: %v", err)
	}
	var n int64
	e.db.Model(&models.MealEntry{}).Joins("JOIN users ON users.id = meal_entries.user_id").Where("users.email = ?", email).Count(&n)
	if n != 0 {
		t.Errorf("smoke run left %d meal entries behind", n)
	}
}

func validProfile() map[string]any {
	return map[string]any{
		"gender": "male", "birthDate": "2000-01-01", "heightCm": 180, "currentWeightKg": 70,
		"activityLevel": "moderate", "goal": "bulk", "targetWeightKg": 75, "timeframeWeeks": 10,
		"timezone": "Europe/Bucharest",
	}
}

func with(base map[string]any, kv ...any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for i := 0; i+1 < len(kv); i += 2 {
		if kv[i+1] == nil {
			delete(out, kv[i].(string))
		} else {
			out[kv[i].(string)] = kv[i+1]
		}
	}
	return out
}

func (c *client) catalogFood(name string) map[string]any {
	c.e.t.Helper()
	res, body := c.do("GET", "/api/foods?q="+url.QueryEscape(name), nil)
	expectStatus(c.e.t, res, http.StatusOK, body)
	for _, f := range body["foods"].([]any) {
		if m := f.(map[string]any); m["name"] == name {
			return m
		}
	}
	c.e.t.Fatalf("food %q not found in %v", name, body)
	return nil
}

func (c *client) logMeal(date, meal string, foodID any, grams float64) map[string]any {
	c.e.t.Helper()
	res, body := c.do("POST", "/api/meals", map[string]any{"date": date, "mealType": meal, "foodId": foodID, "grams": grams})
	expectStatus(c.e.t, res, http.StatusCreated, body)
	return body
}

func (c *client) summary(date string) map[string]any {
	c.e.t.Helper()
	res, body := c.do("GET", "/api/daily-summary?date="+date, nil)
	expectStatus(c.e.t, res, http.StatusOK, body)
	return body
}

func num(m any, path ...string) float64 {
	for _, p := range path {
		m = m.(map[string]any)[p]
	}
	f, _ := m.(float64)
	return f
}

func near(a, b float64) bool { return math.Abs(a-b) < 0.051 }

func mealSubtotal(summary map[string]any, meal string) float64 {
	for _, s := range summary["meals"].([]any) {
		if s.(map[string]any)["mealType"] == meal {
			return num(s, "subtotal", "caloriesKcal")
		}
	}
	return math.NaN()
}

func TestMVPUserDataIsolation(t *testing.T) {
	e := setup(t)
	alice, bob := e.newClient(), e.newClient()
	aliceID := alice.register(e.uniqueEmail("iso-alice"))
	bob.register(e.uniqueEmail("iso-bob"))
	const day = "2026-11-10"

	// Alice onboards and logs a catalog food plus a private food.
	res, body := alice.do("PUT", "/api/profile", validProfile())
	expectStatus(t, res, http.StatusOK, body)
	aliceTarget := num(body, "profile", "targetCalories")

	private := models.Food{Name: fmt.Sprintf("Alice secret shake %d", time.Now().UnixNano()), OwnerUserID: &aliceID,
		Nutrition: models.Macros{CaloriesKcal: 120, ProteinG: 20, CarbsG: 5, FatG: 2}}
	if err := e.db.Create(&private).Error; err != nil {
		t.Fatal(err)
	}
	oats := alice.catalogFood("Rolled Oats")
	aliceEntry := alice.logMeal(day, "breakfast", oats["id"], 100)
	alice.logMeal(day, "snack", private.ID, 50)
	aliceBefore := alice.summary(day)

	// Bob has no profile and cannot see Alice's.
	res, body = bob.do("GET", "/api/profile", nil)
	expectStatus(t, res, http.StatusNotFound, body)
	if errorCode(body) != "profile_not_found" {
		t.Errorf("bob profile: %v", body)
	}

	// Bob's dashboard for the same day is empty and has no targets.
	bs := bob.summary(day)
	if num(bs, "consumed", "caloriesKcal") != 0 || bs["targets"] != nil || bs["remaining"] != nil {
		t.Errorf("bob sees data on alice's day: %v", bs)
	}
	res, body = bob.do("GET", "/api/meals?date="+day, nil)
	expectStatus(t, res, http.StatusOK, body)
	if len(body["entries"].([]any)) != 0 {
		t.Errorf("bob lists alice's entries: %v", body)
	}

	// Bob can neither find nor log Alice's private food.
	res, body = bob.do("GET", "/api/foods?q=Alice+secret", nil)
	expectStatus(t, res, http.StatusOK, body)
	// Fuzzy search may return other, less similar foods, but never Alice's.
	for _, food := range body["foods"].([]any) {
		if food.(map[string]any)["id"] == float64(private.ID) {
			t.Errorf("bob found alice's private food: %v", body)
		}
	}
	res, body = bob.do("POST", "/api/meals", map[string]any{"date": day, "mealType": "lunch", "foodId": private.ID, "grams": 100})
	expectStatus(t, res, http.StatusNotFound, body)
	if errorCode(body) != "food_not_found" {
		t.Errorf("bob logging alice's private food: %v", body)
	}

	// Bob cannot delete Alice's entry; it is reported as not found.
	res, body = bob.do("DELETE", fmt.Sprintf("/api/meals/%v", aliceEntry["id"]), nil)
	expectStatus(t, res, http.StatusNotFound, body)

	// Bob's own onboarding and logging do not touch Alice's targets or totals.
	res, body = bob.do("PUT", "/api/profile", with(validProfile(), "gender", "female", "currentWeightKg", 55, "targetWeightKg", 50, "goal", "cut"))
	expectStatus(t, res, http.StatusOK, body)
	bob.logMeal(day, "breakfast", oats["id"], 300)

	aliceAfter := alice.summary(day)
	if num(aliceAfter, "consumed", "caloriesKcal") != num(aliceBefore, "consumed", "caloriesKcal") {
		t.Errorf("alice's totals changed after bob's activity: %v -> %v", aliceBefore["consumed"], aliceAfter["consumed"])
	}
	if num(aliceAfter, "targets", "caloriesKcal") != aliceTarget {
		t.Errorf("alice's target changed: %v -> %v", aliceTarget, aliceAfter["targets"])
	}
	// Alice still sees her private food.
	res, body = alice.do("GET", "/api/foods?q=Alice+secret", nil)
	expectStatus(t, res, http.StatusOK, body)
	foods := body["foods"].([]any)
	if len(foods) == 0 || foods[0].(map[string]any)["id"] != float64(private.ID) {
		t.Errorf("alice's exact private-food match is not ranked first: %v", body)
	}
}

func TestMVPBiometricAndTargetBoundaries(t *testing.T) {
	e := setup(t)
	c := e.newClient()
	c.register(e.uniqueEmail("bounds"))

	tomorrow := time.Now().UTC().AddDate(0, 0, 2).Format("2006-01-02")
	invalid := []struct {
		name  string
		body  map[string]any
		field string
	}{
		{"negative weight", with(validProfile(), "currentWeightKg", -10), "currentWeightKg"},
		{"600 kg", with(validProfile(), "currentWeightKg", 600), "currentWeightKg"},
		{"missing height", with(validProfile(), "heightCm", nil), "heightCm"},
		{"height above max", with(validProfile(), "heightCm", 250.5), "heightCm"},
		{"missing gender", with(validProfile(), "gender", nil), "gender"},
		{"missing birth date", with(validProfile(), "birthDate", nil), "birthDate"},
		{"birth date in the future", with(validProfile(), "birthDate", tomorrow), "birthDate"},
		{"too young", with(validProfile(), "birthDate", time.Now().UTC().AddDate(-12, 0, 0).Format("2006-01-02")), "birthDate"},
		{"unknown activity", with(validProfile(), "activityLevel", "couch"), "activityLevel"},
		{"bulk target below current", with(validProfile(), "targetWeightKg", 65), "targetWeightKg"},
		{"cut target above current", with(validProfile(), "goal", "cut"), "targetWeightKg"},
		{"zero weeks", with(validProfile(), "timeframeWeeks", 0), "timeframeWeeks"},
		{"bad time zone", with(validProfile(), "timezone", "Mars/Olympus"), "timezone"},
	}
	for _, tc := range invalid {
		res, body := c.do("PUT", "/api/profile", tc.body)
		if res.StatusCode != http.StatusBadRequest || errorCode(body) != "validation_failed" {
			t.Errorf("%s: status %d body %v, want 400 validation_failed", tc.name, res.StatusCode, body)
			continue
		}
		fields, _ := body["error"].(map[string]any)["fields"].(map[string]any)
		if _, ok := fields[tc.field]; !ok {
			t.Errorf("%s: no error for field %q in %v", tc.name, tc.field, fields)
		}
	}
	// Nothing was stored by the rejected submissions.
	res, body := c.do("GET", "/api/profile", nil)
	expectStatus(t, res, http.StatusNotFound, body)

	// An empty form reports every required field at once.
	res, body = c.do("PUT", "/api/profile", map[string]any{})
	expectStatus(t, res, http.StatusBadRequest, body)
	fields := body["error"].(map[string]any)["fields"].(map[string]any)
	for _, f := range []string{"gender", "birthDate", "heightCm", "currentWeightKg", "activityLevel", "goal"} {
		if _, ok := fields[f]; !ok {
			t.Errorf("empty form: missing error for %q (%v)", f, fields)
		}
	}

	// Unknown JSON fields are rejected rather than ignored.
	res, body = c.do("PUT", "/api/profile", with(validProfile(), "userId", 1))
	expectStatus(t, res, http.StatusBadRequest, body)

	// Valid intake: targets are computed and persisted (US-04/05/06 numbers
	// depend on age; the smoke test checks the exact backlog values).
	res, body = c.do("PUT", "/api/profile", validProfile())
	expectStatus(t, res, http.StatusOK, body)
	p := body["profile"].(map[string]any)
	if p["weeklyRateKg"] != 0.5 || p["dailyAdjustmentKcal"] != 400.0 || num(p, "targetCalories") != num(p, "tdee")+400 {
		t.Errorf("bulk targets: %v", p)
	}
	if kcal := num(p, "targetProteinG")*4 + num(p, "targetCarbsG")*4 + num(p, "targetFatG")*9; math.Abs(kcal-num(p, "targetCalories")) > 1 {
		t.Errorf("macros add up to %v kcal, target %v", kcal, p["targetCalories"])
	}
	if w := body["warnings"].([]any); len(w) != 0 {
		t.Errorf("unexpected warnings %v", w)
	}

	// Unrealistic pace is saved but flagged (US-03 scenario 2).
	res, body = c.do("PUT", "/api/profile", with(validProfile(), "targetWeightKg", 80, "timeframeWeeks", 2))
	expectStatus(t, res, http.StatusOK, body)
	if w := body["warnings"].([]any); len(w) != 1 {
		t.Errorf("10 kg in 2 weeks: warnings %v, want 1", w)
	}

	// Switching to maintenance removes the surplus (US-20 recalibration).
	res, body = c.do("PUT", "/api/profile", with(validProfile(), "goal", "maintain", "targetWeightKg", nil, "timeframeWeeks", nil))
	expectStatus(t, res, http.StatusOK, body)
	p = body["profile"].(map[string]any)
	if num(p, "targetCalories") != num(p, "tdee") || num(p, "dailyAdjustmentKcal") != 0 {
		t.Errorf("maintain: %v", p)
	}
}

func TestMVPPortionLoggingAndTotals(t *testing.T) {
	e := setup(t)
	c := e.newClient()
	c.register(e.uniqueEmail("portions"))
	const day = "2026-11-11"

	// Before onboarding the dashboard works but has no targets.
	s := c.summary(day)
	if s["targets"] != nil || len(s["meals"].([]any)) != 4 {
		t.Errorf("summary before onboarding: %v", s)
	}
	res, body := c.do("PUT", "/api/profile", validProfile())
	expectStatus(t, res, http.StatusOK, body)
	target := num(body, "profile", "targetCalories")

	oats := c.catalogFood("Rolled Oats") // 389 kcal / 100 g
	chicken := c.catalogFood("Chicken Breast, cooked")
	oil := c.catalogFood("Extra Virgin Olive Oil") // 884 kcal / 100 g

	// Portion boundaries.
	for _, tc := range []struct {
		name string
		body map[string]any
		want int
	}{
		{"zero grams", map[string]any{"date": day, "mealType": "breakfast", "foodId": oats["id"], "grams": 0}, 400},
		{"negative grams", map[string]any{"date": day, "mealType": "breakfast", "foodId": oats["id"], "grams": -5}, 400},
		{"over 2 kg", map[string]any{"date": day, "mealType": "breakfast", "foodId": oats["id"], "grams": 2000.1}, 400},
		{"bad meal type", map[string]any{"date": day, "mealType": "brunch", "foodId": oats["id"], "grams": 50}, 400},
		{"bad date", map[string]any{"date": "11/11/2026", "mealType": "lunch", "foodId": oats["id"], "grams": 50}, 400},
		{"missing food", map[string]any{"date": day, "mealType": "lunch", "grams": 50}, 400},
		{"unknown food", map[string]any{"date": day, "mealType": "lunch", "foodId": 999999999, "grams": 50}, 404},
		{"grams as text", map[string]any{"date": day, "mealType": "lunch", "foodId": oats["id"], "grams": "80g"}, 400},
	} {
		res, body := c.do("POST", "/api/meals", tc.body)
		if res.StatusCode != tc.want {
			t.Errorf("%s: status %d, want %d (%v)", tc.name, res.StatusCode, tc.want, body)
		}
	}
	if got := num(c.summary(day), "consumed", "caloriesKcal"); got != 0 {
		t.Fatalf("rejected portions were stored: consumed %v", got)
	}

	// Scaling: 80 g oats = 311.2 kcal; smallest and largest allowed portions.
	e1 := c.logMeal(day, "breakfast", oats["id"], 80)
	if num(e1, "nutrition", "caloriesKcal") != 311.2 || num(e1, "nutrition", "proteinG") != 13.5 {
		t.Errorf("80 g oats: %v", e1["nutrition"])
	}
	e2 := c.logMeal(day, "lunch", chicken["id"], 200) // 330 kcal, 62 g protein
	e3 := c.logMeal(day, "lunch", oats["id"], 0.1)    // 0.4 kcal
	if num(e3, "nutrition", "caloriesKcal") != 0.4 {
		t.Errorf("0.1 g oats: %v", e3["nutrition"])
	}

	s = c.summary(day)
	if !near(mealSubtotal(s, "breakfast"), 311.2) || !near(mealSubtotal(s, "lunch"), 330.4) ||
		mealSubtotal(s, "dinner") != 0 || mealSubtotal(s, "snack") != 0 {
		t.Errorf("subtotals: %v", s["meals"])
	}
	if !near(num(s, "consumed", "caloriesKcal"), 641.6) || !near(num(s, "consumed", "proteinG"), 75.5) {
		t.Errorf("consumed: %v", s["consumed"])
	}
	if !near(num(s, "remaining", "caloriesKcal"), target-641.6) {
		t.Errorf("remaining: %v (target %v)", s["remaining"], target)
	}
	// Another day is unaffected.
	if num(c.summary("2026-11-12"), "consumed", "caloriesKcal") != 0 {
		t.Error("entries leaked into the next day")
	}

	// Going over budget yields a negative remaining value: 2000 g oil = 17680 kcal.
	big := c.logMeal(day, "dinner", oil["id"], 2000)
	s = c.summary(day)
	if !near(num(s, "remaining", "caloriesKcal"), target-641.6-17680) || num(s, "remaining", "caloriesKcal") >= 0 {
		t.Errorf("over budget: remaining %v", s["remaining"])
	}

	// Deleting entries updates the totals immediately (US-10 scenario 2).
	for _, e := range []map[string]any{big, e2, e3} {
		res, body := c.do("DELETE", fmt.Sprintf("/api/meals/%v", e["id"]), nil)
		expectStatus(t, res, http.StatusNoContent, body)
	}
	s = c.summary(day)
	if !near(num(s, "consumed", "caloriesKcal"), 311.2) || mealSubtotal(s, "lunch") != 0 {
		t.Errorf("after delete: %v", s["consumed"])
	}
	// Deleting twice is a 404, not a silent success.
	res, body = c.do("DELETE", fmt.Sprintf("/api/meals/%v", e2["id"]), nil)
	expectStatus(t, res, http.StatusNotFound, body)
}

func TestMVPFoodSearchBoundaries(t *testing.T) {
	e := setup(t)
	c := e.newClient()
	c.register(e.uniqueEmail("search"))

	for _, q := range []string{"", "a", "%20%20"} {
		res, body := c.do("GET", "/api/foods?q="+q, nil)
		expectStatus(t, res, http.StatusBadRequest, body)
	}
	// LIKE wildcards are matched literally, so "%%" does not list every food.
	res, body := c.do("GET", "/api/foods?q=%25%25", nil)
	expectStatus(t, res, http.StatusOK, body)
	if n := len(body["foods"].([]any)); n != 0 {
		t.Errorf(`"%%%%" matched %d foods`, n)
	}
	// Case-insensitive partial match, per-100 g values included.
	res, body = c.do("GET", "/api/foods?q=CHICKEN&limit=5", nil)
	expectStatus(t, res, http.StatusOK, body)
	foods := body["foods"].([]any)
	if len(foods) == 0 || num(foods[0], "nutrition", "caloriesKcal") != 165 {
		t.Errorf("chicken search: %v", foods)
	}
	res, body = c.do("GET", "/api/foods?q=chicken&limit=0", nil)
	expectStatus(t, res, http.StatusBadRequest, body)
}
