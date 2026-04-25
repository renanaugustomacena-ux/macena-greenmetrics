# Workload must declare resources.requests + resources.limits.
# Doctrine: Rule 15 (reliability), Rule 16 (scalability), Rule 22 (cost awareness), Rule 42 (resource lifecycle).
# Mitigates: RISK-014, RISK-015, RISK-021 (cardinality + cost predictability).

package main

import future.keywords.contains
import future.keywords.if
import future.keywords.in

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

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    not c.resources.requests.cpu
    msg := sprintf("[%s/%s][container=%s] resources.requests.cpu must be set", [input.kind, input.metadata.name, c.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    not c.resources.requests.memory
    msg := sprintf("[%s/%s][container=%s] resources.requests.memory must be set", [input.kind, input.metadata.name, c.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    not c.resources.limits.cpu
    msg := sprintf("[%s/%s][container=%s] resources.limits.cpu must be set", [input.kind, input.metadata.name, c.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    not c.resources.limits.memory
    msg := sprintf("[%s/%s][container=%s] resources.limits.memory must be set", [input.kind, input.metadata.name, c.name])
}
