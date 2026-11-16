package auth

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	base := Config{SessionTTL: time.Hour, IdleTimeout: time.Minute, CookieName: "dietnow_session", BcryptCost: 12}

	if err := base.Validate(false); err != nil {
		t.Errorf("local defaults rejected: %v", err)
	}
	if err := base.Validate(true); err == nil {
		t.Error("staging without COOKIE_SECURE must be rejected")
	}

	secure := base
	secure.CookieSecure = true
	if err := secure.Validate(true); err != nil {
		t.Errorf("secure staging config rejected: %v", err)
	}

	weak := secure
	weak.BcryptCost = 4
	if err := weak.Validate(true); err == nil {
		t.Error("low bcrypt cost must be rejected outside local")
	}

	host := base
	host.CookieName = "__Host-dietnow_session"
	if err := host.Validate(false); err == nil {
		t.Error("__Host- cookie without Secure must be rejected")
	}
}

func TestValidatePassword(t *testing.T) {
	for pw, ok := range map[string]bool{
		"1234567":                false, // < 8
		"12345678":               true,
		"pässwörd":               true,  // 8 runes
		string(make([]byte, 73)): false, // > bcrypt's 72-byte limit
	} {
		if got := ValidatePassword(pw) == nil; got != ok {
			t.Errorf("ValidatePassword(len=%d) ok=%v, want %v", len(pw), got, ok)
		}
	}
}
