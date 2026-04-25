# mTLS Plan

**Owner:** `@greenmetrics/secops`, `@greenmetrics/platform-team`.
**Doctrine refs:** Rule 19, Rule 39, Rule 62 (identity discipline), Rule 65 (zero-trust).
**Plan ADR:** `docs/adr/0020-cert-manager-vs-spire.md`.

## 1. Decision

Use **cert-manager** + **trust-manager** to issue per-pod certificates for in-cluster service-to-service mTLS. SPIRE rejected for Phase 1 (REJ-31): control-plane operational complexity outweighs benefit at this scale, and cert-manager already covers the cluster CA + Let's Encrypt edge.

## 2. Trust topology

```
                ┌─────────────┐
                │ Public ACME │ Let's Encrypt issuer
                └──────┬──────┘
                       │ ACME HTTP-01
                       ▼
                ┌─────────────────────────┐
                │ cert-manager Issuer     │
                │  (greenmetrics-tls)     │
                └──────┬──────────────────┘
                       │
        ┌──────────────┼─────────────┐
        ▼              ▼             ▼
   Ingress TLS    Internal CA    Backend mTLS
   (edge)         (cert-manager  (per-pod cert
                  ClusterIssuer  trust-manager
                  greenmetrics-  bundle)
                  internal-ca)
```

Two issuers:

- **`greenmetrics-tls`** — Let's Encrypt for public-facing TLS (ingress, Argo CD UI, Grafana UI).
- **`greenmetrics-internal-ca`** — self-signed root + intermediate, distributed to every pod via trust-manager `Bundle`.

## 3. Per-pod certificate

`gitops/base/cert-manager/certificate-backend.yaml` (S4 to ship):

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: greenmetrics-backend-mtls
  namespace: greenmetrics
spec:
  secretName: greenmetrics-backend-mtls
  issuerRef:
    name: greenmetrics-internal-ca
    kind: ClusterIssuer
  commonName: greenmetrics-backend.greenmetrics.svc.cluster.local
  dnsNames:
    - greenmetrics-backend.greenmetrics.svc.cluster.local
    - greenmetrics-backend
  duration: 720h          # 30 days
  renewBefore: 168h       # renew 7 days before expiry
  privateKey:
    algorithm: ECDSA
    size: 256
    rotationPolicy: Always
  usages: [server auth, client auth, digital signature, key encipherment]
```

## 4. Trust bundle

`gitops/base/cert-manager/bundle-internal-ca.yaml`:

```yaml
apiVersion: trust.cert-manager.io/v1alpha1
kind: Bundle
metadata: { name: greenmetrics-internal-ca-bundle }
spec:
  sources:
    - useDefaultCAs: true
    - secret:
        name: greenmetrics-internal-ca-root
        key: ca.crt
        namespace: cert-manager
  target:
    configMap:
      key: ca-bundle.crt
    namespaceSelector:
      matchLabels: { app.kubernetes.io/part-of: greenmetrics }
```

## 5. Backend mTLS configuration

Backend Fiber app reads cert + key from projected secret volume mount; client cert verification on mTLS-required endpoints (e.g. `/api/internal/metrics`, `/api/v1/admin/*`).

Pod template (Kustomize patch in S4):

```yaml
volumes:
  - name: mtls
    secret: { secretName: greenmetrics-backend-mtls }
  - name: trust
    configMap: { name: greenmetrics-internal-ca-bundle }
volumeMounts:
  - mountPath: /etc/greenmetrics/mtls
    name: mtls
    readOnly: true
  - mountPath: /etc/greenmetrics/trust
    name: trust
    readOnly: true
env:
  - { name: TLS_CERT_FILE, value: /etc/greenmetrics/mtls/tls.crt }
  - { name: TLS_KEY_FILE,  value: /etc/greenmetrics/mtls/tls.key }
  - { name: TLS_CA_FILE,   value: /etc/greenmetrics/trust/ca-bundle.crt }
```

## 6. Edge TLS (Let's Encrypt)

- Ingress uses `cert-manager.io/cluster-issuer: greenmetrics-tls` annotation.
- ACME HTTP-01 via nginx ingress controller.
- Renewal automated by cert-manager; expiry alert fires 14d before expiry.
- Runbook: `docs/runbooks/cert-rotation.md` (S4).

## 7. Phase 2 — multi-cluster

If multi-cluster federation appears in roadmap:

- Re-evaluate SPIRE + SPIFFE for cross-cluster identity.
- ADR to flip REJ-31.
- cert-manager remains for in-cluster issuance; SPIRE bridges federation.

## 8. Anti-patterns rejected

- TLS termination at backend without mTLS to upstream — accepted today (ingress → backend), but plan above closes the gap on internal hops.
- Bring-your-own-CA per pod — adds operational burden; trust-manager bundle is the rule.
- Long-lived certs (≥ 1y) — rotation cycle is 30d for per-pod, 90d for edge.
- mTLS for everything — overhead not justified for read-only public endpoints; document per-endpoint mTLS requirement in `docs/TRUST-BOUNDARIES.md`.
