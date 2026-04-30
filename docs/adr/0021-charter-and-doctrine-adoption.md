# 0021 — Charter and doctrine adoption

**Status:** Accepted
**Date:** 2026-04-30
**Authors:** @ciupsciups
**Supersedes:** ADR-0001 §"Tradeoff Stanza" — replaces "regulated-industry SaaS" framing with "regulated-industry engagement template" framing per `docs/MODULAR-TEMPLATE-CHARTER.md` §2.
**Doctrine refs:** all 210 rules of `docs/DOCTRINE.md`.
**Review date:** 2026-10-30

## Context

GreenMetrics passed Mission II audit (PASS, security CLEAN) on 2026-04-18. ADR-0001 adopted the implicit 60-rule doctrine (Platform 9–28, Backend 29–48, DevSecOps 49–68) and the operational plan at `~/.claude/plans/my-brother-i-would-flickering-coral.md`. That ADR explicitly framed the project as "a regulated-industry SaaS" in its Tradeoff Stanza.

Between 2026-04-25 and 2026-04-30 the project completed Sprints S1–S4 against that plan and surfaced two structural realities that the SaaS framing did not anticipate:

1. **Single-operator business.** GreenMetrics is operated as a single-engineer template-and-engagement business by Renan Augusto Macena, not as a multi-team SaaS company. The buyer is the engagement client, not the long-tail SME via self-serve. The operational obligations of a multi-tenant SaaS (Stripe billing, churn dashboards, CAC/LTV measurement, signup flows, customer-support ticket queues) do not exist and are not on the roadmap.

2. **Code reuse is per-engagement, not per-tenant.** The Italian flagship code in `internal/services/` is the pilot for a Pack-extraction model where each new client gets a per-engagement fork plus Pack composition. The "tenant" in the codebase is a defence-in-depth construct, not a billing primitive.

Continuing under the SaaS framing forces vocabulary and governance that don't match the actual delivery model and introduces friction at every PR review. The pivot to a *modular template + engagement* model is recorded in `docs/MODULAR-TEMPLATE-CHARTER.md` (adopted same date) and operationalised in `docs/PLAN.md` (adopted same date). The doctrine grows from 60 rules to 210 rules (`docs/DOCTRINE.md`), preserving the existing rule numbers 9–68 verbatim and adding seven new groups (69–88 Modular Template Integrity, 89–108 Audit-Grade Reproducibility, 109–128 OT Integration Discipline, 129–148 Regulatory Pack Discipline, 149–168 Engagement Lifecycle, 169–188 Cryptographic Invariants, 189–208 AI/ML Reproducibility, 209–210 Meta-rules).

Constraints unchanged from ADR-0001: CLAUDE.md cross-portfolio invariants; Italian residency (eu-south-1 or Aruba); regulatory regimes (GDPR, NIS2 D.Lgs. 138/2024, CSRD/ESRS E1, Piano 5.0, D.Lgs. 102/2014); distroless nonroot; PSS restricted; NetworkPolicy default-deny; Mission II accepted residual items honoured.

## Decision

Adopt `docs/MODULAR-TEMPLATE-CHARTER.md`, `docs/DOCTRINE.md` (210 rules), and `docs/PLAN.md` (Phase E–J across Sprints S5–S22) as the binding governance for GreenMetrics from Sprint S5 forward. PR review references the charter and doctrine; CI gates enforce the mechanically-checkable subset; rejection authority remains as in Rules 26 / 46 / 66 plus the new Rule 87 Pack acceptance authority.

The "regulated-industry SaaS" framing in ADR-0001 is replaced by "regulated-industry engagement template" framing per Charter §2. ADR-0001 is otherwise preserved — its operational plan still applies for the historical Sprints S1–S4, and its choice to adopt the doctrine is reaffirmed at greater scope. ADR-0001 receives an inline annotation pointing to this ADR.

## Alternatives considered

- **Continue under ADR-0001 SaaS framing and treat the charter as marketing.** Rejected because the SaaS framing produces ongoing friction in PR reviews, in document writing, in commercial conversations. Folklore versus doctrine is exactly the failure mode Rule 209 / 210 exists to prevent.
- **Retire ADR-0001 entirely and replace with a fresh ADR.** Rejected because ADR-0001's operational plan covered Sprints S1–S4 and is part of the historical record. Replacement would orphan the audit lineage and the conformance evidence already filed. Annotation + supersession is cheaper.
- **Defer the framing change to v1.0.0 launch.** Rejected because the friction is current; deferring it for ~9 months pushes ~9 months of doc-rewrite-as-side-quest into every PR. The pivot is cheaper to execute now while Phase E is starting, when the rewrite cost can be amortised across the Pack-extraction work.
- **Run the engagement model as an internal experiment without a charter.** Rejected because Rule 9 (platform discipline) and Rule 28 (termination criterion) require a documented charter.

## Consequences

### Positive

- Single coherent governance set (charter + doctrine + plan).
- Pack-extraction work in Sprint S6–S7 has a clear architectural basis.
- The 140 new doctrine rules cover audit-grade reproducibility, OT integration discipline, regulatory pack discipline, engagement lifecycle, cryptographic invariants, and AI/ML reproducibility — each addressing a real gap surfaced by Mission II.
- The competitive moat surfaces in `docs/COMPETITIVE-BRIEF.md` are now backed by doctrine rules a reviewer can audit.
- The risk register (`docs/RISK-REGISTER.md`) gains seven new RISK entries (RISK-024 through RISK-030) with named mitigations.

### Negative

- ~3000 LoC of doctrine-and-plan documentation surface to maintain.
- Quarterly office hours cadence becomes a real obligation; missing it weakens the doctrine.
- Vocabulary mass-rewrites across MODUS_OPERANDI, COST-MODEL, THREAT-MODEL §2, GDPR / NIS2 / CSRD compliance docs in Phase E Sprint S5–S6.
- Charter-supersession in ADR-0001 produces a non-trivial cross-reference graph that future maintainers must navigate.

### Neutral

- Charter, doctrine, and plan all carry the same six-month review cadence — review burden is bounded.
- Rejection authority structure unchanged; just gains the Rule 87 Pack-acceptance authority.

## Residual risks

- The doctrine has not been audited by a third party — internal application only. RISK-006-style insider risks remain.
- Single-operator bus factor (RISK-001) is reaffirmed; mitigation is the engagement-team hire planned for Phase H.
- If the Pack-extraction work in Phase E Sprint S6–S7 produces regressions, the engagement-fork model that Phase E ships gets delivered with bugs.
- The 36-week plan has aggressive scope; slippage beyond Phase J Sprint S22 requires a re-plan ADR.

## References

- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md`.
- Doctrine: `docs/DOCTRINE.md`.
- Plan: `docs/PLAN.md`.
- Competitive brief: `docs/COMPETITIVE-BRIEF.md`.
- Predecessor: `docs/adr/0001-platform-doctrine-adoption.md` (annotated to point here).
- Sprint S1–S4 deliverables: commits `ac0cb63`, `3adb466`, `a990c23`, `db432cd`, `303c74d`.

## Tradeoff Stanza

- **Solves:** the gap between the SaaS framing of ADR-0001 and the actual delivery model of an engagement template; the absence of an explicit Modular Template Integrity / Audit-Grade Reproducibility / OT Integration / Regulatory Pack / Engagement Lifecycle / Cryptographic / AI-ML rule group; the lack of a doctrine-derived plan from current state to v1.0.0 charter-conformant; the lack of a competitive brief positioning the template against the surveyed field.
- **Optimises for:** doctrine continuity (preserve rule numbers 9–68), governance coherence (one charter + one doctrine + one plan), engagement readiness (the lifecycle and tier model are explicit), audit-grade defensibility (every reproducibility property is a rule with a conformance test), competitive moat clarity (every beat-point is doctrine-backed).
- **Sacrifices:** ~3000 LoC of new documentation surface to maintain; ~2 sprints of Phase E velocity dedicated to doc rewrites; the optional simplicity of "we have a 60-rule doctrine"; the optionality to silently amend rules between office hours.
- **Residual risks:** doctrine drift if office hours are skipped (mitigated by Rule 209/210 rotation process); the 36-week plan slips (mitigated by 20% reserved capacity + sprint exit gates); the engagement-team hire in Phase H is delayed (mitigated by deferring T2/T3 expansion); the v1.0.0 license decision deferred past Phase J Sprint S22 (mitigated by ADR-0053 placeholder).
