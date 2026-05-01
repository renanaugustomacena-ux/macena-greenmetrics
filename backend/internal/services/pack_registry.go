// Package services hosts the Pack registry — the runtime lookup table
// that maps Pack identifiers to their concrete instances.
//
// The registry is the seed of REJ-11 (god-service decomposition): it
// makes the 12 Italian-flagship Packs discoverable by id at runtime
// without committing to the full Phase F migration of report_generator.go
// onto the Pack-loader's dispatch path. ReportGenerator.Generate
// continues to use its inline build* functions for v1.0.0; Phase F
// replaces those with delegation to the registry.
//
// Doctrine refs: Rules 73 (manifest lock), 75 (capabilities declared,
// not discovered), 87 (no god service — this registry is a passive
// lookup, not a god service), 88 (Region Pack flagship pattern).
// Charter ref: §4 (Pack contract), §11 (single-tenant by default).
//
// The registry is intentionally a thin façade over package-level Pack
// globals. Each Pack package exports `Pack` of the appropriate contract
// type (reporting.Builder, emissions.FactorSource, region.RegionProfile);
// importing the package is sufficient to make the global available, and
// the registry simply collects references for lookup.
package services

import (
	"github.com/greenmetrics/backend/internal/domain/emissions"
	"github.com/greenmetrics/backend/internal/domain/region"
	"github.com/greenmetrics/backend/internal/domain/reporting"

	// Region Packs (1).
	regionIT "github.com/greenmetrics/backend/packs/region/it"

	// Factor Packs (4).
	factorAIB "github.com/greenmetrics/backend/packs/factor/aib"
	factorGSE "github.com/greenmetrics/backend/packs/factor/gse"
	factorISPRA "github.com/greenmetrics/backend/packs/factor/ispra"
	factorTerna "github.com/greenmetrics/backend/packs/factor/terna"

	// Report Packs (7).
	reportAuditDLgs102 "github.com/greenmetrics/backend/packs/report/audit_dlgs102"
	reportCO2Footprint "github.com/greenmetrics/backend/packs/report/co2_footprint"
	reportContoTermico "github.com/greenmetrics/backend/packs/report/conto_termico"
	reportESRSE1 "github.com/greenmetrics/backend/packs/report/esrs_e1"
	reportMonthlyConsumption "github.com/greenmetrics/backend/packs/report/monthly_consumption"
	reportPiano50 "github.com/greenmetrics/backend/packs/report/piano_5_0"
	reportTEE "github.com/greenmetrics/backend/packs/report/tee"
)

// ReportPacks returns the registry of Report Packs keyed by their declared
// ReportType. The map is constructed fresh on each call (defensive copy);
// callers may mutate without affecting the canonical set.
//
// The 7 Italian-flagship Report Packs are registered as of Phase E Sprint S6:
// monthly_consumption, co2_footprint, esrs_e1, piano_5_0, audit_dlgs102,
// tee, conto_termico.
func ReportPacks() map[reporting.ReportType]reporting.Builder {
	packs := []reporting.Builder{
		reportMonthlyConsumption.Pack,
		reportCO2Footprint.Pack,
		reportESRSE1.Pack,
		reportPiano50.Pack,
		reportAuditDLgs102.Pack,
		reportTEE.Pack,
		reportContoTermico.Pack,
	}
	out := make(map[reporting.ReportType]reporting.Builder, len(packs))
	for _, p := range packs {
		out[p.Type()] = p
	}
	return out
}

// FactorPacks returns the registry of Factor Packs keyed by their declared
// Name. The 4 Italian-flagship Factor Packs are registered as of Phase E
// Sprint S6: ispra, gse, terna, aib.
func FactorPacks() map[string]emissions.FactorSource {
	packs := []emissions.FactorSource{
		factorISPRA.Pack,
		factorGSE.Pack,
		factorTerna.Pack,
		factorAIB.Pack,
	}
	out := make(map[string]emissions.FactorSource, len(packs))
	for _, p := range packs {
		out[p.Name()] = p
	}
	return out
}

// RegionPacks returns the registry of Region Packs keyed by their declared
// Code. The 1 Italian-flagship Region Pack is registered as of Phase E
// Sprint S6: region-it.
func RegionPacks() map[string]region.RegionProfile {
	packs := []region.RegionProfile{
		regionIT.Pack,
	}
	out := make(map[string]region.RegionProfile, len(packs))
	for _, p := range packs {
		out[p.Code()] = p
	}
	return out
}

// PackCount returns the (region, factor, report) tuple of registered Pack
// counts for the runtime health envelope (Rule 74).
func PackCount() (region, factor, report int) {
	return len(RegionPacks()), len(FactorPacks()), len(ReportPacks())
}
