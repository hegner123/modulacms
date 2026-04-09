package search

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple words with punctuation",
			input:    "Hello, World!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "HTML tags stripped",
			input:    "<p>Hello <b>world</b></p>",
			expected: []string{"hello", "world"},
		},
		{
			name:     "unicode preserved",
			input:    "café résumé",
			expected: []string{"café", "résumé"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "non-letter non-digit splits",
			input:    "C++ is great",
			expected: []string{"c", "is", "great"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tokenize(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Tokenize(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tags replaced with spaces",
			input:    "<p>Hello <b>world</b></p>",
			expected: " Hello  world  ",
		},
		{
			name:     "no tags",
			input:    "no tags here",
			expected: "no tags here",
		},
		{
			name:     "HTML entities decoded",
			input:    "&amp; &lt; &gt; &quot; &#39;",
			expected: "& < > \" '",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripHTML(tt.input)
			if got != tt.expected {
				t.Errorf("StripHTML(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTokenizeAndFilter(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		stopWords         map[string]bool
		expectedTerms     []string
		expectedPositions []int
	}{
		{
			name:              "apostrophe splits and stop words filtered",
			input:             "it's a test",
			stopWords:         defaultStopWords,
			expectedTerms:     []string{"s", "test"},
			expectedPositions: []int{1, 3},
		},
		{
			name:              "the filtered as stop word",
			input:             "the quick brown fox",
			stopWords:         defaultStopWords,
			expectedTerms:     []string{"quick", "brown", "fox"},
			expectedPositions: []int{1, 2, 3},
		},
		{
			name:              "empty string",
			input:             "",
			stopWords:         defaultStopWords,
			expectedTerms:     nil,
			expectedPositions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terms, positions := TokenizeAndFilter(tt.input, tt.stopWords)
			if !reflect.DeepEqual(terms, tt.expectedTerms) {
				t.Errorf("TokenizeAndFilter(%q) terms = %v, want %v", tt.input, terms, tt.expectedTerms)
			}
			if !reflect.DeepEqual(positions, tt.expectedPositions) {
				t.Errorf("TokenizeAndFilter(%q) positions = %v, want %v", tt.input, positions, tt.expectedPositions)
			}
		})
	}
}

func TestTokenizeAndFilterStopWords(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTerms []string
	}{
		{
			name:          "apostrophe splits and stop word 'a' filtered",
			input:         "it's a test",
			expectedTerms: []string{"s", "test"},
			// "it" and "a" are both default stop words
		},
		{
			name:          "stop word 'is' filtered",
			input:         "C++ is great",
			expectedTerms: []string{"c", "great"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terms, _ := TokenizeAndFilter(tt.input, defaultStopWords)
			if !reflect.DeepEqual(terms, tt.expectedTerms) {
				t.Errorf("TokenizeAndFilter(%q) terms = %v, want %v", tt.input, terms, tt.expectedTerms)
			}
		})
	}
}

func TestTokenizeWithOffsets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TermOffset
	}{
		{
			name:  "ASCII words",
			input: "hello world",
			expected: []TermOffset{
				{Term: "hello", Start: 0, End: 5},
				{Term: "world", Start: 6, End: 11},
			},
		},
		{
			name:  "multi-byte UTF-8",
			input: "café",
			expected: []TermOffset{
				{Term: "café", Start: 0, End: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TokenizeWithOffsets(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("TokenizeWithOffsets(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
