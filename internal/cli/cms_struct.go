package cli

import "github.com/hegner123/modulacms/internal/db"

type TreeRoot struct {
	Root      *TreeNode
	NodeIndex map[int64]*TreeNode
}

type TreeNode struct {
	Instance       *db.ContentData
	InstanceFields []db.ContentFields
	Datatype       db.Datatypes
	Fields         []db.Fields
	FirstChild     *TreeNode
	NextSibling    *TreeNode
	Expand         bool
	Indent         int
	Wrapped        int
}

func NewTreeRoot() *TreeRoot {
	return &TreeRoot{
		NodeIndex: make(map[int64]*TreeNode),
	}
}

func NewTreeNode(row db.GetRouteTreeByRouteIDRow) *TreeNode {
	cd := db.ContentData{
		ContentDataID: row.ContentDataID,
		ParentID:      row.ParentID,
	}

	return &TreeNode{
		Instance: &cd,
	}

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
		DatatypeID: row.DatatypeID,
		Label:      row.DatatypeLabel,
		Type:       row.DatatypeType,
	}

	return &TreeNode{
		Instance: &cd,
		Datatype: dt,
	}
}

func (page *TreeRoot) Insert(n TreeNode) bool {
	if page.Root == nil {
		if !n.Instance.ParentID.Valid { // null parent = root
			page.Root = &n
			page.NodeIndex[n.Instance.ContentDataID] = page.Root
			return true
		}
		return false // can't insert non-root when no root exists
	}

	if page.Root.InsertRecursive(n) {
		page.NodeIndex[n.Instance.ContentDataID] = &n
		return true
	}
	return false
}

func (node *TreeNode) InsertRecursive(n TreeNode) bool {
	if !n.Instance.ParentID.Valid {
		return false
	}

	if node.Instance.ContentDataID == n.Instance.ParentID.Int64 {
		node.AddNode(n)
		return true
	}

	// Check children first
	if node.FirstChild != nil {
		if node.FirstChild.InsertRecursive(n) {
			return true
		}
	}

	// Then check siblings
	if node.NextSibling != nil {
		return node.NextSibling.InsertRecursive(n)
	}

	return false
}
func (node *TreeNode) AddNode(n TreeNode) bool {
	if node.FirstChild == nil {
		node.FirstChild = &n
		return true
	}
	return node.FirstChild.AddSibling(n) // traverse to end of sibling chain
}

func (node *TreeNode) AddSibling(n TreeNode) bool {
	if node.NextSibling == nil {
		node.NextSibling = &n
		return true
	} else {
		return node.NextSibling.AddSibling(n)
	}
}


func (page *TreeRoot) NodeInsertByIndex(index *TreeNode, n TreeNode) {
	page.NodeIndex[n.Instance.ContentDataID] = &n
}

