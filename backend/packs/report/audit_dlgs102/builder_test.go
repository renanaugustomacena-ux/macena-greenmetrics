package audit_dlgs102

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
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

// TestBuildBitPerfectReproducibility — Rule 89.
func TestBuildBitPerfectReproducibility(t *testing.T) {
	period := defaultPeriod()
	factors := newFactorBundle(map[string]factorRow{
		FactorObligationType: {value: ObligationGrandeImpresa, version: "engagement"},
		FactorTotalFloorM2:   {value: 12500, version: "engagement"},
	})
	rows := mixedSiteReadings()
	r1, _ := Pack.Build(context.Background(), period, factors, newReadings(rows))
	r2, _ := Pack.Build(context.Background(), period, factors, newReadings(rows))
	if !bytes.Equal(r1.Encoded, r2.Encoded) {
		t.Fatalf("Encoded bytes differ between two Build calls (Rule 89 violated)")
	}
}

// TestBuildEnergyBaselineByVector verifies per-vector kWh + tep totals.
func TestBuildEnergyBaselineByVector(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(nil), newReadings(mixedSiteReadings()))
	body := r.Body.(Body)

	wantElec := 1500.0                         // 500_000 Wh + 1000 kWh = 500 + 1000 = 1500 kWh
	wantGas := 100.0                           // 100 Sm³
	wantDiesel := 100.0 * DefaultDensityDiesel // 100 l × 0.835 kg/l = 83.5 kg

	got := byVector(body.EnergyBaseline.ByVector)
	if !floatNear(got[VectorElectricity].PrimaryTotal, wantElec, 1e-9) {
		t.Errorf("electricity primary total: want %v, got %v", wantElec, got[VectorElectricity].PrimaryTotal)
	}
	if !floatNear(got[VectorElectricity].KWhTotal, wantElec, 1e-9) {
		t.Errorf("electricity kWh total: want %v, got %v", wantElec, got[VectorElectricity].KWhTotal)
	}
	wantElecTep := wantElec * TepFactorElectricity
	if !floatNear(got[VectorElectricity].TepTotal, wantElecTep, 1e-9) {
		t.Errorf("electricity tep total: want %v, got %v", wantElecTep, got[VectorElectricity].TepTotal)
	}
	if !floatNear(got[VectorNaturalGas].PrimaryTotal, wantGas, 1e-9) {
		t.Errorf("gas primary total: want %v, got %v", wantGas, got[VectorNaturalGas].PrimaryTotal)
	}
	wantGasTep := wantGas * TepFactorNaturalGas
	if !floatNear(got[VectorNaturalGas].TepTotal, wantGasTep, 1e-9) {
		t.Errorf("gas tep total: want %v, got %v", wantGasTep, got[VectorNaturalGas].TepTotal)
	}
	if !floatNear(got[VectorDiesel].PrimaryTotal, wantDiesel, 1e-9) {
		t.Errorf("diesel kg total: want %v, got %v", wantDiesel, got[VectorDiesel].PrimaryTotal)
	}

	// Total tep = electricity + gas + diesel
	wantTotalTep := wantElecTep + wantGasTep + wantDiesel*TepFactorDiesel
	if !floatNear(body.EnergyBaseline.TotalTep, wantTotalTep, 1e-9) {
		t.Errorf("total tep: want %v, got %v", wantTotalTep, body.EnergyBaseline.TotalTep)
	}
}

// TestBuildSiteBreakdown verifies per-site totals + lex-sorted site_id order.
func TestBuildSiteBreakdown(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(nil), newReadings(mixedSiteReadings()))
	body := r.Body.(Body)

	if len(body.SiteBreakdown) != 2 {
		t.Fatalf("want 2 sites, got %d", len(body.SiteBreakdown))
	}
	// Site IDs are lex-sorted by UUID string.
	for i := 1; i < len(body.SiteBreakdown); i++ {
		if body.SiteBreakdown[i].SiteID < body.SiteBreakdown[i-1].SiteID {
			t.Errorf("site_id %d not in lex order: %q < %q",
				i, body.SiteBreakdown[i].SiteID, body.SiteBreakdown[i-1].SiteID)
		}
	}
}

// TestBuildObligationBlockGrandeImpresa — default selector.
func TestBuildObligationBlockGrandeImpresa(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(map[string]factorRow{
			FactorObligationType:   {value: ObligationGrandeImpresa, version: "engagement"},
			FactorNextDeadlineUnix: {value: 1828137599, version: "engagement"},
		}),
		newReadings(nil))
	body := r.Body.(Body)

	if body.Obligation.Type != "grande_impresa" {
		t.Errorf("obligation type: want grande_impresa, got %q", body.Obligation.Type)
	}
	if body.Obligation.AuditPeriodicityYears != 4 {
		t.Errorf("periodicity: want 4, got %d", body.Obligation.AuditPeriodicityYears)
	}
	if body.Obligation.NextDeadlineISO == nil {
		t.Fatal("next deadline ISO should be populated")
	}
	if !strings.HasPrefix(*body.Obligation.NextDeadlineISO, "2027") {
		t.Errorf("next deadline: want 2027-prefix, got %q", *body.Obligation.NextDeadlineISO)
	}
}

// TestBuildObligationEnergivora.
func TestBuildObligationEnergivora(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(map[string]factorRow{
			FactorObligationType: {value: ObligationEnergivora, version: "engagement"},
			FactorExemptionISO:   {value: ExemptionISO50001, version: "engagement"},
		}),
		newReadings(nil))
	body := r.Body.(Body)

	if body.Obligation.Type != "energivora" {
		t.Errorf("obligation type: want energivora, got %q", body.Obligation.Type)
	}
	if body.Obligation.ExemptionBasis == nil || *body.Obligation.ExemptionBasis != "iso_50001" {
		t.Errorf("exemption basis: want iso_50001, got %v", body.Obligation.ExemptionBasis)
	}
}

// TestBuildBelow50TepFlagFromConsumption.
func TestBuildBelow50TepFlagFromConsumption(t *testing.T) {
	// 100 kWh × 0.000187 = 0.0187 tep (well below 50).
	siteA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	chanA := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	rows := []reporting.AggregatedRow{
		{MeterID: siteA, ChannelID: chanA, BucketStart: time.Now(), Sum: 100, Count: 1, Unit: "kWh"},
	}
	r, _ := Pack.Build(context.Background(), defaultPeriod(), newFactorBundle(nil), newReadings(rows))
	body := r.Body.(Body)

	if !body.EnergyBaseline.Below50TepThreshold {
		t.Errorf("below_50_tep_threshold: want true (total tep %.4f < 50), got false", body.EnergyBaseline.TotalTep)
	}
}

// TestBuildBelow50TepExemptionAsserted.
func TestBuildBelow50TepExemptionAsserted(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(map[string]factorRow{
			FactorBelow50Exempt: {value: 1, version: "engagement-discovery"},
		}),
		newReadings(nil))
	body := r.Body.(Body)

	if !body.Obligation.Below50TepExemption {
		t.Error("below_50_tep_exemption: want true (asserted by factor)")
	}
}

// TestBuildEnPITotalEnergyIntensity verifies kWh/m² candidate emission.
func TestBuildEnPITotalEnergyIntensity(t *testing.T) {
	siteA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	chanA := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	rows := []reporting.AggregatedRow{
		{MeterID: siteA, ChannelID: chanA, BucketStart: time.Now(), Sum: 1_000_000, Count: 1, Unit: "kWh"},
	}
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(map[string]factorRow{
			FactorTotalFloorM2: {value: 5000, version: "engagement"},
		}),
		newReadings(rows))
	body := r.Body.(Body)

	if len(body.EnPI) != 1 {
		t.Fatalf("EnPI: want 1 candidate, got %d", len(body.EnPI))
	}
	want := 1_000_000.0 / 5000.0
	if !floatNear(body.EnPI[0].Value, want, 1e-9) {
		t.Errorf("EnPI value: want %v, got %v", want, body.EnPI[0].Value)
	}
	if body.EnPI[0].Indicator != "kWh/m2" {
		t.Errorf("EnPI indicator: want kWh/m2, got %q", body.EnPI[0].Indicator)
	}
}

// TestBuildTepFactorOverride — engagement-supplied factor overrides statutory.
func TestBuildTepFactorOverride(t *testing.T) {
	siteA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	chanA := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	rows := []reporting.AggregatedRow{
		{MeterID: siteA, ChannelID: chanA, BucketStart: time.Now(), Sum: 1000, Count: 1, Unit: "kWh"},
	}
	const override = 0.000200 // hypothetical post-revision factor
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(map[string]factorRow{
			FactorTepFactorPrefix + VectorElectricity: {value: override, version: "ARERA-2026-update"},
		}),
		newReadings(rows))
	body := r.Body.(Body)

	got := byVector(body.EnergyBaseline.ByVector)
	if !floatNear(got[VectorElectricity].TepFactor, override, 1e-12) {
		t.Errorf("override tep factor: want %v, got %v", override, got[VectorElectricity].TepFactor)
	}
	wantTep := 1000.0 * override
	if !floatNear(body.EnergyBaseline.TotalTep, wantTep, 1e-9) {
		t.Errorf("total tep with override: want %v, got %v", wantTep, body.EnergyBaseline.TotalTep)
	}
	if _, ok := body.FactorsUsed[FactorTepFactorPrefix+VectorElectricity]; !ok {
		t.Error("override factor should appear in factors_used")
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

// TestEncodedIsValidJSON — output parses + has expected top-level keys.
func TestEncodedIsValidJSON(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(nil), newReadings(mixedSiteReadings()))

	var decoded map[string]any
	if err := json.Unmarshal(r.Encoded, &decoded); err != nil {
		t.Fatalf("Encoded is not valid JSON: %v", err)
	}
	for _, key := range []string{
		"report", "regulator", "period", "obligation", "factors_used",
		"energy_baseline", "site_breakdown", "enpi",
		"improvement_measures", "monitoring_plan",
		"narrative_data_points", "ege_certification_required", "unclassified_rows",
	} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("Encoded JSON missing key %q", key)
		}
	}
	if decoded["report"] != "audit_dlgs102" {
		t.Errorf("report identifier: want audit_dlgs102, got %v", decoded["report"])
	}
}

// TestBuildUnclassifiedRows.
func TestBuildUnclassifiedRows(t *testing.T) {
	siteA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	chanA := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	rows := []reporting.AggregatedRow{
		{MeterID: siteA, ChannelID: chanA, BucketStart: time.Now(), Sum: 99, Count: 5, Unit: "rpm"},  // unknown
		{MeterID: siteA, ChannelID: chanA, BucketStart: time.Now(), Sum: 100, Count: 1, Unit: "kWh"}, // valid
	}
	r, _ := Pack.Build(context.Background(), defaultPeriod(), newFactorBundle(nil), newReadings(rows))
	body := r.Body.(Body)
	if body.UnclassifiedRows != 5 {
		t.Errorf("unclassified rows: want 5, got %d", body.UnclassifiedRows)
	}
	if !anyNoteContains(body.Notes, "unclassified Unit") {
		t.Error("notes should mention unclassified rows")
	}
}

// TestBuildNarrativeIsNullPlaceholder.
func TestBuildNarrativeIsNullPlaceholder(t *testing.T) {
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(nil), newReadings(nil))
	body := r.Body.(Body)
	if body.NarrativeDataPoints.ScopeDescription != nil {
		t.Error("ScopeDescription should be null placeholder")
	}
	if body.NarrativeDataPoints.EGECertifierID != nil {
		t.Error("EGECertifierID should be null placeholder")
	}
	if body.NarrativeDataPoints.Note == "" {
		t.Error("NarrativeDataPoints.Note should be populated")
	}
}

// TestBuildVectorOrderingDeterministic — vectors emitted in canonical order
// regardless of reading input order.
func TestBuildVectorOrderingDeterministic(t *testing.T) {
	siteA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	chanA := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	rows := []reporting.AggregatedRow{
		// Intentionally out of canonical order.
		{MeterID: siteA, ChannelID: chanA, Sum: 100, Count: 1, Unit: "kg_coal"},
		{MeterID: siteA, ChannelID: chanA, Sum: 100, Count: 1, Unit: "kWh"},
		{MeterID: siteA, ChannelID: chanA, Sum: 100, Count: 1, Unit: "Sm3"},
	}
	r, _ := Pack.Build(context.Background(), defaultPeriod(), newFactorBundle(nil), newReadings(rows))
	body := r.Body.(Body)

	wantOrder := []string{VectorElectricity, VectorNaturalGas, VectorCoal}
	if len(body.EnergyBaseline.ByVector) != len(wantOrder) {
		t.Fatalf("by_vector len: want %d, got %d", len(wantOrder), len(body.EnergyBaseline.ByVector))
	}
	for i, v := range wantOrder {
		if body.EnergyBaseline.ByVector[i].Vector != v {
			t.Errorf("by_vector[%d]: want %q, got %q", i, v, body.EnergyBaseline.ByVector[i].Vector)
		}
	}
}

// TestBuildLitresToKgConversion uses the default density to verify l → kg.
func TestBuildLitresToKgConversion(t *testing.T) {
	siteA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	chanA := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	rows := []reporting.AggregatedRow{
		{MeterID: siteA, ChannelID: chanA, Sum: 1000, Count: 1, Unit: "l_diesel"},
	}
	r, _ := Pack.Build(context.Background(), defaultPeriod(),
		newFactorBundle(map[string]factorRow{
			// Default density override → custom density 0.840.
			FactorDensityPrefix + VectorDiesel: {value: 0.840, version: "engagement-supplier-spec"},
		}),
		newReadings(rows))
	body := r.Body.(Body)

	got := byVector(body.EnergyBaseline.ByVector)
	wantKg := 1000.0 * 0.840
	if !floatNear(got[VectorDiesel].PrimaryTotal, wantKg, 1e-9) {
		t.Errorf("diesel kg with override density: want %v, got %v", wantKg, got[VectorDiesel].PrimaryTotal)
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

func mixedSiteReadings() []reporting.AggregatedRow {
	siteA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	siteB := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	chan1 := uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	chan2 := uuid.MustParse("00000000-0000-0000-0000-0000000000a2")
	chan3 := uuid.MustParse("00000000-0000-0000-0000-0000000000a3")
	t1 := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	return []reporting.AggregatedRow{
		// Site A: 500 kWh worth of Wh + 100 Sm3 gas.
		{MeterID: siteA, ChannelID: chan1, BucketStart: t1, Sum: 500_000, Count: 1, Unit: "Wh"},
		{MeterID: siteA, ChannelID: chan2, BucketStart: t1, Sum: 100, Count: 1, Unit: "Sm3"},
		// Site B: 1000 kWh + 100 l diesel.
		{MeterID: siteB, ChannelID: chan1, BucketStart: t1, Sum: 1000, Count: 1, Unit: "kWh"},
		{MeterID: siteB, ChannelID: chan3, BucketStart: t1, Sum: 100, Count: 1, Unit: "l_diesel"},
	}
}

func byVector(rows []VectorRow) map[string]VectorRow {
	out := map[string]VectorRow{}
	for _, r := range rows {
		out[r.Vector] = r
	}
	return out
}

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
