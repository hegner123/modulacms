package tree_test

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// ---------------------------------------------------------------------------
// Admin test helpers
// ---------------------------------------------------------------------------

// makeAdminContentID creates a deterministic AdminContentID for test use.
func makeAdminContentID(suffix string) types.AdminContentID {
	base := "01JAAAAAAAAAAAAAAAAAAAAAAA"
	if len(suffix) > 26 {
		suffix = suffix[:26]
	}
	id := base[:26-len(suffix)] + suffix
	return types.AdminContentID(id)
}

// makeAdminDatatypeID creates a deterministic AdminDatatypeID for test use.
func makeAdminDatatypeID(suffix string) types.AdminDatatypeID {
	base := "01JBBBBBBBBBBBBBBBBBBBBBBB"
	if len(suffix) > 26 {
		suffix = suffix[:26]
	}
	id := base[:26-len(suffix)] + suffix
	return types.AdminDatatypeID(id)
}

// nullAdminContentID wraps an AdminContentID in a valid NullableAdminContentID.
func nullAdminContentID(id types.AdminContentID) types.NullableAdminContentID {
	return types.NullableAdminContentID{ID: id, Valid: true}
}

// emptyNullAdminContentID returns an invalid NullableAdminContentID.
func emptyNullAdminContentID() types.NullableAdminContentID {
	return types.NullableAdminContentID{Valid: false}
}

// nullAdminDatatypeID wraps an AdminDatatypeID in a valid NullableAdminDatatypeID.
func nullAdminDatatypeID(id types.AdminDatatypeID) types.NullableAdminDatatypeID {
	return types.NullableAdminDatatypeID{ID: id, Valid: true}
}

// makeAdminData builds paired AdminContentData and AdminDatatypes slices for a single node.
// parentSuffix == "" means root node (no parent).
func makeAdminData(suffix, parentSuffix, dtSuffix, dtLabel, dtType string) (db.AdminContentData, db.AdminDatatypes) {
	cd := db.AdminContentData{
		AdminContentDataID: makeAdminContentID(suffix),
		Status:             types.ContentStatusDraft,
	}
	if parentSuffix == "" {
		cd.ParentID = emptyNullAdminContentID()
	} else {
		cd.ParentID = nullAdminContentID(makeAdminContentID(parentSuffix))
	}

	dt := db.AdminDatatypes{
		AdminDatatypeID: makeAdminDatatypeID(dtSuffix),
		Label:           dtLabel,
		Type:            dtType,
	}

	return cd, dt
}

// makeAdminDataWithSiblings builds a node with explicit sibling pointers.
func makeAdminDataWithSiblings(suffix, parentSuffix, firstChildSuffix, nextSiblingSuffix, prevSiblingSuffix, dtSuffix, dtLabel, dtType string) (db.AdminContentData, db.AdminDatatypes) {
	cd, dt := makeAdminData(suffix, parentSuffix, dtSuffix, dtLabel, dtType)
	if firstChildSuffix != "" {
		cd.FirstChildID = nullAdminContentID(makeAdminContentID(firstChildSuffix))
	}
	if nextSiblingSuffix != "" {
		cd.NextSiblingID = nullAdminContentID(makeAdminContentID(nextSiblingSuffix))
	}
	if prevSiblingSuffix != "" {
		cd.PrevSiblingID = nullAdminContentID(makeAdminContentID(prevSiblingSuffix))
	}
	return cd, dt
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestLoadFromAdminData_SingleRoot(t *testing.T) {
	cd1, dt1 := makeAdminData("root", "", "dt1", "Page", "root")

	root := tree.NewRoot()
	stats, err := root.LoadFromAdminData(
		[]db.AdminContentData{cd1},
		[]db.AdminDatatypes{dt1},
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.NodesCount != 1 {
		t.Errorf("NodesCount = %d, want 1", stats.NodesCount)
	}
	if root.Root == nil {
		t.Fatal("Root is nil")
	}
	// The admin ID is converted to ContentID via string casting
	expectedID := types.ContentID(makeAdminContentID("root").String())
	if root.Root.Instance.ContentDataID != expectedID {
		t.Errorf("Root ContentDataID = %q, want %q", root.Root.Instance.ContentDataID, expectedID)
	}
	if len(root.NodeIndex) != 1 {
		t.Errorf("NodeIndex size = %d, want 1", len(root.NodeIndex))
	}
}

func TestLoadFromAdminData_ParentChild(t *testing.T) {
	cd1, dt1 := makeAdminData("root", "", "dt1", "Page", "root")
	cd2, dt2 := makeAdminData("child1", "root", "dt2", "Block", "block")
	cd3, dt3 := makeAdminData("child2", "root", "dt3", "Block", "block")

	root := tree.NewRoot()
	stats, err := root.LoadFromAdminData(
		[]db.AdminContentData{cd1, cd2, cd3},
		[]db.AdminDatatypes{dt1, dt2, dt3},
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.NodesCount != 3 {
		t.Errorf("NodesCount = %d, want 3", stats.NodesCount)
	}
	if root.Root == nil {
		t.Fatal("Root is nil")
	}
	if root.Root.FirstChild == nil {
		t.Fatal("Root.FirstChild is nil")
	}

	// Verify parent pointers reference tree.Node, not core.Node
	child := root.Root.FirstChild
	if child.Parent != root.Root {
		t.Error("FirstChild.Parent does not point to tree.Root node")
	}

	// Walk siblings to ensure both children are reachable
	childCount := 0
	current := root.Root.FirstChild
	for current != nil {
		childCount++
		if current.Parent != root.Root {
			t.Errorf("child node parent pointer does not reference tree.Root node")
		}
		current = current.NextSibling
	}
	if childCount != 2 {
		t.Errorf("child count = %d, want 2", childCount)
	}
}

func TestLoadFromAdminData_SiblingPointers(t *testing.T) {
	// Build tree with explicit sibling order: root -> [A, B, C] where A<->B<->C
	cdRoot, dtRoot := makeAdminDataWithSiblings("root", "", "A", "", "", "dt0", "Page", "root")
	cdA, dtA := makeAdminDataWithSiblings("A", "root", "", "B", "", "dt1", "Block", "block")
	cdB, dtB := makeAdminDataWithSiblings("B", "root", "", "C", "A", "dt2", "Block", "block")
	cdC, dtC := makeAdminDataWithSiblings("C", "root", "", "", "B", "dt3", "Block", "block")

	root := tree.NewRoot()
	_, err := root.LoadFromAdminData(
		[]db.AdminContentData{cdRoot, cdA, cdB, cdC},
		[]db.AdminDatatypes{dtRoot, dtA, dtB, dtC},
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// FlattenVisible should return: root, A, B, C
	visible := root.FlattenVisible()
	if len(visible) != 4 {
		t.Fatalf("FlattenVisible len = %d, want 4", len(visible))
	}

	expectedOrder := []string{"root", "A", "B", "C"}
	for i, suffix := range expectedOrder {
		expectedID := types.ContentID(makeAdminContentID(suffix).String())
		if visible[i].Instance.ContentDataID != expectedID {
			t.Errorf("visible[%d] = %q, want %q", i, visible[i].Instance.ContentDataID, expectedID)
		}
	}

	// Verify PrevSibling pointers are tree.Node references
	nodeA := visible[1]
	nodeB := visible[2]
	nodeC := visible[3]

	if nodeB.PrevSibling != nodeA {
		t.Error("B.PrevSibling does not point to tree.Node A")
	}
	if nodeC.PrevSibling != nodeB {
		t.Error("C.PrevSibling does not point to tree.Node B")
	}
	if nodeA.NextSibling != nodeB {
		t.Error("A.NextSibling does not point to tree.Node B")
	}
	if nodeB.NextSibling != nodeC {
		t.Error("B.NextSibling does not point to tree.Node C")
	}
}

func TestLoadFromAdminData_PointerRelinking(t *testing.T) {
	// Verify that all pointer fields reference tree.Node, not core.Node.
	// Build: root -> parent -> child
	cdRoot, dtRoot := makeAdminData("root", "", "dt0", "Page", "root")
	cdParent, dtParent := makeAdminData("parent", "root", "dt1", "Section", "block")
	cdChild, dtChild := makeAdminData("child", "parent", "dt2", "Block", "block")

	root := tree.NewRoot()
	_, err := root.LoadFromAdminData(
		[]db.AdminContentData{cdRoot, cdParent, cdChild},
		[]db.AdminDatatypes{dtRoot, dtParent, dtChild},
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parentID := types.ContentID(makeAdminContentID("parent").String())
	childID := types.ContentID(makeAdminContentID("child").String())

	parentNode := root.NodeIndex[parentID]
	childNode := root.NodeIndex[childID]

	if parentNode == nil {
		t.Fatal("parent node not in NodeIndex")
	}
	if childNode == nil {
		t.Fatal("child node not in NodeIndex")
	}

	// Parent's FirstChild should be the tree.Node child
	if parentNode.FirstChild != childNode {
		t.Error("parent.FirstChild does not point to tree.Node child")
	}
	// Child's Parent should be the tree.Node parent
	if childNode.Parent != parentNode {
		t.Error("child.Parent does not point to tree.Node parent")
	}
	// Parent's Parent should be root
	if parentNode.Parent != root.Root {
		t.Error("parent.Parent does not point to tree.Root node")
	}

	// Verify CoreNode is set and is distinct from the tree.Node
	if parentNode.CoreNode == nil {
		t.Error("parent.CoreNode is nil")
	}
	if childNode.CoreNode == nil {
		t.Error("child.CoreNode is nil")
	}
}

func TestLoadFromAdminData_WithFields(t *testing.T) {
	cd1, dt1 := makeAdminData("root", "", "dt1", "Page", "root")

	cf := []db.AdminContentFields{
		{
			AdminContentFieldID: types.AdminContentFieldID("01JCCCCCCCCCCCCCCCCCCCCCCC"),
			AdminContentDataID:  nullAdminContentID(makeAdminContentID("root")),
			AdminFieldID:        types.NullableAdminFieldID{ID: types.AdminFieldID("01JDDDDDDDDDDDDDDDDDDDDDDD"), Valid: true},
			AdminFieldValue:     "hello world",
		},
	}
	df := []db.AdminFields{
		{
			AdminFieldID: types.AdminFieldID("01JDDDDDDDDDDDDDDDDDDDDDDD"),
			Label:        "Title",
			Type:         types.FieldTypeText,
		},
	}

	root := tree.NewRoot()
	_, err := root.LoadFromAdminData(
		[]db.AdminContentData{cd1},
		[]db.AdminDatatypes{dt1},
		cf, df,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root.Root == nil {
		t.Fatal("Root is nil")
	}
	if len(root.Root.InstanceFields) != 1 {
		t.Fatalf("InstanceFields len = %d, want 1", len(root.Root.InstanceFields))
	}
	if root.Root.InstanceFields[0].FieldValue != "hello world" {
		t.Errorf("FieldValue = %q, want %q", root.Root.InstanceFields[0].FieldValue, "hello world")
	}
	if len(root.Root.Fields) != 1 {
		t.Fatalf("Fields len = %d, want 1", len(root.Root.Fields))
	}
	if root.Root.Fields[0].Label != "Title" {
		t.Errorf("Field Label = %q, want %q", root.Root.Fields[0].Label, "Title")
	}
}

func TestLoadFromAdminData_CountVisibleAndNodeAtIndex(t *testing.T) {
	// Build: root -> [A, B]
	cdRoot, dtRoot := makeAdminData("root", "", "dt0", "Page", "root")
	cdA, dtA := makeAdminData("A", "root", "dt1", "Block", "block")
	cdB, dtB := makeAdminData("B", "root", "dt2", "Block", "block")

	root := tree.NewRoot()
	_, err := root.LoadFromAdminData(
		[]db.AdminContentData{cdRoot, cdA, cdB},
		[]db.AdminDatatypes{dtRoot, dtA, dtB},
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := root.CountVisible(); got != 3 {
		t.Errorf("CountVisible = %d, want 3", got)
	}

	// NodeAtIndex(0) should be root
	node0 := root.NodeAtIndex(0)
	if node0 == nil || node0 != root.Root {
		t.Error("NodeAtIndex(0) is not root")
	}

	// FindVisibleIndex for root should be 0
	rootID := types.ContentID(makeAdminContentID("root").String())
	if idx := root.FindVisibleIndex(rootID); idx != 0 {
		t.Errorf("FindVisibleIndex(root) = %d, want 0", idx)
	}
}

func TestLoadFromAdminData_OrphanError(t *testing.T) {
	// A node whose parent doesn't exist should produce an error (orphan warning)
	// but the tree should still be usable.
	cdRoot, dtRoot := makeAdminData("root", "", "dt0", "Page", "root")
	cdOrphan, dtOrphan := makeAdminData("orphan", "missing", "dt1", "Block", "block")

	root := tree.NewRoot()
	stats, err := root.LoadFromAdminData(
		[]db.AdminContentData{cdRoot, cdOrphan},
		[]db.AdminDatatypes{dtRoot, dtOrphan},
		nil, nil,
	)
	if err == nil {
		t.Fatal("expected orphan error, got nil")
	}

	// Tree should still be usable
	if root.Root == nil {
		t.Fatal("Root is nil despite orphan")
	}
	if stats.NodesCount != 2 {
		t.Errorf("NodesCount = %d, want 2", stats.NodesCount)
	}
}

func TestLoadFromAdminData_EmptySlices(t *testing.T) {
	root := tree.NewRoot()
	stats, err := root.LoadFromAdminData(nil, nil, nil, nil)
	if err == nil {
		t.Fatal("expected 'no root node found' error, got nil")
	}

	if stats.NodesCount != 0 {
		t.Errorf("NodesCount = %d, want 0", stats.NodesCount)
	}
	if root.Root != nil {
		t.Error("Root should be nil for empty input")
	}
}
