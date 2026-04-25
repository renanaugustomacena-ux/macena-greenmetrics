# Dockerfile policy gate — consumes hadolint JSON output.
# Doctrine: Rule 19, Rule 39, Rule 53, Rule 54.
# Mitigates: RISK-002, RISK-006, RISK-017.

package main

# Hadolint JSON shape: array of {file, line, code, message, level}.
# Conftest scans `hadolint --format json Dockerfile`.

deny[msg] {
    finding := input[_]
    # DL3002: do not switch to root USER (must end as nonroot)
    finding.code == "DL3002"
    msg := sprintf("[hadolint DL3002] %s:%d %s", [finding.file, finding.line, finding.message])
}

deny[msg] {
    finding := input[_]
    # DL3007: do not pin to :latest
    finding.code == "DL3007"
    msg := sprintf("[hadolint DL3007] %s:%d %s — pin base image by digest (RISK-005)", [finding.file, finding.line, finding.message])
}

deny[msg] {
    finding := input[_]
    # DL3025: ensure JSON-array form for ENTRYPOINT/CMD (avoid /bin/sh -c)
    finding.code == "DL3025"
    msg := sprintf("[hadolint DL3025] %s:%d %s", [finding.file, finding.line, finding.message])
}

deny[msg] {
    finding := input[_]
    # DL3008: pin apt versions
    finding.code == "DL3008"
    msg := sprintf("[hadolint DL3008] %s:%d %s", [finding.file, finding.line, finding.message])
}

# Anything CRITICAL is a deny (default hadolint level mapping).
deny[msg] {
    finding := input[_]
    finding.level == "error"
    msg := sprintf("[hadolint %s/%s] %s:%d %s", [finding.code, finding.level, finding.file, finding.line, finding.message])
}

# Style warnings flagged but not blocking.
warn[msg] {
    finding := input[_]
    finding.level == "warning"
    msg := sprintf("[hadolint %s/%s] %s:%d %s", [finding.code, finding.level, finding.file, finding.line, finding.message])
}
