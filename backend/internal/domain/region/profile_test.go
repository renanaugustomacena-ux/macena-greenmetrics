// Example test for the Region Pack contract. Per Rule 86.

package region_test

import (
	"testing"
	"time"

	"github.com/greenmetrics/backend/internal/domain/region"
)

// stubItaly models a minimal Italian Region Pack.
type stubItaly struct{}

func (stubItaly) Code() string { return "it" }

func (stubItaly) Profile() region.Profile {
	return region.Profile{
		Code:             "it",
		Timezone:         "Europe/Rome",
		Locale:           "it_IT.UTF-8",
		CurrencyISO4217:  "EUR",
		DecimalSeparator: ",",
		DefaultRegimes: []region.RegulatoryRegime{
			region.RegimeCSRDWave2,
			region.RegimePiano50,
			region.RegimeContoTermico,
			region.RegimeTEE,
			region.RegimeAuditDLgs102,
			region.RegimeGDPR,
			region.RegimeNIS2Italia,
			region.RegimeARERA,
		},
		Authorities: map[string]string{
			"GSE":     "https://www.gse.it",
			"ENEA":    "https://www.efficienzaenergetica.enea.it",
			"ARERA":   "https://www.arera.it",
			"Garante": "https://www.garanteprivacy.it",
			"MIMIT":   "https://www.mimit.gov.it",
			"MASE":    "https://www.mase.gov.it",
		},
	}
}

func (stubItaly) HolidayCalendar(year int) []time.Time {
	loc, _ := time.LoadLocation("Europe/Rome")
	return []time.Time{
		time.Date(year, 1, 1, 0, 0, 0, 0, loc),   // Capodanno
		time.Date(year, 1, 6, 0, 0, 0, 0, loc),   // Epifania
		time.Date(year, 4, 25, 0, 0, 0, 0, loc),  // Liberazione
		time.Date(year, 5, 1, 0, 0, 0, 0, loc),   // Festa del Lavoro
		time.Date(year, 6, 2, 0, 0, 0, 0, loc),   // Festa della Repubblica
		time.Date(year, 8, 15, 0, 0, 0, 0, loc),  // Ferragosto
		time.Date(year, 11, 1, 0, 0, 0, 0, loc),  // Tutti i Santi
		time.Date(year, 12, 8, 0, 0, 0, 0, loc),  // Immacolata
		time.Date(year, 12, 25, 0, 0, 0, 0, loc), // Natale
		time.Date(year, 12, 26, 0, 0, 0, 0, loc), // Santo Stefano
	}
}

func (stubItaly) RegulatoryThreshold(name string) (float64, string, bool) {
	thresholds := map[string]struct {
		value   float64
		version string
	}{
		"csrd_wave_2_employee_threshold":          {250, "2024.1"},
		"csrd_wave_2_turnover_eur_threshold":      {50_000_000, "2024.1"},
		"csrd_wave_2_balance_sheet_eur_threshold": {25_000_000, "2024.1"},
		"piano_5_0_process_reduction_pct":         {3.0, "2024.1"},
		"piano_5_0_site_reduction_pct":            {5.0, "2024.1"},
	}
	t, ok := thresholds[name]
	return t.value, t.version, ok
}

func TestExample_RegionProfileSurface(t *testing.T) {
	var rp region.RegionProfile = stubItaly{}
	p := rp.Profile()

	if rp.Code() != "it" {
		t.Errorf("Code = %q; want it", rp.Code())
	}
	if p.Timezone != "Europe/Rome" {
		t.Errorf("Timezone = %q", p.Timezone)
	}
	if p.CurrencyISO4217 != "EUR" {
		t.Errorf("Currency = %q", p.CurrencyISO4217)
	}

	holidays := rp.HolidayCalendar(2026)
	if len(holidays) < 10 {
		t.Errorf("expected 10+ Italian public holidays; got %d", len(holidays))
	}

	v, version, ok := rp.RegulatoryThreshold("piano_5_0_process_reduction_pct")
	if !ok {
		t.Fatal("piano_5_0_process_reduction_pct not found")
	}
	if v != 3.0 {
		t.Errorf("piano_5_0 process threshold = %v; want 3.0", v)
	}
	if version == "" {
		t.Error("threshold version must be non-empty per Rule 139")
	}
}

func TestContractVersion_IsSet(t *testing.T) {
	if region.ContractVersion == "" {
		t.Fatal("ContractVersion empty — Rule 71 requires per-kind contract version")
	}
}
