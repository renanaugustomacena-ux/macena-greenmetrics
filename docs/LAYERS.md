# GreenMetrics Layers

**Doctrine refs:** Rule 10 (multi-layer system), Rule 14 (contracts at crossings).
**Source of truth:** `docs/layers.yaml` — diagram below regenerated via `make layers-doc`.

GreenMetrics exists in five layers. Optimising one layer in isolation is invalid (Rule 10). Each layer crossing is a contract (Rule 14).

## 1. Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│  Layer 5 — Operators                                             │
│  Runbooks, ADRs, on-call rota, postmortems, office hours.        │
│  Owners: @greenmetrics/sre, @greenmetrics/secops, @greenmetrics/ │
│  platform-team.                                                  │
└─────────▲───────────────────────────────────────────────────────┘
          │  Crossing: incident escalation, ADR review, runbook update
          │
┌─────────┴───────────────────────────────────────────────────────┐
│  Layer 4 — Tooling                                               │
│  CI/CD, devcontainer, pre-commit, conftest, Cosign, kubeconform, │
│  Renovate/Dependabot, k6, Playwright, Chaos Mesh, nuclei.        │
│  Owners: @greenmetrics/platform-team, @greenmetrics/secops.      │
└─────────▲───────────────────────────────────────────────────────┘
          │  Crossing: pipeline gates, policy enforcement, artefact signing
          │
┌─────────┴───────────────────────────────────────────────────────┐
│  Layer 3 — Application                                           │
│  Backend (Fiber, pgx, OTel, JWT), frontend (SvelteKit), worker   │
│  (Asynq), simulator. Domain logic in internal/domain/.            │
│  Owners: @greenmetrics/app-team.                                 │
└─────────▲───────────────────────────────────────────────────────┘
          │  Crossing: K8s envFrom (config), ServiceAccount→IRSA (identity),
          │            Secret mount via ESO (secrets), OTLP→collector (traces),
          │            Prometheus scrape (metrics), Loki log shipping (logs)
          │
┌─────────┴───────────────────────────────────────────────────────┐
│  Layer 2 — Platform services                                     │
│  K8s, Argo CD, ESO, OTel collector, Prometheus, Alertmanager,    │
│  Loki, Tempo, Falco, cert-manager, Kyverno admission, ingress.   │
│  Owners: @greenmetrics/platform-team, @greenmetrics/secops,      │
│  @greenmetrics/sre.                                              │
└─────────▲───────────────────────────────────────────────────────┘
          │  Crossing: Terraform output → Helm values, IRSA trust policy,
          │            Secrets Manager ARN → ESO ClusterSecretStore,
          │            VPC subnet IDs → cluster networking, KMS key ARN
          │
┌─────────┴───────────────────────────────────────────────────────┐
│  Layer 1 — Infrastructure                                        │
│  AWS (eu-south-1 Milan): VPC, EKS, RDS PG16 (TimescaleDB),       │
│  ElastiCache Redis, S3, KMS, Secrets Manager, CloudFront + WAF,  │
│  CloudTrail, IAM, Route 53. Aruba Cloud as sovereign alt.        │
│  Owners: @greenmetrics/platform-team, @greenmetrics/secops.      │
└─────────────────────────────────────────────────────────────────┘
```

## 2. Per-layer ownership and gaps

| Layer | Primary owner | Today | Gap closed by |
|---|---|---|---|
| 1 — Infra | platform | Terraform skeleton, state backend commented | Sprint S2 (Rule 63) |
| 2 — Platform | platform / secops / sre | K8s manifests production-grade, monitoring/ empty | Sprints S2–S4 (Rules 18, 50, 54, 63) |
| 3 — App | app | Fiber backend + SvelteKit frontend running, no DDD split | Sprints S2–S3 (Rules 30, 32, 39) |
| 4 — Tooling | platform / secops | CI exists, no policy gate, no Cosign | Sprints S1–S3 (Rules 23, 52, 53, 54) |
| 5 — Operators | sre / secops | RUNBOOK.md exists, no per-failure runbooks, no IR doc | Sprint S4–S5 (Rules 20, 60) |

## 3. Layer-crossing contracts (Rule 14)

| From | To | Contract | Source of truth |
|---|---|---|---|
| App → Platform | K8s | `envFrom` (ConfigMap + Secret) | `k8s/configmap.yaml`, `k8s/secret.yaml` |
| App → Platform | OTel collector | OTLP gRPC | `OTEL_EXPORTER_OTLP_ENDPOINT` env |
| App → Platform | Prometheus | `/metrics` exposition | `internal/metrics/metrics.go` |
| App → Platform | Loki | stdout JSON via Promtail | zap encoder config |
| Platform → Infra | ESO | `ClusterSecretStore` | `gitops/base/external-secrets/clustersecretstore.yaml` |
| Platform → Infra | IRSA | ServiceAccount annotation | `k8s/service.yaml`, `terraform/modules/iam-irsa/` |
| Platform → Infra | DB | pgx connection string | `DATABASE_URL` (pulled from Secrets Manager via ESO) |
| Tooling → App | OpenAPI codegen | `oapi-codegen` types | `api/openapi/v1.yaml` |
| Tooling → Platform | Argo CD sync | `Application` manifest | `gitops/argocd/applications/` |
| Tooling → Tooling | conftest policy gate | Rego rules | `policies/conftest/` |
| Tooling → Platform | Kyverno admission | ClusterPolicy | `policies/kyverno/` |
| Operators → All | Runbook YAML front-matter | severity / mttd / mttr / owner | `docs/runbooks/*.md` |
| Operators → All | ADR Tradeoff stanza | Solves / Optimises / Sacrifices / Residual | `docs/adr/*.md` |

## 4. Cross-layer impact reasoning (Rule 10)

A change at any layer must be reasoned through every other layer it touches:

- **Add a new ingestor protocol (App layer):** requires extension-point interface (Rule 12), per-meter capacity (Rule 16, Layer 1 sizing), Prometheus metric (Rule 18, Layer 2 cardinality), runbook (Rule 20, Layer 5), policy gate (Rule 54, Layer 4), threat model entry (Rule 55, secops review).
- **Bump TimescaleDB version (Infra layer):** requires CAGG migration plan (Rule 33, Layer 3 schema), Prometheus exporter compatibility (Rule 18, Layer 2), runbook update for failure modes (Rule 20, Layer 5), capacity re-baseline (Rule 16, Layer 1 sizing).
- **Add a Grafana alert rule (Platform layer):** requires runbook link (Rule 18, Layer 5), Alertmanager routing (Rule 58, Layer 2), SLI catalog entry (Rule 18, docs/SLI-CATALOG.md), false-positive review cadence (Rule 58, Layer 5).

## 5. Layer reviews

Each layer is reviewed quarterly during platform office hours:

- **Q1 + Q3:** Layers 1, 4 (infra + tooling).
- **Q2 + Q4:** Layers 2, 3, 5 (platform, app, operators).
