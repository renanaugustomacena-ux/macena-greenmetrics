# TEE / Certificati Bianchi — Report Pack — Charter

> **Pack:** `report-tee` · **Kind:** Report · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA (Italian energy-efficiency certificate scheme; tradeable on the GME borsa)
> **Doctrine refs:** Rules 89 (bit-perfect reproducibility), 91 (pure functions), 95 (provenance bundle), 97 (algorithm versioning), 131 (formal-spec validation), 132 (primary-source citation), 137 (regulatory counter-signature), 141 (deterministic serialisation), 144 (signed at finalisation).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Report Packs).
> **Regulatory base (primary sources):**
>
> - **D.M. 11 gennaio 2017** — Decreto interministeriale "Determinazione degli obiettivi quantitativi nazionali di risparmio energetico [...] di cui all'articolo 7 del decreto legislativo 4 luglio 2014 n. 102". Fonte primaria del meccanismo TEE 2017-2024. <https://www.mimit.gov.it/images/stories/normativa/DM-Certificati-Bianchi_2017.pdf>
> - **GSE Linee Guida Operative Certificati Bianchi** (April 2019; aggiornate). Documenti di supporto operativo. <https://www.gse.it/servizi-per-te/efficienza-energetica/certificati-bianchi/documenti>
> - **D.M. MASE 21 luglio 2025** — Decreto MASE che aggiorna le regole TEE per il periodo 2025-2030, in linea col PNIEC 2024.
>
> Sources accessed 2026-04-30 for Pack v1.0.0.

## 1. Purpose

The Italian Titoli di Efficienza Energetica (TEE) — comunemente "Certificati Bianchi" — sono certificati tradeable rilasciati dal GSE che attestano il conseguimento di **un risparmio annuo di 1 tonnellata equivalente di petrolio (1 TEE = 1 tep)** ottenuto attraverso un progetto di efficienza energetica. Operatori obbligati (distributori energia elettrica e gas con >50 000 utenti) devono raggiungere quote annuali di TEE; possono comprare i titoli sulla borsa GME oppure realizzare/acquisire progetti.

Il Pack copre la **half computazionale** del processo TEE:

- **Calcolo del risparmio addizionale annuo in tep** — differenza ex-ante / ex-post sui consumi caratteristici del progetto, opzionalmente convertita in tep tramite i fattori statutari ENEA / ARERA (riusati dal Pack `audit_dlgs102`).
- **Fattore moltiplicativo K** — DM 11/01/2017 sostituì il vecchio τ con i fattori K1=1.2 (prima metà della vita utile del progetto) e K2=0.8 (seconda metà). Questo Pack applica K determinato dal periodo corrente vs vita utile dichiarata.
- **TEE rilasciati** — annual_saving_tep × K_factor.
- **Tipologia di metodo di calcolo** — `consuntivo` / `standardizzato` / `pppm` selezionabile via FactorBundle.
- **Regime version** — `dm-2017` (default; DM 11/01/2017) vs `dm-mase-2025-2030` (selettore preliminare; nuove regole 2025-2030 in fase di consolidamento operativo — coefficienti specifici da verificare in Phase 0 Discovery contro il decreto vigente alla data di richiesta RVC).

Narrative content (project description, EGE/ESCo certifier ID, RVC submission ID, vita utile methodology, baseline calculation methodology) flows from client into the dossier via the engagement-fork's reporting orchestrator. GSE portal RVC submission (`Richiesta di Verifica e Certificazione`) is a separate engagement-side Pack landing in Phase H Sprint S16.

Per Rule 137, il progetto TEE richiede la firma di un EGE (UNI CEI 11339) o di un auditor di una ESCo (UNI CEI 11352); l'engagement Phase 0 Discovery verifies the relationship is in place.

## 2. Algorithm

Pure function `Build(ctx, period, factors, readings) → Report`.

```text
1. Read scenario inputs from FactorBundle:
     tee_method                     — 0=consuntivo, 1=standardizzato, 2=pppm
     tee_ex_ante_tep                — annual tep consumption pre-intervention
     tee_ex_post_tep                — annual tep consumption post-intervention
     tee_vita_utile_years           — declared project useful life
     tee_current_year_in_project    — current year of the certification cycle (1, 2, ...)
     tee_intervention_category      — engagement-supplied integer category code (Allegato 2 GSE)
     tee_regime_version             — 0 = dm-2017 (default), 1 = dm-mase-2025-2030

2. Compute annual saving:
     annual_saving_tep = ex_ante_tep - ex_post_tep
     saving_pct = (annual_saving_tep / ex_ante_tep) × 100   (if ex_ante_tep > 0)

3. Determine vita utile half:
     half_threshold = vita_utile_years / 2
     if current_year_in_project ≤ half_threshold:
       half = "first";  K_factor = 1.2  (K1)
     else:
       half = "second"; K_factor = 0.8  (K2)

4. TEE issued for the current year:
     tee_issued = annual_saving_tep × K_factor

   If annual_saving_tep ≤ 0 → tee_issued = 0; eligibility note emitted.

5. Render the canonical typed Body (project + energy_savings + tee_calculation
   + factors_used + narrative + notes).

6. Serialise deterministically (Rule 141): JSON with sorted struct fields,
   2-space indent, trailing newline.

7. Provenance is populated by Core's reporting orchestrator at signed-state
   transition (Rule 144).
```

The function is **pure** (Rule 91). The conformance test asserts byte-identical Encoded across two consecutive `Build` calls (Rule 89).

## 3. Output shape

```json
{
  "report":         "tee",
  "regulator":      "GSE / MASE — DM 11/01/2017; DM MASE 21/07/2025 (regime 2025-2030)",
  "regime_version": "dm-2017",
  "period":         { ... },
  "factors_used": {
    "tee_method":                  { "value": 0, "unit": "selector",  "version": "engagement-supplied" },
    "tee_ex_ante_tep":             { "value": 250, "unit": "tep/year","version": "engagement-supplied" },
    "tee_ex_post_tep":             { "value": 180, "unit": "tep/year","version": "engagement-supplied" },
    "tee_vita_utile_years":        { "value":  10, "unit": "year",    "version": "engagement-supplied" },
    "tee_current_year_in_project": { "value":   3, "unit": "year",    "version": "engagement-supplied" }
  },
  "project": {
    "method":                  "consuntivo",
    "intervention_category":   null,
    "vita_utile_years":        10,
    "current_year_in_project": 3,
    "current_period_half":     "first"
  },
  "energy_savings": {
    "ex_ante_tep":         250.0,
    "ex_post_tep":         180.0,
    "annual_saving_tep":    70.0,
    "saving_pct":           28.0
  },
  "tee_calculation": {
    "regime_version":    "dm-2017",
    "annual_saving_tep":  70.0,
    "k_factor":            1.2,
    "tee_issued":         84.0
  },
  "narrative_data_points": {
    "project_description":         null,
    "ege_certifier_id":            null,
    "rvc_submission_id":           null,
    "vita_utile_methodology":      null,
    "baseline_calculation_method": null,
    "note": "Narrative content (project description, EGE certifier ID, RVC submission ID, vita utile methodology, baseline calculation methodology) is bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4)."
  },
  "ege_certification_required": true,
  "notes": [...]
}
```

## 4. Tradeoff Stanza

- **Solves:** the absence of a regulator-grade TEE per-project tep-saving + K-factor calculation Builder. The K-factor application is widely mishandled in spreadsheet-based TEE proposals (the "first / second half of vita utile" boundary is an off-by-one source of audit findings); the Pack encodes the regime as a pure deterministic function so two re-runs on the same inputs produce byte-identical output (Rule 89).
- **Optimises for:** primary-source citation on the K1=1.2 / K2=0.8 constants (Rule 132 — every constant carries a `// Source: DM 11/01/2017 art. 8 c. 4` comment); regime-version awareness (`dm-2017` default + `dm-mase-2025-2030` selector for future extension); engagement-overridable scenario inputs via FactorBundle (project parameters flow as tee_* keys).
- **Sacrifices:** Allegato 2 intervention-category catalogue (industrial / agricoltura / edifici / illuminazione / etc.) is NOT enumerated in the Pack — engagement-side knowledge — surfaced as opaque integer code only. The DM MASE 21/07/2025 detailed coefficient changes are NOT yet encoded; the `dm-mase-2025-2030` regime falls back to `dm-2017` semantics and emits an explicit warning note. Engagement Phase 0 Discovery verifies the regime version current at RVC submission. The pre-DM-2017 τ ("tau") regime is intentionally NOT supported — projects under that regime have already concluded their certification cycles.
- **Residual risks:** DM MASE 21/07/2025 coefficient evolution before Pack v1.x.0 minor update (Rule 138 annual review is the closure); EGE / ESCo relationship gap at Phase 0 Discovery (Rule 137 mitigation); GSE RVC portal schema evolution (Rule 138 — the `submitter-gse-tee` Pack version-binds to portal-version current at submission).

## 5. EGE / ESCo counter-signature (Rule 137)

DM 11/01/2017 art. 6 richiede che la richiesta TEE sia presentata da un soggetto qualificato:

- una **ESCo** (Energy Service Company, certificata UNI CEI 11352), o
- un soggetto con un **EGE** (Esperto in Gestione dell'Energia, certificato UNI CEI 11339), o
- un soggetto con un **sistema di gestione dell'energia conforme alla ISO 50001** + responsabile certificato.

Per Rule 137, l'engagement Phase 0 Discovery verifies the certifier relationship is in place; the Pack does NOT block at generation time, but the engagement-fork's signed-state transition (Rule 144) requires the certifier signature before Cosign-signing the dossier.

## 6. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/report/tee/`. Pack manifest + CHARTER stay at the repo-root `packs/report/tee/` per Charter §3.2 discovery convention. See ADR-0024.

## 7. Cross-references

- Pack contract: `backend/internal/domain/reporting/builder.go`.
- Pack manifest: `packs/report/tee/manifest.yaml`.
- Implementation: `backend/packs/report/tee/builder.go`.
- Tests: `backend/packs/report/tee/builder_test.go`.
- Factor dependencies: `packs/factor/{ispra,gse}/` (catalogue-level — the Pack uses FactorBundle as a scenario channel; ISPRA/GSE specific factors are not consumed in v1.0.0).
- Sister Italian Report Packs: `packs/report/{esrs_e1,piano_5_0,conto_termico,audit_dlgs102,monthly_consumption,co2_footprint}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 89, 91, 95, 97, 131, 132, 137, 141, 144.
- Regulator primary sources: see CHARTER §1 (above).
