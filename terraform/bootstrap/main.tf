# One-shot Terraform state backend bootstrap.
#
# Doctrine: Rule 23, Rule 63. Mitigates: RISK-018.
#
# Chicken-and-egg: this root uses **local** state to create the S3 bucket + DynamoDB
# lock table that the rest of the project's remote state lives in. After first apply,
# migrate this root's state to a separate "bootstrap" key in the same bucket and
# protect with `prevent_destroy = true`.
#
# Run once per AWS account/region:
#
#   cd terraform/bootstrap
#   terraform init
#   terraform apply
#   # Then in terraform/ root: terraform init -migrate-state

terraform {
  required_version = ">= 1.8.0"
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.60" }
  }
  # Local state for the bootstrap; later migrated to s3://greenmetrics-tf-state-eu-south-1/bootstrap.tfstate
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS region for state backend (Italian residency: eu-south-1 Milan)."
  type        = string
  default     = "eu-south-1"
}

variable "state_bucket_name" {
  description = "S3 bucket name for Terraform state."
  type        = string
  default     = "greenmetrics-tf-state-eu-south-1"
}

variable "lock_table_name" {
  description = "DynamoDB table for state locks."
  type        = string
  default     = "greenmetrics-tf-locks"
}

# --- KMS key for state encryption -------------------------------------------

resource "aws_kms_key" "tf_state" {
  description             = "GreenMetrics Terraform state encryption (RISK-018)"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = {
    Project        = "greenmetrics"
    Environment    = "shared"
    Owner          = "platform-team"
    CostCenter     = "platform"
    DataResidency  = "EU-South-1-IT"
  }
}

resource "aws_kms_alias" "tf_state" {
  name          = "alias/greenmetrics-tf-state"
  target_key_id = aws_kms_key.tf_state.id
}

# --- S3 bucket --------------------------------------------------------------

resource "aws_s3_bucket" "tf_state" {
  bucket = var.state_bucket_name

  lifecycle {
    prevent_destroy = true
  }

  tags = {
    Project        = "greenmetrics"
    Environment    = "shared"
    Owner          = "platform-team"
    CostCenter     = "platform"
    DataResidency  = "EU-South-1-IT"
  }
}

resource "aws_s3_bucket_versioning" "tf_state" {
  bucket = aws_s3_bucket.tf_state.id
  versioning_configuration {
    status     = "Enabled"
    mfa_delete = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "tf_state" {
  bucket = aws_s3_bucket.tf_state.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.tf_state.arn
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_public_access_block" "tf_state" {
  bucket                  = aws_s3_bucket.tf_state.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_policy" "tf_state_tls_only" {
  bucket = aws_s3_bucket.tf_state.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "DenyNonTLS"
      Effect    = "Deny"
      Principal = "*"
      Action    = "s3:*"
      Resource  = [aws_s3_bucket.tf_state.arn, "${aws_s3_bucket.tf_state.arn}/*"]
      Condition = { Bool = { "aws:SecureTransport" = "false" } }
    }]
  })
}

# --- DynamoDB lock table ----------------------------------------------------

resource "aws_dynamodb_table" "tf_locks" {
  name         = var.lock_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.tf_state.arn
  }

  tags = {
    Project        = "greenmetrics"
    Environment    = "shared"
    Owner          = "platform-team"
    CostCenter     = "platform"
    DataResidency  = "EU-South-1-IT"
  }
}

# --- Outputs ----------------------------------------------------------------

output "state_bucket" { value = aws_s3_bucket.tf_state.bucket }
output "lock_table"   { value = aws_dynamodb_table.tf_locks.name }
output "kms_key_arn"  { value = aws_kms_key.tf_state.arn }
output "next_steps" {
  value = <<-EOT
    Bootstrap complete.

    Migrate the bootstrap state itself to S3 (optional but recommended):

      cd terraform/bootstrap
      terraform init -migrate-state \
        -backend-config="bucket=${aws_s3_bucket.tf_state.bucket}" \
        -backend-config="key=bootstrap/terraform.tfstate" \
        -backend-config="region=${var.aws_region}" \
        -backend-config="dynamodb_table=${aws_dynamodb_table.tf_locks.name}" \
        -backend-config="encrypt=true" \
        -backend-config="kms_key_id=alias/greenmetrics-tf-state"

    Then initialise the root project state:

      cd ..
      terraform init
  EOT
}
