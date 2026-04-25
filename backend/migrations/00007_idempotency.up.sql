-- +goose Up
-- +goose StatementBegin
--
-- Idempotency keys for POST endpoints (Rule 35, RFC 9457 / Stripe-style).
-- Stored per-tenant, hashed request body, replayed response.
-- Retention 24 h via Timescale retention policy.
--
-- Doctrine refs: Rule 14, Rule 34, Rule 35.
-- Mitigates: ingestion duplicates from retry storms; out-of-order client retries.

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

-- Convert to hypertable for fast retention drop.
SELECT create_hypertable(
    'idempotency_keys',
    'created_at',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 24h retention.
SELECT add_retention_policy('idempotency_keys', INTERVAL '24 hours', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idempotency_keys_created_at_idx ON idempotency_keys (created_at DESC);
-- +goose StatementEnd
