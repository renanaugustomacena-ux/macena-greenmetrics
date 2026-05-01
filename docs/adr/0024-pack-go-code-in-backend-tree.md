# 0024 — Pack Go code in `backend/packs/` (interim, single-module phase)

**Status:** Accepted
**Date:** 2026-05-01
**Authors:** @ciupsciups
**Doctrine refs:** Rules 69 (Core and Pack are the load-bearing distinction), 71 (Pack contracts versioned independently of Core), 72 (Pack registration via Registrar, never via global), 87 (god-service decomposition).
**Charter refs:** §3.2 (Pack flavours), §4 (Pack contract), §5.1 (upstream template repository).
**Plan ref:** `docs/PLAN.md` §5.4 (Sprint S6 — Italian Region/Factor/Report Pack extraction).
**Supersession schedule:** target Phase F Sprint S11 (when contract interfaces are extracted to `github.com/greenmetrics/contract` and Pack code moves to repo-root `packs/<kind>/<id>/`).

## Context

Sprint S6 begins populating the Italian-flagship Pack catalogue. Per Charter §3.2:

> A Pack is a self-contained directory of code, schema, fixtures, tests, ADR, and Pack-charter under `packs/<pack-id>/`.

The literal reading is that Pack Go code lives at `packs/<kind>/<id>/`. However, Pack Go code imports the contract interfaces declared in `backend/internal/domain/{region,emissions,reporting,protocol,identity}/`. With the current single-module repo (one `go.mod` at `backend/`):

- If `packs/` is a separate module, it depends on `backend/`. Backend's `cmd/server` (which boots the binary) needs to register the Packs at boot — that requires `backend/cmd/server` to import `packs/<kind>/<id>/`. **Result: a backend ↔ packs module cycle, which Go forbids.**
- If `packs/` is part of the backend module (single `go.mod` at backend root, with import paths spanning the repo), the file layout `packs/<kind>/<id>/` cannot be Go code under that module; it would need to be either inside `backend/` or the module path would need to span the entire repo.

Three structural fixes have been considered:

1. **Move `go.mod` to the repo root**, change the module path from `github.com/greenmetrics/backend` to (e.g.) `github.com/greenmetrics/template`, fold all current `internal/` paths under the new prefix. Big refactor; touches every import line in the project.
2. **Extract contract interfaces** (`internal/domain/{region,emissions,reporting,protocol,identity}/`) into a separate `github.com/greenmetrics/contract` module. Both `backend` and `packs` depend on `contract`; `cmd/server` imports `packs` for boot registration; **no cycle**. Cleanest long-term; touches every existing reference to the contract interfaces.
3. **Place Pack Go code inside the backend module** (e.g. at `backend/packs/<kind>/<id>/`); keep the manifest + CHARTER + non-Go assets at the repo-root `packs/<kind>/<id>/` for Charter §3.2 discovery. Smallest change; keeps each Pack split across two paths.

## Decision

Adopt **option 3 as an interim** for Sprint S6. Pack Go code lives at `backend/packs/<kind>/<id>/`; Pack manifest + CHARTER + fixtures + ADRs live at repo-root `packs/<kind>/<id>/`. Each Pack's CHARTER §"Layout note" calls out the split and references this ADR.

Schedule **option 2 (contract-extraction)** for Phase F Sprint S11. At supersession:

1. Create `contract/go.mod` at repo root with `module github.com/greenmetrics/contract`.
2. Move `backend/internal/domain/{region,emissions,reporting,protocol,identity}/` → `contract/<kind>/`.
3. Update all `backend/` imports to reference the new path.
4. Create `packs/go.mod` at repo root (or use `go.work`).
5. Move every `backend/packs/<kind>/<id>/` → `packs/<kind>/<id>/`.
6. Delete `backend/packs/`.
7. Update each Pack's CHARTER `Layout note` to point at the unified location.
8. Bump Pack contract versions per Rule 71 if the move surfaces any incompatibilities (none expected — pure interface relocation).

## Alternatives considered

- **Option 1 (move `go.mod` to repo root).** Rejected. Touching every import in the codebase produces a >200-file diff with no functional change; the resulting paths (`github.com/greenmetrics/template/internal/...`) are awkward; the historical backend module identity is lost.
- **Option 2 (contract module) right now.** Rejected as "right now" because Sprint S6 is already a 4-week sprint with substantial Italian-flagship work; folding a contract-extraction refactor into the same window puts both at risk. Deferring to Sprint S11 lets Sprint S6 focus on the regulatory pack content.
- **Two-binary split (cmd/server out of backend module).** Rejected. Adds a third Go module for the engagement boot wiring; doesn't actually resolve the cycle (cmd → packs → backend → cmd's shared types) without contract extraction.

## Consequences

### Positive

- Sprint S6 ships immediately. Italian Region Pack, ISPRA Factor Pack, and `monthly_consumption` Report Pack land in this PR with full tests passing under Go 1.26.
- The split path (manifest at `packs/`, Go code at `backend/packs/`) is symmetric across all five Pack kinds; CHARTER §"Layout note" makes the rule explicit per Pack.
- The repo-root `packs/` directory remains the Charter-§3.2 discovery surface — the manifest, CHARTER, and any non-Go assets (fixtures, schemas) live there.
- The supersession schedule is concrete (Phase F Sprint S11) with a defined migration procedure.

### Negative

- Each Pack is split across two directories. A reviewer reading `packs/region/it/CHARTER.md` must follow the §"Layout note" pointer to `backend/packs/region/it/profile.go`. The CHARTER explicitly cross-references both paths.
- Engagement-fork bootstrap creates two parallel paths under `engagements/<id>/` for engagement-specific Packs (one in the engagement's `packs/`, one in the engagement's `backend/packs/`) until Sprint S11 unifies them. The bootstrap runbook is updated to reflect this.
- The single-module status holds the contract package boundaries open — anyone editing `backend/internal/domain/<kind>/` without coordinating Pack-side updates can break Packs at next-build. Mitigated by the conformance suite (`tests/packs/contract_compliance_test.go`).

### Neutral

- No effect on the Pack contract itself — the interfaces in `backend/internal/domain/<kind>/` are unchanged; only the location of implementation Go files is affected.
- No effect on the manifest schema, CHARTER format, or `config/required-packs.yaml` shape.

## Residual risks

- **Pack-loader implementation drifts from this layout convention.** The Phase E Sprint S5 Pack-loader skeleton (`backend/internal/packs/`) has not yet wired the `packs/<kind>/<id>/` ↔ `backend/packs/<kind>/<id>/` discovery; that wiring lands in Sprint S6 PR #2 (the Pack-loader hardening) and the conformance test pins the convention.
- **Engagement forks predating Sprint S11 ship the split layout.** Their `engagements/<id>/packs/` and `engagements/<id>/backend/packs/` overlays follow the same pattern. The Sprint S11 migration guide explicitly handles this case.
- **Contract-interface drift** between Pack code and backend's domain types is a single-module hazard — `tests/packs/contract_compliance_test.go` (Sprint S6 deliverable) closes it.

## References

- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §3.2, §4, §5.1.
- Doctrine: Rules 69, 71, 72, 87.
- Plan: §5.4 (Sprint S6), §6 (Phase F).
- Sister ADRs: 0023 (Pack-contract interfaces and versioning).
- Related Pack CHARTERs: `packs/region/it/CHARTER.md` §9, `packs/factor/ispra/CHARTER.md` §5, `packs/report/monthly_consumption/CHARTER.md` §5.

## Tradeoff Stanza

- **Solves:** the backend ↔ packs module cycle in a single-module repo; the Sprint S6 schedule pressure that makes a Sprint-scope contract-extraction refactor risky.
- **Optimises for:** Sprint S6 velocity, structural symmetry across the five Pack kinds, a concrete supersession schedule, low-friction reviewer experience (CHARTER `Layout note` is the single anchor).
- **Sacrifices:** strict adherence to Charter §3.2 ("self-contained directory under `packs/<pack-id>/`") for the Sprint S6 → Sprint S11 window; one extra path each Pack reviewer must follow; future migration cost in Sprint S11 (mitigated by the documented procedure).
- **Residual risks:** drift between Pack code and contract interfaces (mitigated by the conformance suite); engagement-fork layout duplication during the interim window; risk of the supersession date slipping beyond Sprint S11 (mitigated by the explicit ADR-supersession entry that this ADR carries).
