# Incident Response

**Owner:** `@greenmetrics/secops`, `@greenmetrics/sre`.
**Doctrine refs:** Rule 60 (incident response as design input), Rule 65 (regulated quality), Rule 67 (transparency).
**Cadence:** annual full review; quarterly tabletop drill.

## 1. Severity matrix

| Sev | Definition | Response | Comms | Postmortem |
|---|---|---|---|---|
| **P0** | Customer-impacting outage; suspected data breach; security compromise | Page immediately | Status page (Investigating) + customer email + NIS2 24h clock + GDPR 72h clock if PII | mandatory ≤ 5 business days |
| **P1** | Significant degradation; SLO error budget burning fast; security CRITICAL with no exploitation evidence | Page within 15 min | Status page + customer comms if > 1h | mandatory ≤ 5 business days |
| **P2** | Partial degradation; specific feature affected; security HIGH | Slack `#greenmetrics-ops` | Status page if > 1h | recommended |
| **P3** | Operational toil; non-urgent investigation; security MEDIUM | Issue + next business day | n/a | optional |

## 2. Detection sources

- Alertmanager → Slack / PagerDuty.
- Falco runtime events → Loki → Alertmanager.
- GitHub security advisory → Dependabot PR.
- Customer report → support → on-call.
- External pentest finding (annual).
- Internal pentest finding (quarterly).
- AWS Health Dashboard.
- Synthetic probe failure (`blackbox_exporter`).

## 3. Roles during an active incident

| Role | Responsibilities |
|---|---|
| **Incident Commander (IC)** | Coordinates response; makes decisions on severity escalation; owns customer comms thread |
| **Tech Lead** | Drives mitigation + recovery; runs runbooks |
| **Comms Lead** | Status page + customer email + Slack updates |
| **Scribe** | Maintains UTC ISO 8601 timeline in `#sev1-active` thread |
| **Subject Matter Expert** | Domain-specific (DB, networking, security) — pulled in as needed |

For P0/P1: IC, Tech Lead, Scribe required. Same person may not hold IC + Tech Lead.

## 4. Comms plan

### Internal

- `#sev1-active` Slack channel: timeline, decisions, blockers.
- PagerDuty rota: primary → secondary → manager (per `docs/ON-CALL.md`).

### External

- Status page: `Investigating` → `Identified` → `Monitoring` → `Resolved`. Update every 30 min during active incident.
- Customer email: send at `Identified` (P0/P1) and `Resolved`.
- Pre-approved templates in `docs/templates/incident-comms/`.

### Regulatory

- **NIS2 (`D.Lgs. 138/2024`):** preliminary notification within **24 h** to ACN portal. Full report within **72 h**. Pre-filled credentials in Vault under `greenmetrics/compliance/acn-portal`.
- **GDPR breach (Reg. UE 2016/679 art. 33):** notification to Garante within **72 h** if affecting EU data subjects. Notify DPO immediately.
- **CSRD (ESRS E1):** internal note in audit_log; no external regulatory clock.

## 5. Containment playbooks

| Scenario | Playbook |
|---|---|
| Suspected JWT secret leak | `docs/runbooks/jwt-secret-rotation.md` (P0 path) |
| Suspected tenant data leak | `docs/runbooks/tenant-data-leak.md` |
| Compromised IP / user agent | WAF rate-rule update (in `docs/runbooks/pulse-webhook-flood.md`) |
| Suspected supply chain compromise | Disable affected workflow; rotate all secrets exposed to it; rotate OIDC trust |
| Compromised IAM access | Break-glass IAM with CloudTrail alarm; revoke immediately |
| Pod compromise | `kubectl drain` node; preserve PVC for forensics (do not delete); restore from clean image |
| RDS suspected compromise | Snapshot before any action (`aws rds create-db-snapshot`); IC decides on isolate vs investigate |

Read-only mode toggle:

```bash
kubectl set env deployment/greenmetrics-backend -n greenmetrics READ_ONLY_MODE=true
```

Scale to zero (last resort):

```bash
kubectl scale deployment/greenmetrics-backend -n greenmetrics --replicas=0
```

## 6. Forensic readiness

- **Snapshots:** AWS Backup plan EBS snapshots every 4 h, retain 90 d (per `gitops/base/forensics/snapshot-policy.yaml`).
- **Audit log table:** retained 365 d in DB; shipped to Loki + Object Lock S3 audit bucket for 5 y.
- **K8s audit log:** EKS cluster `enabled_cluster_log_types: [api, audit, authenticator, controllerManager, scheduler]` (`terraform/modules/eks/main.tf:39`); shipped to CloudWatch and Loki.
- **CloudTrail:** all AWS API calls logged to S3 audit bucket with Object Lock `compliance` mode 5 y.
- **Container runtime events:** Falco → Loki, retained 30 d in Loki and 365 d in S3.

Evidence preservation rule: **never delete logs / snapshots / audit rows during active incident** without IC approval. Even after recovery, keep forensic state ≥ 90 d.

## 7. Postmortem template

```markdown
# Postmortem — <SHORT TITLE>

**Date:** YYYY-MM-DD
**Severity:** P0 / P1 / P2
**IC:** @<handle>
**Author:** @<handle>
**Status:** Draft / Final
**Distribution:** internal / customer-facing summary

## Summary

One-paragraph executive summary. What broke, who was affected, how long.

## Timeline (UTC, ISO 8601)

- 2026-04-25T08:14:32Z — first symptom (Alertmanager fired `APIErrorRateHigh`).
- 2026-04-25T08:14:48Z — primary on-call paged.
- 2026-04-25T08:15:30Z — primary acknowledged; opened `#sev1-active`.
- ... (event by event) ...
- 2026-04-25T09:42:11Z — declared resolved; status page updated.

## Impact

- Number of affected tenants: N.
- Estimated affected requests: N.
- SLO error budget consumed: X% of monthly.
- Customer-facing duration: X minutes.

## Contributing factors

Root cause + at least one second-order:

- **Root:** ...
- **Second-order:** ...
- **Trigger:** ...

## What went well

- ...
- ...

## What went poorly

- ...
- ...

## Action items

| ID | Action | Owner | Due | Status |
|---|---|---|---|---|
| AI-01 | ... | @<handle> | YYYY-MM-DD | open |
| AI-02 | ... | @<handle> | YYYY-MM-DD | open |

Each action item links to a tracked issue.

## Public-facing summary

(Status page + customer email copy.)

## Risk register update

`RISK-NNN` reviewed; L×I re-scored to <new>.
```

Postmortems live in `docs/postmortems/YYYY-MM-DD-<slug>.md`. Public summaries on status page.

## 8. Quarterly tabletop drill

`@greenmetrics/secops` runs a tabletop scenario (P0 simulated). Outcomes captured in `docs/CHAOS-LOG.md`.

## 9. Anti-patterns rejected

- Postmortem witch-hunts — focus on systems, not people.
- Skipping postmortem because "we know what happened" — it's the action items that matter.
- Customer comms only after resolution — must update during active incident.
- Deleting forensic state to "speed up recovery" — never.
- Reusing the same IC + Tech Lead person for sustained P0 — burnout risk.
