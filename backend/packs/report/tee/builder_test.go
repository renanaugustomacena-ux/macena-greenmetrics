package tee

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

func projectFactors(method, exAnte, exPost, vitaUtile, currentYear float64) reporting.FactorBundle {
	return newFactorBundle(map[string]factorRow{
		FactorMethod:               {value: method, version: "engagement-supplied"},
		FactorExAnteTep:            {value: exAnte, version: "engagement-supplied"},
		FactorExPostTep:            {value: exPost, version: "engagement-supplied"},
		FactorVitaUtileYears:       {value: vitaUtile, version: "engagement-supplied"},
		FactorCurrentYearInProject: {value: currentYear, version: "engagement-supplied"},
	})
}

// TestBuildBitPerfectReproducibility — Rule 89.
func TestBuildBitPerfectReproducibility(t *testing.T) {
	period := defaultPeriod()
	factors := projectFactors(MethodConsuntivo, 250, 180, 10, 3)
	r1, _ := Pack.Build(context.Background(), period, factors, newReadings(nil))
	r2, _ := Pack.Build(context.Background(), period, factors, newReadings(nil))
	if !bytes.Equal(r1.Encoded, r2.Encoded) {
		t.Fatalf("Encoded bytes differ between two Build calls (Rule 89 violated)")
	}
}

// TestBuildAnnualSavingAndPct verifies basic delta math.
func TestBuildAnnualSavingAndPct(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		projectFactors(MethodConsuntivo, 250, 180, 10, 3), newReadings(nil))
	body := r.Body.(Body)

	if !floatNear(body.EnergySavings.AnnualSavingTep, 70.0, 1e-9) {
		t.Errorf("annual saving: want 70.0, got %v", body.EnergySavings.AnnualSavingTep)
	}
	if !floatNear(body.EnergySavings.SavingPct, 28.0, 1e-9) {
		t.Errorf("saving pct: want 28.0, got %v", body.EnergySavings.SavingPct)
	}
}

// TestBuildKFactorFirstHalf — currentYear ≤ vita_utile/2 → K1 = 1.2.
func TestBuildKFactorFirstHalf(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		projectFactors(MethodConsuntivo, 100, 50, 10, 5), newReadings(nil))
	body := r.Body.(Body)

	if body.Project.CurrentPeriodHalf != "first" {
		t.Errorf("period half: want first, got %q", body.Project.CurrentPeriodHalf)
	}
	if body.TEECalculation.KFactor != KFactorFirstHalf {
		t.Errorf("K factor: want %v, got %v", KFactorFirstHalf, body.TEECalculation.KFactor)
	}
	wantTEE := 50.0 * KFactorFirstHalf // 50 tep × 1.2 = 60 TEE
	if !floatNear(body.TEECalculation.TEEIssued, wantTEE, 1e-9) {
		t.Errorf("TEE issued: want %v, got %v", wantTEE, body.TEECalculation.TEEIssued)
	}
}

// TestBuildKFactorSecondHalf — currentYear > vita_utile/2 → K2 = 0.8.
func TestBuildKFactorSecondHalf(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		projectFactors(MethodConsuntivo, 100, 50, 10, 7), newReadings(nil))
	body := r.Body.(Body)

	if body.Project.CurrentPeriodHalf != "second" {
		t.Errorf("period half: want second, got %q", body.Project.CurrentPeriodHalf)
	}
	if body.TEECalculation.KFactor != KFactorSecondHalf {
		t.Errorf("K factor: want %v, got %v", KFactorSecondHalf, body.TEECalculation.KFactor)
	}
	wantTEE := 50.0 * KFactorSecondHalf // 50 tep × 0.8 = 40 TEE
	if !floatNear(body.TEECalculation.TEEIssued, wantTEE, 1e-9) {
		t.Errorf("TEE issued: want %v, got %v", wantTEE, body.TEECalculation.TEEIssued)
	}
}

// TestBuildKFactorBoundaryExactlyHalf — currentYear = vita_utile/2 → first half (≤).
func TestBuildKFactorBoundaryExactlyHalf(t *testing.T) {
	// vita_utile=10, currentYear=5 (== 10/2 = 5) → first half (boundary inclusive).
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		projectFactors(MethodConsuntivo, 100, 50, 10, 5), newReadings(nil))
	body := r.Body.(Body)

	if body.Project.CurrentPeriodHalf != "first" {
		t.Errorf("boundary at exactly half should be first half, got %q", body.Project.CurrentPeriodHalf)
	}
}

// TestBuildIneligibleNonPositiveSaving.
func TestBuildIneligibleNonPositiveSaving(t *testing.T) {
	// Ex-post > ex-ante → saving negative.
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		projectFactors(MethodConsuntivo, 100, 110, 10, 3), newReadings(nil))
	body := r.Body.(Body)

	if body.TEECalculation.TEEIssued != 0 {
		t.Errorf("TEE issued for non-positive saving: want 0, got %v", body.TEECalculation.TEEIssued)
	}
	if !anyNoteContains(body.Notes, "ineligible") {
		t.Error("expected ineligibility note for non-positive saving")
	}
}

// TestBuildMethodSelectors verifies all three method selectors map correctly.
func TestBuildMethodSelectors(t *testing.T) {
	cases := []struct {
		selector float64
		want     string
	}{
		{selector: MethodConsuntivo, want: "consuntivo"},
		{selector: MethodStandardizzato, want: "standardizzato"},
		{selector: MethodPPPM, want: "pppm"},
	}
	for _, tc := range cases {
		r, _ := Pack.Build(context.Background(), defaultPeriod(),
			projectFactors(tc.selector, 100, 50, 10, 3), newReadings(nil))
		body := r.Body.(Body)
		if body.Project.Method != tc.want {
			t.Errorf("method selector %v: want %q, got %q", tc.selector, tc.want, body.Project.Method)
		}
	}
}

// TestBuildRegimeVersionDM2017Default.
func TestBuildRegimeVersionDM2017Default(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		projectFactors(MethodConsuntivo, 100, 50, 10, 3), newReadings(nil))
	body := r.Body.(Body)
	if body.RegimeVersion != RegimeDM2017 {
		t.Errorf("default regime: want %q, got %q", RegimeDM2017, body.RegimeVersion)
	}
}

// TestBuildRegimeVersionDMMase202530Selector — selector value 1 → mase regime.
func TestBuildRegimeVersionDMMase202530Selector(t *testing.T) {
	rows := map[string]factorRow{
		FactorMethod:               {value: MethodConsuntivo, version: "engagement"},
		FactorExAnteTep:            {value: 100, version: "engagement"},
		FactorExPostTep:            {value: 50, version: "engagement"},
		FactorVitaUtileYears:       {value: 10, version: "engagement"},
		FactorCurrentYearInProject: {value: 3, version: "engagement"},
		FactorRegimeVersion:        {value: 1, version: "engagement"},
	}
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(rows), newReadings(nil))
	body := r.Body.(Body)
	if body.RegimeVersion != RegimeDMMase202530 {
		t.Errorf("regime: want %q, got %q", RegimeDMMase202530, body.RegimeVersion)
	}
	if !anyNoteContains(body.Notes, "DM MASE 21/07/2025") {
		t.Error("expected note about DM MASE 21/07/2025 fallback")
	}
}

// TestBuildMissingScenarioInputs — graceful degradation.
func TestBuildMissingScenarioInputs(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(map[string]factorRow{}), newReadings(nil))
	body := r.Body.(Body)

	if body.TEECalculation.TEEIssued != 0 {
		t.Errorf("TEE issued: want 0, got %v", body.TEECalculation.TEEIssued)
	}
	if body.Project.CurrentPeriodHalf != "undefined" {
		t.Errorf("period half: want undefined, got %q", body.Project.CurrentPeriodHalf)
	}
	if !anyNoteContains(body.Notes, "Ex-ante / ex-post tep inputs missing") {
		t.Error("expected ex-ante/ex-post missing note")
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
		projectFactors(MethodConsuntivo, 250, 180, 10, 3), newReadings(nil))
	var decoded map[string]any
	if err := json.Unmarshal(r.Encoded, &decoded); err != nil {
		t.Fatalf("Encoded is not valid JSON: %v", err)
	}
	for _, key := range []string{
		"report", "regulator", "regime_version", "period",
		"factors_used", "project", "energy_savings", "tee_calculation",
		"narrative_data_points", "ege_certification_required",
	} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("Encoded JSON missing key %q", key)
		}
	}
	if decoded["report"] != "tee" {
		t.Errorf("report identifier: want tee, got %v", decoded["report"])
	}
}

// TestBuildNarrativeIsNullPlaceholder.
func TestBuildNarrativeIsNullPlaceholder(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		projectFactors(MethodConsuntivo, 100, 50, 10, 3), newReadings(nil))
	body := r.Body.(Body)
	if body.NarrativeDataPoints.ProjectDescription != nil {
		t.Error("ProjectDescription should be null placeholder")
	}
	if body.NarrativeDataPoints.RVCSubmissionID != nil {
		t.Error("RVCSubmissionID should be null placeholder")
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
