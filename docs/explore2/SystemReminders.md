# System Reminders: The In-Band Guidance and Prompt-Injection Defense Layer

`packages/agent-modules/system-reminder/` is a self-contained, IO-free package
that assembles a single `<system-reminder>` block injected ahead of (almost)
every message sent to an agent. Two instances of the broader pattern were
already documented in `docs/explore/AgentDelegation.md` — the
`BTW_SIDE_BOUNDARY` reminder (`conversation-fork-service.ts:31-48`) and the
`trust="untrusted_data"` wrapping used for verifier subagent prompts
(`goal/src/verification/subagent.ts`) — without ever looking at the framework
that produces the ordinary per-turn reminders. This doc covers that framework:
its provider taxonomy, injection triggers, interaction with context
compaction, and how the same mechanism functions as a prompt-injection
defense layer, cross-referencing the two known instances rather than
re-deriving them.

## 1. Architecture: chain-of-responsibility over providers

Three pieces (`packages/agent-modules/system-reminder/src/`):

- **`registry.ts`** — `SystemReminderRegistry`. Holds an ordered list of
  `NamedProvider { name, fn, critical }`. `.append()` registers a provider for
  all `AgentFrameworkType`s (`opencode` / `pi-agent` / `codex`,
  `types.ts:50-55`); `.appendFor(frameworkType, ...)` registers one only for a
  specific framework. `.resolve(frameworkType)` returns defaults followed by
  framework-specific providers, in registration order — that order is also the
  block order inside the final `<system-reminder>` text.
- **`providers.ts`** — the individual `ReminderProviderFn`s (~25 of them) plus
  `createDefaultRegistry()`, the factory that registers all built-ins in a
  fixed order (`providers.ts:823-854`).
- **`service.ts`** — `SystemReminderService.buildReminder(session, msg)`, the
  entry point called once per outbound message. It collects data via an
  injected `DataCollector` (host-owned IO), runs every resolved provider in
  order, concatenates non-empty results with blank lines, and wraps the whole
  thing: `` `<system-reminder>\n${blocks.join('\n\n')}\n</system-reminder>` ``
  (`service.ts:210-213`).

A `ReminderProviderFn` is `(input: SystemReminderInput) => string | undefined
| Promise<...>` (`types.ts:105-107`) — pure function of the collected input,
returns `undefined` to skip. This is why the package can stay IO-free: all the
filesystem/sqlite work happens once, up front, in the host's `DataCollector`,
and every provider is a deterministic renderer over that snapshot.

## 2. Reminder type taxonomy

Grouped by what triggers them (provider name → registration in
`providers.ts:823-854`, builder in `blocks.ts`):

**Identity / session framing (every turn, always-on)**
- `agentContextProvider` → `<agent-context>` — full block on turn 1
  (`buildAgentContextBlock`), a slimmer variant on every subsequent turn
  (`buildSlimAgentContextBlock`, `providers.ts:193-199`). Carries session id,
  role (root/branch), scratchpad path, workspace dir, parent-session
  result-delivery instructions, etc.
- `peersUpdateProvider` → `<peers_update>`, only when the peer list changed
  since the last injection (`peersChanged` flag) and team mode is on.
- `inboundMetaProvider` → renders trusted platform/sender/chat metadata
  extracted from an `<inbound-context>` tag (`types.ts:293`, "Layer 1 —
  trusted metadata" per the comment) as a `# Inbound Message Context` block.

**Cold-start / onboarding (cool-down gated, audience-restricted)**
- `personaMissingProvider` → `<persona_missing>`, `bootstrapProvider` →
  `<bootstrap_check>`, `worktreeReminderProvider` → `<worktree-reminder>`.
  All three share `isColdStartAudience()` (`providers.ts:184-188`: fires only
  for orchestrator agents or manually-created agents — `auto`-spawned workers
  stay silent) and `tryEnterColdStart()`, a per-`(sessionId, kind)` in-memory
  cool-down (`COLD_START_COOLDOWN_MS = 6h`, overridable per-provider via
  `ReminderPolicyEntry.thresholds.cooldown_ms`).

**Memory / skill-evolution routing (turn-scheduled, exponential backoff)**
- `memorySkillReminderProvider` → `<memory-skill-reminder>` — file size, a
  skill-signal/proposal trigger checklist, and a three-question
  user/agent/project memory-layer attribution test. Schedule: turn 1, then
  10 → 20 → 40 → 40... turns apart (`MEMORY_SKILL_REMINDER_DEFAULTS`,
  `providers.ts:322`).
- `evolutionReminderProvider` wraps `getEvolutionReminder()` from
  `evolution.ts` — a similar exponential-backoff memory-write nudge (10 → 20
  → 40, `EVOLUTION_CONFIG` in `types.ts:426-444`) that only fires if the
  agent's memory directory snapshot hasn't changed since the last check
  (`current.maxMtime > initial.maxMtime`, `evolution.ts:109-110`).
- `memoryTopicsProvider`, `cliSunsetMemoryNoticeProvider`,
  `skillEvolutionChannelsProvider` — first-message-only, teach available
  memory topic files, warn about stale CLI references in memory
  (`buildCliSunsetMemoryNoticeBlock`), and teach the skill-signal/proposal
  channels respectively.

**Mid-session state-change notices (event-driven, one-shot)**
- `userMemoryUpdateProvider`, `agentMemoryUpdateProvider`,
  `memorySummaryUpdateProvider`, `dailyMemoryUpdateProvider`,
  `identityUpdateProvider`, `configUpdateProvider` — all fire only when the
  corresponding host-collected field is non-empty this turn (memory files or
  config changed since session start), each delivering the *full updated
  content* inline so the agent doesn't need a separate read.
- `pendingSessionRemindersProvider` — delivers host-queued one-shot bodies
  (e.g. session-rotate handoff bootstrap) that are already fully formatted
  text; consumed/cleared on collection so each fires exactly once
  (`types.ts:303-309`).
- `branchNotificationProvider` (factory, takes a host `DateFormatter`) →
  `<branch-finish-alert>` when child branch sessions finished without
  reporting back.

**Turn-scheduled operational nudges (exponential backoff, compaction-aware)**
- `asyncAuditProvider` → `<async-audit>` (5 → 10 → 20 turns), silenced when
  `cronEnabled === false` so hosts without Cron aren't told to call a command
  that fails closed.
- `mediaOutputReminderProvider` → `<media-output-reminder>` (10 → 20 → 40
  turns), orchestrator/manual-agent audience only.
- `taskCompletionReminderProvider` (critical) → `<task-completion-reminder>`
  when active `TodoWrite` items remain unreconciled, on a turn-interval gate
  (`shouldInjectTodoCompletionReminder`, default every 5 turns).
- `activePlanReminderProvider` → `<active-plan-reminder>` for orchestrators
  owning active team plans, nudging new work into the existing plan instead of
  a fresh session.
- `secretEnvReminderProvider` → `<secret-env>`, listing encrypted secret env
  var **names only** (never values — the comment at `types.ts:363` is
  explicit that values are masked by `SecretMasker` on outbound text), on turn
  1 and whenever the name set's djb2 hash changes.
- `createBoardNudgeProvider(nudgeRegistry)` (closure-based, appended by the
  host container after construction) → `<engine-nudge>`, a one-shot poke to
  update a plan-board file, cleared after injection.

**Model/framework-specific tail providers**
- `withMcodeToolsMasterReminder()` (`mcode-tools-master-reminder.ts`) is not a
  registry provider but a post-hoc wrapper applied only for models whose ID
  starts with `MiniMax-M2.7`: it unwraps any existing `<system-reminder>`
  text, appends an `<mcode-tools-master-reminder>` block, and re-wraps. This
  is the one place in the package where reminder injection is keyed off model
  identity rather than session/turn state.
- `plugin-reference.ts` builds `<selected-plugin>` reminders when a user
  message contains an explicit `@Plugin` reference (whitespace-delimited
  detection, `detectPluginReferencesForMessages`), listing that plugin's
  connected app/MCP tools so the model doesn't have to guess capability names.

## 3. `critical` providers and the SR-disable escape hatch

Every provider is registered as either "normal" (`.append`) or
`.appendCritical`. `service.ts:100-125` can put a session into
`criticalOnly` mode when the resolved model matches an entry in
`disableSrModels` (config: `contextManagement.disableSystemReminderModels`) —
in that mode only critical providers run. Currently critical:
`pendingSessionRemindersProvider`, `taskCompletionReminderProvider`,
`mediaOutputReminderProvider`, `asyncAuditProvider`
(`providers.ts:835-836,850-851`). This is a per-model kill switch for the
bulk of the reminder volume while keeping the few reminders judged essential
(pending one-shot handoffs, todo completion, deliverable format, async-audit
safety) always on.

Separately, `SystemReminderInput.systemReminders` (`types.ts:395`) is a
cloud-runtime allowlist sourced from `AgentConfig.system_reminders`: when
present (even empty), only providers whose stripped name
(`stripProviderSuffix`, drops the `Provider` suffix) appears in the allowlist
run — and per-entry `frequency.intervals` / `thresholds.cooldown_ms` /
`thresholds.todo_reminder_interval_turns` override each provider's
hard-coded backoff schedule (read via `findReminderPolicy`). Critical
providers still go through this allowlist gate (`service.ts:166-169` — not
bypassed here, unlike the model-based critical-only mode above).

## 4. Interaction with context compaction

Reference: `docs/explore2/TokenReduction.md` §1 for the compaction mechanism
itself (`context-manager`'s `ContextManager.checkpoint`, token-triggered,
structurally-safe cut points).

System reminders are **not** part of the transcript that gets compacted —
they're generated fresh on every `buildReminder()` call from live
host-collected state (`SystemReminderInput`), not stored as persisted
message history. So there's nothing to "survive" a cut: each reminder is
regenerated for the current turn regardless of what the compactor did to
prior turns.

What *does* interact with compaction is the **turn-count-based scheduling**
used by every exponential-backoff provider (`memorySkillReminderProvider`,
`asyncAuditProvider`, `mediaOutputReminderProvider`, and
`getEvolutionReminder` in `evolution.ts`). Each keeps in-process state keyed
by `sessionId`: `{ lastInjectedAt: turn, backoffLevel }`. Compaction doesn't
touch `turnCounts` (`service.ts:31`, incremented independently per message,
never reset by compaction) — but the pattern all four schedulers implement is
explicit: *"Compaction detected: turnCount dropped below last injection point
→ reset"* (e.g. `providers.ts:614-618`, `evolution.ts:92-101`). The comment
in `evolution.ts:92-93` clarifies the actual mechanism: it's not that
compaction resets `turnCount` directly, but that these providers treat *any*
observed drop in `turnCount` relative to their own last-fired bookmark as a
signal that compaction (or a session restart) happened, and re-arm from the
initial interval rather than the backed-off one. In practice this means a
freshly-compacted session gets the *full* cold-start/backoff reminder set
again on its next few turns — the framework deliberately re-teaches state
that the compactor just cut out of the model's visible history, rather than
assuming the model still remembers it.

`agentContextProvider` reinforces this from the other direction: it emits the
**full** `<agent-context>` block only on turn 1 and the **slim** variant on
every later turn (`providers.ts:193-199`), on the assumption that turn-1
context is still in the live window. If compaction ever cut turn 1's full
context out of history, nothing here re-detects that and re-emits the full
block — only the four explicit `turnCount`-drop-aware providers above have
that resync logic. This is a real, narrow gap: a very long, heavily-compacted
session could lose the full `<agent-context>` fields (scratchpad path,
`taskResultDelivery` mode, etc.) permanently, relying only on the slim block
thereafter.

## 5. The prompt-injection defense framing

The framework functions as a defense layer in three distinct ways:

**(a) It is the delivery channel for hard behavioral boundaries around
untrusted content**, not just informational nudges. The two instances already
documented in `docs/explore/AgentDelegation.md` are the clearest examples:

- `BTW_SIDE_BOUNDARY` (`conversation-fork-service.ts:31-48`) is *itself*
  wrapped in a literal `<system-reminder>...</system-reminder>` string
  (constructed by hand, not through this package's registry — forks are a
  different code path) that explicitly tells the model: content before this
  marker is reference-only history, not active instructions, and sub-agent
  use is off-limits in the forked side-conversation. It borrows the same tag
  name and framing convention this package uses, so the model treats it with
  the same in-band authority as a framework-generated reminder, even though
  it originates from a completely separate service.
- `trust="untrusted_data"` (`goal/src/verification/subagent.ts:242-264`) is a
  sibling convention: XML-tag attributes marking verifier-subagent inputs
  (`evidence_brief_json`, `worker_transcript_json`, `goal_objective`,
  `goal_resources`) as data the model must evaluate, not follow — mirrored in
  `goal/src/verification/evaluator-adapter.ts:291-309` where objects passed to
  the evaluator carry `{ trust: 'untrusted_data', value: ... }` as structured
  metadata rather than an XML attribute. Neither of these is produced by
  `system-reminder`'s registry/provider machinery, but both target the same
  attacker model: content that arrived from outside the current trusted
  instruction chain (a prior conversation turn, a subagent's own transcript,
  external evidence) must be legible to the model as data-with-provenance,
  not as new instructions.

**(b) The registry pattern itself limits injection surface indirectly.**
Every reminder the *framework* proper renders is built from structured,
typed `SystemReminderInput` fields collected by a host `DataCollector` —
never raw string concatenation of arbitrary tool output or user text into an
instruction-bearing block. Where free-form content does flow into a block
(`teamMemoryProvider`'s `teamMemoryIndex`, `relevantMemoryProvider`'s
`relevantMemory`, `inboundMetaProvider`'s `inboundMeta`), it is always
memory the *agent itself* wrote in a prior session or trusted inbound
platform metadata — not attacker-controlled tool output being replayed as an
instruction. This is a narrower trust boundary than `trust="untrusted_data"`,
by construction rather than by explicit tagging: the provider functions
simply don't have a code path that takes untrusted external strings and
renders them without provenance.

**(c) Reminders are structurally separated from tool results and user
messages** — always their own `<system-reminder>`-wrapped block, assembled
once per `buildReminder()` call and prepended to the outbound message rather
than interleaved with tool output. Combined with per-turn regeneration
(§4), this means an attacker who manages to inject text that merely *looks
like* a system-reminder block into a tool result or fetched web page cannot
piggyback on the trust the model extends to blocks the framework itself
emits — the real ones arrive through a distinct, service-controlled channel
each turn, not as something the model has to distinguish by content alone.
That said, nothing in this package enforces that distinction on the model's
behalf (no cryptographic signing or channel-level tagging visible in the
source) — the separation is structural/positional (own paragraph, own tags,
own place in the message) and behavioral convention (the tag name is used
consistently and only by trusted code paths), which is the same class of
defense `BTW_SIDE_BOUNDARY` relies on for fork boundaries.

## 6. Differentiator framing

Most agent CLIs surface static system-prompt text and, at best, a couple of
ad-hoc "reminder" strings hard-coded into specific tool call paths. This
codebase instead has a dedicated, host-decoupled package with:
- a declarative registry (25+ independently-testable, independently-gated
  providers) instead of one monolithic prompt-transform function (the
  monolithic `buildSystemReminders()` predecessor is explicitly noted as
  retired in `blocks.ts:746-749`),
- per-provider audience gating, cooldowns, and exponential backoff that all
  self-heal after compaction,
- a live cloud-runtime allowlist/policy-override wire format
  (`ReminderPolicyEntry`) so operators can retune schedules without a
  redeploy,
- and a consistent `<system-reminder>` / `trust="untrusted_data"` vocabulary
  reused across otherwise-unrelated subsystems (fork boundaries, verifier
  subagents, the reminder framework itself) for marking non-instruction
  content.

That combination — scheduled, self-resetting, audience-aware in-band
guidance plus a shared untrusted-content marking convention — is the
candidate differentiator: it is infrastructure, not a handful of strings.

## Files referenced

- `packages/agent-modules/system-reminder/src/types.ts`
- `packages/agent-modules/system-reminder/src/registry.ts`
- `packages/agent-modules/system-reminder/src/service.ts`
- `packages/agent-modules/system-reminder/src/providers.ts`
- `packages/agent-modules/system-reminder/src/blocks.ts`
- `packages/agent-modules/system-reminder/src/evolution.ts`
- `packages/agent-modules/system-reminder/src/dependencies.ts`
- `packages/agent-modules/system-reminder/src/mcode-tools-master-reminder.ts`
- `packages/agent-modules/system-reminder/src/plugin-reference.ts`
- `packages/agent-modules/system-reminder/src/index.ts`
- `goal/src/verification/subagent.ts` (cross-referenced, not re-derived)
- `goal/src/verification/evaluator-adapter.ts` (cross-referenced, not re-derived)
- `conversation-fork-service.ts:31-48` (cross-referenced, not re-derived)
- `docs/explore/AgentDelegation.md` (prior findings this doc builds on)
- `docs/explore2/TokenReduction.md` (compaction mechanism referenced in §4)
