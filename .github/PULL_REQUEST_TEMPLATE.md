# Pull Request

## Summary

<!-- 1-3 sentences. What changes, why now. -->

## Linked issue / ADR

<!-- Required for any change in /k8s, /terraform, /backend/migrations, /api/openapi/, /.github/workflows/, /policies/. -->

- Issue:
- ADR:

## Doctrine rules touched

<!-- Cite by number. See /home/renan/.claude/plans/my-brother-i-would-flickering-coral.md §1 for the matrix. -->

- Platform (9–28):
- Backend (29–48):
- DevSecOps (49–68):

## Doctrine rules rejected

<!-- If this PR introduces a pattern explicitly rejected in §3 of the plan, cite the override ADR. Otherwise leave blank. -->

## Risk acknowledged

<!-- For non-trivial change, fill the four-part stanza. -->

- **Solves:**
- **Optimises for:**
- **Sacrifices:**
- **Residual risks:**

## Backend addendum (if /backend/ changed)

<!-- Required for backend PRs. -->

- **Domain modelling notes:**
- **Consistency model (strong / eventual / idempotent):**
- **Failure mode plan:**
- **Latency budget impact:**
- **Trust boundary touched (cite docs/TRUST-BOUNDARIES.md):**
- **Doctrine checklist signed:** [ ] yes — see comment

## Runbook update

<!-- For ops-touching changes, link or create the relevant runbook in docs/runbooks/. -->

- Runbook:

## CHANGELOG

<!-- For any /api/v1/ contract change. Keep-a-Changelog format. -->

- Entry:

## Verification checklist

- [ ] CI green (lint / test / build / security / policy gate).
- [ ] `pre-commit run --all-files` clean locally.
- [ ] Coverage ≥ 80% on changed packages (≥ 90% on `internal/domain/`).
- [ ] No new HIGH/CRITICAL findings (Trivy / govulncheck / osv-scanner / Semgrep / CodeQL).
- [ ] Conformance suite green (`make test-conformance`).
- [ ] No regressions on the SLO regression bench (if perf-touching).
- [ ] If changing migrations: `make migrate-up && make migrate-down N=1 && make migrate-up` reversible against testcontainer.
- [ ] If changing /k8s or /terraform: `make policy-check` green.
- [ ] If changing /api/openapi/: `redocly lint` green and `tests/contracts/v1_compat_test.go` green.
- [ ] Documentation touched alongside code in its domain.
