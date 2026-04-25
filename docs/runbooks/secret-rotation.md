---
title: Generic secret rotation (ESO)
severity: P2
mttd_target: 1h
mttr_target: 2h
owner: "@greenmetrics/secops"
related_alerts: [ESONotSynced]
last_tested: 2026-04-25
review_date: 2026-07-25
---

# Runbook — Generic secret rotation via ESO

**Doctrine refs:** Rule 19, Rule 39, Rule 56, Rule 62.

For JWT-specific rotation see `docs/runbooks/jwt-secret-rotation.md`.

## Symptoms

- Alertmanager firing `ESONotSynced` (status != SecretSynced for 10m).
- Pod env still showing old secret value after Secrets Manager update.

## Diagnosis

```bash
# Check ESO sync status.
kubectl get externalsecret -n greenmetrics
kubectl describe externalsecret -n greenmetrics greenmetrics-backend

# Check Secrets Manager value freshness.
aws secretsmanager describe-secret --secret-id greenmetrics/prod/jwt \
  --query '{LastChanged:LastChangedDate, LastAccessed:LastAccessedDate, Version:VersionIdsToStages}'

# Check IRSA assumption from ESO pod.
kubectl exec -n external-secrets deploy/external-secrets -- \
  aws sts get-caller-identity
```

## Mitigation

### M1 — Force ESO sync

```bash
kubectl annotate externalsecret -n greenmetrics greenmetrics-backend \
  force-sync="$(date +%s)" --overwrite
# Wait < 30s for sync.
```

### M2 — Restart ESO controller

```bash
kubectl rollout restart deployment/external-secrets -n external-secrets
```

### M3 — Restart consumer pods

```bash
# Even after Secret refresh, pods using env-var injection need restart to pick up.
# (Projected volume + sub-path mounts auto-reload; env-var injection does not.)
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
```

## Recovery

1. `kubectl get externalsecret -n greenmetrics` shows `Status: SecretSynced` for all.
2. Pod env (`kubectl exec ... env | grep <KEY>`) shows new value.
3. Application functional with new secret.

## Annual rotation cadence

- AWS Secrets Manager → Lambda rotation configured for credentials with native rotators (RDS, Redis).
- Quarterly cron via `.github/workflows/jwt-rotation.yml` for JWT.
- Manual rotation drill quarterly per `docs/SECOPS-RUNBOOK.md`.
