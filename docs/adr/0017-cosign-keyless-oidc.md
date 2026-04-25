# 0017 — Cosign keyless via Sigstore OIDC

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 53, 54, 65, 66, 67
**Review date:** 2027-04-25
**Mitigates:** RISK-005, RISK-017.

## Context

`.github/workflows/cd.yml:55-57` had Cosign installer present but no `cosign sign` invocation — placeholder. Plan calls for signed images at every deploy + Kyverno admission verifying signatures bound to the GitHub workflow identity.

## Decision

Adopt **Cosign keyless** via **Sigstore Fulcio + Rekor**. GitHub OIDC token is exchanged for a short-lived signing certificate bound to the workflow identity (subject: `https://github.com/<owner>/greenmetrics/.github/workflows/cd.yml@<ref>`). Signatures + attestations stored as OCI artefacts alongside the image in GHCR. Rekor transparency log entry per signature.

- `cosign sign --yes <image>:<sha>`
- `cosign attest --predicate sbom.spdx.json --type spdx <image>`
- `slsa-github-generator/.github/workflows/generator_container_slsa3.yml@<sha>` for SLSA L2 provenance attest.
- Verify at admission: Kyverno `verify-images` policy in `policies/kyverno/verify-images.yaml`.

## Alternatives considered

- **Cosign with custom KMS-backed keys (AWS KMS / GCP KMS).** Rejected — adds key-management burden (rotation, access control, audit) without material security improvement. Sigstore Fulcio + Rekor + workflow-identity binding is sufficient for our threat model.
- **Notary v2 / Notation.** Considered. Pro: OCI-aligned. Con: smaller adoption; Cosign + Sigstore is the de facto standard.
- **No signing; rely on Trivy + SBOM only.** Rejected — Rule 53; SLSA L2 requires provenance attestation.
- **Cosign + signing via dedicated CI runner with offline key.** Rejected — operationally complex; defeats the audit trail benefit of Rekor.

## Consequences

### Positive

- Every deployed image traceable to a specific GitHub workflow run.
- Sigstore Rekor transparency log — public audit trail.
- Kyverno admission denies unsigned images regardless of who pushed them.
- No long-lived signing key to rotate / leak.

### Negative

- Sigstore Fulcio CA chain trust requires annual review (`docs/SECOPS-RUNBOOK.md` annual section).
- GitHub OIDC issuer compromise → forged signatures possible. Mitigation: trust policy bound to repo + ref + environment.
- Rekor entries are public — image existence + signing time are public knowledge. Acceptable for our threat model (private image content, public existence).

### Neutral

- Cosign tooling is well-maintained; CNCF project.

## Residual risks

- Sigstore root key compromise → entire trust chain compromised. Mitigation: annual cert pinning review; emergency runbook to switch to alt root.
- GHCR compromise → unsigned image push. Mitigation: Kyverno admission blocks; Renovate verifies signatures on inbound dependency images.
- Workflow identity widening (e.g. forks) → signature subject mismatch. Mitigation: branch protection prevents fork PR from running CD; Kyverno regex strictly bounds subject.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.C.53.
- `.github/workflows/supply-chain.yml`.
- `policies/kyverno/verify-images.yaml`.
- `docs/SUPPLY-CHAIN.md`.
- ADR-0018 (SLSA L2/L3).
- `RISK-005`, `RISK-017`.

## Tradeoff Stanza

- **Solves:** unsigned images deployable to production; missing supply-chain attestation chain.
- **Optimises for:** zero-key-management overhead; public audit trail via Rekor; admission-time enforcement.
- **Sacrifices:** Sigstore root chain dependency (annual review); GitHub OIDC issuer dependency; public Rekor visibility.
- **Residual risks:** root chain compromise (annual review + emergency switch); GHCR compromise (Kyverno blocks); subject widening (branch protection + regex bound).
