# Contacts

**Doctrine refs:** Rule 9, Rule 49, Rule 60.

## Internal

| Role | Contact | Channel |
|---|---|---|
| Repo owner | Renan Augusto Macena | ciupsciups@libero.it |
| Platform team | `@greenmetrics/platform-team` | Slack `#greenmetrics-ops` |
| App team | `@greenmetrics/app-team` | Slack `#greenmetrics-ops` |
| SRE | `@greenmetrics/sre` | Slack `#greenmetrics-ops`, PagerDuty `greenmetrics-primary` |
| SecOps | `@greenmetrics/secops` | Slack `#secops`, `#sev1-active` |
| DPO | dpo@greenmetrics.it | (TBD external counsel) |
| Legal | legal@greenmetrics.it | (TBD external counsel) |

## Regulatory

| Authority | Purpose | Contact |
|---|---|---|
| ACN (Agenzia per la Cybersicurezza Nazionale) | NIS2 24h/72h notification | https://www.csirt.gov.it/ — credentials in Vault `greenmetrics/compliance/acn-portal` |
| Garante per la protezione dei dati personali | GDPR breach 72h notification | https://www.garanteprivacy.it/ — DPO leads |
| ISPRA | Emission factor data source | https://www.isprambiente.gov.it/ |
| Terna | Grid mix data | https://www.terna.it/ |
| GSE (Gestore dei Servizi Energetici) | Conto Termico 2.0 + TEE | https://www.gse.it/ |
| AdE (Agenzia delle Entrate) | Piano 5.0 attestazione path | https://www.agenziaentrate.gov.it/ |

## External

| Vendor | Contact | Purpose |
|---|---|---|
| AWS | TAM (Technical Account Manager) — TBD | infra escalation |
| GitHub | Premium support — TBD | CI/CD escalation |
| Sigstore | Public mailing list / GitHub issues | Cosign trust chain |
| Cert provider | Let's Encrypt (no human contact) | edge TLS |

## Customers

Customer-facing contacts maintained in CRM (out of scope of this repo).

## Update procedure

Edit this file via PR. Routes to `@greenmetrics/platform-team` + `@greenmetrics/secops` review. Quarterly review by `@greenmetrics/sre`.
