# Architecture Decision Records

**Doctrine refs:** Rule 11 (methodological sequence), Rule 27 (decision rationale), Rule 47 (backend rationale), Rule 67 (devsecops rationale).

ADRs capture non-trivial decisions: the context, the choice, the alternatives, the consequences, the residual risks. They are the historical record so a future engineer can understand why something is the way it is.

## Naming

- `NNNN-kebab-case-title.md` (zero-padded 4-digit ID).
- IDs are monotonically increasing; never reuse.
- Filename matches the H1 inside the file.

## Status lifecycle

- `Proposed` — draft, under review.
- `Accepted` — merged; the decision is in force.
- `Superseded by NNNN` — replaced by another ADR; keep the file for history.
- `Rejected` — proposed and rejected; cross-link from `docs/adr/REJECTED.md`.

## Required sections

Every ADR must contain:

1. **Title** (H1)
2. **Status**
3. **Date**
4. **Context** — what is the problem, what constraints exist, why now.
5. **Decision** — what is being decided.
6. **Alternatives considered** — at least one realistic alternative + why not chosen.
7. **Consequences** — positive, negative, neutral.
8. **Residual risks** — what we are still exposed to after this decision.
9. **References** — links to related ADRs, runbooks, plan sections, external sources.
10. **Review date** — when to revisit.
11. **Tradeoff Stanza** (Rule 27 / 47 / 67) — four bullets at the end:
   - **Solves:** ...
   - **Optimises for:** ...
   - **Sacrifices:** ...
   - **Residual risks:** ...

CI `markdownlint` rule + `adr-link-check` job enforce presence of all sections.

## Index

| ID | Title | Status | Date | Doctrine rules | Sprint |
|---|---|---|---|---|---|
| 0000 | Template | — | 2026-04-25 | 11, 27 | — |
| 0001 | Platform doctrine adoption | Accepted | 2026-04-25 | 9–68 (all) | S1 |
| 0002 | Multi-tenant RLS strategy | Proposed | _S2_ | 19, 39 | S2 |
| 0003 | Secret management — ESO over Vault | Proposed | _S2_ | 19, 62 | S2 |
| 0004 | GitOps with Argo CD | Proposed | _S3_ | 50, 56, 63 | S3 |
| 0005 | Migration tool — pressly/goose | Proposed | _S2_ | 21, 33 | S2 |
| 0006 | Observability — OTel + Prometheus | Proposed | _S4_ | 18, 40, 58 | S4 |
| 0007 | Italian residency — AWS eu-south-1 | Proposed | _S2_ | 25, 65 | S2 |
| 0008 | API versioning policy | Proposed | _S3_ | 21, 34 | S3 |
| 0009 | Circuit breakers — sony/gobreaker | Proposed | _S4_ | 15, 36 | S4 |
| 0010 | Hypertable space partitioning | Proposed | _deferred_ | 16, 38 | deferred |
| 0011 | Postgres RLS — defence in depth | Proposed | _S3_ | 19, 39 | S3 |
| 0012 | Validator — go-playground/validator | Proposed | _S2_ | 14, 34 | S2 |
| 0013 | OpenAPI codegen — design-first via oapi-codegen | Proposed | _S2_ | 14, 34 | S2 |
| 0014 | Async report generation — Asynq + Redis | Proposed | _S4_ | 36, 37 | S4 |
| 0015 | Bounded ingest queue with drop policy | Proposed | _S4_ | 15, 36 | S4 |
| 0016 | JWT KID rotation | Proposed | _S3_ | 19, 62 | S3 |
| 0017 | Cosign keyless OIDC | Proposed | _S3_ | 53, 54 | S3 |
| 0018 | SLSA L2 now, L3 plan | Proposed | _S3_ | 53 | S3 |
| 0019 | Falco vs Tetragon | Proposed | _S4_ | 58 | S4 |
| 0020 | cert-manager vs SPIRE | Proposed | _S4_ | 19, 62 | S4 |

## Workflow

1. Open a draft PR with the new ADR (`docs/adr/NNNN-<slug>.md`) and `Status: Proposed`.
2. Discuss in the PR; iterate on alternatives + consequences.
3. On merge: flip to `Status: Accepted`, update this index.
4. If superseded later: flip to `Status: Superseded by NNNN`; keep the file.
5. If rejected: flip to `Status: Rejected`; cross-link from `docs/adr/REJECTED.md`.

## Quarterly audit

Every ADR has a review date. The quarterly platform office hours samples ADRs whose review date has passed and asks: still valid? still in force? supersede?
