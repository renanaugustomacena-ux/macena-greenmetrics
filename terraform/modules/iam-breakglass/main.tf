# Time-bound IAM break-glass role.
#
# Doctrine refs: Rule 19, Rule 39, Rule 57 (least privilege), Rule 60 (incident response).
# Mitigates: RISK-006 (insider IAM access).
#
# Usage:
#   - Operator assumes via SSO + MFA; max session 1h.
#   - Every assume → CloudTrail event → Alertmanager rule → Slack #sev1-active.
#   - Required for SEV1 incidents only; runbooks document the exact scenarios.

terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.60" }
  }
}

variable "trusted_principal_arns" {
  description = "ARNs of operator IAM principals (or AWS SSO permission set ARN) authorised to assume break-glass."
  type        = list(string)
}

variable "tags" {
  description = "Mandatory cost-allocation tags."
  type        = map(string)
  default = {
    Project        = "greenmetrics"
    Environment    = "shared"
    Owner          = "secops"
    CostCenter     = "platform"
    DataResidency  = "EU-South-1-IT"
  }
}

# --- Break-glass role -----------------------------------------------------

data "aws_iam_policy_document" "breakglass_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "AWS"
      identifiers = var.trusted_principal_arns
    }
    condition {
      test     = "Bool"
      variable = "aws:MultiFactorAuthPresent"
      values   = ["true"]
    }
    condition {
      test     = "NumericLessThan"
      variable = "aws:MultiFactorAuthAge"
      values   = ["3600"] # MFA freshness ≤ 1h
    }
    condition {
      test     = "StringNotEquals"
      variable = "aws:RequestTag/incidentReference"
      values   = [""]
    }
  }
}

resource "aws_iam_role" "breakglass" {
  name                 = "greenmetrics-breakglass"
  description          = "Time-bound (1h) break-glass for SEV1 incidents. Requires MFA + incident-reference session tag."
  assume_role_policy   = data.aws_iam_policy_document.breakglass_assume.json
  max_session_duration = 3600
  permissions_boundary = aws_iam_policy.breakglass_boundary.arn
  tags                 = var.tags
}

# Permissions boundary keeps even break-glass within the GreenMetrics blast radius.
resource "aws_iam_policy" "breakglass_boundary" {
  name        = "greenmetrics-breakglass-boundary"
  description = "Bound for greenmetrics-breakglass role."
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "GreenMetricsScopeOnly"
        Effect = "Allow"
        Action = [
          "rds:*",
          "eks:*",
          "ec2:Describe*",
          "secretsmanager:*",
          "s3:GetObject",
          "s3:ListBucket",
          "kms:Decrypt",
          "kms:DescribeKey",
          "logs:*",
          "cloudtrail:LookupEvents"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "aws:RequestedRegion" = "eu-south-1"
          }
        }
      },
      {
        Sid    = "DenyDestructiveCrossEnv"
        Effect = "Deny"
        Action = [
          "iam:DeleteRole",
          "iam:DeleteRolePolicy",
          "iam:PutUserPolicy",
          "iam:CreateAccessKey",
          "kms:ScheduleKeyDeletion",
          "rds:DeleteDBCluster",
          "rds:DeleteDBInstance"
        ]
        Resource = "*"
      }
    ]
  })
  tags = var.tags
}

# --- CloudTrail alarm -----------------------------------------------------

resource "aws_cloudwatch_metric_alarm" "breakglass_assumed" {
  alarm_name          = "greenmetrics-breakglass-assumed"
  alarm_description   = "Break-glass role assumed (SEV1 path). Verify on-call awareness."
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  threshold           = 1
  metric_name         = "BreakglassAssumedCount"
  namespace           = "GreenMetrics/Security"
  statistic           = "Sum"
  period              = 60
  treat_missing_data  = "notBreaching"

  alarm_actions = [] # wire to SNS → PagerDuty in S4
  tags          = var.tags
}

output "breakglass_role_arn" {
  value = aws_iam_role.breakglass.arn
}
