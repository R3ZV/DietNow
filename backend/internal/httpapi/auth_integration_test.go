package httpapi_test

// Integration tests; they need the database from `docker compose up -d` and are skipped without it.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dietnow/backend/internal/auth"
	"github.com/dietnow/backend/internal/database"
	"github.com/dietnow/backend/internal/httpapi"
	"github.com/dietnow/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	testPassword = "correct-horse-1"
	allowedOrig  = "http://localhost:5173"
)

type env struct {
	t   *testing.T
	db  *gorm.DB
	srv *httptest.Server
	cfg auth.Config
}

func setup(t *testing.T) *env {
	t.Helper()
	database.LoadDotEnv("../../.env")
	dbCfg := database.ConfigFromEnv()
	dbCfg.LogLevel = logger.Silent
	db, err := database.Connect(dbCfg)
	if err != nil {
		t.Skipf("database unavailable (run `docker compose up -d`): %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	cfg := auth.Config{
		SessionTTL:  time.Hour,
		IdleTimeout: 30 * time.Minute,
		CookieName:  "dietnow_session",
		BcryptCost:  bcrypt.MinCost,
	}
	svc, err := auth.NewService(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(httpapi.New(db, svc, httpapi.Options{AllowedOrigins: []string{allowedOrig}}))
	t.Cleanup(srv.Close)
	return &env{t: t, db: db, srv: srv, cfg: cfg}
}

type client struct {
	e    *env
	http *http.Client
}

func (e *env) newClient() *client {
	jar, _ := cookiejar.New(nil)
	return &client{e: e, http: &http.Client{Jar: jar}}
}

func (e *env) uniqueEmail(tag string) string {
	email := fmt.Sprintf("auth-test-%s-%d@example.com", tag, time.Now().UnixNano())
	e.t.Cleanup(func() {
		e.db.Unscoped().Where("email = ?", email).Delete(&models.User{})
	})
	return email
}

func (c *client) do(method, path string, body any, hdr ...string) (*http.Response, map[string]any) {
	c.e.t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, c.e.srv.URL+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	res, err := c.http.Do(req)
	if err != nil {
		c.e.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(res.Body).Decode(&out)
	return res, out
}

func (c *client) register(email string) uint {
	c.e.t.Helper()
	res, body := c.do("POST", "/api/auth/register", map[string]string{"email": email, "password": testPassword})
	expectStatus(c.e.t, res, http.StatusCreated, body)
	return uint(body["user"].(map[string]any)["id"].(float64))
}

func expectStatus(t *testing.T, res *http.Response, want int, body map[string]any) {
	t.Helper()
	if res.StatusCode != want {
		t.Fatalf("%s %s: status %d, want %d (body %v)", res.Request.Method, res.Request.URL.Path, res.StatusCode, want, body)
	}
}

func sessionCookie(t *testing.T, res *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, c := range res.Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no %s cookie set", name)
	return nil
}

func TestRegisterStoresSaltedHashAndStartsSession(t *testing.T) {
	e := setup(t)
	c := e.newClient()
	email := e.uniqueEmail("reg")

	res, body := c.do("POST", "/api/auth/register", map[string]string{"email": strings.ToUpper(email), "password": testPassword})
	expectStatus(t, res, http.StatusCreated, body)

	ck := sessionCookie(t, res, e.cfg.CookieName)
	if !ck.HttpOnly || ck.SameSite != http.SameSiteLaxMode || ck.Value == "" {
		t.Errorf("cookie flags: HttpOnly=%v SameSite=%v", ck.HttpOnly, ck.SameSite)
	}
	if _, leaked := body["user"].(map[string]any)["passwordHash"]; leaked {
		t.Error("password hash leaked in JSON")
	}

	var u models.User
	e.db.Where("email = ?", email).First(&u)
	if u.ID == 0 {
		t.Fatal("user not stored with normalized email")
	}
	if u.PasswordHash == testPassword || !strings.HasPrefix(u.PasswordHash, "$2") {
		t.Errorf("password not stored as bcrypt hash: %q", u.PasswordHash)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(testPassword)) != nil {
		t.Error("stored hash does not verify")
	}
	var sess models.Session
	e.db.Where("user_id = ?", u.ID).First(&sess)
	if sess.TokenHash == ck.Value || len(sess.TokenHash) != 64 {
		t.Error("session token must be stored hashed, not raw")
	}

	res, body = c.do("GET", "/api/auth/me", nil)
	expectStatus(t, res, http.StatusOK, body)
	if body["user"].(map[string]any)["email"] != email {
		t.Errorf("me returned %v", body)
	}
}

func TestRegisterValidation(t *testing.T) {
	e := setup(t)
	c := e.newClient()
	email := e.uniqueEmail("val")

	cases := []struct {
		name string
		body map[string]string
		want int
	}{
		{"short password", map[string]string{"email": email, "password": "short"}, http.StatusBadRequest},
		{"bad email", map[string]string{"email": "not-an-email", "password": testPassword}, http.StatusBadRequest},
		{"too long password", map[string]string{"email": email, "password": strings.Repeat("a", 73)}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		res, body := c.do("POST", "/api/auth/register", tc.body)
		if res.StatusCode != tc.want {
			t.Errorf("%s: status %d, want %d (%v)", tc.name, res.StatusCode, tc.want, body)
		}
	}

	c.register(email)
	res, body := e.newClient().do("POST", "/api/auth/register", map[string]string{"email": email, "password": testPassword})
	expectStatus(t, res, http.StatusConflict, body)

	req, _ := http.NewRequest("POST", e.srv.URL+"/api/auth/login", strings.NewReader("email=x&password=y"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r2, _ := http.DefaultClient.Do(req)
	r2.Body.Close()
	if r2.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("form post: status %d, want 415", r2.StatusCode)
	}
}

func TestLoginValidAndInvalid(t *testing.T) {
	e := setup(t)
	email := e.uniqueEmail("login")
	e.newClient().register(email)

	c := e.newClient()
	res, wrongPw := c.do("POST", "/api/auth/login", map[string]string{"email": email, "password": "wrong-password"})
	expectStatus(t, res, http.StatusUnauthorized, wrongPw)
	if len(res.Cookies()) != 0 {
		t.Error("failed login must not set a cookie")
	}
	res, unknown := c.do("POST", "/api/auth/login", map[string]string{"email": "nobody-" + email, "password": testPassword})
	expectStatus(t, res, http.StatusUnauthorized, unknown)
	if wrongPw["error"] != unknown["error"] {
		t.Errorf("error messages differ, enabling account enumeration: %v vs %v", wrongPw, unknown)
	}
	res, body := c.do("GET", "/api/auth/me", nil)
	expectStatus(t, res, http.StatusUnauthorized, body)

	res, body = c.do("POST", "/api/auth/login", map[string]string{"email": email, "password": testPassword})
	expectStatus(t, res, http.StatusOK, body)
	res, body = c.do("GET", "/api/auth/me", nil)
	expectStatus(t, res, http.StatusOK, body)
}

func TestUnauthenticatedRequestsAreRejected(t *testing.T) {
	e := setup(t)
	c := e.newClient()
	for _, p := range []struct{ method, path string }{
		{"GET", "/api/auth/me"},
		{"GET", "/api/meals?date=2026-11-08"},
		{"DELETE", "/api/meals/1"},
		{"GET", "/api/weights?from=2026-11-01&to=2026-11-08"},
	} {
		res, body := c.do(p.method, p.path, nil)
		expectStatus(t, res, http.StatusUnauthorized, body)
		if body["error"] != "unauthenticated" || res.Header.Get("Content-Type") != "application/json" {
			t.Errorf("%s %s: unexpected 401 body %v", p.method, p.path, body)
		}
	}
	res, body := c.do("GET", "/api/auth/me", nil, "Cookie", e.cfg.CookieName+"=forged-token-value")
	expectStatus(t, res, http.StatusUnauthorized, body)
}

func TestLogoutInvalidatesSessionServerSide(t *testing.T) {
	e := setup(t)
	c := e.newClient()
	email := e.uniqueEmail("logout")
	res, body := c.do("POST", "/api/auth/register", map[string]string{"email": email, "password": testPassword})
	expectStatus(t, res, http.StatusCreated, body)
	stolen := sessionCookie(t, res, e.cfg.CookieName).Value

	res, body = c.do("POST", "/api/auth/logout", nil)
	expectStatus(t, res, http.StatusNoContent, body)

	res, body = e.newClient().do("GET", "/api/auth/me", nil, "Cookie", e.cfg.CookieName+"="+stolen)
	expectStatus(t, res, http.StatusUnauthorized, body)

	res, body = c.do("POST", "/api/auth/logout", nil)
	expectStatus(t, res, http.StatusNoContent, body)
}

func TestSessionExpiry(t *testing.T) {
	e := setup(t)

	t.Run("absolute lifetime", func(t *testing.T) {
		c := e.newClient()
		uid := c.register(e.uniqueEmail("exp"))
		e.db.Model(&models.Session{}).Where("user_id = ?", uid).Update("expires_at", time.Now().Add(-time.Second))
		res, body := c.do("GET", "/api/auth/me", nil)
		expectStatus(t, res, http.StatusUnauthorized, body)
		var n int64
		e.db.Model(&models.Session{}).Where("user_id = ?", uid).Count(&n)
		if n != 0 {
			t.Error("expired session row should be deleted")
		}
	})

	t.Run("idle timeout", func(t *testing.T) {
		c := e.newClient()
		uid := c.register(e.uniqueEmail("idle"))
		e.db.Model(&models.Session{}).Where("user_id = ?", uid).Update("last_seen_at", time.Now().Add(-e.cfg.IdleTimeout-time.Minute))
		res, body := c.do("GET", "/api/auth/me", nil)
		expectStatus(t, res, http.StatusUnauthorized, body)
	})
}

func TestUsersCannotAccessEachOthersData(t *testing.T) {
	e := setup(t)
	alice, bob := e.newClient(), e.newClient()
	aliceID := alice.register(e.uniqueEmail("alice"))
	bob.register(e.uniqueEmail("bob"))

	day := models.NewLocalDate(2026, time.November, 8)
	entry := models.MealEntry{
		UserID: aliceID, ConsumedOn: day, MealType: models.MealBreakfast,
		PortionGrams: 100, FoodName: "Alice's oats",
	}
	if err := e.db.Create(&entry).Error; err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/api/meals/%d", entry.ID)

	res, body := bob.do("GET", "/api/meals?date=2026-11-08", nil)
	expectStatus(t, res, http.StatusOK, body)
	if n := len(body["entries"].([]any)); n != 0 {
		t.Errorf("bob sees %d of alice's entries", n)
	}

	res, body = bob.do("DELETE", path, nil)
	expectStatus(t, res, http.StatusNotFound, body)
	var count int64
	e.db.Model(&models.MealEntry{}).Where("id = ?", entry.ID).Count(&count)
	if count != 1 {
		t.Fatal("bob deleted alice's entry")
	}

	res, body = alice.do("GET", "/api/meals?date=2026-11-08", nil)
	expectStatus(t, res, http.StatusOK, body)
	if n := len(body["entries"].([]any)); n != 1 {
		t.Errorf("alice sees %d entries, want 1", n)
	}
	res, body = alice.do("DELETE", path, nil)
	expectStatus(t, res, http.StatusNoContent, body)
}

func TestCrossOriginProtection(t *testing.T) {
	e := setup(t)
	c := e.newClient()
	email := e.uniqueEmail("origin")
	c.register(email)

	creds := map[string]string{"email": email, "password": testPassword}
	res, body := c.do("POST", "/api/auth/login", creds, "Origin", "https://evil.example")
	expectStatus(t, res, http.StatusForbidden, body)

	res, body = c.do("POST", "/api/auth/login", creds, "Origin", allowedOrig)
	expectStatus(t, res, http.StatusOK, body)
	if res.Header.Get("Access-Control-Allow-Origin") != allowedOrig || res.Header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("missing CORS headers for allowed origin: %v", res.Header)
	}
}
