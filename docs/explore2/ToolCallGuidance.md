# Tool Call Guidance: Making Agent Tool Calls Robust and Cheap

Found via `code-index --search "tool policy"` (hit `canonical-tool-policy.ts`) and
`--search "hint"` (hit the runtime-host directories); most other single-word
queries (`tool call`, `schema`, `validation`, `retry`, `polling`, `error`,
`guardrail`, `streaming`, `parallel`) returned no matches — same keyword-vs-
content gap noted in issue #006. The actual material was found by listing
`packages/agent-tools/src/desktop/` directly and reading files whose names
suggested the right concern (`output-limit.ts`, `path-guard.ts`,
`prepare-ask-user-arguments.ts`, `local-task-control.ts`).

## 1. Tolerant argument normalization before schema validation

`packages/agent-tools/src/desktop/prepare-ask-user-arguments.ts` is a clean
example of *robustness* aimed squarely at model fallibility rather than user
input: `prepareAskUserArguments` runs immediately before schema validation and
repairs shapes that some provider adapters are known to emit incorrectly —
a single flattened question instead of a `steps` array, legacy
`{ item: [...] }` wrappers for repeated fields, string `"true"`/`"false"`
instead of a real boolean for `multiSelect`. The comment is explicit about the
philosophy: *"Invalid wrappers remain untouched so validation fails closed"* —
this is normalization of *known* malformed-but-recoverable shapes, not a
general leniency pass that would let genuinely bad tool calls silently through.
The effect: fewer failed tool calls (and fewer round trips where the model has
to see an error, re-read the schema, and retry) for a known class of
model/provider quirks, without weakening validation for anything else.

## 2. Output size limiting with a continuation protocol, not blind truncation

`packages/agent-tools/src/desktop/output-limit.ts` bounds every large tool
result (`read`, `grep`, `bash`) to a fixed byte budget
(`DESKTOP_READ_TEXT_MAX_BYTES = 24KB`, `DESKTOP_GREP_CONTENT_MAX_BYTES = 16KB`,
`DESKTOP_BASH_MAX_BYTES = 24KB`) using one of two strategies:

- `limitDesktopPrefixLines` — keep the largest whole-line prefix that fits,
  binary-search-free (walks forward line by line), append a notice.
- `limitDesktopHeadTailLines` — split the byte budget 45%/55% between head and
  tail of the output (`headBudget`/`tailBudget`), so a huge command's *last*
  lines (often the actually relevant error or result) survive truncation
  alongside its first lines, rather than only the head surviving.

Both respect UTF-8 boundaries when cutting (`fitUtf8Prefix`/`fitUtf8Suffix`
walk back over continuation bytes) so truncation never produces a corrupt
multi-byte character.

Critically, truncation isn't a dead end: `DesktopOutputContinuation` embeds a
`continuation_hint` — `{ tool, preserve_args, instruction }` — directly in the
tool result's `details`, telling the model exactly which tool to call again,
with which arguments preserved, and a `next_offset`/`offset_unit` to resume
from. This converts "your output was cut off" from a dead-end error into a
structured, cheap-to-follow next step, and it means the model doesn't have to
re-read the *whole* file/output again to get the rest — it resumes.

## 3. Workspace path boundary — fail before the syscall, not after

`packages/agent-tools/src/desktop/path-guard.ts` — `resolveWithinWorkspace`
runs before any filesystem tool touches disk. It resolves `..`/absolute paths
and, importantly, walks the path component-by-component through `realpath`
so a symlink partway through the path (not just the final component) can't be
used to escape the workspace root; on `ENOENT`/`ENOTDIR` it keeps peeling
path components and retrying against the parent, so it correctly validates
paths to files that don't exist yet (e.g. a `write` target). This is a
robustness/safety guard rather than a token-reduction one, but it belongs in
"tool call guidance" because it turns a whole class of malformed or
adversarial tool calls into one clear, immediate error instead of an
inconsistent OS-level failure or a silent workspace escape.

## 4. Anti-repolling protocol hints embedded in tool output

`packages/agent-tools/src/desktop/local-task-control.ts` (`LocalTaskOutputTool`)
layers three kinds of in-band guidance directly into the `task_output` result
text, all aimed at stopping the model from making wasteful repeat calls:

- **Wait-time clamp notice** — server-side `wait_ms` is capped at 30s
  (`MAX_TASK_OUTPUT_WAIT_MS`); if the model requested more, a
  `<task_output_hint>` explicitly tells it the cap and what to pass next time,
  rather than silently clamping and leaving the model to infer why its long
  poll returned early.
- **Repolling detection** — `pollingHint` compares the current
  `(sessionId, turnId, taskId, status, nextOffset)` tuple against the last
  read *by that same tool instance*. If nothing changed since the last call,
  it injects an instruction to stop tight-polling and wait for the completion
  notification instead — a self-correcting nudge that fires exactly when it's
  needed (repeated no-op calls) rather than as static documentation the model
  has to remember unprompted.
- **Offset semantics reminder** — the hint text calls out that `offset=0`
  replays existing output and can return immediately even with `wait_ms` set,
  preempting a specific known model mistake (re-reading from the start
  expecting a real wait).

All three examples are the same idea: bake the "what should you do
differently" instruction into the tool result at the moment the mistake would
otherwise repeat, rather than relying on the model to have internalized it
from a system prompt turns earlier.

## 5. Tool-list scoping — fewer choices, less chance to pick wrong

`packages/agent-tools/src/desktop/canonical-tool-policy.ts` (also covered in
`docs/explore/AgentDelegation.md` and `docs/explore2/TokenReduction.md` from
the delegation/token angles) has a robustness dimension too, not just a token
one: `filterCanonicalNativeToolCeiling` removes tools that are structurally
inapplicable to a role (`write`/`edit`/`task` for an `explore` or `verifier`
subagent) *before* the model ever sees them. This isn't only cheaper — it
also means the model can't mistakenly reach for a mutating tool in a context
where using it would be wrong, closing off a class of tool-selection errors
at the list-construction boundary rather than relying on prompt instructions
("don't write files") that a smaller/distracted model might ignore under
pressure.

## Summary

| Concern | Mechanism | File |
|---|---|---|
| Malformed tool-call args from provider quirks | Normalize known-bad shapes before validation; fail closed on anything else | `prepare-ask-user-arguments.ts` |
| Oversized tool output | Byte-budgeted head/tail-preserving truncation + structured continuation hint | `output-limit.ts` |
| Filesystem tool calls escaping the workspace | Component-wise `realpath` boundary check, symlink-safe | `path-guard.ts` |
| Wasteful repeat polling calls | In-band hints injected exactly when the mistake pattern is detected | `local-task-control.ts` |
| Wrong tool for a role | Tool list filtered by role before the model sees it | `canonical-tool-policy.ts` |

The common thread: robustness and token-efficiency aren't separate concerns
here — a tool call that fails validation, returns an oversized blob with no
way to resume, or gets silently repeated all cost the same thing, a wasted
round trip. Every mechanism above turns a likely failure mode into either a
successful call (normalization), a bounded and resumable one (output limits),
or an impossible one (path guard, tool-list filtering).
