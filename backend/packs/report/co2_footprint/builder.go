// Package co2_footprint implements the CO₂ Footprint Report Pack —
// Scope 1 (combustion) + Scope 2 (location-based + market-based) summary
// over a half-open period.
//
// Pure function per Rule 91. Deterministic byte-stream per Rule 141.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/reporting/builder.go
//   - Manifest:         packs/report/co2_footprint/manifest.yaml
//   - Charter:          packs/report/co2_footprint/CHARTER.md
package co2_footprint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/greenmetrics/backend/internal/domain/reporting"
)

// Pack is the singleton instance constructed at boot.
var Pack reporting.Builder = &builder{}

// PackVersion is the Pack's SemVer (matches manifest.yaml version).
const PackVersion = "1.0.0"

// ReportType is the report-type identifier this Builder registers.
const ReportType reporting.ReportType = "co2_footprint"

// Factor-bundle keys this Builder consults.
const (
	FactorScope2Location = "it_grid_mix_location"
	FactorScope2Market   = "it_aib_residual_mix"
)

// builder is the concrete Builder implementation.
type builder struct{}

func (b *builder) Type() reporting.ReportType { return ReportType }
func (b *builder) Version() string            { return PackVersion }

// classification maps a normalised AggregatedRow.Unit to a scope + factor
// code. Entries kept in alphabetical order on `unit` for deterministic
// serialisation of the factors_used map.
type classification struct {
	scope      int    // 1 or 2; 0 = unclassified
	factorCode string // factor-bundle key for the per-source factor
}

// classifyByUnit returns the classification for a known unit, or
// unclassified (scope=0) if the unit is not known.
func classifyByUnit(unit string) classification {
	switch unit {
	case "Wh", "kWh":
		return classification{scope: 2, factorCode: ""} // Scope 2 uses both factors
	case "Sm3":
		return classification{scope: 1, factorCode: "natural_gas_combustion"}
	case "l_diesel":
		return classification{scope: 1, factorCode: "diesel_combustion"}
	case "l_petrol":
		return classification{scope: 1, factorCode: "petrol_road_vehicle"}
	case "l_diesel_vehicle":
		return classification{scope: 1, factorCode: "diesel_road_vehicle"}
	case "kg_lpg":
		return classification{scope: 1, factorCode: "lpg_combustion"}
	case "kg_coal":
		return classification{scope: 1, factorCode: "coal_combustion"}
	case "kg_heavy_fuel":
		return classification{scope: 1, factorCode: "heavy_fuel_oil_combustion"}
	}
	return classification{scope: 0}
}

// Body is the canonical typed payload returned in Report.Body.
type Body struct {
	Period           reporting.Period      `json:"period"`
	FactorsUsed      map[string]FactorRef  `json:"factors_used"`
	Scope1           Scope1Summary         `json:"scope_1"`
	Scope2           Scope2Summary         `json:"scope_2"`
	Scope3           Scope3Summary         `json:"scope_3"`
	Totals           TotalsSummary         `json:"totals"`
	UnclassifiedRows int64                 `json:"unclassified_rows"`
}

// FactorRef captures one factor consulted (identical shape across Report Packs).
type FactorRef struct {
	Value   float64 `json:"value"`
	Unit    string  `json:"unit"`
	Version string  `json:"version"`
}

// Scope1Summary aggregates combustion sources.
type Scope1Summary struct {
	KgCO2eqTotal float64           `json:"kg_co2eq_total"`
	PerSource    []Scope1SourceRow `json:"per_source"`
}

// Scope1SourceRow is one combustion-source row.
type Scope1SourceRow struct {
	Code     string  `json:"code"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	KgCO2eq  float64 `json:"kg_co2eq"`
}

// Scope2Summary aggregates electricity consumption with dual-method emissions.
type Scope2Summary struct {
	KWhTotal              float64  `json:"kwh_total"`
	KgCO2eqLocationBased  *float64 `json:"kg_co2eq_location_based,omitempty"`
	KgCO2eqMarketBased    *float64 `json:"kg_co2eq_market_based,omitempty"`
}

// Scope3Summary is a placeholder (Sprint S15 ingestion path adds inputs).
type Scope3Summary struct {
	KgCO2eqTotal float64 `json:"kg_co2eq_total"`
	Note         string  `json:"note"`
}

// TotalsSummary rolls Scope 1 + Scope 2 (per method) + Scope 3 into a single
// pair of grand totals.
type TotalsSummary struct {
	KgCO2eqLocationBased *float64 `json:"kg_co2eq_location_based,omitempty"`
	KgCO2eqMarketBased   *float64 `json:"kg_co2eq_market_based,omitempty"`
}

// Build implements reporting.Builder. Pure function.
func (b *builder) Build(
	ctx context.Context,
	period reporting.Period,
	factors reporting.FactorBundle,
	readings reporting.AggregatedReadings,
) (reporting.Report, error) {
	if err := ctx.Err(); err != nil {
		return reporting.Report{}, err
	}

	type fuelKey string
	type fuelAcc struct {
		quantity float64
		unit     string
	}
	scope1 := map[fuelKey]*fuelAcc{}
	var scope2WhSum int64
	var unclassifiedCount int64

	iter := readings.Iter()
	for iter.Next() {
		row := iter.Row()
		c := classifyByUnit(row.Unit)
		switch c.scope {
		case 2:
			// Convert kWh → Wh equivalent for accumulation.
			if row.Unit == "kWh" {
				scope2WhSum += row.Sum * 1000
			} else {
				scope2WhSum += row.Sum
			}
		case 1:
			k := fuelKey(c.factorCode)
			acc := scope1[k]
			if acc == nil {
				acc = &fuelAcc{unit: row.Unit}
				scope1[k] = acc
			}
			acc.quantity += float64(row.Sum)
		default:
			unclassifiedCount += row.Count
		}
	}
	if err := iter.Err(); err != nil {
		return reporting.Report{}, fmt.Errorf("readings iteration: %w", err)
	}

	factorsUsed := map[string]FactorRef{}

	// ── Scope 2 — dual method ────────────────────────────────────────
	scope2KWh := float64(scope2WhSum) / 1000.0
	scope2 := Scope2Summary{KWhTotal: scope2KWh}

	if v, ver, ok := factors.Get(FactorScope2Location); ok {
		factorsUsed[FactorScope2Location] = FactorRef{Value: v, Unit: "g CO2eq/kWh", Version: ver}
		kg := scope2KWh * v / 1000.0
		scope2.KgCO2eqLocationBased = &kg
	}
	if v, ver, ok := factors.Get(FactorScope2Market); ok {
		factorsUsed[FactorScope2Market] = FactorRef{Value: v, Unit: "g CO2eq/kWh", Version: ver}
		kg := scope2KWh * v / 1000.0
		scope2.KgCO2eqMarketBased = &kg
	}

	// ── Scope 1 — per-source ─────────────────────────────────────────
	// Sort fuel keys deterministically (Rule 141).
	scope1Keys := make([]string, 0, len(scope1))
	for k := range scope1 {
		scope1Keys = append(scope1Keys, string(k))
	}
	sort.Strings(scope1Keys)

	scope1Sum := Scope1Summary{}
	for _, k := range scope1Keys {
		acc := scope1[fuelKey(k)]
		row := Scope1SourceRow{
			Code:     k,
			Quantity: acc.quantity,
			Unit:     acc.unit,
		}
		if v, ver, ok := factors.Get(k); ok {
			factorsUsed[k] = FactorRef{Value: v, Unit: factorUnitForCode(k), Version: ver}
			row.KgCO2eq = acc.quantity * v
			scope1Sum.KgCO2eqTotal += row.KgCO2eq
		}
		scope1Sum.PerSource = append(scope1Sum.PerSource, row)
	}

	// ── Scope 3 — placeholder ────────────────────────────────────────
	scope3 := Scope3Summary{
		KgCO2eqTotal: 0,
		Note:         "Scope 3 inputs flow through a separate ingestion path; zero by default until populated.",
	}

	// ── Totals ───────────────────────────────────────────────────────
	totals := TotalsSummary{}
	if scope2.KgCO2eqLocationBased != nil {
		v := *scope2.KgCO2eqLocationBased + scope1Sum.KgCO2eqTotal + scope3.KgCO2eqTotal
		totals.KgCO2eqLocationBased = &v
	}
	if scope2.KgCO2eqMarketBased != nil {
		v := *scope2.KgCO2eqMarketBased + scope1Sum.KgCO2eqTotal + scope3.KgCO2eqTotal
		totals.KgCO2eqMarketBased = &v
	}

	body := Body{
		Period:           period,
		FactorsUsed:      factorsUsed,
		Scope1:           scope1Sum,
		Scope2:           scope2,
		Scope3:           scope3,
		Totals:           totals,
		UnclassifiedRows: unclassifiedCount,
	}

	encoded, err := encode(body)
	if err != nil {
		return reporting.Report{}, fmt.Errorf("encode: %w", err)
	}

	notes := []string{}
	if unclassifiedCount > 0 {
		notes = append(notes, fmt.Sprintf(
			"%d reading(s) had unclassified Unit values; not counted in scope totals", unclassifiedCount))
	}
	if scope2.KgCO2eqLocationBased == nil && scope2WhSum > 0 {
		notes = append(notes, fmt.Sprintf(
			"factor %q missing; Scope 2 location-based emissions omitted", FactorScope2Location))
	}
	if scope2.KgCO2eqMarketBased == nil && scope2WhSum > 0 {
		notes = append(notes, fmt.Sprintf(
			"factor %q missing; Scope 2 market-based emissions omitted", FactorScope2Market))
	}

	return reporting.Report{
		Type:    ReportType,
		Period:  period,
		Body:    body,
		Encoded: encoded,
		Notes:   notes,
	}, nil
}

// factorUnitForCode returns the canonical unit string used in factor
// bundles (kg CO2eq / unit-of-quantity) for a Scope 1 factor code. Used
// to populate FactorRef.Unit symmetrically across Scope 1 sources.
func factorUnitForCode(code string) string {
	switch code {
	case "natural_gas_combustion":
		return "kg CO2eq/Sm3"
	case "diesel_combustion", "petrol_road_vehicle", "diesel_road_vehicle":
		return "kg CO2eq/l"
	case "lpg_combustion", "coal_combustion", "heavy_fuel_oil_combustion":
		return "kg CO2eq/kg"
	}
	return ""
}

// encode serialises body deterministically (Rule 141).
func encode(body Body) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
