-- +goose Down
-- +goose StatementBegin
-- Lossy: drops 24h of idempotency state. Production rollback risks duplicate processing
-- on in-flight retries. Do not run against live tenant DB without IC sign-off.

SELECT remove_retention_policy('idempotency_keys', if_exists => TRUE);
DROP TABLE IF EXISTS idempotency_keys;
-- +goose StatementEnd
