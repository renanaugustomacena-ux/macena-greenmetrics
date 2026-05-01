# Italian Region Pack — Charter

> **Pack:** `region-it` · **Kind:** Region · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA (flagship reference per Rule 88)
> **Doctrine refs:** Rules 8 (Italian residency default), 88 (flagship reference), 101 (per-tenant timezone), 132 (Italian regulatory ground truth), 139 (thresholds propagated, not duplicated), 140 (per-tenant regulatory profile is explicit).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Region Packs).

## 1. Purpose

The Italian Region Pack is the **flagship reference Pack** per Rule 88. It bundles every regional default an Italian-flagship engagement deployment needs: timezone, locale, currency, holiday calendar, default regulatory-regime set, and the regulatory threshold table that the Italian Report Packs (`packs/report/{esrs_e1, piano_5_0, conto_termico, tee, audit_dlgs102}`) consult.

Other Region Packs (DE, FR, ES, GB, AT) are reviewed against this Pack for thoroughness — the structure, the documentation depth, and the threshold-source citation pattern are the template.

## 2. Profile

| Field | Value |
|---|---|
| Code | `it` |
| Timezone | `Europe/Rome` (CET in winter, CEST in summer; DST per EU directive) |
| Locale | `it_IT.UTF-8` |
| Currency | `EUR` (comma as decimal separator) |
| Default regimes | `csrd_wave_2`, `csrd_wave_3`, `piano_5_0`, `conto_termico`, `tee`, `audit_dlgs102`, `ets`, `gdpr`, `nis2_italia`, `arera` |

## 3. Holiday calendar

Italian national holidays per L. 27 maggio 1949, n. 260 plus a Veneto-regional holiday (Patrono di Verona, 21 May per local tradition; per-tenant configurability deferred to Phase H Sprint S15). The full list:

| Date | Holiday |
|---|---|
| 1 January | Capodanno |
| 6 January | Epifania |
| Movable (Sun.) | Pasqua |
| Movable (Mon.) | Lunedì dell'Angelo |
| 25 April | Festa della Liberazione |
| 1 May | Festa del Lavoro |
| 2 June | Festa della Repubblica |
| 15 August | Ferragosto |
| 1 November | Tutti i Santi |
| 8 December | Immacolata Concezione |
| 25 December | Natale |
| 26 December | Santo Stefano |
| 21 May | Patrono di Verona (Veneto regional) |

Easter Sunday is computed via the anonymous Gregorian algorithm (Meeus 1991). Easter Monday is Easter Sunday + 1 day.

## 4. Regulatory threshold table

Cited from primary sources per Rule 132. Versioned via the threshold key + an annotation indicating the source-document-version. Per Rule 139, thresholds are propagated to Report Packs (read via `RegulatoryThreshold(name)`) — not duplicated inside Report Pack code.

| Threshold key | Value | Source |
|---|---|---|
| `csrd_wave_2_employee_threshold` | 250 | Dir. UE 2022/2464 art. 19a |
| `csrd_wave_2_turnover_eur` | 50_000_000 | Dir. UE 2022/2464 art. 19a |
| `csrd_wave_2_balance_sheet_eur` | 25_000_000 | Dir. UE 2022/2464 art. 19a |
| `piano_5_0_process_reduction_pct` | 3.0 | DL 19/2024 art. 38 c.4 |
| `piano_5_0_site_reduction_pct` | 5.0 | DL 19/2024 art. 38 c.4 |
| `piano_5_0_max_qualifying_outlay_eur` | 50_000_000 | DL 19/2024 art. 38 c.6 |
| `audit_dlgs102_kwh_threshold` | 0.0 | D.Lgs. 102/2014 art. 8 (no kWh threshold; obligation is by company size) |
| `audit_dlgs102_employee_threshold` | 250 | D.Lgs. 102/2014 art. 8 c.1 |
| `audit_dlgs102_turnover_eur` | 50_000_000 | D.Lgs. 102/2014 art. 8 c.1 |
| `audit_dlgs102_balance_sheet_eur` | 43_000_000 | D.Lgs. 102/2014 art. 8 c.1 |
| `tee_minimum_certificate_toe` | 1.0 | DM 11 gennaio 2017 art. 6 |
| `conto_termico_max_intervento_eur` | 700_000 | DM 16 febbraio 2016 art. 5 |
| `arera_smart_meter_data_window_min` | 15 | ARERA delibera 646/2015 |
| `nis2_notification_window_h_initial` | 24 | D.Lgs. 138/2024 art. 24 c.1 |
| `nis2_notification_window_h_full` | 72 | D.Lgs. 138/2024 art. 24 c.2 |
| `nis2_final_report_window_d` | 30 | D.Lgs. 138/2024 art. 24 c.3 |

## 5. Authorities

Authority lookup map exposed via `Profile().Authorities`:

| Authority | URL / portal |
|---|---|
| `gse` | https://areaclienti.gse.it |
| `enea` | https://audit102.enea.it |
| `ispra` | https://www.isprambiente.gov.it |
| `terna` | https://transparency.entsoe.eu |
| `arera` | https://www.arera.it |
| `garante` | https://www.garanteprivacy.it |
| `acn` | https://www.acn.gov.it |
| `mimit` | https://www.mimit.gov.it |
| `mase` | https://www.mase.gov.it |

## 6. Privacy regime

GDPR (Reg. UE 2016/679) applies. Italian supplements: D.Lgs. 196/2003 as amended by D.Lgs. 101/2018; Garante guidance; ARERA delibera 646/2015 + Measure 147/2023 on near-real-time consumption data as personal data even when the POD owner is a legal entity. The Region Pack does not encode controller-of-record obligations; those are configured per-engagement in `engagements/<id>/CHARTER.md`.

## 7. Tradeoff Stanza

- **Solves:** the absence of a single authoritative answer for "Italian engagement defaults" — timezone, locale, currency, holiday calendar, applicable regulatory regimes, threshold values traceable to primary sources, authority contact map.
- **Optimises for:** thoroughness (matches Rule 88 flagship requirement), citation discipline (every threshold traces to a primary source), reusability (other Region Packs DE/FR/ES/GB/AT inherit this structure).
- **Sacrifices:** customisation depth out of the box — per-tenant regulatory profile (Rule 140) is the override mechanism, and per-engagement overlays in `engagements/<id>/` are the second.
- **Residual risks:** regulatory threshold values can change mid-cycle (e.g. CSRD-wave thresholds review every two years); Pack version bump + downstream Report Pack regression test is the closure path. Annual Pack review per Rule 138 is the anchor.

## 8. Annual review

The Pack is reviewed every Q1 by the engagement-team energy-advisor + EGE counter-signature. The review verifies that every threshold value still cites the current primary-source version. Output is a row in `REVIEW-LOG.md` (this Pack) + an entry in the portfolio-wide `docs/PACK-REVIEW-LOG.md`.

## 9. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/region/it/` while the project is on a single Go module (`github.com/greenmetrics/backend`). Pack manifest + CHARTER stay at the repo-root `packs/region/it/` for discovery per Charter §3.2. When the contract interfaces (`backend/internal/domain/{region,emissions,reporting,protocol,identity}/`) are extracted into a separate `github.com/greenmetrics/contract` module — tracked as a Phase F decision — Pack Go code moves to repo-root `packs/<kind>/<id>/` and the `backend/packs/` tree is deleted. This avoids the backend ↔ packs module-cycle problem. ADR-0024 records the rationale.

## 10. Cross-references

- Pack contract: `backend/internal/domain/region/profile.go` (interface).
- Pack manifest: `packs/region/it/manifest.yaml`.
- Implementation: `backend/packs/region/it/profile.go`.
- Tests: `backend/packs/region/it/profile_test.go`.
- Sister Italian-flagship Packs: `packs/factor/{ispra,gse,terna,aib}/`, `packs/report/{esrs_e1,piano_5_0,conto_termico,tee,audit_dlgs102,monthly_consumption,co2_footprint}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 8, 88, 101, 132, 138, 139, 140.
