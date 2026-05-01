// Package conto_termico implements the Conto Termico Report Pack — the
// Italian thermal-energy-efficiency incentive scheme administered by GSE.
//
// Covers DM 16/02/2016 (Conto Termico 2.0; €900M/year national budget,
// 1-shot pagamento ≤ €5 000 or 2-5 year annual rates) and DM 7/08/2025
// (Conto Termico 3.0; capital grant up to 65% of eligible costs, 90-day
// submission window, expanded beneficiary universe including ETS / CER).
//
// Pack accepts the per-intervention incentive amount from engagement-
// supplied scenario inputs (per-intervention formulas land in Phase F)
// and computes the payment schedule + cap check + regime selector.
// Narrative content (intervention description, ATECO codes, beneficiary
// contact, supplier invoices, certifier ID, building data) is null in
// this Pack — the engagement-fork's reporting orchestrator injects
// client-supplied content before signing (Rule 144).
//
// Pure function per Rule 91. Deterministic serialisation per Rule 141.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/reporting/builder.go
//   - Manifest:         packs/report/conto_termico/manifest.yaml
//   - Charter:          packs/report/conto_termico/CHARTER.md
package conto_termico

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/greenmetrics/backend/internal/domain/reporting"
)

var Pack reporting.Builder = &builder{}

const (
	PackVersion = "1.0.0"

	ReportType reporting.ReportType = "conto_termico"

	// FactorBundle scenario keys consulted by this Builder.
	FactorRegimeVersion        = "conto_termico_regime_version"
	FactorInterventionCategory = "conto_termico_intervention_category"
	FactorBeneficiaryType      = "conto_termico_beneficiary_type"
	FactorIncentiveAmountEUR   = "conto_termico_incentive_amount_eur"
	FactorEligibleCostsEUR     = "conto_termico_eligible_costs_eur"
	FactorClimateZone          = "conto_termico_climate_zone"
	FactorPaymentYearsOverride = "conto_termico_payment_years_override"

	// Regime selectors.
	RegimeCT20 = "ct-2-0"
	RegimeCT30 = "ct-3-0"

	// Beneficiary types.
	BeneficiaryPA      = 0
	BeneficiaryPrivato = 1
	BeneficiaryETS     = 2
	BeneficiaryCER     = 3

	// Payment modes.
	PaymentSingleTranche = "single_tranche"
	PaymentAnnualRates   = "annual_rates"
	PaymentCapitalGrant  = "capital_grant"

	// ── CT 2.0 thresholds. ──────────────────────────────────────────────
	// Source: DM 16/02/2016 art. 8 c. 8 — soglia di erogazione in unica
	//         soluzione innalzata da €600 a €5 000.
	CT20SingleTrancheThresholdEUR = 5_000.0
	// DM 16/02/2016: rate annuali costanti di durata 2-5 anni a seconda
	// dell'intervento. Default 2 anni per impianti FER ≤ 35 kWt; 5 anni
	// per interventi maggiori. Engagement override available via
	// FactorPaymentYearsOverride.
	CT20DefaultPaymentYears = 2
	// Submission window post-completion (CT 2.0).
	CT20SubmissionWindowDays = 60

	// ── CT 3.0 thresholds. ──────────────────────────────────────────────
	// Source: DM 7/08/2025 — contributo in conto capitale fino al 65%
	//         delle spese ammissibili.
	CT30MaxCapPct = 0.65
	// Submission window post-completion (CT 3.0).
	CT30SubmissionWindowDays = 90
)

type builder struct{}

func (b *builder) Type() reporting.ReportType { return ReportType }
func (b *builder) Version() string            { return PackVersion }

type Body struct {
	Report              string               `json:"report"`
	Regulator           string               `json:"regulator"`
	RegimeVersion       string               `json:"regime_version"`
	Period              reporting.Period     `json:"period"`
	FactorsUsed         map[string]FactorRef `json:"factors_used"`
	Intervention        InterventionBlock    `json:"intervention"`
	EligibleCosts       EligibleCostsBlock   `json:"eligible_costs"`
	PaymentSchedule     PaymentScheduleBlock `json:"payment_schedule"`
	NarrativeDataPoints NarrativeBlock       `json:"narrative_data_points"`
	EGECertRequired     bool                 `json:"ege_certification_required"`
	Notes               []string             `json:"notes,omitempty"`
}

type FactorRef struct {
	Value   float64 `json:"value"`
	Unit    string  `json:"unit"`
	Version string  `json:"version"`
}

type InterventionBlock struct {
	CategoryCode *int    `json:"category_code"`
	Beneficiary  string  `json:"beneficiary"`
	ClimateZone  *string `json:"climate_zone"`
}

type EligibleCostsBlock struct {
	TotalEUR     float64  `json:"total_eur"`
	MaxCapPct    *float64 `json:"max_cap_pct"`
	MaxCapEUR    *float64 `json:"max_cap_eur"`
	CapViolation bool     `json:"cap_violation"`
	CapExcessEUR float64  `json:"cap_excess_eur"`
}

type PaymentScheduleBlock struct {
	RegimeVersion        string  `json:"regime_version"`
	IncentiveAmountEUR   float64 `json:"incentive_amount_eur"`
	PaymentMode          string  `json:"payment_mode"`
	PaymentYears         int     `json:"payment_years"`
	AnnualRateEUR        float64 `json:"annual_rate_eur"`
	SubmissionWindowDays int     `json:"submission_window_days"`
}

type NarrativeBlock struct {
	InterventionDescription *string `json:"intervention_description"`
	ATECOCodes              *string `json:"ateco_codes"`
	BeneficiaryContact      *string `json:"beneficiary_contact"`
	SupplierInvoices        *string `json:"supplier_invoices"`
	CertifierID             *string `json:"certifier_id"`
	BuildingData            *string `json:"building_data"`
	Note                    string  `json:"note"`
}

func regimeName(selector float64) string {
	if int(selector) == 1 {
		return RegimeCT30
	}
	return RegimeCT20
}

func beneficiaryName(selector float64) string {
	switch int(selector) {
	case BeneficiaryPrivato:
		return "privato"
	case BeneficiaryETS:
		return "ETS"
	case BeneficiaryCER:
		return "CER"
	default:
		return "PA"
	}
}

// climateZoneName decodes the integer climate-zone selector (1=A..6=F)
// per D.P.R. 412/93 + UNI 5364.
func climateZoneName(selector float64) *string {
	zones := []string{"A", "B", "C", "D", "E", "F"}
	idx := int(selector) - 1
	if idx < 0 || idx >= len(zones) {
		return nil
	}
	return &zones[idx]
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

	readScenario := func(key, unit string) (float64, bool) {
		v, ver, ok := factors.Get(key)
		if !ok {
			return 0, false
		}
		factorsUsed[key] = FactorRef{Value: v, Unit: unit, Version: ver}
		return v, true
	}

	// ── Regime version (default CT 2.0). ─────────────────────────────
	regimeSelector, _ := readScenario(FactorRegimeVersion, "selector")
	regime := regimeName(regimeSelector)

	// ── Intervention details. ────────────────────────────────────────
	var categoryCode *int
	if v, ver, ok := factors.Get(FactorInterventionCategory); ok {
		c := int(v)
		categoryCode = &c
		factorsUsed[FactorInterventionCategory] = FactorRef{Value: v, Unit: "category-code", Version: ver}
	}

	beneficiarySelector, _ := readScenario(FactorBeneficiaryType, "selector")
	beneficiary := beneficiaryName(beneficiarySelector)

	var climateZone *string
	if v, ver, ok := factors.Get(FactorClimateZone); ok {
		climateZone = climateZoneName(v)
		factorsUsed[FactorClimateZone] = FactorRef{Value: v, Unit: "selector", Version: ver}
	}

	// ── Incentive amount + eligible costs. ───────────────────────────
	incentiveAmount, hasIncentive := readScenario(FactorIncentiveAmountEUR, "EUR")
	eligibleCosts, hasCosts := readScenario(FactorEligibleCostsEUR, "EUR")

	if !hasIncentive {
		notes = append(notes, "Incentive amount missing; payment schedule emits zero. Engagement Phase 0 supplies conto_termico_incentive_amount_eur computed via the per-intervention formula (Allegato I-II GSE Regole Applicative).")
	}

	// ── Eligible-cost cap check (CT 3.0 only). ───────────────────────
	costsBlock := EligibleCostsBlock{TotalEUR: eligibleCosts}
	if regime == RegimeCT30 && hasCosts {
		maxCapPct := CT30MaxCapPct
		maxCapEUR := eligibleCosts * maxCapPct
		costsBlock.MaxCapPct = &maxCapPct
		costsBlock.MaxCapEUR = &maxCapEUR
		if hasIncentive && incentiveAmount > maxCapEUR {
			costsBlock.CapViolation = true
			costsBlock.CapExcessEUR = incentiveAmount - maxCapEUR
			notes = append(notes, fmt.Sprintf("Incentive amount €%.2f exceeds the CT 3.0 65%%-of-eligible-costs cap (€%.2f); excess €%.2f surfaced — engagement should reduce request to cap.",
				incentiveAmount, maxCapEUR, costsBlock.CapExcessEUR))
		}
	}

	// ── Payment schedule. ────────────────────────────────────────────
	scheduleBlock := PaymentScheduleBlock{
		RegimeVersion:      regime,
		IncentiveAmountEUR: incentiveAmount,
	}
	switch regime {
	case RegimeCT30:
		scheduleBlock.PaymentMode = PaymentCapitalGrant
		scheduleBlock.PaymentYears = 1
		scheduleBlock.AnnualRateEUR = incentiveAmount
		scheduleBlock.SubmissionWindowDays = CT30SubmissionWindowDays
	default: // RegimeCT20
		scheduleBlock.SubmissionWindowDays = CT20SubmissionWindowDays
		if !hasIncentive || incentiveAmount <= CT20SingleTrancheThresholdEUR {
			scheduleBlock.PaymentMode = PaymentSingleTranche
			scheduleBlock.PaymentYears = 1
			scheduleBlock.AnnualRateEUR = incentiveAmount
		} else {
			scheduleBlock.PaymentMode = PaymentAnnualRates
			years := CT20DefaultPaymentYears
			if v, ver, ok := factors.Get(FactorPaymentYearsOverride); ok && int(v) > 0 {
				years = int(v)
				factorsUsed[FactorPaymentYearsOverride] = FactorRef{Value: v, Unit: "years", Version: ver}
			}
			if years < 2 || years > 5 {
				notes = append(notes, fmt.Sprintf("Override payment_years=%d is outside the DM 16/02/2016 art. 8 c. 6 range [2,5]; verify against the per-intervention class declared in Allegato I-II.", years))
			}
			scheduleBlock.PaymentYears = years
			if years > 0 {
				scheduleBlock.AnnualRateEUR = incentiveAmount / float64(years)
			}
		}
	}

	// ── Drain readings (the Pack is scenario-driven). ────────────────
	var rowCount int64
	iter := readings.Iter()
	for iter.Next() {
		rowCount++
	}
	if err := iter.Err(); err != nil {
		return reporting.Report{}, fmt.Errorf("readings iteration: %w", err)
	}
	if rowCount > 0 {
		notes = append(notes, fmt.Sprintf("Pack consumed scenario inputs from FactorBundle; %d reading row(s) ignored (Phase F adds reading-derived per-intervention formula path).", rowCount))
	}

	body := Body{
		Report:        "conto_termico",
		Regulator:     "GSE / MASE — DM 16/02/2016 (CT 2.0); DM 7/08/2025 (CT 3.0)",
		RegimeVersion: regime,
		Period:        period,
		FactorsUsed:   factorsUsed,
		Intervention: InterventionBlock{
			CategoryCode: categoryCode,
			Beneficiary:  beneficiary,
			ClimateZone:  climateZone,
		},
		EligibleCosts:   costsBlock,
		PaymentSchedule: scheduleBlock,
		NarrativeDataPoints: NarrativeBlock{
			Note: "Narrative content (intervention description, ATECO codes, beneficiary contact, supplier invoices, certifier ID, building data) is bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4).",
		},
		EGECertRequired: true,
		Notes:           notes,
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
