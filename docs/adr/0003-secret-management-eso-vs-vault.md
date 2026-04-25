# 0003 — Secret management: AWS Secrets Manager + ESO vs Vault

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 19, 39, 62, 67
**Review date:** 2027-04-25
**Mitigates:** RISK-003, RISK-013, RISK-024.

## Context

GreenMetrics needs runtime secrets distributed to backend, worker, and Grafana pods: JWT secret + KID, DB credentials, Pulse webhook secret, Redis URL, Grafana admin password, OCPP CS URL, SunSpec address. Today these are placeholders in `k8s/secret.yaml` (`REPLACE_VIA_EXTERNAL_SECRETS`). Compose stack uses `:?must-be-set` env-var refusal; production K8s has no working secret pipeline.

## Decision

Adopt **AWS Secrets Manager** as the secret source of truth + **External Secrets Operator (ESO)** as the in-cluster materialisation engine. IRSA-backed ESO ServiceAccount (`terraform/modules/iam-irsa/`). ClusterSecretStore in `gitops/base/external-secrets/clustersecretstore.yaml`. ExternalSecret per workload mapping `greenmetrics/prod/*` → K8s Secrets.

## Alternatives considered

- **HashiCorp Vault.** Rejected — adds entire control plane (storage backend, unseal, audit, access policies) to operate. We are a small team; AWS already in stack; Secrets Manager native rotation Lambda + AWS-managed HA + IAM federation via IRSA is a turnkey path. Tradeoff: vendor lock-in to AWS Secrets Manager API; mitigation = ESO abstracts the provider, switch path is replacing `SecretStore` reference.
- **K8s native Secrets only.** Rejected — base64-encoded etcd entries; no rotation primitive; no audit trail beyond K8s RBAC; no integration with cloud-native KMS without manual wiring.
- **CSI Secrets Store with AWS provider.** Rejected — pod-volume mount replaces env-var idiom; restart on rotation requires custom controller; ESO with `creationPolicy: Owner` + secret-watching pods is simpler.
- **SOPS-encrypted git secrets.** Rejected — secret material in git history is a structural risk; key management for SOPS is its own problem; rotation requires git rewrite.
- **Aruba Cloud KMS.** Rejected for v1 — Italian sovereignty alternative documented but not deployed; revisit if Aruba is the production target (ADR-007).

## Consequences

### Positive

- Single secret source of truth in Secrets Manager; rotations propagate to pods within ESO `refreshInterval` (1h).
- IRSA per-pod (Rule 57) — no static AWS keys in cluster.
- Audit trail via CloudTrail data events on `GetSecretValue`.
- ESO ExternalSecret manifests live in git; declarative + reviewable.
- Switch path: replace ClusterSecretStore reference; Vault provider is a one-liner change in ESO.

### Negative

- ESO is one more control plane to operate; CRDs + controller pod.
- AWS lock-in for the secret store backend (mitigation: ESO abstraction).
- ESO sync failures are silent unless monitored (mitigation: Alertmanager rule on `Status != SecretSynced`; RISK-024).
- 1h refresh interval means rotation propagation delay; emergency rotation requires manual force-sync annotation (`task rotate:secrets`).

### Neutral

- ExternalSecret YAML is declarative — fits the GitOps story (ADR-004).
- Per-secret pricing (Secrets Manager €0.40/secret/month + €0.05/10k API calls) is a few euros/month at our scale.

## Residual risks

- ESO controller compromise → access to all secrets in mapped paths. Mitigation: scoped IRSA (`secretsmanager:GetSecretValue` on `greenmetrics/prod/*` only); Falco rule on unexpected `GetSecretValue` patterns.
- Secrets Manager API outage halts rotation propagation. Mitigation: existing pod env retains last value; alert at `aws_secret_age_seconds > 7776000` (90d).
- Bus factor on ESO operator knowledge. Mitigation: `docs/SECOPS-RUNBOOK.md` documents force-sync, troubleshooting; runbook `docs/runbooks/secret-rotation.md` (S4).

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.A.19, §6.C.62.
- `gitops/base/external-secrets/`.
- `terraform/modules/iam-irsa/`.
- `terraform/modules/secrets/main.tf` (existing).
- `RISK-003`, `RISK-013`, `RISK-024`.
- ADR-0007 (Italian residency).
- ADR-0016 (JWT KID rotation).

## Tradeoff Stanza

- **Solves:** sentinel secrets reaching production; missing rotation primitive; missing audit trail.
- **Optimises for:** managed-service operability; IRSA-bound least privilege; declarative GitOps fit.
- **Sacrifices:** AWS lock-in (mitigated by ESO abstraction); ESO control plane to operate; 1h refresh latency.
- **Residual risks:** ESO controller compromise (scoped IRSA + Falco rule); Secrets Manager outage (alert at age threshold); bus factor (runbook + drill).
