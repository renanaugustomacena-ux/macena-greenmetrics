//go:build integration

// Up/down migration test — every migration applies, downs cleanly, re-applies on a fresh testcontainer.
//
// Doctrine refs: Rule 21, Rule 33, Rule 44.
// Plan ADR: docs/adr/0005-migration-tool-pressly-goose.md.
// CI gate: required for every PR touching backend/migrations/.

package migrations_test

import (
	"context"
	"database/sql"
	"testing"
)

func TestAllMigrationsApplyDownAndReApply(t *testing.T) {
	t.Skip("scaffold — implement when testcontainers fixture lands; spins TimescaleDB 16 fresh per suite")

	ctx := context.Background()
	db := setupTimescale(t, ctx)
	defer db.Close()

	// Phase 1: goose up to head.
	mustGoose(t, db, "up")
	verifyHeadVersion(t, db)

	// Phase 2: goose down by one.
	mustGoose(t, db, "down")

	// Phase 3: goose up again to head.
	mustGoose(t, db, "up")
	verifyHeadVersion(t, db)
}

func TestEveryMigrationDownIsReversible(t *testing.T) {
	t.Skip("scaffold — for each migration N: up to N, down N, up to N; assert state equivalence")
}

func TestRLSPolicyEnumeration(t *testing.T) {
	t.Skip("scaffold — after migrations, enumerate pg_class for tables with tenant_id column; assert each has rowsecurity=true + at least one policy in pg_policies")
}

func setupTimescale(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	t.Fatalf("not implemented; testcontainers fixture lands in S5 follow-on (tests/integration/postgres_setup.go)")
	return nil
}
func mustGoose(t *testing.T, db *sql.DB, cmd string) { t.Helper(); _ = db; _ = cmd }
func verifyHeadVersion(t *testing.T, db *sql.DB) { t.Helper(); _ = db }
