# API Versioning Policy

**Doctrine refs:** Rule 21 (evolution + change management), Rule 34 (backend contracts).
**ADR:** `docs/adr/0008-api-versioning-policy.md` (to author S3).

## 1. Versioning model

- **Major version in URL path.** `/api/v1/`, `/api/v2/`. v1 → v2 is a breaking change.
- **Minor + patch in `info.version` of `api/openapi/v1.yaml` and in `Server: greenmetrics-backend/X.Y.Z` response header.**
- **SemVer 2.0.0 semantics.** Breaking → major. Additive → minor. Bug fix → patch.
- **CHANGELOG required for every minor + major bump.**

## 2. What is a breaking change

- Removing or renaming a path.
- Removing or renaming a request or response field.
- Tightening a validation rule (e.g. `min: 12` → `min: 16` on existing field).
- Changing the type of a field.
- Changing the URL of a status code (e.g. `404 NotFound` → `410 Gone` for the same scenario).
- Removing a query parameter that clients may rely on.
- Adding a required request field.

## 3. What is additive (minor)

- Adding a new path.
- Adding an optional request field.
- Adding a response field (clients are required to ignore unknown fields).
- Adding a new enum value (clients should treat unknown values as "other"; documented in OpenAPI `x-extensible-enum`).
- Loosening a validation rule (e.g. `min: 12` → `min: 8`).

## 4. Deprecation flow

1. Mark the endpoint or field `deprecated: true` in `api/openapi/v1.yaml`.
2. Add `Sunset:` header per RFC 8594 with date ≥ 90 days out.
3. Add `Deprecation:` header per draft-ietf-httpapi-deprecation-header.
4. CHANGELOG entry under `Deprecated` section.
5. PR template's "Backend addendum" must check the deprecation box.
6. Client teams notified via `#greenmetrics-platform` Slack channel + email to known integrators.
7. After Sunset date: route returns `410 Gone` for ≥ 30 days, then removed in next major.

## 5. Major bump procedure

- New tree under `internal/api/v2/` parallel to v1.
- Both routes registered (`/api/v1` + `/api/v2`) for at least 2 minor versions.
- v1 path returns `Sunset:` header.
- After 6-month parallel-run: ADR to remove v1; CHANGELOG `Removed` entry.

## 6. Backward-compatibility test

`tests/contracts/v1_compat_test.go`:

- Loads previous-release `api/openapi/v1.yaml` (committed as `tests/contracts/golden/v1-prev.yaml`).
- For each operation, generates a sample request from the schema.
- Runs request against current handlers.
- Validates response against the previous-release response schema.
- Fails on any breaking change not accompanied by major bump.

## 7. Server header

Every response carries `Server: greenmetrics-backend/X.Y.Z` derived from `internal/version.Version`. Clients can pin minimum version via header check.

## 8. Anti-patterns rejected

- Silent breaking changes ("it's just a rename").
- Removing a deprecated endpoint without Sunset notice.
- Multiple major versions diverging in feature set indefinitely.
- "Mobile clients can't handle versioning" — they must.

## 9. Exceptions

- Security fixes that require a tightening of validation may be released as patch with a CHANGELOG `Security` entry, no Sunset window. Document in ADR.
- Regulatory changes (e.g. GDPR DSAR enforcement) may be released as required additive minor with explicit notification.
