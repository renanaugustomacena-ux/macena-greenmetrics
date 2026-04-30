# GreenMetrics Threat Model

**Owner:** `@greenmetrics/secops`.
**Doctrine refs:** Rule 8 (Italian residency default), Rule 19 (security as structural), Rule 39 (security as core), Rule 51 (sequence step 1), Rule 55 (continuous threat modelling).
**Charter refs:** §10 (deployment topologies), §11 (tenancy model under the new framing).
**Authoring date:** 2026-04-25.
**Last rewrite:** 2026-04-30 (§1 reframed; §2 Tenancy model added per Charter §11; sections §3–§8 renumbered from §2–§7).
**Review log:** see §8 below; cadence quarterly.

This is a living artefact (Rule 55). Update on system change, dependency change, or new attack-surface introduction. Pin every Dependabot major-version PR to a checkbox: "Threat model unchanged | Threat model updated".

## 1. Scope

GreenMetrics is the modular template for industrial energy management and sustainability reporting, delivered as engagements per `docs/MODULAR-TEMPLATE-CHARTER.md`. Each engagement runs a deployment that ingests meter data via multiple OT protocols (Modbus RTU/TCP, M-Bus, SunSpec, OCPP, Pulse, IEC 61850 / OPC UA / MQTT Sparkplug B in catalogue), computes Scope 1 / 2 / 3 emissions against versioned authoritative factor sources, and generates regulatory disclosures (CSRD ESRS E1, Piano Transizione 5.0, Conto Termico 2.0, Certificati Bianchi TEE, D.Lgs. 102/2014 audit energetico).

The default deployment topology is Topology A (AWS eu-south-1 Milan, Charter §10.1); Topologies B (Italian-sovereign cloud), C (on-prem), and D (hybrid OT/IT) carry the same logical threat surface with topology-specific control implementations. Italian residency is the default per Rule 8; cross-EU transfer is opt-in only.

The template's threat surface is enumerated here. An engagement fork that adds engagement-specific Packs, integrations, or Core overlays per Charter §5.2 inherits this threat model and extends it with per-engagement entries in `engagements/<client>/THREAT-MODEL-OVERLAY.md` (Phase F deliverable). The annual pen-test (Rule 60) covers the template; engagement-specific code is pen-tested either by the engagement client's own security team or by an extended scope of the annual pen-test.

## 2. Tenancy model

**Single-tenant by default; multi-tenant by configuration.** This is the Charter §11 framing.

A default deployment runs one engagement's data on one isolated substrate. The Region Pack's `bootstrap_tenant_id` is materialised at first migration; the frontend's "switch tenant" surface is hidden behind a feature flag. There is no shared production database, no shared backend pool, no shared CloudFront distribution across engagements. Engagements are commercially, operationally, and infrastructurally separate.

A multi-tenant deployment is an opt-in configuration used only when a partner ESCO or system integrator hosts multiple SME clients on one substrate, by their own commercial arrangement. In that case the partner — not Macena — owns the threat model for cross-tenant interactions, the data-processing agreement perimeter under Art. 28 GDPR, and the audit-evidence pack for each hosted SME. Macena's commitment is that the template's defence-in-depth controls work correctly under both modes.

**Defence-in-depth tenant isolation is preserved unchanged across modes.** The following controls are not multi-tenancy features; they are engineering hygiene that eliminates an entire class of bug from the attack surface and makes the multi-tenant opt-in a configuration switch, not a re-architecture:

- the `tenant_id` UUIDv4 invariant on every domain row (Rule 3);
- repository-level `WHERE tenant_id = $1` on every read and write (Rule 39);
- Postgres RLS policies per tenant on the `tenants`, `users`, `meters`, `meter_channels`, `readings`, `readings_15min` / `readings_1h` / `readings_1d`, `emission_factors`, `reports`, `alerts`, `audit_log`, `idempotency_keys` tables (ADR-0002, ADR-0011);
- JWT-claim-driven middleware with `InTxAsTenant` opening every transaction with `SET LOCAL app.tenant_id = …` (Rule 39);
- audit-log row keyed on tenant + actor + action (Rule 62);
- per-tenant DEK + KMS-wrapped MEK on `readings.raw_payload` (Rule 172, Phase F Sprint S10 deliverable);
- per-meter HMAC reading provenance (Rule 173, Phase F Sprint S11 deliverable);
- per-Pack health and per-Pack failure isolation (Rule 76).

The STRIDE per-surface analysis in §4 applies identically to both modes. Where a control implementation differs by deployment mode (e.g., Grafana data-source query scope when more than one tenant is loaded), the difference is called out in the relevant subsection. The Trust boundaries in §3 are unchanged across modes — the surface is structural, not deployment-mode-specific.

**What this framing rejects.** A historical reading of GreenMetrics as a "multi-tenant SaaS" — implying a single shared production deployment serving N customers via tenant-flag isolation — is no longer accurate. The SaaS framing is retired per Charter §2 and ADR-0021. Per-meter pricing and self-serve `/signup` endpoints are explicitly forbidden by Charter §13. The threat-model implication: the cross-tenant compromise vectors that dominated v1 are now defence-in-depth checks rather than primary controls; the new primary surface is per-deployment integrity (Pack-loader manifest-lock signing, Rule 73), supply-chain attestation (Rule 57), and engagement-fork sync hygiene (Rule 79).

**Cross-refs.** Charter §10 (deployment topologies). Charter §11 (tenancy framing). Rule 39 (security as structural). Rule 76 (Pack failure isolation). ADR-0002 (multi-tenant RLS strategy). ADR-0011 (Postgres RLS defence in depth). ADR-0021 (charter and doctrine adoption).

## 3. Trust boundaries (cross-ref `docs/TRUST-BOUNDARIES.md`)

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

## 4. STRIDE per surface

### 4.1 Modbus-TCP ingestor (`backend/internal/services/modbus_ingestor.go`)

- **Spoofing:** untrusted gateway could inject readings → mitigations: NetworkPolicy egress allowlist (Layer 2), TLS client cert (planned), per-meter trust signing (future).
- **Tampering:** in-flight register tampering → mitigations: TLS where supported, anomaly detection (z-score on consumption deltas) flags abnormal jumps.
- **Repudiation:** ingestor logs include source IP + register address; audit log row per readings batch.
- **Information disclosure:** Modbus is plaintext by default; mitigations: TLS where supported, segregated VPN to meter network.
- **Denial of service:** slow Modbus host could block ingest loop → mitigations: per-host breaker, bounded timeout (3s), worker pool (Rule 41), fall back to cached value.
- **Elevation of privilege:** N/A (no auth on Modbus protocol).

### 4.2 Pulse webhook (`backend/internal/handlers/pulse.go`, `PULSE_WEBHOOK_SECRET`)

- **Spoofing:** unsigned webhook payload → mitigations: HMAC-SHA256 over body, constant-time compare (CURRENT GAP — plaintext compare is timing oracle; closing in S3).
- **Tampering:** body modified in transit → mitigations: HMAC verifies body integrity.
- **Repudiation:** every webhook logged with `request_id`, `tenant_id`, `meter_id`, sig hash.
- **Information disclosure:** webhook secret in env; ESO-managed (CURRENT GAP — placeholder; closing in S3).
- **Denial of service:** flood at webhook endpoint → mitigations: rate limit `RATE_LIMIT_INGEST_PER_MINUTE=300`, per-tenant bucket, body-size limit 16 MB, idempotency key dedups replays.
- **Elevation of privilege:** webhook scope limited to `tenants/{tenant_id}/meters/{meter_id}/readings`; no admin escalation possible.

### 4.3 JWT path (`backend/internal/handlers/auth.go`, JWTMiddleware)

- **Spoofing:** forged JWT → mitigations: HS256 pinned at line 179, 236; secret length ≥ 32 bytes; `kid` rotation (planned S3); `alg:none` rejected by `jwt.WithValidMethods([]string{"HS256"})`.
- **Tampering:** signature check via golang-jwt v5; pinned alg means downgrade-to-none impossible.
- **Repudiation:** every authenticated request logged with `user_email`, `tenant_id`, `request_id`, `trace_id`; audit log row for state-changing requests.
- **Information disclosure:** secret never logged (zap field redactor); no token echoed in error responses; refresh token in body only.
- **Denial of service:** bcrypt cost factor 12 deliberately slow; mitigations: per-IP+email lockout (`auth_lockout.go`, threshold 5 failures, 30 min window), rate limit `RATE_LIMIT_LOGIN_PER_MINUTE=5`.
- **Elevation of privilege:** RBAC permission registry middleware (planned S3) gates per-route. Today `role` claim set but not checked — CURRENT GAP, closing in S3.

### 4.4 Grafana (`docker-compose.yml:111`)

- **Spoofing:** Grafana login bypass → mitigations: `GF_AUTH_ANONYMOUS_ENABLED=false` (line 124), admin credentials via ESO, OIDC SSO planned.
- **Tampering:** dashboard tampering → mitigations: provisioning is read-only (`disableDeletion: true`, `allowUiUpdates: false`).
- **Repudiation:** Grafana audit log enabled; shipped to Loki.
- **Information disclosure:** dashboards may leak per-tenant data → mitigations: dashboard variables scoped per-tenant; Grafana data source query whitelist. In a partner-hosted multi-tenant deployment per §2, the query-scope filter is the load-bearing isolation control.
- **Denial of service:** Grafana cardinality explosion → mitigations: Prometheus label budget review (Rule 18).
- **Elevation of privilege:** Grafana admin → cluster admin? No — Grafana SA scoped to its own namespace.

### 4.5 OCPP central system (`backend/internal/services/ocpp_client.go`, `OCPP_CENTRAL_SYSTEM_URL`)

- **Spoofing:** untrusted CS → mitigations: per-CS auth token, mTLS (planned).
- **Tampering:** OCPP message tampering → mitigations: TLS in transit, message schema validation.
- **Repudiation:** every OCPP message logged.
- **Information disclosure:** charge point ID in URL; mitigations: scope-limited per-CP credentials.
- **Denial of service:** WebSocket flood → mitigations: connection cap, per-CP rate limit, breaker.
- **Elevation of privilege:** CS cannot escalate beyond its CP scope.

### 4.6 External egress (ISPRA, Terna, E-Distribuzione)

- **Spoofing:** DNS hijack of factor source → mitigations: TLS pinning to known root, certificate transparency monitoring, fallback to cached factors.
- **Tampering:** in-flight factor tampering → mitigations: TLS, fallback to cached.
- **Repudiation:** every external call logged with response hash.
- **Information disclosure:** outbound calls carry only public lookup keys; no PII leaks outbound.
- **Denial of service:** ISPRA down → mitigations: cached factors, breaker, alert at >1% fallback rate.
- **Elevation of privilege:** N/A (read-only consumers).

### 4.7 GitHub Actions / Supply chain

- **Spoofing:** rogue Action with same tag (the `tj-actions/changed-files` 2025 incident pattern) → mitigations: SHA pinning per Rule 53, `gh api ... actions/permissions` set to `selected`.
- **Tampering:** Action artifact tampering → mitigations: Cosign signature verify, SLSA provenance attest.
- **Repudiation:** every workflow run has GitHub-issued attestation.
- **Information disclosure:** secrets scope minimised; OIDC eliminates static keys.
- **Denial of service:** N/A (ephemeral runners).
- **Elevation of privilege:** least-privilege OIDC trust policy bound to `repo:` and `ref:`.

### 4.8 EKS control plane

- **Spoofing:** stolen kubeconfig → mitigations: `endpoint_public_access=false`, IRSA per-pod, MFA on operator AWS user.
- **Tampering:** API server compromise → mitigations: control plane logs to CloudWatch (api/audit/authenticator/controllerManager/scheduler), Falco DaemonSet on data plane.
- **Repudiation:** EKS audit log shipped to Loki, retained 365d.
- **Information disclosure:** etcd encrypted at rest with KMS.
- **Denial of service:** API server overload → mitigations: AWS-managed scaling, Argo CD batched syncs.
- **Elevation of privilege:** PSS `restricted` enforced; Kyverno verify-images admission policy denies unsigned pods.

### 4.9 Pack-loader and manifest-lock (Phase E Sprint S5)

- **Spoofing:** rogue Pack registered at boot → mitigations: Pack manifest validated against `docs/contracts/pack-manifest.schema.json`; Pack contract-version pinned (Rule 71); `config/required-packs.yaml` declares the expected Pack set; manifest lock signed at boot (Rule 73, Phase F Sprint S11).
- **Tampering:** Pack code or manifest tampering → mitigations: Pack image signed with Cosign keyless (Rule 57); image admission verifies signature (Kyverno); SBOM cross-check (Rule 58).
- **Repudiation:** every Pack load logged with `pack_id`, `pack_kind`, `version`, `min_core_version`, manifest hash.
- **Information disclosure:** Pack code is part of the engagement deliverable; no Pack should hold secrets in code (Rule 20). Pack secrets via ESO/Vault per topology.
- **Denial of service:** Pack init failure → mitigations: Core boots only with the required Pack set per `config/required-packs.yaml` and refuses to boot otherwise (Rule 73); per-Pack failure isolation (Rule 76); Pack health surfaced into `/api/health` envelope (Rule 74).
- **Elevation of privilege:** Pack registration via `Registrar` indirection (Rule 72) — no global access; Core surface exposed to Packs is the typed `CoreHandle` (logger, metrics, tracer, repo, config), not a service-locator.

## 5. Risk-to-control mapping

Every threat above maps to one or more risks in `docs/RISK-REGISTER.md` and one or more controls in `policies/conftest/`, `policies/kyverno/`, or `backend/internal/security/`. The mapping is the bidirectional traceability the regulator will ask for.

## 6. Out-of-scope (today)

- Quantum-resistant signing (Cosign roadmap; revisit when Sigstore adopts).
- Multi-region active-active (active-passive only per Rule 25).
- Customer-supplied SSO at Core boot — covered by the Identity Pack catalogue (SAML / OIDC / SPID / CIE) shipping in Phase E Sprint S8 (ADR-0043 / ADR-0044).
- GDPR DSAR for customer-managed identity provider integration — moves into scope when the Identity Packs land in Phase E Sprint S8; full DSAR endpoint stub ships in Sprint S5 PR #7 with the implementation completing in Phase F Sprint S10.
- Cross-tenant attack surface in a partner-hosted multi-tenant deployment — owned by the partner per §2.
- Per-engagement client-specific code under `engagements/<client>/` — covered by the engagement-fork's own `engagements/<client>/THREAT-MODEL-OVERLAY.md`.

## 7. Anti-patterns rejected (Rule 66)

- "We can't TLS Modbus, customers won't allow it" — accepted today, but TLS-where-supported is the default; non-TLS paths are documented exceptions.
- "Just disable the breaker so the simulator works" — never. Simulator timeouts are the same as production timeouts.
- "Pin the Action to `@latest` for convenience" — never. SHA pinning is the rule.
- "Single-tenant deployments don't need RLS — let's drop it for the engagement fork" — never. Defence in depth survives the per-engagement deployment per §2; dropping RLS removes a class of bug from the floor and makes the multi-tenant opt-in a re-architecture instead of a configuration change.
- "Self-serve `/signup` for SMEs that aren't worth an engagement" — rejected by Charter §13. Industrial customers don't self-serve; an unscoped tenant is an unauthorised attack surface.

## 8. Review log

| Date | Reviewer | Trigger | Outcome |
|---|---|---|---|
| 2026-04-25 | secops | Initial authoring | All surfaces enumerated; gaps tracked in S2/S3 sprints. |
| 2026-04-30 | secops | Charter / doctrine adoption (ADR-0021); §1 reframed; §2 Tenancy model added per Charter §11; §4.9 Pack-loader added; subsequent sections renumbered. | Multi-tenant SaaS framing retired. Defence-in-depth controls preserved unchanged. Pack-loader threat surface enumerated. |
| _next: 2026-07-30_ | secops | Quarterly | TBD. |
