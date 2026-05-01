// Package piano_5_0 implements the Piano Transizione 5.0 attestation Report
// Pack — DL 19/2024 art. 38 + DM 24/07/2024 + L. 207/2024 art. 1 c. 427-429.
//
// Produces the dual-criterion energy-saving classification (struttura
// produttiva vs processi interessati — higher tier wins) and the per-
// bracket tax-credit calculation. Narrative content (ex-ante / ex-post
// EGE certifications, Allegato A categorisation, investment description,
// ATECO codes) is null in this Pack — the engagement-fork's reporting
// orchestrator injects client-supplied content before signing (Rule 144).
//
// Pure function per Rule 91. Deterministic serialisation per Rule 141.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/reporting/builder.go
//   - Manifest:         packs/report/piano_5_0/manifest.yaml
//   - Charter:          packs/report/piano_5_0/CHARTER.md
package piano_5_0

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

	ReportType reporting.ReportType = "piano_5_0"

	// FactorBundle scenario keys consulted by this Builder.
	FactorBaselineEnergyKWh        = "piano5_baseline_energy_kwh"
	FactorCounterfactualEnergyKWh  = "piano5_counterfactual_energy_kwh"
	FactorBaselineProcessKWh       = "piano5_baseline_process_kwh"
	FactorCounterfactualProcessKWh = "piano5_counterfactual_process_kwh"
	FactorInvestmentTotalEUR       = "piano5_investment_total_eur"
	FactorRegimeVersion            = "piano5_regime_version"
	FactorYearlyCapEUR             = "piano5_period_year_eur_cap"

	// Regime selectors.
	RegimeLB2025   = "lb-2025"
	RegimeDM240724 = "dm-2024-07-24"

	// Source: DL 19/2024 art. 38 c. 4 — limite massimo annuo €50M per beneficiario.
	DefaultAnnualCapEUR = 50_000_000.0
)

type builder struct{}

func (b *builder) Type() reporting.ReportType { return ReportType }
func (b *builder) Version() string            { return PackVersion }

// Body is the canonical typed payload. Field order is the deterministic
// JSON-encoding order (Rule 141).
type Body struct {
	Report              string               `json:"report"`
	Regulator           string               `json:"regulator"`
	RegimeVersion       string               `json:"regime_version"`
	Period              reporting.Period     `json:"period"`
	FactorsUsed         map[string]FactorRef `json:"factors_used"`
	Investment          InvestmentBlock      `json:"investment"`
	EnergySavings       EnergySavingsBlock   `json:"energy_savings"`
	TaxCredit           TaxCreditBlock       `json:"tax_credit"`
	NarrativeDataPoints NarrativeBlock       `json:"narrative_data_points"`
	EGECertRequired     bool                 `json:"ege_certification_required"`
	Notes               []string             `json:"notes,omitempty"`
}

// FactorRef captures one factor consulted (identical shape across Report Packs).
type FactorRef struct {
	Value   float64 `json:"value"`
	Unit    string  `json:"unit"`
	Version string  `json:"version"`
}

type InvestmentBlock struct {
	TotalEUR          float64 `json:"total_eur"`
	AnnualCapEUR      float64 `json:"annual_cap_eur"`
	AboveCapExcessEUR float64 `json:"above_cap_excess_eur"`
}

type EnergySavingsBlock struct {
	StrutturaProduttiva ScenarioRow `json:"struttura_produttiva"`
	ProcessiInteressati ScenarioRow `json:"processi_interessati"`
	EffectiveTier       int         `json:"effective_tier"`
	EffectiveBasis      string      `json:"effective_basis"`
}

// ScenarioRow is one savings-criterion result.
type ScenarioRow struct {
	BaselineKWh       float64 `json:"baseline_kwh"`
	CounterfactualKWh float64 `json:"counterfactual_kwh"`
	SavingKWh         float64 `json:"saving_kwh"`
	SavingPct         float64 `json:"saving_pct"`
	Tier              int     `json:"tier"`
}

type TaxCreditBlock struct {
	RegimeVersion  string          `json:"regime_version"`
	AppliedTier    int             `json:"applied_tier"`
	RateTable      []BracketRate   `json:"rate_table"`
	PerBracketEUR  []BracketCredit `json:"per_bracket_eur"`
	TotalCreditEUR float64         `json:"total_credit_eur"`
}

type BracketRate struct {
	BracketEURMin     float64 `json:"bracket_eur_min"`
	BracketEURMax     float64 `json:"bracket_eur_max"`
	RateAtAppliedTier float64 `json:"rate_at_applied_tier"`
}

type BracketCredit struct {
	BracketEURMin          float64 `json:"bracket_eur_min"`
	BracketEURMax          float64 `json:"bracket_eur_max"`
	InvestmentInBracketEUR float64 `json:"investment_in_bracket_eur"`
	CreditEUR              float64 `json:"credit_eur"`
}

// NarrativeBlock — placeholder for client-supplied narrative content
// that the engagement-fork's reporting orchestrator injects.
type NarrativeBlock struct {
	ExAnteCertID          *string  `json:"ex_ante_certification_id"`
	ExPostCertID          *string  `json:"ex_post_certification_id"`
	InvestmentDescription *string  `json:"investment_description"`
	AnnexACategories      []string `json:"annex_a_categories"`
	ATECOCodes            []string `json:"ateco_codes"`
	Note                  string   `json:"note"`
}

// rateRegime captures the per-bracket rate matrix for one regime version.
//
// Source: DL 19/2024 art. 38 c. 7-10; DM MIMIT-MEF 24/07/2024 art. 9 c. 4;
//
//	L. 207/2024 (Legge di bilancio 2025) art. 1 c. 427-429.
//
// Two regimes are encoded:
//   - lb-2025      — current regime; LB 2025 collapsed the original
//     €0–€2.5M and €2.5M–€10M brackets into a single ≤€10M
//     bracket. Applies retroactively to investments started
//     1 January 2024 onwards.
//   - dm-2024-07-24 — original three-bracket regime per DM 24 luglio 2024
//     art. 9; retained for historical / non-retroactive
//     attestations.
type rateRegime struct {
	name     string
	brackets []bracket
}

type bracket struct {
	eurMin, eurMax float64
	// ratesByTier index 0 = T1, 1 = T2, 2 = T3.
	ratesByTier [3]float64
}

// Source: L. 207/2024 art. 1 c. 427-429 (LB 2025).
//
//	Verified against MIMIT primary publication; sources accessed 2026-04-30.
var regimeLB2025 = rateRegime{
	name: RegimeLB2025,
	brackets: []bracket{
		{eurMin: 0, eurMax: 10_000_000, ratesByTier: [3]float64{0.35, 0.40, 0.45}},
		{eurMin: 10_000_000, eurMax: 50_000_000, ratesByTier: [3]float64{0.05, 0.10, 0.15}},
	},
}

// Source: DM MIMIT-MEF 24/07/2024 art. 9 c. 4 (original three-bracket regime).
//
//	Verified against MIMIT primary publication; sources accessed 2026-04-30.
var regimeDM240724 = rateRegime{
	name: RegimeDM240724,
	brackets: []bracket{
		{eurMin: 0, eurMax: 2_500_000, ratesByTier: [3]float64{0.35, 0.40, 0.45}},
		{eurMin: 2_500_000, eurMax: 10_000_000, ratesByTier: [3]float64{0.15, 0.20, 0.25}},
		{eurMin: 10_000_000, eurMax: 50_000_000, ratesByTier: [3]float64{0.05, 0.10, 0.15}},
	},
}

func selectRegime(version string) rateRegime {
	if version == RegimeDM240724 {
		return regimeDM240724
	}
	return regimeLB2025
}

// tierForStruttura maps a struttura-produttiva saving-percentage to its tier.
//
// Source: DM 24/07/2024 art. 9 c. 2-3.
//
//	T1 — saving ≥ 3 %
//	T2 — saving > 6 %
//	T3 — saving > 10 %
func tierForStruttura(savingPct float64) int {
	if savingPct > 10.0 {
		return 3
	}
	if savingPct > 6.0 {
		return 2
	}
	if savingPct >= 3.0 {
		return 1
	}
	return 0
}

// tierForProcessi maps a processi-interessati saving-percentage to its tier.
//
// Source: DM 24/07/2024 art. 9 c. 2-3.
//
//	T1 — saving ≥ 5 %
//	T2 — saving > 10 %
//	T3 — saving > 15 %
func tierForProcessi(savingPct float64) int {
	if savingPct > 15.0 {
		return 3
	}
	if savingPct > 10.0 {
		return 2
	}
	if savingPct >= 5.0 {
		return 1
	}
	return 0
}

func computeSaving(baseline, counterfactual float64) (kwh, pct float64) {
	if baseline <= 0 {
		return 0, 0
	}
	saving := baseline - counterfactual
	return saving, (saving / baseline) * 100.0
}

func bracketInvestment(total, eurMin, eurMax float64) float64 {
	if total <= eurMin {
		return 0
	}
	if total >= eurMax {
		return eurMax - eurMin
	}
	return total - eurMin
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

	baselineEnergy, hasBE := readScenario(FactorBaselineEnergyKWh, "kWh")
	counterfactualEnergy, hasCE := readScenario(FactorCounterfactualEnergyKWh, "kWh")
	baselineProcess, hasBP := readScenario(FactorBaselineProcessKWh, "kWh")
	counterfactualProcess, hasCP := readScenario(FactorCounterfactualProcessKWh, "kWh")
	investmentEUR, hasInv := readScenario(FactorInvestmentTotalEUR, "EUR")

	// Regime selector (default LB-2025).
	regimeVersion := RegimeLB2025
	if v, ver, ok := factors.Get(FactorRegimeVersion); ok {
		if v == 1 {
			regimeVersion = RegimeDM240724
		}
		factorsUsed[FactorRegimeVersion] = FactorRef{Value: v, Unit: "selector", Version: ver}
	}
	regime := selectRegime(regimeVersion)

	// Annual cap (default €50M).
	annualCap := DefaultAnnualCapEUR
	if v, ver, ok := factors.Get(FactorYearlyCapEUR); ok {
		annualCap = v
		factorsUsed[FactorYearlyCapEUR] = FactorRef{Value: v, Unit: "EUR", Version: ver}
	}

	// ── Energy savings — dual criterion ─────────────────────────────
	var struttura, processi ScenarioRow
	if hasBE && hasCE {
		struttura.BaselineKWh = baselineEnergy
		struttura.CounterfactualKWh = counterfactualEnergy
		struttura.SavingKWh, struttura.SavingPct = computeSaving(baselineEnergy, counterfactualEnergy)
		struttura.Tier = tierForStruttura(struttura.SavingPct)
	} else {
		notes = append(notes, "Struttura produttiva scenario inputs missing; struttura tier set to 0 (ineligible).")
	}
	if hasBP && hasCP {
		processi.BaselineKWh = baselineProcess
		processi.CounterfactualKWh = counterfactualProcess
		processi.SavingKWh, processi.SavingPct = computeSaving(baselineProcess, counterfactualProcess)
		processi.Tier = tierForProcessi(processi.SavingPct)
	} else {
		notes = append(notes, "Processi interessati scenario inputs missing; processi tier set to 0 (ineligible).")
	}

	// Effective tier = max(struttura, processi). Higher tier wins; on a
	// tie the struttura-produttiva basis is reported (more conservative
	// audit trail; struttura aggregates the whole production unit).
	effectiveTier := struttura.Tier
	effectiveBasis := "struttura_produttiva"
	if processi.Tier > struttura.Tier {
		effectiveTier = processi.Tier
		effectiveBasis = "processi_interessati"
	}
	if effectiveTier == 0 {
		effectiveBasis = "ineligible"
	}

	// ── Investment block + cap ──────────────────────────────────────
	investBlock := InvestmentBlock{
		TotalEUR:     investmentEUR,
		AnnualCapEUR: annualCap,
	}
	cappedInvestment := investmentEUR
	if cappedInvestment > annualCap {
		investBlock.AboveCapExcessEUR = investmentEUR - annualCap
		cappedInvestment = annualCap
	}

	// ── Tax credit ──────────────────────────────────────────────────
	creditBlock := TaxCreditBlock{
		RegimeVersion: regime.name,
		AppliedTier:   effectiveTier,
		RateTable:     make([]BracketRate, 0, len(regime.brackets)),
		PerBracketEUR: make([]BracketCredit, 0, len(regime.brackets)),
	}

	for _, br := range regime.brackets {
		rateAtTier := 0.0
		if effectiveTier >= 1 && effectiveTier <= 3 {
			rateAtTier = br.ratesByTier[effectiveTier-1]
		}
		creditBlock.RateTable = append(creditBlock.RateTable, BracketRate{
			BracketEURMin:     br.eurMin,
			BracketEURMax:     br.eurMax,
			RateAtAppliedTier: rateAtTier,
		})

		var inBracket, credit float64
		if hasInv {
			inBracket = bracketInvestment(cappedInvestment, br.eurMin, br.eurMax)
			credit = inBracket * rateAtTier
		}
		creditBlock.PerBracketEUR = append(creditBlock.PerBracketEUR, BracketCredit{
			BracketEURMin:          br.eurMin,
			BracketEURMax:          br.eurMax,
			InvestmentInBracketEUR: inBracket,
			CreditEUR:              credit,
		})
		creditBlock.TotalCreditEUR += credit
	}

	if !hasInv {
		notes = append(notes, "Investment total missing; tax credit set to 0.")
	}
	if effectiveTier == 0 {
		notes = append(notes, "Effective tier is 0 (savings below 3% struttura / 5% processi); attestation ineligible per DM 24/07/2024 art. 9.")
	}
	if investBlock.AboveCapExcessEUR > 0 {
		notes = append(notes, fmt.Sprintf("Investment exceeds annual cap of €%.0f; €%.0f excess not eligible for credit this period (split across periods or beneficiary entities per DL 19/2024 art. 38 c. 4).",
			annualCap, investBlock.AboveCapExcessEUR))
	}

	// Pack is purely scenario-driven; readings are not consumed in this
	// version. Phase F adds a reading-derived counterfactual ingestion path.
	var rowCount int64
	iter := readings.Iter()
	for iter.Next() {
		rowCount++
	}
	if err := iter.Err(); err != nil {
		return reporting.Report{}, fmt.Errorf("readings iteration: %w", err)
	}
	if rowCount > 0 {
		notes = append(notes, fmt.Sprintf("Pack consumed scenario inputs from FactorBundle; %d reading row(s) ignored (Phase F adds reading-derived counterfactual path).", rowCount))
	}

	body := Body{
		Report:        "piano_5_0",
		Regulator:     "MIMIT / GSE (DL 19/2024 art. 38; DM 24/07/2024; L. 207/2024 art. 1 c. 427-429)",
		RegimeVersion: regime.name,
		Period:        period,
		FactorsUsed:   factorsUsed,
		Investment:    investBlock,
		EnergySavings: EnergySavingsBlock{
			StrutturaProduttiva: struttura,
			ProcessiInteressati: processi,
			EffectiveTier:       effectiveTier,
			EffectiveBasis:      effectiveBasis,
		},
		TaxCredit: creditBlock,
		NarrativeDataPoints: NarrativeBlock{
			Note: "Narrative content (ex-ante / ex-post EGE certifications, investment description, Allegato A categorisation, ATECO codes) is bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4).",
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
