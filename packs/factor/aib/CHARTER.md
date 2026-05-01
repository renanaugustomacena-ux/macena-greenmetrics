# AIB — European Residual Mix Factor Pack — Charter

> **Pack:** `factor-aib` · **Kind:** Factor · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA (European GO-traced market-based Scope 2 baseline)
> **Doctrine refs:** Rules 90 (temporal-validity factor lookup), 95 (provenance), 97 (algorithm versioning), 125 (signed + cached external sources), 130 (factor sources are first-class), 132 (primary-source citation), 138 (annual review).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Factor Packs).
> **Regulatory base (primary sources):**
>
> - **AIB** (Association of Issuing Bodies) — l'organizzazione che riunisce gli enti emissori di GO (Garanzie di Origine) europei. Pubblica annualmente il **European Residual Mix** per i 30+ paesi partecipanti. <https://www.aib-net.org/facts/european-residual-mix>
> - **AIB Residual Mix Calculation Methodology** — documento metodologico che definisce il calcolo del residual mix sottraendo le quantità di GO emesse ed esportate dal mix di generazione totale del paese. <https://www.aib-net.org/sites/default/files/assets/facts/residual-mix/2024/AIB_2024_Residual_Mix_Results.pdf>
> - **GHG Protocol Scope 2 Guidance** — definisce il dual-method reporting (location-based + market-based); il residual mix è il fattore default per chi non ha contratti rinnovabili specifici.
> - **ISO 14064-1:2018** — definisce il GHG accounting at organization level, includendo il dual-method per Scope 2.
> - **GSE** (Italy) — ente nazionale emissore di GO; partecipa all'AIB. <https://www.gse.it/servizi-per-te/garanzie-di-origine>
>
> Sources accessed 2026-04-30 for Pack v1.0.0.

## 1. Purpose

The AIB European Residual Mix is the **regulator-recognised baseline** for Scope 2 market-based reporting in Europe under the GHG Protocol Scope 2 Guidance + ISO 14064-1. Tenants that do NOT hold Guarantees of Origin (GOs) for their consumption — that is, electricity from the grid without a renewable-source contract — must use the residual mix factor for market-based emissions calculation.

The Pack covers:

- **Italy residual mix** (`it_aib_residual_mix`) — primary use case for the engagement template; supersedes the value previously bundled in `factor-source-gse` with its own dedicated provenance.
- **Key European trading-partner residual mix** — Germany (`de_aib_residual_mix`), France (`fr_aib_residual_mix`), Spain (`es_aib_residual_mix`), Austria (`at_aib_residual_mix`), Switzerland (`ch_aib_residual_mix`). Tenants with multi-country supply contracts use these for location-by-supply Scope 2 attribution.
- **Reporting years** 2022, 2023, 2024, 2025. Each year is a separate Factor row with one-year temporal validity per Rule 90.
- **Renewable-share complement** (`<country>_aib_renewable_share`) — companion percentage that complements the residual-mix factor for tenants that report renewable share alongside emissions.

The AIB residual mix is published with a delay (typically annual report in May/June for the previous calendar year). The Pack therefore pre-computes the static table from publications already issued; the live-fetch path (Phase G) reads the AIB Carbon Footprint Calculator API directly.

## 2. Algorithm

Implements `emissions.FactorSource` with:

- `Name() = "aib"`.
- `Refresh(ctx)` — Phase E returns the checked-in static factor table; Phase G replaces the body with the live-fetch implementation against the AIB Carbon Footprint Calculator API.

Each `Factor` row carries:

- `Code` — `<country>_aib_residual_mix` (g CO2eq/kWh) or `<country>_aib_renewable_share` (%).
- `ValidFromUTC` / `ValidToUTC` — one-year window per Rule 90 (e.g. `[2024-01-01, 2025-01-01)` for the 2024 reporting year).
- `Value`, `Unit`.
- `Source` — citation to the AIB Annual Residual Mix Results report.
- `SourceURL` — link to the AIB publication.
- `Notes` — disambiguation (which countries' GO trade is reflected).

## 3. Refresh signature (Phase G live fetch)

```text
Phase G (Pack v2.0.0):

Refresh(ctx):
  resp, err := tracedHTTP.GET(
      "https://api.aib-net.org/residual-mix/results?year=" + currentYear,
      circuitBreaker, signedCache, retry(...))
  if err != nil: return cachedTable, log warning (Rule 125 — fail open with stale)
  parse JSON → []Factor
  validate ranges (Rule 132 — sanity bounds 100–800 g CO2eq/kWh per country)
  store + sign cache (Rule 169 — every signed+stored payload)
  return parsed
```

The Pack guarantees: (1) `Refresh` is non-blocking on circuit-breaker open; (2) returned factors are deterministic-ordered; (3) failures emit warnings but never `panic` per Rule 76; (4) AIB publishes annually (typically May/June for previous year), so the live cache TTL is set conservatively at 30 days.

## 4. Tradeoff Stanza

- **Solves:** the absence of dedicated AIB residual-mix provenance in the engagement template; the prior value embedded in `factor-source-gse` was correct for Italian-only deployments but lacked the schema discipline (versioning, multi-country support, renewable-share complement) needed for engagements that operate across European borders.
- **Optimises for:** primary-source citation (Rule 132 — every Factor row carries a `Source` annotation linking the AIB Annual Residual Mix Results report); temporal-validity discipline (Rule 90 — one-year windows; non-overlapping); multi-country support (Italy + 5 trading-partner countries in v1.0.0; expandable per Rule 138 annual review); Phase G readiness (the AIB API is REST-stable and the Pack ships the full request shape under TODO).
- **Sacrifices:** the **AIB CFC online calculator's per-source breakdown** (e.g. percentage from coal vs gas vs nuclear) is NOT shipped in v1.0.0 — only the aggregate residual mix factor and the renewable share. Per-source breakdown lands in Pack v1.x.0 minor update if engagement demand surfaces. **Pre-2022 reporting years** are NOT included — those are historical-only and engagement-side hard-coding handles them. **All-Europe (28+ countries)** is NOT shipped in v1.0.0; only Italy + 5 trading-partner countries — the engagement Phase 0 Discovery declares supply geographies and triggers Pack v1.x.0 expansion.
- **Residual risks:** AIB publication delay (typically May/June for previous year); the Pack's "current year" residual mix is the previous year's published value, not the current year's preliminary. This is the regulator-aligned practice (ISO 14064-1 + GHG Protocol Scope 2 Guidance recommend using last-published residual mix). **Cross-country supply contract attribution** (the engagement may have a Spanish renewable contract physically delivered in Italy) is engagement-side bookkeeping, not Pack-side; the Pack provides the country-level baselines and the orchestrator weights by supply-contract geography.

## 5. Refresh cadence

- Phase E: factor table is static (checked-in); Pack version bump triggers a refresh.
- Phase G: live fetch with 30-day cache TTL (AIB publishes annually); circuit breaker per ADR-0009 (60s back-off; degraded-but-serving with stale cache).
- **Annual review** (Rule 138): every July (after the AIB May/June annual report) the engagement team reviews the previous year's residual mix and bumps Pack v1.x.0.

## 6. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/factor/aib/`. Pack manifest + CHARTER stay at the repo-root `packs/factor/aib/` per Charter §3.2 discovery convention. See ADR-0024.

## 7. Cross-references

- Pack contract: `backend/internal/domain/emissions/factor_source.go`.
- Pack manifest: `packs/factor/aib/manifest.yaml`.
- Implementation: `backend/packs/factor/aib/factor.go`.
- Tests: `backend/packs/factor/aib/factor_test.go`.
- Sister Italian Factor Packs: `packs/factor/{ispra,gse,terna}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 90, 95, 97, 125, 130, 132, 138.
- Regulator primary sources: see CHARTER §1 (above).
