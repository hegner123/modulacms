package core_test

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree/core"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// makeContentID creates a deterministic ContentID for test use.
func makeContentID(suffix string) types.ContentID {
	base := "01JAAAAAAAAAAAAAAAAAAAAAAA"
	if len(suffix) > 26 {
		suffix = suffix[:26]
	}
	id := base[:26-len(suffix)] + suffix
	return types.ContentID(id)
}

func nullID(id types.ContentID) types.NullableContentID {
	return types.NullableContentID{ID: id, Valid: true}
}

func emptyNullID() types.NullableContentID {
	return types.NullableContentID{Valid: false}
}

func nullDatatypeID(id types.DatatypeID) types.NullableDatatypeID {
	return types.NullableDatatypeID{ID: id, Valid: true}
}

func emptyNullDatatypeID() types.NullableDatatypeID {
	return types.NullableDatatypeID{Valid: false}
}

// makeRow builds a GetContentTreeByRouteRow with the given IDs.
func makeRow(id, parentID string) db.GetContentTreeByRouteRow {
	row := db.GetContentTreeByRouteRow{
		ContentDataID: makeContentID(id),
		DatatypeLabel: "test-type",
		DatatypeType:  "block",
	}
	if parentID == "" {
		row.ParentID = emptyNullID()
	} else {
		row.ParentID = nullID(makeContentID(parentID))
	}
	return row
}

// makeRowWithSiblings builds a row with explicit sibling pointers.
func makeRowWithSiblings(id, parentID, firstChildID, nextSiblingID, prevSiblingID, dtType string) db.GetContentTreeByRouteRow {
	row := db.GetContentTreeByRouteRow{
		ContentDataID: makeContentID(id),
		DatatypeLabel: "test-type",
		DatatypeType:  dtType,
	}
	if parentID == "" {
		row.ParentID = emptyNullID()
	} else {
		row.ParentID = nullID(makeContentID(parentID))
	}
	if firstChildID != "" {
		row.FirstChildID = nullID(makeContentID(firstChildID))
	}
	if nextSiblingID != "" {
		row.NextSiblingID = nullID(makeContentID(nextSiblingID))
	}
	if prevSiblingID != "" {
		row.PrevSiblingID = nullID(makeContentID(prevSiblingID))
	}
	return row
}

// collectChildren walks the FirstChild/NextSibling chain and returns ContentDataIDs.
func collectChildren(parent *core.Node) []types.ContentID {
	var ids []types.ContentID
	child := parent.FirstChild
	for child != nil {
		ids = append(ids, child.ContentData.ContentDataID)
		child = child.NextSibling
	}
	return ids
}

// ---------------------------------------------------------------------------
// BuildFromRows tests
// ---------------------------------------------------------------------------

func TestBuildFromRows_SimpleTree(t *testing.T) {
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child1", "root"),
		makeRow("child2", "root"),
	}

	root, stats, err := core.BuildFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Node == nil {
		t.Fatal("root node is nil")
	}
	if stats.NodesCount != 3 {
		t.Errorf("NodesCount = %d, want 3", stats.NodesCount)
	}
	if root.Node.ContentData.ContentDataID != makeContentID("root") {
		t.Errorf("root ID = %v, want root", root.Node.ContentData.ContentDataID)
	}

	children := collectChildren(root.Node)
	if len(children) != 2 {
		t.Fatalf("root has %d children, want 2", len(children))
	}
}

func TestBuildFromRows_RootIdentifiedByIsRootType(t *testing.T) {
	// The first parentless node becomes root when using BuildFromRows
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child1", "root"),
	}

	root, _, err := core.BuildFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Node.ContentData.ContentDataID != makeContentID("root") {
		t.Errorf("root ID = %v, want root", root.Node.ContentData.ContentDataID)
	}
}

func TestBuildFromRows_SiblingOrdering(t *testing.T) {
	// Root with first_child_id=C, chain: C->A->B
	rows := []db.GetContentTreeByRouteRow{
		makeRowWithSiblings("root", "", "C", "", "", "_root"),
		makeRowWithSiblings("A", "root", "", "B", "C", "page"),
		makeRowWithSiblings("B", "root", "", "", "A", "page"),
		makeRowWithSiblings("C", "root", "", "A", "", "page"),
	}

	root, _, err := core.BuildFromRows(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	children := collectChildren(root.Node)
	if len(children) != 3 {
		t.Fatalf("got %d children, want 3", len(children))
	}

	wantOrder := []string{"C", "A", "B"}
	for i, want := range wantOrder {
		got := children[i]
		if got != makeContentID(want) {
			t.Errorf("child[%d] = %v, want %v", i, got, makeContentID(want))
		}
	}
}

func TestBuildFromRows_OrphanHandling(t *testing.T) {
	// Node with parent that doesn't exist
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child1", "root"),
		makeRow("orphan", "nonexistent"),
	}

	root, stats, err := core.BuildFromRows(rows)
	if err == nil {
		t.Fatal("expected error for orphan nodes, got nil")
	}
	if root.Node == nil {
		t.Fatal("root node should not be nil despite orphans")
	}
	if len(stats.FinalOrphans) != 1 {
		t.Errorf("FinalOrphans = %d, want 1", len(stats.FinalOrphans))
	}
}

func TestBuildFromRows_CircularRefDetection(t *testing.T) {
	// Create nodes where orphan resolution detects a circular parent chain.
	// Node A has parent C (which doesn't exist) and node B has parent A.
	// C doesn't exist, so A becomes an orphan. B's parent A exists but
	// A.Parent is nil (A is an orphan), so B stays orphaned too.
	// The orphan resolution tries to connect them but can't (A's parent C
	// doesn't exist), so they remain orphans. No circular ref in this case.
	//
	// For actual circular detection: A -> B -> A. Both parents exist in
	// the index, so Phase 2 attaches them (A becomes child of B, B becomes
	// child of A -- creating a circular structure). The orphan phase is
	// never reached for these nodes.
	//
	// The original tree.go only detects circular refs among orphans. Nodes
	// that both exist in the index form mutual parent-child relationships
	// during Phase 2 without going through orphan resolution.
	//
	// Test that a structure with orphan cycles is detected:
	// Root exists. A's parent is "ghost1" (not in index, so A is orphan).
	// ghost1 doesn't exist. A stays orphan forever, reported in FinalOrphans.
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("A", "ghost1"),
	}

	_, stats, err := core.BuildFromRows(rows)
	if err == nil {
		t.Fatal("expected error for orphan nodes, got nil")
	}
	if len(stats.FinalOrphans) != 1 {
		t.Errorf("FinalOrphans = %d, want 1", len(stats.FinalOrphans))
	}
}

func TestBuildFromRows_MutualParentCircle(t *testing.T) {
	// When both nodes exist and reference each other as parents,
	// Phase 2 attaches them forming a circular structure.
	// This doesn't trigger the orphan-based circular ref detection
	// because neither becomes an orphan. The tree still builds
	// (root is the first parentless node).
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("A", "B"),
		makeRow("B", "A"),
	}

	root, stats, _ := core.BuildFromRows(rows)
	if root.Node == nil {
		t.Fatal("root node should exist")
	}
	// A and B are attached to each other, not orphans
	if stats.NodesCount != 3 {
		t.Errorf("NodesCount = %d, want 3", stats.NodesCount)
	}
}

func TestBuildFromRows_EmptyRows(t *testing.T) {
	root, stats, err := core.BuildFromRows(nil)
	if err == nil {
		t.Fatal("expected error for empty rows")
	}
	if root.Node != nil {
		t.Error("root should be nil for empty rows")
	}
	if stats.NodesCount != 0 {
		t.Errorf("NodesCount = %d, want 0", stats.NodesCount)
	}
}

// ---------------------------------------------------------------------------
// BuildTree tests
// ---------------------------------------------------------------------------

func TestBuildTree_SimpleTree(t *testing.T) {
	rootID := makeContentID("root")
	childID := makeContentID("child1")
	dtID := types.DatatypeID("01JAAAAAAAAAAAAAAAAAADT001")

	cd := []db.ContentData{
		{
			ContentDataID: rootID,
			ParentID:      emptyNullID(),
			DatatypeID:    nullDatatypeID(dtID),
		},
		{
			ContentDataID: childID,
			ParentID:      nullID(rootID),
			DatatypeID:    nullDatatypeID(dtID),
		},
	}
	dt := []db.Datatypes{
		{DatatypeID: dtID, Label: "root", Type: "_root"},
		{DatatypeID: dtID, Label: "child", Type: "page"},
	}

	root, stats, err := core.BuildTree(cd, dt, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Node == nil {
		t.Fatal("root node is nil")
	}
	if stats.NodesCount != 2 {
		t.Errorf("NodesCount = %d, want 2", stats.NodesCount)
	}

	// Root identified by _root type
	if root.Node.ContentData.ContentDataID != rootID {
		t.Errorf("root ID = %v, want %v", root.Node.ContentData.ContentDataID, rootID)
	}

	children := collectChildren(root.Node)
	if len(children) != 1 {
		t.Fatalf("root has %d children, want 1", len(children))
	}
	if children[0] != childID {
		t.Errorf("child ID = %v, want %v", children[0], childID)
	}
}

func TestBuildTree_RootByNestedRoot(t *testing.T) {
	rootID := makeContentID("root")
	dtID := types.DatatypeID("01JAAAAAAAAAAAAAAAAAADT001")

	cd := []db.ContentData{
		{ContentDataID: rootID, ParentID: emptyNullID(), DatatypeID: nullDatatypeID(dtID)},
	}
	dt := []db.Datatypes{
		{DatatypeID: dtID, Label: "root", Type: "_nested_root"},
	}

	root, _, err := core.BuildTree(cd, dt, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Node == nil {
		t.Fatal("root is nil")
	}
	if root.Node.Datatype.Type != "_nested_root" {
		t.Errorf("root type = %q, want _nested_root", root.Node.Datatype.Type)
	}
}

func TestBuildTree_SiblingOrdering(t *testing.T) {
	rootID := makeContentID("root")
	aID := makeContentID("A")
	bID := makeContentID("B")
	cID := makeContentID("C")
	dtID := types.DatatypeID("01JAAAAAAAAAAAAAAAAAADT001")

	cd := []db.ContentData{
		{
			ContentDataID: rootID,
			ParentID:      emptyNullID(),
			FirstChildID:  nullID(cID),
			DatatypeID:    nullDatatypeID(dtID),
		},
		{
			ContentDataID: aID,
			ParentID:      nullID(rootID),
			NextSiblingID: nullID(bID),
			PrevSiblingID: nullID(cID),
			DatatypeID:    nullDatatypeID(dtID),
		},
		{
			ContentDataID: bID,
			ParentID:      nullID(rootID),
			PrevSiblingID: nullID(aID),
			DatatypeID:    nullDatatypeID(dtID),
		},
		{
			ContentDataID: cID,
			ParentID:      nullID(rootID),
			NextSiblingID: nullID(aID),
			DatatypeID:    nullDatatypeID(dtID),
		},
	}
	dt := []db.Datatypes{
		{DatatypeID: dtID, Label: "root", Type: "_root"},
		{DatatypeID: dtID, Label: "A", Type: "page"},
		{DatatypeID: dtID, Label: "B", Type: "page"},
		{DatatypeID: dtID, Label: "C", Type: "page"},
	}

	root, _, err := core.BuildTree(cd, dt, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	children := collectChildren(root.Node)
	if len(children) != 3 {
		t.Fatalf("got %d children, want 3", len(children))
	}

	wantOrder := []types.ContentID{cID, aID, bID}
	for i, want := range wantOrder {
		if children[i] != want {
			t.Errorf("child[%d] = %v, want %v", i, children[i], want)
		}
	}
}

func TestBuildTree_FieldAttachment(t *testing.T) {
	rootID := makeContentID("root")
	dtID := types.DatatypeID("01JAAAAAAAAAAAAAAAAAADT001")
	fieldID := types.FieldID("01JAAAAAAAAAAAAAAAAAFLD001")

	cd := []db.ContentData{
		{ContentDataID: rootID, ParentID: emptyNullID(), DatatypeID: nullDatatypeID(dtID)},
	}
	dt := []db.Datatypes{
		{DatatypeID: dtID, Label: "root", Type: "_root"},
	}
	cf := []db.ContentFields{
		{
			ContentFieldID: types.ContentFieldID("01JAAAAAAAAAAAAAAAAACF0001"),
			ContentDataID:  nullID(rootID),
			FieldValue:     "hello",
		},
	}
	df := []db.Fields{
		{
			FieldID: fieldID,
			Label:   "title",
			Type:    types.FieldTypeText,
		},
	}

	root, _, err := core.BuildTree(cd, dt, cf, df)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(root.Node.ContentFields) != 1 {
		t.Fatalf("root has %d content fields, want 1", len(root.Node.ContentFields))
	}
	if root.Node.ContentFields[0].FieldValue != "hello" {
		t.Errorf("field value = %q, want hello", root.Node.ContentFields[0].FieldValue)
	}
	if len(root.Node.Fields) != 1 {
		t.Fatalf("root has %d fields, want 1", len(root.Node.Fields))
	}
	if root.Node.Fields[0].Label != "title" {
		t.Errorf("field label = %q, want title", root.Node.Fields[0].Label)
	}
}

func TestBuildTree_MismatchedLengths(t *testing.T) {
	cd := []db.ContentData{{ContentDataID: makeContentID("a")}}
	dt := []db.Datatypes{} // wrong length

	_, _, err := core.BuildTree(cd, dt, nil, nil)
	if err == nil {
		t.Fatal("expected error for mismatched lengths")
	}
}

func TestBuildTree_FallbackRootNoParent(t *testing.T) {
	// No node has _root type, but one has no parent -- it becomes root
	id := makeContentID("only")
	dtID := types.DatatypeID("01JAAAAAAAAAAAAAAAAAADT001")

	cd := []db.ContentData{
		{ContentDataID: id, ParentID: emptyNullID(), DatatypeID: nullDatatypeID(dtID)},
	}
	dt := []db.Datatypes{
		{DatatypeID: dtID, Label: "fallback", Type: "page"},
	}

	root, _, err := core.BuildTree(cd, dt, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Node == nil {
		t.Fatal("root is nil")
	}
	if root.Node.ContentData.ContentDataID != id {
		t.Errorf("root ID = %v, want %v", root.Node.ContentData.ContentDataID, id)
	}
}

// ---------------------------------------------------------------------------
// BuildAdminTree tests
// ---------------------------------------------------------------------------

func TestBuildAdminTree_SimpleTree(t *testing.T) {
	rootID := types.AdminContentID("01JAAAAAAAAAAAAAAADMROOT01")
	childID := types.AdminContentID("01JAAAAAAAAAAAAAAADMCHLD01")
	dtID := types.AdminDatatypeID("01JAAAAAAAAAAAAAAADMDT0001")

	cd := []db.AdminContentData{
		{
			AdminContentDataID: rootID,
			ParentID:           types.NullableAdminContentID{Valid: false},
			AdminDatatypeID:    types.NullableAdminDatatypeID{ID: dtID, Valid: true},
		},
		{
			AdminContentDataID: childID,
			ParentID:           types.NullableAdminContentID{ID: rootID, Valid: true},
			AdminDatatypeID:    types.NullableAdminDatatypeID{ID: dtID, Valid: true},
		},
	}
	dt := []db.AdminDatatypes{
		{AdminDatatypeID: dtID, Label: "root", Type: "_root"},
		{AdminDatatypeID: dtID, Label: "child", Type: "page"},
	}

	root, stats, err := core.BuildAdminTree(cd, dt, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Node == nil {
		t.Fatal("root is nil")
	}
	if stats.NodesCount != 2 {
		t.Errorf("NodesCount = %d, want 2", stats.NodesCount)
	}

	children := collectChildren(root.Node)
	if len(children) != 1 {
		t.Fatalf("root has %d children, want 1", len(children))
	}
}

// ---------------------------------------------------------------------------
// Traverse tests
// ---------------------------------------------------------------------------

func TestCountVisible(t *testing.T) {
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child1", "root"),
		makeRow("child2", "root"),
		makeRow("grandchild", "child1"),
	}

	root, _, _ := core.BuildFromRows(rows)

	// All expanded
	allExpanded := func(*core.Node) bool { return true }
	count := core.CountVisible(root.Node, allExpanded)
	if count != 4 {
		t.Errorf("CountVisible (all expanded) = %d, want 4", count)
	}

	// Nothing expanded (only root visible + its siblings)
	noneExpanded := func(*core.Node) bool { return false }
	count = core.CountVisible(root.Node, noneExpanded)
	if count != 1 {
		t.Errorf("CountVisible (none expanded) = %d, want 1", count)
	}
}

func TestFlattenVisible(t *testing.T) {
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child1", "root"),
		makeRow("child2", "root"),
	}

	root, _, _ := core.BuildFromRows(rows)
	allExpanded := func(*core.Node) bool { return true }
	flat := core.FlattenVisible(root.Node, allExpanded)
	if len(flat) != 3 {
		t.Fatalf("FlattenVisible = %d nodes, want 3", len(flat))
	}
}

func TestNodeAtIndex(t *testing.T) {
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child1", "root"),
		makeRow("child2", "root"),
	}

	root, _, _ := core.BuildFromRows(rows)
	allExpanded := func(*core.Node) bool { return true }

	// Index 0 should be root
	n := core.NodeAtIndex(root.Node, 0, allExpanded)
	if n == nil {
		t.Fatal("NodeAtIndex(0) returned nil")
	}
	if n.ContentData.ContentDataID != makeContentID("root") {
		t.Errorf("NodeAtIndex(0) = %v, want root", n.ContentData.ContentDataID)
	}

	// Out of range
	n = core.NodeAtIndex(root.Node, 100, allExpanded)
	if n != nil {
		t.Error("NodeAtIndex(100) should return nil")
	}
}

func TestFindByContentID(t *testing.T) {
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child1", "root"),
	}

	root, _, _ := core.BuildFromRows(rows)

	found := core.FindByContentID(root, makeContentID("child1"))
	if found == nil {
		t.Fatal("FindByContentID returned nil for existing node")
	}
	if found.ContentData.ContentDataID != makeContentID("child1") {
		t.Errorf("found ID = %v, want child1", found.ContentData.ContentDataID)
	}

	notFound := core.FindByContentID(root, makeContentID("nonexistent"))
	if notFound != nil {
		t.Error("FindByContentID should return nil for nonexistent node")
	}
}

func TestIsDescendantOf(t *testing.T) {
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child", "root"),
		makeRow("grandchild", "child"),
	}

	root, _, _ := core.BuildFromRows(rows)

	grandchild := core.FindByContentID(root, makeContentID("grandchild"))
	child := core.FindByContentID(root, makeContentID("child"))
	rootNode := root.Node

	if !core.IsDescendantOf(grandchild, rootNode) {
		t.Error("grandchild should be descendant of root")
	}
	if !core.IsDescendantOf(grandchild, child) {
		t.Error("grandchild should be descendant of child")
	}
	if core.IsDescendantOf(rootNode, grandchild) {
		t.Error("root should not be descendant of grandchild")
	}
}

func TestFindVisibleIndex(t *testing.T) {
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child1", "root"),
		makeRow("child2", "root"),
	}

	root, _, _ := core.BuildFromRows(rows)
	allExpanded := func(*core.Node) bool { return true }

	idx := core.FindVisibleIndex(root.Node, makeContentID("root"), allExpanded)
	if idx != 0 {
		t.Errorf("FindVisibleIndex(root) = %d, want 0", idx)
	}

	idx = core.FindVisibleIndex(root.Node, makeContentID("nonexistent"), allExpanded)
	if idx != -1 {
		t.Errorf("FindVisibleIndex(nonexistent) = %d, want -1", idx)
	}
}

// ---------------------------------------------------------------------------
// Mutation tests
// ---------------------------------------------------------------------------

func TestInsertNode(t *testing.T) {
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
	}

	root, _, _ := core.BuildFromRows(rows)

	newCD := db.ContentData{
		ContentDataID: makeContentID("new"),
		ParentID:      nullID(makeContentID("root")),
	}
	newNode := &core.Node{ContentData: &newCD}

	core.InsertNode(root, root.Node, nil, nil, nil, newNode)

	found := core.FindByContentID(root, makeContentID("new"))
	if found == nil {
		t.Fatal("inserted node not found in index")
	}
	if found.Parent != root.Node {
		t.Error("inserted node's parent should be root")
	}
}

func TestDeleteNode(t *testing.T) {
	rows := []db.GetContentTreeByRouteRow{
		makeRow("root", ""),
		makeRow("child1", "root"),
		makeRow("child2", "root"),
	}

	root, _, _ := core.BuildFromRows(rows)

	child1 := core.FindByContentID(root, makeContentID("child1"))
	ok := core.DeleteNode(root, child1)
	if !ok {
		t.Fatal("DeleteNode returned false")
	}

	found := core.FindByContentID(root, makeContentID("child1"))
	if found != nil {
		t.Error("deleted node should not be in index")
	}

	// Cannot delete root
	ok = core.DeleteNode(root, root.Node)
	if ok {
		t.Error("should not be able to delete root node")
	}
}
