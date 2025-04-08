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
    /*
	jd, err := json.Marshal(d)
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
	jf, err := json.Marshal(f)
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
    */
	nodes := BuildNodes(d, f)
	//utility.DefaultLogger.Info("datatypes", string(jd))
	//utility.DefaultLogger.Info("fields", string(jf))
	root := NewRoot(d[0], f[0:1])
	root.Node = nodes

	return root
}

// We assume that each Datatype.Content.ContentDataID is unique and that a node is a root
// when its ParentID matches its own ContentDataID (or when Info.ParentID is nil).
func BuildNodes(datatypes []Datatype, fields []Field) *Node {
	// Create a lookup map keyed by ContentDataID.
	nodesByID := make(map[int64]*Node)

	// First, create all nodes from datatypes.
	for _, dt := range datatypes {
		node := &Node{
			Datatype: dt,
			Fields:   []Field{},
			Nodes:    []*Node{},
		}
		// Use the ContentDataID as the unique identifier.
		nodesByID[dt.Content.ContentDataID] = node
	}

	// Next, set up the parent/child hierarchy.
	var root *Node
	for _, node := range nodesByID {
		// If Info.ParentID is nil, treat it as a root.
		if node.Datatype.Info.Type == "ROOT" {
			root = node
			continue
		}
		// Alternatively, if the content's ParentID differs from its own ContentDataID,
		// then it belongs as a child to the node with that id.
		if node.Datatype.Content.ParentID.Int64 != node.Datatype.Content.ContentDataID {
			// In many systems, a node whose parent_id equals its own id is a root.
			// Otherwise, attach the node to its parent's Nodes slice.
			if parent, ok := nodesByID[node.Datatype.Content.ParentID.Int64]; ok {
				parent.Nodes = append(parent.Nodes, node)
			}
		}
	}

	// Finally, assign fields to their corresponding node based on ContentDataID.
	for _, field := range fields {
		if node, ok := nodesByID[field.Info.ParentID.Int64]; ok {
			node.Fields = append(node.Fields, field)
		} else {
			utility.DefaultLogger.Info("no node found for field", field)
			// Optionally log or handle fields without a node match.
			// log.Printf("No node found for field: %v", field)
		}
	}

	return root
}
