# GreenMetrics SecOps Charter

**Authoring date:** 2026-04-25
**Owners:** `@greenmetrics/secops` (`@ciupsciups`).
**Review cadence:** quarterly (next: 2026-07-25).
**Doctrine refs:** Rules 49 (role), 50 (unified system), 51 (sequence), 65 (regulated quality), 66 (rejection authority), 67 (transparency), 68 (termination).

## 1. Mandate

The SecOps role on GreenMetrics is one role with two equally weighted halves: **secure software supply chain** and **continuous secure delivery**. Both halves operate as a single system (Rule 50): code → build → artefact → deploy → run, with security signals at each stage and feedback loops back into the backlog.

## 2. Scope

In scope:

- Software supply chain (source integrity, dependency provenance, build reproducibility, artefact immutability, deployment trust). See `docs/SUPPLY-CHAIN.md`.
- CI/CD security gates (`.github/workflows/{ci,cd,sast,supply-chain,dast}.yml`).
- Policy as code (`policies/conftest/`, `policies/kyverno/`, `policies/falco/`).
- Identity and secrets (ESO, JWT KID rotation, IRSA, GitHub OIDC, mTLS plan).
- Threat modelling (`docs/THREAT-MODEL.md`) and risk register (`docs/RISK-REGISTER.md`).
- Incident response (`docs/INCIDENT-RESPONSE.md`) and forensic readiness.
- Compliance mapping (NIS2 `D.Lgs. 138/2024`, GDPR, CSRD/ESRS E1, Piano Transizione 5.0, D.Lgs. 102/2014).
- Penetration testing engagements and DAST cadence.
- Audit log integrity and retention.
- Container runtime detection (Falco / Tetragon).
- Network policy and egress allowlist.

Out of scope:

- Product roadmap.
- Feature design beyond security/regulatory inputs.
- Customer success.
- Operational toil unrelated to security (handed off to SRE / platform).

## 3. Authority

SecOps has explicit rejection authority (Rule 66) over:

- Security theatre (controls without backing risk).
- Manual approval as the only gate (must be replaced by automated layered gates).
- Tool sprawl (e.g. Snyk on top of Trivy + govulncheck + osv-scanner).
- "Security slows us down" assumptions (countered by sub-5s pre-commit, canary auto-rollback).
- Blind vendor trust (replaced by SBOM + SLSA + Cosign verification at admission).
- DIY KMS or DIY identity systems.
- Unpinned GitHub Actions tags.
- Cosign with custom keys when keyless OIDC works.

Each rejection is documented in `docs/adr/REJECTED.md` with rationale, alternative, residual risk, and review date (Rule 67).

## 4. Decision sequencing (Rule 51)

Every SecOps initiative follows the seven-step sequence:

1. Threat / risk model — `docs/THREAT-MODEL.md`, `docs/RISK-REGISTER.md`.
2. SDLC mapping — `docs/PIPELINE-MAP.md`.
3. Trust boundaries — `docs/TRUST-BOUNDARIES.md`.
4. Controls / policy definition — `policies/`.
5. Automation / enforcement — CI gates, GitOps, Argo CD.
6. Failure / incident model — `docs/INCIDENT-RESPONSE.md`.
7. Validation — DAST, chaos, pentest, continuous verification (`make verify`).

Skipping a step is professional malpractice (Rule 51 verbatim).

## 5. Cadence

| Cadence | Activity |
|---|---|
| Per PR | Pre-commit, conftest, Cosign verify, Trivy image, govulncheck, osv-scanner, license check, gitleaks, semgrep, CodeQL. |
| Daily | Alertmanager triage; Falco event review. |
| Weekly | DAST sweep (Saturday 03:00 UTC); Dependabot triage; SLO synthetic-probe review. |
| Monthly | Chaos Game Day (last Friday); cost audit; risk register HIGH/CRITICAL review; runbook freshness. |
| Quarterly | External pentest cadence check (annual external + quarterly internal); JWT secret rotation; threat-model review; ADR audit; DevSecOps review meeting. |
| Annually | DR full drill; regulatory citation re-verification; license-allowlist refresh; Cosign root key trust review. |

## 6. Bus-factor mitigation

The charter assumes a 2-person `secops` team. Today the team is one operator (`@ciupsciups`). Mitigations:

- Every secops procedure is a runbook in `docs/runbooks/` with YAML front-matter `last_tested` field.
- Every non-trivial decision is an ADR in `docs/decisions/`.
- Every policy rule references its `RISK-NNN` mitigation target.
- `docs/SECOPS-RUNBOOK.md` documents handoff checklist for when a second operator joins.
- Forensic-readiness requirements (Object Lock, audit log retention 365d, snapshot policy) are infrastructure controls, not operator-knowledge controls.

## 7. Termination criterion (Rule 68)

The charter has done its job when:

- The team thinks security-first by default — every PR carries threat-model annotation when relevant.
- The risk register is treated as a live artefact, not a checkbox doc.
- Quarterly DevSecOps review runs without external prompting.
- "What is your control list" can be answered by exporting `docs/RISK-REGISTER.md` + `docs/SECOPS-CHARTER.md` + the policy bundles.
- Incident response drills complete in <30 min using only the documented playbook.

At that point assistant attendance becomes consultation-only.
