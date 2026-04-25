# SecOps Runbook

**Owner:** `@greenmetrics/secops`.
**Doctrine refs:** Rule 49 (role), Rule 60 (incident response), Rule 67 (transparency).
**Review cadence:** quarterly.

This runbook documents the secops operating procedures so the role is operable by anyone joining `@greenmetrics/secops`.

## 1. Onboarding checklist

When a new operator joins `@greenmetrics/secops`:

- [ ] Add to `@greenmetrics/secops` GitHub team.
- [ ] Grant `Maintainer` on the repo (not Admin — preserve 2-person rule for branch protection changes).
- [ ] Add to PagerDuty rotation (`greenmetrics-secops` schedule).
- [ ] Provision AWS IAM user with break-glass IRSA assume rights only (no static keys).
- [ ] Pair on first PR review focused on `policies/`.
- [ ] Walk through `docs/THREAT-MODEL.md`, `docs/RISK-REGISTER.md`, `docs/INCIDENT-RESPONSE.md` together.
- [ ] Hand over physical YubiKey for Cosign root verification (quarterly).
- [ ] Add to Slack channels: `#greenmetrics-ops`, `#sev1-active`, `#secops`.
- [ ] Schedule first quarterly DevSecOps review attendance.

## 2. Daily routine

- Review Alertmanager dashboard: any unacked critical alerts? Page yourself if so.
- Review Falco events shipped to Loki: any unexplained anomalies? Triage.
- Check CI pipeline health on `main`: any flaky security jobs? Investigate, do not silence.
- Triage Dependabot PRs: auto-merge patch with green CI; review minor; manual ADR for major.

## 3. Weekly routine

- Saturday 03:00 UTC DAST sweep result review (`.github/workflows/dast.yml`).
- Dependabot weekly batch review: check for security-only PRs, fast-track them.
- SLO synthetic-probe report review: any availability dips?
- Review `gitleaks` allowlist: any new entries this week? Confirm rationale.

## 4. Monthly routine

- Last Friday: Chaos Game Day execution per `docs/CHAOS-PLAN.md`. Document outcome in `docs/CHAOS-LOG.md`.
- Cost audit: run `scripts/ops/cost-audit.sh`, review unused EBS / idle replicas / orphaned secrets.
- Risk register: any HIGH/CRITICAL changes since last month? Update L×I scores.
- Runbook freshness: SEV1 runbooks have `last_tested` within 90d? If not, run dry-run.

## 5. Quarterly routine

- DevSecOps review meeting: walk through risk register, chaos log, pentest report (if Q3), incident postmortems.
- JWT secret rotation: trigger `.github/workflows/jwt-rotation.yml`. Verify KID overlap window per `docs/JWT-ROTATION.md`.
- Threat model review: any new dependencies introducing new attack surfaces? Update `docs/THREAT-MODEL.md`.
- ADR audit: any review-date-stale ADRs? Revisit.
- Internal pentest engagement.
- Pre-commit framework SHA refresh (verify pins still resolve).

## 6. Annual routine

- External pentest engagement (Q3).
- DR full drill: region failover; restore from snapshot; measure RPO/RTO against `docs/SLO.md`.
- Regulatory citation re-verification: every citation in `docs/ITALIAN-COMPLIANCE.md` re-checked against primary source. Update access dates.
- License allowlist refresh (`LICENSES.allowed`).
- Cosign root key trust review: verify Sigstore Fulcio cert chain still rooted in expected CT log.
- `docs/COMPLIANCE/` evidence pack refresh.

## 7. Emergency procedures

### 7.1 Suspected secret leak

1. Rotate the secret immediately via AWS Secrets Manager rotation (CLI: `aws secretsmanager rotate-secret --secret-id ...`).
2. Force ESO refresh: `kubectl annotate externalsecret <name> -n greenmetrics force-sync=$(date +%s) --overwrite`.
3. Restart affected pods: `kubectl rollout restart deployment/<name> -n greenmetrics`.
4. Audit GitHub for the leaked secret: `gh search code "<partial-secret-prefix>"` (use partial; never echo full secret).
5. If leak in git history: rewrite history is a last resort; prefer secret rotation + revoke.
6. File postmortem within 72h.

### 7.2 Compromised CI / supply chain

1. Disable affected workflow: `gh workflow disable <name>`.
2. Pin Action ecosystem to known-good SHA snapshot.
3. Audit recent runs: `gh run list --workflow <name> --limit 50`.
4. Re-issue any signing certs touched: revoke OIDC trust to GitHub repo at AWS IAM identity provider.
5. Rotate all secrets exposed to that workflow.
6. File postmortem.

### 7.3 SEV1 production incident

Follow `docs/INCIDENT-RESPONSE.md`. Page on-call primary; switch to `#sev1-active` Slack; appoint incident commander; preserve forensic state (do not destroy logs); track timeline; close-loop with postmortem within 5 business days.

### 7.4 Cluster admission webhook offline

If Kyverno admission webhook is down (cluster cannot pull new images):

1. Check Kyverno deployment: `kubectl get deployment -n kyverno`.
2. If pods are crashlooping, escalate to platform-team (P2).
3. Do **not** disable webhook in production unless explicit incident commander order — disabling webhook removes admission gate (`Rule 54` violation).
4. If absolutely necessary, document in incident log + create immediate follow-up issue with priority `incident-followup`.

## 8. Decision-making default

When in doubt, default to the **safer** option (Rule 65 regulated-industry threshold):

- Block, don't allow.
- Verify, don't trust.
- Log, don't suppress.
- Page, don't silence.

When the safer option blocks legitimate work, the unblock requires an ADR + override-allowed label, not a silent waiver.
