package models

import "time"

// MeterType enumerates supported meter archetypes.
type MeterType string

const (
	MeterTypeElectricity       MeterType = "electricity"
	MeterTypeElectricityPhase  MeterType = "electricity_3p"
	MeterTypeGas               MeterType = "gas"
	MeterTypeWater             MeterType = "water"
	MeterTypeThermal           MeterType = "thermal"
	MeterTypePVInverter        MeterType = "pv_inverter"
	MeterTypeEVCharger         MeterType = "ev_charger"
	MeterTypePulseCounter      MeterType = "pulse_counter"
)

// MeterProtocol enumerates supported physical-layer / application-layer protocols.
type MeterProtocol string

const (
	ProtocolModbusRTU  MeterProtocol = "modbus_rtu"
	ProtocolModbusTCP  MeterProtocol = "modbus_tcp"
	ProtocolMBus       MeterProtocol = "mbus"
	ProtocolSunSpec    MeterProtocol = "sunspec"
	ProtocolOCPP       MeterProtocol = "ocpp"
	ProtocolPulse      MeterProtocol = "pulse"
	ProtocolManual     MeterProtocol = "manual"
	ProtocolSPD        MeterProtocol = "spd"
)

// Meter represents a physical measurement point.
type Meter struct {
	ID          string        `json:"id"`
	TenantID    string        `json:"tenant_id"`
	Label       string        `json:"label"`
	MeterType   MeterType     `json:"meter_type"`
	Protocol    MeterProtocol `json:"protocol"`
	Unit        string        `json:"unit"`
	Site        string        `json:"site"`
	CostCentre  string        `json:"cost_centre,omitempty"`
	SerialNo    string        `json:"serial_no,omitempty"`
	PODCode     string        `json:"pod_code,omitempty"`   // Electricity POD
	PDRCode     string        `json:"pdr_code,omitempty"`   // Gas PDR
	Endpoint    string        `json:"endpoint,omitempty"`   // host:port for IP-based protocols
	SlaveAddr   int           `json:"slave_addr,omitempty"` // Modbus slave / M-Bus primary addr
	Active      bool          `json:"active"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// MeterChannel is a sub-measurement on a compound meter (e.g. three phases).
type MeterChannel struct {
	ID          string  `json:"id"`
	MeterID     string  `json:"meter_id"`
	ChannelCode string  `json:"channel_code"` // kWh, kVArh, voltage_L1, etc.
	Unit        string  `json:"unit"`
	ScaleFactor float64 `json:"scale_factor"`
	Description string  `json:"description,omitempty"`
}
