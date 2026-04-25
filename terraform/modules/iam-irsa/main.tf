# IAM Roles for Service Accounts (IRSA) — per-pod least privilege.
# Doctrine: Rule 19 (security as structural), Rule 39 (backend security as core), Rule 57 (least privilege).
# Mitigates: RISK-006 (insider IAM access), RISK-024 (ESO sync failure narrowing).

terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.60" }
  }
}

variable "cluster_name" {
  description = "EKS cluster name."
  type        = string
}

variable "oidc_provider_arn" {
  description = "EKS cluster OIDC provider ARN."
  type        = string
}

variable "oidc_provider_url" {
  description = "EKS cluster OIDC provider URL (host only, no scheme)."
  type        = string
}

variable "aws_region" {
  description = "AWS region."
  type        = string
}

variable "namespace" {
  description = "Kubernetes namespace for the GreenMetrics application."
  type        = string
  default     = "greenmetrics"
}

variable "secrets_manager_path_prefix" {
  description = "Secrets Manager path prefix scoping ESO + backend reads."
  type        = string
  default     = "greenmetrics/prod"
}

variable "tags" {
  description = "Mandatory cost-allocation tags (Rule 22)."
  type        = map(string)
  default = {
    Project        = "greenmetrics"
    Environment    = "production"
    Owner          = "platform-team"
    CostCenter     = "greenmetrics-runtime"
    DataResidency  = "EU-South-1-IT"
  }
}

locals {
  oidc_host = replace(var.oidc_provider_url, "https://", "")
}

# --- Backend application IRSA ----------------------------------------------

data "aws_iam_policy_document" "backend_assume" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = [var.oidc_provider_arn]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_host}:sub"
      values   = ["system:serviceaccount:${var.namespace}:greenmetrics-backend"]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_host}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "backend_inline" {
  # Read only the GreenMetrics-prefixed secrets — no `*` actions.
  statement {
    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
    ]
    resources = [
      "arn:aws:secretsmanager:${var.aws_region}:*:secret:${var.secrets_manager_path_prefix}/*",
    ]
  }
  statement {
    actions = [
      "kms:Decrypt",
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["secretsmanager.${var.aws_region}.amazonaws.com"]
    }
  }
  # CloudWatch log writes for application logs (read-only on log groups beyond the app's own).
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["arn:aws:logs:${var.aws_region}:*:log-group:/aws/eks/${var.cluster_name}/greenmetrics/*:*"]
  }
}

resource "aws_iam_role" "backend" {
  name               = "greenmetrics-backend-irsa"
  assume_role_policy = data.aws_iam_policy_document.backend_assume.json
  tags               = var.tags
}

resource "aws_iam_role_policy" "backend" {
  name   = "greenmetrics-backend-policy"
  role   = aws_iam_role.backend.id
  policy = data.aws_iam_policy_document.backend_inline.json
}

# --- External Secrets Operator IRSA ----------------------------------------

data "aws_iam_policy_document" "eso_assume" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = [var.oidc_provider_arn]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_host}:sub"
      values   = ["system:serviceaccount:external-secrets:external-secrets"]
    }
  }
}

data "aws_iam_policy_document" "eso_inline" {
  statement {
    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
      "secretsmanager:ListSecrets",
    ]
    resources = [
      "arn:aws:secretsmanager:${var.aws_region}:*:secret:${var.secrets_manager_path_prefix}/*",
    ]
  }
  statement {
    actions = ["kms:Decrypt"]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["secretsmanager.${var.aws_region}.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "eso" {
  name               = "greenmetrics-external-secrets-irsa"
  assume_role_policy = data.aws_iam_policy_document.eso_assume.json
  tags               = var.tags
}

resource "aws_iam_role_policy" "eso" {
  name   = "greenmetrics-eso-policy"
  role   = aws_iam_role.eso.id
  policy = data.aws_iam_policy_document.eso_inline.json
}

# --- Argo CD IRSA (read GHCR + minimal AWS read) ---------------------------

data "aws_iam_policy_document" "argocd_assume" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = [var.oidc_provider_arn]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_host}:sub"
      values   = ["system:serviceaccount:argocd:argocd-repo-server"]
    }
  }
}

resource "aws_iam_role" "argocd" {
  name               = "greenmetrics-argocd-irsa"
  assume_role_policy = data.aws_iam_policy_document.argocd_assume.json
  tags               = var.tags
}

# --- Outputs ---------------------------------------------------------------

output "backend_role_arn" { value = aws_iam_role.backend.arn }
output "eso_role_arn"     { value = aws_iam_role.eso.arn }
output "argocd_role_arn"  { value = aws_iam_role.argocd.arn }
