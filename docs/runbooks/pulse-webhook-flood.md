---
title: Pulse webhook flood
severity: P2
mttd_target: 5m
mttr_target: 30m
owner: "@greenmetrics/app-team"
related_alerts: [IngestQueueSaturated, IngestDropped]
last_tested: 2026-04-25
review_date: 2026-07-25
---

# Runbook — Pulse webhook flood

**Doctrine refs:** Rule 36, Rule 39, Rule 42.

## Symptoms

- Alertmanager `IngestQueueSaturated` or `IngestDropped`.
- 429 Retry-After spike on `/api/v1/pulse/ingest`.
- A single tenant_id dominates `topk(rate(gm_ingest_readings_total))`.
- HMAC-failure rate elevated → possible spoofing.

## Diagnosis

```bash
# Top tenants by pulse rate (PromQL).
# topk(10, sum by (tenant_id) (rate(gm_ingest_readings_total{source="pulse_webhook"}[5m])))

# HMAC failure rate.
# rate(gm_pulse_signature_invalid_total[5m])

# Identify offending source IP via Loki.
# {service="greenmetrics-backend"} |= "pulse_handler" | json | request_id != ""
```

## Mitigation

### M1 — Per-tenant throttle

```bash
# Apply tighter rate limit to the hot tenant (in-memory via API).
curl -X POST http://greenmetrics-backend:8082/api/v1/admin/rate-limit \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"tenant_id":"<uuid>", "ingest_per_minute": 60}'
```

### M2 — Block at WAF

If spoofing or non-customer source:

```bash
aws wafv2 update-ip-set --scope CLOUDFRONT --id <id> --addresses <ip>
```

### M3 — Enable spill mode

```bash
kubectl set env deployment/greenmetrics-backend -n greenmetrics INGEST_SPILL=true
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
```

Drains queue to local disk; replays when queue recovers.

### M4 — Rotate the tenant's `PULSE_WEBHOOK_SECRET` (if HMAC failures point to leaked secret)

Coordinate with customer; provision new per-tenant HMAC secret in Secrets Manager (per-tenant secret model lands S3 → tenant.flags `pulse_v2`).

## Recovery

1. `gm_ingest_queue_depth` < 4000 sustained.
2. `gm_ingest_dropped_total` rate == 0.
3. Per-tenant rate limit normalised.
4. HMAC failure rate < baseline.

## Post-mortem

Required if drop count > 0. Add per-tenant rate limit to the tenant's plan.
