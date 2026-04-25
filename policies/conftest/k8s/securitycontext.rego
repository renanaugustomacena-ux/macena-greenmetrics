# Pod Security Standards — restricted enforcement.
# Doctrine: Rule 19 (security as structural), Rule 39 (backend security as core), Rule 54 (policy as code).
# Mitigates: RISK-006, RISK-017 (defence in depth at PR time).

package main

import future.keywords.contains
import future.keywords.if
import future.keywords.in

workload_kinds := {"Deployment", "StatefulSet", "DaemonSet", "Job", "CronJob", "Pod", "ReplicaSet"}

is_workload(input) if {
    input.kind in workload_kinds
}

# ---------------------------------------------------------------------------
# Pod-spec extraction (handles all wrapper kinds).
# ---------------------------------------------------------------------------

pod_spec(obj) := obj.spec.template.spec if obj.kind in {"Deployment", "StatefulSet", "DaemonSet", "Job", "ReplicaSet"}
pod_spec(obj) := obj.spec.jobTemplate.spec.template.spec if obj.kind == "CronJob"
pod_spec(obj) := obj.spec if obj.kind == "Pod"

containers(spec) := array.concat(
    array.concat(
        object.get(spec, "containers", []),
        object.get(spec, "initContainers", []),
    ),
    object.get(spec, "ephemeralContainers", []),
)

# ---------------------------------------------------------------------------
# Pod-level security context.
# ---------------------------------------------------------------------------

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    not spec.securityContext.runAsNonRoot == true
    msg := sprintf("[%s/%s] pod.spec.securityContext.runAsNonRoot must be true (RISK-006)", [input.kind, input.metadata.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    not spec.securityContext.seccompProfile.type
    msg := sprintf("[%s/%s] pod.spec.securityContext.seccompProfile.type must be set (RuntimeDefault recommended)", [input.kind, input.metadata.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    spec.hostNetwork == true
    msg := sprintf("[%s/%s] hostNetwork: true is forbidden", [input.kind, input.metadata.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    spec.hostPID == true
    msg := sprintf("[%s/%s] hostPID: true is forbidden", [input.kind, input.metadata.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    spec.hostIPC == true
    msg := sprintf("[%s/%s] hostIPC: true is forbidden", [input.kind, input.metadata.name])
}

# ---------------------------------------------------------------------------
# Container-level security context.
# ---------------------------------------------------------------------------

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    c.securityContext.privileged == true
    msg := sprintf("[%s/%s][container=%s] privileged: true is forbidden (RISK-006)", [input.kind, input.metadata.name, c.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    not c.securityContext.allowPrivilegeEscalation == false
    msg := sprintf("[%s/%s][container=%s] securityContext.allowPrivilegeEscalation must be false", [input.kind, input.metadata.name, c.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    not c.securityContext.readOnlyRootFilesystem == true
    msg := sprintf("[%s/%s][container=%s] securityContext.readOnlyRootFilesystem must be true", [input.kind, input.metadata.name, c.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    not "ALL" in c.securityContext.capabilities.drop
    msg := sprintf("[%s/%s][container=%s] securityContext.capabilities.drop must include ALL", [input.kind, input.metadata.name, c.name])
}

deny[msg] {
    is_workload(input)
    spec := pod_spec(input)
    c := containers(spec)[_]
    object.get(c.securityContext.capabilities, "add", []) != []
    msg := sprintf("[%s/%s][container=%s] securityContext.capabilities.add is forbidden (drop=ALL is the rule)", [input.kind, input.metadata.name, c.name])
}
