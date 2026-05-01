# `engagements/` — per-engagement overlays

This directory holds the per-engagement code, fixtures, runbooks, ADRs, and Pack overlays that are unique to one client and must NOT flow upstream into the GreenMetrics template. The model is described in `docs/MODULAR-TEMPLATE-CHARTER.md` §5.2 (engagement forks) and the bootstrap procedure in `docs/runbooks/engagement-fork-bootstrap.md`.

## Layout

```
engagements/
├── README.md                        (this file)
├── .gitkeep                         (empty marker; keeps the dir on main)
└── <engagement-id>/                 one per engagement (created at Phase 1)
    ├── CHARTER.md                   one-page engagement contract
    ├── CLAUDE.engagement.md         engagement-specific invariants
    ├── CORE-CUSTOMISATIONS.md       Core overrides + sunset (Rule 80)
    ├── INCIDENT-RESPONSE.md         engagement-specific incident overlay
    ├── PENTEST-LOG.md               engagement-specific pen-test (Rule 60)
    ├── SYNC-LOG.md                  upstream-sync record (Rule 79)
    ├── THREAT-MODEL-OVERLAY.md      engagement-specific threat surface
    ├── adr/                         engagement-specific ADRs (numbered <id>-NNNN)
    ├── fixtures/                    synthetic-only by default (Rule 165)
    ├── packs/                       engagement-specific Pack overlays
    └── runbooks/                    engagement-specific runbooks
```

## What goes here vs upstream

| Type of artefact | Lives in upstream template | Lives in `engagements/<id>/` |
|---|---|---|
| Pack-contract interfaces | `backend/internal/domain/<kind>/` | — |
| Public Italian-flagship Packs | `packs/<kind>/<id>/` + `backend/packs/<kind>/<id>/` | — |
| Doctrine, Charter, Plan | `docs/` | — |
| Conformance suite | `tests/conformance/` | — |
| Engagement-specific Pack overlays | — | `packs/<kind>/<id>/` |
| Engagement-specific fixtures | — | `fixtures/` |
| Client-IdP-down runbook | — | `runbooks/client-idp-down.md` |
| Client SCADA / ERP integration | — | per-engagement code |
| `template-version.txt` | — | repo root of the engagement fork |

## Lifecycle

Per Rule 150, an `engagements/<id>/` directory is **only created at Phase 1 (Fork & Bootstrap)** of an engagement, after Discovery (Phase 0) signs off. Pre-signature work happens in scratch directories, not under engagement-namespaced paths. The bootstrap runbook at `docs/runbooks/engagement-fork-bootstrap.md` is the procedure.

This template repo's `engagements/` directory is empty (only this README + `.gitkeep`); engagement directories appear in **engagement forks** of this template, not in `main` of the upstream template.

## Cross-references

- Charter: `docs/MODULAR-TEMPLATE-CHARTER.md` §5.2 (engagement forks), §7 (upstream-sync discipline).
- Bootstrap: `docs/runbooks/engagement-fork-bootstrap.md`.
- Sync: `docs/runbooks/upstream-sync.md`.
- Doctrine: Rules 77, 79, 80, 149, 150, 151, 154, 155, 156, 157, 158, 165.
