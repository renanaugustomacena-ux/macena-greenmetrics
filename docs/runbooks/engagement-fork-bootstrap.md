---
title: Engagement fork bootstrap (Phase 1)
owner: "@greenmetrics/engagement-leads"
phase: 1 (Fork & Bootstrap)
calendar_target: 1 week
last_tested: 2026-04-30 (synthetic engagement)
review_date: 2026-10-30
---

# Runbook — Engagement fork bootstrap

**Doctrine refs:** Rule 77 (engagement code lives in `engagements/<client>/`), Rule 79 (engagement forks sync upstream quarterly minimum), Rule 80 (Core customisations time-bounded), Rule 150 (engagement fork created at Phase 1, not earlier), Rule 154 (Handover is operator-readiness).
**Charter refs:** §5.2 (engagement forks), §7 (upstream-sync discipline), §8.1 (engagement lifecycle Phase 1).

This runbook is the Phase 1 checklist for creating an engagement fork from the GreenMetrics modular template. It runs once per engagement, after Discovery (Phase 0) signs off and before any Pack-assembly work begins.

## 0. Pre-conditions

Phase 0 (Discovery) has produced and signed off the following artefacts. Bootstrapping without these is a Rule 149 violation and triggers a re-Discovery event:

- Signed Scope-of-Work (SoW) including engagement-id (e.g. `acme-2026`), client legal entity, deployment topology choice (A / B / C / D per Charter §10), engagement tier (T1 / T2 / T3), engagement-lead-of-record, and EGE / advisory partner roster.
- Pack matrix listing required Region / Protocol / Factor / Report / Identity Packs from the upstream catalogue; net-new Pack work flagged with sized estimates.
- Integration map enumerating client systems that will connect to the deployment (SCADA, ERP, SIEM, IdP, billing, ticketing).
- Discovery ADR filed at `docs/adr/<engagement-id>-0001-discovery-and-topology.md` in a private scratch repo (the engagement fork does not yet exist — see Rule 150).
- Engagement-fork organisation slot allocated (`github.com/<engagement-org>/`).
- Branding inputs from the client (legal entity, support contact, logos, theme colours) for `config/branding.yaml` overlay (Charter §6).

If any of the above is missing, stop. Bootstrapping a fork before SoW signature is a Rule 150 violation and creates phantom commercial liability.

## 1. Create the fork

```bash
# 1.1 Identify the upstream tag to pin.
UPSTREAM_TAG="v1.x.y"      # the latest stable release/v1.x line per Charter §5.1
ENGAGEMENT_ID="acme-2026"
ENGAGEMENT_ORG="acme"
TARGET_REPO="${ENGAGEMENT_ORG}/${ENGAGEMENT_ID}-greenmetrics"

# 1.2 Create the private engagement repo on GitHub.
#     The repo is created empty; the template is then mirrored in (not forked
#     via the GitHub fork button — a clean import preserves git history while
#     decoupling issues / PRs / discussions from the upstream).
gh repo create "$TARGET_REPO" --private --description "GreenMetrics engagement deployment for ${ENGAGEMENT_ORG}"

# 1.3 Mirror the template tag at the engagement-fork's main branch.
git clone --branch "$UPSTREAM_TAG" --single-branch \
  https://github.com/greenmetrics/template.git "${ENGAGEMENT_ID}-greenmetrics"
cd "${ENGAGEMENT_ID}-greenmetrics"
git remote rename origin upstream
git remote add origin "https://github.com/${TARGET_REPO}.git"

# 1.4 Record the upstream tag.
echo "$UPSTREAM_TAG" > template-version.txt

# 1.5 First commit on the engagement fork (signed).
git checkout -b main
git add template-version.txt
git commit -S -m "feat(engagement): bootstrap from upstream template ${UPSTREAM_TAG}"
git push -u origin main
```

The mirror approach (vs GitHub-fork-button) is deliberate: GitHub forks share PR / discussion namespaces with the upstream and surface the engagement repo to anyone who can see the upstream. Engagement repos are private to Macena + the engagement client; mirroring + rename keeps that boundary clean.

## 2. Engagement directory structure

Create the `engagements/<client>/` overlay (Charter §5.2, Rule 77) with the template skeleton:

```
engagements/<engagement-id>/
├── CHARTER.md                       # engagement-specific charter (one-pager)
├── CLAUDE.engagement.md             # engagement-specific invariants on top of CLAUDE.md
├── CORE-CUSTOMISATIONS.md           # any Core overrides + sunset date (Rule 80)
├── INCIDENT-RESPONSE.md             # engagement-specific incident overlay
├── PENTEST-LOG.md                   # engagement-specific pen-test findings (Rule 60)
├── SYNC-LOG.md                      # quarterly upstream sync record (Rule 79)
├── THREAT-MODEL-OVERLAY.md          # engagement-specific threat surface
├── adr/                             # engagement-specific ADRs (numbered <engagement-id>-NNNN)
├── fixtures/                        # engagement fixtures (synthetic-only by default per Rule 165)
├── packs/                           # engagement-specific Packs (overlays + private vendor)
└── runbooks/                        # engagement-specific runbooks (client-IdP-down, etc.)
```

Suggested seed contents (filled in during Phase 1 / Phase 2):

| File | Purpose | Seed content |
|---|---|---|
| `CHARTER.md` | one-page contract: scope, topology, tiers, exit criteria | client name, engagement-id, topology, tier, EGE partner, calendar |
| `CLAUDE.engagement.md` | engagement invariants for AI assistants | client privacy regime, IdP requirements, regulator preferences |
| `CORE-CUSTOMISATIONS.md` | every Core override with Tradeoff Stanza + sunset (Rule 80) | empty header + table; sunset reminder |
| `SYNC-LOG.md` | each upstream-sync record per Rule 79 | empty table: date / upstream-tag / reviewer / outcome |
| `THREAT-MODEL-OVERLAY.md` | engagement-specific entries on top of `docs/THREAT-MODEL.md` | client-specific TBs (VPN concentrators, on-prem IdP, etc.) |

Commit:

```bash
mkdir -p engagements/${ENGAGEMENT_ID}/{adr,fixtures,packs,runbooks}
# Populate the seed templates under engagements/<id>/.
git add engagements/
git commit -S -m "feat(engagement): scaffold engagements/${ENGAGEMENT_ID}/ overlay"
```

## 3. Configuration overlays

### 3.1 Branding

Edit `config/branding.yaml` to match the SoW:

```yaml
product_name: <ClientBrand-or-GreenMetrics>
legal_entity: <client legal entity>
support_contact: <client-support@…>
logo_header: engagements/<engagement-id>/branding/logo-header.svg
logo_login: engagements/<engagement-id>/branding/logo-login.svg
logo_pdf_cover: engagements/<engagement-id>/branding/logo-pdf.svg
favicon: engagements/<engagement-id>/branding/favicon.svg
theme_primary: "#0066CC"
theme_secondary: "#003366"
footer_text: "Powered by Macena GreenMetrics — © <legal entity>"
pdf_cover_template: engagements/<engagement-id>/branding/pdf-cover.html
```

Conformance test `tests/conformance/no_hardcoded_brand_test.go` (Sprint S6 enforcement) ensures no rogue strings outside `config/branding.yaml`.

### 3.2 Required Packs

Edit `config/required-packs.yaml` from the SoW Pack matrix:

```yaml
packs:
  - id: region-it             # always for IT engagements
    kind: region
    version: ">=1.0.0"
  - id: factor-ispra
    kind: factor
    version: ">=1.0.0"
  - id: factor-gse
    kind: factor
    version: ">=1.0.0"
  - id: protocol-modbus_tcp
    kind: protocol
    version: ">=1.0.0"
  # ... plus the rest of the SoW Pack matrix
```

The Pack-loader at boot refuses to start if any required Pack is missing or the version constraint is unsatisfied (Rule 73).

### 3.3 Topology choice

Set the Terraform workspace and `terraform/<topology>/` overlay per the chosen topology:

```bash
# Topology A — public-cloud single-tenant (default)
cd terraform/topology-a
terraform workspace new ${ENGAGEMENT_ID}
terraform init
# Edit terraform/topology-a/<engagement-id>.tfvars with engagement-specific values:
#   - aws_account_id, region (eu-south-1 for Italian residency)
#   - rds_instance_class, rds_storage_gb, rds_replicas
#   - eks_node_count, eks_node_instance_type
#   - cost_tag_engagement_id = "${ENGAGEMENT_ID}"
#   - cost_tag_engagement_tier = "T1" | "T2" | "T3"
```

For Topologies B / C / D, follow `docs/DEPLOYMENT-TOPOLOGY-{B,C,D}.md` (Sprint S6–S9 deliverables).

### 3.4 CLAUDE.engagement.md

Append the engagement-specific invariants on top of `CLAUDE.md`:

```markdown
# Engagement-specific invariants — <client name>

Reads alongside `CLAUDE.md` and `docs/MODULAR-TEMPLATE-CHARTER.md`.

- Regulatory regime: CSRD (wave-2), Piano Transizione 5.0, Conto Termico 2.0.
- Identity provider: <SAML | OIDC | local-DB> — `packs/identity/<id>/`.
- Privacy regime: GDPR + Garante guidance + ARERA classifications.
- Hosting topology: <A | B | C | D>, region <region>, owner <Macena | client>.
- EGE counter-signature partner: <name>, contact <email>, scope <Piano 5.0 | audit 102/2014 | both>.
- Sync cadence: quarterly minimum (Rule 79); sync window <month>.
- On-call ownership: T<1|2|3>; on-call rotation per `engagements/<id>/runbooks/on-call.md`.
```

## 4. Branch protection and CODEOWNERS

```bash
# 4.1 Enable branch protection on main: required reviews, required status checks,
#     no force push, no deletion.
gh api repos/${TARGET_REPO}/branches/main/protection \
  --method PUT --input branch-protection.json
# branch-protection.json carries:
# {
#   "required_status_checks": {"strict": true, "contexts": ["task verify", "policy-gate-*"]},
#   "enforce_admins": false,
#   "required_pull_request_reviews": {"required_approving_review_count": 1, "dismiss_stale_reviews": true},
#   "restrictions": null
# }

# 4.2 Add CODEOWNERS entries for engagement-overlay paths.
cat >> .github/CODEOWNERS <<EOF
/engagements/${ENGAGEMENT_ID}/    @${ENGAGEMENT_ORG}/engagement-leads @greenmetrics/engagement-leads
/config/branding.yaml             @${ENGAGEMENT_ORG}/engagement-leads
/config/required-packs.yaml       @greenmetrics/platform-team @${ENGAGEMENT_ORG}/engagement-leads
/terraform/                       @greenmetrics/platform-team
EOF
git add .github/CODEOWNERS
git commit -S -m "chore(engagement): branch protection + CODEOWNERS for ${ENGAGEMENT_ID}"
```

Per Charter §13 a Core override requires `engagements/<id>/CORE-CUSTOMISATIONS.md` with sunset date — engagement-overlay paths land via PR with the engagement-lead approving; Core changes from inside the fork land via the upstream contribution path documented in `docs/runbooks/upstream-sync.md`.

## 5. Initial verification

```bash
# 5.1 Run the template's conformance + property + security suites.
task verify

# 5.2 Run the engagement's pre-deploy validators.
task pack:list                       # enumerates loaded Packs
task pack:validate                   # validates each manifest
task verify:engagement               # engagement-specific checks (Phase 2 deliverable)

# 5.3 Boot the simulator stack.
docker compose --profile simulator up
curl -s http://localhost:8080/api/health | jq .
```

The Phase 1 success bar is `task verify` green and `/api/health` returning `status: "ok"` with the required Pack set listed under `dependencies`. Failures here gate Phase 2 (Rule 151).

## 6. First staging deploy

Per the chosen topology, deploy to a non-production environment per `terraform/topology-<a|b|c|d>/` overlay. The first staging deploy is the verification that the bootstrap is real, not an artefact of the local dev environment.

```bash
# Topology A example:
cd terraform/topology-a
terraform workspace select ${ENGAGEMENT_ID}-staging
terraform apply -var-file=${ENGAGEMENT_ID}.tfvars
# Wait for the EKS cluster + RDS + ArgoCD + ESO to come up.
# Then ArgoCD picks up the engagement-fork's GitOps manifest.
```

Successful first staging deploy = Phase 1 exit (Rule 154 dictates Phase 5, not Phase 1, but the first staging deploy is the explicit gate to start Phase 2).

## 7. Phase 1 exit gate

Move to Phase 2 (Pack Assembly) when all of the following are green:

- [ ] Engagement repo created at `github.com/<engagement-org>/<engagement-id>-greenmetrics` (private).
- [ ] `template-version.txt` committed and matches the upstream `release/v1.x` tag.
- [ ] `engagements/<engagement-id>/` overlay scaffolded with seed files.
- [ ] `config/branding.yaml` reflects client identity.
- [ ] `config/required-packs.yaml` reflects the SoW Pack matrix.
- [ ] `CLAUDE.engagement.md` drafted.
- [ ] Branch protection on `main` of the engagement fork; CODEOWNERS in place.
- [ ] `task verify` green on the engagement fork.
- [ ] First staging deploy successful via the topology overlay.
- [ ] First entry in `engagements/<engagement-id>/SYNC-LOG.md` recording the bootstrap upstream tag.

If any of the above is incomplete, do not start Phase 2. Phase-creep at the boundary is the most expensive overrun.

## 8. Anti-patterns

- **Forking before SoW signature.** Creates phantom commercial liability per Rule 150. Hold pre-signature work in scratch directories.
- **GitHub-button fork (vs mirror).** Surfaces the engagement repo on the upstream PR / discussion namespace. Use mirror clone + remote rename.
- **Skipping `template-version.txt`.** Without an explicit pin, upstream-sync (Rule 79) becomes guesswork. Always pin.
- **Overrides in Core, not in `engagements/<id>/`.** Core overrides break upstream-sync (Charter §7.2). Anything client-specific lives under `engagements/<id>/`; if it must touch Core, file `CORE-CUSTOMISATIONS.md` with sunset date (Rule 80).
- **Carrying upstream branding into a white-label engagement.** Edit `config/branding.yaml` at Phase 1, not Phase 3 — the conformance test catches it later but the cost of a Phase-3 rebrand is higher than a Phase-1 setting.
- **Skipping branch protection.** A solo engagement-lead pushing to `main` of the fork without review breaks the audit chain. Branch protection is non-negotiable.

## 9. Cross-references

- Charter `docs/MODULAR-TEMPLATE-CHARTER.md` §5.2 (engagement forks), §6 (white-label), §7 (upstream-sync), §8.1 (engagement lifecycle).
- Doctrine `docs/DOCTRINE.md` Rules 77, 79, 80, 149, 150, 151, 154, 165.
- Sister runbook: `docs/runbooks/upstream-sync.md` — the recurring follow-up that keeps the fork healthy.
- Sister doc: `docs/PACK-ACCEPTANCE.md` — the gate every Pack the engagement loads must pass.
- Plan `docs/PLAN.md` §5.3.10 (Sprint S5 deliverable).
