# GreenMetrics Doctrine

> **Status:** Adopted; supersedes the implicit doctrine references scattered across ADRs, runbooks, and the (off-machine) plan file at `~/.claude/plans/my-brother-i-would-flickering-coral.md`.
> **Date adopted:** 2026-04-30
> **Authors:** @ciupsciups (Renan Augusto Macena)
> **Charter parent:** `docs/MODULAR-TEMPLATE-CHARTER.md`
> **Review date:** 2026-10-30 (six months — co-scheduled with the charter office hours)
> **Required reading before contributing:** `docs/MODULAR-TEMPLATE-CHARTER.md`, this file in full, `docs/PLAN.md`.

---

## How to read this document

This is the canonical, normative engineering doctrine for GreenMetrics. It carries **210 rules** organised in eleven groups. Rules **1–8** are universal invariants inherited from the cross-portfolio CLAUDE.md. Rules **9–28** are Platform Engineering. Rules **29–48** are Advanced Backend Engineering. Rules **49–68** are DevSecOps Engineering. Rules **69–88** are Modular Template Integrity. Rules **89–108** are Data Provenance & Audit-Grade Reproducibility. Rules **109–128** are OT Integration Discipline. Rules **129–148** are Regulatory Pack Discipline. Rules **149–168** are Engagement Lifecycle. Rules **169–188** are Cryptographic Invariants. Rules **189–208** are AI/ML Reproducibility. Rules **209–210** are the doctrine's own meta-rules (rotation, supersession).

Every rule has the same shape:

- A **title** in the form *Rule N — Imperative or invariant statement*.
- A **body** explaining the rule.
- A **Why** paragraph explaining the failure mode that motivates the rule (the rule exists because of *something concrete that broke or could break*, not because of taste).
- A **How to apply** paragraph explaining where in the code or in the operational practice the rule lives — so a reviewer can verify compliance without re-deriving the rule.
- **Cross-refs** to ADRs, runbooks, REJECTED entries, conformance tests, charter clauses, and other rules that interact with this one.

A rule is *binding* unless it is explicitly superseded by an ADR carrying a `supersedes-rule: NNN` header and the `override-allowed` PR label. A rule is *invariant* if its body says so explicitly — invariants do not accept supersession; they require an unrejection ADR per Rule 210.

PRs that violate any binding rule without an override are blocked at code review and at the conformance-suite gate. The conformance suite (`backend/tests/conformance/`) implements the mechanically-checkable subset of the doctrine and grows with every new rule that admits a test (most do).

The doctrine is *interdependent*. Rule 14 (contract first) underpins Rule 34 (validator-driven binding) and Rule 54 (policy as code). Rule 33 (data is the system) underpins Rule 39 (RLS) and Rule 35 (idempotency). You cannot cherry-pick — adopting a subset creates incoherent gates. ADR-0001 records this decision and is itself subject to the doctrine.

---

## Table of contents

- [Section 0 — Universal invariants (Rules 1–8)](#section-0--universal-invariants-rules-18)
- [Section 1 — Platform Engineering (Rules 9–28)](#section-1--platform-engineering-rules-928)
- [Section 2 — Advanced Backend Engineering (Rules 29–48)](#section-2--advanced-backend-engineering-rules-2948)
- [Section 3 — DevSecOps Engineering (Rules 49–68)](#section-3--devsecops-engineering-rules-4968)
- [Section 4 — Modular Template Integrity (Rules 69–88)](#section-4--modular-template-integrity-rules-6988)
- [Section 5 — Data Provenance & Audit-Grade Reproducibility (Rules 89–108)](#section-5--data-provenance--audit-grade-reproducibility-rules-89108)
- [Section 6 — OT Integration Discipline (Rules 109–128)](#section-6--ot-integration-discipline-rules-109128)
- [Section 7 — Regulatory Pack Discipline (Rules 129–148)](#section-7--regulatory-pack-discipline-rules-129148)
- [Section 8 — Engagement Lifecycle (Rules 149–168)](#section-8--engagement-lifecycle-rules-149168)
- [Section 9 — Cryptographic Invariants (Rules 169–188)](#section-9--cryptographic-invariants-rules-169188)
- [Section 10 — AI/ML Reproducibility (Rules 189–208)](#section-10--aiml-reproducibility-rules-189208)
- [Section 11 — Doctrine Meta-rules (Rules 209–210)](#section-11--doctrine-meta-rules-rules-209210)

---

## Section 0 — Universal invariants (Rules 1–8)

These rules are inherited from the portfolio-level `CLAUDE.md` and are *invariants* in the strict sense: no PR, no ADR, and no engagement may relax them. They are checked by `backend/tests/conformance/` on every CI run and by Kyverno admission policies on every cluster admit.

### Rule 1 — Money is `(amount_cents int64, currency ISO-4217 string)`

Every monetary quantity in code, schema, request, response, log, event, or report is represented as an integer-cent count plus an ISO-4217 currency code. Float and decimal-string representations are forbidden. Mixed-currency arithmetic is forbidden — the value-object package `internal/domain/money/` rejects addition of differing currencies at compile-time via a method receiver.

**Why.** Floating-point arithmetic loses precision on cents in ways that break Italian SDI invoice reconciliation, GSE tax-credit computation, and CSRD ESRS E1 financial-effect disclosures. A €0.01 rounding error multiplied across 10 000 readings becomes a €100 audit finding. EnEx (energy-exchange) markets settle in micro-Euro and exact cents; we have to be exact.

**How to apply.** Use `internal/domain/money/money.go`'s `Money` value object. The repository layer marshals to two columns (`*_amount_cents bigint`, `*_currency_iso char(3)`). The OpenAPI types model money as `{amount_cents: integer, currency: string}` not `number`. Conformance test: `tests/static/no_float_money_test.go` (S5).

**Cross-refs.** REJ-13. ADR-cross-portfolio invariants. Charter §3.1.

---

### Rule 2 — Timestamps are RFC 3339 UTC with explicit offset

Every timestamp in code, schema, request, response, log, event, audit row, or report carries a `+00:00` offset in serialised form and is stored in UTC in the database. Local-timezone strings, naive `time.Time` without location, Unix-epoch integers in user-visible surfaces, and `varchar`-with-format-string columns are all forbidden.

**Why.** Italian regulatory reports cite Europe/Rome local-time wall-clock for human-readable cover pages but require UTC offsets in the underlying data so that DST transitions and cross-border rollups are unambiguous. CSRD audit reproducibility breaks the moment a timestamp is stored without offset and re-rendered against a different server timezone.

**How to apply.** Postgres columns are `timestamptz`. Go writes `time.Time` already in UTC (`.UTC()` in marshalling). Frontend renders local-time only via `Intl.DateTimeFormat` with `timeZone: 'Europe/Rome'`. Conformance test: `tests/conformance/rfc3339utc_test.go` (S5).

**Cross-refs.** Charter §3.1. Schema-evolution policy §1.

---

### Rule 3 — Tenant identifiers are UUIDv4

Every `tenant_id` is a UUIDv4 generated by the bootstrap migration or by `crypto/rand`-backed generation in identity-provisioning code paths. Sequential or business-meaningful tenant identifiers are forbidden. The UUIDv4 invariant lets RLS policies, audit-log row keys, and cross-microservice CloudEvents subjects agree without coordination.

**Why.** A `tenant_id` that is sequential leaks the count of tenants. A `tenant_id` derived from a business name (e.g., `verona-meccanica-srl`) breaks the moment the legal entity is renamed. UUIDv4 is the schelling point for opaque, globally-unique tenant identity that survives re-orgs, mergers, and rebrands.

**How to apply.** Use `github.com/google/uuid` `uuid.New()`. Postgres column type `uuid not null`. Conformance test: `tests/conformance/uuidv4_test.go` (S5).

**Cross-refs.** Charter §3.1. Rule 12 (RLS). Rule 39.

---

### Rule 4 — Errors are RFC 7807 Problem Details

Every API error response is an RFC 7807 Problem Details JSON (`application/problem+json`) with `type`, `title`, `status`, `detail`, `instance`, and a `code` extension. No bare-string errors, no `{error: "..."}` shorthand, no Rails-style `{errors: [...]}` arrays.

**Why.** A consistent error envelope lets clients (the SvelteKit console, the mobile app, integrators' tools, the auditor's CSV exporter) parse failures uniformly. Rails-style ad-hoc envelopes break automation. Bare-string errors break i18n. Stripe's, Twilio's, and the IETF's choice of RFC 7807 has hundreds of integration-years of evidence behind it.

**How to apply.** Use `internal/handlers/errors.go`'s typed helpers (`BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`, `Conflict`, `Unprocessable`, `TooManyRequests`, `Internal`, `Unavailable`). The middleware stamps `Content-Type: application/problem+json`. Conformance test: `tests/conformance/rfc7807_test.go` (S5).

**Cross-refs.** Charter §3.1. Rule 26. Rule 46.

---

### Rule 5 — Events are CloudEvents 1.0

Every published event (audit-log fan-out, ingestion-completed signal, alert-fired notification, report-generation-completed callback) is wrapped in a CloudEvents 1.0 envelope with `specversion`, `type`, `source`, `id`, `time`, `datacontenttype`, `subject`, and the typed `data` payload. Bare JSON payloads on the wire are forbidden.

**Why.** Events outlive the system that produced them. A consumer five years from now needs to know the event's type, schema, source, and time without re-engineering. CloudEvents is the IETF/CNCF standard for this; reinventing the envelope is a Rule-26 rejection.

**How to apply.** Event schemas live in `docs/contracts/events/` and are validated against `cloudevents-spec-1.0.2.json` in CI. The dispatcher in `internal/services/event_bus.go` (Phase E Sprint S6) wraps every payload. Conformance test: every fixture in `docs/contracts/events/` is round-trip-validated in `tests/contracts/cloudevents_test.go` (S5).

**Cross-refs.** Charter §3.1. Rule 22.

---

### Rule 6 — Health envelope shape is fixed

Every health endpoint (`/api/health`, `/api/ready`, `/api/live`) returns the envelope `{status, service, version, uptime_seconds, time, dependencies}`. `dependencies` is an object keyed by short-name with values `{status, latency_ms, last_checked, message}`. The envelope shape is part of the cross-portfolio invariant and may not be extended without a portfolio-level ADR.

**Why.** Operations tooling (uptime monitors, ALB health checks, CloudWatch synthetics, runbook scripts) depends on a single shape. A backend that drifts from the envelope creates breakage in every operational surface that consumes it. The shape is documented and enforced precisely because operations is the use case that cannot tolerate variation.

**How to apply.** `internal/handlers/health.go` builds the envelope. The OpenAPI schema `Health` constrains it. Conformance test: `tests/conformance/health_test.go` (S5).

**Cross-refs.** Rule 40. SLO doc §1.

---

### Rule 7 — Logs are structured JSON with mandatory context

Every log line is a structured JSON object with mandatory fields: `service`, `env`, `version`, `commit`, `request_id`, `trace_id`, `span_id`, `tenant_id`, `level`, `time`, `message`. `fmt.Sprintf`-style log messages with embedded data are forbidden — fields are first-class.

**Why.** Logs are queried by Loki / CloudWatch / OpenSearch. A log line that lacks `trace_id` cannot be joined to a trace. A log line that lacks `tenant_id` cannot be filtered to a specific engagement. A log line that embeds data in the message string cannot be aggregated.

**How to apply.** zap `Logger` is the only logger; obtained via `obs.Logger(ctx)` so context fields propagate. The redactor strips `password`, `token`, `secret`, `authorization`, `jwt`, and `api_key` before emit. Custom golangci-lint rule blocks `log.Printf` and `fmt.Print*` in non-`cmd/` packages. Conformance test: `tests/static/log_format_test.go` (S5).

**Cross-refs.** Rule 40. Platform-defaults §2. AB-05.

---

### Rule 8 — Italian residency is the default; cross-EU transfer is opt-in only

The default deployment is hosted in `eu-south-1` Milan or in an Italian-sovereign-cloud topology (Aruba, Seeweb, TIM Enterprise). Cross-EU data transfer requires an explicit Order-Form clause, an Article 46 GDPR mechanism (SCCs 2021 or an approved Code of Conduct), and a per-tenant `data_region_allowed` allowlist enforced at the replication layer. Cross-Atlantic transfer is forbidden by default; opt-in requires charter-supersession ADR.

**Why.** ARERA's directives on smart-meter data sharing, the Garante's 2022 Google-Analytics ruling, AgID's Qualificazione Cloud per la PA, NIS2 D.Lgs. 138/2024 essential-entity supply-chain obligations — all converge on Italian-residency-as-default for the buyer profile we serve. A cross-Atlantic transfer that wasn't explicitly opted into is a regulator-grade incident.

**How to apply.** Terraform variable `aws_region` defaults to `eu-south-1`. Topology B (Aruba) is the alternative. The replication layer in `internal/services/replication.go` (Phase F Sprint S10) checks the per-tenant allowlist before each cross-region read.

**Cross-refs.** ADR-0007. Charter §10. Rule 8 below interacts with Rules 169–188 (key residency).

---

## Section 1 — Platform Engineering (Rules 9–28)

The Platform Engineering rules govern how the substrate (Kubernetes, IaC, GitOps, identity, observability, supply chain) is built and operated. They presume a single Macena platform team owning the substrate; they survive team growth without rewrite. Rule numbering is preserved verbatim from the original 60-rule plan; rule text is reconstructed from the references scattered across the codebase, runbooks, and platform-defaults.

### Rule 9 — Platform engineering is a discipline, not a label

A platform team exists when there is a documented charter, a documented scope, a documented authority, and a documented termination criterion (Rule 28). A team that calls itself "platform" without those four artefacts is a service team in disguise. The charter for GreenMetrics' platform team is `docs/TEAM-CHARTER.md`. The scope and authority are in the same file. The termination criterion is "the operator team can run the substrate without us for 30 consecutive days."

**Why.** "Platform" without a charter becomes "the team that owns whatever no other team owns" — a backlog graveyard. With a charter, the team has explicit refusal authority.

**How to apply.** New platform initiatives follow `docs/PLATFORM-INITIATIVE-WORKFLOW.md` (Rule 11/31/51 sequence). Initiatives without that paperwork are blocked.

**Cross-refs.** ADR-0001. Charter §3.1. Rule 28.

---

### Rule 10 — Platform serves users; users are application teams

The platform team's customer is the application team, not the end user. The output is leverage — a single Pack contract instead of N bespoke ingestors, a single ArgoCD application instead of N Kubernetes manifests, a single conformance test instead of N hand-rolled checks. If a platform initiative does not deliver leverage to an application team, it is not a platform initiative — it is a hobby project.

**Why.** Many platform teams drift into building tools for themselves. The leverage criterion forces a per-initiative answer: who is the application-team customer, what is the leverage, and how do we know they got it.

**How to apply.** Every platform initiative's PR description names the application-team customer and the measurable leverage delivered. Office-hours quarterly review (`docs/PLATFORM-PLAYBOOK.md`) re-scores every initiative against this criterion.

**Cross-refs.** Rule 9. Rule 17.

---

### Rule 11 — Initiatives follow the Rule 11/31/51 sequence

Every new platform initiative goes through three gates in order: Rule 11 (problem identification — written, scoped, customer-named), Rule 31 (architecture proposal — ADR with Tradeoff Stanza, named alternative, named rejection), Rule 51 (operational readiness — runbook, metrics, alerts, chaos drill plan, rollback plan). An initiative that skips a gate is rejected at the next gate.

**Why.** The most expensive platform mistakes are infrastructure that solves the wrong problem (Rule 11 skipped), infrastructure that fights the architecture (Rule 31 skipped), or infrastructure that the operator team cannot run (Rule 51 skipped). The sequence is the cheapest possible insurance against these mistakes.

**How to apply.** `docs/PLATFORM-INITIATIVE-WORKFLOW.md` is the canonical workflow. PRs implementing a platform initiative without the three artefacts are blocked.

**Cross-refs.** Rule 31. Rule 51. ADR template.

---

### Rule 12 — Platform thinking is opinionated defaults

The platform team picks one stack and says "no" to the rest. One renderer (Kustomize). One GitOps engine (ArgoCD). One queue (Asynq + Redis). One factor source per region by default. One identity provider per deployment. Optionality without justification is rejected.

**Why.** Optionality has a per-option compounding cost. Two queues = two failure domains, two backup procedures, two on-call playbooks, two skill investments. The leverage from picking one is enormous; the cost of supporting two is rarely justified.

**How to apply.** `docs/PLATFORM-DEFAULTS.md` lists the chosen stack. A divergence requires an ADR and a Tradeoff Stanza naming what the divergence buys and what it costs.

**Cross-refs.** REJ-25, REJ-26, REJ-32, REJ-33, REJ-34. ADR-0006.

---

### Rule 13 — Abstractions are cost centres until proven otherwise

A new abstraction enters `docs/ABSTRACTION-LEDGER.md` carrying its hidden complexity, its lost flexibility, who pays the cost, and a trigger to remove it. Abstractions are added when ≥ 2 implementations already exist or are in flight (Rule 26 rejection authority is the gate). Abstractions are removed when only one implementation remains for >2 quarters.

**Why.** Premature abstraction is the most common cause of Go codebases' slow rot. An interface with one implementation is dead weight that future changes have to navigate around. The ledger forces an explicit cost-benefit at the moment the abstraction is introduced.

**How to apply.** `docs/ABSTRACTION-LEDGER.md` is the source of truth. PRs adding a new abstraction must add a row. PRs removing an abstraction must move the row to the "Removed" history.

**Cross-refs.** ADR-0013. ADR-0014. ABS-01 through ABS-16.

---

### Rule 14 — Contracts come first; code is the implementation

The OpenAPI 3.1 contract at `api/openapi/v1.yaml` is hand-written and is the source of truth. Go server stubs are generated from it (oapi-codegen). Schemas for events and config are JSON Schema in `docs/contracts/`. The Pack contracts in `internal/packs/` are typed Go interfaces. Every contract is versioned and validated in CI.

**Why.** The most expensive coordination failures we see in industrial-software work are contract-first failures: an integrator builds against a spec that didn't match the implementation, and discovery happens at the customer site at 3 AM. Contract-first means the spec is the truth and the implementation is the test.

**How to apply.** `task openapi:lint` is green on every PR. `make openapi-bundle` regenerates Go stubs. `make pack-contract-check` (Sprint S5) validates Pack manifests against the schema. Conformance test: `tests/contracts/v1_compat_test.go` for OpenAPI; `tests/contracts/event_schema_test.go` for events.

**Cross-refs.** ADR-0013. Rule 34. Rule 54.

---

### Rule 15 — Layers are the system map; the system map is the architecture

The repository's mental model lives in `docs/LAYERS.md` (Mermaid diagram) and `docs/layers.yaml` (data). Five layers: Infrastructure → Substrate → Backend → Frontend → Operators. Every component fits in exactly one layer. Cross-layer dependencies are explicit.

**Why.** A team without a system map argues from anecdote. A team with a system map argues from structure. Mission-II audit pivoted on the system map being legible to a third-party reviewer in under an hour.

**How to apply.** `make layers-doc` regenerates the diagram from the YAML. PRs that add a component update the YAML. Mermaid in `docs/LAYERS.md` is generated, not hand-edited.

**Cross-refs.** ADR-0001 §References. Mission-II audit findings.

---

### Rule 16 — Infrastructure is code; state is centralised but partitioned

Every cloud resource is declared in `terraform/`. State lives in S3 with DynamoDB locking and a per-environment partition (REJ-08). One root state per environment is the SPOF/scaling-axis violation; multiple roots per env are sprawl. The `terraform/envs/{dev,staging,prod}/` layout is the canonical answer.

**Why.** Terraform monorepo with one state breaks at scale; per-component states without coordination break under change. Per-environment roots are the Pareto-optimal point.

**How to apply.** S3 + DynamoDB backend in `terraform/versions.tf`. Bootstrap module in `terraform/bootstrap/`. State-file encryption with KMS. PR checks: `terraform fmt`, `terraform validate`, `tflint`, `tfsec`, `checkov`, `conftest test --policy policies/conftest/terraform/`.

**Cross-refs.** REJ-08. ADR-cross-portfolio.

---

### Rule 17 — Developer experience is a first-class platform output

Bootstrap from a fresh clone to a working development stack is ≤ 5 minutes. The bootstrap script (`scripts/dev/bootstrap.sh`) is idempotent. The Devcontainer (`/.devcontainer/`) carries the full toolchain. The Taskfile is the canonical CLI surface. `task --list` discoverable.

**Why.** A platform that takes a day to bootstrap loses a day every time anyone joins. The leverage compounds: every minute saved on bootstrap is a minute spent on feature work.

**How to apply.** `task bootstrap` runs end-to-end on a fresh laptop. The bootstrap drill is run quarterly; failures are platform-team Sev-2 incidents.

**Cross-refs.** `docs/DX.md`. `Taskfile.yaml`. `docs/CLI-CONTRACT.md`.

---

### Rule 18 — Telemetry sampling is policy, not arbitrary

OpenTelemetry sample ratio is 0.1 in production by default. Errors (non-2xx responses) are head-based-sampled at 100%. Critical paths (`/reports`, `/auth/login`) override to 1.0. Configuration is in `internal/config/config.go`, not magic numbers.

**Why.** Trace storms cost money and drown the signal in noise. Trace under-sampling drops the rare failure that's the only one you needed. Sample-ratio policy must be deliberate.

**How to apply.** Sample-ratio constants are in code. Override per-route via the `OTEL_SAMPLE_OVERRIDE` env (KV pairs). Sample-ratio change requires an ADR.

**Cross-refs.** REJ-19. ADR-0006. Rule 40.

---

### Rule 19 — Sentinel detection refuses to boot in production

Production builds refuse to boot if any sentinel placeholder secret is present (`change-me`, `default-pass`, `your-jwt-secret`, the `null`-byte sentinel). The `internal/config/config.go` `Load()` function checks every secret against a deny-list and returns an error. The error is *not* a warning — it's a refusal.

**Why.** Audit-only sentinels become production sentinels at the speed of "we'll fix it later." A refusal at boot is the only enforcement that survives the rush before a launch.

**How to apply.** `config.go:176-194` carries the deny-list. Conformance test: `tests/security/sentinel_refusal_test.go` (S2 done; S5 hardens).

**Cross-refs.** REJ-18. ADR-0001.

---

### Rule 20 — Secrets are managed; never in code, never in K8s

Secrets live in the cloud provider's KMS-backed secret store (AWS Secrets Manager + ESO; Vault for Topology B). Kubernetes Secrets are *materialised* references via External Secrets Operator, not source-of-truth. Plaintext secrets in PRs are gitleaks-blocked at pre-commit.

**Why.** A secret in Git is forever. A secret in a K8s Secret YAML is everywhere a backup of that YAML reaches. A secret in Secrets Manager is rotatable, audited, and KMS-encrypted.

**How to apply.** ESO in `gitops/base/external-secrets/`. Pre-commit gitleaks. CI gitleaks. Rotation runbook `docs/runbooks/secret-rotation.md`.

**Cross-refs.** ADR-0003. ADR-0017 (Cosign keyless OIDC). Rule 178.

---

### Rule 21 — Evolution and change management are first-class

Every breaking change carries a parallel-run window per RFC 8594. Every schema change goes through `docs/SCHEMA-EVOLUTION.md`. Every API version bump follows `docs/API-VERSIONING.md`. The CHANGELOG is updated in the same PR as the change.

**Why.** Industrial customers have multi-year roadmaps; a backend that breaks them mid-cycle is a contract breach. Change management is the engineering rule that prevents accidental breakage.

**How to apply.** `docs/SCHEMA-EVOLUTION.md` is the schema policy. `docs/API-VERSIONING.md` is the API policy. CHANGELOG entries are PR-template-required.

**Cross-refs.** REJ-07 (bespoke migration). ADR-0005 (goose). Rules 100, 102, 145.

---

### Rule 22 — Events are the integration surface; not RPC

Cross-component integration uses CloudEvents over a chosen transport (Asynq + Redis for in-cluster; webhook for cross-org). Cross-component RPC is forbidden between Backend and any Pack — Packs subscribe to events and emit events. The event-bus indirection in `internal/services/event_bus.go` (Phase E Sprint S6) is the seam.

**Why.** RPC tightly couples the lifecycle of caller and callee; a slow callee makes the caller slow. Events decouple. The integration tax of events (schema discipline, idempotent consumers) is paid once and amortised.

**How to apply.** Pack contracts emit / consume events; not synchronous calls into Core. `tests/contracts/event_schema_test.go` validates every event schema.

**Cross-refs.** Rule 5. ADR-0014 (asynq). REJ-32 (no GraphQL gateway).

---

### Rule 23 — The substrate is operable by a single on-call

A single on-call engineer can handle every Sev-1 in a 12-hour shift using only the runbooks. If a runbook requires a vendor's premium support to execute, the runbook is broken. If a runbook requires more than two screens of context, the runbook is broken.

**Why.** On-call burns out engineers. The cheapest insurance is to keep the runbooks fast and self-contained. The Italian-SME engagement-model commitment to RPO 1h / RTO 4h is achievable only if the on-call doesn't have to wake up two more people to execute the recovery.

**How to apply.** `docs/runbooks/` carries one runbook per failure mode. Quarterly runbook drill. Runbook-walkthrough completion is a Phase 5 (Handover) gate.

**Cross-refs.** `docs/ON-CALL.md`. `docs/RUNBOOK.md`. Rule 51. Rule 167.

---

### Rule 24 — Verification is continuous and automatic

`task verify` runs the full quality bar — pre-commit, unit, integration, property, security, conformance, policy, openapi-lint, security-scan — in under 8 minutes on a developer laptop. CI runs the same plus pact + DAST + chaos-light + image-sign + image-scan. A PR that doesn't pass `task verify` doesn't merge.

**Why.** Verification that you have to remember to run never gets run. Verification that runs in 30 minutes never gets run before a PR opens. Sub-8-minute verification on a laptop is the schelling point that makes the discipline survive the rush.

**How to apply.** `Taskfile.yaml` `verify` target. CI mirror. The 8-minute target is enforced by the test budget; tests that exceed it are split.

**Cross-refs.** `docs/QUALITY-BAR.md`. Rule 44. Rule 64.

---

### Rule 25 — Quality threshold is regulator-grade

The threshold for a Core change to merge is the same threshold a regulator would apply to a SOC2 audit: signed artefacts, attested provenance, RBAC enforced, RLS enforced, idempotency enforced, audit log immutable, conformance suite green. The `docs/QUALITY-BAR.md` enumerates the eleven non-negotiables.

**Why.** Industrial customers' procurement processes already include SOC2-equivalent due diligence. A template that doesn't meet that bar can't be sold. Designing to the bar from day 0 is cheaper than retrofitting.

**How to apply.** `docs/QUALITY-BAR.md` §11 — "the line": the reviewer can produce the CSRD audit evidence pack on demand using only the artefacts listed there. If any is missing or stale, the bar is broken.

**Cross-refs.** Rule 45 (backend mirror). Rule 65 (DevSecOps mirror).

---

### Rule 26 — Rejection authority is named and exercised

The platform team has the authority to reject patterns that violate the doctrine. Rejected patterns enter `docs/adr/REJECTED.md` with rule citation, alternative, residual risk, and review date. PRs that propose a rejected pattern are blocked unless accompanied by an unrejection ADR (Rule 209).

**Why.** Without a named authority, every PR re-litigates the doctrine. With a named authority, the doctrine is reviewable in one place.

**How to apply.** `docs/adr/REJECTED.md` is the canonical rejection log. The `make audit-rejections` script scans new ADRs for rejected-pattern keywords. Quarterly office hours walk every rejection past its review date.

**Cross-refs.** Rule 46. Rule 66. REJ-01 through REJ-35.

---

### Rule 27 — Decisions carry the four-part Tradeoff Stanza

Every ADR ends with a Tradeoff Stanza naming what the decision *Solves*, what it *Optimises for*, what it *Sacrifices*, and what *Residual risks* remain. Markdownlint enforces the structure. An ADR without the stanza is rejected at PR review.

**Why.** Decisions without a stanza age badly: future maintainers can't tell whether a residual risk is acceptable now or whether it was always a known compromise. The stanza is the engineering equivalent of the medical-record SOAP note.

**How to apply.** ADR template `docs/adr/0000-template.md` includes the stanza. PR template prompts for ADR linkage. Markdownlint custom rule (`markdownlint-rule-tradeoff-stanza`) verifies the four headings are present.

**Cross-refs.** Rule 47. Rule 67. Every ADR.

---

### Rule 28 — Termination criterion is named at the outset

Every platform initiative carries a termination criterion: when this initiative is done, what does done look like? "We have a working ArgoCD" is not a termination criterion. "ArgoCD reconciles every application within 60s of a Git push, alarms on drift, and is operable by the operator team without our help for 30 consecutive days" is.

**Why.** Initiatives without termination criteria become permanent overhead.

**How to apply.** Every platform-initiative ADR includes a "Termination criterion" section. Quarterly office hours re-scores every initiative against its criterion; initiatives that have hit it are formally closed.

**Cross-refs.** Rule 48 (backend mirror). Rule 68 (DevSecOps mirror). Rule 9.

---

## Section 2 — Advanced Backend Engineering (Rules 29–48)

The Advanced Backend Engineering rules govern the Go backend. They presume Fiber + pgx + TimescaleDB + Asynq as the chosen stack (Rule 12 / 43) and pgrade-grade discipline as the bar (Rule 25 / 45).

### Rule 29 — Backend code is intent-revealing

Variable, function, type, and package names reveal *domain* intent (Rule 32) before *technical* intent. `EmissionFactorSource` not `FactorRepo`. `ReportBuilder` not `Generator`. `MeterIngestor` not `Worker`. The compiler doesn't care; the next reader does.

**Why.** A name that reveals intent is a comment that survives refactoring. A name that hides intent forces the reader to re-derive the intent from the body, every time. Multiplied across ~8000 LOC and a 5-year maintenance horizon, the cost is enormous.

**How to apply.** Code review rejects opaque names. Renamings happen in PRs scoped to renaming.

**Cross-refs.** Rule 32. REJ-17.

---

### Rule 30 — Backend is a first-class system, not a side-effect of the framework

The framework is Fiber; the system is GreenMetrics. Fiber-specific code is confined to `internal/handlers/` and `cmd/server/`. `internal/services/`, `internal/repository/`, `internal/domain/` know nothing about Fiber. Switching frameworks (REJ-34 prevents this on whim, but the day may come) is local to two packages.

**Why.** Frameworks come and go; domains stay. A backend that bleeds framework concerns into domain code is a backend that becomes the framework's hostage.

**How to apply.** Static-analysis rule: no `import "github.com/gofiber/fiber/v2"` outside `internal/handlers/` and `cmd/`. Conformance test: `tests/static/no_fiber_in_domain_test.go` (S5).

**Cross-refs.** REJ-34. Rule 43.

---

### Rule 31 — Architecture proposals carry a named alternative

Every backend ADR names at least one alternative considered and explains why it was rejected. "We chose X" without "we considered Y because Z, but rejected Y because W" is not an ADR — it's a bullet point.

**Why.** A decision without a named alternative is a non-decision. Future maintainers cannot reverse it without reconstructing the alternatives from scratch.

**How to apply.** ADR template enforces the section. PR review enforces the discipline.

**Cross-refs.** Rule 11. Rule 27. Every ADR.

---

### Rule 32 — Domain-driven design lives in `internal/domain/`

Aggregates, value objects, and domain services live in `internal/domain/<aggregate>/`. Repository interfaces are co-located. Anaemic-domain anti-patterns (data classes without behaviour) are rejected. The current backend is mid-migration: `internal/services/` holds today what `internal/domain/` will hold by Sprint S8.

**Why.** Domain-driven structure makes the *meaning* of the code legible. Auditors who don't speak Go can read a `domain/reporting/` directory and grasp the regulatory landscape.

**How to apply.** Migration plan in `docs/PLAN.md` Phase E. Per-aggregate sub-directory under `internal/domain/`. Every aggregate carries a README.

**Cross-refs.** REJ-11 (god service). ADR-0014. Rule 29.

---

### Rule 33 — Data is the system

Schema outlives code. A schema mistake is a 3-year mistake. The TimescaleDB schema is the canonical statement of what the system *is*. Goose migrations in `backend/migrations/` are forward-only in production. CAGGs are not edited in place. Compression and retention policies are part of the schema, not a forgotten ops task.

**Why.** Code can be rewritten in a sprint; data cannot. Every regulatory output (ESRS E1, Piano 5.0, ENEA audit) is reconstructed from the schema; if the schema is shaky, the regulatory outputs are shaky.

**How to apply.** `docs/SCHEMA-EVOLUTION.md` is the policy. ADR-0005 chose goose. ADR-0010 chose hypertable space partitioning. Every migration carries a lock-impact analysis.

**Cross-refs.** Rule 99 (data lineage). Rule 100 (additive only). Rule 145 (regulatory schema).

---

### Rule 34 — Contract-first applies inside the backend too

The OpenAPI spec is the source of truth for the HTTP API; oapi-codegen derives Go types. CloudEvents schemas are the source of truth for events; codegen derives Go types. Pack contract interfaces are the source of truth for Packs; the Pack manifest validates against the schema. *Inside* the backend, contracts apply at module boundaries: `internal/repository/` exposes typed interfaces, `internal/services/` consumes them, no `interface{}` returns.

**Why.** A backend without internal contracts is a backend where every refactor is dangerous. With contracts, a refactor is local.

**How to apply.** Custom golangci-lint rule: no `interface{}` / `any` returns from `internal/repository/`. Conformance test: every public type in `internal/domain/` has a corresponding test in `internal/domain/<aggregate>/<aggregate>_test.go`.

**Cross-refs.** REJ-16. ADR-0013. Rule 14.

---

### Rule 35 — Consistency and state guarantees are explicit

Idempotency is enforced via the `idempotency_keys` Timescale hypertable + 24h retention + middleware in `internal/handlers/idempotency.go`. Repeated submission of the same `Idempotency-Key` with the same request hash returns the original response; conflict on differing hash returns 422. State transitions on `reports`, `alerts`, `meters`, `users` are enumerated and enforced in code, not implicit.

**Why.** Industrial integration partners retry aggressively. A backend that doesn't dedupe is a backend that double-fires GSE submissions, double-creates audit-log rows, and double-bills. None of those are recoverable without operator-grade incident response.

**How to apply.** `internal/handlers/idempotency.go` is the canonical middleware. Every state-change endpoint requires `Idempotency-Key`. Every state machine is expressed as a transitions map; the repository's `UpdateState` method validates against it.

**Cross-refs.** ADR-0014. Rule 41. Rule 124.

---

### Rule 36 — Failure is normal; the system is designed for failure

External calls fail. Database connections drop. Disk fills. The dependency tree rots. Every external integration carries a circuit breaker (Rule 36 + ADR-0009 sony/gobreaker). Every retry is jittered exponential backoff capped at attempts (Rule 36 + cenkalti/backoff). Every long-running job is async (REJ-21 + ADR-0014 asynq). Every stateful resource has a leak detector (`tests/leak/`).

**Why.** Industrial environments have legacy networks, ageing meters, intermittent VPN, vendor portals that go down on Italian holidays. A backend that doesn't expect failure becomes the failure.

**How to apply.** `internal/resilience/` has breaker + retry + http helpers. Every external client uses them. Conformance test: `tests/resilience/breaker_smoke_test.go` (S5).

**Cross-refs.** ADR-0009. ADR-0015 (bounded queue). Rule 41 (concurrency). Rule 76.

---

### Rule 37 — Performance is a design constraint, not a tuning task

Performance budgets are stated upfront and validated in CI: P95 read < 500ms (`SLO.md` §1), P99 ingest < 1s, dashboard queries < 250ms, report generation P99 < 2.5s on a 1-year window. k6 perf bench (`tests/load/`) runs against staging weekly with 1×/3×/5× production-shape; budget assertions block deploys.

**Why.** Performance retrofits are 10× more expensive than performance design. Industrial dashboards that lag 5s feel broken; industrial dashboards that lag 250ms feel responsive. The buyer notices.

**How to apply.** `docs/SLO.md` §1 carries the budgets. `tests/load/` carries the k6 scripts. CD pipeline gates on perf-bench-pass. Async report generation (REJ-21) is the structural commitment to keep reads fast.

**Cross-refs.** REJ-21. ADR-0014. Rule 119 (ingest rate). Rule 124.

---

### Rule 38 — Scalability is intentional, not magical

Scale-up axes are named and prioritised: tenants (multi-tenancy or multi-deployment), meters per tenant (channel cardinality), readings per meter per minute (sample rate), report generation parallelism. Each axis has a documented headroom (`docs/CAPACITY.md`). Hitting headroom triggers a planned scale-out, not an emergency.

**Why.** Scaling without intention produces hacks. Hacks accumulate. Eventually a redesign is unavoidable. A capacity model with named axes makes scale-out a plan, not a panic.

**How to apply.** `docs/CAPACITY.md` is the capacity model. Quarterly capacity review. Hitting 80% on any axis triggers a scale-out ADR.

**Cross-refs.** ADR-0010 (hypertable space partitioning). REJ-09 (no multi-region active-active in Phase 1). Rule 100. Rule 119.

---

### Rule 39 — Security is structural, not bolted on

Multi-tenancy is enforced at three layers: repository-level `WHERE tenant_id = $1` (defence in depth), Postgres RLS policies (Rule 39 + ADR-0011 + REJ-12), JWT-claim-driven middleware (`InTxAsTenant`). RBAC is middleware-enforced (`internal/security/rbac.go`). JWT KID rotation is automated (`docs/JWT-ROTATION.md` + ADR-0016). Sensitive fields are field-level encrypted (`readings.raw_payload` AES-256-GCM, KMS-wrapped DEK).

**Why.** A single missing `WHERE` clause in a hot-path query is a regulator-grade tenant data leak. Defence in depth means the bug-density of any one layer is not the system's bug-density.

**How to apply.** RLS migration `00098_rls_enable.go`. RBAC permission registry. Conformance: `tests/security/rls_test.go`, `tests/security/rbac_test.go`.

**Cross-refs.** ADR-0002. ADR-0011. ADR-0016. REJ-12. Rule 169 (crypto).

---

### Rule 40 — Observability is structural

Every request carries `X-Request-ID`, propagated as a baggage item via OpenTelemetry. Every span carries `tenant.id` attribute (Rule 40 + Platform-defaults §3). Every log line carries the mandatory fields (Rule 7). Every metric has an explicit cardinality budget. Every alert has `runbook_url` annotation pointing to `docs/runbooks/<name>.md`.

**Why.** Observability that lacks request-id propagation is observability that breaks at the first integration boundary. Observability with unbounded cardinality is a budget bomb.

**How to apply.** `internal/observability/` is the canonical wiring. `monitoring/prometheus/rules/` carries alert definitions with `runbook_url`. `tests/conformance/runbook_link_test.go` (S5) verifies every alert links to an existing runbook.

**Cross-refs.** ADR-0006. Rule 7. Rule 18.

---

### Rule 41 — Concurrency is bounded and explicit

Goroutines are bounded (worker pools via `panjf2000/ants`; errgroups for coordinated work). Channels are bounded (the ingest queue is bounded via ADR-0015 + REJ-23 + AB-08). Backpressure is observable (Prometheus counter on drops). No unbounded `go func()` in any package.

**Why.** Unbounded goroutines under load become OOM. Unbounded channels become unbounded latency. The cost of a global concurrency review is high; the cost of writing bounded code from the start is low.

**How to apply.** Custom golangci-lint rule: bare `go func()` outside `cmd/` is flagged. Code review enforces the pool/errgroup pattern. AB-09 documents the worker pool decision.

**Cross-refs.** REJ-23. ADR-0015. Rule 36.

---

### Rule 42 — Resources have lifecycles; lifecycles are managed

Every resource (HTTP body, file handle, DB connection, timer, goroutine, breaker registration, OTel exporter, queue connection) is acquired with a deferred release. Long-running resources (pools, exporters, queue connections) are owned by the `cmd/server/main.go` lifecycle and shut down gracefully (≤ 30s budget). The leak-detector tests in `backend/tests/leak/` are part of `task test:integration`.

**Why.** A long-running backend that leaks 1KB per request leaks 1GB per million requests. We hit million-request days routinely. A goroutine that doesn't exit on shutdown blocks deploys.

**How to apply.** `cmd/server/main.go` carries the canonical lifecycle. `internal/observability/` exposes the `Shutdown(ctx)` for OTel. The leak suite runs in CI.

**Cross-refs.** Rule 36. Rule 41.

---

### Rule 43 — Framework and infrastructure awareness is explicit

The chosen stack is documented and is the default for new code. Diverging from the stack requires an ADR. The current stack: Fiber v2 (REJ-34 keeps it), pgx/v5 (REJ-35 rejects ORM), pressly/goose v3 (ADR-0005), go-playground/validator/v10 (ADR-0012), oapi-codegen (ADR-0013), sony/gobreaker/v2 (ADR-0009), cenkalti/backoff/v5 (resilience), zap (logging), OpenTelemetry SDK (tracing), Prometheus client (metrics), asynq (queue ADR-0014), Redis (asynq backend).

**Why.** A framework choice is a long-term commitment. Drift creates two flavours of every concern (two HTTP frameworks, two queues, two loggers). The cost compounds.

**How to apply.** `docs/PLATFORM-DEFAULTS.md` enumerates the stack. PRs adding a new dependency justify it against an existing alternative.

**Cross-refs.** REJ-25, REJ-26, REJ-32, REJ-34, REJ-35. ADR-0005, ADR-0009, ADR-0012, ADR-0013, ADR-0014.

---

### Rule 44 — Testability is a design output

Every aggregate has a unit test (≥ 90% line coverage in `internal/domain/`). Every handler has a handler-level integration test against a testcontainers Timescale. Every external integration has a contract test (Pact-style for Pack contracts, schema-validation for events). Every algorithm with algebraic invariants has property-based tests (`backend/tests/property/`).

**Why.** Untested code is unreviewable. Tests are the executable specification of intent. Property-based tests catch the cases unit tests miss.

**How to apply.** Coverage gates in CI. Property tests in `backend/tests/property/`. Integration tests in `backend/tests/`. Per-PR test bar in `docs/QUALITY-BAR.md` §2.

**Cross-refs.** Rule 24. Rule 64. Rule 192.

---

### Rule 45 — Backend quality threshold is regulator-grade

The backend mirror of Rule 25. The eleven non-negotiables in `docs/QUALITY-BAR.md` apply: cross-portfolio invariants enforced, code coverage ≥ 80% line, ≥ 90% domain, mutation kill rate ≥ 70% on `internal/domain/`, no `panic()` on user input, no `interface{}` repository returns, no god services, gosec/CodeQL clean, sentinel refusal at boot, RLS+RBAC enforced, body limit set, constant-time webhook verify.

**Why.** A backend that doesn't meet the quality bar can't pass an audit. Audits are the moment of truth.

**How to apply.** `docs/QUALITY-BAR.md`. Conformance suite. CI gates.

**Cross-refs.** Rule 25. Rule 65. Mission-II audit.

---

### Rule 46 — Backend rejection authority is exercised

The backend mirror of Rule 26. Rejected backend patterns enter `docs/adr/REJECTED.md` (REJ-11 through REJ-23). The `make audit-rejections` script catches accidental reintroduction. Code review enforces.

**Why.** Without rejection authority, every PR re-litigates the doctrine. With rejection authority, the doctrine is reviewable in one place.

**How to apply.** `docs/adr/REJECTED.md`. Backend-specific rejections REJ-11 through REJ-23.

**Cross-refs.** Rule 26. Rule 66.

---

### Rule 47 — Decision rationale travels with the decision

Every non-trivial backend decision is an ADR. The ADR carries the four-part Tradeoff Stanza (Rule 27). The ADR is linked in the PR description. The ADR is reviewed at six months and at one year for residual-risk update.

**Why.** Decisions without rationale rot. The next reviewer (in 18 months) cannot tell whether the decision is still valid without the rationale.

**How to apply.** ADR template. PR template prompts for ADR linkage. `docs/adr/` is browsable; `docs/adr/README.md` is the index.

**Cross-refs.** Rule 27. Rule 67. Every ADR.

---

### Rule 48 — Backend termination criterion is named

The backend's termination criterion (in the Rule-28 sense) is "the engagement-team operator can run the backend on the chosen topology without our help for 30 consecutive days." This is the Phase-5 (Handover) gate of every engagement. Until that bar is hit, the engagement is not closed.

**Why.** A backend without a termination criterion becomes permanent overhead for the platform team.

**How to apply.** Engagement Phase 5 gate. Quarterly engagement-health review. Backend ADRs that fail to name a termination criterion are rejected.

**Cross-refs.** Rule 28. Rule 68. Charter §8.

---

## Section 3 — DevSecOps Engineering (Rules 49–68)

The DevSecOps Engineering rules govern the unified CI/CD/security/compliance posture. They presume a single team owning all three concerns (Rule 50) and a regulator-grade bar (Rule 65 / 25 / 45). The numbering and topic structure are preserved verbatim from the original 60-rule plan.

### Rule 49 — DevSecOps is a discipline, not a label

A DevSecOps team exists when there is a documented charter, a documented scope, a documented authority, and a documented termination criterion (Rule 68). The charter is `docs/SECOPS-CHARTER.md`. Without those four artefacts, "DevSecOps" is a slogan.

**Why.** "DevSecOps" without a charter becomes either a security team that signs off on PRs or a CI team that runs Trivy. Neither is the integrated discipline the doctrine requires.

**How to apply.** `docs/SECOPS-CHARTER.md` carries the charter. New DevSecOps initiatives follow Rule 11/31/51 sequence. Quarterly DevSecOps review.

**Cross-refs.** Rule 9. Rule 50.

---

### Rule 50 — DevSecOps is a unified system

Security cannot be a post-step. CI cannot be a separate concern from security. Rollouts cannot be a separate concern from observability. The unified system means: every PR runs the same gates (lint, test, sign, attest, scan, deploy-canary, observe). One team owns the whole pipeline; one set of dashboards observes the whole pipeline.

**Why.** A security team that signs off on PRs after CI runs adds days of latency. A CI team that runs Trivy without owning the deployment gating adds findings that no one acts on. The unified system is the only design that survives at scale.

**How to apply.** `.github/workflows/` carries the unified pipeline. `docs/PIPELINE-MAP.md` is the canonical map.

**Cross-refs.** Rule 56. Rule 64.

---

### Rule 51 — DevSecOps initiatives carry operational readiness

Every DevSecOps initiative includes runbook, metrics, alerts, chaos drill plan, rollback plan. An initiative that ships without these artefacts is rejected at the Rule-51 gate. The runbook is in `docs/runbooks/<initiative>.md`. Metrics are in `monitoring/prometheus/rules/`. Alerts have `runbook_url` annotation.

**Why.** A pipeline that lacks runbooks is a pipeline that fails opaquely. Operator readiness is the difference between a 5-minute outage and a 5-hour outage.

**How to apply.** Rule 11/31/51 sequence in `docs/PLATFORM-INITIATIVE-WORKFLOW.md`. PR template prompts.

**Cross-refs.** Rule 23. Rule 11.

---

### Rule 52 — Identity in CI/CD is keyless and ephemeral

CI authenticates to AWS via GitHub OIDC, never with long-lived access keys. Cosign signs keylessly via Sigstore Fulcio + OIDC, never with stored keys. Image registry pushes are short-lived JWTs. The `terraform/modules/github-oidc/` module is the trust anchor.

**Why.** Long-lived CI credentials are the most common breach vector in the 2023–2025 supply-chain incidents (CircleCI 2023, GitHub-Actions third-party-Action incidents 2024). Keyless OIDC eliminates the credential-storage attack surface.

**How to apply.** GitHub-OIDC Terraform module. Cosign keyless. Per-job claim binding (`repo:ref:branch`). Conformance: `tests/static/no_aws_access_key_test.go` (S5).

**Cross-refs.** ADR-0017. REJ-27 (Cosign with custom KMS). Rule 178.

---

### Rule 53 — Pin every dependency to a digest

Every GitHub Action is SHA-pinned with a `# vX.Y.Z` comment. Every Docker base image is digest-pinned. Every Helm chart is version-pinned (a chart pin is a version, not a `latest`-tag). Every Go module is `go.mod`-pinned with `go.sum` integrity. Every npm package is `package-lock.json`-pinned.

**Why.** The `tj-actions/changed-files` 2025 incident pattern is the canonical example: a tag was rewritten to point to a malicious commit, and unpinned consumers got pwned. SHA-pinning is the only defence.

**How to apply.** Dependabot weekly bumps with `# vX.Y.Z` comment auto-update. Custom workflow lint: `actionlint` rejects unpinned `uses:`. Dockerfile pre-commit: `hadolint` rejects unpinned `FROM`.

**Cross-refs.** REJ-29. ADR-0017.

---

### Rule 54 — Policy as code is the gating mechanism

Conftest policies live in `policies/conftest/{k8s,dockerfile,terraform}/` and run at pre-commit + CI. Kyverno policies live in `policies/kyverno/` and run at admission. Falco rules live in `policies/falco/` and run at runtime. Every policy carries a rationale comment and a doctrine-rule citation.

**Why.** Policy as code is the difference between "we should" and "we do." A Kyverno admission policy that denies unsigned images is the only enforcement that survives the rush before a launch.

**How to apply.** `policies/` is the canonical home. PR template prompts policy-impact. Conformance: `policies/conftest/k8s/test/` carries unit tests for the policies themselves.

**Cross-refs.** ADR-0001. Rule 55. Rule 14.

---

### Rule 55 — Policy gates are layered: IaC → Manifest → Admission → Runtime

Defence in depth applies to policy: Conftest at IaC layer (Terraform), Conftest at K8s manifest layer (kustomize render), Kyverno at admission layer (cluster-time), Falco at runtime layer (host events). A miss at one layer is caught at the next. No layer is the only defence.

**Why.** Each layer has a different threat model. IaC catches mistakes before deploy; manifest catches mistakes before admission; admission catches mistakes before runtime; runtime catches mistakes that escaped policy. The stack is what makes the system regulator-grade.

**How to apply.** `docs/PIPELINE-MAP.md` shows the layers. `policies/` carries the policies for each.

**Cross-refs.** Rule 54. ADR-0019 (Falco vs Tetragon).

---

### Rule 56 — CD gates are layered, not single-step

The CD pipeline is layered: build → sign → SBOM-attest → SLSA-provenance-attest → image-scan → policy-check → deploy-canary → analyse-SLO-burn → promote → observe-post-deploy. A manual approval is one of the gates, not the only gate (REJ-24).

**Why.** Manual approval as the only gate is a bottleneck masquerading as control. Layered automated gates with a single human signoff at the end give both speed and oversight.

**How to apply.** `.github/workflows/` carries the layers. ArgoCD Rollouts AnalysisTemplate reads SLO burn-rate. `docs/ROLLBACK.md` describes the rollback path.

**Cross-refs.** REJ-24. ADR-0004 (ArgoCD). Rule 64.

---

### Rule 57 — Supply chain is end-to-end attested

Every production image is Cosign-signed (keyless OIDC), SLSA-L2-provenance-attested, SBOM-attested (CycloneDX via Syft), and Trivy-scanned. The Kyverno admission policy refuses unsigned/unattested images. Every dependency upgrade triggers a re-attest. The trust chain is documented in `docs/SUPPLY-CHAIN.md`.

**Why.** The 2020–2024 supply-chain incidents (SolarWinds, Codecov, ua-parser-js, Log4Shell, MOVEit, XZ-utils) all share the same shape: an unattested artefact in a trusted spot. Cosign + SLSA + Kyverno + Trivy is the canonical defence.

**How to apply.** `.github/workflows/supply-chain.yml`. `docs/SUPPLY-CHAIN.md`. SLSA L3 plan in ADR-0018.

**Cross-refs.** ADR-0017. ADR-0018. REJ-25, REJ-26, REJ-27.

---

### Rule 58 — SBOM is generated, attested, and queried

CycloneDX SBOM via Syft on every image build. SBOM attached as Cosign attestation. SBOM uploaded to a queryable store (DependencyTrack instance for Phase F). Vulnerability findings (Trivy + osv-scanner + govulncheck + CodeQL) cross-referenced against SBOM.

**Why.** "Are we vulnerable to CVE-X?" is a regulator question after every incident (Log4Shell, OpenSSL 3.x). The answer time is the difference between a 24-hour disclosure and a 30-day fire drill.

**How to apply.** Syft in `.github/workflows/supply-chain.yml`. DependencyTrack in `gitops/base/dependency-track/` (Phase F Sprint S10). SBOM-query runbook `docs/runbooks/sbom-query.md` (Phase F).

**Cross-refs.** Rule 57. Rule 59.

---

### Rule 59 — Vulnerability response is SLA'd

CVSS ≥ 9.0 (critical): patch within 48 hours. CVSS 7.0–8.9 (high): patch within 14 days. CVSS 4.0–6.9 (medium): tracked, patched in next minor release. Below 4.0: tracked. The SLA is documented; the dashboard tracks compliance.

**Why.** Vulnerability response without an SLA becomes "we'll get to it." The SLA is the engineering equivalent of a fire-evacuation plan: you don't wait for the fire to figure out the route.

**How to apply.** `docs/SUPPLY-CHAIN.md` §SLA. Dashboard in Grafana `vuln-response-dashboard.json` (Phase F). PR template for vuln response.

**Cross-refs.** Rule 58. Rule 65.

---

### Rule 60 — Pen-test cadence is annual + post-major-change

Annual third-party penetration test by an OSCP-certified team. Post-major-change re-test (a major change is a Pack-contract version bump, a new Identity Pack, a new Topology). Findings tracked in `docs/PENTEST-CADENCE.md` and remediated against the Rule-59 SLA.

**Why.** A pen-test once a year catches the issues that automated scans miss. A pen-test never catches the regressions introduced by major changes.

**How to apply.** `docs/PENTEST-CADENCE.md` carries the schedule. Findings in a private repo. Remediation tracked.

**Cross-refs.** Rule 59. Rule 65.

---

### Rule 61 — Incident response is rehearsed

NIS2 D.Lgs. 138/2024 timelines apply: 24-hour early warning, 72-hour initial notification, 30-day final report. Annual tabletop exercise with ACN coordination. Postmortem template enforces the structure (timeline, contributing factors, action items, owner, due date). Postmortem within 5 business days for any Sev-1.

**Why.** Incident response without rehearsal is incident response with surprises. NIS2 timelines are not negotiable; missing the 24-hour mark is a regulator-grade failure.

**How to apply.** `docs/INCIDENT-RESPONSE.md`. `docs/runbooks/` per-failure-mode. Tabletop logged in `docs/INCIDENTS-LOG.md` (Phase F).

**Cross-refs.** Rule 167. `docs/COMPLIANCE/NIS2.md`.

---

### Rule 62 — Audit logging is append-only and immutable

The `audit_log` table is append-only at the schema level (REVOKE UPDATE/DELETE on the role used by the application). The application never issues UPDATE/DELETE on audit_log. The 00099_audit_lock.sql migration enforces this. WORM-mirroring to S3 Object Lock (compliance mode) for legal-hold resilience in Topology A.

**Why.** An audit log that can be updated is not an audit log; it's a logbook. Immutability is the property regulators trust.

**How to apply.** Migration `00099_audit_lock.sql`. Conformance test: `tests/security/audit_lock_test.go` (S5).

**Cross-refs.** Rule 145. Rule 169 (signed audit chain). ADR-0011.

---

### Rule 63 — Compliance evidence is exportable

The CSRD audit evidence pack, the Piano 5.0 attestazione bundle, the D.Lgs. 102/2014 audit dossier, the GDPR DSAR export, the NIS2 incident-notification template — all are exportable from the running system without manual surgery. The exports are deterministic (Rule 89) and signed (Rule 169).

**Why.** Auditors don't have time for ad-hoc data pulls. Manual export is the path that produces the wrong answer at the wrong time.

**How to apply.** `docs/COMPLIANCE/` per-regime. Per-regime export endpoint. Per-regime fixture-replay test in `tests/conformance/`.

**Cross-refs.** Rule 89. Rule 145. Charter §1.1.

---

### Rule 64 — Verification is continuous, not periodic

The full verification suite (`task verify`) runs on every PR. The CD pipeline runs the verification suite on every merge to main. Quarterly chaos drill. Quarterly DR drill. Semi-annual upstream-sync drill (for engagement forks). Annual pen-test.

**Why.** Verification on a release cadence creates a backlog of known-broken state. Continuous verification eliminates the backlog.

**How to apply.** CI on every PR. Quarterly drills logged in `docs/CHAOS-LOG.md`. Annual drills in `docs/COMPLIANCE/`.

**Cross-refs.** Rule 24. Rule 56.

---

### Rule 65 — DevSecOps quality threshold is regulator-grade

The DevSecOps mirror of Rule 25. The eleven non-negotiables in `docs/QUALITY-BAR.md` apply. Plus the regulated-industry-specific bar: SLSA L2 minimum (L3 plan in flight), Cosign keyless verify, pen-test annual + post-major, NIS2 timelines, GDPR DSAR endpoint, ARERA classification handled.

**Why.** Industrial customers' procurement processes already include SOC2-equivalent due diligence. A template that doesn't meet that bar can't be sold.

**How to apply.** `docs/QUALITY-BAR.md` §3 + §10. `docs/COMPLIANCE/`. Mission-II audit baseline.

**Cross-refs.** Rule 25. Rule 45. Mission-II audit.

---

### Rule 66 — DevSecOps rejection authority is exercised

The DevSecOps mirror of Rule 26. Rejected patterns enter `docs/adr/REJECTED.md` (REJ-24 through REJ-35). Code review enforces. Sec-ops sign-off required to override.

**Why.** Without rejection authority, every PR re-litigates the security posture.

**How to apply.** `docs/adr/REJECTED.md`. CODEOWNERS routes PRs touching `policies/`, `.github/workflows/`, `terraform/` to `@greenmetrics/secops`.

**Cross-refs.** Rule 26. Rule 46.

---

### Rule 67 — DevSecOps decision rationale travels with the decision

The DevSecOps mirror of Rule 47. Every ADR in the security/CI/CD path carries the four-part Tradeoff Stanza. Quarterly review.

**Why.** A security decision without rationale ages worst of all: nobody wants to revisit a security decision without knowing why.

**How to apply.** ADR template. Per-quarter office hours.

**Cross-refs.** Rule 27. Rule 47.

---

### Rule 68 — DevSecOps termination criterion is named

The DevSecOps mirror of Rule 28. The DevSecOps termination criterion is "the operator team can run the pipeline, respond to incidents, and run a tabletop without our help for 30 consecutive days." This is part of every engagement Phase-5 gate.

**Why.** A DevSecOps team that doesn't name a termination criterion becomes permanent overhead.

**How to apply.** Engagement Phase 5 gate. Quarterly DevSecOps review.

**Cross-refs.** Rule 28. Rule 48.

---

## Section 4 — Modular Template Integrity (Rules 69–88)

These rules are new in 2026-04-30 and exist because the charter (`docs/MODULAR-TEMPLATE-CHARTER.md`) reframes GreenMetrics as a modular template rather than a SaaS product. Modular template integrity is the property that the template can be specialised per engagement *without losing its identity* across engagements; every rule in this section is a constraint that protects that property.

### Rule 69 — Core and Pack are the load-bearing distinction

Every line of code, schema, fixture, and configuration belongs in exactly one of three places: Core (template-wide invariant, lives at the canonical paths in `internal/`, `frontend/src/`, `backend/migrations/`, `terraform/`, `gitops/`, `policies/`, `docs/`), a Pack (per-flavour swappable implementation, lives at `packs/<kind>/<id>/`), or an Engagement (single-client overlay, lives at `engagements/<client>/`). A line that fits in two places is a design smell to be resolved before merge.

**Why.** The whole template economy depends on being able to swap a Pack without touching Core, and to add an Engagement overlay without touching either. A line of Italian-specific factor-source code in `internal/services/` defeats the entire model.

**How to apply.** Conformance test `tests/conformance/core_pack_separation_test.go` (Phase E Sprint S5) walks the file tree and verifies each path matches its layer. Country names, regulator names, factor-source names appear in Core only as `Pack` interface references; concrete implementations live in Packs. CI gate fails the PR if a Core file imports a Pack-namespaced package, or vice versa beyond the documented seam.

**Cross-refs.** Charter §3. Rules 70, 71, 72.

---

### Rule 70 — A Pack manifests itself; a Pack does not announce itself

Every Pack ships a `manifest.yaml` declaring `id`, `kind`, `version`, `min_core_version`, `pack_contract_version`, `author`, `license_spdx`, `capabilities`, `dependencies`. Core's loader reads the manifest at boot, validates the schema, instantiates the Pack only when its declared `min_core_version` and `pack_contract_version` match. A Pack that lacks a valid manifest is rejected at boot.

**Why.** Implicit Pack discovery (e.g., scan a directory and import everything) creates a system whose composition is unknowable. Explicit manifests are the only way to prove "the system that ran on January 5 is the system that produced the audit dossier on March 30."

**How to apply.** `internal/packs/manifest.go` enforces the schema. `tests/packs/manifest_validation_test.go` (S5) is the conformance gate. The manifest schema is at `docs/contracts/pack-manifest.schema.json`.

**Cross-refs.** Rule 71. Rule 89 (provenance).

---

### Rule 71 — Pack contracts are versioned independently of Core

The Pack contract for each Pack-kind (`Ingestor`, `FactorSource`, `Builder`, `IdentityProvider`, `RegionProfile`) carries its own semantic version. Core declares the *supported window* of contract versions in `internal/packs/contracts.go`. A Pack that declares a contract version outside the supported window is rejected at boot. Deprecating a contract version emits an audit-log event on every Pack load.

**Why.** Core evolves; Packs evolve. A single global version number couples them. Independent versioning makes the upgrade surface tractable: a Core minor version may add a contract method (with a default no-op shim) and bump the contract version; existing Packs continue to satisfy the previous version.

**How to apply.** Per-kind contract versions in `internal/packs/contracts.go`. Contract-version deprecation is announced one minor release before removal. The conformance suite verifies contract-version-handling.

**Cross-refs.** Charter §12.2. Rule 72.

---

### Rule 72 — Pack registration is via Registrar, never via global

Packs register their contributions (Ingestors, Factor Sources, Builders, Identity Providers) through the typed `Registrar` indirection passed to `Pack.Register(reg Registrar) error`. Global state, init() side effects, and reflection-based discovery are forbidden.

**Why.** Globals defeat introspection. A Core that cannot answer "what's loaded right now?" cannot produce the SLSA-level provenance the regulator-grade bar requires. The `Registrar` indirection makes the loaded set explicit.

**How to apply.** `internal/packs/registry.go` defines `Registrar`. `Pack.Register` is the only legal registration path. Conformance test asserts no `init()` side effects in any `packs/` package.

**Cross-refs.** Rule 70. Rule 73.

---

### Rule 73 — Boot writes a manifest lock

At successful boot, Core writes `manifest.lock.json` capturing the loaded Pack set: per-Pack `id`, `version`, `pack_contract_version`, `manifest_hash`, `binary_hash`. The lock is signed (Cosign keyless) and shipped as a deployment artefact. Subsequent boots compare the loaded set to the lock; divergence triggers a Sev-2 alert.

**Why.** Audit reproducibility (Rule 89) starts with knowing what code ran. The manifest lock is the cryptographic answer.

**How to apply.** `internal/packs/lock.go`. `tests/packs/lock_test.go`. Cosign signature in CD pipeline.

**Cross-refs.** Rule 89. Rule 169.

---

### Rule 74 — Pack health is part of Core health

`Pack.Health(ctx)` is invoked on every `/api/health` and surfaced in the `dependencies` field of the health envelope (Rule 6). A Pack returning `unhealthy` for more than its configured grace period flips the envelope `status` to `degraded` (or `unhealthy` per Pack policy).

**Why.** A Pack that's broken silently is the worst kind of broken. Surfacing Pack health in the platform health envelope is the cheapest possible early warning.

**How to apply.** `internal/handlers/health.go` aggregates `Pack.Health`. `tests/conformance/pack_health_test.go`.

**Cross-refs.** Rule 6. Rule 23.

---

### Rule 75 — Pack capabilities are declared, not discovered

The `capabilities` field in `manifest.yaml` lists the explicit features a Pack provides (e.g., `protocol.modbus.tcp`, `factor.source.ispra`, `report.builder.esrs_e1`). Core matches required capabilities (declared in `config/required-packs.yaml`) against loaded capabilities; missing capability is a boot refusal.

**Why.** Capability matching makes engagement-time configuration explicit. A deployment that requires the Italian Region Pack and a SunSpec Protocol Pack but loads only the Italian Pack fails fast at boot with a comprehensible error, not at first ingestion with a `nil pointer dereference`.

**How to apply.** `config/required-packs.yaml`. Loader matches before instantiation. Conformance: `tests/packs/capability_matching_test.go`.

**Cross-refs.** Rule 70. Charter §3.2.

---

### Rule 76 — Pack failures are isolated; they do not crash Core

A Pack whose `Init()` returns error fails the Pack's own boot but does not crash Core when the Pack is non-required. A Pack whose request handler panics is recovered by Core's panic-recovery middleware; the failure is logged with `pack.id` context. A Pack that violates resource limits (memory, goroutines, file descriptors) is restartable without Core restart in the worker-pool runtime (Phase F Sprint S10 — pluggable worker isolation).

**Why.** Core stability cannot be the lowest-common-denominator of every loaded Pack. If a Pack misbehaves, the platform must continue to serve the rest.

**How to apply.** Per-Pack error wrapping. Per-Pack panic recovery. Per-Pack resource budgets in the worker pool.

**Cross-refs.** Rule 41. Rule 42.

---

### Rule 77 — Engagement code lives in `engagements/<client>/`

Code, schema, fixtures, and configuration that is uniquely useful to one client lives in `engagements/<client>/` and never in upstream `main`. The fork's `template-version.txt` records the upstream version it last synced from. Sync runbook `docs/runbooks/upstream-sync.md` is the canonical workflow.

**Why.** Engagement code in upstream pollutes the template. The next engagement inherits assumptions that don't apply. The template ages badly.

**How to apply.** `engagements/` is in upstream `.gitignore`-equivalent (or a separate branch). Conformance test: upstream `main` has no `engagements/` directory.

**Cross-refs.** Charter §5.2. Rule 79.

---

### Rule 78 — Core changes are merge-friendly between minor versions

A Core change between minor versions does not break Pack contracts, does not change the schema in a non-additive way, does not remove an OpenAPI operation. Breaking changes batch into major-version releases, cadence ≤ once per 18 months. The `merge-friendliness` test at `tests/contracts/v1_compat_test.go` verifies the OpenAPI invariant; the `additive-only` test at `tests/migrations/additive_only_test.go` verifies the schema invariant.

**Why.** Engagement forks fall behind the day they're created. The longer they stay behind, the more painful the sync. Merge-friendliness is the engineering property that makes quarterly upstream-sync tractable.

**How to apply.** CI gates on the merge-friendliness tests. PR template prompts for breaking-change check.

**Cross-refs.** Charter §7.2. Rule 21. Rule 100.

---

### Rule 79 — Engagement forks sync upstream quarterly minimum

Every engagement fork merges the upstream `release/v1.x` branch at least once per quarter, or before each new client-visible release of the engagement deployment, whichever is sooner. A fork that has not synced for two consecutive quarters is flagged in the engagement health register; the engagement lead has 30 days to either sync or document the exception.

**Why.** The longer a fork lags upstream, the harder the next sync becomes. Quarterly cadence keeps the merge surface small.

**How to apply.** `docs/runbooks/upstream-sync.md`. Engagement health dashboard in `monitoring/engagement-dashboard.json` (Phase E Sprint S6).

**Cross-refs.** Charter §7.1. Rule 78.

---

### Rule 80 — Core customisations in a fork are time-bounded

If an engagement fork must customise Core (vs. just adding Packs or engagement overlays), the customisation is documented in `engagements/<client>/CORE-CUSTOMISATIONS.md` with a Tradeoff Stanza and a sunset date no more than two template releases out. By the sunset, the customisation is either upstreamed (becomes Core) or reverted to a Pack-mediated solution.

**Why.** Core customisations are technical debt that grows with every upstream sync. Time-bounding them forces resolution.

**How to apply.** PR template requires the file when modifying Core in an engagement fork. Quarterly engagement review.

**Cross-refs.** Charter §5.2. Rule 79.

---

### Rule 81 — Branding is configuration, not code

The `config/branding.yaml` file controls every branding surface (`product_name`, `legal_entity`, `support_contact`, `logo_*`, `theme_*`, `footer_text`, `pdf_cover_template`). The conformance suite (`tests/conformance/no_hardcoded_brand_test.go`) verifies no rendered surface contains hard-coded brand strings outside `config/branding.yaml` defaults.

**Why.** Hard-coded brand strings spread through templates, PDFs, email subjects, error messages. Each one is a per-engagement edit. Configuration centralises the cost.

**How to apply.** `config/branding.yaml` schema. PDF templates accept variables. Frontend `app.css` reads theme tokens at build. Conformance test.

**Cross-refs.** Charter §6.

---

### Rule 82 — Configuration is layered: Core defaults → Pack defaults → Engagement overrides

The configuration loader resolves keys via a fixed precedence: engagement override (`engagements/<client>/config/`) wins over Pack default (`packs/<kind>/<id>/config/defaults.yaml`) wins over Core default (`internal/config/defaults.go`). The resolved set is logged at boot for auditability.

**Why.** Configuration without precedence rules creates "why is this value what it is?" mysteries. Explicit layering makes the answer derivable in 5 seconds.

**How to apply.** `internal/config/loader.go`. Conformance: `tests/config/precedence_test.go`.

**Cross-refs.** Rule 73. Rule 81.

---

### Rule 83 — Pack contracts forbid recursive Pack discovery

A Pack does not enumerate other Packs at runtime. A Pack does not register its own children. Pack composition is a Core concern, expressed in `config/required-packs.yaml`. A Pack that needs another Pack declares it in the manifest's `dependencies` field.

**Why.** Recursive Pack discovery is the path to circular load-time dependencies, undefined initialisation order, and dependency hell. Centralised composition is tractable.

**How to apply.** `internal/packs/loader.go` is the only place that enumerates Packs. Conformance: `tests/packs/no_recursive_discovery_test.go`.

**Cross-refs.** Rule 70. Rule 75.

---

### Rule 84 — Pack tests run independently and against Core

Each Pack carries its own test directory (`packs/<kind>/<id>/tests/`) with unit tests, contract conformance tests, and integration tests against a containerised Core. A Pack PR is rejected if its tests don't pass against the *minimum-supported Core version*.

**Why.** Pack-only testing leaves Core-Pack interaction bugs undiscovered. Cross-version testing surfaces compatibility regressions.

**How to apply.** Per-Pack `Makefile` target. CI matrix: Pack × Core-min × Core-current.

**Cross-refs.** Rule 71. Rule 84.

---

### Rule 85 — Pack-loader instrumentation is deep

Loader emits structured logs per Pack: load start, manifest validation result, init duration, registration counts, health-check first result, shutdown duration. Loader emits Prometheus counters: `packs_loaded_total{pack_id, kind, version}`, `pack_load_errors_total`, `pack_health_status{pack_id, status}`. Loader emits OTel spans wrapping each phase.

**Why.** A loader that's a black box is unfixable. Deep instrumentation is the operations-team's fastest debug surface.

**How to apply.** `internal/packs/loader.go` emits the signals. Grafana dashboard `pack-loader-dashboard.json`.

**Cross-refs.** Rule 40. Rule 76.

---

### Rule 86 — Pack contracts are documented in code, not just text

Each Pack-kind contract has a Go doc comment explaining the contract, a code example of a minimal compliant implementation, and a link to the conformance test. A Pack-contract change requires the doc + example + test to update in the same PR.

**Why.** Documentation that lives separately from the contract drifts. Co-located documentation rots less because the next reader/writer is right there.

**How to apply.** Custom golangci-lint rule: package `internal/domain/<kind>` requires a doc-example test. PR template prompts.

**Cross-refs.** Rule 14. Rule 47.

---

### Rule 87 — Pack acceptance criteria are objective

A new Pack is accepted into the upstream template only if: it passes its own tests, it passes the conformance suite, it has a Tradeoff Stanza ADR, it has a Pack-charter (`packs/<kind>/<id>/CHARTER.md`), it does not import any other Pack outside its declared dependencies, and at least one Macena platform-team reviewer has built it from clean clone in under 30 minutes.

**Why.** Subjective acceptance breeds inconsistency. Objective acceptance is reproducible across reviewers.

**How to apply.** Pack-acceptance checklist in `docs/PACK-ACCEPTANCE.md` (Phase E Sprint S5). PR template prompts.

**Cross-refs.** Rule 70. Rule 87.

---

### Rule 88 — The Italian Region Pack is the flagship reference Pack

The `packs/region/it/` Pack is the canonical reference: it is the most complete, most tested, most documented Pack in the upstream. New Packs are reviewed against the Italian flagship for thoroughness. The Italian Pack's CHARTER, manifest, README, conformance tests, and Tradeoff Stanza are the template for new Region Packs.

**Why.** Without a reference, "thorough" is undefined. With a reference, "thorough" is "as thorough as Italy."

**How to apply.** Phase E Sprint S5–S7 extracts current Italian-specific code into `packs/region/it/` and `packs/factor/{ispra,terna,gse,aib}/` and `packs/report/{esrs_e1,piano_5_0,conto_termico,tee,audit_dlgs102}/`. Subsequent Region Packs (`packs/region/{de,fr,es,gb}/`) lift the structure.

**Cross-refs.** Charter §3.2. PLAN.md Phase E.

---

## Section 5 — Data Provenance & Audit-Grade Reproducibility (Rules 89–108)

These rules govern the property that every regulatory report can be re-derived bit-for-bit from raw readings + factor versions + Pack versions, indefinitely. This is the single largest moat against every competitor surveyed in `docs/COMPETITIVE-BRIEF.md`: none of them guarantee bit-perfect re-derivation.

### Rule 89 — Every regulatory output is bit-perfect reproducible from source

Given the same `(tenant_id, period, report_type, manifest_lock_hash, factor_pack_version, report_pack_version)`, the regulatory output (ESRS E1 dossier JSON, Piano 5.0 attestazione, Conto Termico GSE submission, TEE submission, audit 102/2014 dossier) is byte-for-byte identical. Determinism is enforced in code, validated by the conformance suite (`tests/conformance/report_determinism_test.go`), and is the central audit guarantee.

**Why.** A regulator that asks "show me how you computed this number" expects to be able to re-run the computation and get the same answer. A pipeline that produces different outputs on different runs is not auditable.

**How to apply.** Builders are pure functions of `(period, factors, readings)`. No `time.Now()` inside builders (except as parameter). No map iteration in serialisation paths. Sorted keys in JSON output. Stable round-trip across Go minor versions (verified by Go-version-matrix in CI).

**Cross-refs.** Rule 73. Rule 90. Rule 91.

---

### Rule 90 — Factor sources are versioned with temporal validity

Every emission factor is keyed `(code, valid_from)`. `valid_to NULL` means active. A factor query for a temporal point selects the factor whose `valid_from <= ts AND (valid_to IS NULL OR valid_to > ts)` and is the most recent. A regulatory rerun for a past period uses the factor that was valid at that period — even if a newer factor is in the table.

**Why.** When ISPRA publishes the April update reducing the 2024 mix from 0.250 to 0.245, a 2024 report rerun in 2026 must continue to use 0.250 if it was the official figure at the original reporting time. *That* is the property auditors care about.

**How to apply.** `emission_factors` schema (migration 0005). Repository method `factor.Active(ctx, code, ts)`. Conformance: `tests/conformance/factor_temporal_test.go`. Architecture doc §3.

**Cross-refs.** Rule 89. Rule 130. Schema-evolution policy.

---

### Rule 91 — Builders are pure functions

Every Report Pack `Builder.Build(ctx, period, factors, readings)` returns `(Report, error)` and depends only on the inputs. No global state. No I/O outside the `ctx`-bound resources. No `time.Now()`. No environment variables read inside the builder. The conformance test runs every builder twice with the same inputs and asserts byte-identical output.

**Why.** Purity is the only guarantee that re-derivation is possible. A builder that depends on global state is a builder whose output cannot be reproduced.

**How to apply.** Custom golangci-lint rule: `internal/domain/reporting/<builder>/` is forbidden to import `time`, `os`, `crypto/rand`, `math/rand`, or any package from `internal/services/`. Conformance test re-runs.

**Cross-refs.** Rule 89. Rule 92.

---

### Rule 92 — Aggregation is associative, commutative, and verifiable

Every aggregation function used in builders (`Sum`, `Average`, `WeightedAverage`, `Max`, `Min`, `P95`) carries property-based tests asserting associativity (where applicable), commutativity (where applicable), and identity (zero-valued aggregation behaviour). The property tests live in `backend/tests/property/aggregate_invariants_test.go` and run on every CI.

**Why.** A non-associative aggregation gives different answers depending on input order. A sustainability report that depends on the order of meter readings is not a sustainability report; it's a coin flip.

**How to apply.** `backend/tests/property/aggregate_invariants_test.go` exists and runs. New aggregation functions add tests in the same file.

**Cross-refs.** Rule 89. Rule 91.

---

### Rule 93 — Time bucketing is deterministic and named

Every time-bucket boundary is computed deterministically: 15-min buckets at `:00, :15, :30, :45`; 1-h buckets at `:00`; 1-day buckets in `Europe/Rome` *or* UTC depending on the report (declared per Builder); 1-month buckets at the calendar-month boundary in the configured timezone. Bucket boundaries are not derived from "now" relative arithmetic; they are absolute.

**Why.** A bucket boundary that drifts (e.g., "the past 24 hours from now") produces a different report on every run. Absolute boundaries are reproducible.

**How to apply.** `internal/domain/timebucket/` defines the bucketing functions. Conformance: `tests/conformance/timebucket_test.go`.

**Cross-refs.** Rule 89. Rule 91.

---

### Rule 94 — Floats forbidden in regulatory paths

The regulatory-output path (Builders, Carbon calculator, Aggregation) uses fixed-point integer arithmetic for emissions in micrograms and energy in milliwatt-hours. Floats are forbidden in any code that contributes to a Report. A float in this path triggers a custom-lint rule.

**Why.** Float arithmetic introduces non-deterministic rounding. A regulatory output computed with floats is a regulatory output that may differ between runs and between platforms.

**How to apply.** Custom golangci-lint rule scopes a deny-list of `float64` / `float32` to specific packages. Conformance: `tests/static/no_float_in_regulatory_test.go`.

**Cross-refs.** Rule 1. Rule 89. Rule 91.

---

### Rule 95 — Every report carries a provenance bundle

Every persisted report row carries a `provenance` JSONB column with: `manifest_lock_hash`, `factor_pack_versions`, `report_pack_version`, `query_definitions`, `source_data_window`, `tenant_data_region`, `executor_user_id`, `executed_at_utc`. The provenance bundle is signed (Cosign sign-blob) and the signature is stored in the same row.

**Why.** A report without provenance is a report whose claims cannot be independently verified. Provenance is the engineering equivalent of a notarised affidavit.

**How to apply.** `reports` schema carries `provenance jsonb` and `provenance_signature bytea`. The Builder finalises the provenance and signs before persistence.

**Cross-refs.** Rule 73. Rule 169.

---

### Rule 96 — Source data is immutable in the audit window

Raw readings within the audit retention window (10 years for daily aggregates) are immutable. Corrections happen via a `corrections` overlay table, never via in-place updates. The Builder reads `readings JOIN corrections` to produce the final figures. Original readings remain visible for forensic reconstruction.

**Why.** A raw-reading column that can be updated is a raw-reading column that an auditor cannot trust. Immutability + append-only corrections preserve the original record while permitting correction.

**How to apply.** Migration adds `corrections` table (Phase F Sprint S10). Builder query joins. Conformance: `tests/conformance/raw_reading_immutability_test.go`.

**Cross-refs.** Rule 33. Rule 62.

---

### Rule 97 — Algorithm changes are versioned and ADR'd

Every change to a regulatory algorithm (carbon calculator coefficient, Piano 5.0 threshold, ESRS E1 disclosure aggregation) bumps a `report_pack_version` and ships an ADR explaining the change, the previous behaviour, and the migration story for previously-generated reports.

**Why.** Regulatory algorithms evolve. A change to the algorithm without versioning produces inconsistent reports across cohorts. An ADR documents the change so the next auditor can reconstruct the timeline.

**How to apply.** `report_pack_version` in manifest. ADR in `docs/adr/`. Conformance: every Builder bump links an ADR.

**Cross-refs.** Rule 47. Rule 71. Charter §12.2.

---

### Rule 98 — Replay test runs every Builder against pinned fixtures

The conformance suite includes `tests/conformance/builder_replay_test.go` that runs every Builder against versioned fixture data in `tests/fixtures/regulatory/<period>/<scenario>.golden.json` and asserts byte-identical output. Updates to the golden files require a Tradeoff Stanza ADR.

**Why.** Replay tests are the cheapest possible regression detection for the highest-stakes code in the codebase. A change that breaks a Builder produces a clear failure, not a subtle drift.

**How to apply.** Fixtures in `tests/fixtures/regulatory/`. Conformance test. Golden-file update is its own PR.

**Cross-refs.** Rule 89. Rule 91.

---

### Rule 99 — Data lineage is queryable

Every regulatory output exposes a `lineage` endpoint (`/api/v1/reports/{id}/lineage`) returning: input meter set, input channel set, input reading count, factor versions used, builder version, manifest lock, signing certificate. The endpoint is auditor-visible (RBAC role `auditor`) and is part of the standard CSRD evidence pack export.

**Why.** Lineage that has to be reconstructed manually is lineage that's reconstructed wrong. Queryable lineage is the auditor's primary verification tool.

**How to apply.** `internal/handlers/reports.go` adds the endpoint. Conformance: `tests/contracts/lineage_endpoint_test.go`.

**Cross-refs.** Rule 95. Rule 145.

---

### Rule 100 — Schema changes are additive on regulatory paths

Schema columns and indexes that contribute to a regulatory output (`readings`, `meters`, `meter_channels`, `emission_factors`, `reports`, `corrections`, `audit_log`) are additive only between major Core versions. No `ALTER COLUMN TYPE`, no `DROP COLUMN`, no `ALTER TABLE` that rewrites the table. New columns are nullable or default-valued.

**Why.** A regulatory schema that can change shape between Core versions is a schema that breaks audit reproducibility.

**How to apply.** `docs/SCHEMA-EVOLUTION.md` §3. CI rule: edit to applied migration is rejected. Conformance: `tests/migrations/additive_only_test.go`.

**Cross-refs.** Rule 33. Rule 78. Rule 89.

---

### Rule 101 — Time-zone configuration is per-tenant explicit

Every tenant carries an explicit `timezone` (default `Europe/Rome` from the Italian Region Pack; configurable per Region Pack). Every report names the timezone in its provenance. Per-tenant timezone is the only legal way to compute "1 day" in human-meaningful terms.

**Why.** A timezone defaulted from the server breaks the moment the server moves regions. A timezone defaulted from the user breaks the moment the user travels. Per-tenant timezone is the only stable answer.

**How to apply.** `tenants.timezone` column. Region Pack supplies the default. Conformance: `tests/conformance/tenant_timezone_test.go`.

**Cross-refs.** Rule 2. Rule 93.

---

### Rule 102 — Continuous aggregates are not edited in place

CAGGs are dropped and recreated, never `CREATE OR REPLACE`-altered. CAGG schema changes happen during a maintenance window, with the historical-window refresh logged in the audit log. The maintenance window is announced ≥ 14 days in advance.

**Why.** In-place CAGG alteration silently changes the answer auditors get when they query historical data. Drop-and-recreate is loud and traceable.

**How to apply.** `docs/SCHEMA-EVOLUTION.md` §5. Goose `NO TRANSACTION` for CAGGs. Audit-log entry on every CAGG operation.

**Cross-refs.** Rule 33. Rule 100.

---

### Rule 103 — Retention shortenings require ADR + 30-day notice

Reducing a retention period (e.g., raw 90 d → 60 d, or 1d aggregate 10 y → 7 y) requires an ADR justifying the data loss, a 30-day customer notice, and a per-tenant opt-out path for clients with regulatory retention obligations greater than the new default.

**Why.** Retention is a contractual and regulatory commitment. Quietly cutting retention deletes data that auditors will need.

**How to apply.** `docs/SCHEMA-EVOLUTION.md` §6. ADR template. Customer-notice runbook.

**Cross-refs.** Rule 33. `docs/SLO.md` §2.

---

### Rule 104 — Soft-deletes never hard-delete in the audit window

Tenants and users are soft-deleted (`active = false`) and never hard-deleted within the audit retention window (10 years). The right-of-erasure (GDPR Art. 17) for natural persons is satisfied by *crypto-shredding*: the per-tenant DEK is deleted, rendering the encrypted PII fields irrecoverable, while the structural audit row remains.

**Why.** GDPR right-of-erasure conflicts with audit retention. Crypto-shredding resolves the conflict: the row exists for audit, but the personal data inside it is unreadable.

**How to apply.** `internal/domain/tenant/erasure.go`. KMS DEK lifecycle. Conformance: `tests/conformance/crypto_shredding_test.go`.

**Cross-refs.** Rule 169. Rule 184. `docs/COMPLIANCE/GDPR.md`.

---

### Rule 105 — Reading provenance includes meter signature

Every reading carries `meter_id`, `channel_id`, `quality_code`, `raw_payload` (encrypted), and the meter's HMAC signature over `(meter_id, channel_id, ts, value)`. The signature is verified at ingestion (Phase F Sprint S11). Unsigned readings are flagged with a `quality_code = 'unsigned'` and excluded from regulatory builders that require signed input.

**Why.** Without a signature, an attacker who reaches the wire between meter and backend can manipulate readings. Cryptographic signatures break that trust assumption.

**How to apply.** Edge gateway holds the per-meter signing key. `internal/services/ingest_pipeline.go` verifies. Conformance: `tests/conformance/signed_reading_test.go`.

**Cross-refs.** Rule 169. Rule 173. PLAN.md Phase F.

---

### Rule 106 — Submission to GSE/ENEA portals is idempotent and signed

Outbound submissions to GSE (Conto Termico, TEE) and ENEA (audit 102/2014) carry an idempotency key (the `report_id`); on the GSE/ENEA side, the same `report_id` is rejected as duplicate. The submission payload is signed (Cosign sign-blob), the signature accompanies the submission, and the GSE/ENEA `portal_ref` is stored on success.

**Why.** Industrial dossiers run €30k–€100k each; a duplicate submission is a regulatory mess. Idempotency + signature defends against accidental re-submission and against tampering en route.

**How to apply.** `internal/services/portal_submission.go`. ADR on the submission flow. Conformance: `tests/conformance/portal_idempotent_test.go`.

**Cross-refs.** Rule 35. Rule 95.

---

### Rule 107 — Restore-from-backup is rehearsed quarterly

The DR drill in `docs/CHAOS-PLAN.md` includes a quarterly restore-from-backup run on staging: take last week's backup, restore to a fresh Timescale instance, run the conformance + property + replay tests, run a synthetic regulatory generation, compare against the production-saved version. A drill failure is a Sev-1 platform incident.

**Why.** Backups that are never restored are not backups. Restore drills are the only way to verify the pipeline is intact.

**How to apply.** `docs/runbooks/dr-restore.md`. `task dr:restore` Taskfile target. `docs/CHAOS-LOG.md` records each drill.

**Cross-refs.** Rule 64. `docs/SLO.md` §3.

---

### Rule 108 — Audit-evidence pack is single-command exportable

`task audit-evidence-pack` (Phase G Sprint S15) produces a zip containing: every audit-log row in the period, every report with its provenance and signature, every Pack manifest with its lock hash, the OpenAPI spec at the version-of-record, the Cosign signature of every deployed image, the SBOM of every deployed image, the conformance-suite green status, the running-system manifest. The zip is signed.

**Why.** A regulator's ask is "show me everything." An ad-hoc evidence pull at audit time is the path that produces the wrong answer. Single-command export means the answer is always already-shaped.

**How to apply.** `task audit-evidence-pack`. Conformance: pack contents validated against schema. Time budget: ≤ 15 minutes for a 1-year audit window.

**Cross-refs.** Rule 63. Rule 95. Rule 145.

---

## Section 6 — OT Integration Discipline (Rules 109–128)

These rules govern the OT (operational technology) ingestion edge — Modbus, M-Bus, SunSpec, OCPP, Pulse, IEC 61850, OPC UA, MQTT Sparkplug B, BACnet, EtherCAT, PROFINET, IEC 62056-21 IR optical, KNX. The OT edge is the most heterogeneous and most legacy-bound part of the stack; the rules exist because OT bugs are not visible until the customer site fires alarms.

### Rule 109 — Every OT protocol is a Pack

Modbus TCP, Modbus RTU, M-Bus, SunSpec, Pulse, OCPP, IEC 61850, OPC UA, MQTT Sparkplug B, BACnet, EtherCAT, PROFINET, IEC 62056-21, KNX — each is a Pack at `packs/protocol/<id>/`. Core's ingestion runtime (`internal/services/ingestor_runner.go`) speaks only the `Ingestor` interface. New protocols are added as Packs without touching Core.

**Why.** Hard-coding protocols into Core makes adding a new protocol a Core release. Pack-based protocols make new protocols an engagement-team activity.

**How to apply.** Charter §3.2. EP-01 in `docs/EXTENSION-POINTS.md`. Conformance: `tests/packs/protocol_count_test.go`.

**Cross-refs.** Rule 69. ADR (Phase E Sprint S5 records the Pack extraction).

---

### Rule 110 — Protocol Packs declare their wire-format invariants

Each Protocol Pack's CHARTER documents the wire-format invariants it depends on: byte order, register layout, scaling factors, unit-of-measure conventions, slave-ID ranges, function codes used, error-frame handling. The CHARTER is the contract between the Pack and the integrator who wires the meter.

**Why.** Wire-format assumptions that are not documented become tribal knowledge. When the integrator changes, the assumption is lost.

**How to apply.** `packs/protocol/<id>/CHARTER.md` template includes a "Wire format invariants" section. PR template prompts.

**Cross-refs.** Rule 70. Rule 87.

---

### Rule 111 — Edge buffering is the integrator's first-line defence

Every Protocol Pack documents the recommended edge-buffer size (in seconds of readings) for the protocol. The default reference edge gateway (`backend/cmd/simulator` modelled in v1; replaced by a real edge in Phase G Sprint S14) carries a 24-hour disk-backed buffer. Backend unavailability for ≤ 24h does not produce data loss.

**Why.** Industrial environments have intermittent connectivity. A 24-hour edge buffer is the difference between "no data loss when the WAN flaps" and "every WAN flap is a data-loss incident."

**How to apply.** Edge-gateway design at `docs/EDGE-GATEWAY.md` (Phase G Sprint S14). Per-Pack buffer recommendation in CHARTER.

**Cross-refs.** Rule 36. Rule 124.

---

### Rule 112 — Time-source on the edge is NTP'd, GPS'd if available

The edge gateway's clock is NTP-synced to a stratum ≤ 2 source. Where the meter physically supports it (PV inverter, energy management gateway), GPS time is preferred. The reading's `ts` is the *meter's* timestamp where the meter supplies one (Modbus servers with timestamp registers, M-Bus with timestamp objects, SunSpec inverters with their own clock); else the edge gateway's NTP-synced time at sample-pull.

**Why.** A sustainability report whose timestamps drift by hours produces wrong day-night curves. NTP discipline is the cheapest possible defence.

**How to apply.** Edge-gateway runbook documents NTP config. Conformance: incoming readings with `ts` more than 5 minutes off backend's `now` are flagged in the audit log.

**Cross-refs.** Rule 2. Rule 124.

---

### Rule 113 — Reading quality codes are explicit and standardised

Every reading carries a `quality_code` enum: `good`, `unsigned`, `interpolated`, `manually_overridden`, `out_of_range`, `device_error`, `timestamp_anomaly`, `sequence_gap`. The Builders treat each code per documented policy: good is included; interpolated is included with a flag in the report's `notes`; manually_overridden is included with an audit-log reference; out_of_range / device_error / timestamp_anomaly / sequence_gap are excluded.

**Why.** Without quality codes, a reading whose value is `0` could mean "the meter measured zero" or "the meter is broken." Reports built on the latter are wrong.

**How to apply.** `internal/domain/readings/quality.go` enumerates codes. Each Pack assigns them. Builders branch on them.

**Cross-refs.** Rule 91. Rule 105.

---

### Rule 114 — Modbus polling cadence respects the meter's manual

Modbus servers have device-specific minimum intervals between requests. Polling faster than the documented minimum corrupts readings on ageing devices. Each Modbus device profile in `packs/protocol/modbus_tcp/devices/<vendor>-<model>.yaml` declares the documented polling minimum; the runtime enforces it with a per-device rate limiter.

**Why.** A Modbus EM24 polled every 1s is a Modbus EM24 that drops every other reading. The vendor manual is non-negotiable.

**How to apply.** Per-device YAML profile. `internal/services/modbus_ingestor.go` reads it.

**Cross-refs.** Rule 109. Rule 122.

---

### Rule 115 — M-Bus addressing is documented per-installation

The M-Bus primary/secondary address scheme varies between installations. The engagement's `engagements/<client>/m-bus-addressing.yaml` documents the address layout for every connected meter. A reading from an undocumented address is rejected with `device_error`.

**Why.** M-Bus address collisions silently mix meter data. Documented addressing prevents accidental cross-tenant reading-mixing in shared installations.

**How to apply.** Engagement-fork-only file. `internal/services/mbus_ingestor.go` reads it.

**Cross-refs.** Rule 77. Rule 109.

---

### Rule 116 — SunSpec models declared per device

PV inverter SunSpec model coverage varies by manufacturer. `packs/protocol/sunspec/devices/<vendor>-<model>.yaml` declares the supported models and their version. A SunSpec response that doesn't match the declared models is logged and rejected.

**Why.** A new firmware revision that adds a model can silently change the indexing. Explicit model declaration catches the change.

**How to apply.** Per-device YAML. `internal/services/sunspec_profile.go` enforces.

**Cross-refs.** Rule 109. Rule 110.

---

### Rule 117 — OCPP version pinning is per-charger

OCPP 1.6 and 2.0.1 are the two supported wire protocols. Per-charger configuration declares the version and the JSON dialect (Plain / Compact). Mixed-version chargers are deployed only after engagement-team explicit approval and a charger-specific compatibility test.

**Why.** OCPP version mismatches produce undefined behaviour: a 1.6 message fed into a 2.0.1 parser parses to the wrong fields. Explicit pinning eliminates the class.

**How to apply.** `engagements/<client>/ocpp-chargers.yaml`. `internal/services/ocpp_client.go` reads.

**Cross-refs.** Rule 109. Rule 110.

---

### Rule 118 — Pulse webhook is HMAC-SHA256 with constant-time compare

Inbound pulse-meter webhooks carry an HMAC-SHA256 signature in `X-Pulse-Signature`. The verifier uses `subtle.ConstantTimeCompare` (REJ-20). Plaintext compare is rejected at code review. Replay protection: a `nonce` in the body, recently-seen-nonce cache for 5 minutes.

**Why.** Plaintext compare is a timing oracle. Replay without nonce protection allows an attacker to re-fire a captured webhook.

**How to apply.** `internal/services/pulse_ingestor.go`. Conformance: `tests/security/pulse_constant_time_test.go`.

**Cross-refs.** REJ-20. Rule 178. Rule 39.

---

### Rule 119 — Ingest path is bounded; backpressure is observable

The ingest path is bounded in two places: a bounded channel in `internal/services/ingest_pipeline.go` (ADR-0015) and a worker pool with maximum concurrent batch writes. Drop policy is "drop oldest, increment counter, log warning." `ingest_drops_total{reason}` Prometheus counter and `ingest_lag_seconds` histogram are scraped.

**Why.** Unbounded ingest under load OOMs. Bounded ingest with backpressure surfaces the symptom early enough to scale.

**How to apply.** ADR-0015. AB-08. Conformance: `tests/load/ingest_backpressure_test.js` (k6).

**Cross-refs.** Rule 41. Rule 38. REJ-23.

---

### Rule 120 — Unit-of-measure conversions are explicit and tested

Every unit conversion (kWh → MJ, m³ → kWh-thermal at 9.6 kWh/m³ for natural gas, BTU/h → kW, etc.) is implemented in `internal/domain/units/` with a property-based test asserting reversibility. Conversion in handlers, services, or builders is forbidden — they call the units package.

**Why.** Hand-rolled unit conversions in builders are the single most common source of arithmetic errors in regulatory reports. Centralising the conversion is the only defence.

**How to apply.** `internal/domain/units/`. Conformance: `tests/property/units_test.go`. Custom golangci-lint rule scopes the deny-list.

**Cross-refs.** Rule 91. Rule 94.

---

### Rule 121 — Per-protocol simulators ship for development

Every Protocol Pack ships a development-time simulator that speaks the protocol against a known-fixture script. The simulator is reproducible, deterministic, and runs in `docker compose up`. New Packs add a simulator before they are accepted (Rule 87).

**Why.** Protocol bugs that surface only at the customer site are the most expensive bugs in the ecosystem. A simulator is the cheapest possible local-development feedback loop.

**How to apply.** `packs/protocol/<id>/simulator/`. Per-Pack docker-compose service. Conformance: `tests/integration/protocol_simulator_test.go`.

**Cross-refs.** Rule 87. Rule 109.

---

### Rule 122 — Device profiles capture the meter's documentation

Each meter's vendor manual is summarised into a YAML profile in `packs/protocol/<id>/devices/<vendor>-<model>.yaml`: register layout, scaling, polling minimum, error frames, supported function codes. Profile authorship is part of the engagement Phase 2 (Pack assembly) work.

**Why.** Centralising vendor documentation in version-controlled profiles is the alternative to "ask the integrator who wired it last time."

**How to apply.** Per-device YAML schema in `docs/contracts/device-profile.schema.json`. CI validates schema compliance.

**Cross-refs.** Rule 110. Rule 114.

---

### Rule 123 — Channel mapping is auditor-visible

Every meter's `meter_channels` rows describe the physical mapping (Phase R, Phase S, Phase T, Total, Reactive, Frequency, Voltage, Current, etc.) and the unit. The mapping is queryable via `/api/v1/meters/{id}/channels` (RBAC role auditor reachable). A regulatory output that aggregates "Active Energy" across phases must use the channel taxonomy, not name-matching.

**Why.** Channel taxonomy drift (sometimes "Total Active" is the sum of phases, sometimes a separate register) is a frequent source of double-counting.

**How to apply.** `meter_channels.kind` enum. Schema-evolution-policy-compliant additions.

**Cross-refs.** Rule 99. Rule 89.

---

### Rule 124 — Latency budget for ingest is documented per-protocol

Modbus polling: round-trip ≤ 200ms; M-Bus: ≤ 500ms; SunSpec: ≤ 300ms; OCPP: ≤ 1s for transactions; Pulse: ≤ 100ms webhook handling. Latency exceeding the budget triggers a warning alert; sustained exceeding triggers a Sev-3.

**Why.** Latency degradation at the OT edge is an early indicator of meter or network failure. Catching the trend early prevents data loss.

**How to apply.** Per-Pack histogram metric. Per-protocol alert rules.

**Cross-refs.** Rule 37. Rule 119.

---

### Rule 125 — Outbound DSO clients (E-Distribuzione, Terna, SPD) are circuit-breakered

The Italian Region Pack's outbound integrations to E-Distribuzione SMD, Terna Transparency, SPD multi-DSO carry per-host circuit breakers. Cooldown 60s; failure threshold 5 in 60s; max-requests-half-open 1. Graceful degradation: cached fallback with `data_freshness` stamp on the consumer surface.

**Why.** Italian DSO portals have sustained outages every few months. A backend that hard-fails when the portal is down is a backend that hard-fails several times a year.

**How to apply.** ADR-0009. `internal/resilience/breaker.go`. Per-client breaker init.

**Cross-refs.** Rule 36. Rule 132.

---

### Rule 126 — Network segmentation is OT-aware

Topology D (hybrid) deploys the ingest path inside the OT segment of the customer's network. The OT-segment ingest backend speaks only to the OT-segment meters and to a single egress proxy on the IT-segment boundary. Cross-segment traffic is whitelisted, logged, and capped at the egress proxy.

**Why.** OT/IT segmentation is a regulatory and insurance requirement for energy-intensive operations (NIS2 D.Lgs. 138/2024). A backend that violates the segmentation is a backend that voids the customer's insurance.

**How to apply.** Topology-D documentation. NetworkPolicy bundles. Conformance: integration test asserts network egress policy.

**Cross-refs.** Charter §10.4. Rule 8.

---

### Rule 127 — Anomaly detection is layered

Layer 1 (z-score on rolling 7-day baseline per meter): `internal/services/alert_engine.go`. Layer 2 (seasonal-decomposition via STL): Phase I (AI/ML) Sprint S18. Layer 3 (cross-meter correlation): Phase I Sprint S19. Each layer surfaces alerts with `detection_layer` annotation.

**Why.** Single-layer anomaly detection produces too many false positives or too few true positives. Layered detection lets each layer specialise.

**How to apply.** `internal/domain/alerting/detector.go` Pack interface. Per-layer Detector. Operator dashboard segregates by `detection_layer`.

**Cross-refs.** Rule 195. EP-04. Rule 197.

---

### Rule 128 — Real-time data is at-least-once with idempotent consumers

The ingestion path is at-least-once. Consumers (alert engine, real-time dashboards, notification dispatchers) are idempotent on `(tenant_id, meter_id, channel_id, ts)` keys. Duplicate readings within an idempotency window are harmless.

**Why.** Exactly-once across an unreliable wire is impossible. At-least-once + idempotent consumer is the canonical robust shape.

**How to apply.** Consumer idempotency key. `tests/property/idempotent_consumer_test.go`.

**Cross-refs.** Rule 35. Rule 36.

---

## Section 7 — Regulatory Pack Discipline (Rules 129–148)

These rules govern Report Packs and Factor Packs. They formalise the property that regulatory output is auditable, repeatable, and resilient to regulatory change.

### Rule 129 — Each regulatory dossier is a Report Pack

ESRS E1 (CSRD), Piano Transizione 5.0 attestazione, Conto Termico 2.0 GSE submission, Certificati Bianchi TEE submission, audit D.Lgs. 102/2014 dossier, monthly consumption report, CO2 footprint report — each is a Report Pack at `packs/report/<id>/`. Core's reporting orchestrator (`internal/handlers/reports.go`) speaks only the `Builder` interface (EP-02).

**Why.** Hard-coding regulatory shapes into Core makes regulatory change a Core release. Pack-based regulatory shapes make the change a Pack release.

**How to apply.** Charter §3.2. EP-02. Phase E Sprint S6 extracts current regulatory code into Packs.

**Cross-refs.** Rule 69. Rule 71.

---

### Rule 130 — Each authoritative factor source is a Factor Pack

ISPRA Italia, GSE renewable shares, Terna national mix, AIB residual mix, UK BEIS / DEFRA, EPA US eGRID, IEA international, EcoInvent connector — each is a Factor Pack at `packs/factor/<id>/`. Core's factor service (`internal/domain/emissions/factor_source.go`) speaks only the `FactorSource` interface (EP-03).

**Why.** Factor sources change annually (ISPRA every April, GSE quarterly, AIB mid-year). Pack-based factor sources make the update a Pack release.

**How to apply.** Charter §3.2. EP-03. Phase E Sprint S6 extracts current factor code into Packs.

**Cross-refs.** Rule 69. Rule 90.

---

### Rule 131 — Regulatory packs validate against the formal spec

Where the regulator publishes a formal spec (XBRL ESRS taxonomy, GSE XML schema for Conto Termico, ENEA XSD for audit 102/2014), the Report Pack validates its output against the spec at build time. Failed validation blocks submission.

**Why.** A submission that the regulator's portal rejects is a regulatory failure. Pre-submission validation is the cheapest possible defence.

**How to apply.** `packs/report/<id>/validation/` carries the spec. Builder validates before persistence. Conformance: `tests/conformance/report_validation_test.go`.

**Cross-refs.** Rule 89. Rule 132.

---

### Rule 132 — Italian regulatory ground truth is annotated to primary sources

Every Italian-specific assertion in a Report Pack (a threshold, a rate, a XSD reference, a deadline) carries a comment citing the primary regulatory source: GU n. NN del DD/MM/YYYY, deliberation number, decree code, MIMIT/MASE circular ID. The annotation enables annual re-verification.

**Why.** Italian regulation cites multiple primary sources (gazzetta, decreto, deliberation, circolare, FAQ). A code without citation drifts silently when the source updates.

**How to apply.** Code comments. `docs/ITALIAN-COMPLIANCE.md` carries the citation map. Annual re-verification runbook.

**Cross-refs.** Rule 47. `docs/COMPLIANCE/`.

---

### Rule 133 — Piano 5.0 thresholds are configurable per attestazione cycle

The Piano 5.0 attestazione thresholds (3% process / 5% site, credit bands 5/20/35/40/45%) live in `packs/report/piano_5_0/config.yaml` and are versioned with the Pack. Changes to thresholds (driven by MIMIT updates) ship as Pack minor versions. Existing reports retain the threshold version they were generated under.

**Why.** Hard-coded thresholds break the moment MIMIT republishes. Configurable + versioned thresholds preserve historical reproducibility.

**How to apply.** Per-Pack config.yaml. Builder reads at build. Conformance: threshold change versioned in Pack.

**Cross-refs.** Rule 90. Rule 97.

---

### Rule 134 — ESRS E1 disclosures are mapped to the EFRAG taxonomy

The ESRS E1 Builder maps each disclosure (E1-1 through E1-9) to the EFRAG XBRL taxonomy 2024 (and subsequent annual revisions) reference URI. The output JSON / XBRL carries the `dimensionURI` for every datapoint. This makes the dossier mechanically consumable by the auditor's EFRAG-validating tooling.

**Why.** A CSRD dossier that looks human-readable but isn't taxonomy-mapped fails the auditor's first automated check.

**How to apply.** `packs/report/esrs_e1/taxonomy.yaml`. Builder annotates output. Conformance: round-trip validation.

**Cross-refs.** Rule 89. Rule 131.

---

### Rule 135 — Conto Termico submission XML is GSE-spec compliant

The Conto Termico Report Pack outputs XML that validates against the GSE-published XSD for the current call (`packs/report/conto_termico/xsd/`). The GSE schema is checked-in versioned; updates ship as Pack minor versions. Submission rejected by GSE for schema reasons is a Pack regression.

**Why.** GSE rejects schema-noncompliant submissions silently in some flows. Pre-submission validation against the canonical XSD is the only defence.

**How to apply.** Per-Pack XSD. Builder validates. Conformance: `tests/conformance/conto_termico_xsd_test.go`.

**Cross-refs.** Rule 131. Rule 106.

---

### Rule 136 — TEE (Certificati Bianchi) submissions are batch-aware

The TEE Report Pack supports batch submission: multiple interventions submitted in one filing. Idempotency keys per-intervention; rejection of one intervention does not invalidate the batch.

**Why.** Industrial customers submit multi-intervention TEE filings. A batch that fails atomically forces resubmission of approved interventions.

**How to apply.** Builder supports batch. Per-intervention idempotency.

**Cross-refs.** Rule 35. Rule 106.

---

### Rule 137 — D.Lgs. 102/2014 audit dossier carries EGE countersignature

The audit-102 dossier includes a counter-signature slot for the engagement's EGE (energy-management expert) certified per UNI CEI 11339. The dossier is finalised only when the EGE has signed; before that, the dossier is in `draft` state.

**Why.** ENEA accepts only EGE-counter-signed audit dossiers. A platform that produces a dossier without the counter-signature produces a dossier that ENEA rejects.

**How to apply.** Workflow state machine in the Pack. EGE-portal integration. Conformance: state-transition test.

**Cross-refs.** Rule 35. Rule 106.

---

### Rule 138 — Regulatory pack updates are queued for annual review

Every Italian Regulatory Report Pack is reviewed annually against the latest primary sources (GU updates, MIMIT FAQ, GSE call announcements). The review checklist is in `docs/COMPLIANCE/ANNUAL-REVIEW.md` (Phase F Sprint S11). A review-driven Pack version bump is a non-emergency Pack release.

**Why.** Italian regulation drifts. Without an annual review cadence, the Pack stays correct only by accident.

**How to apply.** Annual checklist. Calendar reminder. Quarterly office-hours touchpoint.

**Cross-refs.** Rule 47. Rule 132.

---

### Rule 139 — Regulatory thresholds are propagated, not duplicated

A threshold used in multiple Builders (e.g., the 250-employee CSRD threshold) lives in `packs/region/<id>/thresholds.yaml` and is referenced by the consuming Builders. Hard-coding the threshold in a Builder is a Rule-26 rejection.

**Why.** Duplicated thresholds drift. Centralised thresholds update once.

**How to apply.** Region Pack carries thresholds. Builders reference. Custom golangci-lint rule.

**Cross-refs.** Rule 26. Rule 91.

---

### Rule 140 — Per-tenant regulatory profile is explicit

Each tenant declares its applicable regulatory regimes (CSRD wave, audit-102 obligation, Piano 5.0 eligibility, GHG Protocol scope coverage, ETS participation) in the `tenants.regulatory_profile` JSONB. The frontend's regulatory-calendar surfaces only the regimes the tenant participates in.

**Why.** Showing all regimes to all tenants drowns the relevant ones in noise. Per-tenant profiles eliminate the noise.

**How to apply.** Migration adds the column. Frontend filters. Region Pack supplies defaults.

**Cross-refs.** Rule 101. Rule 137.

---

### Rule 141 — Reports are deterministic in serialisation

JSON output serialises with sorted keys. XML output preserves the schema-declared element order. PDF output uses a deterministic font subset and a fixed PDF/A-2b producer string. The same `(period, factors, readings)` always produces the same byte stream.

**Why.** Non-deterministic serialisation breaks bit-perfect reproduction even when the Builder is pure.

**How to apply.** `internal/domain/serialise/` provides deterministic helpers. PDF renderer uses `gofpdf` with fixed options. Conformance: `tests/conformance/report_byte_identity_test.go`.

**Cross-refs.** Rule 89. Rule 91.

---

### Rule 142 — Reports declare their input data window inclusively

Every report's provenance declares the input window with explicit `period_start_inclusive` and `period_end_exclusive` timestamps in UTC, plus the per-tenant timezone for human-readable rendering. Half-open intervals are the canonical shape.

**Why.** "Last 30 days" is ambiguous. Half-open `[start, end)` is unambiguous.

**How to apply.** All builders use half-open. Conformance: `tests/conformance/report_period_test.go`.

**Cross-refs.** Rule 93. Rule 95.

---

### Rule 143 — Scope 3 is opt-in per category and ADR'd

Scope 3 categories are large and methodology-dependent. Each Scope 3 category (1: purchased goods; 2: capital goods; 3: fuel-and-energy; 4: upstream transport; 5: waste; 6: business travel; 7: employee commuting; 8: upstream leased assets; 9: downstream transport; 10: processing of sold products; 11: use of sold products; 12: end-of-life; 13: downstream leased assets; 14: franchises; 15: investments) is opt-in per tenant, declared in the regulatory profile, and configured against an ADR documenting the methodology and the data sources.

**Why.** Default-on Scope 3 produces aggressive estimates that don't survive auditor scrutiny. Default-off + opt-in + ADR'd methodology is auditable.

**How to apply.** Per-tenant opt-in. ADR per category per tenant. `internal/domain/scope3/`.

**Cross-refs.** Rule 89. Rule 132.

---

### Rule 144 — Reports are signed at finalisation

A report transitions `draft → generated → signed → submitted → accepted | rejected`. The `signed` state requires a Cosign sign-blob over the canonical JSON serialisation; the signature is stored in `reports.signature`. `submitted` is reached only after signing.

**Why.** A report submitted without a signature cannot be later proven to have been the one we generated. Signing is the cryptographic anchor.

**How to apply.** State machine. Cosign integration. Conformance: state-transition test.

**Cross-refs.** Rule 95. Rule 169.

---

### Rule 145 — The audit-evidence pack export is regulatory-pack-aware

`task audit-evidence-pack` includes per-Report-Pack evidence: the Pack manifest version, the threshold versions used, the factor versions used, the signed report binary, the lineage JSON, the XSD or taxonomy validation result, the EGE counter-signature where applicable.

**Why.** A complete evidence pack is a complete defence. A partial pack invites questions.

**How to apply.** Audit-evidence builder iterates over Packs.

**Cross-refs.** Rule 108. Rule 99.

---

### Rule 146 — Regulatory PDF cover-letters are template-driven

The PDF cover-letter for each regulatory dossier is rendered from a Pack-supplied template (`packs/report/<id>/templates/cover.html`). The template references the engagement's branding (`config/branding.yaml`). The rendered PDF carries a PDF/A-2b conformance stamp.

**Why.** Hard-coded cover-letters break the moment the engagement rebrands. Templates absorb the change.

**How to apply.** Per-Pack template. PDF renderer. Conformance: `tests/conformance/pdf_pdfa_test.go`.

**Cross-refs.** Rule 81. Rule 141.

---

### Rule 147 — Notifications to GSE/ENEA are tracked

Every outbound submission to GSE / ENEA is tracked in the `regulatory_submissions` table with: `report_id`, `portal_ref`, `submitted_at_utc`, `submitted_by_user_id`, `response_payload`, `response_status`. The state machine of the submission is queryable.

**Why.** GSE / ENEA portals don't always notify on status change. Polling against the tracked submissions is the operations-team's only reliable source of truth.

**How to apply.** Migration adds the table. Submission service writes. Polling worker checks status periodically.

**Cross-refs.** Rule 33. Rule 106.

---

### Rule 148 — Regulatory packs declare their EGE / auditor dependency

A Report Pack that requires external counter-signature (audit 102/2014, Piano 5.0 above certain spend) declares the dependency in its CHARTER. The engagement Phase 0 Discovery includes verifying the engagement's EGE / auditor relationship.

**Why.** A Pack that requires external sign-off but the engagement doesn't have an EGE relationship is a Pack that produces unfileable dossiers.

**How to apply.** Pack CHARTER. Engagement Discovery checklist.

**Cross-refs.** Rule 137. Rule 70.

---

## Section 8 — Engagement Lifecycle (Rules 149–168)

These rules govern how a contracted engagement runs end-to-end — from the signed SoW through Phase 5 handover. They formalise the property that engagements are predictable, time-bounded, and produce operator-ready deployments.

### Rule 149 — Discovery is a deliverable, not an assumption

Phase 0 (Discovery) ends with a signed Scope-of-Work, a chosen deployment topology (A/B/C/D), a Pack matrix listing required Region/Protocol/Factor/Report/Identity Packs, a documented integration map (which client systems connect to the deployment), and an ADR recording the topology choice. No Phase 1 work begins without these artefacts.

**Why.** Engagements that skip Discovery overrun by 50–100% in calendar time. Discovery is the cheapest defence against scope drift.

**How to apply.** `docs/runbooks/engagement-phase-0-discovery.md` (Phase E Sprint S5). SoW template. Discovery checklist.

**Cross-refs.** Charter §8.1. Rule 11.

---

### Rule 150 — The engagement fork is created at Phase 1, not earlier

The engagement fork is created at Phase 1 (Fork & Bootstrap), after the signed SoW. Pre-signature work happens in scratch directories, not in client-named repositories. Once created, the fork follows the upstream-sync discipline of Rule 79 from Day 1.

**Why.** Fork-before-signature creates a phantom liability: the fork exists, the client thinks they own it, but the SoW isn't in place. Discipline at the fork-creation boundary prevents the phantom.

**How to apply.** Engagement-fork creation runbook. Per-engagement audit log row.

**Cross-refs.** Charter §5.2. Rule 77.

---

### Rule 151 — Pack assembly is bounded at Phase 2

Phase 2 (Pack Assembly) takes 3–6 weeks. The Pack matrix from Discovery dictates the scope. Adding a Pack mid-Phase-2 is permitted with a Tradeoff-Stanza ADR; adding more than two Packs mid-Phase-2 is a re-Discovery event.

**Why.** Unbounded Pack assembly is the path to engagement overrun. Bounding it at the Phase keeps the engagement on schedule.

**How to apply.** Engagement timeline template. Mid-Phase-2 ADR. Phase-gate review.

**Cross-refs.** Charter §8.1. Rule 28.

---

### Rule 152 — Customisation Sprint is bounded at Phase 3

Phase 3 (Customisation Sprint) takes 2–4 weeks. The customisation scope is the SoW + the Pack matrix. Beyond-scope requests trigger a change-order with separate pricing and timeline; they don't extend Phase 3.

**Why.** Customisation drift is the second most common cause of engagement overrun. A bounded Phase 3 with a change-order escape valve is the schedule control.

**How to apply.** Engagement change-order template. PR template prompts for change-order linkage.

**Cross-refs.** Rule 151.

---

### Rule 153 — Hardening + soak is bounded at Phase 4

Phase 4 (Hardening & Soak) takes 2 weeks: production deploy, chaos drill, failover drill, capacity test (1×/3×/5× expected load), runbook walkthrough. A failed drill blocks Phase 5; the engagement extends until the drill passes.

**Why.** A handover before drills pass is a handover that produces incidents. Phase 4 is non-negotiable.

**How to apply.** Phase 4 checklist. Drill log per engagement.

**Cross-refs.** Rule 23. Rule 64. Rule 107.

---

### Rule 154 — Handover is operator-readiness, not paperwork

Phase 5 (Handover or Co-managed Start) ends when the operator team self-reports readiness AND the first 7-day on-call shift is owned by them (or by the co-managed lead in T2/T3). Paperwork without operator-readiness is paperwork; paperwork is not the gate.

**Why.** Handover-as-paperwork produces deployments where the operator team can't run incidents. The 7-day on-call shift is the proof.

**How to apply.** Phase 5 checklist. On-call shadow → reverse-shadow → ownership progression.

**Cross-refs.** Rule 23. Rule 28.

---

### Rule 155 — Engagement runbooks are engagement-specific

Every engagement has its own `engagements/<client>/runbooks/` directory carrying the engagement-specific runbooks: client-IdP-down, client-VPN-flap, client-portal-throttling, client-meter-vendor-firmware-update. The Core runbooks at `docs/runbooks/` continue to apply.

**Why.** Engagement-specific failures are engagement-specific. Folding them into Core runbooks creates noise for other engagements.

**How to apply.** Engagement runbook template. Per-engagement on-call hand-out.

**Cross-refs.** Rule 23. Rule 77.

---

### Rule 156 — Engagement health is monitored monthly

Each engagement has a health score updated monthly: deployment uptime, incident count, runbook drill freshness, upstream-sync recency, regulatory-Pack version recency, customer-NPS, customer-renewal-likelihood. Engagements scoring red on ≥ 3 dimensions trigger an executive-sponsor call.

**Why.** Engagements that fail silently are engagements that fail completely. Monthly health surfaces the early warnings.

**How to apply.** Engagement-health dashboard. Per-engagement monthly review.

**Cross-refs.** Rule 79.

---

### Rule 157 — Engagement renewal is intentional, not automatic

Annual maintenance contracts auto-renew with 60-day notice; the engagement-team initiates the renewal conversation 90 days before expiry. Auto-renewal is not the default sales motion. Renewal is treated as a re-evaluation: did the deployment deliver?

**Why.** Auto-renewal accumulates dead-weight engagements. Intentional renewal forces the conversation.

**How to apply.** Renewal calendar. Pre-renewal checklist.

**Cross-refs.** Charter §9.

---

### Rule 158 — Termination produces an exit pack

When an engagement terminates (annual non-renewal, customer-requested termination, vendor change), Macena delivers an exit pack: full source on the version they were running, all data exports, all runbook walkthrough recordings, all keys (rotated to a client-only chain), all dashboards as JSON, all ADRs frozen. The exit pack is single-command-generated by `task engagement-exit-pack`.

**Why.** Painful exits produce reputation damage. Smooth exits produce referral business.

**How to apply.** `task engagement-exit-pack`. Exit-pack template.

**Cross-refs.** Rule 158. Rule 108.

---

### Rule 159 — Engagement-specific code never lands upstream without generalisation

If an engagement-specific feature is going to be reused, it is generalised and contributed back as a Pack or as a Core enhancement. Direct inclusion of `engagements/<client>/<feature>/` into upstream is forbidden — every upstream contribution passes through generalisation.

**Why.** Direct inclusion lifts customer-specific assumptions into the template. Generalisation forces "what's invariant?" thinking.

**How to apply.** Generalisation checklist. PR template asks "is this a generalisation?"

**Cross-refs.** Rule 69. Rule 77.

---

### Rule 160 — Engagement post-mortems contribute to the doctrine

Every engagement closes with a post-mortem capturing: what went well, what went badly, what's reusable, what's not. A post-mortem item that recurs across two engagements is a candidate for a doctrine rule (Rule 209 process).

**Why.** Post-mortems without a path to the doctrine produce knowledge that's lost.

**How to apply.** Post-mortem template. Per-engagement final document. Quarterly doctrine review.

**Cross-refs.** Rule 209.

---

### Rule 161 — On-call rotation in T2/T3 is documented and rotated

Co-managed (T2) and Fully-managed (T3) engagements run on-call rotations with at least three engineers. Pager-duty is rotated weekly; primary / secondary roles. The rotation is documented in `engagements/<client>/ON-CALL-SCHEDULE.md`.

**Why.** A single-person rotation is a single point of failure. Three-person rotation is the minimum viable resilience.

**How to apply.** PagerDuty / OpsGenie integration. Schedule template.

**Cross-refs.** Rule 23.

---

### Rule 162 — Engagement reviews are tri-annual

Every engagement has a review at month 4 (post-handover stability), month 8 (renewal-readiness), month 12 (renewal-decision). Reviews are short (30 min) and produce a written summary. Reviews are added to the engagement health dashboard.

**Why.** Reviews on a regular cadence catch drift before it becomes a crisis.

**How to apply.** Review template. Calendar reminders.

**Cross-refs.** Rule 156.

---

### Rule 163 — Pricing is transparent at engagement start

The engagement-pricing model (license + customisation + maintenance + tier retainer) is transparent at SoW signing. Every later commercial discussion (additional Packs, additional sites, additional features) is priced against the same rate card. Pricing surprises are not part of the model.

**Why.** Surprise pricing erodes trust. Transparent pricing produces multi-year customers.

**How to apply.** Rate card. Quote template.

**Cross-refs.** Charter §9.

---

### Rule 164 — Customer data ownership is contractual and technical

The MSA explicitly assigns data ownership to the client. The Macena team acts as a data processor under Art. 28 GDPR. Technical implementations match: per-tenant DEK keys are exportable to the client's KMS at termination; data-region-allowlist is per-tenant.

**Why.** Cloud-vendor lock-in arguments lose to "you own your data on your terms." Technical match prevents accidents.

**How to apply.** MSA template. DEK rotation runbook.

**Cross-refs.** Rule 184. `docs/COMPLIANCE/GDPR.md`.

---

### Rule 165 — Engagement testing fixtures are synthetic by default

Engagement test fixtures are synthetic data (`packs/region/<id>/fixtures/synthetic/` for Pack-level; `engagements/<client>/fixtures/synthetic/` for engagement-level). Real client data appears in tests only if (a) the client has explicitly approved (b) the data is anonymised per Italian-Garante guidance.

**Why.** Real customer data in tests becomes real customer data in commits, then in CI logs, then in incident chats. Synthetic by default eliminates the risk class.

**How to apply.** Conformance: `tests/static/no_real_pii_test.go` (Phase F Sprint S10).

**Cross-refs.** Rule 165. `docs/COMPLIANCE/GDPR.md`.

---

### Rule 166 — Engagement support tier upgrades are explicit

Moving an engagement from T1 to T2 (or T2 to T3) is a contracted change, not a slow-motion drift. The change-order names the new tier, the new SLA, the new retainer, the additional named on-call.

**Why.** Tier drift produces engagements where the obligations don't match the revenue. Explicit upgrades match.

**How to apply.** Tier change-order. PagerDuty role update.

**Cross-refs.** Rule 161.

---

### Rule 167 — Crisis communication is rehearsed

The engagement-crisis communication playbook (`docs/INCIDENT-RESPONSE.md` Engagement section, Phase F Sprint S10) covers: who calls the customer first; what's said in the first 15 minutes; the cadence of subsequent updates; the postmortem timeline. Annual tabletop exercise per engagement (T2/T3).

**Why.** A crisis without a communication plan compounds. Rehearsed communication is the only kind that survives the first 15 minutes.

**How to apply.** Crisis playbook. Tabletop exercise.

**Cross-refs.** Rule 61. `docs/COMPLIANCE/NIS2.md`.

---

### Rule 168 — Lessons from the engagement portfolio feed back to roadmap

Quarterly, the engagement-team aggregates the lessons across the active portfolio: most-requested Pack, most-frequent Pack-contract pain, most-common runbook gap, most-recurring incident class. The aggregate feeds the Macena roadmap and the doctrine review.

**Why.** A roadmap built from speculation drifts. A roadmap built from portfolio feedback stays grounded.

**How to apply.** Quarterly portfolio review. Roadmap backlog.

**Cross-refs.** Rule 168. Rule 209.

---

## Section 9 — Cryptographic Invariants (Rules 169–188)

These rules govern the cryptographic surface — signing, encryption, KDF, key rotation, identity. They make tamper-evidence a property, not an aspiration.

### Rule 169 — The audit log is signed in chained-HMAC

Every `audit_log` row carries `prev_hmac` and `row_hmac`. `row_hmac = HMAC-SHA256(audit_chain_key, prev_hmac || row_canonical_json)`. The chain anchors at a published signed checkpoint every hour. Tampering with any row breaks the chain at that row.

**Why.** A merely append-only audit log can be wholesale-replaced. A chained-HMAC audit log can be detected as tampered.

**How to apply.** Migration adds the columns. Insert trigger computes the HMAC. Hourly checkpoint job. Conformance: `tests/security/audit_chain_test.go` (Phase F Sprint S11).

**Cross-refs.** Rule 62. Rule 174.

---

### Rule 170 — JWT tokens carry `kid`, are HS256, are validation-pinned

The JWT `kid` claim identifies the signing key. The validator pins `[]string{"HS256"}` (REJ — alg-confusion attacks). JWT secrets are KMS-stored, rotated quarterly, with a 24h dual-key window per ADR-0016.

**Why.** alg=none and alg=RS256 confusion attacks have produced real breaches. Validation-pinning is the canonical defence. KID rotation prevents long-term key compromise.

**How to apply.** `internal/security/jwt.go`. ADR-0016. Conformance: `tests/security/jwt_alg_pinning_test.go`.

**Cross-refs.** Rule 7 of cryptographic hygiene literature. ADR-0016.

---

### Rule 171 — Cosign signing is keyless via Sigstore Fulcio

Production images, manifest locks, signed reports, and signed audit-log checkpoints are signed via Cosign keyless OIDC. The trust anchor is the Sigstore CT log. No long-lived signing keys; ephemeral signing certs from Fulcio.

**Why.** Long-lived signing keys are the supply-chain compromise vector. Keyless eliminates the class.

**How to apply.** ADR-0017. CI keyless signing job. Verification at admission (Kyverno).

**Cross-refs.** Rule 57. REJ-27.

---

### Rule 172 — Field-level encryption uses per-tenant DEK + KMS-wrapped MEK

`readings.raw_payload` is field-level encrypted with AES-256-GCM. The DEK (data encryption key) is per-tenant, generated at tenant creation, wrapped by a KMS master key (MEK). The KMS key is regional. Key rotation is supported with a versioned `dek_id` column.

**Why.** Per-tenant DEK enables crypto-shredding (Rule 104). KMS-wrapped MEK prevents in-database key exposure. Versioned DEK supports rotation.

**How to apply.** Migration adds the columns. KMS configured per-region. Crypto-shredding runbook.

**Cross-refs.** Rule 104. Rule 184.

---

### Rule 173 — Meter signatures use HMAC-SHA256 with per-meter keys

Each meter is provisioned with a per-meter HMAC key on its first onboarding. The key is stored in the customer's KMS (or in a sealed enclave on the edge gateway). Reading signatures use HMAC-SHA256 over `(meter_id, channel_id, ts_ns, value_int_micro_unit)`.

**Why.** A per-meter key means one compromised meter doesn't compromise all meters. HMAC-SHA256 is fast enough for high-rate sampling and trivially verifiable.

**How to apply.** Meter-provisioning runbook. Edge-gateway sealed-key store. Phase F Sprint S11 wire-up.

**Cross-refs.** Rule 105. Rule 178.

---

### Rule 174 — TLS is 1.3-only on customer-facing edges

External-facing edges (frontend, API gateway, customer-OAuth callbacks) accept TLS 1.3 only. TLS 1.2 is permitted only on internal edges where a legacy client requires it (declared per-engagement in `engagements/<client>/CRYPTO-EXCEPTIONS.md`). HSTS is enabled with `max-age=31536000; includeSubDomains; preload`.

**Why.** TLS 1.3 closes the entire CBC / RC4 / weak-DH attack surface. TLS 1.2 is acceptable for legacy compatibility but is the exception.

**How to apply.** Ingress controller config. TLS-version conformance test.

**Cross-refs.** Rule 8.

---

### Rule 175 — Certificate rotation is automated; manual rotation is the runbook

Public-facing certs rotate via cert-manager + Let's Encrypt ACME. Internal mTLS certs rotate via cert-manager + an in-cluster Issuer. Manual rotation is a runbook (`docs/runbooks/cert-rotation.md`) for the rare case the automation breaks.

**Why.** Cert expiry is the most common preventable outage. Automation eliminates the class.

**How to apply.** cert-manager Helm chart. Runbook.

**Cross-refs.** Rule 23. ADR-0020.

---

### Rule 176 — Random number generation uses `crypto/rand`

UUIDs, nonces, secret material, idempotency keys (for system-generated cases), session tokens — all use `crypto/rand`. `math/rand` is forbidden in any path that touches a security boundary. A custom golangci-lint rule enforces.

**Why.** `math/rand` produces predictable output. Predictability + secret material = breach.

**How to apply.** golangci-lint rule. Conformance: `tests/static/no_math_rand_in_security_test.go`.

**Cross-refs.** Rule 39.

---

### Rule 177 — Password storage is bcrypt with cost factor 12

User passwords (in the local-DB Identity Pack — Identity Packs that delegate to an external IdP don't store passwords) are bcrypt-hashed with cost factor 12, with a per-deployment pepper from KMS. No SHA-1, no MD5, no plain SHA-256. The cost factor is reviewed annually as hardware advances.

**Why.** Bcrypt is the canonical answer for password storage. Cost factor 12 is the 2026 threshold for adequate protection against offline brute force.

**How to apply.** `internal/security/password.go`. Annual cost-factor review.

**Cross-refs.** Rule 39.

---

### Rule 178 — Webhook signatures use HMAC-SHA256 with constant-time compare

Inbound webhook signatures (Pulse meters, GSE callbacks, customer custom integrations) use HMAC-SHA256. Compare uses `subtle.ConstantTimeCompare` (REJ-20). Replay protection: a `nonce` and a recently-seen-nonce cache for at least 5 minutes.

**Why.** Plaintext compare is a timing oracle. Replay without nonce protection lets an attacker re-fire captured webhooks.

**How to apply.** `internal/security/webhook.go`. Conformance: `tests/security/webhook_constant_time_test.go`.

**Cross-refs.** Rule 118. REJ-20.

---

### Rule 179 — KMS keys are regional; cross-region transfer requires explicit policy

KMS keys live in the deployment's region (eu-south-1 or the engagement's region). Cross-region key access requires an explicit IAM policy citing the use case. Default-deny on cross-region.

**Why.** Cross-region key access is an exfiltration vector if it's permissive.

**How to apply.** IAM policy templates. Terraform module.

**Cross-refs.** Rule 8. Rule 172.

---

### Rule 180 — Secrets-rotation runbook is annual or post-incident

Secrets rotate annually as a baseline; post-incident immediately. The rotation runbook (`docs/runbooks/secret-rotation.md`) covers JWT, DB, KMS, ESO ClusterSecretStore, OAuth client secrets, webhook HMAC keys.

**Why.** Long-lived secrets accumulate exposure surface (logs, backups, ex-employee access). Annual rotation caps the half-life.

**How to apply.** Runbook. Calendar reminder. ESO automatic rotation where supported.

**Cross-refs.** Rule 20.

---

### Rule 181 — TOTP is the second factor for admin accounts

Admin and operator accounts in the local-DB Identity Pack require TOTP as a second factor. Identity Packs that delegate to external IdPs inherit the IdP's MFA discipline. Single-factor admin login is forbidden in production.

**Why.** Admin compromise is a regulator-grade incident. MFA cuts the credential-stuffing class.

**How to apply.** `internal/security/totp.go`. Login flow.

**Cross-refs.** Rule 39.

---

### Rule 182 — IP and email-based lockout for repeated failed login

Failed-login lockout: 5 attempts in 15 minutes from the same IP+email pair triggers a 15-minute lockout. The lockout state is in `auth_lockouts` (Phase E Sprint S5 hardens this). Distinct from rate-limiting.

**Why.** Credential-stuffing attacks rely on unbounded retry. Lockout caps the attack rate.

**How to apply.** `internal/security/auth_lockout.go`. Lockout dashboard.

**Cross-refs.** Rule 39.

---

### Rule 183 — Session tokens are revocable and have idle + absolute lifetimes

Session tokens (refresh tokens) carry an idle lifetime (15 min default) and an absolute lifetime (12 hours default). Revocation lists are kept in Redis with the same TTL. Logged-out tokens are immediately revoked.

**Why.** Long-lived sessions extend breach windows. Revocable + lifetime-capped sessions cap the windows.

**How to apply.** Token issuance config. Revocation API.

**Cross-refs.** Rule 170.

---

### Rule 184 — Crypto-shredding handles GDPR Art. 17

Right-of-erasure for natural persons is satisfied by crypto-shredding the per-tenant DEK. The audit-log row remains for audit retention; the encrypted personal data inside it becomes unreadable. The shredding event is itself audit-logged.

**Why.** Hard delete violates audit retention. Soft delete violates Art. 17. Crypto-shredding satisfies both.

**How to apply.** DSAR endpoint. Per-tenant DEK lifecycle. Conformance: `tests/conformance/crypto_shredding_test.go`.

**Cross-refs.** Rule 104. Rule 172.

---

### Rule 185 — Cryptographic primitives use the standard library where possible

`crypto/aes`, `crypto/sha256`, `crypto/hmac`, `crypto/subtle`, `golang.org/x/crypto/...` for vetted extensions. Bring-your-own crypto is forbidden. Custom-rolled crypto patterns are rejected at code review.

**Why.** Custom crypto fails in subtle ways. Standard library is vetted.

**How to apply.** Code review. Custom golangci-lint rule.

**Cross-refs.** Rule 26.

---

### Rule 186 — Crypto agility: algorithm choices are versioned

The DEK algorithm (AES-256-GCM today), the audit-chain HMAC algorithm (HMAC-SHA256 today), the meter-signature algorithm (HMAC-SHA256 today), the password algorithm (bcrypt cost 12 today) are all versioned in the schema. Migration to a future algorithm rolls existing data over a versioned upgrade path.

**Why.** SHA-1 retirement, MD5 retirement, bcrypt-cost increases — they all happen. A schema that doesn't anticipate the migration is a schema that creates a fire drill.

**How to apply.** Schema columns: `dek_algorithm`, `chain_hmac_algorithm`, `signature_algorithm`, `password_algorithm`. Migration plays for each upgrade.

**Cross-refs.** Rule 33.

---

### Rule 187 — Quantum-resistant crypto plan is on the radar (not yet implemented)

A documented plan in `docs/QUANTUM-CRYPTO-PLAN.md` (Phase J Sprint S22) addresses the path to post-quantum signatures (Dilithium / Falcon / SPHINCS+) and KEM (Kyber). Implementation is deferred until NIST finalises and Cosign / Sigstore expose primitives. The plan is a placeholder that reviewers can point at.

**Why.** Post-quantum is going to land in the 2027–2030 window. Having no plan is a strategic vulnerability; a placeholder plan is preparedness.

**How to apply.** Plan document. Annual review of the NIST timeline.

**Cross-refs.** Rule 186.

---

### Rule 188 — Cryptography failures are Sev-1 by default

A cryptography-related defect (signature verification skipped, plaintext compare slipped through, secret leaked, KDF wrong) is Sev-1 by default. The defect is fixed within 24 hours; the post-mortem is filed within 5 business days.

**Why.** Crypto bugs that aren't fixed fast become incidents. The Sev-1 default forces speed.

**How to apply.** PR template prompts severity. Pre-commit gitleaks. Bug-tracker template.

**Cross-refs.** Rule 61.

---

## Section 10 — AI/ML Reproducibility (Rules 189–208)

These rules govern the AI/ML surface — the consumption forecaster, the anomaly detector, the flexibility-market connector, the eventual peak-shaving simulator. They formalise the property that AI-derived outputs are reproducible, explainable, and don't quietly become regulatory artefacts.

### Rule 189 — AI/ML outputs are not regulatory artefacts unless explicitly opted in

Forecasts, anomaly detections, and AI-derived insights are decision-support, not regulatory output. ESRS E1 dossier, Piano 5.0 attestazione, Conto Termico submission, audit 102/2014 dossier — none of them include AI-derived figures by default. Inclusion in a regulatory output requires per-deployment ADR.

**Why.** AI outputs are probabilistic. Regulatory output is deterministic. Mixing them silently is a regulatory failure waiting to happen.

**How to apply.** Builder code paths exclude AI-derived columns. Per-tenant opt-in flag with an ADR.

**Cross-refs.** Rule 89. Rule 95.

---

### Rule 190 — Every model carries a model card

Every deployed model has a model card (`models/<model-id>/MODEL-CARD.md`) documenting: training data window, training data sources, hyperparameters, evaluation metrics, intended use, known failure modes, dataset bias notes. The model card is part of the SBOM-equivalent for ML.

**Why.** A model without a card is a model without a reviewable contract. Auditors who don't speak ML can read the card; auditors who do speak ML can read the card *and* run their own validation against the documented evaluation set.

**How to apply.** Model card template. Per-model directory. CI verifies card exists before model deploys.

**Cross-refs.** Rule 47. Rule 197.

---

### Rule 191 — Training data lineage is captured

Every training run captures: dataset hashes, dataset sources, dataset windows, training environment (Python version, library versions, container hash), hyperparameter set, eval-set hashes. The lineage is stored in `models/<model-id>/training-runs/<run-id>/` and signed (Cosign sign-blob).

**Why.** A model whose training data can't be audited is a model whose claims can't be audited. Training-data lineage is the engineering equivalent of clinical-trial registration.

**How to apply.** Training pipeline emits the lineage manifest. Per-run signed.

**Cross-refs.** Rule 89. Rule 95.

---

### Rule 192 — Forecasts are reproducible from frozen inputs

A forecast for `(tenant_id, period_start, period_end, model_id, model_version, input_data_hash)` is byte-for-byte reproducible. Determinism is enforced by: pinned random seeds, pinned library versions, frozen input data, frozen model artefact.

**Why.** A forecaster whose output drifts on rerun is a forecaster whose claims can't be checked.

**How to apply.** Inference pipeline pins seed. Conformance: `tests/conformance/forecast_determinism_test.go` (Phase I Sprint S18).

**Cross-refs.** Rule 89. Rule 91.

---

### Rule 193 — Forecast accuracy is monitored continuously

Each forecast is paired (after the actual outcome lands) with an error metric (MAE, MAPE, RMSE) per horizon (24h, 7d). The error metrics are tracked in Prometheus. Alerts fire when sustained error exceeds the documented budget (MAPE ≤ 12% day-ahead, ≤ 25% week-ahead).

**Why.** A forecaster without continuous accuracy monitoring is a forecaster that quietly degrades.

**How to apply.** Per-forecast error computation. Grafana dashboard. Alerts.

**Cross-refs.** Rule 24. Rule 197.

---

### Rule 194 — Drift detection is automatic

Distribution drift (Kolmogorov-Smirnov on input features), output drift (mean/variance of predictions over time), and label drift (when ground truth lands later) are computed and alerted. Drift exceeding the documented threshold triggers a model-retraining decision in the next sprint.

**Why.** A model that drifts silently produces increasingly wrong outputs. Drift detection is the early warning.

**How to apply.** Drift-detection job. Per-model thresholds. Alerts.

**Cross-refs.** Rule 193. Rule 197.

---

### Rule 195 — Anomaly-detection layers each carry a model card

Layer 1 (z-score on rolling baseline): model card documents the baseline window and z-score threshold. Layer 2 (STL seasonal decomposition): model card documents the seasonality assumptions. Layer 3 (cross-meter correlation): model card documents the correlation matrix and the threshold.

**Why.** "It's just z-score, no AI" is a misconception that produces undocumented detection. Even simple statistics are model-decisions worthy of a model card.

**How to apply.** Per-layer model card.

**Cross-refs.** Rule 127. Rule 190.

---

### Rule 196 — Forecast inputs declare their staleness budget

Each input feature (weather forecast, production plan, historical baseline) declares a staleness budget. A forecast computed against stale inputs is annotated `data_freshness=stale` in the response. Consumers branch on the annotation.

**Why.** A forecast computed against 8-hour-old weather data is no longer a forecast — it's a number. Staleness annotation makes the property visible.

**How to apply.** Per-feature staleness in metadata. Forecast endpoint annotates.

**Cross-refs.** Rule 36.

---

### Rule 197 — Model retraining is sprint-gated and ADR'd

Retraining a model is a sprint-gated decision: the previous model card is updated, the new training-run lineage is captured, the previous-vs-new evaluation comparison is documented in an ADR. The new model is canary'd before full rollout.

**Why.** Continuous retraining without gates produces a moving target nobody can audit. Sprint-gated retraining produces auditable cadence.

**How to apply.** Retraining sprint task. ADR template. Canary deploy.

**Cross-refs.** Rule 56. Rule 197.

---

### Rule 198 — Models are versioned and rolled back

Every model has a semantic version. The deployed version is recorded in the per-prediction lineage. Rollback to a previous version is a one-command operation. Two versions can run side-by-side during a canary.

**Why.** Without versioning, "which model produced this?" is unanswerable. Without rollback, a bad model is locked in.

**How to apply.** Model registry. Per-prediction `model_version` column.

**Cross-refs.** Rule 71. Rule 197.

---

### Rule 199 — Inference is observable

Every inference request emits OTel spans (model name, version, input hash, output, latency). Latency budgets per model: forecast P95 ≤ 300ms; anomaly detection P95 ≤ 100ms. Budget violations alert.

**Why.** Inference latency that surprises in production is inference latency that wasn't observable. Per-model OTel spans surface the property.

**How to apply.** Inference SDK wraps. Grafana dashboard.

**Cross-refs.** Rule 40. Rule 124.

---

### Rule 200 — AI explainability is a deliverable

Every forecast and every anomaly carries an explanation: top-k input features, SHAP values where applicable, comparable historical example. The explanation is part of the response payload (`explanation` field) and is rendered in the operator UI.

**Why.** "The AI says so" is not a defensible answer to an operator who has to act on the output. Explanations are the bridge to action.

**How to apply.** Per-model explanation routine. Frontend renders. Conformance: `tests/conformance/explanation_present_test.go`.

**Cross-refs.** Rule 99.

---

### Rule 201 — Model fairness review is part of release

A model whose output influences customer-facing decisions (forecast → flexibility-market commitment, anomaly → operator alert prioritisation) is reviewed for fairness: does the model behave equivalently across tenant cohorts? Per-cohort error metrics are tracked.

**Why.** A model that's accurate on the majority cohort but wrong on a minority cohort is a model that produces unfair outcomes. Fairness review surfaces the property.

**How to apply.** Fairness-review checklist in retraining sprint. Per-cohort metrics dashboard.

**Cross-refs.** Rule 197.

---

### Rule 202 — AI/ML libraries are pinned and digest-locked

PyTorch / sklearn / lightgbm / TensorFlow / XGBoost / numpy / pandas / scikit-learn — all pinned via `requirements.lock` with hash-locking. Container base image is digest-pinned. The training image is signed and SLSA-attested.

**Why.** ML libraries change behaviour subtly between minor versions. Pinning eliminates the silent change class.

**How to apply.** Hash-locked requirements. Per-image signature.

**Cross-refs.** Rule 53. Rule 57.

---

### Rule 203 — AI privacy: training data PII handling is documented

Training data PII handling is documented in the model card: which fields were used, which were excluded, which were anonymised, which were aggregated. PII inclusion in training requires a Tradeoff Stanza ADR.

**Why.** PII in training data is a GDPR risk class. Documented handling is the only auditable answer.

**How to apply.** Model card. ADR. Annual data-protection review.

**Cross-refs.** Rule 184. `docs/COMPLIANCE/GDPR.md`.

---

### Rule 204 — AI bias evaluation is documented

Each model's bias evaluation (per-tenant, per-meter-class, per-region) is documented in the model card. Specific metrics: per-cohort MAPE, per-cohort false-positive-rate, per-cohort false-negative-rate.

**Why.** Bias that's not measured is bias that's not addressed. Per-cohort evaluation forces awareness.

**How to apply.** Model card. Bias-evaluation script. Per-cohort dashboards.

**Cross-refs.** Rule 201.

---

### Rule 205 — Consumption forecaster ships only after passing the evaluation budget

The consumption forecaster (Phase I Sprint S18) ships only after passing the documented evaluation budget on a held-out test set: MAPE ≤ 12% day-ahead, ≤ 25% week-ahead, with statistically-significant improvement over the seasonal-baseline. A forecaster that doesn't beat seasonal baseline is not deployed.

**Why.** A forecaster worse than seasonal baseline is overhead. The bar is non-negotiable.

**How to apply.** Pre-deploy evaluation. ADR documents the result.

**Cross-refs.** Rule 197. Rule 200.

---

### Rule 206 — AI integration with the flexibility market (MSD) is opt-in per-tenant

The MSD (Mercato del Servizio di Dispacciamento) integration (Phase I Sprint S19) commits a tenant's peak-shaving capacity to the Italian flexibility market. Opt-in is per-tenant, with an explicit consent capture (the tenant's CFO signs off; the consent is in the audit log). Default off.

**Why.** Auto-opting tenants into a financial market is a contractual and reputational risk class. Explicit consent eliminates the class.

**How to apply.** Consent capture flow. Audit-log entry. Per-tenant flag.

**Cross-refs.** Rule 164.

---

### Rule 207 — AI/ML failures degrade gracefully

A failed inference doesn't crash the request. The forecast endpoint returns a documented fallback (last-known-good or seasonal baseline) with `degraded_mode=true` annotation. The anomaly endpoint returns "no anomaly detected" with the annotation. Calling code branches on the annotation.

**Why.** A request that fails because the model failed is a worse experience than a request that returns the seasonal baseline.

**How to apply.** Per-endpoint fallback. Conformance: `tests/conformance/inference_graceful_test.go`.

**Cross-refs.** Rule 36. Rule 207.

---

### Rule 208 — AI/ML termination criterion is named per-model

Each model has a termination criterion: "the model continues to be deployed only as long as it demonstrates ≥ X% improvement over the documented baseline on the per-cohort evaluation." Models that fail the criterion are retired. Retired models become `deprecated` and the inference endpoints fall back to the baseline.

**Why.** Models without termination criteria become permanent overhead. Models with criteria stay in healthy turnover.

**How to apply.** Per-model termination criterion in the model card. Quarterly review.

**Cross-refs.** Rule 28. Rule 197.

---

## Section 11 — Doctrine Meta-rules (Rules 209–210)

### Rule 209 — Adding a doctrine rule requires evidence and a Tradeoff Stanza

A new doctrine rule is added through an ADR. The ADR includes: the failure mode the rule addresses, evidence from at least one engagement or one near-incident, the proposed rule body (title + Why + How to apply + Cross-refs), the four-part Tradeoff Stanza, and the impact on existing rules (does it supersede / amend / interact with any?). Quarterly office hours confirm new rules.

**Why.** Doctrine that grows without evidence becomes folklore. Evidence-grounded growth produces rules that survive review.

**How to apply.** ADR template. Quarterly office hours.

**Cross-refs.** Rule 27. Rule 47.

---

### Rule 210 — Removing a doctrine rule requires a recorded supersession

A doctrine rule is removed (or significantly amended) only through an ADR explicitly citing `supersedes-rule: NNN` and providing evidence that the rule's failure mode no longer applies (or the rule is being replaced by a stronger rule). Quarterly office hours confirm. The removed rule moves to the "Retired Rules" appendix at the end of this file with the supersession date.

**Why.** Doctrine that's freely removed becomes negotiable. Recorded supersession means every rule's lineage is reviewable.

**How to apply.** ADR template. Quarterly office hours. Retired-rules appendix below.

**Cross-refs.** Rule 209. Rule 26.

---

## Retired Rules (appendix)

(empty — populates as rules are superseded per Rule 210)

---

## Quick-reference table (all 208 active rules)

| # | Title | Group |
|---|---|---|
| 1 | Money is `(amount_cents int64, currency ISO-4217 string)` | Universal invariant |
| 2 | Timestamps are RFC 3339 UTC with explicit offset | Universal invariant |
| 3 | Tenant identifiers are UUIDv4 | Universal invariant |
| 4 | Errors are RFC 7807 Problem Details | Universal invariant |
| 5 | Events are CloudEvents 1.0 | Universal invariant |
| 6 | Health envelope shape is fixed | Universal invariant |
| 7 | Logs are structured JSON with mandatory context | Universal invariant |
| 8 | Italian residency is the default; cross-EU transfer is opt-in only | Universal invariant |
| 9 | Platform engineering is a discipline, not a label | Platform |
| 10 | Platform serves users; users are application teams | Platform |
| 11 | Initiatives follow the Rule 11/31/51 sequence | Platform |
| 12 | Platform thinking is opinionated defaults | Platform |
| 13 | Abstractions are cost centres until proven otherwise | Platform |
| 14 | Contracts come first; code is the implementation | Platform |
| 15 | Layers are the system map; the system map is the architecture | Platform |
| 16 | Infrastructure is code; state is centralised but partitioned | Platform |
| 17 | Developer experience is a first-class platform output | Platform |
| 18 | Telemetry sampling is policy, not arbitrary | Platform |
| 19 | Sentinel detection refuses to boot in production | Platform |
| 20 | Secrets are managed; never in code, never in K8s | Platform |
| 21 | Evolution and change management are first-class | Platform |
| 22 | Events are the integration surface; not RPC | Platform |
| 23 | The substrate is operable by a single on-call | Platform |
| 24 | Verification is continuous and automatic | Platform |
| 25 | Quality threshold is regulator-grade | Platform |
| 26 | Rejection authority is named and exercised | Platform |
| 27 | Decisions carry the four-part Tradeoff Stanza | Platform |
| 28 | Termination criterion is named at the outset | Platform |
| 29 | Backend code is intent-revealing | Backend |
| 30 | Backend is a first-class system, not a side-effect of the framework | Backend |
| 31 | Architecture proposals carry a named alternative | Backend |
| 32 | Domain-driven design lives in `internal/domain/` | Backend |
| 33 | Data is the system | Backend |
| 34 | Contract-first applies inside the backend too | Backend |
| 35 | Consistency and state guarantees are explicit | Backend |
| 36 | Failure is normal; the system is designed for failure | Backend |
| 37 | Performance is a design constraint, not a tuning task | Backend |
| 38 | Scalability is intentional, not magical | Backend |
| 39 | Security is structural, not bolted on | Backend |
| 40 | Observability is structural | Backend |
| 41 | Concurrency is bounded and explicit | Backend |
| 42 | Resources have lifecycles; lifecycles are managed | Backend |
| 43 | Framework and infrastructure awareness is explicit | Backend |
| 44 | Testability is a design output | Backend |
| 45 | Backend quality threshold is regulator-grade | Backend |
| 46 | Backend rejection authority is exercised | Backend |
| 47 | Decision rationale travels with the decision | Backend |
| 48 | Backend termination criterion is named | Backend |
| 49 | DevSecOps is a discipline, not a label | DevSecOps |
| 50 | DevSecOps is a unified system | DevSecOps |
| 51 | DevSecOps initiatives carry operational readiness | DevSecOps |
| 52 | Identity in CI/CD is keyless and ephemeral | DevSecOps |
| 53 | Pin every dependency to a digest | DevSecOps |
| 54 | Policy as code is the gating mechanism | DevSecOps |
| 55 | Policy gates are layered: IaC → Manifest → Admission → Runtime | DevSecOps |
| 56 | CD gates are layered, not single-step | DevSecOps |
| 57 | Supply chain is end-to-end attested | DevSecOps |
| 58 | SBOM is generated, attested, and queried | DevSecOps |
| 59 | Vulnerability response is SLA'd | DevSecOps |
| 60 | Pen-test cadence is annual + post-major-change | DevSecOps |
| 61 | Incident response is rehearsed | DevSecOps |
| 62 | Audit logging is append-only and immutable | DevSecOps |
| 63 | Compliance evidence is exportable | DevSecOps |
| 64 | Verification is continuous, not periodic | DevSecOps |
| 65 | DevSecOps quality threshold is regulator-grade | DevSecOps |
| 66 | DevSecOps rejection authority is exercised | DevSecOps |
| 67 | DevSecOps decision rationale travels with the decision | DevSecOps |
| 68 | DevSecOps termination criterion is named | DevSecOps |
| 69 | Core and Pack are the load-bearing distinction | Modular Template |
| 70 | A Pack manifests itself; a Pack does not announce itself | Modular Template |
| 71 | Pack contracts are versioned independently of Core | Modular Template |
| 72 | Pack registration is via Registrar, never via global | Modular Template |
| 73 | Boot writes a manifest lock | Modular Template |
| 74 | Pack health is part of Core health | Modular Template |
| 75 | Pack capabilities are declared, not discovered | Modular Template |
| 76 | Pack failures are isolated; they do not crash Core | Modular Template |
| 77 | Engagement code lives in `engagements/<client>/` | Modular Template |
| 78 | Core changes are merge-friendly between minor versions | Modular Template |
| 79 | Engagement forks sync upstream quarterly minimum | Modular Template |
| 80 | Core customisations in a fork are time-bounded | Modular Template |
| 81 | Branding is configuration, not code | Modular Template |
| 82 | Configuration is layered: Core defaults → Pack defaults → Engagement overrides | Modular Template |
| 83 | Pack contracts forbid recursive Pack discovery | Modular Template |
| 84 | Pack tests run independently and against Core | Modular Template |
| 85 | Pack-loader instrumentation is deep | Modular Template |
| 86 | Pack contracts are documented in code, not just text | Modular Template |
| 87 | Pack acceptance criteria are objective | Modular Template |
| 88 | The Italian Region Pack is the flagship reference Pack | Modular Template |
| 89 | Every regulatory output is bit-perfect reproducible from source | Audit-Grade |
| 90 | Factor sources are versioned with temporal validity | Audit-Grade |
| 91 | Builders are pure functions | Audit-Grade |
| 92 | Aggregation is associative, commutative, and verifiable | Audit-Grade |
| 93 | Time bucketing is deterministic and named | Audit-Grade |
| 94 | Floats forbidden in regulatory paths | Audit-Grade |
| 95 | Every report carries a provenance bundle | Audit-Grade |
| 96 | Source data is immutable in the audit window | Audit-Grade |
| 97 | Algorithm changes are versioned and ADR'd | Audit-Grade |
| 98 | Replay test runs every Builder against pinned fixtures | Audit-Grade |
| 99 | Data lineage is queryable | Audit-Grade |
| 100 | Schema changes are additive on regulatory paths | Audit-Grade |
| 101 | Time-zone configuration is per-tenant explicit | Audit-Grade |
| 102 | Continuous aggregates are not edited in place | Audit-Grade |
| 103 | Retention shortenings require ADR + 30-day notice | Audit-Grade |
| 104 | Soft-deletes never hard-delete in the audit window | Audit-Grade |
| 105 | Reading provenance includes meter signature | Audit-Grade |
| 106 | Submission to GSE/ENEA portals is idempotent and signed | Audit-Grade |
| 107 | Restore-from-backup is rehearsed quarterly | Audit-Grade |
| 108 | Audit-evidence pack is single-command exportable | Audit-Grade |
| 109 | Every OT protocol is a Pack | OT Integration |
| 110 | Protocol Packs declare their wire-format invariants | OT Integration |
| 111 | Edge buffering is the integrator's first-line defence | OT Integration |
| 112 | Time-source on the edge is NTP'd, GPS'd if available | OT Integration |
| 113 | Reading quality codes are explicit and standardised | OT Integration |
| 114 | Modbus polling cadence respects the meter's manual | OT Integration |
| 115 | M-Bus addressing is documented per-installation | OT Integration |
| 116 | SunSpec models declared per device | OT Integration |
| 117 | OCPP version pinning is per-charger | OT Integration |
| 118 | Pulse webhook is HMAC-SHA256 with constant-time compare | OT Integration |
| 119 | Ingest path is bounded; backpressure is observable | OT Integration |
| 120 | Unit-of-measure conversions are explicit and tested | OT Integration |
| 121 | Per-protocol simulators ship for development | OT Integration |
| 122 | Device profiles capture the meter's documentation | OT Integration |
| 123 | Channel mapping is auditor-visible | OT Integration |
| 124 | Latency budget for ingest is documented per-protocol | OT Integration |
| 125 | Outbound DSO clients (E-Distribuzione, Terna, SPD) are circuit-breakered | OT Integration |
| 126 | Network segmentation is OT-aware | OT Integration |
| 127 | Anomaly detection is layered | OT Integration |
| 128 | Real-time data is at-least-once with idempotent consumers | OT Integration |
| 129 | Each regulatory dossier is a Report Pack | Regulatory Pack |
| 130 | Each authoritative factor source is a Factor Pack | Regulatory Pack |
| 131 | Regulatory packs validate against the formal spec | Regulatory Pack |
| 132 | Italian regulatory ground truth is annotated to primary sources | Regulatory Pack |
| 133 | Piano 5.0 thresholds are configurable per attestazione cycle | Regulatory Pack |
| 134 | ESRS E1 disclosures are mapped to the EFRAG taxonomy | Regulatory Pack |
| 135 | Conto Termico submission XML is GSE-spec compliant | Regulatory Pack |
| 136 | TEE (Certificati Bianchi) submissions are batch-aware | Regulatory Pack |
| 137 | D.Lgs. 102/2014 audit dossier carries EGE countersignature | Regulatory Pack |
| 138 | Regulatory pack updates are queued for annual review | Regulatory Pack |
| 139 | Regulatory thresholds are propagated, not duplicated | Regulatory Pack |
| 140 | Per-tenant regulatory profile is explicit | Regulatory Pack |
| 141 | Reports are deterministic in serialisation | Regulatory Pack |
| 142 | Reports declare their input data window inclusively | Regulatory Pack |
| 143 | Scope 3 is opt-in per category and ADR'd | Regulatory Pack |
| 144 | Reports are signed at finalisation | Regulatory Pack |
| 145 | The audit-evidence pack export is regulatory-pack-aware | Regulatory Pack |
| 146 | Regulatory PDF cover-letters are template-driven | Regulatory Pack |
| 147 | Notifications to GSE/ENEA are tracked | Regulatory Pack |
| 148 | Regulatory packs declare their EGE / auditor dependency | Regulatory Pack |
| 149 | Discovery is a deliverable, not an assumption | Engagement Lifecycle |
| 150 | The engagement fork is created at Phase 1, not earlier | Engagement Lifecycle |
| 151 | Pack assembly is bounded at Phase 2 | Engagement Lifecycle |
| 152 | Customisation Sprint is bounded at Phase 3 | Engagement Lifecycle |
| 153 | Hardening + soak is bounded at Phase 4 | Engagement Lifecycle |
| 154 | Handover is operator-readiness, not paperwork | Engagement Lifecycle |
| 155 | Engagement runbooks are engagement-specific | Engagement Lifecycle |
| 156 | Engagement health is monitored monthly | Engagement Lifecycle |
| 157 | Engagement renewal is intentional, not automatic | Engagement Lifecycle |
| 158 | Termination produces an exit pack | Engagement Lifecycle |
| 159 | Engagement-specific code never lands upstream without generalisation | Engagement Lifecycle |
| 160 | Engagement post-mortems contribute to the doctrine | Engagement Lifecycle |
| 161 | On-call rotation in T2/T3 is documented and rotated | Engagement Lifecycle |
| 162 | Engagement reviews are tri-annual | Engagement Lifecycle |
| 163 | Pricing is transparent at engagement start | Engagement Lifecycle |
| 164 | Customer data ownership is contractual and technical | Engagement Lifecycle |
| 165 | Engagement testing fixtures are synthetic by default | Engagement Lifecycle |
| 166 | Engagement support tier upgrades are explicit | Engagement Lifecycle |
| 167 | Crisis communication is rehearsed | Engagement Lifecycle |
| 168 | Lessons from the engagement portfolio feed back to roadmap | Engagement Lifecycle |
| 169 | The audit log is signed in chained-HMAC | Cryptographic |
| 170 | JWT tokens carry `kid`, are HS256, are validation-pinned | Cryptographic |
| 171 | Cosign signing is keyless via Sigstore Fulcio | Cryptographic |
| 172 | Field-level encryption uses per-tenant DEK + KMS-wrapped MEK | Cryptographic |
| 173 | Meter signatures use HMAC-SHA256 with per-meter keys | Cryptographic |
| 174 | TLS is 1.3-only on customer-facing edges | Cryptographic |
| 175 | Certificate rotation is automated; manual rotation is the runbook | Cryptographic |
| 176 | Random number generation uses `crypto/rand` | Cryptographic |
| 177 | Password storage is bcrypt with cost factor 12 | Cryptographic |
| 178 | Webhook signatures use HMAC-SHA256 with constant-time compare | Cryptographic |
| 179 | KMS keys are regional; cross-region transfer requires explicit policy | Cryptographic |
| 180 | Secrets-rotation runbook is annual or post-incident | Cryptographic |
| 181 | TOTP is the second factor for admin accounts | Cryptographic |
| 182 | IP and email-based lockout for repeated failed login | Cryptographic |
| 183 | Session tokens are revocable and have idle + absolute lifetimes | Cryptographic |
| 184 | Crypto-shredding handles GDPR Art. 17 | Cryptographic |
| 185 | Cryptographic primitives use the standard library where possible | Cryptographic |
| 186 | Crypto agility: algorithm choices are versioned | Cryptographic |
| 187 | Quantum-resistant crypto plan is on the radar | Cryptographic |
| 188 | Cryptography failures are Sev-1 by default | Cryptographic |
| 189 | AI/ML outputs are not regulatory artefacts unless explicitly opted in | AI/ML |
| 190 | Every model carries a model card | AI/ML |
| 191 | Training data lineage is captured | AI/ML |
| 192 | Forecasts are reproducible from frozen inputs | AI/ML |
| 193 | Forecast accuracy is monitored continuously | AI/ML |
| 194 | Drift detection is automatic | AI/ML |
| 195 | Anomaly-detection layers each carry a model card | AI/ML |
| 196 | Forecast inputs declare their staleness budget | AI/ML |
| 197 | Model retraining is sprint-gated and ADR'd | AI/ML |
| 198 | Models are versioned and rolled back | AI/ML |
| 199 | Inference is observable | AI/ML |
| 200 | AI explainability is a deliverable | AI/ML |
| 201 | Model fairness review is part of release | AI/ML |
| 202 | AI/ML libraries are pinned and digest-locked | AI/ML |
| 203 | AI privacy: training data PII handling is documented | AI/ML |
| 204 | AI bias evaluation is documented | AI/ML |
| 205 | Consumption forecaster ships only after passing the evaluation budget | AI/ML |
| 206 | AI integration with the flexibility market (MSD) is opt-in per-tenant | AI/ML |
| 207 | AI/ML failures degrade gracefully | AI/ML |
| 208 | AI/ML termination criterion is named per-model | AI/ML |
| 209 | Adding a doctrine rule requires evidence and a Tradeoff Stanza | Meta |
| 210 | Removing a doctrine rule requires a recorded supersession | Meta |

---

## Cross-rule traceability

The doctrine is interdependent. The following cross-rule chains are the most load-bearing for reviewers:

**Audit-grade reproducibility chain:** Rule 33 → Rule 89 → Rule 90 → Rule 91 → Rule 95 → Rule 96 → Rule 102 → Rule 141 → Rule 144. The chain prescribes that schema is the system, output is reproducible, factors are temporal, builders are pure, every report has provenance, source data is immutable, CAGGs are not edited in place, serialisation is deterministic, and finalised reports are signed. A break anywhere breaks the regulatory defence.

**Modular template chain:** Rule 69 → Rule 70 → Rule 71 → Rule 72 → Rule 73 → Rule 78 → Rule 79 → Rule 87 → Rule 88. The chain prescribes that Core/Pack is the distinction, Packs manifest themselves, contracts are versioned, registration is explicit, boot is locked, Core changes are merge-friendly, forks sync quarterly, Pack acceptance is objective, and the Italian Pack is the reference. A break anywhere breaks the engagement model.

**Defence-in-depth tenant isolation:** Rule 3 → Rule 12 → Rule 39 → Rule 169. UUIDv4 tenant IDs, RLS policies, structural security, signed audit log. A bug in any one layer is caught by the next.

**Regulator-grade quality:** Rule 25 → Rule 45 → Rule 65 → Rule 95 → Rule 108 → Rule 145. The chain prescribes that the threshold is regulator-grade across Platform, Backend, DevSecOps, that every report carries provenance, that the audit-evidence pack is exportable, and that the export is regulatory-pack-aware. A break anywhere breaks the audit defence.

**Resilience chain:** Rule 36 → Rule 41 → Rule 42 → Rule 76 → Rule 119 → Rule 125 → Rule 207. Failure is normal, concurrency is bounded, resources have lifecycles, Pack failures are isolated, ingest is bounded, DSO clients are breakered, AI/ML degrades gracefully. A break anywhere produces compounding outages.

---

## Conformance: how the doctrine is enforced

A subset of these rules is mechanically enforced via the conformance suite. The current state is partial; full enforcement is the Phase E + Phase F deliverable. The conformance suite is at `backend/tests/conformance/` and `backend/tests/static/` and `backend/tests/security/` and `backend/tests/property/` and `backend/tests/migrations/`. Each rule's "How to apply" section names the relevant test where one exists; otherwise, the test is on the Phase E/F backlog.

The remaining (non-mechanical) rules are enforced at code review by named reviewers per CODEOWNERS. The doctrine grants the named reviewers refusal authority (Rules 26 / 46 / 66 / 88). Quarterly office hours review the per-rule enforcement state.

---

## Interaction with `CLAUDE.md`

The cross-portfolio `CLAUDE.md` file at `~/.claude/projects/-media-alexcupsa-SSD-PROGETTI-00-PROGETTI-attivi-macena-greenmetrics/memory/MEMORY.md` carries portfolio-level invariants (the universal-invariants of Section 0). This doctrine is downstream: the universal invariants are repeated here for self-containment but are sourced from CLAUDE.md. A change to a universal invariant requires both a portfolio-level ADR and a per-project re-adoption.

---

## Cross-references

- **Charter:** `docs/MODULAR-TEMPLATE-CHARTER.md`. Charter §3 carries the Core / Pack distinction; Charter §13 carries the explicit forbiddings.
- **Plan:** `docs/PLAN.md`. Per-Phase, per-Sprint application of the doctrine.
- **Quality bar:** `docs/QUALITY-BAR.md`. Eleven non-negotiables.
- **Rejected patterns:** `docs/adr/REJECTED.md`. Per-rule citations for each rejection.
- **ADRs:** `docs/adr/`. The lineage of every non-trivial decision.
- **Risk register:** `docs/RISK-REGISTER.md`. Per-RISK rule-citation.
- **Threat model:** `docs/THREAT-MODEL.md`. STRIDE per attack surface.
- **Runbooks:** `docs/runbooks/`. Per-failure-mode runbook.

---

## Tradeoff Stanza

- **Solves:** the implicit, scattered doctrine references across ADRs and runbooks; the loss of doctrine continuity when the off-machine plan file is unavailable; the absence of explicit modular-template / audit-grade / OT-integration / regulatory-pack / engagement-lifecycle / cryptographic / AI-ML rule groups; the impossibility of mechanically enforcing rules that aren't written down.
- **Optimises for:** reviewability (every rule has a body, why, how-to-apply, cross-refs), enforceability (mechanical-rule subset has named conformance tests), evolvability (Rule 209/210 process), defensibility (regulator can read the doctrine and verify against the codebase), portability (the doctrine survives team growth and personnel change).
- **Sacrifices:** the brevity of "we have a 60-rule doctrine"; the option to silently amend rules; ~3000 lines of documentation surface that must be reviewed and maintained; the velocity of rule-free development.
- **Residual risks:** the doctrine drifts from reality (mitigated by the conformance suite + quarterly office hours); reviewers ignore the doctrine (mitigated by the CODEOWNERS routing + PR-template rule citation); the doctrine becomes its own bottleneck (mitigated by the supersession process + the rule that rules without evidence are rejected); a critical rule is missed (mitigated by the engagement-portfolio feedback loop).

---

*This doctrine governs every PR, ADR, runbook, and engagement decision in the repository at and after Sprint S5. PRs that violate a binding rule without an override-allowed ADR are blocked. The doctrine is reviewed at six-month intervals; the next review is 2026-10-30.*
