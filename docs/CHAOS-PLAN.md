# Chaos Engineering Plan

**Doctrine refs:** Rule 15, Rule 36, Rule 59, Rule 64.
**Owner:** `@greenmetrics/sre`, `@greenmetrics/secops`.
**Cadence:** monthly Game Day (last Friday). Recorded in `docs/CHAOS-LOG.md`.
**Tooling:** Chaos Mesh deployed via Helm; experiments live in `chaos/`.

## 1. Experiments

### CE-01 — DB pod kill

```yaml
# chaos/db-kill.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata: { name: db-pod-kill, namespace: greenmetrics-staging }
spec:
  action: pod-kill
  mode: one
  selector: { namespaces: [greenmetrics-staging], labelSelectors: { app: timescaledb } }
```

**Validation:** RDS Multi-AZ failover < 60s; `gm_http_requests_total{status=~"5.."}` ratio < 1%; on-call paged within 60s.

### CE-02 — Network partition (API ↔ DB)

```yaml
# chaos/network-partition.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata: { name: api-db-partition, namespace: greenmetrics-staging }
spec:
  action: partition
  mode: all
  selector: { namespaces: [greenmetrics-staging], labelSelectors: { app: greenmetrics-backend } }
  direction: to
  target:
    mode: one
    selector: { namespaces: [greenmetrics-staging], labelSelectors: { app: timescaledb } }
  duration: 2m
```

**Validation:** Backend `/api/ready` flips degraded; pgx pool acquire fails; recovery < 30s after partition lifts.

### CE-03 — Modbus simulator crash

```yaml
# chaos/ingestor-crash.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata: { name: simulator-crash, namespace: greenmetrics-staging }
spec:
  action: pod-failure
  mode: all
  selector: { namespaces: [greenmetrics-staging], labelSelectors: { app: greenmetrics-simulator } }
  duration: 5m
```

**Validation:** Per-host breaker opens within 5 calls; degradation banner; `ModbusIngestStalled` alert fires; recovery within 30s of simulator restart.

### CE-04 — CPU pressure on backend

```yaml
# chaos/cpu-pressure.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: StressChaos
metadata: { name: api-cpu-stress, namespace: greenmetrics-staging }
spec:
  mode: one
  selector: { namespaces: [greenmetrics-staging], labelSelectors: { app: greenmetrics-backend } }
  stressors: { cpu: { workers: 4, load: 90 } }
  duration: 5m
```

**Validation:** HPA scales up; SLO p99 holds within budget; pgx pool not saturated.

### CE-05 — Redis down (Asynq)

```yaml
# chaos/redis-down.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata: { name: redis-kill, namespace: greenmetrics-staging }
spec:
  action: pod-kill
  mode: all
  selector: { namespaces: [greenmetrics-staging], labelSelectors: { app: redis } }
```

**Validation:** `POST /v1/reports` returns 503 Retry-After; ingest path unaffected; in-flight jobs retry on resume.

### CE-06 — ISPRA upstream block (egress NetworkPolicy)

```yaml
# chaos/ispra-blocked.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata: { name: ispra-blocked, namespace: greenmetrics-staging }
spec:
  action: partition
  mode: all
  selector: { namespaces: [greenmetrics-staging], labelSelectors: { app: greenmetrics-backend } }
  direction: to
  externalTargets: ["www.isprambiente.gov.it"]
  duration: 5m
```

**Validation:** Breaker opens; reports stamped `data_freshness:"cached_<n>h"`; `gm_external_api_fallback_total{api="ispra"}` increments.

### CE-07 — Disk pressure on TimescaleDB

(Manual; via RDS CloudWatch alarms simulated.)

**Validation:** Alert fires at 80% / 90%; ingest path tightens rate limit; runbook `capacity-spike.md` is followed.

## 2. Schedule

- Last Friday of each month, 10:00 Europe/Rome.
- 1 experiment per month, rotated; revisit each every 6 months.
- Run against **staging only** in steady state. Production drills are full DR drills (annual; `region-failover.md`).

## 3. Pre-flight

- Stakeholders informed in `#greenmetrics-ops` 24h ahead.
- On-call primary on duty + dedicated for the drill window.
- Status page entry not required (staging).
- Rollback plan documented in the experiment YAML.

## 4. Outcome capture

For every experiment, document in `docs/CHAOS-LOG.md`:

- Date.
- Experiment ID.
- Pre-state.
- Trigger time.
- Detection time.
- Mitigation time.
- Recovery time.
- Outcome (Pass / Partial / Fail).
- Action items (with owners + dates).

## 5. Anti-patterns rejected

- Run chaos in production without DR drill cadence — REJ.
- Skip the runbook the experiment is meant to validate — defeats the point.
- Run multiple experiments simultaneously — confounds attribution.
- Run silently — stakeholders + on-call must know.
