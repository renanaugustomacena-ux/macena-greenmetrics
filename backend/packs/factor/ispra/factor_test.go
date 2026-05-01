package ispra

import (
	"context"
	"testing"
	"time"

	"github.com/greenmetrics/backend/internal/domain/emissions"
)

// TestPackImplementsFactorSource is the compile-time + runtime check that
// the Pack satisfies the contract.
func TestPackImplementsFactorSource(t *testing.T) {
	var _ emissions.FactorSource = Pack
	if Pack.Name() != "ispra" {
		t.Fatalf("expected name=ispra, got %q", Pack.Name())
	}
}

// TestRefreshReturnsFactors verifies Refresh returns a non-empty set with
// the expected Scope 2 grid-mix rows.
func TestRefreshReturnsFactors(t *testing.T) {
	ctx := context.Background()
	factors, err := Pack.Refresh(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(factors) < 5 {
		t.Errorf("expected at least 5 factors, got %d", len(factors))
	}

	// Spot-check: 2024 grid-mix should be 233 g/kWh per Rapporto 404/2025.
	t2024 := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	v, ok := lookup(factors, "it_grid_mix_location", t2024)
	if !ok {
		t.Fatal("it_grid_mix_location should have a row covering 2024-06-15")
	}
	if v.Value != 233.0 {
		t.Errorf("2024 grid mix: want 233, got %v", v.Value)
	}
	if v.Unit != "g CO2eq/kWh" {
		t.Errorf("2024 grid mix unit: want g CO2eq/kWh, got %q", v.Unit)
	}
}

// TestRefreshHonoursContext verifies Refresh exits promptly on context
// cancellation.
func TestRefreshHonoursContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	_, err := Pack.Refresh(ctx)
	if err == nil {
		t.Error("expected ctx.Err() on cancelled context")
	}
}

// TestEveryFactorCarriesSource enforces Rule 132 — primary-source citation
// on every row.
func TestEveryFactorCarriesSource(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	for _, f := range factors {
		if f.Source == "" {
			t.Errorf("factor %s @ %s: Source must not be empty (Rule 132)",
				f.Code, f.ValidFromUTC.Format("2006-01-02"))
		}
		if f.Unit == "" {
			t.Errorf("factor %s @ %s: Unit must not be empty",
				f.Code, f.ValidFromUTC.Format("2006-01-02"))
		}
	}
}

// TestTemporalKeyUniqueness enforces (code, valid_from) being a primary key
// per Rule 90.
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

// TestValidIntervals enforces valid_from < valid_to (when set) and that
// successive rows for the same code are non-overlapping.
func TestValidIntervals(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	byCode := map[string][]emissions.Factor{}
	for _, f := range factors {
		byCode[f.Code] = append(byCode[f.Code], f)
	}
	for code, rows := range byCode {
		// Sort by valid_from.
		for i := 0; i < len(rows); i++ {
			for j := i + 1; j < len(rows); j++ {
				if rows[j].ValidFromUTC.Before(rows[i].ValidFromUTC) {
					rows[i], rows[j] = rows[j], rows[i]
				}
			}
		}
		for _, r := range rows {
			if !r.ValidToUTC.IsZero() && !r.ValidFromUTC.Before(r.ValidToUTC) {
				t.Errorf("%s: valid_from (%s) must be before valid_to (%s)",
					code, r.ValidFromUTC, r.ValidToUTC)
			}
		}
		for i := 1; i < len(rows); i++ {
			prev, cur := rows[i-1], rows[i]
			if !prev.ValidToUTC.IsZero() && cur.ValidFromUTC.Before(prev.ValidToUTC) {
				// Non-overlapping required for code with multiple rows.
				t.Errorf("%s: row %d valid_from (%s) overlaps prev valid_to (%s)",
					code, i, cur.ValidFromUTC, prev.ValidToUTC)
			}
		}
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
