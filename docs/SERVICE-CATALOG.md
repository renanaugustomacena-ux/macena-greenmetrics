# Service Catalog

**Doctrine refs:** Rule 12 (platform thinking), Rule 14 (contracts), Rule 16 (scalability), Rule 18 (observability).
**Maintainer:** `@greenmetrics/platform-team`.

The catalog is the source of truth for what services exist, who owns them, what they expose, and what their operational targets are.

## 1. Service inventory

| ID | Name | Owner | Type | Contract | Image | SLO source |
|---|---|---|---|---|---|---|
| svc-01 | `backend-api` | app-team | HTTP API (Fiber + pgx) | `api/openapi/v1.yaml` | `ghcr.io/<owner>/greenmetrics-backend` | `docs/SLO.md`, `docs/backend/slo.md` |
| svc-02 | `frontend-web` | app-team | SvelteKit SSR | implicit (uses `lib/api.ts`) | `ghcr.io/<owner>/greenmetrics-frontend` | `docs/SLO.md` |
| svc-03 | `ingestor-modbus` | app-team | goroutine in backend | `internal/services/modbus_ingestor.go` | colocated in `backend-api` | `docs/SLI-CATALOG.md` (S4) |
| svc-04 | `ingestor-mbus` | app-team | goroutine in backend | `internal/services/mbus_ingestor.go` | colocated | same |
| svc-05 | `ingestor-sunspec` | app-team | goroutine in backend | `internal/services/sunspec_profile.go` | colocated | same |
| svc-06 | `ingestor-pulse` | app-team | HTTP webhook handler | `POST /api/v1/pulse/ingest` | colocated | same |
| svc-07 | `ocpp-client` | app-team | WebSocket client | `internal/services/ocpp_client.go` | colocated | same |
| svc-08 | `worker-reports` | app-team | Asynq worker (`cmd/worker`) | `docs/contracts/events/report.generated.v1.json` (S2) | `ghcr.io/<owner>/greenmetrics-backend` (same image, different command) | `docs/backend/slo.md` |
| svc-09 | `alert-engine` | app-team | function in `backend-api` | `internal/services/alert_engine.go` | colocated | `docs/SLI-CATALOG.md` |
| svc-10 | `db-migrations` | platform-team | one-shot Job (`cmd/migrate`) | `backend/migrations/` | `ghcr.io/<owner>/greenmetrics-backend` | n/a |
| svc-11 | `simulator` | app-team | Modbus TCP server | `cmd/simulator/main.go` | `ghcr.io/<owner>/greenmetrics-simulator` | dev/test only |
| svc-12 | `external-secrets-operator` | secops | Helm-managed | upstream chart | upstream image | upstream SLO |
| svc-13 | `argocd` | platform-team | Helm-managed | upstream chart | upstream image | `docs/runbooks/argocd-down.md` (S4) |
| svc-14 | `prometheus` | sre | Helm-managed (kube-prometheus-stack) | upstream chart | upstream image | self-monitored |
| svc-15 | `alertmanager` | sre | Helm-managed | upstream chart | upstream image | self-monitored + dead-man-switch |
| svc-16 | `loki` | sre | Helm-managed | upstream chart | upstream image | self-monitored |
| svc-17 | `tempo` | sre | Helm-managed | upstream chart | upstream image | self-monitored |
| svc-18 | `otel-collector` | sre | DaemonSet | `monitoring/otel-collector/config.yaml` | upstream image | n/a |
| svc-19 | `falco` | secops | DaemonSet | `policies/falco/greenmetrics-rules.yaml` | upstream image | self-monitored |
| svc-20 | `cert-manager` | platform-team | Helm-managed | upstream chart | upstream image | TLS-expiry alert |
| svc-21 | `kyverno` | secops | Helm-managed | `policies/kyverno/` | upstream image | webhook-latency alert |
| svc-22 | `redis` (Asynq backend) | platform-team | container | `docker-compose.yml`, K8s StatefulSet | `redis:7-alpine@sha256:...` | health endpoint |
| svc-23 | `timescaledb` | platform-team | RDS PG16 + Timescale ext. | `terraform/modules/rds-timescale` | upstream image | RDS SLA |
| svc-24 | `grafana` | sre | container + provisioning | `grafana/provisioning/` | upstream image | health endpoint |
| svc-25 | `landing-page` | app-team | static HTML/CSS | `landing-page/` | n/a (CDN-hosted) | static, no SLO |

## 2. Per-service SLO target (default until per-service overrides ship)

- Availability ≥ 99.5% over 30-day rolling window.
- Latency p99 per endpoint per `docs/backend/slo.md` table.
- Error budget burn-rate: page on 14.4× budget burn over 1h or 6× over 6h.

## 3. Cross-service contracts

| Contract | Producer | Consumer | Type | Source of truth |
|---|---|---|---|---|
| `it.greenmetrics.reading.ingested.v1` | backend-api / ingestor-* | worker-reports, alert-engine | CloudEvents | `docs/contracts/events/reading.ingested.v1.json` (S2) |
| `it.greenmetrics.alert.fired.v1` | alert-engine | (consumer placeholder) | CloudEvents | `docs/contracts/events/alert.fired.v1.json` (S2) |
| `it.greenmetrics.report.generated.v1` | worker-reports | (frontend, customer notification) | CloudEvents | `docs/contracts/events/report.generated.v1.json` (S2) |
| `it.greenmetrics.tenant.created.v1` | backend-api | (provisioning, billing) | CloudEvents | (S2) |

## 4. Onboarding a new service

1. ADR via `docs/PLATFORM-INITIATIVE-WORKFLOW.md`.
2. Add row to this catalog.
3. Define contract in `api/openapi/v1.yaml` (HTTP) or `docs/contracts/events/` (event).
4. Add CODEOWNERS entry.
5. Add SLO entry to `docs/SLO.md`.
6. Add per-service Prometheus rules to `monitoring/prometheus/rules/`.
7. Add per-service runbook to `docs/runbooks/` if it has its own failure modes.
8. Add image to Cosign signing pipeline; add to Kyverno `verify-images.yaml` allowlist.

## 5. Anti-patterns rejected

- Service split before measured bottleneck (REJ-05).
- Generic "service framework" abstraction (Rule 13).
- Mesh control plane for ≤6 services (REJ-01).
