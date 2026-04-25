# JWT KID Rotation Procedure

**Owner:** `@greenmetrics/secops`.
**Doctrine refs:** Rule 19, Rule 39, Rule 62 (secrets discipline).
**Plan ADR:** `docs/adr/0016-jwt-kid-rotation.md`.
**Cron:** `.github/workflows/jwt-rotation.yml` quarterly (`0 3 1 */3 *`).
**Mitigates:** RISK-003 (JWT secret static).

## 1. Model

The backend JWT layer uses HS256 with a **`kid` (key ID) claim** in the token header. Multiple keys may be valid simultaneously (overlap window); only one key is the **current** signer at any time.

Storage:

- AWS Secrets Manager secret `greenmetrics/prod/jwt` JSON shape:

  ```json
  {
    "kid": "v3",
    "secret": "<base64 48 bytes of randomness>",
    "previous": [
      { "kid": "v2", "secret": "<base64 48 bytes>", "expires_at": "2026-04-26T03:00:00Z" }
    ]
  }
  ```

- ESO syncs this secret to K8s Secret `greenmetrics-backend-secrets` keys `JWT_SECRET` (current), `JWT_KID_CURRENT` (kid), `JWT_KIDS_VALID` (JSON array of `{kid, secret, expires_at}`).

## 2. Backend behaviour

- **Sign:** always with `kid = JWT_KID_CURRENT`. Header includes `"kid": "v3"`.
- **Validate:** look up incoming `kid` in `JWT_KIDS_VALID`; reject if expired or absent.
- **Reject `alg:none`** and any non-HS256 (`jwt.WithValidMethods([]string{"HS256"})`).
- **Boot refusal** (Rule 39): `cfg.JWTKidsValid` must be non-empty in production, the `current` kid must be present, and every `secret` must be ≥ 32 bytes.

## 3. Quarterly rotation flow

```
Day 0:  Workflow runs.
        - aws secretsmanager get-secret-value --secret-id greenmetrics/prod/jwt
        - parse current JSON
        - generate new secret (48 random bytes)
        - bump kid: v3 → v4
        - move v3 to previous[] with expires_at = now + 24h
        - aws secretsmanager put-secret-value
Day 0:  ESO refreshes within 1h → K8s Secret updated.
Day 0:  Backend pods detect Secret change via projected-volume reload (S3) or rolling restart.
Day 0:  Tokens issued with new kid v4. Old tokens (kid v3) still validate until expiry of last v3 token.
Day 1:  Workflow scheduled to remove v3 from previous[] (24h overlap window > max session 12h + buffer).
```

## 4. Manual rotation (emergency)

```bash
# Suspected JWT secret leak.
gh workflow run jwt-rotation.yml -f reason="suspected leak"

# After ESO sync (≤ 1h), verify pods picked up new kid:
kubectl logs -n greenmetrics deploy/greenmetrics-backend | jq 'select(.message=="jwt key reloaded") | {kid, kids_valid_count}'

# Force restart if pods don't reload:
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
```

## 5. Verification

- Integration test `backend/tests/security/jwt_rotation_test.go` (S3) covers:
  - Token signed with `kid=v3` validates during overlap.
  - Token signed with `kid=v3` rejects after `expires_at`.
  - Token signed with unknown `kid` rejects always.
  - Boot refuses if `JWT_KID_CURRENT` not in `JWT_KIDS_VALID`.

- Alertmanager rule (`monitoring/prometheus/rules/security.rules.yaml`):

  ```promql
  # Page if rotation hasn't run in > 90 days.
  (time() - aws_secretsmanager_secret_last_rotated_timestamp_seconds{secret="greenmetrics/prod/jwt"}) > 7776000
  ```

## 6. Related

- `terraform/modules/iam-irsa/` — `secops-rotate-irsa` role assumed by the workflow.
- `terraform/modules/secrets/` — Secrets Manager secret + AWS-native rotation Lambda config.
- `gitops/base/external-secrets/externalsecret-backend.yaml` — ESO mapping.
- `docs/runbooks/jwt-secret-rotation.md` (S4) — operator procedure for incident-driven manual rotation.

## 7. Anti-patterns rejected

- "Rotate yearly" — too long; 90 days is regulatory norm for high-risk secrets.
- "Big-bang rotate without overlap" — invalidates in-flight sessions; users see logout storm.
- "Custom KMS keys instead of HMAC secret" — works but adds key-management burden; HS256 + rotation suffices for our threat model.
- "Rotate manually when needed" — Rule 56 (automation over process); cron + workflow is the rule.
