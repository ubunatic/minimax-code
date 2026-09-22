# Explore 2: Index and Query Tools

This directory contains tools for navigating and understanding the minimax-code repository structure.

## Files

### `CodeMap.md`
Comprehensive narrative guide to the repository. Start here for understanding:
- Overall structure and layout
- Purpose of each major package
- Where to find code for specific features
- Critical paths for common tasks
- Key concepts and terminology

**Best for**: First exploration, understanding relationships, learning context.

### `index.csv`
Structured database of all directories and significant files with:
- Path to the resource
- Primary category
- Secondary categories  
- Summary description

**Columns**:
- `type` — "file" or "dir"
- `path` — Full path to the resource
- `primary_category` — Main responsibility/concern
- `secondary_categories` — Related concerns (comma-separated)
- `summary` — One-line description

**Best for**: Programmatic queries, building your own lookup tables, importing into tools.

### `query.go`
Go command-line tool to query `index.csv` with various filters.

**Best for**: Fast lookups while exploring.

## Using the Query Tool

### Install globally
```bash
# Install to /usr/local/bin/code-index (default)
make install

# Or to a custom location
make install PREFIX=$HOME/.local INSTALL_PATH=$HOME/.local/bin

# Then use from anywhere
code-index -help
```

### Build locally
```bash
cd docs/explore2/
go build -o query query.go
./query -help
```

**Note**: The CSV is embedded in the binary, so it works everywhere with no file dependencies.

### Query Examples

After installing with `make install`, you can run these from anywhere:

**Find all CLI/UI related files and directories:**
```bash
code-index -category "cli-ui"
```

**Find all directories related to authentication:**
```bash
code-index -type "dir" -category "authentication"
```

**Find all things related to session management:**
```bash
code-index -category "session-management"
```

**Search for "testing" in summaries:**
```bash
code-index -search "testing"
```

**Find everything in packages/local-runtime:**
```bash
code-index -path "packages/local-runtime"
```

**List all unique categories:**
```bash
code-index -categories
```

**Show all paths:**
```bash
code-index -paths
```

**Find runtime-host-v2 related files:**
```bash
code-index -primary "runtime-host-v2"
```

## Quick Lookups

### "Where do I find code for..."

**...the agent execution engine?**
```bash
code-index -category "agent-runtime"
```
→ `packages/agent-core/` (primary), `packages/agent-extension/`, `packages/agent-runtime/`

**...CLI/TUI features?**
```bash
code-index -category "cli-ui"
```
→ `packages/tui/` (primary)

**...the local runtime?**
```bash
code-index -search "runtime"
```
→ `packages/local-runtime/` (v1), `packages/local-runtime-v2/` (v2)

**...permission and access control?**
```bash
code-index -category "permissions"
```
→ `packages/agent-modules/permission/`

**...authentication and OAuth?**
```bash
code-index -category "authentication"
```
→ `packages/oauth-core/`, `packages/oauth-lease-protocol/`, OAuth-related code in TUI

**...session management?**
```bash
code-index -category "session-management"
```
→ `packages/local-runtime/src/session/`, `packages/local-runtime-v2/src/application/session/`

**...model selection and catalogs?**
```bash
code-index -category "model-system"
```
→ `packages/local-runtime-v2/src/service/model-system/`

**...MCP integration?**
```bash
code-index -category "mcp-integration"
```
→ `packages/agent-modules/mcp/`, `packages/local-runtime-v2/src/service/mcp/`

**...tests and testing infrastructure?**
```bash
code-index -category "testing"
```
→ `test/`, `vitest-suites.json`, all `*/test/` directories

**...build and release processes?**
```bash
code-index -category "build-system"
code-index -category "release-tools"
```
→ `scripts/`, `release/`, build configuration files

## Category Glossary

### Main Categories (ordered by importance/coverage)

- **agent-runtime** — Core agent execution engine and infrastructure
- **runtime-host** — Local runtime implementations (v1 and v2)
- **cli-ui** — Terminal UI and command-line interface
- **testing** — Test infrastructure and test files
- **documentation** — User guides and process docs
- **build-system** — Build scripts and configuration
- **release-tools** — Release, packaging, and publishing
- **protocol** — Protocol definitions and contracts
- **authentication** — OAuth, credentials, and access control
- **session-management** — Session lifecycle and state
- **model-selection** — Model catalog and selection logic
- **permissions** — Permission system and policy
- **mcp-integration** — Model Context Protocol integration
- **capability-system** — Pluggable agent capabilities

### Other Categories

- **configuration** — Configuration management
- **vendor-code** — Vendored upstream packages
- **examples** — Example code and demonstrations
- **issue-tracking** — Issue tracker (harnez-managed)

## Programmatic Usage

You can also use the CSV directly in scripts:

```bash
# Get all CLI-related paths
grep "cli-ui" docs/explore2/index.csv | cut -d',' -f2

# Get all directory paths
grep "^dir," docs/explore2/index.csv | cut -d',' -f2

# Count entries by type
awk -F',' '{print $1}' docs/explore2/index.csv | sort | uniq -c
```

Or import the CSV into Python/Node/etc:
```python
import csv
with open('docs/explore2/index.csv') as f:
    reader = csv.DictReader(f)
    for row in reader:
        if 'cli-ui' in row['secondary_categories']:
            print(row['path'])
```

## Related Documents

- **CodeMap.md** — Narrative guide with relationships and context
- **../../CLAUDE.md** — Project conventions and agent guidance
- **../../CONTRIBUTING.md** — Contribution guidelines
