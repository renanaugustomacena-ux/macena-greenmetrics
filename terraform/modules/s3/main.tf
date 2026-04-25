// S3 buckets for GreenMetrics: reports (generated HTML/PDF), backups, audit
// archives. EU-only, TLS-only, SSE-KMS.

variable "name_prefix" { type = string }
variable "tags"        { type = map(string); default = {} }

resource "aws_kms_key" "this" {
  description             = "${var.name_prefix} S3 encryption"
  deletion_window_in_days = 30
  enable_key_rotation     = true
  tags                    = var.tags
}

locals {
  buckets = {
    reports = {
      versioning = true
      lifecycle_days = 3650  // ESRS assurance trail 10y
    }
    backups = {
      versioning = true
      lifecycle_days = 90
    }
    audit = {
      versioning = true
      lifecycle_days = 1825  // 5y minimum per GDPR evidentiary need
    }
  }
}

resource "aws_s3_bucket" "this" {
  for_each      = local.buckets
  bucket        = "${var.name_prefix}-${each.key}"
  force_destroy = false
  tags          = merge(var.tags, { Purpose = each.key })
}

resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  for_each = local.buckets
  bucket   = aws_s3_bucket.this[each.key].id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.this.arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_versioning" "this" {
  for_each = { for k, v in local.buckets : k => v if v.versioning }
  bucket   = aws_s3_bucket.this[each.key].id
  versioning_configuration { status = "Enabled" }
}

resource "aws_s3_bucket_public_access_block" "this" {
  for_each                = local.buckets
  bucket                  = aws_s3_bucket.this[each.key].id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_policy" "tls_only" {
  for_each = local.buckets
  bucket   = aws_s3_bucket.this[each.key].id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid = "DenyInsecureTransport",
      Effect = "Deny",
      Principal = "*",
      Action = "s3:*",
      Resource = [
        aws_s3_bucket.this[each.key].arn,
        "${aws_s3_bucket.this[each.key].arn}/*",
      ],
      Condition = { Bool = { "aws:SecureTransport" = "false" } }
    }]
  })
}

resource "aws_s3_bucket_lifecycle_configuration" "this" {
  for_each = local.buckets
  bucket   = aws_s3_bucket.this[each.key].id
  rule {
    id     = "${each.key}-retention"
    status = "Enabled"
    expiration { days = each.value.lifecycle_days }
    noncurrent_version_expiration { noncurrent_days = 30 }
  }
}

output "bucket_names" { value = { for k, b in aws_s3_bucket.this : k => b.bucket } }
output "kms_key_arn"  { value = aws_kms_key.this.arn }
