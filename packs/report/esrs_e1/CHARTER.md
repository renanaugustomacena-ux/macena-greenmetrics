# ESRS E1 — Climate Change Report Pack — Charter

> **Pack:** `report-esrs_e1` · **Kind:** Report · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA (regulatory centrepiece for CSRD wave-2 / wave-3 engagements)
> **Doctrine refs:** Rules 89 (bit-perfect reproducibility), 91 (pure functions), 95 (provenance bundle), 97 (algorithm versioning), 131 (formal-spec validation), 137 (regulatory counter-signature), 141 (deterministic serialisation), 144 (signed at finalisation).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Report Packs).
> **Regulatory base:** EFRAG ESRS E1 Climate Change — adopted via Reg. Delegato (UE) 2023/2772 §1; Italian transposition D.Lgs. 6 settembre 2024 n. 125.

## 1. Purpose

This Pack produces the **quantitative ESRS E1 Climate Change disclosures** for an Italian-flagship engagement subject to CSRD reporting obligations (waves 2 and 3 — circa 8 000 + 35 000 Italian entities). Specifically it covers:

- **E1-5** — Energy consumption and mix
- **E1-6** — Gross Scope 1, 2, 3 and Total GHG emissions

These two data-point families are the auditable, instrument-driven half of ESRS E1; everything else (E1-1 transition plan, E1-2 policies, E1-3 actions, E1-4 targets, E1-7 GHG removals, E1-8 internal carbon pricing, E1-9 anticipated financial effects) is **narrative content supplied by the engagement client** and bundled into the final dossier by the engagement-fork's reporting orchestrator (Plan §5.4 + Sprint S15 deliverable).

Per Rule 131, the Pack's output validates against the EFRAG XBRL Taxonomy at finalisation. The XBRL-tagging path lands in Phase H Sprint S15; this Phase E version emits the canonical typed payload + deterministic JSON encoding that the tagger reads.

Per Rule 137, dossiers above thresholds typical for the engagement client (1× CSRD wave-2 entity) require an EGE counter-signature; the Pack's CHARTER declares this dependency and the engagement Phase 0 Discovery verifies the EGE relationship is in place before generation.

## 2. Algorithm

Pure function `Build(ctx, period, factors, readings) → Report`.

```text
1. Classify readings by Unit (same convention as co2_footprint Pack):
     - Wh / kWh         → Scope 2 electricity
     - Sm3              → Scope 1 natural gas
     - l_diesel         → Scope 1 stationary diesel
     - l_petrol         → Scope 1 mobile petrol
     - l_diesel_vehicle → Scope 1 mobile diesel
     - kg_lpg / kg_coal / kg_heavy_fuel → Scope 1
     - any other unit   → unclassified (counts in Notes; not in scope totals)
2. Compute E1-5 quantities:
     - Total energy consumption (MWh) = ∑ kWh equivalent across all classified
       sources. Per-fuel kWh uses ISPRA-published lower heating values (LHV)
       which are not currently in the FactorBundle; for Phase E, only the
       electricity portion is summed in MWh; non-electricity sources are
       reported as primary quantity (Sm³ for gas, l for liquid fuels, kg for
       solid fuels) plus a Notes line about the LHV deferral.
     - Renewable share — read from factors `it_renewable_share` (GSE).
     - Non-renewable share — derived as 100 - renewable_share.
3. Compute E1-6 quantities:
     - Scope 1 per source: ∑ quantity × per-fuel factor (ISPRA Pack).
     - Scope 1 total: ∑ per-source kg.
     - Scope 2 location-based: electricity_kWh × it_grid_mix_location / 1000.
     - Scope 2 market-based:    electricity_kWh × it_aib_residual_mix / 1000.
     - Scope 3: placeholder (zero) until the Sprint S15 ingestion path lands.
     - Total (location-based) = Scope1 + Scope2.location + Scope3.
     - Total (market-based)   = Scope1 + Scope2.market + Scope3.
4. Render the canonical typed Body (E1-5 block + E1-6 block + factors_used + notes).
5. Serialise deterministically (Rule 141): JSON with sorted struct fields,
   2-space indent, trailing newline.
6. Provenance is populated by Core's reporting orchestrator at signed-state
   transition (Rule 144). The Pack does NOT call out to a regulator portal;
   submission to the EFRAG / ENEA filing surface is a separate engagement-side
   Pack ("submitter") landing in Phase H Sprint S16.
```

The function is **pure** (Rule 91). The conformance test asserts byte-identical Encoded across two consecutive `Build` calls (Rule 89).

## 3. Output shape

```json
{
  "report":     "esrs_e1",
  "regulator":  "EFRAG / CSRD wave 2",
  "period":     { ... },
  "factors_used": {
    "it_grid_mix_location":   { "value": 233, "unit": "g CO2eq/kWh", "version": "..." },
    "it_aib_residual_mix":    { "value": 332, "unit": "g CO2eq/kWh", "version": "..." },
    "it_renewable_share":     { "value": 44.5, "unit": "%",         "version": "..." },
    "natural_gas_combustion": { "value": 1.967, "unit": "kg CO2eq/Sm3", "version": "..." }
  },
  "e1_5_energy_consumption": {
    "electricity_mwh":          128.43,
    "non_electricity_sources":  [{ "code": "natural_gas_combustion", "quantity": 6275, "unit": "Sm3" }],
    "renewable_share_pct":      44.5,
    "non_renewable_share_pct":  55.5,
    "notes":                    [...]
  },
  "e1_6_ghg_emissions": {
    "scope_1": { "kg_co2eq_total": 12343, "per_source": [...] },
    "scope_2": { "kg_co2eq_location_based": 29904.2, "kg_co2eq_market_based": 42638.8, "kwh_total": 128430 },
    "scope_3": { "kg_co2eq_total": 0, "note": "..." },
    "totals":  { "kg_co2eq_location_based": 42247.2, "kg_co2eq_market_based": 54981.8 }
  },
  "narrative_data_points": {
    "e1_1": null,
    "e1_2": null,
    "e1_3": null,
    "e1_4": null,
    "e1_7": null,
    "e1_8": null,
    "e1_9": null,
    "note": "Narrative E1-1/2/3/4/7/8/9 are bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4 + Sprint S15)."
  },
  "unclassified_rows": 0
}
```

## 4. Tradeoff Stanza

- **Solves:** the absence of a regulator-grade Italian-flagship CSRD ESRS E1 quantitative-disclosure Builder. The two data-point families this Pack covers (E1-5 + E1-6) are the audit-bearing half of E1; the narrative half is client content that flows around the Pack.
- **Optimises for:** primary-source factor citation (ISPRA + GSE Packs as factor providers — every emission row traces to a regulator publication per Rule 132); audit-grade reproducibility (Rule 89 byte-identical replay); structural readiness for EFRAG XBRL tagging in Phase H Sprint S15 (the Body shape mirrors the EFRAG taxonomy data-point names).
- **Sacrifices:** E1-5 lower-heating-value conversion to MWh for non-electricity sources — Phase E reports primary quantity only with a Notes line explaining the deferral; ISPRA LHV factors are added to the ISPRA Pack in Sprint S6 PR #27.5 (TODO). Narrative content (E1-1/2/3/4/7/8/9) is null in this Pack; the engagement-side reporting orchestrator must inject it before signing (Rule 144).
- **Residual risks:** EFRAG taxonomy changes mid-cycle (Rule 138 annual review is the closure); narrative content quality is the engagement client's responsibility (Rule 137 EGE counter-signature mitigates); LHV factor coverage for non-Italian fuels is open until the per-Region Factor Pack catalogue grows.

## 5. EGE counter-signature (Rule 137)

ESRS E1 dossiers above the typical wave-2 / wave-3 client size require an EGE (Esperto in Gestione dell'Energia, UNI CEI 11339) counter-signature on the energy-baseline calculation that supports E1-5 + E1-6. The engagement Phase 0 Discovery verifies the EGE relationship; the Pack does NOT block at generation time, but the engagement-fork's signed-state transition (Rule 144) requires the EGE signature before Cosign-signing the dossier.

## 6. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/report/esrs_e1/`. Pack manifest + CHARTER stay at the repo-root `packs/report/esrs_e1/` per Charter §3.2 discovery convention. See ADR-0024.

## 7. Cross-references

- Pack contract: `backend/internal/domain/reporting/builder.go`.
- Pack manifest: `packs/report/esrs_e1/manifest.yaml`.
- Implementation: `backend/packs/report/esrs_e1/builder.go`.
- Tests: `backend/packs/report/esrs_e1/builder_test.go`.
- Factor dependencies: `packs/factor/{ispra,gse}/`.
- Sister Italian Report Packs: `packs/report/{piano_5_0,conto_termico,tee,audit_dlgs102,monthly_consumption,co2_footprint}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 89, 91, 95, 97, 131, 137, 141, 144.
- Regulator: EFRAG ESRS E1 — Reg. Delegato (UE) 2023/2772 §1; D.Lgs. 6/9/2024 n. 125.
