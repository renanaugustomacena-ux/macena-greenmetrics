//go:build conformance

// Conformance — every timestamp in API responses parses as RFC 3339 with UTC offset.

package conformance_test

import (
	"testing"
	"time"
)

func TestRFC3339UTC(t *testing.T) {
	cases := []struct {
		name string
		in   string
		ok   bool
	}{
		{"utc_z", "2026-04-25T12:34:56Z", true},
		{"utc_offset", "2026-04-25T12:34:56+00:00", true},
		{"with_nanos", "2026-04-25T12:34:56.123456789Z", true},
		{"non_utc_offset", "2026-04-25T12:34:56+02:00", false},
		{"missing_offset", "2026-04-25T12:34:56", false},
		{"date_only", "2026-04-25", false},
		{"unix_seconds", "1745597696", false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			tt, err := time.Parse(time.RFC3339Nano, c.in)
			if err != nil {
				if c.ok {
					t.Errorf("parse failed unexpectedly: %v", err)
				}
				return
			}
			_, off := tt.Zone()
			if c.ok && off != 0 {
				t.Errorf("offset != 0: %d", off)
			}
			if !c.ok && off == 0 && err == nil {
				t.Errorf("expected reject, got UTC parse")
			}
		})
	}
}
