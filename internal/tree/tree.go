// Package tree provides the TUI-layer tree representation for ModulaCMS.
// It wraps the shared core tree-building algorithms (internal/tree/core)
// and adds TUI-specific state: expand/collapse, indent, and wrapping.
//
// All existing field names and method signatures are preserved so that
// internal/tui/ requires zero changes.
package tree

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree/core"
)

// Root is the TUI-layer tree container. It wraps a core.Root and maintains
// a parallel tree of TUI-specific Node wrappers.
type Root struct {
	Root      *Node
	NodeIndex map[types.ContentID]*Node
	Orphans   map[types.ContentID]*Node
	MaxRetry  int
	rootNodes []*Node // all parentless nodes, used during loading for sibling reorder

	// Core holds the shared core tree for composition layer access.
	Core *core.Root
}

// LoadStats tracks tree construction diagnostics.
type LoadStats struct {
	NodesCount      int
	OrphansResolved int
	RetryAttempts   int
	CircularRefs    []types.ContentID
	FinalOrphans    []types.ContentID
}

func (stats LoadStats) String() string {
	return fmt.Sprintf("Nodes Count: %d\nOrphans Resolved %d\nRetry Attempts%d\nCircular Refs %v\nFinal Orphans %v\n",
		stats.NodesCount,
		stats.OrphansResolved,
		stats.RetryAttempts,
		stats.CircularRefs,
		stats.FinalOrphans,
	)
}

// Node is the TUI-layer tree node. It mirrors core.Node's data fields
// (Instance, Datatype, Fields, InstanceFields) and adds TUI-specific
// state (Expand, Indent, Wrapped). Pointer fields (Parent, FirstChild,
// NextSibling, PrevSibling) reference other tree.Node values, not core.Node.
type Node struct {
	Instance       *db.ContentData
	InstanceFields []db.ContentFields
	Datatype       db.Datatypes
	Fields         []db.Fields
	Parent         *Node
	FirstChild     *Node
	NextSibling    *Node
	PrevSibling    *Node
	Expand         bool
	Indent         int
	Wrapped        int

	// CoreNode holds the corresponding core.Node for composition layer access.
	CoreNode *core.Node
}

func NewRoot() *Root {
	return &Root{
		NodeIndex: make(map[types.ContentID]*Node),
		Orphans:   make(map[types.ContentID]*Node),
		MaxRetry:  100,
	}
}

func NewNode(row db.GetRouteTreeByRouteIDRow) *Node {
	cd := db.ContentData{
		ContentDataID: row.ContentDataID,
		ParentID:      row.ParentID,
	}

	return &Node{
		Instance: &cd,
		Expand:   true,
	}
}

func NewNodeFromContentTree(row db.GetContentTreeByRouteRow) *Node {
	cd := db.ContentData{
		ContentDataID: row.ContentDataID,
		ParentID:      row.ParentID,
		FirstChildID:  row.FirstChildID,
		NextSiblingID: row.NextSiblingID,
		PrevSiblingID: row.PrevSiblingID,
		RouteID:       row.RouteID,
		DatatypeID:    row.DatatypeID,
		AuthorID:      row.AuthorID,
		Status:        row.Status,
		DateCreated:   row.DateCreated,
		DateModified:  row.DateModified,
	}

	dt := db.Datatypes{
		DatatypeID: row.DatatypeID.ID,
		Label:      row.DatatypeLabel,
		Type:       row.DatatypeType,
	}

	return &Node{
		Instance: &cd,
		Datatype: dt,
		Expand:   true,
	}
}

// LoadFromRows builds the tree from query rows by delegating to core.BuildFromRows,
// then wrapping the core nodes with TUI-specific state.
func (page *Root) LoadFromRows(rows *[]db.GetContentTreeByRouteRow) (*LoadStats, error) {
	coreRoot, coreStats, coreErr := core.BuildFromRows(*rows)
	page.Core = coreRoot

	// Build a map from core.Node to tree.Node for pointer re-linking
	coreToTree := make(map[*core.Node]*Node, len(coreRoot.NodeIndex))

	// Phase 1: Create tree.Node wrappers for every core.Node
	for id, cn := range coreRoot.NodeIndex {
		tn := &Node{
			Instance:       cn.ContentData,
			InstanceFields: cn.ContentFields,
			Datatype:       cn.Datatype,
			Fields:         cn.Fields,
			Expand:         true,
			CoreNode:       cn,
		}
		coreToTree[cn] = tn
		page.NodeIndex[id] = tn
	}

	// Phase 2: Re-link pointers in tree.Node space
	for cn, tn := range coreToTree {
		if cn.Parent != nil {
			tn.Parent = coreToTree[cn.Parent]
		}
		if cn.FirstChild != nil {
			tn.FirstChild = coreToTree[cn.FirstChild]
		}
		if cn.NextSibling != nil {
			tn.NextSibling = coreToTree[cn.NextSibling]
		}
		if cn.PrevSibling != nil {
			tn.PrevSibling = coreToTree[cn.PrevSibling]
		}
	}

	// Set root
	if coreRoot.Node != nil {
		page.Root = coreToTree[coreRoot.Node]
	}

	// Collect root-level nodes
	for _, tn := range page.NodeIndex {
		if tn.Parent == nil {
			page.rootNodes = append(page.rootNodes, tn)
		}
	}

	// Convert core stats to tree stats
	stats := &LoadStats{
		NodesCount:      coreStats.NodesCount,
		OrphansResolved: coreStats.OrphansResolved,
		RetryAttempts:   coreStats.RetryAttempts,
		CircularRefs:    coreStats.CircularRefs,
		FinalOrphans:    coreStats.FinalOrphans,
	}

	return stats, coreErr
}

// LoadFromAdminData builds the tree from admin content slices by delegating to
// core.BuildAdminTree, then wrapping the core nodes with TUI-specific state.
// The core layer converts admin ID types to ContentID via string casting, so the
// resulting NodeIndex uses ContentID keys and all tree traversal methods work
// without modification.
func (page *Root) LoadFromAdminData(
	cd []db.AdminContentData,
	dt []db.AdminDatatypes,
	cf []db.AdminContentFields,
	df []db.AdminFields,
) (*LoadStats, error) {
	coreRoot, coreStats, coreErr := core.BuildAdminTree(cd, dt, cf, df)
	page.Core = coreRoot

	// Build a map from core.Node to tree.Node for pointer re-linking
	coreToTree := make(map[*core.Node]*Node, len(coreRoot.NodeIndex))

	// Phase 1: Create tree.Node wrappers for every core.Node
	for id, cn := range coreRoot.NodeIndex {
		tn := &Node{
			Instance:       cn.ContentData,
			InstanceFields: cn.ContentFields,
			Datatype:       cn.Datatype,
			Fields:         cn.Fields,
			Expand:         true,
			CoreNode:       cn,
		}
		coreToTree[cn] = tn
		page.NodeIndex[id] = tn
	}

	// Phase 2: Re-link pointers in tree.Node space
	for cn, tn := range coreToTree {
		if cn.Parent != nil {
			tn.Parent = coreToTree[cn.Parent]
		}
		if cn.FirstChild != nil {
			tn.FirstChild = coreToTree[cn.FirstChild]
		}
		if cn.NextSibling != nil {
			tn.NextSibling = coreToTree[cn.NextSibling]
		}
		if cn.PrevSibling != nil {
			tn.PrevSibling = coreToTree[cn.PrevSibling]
		}
	}

	// Set root
	if coreRoot.Node != nil {
		page.Root = coreToTree[coreRoot.Node]
	}

	// Collect root-level nodes
	for _, tn := range page.NodeIndex {
		if tn.Parent == nil {
			page.rootNodes = append(page.rootNodes, tn)
		}
	}

	// Convert core stats to tree stats
	stats := &LoadStats{
		NodesCount:      coreStats.NodesCount,
		OrphansResolved: coreStats.OrphansResolved,
		RetryAttempts:   coreStats.RetryAttempts,
		CircularRefs:    coreStats.CircularRefs,
		FinalOrphans:    coreStats.FinalOrphans,
	}

	return stats, coreErr
}

// InsertNodeByIndex adds a node to the tree with the given relationships.
func (page *Root) InsertNodeByIndex(parent, firstChild, prevSibling, nextSibling, n *Node) {
	page.NodeIndex[n.Instance.ContentDataID] = n
	n.Parent = parent
	n.FirstChild = firstChild
	n.PrevSibling = prevSibling
	n.NextSibling = nextSibling
}

// DeleteNodeByIndex removes a node from the tree, re-linking its children and siblings.
func (page *Root) DeleteNodeByIndex(n *Node) bool {
	target := page.NodeIndex[n.Instance.ContentDataID]
	if page.Root == nil || target == nil || page.Root == target || target.Parent == nil {
		return false
	}
	if target.Parent.FirstChild == target {
		DeleteFirstChild(target)
		delete(page.NodeIndex, n.Instance.ContentDataID)
		return true
	}
	current := target.Parent.FirstChild
	for current != nil && current != target {
		current = current.NextSibling
	}
	if current == nil {
		return false
	}
	DeleteNestedChild(target)

	delete(page.NodeIndex, n.Instance.ContentDataID)
	return true
}

// DeleteFirstChild handles deletion when the target is the first child of its parent.
func DeleteFirstChild(target *Node) bool {
	if target.FirstChild != nil {
		return deleteFirstChildHasChildren(target)
	}
	return deleteFirstChildNoChildren(target)
}

func deleteFirstChildHasChildren(target *Node) bool {
	if target.NextSibling != nil {
		target.Parent.FirstChild = target.FirstChild
		current := target.FirstChild
		for current.NextSibling != nil {
			current.Parent = target.Parent
			current = current.NextSibling
		}
		target.NextSibling.PrevSibling = current
		current.NextSibling = target.NextSibling
		current.Parent = target.Parent
		return true
	}
	target.Parent.FirstChild = target.FirstChild
	current := target.FirstChild
	for current != nil {
		current.Parent = target.Parent
		current = current.NextSibling
	}
	return true
}

func deleteFirstChildNoChildren(target *Node) bool {
	if target.NextSibling != nil {
		target.Parent.FirstChild = target.NextSibling
		target.NextSibling.PrevSibling = nil
		return true
	}
	target.Parent.FirstChild = nil
	return true
}

// DeleteNestedChild handles deletion when the target is not the first child.
func DeleteNestedChild(target *Node) bool {
	if target.FirstChild != nil {
		return deleteNestedChildHasChildren(target)
	}
	return deleteNestedChildNoChildren(target)
}

func deleteNestedChildHasChildren(target *Node) bool {
	if target.NextSibling != nil {
		if target.PrevSibling == nil {
			return false
		}
		target.PrevSibling.NextSibling = target.FirstChild
		current := target.FirstChild
		for current.NextSibling != nil {
			current.Parent = target.Parent
			current = current.NextSibling
		}
		target.NextSibling.PrevSibling = current
		current.NextSibling = target.NextSibling
		current.Parent = target.Parent
		return true
	}
	if target.PrevSibling == nil || target.FirstChild == nil {
		return false
	}
	target.PrevSibling.NextSibling = target.FirstChild
	target.FirstChild.PrevSibling = target.PrevSibling
	current := target.FirstChild
	for current != nil {
		current.Parent = target.Parent
		current = current.NextSibling
	}
	return true
}

func deleteNestedChildNoChildren(target *Node) bool {
	if target.NextSibling != nil {
		if target.PrevSibling == nil {
			return false
		}
		target.PrevSibling.NextSibling = target.NextSibling
		target.NextSibling.PrevSibling = target.PrevSibling
		return true
	}
	if target.PrevSibling == nil {
		return false
	}
	target.PrevSibling.NextSibling = nil
	return true
}

// CountVisible counts the number of visible nodes in the tree.
func (page *Root) CountVisible() int {
	if page.Root == nil {
		return 0
	}
	count := 0
	page.countNodesRecursive(page.Root, &count)
	return count
}

func (page *Root) countNodesRecursive(node *Node, count *int) {
	if node == nil {
		return
	}
	*count++

	if node.Expand && node.FirstChild != nil {
		page.countNodesRecursive(node.FirstChild, count)
	}

	if node.NextSibling != nil {
		page.countNodesRecursive(node.NextSibling, count)
	}
}

// NodeAtIndex returns the node at the given visible index.
func (page *Root) NodeAtIndex(index int) *Node {
	if page.Root == nil {
		return nil
	}
	currentIndex := 0
	return page.nodeAtIndex(page.Root, index, &currentIndex)
}

func (page *Root) nodeAtIndex(node *Node, targetIndex int, currentIndex *int) *Node {
	if node == nil {
		return nil
	}

	if *currentIndex == targetIndex {
		return node
	}
	*currentIndex++

	if node.Expand && node.FirstChild != nil {
		if result := page.nodeAtIndex(node.FirstChild, targetIndex, currentIndex); result != nil {
			return result
		}
	}

	if node.NextSibling != nil {
		return page.nodeAtIndex(node.NextSibling, targetIndex, currentIndex)
	}

	return nil
}

// IsDescendantOf returns true if node is a descendant of ancestor.
// Walks up the parent chain from node looking for ancestor.
func IsDescendantOf(node, ancestor *Node) bool {
	current := node.Parent
	for current != nil {
		if current == ancestor {
			return true
		}
		current = current.Parent
	}
	return false
}

// FlattenVisible returns a flat slice of all visible nodes in display order,
// respecting the Expand state of each node.
func (page *Root) FlattenVisible() []*Node {
	if page.Root == nil {
		return nil
	}
	var result []*Node
	page.flattenVisibleRecursive(page.Root, &result)
	return result
}

func (page *Root) flattenVisibleRecursive(node *Node, result *[]*Node) {
	if node == nil {
		return
	}
	*result = append(*result, node)

	if node.Expand && node.FirstChild != nil {
		page.flattenVisibleRecursive(node.FirstChild, result)
	}

	if node.NextSibling != nil {
		page.flattenVisibleRecursive(node.NextSibling, result)
	}
}

// FindVisibleIndex returns the visible index of the node with the given
// content ID, or -1 if the node is not currently visible.
func (page *Root) FindVisibleIndex(contentID types.ContentID) int {
	visible := page.FlattenVisible()
	for i, n := range visible {
		if n.Instance != nil && n.Instance.ContentDataID == contentID {
			return i
		}
	}
	return -1
}
