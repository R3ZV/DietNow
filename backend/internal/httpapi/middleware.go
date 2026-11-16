package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// statusRecorder captures the status code written by downstream handlers.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// requestLogger logs method, path, status, and duration for every request.
// Query strings are not logged (they may contain search terms or dates).
func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).String(),
		)
	})
}

// recoverer turns a handler panic into a 500 JSON error instead of a dropped connection.
func recoverer(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}
				log.Error("panic in handler", "panic", v, "path", r.URL.Path)
				WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// jsonFallback makes the mux's built-in 404 / 405 responses use the JSON error
// envelope instead of plain text, while keeping their status codes and the
// Allow header (a catch-all route would turn every 405 into a 404).
func jsonFallback(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := mux.Handler(r); pattern != "" {
			mux.ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(&errorRewriter{ResponseWriter: w}, r)
	})
}

// errorRewriter replaces the status-line body written by http.Error with the
// standard JSON error, discarding the original text body.
type errorRewriter struct {
	http.ResponseWriter
	wrote bool
}

func (e *errorRewriter) WriteHeader(code int) {
	if e.wrote {
		return
	}
	e.wrote = true
	switch code {
	case http.StatusMethodNotAllowed:
		WriteError(e.ResponseWriter, code, "method_not_allowed", "method not allowed for this endpoint")
	case http.StatusNotFound:
		WriteError(e.ResponseWriter, code, "not_found", "no such API endpoint")
	default:
		WriteError(e.ResponseWriter, code, "error", http.StatusText(code))
	}
}

func (e *errorRewriter) Write(b []byte) (int, error) {
	if !e.wrote {
		e.WriteHeader(http.StatusOK)
	}
	return len(b), nil // drop the plain-text body; JSON was already written
}

// originGuard applies CORS for the allowed origins and blocks state-changing
// requests that come from any other origin (CSRF defence in depth, on top of
// the SameSite=Lax session cookie). Requests without an Origin header (curl,
// same-origin GETs, server-to-server smoke tests) pass through.
func originGuard(origins []string, next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(origins))
	for _, o := range origins {
		if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
			allowed[o] = true
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" || origin == requestOrigin(r) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Add("Vary", "Origin")
		if !allowed[origin] {
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				WriteError(w, http.StatusForbidden, "origin_not_allowed", "origin not allowed")
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
			h.Set("Access-Control-Allow-Headers", "Content-Type")
			h.Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requestOrigin is the origin the request was addressed to. Browsers send an
// Origin header on same-origin POSTs too; those must not be treated as
// cross-site. Behind a TLS-terminating proxy (Render) the scheme comes from
// X-Forwarded-Proto.
func requestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
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
