# code-index Agent Guide

Fast repository navigation without grepping. Find implementation files, types, and entry points in ~100ms.

## Quick Start

```bash
# Search for concepts
code-index -s "delegation"          # Find subagent delegation code
code-index -s "permision"           # Fuzzy match: catches typos

# Browse by category
code-index -c "task-management"     # All background-task files
code-index -C                       # List all categories

# Filter by type/path
code-index -p "local-runtime" -t "dir"  # Find specific packages
code-index -P                       # List all indexed paths

# Check status
code-index status                   # Index stats + git state
code-index untracked --git          # Show git-untracked items
```

## When to Use code-index

| Goal | Command | Output |
|------|---------|--------|
| Find implementation file for concept | `code-index -s "concept"` | File path + category |
| Explore all files in a subsystem | `code-index -c "category"` | All related files |
| Verify indexed coverage | `code-index status` | Total entries, gaps, git status |
| Find entry point for a package | `code-index -p "package"` | Directory + main files |
| Check git state | `code-index status` | Modified/staged/untracked counts |

## Search Features

### Exact & Substring Matches
```bash
code-index -s "task"        # Finds "task-management" files
code-index -s "agent-core"  # Finds "packages/agent-core"
```

### Fuzzy Matching (Typos & Partial Words)
```bash
code-index -s "delegat"     # → "delegation" (substring)
code-index -s "permision"   # → "permission" (1 edit distance)
code-index -s "taks"        # → "task" (2 edit distance)
```

**Fuzzy rules:**
- Patterns ≤3 chars: exact/substring only
- Patterns 4+ chars: allows up to 2 character edits (typos, transpositions)
- Checks individual words in summaries for better precision

### Category Filtering
```bash
code-index -C                           # Show all categories
code-index -c "task-management"         # Find by primary category
code-index -c "permissions"             # Works with secondary too
```

**Current categories** (top 10):
- `build-system` (13 entries)
- `release-tools` (7 entries)
- `task-management` (4 entries)
- `documentation` (4 entries)
- `permissions` (3 entries)
- `session-management` (3 entries)
- `delegation` (1 entry)

### Path Filtering
```bash
code-index -p "agent-core"              # Find by path substring
code-index -p "local-runtime" -t "dir"  # Combine with type filter
code-index -P                           # List all indexed paths
```

## Output Formats

### Default (Tab-separated, auto-wraps in TTY)
```
type	path	primary_category	secondary_categories	summary
file	packages/agent-tools/src/desktop/subagent-roles.ts	delegation	subagent-authorization	Subagent role definitions and delegation rules
```

### Flat format (for piping/scripting)
```bash
code-index -s "delegation" -f flat | cut -f2  # Extract just paths
```

### With wrapping
```bash
code-index -s "delegation" --wrap  # Force line wrapping in terminal
```

## Real-World Scenarios

### Scenario 1: Find Agent Delegation Code
```bash
$ code-index -s "delegation"
# → packages/agent-tools/src/desktop/subagent-roles.ts

$ code-index -s "subagent"
# → Same file + AGENTS.md + agent-skills directory

$ code-index -c "delegation"
# → All delegation-related entries
```

**Why it works:** Key implementation files are indexed by name + summary, allowing agents to land directly on code without manual grepping.

### Scenario 2: Explore Background Tasks
```bash
$ code-index -c "task-management"
# Returns:
#   packages/agent-modules/background-task/ (dir)
#   .../background-task/src/index.ts (exports)
#   .../background-task/src/manager.ts (implementation)
#   .../background-task/src/types.ts (type definitions)
```

**Why it works:** Hierarchical indexing—directory summary + key files + type/interface files. No need to `ls` or `find`.

### Scenario 3: Verify Repository State
```bash
$ code-index status
# Index Status: 77 entries (46 files, 31 dirs), 113 categories
# Git Status: 0 staged, 0 modified, 0 untracked

$ code-index untracked --git
# Check what's actually untracked in git (vs just missing from index)
```

**Why it works:** Combines index coverage with live git state for quick sanity checks.

## Index Coverage

**What's indexed:**
- All top-level directories in `packages/`
- Key implementation files: `subagent-roles.ts`, `canonical-tool-policy.ts`, task manager files
- Type/interface files: `types.ts` for critical modules
- Main entry points: `index.ts` for each package
- Build/release files: Makefile, scripts, release contracts
- Documentation: README, AGENTS.md, CLAUDE.md

**What's NOT indexed:**
- Individual test files (use `test/vitest-suites.json` instead)
- Auto-generated files (tsconfig paths, etc.)
- Vendored third-party code (in `third_party/`)

**To expand coverage:**
```bash
code-index extend      # Shows format and workflow
# Edit: index.csv at repo root
# Rebuild: make install
# Verify: code-index -p "your-new-path"
```

## Adding New Categories

Discover missing coverage, then add entries:

```bash
# 1. Find gaps
code-index untracked              # Files not in index
code-index status | head -5       # Low-coverage areas

# 2. Add to index.csv (repo root)
echo 'file,packages/foo/src/bar.ts,new-category,"related-concept","Brief summary"' >> index.csv

# 3. Rebuild & verify
make install
code-index -c "new-category"      # See new entries
code-index -C | grep new-category # Check count

# 4. Commit
git add index.csv
git commit -m "docs: add new-category entries"
```

**Format:**
- `primary_category` — Main responsibility (kebab-case, single word)
- `secondary_categories` — Semicolon-separated related concerns
- `summary` — One line, quoted if contains commas

## Limitations & Workarounds

| Issue | Workaround |
|-------|-----------|
| "spawn" doesn't match agent delegation | Use `"delegation"` or `"subagent"` instead |
| Need to find a specific function inside a large file | Open the file (returned by code-index) and search within it; symbol indexing deferred |
| Index is ~2 minutes old | Use `code-index status` to check; run `make install` to rebuild from latest `index.csv` |

## Commands Reference

```bash
code-index [flags] [command]

Flags:
  -s, --search STRING      Search in summaries (fuzzy + substring match)
  -c, --category STRING    Filter by primary/secondary category
  -p, --path STRING        Filter by path substring
  -t, --type STRING        Filter by type: "file" or "dir"
  -C, --categories         List all categories with counts
  -P, --paths              List all indexed paths
  -f, --format STRING      Output format: "flat" (tab-sep) or "table"
  --wrap                   Wrap long lines

Commands:
  status                   Show index stats + git state
  untracked [--git]        Find files/dirs not in index (or git-untracked with --git)
  check                    Verify all indexed paths exist
  extend                   Guide for adding new entries
  clean                    Remove stale entries from index
```

## Performance Notes

- **Search:** ~5ms (embedded CSV, no network)
- **Works from anywhere:** Binary includes full index
- **Zero dependencies:** No node, no go.mod fetch
- **TTY detection:** Auto-formats for terminal vs pipe

## See Also

- `index.csv` — Source of truth (77 entries, 113 categories)
- `docs/explore2/CodeMap.md` — Narrative repository guide
- `AGENTS.md` — Agent conventions and boundaries
