# Harness Tools and Product Surface

## Scope and evidence standard

This is a source-based analysis of the public projection in this repository. File paths below are evidence; claims about Claude Code, Codex, or Agy are hypotheses unless this repository directly measures them. No benchmark or competitor source is included here.

## Product shape and execution flow

The documented execution path is `TUI / exec / ACP → CliService → local applications → Session / Turn / Agent services → Pi / model providers / local tools` ([`docs/architecture.md`](architecture.md)). The important product property is one runtime behind several surfaces: the terminal UI, a one-shot/headless command, and an Agent Client Protocol stdio server. The CLI registers these in [`packages/tui/src/cli/program.ts`](../packages/tui/src/cli/program.ts), while the CLI architecture guide identifies the adapter boundary and runtime entry points ([`.agents/skills/cli-guide/SKILL.md`](../.agents/skills/cli-guide/SKILL.md)).

| Surface | Evidence-backed capability |
|---|---|
| Interactive TUI | Prompt, model selection, session picker/continue, regular/fullscreen modes, queued steering, plan mode, permission controls ([`README.md`](../README.md), [`packages/tui/src/cli/contract.ts`](../packages/tui/src/cli/contract.ts)). |
| Headless `exec` | Explicit cwd/input/files, model and effort overrides, prompt mode, session continuation, config path, permission policy, timeout, step cap, text/JSON/stream-JSON output, diagnostics, output schema, and final-message file ([`packages/tui/src/cli/contract.ts`](../packages/tui/src/cli/contract.ts)). |
| ACP | stdio Agent Client Protocol server, with the same runtime lane selected by the CLI ([`packages/tui/src/cli/program.ts`](../packages/tui/src/cli/program.ts)). |
| Session/runtime | Local applications own sessions, queues, and interactions; the architecture explicitly avoids an HTTP front door or daemon boundary in this projection ([`docs/architecture.md`](architecture.md)). |

## Built-in tools and strong defaults

The base LLM-facing tool set is intentionally small: `read`, `write`, `edit`, `bash`, `grep`, `glob`, TODO state, and task controls, plus web fetch/deploy ([`packages/config/src/agent-capabilities.ts`](../packages/config/src/agent-capabilities.ts)). The tool definitions encode operational guidance, not just schemas: edit existing files rather than overwrite, read immediately before editing, use exact unique replacements, re-read after edits, and avoid unsolicited documentation ([`packages/agent-core/src/tools/builtin-defs.ts`](../packages/agent-core/src/tools/builtin-defs.ts)). `grep` and `glob` respect `.gitignore` and return bounded output, while reads page large files and support selected media ([same file](../packages/agent-core/src/tools/builtin-defs.ts)).

This is a likely quality multiplier: the model receives executable workflow rules at the tool boundary, where they remain visible whenever the tool is selected. It reduces stale-edit failures and context waste without requiring a separate reviewer. That causal claim is an inference, not a measured result.

Additional capability-owned tools include Matrix/MCP media operations, search, browser, connectors, deployment, and task/subagent operations; the available IDs and feature gates are centralized in [`packages/config/src/agent-capabilities.ts`](../packages/config/src/agent-capabilities.ts). Built-in skills include code review, deep research, initialization, document/spreadsheet/presentation/PDF workflows, plugin/skill creation, and media-oriented workflows ([same file](../packages/config/src/agent-capabilities.ts)).

## Guidance system: recipes portable to other harnesses

1. **Put invariants in tool descriptions.** Example: “read immediately before editing” and “all replacements apply atomically” in the edit schema ([`packages/agent-core/src/tools/builtin-defs.ts`](../packages/agent-core/src/tools/builtin-defs.ts)).
2. **Generate project-local guidance.** `mcode init` invokes the Runtime init skill to create/update `AGENTS.md`; the README tells users to specify expected result, allowed changes, and verification ([`packages/tui/src/cli/program.ts`](../packages/tui/src/cli/program.ts), [`README.md`](../README.md)).
3. **Inject scoped reminders instead of one giant static prompt.** The system-reminder model carries project guidance, scratchpad, available skills, bootstrap state, and feature-gated evolution instructions ([`packages/agent-modules/system-reminder/src/types.ts`](../packages/agent-modules/system-reminder/src/types.ts)).
4. **Make extensions contribute prompt, tools, reminders, and lifecycle hooks through one registry.** The registry assembles those independently per turn and isolates concurrent assemblies ([`packages/agent-runtime/src/registry.ts`](../packages/agent-runtime/src/registry.ts)).
5. **Use skills as bounded, discoverable recipes.** The skill registry ranks project/workspace/agent/global/builtin sources, caches scans, watches changes, and caps rendered content ([`packages/agent-modules/skills/src/registry.ts`](../packages/agent-modules/skills/src/registry.ts)).
6. **Give agents durable task state and controlled delegation.** Task query/output/stop are first-class capability IDs, and delegation is a feature gate ([`packages/config/src/agent-capabilities.ts`](../packages/config/src/agent-capabilities.ts)).

## Permissions, sandbox, and trust boundaries

Interactive permission modes are Ask, Auto, and Full access; headless supports `smart`, `full`, and `off`, while Ask is reserved for TUI/ACP because it needs an interactive confirmation surface ([`docs/installation.md`](installation.md), [`packages/tui/src/cli/contract.ts`](../packages/tui/src/cli/contract.ts)). The permission package contains command tokenization/AST checks, path capability checks, write-target analysis, slow-command scanning, and platform-specific deletion handling ([`packages/agent-modules/permission/src`](../packages/agent-modules/permission/src)).

Sandbox policy is separately modeled with filesystem modes `read_only`, `workspace_write`, `delete_guard`, and `full_access`, plus network policy and denied domains ([`packages/config/src/sandbox-config.ts`](../packages/config/src/sandbox-config.ts), [`packages/config/src/sandbox-settings.ts`](../packages/config/src/sandbox-settings.ts)). The current value defaults normalize disabled/full access to no sandbox; that is a product default, not proof that every deployment is unrestricted. The vendored sandbox runtime supplies native helpers, request filtering, credential masking, proxies, and violation monitoring ([`third_party/sandbox-runtime/src`](../third_party/sandbox-runtime/src)).

## Providers, BYOK, and configuration

Managed MiniMax login and custom providers coexist. The provider CLI supports OpenAI Completions, OpenAI Responses, and Anthropic Messages formats, repeated models, model context/output limits, image support, environment-selected API keys, connection testing, and save-and-select semantics ([`packages/tui/src/cli/program.ts`](../packages/tui/src/cli/program.ts), [`README.md`](../README.md)). Provider routing distinguishes managed token-plan, MiniMax API-key, custom-provider, and configured-provider routes ([`packages/config/src/model-availability.ts`](../packages/config/src/model-availability.ts)).

Configuration is profile/data-directory based, with private config-file handling and explicit config-path support for headless runs ([`docs/installation.md`](installation.md), [`packages/config/src/private-config-file.ts`](../packages/config/src/private-config-file.ts), [`packages/tui/src/cli/contract.ts`](../packages/tui/src/cli/contract.ts)). Telemetry channels are separately opt-in and documented as excluding prompts, responses, paths, command text, plugin names, and credentials for usage events ([`docs/telemetry.md`](telemetry.md)).

## Extension points

The extension SPI permits runtime tools, system/user prompt contributors, reminder providers, and named lifecycle hooks. Initialization is owner-scoped and failure transitions the registry permanently to `failed`, preventing partial contribution leakage ([`packages/agent-runtime/src/registry.ts`](../packages/agent-runtime/src/registry.ts)). Plugin hooks, MCP disclosure/invocation, skills, and capability-owned tools are separate packages, allowing product features to be composed without changing the CLI transport ([`packages/agent-modules/plugin-hooks/src`](../packages/agent-modules/plugin-hooks/src), [`packages/agent-tools/src/mcp-disclosure`](../packages/agent-tools/src/mcp-disclosure)).

## Efficiency and quality: evidence versus inference

Evidence supports these architectural advantages: one in-process runtime is reused by three entry points ([`docs/architecture.md`](architecture.md)); tool schemas carry workflow policy ([`packages/agent-core/src/tools/builtin-defs.ts`](../packages/agent-core/src/tools/builtin-defs.ts)); output is bounded and large-file reads are paged ([same file](../packages/agent-core/src/tools/builtin-defs.ts)); extensions assemble prompt/tools/hooks per turn ([`packages/agent-runtime/src/registry.ts`](../packages/agent-runtime/src/registry.ts)); and sessions can resume or continue ([`README.md`](../README.md)).

The plausible mechanism is lower interaction overhead plus fewer invalid mutations: bounded retrieval keeps context focused, parallel-safe read/search tools expose concurrency metadata, exact-edit contracts reduce retries, and reusable per-turn extensions keep guidance close to the operation. The repository does not provide a controlled speed/quality comparison against Claude Code, Codex, or Agy, so superiority over those tools is a hypothesis.

| Comparison claim | Status |
|---|---|
| MiniMax Code has more built-in media/managed-service surface than a minimal coding CLI | Plausible from the listed MCP tools, connectors, mcode-tools, and plugins; competitor parity is unverified ([`README.md`](../README.md), [`packages/config/src/agent-capabilities.ts`](../packages/config/src/agent-capabilities.ts)). |
| One runtime for TUI/headless/ACP reduces behavioral drift | Evidence-backed architecture; comparative drift rates are unmeasured ([`docs/architecture.md`](architecture.md)). |
| Tool-level guidance improves first-pass correctness | Reasonable hypothesis from the detailed schemas; no experiment in this repository proves effect ([`packages/agent-core/src/tools/builtin-defs.ts`](../packages/agent-core/src/tools/builtin-defs.ts)). |
| Local/in-process execution is faster than a daemon or remote orchestration path | Architecture documents the local boundary, but no latency benchmark establishes this comparison ([`docs/architecture.md`](architecture.md)). |
| Provider flexibility improves model quality/cost choice | Configuration supports managed and BYOK routes; outcome depends on selected models and is unverified ([`packages/config/src/model-availability.ts`](../packages/config/src/model-availability.ts)). |

## Copyable recipe for another harness

```text
For every tool, define: purpose, bounded output, concurrency mode, exact input schema,
and the failure-recovery rule. For edit tools, require a fresh read, unique exact
matches, atomic application, and a post-edit reread. Inject project instructions and
available skills only when relevant. Keep TUI, batch, and editor transports thin over
one runtime. Gate dangerous operations through a policy engine and make the policy
visible in the UI and batch flags. Persist sessions, model choice, and task state so
the agent can resume without replaying completed tool calls.
```

## Summary and open questions

The strongest evidence-backed differentiators are compositional runtime reuse, unusually prescriptive tool contracts, bounded context flow, session continuity, broad managed/BYOK tooling, and a typed extension/skill system. The speed and quality advantage over Claude Code, Codex, or Agy remains a hypothesis until matched-task, same-model measurements are published.

Open questions: What are median tool calls, retries, wall-clock latency, and token counts on matched tasks? How much do tool-description rules versus model/provider differences contribute? Which competitor features have equivalent permission/sandbox semantics? How does local execution behave across platforms and under real network/service failures?
