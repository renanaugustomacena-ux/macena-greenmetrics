-- 0002_hypertables.sql — raw readings hypertable.
--
-- Chunk interval: 1 day (spec requirement). A smaller interval would inflate
-- chunk count; a larger one reduces selectivity for recent queries.

BEGIN;

CREATE TABLE IF NOT EXISTS readings (
    ts           timestamptz NOT NULL,
    tenant_id    uuid NOT NULL,
    meter_id     uuid NOT NULL,
    channel_id   uuid,
    value        double precision NOT NULL,
    unit         text NOT NULL,
    quality_code smallint NOT NULL DEFAULT 0,
    raw_payload  jsonb
);

-- Promote to hypertable with 1-day chunks.
SELECT create_hypertable(
    'readings', 'ts',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Space-partition hint: none. Tenant segmentation handled at query level +
-- chunk compression. A composite (tenant_id, meter_id) space-partition could
-- be added later at >1000 tenants.

CREATE INDEX IF NOT EXISTS readings_meter_ts_idx ON readings(meter_id, ts DESC);
CREATE INDEX IF NOT EXISTS readings_tenant_ts_idx ON readings(tenant_id, ts DESC);

-- Compression policy: compress chunks older than 7 days.
ALTER TABLE readings SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'meter_id',
    timescaledb.compress_orderby   = 'ts DESC'
);

SELECT add_compression_policy('readings', INTERVAL '7 days', if_not_exists => TRUE);

COMMIT;
