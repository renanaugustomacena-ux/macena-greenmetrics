# 0028 — MODUS_OPERANDI v2 — engagement playbook

**Status:** Accepted
**Date:** 2026-04-30
**Authors:** @ciupsciups
**Supersedes:** in-place rewrite of `docs/MODUS_OPERANDI.md` v1 (the SaaS playbook).
**Doctrine refs:** Charter §2 (the pivot), Rules 149–168 (Engagement Lifecycle), Rule 209 (doctrine evolution).
**Plan ref:** `docs/PLAN.md` §5.3 (Sprint S5 deliverable §5.3.2 — MODUS_OPERANDI v2).
**Review date:** 2026-10-30

## Context

`docs/MODUS_OPERANDI.md` v1 is a ~14k-word SaaS playbook covering market analysis, business model with per-meter pricing tiers (€99 Starter / €249 Professionale / Enterprise), CAC/LTV/churn unit economics, customer-success motion, and Series-A growth thesis. It was written before the modular-template pivot recorded in `docs/MODULAR-TEMPLATE-CHARTER.md`.

The v1 framing is incompatible with the post-pivot delivery model. Concretely:

- The v1 text repeatedly describes GreenMetrics as "a multi-tenant SaaS platform." The post-pivot reality is a modular template delivered as engagement.
- The v1 pricing tiers (€99 / €249 / Enterprise) are per-meter monthly charges. The post-pivot revenue model is engagement license + customisation services + annual maintenance + tier retainer.
- The v1 unit economics (CAC, LTV, churn, ARPA, LTV/CAC ratio) are SaaS metrics. The post-pivot KPIs are engagement margin, time-to-customisation, template-fit-score, net-engagement-value, annual-maintenance-attach-rate.
- The v1 GTM lean is "Verona-direct + partner channel + 40% channel-revenue-mix by Y3" sized against a SaaS SOM. The post-pivot motion is engagement-by-engagement with channel partners (ESCO + EGE + commercialisti + system-integrator) on different mechanics.
- The v1 success target is "€5.4M ARR + €2.0M services = €7.4M revenue by Y3". The post-pivot success target is "5 engagements through Phase 5 Handover by end of Phase J" (Plan §3.2).

Continuing under v1 produces friction in every commercial conversation, every doc-rewrite, every customer-onboarding artefact. The pivot is recorded in the charter and the doctrine; the MODUS_OPERANDI must align.

## Decision

Rewrite `docs/MODUS_OPERANDI.md` in place as v2 — the engagement playbook. The rewrite preserves the Italian-market analysis (it remains the flagship-region commercial context) and the technical-architecture overview (it remains accurate post Pack extraction in Sprint S6) but replaces:

- the multi-tenant SaaS framing → modular template + engagement framing
- per-meter pricing tiers → engagement license + customisation services + annual maintenance + tier retainer
- CAC/LTV/churn → engagement margin + time-to-customisation + template-fit-score + net-engagement-value + maintenance-attach-rate
- ARR/ARPA targets → engagements-through-Phase-5 + annual-maintenance-attach-rate
- "Series A growth on EU-sustainability theses" → "engagement-portfolio scaling with measured engagement-margin discipline"

The Italian-flagship + EU-expansion arc is preserved with the framing that flagship Italy is the reference (Rule 88) and other Region Packs (DE, ES, FR, GB, AT) ship as engagement demand opens those geographies (Plan Phase H Sprint S17).

## Alternatives considered

- **Mark v1 deprecated; ship v2 as a separate file.** Rejected because two documents create version drift; the canonical document is the engagement playbook.
- **Leave v1 in place; add a v2 addendum at the top.** Rejected because half the words become wrong-framed; readers would have to guess which paragraphs are still-current.
- **Defer the rewrite to v1.0.0 launch.** Rejected because every commercial conversation between now and Phase J would have to manually translate. The cost of deferring is paid 50+ times.
- **Strip MODUS_OPERANDI entirely; put the commercial playbook in CHARTER.** Rejected because the charter is engineering-focused; the commercial playbook is its own document with a different audience (the sales-leadership reader, future investor, future engagement-team hire).

## Consequences

### Positive

- One coherent commercial framing across charter, doctrine, plan, MODUS_OPERANDI, COST-MODEL, THREAT-MODEL.
- New engagement-team hires (Phase H) onboard from a document that matches reality.
- Future commercial conversations (engagement prospects, channel partners, EGE relationships) reference one source.
- The Italian-flagship + EU-expansion arc gets a clear engagement-by-engagement description rather than a SaaS-cohort description.

### Negative

- ~14k words of in-place rewrite. ~1500–2000 lines of new content.
- Historical readers who memorised the v1 pricing tiers must re-learn.
- Vendor / channel partners who read v1 may need a quick re-orientation conversation.

### Neutral

- The Italian market analysis (TAM, SAM, SOM, regulatory drivers) is preserved unchanged — those numbers are about the underlying market, not about our delivery model.
- The technical-architecture overview is preserved with light annotation pointing at the Pack extraction.

## Residual risks

- Some commercial / regulatory partners have already read v1 and may continue to expect SaaS-style deliverables. Mitigated by an explicit "what changed since 2026-04" section in the v2 introduction.
- The COST-MODEL.md rewrite (Plan §5.3 Sprint S5 deliverable #3) must land in sync; otherwise COST-MODEL contradicts MODUS_OPERANDI v2 on engagement margin vs SaaS gross margin.
- Tax / accounting framings around services revenue vs subscription revenue change between v1 and v2; the Italian commercialista must update their model.

## References

- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §2 (the pivot), §8 (engagement lifecycle), §9 (economic model).
- Doctrine: `docs/DOCTRINE.md` Rules 149–168 (Engagement Lifecycle), Rule 88 (Italian Pack flagship).
- Plan: `docs/PLAN.md` §5.3 (Sprint S5 deliverable list).
- Predecessor: `docs/MODUS_OPERANDI.md` v1 (rewritten in place by this ADR).
- Adjacent ADRs: ADR-0021 (charter and doctrine adoption), ADR-0023 (Pack-contract interfaces).

## Tradeoff Stanza

- **Solves:** the commercial-framing mismatch between v1's SaaS playbook and the post-charter engagement model; the friction in every commercial conversation that has to manually translate; the absence of a coherent reference document for new engagement-team hires.
- **Optimises for:** commercial coherence, engagement-team onboarding, channel-partner alignment, regulatory-partner alignment, future-investor readability.
- **Sacrifices:** 1500–2000 lines of in-place rewrite work; the optionality to keep both framings around for tactical flexibility; some early-adopter partners' familiarity with v1 pricing tiers.
- **Residual risks:** COST-MODEL desync if its rewrite slips (mitigated by Sprint S5 exit gate including COST-MODEL); historical partner re-orientation friction (mitigated by the explicit changelog at the top of v2); tax / accounting model mismatch for the Italian commercialista (mitigated by a separate finance-model briefing).
