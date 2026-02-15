// Package model provides the hierarchical content tree representation for ModulaCMS.
// It defines Node, Root, Datatype, and Field types that combine type definitions
// with instance data, and offers construction/traversal methods used by tree builders
// (BuildTree, BuildAdminTree) and transform serializers (Contentful, Sanity, etc.).
package model

import (
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
)

// Logger is the logging interface consumed by the model package. Callers pass
// a concrete logger (e.g. *utility.Logger) into BuildTree/BuildAdminTree, which
// forward it to internal functions like BuildNodes for orphan warnings.
type Logger interface {
	Warn(message string, err error, args ...any)
}

// Root is the top-level container for a content tree. It holds a single pointer
// to the root Node of the hierarchy. All tree operations (BuildTree, BuildAdminTree)
// produce a Root, which is then passed to transform.Transformer implementations
// for serialization into various CMS output formats (Contentful, Sanity, etc.).
type Root struct {
	Node *Node `json:"root"`
}

// Node represents a single node in the content hierarchy. Each node combines
// a Datatype (type definition + instance data), a slice of Fields (field
// definitions + values), and child Nodes forming the tree structure.
// Nodes are constructed in BuildNodes and linked via parent-child relationships
// derived from ContentData.ParentID.
type Node struct {
	Datatype Datatype `json:"datatype"`
	Fields   []Field  `json:"fields"`
	Nodes    []*Node  `json:"nodes"`
}

// Datatype pairs a type definition (Info: label, type name, slug) with a
// content instance (Content: content data ID, parent ID, route, author, dates).
// Info comes from the datatypes/admin_datatypes table; Content comes from
// the content_data/admin_content_data table.
type Datatype struct {
	Info    db.DatatypeJSON    `json:"info"`
	Content db.ContentDataJSON `json:"content"`
}

// Field pairs a field definition (Info: field ID, label, type, parent ID) with
// a content field value (Content: value, content data ID, dates).
// Info comes from fields/admin_fields; Content comes from
// content_fields/admin_content_fields.
type Field struct {
	Info    db.FieldsJSON        `json:"info"`
	Content db.ContentFieldsJSON `json:"content"`
}

// MarshalJSON provides custom JSON encoding for Node to prevent circular
// references during serialization. It copies the child slice while filtering
// out any node that points back to itself, then delegates to json.Marshal
// on a local CustomNode type (which has omitempty on Nodes to compact output
// for leaf nodes).
func (n Node) MarshalJSON() ([]byte, error) {
	type CustomNode struct {
		Datatype Datatype `json:"datatype"`
		Fields   []Field  `json:"fields"`
		Nodes    []*Node  `json:"nodes,omitempty"`
	}

	// Filter out self-references to prevent infinite recursion during
	// json.Marshal. This can happen if tree construction accidentally
	// adds a node as its own child. Ranging over a nil slice is a no-op
	// in Go, so no nil guard is needed.
	var nodes []*Node
	for _, node := range n.Nodes {
		if node == &n {
			continue
		}
		nodes = append(nodes, node)
	}

	custom := CustomNode{
		Datatype: n.Datatype,
		Fields:   n.Fields,
		Nodes:    nodes,
	}

	return json.Marshal(custom)
}

// UnmarshalJSON uses the NodeAlias pattern to decode JSON into a Node without
// triggering infinite recursion (a direct json.Unmarshal into *Node would call
// this method again). The alias type has the same memory layout but no methods,
// so json.Unmarshal uses the default decoder.
func (n *Node) UnmarshalJSON(data []byte) error {
	type NodeAlias Node

	aux := &struct {
		*NodeAlias
	}{
		NodeAlias: (*NodeAlias)(n),
	}

	return json.Unmarshal(data, aux)
}

// Render serializes the entire Root (including all nested Nodes) to a JSON string.
// Used by the router layer to produce HTTP response bodies.
func (r Root) Render() (string, error) {
	j, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("failed to marshal content tree: %w", err)
	}
	return string(j), nil
}

// FindNodeByID performs a depth-first recursive search through the tree
// looking for a node whose ContentDataID matches the given id.
func (n *Node) FindNodeByID(id string) *Node {
	if n.Datatype.Content.ContentDataID == id {
		return n
	}

	for _, child := range n.Nodes {
		if found := child.FindNodeByID(id); found != nil {
			return found
		}
	}

	return nil
}

// NewRoot returns an empty Root with a nil Node pointer.
func NewRoot() Root {
	return Root{}
}

// SetRootNode sets the root node of the tree if the Root is currently empty (nil Node).
// Does nothing if the root already has a node.
func SetRootNode(r Root, n *Node) Root {
	if r.Node == nil {
		r.Node = n
	}
	return r
}

// NewNode creates a Node from a Datatype with empty fields and children slices.
func NewNode(d Datatype) Node {
	return Node{
		Datatype: d,
		Fields:   []Field{},
		Nodes:    []*Node{},
	}
}

// AddChild appends a child node to this node's Nodes slice.
// This is the correct append-based child addition, unlike the free function
// AddChild which only sets the root.
func (n *Node) AddChild(child *Node) {
	n.Nodes = append(n.Nodes, child)
}
