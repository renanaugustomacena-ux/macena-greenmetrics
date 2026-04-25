###############################################################################
# GreenMetrics — Infrastructure-as-Code skeleton
#
# Primary target: AWS eu-south-1 (Milan) for Italian data-residency.
# Alternative target (commented below): Aruba Cloud for on-shore sovereignty.
#
# This file ships as a SKELETON — no cloud account is created. Developers
# provide credentials via the standard AWS_ACCESS_KEY_ID / AWS_PROFILE path
# and run `terraform init && terraform plan`.
###############################################################################

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
      CostCenter  = "sustainability"
    }
  }
}

# ----- VPC --------------------------------------------------------------------
resource "aws_vpc" "greenmetrics" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.project}-vpc-${var.environment}"
  }
}

data "aws_availability_zones" "available" { state = "available" }

resource "aws_subnet" "public" {
  count             = length(var.public_subnets)
  vpc_id            = aws_vpc.greenmetrics.id
  cidr_block        = var.public_subnets[count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project}-public-${count.index}"
    Tier = "public"
  }
}

resource "aws_subnet" "private" {
  count             = length(var.private_subnets)
  vpc_id            = aws_vpc.greenmetrics.id
  cidr_block        = var.private_subnets[count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "${var.project}-private-${count.index}"
    Tier = "private"
  }
}

# ----- TimescaleDB (managed via RDS for PostgreSQL — TimescaleDB extension) --
# NOTE: A production-grade deployment uses Timescale Cloud in eu-central-1
# (managed Timescale service). The RDS path below is provided as an alternative
# for teams that prefer a pure-AWS bill.
resource "aws_db_subnet_group" "timescale" {
  name       = "${var.project}-timescale-${var.environment}"
  subnet_ids = aws_subnet.private[*].id
}

resource "random_password" "timescale_master" {
  length  = 24
  special = true
}

resource "aws_db_instance" "timescale" {
  identifier              = "${var.project}-timescale-${var.environment}"
  engine                  = "postgres"
  engine_version          = "16"
  instance_class          = var.db_instance_class
  allocated_storage       = 200
  storage_type            = "gp3"
  storage_encrypted       = true
  db_name                 = "greenmetrics"
  username                = "greenmetrics_admin"
  password                = random_password.timescale_master.result
  db_subnet_group_name    = aws_db_subnet_group.timescale.name
  skip_final_snapshot     = var.environment != "production"
  publicly_accessible     = false
  deletion_protection     = var.environment == "production"
  backup_retention_period = var.environment == "production" ? 30 : 7
  multi_az                = var.environment == "production"
  performance_insights_enabled = true
  monitoring_interval          = 60

  # Note: enable the TimescaleDB shared_preload_libraries via parameter group.
  parameter_group_name = aws_db_parameter_group.timescale.name
}

resource "aws_db_parameter_group" "timescale" {
  name   = "${var.project}-timescale-${var.environment}"
  family = "postgres16"

  parameter {
    name         = "shared_preload_libraries"
    value        = "pg_stat_statements,timescaledb"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "timescaledb.telemetry_level"
    value = "off"
  }
}

# ----- Grafana Cloud (optional) ----------------------------------------------
# When grafana_cloud_stack is set, wire Grafana Cloud as the primary dashboarding
# backend. Otherwise the docker-compose self-hosted Grafana is the target.
provider "grafana" {
  cloud_access_policy_token = ""
  url                       = var.grafana_cloud_stack != "" ? "https://${var.grafana_cloud_stack}.grafana.net/" : ""
  alias                     = "cloud"
}

# ----- Alternative: Aruba Cloud ----------------------------------------------
#
# Aruba Cloud (Italian provider — data-residency in Arezzo and Bergamo) is the
# preferred choice for public-administration customers. Uncomment and populate
# with real credentials + resource definitions when using Aruba instead of AWS:
#
# provider "aruba" {
#   username = var.aruba_username
#   password = var.aruba_password
#   url      = "https://api.dc1.computing.cloud.it"
# }
#
# resource "aruba_virtual_server" "greenmetrics_api" {
#   name            = "${var.project}-api-${var.environment}"
#   os_template_id  = 3145
#   package_id      = "LARGE"
#   hypervisor_type = "vmware"
# }

# ----- Outputs consumed by the app's env -------------------------------------
output "timescaledb_endpoint" {
  description = "TimescaleDB (RDS Postgres) endpoint."
  value       = aws_db_instance.timescale.endpoint
}
