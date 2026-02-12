package components

import (
	"regexp"
	"sort"
	"strings"
)

// FuzzyMatch represents the result of a fuzzy match.
// Lower score = better match.
type FuzzyMatch struct {
	Matches bool
	Score   float64
}

// fuzzyMatch performs fuzzy matching between query and text.
// Matches if all query characters appear in order (not necessarily consecutive).
func fuzzyMatch(query, text string) FuzzyMatch {
	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(text)

	matchQuery := func(normalizedQuery string) FuzzyMatch {
		if len(normalizedQuery) == 0 {
			return FuzzyMatch{Matches: true, Score: 0}
		}

		if len(normalizedQuery) > len(textLower) {
			return FuzzyMatch{Matches: false, Score: 0}
		}

		queryIndex := 0
		score := 0.0
		lastMatchIndex := -1
		consecutiveMatches := 0

		for i := 0; i < len(textLower) && queryIndex < len(normalizedQuery); i++ {
			if textLower[i] == normalizedQuery[queryIndex] {
				isWordBoundary := i == 0 || isWordBoundaryChar(textLower[i-1])

				// Reward consecutive matches
				if lastMatchIndex == i-1 {
					consecutiveMatches++
					score -= float64(consecutiveMatches) * 5
				} else {
					consecutiveMatches = 0
					// Penalize gaps
					if lastMatchIndex >= 0 {
						score += float64(i-lastMatchIndex-1) * 2
					}
				}

				// Reward word boundary matches
				if isWordBoundary {
					score -= 10
				}

				// Slight penalty for later matches
				score += float64(i) * 0.1

				lastMatchIndex = i
				queryIndex++
			}
		}

		if queryIndex < len(normalizedQuery) {
			return FuzzyMatch{Matches: false, Score: 0}
		}

		return FuzzyMatch{Matches: true, Score: score}
	}

	// Try primary match
	primaryMatch := matchQuery(queryLower)
	if primaryMatch.Matches {
		return primaryMatch
	}

	// Try swapping letters and digits
	swappedQuery := getSwappedQuery(queryLower)
	if swappedQuery == "" {
		return primaryMatch
	}

	swappedMatch := matchQuery(swappedQuery)
	if !swappedMatch.Matches {
		return primaryMatch
	}

	return FuzzyMatch{Matches: true, Score: swappedMatch.Score + 5}
}

// isWordBoundaryChar checks if a character is a word boundary character.
func isWordBoundaryChar(c byte) bool {
	return c == ' ' || c == '-' || c == '_' || c == '.' || c == '/' || c == ':'
}

// getSwappedQuery attempts to swap letters and digits in the query.
func getSwappedQuery(query string) string {
	// Match "letters+digits" pattern
	alphaNumericRe := regexp.MustCompile(`^([a-z]+)([0-9]+)$`)
	if matches := alphaNumericRe.FindStringSubmatch(query); matches != nil {
		return matches[2] + matches[1]
	}

	// Match "digits+letters" pattern
	numericAlphaRe := regexp.MustCompile(`^([0-9]+)([a-z]+)$`)
	if matches := numericAlphaRe.FindStringSubmatch(query); matches != nil {
		return matches[2] + matches[1]
	}

	return ""
}

// FuzzyFilter filters and sorts items by fuzzy match quality (best matches first).
// Supports space-separated tokens: all tokens must match.
func FuzzyFilter[T any](items []T, query string, getText func(T) string) []T {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return items
	}

	// Split query into tokens
	tokens := strings.Fields(trimmedQuery)
	if len(tokens) == 0 {
		return items
	}

	type result struct {
		item       T
		totalScore float64
	}

	var results []result

	for _, item := range items {
		text := getText(item)
		totalScore := 0.0
		allMatch := true

		for _, token := range tokens {
			match := fuzzyMatch(token, text)
			if match.Matches {
				totalScore += match.Score
			} else {
				allMatch = false
				break
			}
		}

		if allMatch {
			results = append(results, result{item: item, totalScore: totalScore})
		}
	}

	// Sort by score (lower is better)
	sort.Slice(results, func(i, j int) bool {
		return results[i].totalScore < results[j].totalScore
	})

	// Extract sorted items
	filtered := make([]T, len(results))
	for i, r := range results {
		filtered[i] = r.item
	}

	return filtered
}
