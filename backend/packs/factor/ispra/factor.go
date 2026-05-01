// Package ispra implements the ISPRA Factor Pack — Italian national grid
// emission factors and primary-fuel combustion factors per Rapporto 404
// and D.M. 11 gennaio 2017 Allegati 1+2.
//
// The Pack ships a static factor table (checked-in) for Phase E. Phase G
// adds the live-fetch path against ispra.gov.it; the live path replaces
// staticFactors() while keeping Refresh's external behaviour unchanged.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/emissions/factor_source.go
//   - Manifest:         packs/factor/ispra/manifest.yaml
//   - Charter:          packs/factor/ispra/CHARTER.md
package ispra

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
func (s *source) Name() string { return "ispra" }

// Refresh implements emissions.FactorSource. Phase E returns the
// checked-in static table; Phase G replaces this with a live fetch
// against ispra.gov.it (signed + cached per Rule 125 / ADR-0009).
func (s *source) Refresh(ctx context.Context) ([]emissions.Factor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return staticFactors(), nil
}

// staticFactors returns the Phase-E checked-in factor set. Each row's
// `Source` annotation cites the primary regulator publication (Rule 132).
//
// Times are RFC 3339 UTC per Rule 2; valid_to is exclusive (the next
// year's valid_from), or the zero-time for "still active" rows.
func staticFactors() []emissions.Factor {
	mustParse := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err) // checked-in literals — boot-time failure is fine
		}
		return t
	}

	const isprahost = "https://www.isprambiente.gov.it"

	out := []emissions.Factor{
		// Scope 2 — Italian national mix (location-based) — historical + projection.
		{
			Code:         "it_grid_mix_location",
			ValidFromUTC: mustParse("2022-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2023-01-01T00:00:00Z"),
			Value:        268.0,
			Unit:         "g CO2eq/kWh",
			Source:       "ISPRA Rapporto 404/2024 Tabella 1",
			SourceURL:    isprahost + "/it/pubblicazioni/rapporti/fattori-di-emissione",
			Notes:        "Italian national mix, location-based, Scope 2.",
		},
		{
			Code:         "it_grid_mix_location",
			ValidFromUTC: mustParse("2023-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2024-01-01T00:00:00Z"),
			Value:        245.0,
			Unit:         "g CO2eq/kWh",
			Source:       "ISPRA Rapporto 404/2025 Tabella 1",
			SourceURL:    isprahost + "/it/pubblicazioni/rapporti/fattori-di-emissione",
		},
		{
			Code:         "it_grid_mix_location",
			ValidFromUTC: mustParse("2024-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2025-01-01T00:00:00Z"),
			Value:        233.0,
			Unit:         "g CO2eq/kWh",
			Source:       "ISPRA Rapporto 404/2025 Tabella 1",
			SourceURL:    isprahost + "/it/pubblicazioni/rapporti/fattori-di-emissione",
		},
		{
			Code:         "it_grid_mix_location",
			ValidFromUTC: mustParse("2025-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2026-01-01T00:00:00Z"),
			Value:        225.0,
			Unit:         "g CO2eq/kWh",
			Source:       "ISPRA preliminary 2026-04 sezione 2",
			SourceURL:    isprahost + "/it/pubblicazioni/rapporti/fattori-di-emissione",
			Notes:        "Preliminary; final table expected Q3 2026.",
		},
		{
			Code:         "it_grid_mix_location",
			ValidFromUTC: mustParse("2026-01-01T00:00:00Z"),
			ValidToUTC:   mustParse("2027-01-01T00:00:00Z"),
			Value:        218.0,
			Unit:         "g CO2eq/kWh",
			Source:       "ISPRA preliminary 2026-04 sezione 2",
			SourceURL:    isprahost + "/it/pubblicazioni/rapporti/fattori-di-emissione",
			Notes:        "Preliminary projection.",
		},

		// Scope 1 — Stationary combustion (D.M. 11 gennaio 2017 Allegato 1).
		stationary("natural_gas_combustion", 1.967, "kg CO2eq/Sm3"),
		stationary("lpg_combustion", 2.965, "kg CO2eq/kg"),
		stationary("diesel_combustion", 2.642, "kg CO2eq/l"),
		stationary("heavy_fuel_oil_combustion", 3.155, "kg CO2eq/kg"),
		stationary("coal_combustion", 2.394, "kg CO2eq/kg"),

		// Scope 1 — Mobile combustion (D.M. 11 gennaio 2017 Allegato 2).
		mobile("petrol_road_vehicle", 2.318),
		mobile("diesel_road_vehicle", 2.642),
	}
	return out
}

// stationary builds a Scope 1 stationary-combustion Factor row valid
// 2024-01-01 through 2027-12-31 (the D.M. 11/01/2017 publication's
// effective window, until ISPRA / MIMIT supersedes).
func stationary(code string, value float64, unit string) emissions.Factor {
	return emissions.Factor{
		Code:         code,
		ValidFromUTC: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ValidToUTC:   time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC),
		Value:        value,
		Unit:         unit,
		Source:       "D.M. 11/01/2017 Allegato 1",
		SourceURL:    "https://www.normattiva.it",
	}
}

// mobile builds a Scope 1 mobile-combustion Factor row valid 2024-01-01
// through 2027-12-31 (D.M. 11/01/2017 Allegato 2).
func mobile(code string, value float64) emissions.Factor {
	return emissions.Factor{
		Code:         code,
		ValidFromUTC: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ValidToUTC:   time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC),
		Value:        value,
		Unit:         "kg CO2eq/l",
		Source:       "D.M. 11/01/2017 Allegato 2",
		SourceURL:    "https://www.normattiva.it",
	}
}
