# 0015 — Bounded ingest queue with drop policy + optional disk spill

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 15, 36, 37, 41, 42
**Review date:** 2027-04-25
**Mitigates:** RISK-014, RISK-015.

## Context

Ingest sources (Modbus, M-Bus, SunSpec, OCPP, Pulse webhook) all push readings to the repository synchronously today. Without a bounded buffer:

- A slow DB stops ingest sources, which stop polling — Modbus tickers skew.
- A burst of pulse webhooks can saturate the pgx pool — head-of-line latency for everyone.
- Goroutine count climbs unboundedly under sustained load — RSS pressure → OOMKill.

## Decision

Insert a **bounded channel + batched DB writer** between sources and the repository:

- `chan PipelineReading` capacity `INGEST_QUEUE_DEPTH` (default 10000).
- Single batched writer drains up to 1000-row batches every 100 ms via `pgx.CopyFrom`.
- Drop policy on saturation: log + Prometheus counter `gm_ingest_dropped_total{reason="queue_full"}`. Sources never block.
- Sources expose two submit modes:
  - `Submit(r)` — non-blocking (Modbus, M-Bus tickers — must not skew).
  - `SubmitBlocking(ctx, r)` — backpressure honoured up to ctx deadline (HTTP ingest, can return Retry-After).
- Optional disk spill via `INGEST_SPILL=true` (boltdb-backed, S5 follow-on) when the queue is saturated; drained when queue recovers.
- Ingestor refactor (Rule 41): `errgroup` + `panjf2000/ants` worker pool replaces naked goroutines.
- Implementation in `internal/services/ingest_pipeline.go`.

## Alternatives considered

- **Unbounded channel.** Rejected — Goroutine pile + RSS pressure → OOMKill.
- **Push directly to DB without buffer.** Rejected — head-of-line latency + Modbus skew.
- **Use Asynq for ingest too.** Rejected — Asynq is for jobs (minutes); ingest is sub-second; Asynq overhead per-row dominates.
- **Per-source queue.** Rejected — Rule 13 (abstraction cost > leverage); single queue with reason-tagged drops is enough.
- **Kafka.** Rejected — REJ-33; overkill.

## Consequences

### Positive

- Modbus tickers never skew (sources non-blocking).
- pgx pool never directly exhausted by ingest path (writer is single-threaded into the pool).
- Batched `CopyFrom` benchmarked target ≥ 50k readings/s on staging.
- Saturation surfaces as a Prometheus counter, not a silent stall.

### Negative

- Drop policy means data loss under saturation (acceptable per RISK-014 mitigation; spill mode optional escape).
- Per-replica buffer (not shared) — coordination via Prometheus aggregation, not direct.
- Buffered data lost on pod crash (≤ 100 ms × queue depth = up to 10k readings — acceptable for non-billing-critical data).

### Neutral

- Adds `panjf2000/ants` dep for the source-side pool (Rule 41 cross-ref).

## Residual risks

- Drop counter increments hidden if Prometheus down. Mitigation: Alertmanager direct dead-man-switch.
- Disk spill in runaway scenario fills PVC. Mitigation: spill cap; alert at 80% PVC use.
- Writer goroutine panic loses entire batch in flight. Mitigation: recover middleware in writer goroutine; per-batch error metric.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.A.15, §6.B.36.
- `internal/services/ingest_pipeline.go`.
- `RISK-014`, `RISK-015`.
- ADR-0009 (circuit breakers).

## Tradeoff Stanza

- **Solves:** Modbus skew under DB slowdown; pgx pool exhaustion from ingest path; goroutine pile under load.
- **Optimises for:** source loop fidelity (Modbus cadence preserved); writer-side throughput via CopyFrom batching; saturation observability.
- **Sacrifices:** data loss under saturation (drop policy); per-replica buffer (coordination via metrics); ≤ 100 ms × queue depth lost on pod crash.
- **Residual risks:** silent drop if Prometheus down (dead-man); spill PVC fill (cap + alert); writer panic loses batch (recover + metric).
