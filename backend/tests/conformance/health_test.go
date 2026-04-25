//go:build conformance

// Conformance — /api/health response shape matches CLAUDE.md invariant.

package conformance_test

import (
	"encoding/json"
	"testing"
	"time"
)

type healthShape struct {
	Status        string                 `json:"status"`
	Service       string                 `json:"service"`
	Version       string                 `json:"version"`
	UptimeSeconds int                    `json:"uptime_seconds"`
	Time          string                 `json:"time"`
	Dependencies  map[string]interface{} `json:"dependencies"`
}

func TestHealthEnvelope(t *testing.T) {
	t.Skip("scaffold — calls live /api/health when integration fixture lands")

	body := callAPI(t, "GET", "/api/health", "")
	var h healthShape
	if err := json.Unmarshal(body, &h); err != nil {
		t.Fatalf("response not JSON: %v\n%s", err, body)
	}
	if h.Status == "" {
		t.Errorf("status empty")
	}
	if h.Service != "greenmetrics-backend" {
		t.Errorf("service = %q; want greenmetrics-backend", h.Service)
	}
	if h.Version == "" {
		t.Errorf("version empty")
	}
	if _, err := time.Parse(time.RFC3339, h.Time); err != nil {
		t.Errorf("time not RFC 3339: %v", err)
	}
	if h.Dependencies == nil {
		t.Errorf("dependencies missing")
	}
}
