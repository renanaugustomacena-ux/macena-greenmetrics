# NIS2 — Compliance Evidence Pack

**Owner:** `@greenmetrics/secops`.
**Doctrine refs:** Rule 19, Rule 39, Rule 60, Rule 61, Rule 65.
**Citation source:** Direttiva (UE) 2022/2555 (NIS2) + D.Lgs. 4 settembre 2024, n. 138 (in vigore 17 ottobre 2024).

## 1. Regulatory scope

- D.Lgs. 138/2024 transposes NIS2 in Italy.
- Sector: digital service provider (DSP). GreenMetrics qualifies as a "soggetto importante" via the SaaS-for-energy use case.
- Authority: ACN (Agenzia per la Cybersicurezza Nazionale) + CSIRT Italia.

## 2. Mapped controls + evidence

| Requirement | Implementation | Evidence |
|---|---|---|
| Risk management measures (art. 21) | `docs/RISK-REGISTER.md` + quarterly review + monthly HIGH/CRITICAL re-scoring | `docs/RISK-REGISTER.md` |
| Incident handling | `docs/INCIDENT-RESPONSE.md` + runbooks + on-call + postmortem template | `docs/INCIDENT-RESPONSE.md`, `docs/runbooks/`, `docs/ON-CALL.md` |
| Business continuity, crisis mgmt, backup mgmt | `docs/runbooks/region-failover.md` + AWS Backup snapshot policy + DR drill annual + RPO 1h / RTO 4h | `docs/SLO.md`, `docs/CHAOS-LOG.md` |
| Supply chain security | `docs/SUPPLY-CHAIN.md` + Cosign + SLSA L2 + Dependabot + osv-scanner | `.github/workflows/supply-chain.yml` |
| Network and information system security | NetworkPolicy default-deny + Kyverno admission + Falco runtime detection | `k8s/namespace.yaml`, `policies/kyverno/`, `policies/falco/` |
| Vulnerability handling and disclosure | `docs/PENTEST-CADENCE.md` + Dependabot + vuln scanners (govulncheck + Trivy + osv) + CodeQL | `.github/workflows/sast.yml`, `.github/workflows/dast.yml` |
| Cryptography + encryption | TLS 1.3, sslmode=require, KMS at-rest, JWT HS256 with KID rotation | `terraform/`, `internal/handlers/auth.go` |
| Human resources security | Single-operator today; rotation plan per `docs/SECOPS-RUNBOOK.md` | `docs/RACI.md`, `docs/SECOPS-CHARTER.md` |
| Access control and asset mgmt | IRSA per-pod + break-glass + MFA + CODEOWNERS + branch protection | `terraform/modules/iam-irsa/`, `terraform/modules/iam-breakglass/`, `.github/CODEOWNERS` |
| MFA + secure communication channels | MFA on operator IAM (terraform/modules/iam-breakglass condition); Slack `#sev1-active` for IR comms | `terraform/modules/iam-breakglass/main.tf` |

## 3. 24h preliminary notification

Within 24 h of becoming aware of a significant incident, ACN must be notified via the CSIRT portal.

Template (in `docs/INCIDENT-RESPONSE.md` §4):

```
Subject: [GreenMetrics] Notifica preliminare incidente — <date> <time> UTC

Soggetto: GreenMetrics (categoria: soggetto importante)
Identificativo CSIRT: <ID>

Tipologia incidente: <indisponibilità servizio | violazione dati | accesso non autorizzato | altro>
Data e ora di rilevazione: <UTC ISO 8601>
Durata stimata: <attiva | risolta in X minuti>
Servizi impattati: <API ingest, API reporting, ...>
Numero utenti / clienti impattati: <stimato>
Estensione geografica: <eu-south-1 (Italia)>
Misure di contenimento adottate: <es. failover, RLS verifica, JWT rotation>
Prossime azioni: <full report 72h, postmortem ≤ 5 giorni>

Punto di contatto: secops@greenmetrics.it / +39-...
```

## 4. 72h full report

Full structured report covers:

- Detailed timeline (UTC ISO 8601).
- Root cause analysis.
- Mitigation steps + effectiveness.
- Forensic findings (preserved evidence references).
- Action items (with owners + deadlines).
- Notification to other affected parties (clients, GDPR Garante if applicable).

## 5. Annual review

`@greenmetrics/secops` reviews this document + the risk register + the SECOPS-CHARTER + the threat model annually with `@greenmetrics/legal`. Outcome filed in `docs/office-hours/`.

## 6. Anti-patterns rejected

- "We don't qualify under NIS2" — verify with legal annually; do not assume.
- Notification > 24h — strict regulatory window.
- Reporting only after full root-cause analysis — preliminary first, full second.
- Notification without DPO involvement — REJ; DPO is in the loop.
