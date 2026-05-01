// Package it implements the Italian Region Pack — the flagship reference
// Pack per Rule 88.
//
// The Pack ships static data (holiday calendar, regulatory thresholds,
// authority map) plus a stateless Profile implementation. It has no
// runtime dependencies and is safe to construct at boot.
//
// Cross-refs:
//   - Pack contract:    backend/internal/domain/region/profile.go
//   - Manifest:         packs/region/it/manifest.yaml
//   - Charter:          packs/region/it/CHARTER.md
package it

import (
	"sort"
	"time"

	"github.com/greenmetrics/backend/internal/domain/region"
)

// Pack is the singleton instance constructed at boot.
var Pack region.RegionProfile = &profile{}

// PackVersion is the Pack's SemVer (matches manifest.yaml version).
const PackVersion = "1.0.0"

// profile is the concrete RegionProfile for Italy.
type profile struct{}

// Code implements region.RegionProfile.
func (p *profile) Code() string { return "it" }

// Profile implements region.RegionProfile. The values mirror packs/region/it/CHARTER.md §2.
func (p *profile) Profile() region.Profile {
	return region.Profile{
		Code:             "it",
		Timezone:         "Europe/Rome",
		Locale:           "it_IT.UTF-8",
		CurrencyISO4217:  "EUR",
		DecimalSeparator: ",",
		DefaultRegimes: []region.RegulatoryRegime{
			region.RegimeCSRDWave2,
			region.RegimeCSRDWave3,
			region.RegimePiano50,
			region.RegimeContoTermico,
			region.RegimeTEE,
			region.RegimeAuditDLgs102,
			region.RegimeETS,
			region.RegimeGDPR,
			region.RegimeNIS2Italia,
			region.RegimeARERA,
		},
		Authorities: map[string]string{
			"gse":     "https://areaclienti.gse.it",
			"enea":    "https://audit102.enea.it",
			"ispra":   "https://www.isprambiente.gov.it",
			"terna":   "https://transparency.entsoe.eu",
			"arera":   "https://www.arera.it",
			"garante": "https://www.garanteprivacy.it",
			"acn":     "https://www.acn.gov.it",
			"mimit":   "https://www.mimit.gov.it",
			"mase":    "https://www.mase.gov.it",
		},
	}
}

// HolidayCalendar implements region.RegionProfile.
//
// Italian national holidays per L. 27 maggio 1949, n. 260, plus the
// Veneto regional holiday (Patrono di Verona, 21 May). Easter and
// Easter Monday are computed via the Meeus/Jones/Butcher Gregorian
// algorithm. All times are 00:00:00 in the Europe/Rome timezone, then
// converted to UTC for the returned slice.
func (p *profile) HolidayCalendar(year int) []time.Time {
	tz, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		// Defensive: if Europe/Rome is somehow not in the tzdata, fall
		// back to UTC. Boot-time conformance test (Rule 19) catches this.
		tz = time.UTC
	}

	at := func(month, day int) time.Time {
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, tz).UTC()
	}

	easter := easterSunday(year, tz)
	easterMonday := easter.AddDate(0, 0, 1)

	out := []time.Time{
		at(1, 1),   // Capodanno
		at(1, 6),   // Epifania
		easter,     // Pasqua
		easterMonday, // Lunedì dell'Angelo
		at(4, 25),  // Festa della Liberazione
		at(5, 1),   // Festa del Lavoro
		at(5, 21),  // Patrono di Verona (Veneto regional)
		at(6, 2),   // Festa della Repubblica
		at(8, 15),  // Ferragosto
		at(11, 1),  // Tutti i Santi
		at(12, 8),  // Immacolata Concezione
		at(12, 25), // Natale
		at(12, 26), // Santo Stefano
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Before(out[j]) })
	return out
}

// RegulatoryThreshold implements region.RegionProfile. Per Rule 139,
// thresholds are propagated to Report Packs — they are NOT duplicated
// inside Builder code. Each entry pairs a value with the source-document
// version annotation.
func (p *profile) RegulatoryThreshold(name string) (float64, string, bool) {
	v, ok := thresholds[name]
	if !ok {
		return 0, "", false
	}
	return v.value, v.version, true
}

type threshold struct {
	value   float64
	version string // primary-source version annotation
}

// thresholds is the threshold table cited in CHARTER.md §4. Each version
// string identifies the source document at the time the value was
// recorded; bumping the value bumps the version.
var thresholds = map[string]threshold{
	"csrd_wave_2_employee_threshold":      {250, "Dir. UE 2022/2464 art. 19a"},
	"csrd_wave_2_turnover_eur":            {50_000_000, "Dir. UE 2022/2464 art. 19a"},
	"csrd_wave_2_balance_sheet_eur":       {25_000_000, "Dir. UE 2022/2464 art. 19a"},
	"piano_5_0_process_reduction_pct":     {3.0, "DL 19/2024 art. 38 c.4"},
	"piano_5_0_site_reduction_pct":        {5.0, "DL 19/2024 art. 38 c.4"},
	"piano_5_0_max_qualifying_outlay_eur": {50_000_000, "DL 19/2024 art. 38 c.6"},
	"audit_dlgs102_kwh_threshold":         {0.0, "D.Lgs. 102/2014 art. 8 (no kWh threshold)"},
	"audit_dlgs102_employee_threshold":    {250, "D.Lgs. 102/2014 art. 8 c.1"},
	"audit_dlgs102_turnover_eur":          {50_000_000, "D.Lgs. 102/2014 art. 8 c.1"},
	"audit_dlgs102_balance_sheet_eur":     {43_000_000, "D.Lgs. 102/2014 art. 8 c.1"},
	"tee_minimum_certificate_toe":         {1.0, "DM 11 gennaio 2017 art. 6"},
	"conto_termico_max_intervento_eur":    {700_000, "DM 16 febbraio 2016 art. 5"},
	"arera_smart_meter_data_window_min":   {15, "ARERA delibera 646/2015"},
	"nis2_notification_window_h_initial":  {24, "D.Lgs. 138/2024 art. 24 c.1"},
	"nis2_notification_window_h_full":     {72, "D.Lgs. 138/2024 art. 24 c.2"},
	"nis2_final_report_window_d":          {30, "D.Lgs. 138/2024 art. 24 c.3"},
}

// easterSunday returns the date of Easter Sunday in `year` at 00:00 in
// the supplied timezone, converted to UTC. Uses the anonymous Gregorian
// algorithm (Meeus / Jones / Butcher).
func easterSunday(year int, tz *time.Location) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, tz).UTC()
}
