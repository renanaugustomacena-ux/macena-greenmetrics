# Platform Defaults

**Doctrine refs:** Rule 12 (platform thinking — opinionated defaults), Rule 14 (contracts), Rule 25 (quality threshold).

A new service that diverges from these defaults requires an ADR justifying the divergence.

## 1. Configuration

- 12-factor environment variables only. No bespoke config-management framework (REJ-04).
- Typed config struct in `internal/config/config.go`; `Load()` returns an error rather than panic.
- Sentinel detection in production: refuse to boot with placeholder secrets (`config.go:176-186`).
- JSON Schema for every config key in `docs/contracts/config.schema.json` (S2).

## 2. Logging

- Structured JSON via zap (production); human-readable + colour (development).
- Mandatory fields on every log line: `service`, `env`, `version`, `commit`, `request_id`, `trace_id`, `span_id`, `tenant_id`.
- Logger obtained via `obs.Logger(ctx)` so context fields propagate.
- No `fmt.Sprintf` log messages — structured fields only.
- No secret in logs — zap field redactor for `password`, `token`, `secret`, `authorization`.

## 3. Tracing

- OpenTelemetry SDK with OTLP gRPC exporter.
- Sample ratio 0.1 in production, 1.0 elsewhere (config default).
- Span coverage: HTTP handler (otelfiber), pgx (`tracelog.QueryTracer`), outbound HTTP (otelhttp), ingestor poll (manual).
- Every span carries `tenant.id` attribute.

## 4. Metrics

- Prometheus exposition at `/api/internal/metrics` (internal-only).
- Custom metrics defined in `internal/metrics/metrics.go` with explicit cardinality budget.
- Tenant ID label only on counters, never on histograms (use tenant tier instead).

## 5. Errors

- RFC 7807 Problem Details for every error response.
- `internal/handlers/errors.go` provides typed helpers: `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`, `Conflict`, `Unprocessable`, `TooManyRequests`, `Internal`, `Unavailable`.
- No `panic()` on user input; CI grep gate.

## 6. HTTP

- Fiber v2 with `BodyLimit: 4 MB` global; 16 MB per-route override on ingest endpoints.
- `ReadTimeout: 15s`, `WriteTimeout: 20s`, `IdleTimeout: 60s`.
- Recover middleware (panic recovery).
- requestid middleware (X-Request-ID).
- compress middleware.
- CORS allow-list from env.
- SecurityHeaders middleware (CSP, HSTS, X-Frame-Options).

## 7. Auth

- HS256 JWT pinned at validation site (`jwt.WithValidMethods([]string{"HS256"})`).
- `kid` claim required; rotation via dual-key window (S3).
- bcrypt for password hashing.
- IP+email lockout (`auth_lockout.go`).
- Sentinel JWT secret hard-refused in production at boot (Rule 19).

## 8. Authorisation

- RBAC via `RequirePermission(...)` middleware (S3).
- Permission registry in `internal/security/rbac.go`.
- No implicit role escalation.

## 9. Multi-tenancy

- JWT claim → middleware → `c.Locals("tenant_id")`.
- DB context via `InTxAsTenant(ctx, tenantID, fn)` (S3).
- Postgres RLS policies on every tenant-scoped table (S3).
- App-level WHERE filter as defence in depth (`current_setting('app.tenant_id')` is the second line).

## 10. Idempotency

- `Idempotency-Key` header required on POST in production.
- Storage in `idempotency_keys` Timescale hypertable, 24h retention.
- Conflict on differing request hash returns 422.

## 11. Database

- pgx/v5 only. No ORM.
- Parameterised queries always (no string concatenation).
- Connection pool max 25 / min 2 / max-lifetime 30 m / health check 1 m.
- Per-Tx timeout 5 s default.
- TimescaleDB hypertables for time-series; CAGGs for aggregates; retention policies for raw + aggregated.

## 12. Migrations

- pressly/goose v3 (ADR-005).
- `.up.sql` + `.down.sql` pairs.
- Forward-only in production; CI fails edits to applied migrations.
- Down scripts may be lossy; documented.

## 13. Resilience

- Per-host circuit breakers via `sony/gobreaker/v2` (S4).
- Retry via `cenkalti/backoff/v5` — idempotent verbs only, capped attempts, jittered backoff.
- Timeouts derived from `c.Context()` with explicit per-endpoint budgets.
- Backpressure via bounded channel + drop policy + Prometheus counter.
- Graceful degradation: cached fallback + `data_freshness` stamp.

## 14. Observability output

- Health envelope `{status, service, version, uptime_seconds, time, dependencies}` (CLAUDE.md invariant).
- Three health endpoints: `/api/health` (degraded-tolerant), `/api/ready` (strict), `/api/live` (always 200 once boot complete).

## 15. Cross-portfolio invariants

- Money: `(amount_cents int64, currency ISO-4217 string)`. Never float.
- Timestamp: RFC 3339 UTC with offset.
- `tenant_id`: UUIDv4.
- Errors: RFC 7807.
- Events: CloudEvents 1.0.

## 16. Container

- Distroless nonroot image (`gcr.io/distroless/static-debian12:nonroot`).
- Multi-stage build.
- Digest-pinned base.
- HEALTHCHECK directive present.
- USER directive non-root.
- ENTRYPOINT in JSON-array form (avoid `/bin/sh -c`).

## 17. Kubernetes

- Pod Security Standard `restricted` enforced.
- NetworkPolicy default-deny + selective egress allow-list.
- Security context: `runAsNonRoot`, `runAsUser=65532`, `fsGroup=65532`, `seccompProfile.type=RuntimeDefault`, `allowPrivilegeEscalation=false`, `readOnlyRootFilesystem=true`, `capabilities.drop=[ALL]`.
- Resources: requests + limits set on every container.
- Probes: liveness + readiness + startup on long-running workloads.
- HPA + PDB for stateful + stateless deployments.
- Topology spread across AZs.
- ServiceAccount IRSA-bound; `automountServiceAccountToken: false` unless needed.

## 18. Supply chain

- GitHub Actions SHA-pinned (Dependabot weekly).
- Container base images digest-pinned.
- Cosign keyless sign + SBOM attest + SLSA L2 provenance attest.
- Kyverno `verify-images` admission denies unsigned.
- Trivy image scan post-build, fail HIGH/CRITICAL.

## 19. CI/CD

- `pre-commit-ci` first job — fast fail.
- Required status checks per `docs/PLATFORM-INITIATIVE-WORKFLOW.md`.
- CD via Argo CD GitOps; no manual `kubectl` in production.
- Argo Rollouts canary for stateless deploys; AnalysisTemplate reads SLO burn-rate.

## 20. Documentation

- ADRs for every non-trivial decision (`docs/adr/`).
- Tradeoff Stanza required.
- PR template enforces fields.
- Runbooks for every failure mode.
- Quarterly platform office hours.
