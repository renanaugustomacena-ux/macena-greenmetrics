# GreenMetrics Quality Bar

**Doctrine refs:** Rule 25 (Platform quality threshold), Rule 45 (Backend quality threshold), Rule 65 (DevSecOps quality threshold — regulated industry).

These are non-negotiable invariants. PRs violating any of them are blocked, regardless of how convenient the violation seems. Override requires an explicit ADR + `override-allowed` label + secops sign-off.

## 1. Cross-portfolio invariants (CLAUDE.md)

| Invariant | Enforcement |
|---|---|
| Money = `(amount_cents int64, currency ISO-4217 string)` — never float | `tests/static/no_float_money_test.go` (S5) |
| Timestamp = RFC 3339 UTC with offset | `tests/conformance/rfc3339utc_test.go` (S5) |
| `tenant_id` = UUIDv4 | `tests/conformance/uuidv4_test.go` (S5) |
| Errors = RFC 7807 ProblemDetails | `tests/conformance/rfc7807_test.go` (S5) |
| Events = CloudEvents 1.0 envelope | event schemas in `docs/contracts/events/` validated in CI |
| Health envelope `{status, service, version, uptime_seconds, time, dependencies}` | `tests/conformance/health_test.go` (S5) |

## 2. Code quality

- Backend coverage ≥ 80% line; ≥ 90% on `internal/domain/`.
- Mutation kill rate ≥ 70% on `internal/domain/` (quarterly `gremlins`).
- No `panic()` on user input (`grep -rn 'panic(' internal/handlers/` returns 0 in CI).
- No `interface{}` / `any` returns from repository (custom golangci-lint rule).
- No god service (single struct file > 400 LoC with > 5 unrelated methods).
- gofmt + goimports + golangci-lint --fast clean on every commit.

## 3. Security

- gosec / CodeQL clean.
- govulncheck / Trivy FS / Trivy image / osv-scanner: 0 HIGH/CRITICAL unaccepted.
- gitleaks: 0 unallowlisted findings.
- Cosign signature verifies on every deployed image; Kyverno admission gates.
- SLSA L2 provenance attached; L3 plan dated.
- License allowlist clean (`LICENSES.allowed`); deny GPL-3.0+ in commercial path.
- Sentinel JWT secret hard-refused at boot in production (config.go:176-194).
- Postgres RLS enforced + RBAC at middleware + app-level WHERE filter (defence in depth).
- Body limit on Fiber 4 MB / 16 MB ingest.
- Constant-time signature compare on Pulse webhook.

## 4. Observability

- Every log line carries `service`, `env`, `version`, `commit`, `request_id`, `trace_id`, `span_id`, `tenant_id`.
- Every alert has `runbook_url` annotation pointing to `docs/runbooks/<name>.md`.
- OTel sample ratio = 0.1 in production; trace storm prevented.
- Custom metrics defined with explicit cardinality budget; tenant_id only on counters.
- Health envelope `{status, service, version, uptime_seconds, time, dependencies}`.

## 5. Reliability

- Triple health probes (liveness / readiness / startup) on every long-running workload.
- Graceful shutdown ≤ 30 s.
- Per-host circuit breakers on every outbound call.
- Bounded ingest queue + drop policy + Prometheus counter.
- Async report generation; OLTP pool not blocked on long jobs.
- RPO 1 h / RTO 4 h documented in `docs/SLO.md`.

## 6. Container + K8s

- Distroless nonroot base; multi-stage; digest-pinned.
- PSS `restricted` enforced.
- NetworkPolicy default-deny.
- `runAsNonRoot`, `readOnlyRootFilesystem`, `allowPrivilegeEscalation: false`, `capabilities.drop: ALL`, `seccompProfile.type: RuntimeDefault`.
- Resources requests + limits set on every container.
- Probes triple set on every long-running workload.
- ServiceAccount IRSA-bound; `automountServiceAccountToken: false` unless required.

## 7. Data

- pgx parameterised queries (no concat).
- pgx pool 25/2/30m; per-Tx 5 s timeout.
- TimescaleDB hypertables + CAGGs + retention policies versioned via goose.
- Forward-only migrations in production.
- Lossy down scripts documented.

## 8. Contracts

- Hand-written `api/openapi/v1.yaml` is source of truth; oapi-codegen generates Go types.
- `redocly lint` green on every PR.
- v1 backward compat test green (`tests/contracts/v1_compat_test.go`).
- CloudEvents schemas live in `docs/contracts/events/`.
- Config schema in `docs/contracts/config.schema.json`; `ajv validate` green.

## 9. Process

- ADR for every non-trivial decision; four-part Tradeoff Stanza enforced by markdownlint.
- PR template fields populated (linked issue, ADR, doctrine rules, tradeoff, backend addendum, runbook update, CHANGELOG).
- CODEOWNERS routes review; required reviewers per path.
- Conventional Commits.
- Pre-commit installed locally; CI mirror catches bypassers.
- Quarterly platform office hours; quarterly DevSecOps review; quarterly architectural review.

## 10. Regulatory

- CSRD/ESRS E1 evidence pack exportable (`docs/COMPLIANCE/CSRD.md`, S5).
- Piano 5.0 attestazione deterministic (`tests/property/aggregate_invariants_test.go`).
- D.Lgs. 102/2014 audit log queryable.
- GDPR DSAR endpoint + integration test.
- NIS2 24h/72h notification template (`docs/INCIDENT-RESPONSE.md`).
- Italian-compliance citations annotated with `MITIGATES: RISK-NNN`; re-verified annually against primary sources.

## 11. The line

A reviewer can produce, on demand, the evidence pack for a CSRD audit using only:

- `docs/RISK-REGISTER.md`
- `docs/SECOPS-CHARTER.md`
- `docs/THREAT-MODEL.md`
- `docs/INCIDENT-RESPONSE.md`
- `docs/SUPPLY-CHAIN.md`
- `docs/ITALIAN-COMPLIANCE.md`
- `policies/conftest/`, `policies/kyverno/` bundles
- `audit_log` table exports
- Cosign signature verification of deployed images

If any of those is missing or stale: the bar is broken. Fix it before the next deploy.
