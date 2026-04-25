# Developer Experience (DX)

**Doctrine refs:** Rule 17 (DX as first-class).

DX is a quantified metric, not a vibe. Targets:

- Time-to-first-`/api/health 200` from a fresh clone ≤ **5 min** on a stock laptop.
- `make help` lists every target with a one-line description.
- Time-to-first-PR for a new engineer ≤ **3 days** (issue → branch → PR).
- Time-to-first-deploy for a new engineer ≤ **30 days** (per Rule 68 termination).

## Quick links

- **Bootstrap:** `./scripts/dev/bootstrap.sh` (idempotent, ≤ 5 min).
- **Devcontainer:** `.devcontainer/devcontainer.json` for VSCode + Codespaces with the full toolchain (Go 1.26, Node 20, terraform, kubectl, helm, argocd, cosign, syft, trivy, gosec, k6, k9s, conftest, kubeconform, hadolint, actionlint, tfsec).
- **CLI contract:** `docs/CLI-CONTRACT.md` — stable target names (`task`, `make`).
- **Debug guide:** `docs/DEBUG.md` — `dlv`, zap log filtering, Tempo trace correlation, Prometheus + Grafana, Timescale internals, RLS testing, Cosign verify.
- **Troubleshooting tree:** `docs/TROUBLESHOOTING.md` — decision tree per symptom.
- **Contributing guide:** `docs/CONTRIBUTING.md` — bootstrap, dev loop, doctrine, contribution flow, onboarding.
- **Doctrine plan:** `~/.claude/plans/my-brother-i-would-flickering-coral.md`.

## Bootstrap walkthrough

```bash
git clone <repo> && cd GreenMetrics
./scripts/dev/bootstrap.sh
# ⏱️  ~3-5 min on stock laptop with warm Docker cache
```

What the script does:

1. Verify prerequisites (`docker`, `jq`, `curl`, `openssl`).
2. Generate `.env` with cryptographic secrets if missing.
3. `docker compose up -d --build` (5 services: TimescaleDB, backend, frontend, Grafana, simulator).
4. Wait for `/api/health 200` (timeout 120 s).
5. Apply migrations (goose if available, fallback to `psql`).
6. Smoke `/api/health`, `/api/live`, `/api/ready`.
7. Print next-step tips.

## Daily dev loop

```bash
# Backend hot-reload (with delve debug listening on :2345)
make debug-backend

# Frontend hot-reload
cd frontend && npm run dev

# Run a single test
go test -run TestLogin ./internal/handlers/...

# Run all unit tests with race
make test

# Run integration tests against ephemeral testcontainer
make test-integration

# Run E2E suite (Playwright)
task test:e2e

# Lint everything pre-commit lints
make precommit

# Full continuous-verification loop
make verify

# Render docs/LAYERS.md from docs/layers.yaml
make layers-doc

# Print every target available
make help
```

## Pre-commit hooks (mandatory)

```bash
pip install pre-commit
pre-commit install
pre-commit install --hook-type commit-msg
```

Hooks run on every commit. Total time target < 5 s on cache hit. Bypassing with `--no-verify` is **not** acceptable.

## Common pitfalls

- **Docker Desktop on macOS:** allocate ≥ 4 GB RAM and 4 CPUs in Settings, otherwise TimescaleDB OOMs during CAGG creation.
- **Linux UID mismatch:** if `chown` errors during `npm install`, run `sudo chown -R $(id -u):$(id -g) frontend/node_modules`.
- **Port collisions:** ports `3005` (frontend), `3011` (Grafana), `5439` (Timescale), `8082` (backend) — change in `docker-compose.yml` or stop the conflicting service.
- **Slow cold start:** first `docker compose up` builds images and downloads the Timescale image (~600 MB); subsequent runs reuse cache.

## Onboarding (new engineer, first 30 days)

See `docs/CONTRIBUTING.md` §6.
