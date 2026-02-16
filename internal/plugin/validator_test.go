package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePlugin_ValidBookmarksPlugin(t *testing.T) {
	dir := filepath.Join("testdata", "plugins", "test_bookmarks")

	info, results, err := ValidatePlugin(dir)
	if err != nil {
		t.Fatalf("ValidatePlugin returned unexpected error: %s", err)
	}

	// Check no errors in results (warnings are OK).
	for _, r := range results {
		if r.Severity == SeverityError {
			t.Errorf("unexpected error result: field=%q message=%q", r.Field, r.Message)
		}
	}

	if info.Name != "test_bookmarks" {
		t.Errorf("Name = %q, want %q", info.Name, "test_bookmarks")
	}
	if info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", info.Version, "1.0.0")
	}
	if info.Description != "Phase 1 integration test plugin" {
		t.Errorf("Description = %q, want %q", info.Description, "Phase 1 integration test plugin")
	}
	if info.Author != "ModulaCMS Test Suite" {
		t.Errorf("Author = %q, want %q", info.Author, "ModulaCMS Test Suite")
	}
	if !info.HasOnInit {
		t.Error("HasOnInit = false, want true (test_bookmarks defines on_init)")
	}
}

func TestValidatePlugin_InvalidNoManifest(t *testing.T) {
	dir := filepath.Join("testdata", "plugins", "invalid_no_manifest")

	info, results, err := ValidatePlugin(dir)
	if err != nil {
		t.Fatalf("ValidatePlugin returned unexpected error: %s", err)
	}

	// Should have at least one error result for missing manifest.
	hasError := false
	for _, r := range results {
		if r.Severity == SeverityError {
			hasError = true
			if !strings.Contains(r.Message, "missing required plugin_info") {
				t.Errorf("error message = %q, want to contain %q", r.Message, "missing required plugin_info")
			}
		}
	}
	if !hasError {
		t.Error("expected at least one error result for missing manifest")
	}

	// Info should be zero value when extraction fails.
	if info.Name != "" {
		t.Errorf("Name = %q, want empty for failed extraction", info.Name)
	}
}

func TestValidatePlugin_InvalidBadName(t *testing.T) {
	dir := filepath.Join("testdata", "plugins", "invalid_bad_name")

	info, results, err := ValidatePlugin(dir)
	if err != nil {
		t.Fatalf("ValidatePlugin returned unexpected error: %s", err)
	}

	// Should have at least one error result for invalid name.
	hasError := false
	for _, r := range results {
		if r.Severity == SeverityError {
			hasError = true
			if !strings.Contains(r.Message, "invalid character") {
				t.Errorf("error message = %q, want to contain %q", r.Message, "invalid character")
			}
		}
	}
	if !hasError {
		t.Error("expected at least one error result for invalid name")
	}

	// Info should be zero value when validation fails.
	if info.Name != "" {
		t.Errorf("Name = %q, want empty for failed validation", info.Name)
	}
}

func TestValidatePlugin_NonexistentDirectory(t *testing.T) {
	_, _, err := ValidatePlugin(filepath.Join(t.TempDir(), "does_not_exist"))
	if err == nil {
		t.Fatal("expected error for nonexistent directory, got nil")
	}
}

func TestValidatePlugin_NotADirectory(t *testing.T) {
	// Create a regular file, not a directory.
	tmpFile := filepath.Join(t.TempDir(), "not_a_dir")
	if writeErr := os.WriteFile(tmpFile, []byte("hello"), 0o644); writeErr != nil {
		t.Fatalf("creating temp file: %s", writeErr)
	}

	_, _, err := ValidatePlugin(tmpFile)
	if err == nil {
		t.Fatal("expected error for file (not directory), got nil")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "not a directory")
	}
}

func TestValidatePlugin_MissingInitLua(t *testing.T) {
	// Create a directory with no init.lua.
	dir := filepath.Join(t.TempDir(), "empty_plugin")
	if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
		t.Fatalf("creating dir: %s", mkErr)
	}

	info, results, err := ValidatePlugin(dir)
	if err != nil {
		t.Fatalf("ValidatePlugin returned unexpected error: %s", err)
	}

	hasInitError := false
	for _, r := range results {
		if r.Field == "init.lua" && r.Severity == SeverityError {
			hasInitError = true
		}
	}
	if !hasInitError {
		t.Error("expected error result for missing init.lua")
	}
	if info.Name != "" {
		t.Errorf("Name = %q, want empty when init.lua is missing", info.Name)
	}
}

func TestValidatePlugin_SyntaxError(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "bad_syntax")
	if mkErr := os.MkdirAll(pluginDir, 0o755); mkErr != nil {
		t.Fatalf("creating dir: %s", mkErr)
	}
	initPath := filepath.Join(pluginDir, "init.lua")
	if writeErr := os.WriteFile(initPath, []byte("this is not valid lua {{{{"), 0o644); writeErr != nil {
		t.Fatalf("writing init.lua: %s", writeErr)
	}

	info, results, err := ValidatePlugin(pluginDir)
	if err != nil {
		t.Fatalf("ValidatePlugin returned unexpected error: %s", err)
	}

	hasSyntaxError := false
	for _, r := range results {
		if r.Field == "syntax" && r.Severity == SeverityError {
			hasSyntaxError = true
		}
	}
	if !hasSyntaxError {
		t.Error("expected error result for syntax error")
	}
	if info.Name != "" {
		t.Errorf("Name = %q, want empty when syntax check fails", info.Name)
	}
}

func TestValidatePlugin_HasOnInit(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "with_init")
	if mkErr := os.MkdirAll(pluginDir, 0o755); mkErr != nil {
		t.Fatalf("creating dir: %s", mkErr)
	}
	initPath := filepath.Join(pluginDir, "init.lua")
	luaCode := `
plugin_info = {
    name        = "with_init",
    version     = "1.0.0",
    description = "Plugin with on_init defined",
    author      = "test",
    license     = "MIT",
}

function on_init()
    log.info("hello")
end
`
	if writeErr := os.WriteFile(initPath, []byte(luaCode), 0o644); writeErr != nil {
		t.Fatalf("writing init.lua: %s", writeErr)
	}

	info, results, err := ValidatePlugin(pluginDir)
	if err != nil {
		t.Fatalf("ValidatePlugin returned unexpected error: %s", err)
	}

	// No errors expected.
	for _, r := range results {
		if r.Severity == SeverityError {
			t.Errorf("unexpected error: field=%q message=%q", r.Field, r.Message)
		}
	}

	if !info.HasOnInit {
		t.Error("HasOnInit = false, want true")
	}
}

func TestValidatePlugin_NoOnInit(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "no_init_fn")
	if mkErr := os.MkdirAll(pluginDir, 0o755); mkErr != nil {
		t.Fatalf("creating dir: %s", mkErr)
	}
	initPath := filepath.Join(pluginDir, "init.lua")
	luaCode := `
plugin_info = {
    name        = "no_init_fn",
    version     = "1.0.0",
    description = "Plugin without on_init",
    author      = "test",
    license     = "MIT",
}
`
	if writeErr := os.WriteFile(initPath, []byte(luaCode), 0o644); writeErr != nil {
		t.Fatalf("writing init.lua: %s", writeErr)
	}

	info, results, err := ValidatePlugin(pluginDir)
	if err != nil {
		t.Fatalf("ValidatePlugin returned unexpected error: %s", err)
	}

	if info.HasOnInit {
		t.Error("HasOnInit = true, want false")
	}

	// Should have a warning about missing on_init.
	hasOnInitWarning := false
	for _, r := range results {
		if r.Field == "on_init" && r.Severity == SeverityWarning {
			hasOnInitWarning = true
		}
	}
	if !hasOnInitWarning {
		t.Error("expected warning result for missing on_init")
	}
}

func TestValidatePlugin_MissingOptionalFields(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "minimal")
	if mkErr := os.MkdirAll(pluginDir, 0o755); mkErr != nil {
		t.Fatalf("creating dir: %s", mkErr)
	}
	initPath := filepath.Join(pluginDir, "init.lua")
	luaCode := `
plugin_info = {
    name        = "minimal",
    version     = "1.0.0",
    description = "Minimal plugin without optional fields",
}
`
	if writeErr := os.WriteFile(initPath, []byte(luaCode), 0o644); writeErr != nil {
		t.Fatalf("writing init.lua: %s", writeErr)
	}

	info, results, err := ValidatePlugin(pluginDir)
	if err != nil {
		t.Fatalf("ValidatePlugin returned unexpected error: %s", err)
	}

	// No errors expected.
	for _, r := range results {
		if r.Severity == SeverityError {
			t.Errorf("unexpected error: field=%q message=%q", r.Field, r.Message)
		}
	}

	if info.Name != "minimal" {
		t.Errorf("Name = %q, want %q", info.Name, "minimal")
	}

	// Should have warnings for missing author and license.
	warningFields := make(map[string]bool)
	for _, r := range results {
		if r.Severity == SeverityWarning {
			warningFields[r.Field] = true
		}
	}
	if !warningFields["author"] {
		t.Error("expected warning for missing author")
	}
	if !warningFields["license"] {
		t.Error("expected warning for missing license")
	}
}

func TestValidatePlugin_FailingPluginFixture(t *testing.T) {
	// The failing_plugin fixture has a valid manifest but on_init that errors.
	// ValidatePlugin should succeed (it does not execute on_init, only extracts
	// the manifest). The plugin will fail at runtime, not at validation time.
	dir := filepath.Join("testdata", "plugins", "failing_plugin")

	info, results, err := ValidatePlugin(dir)
	if err != nil {
		t.Fatalf("ValidatePlugin returned unexpected error: %s", err)
	}

	// No errors expected -- manifest is valid.
	for _, r := range results {
		if r.Severity == SeverityError {
			t.Errorf("unexpected error: field=%q message=%q", r.Field, r.Message)
		}
	}

	if info.Name != "failing_plugin" {
		t.Errorf("Name = %q, want %q", info.Name, "failing_plugin")
	}
	if !info.HasOnInit {
		t.Error("HasOnInit = false, want true (failing_plugin defines on_init)")
	}
}

func TestValidationSeverity_String(t *testing.T) {
	tests := []struct {
		severity ValidationSeverity
		want     string
	}{
		{SeverityError, "error"},
		{SeverityWarning, "warning"},
		{ValidationSeverity(99), "unknown(99)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.String()
			if got != tt.want {
				t.Errorf("ValidationSeverity(%d).String() = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}
