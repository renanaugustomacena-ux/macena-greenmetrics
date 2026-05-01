# GSE Factor Pack — Charter

> **Pack:** `factor-gse` · **Kind:** Factor · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA
> **Doctrine refs:** Rules 90 (temporal validity), 130 (each factor source is a Pack), 132 (Italian regulatory ground truth annotated to primary sources), 138 (annual review).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Factor Packs).

## 1. Purpose

This Pack provides the **GSE-channelled emission factors** for Italian engagements. GSE (Gestore dei Servizi Energetici) is Italy's national authority for renewable-energy promotion, Conto Termico administration, Certificati Bianchi (TEE) issuance, and the AIB residual-mix publication for the Italian electricity market.

Two factor families:

1. **AIB residual mix (Scope 2 market-based)** — the European residual mix attribute for Italy, published annually by the Association of Issuing Bodies (AIB) with a 6-month publication lag. This is the regulator-recognised baseline for **ESRS E1 Scope 2 market-based** reporting in any year where the engagement client cannot demonstrate full Guarantees of Origin coverage. Companion to ISPRA's location-based factor; both are typically reported.
2. **Italian renewable share** — the percentage of the Italian electricity mix produced from renewable sources, by year. Used in Piano Transizione 5.0 attestazione calculations (DL 19/2024 art. 38) and as a high-level KPI on the operator dashboard.

## 2. Factor catalogue

### 2.1 Scope 2 — AIB Italian residual mix (market-based)

| valid_from | valid_to | code | g CO₂eq / kWh | source |
|---|---|---|---|---|
| 2022-01-01 | 2023-01-01 | `it_aib_residual_mix` | 359 | AIB European Residual Mixes 2022 (published 2023-08), §"IT" |
| 2023-01-01 | 2024-01-01 | `it_aib_residual_mix` | 346 | AIB European Residual Mixes 2023 (published 2024-08), §"IT" |
| 2024-01-01 | 2025-01-01 | `it_aib_residual_mix` | 332 | AIB European Residual Mixes 2024 (published 2025-08), §"IT" |
| 2025-01-01 | 2026-01-01 | `it_aib_residual_mix` | 318 | AIB preliminary 2026-Q1, §"IT" |

### 2.2 Italian renewable share (informational + Piano 5.0 calculations)

| valid_from | valid_to | code | share (%) | source |
|---|---|---|---|---|
| 2022-01-01 | 2023-01-01 | `it_renewable_share` | 36.0 | GSE Rapporto Statistico FER 2023 §1 |
| 2023-01-01 | 2024-01-01 | `it_renewable_share` | 39.5 | GSE Rapporto Statistico FER 2024 §1 |
| 2024-01-01 | 2025-01-01 | `it_renewable_share` | 42.1 | GSE Rapporto Statistico FER 2025 §1 (preliminary) |
| 2025-01-01 | 2026-01-01 | `it_renewable_share` | 44.5 | GSE preliminary 2026-04, §1.2 |

### 2.3 Per-source renewable shares (informational)

| code | share (%) at 2024 | source |
|---|---|---|
| `it_re_share_hydro` | 14.8 | GSE Rapporto Statistico FER 2025 §2.1 |
| `it_re_share_solar` | 11.2 | GSE Rapporto Statistico FER 2025 §2.2 |
| `it_re_share_wind` | 7.4 | GSE Rapporto Statistico FER 2025 §2.3 |
| `it_re_share_geothermal` | 1.9 | GSE Rapporto Statistico FER 2025 §2.4 |
| `it_re_share_biomass` | 6.8 | GSE Rapporto Statistico FER 2025 §2.5 |

(Per-source shares ship with `valid_from = 2024-01-01`; updated annually with the FER report.)

## 3. Refresh cadence

GSE publishes the Rapporto Statistico FER annually in March. AIB publishes the European Residual Mixes in August (6-month lag). The Pack's `Refresh(ctx)` is wired to `0 9 15 3 *` for FER + `0 9 20 8 *` for AIB updates, with manual-rerun escape hatches. Failures surface on `/api/health.dependencies.factor-gse`.

For Phase E development, factors are checked-in static data — `Refresh` returns the bundled set. Phase G adds the live fetch path against `areaclienti.gse.it` (with circuit breaker per ADR-0009).

## 4. Tradeoff Stanza

- **Solves:** the absence of an Italian-channelled Scope 2 market-based factor source bundled with the engagement deliverable. AIB residual mix via GSE is the regulator-recognised baseline; combined with ISPRA's location-based factor it provides both ESRS E1 reporting modes.
- **Optimises for:** primary-source citation discipline (Rule 132), temporal-validity correctness (Rule 90), interoperability with Piano 5.0 calculations (renewable share is the qualifying-investment trigger for the higher credit bands).
- **Sacrifices:** dynamic granularity — the Pack ships annual factors only. Hourly market-based factors (a Phase F sophistication) require a Terna-channelled factor source companion (`packs/factor/terna/`) which ships separately.
- **Residual risks:** AIB or GSE may reissue factor values when their underlying datasets are corrected; the Pack version + `Refresh` cadence are the closure path. Annual review (Rule 138) re-validates against the current FER + AIB publication.

## 5. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/factor/gse/`. Pack manifest + CHARTER stay at the repo-root `packs/factor/gse/` per Charter §3.2 discovery convention. See ADR-0024 for the single-module rationale.

## 6. Cross-references

- Pack contract: `backend/internal/domain/emissions/factor_source.go`.
- Pack manifest: `packs/factor/gse/manifest.yaml`.
- Implementation: `backend/packs/factor/gse/factor.go`.
- Tests: `backend/packs/factor/gse/factor_test.go`.
- Italian Region Pack: `packs/region/it/` (declares `factor-source-gse` as a default).
- Sister Italian-flagship Factor Packs: `packs/factor/{ispra,terna,aib}/`.
- Companion Report Packs (consume these factors): `packs/report/{esrs_e1,piano_5_0}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 90, 130, 132, 138.
