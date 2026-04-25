# 0009 — Circuit breakers: sony/gobreaker per host

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 15, 36, 37
**Review date:** 2027-04-25
**Mitigates:** RISK-004, RISK-008.

## Context

Backend has outbound dependencies: ISPRA (factor refresh), Terna (grid mix), E-Distribuzione (POD lookup), Grafana (dashboard data), OCPP (per-CP WebSocket), Modbus per host (ingestor). Today these are direct calls with single timeouts and no breaker. A slow upstream blocks goroutines, exhausts the pgx pool, and cascades to API SLO failure.

## Decision

Adopt `sony/gobreaker/v2`. New `internal/resilience/breaker.go` exposes `NewBreaker(name, opts...)`. Per-host breaker keys (cardinality bounded by configured external endpoints):

- `ispra_emission_factors`
- `terna_grid_mix`
- `e_distribuzione_pod`
- `grafana_dashboard`
- `ocpp_central_system` (per CP)
- `modbus_<host>` (per Modbus host)

Settings:

- Open after 5 consecutive failures or 50% failure ratio over 20 calls.
- Half-open after 30 s.
- One trial probe in half-open; success closes, failure reopens.
- Breaker timeout < poll interval (e.g. 30 s for Modbus) — recovery within one cycle.

State exposed as `gm_breaker_state{name}` Prometheus gauge (0=closed, 1=open, 2=half-open). Alert `BreakerOpen` fires after 5 min sustained.

## Alternatives considered

- **No breaker; rely on timeouts.** Status quo. Rejected — slow upstream + retry storm exhausts pgx pool.
- **`failsafe-go`.** Rejected — newer, smaller community; gobreaker is battle-tested.
- **Custom breaker.** Rejected per Rule 26 (NIH).
- **Service mesh circuit breaker (Istio).** Rejected — REJ-01 (mesh for 2 services is overkill).

## Consequences

### Positive

- Cascading failure contained to per-host blast radius.
- Graceful degradation path: when ISPRA breaker opens, report generation falls back to cached factors with `data_freshness: "cached_<n>h"` stamp.
- Breaker state observable as a metric; runbook `docs/runbooks/db-outage.md` (S4) cross-references.

### Negative

- One more dependency.
- Tuning thresholds requires production observation; first iteration may have false trips.
- Per-host breaker means cardinality grows with configured endpoints — currently bounded by Modbus hosts (≤ 50 in capacity model).

### Neutral

- gobreaker is a small, single-author lib but stable for years; review at 2027.

## Residual risks

- Wrong threshold → false trip cascades to legitimate-traffic outage. Mitigation: observability + tunable thresholds; quarterly review.
- Breaker state not shared across replicas. Mitigation: each replica's breakers operate independently; with 2–10 replicas, the aggregate behaviour converges to the right answer faster than a centralised breaker would react.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.A.15, §6.B.36.
- `internal/resilience/breaker.go` (S4 to ship).
- `docs/runbooks/db-outage.md` (S4).
- `RISK-004`, `RISK-008`.

## Tradeoff Stanza

- **Solves:** cascading failure from slow / down upstreams; pgx-pool exhaustion under retry storm.
- **Optimises for:** blast-radius containment; degraded but live operation; observability of upstream health.
- **Sacrifices:** one library dependency; tuning effort; per-host cardinality on metrics.
- **Residual risks:** false trip from wrong threshold (observability + quarterly tuning); per-replica breaker independence (acceptable convergence).
