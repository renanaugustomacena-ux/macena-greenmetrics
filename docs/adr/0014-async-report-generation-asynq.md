# 0014 — Async report generation: Asynq + Redis

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 30, 36, 37, 67
**Review date:** 2027-04-25
**Mitigates:** RISK-016.

## Context

`internal/services/report_generator.go` is **525 LoC** with seven `buildXxx` methods on one struct. The HTTP handler `POST /v1/reports` invokes generation **synchronously**: the request holds a pgx connection, runs CAGG aggregation queries, templates HTML/PDF, returns the full payload. For ESRS E1 reports over a 12-month window this can take minutes — head-of-line latency on the OLTP pool, blocking other requests, violating the p99 ingest budget for everyone.

## Decision

Move report generation to an **async worker queue** backed by Redis (Asynq).

- `POST /v1/reports` enqueues a job → returns `202 Accepted` + `Location: /api/v1/jobs/{id}` + `{"job_id":"...","status":"queued"}`.
- New endpoint `GET /v1/jobs/{id}` returns job state {queued, running, succeeded, failed, retry, dead}.
- Worker process: `cmd/worker/main.go`. Same image as `cmd/server`; different command. Identical config + observability boot path.
- Job types: `report:esrs_e1`, `report:piano_5_0`, `report:conto_termico`, `report:tee`, `report:audit_dlgs102`, `report:monthly_consumption`, `report:co2_footprint`, plus `factor:refresh_ispra`, `factor:refresh_terna`.
- Asynq queues: `reports` (weight 6), `factor-refresh` (weight 2), `default` (weight 2). Strict priority.
- Retry: max 3, exponential backoff with jitter; failed job → dead queue + alert.
- Idempotency: Asynq `TaskID` set to `Idempotency-Key` from HTTP layer.
- Custom metric `gm_async_job_duration_seconds{type,result}`.
- Redis added to `docker-compose.yml`; managed Redis (ElastiCache) in production via Terraform.
- Health check: `/api/ready` checks Redis reachability.

## Alternatives considered

- **Keep synchronous.** Rejected — RISK-016, head-of-line latency violates SLO.
- **`riverqueue/river`** (Postgres-only queue). Considered seriously. Pro: no new failure domain, OLTP-DB-only. Con: report jobs are CPU-intensive; running them on the OLTP pool defeats the point. Asynq isolates job traffic from OLTP.
- **Custom in-process goroutine queue.** Rejected — no durability across pod restart.
- **Kafka.** Rejected — overkill for our scale; operational complexity > value.
- **AWS SQS + Lambda.** Rejected — Italian residency simpler with in-cluster Redis; AWS SQS adds cross-AZ network hop + lambda packaging.

## Consequences

### Positive

- `POST /v1/reports` p99 returns in < 100 ms regardless of report size — async ack only.
- Report compute isolated from OLTP pool; no head-of-line latency.
- Worker pool horizontally scalable independently of API replicas.
- Failed jobs visible in dead queue + Asynq Web UI for debugging.

### Negative

- Redis is a **new failure domain** (RISK to be added to register if it materialises).
- Asynq pulls in Redis client + JSON payload encoding.
- Operator must monitor dead queue.
- Status polling pattern: clients call `GET /v1/jobs/{id}` periodically — webhook callback could replace in v2.

### Neutral

- Same image, different command — no extra container build.
- Asynq is well-maintained, single-author project; review at 2027.

## Residual risks

- Redis outage: jobs cannot be enqueued; `POST /v1/reports` returns 503 Retry-After. Mitigation: ElastiCache Multi-AZ in production; Asynq retries on resume.
- Job poisoning via crafted payload: schema validation at enqueue time + bounded payload size.
- Dead queue accumulation: alert on `asynq_queue_size{queue="dead"} > 100`.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.B.37.
- `internal/jobs/queue.go`, `cmd/worker/main.go`.
- `RISK-016`.
- ADR-0009 (circuit breakers — outbound calls from worker).

## Tradeoff Stanza

- **Solves:** synchronous report generation head-of-line latency on OLTP pool.
- **Optimises for:** API SLO protection; horizontally scalable worker tier; durable retry.
- **Sacrifices:** Redis as new failure domain; async pattern complexity for clients (poll-based status).
- **Residual risks:** Redis outage (Multi-AZ + alert); dead queue accumulation (alert); status polling load (mitigate via webhook v2).
