-- 0003_continuous_aggregates.sql — 15min / 1h / 1d roll-ups.
--
-- Each level aggregates the level below, so refresh policies cascade.

BEGIN;

-- 15-min continuous aggregate -----------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS readings_15min
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('15 minutes', ts) AS bucket,
    tenant_id,
    meter_id,
    channel_id,
    SUM(value) AS sum_value,
    AVG(value) AS avg_value,
    MAX(value) AS max_value,
    MIN(value) AS min_value,
    COUNT(*)   AS n_samples,
    unit
FROM readings
GROUP BY bucket, tenant_id, meter_id, channel_id, unit
WITH NO DATA;

SELECT add_continuous_aggregate_policy('readings_15min',
    start_offset => INTERVAL '1 day',
    end_offset   => INTERVAL '15 minutes',
    schedule_interval => INTERVAL '15 minutes');

CREATE INDEX IF NOT EXISTS readings_15min_meter_bucket_idx ON readings_15min(meter_id, bucket DESC);

-- 1-hour continuous aggregate -----------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS readings_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', ts) AS bucket,
    tenant_id,
    meter_id,
    channel_id,
    SUM(value) AS sum_value,
    AVG(value) AS avg_value,
    MAX(value) AS max_value,
    MIN(value) AS min_value,
    COUNT(*)   AS n_samples,
    unit
FROM readings
GROUP BY bucket, tenant_id, meter_id, channel_id, unit
WITH NO DATA;

SELECT add_continuous_aggregate_policy('readings_1h',
    start_offset => INTERVAL '7 days',
    end_offset   => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

CREATE INDEX IF NOT EXISTS readings_1h_meter_bucket_idx ON readings_1h(meter_id, bucket DESC);

-- 1-day continuous aggregate ------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS readings_1d
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', ts) AS bucket,
    tenant_id,
    meter_id,
    channel_id,
    SUM(value) AS sum_value,
    AVG(value) AS avg_value,
    MAX(value) AS max_value,
    MIN(value) AS min_value,
    COUNT(*)   AS n_samples,
    unit
FROM readings
GROUP BY bucket, tenant_id, meter_id, channel_id, unit
WITH NO DATA;

SELECT add_continuous_aggregate_policy('readings_1d',
    start_offset => INTERVAL '30 days',
    end_offset   => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');

CREATE INDEX IF NOT EXISTS readings_1d_meter_bucket_idx ON readings_1d(meter_id, bucket DESC);

COMMIT;
