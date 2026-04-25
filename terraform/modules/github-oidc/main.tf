# GitHub Actions OIDC trust — federated identity for CI → AWS without static keys.
# Doctrine: Rule 19, Rule 39, Rule 53 (supply chain), Rule 57 (least privilege).
# Mitigates: RISK-005, RISK-006.

terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.60" }
  }
}

variable "github_org" {
  description = "GitHub org / user that owns the repo."
  type        = string
}

variable "github_repo" {
  description = "Repository name (e.g. greenmetrics)."
  type        = string
  default     = "greenmetrics"
}

variable "tags" {
  description = "Mandatory cost-allocation tags."
  type        = map(string)
  default = {
    Project        = "greenmetrics"
    Environment    = "shared"
    Owner          = "platform-team"
    CostCenter     = "platform"
    DataResidency  = "EU-South-1-IT"
  }
}

# --- OIDC provider ----------------------------------------------------------

resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  # GitHub Actions OIDC root CA thumbprints. Refresh annually per docs/SECOPS-RUNBOOK.md.
  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd",
  ]
  tags = var.tags
}

# --- Deploy role (push images to GHCR is GitHub-native; here we permit AWS deploys) ---

data "aws_iam_policy_document" "deploy_assume" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }
    # Trust only this repo, only main branch pushes and tag pushes.
    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }
    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values = [
        "repo:${var.github_org}/${var.github_repo}:ref:refs/heads/main",
        "repo:${var.github_org}/${var.github_repo}:ref:refs/tags/v*",
        "repo:${var.github_org}/${var.github_repo}:environment:staging",
        "repo:${var.github_org}/${var.github_repo}:environment:production",
      ]
    }
  }
}

resource "aws_iam_role" "github_deploy" {
  name               = "greenmetrics-github-deploy"
  description        = "Trusted by GitHub Actions OIDC for deploy workflows. Bound to repo + ref."
  assume_role_policy = data.aws_iam_policy_document.deploy_assume.json
  max_session_duration = 3600
  tags               = var.tags
}

# Minimal deploy policy — Argo CD does the cluster mutation; this role only updates
# image tags in S3-backed gitops repo + reads Secrets Manager for ESO bootstrap.
data "aws_iam_policy_document" "deploy_inline" {
  statement {
    actions   = ["s3:GetObject", "s3:PutObject"]
    resources = ["arn:aws:s3:::greenmetrics-gitops/*"]
  }
  statement {
    actions   = ["s3:ListBucket"]
    resources = ["arn:aws:s3:::greenmetrics-gitops"]
  }
}

resource "aws_iam_role_policy" "deploy" {
  name   = "greenmetrics-github-deploy-policy"
  role   = aws_iam_role.github_deploy.id
  policy = data.aws_iam_policy_document.deploy_inline.json
}

# --- Outputs ---------------------------------------------------------------

output "oidc_provider_arn" { value = aws_iam_openid_connect_provider.github.arn }
output "deploy_role_arn"   { value = aws_iam_role.github_deploy.arn }
