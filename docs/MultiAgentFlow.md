# Multi-Agent Flow

## Scope and evidence

This note documents the repository's local task/subagent surface. Claims marked **Evidence** are direct observations from the cited source. Claims marked **Inference** describe the likely efficiency consequence; they are not benchmark results. The public projection includes desktop task adapters and local-runtime integration, but the concrete adapter implementation is injected through interfaces, so this file does not invent an undocumented scheduler limit.

## Flow at a glance

```text
parent Agent
  -> task(description, prompt, agent_name, background?)
  -> child Session/Turn through LocalTaskAdapter
  -> task receipt (task_id; optionally session_id)
  -> foreground result OR background notification
  -> task_output / task_query / task_stop
  -> task_append(task_id, content) for continuation
```

**Evidence:** `LocalTaskTool` chooses `startBackground` or `runForeground`, emits structured XML-like receipts, and preserves both task and child-session handles ([local-task.ts](../packages/agent-tools/src/desktop/local-task.ts)). The tool contract says the child receives no parent conversation history, while inherited Agent instructions, scoped context, and exposed tools still apply ([builtin-defs.ts](../packages/agent-tools/src/desktop/builtin-defs.ts)).

**Inference:** Fresh context plus an explicit self-contained brief reduces parent-context bloat and makes delegation reproducible. Returning a compact result instead of merging a full transcript is a likely speed/quality advantage, but this repository does not provide a cross-harness benchmark.

## Roles and guidance

The canonical roster is intentionally small:

| Role | Intended assignment | Capability boundary |
| --- | --- | --- |
| `explore` | Read-only, unfamiliar, cross-file, evidence-heavy investigation | No writes, task delegation, computer-use tools, or task-output polling |
| `worker` | Bounded production work with ownership and acceptance criteria | Cannot delegate further through `task`/`task_append` |
| `verifier` | Independent validation of an existing deliverable | Read-only like Explore; keeps validation artifacts in an explicitly temporary location |
| `mavis` | Broad or mixed-scope work | General built-in target |

**Evidence:** Role definitions and their “when to use” text are the single source in [`subagent-roles.ts`](../packages/shared/src/subagent-roles.ts). The native ceiling removes delegation from Worker and removes write, memory, user-interaction, computer-use, and (for Explore) task-control tools from read-only roles ([`canonical-tool-policy.ts`](../packages/agent-tools/src/desktop/canonical-tool-policy.ts)). MCP is similarly reduced to the built-in Matrix web-search entry for Explore/Verifier.

**Evidence:** The `task` description teaches the parent to delegate only concrete bounded work, keep independent writers on disjoint files, use Worker for edits, and require a self-contained prompt containing objective, facts, ownership, constraints, deliverable, acceptance criteria, and response format ([`builtin-defs.ts`](../packages/agent-tools/src/desktop/builtin-defs.ts)).

**Inference:** This is strong guidance because it combines semantic role selection with enforced tool ceilings; a prompt mistake is less likely to become an unauthorized write. The role text also makes delegation composable: Explore gathers evidence, Worker edits, and Verifier audits.

## Spawn, isolation, and lifecycle

**Evidence:** `task` creates a fresh child Agent for one bounded subtask. Foreground execution waits for completion. Background execution returns immediately with a `task_id`; completion automatically resumes the owning conversation ([`builtin-defs.ts`](../packages/agent-tools/src/desktop/builtin-defs.ts), [`local-task.ts`](../packages/agent-tools/src/desktop/local-task.ts)). The child has its own context and the parent remains the owner of interpretation, scope, decisions, and final delivery.

**Evidence:** Background task states exposed to the model are `queued`, `running`, `stopping`, `succeeded`, `failed`, `canceled`, and `lost`. `task_stop` cancels queued work or aborts a running child ([`builtin-defs.ts`](../packages/agent-tools/src/desktop/builtin-defs.ts), [`local-task-control.ts`](../packages/agent-tools/src/desktop/local-task-control.ts)). Task records can expose `session_id`, `parent_task_id`, execution mode, timestamps, and errors, while child session identity is only surfaced for `subagent` tasks ([`local-task-control.ts`](../packages/agent-tools/src/desktop/local-task-control.ts)).

**Evidence:** Context compaction captures subagent state, including status counts and task items, so active delegation is not silently lost when the parent context is compacted ([`automatic-context-compactor.ts`](../packages/local-runtime-v2/src/service/turn-system/compaction/automatic-context-compactor.ts), [`compat.ts`](../packages/local-runtime-v2/src/service/turn-system/compaction/compat.ts)).

**Inference:** The lifecycle is durable enough for long work and context pressure: the parent can resume from a handle, inspect terminal/error state, and recover delegation state after compaction. The exact persistence and restart recovery behavior belongs to the injected runtime/background-task service and should be verified there before claiming process-crash guarantees.

## Messages, continuation, and result handoff

`task` is the initial delegation message. Its foreground response contains status, requested/resolved agent names, child handles, event count, final text, and optional verification data; failures retain task/session handles and error details ([`local-task.ts`](../packages/agent-tools/src/desktop/local-task.ts)). This gives the parent structured evidence rather than an opaque success bit.

`task_append` addresses an existing child by `task_id`. Its result is an admission acknowledgement, not completion:

| Append mode | Meaning |
| --- | --- |
| `activated` | Idle child starts a new child Turn under a new task id |
| `steered` | Follow-up joins the Turn already in flight and produces one merged answer |
| `duplicate` | Exact call was already admitted; nothing is delivered twice |

**Evidence:** These modes, ownership checks, pre-admission abort behavior, and the rule that an admitted child continues even if the parent turn ends are specified in [`builtin-defs.ts`](../packages/agent-tools/src/desktop/builtin-defs.ts) and implemented by [`local-task-append.ts`](../packages/agent-tools/src/desktop/local-task-append.ts).

For background work, `task_output` reads incremental byte offsets and can wait up to 30 seconds. Completion wakes the owner; the tool explicitly advises continuing independent work instead of tight polling. `task_query` lists or fetches task state, and `task_stop` is the cancellation path ([`local-task-control.ts`](../packages/agent-tools/src/desktop/local-task-control.ts)).

**Copyable recipe:**

```text
task:
  description: "Inspect authentication flow"
  agent_name: explore
  prompt: "Read-only. Trace login from CLI entry to storage. Cite exact files and line ranges. Do not edit. Return findings, risks, and open questions in <= 20 bullets."

task:
  description: "Implement the parser fix"
  agent_name: worker
  prompt: "Own only packages/foo/src/parser.ts and its focused test. Apply the accepted finding X. Preserve public APIs. Run the relevant check and return files, checks, and remaining risks."

task:
  description: "Audit the change"
  agent_name: verifier
  prompt: "Read-only independent review of the current diff against acceptance criterion X. Do not edit. Report blocking findings first, then evidence and missing tests."
```

## Parallelism and merge discipline

**Evidence:** The repository's process guidance says “Parallel Read, Sequential Write”; multiple read-only advisors may run concurrently, while writes use one writer per workspace. It explicitly prohibits parallel writers on one workspace and requires tracking and terminating every child ([`docs/AgenticLoop.md`](./AgenticLoop.md)). The `task` tool itself is marked sequential, while query/output controls are marked parallel ([`builtin-defs.ts`](../packages/agent-tools/src/desktop/builtin-defs.ts)).

**Evidence:** The task contract permits background execution for independent work and requires disjoint file ownership for parallel writers ([`builtin-defs.ts`](../packages/agent-tools/src/desktop/builtin-defs.ts)). No numeric maximum for simultaneous subagents is declared in the inspected task tool, role definitions, or public orchestration guidance; adapter/service implementations own that detail.

**Inference:** The speed strategy is selective overlap: parallelize independent discovery or long-running tasks, but serialize conflicting writes and integration. Quality comes from staged specialization and an independent verifier, not from indiscriminate fan-out. The handoff is a result/receipt merge into the parent’s reasoning, not an automatic source merge.

## Why this can outperform a single generalist (hypothesis)

The repository supports four reinforcing mechanisms: (1) fresh bounded contexts, (2) role-specific prompts and tool ceilings, (3) asynchronous execution with automatic wake-up and incremental output, and (4) explicit verification plus compaction-safe task state. These are design mechanisms, not proof that MiniMax Code beats Claude Code, Codex, or Agy on identical models. A valid comparison would need the same model, task set, permissions, concurrency budget, and measured wall time, token use, defect rate, and reviewer acceptance.

## Open questions

- Where is the concrete background-task scheduler and its numeric concurrency/backpressure policy implemented?
- What exactly survives process restart, versus only parent-turn cancellation or context compaction?
- How are child files, commits, or patches integrated when a Worker edits a shared workspace?
- What telemetry measures the claimed speed/quality advantage, and are there controlled comparisons against other harnesses?

## Summary

The public flow is handle-based delegation: a parent assigns a bounded self-contained prompt to a role-constrained fresh child, optionally runs it in the background, then receives structured output or a wake-up and decides how to integrate it. Parallelism is encouraged for independent reads/background work and constrained by ownership for writes; the exact scheduler limit remains outside the inspected interfaces.
