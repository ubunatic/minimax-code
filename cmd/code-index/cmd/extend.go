package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var extendCmd = &cobra.Command{
	Use:   "extend",
	Short: "Add new entries to the index",
	Long: `Guide for adding new entries to the repository index.

The index is stored in index.csv (at repository root) with these columns:
  type, path, primary_category, secondary_categories, summary

Where:
  type                  = "file" or "dir"
  path                  = relative path from repo root
  primary_category      = main responsibility (kebab-case)
  secondary_categories  = semicolon-separated list of related concerns; comma-space separated
  summary               = one-line description (quoted if contains special chars)

Steps to extend:
  1. Edit index.csv at repository root
  2. Add line: type,path,primary,"secondary;categories","Summary text"
  3. Rebuild: make install
  4. Verify: code-index -p "your-new-path"
  5. Commit: git add index.csv && git commit -m "docs: add entry to code index"

Example entries:
  dir,packages/new-pkg,feature-system,"capability-system","New feature system"
  file,packages/core/src/main.ts,agent-runtime,"api-exports","Core runtime exports"

Categories to use:
  - Run 'code-index -C' to see existing primary categories
  - Kebab-case for consistency
  - Secondary: split multiple concerns with semicolons`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("📝 Adding entries to the code index")
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println("\nEdit this file to add new entries:")
		fmt.Println("  index.csv (repository root)")
		fmt.Println("\nFormat (CSV with 5 columns):")
		fmt.Println("  type,path,primary_category,secondary_categories,summary")
		fmt.Println("\nExamples:")
		fmt.Println(`  dir,packages/new-pkg,feature-system,"system-capability","Feature package"`)
		fmt.Println(`  file,packages/core/src/main.ts,agent-runtime,"api-exports","Main runtime file"`)
		fmt.Println("\nAfter editing:")
		fmt.Println("  1. make install           # Rebuild the tool (embeds new entries)")
		fmt.Println("  2. code-index -p new-pkg  # Verify entry is found")
		fmt.Println("  3. git add index.csv      # Stage changes")
		fmt.Println("  4. git commit             # Commit with descriptive message")
		fmt.Println("\nTo list all existing categories:")
		fmt.Println("  code-index -C")
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(extendCmd)
}
