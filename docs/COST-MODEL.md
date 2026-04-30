# GreenMetrics Cost Model — v2 (engagement model)

**Status:** v2, supersedes v1 (the SaaS unit-economics model).
**Date adopted:** 2026-04-30.
**Doctrine refs:** Rule 1 (money is cents+ISO-4217), Rule 13 (abstractions are cost centres), Rule 27 (Tradeoff Stanza), Rule 38 (capacity is intentional), Rules 151 / 152 (Phase 2 / Phase 3 bounded), Rule 156 (engagement health monitored monthly), Rule 168 (portfolio feedback to roadmap).
**Charter refs:** §9 (economic model), §10 (deployment topologies).
**ADR refs:** ADR-0021 (charter and doctrine adoption), ADR-0028 (MODUS_OPERANDI v2).
**Owners:** `@greenmetrics/platform-team`, `@greenmetrics/engagement-leads`, finance.
**Review cadence:** per-engagement P&L monthly; portfolio aggregate quarterly; full re-fit at each charter office hours (six-monthly).

---

## 0. What changed since v1

If you read v1, you saw:

- per-tenant unit economics priced at €/meter/month;
- a SaaS gross-margin target of 4× cost;
- a single per-tenant cost line (€2.20/Medium) and a customer-pricing line (€10/Medium);
- AWS Budgets sized against environment totals as if all tenants shared one production deployment.

That model is retired with the pivot recorded in `docs/MODULAR-TEMPLATE-CHARTER.md`. The unit of sale is now the **engagement**, not the meter; the unit of deployment is one isolated single-tenant deployment per engagement (Charter §11) — multi-tenant is a partner-hosted opt-in, not the default. The cost model therefore tracks two things:

1. **Per-engagement P&L** (§1). Revenue across the four streams (license + customisation + annual maintenance + tier retainer + regulatory-deliverable services), set against engagement delivery cost. Margin target by tier per Charter §9.3.
2. **Per-deployment infra cost** (§2). The recurring cloud / on-prem cost of running one deployment, scaled with meter count and topology. This is the cost line in the engagement P&L.

The aggregate-portfolio view (§3) replaces the v1 environment-totals framing. AWS Budgets are now sized per-deployment (§3.3), not per-environment.

Vocabulary changes: CAC / LTV / churn / per-meter ARPA are absent by design. The replacement KPIs are engagement margin, time-to-customisation, template-fit-score, net-engagement-value, annual-maintenance-attach-rate (Charter §9.1, MODUS_OPERANDI §3.9).

---

## 1. Per-engagement P&L

The engagement P&L is the primary commercial-reporting unit. Each active engagement carries a P&L sheet covering Year 0 (Phases 0–5, the delivery year) and Years 1–5 (operations).

### 1.1 Revenue lines

Per Charter §9.1 and MODUS_OPERANDI v2 §3:

| Stream | Cadence | Sizing |
|---|---|---|
| Engagement license | one-time at signing | Light €40–80k / Standard €80–140k / Complex €140–220k / Strategic €220–500k+ |
| Discovery (Phase 0) | one-time | Standard €15–30k (billable; deal-killer surfacing pays for itself) |
| Customisation services | T&M, Phase 2 + 3 | €60–300k engagement-dependent; 30/60/10 milestone billing |
| Annual maintenance | yearly | 18–22% of license; attach target ≥ 90% T1 / ≥ 95% T2/T3 |
| Co-managed retainer (T2) | monthly | €4–12k/mo |
| Fully-managed retainer (T3) | monthly | €18–55k/mo |
| Regulatory-deliverable services | one-off | Piano 5.0 attestazione €3.5–15k; CSRD ESRS E1 dossier €5–50k; audit 102/2014 €6–12k |

### 1.2 Cost lines

| Cost centre | Driver |
|---|---|
| Engagement-lead time | calendar weeks in Phase 0–5 × loaded weekly rate |
| EGE / domain advisor time | counter-signature + dossier review hours × rate |
| Per-deployment infra | §2 below × engagement duration |
| Pack-update labour (maintenance) | quarterly Pack-version review + emergency patches |
| Regulatory-deliverable delivery | per-deliverable hours × rate |
| Template R&D amortisation | template engineering cost spread across the engagement portfolio (see §3) |
| Customer-success / on-call (T2/T3 only) | named on-call hours + quarterly review |

### 1.3 Worked example — Standard T1 engagement, 5-year horizon

Anchored on MODUS_OPERANDI v2 §11.3:

```
Revenue
  Engagement license              €160 000
  Customisation services          €130 000
  Annual maintenance (Y1–Y4)      €120 000  (4 × €30k; 18.75% of license / yr)
  Regulatory-deliverable services  €30 000  (CSRD E1 + Piano 5.0 reviews)
  ── 5-year revenue              €440 000

Delivery cost
  Phase 0–5 engagement-lead        €40 000  (~11 wk × €3.6k loaded)
  Phase 0–5 advisory + EGE         €18 000
  Per-deployment infra (5 yr)      €18 000  (Topology A medium, §2.2)
  Annual Pack-update labour        €25 000  (5 × ~€5k/yr)
  Regulatory-deliverable delivery  €12 000
  Template R&D amortisation        €27 000  (~17% of license per §3.2)
  ── 5-year cost                  €140 000

Engagement margin
  Net (5-year)                    €300 000
  Gross margin                       68.2 %
```

The 68.2% gross margin clears the T1 ≥ 65% target (Charter §9.3) and matches the canonical figure in MODUS_OPERANDI v2 §11.3. The P&L sheet is the deliverable; this snapshot is illustrative only — the actual engagement margin is computed monthly from the engagement-lead's time-tracking and the per-deployment cost-export.

### 1.4 Margin targets (Charter §9.3)

| Tier | Gross margin target | Maintenance attach |
|---|---|---|
| T1 (handover) | ≥ 65% | ≥ 90% |
| T2 (co-managed) | ≥ 55% | ≥ 95% |
| T3 (fully-managed) | ≥ 45% | ≥ 95% |

A closed engagement landing below target by more than 10 percentage points is a Sev-2 commercial event — root cause is filed in the engagement retrospective and feeds the next quarterly portfolio review (Rule 168).

### 1.5 Margin levers (engagement-level)

- **Pack reuse rate.** Engagement #N reusing 95%+ of catalogue Packs spends ~20% on engagement-specific code; engagement #1 (Italian flagship) carries the catalogue-build cost. Tracked per engagement.
- **Customisation-time discipline.** Phase 3 bounded at 2–4 weeks per Rule 152; beyond-scope requests are change-orders, not extensions. The most common margin leak is unbilled customisation drift.
- **Discovery rigor.** A €15–30k Discovery that surfaces a deal-killer before Phase 1 saves 8–14 weeks of Phase 2–5 cost — Discovery margin matters less than the disqualification it produces.
- **T1 → T2 conversion.** A T1 engagement that adds a co-managed retainer at month 12 raises 5-year net-engagement-value by ~€60–150k without proportionate delivery cost.

---

## 2. Per-deployment infrastructure cost

A deployment is one engagement's running system. Cost lines depend on topology (Charter §10) and meter count.

### 2.1 Cost breakdown (Topology A — public-cloud single-tenant, AWS eu-south-1)

| Cost centre | Driver |
|---|---|
| Ingest storage | rows/day × bytes × retention / compression × €/GB-month gp3 (~€0.115) |
| CAGG storage | ~25% of raw |
| RDS-Timescale compute | per-deployment primary + 2 replicas; profile-dependent |
| Backend compute (HPA) | rps × €/h; HPA on `gm_ingest_readings_total` + CPU |
| Frontend serving | static asset size × CDN egress |
| Worker (Asynq) | report jobs/month × avg minutes × €/h |
| Bandwidth | egress on report download + DSO outbound |
| Grafana | flat-cost amortised |
| Secrets Manager + KMS | flat per stored secret + per-request |
| WAF | per-million-requests |

### 2.2 Worked profiles (Topology A)

```
Small        (10 meters,    1 reading / 15 min):    ~€35  / month
Medium       (100 meters,   4 readings / min):      ~€300 / month
Large        (1000 meters,  4 readings / min):      ~€1.4k / month
Stretch      (5000 meters,  12 readings / min):     ~€6.5k / month
```

Per-deployment cost in single-tenant framing is dominated by RDS-Timescale (60–75% of the bill on Medium and above). The Stretch profile is the upper bound encountered in the Italian-flagship engagement pipeline; clients above 5000 meters are routed to Topology D (hybrid OT-segment ingest) to keep costs predictable.

### 2.3 Topology cost multipliers (relative to Topology A baseline)

| Topology | Recurring infra cost | Notes |
|---|---|---|
| A — public-cloud (AWS eu-south-1) | 1.0× | baseline |
| B — Italian-sovereign (Aruba / Seeweb / TIM Enterprise) | ~1.3–1.5× | similar Compute/Storage; Vault adds operational labour |
| C — on-prem (client K3s in client DC) | client capex; Macena ≈ 0.3× of A | Macena cost is config + ops; hardware on the client books |
| D — hybrid | ~1.1–1.2× | OT-side small node + IT-side reduced cloud load |

### 2.4 What is no longer modelled

The v1 "per-tenant share of RDS compute" line is gone — single-tenant deployments do not share RDS compute. The "tenants_share" denominator from v1 §1 is unused in v2; if a partner ESCO operates a multi-tenant deployment, that's the partner's allocation problem, not the template's.

---

## 3. Portfolio aggregate

### 3.1 Aggregate composition

The portfolio aggregate is the sum across active engagements:

- **Active engagement count** (Phase 2 + Phase 3 + Phase 4 + Phase 5 in flight; T2/T3 retainers in run mode).
- **Aggregate engagement margin** (revenue across all engagements minus delivery cost across all engagements + template R&D not yet amortised).
- **Aggregate per-deployment infra cost** (sum of §2 lines for each active deployment).
- **Capacity utilisation** (engagement-lead-weeks committed ÷ engagement-lead-weeks available; trigger to hire when >85% sustained for 6 weeks).
- **Pack-catalogue contribution rate** (Rule 168) — engagements per quarter that contribute at least one generalised Pack upstream.

### 3.2 Template R&D amortisation

Template engineering (Sprints S1–S22, est. ~36 weeks of founder + Phase H hires through end of Phase J) is amortised across the engagement portfolio. The current basis: an engagement license carries an attributed share of template R&D equal to ~15–20% of the license — the remainder of the license value is the IP rights and the regulator-grade engineering substrate. This share is reviewed quarterly and adjusted as the portfolio expands.

### 3.3 AWS Budgets (per-deployment)

Per-deployment caps replace v1's per-environment caps:

| Deployment profile | Soft target | Hard cap |
|---|---|---|
| Small | €100/mo | €200/mo |
| Medium | €350/mo | €600/mo |
| Large | €1.6k/mo | €2.5k/mo |
| Stretch | €7k/mo | €10k/mo |

`terraform/modules/cost/main.tf` provisions per-deployment `aws_budgets_budget` resources with thresholds at 50%, 80%, 100%, 120%; notifications go to `#greenmetrics-ops` + the engagement-lead's email. Budget actions are alert-only (auto-stop on spend would harm availability).

Template environments (`dev`, `staging`) retain the v1 caps:

| Env | Monthly target | Hard cap |
|---|---|---|
| dev | €100 | €200 |
| staging | €500 | €750 |

### 3.4 Cost-allocation tags

Every cloud resource tagged with `Project=greenmetrics-<engagement-id>`, `Environment=<env>`, `Tenant=<tenant-id>`, `EngagementTier=<T1|T2|T3>`, `DataResidency=<region>`. Enforced by `policies/conftest/terraform/tags.rego`. Untagged resources fail policy gate.

### 3.5 AWS CUR pipeline (per-deployment)

CUR delivery to `s3://greenmetrics-aws-cur-<engagement-id>` (KMS-encrypted, per engagement). Daily delivery; CSV. Pulled into the per-deployment Grafana via the AWS Cost Datasource plugin. The portfolio dashboard at `monitoring/grafana/dashboards/engagement-portfolio.json` (Phase E Sprint S8 deliverable) federates per-deployment cost into the aggregate.

---

## 4. Efficiency tradeoffs

The technical levers from v1 are unchanged — they apply per deployment.

| Lever | Saves | Costs |
|---|---|---|
| Timescale compression on ≥ 7d chunks (10×) | 80% storage | 3× decompression query latency on cold data |
| CAGG refresh policy 5 min vs 1 min | 80% refresh CPU | 4 min CAGG lag (within SLO) |
| HPA scale-down stabilization 300s | 30% over-provision avoidance | slight latency spikes on scale-up |
| Asynq + Redis vs in-process | OLTP pool isolation; SLO protection | Redis ≈ €30/mo + 1 failure domain |
| Distroless nonroot vs slim | smaller attack surface | larger image (~300 MB vs ~80 MB), pull bandwidth + GHCR storage |
| Multi-AZ RDS | DR + zero-downtime failover | 2× DB cost |
| Object Lock on audit bucket | compliance + tamper resistance | cannot delete even legitimately stale entries |

Engagement-level levers (new in v2):

| Lever | Saves | Costs |
|---|---|---|
| Pack reuse over Pack-fork | engagement labour 60–80% | upstream-sync discipline tax (Rule 79) |
| Synthetic fixtures only | data-handling labour + risk | dev-realism gap (mitigated by simulators) |
| Per-engagement dashboards from JSON template | UI labour 50–70% | overlay drift if not re-applied at sync |
| EGE counter-signature pre-arranged at Phase 0 | Phase 4–5 calendar shrinkage | Discovery overhead (already paid via §1.1) |
| T1 → T2 conversion at month 12 | recurring revenue without new license labour | named on-call commitment |

Each tradeoff carries a Tradeoff Stanza per Rule 27.

---

## 5. Waste detection

`scripts/ops/cost-audit.sh` runs monthly per active deployment:

- unused EBS volumes (no attachment > 7d);
- idle RDS replicas (CPU < 5% sustained 30d);
- orphaned secrets (Secrets Manager `LastAccessedDate` > 90d);
- Grafana datasources with zero queries 30d;
- unused IAM roles (CloudTrail no `AssumeRole` event 90d);
- under-utilised K8s nodes (CPU < 20% sustained 7d).

Engagement-level waste detection (new in v2):

- **Customisation overrun.** Phase 3 hours billed beyond the SoW without a change-order — flagged at the weekly engagement-lead review (Rule 152).
- **Pack drift cost.** A fork that has not synced for two consecutive quarters per Rule 79 — labour to catch up grows superlinearly; flagged in the engagement-health dashboard.
- **Maintenance under-utilisation.** A T1 engagement consuming < 20% of allocated maintenance hours over 12 months — surface a value-conversation, not an upsell pitch.

Output of each detection batch → GitHub issue `monthly-cost-audit-<engagement-id>-YYYY-MM` (per-deployment) and `monthly-portfolio-audit-YYYY-MM` (portfolio).

---

## 6. Cost regression in CI

`infracost` integrated in `.github/workflows/terraform-plan.yml`:

- PR comment shows monthly cost diff per resource for every targeted deployment.
- Threshold: any +€500/month change requires `cost-approved` label + ADR.
- Per-engagement deployments: the threshold is 10% of the deployment's soft target (§3.3), not a flat €500 — a Stretch deployment can absorb €500/mo of legitimate growth that a Small cannot.

---

## 7. Anti-patterns rejected (Rule 26 / 46 / 66)

- **"Per-meter monthly subscription on top of the engagement license."** Rejected by Charter §13 — we don't bill on volume.
- **"Self-serve signup endpoint at /signup."** Rejected by Charter §13 — industrial customers don't self-serve; an unscoped tenant is an unbillable engagement.
- **"Just spin up a bigger RDS — €€€ are not the user's problem."** Rejected; per-deployment infra cost is the engagement-lead's surface and the engagement margin's denominator.
- **"Disable Cost Explorer, it costs money."** Rejected; Cost Explorer is single-digit euros per month and surfaces thousands of euros of surprise.
- **"Add a tenants_share denominator so we can run multi-tenant on one RDS for cheaper."** Rejected for the default deployment — see Charter §11. Acceptable only for partner-hosted deployments, owned by the partner's allocation model.
- **"Reserved Instances right after engagement signing."** Rejected until per-deployment usage stabilises 6 months — RIs locked too early forfeit flexibility through the Phase 4 / Phase 5 capacity-tuning cycle.
- **"Cost dashboard for marketing."** Rejected — cost data is engagement-confidential and feeds the engagement-lead, finance, and the engagement client; it is not a public-facing surface.

---

## 8. Doctrine cross-references

- Rule 1 — money is `(amount_cents int64, currency ISO-4217 string)`; every figure here is a denomination, never a float.
- Rule 13 — abstractions are cost centres; cost-model abstractions (per-deployment vs per-engagement) carry the same ledger discipline.
- Rule 27 / 47 / 67 — every tradeoff in §4 carries a Tradeoff Stanza in the originating ADR.
- Rule 38 — capacity is intentional; the per-deployment infra cost (§2) is the financial face of the capacity model in `docs/CAPACITY.md`.
- Rule 151 / 152 — Phase 2 / Phase 3 bounding is the structural discipline behind the engagement-margin lever.
- Rule 156 — engagement health is monitored monthly; the engagement P&L is one of the health-score inputs.
- Rule 168 — portfolio feedback to roadmap; aggregate margin trends drive Pack-catalogue investment decisions.
- ADR-0009 — circuit breaker cost / behaviour rationale (DSO outbound).
- ADR-0014 — Asynq + Redis cost / failure domain rationale.
- ADR-0021 — charter and doctrine adoption (umbrella for the v1 → v2 reframe).
- ADR-0028 — MODUS_OPERANDI v2 (sibling rewrite that this document tracks).

---

## 9. Tradeoff Stanza (this document)

- **Solves:** the desync between MODUS_OPERANDI v2's engagement framing and v1's per-tenant SaaS unit economics; the absence of a per-engagement P&L; the v1 AWS Budgets framing that assumed shared-infrastructure multi-tenancy.
- **Optimises for:** engagement-margin discipline at delivery time, predictable per-deployment cost ceilings, portfolio-level visibility into capacity utilisation and amortisation, alignment with Charter §9 economic model.
- **Sacrifices:** the simplicity of a single per-tenant cost number; the directness of "€/meter/month" as a procurement-conversation primitive; some operational labour in maintaining per-deployment cost dashboards instead of one production-wide one.
- **Residual risks:** template R&D amortisation rate is best-guessed at v1.0.0 (§3.2) and will need re-fitting after the first 5 engagements close; per-engagement infra-cost ceilings (§3.3) may understate Stretch deployments until real telemetry lands; engagement-lead time-tracking discipline is required for the P&L to be honest, and skipping it is the most common margin-misreporting failure.
