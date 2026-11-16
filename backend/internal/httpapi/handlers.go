package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net"
	"net/http"
	"strconv"

	"github.com/dietnow/backend/internal/auth"
	"github.com/dietnow/backend/internal/models"
	"github.com/dietnow/backend/internal/repository"
)

const maxBodyBytes = 1 << 20

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
	case errors.Is(err, auth.ErrInvalidEmail), errors.Is(err, auth.ErrWeakPassword):
		writeError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, http.StatusConflict, err.Error())
		return
	case err != nil:
		internalError(w, "register", err)
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
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if err != nil {
		internalError(w, "login", err)
		return
	}
	// Prevent session fixation.
	_ = s.auth.RevokeSession(r.Context(), s.auth.TokenFromRequest(r))
	s.startSession(w, r, user, http.StatusOK)
}

func (s *server) startSession(w http.ResponseWriter, r *http.Request, user models.User, status int) {
	token, sess, err := s.auth.CreateSession(r.Context(), user.ID, r.UserAgent(), clientIP(r))
	if err != nil {
		internalError(w, "create session", err)
		return
	}
	s.auth.SetSessionCookie(w, token, sess.ExpiresAt)
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, status, authResponse{User: user, ExpiresAt: sess.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")})
}

// POST /api/auth/logout
func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if err := s.auth.RevokeSession(r.Context(), s.auth.TokenFromRequest(r)); err != nil {
		internalError(w, "logout", err)
		return
	}
	s.auth.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/auth/me
func (s *server) me(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	sess, _ := auth.SessionFrom(r.Context())
	writeJSON(w, http.StatusOK, authResponse{User: user, ExpiresAt: sess.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")})
}

// GET /api/meals?date=YYYY-MM-DD
func (s *server) listMeals(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	day, err := models.ParseLocalDate(r.URL.Query().Get("date"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "query parameter date=YYYY-MM-DD is required")
		return
	}
	entries, err := s.meals.ListForDay(r.Context(), user.ID, day)
	if err != nil {
		internalError(w, "list meals", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"date": day, "entries": entries})
}

// DELETE /api/meals/{id}
func (s *server) deleteMeal(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	err = s.meals.Delete(r.Context(), user.ID, uint(id))
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		internalError(w, "delete meal", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/weights?from=YYYY-MM-DD&to=YYYY-MM-DD
func (s *server) listWeights(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	q := r.URL.Query()
	from, err1 := models.ParseLocalDate(q.Get("from"))
	to, err2 := models.ParseLocalDate(q.Get("to"))
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "query parameters from and to (YYYY-MM-DD) are required")
		return
	}
	logs, err := s.weights.History(r.Context(), user.ID, from, to)
	if err != nil {
		internalError(w, "list weights", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": logs})
}

// decodeJSON requires a JSON content type and caps the body size.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func internalError(w http.ResponseWriter, op string, err error) {
	log.Printf("%s: %v", op, err)
	writeError(w, http.StatusInternalServerError, "internal error")
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
