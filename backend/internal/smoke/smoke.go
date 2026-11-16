// Package smoke drives the MVP user journey against a running DietNow
// deployment over plain HTTP:
//
//	health -> readiness -> auth -> biometric intake -> targets -> food lookup
//	-> portion logging -> updated totals -> delete -> totals restored -> logout
//
// The same flow is used by the integration tests (against an in-process
// server), by CI (against the freshly built Docker image) and by the staging
// deploy pipeline (against the live service, with a non-production account).
// It cleans up the meal entry it creates, so it can run repeatedly against
// the same account.
package smoke

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Options configures a smoke run.
type Options struct {
	BaseURL  string // e.g. https://dietnow-staging.onrender.com
	Email    string // non-production account
	Password string
	// Register creates the account when login fails with 401. Use it for the
	// first run against a fresh environment and for throwaway test accounts.
	Register bool
	// ExpectVersion, if set, must equal the version reported by /api/health
	// (the git commit the deploy pipeline just shipped).
	ExpectVersion string
	// Logf receives one line per step. Optional.
	Logf func(format string, args ...any)
	// HTTPClient is optional; a client with a cookie jar is created if nil.
	HTTPClient *http.Client
}

// Food search term and portion used by the flow. "Rolled Oats" is part of the
// reference catalog that every environment receives with its migrations.
const (
	FoodQuery = "rolled oats"
	FoodName  = "Rolled Oats"
	Portion   = 80.0
)

// ExpectedTargets are the calorie-engine results for the smoke profile
// (male, 22 y, 180 cm, 70 kg, moderately active, bulk to 75 kg in 10 weeks).
// They match the backlog examples in US-04, US-05 and US-06.
var ExpectedTargets = map[string]float64{
	"bmr": 1720, "tdee": 2666, "targetCalories": 3066,
	"targetProteinG": 140, "targetFatG": 70, "targetCarbsG": 469,
}

// Run executes the flow and returns the first failure.
func Run(ctx context.Context, o Options) (err error) {
	c, err := newClient(o)
	if err != nil {
		return err
	}

	// 1. Liveness and readiness.
	var health struct{ Status, Version string }
	if err := c.expect(ctx, "GET", "/api/health", nil, 200, &health); err != nil {
		return err
	}
	if health.Status != "ok" {
		return fmt.Errorf("health status %q, want ok", health.Status)
	}
	if o.ExpectVersion != "" && health.Version != o.ExpectVersion {
		return fmt.Errorf("deployed version %q, want %q", health.Version, o.ExpectVersion)
	}
	c.logf("health ok (version %s)", health.Version)
	if err := c.expect(ctx, "GET", "/api/health/ready", nil, 200, nil); err != nil {
		return err
	}
	c.logf("ready: database up")

	// 2. Private data requires a session.
	if err := c.expectError(ctx, "GET", "/api/daily-summary", nil, 401, "unauthenticated"); err != nil {
		return err
	}
	c.logf("unauthenticated request rejected with 401")

	// 3. Authenticate.
	creds := map[string]string{"email": o.Email, "password": o.Password}
	status, body, err := c.do(ctx, "POST", "/api/auth/login", creds)
	if err != nil {
		return err
	}
	switch {
	case status == 200:
		c.logf("logged in as %s", o.Email)
	case status == 401 && o.Register:
		if err := c.expect(ctx, "POST", "/api/auth/register", creds, 201, nil); err != nil {
			return err
		}
		c.logf("registered %s", o.Email)
	default:
		return fmt.Errorf("POST /api/auth/login: status %d: %s", status, body)
	}

	// 4. Biometric intake -> targets.
	today := time.Now().UTC()
	profileReq := map[string]any{
		"gender":          "male",
		"birthDate":       fmt.Sprintf("%04d-01-01", today.Year()-22), // 22 years old today
		"heightCm":        180,
		"currentWeightKg": 70,
		"activityLevel":   "moderate",
		"goal":            "bulk",
		"targetWeightKg":  75,
		"timeframeWeeks":  10,
		"timezone":        "UTC",
	}
	var saved struct {
		Profile  map[string]any `json:"profile"`
		Warnings []string       `json:"warnings"`
	}
	if err := c.expect(ctx, "PUT", "/api/profile", profileReq, 200, &saved); err != nil {
		return err
	}
	if err := checkTargets(saved.Profile); err != nil {
		return fmt.Errorf("PUT /api/profile: %w", err)
	}
	var got struct {
		Profile map[string]any `json:"profile"`
	}
	if err := c.expect(ctx, "GET", "/api/profile", nil, 200, &got); err != nil {
		return err
	}
	if err := checkTargets(got.Profile); err != nil {
		return fmt.Errorf("GET /api/profile: %w", err)
	}
	c.logf("profile saved; targets %v kcal (P%v C%v F%v)", got.Profile["targetCalories"],
		got.Profile["targetProteinG"], got.Profile["targetCarbsG"], got.Profile["targetFatG"])

	// 5. Food lookup.
	var found struct {
		Foods []struct {
			ID        uint   `json:"id"`
			Name      string `json:"name"`
			Nutrition Macros `json:"nutrition"`
		} `json:"foods"`
	}
	if err := c.expect(ctx, "GET", "/api/foods?q="+url.QueryEscape(FoodQuery), nil, 200, &found); err != nil {
		return err
	}
	var foodID uint
	var per100 Macros
	for _, f := range found.Foods {
		if f.Name == FoodName {
			foodID, per100 = f.ID, f.Nutrition
		}
	}
	if foodID == 0 {
		return fmt.Errorf("food search %q did not return %q (is the food catalog migrated?)", FoodQuery, FoodName)
	}
	c.logf("found %q: %v kcal/100 g", FoodName, per100.CaloriesKcal)

	// 6. Portion logging and updated totals.
	date := today.Format("2006-01-02")
	before, err := c.summary(ctx, date)
	if err != nil {
		return err
	}

	// A portion outside the allowed range is rejected and changes nothing.
	if err := c.expectError(ctx, "POST", "/api/meals",
		map[string]any{"date": date, "mealType": "breakfast", "foodId": foodID, "grams": 0}, 400, "validation_failed"); err != nil {
		return err
	}

	var entry struct {
		ID           uint    `json:"id"`
		PortionGrams float64 `json:"portionGrams"`
		Nutrition    Macros  `json:"nutrition"`
	}
	if err := c.expect(ctx, "POST", "/api/meals",
		map[string]any{"date": date, "mealType": "breakfast", "foodId": foodID, "grams": Portion}, 201, &entry); err != nil {
		return err
	}
	deleted := false
	defer func() {
		if !deleted && entry.ID != 0 {
			_, _, _ = c.do(context.Background(), "DELETE", fmt.Sprintf("/api/meals/%d", entry.ID), nil)
		}
	}()
	want := per100.Scale(Portion)
	if !want.Equal(entry.Nutrition) {
		return fmt.Errorf("logged %v g: nutrition %+v, want %+v", Portion, entry.Nutrition, want)
	}
	c.logf("logged %v g -> %v kcal", Portion, entry.Nutrition.CaloriesKcal)

	after, err := c.summary(ctx, date)
	if err != nil {
		return err
	}
	if !after.Consumed.Equal(before.Consumed.Plus(want)) {
		return fmt.Errorf("daily total %+v, want %+v + %+v", after.Consumed, before.Consumed, want)
	}
	if !after.meal("breakfast").Equal(before.meal("breakfast").Plus(want)) {
		return fmt.Errorf("breakfast subtotal %+v, want %+v + %+v", after.meal("breakfast"), before.meal("breakfast"), want)
	}
	if after.Targets == nil || after.Remaining == nil {
		return errors.New("daily summary has no targets/remaining after onboarding")
	}
	if !after.Remaining.Equal(after.Targets.Minus(after.Consumed)) {
		return fmt.Errorf("remaining %+v, want targets %+v - consumed %+v", *after.Remaining, *after.Targets, after.Consumed)
	}
	c.logf("totals updated: consumed %v kcal, remaining %v kcal", after.Consumed.CaloriesKcal, after.Remaining.CaloriesKcal)

	// 7. Deleting the entry restores the totals.
	if err := c.expect(ctx, "DELETE", fmt.Sprintf("/api/meals/%d", entry.ID), nil, 204, nil); err != nil {
		return err
	}
	deleted = true
	restored, err := c.summary(ctx, date)
	if err != nil {
		return err
	}
	if !restored.Consumed.Equal(before.Consumed) {
		return fmt.Errorf("after delete: total %+v, want %+v", restored.Consumed, before.Consumed)
	}
	c.logf("entry deleted; totals restored")

	// 8. Logout invalidates the session.
	if err := c.expect(ctx, "POST", "/api/auth/logout", nil, 204, nil); err != nil {
		return err
	}
	if err := c.expectError(ctx, "GET", "/api/auth/me", nil, 401, "unauthenticated"); err != nil {
		return err
	}
	c.logf("logged out; session invalidated")
	return nil
}

func checkTargets(p map[string]any) error {
	for k, want := range ExpectedTargets {
		got, _ := p[k].(float64)
		if math.Abs(got-want) > 0.05 {
			return fmt.Errorf("%s = %v, want %v", k, p[k], want)
		}
	}
	return nil
}

// Macros mirrors the API's calorie/macro objects.
type Macros struct {
	CaloriesKcal float64 `json:"caloriesKcal"`
	ProteinG     float64 `json:"proteinG"`
	CarbsG       float64 `json:"carbsG"`
	FatG         float64 `json:"fatG"`
}

// Scale returns the macros of a portion, rounded to one decimal like the API.
func (m Macros) Scale(grams float64) Macros {
	r := func(v float64) float64 { return math.Round(v*grams/100*10) / 10 }
	return Macros{r(m.CaloriesKcal), r(m.ProteinG), r(m.CarbsG), r(m.FatG)}
}

func (m Macros) Plus(o Macros) Macros {
	return Macros{m.CaloriesKcal + o.CaloriesKcal, m.ProteinG + o.ProteinG, m.CarbsG + o.CarbsG, m.FatG + o.FatG}
}

func (m Macros) Minus(o Macros) Macros {
	return Macros{m.CaloriesKcal - o.CaloriesKcal, m.ProteinG - o.ProteinG, m.CarbsG - o.CarbsG, m.FatG - o.FatG}
}

// Equal compares with a tolerance that absorbs one-decimal rounding.
func (m Macros) Equal(o Macros) bool {
	near := func(a, b float64) bool { return math.Abs(a-b) < 0.11 }
	return near(m.CaloriesKcal, o.CaloriesKcal) && near(m.ProteinG, o.ProteinG) &&
		near(m.CarbsG, o.CarbsG) && near(m.FatG, o.FatG)
}

// Summary mirrors GET /api/daily-summary.
type Summary struct {
	Date      string  `json:"date"`
	Targets   *Macros `json:"targets"`
	Consumed  Macros  `json:"consumed"`
	Remaining *Macros `json:"remaining"`
	Meals     []struct {
		MealType string            `json:"mealType"`
		Entries  []json.RawMessage `json:"entries"`
		Subtotal Macros            `json:"subtotal"`
	} `json:"meals"`
}

func (s Summary) meal(t string) Macros {
	for _, m := range s.Meals {
		if m.MealType == t {
			return m.Subtotal
		}
	}
	return Macros{}
}

type client struct {
	base string
	http *http.Client
	logf func(string, ...any)
}

func newClient(o Options) (*client, error) {
	base := strings.TrimRight(strings.TrimSpace(o.BaseURL), "/")
	if _, err := url.ParseRequestURI(base); err != nil || base == "" {
		return nil, fmt.Errorf("invalid base URL %q", o.BaseURL)
	}
	if o.Email == "" || o.Password == "" {
		return nil, errors.New("smoke account email and password are required")
	}
	hc := o.HTTPClient
	if hc == nil {
		jar, _ := cookiejar.New(nil)
		hc = &http.Client{Jar: jar, Timeout: 30 * time.Second}
	}
	logf := o.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}
	return &client{base: base, http: hc, logf: logf}, nil
}

func (c *client) do(ctx context.Context, method, path string, body any) (int, []byte, error) {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rd)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	return res.StatusCode, data, err
}

func (c *client) expect(ctx context.Context, method, path string, body any, want int, out any) error {
	status, data, err := c.do(ctx, method, path, body)
	if err != nil {
		return err
	}
	if status != want {
		return fmt.Errorf("%s %s: status %d, want %d: %s", method, path, status, want, bytes.TrimSpace(data))
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("%s %s: decode response: %w", method, path, err)
		}
	}
	return nil
}

func (c *client) expectError(ctx context.Context, method, path string, body any, want int, code string) error {
	var e struct {
		Error struct{ Code string } `json:"error"`
	}
	if err := c.expect(ctx, method, path, body, want, &e); err != nil {
		return err
	}
	if e.Error.Code != code {
		return fmt.Errorf("%s %s: error code %q, want %q", method, path, e.Error.Code, code)
	}
	return nil
}

func (c *client) summary(ctx context.Context, date string) (Summary, error) {
	var s Summary
	err := c.expect(ctx, "GET", "/api/daily-summary?date="+date, nil, 200, &s)
	if err == nil && len(s.Meals) != 4 {
		err = fmt.Errorf("daily summary has %d meal sections, want 4", len(s.Meals))
	}
	return s, err
}
