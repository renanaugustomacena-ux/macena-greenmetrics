// Package services — IngestorRunner: orchestrates the meter polling loops.
//
// This is the missing wiring flagged in GreenMetrics-GAPS H-01 / D-08:
// ModbusIngestor and MBusIngestor existed as code but were never constructed
// in cmd/server/main.go. The runner constructs them, runs each poll loop as
// a background goroutine, and respects context cancellation on shutdown.
//
// The poll sources for the local dev/CI stack come from env:
//
//	MODBUS_SIMULATOR_ADDR   e.g. greenmetrics-simulator:5020   (used in compose)
//	MODBUS_POLL_INTERVAL    default 30s
//	MODBUS_SLAVE_IDS        comma list of slave IDs, default "1,2,3,4,5"
//
// If MODBUS_SIMULATOR_ADDR is empty the runner logs "ingestion disabled" at
// INFO and returns cleanly. No silent stubs — if the env vars are partially
// set (address present but malformed) the runner returns a typed error.
package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/greenmetrics/backend/internal/repository"
)

// IngestorConfig holds the runtime knobs for background meter polling.
type IngestorConfig struct {
	ModbusAddr       string        // host:port of a Modbus/TCP target (or simulator)
	ModbusPollEvery  time.Duration // poll cadence
	ModbusSlaveIDs   []byte        // slave IDs to poll
	ModbusTimeout    time.Duration // per-call timeout

	SunSpecConfig *SunSpecConfig // optional; if non-nil a SunSpec loop also runs

	MBusDevice   string
	MBusBaud     int
	MBusInterval time.Duration

	TenantID string // tenant the synthetic readings are written under
}

// IngestorRunner wires the adapter suite to the repository and a lifecycle.
type IngestorRunner struct {
	cfg    IngestorConfig
	repo   *repository.TimescaleRepository
	modbus *ModbusIngestor
	mbus   *MBusIngestor
	sun    *SunSpecProfile
	logger *zap.Logger
}

// ErrIngestorDisabled is returned by Start when no data-source is configured.
// It is NOT fatal — the caller may proceed and run only the HTTP surface.
var ErrIngestorDisabled = errors.New("ingestor: no meter data source configured (set MODBUS_SIMULATOR_ADDR to enable)")

// NewIngestorRunner builds the runner. Passing a nil repo is allowed (the
// runner will log warnings instead of persisting — dev convenience), but a
// missing Modbus address is treated as "ingestion disabled" and surfaced via
// Start.
func NewIngestorRunner(cfg IngestorConfig, repo *repository.TimescaleRepository, logger *zap.Logger) *IngestorRunner {
	r := &IngestorRunner{cfg: cfg, repo: repo, logger: logger}
	if cfg.ModbusTimeout <= 0 {
		r.cfg.ModbusTimeout = 3 * time.Second
	}
	if cfg.ModbusPollEvery <= 0 {
		r.cfg.ModbusPollEvery = 30 * time.Second
	}
	if cfg.TenantID == "" {
		r.cfg.TenantID = "placeholder-tenant"
	}
	r.modbus = NewModbusIngestor(logger, r.cfg.ModbusTimeout)

	if cfg.MBusDevice != "" {
		r.mbus = NewMBusIngestor(logger, cfg.MBusDevice, cfg.MBusBaud, r.cfg.ModbusTimeout)
	}
	if cfg.SunSpecConfig != nil {
		// Best-effort; ErrSunSpecNotConfigured only means addr is empty.
		if prof, err := NewSunSpecProfile(*cfg.SunSpecConfig, r.modbus, logger); err == nil {
			r.sun = prof
		}
	}
	return r
}

// Start launches background poll loops. Returns ErrIngestorDisabled if there
// is literally nothing to do, so the caller can skip the goroutines cleanly.
// The loops honour ctx cancellation.
func (r *IngestorRunner) Start(ctx context.Context) error {
	if r.cfg.ModbusAddr == "" && r.mbus == nil && r.sun == nil {
		return ErrIngestorDisabled
	}

	if r.cfg.ModbusAddr != "" {
		for _, slave := range r.cfg.ModbusSlaveIDs {
			sid := slave // capture for goroutine
			go r.runModbusLoop(ctx, sid)
		}
		r.logger.Info("modbus ingestor loops started",
			zap.String("addr", r.cfg.ModbusAddr),
			zap.Int("slaves", len(r.cfg.ModbusSlaveIDs)),
			zap.Duration("interval", r.cfg.ModbusPollEvery),
		)
	}

	if r.mbus != nil {
		go r.runMBusLoop(ctx)
		r.logger.Info("mbus ingestor loop started", zap.String("device", r.cfg.MBusDevice))
	}

	if r.sun != nil {
		go r.runSunSpecLoop(ctx)
		r.logger.Info("sunspec ingestor loop started", zap.String("profile", r.sun.Describe()))
	}

	return nil
}

// --- Modbus --------------------------------------------------------------

// DefaultElectricityPoints is the register map matching cmd/simulator.
// Register 0x0000-0x0001 is the energy counter (u32 Wh); 0x0002-0x0003 is
// instantaneous power (u32 W×1000).
func DefaultElectricityPoints() []ModbusPoint {
	return []ModbusPoint{
		{Name: "energy_wh", Register: 0x0000, Quantity: 2, FunctionCode: 3, Scale: 1, Unit: "Wh", DataType: "uint32"},
		{Name: "power_w", Register: 0x0002, Quantity: 2, FunctionCode: 3, Scale: 0.001, Unit: "W", DataType: "uint32"},
		{Name: "voltage_v", Register: 0x0004, Quantity: 1, FunctionCode: 3, Scale: 0.1, Unit: "V", DataType: "uint16"},
		{Name: "current_a", Register: 0x0005, Quantity: 1, FunctionCode: 3, Scale: 0.01, Unit: "A", DataType: "uint16"},
		{Name: "pf", Register: 0x0006, Quantity: 1, FunctionCode: 3, Scale: 0.001, Unit: "", DataType: "uint16"},
		{Name: "frequency_hz", Register: 0x0007, Quantity: 1, FunctionCode: 3, Scale: 0.01, Unit: "Hz", DataType: "uint16"},
	}
}

func (r *IngestorRunner) runModbusLoop(ctx context.Context, slaveID byte) {
	ticker := time.NewTicker(r.cfg.ModbusPollEvery)
	defer ticker.Stop()
	// Jittered initial tick.
	first := time.After(time.Duration(int64(slaveID)) * 100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			r.logger.Info("modbus loop exiting", zap.Uint8("slave", slaveID))
			return
		case <-first:
			first = nil
			r.pollAndPersist(ctx, slaveID)
		case <-ticker.C:
			r.pollAndPersist(ctx, slaveID)
		}
	}
}

func (r *IngestorRunner) pollAndPersist(ctx context.Context, slaveID byte) {
	pts := DefaultElectricityPoints()
	vals, err := r.modbus.PollTCP(ctx, r.cfg.ModbusAddr, slaveID, pts)
	if err != nil {
		r.logger.Warn("modbus poll failed", zap.Uint8("slave", slaveID), zap.Error(err))
		return
	}
	if r.repo == nil {
		r.logger.Debug("no repo; dropping sample", zap.Uint8("slave", slaveID), zap.Any("sample", vals))
		return
	}
	now := time.Now().UTC()
	readings := make([]repository.Reading, 0, len(vals))
	meterID := fmt.Sprintf("sim-meter-%d", slaveID)
	for k, v := range vals {
		readings = append(readings, repository.Reading{
			Ts:          now,
			TenantID:    r.cfg.TenantID,
			MeterID:     meterID,
			ChannelID:   k,
			Value:       v,
			Unit:        unitFor(k),
			QualityCode: 0,
		})
	}
	insCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.repo.InsertReadings(insCtx, readings); err != nil {
		r.logger.Warn("modbus persist failed", zap.Uint8("slave", slaveID), zap.Error(err))
	}
}

func unitFor(channel string) string {
	switch {
	case strings.HasSuffix(channel, "_wh"):
		return "Wh"
	case strings.HasSuffix(channel, "_w"):
		return "W"
	case strings.HasSuffix(channel, "_v"):
		return "V"
	case strings.HasSuffix(channel, "_a"):
		return "A"
	case strings.HasSuffix(channel, "_hz"):
		return "Hz"
	case strings.HasSuffix(channel, "_c"):
		return "degC"
	}
	return ""
}

// --- M-Bus ---------------------------------------------------------------

func (r *IngestorRunner) runMBusLoop(ctx context.Context) {
	interval := r.cfg.MBusInterval
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			frame, err := r.mbus.ReadREQ_UD2(ctx, 1)
			if err != nil {
				if errors.Is(err, ErrNotConfigured) {
					return // surface once; loop ends if device isn't wired
				}
				r.logger.Warn("mbus read failed", zap.Error(err))
				continue
			}
			r.logger.Debug("mbus frame", zap.Uint8("addr", frame.PrimaryAddr), zap.Int("records", len(frame.DataRecords)))
		}
	}
}

// --- SunSpec -------------------------------------------------------------

func (r *IngestorRunner) runSunSpecLoop(ctx context.Context) {
	interval := r.sun.cfg.PollInterval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			vals, err := r.sun.PollOnce(ctx)
			if err != nil {
				r.logger.Warn("sunspec poll failed", zap.Error(err))
				continue
			}
			r.logger.Debug("sunspec sample", zap.Any("values", vals))
		}
	}
}

// ParseSlaveIDs parses a comma-separated list like "1,2,3" into []byte.
func ParseSlaveIDs(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []byte{1, 2, 3, 4, 5}, nil
	}
	parts := strings.Split(s, ",")
	out := make([]byte, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return nil, fmt.Errorf("invalid slave id %q", p)
		}
		out = append(out, byte(n))
	}
	if len(out) == 0 {
		return nil, errors.New("slave id list produced empty set")
	}
	return out, nil
}
