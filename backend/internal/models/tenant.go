package models

import "time"

// Tenant is a customer organisation.
type Tenant struct {
	ID             string    `json:"id"`
	RagioneSociale string    `json:"ragione_sociale"`
	PartitaIVA     string    `json:"partita_iva"`
	CodiceFiscale  string    `json:"codice_fiscale,omitempty"`
	SDICode        string    `json:"sdi_code,omitempty"`
	PEC            string    `json:"pec,omitempty"`
	ATECO          string    `json:"ateco,omitempty"`
	LargeEnterprise bool     `json:"large_enterprise"` // D.Lgs. 102/2014 obligations apply
	CSRDInScope    bool      `json:"csrd_in_scope"`
	Province       string    `json:"province"` // e.g. "VR"
	Region         string    `json:"region"`
	Plan           string    `json:"plan"` // starter | professionale | enterprise
	MeterQuota     int       `json:"meter_quota"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
