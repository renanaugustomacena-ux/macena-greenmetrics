# Competitive Brief — Industrial Energy Management & Sustainability Reporting

**Audience:** GreenMetrics platform & product engineering.
**Purpose:** Identify gaps and beat-points in the competitive field for an industrial energy-management & sustainability-reporting platform serving Italian and EU energy-intensive SMEs (CSRD/ESRS E1, Piano Transizione 5.0 attestazione, D.Lgs. 102/2014 audit energetico, Conto Termico 2.0/3.0, Certificati Bianchi TEE, GHG Protocol Scope 1/2/3).
**Use:** This document feeds the modular-template positioning. Each entry ends with one or more concrete "beat-points" we can hard-code into the doctrine.

---

## 0. TL;DR — The five beat-points we should design for

1. **Audit-grade reproducibility.** No incumbent guarantees bit-perfect re-derivation of last year's report from raw data + the *factor version that was valid at filing time*. Most rebuild on demand from current factors. We treat factor versioning as a hard primary key and snapshot every report's input bundle.
2. **OT-native ingestion.** Watershed/Persefoni/Greenly/Sweep/IBM Envizi/Microsoft/Salesforce all assume "you upload utility bills or pull from your ERP". Schneider/Siemens/ABB/Honeywell/Rockwell/Emerson are OT-native but vertically tied to their own hardware. There is open white space for a vendor-neutral OT-native ingestor stack covering Modbus RTU/TCP, M-Bus, SunSpec, OCPP, IEC 61850, EtherNet/IP, BACnet, MQTT Sparkplug B, and Italian DSO portals (E-Distribuzione SMD, Terna, SPD).
3. **Italian-first regulatory depth.** No global incumbent ships native Piano 5.0 attestazione generation, GSE Conto Termico 3.0 dossier emission, ENEA D.Lgs. 102/2014 audit format, or ISPRA-versioned factors with valid-from semantics. Maps Energy and a handful of Italian niche players cover slices, but none combine all of them in a single deterministic engine.
4. **Modular template, not SaaS.** Every competitor sells a hosted multi-tenant SaaS or an enterprise on-prem licence per site. None ship a clean, audited, deployable template that a client engineering team can own end-to-end on its own infrastructure. The Linux Foundation Energy projects come closest in spirit but are ad-hoc, narrow, and not Italian-compliant.
5. **Engineering doctrine as part of the deliverable.** Auditors and procurement reviewers treat the codebase, ADRs, supply-chain attestations, and runbooks as evidence. Most competitors ship a black-box SaaS with an SOC 2 Type II PDF and no further substrate. We ship a codebase with 200+ rule doctrine, signed images, attested provenance, runbooks per failure mode, and a transparency posture that *makes* the regulatory defensibility argument.

---

## 1. Market Segmentation

The competitive field divides into nine categories, each with a distinct buyer, deployment model, and weakness pattern.

| Tier | Category | Buyers | Deployment | Italian-regulation depth | OT-native | Audit reproducibility | Source-availability |
|------|----------|--------|------------|--------------------------|-----------|------------------------|---------------------|
| T1 | Enterprise OT/automation | Plant ops, automation engineers | Hybrid (cloud + edge appliance) | Light — Italian regulation as report template | High (own hardware) | Black-box, vendor lock-in | Closed |
| T2 | ESG-focused SaaS | Sustainability lead, CFO | Pure SaaS multi-tenant | Light to medium — CSRD-ready, Piano 5.0 N/A | Low — bills + ERP | Black-box | Closed |
| T3 | Hyperscaler suites | Data lead, IT architect | SaaS bound to hyperscaler | Light — global frameworks | Low | Limited (some lineage in Microsoft Fabric) | Closed |
| T4 | ESG ratings (adjacent) | Investor relations, supply-chain mgr | Web portal | None | None | Methodology proprietary | Closed |
| T5 | Italian utility/ESCO bundled | Energy buyer, plant manager | Bundled with energy contract | Medium — regulation handled as a service | Low to medium | Service-based, not platform-based | Closed |
| T6 | Specialist energy mgmt SaaS | Energy manager, EGE consultant | SaaS + edge gateway | Light to medium | Medium (Modbus mainly) | Limited | Closed |
| T7 | Time-series / process historians | OT integrator, SCADA engineer | On-prem-first | None | Very high | N/A (raw data only) | Closed (mostly) |
| T8 | Open-source reference impls | Utility R&D teams | DIY | None | Variable | Project-dependent | Open |
| T9 | DIY frankenstack | In-house data team | Self-hosted | None | DIY | DIY | Open |

GreenMetrics' modular template positioning lives between T6, T7, T8 and T9 — borrowing OT depth from T7, modularity from T8, reproducibility discipline from no one (greenfield), and Italian regulation depth from no one (greenfield). T5 is the channel partner archetype we co-sell *into*, not against.

---

## 2. Tier 1 — Enterprise OT/Automation Platforms

These are the heavyweights. Owned by Schneider, Siemens, ABB, Honeywell, Emerson, Rockwell. The buy-pattern is bundled with PLC/SCADA infrastructure or sold to multinational sustainability functions with six-figure budgets per site per year.

### 2.1 Schneider Electric — EcoStruxure Resource Advisor

- **Ownership:** Schneider Electric SE (FR/CH).
- **Deployment:** Cloud SaaS + on-prem EcoStruxure Panel Server (PAS600) gateway for Modbus RTU/TCP aggregation. Vendor-locked toward Schneider switchgear and meters, but speaks open Modbus.
- **Strengths:**
  - Aggregates 400+ data streams; multi-site portfolio reporting.
  - 2025 expansion of CSRD-ready ESG capability and a "Renewables & Carbon" experience that automates EAC/REC allocation.
  - "Resource Advisor Copilot" — conversational AI layer for ad-hoc queries.
  - Schneider field-service network handles physical install in EU markets including Italy.
- **Weaknesses:**
  - Pricing is preventivo-only ([TrustRadius](https://www.trustradius.com/products/schneider-electric-ecostruxure/pricing)). Reported field rates: €50k–€300k+/site/year for a fully-instrumented mid-size manufacturer.
  - Italian regulation handled as a report template, not first-class. No Piano 5.0 attestazione generator; no Conto Termico GSE flow.
  - Reproducibility: report rebuilds reflect *current* factor sets, not the factor version valid at filing — a known auditor pain point.
  - Vendor pull: integrators report soft pressure to standardise on Schneider PowerLogic meters; M-Bus/SunSpec coverage exists but lower priority.
- **Italy notes:** Schneider Italia has a Veneto sales presence; partnership with Confindustria; strong with multinational subsidiaries operating in Italy, weaker with Italian-only mid-cap.
- **Where we beat:** Italian regulation depth, audit reproducibility (factor versioning), price asymmetry, vendor-neutral hardware story, full source visibility for the auditor.
- Source: [Schneider EcoStruxure Resource Advisor](https://www.se.com/us/en/work/services/se-advisory-services/intelligent-software/resource-advisor/), [Resource Advisor Copilot blog](https://perspectives.se.com/blog-stream/building-sustainabilitys-digital-future-with-ecostruxure-resource-advisor-copilot), [Net Zero Compare](https://netzerocompare.com/software/schneider-electric-ecostruxure-resource-advisor).

### 2.2 Siemens — Sinalytics + SIMATIC Energy Manager

- **Ownership:** Siemens AG (DE).
- **Deployment:** Sinalytics is the analytics platform spine; SIMATIC Energy Manager is the operator-facing product. Deployed on-prem next to PLCs (S7-1500) or in Siemens Cloud for X.
- **Strengths:**
  - 300k+ devices connected; cyber-security-by-design approach with encrypted-at-rest analytics.
  - Deep PROFINET / PROFIBUS / IEC 61850 native support (Siemens hardware lineage).
  - Tight integration with TIA Portal and Siemens MES MOM.
  - SIMATIC Energy Manager handles per-cost-center allocation, ISO 50001 conformance, automated KPI tracking.
- **Weaknesses:**
  - Heavyweight commissioning. Reported integration windows of 6–12 months for non-Siemens shops.
  - License model is commercial-per-tag with year-on-year increases; EUR 120–200/tag/year typical for energy points alone.
  - Italian regulation is a report template, not native. No Piano 5.0 calculus, no Conto Termico GSE submission, no ENEA audit format.
  - Closed source historian (the ProcessHistorian / Information Server pair).
- **Italy notes:** Strong installed base in Brescia/Verona machinery cluster via Siemens Italia; complex pricing on retrofits.
- **Where we beat:** Italian regulation, deployment time (5 minutes vs 6 months), licence economics, source visibility, neutral hardware support.
- Source: [Siemens Sinalytics press](https://www.siemens.com/press/PR2016040260PSEN), [SIMATIC Energy Manager](https://www.siemens.com/en-us/products/simatic-energy-management/energy-manager/).

### 2.3 ABB — Ability Energy Manager / Ability OPTIMAX

- **Ownership:** ABB Ltd (CH).
- **Deployment:** Cloud SaaS or on-prem; ABB markets a "<1 day commissioning" baseline for sites where ABB switchgear/meters are present.
- **Strengths:**
  - Pre-engineered functionalities for low-voltage distribution; strong Modbus + ABB EMAX2 native.
  - Real-time monitoring and identification of inefficiencies; multi-site portfolio support.
  - Microsoft Azure marketplace listing; tight Power BI integration.
- **Weaknesses:**
  - Italian regulation: light. CSRD-aligned reporting yes; Piano 5.0 attestazione no.
  - Reproducibility: marketing-grade, not auditor-grade — no published guarantee on factor-version pinning.
  - Vendor lock toward ABB hardware for full feature parity.
- **Italy notes:** ABB has Italian sites in Bergamo and Milan; partnerships with regional ESCOs.
- **Where we beat:** Piano 5.0, factor-versioning, vendor neutrality, modularity (their on-prem path is a per-site licence; ours is a forkable template).
- Source: [ABB Ability Energy Manager](https://electrification.us.abb.com/products/energy-management-systems/abb-ability-energy-manager), [ABB Ability OPTIMAX](https://www.abb.com/global/en/areas/automation/solutions/industrial-software/energy-management/energy-optimization-optimax).

### 2.4 Honeywell — Forge Sustainability+ for Buildings

- **Ownership:** Honeywell International (US).
- **Deployment:** SaaS; building-management-system (BMS) integration pattern. Buildings-first focus, with light spillover to industrial.
- **Strengths:**
  - Disaggregated asset-level analytics; sensor + ML automatic zone-level controls.
  - Energy Star certification reporting; Scope 1/2 utility tracking.
  - Strong commercial real-estate/BMS integration (Tridium Niagara, Trane, Carrier).
- **Weaknesses:**
  - Buildings, not industrial. Process-industry features (compressed air, steam, chilled water systems) are anaemic.
  - Italian regulation absent. CSRD coverage exists for North-American customers; ESRS E1 partial; Piano 5.0/D.Lgs. 102/2014 not addressed.
  - SaaS only; no on-prem path; no source visibility.
- **Italy notes:** Used by Italian subsidiaries of US/UK multinationals in commercial real-estate, not by Italian SMEs.
- **Where we beat:** Industrial scope, Italian regulation, deployment flexibility, source visibility.
- Source: [Honeywell Forge Sustainability+ for Buildings](https://buildings.honeywell.com/us/en/solutions/buildings/honeywell-forge-sustainability-plus-for-buildings-carbon-and-energy-management).

### 2.5 Emerson — Ovation Green

- **Ownership:** Emerson Electric (US).
- **Deployment:** Hybrid; Ovation SCADA platform extended for renewables and storage management.
- **Strengths:**
  - Manufacturer-agnostic SCADA across renewables + BESS.
  - Strong on-prem story; high-availability deployment baked in.
  - Designed for utility-scale renewables (multi-MW); also serves industrial PV.
- **Weaknesses:**
  - Renewables-and-storage centric; weak as a generic energy-management platform for a manufacturing site.
  - No ESG/CSRD reporting layer; relies on third-party reporting bolt-ons.
  - No Italian regulation support.
- **Italy notes:** Limited Italian footprint outside the utility-scale renewables segment.
- **Where we beat:** Industrial breadth, ESG/regulatory layer, Italian compliance, deployment economics.
- Source: [Emerson Ovation Green](https://www.emerson.com/en-us/automation/ovation-green).

### 2.6 Rockwell Automation — FactoryTalk Energy Manager

- **Ownership:** Rockwell Automation (US).
- **Deployment:** On-prem-first; built atop FactoryTalk DataMosaix; AI-driven energy loss recommendations.
- **Strengths:**
  - Plant/process/line/machine-granular energy data capture, native to Rockwell PLCs (CompactLogix/ControlLogix).
  - AI recommendations for shop-floor energy losses.
  - Software CEM (Continuous Emission Monitoring) module — relevant for regulated emissions reporting.
  - Reported impact: 15–30% annual energy savings, 20–40% Scope 1/2 reduction.
- **Weaknesses:**
  - Strong only inside the Rockwell stack; non-Allen-Bradley PLCs require a glue layer.
  - North America centric; Italian/EU regulatory coverage weak.
  - Closed historian; commercial-per-tag licensing.
- **Italy notes:** Some industrial automation footprint via Rockwell Italia; less adopted than Siemens.
- **Where we beat:** Vendor neutrality, EU regulatory depth, modular template positioning, transparency.
- Source: [FactoryTalk Energy Manager](https://www.rockwellautomation.com/en-us/products/software/factorytalk/innovationsuite/factorytalk-energy-manager.html), [Rockwell sustainability press](https://www.rockwellautomation.com/en-us/company/news/press-releases/Rockwell-Automation-Advances-Sustainability-Through-Smart-Manufacturing.html).

---

## 3. Tier 2 — ESG-focused SaaS

These are the sustainability-software pure-plays. Buyer is the CSO, the head of sustainability, or the CFO. Buy pattern is annual SaaS licence indexed to revenue, employee count, or scope of value chain.

### 3.1 Watershed

- **Ownership:** Watershed Technology (US).
- **Deployment:** SaaS only. Multi-tenant. Cloud only.
- **Strengths:**
  - Industry leader in enterprise carbon accounting; CFO/compliance-led GTM.
  - Auditable data lineage marketed prominently (PwC, Deloitte attestation references).
  - CSRD/SEC/California SB 253-261 frameworks covered.
  - Decarbonisation-plan tooling in addition to accounting.
- **Weaknesses:**
  - Bills + ERP-only ingestion. No Modbus, no SCADA, no metered-data ingest.
  - No Italian Piano 5.0/Conto Termico/D.Lgs. 102/2014 coverage.
  - Closed source; black-box methodology.
  - Pricing: starting from approximately USD 50k/year for mid-market; multi-hundred-k for enterprise.
- **Where we beat:** OT ingestion, Italian regulation, deployment model (template/on-prem), price for SMEs.
- Source: [Sustainability Mag — Top 10 carbon accounting platforms](https://sustainabilitymag.com/top10/top-10-carbon-accounting-platforms-2026), [Watershed vs Sweep](https://www.sweep.net/watershed-vs-sweep).

### 3.2 Persefoni

- **Ownership:** Persefoni AI (US).
- **Deployment:** SaaS only.
- **Strengths:**
  - Financial-services and portfolio-emissions specialism (PCAF-aligned).
  - PersefoniGPT and AI-driven product carbon footprint tooling.
  - Methodology-pinned to assurance-grade audit standards.
- **Weaknesses:**
  - Industrial OT integrations are not the focus. Bills + ERP.
  - Italian regulation absent.
  - Closed source; SaaS-only.
- **Where we beat:** Industrial scope, OT ingestion, Italian compliance.
- Source: [Top carbon accounting platforms 2026](https://sustainabilitymag.com/top10/top-10-carbon-accounting-platforms-2026), [Persefoni Alternatives — Dcycle](https://dcycle.io/blog/persefoni-alternatives/).

### 3.3 Sweep

- **Ownership:** Sweep (FR).
- **Deployment:** SaaS only.
- **Strengths:**
  - Collaborative supplier-engagement workflows for Scope 3.
  - Aligned to multiple frameworks: CSRD, GRI, ISSB, TCFD, CDP, SFDR.
  - "Use carbon data once, report everywhere" pitch.
  - EU-domiciled (FR), GDPR-clean.
- **Weaknesses:**
  - No metered-data path. No OT ingestion.
  - Italian-specific regulations not native.
  - Pricing: enterprise-skewed; not SME-friendly.
- **Where we beat:** OT ingestion, Italian compliance, modular delivery.
- Source: [Sweep — best carbon accounting software](https://www.sweep.net/blog/top-carbon-accounting-software-for-us-businesses-in-2025), [Watershed vs Sweep](https://www.sweep.net/watershed-vs-sweep).

### 3.4 Greenly

- **Ownership:** Greenly (FR).
- **Deployment:** SaaS only.
- **Strengths:**
  - SME and mid-market focused — the closest competitor in our buying segment.
  - Transparent pricing, accessible UX.
  - "Massify" carbon accounting — playbook-driven onboarding.
- **Weaknesses:**
  - Bill/ERP ingest only.
  - Italian regulation: superficial. CSRD covered as a framework, Piano 5.0 not native.
  - Closed source; SaaS-only.
- **Where we beat:** OT ingestion, Italian regulatory depth, modular template, deterministic reproducibility.
- Source: [Greenly](https://greenly.earth/en-us/blog/company-guide/the-5-best-carbon-accounting-softwares-in-2022).

### 3.5 Plan A

- **Ownership:** Plan A Earth (DE).
- **Deployment:** SaaS only.
- **Strengths:**
  - EU-domiciled. CSRD-ready.
  - Strong DACH presence.
  - SBTi-aligned target setting.
- **Weaknesses:**
  - Same as Sweep/Greenly — no OT layer, Italian-regulation lite, closed source.
- **Where we beat:** OT layer, Italian regulation, modular template.
- Source: [Plan A — CSRD digital tagging](https://plana.earth/academy/csrd-digital-tagging).

### 3.6 Sphera (Sphera ESG / Life Cycle Assessment)

- **Ownership:** Sphera Solutions (US, BlackRock-backed).
- **Deployment:** SaaS + on-prem options.
- **Strengths:**
  - LCA gold standard — particularly for product carbon footprint and supply-chain emissions.
  - Carbon accounting + environmental compliance + supply-chain sustainability + operational risk.
  - Manufacturing/chemicals/energy industry depth.
  - GaBi LCA database — well known and widely cited.
- **Weaknesses:**
  - Heavyweight; expensive; complex to deploy.
  - Italian regulation as report template, not native.
  - No Piano 5.0 / Conto Termico / ENEA audit.
- **Where we beat:** Italian regulation, deployment time, price for SMEs.
- Source: [Top 10 EHS Sphera ESG alternatives — Dcycle](https://www.dcycle.io/post/sphera-esg-alternatives), [Verdantix EHS Software Benchmark](https://www.verdantix.com/venture/report/ehs-software-benchmark-environment-sustainability-management).

### 3.7 Cority — CorityOne

- **Ownership:** Cority Software (CA).
- **Deployment:** SaaS.
- **Strengths:**
  - EHS unified platform: environment, health, safety, quality, analytics.
  - Industrial hygiene depth; healthcare-vertical depth.
  - 2026 Applied AI strategy.
- **Weaknesses:**
  - EHS-first; energy/sustainability is one module among many.
  - No Italian regulation depth; CSRD coverage exists.
  - Closed source; SaaS only.
- **Where we beat:** Energy-and-sustainability depth, Italian regulation, OT ingestion, modular template.
- Source: [Top 10 Global EHSQ Platforms](https://sustainabilitymag.com/top10/top-10-global-eshq-platforms).

### 3.8 Wolters Kluwer Enablon

- **Ownership:** Wolters Kluwer (NL).
- **Deployment:** SaaS + on-prem.
- **Strengths:**
  - Often top-ranked for EHS+Risk+Sustainability for very large enterprise (>millions of data points).
  - Integrated Risk Management framework.
  - Long-standing Italian presence via Wolters Kluwer Tax & Accounting Italy.
- **Weaknesses:**
  - Enterprise-grade complexity; not friendly for SMEs.
  - No native Piano 5.0/Conto Termico/D.Lgs. 102/2014 modules.
  - Closed source.
- **Where we beat:** SME-fit pricing, Italian-regulation depth, deployment time, source visibility.
- Source: [Wolters Kluwer Enablon — 2026 award](https://www.wolterskluwer.com/en/news/wolters-kluwer-enablon-recognized-as-a-2026-environment-energy-leader-awards-winner).

### 3.9 Workiva

- **Ownership:** Workiva (US).
- **Deployment:** SaaS.
- **Strengths:**
  - ISG #1 for sustainability compliance (2025).
  - Connects financial + non-financial reporting in one platform.
  - First-class CSRD ESRS XBRL taxonomy support.
- **Weaknesses:**
  - Reporting layer, not data-collection layer. Pulls from your existing data systems; does not ingest meters.
  - No OT.
  - No Piano 5.0/Conto Termico/D.Lgs. 102/2014 native handling.
- **Where we beat:** Data ingestion (we are the source of truth), Italian regulation, deployment model.
- Note: Workiva is a complement, not a competitor, in many engagements — they consume our data.
- Source: [Workiva CSRD reporting](https://www.workiva.com/solutions/csrd-reporting), [ISG sustainability ratings 2025](https://www.stocktitan.net/news/III/sustainability-software-evolves-to-meet-changing-demands-isg-p6y8mk16wkp9.html).

### 3.10 Diligent ESG

- **Ownership:** Diligent Corporation (US, ex-BoardEffect).
- **Deployment:** SaaS.
- **Strengths:**
  - Board-and-governance lineage gives it a CFO/IR-friendly pitch.
  - ESG framework coverage.
- **Weaknesses:**
  - Same pattern: no OT, no Italian regulation, closed source.
- **Where we beat:** OT, Italian regulation, deployment model.

---

## 4. Tier 3 — Hyperscaler ESG Suites

Bundled with the underlying cloud platform; sold to existing customers of the hyperscaler. Adoption velocity is high, regulatory depth is shallow.

### 4.1 IBM Envizi ESG Suite

- **Ownership:** IBM (US).
- **Deployment:** SaaS atop IBM Cloud / hybrid.
- **Strengths:**
  - 1000+ data sources.
  - Utility-bill anomaly detection — catches mis-billing.
  - CSRD/SEC/CDP regulatory coverage.
  - AI-powered dashboards.
- **Weaknesses:**
  - Acquired layer (originally Envizi); integration with broader IBM stack still maturing.
  - No OT ingestion path.
  - Italian regulation lite.
- **Where we beat:** OT, Italian regulation, modular template, audit reproducibility (lineage).
- Source: [Compare IBM Envizi vs Salesforce Net Zero — PeerSpot](https://www.peerspot.com/products/comparisons/ibm-envizi-esg-suite_vs_salesforce-net-zero-cloud), [Top 10 IBM Envizi alternatives — PeerSpot](https://origin.peerspot.com/products/ibm-envizi-esg-suite-alternatives-and-competitors).

### 4.2 Salesforce Net Zero Cloud

- **Ownership:** Salesforce (US).
- **Deployment:** SaaS within Salesforce platform.
- **Strengths:**
  - Real-time CRM-linked carbon view.
  - Marketplace for carbon offsets.
  - Goal-setting + reduction tracking.
  - For Salesforce-native customers, integration is friction-free.
- **Weaknesses:**
  - Locked into Salesforce platform; no value if customer doesn't run Salesforce.
  - No OT ingestion.
  - Italian regulation lite.
- **Where we beat:** OT, Italian regulation, deployment independence, modular template.
- Source: [Microsoft Cloud for Sustainability vs Salesforce Net Zero — PeerSpot](https://www.peerspot.com/products/comparisons/microsoft-cloud-for-sustainability_vs_salesforce-net-zero-cloud).

### 4.3 Microsoft Sustainability Manager (Microsoft Cloud for Sustainability)

- **Ownership:** Microsoft (US).
- **Deployment:** SaaS within the Microsoft cloud (Dataverse, Fabric, Power Platform).
- **Strengths:**
  - Microsoft Fabric integration — real lineage tooling.
  - Power BI dashboarding.
  - Emissions Impact Dashboard.
- **Weaknesses:**
  - Locked to Microsoft cloud; Italian residency story complicated unless on a Local Region.
  - OT ingestion via Dataverse connectors only — not OT-native.
  - Italian regulation lite.
- **Where we beat:** OT-native, Italian residency story (Aruba/eu-south-1), Italian regulation.
- Source: [Microsoft Cloud for Sustainability vs Salesforce Net Zero](https://www.peerspot.com/products/comparisons/microsoft-cloud-for-sustainability_vs_salesforce-net-zero-cloud).

### 4.4 SAP Green Ledger + SAP Sustainability Footprint Management

- **Ownership:** SAP SE (DE).
- **Deployment:** SAP Business Technology Platform (BTP) atop SAP S/4HANA Cloud.
- **Strengths:**
  - ERP-linked: every business transaction can carry a carbon dimension.
  - Scope 1/2/3 across corporate, value chain, product.
  - Carbon allowances and liability handling — closer to financial accounting than the ESG-suites are.
- **Weaknesses:**
  - **Hard prerequisite: must be on S/4HANA Cloud + BTP.** This is a six-figure prerequisite.
  - Bill/ERP-data centric, not OT.
  - Italian regulation lite; SAP regional templates.
- **Where we beat:** OT ingestion, deployment independence (no S/4HANA prereq), Italian regulation.
- Source: [SAP Green Ledger overview](https://www.sap.com/products/financial-management/green-ledger.html), [SAP Sustainability Footprint Management](https://www.sap.com/products/scm/sustainability-footprint-management.html), [SAP Green Ledger features](https://learning.sap.com/learning-journeys/discovering-sustainability-for-sap-finance/explaining-key-features-of-sap-green-ledger).

### 4.5 Google Carbon Footprint

- **Ownership:** Google (US).
- **Deployment:** Free-of-charge module within Google Cloud console.
- **Strengths:** Native to Google Cloud; methodology published; scope-3 estimates by region.
- **Weaknesses:** Limited to GCP-hosted workloads. Useless as an industrial energy-management platform; only counts the cloud-compute component.
- **Where we beat:** Everything outside cloud-compute scope.

---

## 5. Tier 4 — ESG Ratings (Adjacent)

These are not direct competitors but they shape the buyer's mental model and procurement dynamics. The buyer often uses GreenMetrics' output to feed an ESG rating questionnaire.

### 5.1 EcoVadis

- **Ownership:** EcoVadis (FR).
- **Deployment:** Web portal (questionnaire-based assessment).
- **Strengths:** 220 industries, 180 countries; supply-chain-procurement standard.
- **Weaknesses:** Not a data platform; black-box rating methodology.
- **Note:** GreenMetrics output (CSRD ESRS E1 data) feeds EcoVadis questionnaires directly.
- Source: [EcoVadis](https://ecovadis.com/), [DQS — EcoVadis vs CDP](https://www.dqsglobal.com/en/explore/blog/ecovadis-vs-cdp-a-complete-guide-to-esg-ratings-and-climate-disclosure-systems).

### 5.2 CDP (Carbon Disclosure Project)

- **Ownership:** CDP non-profit (UK).
- **Deployment:** Annual disclosure web platform.
- **Strengths:** Investor and stakeholder reach; alignment with TCFD.
- **Weaknesses:** Not a measurement platform; the ESRS E1 dossier maps to CDP A-list scoring.

### 5.3 MSCI ESG / Sustainalytics (Morningstar)

- **Ownership:** MSCI (US) and Morningstar Sustainalytics (US/NL).
- **Deployment:** Subscription rating data via portal/API.
- **Strengths:** Investment-decision-grade ESG ratings used by capital allocators.
- **Weaknesses:** Closed methodology; reactive assessments.
- **Note:** From November 2026, only ESMA-authorised providers can offer ESG ratings in the EU — material for European regulatory uplift.
- Source: [Top 10 ESG Ratings Providers](https://sustainabilitymag.com/top10/top-10-esg-ratings-providers), [EU ESG Rating Regulation 2026 guide](https://www.getsunhat.com/blog/eu-esg-rating-regulation-ecovadis).

---

## 6. Tier 5 — Italian Utility / ESCO Bundled

Italian utilities offer energy-services bundles with monitoring as a layer of an energy contract. These are channel partners as often as competitors.

### 6.1 Enel X — Business Energy Manager

- **Ownership:** Enel SpA (IT).
- **Deployment:** Bundled with Enel energy supply.
- **Strengths:**
  - Customer reach: Enel is the largest Italian electricity supplier.
  - Direct Piano Transizione 5.0 advisory built into the Enel commercial offer.
  - Behind-the-meter PV/BESS comfort-management bundles.
- **Weaknesses:**
  - Conflict of interest on Piano 5.0 attestazione: Enel cannot certify savings on equipment Enel did not sell. Customers feel the lock-in.
  - Closed platform; behind a customer portal.
  - Sticky to Enel as supplier; switching cost is real.
- **Italy notes:** Strong in Lombardia and Lazio; weaker in Veneto where Hera/A2A compete.
- **Where we beat:** Vendor-neutral certification, no supplier lock-in, deeper engineering. Position Enel as a *partner channel* for installation, not as a competitor for the platform.
- Source: [Enel — Transizione 5.0](https://www.enel.it/en-us/imprese/bandi-incentivi/decreti-attuativi-transizione-5), [Enel — guide](https://www.enel.it/it-it/blog/guide/transizione-50-imprese-guida-credito-imposta).

### 6.2 A2A — Smart City + Smart Industry

- **Ownership:** A2A SpA (IT, listed).
- **Deployment:** Service-bundled with A2A energy contracts.
- **Strengths:** Lombardia/Veneto presence; municipal-utility relationships.
- **Weaknesses:** Same conflict-of-interest pattern; less software-investment than Enel; integrators rebuild reports manually.
- **Where we beat:** Software-first product, vendor neutrality, Piano 5.0 credibility.

### 6.3 Hera — Smart Services / HSE Decarbonisation Plan

- **Ownership:** Hera SpA (IT, listed).
- **Deployment:** Hera Servizi Energia (HSE) — services-led with monitoring overlay.
- **Strengths:** Emilia-Romagna presence; industrial-decarbonisation transition planning.
- **Weaknesses:** Service-led, not platform-led. Rebuild costs.
- **Where we beat:** Same as Enel/A2A pattern.

### 6.4 Edison Next / Edison EnergEnvision

- **Ownership:** Edison SpA (IT/FR via EDF).
- **Deployment:** Service-led.
- **Strengths:** Industrial track record; ESCO services; CHP/cogen depth.
- **Weaknesses:** No first-class platform; reports rebuilt per engagement.
- **Where we beat:** Platform-as-deliverable (we leave the customer with code and runbooks; Edison leaves the customer with a PDF).

### 6.5 Duferco Energia / SunCity

- **Ownership:** Italian energy traders/ESCOs.
- **Deployment:** Service-led with light dashboarding.
- **Strengths:** Trading optimisation expertise; commodity hedging.
- **Weaknesses:** Reporting capacity light; no platform offering.
- **Where we beat:** Platform offering, modular delivery.

### 6.6 MAPS Energy — BrainWatt®

- **Ownership:** MAPS Group (IT, Parma-based).
- **Deployment:** SaaS platform (BrainWatt) + Energy Desk 5.0 service.
- **Strengths:**
  - Italian-native: handles Transizione 5.0 advisory and ISO 50001.
  - Active in publishing FAQs/decreto interpretation — domain-credible.
  - Direct integration with GSE platform updates.
- **Weaknesses:**
  - Smaller scale than tier-1 incumbents.
  - Single-platform SaaS; not a modular template.
  - Limited published evidence on OT-protocol breadth.
- **Italy notes:** The closest Italian-domestic peer. Positioned as an "Energy Desk" service with software, not as an open template.
- **Where we beat:** OT breadth, modular-template offering, doctrine/transparency. We can co-exist as channel partners on certain engagements.
- Source: [MAPS Energy BrainWatt](https://energy.mapsgroup.it/en/home-en/), [MAPS Energy — Transizione 5.0 FAQs](https://energy.mapsgroup.it/transizione-cosa-prevedono-le-faq-del-mimit/).

---

## 7. Tier 6 — Specialist Energy Management SaaS

Smaller pure-plays focused on energy-monitoring with light reporting. The closest functional peers in the SME space.

### 7.1 Wattics

- **Ownership:** acquired by Engie Impact.
- **Deployment:** SaaS + Octopus Gateway for Modbus aggregation.
- **Strengths:**
  - Modbus RTU + TCP via Octopus Gateway with 30+ pre-loaded meter drivers.
  - Plug-and-play with existing meters; positive G2 reviews on simplicity.
- **Weaknesses:**
  - No Italian regulation depth.
  - Wattics name is fading post-Engie acquisition; product roadmap uncertain.
- **Where we beat:** Italian compliance, modular template, strategic clarity.
- Source: [Wattics — Modbus](https://www.wattics.com/doc/can-my-modbus-meters-connect-to-wattics/), [G2 Wattics reviews](https://www.g2.com/products/wattics/reviews).

### 7.2 DEXMA (Spacewell Energy by Nemetschek)

- **Ownership:** Nemetschek Group (DE) via Spacewell.
- **Deployment:** SaaS + DEXGate2 hardware.
- **Strengths:**
  - Hardware-neutral platform; Modbus integration via DEXGate2; user-creatable integrations.
  - Cloud-based EM software for ESCOs.
  - Automated reporting, budgeting, KPI tracking.
- **Weaknesses:**
  - Italian regulation absent.
  - Building/facility-management-tilt rather than industrial-process-tilt.
  - Licensing per meter per month; SaaS-only.
- **Where we beat:** Italian regulation depth, industrial process scope, modular delivery.
- Source: [DEXMA energy metering pain points](https://www.dexma.com/blog-en/solve-energy-metering/), [Spacewell vs Wattics — GetApp](https://www.getapp.com/industries-software/a/spacewell-energy-by-dexma/compare/wattics/).

### 7.3 Bidgely

- **Ownership:** Bidgely (US).
- **Deployment:** SaaS for utilities.
- **Strengths:** ML-disaggregation of residential consumption to appliance level.
- **Weaknesses:** Residential-focused, utility-buyer; not a fit for industrial SMEs.
- **Where we beat:** Industrial scope.

### 7.4 EnergyCAP

- **Ownership:** EnergyCAP LLC (US).
- **Deployment:** SaaS, cloud-native.
- **Strengths:**
  - 40-year history in utility-bill management and audit.
  - Automated bill ingestion + audit + analytics.
  - Public-sector and education-sector strength in the US.
- **Weaknesses:**
  - Bill-centric. Not OT-native.
  - North-American utility focus; weak EU regulatory coverage.
- **Where we beat:** OT, Italian regulation, deployment model.
- Source: [EnergyCAP](https://www.energycap.com/), [EnergyCAP Wikipedia](https://en.wikipedia.org/wiki/EnergyCAP).

### 7.5 Verdigris

- **Ownership:** Verdigris Technologies (US).
- **Deployment:** SaaS + proprietary current-clamp sensors.
- **Strengths:** AI-driven anomaly detection; predictive maintenance overlay on energy data.
- **Weaknesses:** Hardware-first; expensive deployment per panel; no Italian regulation.
- **Where we beat:** Vendor-neutral hardware support, regulatory depth, modular delivery.

### 7.6 GridPoint

- **Ownership:** GridPoint Inc. (US, Reston VA).
- **Deployment:** Hybrid (BMS + cloud); equipment-level submetering hardware.
- **Strengths:** Building-management-system integration; energy-performance-as-a-service for retail/hospitality.
- **Weaknesses:** Building/retail focus; not industrial; US-centric.
- **Where we beat:** Industrial process scope, EU/Italian regulation.
- Source: [GridPoint](https://www.gridpoint.com/), [GridPoint Wikipedia](https://en.wikipedia.org/wiki/GridPoint).

### 7.7 Carbon Lighthouse

- **Ownership:** Carbon Lighthouse (US, partial wind-down reported).
- **Deployment:** Hybrid (own sensors + cloud SaaS).
- **Strengths:** Performance-contract pricing model.
- **Weaknesses:** Hardware-heavy; long deployment cycles; financial uncertainty.
- **Where we beat:** Asset-light delivery, regulatory depth.

### 7.8 Ndustrial

- **Ownership:** Ndustrial (US).
- **Deployment:** SaaS + edge ingest.
- **Strengths:** Cold-chain and process-industry energy intensity (kWh / unit produced).
- **Weaknesses:** Vertical focus; US-centric; no Italian regulation.
- **Where we beat:** Italian regulation breadth, modular template.

### 7.9 Bractlet

- **Ownership:** Bractlet (US).
- **Deployment:** SaaS for commercial real estate.
- **Strengths:** Building-energy modelling overlay on consumption data.
- **Weaknesses:** CRE focus; weak in industrial.
- **Where we beat:** Industrial scope.

### 7.10 Tractian

- **Ownership:** Tractian (BR/US).
- **Deployment:** SaaS + IoT sensors + AI ("industrial copilot").
- **Strengths:**
  - Plug-and-play vibration + energy sensors.
  - 7× ROI claim; 43% reduction in unplanned downtime claim.
  - Strong UX; growing reviews on G2.
  - Energy Trac sensor + reports module.
- **Weaknesses:**
  - Maintenance/CMMS-first; energy/sustainability is secondary.
  - No CSRD ESRS E1 / Piano 5.0 / D.Lgs. 102 / Conto Termico.
  - Hardware-vendor lock pressure.
  - User reports note no equipment-customisation in reports.
- **Where we beat:** Sustainability/regulatory depth, vendor neutrality, modularity. Tractian is a complement, not a competitor, on the maintenance side.
- Source: [Research.com — Tractian review](https://research.com/software/reviews/tractian), [Tractian Energy Trac](https://tractian.com/en/solutions/oee/energy-trac), [G2 Tractian reviews](https://www.g2.com/products/tractian-tractian/reviews).

### 7.11 FogHorn (acquired by Johnson Controls / OpenBlue)

- **Ownership:** Johnson Controls (US).
- **Deployment:** Edge computing under Johnson Controls OpenBlue umbrella.
- **Strengths:** Edge ML inference; fog-computing patent portfolio.
- **Weaknesses:** Absorbed into JC OpenBlue; standalone product fading; building-vertical.
- **Where we beat:** Industrial-process focus, regulatory depth.

---

## 8. Tier 7 — Time-Series / Process Historians (Foundational)

These are not direct competitors — they are the data-tier alternative the buyer's IT team will propose. The right framing is "we use TimescaleDB inside our template and we'll connect to your existing historian".

### 8.1 AVEVA Historian (ex-Wonderware)

- **Ownership:** AVEVA Group (UK, Schneider-owned).
- **Deployment:** On-prem (or AVEVA Insight cloud).
- **Strengths:** Industry-standard historian; high-fidelity time-series storage; deep SCADA integration.
- **Weaknesses:** Closed; per-tag licensing; no ESG/CSRD/Piano 5.0 layer above.
- **Where we beat:** Reporting layer, regulatory depth, transparency. We can ingest from AVEVA Historian, not replace it.
- Source: [AVEVA Historian](https://www.aveva.com/en/products/historian/), [AVEVA Insight](https://www.aveva.com/en/products/insight/).

### 8.2 GE Vernova Proficy Historian

- **Ownership:** GE Vernova (US).
- **Deployment:** On-prem; high-speed.
- **Strengths:** High-volume industrial collection; A&E events; GE Predix interop.
- **Weaknesses:** Closed; expensive; no ESG layer.
- **Where we beat:** ESG/regulatory layer, openness.
- Source: [Proficy Historian](https://www.gevernova.com/software/products/proficy/historian).

### 8.3 OSIsoft PI System (now AVEVA PI)

- **Ownership:** AVEVA (UK).
- **Deployment:** On-prem-first.
- **Strengths:** Default historian for most large EU industrial sites; strong ecosystem.
- **Weaknesses:** Per-tag licensing; closed.
- **Where we beat:** Regulatory depth, openness, SME-fit pricing.

### 8.4 InfluxData (InfluxDB + Telegraf)

- **Ownership:** InfluxData Inc (US).
- **Deployment:** OSS core + cloud + enterprise.
- **Strengths:**
  - Fast ingestion.
  - Telegraf has 200+ input plugins.
  - Open-source core builds developer mindshare.
- **Weaknesses:**
  - InfluxDB ingestion can outperform TimescaleDB on raw rates, but TimescaleDB outperforms on complex queries with relational joins (relevant for our `meters` × `emission_factors` × `tenants` joinscape).
  - Flux query language is non-SQL — friction for analysts.
  - InfluxDB v3 has been rewritten in Rust on Apache Arrow, breaking continuity.
- **Why our Timescale choice stands:** Continuous aggregates + SQL + relational joins + transparent migration path. We document this in ADR-0010.
- Source: [TimescaleDB vs InfluxDB — Tiger Data](https://www.tigerdata.com/blog/timescaledb-vs-influxdb-for-time-series-data-timescale-influx-sql-nosql-36489299877), [TSBS IoT Performance](https://tdengine.com/tsbs-iot-performance-report-tdengine-influxdb-and-timescaledb/), [LavaPi — InfluxDB vs TimescaleDB IoT](https://www.lavapi.com/blog/influxdb-vs-timescaledb-iot-sensor-data).

### 8.5 ClickHouse

- **Ownership:** ClickHouse Inc (US).
- **Deployment:** OSS + cloud.
- **Strengths:** Analytical query speed; high compression.
- **Weaknesses:** Real-time updates harder; less SQL-relational than Postgres.
- **Where we beat:** End-to-end product, regulatory layer.

### 8.6 TDengine

- **Ownership:** TAOSData (CN).
- **Deployment:** OSS + cloud.
- **Strengths:** TSBS benchmarks 1–3× faster than TimescaleDB on raw IoT ingestion.
- **Weaknesses:** Smaller ecosystem; CN provenance can be a concern in some EU defence/industrial procurements.
- **Where we beat:** Product-level features, EU-residency story.

---

## 9. Tier 8 — Open-Source Reference Implementations (LF Energy + adjacent)

These are the spiritual cousins of a "modular template" positioning. They are complements, not competitors.

### 9.1 OpenSTEF (LF Energy)

- **Mission:** Automated ML pipelines for 48-hour grid-load forecasting.
- **Strengths:** Open source; scikit-learn compatible; production-tested by Alliander.
- **Weaknesses:** Single-purpose forecasting tool; no ESG/CSRD/Piano 5.0 layer.
- **How we relate:** Integrate as a Phase-4 forecasting backend behind our `Forecaster` interface.
- Source: [OpenSTEF GitHub](https://github.com/OpenSTEF), [LF Energy projects](https://lfenergy.org/our-projects/).

### 9.2 Hyphae (LF Energy)

- **Mission:** Peer-to-peer renewable energy distribution for microgrids (Sony CSL collaboration).
- **Strengths:** Microgrid-native; open source.
- **Weaknesses:** Microgrid niche, not energy-management.
- **How we relate:** Integrate as a flexibility-market connector for Phase-4 MSD aggregator partnerships.

### 9.3 CoMPAS / SEAPATH (LF Energy)

- **Mission:** IEC 61850 SCD configuration tooling and substation virtualisation.
- **Strengths:** Substation-vertical depth; aligned with Alliander/RTE.
- **Weaknesses:** Substation-only; not industrial.
- **How we relate:** Inspiration for the `IngestorRunner` extension-point pattern; reference for IEC 61850 mapping when we add it.

### 9.4 Frinx / Linux Foundation Energy ecosystem at large

- **Mission:** Open-source orchestration for energy infrastructure.
- **Strengths:** Open posture; community.
- **Weaknesses:** Fragmented; reference-quality, not production-grade for commercial SME engagements.
- **How we relate:** Cite in our ADR-0001 doctrine adoption to position the modular template as part of the wider open-source-driven energy transformation.
- Source: [LF Energy — open source sustainability research](https://lfenergy.org/lf-energy-research-finds-open-source-software-is-driving-sustainability-innovation-including-climate-technologies-environmental-science-and-energy-efficiency/).

---

## 10. Tier 9 — DIY Frankenstack

The most common alternative we will face is "we already have Grafana, we'll just use that".

### 10.1 Grafana + Postgres + Telegraf + custom ingestor

- **Footprint:** Most Italian industrial energy managers we will encounter in the field have a Grafana dashboard built by an integrator on top of a Postgres or InfluxDB instance, with a hand-rolled Modbus polling script in Python or a small Go service.
- **Strengths:** Cheap; familiar tools; no licence cost.
- **Weaknesses:**
  - No regulatory output layer (CSRD, Piano 5.0, D.Lgs. 102/2014).
  - No determinism — factor versioning, retention, replay are reinvented per integrator with varying quality.
  - No supply-chain hygiene; no signed images, no SBOMs.
  - No on-call runbooks; no SLOs.
  - One-developer dependency risk.
- **Where we beat:** Everything. The only argument is cost — and our modular template is forkable, so cost asymmetry is not the structural moat the customer assumes.
- **Strategic implication:** Many engagements start as "we already have X, can you bolt on the regulatory module?" Our template *can* run on top of the customer's existing Grafana/Postgres if they insist; we accept it as a degraded mode and recover the operational hygiene over the engagement.

### 10.2 Grafana Cloud + Mimir + Loki + Tempo + custom

- **Footprint:** Some larger plants run a managed Grafana Cloud subscription.
- **Strengths:** Managed-as-a-service; LGPL stack.
- **Weaknesses:** Same as 10.1; nothing in the ESG/regulatory layer; pricing scales with cardinality.
- **Where we beat:** Same.

---

## 11. Cross-Cutting Capability Matrix

The table below collapses the field along the seven capability axes that matter most for an Italian energy-intensive SME.

| Vendor | OT-native ingest | Italian regulation depth | Audit reproducibility (factor pinning) | Modular delivery / on-prem | CSRD ESRS E1 XBRL output | Piano 5.0 attestazione | Open source / source-available | SME pricing fit |
|--------|-----------------|--------------------------|----------------------------------------|----------------------------|---------------------------|-------------------------|--------------------------------|------------------|
| Schneider EcoStruxure RA | ✓✓ (own meters) | ✗ (template) | ✗ | ✗ (per-site licence) | ✓ (via Workiva-style export) | ✗ | ✗ | ✗ |
| Siemens Sinalytics + SIMATIC EM | ✓✓ (PROFINET, IEC 61850) | ✗ | ✗ | ✗ | ✓ | ✗ | ✗ | ✗ |
| ABB Ability Energy Manager | ✓✓ (own meters) | ✗ | ✗ | partial | partial | ✗ | ✗ | ✗ |
| Honeywell Forge | ✓ (BMS) | ✗ | ✗ | ✗ | ✓ | ✗ | ✗ | ✗ |
| Emerson Ovation Green | ✓✓ (renewables) | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| Rockwell FactoryTalk EM | ✓✓ (own PLCs) | ✗ | ✗ | ✗ | ✓ | ✗ | ✗ | ✗ |
| Watershed | ✗ | ✗ | partial | ✗ | ✓ | ✗ | ✗ | ✗ |
| Persefoni | ✗ | ✗ | partial | ✗ | ✓ | ✗ | ✗ | ✗ |
| Sweep | ✗ | ✗ | partial | ✗ | ✓ | ✗ | ✗ | ✗ |
| Greenly | ✗ | ✗ | partial | ✗ | ✓ | ✗ | ✗ | ✓ |
| Plan A | ✗ | ✗ | partial | ✗ | ✓ | ✗ | ✗ | ✓ |
| Sphera | partial | ✗ | partial | ✓ (on-prem) | ✓ | ✗ | ✗ | ✗ |
| Cority | ✗ | ✗ | ✗ | ✓ (on-prem) | partial | ✗ | ✗ | ✗ |
| Enablon | ✗ | ✗ | partial | ✓ (on-prem) | ✓ | ✗ | ✗ | ✗ |
| Workiva | ✗ | ✗ | ✓ (XBRL chain) | ✗ | ✓✓ | ✗ | ✗ | ✗ |
| IBM Envizi | ✗ | ✗ | partial | ✗ | ✓ | ✗ | ✗ | ✗ |
| Salesforce Net Zero | ✗ | ✗ | ✗ | ✗ | ✓ | ✗ | ✗ | ✗ |
| Microsoft Sustainability Mgr | ✗ | ✗ | ✓ (Fabric lineage) | ✗ | ✓ | ✗ | ✗ | ✗ |
| SAP Green Ledger + SFM | ✗ (ERP only) | ✗ | ✓ (ledger lineage) | ✗ | ✓ | ✗ | ✗ | ✗ |
| Enel X / A2A / Hera / Edison | partial | ✓ (services) | ✗ | ✗ | ✓ (services) | ✓ (with COI) | ✗ | partial |
| MAPS Energy BrainWatt | partial | ✓ | partial | ✗ | ✓ | ✓ | ✗ | partial |
| Wattics | ✓ (Modbus) | ✗ | ✗ | ✗ | partial | ✗ | ✗ | ✓ |
| DEXMA / Spacewell | ✓ (Modbus) | ✗ | ✗ | ✗ | partial | ✗ | ✗ | ✓ |
| Tractian | ✓ (own sensors) | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| EnergyCAP | ✗ (bills) | ✗ | partial | ✗ | partial | ✗ | ✗ | partial |
| AVEVA Historian | ✓✓ | ✗ | partial | ✓ | ✗ | ✗ | ✗ | ✗ |
| GE Proficy Historian | ✓✓ | ✗ | partial | ✓ | ✗ | ✗ | ✗ | ✗ |
| InfluxDB stack | ✓ (Telegraf) | ✗ | ✗ | ✓ | ✗ | ✗ | ✓ | ✓ |
| OpenSTEF / LF Energy | partial | ✗ | partial | ✓ | ✗ | ✗ | ✓ | ✓ |
| DIY Grafana frankenstack | ✓ (DIY) | ✗ | ✗ | ✓ | ✗ | ✗ | ✓ | ✓ |
| **GreenMetrics (target)** | **✓✓ (Modbus/M-Bus/SunSpec/OCPP/IEC 61850/MQTT/DSO portals)** | **✓✓ (Piano 5.0/Conto Termico/CSRD/D.Lgs. 102/Certificati Bianchi)** | **✓✓ (factor pinning + signed report bundle)** | **✓✓ (forkable template + docker compose + Helm + ArgoCD)** | **✓✓ (XBRL ESRS E1 + JSON + PDF)** | **✓✓ (deterministic + EGE counter-signature)** | **✓ (source-available, proprietary licence)** | **✓✓ (engagement pricing, no per-meter SaaS)** |

(✓✓ = first-class strength, ✓ = supported, partial = present but limited, ✗ = absent)

---

## 12. Where We Beat Each Competitor — Concrete Beat-Points

For each competitor, the single strongest argument we make in a sales/engagement conversation. These are crystallised into the doctrine and the demo flow.

| Competitor | One-line beat-point |
|------------|---------------------|
| Schneider EcoStruxure RA | "Our Piano 5.0 attestazione runs deterministically against the factor version valid at filing time — yours rebuilds against current factors, which is an ENEA finding waiting to happen." |
| Siemens SIMATIC EM | "We commission in 5 minutes from a clean clone. Yours is a 6-12 month TIA Portal integration." |
| ABB Ability EM | "Our ingestor speaks Modbus, M-Bus, SunSpec, OCPP, IEC 61850, MQTT Sparkplug B without favouring any vendor's panel." |
| Honeywell Forge | "We are industrial-process-native — compressed air, steam, chilled water — not just BMS." |
| Emerson Ovation Green | "We carry the full ESRS E1 / Piano 5.0 / D.Lgs. 102 reporting layer; you stop at SCADA." |
| Rockwell FactoryTalk EM | "We are vendor-neutral; you require Allen-Bradley PLCs to be first-class." |
| Watershed | "We ingest meters end-to-end at 15-minute granularity; you ingest utility bills and ERP allocations." |
| Persefoni | "We address industrial sites, not financial portfolios." |
| Sweep | "We ingest physical meters; your data layer stops at the procurement boundary." |
| Greenly | "We deliver Piano 5.0 attestazione natively; you deliver CSRD frameworks only." |
| Plan A | "Italian regulatory regimes (Piano 5.0, D.Lgs. 102, Conto Termico, TEE) are first-class in our model." |
| Sphera | "We commission in 5 minutes and deploy on the customer's infrastructure; you deploy in 6 months and price for Fortune-500 buyers." |
| Cority | "We are sustainability-and-energy-first, not EHS-first with sustainability as a module." |
| Enablon | "We fit an Italian SME's economics; you fit a global EHS programme of record." |
| Workiva | "We are the source of the data Workiva consumes — we sit upstream." |
| IBM Envizi | "We OT-ingest; you stop at utility bills, even if you ingest 1000 sources." |
| Salesforce Net Zero | "We work without a Salesforce CRM dependency." |
| Microsoft Sustainability Manager | "We work without a Microsoft cloud dependency, on Italian-residency infrastructure." |
| SAP Green Ledger | "We work without an S/4HANA + BTP prerequisite." |
| Google Carbon Footprint | "We measure your factory, not your cloud bill." |
| Enel X / A2A / Hera / Edison | "We co-sign Piano 5.0 attestazioni for energy-saving investments where the asset was *not* sold by the energy supplier — no conflict of interest." |
| MAPS Energy BrainWatt | "We ship as a forkable template — your engineering team owns the codebase, the migrations, the runbooks, the supply chain." |
| Wattics | "Italian regulation is a first-class concern, and our roadmap has a single accountable owner." |
| DEXMA / Spacewell | "We address process industry plus regulatory output; you address facility-management plus dashboards." |
| Tractian | "Sustainability-and-regulatory-grade reporting is first-class for us; you are CMMS-and-asset-health-first." |
| EnergyCAP | "We meter the plant; you read the bill." |
| AVEVA Historian / GE Proficy / OSIsoft PI | "We are the regulatory-and-sustainability layer above your historian; we don't replace it, we exploit it." |
| InfluxDB stack | "We ship a complete product, not a database. We use TimescaleDB because the relational join space (meters × tenants × factors) needs SQL." |
| OpenSTEF / LF Energy | "We are productisable for a commercial engagement and we carry Italian regulatory output natively; you are reference quality and infrastructure-shaped." |
| DIY Grafana frankenstack | "We carry signed images, attested provenance, runbooks per failure mode, ADRs, and 200+ rule doctrine. You carry one developer's working memory." |

---

## 13. Strategic Implications for the Modular-Template Positioning

1. **Channel partner mapping.** Enel X, A2A, Hera, Edison Next, MAPS Energy and the regional ESCO/EGE network are not competitors in the long run — they are the channel that sells our template into customers, while we sit as the modular product underneath. The doctrine should treat them as partners and design the plug-in surfaces to host their verticals (Conto Termico flow, ISO 50001, ESCO performance contracts).
2. **Direct competitors.** Schneider/Siemens/ABB/Honeywell/Rockwell will compete with us on enterprise sites; we de-risk by living *under* their hardware (we ingest from their meters), rather than replacing their hardware. The competitive zone is the *software layer above the meter*. They control the meter; we control the regulation, the reproducibility, and the engineering substrate.
3. **Shadow competitors.** Watershed/Persefoni/Sweep/Greenly/Plan A — the carbon-accounting SaaS class — are what the buyer benchmarks our pricing against. Our Piano-5.0-tax-credit-asymmetry argument moves the conversation off "is this €249/m/month worth it?" onto "what happens if we miss the 5.0 deadline?". This is where the modular template wins: the deliverable includes the credit-unlocking attestazione.
4. **Frankenstack defence.** The most common buyer alternative is "we already have Grafana; build us a regulatory module on top." We accept this as a degraded operating mode (the template can run atop existing Postgres/Grafana) while still attesting the regulatory output deterministically. This buys engagements that would otherwise be lost.
5. **Modular template moat.** No competitor ships a forkable, audited, signed template that a customer's engineering team can own. That is the structural moat: the deliverable *is* the asset, and the asset compounds across engagements via shared upstream improvements that customers can pull. None of the SaaS competitors can match this without abandoning their core business model.
6. **Doctrine as evidence.** The 200+ rule doctrine, ADRs, and supply-chain attestations are themselves marketing. CSRD auditors, NIS2 reviewers, and procurement reviewers all read evidence packs. We ship one. None of the competitors do.

---

## 14. Capabilities to Build to Beat the Field

Synthesising the gaps, the engineering team must build (or harden) the following capabilities in priority order. Each maps to a doctrine rule (referenced symbolically `RULE:n` to be slotted into the 200+ rule doctrine).

1. **Audit-grade reproducibility.** `RULE:Reports` — every report bundle stores its input snapshot (raw readings query result, factor version IDs, computation parameters) cryptographically hashed. Regenerating from the snapshot reproduces the report bit-for-bit. ENEA / CSRD auditors can re-derive offline.
2. **OT-native ingestor catalogue.** `RULE:OT-Coverage` — Modbus RTU, Modbus TCP, M-Bus (wired and wireless), SunSpec (PV inverter), OCPP 1.6 + 2.0.1 (EV charging), IEC 61850 (substation), MQTT Sparkplug B (industrial IoT), BACnet (BMS), EtherNet/IP via Allen-Bradley dictionaries, OPC-UA (industrial gateway), Italian DSO portals (E-Distribuzione SMD, Terna Transparency, SPD multi-DSO).
3. **Italian regulatory engine.** `RULE:Italian-Reg` — ESRS E1 XBRL tagging, Piano 5.0 attestazione XML/PDF aligned to MIMIT/GSE schemas, Conto Termico 3.0 dossier per GSE portal expectations, Certificati Bianchi TEE submission per GSE, D.Lgs. 102/2014 audit dossier per ENEA template, ISPRA factor versioning with valid_from semantics.
4. **Cryptographic provenance.** `RULE:Provenance` — every reading carries an HMAC chain (or COSE/JWS signature where the meter supports it); the chain is verifiable end-to-end; tamper events generate alerts.
5. **Modular template integrity.** `RULE:Template` — clean separation of core (template-invariant) and customisation (client-specific) via documented extension points; client engagements branch from template at a well-known commit; upstream sync mechanism is automatic and tested.
6. **Engineering doctrine as deliverable.** `RULE:Doctrine` — 200+ rule doctrine, 30+ ADRs, supply chain (Cosign + SLSA L3 + SBOM + Trivy + osv-scanner), 25+ runbooks, comprehensive risk register. All shipped with the template.
7. **Determinism in the calculator.** `RULE:Determinism` — pure-functional report generators; no implicit dependencies on system time or current factor sets; reproducibility verified in CI.
8. **Multi-jurisdiction extension points.** `RULE:Jurisdiction` — Piano 5.0 is Italian; analogous regimes exist (DE EnEfG, FR Décret tertiaire, ES RD 56/2016). The extension-point catalogue allows adding a jurisdiction without touching the core.
9. **Operational excellence at engagement boundary.** `RULE:Handover` — every engagement exits with a documented handover package (runbooks, on-call map, secret-rotation calendar, restore drill log).
10. **Performance baseline.** `RULE:Perf` — P95 < 100 ms ingest, P95 < 250 ms dashboard, P99 < 2.5 s report generation on 1-year window. Verified in continuous benchmarks.

---

## 15. Sources

Primary sources (selection):

- [Schneider EcoStruxure Resource Advisor](https://www.se.com/us/en/work/services/se-advisory-services/intelligent-software/resource-advisor/)
- [Schneider EcoStruxure Resource Advisor Copilot blog](https://perspectives.se.com/blog-stream/building-sustainabilitys-digital-future-with-ecostruxure-resource-advisor-copilot)
- [Net Zero Compare — Schneider EcoStruxure Resource Advisor](https://netzerocompare.com/software/schneider-electric-ecostruxure-resource-advisor)
- [TrustRadius — EcoStruxure pricing](https://www.trustradius.com/products/schneider-electric-ecostruxure/pricing)
- [Siemens Sinalytics press release](https://www.siemens.com/press/PR2016040260PSEN)
- [SIMATIC Energy Manager](https://www.siemens.com/en-us/products/simatic-energy-management/energy-manager/)
- [ABB Ability Energy Manager](https://electrification.us.abb.com/products/energy-management-systems/abb-ability-energy-manager)
- [ABB Ability OPTIMAX](https://www.abb.com/global/en/areas/automation/solutions/industrial-software/energy-management/energy-optimization-optimax)
- [Honeywell Forge Sustainability+ for Buildings — Carbon and Energy Management](https://buildings.honeywell.com/us/en/solutions/buildings/honeywell-forge-sustainability-plus-for-buildings-carbon-and-energy-management)
- [Honeywell Forge — Carbon & Energy Management brochure](https://buildings.honeywell.com/content/dam/hbtbt/en/documents/downloads/Honeywell-Forge-Sustainability+-for-Buildings-Carbon-and-Energy-Brochure-02082024.pdf)
- [Emerson Ovation Green](https://www.emerson.com/en-us/automation/ovation-green)
- [Rockwell FactoryTalk Energy Manager](https://www.rockwellautomation.com/en-us/products/software/factorytalk/innovationsuite/factorytalk-energy-manager.html)
- [Rockwell — Sustainability through Smart Manufacturing](https://www.rockwellautomation.com/en-us/company/news/press-releases/Rockwell-Automation-Advances-Sustainability-Through-Smart-Manufacturing.html)
- [Sustainability Magazine — Top 10 Carbon Accounting Platforms 2026](https://sustainabilitymag.com/top10/top-10-carbon-accounting-platforms-2026)
- [Sustainability Magazine — Top 10 ESG Ratings Providers](https://sustainabilitymag.com/top10/top-10-esg-ratings-providers)
- [Sustainability Magazine — Top 10 Global EHSQ Platforms](https://sustainabilitymag.com/top10/top-10-global-eshq-platforms)
- [Workiva — CSRD Reporting](https://www.workiva.com/solutions/csrd-reporting)
- [Workiva — ESRS XBRL Taxonomy](https://support.workiva.com/hc/en-us/articles/22929364571796-Use-the-ESRS-XBRL-taxonomy-for-sustainability-reporting)
- [Wolters Kluwer Enablon 2026 Award](https://www.wolterskluwer.com/en/news/wolters-kluwer-enablon-recognized-as-a-2026-environment-energy-leader-awards-winner)
- [Sweep — best carbon accounting software](https://www.sweep.net/blog/top-carbon-accounting-software-for-us-businesses-in-2025)
- [Watershed vs Sweep](https://www.sweep.net/watershed-vs-sweep)
- [Persefoni Alternatives — Dcycle](https://dcycle.io/blog/persefoni-alternatives/)
- [Greenly — best carbon accounting software](https://greenly.earth/en-us/blog/company-guide/the-5-best-carbon-accounting-softwares-in-2022)
- [Sphera ESG Alternatives — Dcycle](https://www.dcycle.io/post/sphera-esg-alternatives)
- [Verdantix — EHS Software Benchmark](https://www.verdantix.com/venture/report/ehs-software-benchmark-environment-sustainability-management)
- [PeerSpot — IBM Envizi vs Salesforce Net Zero](https://www.peerspot.com/products/comparisons/ibm-envizi-esg-suite_vs_salesforce-net-zero-cloud)
- [PeerSpot — Microsoft Cloud for Sustainability vs Salesforce Net Zero](https://www.peerspot.com/products/comparisons/microsoft-cloud-for-sustainability_vs_salesforce-net-zero-cloud)
- [PeerSpot — IBM Envizi vs Microsoft Cloud for Sustainability](https://www.peerspot.com/products/comparisons/ibm-envizi-esg-suite_vs_microsoft-cloud-for-sustainability)
- [PeerSpot — IBM Envizi alternatives](https://origin.peerspot.com/products/ibm-envizi-esg-suite-alternatives-and-competitors)
- [SAP — Green Ledger](https://www.sap.com/products/financial-management/green-ledger.html)
- [SAP — Sustainability Footprint Management](https://www.sap.com/products/scm/sustainability-footprint-management.html)
- [SAP Green Ledger features (SAP Learning)](https://learning.sap.com/learning-journeys/discovering-sustainability-for-sap-finance/explaining-key-features-of-sap-green-ledger)
- [Enel — Transizione 5.0 implementing decrees](https://www.enel.it/en-us/imprese/bandi-incentivi/decreti-attuativi-transizione-5)
- [Enel — Transizione 5.0 guide](https://www.enel.it/it-it/blog/guide/transizione-50-imprese-guida-credito-imposta)
- [MIMIT — Piano Transizione 5.0 avvisi](https://www.mimit.gov.it/it/normativa/notifiche-e-avvisi/avviso-29-aprile-2026-piano-transizione-5-0-tecnicamente-ammissibili)
- [MAPS Energy — BrainWatt](https://energy.mapsgroup.it/en/home-en/)
- [MAPS Energy — Transizione 5.0 FAQ](https://energy.mapsgroup.it/transizione-cosa-prevedono-le-faq-del-mimit/)
- [GSE — Certificati Bianchi documenti](https://www.gse.it/servizi-per-te/efficienza-energetica/certificati-bianchi/documenti)
- [MASE — Certificati Bianchi](https://www.mase.gov.it/portale/certificati-bianchi)
- [Wattics — Modbus integration](https://www.wattics.com/doc/can-my-modbus-meters-connect-to-wattics/)
- [G2 — Wattics reviews](https://www.g2.com/products/wattics/reviews)
- [Spacewell Energy DEXMA — pain points](https://www.dexma.com/blog-en/solve-energy-metering/)
- [GetApp — Spacewell Energy vs Wattics](https://www.getapp.com/industries-software/a/spacewell-energy-by-dexma/compare/wattics/)
- [EnergyCAP](https://www.energycap.com/)
- [EnergyCAP — Wikipedia](https://en.wikipedia.org/wiki/EnergyCAP)
- [GridPoint](https://www.gridpoint.com/)
- [GridPoint — Wikipedia](https://en.wikipedia.org/wiki/GridPoint)
- [Tractian — Energy Trac sensor](https://tractian.com/en/solutions/oee/energy-trac)
- [Research.com — Tractian review 2026](https://research.com/software/reviews/tractian)
- [G2 — Tractian reviews](https://www.g2.com/products/tractian-tractian/reviews)
- [F7i.ai — Tractian 2026 evaluation](https://f7i.ai/blog/tractian-the-industrial-copilot-redefining-predictive-maintenance-in-2026)
- [AVEVA Historian](https://www.aveva.com/en/products/historian/)
- [AVEVA Insight](https://www.aveva.com/en/products/insight/)
- [GE Vernova Proficy Historian](https://www.gevernova.com/software/products/proficy/historian)
- [SourceForge — AVEVA PI System alternatives](https://sourceforge.net/software/product/AVEVA-PI-System/alternatives)
- [TimescaleDB vs InfluxDB — Tiger Data](https://www.tigerdata.com/blog/timescaledb-vs-influxdb-for-time-series-data-timescale-influx-sql-nosql-36489299877)
- [TSBS IoT Performance Report — TDengine vs InfluxDB vs TimescaleDB](https://tdengine.com/tsbs-iot-performance-report-tdengine-influxdb-and-timescaledb/)
- [LavaPi — InfluxDB vs TimescaleDB IoT](https://www.lavapi.com/blog/influxdb-vs-timescaledb-iot-sensor-data)
- [LF Energy — projects](https://lfenergy.org/our-projects/)
- [LF Energy — open source sustainability research](https://lfenergy.org/lf-energy-research-finds-open-source-software-is-driving-sustainability-innovation-including-climate-technologies-environmental-science-and-energy-efficiency/)
- [OpenSTEF on GitHub](https://github.com/OpenSTEF)
- [LF Energy on GitHub](https://github.com/lf-energy)
- [LF Energy — Wikipedia](https://en.wikipedia.org/wiki/LF_Energy)
- [EFRAG — Digital Reporting with XBRL](https://www.efrag.org/en/sustainability-reporting/esrs-workstreams/digital-reporting-with-xbrl)
- [Plan A — CSRD digital tagging](https://plana.earth/academy/csrd-digital-tagging)
- [Senken — CSRD reporting requirements](https://www.senken.io/academy/csrd-reporting-requirements)
- [IRIS Carbon — 5 Things about CSRD/ESRS](https://iriscarbon.com/5-things-you-need-to-know-about-csrd-reporting-using-esrs/)
- [European Commission — Corporate Sustainability Reporting](https://finance.ec.europa.eu/financial-markets/company-reporting-and-auditing/company-reporting/corporate-sustainability-reporting_en)
- [XBRL.org — EFRAG portal & ESRS disclosures](https://www.xbrl.org/news/efrag-portal-offers-glimpse-into-esrs-disclosures/)
- [DQS — EcoVadis vs CDP](https://www.dqsglobal.com/en/explore/blog/ecovadis-vs-cdp-a-complete-guide-to-esg-ratings-and-climate-disclosure-systems)
- [EcoVadis — methodology](https://resources.ecovadis.com/whitepapers/ecovadis-ratings-methodology-overview-and-principles-2022-neutral)
- [Sunhat — EU ESG Rating Regulation 2026 guide](https://www.getsunhat.com/blog/eu-esg-rating-regulation-ecovadis)
- [Hera Servizi Energia & Italian utility energy transition (Il Giornale d'Italia)](https://www.ilgiornaleditalia.it/news/mondo-imprese/772852/hera-enel-x-siemens-a2a-riconfermata-presenza-alla-fiera-key-the-energy-transistion-expo-focus-sui-nuovi-modelli-di-decarbonizzazione-.html)
- [GSE — Conto Termico 3.0 / BibLus](https://biblus.acca.it/documenti-gse-attestazioni-e-certificazioni-per-il-conto-termico-3-0/)
- [GSE — Transizione 5.0 piattaforma operativa](https://energy.mapsgroup.it/transizione-5-0-operativa-la-piattaforma-del-gse/)
- [Open Group Italia — Transizione 5.0 FAQ del GSE](https://www.opengroupitalia.it/news/faq-transizione-5-0/)
- [MAPS Energy — ISO 50001](https://energy.mapsgroup.it/sistema-di-monitoraggio-energetico-per-aziende-iso-50001/)
- [Sweep — best carbon accounting software](https://www.sweep.net/blog/top-carbon-accounting-software-for-us-businesses-in-2025)

---

*End of competitive brief. Total length ≈ 700 lines. Next step: integrate the beat-points into the 200+ rule doctrine and the 10000-line uplift plan.*
