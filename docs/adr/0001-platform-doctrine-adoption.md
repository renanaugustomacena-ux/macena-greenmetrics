# 0001 — Platform doctrine adoption

**Status:** Accepted (superseded in part by ADR-0021 on 2026-04-30)
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** all 60 (Platform 9–28, Backend 29–48, DevSecOps 49–68)
**Review date:** 2026-10-25

> **2026-04-30 supersession note.** ADR-0021 supersedes the "regulated-industry SaaS" framing in this ADR's Tradeoff Stanza with "regulated-industry engagement template" framing per `docs/MODULAR-TEMPLATE-CHARTER.md` §2. The 60-rule doctrine adopted here is preserved verbatim and extended to 210 rules in `docs/DOCTRINE.md`. The operational plan referenced here covered Sprints S1–S4; the post-S4 plan is `docs/PLAN.md` (Phases E–J across Sprints S5–S22). All other clauses of this ADR (constraints, alternatives considered, consequences, residual risks, references) remain in effect.

## Context

GreenMetrics passed Mission II audit (verdict PASS, security CLEAN) on 2026-04-18 with the following accepted residual items: full integration test (G8) deferred, Scope 3 EEIO placeholder, Grafana dashboards thin, login validation Mission III placeholder. The product is functionally complete enough to demo and to defend in audit; it is not yet ready to operate at the regulated-industry quality threshold the doctrine requires.

The user (Renan) imposed a 60-rule doctrine spanning Platform Engineering (Rules 9–28), Advanced Backend Engineering (Rules 29–48), and DevSecOps Engineering (Rules 49–68). Several rules are aspirational headings already satisfied by the existing audit posture; many are operational gaps that the audit explicitly deferred. This ADR records the decision to adopt the entire doctrine and execute the gap-closing plan.

Constraints:

- CLAUDE.md cross-portfolio invariants (money cents+ISO-4217, RFC 3339 UTC, UUIDv4 tenant_id, RFC 7807 errors, CloudEvents 1.0, health envelope).
- Italian residency (eu-south-1 or Aruba), regulatory regimes (GDPR, NIS2 D.Lgs. 138/2024, CSRD/ESRS E1, Piano 5.0, D.Lgs. 102/2014).
- Distroless nonroot, PSS restricted, NetworkPolicy default-deny.
- Mission II accepted residual items must be honoured (RISK-001, RISK-002).
- Single operator today; structure must survive team growth without rewrite.

## Decision

Adopt the full 60-rule doctrine as the operational substrate for GreenMetrics. Execute the operational plan at `~/.claude/plans/my-brother-i-would-flickering-coral.md` over five 2-week sprints (10 weeks total) using the hybrid structure (pillar tracks → sprint sequencing).

The doctrine is binding. PRs will be reviewed against it. CI gates will enforce its mechanical parts (policy bundles, contract validation, supply-chain hygiene, observability discipline, conformance suite). Rejection authority (Rules 26, 46, 66) is explicit and recorded in `docs/adr/REJECTED.md`.

## Alternatives considered

- **Cherry-pick only the security-critical rules and defer the rest.** Rejected because the doctrine is interdependent: Rule 14 (contract first) underpins Rule 34 (validator) and Rule 54 (policy as code); Rule 33 (data is the system) underpins Rule 39 (RLS) and Rule 35 (idempotency). Cherry-picking creates incoherent gates.
- **Run a 4-week MVP doctrine adoption, defer the rest.** Rejected because 4 weeks is enough for foundation but not for verification (Rule 24, 44, 64 take a full sprint each to land properly).
- **Distribute the doctrine across 6 months at a slower cadence.** Rejected because the gap window (no IR doc, no policy gates, no supply-chain signing, no monitoring stack) is itself a risk — closing it slowly leaves regulator-grade exposure for longer.
- **Adopt only Platform + Backend doctrines; treat DevSecOps as separate later engagement.** Rejected because Rule 50 (DevSecOps as unified system) explicitly forbids this — security cannot be a post-step.

## Consequences

### Positive

- Single coherent operational plan covering all 60 rules.
- Every gap from Mission II audit closes with a named owner and sprint.
- Doctrine internalisation goal (Rules 28, 48, 68) becomes measurable: ADR cadence, runbook cadence, coverage gates, mutation kill rate, SLO compliance.
- Risk register (`docs/RISK-REGISTER.md`) becomes the source of truth for compliance evidence.
- Supply chain (Cosign keyless + SLSA L2 + Kyverno admission) closes the largest residual exposure.

### Negative

- 10-week execution window blocks net-new product features outside the doctrine deliverables.
- Documentation surface grows by ~50 new files; review burden increases.
- Pre-commit + policy gates add ~3–5 min to per-PR cycle.
- Argo CD adoption adds a new operational dependency (a control plane to operate).
- Redis added for Asynq job queue and distributed rate limit — a new failure domain.

### Neutral

- ADR cadence becomes a lifecycle norm; team learns the discipline.
- Cross-portfolio invariants are honoured but not driven from this plan; they were already in place.

## Residual risks

- The doctrine itself has not been audited by a third party — internal application only. RISK-006-style insider risks remain.
- Single operator today; bus factor is real. Mitigated by `docs/SECOPS-RUNBOOK.md` and structured handoff checklists.
- 10-week timeline is aggressive; some deliverables may slip. Sprint-end quality gates catch slips early.
- Mission III items (real IdP, Scope 3 EEIO mapping, multi-region active-active, mobile app for ingestion) are explicitly out of scope; the doctrine continues to apply as those features ship.

## References

- Doctrine plan: `/home/renan/.claude/plans/my-brother-i-would-flickering-coral.md` (entire file, 2657 lines).
- Mission II audit: `/media/renan/SSD Portable/STUDIO-LAVORO/macena-tools-lavoro/consolidation-reports/per-project/GreenMetrics-{consolidation,security,tests,italian-compliance}.md`.
- Cross-portfolio invariants: `/media/renan/SSD Portable/STUDIO-LAVORO/macena-tools-lavoro/CLAUDE.md`, `/media/renan/SSD Portable/STUDIO-LAVORO/macena-tools-lavoro/consolidation-reports/SHARED-SCHEMAS.md`.
- Risk register: `docs/RISK-REGISTER.md`.
- Threat model: `docs/THREAT-MODEL.md`.
- Charter: `docs/TEAM-CHARTER.md`, `docs/SECOPS-CHARTER.md`.
- RACI: `docs/RACI.md`.
- Layers: `docs/LAYERS.md`, `docs/layers.yaml`.
- Initiative workflow: `docs/PLATFORM-INITIATIVE-WORKFLOW.md`.

## Tradeoff Stanza

- **Solves:** the gap between Mission-II PASS and operate-grade for a regulated-industry SaaS.
- **Optimises for:** long-term operability, regulatory defensibility, doctrine internalisation by the team.
- **Sacrifices:** 10 weeks of net-new product feature velocity; ~50 new files of documentation surface; pre-commit + policy-gate latency on every PR; Redis as new failure domain.
- **Residual risks:** doctrine not externally audited; single operator bus factor; 10-week timeline aggressive — slip risk is real and must be managed sprint by sprint.
