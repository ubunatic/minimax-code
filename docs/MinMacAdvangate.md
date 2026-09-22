# Why this harness may have an advantage

## Verdict

On the same backend model, MiniMax Code may be better when the task benefits
from bounded context, explicit tool contracts, resumable sessions, or parallel
research followed by serialized edits. That is a design hypothesis, not a
measured ranking. This repository contains no controlled head-to-head benchmark
against Claude Code, Codex, or Agy.

## Claims and public numbers

| Claim | Number | Source | Confidence |
|---|---:|---|---|
| Long-history performance is an explicit workload | 300 rounds × approximately 8 KiB; 1,048,576 UTF-16-unit cache budget exceeded | [`docs/performance-ci.md`](performance-ci.md) | Evidence for the workload; no product advantage shown |
| Ordinary regression testing has a repeatable paired method | 3 measured alternating pairs, with up to 2 confirmation pairs | [`docs/performance-ci.md`](performance-ci.md), [`scripts/perf/config.json`](../scripts/perf/config.json) | High for the method |
| Performance thresholds are bounded | Duration: 25% and 1 second; CPU: 20% and 0.5 core-seconds; RSS: 20% and 32 MiB | [`docs/performance-ci.md`](performance-ci.md), [`scripts/perf/config.json`](../scripts/perf/config.json) | High for repository policy, not a competitor comparison |
| Delegation can return incrementally rather than block the parent | Up to 30 seconds per `task_output` wait | [`docs/MultiAgentFlow.md`](MultiAgentFlow.md), [`packages/agent-tools/src/desktop/local-task-control.ts`](../packages/agent-tools/src/desktop/local-task-control.ts) | High for the interface |
| The product reuses one runtime across entry points | 3 surfaces: TUI, headless `exec`, and ACP | [`docs/HarnessToolsAndProduct.md`](HarnessToolsAndProduct.md), [`docs/architecture.md`](architecture.md) | High for architecture; comparative drift is unmeasured |
| The harness has multiple bounded guidance mechanisms | 8 documented guidance mechanisms | [`docs/AgentGuidanceTricks.md`](AgentGuidanceTricks.md) | High for the document inventory; effect size is unknown |

No number in this table is a result against Claude Code, Codex, or Agy. The
repository does not publish matched-task wall time, token use, defect rate,
review acceptance, or quality scores for those products.

## Speed

The plausible speed case is reduced overhead around the model: provider
implementations are lazy-loaded, stable system/tool assembly can remain
cache-compatible while history grows, history persistence uses deltas, streamed
events are rendered incrementally, and read-only tools can be parallel while
mutations stay sequential. The evidence is [`docs/EfficiencyAndSpeed.md`](EfficiencyAndSpeed.md),
with implementation pointers to
[`packages/agent-core/src/pi-turn-runner/assembly-fingerprint.ts`](../packages/agent-core/src/pi-turn-runner/assembly-fingerprint.ts),
[`packages/agent-core/src/pi-turn-runner/history.ts`](../packages/agent-core/src/pi-turn-runner/history.ts),
and [`packages/agent-core/src/tools/builtin-defs.ts`](../packages/agent-core/src/tools/builtin-defs.ts).

The public performance suite is synthetic, offline, and explicitly does not
measure live-model quality, latency, compaction acceptance, or other products'
rankings ([`docs/performance-ci.md`](performance-ci.md)). Therefore “faster” is
a hypothesis until the same tasks, model, permissions, and concurrency budget
are run against each harness.

## Quality

Quality may benefit from read-before-edit and exact-match edit rules, bounded
tool output, classified retries, typed adapters, independent verifier paths, and
ordered repository gates. These mechanisms are documented in
[`docs/QualityMechanisms.md`](QualityMechanisms.md) and
[`docs/HarnessToolsAndProduct.md`](HarnessToolsAndProduct.md). They provide
strong process evidence, not a measured first-pass-success or defect-rate
advantage.

## Multi-agent flow

The parent can delegate a self-contained task to a fresh role-constrained child,
run independent work in the background, receive structured output, append a
follow-up, and stop work through task handles. The documented roles are
`explore`, `worker`, `verifier`, and `mavis`; the flow requires parallel reads
but one writer per workspace. See [`docs/MultiAgentFlow.md`](MultiAgentFlow.md).

This could outperform a single generalist on decomposable work, especially when
verification is independent. The repository declares no numeric concurrency
limit and has no same-model comparison, so the size of the advantage is unknown.

## Guidance tricks

The notable pattern is layered guidance: stable operational contracts, selective
state-triggered reminders, execution-boundary hooks, ranked skills, explicit
todo/plan state, tool descriptions that carry policy, bounded steering, and
runaway-loop nudges. See [`docs/AgentGuidanceTricks.md`](AgentGuidanceTricks.md).
The likely benefit is fewer invalid mutations and less repeated work; this is a
hypothesis, and the reminder/token cost is not measured.

## Caveats

- The benchmark numbers above describe fixtures and regression policy, not user-facing speed.
- Backend model, provider latency, permissions, machine, task mix, and concurrency can dominate results.
- The public docs explicitly identify comparative claims as hypotheses.
- No head-to-head benchmark against Claude Code, Codex, or Agy exists in this repository.

## Summary and open questions

The strongest defensible case is architectural: MiniMax Code combines bounded
context, prescriptive tools, resumable task state, selective concurrency, and
verification hooks in one runtime. Whether that beats another harness on the
same model remains unproven.

Open questions are the missing paired measurements: wall time, time to first
useful output, input/output/cache tokens, retries, defects, verifier catches,
and reviewer acceptance on an identical public task suite.
