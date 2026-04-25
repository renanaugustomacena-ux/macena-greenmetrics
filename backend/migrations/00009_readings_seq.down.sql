-- +goose Down
-- +goose StatementBegin
-- Lossy: dropping `ingest_seq` discards CDC/replay state. Operator must accept.
DROP INDEX IF EXISTS readings_meter_ingest_seq_idx;
ALTER TABLE readings DROP COLUMN IF EXISTS ingest_seq;
-- +goose StatementEnd
