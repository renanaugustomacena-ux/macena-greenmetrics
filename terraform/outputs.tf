output "vpc_id" {
  description = "ID of the GreenMetrics VPC."
  value       = aws_vpc.greenmetrics.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs (one per AZ)."
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "Private subnet IDs (workloads + databases)."
  value       = aws_subnet.private[*].id
}

output "timescaledb_port" {
  description = "TimescaleDB port."
  value       = aws_db_instance.timescale.port
}

output "timescaledb_master_password_secret" {
  description = "Random master password (sensitive)."
  value       = random_password.timescale_master.result
  sensitive   = true
}

output "grafana_cloud_hint" {
  description = "Grafana Cloud stack slug, if configured."
  value       = var.grafana_cloud_stack
}
