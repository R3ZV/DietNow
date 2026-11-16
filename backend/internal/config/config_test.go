package config

import (
	"reflect"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("API_PORT", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("API_ALLOWED_ORIGINS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 8080 || cfg.Env != "development" || cfg.Addr() != ":8080" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if len(cfg.AllowedOrigins) != 0 {
		t.Fatalf("expected no allowed origins, got %v", cfg.AllowedOrigins)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("API_PORT", "9090")
	t.Setenv("APP_ENV", "production")
	t.Setenv("API_ALLOWED_ORIGINS", " https://a.example , ,https://b.example")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 9090 || cfg.Env != "production" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	want := []string{"https://a.example", "https://b.example"}
	if !reflect.DeepEqual(cfg.AllowedOrigins, want) {
		t.Fatalf("AllowedOrigins = %v, want %v", cfg.AllowedOrigins, want)
	}
}

func TestLoadRejectsBadPort(t *testing.T) {
	for _, p := range []string{"abc", "0", "70000"} {
		t.Setenv("API_PORT", p)
		if _, err := Load(); err == nil {
			t.Errorf("API_PORT=%q: expected error", p)
		}
	}
}
