// Package config loads the HTTP server configuration from the environment.
//
// Database settings live in database.ConfigFromEnv and session/cookie settings
// in auth.ConfigFromEnv; this package covers the server itself. All of them
// read backend/.env (via database.LoadDotEnv) and let real environment
// variables take precedence.
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
	// Env is the deployment environment: development (alias: local), test,
	// staging or production.
	Env string
	// Port is the TCP port the API listens on.
	Port int
	// AllowedOrigins lists browser origins allowed to call the API cross-origin
	// with cookies. Empty means same-origin only (the Vite dev proxy locally,
	// and the single-container deployment on staging).
	AllowedOrigins []string
	// StaticDir, when set, is a directory with the built frontend (index.html +
	// assets). The API then also serves the SPA, so frontend and API share one
	// origin. Empty disables static serving (local development uses Vite).
	StaticDir string
	// AutoMigrate runs schema migrations and the reference-data seed at
	// startup before the server accepts traffic.
	AutoMigrate bool
	// ShutdownTimeout bounds graceful shutdown.
	ShutdownTimeout time.Duration
}

// Addr returns the listen address, e.g. ":8080".
func (c Config) Addr() string { return fmt.Sprintf(":%d", c.Port) }

// Strict reports whether production-grade safety checks apply (secure cookies,
// non-default database password, strong bcrypt cost). Everything except local
// development and tests is strict.
func (c Config) Strict() bool {
	switch c.Env {
	case "development", "local", "test":
		return false
	default:
		return true
	}
}

// Load reads the server variables (after loading .env) and validates them.
//
// The port comes from PORT (set by hosting platforms such as Render), then
// API_PORT, then defaults to 8080.
func Load() (Config, error) {
	database.LoadDotEnv()

	portVar := "PORT"
	rawPort := getenv("PORT", "")
	if rawPort == "" {
		portVar, rawPort = "API_PORT", getenv("API_PORT", "8080")
	}
	port, err := strconv.Atoi(rawPort)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("%s must be a port number between 1 and 65535, got %q", portVar, rawPort)
	}

	autoMigrate, err := strconv.ParseBool(getenv("DB_AUTO_MIGRATE", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("DB_AUTO_MIGRATE must be true or false, got %q", os.Getenv("DB_AUTO_MIGRATE"))
	}

	env := strings.ToLower(getenv("APP_ENV", "development"))
	switch env {
	case "development", "local", "test", "staging", "production":
	default:
		return Config{}, fmt.Errorf("APP_ENV must be development, test, staging or production, got %q", env)
	}

	return Config{
		Env:             env,
		Port:            port,
		AllowedOrigins:  splitList(os.Getenv("API_ALLOWED_ORIGINS")),
		StaticDir:       getenv("STATIC_DIR", ""),
		AutoMigrate:     autoMigrate,
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
		if p := strings.TrimRight(strings.TrimSpace(part), "/"); p != "" {
			out = append(out, p)
		}
	}
	return out
}
