---
title: Capacity spike
severity: P2
mttd_target: 5m
mttr_target: 30m
owner: "@greenmetrics/sre"
related_alerts: [P99IngestBudget, P99LoginBudget, PgxPoolAcquireSlow, IngestQueueSaturated, "Disk usage"]
last_tested: 2026-04-25
review_date: 2026-07-25
---

# Runbook — Capacity spike

**Doctrine refs:** Rule 15, Rule 16, Rule 20, Rule 22, Rule 38.

## Symptoms

- Latency p99 budget breach (`P99IngestBudget`, `P99LoginBudget`).
- HPA at max replicas, CPU/memory near limits.
- pgx pool acquire p99 > 50 ms (`PgxPoolAcquireSlow`).
- Disk approaching 80%.
- Cost dashboard shows budget burn.

## Diagnosis

```bash
# HPA state.
kubectl get hpa -n greenmetrics
kubectl describe hpa -n greenmetrics greenmetrics-backend

# Top tenants by ingest rate.
# (Loki + Grafana panel; or directly in PromQL.)
# topk(10, sum by (tenant_id) (rate(gm_ingest_readings_total[5m])))

# Pool budget headroom.
# histogram_quantile(0.99, sum by (le) (rate(gm_db_pool_acquire_duration_seconds_bucket[5m])))

# Disk per pod.
kubectl top pod -n greenmetrics
df -h on RDS instance via CloudWatch FreeStorageSpace metric.
```

## Mitigation

### M1 — Vertical scale (RDS)

```bash
aws rds modify-db-instance --db-instance-identifier greenmetrics-prod \
  --db-instance-class db.r6g.xlarge --apply-immediately
# Multi-AZ failover ~2 min downtime; acceptable for capacity event.
```

### M2 — Horizontal scale beyond HPA max

```bash
# Temporarily bump HPA max.
kubectl edit hpa -n greenmetrics greenmetrics-backend
# spec.maxReplicas: 10 → 20
```

### M3 — Deploy pgbouncer (≥ 6 replicas trigger)

If sustained replica count ≥ 6, RDS `max_connections=100` saturates. Ship pgbouncer per `docs/CAPACITY.md` §10 (transaction pooling).

### M4 — Tighten ingest rate limit

```bash
kubectl set env deployment/greenmetrics-backend -n greenmetrics \
  RATE_LIMIT_INGEST_PER_MINUTE=150   # halved from 300
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
```

Customer impact: 429 Retry-After on burst clients.

### M5 — Drain ingest backlog

```bash
# Scale temp ingest worker replica.
kubectl scale deployment greenmetrics-backend -n greenmetrics --replicas=10
# Or enable spill to disk (S4 feature flag).
kubectl set env deployment/greenmetrics-backend -n greenmetrics INGEST_SPILL=true
```

### M6 — Disk: extend EBS

```bash
aws ec2 modify-volume --volume-id <id> --size 200
# K8s detects, expands; verify `kubectl get pvc -n greenmetrics`.
```

## Recovery

1. p99 latency back within budget.
2. HPA at < 70% of max.
3. pgx pool acquire p99 < 50 ms.
4. Cost dashboard 30-day trend back within env budget.

## Post-mortem

Required if event affected SLO or cost > +20% / day. Update `docs/CAPACITY.md` worked examples with new data point.
