# CONTENT_TREES.md

Domain guide for content tree operations in ModulaCMS.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/CONTENT_TREES.md`
**Purpose:** Practical guide to tree navigation, manipulation, and query patterns
**Last Updated:** 2026-01-12

---

## Overview

Content in ModulaCMS is organized as trees using sibling pointers. Each route has one tree root, and content nodes form parent-child hierarchies. This document covers practical tree operations.

**See Also:** [TREE_STRUCTURE.md](../architecture/TREE_STRUCTURE.md) for implementation details.

---

## Database Schema

Content trees are stored in the `content_data` table with sibling pointers:

```sql
CREATE TABLE content_data (
    content_data_id INTEGER PRIMARY KEY,
    parent_id INTEGER REFERENCES content_data,
    first_child_id INTEGER REFERENCES content_data,
    next_sibling_id INTEGER REFERENCES content_data,
    prev_sibling_id INTEGER REFERENCES content_data,
    route_id INTEGER NOT NULL,
    datatype_id INTEGER NOT NULL,
    -- other fields...
);
```

**Pointer Relationships:**
- `parent_id` - Links to parent node (NULL for root)
- `first_child_id` - First child in sibling chain
- `next_sibling_id` - Next sibling
- `prev_sibling_id` - Previous sibling (bi-directional)

---

## In-Memory Representation

The TUI maintains trees in memory with expanded references:

```go
// internal/cli/cms_struct.go
type TreeNode struct {
	Instance       *db.ContentData      // Database row
	InstanceFields []db.ContentFields   // Field values
	Datatype       db.Datatypes         // Type metadata
	Fields         []db.Fields          // Field definitions
	Parent         *TreeNode
	FirstChild     *TreeNode
	NextSibling    *TreeNode
	PrevSibling    *TreeNode
	Expand         bool                 // UI state
	Indent         int                  // Display level
	Wrapped        int                  // Line wrap
}

type TreeRoot struct {
	Root      *TreeNode              // Tree root
	NodeIndex map[int64]*TreeNode    // O(1) lookup
	Orphans   map[int64]*TreeNode    // Unresolved nodes
	MaxRetry  int                    // Loop detection
}
```

---

## Loading Trees

Trees load in three phases:

**Phase 1: Create Nodes**
```go
// internal/cli/cms_struct.go
for _, row := range *rows {
	node := NewTreeNodeFromContentTree(row)
	page.NodeIndex[node.Instance.ContentDataID] = node

	if !node.Instance.ParentID.Valid {
		page.Root = node
	}
}
```

**Phase 2: Assign Hierarchy**
```go
for id, node := range page.NodeIndex {
	if node.Instance.ParentID.Valid {
		parentID := node.Instance.ParentID.Int64
		parent := page.NodeIndex[parentID]

		if parent != nil {
			page.attachNodeToParent(node, parent)
		} else {
			page.Orphans[id] = node
		}
	}
}
```

**Phase 3: Resolve Orphans**
```go
for len(page.Orphans) > 0 && attempts < page.MaxRetry {
	for id, orphan := range page.Orphans {
		parent := page.NodeIndex[orphan.Instance.ParentID.Int64]
		if parent != nil {
			page.attachNodeToParent(orphan, parent)
			delete(page.Orphans, id)
		}
	}
	// Detect circular references if no progress
}
```

---

## Tree Queries

**Full Tree Load** (`sql/schema/22_joins/queries.sql`):
```sql
-- name: GetContentTreeByRoute :many
SELECT cd.content_data_id,
        cd.parent_id,
        cd.first_child_id,
        cd.next_sibling_id,
        cd.prev_sibling_id,
        cd.datatype_id,
        dt.label as datatype_label
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
ORDER BY cd.parent_id NULLS FIRST;
```

**Shallow Load** (lazy loading - root + first level only):
```sql
-- name: GetShallowTreeByRouteId :many
SELECT cd.*, dt.label as datatype_label
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
AND (cd.parent_id IS NULL
     OR cd.parent_id IN (SELECT content_data_id
                         FROM content_data
                         WHERE parent_id IS NULL
                         AND route_id = ?))
ORDER BY cd.parent_id NULLS FIRST;
```

---

## Lazy Loading

**Problem:** Loading 500 nodes when only 5 are visible is wasteful.

**Solution:** Load progressively as user expands nodes.

```go
type ExpandNodeMsg struct {
    NodeID       int64
    LoadChildren bool
}

func Update(msg Msg, model Model) (Model, Cmd) {
    switch m := msg.(type) {
    case ExpandNodeMsg:
        if m.LoadChildren {
            return model, LoadNodeChildrenCmd(m.NodeID, 1)
        }
    }
}
```

**Performance:**
- Initial load: ~20 nodes (4KB) vs 500 nodes (100KB) = **25x reduction**
- Load only explored branches

---

## Tree Navigation

**Traversing Siblings:**
```go
current := parent.FirstChild
for current != nil {
	// Process current
	fmt.Println(current.Datatype.Label)
	current = current.NextSibling
}
```

**Finding Node by ID (O(1)):**
```go
node := treeRoot.NodeIndex[contentDataID]
```

**TUI Rendering** (`internal/cli/page_builders.go`):
```go
func FormatRow(node *TreeNode) string {
	indent := strings.Repeat("  ", node.Indent)
	return indent + DecideNodeName(*node)
}

func DecideNodeName(node TreeNode) string {
	// Finds "label" field value for display
	// Falls back to datatype label
}
```

---

## Tree Manipulation

**Attaching Node to Parent:**
```go
// internal/cli/cms_struct.go
func (page *TreeRoot) attachNodeToParent(node, parent *TreeNode) {
	node.Parent = parent

	if parent.FirstChild == nil {
		parent.FirstChild = node
	} else {
		// Find last sibling
		current := parent.FirstChild
		for current.NextSibling != nil {
			current = current.NextSibling
		}
		current.NextSibling = node
		node.PrevSibling = current
	}
}
```

**Deleting Nodes:**

Deletion handles multiple cases:
1. First child with/without children
2. Nested child with/without children
3. Reparenting orphaned children

```go
// internal/cli/cms_struct.go
func (page *TreeRoot) DeleteTreeNodeByIndex(n *TreeNode) bool {
	target := page.NodeIndex[n.Instance.ContentDataID]
	if target == nil || target == page.Root {
		return false  // Can't delete root
	}

	if target.Parent.FirstChild == target {
		DeleteFirstChild(target)
	} else {
		DeleteNestedChild(target)
	}

	delete(page.NodeIndex, n.Instance.ContentDataID)
	return true
}
```

See `internal/cli/cms_struct.go` for complete deletion logic handling sibling chain updates and child reparenting.

---

## Practical Example

**Database:**
```
ID | ParentID | FirstChildID | NextSiblingID | Datatype
1  | NULL     | 4            | NULL          | Page
4  | 1        | 6            | 5             | Hero
5  | 1        | NULL         | NULL          | Footer
6  | 4        | NULL         | 7             | Image
7  | 4        | NULL         | NULL          | Text
```

**Tree Structure:**
```
Page (id=1) [root]
├─ Hero (id=4)
│  ├─ Image (id=6)
│  └─ Text (id=7)
└─ Footer (id=5)
```

**Sibling Chains:**
- Hero.NextSibling = Footer
- Image.NextSibling = Text

---

## Circular Reference Detection

Prevents infinite loops during tree loading:

```go
func (page *TreeRoot) hasCircularReference(
	node *TreeNode, visited map[int64]bool) bool {
	nodeID := node.Instance.ContentDataID

	if visited[nodeID] {
		return true  // Cycle detected
	}

	if !node.Instance.ParentID.Valid {
		return false  // Reached root
	}

	visited[nodeID] = true
	parent := page.NodeIndex[node.Instance.ParentID.Int64]

	if parent == nil {
		return false
	}

	return page.hasCircularReference(parent, visited)
}
```

---

## JSON Serialization

API responses use a different structure:

```go
// internal/model/model.go
type Node struct {
	Datatype Datatype `json:"datatype"`
	Fields   []Field  `json:"fields"`
	Nodes    []*Node  `json:"nodes"`  // Direct child array
}
```

**Difference:**
- **TUI:** Sibling pointers for O(1) operations
- **JSON:** Child arrays for API consumers

---

## Common Patterns

**Creating New Content Node:**
1. Create TreeNode with database row
2. Set parent reference
3. Use `attachNodeToParent()` to link into sibling chain
4. Add to NodeIndex
5. Update database with pointer fields

**Moving Node:**
1. Remove from current parent's sibling chain
2. Update old siblings' pointers
3. Attach to new parent
4. Update database

**Expanding Node (TUI):**
1. Check if children loaded
2. If not, query database for children
3. Attach children to node
4. Set Expand = true
5. Re-render view

---

## Related Documentation

- **[TREE_STRUCTURE.md](../architecture/TREE_STRUCTURE.md)** - Sibling-pointer implementation details
- **[CONTENT_MODEL.md](../architecture/CONTENT_MODEL.md)** - Domain model relationships
- **[CLI_PACKAGE.md](../packages/CLI_PACKAGE.md)** - TUI tree navigation
- **[ROUTES_AND_SITES.md](ROUTES_AND_SITES.md)** - Route as tree root concept

---

## Quick Reference

**Key Files:**
- `internal/cli/cms_struct.go` - TreeNode, TreeRoot, loading, deletion
- `internal/cli/page_builders.go` - Tree rendering
- `internal/model/model.go` - JSON tree structure
- `sql/schema/16_content_data/schema.sql` - Database schema
- `sql/schema/22_joins/queries.sql` - Tree queries

**Key Operations:**
- Load tree: Three-phase algorithm
- Find node: `NodeIndex[id]` - O(1)
- Traverse: Follow NextSibling pointers
- Delete: Handle sibling chains and reparenting
- Lazy load: Progressive expansion on demand
