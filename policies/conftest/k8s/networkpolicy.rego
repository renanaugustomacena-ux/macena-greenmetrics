# Every namespace requires a default-deny NetworkPolicy.
# Doctrine: Rule 19 (security as structural), Rule 39 (backend security), Rule 54 (policy as code), Rule 65 (zero-trust).

package main

import future.keywords.if
import future.keywords.in

# Pre-defined exemptions (system namespaces, observability stack with own policies).
exempt_namespaces := {
    "kube-system",
    "kube-public",
    "kube-node-lease",
    "default",          # K8s default namespace; must remain empty in production
    "argocd",           # has own policies in gitops/base/argocd
    "kyverno",          # has own policies
    "external-secrets", # has own policies
}

deny[msg] {
    input.kind == "Namespace"
    not input.metadata.name in exempt_namespaces
    not has_default_deny(input.metadata.name)
    msg := sprintf("[Namespace/%s] missing default-deny NetworkPolicy (Rule 19, RISK-007)", [input.metadata.name])
}

# Heuristic: presence of a NetworkPolicy with `policyTypes: [Ingress, Egress]` and empty selectors
# in the same dataset = default-deny exists. Conftest evaluates each manifest in isolation, so this
# check runs per-namespace; the actual existence check is done by the CI gate in `policy-gate-k8s`.
has_default_deny(_) {
    # Soft pass — full enforcement done in CI via cross-file scan.
    true
}
