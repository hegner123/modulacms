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

func TestBuildMediaTree_Empty(t *testing.T) {
	result := BuildMediaTree(nil)
	if result != nil {
		t.Fatalf("expected nil for empty input, got %d nodes", len(result))
	}
}

func TestBuildMediaTree_SingleFile(t *testing.T) {
	items := []db.Media{makeMedia("https://example.com/photo.jpg")}
	roots := BuildMediaTree(items)

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].Kind != MediaNodeFile {
		t.Errorf("expected MediaNodeFile, got %d", roots[0].Kind)
	}
	if roots[0].Label != "photo.jpg" {
		t.Errorf("expected label 'photo.jpg', got %q", roots[0].Label)
	}
	if roots[0].Media == nil {
		t.Error("expected non-nil Media pointer")
	}
}

func TestBuildMediaTree_SharedPrefix(t *testing.T) {
	items := []db.Media{
		makeMedia("https://cdn.example.com/images/a.jpg"),
		makeMedia("https://cdn.example.com/images/b.jpg"),
	}
	roots := BuildMediaTree(items)

	// Should have one root: "images/" folder
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
	items := []db.Media{
		makeMedia("https://cdn.example.com/a/b/c/file.jpg"),
	}
	roots := BuildMediaTree(items)

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
	if f == nil || f.Kind != MediaNodeFile || f.Label != "file.jpg" {
		t.Fatalf("expected file 'file.jpg' as child of 'c'")
	}
}

func TestBuildMediaTree_MixedDepths(t *testing.T) {
	items := []db.Media{
		makeMedia("https://cdn.example.com/root.jpg"),
		makeMedia("https://cdn.example.com/images/nested.jpg"),
	}
	roots := BuildMediaTree(items)

	if len(roots) != 2 {
		t.Fatalf("expected 2 roots (images/ folder + root.jpg), got %d", len(roots))
	}

	// Verify we have one folder and one file at root level
	var folders, files int
	for _, r := range roots {
		if r.Kind == MediaNodeFolder {
			folders++
		}
		if r.Kind == MediaNodeFile {
			files++
		}
	}
	if folders != 1 {
		t.Errorf("expected 1 folder at root, got %d", folders)
	}
	if files != 1 {
		t.Errorf("expected 1 file at root, got %d", files)
	}
}

func TestFlattenMediaTree_RespectsExpand(t *testing.T) {
	items := []db.Media{
		makeMedia("https://cdn.example.com/images/a.jpg"),
		makeMedia("https://cdn.example.com/images/b.jpg"),
	}
	roots := BuildMediaTree(items)

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
