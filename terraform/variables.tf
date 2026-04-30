variable "aws_region" {
  description = "AWS region. eu-south-1 = Milan — mandatory for Italian data-residency commitments."
  type        = string
  default     = "eu-south-1"
}

variable "environment" {
  description = "Environment tag: development | staging | production."
  type        = string
  default     = "development"
}

variable "project" {
  description = "Project identifier used for resource tagging and naming."
  type        = string
  default     = "greenmetrics"
}

variable "vpc_cidr" {
  description = "CIDR block for the GreenMetrics VPC."
  type        = string
  default     = "10.80.0.0/16"
}

variable "public_subnets" {
  description = "CIDRs of public subnets, one per AZ."
  type        = list(string)
  default     = ["10.80.10.0/24", "10.80.20.0/24", "10.80.30.0/24"]
}

variable "private_subnets" {
  description = "CIDRs of private subnets for workloads and databases."
  type        = list(string)
  default     = ["10.80.110.0/24", "10.80.120.0/24", "10.80.130.0/24"]
}

variable "timescale_version" {
  description = "TimescaleDB image tag. Use latest-pg16 for production."
  type        = string
  default     = "latest-pg16"
}

variable "db_instance_class" {
  description = "RDS instance class for managed TimescaleDB (if Aiven/Timescale Cloud is not used)."
  type        = string
  default     = "db.m7g.large"
}

variable "grafana_cloud_stack" {
  description = "Grafana Cloud stack slug. Empty to use self-hosted Grafana."
  type        = string
  default     = ""
}

variable "tenant_domain" {
  description = "Root DNS domain for tenant sub-domains (e.g. *.greenmetrics.it)."
  type        = string
  default     = "greenmetrics.it"
}

variable "allowed_ingest_cidrs" {
  description = "CIDR ranges authorised to hit the ingestor API (industrial gateways)."
  type        = list(string)
  default     = ["0.0.0.0/0"]
}
