package cmd

import (
	"testing"
)

func TestScoreMatch(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		pattern string
		minScore int // Should be at least this
	}{
		// Exact and prefix matches
		{"exact match", "agent", "agent", 100},
		{"prefix match", "agent-tools", "agent", 90},
		{"substring match", "agent-tools", "tools", 70},

		// Fuzzy matching with typos
		{"typo 1-edit", "delegation", "delegat", 20},
		{"typo 2-edits", "permission", "permision", 15},

		// No match
		{"no match", "random text", "xyz", 0},
		{"pattern too short", "something", "ab", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scoreMatch(tt.text, tt.pattern)
			if score < tt.minScore {
				t.Errorf("scoreMatch(%q, %q) = %d, want >= %d", tt.text, tt.pattern, score, tt.minScore)
			}
		})
	}
}

func TestScoreMultiWord(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		filter         string
		expectPositive bool
		description    string
	}{
		// Single-word queries should work
		{"single-word match", "delegation rules", "delegation", true, "single word should delegate to scoreMatch"},
		{"single-word no match", "authorization rules", "delegation", false, "single word no match"},

		// Multi-word queries with OR semantics
		{"multi-word both match", "tool definitions", "tool definitions", true, "both words in text"},
		{"multi-word one matches", "tool definitions", "tool call", true, "tool matches (OR semantics)"},
		{"multi-word neither matches", "authorization", "tool call", false, "neither word matches"},

		// Stopword edge case (known behavior: OR + substring matching allows short words to match substrings)
		// This is a precision trade-off of OR semantics with substring matching
		{"stopword matches substring", "gather weather", "the a", true, "short words match substrings ('the' in 'gathe[r]', 'a' in 'weath[er]')"},
		{"meaningful words match", "agent delegation", "agent delegation", true, "meaningful content words"},

		// Whitespace edge cases
		{"single space", "text", " ", false, "whitespace-only filter returns no words"},
		{"empty string", "text", "", false, "empty filter"},

		// Multi-word with fuzzy matching
		{"multi-word with typo", "delegation rules", "delegat call", true, "fuzzy match on first word"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scoreMultiWord(tt.text, tt.filter)
			gotMatch := score > 0
			if gotMatch != tt.expectPositive {
				t.Errorf("scoreMultiWord(%q, %q) = %d (match=%v), expected match=%v\n  %s",
					tt.text, tt.filter, score, gotMatch, tt.expectPositive, tt.description)
			}
		})
	}
}

func TestFilterEntriesMultiWord(t *testing.T) {
	// Integration test: verify multi-word searches work end-to-end
	entries := []Entry{
		{
			Type:      "file",
			Path:      "packages/agent-tools/src/desktop/subagent-roles.ts",
			Primary:   "delegation",
			Secondary: "subagent-authorization",
			Summary:   "Subagent role definitions and delegation rules",
		},
		{
			Type:      "dir",
			Path:      "packages/agent-tools",
			Primary:   "tool-definitions",
			Secondary: "builtin-tools",
			Summary:   "Platform-owned runtime tool definitions and helpers",
		},
		{
			Type:      "file",
			Path:      "README.md",
			Primary:   "documentation",
			Secondary: "",
			Summary:   "Product overview and quick start guide",
		},
	}

	tests := []struct {
		searchFilter string
		expectCount  int
		description  string
	}{
		{"tool", 2, "matches both agent-tools path and tool-definitions summary"},
		{"delegation", 1, "matches delegation in subagent-roles entry"},
		{"agent delegation", 2, "matches agent-tools (agent in path) and subagent-roles (delegation in primary)"},
		{"impossible xyz", 0, "non-existent terms return no results"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			results := FilterEntries(entries, "", "", tt.searchFilter, "")
			if len(results) != tt.expectCount {
				t.Errorf("FilterEntries with search %q: got %d results, expected %d",
					tt.searchFilter, len(results), tt.expectCount)
			}
		})
	}
}
