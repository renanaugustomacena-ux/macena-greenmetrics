# Feature Flags

**Doctrine refs:** Rule 16 (scaling features), Rule 21 (evolution).
**ADR placeholder:** `docs/adr/00NN-feature-flags.md` (open when first non-trivial flag ships).

## When to use a flag

A flag is appropriate when one of the following is true:

- The feature is incomplete and must not reach all users until validated (canary, beta, A/B).
- The feature toggles a regulatory mode per tenant (e.g. CSRD large-enterprise vs SME).
- The feature toggles a high-risk path that must be reversible without redeploy (kill switch).
- The feature is gated by tenant-level configuration (per-plan capabilities).

A flag is **not** appropriate for:

- Permanent A/B variants — pick one and ship.
- Hiding a feature you don't trust — fix the test suite instead.
- Per-environment differences — that's `internal/config/`.
- Ad-hoc debug switches in handlers — use logging level instead.

## Library

Initial: in-process map driven by env var or per-tenant config (`internal/featureflags/`). No external SaaS (Rule 26 — tool sprawl).

```go
// internal/featureflags/flags.go (S3 to ship)
type Flag string

const (
    FlagAsyncReports     Flag = "async_reports"        // S4: async report generation
    FlagRLSEnforced      Flag = "rls_enforced"         // S3: Postgres RLS in addition to app filter
    FlagCircuitBreakers  Flag = "circuit_breakers"     // S4: gobreaker on outbound calls
    FlagPulseHmacOnly    Flag = "pulse_hmac_only"      // S3: drop legacy plaintext sig fallback
    FlagJWTKidRotation   Flag = "jwt_kid_rotation"     // S3: enforce kid claim
    FlagDastWeekly       Flag = "dast_weekly"          // S5: nuclei + ZAP weekly
)
```

## Per-tenant flags

Stored in `tenants.flags` JSONB column (S3 schema migration). Read via `tenant.HasFlag(flag)`.

## Sunset

Flags have an explicit lifecycle:

1. **Proposed** — ADR opened.
2. **Active** — flag in code, controllable via env / per-tenant.
3. **Default-on** — flag is on by default; opt-out path exists for rollback.
4. **Removed** — code path collapses; flag deleted; ADR superseded.

Flags older than **2 quarters in Active state** trigger a quarterly office-hours review: ship to default-on or remove.

## Anti-patterns rejected

- "Just leave the flag in forever" — flag accretion = code complexity. Sunset.
- Flags as config — config goes in env, not flags.
- Flag-driven branching deeper than 1 level — refactor into separate code paths.
- Cross-tenant defaults stored in code constants — store in DB.
