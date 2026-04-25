// Package repository — Tx wrappers + RLS-aware tenant context propagation.
//
// Doctrine refs: Rule 19, Rule 35 (consistency + transaction boundaries), Rule 39 (security as core).
// Plan ADR: docs/adr/0011-postgres-rls-defence-in-depth.md.
// Mitigates: RISK-007.
//
// Usage:
//
//   err := repo.InTxAsTenant(ctx, tenantID, func(tx pgx.Tx) error {
//       _, err := tx.Exec(ctx, "INSERT INTO meters ...")
//       return err
//   })
//
// `InTxAsTenant` calls `SET LOCAL app.tenant_id = $1` before invoking fn, so every
// subsequent statement evaluates the RLS policy `tenant_isolation` correctly.
// Hard-fail if tenantID is empty or invalid UUID.
//
// `InTx` is the non-tenant-scoped variant for ops paths (migrations are run via
// cmd/migrate which uses migration_user with BYPASSRLS).
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxFn is invoked inside a transaction. Returning an error rolls back; returning
// nil commits. The Tx must not be retained beyond fn return.
type TxFn func(pgx.Tx) error

// InTx executes fn inside a Tx with the given isolation level. Default isolation
// is ReadCommitted; pass pgx.TxOptions to override.
func InTx(ctx context.Context, pool *pgxpool.Pool, opts pgx.TxOptions, fn TxFn) error {
	tx, err := pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		// Rollback is a no-op after Commit; safe to defer unconditionally.
		_ = tx.Rollback(ctx)
	}()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// InTxAsTenant executes fn inside a Tx with `app.tenant_id` set to tenantID,
// so Postgres RLS policies evaluate correctly for every statement in fn.
//
// Hard-fail if tenantID is empty or not a parseable UUID; this is a structural
// safety check (Rule 39) — the application must never hit the DB without a
// known tenant context.
func InTxAsTenant(ctx context.Context, pool *pgxpool.Pool, tenantID string, fn TxFn) error {
	if tenantID == "" {
		return errors.New("tx: tenant id required (RLS context cannot be empty)")
	}
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("tx: invalid tenant id %q: %w", tenantID, err)
	}
	return InTx(ctx, pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		// SET LOCAL is scoped to the Tx; set_config returns the value as text but we discard it.
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", id.String()); err != nil {
			return fmt.Errorf("set_config app.tenant_id: %w", err)
		}
		return fn(tx)
	})
}

// InTxAsTenantWithIso lets the caller override isolation for report generators
// or other read-heavy paths needing RepeatableRead.
func InTxAsTenantWithIso(ctx context.Context, pool *pgxpool.Pool, tenantID string, iso pgx.TxIsoLevel, fn TxFn) error {
	if tenantID == "" {
		return errors.New("tx: tenant id required (RLS context cannot be empty)")
	}
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("tx: invalid tenant id %q: %w", tenantID, err)
	}
	return InTx(ctx, pool, pgx.TxOptions{IsoLevel: iso}, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", id.String()); err != nil {
			return fmt.Errorf("set_config app.tenant_id: %w", err)
		}
		return fn(tx)
	})
}
