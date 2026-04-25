# Backend System Map

**Doctrine refs:** Rule 30 (backend as first-class system), Rule 41, Rule 42.
**Owner:** `@greenmetrics/app-team`.
**Last updated:** 2026-04-25.

## 1. Process boundaries

| ID | Boundary | Process | State owned | Failure domain | Durability | Latency budget |
|---|---|---|---|---|---|---|
| PB-01 | HTTP API | `cmd/server` Fiber app | request scope | per-request | none beyond DB | p99 per `docs/backend/slo.md` |
| PB-02 | Ingestor pool | goroutines in `cmd/server` (errgroup + ants) | per-protocol cursor + bounded queue | per-protocol; degradation cascades to DLQ | DB on flush | sub-second poll cycle |
| PB-03 | OCPP WebSocket dialler | per-CP goroutine in `cmd/server` | per-CP session | per-CP | DB on event | bounded by CS RTT |
| PB-04 | Async-job worker | `cmd/worker` Asynq consumer | job state in Redis | per-worker | Redis + DB | minutes per job |
| PB-05 | Migration runner | `cmd/migrate` goose CLI | none | one-shot | DB | minutes |
| PB-06 | Simulator | `cmd/simulator` Modbus TCP server | RNG-seeded synthetic state | per-instance | none (dev only) | sub-second |

PB-02 and PB-03 are co-located inside the API process for now (single binary). Plan: extract PB-02 to a separate Deployment when ingest rate sustains > 5 k readings/s/replica or when ingestor failure cascades start affecting API SLO. Trigger documented in `docs/CAPACITY.md` §10.

PB-04 is a separate process from S4 onward. Same image, different command (`/greenmetrics-backend worker` vs `/greenmetrics-backend serve`).

## 2. Data ownership matrix

For each table, only the listed process boundaries may **write**. Reads are gated by RLS + RBAC at the transport layer.

| Table | Writers | Readers (RLS-scoped) |
|---|---|---|
| `tenants` | `auth_handler` (provisioning), `migrate` | all |
| `users` | `auth_handler` (provisioning + password update), `migrate` | `auth_handler` |
| `meters` | `meter_handler`, `migrate` | all |
| `meter_channels` | `meter_handler`, `migrate` | all |
| `readings` | `readings_handler.Ingest` (HTTP), `ingestor_runner` (PB-02), `pulse_handler.Ingest` | all |
| `reports` | `reports_handler.Generate` (queue), `worker_reports` (write payload + status) | all |
| `alerts` | `alert_engine`, `alerts_handler.Ack` | all |
| `emission_factors` | `emission_factors_handler.Create`, `migrate` (seed), ISPRA refresh job | all |
| `audit_log` | `audit_middleware` only (append-only via WITH CHECK) | `audit_handler` (RBAC `audit:read`) |
| `idempotency_keys` | `idempotency_middleware` only | `idempotency_middleware` only |
| `pulse_dlq` (S4) | `pulse_handler.Ingest` (failure path) | `pulse_handler.Replay` (admin) |
| `schema_migrations` | goose (`migrate`) only | goose only |

CI lint flags any `INSERT|UPDATE|DELETE` against a table outside its writer set (custom golangci-lint rule, S3).

## 3. Failure-domain matrix

| Subsystem | Goes down | Blast radius | Degradation |
|---|---|---|---|
| TimescaleDB primary | reads + writes fail | API readiness flips to "degraded"; `/api/ready` returns 503; pods removed from Service | RDS multi-AZ failover < 60 s; runbook `docs/runbooks/db-outage.md` (S4) |
| TimescaleDB replica | read latency rises; some queries route to primary | API still serves; CAGG queries may slow | depend on read replica CPU |
| Redis (Asynq) | new jobs cannot be enqueued | `POST /v1/reports` returns 503 with Retry-After; ingest path unaffected | Redis is single-AZ; restart in < 60 s; jobs in flight are retried by Asynq on resume |
| ISPRA upstream | factor refresh fails | breaker opens; report generation falls back to cached factors with `data_freshness: "cached_<n>h"` stamp | `gm_external_api_fallback_total` alerts at >0/min; runbook `docs/runbooks/region-failover.md` |
| Modbus simulator (dev only) | dev ingest stops | dev impact only; not a prod path | n/a |
| OCPP central system | per-CP WebSocket fails | per-CP only; other CPs unaffected | breaker per CP; jittered reconnect |
| OTel collector | trace export fails | tracing degraded; logs + metrics unaffected | OTel SDK retries with backoff; spans buffered (bounded) |
| Prometheus | metric collection halts | dashboards stale; alerts based on Prometheus rules silent | dead-man-switch alert via Alertmanager direct; runbook `docs/runbooks/grafana-down.md` (S4) |
| ESO | secret refresh halts | new secret rotations don't propagate; existing secrets in pod env unaffected | runbook `docs/runbooks/secret-rotation.md` (S4) |
| Argo CD | GitOps reconciliation halts | manual `kubectl apply` reverts paused; new gitops PRs not deployed | runbook `docs/runbooks/argocd-down.md` (S4) |
| Kyverno admission | new pod creation blocked | restarts unaffected; new deploys queued | restart Kyverno within MTTR ≤ 10 min |
| API replica panics | per-replica request error | recover middleware catches; pod stays alive; alert if rate > 1% over 5m | replica restart by K8s if liveness fails |

## 4. Consistency model

Per-aggregate consistency declared in `docs/backend/consistency.md` (S3) — summary:

| Aggregate | Consistency | Idempotency | Ordering |
|---|---|---|---|
| tenants | strong | n/a | n/a |
| users | strong | by email PK | n/a |
| meters | strong | by `Idempotency-Key` on POST | n/a |
| readings | strong write + eventual aggregate (CAGG ≤ 60s lag) | by `(tenant, meter, channel, ts)` unique index + `Idempotency-Key` | per-meter via `ingest_seq bigserial` |
| reports | strong (generated state in OLTP); eventual on payload fetch | by `Idempotency-Key` | n/a |
| alerts | strong | n/a (alert_engine writes) | per-meter via `created_at` |
| audit_log | strong append-only | n/a (middleware writes once per request) | per-tenant via `created_at` |
| idempotency_keys | strong | by PK `(tenant_id, key)` | n/a |

## 5. Process boundary diagram

```
                  ┌──────────────┐
        ingress → │   API (PB-01)│ ← HTTP ← clients
                  ├──────────────┤
                  │  Ingestors   │ ← Modbus / M-Bus / SunSpec
                  │   (PB-02)    │ ← OCPP WebSocket
                  │              │ ← Pulse webhook
                  ├──────────────┤
                  │ OCPP dialler │
                  │   (PB-03)    │
                  └──────┬───────┘
                         │ enqueue (Redis)
                         ▼
                  ┌──────────────┐
                  │ Worker (PB-04)│ ← Asynq queues
                  └──────┬───────┘
                         │ DB write (reports, payload)
                         ▼
                  ┌──────────────┐
                  │ TimescaleDB  │
                  │   (RLS on)   │
                  └──────────────┘

   one-shot:  cmd/migrate (PB-05) — pre-deploy K8s Job
   dev only:  cmd/simulator (PB-06)
```
