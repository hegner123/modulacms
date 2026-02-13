# tree

Package tree implements a sibling-pointer content tree structure with lazy loading, orphan resolution, and circular reference detection. It provides O(1) node access via an in-memory index and supports efficient tree traversal, node insertion, and deletion.

## Overview

The tree package manages hierarchical content data loaded from database rows. It uses a three-phase loading algorithm to construct the tree, resolve orphaned nodes, and reorder siblings according to stored database pointers. Each node stores references to parent, first child, previous sibling, and next sibling for constant-time navigation.

The package handles edge cases including circular references, orphaned nodes with missing parents, and partial sibling chains. It provides utilities for counting visible nodes, finding nodes by index, and flattening the tree for display.

## Types

### Root

Root represents the entire content tree and manages node indexing, orphan tracking, and tree construction.

```go
type Root struct {
    Root      *Node
    NodeIndex map[types.ContentID]*Node
    Orphans   map[types.ContentID]*Node
    MaxRetry  int
    rootNodes []*Node
}
```

Fields:
- Root: Primary root node of the tree
- NodeIndex: Map from ContentID to Node pointer for O(1) access
- Orphans: Nodes whose parents are not yet loaded or do not exist
- MaxRetry: Maximum orphan resolution attempts (default 100)
- rootNodes: All parentless nodes collected during loading

### LoadStats

LoadStats tracks statistics from tree loading and reports diagnostics.

```go
type LoadStats struct {
    NodesCount      int
    OrphansResolved int
    RetryAttempts   int
    CircularRefs    []types.ContentID
    FinalOrphans    []types.ContentID
}
```

Fields:
- NodesCount: Total nodes created
- OrphansResolved: Orphans successfully attached during resolution phase
- RetryAttempts: Number of orphan resolution cycles executed
- CircularRefs: ContentIDs forming circular parent references
- FinalOrphans: ContentIDs that remain orphaned after all retries

### Node

Node represents a single content item in the tree with pointers for navigation and display state.

```go
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
```

Fields:
- Instance: Content data from database
- InstanceFields: Field values for this content instance
- Datatype: Type definition for this content
- Fields: Field schema definitions
- Parent: Parent node pointer, nil for root nodes
- FirstChild: First child in the child list
- NextSibling: Next sibling at same level
- PrevSibling: Previous sibling at same level
- Expand: Display state for tree UI expansion
- Indent: Display indentation level
- Wrapped: Line wrapping state for rendering

## Functions

### NewRoot

NewRoot creates a new empty Root with initialized maps and default MaxRetry of 100.

```go
func NewRoot() *Root
```

Returns a Root ready for loading tree data.

### NewNode

NewNode creates a Node from a GetRouteTreeByRouteIDRow database result.

```go
func NewNode(row db.GetRouteTreeByRouteIDRow) *Node
```

Sets Instance with ContentDataID and ParentID. Sets Expand to true by default.

### NewNodeFromContentTree

NewNodeFromContentTree creates a Node from a GetContentTreeByRouteRow database result with full sibling pointer data.

```go
func NewNodeFromContentTree(row db.GetContentTreeByRouteRow) *Node
```

Populates Instance with all ContentData fields and Datatype. Sets Expand to true.

### DeleteFirstChild

DeleteFirstChild removes a target node that is the first child of its parent.

```go
func DeleteFirstChild(target *Node) bool
```

Handles two cases: target has children (promotes them to parent) or target has no children (removes from parent). Returns true on success.

### DeleteNestedChild

DeleteNestedChild removes a target node that is not the first child of its parent.

```go
func DeleteNestedChild(target *Node) bool
```

Handles nodes with and without children, promoting children to parent level if present. Updates sibling pointers. Returns true on success, false if preconditions fail.

### IsDescendantOf

IsDescendantOf checks if node is a descendant of ancestor by walking up the parent chain.

```go
func IsDescendantOf(node, ancestor *Node) bool
```

Returns true if ancestor is found in the parent chain, false otherwise.

## Root Methods

### LoadFromRows

LoadFromRows constructs the tree from database rows using a four-phase algorithm.

```go
func (page *Root) LoadFromRows(rows *[]db.GetContentTreeByRouteRow) (*LoadStats, error)
```

Phase 1 creates all nodes and populates NodeIndex. Phase 2 assigns hierarchy for nodes with existing parents. Phase 3 iteratively resolves orphans. Phase 4 reorders siblings to match database pointers. Returns statistics and error if circular references or final orphans exist.

### InsertNodeByIndex

InsertNodeByIndex adds a node to the NodeIndex and sets all pointers.

```go
func (page *Root) InsertNodeByIndex(parent, firstChild, prevSibling, nextSibling, n *Node)
```

Updates NodeIndex, assigns parent, firstChild, prevSibling, and nextSibling for the new node n.

### DeleteNodeByIndex

DeleteNodeByIndex removes a node from the tree and NodeIndex.

```go
func (page *Root) DeleteNodeByIndex(n *Node) bool
```

Refuses to delete root node or nil nodes. Delegates to DeleteFirstChild or DeleteNestedChild based on node position. Returns true on successful deletion.

### CountVisible

CountVisible counts all visible nodes in the tree respecting Expand state.

```go
func (page *Root) CountVisible() int
```

Recursively traverses expanded nodes and their siblings. Returns total visible node count.

### NodeAtIndex

NodeAtIndex returns the node at the given visible index in display order.

```go
func (page *Root) NodeAtIndex(index int) *Node
```

Uses zero-based indexing. Returns nil if index is out of bounds or tree is empty.

### FlattenVisible

FlattenVisible returns a slice of all visible nodes in display order.

```go
func (page *Root) FlattenVisible() []*Node
```

Respects Expand state of each node. Returns nil if tree is empty.

### FindVisibleIndex

FindVisibleIndex returns the visible index of a node by ContentID.

```go
func (page *Root) FindVisibleIndex(contentID types.ContentID) int
```

Returns minus one if the node is not visible or does not exist.

## LoadStats Methods

### String

String formats LoadStats as a human-readable summary.

```go
func (stats LoadStats) String() string
```

Returns multiline string with node counts, orphan resolution stats, retry attempts, circular references, and final orphans.

## Internal Methods

Root includes unexported methods for tree construction phases:

- createAllNodes: Phase 1, populates NodeIndex and identifies root nodes
- assignImmediateHierarchy: Phase 2, attaches nodes with existing parents
- resolveOrphans: Phase 3, iteratively resolves orphaned nodes
- detectCircularReferences: Identifies circular parent reference chains
- hasCircularReference: Checks if a node forms a cycle
- attachNodeToParent: Links node to parent and sibling chain
- validateFinalState: Checks for remaining orphans and circular references
- reorderByPointers: Phase 4, rebuilds sibling order from database pointers
- buildSiblingChain: Constructs ordered list following NextSiblingID
- applySiblingOrder: Rewrites in-memory pointers to match database order
- reorderRootSiblings: Reorders root-level siblings
- countNodesRecursive: Helper for CountVisible
- nodeAtIndex: Helper for NodeAtIndex
- flattenVisibleRecursive: Helper for FlattenVisible

Internal deletion helpers:

- deleteFirstChildHasChildren: Promotes children when deleting first child
- deleteFirstChildNoChildren: Removes first child with no descendants
- deleteNestedChildHasChildren: Promotes children when deleting nested node
- deleteNestedChildNoChildren: Removes childless nested node
