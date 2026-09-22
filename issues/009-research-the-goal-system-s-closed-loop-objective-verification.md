# 009 — Research: the Goal system's closed-loop objective verification

**Status**: Open
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
