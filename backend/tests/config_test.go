package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/greenmetrics/backend/internal/config"
)

// TestConfig_RefusesDefaultJWTInProduction verifies the production hard gate.
func TestConfig_RefusesDefaultJWTInProduction(t *testing.T) {
	old := snapshotEnv("APP_ENV", "JWT_SECRET", "GRAFANA_ADMIN_PASSWORD", "DATABASE_URL")
	defer restoreEnv(old)

	os.Setenv("APP_ENV", "production")
	os.Setenv("JWT_SECRET", config.DefaultSentinelJWTSecret)
	os.Setenv("GRAFANA_ADMIN_PASSWORD", "rotate-me-32-bytes")
	os.Setenv("DATABASE_URL", "postgres://u:p@h:5432/db?sslmode=require")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error refusing default JWT_SECRET in production")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("expected JWT_SECRET in error, got %v", err)
	}
}

// TestConfig_RefusesDefaultGrafanaInProduction.
func TestConfig_RefusesDefaultGrafanaInProduction(t *testing.T) {
	old := snapshotEnv("APP_ENV", "JWT_SECRET", "GRAFANA_ADMIN_PASSWORD", "DATABASE_URL")
	defer restoreEnv(old)

	os.Setenv("APP_ENV", "production")
	os.Setenv("JWT_SECRET", "long-enough-jwt-secret-of-at-least-32-bytes-ok")
	os.Setenv("GRAFANA_ADMIN_PASSWORD", config.DefaultSentinelGrafanaPassword)
	os.Setenv("DATABASE_URL", "postgres://u:p@h:5432/db?sslmode=require")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error refusing default GRAFANA_ADMIN_PASSWORD in production")
	}
	if !strings.Contains(err.Error(), "GRAFANA_ADMIN_PASSWORD") {
		t.Fatalf("expected GRAFANA_ADMIN_PASSWORD in error, got %v", err)
	}
}

// TestConfig_RefusesSSLModeDisableInProduction.
func TestConfig_RefusesSSLModeDisableInProduction(t *testing.T) {
	old := snapshotEnv("APP_ENV", "JWT_SECRET", "GRAFANA_ADMIN_PASSWORD", "DATABASE_URL")
	defer restoreEnv(old)

	os.Setenv("APP_ENV", "production")
	os.Setenv("JWT_SECRET", "long-enough-jwt-secret-of-at-least-32-bytes-ok")
	os.Setenv("GRAFANA_ADMIN_PASSWORD", "rotate-me-32-bytes-secret-value")
	os.Setenv("DATABASE_URL", "postgres://u:p@h:5432/db?sslmode=disable")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error refusing sslmode=disable in production")
	}
	if !strings.Contains(err.Error(), "sslmode=disable") {
		t.Fatalf("expected sslmode=disable in error, got %v", err)
	}
}

// TestConfig_DevelopmentAllowsDefaults.
func TestConfig_DevelopmentAllowsDefaults(t *testing.T) {
	old := snapshotEnv("APP_ENV", "JWT_SECRET", "GRAFANA_ADMIN_PASSWORD", "DATABASE_URL")
	defer restoreEnv(old)

	os.Setenv("APP_ENV", "development")
	os.Setenv("JWT_SECRET", "")
	os.Setenv("GRAFANA_ADMIN_PASSWORD", "")
	os.Setenv("DATABASE_URL", "")

	if _, err := config.Load(); err != nil {
		t.Fatalf("development should tolerate defaults, got %v", err)
	}
}

func snapshotEnv(keys ...string) map[string]*string {
	out := map[string]*string{}
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			vv := v
			out[k] = &vv
		} else {
			out[k] = nil
		}
	}
	return out
}

func restoreEnv(snap map[string]*string) {
	for k, v := range snap {
		if v == nil {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, *v)
		}
	}
}
