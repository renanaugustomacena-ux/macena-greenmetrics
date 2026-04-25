# Architecture — GreenMetrics

## 1. High-Level View

```
+-------------------+           +-----------------------+           +---------------------+
|  Edge Gateways    |  HTTPS    |  Fiber API (Go 1.26)  |           |  TimescaleDB        |
|  (Modbus/M-Bus/   | --------> |  /api/v1/readings/    |  pgx      |  PG 16 + hypertable |
|   SunSpec/OCPP)   |  JSON     |  ingest, /meters, ... | --------> |  + CAGGs 15m/1h/1d  |
+-------------------+           +-----------+-----------+           +----------+----------+
                                            |                                  |
                                            | OTLP                             |
                                            v                                  |
                                +-------------------------+                    |
                                |  OpenTelemetry Collector|                    |
                                +-------------------------+                    |
                                                                               v
                                                                   +-----------------------+
                                                                   |  Grafana dashboards   |
                                                                   |  (provisioned JSON)   |
                                                                   +-----------------------+
```

External integrations:

- **ISPRA** — periodic pull of emission-factor tables into `emission_factors`.
- **Terna** — hourly national-mix download for market-based accounting.
- **E-Distribuzione SMD** — per-POD 15-min curves (OAuth2).
- **SPD (Servizio Portale Distribuzione)** — multi-DSO federated access via mTLS.
- **GSE** — outbound submission endpoints for Conto Termico 2.0 and Certificati Bianchi TEE.
- **ENEA audit102** — outbound dossier submission for D.Lgs. 102/2014 audits.

## 2. Components

| Layer | Component | Responsibility |
|-------|-----------|----------------|
| API | `cmd/server/main.go` | Bootstrap, config, logger, OTel, graceful shutdown |
| API | `internal/handlers/` | HTTP routing and request handling |
| Domain | `internal/services/` | Analytics, carbon calc, reports, ingestors, alerts |
| Models | `internal/models/` | Public types, DTOs |
| Data | `internal/repository/timescale_repository.go` | pgx pool, hypertable access |
| Schema | `backend/migrations/` | Tables, hypertables, continuous aggregates, retention |
| UI | `frontend/src/` | SvelteKit 2 console |
| Dash | `grafana/provisioning/` | Datasource + two dashboards |

## 3. Data Model

### Core relational tables

- `tenants` — UUID primary key; GDPR / CSRD / 102/2014 flags.
- `users` — RBAC roles (`admin`, `manager`, `operator`, `auditor`, `readonly`).
- `meters` — meter type + protocol + POD/PDR codes + endpoint.
- `meter_channels` — sub-channels (e.g. three-phase breakdown).
- `reports` — generated dossiers with `payload JSONB` and `status` lifecycle.
- `alerts` — fired alerts with `triggered_at / acked_at / resolved_at`.
- `emission_factors` — versioned, keyed on `(code, valid_from)`.
- `audit_log` — append-only audit trail.

### Time-series schema

`readings` is a TimescaleDB hypertable chunked by 1 day on `ts`:

```sql
SELECT create_hypertable('readings', 'ts',
    chunk_time_interval => INTERVAL '1 day');
```

Three continuous aggregates (defined in migration 0003) provide:

- `readings_15min` — refreshed every 15 minutes (start_offset 1 day, end_offset 15 min).
- `readings_1h` — refreshed hourly (start_offset 7 days).
- `readings_1d` — refreshed daily (start_offset 30 days).

Retention (migration 0004): raw 90 d, 15m 1 y, 1h 3 y, 1d 10 y. Compression is applied
to raw chunks older than 7 days with `segmentby = meter_id`.

### Emission factors (versioning)

```
(code, valid_from) primary key
valid_to NULL means "still active".
Query: SELECT * FROM emission_factors
       WHERE code = $1 AND valid_from <= $2
         AND (valid_to IS NULL OR valid_to > $2)
       ORDER BY valid_from DESC LIMIT 1;
```

This preserves reproducibility: a report generated in 2026 for period 2023 uses the
2023-valid factor; re-running the same report in 2027 produces identical output.

## 4. Sequence — Meter ingestion → Timescale write → alert

```
 Edge Gateway           Fiber Backend            TimescaleDB           AlertEngine
      |                      |                        |                     |
      | POST /readings/ingest|                        |                     |
      |--------------------->|                        |                     |
      |                      | pgx COPY → readings    |                     |
      |                      |----------------------->|                     |
      |                      |     bulk-insert ACK    |                     |
      |                      |<-----------------------|                     |
      |       202 Accepted   |                        |                     |
      |<---------------------|                        |                     |
      |                      | Async: trigger rules   |                     |
      |                      |----------------------------->               |
      |                      |                        | QueryAggregated     |
      |                      |                        |<--------------------|
      |                      |                        | return rows         |
      |                      |                        |-------------------->|
      |                      |                        |                     | z-score > 3?
      |                      |                        |                     | yes -> create Alert
      |                      |                        |<--------------------|
      |                      |                        | INSERT alerts       |
      |                      |                        |<--------------------|
```

Continuous aggregates refresh on their own schedule — dashboards query the 15-min
view for recent data, seamlessly falling back to the 1h or 1d view for older ranges.

## 5. Sequence — CSRD ESRS E1 report generation

```
Frontend              Backend /reports             ReportGenerator         Repo/Services
    |                        |                           |                        |
    | POST /reports (type=   |                           |                        |
    |   esrs_e1_csrd)        |                           |                        |
    |----------------------->|                           |                        |
    |                        | reporter.Generate()       |                        |
    |                        |-------------------------->|                        |
    |                        |                           | carbon.Compute(period) |
    |                        |                           |----------------------->|
    |                        |                           |  ListMeters, aggregate |
    |                        |                           |  × emission_factors    |
    |                        |                           |<-----------------------|
    |                        |                           | build ESRS E1 data     |
    |                        |                           | points (E1-5 / E1-6 /  |
    |                        |                           |  E1-7 / E1-8)          |
    |                        |<--------------------------|                        |
    |                        | persist reports row       |                        |
    |                        |--------------------------------------------------->|
    |  201 Created           |                           |                        |
    |<-----------------------|                           |                        |
    | GET /reports/:id/      |                           |                        |
    | download (PDF)         |                           |                        |
    |----------------------->|                           |                        |
    |  binary stream         |                           |                        |
    |<-----------------------|                           |                        |
```

## 6. Sequence — Piano Transizione 5.0 attestazione

```
Frontend              Backend                  ReportGenerator          AlgorithmicCore
    |                    |                            |                        |
    | POST /reports      |                            |                        |
    |   type=piano_5_0   |                            |                        |
    |   options={        |                            |                        |
    |     baseline_kwh,  |                            |                        |
    |     post_intervention_kwh,                       |                        |
    |     eligible_spend}|                            |                        |
    |------------------->|                            |                        |
    |                    | reporter.Generate()        |                        |
    |                    |--------------------------->|                        |
    |                    |                            | If baseline missing:   |
    |                    |                            |  derive from window    |
    |                    |                            |--------------------->  |
    |                    |                            | compute reduction_pct  |
    |                    |                            | compute process/site   |
    |                    |                            | thresholds (3%/5%)     |
    |                    |                            | assign credit band     |
    |                    |                            | compute EUR credit     |
    |                    |                            |<-----------------------|
    |                    |<---------------------------|                        |
    |                    | persist reports row        |                        |
    |  201 Created       |                            |                        |
    |<-------------------|                            |                        |
    | optional:          |                            |                        |
    | POST /reports/:id/submit -> GSE portal stub     |                        |
    |------------------->|                            |                        |
    |  200 OK            |                            |                        |
    |<-------------------|                            |                        |
```

Reference thresholds (Piano 5.0):

| Reduction | Band label | Indicative rate |
|-----------|------------|-----------------|
| < 3% process / < 5% site | `non-ammissibile` | 0% |
| 3–6% | `3-6% (aliquota base)` | 5% |
| 6–10% | `6-10% (intermedia)` | 20% |
| 10–15% | `10-15% (intermedia)` | 35% |
| ≥ 15% | `15%+ (superiore)` | 40% |

Actual decree rates depend on the spend scaglione and the specific call; the
implementation keeps these numeric parameters in `config` for quick alignment.

## 7. Multi-tenancy & security

- Every repo method filters by `tenant_id`.
- Optional PostgreSQL Row-Level Security toggled via env.
- JWT contains `sub`, `tenant_id`, `role`; middleware populates `c.Locals`.
- Field-level AES-256-GCM on `readings.raw_payload` (KMS-wrapped per-tenant DEK).
- HTTPS / TLS 1.3 mandatory in production, HSTS + CSP + X-Content-Type-Options on.
- RFC 7807 problem-details for errors.

## 8. Observability

- `X-Request-ID` propagated end-to-end.
- zap structured JSON logs.
- `otelfiber` spans wrap repository + service calls.
- Prometheus endpoint at `/api/internal/metrics`.
- Grafana dashboards in `grafana/provisioning/dashboards/`.

## 9. Deployment topology (production)

- AWS eu-south-1 (Milan), three AZs.
- Backend: 2–6 replicas behind ALB (HPA on CPU + custom "readings_ingest_rate").
- Frontend: 2 replicas behind CloudFront.
- TimescaleDB: 1 primary + 2 streaming replicas (AZ-split).
- Grafana: single replica + snapshotted data volume.
- All secrets in AWS Secrets Manager + per-env KMS keys.

## 10. Migration discipline

- Numeric-prefixed files in `backend/migrations/`.
- Each migration idempotent and re-runnable.
- No in-place hypertable alteration — create new + cutover.
- Schema changes frozen November–January.
