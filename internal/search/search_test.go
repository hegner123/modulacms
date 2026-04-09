package search

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// --- 10.7 Integration tests ---

func TestSearchRankedResults(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		DatatypeName:  "page",
		Fields:        map[string]string{"title": "Alpha", "body": "Simple introduction to the platform"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		DatatypeName:  "page",
		Fields:        map[string]string{"title": "Platform Guide", "body": "platform platform platform usage details"},
	})
	idx.Add(SearchDocument{
		ID:            "doc3",
		ContentDataID: "cd3",
		DatatypeName:  "article",
		Fields:        map[string]string{"title": "Unrelated", "body": "Nothing matching here"},
	})

	resp := idx.Search("platform", SearchOptions{})

	if resp.Total < 2 {
		t.Fatalf("expected at least 2 results for 'platform', got %d", resp.Total)
	}

	// Results should be ranked by score descending
	for i := 1; i < len(resp.Results); i++ {
		if resp.Results[i].Score > resp.Results[i-1].Score {
			t.Errorf("results not sorted by score: result[%d].Score=%f > result[%d].Score=%f",
				i, resp.Results[i].Score, i-1, resp.Results[i-1].Score)
		}
	}
}

func TestSearchDatatypeFilter(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		DatatypeName:  "page",
		Fields:        map[string]string{"body": "platform documentation guide"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		DatatypeName:  "article",
		Fields:        map[string]string{"body": "platform news and updates"},
	})
	idx.Add(SearchDocument{
		ID:            "doc3",
		ContentDataID: "cd3",
		DatatypeName:  "page",
		Fields:        map[string]string{"body": "platform configuration settings"},
	})

	resp := idx.Search("platform", SearchOptions{DatatypeName: "article"})

	if resp.Total != 1 {
		t.Fatalf("expected 1 result with datatype filter 'article', got %d", resp.Total)
	}
	if resp.Results[0].ContentDataID != "cd2" {
		t.Errorf("expected cd2 (article), got %s", resp.Results[0].ContentDataID)
	}
}

func TestSearchLocaleFilter(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Locale:        "en",
		Fields:        map[string]string{"body": "platform documentation"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		Locale:        "fr",
		Fields:        map[string]string{"body": "platform documentation"},
	})
	idx.Add(SearchDocument{
		ID:            "doc3",
		ContentDataID: "cd3",
		Locale:        "en",
		Fields:        map[string]string{"body": "platform guide"},
	})

	resp := idx.Search("platform", SearchOptions{Locale: "en"})

	for _, r := range resp.Results {
		if r.Locale != "en" {
			t.Errorf("expected only locale 'en' results, got locale %q for %s", r.Locale, r.ContentDataID)
		}
	}
	if resp.Total != 2 {
		t.Errorf("expected 2 results with locale 'en', got %d", resp.Total)
	}
}

func TestSearchWithPrefixPartialTermMatch(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"title": "configuration Management"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		Fields:        map[string]string{"title": "Contact Information"},
	})

	resp := idx.SearchWithPrefix("conf", SearchOptions{})

	if resp.Total == 0 {
		t.Fatal("expected prefix search for 'conf' to return results")
	}

	// Should match "configuration" doc
	foundConfig := false
	for _, r := range resp.Results {
		if r.ContentDataID == "cd1" {
			foundConfig = true
		}
	}
	if !foundConfig {
		t.Error("expected to find cd1 (configuration) via prefix 'conf'")
	}
}

func TestParseQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		query           string
		expectedTerms   []string
		expectedPhrases int
	}{
		{
			name:            "simple terms",
			query:           "hello world",
			expectedTerms:   []string{"hello", "world"},
			expectedPhrases: 0,
		},
		{
			name:            "quoted phrase",
			query:           `"hello world"`,
			expectedTerms:   nil,
			expectedPhrases: 1,
		},
		{
			name:            "stop words removed",
			query:           "the quick brown fox",
			expectedTerms:   []string{"quick", "brown", "fox"},
			expectedPhrases: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pq := ParseQuery(tt.query, defaultStopWords)

			if tt.expectedTerms == nil {
				if len(pq.Terms) != 0 {
					t.Errorf("expected no terms, got %v", pq.Terms)
				}
			} else {
				if len(pq.Terms) != len(tt.expectedTerms) {
					t.Fatalf("expected %d terms, got %d: %v", len(tt.expectedTerms), len(pq.Terms), pq.Terms)
				}
				for i, expected := range tt.expectedTerms {
					if pq.Terms[i] != expected {
						t.Errorf("term[%d] = %q, want %q", i, pq.Terms[i], expected)
					}
				}
			}

			if len(pq.Phrases) != tt.expectedPhrases {
				t.Errorf("expected %d phrases, got %d", tt.expectedPhrases, len(pq.Phrases))
			}
		})
	}
}

// --- 10.8 Adversarial tests ---

func TestAdversarialStopWordsOnlyQuery(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"body": "hello world"},
	})

	// Query with only stop words should return empty, not crash
	resp := idx.Search("the is a an", SearchOptions{})

	if resp.Total != 0 {
		t.Errorf("expected 0 results for stop-words-only query, got %d", resp.Total)
	}
}

func TestAdversarialLongQuery(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"body": "simple test document"},
	})

	// 1000-character query should not crash
	longQuery := strings.Repeat("abcdefghij ", 100) // 1100 chars
	resp := idx.Search(longQuery, SearchOptions{})

	// Should not panic; results may or may not be empty
	_ = resp
}

func TestAdversarialLargeDocument(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	// 100K character document
	largeBody := strings.Repeat("word ", 20000)

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"body": largeBody},
	})

	if idx.Len() != 1 {
		t.Errorf("expected 1 document after adding large doc, got %d", idx.Len())
	}

	resp := idx.Search("word", SearchOptions{})
	if resp.Total != 1 {
		t.Errorf("expected to find the large document, got %d results", resp.Total)
	}
}

func TestAdversarialConcurrentReadsAndWrites(t *testing.T) {
	// NOTE: This test exercises concurrent reads and writes on the index.
	// There is a known data race in ensureSorted() which is called by
	// SearchPrefix under an RLock but mutates sortedTerms/sortedDirty
	// (requires a write lock). This test uses only Search, Len, and Stats
	// (which are race-safe) to avoid triggering that race. SearchPrefix is
	// excluded until the source is fixed.
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	// Pre-populate with some documents
	for i := range 10 {
		idx.Add(SearchDocument{
			ID:            "init-" + strings.Repeat("a", i+1),
			ContentDataID: "cd-init-" + strings.Repeat("a", i+1),
			Fields:        map[string]string{"body": "initial content word" + strings.Repeat("x", i)},
		})
	}

	var wg sync.WaitGroup
	const writers = 5
	const readers = 10
	const iterations = 50

	// Concurrent writers
	for w := range writers {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := range iterations {
				idx.Add(SearchDocument{
					ID:            strings.Repeat("w", writerID+1) + strings.Repeat("i", i+1),
					ContentDataID: strings.Repeat("c", writerID+1) + strings.Repeat("i", i+1),
					Fields:        map[string]string{"body": "concurrent write test content"},
				})
			}
		}(w)
	}

	// Concurrent readers (Len, Stats, Search are all properly locked)
	for r := range readers {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for range iterations {
				_ = idx.Search("content", SearchOptions{})
				_ = idx.Len()
				_ = idx.Stats()
			}
		}(r)
	}

	wg.Wait()

	// Verify index is still functional after concurrent operations
	if idx.Len() == 0 {
		t.Error("index should not be empty after concurrent writes")
	}
}

func TestAdversarialSearchPrefixConcurrent(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())
	for i := 0; i < 10; i++ {
		idx.Add(SearchDocument{
			ID:            fmt.Sprintf("doc-%d", i),
			ContentDataID: fmt.Sprintf("cd-%d", i),
			Fields:        map[string]string{"body": fmt.Sprintf("configuration setup installation doc %d", i)},
		})
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				idx.SearchPrefix("conf")
				idx.SearchPrefix("inst")
			}
		}(i)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				idx.Add(SearchDocument{
					ID:            fmt.Sprintf("w-%d-%d", n, j),
					ContentDataID: fmt.Sprintf("wcd-%d-%d", n, j),
					Fields:        map[string]string{"body": fmt.Sprintf("concurrent write %d %d", n, j)},
				})
			}
		}(i)
	}
	wg.Wait()

	results := idx.SearchPrefix("conf")
	if len(results) == 0 {
		t.Error("expected SearchPrefix results after concurrent operations")
	}
}

func TestAdversarialEmptyIndexSearch(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	resp := idx.Search("anything", SearchOptions{})

	if resp.Total != 0 {
		t.Errorf("expected 0 results on empty index, got %d", resp.Total)
	}
	if resp.Results != nil && len(resp.Results) != 0 {
		t.Errorf("expected nil or empty results on empty index, got %d items", len(resp.Results))
	}
}

func TestAdversarialXSSQuery(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"body": "normal content"},
	})

	// XSS-style query should be tokenized as plain text, not interpreted as code
	resp := idx.Search("<script>alert('xss')</script>", SearchOptions{})

	// The query should be tokenized into words: "script", "alert", "xss", "script"
	// It should not cause any injection or panic
	_ = resp

	// Also test XSS as document content
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		Fields:        map[string]string{"body": "<script>alert('xss')</script>"},
	})

	resp2 := idx.Search("alert", SearchOptions{})
	if resp2.Total == 0 {
		t.Error("expected to find tokenized 'alert' from XSS-style content")
	}

	// Snippets should not contain raw HTML tags
	for _, r := range resp2.Results {
		if strings.Contains(r.Snippet, "<script>") {
			t.Errorf("snippet should not contain raw HTML tags: %q", r.Snippet)
		}
	}
}

func TestAdversarialEmptyQuery(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"body": "hello world"},
	})

	resp := idx.Search("", SearchOptions{})
	if resp.Total != 0 {
		t.Errorf("expected 0 results for empty query, got %d", resp.Total)
	}
}

func TestAdversarialSpecialCharactersQuery(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"body": "hello world"},
	})

	// Queries with only special characters should not crash
	specialQueries := []string{
		"!@#$%^&*()",
		"...",
		"---",
		"\"\"",
		"   ",
		"\t\n\r",
	}

	for _, q := range specialQueries {
		resp := idx.Search(q, SearchOptions{})
		_ = resp // should not panic
	}
}

func TestSearchPagination(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	// Add many documents with the same searchable term
	for i := range 25 {
		idx.Add(SearchDocument{
			ID:            strings.Repeat("d", i+1),
			ContentDataID: strings.Repeat("c", i+1),
			Fields:        map[string]string{"body": "searchable content item"},
		})
	}

	// First page
	resp1 := idx.Search("searchable", SearchOptions{Limit: 10, Offset: 0})
	if len(resp1.Results) != 10 {
		t.Errorf("page 1 results = %d, want 10", len(resp1.Results))
	}
	if resp1.Total != 25 {
		t.Errorf("total = %d, want 25", resp1.Total)
	}

	// Second page
	resp2 := idx.Search("searchable", SearchOptions{Limit: 10, Offset: 10})
	if len(resp2.Results) != 10 {
		t.Errorf("page 2 results = %d, want 10", len(resp2.Results))
	}

	// Third page (partial)
	resp3 := idx.Search("searchable", SearchOptions{Limit: 10, Offset: 20})
	if len(resp3.Results) != 5 {
		t.Errorf("page 3 results = %d, want 5", len(resp3.Results))
	}

	// Offset beyond total
	resp4 := idx.Search("searchable", SearchOptions{Limit: 10, Offset: 100})
	if len(resp4.Results) != 0 {
		t.Errorf("offset beyond total results = %d, want 0", len(resp4.Results))
	}
}
