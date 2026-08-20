package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds application configuration loaded from environment
// variables. HIS base URLs for a given hospital are stored per-row on
// hospitals.his_base_url (see docs/er-diagram.md) rather than in config;
// HospitalABaseURL here is only the default used when seeding that row.
type Config struct {
	Port             string
	DBDSN            string
	JWTSecret        string
	JWTExpiry        time.Duration
	HospitalABaseURL string
}

const defaultJWTExpiry = time.Hour

// Load reads and validates configuration from the environment. DB_DSN and
// JWT_SECRET are required; everything else has a default.
func Load() (Config, error) {
	cfg := Config{
		Port:             os.Getenv("PORT"),
		DBDSN:            os.Getenv("DB_DSN"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		HospitalABaseURL: os.Getenv("HOSPITAL_A_BASE_URL"),
		JWTExpiry:        defaultJWTExpiry,
	}

	if cfg.DBDSN == "" {
		return Config{}, fmt.Errorf("DB_DSN is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.HospitalABaseURL == "" {
		cfg.HospitalABaseURL = "https://hospital-a.api.co.th"
	}
	if v := os.Getenv("JWT_EXPIRY"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid JWT_EXPIRY: %w", err)
		}
		cfg.JWTExpiry = d
	}

	return cfg, nil
}
