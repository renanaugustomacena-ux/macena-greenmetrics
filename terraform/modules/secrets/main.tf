// AWS Secrets Manager — the canonical source for JWT_SECRET, DATABASE_URL,
// OCPP_CENTRAL_SYSTEM_URL, PULSE_WEBHOOK_SECRET, and downstream integration
// credentials (SDI, Terna, E-Distribuzione).

variable "name_prefix" { type = string }
variable "tags"        { type = map(string); default = {} }

locals {
  secret_keys = toset([
    "jwt-secret",
    "database-url",
    "grafana-admin-password",
    "pulse-webhook-secret",
    "ocpp-central-system-url",
    "terna-api-token",
    "edistribuzione-api-cert",
  ])
}

resource "aws_secretsmanager_secret" "this" {
  for_each = local.secret_keys
  name     = "${var.name_prefix}/${each.key}"
  tags     = merge(var.tags, { Key = each.key })
}

resource "aws_secretsmanager_secret_version" "placeholder" {
  for_each      = local.secret_keys
  secret_id     = aws_secretsmanager_secret.this[each.key].id
  secret_string = "REPLACE_ME"
  lifecycle {
    ignore_changes = [secret_string]
  }
}

output "secret_arns" { value = { for k, s in aws_secretsmanager_secret.this : k => s.arn } }
