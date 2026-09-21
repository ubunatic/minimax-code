# 001 — Make the repository Harnez-ready for solo development

**Status**: Open
**Priority**: P1
**Severity**: Major
**Category**: Maintenance

---

## Goal

Convert this fork from an upstream-publication-oriented MiniMax repository into a solo-developer Harnez repository. Remove or replace upstream synchronization, publication, npm-oriented, and multi-contributor workflow assumptions that are outside the fork's scope.

## Acceptance criteria

- Repository guidance describes the fork as the source of truth and no longer requires upstream publication or synchronization workflows.
- Harnez-managed guidance matches the actual repository layout and tooling.
- The project has one concise, discoverable development workflow for the solo maintainer.
- No documentation instructs agents to run npm, npx, pnpm, or other host-side JavaScript tooling.
- The resulting guidance does not claim verification that has not been performed.
