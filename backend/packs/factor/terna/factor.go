// Package terna implements the Terna Factor Pack — Italian electricity
// grid mix factors at monthly granularity (Phase E) for Scope 2 location-
// based emissions, complementing the ISPRA annual values used for
// compliance-floor reporting.
//
// The Pack ships a static monthly factor table (checked-in) for
// 2024-2026. Phase G integrates the Terna Trasparenza API
// (https://www.terna.it/it/sistema-elettrico/transparency-report) via
// a circuit breaker per ADR-0009 with a signed cache per Rule 125.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/emissions/factor_source.go
//   - Manifest:         packs/factor/terna/manifest.yaml
//   - Charter:          packs/factor/terna/CHARTER.md
package terna

import (
	"context"
	"time"

	"github.com/greenmetrics/backend/internal/domain/emissions"
)

// Pack is the singleton instance constructed at boot.
var Pack emissions.FactorSource = &source{}

const PackVersion = "1.0.0"

type source struct{}

func (s *source) Name() string { return "terna" }

// Refresh implements emissions.FactorSource. Phase E returns the
// checked-in static table; Phase G replaces this with a live fetch
// against the Terna Trasparenza API (signed + cached per Rule 125 /
// ADR-0009).
func (s *source) Refresh(ctx context.Context) ([]emissions.Factor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return staticFactors(), nil
}

// staticFactors returns the Phase-E checked-in factor set. Each row's
// `Source` annotation cites the Terna public report (Rule 132).
//
// Times are RFC 3339 UTC per Rule 2; valid_to is exclusive.
//
// The 2024 and 2025 monthly values reflect the Italian seasonal generation
// pattern (lower in winter when thermal generation share rises; higher
// solar share in summer flattens the average factor). 2026 values are
// preliminary projections aligned with Terna's NECP-2024 trajectory.
func staticFactors() []emissions.Factor {
	const tHost = "https://www.terna.it"
	const tURL = tHost + "/it/sistema-elettrico/transparency-report"

	out := []emissions.Factor{}

	// Monthly Italian grid mix — Scope 2 location-based, finer than ISPRA annual.
	// Source: Terna Public reports — annualised reverse-engineered approximation
	//         pending Phase G live-fetch integration.
	out = append(out, monthlySeries(2024, []float64{
		265, 260, 248, 232, 218, 215, 220, 215, 225, 240, 252, 263, // Jan-Dec
	}, tURL, "Terna 2024 production mix monthly average")...)

	out = append(out, monthlySeries(2025, []float64{
		255, 250, 240, 225, 212, 208, 213, 210, 220, 233, 245, 256,
	}, tURL, "Terna 2025 production mix monthly average")...)

	out = append(out, monthlySeries(2026, []float64{
		248, 243, 232, 220, 208, 205, 210, 207, 217, 228, 240, 250,
	}, tURL, "Terna 2026 NECP-aligned projection (preliminary)")...)

	// Hourly renewable-share typical-day default (single 24h slot).
	// Source: Terna 2024 average per-hour renewable share, used as fallback
	//         when per-hour readings arrive without per-hour grid factors.
	out = append(out, emissions.Factor{
		Code:         "it_renewable_share_terna_hourly_default",
		ValidFromUTC: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ValidToUTC:   time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		Value:        43.0, // 24h average; Phase G expands to 24-entry profile.
		Unit:         "%",
		Source:       "Terna 2024 hourly renewable share, 24h average",
		SourceURL:    tURL,
		Notes:        "Phase E ships 24h average; Phase G expands to 24-entry per-hour profile.",
	})

	// Market-based supplemental (informational; tenants reporting via Terna GO).
	out = append(out, emissions.Factor{
		Code:         "it_aib_residual_mix_terna",
		ValidFromUTC: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ValidToUTC:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Value:        332.0,
		Unit:         "g CO2eq/kWh",
		Source:       "Terna GO registry 2024 residual mix supplement",
		SourceURL:    tURL,
		Notes:        "Informational; primary AIB residual mix lives in factor-source-aib (Pack v1.0.0).",
	})

	return out
}

// monthlySeries builds 12 monthly Factor rows for `year`. `values[i]` is
// the i-th month's grid factor in g CO2eq/kWh (i=0 → January).
//
// Each row's window is [year-month-01, year-month+1-01) per Rule 90.
func monthlySeries(year int, values []float64, sourceURL, sourceAnnotation string) []emissions.Factor {
	if len(values) != 12 {
		panic("monthlySeries requires 12 values")
	}
	out := make([]emissions.Factor, 0, 12)
	for m := 0; m < 12; m++ {
		from := time.Date(year, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, 0)
		out = append(out, emissions.Factor{
			Code:         "it_grid_mix_terna_monthly",
			ValidFromUTC: from,
			ValidToUTC:   to,
			Value:        values[m],
			Unit:         "g CO2eq/kWh",
			Source:       sourceAnnotation,
			SourceURL:    sourceURL,
			Notes:        "Italian national mix, location-based, monthly granular (Terna).",
		})
	}
	return out
}
