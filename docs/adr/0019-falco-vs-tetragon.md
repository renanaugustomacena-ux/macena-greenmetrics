# 0019 — Runtime detection: Falco (with Tetragon fallback)

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 19, 39, 58, 65, 66
**Review date:** 2027-04-25

## Context

GreenMetrics has no runtime detection layer. Falco events feed `docs/INCIDENT-RESPONSE.md`; Falco rules in `policies/falco/greenmetrics-rules.yaml` cover unexpected process spawn, /etc writes, outbound to non-allowlisted hosts, shell in distroless container, and mount syscalls.

## Decision

Deploy **Falco** as a DaemonSet with custom rules. Fall back to **Tetragon** (eBPF-native) only if Falco kernel-module compatibility fails on the EKS Bottlerocket node OS.

## Alternatives considered

- **Falco-only commitment.** Rejected — kernel-module compat on AL2023 / Bottlerocket has been historically unreliable; need fallback path.
- **Tetragon as primary.** Considered. Pro: eBPF-native = no kernel module compat issues. Con: smaller ecosystem of community rules; documentation thinner. Pick Falco first; switch if Falco fails.
- **AWS GuardDuty + EKS Runtime Monitoring.** Considered. Pro: managed. Con: vendor lock-in; less customisable; doesn't substitute for in-cluster Falco rules tied to GreenMetrics-specific patterns.
- **No runtime detection.** Rejected — Rule 58 (security observability).
- **Aqua / Sysdig / Trend Micro.** Rejected per Rule 66 (CNCF stack covers it).

## Consequences

### Positive

- Per-pod runtime visibility with custom GreenMetrics rules.
- Events ship to Loki via Promtail; Alertmanager fires on Critical/Error.
- Single runtime detection layer; no tool sprawl.

### Negative

- Falco kernel-module install requires privileged init container.
- Custom rule maintenance (`policies/falco/greenmetrics-rules.yaml`).
- False positives initially; tuning required.

### Neutral

- CNCF graduated project; long-term support reasonable.

## Residual risks

- Kernel-module install fails on a specific node OS update → Tetragon switch.
- Rule false negative: real intrusion missed. Mitigation: layered defence (Kyverno admission + NetworkPolicy + pgx+RLS + IRSA scope + audit log).

## References

- `policies/falco/greenmetrics-rules.yaml`.
- `gitops/base/falco/` (Helm-managed, S5).
- Tetragon: <https://github.com/cilium/tetragon>.

## Tradeoff Stanza

- **Solves:** runtime visibility into container-side anomalies (process spawn, file write, outbound).
- **Optimises for:** CNCF-native stack alignment; per-pod custom rule expression.
- **Sacrifices:** kernel-module compat risk; rule tuning effort; privileged init container.
- **Residual risks:** kernel-module install failure (Tetragon fallback); rule false negative (layered defences).
