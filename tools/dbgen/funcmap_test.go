package main

import (
	"testing"
	"text/template"
)

// helper to extract a func from the FuncMap by name and cast it.
func getFuncMapEntry[T any](t *testing.T, name string) T {
	t.Helper()
	fm := templateFuncMap()
	fn, ok := fm[name]
	if !ok {
		t.Fatalf("templateFuncMap() missing key %q", name)
	}
	typed, ok := fn.(T)
	if !ok {
		t.Fatalf("templateFuncMap()[%q] has unexpected type %T", name, fn)
	}
	return typed
}

func TestTemplateFuncMap_ReturnsAllExpectedKeys(t *testing.T) {
	t.Parallel()
	fm := templateFuncMap()

	expected := []string{
		"lower",
		"driverEntity",
		"driverLabel",
		"driverCommentLabel",
		"paginationCast",
		"idFieldInGetParams",
		"isNullWrapper",
		"stringExpr",
		"wrapParam",
		"sqlcExtraQueryName",
		"sqlcPaginatedQueryName",
	}

	for _, key := range expected {
		if _, ok := fm[key]; !ok {
			t.Errorf("templateFuncMap() missing expected key %q", key)
		}
	}
}

func TestTemplateFuncMap_IsValidForTemplate(t *testing.T) {
	t.Parallel()
	// Verify the FuncMap can be used to create a template without panicking
	fm := templateFuncMap()
	tmpl := template.New("test").Funcs(fm)
	if tmpl == nil {
		t.Fatal("template.New().Funcs() returned nil")
	}
}

// ========== lower ==========

func TestLower(t *testing.T) {
	t.Parallel()
	lower := getFuncMapEntry[func(string) string](t, "lower")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "uppercase first char", input: "User", want: "user"},
		{name: "already lowercase", input: "user", want: "user"},
		{name: "single char uppercase", input: "U", want: "u"},
		{name: "single char lowercase", input: "u", want: "u"},
		{name: "empty string", input: "", want: ""},
		{name: "all uppercase", input: "URL", want: "uRL"},
		{name: "unicode first char", input: "Aabc", want: "aabc"},
		{name: "preserves rest", input: "MediaID", want: "mediaID"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := lower(tt.input)
			if got != tt.want {
				t.Errorf("lower(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ========== driverEntity ==========

func TestDriverEntity(t *testing.T) {
	t.Parallel()
	driverEntity := getFuncMapEntry[func(Entity, DriverConfig) DriverEntityData](t, "driverEntity")

	e := Entity{Name: "Users", Singular: "User"}
	d := DriverConfig{Name: "sqlite", Struct: "Database"}

	got := driverEntity(e, d)
	if got.Entity.Name != "Users" {
		t.Errorf("driverEntity().Entity.Name = %q, want %q", got.Entity.Name, "Users")
	}
	if got.Driver.Name != "sqlite" {
		t.Errorf("driverEntity().Driver.Name = %q, want %q", got.Driver.Name, "sqlite")
	}
}

// ========== driverLabel ==========

func TestDriverLabel(t *testing.T) {
	t.Parallel()
	driverLabel := getFuncMapEntry[func(DriverConfig) string](t, "driverLabel")

	tests := []struct {
		name       string
		driverName string
		want       string
	}{
		{name: "sqlite", driverName: "sqlite", want: "SQLITE"},
		{name: "mysql", driverName: "mysql", want: "MYSQL"},
		{name: "psql", driverName: "psql", want: "POSTGRES"},
		{name: "unknown driver", driverName: "cockroach", want: "COCKROACH"},
		{name: "empty driver", driverName: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := driverLabel(DriverConfig{Name: tt.driverName})
			if got != tt.want {
				t.Errorf("driverLabel(%q) = %q, want %q", tt.driverName, got, tt.want)
			}
		})
	}
}

// ========== driverCommentLabel ==========

func TestDriverCommentLabel(t *testing.T) {
	t.Parallel()
	driverCommentLabel := getFuncMapEntry[func(DriverConfig) string](t, "driverCommentLabel")

	tests := []struct {
		name       string
		driverName string
		want       string
	}{
		{name: "sqlite", driverName: "sqlite", want: "SQLite"},
		{name: "mysql", driverName: "mysql", want: "MySQL"},
		{name: "psql", driverName: "psql", want: "PostgreSQL"},
		{name: "unknown driver returns raw name", driverName: "cockroach", want: "cockroach"},
		{name: "empty driver", driverName: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := driverCommentLabel(DriverConfig{Name: tt.driverName})
			if got != tt.want {
				t.Errorf("driverCommentLabel(%q) = %q, want %q", tt.driverName, got, tt.want)
			}
		})
	}
}

// ========== paginationCast ==========

func TestPaginationCast(t *testing.T) {
	t.Parallel()
	paginationCast := getFuncMapEntry[func(DriverConfig, string) string](t, "paginationCast")

	tests := []struct {
		name            string
		int32Pagination bool
		value           string
		want            string
	}{
		{
			name:            "wraps in int32 when Int32Pagination is true",
			int32Pagination: true,
			value:           "params.Limit",
			want:            "int32(params.Limit)",
		},
		{
			name:            "returns value as-is when Int32Pagination is false",
			int32Pagination: false,
			value:           "params.Limit",
			want:            "params.Limit",
		},
		{
			name:            "empty value with int32",
			int32Pagination: true,
			value:           "",
			want:            "int32()",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := DriverConfig{Int32Pagination: tt.int32Pagination}
			got := paginationCast(d, tt.value)
			if got != tt.want {
				t.Errorf("paginationCast(%v, %q) = %q, want %q", tt.int32Pagination, tt.value, got, tt.want)
			}
		})
	}
}

// ========== idFieldInGetParams ==========

func TestIdFieldInGetParams(t *testing.T) {
	t.Parallel()
	idFieldInGetParams := getFuncMapEntry[func(Entity) string](t, "idFieldInGetParams")

	e := Entity{IDField: "UserID"}
	got := idFieldInGetParams(e)
	if got != "UserID" {
		t.Errorf("idFieldInGetParams() = %q, want %q", got, "UserID")
	}
}

// ========== isNullWrapper ==========

func TestIsNullWrapper(t *testing.T) {
	t.Parallel()
	isNullWrapper := getFuncMapEntry[func(string) bool](t, "isNullWrapper")

	tests := []struct {
		name string
		typ  string
		want bool
	}{
		{name: "NullString", typ: "NullString", want: true},
		{name: "NullInt32", typ: "NullInt32", want: true},
		{name: "NullInt64", typ: "NullInt64", want: true},
		{name: "NullTime", typ: "NullTime", want: true},
		{name: "string is not a null wrapper", typ: "string", want: false},
		{name: "sql.NullString is not matched", typ: "sql.NullString", want: false},
		{name: "empty string", typ: "", want: false},
		{name: "NullBool is not a wrapper", typ: "NullBool", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isNullWrapper(tt.typ)
			if got != tt.want {
				t.Errorf("isNullWrapper(%q) = %v, want %v", tt.typ, got, tt.want)
			}
		})
	}
}

// ========== stringExpr ==========

func TestStringExpr(t *testing.T) {
	t.Parallel()
	stringExpr := getFuncMapEntry[func(string, string) string](t, "stringExpr")

	tests := []struct {
		name    string
		convert string
		expr    string
		want    string
	}{
		{
			name:    "toString appends .String()",
			convert: "toString",
			expr:    "a.UserID",
			want:    "a.UserID.String()",
		},
		{
			name:    "string returns expr as-is",
			convert: "string",
			expr:    "a.Username",
			want:    "a.Username",
		},
		{
			name:    "sprintf wraps in Sprintf with %d",
			convert: "sprintf",
			expr:    "a.Count",
			want:    `fmt.Sprintf("%d", a.Count)`,
		},
		{
			name:    "cast wraps in string()",
			convert: "cast",
			expr:    "a.Byte",
			want:    "string(a.Byte)",
		},
		{
			name:    "nullToString wraps in utility.NullToString",
			convert: "nullToString",
			expr:    "a.Name",
			want:    "utility.NullToString(a.Name)",
		},
		{
			name:    "wrapperNullToString accesses .NullString",
			convert: "wrapperNullToString",
			expr:    "a.Name",
			want:    "utility.NullToString(a.Name.NullString)",
		},
		{
			name:    "nullToEmpty wraps in NullStringToEmpty",
			convert: "nullToEmpty",
			expr:    "a.Bio",
			want:    "NullStringToEmpty(a.Bio)",
		},
		{
			name:    "wrapperNullToEmpty accesses .NullString",
			convert: "wrapperNullToEmpty",
			expr:    "a.Bio",
			want:    "NullStringToEmpty(a.Bio.NullString)",
		},
		{
			name:    "wrapperNullInt64ToString accesses .NullInt64",
			convert: "wrapperNullInt64ToString",
			expr:    "a.Count",
			want:    "utility.NullToString(a.Count.NullInt64)",
		},
		{
			name:    "nullableIDToEmpty wraps in nullableIDToEmpty",
			convert: "nullableIDToEmpty",
			expr:    "a.AuthorID",
			want:    "nullableIDToEmpty(a.AuthorID)",
		},
		{
			name:    "sprintfBool wraps in Sprintf with %t",
			convert: "sprintfBool",
			expr:    "a.Active",
			want:    `fmt.Sprintf("%t", a.Active)`,
		},
		{
			name:    "sprintfFloat64 wraps in Sprintf with .Float64 access",
			convert: "sprintfFloat64",
			expr:    "a.FocalX",
			want:    `fmt.Sprintf("%v", a.FocalX.Float64)`,
		},
		{
			name:    "unknown convert returns expr as-is",
			convert: "unknownConversion",
			expr:    "a.Whatever",
			want:    "a.Whatever",
		},
		{
			name:    "empty convert returns expr as-is",
			convert: "",
			expr:    "a.Skipped",
			want:    "a.Skipped",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stringExpr(tt.convert, tt.expr)
			if got != tt.want {
				t.Errorf("stringExpr(%q, %q) = %q, want %q", tt.convert, tt.expr, got, tt.want)
			}
		})
	}
}

// ========== wrapParam ==========

func TestWrapParam(t *testing.T) {
	t.Parallel()
	wrapParam := getFuncMapEntry[func(string, string) string](t, "wrapParam")

	tests := []struct {
		name     string
		wrapExpr string
		value    string
		want     string
	}{
		{
			name:     "applies wrapping expression",
			wrapExpr: "StringToNullString(%s)",
			value:    "name",
			want:     "StringToNullString(name)",
		},
		{
			name:     "empty wrapExpr returns value as-is",
			wrapExpr: "",
			value:    "email",
			want:     "email",
		},
		{
			name:     "wrapExpr without %s returns wrapExpr unchanged",
			wrapExpr: "SomeFunc(hardcoded)",
			value:    "ignored",
			want:     "SomeFunc(hardcoded)",
		},
		{
			name:     "only first %s is replaced",
			wrapExpr: "Func(%s, %s)",
			value:    "x",
			want:     "Func(x, %s)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := wrapParam(tt.wrapExpr, tt.value)
			if got != tt.want {
				t.Errorf("wrapParam(%q, %q) = %q, want %q", tt.wrapExpr, tt.value, got, tt.want)
			}
		})
	}
}

// ========== sqlcExtraQueryName (FuncMap wrapper) ==========

func TestFuncMap_SqlcExtraQueryName(t *testing.T) {
	t.Parallel()
	fn := getFuncMapEntry[func(Entity, ExtraQuery) string](t, "sqlcExtraQueryName")

	// This is a thin wrapper around Entity.SqlcExtraQueryName -- just verify it delegates
	e := Entity{}
	eq := ExtraQuery{MethodName: "GetByEmail", SqlcName: "GetUserByEmail"}
	got := fn(e, eq)
	if got != "GetUserByEmail" {
		t.Errorf("sqlcExtraQueryName() = %q, want %q", got, "GetUserByEmail")
	}
}

// ========== sqlcPaginatedQueryName (FuncMap wrapper) ==========

func TestFuncMap_SqlcPaginatedQueryName(t *testing.T) {
	t.Parallel()
	fn := getFuncMapEntry[func(Entity, PaginatedExtraQuery) string](t, "sqlcPaginatedQueryName")

	e := Entity{}
	pq := PaginatedExtraQuery{MethodName: "ListByRoute", SqlcName: "ListContentByRoutePaginated"}
	got := fn(e, pq)
	if got != "ListContentByRoutePaginated" {
		t.Errorf("sqlcPaginatedQueryName() = %q, want %q", got, "ListContentByRoutePaginated")
	}
}
