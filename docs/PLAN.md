# GreenMetrics Master Plan — Uplift to Category Leader

> **Status:** Adopted
> **Date adopted:** 2026-04-30
> **Authors:** @ciupsciups (Renan Augusto Macena)
> **Charter parent:** `docs/MODULAR-TEMPLATE-CHARTER.md`
> **Doctrine parent:** `docs/DOCTRINE.md` (210 rules)
> **Review date:** 2026-10-30 (six-month checkpoint, end of Phase E + start of Phase F)
> **Required reading before contributing:** `docs/MODULAR-TEMPLATE-CHARTER.md`, `docs/DOCTRINE.md`, this file in full.

---

## 0. How to read this plan

This is the canonical, normative master plan that takes GreenMetrics from its current state (post Sprint S4, MODUS_OPERANDI v1, Mission II PASS verdict) to the v1.0.0 modular-template release and onward to the v1.5.x state where the template is *measurably better than every credible competitor on the dimensions that matter to a regulator-grade industrial customer*. The plan is organised in six phases (E through J) running across sixteen sprints (S5 through S20, with optional S21–S22 hardening). Each phase derives every deliverable from one or more rules in `docs/DOCTRINE.md` and from clauses in `docs/MODULAR-TEMPLATE-CHARTER.md`.

The plan is also a *commitment*: every deliverable has an exit criterion, a doctrine-rule citation, an ADR trigger (where applicable), and a risk-register delta. PRs that ship without aligning to this plan are blocked at code review unless an explicit replan ADR is filed (Rule 27 / 47 / 67). The plan itself is reviewed at six-month intervals.

The plan is structured for two audiences. The first is the engineer reading on Day 1 of Sprint S5, who needs to know what to do this week. For that audience, the per-sprint sections at the head of each phase are the operational handbook: the Sprint summary, the deliverables list, the exit criteria, the daily-standup check-list. The second is the architect / auditor / acquiring-investor / engagement-prospect reading the plan to assess the platform's maturity. For that audience, the phase intros and the doctrine-traceability matrix at §X are the answer.

You can read the plan front-to-back. You can also read it phase-by-phase as the calendar advances. Each phase is ~1000 lines and self-contained.

---

## 1. Executive summary

GreenMetrics today is a regulator-grade industrial energy-management and sustainability-reporting platform with a Go + TimescaleDB + SvelteKit stack, a 60-rule operational doctrine, 20 ADRs, 35 explicitly-rejected anti-patterns, full GitOps with Cosign keyless signing and SLSA L2 attestation, Italian-compliance depth (CSRD ESRS E1, Piano Transizione 5.0, Conto Termico 2.0, Certificati Bianchi TEE, D.Lgs. 102/2014 audit energetico), and OT-native ingestion across Modbus / M-Bus / SunSpec / Pulse / OCPP. It carries a Mission II PASS audit verdict from 2026-04-18 and four sprints of executed delivery (S1 doctrine adoption + S2 contracts + S3 security wiring + S4 resilience).

GreenMetrics tomorrow is a *modular template* — a deeply opinionated reference implementation that is delivered to clients as a complete artefact (source, infrastructure, doctrine, runbooks, audit evidence) and specialised per engagement out of the template, not built from scratch. The template carries 210 rules of doctrine, audit-grade reproducibility (every regulatory output bit-perfect re-derivable from raw + factor versions + Pack versions), formal-spec validation against EFRAG XBRL / GSE XSD / ENEA XSD, signed audit-log chain, signed manifest lock, signed reports, signed reading provenance, and Italian Region Pack as the flagship reference for German / French / Spanish / British Region Packs that follow.

The plan from today to that future state is six phases over eighteen-to-twenty sprints (S5–S22, ≈ 36–44 weeks), with explicit deliverables, exit criteria, doctrine traceability, and risk deltas at every sprint. Total engineering investment: 80% of capacity over the period; remaining 20% reserved for unplanned engagement work, security incidents, and a 12% regulatory-agility buffer (as MODUS_OPERANDI v1 already accounted for).

The phases:

- **Phase E — Engagement & Pack Extraction (S5–S8, ~16 weeks):** strip the SaaS framing, extract Italian-flagship code into `packs/region/it/` + Factor / Report / Protocol Packs, formalise Pack contracts, ship the engagement-fork model and upstream-sync discipline, hit v1.0.0.
- **Phase F — Audit-Grade Reproducibility (S9–S11, ~12 weeks):** wire the chained-HMAC audit log, the signed reading provenance with per-meter HMAC, the deterministic Builder framework, the bit-perfect replay test for every Builder, the manifest-lock signature and verification at admission, the audit-evidence-pack export.
- **Phase G — OT Integration Maturity (S12–S14, ~12 weeks):** ship the edge gateway with disk-backed buffering, NTP/GPS time discipline, per-protocol simulators, IEC 61850 + OPC UA + MQTT Sparkplug B Packs, network-segmentation-aware Topology D, channel-mapping auditor surface.
- **Phase H — Regulatory Pack Catalogue (S15–S17, ~12 weeks):** validate every Italian Pack against the formal spec, ship UK-SECR + GHG-Protocol-Corporate-Standard + ISO-14064-1 + TCFD + IFRS-S1/S2 Report Packs as the second-wave reference, ship UK-DEFRA + EPA-eGRID Factor Packs, ship the German Region Pack (DACH expansion preparation).
- **Phase I — AI/ML & Forecasting (S18–S20, ~12 weeks):** ship the consumption forecaster (LightGBM + seasonal ETS hybrid) with the documented evaluation budget (MAPE ≤ 12% day-ahead), the layered anomaly detection (z-score + STL + cross-meter), the model-card discipline, the drift detection, the inference observability, the explainability surface, the optional MSD flexibility-market connector.
- **Phase J — Hardening & Certification (S21–S22, ~8 weeks):** SOC 2 Type II preparation, ISO 27001 readiness, AGID Qualificazione Cloud per la PA dossier, NIS2 D.Lgs. 138/2024 incident-response tabletop, post-quantum-crypto plan, third-party penetration test, formal v1.0.0 release announcement and licensing-decision ADR.

At the end of Phase J, GreenMetrics holds the *regulatory-grade modular template* position with no credible competitor in any single dimension and decisive advantage in the dimensions that matter to the Italian-energy-intensive-SME buyer profile and to the partner-ESCO / consulting-accelerator / system-integrator channel that we'll be selling through.

---

## 2. Current state assessment

The current state is the boundary condition of the plan. We assess it across eight axes and surface the gaps that the plan closes. Each axis carries a Mission II audit reference (`MII:`), a doctrine-rule reference, and a current-state-as-of-2026-04-30 score (1 = nascent, 5 = exceeds bar).

### 2.1 Code substrate

- Backend: ≈ 8 000 LoC Go across `internal/{api,config,handlers,jobs,metrics,models,observability,repository,resilience,security,services}`. Layer separation is partial — `internal/services/` carries domain logic that needs to migrate into `internal/domain/` per Rule 32. `internal/repository/timescale_repository.go` is 298 LoC and is mid-decomposition (AB-03 active trigger). The 525-LoC `report_generator.go` god service (REJ-11) is on the Phase E backlog.
- Frontend: SvelteKit 2 with `frontend/src/lib/api.ts` typed client, `frontend/src/lib/components/` (eight components: ConsumptionHeatmap, Scope123Breakdown, EnergyChart, MeterList, ESGScorecard, EmissionFactorEditor, AlarmFeed, CarbonFootprint), routes for `/`, `/meters`, `/readings`, `/reports`, `/carbon`, `/settings`. Tailwind CSS, vite-built. Auth via `hooks.server.ts`. White-label readiness is *partial* — brand strings are coded.
- IaC: Terraform modules for VPC, EKS, RDS-Timescale, IRSA, GitHub-OIDC, S3, Secrets, IAM-breakglass, CloudFront-WAF. State backend (S3 + DynamoDB + KMS) materialised. Bootstrap module in place. Per-environment roots not yet split (REJ-08 partially mitigated).
- GitOps: ArgoCD applications in `gitops/argocd/applications/`. Kyverno policies for image-signature-verify + PSS-restricted + required-resources. Conftest bundles for K8s, Dockerfile, Terraform. Falco rules in place. External Secrets Operator wired.
- Supply chain: Cosign keyless signing pipeline (`.github/workflows/supply-chain.yml`). SLSA L2 provenance attest. SBOM via Syft. Trivy image scan post-build. CodeQL Go + JS extended security pack. OSV scanner. SHA-pinned actions weekly Dependabot.
- Tests: unit tests in `backend/internal/*_test.go`, integration tests scaffolded in `backend/tests/{integration, contracts, security, property, conformance, load, leak, migrations, static, bench}/` (tier directories present; not all populated to bar). Conformance suite is partial — universal-invariants tests are on Phase E backlog.

**Score:** 4/5. Strong stack, strong layer-aware structure, strong supply chain. Gap: domain refactor + conformance suite + test population.

**Doctrine-rule references:** 12, 13, 14, 24, 30, 32, 33, 36, 41, 43, 44, 53, 54, 56, 57.

### 2.2 Schema and data discipline

- TimescaleDB hypertable on `readings` (1-day chunk interval). Three CAGGs (15-min / 1-hour / 1-day). Retention policies (raw 90d, 15m 1y, 1h 3y, 1d 10y). Compression (chunks > 7d, segment-by `meter_id`).
- Migrations under `backend/migrations/` (10 active: init, hypertables, CAGGs, retention, emission factors, idempotency, readings dedup, readings seq, RLS enable, audit lock).
- Goose (pressly/goose v3) is the migration runner per ADR-0005. Forward-only in production. CAGG-aware (NO TRANSACTION).
- pgx/v5 only — no ORM (REJ-35). Parameterised queries. Pool 25/2/30m. Per-Tx 5s timeout.
- Field-level encryption on `readings.raw_payload`: not yet wired (Phase F Sprint S11 deliverable). KMS key wrapping not yet wired.

**Score:** 4/5. Solid time-series schema. Gap: field-level encryption, signed reading provenance, signed audit chain, corrections-overlay table.

**Doctrine-rule references:** 33, 89, 90, 91, 95, 96, 100, 101, 102, 103, 104, 105, 169, 172.

### 2.3 Security posture

- JWT HS256 with `kid` claim (Rule 170). Validation pinning in `internal/security/jwt.go`. KID rotation runbook (`docs/JWT-ROTATION.md`) + ADR-0016. KID rotation automated workflow not yet shipped (Phase E Sprint S6).
- RBAC: middleware `RequirePermission(...)` (`internal/security/rbac.go`). Permission registry. Five roles (admin, manager, operator, auditor, readonly).
- RLS: migration `00098_rls_enable.go` enables tenant RLS. Conformance test `tests/security/rls_test.go` partial.
- Idempotency: middleware in `internal/handlers/idempotency.go`. `idempotency_keys` Timescale hypertable + 24h retention.
- Sentinel detection: refusal at boot (`config.go:176-194`). Conformance test scaffold.
- Audit log: append-only schema. Immutability enforced at role-revoke level in `00099_audit_lock.sql`. Chained-HMAC signatures *not yet shipped* — Phase F Sprint S11 deliverable.
- TLS 1.3 production. HSTS preload. CSP / X-Frame-Options / X-Content-Type-Options. Body limit 4MB / 16MB ingest. Constant-time compare on Pulse webhook (REJ-20 mitigation).
- Pen-test: not yet performed. Phase J Sprint S22 deliverable.

**Score:** 3.5/5. Defence-in-depth architecturally in place. Gap: signed audit chain, field-level encryption wiring, signed reading provenance, pen-test.

**Doctrine-rule references:** 8, 19, 20, 39, 62, 169, 170, 172, 173, 174, 178, 184, 188.

### 2.4 Compliance surface

- Italian-compliance citation map in `docs/ITALIAN-COMPLIANCE.md`.
- CSRD coverage: `docs/COMPLIANCE/CSRD.md` (placeholder; Phase E Sprint S6 hardens).
- D.Lgs. 102/2014 audit: `docs/COMPLIANCE/Dlgs102.md` (placeholder).
- GDPR: `docs/COMPLIANCE/GDPR.md` carries lawful-basis enumeration + DSAR-workflow placeholder + Art. 28 DPA reference. DSAR endpoint not yet wired (Phase F Sprint S11).
- NIS2 D.Lgs. 138/2024: `docs/COMPLIANCE/NIS2.md` carries the soggetto-importante classification + 24h/72h notification template references.
- AgID Qualificazione Cloud per la PA: not yet pursued (Phase J Sprint S21–S22 deliverable).
- ISPRA factor versioning: emission_factors table in place (migration 0005). Annual update workflow not yet wired.
- ENEA / GSE submission stubs: not yet wired (Phase H Sprint S15 deliverable).

**Score:** 3/5. Compliance surface declared. Gap: validation against formal specs, exportable evidence pack, automated factor refresh, ENEA/GSE submission flows, AgID dossier.

**Doctrine-rule references:** 8, 63, 89, 95, 108, 129–148, 164, 165, 167, 184.

### 2.5 OT integration depth

- Modbus TCP ingestor in `internal/services/modbus_ingestor.go` (159 LoC). Goburrow/modbus client. Polling cadence configurable. Slave-IDs configurable. Simulator at `backend/cmd/simulator/main.go` for development.
- Modbus RTU support via the same library (serial-line via socat in dev). Production wiring partial.
- M-Bus stub at `internal/services/mbus_ingestor.go` (91 LoC) — frame parser; no live-line wiring.
- SunSpec profile in `internal/services/sunspec_profile.go` (124 LoC). PV-inverter model coverage partial.
- Pulse webhook in `internal/services/pulse_ingestor.go` (131 LoC). HMAC-SHA256 verify with constant-time compare per REJ-20.
- OCPP client in `internal/services/ocpp_client.go` (292 LoC). Version 1.6 support; 2.0.1 partial. Charging-station onboarding stub.
- IEC 61850, OPC UA, MQTT Sparkplug B, BACnet, EtherCAT, PROFINET, KNX: not yet covered (Phase G Sprint S12–S14 deliverables for the most-requested subset).
- Edge gateway: not yet shipped — `cmd/simulator` is the development stand-in. Phase G Sprint S14 ships the real edge gateway.
- Time-source discipline: NTP enforced in dev; GPS sync not yet wired. Phase G deliverable.

**Score:** 3/5. Major-protocols-covered. Gap: real edge gateway, IEC 61850 + OPC UA + Sparkplug B, time-source GPS, network-segmentation-aware Topology D.

**Doctrine-rule references:** 109–128, 169, 173.

### 2.6 Reporting & analytics

- `report_generator.go` (525 LoC, REJ-11): god service producing ESRS E1, Piano 5.0, Conto Termico, TEE, audit DLgs102, monthly consumption, CO2 footprint reports. Decomposition into per-Builder packages is Phase E Sprint S6 deliverable.
- Carbon calculator in `internal/services/carbon_calculator.go` (185 LoC). Versioned-factor lookup. Aggregation correctness validated by `tests/property/aggregate_invariants_test.go` (initial coverage; expansion Phase F Sprint S9).
- Frontend report-flow: route `/reports/+page.svelte`. Trigger-and-download UX. PDF rendering server-side with `gofpdf`-equivalent (Phase F Sprint S11 hardens with PDF/A-2b).
- Determinism property: not yet enforced. Builders are not yet pure functions (some access `time.Now()`). Phase F Sprint S9 deliverable.
- Replay tests against fixtures: not yet shipped. Phase F Sprint S9.

**Score:** 3/5. Reports work. Gap: pure builders, deterministic serialisation, replay tests, formal-spec validation, signed reports, signed provenance, EFRAG-taxonomy mapping, GSE-XSD validation.

**Doctrine-rule references:** 89–108, 129–148, 169, 144.

### 2.7 Observability & operability

- OTel SDK + OTLP gRPC exporter. Sample ratio 0.1 production / 1.0 dev (Rule 18). Span coverage: HTTP handler (otelfiber), pgx (tracelog QueryTracer), outbound HTTP (otelhttp), ingestor poll (manual).
- Zap structured JSON logs. Mandatory fields per Rule 7. Redactor for password / token / secret / authorisation.
- Prometheus exposition at `/api/internal/metrics`. Custom metrics in `internal/metrics/metrics.go` with cardinality budgets.
- Grafana dashboards: `energy-overview.json`, `carbon-dashboard.json` provisioned. Operator-team-friendly. Engagement-team-specific dashboards added per engagement.
- Health endpoints: `/api/health` (degraded-tolerant), `/api/ready` (strict), `/api/live` (always 200 once boot complete). Health envelope per Rule 6.
- Alertmanager rules: error rate > 1%, ingestion lag > 5min, Timescale connection pool > 80%, OTel exporter backlog > 10s, scheduled compression failed, CAGG refresh lag > 2× interval. Runbook annotations on each.
- Runbooks: 11 on disk (db-outage, ingestor-crash-loop, grafana-down, cert-rotation, secret-rotation, jwt-secret-rotation, capacity-spike, region-failover, tenant-data-leak, pulse-webhook-flood, cost-audit). On-call runbook drill cadence quarterly.

**Score:** 4.5/5. Observability is mature. Gap: pack-loader instrumentation depth, engagement-health dashboard, pack-health-aggregated health envelope.

**Doctrine-rule references:** 7, 18, 23, 40, 74, 85, 124, 156, 161, 199.

### 2.8 Documentation & doctrine

- 50+ docs in `docs/`, including TEAM-CHARTER, RACI, SECOPS-CHARTER, PLATFORM-INITIATIVE-WORKFLOW, ARCHITECTURE, API, LAYERS, MODUS_OPERANDI, THREAT-MODEL, RISK-REGISTER, SUPPLY-CHAIN, ITALIAN-COMPLIANCE, SLO, SLI-CATALOG, OBSERVABILITY, RUNBOOK, INCIDENT-RESPONSE, ON-CALL, ROLLBACK, RELIABILITY-MODEL, CHAOS-PLAN, CHAOS-LOG, CAPACITY, COST-MODEL, PIPELINE-MAP, PENTEST-CADENCE, JWT-ROTATION, MTLS-PLAN, TRUST-BOUNDARIES, A11Y, CONTACTS, CONTRIBUTING, DX, DEBUG, TROUBLESHOOTING, PLATFORM-PLAYBOOK, PLATFORM-DEFAULTS, QUALITY-BAR, SERVICE-CATALOG, EXTENSION-POINTS, ABSTRACTION-LEDGER, SCHEMA-EVOLUTION, API-VERSIONING, FEATURE-FLAGS, CLI-CONTRACT.
- ADRs: 20 numbered ADRs covering platform doctrine, multi-tenant RLS, ESO vs Vault, GitOps with ArgoCD, migration tool (goose), observability OTel+Prometheus, residency, API versioning, breakers, hypertable space partitioning, RLS defence-in-depth, validator, oapi-codegen design-first, async report (asynq), bounded ingest, JWT KID rotation, Cosign keyless OIDC, SLSA L2-now-L3-plan, Falco vs Tetragon, cert-manager vs SPIRE.
- REJECTED.md: 35 anti-patterns rejected (REJ-01 through REJ-35) with rule citation, alternative, residual risk, review date.
- Doctrine: 60-rule doctrine adopted via ADR-0001. The plan file `~/.claude/plans/my-brother-i-would-flickering-coral.md` is on a different machine; the inferred doctrine is reconstructed in `docs/DOCTRINE.md` (210 rules total, including 140+ new rules in groups 4–10).

**Score:** 5/5. Documentation surface is the largest moat against any competitor. Continuing investment expected.

**Doctrine-rule references:** 27, 47, 67, 209, 210.

### 2.9 Aggregate score and gap summary

The current state averages 3.7/5 across the eight axes — which translates to "Mission II PASS, but not yet Charter-conformant." The named gaps are:

1. **SaaS framing** — every commercial document, ADR-0001 references, MODUS_OPERANDI, COST-MODEL, THREAT-MODEL §2 carries SaaS vocabulary. Rewrite in Phase E Sprint S5.
2. **Pack extraction** — Italian-flagship code lives in `internal/services/` rather than `packs/region/it/` + Factor / Report / Protocol Packs. Phase E Sprint S6–S7.
3. **Pack contracts** — `Ingestor`, `FactorSource`, `Builder`, `IdentityProvider`, `RegionProfile` interfaces are not yet formalised. Phase E Sprint S5.
4. **Engagement-fork model** — `template-version.txt`, `engagements/<client>/`, upstream-sync runbook do not yet exist. Phase E Sprint S6.
5. **White-label readiness** — `config/branding.yaml` and the conformance test against hard-coded brand strings do not yet exist. Phase E Sprint S6.
6. **Conformance suite** — universal-invariant tests, core-pack-separation tests, replay tests, manifest-lock tests are not yet shipped. Phase E + F.
7. **Audit-grade reproducibility** — chained-HMAC audit log, signed reading provenance with per-meter HMAC, deterministic Builders, replay tests are not yet shipped. Phase F.
8. **Real edge gateway** — `cmd/simulator` is the dev stand-in; the production edge gateway (with disk-backed buffer + GPS time + signed readings) is not yet shipped. Phase G.
9. **Additional Protocol Packs** — IEC 61850 + OPC UA + MQTT Sparkplug B + BACnet not yet covered. Phase G.
10. **Formal-spec validation** — EFRAG XBRL taxonomy + GSE XSD + ENEA XSD validation not yet wired. Phase H.
11. **Additional Region Packs** — German Region Pack + UK Region Pack not yet shipped. Phase H.
12. **AI/ML surface** — consumption forecaster, layered anomaly detection (only Layer 1 partial), drift detection, model cards, model registry not yet shipped. Phase I.
13. **Hardening** — SOC 2 / ISO 27001 / AgID dossier / pen-test / post-quantum-crypto plan not yet shipped. Phase J.

The plan addresses each gap explicitly, with a phase + sprint + deliverable + exit-criterion + doctrine-rule citation. The gaps are not insurmountable — they are the predictable surface of a Mission-II-PASS project that needs to ship its v1.0.0.

---

## 3. North Star: positioning, success criteria, non-goals

### 3.1 Positioning

GreenMetrics v1.0.0 is **the regulator-grade modular template for industrial energy management and sustainability reporting**. It is delivered as an engagement: a contracted client receives the template, the engagement specialises it via Pack composition + customisation sprint, and the deployment is operated by the client / co-managed / fully-managed depending on tier. The buyer profile is the energy-intensive Italian (later European) industrial SME that needs CSRD ESRS E1 / Piano 5.0 / Conto Termico / TEE / audit 102/2014 capability and is not adequately served by the four credible competitor classes:

1. **Tier 1 enterprise OT/automation platforms** (Schneider EcoStruxure Resource Advisor, Siemens Sinergy / Sinalytics, ABB Ability EM, Honeywell Forge, Emerson Ovation Green, Rockwell FactoryTalk Energy Manager) — six-figure entry, six-month integration, regulator-as-localisation-afterthought. We're audit-grade by default, integration-shaped to the customer in 8–14 weeks, and Italian-regulation-native.
2. **ESG SaaS platforms** (Watershed, Persefoni, Sweep, Greenly, Plan A, Sphera, Cority, Wolters Kluwer Enablon, Workiva, Diligent ESG) — heavy dashboarding, light OT integration, generic CSRD support, English-first UX, US-cloud-first residency. We're OT-native, Italian-residency-by-default, deterministic-builders.
3. **Hyperscaler ESG platforms** (IBM Envizi, Salesforce Net Zero Cloud, Microsoft Sustainability Manager, SAP Green Ledger / SAP SFM, Google Carbon Footprint) — bundled with the hyperscaler relationship, generic, lock-in tax, weak Italian-specific muscle. We're vendor-neutral, Italian-flagship, open-data via SQL access.
4. **Italian utility-bundled / ESCO platforms** (Enel X Way Business Energy Manager, A2A Smart City, Hera Smart Services, Edison EnergEnvision / Edison Next, Duferco Energia, MAPS Energy, SunCity) — utility-conflict, energy-services-bundled, Piano-5.0-ESCO-conflict-of-interest. We're vendor-neutral and willing to certify Piano 5.0 attestazioni produced for investments we did not sell.

The wedge is regulator-grade engineering. We make claims (bit-perfect reproducibility, signed audit chain, formal-spec validation, signed reading provenance, Pack-versioned thresholds, EGE counter-signature workflow, single-command audit-evidence-pack) that no credible competitor makes and cannot quickly acquire. The competitive brief in `docs/COMPETITIVE-BRIEF.md` enumerates the matrix.

### 3.2 Success criteria

The plan succeeds if at the end of Sprint S22 (≈ March 2027) the following are true:

1. **v1.0.0 released.** Template is at `github.com/greenmetrics/template` with the canonical migration from the current private repository complete. Tag `v1.0.0` is signed (Cosign keyless). Release notes published. License decision (BUSL / SSPL / AGPLv3 / proprietary-perpetual) recorded in an ADR.
2. **Five engagements shipped end-to-end through Phase 5 Handover.** Five distinct clients have run the lifecycle, all hit Phase 5 readiness on time, all are paying their annual maintenance, all have generalised at least one engagement-specific feature back upstream as a Pack contribution.
3. **210 doctrine rules enforced.** Every rule has either a mechanical conformance test or a named CODEOWNERS reviewer. The conformance suite runs in under 8 minutes locally per Rule 24.
4. **Six Region Packs shipped.** `packs/region/{it, de, fr, es, gb, at}/` are all charter-conformant (Pack-acceptance checklist green) and have at least one engagement on each.
5. **Twelve Protocol Packs shipped.** Modbus TCP, Modbus RTU, M-Bus, SunSpec, Pulse, OCPP 1.6, OCPP 2.0.1, IEC 61850, OPC UA, MQTT Sparkplug B, BACnet, IEC 62056-21 — all charter-conformant, all with simulators, all with device-profile catalogues populated for at least the top-3 vendor in their class.
6. **Eight Report Packs shipped.** ESRS E1 (CSRD), Piano 5.0, Conto Termico 2.0, Certificati Bianchi TEE, audit 102/2014, monthly consumption, CO2 footprint, plus one of {SECR UK, GHG Protocol Corporate Standard, ISO 14064-1, TCFD, IFRS S1/S2, SEC Climate Disclosure}.
7. **Six Factor Packs shipped.** ISPRA Italia, GSE renewable shares, Terna national mix, AIB residual mix, UK BEIS / DEFRA, EPA US eGRID.
8. **Three Identity Packs shipped.** Local-DB (default), SAML 2.0 (CyberArk / Okta / Azure AD / Keycloak), OIDC (Auth0 / Okta OIDC / Google).
9. **Audit-grade reproducibility property holds.** Every Builder is pure, every report carries a signed provenance bundle, every reading is signed at ingestion with per-meter HMAC, the audit log is chained-HMAC-signed, the manifest lock is signed and verified at admission. The bit-perfect-replay test is green for every Builder against pinned fixtures.
10. **Regulatory-evidence pack is single-command.** `task audit-evidence-pack` produces a complete, signed, schema-validated zip in ≤ 15 minutes for a 1-year audit window.
11. **AI/ML surface is observable and reproducible.** The consumption forecaster ships with model cards, training-data lineage, drift detection, and accuracy monitoring. The forecaster beats the seasonal baseline by ≥ 10% MAPE. Anomaly detection is Layer 1 + 2 + 3 deployed.
12. **SOC 2 Type II readiness.** The audit-readiness checklist is green. A nominated auditor has begun the Type I → Type II progression. Italian-AGID Qualificazione Cloud per la PA dossier is filed for the Aruba topology.
13. **Pen-test cleared.** Annual third-party penetration test by an OSCP-certified firm completed; findings remediated against the Rule-59 SLA.
14. **Engagement-margin target hit.** T1 engagements ship at ≥ 65% gross margin, T2 at ≥ 55%, T3 at ≥ 45%. Annual-maintenance attach rate ≥ 90%.

A subset of these (notably 12 and 13) may slip into S23–S24 for calendar reasons; the plan accepts the slip as long as the rest hold. Slips beyond S24 trigger a re-plan ADR.

### 3.3 Non-goals (explicit)

The following are explicitly *out of scope* for the plan period. Pursuing them is a Rule-26 / Rule-46 / Rule-66 rejection unless a dedicated supersession ADR is filed.

- **A SaaS-edition of the template hosted by Macena.** Charter §14.2 defers this to year 2. Pursuing it now violates the charter.
- **Multi-region active-active deployment.** REJ-09 retains; revisit Sprint S23+ if engagement-pipeline requests it.
- **Service-mesh adoption.** REJ-01 retains.
- **Microservice split of the backend.** REJ-05 retains.
- **GraphQL gateway.** REJ-32 retains.
- **Custom Go web framework.** REJ-34 retains.
- **ORM (gorm / bun) on Timescale.** REJ-35 retains.
- **OPA-everywhere request-time PEP.** REJ-10 retains.
- **Vault adoption in Phase 1 of any engagement.** REJ-30 retains; Vault may appear in Topology B (Aruba) where AWS Secrets Manager is unavailable.
- **A mobile native app for ingestion or operator workflows.** Out of scope for the plan period; Phase J explicitly does not address it.
- **A full-blown CRM / billing / invoicing surface.** Not the template's job; we partner.
- **An `/api/v2` major version.** v1 is frozen-and-additive for 18 months minimum (Charter §12.1).
- **Removing the Italian-flagship Pack.** Charter §3.2 + Rule 88 retain it as the reference.
- **Open-source release before v1.0.0 is locked.** Charter §14.1 defers the licensing decision to v1.0.0 launch.

The non-goals are not a value judgement; they're a scope discipline. Each one is potentially valuable; none of them is in this plan.

---

## 4. The architectural pivot — code-level deltas

The pivot from "multi-tenant SaaS for Italian SMEs" to "modular template for engagement deployments" is not a marketing rebrand. It changes the code in specific ways. This section enumerates those deltas at the resolution a developer can implement against. Each delta has a doctrine-rule citation, an estimated diff size, an owning sprint, and an ADR trigger.

### 4.1 Pack-loader infrastructure

**Current:** `cmd/server/main.go` constructs ingestors, factor sources, and report builders directly via concrete-type constructors. The list of available implementations is hard-coded.

**Target:** `internal/packs/loader.go` reads `config/required-packs.yaml`, walks `packs/`, instantiates each Pack, calls `Pack.Init`, calls `Pack.Register(reg)` to receive contributions, and writes `manifest.lock.json` (Cosign-signed) at successful boot.

**Code surface:** new `internal/packs/{pack.go, manifest.go, contracts.go, registry.go, loader.go, lock.go}` (~1500 LoC). New `config/required-packs.yaml`. New `docs/contracts/pack-manifest.schema.json`. ~400 LoC of conformance tests in `tests/packs/`.

**Doctrine rules:** 69, 70, 71, 72, 73, 75, 76, 83, 85.

**ADR triggers:** new ADR-0021 "Pack loader and contract architecture." new ADR-0022 "Manifest-lock signature and admission verification."

**Sprint:** Phase E Sprint S5.

### 4.2 Pack-contract interface formalisation

**Current:** `internal/services/ingestor.go` defines an informal `Ingestor` interface (used by `IngestorRunner` / AB-01). `internal/domain/reporting/builder.go` *does not yet exist*; report builders are concrete methods on `report_generator.go`. `FactorSource` is implicit. `IdentityProvider` is implicit.

**Target:** explicit Go interfaces at `internal/domain/protocol/ingestor.go`, `internal/domain/reporting/builder.go`, `internal/domain/emissions/factor_source.go`, `internal/domain/identity/provider.go`, `internal/domain/region/profile.go`, each carrying full Godoc, an example test, and a contract version constant.

**Code surface:** five interface files (~400 LoC). Five doc-example tests (~250 LoC). Per-kind contract versioning constants (~50 LoC). Conformance: `tests/packs/pack_contract_compliance_test.go`.

**Doctrine rules:** 14, 34, 70, 71, 86, 87, 109, 129, 130.

**ADR triggers:** new ADR-0023 "Pack-contract interfaces and versioning."

**Sprint:** Phase E Sprint S5.

### 4.3 Italian-flagship Pack extraction

**Current:** Italian-specific code is in `internal/services/{report_generator.go, carbon_calculator.go, smart_meter_client.go, modbus_ingestor.go, mbus_ingestor.go, sunspec_profile.go, pulse_ingestor.go, ocpp_client.go, alert_engine.go, energy_analytics.go, audit_client.go, ingest_pipeline.go, ingestor_runner.go}`. ISPRA / GSE / Terna factor lookup is in `carbon_calculator.go`. ESRS E1 / Piano 5.0 / Conto Termico / TEE / audit-DLgs102 builders are in `report_generator.go`.

**Target:** the Italian Region Pack at `packs/region/it/` carries all Italian-specific defaults (timezone Europe/Rome, locale it_IT.UTF-8, currency EUR with comma-decimal, ARERA / Garante / GDPR / NIS2 invariants). The Italian Factor Packs at `packs/factor/{ispra, terna, gse, aib}/` carry the factor sources. The Italian Report Packs at `packs/report/{esrs_e1, piano_5_0, conto_termico, tee, audit_dlgs102, monthly_consumption, co2_footprint}/` carry the Builders. The Italian Protocol Packs at `packs/protocol/{modbus_tcp, modbus_rtu, mbus, sunspec, pulse, ocpp_1_6, ocpp_2_0_1}/` carry the ingestors.

**Code surface:** ~3000 LoC moves from `internal/services/` to `packs/`. Per-Pack manifest.yaml + Pack-charter + Pack-tests (~2000 LoC of new). Per-Pack README. The `internal/services/` post-extraction shrinks to orchestration only (~1500 LoC remaining). Several REJ-11-flagged god services are decomposed in the same migration.

**Doctrine rules:** 32, 69, 87, 88, 109, 129, 130.

**ADR triggers:** new ADR-0024 "Italian Region Pack and Italian-flagship Pack catalogue." new ADR-0025 "Decomposition of `internal/services/report_generator.go` into per-Builder Packs."

**Sprint:** Phase E Sprint S6 (Italian Region Pack + Factor Packs + Report Packs); Phase E Sprint S7 (Italian Protocol Packs).

### 4.4 White-label / branding configuration

**Current:** brand strings are hard-coded throughout the frontend, the PDF cover-letters, the email templates, the OpenAPI `info.contact`, the config defaults, the docker-compose, the runbooks.

**Target:** `config/branding.yaml` carries `product_name`, `legal_entity`, `support_contact`, `logo_*`, `theme_*`, `footer_text`, `pdf_cover_template`. The frontend reads at build via a Vite plugin. The PDF renderer reads at render. Email templates read at render. The conformance test `tests/conformance/no_hardcoded_brand_test.go` verifies no rendered surface has hard-coded brand outside `config/branding.yaml` defaults.

**Code surface:** new `config/branding.yaml` + schema (~50 LoC). Vite plugin to inject (~80 LoC). PDF-renderer config-read (~60 LoC). Email-template variable substitution (~80 LoC). ~40 source-code edits to remove hard-coded "GreenMetrics" outside the defaults. Conformance test (~150 LoC).

**Doctrine rules:** 81.

**ADR triggers:** new ADR-0026 "White-label configuration via `config/branding.yaml`."

**Sprint:** Phase E Sprint S6.

### 4.5 Engagement-fork model

**Current:** the repository is the canonical source. There is no `template-version.txt`. There is no `engagements/` directory. There is no upstream-sync runbook.

**Target:** the canonical template is at `github.com/greenmetrics/template` (post-v1.0.0). Engagement forks live at `github.com/<engagement-org>/<engagement-id>-greenmetrics`. Each fork carries `template-version.txt` and `engagements/<client>/` overlays. The upstream-sync runbook at `docs/runbooks/upstream-sync.md` describes the quarterly merge workflow.

**Code surface:** `docs/runbooks/upstream-sync.md` (~150 LoC of dense runbook). `docs/runbooks/engagement-fork-bootstrap.md` (~120 LoC). Empty `engagements/` directory in template main with README. The `template-version.txt` template. New `task engagement-bootstrap` task helper.

**Doctrine rules:** 77, 78, 79, 80, 150.

**ADR triggers:** new ADR-0027 "Engagement-fork topology and upstream-sync discipline."

**Sprint:** Phase E Sprint S6.

### 4.6 SaaS-vocabulary doc rewrite

**Current:** MODUS_OPERANDI v1, COST-MODEL, THREAT-MODEL §2, ADR-0001 ("regulated-industry SaaS"), CSRD compliance doc, NIS2 compliance doc, GDPR doc, MODUS_OPERANDI all carry SaaS framing.

**Target:** MODUS_OPERANDI v2 (engagement playbook). COST-MODEL v2 (engagement margin replaces SaaS gross margin; CAC/LTV/churn replaced by engagement-margin / time-to-customisation / template-fit-score / net-engagement-value). THREAT-MODEL §2 ("single-tenant by default, multi-tenant by configuration"). ADR-0001 annotated to clarify that "SaaS" is replaced by "engagement template" with Charter §2 the canonical reframe. CSRD / NIS2 / GDPR docs annotated with the engagement-context.

**Code surface:** ~3000 LoC rewrite across the named docs. No code changes; all documentation.

**Doctrine rules:** 209 (the doc-rewrite is itself a Rule-209-grade evidence). Charter §2.

**ADR triggers:** ADR-0028 "MODUS_OPERANDI v2 — engagement playbook." Charter-supersession annotation in ADR-0001.

**Sprint:** Phase E Sprint S5 (MODUS_OPERANDI + ADR-0001 annotation); Phase E Sprint S6 (other docs).

### 4.7 Conformance-suite expansion

**Current:** `backend/tests/conformance/` directory exists. Population is partial. Universal-invariant tests (Rules 1–8) are not yet shipped. Core/Pack separation test is not yet shipped. Replay tests are not yet shipped. Manifest-lock test is not yet shipped.

**Target:** `tests/conformance/` is populated with: `no_float_money_test.go`, `rfc3339utc_test.go`, `uuidv4_test.go`, `rfc7807_test.go`, `cloudevents_test.go`, `health_test.go`, `log_format_test.go`, `core_pack_separation_test.go`, `no_hardcoded_brand_test.go`, `pack_health_test.go`, `report_determinism_test.go`, `factor_temporal_test.go`, `aggregate_invariants_test.go`, `timebucket_test.go`, `tenant_timezone_test.go`, `report_byte_identity_test.go`, `report_period_test.go`, `pdf_pdfa_test.go`, `lineage_endpoint_test.go`, `crypto_shredding_test.go`, `audit_chain_test.go`, `signed_reading_test.go`, `forecast_determinism_test.go`, `inference_graceful_test.go`, `runbook_link_test.go`, `explanation_present_test.go`, `pack_health_aggregated_test.go`, `manifest_lock_test.go`, `report_provenance_test.go`, `conto_termico_xsd_test.go`. Plus `tests/static/` for compile-time invariants (`no_float_in_regulatory_test.go`, `no_fiber_in_domain_test.go`, `no_aws_access_key_test.go`, `no_math_rand_in_security_test.go`, `no_real_pii_test.go`).

**Code surface:** ~30 conformance tests at ~100 LoC each = ~3000 LoC of test code. ~5 static tests at ~80 LoC each = ~400 LoC. Per-test doc comment naming the rule it enforces.

**Doctrine rules:** 1–8 (universal); 24, 44, 64 (verification); 89–108 (audit-grade); plus rule-specific.

**ADR triggers:** none individually; the rule references the existing ADR-0001.

**Sprint:** Phase E Sprint S5 (universal + Pack); Phase F Sprint S9 (audit-grade reproducibility); Phase F Sprint S11 (signed audit chain + reading provenance); Phase G Sprint S13 (per-protocol); Phase H Sprint S15 (regulatory-pack); Phase I Sprint S18 (AI/ML).

### 4.8 Audit-log chained-HMAC

**Current:** `audit_log` is append-only at the role level. No chained-HMAC.

**Target:** `audit_log` schema augmented with `prev_hmac` and `row_hmac` columns. An insert trigger computes `row_hmac = HMAC-SHA256(audit_chain_key, prev_hmac || row_canonical_json)`. The `audit_chain_key` lives in KMS and rotates annually with key-versioning. An hourly checkpoint job publishes a signed checkpoint to the WORM-mirror. Conformance test `tests/security/audit_chain_test.go` verifies chain integrity on a rolling window.

**Code surface:** migration adding the columns + trigger (~150 LoC). KMS key provisioning in Terraform (~80 LoC). Hourly checkpoint job in `internal/jobs/audit_checkpoint.go` (~120 LoC). Conformance test (~150 LoC).

**Doctrine rules:** 62, 169.

**ADR triggers:** new ADR-0029 "Chained-HMAC audit-log integrity."

**Sprint:** Phase F Sprint S11.

### 4.9 Reading provenance with per-meter HMAC

**Current:** readings are unsigned. No per-meter HMAC keys.

**Target:** every reading carries `meter_hmac` over `(meter_id, channel_id, ts_ns, value_int_micro_unit)`. Per-meter HMAC keys are provisioned at meter onboarding, stored in KMS (or in sealed enclave on the edge gateway). Ingestion verifies; failed-verify readings carry `quality_code = 'unsigned'` and are excluded from regulatory builders. Migration adds the column.

**Code surface:** migration adding `meter_hmac bytea` (~40 LoC). KMS key issuance flow (~150 LoC). Edge-gateway sealed-key store (~200 LoC, Phase G). Backend verification in `internal/services/ingest_pipeline.go` (~80 LoC).

**Doctrine rules:** 105, 173.

**ADR triggers:** new ADR-0030 "Per-meter HMAC reading signatures."

**Sprint:** Phase F Sprint S11.

### 4.10 Field-level encryption

**Current:** `readings.raw_payload` is `bytea`-stored unencrypted. Per-tenant DEK is not yet provisioned.

**Target:** `readings.raw_payload` is AES-256-GCM-encrypted with a per-tenant DEK. The DEK is wrapped by a KMS master key (MEK). The DEK is regional and rotates with versioned `dek_id`. Crypto-shredding on tenant deletion (Rule 184) deletes the DEK.

**Code surface:** migration adding `dek_id`, `dek_version`, `payload_encrypted` columns (~80 LoC). DEK provisioning at tenant creation (~100 LoC). Encryption / decryption helpers in `internal/security/crypto.go` (~150 LoC). Decryption integration in repository read paths (~80 LoC).

**Doctrine rules:** 39, 104, 172, 184.

**ADR triggers:** new ADR-0031 "Field-level encryption with per-tenant DEK and KMS-wrapped MEK."

**Sprint:** Phase F Sprint S11.

### 4.11 Deterministic Builder framework

**Current:** Builders in `report_generator.go` access `time.Now()`, depend on map iteration order, may have nondeterministic float arithmetic.

**Target:** Builders are pure functions per Rule 91. The custom golangci-lint rule blocks `time`, `os`, `crypto/rand`, `math/rand`, `internal/services/` imports inside `internal/domain/reporting/<builder>/`. Deterministic serialisation via `internal/domain/serialise/` (sorted JSON keys; PDF/A-2b producer string; preserved XML element order). Replay test `tests/conformance/builder_replay_test.go` runs every Builder against pinned fixtures.

**Code surface:** Builder refactor (~600 LoC across the seven Builders). Custom golangci-lint rule (~100 LoC). Deterministic-serialise helpers (~200 LoC). Replay test (~250 LoC).

**Doctrine rules:** 89, 91, 94, 141.

**ADR triggers:** new ADR-0032 "Deterministic Builder framework." Charter-supersession annotation in REJ-11.

**Sprint:** Phase F Sprint S9.

### 4.12 Per-tenant timezone

**Current:** timezone is implicitly Europe/Rome (Italian Region Pack default).

**Target:** `tenants.timezone` column is per-tenant explicit. Region Pack supplies the default. All time-bucket operations and human-readable rendering reads the column. Conformance test verifies.

**Code surface:** migration adding `tenants.timezone` (~30 LoC). Region Pack default population (~40 LoC). Time-bucket integration (~60 LoC). Conformance test (~100 LoC).

**Doctrine rules:** 2, 93, 101.

**ADR triggers:** none (within existing ADR-0007 residency).

**Sprint:** Phase F Sprint S10.

### 4.13 Corrections-overlay table

**Current:** raw readings are immutable de-facto (no UPDATE happens), but no formal corrections workflow.

**Target:** new `corrections` table records per-reading corrections (auditor-applied, vendor-replacement, manual override). Builders read `readings JOIN corrections`. The original reading remains visible.

**Code surface:** migration adding `corrections` table (~80 LoC). API endpoint to create a correction (~120 LoC). RBAC role `auditor` allowed; everyone else denied (~30 LoC). Builder query update (~80 LoC).

**Doctrine rules:** 96.

**ADR triggers:** new ADR-0033 "Corrections-overlay table for raw-reading immutability."

**Sprint:** Phase F Sprint S10.

### 4.14 OT edge gateway

**Current:** `cmd/simulator` is the development stand-in. Production deploys do not yet have a real edge gateway shipping signed readings + disk-backed buffer + GPS time.

**Target:** new `cmd/edge-gateway` carries the real edge gateway: protocol drivers loaded from per-Pack drivers, disk-backed buffer (24h default), NTP/GPS time discipline, per-meter HMAC signing at ingestion, mTLS to backend, observability + health surface.

**Code surface:** new `cmd/edge-gateway/` (~1500 LoC). Per-Pack driver wiring. Bootstrap script. Cross-compile Dockerfile (ARM64 + AMD64 + ARMv7 for Raspberry-Pi-class targets). Distroless nonroot image. Cosign-signed.

**Doctrine rules:** 109, 111, 112, 173.

**ADR triggers:** new ADR-0034 "OT edge gateway architecture."

**Sprint:** Phase G Sprint S14.

### 4.15 IEC 61850 Protocol Pack

**Current:** not yet covered.

**Target:** `packs/protocol/iec_61850/` ships an MMS client speaking IEC 61850-8-1 (and substation-automation-relevant parts). Coverage focuses on the LN-class set most common in industrial substations (XCBR, XSWI, MMXU, MMTR, LLN0). Per-IED profile catalogue.

**Code surface:** ~1200 LoC in the Pack. Per-IED profile schema. Simulator (~600 LoC). Conformance + integration tests.

**Doctrine rules:** 109, 110.

**ADR triggers:** new ADR-0035 "IEC 61850 MMS Protocol Pack."

**Sprint:** Phase G Sprint S12.

### 4.16 OPC UA Protocol Pack

**Current:** not yet covered.

**Target:** `packs/protocol/opc_ua/` ships an OPC UA client speaking the binary-encoded protocol with security policy `Basic256Sha256` minimum. Coverage of the OPC UA companion specs most common in industrial automation (DA, AC, HA, PROSYS DI). Per-server-config catalogue.

**Code surface:** ~1500 LoC in the Pack. Per-server-config schema. Simulator (~700 LoC). Tests.

**Doctrine rules:** 109, 110.

**ADR triggers:** new ADR-0036 "OPC UA Protocol Pack."

**Sprint:** Phase G Sprint S12.

### 4.17 MQTT Sparkplug B Protocol Pack

**Current:** not yet covered.

**Target:** `packs/protocol/mqtt_sparkplug_b/` ships an MQTT client subscribing to Sparkplug-B-formatted topics (`spBv1.0/<group_id>/<message_type>/<edge_node_id>/<device_id>`) with NBIRTH / NDATA / NDEATH / DBIRTH / DDATA / DDEATH semantics. The Pack adds an MQTT-broker dependency to the deployment topology — ADR-0037 documents.

**Code surface:** ~1100 LoC in the Pack. Sparkplug B payload protobuf integration. Simulator (~500 LoC). Tests.

**Doctrine rules:** 109, 110.

**ADR triggers:** new ADR-0037 "MQTT Sparkplug B Protocol Pack and broker dependency."

**Sprint:** Phase G Sprint S13.

### 4.18 BACnet Protocol Pack

**Current:** not yet covered.

**Target:** `packs/protocol/bacnet/` ships a BACnet/IP client. Coverage of `analogInput`, `analogOutput`, `analogValue`, `binaryInput`, `binaryValue` object types. Per-device profile catalogue.

**Code surface:** ~900 LoC. Simulator (~400 LoC). Tests.

**Doctrine rules:** 109.

**ADR triggers:** new ADR-0038 "BACnet/IP Protocol Pack."

**Sprint:** Phase G Sprint S13.

### 4.19 EFRAG XBRL taxonomy mapping

**Current:** ESRS E1 Builder produces JSON with English-named keys. No XBRL output. No EFRAG taxonomy mapping.

**Target:** `packs/report/esrs_e1/taxonomy.yaml` declares the mapping from ESRS E1 disclosures to EFRAG XBRL taxonomy URIs. The Builder emits XBRL output (in addition to JSON and PDF). The XBRL is validated against the EFRAG-published taxonomy in CI.

**Code surface:** taxonomy.yaml (~200 LoC). XBRL emit (~400 LoC). Validation (~150 LoC). Conformance test.

**Doctrine rules:** 131, 134.

**ADR triggers:** new ADR-0039 "EFRAG XBRL ESRS taxonomy mapping."

**Sprint:** Phase H Sprint S15.

### 4.20 GSE XSD validation

**Current:** Conto Termico submission produces JSON; no GSE-XSD validation.

**Target:** `packs/report/conto_termico/xsd/` carries the GSE-published XSD. The Builder emits XML and validates against the XSD before persistence. Pre-submission validation prevents schema-noncompliant submissions.

**Code surface:** XSD checked in (~varies by GSE). Validator (~120 LoC). Conformance test.

**Doctrine rules:** 131, 135.

**ADR triggers:** new ADR-0040 "GSE XSD validation for regulatory submissions."

**Sprint:** Phase H Sprint S15.

### 4.21 ENEA XSD validation for audit 102/2014

**Current:** audit 102/2014 dossier produces JSON; ENEA portal expects XSD-shaped XML.

**Target:** `packs/report/audit_dlgs102/xsd/` carries the ENEA-published XSD. The Builder emits XML and validates. EGE counter-signature workflow: dossier transitions `draft → generated → ege_signing → signed → submitted`.

**Code surface:** XSD checked in. Validator. EGE counter-signature workflow + UI surface.

**Doctrine rules:** 131, 137.

**ADR triggers:** new ADR-0041 "ENEA XSD validation and EGE counter-signature workflow."

**Sprint:** Phase H Sprint S16.

### 4.22 Annual ISPRA factor refresh workflow

**Current:** ISPRA factors are manually loaded via SQL.

**Target:** `internal/jobs/ispra_factor_refresh.go` runs annually (April default), pulls the latest ISPRA emission-factor table from the ISPRA portal, validates against schema, generates a Pack release ADR, opens a PR for review. The job is sandboxed — it produces a draft factor table and a PR; merge is human.

**Code surface:** ~300 LoC of job. ISPRA URL + retrieval pattern. Schema validation. PR-creation tooling.

**Doctrine rules:** 90, 138.

**ADR triggers:** new ADR-0042 "Automated annual ISPRA factor refresh workflow."

**Sprint:** Phase F Sprint S11.

### 4.23 Identity Pack — SAML 2.0

**Current:** local-DB authentication only.

**Target:** `packs/identity/saml/` ships a SAML-2.0 SP integration. SP metadata generation. IdP metadata ingestion (CyberArk, Okta, Azure AD, Keycloak validated). Just-in-time provisioning of users from SAML attributes. Per-tenant IdP routing.

**Code surface:** ~1000 LoC. SP metadata. Just-in-time user creation. Tests.

**Doctrine rules:** 70, 87.

**ADR triggers:** new ADR-0043 "SAML 2.0 Identity Pack."

**Sprint:** Phase E Sprint S8.

### 4.24 Identity Pack — OIDC

**Current:** not yet shipped.

**Target:** `packs/identity/oidc/` ships an OIDC RP integration. Discovery via well-known URL. PKCE for confidential clients. Per-tenant OIDC provider routing.

**Code surface:** ~800 LoC. PKCE flow. Tests.

**Doctrine rules:** 70.

**ADR triggers:** new ADR-0044 "OIDC Identity Pack."

**Sprint:** Phase E Sprint S8.

### 4.25 Engagement-health dashboard

**Current:** no engagement-portfolio dashboard.

**Target:** `monitoring/grafana/dashboards/engagement-portfolio.json` shows: per-engagement uptime, per-engagement incident count, per-engagement runbook drill freshness, per-engagement upstream-sync recency, per-engagement regulatory-Pack version recency, per-engagement NPS, per-engagement renewal-likelihood. Per-engagement health score updated monthly.

**Code surface:** Grafana JSON (~300 LoC). Health-score job (~200 LoC).

**Doctrine rules:** 156, 157, 162.

**ADR triggers:** none (within existing ADR-0006 observability).

**Sprint:** Phase E Sprint S8.

### 4.26 Manifest-lock signature verification at admission

**Current:** Kyverno admission verifies image Cosign signatures. Manifest-lock verification not yet implemented.

**Target:** Kyverno policy verifies the manifest-lock signature alongside image signature. A Pod whose image has a Cosign signature but whose `manifest.lock.json` (mounted as ConfigMap) lacks the matching signature is denied.

**Code surface:** Kyverno policy (~120 LoC). Mount the manifest.lock.json as ConfigMap. Per-deployment lock generation in CD.

**Doctrine rules:** 73, 171.

**ADR triggers:** new ADR-0045 "Manifest-lock signature verification at Kyverno admission."

**Sprint:** Phase F Sprint S11.

### 4.27 Audit-evidence-pack export

**Current:** no `task audit-evidence-pack` target.

**Target:** `task audit-evidence-pack` produces a signed zip containing: every audit-log row in the period, every report with provenance + signature, every Pack manifest with lock hash, the OpenAPI spec, the Cosign signatures of every deployed image, the SBOM, the conformance-suite green status, the running-system manifest. The zip is itself signed.

**Code surface:** Taskfile target. New `internal/cmd/audit-evidence-pack` (~600 LoC). Tests against synthetic fixtures.

**Doctrine rules:** 63, 108, 145.

**ADR triggers:** new ADR-0046 "Audit-evidence-pack export tooling."

**Sprint:** Phase H Sprint S15.

### 4.28 Consumption forecaster (LightGBM + seasonal-ETS hybrid)

**Current:** not yet shipped.

**Target:** Phase I Sprint S18 ships the forecaster: LightGBM regressor on engineered features (lagged consumption, weather, calendar features, production plan from ERP) + seasonal-ETS for the seasonality, hybrid combiner with weights tuned on a per-tenant evaluation set. Forecast horizons 1-hour, 24-hour, 7-day. Inference latency P95 ≤ 300ms. MAPE budget: ≤ 12% day-ahead, ≤ 25% week-ahead, with statistically-significant improvement over seasonal baseline.

**Code surface:** Python training pipeline (~1500 LoC). Model artefact format. Go inference SDK (~400 LoC). Per-model model card. Drift detection (~300 LoC). Conformance test for determinism (~200 LoC). Inference observability via OTel.

**Doctrine rules:** 189, 190, 191, 192, 193, 194, 196, 197, 198, 199, 200, 205.

**ADR triggers:** new ADR-0047 "Consumption forecaster — LightGBM + seasonal-ETS hybrid." new ADR-0048 "Model registry and inference SDK."

**Sprint:** Phase I Sprint S18 (training pipeline + first deployable model). Phase I Sprint S19 (inference observability + drift detection).

### 4.29 Layered anomaly detection (Layers 1, 2, 3)

**Current:** Layer 1 (z-score on rolling 7-day baseline per meter) partial in `internal/services/alert_engine.go`.

**Target:** Layer 1 hardened. Layer 2 (STL seasonal decomposition) in Phase I Sprint S19. Layer 3 (cross-meter correlation) in Phase I Sprint S20. Each Layer is a Detector implementing `internal/domain/alerting.Detector` (EP-04). Each Layer has its own model card.

**Code surface:** Layer 2 (~400 LoC + model card). Layer 3 (~600 LoC + model card). Operator UI updates to show `detection_layer` annotation.

**Doctrine rules:** 127, 195, 199.

**ADR triggers:** new ADR-0049 "Layered anomaly detection."

**Sprint:** Phase I Sprint S19–S20.

### 4.30 SOC 2 Type II / ISO 27001 readiness

**Current:** evidence is implicit; no formal SOC 2 / ISO 27001 dossier.

**Target:** Phase J Sprint S21 produces the SOC 2 Trust Services Criteria dossier (Security, Availability, Processing Integrity, Confidentiality, Privacy) with mapped controls. Phase J Sprint S22 produces ISO 27001 Annex A control mapping. AgID Qualificazione Cloud per la PA dossier for Topology B.

**Code surface:** dossiers (~3000 LoC of structured Markdown). Evidence-pack auto-generation extension.

**Doctrine rules:** 25, 65, 108.

**ADR triggers:** new ADR-0050 "SOC 2 Type II readiness." new ADR-0051 "ISO 27001 Annex A control mapping."

**Sprint:** Phase J Sprint S21–S22.

### 4.31 Post-quantum-crypto plan (placeholder)

**Current:** not yet shipped.

**Target:** `docs/QUANTUM-CRYPTO-PLAN.md` (Phase J Sprint S22) addresses the path to post-quantum signatures (Dilithium / Falcon / SPHINCS+) and KEM (Kyber). Implementation deferred until NIST finalises and Cosign / Sigstore expose primitives. Annual review of the NIST timeline.

**Code surface:** plan document only (~300 LoC).

**Doctrine rules:** 187.

**ADR triggers:** new ADR-0052 "Post-quantum cryptography plan."

**Sprint:** Phase J Sprint S22.

### 4.32 v1.0.0 release

**Current:** at v0.x; no formal release tagging beyond commit SHAs.

**Target:** `v1.0.0` tag signed (Cosign keyless). Release notes published. License-decision ADR (BUSL-1.1 / SSPL / AGPLv3 / proprietary-perpetual). Migration of canonical repository to `github.com/greenmetrics/template` (or retention at the current location with public visibility — TBD by license decision).

**Code surface:** Release notes. License decision. Migration script.

**Doctrine rules:** Charter §12.1.

**ADR triggers:** new ADR-0053 "v1.0.0 release decision and licence."

**Sprint:** end of Phase E Sprint S8 (or end of Phase F Sprint S11 if delays).

---

## 5. Phase E — Engagement & Pack Extraction (S5–S8)

### 5.1 Phase intent

Phase E removes the SaaS framing, extracts Italian-flagship code into Packs, formalises Pack contracts, ships the engagement-fork model, lands the white-label configuration, and achieves v1.0.0 readiness. By end of Phase E the template is *charter-conformant*: any new client can be onboarded against the template using the documented engagement lifecycle, the doctrine is mechanically enforced for the modular-template-integrity rules (69–88), and the SaaS-vocabulary documents have been rewritten in place.

Phase E is the most architecturally-disruptive phase. It moves ~3000 LoC from `internal/services/` into Packs. It introduces the Pack-loader, the Pack-contract interfaces, the `engagements/` directory model, and the white-label configuration. It rewrites the four most SaaS-vocabulary-laden documents (MODUS_OPERANDI, COST-MODEL, THREAT-MODEL §2, ADR-0001 annotation). It is bounded at four sprints (S5–S8, ~16 weeks) with a tight scope discipline; over-runs into Phase F are charter-superseded only.

### 5.2 Phase entry gate

Before Sprint S5 begins, the following are in place:

- Charter `docs/MODULAR-TEMPLATE-CHARTER.md` adopted.
- Doctrine `docs/DOCTRINE.md` adopted with 210 rules.
- Plan `docs/PLAN.md` (this file) adopted.
- This plan reviewed by Macena platform-team.
- Sprint S4 deliverables verified (k6 dev script, compose rate-limit env interpolation; commit `303c74d`).

### 5.3 Sprint S5 — Doctrine adoption + Pack-loader infrastructure + SaaS-vocabulary doc rewrite

**Calendar:** weeks 1–2 of Phase E.
**Sprint goal:** the modular-template-integrity rules are enforceable; the four SaaS-vocabulary documents are rewritten; the Pack-loader infrastructure exists in code (even if no Italian Pack yet uses it).

**Deliverables:**

1. **Charter adoption ADR.** `docs/adr/0021-charter-and-doctrine-adoption.md`. Names the charter, the doctrine, the plan; declares all three binding from Sprint S5 forward. Charter-supersession of ADR-0001's "regulated-industry SaaS" framing.
2. **MODUS_OPERANDI v2.** `docs/MODUS_OPERANDI.md` rewritten in place. Engagement playbook replaces the SaaS playbook. Removed: per-meter pricing tiers, CAC/LTV/churn vocabulary, multi-tenant SaaS framing. Added: engagement lifecycle (Phases 0–5), tier T1/T2/T3, engagement margin, time-to-customisation, template-fit-score, Italian-flagship as flagship Pack. New ADR-0028 records the rewrite.
3. **COST-MODEL v2.** `docs/COST-MODEL.md` rewritten. Engagement margin replaces gross margin. Per-deployment cost model.
4. **THREAT-MODEL §2 rewrite.** "Single-tenant by default, multi-tenant by configuration" framing replaces the multi-tenant SaaS framing.
5. **Pack-loader code.** `internal/packs/{pack.go, manifest.go, contracts.go, registry.go, loader.go, lock.go}` (~1500 LoC). `config/required-packs.yaml` (initially empty). `docs/contracts/pack-manifest.schema.json`. Tests in `tests/packs/{manifest_validation_test.go, capability_matching_test.go, lock_test.go}`.
6. **Pack-contract interfaces.** `internal/domain/protocol/ingestor.go`, `internal/domain/reporting/builder.go`, `internal/domain/emissions/factor_source.go`, `internal/domain/identity/provider.go`, `internal/domain/region/profile.go`. Each with full Godoc + example test + version constant. ADR-0023 records.
7. **Universal-invariant conformance tests.** `tests/conformance/{no_float_money_test.go, rfc3339utc_test.go, uuidv4_test.go, rfc7807_test.go, cloudevents_test.go, health_test.go, log_format_test.go}`. Each test enforces one of Rules 1–8.
8. **Core-Pack separation conformance test.** `tests/conformance/core_pack_separation_test.go` walks the file tree and verifies the layering invariant.
9. **`config/branding.yaml` schema and defaults.** Schema in `docs/contracts/branding.schema.json`. Default file in `config/branding.yaml` carrying `product_name: GreenMetrics`. Vite plugin + PDF renderer + email-template integration starts (completes Sprint S6).
10. **Engagement-fork-bootstrap runbook.** `docs/runbooks/engagement-fork-bootstrap.md`.
11. **SECURITY.md / ITALIAN-COMPLIANCE.md / RISK-REGISTER.md SaaS-vocabulary annotations.** Where SaaS framing exists, add an inline note pointing to Charter §2.

**Exit criteria:**

- Charter adoption ADR merged.
- MODUS_OPERANDI v2 merged. SaaS pricing tiers removed.
- COST-MODEL v2 merged. CAC/LTV/churn vocabulary removed.
- THREAT-MODEL §2 rewritten.
- Pack-loader code green on `task verify`.
- Universal-invariant conformance tests green.
- Core-Pack separation test green (template has no Pack-namespaced imports yet, so trivially green).
- `config/branding.yaml` populated; conformance test for hard-coded brand strings *not yet enforced* (Sprint S6 deliverable to populate-and-enforce).

**Doctrine rules invoked:** 1–8, 14, 25, 27, 47, 67, 69, 70, 71, 72, 73, 75, 86, 88, 209.

**Risk-register deltas:** RISK-001 (single-operator bus factor) reaffirmed; mitigation = quarterly office hours plus the engagement-team hire planned for Phase H. RISK-002 (Mission II accepted residuals) starts closing as conformance-test population progresses.

**ADRs filed:** 0021, 0023, 0028.

### 5.4 Sprint S6 — Italian Region Pack + Italian Factor Packs + Italian Report Packs + white-label

**Calendar:** weeks 3–4 of Phase E.
**Sprint goal:** the Italian-specific code is in `packs/region/it/`, `packs/factor/{ispra,terna,gse,aib}/`, `packs/report/{esrs_e1,piano_5_0,conto_termico,tee,audit_dlgs102,monthly_consumption,co2_footprint}/`. White-label is enforced.

**Deliverables:**

1. **Italian Region Pack.** `packs/region/it/` populated with timezone (Europe/Rome), locale (it_IT.UTF-8), currency (EUR), holiday calendar (Italian national + Veneto regional holidays for the most-served region), privacy-regime overlay (GDPR + Garante guidance + ARERA classifications), per-region thresholds (CSRD wave triggers). Manifest. CHARTER. Tests.
2. **Italian Factor Packs.** `packs/factor/ispra/` carries the ISPRA national mix factors with full temporal versioning. `packs/factor/gse/` carries the GSE renewable shares + AIB residual mix entrypoints. `packs/factor/terna/` carries the Terna national mix daily fetcher with circuit-breaker. `packs/factor/aib/` carries the AIB residual mix annual fetcher. Each Pack has manifest + CHARTER + tests + simulator (offline mode for dev).
3. **Italian Report Packs.** `packs/report/esrs_e1/` ships the ESRS E1 Builder. `packs/report/piano_5_0/` ships the Piano 5.0 attestazione Builder with thresholds in `config.yaml`. `packs/report/conto_termico/` ships the Conto Termico GSE Builder (with XSD validation deferred to Phase H). `packs/report/tee/` ships the Certificati Bianchi TEE Builder. `packs/report/audit_dlgs102/` ships the audit 102/2014 Builder (EGE counter-signature workflow deferred to Phase H). `packs/report/monthly_consumption/` and `packs/report/co2_footprint/` ship the simpler periodic reports.
4. **REJ-11 god-service decomposition.** `report_generator.go` is reduced to an orchestrator (~80 LoC, per Rule 87 acceptance) that delegates to the loaded Builder Packs. ADR-0025 records.
5. **White-label conformance enforcement.** The Vite plugin pipes `config/branding.yaml` into the frontend. The PDF renderer reads it. The email templates use it. The conformance test `tests/conformance/no_hardcoded_brand_test.go` enforces no hard-coded brand strings outside the defaults.
6. **Engagement-fork model documents.** `docs/runbooks/upstream-sync.md`. `docs/runbooks/engagement-fork-bootstrap.md`. `docs/PACK-ACCEPTANCE.md`. Empty `engagements/.gitkeep` directory in template `main` with README.
7. **GDPR DSAR endpoint.** `/api/v1/dsar/{tenant_id}/{user_id}/export` and `.../erase`. RBAC role `dpo`. Endpoints implemented; full CDP integration deferred to Phase F.

**Exit criteria:**

- Italian Region + Factor + Report Packs all loaded by the Pack-loader at boot.
- `manifest.lock.json` written and signed at boot.
- White-label conformance test green.
- All previous integration tests green against the extracted Packs (no regression).
- SaaS-vocabulary documents fully aligned with charter (no mention of `SaaS pricing`, `CAC`, `LTV`, `churn`, `signup`, `subscription tier` outside historical-quote contexts).

**Doctrine rules invoked:** 32, 33, 69, 70, 71, 75, 81, 87, 88, 90, 91, 95, 100, 101, 129, 130, 131, 132, 133, 134, 137, 165, 184.

**Risk-register deltas:** RISK-005 (supply-chain SHA-pinning) progresses as Pack-loader code includes its own dependency-pin check at load. New RISK-024 "Pack-extraction regression risk" added; mitigation = run all Mission II audit cases against the post-extraction code with diff < epsilon.

**ADRs filed:** 0024, 0025, 0026, 0027.

### 5.5 Sprint S7 — Italian Protocol Packs + per-protocol simulators + DSO-circuit-breaker hardening

**Calendar:** weeks 5–6 of Phase E.
**Sprint goal:** all five existing Italian-flagship protocols are in Packs with simulators and conformance tests. DSO clients are circuit-breakered.

**Deliverables:**

1. **Italian Protocol Packs.** `packs/protocol/{modbus_tcp, modbus_rtu, mbus, sunspec, pulse, ocpp_1_6, ocpp_2_0_1}/`. Each Pack has manifest, CHARTER (with wire-format invariants), simulator, device-profile catalogue starter (top-3 vendor in each class), tests.
2. **Per-protocol simulators.** Each Pack ships a simulator at `packs/protocol/<id>/simulator/` runnable in `docker compose up`. Simulators are deterministic (script-driven).
3. **Device-profile catalogue.** `packs/protocol/modbus_tcp/devices/{carlo-gavazzi-em24.yaml, socomec-countis-x.yaml, lovato-dmg-300.yaml, schneider-powerlogic-pm3250.yaml, abb-m4m-30.yaml}`. `packs/protocol/sunspec/devices/{sma-stp-25000-tlhe.yaml, fronius-symo-15-2.yaml, abb-tripeak-2-1.yaml}`. Per-device polling minimum, register layout, scaling.
4. **DSO circuit-breaker hardening.** E-Distribuzione SMD client, Terna Transparency client, SPD multi-DSO client all wrapped with `sony/gobreaker/v2` per ADR-0009 + Rule 125. Cached fallback with `data_freshness` stamp on consumer surfaces.
5. **Asynq queue + worker.** Phase 1 of asynq integration per ADR-0014 lands the actual worker with `report:*` job types. The async report-generation surface goes live (REJ-21 mitigation completes).
6. **TLS 1.3 enforcement.** Ingress controller (per Topology) declares TLS 1.3 only. Conformance test `tests/conformance/tls_version_test.go`.

**Exit criteria:**

- All seven Protocol Packs load at boot.
- Each Pack's simulator runnable and tested.
- DSO clients are circuit-breakered with documented cooldown / threshold.
- Async report generation works end-to-end (POST returns 202, worker completes, GET retrieves).
- TLS 1.3 enforced in production topology configurations.

**Doctrine rules invoked:** 36, 37, 109, 110, 111, 113, 114, 115, 116, 117, 118, 119, 121, 122, 124, 125, 174.

**Risk-register deltas:** RISK-014 (DSO portal outage) closing — circuit-breakers in place. New RISK-025 "Pack-extraction performance regression" mitigated — k6 perf bench against pre/post-extraction baseline in CI.

**ADRs filed:** 0023 (re-cited).

### 5.6 Sprint S8 — Identity Packs + engagement-health dashboard + v1.0.0 readiness review

**Calendar:** weeks 7–8 of Phase E.
**Sprint goal:** SAML and OIDC Identity Packs are in place. Engagement-health dashboard is operational. The v1.0.0 readiness review is complete; remaining gaps queued for Phase F.

**Deliverables:**

1. **SAML Identity Pack.** `packs/identity/saml/` with SP metadata generation, IdP metadata ingestion, JIT user provisioning, per-tenant IdP routing. Validated against CyberArk, Okta, Azure AD, Keycloak. ADR-0043.
2. **OIDC Identity Pack.** `packs/identity/oidc/` with discovery, PKCE, per-tenant provider routing. Validated against Auth0, Okta OIDC, Google. ADR-0044.
3. **Engagement-health dashboard.** Grafana JSON in `monitoring/grafana/dashboards/engagement-portfolio.json`. Per-engagement health-score job in `internal/jobs/engagement_health.go`. Per-engagement metrics scraped from per-deployment Prometheus federations.
4. **v1.0.0 readiness review.** A reviewer-led pass through the Charter §16 cross-references and the Doctrine §11 meta-rules. The readiness review identifies any Phase E deliverable that's behind and decides whether v1.0.0 ships at end of S8 or end of S11. The review is recorded in ADR-0053.
5. **Charter office hours #1.** First quarterly charter-and-doctrine office hours (~4 hours): walk every charter clause, every doctrine rule against the post-Sprint-S8 codebase, surface drift, file new rejection ADRs as needed.

**Exit criteria:**

- SAML and OIDC Packs each pass an end-to-end login-flow integration test against a containerised IdP.
- Engagement-health dashboard renders with at least one synthetic engagement.
- v1.0.0 readiness review concluded with explicit ADR.
- Office hours #1 minutes filed.

**Doctrine rules invoked:** 70, 156, 161, 162, 209, 210.

**Risk-register deltas:** new RISK-026 "v1.0.0 license decision deferral" added; mitigation = decision before Phase J Sprint S22.

**ADRs filed:** 0043, 0044, 0053.

### 5.7 Phase E exit gate

Before Phase F begins, the following are in place:

- Sprints S5–S8 deliverables all merged.
- Conformance suite passes including all universal-invariant tests + Core-Pack separation + white-label.
- Italian Region Pack, four Factor Packs, seven Report Packs, seven Protocol Packs, two Identity Packs all loaded at boot via Pack-loader; manifest.lock.json is signed at boot.
- v1.0.0 readiness ADR filed.
- Office-hours minutes #1 filed.

If any of the above is incomplete, Phase F start is deferred until the gap is closed. A two-week slip is absorbed; longer slips trigger a re-plan ADR.

---

## 6. Phase F — Audit-Grade Reproducibility (S9–S11)

### 6.1 Phase intent

Phase F lands the audit-grade reproducibility property. By end of Phase F: every regulatory output is bit-perfect re-derivable from the same `(period, factors, readings, manifest_lock_hash)`; the audit log is chained-HMAC-signed with hourly published checkpoints; reading provenance is signed at ingestion with per-meter HMAC; field-level encryption protects raw payloads with per-tenant DEK + KMS-wrapped MEK; the Pack-manifest lock is signed and verified at admission. The replay test runs every Builder against pinned fixtures and asserts byte-identical output.

Phase F is the regulator-grade differentiation phase. Of every competitor surveyed, none guarantees bit-perfect reproducibility, none ships chained-HMAC audit logs, none signs every reading at ingestion. Phase F is what makes "regulator-grade modular template" a verifiable claim rather than a marketing line.

### 6.2 Phase entry gate

Phase E exit gate is closed. Plus: the engagement-health dashboard has a baseline (at least one synthetic engagement); the conformance-suite cycle time is below 8 minutes locally on a developer laptop (Rule 24).

### 6.3 Sprint S9 — Deterministic Builder framework + replay tests + per-tenant timezone

**Calendar:** weeks 9–10 of Phase F.
**Sprint goal:** every Builder is a pure function. Replay tests against pinned fixtures produce byte-identical output across reruns and across Go-version-matrix runs.

**Deliverables:**

1. **Builder purity refactor.** Every Builder under `internal/domain/reporting/<id>/` is refactored to be a pure function of `(period, factors, readings)`. `time.Now()`, `os.Getenv()`, `crypto/rand`, `math/rand`, and `internal/services/` imports are removed. Custom golangci-lint rule scopes the deny-list. Conformance test `tests/conformance/builder_purity_test.go` re-runs each Builder twice and asserts byte-identical output.
2. **Deterministic serialisation library.** `internal/domain/serialise/` provides `MarshalJSON(v) ([]byte, error)` with sorted keys, `MarshalXML` preserving schema-declared order, `MarshalPDFA2b(template, data)` with deterministic font subset and fixed PDF/A-2b producer string. All Builders route through these.
3. **Replay-test suite.** `tests/conformance/builder_replay_test.go` runs every Builder against pinned fixtures in `tests/fixtures/regulatory/<period>/<scenario>.golden.json`. The fixture set covers each Italian Report Pack with at least three scenarios (small / medium / corner-case). Golden-file update requires its own PR with a Tradeoff Stanza.
4. **Aggregation property tests expansion.** `tests/property/aggregate_invariants_test.go` expanded to cover `Sum`, `Average`, `WeightedAverage`, `Max`, `Min`, `P95` against associativity (where applicable), commutativity (where applicable), identity. Property test runs on every CI.
5. **Per-tenant timezone.** Migration adds `tenants.timezone` column. Region Pack supplies the default. Time-bucket library (`internal/domain/timebucket/`) reads per-tenant. Conformance test `tests/conformance/tenant_timezone_test.go`.
6. **Time-bucket library.** `internal/domain/timebucket/` provides `Bucket15m`, `Bucket1h`, `Bucket1d`, `BucketMonth`, `BucketYear` functions parameterised on timezone. Deterministic. Tested against transitional dates (DST spring-forward, DST fall-back, leap-year February 29).
7. **Float-in-regulatory-path conformance.** `tests/static/no_float_in_regulatory_test.go` enforces no `float64` / `float32` declaration inside `internal/domain/reporting/`, `internal/domain/emissions/`, `internal/domain/units/`. Fixed-point microgram / milliwatt-hour types defined.
8. **Annual ISPRA factor refresh job.** `internal/jobs/ispra_factor_refresh.go` runs annually (April default), pulls from the ISPRA portal, validates against schema, opens a PR for review. Sandbox-only (no auto-merge).

**Exit criteria:**

- Replay test green on every Builder against fixtures, with byte-identical output across 100 reruns.
- Builder purity conformance test green.
- Property tests for aggregations green.
- Per-tenant timezone migration applied and conformance green.
- Float-in-regulatory-path conformance green (zero violations).
- ISPRA factor refresh job tested in dry-run mode.

**Doctrine rules invoked:** 89, 91, 92, 93, 94, 101, 132, 138, 141, 142.

**Risk-register deltas:** new RISK-027 "Determinism regression on Go-version upgrade." Mitigated by Go-version-matrix in CI (Go 1.25 / 1.26 / 1.27).

**ADRs filed:** 0032.

### 6.4 Sprint S10 — Field-level encryption + corrections-overlay + GDPR DSAR endpoint

**Calendar:** weeks 11–12 of Phase F.
**Sprint goal:** raw payloads are field-level encrypted with per-tenant DEK; corrections workflow is in place; DSAR endpoint is wired.

**Deliverables:**

1. **Per-tenant DEK provisioning.** At tenant creation, generate a 256-bit DEK using `crypto/rand`, wrap with the regional KMS MEK, store the wrapped DEK in `tenants.wrapped_dek bytea`. ADR-0031.
2. **Field-level encryption on `readings.raw_payload`.** Migration adds `dek_id`, `dek_version`, `payload_encrypted` columns. Encryption helper in `internal/security/crypto.go` provides `EncryptRawPayload(ctx, tenantID, plaintext) ([]byte, dekID, error)` and `DecryptRawPayload(ctx, tenantID, ciphertext, dekID) ([]byte, error)`. Repository writes use Encrypt; reads use Decrypt; backward-compatible reads tolerate `dek_id IS NULL` for pre-encryption rows.
3. **Crypto-shredding implementation.** Tenant deletion (soft-delete with `active = false`) triggers DEK deletion in KMS within 30 days (configurable per tenant). Audit-log row records the shredding event. Conformance test `tests/conformance/crypto_shredding_test.go` verifies that decryption fails for a shredded tenant.
4. **Corrections-overlay table.** Migration adds `corrections (correction_id uuid pk, tenant_id, meter_id, channel_id, ts, original_value, corrected_value, reason, created_by_user_id, created_at)`. RBAC role `auditor` allowed to insert. Builders read `readings JOIN corrections`. ADR-0033.
5. **GDPR DSAR full endpoint.** `/api/v1/dsar/{tenant_id}/{user_id}/export` returns the full data export as a signed zip. `.../erase` triggers the soft-delete + DEK shred. The DSAR runbook `docs/runbooks/dsar.md`.
6. **DependencyTrack instance.** `gitops/base/dependency-track/` deploys a DependencyTrack instance for vulnerability cross-referencing against SBOMs. SBOM upload integrated into CD pipeline. Vulnerability-response SLA dashboard.
7. **Engagement-health dashboard with real engagement.** First synthetic engagement (a fully-mocked client) running on the engagement-fork pattern; engagement-health dashboard shows real signals. The synthetic engagement's runbooks added under `engagements/synthetic/runbooks/`.

**Exit criteria:**

- Per-tenant DEK provisioned for every tenant.
- All new readings ingest with field-level encryption.
- Backward-compatibility test green: pre-encryption readings still readable.
- Crypto-shredding test green.
- Corrections workflow end-to-end green (auditor inserts a correction; report rerun reflects the correction).
- DSAR endpoint round-trip works on synthetic data.
- DependencyTrack ingests SBOMs.

**Doctrine rules invoked:** 39, 58, 59, 96, 104, 165, 172, 184.

**Risk-register deltas:** RISK-022 "Plaintext payloads at rest" closing — encryption in place.

**ADRs filed:** 0031, 0033.

### 6.5 Sprint S11 — Chained-HMAC audit log + per-meter HMAC reading provenance + manifest-lock admission verify

**Calendar:** weeks 13–14 of Phase F.
**Sprint goal:** the cryptographic chain (audit-log chain + reading provenance + manifest-lock + signed reports) is end-to-end functional.

**Deliverables:**

1. **Chained-HMAC audit log.** Migration adds `prev_hmac bytea`, `row_hmac bytea`, `chain_key_version int` to `audit_log`. Insert trigger computes `row_hmac = HMAC-SHA256(audit_chain_key_v<chain_key_version>, prev_hmac || row_canonical_json)`. Hourly checkpoint job in `internal/jobs/audit_checkpoint.go` publishes a signed checkpoint to S3 Object Lock. Annual key rotation with versioned `chain_key_version`. ADR-0029. Conformance test `tests/security/audit_chain_test.go`.
2. **Per-meter HMAC keys.** Meter onboarding flow generates a per-meter 256-bit HMAC key, stores in KMS with key-name `meter-<meter_id>-<key_version>`. Edge gateway pulls the key into a sealed enclave (or uses KMS API where reachable) at first connect.
3. **Reading provenance with HMAC verify.** Ingest pipeline verifies the HMAC over `(meter_id, channel_id, ts_ns, value_int_micro_unit)`. Failed-verify readings carry `quality_code = 'unsigned'`. Conformance test `tests/conformance/signed_reading_test.go`.
4. **Manifest-lock signature verify at admission.** Kyverno policy verifies `manifest.lock.json` signature alongside image signature. ConfigMap mount of the lock per Pod. ADR-0045.
5. **Report signing at finalisation.** Reports transition `draft → generated → signed → submitted`. `signed` requires Cosign sign-blob over the canonical JSON serialisation. Signature stored in `reports.signature`. Submission only allowed in `signed` state.
6. **Algorithm-versioning columns.** Schema columns for `dek_algorithm`, `chain_hmac_algorithm`, `signature_algorithm`, `password_algorithm` ensure crypto-agility per Rule 186.
7. **DR drill on Phase F deliverables.** Quarterly DR drill exercises: take last week's backup, restore to staging, run audit-chain integrity verification, run report replay tests, run signed-reading verification, compare against production-saved versions. Drill passes per Rule 107.

**Exit criteria:**

- Audit-chain integrity verified across a 7-day rolling window.
- Per-meter HMAC verification on 100% of new readings (synthetic test).
- Manifest-lock signature enforced — unsigned-lock Pod denied.
- Reports signed at finalisation.
- DR drill passes including audit-chain verification.

**Doctrine rules invoked:** 62, 73, 89, 95, 105, 144, 169, 171, 173, 186.

**Risk-register deltas:** RISK-021 "Audit-log tamperability" closing — chained-HMAC in place. RISK-023 "Unsigned readings" closing — per-meter HMAC in place.

**ADRs filed:** 0029, 0030, 0045.

### 6.6 Phase F exit gate

Before Phase G begins:

- Every Builder is pure (purity test green).
- Replay test green for every Builder against pinned fixtures.
- Field-level encryption on raw_payload.
- Chained-HMAC audit log operational.
- Per-meter HMAC reading provenance operational.
- Manifest-lock signature verified at admission.
- Reports signed at finalisation.
- DR drill passed.
- Conformance suite cycle time still ≤ 8 minutes locally.

If any incomplete, Phase G defers. Two-week absorbed; longer triggers re-plan.

---

## 7. Phase G — OT Integration Maturity (S12–S14)

### 7.1 Phase intent

Phase G ships the production-ready edge gateway with disk-backed buffering and GPS time discipline; ships the IEC 61850 + OPC UA + MQTT Sparkplug B + BACnet Protocol Packs; lands Topology D (hybrid OT-segment + IT-segment); makes channel-mapping auditor-visible.

By end of Phase G the OT integration depth is decisive: 12 Protocol Packs, real edge gateway, NTP/GPS time, 24-hour disk-backed buffer, signed readings end-to-end, network-segmentation-aware deployment. The competitor surveyed in `docs/COMPETITIVE-BRIEF.md` is at most 3–4 protocols deep with shallow time-discipline; we are 12 deep with cryptographic provenance.

### 7.2 Sprint S12 — IEC 61850 + OPC UA Protocol Packs

**Calendar:** weeks 15–16 of Phase G.
**Sprint goal:** two of the three highest-value advanced industrial protocols are in Packs with simulators.

**Deliverables:**

1. **IEC 61850 Protocol Pack.** `packs/protocol/iec_61850/` ships an MMS client speaking IEC 61850-8-1. Coverage of LN classes XCBR, XSWI, MMXU, MMTR, LLN0. Per-IED profile catalogue starter (top-3 vendors). Simulator. CHARTER. Tests. ADR-0035.
2. **OPC UA Protocol Pack.** `packs/protocol/opc_ua/` ships an OPC UA client with security-policy `Basic256Sha256` minimum. Coverage of DA, AC, HA, PROSYS DI companion specs. Per-server-config catalogue starter. Simulator. CHARTER. Tests. ADR-0036.
3. **Per-protocol latency budgets.** Per-protocol histogram metrics. Per-protocol latency-exceeded alerts.
4. **Protocol-Pack acceptance review.** Each Pack passes the `docs/PACK-ACCEPTANCE.md` checklist signed by a platform-team reviewer in under 30 minutes from clean clone.

**Exit criteria:**

- IEC 61850 Pack reads MMS data from simulator and from one real-IED hardware target (lab setup).
- OPC UA Pack reads data from simulator and from one real-server target (lab setup).
- Per-protocol latency budgets active.

**Doctrine rules invoked:** 87, 109, 110, 121, 124.

**Risk-register deltas:** none new.

**ADRs filed:** 0035, 0036.

### 7.3 Sprint S13 — MQTT Sparkplug B + BACnet Protocol Packs

**Calendar:** weeks 17–18 of Phase G.
**Sprint goal:** the de-facto modern OT protocol (Sparkplug B) and the dominant building-automation protocol (BACnet) are in Packs.

**Deliverables:**

1. **MQTT Sparkplug B Protocol Pack.** `packs/protocol/mqtt_sparkplug_b/` ships an MQTT client subscribing to Sparkplug-B-formatted topics with NBIRTH / NDATA / NDEATH / DBIRTH / DDATA / DDEATH semantics. Adds an MQTT broker dependency (eclipse-mosquitto or HiveMQ depending on engagement choice). ADR-0037 documents the dependency.
2. **BACnet/IP Protocol Pack.** `packs/protocol/bacnet/` ships a BACnet/IP client with `analogInput`, `analogOutput`, `analogValue`, `binaryInput`, `binaryValue` object support. Per-device profile catalogue. Simulator. ADR-0038.
3. **IEC 62056-21 IR optical Pack stub.** `packs/protocol/iec_62056_21/` minimum viable stub for the IR-optical-port reading flow on IEC 62056-21-compliant electrical meters; full implementation deferred to engagement demand.
4. **Per-protocol simulator-against-Pack integration test matrix.** `tests/integration/protocol_simulator_test.go` matrix grows to cover all 11 active Packs.

**Exit criteria:**

- Sparkplug B Pack reads from a Mosquitto broker fed by a Sparkplug-B simulator.
- BACnet Pack reads from a simulator.
- Integration matrix green.

**Doctrine rules invoked:** 87, 109, 110, 121.

**Risk-register deltas:** new RISK-028 "MQTT broker as new failure domain in deployments using Sparkplug B." Mitigation = per-deployment broker-health monitoring + Pack-health surfaces broker status.

**ADRs filed:** 0037, 0038.

### 7.4 Sprint S14 — Edge gateway + Topology D + channel mapping auditor surface

**Calendar:** weeks 19–20 of Phase G.
**Sprint goal:** the production-grade edge gateway is shipped. Topology D (hybrid) is documented and deployable. Channel-mapping is auditor-visible.

**Deliverables:**

1. **Edge gateway (`cmd/edge-gateway`).** New `cmd/edge-gateway/main.go`. Loads protocol drivers from per-Pack drivers. 24h disk-backed buffer. NTP-synced clock with optional GPS-time stratum-1 source where hardware supports. Per-meter HMAC signing at ingestion. mTLS to backend with per-engagement certificate. Observability + health surface. Distroless nonroot image. Cross-compiled for ARM64 + AMD64 + ARMv7. Cosign-signed. ADR-0034.
2. **Topology D documentation.** `docs/DEPLOYMENT-TOPOLOGY-D.md`: ingest backend deployed inside OT segment, frontend / reporting in IT segment, site-to-site VPN with strict segmentation, NetworkPolicy bundles per-segment.
3. **Channel-mapping auditor surface.** `/api/v1/meters/{id}/channels` endpoint returns the channel taxonomy (Phase R / S / T / Total / Reactive / Frequency / Voltage / Current / etc.) with units. RBAC role `auditor` reachable. Frontend `meters/[id]/channels/+page.svelte` renders the surface for auditor-role users.
4. **Edge-gateway runbook.** `docs/runbooks/edge-gateway-deploy.md`. `docs/runbooks/edge-gateway-secure-boot.md`. `docs/runbooks/edge-gateway-buffer-overflow.md`.
5. **Edge-gateway chaos drill.** Drill scenarios: WAN flap (24 hours offline), GPS antenna fault, time-skew detection, sealed-enclave key rotation, signed-reading verification under load.

**Exit criteria:**

- Edge gateway builds for all target architectures, signed, deploys to a development Raspberry-Pi-class device.
- 24h offline buffer drill: gateway buffers and re-sends correctly on reconnect.
- Channel-mapping endpoint serves auditor-role queries.
- Topology D documented and deployed in lab setup.

**Doctrine rules invoked:** 105, 109, 111, 112, 119, 123, 126, 173.

**Risk-register deltas:** RISK-014 (DSO portal outage) further closing — edge-gateway buffer protects against backend-side outage too.

**ADRs filed:** 0034.

### 7.5 Phase G exit gate

Before Phase H begins:

- 12 Protocol Packs loaded, simulators tested, device profiles populated.
- Edge gateway deployed in lab + first staging engagement.
- Topology D documented and lab-tested.
- Channel-mapping auditor surface live.
- Per-protocol latency budgets active.

If any incomplete, Phase H defers.

---

## 8. Phase H — Regulatory Pack Catalogue (S15–S17)

### 8.1 Phase intent

Phase H ships the formal-spec validation against EFRAG XBRL / GSE XSD / ENEA XSD; ships UK-SECR + GHG-Protocol + ISO-14064-1 + TCFD + IFRS-S1/S2 Report Packs as the second-wave reference; ships UK-DEFRA + EPA-eGRID Factor Packs; ships the German Region Pack and the Spanish Region Pack to support DACH + Iberia expansion.

By end of Phase H the regulatory catalogue is the second-largest moat after audit-grade reproducibility. Every Italian regulatory output validates against the canonical spec before submission. Five non-Italian regulatory shapes are ready for engagement assembly. Two non-Italian Region Packs are charter-conformant.

### 8.2 Sprint S15 — EFRAG XBRL + GSE XSD + audit-evidence-pack export

**Calendar:** weeks 21–22.
**Sprint goal:** ESRS E1 outputs validate against EFRAG XBRL taxonomy. Conto Termico XML validates against GSE XSD. The audit-evidence-pack export is end-to-end functional.

**Deliverables:**

1. **EFRAG XBRL ESRS taxonomy mapping.** `packs/report/esrs_e1/taxonomy.yaml` declares the per-disclosure mapping. Builder emits XBRL output. CI validates against the published taxonomy. ADR-0039.
2. **GSE XSD validation for Conto Termico.** `packs/report/conto_termico/xsd/` carries the GSE XSD. Builder emits XML. Validator enforces. ADR-0040.
3. **Audit-evidence-pack export.** `task audit-evidence-pack` produces a signed zip with all Charter §1.1 artefacts. Per-Pack evidence builder. Time budget: ≤ 15 minutes for 1-year window. ADR-0046.
4. **Annual-review checklist for Italian Packs.** `docs/COMPLIANCE/ANNUAL-REVIEW.md` carries the per-Pack annual-review checklist. Calendar reminder for April annual ISPRA cycle.

**Exit criteria:**

- ESRS E1 XBRL output passes EFRAG validation.
- Conto Termico XML passes GSE XSD validation.
- Audit-evidence-pack zip generated, signed, contents schema-valid.

**Doctrine rules invoked:** 108, 131, 132, 134, 135, 138, 145.

**Risk-register deltas:** RISK-015 (regulator schema change) further mitigated — formal validation catches schema deltas at build.

**ADRs filed:** 0039, 0040, 0046.

### 8.3 Sprint S16 — ENEA XSD + EGE counter-signature workflow + UK-SECR Report Pack

**Calendar:** weeks 23–24.
**Sprint goal:** D.Lgs. 102/2014 audit dossier flows through the EGE counter-signature workflow. UK-SECR Report Pack ships.

**Deliverables:**

1. **ENEA XSD validation for audit 102/2014.** `packs/report/audit_dlgs102/xsd/` carries the ENEA XSD. Builder emits XML and validates. Workflow state machine `draft → generated → ege_signing → signed → submitted`. EGE-portal integration stub. ADR-0041.
2. **UK-SECR (Streamlined Energy and Carbon Reporting) Report Pack.** `packs/report/uk_secr/`. Builder produces the SECR mandatory disclosures (energy use, emissions, intensity ratio, narrative, methodology). PDF formatted to UK Companies-House expected layout.
3. **GHG Protocol Corporate Standard Report Pack.** `packs/report/ghg_protocol/`. Generic Scope 1 / 2 / 3 reporting. Methodology references.
4. **ISO 14064-1 verification Report Pack.** `packs/report/iso_14064_1/`. Quantification and reporting of GHG emissions. Verification statement template.
5. **TCFD Report Pack.** `packs/report/tcfd/`. Governance / Strategy / Risk Management / Metrics-and-Targets disclosures.
6. **Pack-acceptance reviews for each new Pack.** Five acceptance reviews + ADRs for each Pack.

**Exit criteria:**

- audit 102/2014 dossier generates valid ENEA-XML and supports EGE counter-signature.
- All four new Report Packs load, generate output, pass acceptance.

**Doctrine rules invoked:** 87, 129, 131, 132, 137, 144.

**Risk-register deltas:** none new.

**ADRs filed:** 0041 + per-Pack acceptance ADRs.

### 8.4 Sprint S17 — IFRS S1/S2 + UK-DEFRA + EPA-eGRID Factor Packs + German + Spanish Region Packs

**Calendar:** weeks 25–26.
**Sprint goal:** non-Italian Region Packs ship; non-Italian Factor Packs ship; IFRS S1/S2 disclosures supported.

**Deliverables:**

1. **IFRS S1/S2 Report Pack.** `packs/report/ifrs_s_1_s_2/`. Climate-related financial disclosures (S2) plus general sustainability disclosures (S1). The most commercially-relevant non-EU regulatory shape in 2026.
2. **UK-DEFRA Factor Pack.** `packs/factor/uk_defra/`. Annual UK-government-published emission factors.
3. **EPA-eGRID Factor Pack.** `packs/factor/epa_egrid/`. US EPA emission-factor data for the eGRID grid.
4. **German Region Pack.** `packs/region/de/`. Timezone Europe/Berlin, locale de_DE.UTF-8, currency EUR, German national holidays, BAFA energy-management invariants, DSGVO (German GDPR) overlay. References the German factor source (UBA / EEX) — `packs/factor/uba/` ships a stub; full implementation in a later sprint.
5. **Spanish Region Pack.** `packs/region/es/`. Timezone Europe/Madrid, locale es_ES.UTF-8, currency EUR, Spanish national holidays, MITECO invariants. References Spanish factor source (MITECO + REE) — stub.
6. **Pack-extraction lessons-learned ADR.** A retrospective ADR (`docs/adr/0054-lessons-from-pack-extraction.md`) capturing what worked and what didn't in the extraction work of Phase E and Phase H. Feeds the doctrine.

**Exit criteria:**

- IFRS S1/S2 Pack generates output.
- UK-DEFRA + EPA-eGRID Factor Packs operational.
- German + Spanish Region Packs charter-conformant.
- Lessons-learned ADR filed.

**Doctrine rules invoked:** 88, 129, 130, 131, 132, 138, 160, 168, 209.

**Risk-register deltas:** new RISK-029 "Multi-region factor sources require multi-license management." Mitigation = per-Factor-Pack license declaration in manifest.

**ADRs filed:** 0054 + per-Pack acceptance ADRs.

### 8.5 Phase H exit gate

Before Phase I begins:

- All Italian Report Packs validate against formal specs (XBRL / GSE / ENEA).
- Audit-evidence-pack export functional.
- Five additional Report Packs (UK-SECR, GHG Protocol, ISO 14064-1, TCFD, IFRS S1/S2) live.
- Two additional Factor Packs (UK-DEFRA, EPA-eGRID) live.
- Two additional Region Packs (DE, ES) live.

---

## 9. Phase I — AI/ML & Forecasting (S18–S20)

### 9.1 Phase intent

Phase I ships the consumption forecaster, the layered anomaly detection (Layers 2 + 3), the model-card discipline, drift detection, model registry, inference observability, explainability surface, and the optional MSD flexibility-market connector for tenants that opt in.

By end of Phase I, AI/ML is observable, reproducible, drift-monitored, and explainable. The forecaster beats the seasonal baseline by ≥ 10% MAPE on the held-out test set. Anomaly detection produces actionable alerts with explanations.

### 9.2 Sprint S18 — Consumption forecaster (LightGBM + seasonal-ETS hybrid)

**Calendar:** weeks 27–28.
**Sprint goal:** the forecaster ships with model card, training-data lineage, evaluation against budget, and inference SDK.

**Deliverables:**

1. **Training pipeline.** Python pipeline (`models/forecaster/training/`) builds the LightGBM regressor on engineered features (lagged consumption, weather, calendar, production plan from ERP) and the seasonal-ETS for seasonality. Hybrid combiner with weights tuned per-tenant.
2. **Per-model model card.** `models/forecaster/MODEL-CARD.md` documents training data window, sources, hyperparameters, evaluation metrics, intended use, known failure modes, dataset bias notes.
3. **Training-data lineage capture.** Each training run captures dataset hashes, sources, windows, environment, hyperparameters, eval-set hashes, signed via Cosign sign-blob. Stored in `models/forecaster/training-runs/<run-id>/`.
4. **Inference SDK.** `internal/domain/inference/forecaster.go` provides `Forecast(ctx, tenantID, horizon time.Duration) ([]Forecast, error)`. Determinism via pinned-seed model artefact. OTel spans wrapping inference.
5. **Forecast-determinism conformance test.** `tests/conformance/forecast_determinism_test.go` runs the same inference twice and asserts byte-identical output.
6. **Forecast-evaluation-budget gate.** Pre-deploy gate: forecaster must beat seasonal baseline by ≥ 10% MAPE day-ahead, ≥ 5% MAPE week-ahead. ADR-0047 records the methodology and the result.
7. **Explainability surface.** Per-forecast `explanation` field listing top-k input features by SHAP. Frontend renders. Conformance test.

**Exit criteria:**

- Forecaster trained on at least 90 days of synthetic data, evaluated on a held-out 30-day window.
- MAPE budget green.
- Inference SDK integrated.
- Explanation rendered in operator UI.

**Doctrine rules invoked:** 189, 190, 191, 192, 197, 200, 202, 205.

**Risk-register deltas:** new RISK-030 "AI/ML accuracy regression on retraining." Mitigation = per-retraining gate.

**ADRs filed:** 0047, 0048.

### 9.3 Sprint S19 — Layer 2 anomaly detection (STL) + drift detection + inference observability

**Calendar:** weeks 29–30.
**Sprint goal:** Layer 2 detector deployed; drift detection runs continuously; inference observability is mature.

**Deliverables:**

1. **Layer 2 anomaly detector (STL seasonal decomposition).** `internal/domain/alerting/detectors/stl/`. Detects anomalies in the trend + seasonal residual after decomposition. Per-detector model card.
2. **Drift detection.** `internal/jobs/drift_detector.go` computes Kolmogorov-Smirnov on input features over rolling windows; mean/variance of predictions; label drift when ground truth lands. Alerts when sustained drift exceeds documented threshold.
3. **Inference observability.** Per-inference OTel span (model name, version, input hash, output, latency). Per-model latency histograms in Prometheus. Latency budget P95 ≤ 300ms forecast, ≤ 100ms anomaly. Budget-violation alerts.
4. **Model registry.** `models/registry.json` is the canonical list of deployed models with versions. Inference SDK reads at boot. Rollback to previous version is a one-command operation.

**Exit criteria:**

- Layer 2 detector produces anomalies on synthetic-with-injected-anomaly fixture.
- Drift detection runs on synthetic distribution-shift scenarios.
- Inference observability dashboard renders.

**Doctrine rules invoked:** 127, 193, 194, 195, 198, 199.

**Risk-register deltas:** none new.

**ADRs filed:** 0049.

### 9.4 Sprint S20 — Layer 3 anomaly detection (cross-meter) + MSD opt-in connector + AI/ML hardening

**Calendar:** weeks 31–32.
**Sprint goal:** Layer 3 detector deployed; MSD connector for opt-in tenants; AI/ML termination criteria reviewed.

**Deliverables:**

1. **Layer 3 anomaly detector (cross-meter correlation).** `internal/domain/alerting/detectors/cross_meter/`. Detects anomalies via deviation from learned cross-meter correlation matrix. Per-detector model card.
2. **MSD (Mercato del Servizio di Dispacciamento) connector.** `packs/integration/msd/`. For opt-in tenants only (per Rule 206), commits peak-shaving capacity to MSD via aggregator partner API. Settlement reconciliation flows back into reports module. ADR for the integration.
3. **AI bias evaluation.** Per-cohort MAPE / FP / FN metrics. Per-cohort dashboards.
4. **Per-model termination criteria review.** Each deployed model's termination criterion in its model card is reviewed. Models that fail the criterion are retired.
5. **Pen-test preparation #1.** Annual pen-test scoping document. Test environment provisioning. Points-of-contact.

**Exit criteria:**

- Layer 3 detector operational.
- MSD connector tested in sandbox against a partner-provided test endpoint.
- Bias evaluation produces per-cohort metrics.
- Pen-test scoping complete.

**Doctrine rules invoked:** 127, 195, 200, 201, 204, 206, 208.

**Risk-register deltas:** RISK-016 "MSD integration regulatory exposure" — opt-in-only mitigates.

**ADRs filed:** 0049 (re-cited).

### 9.5 Phase I exit gate

Before Phase J begins:

- Forecaster live with model card + lineage + drift detection + observability + explainability.
- Layered anomaly detection (Layers 1–3) live.
- MSD connector available for opt-in tenants.
- Pen-test scoped.

---

## 10. Phase J — Hardening & Certification (S21–S22)

### 10.1 Phase intent

Phase J finalises SOC 2 / ISO 27001 / AgID readiness, conducts the annual pen-test, files the post-quantum-crypto plan, and ships v1.0.0 with the licence decision recorded.

### 10.2 Sprint S21 — SOC 2 Type II readiness + AgID dossier + pen-test execution

**Calendar:** weeks 33–34.
**Sprint goal:** SOC 2 + AgID + pen-test in flight.

**Deliverables:**

1. **SOC 2 Type II readiness dossier.** `docs/COMPLIANCE/SOC2.md`. Trust Services Criteria mapping. Evidence pointers. ADR-0050.
2. **ISO 27001 Annex A control mapping.** `docs/COMPLIANCE/ISO27001.md`. Per-control evidence pointers. ADR-0051.
3. **AgID Qualificazione Cloud per la PA dossier.** `docs/COMPLIANCE/AGID.md`. Dossier for Topology B.
4. **Annual penetration test execution.** OSCP-certified third-party firm. Scope = backend + frontend + edge gateway + KMS surfaces. Findings tracked in `docs/PENTEST-CADENCE.md`. Remediation per Rule-59 SLA.
5. **NIS2 tabletop exercise.** Annual tabletop with simulated incident; ACN coordination simulated. Postmortem template exercised.

**Exit criteria:**

- SOC 2 dossier complete.
- ISO 27001 mapping complete.
- AgID dossier filed.
- Pen-test completed; high-severity findings remediated.
- NIS2 tabletop run; lessons recorded.

**Doctrine rules invoked:** 25, 60, 61, 65, 167.

**Risk-register deltas:** RISK-006 "Insider risk" — pen-test results inform mitigation.

**ADRs filed:** 0050, 0051.

### 10.3 Sprint S22 — Post-quantum-crypto plan + v1.0.0 release + license decision

**Calendar:** weeks 35–36.
**Sprint goal:** v1.0.0 ships.

**Deliverables:**

1. **Post-quantum-crypto plan.** `docs/QUANTUM-CRYPTO-PLAN.md`. Path to Dilithium / Falcon / SPHINCS+ signatures and Kyber KEM. Annual NIST-timeline review. ADR-0052.
2. **License decision.** ADR-0053 records the licence choice (BUSL-1.1 / SSPL / AGPLv3 / proprietary-perpetual). LICENSE file updated. NOTICE / TRADEMARK / CONTRIBUTING aligned.
3. **v1.0.0 release.** Tag `v1.0.0` signed. Release notes published. Migration of repository to `github.com/greenmetrics/template` (or retention with public visibility per the licence decision).
4. **Office hours #4 + final retrospective.** Walk every doctrine rule against the v1.0.0 codebase. File any unrejection ADRs. Capture lessons for the post-v1.0.0 plan (which is a separate document).
5. **Five-engagement portfolio review.** The five engagements that ran the lifecycle through Phase 5 are reviewed against the success criteria of §3.2.

**Exit criteria:**

- v1.0.0 tagged, signed, announced.
- Licence decision recorded.
- Post-quantum plan filed.
- Office hours and portfolio review concluded.

**Doctrine rules invoked:** 187, 209, 210.

**ADRs filed:** 0052, 0053.

---

## 11. Doctrine traceability matrix

The following table maps every Doctrine rule to the Phase / Sprint where it is *first* enforced or substantively addressed. Rules already satisfied at the start of Phase E are marked "S0". Rules with multiple touchpoints have the first listed; the per-sprint deliverables above carry the full picture.

| Rule | Title | First Sprint |
|---|---|---|
| 1 | Money is `(amount_cents, currency)` | S5 |
| 2 | RFC 3339 UTC timestamps | S5 |
| 3 | UUIDv4 tenant IDs | S0 (already in place) |
| 4 | RFC 7807 Problem Details | S0 (already in place; conformance test S5) |
| 5 | CloudEvents 1.0 | S5 |
| 6 | Health envelope | S0 |
| 7 | Structured JSON logs | S0 |
| 8 | Italian residency default | S0 |
| 9 | Platform discipline | S0 |
| 10 | Platform serves application teams | S0 |
| 11 | 11/31/51 sequence | S0 |
| 12 | Opinionated defaults | S0 |
| 13 | Abstractions are cost centres | S0 |
| 14 | Contract-first | S0 (OpenAPI in place); Pack contracts S5 |
| 15 | Layers map | S0 |
| 16 | IaC + state | S0 |
| 17 | DX | S0 |
| 18 | Sample policy | S0 |
| 19 | Sentinel refusal | S0 |
| 20 | Secrets management | S0 |
| 21 | Evolution + change mgmt | S0 |
| 22 | Events as integration | S6 (event bus indirection) |
| 23 | Single-on-call | S0 |
| 24 | Continuous verification | S0 |
| 25 | Quality threshold | ongoing |
| 26 | Rejection authority | S0 |
| 27 | Tradeoff Stanza | S0 |
| 28 | Termination criterion | S0 |
| 29 | Intent-revealing names | ongoing |
| 30 | Backend = first-class system | S0 |
| 31 | Named alternatives | S0 |
| 32 | DDD in `internal/domain/` | S6 (extraction) |
| 33 | Data is the system | S0 |
| 34 | Backend contract-first | S5 |
| 35 | Idempotency | S0 |
| 36 | Failure as normal | S0 |
| 37 | Performance budgets | S0 |
| 38 | Scalability axes | S0 |
| 39 | Security as core | S0 |
| 40 | Observability | S0 |
| 41 | Bounded concurrency | S0 |
| 42 | Resource lifecycles | S0 |
| 43 | Framework awareness | S0 |
| 44 | Testability | S5 (conformance suite) |
| 45 | Backend quality | ongoing |
| 46 | Backend rejection authority | S0 |
| 47 | Decision rationale | S0 |
| 48 | Backend termination | S0 |
| 49 | DevSecOps discipline | S0 |
| 50 | DevSecOps unified system | S0 |
| 51 | DevSecOps op readiness | S0 |
| 52 | Keyless CI/CD identity | S0 |
| 53 | Pinned dependencies | S0 |
| 54 | Policy as code | S0 |
| 55 | Layered policy gates | S0 |
| 56 | Layered CD gates | S0 |
| 57 | Supply-chain attestation | S0 |
| 58 | SBOM exportable | S10 (DependencyTrack) |
| 59 | Vulnerability SLA | S0 (declared) / S10 (dashboard) |
| 60 | Pen-test cadence | S21 |
| 61 | Incident response | S21 (tabletop) |
| 62 | Audit log immutable | S0 (role-revoke); S11 (chained-HMAC) |
| 63 | Compliance evidence exportable | S15 |
| 64 | Continuous verification | S0 |
| 65 | DevSecOps quality | ongoing |
| 66 | DevSecOps rejection | S0 |
| 67 | DevSecOps rationale | S0 |
| 68 | DevSecOps termination | S0 |
| 69 | Core/Pack distinction | S5 |
| 70 | Pack manifest | S5 |
| 71 | Pack-contract versioning | S5 |
| 72 | Pack registration via Registrar | S5 |
| 73 | Manifest lock | S5 (write); S11 (admission verify) |
| 74 | Pack health → core health | S5 |
| 75 | Pack capabilities declared | S5 |
| 76 | Pack failures isolated | S5 |
| 77 | `engagements/<client>/` | S6 |
| 78 | Merge-friendly Core | S5 (test); ongoing |
| 79 | Quarterly upstream sync | S6 (runbook) |
| 80 | Time-bounded Core customisations | S6 (runbook) |
| 81 | Branding configuration | S6 |
| 82 | Layered configuration | S5 |
| 83 | No recursive Pack discovery | S5 |
| 84 | Pack tests against Core | S5 |
| 85 | Pack-loader instrumentation | S5 |
| 86 | Contract-as-code | S5 |
| 87 | Pack acceptance | S6 (review process) |
| 88 | Italian Pack as flagship | S6 |
| 89 | Bit-perfect reproducibility | S9 (purity) |
| 90 | Temporal-validity factors | S0 (in place); refresh S9 |
| 91 | Pure Builders | S9 |
| 92 | Aggregation properties | S9 |
| 93 | Time-bucket determinism | S9 |
| 94 | No floats in regulatory | S9 |
| 95 | Provenance bundle | S9 (foundation); S11 (signed) |
| 96 | Source-data immutability | S10 (corrections-overlay) |
| 97 | Algorithm versioning | S9 |
| 98 | Replay test | S9 |
| 99 | Lineage queryable | S9 |
| 100 | Schema additive in regulatory | S0 (policy) / ongoing |
| 101 | Per-tenant timezone | S9 |
| 102 | CAGGs not edited in place | S0 |
| 103 | Retention shortenings ADR'd | S0 |
| 104 | Crypto-shredding | S10 |
| 105 | Reading provenance | S11 |
| 106 | Submission idempotent + signed | S15 (Conto Termico); ongoing |
| 107 | DR drill quarterly | S11 |
| 108 | Audit-evidence-pack export | S15 |
| 109 | OT protocols as Packs | S7 (Italian); S12-S13 (advanced) |
| 110 | Wire-format invariants documented | S7 |
| 111 | Edge buffering | S14 |
| 112 | NTP/GPS time | S14 |
| 113 | Quality codes | S9 (formalisation) |
| 114 | Modbus polling cadence | S7 |
| 115 | M-Bus addressing | S7 |
| 116 | SunSpec models | S7 |
| 117 | OCPP version pinning | S7 |
| 118 | Pulse webhook constant-time | S0 (in place); conformance S5 |
| 119 | Bounded ingest | S0 (in place via ADR-0015) |
| 120 | Unit conversions explicit | S9 |
| 121 | Per-protocol simulators | S7 (Italian); S12-S13 |
| 122 | Device profiles | S7 |
| 123 | Channel mapping auditor-visible | S14 |
| 124 | Per-protocol latency budget | S12 |
| 125 | Outbound DSO breakered | S7 |
| 126 | OT-aware segmentation | S14 (Topology D) |
| 127 | Layered anomaly detection | S19 (Layer 2); S20 (Layer 3) |
| 128 | At-least-once + idempotent consumers | S0 |
| 129 | Regulatory dossiers as Packs | S6 |
| 130 | Factor sources as Packs | S6 |
| 131 | Formal-spec validation | S15 |
| 132 | Italian primary-source citations | S6 |
| 133 | Piano 5.0 thresholds configurable | S6 |
| 134 | EFRAG taxonomy | S15 |
| 135 | GSE XSD compliance | S15 |
| 136 | TEE batch awareness | S6 (basic); S15 (hardened) |
| 137 | Audit 102/2014 EGE workflow | S16 |
| 138 | Annual Pack reviews | S15 |
| 139 | Thresholds propagated | S6 |
| 140 | Per-tenant regulatory profile | S6 |
| 141 | Deterministic serialisation | S9 |
| 142 | Half-open period intervals | S9 |
| 143 | Scope 3 opt-in | S17 (initial); ongoing |
| 144 | Reports signed at finalisation | S11 |
| 145 | Audit-evidence regulatory-Pack-aware | S15 |
| 146 | PDF cover-letters template-driven | S6 (foundation); S11 (PDF/A-2b) |
| 147 | GSE/ENEA notifications tracked | S15 |
| 148 | EGE / auditor dependency declared | S6 |
| 149 | Discovery as deliverable | S5 (runbook) |
| 150 | Engagement fork at Phase 1 | S6 |
| 151 | Phase 2 bounded | S6 (template) |
| 152 | Phase 3 bounded | S6 |
| 153 | Phase 4 bounded | S6 |
| 154 | Handover = operator readiness | S6 |
| 155 | Engagement runbooks | S10 (synthetic engagement) |
| 156 | Monthly engagement health | S8 |
| 157 | Intentional renewal | S8 (calendar) |
| 158 | Exit pack | S22 (`task engagement-exit-pack`) |
| 159 | Generalisation before upstream | ongoing |
| 160 | Post-mortem → doctrine | ongoing |
| 161 | T2/T3 on-call rotation | S8 (foundation) |
| 162 | Tri-annual reviews | S8 |
| 163 | Transparent pricing | S5 (MODUS_OPERANDI v2) |
| 164 | Customer-data ownership | S5 (MSA template) |
| 165 | Synthetic-by-default fixtures | S10 |
| 166 | Tier upgrades explicit | S8 |
| 167 | Crisis communication rehearsed | S21 (tabletop) |
| 168 | Portfolio feedback to roadmap | S22 |
| 169 | Audit log chained-HMAC | S11 |
| 170 | JWT pinned + KID rotated | S0 (in place); rotation runbook S6 |
| 171 | Cosign keyless | S0 |
| 172 | Field-level encryption | S10 |
| 173 | Per-meter HMAC | S11 |
| 174 | TLS 1.3 only | S7 |
| 175 | Cert rotation automated | S0 |
| 176 | `crypto/rand` | S0 (in place); conformance test S5 |
| 177 | bcrypt cost 12 | S0 |
| 178 | Webhook HMAC constant-time | S0 (in place); conformance S5 |
| 179 | KMS regional | S10 |
| 180 | Annual secret rotation | S0 |
| 181 | TOTP MFA admin | S8 |
| 182 | IP+email lockout | S0 (in place) |
| 183 | Session lifetimes | S0 |
| 184 | Crypto-shredding (Art. 17) | S10 |
| 185 | Standard-library crypto | S0 |
| 186 | Crypto agility versioning | S11 |
| 187 | Post-quantum plan | S22 |
| 188 | Crypto failures Sev-1 | ongoing |
| 189 | AI not regulatory by default | S18 |
| 190 | Model card | S18 |
| 191 | Training-data lineage | S18 |
| 192 | Forecast determinism | S18 |
| 193 | Forecast accuracy monitoring | S19 |
| 194 | Drift detection | S19 |
| 195 | Anomaly-layer model cards | S19 (Layer 2) |
| 196 | Staleness budget | S18 |
| 197 | Sprint-gated retraining | S18 |
| 198 | Model versioning + rollback | S19 |
| 199 | Inference observable | S19 |
| 200 | Explainability deliverable | S18 |
| 201 | Fairness review | S20 |
| 202 | Pinned ML libraries | S18 |
| 203 | Training PII handling | S18 |
| 204 | Bias evaluation | S20 |
| 205 | Forecast evaluation budget | S18 |
| 206 | MSD opt-in | S20 |
| 207 | AI/ML graceful degradation | S18 |
| 208 | AI/ML termination criterion | S20 |
| 209 | Doctrine evidence + Tradeoff Stanza | ongoing (per ADR) |
| 210 | Recorded supersession | ongoing |

---

## 12. Risk-register deltas

The plan introduces, mitigates, or closes the following risks. Cross-reference with `docs/RISK-REGISTER.md` (annotated in Phase E Sprint S6).

### 12.1 Risks introduced

- **RISK-024** "Pack-extraction regression risk." Phase E Sprint S6. Mitigation: run all Mission II audit cases against post-extraction code with diff < epsilon; perf-bench against pre/post-extraction baseline.
- **RISK-025** "Pack-extraction performance regression." Phase E Sprint S7. Mitigation: k6 perf bench in CI.
- **RISK-026** "v1.0.0 license decision deferral." Phase E Sprint S8. Mitigation: decision not later than Phase J Sprint S22.
- **RISK-027** "Determinism regression on Go-version upgrade." Phase F Sprint S9. Mitigation: Go-version-matrix in CI (Go 1.25 / 1.26 / 1.27).
- **RISK-028** "MQTT broker as new failure domain." Phase G Sprint S13. Mitigation: per-deployment broker-health monitoring; Pack-health surfaces broker status.
- **RISK-029** "Multi-region factor sources require multi-license management." Phase H Sprint S17. Mitigation: per-Factor-Pack license declaration in manifest.
- **RISK-030** "AI/ML accuracy regression on retraining." Phase I Sprint S18. Mitigation: per-retraining gate.

### 12.2 Risks mitigated or closing

- **RISK-001** "Single-operator bus factor." Reaffirmed in Phase E. Mitigation = quarterly office hours + planned Phase H engagement-team hire.
- **RISK-002** "Mission II accepted residuals." Closes progressively as conformance-test population progresses.
- **RISK-005** "Supply-chain SHA-pinning." Pack-loader's own dependency-pin check at load adds depth.
- **RISK-006** "Insider risk." Phase J Sprint S21 pen-test results inform.
- **RISK-014** "DSO portal outage." Closing in Phase E Sprint S7 (circuit-breakers) + Phase G Sprint S14 (edge-gateway buffer).
- **RISK-015** "Regulator schema change." Mitigated in Phase H Sprint S15 (formal-spec validation).
- **RISK-016** "MSD integration regulatory exposure." Mitigated in Phase I Sprint S20 (opt-in only).
- **RISK-018** "Terraform state SPOF." Already mitigated by S3 + DynamoDB lock + KMS state backend (Sprint S0); per-environment partitioning in Phase E Sprint S7 deepens.
- **RISK-021** "Audit-log tamperability." Closing in Phase F Sprint S11 (chained-HMAC).
- **RISK-022** "Plaintext payloads at rest." Closing in Phase F Sprint S10.
- **RISK-023** "Unsigned readings." Closing in Phase F Sprint S11.

---

## 13. ADR triggers (consolidated)

The plan triggers the following ADRs. Numbers are reserved in advance to keep the sequence intact; the actual ADR file is filed when its sprint reaches the deliverable. Each ADR carries the four-part Tradeoff Stanza.

| ADR # | Title | Sprint |
|---|---|---|
| 0021 | Charter and doctrine adoption (supersedes ADR-0001 SaaS framing) | S5 |
| 0022 | Manifest-lock signature and admission verification | S5 (foundation); S11 (admission) |
| 0023 | Pack-contract interfaces and versioning | S5 |
| 0024 | Italian Region Pack and Italian-flagship Pack catalogue | S6 |
| 0025 | Decomposition of `report_generator.go` into per-Builder Packs | S6 |
| 0026 | White-label configuration via `config/branding.yaml` | S6 |
| 0027 | Engagement-fork topology and upstream-sync discipline | S6 |
| 0028 | MODUS_OPERANDI v2 — engagement playbook | S5 |
| 0029 | Chained-HMAC audit-log integrity | S11 |
| 0030 | Per-meter HMAC reading signatures | S11 |
| 0031 | Field-level encryption with per-tenant DEK + KMS-wrapped MEK | S10 |
| 0032 | Deterministic Builder framework | S9 |
| 0033 | Corrections-overlay table | S10 |
| 0034 | OT edge gateway architecture | S14 |
| 0035 | IEC 61850 MMS Protocol Pack | S12 |
| 0036 | OPC UA Protocol Pack | S12 |
| 0037 | MQTT Sparkplug B Protocol Pack and broker dependency | S13 |
| 0038 | BACnet/IP Protocol Pack | S13 |
| 0039 | EFRAG XBRL ESRS taxonomy mapping | S15 |
| 0040 | GSE XSD validation for regulatory submissions | S15 |
| 0041 | ENEA XSD validation and EGE counter-signature workflow | S16 |
| 0042 | Automated annual ISPRA factor refresh workflow | S9 |
| 0043 | SAML 2.0 Identity Pack | S8 |
| 0044 | OIDC Identity Pack | S8 |
| 0045 | Manifest-lock signature verification at Kyverno admission | S11 |
| 0046 | Audit-evidence-pack export tooling | S15 |
| 0047 | Consumption forecaster — LightGBM + seasonal-ETS hybrid | S18 |
| 0048 | Model registry and inference SDK | S18 |
| 0049 | Layered anomaly detection | S19 |
| 0050 | SOC 2 Type II readiness | S21 |
| 0051 | ISO 27001 Annex A control mapping | S21 |
| 0052 | Post-quantum cryptography plan | S22 |
| 0053 | v1.0.0 release decision and licence | S22 |
| 0054 | Lessons from Pack extraction (retrospective) | S17 |

---

## 14. Per-sprint daily-standup checklist

A quick reference for the engineer working a sprint: are we hitting the gates?

- **At standup-start of each day:** which Sprint deliverable is in-flight? Is the conformance suite still green locally? Are any test runs over the 8-min budget (Rule 24)?
- **At standup-end of each day:** PRs merged → which doctrine rules just went from "manual" to "mechanical"? Any new RISK identified?
- **At sprint mid-point:** is the deliverable list 50% done? If not, escalation conversation: is scope wrong, or is execution behind?
- **At sprint exit:** every deliverable green? Exit criteria met? ADRs filed? CHANGELOG updated? Charter and doctrine references in PRs cited correctly?

The standup is short. The discipline is the engineering substrate.

---

## 15. Calendar overview

| Phase | Sprints | Calendar weeks | Key deliverable |
|---|---|---|---|
| Phase E | S5–S8 | weeks 1–8 | Charter-conformant; Pack-extracted; v1.0.0 readiness review |
| Phase F | S9–S11 | weeks 9–14 | Audit-grade reproducibility; signed everything |
| Phase G | S12–S14 | weeks 15–20 | OT integration maturity; edge gateway shipped |
| Phase H | S15–S17 | weeks 21–26 | Regulatory catalogue expanded; formal-spec validation |
| Phase I | S18–S20 | weeks 27–32 | AI/ML observable + reproducible |
| Phase J | S21–S22 | weeks 33–36 | SOC 2 / ISO 27001 / AgID / pen-test / v1.0.0 release |

Total: 36 weeks. ~9 months. Aggressive but bounded by the doctrine. Slips beyond Phase J end (calendar week 36) require a re-plan ADR.

---

## 16. Tradeoff Stanza

- **Solves:** the absence of a coherent, doctrine-derived plan from current state to category-leading template; the SaaS-framing residual across multiple documents; the gap between "Mission II PASS" and "v1.0.0 charter-conformant"; the lack of explicit deliverables, sprint-bounds, exit criteria, and ADR triggers for the 140+ new doctrine rules; the unspecified path to OT-integration-maturity and AI/ML-reproducibility and regulatory-catalogue-expansion.
- **Optimises for:** doctrine traceability (every deliverable cites the rules it addresses), sprint-bound execution (every Phase has a tight scope), audit-grade defensibility (every reproducibility property is wired in code), engagement readiness (the lifecycle is exercised end-to-end against a synthetic engagement before client #1), competitive moat (every dimension where we plan to beat the surveyed competitors has a concrete sprint).
- **Sacrifices:** velocity on net-new product features outside the plan deliverables (the 80% capacity allocation is the constraint); ~36 weeks of front-loaded engineering work before the plan's success criteria are evaluable; the optionality to silently re-prioritise without filing a re-plan ADR.
- **Residual risks:** plan slip due to unscoped engagement work (mitigated by 20% reserved capacity); doctrine drift if office hours are skipped (mitigated by quarterly cadence); engagement-team hire delay (mitigated by deferring T2/T3 expansion until the hire lands); v1.0.0 license decision delayed past Phase J Sprint S22 (mitigated by ADR-0053 placeholder).

---

*This plan governs every Sprint between Sprint S5 and Sprint S22. PRs that are not derivable from this plan + the doctrine + the charter are blocked unless a re-plan ADR is filed. The plan is reviewed at six-month intervals; the next review is 2026-10-30 (end of Phase E + start of Phase F).*

---

## 17. Where to start on Day 1 of Sprint S5

The first PR of Sprint S5 should:

1. Add ADR-0021 (Charter and doctrine adoption).
2. Add the conformance test stubs for Rules 1–8 (initially failing where the rule isn't yet enforced; passing where it is).
3. Add `internal/packs/pack.go` and `internal/packs/manifest.go` with the interface and schema.
4. Add `docs/contracts/pack-manifest.schema.json`.
5. Update `Taskfile.yaml` with `task pack-list` and `task pack-validate` helpers.
6. Update CHANGELOG.

The second PR of Sprint S5 should:

1. Add ADR-0023 (Pack-contract interfaces).
2. Add `internal/domain/protocol/ingestor.go`, `internal/domain/reporting/builder.go`, `internal/domain/emissions/factor_source.go`, `internal/domain/identity/provider.go`, `internal/domain/region/profile.go`.
3. Add per-interface example tests.

The third PR of Sprint S5 should rewrite `docs/MODUS_OPERANDI.md` per ADR-0028 (Phase E Sprint S5 deliverable §5.3.2).

These are the first three concrete moves. From there, Sprint S5 deliverables §5.3.4 onward are scheduled per the remaining 7 weekdays.

The plan is lived, not read. Start.

