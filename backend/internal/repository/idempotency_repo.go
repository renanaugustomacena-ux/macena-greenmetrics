// Package repository — pgx-backed implementation of the IdempotencyStore
// interface declared in internal/api/v1/idempotency.go.
//
// Doctrine refs: Rule 14, Rule 35 (idempotency), Rule 39, Rule 65.
// Mitigates: ingest duplicates from retry storms.
//
// Storage: `idempotency_keys` table from migration 00007_idempotency.sql.
// Plain table (NOT a hypertable) — see migration design note. Cleanup is
// out-of-band via scripts/ops/idempotency-gc.sh (S5 follow-on).
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apiv1 "github.com/greenmetrics/backend/internal/api/v1"
)

// IdempotencyRepo persists idempotency replay state in Postgres.
type IdempotencyRepo struct {
	pool *pgxpool.Pool
}

// NewIdempotencyRepo constructs the repo. Caller passes the existing pgx pool
// (reuse the one from TimescaleRepository to avoid a second connection budget).
func NewIdempotencyRepo(pool *pgxpool.Pool) *IdempotencyRepo {
	return &IdempotencyRepo{pool: pool}
}

// Get implements apiv1.IdempotencyStore. Returns apiv1.ErrIdempotencyMiss on
// no rows; any other error surfaces to the middleware as a 500 (so middleware
// can decide to fail open vs. closed).
func (r *IdempotencyRepo) Get(ctx context.Context, tenantID uuid.UUID, key string) (*apiv1.IdempotencyRecord, error) {
	rec := &apiv1.IdempotencyRecord{TenantID: tenantID, Key: key, Headers: map[string]string{}}

	var hashBytes []byte
	var headersJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT request_hash, response_status, response_body, response_headers
		FROM idempotency_keys
		WHERE tenant_id = $1 AND key = $2
	`, tenantID, key).Scan(&hashBytes, &rec.Status, &rec.Body, &headersJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apiv1.ErrIdempotencyMiss
	}
	if err != nil {
		return nil, fmt.Errorf("idempotency get: %w", err)
	}

	if len(hashBytes) != len(rec.RequestHash) {
		return nil, fmt.Errorf("idempotency get: stored hash size %d != expected %d", len(hashBytes), len(rec.RequestHash))
	}
	copy(rec.RequestHash[:], hashBytes)

	if len(headersJSON) > 0 {
		if err := json.Unmarshal(headersJSON, &rec.Headers); err != nil {
			return nil, fmt.Errorf("idempotency get: unmarshal headers: %w", err)
		}
	}
	return rec, nil
}

// Put implements apiv1.IdempotencyStore. INSERT ... ON CONFLICT DO NOTHING so
// concurrent middleware invocations on the same (tenant, key) are race-safe;
// the second writer's Put becomes a no-op and the first writer's response wins
// the replay slot.
func (r *IdempotencyRepo) Put(ctx context.Context, rec *apiv1.IdempotencyRecord) error {
	if rec == nil {
		return errors.New("idempotency put: nil record")
	}
	headersJSON, err := json.Marshal(rec.Headers)
	if err != nil {
		return fmt.Errorf("idempotency put: marshal headers: %w", err)
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO idempotency_keys
			(tenant_id, key, request_hash, response_status, response_body, response_headers, created_at)
		VALUES
			($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (tenant_id, key) DO NOTHING
	`, rec.TenantID, rec.Key, rec.RequestHash[:], rec.Status, rec.Body, headersJSON)
	if err != nil {
		return fmt.Errorf("idempotency put: %w", err)
	}
	return nil
}
