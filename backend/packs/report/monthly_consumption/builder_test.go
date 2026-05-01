package monthly_consumption

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/greenmetrics/backend/internal/domain/reporting"
)

// TestPackImplementsBuilder is the compile-time + runtime check that the
// Pack satisfies the contract.
func TestPackImplementsBuilder(t *testing.T) {
	var _ reporting.Builder = Pack
	if Pack.Type() != ReportType {
		t.Fatalf("expected type=%q, got %q", ReportType, Pack.Type())
	}
	if Pack.Version() != PackVersion {
		t.Fatalf("expected version=%q, got %q", PackVersion, Pack.Version())
	}
}

// TestBuildBitPerfectReproducibility — Rule 89. Two consecutive Build
// calls with the same inputs produce byte-identical Encoded payloads.
func TestBuildBitPerfectReproducibility(t *testing.T) {
	period := reporting.Period{
		StartInclusiveUTC: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndExclusiveUTC:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Timezone:          "Europe/Rome",
	}
	factors := newFactorBundle(map[string]factorRow{
		ScopeTwoFactorKey: {value: 233.0, version: "ispra@2025.04"},
	})
	readings := newReadings(sampleRows())

	r1, err := Pack.Build(context.Background(), period, factors, readings)
	if err != nil {
		t.Fatalf("first Build: %v", err)
	}
	r2, err := Pack.Build(context.Background(), period, factors, newReadings(sampleRows()))
	if err != nil {
		t.Fatalf("second Build: %v", err)
	}
	if !bytes.Equal(r1.Encoded, r2.Encoded) {
		t.Fatalf("Encoded bytes differ between two Build calls (Rule 89 violated):\n--- first ---\n%s\n--- second ---\n%s",
			string(r1.Encoded), string(r2.Encoded))
	}
}

// TestBuildAggregates verifies the consumption arithmetic: 4 groups of
// readings → 4 group rows; totals add up; Scope 2 derived from factor.
func TestBuildAggregates(t *testing.T) {
	period := reporting.Period{
		StartInclusiveUTC: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndExclusiveUTC:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Timezone:          "Europe/Rome",
	}
	factors := newFactorBundle(map[string]factorRow{
		ScopeTwoFactorKey: {value: 233.0, version: "ispra@2025.04"},
	})
	readings := newReadings(sampleRows())

	report, err := Pack.Build(context.Background(), period, factors, readings)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	body, ok := report.Body.(Body)
	if !ok {
		t.Fatalf("Body type: want monthly_consumption.Body, got %T", report.Body)
	}
	if body.Total.GroupCount != 2 {
		t.Errorf("GroupCount: want 2 (two unique meter+channel pairs), got %d", body.Total.GroupCount)
	}

	// sampleRows: meter A has 2 readings of 100Wh each = 200Wh = 0.2 kWh.
	// meter B has 2 readings of 250Wh each = 500Wh = 0.5 kWh.
	// Total = 0.7 kWh.
	const want = 0.7
	if got := body.Total.EnergyKWh; got != want {
		t.Errorf("Total.EnergyKWh: want %v, got %v", want, got)
	}
	// Scope 2 = 0.7 kWh × 233 g/kWh / 1000 = 0.1631 kg.
	if body.Total.Scope2KgCO2eq == nil {
		t.Fatal("Total.Scope2KgCO2eq should not be nil when factor present")
	}
	const wantCO2 = 0.7 * 233.0 / 1000.0
	if got := *body.Total.Scope2KgCO2eq; got != wantCO2 {
		t.Errorf("Total.Scope2KgCO2eq: want %v, got %v", wantCO2, got)
	}
}

// TestBuildMissingFactorEmitsNote verifies the report degrades gracefully
// when the Scope 2 factor is absent (per CHARTER §2 step 4).
func TestBuildMissingFactorEmitsNote(t *testing.T) {
	period := reporting.Period{
		StartInclusiveUTC: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndExclusiveUTC:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Timezone:          "Europe/Rome",
	}
	factors := newFactorBundle(nil) // empty — no factor present
	readings := newReadings(sampleRows())

	report, err := Pack.Build(context.Background(), period, factors, readings)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	body := report.Body.(Body)
	if body.FactorUsed != nil {
		t.Error("FactorUsed should be nil when factor missing")
	}
	if body.Total.Scope2KgCO2eq != nil {
		t.Error("Total.Scope2KgCO2eq should be nil when factor missing")
	}
	if len(report.Notes) == 0 {
		t.Error("Notes should mention missing-factor case")
	}
}

// TestBuildSortedDeterministic verifies group-row ordering is by
// (meter_id, channel_id) bytewise, regardless of input order.
func TestBuildSortedDeterministic(t *testing.T) {
	period := reporting.Period{
		StartInclusiveUTC: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndExclusiveUTC:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Timezone:          "Europe/Rome",
	}
	factors := newFactorBundle(nil)

	// Build twice with rows in different orders — Encoded must match.
	rowsA := sampleRows()
	rowsB := []reporting.AggregatedRow{rowsA[3], rowsA[0], rowsA[2], rowsA[1]}

	r1, _ := Pack.Build(context.Background(), period, factors, newReadings(rowsA))
	r2, _ := Pack.Build(context.Background(), period, factors, newReadings(rowsB))

	if !bytes.Equal(r1.Encoded, r2.Encoded) {
		t.Errorf("group ordering not deterministic on input permutation:\n--- A ---\n%s\n--- B ---\n%s",
			string(r1.Encoded), string(r2.Encoded))
	}
}

// TestBuildHonoursContext verifies Build exits on context cancellation.
func TestBuildHonoursContext(t *testing.T) {
	period := reporting.Period{
		StartInclusiveUTC: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndExclusiveUTC:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Timezone:          "Europe/Rome",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Pack.Build(ctx, period, newFactorBundle(nil), newReadings(nil))
	if err == nil {
		t.Error("expected ctx.Err() on cancelled context")
	}
}

// TestEncodedIsValidJSON verifies the deterministic encoder produces
// parseable JSON with the expected top-level keys.
func TestEncodedIsValidJSON(t *testing.T) {
	period := reporting.Period{
		StartInclusiveUTC: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndExclusiveUTC:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Timezone:          "Europe/Rome",
	}
	factors := newFactorBundle(map[string]factorRow{
		ScopeTwoFactorKey: {value: 233.0, version: "ispra@2025.04"},
	})
	report, err := Pack.Build(context.Background(), period, factors, newReadings(sampleRows()))
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(report.Encoded, &decoded); err != nil {
		t.Fatalf("Encoded is not valid JSON: %v", err)
	}
	for _, key := range []string{"period", "factor_used", "per_group", "total"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("Encoded JSON missing key %q", key)
		}
	}
}

// ── Test helpers ───────────────────────────────────────────────────────

// factorRow is the minimal struct returned by the test FactorBundle.
type factorRow struct {
	value   float64
	version string
}

// stubFactorBundle satisfies reporting.FactorBundle for tests.
type stubFactorBundle struct {
	rows map[string]factorRow
}

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

// stubReadings satisfies reporting.AggregatedReadings for tests.
type stubReadings struct {
	rows []reporting.AggregatedRow
}

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

// sampleRows returns a deterministic 4-row dataset across two
// (meter, channel) groups: 2 rows of 100 Wh each on group A, 2 rows of
// 250 Wh each on group B.
func sampleRows() []reporting.AggregatedRow {
	meterA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	chanA := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	meterB := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	chanB := uuid.MustParse("00000000-0000-0000-0000-0000000000b1")
	t1 := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	return []reporting.AggregatedRow{
		{MeterID: meterA, ChannelID: chanA, BucketStart: t1, BucketEnd: t1.Add(15 * time.Minute), Sum: 100, Count: 1, Unit: "Wh"},
		{MeterID: meterA, ChannelID: chanA, BucketStart: t2, BucketEnd: t2.Add(15 * time.Minute), Sum: 100, Count: 1, Unit: "Wh"},
		{MeterID: meterB, ChannelID: chanB, BucketStart: t1, BucketEnd: t1.Add(15 * time.Minute), Sum: 250, Count: 1, Unit: "Wh"},
		{MeterID: meterB, ChannelID: chanB, BucketStart: t2, BucketEnd: t2.Add(15 * time.Minute), Sum: 250, Count: 1, Unit: "Wh"},
	}
}
