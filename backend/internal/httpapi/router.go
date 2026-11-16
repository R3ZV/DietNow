// Package httpapi wires the HTTP routes, middleware, and handlers of the API.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/dietnow/backend/internal/auth"
	"github.com/dietnow/backend/internal/repository"
	"gorm.io/gorm"
)

// Pinger is satisfied by *sql.DB; it lets the readiness check verify the
// database without depending on GORM, and lets tests substitute a fake.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// Deps are the dependencies handlers need.
type Deps struct {
	Logger *slog.Logger
	// DB is nil when the database could not be reached at startup. Health
	// endpoints keep working; every data endpoint then answers 503.
	DB *gorm.DB
	// Pinger is used by the readiness probe. When nil it is derived from DB.
	Pinger Pinger
	// Auth resolves session cookies. Required for the data endpoints.
	Auth    *auth.Service
	Version string
	// AllowedOrigins get CORS headers with credentials; requests that change
	// state from any other Origin are refused with 403.
	AllowedOrigins []string
	// StaticDir, when set, serves the built frontend for every non-/api path.
	StaticDir string
}

// NewRouter builds the root handler. All API routes live under /api.
func NewRouter(d Deps) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}
	if d.Pinger == nil && d.DB != nil {
		if sqlDB, err := d.DB.DB(); err == nil {
			d.Pinger = sqlDB
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", d.handleHealth)
	mux.HandleFunc("GET /api/health/ready", d.handleReady)
	registerAppRoutes(mux, d)

	var h http.Handler = jsonFallback(mux)
	if d.StaticDir != "" {
		h = withSPA(d.StaticDir, h)
	}
	h = originGuard(d.AllowedOrigins, h)
	h = securityHeaders(h)
	h = recoverer(d.Logger, h)
	h = requestLogger(d.Logger, h)
	return h
}

// server holds the per-feature repositories used by the handlers.
type server struct {
	log      *slog.Logger
	auth     *auth.Service
	profiles *repository.ProfileRepository
	foods    *repository.FoodRepository
	meals    *repository.MealEntryRepository
	weights  *repository.WeightLogRepository
}

func registerAppRoutes(mux *http.ServeMux, d Deps) {
	s := &server{log: d.Logger, auth: d.Auth}
	if d.DB != nil {
		s.profiles = repository.NewProfileRepository(d.DB)
		s.foods = repository.NewFoodRepository(d.DB)
		s.meals = repository.NewMealEntryRepository(d.DB)
		s.weights = repository.NewWeightLogRepository(d.DB)
	}
	available := d.DB != nil && d.Auth != nil

	route := func(pattern string, h http.HandlerFunc, private bool) {
		var handler http.Handler = h
		if !available {
			handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				WriteError(w, http.StatusServiceUnavailable, "database_unavailable", "database is not connected")
			})
		} else if private {
			handler = d.Auth.RequireAuth(handler)
		}
		mux.Handle(pattern, handler)
	}

	// Authentication (US-19).
	route("POST /api/auth/register", s.register, false)
	route("POST /api/auth/login", s.login, false)
	route("POST /api/auth/logout", s.logout, false)
	route("GET /api/auth/me", s.me, true)

	// Biometric intake + calculated targets (US-01..US-06).
	route("GET /api/profile", s.getProfile, true)
	route("PUT /api/profile", s.putProfile, true)

	// Food lookup (US-09).
	route("GET /api/foods", s.searchFoods, true)

	// Portion logging and daily totals (US-08, US-10, US-11).
	route("GET /api/meals", s.listMeals, true)
	route("POST /api/meals", s.createMeal, true)
	route("DELETE /api/meals/{id}", s.deleteMeal, true)
	route("GET /api/daily-summary", s.dailySummary, true)

	// Weight history (US-12).
	route("GET /api/weights", s.listWeights, true)
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
	Time    string `json:"time"`
}

// handleHealth is the liveness probe: it succeeds whenever the process serves.
// Version is the deployed git commit, which deploy verification compares.
func (d Deps) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	WriteJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Service: "dietnow-api",
		Version: d.Version,
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}

type readyResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Version  string `json:"version"`
}

// handleReady is the readiness probe: it also checks the database.
func (d Deps) handleReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if d.Pinger == nil {
		WriteError(w, http.StatusServiceUnavailable, "database_unavailable", "database is not connected")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := d.Pinger.PingContext(ctx); err != nil {
		d.Logger.Warn("readiness: database ping failed", "err", err)
		WriteError(w, http.StatusServiceUnavailable, "database_unavailable", "database did not respond")
		return
	}
	WriteJSON(w, http.StatusOK, readyResponse{Status: "ok", Database: "up", Version: d.Version})
}
