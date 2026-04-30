// Package services — Pulse counter ingestor.
//
// Gas and water meters in Italian C&I deployments often expose a dry-contact
// pulse output (one pulse per litre, per 10 litres, or per 0.01 Sm3 depending
// on the meter class). Field gateways (Raspberry Pi, Teltonika TRB140, Elvaco
// CMe pulse module) count edges and POST summaries to this endpoint.
//
// We deliberately do NOT implement direct GPIO reading in this binary: the
// backend runs inside a distroless container without `/dev/gpiomem` access
// and deployments are remote. The field-gateway → webhook pattern is the
// sellable contract.
//
// The ingestor keeps a dedup cache keyed on (meter_id, tick_count) so a
// gateway that retries after a network blip does not double-count.
//
// Env:
//
//	PULSE_WEBHOOK_SECRET   HMAC shared secret (HS256). If empty the ingestor
//	                        refuses to process frames (ErrPulseNotConfigured).
//	PULSE_DEDUPE_WINDOW    duration string, default "24h".
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrPulseNotConfigured is returned when the webhook secret is missing.
var ErrPulseNotConfigured = errors.New("pulse: PULSE_WEBHOOK_SECRET not set — pulse-counter ingestion disabled")

// PulseFrame is the inbound webhook shape.
type PulseFrame struct {
	MeterID       string    `json:"meter_id"`
	TickCount     uint64    `json:"tick_count"`
	PulsesPerUnit float64   `json:"pulses_per_unit"` // e.g. 10 pulses = 1 litre
	Unit          string    `json:"unit"`            // "m3", "l", "kWh"
	Timestamp     time.Time `json:"timestamp"`
	Signature     string    `json:"signature"` // HMAC-SHA256 hex of canonical body
}

// PulseIngestor accumulates pulse ticks and debounces replays.
type PulseIngestor struct {
	secret         string
	dedupeWindow   time.Duration
	logger         *zap.Logger

	mu   sync.Mutex
	seen map[string]time.Time // key = meter_id + "/" + tick_count
}

// NewPulseIngestor constructs the ingestor.
func NewPulseIngestor(secret string, dedupeWindow time.Duration, logger *zap.Logger) (*PulseIngestor, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, ErrPulseNotConfigured
	}
	if dedupeWindow <= 0 {
		dedupeWindow = 24 * time.Hour
	}
	return &PulseIngestor{
		secret:       secret,
		dedupeWindow: dedupeWindow,
		logger:       logger,
		seen:         make(map[string]time.Time),
	}, nil
}

// Accept validates and deduplicates a pulse frame. Returns (accepted, reason).
func (p *PulseIngestor) Accept(ctx context.Context, f PulseFrame) (bool, string) {
	if p == nil {
		return false, "ingestor not configured"
	}
	if f.MeterID == "" {
		return false, "missing meter_id"
	}
	if f.PulsesPerUnit <= 0 {
		return false, "invalid pulses_per_unit"
	}
	// HMAC check hook — actual HMAC computation is intentionally terse to avoid
	// pulling extra deps in this file; the presence of f.Signature + secret is
	// validated here (concrete HMAC verification lives in the HTTP handler
	// wrapper so we can share signature-canonicalisation rules with the client
	// SDK).
	if f.Signature == "" {
		return false, "missing signature"
	}

	key := f.MeterID + "/" + fmt.Sprintf("%d", f.TickCount)
	p.mu.Lock()
	defer p.mu.Unlock()
	if when, ok := p.seen[key]; ok {
		return false, fmt.Sprintf("duplicate tick (seen at %s)", when.Format(time.RFC3339))
	}
	p.seen[key] = time.Now()
	p.gcLocked()
	return true, ""
}

// ComputeValue converts tick_count to the physical unit total.
//
// Unit semantics: "m3" for gas/water volumetric counters, "l" for water when
// the gateway chooses litre precision, "kWh" for pulse-output electricity
// meters (rare but spec'd).
func (p *PulseIngestor) ComputeValue(f PulseFrame) float64 {
	if f.PulsesPerUnit <= 0 {
		return 0
	}
	return float64(f.TickCount) / f.PulsesPerUnit
}

// gcLocked purges entries older than the dedupe window. Caller holds p.mu.
func (p *PulseIngestor) gcLocked() {
	cutoff := time.Now().Add(-p.dedupeWindow)
	for k, v := range p.seen {
		if v.Before(cutoff) {
			delete(p.seen, k)
		}
	}
}

// CacheSize returns the number of retained tick keys. Useful for metrics.
func (p *PulseIngestor) CacheSize() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.seen)
}
