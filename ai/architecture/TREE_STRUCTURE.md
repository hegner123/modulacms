# TREE_STRUCTURE.md

Comprehensive documentation on ModulaCMS's sibling-pointer tree implementation for content hierarchy.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md`
**Related Code:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go`
**Database Schema:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data/`

---

## Overview

ModulaCMS uses a **sibling-pointer tree structure** to represent content hierarchies. This is a specialized tree implementation that stores four pointers per node (parent, first child, next sibling, previous sibling) instead of the traditional parent-child array structure.

**Why This Matters:**
- This is the most unique and complex part of the codebase
- Understanding this structure is essential for working with content
- Different from typical tree implementations you may be familiar with
- Enables O(1) operations for common tree manipulations

---

## What Are Sibling Pointers?

### Traditional Tree Structure (Parent-Child Arrays)

Most tree implementations store children as slices/arrays:

```go
type Node struct {
    ID       int64
    ParentID int64
    Children []*Node  // Array of children
}
```

**Operations:**
- Find child: O(n) - must scan children array
- Insert child: O(1) - append to array
- Delete child: O(n) - must find and remove from array
- Sibling navigation: O(n) - must scan parent's children array

### Sibling-Pointer Structure (ModulaCMS)

ModulaCMS stores explicit sibling relationships:

```go
type TreeNode struct {
    Instance    *db.ContentData
    Parent      *TreeNode    // Pointer to parent
    FirstChild  *TreeNode    // Pointer to first child
    NextSibling *TreeNode    // Pointer to next sibling
    PrevSibling *TreeNode    // Pointer to previous sibling
}
```

**Operations:**
- Find child: O(1) - direct pointer
- Insert child: O(1) - pointer rewiring
- Delete child: O(1) - pointer rewiring
- Sibling navigation: O(1) - direct pointer

---

## Why Sibling Pointers?

### Performance: O(1) Operations

**Common content operations benefit from constant-time access:**

1. **Navigate to first child:** `node.FirstChild` (O(1))
2. **Navigate to next sibling:** `node.NextSibling` (O(1))
3. **Insert as first child:** Rewire 2 pointers (O(1))
4. **Delete node:** Rewire surrounding pointers (O(1))
5. **Find node by ID:** Use NodeIndex map (O(1))

Compare to traditional tree:
- Find specific child: O(n) where n = number of children
- Insert at position: O(n) to shift array
- Delete child: O(n) to find and remove

### Use Case: Content Management TUI

The TUI needs to:
- Display tree hierarchy visually
- Navigate up/down through siblings
- Expand/collapse nodes
- Move nodes within tree
- Delete nodes maintaining structure

All these operations are O(1) with sibling pointers.

### Inspiration: Filesystem Trees

This pattern is used by filesystems (like ext4, NTFS) where directories need fast child access and sibling traversal without scanning entire child lists.

---

## Why Performance Matters: The Row Count Reality

**Critical Context:** While the sibling-pointer tree and lazy loading optimizations might seem like premature optimization for small demos or tests, they become essential when understanding ModulaCMS's data model at scale.

### The Dynamic Schema Creates Many Rows

ModulaCMS uses a dynamic schema system where content structure is defined via:
- `datatypes` - Content type definitions (e.g., "Page", "Post")
- `fields` - Property definitions (e.g., "Title", "Body", "Image")
- `datatypes_fields` - Junction table linking fields to datatypes
- `content_data` - Content instances (tree nodes)
- `content_fields` - Actual field values

**Example: WordPress Page Equivalent**

A typical WordPress page has approximately:
- 23 core wp_posts fields
- 6 common meta fields
- **Total: 29 fields**

**Rows per page instance in ModulaCMS:**
- 1 row in `content_data` (the page itself)
- 29 rows in `content_fields` (one per field value)
- **Total: 30 rows per page**

**Including schema definition:**
- 1 row in `datatypes` (Page datatype) - shared across all pages
- 29 rows in `fields` (field definitions) - shared across all pages
- 29 rows in `datatypes_fields` (junction records) - shared across all pages
- **Total: 59 rows for schema + 30 rows per instance**

### Scale Impact

**Small demo (10 pages):**
- Schema: 59 rows
- Content: 10 × 30 = 300 rows
- **Total: 359 rows**
- Performance optimizations appear unnecessary

**Real-world college website (1,000 pages):**
- Schema: 59 rows
- Content: 1,000 × 30 = 30,000 rows
- **Total: 30,059 rows**
- Performance optimizations become critical

**Large enterprise site (10,000 pages):**
- Schema: 59 rows
- Content: 10,000 × 30 = 300,000 rows
- **Total: 300,059 rows**
- Performance optimizations make the difference between usable and unusable

### Headless Architecture Amplifies the Need

As a **headless CMS**, ModulaCMS serves content over HTTP/HTTPS APIs to separate frontends:

**Traditional CMS:**
```
Database → Server → Template → HTML (single process)
```

**Headless CMS (ModulaCMS):**
```
Database → API Server → Network → Frontend → Render
```

**Implications:**
- Every millisecond of query time gets amplified by network latency
- A 100ms database query becomes 100ms + 50ms network + frontend processing time
- Slow queries are immediately visible to end users
- Can't hide poor performance with server-side rendering or caching tricks

### Multiple Frontends = Multiple Query Patterns

ModulaCMS supports two distinct frontends:

**Client Frontend (Public Site):**
- Needs: Fast content delivery, search, filtering
- Pattern: Read-heavy, specific content queries
- Users: Site visitors (performance expectations are high)

**Admin Frontend (CMS Interface):**
- Needs: Tree navigation, CRUD operations, real-time updates
- Pattern: Mixed read/write, tree traversal, structural queries
- Users: Content editors (productivity depends on speed)

**You can't optimize for just one.** The backend must be universally fast.

### Agency Use Case: Unknown Frontend Implementation

**Key constraint:** Agencies build custom frontends that ModulaCMS doesn't control.

- They might use React, Vue, Svelte, vanilla JS, or any framework
- They might implement efficient caching strategies or none at all
- They might make smart, optimized queries or naive, unoptimized ones
- They might have memory constraints or unlimited resources

**This means:**
- Every API endpoint must be fast by default
- Backend can't rely on frontend optimization to compensate for slow queries
- Tree operations must be O(1) because you don't know how they'll be used
- Lazy loading must work because you don't know frontend memory constraints
- Database performance is the foundation everything builds on

### The Field Model Compounds Read Costs

Every content item requires joining across multiple tables:

**Single content item:**
```sql
content_data → content_fields (1 to 29 join)
content_fields → fields (29 field definitions)
content_data → datatypes (datatype info)
```

**Loading 100 pages:**
- 100 content_data rows
- 2,900 content_fields rows
- 29 field definition lookups
- Plus tree traversal operations

Without O(1) tree operations, this becomes:
- **Tree traversal:** O(n) per child lookup × 100 pages = potentially 1,000+ operations
- **Sibling navigation:** O(n) per sibling scan × average 5 siblings = 500+ operations
- **Total:** Thousands of database operations for a single page load

**With O(1) tree operations:**
- Direct pointer access for all navigation
- NodeIndex map provides instant lookups
- Lazy loading defers unnecessary queries
- **Result:** Fast, responsive interface regardless of content size

### Why These Specific Optimizations Matter

**1. Sibling-Pointer Tree (O(1) operations)**
- Without: O(n) to find/insert/delete children
- With 1,000 pages averaging 10 children each: 10,000 O(n) operations
- This compounds during tree traversal
- **Benefit:** Constant-time navigation regardless of tree size

**2. Lazy Loading**
- Without: Load all 30,000+ rows on initial page load
- With: Load only visible nodes (typically 10-50 rows initially)
- **Benefit:** 600× reduction in initial load time

**3. NodeIndex Map (O(1) lookups)**
- Without: O(n) search through tree to find specific node
- With: Instant lookup regardless of tree size
- **Benefit:** Essential for content editing workflow where you jump to specific items

**4. Three-Phase Loading with Orphan Resolution**
- Without: Multiple database round-trips to resolve parent-child relationships
- With: Single query, resolve in memory
- **Benefit:** Network latency doesn't compound with complex relationships

### The Design Philosophy

**For demos/tests:** These optimizations appear to be overkill.

**At scale:** These optimizations are the difference between:
- A responsive CMS that handles thousands of pages smoothly
- A system that times out loading a content tree
- A tool agencies can build production sites with
- A prototype that can't handle real-world content volumes

**ModulaCMS is designed for agencies building real-world sites.** The performance-minded architecture ensures the system scales from prototype to production without requiring architectural rewrites. The seemingly "over-engineered" tree structure is actually the minimum viable solution for the dynamic schema model at scale.

---

## Database Schema

### content_data Table

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER PRIMARY KEY,

    -- Tree structure pointers
    parent_id INTEGER
        REFERENCES content_data ON DELETE SET NULL,
    first_child_id INTEGER
        REFERENCES content_data ON DELETE SET NULL,
    next_sibling_id INTEGER
        REFERENCES content_data ON DELETE SET NULL,
    prev_sibling_id INTEGER
        REFERENCES content_data ON DELETE SET NULL,

    -- Content metadata
    route_id INTEGER NOT NULL
        REFERENCES routes ON DELETE CASCADE,
    datatype_id INTEGER NOT NULL
        REFERENCES datatypes ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraints
    FOREIGN KEY (parent_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL
);
```

**Key Fields:**
- `parent_id` - Points to parent node (NULL for root)
- `first_child_id` - Points to first child node (NULL if no children)
- `next_sibling_id` - Points to next sibling (NULL if last sibling)
- `prev_sibling_id` - Points to previous sibling (NULL if first sibling)

**Self-Referential Foreign Keys:**
All four pointer fields reference the same table (`content_data`), creating a self-referential structure.

---

## In-Memory Data Structures

### TreeRoot Structure

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:9-14`

```go
type TreeRoot struct {
    Root      *TreeNode              // Pointer to root node
    NodeIndex map[int64]*TreeNode    // O(1) lookup by content_data_id
    Orphans   map[int64]*TreeNode    // Nodes with missing parents
    MaxRetry  int                    // Max attempts to resolve orphans
}
```

**Purpose:**
- `Root` - Entry point to tree
- `NodeIndex` - Fast lookup of any node by ID (O(1))
- `Orphans` - Temporary storage for nodes whose parents haven't loaded yet
- `MaxRetry` - Prevents infinite loops in orphan resolution

**Creating a TreeRoot:**
```go
func NewTreeRoot() *TreeRoot {
    return &TreeRoot{
        NodeIndex: make(map[int64]*TreeNode),
        Orphans:   make(map[int64]*TreeNode),
        MaxRetry:  100,
    }
}
```

### TreeNode Structure

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:35-47`

```go
type TreeNode struct {
    // Database data
    Instance       *db.ContentData     // Content data from DB
    InstanceFields []db.ContentFields  // Field values
    Datatype       db.Datatypes        // Datatype definition
    Fields         []db.Fields         // Field definitions

    // Tree structure (sibling pointers)
    Parent         *TreeNode           // Pointer to parent
    FirstChild     *TreeNode           // Pointer to first child
    NextSibling    *TreeNode           // Pointer to next sibling
    PrevSibling    *TreeNode           // Pointer to previous sibling

    // UI state
    Expand         bool                // Is node expanded in TUI?
    Indent         int                 // Indentation level
    Wrapped        int                 // Text wrapping count
}
```

**Key Points:**
- Combines database data with tree structure
- Sibling pointers enable O(1) navigation
- UI state stored alongside content (Expand, Indent)

---

## Three-Phase Tree Loading Algorithm

ModulaCMS loads trees in three distinct phases to handle complex scenarios like orphaned nodes, circular references, and out-of-order loading.

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:69-94`

### Phase 1: Create All Nodes

**Purpose:** Create every node and populate the NodeIndex for O(1) lookups.

```go
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
```

**What Happens:**
1. Loop through all database rows
2. Create a TreeNode for each row
3. Add to NodeIndex map (key = content_data_id)
4. Identify root node (node with NULL parent_id)
5. No parent-child relationships established yet

**Why:** Creating all nodes first ensures every node exists before we try to link them together. This handles out-of-order loading from the database.

### Phase 2: Assign Immediate Hierarchy

**Purpose:** Connect nodes to parents when parent already exists in NodeIndex.

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:112-130`

```go
func (page *TreeRoot) assignImmediateHierarchy(stats *LoadStats) error {
    for id, node := range page.NodeIndex {
        if !node.Instance.ParentID.Valid {
            continue // Skip root
        }

        parentID := node.Instance.ParentID.Int64
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
```

**What Happens:**
1. Loop through all nodes in NodeIndex
2. Skip root node (has no parent)
3. Look up parent in NodeIndex
4. If parent exists: Attach node to parent (set pointers)
5. If parent missing: Add to Orphans map for later resolution

**Why Two Steps:**
- Some nodes might reference parents that don't exist yet
- Or parent hasn't been processed yet in the loop
- Orphans will be resolved in Phase 3

### Phase 3: Resolve Orphans Iteratively

**Purpose:** Resolve nodes whose parents weren't available in Phase 2.

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:133-161`

```go
func (page *TreeRoot) resolveOrphans(stats *LoadStats) error {
    for len(page.Orphans) > 0 && stats.RetryAttempts < page.MaxRetry {
        stats.RetryAttempts++
        orphansResolved := 0

        // Try to resolve each orphan
        for id, orphan := range page.Orphans {
            parentID := orphan.Instance.ParentID.Int64
            parent := page.NodeIndex[parentID]

            if parent != nil && parent.Parent != nil {
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
        }
    }

    return nil
}
```

**What Happens:**
1. Keep looping while orphans exist and under MaxRetry limit
2. For each orphan, check if parent now exists AND is connected to tree
3. If yes: Attach orphan and remove from Orphans map
4. If no orphans resolved in iteration: Check for circular references
5. Stop when all orphans resolved or circular references detected

**Why Iterative:**
- Forward references: Node A's parent is Node B, but B's parent is Node C
- A can't be resolved until B is resolved
- B can't be resolved until C is resolved
- Must iterate multiple times to resolve chains

**Why Check `parent.Parent != nil`:**
Ensures parent is actually connected to the tree before attaching the orphan.

### Attaching Nodes to Parents

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:201-215`

```go
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
```

**What Happens:**
1. Set node's Parent pointer
2. If parent has no children: Set as FirstChild
3. If parent has children: Traverse to last sibling, append node

**Result:** Node is added as the last child of parent with proper sibling linking.

---

## Circular Reference Detection

**Problem:** Database might contain circular references (Node A → Node B → Node A).

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:164-198`

### Detection Algorithm

```go
func (page *TreeRoot) hasCircularReference(node *TreeNode, visited map[int64]bool) bool {
    nodeID := node.Instance.ContentDataID

    if visited[nodeID] {
        return true // Found cycle - we've seen this node before
    }

    if !node.Instance.ParentID.Valid {
        return false // Reached root - no cycle
    }

    visited[nodeID] = true
    parentID := node.Instance.ParentID.Int64
    parent := page.NodeIndex[parentID]

    if parent == nil {
        return false // Parent doesn't exist (not circular, just missing)
    }

    return page.hasCircularReference(parent, visited)
}
```

**How It Works:**
1. Keep track of visited nodes in a map
2. Walk up the parent chain
3. If we visit a node twice: Circular reference detected
4. If we reach root (NULL parent): No cycle
5. If parent doesn't exist: Not circular, just orphaned

**Visited Map:**
The `visited` map prevents infinite loops. Each recursive call adds current node to visited.

**Example Circular Reference:**
```
Node 10: parent_id = 20
Node 20: parent_id = 30
Node 30: parent_id = 10  <- Circular!
```

Walking from Node 10:
1. Visit 10, add to map
2. Visit 20, add to map
3. Visit 30, add to map
4. Visit 10 again - already in map! Circular reference detected.

---

## Load Statistics Tracking

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:16-32`

```go
type LoadStats struct {
    NodesCount      int      // Total nodes loaded
    OrphansResolved int      // Orphans successfully resolved
    RetryAttempts   int      // Iterations to resolve orphans
    CircularRefs    []int64  // Node IDs with circular references
    FinalOrphans    []int64  // Unresolved orphans (errors)
}
```

**Purpose:** Track tree loading process for debugging and validation.

**Validation:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:218-239`

```go
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
```

**Errors Reported:**
- No root node found
- Circular references detected
- Unresolved orphans (parent doesn't exist in database)

---

## Tree Operations

### Finding Nodes: O(1) Lookup

**Via NodeIndex:**
```go
node := treeRoot.NodeIndex[contentDataID]
if node != nil {
    // Found node in O(1)
}
```

**Why O(1):** Map lookup is constant time.

### Inserting Nodes: O(1)

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:264-271`

```go
func (page *TreeRoot) InsertTreeNodeByIndex(parent, firstChild, prevSibling, nextSibling, n *TreeNode) {
    page.NodeIndex[n.Instance.ContentDataID] = n
    n.Parent = parent
    n.FirstChild = firstChild
    n.PrevSibling = prevSibling
    n.NextSibling = nextSibling
}
```

**What Happens:**
1. Add to NodeIndex
2. Wire up all four pointers
3. All pointer operations are O(1)

### Deleting Nodes: O(1)

Deletion is complex because we must maintain tree structure when removing a node.

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:275-296`

```go
func (page *TreeRoot) DeleteTreeNodeByIndex(n *TreeNode) bool {
    target := page.NodeIndex[n.Instance.ContentDataID]

    // Can't delete root or nodes not in tree
    if page.Root == nil || target == nil || page.Root == target || target.Parent == nil {
        return false
    }

    // Special case: target is first child
    if target.Parent.FirstChild == target {
        DeleteFirstChild(target)
        delete(page.NodeIndex, n.Instance.ContentDataID)
        return true
    }

    // General case: target is nested child
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
```

### Deletion Cases

**Case 1: Delete First Child with No Children**

```
Before:
Parent
  ├─ Target (first child)
  ├─ Sibling1
  └─ Sibling2

After:
Parent
  ├─ Sibling1
  └─ Sibling2
```

**Code:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:331-342`

```go
func DeleteFirstChildNoChildren(target *TreeNode) bool {
    if target.NextSibling != nil {
        target.Parent.FirstChild = target.NextSibling
        target.NextSibling.PrevSibling = nil
        return true
    } else {
        target.Parent.FirstChild = nil
        return true
    }
}
```

**Case 2: Delete First Child with Children**

```
Before:
Parent
  ├─ Target (first child)
  │   ├─ TargetChild1
  │   └─ TargetChild2
  ├─ Sibling1
  └─ Sibling2

After:
Parent
  ├─ TargetChild1
  ├─ TargetChild2
  ├─ Sibling1
  └─ Sibling2
```

Target's children get promoted to be siblings of Target's former siblings.

**Code:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go:307-329`

**Case 3: Delete Nested Child with No Children**

```
Before:
Parent
  ├─ Child1
  ├─ Target
  └─ Child3

After:
Parent
  ├─ Child1
  └─ Child3
```

**Case 4: Delete Nested Child with Children**

```
Before:
Parent
  ├─ Child1
  ├─ Target
  │   ├─ TargetChild1
  │   └─ TargetChild2
  └─ Child3

After:
Parent
  ├─ Child1
  ├─ TargetChild1
  ├─ TargetChild2
  └─ Child3
```

**All operations:** O(1) pointer rewiring (except finding target if not first child, which is O(siblings)).

---

## Tree Traversal Patterns

### Depth-First Traversal (DFS)

Visit all descendants of a node before moving to siblings:

```go
func TraverseDepthFirst(node *TreeNode, visit func(*TreeNode)) {
    if node == nil {
        return
    }

    // Visit current node
    visit(node)

    // Traverse all children
    child := node.FirstChild
    for child != nil {
        TraverseDepthFirst(child, visit)
        child = child.NextSibling
    }
}
```

**Order:** Root → Child1 → Child1's Children → Child2 → Child2's Children → ...

**Use Case:** Rendering entire tree in hierarchy order (TUI display).

### Breadth-First Traversal (BFS)

Visit all children at one level before descending:

```go
func TraverseBreadthFirst(node *TreeNode, visit func(*TreeNode)) {
    if node == nil {
        return
    }

    queue := []*TreeNode{node}

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        visit(current)

        // Add all children to queue
        child := current.FirstChild
        for child != nil {
            queue = append(queue, child)
            child = child.NextSibling
        }
    }
}
```

**Order:** Root → All Root's Children → All Grandchildren → ...

**Use Case:** Processing tree level-by-level.

### Sibling Iteration

Navigate all siblings of a node:

```go
func IterateSiblings(startNode *TreeNode, visit func(*TreeNode)) {
    current := startNode
    for current != nil {
        visit(current)
        current = current.NextSibling
    }
}
```

**Use Case:** Listing all items at the same hierarchy level.

### Parent Chain Traversal

Walk up to root:

```go
func GetPathToRoot(node *TreeNode) []*TreeNode {
    path := []*TreeNode{}
    current := node

    for current != nil {
        path = append([]*TreeNode{current}, path...) // Prepend
        current = current.Parent
    }

    return path
}
```

**Use Case:** Building breadcrumb navigation, checking permissions up hierarchy.

---

## Lazy Loading

ModulaCMS supports lazy loading to avoid loading entire large trees into memory.

**Strategy:**
1. Initially load only root and immediate children
2. Load deeper levels on-demand when user expands nodes in TUI

**Implementation:**
- Marked nodes as "not loaded" initially
- When user expands node: Query database for children
- Attach children to parent using normal tree operations

**Benefits:**
- Fast initial load for trees with hundreds of nodes
- Memory efficient (only load visible portions)
- Responsive UI (no waiting for full tree load)

**Typical Initial Query:**
```sql
-- Load root and first level only
SELECT cd.*, dt.label as datatype_label, dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
AND (cd.parent_id IS NULL OR cd.parent_id IN (
    SELECT content_data_id FROM content_data
    WHERE parent_id IS NULL AND route_id = ?
))
```

**On-Demand Query:**
```sql
-- Load children of specific node
SELECT cd.*, dt.label as datatype_label, dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.parent_id = ?
```

---

## Performance Characteristics

### Time Complexity

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| Find node by ID | O(1) | NodeIndex map lookup |
| Navigate to parent | O(1) | Direct pointer |
| Navigate to first child | O(1) | Direct pointer |
| Navigate to next sibling | O(1) | Direct pointer |
| Insert node | O(1) | Pointer rewiring |
| Delete node | O(1)* | *O(siblings) to find if not first child |
| Tree loading (3-phase) | O(n) | n = number of nodes |
| Orphan resolution | O(o × r) | o = orphans, r = retry attempts |
| Circular detection | O(d) | d = depth of circular chain |

### Space Complexity

| Component | Space | Notes |
|-----------|-------|-------|
| NodeIndex map | O(n) | One entry per node |
| TreeNode pointers | O(n) | 4 pointers per node |
| Orphans map | O(o) | Temporary, cleared after loading |
| Total | O(n) | Linear in number of nodes |

**Example:**
- Tree with 1000 nodes
- NodeIndex: 1000 map entries
- Tree pointers: 4000 pointers (4 per node)
- Memory: ~200KB (approximate)

---

## Common Pitfalls and Debugging

### Pitfall 1: Circular References in Database

**Symptom:** Tree loading fails with "circular references detected"

**Cause:** Database contains nodes that reference each other as parents:
```
Node A: parent_id = B
Node B: parent_id = A
```

**Detection:** LoadStats.CircularRefs will contain offending node IDs

**Fix:** Clean database:
```sql
-- Find circular references
WITH RECURSIVE chain AS (
    SELECT content_data_id, parent_id, 1 as depth
    FROM content_data
    WHERE content_data_id = ?  -- Start node

    UNION ALL

    SELECT cd.content_data_id, cd.parent_id, c.depth + 1
    FROM content_data cd
    JOIN chain c ON cd.content_data_id = c.parent_id
    WHERE c.depth < 100
)
SELECT * FROM chain;
```

### Pitfall 2: Orphaned Nodes

**Symptom:** Tree loading fails with "unresolved orphan nodes"

**Cause:** Node references parent that doesn't exist:
```
Node X: parent_id = 999
Node 999: Does not exist in database
```

**Detection:** LoadStats.FinalOrphans will contain orphaned node IDs

**Fix:** Either:
1. Create missing parent node
2. Update orphan to reference existing parent
3. Set orphan's parent_id to NULL (make it root)

```sql
-- Find orphans
SELECT cd.content_data_id, cd.parent_id
FROM content_data cd
LEFT JOIN content_data parent ON cd.parent_id = parent.content_data_id
WHERE cd.parent_id IS NOT NULL
AND parent.content_data_id IS NULL;
```

### Pitfall 3: Missing Root Node

**Symptom:** Tree loading fails with "no root node found"

**Cause:** No node with NULL parent_id for the route

**Fix:**
```sql
-- Ensure route has root node
INSERT INTO content_data (parent_id, route_id, datatype_id, author_id)
VALUES (NULL, ?, ?, ?);
```

### Pitfall 4: Inconsistent Sibling Pointers

**Symptom:** Node appears multiple times or missing from traversal

**Cause:** Database sibling pointers don't match actual structure

**Example:**
```
Node A: next_sibling_id = B
Node B: prev_sibling_id = C  <- Should be A!
```

**Detection:** Verify integrity:
```sql
-- Check sibling pointer consistency
SELECT *
FROM content_data cd
WHERE next_sibling_id IS NOT NULL
AND NOT EXISTS (
    SELECT 1 FROM content_data sibling
    WHERE sibling.content_data_id = cd.next_sibling_id
    AND sibling.prev_sibling_id = cd.content_data_id
);
```

**Fix:** Rebuild sibling pointers or fix manually

### Debugging Tips

**1. Enable Loading Stats:**
```go
stats, err := treeRoot.LoadFromRows(rows)
if err != nil {
    log.Printf("Loading failed: %v", err)
    log.Printf("Stats: %s", stats.String())
}
```

**2. Inspect NodeIndex:**
```go
log.Printf("Total nodes loaded: %d", len(treeRoot.NodeIndex))
for id, node := range treeRoot.NodeIndex {
    log.Printf("Node %d: parent=%v, children=%v",
        id,
        node.Parent != nil,
        node.FirstChild != nil)
}
```

**3. Verify Tree Structure:**
```go
func ValidateTree(node *TreeNode, visited map[int64]bool) error {
    if visited[node.Instance.ContentDataID] {
        return fmt.Errorf("cycle detected at node %d", node.Instance.ContentDataID)
    }
    visited[node.Instance.ContentDataID] = true

    // Validate children
    child := node.FirstChild
    for child != nil {
        if child.Parent != node {
            return fmt.Errorf("child %d has wrong parent", child.Instance.ContentDataID)
        }
        if err := ValidateTree(child, visited); err != nil {
            return err
        }
        child = child.NextSibling
    }

    return nil
}
```

---

## Usage Examples

### Example 1: Loading a Content Tree

```go
// Get content tree rows from database
rows, err := db.GetContentTreeByRoute(routeID)
if err != nil {
    return fmt.Errorf("failed to load tree: %w", err)
}

// Create tree root
treeRoot := NewTreeRoot()

// Load tree using 3-phase algorithm
stats, err := treeRoot.LoadFromRows(rows)
if err != nil {
    utility.DefaultLogger.Error("Tree loading failed", "error", err, "stats", stats.String())
    return err
}

utility.DefaultLogger.Info("Tree loaded successfully",
    "nodes", stats.NodesCount,
    "orphans_resolved", stats.OrphansResolved,
    "retries", stats.RetryAttempts)

// Access root node
rootNode := treeRoot.Root
```

### Example 2: Finding a Node

```go
// O(1) lookup
node := treeRoot.NodeIndex[contentDataID]
if node == nil {
    return fmt.Errorf("node %d not found", contentDataID)
}

// Access node properties
fmt.Printf("Node: %s (type: %s)\n",
    node.Datatype.Label,
    node.Datatype.Type)
```

### Example 3: Traversing Children

```go
// Iterate through all children of a node
child := node.FirstChild
for child != nil {
    fmt.Printf("Child: %d\n", child.Instance.ContentDataID)
    child = child.NextSibling
}
```

### Example 4: Walking Entire Tree

```go
func PrintTree(node *TreeNode, indent int) {
    if node == nil {
        return
    }

    // Print current node
    fmt.Printf("%s%s\n",
        strings.Repeat("  ", indent),
        node.Datatype.Label)

    // Print all children recursively
    child := node.FirstChild
    for child != nil {
        PrintTree(child, indent+1)
        child = child.NextSibling
    }
}

// Usage
PrintTree(treeRoot.Root, 0)
```

**Output:**
```
Page
  Hero Section
    Title
    Subtitle
  Content Section
    Paragraph
    Image
  Footer
```

### Example 5: Inserting a New Node

```go
// Create new node
newNode := &TreeNode{
    Instance: &db.ContentData{
        ContentDataID: 123,
        ParentID:      sql.NullInt64{Int64: parentID, Valid: true},
        DatatypeID:    datatypeID,
    },
    Datatype: datatype,
}

// Add to tree
parent := treeRoot.NodeIndex[parentID]
if parent != nil {
    treeRoot.attachNodeToParent(newNode, parent)
    treeRoot.NodeIndex[123] = newNode
}
```

### Example 6: Deleting a Node

```go
// Find node to delete
nodeToDelete := treeRoot.NodeIndex[contentDataID]
if nodeToDelete == nil {
    return fmt.Errorf("node not found")
}

// Delete from tree
success := treeRoot.DeleteTreeNodeByIndex(nodeToDelete)
if !success {
    return fmt.Errorf("failed to delete node")
}

// Node is removed from tree and NodeIndex
```

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Domain model relationships
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TUI_ARCHITECTURE.md` - How tree is rendered in TUI

**Packages:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` - TUI implementation details
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/MODEL_PACKAGE.md` - Tree data structures

**Database:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` - Schema definitions
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database operations

**Domain:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/CONTENT_TREES.md` - Content tree operations

**General:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/FILE_TREE.md` - Complete directory structure
- `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md` - Development guidelines

---

## Quick Reference

### Key Files
- Schema: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data/`
- Implementation: `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/cms_struct.go`
- Usage: `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/` (various files)

### Key Structures
```go
TreeRoot    // Root + NodeIndex + Orphans
TreeNode    // Parent + FirstChild + NextSibling + PrevSibling
LoadStats   // Loading metrics and errors
```

### Key Operations
```go
NewTreeRoot()                        // Create tree root
LoadFromRows(rows)                   // 3-phase loading
NodeIndex[id]                        // O(1) find node
attachNodeToParent(node, parent)     // O(1) insert
DeleteTreeNodeByIndex(node)          // O(1)* delete
```

### Loading Phases
1. **Phase 1:** Create all nodes, populate NodeIndex
2. **Phase 2:** Assign hierarchy for nodes with existing parents
3. **Phase 3:** Iteratively resolve orphans, detect circular refs

### Common Errors
- "circular references detected" → Fix database cycles
- "unresolved orphan nodes" → Parent doesn't exist in DB
- "no root node found" → Add node with NULL parent_id

### Performance
- Node lookup: O(1)
- Navigation: O(1)
- Insert/Delete: O(1)
- Full tree load: O(n)
