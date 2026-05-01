# Piano Transizione 5.0 — Attestation Report Pack — Charter

> **Pack:** `report-piano_5_0` · **Kind:** Report · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA (regulatory centrepiece for industrial-transformation engagements)
> **Doctrine refs:** Rules 89 (bit-perfect reproducibility), 91 (pure functions), 95 (provenance bundle), 97 (algorithm versioning), 131 (formal-spec validation), 132 (primary-source citation), 137 (regulatory counter-signature), 141 (deterministic serialisation), 144 (signed at finalisation).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Report Packs).
> **Regulatory base (primary sources):**
>
> - **DL 19/2024** art. 38 — Decreto-legge 2 marzo 2024 n. 19 ("PNRR-quater"), convertito con L. 29 aprile 2024 n. 56. <https://www.gazzettaufficiale.it/eli/id/2024/03/02/24G00028/sg>
> - **DM MIMIT-MEF 24/07/2024** — Decreto interministeriale "Attuazione dell'articolo 38 del decreto-legge 2 marzo 2024 n. 19, convertito con modificazioni dalla legge 29 aprile 2024 n. 56, recante le modalità attuative del Piano Transizione 5.0". <https://www.mimit.gov.it/it/normativa/decreti-interministeriali/decreto-interministeriale-24-luglio-2024-modalita-attuative-piano-transizione-5-0>
> - **L. 207/2024** — Legge di bilancio 2025, art. 1 commi 427-429. Rate-table simplification (collapse of the €0–€2.5M and €2.5M–€10M brackets into a single ≤€10M bracket); applies retroactively to investments from 1 January 2024.
> - **MIMIT FAQ Transizione 5.0** (revision 10/04/2025). <https://www.mimit.gov.it/images/stories/documenti/allegati/FAQ_Transizione_50_-_10_aprile_2025.pdf>
>
> Sources accessed 2026-04-30 for Pack v1.0.0.

## 1. Purpose

This Pack produces the **quantitative attestation** for the Piano Transizione 5.0 incentive scheme — a €6.3 billion industrial digital-and-green tax-credit that runs alongside Piano Transizione 4.0 and is administered by GSE. Beneficiary universe: Italian-resident enterprises (excluding agricultural with cadastral-income regime, finance, insurance) that invest 2024-01-01 → 2025-12-31 in projects combining Annex A (4.0/5.0 enabling tech) plus an explicit energy-saving deliverable.

The Pack covers the **two audit-bearing computations**:

- **E** — *energy savings vs counterfactual scenario*, computed twice (struttura produttiva + processi interessati), with the **higher tier** taken as the eligible attestation tier (DM 24/07/2024 art. 9 c. 2-3 dual-criterion).
- **T** — *tax credit per investment bracket × tier rate matrix*, summed across brackets and capped at €50M/year per beneficiary.

Narrative content (ex-ante / ex-post EGE certifications, Allegato A categorisation, investment description, Codici ATECO of beneficiary) flows from client into the Pack output via the engagement-fork's reporting orchestrator (Plan §5.4). Per Rule 137, the EGE (Esperto in Gestione dell'Energia, UNI CEI 11339) or auditor ESCo (UNI CEI 11352) counter-signature on the energy baseline is verified at engagement Phase 0 Discovery and gates the signed-state transition (Rule 144).

## 2. Algorithm

Pure function `Build(ctx, period, factors, readings) → Report`.

```text
1. Read scenario inputs from FactorBundle (per Pack contract; the FactorBundle
   is the keyed-numeric channel for both regulator-published factors and
   engagement-supplied scenario data — same convention as ESRS E1's use of
   `it_renewable_share`):
     - piano5_baseline_energy_kwh        (struttura produttiva, pre-investment)
     - piano5_counterfactual_energy_kwh  (struttura produttiva, post-investment)
     - piano5_baseline_process_kwh       (processi interessati, pre-investment)
     - piano5_counterfactual_process_kwh (processi interessati, post-investment)
     - piano5_investment_total_eur       (eligible investment value)
     - piano5_regime_version             (0 = LB-2025 default, 1 = DM-2024-07-24)
     - piano5_period_year_eur_cap        (optional override of 50M annual cap)

2. Compute saving percentages:
     saving_struttura_pct = (baseline - counterfactual) / baseline * 100
     saving_processi_pct  = (baseline_proc - counterfactual_proc) / baseline_proc * 100

3. Classify into tiers (DM 24/07/2024 art. 9):
     Tier 1 (T1) — saving_struttura ≥ 3% OR saving_processi ≥ 5%
     Tier 2 (T2) — saving_struttura > 6% OR saving_processi > 10%
     Tier 3 (T3) — saving_struttura > 10% OR saving_processi > 15%

   Effective tier = max(tier_struttura, tier_processi). If both 0 → "ineligible".

4. Apply rate matrix per regime_version:

   ──── LB-2025 (default; investments from 1 Jan 2024 onwards) ────
   Bracket 1: investment ≤ €10M
       T1 = 35%   T2 = 40%   T3 = 45%
   Bracket 2: €10M < investment ≤ €50M
       T1 =  5%   T2 = 10%   T3 = 15%

   ──── DM-2024-07-24 (original three-bracket regime) ────
   Bracket 1: investment ≤ €2.5M
       T1 = 35%   T2 = 40%   T3 = 45%
   Bracket 2: €2.5M < investment ≤ €10M
       T1 = 15%   T2 = 20%   T3 = 25%
   Bracket 3: €10M < investment ≤ €50M
       T1 =  5%   T2 = 10%   T3 = 15%

5. Tax credit:
     credit = Σ (investment_in_bracket_i × rate_i_at_tier)
     The €50M annual cap is applied as a hard cut on (credit, investment),
     surfacing the excess in `investment.above_cap_excess_eur`.

6. Render the canonical typed Body (energy_savings + tax_credit + factors_used
   + investment + narrative + notes).

7. Serialise deterministically (Rule 141): JSON with sorted struct fields,
   2-space indent, trailing newline.

8. Provenance is populated by Core's reporting orchestrator at signed-state
   transition (Rule 144). The Pack does NOT call out to a regulator portal;
   GSE submission via the dedicated portal is a separate engagement-side
   Pack ("submitter-gse-piano50") landing in Phase H Sprint S16.
```

The function is **pure** (Rule 91). The conformance test asserts byte-identical Encoded across two consecutive `Build` calls (Rule 89).

## 3. Output shape

```json
{
  "report":         "piano_5_0",
  "regulator":      "MIMIT / GSE (DL 19/2024 art. 38; DM 24/07/2024; L. 207/2024 art. 1 c. 427-429)",
  "regime_version": "lb-2025",
  "period":         { ... },
  "factors_used": {
    "piano5_baseline_energy_kwh":         { "value": 1200000, "unit": "kWh", "version": "engagement-supplied" },
    "piano5_counterfactual_energy_kwh":   { "value": 1080000, "unit": "kWh", "version": "engagement-supplied" },
    "piano5_baseline_process_kwh":        { "value":  400000, "unit": "kWh", "version": "engagement-supplied" },
    "piano5_counterfactual_process_kwh":  { "value":  340000, "unit": "kWh", "version": "engagement-supplied" },
    "piano5_investment_total_eur":        { "value": 4000000, "unit": "EUR", "version": "engagement-supplied" }
  },
  "investment": {
    "total_eur":             4000000,
    "annual_cap_eur":        50000000,
    "above_cap_excess_eur":  0
  },
  "energy_savings": {
    "struttura_produttiva": {
      "baseline_kwh":       1200000,
      "counterfactual_kwh": 1080000,
      "saving_kwh":          120000,
      "saving_pct":              10.0,
      "tier":                       2
    },
    "processi_interessati": {
      "baseline_kwh":        400000,
      "counterfactual_kwh":  340000,
      "saving_kwh":           60000,
      "saving_pct":              15.0,
      "tier":                       2
    },
    "effective_tier":      2,
    "effective_basis":     "struttura_produttiva"
  },
  "tax_credit": {
    "regime_version":      "lb-2025",
    "applied_tier":        2,
    "rate_table": [
      { "bracket_eur_min": 0,         "bracket_eur_max": 10000000, "rate_at_applied_tier": 0.40 },
      { "bracket_eur_min": 10000000,  "bracket_eur_max": 50000000, "rate_at_applied_tier": 0.10 }
    ],
    "per_bracket_eur": [
      { "bracket_eur_min": 0,         "bracket_eur_max": 10000000, "investment_in_bracket_eur": 4000000, "credit_eur": 1600000 },
      { "bracket_eur_min": 10000000,  "bracket_eur_max": 50000000, "investment_in_bracket_eur":       0, "credit_eur":       0 }
    ],
    "total_credit_eur":    1600000
  },
  "narrative_data_points": {
    "ex_ante_certification_id":  null,
    "ex_post_certification_id":  null,
    "investment_description":    null,
    "annex_a_categories":        null,
    "ateco_codes":               null,
    "note": "Narrative content (ex-ante / ex-post EGE certifications, investment description, Allegato A categorisation, ATECO codes) is bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4)."
  },
  "ege_certification_required": true,
  "notes": [...]
}
```

## 4. Tradeoff Stanza

- **Solves:** the absence of an audit-grade Piano Transizione 5.0 attestation Builder. Italian industrial enterprises engaging with the €6.3B incentive face commercially valuable but error-prone tier-classification + per-bracket math; the dual-criterion test in particular is mishandled by spreadsheet-based attestation drafts. The Pack encodes the regime as a pure deterministic function so two re-runs on the same inputs produce byte-identical output (Rule 89).
- **Optimises for:** primary-source citation (Rule 132 — every rate constant carries a `// Source:` comment block referencing DL 19/2024 art. 38 + DM 24/07/2024 art. 9 + L. 207/2024 art. 1 c. 427-429); regime-version awareness (LB-2025 default, original DM-24-07-24 retained for retroactive amendments / historical attestations); engagement-overridable scenario inputs via FactorBundle (the engagement provides the counterfactual model values; the Pack mechanises the math).
- **Sacrifices:** Allegato A / Annex A categorisation of the investment goods is narrative content supplied by the engagement client — the Pack does not enumerate Allegato A category codes, instead leaving that block null with an injection-point note for the orchestrator. ATECO eligibility check is also out of scope (Phase 0 Discovery responsibility). The GSE portal submission (`Comunicazione preventiva` / `Comunicazione di completamento`) is a separate Phase H Pack.
- **Residual risks:** rate-table change mid-engagement (Rule 138 annual review is the closure — `regime_version` is a stable Pack contract field that allows a follow-on `lb-2026` regime to be added without a breaking change); EGE counter-signature relationship gap at Phase 0 Discovery (Rule 137 mitigation — engagement bootstrap runbook explicitly verifies); MIMIT clarification FAQ revisions on edge cases (Rule 132 — the FAQ revision date is captured in this CHARTER §1; major FAQ-driven semantic changes ship as a Pack v1.x.0 minor with a CHANGELOG entry).

## 5. EGE / ESCo counter-signature (Rule 137)

DL 19/2024 art. 38 c. 11 requires that the energy-saving calculation supporting the attestation be signed by either:

- an **EGE** (Esperto in Gestione dell'Energia, certified per UNI CEI 11339), or
- a **ESCo** (Energy Service Company, certified per UNI CEI 11352).

Per Rule 137, the engagement Phase 0 Discovery verifies the EGE/ESCo relationship is in place; the Pack does NOT block at generation time, but the engagement-fork's signed-state transition (Rule 144) requires the counter-signature before Cosign-signing the dossier.

## 6. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/report/piano_5_0/`. Pack manifest + CHARTER stay at the repo-root `packs/report/piano_5_0/` per Charter §3.2 discovery convention. See ADR-0024.

## 7. Cross-references

- Pack contract: `backend/internal/domain/reporting/builder.go`.
- Pack manifest: `packs/report/piano_5_0/manifest.yaml`.
- Implementation: `backend/packs/report/piano_5_0/builder.go`.
- Tests: `backend/packs/report/piano_5_0/builder_test.go`.
- Factor dependencies: `packs/factor/{ispra,gse}/` (the Pack uses the FactorBundle as a scenario-input channel; the actual ISPRA/GSE factors are not consumed in this Pack — the dependency record is for catalogue-level Phase 0 Discovery so the engagement boot-checks the same chain regardless of Report Pack mix).
- Sister Italian Report Packs: `packs/report/{esrs_e1,conto_termico,tee,audit_dlgs102,monthly_consumption,co2_footprint}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 89, 91, 95, 97, 131, 132, 137, 141, 144.
- Regulator primary sources: see CHARTER §1 (above).
