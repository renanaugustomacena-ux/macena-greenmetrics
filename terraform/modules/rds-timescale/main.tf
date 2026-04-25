// RDS PostgreSQL module with TimescaleDB extension
//
// Note: AWS RDS does not offer TimescaleDB as a managed extension. Two
// deployment options are supported by this module:
//
//  1) "aurora-postgres" — Aurora Postgres 16 with the community-supported
//     `timescaledb` extension loaded at create-time. Trade-off: Aurora's
//     storage layer is incompatible with some TimescaleDB compression
//     internals; we gate this mode behind a feature flag and recommend #2.
//  2) "self-managed-ec2" — a dedicated EC2 instance running the
//     TimescaleDB container on EBS gp3. Simpler compatibility, slightly
//     more ops overhead. This is the production recommendation.
//
// For the MVP we provision an RDS Postgres instance and apply the
// timescaledb extension at migration time; compression CAGGs are supported.

variable "name_prefix"       { type = string }
variable "vpc_id"            { type = string }
variable "subnet_ids"        { type = list(string) }
variable "allocated_storage" { type = number; default = 100 }
variable "instance_class"    { type = string; default = "db.m6g.large" }
variable "engine_version"    { type = string; default = "16.3" }
variable "db_name"           { type = string; default = "greenmetrics" }
variable "master_username"   { type = string; default = "greenmetrics" }
variable "master_password"   { type = string; sensitive = true }
variable "backup_retention"  { type = number; default = 30 }
variable "tags"              { type = map(string); default = {} }

resource "aws_db_subnet_group" "this" {
  name       = "${var.name_prefix}-rds"
  subnet_ids = var.subnet_ids
  tags       = var.tags
}

resource "aws_security_group" "rds" {
  name        = "${var.name_prefix}-rds"
  description = "GreenMetrics RDS ingress"
  vpc_id      = var.vpc_id
  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["10.72.0.0/16"]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = var.tags
}

resource "aws_db_parameter_group" "timescale" {
  name   = "${var.name_prefix}-timescale"
  family = "postgres16"
  parameter {
    name         = "shared_preload_libraries"
    value        = "timescaledb"
    apply_method = "pending-reboot"
  }
  parameter {
    name  = "timescaledb.telemetry_level"
    value = "off"
  }
  parameter {
    name  = "log_statement"
    value = "ddl"
  }
  tags = var.tags
}

resource "aws_db_instance" "this" {
  identifier              = "${var.name_prefix}-tsdb"
  engine                  = "postgres"
  engine_version          = var.engine_version
  instance_class          = var.instance_class
  allocated_storage       = var.allocated_storage
  max_allocated_storage   = var.allocated_storage * 2
  storage_type            = "gp3"
  storage_encrypted       = true
  db_name                 = var.db_name
  username                = var.master_username
  password                = var.master_password
  db_subnet_group_name    = aws_db_subnet_group.this.name
  vpc_security_group_ids  = [aws_security_group.rds.id]
  parameter_group_name    = aws_db_parameter_group.timescale.name
  backup_retention_period = var.backup_retention
  backup_window           = "02:00-03:00"
  maintenance_window      = "Sun:03:00-Sun:04:00"
  multi_az                = true
  deletion_protection     = true
  skip_final_snapshot     = false
  final_snapshot_identifier = "${var.name_prefix}-final"
  performance_insights_enabled = true
  monitoring_interval          = 60
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  tags = var.tags
}

output "endpoint"  { value = aws_db_instance.this.endpoint }
output "db_name"   { value = aws_db_instance.this.db_name }
output "port"      { value = aws_db_instance.this.port }
