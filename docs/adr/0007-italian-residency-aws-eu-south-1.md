# 0007 — Italian residency on AWS eu-south-1

**Status:** Accepted
**Date:** 2026-04-25
**Authors:** @ciupsciups
**Doctrine rules:** 19, 25 (quality threshold), 65 (regulated quality), 67 (transparency)
**Review date:** 2027-04-25

## Context

GreenMetrics processes commercial energy data on Italian soil for tenants subject to GDPR (Reg. UE 2016/679 + D.Lgs. 196/2003), NIS2 (`D.Lgs. 138/2024`), and CSRD (Dir. UE 2022/2464). Data residency on Italian / EU territory is a binding regulatory and customer-trust constraint.

## Decision

Deploy production infrastructure to AWS region `eu-south-1` (Milan). Maintain an Aruba Cloud (Italian sovereign provider) deployment path as documented alternative for tenants requiring Italian-only sovereignty. State backend, RDS, EKS, S3, KMS, Secrets Manager, CloudFront origin all in `eu-south-1`; CloudFront edge nodes operate globally per AWS WAF managed rules.

## Alternatives considered

- **`eu-west-1` Ireland.** Rejected — physically outside Italy; some Italian tenants reject non-Italian residency in procurement.
- **Aruba Cloud only.** Rejected for v1 — operational maturity gap (less mature managed PostgreSQL, no managed K8s, narrower observability ecosystem) outweighs sovereignty benefit for the target tenant profile (Veneto SMEs, not state critical-infrastructure operators). Documented as Phase-2 alt in `terraform/main.tf`.
- **Multi-region active-active across Italy + Ireland.** Rejected per Rule 25 (over-scoping for current scale). Active-passive eu-south-1 is the v1 target (REJ-09).
- **OVHcloud Strasbourg.** Rejected — better French residency, but Italian tenant trust is best served by Italian-territory presence.

## Consequences

### Positive

- Italian residency satisfied at infrastructure layer — no cross-border data flow under default operation.
- AWS managed services (RDS PG16, EKS, KMS, Secrets Manager) accelerate operability.
- CloudFront + WAF protect ingress with DDoS + rate-based rules.
- Argo CD GitOps + Cosign + Kyverno mature in eu-south-1.

### Negative

- AWS lock-in — switching to Aruba requires Terraform module fork. Mitigated: every AWS-specific module wraps a generic concept (state backend, DB, secrets manager) so Aruba modules can ship in parallel.
- eu-south-1 has slightly higher service launch lag than eu-west-1 (AWS prioritises Ireland/Frankfurt for new services).
- Some AWS services not GA in eu-south-1 (revisit annually).

### Neutral

- Cost per resource similar to eu-west-1; minor variance.

## Residual risks

- AWS account compromise → cross-region pivot. Mitigations: IRSA per-pod (Rule 57), MFA on operator IAM, CloudTrail in same region with Object Lock on audit bucket.
- AWS region-wide outage. Mitigations: documented active-passive plan; quarterly DR drill (`docs/runbooks/region-failover.md` lands S4); annual full-region failover.
- Vendor lock-in. Mitigations: Aruba module path documented; ESO abstracts secrets manager; pgx/Timescale runs anywhere.

## References

- CLAUDE.md — Italian residency invariant.
- `terraform/main.tf:127-144` — Aruba Cloud commented stub.
- `~/.claude/plans/my-brother-i-would-flickering-coral.md` §0.4.

## Tradeoff Stanza

- **Solves:** regulatory residency requirement; tenant-trust positioning.
- **Optimises for:** operational maturity, AWS-managed-service velocity.
- **Sacrifices:** AWS vendor lock-in (mitigated by abstraction layer); eu-south-1 service launch lag vs eu-west-1.
- **Residual risks:** account compromise (IRSA + MFA + audit), region outage (active-passive + DR drill), lock-in (Aruba parallel path documented).
