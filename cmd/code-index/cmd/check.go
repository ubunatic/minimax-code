package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	checkRepoRoot string
)

var checkCmd = &cobra.Command{
	Use:   "check [repo-root]",
	Short: "Check for missing files in index",
	Long: `Verify that all indexed paths exist in the repository.

Reports:
- Missing files/directories (indexed but not found)
- Suggestions to run 'clean' to remove stale entries`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := LoadIndex()
		if err != nil {
			return err
		}

		// Get repo root from argument or flag
		root := checkRepoRoot
		if len(args) > 0 {
			root = args[0]
		}
		if root == "" {
			root = "."
		}

		var missing []Entry
		missing_count := 0

		for _, e := range entries {
			fullPath := filepath.Join(root, e.Path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				missing = append(missing, e)
				missing_count++
			}
		}

		if missing_count == 0 {
			fmt.Println("✓ All indexed paths exist")
			return nil
		}

		fmt.Printf("\n⚠ Found %d missing entries:\n", missing_count)
		fmt.Println("─────────────────────────────────────────────────────")
		for _, e := range missing {
			fmt.Printf("%s  %s\n", e.Type[0:1], e.Path)
		}

		fmt.Println("\nRemove stale entries with:")
		fmt.Println("  code-index clean [missing-paths...]")
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVarP(&checkRepoRoot, "root", "r", ".", "Repository root directory")
}
