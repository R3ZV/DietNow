package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"mime"
	"net/http"
)

// ErrorBody is the envelope for every non-2xx response (see docs/ARCHITECTURE.md).
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail carries a stable machine-readable code and a human message.
// Fields is set for validation errors and maps JSON field names to problems,
// so a form can highlight the offending inputs.
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// WriteJSON writes v as a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json response", "err", err)
	}
}

// WriteError writes the standard error envelope.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorBody{Error: ErrorDetail{Code: code, Message: message}})
}

// WriteValidationError writes a 400 with per-field messages.
func WriteValidationError(w http.ResponseWriter, fields map[string]string) {
	WriteJSON(w, http.StatusBadRequest, ErrorBody{Error: ErrorDetail{
		Code:    "validation_failed",
		Message: "some fields are invalid",
		Fields:  fields,
	}})
}

const maxBodyBytes = 1 << 20

// decodeJSON requires a JSON content type, caps the body size and rejects
// unknown fields. It writes the error response itself and returns false on
// failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if mt, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mt != "application/json" {
		WriteError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
		return false
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			WriteError(w, http.StatusRequestEntityTooLarge, "body_too_large", "request body is too large")
			return false
		}
		WriteError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON for this endpoint")
		return false
	}
	return true
}

func (s *server) internalError(w http.ResponseWriter, op string, err error) {
	s.log.Error("request failed", "op", op, "err", err)
	WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
}
