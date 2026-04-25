# Changelog

All notable changes to GreenMetrics will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Doctrine refs: Rule 21 (evolution + change management), Rule 34 (backend contracts), `docs/API-VERSIONING.md`.

## [Unreleased]

### Added

- **Repository extraction.** Extracted from monorepo `macena-tools-lavoro` into a standalone private repository at `github.com/renanaugustomacena-ux/macena-greenmetrics`. Fresh git history; loses parent monorepo commit log but does not leak history of sibling projects.
- **IP-protection surface.** Replaced previous MIT licence with proprietary "All Rights Reserved" notice attributing copyright exclusively to Renan Augusto Macena (LICENSE). Added NOTICE (trademark + dependency attribution + SBOM pointer), AUTHORS (sole-author + citation format), COPYRIGHT (short pointer), TRADEMARK.md (full trademark policy), top-level CONTRIBUTING.md (CLA-style contributor terms with assignment / licence grant + patent grant + governing law), top-level SECURITY.md (responsible-disclosure policy with NIS2-aligned timelines), CODE_OF_CONDUCT.md (Contributor Covenant v2.1), .github/copyright-header.txt (per-file header template), .github/ISSUE_TEMPLATE/{bug,feature,security}.md (security template redirects to private channel), .github/FUNDING.yml (placeholder).
- **Doctrine adoption.** Full 60-rule operational doctrine plan at `~/.claude/plans/my-brother-i-would-flickering-coral.md` plus Sprint S1 deliverables: charters (`docs/TEAM-CHARTER.md`, `docs/SECOPS-CHARTER.md`), RACI (`docs/RACI.md`), CODEOWNERS, PR template, layers map (`docs/LAYERS.md`, `docs/layers.yaml`), threat model (`docs/THREAT-MODEL.md`), risk register (`docs/RISK-REGISTER.md`), platform initiative workflow (`docs/PLATFORM-INITIATIVE-WORKFLOW.md`), supply chain trust model (`docs/SUPPLY-CHAIN.md`), CONTRIBUTING guide (`docs/CONTRIBUTING.md`).
- **ADR scaffolding.** `docs/adr/0000-template.md`, `docs/adr/README.md`, `docs/adr/REJECTED.md` (35 anti-patterns), and ADR-0001 (platform doctrine adoption).
- **Pre-commit framework.** `.pre-commit-config.yaml` covering gitleaks, gofmt, goimports, golangci-lint, prettier, eslint, shellcheck, hadolint, terraform-fmt/validate/tflint/tfsec, kubeconform, conftest, actionlint, markdownlint, yamllint, conventional-pre-commit.
- **Dependabot.** Weekly bumps for gomod / npm / github-actions / docker / terraform.
- **CI extensions.** `pre-commit-ci`, `policy-gate-k8s` (kubeconform + conftest), `policy-gate-dockerfile` (hadolint), `policy-gate-terraform` (tfsec + checkov), `actions-lint`, `openapi-lint`, `config-schema`, `adr-link-check`, `license` (go-licenses + npm license-checker), `osv-scanner` cross-source vuln check.
- **Supply chain.** `.github/workflows/sast.yml` (CodeQL Go + JS, security-extended), `.github/workflows/supply-chain.yml` (Cosign keyless sign + SLSA L2 provenance + Trivy image post-build).
- **OpenAPI 3.1 canonical spec** at `api/openapi/v1.yaml`. swag remains as Swagger UI shell only.
- **Migration tooling adopted: pressly/goose** (ADR-0005). New Makefile targets `migrate-up`, `migrate-down N=`, `migrate-status`, `migrate-create NAME=`.
- **Conftest policy bundles** at `policies/conftest/{k8s,dockerfile,terraform}/` enforcing PSS-restricted, resource limits, probes, image-digest pinning, RDS encryption, mandatory cost-allocation tags. Audit mode in S2; enforce mode in S3.
- **Kyverno admission policies** at `policies/kyverno/` enforcing image signature verification, PSS-restricted, required resources.
- **Devcontainer** at `.devcontainer/` for VSCode + Codespaces with full toolchain (Go, Node, terraform, kubectl, helm, argocd, cosign, syft, trivy, gosec, k6, k9s).
- **Bootstrap script** at `scripts/dev/bootstrap.sh` — fresh-clone bootstrap with cryptographic-secret generation, migration application, smoke test. Idempotent, ≤ 5 min target.
- **Taskfile.yaml** at repo root + reorganised `backend/Makefile` exposing the stable CLI contract (`docs/CLI-CONTRACT.md`).
- **Terraform state backend activated** (S3 + DynamoDB lock + KMS) per `terraform/versions.tf`; `terraform/bootstrap/` for one-shot creation. Mitigates RISK-018.
- **IRSA module** at `terraform/modules/iam-irsa/` for per-pod least-privilege IAM (backend, ESO, Argo CD).
- **GitHub OIDC module** at `terraform/modules/github-oidc/` for keyless CI → AWS auth bound to repo + ref claims.
- **Capacity model** at `docs/CAPACITY.md` with worked small/medium/large/stretch examples.
- **Cost model** at `docs/COST-MODEL.md` with per-tenant unit economics + AWS Budgets + waste detection.
- **Service catalog** at `docs/SERVICE-CATALOG.md` with 25 services + per-service SLO defaults.
- **Platform defaults** at `docs/PLATFORM-DEFAULTS.md` enumerating opinionated stack choices.
- **Extension points** at `docs/EXTENSION-POINTS.md` — 6 documented seams with interface templates.
- **Abstraction ledger** at `docs/ABSTRACTION-LEDGER.md` — 16 active abstractions with cost/leverage rationale.
- **Schema evolution policy** at `docs/SCHEMA-EVOLUTION.md` — forward-only, hot-path additive-only, CAGG-aware.
- **API versioning policy** at `docs/API-VERSIONING.md` — SemVer, RFC 8594 Sunset header, parallel-run window.
- **Debug guide** at `docs/DEBUG.md` — backend, frontend, auth, RLS, K8s, Argo CD, Cosign.
- **Troubleshooting tree** at `docs/TROUBLESHOOTING.md` — decision tree per symptom.
- **Quality bar** at `docs/QUALITY-BAR.md` — non-negotiable invariants.
- **Doctrine checklist** at `docs/backend/doctrine-checklist.md` — reviewer-signed per backend PR.

### Changed

- **README.md** — added documentation index linking new docs.
- **frontend/Dockerfile** — added supply-chain comment block + RISK-002/-005 cross-reference.
- **.github/workflows/ci.yml** — restructured around new gate jobs (pre-commit-ci, policy-gate-*, openapi-lint, license, osv-scanner, adr-link-check); SHA-pinning will land via Dependabot weekly migration (RISK-005).
- **backend/Makefile** — added `help`, `verify`, `migrate-{up,down,status,create}`, `test-{integration,e2e,property,security,conformance,static}`, `bench`, `precommit`, `policy-check`, `audit-rejections`, `layers-doc`.

### Security

- (none in S1; S3 ships RLS, RBAC, idempotency, Cosign verify enforce.)

### Deprecated

- (none yet.)

### Removed

- (none yet.)

### Fixed

- (none yet.)

---

## How to add an entry

- Add to `[Unreleased]` first.
- On release: rename `[Unreleased]` to `[X.Y.Z] - YYYY-MM-DD` and create a new `[Unreleased]`.
- Sections: Added / Changed / Security / Deprecated / Removed / Fixed.
- Cite ADR + doctrine rule where relevant.
- Breaking changes prefixed `BREAKING:` in commit message AND in this entry.
- Sunset notices for deprecations include the Sunset date per RFC 8594.
