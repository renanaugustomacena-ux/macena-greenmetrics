# Changelog

All notable changes to GreenMetrics will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Doctrine refs: Rule 21 (evolution + change management), Rule 34 (backend contracts), `docs/API-VERSIONING.md`.

## [Unreleased]

### Added

- **Sprint S5 — Charter, doctrine, plan, competitive brief.** Adopted `docs/MODULAR-TEMPLATE-CHARTER.md` (446 lines, reframes the project from SaaS to engagement template), `docs/DOCTRINE.md` (210 rules across 11 sections, 2893 lines, extends the existing 60-rule plan to cover modular-template integrity, audit-grade reproducibility, OT integration, regulatory packs, engagement lifecycle, cryptographic invariants, AI/ML reproducibility), `docs/PLAN.md` (1666 lines, six phases E–J across 18 sprints with explicit deliverables / exit criteria / doctrine traceability / ADR triggers / risk deltas), `docs/COMPETITIVE-BRIEF.md` (919 lines, surveys 30+ competitors across 9 tiers and identifies the 5 beat-points the template designs for). ADR-0021 records adoption.
- **Sprint S5 — Pack-loader skeleton.** New `backend/internal/packs/{pack.go, manifest.go, manifest_test.go}` implements the Pack contract per Charter §4 and Rules 69–88. Five PackKind values (protocol/factor/report/identity/region). Manifest schema with id/kind/version/min_core_version/pack_contract_version/author/license_spdx/capabilities/dependencies fields. ValidateBasic with SemVer + id-pattern + license + capabilities checks. Eight unit tests covering happy path + bad id + bad version + unknown kind + missing license + missing capabilities + SemVer pre-release + SemVer build metadata.
- **Sprint S5 — Pack-contract interfaces.** New `backend/internal/domain/{protocol,reporting,emissions,identity,region}/` packages each carrying the Pack-kind contract per Rule 32 (DDD) and ADR-0023. `protocol.Ingestor` (Start/Stop + ReadingSink), `reporting.Builder` (Type/Version/Build with pure-function semantics), `emissions.FactorSource` (Name/Refresh with temporal validity), `identity.IdentityProvider` (Authenticate/LookupUser supporting local-DB / SAML / OIDC dialects), `region.RegionProfile` (Code/Profile/HolidayCalendar/RegulatoryThreshold). Each package carries a `ContractVersion = "1.0.0"` constant per Rule 71 and an example test demonstrating compile-time + runtime contract compliance.
- **Sprint S5 — MODUS_OPERANDI v2 (engagement playbook).** `docs/MODUS_OPERANDI.md` rewritten in place per ADR-0028. Replaces the SaaS framing (per-meter pricing tiers €99/€249/Enterprise, CAC/LTV/churn unit economics, multi-tenant SaaS framing, ARR/ARPA targets) with the engagement model (license + customisation + annual maintenance + tier retainer; engagement-margin / time-to-customisation / template-fit-score / net-engagement-value / annual-maintenance-attach-rate KPIs; Phase 0–5 engagement lifecycle). Italian market analysis (CSRD wave-2 / Piano 5.0 / D.Lgs. 102/2014 tailwinds; sector composition; competitive positioning) preserved unchanged.
- **Sprint S5 — COST-MODEL v2 (engagement-margin model).** `docs/COST-MODEL.md` rewritten in place per Plan §5.3.3, sibling to ADR-0028. Replaces the v1 per-tenant unit economics (€/meter/month, 4×-cost gross-margin target, environment-totals AWS Budgets) with the per-engagement P&L (license + customisation + maintenance + tier-retainer + regulatory-deliverable revenue lines against engagement-lead time + advisory + per-deployment infra + Pack-update labour + regulatory-deliverable delivery + template-R&D amortisation cost lines), the per-deployment infra-cost model across Topologies A/B/C/D, the portfolio aggregate view (capacity utilisation + amortisation + Pack-contribution rate), per-deployment AWS Budgets, and engagement-level efficiency / waste-detection / anti-pattern coverage. Worked Standard T1 5-year P&L anchored on MODUS_OPERANDI v2 §11.3 (€440k revenue / €140k cost / €300k net / 68.2% margin). Doctrine refs: Rules 1, 13, 27, 38, 151, 152, 156, 168.
- **Sprint S5 — THREAT-MODEL rewrite (charter alignment).** `docs/THREAT-MODEL.md` §1 reframed (modular template / engagement / single-tenant by default replaces "multi-tenant SaaS"); new §2 *Tenancy model* per Charter §11 makes "single-tenant by default, multi-tenant by configuration" explicit and preserves all defence-in-depth tenant-isolation controls (RLS / RBAC / repo-level WHERE / `InTxAsTenant` / per-meter HMAC) as engineering hygiene; new §4.9 enumerates the Pack-loader threat surface (manifest validation, Cosign signature, manifest-lock per Rule 73, per-Pack failure isolation per Rule 76); subsequent sections renumbered §3–§8 from §2–§7. Cross-ref in `docs/TRUST-BOUNDARIES.md` updated (§3.3 → §4.3).

- **Sprint S5 — Pack manifest schema.** New `docs/contracts/pack-manifest.schema.json` (JSON Schema 2020-12) is the source of truth for the manifest format; the Go struct mirrors it.
- **Sprint S5 — `packs/` repository directory + README.** Documents the target Pack catalogue layout (12 protocol Packs, 6 factor Packs, 12 report Packs, 2 identity Packs, 6 region Packs) and the per-Pack contents convention.
- **Sprint S5 — `config/required-packs.yaml` skeleton.** Declarative list of Packs required for this deployment per Rule 75; Phase E Sprint S5 ships empty; Sprint S6+ populates with Italian-flagship Packs.
- **Sprint S5 — Universal-invariant conformance test stubs.** New `backend/tests/conformance/{cloudevents_test.go, log_format_test.go}` complete the Rule 1–8 universal-invariant test set (joining the existing health/money/rfc3339utc/rfc7807/uuidv4 stubs). Build tag `conformance`; tests scaffolded and skipped pending the integration fixture.
- **Sprint S5 — Taskfile pack helpers.** New `task pack:list` and `task pack:validate` targets for inspecting and validating Pack manifests against the schema.
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

- **`docs/adr/0001-platform-doctrine-adoption.md`** annotated with a 2026-04-30 supersession note pointing to ADR-0021. The ADR's "regulated-industry SaaS" framing is replaced; the operational decisions for Sprints S1–S4 remain in effect.
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

- **CI hotfix — `.pre-commit-config.yaml` rev SHAs.** Nine of fourteen pinned hook SHAs did not exist in their upstream repositories (gitleaks, markdownlint-cli, golangci-lint, dnephin/pre-commit-golang, mirrors-eslint, hadolint, pre-commit-terraform, kubeconform, conventional-pre-commit), causing every CI run to fail on `pre-commit run --all-files` at hook-environment-init time with `fatal: unable to read tree`. SHAs corrected to match each existing version label; gitleaks bumped from invalid v8.21.2 pin to valid v8.30.1 (`83d9cd684c87d95d656c1458ef04895a7f1cbd8e`). All 14 hook SHAs re-verified via `gh api repos/<owner>/<repo>/commits/<sha>` after the fix. Per Rule 53 (pin every dependency to a digest); Dependabot weekly bumps continue from the corrected baseline.

---

## How to add an entry

- Add to `[Unreleased]` first.
- On release: rename `[Unreleased]` to `[X.Y.Z] - YYYY-MM-DD` and create a new `[Unreleased]`.
- Sections: Added / Changed / Security / Deprecated / Removed / Fixed.
- Cite ADR + doctrine rule where relevant.
- Breaking changes prefixed `BREAKING:` in commit message AND in this entry.
- Sunset notices for deprecations include the Sunset date per RFC 8594.
