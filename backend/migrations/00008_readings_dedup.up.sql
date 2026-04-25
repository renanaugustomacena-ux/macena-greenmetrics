-- +goose Up
-- +goose StatementBegin
--
-- Per-meter, per-channel, per-timestamp uniqueness on readings to dedup ingestion.
-- Required by ingest pipeline ON CONFLICT DO NOTHING semantics (Rule 35).
-- Lock impact: CREATE UNIQUE INDEX CONCURRENTLY against existing hypertable rows.
--   On a populated table this can take minutes. Schedule maintenance window per
--   docs/runbooks/db-outage.md (S4).
--
-- Doctrine: Rule 35 (consistency / idempotency), Rule 33 (data integrity).

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS readings_dedup_idx
    ON readings (tenant_id, meter_id, channel_id, ts);
-- +goose StatementEnd
