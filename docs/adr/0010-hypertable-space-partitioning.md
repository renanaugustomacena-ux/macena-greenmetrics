# 0010 — Hypertable space partitioning by tenant_id

**Status:** Proposed (deferred — trigger condition not yet met)
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 16, 22, 38
**Review date:** when capacity profile reaches Medium (50 tenants) per `docs/CAPACITY.md`

## Context

TimescaleDB hypertables today partition by **time** (1d chunks). At small/medium profiles (≤50 tenants) this is fine. At large/stretch (≥500 tenants), a hot tenant can dominate a chunk's CPU/IO during compression or query, hurting per-tenant tail latency.

Timescale supports an additional **space dimension** (`partitioning_column`) — chunks within a time period are further sharded by `tenant_id` hash, parallelising query and compression work.

## Decision (deferred)

When `docs/CAPACITY.md` reaches the Medium profile (50 tenants):

- Introduce space partitioning: `SELECT add_dimension('readings', 'tenant_id', number_partitions => 4);`
- Chunk tuple becomes `(time, tenant_id_hash)`.
- Existing data unchanged; new chunks created with the new dimension.
- CAGGs continue to work (CAGG aggregates over the time dimension).
- Compression policy continues to work per chunk.

## Alternatives considered

- **Per-tenant table.** Rejected — at 500+ tenants, table count is unmanageable; CAGG topology multiplies.
- **Per-tenant database.** Rejected — backup, monitoring, migration topology multiplies; operational overhead unmanageable.
- **Hash partition at app layer.** Rejected — Timescale handles this natively.
- **Stay time-only.** Defer — works until measured bottleneck.

## Consequences

### Positive

- Per-tenant query work parallelised across CPUs.
- Hot-tenant compression work doesn't block cold tenants.
- Linear scaling beyond Medium profile.

### Negative

- Chunk count multiplies by `number_partitions` factor.
- Catalog overhead grows.
- Re-partitioning (changing `number_partitions`) requires hypertable recreation — plan ahead.

### Neutral

- Timescale-native; no app changes beyond connection pool sizing.

## Residual risks

- Wrong `number_partitions` choice — start at 4, scale to 8/16 as profile grows. Document in `docs/CAPACITY.md`.

## Trigger

Capacity profile reaches Medium (50 tenants × 20 meters × 4 r/min = 5.76 M rows/day) sustained for 30 days; or a single-tenant query consistently exceeds the SLO budget for `/v1/readings/aggregated`.

## Tradeoff Stanza

- **Solves:** future hot-tenant tail latency at large/stretch profile.
- **Optimises for:** parallelisable query + compression work; linear horizontal scaling beyond Medium.
- **Sacrifices:** chunk catalogue size; re-partitioning cost if `number_partitions` chosen wrong.
- **Residual risks:** wrong partition count (start small, scale up); deferred decision (can't be applied retroactively without hypertable recreation).
