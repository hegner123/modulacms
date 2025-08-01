package model

import (
	"encoding/json"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

type Root struct {
	Node *Node `json:"root"`
}

// Node represents a node in the content tree
type Node struct {
	Datatype Datatype `json:"datatype"`
	Fields   []Field  `json:"fields"`
	Nodes    []*Node  `json:"nodes"`
}

type Datatype struct {
	Info    db.DatatypeJSON    `json:"info"`
	Content db.ContentDataJSON `json:"content"`
}

type Field struct {
	Info    db.FieldsJSON        `json:"info"`
	Content db.ContentFieldsJSON `json:"content"`
}

func (n Node) MarshalJSON() ([]byte, error) {
	// Create a custom structure that mirrors Node but without circular references
	type CustomNode struct {
		Datatype Datatype `json:"datatype"`
		Fields   []Field  `json:"fields"`
		Nodes    []*Node  `json:"nodes,omitempty"`
	}

	// Create a copy to avoid circular reference
	var nodes []*Node
	if n.Nodes != nil {
		// Avoid self-references
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

func (n *Node) UnmarshalJSON(data []byte) error {
	type NodeAlias Node // Create an alias to avoid infinite recursion

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

// Render renders this node and all its children recursively
func (r Root) Render() string {
	j, err := json.Marshal(r)
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	return string(j)
}

// FindNodeByID searches the tree for a node with the given ID
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

func NewRoot() Root {
	return Root{}

}

func AddChild(r Root, n *Node) Root {
	if r.Node == nil {
		r.Node = n
	}
	return r
}

func NewNode(d Datatype) Node {
	return Node{
		Datatype: d,
	}
}

// AddChild adds a child node to this node
func (n *Node) AddChild(child *Node) {
	n.Nodes = append(n.Nodes, child)
}
