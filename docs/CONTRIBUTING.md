# Contributing to GreenMetrics

**Doctrine refs:** Rule 17 (DX), Rule 23 (tooling discipline), Rule 52 (shift-left), Rule 68 (termination objective).

This guide covers the bootstrap, dev loop, contribution flow, and the doctrine your PR will be reviewed against. Read this before your first PR.

## 1. Bootstrap

Fresh-clone bootstrap target ≤ 5 min:

```bash
git clone <repo> && cd GreenMetrics
./scripts/dev/bootstrap.sh
```

What the bootstrap script does (idempotent, safe to re-run):

1. `cp .env.example .env` if `.env` missing.
2. Generate `JWT_SECRET`, `POSTGRES_PASSWORD`, `GRAFANA_ADMIN_PASSWORD` via `openssl rand -base64 48`.
3. `docker compose up -d` — TimescaleDB, backend, frontend, Grafana, simulator.
4. Wait for `/api/health` to return `200`.
5. Run migrations via `goose -dir backend/migrations postgres "$DATABASE_URL" up`.
6. Smoke `curl http://localhost:8082/api/health | jq .`.
7. Print next-step tips (login, dashboards, runbooks).

If you prefer a containerised dev environment, open this repo in VSCode + Dev Containers (Microsoft extension) — the `.devcontainer/devcontainer.json` ships every tool the CI uses (Go 1.26, Node 20, postgres-client, pre-commit, conftest, terraform, tflint, tfsec, kubeconform, hadolint, actionlint, k6, cosign, syft, trivy, gosec, k9s, kubectl, helm, argocd-cli).

Nix users: `nix develop` against `flake.nix` for the equivalent shell.

## 2. Daily dev loop

```bash
# Run backend in hot-reload mode.
make debug-backend

# Run frontend in dev mode.
cd frontend && npm run dev

# Run all unit tests.
make test

# Run only the suite touching the package you changed.
go test ./internal/handlers/...

# Run integration tests against testcontainers.
make test-integration

# Run E2E suite against ephemeral compose.
make test-e2e

# Run k6 perf bench locally.
make bench

# Render Mermaid diagrams from docs/layers.yaml.
make layers-doc

# Lint everything that pre-commit will lint.
make precommit
```

`make help` prints the full target inventory.

## 3. Pre-commit hooks (mandatory)

Install once:

```bash
pip install pre-commit
pre-commit install
```

Hooks run on every commit. Total time target <5s on cache hit. Bypassing with `--no-verify` is **not** acceptable — if a hook is broken, fix it or open an issue tagged `tooling-bug`. CI mirrors the same hook set as the first job, so you cannot escape by pushing.

What runs:

- gitleaks (secret scan)
- golangci-lint --fast, gofmt, goimports, go-vet
- prettier, eslint
- shellcheck
- yamllint, markdownlint
- hadolint (Dockerfile)
- terraform fmt + validate + tflint + tfsec
- actionlint (GitHub Actions)
- kubeconform (K8s manifests)
- conftest (audit mode locally; CI enforces)
- conventional-pre-commit (commit message format)

Commit messages follow Conventional Commits: `feat(scope): summary`, `fix(scope): summary`, `chore(deps): bump x to y`, etc.

## 4. Contribution flow

1. Open an issue or pick one from the backlog. Tag the doctrine rules involved.
2. Branch from `main`: `git switch -c feat/<short-slug>`.
3. Write tests first (Rule 24 testability, Rule 44 backend testability). The PR template asks for test plan.
4. Implement. Run `make verify` locally before pushing.
5. Push and open a PR using the template at `.github/PULL_REQUEST_TEMPLATE.md`.
6. The CODEOWNERS file routes review to the right team. Do not request review manually unless overriding.
7. Required checks must be green: lint, test, integration, build, security (Trivy + govulncheck + osv + gitleaks + semgrep + CodeQL), policy gate (conftest + kubeconform + Kyverno admission), SBOM (Syft), Cosign verify on staging deploys.
8. For platform changes (`/k8s`, `/terraform`, `/backend/migrations`, `/api/openapi/`, `/.github/workflows/`, `/policies/`): an ADR is required. PR description must link the ADR file.
9. Squash-merge on green review. Squash commit message inherits the PR title (Conventional Commits format).

## 5. The doctrine your PR will be reviewed against

GreenMetrics is governed by a 60-rule operational doctrine spanning Platform Engineering (Rules 9–28), Advanced Backend Engineering (Rules 29–48), and DevSecOps Engineering (Rules 49–68). Full text in `~/.claude/plans/my-brother-i-would-flickering-coral.md`. The ten things you will be asked about most:

1. **Rule 14 — Contract first.** Hand-written `api/openapi/v1.yaml` is the source of truth. Code generated, not the other way around.
2. **Rule 27 / 47 / 67 — Tradeoff stanza.** Every non-trivial decision states what it solves, optimises for, sacrifices, residual.
3. **Rule 32 — DDD.** Domain code in `internal/domain/`; no `pgx`, `fiber`, or `zap` imports there.
4. **Rule 35 — Idempotency.** Every POST that mutates state requires `Idempotency-Key` header in production.
5. **Rule 36 — Failure as normal.** Every outbound call wrapped with breaker + retry + bounded timeout.
6. **Rule 39 — Security as core.** RLS at DB layer, RBAC at middleware, validation at every boundary, body limit on Fiber, constant-time compare on signature checks.
7. **Rule 40 — Observability.** Every log line carries `trace_id` and `span_id`. Custom metrics defined with explicit cardinality budget.
8. **Rule 46 — Rejection.** Anti-patterns from §3 of the plan are blocked by lint or policy. If you propose one, expect citation.
9. **Rule 53 — Supply chain.** GitHub Actions pinned by SHA. Container base images pinned by digest. Cosign signature required for deploy.
10. **Rule 54 — Policy as code.** Conftest and Kyverno bundles enforce architectural and security defaults at PR + admission time.

## 6. Onboarding (new engineer, first 30 days)

- Day 1: read `docs/TEAM-CHARTER.md`, `docs/RACI.md`, `docs/LAYERS.md`, `docs/SECOPS-CHARTER.md`. Run bootstrap. Get pre-commit installed.
- Day 2: walk through `docs/THREAT-MODEL.md`, `docs/RISK-REGISTER.md` with `secops`. Pair on one PR review focused on `/policies/`.
- Week 1: pick a starter issue tagged `good-first-issue`. Ship one ADR + one PR through the full flow.
- Week 2: pair on one runbook dry-run. Update one runbook with `last_tested` date.
- Week 3: present a small feature against the doctrine grid at platform office hours.
- Week 4: shadow on-call.

By day 30 you should be able to produce an ADR + a policy-clean PR + an updated runbook without external coaching beyond what lives in `docs/`. That is Rule 68 (DevSecOps termination objective).
