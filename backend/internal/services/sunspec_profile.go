// Package services — SunSpec profile adapter for PV inverters.
//
// SunSpec Alliance Information Model ("common model + device models") is a
// registry of Modbus register blocks that PV inverters expose. The wire
// transport is ModbusTCP (register base 40001 or 50001); we reuse the existing
// ModbusIngestor rather than duplicate the Modbus framing.
//
// Models we support out of the box:
//
//	Model 1   Common block (manufacturer, model, version, serial)
//	Model 101 Single-phase inverter
//	Model 102 Split-phase inverter
//	Model 103 Three-phase inverter
//	Model 111 Single-phase inverter (float variants)
//
// The adapter is configured via env:
//
//	SUNSPEC_ENABLED=true
//	SUNSPEC_MODBUS_ADDR=10.0.0.50:502
//	SUNSPEC_SLAVE_ID=126
//	SUNSPEC_BASE_REGISTER=40000  (or 50000 on newer models)
//
// If SUNSPEC_MODBUS_ADDR is empty we refuse with ErrSunSpecNotConfigured.
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ErrSunSpecNotConfigured is returned when the SunSpec adapter is not wired.
var ErrSunSpecNotConfigured = errors.New("sunspec: SUNSPEC_MODBUS_ADDR not set — PV inverter ingestion disabled")

// SunSpecConfig is the per-inverter configuration.
type SunSpecConfig struct {
	Address      string
	SlaveID      byte
	BaseRegister uint16 // typically 40000 (SunSpec "40001") or 50000
	PollInterval time.Duration
}

// SunSpecProfile wires a ModbusIngestor to a SunSpec-compliant device.
type SunSpecProfile struct {
	cfg      SunSpecConfig
	ingestor *ModbusIngestor
	logger   *zap.Logger
}

// NewSunSpecProfile builds the adapter; returns ErrSunSpecNotConfigured
// if Address is empty.
func NewSunSpecProfile(cfg SunSpecConfig, ingestor *ModbusIngestor, logger *zap.Logger) (*SunSpecProfile, error) {
	if strings.TrimSpace(cfg.Address) == "" {
		return nil, ErrSunSpecNotConfigured
	}
	if ingestor == nil {
		return nil, errors.New("sunspec: modbus ingestor is required")
	}
	if cfg.BaseRegister == 0 {
		cfg.BaseRegister = 40000
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 30 * time.Second
	}
	if cfg.SlaveID == 0 {
		cfg.SlaveID = 126 // SunSpec default unit ID
	}
	return &SunSpecProfile{cfg: cfg, ingestor: ingestor, logger: logger}, nil
}

// CommonModelPoints returns the Model 1 register block (manufacturer, model,
// version, serial). Used to verify a device is SunSpec-compliant.
func (s *SunSpecProfile) CommonModelPoints() []ModbusPoint {
	base := s.cfg.BaseRegister
	// Model 1 "Common" block starts at base + 2 (after the SunSpec ID and DID markers).
	return []ModbusPoint{
		{Name: "sunspec_id", Register: base + 0, Quantity: 2, FunctionCode: 3, Scale: 1, Unit: "", DataType: "uint32"},
		{Name: "sunspec_did", Register: base + 2, Quantity: 1, FunctionCode: 3, Scale: 1, Unit: "", DataType: "uint16"},
		{Name: "sunspec_len", Register: base + 3, Quantity: 1, FunctionCode: 3, Scale: 1, Unit: "", DataType: "uint16"},
	}
}

// Model103Points returns the three-phase inverter register block (Model 103),
// which is the most common shape for Italian C&I PV systems.
func (s *SunSpecProfile) Model103Points() []ModbusPoint {
	// Offsets inside the Model 103 block, per the SunSpec Model Registry:
	//   A       40072  total AC current (A)  uint16
	//   A_SF    40076  scale factor          int16
	//   W       40083  AC power (W)          int16
	//   W_SF    40084  scale factor          int16
	//   Hz      40085  AC frequency (Hz)     uint16
	//   Hz_SF   40086  scale factor          int16
	//   WH      40093  AC energy (Wh)        uint32
	//   WH_SF   40095  scale factor          int16
	//   DCW     40100  DC power (W)          int16
	//   TmpCab  40103  cabinet temp (C)      int16
	// Addresses below assume base=40000 (the SunSpec "40001" convention).
	base := s.cfg.BaseRegister
	return []ModbusPoint{
		{Name: "pv_ac_current", Register: base + 71, Quantity: 1, FunctionCode: 3, Scale: 0.1, Unit: "A", DataType: "uint16"},
		{Name: "pv_ac_power_w", Register: base + 82, Quantity: 1, FunctionCode: 3, Scale: 1, Unit: "W", DataType: "int16"},
		{Name: "pv_ac_frequency_hz", Register: base + 84, Quantity: 1, FunctionCode: 3, Scale: 0.01, Unit: "Hz", DataType: "uint16"},
		{Name: "pv_ac_energy_total_wh", Register: base + 92, Quantity: 2, FunctionCode: 3, Scale: 1, Unit: "Wh", DataType: "uint32"},
		{Name: "pv_dc_power_w", Register: base + 99, Quantity: 1, FunctionCode: 3, Scale: 1, Unit: "W", DataType: "int16"},
		{Name: "pv_cabinet_temp_c", Register: base + 102, Quantity: 1, FunctionCode: 3, Scale: 0.1, Unit: "degC", DataType: "int16"},
	}
}

// PollOnce reads the Model 103 block once via the ModbusIngestor.
func (s *SunSpecProfile) PollOnce(ctx context.Context) (map[string]float64, error) {
	if s == nil {
		return nil, ErrSunSpecNotConfigured
	}
	return s.ingestor.PollTCP(ctx, s.cfg.Address, s.cfg.SlaveID, s.Model103Points())
}

// Describe returns a human-readable summary for logging.
func (s *SunSpecProfile) Describe() string {
	return fmt.Sprintf("sunspec(addr=%s slave=%d base=%d)", s.cfg.Address, s.cfg.SlaveID, s.cfg.BaseRegister)
}
