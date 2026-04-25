# 0013 — OpenAPI codegen: design-first via oapi-codegen

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 12 (platform thinking), 14 (contract first), 21 (evolution), 34 (backend contracts)
**Review date:** 2027-04-25

## Context

Today the OpenAPI spec lives as an embedded JSON string in `backend/internal/handlers/openapi.go:50-172` — reverse-engineered from handler code via swag annotations. The spec is derived from code, not the other way around. This violates Rule 14 (contract first) — the contract is implicit and any handler change can silently drift the contract.

## Decision

Flip to **design-first**. Source of truth becomes `api/openapi/v1.yaml` (hand-written, semver tagged). Adopt `github.com/oapi-codegen/oapi-codegen/v2` to generate `internal/api/v1/types.go` (request/response structs with `validate:` tags) and `internal/api/v1/server.go` (Fiber-compatible router shim). Retire swag annotations or keep `internal/handlers/openapi.go` only as a tiny `//go:embed` shim that serves the canonical YAML through Swagger UI.

Wire `cmd/oapi-gen/main.go` invoking oapi-codegen with config:

```yaml
package: v1
output: internal/api/v1/types.go
generate:
  models: true
  fiber-server: true
  strict-server: true
output-options:
  skip-prune: true
```

CI gate `openapi-lint` runs `redocly lint api/openapi/v1.yaml`. Backward compatibility test `tests/contracts/v1_compat_test.go` runs the previous-release OpenAPI against current handlers via `kin-openapi` validator — fail on breaking change.

## Alternatives considered

- **Continue with swag (annotation-driven).** Rejected — see Context; Rule 14 violation.
- **Generate the YAML from Go types via swag, then treat that as source of truth.** Rejected — same problem; the Go types are still authoritative, the YAML is a downstream artefact.
- **`ogen-go/ogen`.** Rejected — younger ecosystem, less Fiber integration; revisit in 2027 if oapi-codegen becomes unmaintained.
- **Buf + Connect (gRPC + REST).** Rejected — overkill for our REST-only API; would be reconsidered if internal services proliferate.
- **Hand-written types + hand-written YAML.** Rejected — drift risk too high; codegen is the gate.

## Consequences

### Positive

- Contract is the source of truth (Rule 14); Go types regenerated, never the reverse.
- Generated types carry `validate:` tags that align with ADR-012 (validator).
- Breaking-change protection via `tests/contracts/v1_compat_test.go`.
- OpenAPI document is human-editable in YAML; `redocly lint` catches schema bugs at PR time.
- `Sunset:` headers and deprecation policy can be expressed declaratively in OpenAPI `deprecated: true`.

### Negative

- Migration is invasive — every handler signature changes to use generated types.
- Stage cutover required: keep swag during one release; gate on contract-test green.
- oapi-codegen generates a Fiber adapter that wraps Fiber's `*fiber.Ctx` — reviewer learning curve.

### Neutral

- Swagger UI shell (`internal/handlers/openapi.go`'s `swaggerUIHTML`) keeps working — only the spec source changes.
- One more cmd/ binary (`cmd/oapi-gen/main.go`).

## Residual risks

- Generated code may be regenerated with diff noise on each oapi-codegen bump; commit the generated files and gate with a CI check that `make openapi-bundle` produces no diff vs committed.
- v1 contract test relies on `kin-openapi` validator — if validator has bugs, breaking changes can slip; mitigated by property tests on response shapes.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.A.14, §6.B.34.
- `api/openapi/v1.yaml` — canonical spec (lands S2).
- `https://github.com/oapi-codegen/oapi-codegen`.
- ADR-012 (validator).
- ADR-008 (API versioning, S3).

## Tradeoff Stanza

- **Solves:** contract drift; missing per-boundary validation; reverse-engineered spec.
- **Optimises for:** contract-as-source-of-truth, breaking-change protection, Italian-specific schema reuse.
- **Sacrifices:** invasive handler refactor; staged cutover effort; one more codegen step.
- **Residual risks:** codegen diff noise (committed + CI-checked); validator bugs in v1 compat (property tests supplement).
