# MiniMax Code: Code Map

## Purpose and layout

MiniMax Code is a standalone Node/TypeScript CLI and terminal UI with local and cloud agent runtimes, tools, managed accounts, BYOK models, plugins, and ACP support.

- `packages/` — first-party workspace packages. `packages/agent-modules/*` contains independent capability packages, not a package itself.
- `third_party/` — vendored `pi-mono` workspace packages and `sandbox-runtime` integrations; their tests are outside this distribution's verification.
- `release/` — `extraction.json` defines package roots/source baseline, `public-source.json` inventories publishable files, and `dependency-licenses.json` records declared licenses.
- `scripts/` — build, source-sync, release, inventory, standalone-boundary, and verification tooling; shared helpers live in `scripts/lib/`.
- `test/` — repository-level Node tests plus `vitest-suites.json`, the single declaration of Vitest files and gates.
- `docs/` — architecture, operations, release, contributor, and product documentation.
- `examples/` — small runnable examples (currently `clamp`).
- Root config — `package.json`, `pnpm-workspace.yaml`/lockfile, TypeScript configs, `vitest.oss.config.mjs`, and `Makefile` define the development and local verification toolchain.

## First-party packages

Every package uses `src/index.ts` as its TypeScript entry unless noted; exports below are the public subpaths declared in its manifest.

- `@mavis/agent-core` — runtime contracts and PiTurnRunner assembly; exports protocol, event bridge, runner, tools, bash environment, and prompt-read APIs.
- `@mavis/agent-extension` — built-in adapters for the host-neutral runtime SPI; exports `.`.
- `@mavis/agent-runtime` — per-turn extension SPI, registry, and assembler; exports `.`.
- `@mavis/agent-tools` — platform-owned runtime tools; exports desktop, routing, read-contract, plugin-hooks, and Matrix APIs.
- `@mavis/background-task` — background-task domain model and lifecycle primitives; exports `.`.
- `@mavis/context-manager` — context-compaction lifecycle hook; exports `.`.
- `@mavis/conversation-contract` — runtime-neutral Session, Queue, Turn, and compaction contract; exports `.`.
- `@mavis/cron` — scheduled task capability built on agent-core; exports `.`.
- `@mavis/goal` — thread goals, goal tools, continuation, and budget steering; exports `.`.
- `@mavis/mcp` — MCP substrate, connection pool, naming, and HTTP/stdio transports; exports those runtime APIs.
- `@mavis/permission` — injected-host tool-call permission engine; exports `.`.
- `@mavis/plugin-hooks` — plugin-hook parsing, command execution, and session lifecycle coordination; exports `.`.
- `@mavis/runaway-guard` — bounded turn-local observation and reminder policy; exports `.`.
- `@mavis/session-report` — host-independent session diagnostics and feedback artifacts; exports `.`.
- `@mavis/skills` — filesystem-backed skill registry; exports `.`.
- `@mavis/system-reminder` — framework-agnostic system-reminder assembly; exports `.`.
- `@mavis/browser-core` — provider-neutral browser core for Electron and headless Chrome; exports `.`.
- `@mavis/config` — shared configuration, data-directory, CU-backend, and sandbox-settings APIs.
- `@mavis/local-runtime` — clean local runtime host for desktop and CLI paths; exports `.`.
- `@mavis/local-runtime-v2` — newer local runtime services; exports process-local, CLI service, logging, history, session, turn, and agent APIs.
- `@mavis/protocol` — shared protocol contracts; exports `.` and `./local`.
- `@mavis/shared` — shared domain types and utilities including citations, runtime transport, logging, diagnostics, safety, and model helpers; exports many focused subpaths.
- `@minimax/code` — CLI/TUI product package; its runtime entry is `packages/tui/src/index.ts` (manifest only exports `./package.json`).
- `@mavis/oauth-core` — local OAuth state, credential store, and cross-process coordination; exports `.`.
- `@mavis/oauth-lease-protocol` — local OAuth access-token lease protocol; exports `.`.
- `@mavis/mcode-tools-host` — OAuth lease broker and embedded mcode-tools launcher lifecycle; exports `.`.

## Key scripts and flow

- `scripts/build.mjs` bundles the TUI CLI plus image-preview, mcode-tools, and Matrix MCP entry points with esbuild, resolves workspace exports to checked-in sources, copies runtime/native assets, and writes `dist/`.
- `scripts/gen-tsconfig-paths.mjs` derives standalone TypeScript paths from package exports; `scripts/source-inventory.mjs` checks or writes the public-file inventory.
- `scripts/run-vitest-suite.mjs` runs a named group from `test/vitest-suites.json`; repository Node tests are invoked directly by package scripts.
- `scripts/check-standalone-boundary.mjs` enforces that the standalone bundle stays within the published boundary; `scripts/ci-changes.mjs` classifies documentation-only changes.
- `scripts/export-source-preview.mjs`, `source-candidate.mjs`, `prepare-source-sync.mjs`, and `scripts/lib/source-archive.mjs` support source synchronization and archive validation; `scripts/lib/retired-sources.mjs` enforces retired paths.
- `scripts/release-cli.mjs`, `package-cli-release.mjs`, `publish-cli-release.mjs`, and `verify-cli-release.mjs` assemble, publish, and validate CLI release artifacts.

The normal path is `pnpm build` for `dist/`, focused package scripts for iteration, and `pnpm verify` for the CI-equivalent ordered gates. Verification checks source/inventory and generated paths, exports a source preview, runs release-tool tests, typechecks (full profile), builds, checks standalone boundaries and artifacts, then runs capability, status, smoke, BYOK, policy, and platform-specific sandbox/package gates. `pnpm verify --list` shows applicable gates; profiles (`full`, `platform`, `docs`, `archive`, `package`) control scope. `pnpm gen:tsconfig` regenerates the standalone path map; `pnpm check:tsconfig` validates it.
