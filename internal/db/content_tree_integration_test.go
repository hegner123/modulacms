// Integration tests for the content tree lifecycle.
//
// These tests exercise the full CMS content tree pipeline against a real SQLite
// database: creating content_data rows with parent/sibling pointers, reading them
// back via GetContentTreeByRoute, building the tree with core.BuildFromRows, and
// verifying the resulting structure.
//
// This is what makes ModulaCMS a CMS -- the ability to store, retrieve, and
// re-assemble hierarchical content from the database.
//
// Package db_test (external) to avoid an import cycle: this file imports
// internal/tree/core which itself imports internal/db. Test helpers are
// re-exported via export_test.go.
package db_test

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree/core"
)

// ---------------------------------------------------------------------------
// Seed helpers: build a realistic content tree scenario
// ---------------------------------------------------------------------------

// contentTreeSeed holds all IDs needed to verify a content tree scenario.
type contentTreeSeed struct {
	DB       db.Database
	AC       audited.AuditContext
	User     *db.Users
	RouteID  types.NullableRouteID
	DtRootID types.NullableDatatypeID // _root datatype
	DtPageID types.NullableDatatypeID // page datatype
	FieldID  types.NullableFieldID    // text field for content_fields
}

// seedContentTreeBase creates the prerequisite entities (role, user, route,
// datatypes, field) needed for content tree tests. Returns everything needed
// to start creating content_data.
func seedContentTreeBase(t *testing.T) contentTreeSeed {
	t.Helper()
	d := db.ExportedIntegrationDB(t)
	ctx := d.Context
	ac := db.ExportedAuditCtx(d)
	now := types.TimestampNow()

	role, err := d.CreateRole(ctx, ac, db.CreateRoleParams{Label: "tree-test-role"})
	if err != nil {
		t.Fatalf("seed CreateRole: %v", err)
	}

	user, err := d.CreateUser(ctx, ac, db.CreateUserParams{
		Username:     "treeuser",
		Name:         "Tree User",
		Email:        types.Email("tree@test.com"),
		Hash:         "fakehash",
		Role:         role.RoleID.String(),
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}

	acUser := db.ExportedAuditCtxWithUser(d, user.UserID)

	route, err := d.CreateRoute(ctx, acUser, db.CreateRouteParams{
		Slug:         types.Slug("tree-test"),
		Title:        "Tree Test Route",
		Status:       1,
		AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateRoute: %v", err)
	}

	// _root datatype: identifies the tree root
	dtRoot, err := d.CreateDatatype(ctx, acUser, db.CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "Root",
		Type:         "_root",
		AuthorID:     user.UserID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateDatatype(_root): %v", err)
	}

	// page datatype: regular content nodes
	dtPage, err := d.CreateDatatype(ctx, acUser, db.CreateDatatypeParams{
		DatatypeID:   types.NewDatatypeID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "Page",
		Type:         "page",
		AuthorID:     user.UserID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateDatatype(page): %v", err)
	}

	field, err := d.CreateField(ctx, acUser, db.CreateFieldParams{
		FieldID:      types.NewFieldID(),
		ParentID:     types.NullableDatatypeID{},
		Label:        "title",
		Data:         "",
		Validation:   types.EmptyJSON,
		UIConfig:     types.EmptyJSON,
		Type:         types.FieldTypeText,
		AuthorID:     types.NullableUserID{ID: user.UserID, Valid: true},
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seed CreateField: %v", err)
	}

	return contentTreeSeed{
		DB:       d,
		AC:       acUser,
		User:     user,
		RouteID:  types.NullableRouteID{ID: route.RouteID, Valid: true},
		DtRootID: types.NullableDatatypeID{ID: dtRoot.DatatypeID, Valid: true},
		DtPageID: types.NullableDatatypeID{ID: dtPage.DatatypeID, Valid: true},
		FieldID:  types.NullableFieldID{ID: field.FieldID, Valid: true},
	}
}

// createContentNode is a helper that creates a content_data row with the given pointers.
func createContentNode(t *testing.T, s contentTreeSeed, datatypeID types.NullableDatatypeID, parentID, firstChildID, nextSiblingID, prevSiblingID types.NullableContentID) *db.ContentData {
	t.Helper()
	now := types.TimestampNow()

	created, err := s.DB.CreateContentData(s.DB.Context, s.AC, db.CreateContentDataParams{
		RouteID:       s.RouteID,
		ParentID:      parentID,
		FirstChildID:  firstChildID,
		NextSiblingID: nextSiblingID,
		PrevSiblingID: prevSiblingID,
		DatatypeID:    datatypeID,
		AuthorID:      s.User.UserID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("createContentNode: %v", err)
	}
	return created
}

// updateContentNode updates a content_data row's pointers.
func updateContentNode(t *testing.T, s contentTreeSeed, id types.ContentID, datatypeID types.NullableDatatypeID, parentID, firstChildID, nextSiblingID, prevSiblingID types.NullableContentID) {
	t.Helper()
	now := types.TimestampNow()

	_, err := s.DB.UpdateContentData(s.DB.Context, s.AC, db.UpdateContentDataParams{
		ContentDataID: id,
		RouteID:       s.RouteID,
		ParentID:      parentID,
		FirstChildID:  firstChildID,
		NextSiblingID: nextSiblingID,
		PrevSiblingID: prevSiblingID,
		DatatypeID:    datatypeID,
		AuthorID:      s.User.UserID,
		Status:        types.ContentStatusPublished,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("updateContentNode(%v): %v", id, err)
	}
}

// nid wraps a ContentID in a valid NullableContentID.
func nid(id types.ContentID) types.NullableContentID {
	return types.NullableContentID{ID: id, Valid: true}
}

// noID returns an invalid NullableContentID (no reference).
func noID() types.NullableContentID {
	return types.NullableContentID{Valid: false}
}

// collectChildIDs walks the FirstChild/NextSibling chain and returns ContentDataIDs.
func collectChildIDs(parent *core.Node) []types.ContentID {
	var ids []types.ContentID
	child := parent.FirstChild
	for child != nil {
		ids = append(ids, child.ContentData.ContentDataID)
		child = child.NextSibling
	}
	return ids
}

// ---------------------------------------------------------------------------
// Test: Single root node round-trip
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_SingleRoot(t *testing.T) {
	// Simplest possible tree: one _root node. Verifies that a single
	// content_data row round-trips through the database and builds into
	// a valid tree with one node.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())

	// Read back via GetContentTreeByRoute
	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}
	if rows == nil || len(*rows) != 1 {
		t.Fatalf("expected 1 row, got %d", lenPtrSlice(rows))
	}

	// Build tree
	tree, stats, err := core.BuildFromRows(*rows)
	if err != nil {
		t.Fatalf("BuildFromRows: %v", err)
	}
	if tree.Node == nil {
		t.Fatal("tree root is nil")
	}
	if stats.NodesCount != 1 {
		t.Errorf("NodesCount = %d, want 1", stats.NodesCount)
	}
	if tree.Node.ContentData.ContentDataID != root.ContentDataID {
		t.Errorf("root ID = %v, want %v", tree.Node.ContentData.ContentDataID, root.ContentDataID)
	}
	if tree.Node.FirstChild != nil {
		t.Error("root should have no children")
	}
}

// ---------------------------------------------------------------------------
// Test: Root with children (parent pointers only)
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_RootWithChildren(t *testing.T) {
	// A root node with two children. Children reference the root via ParentID.
	// No sibling pointers set -- children are appended in insertion order.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	child1 := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	child2 := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())

	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}
	if rows == nil || len(*rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", lenPtrSlice(rows))
	}

	tree, _, err := core.BuildFromRows(*rows)
	if err != nil {
		t.Fatalf("BuildFromRows: %v", err)
	}

	children := collectChildIDs(tree.Node)
	if len(children) != 2 {
		t.Fatalf("root has %d children, want 2", len(children))
	}

	// Both child IDs should be present (order depends on map iteration in BuildFromRows)
	childSet := make(map[types.ContentID]bool)
	for _, id := range children {
		childSet[id] = true
	}
	if !childSet[child1.ContentDataID] {
		t.Errorf("child1 (%v) not found in tree children", child1.ContentDataID)
	}
	if !childSet[child2.ContentDataID] {
		t.Errorf("child2 (%v) not found in tree children", child2.ContentDataID)
	}
}

// ---------------------------------------------------------------------------
// Test: Sibling pointer ordering
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_SiblingOrdering(t *testing.T) {
	// Creates root with 3 children and explicit sibling pointers to enforce
	// a specific order: C -> A -> B. Verifies BuildFromRows respects the
	// stored sibling chain.
	t.Parallel()
	s := seedContentTreeBase(t)

	// Step 1: Create all nodes without sibling pointers
	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	nodeA := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	nodeB := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	nodeC := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())

	// Step 2: Update with sibling pointers to enforce order C -> A -> B
	// Root's first_child_id = C
	updateContentNode(t, s, root.ContentDataID, s.DtRootID,
		noID(),                   // parent
		nid(nodeC.ContentDataID), // first_child
		noID(),                   // next_sibling
		noID(),                   // prev_sibling
	)
	// C: prev=nil, next=A
	updateContentNode(t, s, nodeC.ContentDataID, s.DtPageID,
		nid(root.ContentDataID),  // parent
		noID(),                   // first_child
		nid(nodeA.ContentDataID), // next_sibling
		noID(),                   // prev_sibling
	)
	// A: prev=C, next=B
	updateContentNode(t, s, nodeA.ContentDataID, s.DtPageID,
		nid(root.ContentDataID),  // parent
		noID(),                   // first_child
		nid(nodeB.ContentDataID), // next_sibling
		nid(nodeC.ContentDataID), // prev_sibling
	)
	// B: prev=A, next=nil
	updateContentNode(t, s, nodeB.ContentDataID, s.DtPageID,
		nid(root.ContentDataID),  // parent
		noID(),                   // first_child
		noID(),                   // next_sibling
		nid(nodeA.ContentDataID), // prev_sibling
	)

	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}

	tree, _, err := core.BuildFromRows(*rows)
	if err != nil {
		t.Fatalf("BuildFromRows: %v", err)
	}

	children := collectChildIDs(tree.Node)
	if len(children) != 3 {
		t.Fatalf("root has %d children, want 3", len(children))
	}

	// Verify order: C, A, B
	wantOrder := []types.ContentID{nodeC.ContentDataID, nodeA.ContentDataID, nodeB.ContentDataID}
	for i, want := range wantOrder {
		if children[i] != want {
			t.Errorf("child[%d] = %v, want %v", i, children[i], want)
		}
	}

	// Verify prev_sibling pointers
	cNode := tree.Node.FirstChild
	if cNode.PrevSibling != nil {
		t.Error("C should have no prev sibling")
	}
	aNode := cNode.NextSibling
	if aNode.PrevSibling != cNode {
		t.Error("A's prev sibling should be C")
	}
	bNode := aNode.NextSibling
	if bNode.PrevSibling != aNode {
		t.Error("B's prev sibling should be A")
	}
	if bNode.NextSibling != nil {
		t.Error("B should have no next sibling")
	}
}

// ---------------------------------------------------------------------------
// Test: Nested tree (grandchildren)
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_NestedDepth(t *testing.T) {
	// Root -> Child -> Grandchild. Verifies multi-level hierarchy round-trips.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	child := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	grandchild := createContentNode(t, s, s.DtPageID, nid(child.ContentDataID), noID(), noID(), noID())

	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}
	if rows == nil || len(*rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", lenPtrSlice(rows))
	}

	tree, _, err := core.BuildFromRows(*rows)
	if err != nil {
		t.Fatalf("BuildFromRows: %v", err)
	}

	// Root -> Child
	rootChildren := collectChildIDs(tree.Node)
	if len(rootChildren) != 1 {
		t.Fatalf("root has %d children, want 1", len(rootChildren))
	}
	if rootChildren[0] != child.ContentDataID {
		t.Errorf("root child = %v, want %v", rootChildren[0], child.ContentDataID)
	}

	// Child -> Grandchild
	childNode := tree.Node.FirstChild
	grandChildren := collectChildIDs(childNode)
	if len(grandChildren) != 1 {
		t.Fatalf("child has %d children, want 1", len(grandChildren))
	}
	if grandChildren[0] != grandchild.ContentDataID {
		t.Errorf("grandchild = %v, want %v", grandChildren[0], grandchild.ContentDataID)
	}

	// Grandchild has no children
	gcNode := childNode.FirstChild
	if gcNode.FirstChild != nil {
		t.Error("grandchild should have no children")
	}

	// Parent pointers
	if gcNode.Parent != childNode {
		t.Error("grandchild's parent should be child")
	}
	if childNode.Parent != tree.Node {
		t.Error("child's parent should be root")
	}
}

// ---------------------------------------------------------------------------
// Test: Tree with content fields attached
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_WithContentFields(t *testing.T) {
	// Creates content_data with associated content_fields. Uses BuildTree
	// (not BuildFromRows) since field attachment requires the parallel-slice
	// interface.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	child := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())

	// Attach a content_field to the child
	now := types.TimestampNow()
	cf, err := s.DB.CreateContentField(s.DB.Context, s.AC, db.CreateContentFieldParams{
		RouteID:       s.RouteID,
		ContentDataID: nid(child.ContentDataID),
		FieldID:       s.FieldID,
		FieldValue:    "Hello World",
		AuthorID:      s.User.UserID,
		DateCreated:   now,
		DateModified:  now,
	})
	if err != nil {
		t.Fatalf("CreateContentField: %v", err)
	}

	// Read back content data
	cdList, err := s.DB.ListContentDataByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("ListContentDataByRoute: %v", err)
	}
	if cdList == nil || len(*cdList) != 2 {
		t.Fatalf("expected 2 content data, got %d", lenPtrSliceCD(cdList))
	}

	// Get the datatype for each content_data
	cd := *cdList
	dt := make([]db.Datatypes, len(cd))
	for i, c := range cd {
		if !c.DatatypeID.Valid {
			t.Fatalf("content_data[%d] has no datatype_id", i)
		}
		dtRow, dtErr := s.DB.GetDatatype(c.DatatypeID.ID)
		if dtErr != nil {
			t.Fatalf("GetDatatype(%v): %v", c.DatatypeID.ID, dtErr)
		}
		dt[i] = *dtRow
	}

	// Get content_fields and their field definitions
	var allCF []db.ContentFields
	var allDF []db.Fields
	for _, c := range cd {
		cfList, cfErr := s.DB.ListContentFieldsByContentData(nid(c.ContentDataID))
		if cfErr != nil {
			t.Fatalf("ListContentFieldsByContentData(%v): %v", c.ContentDataID, cfErr)
		}
		if cfList == nil {
			continue
		}
		for _, cfRow := range *cfList {
			if !cfRow.FieldID.Valid {
				continue
			}
			fDef, fErr := s.DB.GetField(cfRow.FieldID.ID)
			if fErr != nil {
				t.Fatalf("GetField(%v): %v", cfRow.FieldID.ID, fErr)
			}
			allCF = append(allCF, cfRow)
			allDF = append(allDF, *fDef)
		}
	}

	// Build tree with fields
	tree, _, err := core.BuildTree(cd, dt, allCF, allDF)
	if err != nil {
		t.Fatalf("BuildTree: %v", err)
	}

	// Find the child node in the tree
	childNode := core.FindByContentID(tree, child.ContentDataID)
	if childNode == nil {
		t.Fatal("child node not found in tree")
	}

	// Verify the field is attached
	if len(childNode.ContentFields) != 1 {
		t.Fatalf("child has %d content fields, want 1", len(childNode.ContentFields))
	}
	if childNode.ContentFields[0].FieldValue != "Hello World" {
		t.Errorf("field value = %q, want %q", childNode.ContentFields[0].FieldValue, "Hello World")
	}
	if childNode.ContentFields[0].ContentFieldID != cf.ContentFieldID {
		t.Errorf("content field ID mismatch")
	}

	// Verify the field definition is attached
	if len(childNode.Fields) != 1 {
		t.Fatalf("child has %d field defs, want 1", len(childNode.Fields))
	}
	if childNode.Fields[0].Label != "title" {
		t.Errorf("field label = %q, want %q", childNode.Fields[0].Label, "title")
	}

	// Suppress unused variable warnings
	_ = root
}

// ---------------------------------------------------------------------------
// Test: Delete a node, verify children are re-parentable
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_DeleteMiddleNode(t *testing.T) {
	// Creates root -> middle -> leaf, deletes the middle node from the DB.
	//
	// The content_data schema uses ON DELETE SET NULL on parent_id, so when the
	// middle node is deleted, the leaf's parent_id is set to NULL. This makes
	// the leaf a parentless node that BuildFromRows treats as a root-level node.
	//
	// This tests the real-world cascading delete semantics of the content tree.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	middle := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	leaf := createContentNode(t, s, s.DtPageID, nid(middle.ContentDataID), noID(), noID(), noID())

	// Delete the middle node
	err := s.DB.DeleteContentData(s.DB.Context, s.AC, middle.ContentDataID)
	if err != nil {
		t.Fatalf("DeleteContentData(middle): %v", err)
	}

	// Verify ON DELETE SET NULL: leaf's parent_id should now be NULL
	leafRow, err := s.DB.GetContentData(leaf.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentData(leaf): %v", err)
	}
	if leafRow.ParentID.Valid {
		t.Errorf("leaf parent_id should be NULL after middle deletion, got %v", leafRow.ParentID.ID)
	}

	// Re-read the route tree
	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}

	// Should have 2 rows: root and leaf (middle is deleted, leaf re-parented to NULL)
	if rows == nil || len(*rows) != 2 {
		t.Fatalf("expected 2 rows after delete, got %d", lenPtrSlice(rows))
	}

	// BuildFromRows: both root and leaf have no parent, so both become root-level
	tree, stats, buildErr := core.BuildFromRows(*rows)
	if buildErr != nil {
		t.Fatalf("BuildFromRows: %v", buildErr)
	}
	if tree.Node == nil {
		t.Fatal("tree root should not be nil")
	}

	// No orphans -- the leaf's parent was SET NULL, making it a root-level node
	if len(stats.FinalOrphans) != 0 {
		t.Errorf("FinalOrphans = %d, want 0", len(stats.FinalOrphans))
	}

	// Two nodes total in the index
	if stats.NodesCount != 2 {
		t.Errorf("NodesCount = %d, want 2", stats.NodesCount)
	}

	// Both original root and the orphaned leaf should be in the NodeIndex
	if tree.NodeIndex[root.ContentDataID] == nil {
		t.Error("root not found in NodeIndex")
	}
	if tree.NodeIndex[leaf.ContentDataID] == nil {
		t.Error("leaf not found in NodeIndex")
	}
}

// ---------------------------------------------------------------------------
// Test: Update sibling order and re-read
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_ReorderSiblings(t *testing.T) {
	// Creates root with children A, B, C ordered A->B->C.
	// Reorders to B->C->A by updating sibling pointers.
	// Verifies the tree rebuilds in the new order.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	nodeA := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	nodeB := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	nodeC := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())

	// Set order: B -> C -> A
	updateContentNode(t, s, root.ContentDataID, s.DtRootID,
		noID(), nid(nodeB.ContentDataID), noID(), noID())
	updateContentNode(t, s, nodeB.ContentDataID, s.DtPageID,
		nid(root.ContentDataID), noID(), nid(nodeC.ContentDataID), noID())
	updateContentNode(t, s, nodeC.ContentDataID, s.DtPageID,
		nid(root.ContentDataID), noID(), nid(nodeA.ContentDataID), nid(nodeB.ContentDataID))
	updateContentNode(t, s, nodeA.ContentDataID, s.DtPageID,
		nid(root.ContentDataID), noID(), noID(), nid(nodeC.ContentDataID))

	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}

	tree, _, err := core.BuildFromRows(*rows)
	if err != nil {
		t.Fatalf("BuildFromRows: %v", err)
	}

	children := collectChildIDs(tree.Node)
	if len(children) != 3 {
		t.Fatalf("root has %d children, want 3", len(children))
	}

	wantOrder := []types.ContentID{nodeB.ContentDataID, nodeC.ContentDataID, nodeA.ContentDataID}
	for i, want := range wantOrder {
		if children[i] != want {
			t.Errorf("after reorder child[%d] = %v, want %v", i, children[i], want)
		}
	}
}

// ---------------------------------------------------------------------------
// Test: GetContentDataDescendants
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_GetDescendants(t *testing.T) {
	// Verifies GetContentDataDescendants returns a node and all its descendants.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	child := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	grandchild := createContentNode(t, s, s.DtPageID, nid(child.ContentDataID), noID(), noID(), noID())

	// Descendants of root should include root + child + grandchild
	desc, err := s.DB.GetContentDataDescendants(s.DB.Context, root.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentDataDescendants(root): %v", err)
	}
	if desc == nil || len(*desc) != 3 {
		t.Fatalf("expected 3 descendants, got %d", lenPtrSliceCD(desc))
	}

	// Descendants of child should include child + grandchild
	desc, err = s.DB.GetContentDataDescendants(s.DB.Context, child.ContentDataID)
	if err != nil {
		t.Fatalf("GetContentDataDescendants(child): %v", err)
	}
	if desc == nil || len(*desc) != 2 {
		t.Fatalf("expected 2 descendants of child, got %d", lenPtrSliceCD(desc))
	}

	descIDs := make(map[types.ContentID]bool)
	for _, d := range *desc {
		descIDs[d.ContentDataID] = true
	}
	if !descIDs[child.ContentDataID] {
		t.Error("child should be in its own descendants")
	}
	if !descIDs[grandchild.ContentDataID] {
		t.Error("grandchild should be in child's descendants")
	}
}

// ---------------------------------------------------------------------------
// Test: Audit trail for tree mutations
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_AuditTrailOnMutations(t *testing.T) {
	// Every content_data mutation (create, update sibling pointers, delete)
	// should produce change_events. This bridges the audited and tree tests.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	child := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())

	// Update child's sibling pointer
	updateContentNode(t, s, child.ContentDataID, s.DtPageID,
		nid(root.ContentDataID), noID(), noID(), noID())

	// Delete child
	err := s.DB.DeleteContentData(s.DB.Context, s.AC, child.ContentDataID)
	if err != nil {
		t.Fatalf("DeleteContentData: %v", err)
	}

	// Check audit trail for the child: should have create + update + delete = 3
	events, err := s.DB.GetChangeEventsByRecord("content_data", string(child.ContentDataID))
	if err != nil {
		t.Fatalf("GetChangeEventsByRecord: %v", err)
	}
	if events == nil || len(*events) != 3 {
		t.Fatalf("expected 3 change events for child, got %d", lenOrZeroEvents(events))
	}

	// Suppress unused variable warnings
	_ = root
}

// ---------------------------------------------------------------------------
// Test: NodeIndex contains all nodes
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_NodeIndexComplete(t *testing.T) {
	// After building a tree from DB rows, every content_data ID should be
	// present in the NodeIndex for O(1) lookup.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	child1 := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	child2 := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	grandchild := createContentNode(t, s, s.DtPageID, nid(child1.ContentDataID), noID(), noID(), noID())

	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}

	tree, _, err := core.BuildFromRows(*rows)
	if err != nil {
		t.Fatalf("BuildFromRows: %v", err)
	}

	allIDs := []types.ContentID{
		root.ContentDataID,
		child1.ContentDataID,
		child2.ContentDataID,
		grandchild.ContentDataID,
	}

	for _, id := range allIDs {
		if tree.NodeIndex[id] == nil {
			t.Errorf("NodeIndex missing %v", id)
		}
	}

	if len(tree.NodeIndex) != 4 {
		t.Errorf("NodeIndex has %d entries, want 4", len(tree.NodeIndex))
	}
}

// ---------------------------------------------------------------------------
// Test: Traverse functions on DB-built tree
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_TraverseFunctions(t *testing.T) {
	// Builds a tree from the DB and verifies core.CountVisible,
	// core.FlattenVisible, and core.FindVisibleIndex work correctly.
	t.Parallel()
	s := seedContentTreeBase(t)

	root := createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	child1 := createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	createContentNode(t, s, s.DtPageID, nid(root.ContentDataID), noID(), noID(), noID())
	createContentNode(t, s, s.DtPageID, nid(child1.ContentDataID), noID(), noID(), noID())

	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}

	tree, _, err := core.BuildFromRows(*rows)
	if err != nil {
		t.Fatalf("BuildFromRows: %v", err)
	}

	allExpanded := func(*core.Node) bool { return true }

	// CountVisible: all 4 nodes
	count := core.CountVisible(tree.Node, allExpanded)
	if count != 4 {
		t.Errorf("CountVisible = %d, want 4", count)
	}

	// FlattenVisible: 4 nodes
	flat := core.FlattenVisible(tree.Node, allExpanded)
	if len(flat) != 4 {
		t.Errorf("FlattenVisible = %d nodes, want 4", len(flat))
	}

	// FindVisibleIndex for root should be 0
	idx := core.FindVisibleIndex(tree.Node, root.ContentDataID, allExpanded)
	if idx != 0 {
		t.Errorf("FindVisibleIndex(root) = %d, want 0", idx)
	}

	// FindByContentID
	found := core.FindByContentID(tree, child1.ContentDataID)
	if found == nil {
		t.Fatal("FindByContentID(child1) returned nil")
	}
	if found.ContentData.ContentDataID != child1.ContentDataID {
		t.Errorf("FindByContentID returned %v, want %v", found.ContentData.ContentDataID, child1.ContentDataID)
	}
}

// ---------------------------------------------------------------------------
// Test: Datatype metadata preserved in tree
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_DatatypeMetadata(t *testing.T) {
	// Verifies that the datatype label and type are preserved in tree nodes
	// after the round-trip through the database.
	t.Parallel()
	s := seedContentTreeBase(t)

	createContentNode(t, s, s.DtRootID, noID(), noID(), noID(), noID())
	createContentNode(t, s, s.DtPageID, noID(), noID(), noID(), noID())

	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}

	tree, _, _ := core.BuildFromRows(*rows)
	if tree.Node == nil {
		t.Fatal("tree root is nil")
	}

	// Find a node with _root type
	var foundRoot bool
	for _, node := range tree.NodeIndex {
		if node.Datatype.Type == "_root" {
			foundRoot = true
			if node.Datatype.Label != "Root" {
				t.Errorf("_root label = %q, want %q", node.Datatype.Label, "Root")
			}
		}
	}
	if !foundRoot {
		t.Error("no node with _root datatype type found")
	}
}

// ---------------------------------------------------------------------------
// Test: Empty route returns empty tree
// ---------------------------------------------------------------------------

func TestIntegration_ContentTree_EmptyRoute(t *testing.T) {
	// A route with no content_data should return no rows and BuildFromRows
	// should return an error.
	t.Parallel()
	s := seedContentTreeBase(t)

	rows, err := s.DB.GetContentTreeByRoute(s.RouteID)
	if err != nil {
		t.Fatalf("GetContentTreeByRoute: %v", err)
	}

	// No content_data for this route yet
	if rows != nil && len(*rows) != 0 {
		t.Fatalf("expected 0 rows for empty route, got %d", len(*rows))
	}

	var emptyRows []db.GetContentTreeByRouteRow
	if rows != nil {
		emptyRows = *rows
	}
	_, _, buildErr := core.BuildFromRows(emptyRows)
	if buildErr == nil {
		t.Error("expected error from BuildFromRows with empty rows")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func lenPtrSlice(s *[]db.GetContentTreeByRouteRow) int {
	if s == nil {
		return 0
	}
	return len(*s)
}

func lenPtrSliceCD(s *[]db.ContentData) int {
	if s == nil {
		return 0
	}
	return len(*s)
}

func lenOrZeroEvents(events *[]db.ChangeEvent) int {
	if events == nil {
		return 0
	}
	return len(*events)
}
