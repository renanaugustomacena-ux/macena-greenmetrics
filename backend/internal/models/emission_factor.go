package models

import "time"

// EmissionFactor is a versioned conversion factor (value × factor = kg CO2e).
//
// Sources supported:
//   - ISPRA (Italian national GHG inventory; updated annually).
//   - GSE / GME (electricity residual mix).
//   - IPCC AR6.
//   - EcoInvent (Scope 3 supplier-side, placeholder).
type EmissionFactor struct {
	Code       string     `json:"code"`        // e.g. "IT_ELEC_MIX_2023"
	Scope      int        `json:"scope"`       // 1 | 2 | 3
	Category   string     `json:"category"`    // "electricity_mix", "natural_gas", "diesel"...
	Unit       string     `json:"unit"`        // "kWh", "m3", "L", "kg"
	KgCO2ePer  float64    `json:"kg_co2e_per"` // e.g. 0.250 for 2023 Italian mix
	Source     string     `json:"source"`      // "ISPRA 2024 Rapporto 404"
	ValidFrom  time.Time  `json:"valid_from"`
	ValidTo    *time.Time `json:"valid_to,omitempty"`
	Version    string     `json:"version"` // e.g. "2024.1"
	Notes      string     `json:"notes,omitempty"`
}
