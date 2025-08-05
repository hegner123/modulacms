# Tree Query Patterns for Content Hierarchy

## Database Tree Structure

The `content_data` table forms a tree using `parent_id` self-references:
- Root nodes: `parent_id = NULL`
- Child nodes: `parent_id` points to parent's `content_data_id`

## Query Approaches

### 1. Recursive CTE Approach (SQLite 3.8.3+)
```sql
-- Get entire tree starting from root
WITH RECURSIVE content_tree AS (
    -- Base case: root nodes
    SELECT content_data_id, parent_id, route_id, datatype_id, 0 as depth, 
           CAST(content_data_id AS TEXT) as path
    FROM content_data 
    WHERE parent_id IS NULL
    
    UNION ALL
    
    -- Recursive case: children
    SELECT c.content_data_id, c.parent_id, c.route_id, c.datatype_id, 
           ct.depth + 1, ct.path || '/' || c.content_data_id
    FROM content_data c
    JOIN content_tree ct ON c.parent_id = ct.content_data_id
)
SELECT * FROM content_tree ORDER BY path;
```

### 2. Go Implementation Considerations

#### Data Structures
```go
type ContentNode struct {
    ContentDataID int64
    ParentID      *int64  // pointer for NULL handling
    RouteID       int64
    DatatypeID    int64
    Depth         int
    Path          string
    Children      []*ContentNode  // slice of pointers
    Fields        []ContentField
}

type ContentTree struct {
    Nodes    map[int64]*ContentNode  // fast lookup by ID
    Roots    []*ContentNode          // top-level nodes
    MaxDepth int
}
```

#### Go Map Limitations & Solutions
```go
// Problem: Maps are not ordered
// Solution: Use additional slice for ordered traversal
type OrderedTree struct {
    NodeMap   map[int64]*ContentNode  // O(1) lookup
    NodeOrder []int64                 // preserve insertion/sort order
    RootIDs   []int64                 // root node IDs in order
}

// Problem: Maps don't preserve hierarchy
// Solution: Build parent-child relationships separately
func BuildTreeFromFlat(flatNodes []ContentNode) *ContentTree {
    tree := &ContentTree{
        Nodes: make(map[int64]*ContentNode),
        Roots: make([]*ContentNode, 0),
    }
    
    // First pass: create all nodes in map
    for i := range flatNodes {
        node := &flatNodes[i]
        node.Children = make([]*ContentNode, 0)
        tree.Nodes[node.ContentDataID] = node
    }
    
    // Second pass: build parent-child relationships
    for _, node := range tree.Nodes {
        if node.ParentID == nil {
            tree.Roots = append(tree.Roots, node)
        } else {
            if parent, exists := tree.Nodes[*node.ParentID]; exists {
                parent.Children = append(parent.Children, node)
            }
        }
    }
    
    return tree
}
```

### 3. Query Strategies by Use Case

#### A. Load Specific Subtree
```go
// Get node and all descendants
func LoadSubtree(db *sql.DB, rootID int64) (*ContentNode, error) {
    // Use CTE to get all descendants
    query := `
    WITH RECURSIVE subtree AS (
        SELECT content_data_id, parent_id, route_id, datatype_id, 0 as depth
        FROM content_data WHERE content_data_id = ?
        UNION ALL
        SELECT c.content_data_id, c.parent_id, c.route_id, c.datatype_id, s.depth + 1
        FROM content_data c
        JOIN subtree s ON c.parent_id = s.content_data_id
    )
    SELECT * FROM subtree ORDER BY depth, content_data_id`
    
    // Build tree from flat result
    // Return root node with populated Children slices
}
```

#### B. Breadth-First Traversal
```go
func TraverseBreadthFirst(root *ContentNode, visitor func(*ContentNode)) {
    if root == nil {
        return
    }
    
    queue := []*ContentNode{root}  // slice as queue
    
    for len(queue) > 0 {
        current := queue[0]        // dequeue from front
        queue = queue[1:]          // slice limitation: O(n) operation
        
        visitor(current)
        
        // Add children to queue
        queue = append(queue, current.Children...)
    }
}

// Optimization: Use ring buffer or proper queue for large trees
```

#### C. Depth-First Traversal
```go
func TraverseDepthFirst(node *ContentNode, visitor func(*ContentNode, int), depth int) {
    if node == nil {
        return
    }
    
    visitor(node, depth)
    
    // Traverse children
    for _, child := range node.Children {
        TraverseDepthFirst(child, visitor, depth+1)
    }
}
```

### 4. Loading Fields with Tree

#### Join Query Limitation
```go
// Problem: JOIN with content_fields creates cartesian product
// One content_data row becomes N rows (one per field)

// Solution 1: Separate queries
func LoadTreeWithFields(db *sql.DB, rootID int64) (*ContentNode, error) {
    // First: Load tree structure
    tree := LoadSubtree(db, rootID)
    
    // Second: Load all fields for tree nodes
    nodeIDs := ExtractNodeIDs(tree)
    fieldsMap := LoadFieldsForNodes(db, nodeIDs)
    
    // Third: Attach fields to nodes
    AttachFieldsToTree(tree, fieldsMap)
    
    return tree
}

// Solution 2: JSON aggregation (SQLite 3.45.0+)
query := `
WITH RECURSIVE subtree AS (...),
fields_json AS (
    SELECT content_data_id, 
           json_group_array(json_object(
               'field_id', field_id,
               'field_value', field_value
           )) as fields
    FROM content_fields 
    WHERE content_data_id IN (SELECT content_data_id FROM subtree)
    GROUP BY content_data_id
)
SELECT s.*, COALESCE(f.fields, '[]') as fields_json
FROM subtree s
LEFT JOIN fields_json f ON s.content_data_id = f.content_data_id
`
```

### 5. Go Slice Performance Considerations

```go
// Problem: Appending to slices may cause reallocations
children := make([]*ContentNode, 0)          // starts with 0 capacity
children = append(children, child)           // may reallocate

// Solution: Pre-allocate when size is known
children := make([]*ContentNode, 0, expectedCount)

// Problem: Removing from middle of slice is O(n)
// Solution: Use tombstone pattern or rebuild slice
func RemoveChild(parent *ContentNode, childID int64) {
    filtered := make([]*ContentNode, 0, len(parent.Children))
    for _, child := range parent.Children {
        if child.ContentDataID != childID {
            filtered = append(filtered, child)
        }
    }
    parent.Children = filtered
}
```

### 6. Memory Efficiency Patterns

```go
// For large trees, consider streaming/pagination
type TreeCursor struct {
    CurrentPath []int64
    Depth       int
    BatchSize   int
}

func LoadTreeBatch(db *sql.DB, cursor *TreeCursor) ([]*ContentNode, *TreeCursor, error) {
    // Load next batch of nodes at current level
    // Return new cursor for next batch
    // Allows processing huge trees without loading everything
}

// For read-only trees, use sync.Pool for node reuse
var nodePool = sync.Pool{
    New: func() any {
        return &ContentNode{
            Children: make([]*ContentNode, 0, 4), // common case
        }
    },
}
```

## Key Takeaways

1. **Use maps for O(1) lookups**, slices for ordered traversal
2. **Two-pass construction** works well for tree building
3. **Separate field loading** avoids cartesian product issues
4. **Pre-allocate slices** when size is predictable
5. **Consider memory usage** for large trees (streaming/pooling)
6. **Recursive CTEs** handle arbitrary depth efficiently