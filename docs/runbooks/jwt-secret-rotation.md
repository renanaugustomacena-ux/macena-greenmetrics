---
title: JWT secret rotation (KID dual-key window)
severity: P2 (P0 on suspected leak)
mttd_target: 1h
mttr_target: 24h overlap window
owner: "@greenmetrics/secops"
related_alerts: [JWTSecretAgeExceeded, JWTVerifySlow]
last_tested: 2026-04-25
review_date: 2026-07-25
---

# Runbook — JWT secret rotation

**Doctrine refs:** Rule 19, Rule 39, Rule 56, Rule 62.
**ADR:** `docs/adr/0016-jwt-kid-rotation.md`. **Procedure detail:** `docs/JWT-ROTATION.md`.

## When to use

- Quarterly scheduled (cron `0 3 1 */3 *`).
- Suspected JWT secret leak (P0).
- Alertmanager `JWTSecretAgeExceeded` (90d cap).
- Compliance request (annual audit).

## Procedure (suspected leak — emergency P0)

```bash
# 1. Trigger rotation workflow with explicit reason.
gh workflow run jwt-rotation.yml -f reason="suspected leak: <description>"

# 2. Force ESO sync immediately.
kubectl annotate externalsecret -n greenmetrics greenmetrics-backend \
  force-sync="$(date +%s)" --overwrite

# 3. Restart all backend pods to pick up new kid.
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
kubectl rollout status deployment/greenmetrics-backend -n greenmetrics --timeout=180s

# 4. Verify new kid is current.
kubectl logs -n greenmetrics deploy/greenmetrics-backend --tail=50 \
  | jq 'select(.message=="jwt key reloaded") | {kid, kids_valid_count}'

# 5. Manually purge old kid from previous[] (suspected leak: skip 24h overlap).
aws secretsmanager get-secret-value --secret-id greenmetrics/prod/jwt --query SecretString --output text \
  | jq '.previous = []' \
  | aws secretsmanager put-secret-value --secret-id greenmetrics/prod/jwt --secret-string file:///dev/stdin

# 6. All in-flight tokens with old kid will reject; users see logout storm.
#    Status page comms: "We have rotated authentication credentials; please log in again."

# 7. File postmortem within 72h.
```

## Procedure (scheduled, no overlap purge)

The cron workflow handles steps 1–4; nothing manual required. Verify completion in `#greenmetrics-secops` Slack channel within 24h.

## Rollback

If new kid causes mass-validation failures (e.g. ESO sync failure leaves pods on old secret while new tokens are signed with new kid):

```bash
# Roll the secret back via AWS Secrets Manager versioning.
aws secretsmanager update-secret-version-stage \
  --secret-id greenmetrics/prod/jwt \
  --version-stage AWSCURRENT \
  --move-to-version-id <previous-version-id>
# Force ESO sync + pod restart (steps 2-3 above).
```

## Verification

- Integration test `backend/tests/security/jwt_rotation_test.go` covers overlap + expiry.
- `gm_jwt_verify_duration_seconds` p99 < 20 ms post-rotation.
- `aws_secretsmanager_secret_last_rotated_timestamp_seconds` recent.

## Anti-patterns

- Skip overlap window without leak reason — invalidates in-flight sessions unnecessarily.
- Manual rotation without GH workflow — no audit trail; Rule 56 violation.
- Rotate without verifying ESO sync — pods stuck on old secret while new tokens reject.
