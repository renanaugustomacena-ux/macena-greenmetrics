-- Dev seed — single tenant, single admin user, 5 Modbus meters tied to the
-- in-cluster simulator. Idempotent: ON CONFLICT DO NOTHING everywhere.
--
-- Tenant ID matches the constant baked into:
--   - backend/internal/services/ingestor_runner.go:75
--   - backend/internal/handlers/readings.go:41,79
--   - backend/internal/handlers/auth.go:126 (dev fallback)
--
-- Apply via:
--   docker compose exec -T greenmetrics-timescaledb \
--       psql -U greenmetrics -d greenmetrics -f /tmp/seed.sql
--
-- Login as: dev@greenmetrics.local / OperatorPass2026!

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1. Tenant ----------------------------------------------------------------
INSERT INTO tenants (
    id, ragione_sociale, partita_iva, codice_fiscale, sdi_code,
    pec, ateco, large_enterprise, csrd_in_scope,
    province, region, plan, meter_quota, active
) VALUES (
    '00000000-0000-4000-8000-000000000001',
    'GreenMetrics Development Tenant',
    '00000000001',
    '00000000001',
    '0000000',
    'pec@greenmetrics.local',
    '70.10.00',
    false,
    false,
    'VR',
    'Veneto',
    'starter',
    100,
    true
) ON CONFLICT (id) DO NOTHING;

-- 2. Admin user ------------------------------------------------------------
-- bcrypt cost 12 of "OperatorPass2026!"; pgcrypto's blowfish is byte-compatible
-- with golang.org/x/crypto/bcrypt.
INSERT INTO users (
    id, tenant_id, email, password_hash, role, full_name, mfa_enabled
) VALUES (
    '00000000-0000-4000-8000-00000000a001',
    '00000000-0000-4000-8000-000000000001',
    'dev@greenmetrics.local',
    crypt('OperatorPass2026!', gen_salt('bf', 12)),
    'admin',
    'Dev Admin',
    false
) ON CONFLICT (email) DO UPDATE
    SET password_hash = EXCLUDED.password_hash,
        role          = EXCLUDED.role;

-- 3. Modbus meters ---------------------------------------------------------
-- Match MODBUS_SLAVE_IDS=1,2,3,4,5 (default config); endpoint points to the
-- simulator container on the docker compose network.
INSERT INTO meters (
    id, tenant_id, label, meter_type, protocol, unit, site, cost_centre,
    serial_no, endpoint, slave_addr, active
) VALUES
    ('00000000-0000-4000-8000-0000000eee01', '00000000-0000-4000-8000-000000000001',
     'Linea Produzione 1', 'electricity', 'modbus_tcp', 'kWh',
     'Stabilimento Verona', 'Produzione',
     'SIM-001', 'greenmetrics-simulator:5020', 1, true),
    ('00000000-0000-4000-8000-0000000eee02', '00000000-0000-4000-8000-000000000001',
     'Linea Produzione 2', 'electricity', 'modbus_tcp', 'kWh',
     'Stabilimento Verona', 'Produzione',
     'SIM-002', 'greenmetrics-simulator:5020', 2, true),
    ('00000000-0000-4000-8000-0000000eee03', '00000000-0000-4000-8000-000000000001',
     'Compressore Aria', 'electricity', 'modbus_tcp', 'kW',
     'Stabilimento Verona', 'Servizi',
     'SIM-003', 'greenmetrics-simulator:5020', 3, true),
    ('00000000-0000-4000-8000-0000000eee04', '00000000-0000-4000-8000-000000000001',
     'Caldaia Gas Naturale', 'gas', 'modbus_tcp', 'm3',
     'Stabilimento Verona', 'Servizi',
     'SIM-004', 'greenmetrics-simulator:5020', 4, true),
    ('00000000-0000-4000-8000-0000000eee05', '00000000-0000-4000-8000-000000000001',
     'Impianto FV Tetto', 'photovoltaic', 'modbus_tcp', 'kWh',
     'Stabilimento Verona', 'Generazione',
     'SIM-005', 'greenmetrics-simulator:5020', 5, true)
ON CONFLICT (id) DO NOTHING;

-- 4. Meter channels --------------------------------------------------------
-- One default channel per meter (active power / cumulative energy).
INSERT INTO meter_channels (
    id, meter_id, channel_code, unit, scale_factor, description
) VALUES
    ('00000000-0000-4000-8000-0000ccc00001', '00000000-0000-4000-8000-0000000eee01', 'P_active', 'kW', 1.0, 'Active power'),
    ('00000000-0000-4000-8000-0000ccc00002', '00000000-0000-4000-8000-0000000eee02', 'P_active', 'kW', 1.0, 'Active power'),
    ('00000000-0000-4000-8000-0000ccc00003', '00000000-0000-4000-8000-0000000eee03', 'P_active', 'kW', 1.0, 'Active power'),
    ('00000000-0000-4000-8000-0000ccc00004', '00000000-0000-4000-8000-0000000eee04', 'V_flow',   'm3/h', 1.0, 'Gas flow rate'),
    ('00000000-0000-4000-8000-0000ccc00005', '00000000-0000-4000-8000-0000000eee05', 'P_pv',     'kW', 1.0, 'PV active power')
ON CONFLICT (id) DO NOTHING;

-- 5. Verification queries (psql will show counts) --------------------------
SELECT 'tenants'       AS table_name, count(*) AS rows FROM tenants
UNION ALL SELECT 'users', count(*)    FROM users
UNION ALL SELECT 'meters', count(*)   FROM meters
UNION ALL SELECT 'meter_channels', count(*) FROM meter_channels;
