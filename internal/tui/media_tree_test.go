package tui

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

func makeMedia(url string) db.Media {
	return db.Media{
		MediaID: types.NewMediaID(),
		URL:     types.URL(url),
		Name:    db.NewNullString(url),
	}
}

func makeMediaWithName(url, name, displayName string) db.Media {
	return db.Media{
		MediaID:     types.NewMediaID(),
		URL:         types.URL(url),
		Name:        db.NewNullString(name),
		DisplayName: db.NewNullString(displayName),
	}
}

// =============================================================================
// DB-BACKED FOLDER TREE TESTS
// =============================================================================

func TestBuildMediaTree_EmptyBoth(t *testing.T) {
	result := BuildMediaTree(nil, nil)
	if result != nil {
		t.Fatalf("expected nil for empty input, got %d nodes", len(result))
	}
}

func TestBuildMediaTree_UnfiledMediaAtRoot(t *testing.T) {
	items := []db.Media{makeMedia("https://example.com/photo.jpg")}
	roots := BuildMediaTree(nil, items)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].Kind != MediaNodeFile {
		t.Errorf("expected MediaNodeFile, got %d", roots[0].Kind)
	}
	if roots[0].Media == nil {
		t.Error("expected non-nil Media pointer")
	}
}

func TestBuildMediaTree_FolderWithMedia(t *testing.T) {
	folderID := types.NewMediaFolderID()
	folders := []db.MediaFolder{
		{FolderID: folderID, Name: "images"},
	}
	items := []db.Media{
		makeMediaInFolderForTree("a.jpg", folderID),
		makeMediaInFolderForTree("b.jpg", folderID),
	}
	roots := BuildMediaTree(folders, items)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root folder, got %d", len(roots))
	}
	root := roots[0]
	if root.Kind != MediaNodeFolder {
		t.Fatalf("expected folder at root, got kind %d", root.Kind)
	}
	if root.Label != "images" {
		t.Errorf("expected label 'images', got %q", root.Label)
	}
	if root.FolderID != folderID {
		t.Errorf("expected FolderID %s, got %s", folderID, root.FolderID)
	}

	// Folder should have 2 file children
	var children int
	child := root.FirstChild
	for child != nil {
		if child.Kind != MediaNodeFile {
			t.Errorf("expected file child, got kind %d", child.Kind)
		}
		children++
		child = child.NextSibling
	}
	if children != 2 {
		t.Errorf("expected 2 children, got %d", children)
	}
}

func TestBuildMediaTree_NestedFolders(t *testing.T) {
	parentID := types.NewMediaFolderID()
	childID := types.NewMediaFolderID()
	grandchildID := types.NewMediaFolderID()

	folders := []db.MediaFolder{
		{FolderID: parentID, Name: "a"},
		{FolderID: childID, Name: "b", ParentID: types.NullableMediaFolderID{ID: parentID, Valid: true}},
		{FolderID: grandchildID, Name: "c", ParentID: types.NullableMediaFolderID{ID: childID, Valid: true}},
	}
	items := []db.Media{
		makeMediaInFolderForTree("file.jpg", grandchildID),
	}
	roots := BuildMediaTree(folders, items)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	// Should be: a/ -> b/ -> c/ -> file.jpg
	a := roots[0]
	if a.Kind != MediaNodeFolder || a.Label != "a" {
		t.Fatalf("expected folder 'a', got kind=%d label=%q", a.Kind, a.Label)
	}
	b := a.FirstChild
	if b == nil || b.Kind != MediaNodeFolder || b.Label != "b" {
		t.Fatalf("expected folder 'b' as child of 'a'")
	}
	c := b.FirstChild
	if c == nil || c.Kind != MediaNodeFolder || c.Label != "c" {
		t.Fatalf("expected folder 'c' as child of 'b'")
	}
	f := c.FirstChild
	if f == nil || f.Kind != MediaNodeFile {
		t.Fatalf("expected file as child of 'c'")
	}
}

func TestBuildMediaTree_MixedFolderAndUnfiled(t *testing.T) {
	folderID := types.NewMediaFolderID()
	folders := []db.MediaFolder{
		{FolderID: folderID, Name: "images"},
	}
	items := []db.Media{
		makeMedia("https://cdn.example.com/root.jpg"),
		makeMediaInFolderForTree("nested.jpg", folderID),
	}
	roots := BuildMediaTree(folders, items)

	// Should have: images/ folder + root.jpg file at root = 2
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}

	// Verify we have one folder and one file at root level
	var folderCount, fileCount int
	for _, r := range roots {
		if r.Kind == MediaNodeFolder {
			folderCount++
		}
		if r.Kind == MediaNodeFile {
			fileCount++
		}
	}
	if folderCount != 1 {
		t.Errorf("expected 1 folder at root, got %d", folderCount)
	}
	if fileCount != 1 {
		t.Errorf("expected 1 file at root, got %d", fileCount)
	}
}

func TestBuildMediaTree_FoldersSortedAlphabetically(t *testing.T) {
	id1 := types.NewMediaFolderID()
	id2 := types.NewMediaFolderID()
	id3 := types.NewMediaFolderID()
	folders := []db.MediaFolder{
		{FolderID: id3, Name: "zebra"},
		{FolderID: id1, Name: "alpha"},
		{FolderID: id2, Name: "middle"},
	}
	roots := BuildMediaTree(folders, nil)

	if len(roots) != 3 {
		t.Fatalf("expected 3 roots, got %d", len(roots))
	}
	if roots[0].Label != "alpha" {
		t.Errorf("expected first folder 'alpha', got %q", roots[0].Label)
	}
	if roots[1].Label != "middle" {
		t.Errorf("expected second folder 'middle', got %q", roots[1].Label)
	}
	if roots[2].Label != "zebra" {
		t.Errorf("expected third folder 'zebra', got %q", roots[2].Label)
	}
}

// =============================================================================
// LEGACY URL-BASED TREE TESTS
// =============================================================================

func TestBuildMediaTreeLegacy_Empty(t *testing.T) {
	result := BuildMediaTreeLegacy(nil)
	if result != nil {
		t.Fatalf("expected nil for empty input, got %d nodes", len(result))
	}
}

func TestBuildMediaTreeLegacy_SingleFile(t *testing.T) {
	items := []db.Media{makeMedia("https://example.com/photo.jpg")}
	roots := BuildMediaTreeLegacy(items)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].Kind != MediaNodeFile {
		t.Errorf("expected MediaNodeFile, got %d", roots[0].Kind)
	}
	if roots[0].Label != "photo.jpg" {
		t.Errorf("expected label 'photo.jpg', got %q", roots[0].Label)
	}
}

func TestBuildMediaTreeLegacy_SharedPrefix(t *testing.T) {
	items := []db.Media{
		makeMedia("https://cdn.example.com/images/a.jpg"),
		makeMedia("https://cdn.example.com/images/b.jpg"),
	}
	roots := BuildMediaTreeLegacy(items)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root folder, got %d roots", len(roots))
	}
	root := roots[0]
	if root.Kind != MediaNodeFolder {
		t.Fatalf("expected folder at root, got kind %d", root.Kind)
	}
	if root.Label != "images" {
		t.Errorf("expected label 'images', got %q", root.Label)
	}
}

// =============================================================================
// FLATTEN TESTS
// =============================================================================

func TestFlattenMediaTree_RespectsExpand(t *testing.T) {
	folderID := types.NewMediaFolderID()
	folders := []db.MediaFolder{
		{FolderID: folderID, Name: "images"},
	}
	items := []db.Media{
		makeMediaInFolderForTree("a.jpg", folderID),
		makeMediaInFolderForTree("b.jpg", folderID),
	}
	roots := BuildMediaTree(folders, items)

	// With expand=true (default): folder + 2 files = 3
	flat := FlattenMediaTree(roots)
	if len(flat) != 3 {
		t.Fatalf("expanded: expected 3 nodes, got %d", len(flat))
	}

	// Collapse the folder
	roots[0].Expand = false
	flat = FlattenMediaTree(roots)
	if len(flat) != 1 {
		t.Fatalf("collapsed: expected 1 node (folder only), got %d", len(flat))
	}
	if flat[0].Kind != MediaNodeFolder {
		t.Error("collapsed: expected folder node")
	}
}

// =============================================================================
// FILTER TESTS
// =============================================================================

func TestFilterMediaList_MatchesName(t *testing.T) {
	items := []db.Media{
		makeMediaWithName("https://cdn.example.com/a.jpg", "alpha", "Alpha Image"),
		makeMediaWithName("https://cdn.example.com/b.png", "beta", "Beta Photo"),
	}

	result := FilterMediaList(items, "alpha")
	if len(result) != 1 {
		t.Fatalf("expected 1 match for 'alpha', got %d", len(result))
	}
	if string(result[0].URL) != "https://cdn.example.com/a.jpg" {
		t.Errorf("wrong match: %s", result[0].URL)
	}
}

func TestFilterMediaList_CaseInsensitive(t *testing.T) {
	items := []db.Media{
		makeMediaWithName("https://cdn.example.com/a.jpg", "Photo", ""),
	}
	result := FilterMediaList(items, "PHOTO")
	if len(result) != 1 {
		t.Fatalf("expected case-insensitive match, got %d results", len(result))
	}
}

func TestFilterMediaList_MatchesMimetype(t *testing.T) {
	items := []db.Media{
		{
			MediaID:  types.NewMediaID(),
			URL:      types.URL("https://cdn.example.com/a.jpg"),
			Mimetype: db.NewNullString("image/jpeg"),
		},
		{
			MediaID:  types.NewMediaID(),
			URL:      types.URL("https://cdn.example.com/b.mp4"),
			Mimetype: db.NewNullString("video/mp4"),
		},
	}
	result := FilterMediaList(items, "video")
	if len(result) != 1 {
		t.Fatalf("expected 1 match for 'video', got %d", len(result))
	}
}

func TestFilterMediaList_MatchesURL(t *testing.T) {
	items := []db.Media{
		makeMedia("https://cdn.example.com/uploads/photos/sunset.jpg"),
		makeMedia("https://cdn.example.com/uploads/docs/readme.pdf"),
	}
	result := FilterMediaList(items, "photos")
	if len(result) != 1 {
		t.Fatalf("expected 1 match for 'photos' in URL, got %d", len(result))
	}
}

func TestFilterMediaList_EmptyQuery(t *testing.T) {
	items := []db.Media{makeMedia("a.jpg"), makeMedia("b.jpg")}
	result := FilterMediaList(items, "")
	if len(result) != 2 {
		t.Fatalf("empty query should return all items, got %d", len(result))
	}
}

func TestFilterMediaTree_PreservesAncestorFolders(t *testing.T) {
	parentID := types.NewMediaFolderID()
	childID := types.NewMediaFolderID()
	otherID := types.NewMediaFolderID()

	folders := []db.MediaFolder{
		{FolderID: parentID, Name: "parent"},
		{FolderID: childID, Name: "child", ParentID: types.NullableMediaFolderID{ID: parentID, Valid: true}},
		{FolderID: otherID, Name: "other"},
	}
	items := []db.Media{
		makeMediaInFolderWithName("sunset.jpg", "sunset", childID),
		makeMediaInFolderWithName("portrait.jpg", "portrait", otherID),
	}

	filteredFolders, filteredItems := FilterMediaTree(folders, items, "sunset")

	if len(filteredItems) != 1 {
		t.Fatalf("expected 1 matching item, got %d", len(filteredItems))
	}
	// Should preserve parent and child folders, but not other
	if len(filteredFolders) != 2 {
		t.Fatalf("expected 2 ancestor folders (parent + child), got %d", len(filteredFolders))
	}
}

// =============================================================================
// TEST HELPERS
// =============================================================================

func makeMediaInFolderForTree(name string, folderID types.MediaFolderID) db.Media {
	m := makeMediaWithName("https://cdn.example.com/"+name, name, "")
	m.FolderID = types.NullableMediaFolderID{ID: folderID, Valid: true}
	return m
}

func makeMediaInFolderWithName(url, name string, folderID types.MediaFolderID) db.Media {
	m := makeMediaWithName("https://cdn.example.com/"+url, name, "")
	m.FolderID = types.NullableMediaFolderID{ID: folderID, Valid: true}
	return m
}
