# 008 — Research: how the permission/approval system enforces agent autonomy boundaries

**Status**: Closed — Resolved
**Priority**: P3
**Severity**: Minor
**Category**: Documentation

---

## Goal

Write `docs/explore2/PermissionSystem.md` documenting how this codebase
decides which agent actions require user confirmation versus running
autonomously, and how that boundary is enforced (not just where it's
configured). This is a candidate differentiator versus other agent CLIs,
where the autonomy/confirmation boundary is often a flat allow/deny list
rather than something contextual and structurally enforced.

## Background

Identified while researching tool-call robustness and token efficiency
(`docs/explore2/ToolCallGuidance.md`, `docs/explore2/TokenReduction.md`) —
`packages/agent-modules/permission` was noted via `code-index --search
"permission"` but not explored in depth; only the tool-list-ceiling angle
(`canonical-tool-policy.ts`) was covered so far, which is adjacent but not
the same as the approval/confirmation system itself.

## Starting points

- `packages/agent-modules/permission/src/index.ts` and the rest of that
  package.
- Cross-reference with `packages/agent-tools/src/desktop/canonical-tool-policy.ts`
  (already documented) to see how tool-ceiling and per-call permission
  checks relate or differ.
- Look for how risky/destructive actions (file writes, deletes, git push,
  shell commands) are classified and whether classification is static
  (per-tool) or dynamic (per-call, based on arguments).

## Acceptance criteria

- `docs/explore2/PermissionSystem.md` exists, covers what triggers a
  confirmation prompt, what runs autonomously, and where that decision is
  enforced in the code (not just documented in a policy file).
- Notes any prompt-injection-relevant angle (e.g. can untrusted content
  influence what gets auto-approved).

## Notes

Exploratory/documentation task, not a bug fix. Use `code-index` as a first
pass per the established workflow, expect to need direct `ls`/`grep`
fallback for parts the index doesn't surface (see
`docs/studies/CodeIndexEffectiveness.md`).

## Implementation

Wrote `docs/explore2/PermissionSystem.md`. Key findings:

- The tool-list ceiling (`canonical-tool-policy.ts`) and the permission
  engine (`packages/agent-modules/permission`) are two independent
  mechanisms: the ceiling statically hides tools per subagent role at
  list-construction time; the engine dynamically judges every call to a
  tool that *is* visible, based on parsed arguments.
- Classification is per-call, not per-tool: `bash-checker.ts` parses the
  actual command (unwrapping `bash -c`, heredocs, wrappers) against
  `HARD_BLOCKED_REGISTRY`/`SOFT_RISK_REGISTRY`; `fs-checker.ts` resolves the
  real path per invocation. The same tool name can land on either side of
  the boundary depending on arguments (e.g. `cat README.md` vs.
  `cat ~/.ssh/id_rsa`).
- The precedence (deny > bypass-immune ask > mode/bypass > rule allow >
  default) is enforced in `permission-core.ts::reducePermissionDecision` /
  `engine.ts::checkPermission`, with `isBypassPermissions` checked only
  after all bypass-immune branches have had a chance to return — so hard
  blocks are structurally unreachable by "always allow" mode, not just
  documented as such.
- Enforcement actually gates execution in
  `LocalPermissionFacade.checkPermission`
  (`packages/local-runtime/src/permissions/facade.ts`): a `deny` verdict
  never reaches `ExecutionPlan` construction, so nothing downstream can run
  the tool.
- Prompt-injection angle: `auto` mode's cloud classifier receives recent
  tool call/result text via `conversation-renderer.ts`, so untrusted tool
  output can influence its verdict — but `reducePermissionClassifierDecision`
  restricts the classifier to refining an existing `ask` into `allow`/`deny`;
  it can never override a deterministic `allow`/`deny` Core already reached,
  bounding how much an injected instruction can achieve.
