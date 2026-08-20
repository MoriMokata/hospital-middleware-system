package config

import (
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"PORT", "DB_DSN", "JWT_SECRET", "HOSPITAL_A_BASE_URL", "JWT_EXPIRY"} {
		t.Setenv(key, "")
	}
}

func TestLoad_AllEnvPresent(t *testing.T) {
	clearEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("DB_DSN", "postgres://user:pass@host:5432/db")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("HOSPITAL_A_BASE_URL", "https://hospital-a.example.com")
	t.Setenv("JWT_EXPIRY", "30m")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.DBDSN != "postgres://user:pass@host:5432/db" {
		t.Errorf("DBDSN = %q", cfg.DBDSN)
	}
	if cfg.JWTSecret != "supersecret" {
		t.Errorf("JWTSecret = %q", cfg.JWTSecret)
	}
	if cfg.HospitalABaseURL != "https://hospital-a.example.com" {
		t.Errorf("HospitalABaseURL = %q", cfg.HospitalABaseURL)
	}
	if cfg.JWTExpiry != 30*time.Minute {
		t.Errorf("JWTExpiry = %v, want 30m", cfg.JWTExpiry)
	}
}

func TestLoad_DefaultsAppliedWhenOptionalEnvMissing(t *testing.T) {
	clearEnv(t)
	t.Setenv("DB_DSN", "postgres://user:pass@host:5432/db")
	t.Setenv("JWT_SECRET", "supersecret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want default 8080", cfg.Port)
	}
	if cfg.HospitalABaseURL != "https://hospital-a.api.co.th" {
		t.Errorf("HospitalABaseURL = %q, want default", cfg.HospitalABaseURL)
	}
	if cfg.JWTExpiry != defaultJWTExpiry {
		t.Errorf("JWTExpiry = %v, want default %v", cfg.JWTExpiry, defaultJWTExpiry)
	}
}

func TestLoad_MissingDBDSN(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_SECRET", "supersecret")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected error for missing DB_DSN, got nil")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	clearEnv(t)
	t.Setenv("DB_DSN", "postgres://user:pass@host:5432/db")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected error for missing JWT_SECRET, got nil")
	}
}

func TestLoad_InvalidJWTExpiry(t *testing.T) {
	clearEnv(t)
	t.Setenv("DB_DSN", "postgres://user:pass@host:5432/db")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("JWT_EXPIRY", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected error for invalid JWT_EXPIRY, got nil")
	}
}
