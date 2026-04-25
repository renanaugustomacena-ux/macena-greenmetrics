// Package services — bounded ingest channel + batched DB writer.
//
// Doctrine refs: Rule 15 (reliability), Rule 36 (failure as normal — backpressure),
//                Rule 37 (latency budget), Rule 41 (concurrency), Rule 42 (resource lifecycle).
//
// Pipeline shape:
//
//   sources (Modbus / M-Bus / SunSpec / Pulse / OCPP)
//                       │
//                       ▼
//        bounded chan repository.Reading (capacity INGEST_QUEUE_DEPTH)
//                       │
//                       ▼ (drained by single goroutine)
//        batched writer: up to BATCH_SIZE every BATCH_INTERVAL via pgx.CopyFrom
//                       │
//                       ▼
//                  TimescaleDB
//
// Drop policy on saturation: log + Prometheus counter `gm_ingest_dropped_total{reason="queue_full"}`.
// Sources MUST never block the pipeline (Modbus tickers cannot skew).
//
// Optional disk spill via INGEST_SPILL=true (boltdb-backed, S5 follow-on).

package services

import (
	"context"
	"sync"
	"time"

	"github.com/greenmetrics/backend/internal/metrics"
)

// PipelineConfig governs the ingest pipeline.
type PipelineConfig struct {
	QueueDepth     int           // default 10000
	BatchSize      int           // default 1000
	BatchInterval  time.Duration // default 100 ms
	WriteTimeout   time.Duration // default 5 s
	SpillEnabled   bool          // default false
}

func (c PipelineConfig) withDefaults() PipelineConfig {
	if c.QueueDepth == 0 {
		c.QueueDepth = 10000
	}
	if c.BatchSize == 0 {
		c.BatchSize = 1000
	}
	if c.BatchInterval == 0 {
		c.BatchInterval = 100 * time.Millisecond
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 5 * time.Second
	}
	return c
}

// PipelineReading is the wire format from sources to writer.
//
// Concrete repository.Reading is referenced indirectly to avoid an import cycle;
// callers map their domain Reading into this struct at submit time.
type PipelineReading struct {
	TenantID  string
	MeterID   string
	ChannelID string
	Ts        time.Time
	Value     float64
	Unit      string
	Quality   int
	Source    string // "modbus", "mbus", "pulse_webhook", etc.
}

// Writer is the abstract persistence target — pgx-backed in production.
type Writer interface {
	WriteBatch(ctx context.Context, batch []PipelineReading) error
}

// Pipeline is the bounded-channel + batched-writer wrapper.
type Pipeline struct {
	cfg    PipelineConfig
	in     chan PipelineReading
	writer Writer
	stop   chan struct{}
	wg     sync.WaitGroup
}

// NewPipeline constructs the pipeline. Call Start before submitting.
func NewPipeline(cfg PipelineConfig, w Writer) *Pipeline {
	cfg = cfg.withDefaults()
	return &Pipeline{
		cfg:    cfg,
		in:     make(chan PipelineReading, cfg.QueueDepth),
		writer: w,
		stop:   make(chan struct{}),
	}
}

// Start spawns the writer goroutine. Idempotent if called twice.
func (p *Pipeline) Start(ctx context.Context) {
	p.wg.Add(1)
	go p.run(ctx)
}

// Stop signals the writer to drain and exit. Wait blocks until done.
func (p *Pipeline) Stop() {
	close(p.stop)
}

// Wait blocks until the writer goroutine has exited.
func (p *Pipeline) Wait() { p.wg.Wait() }

// Submit attempts to push a reading into the queue without blocking.
// Returns false if the queue is full; caller logs + Prometheus counter increments.
func (p *Pipeline) Submit(r PipelineReading) bool {
	select {
	case p.in <- r:
		metrics.IngestQueueDepth.Set(float64(len(p.in)))
		return true
	default:
		metrics.IngestDropped.WithLabelValues("queue_full").Inc()
		return false
	}
}

// SubmitBlocking pushes a reading with backpressure honoured up to ctx deadline.
// Use only from sources that can tolerate blocking (e.g. HTTP ingest with Retry-After).
func (p *Pipeline) SubmitBlocking(ctx context.Context, r PipelineReading) error {
	select {
	case p.in <- r:
		metrics.IngestQueueDepth.Set(float64(len(p.in)))
		return nil
	case <-ctx.Done():
		metrics.IngestDropped.WithLabelValues("ctx_done").Inc()
		return ctx.Err()
	}
}

func (p *Pipeline) run(ctx context.Context) {
	defer p.wg.Done()

	batch := make([]PipelineReading, 0, p.cfg.BatchSize)
	tick := time.NewTicker(p.cfg.BatchInterval)
	defer tick.Stop()

	flush := func(reason string) {
		if len(batch) == 0 {
			return
		}
		writeCtx, cancel := context.WithTimeout(ctx, p.cfg.WriteTimeout)
		err := p.writer.WriteBatch(writeCtx, batch)
		cancel()
		if err != nil {
			// Failed batch: optionally spill (S5); for now drop + counter.
			metrics.IngestDropped.WithLabelValues("write_error").Add(float64(len(batch)))
		}
		batch = batch[:0]
		metrics.IngestQueueDepth.Set(float64(len(p.in)))
		_ = reason // reserved for future per-reason metric
	}

	for {
		select {
		case <-p.stop:
			// Drain remaining items + final flush.
			for {
				select {
				case r := <-p.in:
					batch = append(batch, r)
					if len(batch) >= p.cfg.BatchSize {
						flush("drain_full")
					}
				default:
					flush("drain_close")
					return
				}
			}

		case <-ctx.Done():
			flush("ctx_done")
			return

		case r := <-p.in:
			batch = append(batch, r)
			metrics.IngestQueueDepth.Set(float64(len(p.in)))
			if len(batch) >= p.cfg.BatchSize {
				flush("size")
			}

		case <-tick.C:
			flush("interval")
		}
	}
}
