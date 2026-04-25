---
title: Region failover (eu-south-1 outage)
severity: P0
mttd_target: 5m
mttr_target: 4h
owner: "@greenmetrics/sre"
related_alerts: [DBPrimaryDown, APIDown, AbsentMetricsDeadMan]
last_tested: 2026-04-25 (drill)
review_date: 2026-10-25
---

# Runbook — Region failover

**Doctrine refs:** Rule 15, Rule 20, Rule 25, Rule 60, Rule 65.

## Scope

Active-passive disaster recovery from eu-south-1 (Milan) to a hot-standby region. RPO 1 h / RTO 4 h per `docs/SLO.md`.

**NOT in scope (today):** active-active multi-region — REJ-09 (over-scoping).

## Pre-requisites (verified annually)

- RDS automated backup (every 4 h, retain 90 d) + cross-region snapshot copy to eu-west-1.
- S3 buckets cross-region replication on `greenmetrics-audit` (Object Lock preserved).
- Terraform module `terraform/modules/rds-timescale-replica/` (S5) provisions DR replica.
- DNS Route 53 health-checked failover record.

## Decision criteria

Engage region failover only if:

- eu-south-1 outage confirmed via AWS Health Dashboard.
- Mitigation steps in `docs/runbooks/db-outage.md` exhausted.
- Estimated MTTR for in-region recovery > 4 h.

## Procedure

### Phase 1 — Acknowledge

1. IC opens incident in `#sev1-active`.
2. Status page entry (`Investigating`).
3. Notify GDPR DPO if customer data access disrupted (NIS2 24h notification clock starts).

### Phase 2 — Promote DR replica

```bash
# Promote DR replica to standalone primary.
aws rds promote-read-replica --db-instance-identifier greenmetrics-prod-dr --region eu-west-1
# Verify status.
aws rds describe-db-instances --db-instance-identifier greenmetrics-prod-dr --region eu-west-1 \
  --query 'DBInstances[].DBInstanceStatus'
```

### Phase 3 — Reroute application

```bash
# Update DATABASE_URL in Secrets Manager (AWS-side).
aws secretsmanager put-secret-value --secret-id greenmetrics/prod/db \
  --secret-string '{"url":"postgres://app_user:****@<dr-endpoint>:5432/greenmetrics?sslmode=require"}'
# ESO sync (≤ 1h or force).
kubectl annotate externalsecret -n greenmetrics greenmetrics-backend \
  force-sync="$(date +%s)" --overwrite
# Restart backend.
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
```

### Phase 4 — Reroute DNS

```bash
# Update Route 53 record (or rely on health-check failover if configured).
aws route53 change-resource-record-sets --hosted-zone-id <zone> \
  --change-batch file://failover-rrset.json
```

### Phase 5 — Bring up cluster in DR region

If EKS in eu-south-1 also down: Argo CD App-of-Apps points to gitops repo; bring up EKS in eu-west-1, install Argo CD, point at `gitops/production-dr/` overlay.

### Phase 6 — Customer comms

Status page: `Identified` → `Monitoring` → `Resolved`.
Email: customer success notifies key accounts.
NIS2 preliminary report within 24h to ACN portal (template in `docs/INCIDENT-RESPONSE.md`).
GDPR breach notification to Garante within 72h if PII access affected.

## Failback

Once eu-south-1 restored:

1. Re-enable RDS read replica in eu-south-1; allow sync to catch up.
2. Schedule maintenance window (announce 48h+ ahead).
3. Reverse DNS, Secrets Manager, ESO, deployment routing.
4. Promote eu-south-1 instance back to primary.
5. Return DR to read-replica role.

## Annual drill

Q3 each year. Recorded in `docs/CHAOS-LOG.md`. Validates RPO/RTO. Updates this runbook.
