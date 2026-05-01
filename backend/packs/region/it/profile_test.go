package it

import (
	"testing"
	"time"

	"github.com/greenmetrics/backend/internal/domain/region"
)

// TestPackImplementsRegionProfile is the compile-time + runtime check that
// the Pack satisfies the contract.
func TestPackImplementsRegionProfile(t *testing.T) {
	var _ region.RegionProfile = Pack
	if Pack.Code() != "it" {
		t.Fatalf("expected code=it, got %q", Pack.Code())
	}
}

// TestProfile asserts the static fields match the manifest + CHARTER.
func TestProfile(t *testing.T) {
	p := Pack.Profile()
	if p.Code != "it" {
		t.Errorf("Code: want it, got %q", p.Code)
	}
	if p.Timezone != "Europe/Rome" {
		t.Errorf("Timezone: want Europe/Rome, got %q", p.Timezone)
	}
	if p.Locale != "it_IT.UTF-8" {
		t.Errorf("Locale: want it_IT.UTF-8, got %q", p.Locale)
	}
	if p.CurrencyISO4217 != "EUR" {
		t.Errorf("Currency: want EUR, got %q", p.CurrencyISO4217)
	}
	if p.DecimalSeparator != "," {
		t.Errorf("DecimalSeparator: want comma, got %q", p.DecimalSeparator)
	}
	if got := len(p.DefaultRegimes); got < 5 {
		t.Errorf("expected ≥5 default regimes, got %d", got)
	}
	if _, ok := p.Authorities["gse"]; !ok {
		t.Error("Authorities should contain gse")
	}
}

// TestHolidayCalendar2026 verifies the fixed-date holidays are present
// for a representative year and that Easter falls on the expected date.
// Easter 2026 is 5 April per the Gregorian algorithm.
func TestHolidayCalendar2026(t *testing.T) {
	hols := Pack.HolidayCalendar(2026)
	if len(hols) != 13 {
		t.Errorf("expected 13 holidays in 2026, got %d", len(hols))
	}

	// Check sorted ascending.
	for i := 1; i < len(hols); i++ {
		if hols[i-1].After(hols[i]) {
			t.Errorf("holidays not sorted ascending at index %d: %s after %s",
				i, hols[i-1], hols[i])
		}
	}

	// Spot-check fixed Italian holidays (in Europe/Rome → UTC: 00:00 local
	// in CET = 23:00 UTC previous day in winter; 00:00 in CEST = 22:00 UTC
	// previous day in summer; we just verify the date is in the set).
	mustHave := []struct {
		month time.Month
		day   int
		name  string
	}{
		{time.January, 1, "Capodanno"},
		{time.January, 6, "Epifania"},
		{time.April, 25, "Festa della Liberazione"},
		{time.May, 1, "Festa del Lavoro"},
		{time.May, 21, "Patrono di Verona"},
		{time.June, 2, "Festa della Repubblica"},
		{time.August, 15, "Ferragosto"},
		{time.November, 1, "Tutti i Santi"},
		{time.December, 8, "Immacolata Concezione"},
		{time.December, 25, "Natale"},
		{time.December, 26, "Santo Stefano"},
	}
	for _, mh := range mustHave {
		if !hasDateAt(hols, 2026, mh.month, mh.day) {
			t.Errorf("missing holiday %s (%d-%d)", mh.name, mh.month, mh.day)
		}
	}
}

// hasDateAt checks that any of the holidays falls on (year, month, day) in
// either UTC or Europe/Rome — robust against the DST/UTC rollover.
func hasDateAt(hols []time.Time, year int, month time.Month, day int) bool {
	rome, _ := time.LoadLocation("Europe/Rome")
	for _, h := range hols {
		hUTC := h.UTC()
		hRome := h.In(rome)
		if (hUTC.Year() == year && hUTC.Month() == month && hUTC.Day() == day) ||
			(hRome.Year() == year && hRome.Month() == month && hRome.Day() == day) {
			return true
		}
	}
	return false
}

// TestEasterSundayKnownDates checks the Meeus algorithm against published
// Easter dates for a wide year range.
func TestEasterSundayKnownDates(t *testing.T) {
	cases := []struct {
		year  int
		month time.Month
		day   int
	}{
		{2024, time.March, 31},
		{2025, time.April, 20},
		{2026, time.April, 5},
		{2027, time.March, 28},
		{2028, time.April, 16},
		{2030, time.April, 21},
		{2038, time.April, 25}, // late Easter
		{2008, time.March, 23}, // early Easter
	}
	rome, _ := time.LoadLocation("Europe/Rome")
	for _, c := range cases {
		got := easterSunday(c.year, rome).In(rome)
		if got.Year() != c.year || got.Month() != c.month || got.Day() != c.day {
			t.Errorf("easter %d: want %d-%02d-%02d, got %s",
				c.year, c.year, c.month, c.day, got.Format("2006-01-02"))
		}
	}
}

// TestRegulatoryThreshold verifies the table is queryable by symbolic
// name and that unknown names return ok=false.
func TestRegulatoryThreshold(t *testing.T) {
	v, ver, ok := Pack.RegulatoryThreshold("csrd_wave_2_employee_threshold")
	if !ok {
		t.Fatal("csrd_wave_2_employee_threshold should be present")
	}
	if v != 250 {
		t.Errorf("csrd_wave_2_employee_threshold value: want 250, got %v", v)
	}
	if ver == "" {
		t.Error("version annotation should be non-empty")
	}

	v, _, ok = Pack.RegulatoryThreshold("piano_5_0_process_reduction_pct")
	if !ok || v != 3.0 {
		t.Errorf("piano_5_0_process_reduction_pct: got value=%v ok=%v", v, ok)
	}

	_, _, ok = Pack.RegulatoryThreshold("nonexistent_threshold")
	if ok {
		t.Error("unknown threshold should return ok=false")
	}
}

// TestThresholdsCarryVersion verifies every threshold has a non-empty
// version annotation per Rule 132 (regulatory ground truth annotated to
// primary sources).
func TestThresholdsCarryVersion(t *testing.T) {
	for name, th := range thresholds {
		if th.version == "" {
			t.Errorf("%s: version annotation must not be empty (Rule 132)", name)
		}
	}
}
