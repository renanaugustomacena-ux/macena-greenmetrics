# 0005 — Migration tool: pressly/goose

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 21 (evolution), 33 (data is the system)
**Review date:** 2027-04-25

## Context

GreenMetrics ships 5 raw SQL migrations applied via `psql` shell-out from `backend/Makefile:40-45`. There is no version table, no rollback, no idempotency guard beyond `IF NOT EXISTS`. TimescaleDB CAGGs and retention policies are non-trivial objects — CAGGs cannot be `CREATE OR REPLACE`d; retention policies are jobs, not DDL — so production rollback paths require Go-coded migrations rather than pure SQL.

## Decision

Adopt `pressly/goose` v3 as the migration runner. Add `cmd/migrate/main.go` that invokes `goose.Up`/`Down`. Restructure `backend/migrations/` into `XXXXX_<name>.up.sql` / `.down.sql` pairs. Add `00006_rls_enable.go` as a Go-coded migration for templating per-table RLS policies. Track applied migrations in the goose `schema_migrations` table.

## Alternatives considered

- **`golang-migrate/migrate`.** Rejected because Go-coded migrations are second-class — needed for the Timescale CAGG drop+recreate dance and the per-table RLS policy templating in S3. golang-migrate is more mature and has a richer driver matrix, but the Go-migration ergonomics matter more for our workload.
- **`tern` (jackc).** Rejected because adoption is narrower; goose has a larger community, more documentation, and is widely used in Timescale-shaped projects.
- **Sqitch.** Rejected because tooling sprawl — adds Perl runtime to the stack.
- **DIY Go migration runner.** Rejected per Rule 13 (abstraction cost) and Rule 26 (no NIH).

## Consequences

### Positive

- Forward-only enforcement: `goose status` reveals applied + pending; CI fails PR that edits an applied migration.
- Rollback path: every `.up.sql` has a paired `.down.sql`; integration test `tests/migrations/up_down_test.go` runs both directions on every PR.
- Go-coded migrations enabled — RLS migration `00006_rls_enable.go` templates per-table policies from a generated list.
- Single binary `cmd/migrate` replaces shell-out; deployable as a K8s Job pre-deploy.

### Negative

- Doubles migration file count (`.up.sql` + `.down.sql`).
- Down scripts may be lossy (`DROP COLUMN` discards data); operator must accept and document.
- `goose` is a single-author maintained project — review at 2027.

### Neutral

- Migration policy doc lives in `docs/backend/migrations.md` (lands S2).

## Residual risks

- CAGG migrations on populated hypertables take minutes — operator must schedule a maintenance window. Documented in `docs/runbooks/db-outage.md` (lands S4).
- Migration runner crash mid-application leaves DB in inconsistent state — goose tracks per-migration so re-run is safe; documented in operator runbook.
- Major DB rollbacks rely on `pg_dump`/`pg_restore` (DR runbook), not migration `down` scripts.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.B.33.
- `https://github.com/pressly/goose`.
- `docs/backend/migrations.md` (lands S2).
- `RISK-018` — Terraform state on local FS (closes S2 separately).

## Tradeoff Stanza

- **Solves:** ad-hoc DB drift, no rollback path, no version tracking.
- **Optimises for:** rollback safety, Timescale-aware migration ergonomics, single-binary deploy.
- **Sacrifices:** 2× SQL files, a single-author dependency, occasional lossy down scripts.
- **Residual risks:** CAGG migrations need maintenance windows; major DB rollbacks rely on dump/restore; goose project sustainability — review at 2027.
