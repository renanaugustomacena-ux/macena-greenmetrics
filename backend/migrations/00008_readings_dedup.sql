-- +goose Up
-- +goose StatementBegin
--
-- Per-meter, per-channel, per-timestamp uniqueness on readings to dedup ingestion.
-- Required by ingest pipeline ON CONFLICT DO NOTHING semantics (Rule 35).
--
-- TimescaleDB hypertables do NOT support CREATE INDEX CONCURRENTLY (creation
-- propagates to every chunk; concurrency would race per-chunk metadata).
-- Locks the hypertable + every chunk briefly during creation.
--
-- Lock impact: schedule maintenance window per docs/runbooks/db-outage.md
-- on populated production tables (≥ 100M rows). Empty / development DB:
-- sub-second.
--
-- Doctrine: Rule 35 (consistency / idempotency), Rule 33 (data integrity).

CREATE UNIQUE INDEX IF NOT EXISTS readings_dedup_idx
    ON readings (tenant_id, meter_id, channel_id, ts);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS readings_dedup_idx;
-- +goose StatementEnd
