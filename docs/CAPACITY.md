# GreenMetrics Capacity Model

**Doctrine refs:** Rule 16 (scalability), Rule 22 (cost awareness), Rule 38 (backend scalability).
**Owners:** `@greenmetrics/platform-team`, `@greenmetrics/sre`.
**Review cadence:** quarterly; re-baseline whenever K8s `resources.limits` change or after any DR drill.

## 1. Variables

| Symbol | Meaning | Source |
|---|---|---|
| T | tenants | tenant table count |
| M | meters per tenant (mean) | meter table count / T |
| R | readings per minute per meter | ingest rate sample over 7d |
| D | retention days for raw readings | `migrations/0004_retention.sql` (currently 90d raw / 1y / 3y / 10y) |
| B | average payload bytes per reading | sample of `readings` row size |
| C | concurrent users (peak) | from auth log p95 |
| W | workers per replica | pgx pool max (currently 25) |
| H | hold time per request (p99) | bench |

## 2. Per-meter ingest formula

```
rows_per_day  = T · M · R · 1440
raw_bytes/day ≈ rows_per_day · B
chunk_count   ≈ D
CAGG_storage  ≈ aggregate(15min, 1h, 1d) ≈ 0.20 × raw_bytes/day for 15min CAGG
```

`B ≈ 80 bytes` (timestamp + 2 UUID + double + smallint quality + 2 indexes' overhead) measured against current schema. Compression on chunks ≥ 7d typically yields 8–12× ratio (segment_by `meter_id`); use 10× as planning constant.

Steady-state storage: `raw_bytes/day · D / 10` (compressed) + `CAGG storage at 0.20× / 0.05× / 0.01× across 15m/1h/1d bands`.

## 3. Worked examples

| Profile | T | M | R | rows/day | raw GB/day | compressed @ 90d (GB) | rec. RDS class |
|---|---|---|---|---|---|---|---|
| Small | 1 | 5 | 4 | 28 800 | 0.002 | 0.02 | db.m6g.large (2 vCPU / 8 GB) |
| Medium | 50 | 20 | 4 | 5 760 000 | 0.46 | 4.1 | db.r6g.large (2 vCPU / 16 GB) |
| Large | 500 | 50 | 4 | 144 000 000 | 11.5 | 103.5 | db.r6g.4xlarge (16 vCPU / 128 GB) + read replica |
| Stretch | 2 000 | 100 | 12 | 3 456 000 000 | 276 | 2 484 | db.r6g.16xlarge + 2 read replicas + Timescale space partitioning by tenant_id (ADR-010 trigger) |

Add 30% headroom for re-aggregation, vacuum, replication slot churn.

## 4. Read scaling

- Single-instance ceiling at ~70% sustained CPU over 1h → add `aws_db_instance.timescale_replica` (Terraform module to ship in S3 if profile reaches medium-large).
- CAGG hit rate target ≥ 90% on time-bucket queries; `aggregated` endpoint gates queries to the smallest bucket that satisfies the request window.

## 5. Compute (HTTP API)

- pgx pool max 25 (`internal/repository/timescale_repository.go:46`).
- Average request hold time p99 = 100 ms (default budget).
- Per-pod ceiling = `25 / 0.1s = 250 RPS sustained` (CPU/IO permitting).
- Scaling: HPA 2–10 replicas → 500–2 500 RPS sustained API capacity.
- Beyond 6 replicas: deploy pgbouncer (transaction pooling) to avoid `max_connections=100` saturation on RDS.

## 6. Compute (Ingest path)

- Modbus TCP poll cadence default 30s; per-host ceiling ~10 slaves per goroutine before tail latency creeps.
- M-Bus serial 300s, single port per replica → strictly vertical.
- SunSpec poll 30s.
- OCPP per-CP WebSocket; sticky session per charge-point ID.
- Bounded ingest channel `INGEST_QUEUE_DEPTH=10000` between source goroutines and batched writer; `gm_ingest_dropped_total` alerts at >0/min (S4 deliverable).
- Batched writer drains 1000-row batches every 100 ms via `pgx.CopyFrom`; benchmarked target ≥ 50k readings/s (`tests/bench/repo_insert_bench_test.go`, S5).

## 7. Compute (Async worker)

- Asynq + Redis (S4 deliverable).
- Report builders are CPU-bound for ESRS E1 (large CAGG aggregation queries) and Piano 5.0 (HTML/PDF templating).
- Worker pool size: per-replica `min(GOMAXPROCS, 8)` workers; HPA 1–4 replicas.

## 8. Network

- Backend ingress p95 < 200 ms target (CloudFront + nginx ingress + Fiber).
- Egress: ISPRA/Terna/E-Distribuzione allow-listed in NetworkPolicy; outbound TLS only.
- DNS: in-cluster CoreDNS; cache TTL respected.

## 9. Recovery + DR

- RPO 1 h (RDS automated backup interval).
- RTO 4 h for region failover (DR runbook lands S5 + annual drill).
- Snapshot policy: AWS Backup every 4 h, retain 90 d.
- Audit bucket: Object Lock `compliance` mode, 5-y retention.

## 10. Triggers for ADR

| Trigger | ADR |
|---|---|
| Profile reaches Medium (50 tenants) | ADR-010 — TimescaleDB space partitioning |
| Replicas ≥ 6 | ADR for pgbouncer transaction pooling |
| Sustained pgx-pool acquire latency > 50 ms p99 | ADR for pool tuning + maybe move audit writes async |
| CAGG refresh lag > 120 s | ADR for CAGG refresh frequency tuning (cross-ref `RISK-030`) |
| Egress to >5 external APIs | ADR for service mesh re-evaluation (REJ-01 review) |
| New tenant onboarding rate > 10/week | Automation review + capacity re-baseline |

## 11. Observability

Capacity-model variables surface as metrics:

- `gm_ingest_readings_total{tenant_id,protocol,result}` → R per tenant
- `gm_ingest_queue_depth` → backpressure
- `gm_ingest_dropped_total{reason}` → saturation
- `gm_db_pool_acquire_duration_seconds` → pool budget headroom
- `gm_cagg_refresh_duration_seconds{view}` → CAGG cost
- `gm_async_job_duration_seconds{type,result}` → worker capacity

Quarterly review verifies the model against actuals within ±15%.
