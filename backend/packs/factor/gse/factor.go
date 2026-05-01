// Package gse implements the GSE Factor Pack — Italian renewable-energy
// mix shares (GSE Rapporto Statistico FER) plus the AIB European Residual
// Mix entrypoint published via GSE's "Energia statistica" channel.
//
// Phase E ships the static factor table; Phase G adds the live-fetch path
// against areaclienti.gse.it with circuit breaker per ADR-0009.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/emissions/factor_source.go
//   - Manifest:         packs/factor/gse/manifest.yaml
//   - Charter:          packs/factor/gse/CHARTER.md
package gse

import (
	"context"
	"time"

	"github.com/greenmetrics/backend/internal/domain/emissions"
)

// Pack is the singleton instance constructed at boot.
var Pack emissions.FactorSource = &source{}

// PackVersion is the Pack's SemVer (matches manifest.yaml version).
const PackVersion = "1.0.0"

// source is the concrete FactorSource implementation.
type source struct{}

// Name implements emissions.FactorSource.
func (s *source) Name() string { return "gse" }

// Refresh implements emissions.FactorSource. Phase E returns the
// checked-in static table; Phase G replaces this with a live fetch
// against areaclienti.gse.it.
func (s *source) Refresh(ctx context.Context) ([]emissions.Factor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return staticFactors(), nil
}

// staticFactors returns the Phase-E checked-in factor set. Each row's
// Source annotation cites the primary regulator publication (Rule 132).
func staticFactors() []emissions.Factor {
	mustParse := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t
	}

	const aibSource = "AIB European Residual Mixes (published via GSE)"
	const aibURL = "https://www.aib-net.org/facts/european-residual-mix"
	const fer = "GSE Rapporto Statistico FER"
	const ferURL = "https://www.gse.it/dati-e-scenari/statistiche"

	out := []emissions.Factor{
		// AIB Italian residual mix (Scope 2 market-based).
		{
			Code:         "it_aib_residual_mix",
			ValidFromUTC: mustParse("2022-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2023-01-01T00:00:00Z"),
			Value:        359.0,
			Unit:         "g CO2eq/kWh",
			Source:       aibSource + " 2022 §IT",
			SourceURL:    aibURL,
			Notes:        "Italian residual mix, AIB methodology, Scope 2 market-based.",
		},
		{
			Code:         "it_aib_residual_mix",
			ValidFromUTC: mustParse("2023-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2024-01-01T00:00:00Z"),
			Value:        346.0,
			Unit:         "g CO2eq/kWh",
			Source:       aibSource + " 2023 §IT",
			SourceURL:    aibURL,
		},
		{
			Code:         "it_aib_residual_mix",
			ValidFromUTC: mustParse("2024-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2025-01-01T00:00:00Z"),
			Value:        332.0,
			Unit:         "g CO2eq/kWh",
			Source:       aibSource + " 2024 §IT",
			SourceURL:    aibURL,
		},
		{
			Code:         "it_aib_residual_mix",
			ValidFromUTC: mustParse("2025-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2026-01-01T00:00:00Z"),
			Value:        318.0,
			Unit:         "g CO2eq/kWh",
			Source:       aibSource + " preliminary 2026-Q1 §IT",
			SourceURL:    aibURL,
			Notes:        "Preliminary; final table expected 2026-08.",
		},

		// Italian renewable share (Piano 5.0 input + dashboard KPI).
		{
			Code:         "it_renewable_share",
			ValidFromUTC: mustParse("2022-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2023-01-01T00:00:00Z"),
			Value:        36.0,
			Unit:         "%",
			Source:       fer + " 2023 §1",
			SourceURL:    ferURL,
		},
		{
			Code:         "it_renewable_share",
			ValidFromUTC: mustParse("2023-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2024-01-01T00:00:00Z"),
			Value:        39.5,
			Unit:         "%",
			Source:       fer + " 2024 §1",
			SourceURL:    ferURL,
		},
		{
			Code:         "it_renewable_share",
			ValidFromUTC: mustParse("2024-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2025-01-01T00:00:00Z"),
			Value:        42.1,
			Unit:         "%",
			Source:       fer + " 2025 §1 (preliminary)",
			SourceURL:    ferURL,
		},
		{
			Code:         "it_renewable_share",
			ValidFromUTC: mustParse("2025-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2026-01-01T00:00:00Z"),
			Value:        44.5,
			Unit:         "%",
			Source:       fer + " preliminary 2026-04 §1.2",
			SourceURL:    ferURL,
			Notes:        "Preliminary projection.",
		},

		// Per-source renewable shares (2024 baseline).
		perSource("it_re_share_hydro", 14.8, "FER 2025 §2.1"),
		perSource("it_re_share_solar", 11.2, "FER 2025 §2.2"),
		perSource("it_re_share_wind", 7.4, "FER 2025 §2.3"),
		perSource("it_re_share_geothermal", 1.9, "FER 2025 §2.4"),
		perSource("it_re_share_biomass", 6.8, "FER 2025 §2.5"),
	}
	return out
}

// perSource builds an informational per-source renewable share row valid
// from 2024-01-01 with no upper bound (refreshed on annual Pack review).
func perSource(code string, value float64, sourceTail string) emissions.Factor {
	return emissions.Factor{
		Code:         code,
		ValidFromUTC: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ValidToUTC:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Value:        value,
		Unit:         "%",
		Source:       "GSE Rapporto Statistico " + sourceTail,
		SourceURL:    "https://www.gse.it/dati-e-scenari/statistiche",
	}
}
