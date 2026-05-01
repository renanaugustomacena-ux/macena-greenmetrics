// Package tee implements the TEE / Certificati Bianchi (Titoli di
// Efficienza Energetica) Report Pack — the GSE-administered tradeable-
// certificate scheme established by D.M. 11 gennaio 2017 (and updated
// for the period 2025-2030 by DM MASE 21 luglio 2025).
//
// One TEE = 1 tep certified annual energy saving. The Pack computes
// ex-ante / ex-post tep, applies the K1=1.2 / K2=0.8 multiplicative
// coefficient determined by the project's vita utile half, and emits
// the certificate count for the current year. Narrative content
// (project description, EGE certifier ID, RVC submission ID, vita utile
// methodology) is null in this Pack — the engagement-fork's reporting
// orchestrator injects client-supplied content before signing
// (Rule 144).
//
// Pure function per Rule 91. Deterministic serialisation per Rule 141.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/reporting/builder.go
//   - Manifest:         packs/report/tee/manifest.yaml
//   - Charter:          packs/report/tee/CHARTER.md
package tee

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

	ReportType reporting.ReportType = "tee"

	// FactorBundle scenario keys consulted by this Builder.
	FactorMethod               = "tee_method"
	FactorExAnteTep            = "tee_ex_ante_tep"
	FactorExPostTep            = "tee_ex_post_tep"
	FactorVitaUtileYears       = "tee_vita_utile_years"
	FactorCurrentYearInProject = "tee_current_year_in_project"
	FactorInterventionCategory = "tee_intervention_category"
	FactorRegimeVersion        = "tee_regime_version"

	// Method selectors.
	MethodConsuntivo     = 0
	MethodStandardizzato = 1
	MethodPPPM           = 2

	// Regime selectors.
	RegimeDM2017       = "dm-2017"
	RegimeDMMase202530 = "dm-mase-2025-2030"

	// ── K-factor constants — DM 11/01/2017. ─────────────────────────────
	// Source: DM 11/01/2017 art. 8 c. 4. The vecchio coefficiente "tau"
	//         (DM 28/12/2012) è stato sostituito da K1=1.2 (prima metà
	//         vita utile) e K2=0.8 (seconda metà vita utile).
	KFactorFirstHalf  = 1.2 // K1 — first half of vita utile
	KFactorSecondHalf = 0.8 // K2 — second half of vita utile
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
	Project             ProjectBlock         `json:"project"`
	EnergySavings       EnergySavingsBlock   `json:"energy_savings"`
	TEECalculation      TEECalculationBlock  `json:"tee_calculation"`
	NarrativeDataPoints NarrativeBlock       `json:"narrative_data_points"`
	EGECertRequired     bool                 `json:"ege_certification_required"`
	Notes               []string             `json:"notes,omitempty"`
}

type FactorRef struct {
	Value   float64 `json:"value"`
	Unit    string  `json:"unit"`
	Version string  `json:"version"`
}

type ProjectBlock struct {
	Method               string  `json:"method"`
	InterventionCategory *int    `json:"intervention_category"`
	VitaUtileYears       float64 `json:"vita_utile_years"`
	CurrentYearInProject float64 `json:"current_year_in_project"`
	CurrentPeriodHalf    string  `json:"current_period_half"`
}

type EnergySavingsBlock struct {
	ExAnteTep       float64 `json:"ex_ante_tep"`
	ExPostTep       float64 `json:"ex_post_tep"`
	AnnualSavingTep float64 `json:"annual_saving_tep"`
	SavingPct       float64 `json:"saving_pct"`
}

type TEECalculationBlock struct {
	RegimeVersion   string  `json:"regime_version"`
	AnnualSavingTep float64 `json:"annual_saving_tep"`
	KFactor         float64 `json:"k_factor"`
	TEEIssued       float64 `json:"tee_issued"`
}

// NarrativeBlock — placeholder for client-supplied narrative content.
type NarrativeBlock struct {
	ProjectDescription        *string `json:"project_description"`
	EGECertifierID            *string `json:"ege_certifier_id"`
	RVCSubmissionID           *string `json:"rvc_submission_id"`
	VitaUtileMethodology      *string `json:"vita_utile_methodology"`
	BaselineCalculationMethod *string `json:"baseline_calculation_method"`
	Note                      string  `json:"note"`
}

func methodName(selector float64) string {
	switch int(selector) {
	case MethodStandardizzato:
		return "standardizzato"
	case MethodPPPM:
		return "pppm"
	default:
		return "consuntivo"
	}
}

func regimeName(selector float64) string {
	if int(selector) == 1 {
		return RegimeDMMase202530
	}
	return RegimeDM2017
}

// kFactorForHalf returns K1 (first half of vita utile) or K2 (second half).
//
// Source: DM 11/01/2017 art. 8 c. 4. The boundary at half_threshold is
//
//	inclusive on the lower end (current_year ≤ half ⇒ first half).
func kFactorForHalf(currentYear, vitaUtile float64) (float64, string) {
	if vitaUtile <= 0 {
		return 0, "undefined"
	}
	halfThreshold := vitaUtile / 2.0
	if currentYear <= halfThreshold {
		return KFactorFirstHalf, "first"
	}
	return KFactorSecondHalf, "second"
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

	exAnteTep, hasExAnte := readScenario(FactorExAnteTep, "tep/year")
	exPostTep, hasExPost := readScenario(FactorExPostTep, "tep/year")
	vitaUtile, hasVita := readScenario(FactorVitaUtileYears, "year")
	currentYear, hasCY := readScenario(FactorCurrentYearInProject, "year")

	// Method selector (default consuntivo).
	method := methodName(0)
	if v, ver, ok := factors.Get(FactorMethod); ok {
		method = methodName(v)
		factorsUsed[FactorMethod] = FactorRef{Value: v, Unit: "selector", Version: ver}
	}

	// Intervention category (engagement-supplied opaque integer).
	var interventionCategory *int
	if v, ver, ok := factors.Get(FactorInterventionCategory); ok {
		c := int(v)
		interventionCategory = &c
		factorsUsed[FactorInterventionCategory] = FactorRef{Value: v, Unit: "category-code", Version: ver}
	}

	// Regime version (default DM-2017).
	regimeSelector := 0.0
	if v, ver, ok := factors.Get(FactorRegimeVersion); ok {
		regimeSelector = v
		factorsUsed[FactorRegimeVersion] = FactorRef{Value: v, Unit: "selector", Version: ver}
	}
	regime := regimeName(regimeSelector)
	if regime == RegimeDMMase202530 {
		notes = append(notes, "Regime DM MASE 21/07/2025 (2025-2030) selected; v1.0.0 of this Pack falls back to DM-2017 K-factor semantics — engagement Phase 0 Discovery must verify the 2025-2030 detailed coefficients against the decreto vigente at RVC submission. Pack v1.x.0 minor update to follow regulatory consolidation per Rule 138.")
	}

	// ── Energy savings. ──────────────────────────────────────────────
	var annualSaving, savingPct float64
	if hasExAnte && hasExPost {
		annualSaving = exAnteTep - exPostTep
		if exAnteTep > 0 {
			savingPct = (annualSaving / exAnteTep) * 100.0
		}
	} else {
		notes = append(notes, "Ex-ante / ex-post tep inputs missing; annual saving and TEE issued set to 0.")
	}

	// ── K-factor. ────────────────────────────────────────────────────
	var kFactor float64
	half := "undefined"
	if hasVita && hasCY {
		kFactor, half = kFactorForHalf(currentYear, vitaUtile)
	} else {
		notes = append(notes, "Vita utile / current_year_in_project missing; K-factor undefined; TEE issued set to 0.")
	}

	// ── TEE issued. ──────────────────────────────────────────────────
	teeIssued := annualSaving * kFactor
	if annualSaving <= 0 {
		teeIssued = 0
		if hasExAnte && hasExPost {
			notes = append(notes, "Annual saving is non-positive; project ineligible for TEE issuance under DM 11/01/2017 art. 4.")
		}
	}

	// Drain readings — Pack is purely scenario-driven; Phase F adds reading-
	// derived ex-ante / ex-post via a separate ingestion pre-step.
	var rowCount int64
	iter := readings.Iter()
	for iter.Next() {
		rowCount++
	}
	if err := iter.Err(); err != nil {
		return reporting.Report{}, fmt.Errorf("readings iteration: %w", err)
	}
	if rowCount > 0 {
		notes = append(notes, fmt.Sprintf("Pack consumed scenario inputs from FactorBundle; %d reading row(s) ignored (Phase F adds reading-derived ex-ante / ex-post path).", rowCount))
	}

	body := Body{
		Report:        "tee",
		Regulator:     "GSE / MASE — DM 11/01/2017; DM MASE 21/07/2025 (regime 2025-2030)",
		RegimeVersion: regime,
		Period:        period,
		FactorsUsed:   factorsUsed,
		Project: ProjectBlock{
			Method:               method,
			InterventionCategory: interventionCategory,
			VitaUtileYears:       vitaUtile,
			CurrentYearInProject: currentYear,
			CurrentPeriodHalf:    half,
		},
		EnergySavings: EnergySavingsBlock{
			ExAnteTep:       exAnteTep,
			ExPostTep:       exPostTep,
			AnnualSavingTep: annualSaving,
			SavingPct:       savingPct,
		},
		TEECalculation: TEECalculationBlock{
			RegimeVersion:   regime,
			AnnualSavingTep: annualSaving,
			KFactor:         kFactor,
			TEEIssued:       teeIssued,
		},
		NarrativeDataPoints: NarrativeBlock{
			Note: "Narrative content (project description, EGE certifier ID, RVC submission ID, vita utile methodology, baseline calculation methodology) is bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4).",
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
