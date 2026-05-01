// Package audit_dlgs102 implements the D.Lgs. 102/2014 art. 8 mandatory
// energy-audit Report Pack ("diagnosi energetica obbligatoria") — the
// highest-volume Italian regulatory dossier, applying to ~10 000 grandi
// imprese + imprese energivore on a 4-yearly cadence.
//
// Produces the quantitative half of the audit dossier: per-vector +
// per-site energy baseline (kWh + tep), obligation block (audit type,
// 4-yearly cadence, 5-December deadline, exemption basis), and EnPI
// candidates. Narrative content (UNI CEI EN 16247-1..4 compliance
// statement, site-visit summaries, scope description, EGE / ESCo
// certifier identity, improvement-measure list with payback analysis,
// monitoring-plan KPIs) is null in this Pack — the engagement-fork's
// reporting orchestrator injects client-supplied content before signing
// (Rule 144).
//
// Pure function per Rule 91. Deterministic serialisation per Rule 141.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/reporting/builder.go
//   - Manifest:         packs/report/audit_dlgs102/manifest.yaml
//   - Charter:          packs/report/audit_dlgs102/CHARTER.md
package audit_dlgs102

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/greenmetrics/backend/internal/domain/reporting"
)

var Pack reporting.Builder = &builder{}

const (
	PackVersion = "1.0.0"

	ReportType reporting.ReportType = "audit_dlgs102"

	// ── Vector identifiers (used in by_vector ordering + factor keys). ─
	VectorElectricity  = "electricity"
	VectorNaturalGas   = "natural_gas"
	VectorDiesel       = "diesel"
	VectorPetrol       = "petrol"
	VectorLPG          = "lpg"
	VectorCoal         = "coal"
	VectorHeavyFuelOil = "heavy_fuel_oil"

	// ── FactorBundle keys consulted by this Builder. ───────────────────
	FactorTepFactorPrefix  = "audit102_tep_factor_"
	FactorDensityPrefix    = "audit102_density_"
	FactorLHVPrefix        = "audit102_lhv_"
	FactorObligationType   = "audit102_obligation_type"
	FactorBelow50Exempt    = "audit102_below_50_exempt"
	FactorExemptionISO     = "audit102_exemption_iso"
	FactorNextDeadlineUnix = "audit102_next_deadline_unix"
	FactorTotalFloorM2     = "audit102_total_floor_m2"

	// Periodicity per art. 8 c. 1.
	AuditPeriodicityYears = 4

	// 50 tep/year exemption threshold per D.Lgs. 73/2020 (introducing
	// art. 8 c. 3-bis).
	Below50TepThreshold = 50.0

	// Obligation-type selector values.
	ObligationGrandeImpresa = 0
	ObligationEnergivora    = 1
	ObligationVoluntary     = 2

	// Exemption-basis selector values.
	ExemptionNone     = 0
	ExemptionISO50001 = 1
	ExemptionEMAS     = 2
	ExemptionISO14001 = 3
)

// ── Statutory tep conversion factors per D.M. 20/07/2004 + ARERA Delibera ──
//
//	03/08 EEN. Sources accessed 2026-04-30. Engagement Phase 0 Discovery
//	verifies the values match the audit102.enea.it portal version current
//	at submission; overridable via FactorBundle audit102_tep_factor_<vector>.
//
// Source: ARERA Delibera 03/08 EEN ("Aggiornamento del fattore di conversione
//
//	dei kWh in tonnellate equivalenti di petrolio").
const TepFactorElectricity = 0.000187 // tep/kWh

// Source: D.M. 20/07/2004 (gas naturale) — 0.836 tep / 1000 Sm3.
const TepFactorNaturalGas = 0.000836 // tep/Sm3

// Source: D.M. 20/07/2004 + ARERA TEE table.
const (
	TepFactorDiesel       = 0.001080 // tep/kg
	TepFactorPetrol       = 0.001050 // tep/kg
	TepFactorLPG          = 0.001099 // tep/kg
	TepFactorCoal         = 0.000700 // tep/kg
	TepFactorHeavyFuelOil = 0.000980 // tep/kg
)

// ── Default fuel densities (kg per litre). ────────────────────────────
// Source: D.M. 20/07/2004 default values; engagement may override via
//
//	FactorBundle audit102_density_<vector>.
const (
	DefaultDensityDiesel = 0.835 // kg/l
	DefaultDensityPetrol = 0.745 // kg/l
)

type builder struct{}

func (b *builder) Type() reporting.ReportType { return ReportType }
func (b *builder) Version() string            { return PackVersion }

// Body is the canonical typed payload. Field order is the deterministic
// JSON-encoding order (Rule 141).
type Body struct {
	Report              string               `json:"report"`
	Regulator           string               `json:"regulator"`
	Period              reporting.Period     `json:"period"`
	Obligation          ObligationBlock      `json:"obligation"`
	FactorsUsed         map[string]FactorRef `json:"factors_used"`
	EnergyBaseline      EnergyBaselineBlock  `json:"energy_baseline"`
	SiteBreakdown       []SiteRow            `json:"site_breakdown"`
	EnPI                []EnPIRow            `json:"enpi"`
	ImprovementMeasures []ImprovementMeasure `json:"improvement_measures"`
	MonitoringPlan      MonitoringPlanBlock  `json:"monitoring_plan"`
	NarrativeDataPoints NarrativeBlock       `json:"narrative_data_points"`
	EGECertRequired     bool                 `json:"ege_certification_required"`
	UnclassifiedRows    int64                `json:"unclassified_rows"`
	Notes               []string             `json:"notes,omitempty"`
}

type FactorRef struct {
	Value   float64 `json:"value"`
	Unit    string  `json:"unit"`
	Version string  `json:"version"`
}

type ObligationBlock struct {
	Type                  string  `json:"type"`
	AuditPeriodicityYears int     `json:"audit_periodicity_years"`
	NextDeadlineISO       *string `json:"next_deadline_iso"`
	ExemptionBasis        *string `json:"exemption_basis"`
	Below50TepExemption   bool    `json:"below_50_tep_exemption"`
}

type EnergyBaselineBlock struct {
	ByVector            []VectorRow `json:"by_vector"`
	TotalKWh            float64     `json:"total_kwh"`
	TotalTep            float64     `json:"total_tep"`
	Below50TepThreshold bool        `json:"below_50_tep_threshold"`
}

type VectorRow struct {
	Vector       string  `json:"vector"`
	PrimaryUnit  string  `json:"primary_unit"`
	PrimaryTotal float64 `json:"primary_total"`
	KWhTotal     float64 `json:"kwh_total"`
	TepTotal     float64 `json:"tep_total"`
	TepFactor    float64 `json:"tep_factor"`
}

type SiteRow struct {
	SiteID   string      `json:"site_id"`
	ByVector []VectorRow `json:"by_vector"`
	KWhTotal float64     `json:"kwh_total"`
	TepTotal float64     `json:"tep_total"`
}

type EnPIRow struct {
	Name           string  `json:"name"`
	Indicator      string  `json:"indicator"`
	Value          float64 `json:"value"`
	BaselinePeriod string  `json:"baseline_period"`
}

type ImprovementMeasure struct {
	ID                     string   `json:"id"`
	Description            *string  `json:"description"`
	EstimatedCostEUR       *float64 `json:"estimated_cost_eur"`
	EstimatedSavingKWhYear *float64 `json:"estimated_saving_kwh_year"`
	EstimatedSavingTepYear *float64 `json:"estimated_saving_tep_year"`
	PaybackYears           *float64 `json:"payback_years"`
	Priority               *string  `json:"priority"`
}

type MonitoringPlanBlock struct {
	KPIs                  []string `json:"kpis"`
	ReviewFrequencyMonths int      `json:"review_frequency_months"`
}

type NarrativeBlock struct {
	ScopeDescription                 *string `json:"scope_description"`
	SiteVisitSummaries               *string `json:"site_visit_summaries"`
	UNICEIEN16247ComplianceStatement *string `json:"uni_cei_en_16247_compliance_statement"`
	EGECertifierID                   *string `json:"ege_certifier_id"`
	Note                             string  `json:"note"`
}

// vectorClass captures classification of a single reading row.
type vectorClass struct {
	vector      string
	primaryUnit string
	// kWhPerPrimary: multiplier from primary unit to kWh-equivalent. Zero if
	// the LHV-equivalent kWh is not available without an explicit LHV factor.
	kWhPerPrimary float64
}

// classifyByUnit maps an AggregatedRow.Unit to its vector classification.
// kWhPerPrimary is non-zero only for electricity (where the primary unit IS
// kWh / Wh). For non-electricity vectors the kWh-equivalent requires an
// LHV factor supplied via FactorBundle audit102_lhv_<vector> (TODO Phase F).
func classifyByUnit(unit string) (vectorClass, bool) {
	switch unit {
	case "Wh":
		return vectorClass{vector: VectorElectricity, primaryUnit: "kWh", kWhPerPrimary: 1}, true
	case "kWh":
		return vectorClass{vector: VectorElectricity, primaryUnit: "kWh", kWhPerPrimary: 1}, true
	case "Sm3":
		return vectorClass{vector: VectorNaturalGas, primaryUnit: "Sm3"}, true
	case "l_diesel", "l_diesel_vehicle":
		// Litres → kg via density.
		return vectorClass{vector: VectorDiesel, primaryUnit: "kg"}, true
	case "l_petrol":
		return vectorClass{vector: VectorPetrol, primaryUnit: "kg"}, true
	case "kg_lpg":
		return vectorClass{vector: VectorLPG, primaryUnit: "kg"}, true
	case "kg_coal":
		return vectorClass{vector: VectorCoal, primaryUnit: "kg"}, true
	case "kg_heavy_fuel":
		return vectorClass{vector: VectorHeavyFuelOil, primaryUnit: "kg"}, true
	}
	return vectorClass{}, false
}

// statutoryTepFactor returns the default tep-per-primary-unit factor for
// `vector`, before any FactorBundle override.
func statutoryTepFactor(vector string) (value float64, factorUnit string) {
	switch vector {
	case VectorElectricity:
		return TepFactorElectricity, "tep/kWh"
	case VectorNaturalGas:
		return TepFactorNaturalGas, "tep/Sm3"
	case VectorDiesel:
		return TepFactorDiesel, "tep/kg"
	case VectorPetrol:
		return TepFactorPetrol, "tep/kg"
	case VectorLPG:
		return TepFactorLPG, "tep/kg"
	case VectorCoal:
		return TepFactorCoal, "tep/kg"
	case VectorHeavyFuelOil:
		return TepFactorHeavyFuelOil, "tep/kg"
	}
	return 0, ""
}

// defaultDensity returns the default fuel density (kg/l) for liquid fuels
// reported in litres. Returns 0 for vectors not in litres.
func defaultDensity(vector string) float64 {
	switch vector {
	case VectorDiesel:
		return DefaultDensityDiesel
	case VectorPetrol:
		return DefaultDensityPetrol
	}
	return 0
}

// obligationName decodes the audit102_obligation_type selector.
func obligationName(selector float64) string {
	switch int(selector) {
	case ObligationEnergivora:
		return "energivora"
	case ObligationVoluntary:
		return "voluntary"
	default:
		return "grande_impresa"
	}
}

// exemptionName decodes the audit102_exemption_iso selector.
func exemptionName(selector float64) *string {
	switch int(selector) {
	case ExemptionISO50001:
		s := "iso_50001"
		return &s
	case ExemptionEMAS:
		s := "emas"
		return &s
	case ExemptionISO14001:
		s := "iso_14001"
		return &s
	}
	return nil
}

// vectorOrdering is the canonical ordering of vectors for deterministic
// JSON emission (Rule 141).
var vectorOrdering = []string{
	VectorElectricity,
	VectorNaturalGas,
	VectorDiesel,
	VectorPetrol,
	VectorLPG,
	VectorCoal,
	VectorHeavyFuelOil,
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

	notes := []string{}
	factorsUsed := map[string]FactorRef{}

	// ── Read tep-factor + density overrides from FactorBundle. ──────────
	tepFactorOf := func(vector string) (value float64, version string) {
		v, ver, ok := factors.Get(FactorTepFactorPrefix + vector)
		if ok {
			factorsUsed[FactorTepFactorPrefix+vector] = FactorRef{Value: v, Unit: tepFactorUnit(vector), Version: ver}
			return v, ver
		}
		def, _ := statutoryTepFactor(vector)
		return def, "statutory-default"
	}
	densityOf := func(vector string) float64 {
		v, ver, ok := factors.Get(FactorDensityPrefix + vector)
		if ok {
			factorsUsed[FactorDensityPrefix+vector] = FactorRef{Value: v, Unit: "kg/l", Version: ver}
			return v
		}
		return defaultDensity(vector)
	}

	// ── Aggregate readings → per-(site, vector) primary-unit totals. ────
	// Map structure: site_id → vector → primary-unit total.
	type vecAcc struct {
		primaryTotal float64
		kWhTotal     float64
	}
	siteData := map[uuid.UUID]map[string]*vecAcc{}
	totalsByVector := map[string]*vecAcc{}
	var unclassifiedRows int64

	iter := readings.Iter()
	for iter.Next() {
		row := iter.Row()
		c, ok := classifyByUnit(row.Unit)
		if !ok {
			unclassifiedRows += row.Count
			continue
		}

		// Convert primary value:
		// - Electricity: Wh → kWh requires /1000; kWh as-is.
		// - Liquid fuels (diesel / petrol): l → kg via density.
		// - Others (Sm3, kg_*): primary as-is.
		var primary float64
		switch row.Unit {
		case "Wh":
			primary = float64(row.Sum) / 1000.0
		case "kWh", "Sm3", "kg_lpg", "kg_coal", "kg_heavy_fuel":
			primary = float64(row.Sum)
		case "l_diesel", "l_diesel_vehicle", "l_petrol":
			d := densityOf(c.vector)
			primary = float64(row.Sum) * d
		default:
			primary = float64(row.Sum)
		}

		kWhEq := primary * c.kWhPerPrimary // 0 for non-electricity until LHV lands

		if _, ok := siteData[row.MeterID]; !ok {
			siteData[row.MeterID] = map[string]*vecAcc{}
		}
		acc := siteData[row.MeterID][c.vector]
		if acc == nil {
			acc = &vecAcc{}
			siteData[row.MeterID][c.vector] = acc
		}
		acc.primaryTotal += primary
		acc.kWhTotal += kWhEq

		totalAcc := totalsByVector[c.vector]
		if totalAcc == nil {
			totalAcc = &vecAcc{}
			totalsByVector[c.vector] = totalAcc
		}
		totalAcc.primaryTotal += primary
		totalAcc.kWhTotal += kWhEq
	}
	if err := iter.Err(); err != nil {
		return reporting.Report{}, fmt.Errorf("readings iteration: %w", err)
	}

	// ── Build by-vector totals (canonical order, every vector that appears). ──
	byVector := make([]VectorRow, 0, len(vectorOrdering))
	var totalKWh, totalTep float64
	for _, v := range vectorOrdering {
		acc, ok := totalsByVector[v]
		if !ok {
			continue
		}
		factor, _ := tepFactorOf(v)
		tep := acc.primaryTotal * factor
		byVector = append(byVector, VectorRow{
			Vector:       v,
			PrimaryUnit:  primaryUnitOf(v),
			PrimaryTotal: acc.primaryTotal,
			KWhTotal:     acc.kWhTotal,
			TepTotal:     tep,
			TepFactor:    factor,
		})
		totalKWh += acc.kWhTotal
		totalTep += tep
	}

	// ── Build site breakdown (sorted by site_id for determinism). ─────
	siteIDs := make([]uuid.UUID, 0, len(siteData))
	for id := range siteData {
		siteIDs = append(siteIDs, id)
	}
	sort.Slice(siteIDs, func(i, j int) bool {
		return siteIDs[i].String() < siteIDs[j].String()
	})
	sites := make([]SiteRow, 0, len(siteIDs))
	for _, id := range siteIDs {
		vmap := siteData[id]
		row := SiteRow{SiteID: id.String()}
		for _, v := range vectorOrdering {
			acc, ok := vmap[v]
			if !ok {
				continue
			}
			factor, _ := tepFactorOf(v)
			tep := acc.primaryTotal * factor
			row.ByVector = append(row.ByVector, VectorRow{
				Vector:       v,
				PrimaryUnit:  primaryUnitOf(v),
				PrimaryTotal: acc.primaryTotal,
				KWhTotal:     acc.kWhTotal,
				TepTotal:     tep,
				TepFactor:    factor,
			})
			row.KWhTotal += acc.kWhTotal
			row.TepTotal += tep
		}
		sites = append(sites, row)
	}

	// ── Obligation block. ───────────────────────────────────────────────
	obligation := ObligationBlock{
		Type:                  obligationName(0),
		AuditPeriodicityYears: AuditPeriodicityYears,
	}
	if v, ver, ok := factors.Get(FactorObligationType); ok {
		obligation.Type = obligationName(v)
		factorsUsed[FactorObligationType] = FactorRef{Value: v, Unit: "selector", Version: ver}
	}
	if v, ver, ok := factors.Get(FactorBelow50Exempt); ok {
		if int(v) == 1 {
			obligation.Below50TepExemption = true
		}
		factorsUsed[FactorBelow50Exempt] = FactorRef{Value: v, Unit: "selector", Version: ver}
	}
	if v, ver, ok := factors.Get(FactorExemptionISO); ok {
		obligation.ExemptionBasis = exemptionName(v)
		factorsUsed[FactorExemptionISO] = FactorRef{Value: v, Unit: "selector", Version: ver}
	}
	if v, ver, ok := factors.Get(FactorNextDeadlineUnix); ok {
		t := time.Unix(int64(v), 0).UTC()
		iso := t.Format(time.RFC3339)
		obligation.NextDeadlineISO = &iso
		factorsUsed[FactorNextDeadlineUnix] = FactorRef{Value: v, Unit: "unix-seconds", Version: ver}
	}

	// ── EnPI candidates. ────────────────────────────────────────────────
	enpi := []EnPIRow{}
	if floor, ver, ok := factors.Get(FactorTotalFloorM2); ok && floor > 0 && totalKWh > 0 {
		factorsUsed[FactorTotalFloorM2] = FactorRef{Value: floor, Unit: "m2", Version: ver}
		baselineLabel := period.StartInclusiveUTC.Format("2006") + "–" + period.EndExclusiveUTC.Add(-time.Second).Format("2006")
		enpi = append(enpi, EnPIRow{
			Name:           "Total energy intensity",
			Indicator:      "kWh/m2",
			Value:          totalKWh / floor,
			BaselinePeriod: baselineLabel,
		})
	}

	// ── Notes. ──────────────────────────────────────────────────────────
	if unclassifiedRows > 0 {
		notes = append(notes, fmt.Sprintf("%d reading(s) had unclassified Unit values; not in baseline totals.", unclassifiedRows))
	}
	if obligation.Below50TepExemption {
		notes = append(notes, "Below-50-tep exemption asserted by engagement Phase 0 Discovery (D.Lgs. 73/2020 → art. 8 c. 3-bis); Pack does not block but flags exemption_basis.")
	}
	hasNonElectricity := false
	for _, vr := range byVector {
		if vr.Vector != VectorElectricity && vr.PrimaryTotal > 0 {
			hasNonElectricity = true
			break
		}
	}
	if hasNonElectricity {
		notes = append(notes, "Non-electricity vectors are reported in primary unit + tep; kWh-equivalent for these vectors requires LHV factors via FactorBundle audit102_lhv_<vector> (deferred to Phase F per CHARTER §4 Tradeoff stanza).")
	}
	notes = append(notes, "Site-classification (industrial vs tertiary) and the Allegato 2 cluster-selection rules (≥10 000 tep industrial mandatory; ≥1 000 tep tertiary mandatory; <100 tep optionally excludable subject to the 20%-of-total cap) are computed by the engagement-fork's reporting orchestrator from ATECO codes (engagement-side knowledge).")

	body := Body{
		Report:      "audit_dlgs102",
		Regulator:   "ENEA / MASE — D.Lgs. 4 luglio 2014 n. 102 art. 8",
		Period:      period,
		Obligation:  obligation,
		FactorsUsed: factorsUsed,
		EnergyBaseline: EnergyBaselineBlock{
			ByVector:            byVector,
			TotalKWh:            totalKWh,
			TotalTep:            totalTep,
			Below50TepThreshold: totalTep < Below50TepThreshold,
		},
		SiteBreakdown:       sites,
		EnPI:                enpi,
		ImprovementMeasures: []ImprovementMeasure{},
		MonitoringPlan: MonitoringPlanBlock{
			KPIs:                  []string{},
			ReviewFrequencyMonths: 12,
		},
		NarrativeDataPoints: NarrativeBlock{
			Note: "Narrative content (UNI CEI EN 16247-1..4 compliance statement, site-visit summaries, scope description, EGE / ESCo certifier identity, improvement-measure list with payback analysis, monitoring-plan KPIs) is bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4).",
		},
		EGECertRequired:  true,
		UnclassifiedRows: unclassifiedRows,
		Notes:            notes,
	}

	encoded, err := encode(body)
	if err != nil {
		return reporting.Report{}, fmt.Errorf("encode: %w", err)
	}
	return reporting.Report{
		Type:    ReportType,
		Period:  period,
		Body:    body,
		Encoded: encoded,
	}, nil
}

func primaryUnitOf(vector string) string {
	switch vector {
	case VectorElectricity:
		return "kWh"
	case VectorNaturalGas:
		return "Sm3"
	}
	// All other vectors are reported in kg.
	return "kg"
}

func tepFactorUnit(vector string) string {
	switch vector {
	case VectorElectricity:
		return "tep/kWh"
	case VectorNaturalGas:
		return "tep/Sm3"
	}
	return "tep/kg"
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
