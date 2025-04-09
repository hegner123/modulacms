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
	nodesByID := make(map[int64]*Node)

	for _, dt := range datatypes {
		node := &Node{
			Datatype: dt,
			Fields:   []Field{},
			Nodes:    []*Node{},
		}
		nodesByID[dt.Content.ContentDataID] = node
	}

	var root *Node
	for _, node := range nodesByID {
		if node.Datatype.Info.Type == "ROOT" {
			root = node
			continue
		}
		if node.Datatype.Content.ParentID.Int64 != node.Datatype.Content.ContentDataID {
			if parent, ok := nodesByID[node.Datatype.Content.ParentID.Int64]; ok {
				parent.Nodes = append(parent.Nodes, node)
			}
		}
	}

	for _, field := range fields {
		if node, ok := nodesByID[field.Info.ParentID.Int64]; ok {
			node.Fields = append(node.Fields, field)
		} else {
			utility.DefaultLogger.Info("no node found for field", field)
		}
	}

	return root
}
