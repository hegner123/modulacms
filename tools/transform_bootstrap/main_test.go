package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// White-box test file: we test the unexported min helper directly,
// and use exec.Command integration tests for the main function since
// all replacement logic lives inside main().

func TestMin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{name: "a_less_than_b", a: 3, b: 10, want: 3},
		{name: "b_less_than_a", a: 10, b: 3, want: 3},
		{name: "equal", a: 5, b: 5, want: 5},
		{name: "zero_and_positive", a: 0, b: 7, want: 0},
		{name: "negative_values", a: -3, b: -1, want: -3},
		{name: "zero_both", a: 0, b: 0, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := min(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// buildBinary compiles the transform_bootstrap binary into a temp directory
// and returns the path to the binary. The caller should use t.Cleanup or
// defer to remove the temp dir.
func buildBinary(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "transform_bootstrap")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = filepath.Join(projectRoot(t), "tools", "transform_bootstrap")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}
	return binPath
}

// projectRoot walks up from the test file to find the module root (where go.mod lives).
func projectRoot(t *testing.T) string {
	t.Helper()
	// We know the test is in tools/transform_bootstrap/ within the module
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	// Walk up until we find go.mod
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

func TestMain_NoArgs_ExitNonZero(t *testing.T) {
	t.Parallel()
	bin := buildBinary(t)

	cmd := exec.Command(bin)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit when no args provided, got exit 0")
	}
}

func TestMain_NonexistentFile_ExitOne(t *testing.T) {
	t.Parallel()
	bin := buildBinary(t)

	cmd := exec.Command(bin, "/nonexistent/path/file.go")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit for nonexistent file, got exit 0")
	}
	if !strings.Contains(string(out), "Error reading") {
		t.Errorf("expected stderr to contain 'Error reading', got: %s", out)
	}
}

func TestMain_AppliesReplacements(t *testing.T) {
	t.Parallel()
	bin := buildBinary(t)

	// Build a minimal input that contains one of the replacement patterns.
	// We use the SQLite bootstrap header pattern -- the first replacement in the list.
	input := `package db

func (d Database) CreateBootstrapData() error {
	// 1. Create system admin permission (permission_id = 1)
	permission := d.CreatePermission(CreatePermissionParams{
		Label: "admin",
	})
	if permission.PermissionID.IsZero() {
		return fmt.Errorf("failed to create system admin permission")
	}

	// 2. Create system admin role (role_id = 1)
	adminRole := d.CreateRole(CreateRoleParams{
		Label: "admin",
	})
	if adminRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create system admin role")
	}

	// 3. Create viewer role (role_id = 4)
	viewerRole := d.CreateRole(CreateRoleParams{
		Label: "viewer",
	})
	if viewerRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create viewer role")
	}

	// 4. Create system admin user (user_id = 1)
	systemUser, err := d.CreateUser(CreateUserParams{
		Username: "admin",
	})

	// 5. Create default home route (route_id = 1) - Recommended
	homeRoute := d.CreateRoute(CreateRouteParams{
		Slug: "/",
	})
	if homeRoute.RouteID.IsZero() {
		return fmt.Errorf("failed to create default home route")
	}

	// 6. Create default page datatype (datatype_id = 1)
	pageDatatype := d.CreateDatatype(CreateDatatypeParams{
		Label: "page",
	})
	if pageDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create default page datatype")
	}

	// 7. Create default admin route (admin_route_id = 1)
	adminRoute := d.CreateAdminRoute(CreateAdminRouteParams{
		Slug: "/admin",
	})
	if adminRoute.AdminRouteID.IsZero() {
		return fmt.Errorf("failed to create default admin route")
	}

	// 8. Create default admin datatype (admin_datatype_id = 1)
	adminDatatype := d.CreateAdminDatatype(CreateAdminDatatypeParams{
		Label: "admin_page",
	})
	if adminDatatype.AdminDatatypeID.IsZero() {
		return fmt.Errorf("failed to create default admin datatype")
	}

	// 9. Create default admin field (admin_field_id = 1)
	adminField := d.CreateAdminField(CreateAdminFieldParams{
		Label: "title",
	})
	if adminField.AdminFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin field")
	}

	// 10. Create default field (field_id = 1)
	field := d.CreateField(CreateFieldParams{
		Label: "title",
	})
	if field.FieldID.IsZero() {
		return fmt.Errorf("failed to create default field")
	}

	// 11. Create default content_data record (content_data_id = 1)
	contentData := d.CreateContentData(CreateContentDataParams{
		Title: "Home",
	})
	if contentData.ContentDataID.IsZero() {
		return fmt.Errorf("failed to create default content_data")
	}

	// 12. Create default admin_content_data record (admin_content_data_id = 1)
	adminContentData := d.CreateAdminContentData(CreateAdminContentDataParams{
		Title: "Admin Home",
	})
	if adminContentData.AdminContentDataID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_data")
	}

	// 13. Create default content_field (content_field_id = 1)
	contentField := d.CreateContentField(CreateContentFieldParams{
		Label: "title",
	})
	if contentField.ContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default content_field")
	}

	// 14. Create default admin_content_field (admin_content_field_id = 1)
	adminContentField := d.CreateAdminContentField(CreateAdminContentFieldParams{
		Label: "title",
	})
	if adminContentField.AdminContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_field")
	}

	// 15. Create default media_dimension (md_id = 1) - Validation record
	mediaDimension := d.CreateMediaDimension(CreateMediaDimensionParams{
		Label: "thumbnail",
	})
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media := d.CreateMedia(CreateMediaParams{
		Label: "placeholder",
	})
	if media.MediaID.IsZero() {
		return fmt.Errorf("failed to create default media")
	}

	// 17. Create default token (id = 1) - Validation record
	token := d.CreateToken(CreateTokenParams{
		Label: "system",
	})
	if token.ID == "" {
		return fmt.Errorf("failed to create default token")
	}

	// 18. Create default session (session_id = 1) - Validation record
	session, err := d.CreateSession(CreateSessionParams{
		Label: "system",
	})

	// 19. Create default user_oauth record (user_oauth_id = 1) - Validation record
	userOauth, err := d.CreateUserOauth(CreateUserOauthParams{
		Provider: "system",
	})

	userSshKey, err := d.CreateUserSshKey(CreateUserSshKeyParams{
		Label: "system",
	})

		table := d.CreateTable(CreateTableParams{Label: tableName})
		if table.ID == "" {
			return fmt.Errorf("failed to register table")
		}

	datatypeField := d.CreateDatatypeField(CreateDatatypeFieldParams{
		Label: "default",
	})
	if datatypeField.ID == "" {
		return fmt.Errorf("failed to create default datatypes_fields")
	}

	// 22. Create default admin_datatypes_fields junction record
	adminDatatypeField := d.CreateAdminDatatypeField(CreateAdminDatatypeFieldParams{
		Label: "default",
	})
	if adminDatatypeField.ID == "" {
		return fmt.Errorf("failed to create default admin_datatypes_fields")
	}

	utility.DefaultLogger.Finfo(
		"Bootstrap complete",
	)
	return nil
}
`

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "bootstrap.go")
	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	cmd := exec.Command(bin, inputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary exited with error: %v\noutput: %s", err, out)
	}

	result, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("failed to read result file: %v", err)
	}
	content := string(result)

	// Verify the SQLite bootstrap header was injected (ctx and ac lines)
	if !strings.Contains(content, "ctx := context.Background()") {
		t.Error("expected 'ctx := context.Background()' in output")
	}
	if !strings.Contains(content, `ac := audited.Ctx(types.NodeID(d.Config.Node_ID)`) {
		t.Error("expected audited.Ctx line in output")
	}

	// Verify at least one Create was converted to audited form (permission)
	if !strings.Contains(content, "permission, err := d.CreatePermission(ctx, ac, CreatePermissionParams{") {
		t.Error("expected permission Create to be converted to audited form")
	}

	// Verify error check was added after permission creation
	if !strings.Contains(content, `return fmt.Errorf("failed to create system admin permission: %w", err)`) {
		t.Error("expected wrapped error for permission creation")
	}

	// Verify stdout contains success message
	if !strings.Contains(string(out), "Bootstrap methods updated successfully") {
		t.Errorf("expected success message in stdout, got: %s", out)
	}
}

func TestMain_WarnsOnMissingPattern(t *testing.T) {
	t.Parallel()
	bin := buildBinary(t)

	// Provide a file that does NOT contain any of the expected patterns.
	// The tool should warn on stderr for each missing pattern but still
	// write the file and exit 0.
	input := `package db

func (d Database) SomethingElse() error {
	return nil
}
`

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "bootstrap.go")
	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	cmd := exec.Command(bin, inputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected exit 0 even with missing patterns, got error: %v\noutput: %s", err, out)
	}

	// Should contain WARNING lines for missing patterns
	if !strings.Contains(string(out), "WARNING: pattern not found") {
		t.Errorf("expected warnings for missing patterns, got: %s", out)
	}
}

func TestMain_MysqlHeaderInjection(t *testing.T) {
	t.Parallel()
	bin := buildBinary(t)

	// Test that the MySQL-specific header injection works
	input := `package db

func (d MysqlDatabase) CreateBootstrapData() error {
	// 1. Create system admin permission (permission_id = 1)
	return nil
}
`

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "bootstrap.go")
	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	cmd := exec.Command(bin, inputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary exited with error: %v\noutput: %s", err, out)
	}

	result, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("failed to read result file: %v", err)
	}
	content := string(result)

	if !strings.Contains(content, "func (d MysqlDatabase) CreateBootstrapData() error {") {
		t.Error("expected MysqlDatabase function signature to remain")
	}
	if !strings.Contains(content, "ctx := context.Background()") {
		t.Error("expected ctx line injected into MySQL bootstrap")
	}
}

func TestMain_PsqlHeaderInjection(t *testing.T) {
	t.Parallel()
	bin := buildBinary(t)

	// Test that the PostgreSQL-specific header injection works
	input := `package db

func (d PsqlDatabase) CreateBootstrapData() error {
	// 1. Create system admin permission (permission_id = 1)
	return nil
}
`

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "bootstrap.go")
	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	cmd := exec.Command(bin, inputPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary exited with error: %v\noutput: %s", err, out)
	}

	result, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("failed to read result file: %v", err)
	}
	content := string(result)

	if !strings.Contains(content, "func (d PsqlDatabase) CreateBootstrapData() error {") {
		t.Error("expected PsqlDatabase function signature to remain")
	}
	if !strings.Contains(content, `ac := audited.Ctx(types.NodeID(d.Config.Node_ID)`) {
		t.Error("expected audited.Ctx line injected into PostgreSQL bootstrap")
	}
}

func TestMain_FileContentUnchangedWhenNoPatternsMatch(t *testing.T) {
	t.Parallel()
	bin := buildBinary(t)

	input := `package db

func (d Database) Unrelated() error {
	return nil
}
`

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "bootstrap.go")
	if err := os.WriteFile(inputPath, []byte(input), 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	cmd := exec.Command(bin, inputPath)
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("failed to read result file: %v", err)
	}

	// The MySQL and PostgreSQL header replacements won't match either,
	// so the content should still be written back (tool always writes).
	// Verify the original content is preserved since no patterns matched.
	if !strings.Contains(string(result), "func (d Database) Unrelated() error {") {
		t.Error("expected original content to be preserved")
	}
}
