# Debug Guide

**Doctrine refs:** Rule 17 (DX), Rule 18 (observability), Rule 40 (backend observability).

## 1. Backend (Go + Fiber)

### Attach `dlv` debugger

```bash
make debug-backend
# Then in VSCode: "Attach to running Go process" → host=localhost, port=2345.
```

Or remote:

```bash
dlv connect <pod-ip>:2345
```

In production, `dlv` is **not** installed in the distroless image. For prod debugging:

1. Reproduce locally with same env.
2. Use OTel traces via Tempo to localise.
3. Use Prometheus + Grafana to localise.
4. Last resort: temporary debug pod with `kubectl debug` (PR-required).

### Read zap JSON logs

Pretty-print:

```bash
docker compose logs -f greenmetrics-backend | jq 'select(.level=="ERROR")'
```

Filter by request:

```bash
kubectl logs -n greenmetrics deploy/greenmetrics-backend | jq 'select(.request_id=="abc123")'
```

Filter by tenant:

```bash
kubectl logs -n greenmetrics deploy/greenmetrics-backend | jq 'select(.tenant_id=="<uuid>")'
```

Loki from Grafana:

```logql
{app="greenmetrics-backend", env="production"} | json | tenant_id="<uuid>" | level="error"
```

### Trace ↔ log correlation

Every log line contains `trace_id` and `span_id`. Click the `trace_id` in Grafana Loki panel to jump to Tempo.

### Query Prometheus

```promql
# p99 ingest latency
histogram_quantile(0.99, sum by (le) (rate(gm_http_request_duration_seconds_bucket{path="/api/v1/readings/ingest"}[5m])))

# Ingest queue depth
gm_ingest_queue_depth

# Breaker state per host
gm_breaker_state

# Top tenants by ingest rate
topk(10, sum by (tenant_id) (rate(gm_ingest_readings_total[5m])))
```

### Inspect TimescaleDB

```bash
docker compose exec greenmetrics-timescaledb psql -U greenmetrics -d greenmetrics

# Hypertable info
\d+ readings
SELECT * FROM timescaledb_information.hypertables;

# CAGG status
SELECT view_name, materialization_hypertable_name, refresh_lag
FROM timescaledb_information.continuous_aggregates;

# Chunk count
SELECT hypertable_name, count(*) chunks
FROM timescaledb_information.chunks
GROUP BY hypertable_name;

# Compression status
SELECT chunk_name, before_compression_total_bytes, after_compression_total_bytes
FROM chunk_compression_stats('readings');
```

### Inspect pgx pool

```promql
gm_db_pool_acquire_duration_seconds
```

If acquire latency > 50 ms p99 → pool saturated; check `pg_stat_activity`:

```sql
SELECT count(*), state FROM pg_stat_activity WHERE usename='greenmetrics' GROUP BY state;
```

### Inspect ingestor

```promql
gm_ingest_readings_total{protocol="modbus"}
gm_ingest_dropped_total
gm_ingest_queue_depth
```

If queue depth saturates (`> 8000`) sustained 2 min → AlertManager fires; runbook `docs/runbooks/pulse-webhook-flood.md`.

### Tail Falco events

```bash
kubectl logs -n falco -l app=falco | jq 'select(.priority=="Critical" or .priority=="Error")'
```

## 2. Frontend (SvelteKit)

### Hot-reload dev

```bash
cd frontend && npm run dev
```

Vite serves on <http://localhost:5173> (dev) or <http://localhost:3005> (compose).

### Browser devtools

- Network tab: every request carries `X-Request-ID` (matches log/trace).
- Console: SvelteKit dev errors are full-stack.

### Type-check

```bash
cd frontend && npm run check
```

### Inspect API client errors

`src/lib/api.ts` `APIError` has `.status`, `.problem` (RFC 7807 body), `.requestId`. Log it:

```ts
try { await api.post('/v1/readings/ingest', body); }
catch (e) {
  if (e instanceof APIError) {
    console.error('api error', e.status, e.requestId, e.problem);
  }
  throw e;
}
```

## 3. Auth

### Decode JWT

```bash
ACCESS=$(curl -s http://localhost:8082/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"operator@greenmetrics.local","password":"long-enough-pw"}' \
  | jq -r .access_token)

echo "$ACCESS" | cut -d. -f2 | base64 -d 2>/dev/null | jq .
```

### Test RBAC denial

```bash
curl -X DELETE http://localhost:8082/api/v1/meters/<id> \
  -H "Authorization: Bearer $VIEWER_TOKEN"
# Expect 403 RFC 7807.
```

### Test RLS isolation

In `psql`:

```sql
SET LOCAL app.tenant_id = 'tenant-A-uuid';
SELECT count(*) FROM meters; -- only tenant A's rows

SET LOCAL app.tenant_id = 'tenant-B-uuid';
SELECT count(*) FROM meters; -- only tenant B's rows

RESET app.tenant_id; -- back to BYPASSRLS for migration_user only
```

## 4. Async jobs (Asynq, S4)

```bash
asynq stats --redis "$REDIS_URL"
asynq queues
asynq job ls --state retry
```

## 5. K8s

```bash
# Pod status
kubectl get pods -n greenmetrics
kubectl describe pod -n greenmetrics <pod>

# Logs
kubectl logs -n greenmetrics deploy/greenmetrics-backend --tail=100 -f

# Exec into a pod (only if shell available — distroless has none)
kubectl debug -n greenmetrics <pod> --image=busybox --target=greenmetrics-backend

# Events
kubectl get events -n greenmetrics --sort-by=.metadata.creationTimestamp

# NetworkPolicy denial
kubectl get networkpolicies -n greenmetrics
# Cilium / Calico flow log: see Loki query in §1
```

## 6. Argo CD

```bash
argocd app get greenmetrics
argocd app diff greenmetrics
argocd app sync greenmetrics --dry-run
```

## 7. Cosign verify

See `docs/SUPPLY-CHAIN.md` §7.
