# CSRD / ESRS E1 — Compliance Evidence Pack

**Owner:** `@greenmetrics/secops`, `@greenmetrics/legal`.
**Doctrine refs:** Rule 25 (quality threshold), Rule 61 (compliance ≠ goal — controls map to risks not checkboxes).
**Citation source:** `docs/ITALIAN-COMPLIANCE.md` + EU directive verified annually against `eur-lex.europa.eu`.

## 1. Regulatory scope

- **Directive (EU) 2022/2464** (CSRD — Corporate Sustainability Reporting Directive).
- **ESRS E1** (European Sustainability Reporting Standards — Climate Change).
- Italian transposition: D.Lgs. 6 settembre 2024, n. 125.
- Applicable to GreenMetrics tenants flagged `csrd_in_scope = true`.

## 2. Mapped controls + evidence

| Requirement | Implementation | Mitigation of | Evidence artefact |
|---|---|---|---|
| ESRS E1-1 — Transition plan | Tenant-managed (out of system scope); GreenMetrics provides baseline measurements | n/a | tenant-supplied PDF |
| ESRS E1-4 — Energy consumption + mix | `readings` hypertable per meter + scope category mapping | n/a | report `monthly_consumption.html` |
| ESRS E1-5 — Energy intensity per net revenue | tenant supplies revenue; GreenMetrics computes ratio | n/a | `report:esrs_e1` |
| ESRS E1-6 — GHG emissions Scope 1/2/3 | `emission_factors` × `readings` aggregated; report builder `internal/domain/reporting/esrs_e1.go` | RISK-007 (cross-tenant leak) — RLS enforced | `report:esrs_e1` JSON + HTML; audit_log row per generation |
| ESRS E1-7 — GHG removals + carbon credits | tenant-supplied inputs; GreenMetrics persists | n/a | report payload |
| Audit trail of disclosure preparation | `audit_log` table append-only + S3 Object Lock 5y | RISK-009 (tampering) | `audit_log` exports to S3 |
| Data freshness on Scope 2 (electricity) | ISPRA factor refresh; on breaker open, fallback to cached factors with `data_freshness:"cached_<n>h"` stamp | RISK-004 | `gm_external_api_fallback_total` metric + report stamp |
| Versioned emission factors | `emission_factors` table `(code, valid_from, valid_to)` | RISK-009 | `emission_factor.updated.v1` event archive |
| Tenant isolation | Postgres RLS per `migrations/00006_rls_enable.go` + RBAC per `internal/security/rbac.go` | RISK-007 | `tests/security/rls_isolation_test.go` |

## 3. Audit query

For an ESRS E1 audit, `@greenmetrics/secops` exports:

```sql
-- Audit trail of report generation events for the period.
SELECT id, tenant_id, action, target, request_id, trace_id, created_at, payload
FROM audit_log
WHERE tenant_id = '<tenant>'
  AND action LIKE 'report:%'
  AND created_at BETWEEN '<from>' AND '<to>'
ORDER BY created_at;

-- Emission factor versions used during the period.
SELECT code, scope, year, value, source, valid_from, valid_to
FROM emission_factors
WHERE valid_from <= '<to>' AND (valid_to IS NULL OR valid_to >= '<from>')
ORDER BY scope, code, valid_from;

-- Cached-fallback events that may have affected report freshness.
-- (Prometheus → Grafana panel: gm_external_api_fallback_total)
```

## 4. Evidence delivery

Auditor receives:

- `audit_log` SQL dump signed (PGP) by repo owner.
- `emission_factors` snapshot signed.
- Report PDFs + JSON payloads (immutable in S3 reports bucket).
- Cosign verification chain for the deployed image that generated each report.
- This document + `docs/ITALIAN-COMPLIANCE.md` + `docs/RISK-REGISTER.md`.
- Pentest reports (annual external; redacted for non-`secops` audience).

## 5. Anti-patterns rejected

- Hand-typed evidence — must come from queries against immutable artefacts.
- Removing audit_log entries to "clean up" — REJ; append-only by trigger.
- Compliance checklist without backing risk — Rule 61 violation.
