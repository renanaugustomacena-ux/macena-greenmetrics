# 0020 — In-cluster identity: cert-manager (SPIRE deferred)

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 19, 39, 62, 65
**Review date:** 2027-04-25

## Context

`docs/MTLS-PLAN.md` calls for in-cluster mTLS for service-to-service hops (backend ↔ Grafana, backend ↔ TimescaleDB application connection where feasible, OTel collector ↔ Tempo). cert-manager is the de facto K8s issuer; SPIRE/SPIFFE is the deeper identity standard.

## Decision

Use **cert-manager + trust-manager** for Phase 1. SPIRE deferred to Phase 2 if multi-cluster federation appears in roadmap.

- `cert-manager` issues per-pod certificates from a self-signed `ClusterIssuer greenmetrics-internal-ca`.
- `trust-manager` distributes the CA bundle to every pod via a `Bundle` mounted as a ConfigMap.
- Edge TLS uses `letsencrypt` ClusterIssuer (Let's Encrypt HTTP-01 via nginx ingress).
- Per-pod cert: 30d duration, 7d renewBefore, ECDSA P-256, `Always` rotation policy.

## Alternatives considered

- **SPIRE / SPIFFE on Phase 1.** Rejected — control-plane operational complexity (SPIRE Server + SPIRE Agent DaemonSet + trust bundle distribution + workload registration entries) > value at our scale. SPIRE shines for multi-cluster federation; we're single-cluster.
- **Linkerd auto-mTLS.** Rejected — REJ-01 (mesh for 2 services).
- **Self-managed CA + scripts.** Rejected — REJ-28 (DIY KMS family).
- **AWS Private CA.** Rejected — vendor lock-in for in-cluster identity; cert-manager is the K8s-native path.

## Consequences

### Positive

- cert-manager is mature, broadly used, well-documented.
- 30d cert rotation routine; trust-manager bundles CA to consumers.
- Switch to SPIRE in Phase 2 doesn't invalidate this — replace ClusterIssuer reference.

### Negative

- Per-pod cert isn't workload-identity in the SPIFFE sense; pod identity = ServiceAccount, not workload-attested.
- Manual mTLS wiring per service (read cert + key from volume; configure TLS).
- Trust bundle drift if trust-manager `Bundle` resource updates lag.

### Neutral

- Edge ACME via Let's Encrypt is already the default — no change there.

## Residual risks

- Renewal failure → cert expires; alert at 14d expiry; runbook `cert-rotation.md`.
- Internal-CA root key compromise → all in-cluster mTLS compromised. Mitigation: root key generated once via cert-manager `selfsigned-issuer`, rotated annually per `docs/SECOPS-RUNBOOK.md`.
- Pod restart needed to pick up new cert. Mitigation: trust-manager + projected volume reload on Secret update where supported.

## References

- `docs/MTLS-PLAN.md`.
- `gitops/base/cert-manager/` (S4 to ship full manifests).
- ADR-0007 (Italian residency).

## Tradeoff Stanza

- **Solves:** in-cluster TLS without service mesh; per-pod cert with rotation.
- **Optimises for:** operational simplicity, K8s-native pattern, switch path to SPIRE preserved.
- **Sacrifices:** workload attestation (vs. SPIRE SPIFFE ID); per-service mTLS wiring effort; pod restart on cert rotate (where projected volume reload not supported).
- **Residual risks:** renewal failure (alert + runbook); root-key compromise (annual rotation, secured generation); cert pickup lag (rolling restart procedure).
