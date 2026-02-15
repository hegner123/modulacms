package definitions

import (
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// callRecord tracks which create method was called and in what order.
type callRecord struct {
	kind string // "field", "datatype", "junction"
}

type mockInstaller struct {
	calls     []callRecord
	datatypes []db.Datatypes
	fields    []db.Fields
	junctions []db.DatatypeFields
}

func (m *mockInstaller) CreateField(p db.CreateFieldParams) db.Fields {
	m.calls = append(m.calls, callRecord{kind: "field"})
	f := db.Fields{
		FieldID:      p.FieldID,
		ParentID:     p.ParentID,
		Label:        p.Label,
		Data:         p.Data,
		Type:         p.Type,
		AuthorID:     p.AuthorID,
		DateCreated:  p.DateCreated,
		DateModified: p.DateModified,
	}
	m.fields = append(m.fields, f)
	return f
}

func (m *mockInstaller) CreateDatatype(p db.CreateDatatypeParams) db.Datatypes {
	m.calls = append(m.calls, callRecord{kind: "datatype"})
	d := db.Datatypes{
		DatatypeID:   p.DatatypeID,
		ParentID:     p.ParentID,
		Label:        p.Label,
		Type:         p.Type,
		AuthorID:     p.AuthorID,
		DateCreated:  p.DateCreated,
		DateModified: p.DateModified,
	}
	m.datatypes = append(m.datatypes, d)
	return d
}

func (m *mockInstaller) CreateDatatypeField(p db.CreateDatatypeFieldParams) db.DatatypeFields {
	m.calls = append(m.calls, callRecord{kind: "junction"})
	j := db.DatatypeFields{
		ID:         p.ID,
		DatatypeID: p.DatatypeID,
		FieldID:    p.FieldID,
	}
	m.junctions = append(m.junctions, j)
	return j
}

func TestInstall_DefaultSchema(t *testing.T) {
	def, ok := Get("modulacms-default")
	if !ok {
		t.Fatal("modulacms-default definition not found")
	}

	mock := &mockInstaller{}
	authorID := types.NewUserID()
	result, err := Install(mock, def, authorID)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if result.DefinitionName != "modulacms-default" {
		t.Errorf("expected definition name %q, got %q", "modulacms-default", result.DefinitionName)
	}

	expectedFields := len(def.Fields)
	if result.Fields != expectedFields {
		t.Errorf("expected %d fields, got %d", expectedFields, result.Fields)
	}

	expectedDatatypes := len(def.Datatypes)
	if result.Datatypes != expectedDatatypes {
		t.Errorf("expected %d datatypes, got %d", expectedDatatypes, result.Datatypes)
	}

	// Count total field refs across all datatypes
	expectedJunctions := 0
	for _, dt := range def.Datatypes {
		expectedJunctions += len(dt.FieldRefs)
	}
	if result.JunctionLinks != expectedJunctions {
		t.Errorf("expected %d junction links, got %d", expectedJunctions, result.JunctionLinks)
	}
}

func TestInstall_PhaseOrdering(t *testing.T) {
	def, ok := Get("modulacms-default")
	if !ok {
		t.Fatal("modulacms-default definition not found")
	}

	mock := &mockInstaller{}
	authorID := types.NewUserID()
	_, err := Install(mock, def, authorID)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify ordering: all fields first, then datatypes, then junctions
	phase := "field"
	for _, call := range mock.calls {
		switch phase {
		case "field":
			if call.kind == "datatype" {
				phase = "datatype"
			} else if call.kind == "junction" {
				t.Fatal("junction created before all datatypes")
			}
		case "datatype":
			if call.kind == "field" {
				t.Fatal("field created after datatype phase started")
			}
			if call.kind == "junction" {
				phase = "junction"
			}
		case "junction":
			if call.kind == "field" {
				t.Fatal("field created during junction phase")
			}
			if call.kind == "datatype" {
				t.Fatal("datatype created during junction phase")
			}
		}
	}
}

func TestInstall_ChildDatatypeReceivesParentID(t *testing.T) {
	def, ok := Get("modulacms-default")
	if !ok {
		t.Fatal("modulacms-default definition not found")
	}

	mock := &mockInstaller{}
	authorID := types.NewUserID()
	_, err := Install(mock, def, authorID)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// The "section" datatype should have a non-empty ParentID
	foundSection := false
	for _, dt := range mock.datatypes {
		if dt.Label == "Section" {
			foundSection = true
			if !dt.ParentID.Valid {
				t.Error("section datatype should have a valid ParentID")
			}
		}
	}
	if !foundSection {
		t.Error("section datatype not found in created datatypes")
	}

	// The "page" datatype should have no parent
	for _, dt := range mock.datatypes {
		if dt.Label == "Page" {
			if dt.ParentID.Valid {
				t.Error("page datatype should not have a ParentID")
			}
		}
	}
}

func TestInstall_EmptyAuthorID(t *testing.T) {
	def, ok := Get("modulacms-default")
	if !ok {
		t.Fatal("modulacms-default definition not found")
	}

	mock := &mockInstaller{}
	_, err := Install(mock, def, types.UserID(""))
	if err == nil {
		t.Fatal("expected error for empty authorID")
	}
	if !strings.Contains(err.Error(), "authorID cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInstall_InvalidDefinition(t *testing.T) {
	mock := &mockInstaller{}
	authorID := types.NewUserID()

	invalid := SchemaDefinition{
		Name: "", // will fail validation
	}
	_, err := Install(mock, invalid, authorID)
	if err == nil {
		t.Fatal("expected error for invalid definition")
	}
}

func TestInstall_AllRegisteredDefinitions(t *testing.T) {
	for _, def := range List() {
		t.Run(def.Name, func(t *testing.T) {
			mock := &mockInstaller{}
			authorID := types.NewUserID()
			result, err := Install(mock, def, authorID)
			if err != nil {
				t.Fatalf("Install(%q) failed: %v", def.Name, err)
			}

			if result.Fields != len(def.Fields) {
				t.Errorf("expected %d fields, got %d", len(def.Fields), result.Fields)
			}
			if result.Datatypes != len(def.Datatypes) {
				t.Errorf("expected %d datatypes, got %d", len(def.Datatypes), result.Datatypes)
			}

			expectedJunctions := 0
			for _, dt := range def.Datatypes {
				expectedJunctions += len(dt.FieldRefs)
			}
			if result.JunctionLinks != expectedJunctions {
				t.Errorf("expected %d junctions, got %d", expectedJunctions, result.JunctionLinks)
			}
		})
	}
}
