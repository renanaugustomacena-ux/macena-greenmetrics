// VPC module — isolates GreenMetrics workloads in a private network.
// Region: eu-south-1 (Milan) per v2.0 §11 "EU residency".

variable "name_prefix" {
  type    = string
  default = "greenmetrics"
}

variable "cidr_block" {
  type    = string
  default = "10.72.0.0/16"
}

variable "az_count" {
  type    = number
  default = 3
}

variable "tags" {
  type    = map(string)
  default = {}
}

data "aws_availability_zones" "this" {
  state = "available"
}

resource "aws_vpc" "this" {
  cidr_block           = var.cidr_block
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags                 = merge(var.tags, { Name = "${var.name_prefix}-vpc" })
}

resource "aws_subnet" "public" {
  count             = var.az_count
  vpc_id            = aws_vpc.this.id
  cidr_block        = cidrsubnet(var.cidr_block, 4, count.index)
  availability_zone = data.aws_availability_zones.this.names[count.index]
  tags              = merge(var.tags, { Name = "${var.name_prefix}-public-${count.index}", "kubernetes.io/role/elb" = "1" })
}

resource "aws_subnet" "private" {
  count             = var.az_count
  vpc_id            = aws_vpc.this.id
  cidr_block        = cidrsubnet(var.cidr_block, 4, count.index + 8)
  availability_zone = data.aws_availability_zones.this.names[count.index]
  tags              = merge(var.tags, { Name = "${var.name_prefix}-private-${count.index}", "kubernetes.io/role/internal-elb" = "1" })
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id
  tags   = merge(var.tags, { Name = "${var.name_prefix}-igw" })
}

output "vpc_id" { value = aws_vpc.this.id }
output "public_subnet_ids" { value = aws_subnet.public[*].id }
output "private_subnet_ids" { value = aws_subnet.private[*].id }
