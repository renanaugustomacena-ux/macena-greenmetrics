# Monthly Consumption Report Pack — Charter

> **Pack:** `report-monthly_consumption` · **Kind:** Report · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA (reference implementation for Italian Report Packs)
> **Doctrine refs:** Rules 89 (bit-perfect reproducibility), 91 (pure functions), 95 (provenance bundle), 97 (algorithm versioning), 141 (deterministic serialisation).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Report Packs).

## 1. Purpose

The Monthly Consumption Report Pack produces a **periodic operator-facing summary** of energy consumption for an Italian-flagship engagement deployment. Its scope is intentionally narrow:

- aggregate energy consumption (Wh) per meter, per channel, per cost centre, over a half-open period
- compute Scope 2 location-based emissions overlay using the factor row valid at the period midpoint (per Rule 90)
- emit a deterministic byte-stream that makes the report bit-perfectly reproducible

This is the **reference implementation** for the seven Italian Report Packs. Its tests anchor the Builder-purity contract (`tests/conformance/builder_purity_test.go`); its golden fixture demonstrates the Rule 89 reproducibility property; its `provenance.json` shape is the template for the other Italian Report Packs.

It is intentionally NOT regulator-facing — the dossier types (ESRS E1, Piano 5.0, Conto Termico, TEE, audit 102/2014) are produced by their dedicated Report Packs and have their own formal-spec validation paths.

## 2. Algorithm

Pure function `Build(ctx, period, factors, readings) → Report`.

```
1.  Iterate `readings` (already pre-aggregated by Core's reporting orchestrator
    via continuous aggregates; sorted by (meter_id, channel_id, ts)).
2.  Group rows by meter_id + channel_id; sum the `Sum` field across the
    period. Track total reading count per group.
3.  Convert kWh = sum / 1000.0 (assuming Wh ingest unit; honours the row
    `Unit` field).
4.  Look up the Scope 2 factor `it_grid_mix_location` at the period
    midpoint via `factors.Get(...)`. If the factor is missing, the report
    is emitted with `scope_2_kg_co2eq = nil` per group + a `Notes` line
    explaining the gap.
5.  Compute kg CO₂eq = kWh × factor_g_per_kWh / 1000.
6.  Render the canonical typed body (consumption + emissions per group,
    plus a totals row).
7.  Serialise to deterministic JSON: keys sorted alphabetically; `Encoded`
    is `[]byte` of `bytes.NewBuffer(json.MarshalIndent(body, "", "  ") + "\n")`.
8.  Populate Provenance: ManifestLockHash (passed in via context),
    FactorPackVersions{"ispra": <version>}, ReportPackVersion = `Version()`,
    SourceDataWindow = period, executor metadata.
```

The function is **pure**: no `time.Now()`, no `os.Getenv`, no `internal/services/` imports. The conformance test asserts byte-identical output across two consecutive `Build` calls with the same arguments (Rule 89).

## 3. Provenance bundle (Rule 95)

```json
{
  "manifest_lock_hash":   "sha256:...",
  "factor_pack_versions": { "ispra": "1.0.0" },
  "report_pack_version":  "1.0.0",
  "query_definitions":    ["readings_15min view ⨯ period"],
  "source_data_window":   { "period_start_inclusive": "...", "period_end_exclusive": "..." },
  "tenant_data_region":   "it",
  "executor_user_id":     "<uuid>",
  "executed_at_utc":      "<rfc3339>"
}
```

## 4. Tradeoff Stanza

- **Solves:** the simplest end-to-end Report Pack reference; validates the Builder contract (Rule 91), the temporal factor lookup (Rule 90), the deterministic serialisation (Rule 141), and the provenance bundle (Rule 95) on a low-stakes operator-facing dossier.
- **Optimises for:** clarity (the algorithm is six lines of pseudocode), test depth (pure-function determinism + missing-factor edge case), reusability (other Italian Report Packs follow the same shape).
- **Sacrifices:** regulator-facing rigour — this Pack does NOT validate against an external schema (no XBRL, no XSD); that's the job of `report-esrs_e1`, `report-piano_5_0`, `report-conto_termico`, `report-tee`, `report-audit_dlgs102` (each its own Pack with its own formal-spec validation per Rule 131).
- **Residual risks:** the report's interpretation of "month" is the half-open period (Rule 142); if a tenant's month-boundary convention is non-standard (e.g. fiscal month 21st-to-20th), the per-engagement override is via Phase 3 customisation (Rule 152). The Pack does NOT assume calendar months internally.

## 5. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/report/monthly_consumption/`. Pack manifest + CHARTER stay at the repo-root `packs/report/monthly_consumption/` per Charter §3.2 discovery convention. See ADR-0024 for the single-module rationale.

## 6. Cross-references

- Pack contract: `backend/internal/domain/reporting/builder.go`.
- Pack manifest: `packs/report/monthly_consumption/manifest.yaml`.
- Implementation: `backend/packs/report/monthly_consumption/builder.go`.
- Tests: `backend/packs/report/monthly_consumption/builder_test.go`.
- Factor dependency: `packs/factor/ispra/`.
- Sister Italian Report Packs: `packs/report/{esrs_e1,piano_5_0,conto_termico,tee,audit_dlgs102,co2_footprint}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 89, 91, 95, 97, 141.
