-- +goose Up
-- +goose NO TRANSACTION
--
-- Wrapped for goose v3 — file is idempotent (IF NOT EXISTS / OR REPLACE)
-- and was applied originally via TimescaleDB /docker-entrypoint-initdb.d
-- on first DB boot. Wrapping lets /migrate (cmd/migrate) record the
-- version in goose_db_version without re-applying destructively.
--
-- 0005_emission_factors.sql — versioned emission factors seed.
--
-- Factors sourced from:
--   - ISPRA "Fattori di emissione per la produzione ed il consumo di energia
--     elettrica in Italia" (Rapporto 404/2024) — annual publication.
--   - D.M. 11 maggio 2022 (MiTE) — parametri standard nazionali (combustibili).
--   - AIB European Residual Mixes — annual.
--   - GSE Conto Termico 2.0 — district heat reference.

BEGIN;

CREATE TABLE IF NOT EXISTS emission_factors (
    code             text NOT NULL,
    scope            smallint NOT NULL,
    category         text NOT NULL,
    unit             text NOT NULL,
    kg_co2e_per_unit double precision NOT NULL,
    source           text NOT NULL,
    valid_from       timestamptz NOT NULL,
    valid_to         timestamptz,
    version          text NOT NULL,
    notes            text,
    PRIMARY KEY (code, valid_from)
);
CREATE INDEX IF NOT EXISTS emission_factors_code_idx ON emission_factors(code, valid_from DESC);

-- Seed values — update annually from ISPRA publication.
INSERT INTO emission_factors (code, scope, category, unit, kg_co2e_per_unit, source, valid_from, valid_to, version, notes)
VALUES
    ('IT_ELEC_MIX_2022', 2, 'electricity_mix', 'kWh', 0.263, 'ISPRA 2023 Rapporto 386', '2022-01-01', '2023-01-01', '2023.1', 'Mix produzione Italia 2022'),
    ('IT_ELEC_MIX_2023', 2, 'electricity_mix', 'kWh', 0.250, 'ISPRA 2024 Rapporto 404', '2023-01-01', '2024-01-01', '2024.1', 'Mix produzione Italia 2023'),
    ('IT_ELEC_MIX_2024', 2, 'electricity_mix', 'kWh', 0.245, 'ISPRA 2024 stima provvisoria', '2024-01-01', NULL, '2024.1', 'Dato provvisorio — da ricalcolare ad aprile 2026'),
    ('IT_ELEC_RESIDUAL_MIX_2023', 2, 'electricity_market_based', 'kWh', 0.457, 'AIB European Residual Mix 2023', '2023-01-01', '2024-01-01', '2023.1', 'Valore per metodologia market-based'),
    ('NG_STATIONARY_COMBUSTION', 1, 'natural_gas', 'Sm3', 1.975, 'ISPRA D.M. 11/05/2022', '2022-05-11', NULL, '2022.1', 'Combustione stazionaria'),
    ('DIESEL_COMBUSTION', 1, 'diesel', 'L', 2.650, 'ISPRA 2024', '2024-01-01', NULL, '2024.1', 'Gasolio autotrazione'),
    ('LPG_COMBUSTION', 1, 'lpg', 'L', 1.510, 'ISPRA 2024', '2024-01-01', NULL, '2024.1', 'GPL riscaldamento'),
    ('HEATING_OIL_COMBUSTION', 1, 'heating_oil', 'L', 2.771, 'ISPRA 2024', '2024-01-01', NULL, '2024.1', 'Olio combustibile riscaldamento'),
    ('DISTRICT_HEAT_AVERAGE_IT', 2, 'district_heat', 'kWh', 0.200, 'GSE Conto Termico 2.0 reference', '2024-01-01', NULL, '2024.1', 'Teleriscaldamento media nazionale')
ON CONFLICT (code, valid_from) DO NOTHING;

COMMIT;

-- +goose Down
-- +goose NO TRANSACTION
-- Down intentionally empty for this baseline migration. Production
-- rollback of pre-goose schema relies on pg_dump/pg_restore — see
-- docs/runbooks/db-outage.md.
SELECT 1;
