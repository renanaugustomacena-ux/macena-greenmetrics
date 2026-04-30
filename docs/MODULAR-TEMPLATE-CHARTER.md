# Modular Template Charter — GreenMetrics

> **Status:** Adopted (supersedes the implicit SaaS framing in MODUS_OPERANDI v1)
> **Date adopted:** 2026-04-30
> **Authors:** @ciupsciups (Renan Augusto Macena)
> **Doctrine refs:** Rules 9–68 (existing); Rules 69–88 (Modular Template Integrity, this charter is their evidence base)
> **Review date:** 2026-10-30 (six months — co-scheduled with the doctrine office hours)
> **Required reading before contributing:** `docs/DOCTRINE.md`, `docs/PLAN.md`, this file.

---

## 0. Why this document exists

GreenMetrics was, until 2026-04-30, drafted commercially as a multi-tenant Italian-SME SaaS. The MODUS_OPERANDI document, the COST-MODEL, the THREAT-MODEL section 2, and ADR-0001 all carry the SaaS framing. Several artefacts compute CAC, LTV, churn, and per-meter pricing tiers (€99 Starter / €249 Professionale / Enterprise).

That positioning is hereby retired. The product is now a **modular engagement template**: a deeply opinionated, regulator-grade reference implementation that is offered as a complete deliverable to a client; once a client signs, the engagement begins, and the product is then specialised for that client out of the template — never built from scratch, never deployed unchanged. We monetise the *quality and irreducibility* of the template, the *speed* with which we can specialise it, and the *operability* of the resulting per-client deployment. We do not monetise per-meter consumption.

This charter is the spine. Every other document — the doctrine, the plan, the threat model, the cost model, the modus operandi — is downstream of the choices made here, and is being rewritten or annotated to align. Where a downstream document still carries SaaS vocabulary, it is being treated as "not yet aligned" and is queued in `docs/PLAN.md` Phase E for explicit overhaul rather than silently corrected.

---

## 1. Identity

### 1.1 What GreenMetrics *is*

GreenMetrics is a modular template — a single, coherent, opinionated, regulator-grade industrial energy-management and sustainability-reporting platform — delivered as source, infrastructure, doctrine, runbooks, and audit evidence. It is a **production-shaped reference implementation** of the entire stack that an energy-intensive industrial site needs to:

- ingest meter data from any combination of OT protocols (Modbus RTU/TCP, M-Bus, SunSpec, Pulse, OCPP, IEC 61850 / MMS, MQTT Sparkplug B, OPC UA — see the Pack catalogue in `docs/PROTOCOL-PACKS.md`);
- store time-series at TimescaleDB-native scale (hypertables, continuous aggregates, retention, compression, segment-by);
- compute Scope 1 / Scope 2 / Scope 3 emissions against versioned authoritative factor sources (ISPRA, GSE, Terna, AIB, EcoInvent — see `docs/FACTOR-PACKS.md`);
- generate regulatory dossiers (ESRS E1 for CSRD, Piano Transizione 5.0 attestazione, Conto Termico 2.0 GSE submission, Certificati Bianchi TEE, D.Lgs. 102/2014 audit energetico — see `docs/REPORT-PACKS.md`);
- operate at regulator-grade reliability with full supply-chain attestation (Cosign keyless, SLSA L2 → L3, Kyverno admission, PSS-restricted, NetworkPolicy default-deny, RLS, RBAC, idempotency, audit log immutability, SBOM, vulnerability scanning, signed images);
- be specialised for a specific client without restructuring the template, by adding or replacing **Packs** at well-known seams.

GreenMetrics is the *template*. A *Macena GreenMetrics deployment for client X* is the engagement deliverable.

### 1.2 What GreenMetrics is *not*

It is **not**:

- a multi-tenant SaaS product offered on a price list. There is no signup, no Stripe integration, no per-meter monthly billing, no churn dashboard, no CAC funnel.
- a generic IoT platform. It is opinionated about what a meter reading looks like, what a regulatory dossier looks like, and what a Scope 2 calculation must demonstrate. It refuses to become an everything-platform.
- a Grafana template, a Postgres schema, or an OpenAPI YAML in isolation. It is the *whole* shape — code, schema, IaC, runbooks, doctrine, ADRs, conformance suite, supply-chain trust chain, regulatory pack — together. Pieces are not sold separately.
- a competitor to Schneider EcoStruxure or Siemens Sinergy at the multinational-enterprise scale. Those tools are integrated into automation hardware and sold by global account teams to Fortune-500 buyers; GreenMetrics is the layer above — vendor-neutral, audit-grade, and tractable for a 50-to-2000-FTE industrial company that wants the same capability without the integration tail.
- white-label snake-oil. The template is opinionated and the source is open to the engaged client; brand-removability is a configuration knob (see §6) not a value proposition. We are selling the engineering, not the wrapper.

### 1.3 Brand decision (locked)

The template's brand is **GreenMetrics**. Engagement deployments may rebrand the user-facing surface (UI title, login page, PDF cover-letter, email-from header — see §6 "White-label readiness") without forking the template. The Go module path stays `github.com/greenmetrics/backend` for the canonical template; client forks may use a different module path under `github.com/<client>-greenmetrics/backend` per §5 "Repo topology".

This decision is locked because every engagement we already piloted (the Sprint S1–S4 drops to the reference customer) has been delivered under the GreenMetrics brand and the existing rule references in code (`Doctrine refs: Rule 14`, etc.) tie reviewers to a specific lineage. Renaming the template would break that lineage and orphan the Mission II audit verdict.

---

## 2. The pivot: SaaS framing → Engagement model

| Concept (old, SaaS framing) | Concept (new, engagement model) | Why the change |
|---|---|---|
| Multi-tenant SaaS deployed centrally | Per-engagement single-tenant deployment (default) with optional shared-infrastructure mode for partners hosting many clients | Italian regulator preference; eliminates a class of cross-tenant compromise vectors; aligns commercial responsibility with operational responsibility |
| Subscription pricing per-meter per-month | One-time engagement fee + annual maintenance + customisation services + optional managed-operations retainer | Eliminates the "is this €249/m/month worth it" objection; the buyer is buying an outcome, not a metered service |
| Customer self-serve onboarding | Engineer-led 8–14 week engagement, scoped at signing | Industrial clients do not self-serve; pretending they do creates implementation failure risk |
| CAC, LTV, churn metrics | Engagement margin, time-to-customisation, template-fit-score, net-engagement-value, post-handover support attach rate | Different motion, different KPIs |
| One product roadmap | One *template* roadmap + N *engagement* roadmaps that pull from template | Template velocity is independent of any one client's velocity |
| Code reuse via tenant flags | Code reuse via Pack composition + extension-point seams + upstream sync | Tenant flags don't survive the second client; Packs do |
| Customer support tiers | Engagement playbook tiers: handover-only / co-managed / fully-managed | Each tier has a specific SLA and a specific amount of GreenMetrics-team involvement |
| Dashboard for sustainability lead | The template's dashboards are the floor, not the ceiling — every engagement adds the client's domain views | Treating dashboards as the deliverable is what every competitor does; we don't |
| Compliance as a feature | Compliance as the *core* — every Pack must validate against the formal spec for the regulatory output it produces | The differentiator is regulator-grade evidence, not a checkbox |

The pivot is not a rename. It is a *reorganisation of where value lives in the codebase and where revenue lives in the engagement*. The value lives in the template's *core* (everything that is invariant across engagements) and in the *Packs* (everything that swaps in per engagement). The revenue lives in the engagement (license + customisation + optional managed ops), not in a recurring per-meter charge.

---

## 3. Core vs. Pack — the load-bearing distinction

### 3.1 What is *Core*

Core is everything the template is *as a template*. It is invariant across engagements. A change to core ships in a versioned template release that every engagement may pull, with the upstream-sync discipline of §7. Core is:

- The Go backend binary skeleton (`backend/cmd/server`, `backend/cmd/worker`, `backend/cmd/migrate`, `backend/cmd/simulator`).
- The HTTP framework choice (Fiber v2), the data-access layer choice (pgx/v5), the migration runner (pressly/goose), the validator (go-playground/validator/v10), the breaker (sony/gobreaker/v2), the observability libraries (zap, OTel SDK), the queue (asynq).
- The cross-cutting middleware: request-id, structured logging with redactor, OTel span propagation, panic recovery, body-limit, security headers, CORS allow-list, rate-limit, idempotency, JWT validation with `kid`-pinning, RBAC permission middleware, RLS tenant context wrapper, problem-details error envelope.
- The cross-portfolio invariants enforced in `tests/conformance/`: money-cents+ISO-4217, RFC 3339 UTC timestamps, UUIDv4 tenant IDs, RFC 7807 errors, CloudEvents 1.0, health envelope `{status, service, version, uptime_seconds, time, dependencies}`.
- The TimescaleDB schema for `tenants`, `users`, `meters`, `meter_channels`, `readings` (hypertable), `readings_15min` / `readings_1h` / `readings_1d` (continuous aggregates), `emission_factors` (temporal-key versioned), `reports`, `alerts`, `audit_log`, `idempotency_keys`. Every other Core table is in this schema.
- The OpenAPI 3.1 canonical contract at `api/openapi/v1.yaml` and the oapi-codegen pipeline that produces typed Go server stubs.
- The SvelteKit 2 frontend skeleton: `frontend/src/lib/api.ts` (typed client), `frontend/src/lib/components/` (`ConsumptionHeatmap`, `Scope123Breakdown`, `EnergyChart`, `MeterList`, `ESGScorecard`, `EmissionFactorEditor`, `AlarmFeed`, `CarbonFootprint`), `routes/` (`+layout`, `+page`, `meters/`, `readings/`, `reports/`, `carbon/`, `settings/`), the design-token CSS in `app.css`, the auth `hooks.server.ts`.
- The Grafana provisioning skeleton with two flagship dashboards.
- The Terraform module skeleton in `terraform/modules/{eks, vpc, rds-timescale, secrets, s3, iam-irsa, iam-breakglass, cloudfront-waf, github-oidc}/` and the S3 + DynamoDB state backend.
- The GitOps skeleton: ArgoCD applications, Kyverno admission policies, conftest IaC + K8s + Dockerfile policies, Falco rules, External Secrets Operator wiring.
- The supply-chain trust chain: Cosign keyless signing pipeline, SLSA L2 provenance attest pipeline, SBOM generation via Syft, Trivy image scan post-build, OSV scanner, gitleaks, CodeQL, govulncheck.
- The conformance suite: `tests/conformance/` (cross-portfolio invariants), `tests/security/` (RBAC + RLS + boot refusal), `tests/property/` (algebraic invariants on aggregates and reports), `tests/migrations/` (up-down-up).
- The runbook catalogue at `docs/runbooks/`: db-outage, ingestor-crash-loop, grafana-down, cert-rotation, secret-rotation, jwt-secret-rotation, capacity-spike, region-failover, tenant-data-leak, pulse-webhook-flood, cost-audit. Every Core failure mode has one.
- The ADR catalogue at `docs/adr/0000–0020+`. Each ADR is a permanent rationale for a Core decision.
- The 200+ rule doctrine at `docs/DOCTRINE.md`. The doctrine itself is Core.
- This charter.

### 3.2 What is a *Pack*

A Pack is everything that swaps in per engagement. A Pack is a self-contained directory of code, schema, fixtures, tests, ADR, and Pack-charter under `packs/<pack-id>/` that satisfies a documented Pack contract. A Pack is loaded at boot via configuration; the Core engine refuses to boot without at least the Packs it expects (declared in `config/required-packs.yaml`).

Packs come in five flavours:

1. **Protocol Packs** (`packs/protocol/<name>/`): implement the `Ingestor` interface (EP-01) for a specific OT protocol. The Italian-flagship Pack ships Modbus RTU, Modbus TCP, M-Bus, SunSpec, Pulse, and OCPP. New protocols (IEC 61850, MQTT Sparkplug B, OPC UA, BACnet, EtherCAT EoE, PROFINET, IEC 62056-21 IR optical, KNX, LonWorks-via-bridge) are added as Packs.
2. **Factor Packs** (`packs/factor/<name>/`): implement the `FactorSource` interface (EP-03) for a specific authoritative factor source. The Italian-flagship Pack ships ISPRA Italia (national mix, sectoral defaults), GSE (renewable shares, AIB residual mix), Terna (national mix daily). Other Factor Packs include UK BEIS / DEFRA, EPA US (eGRID), DEFRA / IEA international, EcoInvent connector (license-bring-your-own).
3. **Report Packs** (`packs/report/<name>/`): implement the `Builder` interface (EP-02) for a specific dossier type. The Italian-flagship Pack ships ESRS E1 (CSRD), Piano Transizione 5.0 attestazione, Conto Termico 2.0 GSE, Certificati Bianchi TEE, D.Lgs. 102/2014 audit energetico, monthly consumption, CO2 footprint. Other Report Packs include the SECR UK report, GHG Protocol Corporate Standard, ISO 14064-1 verification, TCFD, IFRS S1/S2, SEC Climate Disclosure.
4. **Identity Packs** (`packs/identity/<name>/`): replace the local-DB-backed authentication with an external identity provider. The Core ships local-DB (bcrypt + pepper + lockout). Identity Packs for SAML 2.0 (CyberArk, Okta, Azure AD, Keycloak), OIDC (Auth0, Okta OIDC, Google), LDAP / Active Directory (AD-FS bridge), Italian SPID, and CIE replace the local provider behind a feature flag.
5. **Region Packs** (`packs/region/<name>/`): bundle factor-source defaults + regulatory-pack defaults + locale + timezone + holiday calendar + currency + privacy-regime overlays. The Italian-flagship Pack is `packs/region/it/` and includes Europe/Rome timezone, EUR comma-decimal locale, ARERA / Garante / GDPR / D.Lgs. 138/2024 NIS2 invariants. Region Packs for DE, FR, ES, PT, RO, AT, CH, GB ship as the engagement pipeline opens those geographies.

Packs are versioned independently of Core. Packs declare a *minimum Core version*, an *exact Pack-contract version*, and a four-part Tradeoff Stanza in their Pack-charter. A Pack that violates the Pack contract is rejected by the Core loader at boot — this is enforced in the conformance suite (`tests/packs/contract_compliance_test.go`).

### 3.3 What is *not* in either Core or Pack

Three kinds of artefact are explicitly outside Core and Pack:

1. **Engagement code** — code that is uniquely useful to one client and to no other. It lives in the client's *fork* of the template, under `engagements/<client>/`, and never lands in upstream main. The fork-discipline policy in §7 governs this.
2. **Vendor-specific drivers** that depend on a non-redistributable vendor SDK (e.g., a closed-source Schneider PowerLogic SDK with a binary blob). These live in a private separately-distributed Pack repository under restricted-access licence and are loaded into an engagement only when the client has the right with the vendor.
3. **Customer-confidential data fixtures and golden files**. These live in the customer's fork, encrypted at rest, never in upstream. The conformance suite uses synthetic fixtures in `tests/fixtures/synthetic/` only.

---

## 4. Pack Contract — the formal seam

A Pack is loaded by Core via a *Pack contract*: a Go interface, a manifest schema, a registration entry, a conformance test, and an ADR. The Pack contract for each EP is documented in `docs/EXTENSION-POINTS.md` and is the reference implementation. The summary version:

```go
// internal/packs/pack.go (formalise in Sprint S5 — see PLAN.md)

type PackKind string

const (
    KindProtocol PackKind = "protocol"
    KindFactor   PackKind = "factor"
    KindReport   PackKind = "report"
    KindIdentity PackKind = "identity"
    KindRegion   PackKind = "region"
)

type PackManifest struct {
    ID                 string   `json:"id" validate:"required,oneof_pack_id"`
    Kind               PackKind `json:"kind" validate:"required"`
    Version            string   `json:"version" validate:"required,semver"`
    MinCoreVersion     string   `json:"min_core_version" validate:"required,semver"`
    PackContractVersion string  `json:"pack_contract_version" validate:"required,semver"`
    Author             string   `json:"author" validate:"required"`
    LicenseSPDX        string   `json:"license_spdx" validate:"required"`
    Notes              string   `json:"notes,omitempty"`
    Capabilities       []string `json:"capabilities" validate:"required"`
    Dependencies       []string `json:"dependencies,omitempty"`
}

type Pack interface {
    Manifest() PackManifest
    // Init runs once at boot; receives a typed handle to the Core surface
    // it needs (logger, metrics, tracer, repo, config). Return error to abort boot.
    Init(ctx context.Context, core CoreHandle) error
    // Register is called after Init to register the Pack's contributions
    // (e.g. an Ingestor, a FactorSource, a Builder, an IdentityProvider).
    Register(reg Registrar) error
    // Health reports the Pack-specific dependency health, surfaced into
    // the /api/health envelope.
    Health(ctx context.Context) PackHealth
    // Shutdown is invoked on graceful shutdown with a 30-second budget.
    Shutdown(ctx context.Context) error
}
```

Per-Pack-kind contracts (`Ingestor`, `FactorSource`, `Builder`, `IdentityProvider`, `RegionProfile`) are defined in `internal/domain/<kind>/`. Each Pack-kind contract is the reference for what implementations must satisfy and is itself frozen by the Pack-contract version.

A Pack registers its contributions through the `Registrar` indirection — never via a global. Core can therefore introspect the loaded Pack set, write `manifest.lock.json` for SLSA provenance, and refuse to boot when the loaded set diverges from `config/required-packs.yaml`. This is the regulator-grade evidence that the system you ran on January 5 is the system you generated the audit dossier from on March 30.

---

## 5. Repo topology

### 5.1 Upstream template repository

`github.com/greenmetrics/template` is the canonical repository. The `main` branch is the head of the template. Tagged releases (`v1.0.0`, `v1.1.0`, …) are the artefacts that engagements pull. `Sprint S5` introduces `release/v1.x` branches for long-lived stabilisation.

The current `github.com/renanaugustomacena-ux/macena-greenmetrics` private repository is the *master* of the template; it will be migrated to `github.com/greenmetrics/template` once the template hits v1.0.0 (target end of Sprint S8). Until then the private repo is the canonical source and engagement forks reference it directly.

### 5.2 Engagement forks

Each engagement is a *fork* of the template repository. Fork naming is `github.com/<engagement-org>/<engagement-id>-greenmetrics`. The engagement fork has:

- a single tag-pinned `template-version.txt` at the repo root recording the template version it last synced from;
- an `engagements/<client>/` directory containing all engagement-specific Packs, code, fixtures, runbooks, and ADRs that should not flow upstream;
- a `CLAUDE.engagement.md` listing the engagement-specific invariants on top of the template doctrine (e.g., the client requires SAML, the client uses Aruba Cloud, the client mandates ISO 27001 evidence on top of Italian compliance);
- the standard `main` branch tracking the upstream `release/v1.x` line via the upstream-sync discipline of §7.

A fork *may* customise Core surfaces, but every Core customisation is recorded in `engagements/<client>/CORE-CUSTOMISATIONS.md` with a Tradeoff Stanza and a sunset date — because Core customisations break upstream sync and must be either upstreamed or reverted within two template releases.

### 5.3 Pack repositories

Public Packs live in the template repository under `packs/<kind>/<id>/`. Private Packs live in `github.com/greenmetrics-packs/<id>` (or in the engagement fork's `engagements/<client>/packs/`). The Pack-loader looks up Packs by manifest ID; it does not care whether the Pack lives in the template, in a sibling repo, or in the engagement.

---

## 6. White-label readiness

The template ships with a single configuration file at `config/branding.yaml` controlling all branding surfaces:

- `product_name`: the name shown in the UI header, the login page, the PDF cover-letters, the email-from header (`product_name <noreply@<deploy-domain>>`).
- `legal_entity`: the legal entity that owns the deployment (`Macena Renan Augusto SRL` for Macena-managed engagements; the client's legal entity for handover deployments).
- `support_contact`: the support address and phone.
- `logo_*`: logo SVG / PNG paths for header, login, PDF cover, favicon, social-card.
- `theme_*`: hex colours for primary, secondary, accent, success, warning, danger.
- `footer_text`: shown at the bottom of every page.
- `pdf_cover_template`: optional path to a customised PDF cover-letter template.

The Frontend `app.css` reads the theme tokens at build time. The PDF report builder reads `pdf_cover_template`. The auth pages and emails read `product_name` / `support_contact`. No hard-coded "GreenMetrics" string remains in the rendered UI; the conformance suite enforces this via `tests/conformance/no_hardcoded_brand_test.go` (see Sprint S5).

White-label is a configuration knob, not a forking exercise. A handover deployment can rebrand to the client's name without touching code. The template's Pack registry continues to identify itself as GreenMetrics in machine-readable surfaces (`/api/internal/version`, OpenAPI `info.x-template-source`); the client decides whether the human-readable surfaces inherit that identity or override it.

---

## 7. Upstream-sync discipline

Engagement forks fall behind the template the day they're created. The longer they stay behind, the more painful the sync. We enforce a *sync cadence* and a *merge-friendliness rule* on Core:

### 7.1 Sync cadence

- Each engagement fork must sync the upstream `release/v1.x` line *at least once per quarter*, or before each new client-visible release of the engagement deployment, whichever is sooner.
- The sync runbook at `docs/runbooks/upstream-sync.md` describes the workflow: pull upstream, run conformance + property + security tests on the merged branch, address Pack-contract version bumps, run the engagement's own integration suite, deploy to staging, soak for 48 hours, deploy to production.
- A fork that has not synced for two consecutive quarters is flagged in the engagement health register; the engagement lead has 30 days to either sync or document the exception with an ADR.

### 7.2 Merge-friendliness rule on Core

A change to Core that breaks Pack contracts or changes the database schema in a non-additive way is a *breaking change*. Breaking changes are batched into major-version releases (template v1 → v2) on a cadence no faster than once per 18 months. Between majors, all Core changes are *additive only* on schemas, hot paths, and Pack contracts.

The engineering manifestation of this rule:

- New columns on Core tables: `NULL` or with `DEFAULT`. No `ALTER COLUMN TYPE`. No `DROP COLUMN`. (See Rule 100 in `docs/DOCTRINE.md`.)
- New methods on Pack-contract interfaces are added with default no-op implementations on a versioned shim until all reference Packs implement them and the Pack-contract version bumps.
- Removed Core code stays as a deprecated re-export with a `Deprecated:` godoc line and a sunset date for at least one minor version.
- The OpenAPI v1 contract is frozen except for additive operations. Additive operations include new endpoints, new optional fields, new optional headers, new response examples. Removed operations require a parallel-run window per RFC 8594 — see `docs/API-VERSIONING.md`.

### 7.3 Conflict resolution

When upstream and engagement diverge on the same surface, the conflict is resolved at sync time per `docs/runbooks/upstream-sync.md`. The default resolution is "upstream wins for Core, engagement wins for engagement-specific overlays." The engagement lead documents non-default resolutions in `engagements/<client>/SYNC-LOG.md`. The synced state is signed off by the engagement lead and a Macena platform-team reviewer.

---

## 8. Engagement lifecycle

### 8.1 Phases

An engagement runs through six phases. They are bounded in calendar time at signing — we do not let engagements drift.

| Phase | Calendar | Output | Gate |
|---|---|---|---|
| 0 — Discovery | 2 weeks | Scope-of-work; Pack matrix; integration map; deployment topology choice | Signed SoW; deployment-topology ADR |
| 1 — Fork & bootstrap | 1 week | Engagement repository; bootstrap successful in client target; first staging deploy | `task verify` green on engagement fork |
| 2 — Pack assembly | 3–6 weeks | Required Region/Protocol/Factor/Report Packs assembled; identity Pack wired; client-specific data fixtures loaded; conformance + property + security tests green on engagement fixtures | All required Pack contracts satisfied; conformance gates green |
| 3 — Customisation sprint | 2–4 weeks | Client-specific UI overlays, custom dashboards, custom report layouts, custom alerts, integration with client SCADA / ERP / SIEM | Client UAT green on the agreed acceptance scenarios |
| 4 — Hardening & soak | 2 weeks | Production deploy; chaos drill; failover drill; capacity test at 1×/3×/5× expected load; runbook walkthrough with the operator team | RPO/RTO drill ≤ contractual targets; chaos drill green; runbook drill green |
| 5 — Handover or co-managed start | 1 week | Operator-team training; runbook handover; on-call rotation arrangement; postmortem template embedded; first ADR by the operator team filed | Operator team self-reports readiness; first 7-day on-call shift owned by operator team or co-managed lead |

Total median: 11 weeks from SoW to operator-owned production. Long-tail engagements with multi-site rollouts run 14–18 weeks.

### 8.2 Engagement tiers

Three tiers, indexed on how much of the operator-team work GreenMetrics keeps after handover:

- **Handover-only (T1)**: GreenMetrics finishes Phase 5 and exits. The client owns operations from Day 0 of production. Annual maintenance contract covers Pack updates and security patches for 24 months. Ideal for clients with a strong internal Platform team.
- **Co-managed (T2)**: GreenMetrics retains shared on-call after handover. SLA ≤ 99.5% on key user journeys. Quarterly review with the operator team. Pack updates and security patches included. Ideal for clients with a thin internal team that wants resilience without rebuilding it.
- **Fully-managed (T3)**: GreenMetrics owns operations end-to-end on a hosting platform of the client's choice (AWS eu-south-1, Aruba Cloud, GCP eur-west, on-prem K3s in the client's DC). SLA ≤ 99.9%. Includes capacity planning, cost optimisation, regulatory-update tracking, audit-evidence pack production. Ideal for clients buying outcomes, not infrastructure.

Pricing model per tier is in the rewritten `docs/MODUS_OPERANDI.md` (in flight, Phase E). The principle is: the price reflects risk transfer and the engagement margin is independent of the deployment's transactional volume.

### 8.3 Engagement closure

An engagement closes (in the commercial sense) when one of three conditions hits:

1. The annual maintenance contract is not renewed and the client has either taken full ownership or migrated to another vendor. The fork stays the client's; the template's upstream is unaffected.
2. The client requests termination. We deliver an exit pack: full source on the version they were running, all data exports, all runbook walkthrough recordings, all keys (rotated to a client-only chain), all dashboards as JSON, all ADRs frozen. We retain only the contract and the lessons-learned pack.
3. The deployment is migrated to upstream-of-template — the engagement-specific code is generalised into a Pack and the fork becomes a thin overlay. This is the success path: the engagement's value flows back upstream as a contribution to the template, raising the template's quality for every future engagement.

A closed engagement does not generate a follow-on request; if it does, that's a new engagement under a new SoW.

---

## 9. Economic model

### 9.1 Revenue lines (replaces SaaS pricing tiers)

- **Engagement license** (one-time): grants the client perpetual right to use the version of the template delivered. Sized at a percentage of the engagement project budget. Typical band €40k–€180k depending on complexity, Pack assembly depth, and number of sites.
- **Customisation services**: T&M billed against the engagement scope. Typical band €60k–€300k for the customisation sprint and Pack assembly. Includes the engineering labour, the Macena team's domain advisory, and the regulatory-pack signoff by an EGE-certified partner where applicable.
- **Annual maintenance** (T1+): includes Pack updates as the regulatory landscape evolves (e.g., ISPRA factor table is republished every April), security patches, and major-version migration assistance. Sized as a percentage of license. Typical 18–22%.
- **Co-managed retainer** (T2): a fixed monthly fee buying named on-call hours, SLA, quarterly reviews. Typical band €4k–€12k/month per deployment.
- **Fully-managed retainer** (T3): a fixed monthly fee buying full operations on a hosting platform of the client's choice. Typical band €18k–€55k/month per deployment.
- **Regulatory-deliverable services** (one-off): the EGE-countersigned Piano 5.0 attestazione, the CSRD ESRS E1 dossier review, the D.Lgs. 102/2014 audit countersignature. Priced as in the previous MODUS_OPERANDI but delivered through the deployment, not as a SaaS feature.

### 9.2 What we deliberately do not sell

- We do not sell per-meter monthly subscriptions. The unit of sale is the engagement; the deployment then handles however many meters the client needs at no marginal-revenue model cost to us.
- We do not sell "use the template free, pay for support." The template is not free under the proprietary licence; the engagement license is the entry charge.
- We do not sell access to a centrally-hosted multi-tenant instance for many SMEs. Multiple SMEs can share a deployment if a partner ESCO or system integrator wants to host them, but that's the partner's monetisation, not ours.
- We do not sell unbundled Packs to third parties. Packs are part of the template's value; selling them à la carte invites quality erosion. A partner that wants a Pack must engage with us; the engagement deliverable then includes the Pack.

### 9.3 Margin profile target

Engagement margin (gross) target ≥ 65% on T1 handovers, ≥ 55% on T2 co-managed, ≥ 45% on T3 fully-managed. The mix shifts toward T2/T3 over the engagement's lifetime. Annual maintenance attaches at ≥ 90% of T1 closures (we deliberately scope out clients that won't take maintenance — they are too risky to support).

---

## 10. Deployment topologies

The template ships supporting four deployment topologies. The chosen topology is locked in the Discovery ADR. Switching topologies after Phase 1 is a re-scoping event.

### 10.1 Topology A — Public-cloud single-tenant

Single AWS account or GCP project, single TimescaleDB primary + 2 streaming replicas, single backend deployment, single frontend deployment, ArgoCD for GitOps, IRSA for IAM, AWS KMS for keys, AWS Secrets Manager + ESO for secrets, AWS CloudFront + ALB + CloudFront-WAF in front of frontend, NLB in front of backend. The default. Italian residency satisfied by `eu-south-1` Milan. Documented in `docs/DEPLOYMENT-TOPOLOGY-A.md` (Sprint S6).

### 10.2 Topology B — Italian-sovereign-cloud single-tenant

Aruba Cloud (Arezzo or Bergamo) or Seeweb or TIM Enterprise. AGID Qualificazione Cloud per la PA satisfied. K3s on bare metal or Aruba's managed Kubernetes. Cert-manager + trust-manager for in-cluster PKI. HashiCorp Vault for secrets (no AWS Secrets Manager equivalent in this topology). Documented in `docs/DEPLOYMENT-TOPOLOGY-B.md` (Sprint S7).

### 10.3 Topology C — On-prem single-tenant

K3s on the client's bare metal in the client's DC. The integration with the client's existing identity (SAML / OIDC / AD) via the Identity Pack is mandatory. Backups go to a client-owned S3-compatible store (MinIO, Wasabi, or the client's NAS). Documented in `docs/DEPLOYMENT-TOPOLOGY-C.md` (Sprint S8).

### 10.4 Topology D — Hybrid (cloud frontend + on-prem ingestion)

Backend ingestion deployed on-prem to remain inside the OT segment; the frontend / reporting deployed in public cloud and connected via a site-to-site VPN with strict segmentation. Suitable for clients whose OT/IT separation policy forbids cloud egress from the OT zone. Documented in `docs/DEPLOYMENT-TOPOLOGY-D.md` (Sprint S9).

---

## 11. Tenancy model under the new framing

In the SaaS framing, multi-tenancy was the default and the heart of the threat model. In the engagement framing, single-tenant is the default and multi-tenant is an *opt-in* for partner-hosted deployments.

The implementation rule:

- Every Core surface continues to be tenant-aware. The `tenant_id` UUIDv4 invariant, the `WHERE tenant_id = $1` repository filters, the RLS policies, the JWT claim, the audit-log row keying — all stay. They are *defence in depth* against bugs even in single-tenant deployments.
- A single-tenant deployment runs with one bootstrapped tenant. The Region Pack's `bootstrap_tenant_id` is materialised at first migration. The frontend's "switch tenant" surface is hidden by feature flag.
- A multi-tenant deployment turns the feature flag on. The threat model, the RBAC matrix, and the RLS policies remain the same — they were always defence in depth. Multi-tenant is not a switch from "no isolation" to "isolation"; it is a switch from "one tenant in a defended box" to "many tenants in defended boxes."
- The `THREAT-MODEL.md` Section 2 (which currently calls this "a multi-tenant SaaS") will be rewritten to lead with "single-tenant by default, multi-tenant by configuration" — see Phase E, Sprint S5.

---

## 12. Compatibility commitments

### 12.1 Template versions

The template uses semantic versioning. `v1.0.0` is the first stable release; we hit it at the end of Sprint S8 once the engagement-fork model is complete and at least one client has run the lifecycle end-to-end. Until then we ship `v0.x` releases — every Pack-contract change is documented but no compatibility guarantee is made.

After v1.0.0:

- *Patch* releases (`v1.0.x`): bug fixes, documentation, dependency bumps. No Pack-contract change.
- *Minor* releases (`v1.x.0`): additive changes to Core, Pack contracts, schema, OpenAPI. Must be merge-friendly per §7.2.
- *Major* releases (`v2.0.0`): breaking changes batched. Cadence ≤ once per 18 months. Migration playbook published 90 days before the release. Both `v1` and `v2` supported in parallel for 12 months.

### 12.2 Pack-contract versions

Pack contracts version independently of Core. A Pack declares a `pack_contract_version`; Core maintains a *supported window* of Pack-contract versions in `internal/packs/contracts.go`. A new Core release that deprecates a Pack-contract version emits a `pack-contract-deprecated` event in the audit log on every Pack load until the engagement upgrades the Pack.

### 12.3 Database schema versions

The schema version follows the migration sequence number. Engagements running an older Core can never have a newer schema; the migration runner refuses to up-migrate beyond the `min_core_version` declared in the migration's leading comment.

### 12.4 OpenAPI versions

`v1` of the OpenAPI contract is the only supported version. A `v2` is introduced only when a breaking change is unavoidable; the parallel-run window is 12 months minimum. See `docs/API-VERSIONING.md`.

### 12.5 What we do not promise

We do not promise that a fork that has *customised Core* (vs. just adding Packs) will sync cleanly across template versions. The merge-friendliness rule applies to the template. A fork that overrides Core is responsible for resolving conflicts at sync time per §7.3. The fork-discipline policy explicitly discourages Core overrides; engagement-specific behaviour belongs in a Pack.

---

## 13. What this charter explicitly forbids

The following are rejected without ADR (and rejection authority lives in §3 of the doctrine, Rules 26 / 46 / 66, plus Rules 69–88 of the Modular Template Integrity group):

- Any code change that adds a per-meter usage metric to the billing table (we don't bill on volume).
- Any code change that adds a "Stripe customer ID" field anywhere in Core schema (we don't have a Stripe integration; if a partner needs payments for their hosted deployment, that's their Pack).
- Any code change that registers a tenant via a self-serve `/signup` endpoint without an engagement-id reference (we don't self-serve).
- Any code change that hard-codes "GreenMetrics" outside `config/branding.yaml` defaults (the conformance test enforces this).
- Any code change that removes the tenant scoping from a repository method (defence in depth stays even in single-tenant).
- Any code change that bakes a country, regulator, factor source, or report shape directly into `internal/services/` or `internal/handlers/` rather than into a Pack (Italian-flagship code that exists today is being lifted into `packs/region/it/` in Phase E).
- Any documentation change that re-introduces SaaS pricing-tier vocabulary without a charter-supersession ADR.
- Any change that removes a doctrine rule from `docs/DOCTRINE.md` without an unrejection ADR per the doctrine's rule-rotation process.

---

## 14. Open decisions, deferred

The following are flagged as *deferred* — they require evidence we do not yet have. They are tracked as ADR-stub issues in `docs/adr/PROPOSED.md` and revisited at the next quarterly charter review.

1. **Whether to open-source the template eventually.** Currently proprietary all-rights-reserved (LICENSE). The licence carries a future-open-source clause. Likely candidate licences if we open-source: BUSL-1.1 with a Change Date of 4 years (delays competitor commercialisation while preserving long-term openness); SSPL-1.0 (overlap with the MongoDB / Elastic motion); or AGPLv3 (strongest network-use clause). Decision deferred to v1.0.0 launch.
2. **Whether to ship a vendor-neutral SaaS-edition of the template (we host it for many small clients ourselves)** in addition to the engagement model. This would re-introduce a (controlled) SaaS surface for the long-tail of small clients who can't afford an engagement. Decision deferred to year 2; would require a separate revenue line and a separate operations team.
3. **Whether the Italian Region Pack stays in this repository or splits into `github.com/greenmetrics/pack-region-it`.** Splitting reduces the upstream-template surface area but adds release coordination cost. Decision deferred to v1.0.0; default is keep in-tree for now.
4. **Whether to support a non-TimescaleDB datastore as a Pack** (e.g., a Pack swap to ClickHouse or InfluxDB-3 for clients who refuse PG-anything). Significant Core surface impact. Decision deferred until a real client requests it.
5. **Whether to add an MQTT Sparkplug B Protocol Pack as a Core-tier Pack** (i.e., shipped in the upstream template). MQTT Sparkplug B is the de facto modern OT protocol but adds an MQTT broker dependency. Decision deferred to Sprint S12 when the OT integration discipline rules (Rules 109–128) are solidified.

---

## 15. Glossary

| Term | Meaning |
|---|---|
| **Template** | The canonical upstream `github.com/greenmetrics/template` repository at a given semantic version. |
| **Core** | The invariant code, schema, doctrine, and infra of the template. |
| **Pack** | A self-contained directory under `packs/<kind>/<id>/` implementing a Pack contract. |
| **Pack contract** | The Go interface, manifest schema, and conformance test that every Pack of a given kind must satisfy. |
| **Engagement** | A contracted client deployment of the template. |
| **Engagement fork** | A `github.com/<client-org>/<client-id>-greenmetrics` private fork of the template, with `engagements/<client>/` overlays. |
| **Region Pack** | The Pack that bundles factor-source defaults, regulatory-pack defaults, locale, timezone, and privacy-regime overlays for a country. |
| **Protocol Pack** | The Pack that implements the `Ingestor` interface for an OT protocol. |
| **Factor Pack** | The Pack that implements the `FactorSource` interface for an authoritative factor source. |
| **Report Pack** | The Pack that implements the `Builder` interface for a regulatory dossier shape. |
| **Identity Pack** | The Pack that replaces the Core local-DB authentication with an external IdP. |
| **Topology A/B/C/D** | The four supported deployment topologies (public-cloud, sovereign-cloud, on-prem, hybrid). |
| **Tier T1/T2/T3** | The three engagement tiers (handover, co-managed, fully-managed). |
| **Charter supersession** | The mechanism by which an ADR overrides a charter clause. Requires explicit `supersedes: docs/MODULAR-TEMPLATE-CHARTER.md §X` in the ADR header. |
| **White-label** | The capacity to rebrand the rendered surfaces without forking the template. Implemented via `config/branding.yaml`. |
| **Upstream sync** | The quarterly merge of the template's `release/v1.x` branch into an engagement fork. |

---

## 16. Cross-references

- **Doctrine:** `docs/DOCTRINE.md` (200+ rules, 60 inherited verbatim, 140+ new).
- **Plan:** `docs/PLAN.md` (the work to take the template from current state to Charter-conformant v1.0.0 and beyond).
- **Competitive brief:** `docs/COMPETITIVE-BRIEF.md`.
- **MODUS_OPERANDI:** `docs/MODUS_OPERANDI.md` (rewritten in Phase E to remove SaaS framing; queue tracked in PLAN Phase E Sprint S5).
- **ADR-0001:** `docs/adr/0001-platform-doctrine-adoption.md` (annotated in Phase E Sprint S5 to replace "regulated-industry SaaS" with "regulated-industry engagement template").
- **Threat model:** `docs/THREAT-MODEL.md` (Section 2 rewritten in Phase E Sprint S5).
- **Cost model:** `docs/COST-MODEL.md` (rewritten in Phase E Sprint S5 — engagement margin replaces SaaS gross margin).
- **Service catalog:** `docs/SERVICE-CATALOG.md` (annotated in Phase E Sprint S6).
- **Platform defaults:** `docs/PLATFORM-DEFAULTS.md` (annotated in Phase E Sprint S6).
- **Runbooks:** `docs/runbooks/upstream-sync.md` (new in Phase E Sprint S5).
- **Extension points:** `docs/EXTENSION-POINTS.md` (Pack-contract formalisation in Phase E Sprint S5).
- **Abstraction ledger:** `docs/ABSTRACTION-LEDGER.md` (Pack-loader entry added in Phase E Sprint S5).

---

## 17. Tradeoff Stanza

- **Solves:** the gap between "we built a Mission-II PASS reference implementation" and "we sell it to multiple clients without becoming a maintenance-bound SaaS company"; the absence of an explicit Pack-vs-Core architecture; the SaaS framing that locks us into per-meter economics that don't match the buyer's procurement reality.
- **Optimises for:** engagement margin, time-to-customisation, regulator-grade evidence per deployment, template-fit-score across multiple clients, upstream-sync hygiene, and the engineering team's ability to evolve the template without breaking client forks.
- **Sacrifices:** the simplicity of a single-product-roadmap mindset; the optionality to silently embed client-specific code in core; the marketing simplicity of a per-meter price page; the easy story of self-service signup; ~2 sprints of velocity in Phase E to extract Italian compliance into a Pack and rewrite the SaaS-flavoured docs.
- **Residual risks:** Pack-contract versioning bugs that break client forks at sync time (mitigated by the conformance suite + 12-month parallel-run); engagement scaling becomes labour-intensive (mitigated by the Pack catalogue absorbing more of each new engagement's surface); a client demands a feature that's against the charter (mitigated by §13 and the rejection authority); the rewriting of the SaaS-flavoured docs is partial and inconsistencies leak into client conversations (mitigated by the Phase E sprint scope and the conformance test on doc-vocabulary).

---

*This charter governs every other artefact in the repository at and after Sprint S5. PRs that violate the charter without a charter-supersession ADR are blocked. The charter is reviewed at six-month intervals; the next review is 2026-10-30.*
