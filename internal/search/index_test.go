package search

import (
	"strings"
	"testing"
)

func TestIndexAddAndSearch(t *testing.T) {
	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"title": "Getting Started Guide", "body": "Welcome to the platform documentation"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		Fields:        map[string]string{"title": "Advanced Configuration", "body": "Configure your application settings"},
	})
	idx.Add(SearchDocument{
		ID:            "doc3",
		ContentDataID: "cd3",
		Fields:        map[string]string{"title": "Troubleshooting", "body": "Common problems and solutions"},
	})

	if got := idx.Len(); got != 3 {
		t.Errorf("Len() = %d, want 3", got)
	}

	stats := idx.Stats()
	if stats.Documents != 3 {
		t.Errorf("Stats().Documents = %d, want 3", stats.Documents)
	}
}

func TestIndexRemove(t *testing.T) {
	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"title": "First document"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		Fields:        map[string]string{"title": "Second document"},
	})
	idx.Add(SearchDocument{
		ID:            "doc3",
		ContentDataID: "cd3",
		Fields:        map[string]string{"title": "Third document"},
	})

	idx.RemoveByContentID("cd2")

	if got := idx.Len(); got != 2 {
		t.Errorf("Len() after remove = %d, want 2", got)
	}

	idx.mu.RLock()
	_, found := idx.docsByContentID["cd2"]
	idx.mu.RUnlock()

	if found {
		t.Error("docsByContentID still contains removed content ID")
	}
}

func TestIndexFieldRegistration(t *testing.T) {
	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"title": "Hello", "body": "World"},
	})

	idx.mu.RLock()
	fieldCount := len(idx.fieldNames)
	idx.mu.RUnlock()

	if fieldCount != 2 {
		t.Errorf("fieldNames count after first doc = %d, want 2", fieldCount)
	}

	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		Fields:        map[string]string{"description": "Extra field"},
	})

	idx.mu.RLock()
	fieldCount = len(idx.fieldNames)
	idx.mu.RUnlock()

	if fieldCount != 3 {
		t.Errorf("fieldNames count after second doc = %d, want 3", fieldCount)
	}
}

func TestIndexPostings(t *testing.T) {
	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"body": "the quick brown fox"},
	})

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	quickPostings, hasQuick := idx.postings["quick"]
	if !hasQuick {
		t.Fatal("postings missing entry for 'quick'")
	}
	if len(quickPostings) != 1 {
		t.Errorf("postings['quick'] has %d entries, want 1", len(quickPostings))
	}

	_, hasThe := idx.postings["the"]
	if hasThe {
		t.Error("postings contains 'the' which should be filtered as a stop word")
	}
}

func TestIndexStats(t *testing.T) {
	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"title": "Alpha bravo", "body": "Charlie delta echo"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		Fields:        map[string]string{"title": "Foxtrot golf", "body": "Hotel india juliet"},
	})

	stats := idx.Stats()

	if stats.Documents <= 0 {
		t.Errorf("Stats().Documents = %d, want > 0", stats.Documents)
	}
	if stats.Terms <= 0 {
		t.Errorf("Stats().Terms = %d, want > 0", stats.Terms)
	}
	if stats.Postings <= 0 {
		t.Errorf("Stats().Postings = %d, want > 0", stats.Postings)
	}
	if stats.Fields <= 0 {
		t.Errorf("Stats().Fields = %d, want > 0", stats.Fields)
	}
	if stats.MemEstimate <= 0 {
		t.Errorf("Stats().MemEstimate = %d, want > 0", stats.MemEstimate)
	}
}

func TestIndexSearchReturnsCorrectDoc(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		DatatypeName:  "page",
		Fields:        map[string]string{"title": "Getting Started Guide", "body": "Welcome to the platform"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		DatatypeName:  "page",
		Fields:        map[string]string{"title": "Advanced Configuration", "body": "Configure your application settings"},
	})
	idx.Add(SearchDocument{
		ID:            "doc3",
		ContentDataID: "cd3",
		DatatypeName:  "article",
		Fields:        map[string]string{"title": "Troubleshooting", "body": "Common problems and solutions"},
	})

	// Search for "configuration" which appears only in doc2
	resp := idx.Search("configuration", SearchOptions{})

	if resp.Total == 0 {
		t.Fatal("expected at least one result for 'configuration'")
	}
	if resp.Results[0].ContentDataID != "cd2" {
		t.Errorf("expected first result to be cd2, got %s", resp.Results[0].ContentDataID)
	}
}

func TestIndexSearchPrefixMatch(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"title": "installation Guide"},
	})

	// "install" should match "installation" via prefix search
	resp := idx.SearchWithPrefix("install", SearchOptions{})

	if resp.Total == 0 {
		t.Fatal("expected prefix search for 'install' to match 'installation'")
	}
	if resp.Results[0].ContentDataID != "cd1" {
		t.Errorf("expected first result to be cd1, got %s", resp.Results[0].ContentDataID)
	}
}

func TestIndexFieldWeightTitleRanksHigher(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.FieldWeights["title"] = 3.0
	cfg.FieldWeights["body"] = 1.0
	idx := NewIndex(cfg)

	idx.Add(SearchDocument{
		ID:            "title-match",
		ContentDataID: "cd-title",
		Fields:        map[string]string{"title": "deployment strategies", "body": "unrelated content here"},
	})
	idx.Add(SearchDocument{
		ID:            "body-match",
		ContentDataID: "cd-body",
		Fields:        map[string]string{"title": "unrelated heading", "body": "deployment strategies explained"},
	})

	resp := idx.Search("deployment", SearchOptions{})

	if resp.Total < 2 {
		t.Fatalf("expected 2 results, got %d", resp.Total)
	}
	// Title match should rank higher due to field weight 3.0 vs 1.0
	if resp.Results[0].ContentDataID != "cd-title" {
		t.Errorf("expected title-match (cd-title) to rank first, got %s", resp.Results[0].ContentDataID)
	}
}

func TestIndexSearchNoResults(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"title": "Hello World"},
	})

	resp := idx.Search("nonexistent", SearchOptions{})

	if resp.Total != 0 {
		t.Errorf("expected 0 results for missing term, got %d", resp.Total)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected empty results slice, got %d items", len(resp.Results))
	}
}

func TestIndexRemoveByContentIDSearchVerification(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"title": "Alpha bravo charlie"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		Fields:        map[string]string{"title": "Delta echo foxtrot"},
	})
	idx.Add(SearchDocument{
		ID:            "doc3",
		ContentDataID: "cd3",
		Fields:        map[string]string{"title": "Golf hotel india"},
	})

	// Verify doc2 is findable before removal
	resp := idx.Search("echo", SearchOptions{})
	if resp.Total == 0 {
		t.Fatal("expected to find 'echo' before removal")
	}

	idx.RemoveByContentID("cd2")

	// Verify doc2 is no longer findable
	resp = idx.Search("echo", SearchOptions{})
	if resp.Total != 0 {
		t.Errorf("expected 0 results after removing cd2, got %d", resp.Total)
	}
	for _, r := range resp.Results {
		if r.ContentDataID == "cd2" {
			t.Error("removed document cd2 still appears in search results")
		}
	}

	// Verify other docs still findable
	resp = idx.Search("alpha", SearchOptions{})
	if resp.Total == 0 {
		t.Error("expected cd1 (alpha) to still be searchable after removing cd2")
	}
}

func TestIndexSearchPrefix(t *testing.T) {
	t.Parallel()

	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		Fields:        map[string]string{"body": "configuration management system"},
	})

	matches := idx.SearchPrefix("config")

	if len(matches) == 0 {
		t.Fatal("expected prefix matches for 'config'")
	}
	foundConfiguration := false
	for _, m := range matches {
		if strings.HasPrefix(m, "config") {
			foundConfiguration = true
		}
	}
	if !foundConfiguration {
		t.Errorf("expected a term starting with 'config' in prefix results, got %v", matches)
	}
}
