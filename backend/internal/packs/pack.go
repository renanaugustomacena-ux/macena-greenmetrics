// Package packs defines the Pack contract — the load-bearing seam between
// Core and per-engagement, per-flavour swappable implementations.
//
// Doctrine refs: Rules 69–88 (Modular Template Integrity).
// Charter ref: docs/MODULAR-TEMPLATE-CHARTER.md §3 (Core vs. Pack), §4 (Pack contract).
//
// A Pack is a self-contained directory under packs/<kind>/<id>/ that satisfies
// a documented Pack contract. Core's loader (loader.go) reads each Pack's
// manifest, validates it against the schema, instantiates the Pack, calls
// Init, then Register, and finally writes manifest.lock.json (Cosign-signed)
// at successful boot.
//
// This file ships the public interface. Concrete loader, registry, and lock
// implementations live alongside.
package packs

import (
	"context"
	"time"
)

// PackKind enumerates the five Pack flavours per Charter §3.2. New kinds
// require a charter amendment (Rule 209) and a new contract package under
// internal/domain/<kind>/.
type PackKind string

const (
	KindProtocol PackKind = "protocol"
	KindFactor   PackKind = "factor"
	KindReport   PackKind = "report"
	KindIdentity PackKind = "identity"
	KindRegion   PackKind = "region"
)

// PackContractVersion is the Core-side declaration of which Pack-contract
// versions are supported. Bumping this is a Rule 71 event — every existing
// Pack must declare a contract version inside the supported window.
const PackContractVersion = "1.0.0"

// HealthStatus is the per-Pack health status surfaced into the Core health
// envelope per Rule 74.
type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"
	HealthDegraded  HealthStatus = "degraded"
	HealthUnhealthy HealthStatus = "unhealthy"
)

// PackHealth is the per-Pack health structure. The dependencies field is
// surfaced under the Pack's id in the /api/health envelope's dependencies map.
type PackHealth struct {
	Status      HealthStatus      `json:"status"`
	Message     string            `json:"message,omitempty"`
	LastChecked time.Time         `json:"last_checked"`
	LatencyMS   int64             `json:"latency_ms,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
}

// CoreHandle is the typed surface a Pack receives at Init time. It exposes
// only the Core capabilities a Pack needs — never raw access to Core internals.
//
// The interface is intentionally narrow. Adding a method is a Rule 71 event
// (Pack-contract version bump). Removing a method is a Rule 78 event
// (merge-friendliness break) and requires a major Core release.
type CoreHandle interface {
	// PackID returns the Pack identifier this handle is scoped to.
	PackID() string
	// Logger returns a structured logger pre-bound with the Pack's id.
	Logger() Logger
	// Tracer returns an OpenTelemetry tracer pre-bound with the Pack's id.
	Tracer() Tracer
	// Config exposes Pack-scoped config from config/required-packs.yaml + per-Pack defaults.
	Config() PackConfig
}

// Logger is a minimal structured-logger contract. Implementations bridge to zap.
type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
}

// Tracer is a minimal tracer contract. Implementations bridge to OpenTelemetry.
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

// Span is the trace-span surface.
type Span interface {
	SetAttribute(k string, v any)
	RecordError(err error)
	End()
}

// PackConfig is the typed handle for Pack-scoped configuration. Returns nil
// for keys the Pack hasn't registered.
type PackConfig interface {
	String(key string) string
	Int(key string) int64
	Bool(key string) bool
	Duration(key string) time.Duration
	Has(key string) bool
}

// Registrar receives a Pack's contributions during Pack.Register. The Pack
// calls one of the typed registration methods per its kind. Cross-kind
// registration is forbidden — a Protocol Pack registering a Builder is a
// loader-level reject.
type Registrar interface {
	// RegisterIngestor is called by Protocol Packs.
	RegisterIngestor(ingestor any) error
	// RegisterFactorSource is called by Factor Packs.
	RegisterFactorSource(source any) error
	// RegisterBuilder is called by Report Packs.
	RegisterBuilder(builder any) error
	// RegisterIdentityProvider is called by Identity Packs.
	RegisterIdentityProvider(provider any) error
	// RegisterRegionProfile is called by Region Packs.
	RegisterRegionProfile(profile any) error
}

// Pack is the contract every Pack satisfies.
//
// Lifecycle, in order:
//  1. Loader reads the manifest at packs/<kind>/<id>/manifest.yaml.
//  2. Loader validates the manifest against docs/contracts/pack-manifest.schema.json.
//  3. Loader checks min_core_version and pack_contract_version against the
//     Core-supported window.
//  4. Loader instantiates the Pack (concrete type returned by the Pack's New()
//     constructor — see per-Pack convention).
//  5. Loader calls Pack.Manifest() and confirms the in-code manifest matches
//     the on-disk manifest (defence in depth — the Pack cannot lie about
//     its own identity).
//  6. Loader calls Pack.Init(ctx, core) with a typed CoreHandle. A non-nil
//     error fails the Pack's load (and aborts Core boot if the Pack is
//     declared required in config/required-packs.yaml).
//  7. Loader calls Pack.Register(reg) to receive contributions.
//  8. Loader records the Pack in manifest.lock.json (Rule 73).
//  9. At /api/health, the Pack's Health(ctx) is invoked and surfaced in the
//     health envelope (Rule 74).
//  10. On graceful shutdown, the Pack's Shutdown(ctx) is invoked with a
//     30-second budget (Rule 42).
//
// Implementations must be safe for concurrent calls to Health and Shutdown.
type Pack interface {
	Manifest() PackManifest
	Init(ctx context.Context, core CoreHandle) error
	Register(reg Registrar) error
	Health(ctx context.Context) PackHealth
	Shutdown(ctx context.Context) error
}
