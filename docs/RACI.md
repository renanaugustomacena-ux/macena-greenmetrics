# GreenMetrics RACI

**Authoring date:** 2026-04-25
**Doctrine refs:** Rule 9 (role definition), Rule 49 (secops authority).
**Review cadence:** quarterly.

R = Responsible (does the work). A = Accountable (signs off). C = Consulted (gives input). I = Informed (kept in the loop).

## 1. Subsystem RACI

| Subsystem | R | A | C | I |
|---|---|---|---|---|
| API contract (`api/openapi/v1.yaml`) | platform | platform-lead | app, secops | sre |
| Backend handlers (`backend/internal/handlers/`) | app | app-lead | platform | sre |
| Backend services (`backend/internal/services/`) | app | app-lead | platform | sre |
| Backend repository (`backend/internal/repository/`) | platform | platform-lead | data, app | sre |
| Backend domain (`backend/internal/domain/`) | app | app-lead | platform | sre |
| DB schema + migrations (`backend/migrations/`) | data | platform-lead | secops | sre, app |
| Postgres RLS policies | secops | secops-lead | platform, data | app |
| RBAC permissions registry | secops | secops-lead | app, platform | sre |
| K8s manifests (`k8s/`) | platform | platform-lead | secops, sre | app |
| GitOps (`gitops/`) | platform | platform-lead | secops, sre | app |
| Terraform modules (`terraform/`) | platform | platform-lead | secops | sre |
| IRSA + IAM | secops | secops-lead | platform | sre |
| Secrets distribution (ESO + Secrets Manager) | secops | secops-lead | platform | sre, app |
| JWT KID rotation | secops | secops-lead | app, platform | sre |
| TLS / cert-manager | platform | platform-lead | secops, sre | app |
| Grafana dashboards (`grafana/`) | sre | sre-lead | platform, app | secops |
| Prometheus alert rules (`monitoring/prometheus/rules/`) | sre | sre-lead | platform, secops | app |
| Alertmanager routing | sre | sre-lead | secops | platform, app |
| Loki + log shipping | sre | sre-lead | secops | platform |
| OTel collector + Tempo | sre | sre-lead | platform | app |
| Falco + runtime detection | secops | secops-lead | sre | platform |
| Policy bundles (`policies/conftest/`, `policies/kyverno/`) | secops | secops-lead | platform | sre, app |
| CI workflows (`.github/workflows/ci.yml`) | platform | platform-lead | secops | sre, app |
| CD workflows (`.github/workflows/cd.yml`, supply-chain.yml, dast.yml) | platform | platform-lead | secops | sre, app |
| Cosign signing + SLSA provenance | secops | secops-lead | platform | sre |
| Pre-commit framework | platform | platform-lead | secops, app | — |
| Devcontainer + bootstrap | platform | platform-lead | app | sre |
| Runbooks (`docs/runbooks/`) | sre | sre-lead | platform, secops | app |
| Incident response (`docs/INCIDENT-RESPONSE.md`) | secops | secops-lead | sre, platform | legal, dpo |
| On-call rota | sre | sre-lead | platform | app |
| Postmortem authoring | sre | secops-lead | involved teams | repo owner |
| Threat model (`docs/THREAT-MODEL.md`) | secops | secops-lead | platform, app | sre |
| Risk register (`docs/RISK-REGISTER.md`) | secops | secops-lead | platform, sre, legal | app |
| ADRs (`docs/adr/`) | platform | platform-lead | secops | app, sre |
| Cost model (`docs/COST-MODEL.md`) | platform | platform-lead | finance | repo owner |
| Capacity model (`docs/CAPACITY.md`) | platform | platform-lead | sre, data | app |
| SLO catalog (`docs/SLO.md`, `docs/SLI-CATALOG.md`) | sre | sre-lead | platform | app |
| Italian compliance citations (`docs/ITALIAN-COMPLIANCE.md`) | legal | secops-lead | secops, dpo | repo owner |
| GDPR DSAR endpoint | app | secops-lead | secops | dpo |
| NIS2 incident reporting | secops | secops-lead | legal, dpo | repo owner |
| ESRS audit query | app | secops-lead | data | repo owner |
| Backup + DR | sre | platform-lead | secops | repo owner |
| Region failover | sre | platform-lead | secops | repo owner |
| Chaos engineering (`chaos/`) | sre | platform-lead | secops | app |
| DAST (`dast/`) | secops | secops-lead | sre | app |
| Pentest engagement | secops | repo owner | secops | legal |

## 2. Process RACI

| Process | R | A | C | I |
|---|---|---|---|---|
| Sprint planning | platform-lead | repo owner | all leads | team |
| Quarterly platform office hours | platform | platform-lead | all teams | repo owner |
| Quarterly DevSecOps review | secops | secops-lead | all teams | repo owner |
| Quarterly architectural review (backend) | app | app-lead | platform | sre |
| Annual DR drill | sre | platform-lead | secops | repo owner |
| Annual regulatory citation re-verification | legal | secops-lead | repo owner | dpo |
| Branch protection updates | platform | platform-lead | secops | repo owner |
| Renovate / Dependabot triage | secops | secops-lead | platform | app |
| Secret rotation (quarterly) | secops | secops-lead | platform | sre |
| Cost audit (monthly) | platform | platform-lead | finance | repo owner |
| Runbook freshness review (monthly) | sre | sre-lead | platform | secops |
| Risk register review (monthly) | secops | secops-lead | platform | repo owner |
| Threat model review (quarterly) | secops | secops-lead | app, platform | sre |

## 3. Team membership (initial)

- `@greenmetrics/platform-team`: `@ciupsciups` (lead).
- `@greenmetrics/app-team`: `@ciupsciups` (lead).
- `@greenmetrics/sre`: `@ciupsciups` (lead).
- `@greenmetrics/secops`: `@ciupsciups` (lead) — bus factor noted; rotation plan in `docs/SECOPS-RUNBOOK.md`.
- `@greenmetrics/data-team`: `@ciupsciups`.
- `@greenmetrics/legal`: TBD external counsel.

Initial single-operator structure is honest. Roles will populate as the team grows; RACI is the source of truth for which hat to wear when. Bus-factor mitigations: documented runbooks, no tribal knowledge, ADRs as the historical record.
