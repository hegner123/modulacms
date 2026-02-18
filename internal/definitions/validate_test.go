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
		Datatypes: map[string]DatatypeDef{
			"page": {
				Label: "Page",
				Type:  types.NewNullableString("page"),
				FieldRefs: []FieldDef{
					{Label: "Title", Type: types.FieldTypeText},
				},
			},
		},
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
			name: "parent ref references missing datatype",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["child"] = DatatypeDef{
					Label:     "Child",
					Type:      types.NewNullableString("child"),
					ParentRef: "ghost",
				}
			},
			wantErr: "references unknown parent \"ghost\"",
		},
		{
			name: "self-referencing parent",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label:     "Page",
					Type:      types.NewNullableString("page"),
					ParentRef: "page",
					FieldRefs: []FieldDef{
						{Label: "Title", Type: types.FieldTypeText},
					},
				}
			},
			wantErr: "self-reference",
		},
		{
			name: "empty field label",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label: "Page",
					Type:  types.NewNullableString("page"),
					FieldRefs: []FieldDef{
						{Label: "", Type: types.FieldTypeText},
					},
				}
			},
			wantErr: "has empty label",
		},
		{
			name: "invalid field type",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label: "Page",
					Type:  types.NewNullableString("page"),
					FieldRefs: []FieldDef{
						{Label: "Title", Type: types.FieldType("invalid")},
					},
				}
			},
			wantErr: "has invalid type",
		},
		{
			name: "empty datatype label",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label: "",
					Type:  types.NewNullableString("page"),
					FieldRefs: []FieldDef{
						{Label: "Title", Type: types.FieldTypeText},
					},
				}
			},
			wantErr: "has empty label",
		},
		{
			name: "empty datatype type",
			modify: func(d *SchemaDefinition) {
				d.Datatypes["page"] = DatatypeDef{
					Label: "Page",
					Type:  types.NullableString{},
					FieldRefs: []FieldDef{
						{Label: "Title", Type: types.FieldTypeText},
					},
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
