package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show index summary and statistics",
	Long: `Display statistics about the index:
- Total entries (files and directories)
- Number of categories
- Files vs directories count
- Breakdown by primary category`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := LoadIndex()
		if err != nil {
			return err
		}

		// Count entries
		dirs := 0
		files := 0
		for _, e := range entries {
			if e.Type == "dir" {
				dirs++
			} else {
				files++
			}
		}

		// Get categories
		categories := ListCategories(entries)

		fmt.Println("\nIndex Status")
		fmt.Println("============")
		fmt.Printf("Total entries:    %d\n", len(entries))
		fmt.Printf("Directories:      %d\n", dirs)
		fmt.Printf("Files:            %d\n", files)
		fmt.Printf("Total categories: %d\n", len(categories))

		// Count by primary category
		primaryCounts := make(map[string]int)
		for _, e := range entries {
			primaryCounts[e.Primary]++
		}

		fmt.Println("\nTop 10 Primary Categories:")
		fmt.Println("──────────────────────────")

		type pair struct {
			name  string
			count int
		}
		var sorted []pair
		for k, v := range primaryCounts {
			sorted = append(sorted, pair{k, v})
		}

		// Simple sort by count descending
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].count > sorted[i].count {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		limit := 10
		if len(sorted) < limit {
			limit = len(sorted)
		}

		for i := 0; i < limit; i++ {
			fmt.Printf("  %-30s : %2d entries\n", sorted[i].name, sorted[i].count)
		}

		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
