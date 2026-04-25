package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/greenmetrics/backend/internal/models"
	"github.com/greenmetrics/backend/internal/repository"
)

// CarbonCalculator translates energy consumption into CO2e emissions using
// versioned emission factors from ISPRA / GSE / IPCC / EcoInvent.
type CarbonCalculator struct {
	repo   *repository.TimescaleRepository
	logger *zap.Logger

	// Cached factors — real implementation refreshes from ISPRA_EMISSION_FACTORS_URL.
	cache map[string]*models.EmissionFactor
}

// NewCarbonCalculator builds the calculator.
func NewCarbonCalculator(repo *repository.TimescaleRepository, logger *zap.Logger) *CarbonCalculator {
	return &CarbonCalculator{
		repo:   repo,
		logger: logger,
		cache:  defaultFactors(),
	}
}

// ScopeTotals is the aggregate CO2e result for a window.
type ScopeTotals struct {
	Scope1KgCO2e     float64            `json:"scope1_kg_co2e"`
	Scope2KgCO2e     float64            `json:"scope2_kg_co2e"`
	Scope3KgCO2e     float64            `json:"scope3_kg_co2e"`
	TotalKgCO2e      float64            `json:"total_kg_co2e"`
	ByCategoryKg     map[string]float64 `json:"by_category_kg"`
	FactorsUsed      []string           `json:"factors_used"`
	MethodologyNotes []string           `json:"methodology_notes"`
	Window           struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
	} `json:"window"`
}

// Scope1FromGas converts Sm3 of natural gas into kg CO2e using ISPRA standard factor.
func (c *CarbonCalculator) Scope1FromGas(volumeSm3 float64, at time.Time) (float64, *models.EmissionFactor) {
	f := c.factor("NG_STATIONARY_COMBUSTION", at)
	return volumeSm3 * f.KgCO2ePer, f
}

// Scope1FromDiesel converts litres of diesel into kg CO2e.
func (c *CarbonCalculator) Scope1FromDiesel(litres float64, at time.Time) (float64, *models.EmissionFactor) {
	f := c.factor("DIESEL_COMBUSTION", at)
	return litres * f.KgCO2ePer, f
}

// Scope2FromElectricity converts kWh into kg CO2e using the Italian mix
// factor valid at `at`. Uses location-based accounting by default.
func (c *CarbonCalculator) Scope2FromElectricity(kWh float64, at time.Time, marketBased bool) (float64, *models.EmissionFactor) {
	code := "IT_ELEC_MIX_2023"
	if at.Year() >= 2024 {
		code = "IT_ELEC_MIX_2024"
	}
	if marketBased {
		code = "IT_ELEC_RESIDUAL_MIX_" + year(at)
	}
	f := c.factor(code, at)
	return kWh * f.KgCO2ePer, f
}

// Scope3FromSupplier is a placeholder for supplier-side emissions (EcoInvent).
// Real implementation: look up EcoInvent process IDs × supplier-spend.
func (c *CarbonCalculator) Scope3FromSupplier(spendEUR float64, category string) float64 {
	// Monetary-EF placeholder (kg CO2e / €) drawn from EEIO databases.
	monetaryFactors := map[string]float64{
		"materials_machined":   0.60,
		"logistics_road":       0.45,
		"electronics":          0.85,
		"food_processing":      0.40,
		"services_consultancy": 0.12,
	}
	f, ok := monetaryFactors[category]
	if !ok {
		f = 0.50
	}
	return spendEUR * f
}

// Compute calculates aggregate scope totals for a tenant over a window.
//
// The pragmatic "minimum viable" flow:
//  1. Query electricity meters → Scope 2 (location-based).
//  2. Query gas meters → Scope 1 stationary combustion.
//  3. Apply versioned factors valid at midpoint of the window.
//  4. Accept Scope 3 supplier spend as a per-tenant input (not yet metered).
func (c *CarbonCalculator) Compute(ctx context.Context, tenantID string, from, to time.Time) (*ScopeTotals, error) {
	totals := &ScopeTotals{
		ByCategoryKg: map[string]float64{},
		FactorsUsed:  []string{},
		MethodologyNotes: []string{
			"Scope 2 accounted on location-based method (default) using ISPRA Italian mix factor.",
			"Scope 1 stationary combustion assumes ISPRA D.M. 11/05/2022 parameters for natural gas (1.975 kg CO2e / Sm3).",
			"Scope 3 placeholder uses EEIO monetary factors; production implementation wires EcoInvent.",
		},
	}
	totals.Window.From = from
	totals.Window.To = to
	midpoint := from.Add(to.Sub(from) / 2)

	// Electricity (kWh) from 1-day aggregate — placeholder meter walk.
	meters, err := c.repo.ListMeters(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	for _, m := range meters {
		switch m.MeterType {
		case "electricity", "electricity_3p":
			rows, err := c.repo.QueryAggregated(ctx, tenantID, m.ID, "1d", from, to)
			if err != nil {
				continue
			}
			var kwh float64
			for _, r := range rows {
				kwh += r.SumValue
			}
			kg, f := c.Scope2FromElectricity(kwh, midpoint, false)
			totals.Scope2KgCO2e += kg
			totals.ByCategoryKg["electricity_"+m.Site] += kg
			totals.FactorsUsed = appendUniq(totals.FactorsUsed, f.Code)
		case "gas":
			rows, err := c.repo.QueryAggregated(ctx, tenantID, m.ID, "1d", from, to)
			if err != nil {
				continue
			}
			var sm3 float64
			for _, r := range rows {
				sm3 += r.SumValue
			}
			kg, f := c.Scope1FromGas(sm3, midpoint)
			totals.Scope1KgCO2e += kg
			totals.ByCategoryKg["gas_"+m.Site] += kg
			totals.FactorsUsed = appendUniq(totals.FactorsUsed, f.Code)
		}
	}
	totals.TotalKgCO2e = totals.Scope1KgCO2e + totals.Scope2KgCO2e + totals.Scope3KgCO2e
	return totals, nil
}

func (c *CarbonCalculator) factor(code string, at time.Time) *models.EmissionFactor {
	if ef, ok := c.cache[code]; ok {
		return ef
	}
	// Fallback: neutral factor, warn.
	c.logger.Warn("emission factor not found in cache", zap.String("code", code))
	return &models.EmissionFactor{Code: code, KgCO2ePer: 0, Source: "placeholder"}
}

func defaultFactors() map[string]*models.EmissionFactor {
	return map[string]*models.EmissionFactor{
		"IT_ELEC_MIX_2023":         {Code: "IT_ELEC_MIX_2023", Scope: 2, Unit: "kWh", KgCO2ePer: 0.250, Source: "ISPRA 2024 Rapporto 404"},
		"IT_ELEC_MIX_2024":         {Code: "IT_ELEC_MIX_2024", Scope: 2, Unit: "kWh", KgCO2ePer: 0.245, Source: "ISPRA 2024 stima provvisoria"},
		"IT_ELEC_RESIDUAL_MIX_2023": {Code: "IT_ELEC_RESIDUAL_MIX_2023", Scope: 2, Unit: "kWh", KgCO2ePer: 0.457, Source: "AIB European Residual Mix 2023 (market-based)"},
		"NG_STATIONARY_COMBUSTION": {Code: "NG_STATIONARY_COMBUSTION", Scope: 1, Unit: "Sm3", KgCO2ePer: 1.975, Source: "ISPRA D.M. 11/05/2022"},
		"DIESEL_COMBUSTION":        {Code: "DIESEL_COMBUSTION", Scope: 1, Unit: "L", KgCO2ePer: 2.650, Source: "ISPRA 2024"},
		"LPG_COMBUSTION":           {Code: "LPG_COMBUSTION", Scope: 1, Unit: "L", KgCO2ePer: 1.510, Source: "ISPRA 2024"},
		"HEATING_OIL_COMBUSTION":   {Code: "HEATING_OIL_COMBUSTION", Scope: 1, Unit: "L", KgCO2ePer: 2.771, Source: "ISPRA 2024"},
		"DISTRICT_HEAT_AVERAGE_IT": {Code: "DISTRICT_HEAT_AVERAGE_IT", Scope: 2, Unit: "kWh", KgCO2ePer: 0.200, Source: "GSE Conto Termico 2.0 reference"},
	}
}

func year(t time.Time) string {
	s := time.Time(t).UTC().Format("2006")
	return s
}

func appendUniq(xs []string, s string) []string {
	for _, x := range xs {
		if x == s {
			return xs
		}
	}
	return append(xs, s)
}
