# 0004 — GitOps with Argo CD

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 23, 26 (rejection authority — picked Argo over Flux), 50, 56, 63
**Review date:** 2027-04-25
**Mitigates:** RISK-019.

## Context

CD today is `cd.yml` echo-only placeholder. Manual `kubectl apply` against production is the implicit deploy path. Drift is invisible. Manual approval is the only gate (REJ-24). Plan calls for layered automated gates + image-updater + canary rollout; gitops is the substrate.

## Decision

Adopt **Argo CD** for GitOps reconciliation + **Argo Rollouts** for canary deploys + **argocd-image-updater** for tag promotion.

- `gitops/` directory tree with `base/`, `staging/`, `production/` overlays (Kustomize).
- Argo CD `Application` manifests in `gitops/argocd/applications/`.
- App-of-Apps under `gitops/argocd/applications/root.yaml` so a new env onboards via single PR.
- `selfHeal: true` reverts manual `kubectl apply` mutations within 60s.
- Image-updater watches GHCR; on Cosign-verified signed image landing on `:staging`, opens PR updating digest in `gitops/staging/`. Production promotion stays manual (per Rule 60 — explicit signoff for prod release; per Rule 56 — auto-canary handles the per-deploy decision).
- Argo Rollouts `Rollout` resource for `greenmetrics-backend` + `greenmetrics-frontend`; AnalysisTemplate reads Prometheus SLO burn-rate; auto-undo on AnalysisTemplate failure.
- Argo CD `server.service.type: ClusterIP`; access via Ingress with mTLS + OIDC SSO; dex disabled.
- Argo CD audit log shipped to Loki.

## Alternatives considered

- **FluxCD.** Rejected — Argo CD chosen for native Argo Rollouts integration (canary AnalysisTemplate). Both are valid GitOps engines; Argo's UI is easier for non-platform team members to introspect; image-updater is mature.
- **Manual kubectl + scripts.** Rejected — RISK-019; Rule 56 (automation over process); Rule 63 (immutable infra).
- **Helm only without GitOps.** Rejected — Helm is rendering, not reconciliation; we use Helm for upstream charts (kube-prometheus-stack, ESO, cert-manager) but our own manifests stay in Kustomize.
- **Tekton + custom CD.** Rejected — Argo CD is the de facto standard; no novel value from custom.
- **Spinnaker.** Rejected — operational complexity > value for our scale.

## Consequences

### Positive

- Manual `kubectl apply` reverted within 60s — drift impossible.
- Per-PR `terraform-plan.yml` shows infra diff; PR to `gitops/` shows K8s diff; reviewable.
- Argo Rollouts canary with SLO-driven analysis removes manual approval bottleneck (REJ-24).
- Image-updater automates staging promotion after Cosign verify; production promotion stays explicit.
- Argo CD UI gives non-platform engineers visibility into deploy state.

### Negative

- Argo CD = a new control plane to operate. Phased rollout: Phase 1 staging only; Phase 2 production after Argo CD hardened (RBAC reduced, dex disabled, only OIDC SSO, audit log shipped).
- Image-updater requires GHCR pull credentials in cluster; Cosign verify at admission (Kyverno) is the safety net.
- Rollback via `argocd app sync --revision <prev>` is the operator path; runbook documents.

### Neutral

- Helm + Kustomize coexist (Helm for upstream, Kustomize for ours).

## Residual risks

- Argo CD compromise → arbitrary cluster state. Mitigations:
  - Restrict Argo RBAC to `greenmetrics`+`greenmetrics-staging` namespaces only.
  - ClusterIP + Ingress with mTLS + OIDC SSO only.
  - Audit log shipped to Loki.
  - Kyverno verifies image signatures regardless of manifest content.
  - Branch protection + CODEOWNERS on `gitops/`.
- GitOps repo compromise → arbitrary cluster state. Same mitigations as above; Cosign signature requirement at admission.
- Argo Rollouts AnalysisTemplate misconfigured → bad deploy slips through. Mitigation: AnalysisTemplate is itself versioned + reviewed; tested via Argo Rollouts `Experiment` resource.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.C.50, §6.C.56, §6.C.63.
- `gitops/argocd/` (S4 to ship).
- `RISK-019`, `RISK-023`.

## Tradeoff Stanza

- **Solves:** drift, missing CD automation, manual approval bottleneck.
- **Optimises for:** declarative reconciliation, SLO-driven canary rollout, audit visibility.
- **Sacrifices:** Argo CD operational dependency; staged rollout effort; GHCR credentials in cluster (mitigated by Kyverno).
- **Residual risks:** Argo CD compromise (RBAC + audit + Kyverno); GitOps repo compromise (branch protection + signing); AnalysisTemplate misconfig (versioned + Experiment-tested).
