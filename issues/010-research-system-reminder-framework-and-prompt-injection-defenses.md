# 010 — Research: system-reminder framework and prompt-injection defenses

**Status**: Open
**Priority**: P3
**Severity**: Minor
**Category**: Documentation

---

## Goal

Write `docs/explore2/SystemReminders.md` documenting the system-reminder
framework as its own topic: what reminder types exist, when each is
injected, and how the framework functions as a prompt-injection defense
layer. Candidate differentiator versus other agent CLIs, where this kind of
in-band, context-sensitive guidance is usually absent or much thinner.

## Background

Only fragments of this were seen while researching other topics:
`packages/agent-modules/system-reminder/src/blocks.ts` was grepped (not read
in full) and turned up `MAX_SESSIONS_PER_AGENT = 5`; the `BTW_SIDE_BOUNDARY`
reminder (`conversation-fork-service.ts:31-48`) and the
`trust="untrusted_data"` wrapping pattern used for verifier subagent prompts
(`goal/src/verification/subagent.ts`) were documented in
`docs/explore/AgentDelegation.md` as two independent instances of the same
underlying idea, but the general framework producing these reminders was
never explored directly.

## Starting points

- `packages/agent-modules/system-reminder/` in full.
- Enumerate reminder types/triggers, not just the two instances already
  found.
- How reminders compose with context compaction (`docs/explore2/TokenReduction.md`)
  — do reminders survive a compaction cut, get regenerated, or get dropped?
- Any other untrusted-data-wrapping conventions beyond the verifier-subagent
  one already documented.

## Acceptance criteria

- `docs/explore2/SystemReminders.md` exists, covers the reminder type
  taxonomy, injection triggers, and the prompt-injection-defense framing.
- Cross-references the two known instances already documented in
  `docs/explore/AgentDelegation.md` rather than re-deriving them.

## Notes

Exploratory/documentation task, not a bug fix.
