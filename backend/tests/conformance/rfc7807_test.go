//go:build conformance

// Conformance — every error response is RFC 7807 ProblemDetails.
//
// Doctrine refs: Rule 14, Rule 25, Rule 45.
// Cross-portfolio invariant: CLAUDE.md.

package conformance_test

import (
	"encoding/json"
	"testing"
)

// problemShape mirrors the RFC 7807 minimum + GreenMetrics extensions.
type problemShape struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code,omitempty"`
}

// TestProblemShapeOnAllErrorResponses asserts that every Problem response
// contains required RFC 7807 fields. Run against a live ephemeral compose stack
// when integration env is provisioned (S5 follow-on).
func TestProblemShapeOnAllErrorResponses(t *testing.T) {
	t.Skip("scaffold — implement when integration fixture lands; calls live /api/v1/* with malformed inputs")

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		expect int
	}{
		{"login_invalid_email", "POST", "/api/v1/auth/login", `{"email":"not-an-email","password":"longenoughpassword"}`, 422},
		{"login_short_password", "POST", "/api/v1/auth/login", `{"email":"a@b.it","password":"short"}`, 422},
		{"ingest_no_idempotency_key", "POST", "/api/v1/readings/ingest", `{}`, 400},
		{"ingest_empty_body", "POST", "/api/v1/readings/ingest", ``, 400},
		{"meter_not_found", "GET", "/api/v1/meters/00000000-0000-4000-8000-000000000000", ``, 404},
		{"unauth", "GET", "/api/v1/meters", ``, 401},
		{"forbidden_role", "DELETE", "/api/v1/meters/00000000-0000-4000-8000-000000000000", ``, 403},
		{"too_large", "POST", "/api/v1/readings/ingest", largeBody(20 * 1024 * 1024), 413},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			body := callAPI(t, c.method, c.path, c.body)
			var p problemShape
			if err := json.Unmarshal(body, &p); err != nil {
				t.Fatalf("response not JSON: %v\n%s", err, body)
			}
			if p.Title == "" {
				t.Errorf("Problem.title empty")
			}
			if p.Status == 0 {
				t.Errorf("Problem.status missing")
			}
			if p.Status != c.expect {
				t.Errorf("Problem.status = %d; want %d", p.Status, c.expect)
			}
		})
	}
}

func callAPI(t *testing.T, method, path, body string) []byte {
	t.Helper()
	t.Fatalf("not implemented; ephemeral compose fixture lands in S5 follow-on")
	return nil
}

func largeBody(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}
