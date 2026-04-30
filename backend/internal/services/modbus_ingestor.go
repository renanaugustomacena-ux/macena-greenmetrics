package services

import (
	"context"
	"fmt"
	"time"

	"github.com/goburrow/modbus"
	"go.uber.org/zap"
)

// ModbusIngestor polls Modbus RTU (serial) and Modbus TCP devices.
//
// Typical register maps:
//   - Electricity meters (e.g. Carlo Gavazzi EM24, Socomec A40, Lovato DMG): input
//     registers 0x0000+ for V/A/W, holding registers for energy counters.
//   - PV inverters: SunSpec common model (starts at 40001 / 50001) — we treat that
//     separately in smart_meter_client (not here).
type ModbusIngestor struct {
	logger *zap.Logger
	timeout time.Duration
}

// NewModbusIngestor constructs the ingestor with a timeout.
func NewModbusIngestor(logger *zap.Logger, timeout time.Duration) *ModbusIngestor {
	return &ModbusIngestor{logger: logger, timeout: timeout}
}

// ModbusPoint describes one value we want to read from a device.
type ModbusPoint struct {
	Name         string  `json:"name"`
	Register     uint16  `json:"register"`
	Quantity     uint16  `json:"quantity"` // words to read
	FunctionCode uint8   `json:"function_code"` // 3 = holding, 4 = input
	Scale        float64 `json:"scale"`         // applied after decoding
	Unit         string  `json:"unit"`
	DataType     string  `json:"data_type"` // "int16"|"uint16"|"int32"|"uint32"|"float32"
}

// PollTCP reads a set of points from a Modbus TCP slave.
func (i *ModbusIngestor) PollTCP(ctx context.Context, addr string, slaveID byte, points []ModbusPoint) (map[string]float64, error) {
	handler := modbus.NewTCPClientHandler(addr)
	handler.Timeout = i.timeout
	handler.SlaveId = slaveID
	if err := handler.Connect(); err != nil {
		return nil, fmt.Errorf("modbus connect %s: %w", addr, err)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)
	return i.readPoints(client, points)
}

// PollRTU reads via a serial handler (ttyUSB / ttyS).
func (i *ModbusIngestor) PollRTU(ctx context.Context, device string, baud int, slaveID byte, points []ModbusPoint) (map[string]float64, error) {
	handler := modbus.NewRTUClientHandler(device)
	handler.BaudRate = baud
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = slaveID
	handler.Timeout = i.timeout
	if err := handler.Connect(); err != nil {
		return nil, fmt.Errorf("rtu connect %s: %w", device, err)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)
	return i.readPoints(client, points)
}

func (i *ModbusIngestor) readPoints(client modbus.Client, points []ModbusPoint) (map[string]float64, error) {
	out := map[string]float64{}
	for _, p := range points {
		var raw []byte
		var err error
		switch p.FunctionCode {
		case 3:
			raw, err = client.ReadHoldingRegisters(p.Register, p.Quantity)
		case 4:
			raw, err = client.ReadInputRegisters(p.Register, p.Quantity)
		default:
			err = fmt.Errorf("unsupported function code %d", p.FunctionCode)
		}
		if err != nil {
			i.logger.Warn("modbus read failed", zap.String("point", p.Name), zap.Error(err))
			continue
		}
		out[p.Name] = decode(raw, p)
	}
	return out, nil
}

func decode(raw []byte, p ModbusPoint) float64 {
	if len(raw) == 0 {
		return 0
	}
	switch p.DataType {
	case "uint16":
		if len(raw) < 2 {
			return 0
		}
		v := uint16(raw[0])<<8 | uint16(raw[1])
		return float64(v) * p.Scale
	case "int16":
		if len(raw) < 2 {
			return 0
		}
		v := int16(uint16(raw[0])<<8 | uint16(raw[1]))
		return float64(v) * p.Scale
	case "uint32":
		if len(raw) < 4 {
			return 0
		}
		v := uint32(raw[0])<<24 | uint32(raw[1])<<16 | uint32(raw[2])<<8 | uint32(raw[3])
		return float64(v) * p.Scale
	case "int32":
		if len(raw) < 4 {
			return 0
		}
		v := int32(uint32(raw[0])<<24 | uint32(raw[1])<<16 | uint32(raw[2])<<8 | uint32(raw[3]))
		return float64(v) * p.Scale
	case "float32":
		if len(raw) < 4 {
			return 0
		}
		// IEEE-754 big-endian.
		bits := uint32(raw[0])<<24 | uint32(raw[1])<<16 | uint32(raw[2])<<8 | uint32(raw[3])
		return float64Bits32ToFloat(bits) * p.Scale
	}
	return 0
}

// float64Bits32ToFloat converts a uint32 IEEE-754 representation to float64.
func float64Bits32ToFloat(u uint32) float64 {
	// Replicated locally to avoid depending on math bits package aliases.
	if u == 0 {
		return 0
	}
	sign := 1.0
	if u>>31 == 1 {
		sign = -1.0
	}
	exp := int((u >> 23) & 0xff)
	frac := u & 0x7fffff
	if exp == 0 && frac == 0 {
		return 0
	}
	mantissa := float64(frac)/float64(1<<23) + 1
	return sign * mantissa * pow2(exp-127)
}

func pow2(e int) float64 {
	if e == 0 {
		return 1
	}
	if e > 0 {
		return float64(uint64(1) << uint(e))
	}
	return 1.0 / float64(uint64(1)<<uint(-e))
}
