package utility

import (
	"sort"
	"strings"
	"unicode"
)

// Scoring constants tuned for short CMS identifiers (names, labels, slugs).
const (
	scoreMatch        = 1    // base score per matched character
	bonusConsecutive  = 15   // per char in a consecutive run
	bonusWordBoundary = 30   // after separator or camelCase transition
	bonusFirstChar    = 15   // query[0] matches target[0]
	penaltyGap        = -3   // per unmatched target character between matches
	scoreExact        = 1000 // ceiling for exact and prefix matches
)

// FuzzyResult is the outcome of a fuzzy match operation.
type FuzzyResult struct {
	Index          int   // position of the matched item in the original slice
	Score          int   // higher = better match; can be negative for scattered matches
	MatchedIndices []int // rune positions within the best-scoring field that matched
}

// FuzzyScore scores query against target using subsequence matching.
// Returns (score, matchedRunePositions, matched). ok=false means query is
// not a subsequence of target. Case-insensitive. Both strings are converted
// to rune slices internally. Negative scores are possible for scattered
// matches in long targets — the caller decides whether to filter them.
func FuzzyScore(query, target string) (score int, matched []int, ok bool) {
	if query == "" {
		return 0, []int{}, true
	}
	if target == "" {
		return 0, nil, false
	}

	// Exact match short-circuit
	if strings.EqualFold(query, target) {
		indices := make([]int, len([]rune(target)))
		for i := range indices {
			indices[i] = i
		}
		return scoreExact, indices, true
	}

	queryLower := strings.ToLower(query)
	targetLower := strings.ToLower(target)

	// Prefix match short-circuit
	if strings.HasPrefix(targetLower, queryLower) {
		qLen := len([]rune(query))
		indices := make([]int, qLen)
		for i := range indices {
			indices[i] = i
		}
		return scoreExact - 1, indices, true
	}

	queryRunes := []rune(queryLower)
	targetRunes := []rune(targetLower)
	origTargetRunes := []rune(target)

	if len(queryRunes) > len(targetRunes) {
		return 0, nil, false
	}

	qIdx := 0
	consecutive := 0
	matched = make([]int, 0, len(queryRunes))

	for tIdx := 0; tIdx < len(targetRunes) && qIdx < len(queryRunes); tIdx++ {
		if targetRunes[tIdx] == queryRunes[qIdx] {
			score += scoreMatch + computeBonus(origTargetRunes, tIdx, consecutive)
			matched = append(matched, tIdx)
			consecutive++
			qIdx++
		} else {
			score += penaltyGap
			consecutive = 0
		}
	}

	if qIdx < len(queryRunes) {
		return 0, nil, false
	}

	// First-char bonus
	if len(matched) > 0 && matched[0] == 0 {
		score += bonusFirstChar
	}

	return score, matched, true
}

// FuzzyMatch scores query against each candidate string.
// Returns only matching results, sorted by score descending (best first).
// Empty query returns all candidates with score 0 in original order.
func FuzzyMatch(query string, candidates []string) []FuzzyResult {
	if query == "" {
		results := make([]FuzzyResult, len(candidates))
		for i := range candidates {
			results[i] = FuzzyResult{Index: i, Score: 0, MatchedIndices: []int{}}
		}
		return results
	}

	results := make([]FuzzyResult, 0, len(candidates))
	for i, c := range candidates {
		score, indices, ok := FuzzyScore(query, c)
		if ok {
			results = append(results, FuzzyResult{Index: i, Score: score, MatchedIndices: indices})
		}
	}

	sort.Slice(results, func(a, b int) bool {
		return results[a].Score > results[b].Score
	})
	return results
}

// FuzzyFind scores query against each item in items. fn is called once per
// item to extract one or more searchable strings. The best score across all
// extracted strings for an item is used. Returns only matching items, sorted
// by score descending (best first). Empty query returns all items with
// score 0 in original order.
func FuzzyFind[T any](query string, items []T, fn func(T) []string) []FuzzyResult {
	if query == "" {
		results := make([]FuzzyResult, len(items))
		for i := range items {
			results[i] = FuzzyResult{Index: i, Score: 0, MatchedIndices: []int{}}
		}
		return results
	}

	results := make([]FuzzyResult, 0, len(items))
	for i, item := range items {
		fields := fn(item)
		bestScore := 0
		var bestMatched []int
		found := false

		for _, field := range fields {
			s, m, ok := FuzzyScore(query, field)
			if ok && (!found || s > bestScore) {
				bestScore = s
				bestMatched = m
				found = true
			}
		}

		if found {
			results = append(results, FuzzyResult{Index: i, Score: bestScore, MatchedIndices: bestMatched})
		}
	}

	sort.Slice(results, func(a, b int) bool {
		return results[a].Score > results[b].Score
	})
	return results
}

func computeBonus(targetRunes []rune, tIdx int, consecutive int) int {
	bonus := 0

	if consecutive > 0 {
		bonus += bonusConsecutive
	}

	if tIdx == 0 {
		bonus += bonusWordBoundary
	} else {
		prev := targetRunes[tIdx-1]
		cur := targetRunes[tIdx]

		if isSeparator(prev) {
			bonus += bonusWordBoundary
		} else if unicode.IsLower(prev) && unicode.IsUpper(cur) {
			bonus += bonusWordBoundary
		}
	}

	return bonus
}

func isSeparator(r rune) bool {
	return r == '_' || r == '-' || r == '.' || r == ' ' || r == '/' || r == ':'
}
