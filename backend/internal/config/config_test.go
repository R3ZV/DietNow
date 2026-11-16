package config

import (
	"reflect"
	"testing"
)

func clearEnv(t *testing.T) {
	for _, k := range []string{"PORT", "API_PORT", "APP_ENV", "API_ALLOWED_ORIGINS", "STATIC_DIR", "DB_AUTO_MIGRATE"} {
		t.Setenv(k, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 8080 || cfg.Env != "development" || cfg.Addr() != ":8080" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if len(cfg.AllowedOrigins) != 0 || cfg.StaticDir != "" || !cfg.AutoMigrate || cfg.Strict() {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadFromEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv("API_PORT", "9090")
	t.Setenv("APP_ENV", "Staging")
	t.Setenv("API_ALLOWED_ORIGINS", " https://a.example/ , ,https://b.example")
	t.Setenv("STATIC_DIR", "/app/web")
	t.Setenv("DB_AUTO_MIGRATE", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 9090 || cfg.Env != "staging" || !cfg.Strict() || cfg.StaticDir != "/app/web" || cfg.AutoMigrate {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	want := []string{"https://a.example", "https://b.example"}
	if !reflect.DeepEqual(cfg.AllowedOrigins, want) {
		t.Fatalf("AllowedOrigins = %v, want %v", cfg.AllowedOrigins, want)
	}
}

func TestPlatformPortWins(t *testing.T) {
	clearEnv(t)
	t.Setenv("API_PORT", "9090")
	t.Setenv("PORT", "10000")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 10000 {
		t.Fatalf("Port = %d, want 10000 from PORT", cfg.Port)
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	for _, tc := range []struct{ key, val string }{
		{"API_PORT", "abc"}, {"API_PORT", "0"}, {"API_PORT", "70000"}, {"PORT", "-1"},
		{"APP_ENV", "prod"}, {"DB_AUTO_MIGRATE", "maybe"},
	} {
		clearEnv(t)
		t.Setenv(tc.key, tc.val)
		if _, err := Load(); err == nil {
			t.Errorf("%s=%q: expected error", tc.key, tc.val)
		}
	}
}
