package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type healthHandler struct{ d Dependencies }

func newHealthHandler(d Dependencies) *healthHandler { return &healthHandler{d: d} }

// HealthResponse is the exact JSON contract guaranteed by the spec.
type HealthResponse struct {
	Status          string            `json:"status"`
	Service         string            `json:"service"`
	Version         string            `json:"version"`
	UptimeSeconds   float64           `json:"uptime_seconds"`
	Time            string            `json:"time"`
	Dependencies    map[string]string `json:"dependencies"`
}

// Check handles GET /api/health.
//
// Shape (hard contract):
//
//	{
//	  "status": "ok",
//	  "service": "greenmetrics-backend",
//	  "version": "0.1.0",
//	  "uptime_seconds": 123.45,
//	  "time": "2026-04-17T10:00:00Z",
//	  "dependencies": {
//	    "timescaledb": "ok",
//	    "grafana": "ok"
//	  }
//	}
func (h *healthHandler) Check(c *fiber.Ctx) error {
	now := time.Now().UTC()
	uptime := now.Sub(h.d.StartedAt).Seconds()

	deps := map[string]string{
		"timescaledb": "unknown",
		"grafana":     "unknown",
	}

	// TimescaleDB.
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()
	if h.d.Repo != nil {
		if err := h.d.Repo.Ping(ctx); err != nil {
			deps["timescaledb"] = "degraded"
		} else {
			deps["timescaledb"] = "ok"
		}
	}

	// Grafana HTTP check.
	if h.d.Config.GrafanaURL != "" {
		client := &http.Client{Timeout: 2 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.d.Config.GrafanaURL+"/api/health", nil)
		if err == nil {
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode >= 500 {
				deps["grafana"] = "degraded"
			} else {
				deps["grafana"] = "ok"
			}
			if resp != nil {
				_ = resp.Body.Close()
			}
		} else {
			deps["grafana"] = "degraded"
		}
	}

	status := "ok"
	for _, v := range deps {
		if v == "degraded" || v == "unknown" {
			status = "degraded"
		}
	}

	return c.Status(fiber.StatusOK).JSON(HealthResponse{
		Status:        status,
		Service:       "greenmetrics-backend",
		Version:       h.d.Version,
		UptimeSeconds: uptime,
		Time:          now.Format(time.RFC3339),
		Dependencies:  deps,
	})
}

// Ready is stricter — fails if any dependency is not "ok".
func (h *healthHandler) Ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()
	if h.d.Repo == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready", "reason": "db_pool_not_initialised"})
	}
	if err := h.d.Repo.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready", "reason": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ready"})
}

// Live is a liveness probe — always 200 once the process is up.
func (h *healthHandler) Live(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "live"})
}
