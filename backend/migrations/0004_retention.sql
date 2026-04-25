-- +goose Up
-- +goose NO TRANSACTION
--
-- Wrapped for goose v3 — file is idempotent (IF NOT EXISTS / OR REPLACE)
-- and was applied originally via TimescaleDB /docker-entrypoint-initdb.d
-- on first DB boot. Wrapping lets /migrate (cmd/migrate) record the
-- version in goose_db_version without re-applying destructively.
--
-- 0004_retention.sql — retention policies per spec.
--
-- Raw readings:          90 days
-- 15-min aggregates:     1 year
-- 1-hour aggregates:     3 years
-- 1-day aggregates:      10 years
--
-- These policies interact with the compression policy in 0002; once a chunk
-- is dropped we lose raw fidelity but keep aggregate history.

BEGIN;

SELECT add_retention_policy('readings',       INTERVAL '90 days',   if_not_exists => TRUE);
SELECT add_retention_policy('readings_15min', INTERVAL '365 days',  if_not_exists => TRUE);
SELECT add_retention_policy('readings_1h',    INTERVAL '1095 days', if_not_exists => TRUE);
SELECT add_retention_policy('readings_1d',    INTERVAL '3650 days', if_not_exists => TRUE);

COMMIT;

-- +goose Down
-- +goose NO TRANSACTION
-- Down intentionally empty for this baseline migration. Production
-- rollback of pre-goose schema relies on pg_dump/pg_restore — see
-- docs/runbooks/db-outage.md.
SELECT 1;
