//go:build security && integration

// RLS isolation property test — random tenant pairs prove no cross-tenant access.
//
// Doctrine refs: Rule 19, Rule 39, Rule 65, Rule 44.
// Plan ADR: docs/adr/0011-postgres-rls-defence-in-depth.md.
// Mitigates: RISK-007.
//
// Run: `go test -tags="security integration" ./tests/security/...`
//
// Requires: testcontainers + Docker; spins TimescaleDB 16 fresh per suite,
// applies all migrations including 00006_rls_enable.go, exercises the
// `app_user` role.
//
// Deferred until S3 testcontainers wiring lands. This file is a scaffold:
// it defines the test surface but skips at the top (TODO renan).

package security_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestRLSCrossTenantAccessReturnsZero(t *testing.T) {
	t.Skip("scaffold — implement when testcontainers fixture lands in S3 (cf. tests/integration/postgres_setup.go)")

	ctx := context.Background()
	db := setupTimescale(t, ctx)
	defer db.Close()

	tenantA := "00000000-0000-4000-8000-00000000aaaa"
	tenantB := "00000000-0000-4000-8000-00000000bbbb"

	mustSetTenant(t, ctx, db, tenantA)
	insertMeter(t, ctx, db, tenantA, "00000000-0000-4000-8000-00000000a1a1")

	mustSetTenant(t, ctx, db, tenantB)
	rows := queryMeters(t, ctx, db)
	if rows != 0 {
		t.Errorf("tenant B saw %d meters; expected 0 (RLS leak)", rows)
	}
}

func TestRLSInsertWithMismatchedTenantRejects(t *testing.T) {
	t.Skip("scaffold — implement when testcontainers fixture lands in S3")

	ctx := context.Background()
	db := setupTimescale(t, ctx)
	defer db.Close()

	tenantA := "00000000-0000-4000-8000-00000000aaaa"
	tenantB := "00000000-0000-4000-8000-00000000bbbb"

	mustSetTenant(t, ctx, db, tenantA)
	_, err := db.ExecContext(ctx,
		`INSERT INTO meters (id, tenant_id, label, meter_type, protocol, unit, active) VALUES ($1, $2, 'x', 'electricity', 'modbus_tcp', 'kWh', true)`,
		"00000000-0000-4000-8000-00000000c1c1", tenantB)
	if err == nil {
		t.Fatal("expected RLS WITH CHECK to reject INSERT with mismatched tenant_id; got nil error")
	}
}

func TestRLSAuditLogIsAppendOnly(t *testing.T) {
	t.Skip("scaffold — implement when testcontainers fixture lands in S3")

	ctx := context.Background()
	db := setupTimescale(t, ctx)
	defer db.Close()

	tenantA := "00000000-0000-4000-8000-00000000aaaa"
	mustSetTenant(t, ctx, db, tenantA)

	// Insert OK.
	_, err := db.ExecContext(ctx,
		`INSERT INTO audit_log (id, tenant_id, action, target, created_at) VALUES (gen_random_uuid(), $1, 'test', '/x', now())`,
		tenantA)
	if err != nil {
		t.Fatalf("audit_log insert failed: %v", err)
	}

	// UPDATE should be rejected by trigger.
	_, err = db.ExecContext(ctx, `UPDATE audit_log SET action='tampered' WHERE tenant_id = $1`, tenantA)
	if err == nil {
		t.Fatal("audit_log UPDATE should have been rejected by trigger")
	}

	// DELETE should be rejected by trigger.
	_, err = db.ExecContext(ctx, `DELETE FROM audit_log WHERE tenant_id = $1`, tenantA)
	if err == nil {
		t.Fatal("audit_log DELETE should have been rejected by trigger")
	}
}

// --- testcontainer helpers (stub) -------------------------------------------

func setupTimescale(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	t.Fatalf("not implemented; lands in S3 with tests/integration/postgres_setup.go")
	return nil
}

func mustSetTenant(t *testing.T, ctx context.Context, db *sql.DB, tenant string) {
	t.Helper()
	if _, err := db.ExecContext(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenant); err != nil {
		t.Fatalf("set_config app.tenant_id: %v", err)
	}
}

func insertMeter(t *testing.T, ctx context.Context, db *sql.DB, tenant, id string) {
	t.Helper()
	_, err := db.ExecContext(ctx,
		`INSERT INTO meters (id, tenant_id, label, meter_type, protocol, unit, active) VALUES ($1, $2, 'x', 'electricity', 'modbus_tcp', 'kWh', true)`,
		id, tenant)
	if err != nil {
		t.Fatalf("insert meter: %v", err)
	}
}

func queryMeters(t *testing.T, ctx context.Context, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM meters`).Scan(&n); err != nil {
		t.Fatalf("count meters: %v", err)
	}
	return n
}
