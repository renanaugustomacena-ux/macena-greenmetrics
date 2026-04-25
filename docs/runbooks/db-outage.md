---
title: TimescaleDB outage
severity: P1
mttd_target: 60s
mttr_target: 4h
owner: "@greenmetrics/sre"
related_alerts: [DBPrimaryDown, APIDown, PgxPoolAcquireSlow, CAGGRefreshLag]
last_tested: 2026-04-25
review_date: 2026-07-25
---

# Runbook — TimescaleDB outage

**Severity:** P1.
**MTTD target:** 60 s. **MTTR target:** 4 h (RTO per `docs/SLO.md`).
**Doctrine refs:** Rule 15, Rule 20, Rule 36, Rule 60.

## Symptoms

- Alertmanager firing `DBPrimaryDown` or `APIDown` or `PgxPoolAcquireSlow`.
- Backend `/api/health` returns `{"status":"degraded","dependencies":{"timescaledb":{"status":"error"}}}`.
- Backend `/api/ready` returns 503; pods removed from Service.
- Logs (Loki): `pgx: failed to dial`, `connection refused`, `pool exhausted`.

## Diagnosis

```bash
# 1. Check RDS health.
aws rds describe-db-instances --db-instance-identifier greenmetrics-prod \
  --query 'DBInstances[].[DBInstanceStatus,AvailabilityZone,Engine,EngineVersion]' --output table

# 2. CloudWatch metric — connections, CPU, IOPS.
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS --metric-name DatabaseConnections \
  --dimensions Name=DBInstanceIdentifier,Value=greenmetrics-prod \
  --start-time "$(date -u -d '15 min ago' +%FT%TZ)" --end-time "$(date -u +%FT%TZ)" \
  --period 60 --statistics Average,Maximum

# 3. PG-side check from a debug pod (temporary, distroless production has no shell).
kubectl run -it --rm pg-debug --image=postgres:16-alpine --restart=Never -- \
  psql "$DATABASE_URL" -c "SELECT now(), pg_is_in_recovery();"

# 4. Check active connections + locks.
psql "$DATABASE_URL" <<EOF
SELECT count(*), state, application_name
  FROM pg_stat_activity
  WHERE usename='app_user' GROUP BY 2,3;

SELECT relation::regclass, mode, granted, pid, query
  FROM pg_locks JOIN pg_stat_activity USING (pid)
  WHERE NOT granted;
EOF

# 5. Check Multi-AZ failover state.
aws rds describe-events --source-identifier greenmetrics-prod --duration 60
```

## Mitigation

### M1 — Multi-AZ failover (primary down, replica healthy)

```bash
# RDS Multi-AZ does this automatically; manual trigger:
aws rds reboot-db-instance --db-instance-identifier greenmetrics-prod --force-failover
# Wait 60s; verify endpoint resolves to new primary.
```

Recovery time: < 60 s automatic, < 5 min manual.

### M2 — Read-only mode (DB present but degraded)

```bash
# Toggle backend feature flag to refuse writes.
kubectl set env deployment/greenmetrics-backend -n greenmetrics READ_ONLY_MODE=true
kubectl rollout status deployment/greenmetrics-backend -n greenmetrics
```

`/api/v1/readings/ingest`, `/v1/meters` POST, `/v1/reports` POST return 503 Retry-After. Reads continue.

### M3 — Backpressure shed

If the pgx pool is saturated but DB is otherwise healthy:

```bash
# Reduce ingest replicas to drain queue.
kubectl scale deployment greenmetrics-backend -n greenmetrics --replicas=2
# Or temporarily raise rate limit denial window.
kubectl set env deployment/greenmetrics-backend -n greenmetrics RATE_LIMIT_INGEST_PER_MINUTE=60
```

### M4 — Promote DR replica (region or AZ outage)

See `docs/runbooks/region-failover.md` (linked).

## Recovery

1. Verify RDS endpoint reachable from cluster: `kubectl run -it --rm pg-test --image=postgres:16-alpine --restart=Never -- pg_isready -h <endpoint>`.
2. Restart backend pods to refresh pgx pool: `kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics`.
3. Verify `/api/ready` returns 200 across all replicas.
4. Verify `gm_db_pool_acquire_duration_seconds` p99 < 50 ms.
5. Verify CAGG refresh completes within 60 s of the next refresh window.
6. If `READ_ONLY_MODE` was toggled, reverse: `kubectl set env deployment/greenmetrics-backend -n greenmetrics READ_ONLY_MODE-`.

## Post-mortem

Within 5 business days. Template in `docs/INCIDENT-RESPONSE.md` (S5).

Required fields:

- Timeline (start of impact → detection → mitigation → recovery).
- Contributing factors (root cause + at least one second-order).
- Action items (owner + date for each).
- Public summary (status page).

## Related

- `docs/RELIABILITY-MODEL.md`.
- `docs/INCIDENT-RESPONSE.md`.
- `docs/SLO.md` — RPO 1h / RTO 4h.
- ADR-0009 — circuit breakers.
