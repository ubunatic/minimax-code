# Efficiency and speed

This is a repository-based analysis of why the MiniMax harness can be efficient on the same backend models. It documents mechanisms visible in source, not a claim that the repository proves superiority over Claude Code, Codex, or Agy.

## Evidence boundary

| Claim | Evidence | Interpretation |
|---|---|---|
| Long histories are a first-class performance concern | `docs/performance-ci.md` defines `upstream-100` and `history-300`; the latter is 300 rounds × about 8 KiB and deliberately exceeds a 1,048,576 UTF-16-unit cache budget | The harness is measured under history growth, not only startup latency |
| The benchmark measures process efficiency, not model quality | `docs/performance-ci.md` says the fixture is synthetic, uses an offline network fixture, and does not measure live-model quality, compaction acceptance, latency, or product rankings | Any quality/speed advantage remains an inference until live paired tests exist |
| Provider startup is reduced by lazy loading | `third_party/pi-mono/packages/ai/src/providers/register-builtins.ts` defines `createLazyStream`/`createLazySimpleStream` and dynamic provider imports | A run need not load every provider implementation before its first request |

## Main speed and cost levers

### History reuse and cache-stable assembly

`packages/agent-core/src/pi-turn-runner/assembly-fingerprint.ts` hashes the provider-visible system prompt and tool interfaces with SHA-256, truncating each fingerprint to 16 hex characters. It preserves prompt whitespace, preserves tool order, and canonicalizes object keys. Crucially, the comment and implementation exclude message history: normal loop growth does not look like system/tool assembly churn.

This separates two kinds of change:

```text
stable system prompt + stable tools + growing messages
        └──────────── cacheable prefix ────────────┘
```

Recipe for another harness: fingerprint only the cache-relevant prefix (prompt plus tool definitions), canonicalize structured data, preserve exact prompt bytes, and do not include the append-only conversation tail. Use the fingerprint to detect accidental prompt/tool churn without invalidating history reuse.

The upstream AI contract exposes `cacheRetention` (`none|short|long`) and `sessionId` in `third_party/pi-mono/packages/ai/src/types.ts`; it also reports `cacheRead` and `cacheWrite` usage. The default retention is documented as `short`. This is direct support for provider/session prompt caching, but repository code alone does not establish cache hit rates or dollar savings.

### Hashing and query overhead

Hashing is bounded: assembly fingerprints hash only the system prompt and canonical tool interfaces, while history is excluded (`packages/agent-core/src/pi-turn-runner/assembly-fingerprint.ts`). The performance observer records request bytes and validation duration without headers (`docs/performance-ci.md`), making wire growth visible without adding secret-bearing metadata.

The history runner also uses a cursor rather than repeatedly persisting the complete history: `packages/agent-core/src/pi-turn-runner/history.ts` stores `lastSnapshotLength` and flushes only `agent.state.messages.slice(lastSnapshotLength, total)` as `messageDelta`. Replacement compaction sends a deliberate full replacement. This is an important O(delta) persistence path.

### Runtime loading

AI providers are registered through lazy wrappers and dynamic imports in `third_party/pi-mono/packages/ai/src/providers/register-builtins.ts`. V2 prompt assets are captured once into a `Map` by `captureLocalPromptAssets` (`packages/local-runtime-v2/src/service/agent/builtin/prompt-assets.ts`), then templates are rendered from that in-memory snapshot (`prompt-renderer.ts`). Together these avoid repeatedly reading the prompt package while preserving a coherent prompt snapshot for a run.

Copyable recipe:

```ts
const assets = await captureAssetsOnce(assetRoot);
const prompt = render(templateFrom(assets), stableContext);
```

Lazy-load optional integrations; eagerly snapshot small, immutable prompt assets once per runtime/session; avoid filesystem reads inside every turn.

### Prompt caching

`third_party/pi-mono/packages/ai/src/types.ts` supports `cacheRetention`, `sessionId`, provider-specific `cacheControlFormat`, and usage accounting (`cacheRead`/`cacheWrite`). `assembly-fingerprint.ts` is designed around this boundary: stable prompt/tool assembly can remain cache-compatible while messages grow.

Recipe: keep the system prompt and tool schemas byte-stable, pass one stable session identifier for a conversation, choose retention explicitly, and record cache-read/write usage beside input/output tokens. Do not claim savings from the option alone; verify provider usage telemetry.

### Streaming

The AI layer’s primary API is `AssistantMessageEventStream` (`third_party/pi-mono/packages/ai/src/types.ts` and `utils/event-stream.ts`). Stream events expose start, partial updates, and terminal success/error/abort states. The TUI supervisor consumes runtime events with `for await` (`packages/tui/src/headless/supervisor.ts`).

Streaming improves time-to-first-visible-output and permits incremental tool/progress rendering; it does not necessarily reduce total model tokens. The TUI changelog records a concrete UI optimization: coalescing `requestRender()` calls to a 16 ms frame budget under heavy streaming (`third_party/pi-mono/packages/tui/CHANGELOG.md`).

### Tool-call parallelism

The canonical tool definitions mark read-only `read`, `grep`, and `glob` as `executionMode: 'parallel'`, while `write`, `edit`, `bash`, and `todowrite` are sequential (`packages/agent-core/src/tools/builtin-defs.ts`). This is a quality-preserving concurrency boundary: independent observations can overlap, mutations remain ordered.

Recipe: classify tools by side effects; execute only explicitly parallel-safe calls concurrently; serialize mutations and stateful commands. The schema-level mode is a useful contract for local/cloud implementations to share.

### Compaction, truncation, and token budgeting

The context-manager extension installs a `before_llm_call` hook whose decision can replace messages, and describes its purpose as compacting history before each LLM call (`packages/agent-extension/src/context-manager.ts`). Replacement history is persisted distinctly from incremental deltas through `on_history_changed`.

Tool context is measured in both bytes and estimated tokens by `packages/agent-core/src/pi-turn-runner/tool-context-size.ts`; its histogram buckets reach 256,000 tokens and 16,000,000 bytes. Measurement is fail-open: if token estimation fails, byte counts remain available. Tool definitions also instruct read/search tools to truncate output and page long files (`packages/agent-core/src/tools/builtin-defs.ts`).

The AI contract exposes `maxTokens` and thinking-level budgets (`third_party/pi-mono/packages/ai/src/types.ts`; `providers/simple-options.ts` computes max output while respecting model limits). These controls bound expensive output and reasoning. The performance fixture deliberately sets a 2,000,000-token mock context so `history-300` tests history handling rather than accidental compaction (`docs/performance-ci.md`, `scripts/perf/config.json`).

Recipe: measure tool arguments/results before sending; cap or page large outputs; compact before provider overflow; preserve a durable replacement-history event; set output/reasoning budgets per task rather than using the model maximum universally.

### Storage retry and durable progress

History delivery fails closed for replacement persistence but records bounded diagnostics and prevents repeated delivery after failure (`packages/agent-core/src/pi-turn-runner/history.ts`). The TUI update path retries Windows filesystem operations and verifies SHA-256 artifacts (`packages/tui/src/update/service.ts`). LLM retry is separately modeled and tested through `withLLMRetry` (`packages/agent-core/src/pi-turn-runner/llm-retry.ts`, `packages/agent-core/test/unit/pi-turn-runner/llm-retry.test.ts`), including rate limits, network failures, attempt counts, and usage observation.

The efficiency principle is not “retry everything”: retry transient provider/storage failures with bounded policy, preserve logical-turn identity, and make persistence failures visible. That avoids duplicate work and protects correctness while recovering from flaky infrastructure.

## Benchmark and performance evidence

The declared performance contract is in `docs/performance-ci.md` and `scripts/perf/config.json`:

| Scenario | Stress | What is checked |
|---|---:|---|
| `startup` | 1 tool round | startup/shutdown |
| `upstream-100` | 100 rounds × ~4 KiB | original upstream workload |
| `history-300` | 300 rounds × ~8 KiB | history beyond the 1,048,576-unit cache budget |

Both revisions are built separately, run serially on one runner, and compared with alternating paired samples. The suite measures duration, CPU, and sampled RSS; it rejects incomplete wire histories and incorrect tool results. Initial budgets are duration +25% and +1 s, CPU +20% and +0.5 core-seconds, RSS +20% and +32 MiB (`docs/performance-ci.md`, `scripts/perf/config.json`).

This is strong evidence that performance regressions are treated as correctness-plus-resource regressions, but it is not a published absolute speed number. No checked-in result here establishes a winner against another harness.

## Guidance recipes for other harnesses

1. Make the prompt/tool prefix stable and fingerprint it independently of append-only history.
2. Reuse one session identity when the provider supports prompt caching; record cache reads/writes.
3. Lazy-load provider integrations; snapshot immutable prompt assets once.
4. Stream model events and coalesce UI redraws to a frame budget.
5. Declare side-effect-free tools parallel and serialize mutations.
6. Bound tool output by bytes/tokens, page large files, and compact before overflow.
7. Persist history by cursor/delta; reserve full replacement for compaction.
8. Retry only classified transient failures with bounded attempts and stable logical identity.
9. Benchmark startup, ordinary history, and deliberately over-cache history with paired, noisy-run-resistant measurements.

## Summary and open questions

The likely efficiency advantage is compositional: cache-stable assembly, delta history persistence, lazy providers, streaming, safe tool concurrency, bounded context, and disciplined retries reduce overhead around the same model. The repository demonstrates mechanisms and benchmark methodology, not comparative product rankings.

Open questions: What are real provider cache-hit rates and cost deltas? How much wall-clock gain comes from parallel read/search calls? What are live-model quality and latency results against Claude Code, Codex, and Agy on identical tasks? What is the measured benefit of compaction versus larger-context replay?
