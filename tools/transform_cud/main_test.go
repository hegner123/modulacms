package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// White-box test file: the interesting logic lives in unexported helpers
// (findFuncEnd, getWrapperType, transformCreate, transformUpdate,
// transformDelete, transformDatatypeFieldSortOrder, transformUserSshKeyLabel).
// Testing through main() alone would require massive fixture files and
// would not isolate failures to specific transform functions.

// ---------------------------------------------------------------------------
// findFuncEnd
// ---------------------------------------------------------------------------

func TestFindFuncEnd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		idx     int
		want    int
	}{
		{
			name:    "simple_function",
			content: `func foo() { return }`,
			idx:     0,
			want:    21,
		},
		{
			name:    "nested_braces",
			content: `func foo() { if true { x := 1 } }`,
			idx:     0,
			want:    33,
		},
		{
			name:    "deeply_nested",
			content: `func foo() { { { { } } } }`,
			idx:     0,
			want:    26,
		},
		{
			name: "multiline_function",
			content: `func foo() {
	x := 1
	if x > 0 {
		return x
	}
	return 0
}`,
			idx:  0,
			want: 58,
		},
		{
			name:    "string_with_braces",
			content: `func foo() { s := "hello { world }" }`,
			idx:     0,
			want:    37,
		},
		{
			name:    "raw_string_with_braces",
			content: "func foo() { s := `hello { world }` }",
			idx:     0,
			want:    37,
		},
		{
			name:    "no_opening_brace",
			content: `func foo()`,
			idx:     0,
			want:    -1,
		},
		{
			// Unclosed function -- depth never returns to zero
			name:    "unclosed_function",
			content: `func foo() { if true {`,
			idx:     0,
			want:    -1,
		},
		{
			name:    "idx_in_middle_of_content",
			content: `something before func bar() { return } after`,
			idx:     17, // points to "func bar"
			want:    38,
		},
		{
			name:    "escaped_quote_in_string",
			content: `func foo() { s := "he said \"hi\" { }" }`,
			idx:     0,
			want:    40,
		},
		{
			name:    "empty_function_body",
			content: `func foo() {}`,
			idx:     0,
			want:    13,
		},
		{
			name:    "raw_string_with_quotes_and_braces",
			content: "func foo() { s := `{\"key\": \"value\"}` }",
			idx:     0,
			want:    38,
		},
		{
			// Two functions back to back -- should find end of first
			name: "two_functions",
			content: `func first() {
	return
}
func second() {
	return
}`,
			idx:  0,
			want: 24,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := findFuncEnd(tt.content, tt.idx)
			if got != tt.want {
				t.Errorf("findFuncEnd(content, %d) = %d, want %d", tt.idx, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getWrapperType
// ---------------------------------------------------------------------------

func TestGetWrapperType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{name: "Route", want: "Routes"},
		{name: "Role", want: "Roles"},
		{name: "Permission", want: "Permissions"},
		{name: "Session", want: "Sessions"},
		{name: "Token", want: "Tokens"},
		{name: "Table", want: "Tables"},
		{name: "Field", want: "Fields"},
		{name: "Datatype", want: "Datatypes"},
		{name: "DatatypeField", want: "DatatypeFields"},
		{name: "ContentData", want: "ContentData"},
		{name: "ContentField", want: "ContentFields"},
		{name: "Media", want: "Media"},
		{name: "MediaDimension", want: "MediaDimensions"},
		{name: "AdminContentData", want: "AdminContentData"},
		{name: "AdminContentField", want: "AdminContentFields"},
		{name: "AdminDatatype", want: "AdminDatatypes"},
		{name: "AdminDatatypeField", want: "AdminDatatypeFields"},
		{name: "AdminField", want: "AdminFields"},
		{name: "AdminRoute", want: "AdminRoutes"},
		{name: "UserOauth", want: "UserOauth"},
		{name: "UserSshKey", want: "UserSshKeys"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := entity{name: tt.name}
			got := getWrapperType(e)
			if got != tt.want {
				t.Errorf("getWrapperType(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestGetWrapperType_UnknownFallsBackToName(t *testing.T) {
	t.Parallel()
	e := entity{name: "UnknownEntity"}
	got := getWrapperType(e)
	if got != "UnknownEntity" {
		t.Errorf("getWrapperType(UnknownEntity) = %q, want %q", got, "UnknownEntity")
	}
}

// ---------------------------------------------------------------------------
// transformCreate
// ---------------------------------------------------------------------------

func TestTransformCreate_BareReturn(t *testing.T) {
	t.Parallel()
	// Simulates a bare-return Create (e.g., Route) for the Database driver
	e := entity{
		file:              "route.go",
		name:              "Route",
		mapName:           "MapRoute",
		idType:            "types.RouteID",
		idField:           "RouteID",
		table:             "routes",
		createReturnsBare: true,
	}

	input := `package db

func (d Database) CreateRoute(s CreateRouteParams) Routes {
	queries := mdb.New(d.Connection)
	r, _ := queries.CreateRoute(d.Context, mdb.CreateRouteParams{
		Slug: s.Slug,
	})
	return d.MapRoute(r)
}
`

	content, modified := transformCreate(input, e, "Database", false)
	if !modified {
		t.Fatal("expected modified=true, got false")
	}
	if !strings.Contains(content, "func (d Database) CreateRoute(ctx context.Context, ac audited.AuditContext, s CreateRouteParams) (*Routes, error) {") {
		t.Error("expected new signature with ctx, ac, and (*Routes, error) return")
	}
	if !strings.Contains(content, "cmd := d.NewRouteCmd(ctx, ac, s)") {
		t.Error("expected NewRouteCmd call")
	}
	if !strings.Contains(content, "audited.Create(cmd)") {
		t.Error("expected audited.Create call")
	}
	if !strings.Contains(content, "d.MapRoute(result)") {
		t.Error("expected MapRoute call on result")
	}
}

func TestTransformCreate_PointerReturn(t *testing.T) {
	t.Parallel()
	// Session uses pointer return
	e := entity{
		file:              "session.go",
		name:              "Session",
		mapName:           "MapSession",
		idType:            "types.SessionID",
		idField:           "SessionID",
		table:             "sessions",
		createReturnsBare: false,
	}

	input := `package db

func (d Database) CreateSession(s CreateSessionParams) (*Sessions, error) {
	queries := mdb.New(d.Connection)
	r, err := queries.CreateSession(d.Context, mdb.CreateSessionParams{})
	if err != nil {
		return nil, err
	}
	result := d.MapSession(r)
	return &result, nil
}
`

	content, modified := transformCreate(input, e, "Database", false)
	if !modified {
		t.Fatal("expected modified=true, got false")
	}
	if !strings.Contains(content, "func (d Database) CreateSession(ctx context.Context, ac audited.AuditContext, s CreateSessionParams) (*Sessions, error) {") {
		t.Error("expected new signature")
	}
}

func TestTransformCreate_UserSshKeySpecialCase(t *testing.T) {
	t.Parallel()
	e := entity{
		file:              "user_ssh_keys.go",
		name:              "UserSshKey",
		mapName:           "MapUserSshKeys",
		idType:            "string",
		idField:           "SshKeyID",
		table:             "user_ssh_keys",
		createReturnsBare: false,
	}

	input := `package db

func (d Database) CreateUserSshKey(params CreateUserSshKeyParams) (*UserSshKeys, error) {
	queries := mdb.New(d.Connection)
	r, err := queries.CreateUserSshKey(d.Context, mdb.CreateUserSshKeyParams{})
	if err != nil {
		return nil, err
	}
	result := d.MapUserSshKeys(r)
	return &result, nil
}
`

	content, modified := transformCreate(input, e, "Database", false)
	if !modified {
		t.Fatal("expected modified=true, got false")
	}
	// user_ssh_keys uses "params" not "s"
	if !strings.Contains(content, "func (d Database) CreateUserSshKey(ctx context.Context, ac audited.AuditContext, params CreateUserSshKeyParams) (*UserSshKeys, error) {") {
		t.Error("expected user_ssh_keys special signature with 'params'")
	}
	if !strings.Contains(content, "NewUserSshKeyCmd(ctx, ac, params)") {
		t.Error("expected NewUserSshKeyCmd with params argument")
	}
}

func TestTransformCreate_PatternNotFound(t *testing.T) {
	t.Parallel()
	e := entity{
		file:              "route.go",
		name:              "Route",
		mapName:           "MapRoute",
		idType:            "types.RouteID",
		idField:           "RouteID",
		table:             "routes",
		createReturnsBare: true,
	}

	input := `package db

func (d Database) SomethingElse() error {
	return nil
}
`

	content, modified := transformCreate(input, e, "Database", false)
	if modified {
		t.Error("expected modified=false when pattern not found")
	}
	if content != input {
		t.Error("expected content unchanged when pattern not found")
	}
}

func TestTransformCreate_PreservesModifiedFlag(t *testing.T) {
	t.Parallel()
	// When prevModified is true and pattern not found, modified should stay true
	e := entity{
		file:              "route.go",
		name:              "Route",
		mapName:           "MapRoute",
		idType:            "types.RouteID",
		idField:           "RouteID",
		table:             "routes",
		createReturnsBare: true,
	}

	input := `package db

func unrelated() {}
`

	_, modified := transformCreate(input, e, "Database", true)
	if !modified {
		t.Error("expected prevModified=true to be preserved when no pattern found")
	}
}

func TestTransformCreate_AllDrivers(t *testing.T) {
	t.Parallel()
	drivers := []string{"Database", "MysqlDatabase", "PsqlDatabase"}
	e := entity{
		file:              "role.go",
		name:              "Role",
		mapName:           "MapRole",
		idType:            "types.RoleID",
		idField:           "RoleID",
		table:             "roles",
		createReturnsBare: true,
	}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			t.Parallel()
			input := "package db\n\nfunc (d " + driver + ") CreateRole(s CreateRoleParams) Roles {\n\treturn Roles{}\n}\n"
			content, modified := transformCreate(input, e, driver, false)
			if !modified {
				t.Fatalf("expected modified=true for driver %s", driver)
			}
			wantSig := "func (d " + driver + ") CreateRole(ctx context.Context, ac audited.AuditContext, s CreateRoleParams) (*Roles, error) {"
			if !strings.Contains(content, wantSig) {
				t.Errorf("expected signature %q in output for driver %s", wantSig, driver)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// transformUpdate
// ---------------------------------------------------------------------------

func TestTransformUpdate_Standard(t *testing.T) {
	t.Parallel()
	e := entity{
		name:               "Route",
		updateSuccessField: `s.Slug`,
	}

	input := `package db

func (d Database) UpdateRoute(s UpdateRouteParams) (*string, error) {
	queries := mdb.New(d.Connection)
	err := queries.UpdateRoute(d.Context, mdb.UpdateRouteParams{
		Slug: s.Slug,
	})
	if err != nil {
		return nil, err
	}
	msg := "updated"
	return &msg, nil
}
`

	content, modified := transformUpdate(input, e, "Database", false)
	if !modified {
		t.Fatal("expected modified=true")
	}
	if !strings.Contains(content, "func (d Database) UpdateRoute(ctx context.Context, ac audited.AuditContext, s UpdateRouteParams) (*string, error) {") {
		t.Error("expected new signature with ctx, ac")
	}
	if !strings.Contains(content, "cmd := d.UpdateRouteCmd(ctx, ac, s)") {
		t.Error("expected UpdateRouteCmd call")
	}
	if !strings.Contains(content, "audited.Update(cmd)") {
		t.Error("expected audited.Update call")
	}
	if !strings.Contains(content, "s.Slug") {
		t.Error("expected success field s.Slug in output")
	}
}

func TestTransformUpdate_EmptySuccessField(t *testing.T) {
	t.Parallel()
	// UserSshKey has empty updateSuccessField -- should fall back to "updated"
	e := entity{
		name:               "UserSshKey",
		updateSuccessField: "",
	}

	input := `package db

func (d Database) UpdateUserSshKey(s UpdateUserSshKeyParams) (*string, error) {
	return nil, nil
}
`

	content, modified := transformUpdate(input, e, "Database", false)
	if !modified {
		t.Fatal("expected modified=true")
	}
	// When updateSuccessField is empty, it uses `"updated"` literal
	if !strings.Contains(content, `"updated"`) {
		t.Error("expected fallback to \"updated\" when updateSuccessField is empty")
	}
}

func TestTransformUpdate_PatternNotFound(t *testing.T) {
	t.Parallel()
	e := entity{name: "Route", updateSuccessField: "s.Slug"}

	input := `package db

func unrelated() {}
`

	content, modified := transformUpdate(input, e, "Database", false)
	if modified {
		t.Error("expected modified=false when pattern not found")
	}
	if content != input {
		t.Error("expected content unchanged")
	}
}

// ---------------------------------------------------------------------------
// transformDelete
// ---------------------------------------------------------------------------

func TestTransformDelete_Standard(t *testing.T) {
	t.Parallel()
	e := entity{
		file:   "route.go",
		name:   "Route",
		idType: "types.RouteID",
	}

	input := `package db

func (d Database) DeleteRoute(id types.RouteID) error {
	queries := mdb.New(d.Connection)
	return queries.DeleteRoute(d.Context, string(id))
}
`

	content, modified := transformDelete(input, e, "Database", false)
	if !modified {
		t.Fatal("expected modified=true")
	}
	if !strings.Contains(content, "func (d Database) DeleteRoute(ctx context.Context, ac audited.AuditContext, id types.RouteID) error {") {
		t.Error("expected new signature with ctx, ac")
	}
	if !strings.Contains(content, "cmd := d.DeleteRouteCmd(ctx, ac, id)") {
		t.Error("expected DeleteRouteCmd call")
	}
	if !strings.Contains(content, "return audited.Delete(cmd)") {
		t.Error("expected audited.Delete return")
	}
}

func TestTransformDelete_UserSshKeySpecialCase(t *testing.T) {
	t.Parallel()
	e := entity{
		file:   "user_ssh_keys.go",
		name:   "UserSshKey",
		idType: "string",
	}

	input := `package db

func (d Database) DeleteUserSshKey(id string) error {
	queries := mdb.New(d.Connection)
	return queries.DeleteUserSshKey(d.Context, id)
}
`

	content, modified := transformDelete(input, e, "Database", false)
	if !modified {
		t.Fatal("expected modified=true")
	}
	if !strings.Contains(content, "func (d Database) DeleteUserSshKey(ctx context.Context, ac audited.AuditContext, id string) error {") {
		t.Error("expected user_ssh_keys special delete signature")
	}
	if !strings.Contains(content, "DeleteUserSshKeyCmd(ctx, ac, id)") {
		t.Error("expected DeleteUserSshKeyCmd call")
	}
}

func TestTransformDelete_PatternNotFound(t *testing.T) {
	t.Parallel()
	e := entity{file: "route.go", name: "Route", idType: "types.RouteID"}

	input := `package db

func unrelated() {}
`

	content, modified := transformDelete(input, e, "Database", false)
	if modified {
		t.Error("expected modified=false when pattern not found")
	}
	if content != input {
		t.Error("expected content unchanged")
	}
}

func TestTransformDelete_AllDrivers(t *testing.T) {
	t.Parallel()
	drivers := []string{"Database", "MysqlDatabase", "PsqlDatabase"}
	e := entity{file: "role.go", name: "Role", idType: "types.RoleID"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			t.Parallel()
			input := "package db\n\nfunc (d " + driver + ") DeleteRole(id types.RoleID) error {\n\treturn nil\n}\n"
			content, modified := transformDelete(input, e, driver, false)
			if !modified {
				t.Fatalf("expected modified=true for driver %s", driver)
			}
			wantSig := "func (d " + driver + ") DeleteRole(ctx context.Context, ac audited.AuditContext, id types.RoleID) error {"
			if !strings.Contains(content, wantSig) {
				t.Errorf("expected signature %q for driver %s", wantSig, driver)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// transformDatatypeFieldSortOrder
// ---------------------------------------------------------------------------

func TestTransformDatatypeFieldSortOrder_SQLite(t *testing.T) {
	t.Parallel()
	input := `package db

func (d Database) UpdateDatatypeFieldSortOrder(id string, sortOrder int64) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateDatatypeFieldSortOrder(d.Context, mdb.UpdateDatatypeFieldSortOrderParams{
		SortOrder: sortOrder,
		ID:        id,
	})
}
`

	content, modified := transformDatatypeFieldSortOrder(input, false)
	if !modified {
		t.Fatal("expected modified=true for SQLite variant")
	}
	if !strings.Contains(content, "func (d Database) UpdateDatatypeFieldSortOrder(ctx context.Context, ac audited.AuditContext, id string, sortOrder int64) error {") {
		t.Error("expected new SQLite signature")
	}
	if !strings.Contains(content, "UpdateDatatypeFieldSortOrderCmd(ctx, ac, id, sortOrder)") {
		t.Error("expected cmd factory call")
	}
}

func TestTransformDatatypeFieldSortOrder_MySQL(t *testing.T) {
	t.Parallel()
	input := `package db

func (d MysqlDatabase) UpdateDatatypeFieldSortOrder(id string, sortOrder int64) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateDatatypeFieldSortOrder(d.Context, mdbm.UpdateDatatypeFieldSortOrderParams{
		SortOrder: int32(sortOrder),
		ID:        id,
	})
}
`

	content, modified := transformDatatypeFieldSortOrder(input, false)
	if !modified {
		t.Fatal("expected modified=true for MySQL variant")
	}
	if !strings.Contains(content, "func (d MysqlDatabase) UpdateDatatypeFieldSortOrder(ctx context.Context, ac audited.AuditContext, id string, sortOrder int64) error {") {
		t.Error("expected new MySQL signature")
	}
}

func TestTransformDatatypeFieldSortOrder_PostgreSQL(t *testing.T) {
	t.Parallel()
	input := `package db

func (d PsqlDatabase) UpdateDatatypeFieldSortOrder(id string, sortOrder int64) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateDatatypeFieldSortOrder(d.Context, mdbp.UpdateDatatypeFieldSortOrderParams{
		SortOrder: int32(sortOrder),
		ID:        id,
	})
}
`

	content, modified := transformDatatypeFieldSortOrder(input, false)
	if !modified {
		t.Fatal("expected modified=true for PostgreSQL variant")
	}
	if !strings.Contains(content, "func (d PsqlDatabase) UpdateDatatypeFieldSortOrder(ctx context.Context, ac audited.AuditContext, id string, sortOrder int64) error {") {
		t.Error("expected new PostgreSQL signature")
	}
}

func TestTransformDatatypeFieldSortOrder_NoMatch(t *testing.T) {
	t.Parallel()
	input := `package db

func unrelated() {}
`

	content, modified := transformDatatypeFieldSortOrder(input, false)
	if modified {
		t.Error("expected modified=false when no patterns match")
	}
	if content != input {
		t.Error("expected content unchanged")
	}
}

func TestTransformDatatypeFieldSortOrder_PreservesModifiedFlag(t *testing.T) {
	t.Parallel()
	input := `package db

func unrelated() {}
`

	_, modified := transformDatatypeFieldSortOrder(input, true)
	if !modified {
		t.Error("expected prevModified=true to be preserved")
	}
}

// ---------------------------------------------------------------------------
// transformUserSshKeyLabel
// ---------------------------------------------------------------------------

func TestTransformUserSshKeyLabel_SQLite(t *testing.T) {
	t.Parallel()
	input := `package db

func (d Database) UpdateUserSshKeyLabel(id string, label string) error {
	queries := mdb.New(d.Connection)
	err := queries.UpdateUserSshKeyLabel(d.Context, mdb.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: label, Valid: label != ""},
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key label: %v", err)
	}
	return nil
}
`

	content, modified := transformUserSshKeyLabel(input, false)
	if !modified {
		t.Fatal("expected modified=true for SQLite variant")
	}
	if !strings.Contains(content, "func (d Database) UpdateUserSshKeyLabel(ctx context.Context, ac audited.AuditContext, id string, label string) error {") {
		t.Error("expected new SQLite signature")
	}
	if !strings.Contains(content, "UpdateUserSshKeyLabelCmd(ctx, ac, id, label)") {
		t.Error("expected cmd factory call")
	}
	if !strings.Contains(content, "return audited.Update(cmd)") {
		t.Error("expected audited.Update return")
	}
}

func TestTransformUserSshKeyLabel_MySQL(t *testing.T) {
	t.Parallel()
	input := `package db

func (d MysqlDatabase) UpdateUserSshKeyLabel(id string, label string) error {
	queries := mdbm.New(d.Connection)
	err := queries.UpdateUserSshKeyLabel(d.Context, mdbm.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: label, Valid: label != ""},
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key label: %v", err)
	}
	return nil
}
`

	content, modified := transformUserSshKeyLabel(input, false)
	if !modified {
		t.Fatal("expected modified=true for MySQL variant")
	}
	if !strings.Contains(content, "func (d MysqlDatabase) UpdateUserSshKeyLabel(ctx context.Context, ac audited.AuditContext, id string, label string) error {") {
		t.Error("expected new MySQL signature")
	}
}

func TestTransformUserSshKeyLabel_PostgreSQL(t *testing.T) {
	t.Parallel()
	input := `package db

func (d PsqlDatabase) UpdateUserSshKeyLabel(id string, label string) error {
	queries := mdbp.New(d.Connection)
	err := queries.UpdateUserSshKeyLabel(d.Context, mdbp.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: label, Valid: label != ""},
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key label: %v", err)
	}
	return nil
}
`

	content, modified := transformUserSshKeyLabel(input, false)
	if !modified {
		t.Fatal("expected modified=true for PostgreSQL variant")
	}
	if !strings.Contains(content, "func (d PsqlDatabase) UpdateUserSshKeyLabel(ctx context.Context, ac audited.AuditContext, id string, label string) error {") {
		t.Error("expected new PostgreSQL signature")
	}
}

func TestTransformUserSshKeyLabel_NoMatch(t *testing.T) {
	t.Parallel()
	input := `package db

func unrelated() {}
`

	content, modified := transformUserSshKeyLabel(input, false)
	if modified {
		t.Error("expected modified=false when no patterns match")
	}
	if content != input {
		t.Error("expected content unchanged")
	}
}

// ---------------------------------------------------------------------------
// Integration: main via exec.Command
// ---------------------------------------------------------------------------

// buildCUDBinary compiles the transform_cud binary into a temp directory.
func buildCUDBinary(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "transform_cud")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = filepath.Join(projectRoot(t), "tools", "transform_cud")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}
	return binPath
}

func projectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod)")
		}
		dir = parent
	}
}

func TestMainCUD_NoArgs_ExitNonZero(t *testing.T) {
	t.Parallel()
	bin := buildCUDBinary(t)

	cmd := exec.Command(bin)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit when no args provided, got exit 0")
	}
}

func TestMainCUD_NonexistentDir_ExitOne(t *testing.T) {
	t.Parallel()
	bin := buildCUDBinary(t)

	cmd := exec.Command(bin, "/nonexistent/path")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for nonexistent directory, got exit 0")
	}
	if !strings.Contains(string(out), "Error reading") {
		t.Errorf("expected error message about reading, got: %s", out)
	}
}

func TestMainCUD_ProcessesEntityFiles(t *testing.T) {
	t.Parallel()
	bin := buildCUDBinary(t)

	// Create a temp directory mimicking the internal/db/ structure
	// with a minimal route.go containing all three CUD methods for Database
	tmpDir := t.TempDir()

	routeContent := `package db

func (d Database) CreateRoute(s CreateRouteParams) Routes {
	queries := mdb.New(d.Connection)
	r, _ := queries.CreateRoute(d.Context, mdb.CreateRouteParams{
		Slug: s.Slug,
	})
	return d.MapRoute(r)
}

func (d Database) UpdateRoute(s UpdateRouteParams) (*string, error) {
	queries := mdb.New(d.Connection)
	err := queries.UpdateRoute(d.Context, mdb.UpdateRouteParams{
		Slug: s.Slug,
	})
	if err != nil {
		return nil, err
	}
	msg := "updated"
	return &msg, nil
}

func (d Database) DeleteRoute(id types.RouteID) error {
	queries := mdb.New(d.Connection)
	return queries.DeleteRoute(d.Context, string(id))
}
`

	routePath := filepath.Join(tmpDir, "route.go")
	if err := os.WriteFile(routePath, []byte(routeContent), 0644); err != nil {
		t.Fatalf("failed to write route.go: %v", err)
	}

	// Create empty files for all other entities the tool expects
	otherFiles := []string{
		"role.go", "permission.go", "session.go", "token.go", "table.go",
		"field.go", "datatype.go", "datatype_field.go", "content_data.go",
		"content_field.go", "media.go", "media_dimension.go",
		"admin_content_data.go", "admin_content_field.go", "admin_datatype.go",
		"admin_datatype_field.go", "admin_field.go", "admin_route.go",
		"user_oauth.go", "user_ssh_keys.go",
	}
	for _, f := range otherFiles {
		p := filepath.Join(tmpDir, f)
		if err := os.WriteFile(p, []byte("package db\n"), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", f, err)
		}
	}

	cmd := exec.Command(bin, tmpDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary exited with error: %v\noutput: %s", err, out)
	}

	// Verify route.go was transformed
	result, err := os.ReadFile(routePath)
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}
	content := string(result)

	if !strings.Contains(content, "ctx context.Context") {
		t.Error("expected ctx parameter in transformed route.go")
	}
	if !strings.Contains(content, "audited.AuditContext") {
		t.Error("expected audited.AuditContext parameter in transformed route.go")
	}

	// Verify stdout reports the update
	if !strings.Contains(string(out), "Updated route.go") {
		t.Errorf("expected 'Updated route.go' in stdout, got: %s", out)
	}
}

func TestMainCUD_NoChangesNeeded(t *testing.T) {
	t.Parallel()
	bin := buildCUDBinary(t)

	// Create entity files with no matching patterns
	tmpDir := t.TempDir()

	files := []string{
		"route.go", "role.go", "permission.go", "session.go", "token.go",
		"table.go", "field.go", "datatype.go", "datatype_field.go",
		"content_data.go", "content_field.go", "media.go", "media_dimension.go",
		"admin_content_data.go", "admin_content_field.go", "admin_datatype.go",
		"admin_datatype_field.go", "admin_field.go", "admin_route.go",
		"user_oauth.go", "user_ssh_keys.go",
	}
	for _, f := range files {
		p := filepath.Join(tmpDir, f)
		if err := os.WriteFile(p, []byte("package db\n"), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", f, err)
		}
	}

	cmd := exec.Command(bin, tmpDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	// Should report no changes for every file
	if !strings.Contains(string(out), "No changes needed") {
		t.Errorf("expected 'No changes needed' messages, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// Edge cases for findFuncEnd with interleaved string types
// ---------------------------------------------------------------------------

func TestFindFuncEnd_RawStringContainingDoubleQuotes(t *testing.T) {
	t.Parallel()
	// Raw strings can contain " characters without escaping.
	// Braces inside raw strings should not affect depth counting.
	content := "func foo() { s := `hello \"world\" { }` }"
	got := findFuncEnd(content, 0)
	// The function body ends at the last }
	if got != len(content) {
		t.Errorf("findFuncEnd = %d, want %d", got, len(content))
	}
}

func TestFindFuncEnd_ConsecutiveStrings(t *testing.T) {
	t.Parallel()
	// Multiple strings with braces in sequence
	content := `func foo() { a := "{ "; b := "} {"; return }`
	got := findFuncEnd(content, 0)
	if got != len(content) {
		t.Errorf("findFuncEnd = %d, want %d", got, len(content))
	}
}

func TestFindFuncEnd_MixedStringTypes(t *testing.T) {
	t.Parallel()
	// Mix of raw and interpreted strings containing braces
	content := "func foo() { a := `{`; b := \"}\"; c := `}`; return }"
	got := findFuncEnd(content, 0)
	if got != len(content) {
		t.Errorf("findFuncEnd = %d, want %d", got, len(content))
	}
}
