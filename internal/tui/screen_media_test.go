package tui

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
)

func TestNewMediaScreen_InitialState(t *testing.T) {
	items := []db.Media{
		makeMedia("https://cdn.example.com/images/a.jpg"),
		makeMedia("https://cdn.example.com/images/b.jpg"),
		makeMedia("https://cdn.example.com/docs/c.pdf"),
	}
	s := NewMediaScreen(items)

	if s.Cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", s.Cursor)
	}
	if s.Searching {
		t.Error("expected Searching=false")
	}
	if s.SearchQuery != "" {
		t.Errorf("expected empty SearchQuery, got %q", s.SearchQuery)
	}
	if len(s.MediaList) != 3 {
		t.Errorf("expected 3 items in MediaList, got %d", len(s.MediaList))
	}
	if len(s.FlatList) == 0 {
		t.Error("expected non-empty FlatList after tree build")
	}
	if s.PageIndex() != MEDIA {
		t.Errorf("expected PageIndex MEDIA, got %d", s.PageIndex())
	}
}

func TestNewMediaScreen_NilList(t *testing.T) {
	s := NewMediaScreen(nil)
	if len(s.MediaList) != 0 {
		t.Errorf("expected empty MediaList for nil input, got %d", len(s.MediaList))
	}
	if len(s.FlatList) != 0 {
		t.Errorf("expected empty FlatList, got %d", len(s.FlatList))
	}
}

func TestMediaScreen_FilterRebuildsTree(t *testing.T) {
	items := []db.Media{
		makeMediaWithName("https://cdn.example.com/a.jpg", "alpha", ""),
		makeMediaWithName("https://cdn.example.com/b.png", "beta", ""),
		makeMediaWithName("https://cdn.example.com/c.gif", "gamma", ""),
	}
	s := NewMediaScreen(items)
	originalLen := len(s.FlatList)

	// Apply filter
	s.FilteredList = FilterMediaList(s.MediaList, "alpha")
	s.Cursor = 0
	s.rebuildTree()

	if len(s.FlatList) >= originalLen {
		t.Errorf("expected fewer nodes after filter, got %d (was %d)", len(s.FlatList), originalLen)
	}
}

func TestMediaScreen_CursorBoundsAfterFilter(t *testing.T) {
	items := []db.Media{
		makeMediaWithName("https://cdn.example.com/a.jpg", "alpha", ""),
		makeMediaWithName("https://cdn.example.com/b.png", "beta", ""),
		makeMediaWithName("https://cdn.example.com/c.gif", "gamma", ""),
	}
	s := NewMediaScreen(items)

	// Set cursor to last item
	s.Cursor = len(s.FlatList) - 1

	// Filter to single item — cursor should be clamped
	s.FilteredList = FilterMediaList(s.MediaList, "alpha")
	s.rebuildTree()

	if s.Cursor >= len(s.FlatList) {
		t.Errorf("cursor %d out of bounds for FlatList len %d", s.Cursor, len(s.FlatList))
	}
}

func TestMediaScreen_ExpandCollapse(t *testing.T) {
	items := []db.Media{
		makeMedia("https://cdn.example.com/images/a.jpg"),
		makeMedia("https://cdn.example.com/images/b.jpg"),
	}
	s := NewMediaScreen(items)

	// Should start with folder expanded: 1 folder + 2 files = 3
	if len(s.FlatList) != 3 {
		t.Fatalf("expected 3 flat nodes, got %d", len(s.FlatList))
	}

	// Find the folder node and collapse it
	for _, node := range s.FlatList {
		if node.Kind == MediaNodeFolder {
			node.Expand = false
			break
		}
	}
	s.FlatList = FlattenMediaTree(s.MediaTree)
	s.CursorMax = len(s.FlatList) - 1

	if len(s.FlatList) != 1 {
		t.Fatalf("after collapse expected 1 node, got %d", len(s.FlatList))
	}

	// Re-expand
	s.FlatList[0].Expand = true
	s.FlatList = FlattenMediaTree(s.MediaTree)

	if len(s.FlatList) != 3 {
		t.Fatalf("after re-expand expected 3 nodes, got %d", len(s.FlatList))
	}
}

func TestMediaScreen_SelectedMedia(t *testing.T) {
	items := []db.Media{
		makeMedia("https://cdn.example.com/images/a.jpg"),
		makeMedia("https://cdn.example.com/images/b.jpg"),
	}
	s := NewMediaScreen(items)

	// Cursor 0 is the folder — selectedMedia should return nil
	if s.FlatList[0].Kind == MediaNodeFolder {
		media := s.selectedMedia()
		if media != nil {
			t.Error("expected nil selectedMedia when cursor is on folder")
		}
	}

	// Move cursor to a file node
	for i, node := range s.FlatList {
		if node.Kind == MediaNodeFile {
			s.Cursor = i
			break
		}
	}
	media := s.selectedMedia()
	if media == nil {
		t.Error("expected non-nil selectedMedia when cursor is on file")
	}
}
