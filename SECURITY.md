# Security — Vulnerability Disclosure Policy

**Copyright (c) 2026 Renan Augusto Macena. All rights reserved.**

This file describes how to **report** a security vulnerability in
GreenMetrics. The internal security architecture of the Software is
documented at [`docs/SECURITY.md`](./docs/SECURITY.md).

## 1. Supported versions

This is currently a private repository under sole authorship of Renan
Augusto Macena. Security fixes are applied to:

- The `main` branch.
- The most recent release tag.

Older release tags do not receive security backports.

If you are running a fork, an unauthorised redistribution, or any
modified version of the Software, you are responsible for your own
security maintenance and will not receive support.

## 2. Reporting a vulnerability

**Do not open a public issue or pull request for any suspected security
vulnerability.** Public disclosure prior to a coordinated patch puts
users at risk and is incompatible with NIS2 (`D.Lgs. 138/2024`)
responsible-disclosure obligations.

Instead, report privately by **either** of the following channels:

- **GitHub private vulnerability reporting:**
  <https://github.com/renanaugustomacena-ux/macena-greenmetrics/security/advisories/new>
  (preferred when GitHub Security Advisories is enabled on the repository.)

- **Email (PGP-encrypted preferred):**
  <renanaugustomacena@gmail.com> — subject line `[SECURITY] GreenMetrics`.

Include the following information:

- A description of the vulnerability and the affected component(s).
- Reproduction steps, including any proof-of-concept code, payloads, or
  network captures.
- The version, commit SHA, and deployment environment in which the
  vulnerability was observed.
- Your assessment of impact and severity (CVSS v3.1 vector preferred).
- Your contact details and whether you wish to be credited in the
  advisory.

## 3. Triage timeline

The Author commits to the following response timeline. All times are
European business days unless stated otherwise.

| Phase | Target |
|---|---|
| Acknowledge receipt | 2 business days |
| Initial triage + severity classification | 5 business days |
| Status update cadence during investigation | every 7 days |
| Patch development for CRITICAL / HIGH | 30 days |
| Patch development for MEDIUM | 90 days |
| Coordinated public disclosure (after patch + buffer) | 30 days post-patch (negotiable with reporter) |

Where the vulnerability concerns a NIS2-relevant system at a deployed
tenant, the Author will additionally coordinate with the affected
tenant's incident response and may notify ACN within the regulatory
24h/72h window in accordance with `docs/INCIDENT-RESPONSE.md`.

## 4. Coordinated disclosure

Once a patch is available and deployed, a public CVE will be requested
where appropriate, and a security advisory will be published via GitHub
Security Advisories. The reporter will be credited (with their consent)
in the advisory and in the relevant entry of `CHANGELOG.md` under the
`Security` section.

## 5. Safe harbour

The Author will not pursue legal action against, or report to law
enforcement, any researcher who:

- Reports vulnerabilities in good faith and in accordance with this
  policy.
- Acts in accordance with applicable Italian and European Union law
  (including without limitation Codice Penale art. 615-ter, 635-bis,
  640-ter, and Directive 2013/40/EU on attacks against information
  systems).
- Avoids privacy violations, destruction of data, or interruption or
  degradation of any service.
- Does not exfiltrate, copy, retain, or share any production or customer
  data beyond the minimum strictly necessary to demonstrate the
  vulnerability.
- Provides reasonable time for the Author to investigate and remediate
  before any public disclosure.

This safe-harbour is offered in good faith but does not constitute legal
advice or override applicable law. Researchers acting outside this
policy assume their own legal risk.

## 6. Bounty programme

There is no monetary bug-bounty programme at this time. The Author may
elect to recognise significant disclosures with attribution and (at the
Author's sole discretion) a token of appreciation.

## 7. Out of scope

The following are explicitly out of scope for this disclosure policy:

- Findings against forks, mirrors, unauthorised redistributions, or
  modified versions of the Software.
- Findings against third-party services integrated with the Software
  (AWS, GitHub, Sigstore, ISPRA, Terna, etc.); report those to the
  respective vendor.
- Theoretical findings without a working proof-of-concept against the
  current `main` branch or the most recent release tag.
- Denial-of-service findings that require sustained, high-volume traffic
  unless they reveal a structural design flaw.

## 8. Updates to this policy

This policy may be updated from time to time. The version in force at
the time of your report governs the handling of that report.

## 9. Contact

  Renan Augusto Macena
  <renanaugustomacena@gmail.com>

End of vulnerability disclosure policy.
