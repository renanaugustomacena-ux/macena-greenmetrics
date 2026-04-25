# GreenMetrics Threat Model

**Owner:** `@greenmetrics/secops`.
**Doctrine refs:** Rule 19 (security as structural), Rule 39 (security as core), Rule 51 (sequence step 1), Rule 55 (continuous threat modelling).
**Authoring date:** 2026-04-25.
**Review log:** see §7 below; cadence quarterly.

This is a living artefact (Rule 55). Update on system change, dependency change, or new attack-surface introduction. Pin every Dependabot major-version PR to a checkbox: "Threat model unchanged | Threat model updated".

## 1. Scope

GreenMetrics is a multi-tenant SaaS for energy and sustainability reporting in Italy. It ingests meter data via multiple protocols, computes Scope 1/2/3 emissions, and generates regulatory disclosures (CSRD/ESRS E1, Piano Transizione 5.0, Conto Termico, TEE, D.Lgs. 102/2014 audit). Operates in eu-south-1 (Milan); Italian residency invariant.

## 2. Trust boundaries (cross-ref `docs/TRUST-BOUNDARIES.md`)

| ID | Boundary | Direction | Defence summary |
|---|---|---|---|
| TB-01 | Browser → Ingress | inbound | TLS 1.3, WAF (CloudFront + AWS WAF rate-based + managed core rules), rate limit |
| TB-02 | Ingress → Backend | inbound | mTLS (planned), JWT HS256 with `kid`, RBAC permission registry, body-size limit |
| TB-03 | Backend → DB (TimescaleDB) | outbound | pgx prepared statements, `app_user` PG role with `BYPASSRLS=false`, RLS policies per tenant, TLS sslmode=require |
| TB-04 | Backend → Grafana admin API | outbound | basic auth via ESO-managed admin credentials, internal-only ClusterIP |
| TB-05 | Backend → ISPRA factor refresh | outbound | TLS pinning, breaker, fallback to cached factors |
| TB-06 | Backend → Terna grid mix | outbound | TLS pinning, breaker |
| TB-07 | Backend → E-Distribuzione POD lookup | outbound | TLS pinning, breaker |
| TB-08 | Modbus ingestor → Modbus host | outbound | NetworkPolicy egress allowlist, breaker, future TLS |
| TB-09 | M-Bus ingestor → serial device | outbound | physical device path, no network exposure |
| TB-10 | OCPP central system → Backend WebSocket | inbound | mTLS (planned), per-charge-point auth token, sticky session |
| TB-11 | Pulse webhook → Backend | inbound | HMAC-SHA256 signature with constant-time compare, body-size limit, idempotency key |
| TB-12 | CI runner → GHCR | outbound | GitHub OIDC, Cosign keyless sign |
| TB-13 | GHCR → Cluster (image pull) | outbound | Kyverno admission `verify-images` against Sigstore root |
| TB-14 | Cluster → AWS Secrets Manager | outbound | IRSA per-pod, scoped to `greenmetrics/prod/*` |
| TB-15 | Operator → AWS Console / kubectl | inbound | MFA, OIDC SSO, break-glass IAM with `max_session_duration=1h` and CloudTrail alarm |
| TB-16 | Customer → Status page | inbound | static site, no auth, no PII |
| TB-17 | Audit log → S3 (long retention) | outbound | KMS encryption, Object Lock `compliance` mode, 5y retention |
| TB-18 | EKS control plane | inbound | `endpoint_public_access=false` for prod; operator VPN CIDR only |
| TB-19 | Argo CD UI | inbound | OIDC SSO, ClusterIP only, Ingress with mTLS |

## 3. STRIDE per surface

### 3.1 Modbus-TCP ingestor (`backend/internal/services/modbus_ingestor.go`)

- **Spoofing:** untrusted gateway could inject readings → mitigations: NetworkPolicy egress allowlist (Layer 2), TLS client cert (planned), per-meter trust signing (future).
- **Tampering:** in-flight register tampering → mitigations: TLS where supported, anomaly detection (z-score on consumption deltas) flags abnormal jumps.
- **Repudiation:** ingestor logs include source IP + register address; audit log row per readings batch.
- **Information disclosure:** Modbus is plaintext by default; mitigations: TLS where supported, segregated VPN to meter network.
- **Denial of service:** slow Modbus host could block ingest loop → mitigations: per-host breaker, bounded timeout (3s), worker pool (Rule 41), fall back to cached value.
- **Elevation of privilege:** N/A (no auth on Modbus protocol).

### 3.2 Pulse webhook (`backend/internal/handlers/pulse.go`, `PULSE_WEBHOOK_SECRET`)

- **Spoofing:** unsigned webhook payload → mitigations: HMAC-SHA256 over body, constant-time compare (CURRENT GAP — plaintext compare is timing oracle; closing in S3).
- **Tampering:** body modified in transit → mitigations: HMAC verifies body integrity.
- **Repudiation:** every webhook logged with `request_id`, `tenant_id`, `meter_id`, sig hash.
- **Information disclosure:** webhook secret in env; ESO-managed (CURRENT GAP — placeholder; closing in S3).
- **Denial of service:** flood at webhook endpoint → mitigations: rate limit `RATE_LIMIT_INGEST_PER_MINUTE=300`, per-tenant bucket, body-size limit 16 MB, idempotency key dedups replays.
- **Elevation of privilege:** webhook scope limited to `tenants/{tenant_id}/meters/{meter_id}/readings`; no admin escalation possible.

### 3.3 JWT path (`backend/internal/handlers/auth.go`, JWTMiddleware)

- **Spoofing:** forged JWT → mitigations: HS256 pinned at line 179, 236; secret length ≥ 32 bytes; `kid` rotation (planned S3); `alg:none` rejected by `jwt.WithValidMethods([]string{"HS256"})`.
- **Tampering:** signature check via golang-jwt v5; pinned alg means downgrade-to-none impossible.
- **Repudiation:** every authenticated request logged with `user_email`, `tenant_id`, `request_id`, `trace_id`; audit log row for state-changing requests.
- **Information disclosure:** secret never logged (zap field redactor); no token echoed in error responses; refresh token in body only.
- **Denial of service:** bcrypt cost factor 12 deliberately slow; mitigations: per-IP+email lockout (`auth_lockout.go`, threshold 5 failures, 30 min window), rate limit `RATE_LIMIT_LOGIN_PER_MINUTE=5`.
- **Elevation of privilege:** RBAC permission registry middleware (planned S3) gates per-route. Today `role` claim set but not checked — CURRENT GAP, closing in S3.

### 3.4 Grafana (`docker-compose.yml:111`)

- **Spoofing:** Grafana login bypass → mitigations: `GF_AUTH_ANONYMOUS_ENABLED=false` (line 124), admin credentials via ESO, OIDC SSO planned.
- **Tampering:** dashboard tampering → mitigations: provisioning is read-only (`disableDeletion: true`, `allowUiUpdates: false`).
- **Repudiation:** Grafana audit log enabled; shipped to Loki.
- **Information disclosure:** dashboards may leak per-tenant data → mitigations: dashboard variables scoped per-tenant; Grafana data source query whitelist.
- **Denial of service:** Grafana cardinality explosion → mitigations: Prometheus label budget review (Rule 18).
- **Elevation of privilege:** Grafana admin → cluster admin? No — Grafana SA scoped to its own namespace.

### 3.5 OCPP central system (`backend/internal/services/ocpp_client.go`, `OCPP_CENTRAL_SYSTEM_URL`)

- **Spoofing:** untrusted CS → mitigations: per-CS auth token, mTLS (planned).
- **Tampering:** OCPP message tampering → mitigations: TLS in transit, message schema validation.
- **Repudiation:** every OCPP message logged.
- **Information disclosure:** charge point ID in URL; mitigations: scope-limited per-CP credentials.
- **Denial of service:** WebSocket flood → mitigations: connection cap, per-CP rate limit, breaker.
- **Elevation of privilege:** CS cannot escalate beyond its CP scope.

### 3.6 External egress (ISPRA, Terna, E-Distribuzione)

- **Spoofing:** DNS hijack of factor source → mitigations: TLS pinning to known root, certificate transparency monitoring, fallback to cached factors.
- **Tampering:** in-flight factor tampering → mitigations: TLS, fallback to cached.
- **Repudiation:** every external call logged with response hash.
- **Information disclosure:** outbound calls carry only public lookup keys; no PII leaks outbound.
- **Denial of service:** ISPRA down → mitigations: cached factors, breaker, alert at >1% fallback rate.
- **Elevation of privilege:** N/A (read-only consumers).

### 3.7 GitHub Actions / Supply chain

- **Spoofing:** rogue Action with same tag (the `tj-actions/changed-files` 2025 incident pattern) → mitigations: SHA pinning per Rule 53, `gh api ... actions/permissions` set to `selected`.
- **Tampering:** Action artifact tampering → mitigations: Cosign signature verify, SLSA provenance attest.
- **Repudiation:** every workflow run has GitHub-issued attestation.
- **Information disclosure:** secrets scope minimised; OIDC eliminates static keys.
- **Denial of service:** N/A (ephemeral runners).
- **Elevation of privilege:** least-privilege OIDC trust policy bound to `repo:` and `ref:`.

### 3.8 EKS control plane

- **Spoofing:** stolen kubeconfig → mitigations: `endpoint_public_access=false`, IRSA per-pod, MFA on operator AWS user.
- **Tampering:** API server compromise → mitigations: control plane logs to CloudWatch (api/audit/authenticator/controllerManager/scheduler), Falco DaemonSet on data plane.
- **Repudiation:** EKS audit log shipped to Loki, retained 365d.
- **Information disclosure:** etcd encrypted at rest with KMS.
- **Denial of service:** API server overload → mitigations: AWS-managed scaling, Argo CD batched syncs.
- **Elevation of privilege:** PSS `restricted` enforced; Kyverno verify-images admission policy denies unsigned pods.

## 4. Risk-to-control mapping

Every threat above maps to one or more risks in `docs/RISK-REGISTER.md` and one or more controls in `policies/conftest/`, `policies/kyverno/`, or `backend/internal/security/`. The mapping is the bidirectional traceability the regulator will ask for.

## 5. Out-of-scope (today)

- Quantum-resistant signing (Cosign roadmap; revisit when Sigstore adopts).
- Multi-region active-active (active-passive only per Rule 25).
- Customer-supplied SSO (Mission III + roadmap).
- GDPR DSAR for customer-managed identity provider integration (out until SSO ships).

## 6. Anti-patterns rejected (Rule 66)

- "We can't TLS Modbus, customers won't allow it" — accepted today, but TLS-where-supported is the default; non-TLS paths are documented exceptions.
- "Just disable the breaker so the simulator works" — never. Simulator timeouts are the same as production timeouts.
- "Pin the Action to `@latest` for convenience" — never. SHA pinning is the rule.

## 7. Review log

| Date | Reviewer | Trigger | Outcome |
|---|---|---|---|
| 2026-04-25 | secops | Initial authoring | All surfaces enumerated; gaps tracked in S2/S3 sprints. |
| _next: 2026-07-25_ | secops | Quarterly | TBD. |
