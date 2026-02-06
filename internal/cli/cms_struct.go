package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

type TreeRoot struct {
	Root      *TreeNode
	NodeIndex map[types.ContentID]*TreeNode
	Orphans   map[types.ContentID]*TreeNode
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

type TreeNode struct {
	Instance       *db.ContentData
	InstanceFields []db.ContentFields
	Datatype       db.Datatypes
	Fields         []db.Fields
	Parent         *TreeNode
	FirstChild     *TreeNode
	NextSibling    *TreeNode
	PrevSibling    *TreeNode
	Expand         bool
	Indent         int
	Wrapped        int
}

func NewTreeRoot() *TreeRoot {
	return &TreeRoot{
		NodeIndex: make(map[types.ContentID]*TreeNode),
		Orphans:   make(map[types.ContentID]*TreeNode),
		MaxRetry:  100,
	}
}

func NewTreeNode(row db.GetRouteTreeByRouteIDRow) *TreeNode {
	cd := db.ContentData{
		ContentDataID: row.ContentDataID,
		ParentID:      row.ParentID,
	}

	return &TreeNode{
		Instance: &cd,
		Expand:   true, // Expanded by default
	}

}

func (page *TreeRoot) LoadFromRows(rows *[]db.GetContentTreeByRouteRow) (*LoadStats, error) {
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

func (page *TreeRoot) createAllNodes(rows *[]db.GetContentTreeByRouteRow, stats *LoadStats) error {
	for _, row := range *rows {
		node := NewTreeNodeFromContentTree(row)
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
func (page *TreeRoot) assignImmediateHierarchy(stats *LoadStats) error {
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
func (page *TreeRoot) resolveOrphans(stats *LoadStats) error {
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
func (page *TreeRoot) detectCircularReferences(stats *LoadStats) bool {
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
func (page *TreeRoot) hasCircularReference(node *TreeNode, visited map[types.ContentID]bool) bool {
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
func (page *TreeRoot) attachNodeToParent(node, parent *TreeNode) {
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
func (page *TreeRoot) validateFinalState(stats *LoadStats) error {
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

func NewTreeNodeFromContentTree(row db.GetContentTreeByRouteRow) *TreeNode {
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

	return &TreeNode{
		Instance: &cd,
		Datatype: dt,
		Expand:   true, // Expanded by default
	}
}

func (page *TreeRoot) InsertTreeNodeByIndex(parent, firstChild, prevSibling, nextSibling, n *TreeNode) {
	page.NodeIndex[n.Instance.ContentDataID] = n
	n.Parent = parent
	n.FirstChild = firstChild
	n.PrevSibling = prevSibling
	n.NextSibling = nextSibling

}

// Functions for working with Nodes in a tree where access is constant

func (page *TreeRoot) DeleteTreeNodeByIndex(n *TreeNode) bool {
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

func DeleteFirstChild(target *TreeNode) bool {
	if target.FirstChild != nil {
		return DeleteFirstChildHasChildren(target)
	} else {
		return DeleteFirstChildNoChildren(target)

	}
}

func DeleteFirstChildHasChildren(target *TreeNode) bool {
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
	} else {
		target.Parent.FirstChild = target.FirstChild
		current := target.FirstChild
		for current != nil {
			current.Parent = target.Parent
			current = current.NextSibling
		}
		return true
	}
}

func DeleteFirstChildNoChildren(target *TreeNode) bool {
	// else (No children but is first child)
	if target.NextSibling != nil {
		target.Parent.FirstChild = target.NextSibling
		target.NextSibling.PrevSibling = nil
		return true
	} else {
		// if target.NextSibling = nil
		target.Parent.FirstChild = nil
		return true
	}

}

func DeleteNestedChild(target *TreeNode) bool {
	if target.FirstChild != nil {
		return DeleteNestedChildHasChildren(target)
	} else {
		return DeleteNestedChildNoChildren(target)
	}

}
func DeleteNestedChildHasChildren(target *TreeNode) bool {
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
	} else {
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

}

func DeleteNestedChildNoChildren(target *TreeNode) bool {
	// else (No children but isn't first child)
	if target.NextSibling != nil {
		if target.PrevSibling == nil {
			return false
		}
		target.PrevSibling.NextSibling = target.NextSibling
		target.NextSibling.PrevSibling = target.PrevSibling
		return true
	} else {
		if target.PrevSibling == nil {
			return false
		}
		target.PrevSibling.NextSibling = nil
		return true
	}
}

// Message types
type BuildTreeFromRouteMsg struct {
	RouteID int64
}

// Constructors
func BuildTreeFromRouteCMD(id int64) tea.Cmd {
	return func() tea.Msg {
		return BuildTreeFromRouteMsg{
			RouteID: id,
		}
	}
}
