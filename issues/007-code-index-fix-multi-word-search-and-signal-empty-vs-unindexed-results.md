# 007 — code-index: fix multi-word search and signal empty-vs-unindexed results

**Status**: Closed — Resolved
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

## Implementation

**Status**: Completed and tested

**Changes made**:
1. Added `scoreMultiWord()` function to handle multi-word queries using OR semantics
   - Splits query into individual words
   - Returns best match score from ANY query word (not all required)
   - Single-word queries delegate to original `scoreMatch()` for backward compatibility

2. Updated `FilterEntries()` to use `scoreMultiWord()` for search filtering
   - Maintains all existing filters (path, category, type)
   - Preserves fuzzy matching and typo tolerance

3. Updated empty result message in root.go
   - Changed "No matches found" to "0 matches against N indexed entries"
   - N is the total count of indexed entries (currently 78)

**Test results**:
- Single-word "tool": 8 matches (baseline unchanged)
- Multi-word "tool call": 11 matches (was 0, now works with OR semantics)
- Multi-word "agent delegation": 21 matches including delegation-specific files
- Empty search "impossible": Shows "0 matches against 78 indexed entries"
- Fuzzy matching "delegat": Still matches "delegation" correctly
- All existing filters continue to work

**Test coverage** (commit bd78899):
- 14 unit/integration tests added via index_test.go
- TestScoreMatch: 7 cases covering exact/prefix/substring matching, fuzzy typos, no-match
- TestScoreMultiWord: 10 cases covering single-word, multi-word, edge cases, typos, stopwords
- TestFilterEntriesMultiWord: 4 integration cases with real Entry data
- All tests passing

**Known behavior documented**:
- OR semantics with substring matching allows short words to match substrings
- Example: "the" matches "gather" via substring (precision trade-off for recall)
- This is expected behavior within acceptance criteria

**Addresses**:
- ✅ Criterion 1: Multi-word queries return matches at least as good as best single-word subset
- ✅ Criterion 2: Empty results signal indexed entry count for context
- ✅ Code review feedback: Added unit tests covering all scoring functions and edge cases
