# Token Reduction: How Agents and Subagents Keep Context Small

Found via `code-index -s "token"/"compaction"/"context window"`, which point
at one primary package: `packages/agent-modules/context-manager`. The
subagent-specific bounding logic lives in `packages/agent-modules/goal` and
`packages/agent-tools/src/desktop`, found earlier while writing
`docs/explore/AgentDelegation.md`.

`code-index` gaps hit along the way: `-s "budget"`, `-s "summariz"`,
`-s "reminder"`, `-s "lazy"`, `-s "defer"` all returned no matches even
though all five concepts have real implementations — the indexed summary
text just doesn't happen to contain those words. Directory- and file-level
hits only fire on vocabulary actually present in the stored summaries.

## 1. Automatic context compaction (`context-manager`)

`packages/agent-modules/context-manager/src/manager.ts` is a `beforeLlmCall`
hook (`ContextManager.checkpoint`) that runs before every LLM call and can
replace the message history with a shorter one.

**Trigger.** `evaluateTrigger` counts context tokens (via a real
`ContextTokenCounter` when available — important, because tool declarations
alone can be tens of thousands of tokens in MCP-heavy sessions and must not
be omitted from the count, per the comment at `manager.ts:189-192` — falling
back to a local estimator otherwise) and compares against `triggerAt`, computed
by `computeCompactionTriggerAt` (`settings.ts`):
- MiniMax-M3 in 512K/1M context mode: fixed 90% of the context window.
- Everything else: `contextWindow - max(reserveTokens, perTurnMaxTokens + safetyMarginTokens)`,
  i.e. always leaves room for the model's own next output plus a safety margin.

Defaults (`DEFAULT_CONTEXT_MANAGER_SETTINGS`): `reserveTokens: 16_384`,
`keepRecentTokens: 20_000`, `minMessagesToCompact: 4`,
`safetyMarginTokens: 2_048`.

**Plan selection.** `selectPlan` never cuts blindly at a token count — it
walks backward from the end of the transcript accumulating token estimates
until `keepRecentTokens` is reached, then calls `safeCutForIndex` to find a
*structurally safe* cut point: never inside a tool-call/tool-result group
(`findToolGroupStartForIndex`/`findToolGroupEnd`), and never leaving an
orphaned `toolResult` without its originating assistant `toolCall`
(`findAssistantForToolResult`). If no safe cut point exists, compaction is
skipped rather than risk a corrupt transcript.

**Compaction.** Everything before the cut point is summarized
(`generateSummary`, via an injected `ContextSummaryGenerator` or the
`pi-agent-core` default) and replaced with one `compactionSummary` message,
chained onto any previous summary (`findLatestCompactionSummaryIndex`, so
summaries compose rather than re-summarizing already-summarized history).
Everything after the cut point is kept verbatim. The result: bounded context
growth independent of how long a session runs, with the actual mechanics of
"never cut mid-tool-call" doing the safety work.

## 2. Dynamic per-call token budgets (`provider-budget.ts`)

Separately from compaction, `resolveDynamicMaxTokens` and
`resolveCompactionTokenBudget` (same package,
`packages/agent-modules/context-manager/src/provider-budget.ts`) shrink the
*requested max output tokens* for a call based on how much context is
already used, rather than always requesting a fixed ceiling:

```
remaining = contextWindow - estimatedContextTokens - safetyMarginTokens
maxTokens = min(configuredMaxTokens, max(outputFloor, remaining))
```

`resolveCompactionTokenBudget` additionally computes a `providerInputLimit`
(capped at 95% of the context window) and a proactive `automaticTriggerAt`
that reserves `min(2 * reserveTokens, contextWindow / 4)` — so local-runtime-v2
sessions get an early, provider-aware compaction trigger distinct from the
model-declared `contextWindow` alone.

## 3. Subagents get a smaller tool surface, not just a smaller prompt

From `packages/agent-tools/src/desktop/canonical-tool-policy.ts`
(detailed in `docs/explore/AgentDelegation.md`): canonical subagent roles
(`worker`, `explore`, `verifier`) get a filtered tool list —
`filterCanonicalNativeToolCeiling` strips write/mutate/delegate tools for
`explore`/`verifier`, and `task`/`task_append` for `worker`. Fewer tool
definitions sent to the model on every call is a direct, compounding token
saving in exactly the sessions (delegated subagents) that run the most calls
— tool schemas are resent on every turn, so trimming the tool list has an
effect that multiplies by turn count.

## 4. Verifier subagents get a hard-bounded evidence payload, not a raw transcript

`packages/agent-modules/goal/src/verification/evidence-brief.ts` builds the
evidence a verifier subagent reads, under multiple hard character caps:

| Constant | Value | Purpose |
|---|---|---|
| `MAX_VERIFICATION_BRIEF_CHARS` | 14,000 | Total structured brief (objective, claim, file/command changes, recent tail) |
| `MAX_SUBAGENT_VERIFICATION_PROMPT_CHARS` | 16,000 | Whole rendered subagent prompt — checked in `subagent.ts`, throws `input_too_large` if exceeded |
| `MAX_EVALUATOR_TAIL_CHARS` / `_MESSAGES` | 32,000 / 20 | Bounded raw-transcript fallback tail |
| `MAX_TRANSCRIPT_FALLBACK_CHARS` / `_MESSAGES` | 512,000 / 200 | Full-transcript fallback mode ceiling |

`fitBrief` enforces the 14K cap by *progressively degrading* the brief rather
than truncating blindly: drop oldest recent-tail entries first, then drop
file/command change entries, then drop `baselineRef`, then truncate `claim`,
then truncate `objective.text`, only throwing if it still doesn't fit. This
means the verifier's prompt cost is a small, predictable constant regardless
of how large or chatty the worker's actual session was — a worker with a
50k-line transcript and a worker with a 50-line transcript cost the verifier
roughly the same.

`serializeBoundedMessages` (used for both the evaluator tail and the full
fallback transcript) applies the same pattern generically: shrink the message
window until it fits, and if a single message still overflows, replace the
whole payload with one `truncatedMessagePreview` string rather than emit
something over budget.

## 5. Background task output is read incrementally, not replayed

From `packages/agent-tools/src/desktop/local-task-control.ts` (see
`docs/explore/AgentDelegation.md` for the full task lifecycle): `task_output`
reads by `offset` cursor instead of re-fetching a subagent's full transcript
on every check. The tool actively discourages token-wasteful polling:
`pollingHint` detects when a caller re-reads the same unchanged
`(status, nextOffset)` and injects `<task_output_hint>...continue independent
work and wait for the completion notification when available.</task_output_hint>`
into the tool result — a self-correcting nudge embedded in the protocol
itself, not just a doc comment, so the calling agent (parent or another
subagent) doesn't burn turns/tokens tight-polling a running child.

## 6. Everything converges on one idea

Token reduction here isn't one mechanism — it's the same discipline applied
at every layer a session touches:

| Layer | Mechanism | File |
|---|---|---|
| Whole transcript | Auto-compaction at a computed trigger, safe-cut-aware | `context-manager/src/manager.ts` |
| Per-call output | Dynamic max-tokens shrink as context fills | `context-manager/src/provider-budget.ts` |
| Tool schemas sent per turn | Role-based tool-list ceiling for subagents | `agent-tools/src/desktop/canonical-tool-policy.ts` |
| Verifier subagent input | Hard-capped, progressively-degraded evidence brief | `goal/src/verification/evidence-brief.ts` |
| Background/subagent polling | Cursor-based reads + anti-repolling hint | `agent-tools/src/desktop/local-task-control.ts` |

None of these are "compress the model's actual reasoning" tricks — they're
all about bounding *what gets sent on the wire on every call*: fewer messages,
fewer tool defs, a fixed-size evidence payload instead of a full transcript,
and fewer redundant polling round-trips. The subagent-specific pieces (tool
ceiling, evidence brief) exist because subagents are exactly where those
costs would otherwise multiply fastest — many short-lived, high-turn-count
children each re-paying the same per-call overhead.
