# Observability — Operator Reference

**Doctrine refs:** Rule 18, Rule 40, Rule 58.
**Owner:** `@greenmetrics/sre`.

## 1. Stack

| Concern | Component | Where |
|---|---|---|
| Metrics | Prometheus (kube-prometheus-stack) | `monitoring/prometheus/values.yaml` |
| Logs | Loki + Promtail | `monitoring/loki/values.yaml` |
| Traces | Tempo | `monitoring/otel-collector/config.yaml` (exporter) |
| Telemetry router | OpenTelemetry Collector (DaemonSet) | `monitoring/otel-collector/config.yaml` + `gitops/base/monitoring/otel-collector.yaml` |
| Dashboards | Grafana (separate from kps) | `grafana/provisioning/`, `grafana/alerting/` |
| Alerting | Alertmanager | `monitoring/prometheus/values.yaml` (config block) |
| Runtime detection | Falco DaemonSet | `gitops/base/falco/`, `policies/falco/greenmetrics-rules.yaml` |
| Audit log shipping | Loki via Promtail | K8s audit policy `gitops/base/eks-audit-policy.yaml` (S5) |

## 2. What each alert means + on-call action

| Alert | Severity | Page? | Action | Runbook |
|---|---|---|---|---|
| APIErrorRateHigh | critical | yes | Investigate 5xx pattern; likely DB or breaker | `db-outage.md` |
| APIDown | critical | yes | All replicas unreachable; check K8s + RDS | `db-outage.md` |
| AbsentMetricsDeadMan | critical | yes | No metrics from backend; Prometheus / SM issue | `db-outage.md` |
| P99IngestBudget | warning | no | Latency budget breach; investigate pgx pool, queue | `capacity-spike.md` |
| P99LoginBudget | warning | no | Login slow; investigate bcrypt + pool | `capacity-spike.md` |
| IngestQueueSaturated | warning | no | Drops imminent; scale or throttle | `pulse-webhook-flood.md` |
| IngestDropped | critical | yes | Readings discarded; investigate source rate + scale | `pulse-webhook-flood.md` |
| BreakerOpen | warning | no | Specific upstream failing; verify external service | `db-outage.md` |
| ExternalAPIFallback | warning | no | Cached factors served; confirm upstream health | `db-outage.md` |
| ModbusIngestStalled | warning | no | Modbus host or NetworkPolicy issue | `ingestor-crash-loop.md` |
| DBPrimaryDown | critical | yes | RDS down; trigger Multi-AZ failover | `db-outage.md` |
| PgxPoolAcquireSlow | warning | no | Pool saturated; consider pgbouncer or scale | `capacity-spike.md` |
| CAGGRefreshLag | warning | no | Aggregated dashboards stale; manual refresh | `db-outage.md` |
| TimescaleStorageHigh | warning | no | Storage > shared_buffers ratio; check retention | `cost-audit.md` |
| DBReplicaLag | warning | no | Replica behind primary; reads may be stale | `db-outage.md` |
| LoginFailureSpike | warning | no | Possible brute-force; verify lockout + WAF | `tenant-data-leak.md` |
| JWTVerifySlow | warning | no | JWT verify slow; check kid map size | `jwt-secret-rotation.md` |
| JWTSecretAgeExceeded | critical | yes | Quarterly rotation skipped; investigate workflow | `jwt-secret-rotation.md` |
| ESONotSynced | warning | no | Secret stuck; check IRSA + Secrets Manager | `secret-rotation.md` |
| CertificateExpirySoon | warning | no | < 14d to expiry; cert-manager renewal failing | `cert-rotation.md` |
| KyvernoAdmissionSlow | warning | no | Webhook slow; check Kyverno HA + resources | `grafana-down.md` (similar pattern) |
| FalcoCriticalRule | critical | yes | Runtime intrusion suspected | `tenant-data-leak.md` |
| Disk usage > 80/90% | warning/critical | warn/yes | EBS extend or cleanup | `capacity-spike.md` |
| Pod restart loop | warning | no | Investigate panic / OOM | `ingestor-crash-loop.md` |
| NetworkPolicy denied | critical | yes | Possible exfil attempt | `tenant-data-leak.md` |
| SLOErrorBudgetBurnFast | critical | yes | SLO burn 14.4× over 1h | `db-outage.md` |
| SLOErrorBudgetBurnSlow | warning | no | Sustained slow burn | `db-outage.md` |

## 3. Loki query examples

```logql
# Errors for a specific request id
{service="greenmetrics-backend"} |= "request_id" | json | request_id="abc-123"

# Per-tenant filter
{service="greenmetrics-backend"} | json | tenant_id="<uuid>"

# Pulse webhook signature failures
{service="greenmetrics-backend"} | json | path="/api/v1/pulse/ingest" | code="PULSE_SIGNATURE_INVALID"

# Falco events
{app="falco"} | json | priority=~"Critical|Error"

# Argo CD audit
{app="argocd-server"} | json | level="info" | actionType="resource-action"
```

## 4. Tempo trace lookup

Click any `trace_id` field in a Loki panel → Tempo opens with the trace tree. Span attributes carry `tenant.id`, `db.statement`, `http.target`, etc.

## 5. Adding a new alert

1. Add a rule to `monitoring/prometheus/rules/<area>.rules.yaml`.
2. Annotate `runbook_url` to a `docs/runbooks/<file>.md`.
3. Add the runbook if it doesn't exist.
4. Test with `promtool test rules`.
5. PR + Argo CD sync.

## 6. Adding a new metric

1. Define in `internal/metrics/metrics.go` with cardinality budget comment.
2. Update `docs/SLI-CATALOG.md` if it gates an SLO.
3. Update relevant Grafana dashboard.
4. Add to RED/USE per-service table if applicable.

## 7. Anti-patterns rejected

- Alerts without `runbook_url`.
- Alerts firing > 3×/week with no action — downgrade or remove.
- High-cardinality labels on histograms (tenant_id, meter_id) — cost > value.
- Logs without `request_id` / `trace_id` / `span_id` / `tenant_id` — wrap via `obs.Logger(ctx)`.
- Custom alerting bus instead of Alertmanager (REJ-06).
