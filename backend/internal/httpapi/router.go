// Package httpapi wires the HTTP routes, middleware, and handlers of the API.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Pinger is satisfied by *sql.DB; it lets the readiness check verify the
// database without depending on GORM, and lets tests substitute a fake.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// Deps are the dependencies handlers need. Feature tasks add their
// repositories here (e.g. MealEntries *repository.MealEntryRepository).
type Deps struct {
	Logger  *slog.Logger
	DB      Pinger // nil when the database could not be reached at startup
	Version string
	// AllowedOrigins enables CORS for these browser origins (optional).
	AllowedOrigins []string
}

// NewRouter builds the root handler. All API routes live under /api.
func NewRouter(d Deps) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", d.handleHealth)
	mux.HandleFunc("GET /api/health/ready", d.handleReady)

	var h http.Handler = jsonFallback(mux)
	h = cors(d.AllowedOrigins, h)
	h = recoverer(d.Logger, h)
	h = requestLogger(d.Logger, h)
	return h
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
	Time    string `json:"time"`
}

// handleHealth is the liveness probe: it succeeds whenever the process serves.
func (d Deps) handleHealth(w http.ResponseWriter, _ *http.Request) {
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
}

// handleReady is the readiness probe: it also checks the database.
func (d Deps) handleReady(w http.ResponseWriter, r *http.Request) {
	if d.DB == nil {
		WriteError(w, http.StatusServiceUnavailable, "database_unavailable", "database is not connected")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := d.DB.PingContext(ctx); err != nil {
		d.Logger.Warn("readiness: database ping failed", "err", err)
		WriteError(w, http.StatusServiceUnavailable, "database_unavailable", "database did not respond")
		return
	}
	WriteJSON(w, http.StatusOK, readyResponse{Status: "ok", Database: "up"})
}
