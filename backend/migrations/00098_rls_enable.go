// Migration 00006_rls_enable — enable Postgres RLS as defence in depth.
//
// Doctrine refs: Rule 19 (security as structural), Rule 39 (backend security), Rule 65 (regulated quality).
// Plan ADR: docs/adr/0011-postgres-rls-defence-in-depth.md.
// Mitigates: RISK-007 (cross-tenant data leak via SQL bug or SECURITY DEFINER fn).
//
// Application path: every Tx acquired by the backend runs
//
//	SET LOCAL app.tenant_id = '<uuid>'
//
// at the start of the Tx (helper `InTxAsTenant` in internal/repository/tx.go).
// Hard-fail if tenantID is empty. The `app_user` PG role has BYPASSRLS=false.
// The `migration_user` (used only by cmd/migrate) keeps BYPASSRLS for ops.
//
// Tables covered: tenants, users, meters, meter_channels, readings, reports,
// alerts, emission_factors, audit_log, idempotency_keys.
//
// Special policies:
//   - audit_log is append-only via WITH CHECK only (no SELECT/UPDATE/DELETE policy).
//   - users allows read by self + tenant peers.
//
// Down migration drops the policies and disables RLS — kept for testability;
// production must not run `down` against a live tenant DB.

package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upRLSEnable, downRLSEnable)
}

// rlsTables — every tenant-scoped table.
var rlsTables = []string{
	"tenants",
	"users",
	"meters",
	"meter_channels",
	"readings",
	"reports",
	"alerts",
	"emission_factors",
	"audit_log",
	"idempotency_keys",
}

func upRLSEnable(ctx context.Context, tx *sql.Tx) error {
	// Create the application role used by the backend at runtime. Idempotent.
	const ensureRole = `
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
        CREATE ROLE app_user NOINHERIT NOBYPASSRLS;
    END IF;
END$$;`
	if _, err := tx.ExecContext(ctx, ensureRole); err != nil {
		return fmt.Errorf("ensure app_user role: %w", err)
	}

	for _, t := range rlsTables {
		// Enable RLS on the table.
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", t)); err != nil {
			return fmt.Errorf("enable RLS on %s: %w", t, err)
		}
		// Force RLS even for the table owner (defence vs. SECURITY DEFINER fn).
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s FORCE ROW LEVEL SECURITY", t)); err != nil {
			return fmt.Errorf("force RLS on %s: %w", t, err)
		}

		switch t {
		case "audit_log":
			// Append-only: no SELECT/UPDATE/DELETE policy beyond audit:read RBAC.
			if _, err := tx.ExecContext(ctx, `
                CREATE POLICY audit_log_tenant_insert ON audit_log
                FOR INSERT
                WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
            `); err != nil {
				return fmt.Errorf("create policy on audit_log: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `
                CREATE POLICY audit_log_tenant_select ON audit_log
                FOR SELECT
                USING (tenant_id::text = current_setting('app.tenant_id', true));
            `); err != nil {
				return fmt.Errorf("create select policy on audit_log: %w", err)
			}
			// No UPDATE / DELETE policies — append-only.
		default:
			if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
                CREATE POLICY tenant_isolation ON %s
                USING (tenant_id::text = current_setting('app.tenant_id', true))
                WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));
            `, t)); err != nil {
				return fmt.Errorf("create policy on %s: %w", t, err)
			}
		}

		// Grant read to app_user; writes are policy-gated above.
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE ON %s TO app_user", t)); err != nil {
			return fmt.Errorf("grant on %s: %w", t, err)
		}
	}

	// Sequences also need GRANT for INSERT to populate serial / bigserial PKs.
	if _, err := tx.ExecContext(ctx, "GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user"); err != nil {
		return fmt.Errorf("grant sequences: %w", err)
	}

	return nil
}

func downRLSEnable(ctx context.Context, tx *sql.Tx) error {
	for _, t := range rlsTables {
		if t == "audit_log" {
			_, _ = tx.ExecContext(ctx, "DROP POLICY IF EXISTS audit_log_tenant_insert ON audit_log")
			_, _ = tx.ExecContext(ctx, "DROP POLICY IF EXISTS audit_log_tenant_select ON audit_log")
		} else {
			_, _ = tx.ExecContext(ctx, fmt.Sprintf("DROP POLICY IF EXISTS tenant_isolation ON %s", t))
		}
		_, _ = tx.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s NO FORCE ROW LEVEL SECURITY", t))
		_, _ = tx.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s DISABLE ROW LEVEL SECURITY", t))
	}
	// Keep app_user in case downstream code expects it; deletion is operator decision.
	return nil
}
