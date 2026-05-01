package piano_5_0

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/greenmetrics/backend/internal/domain/reporting"
)

func TestPackImplementsBuilder(t *testing.T) {
	var _ reporting.Builder = Pack
	if Pack.Type() != ReportType {
		t.Fatalf("type: want %q, got %q", ReportType, Pack.Type())
	}
	if Pack.Version() != PackVersion {
		t.Fatalf("version: want %q, got %q", PackVersion, Pack.Version())
	}
}

func defaultPeriod() reporting.Period {
	return reporting.Period{
		StartInclusiveUTC: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndExclusiveUTC:   time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		Timezone:          "Europe/Rome",
	}
}

// scenarioFactors builds a FactorBundle for a Piano 5.0 scenario.
//
//	strutturaSavingPct, processiSavingPct — desired saving %; baselines fixed
//	investmentEUR — total eligible investment
//	regime — RegimeLB2025 (default) or RegimeDM240724 (selector value 1)
func scenarioFactors(strutturaSavingPct, processiSavingPct, investmentEUR float64, regime string) reporting.FactorBundle {
	const (
		baselineEnergy  = 1_000_000.0 // 1 GWh
		baselineProcess = 200_000.0   // 200 MWh
	)
	rows := map[string]factorRow{
		FactorBaselineEnergyKWh:        {value: baselineEnergy, version: "engagement-supplied"},
		FactorCounterfactualEnergyKWh:  {value: baselineEnergy * (1 - strutturaSavingPct/100), version: "engagement-supplied"},
		FactorBaselineProcessKWh:       {value: baselineProcess, version: "engagement-supplied"},
		FactorCounterfactualProcessKWh: {value: baselineProcess * (1 - processiSavingPct/100), version: "engagement-supplied"},
		FactorInvestmentTotalEUR:       {value: investmentEUR, version: "engagement-supplied"},
	}
	if regime == RegimeDM240724 {
		rows[FactorRegimeVersion] = factorRow{value: 1, version: "regime-selector"}
	}
	return newFactorBundle(rows)
}

// TestBuildBitPerfectReproducibility — Rule 89.
func TestBuildBitPerfectReproducibility(t *testing.T) {
	period := defaultPeriod()
	factors := scenarioFactors(10.0, 15.0, 4_000_000, RegimeLB2025)
	r1, _ := Pack.Build(context.Background(), period, factors, newReadings(nil))
	r2, _ := Pack.Build(context.Background(), period, factors, newReadings(nil))
	if !bytes.Equal(r1.Encoded, r2.Encoded) {
		t.Fatalf("Encoded bytes differ between two Build calls (Rule 89 violated)")
	}
}

// TestBuildTierClassificationStruttura verifies tier thresholds for struttura
// produttiva: T1 ≥ 3 %, T2 > 6 %, T3 > 10 %. (DM 24/07/2024 art. 9.)
func TestBuildTierClassificationStruttura(t *testing.T) {
	cases := []struct {
		savingPct float64
		wantTier  int
		wantBasis string
	}{
		{savingPct: 1.0, wantTier: 0, wantBasis: "ineligible"},
		{savingPct: 2.99, wantTier: 0, wantBasis: "ineligible"},
		{savingPct: 3.0, wantTier: 1, wantBasis: "struttura_produttiva"},
		{savingPct: 5.0, wantTier: 1, wantBasis: "struttura_produttiva"},
		{savingPct: 6.0, wantTier: 1, wantBasis: "struttura_produttiva"}, // exactly 6 % — strict >.
		{savingPct: 6.01, wantTier: 2, wantBasis: "struttura_produttiva"},
		{savingPct: 10.0, wantTier: 2, wantBasis: "struttura_produttiva"}, // exactly 10 % — strict >.
		{savingPct: 10.01, wantTier: 3, wantBasis: "struttura_produttiva"},
		{savingPct: 25.0, wantTier: 3, wantBasis: "struttura_produttiva"},
	}
	for _, tc := range cases {
		factors := scenarioFactors(tc.savingPct, 0, 1_000_000, RegimeLB2025)
		r, _ := Pack.Build(context.Background(), defaultPeriod(), factors, newReadings(nil))
		body := r.Body.(Body)
		if body.EnergySavings.StrutturaProduttiva.Tier != tc.wantTier {
			t.Errorf("struttura tier @ %v %%: want %d, got %d",
				tc.savingPct, tc.wantTier, body.EnergySavings.StrutturaProduttiva.Tier)
		}
		if body.EnergySavings.EffectiveTier != tc.wantTier {
			t.Errorf("effective tier @ %v %%: want %d, got %d",
				tc.savingPct, tc.wantTier, body.EnergySavings.EffectiveTier)
		}
		if body.EnergySavings.EffectiveBasis != tc.wantBasis {
			t.Errorf("effective basis @ %v %%: want %q, got %q",
				tc.savingPct, tc.wantBasis, body.EnergySavings.EffectiveBasis)
		}
	}
}

// TestBuildTierClassificationProcessi verifies tier thresholds for processi
// interessati: T1 ≥ 5 %, T2 > 10 %, T3 > 15 %. (DM 24/07/2024 art. 9.)
func TestBuildTierClassificationProcessi(t *testing.T) {
	cases := []struct {
		savingPct float64
		wantTier  int
	}{
		{savingPct: 4.99, wantTier: 0},
		{savingPct: 5.0, wantTier: 1},
		{savingPct: 10.0, wantTier: 1},
		{savingPct: 10.01, wantTier: 2},
		{savingPct: 15.0, wantTier: 2},
		{savingPct: 15.01, wantTier: 3},
	}
	for _, tc := range cases {
		factors := scenarioFactors(0, tc.savingPct, 1_000_000, RegimeLB2025)
		r, _ := Pack.Build(context.Background(), defaultPeriod(), factors, newReadings(nil))
		body := r.Body.(Body)
		if body.EnergySavings.ProcessiInteressati.Tier != tc.wantTier {
			t.Errorf("processi tier @ %v %%: want %d, got %d",
				tc.savingPct, tc.wantTier, body.EnergySavings.ProcessiInteressati.Tier)
		}
	}
}

// TestBuildEffectiveTierMaxOfBoth verifies that the higher of the two tiers
// wins, with effective_basis pointing at the criterion that delivered it.
func TestBuildEffectiveTierMaxOfBoth(t *testing.T) {
	cases := []struct {
		struttura, processi float64
		wantTier            int
		wantBasis           string
	}{
		// processi outranks struttura → basis = processi
		{struttura: 4.0, processi: 16.0, wantTier: 3, wantBasis: "processi_interessati"},
		// struttura outranks processi → basis = struttura
		{struttura: 12.0, processi: 6.0, wantTier: 3, wantBasis: "struttura_produttiva"},
		// tie at T2 → basis = struttura (more conservative)
		{struttura: 8.0, processi: 12.0, wantTier: 2, wantBasis: "struttura_produttiva"},
		// both ineligible
		{struttura: 1.0, processi: 4.0, wantTier: 0, wantBasis: "ineligible"},
	}
	for _, tc := range cases {
		factors := scenarioFactors(tc.struttura, tc.processi, 1_000_000, RegimeLB2025)
		r, _ := Pack.Build(context.Background(), defaultPeriod(), factors, newReadings(nil))
		body := r.Body.(Body)
		if body.EnergySavings.EffectiveTier != tc.wantTier {
			t.Errorf("effective tier (struttura %v %% / processi %v %%): want %d, got %d",
				tc.struttura, tc.processi, tc.wantTier, body.EnergySavings.EffectiveTier)
		}
		if body.EnergySavings.EffectiveBasis != tc.wantBasis {
			t.Errorf("effective basis (struttura %v %% / processi %v %%): want %q, got %q",
				tc.struttura, tc.processi, tc.wantBasis, body.EnergySavings.EffectiveBasis)
		}
	}
}

// TestBuildLB2025RateMatrix verifies the post-Legge-di-Bilancio-2025
// two-bracket rate computation. Investment €4M @ T2 → 4M × 0.40 = €1.6M.
func TestBuildLB2025RateMatrix(t *testing.T) {
	factors := scenarioFactors(7.0, 0, 4_000_000, RegimeLB2025) // 7% struttura → T2
	r, _ := Pack.Build(context.Background(), defaultPeriod(), factors, newReadings(nil))
	body := r.Body.(Body)

	if body.RegimeVersion != RegimeLB2025 {
		t.Errorf("regime: want %q, got %q", RegimeLB2025, body.RegimeVersion)
	}
	if body.TaxCredit.AppliedTier != 2 {
		t.Errorf("applied tier: want 2, got %d", body.TaxCredit.AppliedTier)
	}
	if len(body.TaxCredit.PerBracketEUR) != 2 {
		t.Fatalf("LB 2025 should have 2 brackets, got %d", len(body.TaxCredit.PerBracketEUR))
	}
	wantTotal := 4_000_000.0 * 0.40
	if !floatNear(body.TaxCredit.TotalCreditEUR, wantTotal, 1e-6) {
		t.Errorf("total credit @ T2: want %.2f, got %.2f", wantTotal, body.TaxCredit.TotalCreditEUR)
	}
	// First bracket: 4M investment, 0.40 rate → 1.6M credit.
	b0 := body.TaxCredit.PerBracketEUR[0]
	if !floatNear(b0.InvestmentInBracketEUR, 4_000_000, 1e-6) {
		t.Errorf("bracket 0 investment: want 4M, got %v", b0.InvestmentInBracketEUR)
	}
	if !floatNear(b0.CreditEUR, 1_600_000, 1e-6) {
		t.Errorf("bracket 0 credit: want 1.6M, got %v", b0.CreditEUR)
	}
	// Second bracket: 0 investment, 0 credit.
	b1 := body.TaxCredit.PerBracketEUR[1]
	if b1.InvestmentInBracketEUR != 0 || b1.CreditEUR != 0 {
		t.Errorf("bracket 1: want 0/0, got %v/%v", b1.InvestmentInBracketEUR, b1.CreditEUR)
	}
}

// TestBuildDM240724RateMatrix verifies the original three-bracket regime.
// Investment €15M @ T3 (12% struttura) → 2.5M × 0.45 + 7.5M × 0.25 + 5M × 0.15
//
//	= 1.125M + 1.875M + 0.75M = €3.75M.
func TestBuildDM240724RateMatrix(t *testing.T) {
	factors := scenarioFactors(12.0, 0, 15_000_000, RegimeDM240724)
	r, _ := Pack.Build(context.Background(), defaultPeriod(), factors, newReadings(nil))
	body := r.Body.(Body)

	if body.RegimeVersion != RegimeDM240724 {
		t.Errorf("regime: want %q, got %q", RegimeDM240724, body.RegimeVersion)
	}
	if body.TaxCredit.AppliedTier != 3 {
		t.Errorf("applied tier: want 3, got %d", body.TaxCredit.AppliedTier)
	}
	if len(body.TaxCredit.PerBracketEUR) != 3 {
		t.Fatalf("DM 24/07/2024 should have 3 brackets, got %d", len(body.TaxCredit.PerBracketEUR))
	}
	wantTotal := 2_500_000.0*0.45 + 7_500_000.0*0.25 + 5_000_000.0*0.15
	if !floatNear(body.TaxCredit.TotalCreditEUR, wantTotal, 1e-6) {
		t.Errorf("total credit @ T3 / DM regime: want %.2f, got %.2f",
			wantTotal, body.TaxCredit.TotalCreditEUR)
	}
}

// TestBuildAnnualCapApplied verifies investment > €50M is capped, excess surfaced.
func TestBuildAnnualCapApplied(t *testing.T) {
	factors := scenarioFactors(8.0, 0, 60_000_000, RegimeLB2025) // T2
	r, _ := Pack.Build(context.Background(), defaultPeriod(), factors, newReadings(nil))
	body := r.Body.(Body)

	if !floatNear(body.Investment.AboveCapExcessEUR, 10_000_000, 1e-6) {
		t.Errorf("above cap excess: want 10M, got %v", body.Investment.AboveCapExcessEUR)
	}
	// At T2, capped to €50M:
	//   bracket 1 (≤10M): 10M × 0.40 = 4M
	//   bracket 2 (10-50M): 40M × 0.10 = 4M
	//   total = 8M
	wantTotal := 10_000_000.0*0.40 + 40_000_000.0*0.10
	if !floatNear(body.TaxCredit.TotalCreditEUR, wantTotal, 1e-6) {
		t.Errorf("capped credit: want %.2f, got %.2f", wantTotal, body.TaxCredit.TotalCreditEUR)
	}
	// Note about the cap should be present.
	if !anyNoteContains(body.Notes, "exceeds annual cap") {
		t.Error("expected note about exceeding annual cap")
	}
}

// TestBuildIneligibleBelowMinimum — both savings under thresholds.
func TestBuildIneligibleBelowMinimum(t *testing.T) {
	factors := scenarioFactors(2.0, 4.0, 1_000_000, RegimeLB2025)
	r, _ := Pack.Build(context.Background(), defaultPeriod(), factors, newReadings(nil))
	body := r.Body.(Body)

	if body.EnergySavings.EffectiveTier != 0 {
		t.Errorf("effective tier: want 0, got %d", body.EnergySavings.EffectiveTier)
	}
	if body.TaxCredit.TotalCreditEUR != 0 {
		t.Errorf("ineligible credit: want 0, got %v", body.TaxCredit.TotalCreditEUR)
	}
	if !anyNoteContains(body.Notes, "ineligible") {
		t.Error("expected ineligibility note")
	}
}

// TestBuildHonoursContext.
func TestBuildHonoursContext(t *testing.T) {
	period := defaultPeriod()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Pack.Build(ctx, period, scenarioFactors(10, 15, 4_000_000, RegimeLB2025), newReadings(nil)); err == nil {
		t.Error("expected ctx.Err() on cancelled context")
	}
}

// TestEncodedIsValidJSON — output parses + has expected top-level keys.
func TestEncodedIsValidJSON(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		scenarioFactors(10, 15, 4_000_000, RegimeLB2025), newReadings(nil))

	var decoded map[string]any
	if err := json.Unmarshal(r.Encoded, &decoded); err != nil {
		t.Fatalf("Encoded is not valid JSON: %v", err)
	}
	for _, key := range []string{
		"report", "regulator", "regime_version", "period", "factors_used",
		"investment", "energy_savings", "tax_credit",
		"narrative_data_points", "ege_certification_required",
	} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("Encoded JSON missing key %q", key)
		}
	}
	if decoded["report"] != "piano_5_0" {
		t.Errorf("report identifier: want piano_5_0, got %v", decoded["report"])
	}
	if decoded["ege_certification_required"] != true {
		t.Errorf("ege_certification_required: want true, got %v", decoded["ege_certification_required"])
	}
}

// TestBuildMissingFactorsEmitsNotes — graceful degradation.
func TestBuildMissingFactorsEmitsNotes(t *testing.T) {
	factors := newFactorBundle(map[string]factorRow{}) // entirely empty
	r, _ := Pack.Build(context.Background(), defaultPeriod(), factors, newReadings(nil))
	body := r.Body.(Body)

	if body.EnergySavings.EffectiveTier != 0 {
		t.Errorf("with no scenario inputs, effective tier should be 0")
	}
	if body.TaxCredit.TotalCreditEUR != 0 {
		t.Errorf("with no investment, total credit should be 0")
	}
	if !anyNoteContains(body.Notes, "Struttura produttiva scenario inputs missing") {
		t.Error("expected struttura missing-input note")
	}
	if !anyNoteContains(body.Notes, "Investment total missing") {
		t.Error("expected investment-missing note")
	}
}

// TestBuildNarrativeIsNullPlaceholder verifies the narrative block is nullable.
func TestBuildNarrativeIsNullPlaceholder(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		scenarioFactors(10, 15, 4_000_000, RegimeLB2025), newReadings(nil))
	body := r.Body.(Body)

	if body.NarrativeDataPoints.ExAnteCertID != nil {
		t.Error("ExAnteCertID should be null placeholder")
	}
	if body.NarrativeDataPoints.InvestmentDescription != nil {
		t.Error("InvestmentDescription should be null placeholder")
	}
	if body.NarrativeDataPoints.Note == "" {
		t.Error("NarrativeDataPoints.Note should be populated")
	}
}

// ── Test helpers ───────────────────────────────────────────────────────

type factorRow struct {
	value   float64
	version string
}

type stubFactorBundle struct{ rows map[string]factorRow }

func newFactorBundle(rows map[string]factorRow) reporting.FactorBundle {
	return &stubFactorBundle{rows: rows}
}
func (b *stubFactorBundle) Get(key string) (float64, string, bool) {
	r, ok := b.rows[key]
	if !ok {
		return 0, "", false
	}
	return r.value, r.version, true
}
func (b *stubFactorBundle) Versions() map[string]string {
	out := map[string]string{}
	for k, v := range b.rows {
		out[k] = v.version
	}
	return out
}

type stubReadings struct{ rows []reporting.AggregatedRow }

func newReadings(rows []reporting.AggregatedRow) reporting.AggregatedReadings {
	return &stubReadings{rows: rows}
}
func (r *stubReadings) Iter() reporting.AggregatedIter {
	return &stubIter{rows: r.rows, idx: -1}
}

type stubIter struct {
	rows []reporting.AggregatedRow
	idx  int
}

func (i *stubIter) Next() bool { i.idx++; return i.idx < len(i.rows) }
func (i *stubIter) Row() reporting.AggregatedRow {
	if i.idx < 0 || i.idx >= len(i.rows) {
		return reporting.AggregatedRow{}
	}
	return i.rows[i.idx]
}
func (i *stubIter) Err() error { return nil }

func floatNear(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

func anyNoteContains(notes []string, sub string) bool {
	for _, n := range notes {
		if containsSub(n, sub) {
			return true
		}
	}
	return false
}

func containsSub(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
