# Lazy Tree Loading Pattern for Content Management

## Core Concept

Use ROOT datatypes as content creation templates, then lazy-load tree hierarchy as users navigate deeper.

## ROOT Types as Creation Templates

### From Database Analysis
```sql
-- ROOT datatypes serve as top-level content types
SELECT * FROM datatypes WHERE type = 'ROOT';
-- Results: Page (id=1), Post (id=2), Forms (id=3)
```

### Creation Flow
```go
type ContentCreationFlow struct {
    AvailableRoots []Datatype  // Page, Post, Forms
    SelectedRoot   *Datatype   // User's choice
    FormFields     []Field     // Fields for selected root type
}

// When user clicks "Create Content"
func ShowContentCreationDialog() ContentCreationFlow {
    roots := database.GetRootDatatypes() // WHERE type = 'ROOT'
    return ContentCreationFlow{
        AvailableRoots: roots,
        SelectedRoot:   nil,
        FormFields:     nil,
    }
}

// When user selects a root type (e.g., "Page")
func SelectRootType(rootID int64) ContentCreationFlow {
    root := database.GetDatatype(rootID)
    fields := database.GetFieldsForDatatype(rootID)
    
    return ContentCreationFlow{
        SelectedRoot: root,
        FormFields:   fields, // Title, Favicon, Description for Page
    }
}
```

## Shallow Loading Strategy

### Initial Tree Load
```go
// Load only immediate children of root content
func LoadContentTreeShallow(routeID int64) *ContentTree {
    query := `
    SELECT cd.*, dt.label as datatype_label, dt.type as datatype_type
    FROM content_data cd
    JOIN datatypes dt ON cd.datatype_id = dt.datatype_id  
    WHERE cd.route_id = ? 
    AND (cd.parent_id IS NULL OR cd.parent_id IN (
        SELECT content_data_id FROM content_data 
        WHERE parent_id IS NULL AND route_id = ?
    ))
    ORDER BY cd.parent_id NULLS FIRST, cd.content_data_id`
    
    // This gives us:
    // - Root content (parent_id = NULL) 
    // - First level children only
    // - No deeper descendants
}

// Example result for a Page:
// content_data_id=1, parent_id=NULL, datatype="Page"     <- ROOT
// content_data_id=4, parent_id=1,    datatype="Hero"     <- Level 1
// content_data_id=5, parent_id=1,    datatype="Row"      <- Level 1
// (Row's children not loaded yet)
```

### ELM Model for Shallow Loading
```go
type Model struct {
    ContentTree     *ContentTree
    LoadedDepths    map[int64]int      // nodeID -> deepest loaded level
    LazyLoadStates  map[int64]bool     // nodeID -> is loading children
    ExpandableNodes map[int64]bool     // nodeID -> has unloaded children
}

type ContentNode struct {
    ContentDataID   int64
    ParentID        *int64
    DatatypeLabel   string
    DatatypeType    string
    Children        []*ContentNode
    HasMoreChildren bool               // indicates lazy loading needed
    LoadedDepth     int                // how deep we've loaded from this node
}
```

## Navigation-Triggered Loading

### User Expands Node
```go
// Message when user clicks expand arrow
type ExpandNodeMsg struct {
    NodeID int64
    LoadChildren bool  // true if children not yet loaded
}

func Update(msg Msg, model Model) (Model, Cmd) {
    switch m := msg.(type) {
    case ExpandNodeMsg:
        // Mark as expanded in UI
        model.ContentTree.ExpandedNodes[m.NodeID] = true
        
        // Check if we need to load children
        if m.LoadChildren && !model.LazyLoadStates[m.NodeID] {
            model.LazyLoadStates[m.NodeID] = true
            
            return model, LoadNodeChildrenCmd{
                NodeID: m.NodeID,
                Depth:  1, // load one level deeper
            }
        }
        
        return model, NoCmd()
        
    case NodeChildrenLoadedMsg:
        delete(model.LazyLoadStates, m.NodeID)
        
        // Add children to parent node
        if parent, exists := model.ContentTree.Nodes[m.NodeID]; exists {
            parent.Children = m.Children
            parent.HasMoreChildren = m.HasMoreChildren
            
            // Add children to lookup map
            for _, child := range m.Children {
                model.ContentTree.Nodes[child.ContentDataID] = child
            }
        }
        
        return model, NoCmd()
    }
}
```

### Smart Loading Query
```go
type LoadNodeChildrenCmd struct {
    NodeID int64
    Depth  int    // how many levels to load
}

func (c LoadNodeChildrenCmd) Execute() Msg {
    // Load children and check if they have children
    query := `
    WITH RECURSIVE children AS (
        -- Direct children
        SELECT cd.*, dt.label, dt.type, 1 as level,
               EXISTS(SELECT 1 FROM content_data WHERE parent_id = cd.content_data_id) as has_children
        FROM content_data cd
        JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
        WHERE cd.parent_id = ?
        
        UNION ALL
        
        -- Recursive children (up to specified depth)
        SELECT cd.*, dt.label, dt.type, c.level + 1,
               EXISTS(SELECT 1 FROM content_data WHERE parent_id = cd.content_data_id) as has_children
        FROM content_data cd
        JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
        JOIN children c ON cd.parent_id = c.content_data_id
        WHERE c.level < ?
    )
    SELECT * FROM children ORDER BY level, content_data_id`
    
    children, err := database.Query(query, c.NodeID, c.Depth)
    
    return NodeChildrenLoadedMsg{
        NodeID:   c.NodeID,
        Children: children,
        Err:      err,
    }
}
```

## UI Indicators for Lazy Loading

### Tree Node Rendering
```go
func renderTreeNode(node *ContentNode, model Model, depth int) HTML {
    isLoading := model.LazyLoadStates[node.ContentDataID]
    isExpanded := model.ContentTree.ExpandedNodes[node.ContentDataID]
    
    // Determine expand/collapse icon
    var expandIcon HTML
    if isLoading {
        expandIcon = spinner()
    } else if node.HasMoreChildren {
        if isExpanded {
            expandIcon = icon("chevron-down", onClick(CollapseNodeMsg{node.ContentDataID}))
        } else {
            expandIcon = icon("chevron-right", onClick(ExpandNodeMsg{
                NodeID: node.ContentDataID,
                LoadChildren: len(node.Children) == 0, // load if empty
            }))
        }
    } else if len(node.Children) > 0 {
        // Has loaded children, show expand/collapse
        if isExpanded {
            expandIcon = icon("chevron-down", onClick(CollapseNodeMsg{node.ContentDataID}))
        } else {
            expandIcon = icon("chevron-right", onClick(ExpandNodeMsg{
                NodeID: node.ContentDataID,
                LoadChildren: false, // don't load, just expand
            }))
        }
    } else {
        // Leaf node
        expandIcon = div([]Attribute{class("leaf-spacer")}, []HTML{})
    }
    
    return div(
        []Attribute{class("tree-node")},
        []HTML{
            expandIcon,
            text(fmt.Sprintf("%s (%s)", node.DatatypeLabel, node.DatatypeType)),
        },
    )
}
```

## Performance Benefits

### Database Efficiency
```go
// Instead of: Loading entire tree (potentially hundreds of nodes)
SELECT * FROM content_data WHERE route_id = 1; -- 500+ rows

// We do: Load incrementally as needed
SELECT * FROM content_data WHERE route_id = 1 AND parent_id IS NULL;           -- 1 row
SELECT * FROM content_data WHERE parent_id = 1;                                -- 5 rows  
SELECT * FROM content_data WHERE parent_id IN (4, 5, 6, 7, 8);               -- 12 rows
// Total: 18 rows loaded vs 500+ rows
```

### Memory Efficiency
```go
type TreeMemoryFootprint struct {
    ShallowLoading int // Only loaded nodes in memory
    FullLoading    int // Entire tree in memory
}

// Example for a typical page:
// Shallow: ~20 nodes * 200 bytes = 4KB
// Full:    ~500 nodes * 200 bytes = 100KB
// 25x reduction in memory usage
```

### User Experience Benefits
```go
// Immediate feedback
// - Root content types show instantly
// - Tree expands quickly (only loading needed children)
// - No waiting for massive tree to load
// - Progressive disclosure matches user mental model

// Bandwidth efficiency  
// - Mobile users don't download unused tree data
// - Faster initial page loads
// - Network requests only when user explores deeper
```

## Content Creation Integration

### Root-Type Selection
```go
func ShowCreateContentDialog(model Model) HTML {
    rootTypes := []Datatype{
        {ID: 1, Label: "Page", Type: "ROOT"},
        {ID: 2, Label: "Post", Type: "ROOT"}, 
        {ID: 3, Label: "Forms", Type: "ROOT"},
    }
    
    options := make([]HTML, len(rootTypes))
    for i, rootType := range rootTypes {
        options[i] = button(
            []Attribute{
                onClick(SelectContentTypeMsg{rootType.ID}),
                class("content-type-option"),
            },
            []HTML{
                icon(rootType.Type),
                text(rootType.Label),
                text(fmt.Sprintf("Create new %s", strings.ToLower(rootType.Label))),
            },
        )
    }
    
    return dialog(
        []Attribute{class("create-content-dialog")},
        []HTML{
            h2([]HTML{text("What would you like to create?")}),
            div([]Attribute{class("type-grid")}, options),
        },
    )
}
```

This pattern provides:
1. **Fast initial loads** - only essential tree structure
2. **Responsive navigation** - load content as users explore  
3. **Efficient memory usage** - only loaded nodes in memory
4. **Clear content types** - ROOT types as creation templates
5. **Progressive disclosure** - complexity revealed as needed