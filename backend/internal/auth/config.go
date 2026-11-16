// Package auth implements bcrypt password hashing and server-side sessions
// carried in an HttpOnly cookie. See AUTH.md.
package auth

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Config controls password hashing and session/cookie behaviour.
type Config struct {
	SessionTTL  time.Duration // absolute lifetime from login
	IdleTimeout time.Duration // inactivity limit

	CookieName   string
	CookieSecure bool
	CookieDomain string

	BcryptCost int
}

// ConfigFromEnv reads the SESSION_* / COOKIE_* / BCRYPT_COST variables.
func ConfigFromEnv() Config {
	return Config{
		SessionTTL:   time.Duration(envInt("SESSION_TTL_HOURS", 24*7)) * time.Hour,
		IdleTimeout:  time.Duration(envInt("SESSION_IDLE_MINUTES", 60*24)) * time.Minute,
		CookieName:   envStr("SESSION_COOKIE_NAME", "dietnow_session"),
		CookieSecure: envBool("COOKIE_SECURE", false),
		CookieDomain: envStr("COOKIE_DOMAIN", ""),
		BcryptCost:   envInt("BCRYPT_COST", 12),
	}
}

// Validate rejects insecure or broken configurations; production enables stricter rules.
func (c Config) Validate(production bool) error {
	if c.SessionTTL <= 0 || c.IdleTimeout <= 0 {
		return fmt.Errorf("SESSION_TTL_HOURS and SESSION_IDLE_MINUTES must be positive")
	}
	if c.BcryptCost < bcrypt.MinCost || c.BcryptCost > bcrypt.MaxCost {
		return fmt.Errorf("BCRYPT_COST must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}
	if strings.HasPrefix(c.CookieName, "__Host-") && (!c.CookieSecure || c.CookieDomain != "") {
		return fmt.Errorf("a __Host- cookie requires COOKIE_SECURE=true and an empty COOKIE_DOMAIN")
	}
	if production {
		if !c.CookieSecure {
			return fmt.Errorf("COOKIE_SECURE must be true outside local development")
		}
		if c.BcryptCost < 10 {
			return fmt.Errorf("BCRYPT_COST must be at least 10 outside local development")
		}
	}
	return nil
}

func envStr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key))); err == nil {
		return n
	}
	return def
}

func envBool(key string, def bool) bool {
	if b, err := strconv.ParseBool(strings.TrimSpace(os.Getenv(key))); err == nil {
		return b
	}
	return def
}
