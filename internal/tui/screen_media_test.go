package tui

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

func TestNewMediaScreen_InitialState(t *testing.T) {
	items := []db.Media{
		makeMedia("https://cdn.example.com/images/a.jpg"),
		makeMedia("https://cdn.example.com/images/b.jpg"),
		makeMedia("https://cdn.example.com/docs/c.pdf"),
	}
	s := NewMediaScreen(items, nil)

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
	// All items are unfiled (no folder IDs), so they appear at root
	if len(s.FlatList) == 0 {
		t.Error("expected non-empty FlatList after tree build")
	}
	if s.PageIndex() != MEDIA {
		t.Errorf("expected PageIndex MEDIA, got %d", s.PageIndex())
	}
}

func TestNewMediaScreen_NilList(t *testing.T) {
	s := NewMediaScreen(nil, nil)
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
	s := NewMediaScreen(items, nil)
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
	s := NewMediaScreen(items, nil)

	// Set cursor to last item
	s.Cursor = len(s.FlatList) - 1

	// Filter to single item -- cursor should be clamped
	s.FilteredList = FilterMediaList(s.MediaList, "alpha")
	s.rebuildTree()

	if s.Cursor >= len(s.FlatList) {
		t.Errorf("cursor %d out of bounds for FlatList len %d", s.Cursor, len(s.FlatList))
	}
}

func TestMediaScreen_FolderExpandCollapse(t *testing.T) {
	folderID := types.NewMediaFolderID()
	folders := []db.MediaFolder{
		{FolderID: folderID, Name: "images"},
	}
	items := []db.Media{
		makeMediaInFolder("a.jpg", folderID),
		makeMediaInFolder("b.jpg", folderID),
	}
	s := NewMediaScreen(items, folders)

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
	folderID := types.NewMediaFolderID()
	folders := []db.MediaFolder{
		{FolderID: folderID, Name: "images"},
	}
	items := []db.Media{
		makeMediaInFolder("a.jpg", folderID),
		makeMediaInFolder("b.jpg", folderID),
	}
	s := NewMediaScreen(items, folders)

	// Cursor 0 is the folder -- selectedMedia should return nil
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

func TestMediaScreen_UnfiledMediaAtRoot(t *testing.T) {
	// Media items without folders should appear at root level
	items := []db.Media{
		makeMedia("https://cdn.example.com/a.jpg"),
		makeMedia("https://cdn.example.com/b.jpg"),
	}
	s := NewMediaScreen(items, nil)

	// All items should be at root (no folders)
	if len(s.FlatList) != 2 {
		t.Errorf("expected 2 root items, got %d", len(s.FlatList))
	}
	for _, node := range s.FlatList {
		if node.Kind != MediaNodeFile {
			t.Error("expected all root nodes to be files when no folders exist")
		}
	}
}

func TestMediaScreen_NestedFolders(t *testing.T) {
	parentID := types.NewMediaFolderID()
	childID := types.NewMediaFolderID()
	folders := []db.MediaFolder{
		{FolderID: parentID, Name: "parent"},
		{FolderID: childID, Name: "child", ParentID: types.NullableMediaFolderID{ID: parentID, Valid: true}},
	}
	items := []db.Media{
		makeMediaInFolder("file.txt", childID),
	}
	s := NewMediaScreen(items, folders)

	// Should have: parent folder -> child folder -> file = 3 nodes
	if len(s.FlatList) != 3 {
		t.Errorf("expected 3 flat nodes for nested folders, got %d", len(s.FlatList))
	}

	// Verify depth: parent=0, child=1, file=2
	if s.FlatList[0].Depth != 0 {
		t.Errorf("expected parent depth 0, got %d", s.FlatList[0].Depth)
	}
	if s.FlatList[1].Depth != 1 {
		t.Errorf("expected child depth 1, got %d", s.FlatList[1].Depth)
	}
	if s.FlatList[2].Depth != 2 {
		t.Errorf("expected file depth 2, got %d", s.FlatList[2].Depth)
	}
}

// TestMediaScreen_ViewRendersAllStructFields verifies that every field from
// db.Media is represented in the combined summary + metadata panels.
// If a field is added to the struct but not rendered, this test fails.
func TestMediaScreen_ViewRendersAllStructFields(t *testing.T) {
	folderID := types.NewMediaFolderID()
	folders := []db.MediaFolder{
		{FolderID: folderID, Name: "test-folder"},
	}

	// Populate EVERY field so conditional renders trigger
	media := db.Media{
		MediaID:      types.NewMediaID(),
		Name:         db.NullString{NullString: sql.NullString{String: "test.jpg", Valid: true}},
		DisplayName:  db.NullString{NullString: sql.NullString{String: "Test Image", Valid: true}},
		Alt:          db.NullString{NullString: sql.NullString{String: "alt text", Valid: true}},
		Caption:      db.NullString{NullString: sql.NullString{String: "caption text", Valid: true}},
		Description:  db.NullString{NullString: sql.NullString{String: "description text", Valid: true}},
		Class:        db.NullString{NullString: sql.NullString{String: "image", Valid: true}},
		Mimetype:     db.NullString{NullString: sql.NullString{String: "image/jpeg", Valid: true}},
		Dimensions:   db.NullString{NullString: sql.NullString{String: "1920x1080", Valid: true}},
		URL:          types.URL("https://cdn.example.com/test.jpg"),
		Srcset:       db.NullString{NullString: sql.NullString{String: "https://cdn.example.com/test-sm.jpg 480w", Valid: true}},
		FocalX:       types.NullableFloat64{Float64: 0.5, Valid: true},
		FocalY:       types.NullableFloat64{Float64: 0.5, Valid: true},
		AuthorID:     types.NullableUserID{ID: types.UserID("01TESTAUTHOR0000000000000"), Valid: true},
		FolderID:     types.NullableMediaFolderID{ID: folderID, Valid: true},
		DateCreated:  types.TimestampNow(),
		DateModified: types.TimestampNow(),
	}

	s := NewMediaScreen([]db.Media{media}, folders)
	// Move cursor to the file node (may be under a folder)
	for i, node := range s.FlatList {
		if node.Kind == MediaNodeFile {
			s.Cursor = i
			break
		}
	}

	summary := s.renderMediaSummary()
	metadata := s.renderMediaMetadata()
	combined := summary + "\n" + metadata

	// Map each struct field to what we expect in the output.
	// Key: json tag from db.Media, Value: the string that should appear.
	fieldExpectations := map[string]string{
		"media_id":      media.MediaID.String(),
		"name":          "test.jpg",
		"display_name":  "Test Image",
		"alt":           "alt text",
		"caption":       "caption text",
		"description":   "description text",
		"class":         "image",
		"mimetype":      "image/jpeg",
		"dimensions":    "1920x1080",
		"url":           "https://cdn.example.com/test.jpg",
		"srcset":        "https://cdn.example.com/test-sm.jpg 480w",
		"focal_x":       "0.50",
		"focal_y":       "0.50",
		"author_id":     "01TESTAUTHOR0000000000000",
		"folder_id":     folderID.String(),
		"date_created":  media.DateCreated.String(),
		"date_modified": media.DateModified.String(),
	}

	// Cross-check: use reflection to ensure we're testing every field
	rt := reflect.TypeOf(db.Media{})
	for i := range rt.NumField() {
		field := rt.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		if _, ok := fieldExpectations[jsonTag]; !ok {
			t.Errorf("db.Media has field %q (json:%q) with no entry in fieldExpectations — add it to the test and the view", field.Name, jsonTag)
		}
	}

	// Verify each expected value appears in the rendered output
	for jsonTag, expected := range fieldExpectations {
		if !strings.Contains(combined, expected) {
			t.Errorf("field %q: expected value %q not found in rendered view", jsonTag, expected)
		}
	}
}

// makeMediaInFolder creates a test media item assigned to a folder.
func makeMediaInFolder(name string, folderID types.MediaFolderID) db.Media {
	m := makeMediaWithName("https://cdn.example.com/"+name, name, "")
	m.FolderID = types.NullableMediaFolderID{ID: folderID, Valid: true}
	return m
}
