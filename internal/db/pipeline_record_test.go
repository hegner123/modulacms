// Integration tests for the pipeline entity CRUD lifecycle.
// Uses testIntegrationDB (Tier 0.5: requires plugin for FK dependency).
package db

import (
	"encoding/json"
	"testing"

	"github.com/hegner123/modulacms/internal/db/types"
)

// createTestPlugin is a helper that inserts a plugin and returns it.
// The caller must have already created the plugins table via testIntegrationDB.
func createTestPlugin(t *testing.T, d Database) *Plugin {
	t.Helper()
	ctx := d.Context
	ac := testAuditCtx(d)

	p, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
		Name:           "pipeline-test-plugin",
		Version:        "1.0.0",
		Description:    "plugin for pipeline FK tests",
		Author:         "tester",
		Status:         types.PluginStatusInstalled,
		Capabilities:   "[]",
		ApprovedAccess: "{}",
		ManifestHash:   "pipelinehash",
	})
	if err != nil {
		t.Fatalf("seed CreatePlugin: %v", err)
	}
	return p
}

func TestDatabase_CRUD_Pipeline(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)
	plugin := createTestPlugin(t, d)

	// --- Count: starts at zero ---
	count, err := d.CountPipelines()
	if err != nil {
		t.Fatalf("initial CountPipelines: %v", err)
	}
	if *count != 0 {
		t.Fatalf("initial CountPipelines = %d, want 0", *count)
	}

	// --- Create ---
	created, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
		PluginID:   plugin.PluginID,
		TableName:  "content_fields",
		Operation:  "before_create",
		PluginName: "pipeline-test-plugin",
		Handler:    "validate",
		Priority:   10,
		Enabled:    true,
		Config:     types.NewJSONData(json.RawMessage("{}")),
	})
	if err != nil {
		t.Fatalf("CreatePipeline: %v", err)
	}
	if created == nil {
		t.Fatal("CreatePipeline returned nil")
	}
	if created.PipelineID.IsZero() {
		t.Fatal("CreatePipeline returned zero PipelineID")
	}
	if created.PluginID != plugin.PluginID {
		t.Errorf("PluginID = %v, want %v", created.PluginID, plugin.PluginID)
	}
	if created.TableName != "content_fields" {
		t.Errorf("TableName = %q, want %q", created.TableName, "content_fields")
	}
	if created.Operation != "before_create" {
		t.Errorf("Operation = %q, want %q", created.Operation, "before_create")
	}
	if created.PluginName != "pipeline-test-plugin" {
		t.Errorf("PluginName = %q, want %q", created.PluginName, "pipeline-test-plugin")
	}
	if created.Handler != "validate" {
		t.Errorf("Handler = %q, want %q", created.Handler, "validate")
	}
	if created.Priority != 10 {
		t.Errorf("Priority = %d, want %d", created.Priority, 10)
	}
	if !created.Enabled {
		t.Error("Enabled = false, want true")
	}
	if created.DateCreated.Time.IsZero() {
		t.Error("DateCreated is zero")
	}
	if created.DateModified.Time.IsZero() {
		t.Error("DateModified is zero")
	}

	// --- Get ---
	got, err := d.GetPipeline(created.PipelineID)
	if err != nil {
		t.Fatalf("GetPipeline: %v", err)
	}
	if got == nil {
		t.Fatal("GetPipeline returned nil")
	}
	if got.PipelineID != created.PipelineID {
		t.Errorf("GetPipeline PipelineID = %v, want %v", got.PipelineID, created.PipelineID)
	}
	if got.PluginID != plugin.PluginID {
		t.Errorf("GetPipeline PluginID = %v, want %v", got.PluginID, plugin.PluginID)
	}
	if got.TableName != "content_fields" {
		t.Errorf("GetPipeline TableName = %q, want %q", got.TableName, "content_fields")
	}
	if got.Handler != "validate" {
		t.Errorf("GetPipeline Handler = %q, want %q", got.Handler, "validate")
	}
	if got.Priority != 10 {
		t.Errorf("GetPipeline Priority = %d, want %d", got.Priority, 10)
	}
	if !got.Enabled {
		t.Error("GetPipeline Enabled = false, want true")
	}

	// --- List ---
	list, err := d.ListPipelines()
	if err != nil {
		t.Fatalf("ListPipelines: %v", err)
	}
	if list == nil {
		t.Fatal("ListPipelines returned nil")
	}
	if len(*list) != 1 {
		t.Fatalf("ListPipelines len = %d, want 1", len(*list))
	}
	if (*list)[0].PipelineID != created.PipelineID {
		t.Errorf("ListPipelines[0].PipelineID = %v, want %v", (*list)[0].PipelineID, created.PipelineID)
	}

	// --- Count: now 1 ---
	count, err = d.CountPipelines()
	if err != nil {
		t.Fatalf("CountPipelines after create: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountPipelines after create = %d, want 1", *count)
	}

	// --- Update ---
	err = d.UpdatePipeline(ctx, ac, UpdatePipelineParams{
		PipelineID: created.PipelineID,
		TableName:  "content_data",
		Operation:  "after_update",
		Handler:    "transform",
		Priority:   20,
		Enabled:    false,
		Config:     types.NewJSONData(json.RawMessage(`{"key":"value"}`)),
	})
	if err != nil {
		t.Fatalf("UpdatePipeline: %v", err)
	}

	// --- Get after update ---
	updated, err := d.GetPipeline(created.PipelineID)
	if err != nil {
		t.Fatalf("GetPipeline after update: %v", err)
	}
	if updated.TableName != "content_data" {
		t.Errorf("updated TableName = %q, want %q", updated.TableName, "content_data")
	}
	if updated.Operation != "after_update" {
		t.Errorf("updated Operation = %q, want %q", updated.Operation, "after_update")
	}
	if updated.Handler != "transform" {
		t.Errorf("updated Handler = %q, want %q", updated.Handler, "transform")
	}
	if updated.Priority != 20 {
		t.Errorf("updated Priority = %d, want %d", updated.Priority, 20)
	}
	if updated.Enabled {
		t.Error("updated Enabled = true, want false")
	}
	// PluginID should remain unchanged
	if updated.PluginID != plugin.PluginID {
		t.Errorf("updated PluginID = %v, want %v (should not change)", updated.PluginID, plugin.PluginID)
	}

	// --- Delete ---
	err = d.DeletePipeline(ctx, ac, created.PipelineID)
	if err != nil {
		t.Fatalf("DeletePipeline: %v", err)
	}

	// --- Get after delete: expect error ---
	_, err = d.GetPipeline(created.PipelineID)
	if err == nil {
		t.Fatal("GetPipeline after delete: expected error, got nil")
	}

	// --- Count: back to zero ---
	count, err = d.CountPipelines()
	if err != nil {
		t.Fatalf("CountPipelines after delete: %v", err)
	}
	if *count != 0 {
		t.Fatalf("CountPipelines after delete = %d, want 0", *count)
	}
}

func TestDatabase_CRUD_Pipeline_ListByTable(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)
	plugin := createTestPlugin(t, d)

	tables := []string{"content_fields", "content_data", "content_fields"}
	for i, tbl := range tables {
		_, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
			PluginID:   plugin.PluginID,
			TableName:  tbl,
			Operation:  "before_create",
			PluginName: "pipeline-test-plugin",
			Handler:    "handler-" + tbl,
			Priority:   int(i + 1),
			Enabled:    true,
			Config:     types.NewJSONData(json.RawMessage("{}")),
		})
		if err != nil {
			t.Fatalf("CreatePipeline[%d]: %v", i, err)
		}
	}

	byTable, err := d.ListPipelinesByTable("content_fields")
	if err != nil {
		t.Fatalf("ListPipelinesByTable: %v", err)
	}
	if byTable == nil {
		t.Fatal("ListPipelinesByTable returned nil")
	}
	if len(*byTable) != 2 {
		t.Fatalf("ListPipelinesByTable(content_fields) len = %d, want 2", len(*byTable))
	}
	for _, p := range *byTable {
		if p.TableName != "content_fields" {
			t.Errorf("unexpected TableName = %q, want %q", p.TableName, "content_fields")
		}
	}

	byTableData, err := d.ListPipelinesByTable("content_data")
	if err != nil {
		t.Fatalf("ListPipelinesByTable(content_data): %v", err)
	}
	if len(*byTableData) != 1 {
		t.Fatalf("ListPipelinesByTable(content_data) len = %d, want 1", len(*byTableData))
	}
}

func TestDatabase_CRUD_Pipeline_ListByPluginID(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// Create two distinct plugins
	p1 := createTestPlugin(t, d)

	p2, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
		Name:           "second-plugin",
		Version:        "1.0.0",
		Description:    "Second plugin",
		Author:         "tester",
		Status:         types.PluginStatusInstalled,
		Capabilities:   "[]",
		ApprovedAccess: "{}",
		ManifestHash:   "hash2",
	})
	if err != nil {
		t.Fatalf("CreatePlugin (second): %v", err)
	}

	// 2 pipelines for p1, 1 for p2
	for i := range 2 {
		_, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
			PluginID:   p1.PluginID,
			TableName:  "content_data",
			Operation:  "before_create",
			PluginName: "pipeline-test-plugin",
			Handler:    "handler-p1-" + string(rune('a'+i)),
			Priority:   i + 1,
			Enabled:    true,
			Config:     types.NewJSONData(json.RawMessage("{}")),
		})
		if err != nil {
			t.Fatalf("CreatePipeline p1[%d]: %v", i, err)
		}
	}

	_, err = d.CreatePipeline(ctx, ac, CreatePipelineParams{
		PluginID:   p2.PluginID,
		TableName:  "content_data",
		Operation:  "after_create",
		PluginName: "second-plugin",
		Handler:    "handler-p2",
		Priority:   1,
		Enabled:    true,
		Config:     types.NewJSONData(json.RawMessage("{}")),
	})
	if err != nil {
		t.Fatalf("CreatePipeline p2: %v", err)
	}

	byP1, err := d.ListPipelinesByPluginID(p1.PluginID)
	if err != nil {
		t.Fatalf("ListPipelinesByPluginID(p1): %v", err)
	}
	if len(*byP1) != 2 {
		t.Fatalf("ListPipelinesByPluginID(p1) len = %d, want 2", len(*byP1))
	}
	for _, p := range *byP1 {
		if p.PluginID != p1.PluginID {
			t.Errorf("unexpected PluginID = %v, want %v", p.PluginID, p1.PluginID)
		}
	}

	byP2, err := d.ListPipelinesByPluginID(p2.PluginID)
	if err != nil {
		t.Fatalf("ListPipelinesByPluginID(p2): %v", err)
	}
	if len(*byP2) != 1 {
		t.Fatalf("ListPipelinesByPluginID(p2) len = %d, want 1", len(*byP2))
	}
}

func TestDatabase_CRUD_Pipeline_ListByTableOperation(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)
	plugin := createTestPlugin(t, d)

	combos := []struct {
		table string
		op    string
	}{
		{"content_data", "before_create"},
		{"content_data", "after_create"},
		{"content_data", "before_create"},
		{"content_fields", "before_create"},
	}

	for i, c := range combos {
		_, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
			PluginID:   plugin.PluginID,
			TableName:  c.table,
			Operation:  c.op,
			PluginName: "pipeline-test-plugin",
			Handler:    "handler",
			Priority:   i + 1,
			Enabled:    true,
			Config:     types.NewJSONData(json.RawMessage("{}")),
		})
		if err != nil {
			t.Fatalf("CreatePipeline[%d]: %v", i, err)
		}
	}

	byTableOp, err := d.ListPipelinesByTableOperation("content_data", "before_create")
	if err != nil {
		t.Fatalf("ListPipelinesByTableOperation: %v", err)
	}
	if byTableOp == nil {
		t.Fatal("ListPipelinesByTableOperation returned nil")
	}
	if len(*byTableOp) != 2 {
		t.Fatalf("ListPipelinesByTableOperation len = %d, want 2", len(*byTableOp))
	}
	for _, p := range *byTableOp {
		if p.TableName != "content_data" {
			t.Errorf("unexpected TableName = %q", p.TableName)
		}
		if p.Operation != "before_create" {
			t.Errorf("unexpected Operation = %q", p.Operation)
		}
	}

	// after_create for content_data should return 1
	byAfter, err := d.ListPipelinesByTableOperation("content_data", "after_create")
	if err != nil {
		t.Fatalf("ListPipelinesByTableOperation(after_create): %v", err)
	}
	if len(*byAfter) != 1 {
		t.Fatalf("ListPipelinesByTableOperation(after_create) len = %d, want 1", len(*byAfter))
	}
}

func TestDatabase_CRUD_Pipeline_ListEnabled(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)
	plugin := createTestPlugin(t, d)

	// Create 2 enabled, 1 disabled
	for i, enabled := range []bool{true, true, false} {
		_, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
			PluginID:   plugin.PluginID,
			TableName:  "content_data",
			Operation:  "before_create",
			PluginName: "pipeline-test-plugin",
			Handler:    "handler",
			Priority:   i + 1,
			Enabled:    enabled,
			Config:     types.NewJSONData(json.RawMessage("{}")),
		})
		if err != nil {
			t.Fatalf("CreatePipeline[%d]: %v", i, err)
		}
	}

	enabledList, err := d.ListEnabledPipelines()
	if err != nil {
		t.Fatalf("ListEnabledPipelines: %v", err)
	}
	if enabledList == nil {
		t.Fatal("ListEnabledPipelines returned nil")
	}
	if len(*enabledList) != 2 {
		t.Fatalf("ListEnabledPipelines len = %d, want 2", len(*enabledList))
	}
	for _, p := range *enabledList {
		if !p.Enabled {
			t.Errorf("ListEnabledPipelines returned disabled pipeline %v", p.PipelineID)
		}
	}
}

func TestDatabase_CRUD_Pipeline_UpdateEnabled(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)
	plugin := createTestPlugin(t, d)

	created, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
		PluginID:   plugin.PluginID,
		TableName:  "content_data",
		Operation:  "before_create",
		PluginName: "pipeline-test-plugin",
		Handler:    "validate",
		Priority:   10,
		Enabled:    true,
		Config:     types.NewJSONData(json.RawMessage("{}")),
	})
	if err != nil {
		t.Fatalf("CreatePipeline: %v", err)
	}
	if !created.Enabled {
		t.Fatal("initial Enabled = false, want true")
	}

	// Disable
	err = d.UpdatePipelineEnabled(ctx, ac, created.PipelineID, false)
	if err != nil {
		t.Fatalf("UpdatePipelineEnabled(false): %v", err)
	}

	got, err := d.GetPipeline(created.PipelineID)
	if err != nil {
		t.Fatalf("GetPipeline after disable: %v", err)
	}
	if got.Enabled {
		t.Error("Enabled = true after UpdatePipelineEnabled(false), want false")
	}

	// Verify other fields are unchanged
	if got.Handler != "validate" {
		t.Errorf("Handler changed after toggle: got %q, want %q", got.Handler, "validate")
	}
	if got.Priority != 10 {
		t.Errorf("Priority changed after toggle: got %d, want %d", got.Priority, 10)
	}

	// Re-enable
	err = d.UpdatePipelineEnabled(ctx, ac, created.PipelineID, true)
	if err != nil {
		t.Fatalf("UpdatePipelineEnabled(true): %v", err)
	}

	got2, err := d.GetPipeline(created.PipelineID)
	if err != nil {
		t.Fatalf("GetPipeline after re-enable: %v", err)
	}
	if !got2.Enabled {
		t.Error("Enabled = false after UpdatePipelineEnabled(true), want true")
	}
}

func TestDatabase_CRUD_Pipeline_CountEmpty(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)

	count, err := d.CountPipelines()
	if err != nil {
		t.Fatalf("CountPipelines: %v", err)
	}
	if *count != 0 {
		t.Errorf("CountPipelines = %d, want 0", *count)
	}
}

func TestDatabase_CRUD_Pipeline_DeleteByPluginID(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)

	// Create two plugins
	p1 := createTestPlugin(t, d)

	p2, err := d.CreatePlugin(ctx, ac, CreatePluginParams{
		Name:           "delete-cascade-plugin",
		Version:        "1.0.0",
		Description:    "plugin for cascade delete",
		Author:         "tester",
		Status:         types.PluginStatusInstalled,
		Capabilities:   "[]",
		ApprovedAccess: "{}",
		ManifestHash:   "hash2",
	})
	if err != nil {
		t.Fatalf("CreatePlugin (second): %v", err)
	}

	// 2 pipelines for p1, 1 for p2
	for i := range 2 {
		_, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
			PluginID:   p1.PluginID,
			TableName:  "content_data",
			Operation:  "before_create",
			PluginName: "pipeline-test-plugin",
			Handler:    "handler-p1",
			Priority:   i + 1,
			Enabled:    true,
			Config:     types.NewJSONData(json.RawMessage("{}")),
		})
		if err != nil {
			t.Fatalf("CreatePipeline p1[%d]: %v", i, err)
		}
	}

	p2Pipeline, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
		PluginID:   p2.PluginID,
		TableName:  "content_fields",
		Operation:  "after_create",
		PluginName: "delete-cascade-plugin",
		Handler:    "handler-p2",
		Priority:   1,
		Enabled:    true,
		Config:     types.NewJSONData(json.RawMessage("{}")),
	})
	if err != nil {
		t.Fatalf("CreatePipeline p2: %v", err)
	}

	// Verify 3 total
	count, err := d.CountPipelines()
	if err != nil {
		t.Fatalf("CountPipelines: %v", err)
	}
	if *count != 3 {
		t.Fatalf("CountPipelines = %d, want 3", *count)
	}

	// Delete all pipelines for p1
	err = d.DeletePipelinesByPluginID(ctx, ac, p1.PluginID)
	if err != nil {
		t.Fatalf("DeletePipelinesByPluginID: %v", err)
	}

	// Only p2's pipeline should remain
	count, err = d.CountPipelines()
	if err != nil {
		t.Fatalf("CountPipelines after bulk delete: %v", err)
	}
	if *count != 1 {
		t.Fatalf("CountPipelines after bulk delete = %d, want 1", *count)
	}

	// Verify p2's pipeline is still accessible
	got, err := d.GetPipeline(p2Pipeline.PipelineID)
	if err != nil {
		t.Fatalf("GetPipeline(p2) after bulk delete: %v", err)
	}
	if got.PluginID != p2.PluginID {
		t.Errorf("remaining pipeline PluginID = %v, want %v", got.PluginID, p2.PluginID)
	}

	// Verify p1 has no pipelines
	byP1, err := d.ListPipelinesByPluginID(p1.PluginID)
	if err != nil {
		t.Fatalf("ListPipelinesByPluginID(p1) after bulk delete: %v", err)
	}
	if len(*byP1) != 0 {
		t.Fatalf("ListPipelinesByPluginID(p1) len = %d, want 0", len(*byP1))
	}
}

func TestDatabase_CRUD_Pipeline_ConfigRoundTrip(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)
	plugin := createTestPlugin(t, d)

	configJSON := `{"timeout":30,"retries":3,"tags":["a","b"]}`
	created, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
		PluginID:   plugin.PluginID,
		TableName:  "content_data",
		Operation:  "before_create",
		PluginName: "pipeline-test-plugin",
		Handler:    "validate",
		Priority:   1,
		Enabled:    true,
		Config:     types.NewJSONData(json.RawMessage(configJSON)),
	})
	if err != nil {
		t.Fatalf("CreatePipeline: %v", err)
	}

	got, err := d.GetPipeline(created.PipelineID)
	if err != nil {
		t.Fatalf("GetPipeline: %v", err)
	}

	configBytes, err := json.Marshal(got.Config.Data)
	if err != nil {
		t.Fatalf("marshal Config: %v", err)
	}

	// Parse both to compare structurally (key ordering may differ)
	var expected, actual map[string]any
	if err := json.Unmarshal([]byte(configJSON), &expected); err != nil {
		t.Fatalf("unmarshal expected: %v", err)
	}
	if err := json.Unmarshal(configBytes, &actual); err != nil {
		t.Fatalf("unmarshal actual: %v", err)
	}

	// Compare individual keys since reflect.DeepEqual may be fragile with any
	if actual["timeout"] != expected["timeout"] {
		t.Errorf("Config.timeout = %v, want %v", actual["timeout"], expected["timeout"])
	}
	if actual["retries"] != expected["retries"] {
		t.Errorf("Config.retries = %v, want %v", actual["retries"], expected["retries"])
	}
}

func TestDatabase_CRUD_Pipeline_MultipleRecords(t *testing.T) {
	t.Parallel()
	d := testIntegrationDB(t)
	ctx := d.Context
	ac := testAuditCtx(d)
	plugin := createTestPlugin(t, d)

	ids := make([]types.PipelineID, 3)
	for i := range 3 {
		p, err := d.CreatePipeline(ctx, ac, CreatePipelineParams{
			PluginID:   plugin.PluginID,
			TableName:  "content_data",
			Operation:  "before_create",
			PluginName: "pipeline-test-plugin",
			Handler:    "handler",
			Priority:   i + 1,
			Enabled:    true,
			Config:     types.NewJSONData(json.RawMessage("{}")),
		})
		if err != nil {
			t.Fatalf("CreatePipeline[%d]: %v", i, err)
		}
		ids[i] = p.PipelineID
	}

	count, err := d.CountPipelines()
	if err != nil {
		t.Fatalf("CountPipelines: %v", err)
	}
	if *count != 3 {
		t.Fatalf("CountPipelines = %d, want 3", *count)
	}

	// Delete the middle one
	err = d.DeletePipeline(ctx, ac, ids[1])
	if err != nil {
		t.Fatalf("DeletePipeline: %v", err)
	}

	count, err = d.CountPipelines()
	if err != nil {
		t.Fatalf("CountPipelines after delete: %v", err)
	}
	if *count != 2 {
		t.Fatalf("CountPipelines after delete = %d, want 2", *count)
	}

	// Remaining pipelines should still be accessible
	for _, idx := range []int{0, 2} {
		got, err := d.GetPipeline(ids[idx])
		if err != nil {
			t.Fatalf("GetPipeline(%v) after delete: %v", ids[idx], err)
		}
		if got.PipelineID != ids[idx] {
			t.Errorf("remaining pipeline ID = %v, want %v", got.PipelineID, ids[idx])
		}
	}

	// Deleted pipeline should be gone
	_, err = d.GetPipeline(ids[1])
	if err == nil {
		t.Fatal("GetPipeline for deleted pipeline: expected error, got nil")
	}
}
