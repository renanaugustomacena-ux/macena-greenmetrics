# 0016 — JWT KID rotation

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 19, 39, 56, 62
**Review date:** 2027-04-25
**Mitigates:** RISK-003.

## Context

Today the JWT secret is a single static value held in `cfg.JWTSecret`, sourced from env (`JWT_SECRET`). HS256 is pinned at validation (`internal/handlers/auth.go:179, 236`). There is no rotation procedure. A leak of the JWT secret today requires a hard rollover that invalidates every in-flight session — operationally hostile and rarely done. Result: the secret stays static, RISK-003 stays open.

## Decision

Adopt JWT `kid` (key ID) header claim with **dual-key overlap window**.

- Backend reads `JWT_KID_CURRENT` (string) and `JWT_KIDS_VALID` (JSON array of `{kid, secret, expires_at}`) from K8s Secret materialised by ESO from AWS Secrets Manager `greenmetrics/prod/jwt`.
- Sign tokens with `kid = JWT_KID_CURRENT`; HS256 always.
- Validate tokens by looking up incoming `kid` in `JWT_KIDS_VALID`; reject if expired or absent; reject `alg:none`; reject non-HS256.
- Quarterly rotation via cron workflow `.github/workflows/jwt-rotation.yml` (`0 3 1 */3 *`):
  1. Generate new 48-byte random secret.
  2. Bump `kid` (e.g. `v3 → v4`).
  3. Move old kid into `previous[]` with `expires_at = now + 24h` (overlap window > max session 12h + buffer).
  4. `aws secretsmanager put-secret-value`.
  5. ESO syncs within 1h; pods reload.
  6. After 24h: GH Action prunes expired kids.
- Boot refusal in production if `JWT_KID_CURRENT` not in `JWT_KIDS_VALID` or any secret < 32 bytes.
- Manual rotation (emergency) via `gh workflow run jwt-rotation.yml -f reason="suspected leak"`.

## Alternatives considered

- **Single static secret.** Status quo. Rejected — RISK-003 score 8 (L=2, I=4); a leak forces user-visible logout storm to recover.
- **JWKS with public-key algorithm (RS256/ES256).** Rejected — requires distributing private keys to signing service; HS256 + ESO + KID rotation is simpler; revisit if external IdP appears in roadmap.
- **Single rotation without overlap.** Rejected — invalidates in-flight sessions; users see logout storm.
- **Rotation cadence yearly.** Rejected — too long; 90 days is regulatory norm for high-risk secrets.
- **Manual rotation only.** Rejected per Rule 56 (automation over process).

## Consequences

### Positive

- Quarterly rotation routine; secret leak recovery is "rotate, wait 24h, done" — no hard logout storm.
- Operator can trigger emergency rotation in seconds via workflow_dispatch.
- Boot refusal prevents stale-secret deploy.
- Integration test (`backend/tests/security/jwt_rotation_test.go`) covers overlap + expiry.

### Negative

- `kid`-aware validation requires a non-trivial backend code change to `JWTMiddleware` (lines 215-255 today): replace single `[]byte(d.Config.JWTSecret)` with kid lookup.
- ESO sync delay (≤ 1h) means rotation propagation lag; emergency rotation requires manual force-sync.
- Workflow needs IAM role (`secops-rotate-irsa`) with `secretsmanager:PutSecretValue` on the JWT secret.

### Neutral

- HS256 stays — no migration to RS256 needed.
- KID claim is a standard JWT header field; clients should not interpret it.

## Residual risks

- Rotation workflow failure leaves stale secret. Mitigation: Alertmanager rule on `aws_secret_age_seconds > 7776000` (90d).
- All kids leaked simultaneously. Mitigation: rotation and KID change is independent of secret rotation; emergency rotation purges all `previous[]` immediately.
- 24h overlap window > session TTL of any user with custom long session. Mitigation: enforce `SESSION_ABSOLUTE_HOURS=12` cap; no session > 12h.
- Workflow OIDC trust compromise. Mitigation: trust policy bound to repo + ref + environment (production-secrets) (terraform/modules/github-oidc).

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.C.62.
- `docs/JWT-ROTATION.md`.
- `.github/workflows/jwt-rotation.yml`.
- ADR-0003 (ESO).
- `RISK-003`.

## Tradeoff Stanza

- **Solves:** static long-lived JWT secret; secret leak recovery requires user-visible logout storm.
- **Optimises for:** routine rotation; emergency rotation in minutes; zero user-visible impact during rotation.
- **Sacrifices:** non-trivial backend code change to JWTMiddleware; IAM role for rotation workflow; ESO sync lag.
- **Residual risks:** rotation workflow failure (age alert); all-kids leak (rotation + purge); session > overlap (cap session); OIDC trust compromise (scoped trust policy).
