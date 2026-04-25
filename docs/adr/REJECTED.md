# Rejected Patterns Log

**Doctrine refs:** Rule 26 (Platform rejection authority), Rule 46 (Backend rejection authority), Rule 66 (DevSecOps rejection authority).
**Maintainer:** `@greenmetrics/platform-team`, `@greenmetrics/secops`.

This file lists patterns explicitly rejected for GreenMetrics. Each rejection cites the doctrine rule, names the alternative chosen, the residual risk, and a review date. PRs proposing a rejected pattern are blocked unless accompanied by an explicit unrejection ADR.

The `make audit-rejections` script scans new ADRs and PRs for keywords from this list and surfaces them for triage.

## Rejected — Architecture / Platform

| ID | Pattern | Rule | Reason | Alternative | Residual | Review |
|---|---|---|---|---|---|---|
| REJ-01 | Service mesh (Istio / Linkerd) for two services | 26 | Mesh control-plane cost > leverage at this scale; complexity ≠ value | NetworkPolicy default-deny + cert-manager-issued mTLS where needed | Re-evaluate at ≥ 6 services or zero-trust east-west requirement | 2026-10-25 |
| REJ-02 | Helm + Kustomize + Argo CD + Flux simultaneously | 26 | Tool sprawl; pick one renderer + one GitOps engine | Kustomize for rendering + Argo CD for GitOps | None of consequence | 2026-10-25 |
| REJ-03 | OpenTelemetry collector as ingest bus | 26, 46 | Hidden coupling; OTel is for telemetry only, not data plane | Fiber → TimescaleDB CopyFrom on ingest; OTel exclusively for traces/metrics/logs | None | 2026-10-25 |
| REJ-04 | Per-service config-management framework (Consul/etcd/Vault for app config) | 13, 26 | Abstraction cost > leverage | 12-factor env via `internal/config/config.go`; secrets via ESO | Adds ESO dep — accepted | 2026-10-25 |
| REJ-05 | Microservice split of the backend | 46 (under-engineering inversion) | Single-team monolith stays until measured bottleneck | Per-process boundary refactor only, no service split | None until traffic warrants | 2027-01-25 |
| REJ-06 | Bespoke alert routing | 26 | Tool sprawl | Alertmanager → Slack + PagerDuty + email | None | 2026-10-25 |
| REJ-07 | Bespoke schema migration framework | 13 | NIH; cost > leverage | `pressly/goose` v3 (Go-coded support for Timescale CAGG dance) | Goose is single-author maintained — review at 2027 | 2027-04-25 |
| REJ-08 | Terraform monorepo with one root state | 16, 26 | Single state file is a SPOF and scaling axis violation | Env-scoped roots `terraform/envs/{dev,staging,prod}/` + shared modules | None | 2026-10-25 |
| REJ-09 | Multi-region active-active as Phase 1 | 25, 65 | Over-scoping; Italian residency satisfied by single region | Active-passive on eu-south-1 v1; revisit when usage demands | RPO 1h / RTO 4h tradeoff documented | 2027-04-25 |
| REJ-10 | OPA-everywhere (request-time PEP in Fiber) | 26 | Cost > leverage; admission + IaC layer is enough | Conftest at IaC + K8s manifest + Kyverno admission only | Authz logic in middleware code — accepted | 2026-10-25 |

## Rejected — Backend

| ID | Pattern | Rule | Reason | Alternative | Residual | Review |
|---|---|---|---|---|---|---|
| REJ-11 | God report generator (`internal/services/report_generator.go` 525 LoC, 7 builders) | 46 | God service; tight coupling | Decompose into `internal/domain/reporting/<dossier>.go`; orchestrator ≤ 80 LoC | None post-refactor | 2026-10-25 |
| REJ-12 | Implicit tenant scoping (developer discipline only) | 39, 46 | Hidden coupling; one missed `WHERE` is a regulator-grade leak | Postgres RLS + RBAC + app-level filter (defence in depth) | RLS bypass via SECURITY DEFINER fn — code-review gate | 2027-04-25 |
| REJ-13 | Float for money | 46 (CLAUDE.md invariant) | FP rounding errors in money are a regulator-grade integrity bug | `(amount_cents int64, currency_iso4217 string)` value object in `internal/domain/money/` | None | 2026-10-25 |
| REJ-14 | swag-only contract | 14, 34 | Implicit contract; reverse-engineered from comments | Hand-written `api/openapi/v1.yaml` + oapi-codegen | swag retired completely | 2026-10-25 |
| REJ-15 | `panic()` on user input | 46 | Crash as DoS amplifier | RFC 7807 error helper everywhere; CI grep gate | None | 2026-10-25 |
| REJ-16 | Untyped `interface{}` / `any` returns from repository | 46 | Type leakage; loss of compile-time safety | Per-aggregate typed methods | None | 2026-10-25 |
| REJ-17 | `placeholder-tenant` literal in code | 32, 46 | Naming reflects domain (Rule 32); placeholder is a footgun | Explicit `INGESTOR_TENANT_ID` env; boot fails if absent | None post-fix | 2026-10-25 |
| REJ-18 | Logged-only sentinel JWT secret | 19, 39 | Audit-only ≠ enforcement | Hard boot refusal in `config.Load` for production env | None | 2026-10-25 |
| REJ-19 | OTel sample ratio 1.0 in production | 13, 18, 40 | Cost > insight at 5k req/s | Default 0.1 in production with explicit env override | Loss of trace fidelity on rare paths — Loki sampling supplements | 2026-10-25 |
| REJ-20 | Plaintext compare on Pulse webhook | 39 | Timing oracle | HMAC-SHA256 over body with constant-time compare | None post-fix | 2026-10-25 |
| REJ-21 | Synchronous report generation | 37 | Head-of-line latency violation | Async via Asynq + Redis worker; `POST /v1/reports` returns 202 | Redis = new failure domain — accepted, justified in ADR-014 | 2026-10-25 |
| REJ-22 | Unbounded Fiber body size | 39, 42 | Payload-flood DoS vector | `BodyLimit: 4 MB` global + 16 MB per-route override on ingest | None | 2026-10-25 |
| REJ-23 | Unbounded goroutine spawn in ingestor | 41, 42 | Resource leak under load | `errgroup` + `panjf2000/ants` worker pool | ants is a dependency — accepted | 2026-10-25 |

## Rejected — DevSecOps

| ID | Pattern | Rule | Reason | Alternative | Residual | Review |
|---|---|---|---|---|---|---|
| REJ-24 | Manual approval as the only CD gate | 56, 66 | Bottleneck masquerading as control | Layered automated gates (sign / SBOM / SLSA / Trivy / DAST / smoke / canary with Argo Rollouts) + per-release human signoff | Bad analysis template lets regression through — analysis template is itself versioned + tested | 2026-10-25 |
| REJ-25 | Snyk on top of Trivy + govulncheck + osv-scanner + CodeQL | 26, 66 | Tool sprawl; marginal signal | Trivy + govulncheck + osv-scanner + CodeQL only | None | 2027-04-25 |
| REJ-26 | Aqua / Sysdig on top of Falco / Tetragon | 26, 66 | Tool sprawl; CNCF stack sufficient | Falco DaemonSet (or Tetragon if Falco fails on Bottlerocket) | None | 2027-04-25 |
| REJ-27 | Cosign with custom KMS keys | 66 | Key management burden ≠ improved security | Keyless OIDC via Sigstore Fulcio | Trust chain rooted in Sigstore CT log — annual review | 2027-04-25 |
| REJ-28 | DIY KMS | 66 | NIH; reduces security | AWS KMS in `terraform/modules/s3/main.tf:7-12` | None | 2027-04-25 |
| REJ-29 | Unpinned GitHub Actions tags | 53 | `tj-actions/changed-files` 2025 incident pattern | SHA-pin every Action with `# vX.Y.Z` comment; Dependabot weekly | Dependabot churn — accepted | 2026-10-25 |
| REJ-30 | Vault on Phase 1 | 13, 67 | Control plane cost > leverage at this scale; AWS already in stack | AWS Secrets Manager + ESO; switch path is replacing `SecretStore` ref | Vendor lock-in — abstracted by ESO | 2027-04-25 |
| REJ-31 | SPIRE on Phase 1 | 13 | Federation complexity not yet needed | cert-manager + trust-manager for in-cluster PKI; SPIRE if multi-cluster federation appears | No multi-cluster identity federation today | 2027-04-25 |
| REJ-32 | GraphQL gateway in front of REST | 26, 46 | Adds layer without leverage; clients are simple | REST + RFC 7807 + OpenAPI 3.1 stays | None | 2027-04-25 |
| REJ-33 | Kafka for telemetry | 26 | OTel + Prometheus + Loki + Tempo is the canonical telemetry path | OTLP → collector → backends | None | 2027-04-25 |
| REJ-34 | Custom Go web framework | 26 | NIH | Fiber stays | None | 2027-04-25 |
| REJ-35 | ORM (gorm / bun) on Timescale | 13, 46 | pgx parameterised queries scale better for Timescale-heavy workload | pgx + per-aggregate repositories | Manual SQL — accepted; revisit if migration toolchain pain | 2027-04-25 |

## Process

- New rejection: open PR adding row to this table + ADR explaining the reasoning.
- Unrejection: PR adds row to "Unrejected" log (below) + new ADR explaining what changed.
- Quarterly review: walk every rejection past its review date.

## Unrejected log

(empty — populates if a rejection is reversed with rationale)
