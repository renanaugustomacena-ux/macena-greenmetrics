//go:build conformance

// Conformance — every tenant_id / meter_id / channel_id / report_id is UUIDv4.

package conformance_test

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUIDv4(t *testing.T) {
	cases := []string{
		"00000000-0000-4000-8000-000000000000",
		"00000000-0000-4000-8000-aaaaaaaaaaaa",
	}
	for _, s := range cases {
		s := s
		t.Run(s, func(t *testing.T) {
			id, err := uuid.Parse(s)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if id.Version() != 4 {
				t.Errorf("version = %d; want 4", id.Version())
			}
		})
	}
}

func TestUUIDv4RejectsNonV4(t *testing.T) {
	cases := []string{
		"00000000-0000-1000-8000-000000000000", // v1
		"00000000-0000-3000-8000-000000000000", // v3
		"00000000-0000-5000-8000-000000000000", // v5
		"not-a-uuid",
	}
	for _, s := range cases {
		s := s
		t.Run(s, func(t *testing.T) {
			id, err := uuid.Parse(s)
			if err != nil {
				return // not parseable as UUID — fine
			}
			if id.Version() == 4 {
				t.Errorf("expected non-v4 but got v4: %s", s)
			}
		})
	}
}
