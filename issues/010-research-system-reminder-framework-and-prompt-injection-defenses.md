# 010 — Research: system-reminder framework and prompt-injection defenses

**Status**: Closed — Resolved
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

## Implementation

Wrote `docs/explore2/SystemReminders.md`, documenting the
`packages/agent-modules/system-reminder/` package in full:

- **Architecture**: chain-of-responsibility registry (`registry.ts`) +
  provider functions (`providers.ts`) + orchestrating service
  (`service.ts`), IO-free by design (host supplies a `DataCollector`).
- **Taxonomy**: ~25 providers grouped into identity/session framing,
  cold-start/onboarding (cooldown + audience gated), memory/skill-evolution
  routing (exponential backoff), mid-session update notices (event-driven
  one-shot), turn-scheduled operational nudges, and model/framework-specific
  tail providers (`mcode-tools-master-reminder.ts`, `plugin-reference.ts`).
- **`critical` providers and allowlists**: per-model SR-disable kill switch
  vs. cloud-runtime `ReminderPolicyEntry` allowlist/policy overrides — two
  independent gating mechanisms, documented with their different bypass
  semantics for critical providers.
- **Compaction interaction** (cross-referencing `docs/explore2/TokenReduction.md`):
  reminders are regenerated per turn, not part of compacted transcript state;
  the turn-count-based backoff schedulers explicitly detect a `turnCount`
  drop as a compaction signal and re-arm from the initial interval — but
  `agentContextProvider`'s full-vs-slim split has no such resync, a real gap
  noted in the doc.
- **Prompt-injection defense framing**: ties the framework to the two
  instances already documented in `docs/explore/AgentDelegation.md`
  (`BTW_SIDE_BOUNDARY`, `trust="untrusted_data"`) without re-deriving them,
  and identifies a third convention
  (`goal/src/verification/evaluator-adapter.ts:291-309`, structured
  `{ trust: 'untrusted_data', value }` objects) as a sibling of the XML-tag
  form.

Acceptance criteria met: type taxonomy, injection triggers, and
prompt-injection-defense framing are all covered; the two known instances are
cross-referenced rather than re-derived.
