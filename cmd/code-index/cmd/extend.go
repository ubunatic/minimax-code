package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var extendCmd = &cobra.Command{
	Use:   "extend",
	Short: "Add new entries to the index",
	Long: `Guide for adding new entries to the repository index.

The index is stored in docs/explore2/index.csv with these columns:
  type, path, primary_category, secondary_categories, summary

Where:
  type                  = "file" or "dir"
  path                  = relative path from repo root
  primary_category      = main responsibility (kebab-case)
  secondary_categories  = semicolon-separated list of related concerns
  summary               = one-line description (quoted if contains comma)

Steps to extend:
  1. Edit docs/explore2/index.csv directly
  2. Add line: type,path,primary,secondary;list,"summary text"
  3. Rebuild: make install
  4. Verify: code-index -p "your-new-path"
  5. Commit: git add docs/explore2/index.csv && git commit ...

Example entry:
  dir,packages/new-feature,feature-system,capability-system,"New feature system"

Categories to use:
  - Existing primary categories (run 'code-index -C' to see)
  - Kebab-case for consistency
  - Secondary: split multiple concerns with semicolon`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("📝 Adding entries to the index")
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println("\nEdit this file to add new entries:")
		fmt.Println("  index.csv (at repo root)")
		fmt.Println("\nFormat (CSV with 5 columns):")
		fmt.Println("  type,path,primary_category,secondary_categories,summary")
		fmt.Println("\nExample:")
		fmt.Println(`  dir,packages/new-pkg,feature-system,"system-capability","Description"`)
		fmt.Println("\nAfter editing:")
		fmt.Println("  1. make install              # Rebuild the tool")
		fmt.Println("  2. code-index -p new-pkg     # Verify entry")
		fmt.Println("  3. git add docs/explore2/    # Stage changes")
		fmt.Println("  4. git commit -m 'docs: ...' # Commit")
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(extendCmd)
}
