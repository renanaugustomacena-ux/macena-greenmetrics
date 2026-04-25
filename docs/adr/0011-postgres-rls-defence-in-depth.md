# 0011 — Postgres RLS as defence in depth

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 19, 35, 39, 65
**Review date:** 2027-04-25
**Mitigates:** RISK-007, RISK-009.

## Context

See ADR-0002 (multi-tenant RLS strategy) for the policy decision. This ADR is the **how**: the in-code mechanics + test guarantees + operator escape hatches.

## Decision

- `backend/internal/repository/tx.go` provides `InTxAsTenant(ctx, tenantID, fn)` — wrapper that issues `SELECT set_config('app.tenant_id', $1, true)` at Tx start. Hard-fail if tenantID empty or invalid UUID.
- Every repository write (and every read that should be RLS-scoped) goes through `InTxAsTenant`; CI lint flags any direct `pool.Exec` / `pool.Query` outside `repository/`.
- `app_user` PG role created in migration `00006_rls_enable.go`; `BYPASSRLS=false`; `FORCE ROW LEVEL SECURITY` on every tenant-scoped table.
- `migration_user` (used only by `cmd/migrate`) retains `BYPASSRLS` for ops paths (CAGG creation, retention policy edits).
- Property test `tests/security/rls_isolation_test.go` (S3) generates random tenant pairs; asserts no cross-tenant access via SELECT/UPDATE/DELETE/INSERT (with mismatched tenant_id).
- Failure mode: missed `InTxAsTenant` → `current_setting('app.tenant_id', true)` returns empty → RLS policy rejects → query returns 0 rows for SELECT, 42501 permission denied for INSERT/UPDATE.

## Alternatives considered

- **App-level filters only** (status quo) — see ADR-0002 alternatives.
- **`SECURITY DEFINER` functions for queries** — rejected; SECURITY DEFINER is the documented bypass vector.
- **Custom PG GUC + per-query parameter** — `set_config(..., true)` is the standard; reinventing is Rule 26.

## Consequences

### Positive

- Cross-tenant data leak via SQL bug becomes a 0-row query, not a leak.
- Audit log immutability enforced at DB level (migration `00010_audit_lock`).
- New table is RLS-protected by adding to `rlsTables` list in `00006_rls_enable.go` — single point of policy.
- Property test catches regression in policy enforcement before merge.

### Negative

- Every Tx must use `InTxAsTenant` — ergonomic friction.
- Per-Tx `set_config` round-trip overhead (~1 ms over the wire on local socket; ~3–5 ms on RDS).
- `BYPASSRLS` migration path is a privileged seam — must be guarded.
- Test infra (testcontainers) needs a non-superuser app_user role configured per-suite.

### Neutral

- pgx driver supports `set_config` natively; no library change.
- Tx isolation default is ReadCommitted; RLS evaluation does not change per isolation level.

## Residual risks

- `SECURITY DEFINER` functions added later silently bypass RLS. Mitigation: grep gate in CI; code review checklist; PG audit extension logging when ops use BYPASSRLS.
- `migration_user` credentials stolen → unrestricted DB access. Mitigation: BYPASSRLS role assumed only by migration Job, not by app pods; CloudTrail alarm on `migration_user` connection from non-Job IP.
- Forgotten RLS on a new table. Mitigation: migration test enumerates `pg_class` for `tenant_id` columns and asserts each table has both `rowsecurity=true` and at least one policy.

## References

- ADR-0002 (multi-tenant RLS strategy).
- `backend/migrations/00006_rls_enable.go`.
- `backend/migrations/00010_audit_lock.up.sql`.
- `backend/internal/repository/tx.go`.
- `backend/tests/security/rls_isolation_test.go` (S3).

## Tradeoff Stanza

- **Solves:** structural enforcement of multi-tenant isolation at the DB layer; audit log tampering at any layer.
- **Optimises for:** defence in depth; regulator-grade defensibility.
- **Sacrifices:** ~5–15% query overhead; mandatory `InTxAsTenant` discipline; non-superuser test fixture cost.
- **Residual risks:** SECURITY DEFINER bypass (CI grep + audit); BYPASSRLS role compromise (CloudTrail alarm); forgotten table on new schema (migration enumeration test).
