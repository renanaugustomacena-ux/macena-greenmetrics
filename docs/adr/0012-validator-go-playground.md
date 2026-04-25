# 0012 — Validator: go-playground/validator

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 14 (contract first), 34 (backend contracts), 39 (security as core)
**Review date:** 2027-04-25

## Context

Backend handlers today use manual `BodyParser + nil checks` (`backend/internal/handlers/auth.go:49-55`, `readings.go:46-48`). No central validation; each handler re-implements its own checks. This violates Rule 14 (contracts implicit) and Rule 39 (security at boundary inconsistent).

## Decision

Adopt `github.com/go-playground/validator/v10`. Wire a single `Validator` instance constructed in `internal/api/v1/validate.go` with custom validators registered: `iso4217`, `rfc3339utc`, `uuidv4`, `tenant_id`, `pod_code` (regex `IT001E[0-9A-Z]{8}`), `pdr_code` (14 digits), `cf` (codice fiscale), `piva` (partita IVA). Per-handler request structs use `validate:"..."` tags. Centralised binding helper `func Bind[T any](c *fiber.Ctx, dst *T) error` in `internal/api/v1/bind.go` does `BodyParser` + `Validate` + RFC 7807 error build, replacing every manual `BodyParser` call.

## Alternatives considered

- **`github.com/go-ozzo/ozzo-validation`.** Rejected — ergonomics are method-based not tag-based; couples validation logic into the type, harder to override per-endpoint, performance is lower at high RPS. Tag-based ergonomics matter for the oapi-codegen-generated DTOs (ADR-013).
- **Hand-rolled validation in each handler.** Rejected — current state; violates Rule 14 and Rule 39.
- **`github.com/asaskevich/govalidator`.** Rejected — community has migrated to go-playground; smaller ecosystem.
- **`mold` + `valgo`.** Rejected — niche, less integration with Fiber.

## Consequences

### Positive

- Every request body validated at the boundary (Rule 39); CI lint can grep for any handler missing `Bind[T]` call.
- Custom validators encode Italian-specific schemas (POD, PDR, CF, P.IVA) in one place.
- Generated OpenAPI types from oapi-codegen carry `validate:` tags out of the box; types stay in sync with contract.
- RFC 7807 error response is uniform — `Bind[T]` builds the Problem body.

### Negative

- Reflection-based validation has a small per-request cost (~10–30 µs per medium struct); negligible at our SLO.
- Validation tags are stringly typed; typos found at runtime (mitigated by table-driven tests in `tests/property/`).

### Neutral

- One more direct Go dep; well-maintained, large user base.

## Residual risks

- Custom validator regex mistakes (e.g. POD code regex too strict / too loose) — covered by table-driven tests in `internal/api/v1/validate_test.go`.
- Reflection allocation under high RPS — bench in `tests/bench/validator_bench_test.go`; if ever a hot path, pre-compile validators per type via `validator.RegisterStructValidation`.

## References

- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §6.B.34.
- ADR-013 (oapi-codegen design-first).
- `https://github.com/go-playground/validator`.

## Tradeoff Stanza

- **Solves:** missing per-boundary validation; non-uniform error responses; Italian-specific schema duplication.
- **Optimises for:** tag ergonomics, oapi-codegen integration, RFC 7807 uniformity.
- **Sacrifices:** reflection cost (negligible at SLO), one more direct dep, runtime tag-typo risk.
- **Residual risks:** custom validator regex errors (table-driven tests cover); reflection cost at extreme RPS (bench gates).
