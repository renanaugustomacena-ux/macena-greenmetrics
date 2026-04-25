terraform {
  required_version = ">= 1.8.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.42"
    }
    # Alternative for Italian-sovereignty deployments:
    # aruba = { source = "ArubaCloud/aruba", version = "~> 0.4" }
    grafana = {
      source  = "grafana/grafana"
      version = "~> 3.8"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # Remote state — S3 + DynamoDB lock + KMS encryption.
  # Doctrine: Rule 23 (tooling discipline), Rule 56 (automation), Rule 63 (immutable infra).
  # Mitigates: RISK-018.
  # Bootstrap: see terraform/bootstrap/ for one-shot creation of the bucket + lock table.
  # ADR: docs/adr/0007-italian-residency-aws-eu-south-1.md.
  backend "s3" {
    bucket         = "greenmetrics-tf-state-eu-south-1"
    key            = "greenmetrics/terraform.tfstate"
    region         = "eu-south-1"
    encrypt        = true
    kms_key_id     = "alias/greenmetrics-tf-state"
    dynamodb_table = "greenmetrics-tf-locks"
    use_lockfile   = true
  }
}
