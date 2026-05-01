# Terna — Italian Grid Factor Pack — Charter

> **Pack:** `factor-terna` · **Kind:** Factor · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA (Italian transmission system operator data)
> **Doctrine refs:** Rules 90 (temporal-validity factor lookup), 95 (provenance), 97 (algorithm versioning), 125 (signed + cached external sources), 130 (factor sources are first-class), 132 (primary-source citation), 138 (annual review).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Factor Packs).
> **Regulatory base (primary sources):**
>
> - **Terna S.p.A.** — Italian transmission system operator (TSO). Pubblica i dati di produzione, scambio e mix di energia elettrica per il sistema elettrico nazionale.
> - **Terna Trasparenza** — portale dati pubblici (Transparency Platform). <https://www.terna.it/it/sistema-elettrico/transparency-report>
> - **Terna API** — pubblicano dati orari del mix di generazione + scambi transfrontalieri. Phase G integration target.
> - **ARERA** — autorità che stabilisce le regole di reportistica del settore elettrico italiano; Terna è soggetto regolato.
> - **ISO 14064-1** + **GHG Protocol Scope 2 Guidance** — i fattori location-based mensili Terna sono usati per il reporting Scope 2 location-based con granularità mensile (best practice market-leading per audit-grade ESRS E1).
>
> Sources accessed 2026-04-30 for Pack v1.0.0.

## 1. Purpose

The Pack ships Italian electricity-grid emission factors at **monthly granularity** for Scope 2 location-based reporting, complementing the annual values from ISPRA (ENEA Rapporto 404). Monthly granularity is increasingly demanded by audit-grade ESRS E1 dossiers and by tenants that operate energy-intensive shifts: a manufacturing site running 24×7 has a different annualised exposure than a daytime-only office building, and the market-leading reporting practice is to weight consumption by month-of-emission.

Phase E covers:

- **Monthly Italian grid mix factors** (`it_grid_mix_terna_monthly`) — 12 entries per year for 2024-2026, derived from Terna public reports. These supersede ISPRA's annual rolling average for tenants opting into monthly granularity.
- **Hourly renewable-share patterns** (`it_renewable_share_terna_hourly_default`) — 24-entry typical-day curve used as the default if hourly readings flow without per-hour grid factors. The curve reflects Italian solar / wind / hydro production cycles.
- **Scope 2 market-based supplemental** (`it_aib_residual_mix_terna`) — informational override for tenants reporting via Terna's GO (Garanzie di Origine) registry.

Phase G replaces the static factor table with a live fetch against the Terna Trasparenza API (signed + cached per Rule 125; circuit breaker per ADR-0009 — rate-limit 60 req/min, exponential back-off 1s → 60s, signed cache 24h TTL with stale-while-revalidate).

## 2. Algorithm

The Pack implements `emissions.FactorSource` with:

- `Name() = "terna"`.
- `Refresh(ctx)` — Phase E returns the checked-in static factor table; Phase G replaces the body with the live-fetch implementation.

Each `Factor` row carries:

- `Code` — factor identifier consumed by Builder Packs (`it_grid_mix_terna_monthly`, `it_renewable_share_terna_hourly_default`, `it_aib_residual_mix_terna`).
- `ValidFromUTC` / `ValidToUTC` — temporal validity window per Rule 90. Monthly factors carry one-month windows; hourly default profile carries a year-long window for the typical-day pattern.
- `Value`, `Unit` — the numeric payload (g CO2eq/kWh for grid mix; % for renewable share).
- `Source`, `SourceURL` — primary-source citation per Rule 132.
- `Notes` — disambiguation for tenants on what the row applies to (e.g., "Monthly average; not for hourly-resolved reporting").

The static table values for 2024-2026 are derived from Terna public reports for 2024 + 2025 historical and from Terna's NECP-aligned projection for 2026.

## 3. Refresh signature (Phase G live fetch)

```text
Phase G (TODO — Pack v2.0.0):

Refresh(ctx):
  resp, err := tracedHTTP.GET(
      "https://www.terna.it/api/transparency/electricity-generation-mix?period=monthly",
      circuitBreaker, signedCache, retry(...))
  if err != nil: return cachedTable, log warning (Rule 125 — fail open with stale)
  parse XML → []Factor
  validate ranges (Rule 132 — sanity bounds 50–500 g CO2eq/kWh)
  store + sign cache (Rule 169 — every signed+stored payload)
  return parsed
```

The Pack guarantees: (1) `Refresh` is non-blocking on circuit-breaker open — falls back to last-known signed cache; (2) returned factors are deterministic-ordered for `Versions()` reproducibility per Rule 89; (3) failures emit warnings but never `panic` per Rule 76.

## 4. Tradeoff Stanza

- **Solves:** the absence of monthly-granular Italian grid factors in the engagement template. Audit-grade ESRS E1 + Piano 5.0 dossiers benefit from monthly Scope 2 weighting; ISPRA's annual-only series is adequate for compliance-floor reporting but not for tenants pursuing best-in-class disclosure.
- **Optimises for:** primary-source citation (Rule 132 — every Factor row carries a `Source` annotation referencing Terna's published reports); temporal-validity discipline (Rule 90 — one-month windows for monthly factors; non-overlapping intervals enforced by tests); Phase G readiness (the static-table → live-fetch migration replaces the body of `Refresh` without changing the factor codes consumed by Builder Packs).
- **Sacrifices:** **hourly real-time factors** are NOT shipped in Phase E — the `_hourly_default` factor is a single 24-point curve for fallback when per-hour readings flow without per-hour factors. Real hourly resolution lands in Phase G. **Cross-zone (Italy zonal market — North / Centre-N / Centre-S / South / Sicily / Sardinia) factors** are NOT shipped — Phase G adds zonal-market awareness if tenants demand it. **2027+ projections** are NOT included; the Pack ships through 2026 and Pack v1.x.0 minor updates expand annually per Rule 138.
- **Residual risks:** Terna API schema evolution mid-cycle (Rule 138 annual review is the closure; circuit-breaker fallback to ISPRA annual values is the engagement default per `engagements/<id>/config/factor-fallback.yaml`); stale signed cache during prolonged Terna API outages (Rule 125 mitigation — cache TTL is 24h with stale-while-revalidate, beyond which the Pack reverts to ISPRA annual fallback and surfaces a degraded health status).

## 5. Refresh cadence

- Phase E: factor table is static (checked-in); Pack version bump triggers a refresh.
- Phase G: live fetch with 24h cache TTL; circuit-breaker per ADR-0009 (60s back-off on transient failure; degraded-but-serving with stale cache during outages).
- **Annual review** (Rule 138): every January Q1 the engagement team reviews the previous year's monthly factors against the published Terna annual report and bumps Pack v1.x.0.

## 6. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/factor/terna/`. Pack manifest + CHARTER stay at the repo-root `packs/factor/terna/` per Charter §3.2 discovery convention. See ADR-0024.

## 7. Cross-references

- Pack contract: `backend/internal/domain/emissions/factor_source.go`.
- Pack manifest: `packs/factor/terna/manifest.yaml`.
- Implementation: `backend/packs/factor/terna/factor.go`.
- Tests: `backend/packs/factor/terna/factor_test.go`.
- Sister Italian Factor Packs: `packs/factor/{ispra,gse,aib}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 90, 95, 97, 125, 130, 132, 138.
- Regulator primary sources: see CHARTER §1 (above).
