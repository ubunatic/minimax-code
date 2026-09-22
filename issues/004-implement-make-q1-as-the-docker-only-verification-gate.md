# 004 — Implement make-q1 as the Docker-only verification gate

**Status**: Open
**Priority**: P0
**Severity**: Critical
**Category**: Testing

---

## Goal

Replace the current host-oriented Q1 placeholder with `make-q1`, a verification command that runs the tool and its tests entirely inside Docker through the approved execution boundary.

## Acceptance criteria

- `make-q1` has a deterministic Docker invocation and does not call `make test`, npm, npx, pnpm, or Node directly on the host.
- The gate builds or selects a pinned image and runs the applicable unit, integration, and policy checks inside it.
- Failure, timeout, signal, and cleanup behavior are deterministic and leave no containers or background processes behind.
- The gate supports synthetic fixtures and temporary data only; it never uses account data or host credentials by default.
- Documentation records the exact command and what it does not prove.

## Progress notes — 2026-09-22

`make-q1` now rebuilds the local Podman image before running `pnpm verify` with
`--network=none`, and `test-q1` delegates to it instead of running the old
host-side `harnez ... make test` recipe. Repository Harnez execution timeout is
configured to `2h` in `config.yaml`; the longer timeout works.

The current Container build reaches `pnpm build` but fails because the checkout
has unresolved source/dependency gaps (37 esbuild resolution errors, including
missing local modules and `zod`). Resolve those baseline build failures before
using the verification gate as evidence. Then add deterministic image
selection/pinning, timeout and signal cleanup tests, and route the gate through
the issue-003 companion CLI.
