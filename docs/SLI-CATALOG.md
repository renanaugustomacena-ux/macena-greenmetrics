# SLI Catalog

**Doctrine refs:** Rule 18, Rule 40, Rule 58.
**Owner:** `@greenmetrics/sre`.
**Cross-ref:** `docs/SLO.md` for the contractual targets.

Per Google SRE workbook: each SLO has a measurable SLI, a recording rule, an alert with multi-window-multi-burn-rate, and a `runbook_url` annotation.

## 1. Availability — backend API

- **SLI:** `success_count / total_count` over rolling 5m, where success is HTTP status not in `5xx`.
- **PromQL (recording rule `gm:slo_availability:ratio_5m`):**

  ```
  sum by (service) (rate(gm_http_requests_total{status!~"5.."}[5m]))
    / clamp_min(sum by (service) (rate(gm_http_requests_total[5m])), 1)
  ```

- **SLO target:** 99.5% over 30-day rolling window.
- **Burn-rate alert:** see `monitoring/prometheus/rules/slo.rules.yaml` — fast-burn (14.4× over 1h with 6× over 6h confirm) → critical.
- **Runbook:** `docs/runbooks/db-outage.md`.

## 2. Latency — POST /api/v1/readings/ingest

- **SLI:** p99 of `gm_http_request_duration_seconds_bucket{path="/api/v1/readings/ingest"}` over 5m.
- **SLO target:** p99 ≤ 120 ms (per `docs/backend/slo.md`).
- **Alert:** `P99IngestBudget` after 10m breach.
- **Runbook:** `docs/runbooks/capacity-spike.md`.

## 3. Latency — POST /api/v1/auth/login

- **SLI:** p99 of `gm_http_request_duration_seconds_bucket{path="/api/v1/auth/login"}` over 5m.
- **SLO target:** p99 ≤ 400 ms (bcrypt cost factor 12 baseline ~60 ms; allow 6× headroom).
- **Alert:** `P99LoginBudget` after 5m breach (warning at 500ms).
- **Runbook:** `docs/runbooks/capacity-spike.md`.

## 4. Latency — async report acknowledgement

- **SLI:** p99 of `POST /api/v1/reports` (returns 202) over 5m.
- **SLO target:** p99 ≤ 100 ms (regardless of report size — async).
- **Alert:** TBD; covered by general `APIErrorRateHigh` and per-endpoint p99 monitor.

## 5. CAGG freshness

- **SLI:** `time() - gm_cagg_last_refresh_timestamp_seconds{view="readings_15min"}`.
- **SLO target:** ≤ 60 s for `readings_15min`, ≤ 5 min for `readings_1h`, ≤ 1 h for `readings_1d`.
- **Alert:** `CAGGRefreshLag` > 120 s for 5m → warning.
- **Runbook:** `docs/runbooks/db-outage.md`.

## 6. Ingest success rate per protocol

- **SLI:** `rate(gm_ingest_readings_total{result="success", protocol=X}[5m]) / rate(gm_ingest_readings_total{protocol=X}[5m])`.
- **SLO target:** ≥ 99.0% per protocol over 30-day rolling window.
- **Alert:** Modbus stalled (zero rate 15m) → warning. Drop counter > 0 → critical.
- **Runbooks:** `docs/runbooks/ingestor-crash-loop.md`, `docs/runbooks/pulse-webhook-flood.md`.

## 7. Login lockout false-positive rate

- **SLI:** `rate(gm_login_locked_total{reason="false_positive"}[1h]) / rate(gm_login_attempted_total[1h])`.
- **SLO target:** ≤ 0.1% (FP rate as fraction of legitimate attempts).
- **Alert:** baseline anomaly > 5× over 1d (`LoginFailureSpike`).
- **Runbook:** `docs/runbooks/tenant-data-leak.md`.

## 8. Secret rotation freshness

- **SLI:** `time() - aws_secretsmanager_secret_last_rotated_timestamp_seconds`.
- **SLO target:** ≤ 90 days for `greenmetrics/prod/jwt`; ≤ 90 days for DB credentials.
- **Alert:** `JWTSecretAgeExceeded` (90d) → critical.
- **Runbook:** `docs/runbooks/jwt-secret-rotation.md`.

## 9. Certificate expiry buffer

- **SLI:** `cert_exporter_not_after - time()`.
- **SLO target:** ≥ 14 days at all times (cert-manager renews 7d before expiry).
- **Alert:** `CertificateExpirySoon` < 14 days for 1h → warning.
- **Runbook:** `docs/runbooks/cert-rotation.md`.

## 10. RED + USE per service

### RED (request-driven)

For each `service` label:

- **Rate:** `sum by (service) (rate(gm_http_requests_total[5m]))`.
- **Errors:** `sum by (service) (rate(gm_http_requests_total{status=~"5.."}[5m])) / sum by (service) (rate(gm_http_requests_total[5m]))`.
- **Duration:** `histogram_quantile(0.99, sum by (le, service) (rate(gm_http_request_duration_seconds_bucket[5m])))`.

### USE (resource-driven)

- **Utilisation:** pgx pool — `gm_db_pool_acquire_duration_seconds` p99; CPU — `node_cpu_seconds_total{mode!="idle"}`; Memory — container working set.
- **Saturation:** ingest queue — `gm_ingest_queue_depth`; HPA — `kube_horizontalpodautoscaler_status_current_replicas / kube_horizontalpodautoscaler_spec_max_replicas`.
- **Errors:** request 5xx; pgx errors; circuit breaker state; ingest dropped count.

## 11. Cardinality budget

- `tenant_id` label only on **counters** (`gm_ingest_readings_total`, `gm_alert_fired_total`).
- **Histograms** (`gm_http_request_duration_seconds`, `gm_db_pool_acquire_duration_seconds`) use `path` / `service` only.
- New metric proposing tenant_id on a histogram → ADR explaining why cardinality cost is acceptable.

Total expected series count today: ~6k at 50 tenants × 4 protocols × 16 buckets per histogram. Re-baseline at 500 tenants.
