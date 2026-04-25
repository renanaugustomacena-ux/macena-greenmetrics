# Pipeline Map

**Doctrine refs:** Rule 50 (DevSecOps unified system), Rule 51 (sequence), Rule 53, Rule 54, Rule 56.

## 1. Stages + policy enforcement points

```
┌──────────────────────────────────────────────────────────────────┐
│  CODE                                                            │
│   author commit (GPG-signed)                                     │
│   pre-commit hooks: gitleaks, gofmt, golangci-lint, prettier,    │
│                     eslint, hadolint, terraform_fmt+validate,    │
│                     tflint, tfsec, kubeconform, conftest         │
│                     (audit), actionlint, markdownlint, yamllint  │
│   git push → branch protection: signed commits, CODEOWNERS       │
└──────────────────────┬───────────────────────────────────────────┘
                       ▼
┌──────────────────────────────────────────────────────────────────┐
│  PR (CI)                                                          │
│   pre-commit-ci (mirror; fast-fail < 30s on cache hit)           │
│   backend-lint (golangci-lint + vet)                             │
│   backend-test (race + coverage)                                 │
│   backend-integration (testcontainers Timescale 16)              │
│   backend-property (gopter)                                      │
│   backend-security (RBAC + RLS + boot-refusal)                   │
│   backend-conformance (RFC 7807 + RFC 3339 UTC + UUIDv4 + money) │
│   backend-static (no-float-money, no-panic, no-iface-return)     │
│   backend-build (CGO=0, -trimpath, -ldflags=-s -w)               │
│   frontend-lint + svelte-check                                   │
│   frontend-test (vitest)                                         │
│   frontend-build (SvelteKit)                                     │
│   docker (multi-stage, BuildKit GHA cache)                       │
│   security:                                                      │
│     gitleaks (secret scan)                                       │
│     semgrep (p/default + p/owasp-top-ten + p/golang)             │
│     govulncheck (Go vuln DB)                                     │
│     osv-scanner (cross-source vuln)                              │
│     trivy fs (HIGH/CRITICAL block)                               │
│   sast: CodeQL Go + JS (security-extended)                       │
│   license: go-licenses + npm license-checker (LICENSES.allowed)  │
│   policy-gate-k8s: kubeconform + conftest k8s                    │
│   policy-gate-dockerfile: hadolint + conftest dockerfile         │
│   policy-gate-terraform: tfsec + checkov + conftest terraform    │
│     + infracost (PR comment) + drift detection (nightly)         │
│   actions-lint: actionlint                                       │
│   openapi-lint: redocly lint api/openapi/v1.yaml                 │
│   openapi-compat: tests/contracts/v1_compat_test.go              │
│   config-schema: ajv validate config.schema.json vs .env.example │
│   adr-link-check: markdownlint ADRs + tradeoff-stanza grep       │
│   sbom: Syft SPDX                                                │
│                                                                  │
│   FAIL on any HIGH/CRITICAL or policy violation (no override     │
│   without `override-allowed` label + linked ADR).                │
└──────────────────────┬───────────────────────────────────────────┘
                       ▼
┌──────────────────────────────────────────────────────────────────┐
│  BUILD (CD)                                                       │
│   docker buildx → ghcr.io                                        │
│     tags: ${sha}, ${version}, latest                             │
│     base images digest-pinned                                    │
│     multi-stage distroless nonroot                               │
│   GitHub OIDC → AWS auth (no static keys)                        │
└──────────────────────┬───────────────────────────────────────────┘
                       ▼
┌──────────────────────────────────────────────────────────────────┐
│  ARTEFACT                                                         │
│   Cosign sign --keyless (Sigstore Fulcio cert bound to workflow) │
│   Cosign attest --type spdx (SBOM)                               │
│   slsa-github-generator (SLSA L2 provenance attest)              │
│   trivy image scan (HIGH/CRITICAL block)                         │
│   GHCR stores image + signature + attestations                   │
└──────────────────────┬───────────────────────────────────────────┘
                       ▼
┌──────────────────────────────────────────────────────────────────┐
│  DEPLOY                                                           │
│   argocd-image-updater watches GHCR                              │
│     → on Cosign-verified signed image lands :staging             │
│     → opens PR updating digest in gitops/staging/                │
│     → Argo CD syncs (selfHeal: true) within 60s                  │
│   Argo Rollouts canary 10% → 30% → 100%                          │
│     AnalysisTemplate reads Prometheus SLO burn-rate              │
│     AUTO-ROLLBACK on threshold breach (no human in critical path)│
│   Production: per-release human signoff, then auto-canary        │
│   Kyverno admission verify-images (Sigstore Fulcio + workflow id)│
│     → DENY unsigned image                                        │
│     → DENY pod missing PSS-restricted security context           │
│     → DENY pod missing resources.requests + limits               │
└──────────────────────┬───────────────────────────────────────────┘
                       ▼
┌──────────────────────────────────────────────────────────────────┐
│  RUN                                                              │
│   K8s pod runs (distroless nonroot)                              │
│   Falco DaemonSet eBPF runtime detection                         │
│     → custom rules in policies/falco/greenmetrics-rules.yaml     │
│   ESO refresh secrets every 1h                                   │
│   Prometheus scrape (kube-prometheus-stack)                      │
│   OTel collector → Tempo + Loki + Prometheus remote_write        │
│   Alertmanager → Slack/PagerDuty (per severity)                  │
│   Audit log → DB + Loki (+ S3 audit bucket Object Lock 5y)       │
│   Application audit_log table append-only (RLS + trigger)        │
│   Continuous verification: nightly `make verify` against main    │
└──────────────────────────────────────────────────────────────────┘
```

## 2. Feedback loop (runtime → backlog)

```
Falco event (Critical) → Loki → Alertmanager → GitHub issue (auto)
   → PR with fix or ADR for accepted residual
   → CI pipeline applies the change
   → Cosign-signed image deployed
   → Argo Rollouts canary
   → AnalysisTemplate green → 100% rollout
   → next Falco event compares against the new baseline
```

## 3. Ownership

Every gate has a CODEOWNERS-routed owner per `docs/RACI.md`. Pipeline changes require `@greenmetrics/platform-team` + `@greenmetrics/secops` review.

## 4. Anti-patterns rejected

- Skipping a stage to "ship faster" — every stage is a gate; bypass requires explicit ADR.
- Manual approval as the only gate — REJ-24.
- Unpinned Action tags — REJ-29.
- Cosign with custom keys — REJ-27.
- Production deploy via direct `kubectl` — RISK-019; Argo CD reverts.
