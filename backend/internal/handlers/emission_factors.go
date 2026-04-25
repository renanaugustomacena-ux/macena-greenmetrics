package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/greenmetrics/backend/internal/models"
)

type emissionFactorsHandler struct{ d Dependencies }

func newEmissionFactorsHandler(d Dependencies) *emissionFactorsHandler {
	return &emissionFactorsHandler{d: d}
}

// List returns the active emission-factor set.
// Seed values (public knowledge from ISPRA 2024 Rapporto 404 and GSE reports).
func (h *emissionFactorsHandler) List(c *fiber.Ctx) error {
	base := []models.EmissionFactor{
		{Code: "IT_ELEC_MIX_2023", Scope: 2, Category: "electricity_mix", Unit: "kWh", KgCO2ePer: 0.250, Source: "ISPRA 2024 Rapporto 404 (Fattori di emissione mix elettrico)", ValidFrom: date(2023, 1, 1), ValidTo: ptrTime(date(2023, 12, 31)), Version: "2024.1"},
		{Code: "IT_ELEC_MIX_2024", Scope: 2, Category: "electricity_mix", Unit: "kWh", KgCO2ePer: 0.245, Source: "ISPRA 2024 Rapporto 404 (stima provvisoria)", ValidFrom: date(2024, 1, 1), Version: "2024.1"},
		{Code: "NG_STATIONARY_COMBUSTION", Scope: 1, Category: "natural_gas", Unit: "Sm3", KgCO2ePer: 1.975, Source: "ISPRA — Tabella parametri standard nazionali (D.M. 11/05/2022)", ValidFrom: date(2022, 5, 11), Version: "2022.1"},
		{Code: "DIESEL_COMBUSTION", Scope: 1, Category: "diesel", Unit: "L", KgCO2ePer: 2.650, Source: "ISPRA 2024 (gasolio autotrazione)", ValidFrom: date(2024, 1, 1), Version: "2024.1"},
		{Code: "LPG_COMBUSTION", Scope: 1, Category: "lpg", Unit: "L", KgCO2ePer: 1.510, Source: "ISPRA 2024", ValidFrom: date(2024, 1, 1), Version: "2024.1"},
		{Code: "HEATING_OIL_COMBUSTION", Scope: 1, Category: "heating_oil", Unit: "L", KgCO2ePer: 2.771, Source: "ISPRA 2024", ValidFrom: date(2024, 1, 1), Version: "2024.1"},
		{Code: "DISTRICT_HEAT_AVERAGE_IT", Scope: 2, Category: "district_heat", Unit: "kWh", KgCO2ePer: 0.200, Source: "Conto Termico 2.0 (GSE) reference average", ValidFrom: date(2024, 1, 1), Version: "2024.1"},
	}
	return c.JSON(fiber.Map{"items": base, "total": len(base)})
}

// Create adds a new factor (admin only in real implementation).
func (h *emissionFactorsHandler) Create(c *fiber.Ctx) error {
	var ef models.EmissionFactor
	if err := c.BodyParser(&ef); err != nil {
		return BadRequest("invalid JSON")
	}
	if ef.Code == "" || ef.Unit == "" || ef.KgCO2ePer <= 0 {
		return BadRequest("code, unit, kg_co2e_per required")
	}
	return c.Status(fiber.StatusCreated).JSON(ef)
}

// Get fetches by code.
func (h *emissionFactorsHandler) Get(c *fiber.Ctx) error {
	code := c.Params("code")
	return c.JSON(fiber.Map{"code": code, "placeholder": true})
}

// Update replaces an existing factor version.
func (h *emissionFactorsHandler) Update(c *fiber.Ctx) error {
	code := c.Params("code")
	var ef models.EmissionFactor
	if err := c.BodyParser(&ef); err != nil {
		return BadRequest("invalid JSON")
	}
	ef.Code = code
	return c.JSON(ef)
}

func date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}
func ptrTime(t time.Time) *time.Time { return &t }
