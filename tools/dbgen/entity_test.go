package main

import (
	"testing"
)

// ========== Field ==========

func TestField_SqlcFieldName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		f    Field
		want string
	}{
		{
			name: "returns SqlcName when set",
			f:    Field{AppName: "Role", SqlcName: "Roles"},
			want: "Roles",
		},
		{
			name: "falls back to AppName when SqlcName empty",
			f:    Field{AppName: "Username", SqlcName: ""},
			want: "Username",
		},
		{
			name: "both empty",
			f:    Field{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.f.SqlcFieldName()
			if got != tt.want {
				t.Errorf("SqlcFieldName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ========== Entity query name methods ==========

func TestEntity_CreateTableQueryName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		e    Entity
		want string
	}{
		{
			name: "uses override when set",
			e:    Entity{Singular: "Field", SqlcCreateTableName: "CreateDatatypesFieldsTable"},
			want: "CreateDatatypesFieldsTable",
		},
		{
			name: "derives from singular when no override",
			e:    Entity{Singular: "User"},
			want: "CreateUserTable",
		},
		{
			name: "derives from singular Media",
			e:    Entity{Singular: "Media"},
			want: "CreateMediaTable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.e.CreateTableQueryName()
			if got != tt.want {
				t.Errorf("CreateTableQueryName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEntity_CountQueryName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		e    Entity
		want string
	}{
		{
			name: "uses override when set",
			e:    Entity{Singular: "AdminRoute", SqlcCountName: "CountAdminroute"},
			want: "CountAdminroute",
		},
		{
			name: "derives from singular when no override",
			e:    Entity{Singular: "User"},
			want: "CountUser",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.e.CountQueryName()
			if got != tt.want {
				t.Errorf("CountQueryName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEntity_GetQueryName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		e    Entity
		want string
	}{
		{
			name: "uses override when set",
			e:    Entity{Singular: "DatatypeField", SqlcGetName: "GetDatatypeField"},
			want: "GetDatatypeField",
		},
		{
			name: "derives from singular when no override",
			e:    Entity{Singular: "Token"},
			want: "GetToken",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.e.GetQueryName()
			if got != tt.want {
				t.Errorf("GetQueryName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEntity_ListQueryName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		e    Entity
		want string
	}{
		{
			name: "uses override when set",
			e:    Entity{Singular: "DatatypeField", SqlcListName: "ListDatatypeField"},
			want: "ListDatatypeField",
		},
		{
			name: "derives from singular when no override",
			e:    Entity{Singular: "Table"},
			want: "ListTable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.e.ListQueryName()
			if got != tt.want {
				t.Errorf("ListQueryName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEntity_PaginatedListQueryName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		e    Entity
		want string
	}{
		{
			name: "uses override when set",
			e:    Entity{Singular: "ContentField", SqlcListPaginatedName: "ListContentFieldsPaginated"},
			want: "ListContentFieldsPaginated",
		},
		{
			name: "derives from singular when no override",
			e:    Entity{Singular: "Media"},
			want: "ListMediaPaginated",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.e.PaginatedListQueryName()
			if got != tt.want {
				t.Errorf("PaginatedListQueryName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ========== Entity field filter methods ==========

func TestEntity_StructFields(t *testing.T) {
	t.Parallel()
	e := Entity{
		Fields: []Field{
			{AppName: "ID", UpdateParamsOnly: false},
			{AppName: "Name", UpdateParamsOnly: false},
			{AppName: "Status", UpdateParamsOnly: true},
		},
	}
	got := e.StructFields()
	if len(got) != 2 {
		t.Fatalf("StructFields() returned %d fields, want 2", len(got))
	}
	if got[0].AppName != "ID" {
		t.Errorf("StructFields()[0].AppName = %q, want %q", got[0].AppName, "ID")
	}
	if got[1].AppName != "Name" {
		t.Errorf("StructFields()[1].AppName = %q, want %q", got[1].AppName, "Name")
	}
}

func TestEntity_StructFields_Empty(t *testing.T) {
	t.Parallel()
	e := Entity{}
	got := e.StructFields()
	if got != nil {
		t.Errorf("StructFields() on empty entity = %v, want nil", got)
	}
}

func TestEntity_StructFields_AllUpdateParamsOnly(t *testing.T) {
	t.Parallel()
	// When all fields are UpdateParamsOnly, StructFields returns nil
	e := Entity{
		Fields: []Field{
			{AppName: "A", UpdateParamsOnly: true},
			{AppName: "B", UpdateParamsOnly: true},
		},
	}
	got := e.StructFields()
	if got != nil {
		t.Errorf("StructFields() when all UpdateParamsOnly = %v, want nil", got)
	}
}

func TestEntity_CreateFields(t *testing.T) {
	t.Parallel()
	e := Entity{
		Fields: []Field{
			{AppName: "ID", IsPrimaryID: true, InCreate: false},
			{AppName: "Name", InCreate: true},
			{AppName: "Email", InCreate: true},
			{AppName: "ReadOnly", InCreate: false},
		},
	}
	got := e.CreateFields()
	if len(got) != 2 {
		t.Fatalf("CreateFields() returned %d fields, want 2", len(got))
	}
	if got[0].AppName != "Name" || got[1].AppName != "Email" {
		t.Errorf("CreateFields() = [%q, %q], want [Name, Email]", got[0].AppName, got[1].AppName)
	}
}

func TestEntity_CreateFields_Empty(t *testing.T) {
	t.Parallel()
	e := Entity{}
	got := e.CreateFields()
	if got != nil {
		t.Errorf("CreateFields() on empty entity = %v, want nil", got)
	}
}

func TestEntity_UpdateFields(t *testing.T) {
	t.Parallel()
	e := Entity{
		Fields: []Field{
			{AppName: "ID", IsPrimaryID: true, InUpdate: true},
			{AppName: "Name", InUpdate: true},
			{AppName: "CreatedAt", InUpdate: false},
		},
	}
	got := e.UpdateFields()
	if len(got) != 2 {
		t.Fatalf("UpdateFields() returned %d fields, want 2", len(got))
	}
	if got[0].AppName != "ID" || got[1].AppName != "Name" {
		t.Errorf("UpdateFields() = [%q, %q], want [ID, Name]", got[0].AppName, got[1].AppName)
	}
}

func TestEntity_NonIDCreateFields(t *testing.T) {
	t.Parallel()
	e := Entity{
		Fields: []Field{
			{AppName: "ID", IsPrimaryID: true, InCreate: true},
			{AppName: "Name", InCreate: true},
			{AppName: "Email", InCreate: true},
		},
	}
	got := e.NonIDCreateFields()
	if len(got) != 2 {
		t.Fatalf("NonIDCreateFields() returned %d fields, want 2", len(got))
	}
	for _, f := range got {
		if f.IsPrimaryID {
			t.Errorf("NonIDCreateFields() included primary ID field %q", f.AppName)
		}
	}
}

func TestEntity_NonIDUpdateFields(t *testing.T) {
	t.Parallel()
	e := Entity{
		Fields: []Field{
			{AppName: "ID", IsPrimaryID: true, InUpdate: true},
			{AppName: "Name", InUpdate: true},
			{AppName: "ReadOnly", InUpdate: false},
		},
	}
	got := e.NonIDUpdateFields()
	if len(got) != 1 {
		t.Fatalf("NonIDUpdateFields() returned %d fields, want 1", len(got))
	}
	if got[0].AppName != "Name" {
		t.Errorf("NonIDUpdateFields()[0].AppName = %q, want %q", got[0].AppName, "Name")
	}
}

// ========== Entity ID helpers ==========

func TestEntity_IDIsTyped(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		idType string
		want   bool
	}{
		{name: "typed ID", idType: "types.UserID", want: true},
		{name: "plain string", idType: "string", want: false},
		{name: "empty string is not string literal", idType: "", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := Entity{IDType: tt.idType}
			got := e.IDIsTyped()
			if got != tt.want {
				t.Errorf("IDIsTyped() with IDType=%q = %v, want %v", tt.idType, got, tt.want)
			}
		})
	}
}

func TestEntity_IDToString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		idType string
		expr   string
		want   string
	}{
		{
			name:   "typed ID wraps in string()",
			idType: "types.UserID",
			expr:   "u.UserID",
			want:   "string(u.UserID)",
		},
		{
			name:   "plain string returns expr as-is",
			idType: "string",
			expr:   "u.ID",
			want:   "u.ID",
		},
		{
			name:   "empty expr with typed ID",
			idType: "types.MediaID",
			expr:   "",
			want:   "string()",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := Entity{IDType: tt.idType}
			got := e.IDToString(tt.expr)
			if got != tt.want {
				t.Errorf("IDToString(%q) = %q, want %q", tt.expr, got, tt.want)
			}
		})
	}
}

// ========== Entity import detection ==========

func TestEntity_NeedsTypesImport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		fields []Field
		want   bool
	}{
		{
			name:   "field with types. prefix",
			fields: []Field{{Type: "types.UserID"}},
			want:   true,
		},
		{
			name:   "no types. prefix",
			fields: []Field{{Type: "string"}, {Type: "int64"}},
			want:   false,
		},
		{
			name:   "empty fields",
			fields: nil,
			want:   false,
		},
		{
			name: "short type name that starts with types but is not types.",
			// "types" is only 5 chars, needs > 6 and prefix "types."
			fields: []Field{{Type: "types"}},
			want:   false,
		},
		{
			name:   "exactly types. with something after",
			fields: []Field{{Type: "types.X"}},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := Entity{Fields: tt.fields}
			got := e.NeedsTypesImport()
			if got != tt.want {
				t.Errorf("NeedsTypesImport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntity_NeedsSqlImport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		fields []Field
		want   bool
	}{
		{
			name:   "field with sql. prefix",
			fields: []Field{{Type: "sql.NullString"}},
			want:   true,
		},
		{
			name:   "no sql. prefix",
			fields: []Field{{Type: "string"}, {Type: "NullString"}},
			want:   false,
		},
		{
			name:   "empty fields",
			fields: nil,
			want:   false,
		},
		{
			name:   "short type sql without dot",
			fields: []Field{{Type: "sql"}},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := Entity{Fields: tt.fields}
			got := e.NeedsSqlImport()
			if got != tt.want {
				t.Errorf("NeedsSqlImport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntity_NeedsUtilityImport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		fields []Field
		want   bool
	}{
		{
			name:   "nullToString triggers import",
			fields: []Field{{StringConvert: "nullToString"}},
			want:   true,
		},
		{
			name:   "wrapperNullToString triggers import",
			fields: []Field{{StringConvert: "wrapperNullToString"}},
			want:   true,
		},
		{
			name:   "wrapperNullInt64ToString triggers import",
			fields: []Field{{StringConvert: "wrapperNullInt64ToString"}},
			want:   true,
		},
		{
			name:   "toString does not trigger import",
			fields: []Field{{StringConvert: "toString"}},
			want:   false,
		},
		{
			name:   "empty StringConvert does not trigger import",
			fields: []Field{{StringConvert: ""}},
			want:   false,
		},
		{
			name:   "empty fields",
			fields: nil,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := Entity{Fields: tt.fields}
			got := e.NeedsUtilityImport()
			if got != tt.want {
				t.Errorf("NeedsUtilityImport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntity_NeedsFmtImport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		fields []Field
		want   bool
	}{
		{
			name:   "sprintf triggers import",
			fields: []Field{{StringConvert: "sprintf"}},
			want:   true,
		},
		{
			name:   "sprintfBool triggers import",
			fields: []Field{{StringConvert: "sprintfBool"}},
			want:   true,
		},
		{
			name:   "sprintfFloat64 triggers import",
			fields: []Field{{StringConvert: "sprintfFloat64"}},
			want:   true,
		},
		{
			name:   "string does not trigger import",
			fields: []Field{{StringConvert: "string"}},
			want:   false,
		},
		{
			name:   "empty fields",
			fields: nil,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := Entity{Fields: tt.fields}
			got := e.NeedsFmtImport()
			if got != tt.want {
				t.Errorf("NeedsFmtImport() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ========== Entity extra query name helpers ==========

func TestEntity_SqlcExtraQueryName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		eq   ExtraQuery
		want string
	}{
		{
			name: "uses SqlcName when set",
			eq:   ExtraQuery{MethodName: "GetMediaByURL", SqlcName: "GetMediaByUrl"},
			want: "GetMediaByUrl",
		},
		{
			name: "falls back to MethodName when SqlcName empty",
			eq:   ExtraQuery{MethodName: "GetUserByEmail", SqlcName: ""},
			want: "GetUserByEmail",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := Entity{}
			got := e.SqlcExtraQueryName(tt.eq)
			if got != tt.want {
				t.Errorf("SqlcExtraQueryName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEntity_SqlcPaginatedQueryName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		pq   PaginatedExtraQuery
		want string
	}{
		{
			name: "uses SqlcName when set",
			pq:   PaginatedExtraQuery{MethodName: "ListByRoute", SqlcName: "ListContentByRoutePaginated"},
			want: "ListContentByRoutePaginated",
		},
		{
			name: "falls back to MethodName when SqlcName empty",
			pq:   PaginatedExtraQuery{MethodName: "ListByStatus"},
			want: "ListByStatus",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := Entity{}
			got := e.SqlcPaginatedQueryName(tt.pq)
			if got != tt.want {
				t.Errorf("SqlcPaginatedQueryName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ========== Entity HasStringConvertFields ==========

func TestEntity_HasStringConvertFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		fields []Field
		want   bool
	}{
		{
			name:   "has fields with StringConvert",
			fields: []Field{{StringConvert: "toString"}, {StringConvert: ""}},
			want:   true,
		},
		{
			name:   "all fields empty StringConvert",
			fields: []Field{{StringConvert: ""}, {StringConvert: ""}},
			want:   false,
		},
		{
			name:   "no fields",
			fields: nil,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := Entity{Fields: tt.fields}
			got := e.HasStringConvertFields()
			if got != tt.want {
				t.Errorf("HasStringConvertFields() = %v, want %v", got, tt.want)
			}
		})
	}
}
