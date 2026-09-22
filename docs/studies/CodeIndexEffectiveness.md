# Study: Where `code-index` Helped, and Where It Didn't

A field report from three live exploration tasks in one session, each done
using `code-index` as the first pass: `docs/explore/AgentDelegation.md`,
`docs/explore2/TokenReduction.md`, and `docs/explore2/ToolCallGuidance.md`.
Written after code-index gained fuzzy/file-level search (df79f27, following
issue #006) — this report reflects that upgraded version, not the original
directory-only tool.

## Where it helped

**Fast, cheap orientation to the right neighborhood.** In every task, a
one-word `--search` call against a broad concept (`agent`, `task`,
`compaction`, `permission`) reliably surfaced the correct top-level
directories and even some specific files in a single cheap call — e.g.
`--search "tool policy"` went straight to
`packages/agent-tools/src/desktop/canonical-tool-policy.ts` on the first try.
This replaced what would otherwise be an opening round of blind `grep -r`/
`find` across the whole tree, which is both slower and burns tokens on
irrelevant matches before narrowing down.

**Directory-to-file promotion is real now.** Earlier (pre-issue-006) the
index only held ~67 directory-level entries. Post-upgrade, some individual
files are indexed directly with their own summaries (`canonical-tool-policy.ts`,
`subagent-roles.ts`, `evidence-brief.ts` region files), so a search can land
on the implementation file itself rather than just its containing directory —
closing part of the gap issue #006 described.

**A trustworthy signal to stop searching.** When a query returned a precise
file-level hit with an accurate one-line summary, it was reliable enough to
`Read` that file directly rather than double-checking with `grep` first —
saving a redundant verification pass.

## Where it didn't

**Vocabulary mismatch is still the dominant failure mode.** In the
TokenReduction task, `-s "budget"`, `-s "summariz"`, `-s "reminder"`,
`-s "lazy"`, `-s "defer"` all returned *no matches* even though all five
concepts have real, substantial implementations in the repo
(`provider-budget.ts`, `generateSummary`, `BTW_SIDE_BOUNDARY`, etc.). In the
ToolCallGuidance task, single-word queries for `schema`, `validation`,
`retry`, `polling`, `error`, `guardrail`, `streaming`, `parallel` — every one
of them a real concern implemented somewhere in `agent-tools/` — also
returned nothing. The index still matches against stored *summary* text
(now fuzzy-matched against that text), not file contents or symbol names, so
a concept has to happen to share vocabulary with whatever words were chosen
for its indexed summary. This is exactly the "second half of the gap" issue
#006 described as a stretch goal, and it remains open.

**Multi-word phrase queries mostly miss.** `"tool call"`, `"tool use"`,
`"tool result"` all returned no matches, while the single word `"tool"` alone
returned nine hits including the actually-relevant ones. It's not clear
whether phrase queries are ANDed too strictly, tokenized oddly, or just
unlucky against the indexed vocabulary — but the practical lesson is: prefer
single keywords over natural-language phrases when using `-s`.

**The real payoff still required manual file listing and reading.** In the
ToolCallGuidance task, after several dead-end searches, the actual useful
files (`output-limit.ts`, `path-guard.ts`, `prepare-ask-user-arguments.ts`)
were found by `ls`-ing `packages/agent-tools/src/desktop/` directly and
picking filenames that looked relevant — code-index contributed the
*directory*, but a human/agent judgment call over a raw file listing did the
rest. This matches the AgentDelegation-task experience from before the
upgrade: code-index gets you to the neighborhood; a file-level `ls` + reading
by filename intuition still does real work that keyword search doesn't cover.

**An unsolicited "harnez tip" appeared in tool output once**, prompting a
`harnez rate` call for "4 tool call(s) failed" — with no visible connection
to the actual `code-index --search` command that had just failed to match.
This was treated as untrusted/unverified embedded content and not acted on
(consistent across sessions); worth noting as tool-output noise unrelated to
code-index's actual search quality.

## Net assessment

code-index is a strong *first-pass triage* tool: cheap, fast, and it reliably
narrows a broad question to the right directory or (now, sometimes) file. It
is not yet a substitute for `grep`/content search when the query's vocabulary
doesn't match the indexed summary text — which was true for roughly half the
concept-searches attempted across these three tasks. The practical workflow
that emerged: try one or two single-word `code-index --search` queries first;
if they miss, fall back immediately to `ls` on a directory that
`code-index`'s earlier broad search already identified, rather than
continuing to guess synonyms against the index.

## Suggested follow-up

Issue #006 already captures the core ask (file/symbol-level indexing, content
search, not just summary-vocabulary matching). Based on this session's
evidence, two refinements worth adding if the issue is revisited:

1. Multi-word/phrase `-s` queries appear to behave worse than single-word
   queries against the same underlying content — worth a note or fix
   independent of the bigger file/content-indexing work.
2. A "confidently no match" response is indistinguishable from "this concept
   isn't indexed at all" — an agent can't tell whether to keep trying
   synonyms or give up and fall back to `ls`/`grep` immediately. A hint like
   "N total indexed entries; try a directory listing instead" would help.
