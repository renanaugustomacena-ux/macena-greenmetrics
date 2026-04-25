---
title: Grafana down
severity: P3
mttd_target: 10m
mttr_target: 30m
owner: "@greenmetrics/sre"
related_alerts: []
last_tested: 2026-04-25
review_date: 2026-07-25
---

# Runbook — Grafana down

**Doctrine refs:** Rule 15, Rule 20.

**Important:** Grafana is dashboard-only. Alerts route via **Alertmanager**, not Grafana, so a Grafana outage does **not** silence alerts. Operational paging continues.

## Symptoms

- `kubectl get pods -n greenmetrics` shows `greenmetrics-grafana` `CrashLoopBackOff` or `0/1 Running`.
- Internal users report Grafana UI unreachable.

## Diagnosis

```bash
kubectl describe pod -n greenmetrics -l app=greenmetrics-grafana
kubectl logs -n greenmetrics -l app=greenmetrics-grafana --tail=100
```

Common causes:

- ESO sync failure → Grafana admin password missing → boot refusal.
- PVC full → `grafana.db` write failure.
- Datasource pointing to deleted Prometheus endpoint.

## Mitigation

```bash
# Restart pod.
kubectl rollout restart deployment/greenmetrics-grafana -n greenmetrics

# Check ESO sync.
kubectl get externalsecret greenmetrics-grafana -n greenmetrics

# Reset PVC if disk is full and dashboards are committed in `grafana/provisioning/`.
kubectl delete pvc greenmetrics-grafana -n greenmetrics
kubectl rollout restart deployment/greenmetrics-grafana -n greenmetrics
```

## Recovery

1. Pod `Running` `1/1`.
2. Login with ESO-managed admin password works.
3. Provisioned dashboards visible.
4. Datasource probe returns OK.

## Important reminder

Grafana down is **not** an outage of the alerting path. Alertmanager continues to receive alerts and route them. If Grafana is down for > 24h, escalate to platform office hours for capacity / configuration discussion.
