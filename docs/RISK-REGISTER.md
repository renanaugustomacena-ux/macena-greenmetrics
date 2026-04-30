# GreenMetrics Risk Register

**Owner:** `@greenmetrics/secops`.
**Doctrine refs:** Rule 55 (continuous threat modelling), Rule 61 (compliance ≠ goal), Rule 67 (transparency).
**Authoring date:** 2026-04-25.
**Last reviewed:** 2026-04-30 (charter-alignment annotation added; risk scoring unchanged).
**Review cadence:** monthly for HIGH/CRITICAL changes; quarterly full re-scoring.

> **Framing note (2026-04-30 charter alignment):** The deployment model is
> single-tenant by default per `docs/MODULAR-TEMPLATE-CHARTER.md` §11.
> "Cross-tenant" risks below (e.g. RISK-007) remain in scope **as defence in
> depth** — the engineering controls (RLS, RBAC, repository-level WHERE)
> survive across deployment modes per Charter §2 / Plan §5.3, and are
> additionally load-bearing in any partner-hosted multi-tenant configuration.
> Risk likelihoods reflect both modes.

Every Rego rule in `policies/conftest/` and `policies/kyverno/` references its `RISK-NNN` mitigation target. Every regulatory citation in `docs/ITALIAN-COMPLIANCE.md` is annotated with `MITIGATES: RISK-NNN` where applicable.

L = likelihood (1–5). I = impact (1–5). Score = L × I. Residual = L_after × I_after.

## 1. Active risks

| ID | Description | L | I | Score | Owner | Mitigations | Residual L | Residual I | Residual Score | Review date |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-001 | Modbus simulator `math/rand` seeding observable; deterministic seed predictable | 1 | 2 | 2 | secops | Simulator only in dev/test; not deployed in prod (`backend/cmd/simulator/main.go //nolint:gosec` annotated); CI gate refuses simulator image in production manifest | 1 | 2 | 2 | 2026-07-25 |
| RISK-002 | Alpine rolling apk in `backend/Dockerfile:8` (hadolint ignore for `--no-install-recommends`) | 2 | 2 | 4 | secops | Monthly Renovate base image refresh; Trivy image scan blocks HIGH/CRITICAL on any apk diff; SBOM tracks per-image package versions | 1 | 2 | 2 | 2026-07-25 |
| RISK-003 | JWT secret static; no rotation procedure | 2 | 4 | 8 | secops | Rotation procedure in `docs/JWT-ROTATION.md`; quarterly cron via `.github/workflows/jwt-rotation.yml`; `kid`-aware validation; dual-key window 24h | 1 | 2 | 2 | 2026-07-25 (post S3) |
| RISK-004 | Egress to ISPRA / Terna / E-Distribuzione can be tampered or DNS-hijacked | 1 | 3 | 3 | secops | TLS pinning to known roots; certificate transparency monitoring; per-host breaker; fallback to cached factors with `data_freshness` stamp; alert at fallback rate >1%/min | 1 | 2 | 2 | 2026-07-25 |
| RISK-005 | GitHub Actions tag mutation (`tj-actions/changed-files` 2025 incident pattern) | 2 | 4 | 8 | secops | SHA-pinning per Rule 53; `actions.permissions: selected` + sha-pinned allowlist; weekly Dependabot updates; Cosign verify on artefacts | 1 | 2 | 2 | 2026-07-25 (post S1) |
| RISK-006 | Insider with prod IAM access could exfiltrate data or pivot | 1 | 5 | 5 | secops | IRSA per-pod (no human IAM principal at runtime); break-glass IAM with `max_session_duration=1h` + CloudTrail alarm; 4-eyes rule for prod policy changes via CODEOWNERS + branch protection; MFA mandatory | 1 | 4 | 4 | 2026-10-25 |
| RISK-007 | Cross-tenant data leak via SQL bug or SECURITY DEFINER function | 2 | 5 | 10 | secops | RLS at DB layer (`migrations/00006_rls_enable.go`) + RBAC at middleware + app-level WHERE filter (defence in depth); RLS isolation property test (`tests/security/rls_isolation_test.go`); CI lint flags any new query without tenant context | 1 | 4 | 4 | 2026-07-25 (post S3) |
| RISK-008 | OCPP central system DoS amplification (charge point flood) | 2 | 3 | 6 | secops | Per-CP connection cap; per-CP rate limit; breaker; bounded WebSocket lifecycle (60s idle close + ping/pong) | 1 | 2 | 2 | 2026-07-25 (post S4) |
| RISK-009 | Audit log tampering by privileged user | 1 | 4 | 4 | secops | RLS append-only policy on `audit_log`; ship to Loki immutably; S3 audit bucket Object Lock `compliance` mode 5y; KMS encryption | 1 | 2 | 2 | 2026-07-25 (post S5) |
| RISK-010 | Backup restore key compromise | 1 | 5 | 5 | secops | KMS key rotation; separate IAM role for restore (not assumed by app); 4-eyes restore approval; CloudTrail alarm on `kms:Decrypt` against backup key | 1 | 4 | 4 | 2026-10-25 |
| RISK-011 | Pulse webhook signature timing oracle | 3 | 3 | 9 | secops | Constant-time compare via `subtle.ConstantTimeCompare` (closing S3); HMAC-SHA256 over body; replay protection via timestamp window; idempotency key dedups replays | 1 | 2 | 2 | 2026-07-25 (post S3) |
| RISK-012 | OTel sample ratio 1.0 in production overwhelms collector | 3 | 2 | 6 | platform | Default flipped to 0.1 in production (`config.Load`); explicit env override available; SLO monitoring on collector ingestion rate | 1 | 2 | 2 | 2026-07-25 (post S4) |
| RISK-013 | Sentinel JWT secret reaches production environment | 2 | 5 | 10 | secops | Hard boot refusal in `config.Load` when `cfg.AppEnv == "production"` and secret matches sentinel patterns; integration test `tests/security/boot_refuses_dev_secret_test.go` | 1 | 5 | 5 | 2026-07-25 (post S3) |
| RISK-014 | Unbounded Fiber body size enables payload-flood DoS | 3 | 3 | 9 | platform | `BodyLimit: 4 MB` global; 16 MB per-route override on ingest; reject 413 RFC 7807 | 1 | 2 | 2 | 2026-07-25 (post S3) |
| RISK-015 | Goroutine leak in ingestor crashes pod under sustained load | 3 | 4 | 12 | platform | Refactor to `errgroup` + `panjf2000/ants` pool; `goleak.VerifyNone` in tests; nightly soak test verifies stable RSS + goroutine count | 1 | 3 | 3 | 2026-07-25 (post S4) |
| RISK-016 | Synchronous report generation creates head-of-line latency | 4 | 3 | 12 | platform | Async via Asynq + Redis worker; `POST /v1/reports` returns 202 + Location | 1 | 2 | 2 | 2026-07-25 (post S4) |
| RISK-017 | Unsigned image deployed to production | 2 | 5 | 10 | secops | Cosign keyless sign in `cd.yml` supply-chain workflow; Kyverno `verify-images` admission policy denies unsigned; SLSA L2 provenance attached | 1 | 5 | 5 | 2026-07-25 (post S3) |
| RISK-018 | Terraform state on local FS, no lock, no encryption | 4 | 4 | 16 | platform | S3 backend + DynamoDB lock + KMS encryption (closing S2); MFA delete; versioning enabled | 1 | 3 | 3 | 2026-07-25 (post S2) |
| RISK-019 | Manual `kubectl apply` drift in production | 3 | 3 | 9 | platform | Argo CD GitOps with `selfHeal: true`; reverts manual mutations within 60s; Kyverno admission denies unsigned regardless | 1 | 2 | 2 | 2026-07-25 (post S4) |
| RISK-020 | Missing IR doc — first SEV1 will be ad-hoc | 4 | 4 | 16 | secops | `docs/INCIDENT-RESPONSE.md` (closing S5); on-call rota in PagerDuty; severity matrix; postmortem template; NIS2 24h/72h notification template | 1 | 3 | 3 | 2026-07-25 (post S5) |
| RISK-021 | Cardinality explosion in Prometheus from per-tenant labels | 3 | 3 | 9 | sre | Cardinality budget review in code review; tenant_id only on counters not histograms; tenant tier as label proxy | 1 | 2 | 2 | 2026-07-25 (post S4) |
| RISK-022 | Cert-manager TLS expiry without monitoring | 2 | 4 | 8 | sre | Prometheus rule `cert_exporter_not_after - time() < 1209600` (14d) → page; runbook `docs/runbooks/cert-rotation.md` | 1 | 2 | 2 | 2026-07-25 (post S4) |
| RISK-023 | Argo CD compromised → arbitrary cluster state | 1 | 5 | 5 | secops | Restrict Argo CD RBAC; ClusterIP + Ingress with mTLS; OIDC SSO only; audit log shipped to Loki; Kyverno verifies image signatures regardless of manifest claim | 1 | 4 | 4 | 2026-07-25 (post S4) |
| RISK-024 | ESO sync failure leaves stale or default secret | 2 | 4 | 8 | secops | Alertmanager rule on ESO `Status != SecretSynced`; ESO `refreshInterval: 1h`; runbook `docs/runbooks/secret-rotation.md` | 1 | 3 | 3 | 2026-07-25 (post S3) |
| RISK-025 | DR not tested → restore plan unproven | 3 | 5 | 15 | sre | Annual full DR drill; quarterly snapshot restore validation; `docs/runbooks/region-failover.md` | 1 | 3 | 3 | 2027-04-25 |
| RISK-026 | NIS2 24h notification window missed | 2 | 4 | 8 | secops | Pre-filled ACN portal credentials in Vault; template in `docs/INCIDENT-RESPONSE.md`; on-call drill validates window | 1 | 3 | 3 | 2026-10-25 |
| RISK-027 | GDPR DSAR endpoint missing → 30-day window violated | 3 | 4 | 12 | app | `DELETE /api/v1/tenants/me` endpoint + integration test; PII purge cascade; SLA monitoring | 1 | 3 | 3 | 2026-07-25 (post S5) |
| RISK-028 | License-incompatible dep merged silently | 2 | 3 | 6 | secops | `licensee` / `go-licenses` CI scan; `LICENSES.allowed` allowlist; deny GPL-3.0-or-later in commercial path | 1 | 2 | 2 | 2026-07-25 (post S2) |
| RISK-029 | Kyverno admission webhook offline blocks deploys | 2 | 3 | 6 | secops | High-availability Kyverno install (3 replicas, PDB); Alertmanager rule on webhook latency; runbook documents safe fallback (do not disable webhook in prod without IC order) | 1 | 2 | 2 | 2026-07-25 (post S3) |
| RISK-030 | TimescaleDB CAGG refresh lag breaks reporting freshness | 3 | 3 | 9 | sre | `gm_cagg_refresh_duration_seconds` histogram + alert at lag >120s; runbook documents manual refresh procedure | 1 | 2 | 2 | 2026-07-25 (post S4) |

## 2. Accepted risks (Mission II audit)

| ID | Description | Rationale | Annotation |
|---|---|---|---|
| ACC-001 | math/rand in simulator | Deterministic SIM_SEED required for reproducible scenarios; not a crypto path; simulator never in prod | `//nolint:gosec` in `backend/cmd/simulator/main.go`, plan `RISK-001` |
| ACC-002 | Alpine rolling apk | Pinning provides no security benefit (Alpine policy is rolling); SBOM + Trivy block on diff | `# hadolint ignore=DL3018` in `backend/Dockerfile:8`, plan `RISK-002` |

## 3. Closed risks

(none yet — populates as mitigations land and verification runs green for two consecutive sprints)

## 4. Process

- Every PR adding a control references the `RISK-NNN` it mitigates in the PR body.
- Every Dependabot major-version PR adds a checkbox: "Risk register updated | No new risk".
- Monthly secops review: HIGH/CRITICAL changes; new threats from `docs/THREAT-MODEL.md` review.
- Quarterly full re-scoring with sign-off in `docs/office-hours/YYYY-MM-DD.md`.

## 5. Risk-to-rule mapping

| Risk | Doctrine rule(s) | Sprint closure |
|---|---|---|
| RISK-001 | 53, 61 | accepted |
| RISK-002 | 53, 61 | accepted |
| RISK-003 | 19, 39, 62 | S3 |
| RISK-004 | 15, 36, 39 | S4 |
| RISK-005 | 53 | S1 |
| RISK-006 | 19, 39, 57 | S3 |
| RISK-007 | 19, 39, 46 | S3 |
| RISK-008 | 36, 39, 41 | S4 |
| RISK-009 | 19, 39, 60 | S5 |
| RISK-010 | 19, 39, 60 | S5 |
| RISK-011 | 39 | S3 |
| RISK-012 | 18, 40 | S4 |
| RISK-013 | 19, 39 | S3 |
| RISK-014 | 39, 42 | S3 |
| RISK-015 | 41, 42 | S4 |
| RISK-016 | 36, 37 | S4 |
| RISK-017 | 53, 54 | S3 |
| RISK-018 | 23, 63 | S2 |
| RISK-019 | 23, 56, 63 | S4 |
| RISK-020 | 20, 60 | S5 |
| RISK-021 | 18, 40 | S4 |
| RISK-022 | 18, 19, 20 | S4 |
| RISK-023 | 19, 50, 65 | S4 |
| RISK-024 | 19, 62 | S3 |
| RISK-025 | 15, 20, 60 | S5 + annual |
| RISK-026 | 60, 61 | S5 |
| RISK-027 | 39, 64 | S5 |
| RISK-028 | 23, 53 | S2 |
| RISK-029 | 54, 59 | S3 |
| RISK-030 | 18, 33, 36 | S4 |
