package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	pathFilter     string
	categoryFilter string
	searchFilter   string
	typeFilter     string
	listCategories bool
	listPaths      bool
	format         string
	wrap           bool
)

var rootCmd = &cobra.Command{
	Use:   "code-index",
	Short: "Search the minimax-code repository index",
	Long: `code-index: Query the minimax-code directory and file index

Search for files and directories by category, path, or summary text.
The index is embedded in the binary and works from anywhere.

Features:
- Fast file/directory lookup without grepping
- Fuzzy matching: "delegat" matches "delegation", "taks" matches "task"
- Supports partial words and typos via edit distance (Levenshtein)
- Includes both directories and key implementation files
- TTY-aware output formatting

Examples:
  # Find delegation/subagent code
  code-index -s "delegation"

  # Find all CLI/UI related files
  code-index -c "cli-ui"

  # Search with fuzzy matching (typos ok)
  code-index -s "permision"  # matches "permission"

  # List all categories with counts
  code-index -C

  # Show git status and index statistics
  code-index status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the index
		entries, err := LoadIndex()
		if err != nil {
			return fmt.Errorf("error loading index: %w", err)
		}

		// Handle category listing
		if listCategories {
			categories := ListCategories(entries)
			PrintCategories(categories)
			return nil
		}

		// Handle path listing
		if listPaths {
			PrintPaths(entries)
			return nil
		}

		// Filter entries
		results := FilterEntries(entries, pathFilter, categoryFilter, searchFilter, typeFilter)

		if len(results) == 0 {
			fmt.Printf("0 matches against %d indexed entries\n", len(entries))
			return nil
		}

		PrintResults(results, format, wrap)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&pathFilter, "path", "p", "", "Filter by path (substring match)")
	rootCmd.Flags().StringVarP(&categoryFilter, "category", "c", "", "Filter by category (primary or secondary)")
	rootCmd.Flags().StringVarP(&searchFilter, "search", "s", "", "Search in summary text (case-insensitive)")
	rootCmd.Flags().StringVarP(&typeFilter, "type", "t", "", "Filter by type: 'file' or 'dir'")
	rootCmd.Flags().BoolVarP(&listCategories, "categories", "C", false, "List all unique categories with entry counts")
	rootCmd.Flags().BoolVarP(&listPaths, "paths", "P", false, "List all paths (files and directories)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "flat", "Output format: 'flat' (default, tab-separated) or 'table'")
	rootCmd.Flags().BoolVar(&wrap, "wrap", false, "Wrap long lines (auto-enabled for TTY)")
}
