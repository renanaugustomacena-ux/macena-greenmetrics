# MODUS OPERANDI — GreenMetrics v2

> Engagement playbook for the GreenMetrics modular template. Strategic, technical, and operational.
> **Supersedes v1 (the SaaS playbook) per ADR-0028 on 2026-04-30.**
> Reads alongside `docs/MODULAR-TEMPLATE-CHARTER.md`, `docs/DOCTRINE.md`, `docs/PLAN.md`, `docs/COMPETITIVE-BRIEF.md`.

---

## 0. What changed since v1

If you read MODUS_OPERANDI v1 and absorbed €99/€249 pricing tiers, CAC funnels, LTV/churn metrics, "multi-tenant SaaS" framing, and a 2-person-direct-sales-org targeting €5.4M ARR by Y3 — that document is retired. The pivot recorded in `docs/MODULAR-TEMPLATE-CHARTER.md` reframes the project as a *modular template* delivered as *engagements*, not a hosted SaaS product.

The four most important deltas:

1. **No multi-tenant SaaS.** Every engagement is a single-tenant deployment by default (multi-tenant by configuration only when a partner ESCO or system-integrator wants to host multiple clients on one substrate). The threat model, the RBAC matrix, the RLS policies — all preserved as defence in depth, but the tenant primitive is engineering hygiene, not a billing primitive.

2. **No per-meter pricing.** The unit of sale is the engagement: a contracted commercial relationship covering license + customisation services + annual maintenance + (optionally) co-managed or fully-managed retainer. A deployment with 5 meters and a deployment with 5,000 meters cost the same on the license line; the engagement margin scales with engagement complexity, not with meter count.

3. **No churn / CAC / LTV.** The KPIs are engagement margin, time-to-customisation, template-fit-score, net-engagement-value, annual-maintenance-attach-rate. Engagements are won or lost end-to-end, not as cohort attrition.

4. **No self-serve.** Industrial customers don't self-serve. The engagement starts with a discovery conversation, runs through Phase 0–5 of the lifecycle, and ends in operator-readiness handover (or co-managed / fully-managed continuation).

The Italian market analysis from v1 is preserved unchanged in §2 (the underlying TAM / SAM / SOM is about the market, not our delivery model). The technical-architecture overview is preserved with light annotation pointing at the Pack extraction (Sprint S6–S7 of `docs/PLAN.md`).

---

## 1. Executive Summary

GreenMetrics is **the regulator-grade modular template for industrial energy management and sustainability reporting**. We deliver it to clients as an engagement: the contracted client receives the template (source, infrastructure, doctrine, runbooks, audit evidence), the engagement specialises it via Pack composition + a customisation sprint, and the deployment is operated by the client (T1 Handover), co-managed (T2), or fully-managed by Macena (T3) depending on tier.

The buyer profile is the energy-intensive Italian industrial SME (later: European industrial SME) that needs CSRD ESRS E1, Piano Transizione 5.0 attestazione, Conto Termico 2.0 GSE submission, Certificati Bianchi TEE, and D.Lgs. 102/2014 audit energetico capability and is not adequately served by:

- the Tier 1 enterprise OT/automation platforms (Schneider EcoStruxure / Siemens Sinergy / ABB Ability / Honeywell Forge / Emerson Ovation Green / Rockwell FactoryTalk Energy) that price at €50k–€300k per site per year and require six-month integrations;
- the ESG SaaS platforms (Watershed / Persefoni / Sweep / Greenly / Plan A / Sphera / Cority / Wolters Kluwer Enablon / Workiva) that lack OT-native ingestion and Italian-regulation depth;
- the hyperscaler ESG suites (IBM Envizi / Salesforce Net Zero / Microsoft Sustainability Manager / SAP Green Ledger / Google Carbon Footprint) that ship generic, vendor-locked, US-cloud-first surfaces;
- the Italian utility-bundled / ESCO platforms (Enel X / A2A / Hera / Edison Next / Duferco Energia / SunCity) that face structural conflicts of interest on Piano 5.0 attestazione for investments they did not sell.

The wedge is regulator-grade engineering. We make claims (bit-perfect reproducibility, signed audit chain, formal-spec validation against EFRAG XBRL / GSE XSD / ENEA XSD, signed reading provenance with per-meter HMAC, Pack-versioned thresholds, EGE counter-signature workflow, single-command audit-evidence-pack export) that no surveyed competitor makes (per `docs/COMPETITIVE-BRIEF.md`) and cannot quickly acquire. The 210-rule doctrine in `docs/DOCTRINE.md` makes those claims auditable and the codebase reviewable.

The commercial mechanics: each engagement is sized in the €100k–€500k band combining license + customisation + first-year maintenance, with annual maintenance attach at ≥ 90% of T1 closures and an optional co-managed retainer at €4k–€12k/month or fully-managed retainer at €18k–€55k/month for clients who want operations-as-a-service. The Phase E–J plan (`docs/PLAN.md`) targets *five engagements through Phase 5 Handover* by end of Phase J Sprint S22 (≈ March 2027). Engagement portfolio expansion happens engagement-by-engagement, with the Italian flagship Pack + 5 additional Region Packs + 12 Protocol Packs + 6 Factor Packs + 8 Report Packs + 3 Identity Packs in the catalogue by Phase H exit.

The remainder of this document lays out the market analysis, the engagement model, the technical architecture (post Pack extraction), the engagement lifecycle, scaling, security, infrastructure topology, hiring, GTM, financial framing, and operational handbook.

---

## 2. Market Analysis (preserved from v1)

The addressable market for GreenMetrics in Italy sits at the intersection of three policy tailwinds that are all accelerating simultaneously and all share the same underlying data requirement — namely, continuous and auditable measurement of energy consumption and greenhouse-gas emissions at production-site granularity.

### 2.1 The CSRD tailwind

The first tailwind is the **Corporate Sustainability Reporting Directive** (Dir. UE 2022/2464), which phases in reporting obligations between 2024 and 2028. The first wave (FY 2024, reports published in 2025) captures large public-interest entities that were already under the Non-Financial Reporting Directive: a population of roughly 200 Italian companies. The second wave (FY 2025, reports 2026) extends to all large companies satisfying two of three criteria — 250 employees, €50M turnover, €25M balance sheet — which Confindustria and CDP estimate at around 8 000 Italian entities. The third wave (FY 2026, reports 2027) adds listed SMEs, bringing the total to approximately 35 000 companies that will have to publish ESRS-compliant sustainability statements by 2028. Within that population, ESRS E1 on climate change is the anchor topical standard, because it is the only one both quantitatively measurable and evidence-driven, and because Scope 1 and Scope 2 emissions reporting is a prerequisite for the rest of the narrative (transition plan, metrics, targets, financial effects).

### 2.2 The Piano Transizione 5.0 tailwind

The second tailwind is **Piano Transizione 5.0**, introduced by DL 19/2024 and operationalised by the decreto applicativo MIMIT-MASE of 24 July 2024 with GSE as the paying agent. The plan extends the earlier Piano Transizione 4.0 tax-credit regime to investments that combine Industry 4.0 equipment with demonstrable energy-efficiency improvements — energy reductions of ≥ 3% at process level or ≥ 5% at production-site level unlock credit bands that scale with reduction and spend, producing marginal rates of 5–45% on qualifying outlays up to €50M per company. The national envelope is approximately €6.3B (PNRR allocation) on top of the residual Piano 4.0 balance, bringing the total 4.0+5.0 intervention to above €13B. Uptake in 2024 was slow because the attestazione process demands calibrated baseline-vs-post-intervention measurement that most SMEs could not produce; this is precisely the engagement-fit case for GreenMetrics. MIMIT and MASE have signalled that the 2025 and 2026 tranches will be paid with a more aggressive calendar and stricter technical documentation — further widening the need for a measurement template that outputs the attestazione in the exact format the paying agent expects.

### 2.3 The D.Lgs. 102/2014 audit cycle tailwind

The third tailwind is **D.Lgs. 102/2014**, the Italian transposition of Directive 2012/27/UE, which has since 2015 required large enterprises (>250 employees or >€50M turnover / >€43M balance sheet) and all companies registered as "energivore" with CSEA to commission an independent energy audit every four years. The 2023 cycle's ENEA analysis counted roughly 11 000 audits filed, for about 8 500 obligated undertakings. The next cycle closes 5 December 2027, and the audit must meet EN 16247-1 plus 16247-3 (processes) and/or 16247-4 (transport). A rigorous audit depends on three-year energy-consumption history at a minimum, preferably sub-annual, and preferably per cost-centre. In practice the audit is often stitched together from monthly invoices, producing an expensive but low-quality output. GreenMetrics short-circuits this: if the plant has been on the platform for 12+ months, the audit dossier is essentially a clicked export.

### 2.4 Sector composition

The Verona province counts roughly 92 000 active enterprises in Chamber of Commerce data, of which some 13 500 are industrial (ATECO sections B, C, D, E) and roughly 1 400 are agri-industrial processors (ATECO 10–11). Applying the CSRD / 102/2014 size thresholds, a realistic upper bound of obligated undertakings in the province is on the order of 2 000–2 500. Nationally the equivalent numbers scale to roughly 35 000–45 000 obligated enterprises once the full CSRD perimeter phases in.

### 2.5 TAM / SAM / SOM under the engagement model

Under v1's SaaS framing, TAM was sized at €840M–€1.1B/year priced at €24k/year average. Under v2's engagement model, the right framing is *engagement-eligible companies* — the count of industrial sites that need the regulatory deliverables we ship and have a budget consistent with our €100k–€500k engagement band. Of the 35 000–45 000 nationally-obligated industrial enterprises, the subset with capex programs in the Piano-5.0-relevant €500k–€2M range and the procurement maturity to engage a regulator-grade template is on the order of 5 000–8 000 sites in the 36-month plan horizon. SAM over the same horizon is the Veneto + Lombardy + Emilia-Romagna + Piemonte arc, ~2 500 sites. SOM is what the engagement-team hire plan in §9 supports — 5 engagements through Phase 5 by end of Phase J, scaling to 15–25 engagements/year by Y3 of the engagement business.

### 2.6 Competitive positioning (preserved from v1, sharpened)

Global enterprise ESG platforms target Fortune-500 multinationals, price in five or six figures per site per year, require six-month integrations, and address Italian regulation only as a localisation afterthought. National competitors are usually vertically integrated with a utility or an ESCO relationship, which is great for their own energy-service contracts but adds a procurement layer the customer frequently does not want. GreenMetrics slots between the two: a vendor-neutral, Italian-domiciled specialist, delivered as a forkable template rather than a hosted multi-tenant service.

The full competitive matrix — 30+ competitors across 9 tiers — is in `docs/COMPETITIVE-BRIEF.md`. The TL;DR five-beat-points: (i) audit-grade reproducibility, (ii) OT-native ingestion, (iii) Italian-first regulatory depth, (iv) modular template, not SaaS, (v) engineering doctrine as part of the deliverable. Each beat-point is doctrine-backed (per the Doctrine section/rule references in the brief).

### 2.7 SWOT, post-pivot

- **Strengths:** doctrine-grade engineering, Italian regulatory depth, OT-native ingestion across 7 (now) → 12 (Phase G) protocols, transparent template / source available to engaged client, engagement-margin economics not dependent on meter count, regulatory-deliverable services as a high-margin add-on.
- **Weaknesses:** single-operator business at the start (mitigated by Phase H engagement-team hire), no large pre-existing customer base, requires local partner network to install gateways at scale (mitigated by §9 channel partners), no AGID Qualificazione Cloud per la PA yet (Phase J Sprint S21–S22 deliverable).
- **Opportunities:** the CSRD wave-2/wave-3 deadlines and the 102/2014 December 2027 audit window create a forced-buying calendar; the Italian utility-bundled offerings have a structural Piano-5.0 conflict-of-interest we exploit; partner ESCOs and EGE consultants need an audit-grade platform underneath their consulting work.
- **Threats:** utilities may bundle "free" energy management with supply contracts; hyperscaler ESG platforms may commoditise the dashboard layer; ISPRA or GSE may change the attestazione templates mid-cycle and force a Pack release. Mitigations: vendor-neutral story, doctrine-grade signal-vs-noise differentiation, annual Pack review cadence per Rule 138.

### 2.8 Pricing benchmarks (engagement-model-aligned)

Triangulated from RFP datasets and channel interviews. Entry-level energy-monitoring (10 meters, dashboards only) clears €50–150/meter/month at the SaaS competitor band — that's not what we sell. Mid-tier with reporting (CSRD E1 or ISO 50001 support) clears €200–500/meter/month at the SaaS competitor band — also not what we sell. Enterprise-grade integrated ESG / energy management at the Schneider / Siemens band starts at €2 000/meter/month or moves to per-MWh pricing above a threshold — and *crucially* requires a 6-month integration and a proprietary-stack lock-in.

GreenMetrics' engagement-model pricing in §3 is structured to land at *less than the equivalent 5-year cost* of a Schneider / Siemens deployment at any given site complexity, while delivering audit-grade reproducibility, transparent source, and a 8–14 week Phase 0–5 timeline. The buyer's procurement compare is not "is this €249/meter/month worth it?"; it is "is a €280k engagement delivering audit-grade Piano 5.0 + CSRD E1 + audit-102 + 5 sites onboarded in 12 weeks worth it vs. €1.2M for Schneider over 5 years?"

---

## 3. Engagement Model & Revenue Strategy

GreenMetrics monetises through four stacked streams: engagement license, customisation services, annual maintenance, and (optional) tier retainer. Each stream maps to a different commercial obligation and together they form a margin-resilient mix that holds up under engagement-portfolio expansion.

### 3.1 Engagement license (one-time)

Grants the client perpetual right to use the version of the template delivered. Sized at a percentage of the engagement project budget. Typical band:

- **Light engagement** (1 site, 1 Region Pack already in catalogue, 3–5 Protocol Packs already in catalogue, 1 Report Pack stack — e.g. just CSRD E1 + monthly_consumption): **€40k–€80k**.
- **Standard engagement** (2–3 sites, 1 Region Pack, 5–7 Protocol Packs, 4 Report Packs — e.g. CSRD E1 + Piano 5.0 + Conto Termico + audit-102): **€80k–€140k**.
- **Complex engagement** (4+ sites, multi-Region-Pack, 7+ Protocol Packs, 5+ Report Packs, Identity Pack work, integration with the client's existing SCADA / ERP / SIEM): **€140k–€220k**.
- **Strategic engagement** (an industrial group with 10+ sites, custom Pack development, hybrid topology, fully-managed retainer): **€220k–€500k+**.

The license is *structural*: it pays for the 36 weeks of engineering investment behind v1.0.0, the doctrine, the conformance suite, the runbook catalogue, and the supply-chain attestation that no greenfield build could deliver in the engagement timebox. It is sold once at the start of the engagement; a renewal of the engagement (e.g. major upgrade to v2.0.0) re-prices.

### 3.2 Customisation services (T&M)

Time-and-materials billing against the engagement scope. Typical band €60k–€300k for the customisation sprint and Pack assembly. Includes:

- engineering labour (3–6 weeks Phase 2 Pack Assembly + 2–4 weeks Phase 3 Customisation + 2 weeks Phase 4 Hardening & Soak + 1 week Phase 5 Handover);
- the Macena team's domain advisory (energy-management expertise from the engagement-advisor hire planned for Phase H);
- the regulatory-pack signoff by an EGE-certified partner where applicable (counter-signature work is itself billable to the client per the underlying regulatory-deliverable rate-card).

Customisation services are billed milestone-based — 30% kickoff, 60% draft delivery, 10% acceptance — to align cash flow with delivery progress.

### 3.3 Annual maintenance (T1+)

Annual subscription including Pack updates as the regulatory landscape evolves (e.g., ISPRA factor table is republished every April; GSE Conto Termico XSD updates), security patches against the Rule-59 SLA, and major-version migration assistance. Sized as a percentage of license. Typical 18–22%. Maintenance attach rate target ≥ 90% of T1 closures.

### 3.4 Co-managed retainer (T2 only)

Fixed monthly fee buying named on-call hours, SLA, quarterly reviews. Typical band €4k–€12k/month per deployment. Buys: the Macena team retains shared on-call after handover; SLA ≤ 99.5% on key user journeys; quarterly review with the operator team; Pack updates and security patches included; no surprise per-incident charges.

### 3.5 Fully-managed retainer (T3 only)

Fixed monthly fee buying full operations on a hosting platform of the client's choice (AWS eu-south-1, Aruba Cloud, GCP eur-west, on-prem K3s in the client's DC). Typical band €18k–€55k/month per deployment. Buys: SLA ≤ 99.9%; capacity planning; cost optimisation; regulatory-update tracking; audit-evidence pack production. Ideal for clients buying outcomes, not infrastructure.

### 3.6 Regulatory-deliverable services (one-off, transactional)

The EGE-countersigned Piano 5.0 attestazione, the CSRD ESRS E1 dossier review, the D.Lgs. 102/2014 audit countersignature. Priced as in v1 — €3.5k–€15k for Piano 5.0 attestazione (midpoint €8k), €5k–€50k for CSRD ESRS E1 dossier, €6k–€12k for audit 102/2014 — but delivered through the engagement deployment, not as a separate SaaS feature. These are a recurring revenue line on every active deployment.

### 3.7 What we deliberately do not sell

- We do not sell per-meter monthly subscriptions. The unit of sale is the engagement; the deployment then handles however many meters the client needs at no marginal-revenue model cost to us.
- We do not sell "use the template free, pay for support." The template is not free under the proprietary licence; the engagement license is the entry charge.
- We do not sell access to a centrally-hosted multi-tenant instance for many SMEs. Multiple SMEs can share a deployment if a partner ESCO or system integrator wants to host them, but that's the partner's monetisation, not ours.
- We do not sell unbundled Packs to third parties. Packs are part of the template's value; selling them à la carte invites quality erosion.

### 3.8 Margin profile target

Engagement margin (gross) target ≥ 65% on T1 handovers, ≥ 55% on T2 co-managed, ≥ 45% on T3 fully-managed. The mix shifts toward T2/T3 over the engagement's lifetime as the operator team's risk-tolerance for self-operation declines and the regulatory exposure grows. Annual maintenance attaches at ≥ 90% of T1 closures.

### 3.9 Engagement KPIs (replaces SaaS metrics)

- **Engagement margin.** Gross margin per engagement, computed on a per-engagement P&L (revenue minus engineering + advisory + EGE costs).
- **Time-to-customisation.** Calendar time from SoW signing to Phase 5 Handover. Target median 11 weeks; long-tail 14–18 weeks for multi-site. Phase-overrun beyond 16 weeks is a Sev-2 commercial event.
- **Template-fit-score.** A subjective 1–5 rating from the engagement lead at Phase 5 measuring how well the template (Core + existing Packs) fit the engagement; a 1–2 rating triggers a Pack-or-doctrine evolution proposal. Aggregate target: ≥ 4 across the portfolio.
- **Net-engagement-value.** License + customisation + 5-year-maintenance + 5-year-tier-retainer projection, minus delivery cost. Used to compare engagements at the portfolio level.
- **Annual-maintenance-attach rate.** Percentage of closed engagements that take annual maintenance. Target ≥ 90% on T1; ≥ 95% on T2/T3.
- **Engagement-portfolio velocity.** Engagements through Phase 5 per quarter. Target Q4 2026 = 1 (synthetic), Q1 2027 = 2 (first two real), Q2–Q4 2027 cumulative ≥ 5.
- **Pack-catalogue contribution rate.** Engagements per quarter that contribute at least one generalised Pack back upstream. Target ≥ 50% (every other engagement should improve the template).

### 3.10 Channel partners (preserved from v1, reframed)

Three legs of channel partnership:

- **EGE (energy-management experts) and ESCOs.** Co-sell the platform alongside their consulting mandates. Engagement-margin share: 15–20% on placed deals; the partner is the engagement-lead-of-record; Macena delivers the technical engagement underneath.
- **Commercialisti and Italian fiscal consulting.** Refer engagements where a client is heading toward Piano 5.0 or audit 102/2014 deadlines. Referral commission: 5–10% on first-year engagement license.
- **System integrators (regional Modbus / M-Bus / SunSpec wiring expertise).** Subcontract on Phase 2 Pack Assembly hardware-side work; paid through a set-hours rate-card. The system integrator owns the physical install; Macena owns the platform.

By Y3 of the engagement business, ~35–40% of new engagement-license revenue is expected to come through these channels. Channel partners receive engagement-team training, access to the upstream template via a partner license, and quarterly portfolio reviews.

### 3.11 Compliance calendar as a sales tool

Every quarter the engagement team produces a "compliance clock" for each active engagement — a one-page document showing when the next CSRD filing is due, when the 102/2014 audit window opens, which Piano 5.0 claim windows are approaching, and which Conto Termico calls match the site profile. This is a non-technical artifact pushed to the engagement client's CFO and sustainability lead; it surfaces opportunities for additional Packs / additional sites / additional regulatory deliverables while positioning the deployment as the anchor of the client's regulatory calendar.

---

## 4. Technical Architecture (post Pack extraction)

The GreenMetrics stack is built from five concerns, each of which dictates a discrete technical choice. Ingestion of high-cardinality, high-frequency meter data requires an efficient, concurrent, low-overhead runtime; that is why the backend is written in **Go with the Fiber framework**. The data-layer constraints — time-series volumes in the tens of billions of samples per tenant per year, with on-line rollups needed to meet interactive-dashboard latency SLOs — are exactly the scenario TimescaleDB was built for, which is why we standardise on **TimescaleDB with PostgreSQL 16**. The reporting and operator UI needs a small, fast, accessible surface that avoids the weight of Next.js or the context-switch of Angular; **SvelteKit 2** with Tailwind gives us a sub-50KB critical-path payload and a development velocity consistent with engagement timeboxes. For operator-centred dashboarding we lean on **Grafana**, provisioned from declarative JSON in the repository. Finally, the regulatory-reporting pipeline is implemented as Pure-Function Builders inside Report Packs, with versioned emission factors from Factor Packs and signed JSONB payload storage, so that auditors can reconstruct any report from the source data at any point.

### 4.1 Layer model

Five layers per `docs/LAYERS.md`:

1. **Infrastructure** — Terraform, S3+DynamoDB state backend, KMS, AWS Secrets Manager (Topology A) / Vault (Topology B), per-region partitioning.
2. **Substrate** — Kubernetes (EKS / K3s / on-prem K3s depending on topology), ArgoCD, Kyverno, Falco, External Secrets Operator, cert-manager, Cosign keyless signing, SLSA L2 → L3 attestation.
3. **Backend** — Go 1.26 + Fiber + pgx + TimescaleDB + Asynq + Redis. Layer separation: `internal/api/`, `internal/handlers/`, `internal/domain/<aggregate>/` (DDD per Rule 32), `internal/services/` (orchestration), `internal/repository/`, `internal/security/`, `internal/observability/`, `internal/resilience/`, `internal/jobs/`, `internal/packs/` (Pack-loader). Plus per-Pack code at `packs/<kind>/<id>/`.
4. **Frontend** — SvelteKit 2 + Tailwind. White-label-readiness via `config/branding.yaml`. PDF rendering server-side with deterministic PDF/A-2b output.
5. **Operators** — Engagement-lead, on-call (T2/T3), EGE counter-signers, auditors. Each role has a documented runbook surface and a per-role RBAC mapping.

### 4.2 High-level data flow

A meter — electrical, gas, thermal, water, PV inverter, EV charger — is read by the GreenMetrics edge gateway (Phase G Sprint S14 deliverable) which speaks Modbus RTU / TCP, M-Bus, SunSpec, OCPP, IEC 61850, OPC UA, MQTT Sparkplug B, BACnet, Pulse, IEC 62056-21 (per Phase G Pack catalogue). The edge gateway carries a 24-hour disk-backed buffer (Rule 111), NTP-synced clock with optional GPS time stratum-1 (Rule 112), per-meter HMAC signing at ingestion (Rule 173), and mTLS to the backend.

The backend's `/api/v1/readings/ingest` endpoint receives signed, pre-bundled batches, verifies the HMAC, performs a `pgx` bulk COPY into the `readings` hypertable. TimescaleDB's chunk policy is a 1-day chunk interval; compression policy compresses chunks older than 7 days at 10–20× size reduction. Continuous aggregates roll up to 15-min, 1-hour, 1-day buckets. Retention drops raw rows after 90 days, 15-min after 1 year, 1-hour after 3 years, 1-day after 10 years.

### 4.3 Carbon and reporting

The carbon calculator queries against the appropriate aggregate view and joins versioned `emission_factors` at the temporal midpoint of the query window per Rule 90. Report Packs (`packs/report/<id>/`) implement `Builder.Build(ctx, period, factors, readings)` as pure functions per Rule 91; Core's reporting orchestrator dispatches to the registered Builder by ReportType. Builder output is byte-for-byte deterministic per Rule 89 + Rule 141, signed at finalisation per Rule 144, carries a provenance bundle per Rule 95, and is queryable for lineage per Rule 99.

### 4.4 API

The API surface is versioned at `/api/v1` and documented in `api/openapi/v1.yaml` (the source of truth per Rule 14 and ADR-0013). Authentication is JWT over HTTPS: HS256 with `kid` claim per Rule 170, KID rotation per ADR-0016. Identity Packs (`packs/identity/<id>/`) handle the actual proof-of-identity for local-DB / SAML / OIDC dialects. RBAC via `RequirePermission(...)` middleware per Rule 39. Errors are RFC 7807 Problem Details per Rule 4. Idempotency-Key required on POST per Rule 35 with replay storage in `idempotency_keys` Timescale hypertable (24h retention).

### 4.5 Security architecture

Defence in depth at three layers per Rule 39: repository-level `WHERE tenant_id = $1`, Postgres RLS policies, JWT-claim-driven middleware with `InTxAsTenant`. Field-level encryption on `readings.raw_payload` with per-tenant DEK + KMS-wrapped MEK per Rule 172 (Phase F Sprint S10 deliverable). Per-meter HMAC reading provenance per Rule 173 (Phase F Sprint S11). Chained-HMAC audit log with hourly checkpoint per Rule 169 (Phase F Sprint S11). Manifest-lock signed and verified at admission per Rule 73 (Phase F Sprint S11).

### 4.6 Observability

OTel SDK + OTLP gRPC exporter, sample ratio 0.1 production / 1.0 dev (Rule 18), span coverage at HTTP / pgx / outbound HTTP / ingestor poll. Zap structured JSON logs with mandatory fields per Rule 7. Prometheus exposition at `/api/internal/metrics` with cardinality budgets per Rule 40. Grafana dashboards: energy-overview, carbon-dashboard, engagement-portfolio (Phase E Sprint S8). Health envelope `{status, service, version, uptime_seconds, time, dependencies}` per Rule 6, with per-Pack health surfaced under `dependencies` per Rule 74.

### 4.7 Deployment topologies (per Charter §10)

- **Topology A — Public-cloud single-tenant.** AWS eu-south-1 Milan + AWS Secrets Manager + IRSA. Default. Italian residency satisfied.
- **Topology B — Italian-sovereign-cloud single-tenant.** Aruba Cloud / Seeweb / TIM Enterprise + Vault + K3s. AGID Qualificazione Cloud per la PA satisfied (Phase J).
- **Topology C — On-prem single-tenant.** Client's bare-metal K3s + client's IdP via Identity Pack + client-owned S3-compatible backup.
- **Topology D — Hybrid.** OT-segment ingest backend + IT-segment frontend/reporting + site-to-site VPN + strict NetworkPolicy segmentation.

The chosen topology is locked in the Discovery ADR; switching topologies after Phase 1 is a re-scoping event.

### 4.8 Pack catalogue (target by end of Phase H)

- **Region Packs:** it (flagship), de, fr, es, gb, at.
- **Protocol Packs:** modbus_tcp, modbus_rtu, mbus, sunspec, pulse, ocpp_1_6, ocpp_2_0_1, iec_61850, opc_ua, mqtt_sparkplug_b, bacnet, iec_62056_21.
- **Factor Packs:** ispra, gse, terna, aib, uk_defra, epa_egrid.
- **Report Packs:** esrs_e1, piano_5_0, conto_termico, tee, audit_dlgs102, monthly_consumption, co2_footprint, uk_secr, ghg_protocol, iso_14064_1, tcfd, ifrs_s_1_s_2.
- **Identity Packs:** local_db (default), saml, oidc.

---

## 5. Engagement Lifecycle (replaces v1's "Development Roadmap")

The product roadmap in v1 ("Phase 1 Core Metering, Phase 2 CSRD+Piano 5.0, Phase 3 Scope 3, Phase 4 AI Forecasting") is replaced by the *template roadmap* in `docs/PLAN.md` (Phases E–J). Each *engagement* runs through its own six-phase lifecycle (Phase 0–5) per Charter §8 and Doctrine Rules 149–168.

### 5.1 Phase 0 — Discovery (2 weeks)

Output: signed Scope-of-Work; Pack matrix listing required Region/Protocol/Factor/Report/Identity Packs; integration map (which client systems connect); deployment topology choice; Discovery ADR. No Phase 1 work begins without these artefacts (Rule 149).

Discovery is engineer-led. Discovery deliverables are billable on a fixed-fee basis (typical €15k–€30k for a Standard engagement). If Discovery surfaces a deal-killer (the client doesn't have a clean OT/IT network, the client's regulatory needs are out of catalogue, the client's procurement timeline doesn't fit the engagement model), the Discovery is the deliverable and the engagement closes at Phase 0.

### 5.2 Phase 1 — Fork & Bootstrap (1 week)

Engagement repository created at `github.com/<engagement-org>/<engagement-id>-greenmetrics` per Rule 150. `template-version.txt` records the upstream version. `engagements/<client>/` overlay populated with engagement-specific defaults. Bootstrap successful in client target topology. First staging deploy. `task verify` green on the engagement fork.

### 5.3 Phase 2 — Pack Assembly (3–6 weeks)

Required Region / Protocol / Factor / Report Packs assembled from the catalogue. Identity Pack wired against the client's IdP if applicable. Client-specific data fixtures loaded (synthetic by default per Rule 165). All required Pack contracts satisfied; conformance + property + security tests green on engagement fixtures. Phase 2 is bounded — adding more than two Packs mid-Phase-2 is a re-Discovery event per Rule 151.

### 5.4 Phase 3 — Customisation Sprint (2–4 weeks)

Client-specific UI overlays, custom dashboards, custom report layouts, custom alerts, integration with client SCADA / ERP / SIEM. Branding via `config/branding.yaml`. Customisation scope is the SoW + Pack matrix — beyond-scope requests are change-orders with separate pricing per Rule 152. Phase 3 ends when the client UAT acceptance scenarios are green.

### 5.5 Phase 4 — Hardening & Soak (2 weeks)

Production deploy. Chaos drill (per `docs/CHAOS-PLAN.md`). Failover drill. Capacity test at 1×/3×/5× expected load. Runbook walkthrough with the operator team. Phase 4 is bounded per Rule 153 — a failed drill defers Phase 5 until the drill passes.

### 5.6 Phase 5 — Handover (1 week)

Operator-team training. Runbook handover. On-call rotation arrangement. Postmortem template embedded. First ADR by the operator team filed. Phase 5 ends when the operator team self-reports readiness AND owns the first 7-day on-call shift (T1) or the co-managed lead picks up (T2/T3) per Rule 154.

### 5.7 Total median calendar

11 weeks from SoW to operator-owned production for a Standard engagement. Long-tail engagements with multi-site rollouts run 14–18 weeks. A Phase-overrun beyond 16 weeks is a Sev-2 commercial event triggering an executive-sponsor call.

### 5.8 Engagement closure (Charter §8.3)

Three closure conditions: (1) annual maintenance not renewed and the client has either taken full ownership or migrated to another vendor; (2) client requests termination — Macena delivers the exit pack via `task engagement-exit-pack`; (3) deployment migrates to upstream-of-template — engagement-specific code generalised back as a Pack contribution.

### 5.9 Engagement portfolio expansion arc

- **Q4 2026:** synthetic engagement #0 (internal) ships through Phase 5 to validate the lifecycle. End-of-Phase-E deliverable.
- **Q1 2027:** real engagements #1 and #2 begin. Sized at Standard engagements. Italian flagship Pack matrix.
- **Q2–Q4 2027:** real engagements #3, #4, #5 ship. Mix of Standard and Light. By end of Phase J Sprint S22, five engagements have run end-to-end through Phase 5.
- **2028:** scale to 15–25 engagements/year based on the engagement-team hire plan in §9. Region-Pack expansion to DE / ES drives international engagement opportunities.

---

## 6. Scaling Strategy (engagement portfolio)

Scaling under the engagement model is not the SaaS scaling axes (request rate, tenant count, meter density at the central node). It is engagement-portfolio scale: how many concurrent engagements can the Macena team support without quality regression, and how does the Pack catalogue absorb each new engagement so that engagement #N is faster than engagement #N-1.

### 6.1 Engagement-portfolio capacity

A single engagement-lead supports 2–3 concurrent active engagements (one in Phase 2 Pack Assembly + one in Phase 3 Customisation + one in Phase 4 Hardening). With the founder-only team in 2026–early 2027, capacity is ~3 active engagements at any given time. Phase H engagement-team hire (a second engagement lead + an engagement-advisor) raises capacity to 6–8 active engagements. By end of Y3, with five engagement leads the capacity is ~15 active engagements.

### 6.2 Pack-catalogue absorption

The whole point of the Pack model is that engagement #N's Pack work absorbs into engagement #N+1's reuse. Concretely:

- Engagement #1 builds the Italian-flagship Pack catalogue (Phase E Sprint S6–S7 already started).
- Engagement #2 reuses 95% of Pack code; spends ~20% of its time on engagement-specific code.
- Engagement #3 reuses 95%+ on a Standard engagement; spends time on edge cases.
- Engagement #4 (German) extends the Region Pack catalogue with `packs/region/de/`; subsequent German engagements reuse.

The Pack-contribution rate (Rule 168 portfolio feedback) measures how well the absorption is happening: target ≥ 50% of engagements contribute at least one generalised Pack back upstream.

### 6.3 Per-deployment scale (engagement size)

An individual deployment's scale axes are still the v1 axes — meter ingest rate, query latency, storage volume, dashboard concurrency — but they are now per-deployment, not per-tenant. A deployment running 100 meters with 15-minute resolution is the entry-level Topology A; a deployment running 5 000 meters with 1-minute resolution is a Topology A or D with horizontal-scale-out backend pool. The capacity model in `docs/CAPACITY.md` carries the per-deployment sizing.

### 6.4 Geographic expansion (engagement-driven)

Geographic expansion in v2 is engagement-driven, not roadmap-driven. The first engagement in DACH (likely Austria via Italian industrial-district relationships across the Alpine border) triggers `packs/region/at/` and `packs/region/de/` extraction. The first engagement in Iberia triggers `packs/region/es/` and `packs/region/pt/`. The first engagement in the UK triggers `packs/region/gb/`. The Pack catalogue absorbs each new region; subsequent engagements in that region reuse the work.

The 24-month horizon does not target outside-EU engagements — non-EU data residency, factor-set fragmentation (Italian mix ~0.245 vs. German 0.380 vs. French 0.055), and regulatory-pack divergence are distractions from the core ICP.

---

## 7. Security & Compliance (engagement-aware)

Security posture rests on the four pillars of v1 (identity, transport, storage, audit) plus three engagement-aware pillars (per-deployment isolation, per-engagement pen-test scope, per-engagement compliance evidence). The doctrine governs (Rules 39, 169–188); this section narrates.

### 7.1 Per-deployment isolation

Every engagement runs in a dedicated deployment by default (single-tenant). Multi-tenant deployments occur only when a partner ESCO / system-integrator hosts multiple SMEs — and that's the partner's risk model, not Macena's. Defence-in-depth tenant isolation (RLS + RBAC + repository-level WHERE) remains in place even in single-tenant deployments because it eliminates an entire class of bug from the attack surface.

### 7.2 Per-engagement pen-test scope

Annual pen-test (Rule 60, Phase J Sprint S21–S22) covers the *template*. Engagement-specific code (under `engagements/<client>/`) is pen-tested either by the engagement client's own security team or by an extended scope of the annual pen-test, billable separately per the engagement contract. Findings are tracked in `docs/PENTEST-CADENCE.md` (template) and `engagements/<client>/PENTEST-LOG.md` (engagement-specific).

### 7.3 Per-engagement compliance evidence

Each engagement deployment can produce its own audit-evidence pack via `task audit-evidence-pack` per Rule 108 (Phase H Sprint S15 deliverable). The pack contains: every audit-log row in the period, every report with provenance and signature, every Pack manifest with lock hash, the OpenAPI spec, the Cosign signatures, the SBOM, the conformance-suite green status, the running-system manifest. The pack is itself signed.

### 7.4 GDPR posture

Stricter than baseline because ARERA deliberation 646/2015 and Measure 147/2023 designate near-real-time consumption data as personal data even when the POD owner is a legal entity, whenever the data can be re-identified to a natural person. We treat meter readings as potentially-personal by default and apply: (a) field-level encryption on `readings.raw_payload` via AES-256-GCM with per-tenant DEK + KMS-wrapped MEK (Rule 172); (b) data minimisation; (c) DSAR endpoint (Phase F Sprint S10); (d) crypto-shredding for Art. 17 erasure (Rule 184). Cross-EU transfer restricted to opt-in only per Rule 8.

### 7.5 Italian-specific regulations (preserved from v1)

- AgID Cloud classification (circolare n.1/2021): Topology B deployments serving PA / municipal utilities / public hospitals run on Aruba / Seeweb / TIM Enterprise.
- NIS2 D.Lgs. 138/2024: GreenMetrics inherits soggetto-importante perimeter via supply-chain obligations under Annex III. Annual tabletop exercise per Rule 167.
- AEEGSI / ARERA directives on smart-meter data sharing operationalised in the Italian Region Pack's E-Distribuzione SMD / Terna / SPD client integrations (Rule 125).
- D.Lgs. 102/2014 audit integrity protected by audit-log immutability (Rule 62) + report-signing at finalisation (Rule 144) + EGE counter-signature workflow (Rule 137).

### 7.6 Audit-log immutability and chain integrity

Beyond the role-revoke append-only invariant (`00099_audit_lock.sql`), Phase F Sprint S11 wires the chained-HMAC audit log per Rule 169 with hourly published checkpoints to a WORM-mirror (S3 Object Lock compliance mode for Topology A; equivalent for B/C/D). Tampering with any row breaks the chain at that row and surfaces in the conformance-suite verification.

### 7.7 Encryption and key management

- At rest: AES-256-GCM on `raw_payload`; bcrypt+pepper on `users.password_hash`; KMS-wrapped at the volume level.
- In transit: TLS 1.3 only on customer-facing edges (Rule 174); HSTS preload; mTLS internal where reasonable.
- Cert rotation: cert-manager + Let's Encrypt (public); cert-manager + in-cluster CA (internal).

### 7.8 Incident response

NIS2 timelines: 24h early warning, 72h initial notification, 30d final report. Sev-1 within 1 hour, Sev-2 within 4h, Sev-3 within 24h to runbook completion. Annual tabletop with simulated ACN coordination (Rule 167). Engagement-specific incidents follow the engagement-fork's `engagements/<client>/INCIDENT-RESPONSE.md` overlay.

### 7.9 Supply chain

Cosign keyless signing (ADR-0017) on every production image. SLSA L2 provenance (ADR-0018), L3 plan dated. SBOM via Syft. Trivy scan post-build, fail HIGH/CRITICAL. Kyverno admission policy denies unsigned. Manifest-lock signature verified at admission per Rule 73 (Phase F Sprint S11).

---

## 8. Infrastructure & DevOps (per topology)

### 8.1 Topology A — Public-cloud single-tenant (default)

AWS eu-south-1 Milan. Three AZs. Backend: 2–6 replicas behind ALB with HPA on CPU + custom `readings_ingest_rate`. Frontend: 2 replicas behind CloudFront. TimescaleDB: 1 primary + 2 streaming replicas (AZ-split). Grafana: single replica with snapshotted volume. AWS Secrets Manager + ESO. KMS per-env keys. RDS-Timescale Terraform module per `terraform/modules/rds-timescale/`. ArgoCD GitOps. Cosign verify at admission.

### 8.2 Topology B — Italian-sovereign-cloud single-tenant

Aruba Cloud (Arezzo or Bergamo) or Seeweb or TIM Enterprise. K3s. Vault for secrets. cert-manager + trust-manager for in-cluster PKI (no AWS Secrets Manager equivalent). ArgoCD. Cosign + Kyverno same as Topology A.

### 8.3 Topology C — On-prem single-tenant

K3s on client's bare-metal in client's DC. Identity Pack against client IdP (SAML / OIDC / AD) is mandatory. Backups to client-owned S3-compatible store. ArgoCD. Cosign + Kyverno enforced.

### 8.4 Topology D — Hybrid

OT-segment ingest deployed inside the OT zone; frontend / reporting in IT segment; site-to-site VPN with strict segmentation. NetworkPolicy bundles per-segment. The OT-segment ingest backend speaks only to OT-segment meters and to a single egress proxy at the IT-segment boundary.

### 8.5 Monitoring and alerting

Prometheus scrapes the backend's `/api/internal/metrics`. Alertmanager rules: error rate > 1% for 5min (P1), ingestion lag > 5min for 10min (P1), Timescale connection pool > 80% (P2), OTel exporter backlog > 10s (P3), scheduled compression failed (P2), CAGG refresh lag > 2× interval (P2). Each alert annotated with `runbook_url` per Rule 40.

### 8.6 Disaster recovery

RPO 1 hour (Timescale WAL streaming + nightly pg_dump). RTO 4 hours (automated restore runbook with pre-warmed replica). DR drills quarterly per Rule 107. Last drill recorded in `docs/CHAOS-LOG.md`. Backups encrypted with separate KMS key (single-key-compromise resilience).

### 8.7 CI/CD

Per `docs/PIPELINE-MAP.md`. Build → sign → SBOM-attest → SLSA-provenance-attest → image-scan → policy-check → deploy-canary → analyse-SLO-burn → promote → observe-post-deploy. Argo Rollouts AnalysisTemplate reads SLO burn-rate. Manual approval is one of the gates, not the only gate (REJ-24).

### 8.8 Cost observability per deployment

Every cloud resource tagged with `Project=greenmetrics-<engagement-id>`, `Environment=<env>`, `Tenant=<tenant-id>`. Daily cost-report job aggregates CUR into a Grafana dashboard. SRE reviews anomalies > 10% day-over-day. Per-engagement cost-margin reporting flows into the engagement P&L.

---

## 9. Team Structure & Hiring Plan

### 9.1 Year 1 (now → end of Phase J)

**Founder / CTO / Lead Engineer (existing).** Owns the Go backend, TimescaleDB schema, Pack architecture, charter, doctrine, plan. Profile: Renan Augusto Macena. Compensation: founder.

**Founding Engagement-Lead Engineer (Phase H hire, Sprint S15).** Profile: 8–12 years backend Go / industrial-software experience, prior Italian-industrial-energy or SCADA work, fluent Italian and English. Owns the engagement lifecycle Phase 0 → Phase 5 for engagements #1–#3. Compensation: €75–95k base + equity.

**Founding Energy / EGE Advisor (Phase H hire, Sprint S15).** Profile: EGE certified per UNI CEI 11339 or CMVP-certified auditor, 10+ years in Italian industrial energy management, active relationships with Federesco / FIRE / Assoege, prior ESCO and commercialisti work. Co-signs the first 50 Piano 5.0 attestazioni and audit 102/2014 dossiers. Compensation: €60–80k base + 5% warrants, or board-advisor + revenue-linked retainer.

**Founding Sales / Engagement-Director (Phase H hire, Sprint S16, contingent on engagement #1 closing).** Profile: 7+ years industrial B2B sales in Northern Italy, prior Confindustria network, fluent in commercialista / CFO procurement language. Owns the engagement-pipeline development + channel-partner relationships. Compensation: €65–85k base + commission on engagement margin.

### 9.2 Year 2 (post-v1.0.0, engagement #4–#10)

Add: second engagement-lead engineer, second EGE advisor (or formal partner-EGE network), one-day-a-week security advisor (formalising RISK-006 mitigation).

### 9.3 Year 3 (engagement portfolio scaling)

Engineering: 4–5 engagement leads, 1 platform engineer (substrate ops), 1 security engineer (DevSecOps), 1 data/ML engineer (forecaster + drift detection ops), 1 frontend engineer.
Engagement: 2–3 sales / engagement directors, 2 customer-success leads (T2/T3 retainer support), 1 EGE-network manager.
Operations: founder transitions to CEO; CFO / COO at VP level for series-A / strategic-investor conversation.

### 9.4 Total Y3 headcount target

15–20 FTE. The engagement-margin economics support this headcount at portfolio scale of 15–25 engagements/year.

### 9.5 Hiring discipline

Per Rule 9 (platform discipline) and Rule 49 (DevSecOps discipline), every team has a charter, a scope, an authority, a termination criterion. New roles file an ADR justifying the role's existence and the hire.

---

## 10. Go-To-Market

### 10.1 Direct sales (Y1)

Founder-led + Engagement-Director-led conversations with industrial CFOs, Sustainability Leads, Plant Managers in Verona / Veneto / Lombardia. The wedge is the Piano 5.0 + 102/2014 + CSRD trio: a single engagement covers all three calendar events.

### 10.2 Channel partners (Y1–Y3)

Three legs per §3.10:

- EGEs / ESCOs co-sell at 15–20% engagement-margin share.
- Commercialisti refer at 5–10% first-year license commission.
- System integrators handle physical install at set-hours rate-card.

### 10.3 Industry-body engagement

Federesco, FIRE, Assoege, Confindustria Verona, ANIE, Assoege. Speaking engagements at Pre-CSRD-wave-2 conferences (Q4 2026). Guest articles in Eunoé / Energia Italia / Forum PA. Quarterly white-paper drop pushed to industry-body newsletters (a regulatory-pack-update primer for the obligated cohort).

### 10.4 Reference architecture publications

Open-publish (after license decision in Phase J Sprint S22, per ADR-0053) the architecture, the doctrine (or the doctrine's structure), the conformance-suite contents, the runbook catalogue. Auditors and procurement reviewers like to read what they can audit; transparency is a sales asset.

### 10.5 The compliance calendar

§3.11 already described. The compliance clock as a quarterly outbound is the single most effective up-sell mechanism — per the v1 SaaS analysis, accounts that received the clock showed 18–22% annual up-sell vs 8–10% without; the engagement-model equivalent is engagements that move from T1 to T2 within 12 months and engagements that add Packs (e.g., a customer that started with CSRD E1 + Piano 5.0 adds audit-102 + Conto Termico in year 2).

---

## 11. Financial Framing

The v1 financial framing was SaaS-shaped (ARR, ARPA, churn, blended CAC, LTV/CAC). The engagement-model framing replaces those with engagement-portfolio metrics.

### 11.1 Revenue projection

- **Y1 (2026 H2 + 2027 H1).** 1 synthetic engagement (internal, no revenue) + first 2 real engagements closing in Q1–Q2 2027. Revenue = €280k (engagement #1 license + customisation) + €240k (engagement #2 license + customisation) + €40k (regulatory-deliverable services on engagement #1 + #2) = **€560k**.
- **Y2 (2027 H2 + 2028 H1).** Engagements #3–#7. Revenue = 5 × €260k average + 2× €18k T2 retainer × 12 mo + €120k regulatory-deliverable services = **€1.95M**.
- **Y3 (2028 H2 + 2029 H1).** Engagements #8–#15. Revenue = 8 × €280k + 4× €25k T2 retainer × 12 mo + 1× €40k T3 retainer × 12 mo + €280k regulatory-deliverable services + €380k annual maintenance attach = **€4.4M**.

### 11.2 Margin projection

Engagement-margin (gross) target ≥ 65% T1, ≥ 55% T2, ≥ 45% T3. Y1 gross margin ~62% (mostly T1 with founder-time-heavy customisation). Y2 gross margin ~58% (mix of T1/T2). Y3 gross margin ~57% (mix of T1/T2/T3 with T3 lower-margin but more recurring).

### 11.3 Net engagement value (5-year)

A typical Standard T1 engagement at €160k license + €130k customisation + €120k 5-year-maintenance + €30k 5-year-regulatory-deliverable services = **€440k** 5-year revenue, with a 5-year delivery cost of ~€140k = €300k net. Compared to v1's "LTV/CAC > 20" claim — those metrics don't apply, but the engagement net-value is similar in magnitude per engagement.

### 11.4 Investor framing

Engagement-portfolio companies of this profile (specialised industrial software with regulatory wedge) are typically valued at 4–8× revenue or 8–12× engagement-margin EBITDA, depending on the growth-rate. Y3 target supports a €17M–€53M valuation envelope at exit.

The investor narrative is *not* "we're a SaaS doing €5M ARR by Y3 with 30%+ growth" (the v1 narrative). It is "we're an engagement-platform company with 15+ regulator-grade industrial deployments, a defensible Pack catalogue, and the only audit-grade modular template in the surveyed competitive field."

### 11.5 Strategic-acquisition scenarios

Plausible acquirers:

- A large Italian / European industrial-software vendor wanting Italian regulatory depth (e.g., Var Group, Engineering Ingegneria Informatica, Almaviva, Cybertec).
- A global ESG / sustainability platform wanting OT-native ingestion + Italian flagship (e.g., Watershed, Persefoni at scale).
- A Tier 1 OT vendor wanting a vendor-neutral wedge (e.g., one of Schneider / Siemens / ABB / Honeywell).

The engagement-team's bar for conversation is "the acquirer commits to maintaining the modular-template model post-acquisition." Sale to an acquirer that would convert it to a closed SaaS is rejected on principle — that defeats every doctrine rule and every existing engagement client's expectation.

### 11.6 Strategic-independence scenarios

If neither acquisition nor strategic investor materialises, the base case is a cash-flow-positive operating company growing at 25–35%/year on retained earnings. Italian engagement-software companies in the €5–15M revenue range have historically been viable indefinitely as bootstrapped specialists; the engagement-margin economics support founder-and-team livelihood without external capital.

---

## 12. Operational Handbook

### 12.1 Daily

- Stand-up across active engagements (Phase 2 / 3 / 4 / 5).
- Per-engagement health-dashboard review (uptime, incident count, drill freshness, sync recency).
- Continuous-verification loop (Rule 64) — `task verify` green on every PR.

### 12.2 Weekly

- Per-engagement client touchpoint (typically the engagement-lead with the client's energy manager).
- Pack-catalogue housekeeping (open PRs, sync upstream into engagement forks per Rule 79).
- Cost-anomaly review on per-deployment dashboards.

### 12.3 Monthly

- Engagement-portfolio health review.
- Compliance-clock production for each active engagement.
- Pack-version review for the active engagements.

### 12.4 Quarterly

- Charter office hours (every Q1).
- Doctrine office hours (every Q2).
- Engagement-portfolio review with finance (every Q).
- Chaos drill on the synthetic engagement (every Q).
- DR drill on staging (every Q per Rule 107).
- Capacity model refresh.
- Pack annual-review checklist for Italian Packs (every Q1, after April ISPRA update).

### 12.5 Annual

- Pen-test (Rule 60).
- Tabletop exercise (Rule 167).
- License-and-charter review (Rule 209).
- Doctrine evolution session (Rule 209/210).
- Y-end engagement-portfolio retrospective.

---

## 13. The Italian-flagship Pack — what it is and why

Per Rule 88 and Charter §3.2, the Italian Region Pack is the flagship reference. The pack is the most complete, most tested, most documented Pack in the upstream. Other Region Packs (DE / ES / FR / GB / AT) are reviewed against the Italian flagship for thoroughness. The flagship's CHARTER, manifest, README, conformance tests, and Tradeoff Stanza are the template for new Region Packs.

The Italian flagship comprises:

- **`packs/region/it/`** — timezone Europe/Rome, locale it_IT.UTF-8, currency EUR with comma decimals, Italian national + Veneto regional holidays, ARERA / Garante / GDPR / NIS2 D.Lgs. 138/2024 invariants, default regulatory regimes (CSRD wave 2/3, Piano 5.0, Conto Termico, TEE, audit 102/2014, ETS, ARERA, GDPR, NIS2 Italia).
- **`packs/factor/{ispra, gse, terna, aib}/`** — temporal-validity factor sources for the Italian electrical mix, the GSE renewable shares, the AIB residual mix, the Terna daily national mix.
- **`packs/report/{esrs_e1, piano_5_0, conto_termico, tee, audit_dlgs102, monthly_consumption, co2_footprint}/`** — pure-function Builders for the seven Italian regulatory dossiers (with EFRAG XBRL / GSE XSD / ENEA XSD validation in Phase H Sprint S15–S16).
- **`packs/protocol/{modbus_tcp, modbus_rtu, mbus, sunspec, pulse, ocpp_1_6, ocpp_2_0_1}/`** — the seven OT protocols of the Italian flagship (with IEC 61850 + OPC UA + MQTT Sparkplug B + BACnet + IEC 62056-21 added in Phase G Sprint S12–S14).

The Italian flagship is the engagement-team's primary reference. New engagements that are Italian deployments compose entirely from the flagship's Packs + the engagement-fork's overlay; no new Pack development is required unless the engagement has a non-flagship vendor (e.g. a non-Schneider, non-Carlo-Gavazzi, non-Socomec meter).

---

## 14. The doctrine and the codebase as marketable assets

Two non-traditional sales assets:

### 14.1 The doctrine

`docs/DOCTRINE.md` carries 210 rules. Each rule has a body, a Why, a How-to-apply, cross-refs. The conformance suite mechanically enforces a subset; the rest is named-CODEOWNERS reviewable. A regulator / auditor / acquirer / engagement-prospect who reads the doctrine sees the engineering substrate that makes our claims auditable. No surveyed competitor has a comparable artefact.

### 14.2 The codebase + supply chain

The codebase is open to the engaged client (proprietary licence; future open-source decision in Phase J Sprint S22 per ADR-0053). The supply chain is signed (Cosign keyless), attested (SLSA L2 → L3 plan), SBOM-tracked (Syft + DependencyTrack), vulnerability-scanned (Trivy + osv-scanner + govulncheck + CodeQL), and the manifest-lock is itself signed (Phase F Sprint S11). A regulator running their own diff-scan against our images can verify what they're auditing.

Both assets are sold *together* with the engagement; they are not unbundled.

---

## 15. The engagement contract structure

Two-part commercial agreement:

- **Master Services Agreement (MSA).** Carries standard terms: SLA, data-processing under Art. 28 GDPR, liability caps, termination for cause, intellectual-property allocation (Macena retains template ownership; client owns engagement-fork and engagement-specific code), Italian-law / Verona-court jurisdiction, GDPR-DPA compliance.
- **Engagement Order Form (per engagement).** Names the engagement license amount, the customisation services scope and budget, the deployment topology, the Pack matrix, the calendar, the Phase-gate definitions, the optional T2/T3 retainer terms, the regulatory-deliverable services rate-card.

Annual maintenance is governed by an **Annual Maintenance Order** under the same MSA. T2/T3 retainers are governed by a **Retainer Order** with month-to-month terms after a 6-month minimum.

The MSA is tested for the Italian SME procurement context — typical procurement cycle is 6–12 weeks from initial conversation to signed MSA + first Engagement Order. Macena retains a Verona-based legal counsel for MSA negotiation cycles.

---

## 16. Risks and mitigations (engagement-model-specific)

- **Engagement-overrun risk.** A Phase 2/3 that overruns its scope. Mitigated by Rules 151/152 (Phase-bounded + change-order discipline) and the engagement-margin tracking in §3.9.
- **Engagement-lead bus-factor.** A single engagement-lead is the only person who knows engagement #N. Mitigated by Rule 161 (T2/T3 on-call rotation), Rule 23 (substrate operable by single on-call), and the runbook discipline.
- **Pack-extraction regression.** Sprint S6–S7 Pack extraction breaks an existing capability. Mitigated by RISK-024 (Plan §12.1) and the Mission II audit cases run as regression tests.
- **Regulatory schema change mid-cycle.** ISPRA / GSE / ENEA / EFRAG publish a schema delta during a Pack release window. Mitigated by Rule 138 (annual review) + formal-spec validation (Rule 131) + 12% regulatory-agility staffing buffer (preserved from v1).
- **Engagement-team hiring risk.** Phase H hires don't land on time. Mitigated by deferring T2/T3 expansion until the hire lands; founder remains engagement-lead for engagements #1–#3 if necessary.
- **License-decision delay.** v1.0.0 license decision (BUSL / SSPL / AGPL / proprietary) is deferred past Phase J. Mitigated by ADR-0053 placeholder and the explicit Charter §14.1 decision-deferred record.

---

## 17. The contract with the reader of this document

If you read this MODUS_OPERANDI v2:

- as an engineer joining the engagement-team — read it alongside the Charter and the Doctrine. The Doctrine binds your PRs; the Charter binds your architectural moves; this document binds your commercial expectations.
- as an engagement client — read it before signing the SoW. Section 3 governs the commercial mechanics; section 5 the lifecycle you'll experience; section 7 the security posture you can audit.
- as a channel partner — read sections 3.10, 9, 10. The engagement-margin share, the founder-led + sales-led ramp, the industry-body engagement plan.
- as an investor / strategic acquirer — read sections 11, 14, 16. The financial framing, the doctrine-and-codebase asset, the risks.
- as a regulator / auditor — read sections 7, 13, 14. The compliance posture, the Italian-flagship Pack, the doctrine-and-codebase as auditable evidence.

The contract: this document, the Charter, the Doctrine, and the Plan together form the commitment Macena has made to the GreenMetrics modular template. PRs / engagements / charters that violate the commitments without an explicit supersession are blocked. Quarterly office hours surface drift. Annual review re-affirms (or amends with evidence).

The plan is the substrate. The doctrine is the bar. The charter is the spine. This document is the commercial face. The codebase is what the customer pays for.

---

*Document version 2.0 — adopted 2026-04-30. Supersedes v1 per ADR-0028. Reviewed at six-month intervals; next review 2026-10-30.*
