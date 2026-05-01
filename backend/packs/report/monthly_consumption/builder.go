// Package monthly_consumption implements the Monthly Consumption Report
// Pack — the simplest of the Italian-flagship Report Packs and the
// reference implementation for the Builder contract (Rule 91), temporal
// factor lookup (Rule 90), deterministic serialisation (Rule 141), and
// provenance bundle (Rule 95).
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/reporting/builder.go
//   - Manifest:         packs/report/monthly_consumption/manifest.yaml
//   - Charter:          packs/report/monthly_consumption/CHARTER.md
package monthly_consumption

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/greenmetrics/backend/internal/domain/reporting"
)

// Pack is the singleton instance constructed at boot.
var Pack reporting.Builder = &builder{}

// PackVersion is the Pack's SemVer (matches manifest.yaml version).
const PackVersion = "1.0.0"

// ReportType is the report-type identifier this Builder registers.
const ReportType reporting.ReportType = "monthly_consumption"

// ScopeTwoFactorKey is the FactorBundle key consulted for the Scope 2
// location-based emissions overlay. The Italian Region Pack defaults to
// the ISPRA factor source publishing under this code.
const ScopeTwoFactorKey = "it_grid_mix_location"

// builder is the concrete Builder implementation.
type builder struct{}

// Type implements reporting.Builder.
func (b *builder) Type() reporting.ReportType { return ReportType }

// Version implements reporting.Builder.
func (b *builder) Version() string { return PackVersion }

// Body is the canonical typed payload returned in Report.Body.
type Body struct {
	Period       reporting.Period `json:"period"`
	FactorUsed   *FactorRef       `json:"factor_used,omitempty"`
	PerGroup     []GroupRow       `json:"per_group"`
	Total        TotalsRow        `json:"total"`
}

// FactorRef captures the factor value used for the Scope 2 overlay. The
// version field is the Factor Pack's version (per Rule 95 provenance).
type FactorRef struct {
	Key     string  `json:"key"`
	Value   float64 `json:"value"`
	Unit    string  `json:"unit"`
	Version string  `json:"version"`
}

// GroupRow is one (meter, channel) aggregated row.
type GroupRow struct {
	MeterID         uuid.UUID `json:"meter_id"`
	ChannelID       uuid.UUID `json:"channel_id"`
	ReadingCount    int64     `json:"reading_count"`
	EnergyKWh       float64   `json:"energy_kwh"`
	Scope2KgCO2eq   *float64  `json:"scope_2_kg_co2eq,omitempty"`
}

// TotalsRow aggregates across all groups.
type TotalsRow struct {
	GroupCount    int      `json:"group_count"`
	ReadingCount  int64    `json:"reading_count"`
	EnergyKWh     float64  `json:"energy_kwh"`
	Scope2KgCO2eq *float64 `json:"scope_2_kg_co2eq,omitempty"`
}

// Build implements reporting.Builder. Pure function — reads only the
// arguments. No time.Now(), no env, no I/O outside ctx.
func (b *builder) Build(
	ctx context.Context,
	period reporting.Period,
	factors reporting.FactorBundle,
	readings reporting.AggregatedReadings,
) (reporting.Report, error) {
	if err := ctx.Err(); err != nil {
		return reporting.Report{}, err
	}

	type groupKey struct {
		meterID, channelID uuid.UUID
	}
	type groupAcc struct {
		readingCount int64
		sumWh        int64
		unit         string
	}
	groups := map[groupKey]*groupAcc{}

	iter := readings.Iter()
	for iter.Next() {
		row := iter.Row()
		k := groupKey{row.MeterID, row.ChannelID}
		acc := groups[k]
		if acc == nil {
			acc = &groupAcc{unit: row.Unit}
			groups[k] = acc
		}
		acc.readingCount += row.Count
		acc.sumWh += row.Sum
	}
	if err := iter.Err(); err != nil {
		return reporting.Report{}, fmt.Errorf("readings iteration: %w", err)
	}

	// Look up the Scope 2 factor at the period midpoint (Rule 90).
	var factorRef *FactorRef
	factorVal, factorVer, ok := factors.Get(ScopeTwoFactorKey)
	if ok {
		factorRef = &FactorRef{
			Key:     ScopeTwoFactorKey,
			Value:   factorVal,
			Unit:    "g CO2eq/kWh",
			Version: factorVer,
		}
	}

	// Sort group keys deterministically (Rule 141).
	sortedKeys := make([]groupKey, 0, len(groups))
	for k := range groups {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		if c := bytes.Compare(sortedKeys[i].meterID[:], sortedKeys[j].meterID[:]); c != 0 {
			return c < 0
		}
		return bytes.Compare(sortedKeys[i].channelID[:], sortedKeys[j].channelID[:]) < 0
	})

	notes := []string{}
	perGroup := make([]GroupRow, 0, len(sortedKeys))
	var totalKWh float64
	var totalReadingCount int64
	var anyFactorMissing bool

	for _, k := range sortedKeys {
		acc := groups[k]
		kWh := float64(acc.sumWh) / 1000.0
		row := GroupRow{
			MeterID:      k.meterID,
			ChannelID:    k.channelID,
			ReadingCount: acc.readingCount,
			EnergyKWh:    kWh,
		}
		if factorRef != nil {
			kg := kWh * factorRef.Value / 1000.0
			row.Scope2KgCO2eq = &kg
		} else {
			anyFactorMissing = true
		}
		perGroup = append(perGroup, row)
		totalKWh += kWh
		totalReadingCount += acc.readingCount
	}

	totals := TotalsRow{
		GroupCount:   len(perGroup),
		ReadingCount: totalReadingCount,
		EnergyKWh:    totalKWh,
	}
	if factorRef != nil {
		totalScope2 := totalKWh * factorRef.Value / 1000.0
		totals.Scope2KgCO2eq = &totalScope2
	}

	if anyFactorMissing {
		notes = append(notes, fmt.Sprintf(
			"factor %q missing for the period; Scope 2 overlay omitted on affected groups",
			ScopeTwoFactorKey))
	}

	body := Body{
		Period:     period,
		FactorUsed: factorRef,
		PerGroup:   perGroup,
		Total:      totals,
	}

	encoded, err := encode(body)
	if err != nil {
		return reporting.Report{}, fmt.Errorf("encode: %w", err)
	}

	report := reporting.Report{
		Type:    ReportType,
		Period:  period,
		Body:    body,
		Encoded: encoded,
		Notes:   notes,
		// Provenance is populated by Core's reporting orchestrator at the
		// signed state transition; the Builder leaves it unset.
	}
	return report, nil
}

// encode serialises body deterministically (Rule 141): JSON with
// alphabetically-sorted keys + 2-space indent + trailing newline.
//
// We use json.MarshalIndent which sorts struct field order from the
// declaration; for map values inside Body we explicitly sort keys via
// the typed shape (no map[string]any here, so encoding/json's order is
// already deterministic).
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
