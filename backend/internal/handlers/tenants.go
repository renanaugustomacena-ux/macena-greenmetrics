package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/greenmetrics/backend/internal/models"
)

type tenantsHandler struct{ d Dependencies }

func newTenantsHandler(d Dependencies) *tenantsHandler { return &tenantsHandler{d: d} }

// GetSelf returns the authenticated user's tenant.
func (h *tenantsHandler) GetSelf(c *fiber.Ctx) error {
	return c.JSON(models.Tenant{
		ID:              "placeholder-tenant",
		RagioneSociale:  "Industria Esempio S.r.l.",
		PartitaIVA:      "IT01234567890",
		ATECO:           "10.89.09",
		Province:        "VR",
		Region:          "Veneto",
		LargeEnterprise: false,
		CSRDInScope:     false,
		Plan:            "professionale",
		MeterQuota:      50,
		Active:          true,
		CreatedAt:       time.Now().UTC(),
	})
}

// UpdateSelf edits the authenticated user's tenant.
func (h *tenantsHandler) UpdateSelf(c *fiber.Ctx) error {
	var t models.Tenant
	if err := c.BodyParser(&t); err != nil {
		return BadRequest("invalid JSON")
	}
	t.UpdatedAt = time.Now().UTC()
	return c.JSON(t)
}
