# SECURITY — GreenMetrics

> **Framing note (2026-04-30 charter alignment):** GreenMetrics is delivered
> as a **modular template + engagement** model per `docs/MODULAR-TEMPLATE-CHARTER.md`
> §2. Each engagement runs a **single-tenant deployment by default**; multi-tenant
> is a partner-hosted opt-in (Charter §11). The tenant primitive (`tenant_id`,
> RLS policies, repository-level WHERE filters) below is **engineering hygiene
> preserved across deployment modes** — not multi-tenant SaaS framing. See
> `docs/THREAT-MODEL.md` §2 (Tenancy model) for the tenancy-mode threat surface.

## 1. Threat model (summary)

- **Ingestion attack surface:** Modbus-TCP simulator (dev-only) on :5020 is
  the only raw-socket listener. In production, the only external surface is
  the `/api/*` endpoints on :8082.
- **Sensitive data:** no PII in time-series; PII is limited to login email +
  password hash. Emission factors are public. Tenant ragione_sociale /
  partita_iva are regulated as business data (not sensitive under GDPR).
- **Primary trust boundary:** JWT verification at the `JWTMiddleware`.
  Secrets are loaded from env; defaults are refused in `APP_ENV=production`.

## 2. Hardening

- **Default-password refusal:** `config.Load` refuses to boot with
  `JWT_SECRET=change-me-*` or `GRAFANA_ADMIN_PASSWORD=change-me` in
  production. Tested by `tests/config_test.go`.
- **TLS to TimescaleDB:** `DATABASE_URL` with `sslmode=disable` is refused
  in production; default template is `sslmode=prefer`.
- **JWT strength:** secret minimum 32 bytes in production; algorithm is
  HS256 (hash-based MAC, rotatable from env).
- **CORS:** `CORS_ALLOWED_ORIGINS` is an explicit allow-list; default is
  localhost only.
- **HTTP middleware:** `recover`, `requestid`, `otelfiber`, `compress`,
  `cors` all applied in `cmd/server/main.go`.
- **Modbus-TCP client timeout:** bounded via `MODBUS_TCP_TIMEOUT_MS`
  (default 3000 ms), enforced in `modbus_ingestor.go`.
- **Modbus-TCP simulator timeout:** bounded via `connIdleDeadline = 30s`
  per connection; max 64 concurrent connections.
- **Docker:** backend builds distroless nonroot; `USER nonroot:nonroot`.
- **Healthcheck:** the binary's `--healthcheck` flag is used by
  `HEALTHCHECK` (no curl/wget in the runtime image).

## 3. Scans run in CI

- `govulncheck ./...` — Go advisories.
- `trivy fs . --severity CRITICAL,HIGH` — filesystem secrets and known CVEs.
- `trivy image greenmetrics-backend:audit` — built image CVE scan.
- `semgrep --config=p/owasp-top-ten --config=p/security-audit --config=p/golang`.
- `gitleaks detect --source=.` — secret detection.

Findings at HIGH or CRITICAL block CI (see `docs/ARCHITECTURE.md` §CI).
MEDIUM findings are tolerated with a written rationale.

## 4. Known accepted-risk items

- Login endpoint currently accepts any non-empty password (placeholder; real
  bcrypt/argon2 hash check is pending Mission III user-management work). In
  dev compose, the emitted JWT is purely for demoing protected routes.
- The ERP/FatturaFlow bridge reads the cost-centre map without write auth;
  documented as in-cluster service-to-service trust.

## 5. Incident response

See `docs/INCIDENT-RESPONSE.md` (to be added Mission II.5). For now,
on-call rotation is defined in `docs/RUNBOOK.md` §8.
