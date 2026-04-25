# Trust Boundaries Catalog

**Owner:** `@greenmetrics/secops`.
**Doctrine refs:** Rule 19 (security as structural), Rule 39 (backend security as core), Rule 65 (regulated quality).
**Cross-ref:** `docs/THREAT-MODEL.md` STRIDE per surface; `docs/RISK-REGISTER.md` per-control mitigation.

A trust boundary is a crossing where data, identity, or privilege transitions between zones with different trust assumptions. Every crossing is named, owned, and gated.

| ID | Crossing | Direction | Protocol | Authn | Authz | Encryption | Audit | Threat ref |
|---|---|---|---|---|---|---|---|---|
| TB-01 | Browser → CloudFront edge | inbound | HTTPS | n/a (public assets) / JWT (XHR) | WAF rate-based | TLS 1.3 | CloudFront access log → S3 (KMS) | `docs/THREAT-MODEL.md` §3.3 |
| TB-02 | CloudFront → AWS WAF | inbound | HTTP | n/a | managed core rules + custom rate rules | inside AWS network | WAF logs → S3 (KMS) | DDoS amplification |
| TB-03 | WAF → nginx ingress | inbound | HTTPS (re-encrypt) | n/a | NetworkPolicy | TLS 1.3 (cert-manager / Let's Encrypt) | nginx access log → Loki | TLS downgrade |
| TB-04 | Ingress → Backend pod | inbound | HTTP (mTLS planned S4) | JWT HS256 + `kid` | RBAC `RequirePermission` | mTLS planned (`docs/MTLS-PLAN.md`) | audit_log table per state-changing op | TB-04 spoofing |
| TB-05 | Backend → TimescaleDB | outbound | PG wire (sslmode=require in prod) | `app_user` PG role (BYPASSRLS=false) | RLS per-row policy `tenant_isolation` | TLS to RDS | none (DB internal) | RISK-007 |
| TB-06 | Backend → Grafana admin API | outbound | HTTPS | basic auth via ESO-managed creds | n/a (admin only) | TLS internal | Grafana admin audit log → Loki | TB-06 |
| TB-07 | Backend → ISPRA factor refresh | outbound | HTTPS | none (public) | n/a | TLS pinning | every fetch logged + hash recorded | RISK-004 |
| TB-08 | Backend → Terna grid mix | outbound | HTTPS | API key (ESO) | n/a | TLS pinning | every fetch logged + hash recorded | RISK-004 |
| TB-09 | Backend → E-Distribuzione POD lookup | outbound | HTTPS | API key (ESO) | n/a | TLS pinning | every fetch logged | RISK-004 |
| TB-10 | Backend → Redis (Asynq + rate limit) | outbound | RESP / TLS (rediss in prod) | password auth (ESO) | n/a | TLS in prod | Redis SLOWLOG sampled to Loki | Redis compromise → job poisoning |
| TB-11 | Modbus ingestor → Modbus host | outbound | Modbus TCP / RTU | none | NetworkPolicy egress allowlist | none (legacy protocol) | per-poll logged | RISK-008, untrusted-gateway register injection |
| TB-12 | M-Bus ingestor → serial device | outbound | M-Bus on /dev/ttyS* | physical device | n/a | physical | per-frame logged | physical-access compromise |
| TB-13 | OCPP central system → Backend WebSocket | inbound | WSS | per-CP token (ESO) | per-CP scope | TLS | per-message logged + audit row | RISK-008 |
| TB-14 | Pulse webhook → Backend | inbound | HTTPS | HMAC-SHA256 signature (constant-time compare) + `Idempotency-Key` | rate limit | TLS | per-pulse audit row | RISK-011 (timing oracle) |
| TB-15 | OTel collector → Tempo / Prometheus / Loki | outbound | OTLP gRPC / Prometheus remote_write / Loki HTTP | mTLS (planned S4) | n/a | TLS internal | self-monitored | telemetry tampering |
| TB-16 | CI runner → GHCR | outbound | HTTPS | GitHub OIDC token | repo + ref scoped | TLS | GitHub audit log | RISK-005, RISK-017 |
| TB-17 | GHCR → Cluster (image pull) | outbound | HTTPS | imagePullSecret (Argo CD reads) | Kyverno admission `verify-images` | TLS | image pull audit + Kyverno log | RISK-017 |
| TB-18 | Cluster → AWS Secrets Manager | outbound | HTTPS | IRSA per-pod (ESO) | scoped to `greenmetrics/prod/*` | TLS internal | CloudTrail data event on `GetSecretValue` | RISK-024 |
| TB-19 | Cluster → AWS KMS | outbound | HTTPS | IRSA per-pod | resource policy + condition `kms:ViaService` | TLS internal | CloudTrail data event on `Decrypt` | RISK-010 |
| TB-20 | Operator → AWS Console | inbound | HTTPS | SSO + MFA | least-privilege IAM + break-glass `max_session_duration=1h` | TLS | CloudTrail | RISK-006 |
| TB-21 | Operator → kubectl (production) | inbound | HTTPS via VPN | OIDC SSO | EKS RBAC | mTLS over VPN | EKS audit log → Loki | RISK-006 |
| TB-22 | Operator → Argo CD UI | inbound | HTTPS via Ingress | OIDC SSO | Argo RBAC | TLS + mTLS edge | Argo audit log → Loki | RISK-023 |
| TB-23 | Audit log writer → audit_log table | outbound | PG wire | audit_middleware only | append-only via WITH CHECK + trigger | TLS | n/a (this is the audit) | RISK-009 |
| TB-24 | S3 audit bucket | inbound (write) / outbound (read) | HTTPS | IRSA + bucket policy + Object Lock | Object Lock `compliance` mode 5y | TLS + KMS | CloudTrail data event | RISK-009 |
| TB-25 | Backend → Sigstore Rekor (Cosign verify path indirect) | n/a | n/a | n/a | trust chain to Sigstore root (annual review) | TLS | n/a | trust-chain compromise |

## Crossings under construction (close in S3/S4)

| ID | Crossing | Closes in | Tracked under |
|---|---|---|---|
| TB-04 mTLS | Ingress → Backend mTLS | S4 | `docs/MTLS-PLAN.md` |
| TB-15 mTLS | OTel exports | S4 | `docs/MTLS-PLAN.md` |
| TB-14 HMAC | Pulse webhook constant-time HMAC | S3 | RISK-011 |
| TB-04 KID | JWT KID rotation | S3 | `docs/JWT-ROTATION.md` |
| TB-05 RLS | Postgres RLS enforcement | S3 | `migrations/00006_rls_enable.go` |
| TB-04 RBAC | RBAC middleware on every protected route | S3 | `internal/security/rbac.go` |

## Operator commands per boundary

```bash
# TB-04 verify JWT issuance
curl -sX POST $API/api/v1/auth/login -d '{...}' | jq .access_token | cut -d. -f1 | base64 -d | jq

# TB-05 verify RLS isolation
psql -c "SELECT set_config('app.tenant_id','TENANT_A',true); SELECT count(*) FROM meters;"  # only A's meters

# TB-17 verify image admission
kubectl apply -f unsigned-test.yaml   # expect: admission webhook denial

# TB-18 verify IRSA scope
aws iam simulate-principal-policy \
  --policy-source-arn arn:aws:iam::ACCT:role/greenmetrics-backend-irsa \
  --action-names secretsmanager:GetSecretValue \
  --resource-arns arn:aws:secretsmanager:eu-south-1:ACCT:secret:greenmetrics/prod/jwt-AAAAA
```

## Anti-patterns rejected

- Implicit trust on internal hops — every crossing is named.
- "It's behind the VPN" as a security argument — VPN is one defence; not the only one (RISK-006).
- Long-lived API keys for service-to-service — use IRSA + short-lived OIDC tokens.
- Trust-on-first-use for outbound TLS — pin known roots.
