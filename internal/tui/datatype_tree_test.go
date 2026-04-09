package tui

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

func TestBuildDatatypeTree_Flat(t *testing.T) {
	items := []db.Datatypes{
		{DatatypeID: "dt1", Label: "Blog", SortOrder: 1},
		{DatatypeID: "dt2", Label: "Page", SortOrder: 2},
	}
	roots := BuildDatatypeTree(items)
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	if roots[0].Label != "Blog" {
		t.Errorf("expected first root Blog, got %s", roots[0].Label)
	}
	if roots[0].Kind != DatatypeNodeItem {
		t.Errorf("expected Blog to be item, got %d", roots[0].Kind)
	}
}

func TestBuildDatatypeTree_Hierarchy(t *testing.T) {
	items := []db.Datatypes{
		{DatatypeID: "dt1", Label: "Content", SortOrder: 1},
		{DatatypeID: "dt2", Label: "Blog Post", SortOrder: 1, ParentID: types.NullableDatatypeID{ID: "dt1", Valid: true}},
		{DatatypeID: "dt3", Label: "Article", SortOrder: 2, ParentID: types.NullableDatatypeID{ID: "dt1", Valid: true}},
	}
	roots := BuildDatatypeTree(items)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].Kind != DatatypeNodeGroup {
		t.Errorf("expected Content to be group, got %d", roots[0].Kind)
	}
	if len(roots[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(roots[0].Children))
	}
	if roots[0].Children[0].Label != "Blog Post" {
		t.Errorf("expected first child Blog Post, got %s", roots[0].Children[0].Label)
	}
}

func TestFlattenDatatypeTree_Expanded(t *testing.T) {
	items := []db.Datatypes{
		{DatatypeID: "dt1", Label: "Content", SortOrder: 1},
		{DatatypeID: "dt2", Label: "Blog", SortOrder: 1, ParentID: types.NullableDatatypeID{ID: "dt1", Valid: true}},
		{DatatypeID: "dt3", Label: "Page", SortOrder: 2, ParentID: types.NullableDatatypeID{ID: "dt1", Valid: true}},
	}
	roots := BuildDatatypeTree(items)
	flat := FlattenDatatypeTree(roots)
	if len(flat) != 3 {
		t.Fatalf("expected 3 nodes in flat list, got %d", len(flat))
	}
	if flat[0].Depth != 0 {
		t.Errorf("expected depth 0 for root, got %d", flat[0].Depth)
	}
	if flat[1].Depth != 1 {
		t.Errorf("expected depth 1 for child, got %d", flat[1].Depth)
	}
}

func TestFlattenDatatypeTree_Collapsed(t *testing.T) {
	items := []db.Datatypes{
		{DatatypeID: "dt1", Label: "Content", SortOrder: 1},
		{DatatypeID: "dt2", Label: "Blog", SortOrder: 1, ParentID: types.NullableDatatypeID{ID: "dt1", Valid: true}},
	}
	roots := BuildDatatypeTree(items)
	roots[0].Expand = false
	flat := FlattenDatatypeTree(roots)
	if len(flat) != 1 {
		t.Fatalf("expected 1 node when collapsed, got %d", len(flat))
	}
}

func TestFilterDatatypeList(t *testing.T) {
	items := []db.Datatypes{
		{DatatypeID: "dt1", Label: "Blog Post"},
		{DatatypeID: "dt2", Label: "Page"},
		{DatatypeID: "dt3", Label: "Blog Category"},
	}
	filtered := FilterDatatypeList(items, "blog")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(filtered))
	}
}

func TestFilterDatatypeList_Empty(t *testing.T) {
	items := []db.Datatypes{
		{DatatypeID: "dt1", Label: "Blog"},
	}
	filtered := FilterDatatypeList(items, "")
	if len(filtered) != 1 {
		t.Fatalf("expected full list on empty query, got %d", len(filtered))
	}
}

func TestBuildDatatypeTree_Empty(t *testing.T) {
	roots := BuildDatatypeTree(nil)
	if roots != nil {
		t.Fatalf("expected nil for empty input, got %v", roots)
	}
}

func TestFlattenDatatypeTree_Empty(t *testing.T) {
	flat := FlattenDatatypeTree(nil)
	if len(flat) != 0 {
		t.Fatalf("expected empty flat list, got %d", len(flat))
	}
}

func TestBuildDatatypeTree_OrphanParent(t *testing.T) {
	// Child references a parent that doesn't exist -- should become a root
	items := []db.Datatypes{
		{DatatypeID: "dt1", Label: "Orphan", ParentID: types.NullableDatatypeID{ID: "missing", Valid: true}},
	}
	roots := BuildDatatypeTree(items)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root for orphan, got %d", len(roots))
	}
}

func TestBuildAdminDatatypeTree_Hierarchy(t *testing.T) {
	items := []db.AdminDatatypes{
		{AdminDatatypeID: "adt1", Label: "Root", SortOrder: 1},
		{AdminDatatypeID: "adt2", Label: "Child", SortOrder: 1, ParentID: types.NullableAdminDatatypeID{ID: "adt1", Valid: true}},
	}
	roots := BuildAdminDatatypeTree(items)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if len(roots[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(roots[0].Children))
	}
}

func TestFilterAdminDatatypeList(t *testing.T) {
	items := []db.AdminDatatypes{
		{AdminDatatypeID: "adt1", Label: "System Config"},
		{AdminDatatypeID: "adt2", Label: "user Settings"},
	}
	filtered := FilterAdminDatatypeList(items, "config")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 match, got %d", len(filtered))
	}
}

func TestFlattenDatatypeTree_DeepNesting(t *testing.T) {
	items := []db.Datatypes{
		{DatatypeID: "dt1", Label: "L0", SortOrder: 1},
		{DatatypeID: "dt2", Label: "L1", SortOrder: 1, ParentID: types.NullableDatatypeID{ID: "dt1", Valid: true}},
		{DatatypeID: "dt3", Label: "L2", SortOrder: 1, ParentID: types.NullableDatatypeID{ID: "dt2", Valid: true}},
	}
	roots := BuildDatatypeTree(items)
	flat := FlattenDatatypeTree(roots)
	if len(flat) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(flat))
	}
	for i, expected := range []int{0, 1, 2} {
		if flat[i].Depth != expected {
			t.Errorf("flat[%d] depth: expected %d, got %d", i, expected, flat[i].Depth)
		}
	}
}
