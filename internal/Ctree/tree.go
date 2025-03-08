package ctree

import (
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

type (
	Scanner func(parent, child mdb.ListAdminDatatypeByRouteIdRow) int
)

func Scan(parent, child mdb.ListAdminDatatypeByRouteIdRow) int {
	if parent.AdminDatatypeID == child.ParentID.Int64 {
		return 1
	} else {
		return 0
	}
}

type Tree struct {
	fn   Scanner
	Root *Node `json:"root"`
}

type Node struct {
	Val      mdb.ListAdminDatatypeByRouteIdRow `json:"node"`
	Children []*Node                           `json:"children"`
	Fields   []mdb.AdminFields                 `json:"fields"`
}

func NewTree(scanFunction Scanner) *Tree {
	return &Tree{
		fn: scanFunction,
	}
}

func (tree *Tree) Add(dt mdb.ListAdminDatatypeByRouteIdRow) {
	tree.Root = tree.Root.Add(tree.fn, dt)
}

func (parent *Node) Add(compare Scanner, child mdb.ListAdminDatatypeByRouteIdRow) *Node {
	if parent == nil {
		return &Node{Val: child}
	}
	switch r := compare(parent.Val, child); {
	case r == 1:
		parent.Children = append(parent.Children, &Node{Val: child})
	case r == 0:
		for i := range parent.Children {
			parent.Children[i].Add(compare, child)
		}
	}

	return parent
}

func (node *Node) PrintTree(level int) {
	if node == nil {
		return
	}
	indent := ""
	for range level {
		indent += "  "
	}
	fmt.Printf("%s%d (%s)\n", indent, node.Val.AdminDatatypeID, node.Val.Label)
	for _, v := range node.Fields {
		fmt.Printf("%s%d (%s)\n", "|"+indent, v.AdminFieldID, v.Label)
	}
	for _, junior := range node.Children {
		junior.PrintTree(level + 1)
	}
}
