//go:build security

// Boot-refusal test — production env must reject sentinel JWT secret.
//
// Doctrine refs: Rule 19, Rule 39.
// Mitigates: RISK-013.

package security_test

import (
	"os"
	"strings"
	"testing"

	"github.com/greenmetrics/backend/internal/config"
)

func TestProductionRefusesSentinelJWTSecret(t *testing.T) {
	withEnv(t, map[string]string{
		"APP_ENV":                  "production",
		"DATABASE_URL":             "postgres://app:app@db:5432/app?sslmode=require",
		"JWT_SECRET":               config.DefaultSentinelJWTSecret,
		"GRAFANA_ADMIN_PASSWORD":   "any-non-default-32-char-or-longer-pwd",
	}, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected config.Load() to refuse sentinel JWT secret in production; got nil error")
		}
		if !strings.Contains(err.Error(), "JWT_SECRET") {
			t.Errorf("error did not mention JWT_SECRET: %v", err)
		}
	})
}

func TestProductionRefusesShortJWTSecret(t *testing.T) {
	withEnv(t, map[string]string{
		"APP_ENV":                "production",
		"DATABASE_URL":           "postgres://app:app@db:5432/app?sslmode=require",
		"JWT_SECRET":             "short", // < 32 bytes
		"GRAFANA_ADMIN_PASSWORD": "any-non-default-32-char-or-longer-pwd",
	}, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected config.Load() to refuse short JWT secret in production; got nil")
		}
	})
}

func TestProductionRefusesDefaultGrafanaPassword(t *testing.T) {
	withEnv(t, map[string]string{
		"APP_ENV":                "production",
		"DATABASE_URL":           "postgres://app:app@db:5432/app?sslmode=require",
		"JWT_SECRET":             "this-is-a-perfectly-fine-32-byte-secret-for-tests-only",
		"GRAFANA_ADMIN_PASSWORD": config.DefaultSentinelGrafanaPassword,
	}, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected config.Load() to refuse default Grafana password in production; got nil")
		}
		if !strings.Contains(err.Error(), "GRAFANA_ADMIN_PASSWORD") {
			t.Errorf("error did not mention GRAFANA_ADMIN_PASSWORD: %v", err)
		}
	})
}

func TestProductionRefusesSslModeDisable(t *testing.T) {
	withEnv(t, map[string]string{
		"APP_ENV":                "production",
		"DATABASE_URL":           "postgres://app:app@db:5432/app?sslmode=disable",
		"JWT_SECRET":             "this-is-a-perfectly-fine-32-byte-secret-for-tests-only",
		"GRAFANA_ADMIN_PASSWORD": "any-non-default-32-char-or-longer-pwd",
	}, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected config.Load() to refuse sslmode=disable in production; got nil")
		}
		if !strings.Contains(err.Error(), "sslmode=disable") {
			t.Errorf("error did not mention sslmode=disable: %v", err)
		}
	})
}

func TestDevAcceptsDefaults(t *testing.T) {
	withEnv(t, map[string]string{
		"APP_ENV":      "development",
		"DATABASE_URL": "postgres://app:app@db:5432/app?sslmode=disable",
		// JWT_SECRET defaults to sentinel; expected to be tolerated in dev.
		"JWT_SECRET":             config.DefaultSentinelJWTSecret,
		"GRAFANA_ADMIN_PASSWORD": config.DefaultSentinelGrafanaPassword,
	}, func() {
		_, err := config.Load()
		if err != nil {
			t.Fatalf("dev env unexpectedly rejected defaults: %v", err)
		}
	})
}

// --- helpers ----------------------------------------------------------------

func withEnv(t *testing.T, env map[string]string, fn func()) {
	t.Helper()
	saved := map[string]string{}
	for k := range env {
		saved[k] = os.Getenv(k)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	})
	for k, v := range env {
		_ = os.Setenv(k, v)
	}
	fn()
}
