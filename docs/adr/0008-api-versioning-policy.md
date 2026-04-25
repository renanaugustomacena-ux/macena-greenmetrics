# 0008 — API versioning policy

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 21, 34
**Review date:** 2027-04-25

## Context

API exists at `/api/v1/`. There is no documented versioning policy. Frontend `lib/api.ts` is typed against current API. Without a policy, breaking changes can ship without notice; clients (frontend, future mobile, integrators) have no contract on stability.

## Decision

SemVer 2.0.0 with major in URL path:

- **Major version in URL path** (`/api/v1`, `/api/v2`). Breaking change → new major.
- **Minor + patch in `info.version`** of `api/openapi/v1.yaml` and in `Server: greenmetrics-backend/X.Y.Z` response header.
- **`Sunset:` header per RFC 8594** on deprecated routes ≥ 90 days before removal.
- **`Deprecation:` header** per draft-ietf-httpapi-deprecation-header.
- **2-minor parallel-running window** for major migrations: `/api/v1` and `/api/v2` both served for ≥ 6 months.
- **CHANGELOG entry required** for every minor + major bump (PR template enforces).
- **Backward compat test** `tests/contracts/v1_compat_test.go` runs previous-release OpenAPI against current handlers via `kin-openapi`; fails on breaking change without major bump.

See `docs/API-VERSIONING.md` for full classification of breaking vs additive.

## Alternatives considered

- **Date-based versioning (Stripe-style 2026-04-25).** Rejected — clients pinning to date is unfamiliar to integrators; SemVer is the lingua franca.
- **Header-only versioning (`Accept: application/vnd.greenmetrics.v1+json`).** Rejected — URL versioning is debuggable from a browser; integrators expect URL-based.
- **No versioning, "we'll evolve gracefully".** Rejected — magical thinking; one breaking change = client breakage.

## Consequences

### Positive

- Predictable contract for integrators.
- Sunset header gives clients structured deprecation signal.
- Backward compat test catches breaking changes at PR time.
- v1 → v2 migration path documented; ≥ 6 month overlap = generous.

### Negative

- v1 → v2 means maintaining two code paths in `internal/api/v1/` and `internal/api/v2/` for ≥ 6 months.
- Sunset windows take patience; emergencies (security breakage in v1) require explicit ADR override.

### Neutral

- OpenAPI `deprecated: true` is the marker; `redocly lint` catches inconsistencies.

## Residual risks

- Forgotten Sunset/Deprecation header on a deprecated route. Mitigation: `redocly` rule fails PR; CI gate.
- Major migration drag — integrators stay on v1 forever. Mitigation: 6-month sunset is firm; communication via `#greenmetrics-platform` Slack + email + status page.

## References

- `docs/API-VERSIONING.md`.
- `api/openapi/v1.yaml`.
- ADR-0013 (oapi-codegen design-first).

## Tradeoff Stanza

- **Solves:** silent breaking changes; no contract on stability.
- **Optimises for:** predictable client experience; integrator trust; testable backward compat.
- **Sacrifices:** 6-month parallel-run effort on major migrations; CHANGELOG discipline.
- **Residual risks:** forgotten Sunset header (CI gate); migration drag (firm sunset windows).
