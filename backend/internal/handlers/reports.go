package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/greenmetrics/backend/internal/models"
)

type reportsHandler struct{ d Dependencies }

func newReportsHandler(d Dependencies) *reportsHandler { return &reportsHandler{d: d} }

// GenerateReportRequest is the payload used to kick off a report.
type GenerateReportRequest struct {
	Type       models.ReportType `json:"type"`
	PeriodFrom time.Time         `json:"period_from"`
	PeriodTo   time.Time         `json:"period_to"`
	Options    map[string]any    `json:"options,omitempty"`
	Render     string            `json:"render,omitempty"` // "json"|"html"
}

// List returns tenant reports (optionally filtered by type).
func (h *reportsHandler) List(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"items": []any{}, "total": 0})
}

// Generate dispatches report generation; the response is the created Report
// resource. When render=html, an HTML body is returned for ESRS E1 and
// Piano 5.0 attestations; other types respond JSON.
func (h *reportsHandler) Generate(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = "placeholder-tenant"
	}
	userEmail, _ := c.Locals("user_email").(string)
	var req GenerateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest("invalid JSON")
	}
	if req.Type == "" || req.PeriodFrom.IsZero() || req.PeriodTo.IsZero() {
		return BadRequest("type, period_from and period_to required")
	}
	if !req.PeriodTo.After(req.PeriodFrom) {
		return BadRequest("period_to must be after period_from")
	}
	rep, err := h.d.Reporter.Generate(c.Context(), tenantID, userEmail, req.Type, req.PeriodFrom, req.PeriodTo, req.Options)
	if err != nil {
		return Internal("report generation failed: " + err.Error())
	}

	if req.Render == "html" {
		switch rep.Type {
		case models.ReportESRSE1:
			b, rerr := h.d.Reporter.RenderESRSE1HTML(rep.Payload)
			if rerr != nil {
				return Internal("esrs_e1 render: " + rerr.Error())
			}
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.Status(fiber.StatusCreated).Send(b)
		case models.ReportPianoTransizione50:
			b, rerr := h.d.Reporter.RenderPianoTransizione50HTML(rep.Payload)
			if rerr != nil {
				return Internal("piano_5_0 render: " + rerr.Error())
			}
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.Status(fiber.StatusCreated).Send(b)
		}
	}
	return c.Status(fiber.StatusCreated).JSON(rep)
}

// Get retrieves a single report.
func (h *reportsHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{"id": id, "placeholder": true})
}

// Download returns the file artefact.
func (h *reportsHandler) Download(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", `attachment; filename="report.pdf"`)
	// Placeholder: real impl streams a rendered PDF from object storage.
	return c.SendString("%PDF-1.4\n%placeholder\n")
}

// Submit marks the report as submitted to the relevant authority (GSE / MASE / ENEA).
func (h *reportsHandler) Submit(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"id":           id,
		"status":       string(models.ReportStatusSubmitted),
		"submitted_at": time.Now().UTC().Format(time.RFC3339),
	})
}
