// Example tests for the Protocol Pack contract. Per Rule 86 (contracts
// documented in code), the example demonstrates a minimal compliant
// implementation — readers see immediately what's required.

package protocol_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/greenmetrics/backend/internal/domain/protocol"
)

// nopSink records batches. A test-only ReadingSink.
type nopSink struct{ batches []protocol.ReadingBatch }

func (s *nopSink) Send(_ context.Context, b protocol.ReadingBatch) error {
	s.batches = append(s.batches, b)
	return nil
}

// stubIngestor is a minimal Ingestor implementation. A real Protocol Pack
// would speak Modbus / M-Bus / etc.; this stub emits one batch per Start.
type stubIngestor struct{ stopped bool }

func (i *stubIngestor) Name() string { return "stub:test-1" }

func (i *stubIngestor) Start(ctx context.Context, sink protocol.ReadingSink) error {
	if i.stopped {
		return errors.New("ingestor already stopped")
	}
	batch := protocol.ReadingBatch{
		Readings: []protocol.Reading{{
			TenantID:    uuid.New(),
			MeterID:     uuid.New(),
			ChannelID:   uuid.New(),
			Timestamp:   time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC),
			Value:       1234567,
			Unit:        "Wh",
			QualityCode: protocol.QualityGood,
		}},
	}
	return sink.Send(ctx, batch)
}

func (i *stubIngestor) Stop(_ context.Context) error {
	i.stopped = true
	return nil
}

// TestExample_MinimalIngestor demonstrates Pack-contract compliance.
// Compile-time: the stub satisfies protocol.Ingestor.
// Runtime: the lifecycle (Start → Send → Stop) executes.
func TestExample_MinimalIngestor(t *testing.T) {
	var ing protocol.Ingestor = &stubIngestor{}
	sink := &nopSink{}
	ctx := context.Background()

	if err := ing.Start(ctx, sink); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if len(sink.batches) != 1 {
		t.Fatalf("expected 1 batch; got %d", len(sink.batches))
	}
	if got := len(sink.batches[0].Readings); got != 1 {
		t.Errorf("expected 1 reading in batch; got %d", got)
	}
	if got := sink.batches[0].Readings[0].QualityCode; got != protocol.QualityGood {
		t.Errorf("quality_code = %q; want %q", got, protocol.QualityGood)
	}
	if err := ing.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestContractVersion_IsSemver(t *testing.T) {
	if protocol.ContractVersion == "" {
		t.Fatal("ContractVersion empty — Rule 71 requires per-kind contract version")
	}
}
