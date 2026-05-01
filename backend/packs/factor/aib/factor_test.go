package aib

import (
	"context"
	"strings"
	"testing"

	"github.com/greenmetrics/backend/internal/domain/emissions"
)

func TestPackImplementsFactorSource(t *testing.T) {
	var _ emissions.FactorSource = Pack
	if Pack.Name() != "aib" {
		t.Fatalf("name: want %q, got %q", "aib", Pack.Name())
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

// TestItalyResidualMixCovered2022To2025 — Italy is the primary use case.
func TestItalyResidualMixCovered2022To2025(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	years := map[int]bool{}
	for _, f := range factors {
		if f.Code == "it_aib_residual_mix" {
			years[f.ValidFromUTC.Year()] = true
		}
	}
	for _, y := range []int{2022, 2023, 2024, 2025} {
		if !years[y] {
			t.Errorf("Italy residual mix missing for year %d", y)
		}
	}
}

// TestAllSixCountriesShipped.
func TestAllSixCountriesShipped(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	countries := map[string]bool{}
	for _, f := range factors {
		if !strings.HasSuffix(f.Code, "_aib_residual_mix") {
			continue
		}
		country := strings.TrimSuffix(f.Code, "_aib_residual_mix")
		countries[country] = true
	}
	for _, c := range []string{"it", "de", "fr", "es", "at", "ch"} {
		if !countries[c] {
			t.Errorf("country %q missing from residual mix coverage", c)
		}
	}
}

// TestRenewableShareComplements — every country with residual mix also has
// renewable share for the same years.
func TestRenewableShareComplements(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	type key struct {
		country string
		year    int
	}
	mixKeys := map[key]bool{}
	shareKeys := map[key]bool{}
	for _, f := range factors {
		switch {
		case strings.HasSuffix(f.Code, "_aib_residual_mix"):
			country := strings.TrimSuffix(f.Code, "_aib_residual_mix")
			mixKeys[key{country: country, year: f.ValidFromUTC.Year()}] = true
		case strings.HasSuffix(f.Code, "_aib_renewable_share"):
			country := strings.TrimSuffix(f.Code, "_aib_renewable_share")
			shareKeys[key{country: country, year: f.ValidFromUTC.Year()}] = true
		}
	}
	for k := range mixKeys {
		if !shareKeys[k] {
			t.Errorf("country=%q year=%d has residual mix but no renewable share complement", k.country, k.year)
		}
	}
}

// TestTemporalValidityNonOverlapping — Rule 90.
func TestTemporalValidityNonOverlapping(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	type key struct {
		code string
		from int64
	}
	seen := map[key]bool{}
	for _, f := range factors {
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

// TestResidualMixSanityRange — values within plausible European range.
func TestResidualMixSanityRange(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	for _, f := range factors {
		if !strings.HasSuffix(f.Code, "_aib_residual_mix") {
			continue
		}
		if f.Value < 10 || f.Value > 800 {
			t.Errorf("residual mix %v g CO2eq/kWh outside plausible European range [10, 800] for code=%q valid_from=%v",
				f.Value, f.Code, f.ValidFromUTC)
		}
	}
}

// TestRenewableShareSanityRange — 0 ≤ share ≤ 100.
func TestRenewableShareSanityRange(t *testing.T) {
	factors, _ := Pack.Refresh(context.Background())
	for _, f := range factors {
		if !strings.HasSuffix(f.Code, "_aib_renewable_share") {
			continue
		}
		if f.Value < 0 || f.Value > 100 {
			t.Errorf("renewable share %v %% out of [0, 100] for code=%q valid_from=%v",
				f.Value, f.Code, f.ValidFromUTC)
		}
	}
}
