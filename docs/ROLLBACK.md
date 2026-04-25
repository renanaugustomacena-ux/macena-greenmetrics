# Rollback Policy

**Doctrine refs:** Rule 15, Rule 20, Rule 21, Rule 56, Rule 60, Rule 63.

## 1. Argo Rollouts (Kubernetes Deployments)

```bash
# List recent revisions.
argocd app history greenmetrics

# Roll back to previous revision.
argocd app rollback greenmetrics <revision>

# Or directly via Argo Rollouts.
kubectl argo rollouts undo rollout/greenmetrics-backend -n greenmetrics
```

Auto-rollback: AnalysisTemplate watches Prometheus burn-rate; on threshold breach calls `argo rollouts undo` automatically (no human in critical path; Rule 56). See `gitops/base/argo-rollouts/analysis-template.yaml` (S4).

## 2. Container image rollback (Argo CD GitOps)

```bash
# Edit gitops/staging/ or gitops/production/ to pin previous digest.
git -C gitops checkout -b rollback/greenmetrics-backend-<revision>
sed -i "s|@sha256:.*|@sha256:<previous>|" gitops/production/kustomization.yaml
git commit -am "rollback: greenmetrics-backend to <previous>"
gh pr create --title "rollback: greenmetrics-backend to <previous>" --body "..."
# Merge → Argo CD reconciles within 60s.
```

## 3. Database schema rollback

**Forward-only in production** (Rule 21). For schema rollback in production:

1. Take cluster read-only (`READ_ONLY_MODE=true`).
2. Use `pg_dump` / `pg_restore` from snapshot (last 4h).
3. Replay forward-fix migration.
4. Resume writes.

Down migrations exist for testability + dev rollback only.

For rollback within a single transaction's worth of damage, see `docs/runbooks/db-outage.md`.

## 4. Terraform rollback

```bash
# Revert the offending PR or apply a corrective PR.
git revert <commit>
# CI runs terraform plan in production env → review → apply via gitops apply path.
```

**Never** run `terraform destroy -target` against production resources without 4-eyes approval + ADR.

## 5. Configuration rollback

```bash
# ConfigMap / env-var rollback via Argo CD or kubectl.
kubectl set env deployment/greenmetrics-backend -n greenmetrics KEY=<previous-value>
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
```

## 6. Secret rollback

```bash
# AWS Secrets Manager versioning.
aws secretsmanager update-secret-version-stage \
  --secret-id greenmetrics/prod/jwt \
  --version-stage AWSCURRENT \
  --move-to-version-id <previous>
# Force ESO sync; restart pods.
kubectl annotate externalsecret -n greenmetrics greenmetrics-backend force-sync="$(date +%s)" --overwrite
kubectl rollout restart deployment/greenmetrics-backend -n greenmetrics
```

## 7. JWT rollback

See `docs/runbooks/jwt-secret-rotation.md` §Rollback.

## 8. Anti-patterns rejected

- Direct `kubectl edit` against production — Argo CD reverts within 60s; you're fighting a losing battle. Make the change in `gitops/`.
- Rollback by deleting a Kustomize patch line without commit history — leaves no audit trail.
- Down-migration in production — use forward-fix.
- "Just nuke and re-deploy" — no — staged rollback preserves audit + RPO.
