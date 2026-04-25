package tests

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/greenmetrics/backend/internal/config"
	"github.com/greenmetrics/backend/internal/handlers"
)

// TestHealthEndpoint verifies the /api/health JSON contract.
func TestHealthEndpoint(t *testing.T) {
	app, _ := buildTestApp(t)

	req := httptest.NewRequest("GET", "/api/health", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, k := range []string{"status", "service", "version", "uptime_seconds", "time", "dependencies"} {
		if _, ok := payload[k]; !ok {
			t.Errorf("health response missing %q", k)
		}
	}
	deps, ok := payload["dependencies"].(map[string]any)
	if !ok {
		t.Fatalf("dependencies not a map")
	}
	for _, d := range []string{"timescaledb", "grafana"} {
		if _, ok := deps[d]; !ok {
			t.Errorf("dependencies missing %q", d)
		}
	}
	if payload["service"].(string) != "greenmetrics-backend" {
		t.Errorf("service mismatch: %v", payload["service"])
	}
}

// TestJWTMiddlewareRejectsMissing verifies unauthorized responses.
func TestJWTMiddlewareRejectsMissing(t *testing.T) {
	app, _ := buildTestApp(t)

	cases := []struct {
		name string
		path string
	}{
		{"meters", "/api/v1/meters"},
		{"readings", "/api/v1/readings?meter_id=foo&from=2026-01-01T00:00:00Z&to=2026-01-02T00:00:00Z"},
		{"reports", "/api/v1/reports"},
		{"alerts", "/api/v1/alerts"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Test: %v", err)
			}
			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Errorf("expected 401 for %s, got %d", tc.path, resp.StatusCode)
			}
		})
	}
}

// TestLoginValidatesInput asserts basic input validation.
func TestLoginValidatesInput(t *testing.T) {
	app, _ := buildTestApp(t)
	cases := []struct {
		body string
		code int
	}{
		{`{}`, fiber.StatusBadRequest},
		{`{"email":"a@b.it","password":""}`, fiber.StatusBadRequest},
		{`{"email":"","password":"x"}`, fiber.StatusBadRequest},
	}
	for _, tc := range cases {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(tc.body))
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("test: %v", err)
		}
		if resp.StatusCode != tc.code {
			t.Errorf("body=%s expected %d got %d", tc.body, tc.code, resp.StatusCode)
		}
	}
}

func buildTestApp(t *testing.T) (*fiber.App, *zap.Logger) {
	t.Helper()
	cfg := &config.Config{
		AppEnv:             "test",
		AppPort:            "8080",
		JWTSecret:          "test-secret",
		JWTAccessTTL:       15 * time.Minute,
		JWTRefreshTTL:      time.Hour,
		CORSAllowedOrigins: "*",
	}
	logger, _ := zap.NewDevelopment()
	app := fiber.New(fiber.Config{ErrorHandler: handlers.ErrorHandler(logger)})
	handlers.Register(app, handlers.Dependencies{
		Config:    cfg,
		Logger:    logger,
		StartedAt: time.Now().UTC(),
		Version:   "test",
		Commit:    "test",
	})
	return app, logger
}

// _ retains io import via blank usage if needed elsewhere.
var _ = io.EOF
