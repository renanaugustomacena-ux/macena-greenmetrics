//go:build conformance

// Conformance — every event published by the system carries a CloudEvents 1.0
// envelope with `specversion`, `type`, `source`, `id`, `time`, `datacontenttype`,
// `subject`, and a typed `data` payload. Doctrine refs: Rule 5, Rule 22.

package conformance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// envelope models the required CloudEvents 1.0 attributes for our wire format.
type envelope struct {
	SpecVersion     string          `json:"specversion"`
	Type            string          `json:"type"`
	Source          string          `json:"source"`
	ID              string          `json:"id"`
	Time            string          `json:"time"`
	DataContentType string          `json:"datacontenttype"`
	Subject         string          `json:"subject"`
	Data            json.RawMessage `json:"data"`
}

// TestCloudEventsFixturesValidate walks docs/contracts/events/ and asserts every
// JSON fixture there parses as a CloudEvents 1.0 envelope with all required
// attributes. The fixture directory is the source of truth; the dispatcher in
// internal/services/event_bus.go (Phase E Sprint S6) is verified against it.
func TestCloudEventsFixturesValidate(t *testing.T) {
	t.Skip("scaffold — runs once docs/contracts/events/*.json fixtures land in Sprint S6")

	root := "../../../docs/contracts/events"
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("event fixture root: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("no event fixtures found — at least one schema + example must exist before Sprint S6 exit")
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".example.json") {
			continue
		}
		t.Run(e.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, e.Name()))
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			var env envelope
			if err := json.Unmarshal(data, &env); err != nil {
				t.Fatalf("not JSON: %v", err)
			}
			if env.SpecVersion != "1.0" {
				t.Errorf("specversion=%q; want 1.0 (RFC 9322 / CloudEvents 1.0)", env.SpecVersion)
			}
			if env.Type == "" {
				t.Error("type empty — required by CloudEvents 1.0 §3.1")
			}
			if env.Source == "" {
				t.Error("source empty — required by CloudEvents 1.0 §3.1")
			}
			if env.ID == "" {
				t.Error("id empty — required by CloudEvents 1.0 §3.1")
			}
			if env.Time != "" {
				if _, err := time.Parse(time.RFC3339, env.Time); err != nil {
					t.Errorf("time not RFC 3339: %v", err)
				}
			}
			if env.Subject == "" {
				t.Error("subject empty — required by GreenMetrics doctrine Rule 5 (tighter than CloudEvents)")
			}
			if env.DataContentType == "" {
				t.Error("datacontenttype empty — required by GreenMetrics doctrine Rule 5")
			}
			if len(env.Data) == 0 {
				t.Error("data missing — required by GreenMetrics doctrine Rule 5")
			}
		})
	}
}
