# 0023 — Pack-contract interfaces and versioning

**Status:** Accepted
**Date:** 2026-04-30
**Authors:** @ciupsciups
**Doctrine refs:** Rules 14, 32, 34, 70, 71, 86, 87, 109, 129, 130.
**Charter ref:** `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2 (Pack flavours), §4 (Pack contract — the formal seam).
**Plan ref:** `docs/PLAN.md` §4.2 (Pack-contract interface formalisation), §5.3 (Sprint S5 deliverable).
**Review date:** 2026-10-30

## Context

The Charter introduces five Pack flavours (protocol, factor, report, identity, region) each implementing a typed Go interface that defines its contract with Core. ADR-0021 adopted the doctrine and plan; this ADR formalises the five interfaces themselves.

Today's codebase contains:

- An informal `Ingestor` interface used by `IngestorRunner` / AB-01 (`internal/services/ingestor_runner.go`).
- No `Builder` interface — report builders are concrete methods on `internal/services/report_generator.go` (525 LoC, REJ-11).
- No `FactorSource` interface — factor lookup is inlined in `internal/services/carbon_calculator.go`.
- No `IdentityProvider` interface — the local-DB authenticator is the only path.
- No `RegionProfile` interface.

The Pack-loader (`internal/packs/`) shipped in PR #1 of Sprint S5 expects these interfaces as the *typed* surface a Pack registers via `Registrar`. Today's `Registrar` accepts `any` for each kind because the typed interfaces don't yet exist. PR #2 of Sprint S5 ships them.

## Decision

Add five Pack-contract interfaces under `internal/domain/<kind>/`:

- `internal/domain/protocol/ingestor.go` — `Ingestor` interface with `Name`, `Start(ctx, sink)`, `Stop(ctx)`. Plus `ReadingSink` and `ReadingBatch` types.
- `internal/domain/reporting/builder.go` — `Builder` interface with `Type()`, `Build(ctx, period, factors, readings) (Report, error)`. Plus `Period`, `Report`, `ReportType` types.
- `internal/domain/emissions/factor_source.go` — `FactorSource` interface with `Name`, `Refresh(ctx)`. Plus `Factor`, `FactorBundle` types.
- `internal/domain/identity/provider.go` — `IdentityProvider` interface with `Name`, `Authenticate(ctx, credentials) (Identity, error)`, `LookupUser(ctx, id) (User, error)`. Plus `Identity`, `User`, `Credentials` types.
- `internal/domain/region/profile.go` — `RegionProfile` interface with `Code`, `Timezone`, `Locale`, `Currency`, `HolidayCalendar(year)`, `RegulatoryProfile()`. Plus `Profile`, `RegulatoryRegime` types.

Each interface carries:

- Full Godoc explaining the contract and the Pack-kind that implements it.
- A minimal compliant implementation example in `_test.go`.
- A version constant `ContractVersion = "1.0.0"` per Rule 71.
- A doc-comment cross-reference to `Pack.Register(reg)` in `internal/packs/pack.go`.

## Alternatives considered

- **One interface per concrete Pack id (e.g. `ModbusIngestor`, `MBusIngestor`).** Rejected because it defeats the entire Pack model — Core would couple to every concrete Pack type. The kind-level interface is the right abstraction (Rule 87 acceptance criterion: a new Pack of an existing kind doesn't change Core).
- **Single `Pack` interface with a tagged-union return.** Rejected because a tagged union loses compile-time type safety — registering an Ingestor as a Builder would compile and crash at runtime. Five typed interfaces + five typed registration methods catch the mistake at compile time.
- **Generic `Pack[Contract any]` interface.** Rejected because Go generics on Pack contracts produce hard-to-read errors when a Pack mismatches its declared kind. Five concrete interfaces are clearer.
- **Interface registration via reflection.** REJ-style rejection (REJ-04 generic config-management framework, similar pattern). Reflection registration is a runtime-failure path with no compile-time check.
- **Interfaces in `internal/packs/` rather than `internal/domain/<kind>/`.** Rejected because Pack-contract interfaces are *domain* interfaces (per Rule 32 DDD); the `internal/packs/` package is the loader infrastructure, not the contract.

## Consequences

### Positive

- Compile-time type safety on Pack registration — a Builder-registering-as-Ingestor mistake produces a build error.
- Pack-kind contracts can evolve independently per Rule 71 — each carries its own `ContractVersion`.
- The five-interface surface is the documentation of Pack capabilities; reading the package doc-comments tells you what a Pack of each kind must satisfy.
- Future Pack kinds (a hypothetical "Pricing" Pack or "Notification" Pack) follow the same template.

### Negative

- Five new packages in `internal/domain/` to maintain.
- The `Registrar` indirection in `internal/packs/pack.go` currently uses `any` — it tightens to typed parameters in PR #2 of Sprint S5 once the interfaces exist (a coupled change).
- Pack-contract interface evolution must be additive between minor Core versions per Rule 78 — adds friction to interface design.

### Neutral

- The interfaces are small (3–6 methods each) — Go's "interfaces should be small" idiom holds.
- Pack-contract version constants live alongside the interface — drift between version and surface is local.

## Residual risks

- A Pack-contract interface change between Core versions triggers a Pack-contract version bump; Packs in the field may need to update. Mitigated by the supported window in `internal/packs/contracts.go` and the deprecation grace period.
- The `Registrar` interface in `internal/packs/pack.go` continues to use `any` if PR #2 delivery slips; mitigated by the Sprint S5 exit gate.
- Interface design errors (e.g. missing the `Refresh(ctx)` knob on FactorSource) are caught only after the first non-Italian Factor Pack is built. Mitigated by the Pack-acceptance review of Sprint S6.

## References

- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3, §4.
- Doctrine: `docs/DOCTRINE.md` Rules 70, 71, 86, 87.
- Predecessor: `internal/services/ingestor.go` (informal Ingestor) — gets superseded by `internal/domain/protocol/ingestor.go`.
- Plan: `docs/PLAN.md` §4.2, §5.3.
- Adjacent ADR: ADR-0021 (charter and doctrine adoption).
- Future ADRs: ADR-0024 (Italian Region Pack), ADR-0025 (decomposition of report_generator.go), per-Pack acceptance ADRs.

## Tradeoff Stanza

- **Solves:** the absence of typed Pack-contract surfaces; the inability to register a Pack with compile-time type safety; the diffuse domain-logic placement (currently under `internal/services/`); the lack of Pack-kind-versioning that lets contracts evolve independently of Core.
- **Optimises for:** compile-time safety, Pack-acceptance reviewability (each Pack kind has one interface to satisfy), Pack-contract versioning (each kind versions independently), domain-driven structure (Rule 32: domain types in `internal/domain/`).
- **Sacrifices:** five new packages to maintain; the Registrar's typed-parameter signature must change in lock-step with this ADR; interface-design errors are caught only at the first non-flagship Pack of a kind; Go-version-matrix CI must run against any future Go-major-version bump.
- **Residual risks:** Pack-contract interface design errors surface late (mitigated by the Sprint S6 Italian-Pack-extraction work which exercises every interface); Pack-contract version bumps cascade into ecosystem Pack updates (mitigated by the supported-window discipline in `internal/packs/contracts.go`); the `Registrar` typed-parameter migration is a coupled change that must land in the same PR (mitigated by PR-scope discipline).
