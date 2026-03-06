package tui

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

func makeTopLevel(slug, datatypeLabel, datatypeType string, hasRoute bool) db.ContentDataTopLevel {
	t := db.ContentDataTopLevel{
		RouteSlug:     types.Slug(slug),
		RouteTitle:    slug,
		DatatypeLabel: datatypeLabel,
		DatatypeType:  datatypeType,
	}
	t.ContentDataID = types.NewContentID()
	if hasRoute {
		t.RouteID = types.NullableRouteID{ID: types.NewRouteID(), Valid: true}
	}
	return t
}

func TestBuildContentSelectTree_RoutedSorting(t *testing.T) {
	items := []db.ContentDataTopLevel{
		makeTopLevel("/contact", "Page", "_root", true),
		makeTopLevel("/about/team", "Page", "_root", true),
		makeTopLevel("/about", "Page", "_root", true),
		makeTopLevel("/", "Page", "_root", true),
		makeTopLevel("/blog", "Blog", "_root", true),
		makeTopLevel("/about/visit", "Page", "_root", true),
	}

	tree := BuildContentSelectTree(items)

	// Flatten fully expanded
	flat := FlattenSelectTree(tree)

	expectedSlugs := []string{"/", "/about", "/about/team", "/about/visit", "/blog", "/contact"}
	if len(flat) != len(expectedSlugs) {
		t.Fatalf("flat len = %d, want %d", len(flat), len(expectedSlugs))
	}
	for i, n := range flat {
		if n.Slug != expectedSlugs[i] {
			t.Errorf("flat[%d].Slug = %q, want %q", i, n.Slug, expectedSlugs[i])
		}
	}
}

func TestBuildContentSelectTree_CollapsedGroup(t *testing.T) {
	items := []db.ContentDataTopLevel{
		makeTopLevel("/", "Page", "_root", true),
		makeTopLevel("/about", "Page", "_root", true),
		makeTopLevel("/about/team", "Page", "_root", true),
	}

	tree := BuildContentSelectTree(items)
	flat := FlattenSelectTree(tree)

	// Find the /about node and collapse it
	for _, n := range flat {
		if n.Slug == "/about" {
			n.Expand = false
			break
		}
	}

	flat = FlattenSelectTree(tree)
	// Should have / and /about but NOT /about/team
	for _, n := range flat {
		if n.Slug == "/about/team" {
			t.Error("collapsed group should not include /about/team")
		}
	}
	if len(flat) != 2 {
		t.Errorf("collapsed flat len = %d, want 2", len(flat))
	}
}

func TestBuildContentSelectTree_SectionHeaderExclusion(t *testing.T) {
	items := []db.ContentDataTopLevel{
		makeTopLevel("/", "Page", "_root", true),
	}

	tree := BuildContentSelectTree(items)
	flat := FlattenSelectTree(tree)

	for _, n := range flat {
		if n.Kind == NodeSection {
			t.Error("NodeSection should not appear in flat list")
		}
	}
}

func TestBuildContentSelectTree_MixedRoutedStandalone(t *testing.T) {
	items := []db.ContentDataTopLevel{
		makeTopLevel("/", "Page", "_root", true),
		makeTopLevel("", "Menu", "_global", false),
		makeTopLevel("/contact", "Page", "_root", true),
		makeTopLevel("", "Config", "_global", false),
	}
	items[1].RouteTitle = "Main Menu"
	items[3].RouteTitle = "Site Settings"

	tree := BuildContentSelectTree(items)
	flat := FlattenSelectTree(tree)

	// Routed items come first, then standalone
	if len(flat) < 4 {
		t.Fatalf("flat len = %d, want at least 4", len(flat))
	}

	// First two should be routed (/ and /contact)
	if flat[0].Slug != "/" {
		t.Errorf("flat[0].Slug = %q, want /", flat[0].Slug)
	}
	if flat[1].Slug != "/contact" {
		t.Errorf("flat[1].Slug = %q, want /contact", flat[1].Slug)
	}
}

func TestBuildContentSelectTree_GlobalsSectionSeparation(t *testing.T) {
	items := []db.ContentDataTopLevel{
		makeTopLevel("/", "Page", "_root", true),
		makeTopLevel("", "Menu", "_global", false),
		makeTopLevel("", "Footer", "_global", false),
		makeTopLevel("", "Banner", "component", false),
	}
	items[1].RouteTitle = "Main Menu"
	items[2].RouteTitle = "Footer Nav"
	items[3].RouteTitle = "Hero Banner"

	tree := BuildContentSelectTree(items)

	// Check section headers exist in correct order
	var sections []string
	for _, n := range tree {
		if n.Kind == NodeSection {
			sections = append(sections, n.Label)
		}
	}
	want := []string{"Pages", "Globals", "Standalone"}
	if len(sections) != len(want) {
		t.Fatalf("sections = %v, want %v", sections, want)
	}
	for i, s := range sections {
		if s != want[i] {
			t.Errorf("sections[%d] = %q, want %q", i, s, want[i])
		}
	}

	// Flat list should have 4 items total (1 routed + 1 globals group with 2 children + 1 standalone group with 1 child)
	flat := FlattenSelectTree(tree)
	if len(flat) < 4 {
		t.Errorf("flat len = %d, want at least 4", len(flat))
	}
}

func TestBuildContentSelectTree_StandaloneGrouping(t *testing.T) {
	items := []db.ContentDataTopLevel{
		makeTopLevel("", "Menu", "_global", false),
		makeTopLevel("", "Menu", "_global", false),
		makeTopLevel("", "Config", "_global", false),
	}
	items[0].RouteTitle = "Main Menu"
	items[1].RouteTitle = "Footer"
	items[2].RouteTitle = "Site Settings"

	tree := BuildContentSelectTree(items)
	flat := FlattenSelectTree(tree)

	// Should have: Config group, Site Settings, Menu group, Main Menu, Footer
	// Groups: Config (1 child), Menu (2 children)
	groupCount := 0
	contentCount := 0
	for _, n := range flat {
		switch n.Kind {
		case NodeGroup:
			groupCount++
		case NodeContent:
			contentCount++
		}
	}
	if groupCount != 2 {
		t.Errorf("group count = %d, want 2", groupCount)
	}
	if contentCount != 3 {
		t.Errorf("content count = %d, want 3", contentCount)
	}
}

func TestFlattenSelectTree_Empty(t *testing.T) {
	flat := FlattenSelectTree(nil)
	if len(flat) != 0 {
		t.Errorf("flat len = %d, want 0", len(flat))
	}
}

func TestCompareSlugSegments(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"/", "/about", -1},
		{"/about", "/about/team", -1},
		{"/about/team", "/about/visit", -1},
		{"/about", "/blog", -1},
		{"/", "/", 0},
		{"/blog", "/about", 1},
	}

	for _, tt := range tests {
		got := compareSlugSegments(tt.a, tt.b)
		if (tt.want < 0 && got >= 0) || (tt.want > 0 && got <= 0) || (tt.want == 0 && got != 0) {
			t.Errorf("compareSlugSegments(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestBuildAdminContentSelectTree(t *testing.T) {
	items := []db.AdminContentDataTopLevel{
		{
			AdminContentData: db.AdminContentData{
				AdminContentDataID: types.AdminContentID(types.NewContentID()),
				AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID(types.NewRouteID()), Valid: true},
			},
			RouteSlug:     "/",
			DatatypeLabel: "Page",
			DatatypeType:  "_root",
		},
		{
			AdminContentData: db.AdminContentData{
				AdminContentDataID: types.AdminContentID(types.NewContentID()),
				AdminRouteID:       types.NullableAdminRouteID{ID: types.AdminRouteID(types.NewRouteID()), Valid: true},
			},
			RouteSlug:     "/about",
			DatatypeLabel: "Page",
			DatatypeType:  "_root",
		},
	}

	tree := BuildAdminContentSelectTree(items)
	flat := FlattenSelectTree(tree)

	if len(flat) != 2 {
		t.Fatalf("flat len = %d, want 2", len(flat))
	}
	// Both should have AdminContent set, not Content
	for i, n := range flat {
		if n.AdminContent == nil {
			t.Errorf("flat[%d].AdminContent is nil, want non-nil", i)
		}
		if n.Content != nil {
			t.Errorf("flat[%d].Content should be nil for admin tree", i)
		}
	}
}
