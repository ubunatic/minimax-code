# code-index — Repository Index CLI

A Cobra CLI tool for querying the minimax-code repository structure and file index.

## Building

```bash
cd cmd/code-index
go build -o code-index ./main.go
./code-index --help
```

## Installation

From the repository root:

```bash
# Install to /usr/local/bin (default)
make install

# Install to custom location
make install PREFIX=$HOME/.local INSTALL_PATH=$HOME/.local/bin

# Verify installation
code-index --help
```

## Project Structure

```
cmd/code-index/
├── main.go              # Entry point
├── cmd/
│   ├── root.go          # Root command with all flags
│   ├── index.go         # Index data structure and query logic
│   ├── output.go        # Output formatting
│   └── data.csv         # Embedded repository index
├── go.mod               # Go module definition
├── go.sum               # Go module checksums (generated)
└── code-index          # Compiled binary (generated)
```

## Features

- **Embedded data** — CSV index compiled into binary
- **No dependencies at runtime** — Works from anywhere
- **Cobra framework** — Standard Go CLI structure
- **Multiple query types** — Category, path, text search, type filter

## CLI Reference

```
code-index [flags]

Flags:
  -c, --category string   Filter by category (primary or secondary)
  -p, --path string       Filter by path (substring match)
  -s, --search string     Search in summary text (case-insensitive)
  -t, --type string       Filter by type: 'file' or 'dir'
  -C, --categories        List all unique categories with counts
  -P, --paths             List all paths
  -h, --help              Show help message
```

## Usage Examples

```bash
# Find all CLI/UI files
code-index -c cli-ui

# Find directories related to authentication
code-index -t dir -c authentication

# Search for session-related code
code-index -s "session"

# List all categories
code-index -C

# Combine filters
code-index -p "packages" -t dir -s "runtime"
```

## Data Source

The index data (`cmd/code-index/cmd/data.csv`) is:
- Generated from the reference at `docs/explore2/index.csv`
- Embedded at compile time via `//go:embed data.csv`
- Contains 60+ categories across 65 files/directories

## Maintaining the Index

When the repository structure changes:

1. Update `docs/explore2/index.csv`
2. Copy to `cmd/code-index/cmd/data.csv`:
   ```bash
   cp docs/explore2/index.csv cmd/code-index/cmd/data.csv
   ```
3. Rebuild and reinstall:
   ```bash
   make install
   ```

## Development

The tool is organized into logical modules:

- **index.go** — Data loading and filtering logic
- **output.go** — Result formatting and presentation
- **root.go** — Cobra command definition and flags
- **main.go** — Simple entry point

To add a new command (e.g., `code-index search`), create a new file in `cmd/` and add it to the root command.
