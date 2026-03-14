package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	idx := NewIndex(DefaultConfig())

	idx.Add(SearchDocument{
		ID:            "doc1",
		ContentDataID: "cd1",
		RouteSlug:     "/getting-started",
		DatatypeName:  "page",
		Fields:        map[string]string{"title": "Getting Started Guide", "body": "Welcome to the platform documentation"},
	})
	idx.Add(SearchDocument{
		ID:            "doc2",
		ContentDataID: "cd2",
		RouteSlug:     "/config",
		DatatypeName:  "page",
		Fields:        map[string]string{"title": "Advanced Configuration", "body": "Configure your application settings"},
	})
	idx.Add(SearchDocument{
		ID:            "doc3",
		ContentDataID: "cd3",
		RouteSlug:     "/troubleshooting",
		DatatypeName:  "article",
		Fields:        map[string]string{"title": "Troubleshooting", "description": "Common problems and solutions"},
	})

	origStats := idx.Stats()

	path := filepath.Join(t.TempDir(), "test.idx")
	if err := idx.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path, DefaultConfig())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if got := loaded.Len(); got != idx.Len() {
		t.Errorf("loaded Len() = %d, want %d", got, idx.Len())
	}

	loadedStats := loaded.Stats()
	if loadedStats.Documents != origStats.Documents {
		t.Errorf("loaded Stats().Documents = %d, want %d", loadedStats.Documents, origStats.Documents)
	}
	if loadedStats.Terms != origStats.Terms {
		t.Errorf("loaded Stats().Terms = %d, want %d", loadedStats.Terms, origStats.Terms)
	}
	if loadedStats.Postings != origStats.Postings {
		t.Errorf("loaded Stats().Postings = %d, want %d", loadedStats.Postings, origStats.Postings)
	}
	if loadedStats.Fields != origStats.Fields {
		t.Errorf("loaded Stats().Fields = %d, want %d", loadedStats.Fields, origStats.Fields)
	}

	// Verify postings survived the round-trip by checking a known term
	loaded.mu.RLock()
	postings, hasGuide := loaded.postings["guid"]
	loaded.mu.RUnlock()
	if !hasGuide {
		// Try the stemmed or tokenized form; check what actually got indexed
		loaded.mu.RLock()
		// Look for any term from "Getting Started Guide"
		found := false
		for term := range loaded.postings {
			if term == "get" || term == "start" || term == "guide" || term == "guid" {
				found = true
				break
			}
		}
		loaded.mu.RUnlock()
		if !found {
			t.Error("loaded index postings missing expected terms from indexed documents")
		}
	} else if len(postings) == 0 {
		t.Error("loaded index has empty postings for expected term")
	}
}

func TestLoadCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.idx")
	if err := os.WriteFile(path, []byte("this is garbage data not an index"), 0644); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	_, err := Load(path, DefaultConfig())
	if err == nil {
		t.Fatal("Load of corrupt file should return error")
	}
}

func TestLoadWrongMagic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wrongmagic.idx")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	// Write wrong magic bytes
	f.Write([]byte("XXXX"))
	// Write some padding so the file is not trivially empty
	f.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	f.Close()

	_, err = Load(path, DefaultConfig())
	if err == nil {
		t.Fatal("Load with wrong magic bytes should return error")
	}
	if !strings.Contains(err.Error(), "magic") {
		t.Errorf("error should mention 'magic', got: %v", err)
	}
}

func TestSaveLoadEmpty(t *testing.T) {
	idx := NewIndex(DefaultConfig())

	path := filepath.Join(t.TempDir(), "empty.idx")
	if err := idx.Save(path); err != nil {
		t.Fatalf("Save empty index failed: %v", err)
	}

	loaded, err := Load(path, DefaultConfig())
	if err != nil {
		t.Fatalf("Load empty index failed: %v", err)
	}

	if got := loaded.Len(); got != 0 {
		t.Errorf("loaded empty index Len() = %d, want 0", got)
	}

	stats := loaded.Stats()
	if stats.Documents != 0 {
		t.Errorf("loaded empty Stats().Documents = %d, want 0", stats.Documents)
	}
	if stats.Terms != 0 {
		t.Errorf("loaded empty Stats().Terms = %d, want 0", stats.Terms)
	}
}
