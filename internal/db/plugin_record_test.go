// Integration tests for the plugin entity CRUD lifecycle.
// Uses testIntegrationDB (Tier 0: no FK dependencies).
package db

import (
	"encoding/json"
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

func TestDatabase_CRUD_Plugin(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// --- Count: starts at zero ---
	count, err := d.CountPlugins()
	if err != nil {
		t.Fatalf("initial CountPlugins: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountPlugins = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
		Name:           "test-plugin",
		Version:        "1.0.0",
		Description:    "A test plugin",
		Author:         "tester",
		Status:         types.PluginStatusInstalled,
		Capabilities:   "[]",
		ApprovedAccess: "{}",
		ManifestHash:   "abc123",
	})
	if err != nil {
		t.Fatalf("CreatePlugin: %v", err)
	}
	if created == nil {
		t.Fatal("CreatePlugin returned nil")
	}
	if created.PluginID.IsZero() {
		t.Fatal("CreatePlugin returned zero PluginID")
	}
	if created.Name != "test-plugin" {
		t.Errorf("Name = %q, want %q", created.Name, "test-plugin")
	}
	if created.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", created.Version, "1.0.0")
	}
	if created.Description != "A test plugin" {
		t.Errorf("Description = %q, want %q", created.Description, "A test plugin")
	}
	if created.Author != "tester" {
		t.Errorf("Author = %q, want %q", created.Author, "tester")
	}
	if created.Status != types.PluginStatusInstalled {
		t.Errorf("Status = %q, want %q", created.Status, types.PluginStatusInstalled)
	}
	if created.ManifestHash != "abc123" {
		t.Errorf("ManifestHash = %q, want %q", created.ManifestHash, "abc123")
	}
	if created.DateInstalled.Time.IsZero() {
		t.Error("DateInstalled is zero")
	}
	if created.DateModified.Time.IsZero() {
		t.Error("DateModified is zero")
	}

	// --- Get ---
	got, err := d.GetPlugin(created.PluginID)
	if err != nil {
		t.Fatalf("GetPlugin: %v", err)
	}
	if got == nil {
		t.Fatal("GetPlugin returned nil")
	}
	if got.PluginID != created.PluginID {
		t.Errorf("GetPlugin PluginID = %v, want %v", got.PluginID, created.PluginID)
	}
	if got.Name != "test-plugin" {
		t.Errorf("GetPlugin Name = %q, want %q", got.Name, "test-plugin")
	}
	if got.Version != "1.0.0" {
		t.Errorf("GetPlugin Version = %q, want %q", got.Version, "1.0.0")
	}
	if got.Status != types.PluginStatusInstalled {
		t.Errorf("GetPlugin Status = %q, want %q", got.Status, types.PluginStatusInstalled)
	}

	// --- List ---
	list, err := d.ListPlugins()
	if err != nil {
		t.Fatalf("ListPlugins: %v", err)
	}
	if list == nil {
		t.Fatal("ListPlugins returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListPlugins len = %d, want 1", len(*list))
	}
	if (*list)[0].PluginID != created.PluginID {
		t.Errorf("ListPlugins[0].PluginID = %v, want %v", (*list)[0].PluginID, created.PluginID)
	}

	// --- Count: now 1 ---
	count, err = d.CountPlugins()
	if err != nil {
		t.Fatalf("CountPlugins after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountPlugins after create = %d, want 1", *count)
	}

	// --- Update ---
	err = d.UpdatePlugin(ctx, ac, UpdatePluginParams{
		PluginID:       created.PluginID,
		Version:        "2.0.0",
		Description:    "Updated description",
		Author:         "updated-author",
		Status:         types.PluginStatusEnabled,
		Capabilities:   `["cap1"]`,
		ApprovedAccess: `{"db":"read"}`,
		ManifestHash:   "def456",
	})
	if err != nil {
		t.Fatalf("UpdatePlugin: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetPlugin(created.PluginID)
	if err != nil {
		t.Fatalf("GetPlugin after update: %v", err)
	}
	if updated.Version != "2.0.0" {
		t.Errorf("updated Version = %q, want %q", updated.Version, "2.0.0")
	}
	if updated.Description != "Updated description" {
		t.Errorf("updated Description = %q, want %q", updated.Description, "Updated description")
	}
	if updated.Author != "updated-author" {
		t.Errorf("updated Author = %q, want %q", updated.Author, "updated-author")
	}
	if updated.Status != types.PluginStatusEnabled {
		t.Errorf("updated Status = %q, want %q", updated.Status, types.PluginStatusEnabled)
	}
	if updated.ManifestHash != "def456" {
		t.Errorf("updated ManifestHash = %q, want %q", updated.ManifestHash, "def456")
	}

	// --- Delete ---
	err = d.DeletePlugin(ctx, ac, created.PluginID)
	if err != nil {
		t.Fatalf("DeletePlugin: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetPlugin(created.PluginID)
	if err == nil {
		t.Fatal("GetPlugin after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountPlugins()
	if err != nil {
		t.Fatalf("CountPlugins after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountPlugins after delete = %d, want 0", *count)
	}
}

func TestDatabase_CRUD_Plugin_GetByName(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	created, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
		Name:           "named-plugin",
		Version:        "1.0.0",
		Description:    "plugin found by name",
		Author:         "tester",
		Status:         types.PluginStatusInstalled,
		Capabilities:   "[]",
		ApprovedAccess: "{}",
		ManifestHash:   "hash1",
	})
	if err != nil {
		t.Fatalf("CreatePlugin: %v", err)
	}

	got, err := d.GetPluginByName("named-plugin")
	if err != nil {
		t.Fatalf("GetPluginByName: %v", err)
	}
	if got == nil {
		t.Fatal("GetPluginByName returned nil")
	}
	if got.PluginID != created.PluginID {
		t.Errorf("PluginID = %v, want %v", got.PluginID, created.PluginID)
	}
	if got.Name != "named-plugin" {
		t.Errorf("Name = %q, want %q", got.Name, "named-plugin")
	}
	if got.Description != "plugin found by name" {
		t.Errorf("Description = %q, want %q", got.Description, "plugin found by name")
	}
}

func TestDatabase_CRUD_Plugin_GetByName_NotFound(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	_, err := d.GetPluginByName("nonexistent")
	if err == nil {
		t.Fatal("GetPluginByName for nonexistent: expected error, got nil")
	}
}

func TestDatabase_CRUD_Plugin_ListByStatus(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// Create two installed, one enabled
	for _, name := range []string{"plugin-a", "plugin-b"} {
		_, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
			Name:           name,
			Version:        "1.0.0",
			Description:    "installed plugin",
			Author:         "tester",
			Status:         types.PluginStatusInstalled,
			Capabilities:   "[]",
			ApprovedAccess: "{}",
			ManifestHash:   "hash",
		})
		if err != nil {
			t.Fatalf("CreatePlugin(%s): %v", name, err)
		}
	}

	_, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
		Name:           "plugin-c",
		Version:        "1.0.0",
		Description:    "enabled plugin",
		Author:         "tester",
		Status:         types.PluginStatusEnabled,
		Capabilities:   "[]",
		ApprovedAccess: "{}",
		ManifestHash:   "hash",
	})
	if err != nil {
		t.Fatalf("CreatePlugin(plugin-c): %v", err)
	}

	// ListPluginsByStatus "installed" should return 2
	installed, err := d.ListPluginsByStatus(types.PluginStatusInstalled)
	if err != nil {
		t.Fatalf("ListPluginsByStatus(installed): %v", err)
	}
	if installed == nil {
		t.Fatal("ListPluginsByStatus(installed) returned nil")
	}
	if len(*installed) != 2 {
		t.Fatalf("ListPluginsByStatus(installed) len = %d, want 2", len(*installed))
	}
	for _, p := range *installed {
		if p.Status != types.PluginStatusInstalled {
			t.Errorf("expected status installed, got %q for plugin %q", p.Status, p.Name)
		}
	}

	// ListPluginsByStatus "enabled" should return 1
	enabled, err := d.ListPluginsByStatus(types.PluginStatusEnabled)
	if err != nil {
		t.Fatalf("ListPluginsByStatus(enabled): %v", err)
	}
	if enabled == nil {
		t.Fatal("ListPluginsByStatus(enabled) returned nil")
	}
	if len(*enabled) != 1 {
		t.Fatalf("ListPluginsByStatus(enabled) len = %d, want 1", len(*enabled))
	}
	if (*enabled)[0].Name != "plugin-c" {
		t.Errorf("enabled plugin Name = %q, want %q", (*enabled)[0].Name, "plugin-c")
	}
}

func TestDatabase_CRUD_Plugin_UpdateStatus(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	created, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
		Name:           "status-change-plugin",
		Version:        "1.0.0",
		Description:    "plugin to test status change",
		Author:         "tester",
		Status:         types.PluginStatusInstalled,
		Capabilities:   "[]",
		ApprovedAccess: "{}",
		ManifestHash:   "hash1",
	})
	if err != nil {
		t.Fatalf("CreatePlugin: %v", err)
	}
	if created.Status != types.PluginStatusInstalled {
		t.Fatalf("initial Status = %q, want %q", created.Status, types.PluginStatusInstalled)
	}

	// UpdatePluginStatus to enabled
	err = d.UpdatePluginStatus(ctx, ac, created.PluginID, types.PluginStatusEnabled)
	if err != nil {
		t.Fatalf("UpdatePluginStatus: %v", err)
	}

	got, err := d.GetPlugin(created.PluginID)
	if err != nil {
		t.Fatalf("GetPlugin after UpdatePluginStatus: %v", err)
	}
	if got.Status != types.PluginStatusEnabled {
		t.Errorf("Status after update = %q, want %q", got.Status, types.PluginStatusEnabled)
	}

	// Verify other fields were NOT changed
	if got.Name != "status-change-plugin" {
		t.Errorf("Name changed after UpdatePluginStatus: got %q, want %q", got.Name, "status-change-plugin")
	}
	if got.Version != "1.0.0" {
		t.Errorf("Version changed after UpdatePluginStatus: got %q, want %q", got.Version, "1.0.0")
	}
	if got.ManifestHash != "hash1" {
		t.Errorf("ManifestHash changed after UpdatePluginStatus: got %q, want %q", got.ManifestHash, "hash1")
	}
}

func TestDatabase_CRUD_Plugin_CountEmpty(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	count, err := d.CountPlugins()
	if err != nil {
		t.Fatalf("CountPlugins: %v", err)
	}
	if *count != 0 {
		t.Errorf("CountPlugins = %d, want 0", *count)
	}
}

func TestDatabase_CRUD_Plugin_MultipleRecords(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	names := []string{"multi-a", "multi-b", "multi-c"}
	ids := make([]types.PluginID, len(names))

	for i, name := range names {
		p, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
			Name:           name,
			Version:        "1.0.0",
			Description:    "Multi record test",
			Author:         "tester",
			Status:         types.PluginStatusInstalled,
			Capabilities:   "[]",
			ApprovedAccess: "{}",
			ManifestHash:   "hash",
		})
		if err != nil {
			t.Fatalf("CreatePlugin(%s): %v", name, err)
		}
		ids[i] = p.PluginID
	}

	count, err := d.CountPlugins()
	if err != nil {
		t.Fatalf("CountPlugins: %v", err)
	}
	if *count != 3 {
		t.Fatalf("CountPlugins = %d, want 3", *count)
	}

	list, err := d.ListPlugins()
	if err != nil {
		t.Fatalf("ListPlugins: %v", err)
	}
	if len(*list) != 3 {
		t.Fatalf("ListPlugins len = %d, want 3", len(*list))
	}

	// Delete the middle one
	err = d.DeletePlugin(ctx, ac, ids[1])
	if err != nil {
		t.Fatalf("DeletePlugin: %v", err)
	}

	count, err = d.CountPlugins()
	if err != nil {
		t.Fatalf("CountPlugins after delete: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountPlugins after delete = %d, want 2", *count)
	}

	// Remaining plugins should still be accessible
	for _, idx := range []int{0, 2} {
		got, err := d.GetPlugin(ids[idx])
		if err != nil {
			t.Fatalf("GetPlugin(%v) after delete: %v", ids[idx], err)
		}
		if got.Name != names[idx] {
			t.Errorf("remaining plugin Name = %q, want %q", got.Name, names[idx])
		}
	}

	// Deleted plugin should be gone
	_, err = d.GetPlugin(ids[1])
	if err == nil {
		t.Fatal("GetPlugin for deleted plugin: expected error, got nil")
	}
}

func TestDatabase_CRUD_Plugin_JSONFields(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// Create with specific JSON capabilities and approved access
	created, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
		Name:           "json-plugin",
		Version:        "1.0.0",
		Description:    "JSON field test",
		Author:         "tester",
		Status:         types.PluginStatusInstalled,
		Capabilities:   `["read","write"]`,
		ApprovedAccess: `{"tables":["content_data"],"operations":["before_create"]}`,
		ManifestHash:   "jsonhash",
	})
	if err != nil {
		t.Fatalf("CreatePlugin: %v", err)
	}

	got, err := d.GetPlugin(created.PluginID)
	if err != nil {
		t.Fatalf("GetPlugin: %v", err)
	}

	// Verify Capabilities JSON round-trips correctly
	capBytes, err := json.Marshal(got.Capabilities.Data)
	if err != nil {
		t.Fatalf("marshal Capabilities: %v", err)
	}
	var capSlice []string
	if err := json.Unmarshal(capBytes, &capSlice); err != nil {
		t.Fatalf("unmarshal Capabilities: %v", err)
	}
	if len(capSlice) != 2 || capSlice[0] != "read" || capSlice[1] != "write" {
		t.Errorf("Capabilities = %v, want [read write]", capSlice)
	}

	// Verify ApprovedAccess JSON round-trips correctly
	accessBytes, err := json.Marshal(got.ApprovedAccess.Data)
	if err != nil {
		t.Fatalf("marshal ApprovedAccess: %v", err)
	}
	var accessMap map[string]any
	if err := json.Unmarshal(accessBytes, &accessMap); err != nil {
		t.Fatalf("unmarshal ApprovedAccess: %v", err)
	}
	tablesRaw, ok := accessMap["tables"]
	if !ok {
		t.Error("ApprovedAccess missing 'tables' key")
	} else {
		tables, ok := tablesRaw.([]any)
		if !ok || len(tables) != 1 || tables[0] != "content_data" {
			t.Errorf("ApprovedAccess.tables = %v, want [content_data]", tablesRaw)
		}
	}
	opsRaw, ok := accessMap["operations"]
	if !ok {
		t.Error("ApprovedAccess missing 'operations' key")
	} else {
		ops, ok := opsRaw.([]any)
		if !ok || len(ops) != 1 || ops[0] != "before_create" {
			t.Errorf("ApprovedAccess.operations = %v, want [before_create]", opsRaw)
		}
	}
}
