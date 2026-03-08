package tui

// =============================================================================
// CONTENT TREE POINTER HELPERS
// =============================================================================
//
// Tree pointer manipulation has been consolidated into content_ops.go using the
// treeOps abstraction. The three unified functions (detachFromSiblings,
// attachAsLastChild, swapSiblings) replace the 6 duplicated regular/admin
// variants that previously lived here.
//
// See content_ops.go for:
//   - treeOps, treeNode, treeNullID types
//   - newContentTreeOps() / newAdminTreeOps() constructors
//   - detachFromSiblings(), attachAsLastChild(), swapSiblings()
