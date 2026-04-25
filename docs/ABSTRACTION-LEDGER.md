# Abstraction Ledger

**Doctrine refs:** Rule 13 (abstraction as cost centre), Rule 47 (decision rationale).

Every major abstraction in the codebase is an entry here. Each row carries: hidden complexity, lost flexibility, who pays the cost, trigger to remove.

## 1. Active abstractions

| ID | Abstraction | Hidden complexity | Lost flexibility | Who pays | Trigger to remove |
|---|---|---|---|---|---|
| AB-01 | `IngestorRunner` (`internal/services/ingestor_runner.go`) | per-protocol polymorphism, lifecycle management | per-protocol fine-tuning happens via interface options; no protocol-specific shortcut | platform-team (orchestration) | only one protocol remains in active use for >2 quarters |
| AB-02 | Embedded OpenAPI string (`internal/handlers/openapi.go`) | spec lives inside Go binary | cannot diff/lint as text without extraction | platform-team | retiring in S2 — extracting to `api/openapi/v1.yaml` |
| AB-03 | Single `TimescaleRepository` struct (`internal/repository/timescale_repository.go` 298 LoC) | one struct holds all DB access | no per-aggregate visibility | data-team | already triggered: decompose in S3 |
| AB-04 | RFC 7807 helpers (`internal/handlers/errors.go`) | uniform error shape | new error fields require helper changes | platform-team | invariant — keep |
| AB-05 | Zap field redactor (planned S3) | log redaction logic | depends on field name list | secops | invariant — keep |
| AB-06 | `Bind[T]` request binding (planned S2) | reflection on tags | per-handler escape hatches forbidden | app-team | only one body parser ever used → fold into helper |
| AB-07 | Per-host circuit breakers (planned S4) | shared breaker state across goroutines | per-call breaker disabled-by-default | platform-team | external services consolidate to one + breaker is overhead |
| AB-08 | Bounded ingest channel + batched writer (planned S4) | indirection between source and DB | source cannot push directly | platform-team | ingest rates drop below threshold sustained |
| AB-09 | Worker pool via `panjf2000/ants` (planned S4) | bounded concurrency | new ingestor cannot spawn unbounded workers | platform-team | GOMAXPROCS-based naked goroutines suffice |
| AB-10 | RBAC permission registry (planned S3) | role→perms map indirection | per-handler ad-hoc auth forbidden | secops | invariant — keep |
| AB-11 | Postgres RLS + tenant context wrapper (planned S3) | implicit `app.tenant_id` GUC | every Tx requires `InTxAsTenant` | platform-team | invariant — keep |
| AB-12 | ESO ClusterSecretStore (planned S3) | secret materialisation indirection | direct K8s Secret edits forbidden | secops | switch to alternative provider — abstraction stays, ref point changes |
| AB-13 | Cosign verify at admission (planned S3) | signature check per pod start | unsigned pods rejected | secops | invariant — keep |
| AB-14 | Argo CD App-of-Apps (planned S4) | GitOps reconciliation indirection | direct `kubectl apply` reverted | platform-team | switch to alternative GitOps engine — abstraction stays |
| AB-15 | Kustomize overlays (planned S4) | env-specific patches | per-env divergence visible | platform-team | overlay count > 5, switch to Helm chart values |
| AB-16 | Asynq + Redis worker queue (planned S4) | job queueing layer | sync handlers forbidden for long-running | platform-team | report generation re-becomes synchronous (would require SLO renegotiation) |

## 2. Rejected abstractions

| Pattern | Why not |
|---|---|
| Generic `Repository[T]` over pgx | Rule 13 — cost > leverage; pgx is already typed-enough |
| Generic `Pipeline[A, B]` for ingest | Rule 26 — only one consumer; YAGNI |
| Plugin system via `plugin.Open` | Security + ops complexity |
| Service mesh for 2 services | REJ-01 |
| Generic config-management framework | REJ-04 |
| OPA in request path | REJ-10 |
| ORM (gorm/bun) | REJ-35 |

## 3. Quarterly review process

Each quarterly platform office hours:

1. Re-read every active abstraction.
2. For each: any new evidence of cost > leverage?
3. For each rejected pattern: any new evidence to flip?
4. New abstractions added in the quarter: did they justify the cost in their first 90 days?
5. Outcomes recorded in `docs/office-hours/YYYY-MM-DD.md`.

## 4. Adding a new abstraction

PR introducing a new abstraction must:

- Add a row here.
- Link an ADR.
- ADR ends with the four-part Tradeoff Stanza (Rule 27).
- Justify "Cost of this abstraction" paragraph naming hidden complexity, lost flexibility, who pays.

## 5. Removing an abstraction

PR removing an active abstraction must:

- Update the row to `Removed` and date.
- Link an ADR explaining the trigger.
- Move the entry to "Removed" history at the bottom of this file.
- Refactor the call sites to the concrete type.

## 6. Removed history

(empty — populates as abstractions retire)
