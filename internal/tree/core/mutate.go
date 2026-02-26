package core

import (
	"github.com/hegner123/modulacms/internal/db/types"
)

// InsertNode adds a node to the tree, setting its parent and sibling pointers,
// and registers it in the NodeIndex.
func InsertNode(root *Root, parent, firstChild, prevSibling, nextSibling, node *Node) {
	if node.ContentData == nil {
		return
	}
	root.NodeIndex[node.ContentData.ContentDataID] = node
	node.Parent = parent
	node.FirstChild = firstChild
	node.PrevSibling = prevSibling
	node.NextSibling = nextSibling
}

// DeleteNode removes a node from the tree, re-linking its children and siblings.
// Returns true if the node was successfully deleted, false otherwise.
// The root node itself cannot be deleted.
func DeleteNode(root *Root, node *Node) bool {
	if node == nil || node.ContentData == nil {
		return false
	}
	target := root.NodeIndex[node.ContentData.ContentDataID]
	if root.Node == nil || target == nil || root.Node == target || target.Parent == nil {
		return false
	}

	if target.Parent.FirstChild == target {
		deleteFirstChild(target)
		delete(root.NodeIndex, node.ContentData.ContentDataID)
		return true
	}

	// Walk siblings to confirm target is actually a child of its parent
	current := target.Parent.FirstChild
	for current != nil && current != target {
		current = current.NextSibling
	}
	if current == nil {
		return false
	}

	deleteNestedChild(target)
	delete(root.NodeIndex, node.ContentData.ContentDataID)
	return true
}

// deleteFirstChild handles deletion when the target is the first child of its parent.
func deleteFirstChild(target *Node) bool {
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

// deleteNestedChild handles deletion when the target is not the first child.
func deleteNestedChild(target *Node) bool {
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

// RemoveFromIndex removes a node from the index without unlinking it from the tree.
// Used when a caller needs to manage unlinking separately.
func RemoveFromIndex(root *Root, id types.ContentID) {
	delete(root.NodeIndex, id)
}
