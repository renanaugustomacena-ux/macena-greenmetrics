# Terraform encryption + backup policy.
# Doctrine: Rule 19, Rule 39, Rule 54, Rule 60 (forensic readiness), Rule 63.
# Mitigates: RISK-009, RISK-010, RISK-018, RISK-025.

package main

import future.keywords.contains
import future.keywords.if
import future.keywords.in

# Conftest input is `terraform show -json plan.json` output.
resources := input.planned_values.root_module.resources
modules  := object.get(input.planned_values.root_module, "child_modules", [])

# Recursively collect all resources across modules.
all_resources[r] {
    r := resources[_]
}
all_resources[r] {
    m := modules[_]
    r := m.resources[_]
}
all_resources[r] {
    m := modules[_]
    sub := m.child_modules[_]
    r := sub.resources[_]
}

# --------------- RDS ---------------

deny[msg] {
    r := all_resources[_]
    r.type == "aws_db_instance"
    not r.values.storage_encrypted
    msg := sprintf("[aws_db_instance/%s] storage_encrypted must be true (RISK-018)", [r.name])
}

deny[msg] {
    r := all_resources[_]
    r.type == "aws_db_instance"
    r.values.publicly_accessible == true
    msg := sprintf("[aws_db_instance/%s] publicly_accessible: true is forbidden", [r.name])
}

# --------------- S3 ---------------

deny[msg] {
    r := all_resources[_]
    r.type == "aws_s3_bucket"
    # SSE configuration is tracked separately via aws_s3_bucket_server_side_encryption_configuration;
    # warn if a bucket has no matching SSE resource in the same plan.
    not has_sse(r.values.bucket)
    msg := sprintf("[aws_s3_bucket/%s] missing aws_s3_bucket_server_side_encryption_configuration (RISK-009)", [r.name])
}

has_sse(bucket_name) {
    sse := all_resources[_]
    sse.type == "aws_s3_bucket_server_side_encryption_configuration"
    sse.values.bucket == bucket_name
}

deny[msg] {
    r := all_resources[_]
    r.type == "aws_s3_bucket"
    not has_public_access_block(r.values.bucket)
    msg := sprintf("[aws_s3_bucket/%s] missing aws_s3_bucket_public_access_block", [r.name])
}

has_public_access_block(bucket_name) {
    pab := all_resources[_]
    pab.type == "aws_s3_bucket_public_access_block"
    pab.values.bucket == bucket_name
}

# --------------- IAM ---------------

deny[msg] {
    r := all_resources[_]
    r.type == "aws_iam_role_policy"
    contains(r.values.policy, "\"Action\": \"*\"")
    not has_secops_approved_marker(r)
    msg := sprintf("[aws_iam_role_policy/%s] '*' Action requires '# secops-approved:' justification comment (RISK-006)", [r.name])
}

has_secops_approved_marker(_) := false

# --------------- Security groups ---------------

deny[msg] {
    r := all_resources[_]
    r.type == "aws_security_group_rule"
    r.values.type == "ingress"
    r.values.cidr_blocks[_] == "0.0.0.0/0"
    not r.values.from_port in {80, 443}
    msg := sprintf("[aws_security_group_rule/%s] 0.0.0.0/0 ingress on port %v requires explicit IP allowlist", [r.name, r.values.from_port])
}
