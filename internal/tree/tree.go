package tree

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

type Root struct {
	Root      *Node
	NodeIndex map[types.ContentID]*Node
	Orphans   map[types.ContentID]*Node
	MaxRetry  int
}

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
		Expand:   true, // Expanded by default
	}

}

func NewNodeFromContentTree(row db.GetContentTreeByRouteRow) *Node {
	cd := db.ContentData{
		ContentDataID: row.ContentDataID,
		ParentID:      row.ParentID,
		RouteID:       row.RouteID,
		DatatypeID:    row.DatatypeID,
		AuthorID:      row.AuthorID,
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
		Expand:   true, // Expanded by default
	}
}

func (page *Root) LoadFromRows(rows *[]db.GetContentTreeByRouteRow) (*LoadStats, error) {
	stats := &LoadStats{
		NodesCount:      0,
		OrphansResolved: 0,
		RetryAttempts:   0,
		CircularRefs:    make([]types.ContentID, 0),
		FinalOrphans:    make([]types.ContentID, 0),
	}

	// Phase 1: Create all nodes and populate indexes
	if err := page.createAllNodes(rows, stats); err != nil {
		return stats, err
	}

	// Phase 2: Assign hierarchy for nodes with existing parents
	if err := page.assignImmediateHierarchy(stats); err != nil {
		return stats, err
	}

	// Phase 3: Iteratively resolve orphaned nodes
	if err := page.resolveOrphans(stats); err != nil {
		return stats, err
	}

	return stats, page.validateFinalState(stats)
}

func (page *Root) createAllNodes(rows *[]db.GetContentTreeByRouteRow, stats *LoadStats) error {
	for _, row := range *rows {
		node := NewNodeFromContentTree(row)
		page.NodeIndex[node.Instance.ContentDataID] = node
		stats.NodesCount++

		// Set root immediately if no parent
		if !node.Instance.ParentID.Valid && page.Root == nil {
			page.Root = node
			node.Parent = nil
		}
	}
	return nil
}

// Phase 2: Immediate Hierarchy Assignment
func (page *Root) assignImmediateHierarchy(stats *LoadStats) error {
	for id, node := range page.NodeIndex {
		if !node.Instance.ParentID.Valid {
			continue // Skip root
		}

		parentID := node.Instance.ParentID.ID
		parent := page.NodeIndex[parentID]

		if parent != nil {
			// Parent exists - assign immediately
			page.attachNodeToParent(node, parent)
		} else {
			// Parent missing - defer resolution
			page.Orphans[id] = node
		}
	}
	return nil
}

// Phase 3: Iterative Orphan Resolution
func (page *Root) resolveOrphans(stats *LoadStats) error {
	for len(page.Orphans) > 0 && stats.RetryAttempts < page.MaxRetry {
		stats.RetryAttempts++
		orphansResolved := 0

		// Try to resolve each orphan
		for id, orphan := range page.Orphans {
			parentID := orphan.Instance.ParentID.ID
			parent := page.NodeIndex[parentID]

			if parent != nil && parent.Parent != nil { // Parent now exists and is connected
				page.attachNodeToParent(orphan, parent)
				delete(page.Orphans, id)
				orphansResolved++
				stats.OrphansResolved++
			}
		}

		// No progress made - check for circular references
		if orphansResolved == 0 {
			if page.detectCircularReferences(stats) {
				break // Found circular refs, stop trying
			}
			// Could be legitimate forward references, continue trying
		}
	}

	return nil
}

// Detect circular reference chains
func (page *Root) detectCircularReferences(stats *LoadStats) bool {
	circularRefs := []types.ContentID{}

	for id, orphan := range page.Orphans {
		if page.hasCircularReference(orphan, make(map[types.ContentID]bool)) {
			circularRefs = append(circularRefs, id)
		}
	}

	stats.CircularRefs = circularRefs
	return len(circularRefs) > 0
}

// Check if node creates circular reference
func (page *Root) hasCircularReference(node *Node, visited map[types.ContentID]bool) bool {
	nodeID := node.Instance.ContentDataID

	if visited[nodeID] {
		return true // Found cycle
	}

	if !node.Instance.ParentID.Valid {
		return false // Reached root
	}

	visited[nodeID] = true
	parentID := node.Instance.ParentID.ID
	parent := page.NodeIndex[parentID]

	if parent == nil {
		return false // Parent doesn't exist (not circular, just missing)
	}

	return page.hasCircularReference(parent, visited)
}

// Attach node to parent with proper sibling linking
func (page *Root) attachNodeToParent(node, parent *Node) {
	node.Parent = parent

	if parent.FirstChild == nil {
		parent.FirstChild = node
	} else {
		// Add as last sibling
		current := parent.FirstChild
		for current.NextSibling != nil {
			current = current.NextSibling
		}
		current.NextSibling = node
		node.PrevSibling = current
	}
}

// Final validation and error reporting
func (page *Root) validateFinalState(stats *LoadStats) error {
	// Record final orphans
	for id := range page.Orphans {
		stats.FinalOrphans = append(stats.FinalOrphans, id)
	}

	// Validate tree integrity
	if page.Root == nil {
		return fmt.Errorf("no root node found")
	}

	// Report results
	if len(stats.CircularRefs) > 0 {
		return fmt.Errorf("circular references detected in nodes: %v", stats.CircularRefs)
	}

	if len(stats.FinalOrphans) > 0 {
		return fmt.Errorf("unresolved orphan nodes: %v (parents may not exist)", stats.FinalOrphans)
	}

	return nil
}

func (page *Root) InsertNodeByIndex(parent, firstChild, prevSibling, nextSibling, n *Node) {
	page.NodeIndex[n.Instance.ContentDataID] = n
	n.Parent = parent
	n.FirstChild = firstChild
	n.PrevSibling = prevSibling
	n.NextSibling = nextSibling

}

// Functions for working with Nodes in a tree where access is constant

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
		// traverse target.FirstChild.NextSibling until NextSibling == nil, LastSibling
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
	// else (No children but is first child)
	if target.NextSibling != nil {
		target.Parent.FirstChild = target.NextSibling
		target.NextSibling.PrevSibling = nil
		return true
	}
	// if target.NextSibling = nil
	target.Parent.FirstChild = nil
	return true
}

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
	// else (No children but isn't first child)
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

// CountVisible counts the number of visible nodes in the tree
func (page *Root) CountVisible() int {
	if page.Root == nil {
		return 0
	}
	count := 0
	page.countNodesRecursive(page.Root, &count)
	return count
}

// countNodesRecursive recursively counts visible nodes
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

// NodeAtIndex returns the node at the given visible index
func (page *Root) NodeAtIndex(index int) *Node {
	if page.Root == nil {
		return nil
	}
	currentIndex := 0
	return page.nodeAtIndex(page.Root, index, &currentIndex)
}

// nodeAtIndex finds the node at the given index
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
