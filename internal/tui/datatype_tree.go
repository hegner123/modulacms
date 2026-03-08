package tui

import (
	"sort"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
)

// DatatypeNodeKind discriminates item and group nodes in the datatype tree.
type DatatypeNodeKind int

const (
	DatatypeNodeItem  DatatypeNodeKind = iota // leaf datatype (no children)
	DatatypeNodeGroup                         // datatype with children
)

// DatatypeTreeNode represents a node in the parent_id-grouped datatype tree.
type DatatypeTreeNode struct {
	Kind       DatatypeNodeKind
	Label      string
	Depth      int
	Expand     bool
	Datatype   *db.Datatypes       // non-nil for regular mode
	AdminDT    *db.AdminDatatypes  // non-nil for admin mode
	FieldCount int                 // cached count from field preview
	Children   []*DatatypeTreeNode // child datatypes (sorted by sort_order, then label)
}

// DatatypeID returns the ID string regardless of mode.
func (n *DatatypeTreeNode) DatatypeID() string {
	if n.Datatype != nil {
		return string(n.Datatype.DatatypeID)
	}
	if n.AdminDT != nil {
		return string(n.AdminDT.AdminDatatypeID)
	}
	return ""
}

// BuildDatatypeTree converts a flat list of datatypes into a parent_id-grouped tree.
func BuildDatatypeTree(items []db.Datatypes) []*DatatypeTreeNode {
	if len(items) == 0 {
		return nil
	}

	// Index nodes by ID
	nodeMap := make(map[string]*DatatypeTreeNode, len(items))
	for i := range items {
		dt := items[i]
		nodeMap[string(dt.DatatypeID)] = &DatatypeTreeNode{
			Label:    dt.Label,
			Expand:   true,
			Datatype: &dt,
		}
	}

	// Build tree
	var roots []*DatatypeTreeNode
	for i := range items {
		dt := items[i]
		node := nodeMap[string(dt.DatatypeID)]
		if dt.ParentID.Valid {
			parent, ok := nodeMap[string(dt.ParentID.ID)]
			if ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	// Sort children and assign Kind
	var assignKind func(nodes []*DatatypeTreeNode)
	assignKind = func(nodes []*DatatypeTreeNode) {
		for _, n := range nodes {
			sortDatatypeNodes(n.Children)
			if len(n.Children) > 0 {
				n.Kind = DatatypeNodeGroup
				assignKind(n.Children)
			}
		}
	}
	sortDatatypeNodes(roots)
	assignKind(roots)

	return roots
}

// BuildAdminDatatypeTree converts a flat list of admin datatypes into a parent_id-grouped tree.
func BuildAdminDatatypeTree(items []db.AdminDatatypes) []*DatatypeTreeNode {
	if len(items) == 0 {
		return nil
	}

	nodeMap := make(map[string]*DatatypeTreeNode, len(items))
	for i := range items {
		dt := items[i]
		nodeMap[string(dt.AdminDatatypeID)] = &DatatypeTreeNode{
			Label:   dt.Label,
			Expand:  true,
			AdminDT: &dt,
		}
	}

	var roots []*DatatypeTreeNode
	for i := range items {
		dt := items[i]
		node := nodeMap[string(dt.AdminDatatypeID)]
		if dt.ParentID.Valid {
			parent, ok := nodeMap[string(dt.ParentID.ID)]
			if ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	var assignKind func(nodes []*DatatypeTreeNode)
	assignKind = func(nodes []*DatatypeTreeNode) {
		for _, n := range nodes {
			sortDatatypeNodes(n.Children)
			if len(n.Children) > 0 {
				n.Kind = DatatypeNodeGroup
				assignKind(n.Children)
			}
		}
	}
	sortDatatypeNodes(roots)
	assignKind(roots)

	return roots
}

// FlattenDatatypeTree produces a depth-first flat list from the tree roots.
// Collapsed nodes (Expand=false) hide their children.
func FlattenDatatypeTree(roots []*DatatypeTreeNode) []*DatatypeTreeNode {
	var result []*DatatypeTreeNode
	var walk func(nodes []*DatatypeTreeNode, depth int)
	walk = func(nodes []*DatatypeTreeNode, depth int) {
		for _, n := range nodes {
			n.Depth = depth
			result = append(result, n)
			if n.Expand && len(n.Children) > 0 {
				walk(n.Children, depth+1)
			}
		}
	}
	walk(roots, 0)
	return result
}

// FilterDatatypeList returns datatypes whose Label contains query (case-insensitive).
func FilterDatatypeList(items []db.Datatypes, query string) []db.Datatypes {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	var result []db.Datatypes
	for _, dt := range items {
		if strings.Contains(strings.ToLower(dt.Label), q) {
			result = append(result, dt)
		}
	}
	return result
}

// FilterAdminDatatypeList returns admin datatypes whose Label contains query (case-insensitive).
func FilterAdminDatatypeList(items []db.AdminDatatypes, query string) []db.AdminDatatypes {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	var result []db.AdminDatatypes
	for _, dt := range items {
		if strings.Contains(strings.ToLower(dt.Label), q) {
			result = append(result, dt)
		}
	}
	return result
}

// sortDatatypeNodes sorts nodes by sort_order, then label.
func sortDatatypeNodes(nodes []*DatatypeTreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		a, b := sortOrderOf(nodes[i]), sortOrderOf(nodes[j])
		if a != b {
			return a < b
		}
		return strings.ToLower(nodes[i].Label) < strings.ToLower(nodes[j].Label)
	})
}

func sortOrderOf(n *DatatypeTreeNode) int64 {
	if n.Datatype != nil {
		return n.Datatype.SortOrder
	}
	if n.AdminDT != nil {
		return n.AdminDT.SortOrder
	}
	return 0
}
