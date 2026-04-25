# GreenMetrics Platform Playbook

**Owner:** `@greenmetrics/platform-team`.
**Doctrine refs:** Rule 28 (Platform termination objective), Rule 48 (Backend), Rule 68 (DevSecOps).
**Audience:** every engineer shipping into GreenMetrics.

The plan at `~/.claude/plans/my-brother-i-would-flickering-coral.md` is the *what*. This playbook is the *how* — the working internalisation of the 60-rule doctrine for the team that ships every day.

## 1. The five layers (Rule 10)

| Layer | What lives here | Owner |
|---|---|---|
| Infra | AWS (VPC, EKS, RDS Timescale, S3, KMS, Secrets Manager, CloudFront+WAF) | platform |
| Platform | K8s, Argo CD, ESO, OTel collector, Prometheus, Loki, Tempo, Falco, cert-manager, Kyverno | platform / secops / sre |
| App | Backend (Fiber + pgx), worker (Asynq), frontend (SvelteKit), simulator | app |
| Tooling | CI/CD, devcontainer, pre-commit, conftest, Cosign, Renovate | platform / secops |
| Operators | Runbooks, ADRs, on-call, postmortems, office hours | sre / secops |

Every change reasons across multiple layers. If you only consider one layer, you missed something.

## 2. The Rule 11 / 31 / 51 sequence

For any new platform / backend / secops feature:

1. **Purpose** — what problem? linked issue or RISK-NNN?
2. **Stakeholders** — who's impacted?
3. **Constraints** — CLAUDE.md invariants, regulatory, cost, SLO.
4. **Decompose** — which layers? backend spine = domain → constraints → data → services → failure → contracts → validation. devsecops spine = threat → SDLC → trust boundaries → controls → automation → incident → validation.
5. **Contract** — OpenAPI, JSON Schema, CloudEvents, runbook front-matter, Rego policy.
6. **Failure / Scaling / Evolution** — what fails, blast radius, secure degradation; scaling axis; SemVer + Sunset.
7. **Validate** — unit + integration + e2e + property + conformance + load + chaos + DAST.

End with an ADR (`docs/adr/`). Tradeoff Stanza required: **Solves / Optimises for / Sacrifices / Residual risks**.

## 3. The non-negotiables (Rule 25 / 45 / 65)

Cross-portfolio invariants:

- Money = `(amount_cents int64, currency ISO-4217 string)`. Never float.
- Timestamp = RFC 3339 UTC with offset.
- `tenant_id` = UUIDv4.
- Errors = RFC 7807.
- Events = CloudEvents 1.0.
- Health = `{status, service, version, uptime_seconds, time, dependencies}`.

Every PR must:

- Pass pre-commit (locally + CI mirror).
- Carry a populated PR template.
- Include the four-part Tradeoff Stanza when introducing a non-trivial decision.
- Pass the doctrine checklist (`docs/backend/doctrine-checklist.md`).
- Verify CI gates green: lint, test, integration, build, security (Trivy + govulncheck + osv + gitleaks + semgrep + CodeQL + license), policy gate (conftest + kubeconform + Kyverno), SBOM, OpenAPI lint + compat.

## 4. The Rule 26 / 46 / 66 rejection authority

When you see one of the 35 anti-patterns from `docs/adr/REJECTED.md` proposed:

- Cite the rejection ID (`REJ-NN`).
- Point at the alternative.
- Allow override only with explicit unrejection ADR + secops/platform sign-off.

Common ones to watch:

- Service mesh for ≤6 services.
- Manual approval as the only CD gate.
- Snyk / Aqua / Sysdig on top of CNCF stack.
- DIY KMS / DIY identity.
- Unpinned GitHub Actions tags.
- swag-only API contract.
- Float for money.
- `panic()` on user input.
- Implicit tenant scoping (must use `InTxAsTenant` + RLS).
- Synchronous report generation.

## 5. Daily cadence

- Pre-commit catches the cheap stuff.
- CI mirror catches the bypassers.
- Reviewers apply the doctrine checklist.
- Runbooks are dry-run-tested quarterly.
- Risk register is reviewed monthly for HIGH/CRITICAL changes.
- Threat model is reviewed quarterly + on every dependency major-version bump.
- ADR cadence: ≥ 3/quarter sustained.

## 6. Quarterly cadence

- **Platform office hours** — sample PRs, walk doctrine application, collect questions for FAQ.
- **DevSecOps review** — risk register + chaos log + pentest report + incident postmortems.
- **Architectural review** — engineers present recent changes against the 20-rule grid.
- **JWT rotation** (cron `0 3 1 */3 *`).
- **Internal pentest** (rotating scope per `docs/PENTEST-CADENCE.md`).
- **ADR audit** — review-date-stale ADRs revisited.

## 7. Annual cadence

- DR full drill (region failover).
- Regulatory citation re-verification (`docs/ITALIAN-COMPLIANCE.md` against primary sources).
- Compliance evidence pack refresh (`docs/COMPLIANCE/`).
- License-allowlist refresh (`LICENSES.allowed`).
- Cosign root key trust review.
- External pentest (Q3).

## 8. The termination objective (Rules 28, 48, 68)

You have internalised the doctrine when:

- You apply the Rule 11 / 31 / 51 sequence on new initiatives without prompting.
- You name tradeoffs in PR descriptions without being asked.
- You reject anti-patterns with a citation, not a vibe.
- You can articulate, for any new endpoint, its consistency model + failure budget + latency target + isolation strategy + trust boundary.
- You can answer "what is the threat model for the OCPP path?" / "what control mitigates RISK-005?" / "where does Falco ship its events?" — without prompting.
- The team produces ≥ 3 ADRs and ≥ 1 runbook per quarter without external coaching.
- The quarterly office hours run with engineer-led presentations; the assistant is there as observer only.

When all of the above is sustained over two consecutive quarters, the assistant moves to consultation-only mode. The team owns the substrate.

That's success.

## 9. Where to start (new engineer day 1)

Read in this order:

1. `docs/CONTRIBUTING.md`
2. `docs/TEAM-CHARTER.md`
3. `docs/RACI.md`
4. `docs/LAYERS.md`
5. `docs/QUALITY-BAR.md`
6. `docs/THREAT-MODEL.md`
7. `docs/RISK-REGISTER.md`
8. `docs/SUPPLY-CHAIN.md`
9. This file.
10. `~/.claude/plans/my-brother-i-would-flickering-coral.md` (reference).

Then run `./scripts/dev/bootstrap.sh` and ship your first PR within 3 days.
