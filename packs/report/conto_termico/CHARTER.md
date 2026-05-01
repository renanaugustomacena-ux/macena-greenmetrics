# Conto Termico — Report Pack — Charter

> **Pack:** `report-conto_termico` · **Kind:** Report · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA (Italian thermal-energy-efficiency incentive scheme; €900M/year national budget, 50% PA + 50% privato sotto CT 2.0)
> **Doctrine refs:** Rules 89 (bit-perfect reproducibility), 91 (pure functions), 95 (provenance bundle), 97 (algorithm versioning), 131 (formal-spec validation), 132 (primary-source citation), 137 (regulatory counter-signature), 141 (deterministic serialisation), 144 (signed at finalisation).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Report Packs).
> **Regulatory base (primary sources):**
>
> - **D.M. 16 febbraio 2016** ("Conto Termico 2.0") — Decreto interministeriale "Aggiornamento della disciplina per l'incentivazione di interventi di piccole dimensioni per l'incremento dell'efficienza energetica e per la produzione di energia termica da fonti rinnovabili". Sostituì il D.M. 28/12/2012 originale. <https://www.mimit.gov.it/it/normativa/decreti-interministeriali/conto-termico-aggiornamento-disciplina-incentivazione-interventi-di-piccole-dimensioni>
> - **D.M. 7 agosto 2025** ("Conto Termico 3.0") — Decreto MASE che aggiorna la disciplina con dotazione di €900M/anno, contributo in conto capitale fino al 65% delle spese ammissibili, finestra di richiesta estesa a 90 giorni, beneficiari ampliati (enti del terzo settore, autoconsumo collettivo, comunità energetiche rinnovabili). <https://www.mase.gov.it/portale/conto-termico>
> - **GSE Regole Applicative Conto Termico 2.0** — documento operativo aggiornato. <https://www.gse.it/documenti_site/Documenti%20GSE/Servizi%20per%20te/CONTO%20TERMICO/REGOLE%20APPLICATIVE/REGOLE_APPLICATIVE_CT.pdf>
> - **GSE Regole Applicative Conto Termico 3.0** — pubblicate 19 dicembre 2025. <https://www.gse.it/servizi-per-te/news/conto-termico-3-0-pubblicate-le-regole-applicative>
>
> Sources accessed 2026-04-30 for Pack v1.0.0.

## 1. Purpose

Il Conto Termico è lo strumento principale italiano per incentivare interventi di efficienza energetica e di produzione di energia termica da fonti rinnovabili **di piccole dimensioni**. Beneficiari (CT 3.0): pubbliche amministrazioni, imprese, soggetti privati, terzo settore, autoconsumo collettivo, CER. Interventi tipici: pompe di calore (aria/acqua/geotermiche), impianti a biomassa certificata (≥ 4 o 5 stelle ambientali per CT 3.0), impianti solari termici, microcogenerazione da FER, allaccio a teleriscaldamento efficiente. Per la PA: anche miglioramenti dell'involucro edilizio (isolamento, serramenti) e sostituzione di chiusure trasparenti, illuminazione, sistemi di automazione.

Il Pack copre la **half computazionale** del processo Conto Termico:

- **Importo dell'incentivo per il singolo intervento** — accettato dal FactorBundle come `conto_termico_incentive_amount_eur`. La formula di calcolo per intervento è regolatorialmente densa (per-intervento, per-categoria climatica, per-tipo di tecnologia, con coefficienti di Allegato I-II); v1.0.0 la lascia all'engagement orchestrator (in linea col pattern di Piano 5.0 e TEE che accettano scenari di savings da engagement). Phase F integrerà l'API GSE Regole Applicative per il computo automatico.
- **Schedulazione del pagamento** — per CT 2.0:
  - Erogazione in **un'unica soluzione** se l'importo ≤ €5 000 (DM 16/02/2016 art. 8 c. 8).
  - Erogazione in **rate annuali costanti** di durata 2-5 anni a seconda della tipologia dell'intervento (default 2 anni per impianti FER fino a 35 kWt; 5 anni per interventi maggiori).
- **Schedulazione del pagamento — CT 3.0**: erogazione in conto capitale fino al 65% delle spese ammissibili; il pagamento può essere immediato (single tranche) o suddiviso secondo le Regole Applicative (engagement-side).
- **Massimale di cumulo** — per la PA (CT 2.0): cumulo con incentivi in conto capitale statali / non statali fino al 100% delle spese ammissibili; per i privati: limiti specifici per categoria.
- **Selettore regime** — `ct-2-0` default + `ct-3-0` selettore valore 1; il regime scelto determina la finestra di submission (60 vs 90 giorni post-completamento) e la struttura dei pagamenti.

Narrative content (intervention description, ATECO codes, beneficiary contact, supplier invoices, certifier ID, climate zone, building data) flows from client into the dossier via the engagement-fork's reporting orchestrator. GSE Portaltermico submission is a separate engagement-side Pack landing in Phase H Sprint S16.

Per Rule 137, gli interventi sopra soglie tecniche tipiche (es. installazione di pompe di calore di taglia rilevante) richiedono che il progetto sia validato da un **EGE** (UNI CEI 11339) o da un **auditor di una ESCo** (UNI CEI 11352); l'engagement Phase 0 Discovery verifies the relationship is in place.

## 2. Algorithm

Pure function `Build(ctx, period, factors, readings) → Report`.

```text
1. Read scenario inputs from FactorBundle:
     conto_termico_regime_version           — 0 = ct-2-0 (default), 1 = ct-3-0
     conto_termico_intervention_category    — engagement-supplied integer
                                              (Allegato I-II GSE Regole Applicative)
     conto_termico_beneficiary_type         — 0 = PA, 1 = privato, 2 = ETS, 3 = CER
     conto_termico_incentive_amount_eur     — engagement-computed per-intervention
                                              incentive (Phase F = automated formula)
     conto_termico_eligible_costs_eur       — total eligible costs of the intervention
     conto_termico_climate_zone             — 1=A, 2=B, 3=C, 4=D, 5=E, 6=F
     conto_termico_payment_years_override   — optional; engagement override of
                                              the default annual-rate count

2. Determine regime:
     regime = "ct-2-0"  if regime_version == 0  (default)
              "ct-3-0"  if regime_version == 1

3. Compute payment schedule (CT 2.0):
     if incentive_amount ≤ 5000 EUR:
       payment_mode = "single_tranche"
       annual_rate_eur = incentive_amount
       payment_years = 1
     else:
       payment_mode = "annual_rates"
       payment_years = override OR default-by-category
                       (default: 2 years for FER interventions ≤ 35 kWt;
                                 5 years for larger / building-envelope)
       annual_rate_eur = incentive_amount / payment_years

   Compute payment schedule (CT 3.0):
     payment_mode = "capital_grant"
     payment_years = 1   (default; orchestrator may split per Regole Applicative)
     annual_rate_eur = incentive_amount

4. Eligible-cost cap check (CT 3.0):
     max_cap_pct = 0.65   (DM 7/08/2025: contributo in conto capitale fino al 65%)
     max_cap_eur = eligible_costs_eur × max_cap_pct
     if incentive_amount > max_cap_eur:
       cap_violation = true
       cap_excess_eur = incentive_amount - max_cap_eur
     else:
       cap_violation = false

5. Render the canonical typed Body (intervention + payment_schedule
   + eligible_costs + factors_used + narrative + notes).

6. Serialise deterministically (Rule 141): JSON with sorted struct fields,
   2-space indent, trailing newline.

7. Provenance is populated by Core's reporting orchestrator at signed-state
   transition (Rule 144).
```

The function is **pure** (Rule 91). The conformance test asserts byte-identical Encoded across two consecutive `Build` calls (Rule 89).

## 3. Output shape

```json
{
  "report":         "conto_termico",
  "regulator":      "GSE / MASE — DM 16/02/2016 (CT 2.0); DM 7/08/2025 (CT 3.0)",
  "regime_version": "ct-2-0",
  "period":         { ... },
  "factors_used": {
    "conto_termico_intervention_category": { "value": 12, "unit": "category-code", "version": "engagement-supplied" },
    "conto_termico_beneficiary_type":      { "value":  0, "unit": "selector",      "version": "engagement-supplied" },
    "conto_termico_incentive_amount_eur":  { "value": 4500, "unit": "EUR",         "version": "engagement-supplied" }
  },
  "intervention": {
    "category_code":  12,
    "beneficiary":    "PA",
    "climate_zone":   "E"
  },
  "eligible_costs": {
    "total_eur":      9000,
    "max_cap_pct":    null,
    "max_cap_eur":    null,
    "cap_violation":  false,
    "cap_excess_eur": 0
  },
  "payment_schedule": {
    "regime_version":   "ct-2-0",
    "incentive_amount_eur": 4500,
    "payment_mode":     "single_tranche",
    "payment_years":    1,
    "annual_rate_eur":  4500,
    "submission_window_days": 60
  },
  "narrative_data_points": {
    "intervention_description": null,
    "ateco_codes":              null,
    "beneficiary_contact":      null,
    "supplier_invoices":        null,
    "certifier_id":             null,
    "building_data":            null,
    "note": "Narrative content (intervention description, ATECO codes, beneficiary contact, supplier invoices, certifier ID, building data) is bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4)."
  },
  "ege_certification_required": true,
  "notes": [...]
}
```

## 4. Tradeoff Stanza

- **Solves:** the absence of a regulator-grade Conto Termico payment-schedule + cap-violation Builder. The 1-shot vs annual-rates threshold (€5 000 under CT 2.0) and the CT 3.0 65%-of-eligible-costs cap are off-by-one prone in spreadsheet-based dossiers; the Pack encodes the regime as a pure deterministic function so two re-runs on the same inputs produce byte-identical output (Rule 89).
- **Optimises for:** primary-source citation on the €5 000 threshold + 65% cap (Rule 132); dual-regime support via `regime_version` selector with submission-window awareness (60 days CT 2.0, 90 days CT 3.0); engagement-supplied per-intervention incentive amount (consistent with the Pack pattern of Piano 5.0 + TEE — engagement-side knowledge enters via FactorBundle).
- **Sacrifices:** per-intervention formulas (Allegato I-II GSE Regole Applicative — heat pump COP × climate zone × power × tariff coefficients; biomass kWt × stelle-ambientali coefficient; solar thermal m² × yield) are NOT encoded in v1.0.0 — engagement supplies the pre-computed `conto_termico_incentive_amount_eur`. Phase F integrates the GSE Linee Guida API for automated formula evaluation. The CT 3.0 detailed payment-tranche structure and ETS / CER beneficiary eligibility checks land in Pack v1.x.0 minor update following GSE Regole Applicative consolidation. Catalogue of intervention-category codes (the 30+ entries of Allegato I-II) is engagement-side knowledge, surfaced as opaque integer.
- **Residual risks:** GSE Portaltermico schema evolution (Rule 138 — the `submitter-gse-conto-termico` Pack version-binds at submission); CT 3.0 evolutionary clarifications post Regole Applicative 19/12/2025 (Rule 138 annual review); EGE / ESCo relationship gap at Phase 0 Discovery (Rule 137 mitigation — engagement bootstrap runbook explicitly verifies).

## 5. EGE / ESCo counter-signature (Rule 137)

Le Regole Applicative GSE richiedono che gli interventi sopra soglie di taglia tipica del Conto Termico (es. pompe di calore > 35 kWt) siano validati da un soggetto qualificato:

- un **EGE** (Esperto in Gestione dell'Energia, certificato UNI CEI 11339), o
- un **auditor di una ESCo** (Energy Service Company, certificata UNI CEI 11352).

Per Rule 137, l'engagement Phase 0 Discovery verifies the relationship is in place; the Pack does NOT block at generation time, but the engagement-fork's signed-state transition (Rule 144) requires the certifier signature before Cosign-signing the dossier.

## 6. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/report/conto_termico/`. Pack manifest + CHARTER stay at the repo-root `packs/report/conto_termico/` per Charter §3.2 discovery convention. See ADR-0024.

## 7. Cross-references

- Pack contract: `backend/internal/domain/reporting/builder.go`.
- Pack manifest: `packs/report/conto_termico/manifest.yaml`.
- Implementation: `backend/packs/report/conto_termico/builder.go`.
- Tests: `backend/packs/report/conto_termico/builder_test.go`.
- Factor dependencies: `packs/factor/{ispra,gse}/`.
- Sister Italian Report Packs: `packs/report/{esrs_e1,piano_5_0,tee,audit_dlgs102,monthly_consumption,co2_footprint}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 89, 91, 95, 97, 131, 132, 137, 141, 144.
- Regulator primary sources: see CHARTER §1 (above).
