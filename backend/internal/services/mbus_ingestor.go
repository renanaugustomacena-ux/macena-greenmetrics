package services

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"go.uber.org/zap"
)

// MBusIngestor is a placeholder M-Bus (EN 13757) master implementation.
//
// Real production stack options:
//   - libmbus (C library) via CGo — most complete but ties us to C toolchain.
//   - mbus-go (pure Go subset) — lightweight, PRIMARY ADDRESS + REQ_UD2 support.
//   - External gateway (e.g. Elvaco CMe) that already speaks REST/MQTT — preferred
//     in field deployments; GreenMetrics consumes its JSON via the generic HTTP
//     ingestor instead of implementing M-Bus directly.
type MBusIngestor struct {
	logger  *zap.Logger
	device  string
	baud    int
	timeout time.Duration
}

// NewMBusIngestor constructs the ingestor.
func NewMBusIngestor(logger *zap.Logger, device string, baud int, timeout time.Duration) *MBusIngestor {
	return &MBusIngestor{logger: logger, device: device, baud: baud, timeout: timeout}
}

// MBusFrame is the decoded REQ_UD2 response.
type MBusFrame struct {
	PrimaryAddr   byte
	Manufacturer  string
	MeterID       uint32
	Version       byte
	Medium        string
	DataRecords   []MBusDataRecord
}

// MBusDataRecord is a single VIF/DIF parsed value.
type MBusDataRecord struct {
	Tag   string
	Unit  string
	Value float64
}

// ReadREQ_UD2 issues a short-frame REQ_UD2 to the given primary address and
// parses a single long-frame response. Stub implementation returns ErrNotConfigured
// on empty device string; a CI-friendly loopback scenario would hand back a
// synthetic frame for unit tests.
func (i *MBusIngestor) ReadREQ_UD2(ctx context.Context, primaryAddr byte) (*MBusFrame, error) {
	if i.device == "" {
		return nil, ErrNotConfigured
	}
	// Real path (placeholder):
	//   - open serial port at baud (2400 / 9600 / 19200).
	//   - send SHORT FRAME: 0x10 | C=0x5B (REQ_UD2) | primaryAddr | checksum | 0x16.
	//   - read LONG FRAME starting with 0x68 LEN LEN 0x68 ... CRC 0x16.
	//   - decode CI-field, DIF/VIF chain.
	i.logger.Info("mbus REQ_UD2 stub",
		zap.Uint8("addr", primaryAddr),
		zap.String("device", i.device),
		zap.String("note", "stub returns synthetic frame"),
	)
	return &MBusFrame{
		PrimaryAddr:  primaryAddr,
		Manufacturer: "ELV",
		MeterID:      12345678,
		Medium:       "gas",
		DataRecords: []MBusDataRecord{
			{Tag: "Volume", Unit: "Sm3", Value: 123.456},
			{Tag: "Flow", Unit: "Sm3/h", Value: 0.42},
			{Tag: "Temperature", Unit: "degC", Value: 21.3},
		},
	}, nil
}

// ErrNotConfigured is returned when the serial device is not wired up.
var ErrNotConfigured = errors.New("mbus: serial device not configured")

// EncodeShortFrame is a helper that produces the bytes of a REQ_UD2 short frame.
func EncodeShortFrame(cField, addr byte) []byte {
	checksum := cField + addr
	frame := []byte{0x10, cField, addr, checksum, 0x16}
	return frame
}

// Hex returns a debug hex string for a frame.
func Hex(b []byte) string { return hex.EncodeToString(b) }
