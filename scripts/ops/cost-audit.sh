#!/usr/bin/env bash
# Monthly cost audit — list waste candidates across AWS + K8s + Grafana.
# Doctrine refs: Rule 22.
# Output: stdout markdown report; redirected to GitHub issue per docs/runbooks/cost-audit.md.

set -Eeuo pipefail

REGION="${AWS_REGION:-eu-south-1}"

echo "# GreenMetrics cost audit — $(date -u +%Y-%m-%d)"
echo
echo "Scope: AWS region ${REGION}; K8s + Grafana in cluster."
echo

# --- 1. Unattached EBS volumes (> 7d) ---

echo "## Unattached EBS volumes (potential waste)"
CUTOFF_7D="$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ)"
aws ec2 describe-volumes --region "$REGION" \
  --filters Name=status,Values=available \
  --query "Volumes[?CreateTime<\`${CUTOFF_7D}\`].[VolumeId,Size,VolumeType,CreateTime,Tags[?Key==\`Name\`].Value|[0]]" \
  --output table || echo "(no EBS API access or none found)"
echo

# --- 2. Idle RDS replicas ---

echo "## RDS instances with low CPU (consider scaling down)"
aws rds describe-db-instances --region "$REGION" \
  --query 'DBInstances[].[DBInstanceIdentifier,DBInstanceClass,DBInstanceStatus]' \
  --output table || true
echo "(Verify CloudWatch CPUUtilization < 5% over 30d for each instance.)"
echo

# --- 3. Orphaned secrets ---

echo "## Secrets Manager — last accessed > 90 days ago"
CUTOFF_90D="$(date -u -d '90 days ago' +%Y-%m-%dT%H:%M:%SZ)"
aws secretsmanager list-secrets --region "$REGION" \
  --query "SecretList[?LastAccessedDate<\`${CUTOFF_90D}\`].[Name,LastAccessedDate,LastChangedDate]" \
  --output table || true
echo

# --- 4. Unused IAM roles ---

echo "## IAM roles with no AssumeRole event in 90d"
aws cloudtrail lookup-events --region "$REGION" \
  --lookup-attributes AttributeKey=EventName,AttributeValue=AssumeRole \
  --start-time "$(date -u -d '90 days ago' +%Y-%m-%dT%H:%M:%SZ)" \
  --query 'Events[].UserIdentity.Arn' --output table 2>/dev/null \
  | sort -u || true
echo "(Cross-check against `aws iam list-roles` for unused.)"
echo

# --- 5. K8s underutilised nodes ---

echo "## K8s nodes (verify CPU < 20% sustained 7d on Grafana)"
if command -v kubectl >/dev/null 2>&1; then
  kubectl top node 2>/dev/null || echo "(metrics-server unavailable)"
else
  echo "(kubectl not in PATH)"
fi
echo

# --- 6. Grafana datasources with zero queries 30d ---

echo "## Grafana datasources idle 30d"
echo "(Manual check: Grafana → Configuration → Data sources → Last queried column)"
echo

echo
echo "---"
echo
echo "Triage per docs/runbooks/cost-audit.md."
