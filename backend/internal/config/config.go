// Package config loads the HTTP server configuration from the environment.
//
// Database settings live in database.ConfigFromEnv; this package covers the
// API_* variables. Both read backend/.env (via database.LoadDotEnv) and let
// real environment variables take precedence.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dietnow/backend/internal/database"
)

// Config is the HTTP server configuration.
type Config struct {
	// Env is the deployment environment: "development" or "production".
	Env string
	// Port is the TCP port the API listens on.
	Port int
	// AllowedOrigins lists browser origins allowed to call the API cross-origin.
	// Empty means same-origin only (the dev server proxies /api, so this is
	// normally only needed when the frontend is hosted on another domain).
	AllowedOrigins []string
	// ShutdownTimeout bounds graceful shutdown.
	ShutdownTimeout time.Duration
}

// Addr returns the listen address, e.g. ":8080".
func (c Config) Addr() string { return fmt.Sprintf(":%d", c.Port) }

// Load reads API_* variables (after loading .env) and validates them.
func Load() (Config, error) {
	database.LoadDotEnv()

	port, err := strconv.Atoi(getenv("API_PORT", "8080"))
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("API_PORT must be a port number between 1 and 65535, got %q", os.Getenv("API_PORT"))
	}

	return Config{
		Env:             getenv("APP_ENV", "development"),
		Port:            port,
		AllowedOrigins:  splitList(os.Getenv("API_ALLOWED_ORIGINS")),
		ShutdownTimeout: 10 * time.Second,
	}, nil
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
