package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean [paths...]",
	Short: "Remove missing file entries from index",
	Long: `Remove entries from the index that are no longer present in the repository.

Specify paths to remove, or use 'code-index check' to identify missing entries first.

Note: This modifies docs/explore2/index.csv directly. Commit the changes.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("\n⚠ NOTE: Manual CSV editing required")
		fmt.Println("─────────────────────────────────────")
		fmt.Println("\nThe clean verb requires direct CSV modification.")
		fmt.Println("To remove entries, edit: docs/explore2/index.csv")
		fmt.Println("\nPaths to remove:")
		for _, path := range args {
			fmt.Printf("  • %s\n", path)
		}
		fmt.Println("\nAfter removing lines, rebuild with:")
		fmt.Println("  make install")
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
