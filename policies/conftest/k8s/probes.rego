# Long-running workloads must declare liveness + readiness + startup probes.
# Doctrine: Rule 15 (reliability), Rule 18 (observability), Rule 20 (operational reality), Rule 43 (framework awareness).

package main

import future.keywords.contains
import future.keywords.if
import future.keywords.in

long_running_kinds := {"Deployment", "StatefulSet", "DaemonSet"}

is_long_running(input) if {
    input.kind in long_running_kinds
}

pod_spec(obj) := obj.spec.template.spec

deny[msg] {
    is_long_running(input)
    spec := pod_spec(input)
    c := spec.containers[_]
    not c.livenessProbe
    msg := sprintf("[%s/%s][container=%s] livenessProbe must be set on long-running workloads", [input.kind, input.metadata.name, c.name])
}

deny[msg] {
    is_long_running(input)
    spec := pod_spec(input)
    c := spec.containers[_]
    not c.readinessProbe
    msg := sprintf("[%s/%s][container=%s] readinessProbe must be set on long-running workloads", [input.kind, input.metadata.name, c.name])
}

# Startup probe is recommended but not strictly required for fast-starting containers.
warn[msg] {
    is_long_running(input)
    spec := pod_spec(input)
    c := spec.containers[_]
    not c.startupProbe
    msg := sprintf("[%s/%s][container=%s] startupProbe recommended on long-running workloads", [input.kind, input.metadata.name, c.name])
}
