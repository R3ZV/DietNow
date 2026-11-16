package httpapi

import (
	"log/slog"
	"net/http"
	"slices"
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

// cors allows the configured browser origins to call the API. With no origins
// configured it is a no-op (same-origin only, which is the dev-proxy setup).
func cors(allowed []string, next http.Handler) http.Handler {
	if len(allowed) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && slices.Contains(allowed, origin) {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			h.Add("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
