-- +goose Down
-- +goose StatementBegin
DROP INDEX CONCURRENTLY IF EXISTS readings_dedup_idx;
-- +goose StatementEnd
