# D.Lgs. 102/2014 — Energy Audit (4-yearly) Evidence Pack

**Owner:** `@greenmetrics/legal`.
**Doctrine refs:** Rule 25, Rule 61.
**Citation source:** D.Lgs. 4 luglio 2014, n. 102 art. 8 + ENEA guidelines (`docs/ITALIAN-COMPLIANCE.md`).

## 1. Regulatory scope

- D.Lgs. 102/2014 art. 8: large enterprises + energy-intensive SMEs must conduct a quadriennial energy audit.
- ENEA-format report submitted via the diagnosi-energetiche portal.

## 2. Mapped controls + evidence

| Requirement | Implementation | Evidence |
|---|---|---|
| Site-level + process-level meter coverage | `meters` with `site` + `cost_centre` taxonomies; ≥ 90% of energy consumption metered | `meter` inventory + reading completeness query |
| 12-month rolling baseline + audit period | `readings` hypertable + retention 90d raw + 1y/3y/10y aggregated | `report:audit_dlgs102` payload |
| Significant energy uses (SEU) identification | `internal/domain/reporting/audit_dlgs102.go` builder | report SEU section |
| Energy performance indicators (EnPI) | per-site + per-process indicators computed | report EnPI section |
| Recommended energy-saving measures | tenant + auditor input; persisted in report payload | report measures section |
| Audit-log of submission | `audit_log` with action `report:audit_dlgs102` | `audit_log` query |

## 3. Operator query

```sql
SELECT count(*) AS meters, sum(case when active then 1 else 0 end) AS active_meters,
       count(distinct site) AS sites, count(distinct cost_centre) AS cost_centres
FROM meters WHERE tenant_id = '<tenant>';

SELECT date_trunc('day', ts) AS day, count(*) AS readings, count(distinct meter_id) AS reporting_meters
FROM readings
WHERE tenant_id = '<tenant>' AND ts BETWEEN '<from>' AND '<to>'
GROUP BY 1 ORDER BY 1;
```

## 4. Evidence delivery

Same pattern as CSRD — immutable artefacts + Cosign verification + audit_log SQL dump signed.
