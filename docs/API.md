# API Reference — GreenMetrics

All endpoints are served from the Fiber backend. The base URL in production is
`https://api.greenmetrics.it`; in the local docker-compose environment it is
`http://localhost:8080`.

Versioned surface is at `/api/v1`. Errors follow RFC 7807 problem-details JSON:

```json
{
  "type": "about:blank",
  "title": "Bad Request",
  "status": 400,
  "detail": "meter_id required",
  "instance": "/api/v1/readings?from=..."
}
```

Authentication: `Authorization: Bearer <access_token>` JWT. Obtain via `/api/v1/auth/login`.
Correlation: every response carries `X-Request-ID`.

---

## 1. Health

### GET /api/health

Public. Returns platform health + dependency status.

Response:

```json
{
  "status": "ok",
  "service": "greenmetrics-backend",
  "version": "0.1.0",
  "uptime_seconds": 12345.67,
  "time": "2026-04-17T10:30:00Z",
  "dependencies": {
    "timescaledb": "ok",
    "grafana": "ok"
  }
}
```

`status` is `ok` when every dependency is `ok`; `degraded` otherwise.

### GET /api/ready

Strict readiness probe; returns `503 Service Unavailable` if TimescaleDB is not reachable.

### GET /api/live

Liveness probe; returns 200 whenever the process is serving.

---

## 2. Authentication

### POST /api/v1/auth/login

Body:

```json
{ "email": "user@example.it", "password": "secret" }
```

Response 200:

```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

Errors: 400 `email and password required`, 401 `invalid credentials`.

### POST /api/v1/auth/refresh

Body: `{ "refresh_token": "..." }`. Returns a new access token.

### POST /api/v1/auth/logout

Returns 204. Client-side discards tokens; revocation list optional.

---

## 3. Tenants

### GET /api/v1/tenants/me

Returns the caller's tenant:

```json
{
  "id": "...",
  "ragione_sociale": "Industria Esempio S.r.l.",
  "partita_iva": "IT01234567890",
  "ateco": "10.89.09",
  "province": "VR",
  "large_enterprise": false,
  "csrd_in_scope": false,
  "plan": "professionale",
  "meter_quota": 50,
  "active": true
}
```

### PUT /api/v1/tenants/me

Body: full tenant object. Returns updated tenant.

---

## 4. Meters

### GET /api/v1/meters

Query parameters: `active`, `meter_type`, `protocol`, `site`.

Response:

```json
{ "items": [{ "id": "...", "label": "...", "meter_type": "electricity_3p", "protocol": "modbus_tcp", "site": "Stabilimento A", "active": true, "created_at": "..." }], "total": 1 }
```

### POST /api/v1/meters

Body:

```json
{
  "label": "Quadro Generale CE-001",
  "meter_type": "electricity_3p",
  "protocol": "modbus_tcp",
  "unit": "kWh",
  "site": "Stabilimento A",
  "cost_centre": "CC-001",
  "endpoint": "192.168.10.50:502",
  "slave_addr": 1,
  "pod_code": "IT001E00000000"
}
```

Response 201 with created meter.

Errors: 400 on missing required fields.

### GET /api/v1/meters/:id · PUT /api/v1/meters/:id · DELETE /api/v1/meters/:id

Standard CRUD. DELETE is a soft-delete (sets `active = false`).

### POST /api/v1/meters/:id/probe

Triggers a synchronous read of the meter for diagnostics:

```json
{ "meter_id": "...", "probed_at": "2026-04-17T10:30:00Z", "status": "ok", "latency_ms": 42 }
```

---

## 5. Readings

### POST /api/v1/readings/ingest

Bulk-ingest endpoint used by edge gateways.

Body:

```json
{
  "readings": [
    { "ts": "2026-04-17T10:00:00Z", "meter_id": "...", "channel_id": "...", "value": 42.1, "unit": "kWh", "quality_code": 0 }
  ]
}
```

Response 202:

```json
{ "accepted": 1, "received": 1 }
```

The body is idempotent within a replay window; de-duplication at the DB layer is
not enforced (cost vs. benefit trade-off).

### GET /api/v1/readings?meter_id=...&from=...&to=...&limit=1000

Returns raw readings (capped at `limit`, default 1000).

Errors: 400 if `from`/`to` missing or invalid RFC 3339.

### GET /api/v1/readings/aggregated?meter_id=...&resolution=15min|1h|1d&from=...&to=

Returns rows from the corresponding continuous aggregate view.

```json
{
  "resolution": "1h",
  "items": [
    { "bucket": "2026-04-17T10:00:00Z", "meter_id": "...", "channel_id": "...", "sum_value": 123.4, "avg_value": 42.1, "max_value": 55.5, "unit": "kWh" }
  ],
  "count": 1
}
```

### GET /api/v1/readings/export?meter_id=...&from=...&to=

Streams CSV with header `ts,tenant_id,meter_id,channel_id,value,unit,quality_code`.

---

## 6. Reports

### POST /api/v1/reports

Body:

```json
{
  "type": "piano_5_0_attestazione",
  "period_from": "2025-01-01T00:00:00Z",
  "period_to": "2025-12-31T23:59:59Z",
  "options": {
    "baseline_kwh": 1200000,
    "post_intervention_kwh": 1100000,
    "eligible_spend_eur": 850000,
    "process_scope": true
  }
}
```

`type` ∈
`{monthly_consumption, co2_footprint, esrs_e1_csrd, piano_5_0_attestazione, conto_termico_2_0, certificati_bianchi_tee, audit_dlgs_102_2014}`.

Response 201 — a `Report` resource:

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "type": "piano_5_0_attestazione",
  "period_from": "2025-01-01T00:00:00Z",
  "period_to": "2025-12-31T23:59:59Z",
  "status": "generated",
  "payload": { ... },
  "generated_by": "user@example.it",
  "created_at": "2026-04-17T10:30:00Z",
  "updated_at": "2026-04-17T10:30:00Z"
}
```

#### Piano 5.0 attestazione payload schema

```json
{
  "attestazione": {
    "baseline_kwh": 1200000,
    "post_intervention_kwh": 1100000,
    "energy_reduction_pct": 8.33,
    "process_reduction_pct": 8.33,
    "site_reduction_pct": 0,
    "meets_process_threshold": true,
    "meets_site_threshold": false,
    "tax_credit_band": "6-10% (aliquota intermedia)",
    "eligible_amount_eur": 850000,
    "expected_credit_eur": 170000
  },
  "methodology": "Confronto consumi baseline vs post-intervento, normalizzato su produzione e gradi-giorno (EN 16247-3).",
  "normative_ref": "DL 19/2024, DM applicativo MIMIT-MASE 24/07/2024, Linee Guida GSE.",
  "signer": { "role": "EGE (UNI CEI 11339) o Auditor Energetico (EN 16247-5)", "placeholder": true },
  "generated_at": "2026-04-17T10:30:00Z"
}
```

#### CSRD ESRS E1 payload schema

```json
{
  "disclosure_standard": "ESRS E1 (CSRD — Dir. UE 2022/2464)",
  "data_points": [
    { "code": "E1-5", "description": "Consumo energetico totale da fonti non rinnovabili", "value": 812500, "unit": "kWh", "source": "GreenMetrics", "methodology": "ISPRA residual mix split" },
    { "code": "E1-6", "description": "Emissioni lorde Scope 1", "value": 632125.0, "unit": "kg CO2e" },
    { "code": "E1-6", "description": "Emissioni lorde Scope 2 (location-based)", "value": 306250.0, "unit": "kg CO2e" },
    { "code": "E1-7", "description": "Intensità GHG rispetto ai ricavi netti", "value": 0, "unit": "kg CO2e / €", "methodology": "Richiede input ricavi dal tenant" }
  ],
  "reporting_period": { "from": "2025-01-01", "to": "2025-12-31" },
  "generated_at": "2026-04-17T10:30:00Z"
}
```

### GET /api/v1/reports · GET /api/v1/reports/:id

List / retrieve persisted reports.

### GET /api/v1/reports/:id/download

Streams the rendered PDF (`Content-Type: application/pdf`).

### POST /api/v1/reports/:id/submit

Submits to the relevant authority portal (GSE / ENEA). Returns the updated
`{ id, status: "submitted", submitted_at }`.

---

## 7. Alerts

### GET /api/v1/alerts

```json
{
  "items": [
    { "id": "...", "kind": "consumption_anomaly", "severity": "warning", "message": "Consumo anomalo...", "triggered_at": "...", "resolved_at": null }
  ],
  "total": 1
}
```

### POST /api/v1/alerts/:id/ack · POST /api/v1/alerts/:id/resolve

Mark an alert acknowledged or resolved.

---

## 8. Emission Factors

### GET /api/v1/emission-factors

Returns active factor set. Example row:

```json
{ "code": "IT_ELEC_MIX_2023", "scope": 2, "category": "electricity_mix", "unit": "kWh", "kg_co2e_per": 0.250, "source": "ISPRA 2024 Rapporto 404", "valid_from": "2023-01-01T00:00:00Z", "valid_to": "2023-12-31T00:00:00Z", "version": "2024.1" }
```

### POST /api/v1/emission-factors (admin) · PUT /api/v1/emission-factors/:code

Admin-only write endpoints to extend or revise factors (usually run by the ISPRA
import job, but exposed for manual override).

---

## 9. Metrics

### GET /api/internal/metrics

Prometheus text format. Not versioned, not authenticated on internal networks.

Sample series: `greenmetrics_ingest_total`, `greenmetrics_report_generated_total`,
`greenmetrics_alerts_active`.

---

## 10. Error catalogue

| Status | Title | When |
|--------|-------|------|
| 400 | Bad Request | Invalid JSON, missing required parameter, bad date |
| 401 | Unauthorized | Missing / invalid / expired Bearer token |
| 403 | Forbidden | Role insufficient for endpoint |
| 404 | Not Found | Entity does not exist for tenant |
| 409 | Conflict | Unique constraint (POD / PDR already exists) |
| 422 | Unprocessable Entity | Business rule violation (e.g. period_to before period_from) |
| 429 | Too Many Requests | Rate-limit exceeded |
| 500 | Internal Server Error | Uncaught error — check `X-Request-ID` in logs |
| 503 | Service Unavailable | Dependency down (see `/api/ready`) |
