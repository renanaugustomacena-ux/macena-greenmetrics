# Terraform mandatory cost-allocation tags.
# Doctrine: Rule 22 (cost awareness), Rule 54 (policy as code).

package main

import future.keywords.if
import future.keywords.in

required_tags := {"Project", "Environment", "Owner", "CostCenter", "DataResidency"}

# Tagged resource types (extend as the plan grows).
taggable_types := {
    "aws_db_instance",
    "aws_s3_bucket",
    "aws_eks_cluster",
    "aws_eks_node_group",
    "aws_vpc",
    "aws_subnet",
    "aws_security_group",
    "aws_iam_role",
    "aws_secretsmanager_secret",
    "aws_kms_key",
    "aws_cloudfront_distribution",
    "aws_wafv2_web_acl",
}

resources := input.planned_values.root_module.resources
modules  := object.get(input.planned_values.root_module, "child_modules", [])

all_resources[r] {
    r := resources[_]
}
all_resources[r] {
    m := modules[_]
    r := m.resources[_]
}

deny[msg] {
    r := all_resources[_]
    r.type in taggable_types
    tag := required_tags[_]
    not r.values.tags[tag]
    msg := sprintf("[%s/%s] missing required tag '%s' (Rule 22)", [r.type, r.name, tag])
}
