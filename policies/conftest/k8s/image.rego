# Image references must be by digest (@sha256:...), not tag.
# Doctrine: Rule 53 (supply chain), Rule 54 (policy as code).
# Mitigates: RISK-002 (Alpine rolling apk), RISK-005 (tag mutation), RISK-017 (unsigned deploy).

package main

import future.keywords.contains
import future.keywords.if
import future.keywords.in
import future.keywords.regex

workload_kinds := {"Deployment", "StatefulSet", "DaemonSet", "Job", "CronJob", "Pod", "ReplicaSet"}

is_workload(input) if {
    input.kind in workload_kinds
}

pod_spec(obj) := obj.spec.template.spec if obj.kind in {"Deployment", "StatefulSet", "DaemonSet", "Job", "ReplicaSet"}
pod_spec(obj) := obj.spec.jobTemplate.spec.template.spec if obj.kind == "CronJob"
pod_spec(obj) := obj.spec if obj.kind == "Pod"

containers(spec) := array.concat(
    object.get(spec, "containers", []),
    object.get(spec, "initContainers", []),
)

# Hard fail on `:latest`.
deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    endswith(c.image, ":latest")
    msg := sprintf("[%s/%s][container=%s] image '%s' uses :latest tag — must be digest-pinned (@sha256:...)", [input.kind, input.metadata.name, c.name, c.image])
}

# Warn on tag without digest (audit S2; flips to deny in S3).
warn[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    not contains(c.image, "@sha256:")
    msg := sprintf("[%s/%s][container=%s] image '%s' is not digest-pinned (RISK-002, RISK-005); will be enforced in S3", [input.kind, input.metadata.name, c.name, c.image])
}
