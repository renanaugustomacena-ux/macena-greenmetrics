// Package reporting defines the Pack contract for regulatory-dossier builders.
//
// Doctrine refs: Rules 89 (bit-perfect reproducibility), 91 (pure functions),
// 95 (provenance bundle), 97 (algorithm versioning), 129 (each dossier is a
// Pack), 141 (deterministic serialisation), 144 (signed at finalisation).
// Charter ref: §3.2 Report Packs. ADR-0023 records the interface adoption.
//
// A Report Pack at packs/report/<id>/ implements the Builder interface
// below. Core's reporting orchestrator (internal/handlers/reports.go)
// dispatches a generation request to the registered Builder by ReportType.
//
// Builders MUST be pure functions (Rule 91). The runtime enforces this at
// the conformance suite (tests/conformance/builder_purity_test.go) and at
// the custom golangci-lint rule that bans `time.Now()`, `os.Getenv`, and
// `internal/services/` imports inside `internal/domain/reporting/<id>/`.
package reporting

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ContractVersion is the SemVer of this Pack-contract package. Per Rule 71.
const ContractVersion = "1.0.0"

// ReportType identifies the kind of regulatory dossier a Builder produces.
// Each Report Pack registers exactly one type. Reports referenced by the
// Italian-flagship Pack: esrs_e1, piano_5_0, conto_termico, tee,
// audit_dlgs102, monthly_consumption, co2_footprint.
type ReportType string

// Period is a half-open time interval per Rule 142. Both endpoints are
// RFC 3339 UTC per Rule 2. The Builder reads readings WHERE
// `ts >= StartInclusiveUTC AND ts < EndExclusiveUTC`.
type Period struct {
	StartInclusiveUTC time.Time `json:"period_start_inclusive"`
	EndExclusiveUTC   time.Time `json:"period_end_exclusive"`
	// Timezone is the per-tenant timezone (Rule 101) used for human-readable
	// rendering; the underlying arithmetic remains in UTC.
	Timezone string `json:"timezone"`
}

// Provenance bundles the metadata that makes a report bit-perfectly
// re-derivable (Rule 95). Stored alongside the report in the database;
// signed via Cosign sign-blob at the `signed` state transition.
type Provenance struct {
	ManifestLockHash   string            `json:"manifest_lock_hash"`
	FactorPackVersions map[string]string `json:"factor_pack_versions"`
	ReportPackVersion  string            `json:"report_pack_version"`
	QueryDefinitions   []string          `json:"query_definitions"`
	SourceDataWindow   Period            `json:"source_data_window"`
	TenantDataRegion   string            `json:"tenant_data_region"`
	ExecutorUserID     uuid.UUID         `json:"executor_user_id"`
	ExecutedAtUTC      time.Time         `json:"executed_at_utc"`
}

// Report is the output of a Builder. Body is the canonical typed payload
// (the report's data); Encoded is the deterministic serialisation that
// gets signed and stored. Lineage is the queryable view (Rule 99).
type Report struct {
	Type       ReportType `json:"type"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	Period     Period     `json:"period"`
	Body       any        `json:"body"`
	Encoded    []byte     `json:"-"`
	Provenance Provenance `json:"provenance"`
	Notes      []string   `json:"notes,omitempty"`
}

// FactorBundle is the set of factor values valid at the report period. The
// Builder reads factors via the keys it registers; lookups outside the
// bundle are forbidden (defeat purity).
//
// Concrete shape and lookup methods live in internal/domain/emissions/
// (FactorSource Pack contract). The reporting package re-exports the
// type minimally to avoid circular imports.
type FactorBundle interface {
	// Get returns the factor value for `key` valid at the period start, or
	// `false` if not in the bundle.
	Get(key string) (value float64, version string, ok bool)
	// Versions returns the per-source versions used; surfaces in Provenance.
	Versions() map[string]string
}

// AggregatedReadings is the input the Builder reads. The reporting orchestrator
// pre-aggregates readings (per CAGG window, per channel) before dispatching
// to the Builder so that Builders don't run raw queries.
type AggregatedReadings interface {
	// Iter returns an iterator over aggregated rows. Rows are sorted by
	// (meter_id, channel_id, ts) per Rule 141 (deterministic input order).
	Iter() AggregatedIter
}

// AggregatedIter is a read-only iterator over aggregated rows.
type AggregatedIter interface {
	Next() bool
	Row() AggregatedRow
	Err() error
}

// AggregatedRow is one CAGG row.
type AggregatedRow struct {
	MeterID     uuid.UUID
	ChannelID   uuid.UUID
	BucketStart time.Time
	BucketEnd   time.Time
	Sum         int64
	Count       int64
	Unit        string
	QualityMix  map[string]int
}

// Builder is the Pack-contract for regulatory-dossier builders.
//
// Implementations MUST satisfy purity (Rule 91). Implementations MUST be
// deterministic — same inputs → same Encoded bytes. The conformance test
// runs each Builder twice with the same inputs and asserts byte-identical
// Encoded.
type Builder interface {
	// Type returns the ReportType this Builder produces. Each Builder
	// registers exactly one type.
	Type() ReportType

	// Version returns the report-pack-version per Rule 97. A change to the
	// Builder's algorithm bumps this and ships an ADR explaining the change.
	Version() string

	// Build is the pure computation. Implementations read only the
	// arguments — no global state, no I/O outside ctx-bound resources, no
	// time.Now(), no environment variables.
	Build(ctx context.Context, period Period, factors FactorBundle, readings AggregatedReadings) (Report, error)
}
