<!-- harnez:begin Local Overlays -->
- Local ephemeral overrides: @AGENTS.local.md
<!-- harnez:end Local Overlays -->

# Agent guide

This repository is the reviewed public projection of an internal monorepo, not an ordinary workspace. Every published file is listed in `release/public-source.json`, and upstream changes arrive through a three-way merge described in `docs/source-sync.md`. Moving or renaming files therefore has a cost that a normal repository does not have: it shows up as a conflict or an unreviewed new file at the next synchronization. Prefer changing content over changing layout.

## Layout

- `packages/` — first-party workspace packages. `packages/agent-modules/*` is a second level of packages, not a package itself.
- `third_party/` — vendored upstream packages (`pi-mono`, `sandbox-runtime`) with their own licenses. Their test suites are not part of this distribution's verification.
- `release/` — machine-read release contracts. `extraction.json` pins the source baseline and package scope, `public-source.json` is the file inventory, `dependency-licenses.json` records declared dependency licenses. Build, type-check, test resolution, and source checks all read from here.
- `docs/` — human-readable documentation and supporting media.
- `scripts/` — build and verification tooling. Shared constants live in `scripts/lib/`; import them instead of repeating literal paths or lists.
- `test/` — repository-level tests and `vitest-suites.json`, the declaration of every Vitest file this distribution runs.

## Generated files

Do not edit these by hand; regenerate them and commit the result.

| File | Regenerate with | Checked by |
| --- | --- | --- |
| `release/public-source.json` | `node scripts/source-inventory.mjs --write` | `pnpm check:source` |
| `tsconfig.standalone.json` (the `paths` block) | `pnpm gen:tsconfig` | `pnpm check:tsconfig` |

Review added, removed, and renamed files before regenerating the inventory: recording a file does not make it suitable for publication. Content-only edits to existing files do not require inventory regeneration. Keep private review material, verification reports, and temporary artifacts outside the repository; the inventory scans the working tree, including untracked files outside its explicit exclusions.

## Branches and source synchronization

Use a feature branch and submit a pull request; do not push directly to the default branch. See `CONTRIBUTING.md` for contribution and review requirements.

Name new branches by the purpose of the change: `feat/<short-description>` for features, `fix/<short-description>` for bug fixes, and corresponding prefixes such as `docs/`, `refactor/`, `test/`, or `chore/` for other work. Use concise English descriptions in lowercase kebab-case, for example `feat/provider-limits` or `fix/session-restore`. Do not use agent or tool names as branch prefixes, including `codex/`. Follow an explicitly requested branch name when one is provided.

Never merge internal Git history or cherry-pick internal commits into this repository. Follow `docs/source-sync.md`, keep unreviewed candidates outside the repository, and apply reviewed files individually. Advance `release/extraction.json`'s `sourceRevision` only after reviewing all differences for the selected source revision.

## Single sources of truth

| Concern | Declared in | Consumed by |
| --- | --- | --- |
| Package scope | `release/extraction.json` (`packageRoots`) | build, type-check paths, Vitest aliases, source check, source sync |
| Package export → source file | each package's `exports` via `scripts/lib/package-exports.mjs` | `tsconfig.standalone.json`, `vitest.oss.config.mjs` |
| Vitest files per gate | `test/vitest-suites.json` | `vitest.oss.config.mjs`, `scripts/run-vitest-suite.mjs` |
| Retired source paths | `scripts/lib/retired-sources.mjs` | `check:source` (must not exist), `check:standalone` (must not be bundled) |
| Verification pipeline | `scripts/verify.mjs` | GitHub CI, `pnpm verify` |
| Documentation-only classification | `scripts/ci-changes.mjs` | source verification, release audit |
| Source archive validation/extraction | `scripts/lib/source-archive.mjs` | source export, candidate validation |

## Common changes

Adding a workspace package: add it to `pnpm-workspace.yaml` and to `packageRoots` in `release/extraction.json`, then run `pnpm gen:tsconfig` and `node scripts/source-inventory.mjs --write`.

Changing package exports: run `pnpm gen:tsconfig` after adding or changing an export subpath.

Adding a Vitest file: add its path to the appropriate group in `test/vitest-suites.json`. Do not hard-code Vitest file paths in `package.json` scripts or the Vitest config. Repository-level `node:test` suites remain in their existing gates; add workflow safety and release-tool regressions to `test/source-sync.test.mjs`.

Adding a verification gate: add a step to `scripts/verify.mjs`, with `platforms` when it cannot run everywhere. Do not add steps to the workflow file.

## Verification

`pnpm verify` runs the same gates as CI in the same order; `pnpm verify --list` shows which apply on the current platform. Run it before opening a pull request. Individual gates such as `pnpm typecheck`, `pnpm build`, and `pnpm test:byok` remain available for iteration.

The full profile is the local default. `platform` omits only duplicate type checking. Use `pnpm verify --profile docs` only when every changed path qualifies under `scripts/ci-changes.mjs`; it runs source inventory, generated-path, source-export, and release-tool checks. `AGENTS.md`, bundled runtime prompts, and `release/` changes do not qualify for that profile. `archive` skips Git export for source archives; the candidate workflow authenticates the archive before invoking it. Keep profile selection in the shared verifier.

Source export reads committed `HEAD` and rejects uncommitted tracked changes. During editing, run the relevant individual gates; run the complete applicable profile on the reviewed commit with a clean tracked working tree before opening a PR. Report the checks actually run and any blocked or untested boundaries. Offline tests do not establish live-service or cross-platform acceptance.

## Boundaries

Do not reference internal hosts, generated IDL, or private services; `check:source` catches known patterns but does not replace publication review. Do not restore paths listed in `scripts/lib/retired-sources.mjs` or remove supported capabilities to make standalone checks pass. Do not commit account data, sessions, logs, credentials, or real user content; use temporary data directories and synthetic test inputs. Documentation and commit messages are written in English; preserve the required languages of localized product strings and bundled runtime prompts. See `CONTRIBUTING.md`.
<!-- harnez:begin Harnez Managed Conventions -->
## Harnez Managed Conventions

Managed by harnez — local edits here are overwritten on the next `harnez init`.
Put project-specific rules outside this block.

### Editing Discipline
- Prefer structured patch tools (`apply_patch`) or whole-block replacements over
  narrow string substitution edits.
- When making multi-line edits, ensure sufficient surrounding context lines to
  avoid ambiguous pattern matches.
- **Reading & Context Discipline (Recommended for Large Files)**: Prefer
  `harnez read -I <file>` (dense visual PNG context card) or
  `harnez read -L <range>` / `harnez read -n` for medium/large files (>100 lines)
  to preserve token quota and prevent context fatigue. Native reads remain valid
  for targeted inspection; hook-level blocking is conditional on the active
  `reading_discipline.enforce` mode in `~/.harnez/config.yaml` (or
  `HARNEZ_READ_ENFORCE`).

### Issue Tracker Discovery (harnez find)
Applies when this project has an `issues/` tracker. To search existing issues,
compute the next ticket number, or allocate one, use `harnez find` / `harnez issues`
instead of `ls issues/`, `find`, or raw grep:
- `harnez find -d <repo> issues -a status:open` — list active open issues (use `-I` for visual overview PNG card)
- `harnez find -d <repo> issues "<query>"` — fuzzy search across titles and body text (use `-I` for visual overview)
- `harnez issues show -d <repo> <n> -I` — render single issue as styled visual PNG card (inspect via `view_file`)
- `harnez find -d <repo> issues next` — report the next free ticket number (read-only)
- `harnez issues new -d <repo> "<title>"` — atomically reserve that number and create
  a placeholder ticket file; write the ticket to the printed path
- `harnez issues <verb> -d <repo> <n> [reason]` — change a ticket's status, resync
  `issues/README.md`, and commit, in one call
- `harnez index -d <repo>` — update `issues/README.md` after filing or updating tickets
- Commit documentation and `issues/*.md` changes immediately; don't batch them behind
  pending code work.

### Agentic Loop Invariants
Where `@docs/AgenticLoop.md` is present in this project, follow it rather than
restating it here — in particular Invariant 1 (Parallel Read, Sequential Write:
one writer per workspace), Invariant 3 (Zero Zombie Guarantee: track and terminate
every background task and subagent), Invariant 6 (Context Discipline: no whole-file
reads of AGENTS.md/CLAUDE.md — grep or range-bounded reads), and Invariant 10
(Media & Demo Verification Gate: explicit user confirmation before publishing
recordings or screenshots).
<!-- harnez:end Harnez Managed Conventions -->
<!-- harnez:begin Language Conventions -->
Adhere to the following conventions.

Docs in `./docs/` are managed by harnez. <!-- harnez:bundled -->

- Issue Tracking Practices @docs/IssueTracking.md,
  P0-P3 priorities, metadata headers (Status, Priority, Severity, Category), tracker sync
- Agentic Loop Practices @docs/AgenticLoop.md,
  5-phase loop (Advisory -> Dev -> Review -> Hygiene -> Retro), zero zombie guarantee
- Canary-first development @docs/Canary.md,
  probe external mechanisms before building features on them
- Code Exploration @docs/CodeIndexGuide.md,
  use `code-index` CLI for fast repository navigation without grepping
- Git @docs/Git.md,
  conventional commits, work on the default branch, don't push unless asked
- Markdown @docs/Markdown.md,
  PascalCase for evergreens, kebab-case for ephemeral docs; ASCII art in chat, Mermaid only in docs/
- Spec system @docs/Spec.md,
  YAML spec files as single source of truth; Go code must not duplicate spec values
<!-- harnez:end Language Conventions -->
<!-- harnez:begin Quota-1 Guardrails -->
## Quota-1 Guardrails

- **Single-Test Boundary**: Under Quota-1 rules, the agent may only run the test suite once per step/turn.
- **Code Modification Required**: If tests fail or complete, you MUST modify repository source files before running tests again. Repeated test runs without intermediate code modifications are blocked.
- **Enforced Test Target**: Execute tests via `make test-q1` (or `harnez exec --quota-1 -- <test-cmd>`).
- **Unauthorized Bypass Forbidden**: Bypassing guardrails via `QUOTA_BYPASS=1` or `HARNEZ_QUOTA_BYPASS=1` is strictly reserved for human developers and CI environments. Agent loops must not set or pass bypass flags.
<!-- harnez:end Quota-1 Guardrails -->
