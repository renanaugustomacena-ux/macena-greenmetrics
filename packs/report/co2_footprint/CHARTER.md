# CO₂ Footprint Report Pack — Charter

> **Pack:** `report-co2_footprint` · **Kind:** Report · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA
> **Doctrine refs:** Rules 89 (bit-perfect reproducibility), 91 (pure functions), 95 (provenance bundle), 97 (algorithm versioning), 141 (deterministic serialisation).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Report Packs).

## 1. Purpose

The CO₂ Footprint Report Pack produces an **operator-dashboard carbon summary** for an Italian-flagship engagement deployment. Its scope: Scope 1 (combustion) + Scope 2 (location-based + market-based, dual-method per ESRS E1 best practice). Scope 3 inputs flow through a separate ingestion path (Sprint S15 deliverable) and are surfaced as zero plus a documented note for now.

This is the **second Italian Report Pack reference** (companion to `report-monthly_consumption`). It exercises the multi-factor lookup pattern — same Builder consults both ISPRA (location-based grid mix + Scope 1 combustion) and GSE (AIB residual mix for Scope 2 market-based).

## 2. Algorithm

Pure function `Build(ctx, period, factors, readings) → Report`.

```text
1. Iterate `readings` (pre-aggregated by Core's reporting orchestrator).
2. For each row, classify by Unit:
     - "Wh" / "kWh"      → Scope 2 electricity
     - "Sm3"             → Scope 1 natural_gas_combustion
     - "l_diesel"        → Scope 1 diesel_combustion
     - "l_petrol"        → Scope 1 petrol_road_vehicle
     - "kg_lpg"          → Scope 1 lpg_combustion
     - "kg_coal"         → Scope 1 coal_combustion
     - "kg_heavy_fuel"   → Scope 1 heavy_fuel_oil_combustion
     - "l_diesel_vehicle"→ Scope 1 diesel_road_vehicle (mobile)
     - any other unit    → unclassified (counts in Notes; not in scope totals)
3. Sum quantities by classification.
4. Convert Wh → kWh; look up factors at period midpoint:
     - location-based: factors.Get("it_grid_mix_location")
     - market-based:   factors.Get("it_aib_residual_mix")
   Compute Scope 2 emissions both ways; emit nil for the missing method.
5. Look up Scope 1 factors per fuel type from the same factor bundle.
6. Render canonical typed Body — Scope 1 per source + Scope 2 dual + grand totals.
7. Serialise deterministically (Rule 141): JSON, sorted struct fields,
   2-space indent, trailing newline.
8. Provenance is populated by Core's reporting orchestrator at signed-state.
```

The function is **pure** (Rule 91). The conformance test asserts byte-identical output across two consecutive `Build` calls (Rule 89).

## 3. Body shape

```json
{
  "period":     { ... },
  "factors_used": {
    "it_grid_mix_location":   { "value": 233.0, "unit": "g CO2eq/kWh", "version": "..." },
    "it_aib_residual_mix":    { "value": 332.0, "unit": "g CO2eq/kWh", "version": "..." },
    "natural_gas_combustion": { "value": 1.967, "unit": "kg CO2eq/Sm3", "version": "..." }
  },
  "scope_1": {
    "kg_co2eq_total": 12345.6,
    "per_source": [
      { "code": "natural_gas_combustion", "quantity": 6275.0, "unit": "Sm3", "kg_co2eq": 12343.0 }
    ]
  },
  "scope_2": {
    "kwh_total": 128430.0,
    "kg_co2eq_location_based": 29904.2,
    "kg_co2eq_market_based":   42638.8
  },
  "scope_3": {
    "kg_co2eq_total": 0,
    "note": "Scope 3 inputs flow through a separate ingestion path; zero by default until populated."
  },
  "totals": {
    "kg_co2eq_location_based": 42249.8,
    "kg_co2eq_market_based":   54984.4
  },
  "unclassified_rows": 0
}
```

## 4. Tradeoff Stanza

- **Solves:** the operator-facing carbon summary that the CSRD wave-2/3 client is being held accountable for. Dual-method Scope 2 (location + market) matches ESRS E1 best practice. Per-source Scope 1 breakdown surfaces the highest-impact mitigation levers to the engagement client.
- **Optimises for:** clarity (scope semantics map 1:1 to GHG Protocol Corporate Standard), interoperability (shared factor lookups with `report-esrs_e1` + `report-piano_5_0`), audit-grade reproducibility (Rule 89 byte-identical replay).
- **Sacrifices:** Scope 3 depth — purchased goods, transport, waste flow through a Sprint S15 ingestion path; this Pack emits zero with a documented note. Unit classification is keyed off `AggregatedRow.Unit` (the ingest layer is the canonical normaliser).
- **Residual risks:** missing-factor degradation — if the FactorBundle does not include both `it_grid_mix_location` and `it_aib_residual_mix`, only the present one populates; the other is `nil` with a note. Test enforces graceful behaviour.

## 5. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/report/co2_footprint/`. Pack manifest + CHARTER stay at the repo-root `packs/report/co2_footprint/` per Charter §3.2 discovery convention. See ADR-0024.

## 6. Cross-references

- Pack contract: `backend/internal/domain/reporting/builder.go`.
- Pack manifest: `packs/report/co2_footprint/manifest.yaml`.
- Implementation: `backend/packs/report/co2_footprint/builder.go`.
- Tests: `backend/packs/report/co2_footprint/builder_test.go`.
- Factor dependencies: `packs/factor/{ispra,gse}/`.
- Sister Italian Report Packs: `packs/report/{esrs_e1,piano_5_0,conto_termico,tee,audit_dlgs102,monthly_consumption}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 89, 91, 95, 97, 141.
