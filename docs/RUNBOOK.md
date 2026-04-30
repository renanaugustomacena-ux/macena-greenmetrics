# RUNBOOK — GreenMetrics

Operational runbook for the GreenMetrics energy & sustainability platform.

---

## 1. Stack anatomy

| Service | Port | Role |
|---------|------|------|
| greenmetrics-backend | 8082 | Go/Fiber API. |
| greenmetrics-frontend | 3005 | SvelteKit dashboard. |
| greenmetrics-timescaledb | 5439 | Timescale hypertable + CAGGs. |
| greenmetrics-grafana | 3011 | Dashboards + alerting. |
| greenmetrics-simulator | 5020 | Modbus-TCP server (dev/CI only). |

All services run on the `greenmetrics` docker network.

## 2. Boot / shutdown

```bash
cp .env.example .env
# edit .env — set POSTGRES_PASSWORD, GRAFANA_ADMIN_PASSWORD, JWT_SECRET
docker compose up -d --build
# G1 gate: all services healthy within 90s
docker compose ps
```

Shutdown:

```bash
docker compose down           # preserves volumes
docker compose down -v        # wipes timescale/grafana volumes
```

## 3. Health checks

- `curl -sf http://localhost:8082/api/health | jq` — canonical JSON with
  `status`, `service`, `version`, `uptime_seconds`, `time`,
  `dependencies: { timescaledb, grafana }`.
- `curl -sf http://localhost:8082/api/ready` — strict readiness.
- `curl -sf http://localhost:8082/api/live` — liveness.

## 4. Migrations

Migrations run automatically via `docker-entrypoint-initdb.d` on a fresh
Timescale volume. For upgrades:

```bash
make migrate          # from backend/ dir — runs 0001..0005 in order
```

Migration rollbacks are **not** scripted; for production, Timescale dumps
are taken before each migration.

## 5. Common operational tasks

### 5.1 Refresh emission factors (annual — April)

1. Download the new ISPRA rapporto (link in `docs/ITALIAN-COMPLIANCE.md`).
2. Add a new row to migration `0005_emission_factors.sql` with
   `valid_from = {year}-01-01`; close the previous row's `valid_to`.
3. Bump `cache` in `carbon_calculator.go`'s `defaultFactors`.
4. Add a test case in `tests/carbon_calculator_test.go`.

### 5.2 Rotate Grafana admin password

1. Generate a new 32-byte secret: `openssl rand -base64 32`.
2. Update `.env` → `GRAFANA_ADMIN_PASSWORD`.
3. `docker compose up -d greenmetrics-grafana` — rotation is no-op if
   volumes persisted; use `grafana-cli admin reset-admin-password`.

### 5.3 Trigger a Piano 5.0 attestazione

```bash
TOKEN=$(curl -s -X POST http://localhost:8082/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"audit@example.it","password":"anything"}' | jq -r .access_token)
curl -X POST http://localhost:8082/api/v1/reports \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "type": "piano_5_0_attestazione",
    "period_from": "2026-01-01T00:00:00Z",
    "period_to": "2026-03-31T23:59:59Z",
    "render": "html",
    "options": {
      "baseline_kwh": 100000,
      "post_intervention_kwh": 97000,
      "eligible_spend_eur": 500000,
      "process_scope": true
    }
  }'
```

## 6. Troubleshooting

| Symptom | Likely cause | Remediation |
|---------|--------------|-------------|
| Backend exits with `refusing default JWT_SECRET` | `APP_ENV=production` with placeholder secret | set a real `JWT_SECRET` in env. |
| `pg_isready` healthcheck failing | Timescale extension load on first init | wait 30s; check `docker logs greenmetrics-timescaledb`. |
| Health reports `timescaledb: degraded` | DB pool pings failing | check `DATABASE_URL` and network reachability. |
| Grafana shows `Plugin not found` | slow `GF_INSTALL_PLUGINS` fetch | restart container. |
| Simulator has no data in Grafana | CAGG not refreshed | wait 15 minutes or call `CALL refresh_continuous_aggregate('readings_15min', NULL, NULL);`. |

## 7. Disaster recovery

- **Backup:** nightly `pg_dump -Fd` of the `greenmetrics` database to S3
  (wire outside compose; see `docs/SLO.md`).
- **Restore:** `pg_restore -d greenmetrics -Fd <dir>/` into a fresh
  Timescale instance, then re-run migrations for idempotent schema state.

## 8. Observability

- Prometheus scraping: `GET /api/internal/metrics`.
- OpenTelemetry OTLP/gRPC → set `OTEL_EXPORTER_OTLP_ENDPOINT` to your
  collector endpoint (e.g. `otel-collector:4317`).
- Structured logs: JSON on stdout (zap production mode) when
  `APP_ENV=production`.
