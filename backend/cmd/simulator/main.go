// Command simulator is a Modbus-TCP server producing realistic energy
// telemetry for 5 virtual meters used by the GreenMetrics dev and CI stack.
//
// Meters simulated (matches plan §5.9):
//
//	01 — LINE_A_ELEC         3-phase electricity (Carlo Gavazzi EM24 layout)
//	02 — LINE_B_ELEC         3-phase electricity (Socomec A40 layout)
//	03 — HVAC_ELEC           single-phase electricity
//	04 — BOILER_GAS          natural gas (Sm3)
//	05 — DISTRICT_HEAT       thermal (kWh-eq)
//
// Consumption pattern (Europe/Rome local clock):
//
//   - factory line ramps up 06:00, peak 08:00–18:00, ramps down 18:00–22:00;
//   - weekend (Sat/Sun) reduced to 15% of weekday peak — maintenance / cleaning;
//   - gas boiler tracks HDD (heating degree days) modulo daily minimum for DHW;
//   - HVAC peaks mid-day in summer, mornings in winter;
//   - district heat base-load + morning/evening bumps.
//
// Register map (convention for all meters):
//
//	HOLDING regs (function 3) at base 0x0000, 32-bit Big-Endian unsigned:
//	  0x0000  energy_counter_low        u32 — cumulative counter (Wh or mSm3 or Wh-eq)
//	  0x0002  instantaneous_power_low   u32 — present W / Sm3h*1000 / Wh-eq/h
//	  0x0004  voltage_v_scaled          u16 — V × 10 (electrical only)
//	  0x0005  current_a_scaled          u16 — A × 100
//	  0x0006  power_factor_scaled       u16 — cosφ × 1000
//	  0x0007  frequency_hz_scaled       u16 — Hz × 100
//	  0x0008  status_code               u16 — 0 = OK, bitfield for faults
//	  0x0009  timestamp_unix_low        u16 — seconds since epoch low word
//	  0x000A  timestamp_unix_high       u16 — seconds since epoch high word
//
// Listens on SIM_LISTEN (default 0.0.0.0:5020). Slave IDs 1..5.
//
// Hardening:
//   - each TCP connection has a 30s idle deadline (protocol timeouts bounded);
//   - malformed MBAP frames are dropped silently to avoid amplification;
//   - concurrent connections capped at 64; the 65th is refused.
//
// The simulator is designed to be fully deterministic when SIM_SEED is set,
// which makes the E2E test "1h simulation → Grafana trend" reproducible.
package main

import (
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	//nolint:gosec // G404: math/rand is intentional — SIM_SEED must produce reproducible traces; this is a telemetry simulator, not a crypto path.
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ----------------------------------------------------------------------------
// Constants / types.
// ----------------------------------------------------------------------------

const (
	// Modbus MBAP header length (transaction id 2, protocol 2, length 2, unit 1 = 7).
	mbapHeaderLen = 7
	// Max TCP connections held simultaneously.
	maxConns = 64
	// Per-connection idle deadline.
	connIdleDeadline = 30 * time.Second
)

// meterProfile describes a simulated meter.
type meterProfile struct {
	slaveID byte
	label   string
	kind    string // "electricity_3p" | "electricity" | "gas" | "thermal"
	// Base load and peak load (in meter native unit per hour):
	//   electricity: kW
	//   gas:         Sm3/h
	//   thermal:     kWh-eq/h
	baseLoad float64
	peakLoad float64
	// Process variance applied at each sample (log-normal multiplicative).
	processVariance float64
	// Voltage nominal (electrical only).
	voltageNominal float64
	// Rotating state (counter and last sample).
	mu             sync.RWMutex
	counter        float64 // cumulative Wh or mSm3 or Wh-eq
	lastSampleAt   time.Time
	lastInstantPwr float64
	lastVoltage    float64
	lastCurrent    float64
	lastPF         float64
	lastFreq       float64
	lastStatus     uint16
}

// ----------------------------------------------------------------------------
// Main / wiring.
// ----------------------------------------------------------------------------

func main() {
	// Support --healthcheck CLI flag used by Docker HEALTHCHECK (distroless has
	// no shell tools like `nc`, so the binary self-probes its listen port).
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		runHealthcheck()
		return
	}

	listen := getenv("SIM_LISTEN", "0.0.0.0:5020")
	tzName := getenv("SIM_TZ", "Europe/Rome")
	profile := getenv("SIM_PROFILE", "factory_line_verona")
	seedStr := getenv("SIM_SEED", "")

	flagHelp := flag.Bool("help", false, "print help")
	flag.Parse()
	if *flagHelp {
		fmt.Println("greenmetrics-simulator — Modbus-TCP server for GreenMetrics.")
		fmt.Println("  env: SIM_LISTEN, SIM_TZ, SIM_PROFILE, SIM_SEED")
		fmt.Println("  flags: --healthcheck  probe listen port, exit 0/1")
		return
	}

	tz, err := time.LoadLocation(tzName)
	if err != nil {
		log.Printf("warn: unknown SIM_TZ=%q, falling back to UTC", tzName)
		tz = time.UTC
	}

	var seed int64
	if seedStr != "" {
		seed, _ = strconv.ParseInt(seedStr, 10, 64)
	} else {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))

	meters := buildMeters(profile)
	log.Printf("greenmetrics-simulator starting listen=%s tz=%s profile=%s seed=%d meters=%d",
		listen, tz, profile, seed, len(meters))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Update meter state every 5 seconds.
	go runMeterClock(ctx, meters, tz, rng)

	// Accept TCP.
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	// Trap signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Print("shutdown signal received")
		cancel()
		_ = listener.Close()
	}()

	var active atomic.Int32
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || ctx.Err() != nil {
				log.Print("listener closed; exiting")
				return
			}
			log.Printf("accept: %v", err)
			continue
		}
		if active.Load() >= maxConns {
			log.Printf("connection cap %d reached, refusing %s", maxConns, conn.RemoteAddr())
			_ = conn.Close()
			continue
		}
		active.Add(1)
		go func(c net.Conn) {
			defer func() {
				active.Add(-1)
				_ = c.Close()
			}()
			handleConn(ctx, c, meters)
		}(conn)
	}
}

// ----------------------------------------------------------------------------
// Meter construction.
// ----------------------------------------------------------------------------

func buildMeters(profile string) []*meterProfile {
	// `factory_line_verona` — reference profile from plan §5.9.
	// Base and peak loads calibrated to a 120 kWp peak factory in Verona SME belt.
	switch profile {
	case "factory_line_verona":
		return []*meterProfile{
			{slaveID: 1, label: "LINE_A_ELEC", kind: "electricity_3p", baseLoad: 8, peakLoad: 42, processVariance: 0.08, voltageNominal: 400},
			{slaveID: 2, label: "LINE_B_ELEC", kind: "electricity_3p", baseLoad: 6, peakLoad: 36, processVariance: 0.09, voltageNominal: 400},
			{slaveID: 3, label: "HVAC_ELEC", kind: "electricity", baseLoad: 3, peakLoad: 18, processVariance: 0.12, voltageNominal: 230},
			{slaveID: 4, label: "BOILER_GAS", kind: "gas", baseLoad: 0.4, peakLoad: 6.5, processVariance: 0.15},
			{slaveID: 5, label: "DISTRICT_HEAT", kind: "thermal", baseLoad: 4, peakLoad: 20, processVariance: 0.10},
		}
	default:
		return []*meterProfile{
			{slaveID: 1, label: "DEFAULT_ELEC", kind: "electricity", baseLoad: 1, peakLoad: 5, processVariance: 0.05, voltageNominal: 230},
		}
	}
}

// ----------------------------------------------------------------------------
// Load profile (realistic patterns).
// ----------------------------------------------------------------------------

// loadFraction returns a value in [0, 1] representing how close we are to peak
// at the given Italian clock time. It encodes the weekday/weekend split,
// business hours ramp, and an annual seasonal modulation.
func loadFraction(now time.Time) float64 {
	wd := now.Weekday()
	h := float64(now.Hour()) + float64(now.Minute())/60.0

	// Weekend dip.
	weekendFactor := 1.0
	if wd == time.Saturday || wd == time.Sunday {
		weekendFactor = 0.15
	}

	// Business-hours ramp: 0 at night, smooth ramp-up 06–08, plateau 08–18,
	// ramp-down 18–22, near-zero 22–06.
	var daily float64
	switch {
	case h < 6.0:
		daily = 0.05
	case h < 8.0:
		daily = 0.05 + (h-6.0)/(8.0-6.0)*0.90
	case h < 18.0:
		// Gentle bell curve across the workday around 13:00.
		daily = 0.95 - 0.05*math.Cos(2.0*math.Pi*(h-8.0)/10.0)
	case h < 22.0:
		daily = 0.95 - (h-18.0)/(22.0-18.0)*0.90
	default:
		daily = 0.05
	}

	// Seasonal modulation (winter higher for heating-dominated meters).
	doy := now.YearDay()
	season := 0.85 + 0.15*math.Cos(2.0*math.Pi*float64(doy-15)/365.0) // peak mid-January

	return clamp(daily*weekendFactor*season, 0, 1)
}

// gasSeasonal is a hotter seasonal curve (boiler).
func gasSeasonal(now time.Time) float64 {
	doy := now.YearDay()
	// Peak in early January (day ~15), minimum in mid-July (day ~196).
	return clamp(0.55+0.45*math.Cos(2.0*math.Pi*float64(doy-15)/365.0), 0.05, 1.0)
}

// thermalSeasonal — tracks district-heat delivery in Verona.
func thermalSeasonal(now time.Time) float64 {
	doy := now.YearDay()
	return clamp(0.60+0.40*math.Cos(2.0*math.Pi*float64(doy-10)/365.0), 0.10, 1.0)
}

// ----------------------------------------------------------------------------
// Meter clock — advances counters every tick.
// ----------------------------------------------------------------------------

func runMeterClock(ctx context.Context, meters []*meterProfile, tz *time.Location, rng *rand.Rand) {
	const tick = 5 * time.Second
	t := time.NewTicker(tick)
	defer t.Stop()

	// Initialise counters with a plausible mid-year baseline (avoids zero at startup).
	startWh := rng.Float64()*100_000 + 500_000
	startSm3 := rng.Float64()*5_000 + 20_000
	for _, m := range meters {
		switch m.kind {
		case "gas":
			m.counter = startSm3 * 1000 // store as mSm3 for integer-friendly register
		case "thermal":
			m.counter = startWh // store as Wh-eq
		default:
			m.counter = startWh
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			localNow := now.In(tz)
			for _, m := range meters {
				updateMeter(m, localNow, rng)
			}
		}
	}
}

func updateMeter(m *meterProfile, now time.Time, rng *rand.Rand) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var frac float64
	switch m.kind {
	case "gas":
		frac = 0.2*loadFraction(now) + 0.8*gasSeasonal(now)
	case "thermal":
		frac = 0.4*loadFraction(now) + 0.6*thermalSeasonal(now)
	default:
		frac = loadFraction(now)
	}
	// Instantaneous power in native unit/hour.
	nominal := m.baseLoad + (m.peakLoad-m.baseLoad)*frac
	// Multiplicative log-normal noise.
	noise := math.Exp(rng.NormFloat64() * m.processVariance)
	instant := nominal * noise

	// Time delta since last sample.
	var dt float64
	if m.lastSampleAt.IsZero() {
		dt = 5.0
	} else {
		dt = now.Sub(m.lastSampleAt).Seconds()
		if dt <= 0 || dt > 3600 {
			dt = 5.0
		}
	}
	// Accumulate counter in native sub-units.
	switch m.kind {
	case "gas":
		// instant = Sm3/h → Sm3 = instant*dt/3600 → mSm3 = *1000
		m.counter += instant * (dt / 3600.0) * 1000.0
	default:
		// electricity kW → Wh; thermal kWh-eq/h → Wh-eq
		m.counter += instant * (dt / 3600.0) * 1000.0
	}

	m.lastSampleAt = now
	m.lastInstantPwr = instant

	// Electrical-only fields.
	if m.kind == "electricity" || m.kind == "electricity_3p" {
		// Voltage drifts ±1% around nominal.
		m.lastVoltage = m.voltageNominal * (1.0 + (rng.Float64()-0.5)*0.02)
		// Current = instant*1000 / (voltage * sqrt3_for_3p * pf). Simplified:
		pf := 0.92 + (rng.Float64()-0.5)*0.06
		m.lastPF = clamp(pf, 0.80, 0.99)
		denom := m.lastVoltage * m.lastPF
		if m.kind == "electricity_3p" {
			denom *= math.Sqrt(3)
		}
		if denom > 0 {
			m.lastCurrent = instant * 1000.0 / denom
		}
		m.lastFreq = 50.0 + (rng.Float64()-0.5)*0.1
		// Status = OK.
		m.lastStatus = 0
		// Inject a rare fault bit (~0.01% of ticks) — useful for alert-engine smoke tests.
		if rng.Float64() < 0.0001 {
			m.lastStatus = 0x0010 // bit 4: over-voltage warning
		}
	} else {
		m.lastStatus = 0
	}
}

// ----------------------------------------------------------------------------
// TCP connection handler — Modbus MBAP framing.
// ----------------------------------------------------------------------------

func handleConn(ctx context.Context, c net.Conn, meters []*meterProfile) {
	buf := make([]byte, 256)
	for {
		if ctx.Err() != nil {
			return
		}
		_ = c.SetReadDeadline(time.Now().Add(connIdleDeadline))

		// Read MBAP header.
		if _, err := io.ReadFull(c, buf[:mbapHeaderLen]); err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
				log.Printf("mbap header read: %v", err)
			}
			return
		}
		transactionID := binary.BigEndian.Uint16(buf[0:2])
		protocolID := binary.BigEndian.Uint16(buf[2:4])
		length := binary.BigEndian.Uint16(buf[4:6])
		unitID := buf[6]
		if protocolID != 0 {
			return // not Modbus — drop silently
		}
		if length < 2 || length > 250 {
			return
		}
		// Read PDU (length - 1 byte for unit id already consumed).
		pdu := make([]byte, length-1)
		if _, err := io.ReadFull(c, pdu); err != nil {
			log.Printf("pdu read: %v", err)
			return
		}
		if len(pdu) < 1 {
			return
		}
		funcCode := pdu[0]
		resp := buildResponse(unitID, funcCode, pdu[1:], meters)
		writeResponse(c, transactionID, unitID, resp)
	}
}

func buildResponse(unitID, funcCode byte, body []byte, meters []*meterProfile) []byte {
	// Locate meter.
	var m *meterProfile
	for _, x := range meters {
		if x.slaveID == unitID {
			m = x
			break
		}
	}
	if m == nil {
		return exceptionPDU(funcCode, 0x0B) // gateway target device failed to respond
	}

	switch funcCode {
	case 0x03, 0x04: // Read holding / input registers
		if len(body) != 4 {
			return exceptionPDU(funcCode, 0x03)
		}
		startAddr := binary.BigEndian.Uint16(body[0:2])
		quantity := binary.BigEndian.Uint16(body[2:4])
		if quantity == 0 || quantity > 125 {
			return exceptionPDU(funcCode, 0x03)
		}
		regs := readMeterRegisters(m, startAddr, quantity)
		if regs == nil {
			return exceptionPDU(funcCode, 0x02)
		}
		byteCount := byte(quantity * 2)
		out := make([]byte, 0, 2+int(byteCount))
		out = append(out, funcCode, byteCount)
		for _, r := range regs {
			out = append(out, byte(r>>8), byte(r&0xff))
		}
		return out
	default:
		return exceptionPDU(funcCode, 0x01) // illegal function
	}
}

// readMeterRegisters returns the register values for the [start, start+qty) range.
// Returns nil if the range is out of bounds.
func readMeterRegisters(m *meterProfile, start, qty uint16) []uint16 {
	const maxReg = uint16(16)
	if start+qty > maxReg {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	regs := make([]uint16, qty)
	// Build a 16-register view first, then slice.
	view := make([]uint16, maxReg)

	counter := uint32(math.Min(math.MaxUint32, math.Max(0, m.counter)))
	view[0] = uint16(counter >> 16)
	view[1] = uint16(counter & 0xffff)
	instantWord := uint32(math.Min(math.MaxUint32, math.Max(0, m.lastInstantPwr*1000.0)))
	view[2] = uint16(instantWord >> 16)
	view[3] = uint16(instantWord & 0xffff)
	view[4] = uint16(math.Min(65535, m.lastVoltage*10))
	view[5] = uint16(math.Min(65535, m.lastCurrent*100))
	view[6] = uint16(math.Min(65535, m.lastPF*1000))
	view[7] = uint16(math.Min(65535, m.lastFreq*100))
	view[8] = m.lastStatus
	ts := m.lastSampleAt.Unix()
	if ts < 0 {
		ts = 0
	}
	view[9] = uint16(uint32(ts) & 0xffff)
	view[10] = uint16(uint32(ts) >> 16)
	// 11..15 reserved zeros.

	for i := uint16(0); i < qty; i++ {
		regs[i] = view[start+i]
	}
	return regs
}

func exceptionPDU(funcCode, code byte) []byte {
	return []byte{funcCode | 0x80, code}
}

func writeResponse(c net.Conn, transactionID uint16, unitID byte, pdu []byte) {
	header := make([]byte, mbapHeaderLen)
	binary.BigEndian.PutUint16(header[0:2], transactionID)
	binary.BigEndian.PutUint16(header[2:4], 0) // protocol id
	binary.BigEndian.PutUint16(header[4:6], uint16(len(pdu)+1))
	header[6] = unitID
	_ = c.SetWriteDeadline(time.Now().Add(connIdleDeadline))
	if _, err := c.Write(append(header, pdu...)); err != nil {
		log.Printf("write: %v", err)
	}
}

// ----------------------------------------------------------------------------
// Helpers.
// ----------------------------------------------------------------------------

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// runHealthcheck implements the --healthcheck CLI flag. It dials the
// configured listen address (or 127.0.0.1:<port-from-SIM_LISTEN>) and exits
// 0 on success, 1 on any error. Distroless-safe: no shell utilities needed.
func runHealthcheck() {
	target := getenv("SIM_LISTEN", "0.0.0.0:5020")
	// A listen address like "0.0.0.0:5020" must be probed as "127.0.0.1:5020".
	if host, port, err := net.SplitHostPort(target); err == nil {
		if host == "0.0.0.0" || host == "::" || host == "" {
			target = net.JoinHostPort("127.0.0.1", port)
		}
	}
	conn, err := net.DialTimeout("tcp", target, 3*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: dial %s: %v\n", target, err)
		os.Exit(1)
	}
	_ = conn.Close()
}
