# SLO — GreenMetrics

## 1. Service Level Objectives

| Indicator | Target | Window | Error budget |
|-----------|--------|--------|--------------|
| Availability (`/api/health` 200) | 99,5% | 30-day rolling | 3h 39m |
| Read latency p95 (`/api/v1/readings/aggregated`) | ≤ 500 ms | 30-day rolling | 5% |
| Write latency p95 (`/api/v1/readings/ingest`) | ≤ 1 000 ms | 30-day rolling | 5% |
| Error rate 5xx | ≤ 1% | 5-min rolling | 0 |
| Continuous-aggregate freshness | ≤ 30 minutes lag | continuous | — |

## 2. Data retention

- Raw `readings` hypertable: 90 days (chunk interval 1 day, compressed after
  7 days segmented by `meter_id`).
- `readings_15min` continuous aggregate: 365 days.
- `readings_1h` continuous aggregate: 3 years.
- `readings_1d` continuous aggregate: 10 years.

## 3. RPO / RTO

- **RPO** (Recovery Point Objective): 1 hour — achieved via Timescale WAL
  streaming to a read replica (production deployment) + nightly `pg_dump`.
- **RTO** (Recovery Time Objective): 4 hours — full Timescale restore from
  the latest dump into a cold replica, then cut-over DNS.

## 4. Capacity planning

The 1-day chunk interval + CAGG hierarchy is sized for ~10 tenants × 10
meters × 4 readings/minute (≈ 5.7M rows/day). The hypertable compression
ratio is expected ≥ 10× on cold chunks.

## 5. Alerts (Grafana)

- Error rate > 1% for 5 min → warning (Slack `#greenmetrics-ops`).
- Error rate > 5% for 5 min → critical (page on-call).
- `timescaledb: degraded` on `/api/health` → critical.
- CAGG lag > 60 min → warning.
- Saturation CPU > 80% for 15 min → warning.
- Daily energy delta per tenant > 3σ from 30-day baseline → warning
  (consumption anomaly rule in `alert_engine.go`).
