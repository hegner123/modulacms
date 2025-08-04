package cli

import "github.com/hegner123/modulacms/internal/db"

type TreeRoot struct {
	Root       TreeNode
	NodeIndex  map[int64]*TreeNode
}

type TreeNode struct {
	Node           db.ContentData
	NodeFields     []db.ContentFields
	NodeDatatype   db.Datatypes
	NodeFieldTypes []db.Fields
	Nodes          *[]*TreeNode
}

func NewTreeRoot(root TreeNode) *TreeRoot {
	return &TreeRoot{
		Root:       root,
	}
}

func (page *TreeRoot) Insert(n TreeNode, parent int64) bool {
	if page.Root.Nodes == nil {
		if page.Root.Node.ContentDataID == parent {
			nn := make([]*TreeNode, 0)
			nn = append(nn, &n)
			page.Root.Nodes = &nn
			page.NodeIndex[n.Node.ContentDataID] = &n
			return true
		}
		return false
	} else if page.Root.Node.ContentDataID == parent {
		instance := *page.Root.Nodes
		instance = append(instance, &n)
		page.Root.Nodes = &instance
		page.NodeIndex[n.Node.ContentDataID] = &n
		return true
	}
	instance := *page.Root.Nodes
	for _, v := range instance {
		if v.Node.ContentDataID == parent {
			res := v.Insert(n, parent)
			if res {
				page.NodeIndex[n.Node.ContentDataID] = &n
				return res
			}

			return res
		}
	}
	return false

}

func (node *TreeNode) Insert(n TreeNode, parent int64) bool {
	if node.Nodes == nil {
		if node.Node.ContentDataID == parent {
			nodeSlice := make([]*TreeNode, 0)
			nodeSlice = append(nodeSlice, &n)
			node.Nodes = &nodeSlice
			return true
		}
		return false
	}
	if node.Node.ContentDataID == parent {
		nodeSlice := *node.Nodes
		nodeSlice = append(nodeSlice, &n)
		node.Nodes = &nodeSlice
		return true
	} else {
		nodeSlice := *node.Nodes
		for _, v := range nodeSlice {
			if v.Node.ContentDataID == parent {
				return v.Insert(n, parent)

			}

		}
	}
	return false

}

func (page *TreeRoot) NodeInsertByIndex(index *TreeNode, n TreeNode) {
	if index.Nodes == nil {
		nodeSlice := make([]*TreeNode, 0)
		nodeSlice = append(nodeSlice, &n)
		index.Nodes = &nodeSlice
	} else {
		nodeSlice := *index.Nodes
		nodeSlice = append(nodeSlice, &n)
		index.Nodes = &nodeSlice
	}
	page.NodeIndex[n.Node.ContentDataID] = &n

}

func (page TreeRoot) GetContentByRouteID(id int64) ([]db.ContentData, error) {
	out := make([]db.ContentData, 0)
	return out, nil
}

