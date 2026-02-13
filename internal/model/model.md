# model

Package model provides the core data structures and tree-building logic for ModulaCMS content hierarchies. It constructs in-memory content trees from flat database entities and provides JSON serialization for HTTP responses.

## Overview

The model package defines the Node-based tree structure used throughout ModulaCMS. It pairs type definitions with content instances, assembles hierarchical relationships via parent-child pointers, and handles sibling ordering using database-stored chains. The package supports both public API trees and admin API trees through separate build functions.

## Types

### Logger

Logger is the logging interface consumed by the model package.

```go
type Logger interface {
    Warn(message string, err error, args ...any)
}
```

Callers pass a concrete logger into BuildTree and BuildAdminTree, which forward it to internal functions for orphan warnings.

### Root

Root is the top-level container for a content tree.

```go
type Root struct {
    Node *Node `json:"root"`
}
```

All tree operations produce a Root, which is then passed to transform.Transformer implementations for serialization into various CMS output formats like Contentful or Sanity.

### Node

Node represents a single node in the content hierarchy.

```go
type Node struct {
    Datatype Datatype `json:"datatype"`
    Fields   []Field  `json:"fields"`
    Nodes    []*Node  `json:"nodes"`
}
```

Each node combines a Datatype type definition plus instance data, a slice of Fields with field definitions plus values, and child Nodes forming the tree structure. Nodes are constructed in BuildNodes and linked via parent-child relationships derived from ContentData.ParentID.

### Datatype

Datatype pairs a type definition with a content instance.

```go
type Datatype struct {
    Info    db.DatatypeJSON    `json:"info"`
    Content db.ContentDataJSON `json:"content"`
}
```

Info contains label, type name, and slug from the datatypes or admin_datatypes table. Content contains content data ID, parent ID, route, author, and dates from the content_data or admin_content_data table.

### Field

Field pairs a field definition with a content field value.

```go
type Field struct {
    Info    db.FieldsJSON        `json:"info"`
    Content db.ContentFieldsJSON `json:"content"`
}
```

Info comes from fields or admin_fields table with field ID, label, type, and parent ID. Content comes from content_fields or admin_content_fields with value, content data ID, and dates.

## Functions

### NewRoot

```go
func NewRoot() Root
```

NewRoot returns an empty Root with a nil Node pointer.

### SetRootNode

```go
func SetRootNode(r Root, n *Node) Root
```

SetRootNode sets the root node of the tree if the Root is currently empty. Does nothing if the root already has a node.

### NewNode

```go
func NewNode(d Datatype) Node
```

NewNode creates a Node from a Datatype with empty fields and children slices.

### BuildTree

```go
func BuildTree(log Logger, cd []db.ContentData, dt []db.Datatypes, cf []db.ContentFields, df []db.Fields) (Root, error)
```

BuildTree constructs a content tree from public non-admin database entities. It accepts four parallel slices that are correlated by index. The cd and dt slices are paired where cd[i] is the content instance and dt[i] is its type definition. The cf and df slices are paired where cf[i] is the field value and df[i] is its field definition.

Each pair is mapped into the model Datatype and Field types using the db package MapXxxJSON converters, then passed to BuildNodes to assemble the hierarchy. Called by the router layer to build trees for public API responses.

Returns an error if slice lengths do not match or if orphan nodes or fields are detected.

### BuildAdminTree

```go
func BuildAdminTree(log Logger, cd []db.AdminContentData, dt []db.AdminDatatypes, cf []db.AdminContentFields, df []db.AdminFields) (Root, error)
```

BuildAdminTree constructs a content tree from admin database entities. Same structure as BuildTree but uses admin-prefixed DB types. Called by the router layer to build trees for admin API responses.

Mutates the mapped FieldsJSON.ParentID after creation because AdminFields.ParentID refers to the datatype owner at type-level, but BuildNodes needs ParentID to point to the content data instance. The fix reassigns ParentID from AdminContentFields.AdminContentDataID.

Returns an error if slice lengths do not match or if orphan nodes or fields are detected.

### BuildNodes

```go
func BuildNodes(log Logger, datatypes []Datatype, fields []Field) (*Node, error)
```

BuildNodes is the core tree-assembly algorithm. It takes flat slices of Datatypes one per content node and Fields one per field value and produces a tree by creating one Node per Datatype with empty Fields and Nodes slices, building a map index for O(1) lookups by ContentDataID, identifying the root node with Type equals ROOT, linking each non-root node to its parent via ContentData.ParentID, and attaching fields to their owning node via Field.Info.ParentID.

Returns the root Node pointer which is nil if no node has Type ROOT and an error if orphan nodes or fields were encountered.

## Methods

### Node.MarshalJSON

```go
func (n Node) MarshalJSON() ([]byte, error)
```

MarshalJSON provides custom JSON encoding for Node to prevent circular references during serialization. It copies the child slice while filtering out any node that points back to itself, then delegates to json.Marshal on a local CustomNode type which has omitempty on Nodes to compact output for leaf nodes.

### Node.UnmarshalJSON

```go
func (n *Node) UnmarshalJSON(data []byte) error
```

UnmarshalJSON uses the NodeAlias pattern to decode JSON into a Node without triggering infinite recursion. A direct json.Unmarshal into Node pointer would call this method again. The alias type has the same memory layout but no methods, so json.Unmarshal uses the default decoder.

### Node.AddChild

```go
func (n *Node) AddChild(child *Node)
```

AddChild appends a child node to this node Nodes slice. This is the correct append-based child addition.

### Node.FindNodeByID

```go
func (n *Node) FindNodeByID(id string) *Node
```

FindNodeByID performs a depth-first recursive search through the tree looking for a node whose ContentDataID matches the given id. Returns nil if not found.

### Root.Render

```go
func (r Root) Render() (string, error)
```

Render serializes the entire Root including all nested Nodes to a JSON string. Used by the router layer to produce HTTP response bodies. Returns an error if marshaling fails.

## Tree Assembly Algorithm

BuildNodes executes in three phases. Phase 1 creates a flat slice of Node pointers, one per Datatype, and builds a map index for O(1) lookups by ContentDataID. Phase 2 builds the tree hierarchy by linking children to parents, identifying the root node with Type ROOT, and appending all other nodes to their parent Nodes slice. Phase 2.5 reorders children to match stored sibling pointer ordering by calling reorderChildren. Phase 3 associates each field with its owning node by matching Field.Info.ParentID to node Datatype.Content.ContentDataID.

Orphan nodes occur when a node references a parent that does not exist in the index. Orphan fields occur when a field references a parent node that does not exist. All orphans are logged as warnings and returned as part of the error.

Self-parenting is guarded against to prevent infinite loops during traversal or serialization. Nodes where ParentID equals ContentDataID are skipped during hierarchy construction.

## Sibling Ordering

The reorderChildren function sorts each node Nodes slice to match the stored sibling-pointer ordering defined by FirstChildID and NextSiblingID chain. Nodes not found in the chain are appended at the end, preserving partial ordering when chains are incomplete.

The buildSiblingChain function follows the NextSiblingID chain starting from firstChildID and returns nodes in correct display order. Returns nil on cycle detection or if firstChildID is not found in the index.

The mergeOrdered function returns a slice with chain nodes first, followed by any nodes in existing that are not in the chain. Preserves all children even when the pointer chain is incomplete.

Cycle detection prevents infinite loops when NextSiblingID chains point back to an already-visited node. When a cycle is detected, buildSiblingChain returns nil and the original BuildNodes append order is preserved.
