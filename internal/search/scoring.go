package search

import (
	"math"
	"sort"
)

// BM25 computes the BM25 score for a single term in a single field of a single document.
func BM25(tf float64, df float64, docLen float64, avgDocLen float64, totalDocs int, k1 float64, b float64) float64 {
	idf := math.Log((float64(totalDocs)-df+0.5)/(df+0.5) + 1.0)
	tfNorm := (tf * (k1 + 1.0)) / (tf + k1*(1.0-b+b*(docLen/avgDocLen)))
	return idf * tfNorm
}

type posEntry struct {
	position  int
	termIndex int
}

// ProximityBonus returns a multiplier based on how close query terms are
// to each other within a field. Adjacent terms (phrase match) get the
// highest bonus.
func ProximityBonus(positions [][]int, windowSize int) float64 {
	// Count how many terms actually have positions
	termsWithPositions := 0
	for _, ps := range positions {
		if len(ps) > 0 {
			termsWithPositions++
		}
	}
	if termsWithPositions < 2 {
		return 1.0
	}

	// Collect all (position, termIndex) pairs
	var entries []posEntry
	for termIdx, ps := range positions {
		for _, p := range ps {
			entries = append(entries, posEntry{position: p, termIndex: termIdx})
		}
	}

	// Sort by position
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].position < entries[j].position
	})

	// Sliding window: expand right until all terms present, then shrink left
	termCount := make(map[int]int)
	distinctTerms := termsWithPositions
	minWindow := math.MaxInt

	left := 0
	for right := 0; right < len(entries); right++ {
		termCount[entries[right].termIndex]++

		// Try to shrink from the left while all terms are still represented
		for len(termCount) == distinctTerms {
			w := entries[right].position - entries[left].position + 1
			if w < minWindow {
				minWindow = w
			}
			termCount[entries[left].termIndex]--
			if termCount[entries[left].termIndex] == 0 {
				delete(termCount, entries[left].termIndex)
			}
			left++
		}
	}

	if minWindow == math.MaxInt {
		return 1.0
	}

	// Return multiplier based on window size
	if minWindow <= distinctTerms {
		return 2.0
	}
	if minWindow <= 3 {
		return 1.5
	}
	if minWindow <= 10 {
		return 1.2
	}
	return 1.0
}
