package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type alertsHandler struct{ d Dependencies }

func newAlertsHandler(d Dependencies) *alertsHandler { return &alertsHandler{d: d} }

// List returns current + recent alerts.
func (h *alertsHandler) List(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"items": []any{}, "total": 0})
}

// Ack marks an alert as acknowledged.
func (h *alertsHandler) Ack(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"id":       id,
		"acked_at": time.Now().UTC().Format(time.RFC3339),
		"acked_by": c.Locals("user_email"),
	})
}

// Resolve closes an alert.
func (h *alertsHandler) Resolve(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"id":          id,
		"resolved_at": time.Now().UTC().Format(time.RFC3339),
	})
}
