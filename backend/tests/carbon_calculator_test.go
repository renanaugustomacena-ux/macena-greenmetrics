package tests

import (
	"math"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/greenmetrics/backend/internal/services"
)

// TestCarbonCalculator_Scope2_ISPRA2023 verifies the ISPRA 2023 Italian
// electricity-mix factor (0.250 kg CO2e / kWh) is applied exactly for a
// reference year.
func TestCarbonCalculator_Scope2_ISPRA2023(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	c := services.NewCarbonCalculator(nil, logger)

	// 100 000 kWh consumed mid-2023.
	at := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	kg, f := c.Scope2FromElectricity(100_000, at, false)

	if f.Code != "IT_ELEC_MIX_2023" {
		t.Fatalf("expected factor IT_ELEC_MIX_2023, got %s", f.Code)
	}
	if f.KgCO2ePer != 0.250 {
		t.Fatalf("expected factor value 0.250, got %v", f.KgCO2ePer)
	}
	expected := 25_000.0
	if math.Abs(kg-expected) > 1e-6 {
		t.Fatalf("expected %.2f kgCO2e, got %.2f", expected, kg)
	}
}

// TestCarbonCalculator_Scope2_ISPRA2024 verifies the 2024 provisional factor.
func TestCarbonCalculator_Scope2_ISPRA2024(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	c := services.NewCarbonCalculator(nil, logger)

	at := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	kg, f := c.Scope2FromElectricity(100_000, at, false)

	if f.Code != "IT_ELEC_MIX_2024" {
		t.Fatalf("expected factor IT_ELEC_MIX_2024, got %s", f.Code)
	}
	if math.Abs(f.KgCO2ePer-0.245) > 1e-6 {
		t.Fatalf("expected factor value 0.245, got %v", f.KgCO2ePer)
	}
	expected := 24_500.0
	if math.Abs(kg-expected) > 1e-6 {
		t.Fatalf("expected %.2f kgCO2e, got %.2f", expected, kg)
	}
}

// TestCarbonCalculator_Scope1_NaturalGas verifies the D.M. 11/05/2022 factor.
// 1.975 kg CO2e / Sm3.
func TestCarbonCalculator_Scope1_NaturalGas(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	c := services.NewCarbonCalculator(nil, logger)

	at := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	// 10 000 Sm3 combusted.
	kg, f := c.Scope1FromGas(10_000, at)

	if f.Code != "NG_STATIONARY_COMBUSTION" {
		t.Fatalf("expected factor NG_STATIONARY_COMBUSTION, got %s", f.Code)
	}
	if math.Abs(f.KgCO2ePer-1.975) > 1e-6 {
		t.Fatalf("expected factor value 1.975, got %v", f.KgCO2ePer)
	}
	expected := 19_750.0
	if math.Abs(kg-expected) > 1e-6 {
		t.Fatalf("expected %.2f kgCO2e, got %.2f", expected, kg)
	}
}

// TestCarbonCalculator_MarketBased verifies market-based residual mix.
func TestCarbonCalculator_MarketBased(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	c := services.NewCarbonCalculator(nil, logger)

	at := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
	kg, f := c.Scope2FromElectricity(1_000, at, true)
	if f.Code != "IT_ELEC_RESIDUAL_MIX_2023" {
		t.Fatalf("expected IT_ELEC_RESIDUAL_MIX_2023, got %s", f.Code)
	}
	if math.Abs(f.KgCO2ePer-0.457) > 1e-6 {
		t.Fatalf("expected 0.457 (AIB residual 2023), got %v", f.KgCO2ePer)
	}
	if math.Abs(kg-457.0) > 1e-6 {
		t.Fatalf("expected 457 kg CO2e, got %.2f", kg)
	}
}

// TestCarbonCalculator_Diesel verifies Scope 1 diesel factor.
func TestCarbonCalculator_Diesel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	c := services.NewCarbonCalculator(nil, logger)
	at := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	// 1 000 L diesel.
	kg, f := c.Scope1FromDiesel(1_000, at)
	if f.Code != "DIESEL_COMBUSTION" {
		t.Fatalf("expected DIESEL_COMBUSTION, got %s", f.Code)
	}
	if math.Abs(kg-2_650.0) > 1e-6 {
		t.Fatalf("expected 2 650 kg CO2e, got %.2f", kg)
	}
}

// TestCarbonCalculator_Scope3_Stub verifies the EEIO placeholder is finite.
func TestCarbonCalculator_Scope3_Stub(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	c := services.NewCarbonCalculator(nil, logger)
	kg := c.Scope3FromSupplier(10_000, "logistics_road")
	if kg != 4_500 {
		t.Fatalf("expected 4500 for logistics_road, got %.2f", kg)
	}
	// Unknown category falls back to 0.50.
	kgUnknown := c.Scope3FromSupplier(1_000, "cosmic-rays")
	if kgUnknown != 500 {
		t.Fatalf("expected 500 for fallback, got %.2f", kgUnknown)
	}
}
