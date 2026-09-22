package cmd

import (
	"fmt"
	"os"
	"strings"
	"golang.org/x/term"
)

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func PrintResults(entries []Entry, format string, wrapFlag bool) {
	// Auto-enable wrapping for TTY unless explicitly disabled
	shouldWrap := wrapFlag || (isTTY() && !wrapFlag)

	if format == "flat" {
		PrintResultsFlat(entries, shouldWrap)
	} else {
		PrintResultsTable(entries)
	}
}

func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	// Break on word boundaries
	words := strings.Fields(text)
	var lines []string
	var line string

	for _, word := range words {
		if len(line)+len(word)+1 <= width {
			if line == "" {
				line = word
			} else {
				line += " " + word
			}
		} else {
			if line != "" {
				lines = append(lines, line)
			}
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n  ")
}

func truncate(s string, width int) string {
	if len(s) <= width {
		return s
	}
	return s[:width-3] + "..."
}

func PrintResultsFlat(entries []Entry, wrap bool) {
	if wrap {
		// Wrapped format with fixed-width columns
		const (
			typeW      = 5
			pathW      = 35
			primaryW   = 20
			secondaryW = 30
			summaryW   = 50
		)

		// Header
		fmt.Printf("%-*s\t%-*s\t%-*s\t%-*s\t%s\n",
			typeW, "type", pathW, "path", primaryW, "primary", secondaryW, "secondary", "summary")

		// Data rows
		for _, e := range entries {
			path := truncate(e.Path, pathW)
			primary := truncate(e.Primary, primaryW)
			secondary := truncate(e.Secondary, secondaryW)
			summary := truncate(e.Summary, summaryW)

			fmt.Printf("%-*s\t%-*s\t%-*s\t%-*s\t%s\n",
				typeW, e.Type,
				pathW, path,
				primaryW, primary,
				secondaryW, secondary,
				summary)
		}
	} else {
		// Flat format: pipeline-friendly (no truncation)
		fmt.Println("type\tpath\tprimary_category\tsecondary_categories\tsummary")
		for _, e := range entries {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", e.Type, e.Path, e.Primary, e.Secondary, e.Summary)
		}
	}
}

func PrintResultsTable(entries []Entry) {
	fmt.Println("┌─────────────┬──────────────────────────────────────┬──────────────────────┬───────────────────────────────────────┐")
	fmt.Println("│ Type │ Path │ Primary Category │ Secondary Categories │ Summary │")
	fmt.Println("├─────────────┼──────────────────────────────────────┼──────────────────────┼───────────────────────────────────────┤")

	for _, e := range entries {
		typeStr := e.Type
		if len(typeStr) < 4 {
			typeStr += strings.Repeat(" ", 4-len(typeStr))
		}

		summary := e.Summary
		if len(summary) > 35 {
			summary = summary[:32] + "..."
		}

		secondary := e.Secondary
		if len(secondary) > 35 {
			secondary = secondary[:32] + "..."
		}

		fmt.Printf("│ %-11s │ %-36s │ %-20s │ %-37s │\n",
			typeStr, e.Path, e.Primary, secondary)
		fmt.Printf("│             │                                      │                      │ %s │\n", summary)
		fmt.Println("├─────────────┼──────────────────────────────────────┼──────────────────────┼───────────────────────────────────────┤")
	}
}

func PrintCategories(categories []Category) {
	fmt.Println("\nCategories (sorted by frequency):")
	fmt.Println("───────────────────────────────────")
	for _, cat := range categories {
		fmt.Printf("%-30s : %2d entries\n", cat.Name, cat.Count)
	}
}

func PrintPaths(entries []Entry) {
	fmt.Println("\nAll Paths:")
	fmt.Println("──────────")
	for i, e := range entries {
		prefix := "├─ "
		if i == len(entries)-1 {
			prefix = "└─ "
		}
		typeStr := ""
		if e.Type == "dir" {
			typeStr = "📁 "
		} else {
			typeStr = "📄 "
		}
		fmt.Printf("%s%s%s\n", prefix, typeStr, e.Path)
	}
}
