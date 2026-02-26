package core

import (
	"github.com/hegner123/modulacms/internal/db/types"
)

// CountVisible counts the number of visible nodes in the tree.
// The isExpanded callback determines whether a node's children are visible.
// Pass func(*Node) bool { return true } to count all nodes.
func CountVisible(root *Node, isExpanded func(*Node) bool) int {
	if root == nil {
		return 0
	}
	count := 0
	countVisibleRecursive(root, isExpanded, &count)
	return count
}

func countVisibleRecursive(node *Node, isExpanded func(*Node) bool, count *int) {
	if node == nil {
		return
	}
	*count++

	if isExpanded(node) && node.FirstChild != nil {
		countVisibleRecursive(node.FirstChild, isExpanded, count)
	}

	if node.NextSibling != nil {
		countVisibleRecursive(node.NextSibling, isExpanded, count)
	}
}

// FlattenVisible returns a flat slice of all visible nodes in display order,
// respecting the expand state provided by the callback.
func FlattenVisible(root *Node, isExpanded func(*Node) bool) []*Node {
	if root == nil {
		return nil
	}
	var result []*Node
	flattenVisibleRecursive(root, isExpanded, &result)
	return result
}

func flattenVisibleRecursive(node *Node, isExpanded func(*Node) bool, result *[]*Node) {
	if node == nil {
		return
	}
	*result = append(*result, node)

	if isExpanded(node) && node.FirstChild != nil {
		flattenVisibleRecursive(node.FirstChild, isExpanded, result)
	}

	if node.NextSibling != nil {
		flattenVisibleRecursive(node.NextSibling, isExpanded, result)
	}
}

// NodeAtIndex returns the node at the given visible index.
// Returns nil if the index is out of range.
func NodeAtIndex(root *Node, index int, isExpanded func(*Node) bool) *Node {
	if root == nil {
		return nil
	}
	currentIndex := 0
	return nodeAtIndexRecursive(root, index, isExpanded, &currentIndex)
}

func nodeAtIndexRecursive(node *Node, targetIndex int, isExpanded func(*Node) bool, currentIndex *int) *Node {
	if node == nil {
		return nil
	}

	if *currentIndex == targetIndex {
		return node
	}
	*currentIndex++

	if isExpanded(node) && node.FirstChild != nil {
		if result := nodeAtIndexRecursive(node.FirstChild, targetIndex, isExpanded, currentIndex); result != nil {
			return result
		}
	}

	if node.NextSibling != nil {
		return nodeAtIndexRecursive(node.NextSibling, targetIndex, isExpanded, currentIndex)
	}

	return nil
}

// FindByContentID returns the node with the given ContentID via O(1) index lookup.
// Returns nil if the ID is not found.
func FindByContentID(root *Root, id types.ContentID) *Node {
	if root == nil {
		return nil
	}
	return root.NodeIndex[id]
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

// FindVisibleIndex returns the visible index of the node with the given
// content ID, or -1 if the node is not currently visible.
func FindVisibleIndex(root *Node, contentID types.ContentID, isExpanded func(*Node) bool) int {
	visible := FlattenVisible(root, isExpanded)
	for i, n := range visible {
		if n.ContentData != nil && n.ContentData.ContentDataID == contentID {
			return i
		}
	}
	return -1
}
