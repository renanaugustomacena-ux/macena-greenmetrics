// Example test for the Report Pack contract. Per Rule 86.

package reporting_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/greenmetrics/backend/internal/domain/reporting"
)

// stubFactors is a test-only FactorBundle.
type stubFactors struct {
	versions map[string]string
	vals     map[string]float64
}

func (s stubFactors) Get(k string) (float64, string, bool) {
	v, ok := s.vals[k]
	if !ok {
		return 0, "", false
	}
	return v, s.versions[k], true
}
func (s stubFactors) Versions() map[string]string { return s.versions }

// stubReadings is a test-only AggregatedReadings.
type stubReadings struct{ rows []reporting.AggregatedRow }

func (s stubReadings) Iter() reporting.AggregatedIter { return &stubIter{rows: s.rows, idx: -1} }

type stubIter struct {
	rows []reporting.AggregatedRow
	idx  int
}

func (it *stubIter) Next() bool                   { it.idx++; return it.idx < len(it.rows) }
func (it *stubIter) Row() reporting.AggregatedRow { return it.rows[it.idx] }
func (it *stubIter) Err() error                   { return nil }

// stubBuilder is a minimal Builder that sums Wh × kgCO2/kWh and emits a body.
type stubBuilder struct{}

func (stubBuilder) Type() reporting.ReportType { return reporting.ReportType("stub_v1") }
func (stubBuilder) Version() string            { return "1.0.0" }

func (b stubBuilder) Build(ctx context.Context, period reporting.Period, factors reporting.FactorBundle, readings reporting.AggregatedReadings) (reporting.Report, error) {
	totalKWh := int64(0)
	it := readings.Iter()
	for it.Next() {
		row := it.Row()
		totalKWh += row.Sum / 1_000_000 // micro-Wh → Wh→ scaled below
	}
	factor, _, _ := factors.Get("ispra_national_mix_2024")
	emissionsKgCO2 := float64(totalKWh) * factor

	encoded, err := json.Marshal(map[string]any{
		"total_wh":         totalKWh,
		"emissions_kg_co2": emissionsKgCO2,
	})
	if err != nil {
		return reporting.Report{}, err
	}

	return reporting.Report{
		Type:    b.Type(),
		Period:  period,
		Body:    map[string]any{"total_wh": totalKWh, "emissions_kg_co2": emissionsKgCO2},
		Encoded: encoded,
		Provenance: reporting.Provenance{
			FactorPackVersions: factors.Versions(),
			ReportPackVersion:  b.Version(),
			SourceDataWindow:   period,
			ExecutedAtUTC:      time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC),
		},
	}, nil
}

func TestExample_BuilderIsDeterministic(t *testing.T) {
	period := reporting.Period{
		StartInclusiveUTC: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndExclusiveUTC:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		Timezone:          "Europe/Rome",
	}
	factors := stubFactors{
		versions: map[string]string{"ispra_national_mix_2024": "2024.1"},
		vals:     map[string]float64{"ispra_national_mix_2024": 0.245},
	}
	readings := stubReadings{rows: []reporting.AggregatedRow{
		{MeterID: uuid.New(), Sum: 1_000_000_000, Unit: "Wh"},
	}}

	r1, err := stubBuilder{}.Build(context.Background(), period, factors, readings)
	if err != nil {
		t.Fatalf("Build #1: %v", err)
	}
	r2, err := stubBuilder{}.Build(context.Background(), period, factors, readings)
	if err != nil {
		t.Fatalf("Build #2: %v", err)
	}
	if string(r1.Encoded) != string(r2.Encoded) {
		t.Fatalf("Builder is non-deterministic: %s != %s", r1.Encoded, r2.Encoded)
	}
}

func TestContractVersion_IsSet(t *testing.T) {
	if reporting.ContractVersion == "" {
		t.Fatal("ContractVersion empty — Rule 71 requires per-kind contract version")
	}
}
