---
title: Suspected tenant data leak (RLS bypass / cross-tenant read)
severity: P0
mttd_target: 1h
mttr_target: 4h
owner: "@greenmetrics/secops"
related_alerts: [LoginFailureSpike, FalcoCriticalRule, "NetworkPolicy denied"]
last_tested: 2026-04-25
review_date: 2026-07-25
---

# Runbook — Suspected tenant data leak

**Doctrine refs:** Rule 19, Rule 39, Rule 60, Rule 65.

## Severity

P0 if confirmed. Even suspected: treat as P1 until disproven.

## Symptoms

- Customer report: "I saw another tenant's data".
- Falco event: SQL pattern matching cross-tenant SELECT.
- Audit log shows query with tenant_id mismatch.
- WAF anomaly: query patterns bypassing typical RBAC.

## Immediate actions

### A1 — Snapshot for forensics

```bash
# RDS snapshot (immutable point-in-time).
aws rds create-db-snapshot --db-instance-identifier greenmetrics-prod \
  --db-snapshot-identifier "incident-$(date -u +%Y%m%dT%H%M%S)"

# K8s audit log snapshot to S3 (immutable Object Lock bucket).
kubectl logs -n greenmetrics --previous deploy/greenmetrics-backend > /tmp/backend-prev.log
aws s3 cp /tmp/backend-prev.log s3://greenmetrics-audit/incident/$(date -u +%Y%m%d)/backend-prev.log
```

### A2 — Confirm RLS still enabled

```bash
psql "$DATABASE_URL" <<EOF
SELECT relname, relrowsecurity, relforcerowsecurity
FROM pg_class WHERE relname IN ('tenants','users','meters','readings','reports','alerts','audit_log','idempotency_keys');
EOF
```

All rows must show `t` for both columns. If any `f`: structural breach — escalate to IC.

### A3 — Verify policy presence

```bash
psql "$DATABASE_URL" <<EOF
SELECT schemaname, tablename, policyname, cmd, qual, with_check
FROM pg_policies WHERE schemaname='public';
EOF
```

Every RLS-enabled table must have `tenant_isolation` policy (or for `audit_log`: `audit_log_tenant_insert` + `audit_log_tenant_select`).

### A4 — Audit query log for offending pattern

```bash
# Query CloudWatch / pg_stat_statements for queries that read across tenant_ids.
psql "$DATABASE_URL" -c "SELECT query, calls, total_exec_time FROM pg_stat_statements WHERE query LIKE '%WHERE tenant_id%' ORDER BY total_exec_time DESC LIMIT 50;"
```

### A5 — Check for SECURITY DEFINER functions

```bash
psql "$DATABASE_URL" -c "SELECT proname, prosecdef FROM pg_proc WHERE prosecdef = true AND pronamespace = 'public'::regnamespace;"
```

If any: investigate immediately; SECURITY DEFINER bypasses RLS.

## Containment

### C1 — Revoke affected user tokens

If a specific user's tokens are suspected:

```bash
# Force JWT rotation (P0 path; see jwt-secret-rotation runbook).
gh workflow run jwt-rotation.yml -f reason="P0 tenant data leak suspected"
# All in-flight tokens reject.
```

### C2 — Block IP / ASN at WAF

If exfil traffic identified:

```bash
aws wafv2 update-ip-set --scope CLOUDFRONT --id <id> --addresses <new-block>
```

### C3 — Read-only mode

```bash
kubectl set env deployment/greenmetrics-backend -n greenmetrics READ_ONLY_MODE=true
```

### C4 — Take cluster offline (last resort)

```bash
kubectl scale deployment/greenmetrics-backend -n greenmetrics --replicas=0
# Status page: "Service temporarily unavailable for security investigation."
```

## Notification

- **GDPR breach** affecting ≥ 1 EU data subject → notify Garante within **72 h** via `garanteprivacy.it`.
- **NIS2** notification (`D.Lgs. 138/2024`) → ACN within **24 h** preliminary, **72 h** full report.
- Affected customers: direct email within 72 h.
- Status page: continuous updates.

## Recovery

1. Root cause identified + fix shipped via emergency PR.
2. RLS confirmed enabled across all tables.
3. Audit log shows no active offending pattern in last 1 h.
4. JWT rotated.
5. WAF blocks reviewed + maintained.
6. Customer notification complete.

## Postmortem (mandatory within 5 business days)

Sections:

- Timeline (UTC ISO 8601).
- Root cause + at least one second-order contributing factor.
- Number of affected tenants + estimated record count.
- Mitigations now in place.
- Action items with owners + dates.
- Public-facing summary + customer notification copy.
- Risk register update (`docs/RISK-REGISTER.md` RISK-007 review).
