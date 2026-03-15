package search

import (
	"math"
	"testing"
)

func TestBM25(t *testing.T) {
	tests := []struct {
		name      string
		tf        float64
		df        float64
		docLen    float64
		avgDocLen float64
		totalDocs int
		k1        float64
		b         float64
		check     func(t *testing.T, score float64)
	}{
		{
			name:      "basic score matches BM25 formula",
			tf:        1,
			df:        1,
			docLen:    100,
			avgDocLen: 100,
			totalDocs: 10,
			k1:        1.2,
			b:         0.75,
			check: func(t *testing.T, score float64) {
				// idf = ln((10-1+0.5)/(1+0.5) + 1) = ln(7.333) ≈ 1.9924
				// tfNorm = (1 * 2.2) / (1 + 1.2*(1-0.75+0.75*1)) = 2.2 / 2.2 = 1.0
				// score ≈ 1.9924
				if score <= 0 {
					t.Errorf("expected positive score, got %f", score)
				}
				if math.Abs(score-1.99) > 0.01 {
					t.Errorf("expected score ≈ 1.99, got %f", score)
				}
			},
		},
		{
			name:      "higher tf increases score",
			tf:        5,
			df:        1,
			docLen:    100,
			avgDocLen: 100,
			totalDocs: 10,
			k1:        1.2,
			b:         0.75,
			check: func(t *testing.T, score float64) {
				baseline := BM25(1, 1, 100, 100, 10, 1.2, 0.75)
				if score <= baseline {
					t.Errorf("higher tf should increase score: tf=5 got %f, tf=1 got %f", score, baseline)
				}
			},
		},
		{
			name:      "shorter doc scores higher than longer doc",
			tf:        1,
			df:        1,
			docLen:    50,
			avgDocLen: 100,
			totalDocs: 10,
			k1:        1.2,
			b:         0.75,
			check: func(t *testing.T, score float64) {
				longerScore := BM25(1, 1, 200, 100, 10, 1.2, 0.75)
				if score <= longerScore {
					t.Errorf("shorter doc should score higher: short=%f, long=%f", score, longerScore)
				}
			},
		},
		{
			name:      "rarer term scores higher",
			tf:        1,
			df:        1,
			docLen:    100,
			avgDocLen: 100,
			totalDocs: 100,
			k1:        1.2,
			b:         0.75,
			check: func(t *testing.T, score float64) {
				commonScore := BM25(1, 50, 100, 100, 100, 1.2, 0.75)
				if score <= commonScore {
					t.Errorf("rarer term (df=1) should score higher than common (df=50): rare=%f, common=%f", score, commonScore)
				}
			},
		},
		{
			name:      "score is non-negative",
			tf:        1,
			df:        1,
			docLen:    100,
			avgDocLen: 100,
			totalDocs: 10,
			k1:        1.2,
			b:         0.75,
			check: func(t *testing.T, score float64) {
				if score < 0 {
					t.Errorf("score should be non-negative, got %f", score)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := BM25(tt.tf, tt.df, tt.docLen, tt.avgDocLen, tt.totalDocs, tt.k1, tt.b)
			tt.check(t, score)
		})
	}
}

func TestBM25RareTermShortDocVsCommonTermLongDoc(t *testing.T) {
	t.Parallel()

	// Rare term (df=1) in short document (50 words) should score higher
	// than common term (df=50) in long document (500 words)
	rareShortScore := BM25(1, 1, 50, 100, 100, 1.2, 0.75)
	commonLongScore := BM25(1, 50, 500, 100, 100, 1.2, 0.75)

	if rareShortScore <= commonLongScore {
		t.Errorf("rare term in short doc (%f) should score higher than common term in long doc (%f)",
			rareShortScore, commonLongScore)
	}
}

func TestProximityBonus(t *testing.T) {
	tests := []struct {
		name       string
		positions  [][]int
		windowSize int
		expected   float64
	}{
		{
			name:       "adjacent positions",
			positions:  [][]int{{1}, {2}},
			windowSize: 5,
			expected:   2.0,
		},
		{
			name:       "within 3 words",
			positions:  [][]int{{1}, {3}},
			windowSize: 5,
			expected:   1.5,
		},
		{
			name:       "within 10 words",
			positions:  [][]int{{1}, {8}},
			windowSize: 15,
			expected:   1.2,
		},
		{
			name:       "far apart",
			positions:  [][]int{{1}, {20}},
			windowSize: 25,
			expected:   1.0,
		},
		{
			name:       "single term",
			positions:  [][]int{{1, 5, 10}},
			windowSize: 15,
			expected:   1.0,
		},
		{
			name:       "empty positions",
			positions:  [][]int{},
			windowSize: 5,
			expected:   1.0,
		},
		{
			name:       "three terms adjacent",
			positions:  [][]int{{1}, {2}, {3}},
			windowSize: 5,
			expected:   2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProximityBonus(tt.positions, tt.windowSize)
			if got != tt.expected {
				t.Errorf("ProximityBonus(%v, %d) = %f, want %f", tt.positions, tt.windowSize, got, tt.expected)
			}
		})
	}
}
