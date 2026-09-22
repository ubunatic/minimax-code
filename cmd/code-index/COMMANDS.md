# code-index Commands Reference

## Overview

`code-index` is a Cobra CLI for managing and querying the minimax-code repository index. It includes both query commands (flags) and management subcommands.

## Query Commands (Flags)

Use these flags on the root command to search the index:

### `-c, --category STRING`
Filter by category (primary or secondary).

```bash
code-index -c cli-ui
code-index -c authentication
code-index -c "session-management"
```

### `-p, --path STRING`
Filter by path substring match.

```bash
code-index -p packages/tui
code-index -p "agent-modules"
```

### `-s, --search STRING`
Search in summary text (case-insensitive).

```bash
code-index -s "runtime"
code-index -s "protocol"
```

### `-t, --type STRING`
Filter by type: `file` or `dir`.

```bash
code-index -t dir -c authentication
code-index -t file -p scripts
```

### `-C, --categories`
List all unique categories with entry counts.

```bash
code-index -C
code-index -C | grep -i test
```

### `-P, --paths`
List all paths (files and directories).

```bash
code-index -P
```

## Management Subcommands

### `status`
Show index summary and statistics.

Displays:
- Total entries (files and directories)
- Count breakdown
- Top 10 primary categories

```bash
code-index status
```

### `check [repo-root]`
Verify that all indexed paths exist in the repository.

Reports missing files/directories that are in the index but not on disk.

```bash
code-index check .
code-index check /path/to/repo
```

### `clean [paths...]`
Remove missing file entries from the index.

This command guides you to manually edit `docs/explore2/index.csv` and provides instructions for rebuilding.

```bash
code-index clean packages/old-package
code-index clean NOTICE pnpm-lock.yaml
```

### `extend`
Guide for adding new entries to the index.

Explains the CSV format and step-by-step process for extending the index.

```bash
code-index extend
```

### `untracked [root]`
Find untracked top-level directories and files.

Scans the repository for significant entries not in the index. Useful for discovering new packages or documentation areas.

```bash
code-index untracked .
code-index untracked /path/to/repo
code-index untracked . --ignore "vendor,old,legacy"
```

Flags:
- `--root, -r` — Repository root directory
- `--ignore` — Comma-separated ignore patterns

Default ignores: .git, .github, node_modules, dist, build, __pycache__, etc.

## Workflow Examples

### Finding code by concern
```bash
# All authentication-related files
code-index -c authentication

# All testing infrastructure
code-index -C | grep test

# All directories in packages
code-index -t dir -p packages
```

### Maintaining the index

**After code changes, verify the index:**
```bash
code-index check .          # Find missing entries
code-index untracked .      # Find new entries to add
```

**When you add a new directory:**
```bash
code-index extend           # Get formatting help
# (edit docs/explore2/index.csv)
make install                # Rebuild tool
code-index -p "new-pkg"     # Verify it was added
```

**When you remove or rename something:**
```bash
code-index check .          # Find stale entries
code-index clean old-path   # Get instructions for removal
```

## Combining Filters

You can use multiple flags together:

```bash
# Directories in packages related to authentication
code-index -t dir -p packages -c authentication

# Files about runtime
code-index -t file -s "runtime"

# Everything in scripts except test files
code-index -p scripts -s "build"
```

## CSV Format

The index is stored in `docs/explore2/index.csv`:

```
type,path,primary_category,secondary_categories,summary
dir,packages/tui,cli-ui,"authentication; runtime-adapters",Terminal UI and CLI
file,README.md,documentation,"quick-start; product-overview",Product overview
```

**Columns:**
- `type` — "file" or "dir"
- `path` — Relative path from repo root
- `primary_category` — Main responsibility (kebab-case)
- `secondary_categories` — Semicolon-separated related concerns
- `summary` — One-line description (quoted if contains comma)

## Performance

The index is embedded in the binary at compile time, so:
- ✓ No file I/O at runtime
- ✓ Works from anywhere
- ✓ Fast queries
- ✓ Single distributable binary
