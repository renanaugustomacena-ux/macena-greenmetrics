# CLI Contract

**Doctrine refs:** Rule 14 (contract first), Rule 17 (DX), Rule 23 (tooling discipline).

This document is the stable interface for repository CLI targets. Operators script against these names; renaming or removing a target is a breaking change requiring CHANGELOG entry + 2-minor sunset window.

Sources of truth: `Taskfile.yaml` (top-level), `backend/Makefile` (backend-specific). Both expose the same mental model — `task <name>` and `make <name>` resolve identical operations.

## 1. Bootstrap + lifecycle

| Target | Behaviour |
|---|---|
| `task bootstrap` | Fresh-clone bootstrap. Generates `.env` with cryptographic secrets, brings stack up, applies migrations, smoke-tests. Idempotent. |
| `task up` | `docker compose up -d`. |
| `task down` | `docker compose down`. |
| `task build` | `docker compose build`. |
| `task smoke` | `curl /api/health`. |

## 2. Testing

| Target | Behaviour |
|---|---|
| `task test` / `make test` | All unit + integration tests. |
| `task test:unit` / `make test-unit` | Backend unit tests (race detector, coverage). |
| `task test:integration` / `make test-integration` | Backend integration tests via testcontainers. |
| `task test:e2e` / `make test-e2e` | Playwright E2E against ephemeral compose. |
| `task test:property` / `make test-property` | Property-based tests via gopter. |
| `task test:security` / `make test-security` | RBAC + RLS + boot-refusal tests. |
| `task test:conformance` / `make test-conformance` | Cross-portfolio invariants. |
| `make test-static` | AST-level checks (no float for money, no panic on user input). |
| `make bench` | k6 perf bench (alias `task bench`). |
| `make cover` | Coverage HTML report. |

## 3. Lint + format

| Target | Behaviour |
|---|---|
| `task lint` / `task precommit` / `make precommit` | Run all pre-commit hooks. |
| `make lint` | golangci-lint --fast. |
| `make vet` | go vet. |
| `make fmt` | gofmt + goimports. |
| `make tidy` | go mod tidy. |
| `task verify` / `make verify` | Full continuous-verification loop (lint + tests + security + policy). |

## 4. Migrations (goose)

| Target | Behaviour |
|---|---|
| `task migrate:up` / `make migrate-up` | Apply all pending migrations (forward-only). |
| `task migrate:down N=1` / `make migrate-down N=1` | Roll back N migrations. |
| `task migrate:status` / `make migrate-status` | Show migration status. |
| `task migrate:create NAME=...` / `make migrate-create NAME=...` | Create new migration file pair. |

## 5. Contracts

| Target | Behaviour |
|---|---|
| `task openapi:lint` / `make openapi-validate` | Validate canonical `api/openapi/v1.yaml`. |
| `task openapi:bundle` / `make openapi-bundle` | Bundle to single-file JSON. |

## 6. Policies + security

| Target | Behaviour |
|---|---|
| `task policy:check` / `make policy-check` | Run conftest + kubeconform locally. |
| `task security:scan` | gitleaks + trivy fs + govulncheck + osv. |

## 7. Operations

| Target | Behaviour |
|---|---|
| `task rotate:secrets` | Force ESO refresh across the namespace. |
| `task rotate:jwt` | Trigger quarterly JWT KID rotation workflow (`gh workflow run jwt-rotation.yml`). |
| `task dr:restore` | Disaster-recovery restore drill. **Refuses if env=production**. |
| `task cost:audit` / monthly cron | Run cost audit script; emit report. |
| `task audit:rejections` / `make audit-rejections` | Scan repo for keywords from `docs/adr/REJECTED.md`. |
| `task layers:doc` / `make layers-doc` | Regenerate `docs/LAYERS.md` from `docs/layers.yaml`. |

## 8. Help

| Target | Behaviour |
|---|---|
| `task --list` | List all task targets. |
| `make help` | Print Make target inventory with descriptions. |

## 9. Stability guarantees

- Names listed above are stable — renames require deprecation window of one minor release.
- Behaviour changes that break operator scripts require a CHANGELOG entry and a `BREAKING:` prefix on the commit.
- New targets may be added freely; removal requires ADR.

## 10. Anti-patterns rejected

- "Just script it inline in CI" — rejected. CI script must invoke a task target so operators can reproduce locally.
- "Two ways to do the same thing" — rejected. `make` and `task` aliases must resolve to identical behaviour.
- Hidden flags or env requirements not documented — rejected.
