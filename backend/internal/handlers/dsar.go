// Package handlers — GDPR DSAR (Data Subject Access Request) endpoints.
//
// Doctrine refs: Rule 39 (security as structural), Rule 62 (audit logging
// append-only and immutable), Rule 184 (crypto-shredding for Art. 17 erasure).
// Plan ref: docs/PLAN.md §5.4.7 (Sprint S6 GDPR DSAR endpoint).
//
// The endpoints below are **stub implementations** for Phase E:
//   - /api/v1/dsar/{tenant_id}/{user_id}/export — returns an empty bundle
//     with a 202 + Location semantics for asynchronous fulfillment;
//   - /api/v1/dsar/{tenant_id}/{user_id}/erase — accepts the erase request
//     and returns 202 with a job ID; the actual crypto-shredding cascade
//     wires up in Phase F Sprint S10 (DSAR full implementation).
//
// Both endpoints are gated by `security.RequirePermission(security.PermDSARExport)`
// or `PermDSARErase`. Only the DPO role holds these permissions; no other role
// can invoke DSAR.
//
// Every invocation produces an audit-log row (Rule 62) regardless of fulfillment
// state — this is the regulator's evidence that the request was received within
// the GDPR Art. 15 / Art. 17 timeline.
package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// dsarHandler holds the GDPR DSAR endpoints. Composed via Dependencies.
type dsarHandler struct{ d Dependencies }

func newDSARHandler(d Dependencies) *dsarHandler { return &dsarHandler{d: d} }

// DSARExportResponse is the body returned from `POST /dsar/{tenant_id}/{user_id}/export`.
//
// Phase E ships the response shape and the audit-trail discipline; Phase F Sprint S10
// populates `bundle_url` with a signed S3 URL pointing at the actual export zip.
type DSARExportResponse struct {
	JobID        uuid.UUID `json:"job_id"`
	TenantID     string    `json:"tenant_id"`
	UserID       string    `json:"user_id"`
	State        string    `json:"state"`         // "queued" in Phase E; later "ready" / "failed"
	BundleURL    string    `json:"bundle_url,omitempty"`
	RequestedAt  time.Time `json:"requested_at"`
	SLADeadline  time.Time `json:"sla_deadline"`  // GDPR Art. 12 §3 — 30-day window
	Note         string    `json:"note,omitempty"`
}

// DSAREraseResponse is the body returned from `POST /dsar/{tenant_id}/{user_id}/erase`.
type DSAREraseResponse struct {
	JobID         uuid.UUID `json:"job_id"`
	TenantID      string    `json:"tenant_id"`
	UserID        string    `json:"user_id"`
	State         string    `json:"state"`        // "queued" in Phase E
	RequestedAt   time.Time `json:"requested_at"`
	SLADeadline   time.Time `json:"sla_deadline"` // GDPR Art. 12 §3 — 30-day window
	CryptoShredAt time.Time `json:"crypto_shred_at,omitempty"`
	Note          string    `json:"note,omitempty"`
}

// Export handles `POST /api/v1/dsar/{tenant_id}/{user_id}/export`.
//
// Phase E: returns 202 with a queued JobID + 30-day SLA deadline. Audit row
// is recorded by AuditMiddleware. The actual data-bundling logic (Postgres
// dump per tenant_id WHERE user attribution + readings + reports + audit log
// + factors used) lands in Phase F Sprint S10.
func (h *dsarHandler) Export(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	userID := c.Params("user_id")
	if err := validateDSARParams(tenantID, userID); err != nil {
		return err
	}

	now := time.Now().UTC()
	resp := DSARExportResponse{
		JobID:       uuid.New(),
		TenantID:    tenantID,
		UserID:      userID,
		State:       "queued",
		RequestedAt: now,
		SLADeadline: now.AddDate(0, 0, 30), // GDPR Art. 12 §3
		Note:        "DSAR export queued. Phase E stub: bundling lands in Phase F Sprint S10.",
	}
	c.Status(fiber.StatusAccepted)
	c.Set("Location", "/api/v1/dsar/jobs/"+resp.JobID.String())
	return c.JSON(resp)
}

// Erase handles `POST /api/v1/dsar/{tenant_id}/{user_id}/erase`.
//
// Phase E: returns 202 with a queued JobID. Phase F Sprint S10 wires the
// crypto-shredding cascade (per-tenant DEK rotation per Rule 184) +
// pseudonymisation of PII columns +  audit-row preservation (the audit log
// itself is NOT erased per Rule 62).
func (h *dsarHandler) Erase(c *fiber.Ctx) error {
	tenantID := c.Params("tenant_id")
	userID := c.Params("user_id")
	if err := validateDSARParams(tenantID, userID); err != nil {
		return err
	}

	now := time.Now().UTC()
	resp := DSAREraseResponse{
		JobID:       uuid.New(),
		TenantID:    tenantID,
		UserID:      userID,
		State:       "queued",
		RequestedAt: now,
		SLADeadline: now.AddDate(0, 0, 30),
		Note:        "DSAR erase queued. Phase E stub: crypto-shredding cascade lands in Phase F Sprint S10.",
	}
	c.Status(fiber.StatusAccepted)
	c.Set("Location", "/api/v1/dsar/jobs/"+resp.JobID.String())
	return c.JSON(resp)
}

// validateDSARParams sanity-checks the URL parameters before producing
// a job ID. Both must be valid UUIDv4 (Rule 3 + tenant invariant).
func validateDSARParams(tenantID, userID string) error {
	if _, err := uuid.Parse(tenantID); err != nil {
		return BadRequest("tenant_id must be a UUIDv4")
	}
	if _, err := uuid.Parse(userID); err != nil {
		return BadRequest("user_id must be a UUIDv4")
	}
	return nil
}
