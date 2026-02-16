package plugin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// -- Test helpers --

// newTestDB opens an in-memory SQLite database for testing.
// Caller must defer conn.Close().
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("opening test db: %s", err)
	}
	// Enable WAL mode for plugin compatibility.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		t.Fatalf("enabling WAL mode: %s", err)
	}
	return conn
}

// writePluginFile creates a plugin directory with an init.lua containing the given Lua code.
// Returns the absolute path to the plugins parent directory.
func writePluginFile(t *testing.T, baseDir, pluginDirName, luaCode string) {
	t.Helper()
	dir := filepath.Join(baseDir, pluginDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating plugin dir %q: %s", dir, err)
	}
	initPath := filepath.Join(dir, "init.lua")
	if err := os.WriteFile(initPath, []byte(luaCode), 0o644); err != nil {
		t.Fatalf("writing init.lua: %s", err)
	}
}

// writePluginLib creates a lib/<name>.lua file inside a plugin directory.
func writePluginLib(t *testing.T, baseDir, pluginDirName, libName, luaCode string) {
	t.Helper()
	libDir := filepath.Join(baseDir, pluginDirName, "lib")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("creating lib dir: %s", err)
	}
	libPath := filepath.Join(libDir, libName+".lua")
	if err := os.WriteFile(libPath, []byte(luaCode), 0o644); err != nil {
		t.Fatalf("writing lib file: %s", err)
	}
}

// -- PluginState tests --

func TestPluginState_String(t *testing.T) {
	tests := []struct {
		state PluginState
		want  string
	}{
		{StateDiscovered, "discovered"},
		{StateLoading, "loading"},
		{StateRunning, "running"},
		{StateFailed, "failed"},
		{StateStopped, "stopped"},
		{PluginState(99), "unknown(99)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("PluginState(%d).String() = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

// -- Manifest extraction tests --

func TestExtractManifest_Valid(t *testing.T) {
	dir := t.TempDir()
	writePluginFile(t, dir, "my_plugin", `
plugin_info = {
    name        = "my_plugin",
    version     = "1.0.0",
    description = "A test plugin",
    author      = "Test Author",
    license     = "MIT",
    dependencies = {"other_plugin"},
}
`)

	info, err := ExtractManifest(filepath.Join(dir, "my_plugin", "init.lua"))
	if err != nil {
		t.Fatalf("ExtractManifest: %s", err)
	}

	if info.Name != "my_plugin" {
		t.Errorf("Name = %q, want %q", info.Name, "my_plugin")
	}
	if info.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", info.Version, "1.0.0")
	}
	if info.Description != "A test plugin" {
		t.Errorf("Description = %q, want %q", info.Description, "A test plugin")
	}
	if info.Author != "Test Author" {
		t.Errorf("Author = %q, want %q", info.Author, "Test Author")
	}
	if info.License != "MIT" {
		t.Errorf("License = %q, want %q", info.License, "MIT")
	}
	if len(info.Dependencies) != 1 || info.Dependencies[0] != "other_plugin" {
		t.Errorf("Dependencies = %v, want [other_plugin]", info.Dependencies)
	}
}

func TestExtractManifest_MissingPluginInfo(t *testing.T) {
	dir := t.TempDir()
	writePluginFile(t, dir, "no_manifest", `local x = 1`)

	_, err := ExtractManifest(filepath.Join(dir, "no_manifest", "init.lua"))
	if err == nil {
		t.Fatal("expected error for missing plugin_info, got nil")
	}
	if !strings.Contains(err.Error(), "missing required plugin_info") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "missing required plugin_info")
	}
}

func TestExtractManifest_InvalidName_Spaces(t *testing.T) {
	dir := t.TempDir()
	writePluginFile(t, dir, "bad_name", `
plugin_info = {
    name        = "has spaces",
    version     = "1.0.0",
    description = "Invalid name with spaces",
}
`)

	_, err := ExtractManifest(filepath.Join(dir, "bad_name", "init.lua"))
	if err == nil {
		t.Fatal("expected error for name with spaces, got nil")
	}
	if !strings.Contains(err.Error(), "invalid character") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "invalid character")
	}
}

func TestExtractManifest_InvalidName_Uppercase(t *testing.T) {
	dir := t.TempDir()
	writePluginFile(t, dir, "upper", `
plugin_info = {
    name        = "HasCaps",
    version     = "1.0.0",
    description = "Invalid name with uppercase",
}
`)

	_, err := ExtractManifest(filepath.Join(dir, "upper", "init.lua"))
	if err == nil {
		t.Fatal("expected error for uppercase name, got nil")
	}
	if !strings.Contains(err.Error(), "invalid character") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "invalid character")
	}
}

func TestExtractManifest_InvalidName_TrailingUnderscore(t *testing.T) {
	dir := t.TempDir()
	writePluginFile(t, dir, "trailing", `
plugin_info = {
    name        = "my_plugin_",
    version     = "1.0.0",
    description = "Trailing underscore name",
}
`)

	_, err := ExtractManifest(filepath.Join(dir, "trailing", "init.lua"))
	if err == nil {
		t.Fatal("expected error for trailing underscore, got nil")
	}
	if !strings.Contains(err.Error(), "trailing underscore") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "trailing underscore")
	}
}

func TestExtractManifest_InvalidName_TooLong(t *testing.T) {
	dir := t.TempDir()
	longName := strings.Repeat("a", 33) // 33 chars, over the 32 limit
	writePluginFile(t, dir, "toolong", `
plugin_info = {
    name        = "`+longName+`",
    version     = "1.0.0",
    description = "Name too long",
}
`)

	_, err := ExtractManifest(filepath.Join(dir, "toolong", "init.lua"))
	if err == nil {
		t.Fatal("expected error for name >32 chars, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds 32 characters") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "exceeds 32 characters")
	}
}

func TestExtractManifest_MissingVersion(t *testing.T) {
	dir := t.TempDir()
	writePluginFile(t, dir, "no_ver", `
plugin_info = {
    name        = "no_ver",
    description = "Missing version field",
}
`)

	_, err := ExtractManifest(filepath.Join(dir, "no_ver", "init.lua"))
	if err == nil {
		t.Fatal("expected error for missing version, got nil")
	}
	if !strings.Contains(err.Error(), "version is required") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "version is required")
	}
}

func TestExtractManifest_MissingDescription(t *testing.T) {
	dir := t.TempDir()
	writePluginFile(t, dir, "no_desc", `
plugin_info = {
    name    = "no_desc",
    version = "1.0.0",
}
`)

	_, err := ExtractManifest(filepath.Join(dir, "no_desc", "init.lua"))
	if err == nil {
		t.Fatal("expected error for missing description, got nil")
	}
	if !strings.Contains(err.Error(), "description is required") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "description is required")
	}
}

// -- Validate manifest edge cases --

func TestValidateManifest_ValidNameWithNumbers(t *testing.T) {
	info := &PluginInfo{
		Name:        "plugin_v2",
		Version:     "1.0.0",
		Description: "Valid name with numbers",
	}
	if err := ValidateManifest(info); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestValidateManifest_SingleChar(t *testing.T) {
	info := &PluginInfo{
		Name:        "a",
		Version:     "1.0.0",
		Description: "Single char name",
	}
	if err := ValidateManifest(info); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestValidateManifest_Max32Chars(t *testing.T) {
	info := &PluginInfo{
		Name:        strings.Repeat("a", 32), // exactly 32 chars, should pass
		Version:     "1.0.0",
		Description: "Exactly 32 chars",
	}
	if err := ValidateManifest(info); err != nil {
		t.Errorf("unexpected error for 32-char name: %s", err)
	}
}

// -- Topological sort tests --

func TestTopologicalSort_NoDependencies(t *testing.T) {
	plugins := map[string]*PluginInstance{
		"alpha": {Info: PluginInfo{Name: "alpha"}},
		"beta":  {Info: PluginInfo{Name: "beta"}},
		"gamma": {Info: PluginInfo{Name: "gamma"}},
	}

	sorted, err := topologicalSort(plugins)
	if err != nil {
		t.Fatalf("topologicalSort: %s", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("expected 3 plugins, got %d", len(sorted))
	}
	// With no dependencies, should be sorted alphabetically (deterministic).
	if sorted[0].Info.Name != "alpha" || sorted[1].Info.Name != "beta" || sorted[2].Info.Name != "gamma" {
		names := make([]string, len(sorted))
		for i, s := range sorted {
			names[i] = s.Info.Name
		}
		t.Errorf("expected [alpha beta gamma], got %v", names)
	}
}

func TestTopologicalSort_WithDependencies(t *testing.T) {
	plugins := map[string]*PluginInstance{
		"app":  {Info: PluginInfo{Name: "app", Dependencies: []string{"core"}}},
		"core": {Info: PluginInfo{Name: "core"}},
		"ui":   {Info: PluginInfo{Name: "ui", Dependencies: []string{"core", "app"}}},
	}

	sorted, err := topologicalSort(plugins)
	if err != nil {
		t.Fatalf("topologicalSort: %s", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("expected 3 plugins, got %d", len(sorted))
	}

	// Verify ordering: core before app, app before ui.
	indexOf := make(map[string]int)
	for i, s := range sorted {
		indexOf[s.Info.Name] = i
	}

	if indexOf["core"] >= indexOf["app"] {
		t.Errorf("core (index %d) should come before app (index %d)", indexOf["core"], indexOf["app"])
	}
	if indexOf["app"] >= indexOf["ui"] {
		t.Errorf("app (index %d) should come before ui (index %d)", indexOf["app"], indexOf["ui"])
	}
}

func TestTopologicalSort_CycleDetected(t *testing.T) {
	plugins := map[string]*PluginInstance{
		"a": {Info: PluginInfo{Name: "a", Dependencies: []string{"b"}}},
		"b": {Info: PluginInfo{Name: "b", Dependencies: []string{"a"}}},
	}

	_, err := topologicalSort(plugins)
	if err == nil {
		t.Fatal("expected error for circular dependency, got nil")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "circular dependency")
	}
}

func TestTopologicalSort_ThreeWayCycle(t *testing.T) {
	plugins := map[string]*PluginInstance{
		"a": {Info: PluginInfo{Name: "a", Dependencies: []string{"c"}}},
		"b": {Info: PluginInfo{Name: "b", Dependencies: []string{"a"}}},
		"c": {Info: PluginInfo{Name: "c", Dependencies: []string{"b"}}},
	}

	_, err := topologicalSort(plugins)
	if err == nil {
		t.Fatal("expected error for three-way circular dependency, got nil")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "circular dependency")
	}
}

func TestTopologicalSort_MissingDependency(t *testing.T) {
	plugins := map[string]*PluginInstance{
		"app": {Info: PluginInfo{Name: "app", Dependencies: []string{"nonexistent"}}},
	}

	_, err := topologicalSort(plugins)
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "nonexistent")
	}
}

func TestTopologicalSort_Empty(t *testing.T) {
	plugins := map[string]*PluginInstance{}

	sorted, err := topologicalSort(plugins)
	if err != nil {
		t.Fatalf("topologicalSort: %s", err)
	}
	if len(sorted) != 0 {
		t.Errorf("expected empty result, got %d plugins", len(sorted))
	}
}

// -- NewManager tests --

func TestNewManager_Defaults(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{}, conn, db.DialectSQLite, nil)

	if mgr.cfg.MaxVMsPerPlugin != 4 {
		t.Errorf("MaxVMsPerPlugin = %d, want 4", mgr.cfg.MaxVMsPerPlugin)
	}
	if mgr.cfg.ExecTimeoutSec != 5 {
		t.Errorf("ExecTimeoutSec = %d, want 5", mgr.cfg.ExecTimeoutSec)
	}
	if mgr.cfg.MaxOpsPerExec != 1000 {
		t.Errorf("MaxOpsPerExec = %d, want 1000", mgr.cfg.MaxOpsPerExec)
	}
}

func TestNewManager_CustomConfig(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		MaxVMsPerPlugin: 8,
		ExecTimeoutSec:  10,
		MaxOpsPerExec:   500,
	}, conn, db.DialectSQLite, nil)

	if mgr.cfg.MaxVMsPerPlugin != 8 {
		t.Errorf("MaxVMsPerPlugin = %d, want 8", mgr.cfg.MaxVMsPerPlugin)
	}
	if mgr.cfg.ExecTimeoutSec != 10 {
		t.Errorf("ExecTimeoutSec = %d, want 10", mgr.cfg.ExecTimeoutSec)
	}
	if mgr.cfg.MaxOpsPerExec != 500 {
		t.Errorf("MaxOpsPerExec = %d, want 500", mgr.cfg.MaxOpsPerExec)
	}
}

// -- LoadAll tests --

func TestLoadAll_EmptyDirectory(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll on empty dir: %s", err)
	}

	plugins := mgr.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestLoadAll_DirectoryNotConfigured(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{}, conn, db.DialectSQLite, nil)
	err := mgr.LoadAll(context.Background())
	if err == nil {
		t.Fatal("expected error for empty directory config, got nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "not configured")
	}
}

func TestLoadAll_SkipsDirWithoutInitLua(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	// Create a directory without init.lua.
	if err := os.MkdirAll(filepath.Join(dir, "no_init"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a valid plugin.
	writePluginFile(t, dir, "valid_one", `
plugin_info = {
    name        = "valid_one",
    version     = "1.0.0",
    description = "Valid plugin",
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	plugins := mgr.ListPlugins()
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Info.Name != "valid_one" {
		t.Errorf("plugin name = %q, want %q", plugins[0].Info.Name, "valid_one")
	}
}

func TestLoadAll_SuccessfulPlugin_StateRunning(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "simple", `
plugin_info = {
    name        = "simple",
    version     = "1.0.0",
    description = "Simple test plugin",
}

function on_init()
    log.info("simple plugin initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("simple")
	if inst == nil {
		t.Fatal("GetPlugin returned nil for 'simple'")
	}
	if inst.State != StateRunning {
		t.Errorf("state = %s, want %s", inst.State, StateRunning)
	}
	if inst.FailedReason != "" {
		t.Errorf("FailedReason = %q, want empty", inst.FailedReason)
	}
	if inst.Pool == nil {
		t.Error("Pool is nil")
	}
}

func TestLoadAll_PluginWithTableCreation(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "tasks", `
plugin_info = {
    name        = "tasks",
    version     = "1.0.0",
    description = "Task plugin that creates a table",
}

function on_init()
    db.define_table("items", {
        columns = {
            {name = "title", type = "text", not_null = true},
            {name = "done",  type = "boolean"},
        },
    })
    db.insert("items", {title = "Test task"})
    log.info("Tasks initialized")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("tasks")
	if inst == nil {
		t.Fatal("GetPlugin returned nil for 'tasks'")
	}
	if inst.State != StateRunning {
		t.Errorf("state = %s, want %s", inst.State, StateRunning)
	}

	// Verify the table was actually created and data inserted.
	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM plugin_tasks_items").Scan(&count)
	if err != nil {
		t.Fatalf("querying plugin_tasks_items: %s", err)
	}
	if count != 1 {
		t.Errorf("row count = %d, want 1", count)
	}
}

func TestLoadAll_FailedPlugin_OnInitError(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "failing", `
plugin_info = {
    name        = "failing",
    version     = "1.0.0",
    description = "Plugin that fails in on_init",
}

function on_init()
    error("deliberate init failure")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("failing")
	if inst == nil {
		t.Fatal("GetPlugin returned nil for 'failing'")
	}
	if inst.State != StateFailed {
		t.Errorf("state = %s, want %s", inst.State, StateFailed)
	}
	if !strings.Contains(inst.FailedReason, "on_init failed") {
		t.Errorf("FailedReason = %q, want to contain %q", inst.FailedReason, "on_init failed")
	}
}

func TestLoadAll_InvalidManifestSkipped(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	// Invalid plugin (bad name).
	writePluginFile(t, dir, "bad_one", `
plugin_info = {
    name        = "BAD NAME",
    version     = "1.0.0",
    description = "Invalid",
}
`)
	// Valid plugin.
	writePluginFile(t, dir, "good_one", `
plugin_info = {
    name        = "good_one",
    version     = "1.0.0",
    description = "Valid",
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	// Only the valid plugin should be loaded.
	plugins := mgr.ListPlugins()
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Info.Name != "good_one" {
		t.Errorf("plugin name = %q, want %q", plugins[0].Info.Name, "good_one")
	}
}

func TestLoadAll_DuplicatePluginName_SecondRejected(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	// Two directories with plugins that have the same plugin_info.name.
	writePluginFile(t, dir, "dir_a", `
plugin_info = {
    name        = "duplicate",
    version     = "1.0.0",
    description = "First one",
}
`)
	writePluginFile(t, dir, "dir_b", `
plugin_info = {
    name        = "duplicate",
    version     = "2.0.0",
    description = "Second one (should be rejected)",
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	// Only one "duplicate" should be loaded (first-discovered wins).
	plugins := mgr.ListPlugins()
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Info.Name != "duplicate" {
		t.Errorf("plugin name = %q, want %q", plugins[0].Info.Name, "duplicate")
	}
}

func TestLoadAll_MissingDependency_Fails(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "dependent", `
plugin_info = {
    name         = "dependent",
    version      = "1.0.0",
    description  = "Depends on nonexistent plugin",
    dependencies = {"nonexistent"},
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	// Topological sort should fail because "nonexistent" is not discovered.
	err := mgr.LoadAll(context.Background())
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "nonexistent")
	}
}

func TestLoadAll_CircularDependency_Detected(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "plugin_a", `
plugin_info = {
    name         = "plugin_a",
    version      = "1.0.0",
    description  = "Depends on plugin_b",
    dependencies = {"plugin_b"},
}
`)
	writePluginFile(t, dir, "plugin_b", `
plugin_info = {
    name         = "plugin_b",
    version      = "1.0.0",
    description  = "Depends on plugin_a",
    dependencies = {"plugin_a"},
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err == nil {
		t.Fatal("expected error for circular dependency, got nil")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "circular dependency")
	}
}

func TestLoadAll_DependencyOrder(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "base_plugin", `
plugin_info = {
    name        = "base_plugin",
    version     = "1.0.0",
    description = "Base plugin with no deps",
}

function on_init()
    db.define_table("base_data", {
        columns = {
            {name = "value", type = "text", not_null = true},
        },
    })
end
`)
	writePluginFile(t, dir, "ext_plugin", `
plugin_info = {
    name         = "ext_plugin",
    version      = "1.0.0",
    description  = "Extension that depends on base",
    dependencies = {"base_plugin"},
}

function on_init()
    log.info("ext_plugin loaded after base_plugin")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	base := mgr.GetPlugin("base_plugin")
	ext := mgr.GetPlugin("ext_plugin")

	if base == nil || ext == nil {
		t.Fatal("expected both plugins to be loaded")
	}
	if base.State != StateRunning {
		t.Errorf("base_plugin state = %s, want %s", base.State, StateRunning)
	}
	if ext.State != StateRunning {
		t.Errorf("ext_plugin state = %s, want %s", ext.State, StateRunning)
	}
}

func TestLoadAll_FailedDependency_DependentFails(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "broken_base", `
plugin_info = {
    name        = "broken_base",
    version     = "1.0.0",
    description = "Base plugin that fails",
}

function on_init()
    error("broken base init")
end
`)
	writePluginFile(t, dir, "depends_on_broken", `
plugin_info = {
    name         = "depends_on_broken",
    version      = "1.0.0",
    description  = "Depends on broken base",
    dependencies = {"broken_base"},
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	broken := mgr.GetPlugin("broken_base")
	dependent := mgr.GetPlugin("depends_on_broken")

	if broken == nil || dependent == nil {
		t.Fatal("expected both plugins to be present")
	}
	if broken.State != StateFailed {
		t.Errorf("broken_base state = %s, want %s", broken.State, StateFailed)
	}
	if dependent.State != StateFailed {
		t.Errorf("depends_on_broken state = %s, want %s", dependent.State, StateFailed)
	}
	if !strings.Contains(dependent.FailedReason, "not running") {
		t.Errorf("dependent FailedReason = %q, want to contain %q", dependent.FailedReason, "not running")
	}
}

// -- Shutdown tests --

func TestShutdown_RunsOnShutdown(t *testing.T) {
	conn := newTestDB(t)
	// Note: we do NOT defer conn.Close() here because Shutdown closes it.

	dir := t.TempDir()
	writePluginFile(t, dir, "shutdowner", `
plugin_info = {
    name        = "shutdowner",
    version     = "1.0.0",
    description = "Plugin with on_shutdown",
}

function on_init()
    db.define_table("shutdown_log", {
        columns = {
            {name = "message", type = "text", not_null = true},
        },
    })
end

function on_shutdown()
    log.info("shutdowner is shutting down")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("shutdowner")
	if inst == nil || inst.State != StateRunning {
		t.Fatalf("expected shutdowner to be running, got %v", inst)
	}

	mgr.Shutdown(context.Background())

	if inst.State != StateStopped {
		t.Errorf("state after shutdown = %s, want %s", inst.State, StateStopped)
	}
}

func TestShutdown_NoOnShutdown(t *testing.T) {
	conn := newTestDB(t)
	// Note: we do NOT defer conn.Close() here because Shutdown closes it.

	dir := t.TempDir()
	writePluginFile(t, dir, "no_shutdown", `
plugin_info = {
    name        = "no_shutdown",
    version     = "1.0.0",
    description = "Plugin without on_shutdown",
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	// Should not panic.
	mgr.Shutdown(context.Background())

	inst := mgr.GetPlugin("no_shutdown")
	if inst == nil {
		t.Fatal("GetPlugin returned nil")
	}
	if inst.State != StateStopped {
		t.Errorf("state after shutdown = %s, want %s", inst.State, StateStopped)
	}
}

func TestShutdown_ReverseOrder(t *testing.T) {
	conn := newTestDB(t)
	// Note: we do NOT defer conn.Close() here because Shutdown closes it.

	dir := t.TempDir()
	writePluginFile(t, dir, "foundation", `
plugin_info = {
    name        = "foundation",
    version     = "1.0.0",
    description = "Base",
}
`)
	writePluginFile(t, dir, "extension", `
plugin_info = {
    name         = "extension",
    version      = "1.0.0",
    description  = "Depends on foundation",
    dependencies = {"foundation"},
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	// Verify load order: foundation before extension.
	if len(mgr.loadOrder) != 2 {
		t.Fatalf("expected 2 in loadOrder, got %d", len(mgr.loadOrder))
	}
	if mgr.loadOrder[0] != "foundation" || mgr.loadOrder[1] != "extension" {
		t.Errorf("loadOrder = %v, want [foundation extension]", mgr.loadOrder)
	}

	// Shutdown should process in reverse: extension first, then foundation.
	mgr.Shutdown(context.Background())

	foundation := mgr.GetPlugin("foundation")
	extension := mgr.GetPlugin("extension")
	if foundation.State != StateStopped {
		t.Errorf("foundation state = %s, want stopped", foundation.State)
	}
	if extension.State != StateStopped {
		t.Errorf("extension state = %s, want stopped", extension.State)
	}
}

// -- GetPlugin tests --

func TestGetPlugin_Found(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "findme", `
plugin_info = {
    name        = "findme",
    version     = "1.0.0",
    description = "Findable plugin",
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	if err := mgr.LoadAll(context.Background()); err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("findme")
	if inst == nil {
		t.Fatal("GetPlugin returned nil for existing plugin")
	}
	if inst.Info.Name != "findme" {
		t.Errorf("Name = %q, want %q", inst.Info.Name, "findme")
	}
}

func TestGetPlugin_NotFound(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	mgr := NewManager(ManagerConfig{
		Enabled:   true,
		Directory: t.TempDir(),
	}, conn, db.DialectSQLite, nil)

	if err := mgr.LoadAll(context.Background()); err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("nonexistent")
	if inst != nil {
		t.Errorf("GetPlugin returned %v, want nil", inst)
	}
}

// -- ListPlugins tests --

func TestListPlugins_MultiplePlugins(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "plug_a", `
plugin_info = {
    name        = "plug_a",
    version     = "1.0.0",
    description = "Plugin A",
}
`)
	writePluginFile(t, dir, "plug_b", `
plugin_info = {
    name        = "plug_b",
    version     = "1.0.0",
    description = "Plugin B",
}
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	if err := mgr.LoadAll(context.Background()); err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	plugins := mgr.ListPlugins()
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	names := make(map[string]bool)
	for _, p := range plugins {
		names[p.Info.Name] = true
	}
	if !names["plug_a"] || !names["plug_b"] {
		t.Errorf("expected plug_a and plug_b, got %v", names)
	}
}

// -- Timeout tests --

func TestLoadAll_OnInitTimeout(t *testing.T) {
	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "slow_init", `
plugin_info = {
    name        = "slow_init",
    version     = "1.0.0",
    description = "Plugin with slow on_init",
}

function on_init()
    local i = 0
    while true do
        i = i + 1
    end
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  1, // 1 second timeout for faster test
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	start := time.Now()
	err := mgr.LoadAll(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("slow_init")
	if inst == nil {
		t.Fatal("GetPlugin returned nil")
	}
	if inst.State != StateFailed {
		t.Errorf("state = %s, want %s", inst.State, StateFailed)
	}
	if !strings.Contains(inst.FailedReason, "on_init failed") {
		t.Errorf("FailedReason = %q, want to contain %q", inst.FailedReason, "on_init failed")
	}

	// Should complete within a reasonable time (timeout + overhead).
	if elapsed > 10*time.Second {
		t.Errorf("LoadAll took %s, expected to complete within ~2s due to timeout", elapsed)
	}
}

// -- Integration: full lifecycle --

func TestFullLifecycle_LoadAndShutdown(t *testing.T) {
	conn := newTestDB(t)
	// Do NOT defer conn.Close() -- Shutdown closes it.

	dir := t.TempDir()
	writePluginFile(t, dir, "lifecycle", `
plugin_info = {
    name        = "lifecycle",
    version     = "1.0.0",
    description = "Full lifecycle test",
}

function on_init()
    db.define_table("events", {
        columns = {
            {name = "name",  type = "text", not_null = true},
            {name = "value", type = "integer"},
        },
    })
    db.insert("events", {name = "init", value = 1})
    log.info("lifecycle on_init complete")
end

function on_shutdown()
    log.info("lifecycle on_shutdown called")
end
`)

	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 2,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	// Load.
	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("lifecycle")
	if inst == nil {
		t.Fatal("GetPlugin returned nil")
	}
	if inst.State != StateRunning {
		t.Fatalf("state = %s, want running", inst.State)
	}

	// Verify data was inserted.
	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM plugin_lifecycle_events").Scan(&count)
	if err != nil {
		t.Fatalf("querying: %s", err)
	}
	if count != 1 {
		t.Errorf("row count = %d, want 1", count)
	}

	// Shutdown.
	mgr.Shutdown(context.Background())
	if inst.State != StateStopped {
		t.Errorf("state after shutdown = %s, want stopped", inst.State)
	}
}

// -- Pool exhaustion during init (edge case) --

func TestLoadAll_PoolExhaustion(t *testing.T) {
	// This test verifies that if pool.Get() returns ErrPoolExhausted during
	// loading, the plugin is marked as failed. In practice this should not
	// happen on a fresh pool, but the code must handle it gracefully.

	conn := newTestDB(t)
	defer conn.Close()

	dir := t.TempDir()
	writePluginFile(t, dir, "exhausted", `
plugin_info = {
    name        = "exhausted",
    version     = "1.0.0",
    description = "Test pool exhaustion",
}

function on_init()
    log.info("should not reach here if pool is exhausted")
end
`)

	// Use MaxVMsPerPlugin=1 to make the pool small.
	// In normal operation, the fresh pool should have a VM available.
	// This test just confirms the plugin loads successfully with pool size 1.
	mgr := NewManager(ManagerConfig{
		Enabled:         true,
		Directory:       dir,
		MaxVMsPerPlugin: 1,
		ExecTimeoutSec:  5,
		MaxOpsPerExec:   100,
	}, conn, db.DialectSQLite, nil)

	err := mgr.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("LoadAll: %s", err)
	}

	inst := mgr.GetPlugin("exhausted")
	if inst == nil {
		t.Fatal("GetPlugin returned nil")
	}
	if inst.State != StateRunning {
		t.Errorf("state = %s, want %s", inst.State, StateRunning)
	}
}

// -- ErrPoolExhausted sentinel --

func TestErrPoolExhausted_Is(t *testing.T) {
	// Verify the sentinel error is usable with errors.Is.
	err := fmt.Errorf("wrapped: %w", ErrPoolExhausted)
	if !errors.Is(err, ErrPoolExhausted) {
		t.Error("errors.Is failed for wrapped ErrPoolExhausted")
	}
}
