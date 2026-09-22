# Permission System: How Autonomy Boundaries Are Decided and Enforced

Found via the `permission/src/index.ts` starting point in issue #008 plus
direct reads of `permission-core.ts`, `engine.ts`, the `tools/*` checkers, and
`packages/local-runtime/src/permissions/facade.ts` (the host that actually
calls the engine before a tool runs). Cross-referenced with
`packages/agent-tools/src/desktop/canonical-tool-policy.ts`, already covered
in `docs/explore2/ToolCallGuidance.md` §5.

## 1. Two independent boundaries, not one

The codebase has two structurally separate autonomy mechanisms that are easy
to conflate because both narrow what an agent can do:

- **Tool-list ceiling** (`canonical-tool-policy.ts`) — a *static, role-keyed*
  filter applied once when the tool list is assembled for a subagent
  (`filterCanonicalNativeToolCeiling`). It removes tools the model never even
  sees (`write`/`edit`/`task` for an `explore`/`verifier` role). This closes
  off a class of actions entirely; there is no per-call decision, no prompt,
  no way to earn access mid-turn. It only applies to `builtinAgent` +
  `isCanonicalSubagentRole` combinations (`canonical-tool-policy.ts:34-49`).
- **Permission engine** (`packages/agent-modules/permission`) — a *dynamic,
  per-call* decision made every time a tool the model *does* have access to
  is about to run, based on the tool name **and** its actual arguments
  (command string, file path, etc.). This is the system that answers "does
  this specific invocation need to interrupt the user."

The tool-ceiling shrinks the menu; the permission engine judges each order
placed from what's left on it. A `worker`-role agent that keeps `bash` on its
menu still goes through the full per-call engine below for every `bash`
invocation — the ceiling grants no exemption from it.

## 2. What triggers a confirmation prompt (`ask`) vs. runs autonomously

The decision is owned by `PermissionEngine.checkPermission`
(`packages/agent-modules/permission/src/engine.ts:202`), which dispatches to
one of two policy paths (`policyOwner: 'engine' | 'core'`, `engine.ts:209`):
the legacy in-engine three-step reducer (`checkPermissionWithEngine`) or the
newer `reducePermissionDecision` in `permission-core.ts:159` (`policyOwner:
'core'`, the default read in the host: `facade.ts` `readPermissionPolicyOwner`
defaults to `'core'` unless config says otherwise). Both implement the same
precedence, restated once as the source of truth in `permission-core.ts`:

1. **Whole-tool or MCP-server deny rule** → `deny`, bypass-immune.
2. **Tool/content checker deny** (e.g. a hard-blocked bash pattern) → `deny`,
   wins over broader ask rules so a narrow deny can't be diluted by a wide ask.
3. **Whole-tool ask rule** → `ask`, *unless* the sandbox exception applies
   (`toolName === 'bash' && context.sandboxEnabled &&
   context.autoAllowBashIfSandboxed`, `permission-core.ts:174-175`).
4. **Tool checker error** → `ask` (fail toward confirmation, never silently
   allow on an exception).
5. **`isInteractiveTool`** → `ask` (tool declared as requiring user
   interaction regardless of rules).
6. **Checker result `ask` + `bypassImmune`** → `ask`, and explicitly *cannot*
   be overridden by `bypassPermissions` mode later in the chain.
7. **Content-specific ask rule** (a persisted rule scoped to a specific
   command/path substring) → `ask`, also bypass-immune (§4.2 point 3 in the
   design doc referenced in `engine.ts:15`).
8. **Checker `ask` for a plain safety reason**, not in bypass mode → `ask`.
9. Only after all of the above: **bypass mode** → `allow`; then **whole-tool /
   MCP allow rule** → `allow`; then **checker `allow`** → `allow`.
10. **No checker registered for the tool** → `allow` by default (unknown,
    presumably harmless tools are not blocked by omission).
11. **Checker registered but returned nothing conclusive** → `ask` (fail
    closed, not open, on ambiguity).

The key structural point: **deny and bypass-immune asks are evaluated before
bypass/allow mode is even consulted.** `bypassPermissions` (the "never ask"
mode) only gets a chance to flip a decision to `allow` at step 9 — hard
blocks, sandbox-exempt asks, and bypass-immune content asks are checked
first and are unreachable by that mode. This is enforced in code, not just
documented: `isBypassMode(context)` is called at a fixed point deep in the
reducer (`permission-core.ts:253`, `engine.ts:438`), after all the
bypass-immune branches have already had the chance to return.

## 3. Static vs. dynamic classification

Classification is **dynamic, per-call**, not a flat per-tool table — the flat
table exists (`canonicalRole` ceiling above) but only governs *visibility*,
not per-invocation judgement.

- `packages/agent-modules/permission/src/classifier/dangerous-patterns.ts` is
  the pattern registry: `HARD_BLOCKED_REGISTRY` (catastrophic/irreversible —
  final deny, immune to bypass) and `SOFT_RISK_REGISTRY` (suspicious but
  possibly legitimate — `chmod 777`, `sudo`, `python -c`, container exec;
  forced out of the fast-allow path into ask/cloud-LLM judgement rather than
  either flat-allowed or flat-denied).
- `tools/bash-checker.ts` parses the actual command string (`bash-ast.ts`,
  `bash-split.ts`, heredoc body ranges, wrapper unwrapping via
  `bash-wrapper-unwrap.ts`) and extracts path/read/write intents
  (`path-capability.ts`) before matching against these registries — a `rm`
  wrapped in `bash -c "..."`, a heredoc, or an `eval` all get resolved to
  their actual effect rather than judged by the literal string. `rm` targets
  specifically get rewritten to a recoverable-trash command
  (`buildTrashCommand`/`parseRmTargetsWithMetadata`) rather than a flat
  block, so classification isn't binary allow/deny per command name — it can
  produce a *third* outcome: rewrite-then-allow.
- `tools/fs-checker.ts` extracts the file path from tool input
  (`extractFilePath`) and calls `evaluatePathCapability`, which checks
  workspace boundaries, sensitive-path patterns (secrets, credentials,
  private keys — checked *before* the general workspace-boundary check, per
  the file's own docstring), and produces per-call `allow`/`ask`/`deny`.
- Two commands with the identical program name can land on opposite sides of
  the boundary purely from arguments: `cat README.md` resolves as a read
  intent inside the workspace (`allow`); `cat ~/.ssh/id_rsa` matches the
  sensitive-path check regardless of the tool being the same `bash`/`read`
  call (`ask`/gated through the LLM classifier).

So classification is fundamentally per-call: tool name selects *which*
checker runs, but the checker's verdict is computed from the parsed
arguments every time, not looked up from a static per-tool table.

## 4. Where the boundary is enforced in code, not just configured

`docs/*.md` policy language would be advisory without a call site that
actually blocks execution on the verdict. That call site is
`LocalPermissionFacade.checkPermission`
(`packages/local-runtime/src/permissions/facade.ts:449`), invoked from the
API routes that handle tool execution
(`packages/local-runtime/src/api/routes/permissions.ts:99,194`,
`hosted-agent-permission-capabilities.ts:50`) — i.e. this runs on the request
path *before* the tool's actual side-effecting code executes, not as a
logging/audit pass after the fact. The facade:

1. Resolves the effective `PermissionMode` for this call (plugin-hook
   override → per-call override → config), snapshotting it once
   (`rawMode`) so a mid-call mode change (`bypassPermissions` → `default`
   via `PUT /config`) can't let a stale bypass decision leak through — the
   comment at `facade.ts:435-447` spells out the specific race this
   snapshot prevents.
2. Calls `checkPermissionRaw`, which runs the engine/core reducer above.
3. Applies `applyAskGate(verdict, rawMode)` — a second, host-level gate
   distinct from the engine's own bypass check.
4. Only if the verdict isn't `deny` does it build an `ExecutionPlan`
   (`createPermissionExecutionPlan`) carrying the possibly-rewritten input
   (e.g. `rm` → trash command) that the tool executor actually runs. A
   `deny` verdict never reaches plan construction, i.e. never reaches
   anything that could execute a rewritten or original command.

Every decision is logged with `raw_behavior` vs. `final_behavior` and
`ask_gate_applied` (`facade.ts:483-495`) specifically so a bypass-mode
override is distinguishable after the fact from an engine-level allow —
enforcement and its audit trail are the same code path, not reconstructed
from separate log lines.

## 5. `auto` mode: the cloud LLM gate, and the prompt-injection angle

`auto` mode (`AskForApproval: 'on-request-llm'`, `ask-policy.ts:33-53`) does
not remove the engine's deterministic layer — hard blocks and bypass-immune
asks still apply unchanged — it adds a *refinement* step for the
deterministic `ask` outcomes only:
`reducePermissionClassifierDecision` (`permission-core.ts:307-323`) explicitly
only touches decisions already at `ask`; it can turn an `ask` into `allow` or
`deny` but **cannot override an `allow` or `deny` that Core already
finalized**. This is the load-bearing invariant for the prompt-injection
question below.

The classifier call goes through `HttpCloudGatewayClient` →
`POST /mavis/api/v1/permission/check`
(`classifier/cloud-classify-client.ts`), and the request body includes a
`conversation_context` block built by
`conversation-renderer.ts::renderConversationContext`. That function renders
recent user messages **and recent tool calls with their results**
(`"tool:<name> args=<truncated> result=<truncated>"`,
`conversation-renderer.ts:1-27`) into the text the cloud LLM reasons over
before returning its allow/deny/ask verdict.

This is the prompt-injection-relevant seam: tool *output* — which can contain
attacker-controlled content from a file, a fetched URL, or a shell command's
stdout — flows into the same context the classifier uses to decide whether to
auto-approve a *different, later* action. Mitigations that bound the blast
radius rather than eliminate the risk:

- Truncation limits (`TRUNCATE_LIMIT = 200` chars for tool calls/results,
  500/250 for user messages) cap how much injected text can reach the
  classifier per turn.
- The classifier is strictly a downgrade-only refiner of `ask` (§ above) — it
  can turn a deterministic `ask` into `allow`, but it can never override a
  deterministic `deny` or a bypass-immune `ask`. So injected content can at
  most cause an auto-approval of something that was merely *ambiguous* by
  the static checkers (soft-risk bash, borderline path), never something the
  hard-blocked registry or a bypass-immune content-ask rule already refused.
- `shouldUseCloudClassify()` (`cloud-classify-client.ts`) gates the whole
  cloud path to `isManagedRuntime()` — unmanaged/local daemons never send
  conversation context off-box for this purpose, and never get the
  LLM-refinement upgrade path at all (deterministic-only in that mode).

No code path lets tool *output* content mint a persisted allow rule or write
directly to the rule store; `ProposedRule` suggestions
(`ask-policy.ts:60-80`) are surfaced to the UI as a prefill for a human
"always allow" click (`source: 'engine' | 'cloud-gateway'`), not
auto-committed.

## Summary

| Question | Mechanism | Where enforced |
|---|---|---|
| Can the model even call this tool? | Static role ceiling | `canonical-tool-policy.ts` (list construction time) |
| Does this specific call need confirmation? | Dynamic per-call reducer over parsed command/path intent | `permission-core.ts::reducePermissionDecision` / `engine.ts::checkPermission` |
| Can a hard block ever be silently bypassed? | No — checked before `isBypassMode` | `permission-core.ts:174,253`, `engine.ts:309,438` |
| Where does the verdict actually stop execution? | Facade call before `ExecutionPlan` construction | `facade.ts:449-500`, `routes/permissions.ts:99,194` |
| Can untrusted tool output swing an approval? | Only refines existing `ask`→`allow`/`deny`, truncated, managed-runtime only | `permission-core.ts:307-323`, `conversation-renderer.ts`, `cloud-classify-client.ts` |

The differentiator over a flat allow/deny list: the boundary is a layered
reducer with an explicit, code-enforced precedence (deny > bypass-immune ask
> mode/bypass > rule allow > default), evaluated fresh per call against
parsed command/path intent rather than a static per-tool lookup, with the
one externally-influenced input (cloud classifier context) structurally
restricted to narrowing, not widening, what a deterministic `deny` already
decided.
