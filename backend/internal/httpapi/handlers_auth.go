package httpapi

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/dietnow/backend/internal/auth"
	"github.com/dietnow/backend/internal/models"
)

type credentials struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type authResponse struct {
	User      models.User `json:"user"`
	ExpiresAt string      `json:"expiresAt"`
}

// POST /api/auth/register
func (s *server) register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	user, err := s.auth.Register(r.Context(), in.Email, in.Password, in.DisplayName)
	switch {
	case errors.Is(err, auth.ErrInvalidEmail):
		WriteValidationError(w, map[string]string{"email": err.Error()})
		return
	case errors.Is(err, auth.ErrWeakPassword):
		WriteValidationError(w, map[string]string{"password": err.Error()})
		return
	case errors.Is(err, auth.ErrEmailTaken):
		WriteError(w, http.StatusConflict, "email_taken", err.Error())
		return
	case err != nil:
		s.internalError(w, "register", err)
		return
	}
	s.startSession(w, r, user, http.StatusCreated)
}

// POST /api/auth/login
func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	user, err := s.auth.Authenticate(r.Context(), in.Email, in.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		WriteError(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
		return
	}
	if err != nil {
		s.internalError(w, "login", err)
		return
	}
	// Prevent session fixation.
	_ = s.auth.RevokeSession(r.Context(), s.auth.TokenFromRequest(r))
	s.startSession(w, r, user, http.StatusOK)
}

func (s *server) startSession(w http.ResponseWriter, r *http.Request, user models.User, status int) {
	token, sess, err := s.auth.CreateSession(r.Context(), user.ID, r.UserAgent(), clientIP(r))
	if err != nil {
		s.internalError(w, "create session", err)
		return
	}
	s.auth.SetSessionCookie(w, token, sess.ExpiresAt)
	w.Header().Set("Cache-Control", "no-store")
	WriteJSON(w, status, authResponse{User: user, ExpiresAt: sess.ExpiresAt.UTC().Format(time.RFC3339)})
}

// POST /api/auth/logout
func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if err := s.auth.RevokeSession(r.Context(), s.auth.TokenFromRequest(r)); err != nil {
		s.internalError(w, "logout", err)
		return
	}
	s.auth.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/auth/me
func (s *server) me(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	sess, _ := auth.SessionFrom(r.Context())
	WriteJSON(w, http.StatusOK, authResponse{User: user, ExpiresAt: sess.ExpiresAt.UTC().Format(time.RFC3339)})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
