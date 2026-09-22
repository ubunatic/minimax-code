# Agent Guidance Tricks

This note documents mechanisms visible in the repository. “Evidence” names the
implementation; “inference” describes the likely quality or efficiency effect
and should be validated with task-level benchmarks.

## Guidance architecture

| Layer | Evidence | What it contributes |
| --- | --- | --- |
| Stable role contract | `packages/local-runtime-v2/assets/agents/worker/system-prompt.md.hbs` | A bounded assignment, explicit scope, proportional validation, and a predictable final report. |
| Specialized workflows | `packages/local-runtime-v2/assets/agents/workflow/` | Plan entry, goal continuation, recovery, review, and other state-specific instructions are selected as runtime context. |
| Dynamic reminders | `packages/agent-modules/system-reminder/src/service.ts` | Per-turn state is collected, provider blocks are selected, and the result is injected as one `<system-reminder>` envelope. |
| Modular expertise | `packages/agent-modules/skills/src/registry.ts` | Skills are discovered from ranked roots, cached, watched, and resolved with precedence. |
| Control hooks | `packages/agent-core/src/pi-turn-runner/hooks.ts` | Before-LLM hooks can continue, replace request context, append hidden context, respond, abort, or fail closed; tool and after-LLM hooks provide additional control. |
| Tool contracts | `packages/agent-core/src/tools/builtin-defs.ts` | Descriptions encode when to use a tool, safety checks, atomicity, and output constraints alongside schemas. |
| Mid-run steering | `packages/local-runtime-v2/src/service/turn-system/execution/steering/steer-session.service.ts`, `packages/tui/src/tui/controller/run/active-run-flow.ts` | New guidance can enter an active run through an explicit steer path rather than waiting for a new session. |
| Anti-runaway guard | `packages/agent-modules/runaway-guard/src/reminder.ts` | Repeated actions/errors/polling produce one temporary strategy-change reminder. |

Inference: the speed/quality advantage is not one “better prompt”; it is a
feedback loop that keeps stable instructions small, injects only relevant
state, and changes the model’s next decision at the exact boundary where it can
still matter. The repository does not by itself prove superiority over Claude
Code, Codex, or other harnesses on identical models.

## Core mechanisms

### 1. Stable prompts are operational contracts

The worker prompt says to produce one bounded deliverable, inspect conventions,
make the smallest coherent change, validate proportionally, and report result,
paths, validation, assumptions, and blockers
(`packages/local-runtime-v2/assets/agents/worker/system-prompt.md.hbs`). This
reduces repeated user prose and makes completion auditable. The plan-mode entry
prompt also makes entry a protocol decision: planning requests enter plan mode
before investigation (`packages/local-runtime-v2/assets/agents/workflow/plan-mode/agent-entry.md`).

### 2. Reminders are stateful, selective, and scoped

`SystemReminderService.buildReminder` increments a turn counter, collects
current state, updates todo state, computes evolution reminders, filters
providers by an optional allowlist, and emits one envelope
(`packages/agent-modules/system-reminder/src/service.ts`). Providers include
completion and delivery reminders, such as refusing to claim completion while
active todos remain (`packages/agent-modules/system-reminder/src/blocks.ts`).
Model-specific reminders can append a narrowly targeted skill instruction
(`packages/agent-modules/system-reminder/src/mcode-tools-master-reminder.ts`).

This is an important efficiency trick: do not stuff every rule into every
request. Keep a cheap stable base and inject only facts whose state changed or
whose trigger fired.

### 3. Hooks enforce guidance at execution boundaries

`PiBeforeLlmCallHookDecision` supports request replacement without durable
history replacement, durable hidden-message append, trusted response, abort,
and fail-closed decisions (`packages/agent-core/src/pi-turn-runner/hooks.ts`).
The same hook family observes the final prepared request, while tool hooks can
control tool execution. This lets a harness enforce context compaction,
permissions, budgets, and policy after the model has proposed an action but
before the provider or tool boundary commits it.

### 4. Skills are portable, ranked procedures

`SkillRegistry` scans multiple roots, caches unchanged files, resolves winners
by source rank, records diagnostics, and watches directories for refreshes
(`packages/agent-modules/skills/src/registry.ts`). This makes a skill a
versioned procedure that can be selected on demand, rather than a permanent
prompt tax. The runtime also exposes skill instructions through the session
surface, as shown by the TUI skill/ACP tests (`packages/tui/test/unit/acp-agent.test.ts`).

### 5. Todos and plans externalize state

The `todowrite` tool description requires structured tasks only for complex
work, exactly one `in_progress` item, and explicit status transitions
(`packages/agent-core/src/tools/builtin-defs.ts`). The completion reminder then
turns unfinished state into a delivery gate (`packages/agent-modules/system-reminder/src/blocks.ts`).
Goal continuation similarly requires inspecting evidence before continuing and
auditing every explicit requirement before completion
(`packages/local-runtime-v2/assets/agents/workflow/goal/continuation.md`).

### 6. Tool descriptions carry behavioral policy

The edit tool description requires a fresh read, exact unique matches, atomic
multi-edit calls, and a reread after success; write is reserved for new files
or complete rewrites (`packages/agent-core/src/tools/builtin-defs.ts`). These
constraints are delivered at tool-selection time, when they are most useful,
and are backed by schemas and runtime preparation rather than prompt text alone.

### 7. Steering is a bounded feedback channel

The TUI accepts steer input for an active run and forwards it through the
runtime (`packages/tui/src/tui/controller/run/active-run-flow.ts`). Runtime
steering has explicit session and turn services
(`packages/local-runtime-v2/src/service/turn-system/execution/steering/steer-session.service.ts`),
while the turn runner can feed messages into the active agent
(`packages/agent-core/src/pi-turn-runner/events.ts`). This supports correction
without discarding the transcript or restarting expensive setup.

### 8. Guardrails change strategy, not just stop execution

The runaway guard prioritizes unchanged progress, repeated errors, exact action
repeats, and polling; it reserves one reminder before attempting a steer, and
explicitly says to inspect state or change one controlled variable
(`packages/agent-modules/runaway-guard/src/reminder.ts`). This is a compact
anti-loop intervention rather than an unbounded stream of warnings.

## Copyable recipes for another harness

1. **Bound the deliverable.**
   - Problem: agents broaden work and produce inconsistent handoffs.
   - Trick: put role, scope, proportional validation, and a fixed report schema
     in the stable system prompt.
   - Where implemented: `packages/local-runtime-v2/assets/agents/worker/system-prompt.md.hbs`.
   - Sketch: `Work only in scope. Validate proportionally. Return Result / Changes / Validation / Assumptions / Blockers.`
   - Use when: workers execute tickets or delegated subtasks.

2. **Inject only triggered state.**
   - Problem: a giant system prompt wastes context and dilutes important rules.
   - Trick: collect session state each turn, run an allowlisted provider chain,
     and wrap non-empty blocks in one reminder envelope.
   - Where implemented: `packages/agent-modules/system-reminder/src/service.ts`.
   - Sketch: `blocks = providers.filter(p => allowlist.has(p.name)).map(p => p(state)); send(<system-reminder>{blocks}</system-reminder>)`.
   - Use when: reminders depend on todos, model, memory, or session phase.

3. **Enforce before commitment.**
   - Problem: prompt-only rules can be ignored after a tool or context change.
   - Trick: add a pre-provider hook with `continue | replace | append | respond | abort` and fail closed on hook errors.
   - Where implemented: `packages/agent-core/src/pi-turn-runner/hooks.ts` and `packages/agent-core/src/pi-turn-runner/llm.ts`.
   - Sketch: `decision = beforeLLM(request); if (decision.type === 'abort') stop(); if (replace) request = decision.messages;`.
   - Use when: enforcing permissions, context limits, budgets, or canonical history.

4. **Make procedures loadable.**
   - Problem: domain guidance is too large to include in every run.
   - Trick: discover skills from ranked roots, cache unchanged files, and watch for updates.
   - Where implemented: `packages/agent-modules/skills/src/registry.ts`.
   - Sketch: `skill = registry.resolve(name); inject(skill.content only when selected)`.
   - Use when: teams need reusable, domain-specific workflows.

5. **Turn completion into a state check.**
   - Problem: agents announce success with pending work.
   - Trick: maintain structured todos and inject a final reminder that blocks the claim while active items remain.
   - Where implemented: `packages/agent-core/src/tools/builtin-defs.ts`, `packages/agent-modules/system-reminder/src/blocks.ts`.
   - Sketch: `if todos.some(t => t.status in ['pending','in_progress']) remind('finish or cancel before delivery')`.
   - Use when: tasks have multiple files, gates, or acceptance criteria.

6. **Put safety advice in tool schemas.**
   - Problem: the model selects tools with incomplete operational knowledge.
   - Trick: describe use cases, preconditions, atomicity, and output formatting next to the JSON schema.
   - Where implemented: `packages/agent-core/src/tools/builtin-defs.ts`.
   - Sketch: `edit.description = 'Read immediately before; exact unique oldText; atomic edits; reread after.'`.
   - Use when: misuse is predictable and can be prevented before execution.

7. **Steer active runs instead of restarting.**
   - Problem: user corrections arrive while a long run is executing.
   - Trick: expose a turn-scoped steer endpoint that injects a new message through the active runner.
   - Where implemented: `packages/tui/src/tui/controller/run/active-run-flow.ts`, `packages/local-runtime-v2/src/service/turn-system/execution/steering/steer-session.service.ts`.
   - Sketch: `POST /session/steer {text}; admit at a safe boundary; preserve durable history.`
   - Use when: the user needs to redirect scope, cancel a route, or add missing context.

8. **Nudge loops once, then require a strategy change.**
   - Problem: repeated polling or identical tool calls burn latency and tokens.
   - Trick: classify repeated signals, choose the highest-priority one, reserve one temporary reminder, and tell the agent exactly what to inspect or change.
   - Where implemented: `packages/agent-modules/runaway-guard/src/reminder.ts`.
   - Sketch: `if repeated(action) && !reminded: steer('Do not repeat unchanged; inspect state or change one variable.')`.
   - Use when: external tasks, flaky commands, or polling can create loops.

## Summary and open questions

The strongest portable pattern is layered guidance: concise stable contracts,
specialized skills, state-triggered reminders, execution-boundary hooks,
structured progress state, and a single controlled steering channel. This
should improve both speed and quality by reducing prompt overhead, preventing
known failure modes early, and avoiding costly restarts (inference).

Open questions: the repository does not provide a controlled same-model
benchmark against Claude Code, Codex, or Agy; the relative token/latency cost
of reminder providers is not established here; and the best trigger thresholds
for runaway detection require production or replay evaluation.
