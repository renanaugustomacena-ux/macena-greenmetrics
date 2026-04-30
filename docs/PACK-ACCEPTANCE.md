# Pack Acceptance Criteria

**Owner:** `@greenmetrics/platform-team`.
**Doctrine refs:** Rule 27 (Tradeoff Stanza), Rules 69–88 (Modular Template Integrity), Rule 131 (formal-spec validation), Rule 138 (annual Pack review).
**Charter refs:** §3.2 (what is a Pack), §4 (Pack contract), §12.2 (Pack-contract versions).
**Adopted:** 2026-04-30. **Review cadence:** quarterly.

This document is the gate every Pack passes before being loaded by Core in a production deployment. The Pack-loader at boot enforces a subset mechanically (manifest validation, signature verification, manifest-lock match); the rest is reviewer-signed at PR time.

A Pack that fails any criterion in §1 is rejected at boot. A Pack that fails a criterion in §2 may be accepted as **Beta** but not promoted to **GA** until the gap is closed (see §4 acceptance levels).

---

## 1. Mandatory criteria (boot-blocking)

The Pack-loader refuses to start the deployment if any of the following is false. These are mechanically enforced.

### 1.1 Manifest validity

The Pack ships `packs/<kind>/<id>/manifest.yaml` (or `.json`) that validates against `docs/contracts/pack-manifest.schema.json` (JSON Schema 2020-12).

The manifest declares: `id`, `kind` (`protocol` | `factor` | `report` | `identity` | `region`), `version` (SemVer), `min_core_version` (SemVer), `pack_contract_version` (SemVer), `author`, `license_spdx`, `capabilities[]`, `dependencies[]`, optional `notes`.

**Verification:** `task pack:validate` runs `internal/packs.Manifest.ValidateBasic()` against every loaded Pack at boot.

### 1.2 Pack-contract version supported by Core

The Pack's `pack_contract_version` falls inside Core's supported window declared in `backend/internal/packs/contracts.go`.

**Verification:** `task pack:contract-compliance` walks each Pack and checks the version against Core's window. A Pack-contract version inside the deprecated band emits a `pack-contract-deprecated` event in the audit log on every boot until upgraded; one outside the supported band fails boot.

### 1.3 Min-Core version satisfied

The Pack's `min_core_version` is `<=` Core's running version. A Pack that requires a future Core release fails boot.

**Verification:** boot-time check in `internal/packs/loader.go`.

### 1.4 Per-Pack-kind contract implementation

The Pack implements the canonical interface for its kind:

- `kind=protocol` → `internal/domain/protocol.Ingestor`.
- `kind=factor` → `internal/domain/emissions.FactorSource`.
- `kind=report` → `internal/domain/reporting.Builder`.
- `kind=identity` → `internal/domain/identity.IdentityProvider`.
- `kind=region` → `internal/domain/region.RegionProfile`.

**Verification:** Go's compile-time type system + `tests/packs/contract_compliance_test.go`.

### 1.5 Pack image signed (Cosign keyless)

The container image carrying the Pack is signed via Cosign keyless OIDC trust per Rule 57 / ADR-0017.

**Verification:** Kyverno admission policy `verify-images` denies the pod at admission time.

### 1.6 Manifest lock match

The runtime manifest set matches the signed `manifest.lock.json` produced at the most recent CI build per Rule 73. Boot fails on divergence (e.g., a hot-patched Pack at runtime with no provenance trail).

**Verification:** boot-time signature check in `internal/packs/lock.go`.

### 1.7 Required capabilities declared

The Pack declares its capabilities in `manifest.capabilities[]` (e.g. `protocol-modbus-tcp`, `factor-source-ispra`, `report-builder-csrd-esrs-e1`). Core matches required capabilities from `config/required-packs.yaml` against declared capabilities at boot per Rule 75. Missing capabilities fail boot; unannounced capabilities are ignored.

**Verification:** `task pack:list` enumerates declared vs required.

---

## 2. Promotion criteria (Beta → GA)

The following are required for a Pack to be considered **GA** (i.e., cleared for use in production engagements). Beta Packs may be loaded but only with an explicit `engagements/<id>/CORE-CUSTOMISATIONS.md` row (Rule 80) accepting the gaps with a sunset date.

### 2.1 CHARTER document

Each Pack ships `packs/<kind>/<id>/CHARTER.md` carrying:

- the Pack's purpose and scope;
- the four-part Tradeoff Stanza (Rule 27): Solves / Optimises for / Sacrifices / Residual risks;
- vendor / regulator compatibility matrix (Protocol / Factor / Report Packs);
- temporal-validity policy (Factor Packs reference Rule 90);
- the formal-spec reference for the regulatory output (Report Packs reference Rule 131 — EFRAG XBRL / GSE XSD / ENEA XSD / SEC inline-XBRL where applicable).

**Verification:** review-time gate; `tests/packs/charter_present_test.go` checks file existence.

### 2.2 Conformance test for the Pack contract

The Pack ships at least one `_test.go` file under `packs/<kind>/<id>/` exercising the contract methods. Protocol Packs ship a deterministic simulator + a test that drives the Ingestor against the simulator. Report Packs ship golden-fixture tests that assert byte-identical output (per Rule 89). Factor Packs ship temporal-validity tests against pinned reference data.

**Verification:** `task test:packs` runs every Pack's tests.

### 2.3 Property tests for algebraic invariants

Where applicable, property tests assert algebraic invariants (Rule 33):

- Factor Packs: temporal-validity is monotonic; querying at the period midpoint is deterministic across clock skew.
- Report Packs: Builder is a pure function of `(period, factors, readings)`; idempotent under repeated runs (byte-identical output per Rule 89).
- Protocol Packs: `Ingestor.Read()` is idempotent against a static device state.

**Verification:** `tests/property/packs_*.go`.

### 2.4 SBOM + vulnerability scan

The Pack image has an SBOM generated by Syft (Rule 58) and a Trivy scan with no HIGH / CRITICAL findings (Rule 59 SLA timelines apply). SLSA L2 provenance (Rule 57 / ADR-0018) attests the build.

**Verification:** CI pipeline gates `pack:build`.

### 2.5 Per-Pack health endpoint

The Pack's `Health(ctx)` returns within 2s and surfaces Pack-specific dependency state (e.g., a Factor Pack returns `{factor_source_age_days: 12, last_refresh_status: "ok"}`). Per Rule 74 the result is folded into Core's `/api/health.dependencies` envelope.

**Verification:** `tests/integration/health_envelope_test.go`.

### 2.6 Per-Pack panic recovery + resource budget

The Pack's `Init` and `Register` paths recover from panics into structured errors (Rule 76); long-lived goroutines respect `Shutdown(ctx)` within 30 seconds (Rule 42); worker-pool sizes and queue depths are bounded (Rule 41).

**Verification:** `tests/leak/packs_*.go` runs the Pack lifecycle and asserts no leaks.

### 2.7 Documentation: README + capabilities + dependencies

`packs/<kind>/<id>/README.md` explains: what the Pack does, who wrote it, which engagements use it, which capabilities it declares, which dependencies it requires (e.g., a Pulse Protocol Pack depends on Redis for replay storage), known limitations, vendor compatibility (where the Pack interfaces with hardware), regulatory updates the Pack tracks (where applicable).

**Verification:** `tests/packs/readme_present_test.go`.

### 2.8 Annual-review entry

Per Rule 138, each Pack gets an entry in the annual-review log at `packs/<kind>/<id>/REVIEW-LOG.md` (or aggregated at `docs/PACK-REVIEW-LOG.md`). Entries record: review date, reviewer, regulatory delta vs previous year (for Region / Factor / Report Packs), test-fixture refresh date, vendor-compatibility-matrix delta (for Protocol Packs).

**Verification:** annual review process; reviewer signs the log.

---

## 3. Per-kind specifics

### 3.1 Protocol Packs

- Wire-format invariants documented in CHARTER (e.g., Modbus register layouts, M-Bus baudrates, OCPP 2.0.1 message shape).
- Device-profile catalogue at `packs/protocol/<id>/devices/` covers at minimum the top-3 vendor models by deployment frequency.
- Latency budget per Rule 124 (Modbus ≤ 200ms, M-Bus ≤ 500ms, SunSpec ≤ 300ms, OCPP ≤ 1s tx, Pulse ≤ 100ms).
- Per-host circuit breaker per ADR-0009.
- 24-hour disk-backed buffer at the edge gateway (Rule 111).
- NTP-synced clock at the edge with optional GPS time stratum-1 (Rule 112).
- Per-meter HMAC reading provenance at ingestion (Rule 173, Phase F deliverable).

### 3.2 Factor Packs

- Temporal-validity-keyed factors: every factor row carries `valid_from`, `valid_to`, `source_url`, `source_published_at`, `source_hash`.
- Refresh path with cached fallback: if the upstream is down, the previous validity period continues to apply with a `data_freshness=stale` annotation.
- Authoritative-source manifest: ISPRA = National Inventory Report; GSE = AIB residual mix; Terna = daily national mix; AIB = European residual mix; UK DEFRA = GHG conversion factors; EPA eGRID = US grid mix.
- Pack image is reproducible: building the Pack twice with the same manifest produces byte-identical layers.

### 3.3 Report Packs

- Pure function: `Build(ctx, period, factors, readings) -> Output` (Rule 91).
- Byte-identical output across re-runs (Rule 89). Replay test at `tests/replay/<report-pack>_test.go`.
- Formal-spec validation against the regulator's published schema (Rule 131): EFRAG XBRL Taxonomy for ESRS E1; GSE XSD for Conto Termico; ENEA XSD for D.Lgs. 102/2014; etc.
- Provenance bundle: every output carries `provenance.json` with `period`, `factor_pack_versions`, `manifest_lock_hash`, `code_hash`, `report_pack_version`.
- Output signed at finalisation (Rule 144) with a Pack-specific signing key from the deployment's KMS.
- EGE counter-signature workflow declared in the CHARTER where applicable (audit 102/2014, Piano 5.0 above thresholds).

### 3.4 Identity Packs

- The Pack replaces the Core's local-DB identity provider behind a feature flag.
- Per-tenant IdP routing supported for partner-hosted multi-tenant deployments.
- JIT user provisioning + role mapping + per-IdP claim transformation.
- Validated against at least one upstream IdP (e.g., SAML against Keycloak running in CI; OIDC against Auth0 staging).
- Logout / session-revocation surface present (RBAC role revoke must invalidate cached sessions within 60s).

### 3.5 Region Packs

- Bundles: timezone, locale, currency, holiday calendar, privacy-regime overlay, regulatory-thresholds, default Pack matrix for the region (e.g. Italian Region Pack defaults to ISPRA + GSE + ESRS E1 + Piano 5.0 etc.).
- `bootstrap_tenant_id` materialised at first migration (Charter §11).
- Each Region Pack is reviewed against the Italian-flagship Pack for thoroughness (Rule 88).

---

## 4. Acceptance levels

| Level | Definition | Allowed in production engagements |
|---|---|---|
| **Draft** | All §1 mandatory criteria pass; some §2 promotion criteria fail. | No. Draft Packs ship in dev / sandbox only. |
| **Beta** | All §1 + §2 pass except for: missing annual review (§2.8) on a Pack younger than one year, or vendor compatibility limited to a single device. | Yes, with explicit `CORE-CUSTOMISATIONS.md` row + sunset date. |
| **GA** | All §1 + §2 pass; the Pack has at least one full annual review on record. | Yes, default. |
| **Deprecated** | Pack-contract version moved to deprecated band; replacement Pack named. | Yes through the deprecation window; emits `pack-contract-deprecated` events; removed at next major. |
| **Superseded** | A newer Pack replaces this one; this Pack is read-only and not loaded in new engagements. | Existing engagements may keep loading until next sync; new engagements use the successor. |

Acceptance level is recorded in the Pack's `manifest.yaml` under an `acceptance_level` field and surfaced in `task pack:list`.

---

## 5. Promotion process

A Pack moves between acceptance levels via a Pack-acceptance PR:

1. Author opens a PR titled `feat(pack): <pack-id> -> <new-level>` with the CHARTER, manifest, tests, README, and (where applicable) annual-review entry.
2. CI runs the §1 mandatory checks + the §2 promotion checks. Output is a per-criterion green/red report.
3. Reviewers (Macena platform-team + the relevant domain advisor — e.g., EGE for Report Packs, OT-integration lead for Protocol Packs) sign off after walking the criteria.
4. Promotion to GA also requires explicit Rule 26 / 46 / 66 rejection-authority signoff: a named reviewer accepts the Pack for production use.
5. The promotion is recorded in `docs/PACK-CATALOGUE.md` (Sprint S6 deliverable) with the date, the reviewer chain, and the acceptance level.

A Pack that fails promotion does not silently demote; the PR is closed with a list of remaining gaps and a reviewer-named follow-up issue.

---

## 6. Rejection criteria (the floor)

A Pack that exhibits any of the following is rejected without further review (Rule 26 / 46 / 66):

- Hard-codes `tenant_id`, `environment`, or any deployment-specific value.
- Calls global state (a service locator, a shared mutable singleton) to access Core surfaces. Pack contributions go through `Registrar` indirection per Rule 72.
- Uses a global RNG without an injected seed for replay determinism (Rule 89 / Rule 91).
- Fetches authoritative-source data without TLS pinning + certificate-transparency validation (DSO clients, Rule 125).
- Stores secrets in code or in the Pack image's filesystem (Rule 20). Secrets via ESO / Vault per topology.
- Imports `fmt.Println` or `log.Print*` instead of the Core logger surface (Rule 7 — log fields mandatory).
- Bypasses Core's idempotency, RBAC, RLS, or audit-log middlewares.
- Carries a permissive licence in `license_spdx` that conflicts with the engagement client's IP policy (some engagement contracts forbid AGPL or BUSL Packs).

---

## 7. Annual review (Rule 138)

Every Pack is reviewed annually by the Pack owner + the relevant domain advisor:

- Regulatory calendar deltas (April for Italian factor packs after the ISPRA republish; quarterly for the EFRAG taxonomy).
- Vendor-compatibility-matrix updates (new meter firmware, deprecated devices).
- Test-fixture refresh: golden fixtures regenerated if the regulator schema changed.
- CHARTER review: Tradeoff Stanza still valid; sacrifices still acceptable; residual risks closed or re-acknowledged.
- Acceptance-level review: GA remains GA, or demotion path filed.

Output is one row in `packs/<kind>/<id>/REVIEW-LOG.md` and a roll-up entry in `docs/PACK-REVIEW-LOG.md`.

---

## 8. Anti-patterns rejected

- **Quietly bumping `pack_contract_version` to dodge a deprecation warning.** The deprecation warning is the regulator-grade evidence that the Pack-contract evolution is tracked. Suppressing it breaks the audit chain.
- **Bundling per-engagement business logic into a public Pack.** Engagement-specific code goes in `engagements/<id>/packs/`, not the upstream catalogue.
- **Pack with `kind=region` but no temporal-validity factor data.** A Region Pack without factor sources is a façade; reject and split into Region + Factor Packs.
- **"Just lower the latency budget for this engagement" inside a Protocol Pack.** Rule 124 is portfolio-wide; per-engagement latency tuning is a Pack overlay, not a Core Pack edit.
- **Skipping the annual review because "the regulator didn't change anything this year."** Confirming that the regulator didn't change anything is the review.

---

## 9. Cross-references

- Charter `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2, §4, §12.2.
- Doctrine `docs/DOCTRINE.md` Rules 27, 69–88, 124, 125, 131, 138, 144.
- Pack manifest schema: `docs/contracts/pack-manifest.schema.json`.
- Pack-contract Go interfaces: `backend/internal/domain/{protocol,reporting,emissions,identity,region}/`.
- Pack-loader: `backend/internal/packs/`.
- Engagement-fork bootstrap: `docs/runbooks/engagement-fork-bootstrap.md`.
- Upstream sync: `docs/runbooks/upstream-sync.md`.
- ADR-0023: Pack-contract interfaces and versioning.
