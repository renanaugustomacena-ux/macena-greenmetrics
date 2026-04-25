-- +goose Up
-- +goose StatementBegin
--
-- Server-assigned monotonic ingest sequence per meter for total ordering.
-- Consumers needing strict order (CDC, replay, audit) use ingest_seq.
-- Lock impact: ADD COLUMN with DEFAULT (immediate metadata-only operation in PG 11+).
--
-- Doctrine: Rule 35 (ordering guarantees).

ALTER TABLE readings
    ADD COLUMN IF NOT EXISTS ingest_seq BIGSERIAL;

-- Unique index per meter to make ingest_seq actionable for replay/CDC.
CREATE INDEX IF NOT EXISTS readings_meter_ingest_seq_idx
    ON readings (tenant_id, meter_id, ingest_seq);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Lossy: dropping `ingest_seq` discards CDC/replay state. Operator must accept.
DROP INDEX IF EXISTS readings_meter_ingest_seq_idx;
ALTER TABLE readings DROP COLUMN IF EXISTS ingest_seq;
-- +goose StatementEnd
