# Troubleshooting Tree

**Doctrine refs:** Rule 17 (DX — debuggability), Rule 20 (operational reality).

Decision tree for the most common failure symptoms. For full incident response see `docs/INCIDENT-RESPONSE.md` (S5). For per-failure runbook see `docs/runbooks/`.

## 1. Backend won't boot

```
fresh `make run` / docker compose up → backend exits immediately
│
├── error: refusing default JWT_SECRET in production
│     → cause: APP_ENV=production but JWT_SECRET still placeholder
│     → fix: rotate JWT_SECRET via ESO / Secrets Manager (RISK-013)
│
├── error: refusing default GRAFANA_ADMIN_PASSWORD in production
│     → cause: APP_ENV=production but GRAFANA_ADMIN_PASSWORD still placeholder
│     → fix: rotate via ESO
│
├── error: refusing DATABASE_URL with sslmode=disable in production
│     → cause: DATABASE_URL has sslmode=disable in production
│     → fix: switch to sslmode=require; ensure RDS TLS cert present
│
├── error: JWT_SECRET too short (<32)
│     → cause: env override too short
│     → fix: openssl rand -base64 48
│
├── error: failed to dial OTel collector
│     → cause: OTEL_EXPORTER_OTLP_ENDPOINT set to unreachable host
│     → fix: clear var (disables tracing) or fix collector deploy
│
└── error: pgx pool acquire failed
      → cause: DATABASE_URL wrong / DB down
      → fix: check `docker compose ps`, `pg_isready`, network
```

## 2. Readings missing in Grafana

```
ingest endpoint reports 200 / 202 but Grafana dashboard empty
│
├── Did the read endpoint find them? → curl /api/v1/readings?meter_id=...
│
├── If raw readings present but aggregated empty:
│     → CAGG refresh lag; see metric `gm_cagg_refresh_duration_seconds`
│     → manual refresh: SELECT refresh_continuous_aggregate('readings_15min', NULL, NULL)
│
├── If raw readings missing:
│     → tenant_id mismatch in query? RLS hiding?
│     → check `gm_ingest_dropped_total{reason="queue_full"}` (S4)
│     → check ingestor goroutine alive: `gm_ingest_readings_total{protocol="modbus"}`
│
└── If specific meter empty:
      → meter `active=false` in DB? `SELECT * FROM meters WHERE id=...`
      → ingestor for that protocol up? `kubectl logs deploy/greenmetrics-backend | grep modbus`
```

## 3. Reports come back empty / wrong

```
POST /api/v1/reports returns 202; GET /api/v1/jobs/{id} → status=succeeded; report empty
│
├── Period bounds wrong? → check period_from/period_to are RFC 3339 UTC
│
├── Emission factors missing for the period?
│     → SELECT * FROM emission_factors WHERE valid_from <= ... AND (valid_to IS NULL OR valid_to >= ...)
│     → if missing: trigger ISPRA factor refresh; check breaker state `gm_breaker_state{name="ispra_emission_factors"}`
│
├── Tenant has no meters in scope?
│     → SELECT * FROM meters WHERE tenant_id=... AND active=true
│
└── Report builder bug?
      → check golden-file test for that dossier (`internal/domain/reporting/testdata/`)
      → file an issue with full job_id + tenant_id + payload
```

## 4. Alerts not firing

```
threshold breached but no Alertmanager notification
│
├── Prometheus rule active?
│     → Grafana → Alerting → Rules: rule status `Firing` / `Pending` / `Inactive`
│
├── Alertmanager routing matches?
│     → kubectl exec -n monitoring deploy/alertmanager -- amtool config show
│
├── Slack webhook valid?
│     → kubectl logs deploy/alertmanager | grep webhook
│
├── App-level alert engine running?
│     → `gm_alert_fired_total{rule="..."}` ticking? if not, alert engine bug
│
└── Inhibition rule silencing?
      → amtool silence query; remove if stale
```

## 5. Grafana panel empty

```
Grafana shows "No data" on a known-good metric
│
├── Datasource UID matches dashboard?
│     → Settings → Data sources → check UID
│
├── Datasource reachable from Grafana pod?
│     → kubectl exec deploy/grafana -- curl prometheus:9090/api/v1/query
│
├── Cardinality dropped?
│     → Prometheus `up` metric for the target
│
└── Time range too narrow?
      → expand to 1h
```

## 6. ESO not syncing

```
kubectl get externalsecret -n greenmetrics → Status NotSynced
│
├── ClusterSecretStore reachable?
│     → kubectl describe clustersecretstore aws-secrets-manager
│
├── IRSA assume working?
│     → kubectl exec deploy/external-secrets -- aws sts get-caller-identity
│     → expect identity = greenmetrics-external-secrets-irsa
│
├── Secret exists in AWS Secrets Manager?
│     → aws secretsmanager describe-secret --secret-id greenmetrics/prod/jwt
│
└── Refresh interval too long?
      → kubectl annotate externalsecret -n greenmetrics --all force-sync=$(date +%s) --overwrite
```

## 7. Cosign verify failing

```
kubectl apply triggers Kyverno admission denial
│
├── Image actually signed?
│     → cosign verify ghcr.io/.../image@sha256:... (see SUPPLY-CHAIN.md §7)
│
├── Identity / issuer match Kyverno policy?
│     → kubectl get clusterpolicy verify-greenmetrics-images -o yaml
│     → verify subject regex matches GitHub workflow URL
│
├── Sigstore Rekor reachable from cluster?
│     → kubectl exec test-pod -- curl https://rekor.sigstore.dev/api/v1/log
│
└── Certificate chain valid?
      → annual review per docs/SECOPS-RUNBOOK.md
```

## 8. ArgoCD reverting your change

```
manual `kubectl apply` reverted within 60s
│
└── This is the design (RISK-019). Argo CD `selfHeal: true` enforces gitops as truth.
    Make the change in `gitops/` instead. Or open a PR.
```

## 9. CI failing: pre-commit-ci

```
pre-commit-ci job red on PR
│
├── gitleaks finding → check `.gitleaks.toml` allowlist with rationale
├── golangci-lint → run `make lint` locally; fix
├── prettier / eslint → run `make precommit` locally; fix
├── kubeconform → run `kubeconform -strict k8s/` locally; fix
├── conftest → run `make policy-check` locally; fix or open ADR for waiver
├── actionlint → check workflow file syntax
├── markdownlint → check ADR + docs files; especially Tradeoff Stanza presence
└── tradeoff-stanza missing → add Solves/Optimises/Sacrifices/Residual to ADR
```

## 10. Anything else

Open an issue with:

- Symptom (what you saw).
- Expected vs actual.
- Steps to reproduce.
- Logs (filtered by `request_id` if available).
- Trace ID (Tempo URL if applicable).
- `docker compose ps` output.
- Backend version (`curl /api/health | jq .version`).

Add label `troubleshooting`. Triage in next platform office hours.
