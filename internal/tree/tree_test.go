package tree_test

import (
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// makeContentID creates a deterministic ContentID for test use.
// Uses a fixed ULID prefix with the last characters replaced by the suffix.
// This avoids the overhead of NewContentID() and makes test output readable.
func makeContentID(suffix string) types.ContentID {
	// ULID is 26 chars. Pad suffix to 26 chars with leading zeros.
	base := "01JAAAAAAAAAAAAAAAAAAAAAAA"
	if len(suffix) > 26 {
		suffix = suffix[:26]
	}
	id := base[:26-len(suffix)] + suffix
	return types.ContentID(id)
}

// nullID wraps a ContentID in a valid NullableContentID.
func nullID(id types.ContentID) types.NullableContentID {
	return types.NullableContentID{ID: id, Valid: true}
}

// emptyNullID returns an invalid NullableContentID (no parent).
func emptyNullID() types.NullableContentID {
	return types.NullableContentID{Valid: false}
}

// makeRow builds a GetContentTreeByRouteRow with the given IDs.
// parentID == "" means root node (no parent).
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

// makeRowWithSiblings builds a row with explicit sibling pointer data for
// testing the reorderByPointers phase.
func makeRowWithSiblings(id, parentID, firstChildID, nextSiblingID, prevSiblingID string) db.GetContentTreeByRouteRow {
	row := makeRow(id, parentID)
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

// collectSiblingIDs walks the NextSibling chain from the given node
// (including the node itself) and returns ContentDataIDs in order.
func collectSiblingIDs(node *tree.Node) []types.ContentID {
	var ids []types.ContentID
	current := node
	for current != nil {
		ids = append(ids, current.Instance.ContentDataID)
		current = current.NextSibling
	}
	return ids
}

// buildSimpleTree constructs a tree via LoadFromRows from the given rows,
// failing the test if an unexpected error occurs.
func buildSimpleTree(t *testing.T, rows []db.GetContentTreeByRouteRow) *tree.Root {
	t.Helper()
	root := tree.NewRoot()
	_, err := root.LoadFromRows(&rows)
	if err != nil {
		t.Fatalf("LoadFromRows failed: %v", err)
	}
	return root
}

// makeNode creates a standalone Node with the given ContentDataID.
func makeNode(id string) *tree.Node {
	return &tree.Node{
		Instance: &db.ContentData{
			ContentDataID: makeContentID(id),
		},
		Expand: true,
	}
}

// makeNodeWithParent creates a Node with a valid ParentID set in Instance.
func makeNodeWithParent(id, parentID string) *tree.Node {
	return &tree.Node{
		Instance: &db.ContentData{
			ContentDataID: makeContentID(id),
			ParentID:      nullID(makeContentID(parentID)),
		},
		Expand: true,
	}
}

// ---------------------------------------------------------------------------
// NewRoot tests
// ---------------------------------------------------------------------------

func TestNewRoot(t *testing.T) {
	t.Parallel()
	root := tree.NewRoot()

	if root.Root != nil {
		t.Errorf("expected Root to be nil, got %v", root.Root)
	}
	if root.NodeIndex == nil {
		t.Fatal("expected NodeIndex to be initialized, got nil")
	}
	if root.Orphans == nil {
		t.Fatal("expected Orphans to be initialized, got nil")
	}
	if root.MaxRetry != 100 {
		t.Errorf("expected MaxRetry=100, got %d", root.MaxRetry)
	}
	if len(root.NodeIndex) != 0 {
		t.Errorf("expected empty NodeIndex, got %d entries", len(root.NodeIndex))
	}
	if len(root.Orphans) != 0 {
		t.Errorf("expected empty Orphans, got %d entries", len(root.Orphans))
	}
}

// ---------------------------------------------------------------------------
// NewNode tests
// ---------------------------------------------------------------------------

func TestNewNode(t *testing.T) {
	t.Parallel()
	cid := makeContentID("A1")
	pid := makeContentID("P1")
	row := db.GetRouteTreeByRouteIDRow{
		ContentDataID: cid,
		ParentID:      nullID(pid),
	}
	node := tree.NewNode(row)

	if node.Instance == nil {
		t.Fatal("expected Instance to be set")
	}
	if node.Instance.ContentDataID != cid {
		t.Errorf("expected ContentDataID=%s, got %s", cid, node.Instance.ContentDataID)
	}
	if node.Instance.ParentID.ID != pid {
		t.Errorf("expected ParentID=%s, got %s", pid, node.Instance.ParentID.ID)
	}
	if !node.Expand {
		t.Error("expected Expand=true by default")
	}
}

func TestNewNode_RootNode(t *testing.T) {
	t.Parallel()
	row := db.GetRouteTreeByRouteIDRow{
		ContentDataID: makeContentID("R1"),
		ParentID:      emptyNullID(),
	}
	node := tree.NewNode(row)

	if node.Instance.ParentID.Valid {
		t.Error("expected ParentID.Valid=false for root node")
	}
}

// ---------------------------------------------------------------------------
// NewNodeFromContentTree tests
// ---------------------------------------------------------------------------

func TestNewNodeFromContentTree(t *testing.T) {
	t.Parallel()
	row := db.GetContentTreeByRouteRow{
		ContentDataID: makeContentID("CT1"),
		ParentID:      nullID(makeContentID("P1")),
		FirstChildID:  nullID(makeContentID("FC1")),
		NextSiblingID: nullID(makeContentID("NS1")),
		PrevSiblingID: nullID(makeContentID("PS1")),
		DatatypeLabel: "page",
		DatatypeType:  "container",
		Status:        types.ContentStatus("published"),
	}

	node := tree.NewNodeFromContentTree(row)

	if node.Instance == nil {
		t.Fatal("expected Instance to be set")
	}
	if node.Instance.ContentDataID != row.ContentDataID {
		t.Errorf("ContentDataID mismatch: got %s, want %s", node.Instance.ContentDataID, row.ContentDataID)
	}
	if node.Instance.FirstChildID.ID != row.FirstChildID.ID {
		t.Errorf("FirstChildID mismatch: got %s, want %s", node.Instance.FirstChildID.ID, row.FirstChildID.ID)
	}
	if node.Instance.NextSiblingID.ID != row.NextSiblingID.ID {
		t.Errorf("NextSiblingID mismatch")
	}
	if node.Instance.PrevSiblingID.ID != row.PrevSiblingID.ID {
		t.Errorf("PrevSiblingID mismatch")
	}
	if node.Datatype.Label != "page" {
		t.Errorf("Datatype.Label mismatch: got %q, want %q", node.Datatype.Label, "page")
	}
	if node.Datatype.Type != "container" {
		t.Errorf("Datatype.Type mismatch: got %q, want %q", node.Datatype.Type, "container")
	}
	if node.Instance.Status != types.ContentStatus("published") {
		t.Errorf("Status mismatch")
	}
	if !node.Expand {
		t.Error("expected Expand=true by default")
	}
}

// ---------------------------------------------------------------------------
// LoadStats.String tests
// ---------------------------------------------------------------------------

func TestLoadStats_String(t *testing.T) {
	t.Parallel()
	stats := tree.LoadStats{
		NodesCount:      5,
		OrphansResolved: 2,
		RetryAttempts:   3,
		CircularRefs:    []types.ContentID{makeContentID("C1")},
		FinalOrphans:    []types.ContentID{makeContentID("O1")},
	}
	s := stats.String()

	if !strings.Contains(s, "5") {
		t.Errorf("expected nodes count in string, got %q", s)
	}
	if !strings.Contains(s, "2") {
		t.Errorf("expected orphans resolved in string, got %q", s)
	}
	if !strings.Contains(s, "3") {
		t.Errorf("expected retry attempts in string, got %q", s)
	}
}

// ---------------------------------------------------------------------------
// LoadFromRows tests
// ---------------------------------------------------------------------------

func TestLoadFromRows_EmptyRows(t *testing.T) {
	t.Parallel()
	root := tree.NewRoot()
	rows := []db.GetContentTreeByRouteRow{}

	stats, err := root.LoadFromRows(&rows)
	// Empty rows means no root node found
	if err == nil {
		t.Fatal("expected error for empty rows, got nil")
	}
	if !strings.Contains(err.Error(), "no root node found") {
		t.Errorf("expected 'no root node' error, got: %v", err)
	}
	if stats.NodesCount != 0 {
		t.Errorf("expected 0 nodes, got %d", stats.NodesCount)
	}
}

func TestLoadFromRows_SingleRoot(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
	}
	root := tree.NewRoot()
	stats, err := root.LoadFromRows(&rows)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Root == nil {
		t.Fatal("expected Root to be set")
	}
	if root.Root.Instance.ContentDataID != makeContentID("R1") {
		t.Errorf("root ID mismatch")
	}
	if stats.NodesCount != 1 {
		t.Errorf("expected 1 node, got %d", stats.NodesCount)
	}
	if len(root.NodeIndex) != 1 {
		t.Errorf("expected 1 entry in NodeIndex, got %d", len(root.NodeIndex))
	}
}

func TestLoadFromRows_ParentChildHierarchy(t *testing.T) {
	t.Parallel()
	// root -> child1, child2
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("C1", "R1"),
		makeRow("C2", "R1"),
	}
	root := buildSimpleTree(t, rows)

	if root.Root.FirstChild == nil {
		t.Fatal("expected root to have children")
	}
	// Both children should be attached to root
	childIDs := collectSiblingIDs(root.Root.FirstChild)
	if len(childIDs) != 2 {
		t.Fatalf("expected 2 children, got %d", len(childIDs))
	}

	// Verify parent pointers
	child := root.Root.FirstChild
	for child != nil {
		if child.Parent != root.Root {
			t.Errorf("child %s parent pointer does not point to root", child.Instance.ContentDataID)
		}
		child = child.NextSibling
	}
}

func TestLoadFromRows_DeepHierarchy(t *testing.T) {
	t.Parallel()
	// root -> child -> grandchild -> great-grandchild
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("C1", "R1"),
		makeRow("G1", "C1"),
		makeRow("GG1", "G1"),
	}
	root := buildSimpleTree(t, rows)

	if root.Root.FirstChild == nil {
		t.Fatal("root should have a child")
	}
	child := root.Root.FirstChild
	if child.Instance.ContentDataID != makeContentID("C1") {
		t.Errorf("first child ID mismatch")
	}
	if child.FirstChild == nil {
		t.Fatal("child should have a grandchild")
	}
	grandchild := child.FirstChild
	if grandchild.Instance.ContentDataID != makeContentID("G1") {
		t.Errorf("grandchild ID mismatch")
	}
	if grandchild.FirstChild == nil {
		t.Fatal("grandchild should have a great-grandchild")
	}
	ggchild := grandchild.FirstChild
	if ggchild.Instance.ContentDataID != makeContentID("GG1") {
		t.Errorf("great-grandchild ID mismatch")
	}
}

func TestLoadFromRows_SiblingPointers(t *testing.T) {
	t.Parallel()
	// Verify doubly-linked sibling list: root -> A, B, C
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
		makeRow("CC", "R1"),
	}
	root := buildSimpleTree(t, rows)

	first := root.Root.FirstChild
	if first == nil {
		t.Fatal("expected children")
	}

	// Walk forward
	var nodes []*tree.Node
	current := first
	for current != nil {
		nodes = append(nodes, current)
		current = current.NextSibling
	}
	if len(nodes) != 3 {
		t.Fatalf("expected 3 siblings, got %d", len(nodes))
	}

	// Verify PrevSibling pointers
	if nodes[0].PrevSibling != nil {
		t.Error("first sibling should have nil PrevSibling")
	}
	if nodes[1].PrevSibling != nodes[0] {
		t.Error("second sibling PrevSibling should point to first")
	}
	if nodes[2].PrevSibling != nodes[1] {
		t.Error("third sibling PrevSibling should point to second")
	}
	if nodes[2].NextSibling != nil {
		t.Error("last sibling should have nil NextSibling")
	}
}

func TestLoadFromRows_OrphanWithMissingParent(t *testing.T) {
	t.Parallel()
	// Node "C1" references parent "MISSING" which does not exist in the rows
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("C1", "MISSING"),
	}
	root := tree.NewRoot()
	stats, err := root.LoadFromRows(&rows)

	if err == nil {
		t.Fatal("expected error for unresolved orphan")
	}
	if !strings.Contains(err.Error(), "unresolved orphan") {
		t.Errorf("expected 'unresolved orphan' error, got: %v", err)
	}
	if len(stats.FinalOrphans) != 1 {
		t.Errorf("expected 1 final orphan, got %d", len(stats.FinalOrphans))
	}
}

func TestLoadFromRows_MutualParentReferences(t *testing.T) {
	t.Parallel()
	// A -> B -> A (mutual parent references)
	// Both A and B exist in the NodeIndex, so phase 2 attaches them to
	// each other (A's parent = B, B's parent = A). This creates an in-memory
	// cycle in the parent chain. The code does NOT flag this as an error
	// because circular reference detection only runs during orphan resolution
	// (phase 3), and these nodes are never orphaned.
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),   // real root
		makeRow("AA", "BB"), // A's parent is B
		makeRow("BB", "AA"), // B's parent is A
	}
	root := tree.NewRoot()
	stats, err := root.LoadFromRows(&rows)

	// Both nodes get attached in phase 2 (parents exist in index),
	// so no orphans and no circular ref detection fires.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.NodesCount != 3 {
		t.Errorf("expected 3 nodes, got %d", stats.NodesCount)
	}
	// Both A and B are in the index
	if root.NodeIndex[makeContentID("AA")] == nil {
		t.Error("AA should be in index")
	}
	if root.NodeIndex[makeContentID("BB")] == nil {
		t.Error("BB should be in index")
	}
}

func TestLoadFromRows_OrphanCircularDetection(t *testing.T) {
	t.Parallel()
	// Orphan resolution circular detection:
	// Nodes A and B both reference parents that don't exist in the rows,
	// making them true orphans. Circular refs only fire for orphans.
	// With missing parents, hasCircularReference returns false (parent==nil),
	// so they remain as unresolved orphans, not circular refs.
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),        // real root
		makeRow("AA", "MISSING1"), // parent not in rows
		makeRow("BB", "MISSING2"), // parent not in rows
	}
	root := tree.NewRoot()
	stats, err := root.LoadFromRows(&rows)

	if err == nil {
		t.Fatal("expected error for unresolved orphans")
	}
	if !strings.Contains(err.Error(), "unresolved orphan") {
		t.Errorf("expected 'unresolved orphan' error, got: %v", err)
	}
	if len(stats.FinalOrphans) != 2 {
		t.Errorf("expected 2 final orphans, got %d", len(stats.FinalOrphans))
	}
}

func TestLoadFromRows_MultipleRoots(t *testing.T) {
	t.Parallel()
	// Two parentless nodes -- both should be indexed, first becomes Root
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("R2", ""),
	}
	root := tree.NewRoot()
	_, err := root.LoadFromRows(&rows)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root.Root == nil {
		t.Fatal("expected a root node")
	}
	// Both nodes should be in the index
	if len(root.NodeIndex) != 2 {
		t.Errorf("expected 2 nodes in index, got %d", len(root.NodeIndex))
	}
}

// ---------------------------------------------------------------------------
// LoadFromRows with sibling pointer reordering (Phase 4)
// ---------------------------------------------------------------------------

func TestLoadFromRows_ReorderByPointers(t *testing.T) {
	t.Parallel()
	// Root has children A, B, C. The stored sibling order is C -> A -> B.
	// Phase 2 will attach them in map-iteration order (random),
	// but Phase 4 should reorder them to C, A, B.
	cID := makeContentID("CC")
	aID := makeContentID("AA")
	bID := makeContentID("BB")

	rows := []db.GetContentTreeByRouteRow{
		// root with FirstChildID = C
		makeRowWithSiblings("R1", "", "CC", "", ""),
		// C is first, next is A
		makeRowWithSiblings("CC", "R1", "", "AA", ""),
		// A is middle, prev is C, next is B
		makeRowWithSiblings("AA", "R1", "", "BB", "CC"),
		// B is last, prev is A
		makeRowWithSiblings("BB", "R1", "", "", "AA"),
	}

	root := buildSimpleTree(t, rows)

	// Verify the reordered child chain: C -> A -> B
	first := root.Root.FirstChild
	if first == nil {
		t.Fatal("expected children")
	}
	if first.Instance.ContentDataID != cID {
		t.Errorf("first child: got %s, want %s", first.Instance.ContentDataID, cID)
	}

	second := first.NextSibling
	if second == nil {
		t.Fatal("expected second child")
	}
	if second.Instance.ContentDataID != aID {
		t.Errorf("second child: got %s, want %s", second.Instance.ContentDataID, aID)
	}

	third := second.NextSibling
	if third == nil {
		t.Fatal("expected third child")
	}
	if third.Instance.ContentDataID != bID {
		t.Errorf("third child: got %s, want %s", third.Instance.ContentDataID, bID)
	}

	if third.NextSibling != nil {
		t.Error("last child should have nil NextSibling")
	}

	// Verify PrevSibling chain
	if first.PrevSibling != nil {
		t.Error("first child should have nil PrevSibling")
	}
	if second.PrevSibling != first {
		t.Error("second child PrevSibling should point to first")
	}
	if third.PrevSibling != second {
		t.Error("third child PrevSibling should point to second")
	}
}

func TestLoadFromRows_ReorderRootSiblings(t *testing.T) {
	t.Parallel()
	// Two root nodes: R1 and R2. Stored order: R2 -> R1.
	rows := []db.GetContentTreeByRouteRow{
		makeRowWithSiblings("R1", "", "", "", "R2"),
		makeRowWithSiblings("R2", "", "", "R1", ""),
	}

	root := tree.NewRoot()
	_, err := root.LoadFromRows(&rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// R2 should be first (PrevSiblingID is empty => it is the head)
	if root.Root.Instance.ContentDataID != makeContentID("R2") {
		t.Errorf("expected R2 as root after reorder, got %s", root.Root.Instance.ContentDataID)
	}
	if root.Root.NextSibling == nil {
		t.Fatal("expected R2 to have a next sibling")
	}
	if root.Root.NextSibling.Instance.ContentDataID != makeContentID("R1") {
		t.Errorf("expected R1 as second root, got %s", root.Root.NextSibling.Instance.ContentDataID)
	}
}

// ---------------------------------------------------------------------------
// InsertNodeByIndex tests
// ---------------------------------------------------------------------------

func TestInsertNodeByIndex(t *testing.T) {
	t.Parallel()
	root := tree.NewRoot()

	parent := makeNode("P1")
	root.NodeIndex[parent.Instance.ContentDataID] = parent
	root.Root = parent

	child := makeNode("C1")
	root.InsertNodeByIndex(parent, nil, nil, nil, child)

	if root.NodeIndex[makeContentID("C1")] != child {
		t.Error("child should be in NodeIndex")
	}
	if child.Parent != parent {
		t.Error("child.Parent should point to parent")
	}
	if child.FirstChild != nil {
		t.Error("child.FirstChild should be nil")
	}
}

func TestInsertNodeByIndex_WithSiblings(t *testing.T) {
	t.Parallel()
	root := tree.NewRoot()

	parent := makeNode("P1")
	root.NodeIndex[parent.Instance.ContentDataID] = parent
	root.Root = parent

	prev := makeNode("PREV")
	next := makeNode("NEXT")
	node := makeNode("MID")

	root.InsertNodeByIndex(parent, nil, prev, next, node)

	if node.PrevSibling != prev {
		t.Error("node.PrevSibling should be prev")
	}
	if node.NextSibling != next {
		t.Error("node.NextSibling should be next")
	}
}

// ---------------------------------------------------------------------------
// DeleteNodeByIndex tests
// ---------------------------------------------------------------------------

func TestDeleteNodeByIndex_NilRoot(t *testing.T) {
	t.Parallel()
	root := tree.NewRoot()
	node := makeNode("X1")
	if root.DeleteNodeByIndex(node) {
		t.Error("should return false when root is nil")
	}
}

func TestDeleteNodeByIndex_CannotDeleteRoot(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
	}
	root := buildSimpleTree(t, rows)

	if root.DeleteNodeByIndex(root.Root) {
		t.Error("should not allow deleting the root node")
	}
}

func TestDeleteNodeByIndex_DeleteFirstChildNoChildren(t *testing.T) {
	t.Parallel()
	// root -> A, B ; delete A
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
	}
	root := buildSimpleTree(t, rows)

	firstChild := root.Root.FirstChild
	if firstChild == nil {
		t.Fatal("expected first child")
	}
	firstID := firstChild.Instance.ContentDataID

	ok := root.DeleteNodeByIndex(firstChild)
	if !ok {
		t.Fatal("expected successful deletion")
	}

	// Node should be removed from index
	if root.NodeIndex[firstID] != nil {
		t.Error("deleted node should be removed from NodeIndex")
	}

	// Root's first child should now be the other child
	if root.Root.FirstChild == nil {
		t.Fatal("root should still have a child")
	}
}

func TestDeleteNodeByIndex_DeleteFirstChildWithChildren(t *testing.T) {
	t.Parallel()
	// root -> A -> (A1, A2); delete A -- children should be promoted
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("A1", "AA"),
		makeRow("A2", "AA"),
	}
	root := buildSimpleTree(t, rows)

	nodeA := root.NodeIndex[makeContentID("AA")]
	if nodeA == nil {
		t.Fatal("node AA not found")
	}

	ok := root.DeleteNodeByIndex(nodeA)
	if !ok {
		t.Fatal("expected successful deletion")
	}

	// A's children should now be children of root
	if root.Root.FirstChild == nil {
		t.Fatal("root should have children after promoting A's children")
	}

	// A should be removed from index
	if root.NodeIndex[makeContentID("AA")] != nil {
		t.Error("deleted node should be removed from NodeIndex")
	}
}

func TestDeleteNodeByIndex_DeleteMiddleSibling(t *testing.T) {
	t.Parallel()
	// root -> A, B, C; delete B (not first child, no children)
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
		makeRow("CC", "R1"),
	}
	root := buildSimpleTree(t, rows)

	// Find node B
	nodeB := root.NodeIndex[makeContentID("BB")]
	if nodeB == nil {
		t.Fatal("node BB not found")
	}

	// Ensure B is not the first child (it could be due to map ordering)
	// We need to find a non-first-child node to test this path.
	// If B happens to be first child, swap our target.
	target := nodeB
	if root.Root.FirstChild == nodeB {
		// B is first child, use C instead
		nodeC := root.NodeIndex[makeContentID("CC")]
		if nodeC == nil || root.Root.FirstChild == nodeC {
			// Both are first -- this means the tree only has 1 child (shouldn't happen)
			t.Skip("cannot set up middle sibling scenario due to map ordering")
		}
		target = nodeC
	}

	targetID := target.Instance.ContentDataID
	ok := root.DeleteNodeByIndex(target)
	if !ok {
		t.Fatal("expected successful deletion")
	}

	if root.NodeIndex[targetID] != nil {
		t.Error("deleted node should be removed from NodeIndex")
	}
}

func TestDeleteNodeByIndex_DeleteLastSibling(t *testing.T) {
	t.Parallel()
	// root -> A, B; delete last sibling (B if it's last)
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
	}
	root := buildSimpleTree(t, rows)

	// Walk to the last sibling
	current := root.Root.FirstChild
	for current.NextSibling != nil {
		current = current.NextSibling
	}
	lastID := current.Instance.ContentDataID

	ok := root.DeleteNodeByIndex(current)
	if !ok {
		t.Fatal("expected successful deletion of last sibling")
	}
	if root.NodeIndex[lastID] != nil {
		t.Error("deleted node should be removed from NodeIndex")
	}
}

func TestDeleteNodeByIndex_NodeNotInTree(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
	}
	root := buildSimpleTree(t, rows)

	outsider := makeNode("ZZ")
	if root.DeleteNodeByIndex(outsider) {
		t.Error("should return false for node not in index")
	}
}

func TestDeleteNodeByIndex_NodeWithNilParent(t *testing.T) {
	t.Parallel()
	// A parentless non-root node added directly to index
	root := tree.NewRoot()
	rootNode := makeNode("R1")
	root.Root = rootNode
	root.NodeIndex[rootNode.Instance.ContentDataID] = rootNode

	orphan := makeNode("O1")
	orphan.Parent = nil // explicitly nil parent
	root.NodeIndex[orphan.Instance.ContentDataID] = orphan

	if root.DeleteNodeByIndex(orphan) {
		t.Error("should return false for node with nil parent (other than root)")
	}
}

// ---------------------------------------------------------------------------
// DeleteFirstChild / DeleteNestedChild (exported helpers)
// ---------------------------------------------------------------------------

func TestDeleteFirstChild_OnlyChild(t *testing.T) {
	t.Parallel()
	parent := makeNode("P1")
	child := makeNode("C1")

	parent.FirstChild = child
	child.Parent = parent

	ok := tree.DeleteFirstChild(child)
	if !ok {
		t.Error("expected true")
	}
	if parent.FirstChild != nil {
		t.Error("parent should have no children after deleting only child")
	}
}

func TestDeleteFirstChild_WithNextSibling(t *testing.T) {
	t.Parallel()
	parent := makeNode("P1")
	child := makeNode("C1")
	sibling := makeNode("C2")

	parent.FirstChild = child
	child.Parent = parent
	child.NextSibling = sibling
	sibling.PrevSibling = child
	sibling.Parent = parent

	ok := tree.DeleteFirstChild(child)
	if !ok {
		t.Error("expected true")
	}
	if parent.FirstChild != sibling {
		t.Error("parent's first child should now be sibling")
	}
	if sibling.PrevSibling != nil {
		t.Error("promoted sibling should have nil PrevSibling")
	}
}

func TestDeleteFirstChild_HasChildrenAndNextSibling(t *testing.T) {
	t.Parallel()
	// parent -> target -> (gc1, gc2), target.NextSibling = sibling
	parent := makeNode("P1")
	target := makeNode("T1")
	gc1 := makeNode("G1")
	gc2 := makeNode("G2")
	sibling := makeNode("S1")

	parent.FirstChild = target
	target.Parent = parent
	target.FirstChild = gc1
	target.NextSibling = sibling
	sibling.PrevSibling = target
	sibling.Parent = parent

	gc1.Parent = target
	gc1.NextSibling = gc2
	gc2.PrevSibling = gc1
	gc2.Parent = target

	ok := tree.DeleteFirstChild(target)
	if !ok {
		t.Error("expected true")
	}

	// gc1 (and gc2) should be promoted to parent's children
	if parent.FirstChild != gc1 {
		t.Error("parent's first child should be gc1")
	}
	// gc2's NextSibling should be sibling
	if gc2.NextSibling != sibling {
		t.Error("gc2's NextSibling should be the old sibling")
	}
	if sibling.PrevSibling != gc2 {
		t.Error("sibling's PrevSibling should be gc2")
	}
}

func TestDeleteNestedChild_NoChildrenWithSiblings(t *testing.T) {
	t.Parallel()
	parent := makeNode("P1")
	prev := makeNode("PREV")
	target := makeNode("TGT")
	next := makeNode("NEXT")

	parent.FirstChild = prev
	prev.Parent = parent
	prev.NextSibling = target
	target.PrevSibling = prev
	target.Parent = parent
	target.NextSibling = next
	next.PrevSibling = target
	next.Parent = parent

	ok := tree.DeleteNestedChild(target)
	if !ok {
		t.Error("expected true")
	}
	if prev.NextSibling != next {
		t.Error("prev's NextSibling should skip to next")
	}
	if next.PrevSibling != prev {
		t.Error("next's PrevSibling should skip to prev")
	}
}

func TestDeleteNestedChild_NoChildrenLastSibling(t *testing.T) {
	t.Parallel()
	parent := makeNode("P1")
	prev := makeNode("PREV")
	target := makeNode("TGT")

	parent.FirstChild = prev
	prev.Parent = parent
	prev.NextSibling = target
	target.PrevSibling = prev
	target.Parent = parent

	ok := tree.DeleteNestedChild(target)
	if !ok {
		t.Error("expected true")
	}
	if prev.NextSibling != nil {
		t.Error("prev should now be the last sibling")
	}
}

func TestDeleteNestedChild_NoPrevSibling_ReturnsFalse(t *testing.T) {
	t.Parallel()
	// Edge case: nested child with no PrevSibling should return false
	parent := makeNode("P1")
	target := makeNode("TGT")
	next := makeNode("NEXT")

	parent.FirstChild = target
	target.Parent = parent
	target.NextSibling = next
	next.PrevSibling = target
	next.Parent = parent
	// PrevSibling is nil -- this means it IS the first child,
	// but DeleteNestedChild is being called directly (not through DeleteNodeByIndex)
	target.PrevSibling = nil

	ok := tree.DeleteNestedChild(target)
	if ok {
		t.Error("expected false when PrevSibling is nil for nested child with NextSibling")
	}
}

// ---------------------------------------------------------------------------
// CountVisible tests
// ---------------------------------------------------------------------------

func TestCountVisible_NilRoot(t *testing.T) {
	t.Parallel()
	root := tree.NewRoot()
	if root.CountVisible() != 0 {
		t.Error("expected 0 for nil root")
	}
}

func TestCountVisible_AllExpanded(t *testing.T) {
	t.Parallel()
	// root -> A, B; A -> A1
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
		makeRow("A1", "AA"),
	}
	root := buildSimpleTree(t, rows)

	// All nodes have Expand=true by default
	count := root.CountVisible()
	if count != 4 {
		t.Errorf("expected 4 visible nodes, got %d", count)
	}
}

func TestCountVisible_CollapsedNode(t *testing.T) {
	t.Parallel()
	// root -> A -> (A1, A2); collapse A
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("A1", "AA"),
		makeRow("A2", "AA"),
	}
	root := buildSimpleTree(t, rows)

	nodeA := root.NodeIndex[makeContentID("AA")]
	nodeA.Expand = false

	// root + A are visible, A1 and A2 are hidden
	count := root.CountVisible()
	if count != 2 {
		t.Errorf("expected 2 visible nodes after collapsing A, got %d", count)
	}
}

func TestCountVisible_DeeplyCollapsed(t *testing.T) {
	t.Parallel()
	// root -> A -> B -> C; collapse A
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "AA"),
		makeRow("CC", "BB"),
	}
	root := buildSimpleTree(t, rows)

	nodeA := root.NodeIndex[makeContentID("AA")]
	nodeA.Expand = false

	// Only root and A visible; B and C are hidden under collapsed A
	count := root.CountVisible()
	if count != 2 {
		t.Errorf("expected 2 visible after collapsing A, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// NodeAtIndex tests
// ---------------------------------------------------------------------------

func TestNodeAtIndex_NilRoot(t *testing.T) {
	t.Parallel()
	root := tree.NewRoot()
	if root.NodeAtIndex(0) != nil {
		t.Error("expected nil for nil root")
	}
}

func TestNodeAtIndex(t *testing.T) {
	t.Parallel()
	// root -> A, B; A -> A1
	// Visible order: root(0), A(1), A1(2), B(3)
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
		makeRow("A1", "AA"),
	}
	root := buildSimpleTree(t, rows)

	// We need to verify the order matches the tree structure.
	// Due to map ordering in phase 2, we use FlattenVisible to get the actual order.
	visible := root.FlattenVisible()

	tests := []struct {
		name  string
		index int
		want  types.ContentID
	}{
		{"index 0 is root", 0, visible[0].Instance.ContentDataID},
		{"index 1", 1, visible[1].Instance.ContentDataID},
		{"index 2", 2, visible[2].Instance.ContentDataID},
		{"index 3", 3, visible[3].Instance.ContentDataID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := root.NodeAtIndex(tt.index)
			if got == nil {
				t.Fatalf("expected node at index %d, got nil", tt.index)
			}
			if got.Instance.ContentDataID != tt.want {
				t.Errorf("index %d: got %s, want %s", tt.index, got.Instance.ContentDataID, tt.want)
			}
		})
	}
}

func TestNodeAtIndex_OutOfBounds(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
	}
	root := buildSimpleTree(t, rows)

	tests := []struct {
		name  string
		index int
	}{
		{"negative", -1},
		{"beyond count", 10},
		{"exactly at count", root.CountVisible()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := root.NodeAtIndex(tt.index)
			if got != nil {
				t.Errorf("expected nil for out-of-bounds index %d, got node %s", tt.index, got.Instance.ContentDataID)
			}
		})
	}
}

func TestNodeAtIndex_CollapsedSkipsChildren(t *testing.T) {
	t.Parallel()
	// root -> A -> (A1); A -> collapsed; B is sibling of A
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("A1", "AA"),
		makeRow("BB", "R1"),
	}
	root := buildSimpleTree(t, rows)

	nodeA := root.NodeIndex[makeContentID("AA")]
	nodeA.Expand = false

	// Visible: root(0), then the two children of root in order with A collapsed
	// The visible order depends on attachment order, but A1 should be hidden.
	visible := root.FlattenVisible()

	for _, v := range visible {
		if v.Instance.ContentDataID == makeContentID("A1") {
			t.Error("A1 should not be visible when A is collapsed")
		}
	}
}

// ---------------------------------------------------------------------------
// FlattenVisible tests
// ---------------------------------------------------------------------------

func TestFlattenVisible_NilRoot(t *testing.T) {
	t.Parallel()
	root := tree.NewRoot()
	result := root.FlattenVisible()
	if result != nil {
		t.Errorf("expected nil, got %d nodes", len(result))
	}
}

func TestFlattenVisible_SingleNode(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
	}
	root := buildSimpleTree(t, rows)

	visible := root.FlattenVisible()
	if len(visible) != 1 {
		t.Fatalf("expected 1 visible node, got %d", len(visible))
	}
	if visible[0].Instance.ContentDataID != makeContentID("R1") {
		t.Error("visible node should be root")
	}
}

func TestFlattenVisible_ConsistentWithCountVisible(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
		makeRow("A1", "AA"),
		makeRow("B1", "BB"),
	}
	root := buildSimpleTree(t, rows)

	visible := root.FlattenVisible()
	count := root.CountVisible()

	if len(visible) != count {
		t.Errorf("FlattenVisible returned %d nodes, CountVisible returned %d", len(visible), count)
	}
}

func TestFlattenVisible_ConsistentWithNodeAtIndex(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
		makeRow("A1", "AA"),
	}
	root := buildSimpleTree(t, rows)

	visible := root.FlattenVisible()
	for i, node := range visible {
		atIndex := root.NodeAtIndex(i)
		if atIndex != node {
			t.Errorf("FlattenVisible[%d] != NodeAtIndex(%d)", i, i)
		}
	}
}

func TestFlattenVisible_ExpandCollapse(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("A1", "AA"),
		makeRow("A2", "AA"),
	}
	root := buildSimpleTree(t, rows)

	// All expanded: 4 nodes
	all := root.FlattenVisible()
	if len(all) != 4 {
		t.Fatalf("expected 4 visible nodes, got %d", len(all))
	}

	// Collapse A: only root + A
	nodeA := root.NodeIndex[makeContentID("AA")]
	nodeA.Expand = false

	collapsed := root.FlattenVisible()
	if len(collapsed) != 2 {
		t.Errorf("expected 2 visible after collapse, got %d", len(collapsed))
	}

	// Re-expand A: back to 4
	nodeA.Expand = true
	reexpanded := root.FlattenVisible()
	if len(reexpanded) != 4 {
		t.Errorf("expected 4 visible after re-expand, got %d", len(reexpanded))
	}
}

// ---------------------------------------------------------------------------
// FindVisibleIndex tests
// ---------------------------------------------------------------------------

func TestFindVisibleIndex_NilRoot(t *testing.T) {
	t.Parallel()
	root := tree.NewRoot()
	if root.FindVisibleIndex(makeContentID("X1")) != -1 {
		t.Error("expected -1 for nil root")
	}
}

func TestFindVisibleIndex_Found(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
	}
	root := buildSimpleTree(t, rows)

	// Root should be at index 0
	idx := root.FindVisibleIndex(makeContentID("R1"))
	if idx != 0 {
		t.Errorf("expected root at index 0, got %d", idx)
	}

	// All nodes should have a valid index
	for id := range root.NodeIndex {
		idx := root.FindVisibleIndex(id)
		if idx == -1 {
			t.Errorf("node %s should be visible", id)
		}
	}
}

func TestFindVisibleIndex_NotFound(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
	}
	root := buildSimpleTree(t, rows)

	idx := root.FindVisibleIndex(makeContentID("NONEXISTENT"))
	if idx != -1 {
		t.Errorf("expected -1 for non-existent node, got %d", idx)
	}
}

func TestFindVisibleIndex_HiddenByCollapse(t *testing.T) {
	t.Parallel()
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("A1", "AA"),
	}
	root := buildSimpleTree(t, rows)

	// A1 should be visible initially
	idx := root.FindVisibleIndex(makeContentID("A1"))
	if idx == -1 {
		t.Fatal("A1 should be visible before collapse")
	}

	// Collapse A -- A1 becomes hidden
	nodeA := root.NodeIndex[makeContentID("AA")]
	nodeA.Expand = false

	idx = root.FindVisibleIndex(makeContentID("A1"))
	if idx != -1 {
		t.Errorf("A1 should be hidden after collapse, got index %d", idx)
	}
}

// ---------------------------------------------------------------------------
// IsDescendantOf tests
// ---------------------------------------------------------------------------

func TestIsDescendantOf(t *testing.T) {
	t.Parallel()

	grandparent := makeNode("GP")
	parent := makeNode("P1")
	child := makeNode("C1")
	unrelated := makeNode("U1")

	parent.Parent = grandparent
	child.Parent = parent

	tests := []struct {
		name     string
		node     *tree.Node
		ancestor *tree.Node
		want     bool
	}{
		{"child is descendant of parent", child, parent, true},
		{"child is descendant of grandparent", child, grandparent, true},
		{"parent is descendant of grandparent", parent, grandparent, true},
		{"grandparent is not descendant of child", grandparent, child, false},
		{"parent is not descendant of child", parent, child, false},
		{"node is not descendant of itself", child, child, false},
		{"unrelated is not descendant", unrelated, grandparent, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tree.IsDescendantOf(tt.node, tt.ancestor)
			if got != tt.want {
				t.Errorf("IsDescendantOf: got %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Integration: full tree lifecycle
// ---------------------------------------------------------------------------

func TestFullTreeLifecycle(t *testing.T) {
	t.Parallel()

	// Build a tree: root -> (A, B); A -> (A1, A2); B -> (B1)
	rows := []db.GetContentTreeByRouteRow{
		makeRow("R1", ""),
		makeRow("AA", "R1"),
		makeRow("BB", "R1"),
		makeRow("A1", "AA"),
		makeRow("A2", "AA"),
		makeRow("B1", "BB"),
	}

	root := tree.NewRoot()
	stats, err := root.LoadFromRows(&rows)
	if err != nil {
		t.Fatalf("LoadFromRows: %v", err)
	}

	// Verify stats
	if stats.NodesCount != 6 {
		t.Errorf("expected 6 nodes, got %d", stats.NodesCount)
	}
	if len(stats.FinalOrphans) != 0 {
		t.Errorf("expected 0 orphans, got %d", len(stats.FinalOrphans))
	}
	if len(stats.CircularRefs) != 0 {
		t.Errorf("expected 0 circular refs, got %d", len(stats.CircularRefs))
	}

	// All expanded: 6 visible
	if root.CountVisible() != 6 {
		t.Errorf("expected 6 visible, got %d", root.CountVisible())
	}

	// Collapse A: should hide A1 and A2
	nodeA := root.NodeIndex[makeContentID("AA")]
	nodeA.Expand = false
	if root.CountVisible() != 4 {
		t.Errorf("expected 4 visible after collapsing A, got %d", root.CountVisible())
	}

	// Re-expand
	nodeA.Expand = true

	// Delete A1 -- should succeed and reduce count
	nodeA1 := root.NodeIndex[makeContentID("A1")]
	if nodeA1 == nil {
		t.Fatal("A1 not in index")
	}
	ok := root.DeleteNodeByIndex(nodeA1)
	if !ok {
		t.Fatal("expected deletion of A1 to succeed")
	}
	if root.NodeIndex[makeContentID("A1")] != nil {
		t.Error("A1 should be removed from index")
	}
	if root.CountVisible() != 5 {
		t.Errorf("expected 5 visible after deleting A1, got %d", root.CountVisible())
	}

	// Insert a new node
	newNode := makeNodeWithParent("NN", "BB")
	nodeB := root.NodeIndex[makeContentID("BB")]
	root.InsertNodeByIndex(nodeB, nil, nil, nil, newNode)
	if root.NodeIndex[makeContentID("NN")] == nil {
		t.Error("new node should be in index")
	}

	// FindVisibleIndex for the new node
	// Note: InsertNodeByIndex does not wire into parent's child chain,
	// so it won't be visible through traversal. This tests the index-only path.
	if root.NodeIndex[makeContentID("NN")].Parent != nodeB {
		t.Error("new node should have nodeB as parent")
	}
}
