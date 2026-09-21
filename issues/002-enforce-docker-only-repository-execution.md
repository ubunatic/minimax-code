# 002 — Enforce Docker-only repository execution

**Status**: Open
**Priority**: P0
**Severity**: Critical
**Category**: Security

---

## Goal

Make Docker the only execution environment for repository tools and tests. Nothing in the repository may trigger Node.js, npm, npx, pnpm, build, test, network, or mutation behavior on the host unless explicitly launched through the approved Go companion CLI.

## Acceptance criteria

- Host-side entry points fail closed or delegate into Docker.
- Docker images and scripts provide the complete build, test, and runtime environment.
- Host execution is covered by tests or canaries proving that disallowed commands are not run directly.
- Documentation clearly distinguishes safe host orchestration from container execution.
- Secrets, credentials, sessions, and user data are not copied into images or committed artifacts.
