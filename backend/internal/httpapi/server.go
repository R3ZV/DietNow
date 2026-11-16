// Package httpapi exposes DietNow's JSON HTTP API.
package httpapi

import (
	"net/http"
	"strings"

	"github.com/dietnow/backend/internal/auth"
	"github.com/dietnow/backend/internal/repository"
	"gorm.io/gorm"
)

type Options struct {
	AllowedOrigins []string
}

type server struct {
	auth    *auth.Service
	meals   *repository.MealEntryRepository
	weights *repository.WeightLogRepository
}

func New(db *gorm.DB, authSvc *auth.Service, opts Options) http.Handler {
	s := &server{
		auth:    authSvc,
		meals:   repository.NewMealEntryRepository(db),
		weights: repository.NewWeightLogRepository(db),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/auth/register", s.register)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/logout", s.logout)

	private := func(h http.HandlerFunc) http.Handler { return authSvc.RequireAuth(h) }
	mux.Handle("GET /api/auth/me", private(s.me))
	mux.Handle("GET /api/meals", private(s.listMeals))
	mux.Handle("DELETE /api/meals/{id}", private(s.deleteMeal))
	mux.Handle("GET /api/weights", private(s.listWeights))

	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})

	return securityHeaders(originGuard(normalizeOrigins(opts.AllowedOrigins), mux))
}

func normalizeOrigins(in []string) map[string]bool {
	out := make(map[string]bool, len(in))
	for _, o := range in {
		if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
			out[o] = true
		}
	}
	return out
}

// originGuard applies CORS for allowed origins and blocks state-changing
// requests from any other origin.
func originGuard(allowed map[string]bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Add("Vary", "Origin")
			if !allowed[origin] {
				if r.Method != http.MethodGet && r.Method != http.MethodHead {
					writeError(w, http.StatusForbidden, "origin not allowed")
					return
				}
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				if r.Method == http.MethodOptions {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					w.Header().Set("Access-Control-Max-Age", "600")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
