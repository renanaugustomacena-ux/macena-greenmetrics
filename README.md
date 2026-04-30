# GreenMetrics

> **Monitora i Consumi, Riduci le Emissioni, Risparmia.**
> Energy management & sustainability reporting platform for Italian energy-intensive SMEs.

GreenMetrics is an open-source platform that ingests multi-protocol meter readings, turns them into compliant sustainability reports (CSRD / ESRS E1, Piano Transizione 5.0 attestazione, Conto Termico 2.0, Certificati Bianchi TEE, D.Lgs. 102/2014 audit energetico), and helps manufacturers monetise the Piano Transizione 5.0 tax credit (up to 45%).

---

## Stack

| Layer      | Technology                                        |
|------------|---------------------------------------------------|
| Backend    | Go 1.26 + [Fiber](https://gofiber.io)             |
| Frontend   | [SvelteKit 2](https://kit.svelte.dev) + Tailwind  |
| Database   | [TimescaleDB](https://www.timescale.com) (PG 16)  |
| Dashboards | Grafana (provisioned)                             |
| IaC        | Terraform (AWS eu-south-1 Milan + Aruba Cloud alt.)|
| Observability | OpenTelemetry (OTLP/gRPC) + Prometheus         |

---

## Highlight: Piano Transizione 5.0 — il prodotto si ripaga da solo

The `piano_5_0_attestazione` report computes baseline vs post-intervention energy
savings (thresholds 3% per process / 5% per site) and generates the attestazione
package that unlocks the 5–45% tax credit on qualifying spend. For a typical
mid-sized Verona manufacturer (€500k–€2M eligible investment), the credit alone
often exceeds the lifetime cost of a GreenMetrics subscription.

---

## Quick Start (English)

```bash
cp .env.example .env
docker compose up -d
# backend   → http://localhost:8080
# frontend  → http://localhost:3000
# grafana   → http://localhost:3001  (admin / change-me)
# timescale → localhost:5432
```

Apply migrations (first run):

```bash
docker compose exec greenmetrics-timescaledb psql -U greenmetrics -d greenmetrics \
  -f /docker-entrypoint-initdb.d/0001_init.sql
# ... and likewise for 0002_hypertables.sql, 0003_continuous_aggregates.sql,
# 0004_retention.sql, 0005_emission_factors.sql.
```

Smoke test:

```bash
curl -sS http://localhost:8080/api/health | jq
```

## Avvio rapido (Italiano)

```bash
cp .env.example .env
docker compose up -d
# backend   → http://localhost:8080
# frontend  → http://localhost:3000
# grafana   → http://localhost:3001
```

---

## API overview

Fully documented in [`docs/API.md`](./docs/API.md). Key endpoints:

| Method | Path                              | Description                                 |
|--------|-----------------------------------|---------------------------------------------|
| GET    | `/api/health`                     | Health + dependency status                  |
| POST   | `/api/v1/auth/login`              | Issue access + refresh JWT                  |
| GET    | `/api/v1/meters`                  | List meters for the authenticated tenant    |
| POST   | `/api/v1/readings/ingest`         | Bulk ingest readings (hypertable copy)      |
| GET    | `/api/v1/readings/aggregated`     | Query continuous-aggregate view             |
| POST   | `/api/v1/reports`                 | Generate CSRD / Piano 5.0 / audit reports   |
| GET    | `/api/v1/emission-factors`        | List ISPRA / GSE versioned factors          |
| GET    | `/api/v1/alerts`                  | Current and recent alerts                   |

---

## Documentation

### Product & architecture

- [`docs/MODUS_OPERANDI.md`](./docs/MODUS_OPERANDI.md) — commercial + technical playbook (~14k words).
- [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) — architecture & sequence diagrams.
- [`docs/API.md`](./docs/API.md) — REST reference.
- [`docs/LAYERS.md`](./docs/LAYERS.md) — five-layer system map (Infra → Operators).

### Doctrine & governance

- [`docs/TEAM-CHARTER.md`](./docs/TEAM-CHARTER.md) — platform team mandate, scope, authority.
- [`docs/RACI.md`](./docs/RACI.md) — who is Responsible/Accountable/Consulted/Informed for each subsystem.
- [`docs/SECOPS-CHARTER.md`](./docs/SECOPS-CHARTER.md) — secops mandate.
- [`docs/PLATFORM-INITIATIVE-WORKFLOW.md`](./docs/PLATFORM-INITIATIVE-WORKFLOW.md) — Rule 11/31/51 sequence for any new platform feature.
- [`docs/adr/`](./docs/adr/) — Architecture Decision Records.
- [`docs/adr/REJECTED.md`](./docs/adr/REJECTED.md) — anti-patterns explicitly rejected.

### Security & risk

- [`docs/SECURITY.md`](./docs/SECURITY.md) — threat overview + accepted risks.
- [`docs/THREAT-MODEL.md`](./docs/THREAT-MODEL.md) — STRIDE per attack surface.
- [`docs/RISK-REGISTER.md`](./docs/RISK-REGISTER.md) — L×I scored, mitigation-mapped.
- [`docs/SUPPLY-CHAIN.md`](./docs/SUPPLY-CHAIN.md) — Cosign + SLSA + Kyverno trust chain.
- [`docs/ITALIAN-COMPLIANCE.md`](./docs/ITALIAN-COMPLIANCE.md) — regulatory citation map (CSRD, Piano 5.0, D.Lgs. 102/2014, GDPR, NIS2).

### Operations

- [`docs/SLO.md`](./docs/SLO.md), [`docs/SLI-CATALOG.md`](./docs/SLI-CATALOG.md), [`docs/OBSERVABILITY.md`](./docs/OBSERVABILITY.md).
- [`docs/RUNBOOK.md`](./docs/RUNBOOK.md) + [`docs/runbooks/`](./docs/runbooks/) (db-outage, ingestor-crash-loop, grafana-down, cert-rotation, secret-rotation, jwt-secret-rotation, capacity-spike, region-failover, tenant-data-leak, pulse-webhook-flood, cost-audit).
- [`docs/INCIDENT-RESPONSE.md`](./docs/INCIDENT-RESPONSE.md), [`docs/ON-CALL.md`](./docs/ON-CALL.md), [`docs/ROLLBACK.md`](./docs/ROLLBACK.md).
- [`docs/RELIABILITY-MODEL.md`](./docs/RELIABILITY-MODEL.md), [`docs/CHAOS-PLAN.md`](./docs/CHAOS-PLAN.md), [`docs/CHAOS-LOG.md`](./docs/CHAOS-LOG.md).
- [`docs/CAPACITY.md`](./docs/CAPACITY.md), [`docs/COST-MODEL.md`](./docs/COST-MODEL.md).
- [`docs/PIPELINE-MAP.md`](./docs/PIPELINE-MAP.md), [`docs/PENTEST-CADENCE.md`](./docs/PENTEST-CADENCE.md).
- [`docs/JWT-ROTATION.md`](./docs/JWT-ROTATION.md), [`docs/MTLS-PLAN.md`](./docs/MTLS-PLAN.md), [`docs/TRUST-BOUNDARIES.md`](./docs/TRUST-BOUNDARIES.md).
- [`docs/A11Y.md`](./docs/A11Y.md), [`docs/CONTACTS.md`](./docs/CONTACTS.md).
- [`docs/COMPLIANCE/`](./docs/COMPLIANCE/) — CSRD, Piano 5.0, D.Lgs. 102/2014, GDPR, NIS2 evidence packs.

### Contributing

- [`docs/CONTRIBUTING.md`](./docs/CONTRIBUTING.md) — bootstrap, dev loop, doctrine your PR will be reviewed against.
- [`docs/DX.md`](./docs/DX.md), [`docs/DEBUG.md`](./docs/DEBUG.md), [`docs/TROUBLESHOOTING.md`](./docs/TROUBLESHOOTING.md).
- [`docs/PLATFORM-PLAYBOOK.md`](./docs/PLATFORM-PLAYBOOK.md) — daily / quarterly / annual cadence + termination criteria.
- [`docs/PLATFORM-DEFAULTS.md`](./docs/PLATFORM-DEFAULTS.md) — opinionated stack choices.
- [`docs/QUALITY-BAR.md`](./docs/QUALITY-BAR.md) — non-negotiable invariants.
- [`docs/SERVICE-CATALOG.md`](./docs/SERVICE-CATALOG.md), [`docs/EXTENSION-POINTS.md`](./docs/EXTENSION-POINTS.md), [`docs/ABSTRACTION-LEDGER.md`](./docs/ABSTRACTION-LEDGER.md).
- [`docs/SCHEMA-EVOLUTION.md`](./docs/SCHEMA-EVOLUTION.md), [`docs/API-VERSIONING.md`](./docs/API-VERSIONING.md), [`docs/FEATURE-FLAGS.md`](./docs/FEATURE-FLAGS.md).
- [`docs/CLI-CONTRACT.md`](./docs/CLI-CONTRACT.md), [`Taskfile.yaml`](./Taskfile.yaml), [`backend/Makefile`](./backend/Makefile).
- [`docs/contracts/`](./docs/contracts/) — config schema + CloudEvents schemas.
- [`api/openapi/v1.yaml`](./api/openapi/v1.yaml) — canonical API contract.
- [`docs/adr/`](./docs/adr/) — 20 ADRs covering every non-trivial decision.
- [`docs/adr/REJECTED.md`](./docs/adr/REJECTED.md) — 35 anti-patterns explicitly rejected.
- [`docs/backend/`](./docs/backend/) — backend-specific docs (system map, migrations, doctrine checklist).
- [`policies/`](./policies/) — conftest + Kyverno + Falco bundles.
- [`.github/CODEOWNERS`](./.github/CODEOWNERS), [`.github/PULL_REQUEST_TEMPLATE.md`](./.github/PULL_REQUEST_TEMPLATE.md), [`.github/dependabot.yml`](./.github/dependabot.yml).
- [`.pre-commit-config.yaml`](./.pre-commit-config.yaml), [`.devcontainer/`](./.devcontainer/).

---

## Compliance surface

- **CSRD** (Dir. UE 2022/2464) + **ESRS E1**.
- **D.Lgs. 102/2014** — audit energetico quadriennale.
- **Piano Transizione 5.0** — attestazione risparmi, credito 5–45%.
- **Conto Termico 2.0** — domanda GSE.
- **Certificati Bianchi (TEE)** — submission GSE.
- **ETS** (cap-and-trade, opzionale).
- **GHG Protocol** Scope 1 / 2 / 3.
- **GDPR** — data-subject-request workflow.

---

## Made in Verona · Licence + IP

GreenMetrics is **proprietary software, all rights reserved**.

**Copyright (c) 2026 Renan Augusto Macena. All rights reserved.**

The full proprietary licence governing this Software is set out in
[`LICENSE`](./LICENSE). No use, reproduction, modification, or distribution
is permitted without the express prior written permission of the author.

The author reserves the right to release this Software under an open-source
licence at a future date; until that election is made and published in the
LICENSE file, the proprietary "All Rights Reserved" notice is the sole and
exclusive licence governing this Software.

| Surface | File | Purpose |
|---|---|---|
| Licence | [`LICENSE`](./LICENSE) | Proprietary "All Rights Reserved" + future-open-source clause + Italian/EU enforceability. |
| Copyright | [`COPYRIGHT`](./COPYRIGHT) | Short copyright pointer. |
| Notice | [`NOTICE`](./NOTICE) | Trademark + dependency attribution + SBOM pointer. |
| Authors | [`AUTHORS`](./AUTHORS) | Sole author + citation format. |
| Trademark | [`TRADEMARK.md`](./TRADEMARK.md) | Trademark policy for "GreenMetrics", "Macena", and related Marks. |
| Contributing | [`CONTRIBUTING.md`](./CONTRIBUTING.md) | Contributor terms (CLA-style assignment / licence grant). Read before opening any PR. |
| Conduct | [`CODE_OF_CONDUCT.md`](./CODE_OF_CONDUCT.md) | Contributor Covenant v2.1. |
| Security | [`SECURITY.md`](./SECURITY.md) | Vulnerability disclosure policy. |

For licensing enquiries: **<renanaugustomacena@gmail.com>**.
