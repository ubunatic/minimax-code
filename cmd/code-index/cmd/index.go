package cmd

import (
	_ "embed"
	"encoding/csv"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

//go:embed data.csv
var indexCSV string

type Entry struct {
	Type       string
	Path       string
	Primary    string
	Secondary  string
	Summary    string
}

func LoadIndex() ([]Entry, error) {
	reader := csv.NewReader(strings.NewReader(indexCSV))
	reader.FieldsPerRecord = 5

	// Skip header
	_, err := reader.Read()
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		entries = append(entries, Entry{
			Type:       record[0],
			Path:       record[1],
			Primary:    record[2],
			Secondary:  record[3],
			Summary:    record[4],
		})
	}

	return entries, nil
}

// levenshteinDistance computes the edit distance between two strings
func levenshteinDistance(a, b string) int {
	a, b = strings.ToLower(a), strings.ToLower(b)
	lenA, lenB := utf8.RuneCountInString(a), utf8.RuneCountInString(b)

	if lenA == 0 {
		return lenB
	}
	if lenB == 0 {
		return lenA
	}

	aRunes := []rune(a)
	bRunes := []rune(b)

	prev := make([]int, lenB+1)
	for j := 0; j <= lenB; j++ {
		prev[j] = j
	}

	for i := 1; i <= lenA; i++ {
		curr := make([]int, lenB+1)
		curr[0] = i

		for j := 1; j <= lenB; j++ {
			cost := 0
			if aRunes[i-1] != bRunes[j-1] {
				cost = 1
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev = curr
	}

	return prev[lenB]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// scoreMatch returns a relevance score (0-100) for how well pattern matches text
// Higher scores = better match. Supports fuzzy matching, substrings, and typos.
func scoreMatch(text, pattern string) int {
	text = strings.ToLower(text)
	pattern = strings.ToLower(pattern)

	// Exact match
	if text == pattern {
		return 100
	}

	// Prefix match (starts with pattern)
	if strings.HasPrefix(text, pattern) {
		return 90
	}

	// Exact word match (with hyphens normalized)
	patternNorm := strings.ReplaceAll(pattern, " ", "-")
	if text == patternNorm {
		return 95
	}

	// Substring match
	if strings.Contains(text, pattern) {
		return 70
	}

	// Hyphen-normalized substring
	if strings.Contains(text, patternNorm) {
		return 75
	}

	// Fuzzy match using edit distance (for typos and partial matches)
	// Only consider if pattern length >= 4 to avoid too many false positives
	if len(pattern) >= 4 && len(text) >= 4 {
		distance := levenshteinDistance(text, pattern)
		// Allow up to 2 edits for reasonable typo matching
		if distance > 0 && distance <= 2 {
			// Score based on how close the match is
			// 1 edit = 30-35 pts, 2 edits = 20-25 pts
			score := 35 - (distance * 8)
			return max(15, score)
		}
	}

	// No match
	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// scoreMultiWord handles multi-word queries by matching against individual query words
// Returns highest score from any matching query word (OR semantics)
// Multi-word queries match if ANY word in the query matches the text
func scoreMultiWord(text, filter string) int {
	queryWords := strings.Fields(filter)
	if len(queryWords) == 0 {
		return 0
	}

	// For single-word queries, match against the full text
	if len(queryWords) == 1 {
		return scoreMatch(text, queryWords[0])
	}

	// For multi-word queries: use OR semantics
	// Return the best score from any query word matching the text
	maxScore := 0
	for _, queryWord := range queryWords {
		// Try matching the query word against full text
		if score := scoreMatch(text, queryWord); score > maxScore {
			maxScore = score
		}
		// Also try individual words in the text for better fuzzy matching
		for _, textWord := range strings.Fields(text) {
			if score := scoreMatch(textWord, queryWord); score > maxScore {
				maxScore = score
			}
		}
	}

	return maxScore
}

// scoreCategory checks if category matches any word in filter (OR mode)
// Returns highest score from any matching word
func scoreCategory(text, filter string) int {
	words := strings.Fields(filter)
	maxScore := 0

	for _, word := range words {
		score := scoreMatch(text, word)
		if score > maxScore {
			maxScore = score
		}
	}

	return maxScore
}

type scoredEntry struct {
	entry Entry
	score int
}

func FilterEntries(entries []Entry, pathFilter, categoryFilter, searchFilter, typeFilter string) []Entry {
	var results []scoredEntry

	for _, e := range entries {
		// Type filter
		if typeFilter != "" && e.Type != typeFilter {
			continue
		}

		// Path filter
		if pathFilter != "" && !strings.Contains(e.Path, pathFilter) {
			continue
		}

		// Category filter with scoring (OR mode for multiple words)
		score := 0
		if categoryFilter != "" {
			// Check primary category
			primaryScore := scoreCategory(e.Primary, categoryFilter)
			if primaryScore > 0 {
				score = primaryScore + 10 // Boost primary matches
			} else if e.Secondary != "" {
				// Check secondary categories
				for _, cat := range strings.Split(e.Secondary, ";") {
					catScore := scoreCategory(strings.TrimSpace(cat), categoryFilter)
					if catScore > score {
						score = catScore
					}
				}
			}

			if score == 0 {
				continue
			}
		}

		// Search filter with fuzzy matching on summary and keywords
		if searchFilter != "" {
			searchScore := scoreMultiWord(e.Summary, searchFilter)
			if searchScore == 0 {
				// Also check path for matches
				if pathScore := scoreMultiWord(e.Path, searchFilter); pathScore > 0 {
					score += pathScore / 2 // Lower priority for path matches
				} else {
					continue
				}
			} else {
				score += searchScore
			}
		}

		results = append(results, scoredEntry{e, score})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		// Secondary sort by path if scores are equal
		return results[i].entry.Path < results[j].entry.Path
	})

	// Extract just the entries
	var sorted []Entry
	for _, se := range results {
		sorted = append(sorted, se.entry)
	}
	return sorted
}

type Category struct {
	Name  string
	Count int
}

func ListCategories(entries []Entry) []Category {
	catCount := make(map[string]int)

	for _, e := range entries {
		catCount[e.Primary]++

		if e.Secondary != "" {
			for _, cat := range strings.Split(e.Secondary, ";") {
				cat = strings.TrimSpace(cat)
				if cat != "" {
					catCount[cat]++
				}
			}
		}
	}

	// Sort categories by count (descending), then alphabetically
	var sorted []Category
	for k, v := range catCount {
		sorted = append(sorted, Category{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Count != sorted[j].Count {
			return sorted[i].Count > sorted[j].Count
		}
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}
