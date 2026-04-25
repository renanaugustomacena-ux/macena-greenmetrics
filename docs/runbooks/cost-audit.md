---
title: Monthly cost audit
severity: routine (P3 if findings)
mttd_target: monthly
mttr_target: end-of-month
owner: "@greenmetrics/platform-team"
related_alerts: ["AWS Budgets at 80%", "AWS Budgets at 100%"]
last_tested: 2026-04-25
review_date: 2026-05-25
---

# Runbook — Monthly cost audit

**Doctrine refs:** Rule 22.
**Tool:** `scripts/ops/cost-audit.sh` (S2).

## Cadence

First Friday of each month at 09:00 Europe/Rome.

## Procedure

```bash
# 1. Run the audit.
./scripts/ops/cost-audit.sh > /tmp/cost-audit-$(date +%Y-%m).md

# 2. Review the per-cost-centre dashboard.
# Grafana → "Cost overview" → last 30d trend per CostCenter tag.

# 3. Identify top movers.
# Anything with > +10% MoM growth → investigate.

# 4. Open GitHub issue for any waste candidate.
gh issue create --title "Cost audit: $(date +%Y-%m)" \
  --label cost,priority-medium \
  --body-file /tmp/cost-audit-$(date +%Y-%m).md

# 5. If a budget threshold (80%/100%) is breached, escalate to platform office hours.
```

## Waste candidates

`scripts/ops/cost-audit.sh` lists:

- EBS volumes unattached > 7d.
- RDS replicas idle (CPU < 5% over 30d).
- Orphaned secrets (Secrets Manager `LastAccessedDate` > 90d).
- Grafana datasources with zero queries 30d.
- IAM roles unused (no `AssumeRole` event 90d).
- K8s nodes underutilised (CPU < 20% sustained 7d).

## Anti-patterns rejected

- Reserved Instances year 1 — too early; commit at month 12 after baseline stabilises.
- "Just throw a bigger instance at it" — Rule 22 violation; capacity plan first.
- Disable Cost Explorer — costs cents per month; saves thousands.
