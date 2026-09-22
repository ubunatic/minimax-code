package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	manInstall bool
	manDir     string
)

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate and install roff man pages",
	Long: `Generate standard roff/troff man pages for code-index.

Without flags, prints the roff man page to stdout (e.g. code-index man | man -l -).
With --install, writes .1 man pages to ~/.local/share/man/man1 (or /usr/local/share/man/man1 if root).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if manInstall || manDir != "" {
			return installManPages(rootCmd, manDir, cmd.OutOrStdout())
		}
		header := &doc.GenManHeader{
			Title:   "code-index",
			Section: "1",
			Source:  "code-index 1.0",
			Manual:  "Repository Index Query Tool",
		}
		return doc.GenMan(rootCmd, header, cmd.OutOrStdout())
	},
}

func defaultManDir() string {
	if os.Geteuid() == 0 {
		return "/usr/local/share/man/man1"
	}
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "man", "man1")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".local", "share", "man", "man1")
	}
	return "/usr/local/share/man/man1"
}

func installManPages(rootCmd *cobra.Command, targetDir string, out io.Writer) error {
	if targetDir == "" {
		targetDir = defaultManDir()
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("creating man directory %s: %w", targetDir, err)
	}
	header := &doc.GenManHeader{
		Title:   "code-index",
		Section: "1",
		Source:  "code-index 1.0",
		Manual:  "Repository Index Query Tool",
	}
	if err := doc.GenManTree(rootCmd, header, targetDir); err != nil {
		return fmt.Errorf("generating man pages in %s: %w", targetDir, err)
	}
	fmt.Fprintf(out, "✓ Installed man pages to %s\nRun 'man code-index' to view.\n", targetDir)
	return nil
}

func init() {
	rootCmd.AddCommand(manCmd)
	manCmd.Flags().BoolVar(&manInstall, "install", false, "Install man pages into standard system or user man directory")
	manCmd.Flags().StringVarP(&manDir, "dir", "d", "", "Custom target directory for man pages (implies --install)")
}
