// Package protocol defines the Pack contract for OT-protocol ingestors.
//
// Doctrine refs: Rules 109 (every OT protocol is a Pack), 110 (wire-format
// invariants documented), 113 (quality codes), 121 (per-protocol simulators).
// Charter ref: §3.2 Protocol Packs. ADR-0023 records the interface adoption.
//
// A Protocol Pack at packs/protocol/<id>/ implements the Ingestor interface
// below. Core's IngestorRunner (internal/services/ingestor_runner.go) starts
// every registered Ingestor and feeds their batches into the ReadingSink.
//
// New protocols are added by writing a new Pack — Core's runner does not
// change. The flagship Italian Region Pack today ships seven Protocol Packs
// (modbus_tcp, modbus_rtu, mbus, sunspec, pulse, ocpp_1_6, ocpp_2_0_1);
// Phase G adds iec_61850, opc_ua, mqtt_sparkplug_b, bacnet, iec_62056_21.
package protocol

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ContractVersion is the SemVer of this Pack-contract package. Per Rule 71
// it evolves independently of Core's version. A breaking change to the
// Ingestor interface bumps this; Packs declare which contract version they
// satisfy in their manifest.
const ContractVersion = "1.0.0"

// QualityCode enumerates the quality stamps a Protocol Pack may attach to
// each reading per Rule 113. Builders branch on these — `good` is included
// in regulatory output, `interpolated` is included with a flag, others are
// excluded.
type QualityCode string

const (
	QualityGood               QualityCode = "good"
	QualityUnsigned           QualityCode = "unsigned"
	QualityInterpolated       QualityCode = "interpolated"
	QualityManuallyOverridden QualityCode = "manually_overridden"
	QualityOutOfRange         QualityCode = "out_of_range"
	QualityDeviceError        QualityCode = "device_error"
	QualityTimestampAnomaly   QualityCode = "timestamp_anomaly"
	QualitySequenceGap        QualityCode = "sequence_gap"
)

// Reading is the canonical normalised reading that flows from Protocol Packs
// into Core. Rule 105 prescribes the meter HMAC over (meter_id, channel_id,
// ts_ns, value_int_micro_unit); Phase F Sprint S11 wires verification at
// ingestion.
type Reading struct {
	TenantID       uuid.UUID   `json:"tenant_id"`
	MeterID        uuid.UUID   `json:"meter_id"`
	ChannelID      uuid.UUID   `json:"channel_id"`
	Timestamp      time.Time   `json:"ts"`
	Value          int64       `json:"value_micro_unit"`
	Unit           string      `json:"unit"`
	QualityCode    QualityCode `json:"quality_code"`
	RawPayload     []byte      `json:"raw_payload,omitempty"`
	MeterSignature []byte      `json:"meter_hmac,omitempty"`
}

// ReadingBatch is a transport-level grouping of Readings. Protocol Packs
// emit batches sized for amortised cost; the runner has no opinion on size.
type ReadingBatch struct {
	Readings []Reading `json:"readings"`
}

// ReadingSink is the contract the IngestorRunner exposes to Protocol Packs
// for delivery. The Sink is bounded (Rule 119): backpressure surfaces as
// blocking on Send when the bound is reached.
type ReadingSink interface {
	// Send delivers a batch into Core. A non-nil error indicates either a
	// transient failure (retry-eligible) or a permanent failure (drop +
	// counter increment). Implementations of Ingestor SHOULD honour the
	// ctx deadline.
	Send(ctx context.Context, batch ReadingBatch) error
}

// Ingestor is the Pack-contract for protocol implementations.
//
// Lifecycle, in order:
//  1. Loader instantiates the Pack and calls Pack.Init.
//  2. Pack registers an Ingestor via Registrar.RegisterIngestor.
//  3. IngestorRunner calls Ingestor.Start(ctx, sink); the goroutine runs
//     until ctx is cancelled or Stop is called.
//  4. On graceful shutdown the runner cancels ctx; if Stop returns within
//     the shutdown budget (≤ 30s per Rule 42), the Pack is considered
//     cleanly drained.
//
// Implementations MUST be safe for concurrent calls to Stop. Implementations
// MUST honour the ctx and the shutdown budget.
type Ingestor interface {
	// Name returns a stable identifier for this Ingestor instance, used in
	// logs, traces, and metrics. Convention: "<pack-id>:<instance-key>"
	// (e.g. "modbus_tcp:plant-a-cabin-1").
	Name() string

	// Start begins reading from the protocol and pushing to the sink. It
	// runs until ctx is cancelled or returns an error. Errors are
	// terminal — the runner does not auto-restart; runbooks govern restart.
	Start(ctx context.Context, sink ReadingSink) error

	// Stop is invoked by the runner during graceful shutdown. It must
	// return within the runner's shutdown budget. Subsequent Send calls
	// after Stop returns are forbidden.
	Stop(ctx context.Context) error
}
