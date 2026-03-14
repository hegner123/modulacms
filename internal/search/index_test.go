package search

import (
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
