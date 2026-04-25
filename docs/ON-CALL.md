# On-Call Rota

**Doctrine refs:** Rule 20 (operational reality), Rule 60 (incident response).
**Owner:** `@greenmetrics/sre`.

## Rota

PagerDuty schedule: `greenmetrics-primary`. Weekly handover Friday 17:00 Europe/Rome.

| Tier | Page latency | Role |
|---|---|---|
| Primary | 0 min | Drive incident; coordinate response |
| Secondary | 15 min | Backup; second pair of eyes |
| Manager | 30 min | Customer comms + escalation |
| Repo owner | 60 min | Last-resort decision authority |

## Severity matrix

| Sev | Definition | Response | Comms |
|---|---|---|---|
| P0 | Customer-impacting outage; data leak; security breach | Page immediately | Status page + customer email + NIS2/GDPR clock |
| P1 | Significant degradation; SLO error budget burning fast | Page within 15 min | Status page |
| P2 | Partial degradation; specific feature affected | Slack `#greenmetrics-ops` | Status page if > 1h |
| P3 | Operational toil; non-urgent investigation | Issue + next business day | n/a |

## Handover checklist

- [ ] Open incidents reviewed; mitigations in flight or resolved.
- [ ] Recent postmortems flagged; action items in progress.
- [ ] Cron health (DAST sweep, JWT rotation, cost audit) verified.
- [ ] Calendar conflicts shared (planned PTO, holidays).
- [ ] Slack handover post in `#greenmetrics-ops`.

## Page acknowledgement

- Acknowledge within page-latency window.
- If acknowledgement misses window, secondary auto-paged.
- Self-page only for verification scenarios; never to test.

## Decision authority during incidents

- Primary may take any action in `docs/runbooks/`.
- Manager-level approval required for: scaling deletes, region failover, customer comms beyond status page, NIS2/GDPR notification trigger, JWT emergency rotation purge of overlap window.
- Repo owner approval required for: changes outside doctrine (e.g. disabling Kyverno admission webhook).

## Anti-patterns rejected

- "I'll fix it later" without filing follow-up issue — follow-up is part of recovery.
- Silent runbook deviation — note deviation in incident timeline; postmortem reviews.
- Acknowledging without response capacity — re-route to secondary instead.
