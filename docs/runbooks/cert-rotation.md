---
title: TLS certificate rotation
severity: P2
mttd_target: 14d (early warning)
mttr_target: 24h
owner: "@greenmetrics/sre"
related_alerts: [CertificateExpirySoon]
last_tested: 2026-04-25
review_date: 2026-07-25
---

# Runbook — TLS certificate rotation

**Doctrine refs:** Rule 15, Rule 19, Rule 20, Rule 62.

## Symptoms

- Alertmanager firing `CertificateExpirySoon` (< 14 days to expiry).
- cert-manager Certificate `Status: NotReady`.

## Diagnosis

```bash
# Check Certificate status.
kubectl get certificate -A
kubectl describe certificate -n greenmetrics greenmetrics-tls

# Check cert-manager events.
kubectl get events -n cert-manager --sort-by=.metadata.creationTimestamp | tail -50

# Check ACME challenge.
kubectl get challenges -A
kubectl describe challenge -n greenmetrics <name>
```

Common causes:

- Let's Encrypt rate limit hit.
- HTTP-01 challenge unreachable (Ingress / DNS misconfiguration).
- ACME account lockout.

## Mitigation

### M1 — Force renewal

```bash
# cert-manager re-issues on annotation bump.
kubectl annotate certificate -n greenmetrics greenmetrics-tls \
  cert-manager.io/issue-temporary-certificate="$(date +%s)" --overwrite
```

### M2 — Manual renewal via cmctl

```bash
cmctl renew greenmetrics-tls -n greenmetrics
cmctl status certificate greenmetrics-tls -n greenmetrics
```

### M3 — Bypass ACME for emergency (use staging issuer)

```bash
kubectl edit certificate -n greenmetrics greenmetrics-tls
# Temporarily switch issuerRef.name from greenmetrics-tls to letsencrypt-staging.
# Browser warnings expected; revert ASAP.
```

### M4 — Switch to internal CA fallback

If Let's Encrypt outage:

```bash
# Use the greenmetrics-internal-ca issuer for an emergency self-signed cert.
# Customers see browser warnings; status page must communicate.
```

## Recovery

1. Certificate `Status: Ready: True`.
2. Browser shows valid certificate chain.
3. `cert_exporter_not_after - time()` > 30d.
4. Alertmanager `CertificateExpirySoon` resolves.

## Annual review

Cosign root key trust review (Sigstore Fulcio CA chain) — separate task in `docs/SECOPS-RUNBOOK.md`.
