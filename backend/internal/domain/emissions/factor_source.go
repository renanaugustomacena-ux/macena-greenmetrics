// Package emissions defines the Pack contract for authoritative
// emission-factor sources.
//
// Doctrine refs: Rules 90 (versioned with temporal validity), 130 (each
// authoritative factor source is a Pack), 138 (annual review), 139
// (thresholds propagated, not duplicated).
// Charter ref: §3.2 Factor Packs. ADR-0023 records the interface adoption.
//
// A Factor Pack at packs/factor/<id>/ implements the FactorSource interface
// below. Core's emissions service reads the registered FactorSources at
// boot and refreshes them on the Pack's documented cadence (annually for
// ISPRA, daily for Terna, quarterly for GSE, etc.).
//
// Factors are temporally-versioned per Rule 90: a query for `(code, ts)`
// returns the factor whose [valid_from, valid_to) interval covers `ts`.
// Reports re-run for past periods continue to use the factor that was
// valid at the original reporting time.
package emissions

import (
	"context"
	"time"
)

// ContractVersion is the SemVer of this Pack-contract package. Per Rule 71.
const ContractVersion = "1.0.0"

// Factor is one row in the temporal factor table. (Code, ValidFromUTC) is
// the natural key; ValidToUTC is exclusive (or zero for "still active").
type Factor struct {
	Code         string    `json:"code"`
	ValidFromUTC time.Time `json:"valid_from"`
	ValidToUTC   time.Time `json:"valid_to,omitempty"`
	Value        float64   `json:"value"`
	Unit         string    `json:"unit"`
	Source       string    `json:"source"`
	SourceURL    string    `json:"source_url,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

// FactorSource is the Pack-contract for authoritative emission-factor sources.
//
// Implementations are responsible for:
//   - declaring their Name (matches the Pack's id);
//   - producing their factor rows from the authoritative upstream;
//   - returning the factor set in `Refresh`; Core writes them to
//     `emission_factors` with `(code, valid_from)` as the natural key.
//
// Refresh is called on the Pack's documented cadence by Core's scheduler.
// The cadence is configured per-Pack (e.g. annual for ISPRA — see
// internal/jobs/ispra_factor_refresh.go in Phase F Sprint S9). A Refresh
// returning an error does not crash Core; the audit log records the failure.
type FactorSource interface {
	// Name is the source identifier (matches Pack id; e.g. "ispra", "gse",
	// "terna", "aib", "uk_defra", "epa_egrid").
	Name() string

	// Refresh pulls the latest factor set from the authoritative upstream
	// (or from a checked-in offline manifest in dev). Core merges the
	// returned set into the temporal table; new (code, valid_from) tuples
	// are inserted, existing ones are not updated (immutability per Rule 96).
	//
	// Refresh MUST honour ctx; long-running pulls should chunk and check
	// ctx.Done() periodically.
	Refresh(ctx context.Context) ([]Factor, error)
}
