package cli

import "github.com/hegner123/modulacms/internal/db"

type PageCMS struct {
	Root       db.ContentData
	RootType   db.Datatypes
	Fields     []db.ContentFields
	FieldTypes []db.Fields
	Nodes      *[]*NodeCMS
	NodeIndex  map[int64]*NodeCMS
}

type NodeCMS struct {
	Node           db.ContentData
	NodeFields     []db.ContentFields
	NodeDatatype   db.Datatypes
	NodeFieldTypes []db.Fields
	Nodes          *[]*NodeCMS
}

func NewCmsPage(root db.ContentData, rootType db.Datatypes, fields []db.ContentFields, fieldTypes []db.Fields) *PageCMS {
	return &PageCMS{
		Root:       root,
		RootType:   rootType,
		Fields:     fields,
		FieldTypes: fieldTypes,
		Nodes:      nil,
		NodeIndex:  make(map[int64]*NodeCMS, 0),
	}

}

func (page *PageCMS) Insert(newNode NodeCMS, parent int64) bool {
	if page.Nodes == nil {
		if page.Root.ContentDataID == parent {
			nn := make([]*NodeCMS, 0)
			nn = append(nn, &newNode)
			page.Nodes = &nn
			page.NodeIndex[newNode.Node.ContentDataID] = &newNode
			return true
		}
		return false
	} else if page.Root.ContentDataID == parent {
		instance := *page.Nodes
		instance = append(instance, &newNode)
		page.Nodes = &instance
		page.NodeIndex[newNode.Node.ContentDataID] = &newNode
		return true
	}
	instance := *page.Nodes
	for _, v := range instance {
		if v.Node.ContentDataID == parent {
			res := v.Insert(newNode, parent)
			if res {
				page.NodeIndex[newNode.Node.ContentDataID] = &newNode
				return res
			}

			return res
		}
	}
	return false

}

func (node *NodeCMS) Insert(newNode NodeCMS, parent int64) bool {
	if node.Nodes == nil {
		if node.Node.ContentDataID == parent {
			nn := make([]*NodeCMS, 0)
			nn = append(nn, &newNode)
			node.Nodes = &nn
			return true
		}
		return false
	}
	if node.Node.ContentDataID == parent {
		instance := *node.Nodes
		instance = append(instance, &newNode)
		node.Nodes = &instance
		return true
	} else {
		instance := *node.Nodes
		for _, v := range instance {
			if v.Node.ContentDataID == parent {
				return v.Insert(newNode, parent)

			}

		}
	}
	return false

}

func (page *PageCMS) NodeInsertByIndex(index *NodeCMS, newNode NodeCMS) {
	if index.Nodes == nil {
		nn := make([]*NodeCMS, 0)
		nn = append(nn, &newNode)
		index.Nodes = &nn
	} else {
		instance := *index.Nodes
		instance = append(instance, &newNode)
		index.Nodes = &instance
	}
	page.NodeIndex[newNode.Node.ContentDataID] = &newNode

}

func (page *PageCMS) GetFieldValues() []map[string]any {
	var result []map[string]any
	
	// Root field
	rootMap := map[string]any{
		"field_name": "Root",
		"field_type": "db.ContentData",
		"value":      page.Root,
	}
	result = append(result, rootMap)
	
	// RootType field
	rootTypeMap := map[string]any{
		"field_name": "RootType",
		"field_type": "db.Datatypes",
		"value":      page.RootType,
	}
	result = append(result, rootTypeMap)
	
	// Fields field
	fieldsMap := map[string]any{
		"field_name": "Fields",
		"field_type": "[]db.ContentFields",
		"value":      page.Fields,
	}
	result = append(result, fieldsMap)
	
	// FieldTypes field
	fieldTypesMap := map[string]any{
		"field_name": "FieldTypes",
		"field_type": "[]db.Fields",
		"value":      page.FieldTypes,
	}
	result = append(result, fieldTypesMap)
	
	// Nodes field
	nodesMap := map[string]any{
		"field_name": "Nodes",
		"field_type": "*[]*NodeCMS",
		"value":      page.Nodes,
	}
	result = append(result, nodesMap)
	
	// NodeIndex field
	nodeIndexMap := map[string]any{
		"field_name": "NodeIndex",
		"field_type": "map[int64]*NodeCMS",
		"value":      page.NodeIndex,
	}
	result = append(result, nodeIndexMap)
	
	return result
}
