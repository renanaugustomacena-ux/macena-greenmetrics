package models

// ScopeCategory encodes the 15 GHG-Protocol Scope 3 categories plus Scope 1/2 buckets.
type ScopeCategory string

const (
	// Scope 1 — direct.
	ScopeStationaryCombustion ScopeCategory = "s1_stationary_combustion"
	ScopeMobileCombustion     ScopeCategory = "s1_mobile_combustion"
	ScopeFugitiveEmissions    ScopeCategory = "s1_fugitive_emissions"
	ScopeProcessEmissions     ScopeCategory = "s1_process_emissions"

	// Scope 2 — purchased electricity / heat.
	Scope2ElectricityLocationBased ScopeCategory = "s2_electricity_location_based"
	Scope2ElectricityMarketBased   ScopeCategory = "s2_electricity_market_based"
	Scope2HeatSteamCooling         ScopeCategory = "s2_heat_steam_cooling"

	// Scope 3 — value chain (15 categories per GHG Protocol).
	Scope3Cat1PurchasedGoods        ScopeCategory = "s3_cat1_purchased_goods"
	Scope3Cat2CapitalGoods          ScopeCategory = "s3_cat2_capital_goods"
	Scope3Cat3FuelEnergyRelated     ScopeCategory = "s3_cat3_fuel_energy_related"
	Scope3Cat4UpstreamTransport     ScopeCategory = "s3_cat4_upstream_transport"
	Scope3Cat5WasteOperations       ScopeCategory = "s3_cat5_waste_operations"
	Scope3Cat6BusinessTravel        ScopeCategory = "s3_cat6_business_travel"
	Scope3Cat7EmployeeCommuting     ScopeCategory = "s3_cat7_employee_commuting"
	Scope3Cat8UpstreamLeasedAssets  ScopeCategory = "s3_cat8_upstream_leased_assets"
	Scope3Cat9DownstreamTransport   ScopeCategory = "s3_cat9_downstream_transport"
	Scope3Cat10ProcessingSoldProd   ScopeCategory = "s3_cat10_processing_sold_products"
	Scope3Cat11UseSoldProd          ScopeCategory = "s3_cat11_use_sold_products"
	Scope3Cat12EoLSoldProd          ScopeCategory = "s3_cat12_end_of_life_sold_products"
	Scope3Cat13DownstreamLeasedAss  ScopeCategory = "s3_cat13_downstream_leased_assets"
	Scope3Cat14Franchises           ScopeCategory = "s3_cat14_franchises"
	Scope3Cat15Investments          ScopeCategory = "s3_cat15_investments"
)
