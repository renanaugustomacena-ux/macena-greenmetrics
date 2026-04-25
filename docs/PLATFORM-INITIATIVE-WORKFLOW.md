# Platform Initiative Workflow

**Doctrine refs:** Rule 11 (methodological sequence — Platform), Rule 31 (sequence — Backend), Rule 51 (sequence — DevSecOps).

Every new platform feature follows this sequence. Skipping a step is professional malpractice (Rule 11 verbatim).

A "platform feature" is any change in `/k8s`, `/gitops`, `/terraform`, `/backend/migrations`, `/api/openapi/`, `/.github/workflows/`, `/policies/`, `/monitoring/`, `/grafana/`, or any new `internal/` Go package.

## 1. Sequence

### Step 1 — Purpose

State the problem the initiative solves in one sentence. Cite the issue or risk it addresses (link to `docs/RISK-REGISTER.md` if applicable).

### Step 2 — Stakeholders

List who is impacted — owner team(s), consumer team(s), operator(s), regulator (if applicable). Cross-reference `docs/RACI.md`.

### Step 3 — Constraints

Enumerate the constraints. At minimum:

- CLAUDE.md cross-portfolio invariants (money cents+ISO-4217, RFC 3339 UTC, UUIDv4 tenant_id, RFC 7807 errors, CloudEvents 1.0, health envelope).
- Italian residency (eu-south-1 or Aruba).
- Distroless nonroot, PSS restricted.
- Cost budget (per `docs/COST-MODEL.md`).
- SLO budget (per `docs/SLO.md`).
- Rule 11 / 31 / 51 sequence application.

### Step 4 — Decomposition

Apply Rule 10 (multi-layer system). Identify which of the five layers is touched. State whether the change is single-layer (rare, suspect) or cross-layer (common). Use `docs/LAYERS.md` as the framework.

For backend: apply Rule 31 spine — domain → constraints → data → services → failure → contracts → validation. State which aggregate(s) the change touches in `internal/domain/`.

For devsecops: apply Rule 51 — threat → SDLC → trust boundaries → controls → automation → incident → validation.

### Step 5 — Contracts

For every layer crossing introduced, define the contract:

- API: OpenAPI 3.1 entry in `api/openapi/v1.yaml`.
- Config: JSON Schema entry in `docs/contracts/config.schema.json`.
- Event: CloudEvents 1.0 schema in `docs/contracts/events/`.
- CLI: target in `Taskfile.yaml` + `docs/CLI-CONTRACT.md`.
- IaC: Terraform module README via `terraform-docs`.
- Operational: runbook YAML front-matter in `docs/runbooks/`.
- Policy: Rego rule in `policies/conftest/` referencing `RISK-NNN`.

### Step 6 — Failure / Scaling / Evolution

- **Failure:** what fails, what is the blast radius, what is the secure-degradation path? Add entry to `docs/RELIABILITY-MODEL.md` if new failure mode.
- **Scaling:** which axis (Rule 16) — traffic, data, teams, features, ops load? What is the ceiling?
- **Evolution:** API versioning policy (Rule 21) applies. State `Sunset` plan if deprecating. Update `CHANGELOG.md`.

### Step 7 — Validation

State how the change will be tested:

- Unit (`internal/domain/` ≥ 90% coverage).
- Integration (testcontainers).
- E2E (Playwright if user-facing).
- Property (gopter for invariants).
- Conformance (RFC 7807, RFC 3339 UTC, etc.).
- Load (k6) if perf-touching.
- Chaos (Chaos Mesh) if reliability-touching.
- DAST (nuclei) if attack-surface-introducing.
- Pentest (manual) if regulator-touching.

State the SLO budget impact in numbers.

## 2. ADR requirement

The seven steps produce an ADR in `docs/adr/NNNN-<slug>.md`. The ADR ends with the four-part Tradeoff Stanza (Rule 27, 47, 67):

- **Solves:** ...
- **Optimises for:** ...
- **Sacrifices:** ...
- **Residual risks:** ...

ADRs without all four parts are blocked by `markdownlint` rule.

## 3. PR requirement

The PR template (`/.github/PULL_REQUEST_TEMPLATE.md`) requires:

- Linked issue + ADR.
- Doctrine rules touched (cite by number).
- Doctrine rules rejected (cite override ADR if any).
- Risk acknowledged stanza.
- Backend addendum (if `/backend/` touched): domain modelling notes, consistency model, failure mode plan, latency budget, trust boundary touched, doctrine checklist signed.
- Runbook update.
- CHANGELOG entry (if `/api/v1/` touched).
- Verification checklist completed.

## 4. Review requirement

CODEOWNERS routes the PR. Required reviewers per the path table in `.github/CODEOWNERS`. Reviewer applies the doctrine checklist from `docs/backend/doctrine-checklist.md` (for backend) or this document (for platform).

Required CI checks (branch protection): pre-commit-ci, backend-lint, backend-test, backend-integration, backend-property, backend-security, backend-conformance, backend-static, backend-build, frontend-lint, frontend-test, frontend-build, frontend-e2e, docker, trivy-fs, sast (CodeQL), semgrep, gitleaks, govulncheck, osv, license, kubeconform, policy-gate-k8s, policy-gate-dockerfile, policy-gate-terraform, openapi-lint, openapi-compat, config-schema, event-schema, adr-link-check, doctrine-checklist, sbom, coverage.

## 5. Anti-patterns

- "Just merge it, ADR later" — rejected. ADR is the design, not the documentation.
- "Single-layer optimisation" (Rule 10) — every change touches multiple layers; reason about all of them.
- "Implicit contract" (Rule 14) — undocumented seams are debt; document at design time.
- "Manual approval as the only gate" (Rule 56) — automate or make explicit.
- "We'll add tests later" (Rule 24, 44) — tests first; the PR template requires test plan.

## 6. Quarterly review

Office hours review randomly samples a few merged PRs against this workflow. Failure to apply the sequence escalates to platform office hours topic.
