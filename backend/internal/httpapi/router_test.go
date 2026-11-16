package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger struct{ err error }

func (f fakePinger) PingContext(context.Context) error { return f.err }

func newTestRouter(d Deps) http.Handler {
	d.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(d)
}

func do(t *testing.T, h http.Handler, method, path string, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	rec := do(t, newTestRouter(Deps{Version: "test"}), http.MethodGet, "/api/health", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
	var body healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ok" || body.Service != "dietnow-api" || body.Version != "test" || body.Time == "" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestHealthRejectsWrongMethod(t *testing.T) {
	rec := do(t, newTestRouter(Deps{}), http.MethodPost, "/api/health", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow == "" {
		t.Fatal("expected Allow header on 405")
	}
	var body ErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode 405 body %q: %v", rec.Body, err)
	}
	if body.Error.Code != "method_not_allowed" {
		t.Fatalf("error code = %q", body.Error.Code)
	}
}

func TestReady(t *testing.T) {
	tests := []struct {
		name     string
		db       Pinger
		wantCode int
	}{
		{"database up", fakePinger{}, http.StatusOK},
		{"database ping fails", fakePinger{err: errors.New("boom")}, http.StatusServiceUnavailable},
		{"database never connected", nil, http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, newTestRouter(Deps{Pinger: tt.db}), http.MethodGet, "/api/health/ready", nil)
			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.wantCode, rec.Body)
			}
			if tt.wantCode != http.StatusOK {
				var body ErrorBody
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if body.Error.Code != "database_unavailable" {
					t.Fatalf("error code = %q", body.Error.Code)
				}
			}
		})
	}
}

func TestUnknownAPIRouteReturnsJSON404(t *testing.T) {
	rec := do(t, newTestRouter(Deps{}), http.MethodGet, "/api/does-not-exist", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body ErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "not_found" {
		t.Fatalf("error code = %q", body.Error.Code)
	}
}

func TestCORS(t *testing.T) {
	h := newTestRouter(Deps{AllowedOrigins: []string{"https://app.example"}})

	rec := do(t, h, http.MethodOptions, "/api/health", map[string]string{"Origin": "https://app.example"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Fatalf("allowed origin header = %q", got)
	}

	rec = do(t, h, http.MethodGet, "/api/health", map[string]string{"Origin": "https://evil.example"})
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("disallowed origin got CORS header %q", got)
	}
}

func TestPanicIsRecovered(t *testing.T) {
	h := recoverer(slog.New(slog.NewTextHandler(io.Discard, nil)), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("kaboom")
	}))
	rec := do(t, h, http.MethodGet, "/", nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}
