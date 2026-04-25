# Reliability Model

**Doctrine refs:** Rule 15, Rule 36, Rule 59 (reliability + security coupled), Rule 65.
**Owner:** `@greenmetrics/sre`, `@greenmetrics/platform-team`.

## 1. Failure modes (enumerated)

| Failure | Blast radius | Detection | Mitigation | Secure-degradation path |
|---|---|---|---|---|
| TimescaleDB primary loss | API readiness flips degraded; pods removed from Service | `DBPrimaryDown`; `/api/ready` 503 | Multi-AZ failover (RDS auto, < 60s); manual via `aws rds reboot-db-instance --force-failover` | Read-only mode (`READ_ONLY_MODE=true`); ingest queue spills if enabled |
| AZ outage | One AZ's pods unavailable; others continue | Standard K8s topology spread + HPA | K8s reschedules to surviving AZs; RDS Multi-AZ failover | Capacity halved; degrade gracefully |
| Region (eu-south-1) outage | Full service down | `AbsentMetricsDeadMan`; AWS Health | DR replica promote in eu-west-1 (4h RTO); see `docs/runbooks/region-failover.md` | DR region fully operational once promoted |
| Modbus simulator / host down | Per-host Modbus ingest stops | `ModbusIngestStalled`; `BreakerOpen{name~modbus_*}` | Disable failing host (M1 in `docs/runbooks/ingestor-crash-loop.md`); breaker opens automatically | Pulse + manual ingest paths continue; CAGGs lag for that meter |
| ISPRA endpoint 5xx | Factor refresh fails | `BreakerOpen{name=ispra_emission_factors}`; `ExternalAPIFallback{api=ispra}` | Breaker opens; backend serves cached factors with `data_freshness:"cached_<n>h"` | Reports stamped cached; customer notified via banner |
| Terna / E-Distribuzione 5xx | Specific lookup fails | Breaker opens | Same as ISPRA — cached fallback | Same |
| Grafana down | Dashboards unreachable | Pod CrashLoopBackOff | Restart pod; check ESO sync; `docs/runbooks/grafana-down.md` | **Alerts route via Alertmanager directly — not affected** |
| Argo CD compromised | Arbitrary cluster state possible | Falco rule + audit log on Argo SA | Restrict Argo RBAC; OIDC SSO; Cosign verify at admission acts independently | Kyverno admission denies unsigned images regardless of Argo intent |
| Kyverno admission webhook offline | New pod creation blocked | `KyvernoAdmissionSlow`; `kubectl get clusterpolicy` denies | Restart Kyverno; do not disable webhook in prod without IC | Existing pods continue; new deploys queued |
| Redis (Asynq) outage | New jobs cannot enqueue | Health check fails on `/api/ready` | Restart Redis; in-flight jobs retry on resume | `POST /v1/reports` returns 503 Retry-After; ingest path unaffected |
| Cert-manager renewal failure | TLS expires | `CertificateExpirySoon` 14d ahead | `docs/runbooks/cert-rotation.md` | Continue serving until expiry; if expired, browser warnings — comms |
| ESO sync failure | New secret rotations don't propagate | `ESONotSynced` | `docs/runbooks/secret-rotation.md` | Existing pod env retains last value |
| OTel collector down | Trace export fails | OTel SDK retries internally | Pod restart; bounded buffer in SDK | Logs + metrics paths continue |
| Prometheus down | Metric collection halts | `AbsentMetricsDeadMan` (dead-man-switch via Alertmanager direct) | Restart kps; check storage | Dashboards stale; alerts on Prometheus rules silent |
| API replica panic | Per-replica request error | Recover middleware catches; pod stays alive | Investigate; alert if rate > 1% over 5m | K8s liveness restarts pod if necessary |

## 2. Cascading failure analysis

### 2.1 DB outage cascade

```
DB primary loss
  → pgx pool acquire fails
  → API requests return 5xx
  → APIErrorRateHigh fires
  → SLO error budget burn (multi-window-multi-burn-rate)
  → Argo Rollouts AnalysisTemplate fires → auto-rollback (if recent deploy)
  → Customer impact
```

Containment: Multi-AZ failover < 60s; READ_ONLY_MODE for sustained outages.

### 2.2 Modbus host overload cascade

```
Modbus simulator slow
  → Polling goroutines block on TCP
  → Without breaker: goroutine count climbs → memory pressure → OOMKill
  → With breaker: open after 5 failures, rejects further calls in window
  → Bounded ingest channel saturates → dropped readings → IngestDropped fires
  → Backpressure: HTTP ingest path serves 503 Retry-After
```

Containment: per-host breaker (Rule 36) + bounded queue + drop policy + worker pool (`panjf2000/ants`).

### 2.3 Identity rotation cascade

```
JWT rotation runs but ESO doesn't sync
  → Pods sign with new kid v4 (env)
  → Pods validate against old JWT_KIDS_VALID (memory) → reject new tokens
  → Login storm; users see invalid_token loop
```

Containment: dual-key overlap window 24h; rotation workflow waits for ESO confirm before purging old kid.

## 3. RPO / RTO

| Service | RPO | RTO | Mechanism |
|---|---|---|---|
| TimescaleDB | 1 h | 4 h | RDS automated backup every 4h; cross-region replica + DR procedure |
| Application state (config, secrets) | 0 (idempotent) | < 5 min | Argo CD reconciliation; ESO refresh |
| Audit log | 0 (synchronous write) | n/a | Object Lock 5y on S3 audit bucket |
| Reports archive (S3) | 4 h | 4 h | Cross-region replication |
| Kubernetes manifests | 0 (gitops) | < 5 min | Argo CD App-of-Apps |

## 4. Quarterly chaos validation

`docs/CHAOS-PLAN.md` enumerates monthly Game Day experiments:

- DB pod kill (verify M1 db-outage runbook).
- Modbus simulator timeouts (verify breaker + degradation).
- OCPP CS unreachable (verify per-CP breaker + reconnect).
- Grafana down (verify Alertmanager continues).
- Timescale CAGG lag (verify alert fires + manual refresh procedure).
- Network partition between API and DB.

Outcomes recorded in `docs/CHAOS-LOG.md`.

## 5. Anti-patterns rejected

- Single-AZ deploy — REJ-09.
- Multi-region active-active in v1 — REJ-09.
- Disable breakers "to make tests pass" — REJ.
- Deploy without canary — Argo Rollouts gates.
- Rely on K8s liveness probe for DB connectivity — that turns a Timescale blip into a fleet-wide restart storm.
