# 0002 — Multi-tenant RLS strategy

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 19, 39, 65
**Review date:** 2027-04-25
**Mitigates:** RISK-007, RISK-009.

## Context

GreenMetrics is multi-tenant. Today tenant isolation is enforced **only** at the application layer: JWT middleware extracts `tenant_id`, every repository query carries `WHERE tenant_id = $1`. One missed `WHERE` is a regulator-grade data leak (`RISK-007`, score 10). Audit log is also tenant-scoped but not append-only at the DB level (`RISK-009`).

## Decision

Enforce Postgres Row-Level Security (RLS) on every tenant-scoped table as **defence in depth**, in addition to the application-level filter — not as a replacement.

- Migration `backend/migrations/00006_rls_enable.go` enables RLS + policies on: `tenants`, `users`, `meters`, `meter_channels`, `readings`, `reports`, `alerts`, `emission_factors`, `audit_log`, `idempotency_keys`.
- Policy: `USING (tenant_id::text = current_setting('app.tenant_id', true))` with matching `WITH CHECK`.
- App role `app_user` with `BYPASSRLS=false`. Migration role `migration_user` keeps `BYPASSRLS` for ops paths.
- `FORCE ROW LEVEL SECURITY` set so even table-owner sessions evaluate the policy (defence vs. SECURITY DEFINER functions).
- Application connections set `app.tenant_id` per-Tx via `repository.InTxAsTenant(ctx, tenantID, fn)` (`backend/internal/repository/tx.go`); empty tenantID is hard-failed by the wrapper.
- Audit log gets a separate INSERT-only policy + tx-trigger preventing UPDATE/DELETE (migration `00010_audit_lock`).

## Alternatives considered

- **Application filters only.** Status quo. Rejected — one bug = leak; doctrine Rule 39 requires defence in depth.
- **Per-tenant database (one PG schema per tenant, or one DB per tenant).** Rejected — operational overhead at 50+ tenants is unmanageable; backup, monitoring, Timescale CAGG topology multiply.
- **Per-tenant table partitioning.** Rejected — Timescale already partitions by time; tenant partitioning collides and complicates CAGG semantics. Revisit only at stretch profile (ADR-010 trigger).
- **Application-level encryption per tenant.** Rejected — masks data at rest but does nothing for query-time isolation; adds operational load.
- **External authorisation policy engine (OPA in request path).** Rejected per REJ-10 (OPA-everywhere) — not the right shape for row-level filtering.

## Consequences

### Positive

- A bug in any handler's tenant_id extraction can no longer leak data — DB returns 0 rows (or 42501 permission denied for INSERT/UPDATE that violates `WITH CHECK`).
- Audit log is tamper-evident at the DB layer: `audit_log_no_update` and `audit_log_no_delete` triggers raise if even a privileged role attempts mutation.
- Property-tested in `tests/security/rls_isolation_test.go` — random tenant pairs prove no cross-tenant access.

### Negative

- ~5–15% query overhead from RLS evaluation per row (measured in `tests/bench/repo_insert_bench_test.go`, S5).
- Every Tx must call `InTxAsTenant` — operator pattern; CI grep gate enforces (S5).
- Test infrastructure must run as `app_user`, not `postgres`, to exercise policies — testcontainers config updated.
- `BYPASSRLS` paths (migrations, ops queries) are an attack surface — restricted to `migration_user` + audited via CloudTrail equivalent (PG audit extension if enabled).

### Neutral

- App-level filter remains in code as defence in depth — the two are not redundant; they catch different bug classes.

## Residual risks

- RLS bypass via `SECURITY DEFINER` PL/pgSQL functions executed as table owner — mitigation: code-review gate forbids `SECURITY DEFINER` outside migration files; static check via grep.
- RLS bypass via direct PG superuser session — mitigation: superuser access only via break-glass IAM (RISK-006); CloudTrail alarm.
- New table missed in `rlsTables` list — mitigation: migration test enumerates `pg_class` and asserts every tenant_id-bearing table has a policy.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.A.19, §6.B.39.
- `backend/migrations/00006_rls_enable.go`.
- `backend/internal/repository/tx.go`.
- `backend/internal/security/rbac.go`.
- `RISK-007`, `RISK-009`.

## Tradeoff Stanza

- **Solves:** cross-tenant data leak via SQL bug; audit log tampering by privileged user.
- **Optimises for:** defence in depth; regulatory defensibility; explicit tenant context flow.
- **Sacrifices:** 5–15% query overhead; mandatory `InTxAsTenant` discipline; testcontainer DB role complication.
- **Residual risks:** SECURITY DEFINER function bypass (code-review gate); break-glass superuser sessions (CloudTrail alarm); new tables missed in policy list (migration test enumerates `pg_class`).
