package search

import (
	"strings"
	"testing"
)

func TestExtractSnippet(t *testing.T) {
	tests := []struct {
		name       string
		doc        SearchDocument
		queryTerms []string
		maxLen     int
		checks     func(t *testing.T, snippet string)
	}{
		{
			name: "terms at beginning",
			doc: SearchDocument{
				Fields: map[string]string{
					"body": "installation guide for setting up the system",
				},
			},
			queryTerms: []string{"installation"},
			maxLen:     200,
			checks: func(t *testing.T, snippet string) {
				if !strings.Contains(snippet, "installation") {
					t.Errorf("snippet should contain 'installation', got: %q", snippet)
				}
				if strings.HasPrefix(snippet, "...") {
					t.Errorf("snippet should not have '...' prefix when match is at beginning, got: %q", snippet)
				}
			},
		},
		{
			name: "terms in middle",
			doc: SearchDocument{
				Fields: map[string]string{
					"body": "first part of the document then the installation guide appears here followed by more text that extends beyond the limit",
				},
			},
			queryTerms: []string{"installation"},
			maxLen:     50,
			checks: func(t *testing.T, snippet string) {
				if !strings.Contains(snippet, "installation") {
					t.Errorf("snippet should contain 'installation', got: %q", snippet)
				}
				if !strings.HasPrefix(snippet, "...") {
					t.Errorf("snippet should have '...' prefix when match is in middle, got: %q", snippet)
				}
			},
		},
		{
			name: "multiple query terms",
			doc: SearchDocument{
				Fields: map[string]string{
					"body": "the quick brown fox jumps over the lazy dog",
				},
			},
			queryTerms: []string{"quick", "fox"},
			maxLen:     200,
			checks: func(t *testing.T, snippet string) {
				if !strings.Contains(snippet, "quick") {
					t.Errorf("snippet should contain 'quick', got: %q", snippet)
				}
				if !strings.Contains(snippet, "fox") {
					t.Errorf("snippet should contain 'fox', got: %q", snippet)
				}
			},
		},
		{
			name: "no matching terms",
			doc: SearchDocument{
				Fields: map[string]string{
					"body": "hello world this is a test document",
				},
			},
			queryTerms: []string{"nonexistent"},
			maxLen:     20,
			checks: func(t *testing.T, snippet string) {
				if len(snippet) == 0 {
					t.Error("snippet should not be empty for non-empty document")
				}
				// Should return first ~20 chars of body, likely truncated.
				if !strings.Contains(snippet, "hello") {
					t.Errorf("snippet should contain beginning of text, got: %q", snippet)
				}
			},
		},
		{
			name: "empty document",
			doc: SearchDocument{
				Fields: map[string]string{},
			},
			queryTerms: []string{"test"},
			maxLen:     200,
			checks: func(t *testing.T, snippet string) {
				if snippet != "" {
					t.Errorf("snippet should be empty for empty document, got: %q", snippet)
				}
			},
		},
		{
			name: "html stripped",
			doc: SearchDocument{
				Fields: map[string]string{
					"body": "<p>hello <b>world</b></p>",
				},
			},
			queryTerms: []string{"world"},
			maxLen:     200,
			checks: func(t *testing.T, snippet string) {
				if !strings.Contains(snippet, "world") {
					t.Errorf("snippet should contain 'world', got: %q", snippet)
				}
				if strings.Contains(snippet, "<") || strings.Contains(snippet, ">") {
					t.Errorf("snippet should not contain HTML tags, got: %q", snippet)
				}
			},
		},
		{
			name: "short text",
			doc: SearchDocument{
				Fields: map[string]string{
					"body": "short",
				},
			},
			queryTerms: []string{"short"},
			maxLen:     200,
			checks: func(t *testing.T, snippet string) {
				if snippet != "short" {
					t.Errorf("snippet should be exactly 'short', got: %q", snippet)
				}
				if strings.Contains(snippet, "...") {
					t.Errorf("snippet should not have ellipsis for short text, got: %q", snippet)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snippet := ExtractSnippet(tt.doc, tt.queryTerms, tt.maxLen)
			tt.checks(t, snippet)
		})
	}
}
