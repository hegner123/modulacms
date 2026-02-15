package db

import (
	"reflect"
	"testing"
)

func TestTableStructMap_Completeness(t *testing.T) {
	t.Parallel()
	// Every DBTable constant should have an entry in TableStructMap
	allTables := []DBTable{
		Admin_content_data,
		Admin_content_fields,
		Admin_datatype,
		Admin_datatype_fields,
		Admin_field,
		Admin_route,
		Content_data,
		Content_fields,
		Datatype_fields,
		Datatype,
		Field,
		MediaT,
		Media_dimension,
		Permission,
		Role,
		Route,
		Session,
		Table,
		Token,
		User,
		User_oauth,
	}

	for _, tbl := range allTables {
		t.Run(string(tbl), func(t *testing.T) {
			t.Parallel()
			typ, ok := TableStructMap[tbl]
			if !ok {
				t.Fatalf("TableStructMap missing entry for %q", tbl)
			}
			if typ == nil {
				t.Fatalf("TableStructMap[%q] is nil", tbl)
			}
			if typ.Kind() != reflect.Struct {
				t.Errorf("TableStructMap[%q] kind = %v, want struct", tbl, typ.Kind())
			}
		})
	}
}

func TestTableStructMap_CorrectTypes(t *testing.T) {
	t.Parallel()
	// Spot-check a few entries to ensure they map to the right type
	tests := []struct {
		table    DBTable
		wantType reflect.Type
	}{
		{Route, reflect.TypeOf(Routes{})},
		{User, reflect.TypeOf(Users{})},
		{Permission, reflect.TypeOf(Permissions{})},
		{Session, reflect.TypeOf(Sessions{})},
		{MediaT, reflect.TypeOf(Media{})},
		{Role, reflect.TypeOf(Roles{})},
		{Token, reflect.TypeOf(Tokens{})},
		{Table, reflect.TypeOf(Tables{})},
		{Content_data, reflect.TypeOf(ContentData{})},
		{Datatype, reflect.TypeOf(Datatypes{})},
	}
	for _, tt := range tests {
		t.Run(string(tt.table), func(t *testing.T) {
			t.Parallel()
			got := TableStructMap[tt.table]
			if got != tt.wantType {
				t.Errorf("TableStructMap[%q] = %v, want %v", tt.table, got, tt.wantType)
			}
		})
	}
}

func TestCastToTypedSlice_NilInput(t *testing.T) {
	t.Parallel()
	got := CastToTypedSlice(nil, Route)
	if got != nil {
		t.Errorf("CastToTypedSlice(nil, Route) = %v, want nil", got)
	}
}

func TestCastToTypedSlice_CorrectType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		table DBTable
		input any
	}{
		{name: "routes", table: Route, input: []Routes{{}}},
		{name: "users", table: User, input: []Users{{}}},
		{name: "permissions", table: Permission, input: []Permissions{{}}},
		{name: "sessions", table: Session, input: []Sessions{{}}},
		{name: "media", table: MediaT, input: []Media{{}}},
		{name: "roles", table: Role, input: []Roles{{}}},
		{name: "tokens", table: Token, input: []Tokens{{}}},
		{name: "tables", table: Table, input: []Tables{{}}},
		{name: "datatypes", table: Datatype, input: []Datatypes{{}}},
		{name: "fields", table: Field, input: []Fields{{}}},
		{name: "content_data", table: Content_data, input: []ContentData{{}}},
		{name: "content_fields", table: Content_fields, input: []ContentFields{{}}},
		{name: "admin_content_data", table: Admin_content_data, input: []AdminContentData{{}}},
		{name: "admin_content_fields", table: Admin_content_fields, input: []AdminContentFields{{}}},
		{name: "admin_datatypes", table: Admin_datatype, input: []AdminDatatypes{{}}},
		{name: "admin_datatype_fields", table: Admin_datatype_fields, input: []AdminDatatypeFields{{}}},
		{name: "admin_fields", table: Admin_field, input: []AdminFields{{}}},
		{name: "admin_routes", table: Admin_route, input: []AdminRoutes{{}}},
		{name: "media_dimensions", table: Media_dimension, input: []MediaDimensions{{}}},
		{name: "datatype_fields", table: Datatype_fields, input: []DatatypeFields{{}}},
		{name: "user_oauth", table: User_oauth, input: []UserOauth{{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CastToTypedSlice(tt.input, tt.table)
			if got == nil {
				t.Fatal("CastToTypedSlice returned nil for valid input")
			}
			// The returned value should be the same pointer (identity) as the input
			if reflect.TypeOf(got) != reflect.TypeOf(tt.input) {
				t.Errorf("type = %T, want %T", got, tt.input)
			}
		})
	}
}

func TestCastToTypedSlice_WrongType_ReturnsAsIs(t *testing.T) {
	t.Parallel()
	// When the type assertion fails, the function should return the input as-is
	input := []map[string]any{{"key": "value"}}
	got := CastToTypedSlice(input, Route)
	if got == nil {
		t.Fatal("expected non-nil return for unmatched type")
	}
	// Should be returned as-is
	if _, ok := got.([]map[string]any); !ok {
		t.Errorf("expected []map[string]any passthrough, got %T", got)
	}
}

func TestCastToTypedSlice_UnknownTable_ReturnsAsIs(t *testing.T) {
	t.Parallel()
	input := []string{"a", "b"}
	got := CastToTypedSlice(input, DBTable("nonexistent"))
	if got == nil {
		t.Fatal("expected non-nil return for unknown table")
	}
}
