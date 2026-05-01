package tests

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// TestDSARRequiresAuth verifies both DSAR endpoints reject unauthenticated
// requests. Either 401 (Bearer required) or 403 (CSRF token missing for the
// non-Bearer path) is acceptable — both signal "not authorised".
func TestDSARRequiresAuth(t *testing.T) {
	app, _ := buildTestApp(t)
	const tenantID = "00000000-0000-0000-0000-000000000001"
	const userID = "00000000-0000-0000-0000-000000000002"

	for _, path := range []string{
		"/api/v1/dsar/" + tenantID + "/" + userID + "/export",
		"/api/v1/dsar/" + tenantID + "/" + userID + "/erase",
	} {
		req := httptest.NewRequest("POST", path, nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("Test: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized && resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("%s: want 401 or 403, got %d", path, resp.StatusCode)
		}
	}
}

// TestDSAROnlyDPOCanInvoke verifies that non-DPO roles get 403.
func TestDSAROnlyDPOCanInvoke(t *testing.T) {
	app, _ := buildTestApp(t)
	const tenantID = "00000000-0000-0000-0000-000000000001"
	const userID = "00000000-0000-0000-0000-000000000002"
	path := "/api/v1/dsar/" + tenantID + "/" + userID + "/export"

	for _, role := range []string{"admin", "manager", "operator", "auditor", "viewer"} {
		t.Run(role, func(t *testing.T) {
			tok := signTestJWT(t, role, tenantID)
			req := httptest.NewRequest("POST", path, nil)
			req.Header.Set("Authorization", "Bearer "+tok)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Test: %v", err)
			}
			if resp.StatusCode != fiber.StatusForbidden {
				t.Errorf("role %q: want 403, got %d", role, resp.StatusCode)
			}
		})
	}
}

// TestDSARExportAcceptsDPO verifies the DPO role can invoke /export and
// gets 202 with a job-ID + 30-day SLA deadline.
func TestDSARExportAcceptsDPO(t *testing.T) {
	app, _ := buildTestApp(t)
	const tenantID = "00000000-0000-0000-0000-000000000001"
	const userID = "00000000-0000-0000-0000-000000000002"

	tok := signTestJWT(t, "dpo", tenantID)
	req := httptest.NewRequest("POST",
		"/api/v1/dsar/"+tenantID+"/"+userID+"/export", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusAccepted {
		t.Fatalf("DPO export: want 202, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc == "" {
		t.Error("Location header should be set on 202")
	}

	body, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, k := range []string{"job_id", "tenant_id", "user_id", "state", "requested_at", "sla_deadline"} {
		if _, ok := payload[k]; !ok {
			t.Errorf("response missing %q", k)
		}
	}
	if payload["state"] != "queued" {
		t.Errorf("state: want 'queued', got %v", payload["state"])
	}
	// SLA deadline should be 30 days after requested_at; spot-check the format.
	if _, ok := payload["sla_deadline"].(string); !ok {
		t.Errorf("sla_deadline should be a string")
	}
}

// TestDSAREraseAcceptsDPO mirrors the export test for /erase.
func TestDSAREraseAcceptsDPO(t *testing.T) {
	app, _ := buildTestApp(t)
	const tenantID = "00000000-0000-0000-0000-000000000001"
	const userID = "00000000-0000-0000-0000-000000000002"

	tok := signTestJWT(t, "dpo", tenantID)
	req := httptest.NewRequest("POST",
		"/api/v1/dsar/"+tenantID+"/"+userID+"/erase", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusAccepted {
		t.Fatalf("DPO erase: want 202, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["state"] != "queued" {
		t.Errorf("state: want 'queued', got %v", payload["state"])
	}
}

// TestDSARRejectsInvalidUUID verifies the params are validated.
func TestDSARRejectsInvalidUUID(t *testing.T) {
	app, _ := buildTestApp(t)

	tok := signTestJWT(t, "dpo", "00000000-0000-0000-0000-000000000001")
	cases := []struct {
		path string
		name string
	}{
		{"/api/v1/dsar/not-a-uuid/00000000-0000-0000-0000-000000000002/export", "bad-tenant"},
		{"/api/v1/dsar/00000000-0000-0000-0000-000000000001/not-a-uuid/erase", "bad-user"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+tok)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Test: %v", err)
			}
			if resp.StatusCode != fiber.StatusBadRequest {
				t.Errorf("%s: want 400, got %d", tc.path, resp.StatusCode)
			}
		})
	}
}

// signTestJWT signs a HS256 access JWT with the supplied role + tenant.
// Mirrors the production token format from internal/handlers/auth.go.
func signTestJWT(t *testing.T, role, tenantID string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub":       "test@greenmetrics.local",
		"tenant_id": tenantID,
		"role":      role,
		"exp":       time.Now().Add(15 * time.Minute).Unix(),
		"iat":       time.Now().Unix(),
		"typ":       "access",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok.Header["kid"] = "test"
	signed, err := tok.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}
