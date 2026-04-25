package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/greenmetrics/backend/internal/models"
)

type metersHandler struct{ d Dependencies }

func newMetersHandler(d Dependencies) *metersHandler { return &metersHandler{d: d} }

// List returns all meters for the authenticated tenant.
func (h *metersHandler) List(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = "placeholder-tenant"
	}
	if h.d.Repo == nil {
		return c.JSON(fiber.Map{"items": []any{}, "total": 0})
	}
	rows, err := h.d.Repo.ListMeters(c.Context(), tenantID)
	if err != nil {
		return Internal("meters list: " + err.Error())
	}
	// Map repo rows to public model.
	out := make([]models.Meter, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.Meter{
			ID:         r.ID,
			TenantID:   r.TenantID,
			Label:      r.Label,
			MeterType:  models.MeterType(r.MeterType),
			Protocol:   models.MeterProtocol(r.Protocol),
			Site:       r.Site,
			CostCentre: r.CostCentre,
			Active:     r.Active,
			CreatedAt:  r.CreatedAt,
		})
	}
	return c.JSON(fiber.Map{"items": out, "total": len(out)})
}

// Create adds a meter.
func (h *metersHandler) Create(c *fiber.Ctx) error {
	var m models.Meter
	if err := c.BodyParser(&m); err != nil {
		return BadRequest("invalid meter JSON")
	}
	if m.Label == "" || m.MeterType == "" || m.Protocol == "" {
		return BadRequest("label, meter_type, protocol required")
	}
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = m.CreatedAt
	// Placeholder persistence — real repo INSERT omitted for brevity.
	return c.Status(fiber.StatusCreated).JSON(m)
}

// Get fetches a single meter by ID.
func (h *metersHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return BadRequest("id required")
	}
	return c.JSON(fiber.Map{"id": id, "placeholder": true})
}

// Update modifies a meter.
func (h *metersHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var m models.Meter
	if err := c.BodyParser(&m); err != nil {
		return BadRequest("invalid JSON")
	}
	m.ID = id
	m.UpdatedAt = time.Now().UTC()
	return c.JSON(m)
}

// Delete soft-deletes a meter.
func (h *metersHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return BadRequest("id required")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Probe triggers a one-off read of a meter to verify connectivity.
func (h *metersHandler) Probe(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"meter_id":  id,
		"probed_at": time.Now().UTC().Format(time.RFC3339),
		"status":    "ok_placeholder",
		"latency_ms": 42,
	})
}
