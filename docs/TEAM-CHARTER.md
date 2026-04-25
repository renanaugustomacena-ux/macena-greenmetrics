# GreenMetrics Team Charter

**Authoring date:** 2026-04-25
**Owners:** `@greenmetrics/platform-team`, `@ciupsciups`
**Review cadence:** quarterly (next review: 2026-07-25)
**Doctrine refs:** Rule 9 (Role: Platform Engineering Authority).

This charter defines the platform-engineer role on GreenMetrics, its mandate, decision rights, and escalation paths. It is the source of truth for "who decides what" on the platform substrate.

## 1. Mandate

The platform-engineer role exists to enable other software (Rule 12) by providing:

- Stable contracts at every layer crossing (Rule 14).
- Opinionated defaults and controlled extension points (Rule 12).
- Reliable, observable, secure substrate that the application team can build on without re-deriving operational truths each release (Rules 15, 18, 19, 20).
- Cost-aware, scalable, evolvable infrastructure (Rules 16, 21, 22).

The platform team is **not** a service provider for individual feature requests. It is the substrate that other teams ship through.

## 2. Scope

In scope:

- Backend service contracts (`api/openapi/v1.yaml`, JSON Schemas).
- Database schema lifecycle (migrations, RLS, retention, CAGGs).
- Application boundaries (handlers / services / repository / domain).
- Kubernetes manifests, Argo CD applications, GitOps tree.
- Terraform modules and IaC contracts.
- CI/CD pipelines and policy gates.
- Observability stack (Prometheus, Alertmanager, Grafana, Loki, Tempo, OTel collector).
- Operational runbooks, ADRs, on-call rota.
- Developer experience (devcontainer, bootstrap script, troubleshooting).
- Cost model and budget enforcement.
- Cross-cutting platform abstractions (resilience, idempotency, validation, observability helpers).

Out of scope:

- Product roadmap (owned by product).
- Feature design beyond security/operational inputs (owned by app team).
- Customer-success and account management (owned by go-to-market).
- Regulatory text authoring (owned by legal / compliance, with secops review).

## 3. Decision rights

The platform team has explicit rejection authority (Rules 26, 46, 66) over:

- Architectural changes that violate the doctrine.
- New abstractions whose cost > leverage (Rule 13).
- Tool sprawl or "platform-by-committee" designs (Rule 26).
- Dependency adoption without supply-chain verification (Rule 53).
- Manual gates that should be automated (Rule 56).
- Implicit contracts (Rule 14).

Rejections must be explicit and justified, recorded in `docs/adr/REJECTED.md`, and revisited quarterly.

## 4. Escalation

| Concern | First escalation | Second escalation |
|---|---|---|
| Security vulnerability (CRITICAL) | `@greenmetrics/secops` | repo owner |
| Production incident SEV1 | on-call primary (PagerDuty) | on-call secondary, then engineering manager |
| Architectural disagreement | platform-team review meeting | quarterly platform office hours |
| Regulatory / compliance question | `@greenmetrics/secops` + `@greenmetrics/legal` | DPO |
| Cost anomaly > budget | `@greenmetrics/platform-team` | engineering manager |
| Vendor / supply-chain concern | `@greenmetrics/secops` | repo owner |

## 5. Authority limits

The platform team **may not**:

- Bypass `secops` review on security-relevant changes (Rule 49).
- Skip ADR documentation for non-trivial decisions (Rule 27, 47, 67).
- Force-push to `main` or rewrite published history.
- Disable CI gates without an override label tied to an ADR (Rule 23 guardrails).
- Introduce production changes outside the GitOps reconciliation loop (Rule 63).
- Authorise destructive Terraform applies without 4-eyes review.

## 6. 90-day platform health review

The platform team self-reviews against the doctrine quarterly:

- Are the Rule 11 / 31 / 51 sequences being followed without prompting?
- Are tradeoffs (Rule 27 / 47 / 67) named in every non-trivial PR?
- Are anti-patterns (§3 of the plan) being rejected with citation?
- Is the on-call rota healthy (incident count, MTTR trend, runbook freshness)?
- Is the cost model tracking actual spend within ±5%?
- Is the test pyramid healthy (coverage, mutation kill rate, flaky-test rate)?
- Has the threat model been refreshed since last quarter?

Outcomes captured in `docs/office-hours/YYYY-MM-DD.md`.

## 7. Termination criterion

The charter has done its job when (Rule 28):

- New engineers can apply the doctrine without external prompting.
- ADR cadence ≥ 3/quarter is sustained.
- Rejection log is being maintained without external review.
- Quarterly office hours run without assistant attendance.

At that point the assistant moves to consultation-only mode. The team owns the substrate.
