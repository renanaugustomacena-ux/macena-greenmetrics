package co2_footprint

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
		StartInclusiveUTC: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndExclusiveUTC:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Timezone:          "Europe/Rome",
	}
}

// allFactors returns a FactorBundle populated with the typical Italian
// flagship factors for Scope 1 + Scope 2.
func allFactors() reporting.FactorBundle {
	return newFactorBundle(map[string]factorRow{
		FactorScope2Location:        {value: 233.0, version: "ispra@2025.04"},
		FactorScope2Market:          {value: 332.0, version: "gse-aib@2025"},
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

// TestBuildScope2DualMethod verifies both location-based and market-based
// emissions are computed when both factors are present.
func TestBuildScope2DualMethod(t *testing.T) {
	period := defaultPeriod()
	factors := allFactors()
	report, err := Pack.Build(context.Background(), period, factors, newReadings(electricityOnly()))
	if err != nil {
		t.Fatal(err)
	}
	body := report.Body.(Body)

	if body.Scope2.KWhTotal != 100.0 {
		t.Errorf("KWhTotal: want 100, got %v (sample = 100_000 Wh)", body.Scope2.KWhTotal)
	}
	if body.Scope2.KgCO2eqLocationBased == nil {
		t.Fatal("location-based Scope 2 should be populated")
	}
	want := 100.0 * 233.0 / 1000.0 // 23.3 kg
	if got := *body.Scope2.KgCO2eqLocationBased; got != want {
		t.Errorf("location-based: want %v, got %v", want, got)
	}
	if body.Scope2.KgCO2eqMarketBased == nil {
		t.Fatal("market-based Scope 2 should be populated")
	}
	want = 100.0 * 332.0 / 1000.0 // 33.2 kg
	if got := *body.Scope2.KgCO2eqMarketBased; got != want {
		t.Errorf("market-based: want %v, got %v", want, got)
	}
}

// TestBuildScope1PerSource verifies Scope 1 per-source aggregation +
// emissions per fuel.
func TestBuildScope1PerSource(t *testing.T) {
	period := defaultPeriod()
	factors := allFactors()
	report, _ := Pack.Build(context.Background(), period, factors, newReadings(scope1Mixed()))
	body := report.Body.(Body)

	if len(body.Scope1.PerSource) != 2 {
		t.Errorf("PerSource: want 2 entries, got %d", len(body.Scope1.PerSource))
	}
	// Sorted alphabetically: diesel_combustion, natural_gas_combustion.
	if body.Scope1.PerSource[0].Code != "diesel_combustion" {
		t.Errorf("expected diesel_combustion first, got %s", body.Scope1.PerSource[0].Code)
	}
	// 50 l diesel × 2.642 kg/l = 132.1 kg.
	if !floatNear(body.Scope1.PerSource[0].KgCO2eq, 50.0*2.642, 1e-9) {
		t.Errorf("diesel kg: got %v, want ~%v", body.Scope1.PerSource[0].KgCO2eq, 50.0*2.642)
	}
	// 100 Sm3 natural gas × 1.967 kg/Sm3 = 196.7 kg.
	if !floatNear(body.Scope1.PerSource[1].KgCO2eq, 100.0*1.967, 1e-9) {
		t.Errorf("natural gas kg: got %v, want ~%v", body.Scope1.PerSource[1].KgCO2eq, 100.0*1.967)
	}
}

// TestBuildMissingFactorEmitsNote — graceful degradation when one
// Scope 2 factor is absent.
func TestBuildMissingFactorEmitsNote(t *testing.T) {
	period := defaultPeriod()
	factors := newFactorBundle(map[string]factorRow{
		FactorScope2Location: {value: 233.0, version: "ispra@2025.04"},
		// market-based intentionally absent
	})
	report, _ := Pack.Build(context.Background(), period, factors, newReadings(electricityOnly()))
	body := report.Body.(Body)

	if body.Scope2.KgCO2eqLocationBased == nil {
		t.Error("location-based should be present")
	}
	if body.Scope2.KgCO2eqMarketBased != nil {
		t.Error("market-based should be nil when factor absent")
	}
	if len(report.Notes) == 0 {
		t.Error("expected a note about the missing factor")
	}
}

// TestBuildUnclassifiedRows — readings with unknown units are counted but
// not included in scope totals.
func TestBuildUnclassifiedRows(t *testing.T) {
	period := defaultPeriod()
	factors := allFactors()
	rows := []reporting.AggregatedRow{
		{MeterID: uuid.New(), ChannelID: uuid.New(), Sum: 10, Count: 1, Unit: "weird_unit"},
		{MeterID: uuid.New(), ChannelID: uuid.New(), Sum: 20, Count: 2, Unit: "another_weird"},
	}
	report, _ := Pack.Build(context.Background(), period, factors, newReadings(rows))
	body := report.Body.(Body)
	if body.UnclassifiedRows != 3 { // 1 + 2 counts
		t.Errorf("UnclassifiedRows: want 3, got %d", body.UnclassifiedRows)
	}
	if len(report.Notes) == 0 {
		t.Error("expected a note about unclassified rows")
	}
}

// TestBuildHonoursContext — pure-function ctx cancellation.
func TestBuildHonoursContext(t *testing.T) {
	period := defaultPeriod()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Pack.Build(ctx, period, allFactors(), newReadings(nil)); err == nil {
		t.Error("expected ctx.Err() on cancelled context")
	}
}

// TestEncodedIsValidJSON — output parses as JSON with expected keys.
func TestEncodedIsValidJSON(t *testing.T) {
	period := defaultPeriod()
	report, _ := Pack.Build(context.Background(), period, allFactors(), newReadings(sampleRows()))
	var decoded map[string]any
	if err := json.Unmarshal(report.Encoded, &decoded); err != nil {
		t.Fatalf("Encoded is not valid JSON: %v", err)
	}
	for _, key := range []string{"period", "factors_used", "scope_1", "scope_2", "scope_3", "totals", "unclassified_rows"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("Encoded JSON missing key %q", key)
		}
	}
}

// TestBuildTotalsCombineScope1And2 — totals are Scope 1 + Scope 2 +
// Scope 3 (the latter being zero placeholder).
func TestBuildTotalsCombineScope1And2(t *testing.T) {
	period := defaultPeriod()
	rows := append([]reporting.AggregatedRow{}, electricityOnly()...)
	rows = append(rows, scope1Mixed()...)
	report, _ := Pack.Build(context.Background(), period, allFactors(), newReadings(rows))
	body := report.Body.(Body)

	wantS1 := 50.0*2.642 + 100.0*1.967
	if !floatNear(body.Scope1.KgCO2eqTotal, wantS1, 1e-9) {
		t.Errorf("Scope1.KgCO2eqTotal: want %v, got %v", wantS1, body.Scope1.KgCO2eqTotal)
	}
	wantTotalLoc := wantS1 + 100.0*233.0/1000.0
	if body.Totals.KgCO2eqLocationBased == nil {
		t.Fatal("Totals.KgCO2eqLocationBased should be populated")
	}
	if got := *body.Totals.KgCO2eqLocationBased; !floatNear(got, wantTotalLoc, 1e-9) {
		t.Errorf("Totals.location: want %v, got %v", wantTotalLoc, got)
	}
}

// floatNear is a tolerance comparison for float arithmetic that may be
// off by ULP-scale rounding (e.g. 100.0 * 1.967 = 196.70000000000002).
func floatNear(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
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
	// 100_000 Wh = 100 kWh total across two readings.
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
		// 100 Sm3 of natural gas
		{MeterID: meterB, ChannelID: chanB, BucketStart: t1, Sum: 100, Count: 1, Unit: "Sm3"},
		// 50 l of diesel
		{MeterID: meterC, ChannelID: chanC, BucketStart: t1, Sum: 50, Count: 1, Unit: "l_diesel"},
	}
}
