# ITALIAN-COMPLIANCE.md — GreenMetrics

**Project:** GreenMetrics
**Last reviewed:** 2026-04-30 (charter-alignment annotation added; regulatory citations unchanged)
**Scope:** mapping of Italian and European regulatory obligations to GreenMetrics
code paths, seed data, and operational guarantees.

> **Framing note (2026-04-30 charter alignment):** GreenMetrics is delivered
> as a **modular template + engagement** model per `docs/MODULAR-TEMPLATE-CHARTER.md`
> §2 — not as a hosted SaaS. Italian regulatory obligations cited below are
> satisfied at the **per-engagement deployment** level (the deployment is the
> regulatory artefact-producing surface), with the Italian-flagship Region
> Pack (`packs/region/it/`) + Factor Packs (ISPRA / GSE / Terna / AIB) +
> Report Packs (ESRS E1 / Piano 5.0 / Conto Termico / TEE / audit 102/2014)
> being the load-bearing components per Rule 88. Code-path references below
> currently point at `internal/services/`; those paths migrate into Pack
> directories during Phase E Sprint S6 (Plan §5.4) — the regulatory mapping
> stays unchanged.

This file is the source of truth for the sales and legal team. Every citation
is traceable to a module in the codebase; every code path that ships as a
documented stub is marked explicitly.

---

## 1. Audit energetico / Efficienza energetica

### 1.1 D.Lgs. 4 luglio 2014, n. 102

Recepimento della Direttiva 2012/27/UE sull'efficienza energetica. L'art. 8
impone alle grandi imprese (e alle imprese energivore iscritte al registro
CSEA) una diagnosi energetica quadriennale secondo le norme tecniche
EN 16247-1/2/3/4. La diagnosi va depositata sul portale `audit102.enea.it`
entro il 5 dicembre dell'anno di riferimento.

- **Verifica fonte:** <https://www.normattiva.it> — atto 2014-07-04;102
- **Codice:** `backend/internal/services/audit_client.go` (preparazione
  dossier) + `backend/internal/services/report_generator.go`
  (`buildAuditDLgs102`).
- **Output GreenMetrics:** sezioni "Contesto energetico", "Analisi storica
  consumi ≥3 anni", "Modello energetico dei processi", "Determinazione IPE",
  "Opportunità di miglioramento e tempo di ritorno".

### 1.2 D.Lgs. 14 luglio 2020, n. 73

Aggiornamento del recepimento italiano a seguito della Direttiva 2018/2002
(pacchetto "Energia pulita per tutti gli europei"). Rafforza gli obblighi di
misura e rendicontazione; introduce criteri aggiuntivi per l'indicazione dei
risparmi energetici nell'ambito dei meccanismi di incentivazione (TEE, Conto
Termico, Piano Transizione 5.0).

- **Verifica fonte:** <https://www.normattiva.it> — atto 2020-07-14;73
- **Codice:** citato in `report_generator.go` buildAuditDLgs102; i calcoli di
  risparmio energetico in `ComputePianoTransizione50Result` rispettano le
  soglie e le banding tables conformi al decreto.

### 1.3 EN 16247

Pacchetto di norme tecniche UNI CEI EN 16247-1/2/3/4/5 per le diagnosi
energetiche (generale / edifici / processi industriali / trasporti /
competenze dell'auditor energetico). Non sono "leggi" ma sono richiamate
normativamente dal D.Lgs. 102/2014.

- **Codice:** `report_generator.go` — funzione `buildPianoTransizione50`
  cita EN 16247-3 come metodologia; `audit_client.go` menziona EN 16247-1/2/3/4.

---

## 2. CSRD / ESRS (sostenibilità)

### 2.1 Direttiva (UE) 2022/2464 — CSRD

Corporate Sustainability Reporting Directive. Estende l'obbligo di
rendicontazione di sostenibilità a ~50 000 aziende europee in modo
progressivo 2024–2028. Recepimento italiano: D.Lgs. 6 settembre 2024, n. 125.

- **Verifica fonte:** <https://eur-lex.europa.eu/eli/dir/2022/2464/oj> (EN/IT).
- **Codice:** `report_generator.go` — funzione `buildESRSE1` produce il
  payload dei data-point ESRS E1.
- **Test:** `backend/tests/report_generator_test.go` —
  `TestRenderESRSE1HTML` verifica la presenza di tutte le sezioni E1-5 /
  E1-6 / E1-7 nell'HTML.

### 2.2 Reg. Delegato (UE) 2023/2772 — ESRS set 1

Adozione del primo set di standard ESRS (E1 Climate, E2 Pollution, E3 Water,
E4 Biodiversity, E5 Resource use & circular economy, S1–S4 Social, G1
Business conduct). GreenMetrics supporta in modo nativo ESRS E1; gli altri
standard sono disponibili tramite data-import dal cliente.

- **Verifica fonte:** <https://eur-lex.europa.eu/eli/reg_del/2023/2772/oj>
- **Codice:** `buildESRSE1` emette:
  - **E1-5** consumo energetico totale, con split rinnovabile / non rinnovabile
    (split 0,35 / 0,65 come valore predefinito — configurabile per tenant).
  - **E1-6** emissioni lorde Scope 1, Scope 2 (location-based),
    Scope 3 rilevanti.
  - **E1-7** intensità GHG rispetto ai ricavi netti (valore 0 finché i ricavi
    non vengono inseriti dal tenant).

### 2.3 Stato degli ESRS non coperti nativamente

| Standard | Stato GreenMetrics |
|----------|-------------------|
| ESRS E2 — Pollution | fuori scope (documentato) — si importa da ASL / ARPA. |
| ESRS E3 — Water & marine | fuori scope — dati da fatturazione idrica. |
| ESRS E4 — Biodiversity | fuori scope — gestione documentale. |
| ESRS E5 — Resource use | fuori scope — integrabile con dati gestione rifiuti. |
| ESRS S1–S4 | fuori scope — integrabile con HR/TeamFlow. |
| ESRS G1 | fuori scope — gestione documentale. |

---

## 3. Piano Transizione 5.0

### 3.1 D.L. 2 marzo 2024, n. 19 (convertito in L. 29 aprile 2024, n. 56)

Istituisce il Piano Transizione 5.0 — credito d'imposta per investimenti in
beni strumentali materiali e immateriali "4.0" funzionali al risparmio
energetico. Soglie di ammissibilità:

- riduzione ≥ 3% dei consumi energetici del **processo** interessato, OR
- riduzione ≥ 5% dei consumi energetici della **struttura produttiva**.

Banding del credito (scaglioni di spesa ammissibile e aliquote, rif.
D.M. MIMIT-MASE 24 luglio 2024):

| Riduzione energetica | Aliquota base |
|----------------------|---------------|
| 3% – 6% | 5% |
| 6% – 10% | 20% |
| 10% – 15% | 35% |
| ≥ 15% | 40% |

- **Verifica fonte:** <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:decreto.legge:2024-03-02;19>
- **Verifica conversione:** <https://www.normattiva.it/uri-res/N2Ls?urn:nir:stato:legge:2024-04-29;56>
- **Codice:** `services.ComputePianoTransizione50Result` — calcolo
  deterministico esportato e testato in `report_generator_test.go`:
  - `TestPiano50_BaselineSaving3Percent` — baseline 100 000 kWh con
    risparmio esatto del 3% → fascia base 5%.
  - `TestPiano50_SiteSaving6Percent` — fascia intermedia 20% con soglia
    di sito raggiunta.
  - `TestPiano50_Saving15Percent` — fascia superiore 40%.
  - `TestPiano50_BelowThreshold` — al di sotto del 3% → non ammissibile.

### 3.2 D.M. MIMIT-MASE 24 luglio 2024

Decreto applicativo del Piano Transizione 5.0. GSE è incaricato del rilascio
del codice unico e della verifica dei risparmi. GreenMetrics produce
l'attestazione (HTML e JSON) firmabile da un EGE (UNI CEI 11339) o da un
auditor EN 16247-5.

- **Verifica fonte:** <https://www.mimit.gov.it/> — pubblicato sulla Gazzetta
  Ufficiale del 06/08/2024.
- **Codice:** firma nel template
  `pianoTransizione50HTMLTemplate` in `report_generator.go`.

---

## 4. Altri incentivi energetici italiani

### 4.1 Conto Termico 2.0 — D.M. 16 febbraio 2016

Incentivi GSE per interventi di incremento dell'efficienza energetica e
produzione di energia termica da fonti rinnovabili. Copre pompe di calore,
solare termico, caldaie a biomassa, isolamento, impianti ibridi.

- **Verifica fonte:** <https://www.gse.it/servizi-per-te/efficienza-energetica/conto-termico>
- **Codice:** `report_generator.go` — `buildContoTermico`. Lo stub produce
  il tipo di intervento (esempio "2.C — sostituzione di impianti di
  climatizzazione invernale esistenti con pompe di calore"), il portale GSE
  di presentazione e il riferimento normativo.

### 4.2 Certificati Bianchi / TEE — D.M. 11 gennaio 2017

Titoli di Efficienza Energetica. Il progetto di risparmio è misurato in
tonnellate equivalenti di petrolio (tep) risparmiate per anno. Emissione
dei TEE gestita dal GSE.

- **Verifica fonte:** <https://www.gse.it/servizi-per-te/efficienza-energetica/certificati-bianchi>
- **Codice:** `report_generator.go` — `buildCertificatiBianchi`. Il report
  produce il payload di submission iniziale; la presentazione al portale
  GSE CB è esterna.

### 4.3 Direttiva 2003/87/CE — EU ETS

Sistema europeo di scambio quote emissioni (cap-and-trade). Applicabile a
grandi impianti (potenza termica ≥ 20 MW) e settori industriali inclusi
negli allegati. GreenMetrics non opera il registro ETS ma può esportare i
dati di emissione verso i sistemi di monitoraggio ETS verificati da
ente accreditato (art. 18 Dir. 2003/87/CE).

- **Verifica fonte:** <https://eur-lex.europa.eu/legal-content/IT/TXT/?uri=CELEX%3A02003L0087-20240301>
- **Stato GreenMetrics:** export dati via API `/api/v1/readings/export` (CSV).
  Nessun'integrazione diretta con il registro; documentata come export.

---

## 5. Fattori di emissione (ISPRA)

### 5.1 Mix elettrico italiano

ISPRA pubblica annualmente i fattori di emissione della produzione elettrica
italiana nel rapporto "Fattori di emissione per la produzione ed il consumo
di energia elettrica in Italia". Versioni in uso:

| Anno | Fattore (kg CO2e / kWh) | Fonte |
|------|------------------------|-------|
| 2022 | 0,263 | ISPRA Rapporto 386/2023 |
| 2023 | 0,250 | ISPRA Rapporto 404/2024 |
| 2024 | 0,245 (provvisorio) | ISPRA 2024 stima provvisoria |

- **Verifica fonte:** <https://www.isprambiente.gov.it/it/pubblicazioni/rapporti/fattori-di-emissione-per-la-produzione-ed-il-consumo-di-energia-elettrica-in-italia>
- **Codice:** migration `0005_emission_factors.sql` — seed con chiave
  composita `(code, valid_from)`; default cache in
  `carbon_calculator.go` (`defaultFactors`).
- **Test:** `carbon_calculator_test.go` — `TestCarbonCalculator_Scope2_ISPRA2023`
  verifica esattamente 0,250 × kWh = kg CO2e; la controparte 2024 è
  verificata da `TestCarbonCalculator_Scope2_ISPRA2024`.

### 5.2 Gas naturale — D.M. 11 maggio 2022 (MiTE)

Parametri standard nazionali per la combustione del gas naturale: 1,975 kg
CO2e per Sm3 (combustione stazionaria).

- **Verifica fonte:** <https://www.mase.gov.it/sites/default/files/archivio/allegati/DGP/dm_11-05-2022.pdf>
- **Codice:** factor `NG_STATIONARY_COMBUSTION` in `defaultFactors`.
- **Test:** `TestCarbonCalculator_Scope1_NaturalGas` — 10 000 Sm3 × 1,975 =
  19 750 kg CO2e.

### 5.3 Mix europeo residuale — AIB

Association of Issuing Bodies (AIB) — residual mix per il calcolo
"market-based" secondo GHG Protocol Scope 2 Guidance. Valore 2023 per
l'Italia: 0,457 kg CO2e / kWh.

- **Verifica fonte:** <https://www.aib-net.org/facts/european-residual-mix>
- **Codice:** factor `IT_ELEC_RESIDUAL_MIX_2023`.
- **Test:** `TestCarbonCalculator_MarketBased`.

### 5.4 Altri combustibili

| Fattore | Valore | Fonte |
|---------|--------|-------|
| Gasolio (diesel) | 2,650 kg CO2e / L | ISPRA 2024 |
| GPL | 1,510 kg CO2e / L | ISPRA 2024 |
| Olio combustibile | 2,771 kg CO2e / L | ISPRA 2024 |
| Teleriscaldamento (media IT) | 0,200 kg CO2e / kWh | Riferimento Conto Termico 2.0 |

---

## 6. Protezione dati (GDPR)

GreenMetrics tratta dati aziendali (consumi energetici, fatture, risorse
umane aggregate). Nessun dato personale sensibile è raccolto di default.

- **Base normativa:** Reg. (UE) 2016/679 (GDPR); D.Lgs. 30 giugno 2003, n.
  196 come modificato dal D.Lgs. 10 agosto 2018, n. 101.
- **Privacy by design:** ingestion Modbus/M-Bus non porta PII; login
  (email + password hash bcrypt/argon2 — TBD) è l'unico punto PII.
- **Diritto di cancellazione:** endpoint `DELETE /api/v1/tenants/me` da
  implementare in `Mission III`; nel frattempo procedura manuale.

---

## 7. Interoperabilità con altri progetti Mission II

- **FatturaFlow:** `POST /api/v1/invoices/from-cost-centre-bill` — il
  cost-centre di GreenMetrics produce la righe per fatturazione interna.
- **TeamFlow:** in lettura, per incroci "energia per addetto".
- **SmartERP:** in lettura, per calcolo "kWh per unità prodotta".

---

## 8. Integrazioni con sistemi italiani (stato)

| Sistema | Stato | Note |
|---------|-------|------|
| ISPRA (fattori) | **Implementato** — seed + cache | Aggiornamento annuale documentato. |
| ENEA `audit102` | **Stub documentato** | Firma digitale richiede SPID/CNS. |
| GSE Conto Termico 2.0 | **Stub documentato** | Portale non API; invio manuale. |
| GSE CB (TEE) | **Stub documentato** | Idem. |
| Terna (Download Centre) | **Stub documentato** con shape | Autenticazione OAuth2 richiesta. |
| E-Distribuzione (SMD Chain2) | **Stub documentato** con shape | Richiede OAuth2 + contratto. |
| SPD (Servizio Portale Distribuzione) | **Stub documentato** | Client-certificate. |

Tutti gli stub sono incapsulati in `smart_meter_client.go` (Terna,
E-Distribuzione, SPD), `audit_client.go` (ENEA) e nelle funzioni `build*`
dei report GSE; ogni stub emette un log strutturato e restituisce un payload
vuoto coerente.

---

## 9. Roadmap di verifica periodica

1. **Annuale — aprile:** aggiornare il fattore ISPRA del mix elettrico
   italiano (nuova versione del Rapporto ISPRA). Aggiungere un nuovo record
   alla migration `0005_emission_factors.sql`; **non** modificare i record
   storici.
2. **Annuale — luglio:** rivedere i residual mix AIB.
3. **Semestrale:** rivedere il listato dei fattori (diesel, GPL, olio
   combustibile) rispetto alla tabella ISPRA.
4. **Evento — pubblicazione DM attuativo Piano 5.0:** rivedere le aliquote e
   le soglie in `ComputePianoTransizione50Result`; aggiungere test.
5. **Evento — pubblicazione nuovo set ESRS:** estendere `buildESRSE1` o
   aggiungere un nuovo builder (`buildESRSE2`, `buildESRSE3`, ecc.).

---

## 10. Riferimenti ufficiali

- <https://www.normattiva.it> — normativa italiana vigente.
- <https://eur-lex.europa.eu> — diritto dell'Unione.
- <https://www.isprambiente.gov.it> — fattori di emissione.
- <https://www.gse.it> — GSE (Conto Termico 2.0, Certificati Bianchi, Piano 5.0).
- <https://www.mimit.gov.it> — MIMIT (Ministero delle Imprese e del Made in Italy).
- <https://www.mase.gov.it> — MASE (Ambiente e Sicurezza Energetica).
- <https://www.terna.it> — Terna S.p.A. (TSO).
- <https://www.e-distribuzione.it> — E-Distribuzione (DSO).
- <https://www.arera.it> — ARERA.
- <https://www.mercatoelettrico.org> — GME.
