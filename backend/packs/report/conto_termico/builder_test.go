package conto_termico

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
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

func ct20Factors(incentive, eligibleCosts float64) reporting.FactorBundle {
	return newFactorBundle(map[string]factorRow{
		FactorInterventionCategory: {value: 12, version: "engagement"},
		FactorBeneficiaryType:      {value: BeneficiaryPA, version: "engagement"},
		FactorIncentiveAmountEUR:   {value: incentive, version: "engagement"},
		FactorEligibleCostsEUR:     {value: eligibleCosts, version: "engagement"},
		FactorClimateZone:          {value: 5, version: "engagement"}, // E
	})
}

func ct30Factors(incentive, eligibleCosts float64) reporting.FactorBundle {
	return newFactorBundle(map[string]factorRow{
		FactorRegimeVersion:        {value: 1, version: "engagement"}, // CT 3.0
		FactorInterventionCategory: {value: 12, version: "engagement"},
		FactorBeneficiaryType:      {value: BeneficiaryCER, version: "engagement"},
		FactorIncentiveAmountEUR:   {value: incentive, version: "engagement"},
		FactorEligibleCostsEUR:     {value: eligibleCosts, version: "engagement"},
		FactorClimateZone:          {value: 5, version: "engagement"}, // E
	})
}

// TestBuildBitPerfectReproducibility — Rule 89.
func TestBuildBitPerfectReproducibility(t *testing.T) {
	period := defaultPeriod()
	factors := ct20Factors(4500, 9000)
	r1, _ := Pack.Build(context.Background(), period, factors, newReadings(nil))
	r2, _ := Pack.Build(context.Background(), period, factors, newReadings(nil))
	if !bytes.Equal(r1.Encoded, r2.Encoded) {
		t.Fatalf("Encoded bytes differ between two Build calls (Rule 89 violated)")
	}
}

// TestBuildCT20SingleTrancheBelowThreshold — incentive ≤ €5 000 → 1-shot.
func TestBuildCT20SingleTrancheBelowThreshold(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		ct20Factors(4500, 9000), newReadings(nil))
	body := r.Body.(Body)

	if body.RegimeVersion != RegimeCT20 {
		t.Errorf("regime: want %q, got %q", RegimeCT20, body.RegimeVersion)
	}
	if body.PaymentSchedule.PaymentMode != PaymentSingleTranche {
		t.Errorf("payment mode: want %q, got %q", PaymentSingleTranche, body.PaymentSchedule.PaymentMode)
	}
	if body.PaymentSchedule.PaymentYears != 1 {
		t.Errorf("payment years: want 1, got %d", body.PaymentSchedule.PaymentYears)
	}
	if !floatNear(body.PaymentSchedule.AnnualRateEUR, 4500, 1e-9) {
		t.Errorf("annual rate: want 4500, got %v", body.PaymentSchedule.AnnualRateEUR)
	}
	if body.PaymentSchedule.SubmissionWindowDays != 60 {
		t.Errorf("submission window: want 60, got %d", body.PaymentSchedule.SubmissionWindowDays)
	}
}

// TestBuildCT20AnnualRatesAboveThreshold — incentive > €5 000 → 2 default years.
func TestBuildCT20AnnualRatesAboveThreshold(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		ct20Factors(20_000, 40_000), newReadings(nil))
	body := r.Body.(Body)

	if body.PaymentSchedule.PaymentMode != PaymentAnnualRates {
		t.Errorf("payment mode: want %q, got %q", PaymentAnnualRates, body.PaymentSchedule.PaymentMode)
	}
	if body.PaymentSchedule.PaymentYears != 2 {
		t.Errorf("payment years (default): want 2, got %d", body.PaymentSchedule.PaymentYears)
	}
	wantAnnual := 20_000.0 / 2.0
	if !floatNear(body.PaymentSchedule.AnnualRateEUR, wantAnnual, 1e-9) {
		t.Errorf("annual rate: want %v, got %v", wantAnnual, body.PaymentSchedule.AnnualRateEUR)
	}
}

// TestBuildCT20BoundaryExactlyThreshold — incentive == €5 000 → still single-tranche.
func TestBuildCT20BoundaryExactlyThreshold(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		ct20Factors(5000, 10_000), newReadings(nil))
	body := r.Body.(Body)
	if body.PaymentSchedule.PaymentMode != PaymentSingleTranche {
		t.Errorf("incentive at exactly €5 000 should remain single_tranche; got %q",
			body.PaymentSchedule.PaymentMode)
	}
}

// TestBuildCT20PaymentYearsOverride.
func TestBuildCT20PaymentYearsOverride(t *testing.T) {
	rows := map[string]factorRow{
		FactorIncentiveAmountEUR:   {value: 25_000, version: "engagement"},
		FactorPaymentYearsOverride: {value: 5, version: "engagement"},
	}
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(rows), newReadings(nil))
	body := r.Body.(Body)

	if body.PaymentSchedule.PaymentYears != 5 {
		t.Errorf("payment years (override): want 5, got %d", body.PaymentSchedule.PaymentYears)
	}
	if !floatNear(body.PaymentSchedule.AnnualRateEUR, 5000, 1e-9) {
		t.Errorf("annual rate over 5 years: want 5000, got %v", body.PaymentSchedule.AnnualRateEUR)
	}
}

// TestBuildCT20PaymentYearsOverrideOutOfRange — surface note.
func TestBuildCT20PaymentYearsOverrideOutOfRange(t *testing.T) {
	rows := map[string]factorRow{
		FactorIncentiveAmountEUR:   {value: 25_000, version: "engagement"},
		FactorPaymentYearsOverride: {value: 10, version: "engagement"}, // out of [2,5]
	}
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(rows), newReadings(nil))
	body := r.Body.(Body)
	if !anyNoteContains(body.Notes, "outside the DM 16/02/2016") {
		t.Error("expected out-of-range note for payment_years override")
	}
}

// TestBuildCT30CapitalGrantMode.
func TestBuildCT30CapitalGrantMode(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		ct30Factors(50_000, 100_000), newReadings(nil)) // 50k incentive of 100k costs = 50% (under 65% cap)
	body := r.Body.(Body)

	if body.RegimeVersion != RegimeCT30 {
		t.Errorf("regime: want %q, got %q", RegimeCT30, body.RegimeVersion)
	}
	if body.PaymentSchedule.PaymentMode != PaymentCapitalGrant {
		t.Errorf("payment mode: want %q, got %q", PaymentCapitalGrant, body.PaymentSchedule.PaymentMode)
	}
	if body.PaymentSchedule.SubmissionWindowDays != 90 {
		t.Errorf("submission window: want 90 (CT 3.0), got %d", body.PaymentSchedule.SubmissionWindowDays)
	}
	if body.EligibleCosts.MaxCapPct == nil || *body.EligibleCosts.MaxCapPct != 0.65 {
		t.Errorf("max cap pct: want 0.65, got %v", body.EligibleCosts.MaxCapPct)
	}
	if body.EligibleCosts.MaxCapEUR == nil || !floatNear(*body.EligibleCosts.MaxCapEUR, 65_000, 1e-9) {
		t.Errorf("max cap EUR: want 65000, got %v", body.EligibleCosts.MaxCapEUR)
	}
	if body.EligibleCosts.CapViolation {
		t.Error("cap violation should be false (50k under 65k cap)")
	}
}

// TestBuildCT30CapViolation — incentive > 65% cap.
func TestBuildCT30CapViolation(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		ct30Factors(80_000, 100_000), newReadings(nil)) // 80% requested > 65% cap
	body := r.Body.(Body)

	if !body.EligibleCosts.CapViolation {
		t.Error("cap violation should be true (80k > 65k cap)")
	}
	if !floatNear(body.EligibleCosts.CapExcessEUR, 15_000, 1e-9) {
		t.Errorf("cap excess: want 15000, got %v", body.EligibleCosts.CapExcessEUR)
	}
	if !anyNoteContains(body.Notes, "exceeds the CT 3.0 65%") {
		t.Error("expected cap-violation note")
	}
}

// TestBuildBeneficiaryTypeMapping verifies all 4 beneficiary selectors.
func TestBuildBeneficiaryTypeMapping(t *testing.T) {
	cases := []struct {
		selector float64
		want     string
	}{
		{selector: BeneficiaryPA, want: "PA"},
		{selector: BeneficiaryPrivato, want: "privato"},
		{selector: BeneficiaryETS, want: "ETS"},
		{selector: BeneficiaryCER, want: "CER"},
	}
	for _, tc := range cases {
		rows := map[string]factorRow{
			FactorBeneficiaryType:    {value: tc.selector, version: "engagement"},
			FactorIncentiveAmountEUR: {value: 4500, version: "engagement"},
		}
		r, _ := Pack.Build(context.Background(), defaultPeriod(),
			newFactorBundle(rows), newReadings(nil))
		body := r.Body.(Body)
		if body.Intervention.Beneficiary != tc.want {
			t.Errorf("beneficiary @ selector %v: want %q, got %q",
				tc.selector, tc.want, body.Intervention.Beneficiary)
		}
	}
}

// TestBuildClimateZoneMapping.
func TestBuildClimateZoneMapping(t *testing.T) {
	for i, want := range []string{"A", "B", "C", "D", "E", "F"} {
		rows := map[string]factorRow{
			FactorClimateZone:        {value: float64(i + 1), version: "engagement"},
			FactorIncentiveAmountEUR: {value: 4500, version: "engagement"},
		}
		r, _ := Pack.Build(context.Background(), defaultPeriod(),
			newFactorBundle(rows), newReadings(nil))
		body := r.Body.(Body)
		if body.Intervention.ClimateZone == nil || *body.Intervention.ClimateZone != want {
			t.Errorf("climate zone @ selector %d: want %q, got %v",
				i+1, want, body.Intervention.ClimateZone)
		}
	}
}

// TestBuildMissingIncentiveEmitsNote.
func TestBuildMissingIncentiveEmitsNote(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(nil), newReadings(nil))
	body := r.Body.(Body)
	if !anyNoteContains(body.Notes, "Incentive amount missing") {
		t.Error("expected missing-incentive note")
	}
	if body.PaymentSchedule.IncentiveAmountEUR != 0 {
		t.Errorf("incentive amount: want 0, got %v", body.PaymentSchedule.IncentiveAmountEUR)
	}
}

// TestBuildHonoursContext.
func TestBuildHonoursContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Pack.Build(ctx, defaultPeriod(), newFactorBundle(nil), newReadings(nil)); err == nil {
		t.Error("expected ctx.Err() on cancelled context")
	}
}

// TestEncodedIsValidJSON.
func TestEncodedIsValidJSON(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		ct20Factors(4500, 9000), newReadings(nil))
	var decoded map[string]any
	if err := json.Unmarshal(r.Encoded, &decoded); err != nil {
		t.Fatalf("Encoded is not valid JSON: %v", err)
	}
	for _, key := range []string{
		"report", "regulator", "regime_version", "period",
		"factors_used", "intervention", "eligible_costs", "payment_schedule",
		"narrative_data_points", "ege_certification_required",
	} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("Encoded JSON missing key %q", key)
		}
	}
	if decoded["report"] != "conto_termico" {
		t.Errorf("report identifier: want conto_termico, got %v", decoded["report"])
	}
}

// TestBuildNarrativeIsNullPlaceholder.
func TestBuildNarrativeIsNullPlaceholder(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		ct20Factors(4500, 9000), newReadings(nil))
	body := r.Body.(Body)
	if body.NarrativeDataPoints.InterventionDescription != nil {
		t.Error("InterventionDescription should be null placeholder")
	}
	if body.NarrativeDataPoints.CertifierID != nil {
		t.Error("CertifierID should be null placeholder")
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
	if rows == nil {
		rows = map[string]factorRow{}
	}
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
		if strings.Contains(n, sub) {
			return true
		}
	}
	return false
}
