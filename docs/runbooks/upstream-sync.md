---
title: Upstream sync (engagement fork → template release/v1.x)
owner: "@greenmetrics/engagement-leads"
cadence: quarterly minimum
soak_window: 48h staging
last_tested: 2026-04-30 (synthetic engagement)
review_date: 2026-10-30
---

# Runbook — Upstream sync

**Doctrine refs:** Rule 78 (Core changes are merge-friendly between minor versions), Rule 79 (engagement forks sync upstream quarterly minimum), Rule 80 (Core customisations time-bounded), Rule 138 (regulatory Pack annual review).
**Charter refs:** §7 (upstream-sync discipline), §12 (compatibility commitments).

This runbook is the recurring procedure for syncing an engagement fork's `main` branch with the GreenMetrics template's `release/v1.x` line. Quarterly minimum per Rule 79; or before each engagement-deployment release, whichever is sooner.

## 0. Pre-conditions

- Engagement fork active for at least 1 quarter (recently-bootstrapped forks at `<v1.0.0` follow `engagement-fork-bootstrap.md` instead).
- Engagement-lead identified and on shift through the soak window.
- Staging environment available and not in use for client UAT during the sync window.
- Upstream `release/v1.x` head identified; release notes scanned for breaking changes (none expected per Rule 78, but verify).
- `engagements/<engagement-id>/SYNC-LOG.md` reviewed: most recent sync entry, any deferred resolutions, any open `CORE-CUSTOMISATIONS.md` rows approaching sunset (Rule 80).

A sync that has been deferred more than two consecutive quarters triggers an executive-sponsor call per Rule 79 — the fork is at risk of falling behind the upstream beyond the merge-friendly horizon.

## 1. Determine the target

```bash
ENGAGEMENT_ID="acme-2026"
cd "${ENGAGEMENT_ID}-greenmetrics"

# 1.1 Fetch upstream.
git fetch upstream

# 1.2 Identify the target tag on the same major-version line.
git tag -l --sort=-v:refname --merged upstream/release/v1.x 'v1.*' | head -5
# Pick the most recent tag that is at least one minor or patch ahead of
# template-version.txt. A jump across a major-version line is NOT a normal
# sync; that's a major-version migration with its own runbook (Charter §12.1).

CURRENT_TAG=$(cat template-version.txt)
TARGET_TAG="v1.x.y"   # the chosen target
```

If `TARGET_TAG` and `CURRENT_TAG` straddle a major-version boundary, stop and follow `docs/runbooks/major-version-migration.md` instead (Phase F deliverable).

## 2. Create the sync branch

```bash
git checkout main
git pull origin main

git checkout -b sync/upstream-${TARGET_TAG}-$(date +%Y%m%d)
```

The branch name encodes the target tag and date; it is short-lived and deleted after the merge to `main`.

## 3. Merge the upstream tag

```bash
git merge --no-ff "${TARGET_TAG}" -m "chore(sync): merge upstream ${TARGET_TAG}"
```

Conflicts will surface here. Resolve per the conflict policy in §6 below — do not auto-resolve in favour of either side without recording the choice.

## 4. Run the verification suites

The upstream's verification + the engagement's verification both run against the merged tree:

```bash
# 4.1 Template's conformance + property + security suites.
task verify
task test:integration
task test:property
task test:security
task test:conformance

# 4.2 Pack-loader contract checks against the engagement's required Pack set.
task pack:list
task pack:validate
task pack:contract-compliance

# 4.3 Engagement's own integration suite.
task verify:engagement
task test:engagement
task test:engagement:fixtures

# 4.4 Migration verification (forward-then-backward-then-forward).
task migrate:up-down-up

# 4.5 Performance baseline regression check.
task perf:bench
# Compare against the engagement's pinned baseline at engagements/<id>/perf-baseline.json.
# A >10% regression on any tracked SLO triggers a deeper investigation before deploy.
```

Any failure in §4 blocks the sync. Common categories and resolutions:

| Failure category | Resolution |
|---|---|
| Pack-contract version bumped | Update the engagement's Pack pin per Rule 71; rerun `pack:contract-compliance`. |
| Doctrine rule added that the engagement violates | File a `CORE-CUSTOMISATIONS.md` entry with a Tradeoff Stanza or fix the violation. |
| Schema migration conflict | Resolve in favour of upstream + apply engagement-specific data migration in `engagements/<id>/migrations/`. |
| New conformance test failing | The conformance test is the upstream's evidence; do not skip. Either fix the engagement code or escalate to upstream as a Pack-contract change request. |
| Performance regression | Investigate. If genuinely a Core regression, file an upstream issue and defer the sync. If engagement-specific (e.g. a Pack overlay), tune locally. |

## 5. Pack-contract version review

Per Rule 71, Pack contracts are versioned independently of Core. The sync is the moment to walk the supported window:

```bash
# Check Core's supported Pack-contract window.
grep -n PackContractVersion backend/internal/packs/contracts.go

# For each loaded Pack, check the manifest's pack_contract_version.
yq '.packs[] | "\(.id) \(.kind) \(.version)"' config/required-packs.yaml | while read pack kind version; do
  echo "=== ${kind}/${pack} @ ${version} ==="
  yq '.pack_contract_version' "packs/${kind}/${pack}/manifest.yaml" 2>/dev/null \
    || yq '.pack_contract_version' "engagements/${ENGAGEMENT_ID}/packs/${kind}/${pack}/manifest.yaml"
done
```

If a Pack-contract version has been deprecated by Core, the audit log emits `pack-contract-deprecated` events at boot until the engagement upgrades the Pack. The sync is the natural moment to do the upgrade; a deferred upgrade is fine but recorded in `SYNC-LOG.md`.

## 6. Conflict resolution policy

Per Charter §7.3:

- **Default:** upstream wins for Core; engagement wins for engagement-specific overlays under `engagements/<id>/`.
- **Engagement Core override:** if `engagements/<id>/CORE-CUSTOMISATIONS.md` carries a row covering the conflicting Core surface, the engagement override wins for the duration of the row's sunset window. After sunset (Rule 80), the row must either be upstreamed or removed; sunset rows are reviewed at every sync.
- **Schema conflicts:** upstream wins; the engagement's overlay carries an `engagements/<id>/migrations/` data-migration to bridge.
- **Doctrine conflicts:** upstream's doctrine wins; the engagement files an unrejection ADR (Rule 209) if it disagrees.

Record every non-default resolution in `engagements/<id>/SYNC-LOG.md`:

```markdown
| Date | Upstream tag | Reviewer | Conflict | Resolution | Rationale |
|---|---|---|---|---|---|
| 2026-07-15 | v1.3.2 | engagement-lead + platform-reviewer | Core schema added column `meters.commissioned_at`; engagement overlay had column with same name but different default | upstream wins; engagement migration drops the engagement column and reuses the new Core column with a backfill job | upstream evidence-base + Charter §7.3 default |
```

## 7. Pre-deploy review and signoff

```bash
# 7.1 Open a sync PR for review.
git push -u origin "sync/upstream-${TARGET_TAG}-$(date +%Y%m%d)"
gh pr create --base main --head "sync/upstream-${TARGET_TAG}-$(date +%Y%m%d)" \
  --title "chore(sync): upstream ${CURRENT_TAG} -> ${TARGET_TAG}" \
  --reviewer @<engagement-lead>,@greenmetrics/platform-team

# 7.2 Reviewers' checklist (engagement-lead + Macena platform-team reviewer):
#   - All §4 suites green in CI.
#   - Pack-contract versions consistent.
#   - Conflict-resolution entries in SYNC-LOG.md complete and rationaled.
#   - CORE-CUSTOMISATIONS.md sunsets reviewed; expired rows decisioned.
#   - No new feature flags forced on (sync is not the moment for behaviour changes).
#   - Performance regression check within tolerance.
```

Both sign-offs are required (engagement-lead + Macena platform-team reviewer) per Charter §7.3.

## 8. Staging deploy + 48-hour soak

```bash
# 8.1 Merge sync PR into main.
gh pr merge --squash --auto

# 8.2 Deploy main to staging via ArgoCD (auto-sync if enabled, or manual).
argocd app sync ${ENGAGEMENT_ID}-staging

# 8.3 Validate the staging deploy.
curl -s https://staging.${ENGAGEMENT_ID}.greenmetrics.example/api/health | jq .
task smoke:staging

# 8.4 Soak window: 48 hours.
#   - Monitor SLO dashboards (every 4h is enough).
#   - Track incident rate; any P1/P2 incident in the window blocks production.
#   - Run the simulator-driven property suite hourly.
#   - Validate one report-generation cycle (CSRD E1 or Piano 5.0) end-to-end.
```

A failed soak is not a sync failure; it is a deferred deploy. Production stays on the previous tag until the soak passes.

## 9. Production deploy

```bash
argocd app sync ${ENGAGEMENT_ID}-production --revision main
# Argo Rollouts AnalysisTemplate watches SLO burn-rate; auto-rollback on burn.
```

Post-deploy:

- Update `template-version.txt` to the new `${TARGET_TAG}`.
- Append the sync record to `engagements/<id>/SYNC-LOG.md`.
- Notify the engagement client per the SoW communication clause (e.g., "monthly status email" or "release notes in the client portal").
- Close the sync PR; delete the sync branch.

## 10. Rollback (Sev-2 path)

If a regression surfaces in production within the first 7 days post-sync:

```bash
# 10.1 Roll back the deployment to the previous tag.
git tag -l --sort=-v:refname 'v1.*' | head -5
PREV_TAG=${CURRENT_TAG}   # the tag that was deployed before the sync
argocd app set ${ENGAGEMENT_ID}-production --revision "${PREV_TAG}"
argocd app sync ${ENGAGEMENT_ID}-production

# 10.2 Capture the regression in the engagement's incident log + a draft upstream
#       issue. The next sync attempt waits for the upstream fix or for the
#       engagement-specific workaround.
```

Rollbacks are a last resort. The 48-hour staging soak is designed to surface regressions before production; if a regression makes it to production, the soak window or the perf-bench tolerance probably needs tightening.

## 11. Annual review

At each Q1 sync (the post-Christmas / pre-Easter window), perform additional housekeeping per Rule 138:

- Walk every `CORE-CUSTOMISATIONS.md` row; sunsets within 6 months get a decision now.
- Walk every Pack version pin; bump to the latest minor for Region / Factor / Report Packs that follow regulator update calendars.
- Review the Pack-contract supported window in `internal/packs/contracts.go`; deprecation announcements for the next year are noted in the engagement's roadmap.
- Re-verify the `engagements/<id>/perf-baseline.json` against current production traffic.

## 12. Anti-patterns

- **Skipping conformance tests.** They are the floor. Skipping them produces an untested merge.
- **Resolving conflicts in favour of stale fork code without recording the rationale.** SYNC-LOG.md is a regulatory artefact; `git merge -X theirs` without commentary is unauditable.
- **Merging without staging soak.** The 48-hour soak is the cheapest insurance against regressions reaching production.
- **Bundling a sync with a feature change.** A sync is a sync. New features ship in their own PR after the sync lands.
- **Deferring two consecutive quarters without an ADR.** Rule 79 sets the cadence; deferral requires explicit ADR per the same rule.

## 13. Cross-references

- Charter `docs/MODULAR-TEMPLATE-CHARTER.md` §7 (upstream-sync discipline), §12 (compatibility commitments).
- Doctrine `docs/DOCTRINE.md` Rules 78, 79, 80, 138.
- Sister runbook: `docs/runbooks/engagement-fork-bootstrap.md` — the one-time setup that this runbook recurs against.
- Sister doc: `docs/PACK-ACCEPTANCE.md` — the gate every Pack version bump must pass at sync time.
- Plan `docs/PLAN.md` §5.4.6 (Sprint S6 engagement-fork model documents).
