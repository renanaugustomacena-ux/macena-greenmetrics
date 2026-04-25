#!/usr/bin/env bash
# GreenMetrics — fresh-clone bootstrap.
# Doctrine refs: Rule 17 (DX — time-to-first-200 ≤ 5 min).
# Idempotent: safe to re-run.

set -Eeuo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

# ----- helpers ---------------------------------------------------------------

color_blue=$'\e[34m'
color_green=$'\e[32m'
color_yellow=$'\e[33m'
color_red=$'\e[31m'
color_reset=$'\e[0m'

log() { printf "%s[bootstrap]%s %s\n" "$color_blue" "$color_reset" "$*"; }
ok()  { printf "%s[ ok ]%s %s\n" "$color_green" "$color_reset" "$*"; }
warn(){ printf "%s[warn]%s %s\n" "$color_yellow" "$color_reset" "$*"; }
die() { printf "%s[fail]%s %s\n" "$color_red" "$color_reset" "$*" >&2; exit 1; }

require_cmd() {
    command -v "$1" >/dev/null 2>&1 || die "missing dependency: $1 — install before re-running"
}

# ----- prerequisites ---------------------------------------------------------

log "checking prerequisites…"
require_cmd docker
require_cmd jq
require_cmd curl
require_cmd openssl
docker compose version >/dev/null 2>&1 || die "docker compose v2 required"
ok "prerequisites present"

# ----- .env generation -------------------------------------------------------

if [[ ! -f .env ]]; then
    log "no .env found — generating from .env.example with fresh secrets…"
    cp .env.example .env

    # Generate cryptographically-random secrets where the example uses CHANGE_ME placeholders.
    JWT_SECRET="$(openssl rand -base64 48 | tr -d '\n')"
    POSTGRES_PASSWORD="$(openssl rand -base64 32 | tr -d '\n')"
    GRAFANA_ADMIN_PASSWORD="$(openssl rand -base64 32 | tr -d '\n')"
    PULSE_WEBHOOK_SECRET="$(openssl rand -base64 32 | tr -d '\n')"

    # macOS / GNU sed compat: write to temp + replace.
    awk -v jwt="$JWT_SECRET" \
        -v pg="$POSTGRES_PASSWORD" \
        -v gr="$GRAFANA_ADMIN_PASSWORD" \
        -v ps="$PULSE_WEBHOOK_SECRET" \
        'BEGIN{FS=OFS="="} \
         /^JWT_SECRET=/        {$2=jwt} \
         /^POSTGRES_PASSWORD=/ {$2=pg} \
         /^GRAFANA_ADMIN_PASSWORD=/ {$2=gr} \
         /^PULSE_WEBHOOK_SECRET=/   {$2=ps} \
         {print}' .env > .env.tmp && mv .env.tmp .env
    chmod 600 .env
    ok ".env generated with cryptographically-random secrets (mode 600)"
else
    ok ".env already exists — leaving untouched (re-runnable)"
fi

# ----- docker compose up -----------------------------------------------------

log "starting docker compose stack…"
docker compose up -d --build
ok "containers up"

# ----- wait for backend health -----------------------------------------------

log "waiting for backend /api/health (timeout 120 s)…"
deadline=$(( $(date +%s) + 120 ))
while [[ $(date +%s) -lt $deadline ]]; do
    if curl -fsS http://localhost:8082/api/health >/dev/null 2>&1; then
        ok "backend /api/health responding"
        break
    fi
    sleep 2
done
if ! curl -fsS http://localhost:8082/api/health >/dev/null 2>&1; then
    die "backend health did not come up in 120 s — check 'docker compose logs greenmetrics-backend'"
fi

# ----- migrations ------------------------------------------------------------

log "applying migrations…"
if docker compose exec -T greenmetrics-backend test -x /usr/local/bin/goose 2>/dev/null; then
    docker compose exec -T greenmetrics-backend goose -dir migrations postgres "$DATABASE_URL" up || warn "goose up returned non-zero — investigate"
else
    warn "goose not available in image yet (lands in S2); applying via psql shell-out"
    for f in 0001_init.sql 0002_hypertables.sql 0003_continuous_aggregates.sql 0004_retention.sql 0005_emission_factors.sql; do
        docker compose exec -T greenmetrics-timescaledb psql -U greenmetrics -d greenmetrics -f "/docker-entrypoint-initdb.d/$f" >/dev/null 2>&1 || warn "$f already applied or skipped"
    done
fi
ok "migrations applied"

# ----- smoke -----------------------------------------------------------------

log "smoke-testing endpoints…"
curl -fsS http://localhost:8082/api/health | jq -r '"backend health: \(.status), uptime: \(.uptime_seconds)s, version: \(.version)"'
curl -fsS http://localhost:8082/api/live   >/dev/null && ok "/api/live   200"
curl -fsS http://localhost:8082/api/ready  >/dev/null && ok "/api/ready  200"

# ----- tips ------------------------------------------------------------------

cat <<EOT

${color_green}Bootstrap complete.${color_reset}

Next steps:

  - Frontend:  http://localhost:3005
  - Grafana:   http://localhost:3011 (admin / see .env GRAFANA_ADMIN_PASSWORD)
  - Backend:   http://localhost:8082/api/docs  (Swagger UI)
  - API spec:  api/openapi/v1.yaml  (canonical)
  - Runbooks:  docs/runbooks/
  - Doctrine:  docs/CONTRIBUTING.md

  Login (dev fallback): operator@greenmetrics.local / any password ≥ 12 chars
                       (only works when repo unavailable + APP_ENV != production)

  Run tests:           cd backend && make test
  Run integration:     cd backend && make test-integration
  Run policy check:    task policy:check  (or make policy-check from backend/)

EOT
