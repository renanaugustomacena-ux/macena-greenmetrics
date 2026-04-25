-- +goose Up
-- +goose StatementBegin
--
-- Idempotency keys for POST endpoints (Rule 35, RFC 9457 / Stripe-style).
-- Stored per-tenant, hashed request body, replayed response.
--
-- Doctrine refs: Rule 14, Rule 34, Rule 35.
-- Mitigates: ingestion duplicates from retry storms; out-of-order client retries.
--
-- DESIGN NOTE: kept as a plain table (NOT a TimescaleDB hypertable) because the
-- middleware's natural lookup key is (tenant_id, key) and Timescale requires
-- the partition column to participate in every unique index. A 24h cleanup job
-- handles retention out-of-band (see scripts/ops/idempotency-gc.sh, S5
-- follow-on; for now, manual VACUUM + DELETE WHERE created_at < now() - '24h').

CREATE TABLE IF NOT EXISTS idempotency_keys (
    tenant_id        UUID         NOT NULL,
    key              TEXT         NOT NULL,
    request_hash     BYTEA        NOT NULL,
    response_status  INTEGER      NOT NULL,
    response_body    BYTEA        NOT NULL,
    response_headers JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, key)
);

CREATE INDEX IF NOT EXISTS idempotency_keys_created_at_idx ON idempotency_keys (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS idempotency_keys;
-- +goose StatementEnd
