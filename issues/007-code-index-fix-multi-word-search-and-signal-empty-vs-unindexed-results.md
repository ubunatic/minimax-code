# 007 — code-index: fix multi-word search and signal empty-vs-unindexed results

**Status**: Open
**Priority**: P3
**Severity**: Minor
**Category**: Tooling

---

## Goal

Two small, independently-fixable follow-ups from issue #006, observed while
using the fuzzy/file-level search added in df79f27 across three live
exploration tasks (see `docs/studies/CodeIndexEffectiveness.md`). Neither
requires the bigger content/symbol-indexing work #006 already tracks.

## Background

1. **Multi-word/phrase queries underperform single-word queries against the
   same content.** `-s "tool call"`, `-s "tool use"`, `-s "tool result"` all
   returned no matches, while `-s "tool"` alone returned nine relevant hits
   including the same files a phrase query should have matched. Similarly in
   an earlier task, `-s "agent delegation"`-style phrases missed where single
   words hit. It's unclear whether phrase terms are ANDed too strictly,
   tokenized unexpectedly, or something else — but the practical effect is
   that natural-language queries (the mode an agent defaults to) perform
   worse than keyword queries against the identical underlying index.

2. **"No matches" is indistinguishable from "not indexed at all."** When a
   search returns nothing, an agent can't tell whether to keep trying
   synonyms (the concept might be indexed under different words) or give up
   and fall back to `ls`/`grep` immediately (the concept might not be
   indexed at any granularity). This ambiguity caused repeated synonym
   retries in two separate exploration tasks before falling back to manual
   directory listing, which is where the actually useful files were found
   both times.

## Acceptance criteria

- Multi-word `-s` queries return matches at least as good as the best
  single-word subset of that query, or the tool documents why phrase
  matching is intentionally stricter.
- A "no matches" result includes a lightweight signal distinguishing
  "searched but nothing matched" from useful context like total indexed
  entry count, so an agent can decide whether to retry with different
  vocabulary or fall back to `ls`/`grep` immediately — e.g. "0 matches
  against N indexed entries; try a directory listing instead."

## Notes

Not urgent — code-index remains a net positive as first-pass triage even
with these gaps (see the study doc for the full before/after assessment).
Complements #006 rather than duplicating it; #006 covers indexing content/
symbols at finer granularity, this covers search-quality and UX issues in
the existing summary-based search.
