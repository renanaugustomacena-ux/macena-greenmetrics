# GDPR — Compliance Evidence Pack

**Owner:** `@greenmetrics/secops`, DPO.
**Doctrine refs:** Rule 19, Rule 39, Rule 60, Rule 61, Rule 65.
**Citation source:** Reg. UE 2016/679 (GDPR) + D.Lgs. 196/2003 (101/2018) Codice in materia di protezione dei dati personali.

## 1. Lawful basis

- Contract performance (Reg. UE 2016/679 art. 6.1.b) for SaaS service to tenants.
- Legal obligation (art. 6.1.c) for D.Lgs. 102/2014 audit + CSRD reporting that the tenant must produce.

## 2. Mapped controls + evidence

| Requirement | Implementation | Evidence |
|---|---|---|
| Encryption at rest | RDS `storage_encrypted = true` + KMS key; S3 SSE-KMS; cluster-level secrets via ESO + KMS | Terraform `terraform/modules/rds-timescale/main.tf` + `terraform/modules/s3/main.tf` |
| Encryption in transit | TLS 1.3 ingress; `sslmode=require` to RDS; mTLS plan in S4 | `terraform/versions.tf` config schema requires sslmode; `docs/MTLS-PLAN.md` |
| Access control | JWT HS256 + KID rotation + RBAC per `internal/security/rbac.go` + Postgres RLS | `tests/security/rbac_test.go` + `tests/security/rls_isolation_test.go` |
| Audit log | `audit_log` table append-only + S3 Object Lock 5y | `migrations/00010_audit_lock.up.sql` |
| Data retention | TimescaleDB retention policy 90d/1y/3y/10y per data class | `migrations/0004_retention.sql` |
| DSAR (data subject access request) | `DELETE /api/v1/tenants/me` endpoint (S5 to ship) — purges PII cascading | `tests/integration/gdpr_dsar_test.go` (S5) |
| Breach notification 72h | Template in `docs/INCIDENT-RESPONSE.md` §4 | DPO contact in `docs/CONTACTS.md` |
| Pseudonymisation | `tenant_id` UUIDv4 stable; user emails not exposed beyond authentication path | code review + RBAC |
| Data minimisation | `readings` carry no PII (numeric values + timestamps); user emails are admin-only | `internal/api/v1/dto/` mappers |

## 3. DSAR procedure

1. DPO receives request via dpo@greenmetrics.it.
2. Identity verification (out-of-band).
3. Operator runs `DELETE /api/v1/tenants/<tenant_id>/users/<email>` (admin RBAC required).
4. Cascade purge: user → audit_log redaction → session revocation → JWT KID rotation if requested.
5. DPO confirms completion within 30 days.

`tests/integration/gdpr_dsar_test.go` (S5) verifies cascade.

## 4. Breach notification template

In `docs/INCIDENT-RESPONSE.md` §4 + per-template files in `docs/templates/incident-comms/`. NIS2 + GDPR templates colocated.

## 5. Anti-patterns rejected

- "Soft delete" PII (mark deleted but keep) — REJ; GDPR right-to-erasure means actual purge.
- Audit log redaction skipping audit log of redaction — REJ; redaction itself is logged.
- Email user-id leaked in error messages — REJ; pseudonymisation rule.
- DSAR > 30 days — REJ; SLA.
