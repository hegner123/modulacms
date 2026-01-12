# MODEL_PACKAGE.md

Comprehensive guide to the internal/model package for JSON-serializable content representation.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/`
**Purpose:** Provides JSON-serializable data structures for content tree representation, API responses, and content export
**Last Updated:** 2026-01-12

---

## Table of Contents

1. [Overview](#overview)
2. [Core Data Structures](#core-data-structures)
3. [Key Functions](#key-functions)
4. [JSON Serialization](#json-serialization)
5. [Tree Building](#tree-building)
6. [Tree Operations](#tree-operations)
7. [Relationship to CLI TreeRoot](#relationship-to-cli-treeroot)
8. [Usage Patterns](#usage-patterns)
9. [Common Workflows](#common-workflows)
10. [Testing](#testing)
11. [Related Documentation](#related-documentation)
12. [Quick Reference](#quick-reference)

---

## Overview

The **internal/model** package provides JSON-serializable data structures for representing ModulaCMS content trees. Unlike the CLI's sibling-pointer tree structure (TreeRoot/TreeNode documented in TREE_STRUCTURE.md), this package focuses on creating hierarchical representations suitable for JSON output, API responses, and content export.

**Key Characteristics:**
- **JSON-safe**: All structures can be marshaled to/from JSON
- **Hierarchical**: Parent-child relationships stored as nested slices
- **Self-contained**: Each node contains its datatype info, content data, and field values
- **Circular reference protection**: Custom marshaling prevents infinite loops
- **Database-agnostic**: Works with data from any DbDriver implementation

**When to Use:**
- Building JSON API responses for content
- Exporting content trees to JSON format
- Creating snapshots of content for backup/restore
- Rendering content for external systems (WordPress adapter, etc.)
- Testing content structure without full tree complexity

**When NOT to Use:**
- For TUI tree navigation (use CLI's TreeRoot instead)
- For lazy-loading large trees (use CLI's TreeRoot instead)
- For O(1) sibling operations (use CLI's TreeRoot instead)
- For in-memory content manipulation in the TUI

---

## Core Data Structures

### Root

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:10-12`

The Root struct represents the top-level container for a content tree.

```go
type Root struct {
	Node *Node `json:"root"`
}
```

**Fields:**
- `Node`: Pointer to the root node of the content tree (typically the route node)

**Usage:**
```go
root := model.NewRoot()
root.Node = &someNode
```

---

### Node

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:14-19`

The Node struct represents a single node in the content tree.

```go
type Node struct {
	Datatype Datatype `json:"datatype"`
	Fields   []Field  `json:"fields"`
	Nodes    []*Node  `json:"nodes"`
}
```

**Fields:**
- `Datatype`: The datatype information and content data for this node
- `Fields`: Slice of field values attached to this content node
- `Nodes`: Slice of child nodes (hierarchical structure)

**Key Characteristics:**
- **Self-contained**: Contains all information about the content node
- **Hierarchical**: Child nodes stored in `Nodes` slice
- **Fields included**: All field values are attached to the node
- **Recursive**: Can contain unlimited depth of child nodes

**Example Structure:**
```json
{
  "datatype": {
    "info": { "datatype_id": 1, "type": "page", "name": "Home Page" },
    "content": { "content_data_id": 100, "parent_id": null }
  },
  "fields": [
    { "info": { "field_id": 1, "name": "title" }, "content": { "value": "Welcome" } }
  ],
  "nodes": [
    { "datatype": { ... }, "fields": [ ... ], "nodes": [] }
  ]
}
```

---

### Datatype

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:21-24`

The Datatype struct wraps datatype schema information and content instance data.

```go
type Datatype struct {
	Info    db.DatatypeJSON    `json:"info"`
	Content db.ContentDataJSON `json:"content"`
}
```

**Fields:**
- `Info`: Datatype schema definition (from datatypes table) via `db.DatatypeJSON`
- `Content`: Content instance data (from content_data table) via `db.ContentDataJSON`

**Purpose:**
Combines the **schema** (what kind of content this is) with the **instance** (this specific piece of content) for a complete representation.

**Database Mapping:**
- `Info` maps to the `datatypes` table (columns: datatype_id, name, type, etc.)
- `Content` maps to the `content_data` table (columns: content_data_id, parent_id, route_id, etc.)

**See Also:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` for datatype/content relationships
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` for db.DatatypeJSON and db.ContentDataJSON types

---

### Field

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:26-29`

The Field struct wraps field schema definition and field value data.

```go
type Field struct {
	Info    db.FieldsJSON        `json:"info"`
	Content db.ContentFieldsJSON `json:"content"`
}
```

**Fields:**
- `Info`: Field definition (from fields table) via `db.FieldsJSON`
- `Content`: Field value data (from content_fields table) via `db.ContentFieldsJSON`

**Purpose:**
Combines the **field schema** (what this field is) with the **field value** (the actual content) for a complete representation.

**Database Mapping:**
- `Info` maps to the `fields` table (columns: field_id, name, type, default_value, etc.)
- `Content` maps to the `content_fields` table (columns: content_field_id, content_data_id, value, etc.)

**See Also:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/DATATYPES_AND_FIELDS.md` for field system details

---

## Key Functions

### NewRoot()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:99-102`

Creates a new empty Root instance.

```go
func NewRoot() Root {
	return Root{}
}
```

**Returns:** Empty Root struct with nil Node

**Usage:**
```go
root := model.NewRoot()
// root.Node is nil, ready to be populated
```

---

### NewNode()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:111-115`

Creates a new Node with the given Datatype.

```go
func NewNode(d Datatype) Node {
	return Node{
		Datatype: d,
	}
}
```

**Parameters:**
- `d Datatype`: The datatype to assign to this node

**Returns:** Node with empty Fields and Nodes slices

**Usage:**
```go
datatype := model.Datatype{
	Info: db.DatatypeJSON{DatatypeID: 1},
	Content: db.ContentDataJSON{ContentDataID: 100},
}
node := model.NewNode(datatype)
```

---

### AddChild()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:104-109`

Adds a node as the root node of a Root struct.

```go
func AddChild(r Root, n *Node) Root {
	if r.Node == nil {
		r.Node = n
	}
	return r
}
```

**Parameters:**
- `r Root`: The root to modify
- `n *Node`: The node to add as root

**Returns:** Modified Root

**Note:** Despite the name "AddChild", this function sets the root node, not a child. Use `Node.AddChild()` for adding children to nodes.

**Usage:**
```go
root := model.NewRoot()
node := &model.Node{}
root = model.AddChild(root, node)
```

---

### Node.AddChild()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:118-120`

Adds a child node to a Node.

```go
func (n *Node) AddChild(child *Node) {
	n.Nodes = append(n.Nodes, child)
}
```

**Parameters:**
- `child *Node`: The child node to add

**Effects:** Appends child to n.Nodes slice

**Usage:**
```go
parentNode := &model.Node{}
childNode := &model.Node{}
parentNode.AddChild(childNode)
```

---

### Node.FindNodeByID()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:84-97`

Recursively searches the tree for a node with the given datatype ID.

```go
func (n *Node) FindNodeByID(id int64) *Node {
	if n.Datatype.Info.DatatypeID == id {
		return n
	}

	for _, child := range n.Nodes {
		if found := child.FindNodeByID(id); found != nil {
			return found
		}
	}

	return nil
}
```

**Parameters:**
- `id int64`: The datatype_id to search for

**Returns:** Pointer to matching Node, or nil if not found

**Algorithm:** Depth-first search through node hierarchy

**Usage:**
```go
found := root.Node.FindNodeByID(42)
if found != nil {
	// Process found node
}
```

---

### Root.Render()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:76-82`

Renders the entire tree as a JSON string.

```go
func (r Root) Render() string {
	j, err := json.Marshal(r)
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	return string(j)
}
```

**Returns:** JSON string representation of the tree

**Usage:**
```go
root := model.BuildTree(contentData, datatypes, contentFields, fields)
jsonOutput := root.Render()
// Send jsonOutput as API response
```

---

## JSON Serialization

The model package implements custom JSON marshaling to handle circular references and provide clean JSON output.

### Node.MarshalJSON()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:31-57`

Custom JSON marshaling for Node that prevents circular references.

```go
func (n Node) MarshalJSON() ([]byte, error) {
	type CustomNode struct {
		Datatype Datatype `json:"datatype"`
		Fields   []Field  `json:"fields"`
		Nodes    []*Node  `json:"nodes,omitempty"`
	}

	var nodes []*Node
	if n.Nodes != nil {
		for _, node := range n.Nodes {
			if node != &n { // Avoid self-reference
				nodes = append(nodes, node)
			}
		}
	}

	custom := CustomNode{
		Datatype: n.Datatype,
		Fields:   n.Fields,
		Nodes:    nodes,
	}

	return json.Marshal(custom)
}
```

**Key Features:**
1. **Self-reference detection**: Skips nodes that reference themselves
2. **Omitempty on nodes**: Empty child arrays not included in output
3. **Type aliasing**: Prevents infinite recursion during marshaling

**Why This Matters:**
Without this custom marshaling, a node that accidentally references itself would cause `json.Marshal` to infinitely recurse and crash.

---

### Node.UnmarshalJSON()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model.go:59-73`

Custom JSON unmarshaling for Node.

```go
func (n *Node) UnmarshalJSON(data []byte) error {
	type NodeAlias Node

	aux := &struct {
		*NodeAlias
	}{
		NodeAlias: (*NodeAlias)(n),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	return nil
}
```

**Purpose:** Provides symmetry with MarshalJSON and allows for future custom deserialization logic.

---

## Tree Building

### BuildTree()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/build.go:8-25`

Builds a complete Root tree from database records.

```go
func BuildTree(cd []db.ContentData, dt []db.Datatypes, cf []db.ContentFields, df []db.Fields) Root {
	d := make([]Datatype, len(cd))
	f := make([]Field, len(cf))

	// Map database types to model types
	for i, v := range cd {
		d[i].Info = db.MapDatatypeJSON(dt[i])
		d[i].Content = db.MapContentDataJSON(v)
	}
	for i, v := range cf {
		f[i].Info = db.MapFieldJSON(df[i])
		f[i].Content = db.MapContentFieldJSON(v)
	}

	// Build node hierarchy
	nodes := BuildNodes(d, f)
	root := NewRoot()
	root.Node = nodes

	return root
}
```

**Parameters:**
- `cd []db.ContentData`: Content data records from database
- `dt []db.Datatypes`: Datatype definitions from database
- `cf []db.ContentFields`: Content field values from database
- `df []db.Fields`: Field definitions from database

**Returns:** Complete Root tree with all nodes and fields populated

**Process:**
1. Map database types to JSON types using db.Map* functions
2. Build node hierarchy with BuildNodes()
3. Assign field values to nodes
4. Return complete tree

**Usage:**
```go
// Fetch from database
contentData, _ := driver.GetContentData(ctx, routeID)
datatypes, _ := driver.GetDatatypes(ctx)
contentFields, _ := driver.GetContentFields(ctx, routeID)
fields, _ := driver.GetFields(ctx)

// Build tree
root := model.BuildTree(contentData, datatypes, contentFields, fields)

// Render as JSON
jsonOutput := root.Render()
```

---

### BuildNodes()

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/build.go:26-78`

Builds the node hierarchy and attaches fields to nodes.

```go
func BuildNodes(datatypes []Datatype, fields []Field) *Node {
	// Build a slice of nodes from the datatypes
	nodes := make([]*Node, len(datatypes))
	for i, dt := range datatypes {
		nodes[i] = &Node{
			Datatype: dt,
			Fields:   []Field{},
			Nodes:    []*Node{},
		}
	}

	// Helper function to find a node by ContentDataID
	findNode := func(id int64) *Node {
		for _, node := range nodes {
			if node.Datatype.Content.ContentDataID == id {
				return node
			}
		}
		return nil
	}

	var root *Node

	// Build the tree by assigning each node to its parent
	for _, node := range nodes {
		// Identify the root node
		if node.Datatype.Info.Type == "ROOT" {
			root = node
			continue
		}

		// Avoid self-parenting
		if node.Datatype.Content.ParentID.Int64 != node.Datatype.Content.ContentDataID {
			parent := findNode(node.Datatype.Content.ParentID.Int64)
			if parent != nil {
				parent.Nodes = append(parent.Nodes, node)
			}
		}
	}

	// Associate fields with the corresponding nodes
	for _, field := range fields {
		node := findNode(field.Info.ParentID.Int64)
		if node != nil {
			node.Fields = append(node.Fields, field)
		} else {
			utility.DefaultLogger.Info("no node found for field", field)
		}
	}

	return root
}
```

**Parameters:**
- `datatypes []Datatype`: All datatypes to build into nodes
- `fields []Field`: All fields to attach to nodes

**Returns:** Pointer to root node with complete hierarchy

**Algorithm:**
1. **Phase 1 - Create nodes**: Create a node for each datatype
2. **Phase 2 - Build hierarchy**: Link nodes to their parents using ParentID
3. **Phase 3 - Attach fields**: Associate fields with nodes by parent_id
4. **Root detection**: Find node with Type == "ROOT"
5. **Self-parenting protection**: Skip nodes where ParentID == ContentDataID

**Key Differences from CLI TreeRoot:**
- No sibling pointers (uses parent-child only)
- No NodeIndex map (linear search with findNode helper)
- No orphan resolution
- No lazy loading support
- Simpler structure suitable for JSON output

---

## Tree Operations

### Creating a New Tree

**Example: Creating a simple content tree programmatically**

```go
package main

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

func createContentTree() model.Root {
	// Create root node
	rootDatatype := model.Datatype{
		Info: db.DatatypeJSON{
			DatatypeID: 1,
			Type:       "ROOT",
			Name:       "Site Root",
		},
		Content: db.ContentDataJSON{
			ContentDataID: 1,
			RouteID:       1,
		},
	}
	rootNode := model.NewNode(rootDatatype)

	// Create page node
	pageDatatype := model.Datatype{
		Info: db.DatatypeJSON{
			DatatypeID: 2,
			Type:       "page",
			Name:       "Home Page",
		},
		Content: db.ContentDataJSON{
			ContentDataID: 2,
			ParentID:      sql.NullInt64{Int64: 1, Valid: true},
			RouteID:       1,
		},
	}
	pageNode := model.NewNode(pageDatatype)

	// Add fields to page
	titleField := model.Field{
		Info: db.FieldsJSON{
			FieldID: 1,
			Name:    "title",
			Type:    "text",
		},
		Content: db.ContentFieldsJSON{
			ContentFieldID: 1,
			ContentDataID:  2,
			Value:          "Welcome to ModulaCMS",
		},
	}
	pageNode.Fields = append(pageNode.Fields, titleField)

	// Build tree
	rootNode.AddChild(&pageNode)

	root := model.NewRoot()
	root.Node = &rootNode

	return root
}
```

---

### Traversing a Tree

**Example: Visiting all nodes in a tree**

```go
func traverseTree(node *model.Node, depth int) {
	if node == nil {
		return
	}

	// Process current node
	indent := strings.Repeat("  ", depth)
	fmt.Printf("%s- %s (ID: %d)\n",
		indent,
		node.Datatype.Info.Name,
		node.Datatype.Content.ContentDataID)

	// Process fields
	for _, field := range node.Fields {
		fmt.Printf("%s  [%s]: %s\n",
			indent,
			field.Info.Name,
			field.Content.Value)
	}

	// Recursively traverse children
	for _, child := range node.Nodes {
		traverseTree(child, depth+1)
	}
}

// Usage
root := model.BuildTree(contentData, datatypes, contentFields, fields)
traverseTree(root.Node, 0)
```

---

### Finding Nodes

**Example: Finding a node and modifying it**

```go
func updateNodeTitle(root model.Root, nodeID int64, newTitle string) error {
	node := root.Node.FindNodeByID(nodeID)
	if node == nil {
		return fmt.Errorf("node %d not found", nodeID)
	}

	// Find title field
	for i, field := range node.Fields {
		if field.Info.Name == "title" {
			node.Fields[i].Content.Value = newTitle
			return nil
		}
	}

	return fmt.Errorf("title field not found on node %d", nodeID)
}
```

---

## Relationship to CLI TreeRoot

The model package's tree structure is **different** from the CLI's TreeRoot/TreeNode structure documented in TREE_STRUCTURE.md. Understanding the differences is critical:

### Model Package (internal/model)

**Structure:**
```go
type Root struct {
	Node *Node
}

type Node struct {
	Datatype Datatype
	Fields   []Field
	Nodes    []*Node  // Children stored in slice
}
```

**Characteristics:**
- **Hierarchical**: Parent-child via slices
- **JSON-optimized**: Easy serialization
- **Simple**: No sibling pointers
- **Complete**: All fields included in nodes
- **Static**: No lazy loading
- **For output**: API responses, exports

---

### CLI TreeRoot (internal/cli)

**Structure:**
```go
type TreeRoot struct {
	Root      *TreeNode
	NodeIndex map[int64]*TreeNode
	Orphans   map[int64]*TreeNode
	MaxRetry  int
}

type TreeNode struct {
	ContentDataID int64
	ParentID      sql.NullInt64
	FirstChildID  sql.NullInt64
	NextSiblingID sql.NullInt64
	PrevSiblingID sql.NullInt64
	// ... other fields
}
```

**Characteristics:**
- **Sibling-pointer tree**: O(1) sibling operations
- **Indexed**: NodeIndex map for O(1) lookups
- **Complex**: Three-phase loading, orphan resolution
- **Lazy**: Supports on-demand child loading
- **Minimal**: Fields loaded separately
- **For navigation**: TUI tree operations

---

### When to Use Which

**Use model.Root when:**
- Building JSON API responses
- Exporting content to external systems
- Creating content snapshots
- Testing content structure
- You need a complete self-contained representation

**Use CLI TreeRoot when:**
- Building TUI navigation trees
- Need O(1) sibling access
- Working with large trees (lazy loading)
- Need efficient tree manipulation
- Building the content editor interface

**See Also:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` for CLI TreeRoot documentation
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` for TUI tree usage

---

## Usage Patterns

### Pattern 1: Building JSON API Response

```go
// In an HTTP handler
func handleGetContent(w http.ResponseWriter, r *http.Request) {
	routeID := getRouteIDFromRequest(r)

	// Fetch data from database
	contentData, _ := driver.GetContentData(ctx, routeID)
	datatypes, _ := driver.GetDatatypes(ctx)
	contentFields, _ := driver.GetContentFields(ctx, routeID)
	fields, _ := driver.GetFields(ctx)

	// Build tree
	root := model.BuildTree(contentData, datatypes, contentFields, fields)

	// Render as JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(root.Render()))
}
```

---

### Pattern 2: Content Export

```go
func exportContentToFile(routeID int64, filename string) error {
	// Build tree from database
	contentData, _ := driver.GetContentData(ctx, routeID)
	datatypes, _ := driver.GetDatatypes(ctx)
	contentFields, _ := driver.GetContentFields(ctx, routeID)
	fields, _ := driver.GetFields(ctx)

	root := model.BuildTree(contentData, datatypes, contentFields, fields)

	// Write to file
	jsonData := root.Render()
	return os.WriteFile(filename, []byte(jsonData), 0644)
}
```

---

### Pattern 3: Content Import

```go
func importContentFromFile(filename string) (model.Root, error) {
	// Read file
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return model.Root{}, err
	}

	// Unmarshal into Root
	var root model.Root
	err = json.Unmarshal(jsonData, &root)
	if err != nil {
		return model.Root{}, err
	}

	return root, nil
}
```

---

### Pattern 4: Testing Content Structure

```go
func TestContentHierarchy(t *testing.T) {
	// Create test data
	root := createTestContentTree()

	// Verify structure
	if root.Node == nil {
		t.Fatal("Expected root node")
	}

	if len(root.Node.Nodes) != 2 {
		t.Errorf("Expected 2 children, got %d", len(root.Node.Nodes))
	}

	// Find specific node
	found := root.Node.FindNodeByID(42)
	if found == nil {
		t.Error("Expected to find node 42")
	}
}
```

---

## Common Workflows

### Workflow 1: Adding model.Root to API Endpoint

**Goal:** Create a new API endpoint that returns content as JSON

**Steps:**

1. **Create handler function**

```go
// In cmd/main.go or a new handlers package
func handleContentAPI(db db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 2: Parse route ID
		routeID := r.URL.Query().Get("route_id")
		id, err := strconv.ParseInt(routeID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid route_id", http.StatusBadRequest)
			return
		}

		// Step 3: Fetch data
		ctx := r.Context()
		contentData, err := db.GetContentData(ctx, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		datatypes, _ := db.GetDatatypes(ctx)
		contentFields, _ := db.GetContentFields(ctx, id)
		fields, _ := db.GetFields(ctx)

		// Step 4: Build tree
		root := model.BuildTree(contentData, datatypes, contentFields, fields)

		// Step 5: Return JSON
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(root.Render()))
	}
}
```

2. **Register route in main.go**

```go
mux.HandleFunc("/api/content", handleContentAPI(dbDriver))
```

3. **Test endpoint**

```bash
curl http://localhost:8080/api/content?route_id=1
```

---

### Workflow 2: Creating Custom JSON Structure

**Goal:** Transform model.Root into a custom JSON format for external system

**Steps:**

1. **Define custom structure**

```go
type CustomOutput struct {
	PageTitle string                 `json:"page_title"`
	Content   map[string]interface{} `json:"content"`
	Children  []CustomOutput         `json:"children,omitempty"`
}
```

2. **Create transformation function**

```go
func transformNode(node *model.Node) CustomOutput {
	output := CustomOutput{
		Content:  make(map[string]interface{}),
		Children: []CustomOutput{},
	}

	// Extract title from fields
	for _, field := range node.Fields {
		if field.Info.Name == "title" {
			output.PageTitle = field.Content.Value
		}
		output.Content[field.Info.Name] = field.Content.Value
	}

	// Transform children recursively
	for _, child := range node.Nodes {
		output.Children = append(output.Children, transformNode(child))
	}

	return output
}
```

3. **Use in handler**

```go
root := model.BuildTree(contentData, datatypes, contentFields, fields)
customOutput := transformNode(root.Node)
json.NewEncoder(w).Encode(customOutput)
```

---

### Workflow 3: Programmatically Building Content

**Goal:** Create content programmatically without database

**Use Case:** Testing, fixtures, or content generation

```go
func buildSampleSite() model.Root {
	// Create root
	root := model.NewRoot()

	// Create home page
	homeDatatype := model.Datatype{
		Info:    db.DatatypeJSON{DatatypeID: 1, Name: "Home", Type: "page"},
		Content: db.ContentDataJSON{ContentDataID: 1},
	}
	homeNode := model.NewNode(homeDatatype)
	homeNode.Fields = []model.Field{
		{
			Info:    db.FieldsJSON{FieldID: 1, Name: "title"},
			Content: db.ContentFieldsJSON{Value: "Home Page"},
		},
	}

	// Create about page
	aboutDatatype := model.Datatype{
		Info:    db.DatatypeJSON{DatatypeID: 2, Name: "About", Type: "page"},
		Content: db.ContentDataJSON{ContentDataID: 2, ParentID: sql.NullInt64{Int64: 1, Valid: true}},
	}
	aboutNode := model.NewNode(aboutDatatype)

	// Build tree
	homeNode.AddChild(&aboutNode)
	root.Node = &homeNode

	return root
}
```

---

## Testing

### Test Files

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/model_test.go`

The model package includes comprehensive tests for tree building and operations.

---

### Running Tests

```bash
# Run all model tests
go test -v ./internal/model/

# Run specific test
go test -v ./internal/model/ -run TestBuildTree

# Run with coverage
go test -cover ./internal/model/
```

---

### Example Test: Building a Tree

```go
func TestBuildTree(t *testing.T) {
	// Create test data
	contentData := []db.ContentData{
		{ContentDataID: 1, DatatypeID: 1, ParentID: sql.NullInt64{}},
		{ContentDataID: 2, DatatypeID: 2, ParentID: sql.NullInt64{Int64: 1, Valid: true}},
	}

	datatypes := []db.Datatypes{
		{DatatypeID: 1, Name: "Root", Type: "ROOT"},
		{DatatypeID: 2, Name: "Page", Type: "page"},
	}

	contentFields := []db.ContentFields{
		{ContentFieldID: 1, ContentDataID: 2, FieldID: 1, Value: "Test Page"},
	}

	fields := []db.Fields{
		{FieldID: 1, Name: "title", Type: "text"},
	}

	// Build tree
	root := model.BuildTree(contentData, datatypes, contentFields, fields)

	// Assertions
	if root.Node == nil {
		t.Fatal("Expected root node")
	}

	if root.Node.Datatype.Info.Type != "ROOT" {
		t.Errorf("Expected ROOT type, got %s", root.Node.Datatype.Info.Type)
	}

	if len(root.Node.Nodes) != 1 {
		t.Errorf("Expected 1 child, got %d", len(root.Node.Nodes))
	}

	child := root.Node.Nodes[0]
	if len(child.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(child.Fields))
	}
}
```

---

### Example Test: JSON Marshaling

```go
func TestJSONMarshaling(t *testing.T) {
	// Create node
	node := model.NewNode(model.Datatype{
		Info: db.DatatypeJSON{DatatypeID: 1, Name: "Test"},
	})

	// Marshal to JSON
	jsonData, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded model.Node
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify
	if decoded.Datatype.Info.DatatypeID != 1 {
		t.Errorf("Expected DatatypeID 1, got %d", decoded.Datatype.Info.DatatypeID)
	}
}
```

---

### Example Test: Circular Reference Protection

```go
func TestCircularReferenceProtection(t *testing.T) {
	// Create node that references itself
	node := model.NewNode(model.Datatype{
		Info: db.DatatypeJSON{DatatypeID: 1},
	})

	// Create circular reference (should be handled by MarshalJSON)
	node.Nodes = append(node.Nodes, &node)

	// This should not panic or infinite loop
	jsonData, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Failed to marshal with circular ref: %v", err)
	}

	// Verify JSON is valid
	var decoded model.Node
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
}
```

---

## Related Documentation

### Architecture
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - CLI TreeRoot sibling-pointer implementation
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Domain model and database relationships
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - Database abstraction

### Packages
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database types and DbDriver interface
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` - TUI usage of TreeRoot

### Domain
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/DATATYPES_AND_FIELDS.md` - Content schema system
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/CONTENT_TREES.md` - Tree operations and navigation

### Workflows
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md` - Testing strategies
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/DEBUGGING.md` - Debugging guide

---

## Quick Reference

### Package Location
```
/Users/home/Documents/Code/Go_dev/modulacms/internal/model/
```

### Key Files
- `model.go` - Core structures (Root, Node, Datatype, Field) and operations
- `cms_model.go` - High-level content functions
- `build.go` - Tree building functions
- `model_test.go` - Tests
- `build_test.go` - Tree building tests

### Core Types
```go
type Root struct { Node *Node }
type Node struct { Datatype Datatype; Fields []Field; Nodes []*Node }
type Datatype struct { Info db.DatatypeJSON; Content db.ContentDataJSON }
type Field struct { Info db.FieldsJSON; Content db.ContentFieldsJSON }
```

### Essential Functions
```go
model.NewRoot() Root                                    // Create empty root
model.NewNode(d Datatype) Node                         // Create node
model.AddChild(r Root, n *Node) Root                   // Set root node
node.AddChild(child *Node)                             // Add child to node
node.FindNodeByID(id int64) *Node                      // Find node by ID
root.Render() string                                   // Render to JSON
model.BuildTree(cd, dt, cf, df) Root                   // Build from DB
model.BuildNodes(datatypes, fields) *Node              // Build hierarchy
```

### Import Statement
```go
import "github.com/hegner123/modulacms/internal/model"
```

### Typical Usage Flow
```go
// 1. Fetch data from database
contentData := driver.GetContentData(ctx, routeID)
datatypes := driver.GetDatatypes(ctx)
contentFields := driver.GetContentFields(ctx, routeID)
fields := driver.GetFields(ctx)

// 2. Build tree
root := model.BuildTree(contentData, datatypes, contentFields, fields)

// 3. Output as JSON
jsonOutput := root.Render()

// 4. Or find and modify nodes
node := root.Node.FindNodeByID(42)
if node != nil {
	// Modify node or fields
}
```

### Key Differences from CLI TreeRoot
| Feature              | model.Root          | CLI TreeRoot        |
|----------------------|---------------------|---------------------|
| Structure            | Parent-child slices | Sibling pointers    |
| Lookup speed         | O(n) linear search  | O(1) via NodeIndex  |
| JSON serialization   | Native support      | Manual conversion   |
| Fields included      | Yes, in nodes       | Loaded separately   |
| Lazy loading         | No                  | Yes                 |
| Use case             | JSON output/export  | TUI navigation      |

### Common Gotchas
1. **AddChild() on Root** - Despite the name, sets root node, not a child
2. **FindNodeByID uses DatatypeID** - Not ContentDataID
3. **Fields attached to nodes** - No separate field loading needed
4. **Self-reference protection** - MarshalJSON handles circular refs automatically
5. **Different from CLI tree** - Don't confuse with TreeRoot/TreeNode

---

**Last Updated:** 2026-01-12
**Next Review:** After Phase 4 completion
