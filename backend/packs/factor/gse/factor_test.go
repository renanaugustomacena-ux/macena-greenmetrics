package gse

import (
	"context"
	"testing"
	"time"

	"github.com/greenmetrics/backend/internal/domain/emissions"
)

func TestPackImplementsFactorSource(t *testing.T) {
	var _ emissions.FactorSource = Pack
	if Pack.Name() != "gse" {
		t.Fatalf("expected name=gse, got %q", Pack.Name())
	}
}

func TestRefreshReturnsAIBResidualMix(t *testing.T) {
	factors, err := Pack.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	t2024 := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	v, ok := lookup(factors, "it_aib_residual_mix", t2024)
	if !ok {
		t.Fatal("it_aib_residual_mix should have a row covering 2024-06-15")
	}
	if v.Value != 332.0 {
		t.Errorf("2024 AIB residual mix: want 332, got %v", v.Value)
	}
	if v.Unit != "g CO2eq/kWh" {
		t.Errorf("unit: want g CO2eq/kWh, got %q", v.Unit)
	}
}

func TestRefreshReturnsRenewableShare(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	t2025 := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	v, ok := lookup(factors, "it_renewable_share", t2025)
	if !ok {
		t.Fatal("it_renewable_share should have a row covering 2025-06-15")
	}
	if v.Value != 44.5 {
		t.Errorf("2025 renewable share: want 44.5, got %v", v.Value)
	}
	if v.Unit != "%" {
		t.Errorf("unit: want %%, got %q", v.Unit)
	}
}

func TestRefreshHonoursContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Pack.Refresh(ctx); err == nil {
		t.Error("expected ctx.Err() on cancelled context")
	}
}

func TestEveryFactorCarriesSource(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	if len(factors) < 10 {
		t.Errorf("expected at least 10 factors, got %d", len(factors))
	}
	for _, f := range factors {
		if f.Source == "" {
			t.Errorf("%s @ %s: Source must not be empty (Rule 132)",
				f.Code, f.ValidFromUTC.Format("2006-01-02"))
		}
		if f.Unit == "" {
			t.Errorf("%s @ %s: Unit must not be empty",
				f.Code, f.ValidFromUTC.Format("2006-01-02"))
		}
	}
}

func TestTemporalKeyUniqueness(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	seen := make(map[string]bool)
	for _, f := range factors {
		k := f.Code + "@" + f.ValidFromUTC.Format(time.RFC3339)
		if seen[k] {
			t.Errorf("duplicate (code, valid_from) tuple: %s", k)
		}
		seen[k] = true
	}
}

// lookup returns the factor row whose [valid_from, valid_to) interval
// covers `at`, or zero + false if none.
func lookup(factors []emissions.Factor, code string, at time.Time) (emissions.Factor, bool) {
	for _, f := range factors {
		if f.Code != code {
			continue
		}
		if at.Before(f.ValidFromUTC) {
			continue
		}
		if !f.ValidToUTC.IsZero() && !at.Before(f.ValidToUTC) {
			continue
		}
		return f, true
	}
	return emissions.Factor{}, false
}
