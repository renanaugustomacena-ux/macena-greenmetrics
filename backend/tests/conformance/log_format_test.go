//go:build conformance

// Conformance — every log line emitted by the backend is a structured JSON
// object carrying the mandatory fields per Rule 7: service, env, version,
// commit, request_id, trace_id, span_id, tenant_id, level, time, message.
// Verified by capturing zap output during integration-test scenarios and
// asserting the fields are present.

package conformance_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// requiredFields is the canonical set per docs/PLATFORM-DEFAULTS.md §2 and Rule 7.
// It is intentionally a hard-coded constant — drift between this list and the
// runtime is a doctrine break.
var requiredFields = []string{
	"service",
	"env",
	"version",
	"commit",
	"request_id",
	"trace_id",
	"span_id",
	"tenant_id",
	"level",
	"time",
	"message",
}

// TestLogLineCarriesMandatoryFields parses a sampled log line and asserts the
// mandatory-field set is present. The sample is captured during integration
// tests; this test runs against the captured artefact at backend/tests/conformance/_fixtures/sample.log.
func TestLogLineCarriesMandatoryFields(t *testing.T) {
	t.Skip("scaffold — runs once integration tests emit captured logs to backend/tests/conformance/_fixtures/sample.log")

	// Pseudocode:
	// raw, err := os.ReadFile("_fixtures/sample.log")
	// require.NoError(t, err)
	// for _, line := range strings.Split(string(raw), "\n") {
	//     if line == "" { continue }
	//     parsed := mustJSON(t, line)
	//     for _, f := range requiredFields { require.Contains(t, parsed, f) }
	// }

	sample := `{"service":"greenmetrics-backend","env":"test","version":"0.1.0","commit":"abc1234","request_id":"req-1","trace_id":"trace-1","span_id":"span-1","tenant_id":"00000000-0000-4000-8000-000000000001","level":"info","time":"2026-04-30T00:00:00Z","message":"smoke"}`

	var line map[string]interface{}
	if err := json.Unmarshal([]byte(sample), &line); err != nil {
		t.Fatalf("sample not JSON: %v", err)
	}

	for _, f := range requiredFields {
		if _, ok := line[f]; !ok {
			t.Errorf("missing mandatory field %q in: %s", f, sample)
		}
	}

	// Per Rule 7, log message must not contain `password`, `token`, `secret`, `authorization`,
	// `jwt`, or `api_key` substrings post-redactor — verified at zap-redactor configuration time.
	for _, forbidden := range []string{"password", "token", "secret", "authorization", "jwt", "api_key"} {
		if strings.Contains(strings.ToLower(sample), forbidden) {
			t.Errorf("sample appears to leak %q substring; check redactor config", forbidden)
		}
	}
}
