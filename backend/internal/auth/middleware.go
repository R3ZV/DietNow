package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dietnow/backend/internal/models"
)

type ctxKey struct{}

type principal struct {
	user    models.User
	session models.Session
}

// UserFrom returns the authenticated user set by RequireAuth.
func UserFrom(ctx context.Context) (models.User, bool) {
	p, ok := ctx.Value(ctxKey{}).(principal)
	return p.user, ok
}

// SessionFrom returns the current session.
func SessionFrom(ctx context.Context) (models.Session, bool) {
	p, ok := ctx.Value(ctxKey{}).(principal)
	return p.session, ok
}

// RequireAuth responds 401 unless the request carries a valid session cookie.
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, sess, err := s.ValidateSession(r.Context(), s.TokenFromRequest(r))
		if err != nil {
			if !errors.Is(err, ErrNoSession) {
				log.Printf("auth: validate session: %v", err)
				writeErr(w, http.StatusInternalServerError, "internal error")
				return
			}
			s.ClearSessionCookie(w)
			w.Header().Set("Cache-Control", "no-store")
			writeErr(w, http.StatusUnauthorized, "unauthenticated")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		ctx := context.WithValue(r.Context(), ctxKey{}, principal{user: user, session: sess})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Service) TokenFromRequest(r *http.Request) string {
	c, err := r.Cookie(s.cfg.CookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// SetSessionCookie writes an HttpOnly, SameSite=Lax session cookie.
func (s *Service) SetSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   s.cfg.CookieDomain,
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   s.cfg.CookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
