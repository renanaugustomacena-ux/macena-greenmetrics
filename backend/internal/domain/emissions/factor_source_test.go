// Example test for the Factor Pack contract. Per Rule 86.

package emissions_test

import (
	"context"
	"testing"
	"time"

	"github.com/greenmetrics/backend/internal/domain/emissions"
)

// stubFactorSource models a minimal Factor Pack returning a deterministic factor set.
type stubFactorSource struct{}

func (stubFactorSource) Name() string { return "stub" }

func (stubFactorSource) Refresh(_ context.Context) ([]emissions.Factor, error) {
	return []emissions.Factor{
		{
			Code:         "national_mix_2024",
			ValidFromUTC: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ValidToUTC:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Value:        0.245,
			Unit:         "kgCO2e/kWh",
			Source:       "test",
		},
		{
			Code:         "national_mix_2025",
			ValidFromUTC: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Value:        0.230,
			Unit:         "kgCO2e/kWh",
			Source:       "test",
		},
	}, nil
}

func TestExample_FactorSourceReturnsTemporalSet(t *testing.T) {
	var fs emissions.FactorSource = stubFactorSource{}

	factors, err := fs.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if len(factors) != 2 {
		t.Fatalf("expected 2 factors; got %d", len(factors))
	}

	// Per Rule 90: temporal versioning — `valid_to` exclusive, `valid_from` inclusive.
	if !factors[0].ValidToUTC.Equal(factors[1].ValidFromUTC) {
		t.Errorf("temporal interval discontinuity: factor[0].valid_to=%v, factor[1].valid_from=%v",
			factors[0].ValidToUTC, factors[1].ValidFromUTC)
	}

	if fs.Name() == "" {
		t.Error("Name must not be empty")
	}
}

func TestContractVersion_IsSet(t *testing.T) {
	if emissions.ContractVersion == "" {
		t.Fatal("ContractVersion empty — Rule 71 requires per-kind contract version")
	}
}
