package model

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

func BuildTree(cd []db.ContentData, dt []db.Datatypes, cf []db.ContentFields, df []db.Fields) Root {
	d := make([]Datatype, len(cd))
	f := make([]Field, len(cf))
	for i, v := range cd {
		d[i].Info = db.MapDatatypeJSON(dt[i])
		d[i].Content = db.MapContentDataJSON(v)
	}
	for i, v := range cf {
		f[i].Info = db.MapFieldJSON(df[i])
		f[i].Content = db.MapContentFieldJSON(v)

	}
	nodes := BuildNodes(d, f)
	root := NewRoot()
	root.Node = nodes

	return root
}
func BuildNodes(datatypes []Datatype, fields []Field) *Node {
	// Build a slice of nodes from the datatypes.
	nodes := make([]*Node, len(datatypes))
	for i, dt := range datatypes {
		nodes[i] = &Node{
			Datatype: dt,
			Fields:   []Field{},
			Nodes:    []*Node{},
		}
	}

	// Helper function to find a node in the slice by ContentDataID.
	findNode := func(id int64) *Node {
		for _, node := range nodes {
			if node.Datatype.Content.ContentDataID == id {
				return node
			}
		}
		return nil
	}

	var root *Node

	// Build the tree by assigning each node to its parent.
	for _, node := range nodes {
		// Identify the root node.
		if node.Datatype.Info.Type == "ROOT" {
			root = node
			continue
		}

		// Avoid self-parenting.
		if node.Datatype.Content.ParentID.Int64 != node.Datatype.Content.ContentDataID {
			parent := findNode(node.Datatype.Content.ParentID.Int64)
			if parent != nil {
				parent.Nodes = append(parent.Nodes, node)
			}
		}
	}

	// Associate fields with the corresponding nodes.
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

