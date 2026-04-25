# 0006 — Observability stack: OTel + Prometheus + Loki + Tempo

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 18, 40, 58
**Review date:** 2027-04-25

## Context

Backend already emits OpenTelemetry traces via OTLP gRPC and exposes Prometheus metrics on `/api/internal/metrics`. The `monitoring/` directory was empty (no Prometheus, Alertmanager, OTel collector YAML); 0 Grafana alert rules; SLOs documented but not encoded. Plan calls for a full observability stack with log/trace correlation and runbook-linked alerts.

## Decision

Adopt the canonical CNCF observability stack:

- **Metrics:** Prometheus via `kube-prometheus-stack` Helm chart.
- **Logs:** Loki + Promtail (Promtail extracts `level`, `request_id`, `trace_id`, `span_id`, `tenant_id`, `service` as labels).
- **Traces:** Tempo (OTLP gRPC ingestion).
- **Telemetry routing:** OpenTelemetry Collector DaemonSet (`monitoring/otel-collector/config.yaml`) — receives OTLP from backend, applies tail sampling 10% (production), exports to Tempo + Prometheus remote_write + Loki.
- **Alerting:** Alertmanager (built into kps) → Slack `#greenmetrics-ops` (warning), PagerDuty (critical), email (digest). Inhibition: DB-down silences downstream.
- **Dashboards:** Grafana (separate from kps) provisioned from `grafana/provisioning/`; Grafana unified alerting in `grafana/alerting/`.
- **Trace ↔ log correlation:** `internal/observability/zap_trace.go` injects `trace_id`, `span_id`, `request_id`, `tenant_id` into every log line via `obs.Logger(ctx)`.
- **OTel sample ratio:** default 0.1 in production (`config.Load`); 1.0 elsewhere.
- **Custom metrics:** `internal/metrics/metrics.go` — `gm_*` counters/gauges/histograms with explicit cardinality budget.
- **Alert rules:** `monitoring/prometheus/rules/{api,db,ingest,security,slo}.rules.yaml`. Every alert carries `runbook_url` annotation.

## Alternatives considered

- **Grafana Cloud (managed).** Rejected — Italian residency + cost; we host our own.
- **Datadog / New Relic.** Rejected — same residency + tool sprawl + per-host pricing.
- **Elastic Stack (ELK).** Rejected — Loki + Tempo are cheaper at our scale and integrate natively with Grafana provisioning.
- **Jaeger** (instead of Tempo). Rejected — Tempo's object-storage backend is cheaper; trace UI in Grafana parity.
- **Pure Prometheus without remote write.** Rejected — OTLP is the application-side standard; collector decouples app from backends.

## Consequences

### Positive

- Single observability story; one Grafana to learn.
- Trace-log-metric correlation via `request_id` / `trace_id`.
- 10% tail sampling on production reduces trace cost ~10× without losing error/slow-request fidelity (tail policy keeps errors + slow).
- Alert routing automated by severity; runbook URL attached to every alert.

### Negative

- OpenTelemetry Collector adds a control plane (DaemonSet + config).
- Loki object-storage requires S3 bucket + lifecycle (Terraform).
- Prometheus retention 30d / 50GB sized to Capacity model; revisit at stretch profile.
- Alert tuning iterative; first deployment will see false positives.

### Neutral

- All choices are CNCF / open-source; switch path is documented.

## Residual risks

- Cardinality explosion: tenant_id label only on counters (Rule 18). Gauges + histograms use `service` / `tier`. Code review enforces.
- Sampling missed root-cause spans: tail sampling keeps `status:ERROR` + `latency>500ms` + 10% probabilistic. Rare paths underrepresented.
- Loki retention of 30d may not satisfy NIS2 audit retention (1y for security-relevant logs). Mitigation: K8s audit log + audit_log table shipped separately to S3 audit bucket with 5y Object Lock retention.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.A.18, §6.B.40, §6.C.58.
- `monitoring/prometheus/values.yaml`, `monitoring/loki/values.yaml`, `monitoring/otel-collector/config.yaml`.
- `monitoring/prometheus/rules/`.
- `grafana/alerting/`.
- `internal/observability/zap_trace.go`, `internal/metrics/metrics.go`.
- `docs/SLI-CATALOG.md`, `docs/OBSERVABILITY.md`.

## Tradeoff Stanza

- **Solves:** empty `monitoring/`, no alert rules, no log/trace correlation, OTel sample 1.0 trace storm risk.
- **Optimises for:** unified pane of glass, runbook-linked actionable alerts, low-cost long-term storage.
- **Sacrifices:** OTel collector control plane; iterative alert tuning; Loki retention < audit retention (compensated by S3 audit bucket).
- **Residual risks:** cardinality explosion (label budget review); tail-sampling rare-path miss (probabilistic floor); Loki cost at scale (revisit at stretch).
