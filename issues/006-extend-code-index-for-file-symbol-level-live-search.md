# 006 — Extend code-index for file/symbol-level, live search

**Status**: Open
**Priority**: P2
**Severity**: Minor
**Category**: Tooling

---

## Goal

Make `code-index` (`cmd/code-index`) a genuinely token-efficient first pass for
agent exploration, not just a directory map. An agent should be able to land
on the *specific file* (and ideally symbol) relevant to a query, and trust
`status`/`untracked` to reflect the live working tree, without falling back to
`grep`/`find`/full-file `Read` for the second half of every search.

## Background

Used during a live exploration task (find "agent delegation" code and write it
up). `code-index --search agent`/`task` correctly surfaced the right
top-level directories (`AGENTS.md`, `packages/agent-core`,
`packages/agent-modules/background-task`, `packages/agent-tools`) in one cheap
call — a real win over blind grepping. But:

- The index only has ~67 entries, almost all directories, each with a
  one-line summary. It has no per-file or per-symbol entries, so once the
  right directory was found, locating the actual implementation file
  (`subagent-roles.ts`, `canonical-tool-policy.ts`,
  `background-task/src/types.ts`, etc.) still required manual `grep -rl`/
  `find`, followed by full `Read`s of large files (one was ~1000 lines) to
  find the ~150 relevant lines.
- Second-round searches for close synonyms (`delegat`, `subagent`, `spawn`)
  returned no matches even though the concept is all over the actual code,
  because only directory-summary text is indexed, not file contents or
  symbol names.
- `code-index status` / `code-index untracked` reported a stale view: it said
  "no untracked directories/files found" while `git status` showed several
  untracked paths (`cmd/`, `docs/explore2/`, `index.csv`) in the same
  working tree. The index isn't checked/refreshed against live git state.
- Minor CLI friction: the short flag form (`-s <query>`) is parsed as a
  subcommand rather than the `--search` flag, producing a confusing
  "unknown command" error instead of running the search.

## Acceptance criteria

- `code-index` can index and search at file granularity (not only
  directories), with a short per-file summary, for at least the primary
  first-party source trees.
- Search matches file/symbol content or richer indexed keywords, not only the
  literal words already present in the stored summary text, so reasonable
  synonyms for an indexed concept still surface the right file.
- `status`/`untracked` are accurate against the current working tree (or the
  tool clearly documents that its index is a point-in-time snapshot and
  states its age, so an agent knows to cross-check `git status`).
- The `-s`/`-search`/`--search` short-flag ambiguity is fixed or documented so
  it fails obviously rather than silently mis-parsing as a subcommand.
- (Stretch) A "list symbols in path" or similar command lets an agent get an
  exported-function/class summary for a specific file without a full `Read`,
  for large files where only one function/class is relevant.

## Notes

Not urgent — `code-index` already provides real value as a cheap first pass
for directory-level orientation. This ticket is about closing the gap between
"found the right neighborhood" and "found the right file," which is where
token cost was still being spent during the observed session.
