package services

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// AuditClient prepares the D.Lgs. 102/2014 audit energetico dossier and
// submits to ENEA's audit102 portal (placeholder).
//
// Obligation: every 4 years for large enterprises (>250 employees or
// revenue >50M€ / balance sheet >43M€) plus energivore registered with CSEA.
// Deadline: 5 December of the reporting year; methodology per EN 16247-1/2/3/4.
type AuditClient struct {
	logger *zap.Logger
}

// NewAuditClient creates the client.
func NewAuditClient(logger *zap.Logger) *AuditClient {
	return &AuditClient{logger: logger}
}

// AuditDossier is the structured output presented for review before submission.
type AuditDossier struct {
	CompanyName      string            `json:"company_name"`
	CompanyVAT       string            `json:"company_vat"`
	Sector           string            `json:"sector"`              // ATECO
	SiteCount        int               `json:"site_count"`
	ReportingPeriod  time.Time         `json:"reporting_period"`    // year of reference
	TotalEnergyKWh   float64           `json:"total_energy_kwh"`
	Breakdown        map[string]float64 `json:"breakdown"`          // vector → kWh
	IPE              float64           `json:"ipe"`                 // kWh / unit output
	Auditor          string            `json:"auditor"`             // EGE / ESCo
	AuditorCert      string            `json:"auditor_certification"` // UNI CEI 11339 / SECEM
	SubmittedAt      *time.Time        `json:"submitted_at,omitempty"`
	PortalRef        string            `json:"portal_ref,omitempty"`
}

// Prepare builds an audit dossier skeleton based on tenant data.
func (c *AuditClient) Prepare(ctx context.Context, tenantID string) (*AuditDossier, error) {
	// Placeholder: real implementation aggregates 3+ years from hypertables
	// and computes IPE per production cost-centre.
	return &AuditDossier{
		CompanyName:     "Industria Esempio S.r.l.",
		CompanyVAT:      "IT01234567890",
		Sector:          "ATECO 10.89.09",
		SiteCount:       1,
		ReportingPeriod: time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC),
		Breakdown: map[string]float64{
			"electricity_kwh": 1_250_000,
			"natural_gas_sm3": 320_000,
			"district_heat_kwh": 0,
		},
		Auditor:      "EGE certificato — placeholder",
		AuditorCert:  "UNI CEI 11339",
	}, nil
}

// Submit uploads the dossier to ENEA audit102 portal.
// Placeholder: real implementation uses ENEA-provided API token + multipart upload.
func (c *AuditClient) Submit(ctx context.Context, dossier *AuditDossier) error {
	now := time.Now().UTC()
	dossier.SubmittedAt = &now
	dossier.PortalRef = "ENEA-AUDIT102-PLACEHOLDER-REF"
	c.logger.Info("audit dossier submitted (stub)",
		zap.String("vat", dossier.CompanyVAT),
		zap.String("portal_ref", dossier.PortalRef),
	)
	return nil
}
