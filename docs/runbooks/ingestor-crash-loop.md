---
title: Ingestor crash loop
severity: P2
mttd_target: 5m
mttr_target: 30m
owner: "@greenmetrics/app-team"
related_alerts: [ModbusIngestStalled, BreakerOpen, IngestQueueSaturated]
last_tested: 2026-04-25
review_date: 2026-07-25
---

# Runbook — Ingestor crash loop

**Doctrine refs:** Rule 15, Rule 20, Rule 36, Rule 41.

## Symptoms

- Alertmanager firing `ModbusIngestStalled` or `BreakerOpen{name=~"modbus_.*"}`.
- `gm_ingest_readings_total{protocol="modbus"}` flatlined.
- Backend logs (Loki): `ingestor: Modbus dial failed`, `goroutine panic`, `errgroup error`.
- Pod `RESTARTS` count climbs (`kube_pod_container_status_restarts_total`).

## Diagnosis

```bash
# 1. Check ingestor goroutine logs filtered.
kubectl logs -n greenmetrics deploy/greenmetrics-backend --since=10m | jq 'select(.component=="ingestor")'

# 2. Check breaker state per host.
curl -s http://greenmetrics-backend.greenmetrics:8082/api/internal/metrics \
  | grep -E '^gm_breaker_state'

# 3. Check Modbus host reachability.
kubectl run -it --rm net-debug --image=nicolaka/netshoot --restart=Never -- \
  bash -c 'for h in $MODBUS_HOSTS; do nc -zv -w2 "$h" 502; done'

# 4. Verify NetworkPolicy allows the egress.
kubectl describe networkpolicy -n greenmetrics allow-backend-to-meters
```

## Mitigation

### M1 — Disable a single failing host

```bash
# Patch ConfigMap to drop the failing slave ID(s).
kubectl edit configmap greenmetrics-config -n greenmetrics
# MODBUS_SLAVE_IDS=1,2,3,4,5 → 1,2,3 (drop 4,5)
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
```

### M2 — Disable Modbus protocol entirely

```bash
kubectl set env deployment/greenmetrics-backend -n greenmetrics INGESTOR_MODBUS_DISABLED=true
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
```

Customer impact: no Modbus ingest until reversed; pulse + manual paths continue.

### M3 — Force breaker reset

If breakers stuck open due to historic failures even though upstream is now healthy:

```bash
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
# Per-replica breakers reinitialise on boot.
```

## Recovery

1. Verify the failing host is reachable from cluster.
2. If host fixed: re-enable in ConfigMap (M1 reverse) or `INGESTOR_MODBUS_DISABLED-` (M2 reverse).
3. Restart backend; verify `gm_ingest_readings_total{protocol="modbus"}` resumes.
4. Verify `gm_breaker_state{name=~"modbus_.*"} == 0` within 30 s.

## Post-mortem

If recurring (≥ 2 incidents in 30d): file an issue to add per-host monitoring + auto-quarantine.
