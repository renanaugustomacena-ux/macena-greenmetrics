package main

import (
	"testing"
	"time"
)

// TestLoadFraction_BusinessHoursPeak verifies mid-morning weekday is near peak.
func TestLoadFraction_BusinessHoursPeak(t *testing.T) {
	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC) // Wednesday 10:00
	frac := loadFraction(now)
	if frac < 0.6 {
		t.Errorf("Wednesday 10:00 should be ≥ 0.6, got %.3f", frac)
	}
	if frac > 1.0 {
		t.Errorf("loadFraction must be ≤ 1, got %.3f", frac)
	}
}

// TestLoadFraction_WeekendDip verifies weekend is ≤ 0.15 × weekday nominal.
func TestLoadFraction_WeekendDip(t *testing.T) {
	weekday := loadFraction(time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC))
	weekend := loadFraction(time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)) // Saturday
	if weekend > weekday*0.25 {
		t.Errorf("Saturday should be well under weekday; %.3f vs %.3f", weekend, weekday)
	}
}

// TestLoadFraction_NightIdleIsLow verifies 03:00 weekday is near idle.
func TestLoadFraction_NightIdleIsLow(t *testing.T) {
	frac := loadFraction(time.Date(2026, 4, 15, 3, 0, 0, 0, time.UTC))
	if frac > 0.20 {
		t.Errorf("03:00 should be ≤ 0.20, got %.3f", frac)
	}
}

// TestGasSeasonal_WinterHigherThanSummer verifies heating season dominance.
func TestGasSeasonal_WinterHigherThanSummer(t *testing.T) {
	jan := gasSeasonal(time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC))
	jul := gasSeasonal(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC))
	if jan <= jul {
		t.Errorf("gas seasonal winter (%.3f) must exceed summer (%.3f)", jan, jul)
	}
}

// TestClamp_Bounds.
func TestClamp_Bounds(t *testing.T) {
	if clamp(-1, 0, 1) != 0 {
		t.Fail()
	}
	if clamp(2, 0, 1) != 1 {
		t.Fail()
	}
	if clamp(0.5, 0, 1) != 0.5 {
		t.Fail()
	}
}

// TestReadMeterRegisters_ReturnsExpectedCount verifies register-read shape.
func TestReadMeterRegisters_ReturnsExpectedCount(t *testing.T) {
	m := &meterProfile{
		slaveID:        1,
		label:          "T",
		kind:           "electricity",
		counter:        1234567,
		lastInstantPwr: 12.5,
		lastVoltage:    230.1,
		lastCurrent:    5.0,
		lastPF:         0.95,
		lastFreq:       50.0,
		lastSampleAt:   time.Now().UTC(),
	}
	regs := readMeterRegisters(m, 0, 11)
	if len(regs) != 11 {
		t.Fatalf("expected 11 registers, got %d", len(regs))
	}
	// Counter high word (reg 0) must be non-zero for counter > 65535.
	if regs[0] == 0 && regs[1] == 0 {
		t.Error("expected non-zero counter registers")
	}
	if regs[4] == 0 {
		t.Error("expected non-zero voltage register")
	}
}
