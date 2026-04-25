-- +goose Up
-- +goose NO TRANSACTION
--
-- Wrapped for goose v3 — file is idempotent (IF NOT EXISTS / OR REPLACE)
-- and was applied originally via TimescaleDB /docker-entrypoint-initdb.d
-- on first DB boot. Wrapping lets /migrate (cmd/migrate) record the
-- version in goose_db_version without re-applying destructively.
--
-- 0001_init.sql — GreenMetrics base schema
--
-- Prerequisites:
--   CREATE EXTENSION IF NOT EXISTS timescaledb;
--   CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
--   CREATE EXTENSION IF NOT EXISTS pgcrypto;

BEGIN;

CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Tenants -------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tenants (
    id                uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    ragione_sociale   text NOT NULL,
    partita_iva       text NOT NULL UNIQUE,
    codice_fiscale    text,
    sdi_code          text,
    pec               text,
    ateco             text,
    large_enterprise  boolean NOT NULL DEFAULT false,
    csrd_in_scope     boolean NOT NULL DEFAULT false,
    province          text,
    region            text,
    plan              text NOT NULL DEFAULT 'starter',
    meter_quota       int NOT NULL DEFAULT 10,
    active            boolean NOT NULL DEFAULT true,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);

-- Users ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email         text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role          text NOT NULL DEFAULT 'operator',
    full_name     text,
    mfa_enabled   boolean NOT NULL DEFAULT false,
    last_login_at timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS users_tenant_id_idx ON users(tenant_id);

-- Meters --------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS meters (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    label       text NOT NULL,
    meter_type  text NOT NULL,
    protocol    text NOT NULL,
    unit        text NOT NULL DEFAULT 'kWh',
    site        text,
    cost_centre text,
    serial_no   text,
    pod_code    text,
    pdr_code    text,
    endpoint    text,
    slave_addr  int,
    active      boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS meters_tenant_id_idx ON meters(tenant_id);
CREATE INDEX IF NOT EXISTS meters_active_idx ON meters(tenant_id, active);
CREATE UNIQUE INDEX IF NOT EXISTS meters_pod_uq ON meters(pod_code) WHERE pod_code IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS meters_pdr_uq ON meters(pdr_code) WHERE pdr_code IS NOT NULL;

CREATE TABLE IF NOT EXISTS meter_channels (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    meter_id     uuid NOT NULL REFERENCES meters(id) ON DELETE CASCADE,
    channel_code text NOT NULL,
    unit         text NOT NULL,
    scale_factor numeric NOT NULL DEFAULT 1.0,
    description  text
);
CREATE INDEX IF NOT EXISTS meter_channels_meter_idx ON meter_channels(meter_id);

-- Reports -------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS reports (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    type         text NOT NULL,
    period_from  timestamptz NOT NULL,
    period_to    timestamptz NOT NULL,
    status       text NOT NULL DEFAULT 'draft',
    payload      jsonb NOT NULL DEFAULT '{}'::jsonb,
    file_url     text,
    generated_by text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS reports_tenant_idx ON reports(tenant_id, type, period_from);

-- Alerts --------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS alerts (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    meter_id     uuid REFERENCES meters(id) ON DELETE SET NULL,
    kind         text NOT NULL,
    severity     text NOT NULL,
    message      text NOT NULL,
    context      jsonb,
    triggered_at timestamptz NOT NULL DEFAULT now(),
    acked_at     timestamptz,
    acked_by     text,
    resolved_at  timestamptz
);
CREATE INDEX IF NOT EXISTS alerts_tenant_idx ON alerts(tenant_id, triggered_at DESC);
CREATE INDEX IF NOT EXISTS alerts_open_idx ON alerts(tenant_id) WHERE resolved_at IS NULL;

-- Audit log -----------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_log (
    id             bigserial PRIMARY KEY,
    ts             timestamptz NOT NULL DEFAULT now(),
    tenant_id      uuid,
    actor_email    text,
    action         text NOT NULL,
    entity_type    text,
    entity_id      text,
    correlation_id text,
    details        jsonb
);
CREATE INDEX IF NOT EXISTS audit_log_tenant_ts_idx ON audit_log(tenant_id, ts DESC);

COMMIT;

-- +goose Down
-- +goose NO TRANSACTION
-- Down intentionally empty for this baseline migration. Production
-- rollback of pre-goose schema relies on pg_dump/pg_restore — see
-- docs/runbooks/db-outage.md.
SELECT 1;
