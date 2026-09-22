# 009 — Research: the Goal system's closed-loop objective verification

**Status**: Closed — Resolved
**Priority**: P3
**Severity**: Minor
**Category**: Documentation

---

## Goal

Write `docs/explore2/GoalVerification.md` documenting the `Goal` system's
broader closed-loop "did the agent actually finish the objective" tracking,
beyond the single verifier-subagent mechanism already covered in
`docs/explore/AgentDelegation.md`. This is a candidate differentiator versus
other agent CLIs, where completion is usually just "the agent claims done"
with no independent check.

## Background

`docs/explore/AgentDelegation.md` covered one piece —
`packages/agent-modules/goal/src/verification/subagent.ts`
(`createSubagentVerifierAdapter`) and its verdict mapping
(`pass/fail/partial -> met/not_met/inconclusive`), including the
`notMetStreak` breaker mentioned in a doc comment there. The rest of
`packages/agent-modules/goal` — state machine, retry/streak logic, how goals
are created/updated, other verification backends besides `subagent` — was
not explored.

## Starting points

- `packages/agent-modules/goal/` in full — `ls` it directly since
  `code-index` searches for "goal"/"objective"/"verification" were not yet
  tried systematically.
- The `notMetStreak` breaker across five independent verifications mentioned
  in `subagent.ts`'s doc comment — find and document the actual streak/state
  logic, not just the comment referencing it.
- Other `VerifierPort` backends besides `subagent`, if any exist.
- How a Goal's state transitions are surfaced to the user/parent session.

## Acceptance criteria

- `docs/explore2/GoalVerification.md` exists and documents the full Goal
  lifecycle (creation, attempts, verification, streak/retry logic, terminal
  states), not just the subagent verifier already covered elsewhere.
- Cross-references `docs/explore/AgentDelegation.md` rather than duplicating
  the verifier-subagent section.

## Notes

Exploratory/documentation task, not a bug fix.

## Implementation

Wrote `docs/explore2/GoalVerification.md`, covering:

- The `ThreadGoalStatus` lifecycle (types.ts) and the three owners of state
  transitions (model proposal, user REST surface, system accounting).
- Both `VerifierPort` backends — `subagent` (cross-referenced to
  `docs/explore/AgentDelegation.md` rather than re-derived) and the
  previously-undocumented `evaluator` adapter
  (`verification/evaluator-adapter.ts`), plus `verification-policy.ts`'s
  route-derived selection between them and `none`.
- The real `notMetStreak` breaker logic in
  `packages/local-runtime/src/thread-goal/store-verification.ts`
  (`normalizeVerificationResult` / `recordThreadGoalVerification`), which
  `subagent.ts`'s doc comment references but does not implement.
- The independent `noProgressStreak` / `noToolStreak` breakers in
  `packages/local-runtime/src/thread-goal/breaker.ts` and
  `store-breaker.ts`, and the separate five-turn scheduled terminal audit in
  `reminder-policy.ts` (`GOAL_TERMINAL_AUDIT_INTERVAL = 5`).
- How all of this surfaces to the user: the `thread_goal.updated` /
  `GlobalThreadGoal` event and the TUI goal banner
  (`packages/tui/src/tui/features/goal/banner.ts`), versus the finer-grained
  internal `goal.*` runtime event stream.
