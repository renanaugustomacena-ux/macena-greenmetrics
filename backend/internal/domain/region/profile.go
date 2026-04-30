// Package region defines the Pack contract for regional defaults and
// regulatory-regime overlays.
//
// Doctrine refs: Rules 88 (Italian Region Pack as flagship reference),
// 101 (per-tenant timezone), 132 (Italian regulatory ground truth annotated
// to primary sources), 139 (thresholds propagated, not duplicated), 140
// (per-tenant regulatory profile is explicit).
// Charter ref: §3.2 Region Packs. ADR-0023 records the interface adoption.
//
// A Region Pack at packs/region/<id>/ implements the RegionProfile interface
// below. Each Region Pack bundles: timezone, locale, currency, holiday
// calendar, regulatory-regime defaults (CSRD wave, audit obligation,
// applicable-decree set), privacy-regime overlay (GDPR + national
// supplements), and the per-region threshold set referenced by Report Packs.
//
// The Italian Region Pack at packs/region/it/ is the flagship. New Region
// Packs (DE, FR, ES, GB, AT) follow its structure verbatim.
package region

import (
	"time"
)

// ContractVersion is the SemVer of this Pack-contract package. Per Rule 71.
const ContractVersion = "1.0.0"

// RegulatoryRegime identifies a regulator-driven obligation that the
// region's tenants may participate in. Each Pack declares which regimes its
// region's tenants face by default; per-tenant regulatory profile (Rule 140)
// can opt in / opt out.
type RegulatoryRegime string

const (
	// Italian regimes.
	RegimeCSRDWave1    RegulatoryRegime = "csrd_wave_1"
	RegimeCSRDWave2    RegulatoryRegime = "csrd_wave_2"
	RegimeCSRDWave3    RegulatoryRegime = "csrd_wave_3"
	RegimePiano50      RegulatoryRegime = "piano_5_0"
	RegimeContoTermico RegulatoryRegime = "conto_termico"
	RegimeTEE          RegulatoryRegime = "tee"
	RegimeAuditDLgs102 RegulatoryRegime = "audit_dlgs102"
	RegimeETS          RegulatoryRegime = "ets"
	RegimeGDPR         RegulatoryRegime = "gdpr"
	RegimeNIS2Italia   RegulatoryRegime = "nis2_italia"
	RegimeARERA        RegulatoryRegime = "arera"

	// Cross-EU.
	RegimeEUAllowanceTrading RegulatoryRegime = "eu_ets"
	RegimeIFRS_S1S2          RegulatoryRegime = "ifrs_s1_s2"
	RegimeTCFD               RegulatoryRegime = "tcfd"

	// UK.
	RegimeUKSECR RegulatoryRegime = "uk_secr"

	// US.
	RegimeSECClimateDisclosure RegulatoryRegime = "sec_climate_disclosure"
)

// Profile is the static surface of a Region Pack. Returned by RegionProfile
// methods. The profile is the read-only authoritative answer for the region.
type Profile struct {
	// Region code (matches Pack id; e.g. "it", "de", "fr", "es", "gb", "at").
	Code string `json:"code"`
	// IANA timezone (e.g. "Europe/Rome", "Europe/Berlin").
	Timezone string `json:"timezone"`
	// POSIX locale (e.g. "it_IT.UTF-8", "de_DE.UTF-8").
	Locale string `json:"locale"`
	// ISO-4217 currency (e.g. "EUR", "GBP", "CHF").
	CurrencyISO4217 string `json:"currency_iso4217"`
	// Default decimal separator for human-readable rendering ("," or ".").
	DecimalSeparator string `json:"decimal_separator"`
	// Default regulatory regimes participated in by the region's tenants.
	DefaultRegimes []RegulatoryRegime `json:"default_regimes"`
	// Regulator authority contact map (key = regulator short name; value = URL or address).
	Authorities map[string]string `json:"authorities"`
}

// RegionProfile is the Pack-contract for regional defaults.
type RegionProfile interface {
	// Code is the region identifier (matches Pack id).
	Code() string

	// Profile returns the static profile (timezone, locale, currency, regimes).
	Profile() Profile

	// HolidayCalendar returns the set of public holidays for the region in the given year.
	// Used by reporting time-bucket logic and by the operator dashboard.
	HolidayCalendar(year int) []time.Time

	// RegulatoryThreshold looks up a threshold by symbolic name (e.g.
	// "csrd_wave_2_employee_threshold", "piano_5_0_process_reduction_pct",
	// "audit_dlgs102_kwh_threshold"). Returns the value and the version
	// of the threshold (per Rule 139, thresholds are propagated, not duplicated).
	RegulatoryThreshold(name string) (value float64, version string, ok bool)
}
