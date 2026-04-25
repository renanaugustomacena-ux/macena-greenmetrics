# Recommended labels per K8s convention.
# Doctrine: Rule 14 (contracts at crossings), Rule 22 (cost awareness via tagging).

package main

import future.keywords.contains
import future.keywords.if
import future.keywords.in

required_labels := {
    "app.kubernetes.io/name",
    "app.kubernetes.io/version",
    "app.kubernetes.io/component",
    "app.kubernetes.io/part-of",
}

labelled_kinds := {"Deployment", "StatefulSet", "DaemonSet", "Service", "Ingress"}

warn[msg] {
    input.kind in labelled_kinds
    label := required_labels[_]
    not input.metadata.labels[label]
    msg := sprintf("[%s/%s] missing recommended label '%s'", [input.kind, input.metadata.name, label])
}
