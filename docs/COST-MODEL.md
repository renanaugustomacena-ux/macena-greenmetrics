# GreenMetrics Cost Model

**Doctrine refs:** Rule 22 (cost awareness), Rule 16 (scalability dimensions).
**Owners:** `@greenmetrics/platform-team`, finance.
**Review cadence:** monthly delta vs AWS bill; quarterly full re-fit.

## 1. Per-tenant unit economics

Per-tenant monthly cost broken into:

| Cost centre | Variable | Driver |
|---|---|---|
| Ingest storage | rows_per_day_per_tenant · B · D / 10 (compressed) · €/GB-month | Capacity model; ~€0.115/GB-month gp3 |
| CAGG storage | ~25% of raw storage | hypertable + aggregation overhead |
| RDS compute | replica-hours · €/h / tenants_share | profile-dependent; small tenants share infrastructure cost |
| Backend compute (HPA) | rps_share · €/h | based on `gm_ingest_readings_total` and `gm_http_requests_total` |
| Frontend serving | static asset size · CDN egress €/GB | CloudFront pricing |
| Worker (Asynq) | report_jobs/month · avg_minutes · €/h | pricing per job class |
| Bandwidth | bytes_per_tenant_per_month · €/GB | ingress is free; egress on report download |
| Grafana | flat-cost amortised | per-replica overhead |
| Secrets Manager | secrets_count · €/secret-month | flat per stored secret |
| KMS | requests · €/10k requests | cheap unless audit-heavy |
| WAF | requests · €/M requests | cheap unless attack |

Worked illustrative numbers for a Medium-profile tenant (50 meters @ 4 readings/min):

```
ingest storage: 9.2 GB raw/month → ~0.92 GB compressed → ~€0.11
CAGG storage:    ~€0.03
RDS compute:    €60/month / 50 tenants = €1.20
backend compute: ~€0.50 amortised
frontend:       ~€0.05
worker:         ~€0.02
bandwidth:      ~€0.20
secrets/KMS/WAF: ~€0.10 amortised
TOTAL:          ~€2.20 / Medium tenant / month
```

Stretch tenant (2000 meters @ 12 r/min) costs ~€350/month dominated by RDS.

Customer pricing target ≥ 4× cost for healthy SaaS gross margin → ~€10/Medium, ~€1500/Stretch (matches MODUS_OPERANDI commercial table).

## 2. Infra cost breakdown (env totals, planning anchor)

| Env | Monthly target | Cap (AWS Budgets) |
|---|---|---|
| dev | €100 | €200 |
| staging | €500 | €750 |
| production (year 1) | €2 000 | €5 000 |

Cost-allocation tags enforced on every resource (`Project`, `Environment`, `Owner`, `CostCenter`, `DataResidency`) — `policies/conftest/terraform/tags.rego`.

## 3. AWS Budgets

`terraform/modules/cost/main.tf` (S2 to ship):

- Per-env `aws_budgets_budget` with notification thresholds 50%, 80%, 100%, 120%.
- Notifications → Slack `#greenmetrics-ops` + secops email.
- Budget actions disabled (auto-stop on spend would harm availability); alert-only.

## 4. AWS CUR pipeline

- Cost & Usage Report enabled to S3 bucket `greenmetrics-aws-cur` (KMS encrypted).
- Daily delivery; CSV.
- Pulled into Grafana via AWS Cost Datasource plugin (ships with `kube-prometheus-stack` overlay).
- Dashboard `cost-overview.json` (S4 deliverable): per-cost-centre, per-tenant, per-resource, daily trend.

## 5. Efficiency tradeoffs

| Lever | Saves | Costs |
|---|---|---|
| Timescale compression on ≥7d chunks (10×) | 80% storage | 3× decompression query latency on cold data |
| CAGG refresh policy 5min vs 1min | 80% refresh CPU | 4 min CAGG lag (ok per SLO) |
| HPA scale-down stabilization 300s | 30% over-provision avoidance | slight latency spikes on scale-up |
| Asynq + Redis vs in-process | OLTP pool isolation; SLO protection | Redis = €30/month + 1 failure domain |
| Distroless nonroot vs slim | larger image (~300 MB vs ~80 MB) | image pull bandwidth + GHCR storage |
| Multi-AZ RDS | DR + zero-downtime failover | 2× DB cost |
| Object Lock on audit bucket | compliance + tamper resistance | cannot delete even legitimate stale entries |

Each tradeoff documented in the relevant ADR.

## 6. Waste detection

`scripts/ops/cost-audit.sh` (S2 to ship), monthly cron:

- List unused EBS volumes (no attachment > 7d).
- List idle RDS replicas (CPU < 5% sustained 30d).
- List orphaned secrets (Secrets Manager `LastAccessedDate` > 90d).
- List Grafana datasources with zero queries 30d.
- List unused IAM roles (CloudTrail no `AssumeRole` event 90d).
- List underutilised K8s nodes (CPU < 20% sustained 7d).

Output → GitHub issue `monthly-cost-audit-YYYY-MM`.

## 7. Cost regression in CI

`infracost` integrated in `.github/workflows/terraform-plan.yml` (S2 to ship):

- PR comment shows monthly cost diff per resource.
- Threshold: any +€500/month change requires `cost-approved` label + ADR.

## 8. Anti-patterns rejected (Rule 66)

- "Just spin up a bigger RDS — €€€ are not the user's problem." Rejected; cost is a structural concern (Rule 22).
- "Disable Cost Explorer, it costs money." Rejected; Cost Explorer is a few cents per month and saves thousands.
- Reserved Instances v1 — rejected until usage stabilises 6 months; commit too early and lose flexibility.

## 9. Doctrine cross-references

- Rule 22 — primary owner.
- Rule 16 — capacity model is the upstream input.
- Rule 27 / 47 / 67 — every tradeoff above carries a stanza.
- ADR-014 — Asynq + Redis cost / failure domain rationale.
- ADR-009 — circuit breaker cost / behaviour rationale.
