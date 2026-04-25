// Package repository wraps TimescaleDB/PostgreSQL access.
//
// Schema highlights:
//   - readings is a TimescaleDB hypertable chunked by 1 day on ts.
//   - Continuous aggregates roll-up raw → 15min → 1h → 1d.
//   - Retention policies drop raw > 90d; 15min > 1y; 1h > 3y; 1d > 10y.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// jsonEncode is a small helper so callers do not need to import encoding/json.
func jsonEncode(v any) ([]byte, error) {
	if v == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(v)
}

// TimescaleRepository provides data-access primitives for the GreenMetrics domain.
type TimescaleRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewTimescaleRepository establishes a pool. Missing DB is non-fatal in dev;
// caller logs a warning and can continue serving stub responses.
func NewTimescaleRepository(ctx context.Context, dsn string, logger *zap.Logger) (*TimescaleRepository, error) {
	if dsn == "" {
		return nil, errors.New("empty DATABASE_URL")
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}
	poolCfg.MaxConns = 25
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(pingCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	logger.Info("timescaledb pool ready", zap.Int32("max_conns", poolCfg.MaxConns))
	return &TimescaleRepository{pool: pool, logger: logger}, nil
}

// Close releases the pool.
func (r *TimescaleRepository) Close() {
	if r != nil && r.pool != nil {
		r.pool.Close()
	}
}

// Pool exposes the underlying pgxpool for callers that need raw access
// (audit writer, health probes). Returns nil if the repo was never connected.
func (r *TimescaleRepository) Pool() *pgxpool.Pool {
	if r == nil {
		return nil
	}
	return r.pool
}

// AuditEntry is one row for the audit_log table.
type AuditEntry struct {
	TenantID      string
	ActorEmail    string
	Action        string
	EntityType    string
	EntityID      string
	CorrelationID string
	Details       map[string]any
}

// InsertAudit writes a single audit_log row. tenant_id is cast to UUID only
// when non-empty (the column is nullable).
func (r *TimescaleRepository) InsertAudit(ctx context.Context, e AuditEntry) (int64, error) {
	if r == nil || r.pool == nil {
		return 0, errors.New("pool not initialised")
	}
	const q = `
INSERT INTO audit_log (tenant_id, actor_email, action, entity_type, entity_id, correlation_id, details)
VALUES (NULLIF($1,'')::uuid, NULLIF($2,''), $3, NULLIF($4,''), NULLIF($5,''), NULLIF($6,''), $7::jsonb)
RETURNING id`
	detailsJSON, err := jsonEncode(e.Details)
	if err != nil {
		return 0, fmt.Errorf("audit details encode: %w", err)
	}
	var id int64
	row := r.pool.QueryRow(ctx, q, e.TenantID, e.ActorEmail, e.Action, e.EntityType, e.EntityID, e.CorrelationID, detailsJSON)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// FindUserCredentials loads (password_hash, tenant_id, role) by email.
//
// Returns pgx.ErrNoRows if no user exists (caller should normalise to a
// generic "invalid credentials" error — never reveal which of user/pass failed).
func (r *TimescaleRepository) FindUserCredentials(ctx context.Context, email string) (string, string, string, error) {
	if r == nil || r.pool == nil {
		return "", "", "", errors.New("pool not initialised")
	}
	const q = `SELECT password_hash, tenant_id::text, role FROM users WHERE email = $1 LIMIT 1`
	var hash, tenantID, role string
	row := r.pool.QueryRow(ctx, q, email)
	if err := row.Scan(&hash, &tenantID, &role); err != nil {
		return "", "", "", err
	}
	return hash, tenantID, role, nil
}

// CountAudit returns the number of audit_log rows (helper for tests).
func (r *TimescaleRepository) CountAudit(ctx context.Context) (int64, error) {
	if r == nil || r.pool == nil {
		return 0, errors.New("pool not initialised")
	}
	var n int64
	row := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_log`)
	return n, row.Scan(&n)
}

// Ping tests connectivity (for the /api/health dependencies block).
func (r *TimescaleRepository) Ping(ctx context.Context) error {
	if r == nil || r.pool == nil {
		return errors.New("pool not initialised")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.pool.Ping(ctx)
}

// Reading is an inbound meter reading.
type Reading struct {
	Ts           time.Time
	TenantID     string
	MeterID      string
	ChannelID    string
	Value        float64
	Unit         string
	QualityCode  int
	RawPayload   []byte
}

// InsertReadings performs a bulk copy into the hypertable.
func (r *TimescaleRepository) InsertReadings(ctx context.Context, rows []Reading) (int64, error) {
	if r == nil || r.pool == nil {
		return 0, errors.New("pool not initialised")
	}
	if len(rows) == 0 {
		return 0, nil
	}
	src := make([][]any, len(rows))
	for i, rr := range rows {
		src[i] = []any{rr.Ts, rr.TenantID, rr.MeterID, rr.ChannelID, rr.Value, rr.Unit, rr.QualityCode, rr.RawPayload}
	}
	return r.pool.CopyFrom(ctx,
		pgx.Identifier{"readings"},
		[]string{"ts", "tenant_id", "meter_id", "channel_id", "value", "unit", "quality_code", "raw_payload"},
		pgx.CopyFromRows(src),
	)
}

// AggregateResult represents a single time bucket of aggregated consumption.
type AggregateResult struct {
	Bucket    time.Time
	MeterID   string
	ChannelID string
	SumValue  float64
	AvgValue  float64
	MaxValue  float64
	Unit      string
}

// QueryAggregated pulls from the appropriate continuous aggregate.
//
// resolution ∈ {"15min", "1h", "1d"}. The repository picks the matching view.
func (r *TimescaleRepository) QueryAggregated(ctx context.Context, tenantID, meterID, resolution string, from, to time.Time) ([]AggregateResult, error) {
	if r == nil || r.pool == nil {
		return nil, errors.New("pool not initialised")
	}
	view := "readings_15min"
	switch resolution {
	case "1h":
		view = "readings_1h"
	case "1d":
		view = "readings_1d"
	case "15min":
		view = "readings_15min"
	default:
		return nil, fmt.Errorf("unknown resolution: %s", resolution)
	}
	q := fmt.Sprintf(`
SELECT bucket, meter_id, channel_id, sum_value, avg_value, max_value, unit
FROM %s
WHERE tenant_id = $1 AND meter_id = $2 AND bucket >= $3 AND bucket < $4
ORDER BY bucket ASC`, view)
	rows, err := r.pool.Query(ctx, q, tenantID, meterID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AggregateResult{}
	for rows.Next() {
		var a AggregateResult
		if err := rows.Scan(&a.Bucket, &a.MeterID, &a.ChannelID, &a.SumValue, &a.AvgValue, &a.MaxValue, &a.Unit); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// EmissionFactor is a versioned emission factor (kg CO2e / unit).
type EmissionFactor struct {
	Code       string
	Scope      int
	Category   string
	Unit       string
	KgCO2ePer  float64
	Source     string
	ValidFrom  time.Time
	ValidTo    *time.Time
	Version    string
}

// CurrentEmissionFactor fetches the active factor for a code at a moment in time.
func (r *TimescaleRepository) CurrentEmissionFactor(ctx context.Context, code string, at time.Time) (*EmissionFactor, error) {
	if r == nil || r.pool == nil {
		return nil, errors.New("pool not initialised")
	}
	const q = `
SELECT code, scope, category, unit, kg_co2e_per_unit, source, valid_from, valid_to, version
FROM emission_factors
WHERE code = $1 AND valid_from <= $2 AND (valid_to IS NULL OR valid_to > $2)
ORDER BY valid_from DESC LIMIT 1`
	row := r.pool.QueryRow(ctx, q, code, at)
	var e EmissionFactor
	if err := row.Scan(&e.Code, &e.Scope, &e.Category, &e.Unit, &e.KgCO2ePer, &e.Source, &e.ValidFrom, &e.ValidTo, &e.Version); err != nil {
		return nil, err
	}
	return &e, nil
}

// MeterRow as stored.
type MeterRow struct {
	ID        string
	TenantID  string
	Label     string
	MeterType string
	Protocol  string
	Site      string
	CostCentre string
	Active    bool
	CreatedAt time.Time
}

// ListMeters returns meters for a tenant.
func (r *TimescaleRepository) ListMeters(ctx context.Context, tenantID string) ([]MeterRow, error) {
	if r == nil || r.pool == nil {
		return nil, errors.New("pool not initialised")
	}
	const q = `SELECT id, tenant_id, label, meter_type, protocol, site, cost_centre, active, created_at
FROM meters WHERE tenant_id = $1 ORDER BY label`
	rows, err := r.pool.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MeterRow{}
	for rows.Next() {
		var m MeterRow
		if err := rows.Scan(&m.ID, &m.TenantID, &m.Label, &m.MeterType, &m.Protocol, &m.Site, &m.CostCentre, &m.Active, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
