// Package esrs_e1 implements the ESRS E1 Climate Change Report Pack —
// the regulatory centrepiece for CSRD wave-2 / wave-3 engagements.
//
// Produces the quantitative E1-5 (Energy consumption and mix) and E1-6
// (Gross Scope 1, 2, 3 and Total GHG emissions) data-points. Narrative
// data-points (E1-1/2/3/4/7/8/9) are null in this Pack — the
// engagement-fork's reporting orchestrator injects client-supplied
// content before signing (Rule 144).
//
// Pure function per Rule 91. Deterministic serialisation per Rule 141.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/reporting/builder.go
//   - Manifest:         packs/report/esrs_e1/manifest.yaml
//   - Charter:          packs/report/esrs_e1/CHARTER.md
package esrs_e1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/greenmetrics/backend/internal/domain/reporting"
)

var Pack reporting.Builder = &builder{}

const (
	PackVersion = "1.0.0"

	ReportType reporting.ReportType = "esrs_e1"

	// Factor-bundle keys consulted by this Builder.
	FactorScope2Location  = "it_grid_mix_location"
	FactorScope2Market    = "it_aib_residual_mix"
	FactorRenewableShare  = "it_renewable_share"
)

type builder struct{}

func (b *builder) Type() reporting.ReportType { return ReportType }
func (b *builder) Version() string            { return PackVersion }

// Body is the canonical typed payload. The shape mirrors the EFRAG ESRS E1
// taxonomy data-point names so Phase H Sprint S15's XBRL tagger can map
// 1:1 without re-shaping.
type Body struct {
	Report               string                `json:"report"`
	Regulator            string                `json:"regulator"`
	Period               reporting.Period      `json:"period"`
	FactorsUsed          map[string]FactorRef  `json:"factors_used"`
	E1_5_EnergyConsMix   E1_5Block             `json:"e1_5_energy_consumption"`
	E1_6_GHGEmissions    E1_6Block             `json:"e1_6_ghg_emissions"`
	NarrativeDataPoints  NarrativeBlock        `json:"narrative_data_points"`
	UnclassifiedRows     int64                 `json:"unclassified_rows"`
}

// FactorRef captures one factor consulted (identical shape across Report Packs).
type FactorRef struct {
	Value   float64 `json:"value"`
	Unit    string  `json:"unit"`
	Version string  `json:"version"`
}

// E1_5Block — Energy consumption and mix.
type E1_5Block struct {
	ElectricityMWh         float64                  `json:"electricity_mwh"`
	NonElectricitySources  []NonElectricitySourceRow `json:"non_electricity_sources"`
	RenewableSharePct      *float64                 `json:"renewable_share_pct,omitempty"`
	NonRenewableSharePct   *float64                 `json:"non_renewable_share_pct,omitempty"`
	Notes                  []string                 `json:"notes,omitempty"`
}

// NonElectricitySourceRow is one Scope 1 fuel quantity.
type NonElectricitySourceRow struct {
	Code     string  `json:"code"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

// E1_6Block — Gross Scope 1, 2, 3 and Total GHG emissions.
type E1_6Block struct {
	Scope1 Scope1Sub `json:"scope_1"`
	Scope2 Scope2Sub `json:"scope_2"`
	Scope3 Scope3Sub `json:"scope_3"`
	Totals TotalsSub `json:"totals"`
}

type Scope1Sub struct {
	KgCO2eqTotal float64           `json:"kg_co2eq_total"`
	PerSource    []Scope1SourceRow `json:"per_source"`
}

type Scope1SourceRow struct {
	Code     string  `json:"code"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	KgCO2eq  float64 `json:"kg_co2eq"`
}

type Scope2Sub struct {
	KWhTotal             float64  `json:"kwh_total"`
	KgCO2eqLocationBased *float64 `json:"kg_co2eq_location_based,omitempty"`
	KgCO2eqMarketBased   *float64 `json:"kg_co2eq_market_based,omitempty"`
}

type Scope3Sub struct {
	KgCO2eqTotal float64 `json:"kg_co2eq_total"`
	Note         string  `json:"note"`
}

type TotalsSub struct {
	KgCO2eqLocationBased *float64 `json:"kg_co2eq_location_based,omitempty"`
	KgCO2eqMarketBased   *float64 `json:"kg_co2eq_market_based,omitempty"`
}

// NarrativeBlock — placeholder for client-supplied narrative content
// that the engagement-fork's reporting orchestrator injects.
type NarrativeBlock struct {
	E11  *string `json:"e1_1"`
	E12  *string `json:"e1_2"`
	E13  *string `json:"e1_3"`
	E14  *string `json:"e1_4"`
	E17  *string `json:"e1_7"`
	E18  *string `json:"e1_8"`
	E19  *string `json:"e1_9"`
	Note string  `json:"note"`
}

// classification mirrors the co2_footprint Pack's reading-classifier.
// (Pure-function Builders do not import each other; this is intentional
// duplication.)
type classification struct {
	scope      int    // 1 or 2; 0 = unclassified
	factorCode string // factor-bundle key for the per-source factor
}

func classifyByUnit(unit string) classification {
	switch unit {
	case "Wh", "kWh":
		return classification{scope: 2}
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
	scope2 := Scope2Sub{KWhTotal: scope2KWh}
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

	// ── Scope 1 — per source ─────────────────────────────────────────
	scope1Keys := make([]string, 0, len(scope1))
	for k := range scope1 {
		scope1Keys = append(scope1Keys, string(k))
	}
	sort.Strings(scope1Keys)

	scope1Sub := Scope1Sub{}
	nonElec := make([]NonElectricitySourceRow, 0, len(scope1Keys))
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
			scope1Sub.KgCO2eqTotal += row.KgCO2eq
		}
		scope1Sub.PerSource = append(scope1Sub.PerSource, row)
		nonElec = append(nonElec, NonElectricitySourceRow{
			Code:     k,
			Quantity: acc.quantity,
			Unit:     acc.unit,
		})
	}

	// ── Scope 3 — placeholder ────────────────────────────────────────
	scope3 := Scope3Sub{
		Note: "Scope 3 inputs flow through a separate ingestion path; zero by default until populated (Sprint S15).",
	}

	// ── Totals ───────────────────────────────────────────────────────
	totals := TotalsSub{}
	if scope2.KgCO2eqLocationBased != nil {
		v := *scope2.KgCO2eqLocationBased + scope1Sub.KgCO2eqTotal + scope3.KgCO2eqTotal
		totals.KgCO2eqLocationBased = &v
	}
	if scope2.KgCO2eqMarketBased != nil {
		v := *scope2.KgCO2eqMarketBased + scope1Sub.KgCO2eqTotal + scope3.KgCO2eqTotal
		totals.KgCO2eqMarketBased = &v
	}

	// ── E1-5 — Energy consumption + mix ──────────────────────────────
	e15 := E1_5Block{
		ElectricityMWh:        scope2KWh / 1000.0,
		NonElectricitySources: nonElec,
	}
	notes := []string{
		"E1-5 lower-heating-value conversion to MWh for non-electricity sources is deferred to Sprint S6 ISPRA-LHV update; non-electricity quantities reported in their primary unit.",
	}
	if v, ver, ok := factors.Get(FactorRenewableShare); ok {
		factorsUsed[FactorRenewableShare] = FactorRef{Value: v, Unit: "%", Version: ver}
		share := v
		e15.RenewableSharePct = &share
		nonRen := 100.0 - v
		e15.NonRenewableSharePct = &nonRen
	} else {
		notes = append(notes, fmt.Sprintf("factor %q missing; renewable share omitted", FactorRenewableShare))
	}
	if unclassifiedCount > 0 {
		notes = append(notes, fmt.Sprintf("%d reading(s) had unclassified Unit values; not in scope totals", unclassifiedCount))
	}
	e15.Notes = notes

	// ── E1-6 — GHG emissions ─────────────────────────────────────────
	e16 := E1_6Block{
		Scope1: scope1Sub,
		Scope2: scope2,
		Scope3: scope3,
		Totals: totals,
	}

	// ── Narrative — placeholder ──────────────────────────────────────
	narrative := NarrativeBlock{
		Note: "Narrative E1-1/2/3/4/7/8/9 are bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4 + Sprint S15).",
	}

	body := Body{
		Report:              "esrs_e1",
		Regulator:           "EFRAG / CSRD wave 2",
		Period:              period,
		FactorsUsed:         factorsUsed,
		E1_5_EnergyConsMix:  e15,
		E1_6_GHGEmissions:   e16,
		NarrativeDataPoints: narrative,
		UnclassifiedRows:    unclassifiedCount,
	}

	encoded, err := encode(body)
	if err != nil {
		return reporting.Report{}, fmt.Errorf("encode: %w", err)
	}

	reportNotes := []string{}
	if scope2.KgCO2eqLocationBased == nil && scope2WhSum > 0 {
		reportNotes = append(reportNotes,
			fmt.Sprintf("factor %q missing; Scope 2 location-based emissions omitted", FactorScope2Location))
	}
	if scope2.KgCO2eqMarketBased == nil && scope2WhSum > 0 {
		reportNotes = append(reportNotes,
			fmt.Sprintf("factor %q missing; Scope 2 market-based emissions omitted", FactorScope2Market))
	}

	return reporting.Report{
		Type:    ReportType,
		Period:  period,
		Body:    body,
		Encoded: encoded,
		Notes:   reportNotes,
	}, nil
}

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
