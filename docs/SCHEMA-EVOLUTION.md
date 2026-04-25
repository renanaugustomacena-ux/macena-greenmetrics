# Schema Evolution Policy

**Doctrine refs:** Rule 21 (evolution + change management), Rule 33 (data is the system).
**Owners:** `@greenmetrics/data-team`, `@greenmetrics/platform-team`.

Data outlives code (Rule 33). Schema changes are the most expensive failures to undo. This policy is binding for any change in `backend/migrations/`.

## 1. Forward-only invariant

- New schema → new migration file pair (`XXXXX_<name>.up.sql` / `.down.sql`).
- **Never edit an applied migration.** CI rule fails any PR that modifies `backend/migrations/[0-9]+_*.sql` files already on `main`.
- If a migration was applied and discovered wrong: write a corrective forward migration with a clear name (`XXXXX_fix_<original>.up.sql`).

## 2. Goose conventions

```sql
-- +goose Up
-- +goose StatementBegin
-- DDL here.
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Reverse DDL.
-- +goose StatementEnd
```

For migrations spanning multiple statements that must run in a transaction, wrap in `StatementBegin`/`StatementEnd`. For Timescale CAGG operations, **disable transactions** (`-- +goose NO TRANSACTION`) because CAGG creation cannot run inside a transaction block.

## 3. Hot-path rules

- **Additive only on `readings`.** No `ALTER COLUMN TYPE`. No `DROP COLUMN`. New columns must be `NULL` or have `DEFAULT`.
- **No reordering.** Append new columns at the end.
- **No `ALTER TABLE` that rewrites the table** (anything beyond metadata changes locks the hypertable).

## 4. Hypertable rules

- Never `DROP TABLE` a hypertable in `down` migrations — operator must use `pg_dump`/`pg_restore` for major rollbacks.
- Add chunk-time interval changes only when validated against capacity model.
- New hypertable: include `compress_chunk` policy + `add_retention_policy` in the same migration.

## 5. Continuous aggregate (CAGG) rules

- CAGG creation **cannot** run in a transaction → use `-- +goose NO TRANSACTION`.
- CAGG cannot be `CREATE OR REPLACE`d → drop + recreate; existing CAGG data is regenerated lazily on next refresh.
- CAGG drop+recreate on populated hypertables can take minutes — schedule maintenance window per `docs/runbooks/db-outage.md`.
- After CAGG schema change, run `CALL refresh_continuous_aggregate(...)` for the historical window.

## 6. Retention policy rules

- Retention policies are jobs (`add_retention_policy`), not DDL — change via `remove_retention_policy` + `add_retention_policy`.
- Retention shortening risks data loss — require ADR + 30-day notice.
- Retention lengthening: bumps storage cost; check capacity model.

## 7. Tenant-visible shape changes

- Bump `schema_version` column in `tenants` snapshot exports.
- Update `api/openapi/v1.yaml` if response shape changes.
- Update CHANGELOG.
- Update `tests/contracts/v1_compat_test.go` baseline.

## 8. Down script policy

- Down scripts may be lossy — operator must accept and document.
- Lossy down scripts (e.g. `DROP COLUMN`, `DROP TABLE`) require a comment block at the top of the down section explicitly stating data loss.

## 9. Migration safety check

PR adding a migration must include:

- `tests/migrations/up_down_test.go` runs the new migration up, then down, then up again on a fresh testcontainer Timescale 16 — green.
- Lock impact analysis comment in the migration: "expected lock duration ~X seconds against Y rows".
- Rollback plan in the PR description.

## 10. Emergency rollback path

For migration emergencies that goose `down` cannot solve:

1. Page on-call DBA per `docs/runbooks/db-outage.md`.
2. Take cluster read-only.
3. `pg_dump --schema-only` for current state vs target state diff.
4. Manual SQL surgery, witnessed by 2 operators.
5. Postmortem within 5 business days.

## 11. Anti-patterns rejected

- Editing applied migrations.
- DDL in handler code at runtime.
- Implicit migrations on backend boot.
- Cross-cutting `BEGIN; many statements; COMMIT;` outside goose framework.
- ORM-driven migrations (REJ-35).
