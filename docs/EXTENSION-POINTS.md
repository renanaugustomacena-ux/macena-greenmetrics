# Extension Points

**Doctrine refs:** Rule 12 (controlled extensibility), Rule 13 (abstraction cost), Rule 26 (rejection authority).

An extension point is a documented seam where a new implementation can be added without modifying the surrounding code. Adding an extension point is not free — it is a Rule-13 cost. We add one only when ≥ 2 implementations already exist or are in flight.

## 1. Inventory

| ID | Seam | Interface | Existing impls | Owner |
|---|---|---|---|---|
| EP-01 | Ingestor protocol | `internal/services/ingestor.Ingestor` (formalise in S2) | Modbus TCP, Modbus RTU, M-Bus, SunSpec, Pulse, OCPP | app-team |
| EP-02 | Report dossier builder | `internal/domain/reporting.Builder` (formalise in S3) | ESRS E1, Piano 5.0, Conto Termico, TEE, audit DLgs102, monthly consumption, CO2 footprint | app-team |
| EP-03 | Emission factor source | `internal/domain/emissions.FactorSource` (formalise in S3) | ISPRA, Terna, AIB, GHG Protocol (placeholder), EcoInvent (placeholder) | app-team |
| EP-04 | Alert detector | `internal/domain/alerting.Detector` (formalise in S4) | threshold (existing), z-score (S4) | app-team |
| EP-05 | Async job handler | `internal/jobs.JobHandler` (formalise in S4) | per `report:*` job type | app-team |
| EP-06 | Repository (data store) | `internal/repository.<aggregate>Repo` (per-aggregate) | TimescaleDB | platform-team |

## 2. Contract template

Every extension point has:

- A Go interface in `internal/domain/<aggregate>/`.
- A registration mechanism in `cmd/server/main.go` or in the per-package `Init()`.
- A test that ≥ 2 implementations satisfy the interface (CI gate via build tag).
- An ADR explaining the seam and the cost.

## 3. EP-01 Ingestor protocol

Goal: add a new meter protocol (e.g. SCADA-OPC-UA, BACnet) without touching existing ingestors or the API surface.

```go
// internal/services/ingestor.go (formalise in S2)
type Ingestor interface {
    Name() string
    Start(ctx context.Context, sink ReadingSink) error
    Stop(ctx context.Context) error
}

type ReadingSink interface {
    Send(ctx context.Context, batch ReadingBatch) error
}
```

Registration in `cmd/server/main.go`:

```go
runner.Register(modbus.New(cfg))
runner.Register(mbus.New(cfg))
runner.Register(sunspec.New(cfg))
runner.Register(pulse.New(cfg))
```

CI gate: `tests/extension/ingestor_count_test.go` asserts at least 2 registered implementations exist.

## 4. EP-02 Report dossier builder

Goal: add a new regulatory dossier (e.g. ISO 14064-1) without touching the report orchestrator.

```go
// internal/domain/reporting/builder.go (formalise in S3)
type Builder interface {
    Type() ReportType
    Build(ctx context.Context, period Period, factors emissions.FactorBundle, readings []readings.Aggregated) (Report, error)
}
```

Registration:

```go
reporting.Register(esrs_e1.New())
reporting.Register(piano_5_0.New())
// ...
```

## 5. EP-03 Emission factor source

Goal: add a new factor provider (e.g. EcoInvent for Scope 3 LCA) without touching the calculator.

```go
// internal/domain/emissions/factor_source.go
type FactorSource interface {
    Name() string
    Refresh(ctx context.Context) ([]Factor, error)
}
```

Registration: each source registers itself + a fallback chain ordered by `valid_from` recency.

## 6. EP-04 Alert detector

Goal: add a new anomaly type (e.g. seasonal-decomposition) without touching the alert engine.

```go
// internal/domain/alerting/detector.go
type Detector interface {
    Name() string
    Detect(ctx context.Context, window []Reading) ([]Alert, error)
}
```

## 7. EP-05 Async job handler

Goal: add a new background job type without touching the worker bootstrap.

```go
// internal/jobs/handler.go
type JobHandler interface {
    Type() string
    Handle(ctx context.Context, payload []byte) error
}
```

Registration via Asynq mux.

## 8. EP-06 Repository

Goal: per-aggregate repository so a future swap (e.g. read replica routing) is local. **Not** a generic `Repository[T]` framework — that would be over-abstraction (Rule 13).

```go
// internal/repository/readings_repo.go
type ReadingsRepo interface {
    InsertBatch(ctx context.Context, tenantID uuid.UUID, batch []readings.Reading) (int, int, error)
    QueryAggregated(ctx context.Context, tenantID uuid.UUID, q AggregatedQuery) ([]readings.Aggregated, error)
}
```

## 9. Anti-patterns rejected

- Generic `MeterReadingPipeline[T]` framework while only one consumer exists (Rule 13, REJ via Rule 26).
- Plugin loading via `plugin.Open` — security risk + complexity (REJ).
- Hot-reload of registered impls — adds churn, no value (REJ).

## 10. Sunset of an extension point

If only one implementation remains for >2 quarters, ADR to remove the seam (Rule 13 — pay-as-you-use abstractions). The implementation collapses back into a concrete type.
