# ISPRA Factor Pack — Charter

> **Pack:** `factor-ispra` · **Kind:** Factor · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA
> **Doctrine refs:** Rules 90 (temporal validity), 130 (each factor source is a Pack), 132 (Italian regulatory ground truth annotated to primary sources), 138 (annual review).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Factor Packs).

## 1. Purpose

This Pack provides the **ISPRA-published emission factors** for the Italian electrical mix and for primary fuels. It is the Scope 2 location-based reference for any Italian-flagship engagement and the Scope 1 combustion reference for stationary + mobile combustion.

The authoritative upstream is the ISPRA Rapporto 404 series — *"Fattori di emissione atmosferica di gas a effetto serra nel settore elettrico nazionale e nei principali paesi europei"* — published annually in April with retroactive coverage of the previous calendar year.

## 2. Factor catalogue

### 2.1 Scope 2 — Italian national mix (location-based)

| valid_from | valid_to | code | g CO₂eq / kWh | source |
|---|---|---|---|---|
| 2022-01-01 | 2023-01-01 | `it_grid_mix_location` | 268 | ISPRA Rapporto 404/2024, Tabella 1 |
| 2023-01-01 | 2024-01-01 | `it_grid_mix_location` | 245 | ISPRA Rapporto 404/2025, Tabella 1 |
| 2024-01-01 | 2025-01-01 | `it_grid_mix_location` | 233 | ISPRA Rapporto 404/2025, Tabella 1 |
| 2025-01-01 | 2026-01-01 | `it_grid_mix_location` | 225 | ISPRA preliminary 2026-04, sezione 2 |
| 2026-01-01 | 2027-01-01 | `it_grid_mix_location` | 218 | ISPRA preliminary 2026-04, sezione 2 |

### 2.2 Scope 1 — Stationary combustion

| code | kg CO₂eq / Sm³ or t | unit | source |
|---|---|---|---|
| `natural_gas_combustion` | 1.967 | kg CO₂eq / Sm³ | D.M. 11/01/2017 Allegato 1 |
| `lpg_combustion` | 2.965 | kg CO₂eq / kg | D.M. 11/01/2017 Allegato 1 |
| `diesel_combustion` | 2.642 | kg CO₂eq / l | D.M. 11/01/2017 Allegato 1 |
| `heavy_fuel_oil_combustion` | 3.155 | kg CO₂eq / kg | D.M. 11/01/2017 Allegato 1 |
| `coal_combustion` | 2.394 | kg CO₂eq / kg | D.M. 11/01/2017 Allegato 1 |

(All 2024 publication; valid 2024-01-01 through 2027-12-31 unless ISPRA republishes.)

### 2.3 Scope 1 — Mobile combustion

| code | kg CO₂eq / l | source |
|---|---|---|
| `petrol_road_vehicle` | 2.318 | D.M. 11/01/2017 Allegato 2 |
| `diesel_road_vehicle` | 2.642 | D.M. 11/01/2017 Allegato 2 |

## 3. Refresh cadence

ISPRA publishes annually in April. The Pack's `Refresh(ctx)` is wired to the Core scheduler at `0 9 15 4 *` (every 15 April at 09:00 UTC) with a manual-rerun escape hatch. The 14-day window (15–30 April) accommodates ISPRA's typical staggered publication of preliminary numbers + final tables. Failures are logged via the audit chain and surfaced on `/api/health.dependencies.factor-ispra`.

For Phase E development, factors are checked-in static data — `Refresh` returns the bundled set without an HTTP call. Phase G hardens the live-fetch path.

## 4. Tradeoff Stanza

- **Solves:** the absence of an authoritative Italian Scope 2 location-based factor source bundled with the engagement deliverable. ISPRA Rapporto 404 is the regulator-recognised baseline.
- **Optimises for:** primary-source citation discipline (Rule 132), temporal-validity correctness (Rule 90), audit-grade traceability (every factor row carries `source` + `source_url` + retrieval timestamp).
- **Sacrifices:** complete factor coverage — only the most-used Scope 1 + Scope 2 factors ship; specialised cases (e.g. district heating, flue-gas cleaning losses) are deferred to per-engagement override Packs.
- **Residual risks:** ISPRA may republish or correct factor values between annual editions; the Pack version + `Refresh` cadence are the closure path. Annual review (Rule 138) re-validates against the current Rapporto 404.

## 5. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/factor/ispra/`. Pack manifest + CHARTER stay at the repo-root `packs/factor/ispra/` per Charter §3.2 discovery convention. See ADR-0024 for the single-module rationale.

## 6. Cross-references

- Pack contract: `backend/internal/domain/emissions/factor_source.go`.
- Pack manifest: `packs/factor/ispra/manifest.yaml`.
- Implementation: `backend/packs/factor/ispra/factor.go`.
- Tests: `backend/packs/factor/ispra/factor_test.go`.
- Italian Region Pack: `packs/region/it/` (declares this factor source as a default).
- Sister Italian-flagship Factor Packs: `packs/factor/{gse,terna,aib}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 90, 130, 132, 138.
