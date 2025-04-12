package model

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// Mock data creation functions
func createMockDatatype() db.DatatypeJSON {
	p := sql.NullInt64{Valid: false, Int64: 0}
	dc := sql.NullString{String: "2023-01-01", Valid: true}
	dm := sql.NullString{String: "2023-01-01", Valid: true}
	h := sql.NullString{String: "", Valid: false}
	return db.DatatypeJSON{
		DatatypeID: 1,
		ParentID: db.NullInt64{
			NullInt64: p,
		},
		Label:    "Page",
		Type:     "content",
		AuthorID: 1,
		DateCreated: db.NullString{
			NullString: dc,
		},
		DateModified: db.NullString{
			NullString: dm,
		},
		History: db.NullString{
			NullString: h,
		},
	}
}

func createMockContentData() db.ContentDataJSON {
	p := sql.NullInt64{Int64: 0, Valid: false}
	dc := sql.NullString{String: "2023-01-01", Valid: true}
	dm := sql.NullString{String: "2023-01-01", Valid: true}
	h := sql.NullString{String: "", Valid: false}
	return db.ContentDataJSON{
		ContentDataID: 1,
		ParentID: db.NullInt64{
			NullInt64: p,
		},
		RouteID:    1,
		DatatypeID: 1,
		AuthorID:   1,
		DateCreated: db.NullString{
			NullString: dc,
		},
		DateModified: db.NullString{
			NullString: dm,
		},
		History: db.NullString{
			NullString: h,
		},
	}
}

func createMockField() db.FieldsJSON {
	p := sql.NullInt64{Int64: 0, Valid: false}
	dc := sql.NullString{String: "2023-01-01", Valid: true}
	dm := sql.NullString{String: "2023-01-01", Valid: true}
	h := sql.NullString{String: "", Valid: false}
	return db.FieldsJSON{
		FieldID: 1,
		ParentID: db.NullInt64{
			NullInt64: p,
		},
		Label:    "Title",
		Data:     "string",
		Type:     "text",
		AuthorID: 1,
		DateCreated: db.NullString{
			NullString: dc,
		},
		DateModified: db.NullString{
			NullString: dm,
		},
		History: db.NullString{
			NullString: h,
		},
	}
}

func createMockContentField() db.ContentFieldsJSON {
	r := sql.NullInt64{Int64: 1, Valid: true}
	dc := sql.NullString{String: "2023-01-01", Valid: true}
	dm := sql.NullString{String: "2023-01-01", Valid: true}
	h := sql.NullString{String: "", Valid: false}
	return db.ContentFieldsJSON{
		ContentFieldID: 1,
		RouteID: db.NullInt64{
			NullInt64: r,
		},
		ContentDataID: 1,
		FieldID:       1,
		FieldValue:    "Welcome to my page",
		AuthorID:      1,
		DateCreated: db.NullString{
			NullString: dc,
		},
		DateModified: db.NullString{
			NullString: dm,
		},
		History: db.NullString{
			NullString: h,
		},
	}
}

func createMockDatatype2() db.DatatypeJSON {
	p := sql.NullInt64{Int64: 0, Valid: false}
	dc := sql.NullString{String: "2023-01-01", Valid: true}
	dm := sql.NullString{String: "2023-01-01", Valid: true}
	h := sql.NullString{String: "", Valid: false}
	return db.DatatypeJSON{
		DatatypeID: 2,
		ParentID: db.NullInt64{
			NullInt64: p,
		},
		Label:    "Section",
		Type:     "content",
		AuthorID: 1,
		DateCreated: db.NullString{
			NullString: dc,
		},
		DateModified: db.NullString{
			NullString: dm,
		},
		History: db.NullString{
			NullString: h,
		},
	}
}

func createMockContentData2() db.ContentDataJSON {
	p := sql.NullInt64{Int64: 0, Valid: false}
	dc := sql.NullString{String: "2023-01-01", Valid: true}
	dm := sql.NullString{String: "2023-01-01", Valid: true}
	h := sql.NullString{String: "", Valid: false}
	return db.ContentDataJSON{
		ContentDataID: 2,
		ParentID: db.NullInt64{
			NullInt64: p,
		},
		RouteID:    1,
		DatatypeID: 2,
		AuthorID:   1,
		DateCreated: db.NullString{
			NullString: dc,
		},
		DateModified: db.NullString{
			NullString: dm,
		},
		History: db.NullString{
			NullString: h,
		},
	}
}

func createMockField2() db.FieldsJSON {
	p := sql.NullInt64{Int64: 0, Valid: false}
	dc := sql.NullString{String: "2023-01-01", Valid: true}
	dm := sql.NullString{String: "2023-01-01", Valid: true}
	h := sql.NullString{String: "", Valid: false}
	return db.FieldsJSON{
		FieldID: 2,
		ParentID: db.NullInt64{
			NullInt64: p,
		},
		Label:    "Content",
		Data:     "text",
		Type:     "textarea",
		AuthorID: 1,
		DateCreated: db.NullString{
			NullString: dc,
		},
		DateModified: db.NullString{
			NullString: dm,
		},
		History: db.NullString{
			NullString: h,
		},
	}
}

func createMockContentField2() db.ContentFieldsJSON {
	r := sql.NullInt64{Int64: 1, Valid: true}
	dc := sql.NullString{String: "2023-01-01", Valid: true}
	dm := sql.NullString{String: "2023-01-01", Valid: true}
	h := sql.NullString{String: "", Valid: false}
	return db.ContentFieldsJSON{
		ContentFieldID: 2,
		RouteID: db.NullInt64{
			NullInt64: r,
		},
		ContentDataID: 2,
		FieldID:       2,
		FieldValue:    "This is the content of my page section.",
		AuthorID:      1,
		DateCreated: db.NullString{
			NullString: dc,
		},
		DateModified: db.NullString{
			NullString: dm,
		},
		History: db.NullString{
			NullString: h,
		},
	}
}

// Create test models
func createMockNode() Root {
	root := NewRoot()

	child := &Node{
		Datatype: Datatype{
			Info:    createMockDatatype2(),
			Content: createMockContentData2(),
		},
		Fields: []Field{
			{
				Info:    createMockField2(),
				Content: createMockContentField2(),
			},
		},
		Nodes: nil,
	}

	r := AddChild(root, child)

	// Create new child nodes with unique data instead of reusing the same node
	child2 := &Node{
		Datatype: Datatype{
			Info:    createMockDatatype2(),
			Content: createMockContentData2(),
		},
		Fields: []Field{
			{
				Info:    createMockField2(),
				Content: createMockContentField2(),
			},
		},
		Nodes: nil,
	}

	child3 := &Node{
		Datatype: Datatype{
			Info:    createMockDatatype2(),
			Content: createMockContentData2(),
		},
		Fields: []Field{
			{
				Info:    createMockField2(),
				Content: createMockContentField2(),
			},
		},
		Nodes: nil,
	}
	child4 := &Node{
		Datatype: Datatype{
			Info:    createMockDatatype(),
			Content: createMockContentData(),
		},
		Fields: []Field{
			{
				Info:    createMockField(),
				Content: createMockContentField(),
			},
		},
	}

	// Add different child nodes to avoid circular references
	r.Node.AddChild(child4)
	r.Node.AddChild(child2)
	r.Node.AddChild(child3)
	return r
}

// Test CreateMock function
func TestCreateMock(t *testing.T) {

	threeLayer := createMockNode()
	rendered := threeLayer.Render()
	if len(rendered) < 1 {
		err := fmt.Errorf("BLAHGH")
		utility.DefaultLogger.Error("", err)

	}

}
