# MiniMax Code Repository Map

A terminal coding agent with support for MiniMax, your own models, and tools beyond code.

## Directory Structure Overview

```
minimax-code/
├── packages/                    # First-party workspace packages
│   ├── agent-core/             # Agent runtime contracts and PiTurnRunner assembly
│   ├── agent-extension/        # Built-in adapters for Mavis agent modules
│   ├── agent-modules/          # Domain-specific agent capabilities (second-level packages)
│   ├── agent-runtime/          # Pluggable extension layer on top of agent-core
│   ├── agent-tools/            # Platform-owned runtime tool definitions
│   ├── browser-core/           # Provider-neutral browser core for Electron/headless Chrome
│   ├── config/                 # Configuration management
│   ├── local-runtime/          # Local runtime host (v1)
│   ├── local-runtime-v2/       # Local runtime host (v2)
│   ├── mcode-tools-host/       # OAuth lease broker for mcode-tools
│   ├── oauth-core/             # Shared local OAuth state and credential store
│   ├── oauth-lease-protocol/   # OAuth access-token lease protocol
│   ├── protocol/               # Protocol definitions
│   ├── shared/                 # Shared utilities
│   └── tui/                    # CLI and TUI product entry point
├── third_party/                # Vendored upstream packages
│   ├── pi-mono/                # Upstream agent/AI packages
│   └── sandbox-runtime/        # Sandbox runtime
├── test/                       # Repository-level tests
├── scripts/                    # Build and verification tooling
├── release/                    # Release contracts and inventory
├── docs/                       # Human-readable documentation
├── .agents/                    # Agent skills directory
├── examples/                   # Example code
└── issues/                     # Issue tracker (harnez-managed)
```

## Core Agent Packages

### `packages/agent-core` — Agent Runtime Foundation
**Path**: `packages/agent-core/`

**Purpose**: Shared agent runtime contracts and PiTurnRunner assembly layer.

**Key Components**:
- `src/protocol/` — Protocol definitions (agent-message, runtime-event, context-usage)
- `src/pi-turn-runner/` — PiTurnRunner implementation (turn execution engine)
  - `agent.ts` — Agent execution logic
  - `llm.ts` — LLM interaction
  - `llm-retry.ts` — Retry logic for LLM calls
  - `tools.ts` — Tool binding and execution
  - `hooks.ts` — Hook system
  - `metrics.ts` — Metrics collection
  - `terminal.ts` — Terminal interaction
- `src/event-bridge/` — Event bridge for runtime communication
- `src/tools/` — Tool system (define, bind, builtin-defs, input-validation)
- `test/unit/` — Comprehensive unit tests

**When to visit**: For agent execution flow, tool handling, event protocols, or PiTurnRunner implementation.

### `packages/agent-extension` — Host Adapters
**Path**: `packages/agent-extension/`

**Purpose**: Built-in adapters connecting Mavis agent modules to the host-neutral agent-runtime SPI.

**Key Components**:
- `src/index.ts` — Extension registry
- `src/miniapp.ts` — Miniapp integration
- `src/context-manager.ts` — Context management
- `src/permission.ts` — Permission handling
- `src/plan-mode.ts` — Plan mode support
- `src/miniapp-diagnostics.ts` — Diagnostics

**When to visit**: For host integration, context management, permissions, or plan mode features.

### `packages/agent-runtime` — Extension Layer
**Path**: `packages/agent-runtime/`

**Purpose**: Pluggable extension SPI on top of agent-core.

**When to visit**: For extending agent capabilities, assembly patterns, or per-turn extension hooks.

### `packages/agent-tools` — Tool Definitions
**Path**: `packages/agent-tools/`

**Purpose**: Platform-owned runtime tool definitions and helpers.

**Key Areas**:
- Tool specs and definitions
- Shared editing utilities (replace-all-edit)
- Tool validation and helpers

**When to visit**: For built-in tool specifications or tool manipulation utilities.

## Agent Modules — Domain-Specific Capabilities

**Path**: `packages/agent-modules/` — Second-level packages for agent features.

Each module is a self-contained package with focused responsibility:

| Module | Purpose |
|--------|---------|
| `background-task/` | Background task lifecycle and management |
| `context-manager/` | Context window management and compaction |
| `conversation-contract/` | Conversation protocol contracts |
| `cron/` | Cron/scheduler functionality |
| `goal/` | Thread goals and continuation logic |
| `mcp/` | Model Context Protocol integration |
| `permission/` | Permission system and policy enforcement |
| `plugin-hooks/` | Plugin hook system |
| `runaway-guard/` | Runaway protection (limits and guards) |
| `session-report/` | Session reporting |
| `skills/` | Skill system |
| `system-reminder/` | System reminder management |

**When to visit a specific module**: 
- Permission issues → `permission/`
- Context window/token limit problems → `context-manager/`
- Background task failures → `background-task/`
- Goal continuation issues → `goal/`
- MCP integration → `mcp/`
- Cron scheduling → `cron/`
- Runaway behavior → `runaway-guard/`

## Local Runtime Packages

### `packages/local-runtime` — Local Runtime (v1)
**Path**: `packages/local-runtime/`

**Purpose**: Clean local runtime host for desktop and CLI product paths.

**Key Components**:
- `src/agent-host/` — Agent host implementation
- `src/bash/` — Bash subprocess management
- `src/session/` — Session management
- `src/safety/` — Content safety and security
- `src/tool-impl/` — Tool implementations (bash, memory edit, file operations)
- `test/unit/` — Unit tests for runtime behavior

**When to visit**: For local agent host behavior, bash execution, session management, safety policies, or tool implementations.

### `packages/local-runtime-v2` — Local Runtime (v2)
**Path**: `packages/local-runtime-v2/`

**Purpose**: Newer local runtime implementation with enhanced features.

**Key Components**:
- `src/service/` — Service layer
  - `model-system/` — Model selection and catalog
  - `mcp/` — MCP runtime integration
  - `plugin-system/` — Plugin system
  - `sandbox/` — Sandbox integration
  - `session-system/` — Session management
  - `turn-system/` — Turn execution
- `src/application/` — Application layer
- `test/` — Integration and unit tests

**When to visit**: For v2 runtime features, model system, MCP integration, or session fork functionality.

## TUI — Terminal User Interface

### `packages/tui` — MiniMax Code CLI
**Path**: `packages/tui/`

**Purpose**: Minimax Code CLI and TUI product entry point.

**Key Components**:
- `src/tui/` — TUI implementation
  - `features/` — Feature implementations (composer, settings, etc.)
  - `widgets/` — UI widgets (editor, etc.)
  - `controller/` — User interaction handling
- `src/app/` — Application layer
  - `auth-application.ts` — Authentication
  - `provider-application.ts` — Provider management
  - `plugin-application.ts` — Plugin management
- `src/runtime/` — Runtime adapters
- `test/unit/` — Extensive unit tests (100+ test files)

**When to visit**: For CLI features, TUI rendering, authentication, provider setup, or keybinding issues.

## Authentication & OAuth

### `packages/oauth-core`
**Path**: `packages/oauth-core/`

**Purpose**: Shared local OAuth state, credential store, and cross-process coordination.

**When to visit**: For OAuth state management or credential store issues.

### `packages/oauth-lease-protocol`
**Path**: `packages/oauth-lease-protocol/`

**Purpose**: Credential-minimal local OAuth access-token lease protocol.

**When to visit**: For OAuth token leasing or protocol negotiation.

### `packages/mcode-tools-host`
**Path**: `packages/mcode-tools-host/`

**Purpose**: Host-neutral OAuth lease broker and launcher lifecycle for embedded mcode-tools.

**When to visit**: For mcode-tools integration or launcher lifecycle.

## Infrastructure Packages

### `packages/shared`
**Purpose**: Shared utilities and types across packages.

### `packages/protocol`
**Purpose**: Protocol definitions and contracts.

### `packages/config`
**Path**: `packages/config/`

**Purpose**: Configuration management.

**When to visit**: For config file handling, defaults, or settings persistence.

### `packages/browser-core`
**Path**: `packages/browser-core/`

**Purpose**: Provider-neutral browser core shared by Electron and headless Chrome.

## Third-Party Vendored Code

### `third_party/pi-mono/`
Vendored upstream packages from pi-mono with their own licenses.

**Packages**:
- `packages/agent/` — Upstream agent package
- `packages/ai/` — AI utilities
- `packages/coding-agent/` — Coding agent features
- `packages/tui/` — TUI utilities

**Note**: Test suites for vendored code are not part of distribution verification.

### `third_party/sandbox-runtime/`
Vendored sandbox runtime with its own license.

## Testing Infrastructure

### `test/` — Repository-Level Tests
**Path**: `test/`

**Key Files**:
- `vitest-suites.json` — Declaration of every Vitest file (grouped by gate)
- `history-processing.test.ts` — History processing tests
- `sqlite-message-contention.test.ts` — SQLite contention tests
- `executable-resolution.test.ts` — Executable resolution

**Test Gates**:
- `capability` — Main capability tests (200+ files)
- `status-contract` — Status contract tests
- `policy` — Permission policy tests
- `sandbox` — Sandbox-specific tests

**When to visit**: For understanding test organization, gate definitions, or repository-level test patterns.

## Build & Verification Scripts

### `scripts/`
**Path**: `scripts/`

**Key Scripts** (command reference):

| Script | Purpose |
|--------|---------|
| `build.mjs` | Build TypeScript packages |
| `gen-tsconfig-paths.mjs` | Generate TypeScript path configuration |
| `verify.mjs` | Run full verification pipeline |
| `source-inventory.mjs` | Generate `release/public-source.json` |
| `package-cli-release.mjs` | Package CLI for release |
| `release-cli.mjs` | Release CLI with versioning |
| `publish-cli-release.mjs` | Publish CLI release |
| `verify-cli-release.mjs` | Verify CLI release |
| `source-candidate.mjs` | Validate source candidate |
| `prepare-source-sync.mjs` | Prepare upstream source sync |
| `export-source-preview.mjs` | Preview source export |
| `check-standalone-boundary.mjs` | Check standalone bundling boundaries |
| `ci-changes.mjs` | Classify CI-relevant changes |
| `run-vitest-suite.mjs` | Run specific Vitest suite |

**Shared Constants**:
- `lib/package-exports.mjs` — Package export definitions
- `lib/source-archive.mjs` — Source archive handling
- `lib/retired-sources.mjs` — Retired paths tracking

**When to visit**: For understanding build flow, release process, or verification gates.

## Release Configuration

### `release/`
**Path**: `release/`

**Files** (machine-read contracts):

| File | Purpose |
|------|---------|
| `extraction.json` | Source baseline, package scope, product version |
| `public-source.json` | File inventory for publication |
| `dependency-licenses.json` | Declared dependency licenses |

**Key Fields in `extraction.json`**:
- `packageRoots` — Array of all published packages
- `sourceRevision` — Upstream git revision baseline
- `productBaseline` — Product version (e.g., "0.4.12")
- `distribution` — Distribution type (e.g., "standalone-source")

**When to visit**: For understanding published packages, dependencies, or release constraints.

## Documentation

### `docs/`
**Path**: `docs/`

**User Documentation**:
- `README.md` — Documentation overview
- `installation.md` — Installation guide
- `examples.md` — Usage examples
- `architecture.md` — Architecture overview

**Process Documentation**:
- `releasing.md` — Release process
- `release-audit.md` — Release audit checklist
- `source-sync.md` — Source synchronization from upstream
- `verification.md` — Verification pipeline
- `maintainers.md` — Maintainer guide

**Feature Documentation**:
- `tui-capabilities.md` — TUI feature capabilities
- `telemetry.md` — Telemetry and privacy
- `performance-ci.md` — Performance CI setup

**Governance Documentation**:
- `open-source-status.md` — Open source status
- `publication-authorization.md` — Publication authorization rules

**Project Conventions** (harnez-managed):
- `AgenticLoop.md` — 5-phase agentic loop practices
- `IssueTracking.md` — Issue tracking conventions
- `Canary.md` — Canary-first development
- `Git.md` — Git conventions
- `Markdown.md` — Markdown conventions
- `Spec.md` — Spec system

**Exploration Documents**:
- `explore01/` — First exploration phase docs (ignore for this task)
- `explore02/` — Emerging exploration docs

## Root Configuration Files

### Key Files

| File | Purpose |
|------|---------|
| `CLAUDE.md` | Agent guidance and project conventions |
| `AGENTS.md` | Detailed agent system documentation |
| `package.json` | Workspace root configuration |
| `pnpm-workspace.yaml` | pnpm workspace definition |
| `pnpm-lock.yaml` | Dependency lock file |
| `tsconfig.node.json` | TypeScript config for Node tools |
| `tsconfig.standalone.json` | TypeScript config for bundle (auto-generated) |
| `vitest.oss.config.mjs` | Vitest configuration |
| `Makefile` | Build targets (e.g., `make test-q1`) |

### Root Documentation

- `README.md` — Product overview and quick start
- `README_ZH.md` — Simplified Chinese version
- `CONTRIBUTING.md` — Contribution guidelines
- `LICENSE` — MIT license
- `LICENSE-STATUS.md` — License status for dependencies
- `SECURITY.md` — Security reporting
- `THIRD_PARTY_NOTICES.md` — Third-party notices

## Examples

### `examples/`
**Path**: `examples/`

**Contents**:
- `clamp/` — Simple example (clamp utility with tests)
  - `clamp.mjs` — Implementation
  - `clamp.test.mjs` — Tests
  - `README.md` — Documentation

## Agent Skills

### `.agents/skills/`
**Path**: `.agents/skills/`

**Custom Skills** (agent-specific tools):
- `cli-guide/` — CLI guide skill
- `cross-layer-drift-sweep/` — Cross-layer drift detection
- `retro/` — Retrospective framework
- `testing-workflow/` — Testing workflow
- `verify-all-runtime-sinks/` — Runtime sink verification

Each skill has `SKILL.md` defining its interface.

## Issue Tracker

### `issues/`
**Path**: `issues/`

**Harnez-Managed**:
- `README.md` — Index of open issues
- Individual issue files (numbered, e.g., `001.md`)

**Convention**: Use `harnez find`/`harnez issues` instead of `ls` or `find`.

## Critical Paths for Common Tasks

### Finding Code by Feature

| Need | Visit |
|------|-------|
| Agent execution flow | `packages/agent-core/src/pi-turn-runner/` |
| Tool definitions | `packages/agent-tools/src/` |
| Tool handling | `packages/agent-core/src/tools/` |
| TUI rendering/input | `packages/tui/src/tui/` |
| Authentication | `packages/tui/src/app/auth-application.ts` |
| Session management | `packages/local-runtime/src/session/` or `packages/local-runtime-v2/src/application/session/` |
| Bash execution | `packages/local-runtime/src/bash/` |
| Permission policy | `packages/agent-modules/permission/` |
| Context window | `packages/agent-modules/context-manager/` |
| Model selection | `packages/local-runtime-v2/src/service/model-system/` |
| MCP integration | `packages/agent-modules/mcp/` or `packages/local-runtime-v2/src/service/mcp/` |
| Configuration | `packages/config/` |
| OAuth/credential | `packages/oauth-core/` or `packages/oauth-lease-protocol/` |

### Finding Tests

| Test Type | Visit |
|-----------|-------|
| Find test gate | `test/vitest-suites.json` |
| Agent core tests | `packages/agent-core/test/` |
| TUI tests | `packages/tui/test/unit/` (100+ files) |
| Runtime tests | `packages/local-runtime/test/` or `packages/local-runtime-v2/test/` |
| Permission tests | `packages/agent-modules/permission/test/` |
| Config tests | `packages/config/test/` |
| Integration tests | `packages/*/test/integration/` |

### Finding Build/Release Info

| Need | Visit |
|------|-------|
| What gets published | `release/extraction.json` and `release/public-source.json` |
| Build process | `scripts/build.mjs` |
| Verification steps | `scripts/verify.mjs` |
| Release process | `docs/releasing.md` and `scripts/release-cli.mjs` |
| Package exports | Each package's `package.json` `exports` field |
| Dependencies | `release/dependency-licenses.json` |

## Key Concepts

### PiTurnRunner
The core agent execution engine in `packages/agent-core/src/pi-turn-runner/`. Handles one turn of agent execution including LLM calls, tool invocation, and event emission.

### Agent Modules
Domain-specific capabilities in `packages/agent-modules/*`. Each module is independent and can be composed into an agent.

### Mavis
Internal architecture name for the agent system. Used in package names and code comments.

### Source Sync
Three-way merge process (documented in `docs/source-sync.md`) for pulling changes from internal monorepo. Files moved/renamed during sync show as conflicts in public repo.

### Release Extraction
Process defined by `release/extraction.json` that specifies which packages and files are published to the public npm registry.

### Standalone Source
Distribution type meaning "published source code without build artifacts". All TypeScript is compiled but source is published for inspection.

## Notes for Agents

1. **Do not modify generated files**: `release/public-source.json` and `tsconfig.standalone.json` are regenerated by scripts.
2. **Prefer content changes over layout changes**: File moves/renames create sync conflicts with internal monorepo.
3. **Check test gates before adding tests**: Add new test files to `test/vitest-suites.json` rather than `package.json`.
4. **Verify changes match scope**: Check `release/extraction.json` packageRoots to understand published scope.
5. **Run `pnpm verify` before PR**: This runs same gates as CI in same order.
6. **Issue management**: Use `harnez find`/`harnez issues` commands, not `ls issues/` or `find`.
7. **Context discipline**: Prefer grep/range-bounded reads of CLAUDE.md and AGENTS.md over full file reads.
8. **Source sync constraints**: Never cherry-pick internal commits; follow `docs/source-sync.md` for upstream changes.
