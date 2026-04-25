# Backend Migrations Policy

**Doctrine refs:** Rule 21 (evolution), Rule 33 (data is the system).
**ADR:** `docs/adr/0005-migration-tool-pressly-goose.md`.
**Tool:** `pressly/goose` v3.

## Forward-only in production

- `goose up` only on production. Never `goose down` against a live DB without IC sign-off.
- Down scripts exist for testability and dev rollback only.

## File naming

```
backend/migrations/
  00001_init.up.sql
  00001_init.down.sql
  00002_hypertables.up.sql
  00002_hypertables.down.sql
  ...
  00006_rls_enable.go              # Go-coded for per-table policy templating
  00007_idempotency.up.sql
  00007_idempotency.down.sql
  ...
```

5-digit zero-padded ID; `snake_case` name matching the file's purpose.

## Goose annotations

Plain SQL:

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE meters ADD COLUMN cost_centre TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE meters DROP COLUMN cost_centre;
-- +goose StatementEnd
```

Timescale CAGG (cannot run inside a transaction):

```sql
-- +goose Up
-- +goose NO TRANSACTION
DROP MATERIALIZED VIEW IF EXISTS readings_15min;
CREATE MATERIALIZED VIEW readings_15min ...
WITH (timescaledb.continuous);

-- +goose Down
-- +goose NO TRANSACTION
DROP MATERIALIZED VIEW IF EXISTS readings_15min;
```

Go-coded:

```go
// backend/migrations/00006_rls_enable.go
package migrations

import (
    "context"
    "database/sql"
    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigrationContext(upRLSEnable, downRLSEnable)
}

var rlsTables = []string{
    "tenants", "users", "meters", "meter_channels", "readings",
    "reports", "alerts", "emission_factors", "audit_log", "idempotency_keys",
}

func upRLSEnable(ctx context.Context, tx *sql.Tx) error {
    for _, t := range rlsTables {
        if _, err := tx.ExecContext(ctx, "ALTER TABLE "+t+" ENABLE ROW LEVEL SECURITY"); err != nil { return err }
        if _, err := tx.ExecContext(ctx,
            "CREATE POLICY tenant_isolation ON "+t+
            " USING (tenant_id::text = current_setting('app.tenant_id', true))"+
            " WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true))"); err != nil { return err }
    }
    return nil
}

func downRLSEnable(ctx context.Context, tx *sql.Tx) error { /* drop policies, disable RLS */ return nil }
```

## CI gates

- `tests/migrations/up_down_test.go` runs every migration up, then down, then up again on a fresh testcontainer Timescale 16. Failure blocks PR.
- `forward-only-check` job verifies no committed `migrations/[0-9]+_*.{sql,go}` file already on `main` is modified in the PR.
- `migration-naming-check` enforces `NNNNN_<snake_case>.{up,down}.sql` or `NNNNN_<snake_case>.go`.

## Lock impact analysis

Every PR adding a migration includes a comment block in the PR description:

```
Lock impact:
  - readings: 0 ms (additive ADD COLUMN; nullable)
  - meters: ~50 ms (CREATE INDEX CONCURRENTLY for new column)
Estimated apply duration: < 1 s on staging-sized DB.
Maintenance window required: NO.
```

Hypertable / CAGG mods need maintenance window — see `docs/runbooks/db-outage.md`.

## Hypertable + CAGG + retention

- Never `DROP TABLE` a hypertable in `down` — operator uses `pg_dump`/`pg_restore`.
- New hypertable: same migration must include `compress_chunk` policy + `add_retention_policy`.
- CAGG schema change → drop + recreate → `CALL refresh_continuous_aggregate(...)` for historical window.
- Retention shortening → ADR + 30-day notice.

## Tenant-visible shape change

- Bump `schema_version` column in `tenants` snapshot exports.
- Update `api/openapi/v1.yaml` if response shape changes.
- Update CHANGELOG.
- Update `tests/contracts/v1_compat_test.go` baseline.

## Emergency rollback

See `docs/runbooks/db-outage.md`. Operator is authorised to take cluster read-only; no hard rollback without IC sign-off.

## Anti-patterns

- Editing applied migrations.
- DDL in handler code.
- Implicit migrations on backend boot (boot must require migrations already applied).
- Cross-cutting `BEGIN; many statements; COMMIT;` outside goose framework.
- ORM-driven migrations.
