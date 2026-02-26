package core

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// mockFetcher implements TreeFetcher for testing.
type mockFetcher struct {
	trees      map[types.ContentID]*Root
	fetchCount atomic.Int64
}

func newMockFetcher() *mockFetcher {
	return &mockFetcher{
		trees: make(map[types.ContentID]*Root),
	}
}

func (m *mockFetcher) addTree(id types.ContentID, root *Root) {
	m.trees[id] = root
}

func (m *mockFetcher) FetchAndBuildTree(_ context.Context, id types.ContentID) (*Root, error) {
	m.fetchCount.Add(1)
	root, ok := m.trees[id]
	if !ok {
		return nil, fmt.Errorf("content %s not found", id)
	}
	return root, nil
}

// helper to build a minimal tree node.
func makeNode(id types.ContentID, dtType string, label string) *Node {
	dtID := types.NewDatatypeID()
	now := types.TimestampNow()
	return &Node{
		ContentData: &db.ContentData{
			ContentDataID: id,
			DatatypeID:    types.NullableDatatypeID{ID: dtID, Valid: true},
			Status:        types.ContentStatusPublished,
			DateCreated:   now,
			DateModified:  now,
		},
		Datatype: db.Datatypes{
			DatatypeID:   dtID,
			Label:        label,
			Type:         dtType,
			DateCreated:  now,
			DateModified: now,
		},
	}
}

// helper to add a content_tree_ref field to a node.
func addRefField(node *Node, refValue string) {
	fID := types.NewFieldID()
	now := types.TimestampNow()
	node.Fields = append(node.Fields, db.Fields{
		FieldID:      fID,
		Label:        "ref",
		Type:         types.FieldTypeContentTreeRef,
		DateCreated:  now,
		DateModified: now,
	})
	node.ContentFields = append(node.ContentFields, db.ContentFields{
		ContentFieldID: types.NewContentFieldID(),
		ContentDataID:  types.NullableContentID{ID: node.ContentData.ContentDataID, Valid: true},
		FieldID:        types.NullableFieldID{ID: fID, Valid: true},
		FieldValue:     refValue,
		DateCreated:    now,
		DateModified:   now,
	})
}

// helper to add a non-ref field to a node (should be skipped by composition).
func addTextField(node *Node, label, value string) {
	fID := types.NewFieldID()
	now := types.TimestampNow()
	node.Fields = append(node.Fields, db.Fields{
		FieldID:      fID,
		Label:        label,
		Type:         types.FieldTypeText,
		DateCreated:  now,
		DateModified: now,
	})
	node.ContentFields = append(node.ContentFields, db.ContentFields{
		ContentFieldID: types.NewContentFieldID(),
		ContentDataID:  types.NullableContentID{ID: node.ContentData.ContentDataID, Valid: true},
		FieldID:        types.NullableFieldID{ID: fID, Valid: true},
		FieldValue:     value,
		DateCreated:    now,
		DateModified:   now,
	})
}

// attachChild links child as the last child of parent using sibling pointers.
func attachChild(parent, child *Node) {
	child.Parent = parent
	if parent.FirstChild == nil {
		parent.FirstChild = child
		return
	}
	cur := parent.FirstChild
	for cur.NextSibling != nil {
		cur = cur.NextSibling
	}
	cur.NextSibling = child
	child.PrevSibling = cur
}

func TestComposeTrees_SingleReference(t *testing.T) {
	rootID := types.NewContentID()
	refNodeID := types.NewContentID()
	targetID := types.NewContentID()

	// Build the main tree: root -> _reference node with one ref field
	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")
	refNode := makeNode(refNodeID, string(types.DatatypeTypeReference), "nav_ref")
	addRefField(refNode, string(targetID))
	attachChild(rootNode, refNode)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID:    rootNode,
		refNodeID: refNode,
	}}

	// Build the target subtree
	targetNode := makeNode(targetID, string(types.DatatypeTypeNestedRoot), "navigation")
	targetRoot := &Root{Node: targetNode, NodeIndex: map[types.ContentID]*Node{
		targetID: targetNode,
	}}

	fetcher := newMockFetcher()
	fetcher.addTree(targetID, targetRoot)

	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 10, MaxConcurrency: 5})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	// The _reference node should now have the target as its child
	if refNode.FirstChild == nil {
		t.Fatal("reference node has no children after composition")
	}
	if refNode.FirstChild.ContentData.ContentDataID != targetID {
		t.Errorf("expected child ID %s, got %s", targetID, refNode.FirstChild.ContentData.ContentDataID)
	}
}

func TestComposeTrees_MultipleRefsOnOneNode(t *testing.T) {
	rootID := types.NewContentID()
	refNodeID := types.NewContentID()
	target1ID := types.NewContentID()
	target2ID := types.NewContentID()

	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")
	refNode := makeNode(refNodeID, string(types.DatatypeTypeReference), "multi_ref")
	addRefField(refNode, string(target1ID))
	addRefField(refNode, string(target2ID))
	attachChild(rootNode, refNode)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID:    rootNode,
		refNodeID: refNode,
	}}

	target1 := makeNode(target1ID, string(types.DatatypeTypeNestedRoot), "nav")
	target2 := makeNode(target2ID, string(types.DatatypeTypeNestedRoot), "footer")

	fetcher := newMockFetcher()
	fetcher.addTree(target1ID, &Root{Node: target1, NodeIndex: map[types.ContentID]*Node{target1ID: target1}})
	fetcher.addTree(target2ID, &Root{Node: target2, NodeIndex: map[types.ContentID]*Node{target2ID: target2}})

	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 10, MaxConcurrency: 5})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	// Verify two children in order
	first := refNode.FirstChild
	if first == nil {
		t.Fatal("reference node has no children")
	}
	if first.ContentData.ContentDataID != target1ID {
		t.Errorf("first child: expected %s, got %s", target1ID, first.ContentData.ContentDataID)
	}
	second := first.NextSibling
	if second == nil {
		t.Fatal("reference node has only one child, expected two")
	}
	if second.ContentData.ContentDataID != target2ID {
		t.Errorf("second child: expected %s, got %s", target2ID, second.ContentData.ContentDataID)
	}
}

func TestComposeTrees_NestedReferences(t *testing.T) {
	rootID := types.NewContentID()
	refNodeID := types.NewContentID()
	outerTargetID := types.NewContentID()
	innerRefID := types.NewContentID()
	innerTargetID := types.NewContentID()

	// Main tree: root -> _reference(outer)
	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")
	refNode := makeNode(refNodeID, string(types.DatatypeTypeReference), "outer_ref")
	addRefField(refNode, string(outerTargetID))
	attachChild(rootNode, refNode)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID:    rootNode,
		refNodeID: refNode,
	}}

	// Outer target tree: _nested_root -> _reference(inner)
	outerNode := makeNode(outerTargetID, string(types.DatatypeTypeNestedRoot), "outer")
	innerRefNode := makeNode(innerRefID, string(types.DatatypeTypeReference), "inner_ref")
	addRefField(innerRefNode, string(innerTargetID))
	attachChild(outerNode, innerRefNode)

	outerRoot := &Root{Node: outerNode, NodeIndex: map[types.ContentID]*Node{
		outerTargetID: outerNode,
		innerRefID:    innerRefNode,
	}}

	// Inner target tree
	innerNode := makeNode(innerTargetID, string(types.DatatypeTypeNestedRoot), "inner")
	innerRoot := &Root{Node: innerNode, NodeIndex: map[types.ContentID]*Node{
		innerTargetID: innerNode,
	}}

	fetcher := newMockFetcher()
	fetcher.addTree(outerTargetID, outerRoot)
	fetcher.addTree(innerTargetID, innerRoot)

	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 10, MaxConcurrency: 5})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	// Verify: root -> refNode -> outerNode -> innerRefNode -> innerNode
	if refNode.FirstChild == nil {
		t.Fatal("outer reference not resolved")
	}
	if refNode.FirstChild.ContentData.ContentDataID != outerTargetID {
		t.Errorf("expected outer target %s", outerTargetID)
	}
	innerRef := refNode.FirstChild.FirstChild
	if innerRef == nil {
		t.Fatal("outer target has no children (inner ref)")
	}
	innerChild := innerRef.FirstChild
	if innerChild == nil {
		t.Fatal("inner reference not resolved")
	}
	if innerChild.ContentData.ContentDataID != innerTargetID {
		t.Errorf("expected inner target %s, got %s", innerTargetID, innerChild.ContentData.ContentDataID)
	}
}

func TestComposeTrees_CycleDetection(t *testing.T) {
	rootID := types.NewContentID()
	refNodeID := types.NewContentID()

	// The _reference points back to the root content ID — cycle
	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")
	refNode := makeNode(refNodeID, string(types.DatatypeTypeReference), "self_ref")
	addRefField(refNode, string(rootID)) // points at own root
	attachChild(rootNode, refNode)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID:    rootNode,
		refNodeID: refNode,
	}}

	fetcher := newMockFetcher()
	// Don't need to add any trees — the cycle should be caught before fetch

	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 10, MaxConcurrency: 5})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	// Should have a _system_log child instead of a fetched tree
	child := refNode.FirstChild
	if child == nil {
		t.Fatal("expected _system_log child for circular reference")
	}
	if child.Datatype.Type != string(types.DatatypeTypeSystemLog) {
		t.Errorf("expected _system_log type, got %s", child.Datatype.Type)
	}
	// Check error field
	if len(child.ContentFields) < 1 {
		t.Fatal("system log node has no fields")
	}
	if child.ContentFields[0].FieldValue != "circular_reference" {
		t.Errorf("expected circular_reference error, got %s", child.ContentFields[0].FieldValue)
	}
}

func TestComposeTrees_DepthLimit(t *testing.T) {
	rootID := types.NewContentID()
	refNodeID := types.NewContentID()
	targetID := types.NewContentID()

	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")
	refNode := makeNode(refNodeID, string(types.DatatypeTypeReference), "deep_ref")
	addRefField(refNode, string(targetID))
	attachChild(rootNode, refNode)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID:    rootNode,
		refNodeID: refNode,
	}}

	targetNode := makeNode(targetID, string(types.DatatypeTypeNestedRoot), "target")
	fetcher := newMockFetcher()
	fetcher.addTree(targetID, &Root{Node: targetNode, NodeIndex: map[types.ContentID]*Node{targetID: targetNode}})

	// MaxDepth=1 means depth 0 processes, but depth+1 >= MaxDepth triggers limit
	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 1, MaxConcurrency: 5})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	child := refNode.FirstChild
	if child == nil {
		t.Fatal("expected _system_log child for max_depth")
	}
	if child.Datatype.Type != string(types.DatatypeTypeSystemLog) {
		t.Errorf("expected _system_log type, got %s", child.Datatype.Type)
	}
	if child.ContentFields[0].FieldValue != "max_depth" {
		t.Errorf("expected max_depth error, got %s", child.ContentFields[0].FieldValue)
	}
}

func TestComposeTrees_BrokenReference(t *testing.T) {
	rootID := types.NewContentID()
	refNodeID := types.NewContentID()
	missingID := types.NewContentID()

	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")
	refNode := makeNode(refNodeID, string(types.DatatypeTypeReference), "broken_ref")
	addRefField(refNode, string(missingID))
	attachChild(rootNode, refNode)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID:    rootNode,
		refNodeID: refNode,
	}}

	fetcher := newMockFetcher()
	// Don't add the target — fetch will return error

	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 10, MaxConcurrency: 5})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	child := refNode.FirstChild
	if child == nil {
		t.Fatal("expected _system_log child for broken reference")
	}
	if child.Datatype.Type != string(types.DatatypeTypeSystemLog) {
		t.Errorf("expected _system_log type, got %s", child.Datatype.Type)
	}
	if child.ContentFields[0].FieldValue != "not_found" {
		t.Errorf("expected not_found error, got %s", child.ContentFields[0].FieldValue)
	}
}

func TestComposeTrees_EmptyFieldValueSkipped(t *testing.T) {
	rootID := types.NewContentID()
	refNodeID := types.NewContentID()
	targetID := types.NewContentID()

	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")
	refNode := makeNode(refNodeID, string(types.DatatypeTypeReference), "partial_ref")
	addRefField(refNode, "")              // empty — should be skipped
	addRefField(refNode, string(targetID)) // valid

	attachChild(rootNode, refNode)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID:    rootNode,
		refNodeID: refNode,
	}}

	targetNode := makeNode(targetID, string(types.DatatypeTypeNestedRoot), "nav")
	fetcher := newMockFetcher()
	fetcher.addTree(targetID, &Root{Node: targetNode, NodeIndex: map[types.ContentID]*Node{targetID: targetNode}})

	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 10, MaxConcurrency: 5})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	// Should have exactly one child (the valid ref), not two
	child := refNode.FirstChild
	if child == nil {
		t.Fatal("expected one child")
	}
	if child.ContentData.ContentDataID != targetID {
		t.Errorf("expected target %s, got %s", targetID, child.ContentData.ContentDataID)
	}
	if child.NextSibling != nil {
		t.Error("expected only one child, got a sibling")
	}
}

func TestComposeTrees_NonRefFieldsIgnored(t *testing.T) {
	rootID := types.NewContentID()
	refNodeID := types.NewContentID()
	targetID := types.NewContentID()

	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")
	refNode := makeNode(refNodeID, string(types.DatatypeTypeReference), "mixed_ref")
	addTextField(refNode, "title", "Navigation")       // text field — should be ignored
	addRefField(refNode, string(targetID))              // ref field — should be resolved
	addTextField(refNode, "description", "Site nav")    // text field — should be ignored

	attachChild(rootNode, refNode)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID:    rootNode,
		refNodeID: refNode,
	}}

	targetNode := makeNode(targetID, string(types.DatatypeTypeNestedRoot), "nav")
	fetcher := newMockFetcher()
	fetcher.addTree(targetID, &Root{Node: targetNode, NodeIndex: map[types.ContentID]*Node{targetID: targetNode}})

	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 10, MaxConcurrency: 5})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	child := refNode.FirstChild
	if child == nil {
		t.Fatal("expected one child")
	}
	if child.ContentData.ContentDataID != targetID {
		t.Errorf("expected target %s, got %s", targetID, child.ContentData.ContentDataID)
	}
	if child.NextSibling != nil {
		t.Error("expected only one child (text fields should be ignored)")
	}
}

func TestComposeTrees_ConcurrentResolution(t *testing.T) {
	rootID := types.NewContentID()
	ref1ID := types.NewContentID()
	ref2ID := types.NewContentID()
	ref3ID := types.NewContentID()
	target1ID := types.NewContentID()
	target2ID := types.NewContentID()
	target3ID := types.NewContentID()

	// Root with 3 sibling _reference nodes — should resolve concurrently
	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")

	ref1 := makeNode(ref1ID, string(types.DatatypeTypeReference), "nav_ref")
	addRefField(ref1, string(target1ID))

	ref2 := makeNode(ref2ID, string(types.DatatypeTypeReference), "body_ref")
	addRefField(ref2, string(target2ID))

	ref3 := makeNode(ref3ID, string(types.DatatypeTypeReference), "footer_ref")
	addRefField(ref3, string(target3ID))

	attachChild(rootNode, ref1)
	attachChild(rootNode, ref2)
	attachChild(rootNode, ref3)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID: rootNode,
		ref1ID: ref1,
		ref2ID: ref2,
		ref3ID: ref3,
	}}

	t1 := makeNode(target1ID, string(types.DatatypeTypeNestedRoot), "nav")
	t2 := makeNode(target2ID, string(types.DatatypeTypeNestedRoot), "body")
	t3 := makeNode(target3ID, string(types.DatatypeTypeNestedRoot), "footer")

	fetcher := newMockFetcher()
	fetcher.addTree(target1ID, &Root{Node: t1, NodeIndex: map[types.ContentID]*Node{target1ID: t1}})
	fetcher.addTree(target2ID, &Root{Node: t2, NodeIndex: map[types.ContentID]*Node{target2ID: t2}})
	fetcher.addTree(target3ID, &Root{Node: t3, NodeIndex: map[types.ContentID]*Node{target3ID: t3}})

	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 10, MaxConcurrency: 10})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	// Verify all 3 references resolved
	if ref1.FirstChild == nil || ref1.FirstChild.ContentData.ContentDataID != target1ID {
		t.Error("ref1 not resolved correctly")
	}
	if ref2.FirstChild == nil || ref2.FirstChild.ContentData.ContentDataID != target2ID {
		t.Error("ref2 not resolved correctly")
	}
	if ref3.FirstChild == nil || ref3.FirstChild.ContentData.ContentDataID != target3ID {
		t.Error("ref3 not resolved correctly")
	}

	// Verify fetcher was called 3 times
	if fetcher.fetchCount.Load() != 3 {
		t.Errorf("expected 3 fetches, got %d", fetcher.fetchCount.Load())
	}
}

func TestComposeTrees_NoReferences(t *testing.T) {
	rootID := types.NewContentID()
	childID := types.NewContentID()

	rootNode := makeNode(rootID, string(types.DatatypeTypeRoot), "page")
	childNode := makeNode(childID, "section", "hero")
	attachChild(rootNode, childNode)

	root := &Root{Node: rootNode, NodeIndex: map[types.ContentID]*Node{
		rootID:  rootNode,
		childID: childNode,
	}}

	fetcher := newMockFetcher()

	err := ComposeTrees(context.Background(), root, fetcher, ComposeOptions{MaxDepth: 10, MaxConcurrency: 5})
	if err != nil {
		t.Fatalf("ComposeTrees returned error: %v", err)
	}

	// No fetches should have happened
	if fetcher.fetchCount.Load() != 0 {
		t.Errorf("expected 0 fetches, got %d", fetcher.fetchCount.Load())
	}
}

func TestCachedFetcher_CacheHit(t *testing.T) {
	targetID := types.NewContentID()
	targetNode := makeNode(targetID, string(types.DatatypeTypeNestedRoot), "nav")
	targetRoot := &Root{Node: targetNode, NodeIndex: map[types.ContentID]*Node{targetID: targetNode}}

	inner := newMockFetcher()
	inner.addTree(targetID, targetRoot)

	cached := NewCachedFetcher(inner)

	// First fetch — cache miss
	r1, err := cached.FetchAndBuildTree(context.Background(), targetID)
	if err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	if r1.Node.ContentData.ContentDataID != targetID {
		t.Error("first fetch returned wrong tree")
	}

	// Second fetch — cache hit
	r2, err := cached.FetchAndBuildTree(context.Background(), targetID)
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	if r2 != r1 {
		t.Error("second fetch should return same pointer (cache hit)")
	}

	// Inner fetcher should have been called only once
	if inner.fetchCount.Load() != 1 {
		t.Errorf("expected 1 inner fetch, got %d", inner.fetchCount.Load())
	}
}
