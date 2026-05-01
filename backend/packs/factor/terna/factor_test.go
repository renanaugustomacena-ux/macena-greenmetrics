package terna

import (
	"context"
	"testing"

	"github.com/greenmetrics/backend/internal/domain/emissions"
)

func TestPackImplementsFactorSource(t *testing.T) {
	var _ emissions.FactorSource = Pack
	if Pack.Name() != "terna" {
		t.Fatalf("name: want %q, got %q", "terna", Pack.Name())
	}
}

func TestRefreshReturnsStaticFactors(t *testing.T) {
	factors, err := Pack.Refresh(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(factors) == 0 {
		t.Fatal("expected non-empty factor table")
	}
}

// TestRefreshHonoursContext.
func TestRefreshHonoursContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Pack.Refresh(ctx); err == nil {
		t.Error("expected ctx.Err() on cancelled context")
	}
}

// TestEveryRowCarriesPrimarySource — Rule 132.
func TestEveryRowCarriesPrimarySource(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	for i, f := range factors {
		if f.Source == "" {
			t.Errorf("row %d code=%q missing Source annotation (Rule 132)", i, f.Code)
		}
		if f.SourceURL == "" {
			t.Errorf("row %d code=%q missing SourceURL annotation", i, f.Code)
		}
	}
}

// TestMonthlyFactorTemporalValidity — Rule 90: each (code, month) has exactly
// one valid factor; non-overlapping intervals.
func TestMonthlyFactorTemporalValidity(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	type key struct {
		code string
		from int64
	}
	seen := map[key]bool{}
	for _, f := range factors {
		if f.Code != "it_grid_mix_terna_monthly" {
			continue
		}
		k := key{code: f.Code, from: f.ValidFromUTC.Unix()}
		if seen[k] {
			t.Errorf("duplicate temporal entry for code=%q valid_from=%v", f.Code, f.ValidFromUTC)
		}
		seen[k] = true
		if !f.ValidToUTC.After(f.ValidFromUTC) {
			t.Errorf("non-monotonic interval for code=%q: from=%v to=%v",
				f.Code, f.ValidFromUTC, f.ValidToUTC)
		}
	}
}

// TestMonthlySeriesCoversFullYear — every year shipped has 12 monthly entries.
func TestMonthlySeriesCoversFullYear(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	yearCounts := map[int]int{}
	for _, f := range factors {
		if f.Code != "it_grid_mix_terna_monthly" {
			continue
		}
		yearCounts[f.ValidFromUTC.Year()]++
	}
	for y, n := range yearCounts {
		if n != 12 {
			t.Errorf("year %d has %d monthly entries, want 12", y, n)
		}
	}
	if len(yearCounts) < 3 {
		t.Errorf("expected ≥3 years (2024-2026), got %d", len(yearCounts))
	}
}

// TestFactorRangeSanity — values within plausible Italian grid range.
func TestFactorRangeSanity(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	for _, f := range factors {
		if f.Code != "it_grid_mix_terna_monthly" {
			continue
		}
		if f.Value < 50 || f.Value > 500 {
			t.Errorf("grid factor %v g CO2eq/kWh outside plausible Italian range [50, 500] for code=%q valid_from=%v",
				f.Value, f.Code, f.ValidFromUTC)
		}
	}
}
