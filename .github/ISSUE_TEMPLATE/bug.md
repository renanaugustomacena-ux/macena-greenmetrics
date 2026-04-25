---
name: Bug report
about: Report a defect in the Software (NOT a security vulnerability — see SECURITY.md)
title: "bug: <short description>"
labels: ["bug"]
assignees: ["renanaugustomacena-ux"]
---

> **STOP.** If this is a security vulnerability, **do not file a public issue.**
> Follow the responsible-disclosure process in [`SECURITY.md`](../../SECURITY.md).

## What happened

<!-- one-paragraph description -->

## Expected behaviour

<!-- what should have happened -->

## Reproduction steps

1. ...
2. ...
3. ...

## Environment

- Backend version (`curl /api/health | jq .version`):
- Browser / client (if applicable):
- OS:
- Commit SHA you tested against:
- Deployment target (local / staging / production):

## Logs / traces

<!-- attach via gist or upload artefact -->

- Loki query:
- Trace ID (Tempo URL):
- Request ID:

## Doctrine references (if relevant)

<!-- which rule(s) does this defect violate? See ~/.claude/plans/... or docs/PLATFORM-PLAYBOOK.md -->
