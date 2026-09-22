# Goal Verification: The Closed Loop Beyond the Verifier Subagent

Scope: the `Goal` system's broader "did the agent actually finish" machinery —
lifecycle, verification backends, breaker/streak logic, and how it all
surfaces to the user. The verifier-subagent mechanism itself
(`packages/agent-modules/goal/src/verification/subagent.ts`,
`createSubagentVerifierAdapter`, its `pass/fail/partial -> met/not_met/inconclusive`
mapping) is already covered in `docs/explore/AgentDelegation.md` § "The Goal
system's verifier subagent" — this doc cross-references it rather than
repeating it.

Package layout: `packages/agent-modules/goal/src/` is host-agnostic (types,
prompts, tool impls, the `VerifierPort` seam). The actual persistence,
breaker mutation, and settlement logic that makes verification a *closed
loop* lives in `packages/local-runtime/src/thread-goal/` — the goal package
never owns a database or Goal write authority itself.

## Lifecycle and status model

`packages/agent-modules/goal/src/types.ts:40-49` defines six statuses:
`active`, `paused`, `blocked`, `complete`, `budget_limited`, `usage_limited`.
Three owners can move a Goal between them (types.ts:6-31):

- **The model**, via `update_goal`, may only *propose* `complete` or
  `blocked` — the host converts the proposal into durable state.
- **The user**, via the `/goal/*` REST surface, owns every transition except
  re-entry into `active` from a status whose budget/work is already spent
  (`complete`, `budget_limited` refuse resume with `409`).
- **The system**, via bound token accounting, auto-transitions `active` →
  `budget_limited` when `tokensUsed + delta >= tokenBudget`.

`TERMINAL_STATUSES` (types.ts:228-233) = `complete`, `blocked`,
`budget_limited`, `usage_limited` — all stop the auto-continuation loop
(`CONTINUATION_STATUSES` only contains `active`, types.ts:177). `complete` is
final; `blocked`/`usage_limited` are user-resumable; `budget_limited(token)`
is resumable only through an explicit token-budget edit.

`ThreadGoalStatusReason` (types.ts:85-113) is the closed catalog of *why* —
e.g. `complete(worker_proposal)` vs `complete(verifier_met)`,
`paused(no_progress)`, `blocked(verifier_impossible)`,
`budget_limited(active_time)`. This is what the UI and logs actually render,
not the six-way status alone.

## Verification backends

`ThreadGoalVerification` (types.ts:51) = `'none' | 'evaluator' | 'subagent'`.
The mode is not a stored, user-picked Goal property — it is recomputed on
every settlement from the worker's model route
(`packages/agent-modules/goal/src/verification/verification-policy.ts:18-25`,
`verificationModeForRoute`):

| Worker route | Verification |
|---|---|
| `managed_token_plan`, `minimax_api_key` (routes MiniMax bills itself) | `subagent` — a read-only verifier child (see AgentDelegation.md) |
| `custom_provider`, `configured_provider` (BYOK), or unresolved route | `none` — the worker's own completion proposal settles the Goal |

A second backend, **`evaluator`**
(`packages/agent-modules/goal/src/verification/evaluator-adapter.ts`,
`createEvaluatorVerifierAdapter`), exists as an alternate `VerifierPort`
implementation but is not reachable through `verificationModeForRoute` above
— it is a fresh, tool-less, thinking-off small-fast model call (not a
subagent turn) that returns one of the same four verdicts
(`met`/`not_met`/`impossible`/`inconclusive`) via strict JSON-schema output
(`VERDICT_SCHEMA`, evaluator-adapter.ts:73-104), with `sameRoute` (line 273)
requiring the evaluator to share the worker's route/provider before it will
dispatch. It shares the `VerifierPort` interface
(`verifier-port.ts:138-140`) with `subagent`, so it is a structurally
available, swappable backend even though current routing does not select it.

`none` means no independent check at all: the worker's own `update_goal`
proposal is durably accepted as-is (`LastWorkerProposalV1`, types.ts:130-139).

## The `notMetStreak` breaker: real logic

`subagent.ts:183`'s doc comment references "the `notMetStreak` breaker...on
five independent verifications" but the counter itself is computed by the
host, not the adapter. The actual logic is
`packages/local-runtime/src/thread-goal/store-verification.ts:83-107`
(`normalizeVerificationResult`), invoked from
`recordThreadGoalVerification` (same file, lines 16-81):

- Verdicts other than `not_met` reset `notMetStreak` to `0`.
- A `not_met` verdict hashes its normalized `missing` list into a
  `missingFingerprint` (SHA-256, line 96). If the new fingerprint matches the
  *previous* verification's fingerprint (i.e., the verifier is citing the
  same gaps again), `notMetStreak` increments; otherwise it resets to `1`.
- `recordThreadGoalVerification` then checks `repeatedGap` (lines 32-34):
  `verdict === 'not_met' && notMetStreak >= repeatedNotMetLimit`, where the
  limit defaults from `GOAL_CONFIG_DEFAULTS.verifier.repeatedNotMetLimit`
  (five, per the doc comment) and can be overridden per call. If tripped,
  the host's own decision is overridden and the Goal is force-paused with
  `paused(no_progress)` (line 36) regardless of what the caller's `decision`
  argument requested — a repeated, unchanging verifier gap always wins over
  "keep going."

This is a distinct breaker from the two below: it fires on *verifier*
disagreement across dispatches, not on the worker's own reply/tool pattern.

## The no-progress / no-tool breakers

Two more independent counters live on `ThreadGoalState`
(types.ts:292-303) and are governed by one shared occurrence ladder:

- `replyFingerprint` / `noProgressStreak` — consecutive turns whose final
  assistant text hashes identically (`fingerprintThreadGoalReply`,
  referenced from `packages/local-runtime/src/thread-goal/breaker.ts:2`).
- `noToolStreak` — consecutive Goal-bound turns that ended without a single
  committed tool call. A turn with an *untrustworthy* tool signal (`unknown`)
  resets this to `0` rather than counting as a trustworthy zero
  (types.ts:300-301) — an observation gap can never be stitched into a
  streak.

The ladder (`packages/local-runtime/src/thread-goal/store-breaker.ts:14-23`,
`decideAction`): occurrence 1 is recorded only, occurrence 2 nudges the
model (`renderNudgePrompt` in `continuation.ts:130-151`, choosing
`NO_PROGRESS_NUDGE` or `NO_TOOL_NUDGE` based on which counter fired), and
the configured `limit` (`threadGoalRepeatedReplyLimit`) pauses the Goal with
`paused(no_progress)`. The two conditions are combined via `strongerAction`
(store-breaker.ts:25-29): if either would pause, the Goal pauses; the
`cause` field (`'repeated_reply' | 'no_tool'`) records which one actually
tripped when both are candidates. `updateThreadGoalBreaker`
(store-breaker.ts:40-156) does this in one short SQL transaction guarded by
the Goal's decision epoch (optimistic concurrency — see below), and mutation
is skipped entirely (no epoch bump) when nothing changed
(lines 85-101), so an unrelated concurrent user PATCH cannot be silently
clobbered by a no-op breaker check.

`GoalBreaker.apply` (`packages/local-runtime/src/thread-goal/breaker.ts:73-118`)
is the settlement-time caller: it fingerprints the turn's final text, calls
`updateThreadGoalBreaker`, and on a `pause` action emits both a runtime event
and a `goal.updated` global event, then returns `{ action: 'stop' }` to the
turn pipeline. `clearNoToolStreak` (breaker.ts:56-71) is the inverse fast
path used when a turn used real tools but left the settlement pipeline
before the main breaker stage (e.g., it stopped for a questionnaire) — it
only ever clears the counter, never nudges or pauses, so a dependency wait
can't accidentally count toward the no-tool streak.

## The five-turn terminal audit

Independent of any breaker, every fifth Goal-bound turn gets a forced
self-check even with no other trigger.
`packages/local-runtime/src/thread-goal/reminder-policy.ts`:
`GOAL_TERMINAL_AUDIT_INTERVAL = 5` (line 3);
`selectGoalReminderPromptKind` (lines 8-22) fires when the Goal is `active`
and `turnsUsed % 5 === 0`. It composes with a pending recovery reminder
(after a retracted turn or runtime restart) into one of three prompt kinds:
`recovery`, `terminal-audit`, or `recovery-terminal-audit` (deduplicated so
the model gets one combined reminder, not two). The rendered prompts are
`DEFAULT_GOAL_TERMINAL_AUDIT_TEMPLATE` and
`DEFAULT_GOAL_RECOVERY_TERMINAL_AUDIT_TEMPLATE`
(`packages/agent-modules/goal/src/continuation.ts:90-103`): both instruct the
model to call `get_goal` as the durable source of truth and only call
`update_goal` if completion is proven or the blocked threshold is met —
explicitly *not* as a heartbeat. This is the mechanism that keeps a Goal from
silently running forever on "concrete progress" turns that never re-audit
against the objective.

## Verification dispatch and settlement contract

`VerificationAttempt` / `VerificationResult` / `VerifierPort`
(`packages/agent-modules/goal/src/verification/verifier-port.ts`) define the
adapter-neutral contract both backends implement. Key shape:

- `hostContext` (verifier-port.ts:28-40) is trusted host-authored fact —
  `completionProposal` (the worker's claim) and `settlement.transitionOnMet`
  (`'complete(verifier_met)'`) — versus `objective`/`transcript`/evidence,
  which are always wrapped as untrusted data for the verifier (same
  trust-boundary pattern AgentDelegation.md documents for the subagent
  prompt).
- The verdict domain (`VerificationVerdict`, lines 70-86) has four members:
  `met`, `not_met` (with `missing`), `impossible` (with `blocker`),
  `inconclusive` (with `code`). `impossible` is reachable only from the
  `evaluator` backend — `subagent.ts:175-186` documents why the subagent
  backend deliberately never emits it (see AgentDelegation.md).
- `VerificationDispatchError` (lines 122-135) carries a typed failure code
  (`route_unavailable`, `timeout`, `child_budget_exhausted`,
  `capability_violation`, `aborted`, etc.) plus whatever partial `usage` was
  observed before failure — these map onto the `paused(verifier_*)` status
  reasons in `THREAD_GOAL_STATUS_REASONS` (types.ts:97-102), split by owner
  layer per the comment there (see
  `verification-failure-reason.ts` for the code → reason table).

On the host side, `recordThreadGoalVerification`
(store-verification.ts:16-81) is the single transactional write: it re-reads
the Goal, checks staleness against the caller's expected decision epoch
(`threadGoalStaleDecision`), computes `notMetStreak`, applies the
repeated-gap override if tripped, and does a conditional `UPDATE ... WHERE
updated_at_ms = ? AND status = 'active'` — the same optimistic-concurrency
CAS pattern used by the breaker update, so a verifier verdict racing a user
`/goal pause` cannot land on a Goal it no longer applies to.

## Surfacing to the user/parent session

State reaches the user through two channels:

- **Internal runtime events** (`packages/local-runtime/src/thread-goal/events.ts:57-191`):
  a fine-grained stream — `goal.created`, `goal.admission_decided`,
  `goal.turn_bound`, `goal.turn_settled`, `goal.budget_decided`,
  `goal.breaker_decided`, `goal.verification_dispatched`,
  `goal.verification_child_started`, `goal.verification_decided`,
  `goal.worker_proposal_decided`, `goal.state_transitioned`,
  `goal.continuation_submitted`, `goal.reminder_injected`,
  `goal.queue_item_deferred` — used for logs/diagnostics, distinct from the
  coarser public state.
- **`thread_goal.updated` global/SSE event** (`packages/local-runtime/src/thread-goal/wiring.ts`,
  `publishThreadGoalEvent`): the coarse, client-facing `GlobalThreadGoal`
  snapshot (`packages/shared/src/global-events.ts`), which includes
  `lastVerification.notMetStreak` (global-events.ts:149) alongside status,
  streaks, and usage. This is what the TUI banner
  (`packages/tui/src/tui/features/goal/banner.ts`) renders: status label and
  color (`statusPresentation`, lines 171-179), an action hint per status
  (`actionHint`, lines 157-169, e.g. `/goal resume` for `paused`), the
  `executionWait` reason when `active`, and — when the last verification was
  `not_met` — a literal `not-met streak N` line (banner.ts:234).

The REST surface (`/goal/*`, mentioned in events.ts:10 and wiring.ts:31 as
having replaced an earlier hand-rolled router) is the user-facing control
plane referenced in types.ts:13-20 for pause/stop/resume; it consumes the
same `ThreadGoalState`/`GlobalThreadGoal` shapes described above rather than
a separate representation.

## Summary: what makes this a closed loop

Three independently-tripping breakers, layered on top of whichever
verification backend the route selected:

1. **Verifier disagreement** (`notMetStreak`, store-verification.ts) — the
   verifier keeps citing the same unmet gap.
2. **Worker non-progress** (`noProgressStreak` / `noToolStreak`,
   store-breaker.ts) — the worker's own replies or tool activity stall,
   independent of whether a verifier ever runs.
3. **Scheduled re-audit** (five-turn terminal audit, reminder-policy.ts) —
   a forced `get_goal` + objective-vs-evidence check even when nothing else
   fires, so "still working, still working" can't coast past the objective.

All three funnel into the same host-owned status/reason machinery
(`ThreadGoalStatus` + `ThreadGoalStatusReason`) under optimistic-concurrency
CAS writes, and all surface through the same `thread_goal.updated` event and
TUI banner — the model can propose but never durably decide `complete`;
independent evidence (verifier verdicts) and independent inactivity signals
(streaks, audits) are what the host actually settles on.
