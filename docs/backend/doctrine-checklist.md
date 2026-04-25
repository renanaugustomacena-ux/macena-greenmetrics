# Backend Doctrine Checklist

**Doctrine refs:** Rule 29 (role), Rule 31 (sequence), Rule 45 (quality threshold), Rule 47 (rationale), Rule 48 (termination).

This checklist is signed by the reviewer on every backend PR via a comment in the PR thread. Failure to defend any item escalates to platform office hours mentoring.

## How to use

Reviewer comments on the PR with this checklist filled. Each box is `[x]` (passed), `[ ]` (failed — block merge), or `n/a` (not applicable to this diff).

## Checklist

### Rule 30 — Backend as first-class system

- [ ] Process boundary touched is documented in `docs/backend/system-map.md`.
- [ ] Data ownership matrix not violated (e.g. only `auth_handler` writes `users`; only `audit_middleware` writes `audit_log`).
- [ ] Failure-domain considered (what degrades if this fails).

### Rule 32 — Domain-driven design

- [ ] Domain logic in `internal/domain/`; no `pgx`, `fiber`, or `zap` imports there.
- [ ] DTOs in `internal/api/v1/dto/`, not on domain types.
- [ ] Naming reflects ubiquitous language; no `placeholder-tenant` or magic strings.
- [ ] Mapper boilerplate present where domain ↔ DTO crossing needed.

### Rule 33 — Data is the system

- [ ] If new migration: paired `.up.sql`/`.down.sql` (or `.go` if templating); CAGG-aware down; no `DROP TABLE` on hypertable in down.
- [ ] If existing migration touched: PR reverts and adds new forward-only migration (forward-only invariant).
- [ ] Schema-version impact noted in PR description if tenant-visible shape changes.

### Rule 34 — Contract first

- [ ] OpenAPI updated in `api/openapi/v1.yaml` for any endpoint change.
- [ ] Generated `internal/api/v1/types.go` matches OpenAPI (`oapi-codegen --strict`).
- [ ] Validator tags present on every request struct (`validate:"required,..."`).
- [ ] `Bind[T]` helper used for body parsing; no manual `BodyParser` + nil checks.

### Rule 35 — Consistency and state guarantees

- [ ] Multi-statement operations wrapped in `pgxpool.Tx` with explicit isolation level.
- [ ] POST mutations require `Idempotency-Key` header in production.
- [ ] Strong vs eventual consistency stated for the change.
- [ ] Ordering needs (e.g. `ingest_seq`) considered.

### Rule 36 — Failure as normal

- [ ] Outbound calls wrapped with breaker (`internal/resilience/breaker.go`).
- [ ] Retry policy explicit (idempotent verbs only, capped attempts, backoff with jitter).
- [ ] Timeouts derived from `c.Context()`; per-endpoint budgets honoured.
- [ ] Backpressure: bounded channel + drop policy + Prometheus counter on saturation.
- [ ] Graceful degradation path documented (cached fallback, partial response, `data_freshness` stamp).

### Rule 37 — Performance as design constraint

- [ ] Latency budget stated (p50 / p95 / p99) and falls within `docs/backend/slo.md` table.
- [ ] If long-running: async via Asynq, returns 202 + Location.
- [ ] Contention points identified (bcrypt pool, pgx pool, rate limiter).
- [ ] Bench test added if perf-sensitive (`tests/bench/`); k6 if endpoint-level.

### Rule 38 — Scalability with intent

- [ ] Scaling axis identified (HTTP / ingestor / worker / WebSocket).
- [ ] Ceiling documented if new bottleneck.
- [ ] Cost implication noted (cents per 1k of operation).

### Rule 39 — Security as core

- [ ] Trust boundary touched cited from `docs/TRUST-BOUNDARIES.md`.
- [ ] Input validated at boundary.
- [ ] Auth required (JWT) and authorised (RBAC `RequirePermission`) on protected routes.
- [ ] Tenant context flows to DB via `InTxAsTenant(ctx, tenantID, fn)`.
- [ ] No plaintext compare on signature checks; constant-time only.
- [ ] Body limit honoured.
- [ ] No secret in logs (zap field redactor in scope).

### Rule 40 — Observability

- [ ] Logger called via `obs.Logger(c.Context())` so `trace_id` + `span_id` + `request_id` + `tenant_id` propagate.
- [ ] Custom metric added with explicit cardinality budget if new metric.
- [ ] Span coverage: HTTP handler ✓, pgx ✓, outbound HTTP ✓, ingestor poll ✓.
- [ ] Alert rule added to `monitoring/prometheus/rules/` if new failure mode + `runbook_url` annotation.

### Rule 41 — Concurrency discipline

- [ ] Goroutine has lifecycle owner (`Start` / `Stop` / `Wait`).
- [ ] Errgroup wraps related goroutines; no fire-and-forget outside `main`.
- [ ] Worker pool used (`panjf2000/ants`) for bounded resource use.
- [ ] `goleak.VerifyNone` in tests touching subsystem.

### Rule 42 — Resource lifecycle

- [ ] HTTP outbound uses shared `http.Client` from `internal/resilience/http.go`.
- [ ] DB connections released (deferred `Release` or `pgx.Conn.Release()`).
- [ ] File handles closed.
- [ ] WebSocket has read deadline + ping/pong + idle close.

### Rule 43 — Framework + infra awareness

- [ ] pgx pool budget respected (max 25; ceiling 250 RPS sustained per pod at 100 ms hold time).
- [ ] Fiber `c.Locals` set before middleware that reads them.
- [ ] K8s probe semantic respected: liveness restarts pod, readiness only removes from Service; do not make liveness DB-dependent.
- [ ] Timescale CAGG / retention interaction considered.

### Rule 44 — Testability + verification

- [ ] Unit test for new domain logic (table-driven, deterministic).
- [ ] Integration test for new repository method (testcontainers).
- [ ] Property test for new invariant.
- [ ] Coverage maintained ≥ 80% line, ≥ 90% on `internal/domain/`.
- [ ] No flaky tests introduced; if discovered, quarantined with issue.

### Rule 45 — Quality threshold

- [ ] All CI gates green (lint / test / integration / property / security / conformance / static / build / SBOM / Trivy / govulncheck / semgrep / CodeQL / osv / license / kubeconform / policy-gate / openapi-lint / openapi-compat).
- [ ] Mutation kill rate not regressed on `internal/domain/`.
- [ ] Heap diff < +10% vs main.

### Rule 46 — Rejection

- [ ] No anti-patterns from `docs/adr/REJECTED.md` introduced. If introduced: override ADR linked.
- [ ] No god service (single file > 400 LoC with > 5 unrelated methods on one struct).
- [ ] No implicit tenant scoping (RLS or explicit `InTxAsTenant`).
- [ ] No float for money (static check passes).
- [ ] No `panic()` on user input (CI grep gate passes).
- [ ] No `interface{}` / `any` return from repository (custom golangci-lint rule passes).

### Rule 47 — Decision rationale

- [ ] If non-trivial: ADR linked in PR description.
- [ ] ADR contains four-part Tradeoff Stanza.
- [ ] Alternatives considered + at least one realistic alternative justified.

### Rule 48 — Termination objective

- [ ] Author can articulate, without prompting, the consistency model, failure budget, latency target, isolation strategy, and trust boundary for the change.
- [ ] If reviewer had to provide that articulation, escalate to mentoring loop.

## Reviewer signature

```
Doctrine checklist signed by @<github-handle> on YYYY-MM-DD.
Failures: <list>. Mentoring escalation: <yes/no>.
```
