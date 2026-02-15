package definitions

import (
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func validBase() SchemaDefinition {
	return SchemaDefinition{
		Name:        "test-schema",
		Label:       "Test Schema",
		Description: "A test schema",
		Format:      "test",
		Fields: map[string]FieldDef{
			"title": {Label: "Title", Type: types.FieldTypeText},
		},
		Datatypes: map[string]DatatypeDef{
			"page": {
				Label:     "Page",
				Type:      "page",
				FieldRefs: []string{"title"},
			},
		},
		RootKeys: []string{"page"},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*SchemaDefinition)
		wantErr string
	}{
		{
			name:    "valid definition passes",
			modify:  func(d *SchemaDefinition) {},
			wantErr: "",
		},
		{
			name:    "empty name",
			modify:  func(d *SchemaDefinition) { d.Name = "" },
			wantErr: "name cannot be empty",
		},
		{
			name:    "no datatypes",
			modify:  func(d *SchemaDefinition) { d.Datatypes = map[string]DatatypeDef{} },
			wantErr: "must have at least one datatype",
		},
		{
			name:    "no root keys",
			modify:  func(d *SchemaDefinition) { d.RootKeys = nil },
			wantErr: "must have at least one root key",
		},
		{
			name:    "root key references missing datatype",
			modify:  func(d *SchemaDefinition) { d.RootKeys = []string{"missing"} },
			wantErr: "root key \"missing\" not found",
		},
		{
			name: "field ref references missing field",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label:     "Page",
					Type:      "page",
					FieldRefs: []string{"nonexistent"},
				}
			},
			wantErr: "references unknown field \"nonexistent\"",
		},
		{
			name: "child ref references missing datatype",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label:     "Page",
					Type:      "page",
					FieldRefs: []string{"title"},
					ChildRefs: []string{"ghost"},
				}
			},
			wantErr: "references unknown child datatype \"ghost\"",
		},
		{
			name: "self-referencing child",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label:     "Page",
					Type:      "page",
					FieldRefs: []string{"title"},
					ChildRefs: []string{"page"},
				}
			},
			wantErr: "self-reference",
		},
		{
			name: "empty field label",
			modify: func(d *SchemaDefinition) {
				d.Fields["title"] = FieldDef{Label: "", Type: types.FieldTypeText}
			},
			wantErr: "has empty label",
		},
		{
			name: "invalid field type",
			modify: func(d *SchemaDefinition) {
				d.Fields["title"] = FieldDef{Label: "Title", Type: types.FieldType("invalid")}
			},
			wantErr: "has invalid type",
		},
		{
			name: "empty datatype label",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label:     "",
					Type:      "page",
					FieldRefs: []string{"title"},
				}
			},
			wantErr: "has empty label",
		},
		{
			name: "empty datatype type",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label:     "Page",
					Type:      "",
					FieldRefs: []string{"title"},
				}
			},
			wantErr: "has empty type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			def := validBase()
			tc.modify(&def)
			err := Validate(def)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tc.wantErr, err)
			}
		})
	}
}
