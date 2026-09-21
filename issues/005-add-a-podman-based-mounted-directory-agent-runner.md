# 005 — Add a Podman-based mounted-directory agent runner

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Infrastructure
**Related**: [002](002-enforce-docker-only-repository-execution.md), [003](003-add-a-go-companion-cli-as-the-host-boundary.md)

---

## Goal

After the fork-cleanup work, provide a working Podman-based runner that mounts a user-selected directory, lets the operator choose an agent provider, and starts agentic work inside the mounted project. Minimax should be supported as an optional provider alongside the default provider set.

## Acceptance criteria

- A documented command starts the runner with an explicit, validated directory mount and predictable container working directory.
- Provider selection is explicit, supports the configured provider set, and includes Minimax as an optional provider without making it a required dependency.
- The runner passes provider configuration and credentials through a safe, documented mechanism; secrets and host paths are not baked into images or logs.
- Podman invocation has bounded network, privilege, filesystem, and lifecycle defaults, with clear failure and exit-status reporting.
- A smoke/integration test exercises a synthetic mounted directory and verifies that agent startup reaches the selected provider boundary.
- The implementation aligns with the host-boundary and container-execution contracts in issues #002 and #003, and documents any Podman-specific prerequisites or limitations.

**Status**: Draft

---

Reserved placeholder ticket.
