package tests

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/greenmetrics/backend/internal/services"
)

// TestPiano50_BaselineSaving3Percent covers the deterministic scenario
// described in the plan: baseline 100 000 kWh, saving ≥ 3%, credit 5% bracket.
func TestPiano50_BaselineSaving3Percent(t *testing.T) {
	// Exactly 3% saving, process scope, 500 000 EUR eligible spend.
	r := services.ComputePianoTransizione50Result(
		100_000, 97_000, 500_000, true,
	)
	if math.Abs(r.EnergyReductionPct-3.0) > 1e-6 {
		t.Fatalf("expected 3%% reduction, got %.4f", r.EnergyReductionPct)
	}
	if !r.MeetsProcessThresh {
		t.Fatal("expected MeetsProcessThresh=true at 3%")
	}
	if r.MeetsSiteThresh {
		t.Fatal("site threshold should not trigger at 3% process-only")
	}
	if r.TaxCreditBand != "3-6% (aliquota base 5%)" {
		t.Fatalf("expected base band '3-6%% (aliquota base 5%%)', got %q", r.TaxCreditBand)
	}
	expectedCredit := 500_000 * 0.05
	if math.Abs(r.ExpectedCreditEUR-expectedCredit) > 1e-6 {
		t.Fatalf("expected credit %.2f EUR, got %.2f", expectedCredit, r.ExpectedCreditEUR)
	}
}

// TestPiano50_SiteSaving6Percent exercises the intermediate band.
func TestPiano50_SiteSaving6Percent(t *testing.T) {
	r := services.ComputePianoTransizione50Result(
		1_000_000, 940_000, 2_000_000, false,
	)
	if math.Abs(r.EnergyReductionPct-6.0) > 1e-6 {
		t.Fatalf("expected 6%% reduction, got %.4f", r.EnergyReductionPct)
	}
	if !r.MeetsSiteThresh {
		t.Fatal("expected MeetsSiteThresh=true at 6%")
	}
	if r.TaxCreditBand != "6-10% (aliquota 20%)" {
		t.Fatalf("expected '6-10%% (aliquota 20%%)', got %q", r.TaxCreditBand)
	}
	if math.Abs(r.ExpectedCreditEUR-(2_000_000*0.20)) > 1e-6 {
		t.Fatalf("expected credit 400 000, got %.2f", r.ExpectedCreditEUR)
	}
}

// TestPiano50_Saving15Percent exercises the upper band.
func TestPiano50_Saving15Percent(t *testing.T) {
	r := services.ComputePianoTransizione50Result(
		1_000_000, 850_000, 1_000_000, true,
	)
	if math.Abs(r.EnergyReductionPct-15.0) > 1e-6 {
		t.Fatalf("expected 15%% reduction, got %.4f", r.EnergyReductionPct)
	}
	if r.TaxCreditBand != "15%+ (aliquota superiore 40%)" {
		t.Fatalf("expected upper band, got %q", r.TaxCreditBand)
	}
	if math.Abs(r.ExpectedCreditEUR-400_000) > 1e-6 {
		t.Fatalf("expected credit 400 000, got %.2f", r.ExpectedCreditEUR)
	}
}

// TestPiano50_BelowThreshold verifies non-eligibility.
func TestPiano50_BelowThreshold(t *testing.T) {
	r := services.ComputePianoTransizione50Result(
		100_000, 98_000, 500_000, true,
	)
	if r.MeetsProcessThresh || r.MeetsSiteThresh {
		t.Fatal("2% reduction must not meet any threshold")
	}
	if r.TaxCreditBand != "non-ammissibile" {
		t.Fatalf("expected 'non-ammissibile', got %q", r.TaxCreditBand)
	}
	if r.ExpectedCreditEUR != 0 {
		t.Fatalf("expected 0 credit, got %.2f", r.ExpectedCreditEUR)
	}
}

// TestRenderPianoTransizione50HTML ensures the template binds all key fields
// and emits valid HTML (starts with doctype, contains expected sections).
func TestRenderPianoTransizione50HTML(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rg := services.NewReportGenerator(nil, nil, nil, logger)
	r := services.ComputePianoTransizione50Result(100_000, 97_000, 500_000, true)
	payload := map[string]any{
		"attestazione":   r,
		"company_name":   "Fornace Rossi S.r.l.",
		"company_vat":    "IT01234567890",
		"tenant_id":      "t1",
		"methodology":    "EN 16247-3",
		"normative_ref":  "DL 19/2024",
		"signer_role":    "EGE",
		"signer_name":    "Dott. Mario Bianchi",
		"signer_cert_id": "UNI-CEI-11339-XXXX",
		"generated_at":   "2026-04-17T12:00:00Z",
		"period_from":    "2026-01-01",
		"period_to":      "2026-03-31",
	}
	b, err := rg.RenderPianoTransizione50HTML(payload)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("<!doctype html>")) {
		t.Fatal("expected HTML doctype prefix")
	}
	for _, needle := range []string{
		"Attestazione Piano Transizione 5.0",
		"Fornace Rossi S.r.l.",
		"IT01234567890",
		"Riduzione energetica",
		"3-6% (aliquota base 5%)",
		"25000.00", // credit = 500000 * 0.05
	} {
		if !strings.Contains(string(b), needle) {
			t.Errorf("rendered HTML missing %q", needle)
		}
	}
}

// TestRenderESRSE1HTML ensures the ESRS E1 template binds and shows each
// disclosure code (E1-5 / E1-6 / E1-7).
func TestRenderESRSE1HTML(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rg := services.NewReportGenerator(nil, nil, nil, logger)
	payload := map[string]any{
		"disclosure_standard":   "ESRS E1",
		"reporting_period_from": "2026-01-01",
		"reporting_period_to":   "2026-03-31",
		"generated_at":          "2026-04-17T12:00:00Z",
		"data_points": []map[string]any{
			{"Code": "E1-5", "Description": "Consumo non rinnovabile", "Value": 65000.0, "Unit": "kWh", "Methodology": "ISPRA"},
			{"Code": "E1-5", "Description": "Consumo rinnovabile", "Value": 35000.0, "Unit": "kWh", "Methodology": "GSE"},
			{"Code": "E1-6", "Description": "Scope 1", "Value": 10000.0, "Unit": "kg CO2e"},
			{"Code": "E1-6", "Description": "Scope 2 (LB)", "Value": 16250.0, "Unit": "kg CO2e"},
			{"Code": "E1-6", "Description": "Scope 3", "Value": 0.0, "Unit": "kg CO2e"},
			{"Code": "E1-7", "Description": "Intensità", "Value": 0.000025, "Unit": "kg CO2e / €", "Methodology": "ricavi"},
		},
	}
	b, err := rg.RenderESRSE1HTML(payload)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, needle := range []string{
		"ESRS E1",
		"E1-5",
		"E1-6",
		"E1-7",
		"Consumo non rinnovabile",
		"16250.00",
	} {
		if !strings.Contains(string(b), needle) {
			t.Errorf("rendered HTML missing %q", needle)
		}
	}
}
