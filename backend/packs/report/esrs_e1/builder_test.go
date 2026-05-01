package esrs_e1

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

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

func allFactors() reporting.FactorBundle {
	return newFactorBundle(map[string]factorRow{
		FactorScope2Location:        {value: 233.0, version: "ispra@2025.04"},
		FactorScope2Market:          {value: 332.0, version: "gse-aib@2025"},
		FactorRenewableShare:        {value: 44.5, version: "gse-fer@2026.04"},
		"natural_gas_combustion":    {value: 1.967, version: "dm-2017-01-11"},
		"diesel_combustion":         {value: 2.642, version: "dm-2017-01-11"},
	})
}

// TestBuildBitPerfectReproducibility — Rule 89.
func TestBuildBitPerfectReproducibility(t *testing.T) {
	period := defaultPeriod()
	factors := allFactors()
	r1, _ := Pack.Build(context.Background(), period, factors, newReadings(sampleRows()))
	r2, _ := Pack.Build(context.Background(), period, factors, newReadings(sampleRows()))
	if !bytes.Equal(r1.Encoded, r2.Encoded) {
		t.Fatalf("Encoded bytes differ between two Build calls (Rule 89 violated)")
	}
}

// TestBuildE15EnergyConsumption verifies E1-5 quantitative fields.
func TestBuildE15EnergyConsumption(t *testing.T) {
	period := defaultPeriod()
	factors := allFactors()
	report, err := Pack.Build(context.Background(), period, factors, newReadings(sampleRows()))
	if err != nil {
		t.Fatal(err)
	}
	body := report.Body.(Body)

	// 100_000 Wh = 100 kWh = 0.1 MWh.
	wantMWh := 100.0 / 1000.0
	if !floatNear(body.E1_5_EnergyConsMix.ElectricityMWh, wantMWh, 1e-9) {
		t.Errorf("ElectricityMWh: want %v, got %v", wantMWh, body.E1_5_EnergyConsMix.ElectricityMWh)
	}
	if body.E1_5_EnergyConsMix.RenewableSharePct == nil {
		t.Fatal("RenewableSharePct should be populated")
	}
	if *body.E1_5_EnergyConsMix.RenewableSharePct != 44.5 {
		t.Errorf("RenewableSharePct: want 44.5, got %v", *body.E1_5_EnergyConsMix.RenewableSharePct)
	}
	if body.E1_5_EnergyConsMix.NonRenewableSharePct == nil {
		t.Fatal("NonRenewableSharePct should be populated")
	}
	if *body.E1_5_EnergyConsMix.NonRenewableSharePct != 55.5 {
		t.Errorf("NonRenewableSharePct: want 55.5, got %v", *body.E1_5_EnergyConsMix.NonRenewableSharePct)
	}
	if len(body.E1_5_EnergyConsMix.NonElectricitySources) != 2 {
		t.Errorf("NonElectricitySources: want 2 entries, got %d", len(body.E1_5_EnergyConsMix.NonElectricitySources))
	}
}

// TestBuildE16Emissions verifies E1-6 Scope 1 + 2 + 3 + totals.
func TestBuildE16Emissions(t *testing.T) {
	period := defaultPeriod()
	factors := allFactors()
	report, err := Pack.Build(context.Background(), period, factors, newReadings(sampleRows()))
	if err != nil {
		t.Fatal(err)
	}
	body := report.Body.(Body)
	e16 := body.E1_6_GHGEmissions

	// Scope 1: 100 Sm3 × 1.967 + 50 l × 2.642 = 196.7 + 132.1 = 328.8 kg.
	wantS1 := 100.0*1.967 + 50.0*2.642
	if !floatNear(e16.Scope1.KgCO2eqTotal, wantS1, 1e-9) {
		t.Errorf("Scope1.Total: want %v, got %v", wantS1, e16.Scope1.KgCO2eqTotal)
	}

	// Scope 2 location: 100 kWh × 233 / 1000 = 23.3 kg.
	if e16.Scope2.KgCO2eqLocationBased == nil {
		t.Fatal("Scope2 location should be populated")
	}
	if !floatNear(*e16.Scope2.KgCO2eqLocationBased, 100.0*233.0/1000.0, 1e-9) {
		t.Errorf("Scope2.location: want %v, got %v", 100.0*233.0/1000.0, *e16.Scope2.KgCO2eqLocationBased)
	}

	// Scope 2 market: 100 kWh × 332 / 1000 = 33.2 kg.
	if !floatNear(*e16.Scope2.KgCO2eqMarketBased, 100.0*332.0/1000.0, 1e-9) {
		t.Errorf("Scope2.market: want %v, got %v", 100.0*332.0/1000.0, *e16.Scope2.KgCO2eqMarketBased)
	}

	// Scope 3 zero placeholder.
	if e16.Scope3.KgCO2eqTotal != 0 {
		t.Errorf("Scope3 should be 0 placeholder")
	}
	if e16.Scope3.Note == "" {
		t.Error("Scope3.Note should be populated")
	}

	// Total location-based.
	wantTotalLoc := wantS1 + 100.0*233.0/1000.0
	if e16.Totals.KgCO2eqLocationBased == nil {
		t.Fatal("Totals.location should be populated")
	}
	if !floatNear(*e16.Totals.KgCO2eqLocationBased, wantTotalLoc, 1e-9) {
		t.Errorf("Totals.location: want %v, got %v", wantTotalLoc, *e16.Totals.KgCO2eqLocationBased)
	}
}

// TestBuildNarrativeIsNullPlaceholder verifies the narrative block is the
// expected null-with-note shape; the engagement-fork's orchestrator
// injects content later.
func TestBuildNarrativeIsNullPlaceholder(t *testing.T) {
	period := defaultPeriod()
	report, _ := Pack.Build(context.Background(), period, allFactors(), newReadings(sampleRows()))
	body := report.Body.(Body)

	if body.NarrativeDataPoints.E11 != nil {
		t.Error("E1-1 should be null placeholder")
	}
	if body.NarrativeDataPoints.E14 != nil {
		t.Error("E1-4 should be null placeholder")
	}
	if body.NarrativeDataPoints.Note == "" {
		t.Error("NarrativeDataPoints.Note should be populated")
	}
}

// TestBuildMissingRenewableShareEmitsNote — graceful degradation.
func TestBuildMissingRenewableShareEmitsNote(t *testing.T) {
	period := defaultPeriod()
	factors := newFactorBundle(map[string]factorRow{
		FactorScope2Location: {value: 233.0, version: "ispra@2025.04"},
		// renewable share intentionally absent
	})
	report, _ := Pack.Build(context.Background(), period, factors, newReadings(electricityOnly()))
	body := report.Body.(Body)

	if body.E1_5_EnergyConsMix.RenewableSharePct != nil {
		t.Error("RenewableSharePct should be nil when factor absent")
	}
	gotNote := false
	for _, n := range body.E1_5_EnergyConsMix.Notes {
		if len(n) > 0 && contains(n, "renewable share omitted") {
			gotNote = true
		}
	}
	if !gotNote {
		t.Error("notes should mention missing renewable share")
	}
}

// TestBuildHonoursContext.
func TestBuildHonoursContext(t *testing.T) {
	period := defaultPeriod()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Pack.Build(ctx, period, allFactors(), newReadings(nil)); err == nil {
		t.Error("expected ctx.Err() on cancelled context")
	}
}

// TestEncodedIsValidJSON — output parses + has expected top-level keys.
func TestEncodedIsValidJSON(t *testing.T) {
	period := defaultPeriod()
	report, _ := Pack.Build(context.Background(), period, allFactors(), newReadings(sampleRows()))
	var decoded map[string]any
	if err := json.Unmarshal(report.Encoded, &decoded); err != nil {
		t.Fatalf("Encoded is not valid JSON: %v", err)
	}
	for _, key := range []string{
		"report", "regulator", "period", "factors_used",
		"e1_5_energy_consumption", "e1_6_ghg_emissions",
		"narrative_data_points", "unclassified_rows",
	} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("Encoded JSON missing key %q", key)
		}
	}
	if decoded["report"] != "esrs_e1" {
		t.Errorf("report identifier: want esrs_e1, got %v", decoded["report"])
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

func sampleRows() []reporting.AggregatedRow {
	out := append([]reporting.AggregatedRow{}, electricityOnly()...)
	out = append(out, scope1Mixed()...)
	return out
}

func electricityOnly() []reporting.AggregatedRow {
	meterA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	chanA := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	t1 := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	return []reporting.AggregatedRow{
		{MeterID: meterA, ChannelID: chanA, BucketStart: t1, Sum: 60_000, Count: 1, Unit: "Wh"},
		{MeterID: meterA, ChannelID: chanA, BucketStart: t1.Add(time.Hour), Sum: 40_000, Count: 1, Unit: "Wh"},
	}
}

func scope1Mixed() []reporting.AggregatedRow {
	meterB := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	chanB := uuid.MustParse("00000000-0000-0000-0000-0000000000b1")
	meterC := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	chanC := uuid.MustParse("00000000-0000-0000-0000-0000000000c1")
	t1 := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	return []reporting.AggregatedRow{
		{MeterID: meterB, ChannelID: chanB, BucketStart: t1, Sum: 100, Count: 1, Unit: "Sm3"},
		{MeterID: meterC, ChannelID: chanC, BucketStart: t1, Sum: 50, Count: 1, Unit: "l_diesel"},
	}
}

func floatNear(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
