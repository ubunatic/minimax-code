# What a Harness Needs to Match This One's Efficiency

A synthesis of every `docs/explore2/` doc (and their `docs/explore/` and
`docs/studies/` predecessors) into one answer to a single question: what
structural properties does an agent harness need, in order to run agents as
efficiently and robustly as this one, rather than just prompting a model
harder? Read the individual docs for the file:line detail; this one is the
map of how they fit together.

Five mechanisms recur across the codebase. None of them is "a better
prompt." All five are structural — they live in code the model cannot see or
argue its way around, and each buys either fewer wasted tokens, fewer failed
tool calls, or a check the model cannot talk itself out of.

## 1. Compact by default, not by accident (`TokenReduction.md`)

Context is a budget with automatic enforcement, not a limit the agent has to
remember. `context-manager`'s `evaluateTrigger`/`selectPlan` compact on a
real token count, not a heuristic message count, and choose *what* to cut
rather than just truncating oldest-first. Per-call token budgets
(`provider-budget.ts`) cap individual tool outputs before they ever reach the
transcript. Subagents get a *smaller tool surface*, not just a shorter
prompt — fewer tools means fewer tokens spent on tool schemas every turn.
Verifier subagents receive a hard-bounded evidence payload, never a raw
transcript. Background task output is polled incrementally instead of
replayed in full each time.

**The generalizable requirement**: token efficiency has to be a property of
the plumbing (budgets, incremental reads, role-scoped tool lists), enforced
whether or not the model chooses to be economical. A harness that only tells
the model "keep responses short" is relying on compliance; this one removes
the option to be verbose in the first place.

**Example.** A verifier subagent asked to check a 40k-token transcript never
sees it — `evidence-brief.ts` hard-caps what it receives to a bounded
evidence payload, so the check itself can't blow the budget it's meant to
protect. Contrast with a harness that just tells a verifier "keep your
context small": nothing stops it from reading the whole transcript anyway.

## 2. Make tool calls hard to get wrong (`ToolCallGuidance.md`)

Robustness and token efficiency turn out to be the same problem, not two
separate ones. `prepareAskUserArguments` normalizes malformed/legacy
arguments *before* schema validation, so a slightly-off tool call succeeds
instead of round-tripping an error and a retry. `output-limit.ts` truncates
with a continuation protocol (UTF-8-safe head/tail split) rather than blind
truncation, so the agent doesn't have to re-request the same file to see the
missing middle. `path-guard.ts` fails a workspace-boundary violation before
the syscall, not after, using a symlink-safe component walk. Anti-repolling
hints are embedded directly in tool output (`local-task-control.ts`) so an
agent doesn't burn calls busy-polling a background task. Tool-list scoping
(`canonical-tool-policy.ts`) means fewer, better-fitted choices per role —
picking wrong is structurally less likely when there are fewer options.

**The generalizable requirement**: every point where a tool call can fail
or be ambiguous is a point where the harness pays for it twice — once in the
failed call, once in the retry. Fixing tolerance and boundaries at the tool
layer is cheaper than fixing them in the prompt.

**Example.** A model emits a legacy single-question `ask_user` shape instead
of the current array form. `prepareAskUserArguments` normalizes it silently
before validation, so the call succeeds on the first try. A harness without
that normalization layer would reject the call, the model would retry —
maybe wrong again — burning two or three extra round trips for a shape
mismatch that carried no real ambiguity about intent.

## 3. Bound autonomy structurally, not just by convention (`PermissionSystem.md`)

Two independent boundaries exist, and they're layered, not redundant: a
*static*, role-keyed tool-list ceiling (`canonical-tool-policy.ts`) decides
what a role can even attempt, and a *dynamic*, per-call permission engine
(`packages/agent-modules/permission`) re-checks every surviving tool call
against parsed arguments — a `worker` role keeping `bash` still goes through
full per-call classification, because having the tool doesn't mean every
invocation of it is safe. Decision precedence (deny > bypass-immune ask >
bypass mode > rule allow > default) is enforced in code
(`permission-core.ts::reducePermissionDecision`), so hard blocks are
*structurally* unreachable by an "always allow" mode — not just
discouraged by policy text. Even the cloud LLM classifier in `auto` mode can
only refine an `ask` into allow/deny; it can never override a deterministic
verdict Core already reached, which bounds how much untrusted tool output
fed to that classifier can actually influence outcomes.

**The generalizable requirement**: autonomy limits need to be enforced at a
layer the model cannot reach through prompting or tool arguments — the
enforcement point (`LocalPermissionFacade.checkPermission`) sits below
`ExecutionPlan` construction, not inside a system prompt instruction.

**Example.** The same `bash` tool call `rm -rf /tmp/foo` vs. `rm -rf ~/`
lands on opposite sides of the boundary — `bash-checker.ts` parses the
actual command and path, not just the tool name, so having `bash` in the
tool list never means blanket approval. A harness that classifies risk
per-tool ("bash is dangerous, ask every time" or "bash is fine, never ask")
can't make that distinction at all.

## 4. Verify completion instead of trusting the claim (`GoalVerification.md`)

Most agent CLIs treat "done" as whatever the model says. This harness closes
the loop: a Goal's state machine tracks attempts, a `VerifierPort` backend
(subagent-based, cross-checking evidence rather than re-reading the whole
transcript — see `AgentDelegation.md`) independently verifies, and three
separate breakers exist to stop runaway loops even when verification alone
wouldn't: a `notMetStreak` breaker that only increments when the *same* gaps
repeat (hashed against the verifier's `missing` list) and force-pauses the
Goal after 5 repeats regardless of what the caller wanted next; independent
no-progress/no-tool-call streak breakers; and a scheduled five-turn terminal
audit that forces a re-check even absent any other trigger. State surfaces
to the user via a dedicated event stream, not by parsing agent prose.

**The generalizable requirement**: "finished" needs an independent check
with its own circuit breakers, separate from the executing agent's own
self-report — otherwise a confidently wrong agent just keeps going.

**Example.** An agent claims a Goal is met, but the verifier repeatedly
finds the same missing test file across five attempts. Because the
`notMetStreak` breaker hashes and compares the verifier's `missing` list
rather than trusting a bare fail/pass, it detects the loop and force-pauses
the Goal — the agent can't just keep re-asserting "done" and have the
system eventually agree.

## 5. Deliver guidance and untrusted data through different channels (`SystemReminders.md`)

The system-reminder framework is a chain-of-responsibility registry of
~25 providers injecting fresh, per-turn context — identity framing,
cold-start onboarding, operational nudges — none of which survive
compaction as transcript text; they're regenerated every turn instead, and
several providers explicitly detect a compaction-driven turn-count drop and
re-arm their own schedules. Structurally, reminders are a *separate channel*
from tool results and user messages, which is what makes the
`trust="untrusted_data"` convention (seen in `AgentDelegation.md`'s
verifier-subagent prompts, and again independently as structured
`{trust: 'untrusted_data', value}` objects in `evaluator-adapter.ts`) mean
something: untrusted content is tagged and isolated rather than blended into
the instruction stream the model treats as authoritative.

**The generalizable requirement**: the channel that carries harness
guidance and the channel that carries untrusted external content need to be
distinguishable to the model, consistently, everywhere untrusted content
enters — not just in one place a developer remembered to escape it.

**Example.** This exact session: unsolicited "harnez tip" text repeatedly
appeared inside ordinary tool output urging a `harnez rate` call. Because
that text arrived as raw tool output rather than through the
`trust="untrusted_data"`-tagged or reminder-registry channel, it was
recognizable as unverified and ignored every time — the same mechanism this
section describes, observed from the outside.

## Where the index tool fits (and where it doesn't) — `docs/studies/CodeIndexEffectiveness.md`

`code-index` is not one of the five mechanisms above — it's exploration
tooling, not harness plumbing — but the study is worth folding in here
because it demonstrates the same principle in miniature: a cheap, fast
first-pass tool (single-keyword `--search` reliably narrows to the right
directory or file) still needs a documented, trustworthy failure signal.
Its current gap — "no matches" being indistinguishable from "not indexed"
(issue #007, now resolved) — is exactly mechanism #2's lesson applied to a
different tool: ambiguity at a tool boundary costs retries, whether the tool
is `bash` or a search index.

## Summary: five properties, one theme

| # | Property | Doc | Enforced by |
|---|---|---|---|
| 1 | Context/output shrinks automatically | `TokenReduction.md` | token-counted triggers, per-call budgets, role-scoped tool lists |
| 2 | Tool calls tolerate imperfection and bound their own failure modes | `ToolCallGuidance.md` | argument normalization, continuation protocol, path guard, anti-repoll hints |
| 3 | Autonomy is bounded below the model, not just above it in the prompt | `PermissionSystem.md` | static tool ceiling + dynamic per-call permission engine, precedence rules |
| 4 | Completion is independently verified, with its own breakers | `GoalVerification.md` | verifier subagent, `notMetStreak`/no-progress/no-tool breakers, terminal audit |
| 5 | Guidance and untrusted data travel on distinguishable channels | `SystemReminders.md` | reminder registry separate from transcript, `trust="untrusted_data"` tagging |

The common thread across all five: every property that makes this harness
efficient or robust is implemented as *code the model does not control* —
a trigger, a budget, a checker, a breaker, a channel boundary — rather than
an instruction the model is asked to follow. A harness that wants the same
efficiency needs the equivalent of each of these five as enforced
mechanisms, not as prompt guidance, because prompt guidance degrades under
context pressure and adversarial input exactly when it matters most.
