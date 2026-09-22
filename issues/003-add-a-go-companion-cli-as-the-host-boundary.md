# 003 — Add a Go companion CLI as the host boundary

**Status**: Open
**Priority**: P0
**Severity**: Critical
**Category**: Architecture

---

## Goal

Add a small Go companion CLI that is the sole approved host-side controller for this repository, modeled on the existing `../project/kernel/kpatch` pattern when that reference is available. It should validate inputs, construct safe Docker invocations, and expose bounded repository operations without executing project tooling directly on the host.

## Acceptance criteria

- The CLI has a documented command boundary and explicit allowlist of operations.
- Docker invocation uses fixed image, mounts, working directory, environment, and network policy defaults.
- Arguments and paths are validated against traversal, shell injection, and accidental broad mounts.
- The CLI reports container exit status and logs without leaking secrets.
- Unit and integration tests cover both permitted commands and rejected host-execution paths.

## Pre-work for implementation

The repository now has a preliminary Podman `Containerfile` and Make targets,
but they are deliberately not considered the host boundary. Implement the
companion CLI before claiming issue 002 or 004 complete.

Start with one allowlisted operation for the verification path. Use argument
arrays with `exec.Command`, validate the repository root and every mount, reject
root/home/broad mounts and traversal or symlink escapes, and propagate exit
status and signals. Define the bootstrap path for obtaining the CLI without
running project Node tooling on the host.
