package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	untrackedRoot      string
	untrackedIgnore    string
	untrackedGitStatus bool
)

var untrackedCmd = &cobra.Command{
	Use:   "untracked [root]",
	Short: "Find untracked top-level directories and files",
	Long: `Scan the repository for significant directories and files that are not in the index.

Useful for discovering new packages, feature areas, or documentation that should be indexed.

Ignores:
  - Hidden directories (dot-files)
  - Common build/cache directories (node_modules, .git, dist, etc.)
  - Specified ignore patterns`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If --git is set, show git-untracked items instead
		if untrackedGitStatus {
			return showGitUntrackedItems()
		}

		entries, err := LoadIndex()
		if err != nil {
			return err
		}

		// Get root from argument
		root := untrackedRoot
		if len(args) > 0 {
			root = args[0]
		}
		if root == "" {
			root = "."
		}

		// Build set of indexed paths
		indexed := make(map[string]bool)
		for _, e := range entries {
			indexed[e.Path] = true
		}

		// Default ignore patterns
		ignore := map[string]bool{
			".git":           true,
			".github":        true,
			".agents":        true,
			"node_modules":   true,
			".node_modules":  true,
			"dist":           true,
			"build":          true,
			".DS_Store":      true,
			"__pycache__":    true,
			".pytest_cache":  true,
			"coverage":       true,
			".vscode":        true,
			".idea":          true,
		}

		// Parse additional ignore patterns
		if untrackedIgnore != "" {
			for _, pattern := range strings.Split(untrackedIgnore, ",") {
				pattern = strings.TrimSpace(pattern)
				if pattern != "" {
					ignore[pattern] = true
				}
			}
		}

		// Scan top level
		entries_to_add, err := findUntracked(root, indexed, ignore)
		if err != nil {
			return err
		}

		if len(entries_to_add) == 0 {
			fmt.Println("✓ No untracked directories/files found (not in index)")
			return nil
		}

		fmt.Printf("\n📁 Found %d entries not in index:\n", len(entries_to_add))
		fmt.Println("───────────────────────────────────────────────────────")
		for _, path := range entries_to_add {
			fi, _ := os.Stat(filepath.Join(root, path))
			typeStr := "📄"
			if fi != nil && fi.IsDir() {
				typeStr = "📁"
			}
			fmt.Printf("%s  %s\n", typeStr, path)
		}

		fmt.Println("\nTo add these to the index, run:")
		fmt.Println("  code-index extend")
		fmt.Println()
		return nil
	},
}

func showGitUntrackedItems() error {
	// Get untracked items from git
	cmd := exec.Command("git", "status", "--porcelain", "--untracked-files=all")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var untracked []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		// ?? = untracked files/dirs
		if strings.HasPrefix(line, "??") {
			path := strings.TrimSpace(line[3:])
			untracked = append(untracked, path)
		}
	}

	if len(untracked) == 0 {
		fmt.Println("✓ No git-untracked items found")
		return nil
	}

	fmt.Printf("\n📁 Found %d git-untracked items:\n", len(untracked))
	fmt.Println("───────────────────────────────────────────────────────")
	for _, path := range untracked {
		fi, _ := os.Stat(path)
		typeStr := "📄"
		if fi != nil && fi.IsDir() {
			typeStr = "📁"
		}
		fmt.Printf("%s  %s\n", typeStr, path)
	}
	fmt.Println()
	return nil
}

func findUntracked(root string, indexed, ignore map[string]bool) ([]string, error) {
	var untracked []string

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden and ignored
		if strings.HasPrefix(name, ".") || ignore[name] {
			continue
		}

		// Skip if indexed
		if indexed[name] {
			continue
		}

		// Add to untracked
		untracked = append(untracked, name)
	}

	return untracked, nil
}

func init() {
	rootCmd.AddCommand(untrackedCmd)
	untrackedCmd.Flags().StringVarP(&untrackedRoot, "root", "r", ".", "Repository root directory")
	untrackedCmd.Flags().StringVar(&untrackedIgnore, "ignore", "", "Additional comma-separated ignore patterns")
	untrackedCmd.Flags().BoolVar(&untrackedGitStatus, "git", false, "Show git-untracked items (instead of index-untracked)")
}
