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

## Progress notes — 2026-09-22

The first Container milestone now uses Podman, a non-root Node user, telemetry
opt-out defaults, and a credential-oriented `.dockerignore`. It is explicitly
not the approved host boundary yet: `make install` still runs code-index Go
tooling on the host, and the Make targets invoke Podman directly rather than a
validated companion CLI.

The next implementation must add the bounded Go host controller described in
issue 003, then route verification and runtime commands through it. Keep mount
scope narrow, reject symlink/path escapes, pass an explicit environment
allowlist, and default verification to no network and synthetic data only.
