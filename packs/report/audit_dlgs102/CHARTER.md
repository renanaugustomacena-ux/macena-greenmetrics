# Audit Energetico D.Lgs. 102/2014 — Report Pack — Charter

> **Pack:** `report-audit_dlgs102` · **Kind:** Report · **Version:** 1.0.0
> **Pack-contract version:** 1.0.0 · **Min Core version:** 1.0.0
> **Status:** GA (highest commercial-volume Italian regulatory dossier; ~10 000 grandi imprese + imprese energivore on a 4-yearly cadence)
> **Doctrine refs:** Rules 89 (bit-perfect reproducibility), 91 (pure functions), 95 (provenance bundle), 97 (algorithm versioning), 131 (formal-spec validation), 132 (primary-source citation), 137 (regulatory counter-signature), 141 (deterministic serialisation), 144 (signed at finalisation).
> **Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Report Packs).
> **Regulatory base (primary sources):**
>
> - **D.Lgs. 4 luglio 2014 n. 102** art. 8 — Decreto legislativo "Attuazione della direttiva 2012/27/UE sull'efficienza energetica". <https://www.gazzettaufficiale.it/eli/id/2014/07/18/14G00114/sg>
> - **D.Lgs. 18/07/2016 n. 141** — primo aggiornamento (recepimento osservazioni Commissione UE sulle PMI energivore).
> - **D.Lgs. 73/2020** art. 1 c. 3-ter — esenzione dall'obbligo di diagnosi per imprese con consumi totali inferiori a 50 tep/anno (introduzione c. 3-bis art. 8).
> - **ENEA "Indicazioni Operative"** per la conformità all'art. 8 — portale Audit 102. <https://www.efficienzaenergetica.enea.it/servizi-per/imprese/diagnosi-energetiche/indicazioni-operative.html>
> - **ENEA "Linee Guida e Manuale Operativo"** Diagnosi Energetica + Allegato 2 — clusterizzazione, rapporto di diagnosi, piano di monitoraggio. Portale: <https://audit102.enea.it/>
> - **UNI CEI EN 16247-1..4** — norme tecniche di riferimento per le diagnosi energetiche generali (parte 1), per gli edifici (parte 2), per i processi produttivi (parte 3), per i trasporti (parte 4).
> - **ARERA Delibera 03/08 EEN** — Aggiornamento del fattore di conversione dei kWh in tonnellate equivalenti di petrolio (TEE / TEP); usato anche dai diagnostici art. 8 in assenza di calcolo addizionale di catena di trasformazione. <https://www.arera.it/atti-e-provvedimenti/dettaglio/08/003-08een>
> - **D.M. 20 luglio 2004** — primo provvedimento ministeriale sui fattori di conversione kWh ↔ tep (poi aggiornato da ARERA 03/08).
>
> Sources accessed 2026-04-30 for Pack v1.0.0.

## 1. Purpose

This Pack produces the **quantitative half of the legally-mandated diagnosi energetica** ("audit energetico") that D.Lgs. 102/2014 art. 8 imposes on the grandi imprese (>250 employees AND turnover >€50M OR balance sheet >€43M) and on the imprese energivore (registered in the CSEA list — typically ≥2.4 GWh/year electricity use plus sectoral criteria). The audit cadence is **every 4 years**, deadline **5 December**, the next deadline being typically December 2027 for the cycle that opened December 2023.

Specifically the Pack covers:

- **Energy baseline** — total consumption per energy vector (electricity, natural gas, diesel, petrol, LPG, coal, heavy fuel oil) in kWh and converted to tep using statutory ENEA / ARERA conversion factors per Rule 132. Site-level breakdown (one site = one MeterID).
- **Obligation block** — audit type (`grande_impresa` / `energivora` / `voluntary`), 4-yearly periodicity, next deadline, exemption basis (ISO 50001 / EMAS / ISO 14001) and the <50 tep/year exemption introduced by D.Lgs. 73/2020.
- **Below-50-tep flag** — declarative; if total consumption < 50 tep/year, the obligation does not apply (Art. 8 c. 3-bis).
- **EnPI candidates** — for the simplest case (total energy intensity = total kWh / total floorspace m²); per-site EnPI is left to the orchestrator since site classification (industrial / tertiary) is engagement-side knowledge of ATECO codes.
- **Site breakdown** — per-MeterID kWh + tep totals, per-vector. The Pack does NOT classify sites as industrial vs tertiary; that classification depends on ATECO codes which flow from client.

Narrative content (UNI CEI EN 16247-1..4 compliance statement, site-visit summaries, scope description, EGE / ESCo certifier identity, improvement-measure list with payback analysis, monitoring-plan KPIs) is supplied by the engagement client and bundled into the final dossier by the engagement-fork's reporting orchestrator (Plan §5.4). The ENEA-portal submission flow (audit102.enea.it) is a separate engagement-side Pack ("submitter-enea-audit102") landing in Phase H Sprint S16.

Per Rule 137, the diagnosi must be signed by a UNI CEI 11339 EGE (Esperto in Gestione dell'Energia) or a UNI CEI 11352 ESCo (Energy Service Company); the engagement Phase 0 Discovery verifies the relationship is in place.

## 2. Algorithm

Pure function `Build(ctx, period, factors, readings) → Report`.

```text
1. Iterate readings (each AggregatedRow's MeterID is treated as the site_id):
   for each row:
     classify by Unit:
       Wh / kWh           → electricity   (kWh_per_unit = 1 for kWh, 0.001 for Wh)
       Sm3                → natural_gas
       l_diesel /
       l_diesel_vehicle   → diesel        (× density 0.835 kg/l → kg)
       l_petrol           → petrol        (× density 0.745 kg/l → kg)
       kg_lpg             → lpg
       kg_coal            → coal
       kg_heavy_fuel      → heavy_fuel_oil
       any other          → unclassified  (counts; not in baseline totals)
     accumulate per (site_id, vector) → primary-unit total
                  per vector            → primary-unit total
                  per site_id           → multi-vector dict

2. Convert primary-unit totals to kWh + tep per vector using statutory factors
   (override-able via FactorBundle keys `audit102_tep_factor_<vector>` per
   Rule 90 — per-deployment factor versions are an engagement responsibility):

     vector              statutory tep factor      source
     ─────────────────────────────────────────────────────────────────────
     electricity         0.000187 tep/kWh          ARERA Delibera 03/08 EEN
     natural_gas         0.000836 tep/Sm3          D.M. 20/07/2004 (gas naturale)
     diesel              0.001080 tep/kg           ARERA + D.M. 20/07/2004
     petrol              0.001050 tep/kg           ARERA + D.M. 20/07/2004
     lpg                 0.001099 tep/kg           ARERA + D.M. 20/07/2004
     coal                0.000700 tep/kg           ARERA + D.M. 20/07/2004
     heavy_fuel_oil      0.000980 tep/kg           ARERA + D.M. 20/07/2004

   Each tep factor carries a `// Source:` comment block in builder.go and is
   overrideable via FactorBundle.

3. Compute per-vector aggregates:
     kWh_<vector>  — primary-unit total (electricity in kWh, fuels in their
                     primary unit; non-electricity quantities reported in
                     primary unit AND converted to kWh-equivalent only when
                     LHV factors are supplied via FactorBundle as
                     `audit102_lhv_<vector>` — TODO Phase F).
     tep_<vector>  — primary-unit × statutory-tep-factor.

4. Compute totals:
     total_tep = ∑ tep_<vector>
     total_kwh = ∑ kWh_<vector>
     below_50_tep_flag = (total_tep < 50)

5. Read obligation factors from FactorBundle:
     audit102_obligation_type    — 0 = grande_impresa, 1 = energivora, 2 = voluntary
     audit102_below_50_exempt    — 0/1 (engagement Phase 0 Discovery declares)
     audit102_exemption_iso      — 0 = none, 1 = ISO 50001, 2 = EMAS, 3 = ISO 14001
     audit102_next_deadline_unix — Unix timestamp of next 5-Dec deadline
     audit102_total_floor_m2     — engagement-supplied for total-energy-
                                   intensity EnPI candidate

6. Build EnPI candidates:
     If audit102_total_floor_m2 > 0 and total_kwh > 0:
       emit { name: "Total energy intensity", indicator: "kWh/m²",
              value: total_kwh / floor_m2, baseline_period: <period> }

7. Render the canonical typed Body (energy_baseline + site_breakdown +
   obligation + factors_used + enpi + improvement_measures (empty placeholder)
   + monitoring_plan (empty) + narrative + notes).

8. Serialise deterministically (Rule 141): JSON with sorted struct fields,
   2-space indent, trailing newline. Per-vector + per-site lists are sorted
   by name for determinism.

9. Provenance is populated by Core's reporting orchestrator at signed-state
   transition (Rule 144).
```

The function is **pure** (Rule 91). The conformance test asserts byte-identical Encoded across two consecutive `Build` calls (Rule 89).

## 3. Output shape

```json
{
  "report":     "audit_dlgs102",
  "regulator":  "ENEA / MASE — D.Lgs. 4 luglio 2014 n. 102 art. 8",
  "period":     { ... },
  "obligation": {
    "type":                       "grande_impresa",
    "audit_periodicity_years":    4,
    "next_deadline_iso":          "2027-12-05T23:59:59Z",
    "exemption_basis":            null,
    "below_50_tep_exemption":     false
  },
  "factors_used": {
    "audit102_tep_factor_electricity":   { "value": 0.000187, "unit": "tep/kWh", "version": "ARERA-D-3-08" },
    "audit102_tep_factor_natural_gas":   { "value": 0.000836, "unit": "tep/Sm3", "version": "DM-20-07-2004" },
    "audit102_obligation_type":          { "value": 0, "unit": "selector", "version": "engagement-supplied" },
    "audit102_total_floor_m2":           { "value": 12500, "unit": "m2", "version": "engagement-supplied" }
  },
  "energy_baseline": {
    "by_vector": [
      { "vector": "electricity", "primary_unit": "kWh", "primary_total": 8500000, "tep_total": 1589.5, "tep_factor": 0.000187 },
      { "vector": "natural_gas", "primary_unit": "Sm3", "primary_total":  150000, "tep_total":  125.4, "tep_factor": 0.000836 }
    ],
    "total_kwh":              8500000,
    "total_tep":                1714.9,
    "below_50_tep_threshold": false
  },
  "site_breakdown": [
    {
      "site_id":   "00000000-0000-0000-0000-000000000001",
      "by_vector": [...],
      "kwh_total": 5000000,
      "tep_total": 935.0
    }
  ],
  "enpi": [
    { "name": "Total energy intensity", "indicator": "kWh/m2", "value": 680.0, "baseline_period": "2026" }
  ],
  "improvement_measures": [],
  "monitoring_plan": {
    "kpis": [],
    "review_frequency_months": 12
  },
  "narrative_data_points": {
    "scope_description":                       null,
    "site_visit_summaries":                    null,
    "uni_cei_en_16247_compliance_statement":   null,
    "ege_certifier_id":                        null,
    "note": "Narrative content (UNI CEI EN 16247-1..4 compliance statement, site-visit summaries, scope description, EGE / ESCo certifier identity, improvement-measure list with payback analysis, monitoring-plan KPIs) is bundled by the engagement-fork's reporting orchestrator from client-supplied content (Plan §5.4)."
  },
  "ege_certification_required": true,
  "unclassified_rows":          0,
  "notes": [...]
}
```

## 4. Tradeoff Stanza

- **Solves:** the absence of a regulator-grade Italian-flagship D.Lgs. 102/2014 art. 8 quantitative-baseline + obligation-block Builder. Approximately 10 000 grandi imprese + imprese energivore are subject to the audit on a 4-yearly cadence (next mass deadline 2027-12-05); the Pack mechanises the energy-baseline math and the regulatory bookkeeping (cadence, exemptions, below-50-tep) so the engagement consultant focuses on the audit's qualitative half (UNI CEI EN 16247 compliance, recommendation analysis, site-visit narratives).
- **Optimises for:** primary-source citation on tep conversion factors (Rule 132 — every factor carries a `// Source:` comment block); audit-grade reproducibility (Rule 89 byte-identical replay); structural readiness for ENEA portal submission in Phase H Sprint S16 (the Body shape is field-aligned to the audit102.enea.it XML schema; the submitter Pack reads it 1:1).
- **Sacrifices:** site-classification (industrial vs tertiary) is left to the orchestrator since ATECO classification is engagement-side knowledge; the cluster-selection rules (≥10 000 tep industrial mandatory, ≥1 000 tep tertiary mandatory, <100 tep optionally excludable under 20% rule) are surfaced descriptively in `notes` and computed by the orchestrator. Lower-heating-value (LHV) conversion to kWh-equivalent for non-electricity vectors is deferred to Phase F (engagement supplies LHV via `audit102_lhv_<vector>` FactorBundle keys); current implementation reports primary-unit + tep only for non-electricity sources, which is sufficient for the obligation-status calculation but not for the EnPI kWh-equivalent. Per-fuel density factors (0.835 kg/l for diesel, 0.745 kg/l for petrol) are statutory averages and may diverge from per-supplier batch densities; engagement Phase 0 Discovery may override via factor `audit102_density_<vector>`.
- **Residual risks:** ENEA portal XML schema evolution mid-cycle (Rule 138 annual review is the closure; the `submitter-enea-audit102` Pack version-binds to the portal version current at submission); EGE / ESCo relationship gap at Phase 0 Discovery (Rule 137 mitigation — engagement bootstrap runbook explicitly verifies); tep-factor-version drift between ARERA delibere updates and engagement deployment (Rule 90 mitigation — factors overridable via FactorBundle, version-stamped via Versions()).

## 5. EGE / ESCo counter-signature (Rule 137)

D.Lgs. 102/2014 art. 8 c. 1 requires that the diagnosi energetica be eseguita da:

- un **EGE** (Esperto in Gestione dell'Energia, certificato per UNI CEI 11339), o
- un **auditor di una ESCo** (Energy Service Company, certificata per UNI CEI 11352), o
- un **auditor energetico** certificato per la UNI CEI EN 16247-5.

Per Rule 137, l'engagement Phase 0 Discovery verifies the certifier relationship is in place; the Pack does NOT block at generation time, but the engagement-fork's signed-state transition (Rule 144) requires the certifier signature before Cosign-signing the dossier.

## 6. Layout note (interim, single-module phase)

Pack Go code lives at `backend/packs/report/audit_dlgs102/`. Pack manifest + CHARTER stay at the repo-root `packs/report/audit_dlgs102/` per Charter §3.2 discovery convention. See ADR-0024.

## 7. Cross-references

- Pack contract: `backend/internal/domain/reporting/builder.go`.
- Pack manifest: `packs/report/audit_dlgs102/manifest.yaml`.
- Implementation: `backend/packs/report/audit_dlgs102/builder.go`.
- Tests: `backend/packs/report/audit_dlgs102/builder_test.go`.
- Factor dependencies: `packs/factor/ispra/` (the Pack uses ISPRA-published per-fuel default factors as a fallback when engagement-supplied overrides are absent).
- Sister Italian Report Packs: `packs/report/{esrs_e1,piano_5_0,conto_termico,tee,monthly_consumption,co2_footprint}/`.
- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2.
- Doctrine: `docs/DOCTRINE.md` Rules 89, 90, 91, 95, 97, 131, 132, 137, 141, 144.
- Regulator primary sources: see CHARTER §1 (above).
