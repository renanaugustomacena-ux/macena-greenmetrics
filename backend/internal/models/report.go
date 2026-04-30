package models

import "time"

// ReportType enumerates supported dossier types.
type ReportType string

const (
	ReportMonthlyConsumption    ReportType = "monthly_consumption"
	ReportCO2Footprint          ReportType = "co2_footprint"
	ReportESRSE1                ReportType = "esrs_e1_csrd"
	ReportPianoTransizione50    ReportType = "piano_5_0_attestazione"
	ReportContoTermico20        ReportType = "conto_termico_2_0"
	ReportCertificatiBianchiTEE ReportType = "certificati_bianchi_tee"
	ReportAuditDLgs102          ReportType = "audit_dlgs_102_2014"
)

// ReportStatus represents the lifecycle state of a generated report.
type ReportStatus string

const (
	ReportStatusDraft     ReportStatus = "draft"
	ReportStatusGenerated ReportStatus = "generated"
	ReportStatusSubmitted ReportStatus = "submitted"
	ReportStatusAccepted  ReportStatus = "accepted"
	ReportStatusRejected  ReportStatus = "rejected"
)

// Report is a generated sustainability / regulatory dossier.
type Report struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	Type       ReportType     `json:"type"`
	PeriodFrom time.Time      `json:"period_from"`
	PeriodTo   time.Time      `json:"period_to"`
	Status     ReportStatus   `json:"status"`
	Payload    map[string]any `json:"payload"`
	FileURL    string         `json:"file_url,omitempty"`
	GeneratedBy string        `json:"generated_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// PianoTransizione50Result represents the computed energy-saving outcome.
type PianoTransizione50Result struct {
	BaselineKWh        float64 `json:"baseline_kwh"`
	PostInterventionKWh float64 `json:"post_intervention_kwh"`
	EnergyReductionPct float64 `json:"energy_reduction_pct"`
	ProcessReductionPct float64 `json:"process_reduction_pct"`
	SiteReductionPct   float64 `json:"site_reduction_pct"`
	MeetsProcessThresh bool    `json:"meets_process_threshold"` // ≥3%
	MeetsSiteThresh    bool    `json:"meets_site_threshold"`    // ≥5%
	TaxCreditBand      string  `json:"tax_credit_band"`         // "5-15%", "15-20%", "35-45%"
	EligibleAmountEUR  float64 `json:"eligible_amount_eur"`
	ExpectedCreditEUR  float64 `json:"expected_credit_eur"`
}

// ESRSE1DataPoint represents a single CSRD data point (abridged).
type ESRSE1DataPoint struct {
	Code        string  `json:"code"` // e.g. E1-5 energy-consumption
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Source      string  `json:"source"`
	Methodology string  `json:"methodology"`
}
