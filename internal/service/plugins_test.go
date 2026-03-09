package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/plugin"
	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// Mock PluginManager
// ---------------------------------------------------------------------------

type mockPluginManager struct {
	plugins    map[string]*plugin.PluginInstance
	enableErr  map[string]error
	disableErr map[string]error
	reloadErr  map[string]error

	orphanedTables []string
	droppedTables  []string
	pipelines      *[]db.Pipeline
	dryRunResults  []plugin.DryRunResult

	bridge        *plugin.HTTPBridge
	hookEngine    *plugin.HookEngine
	requestEngine *plugin.RequestEngine
}

func newMockPluginManager() *mockPluginManager {
	return &mockPluginManager{
		plugins:    make(map[string]*plugin.PluginInstance),
		enableErr:  make(map[string]error),
		disableErr: make(map[string]error),
		reloadErr:  make(map[string]error),
	}
}

func (m *mockPluginManager) InstallPlugin(_ context.Context, name string) (*db.Plugin, error) {
	if _, ok := m.plugins[name]; ok {
		return nil, errors.New("already installed")
	}
	return &db.Plugin{Name: name}, nil
}

func (m *mockPluginManager) EnablePlugin(_ context.Context, name string) error {
	if err, ok := m.enableErr[name]; ok {
		return err
	}
	if _, ok := m.plugins[name]; !ok {
		return errors.New("plugin not found: " + name)
	}
	return nil
}

func (m *mockPluginManager) DisablePlugin(_ context.Context, name string) error {
	if err, ok := m.disableErr[name]; ok {
		return err
	}
	if _, ok := m.plugins[name]; !ok {
		return errors.New("plugin not found: " + name)
	}
	return nil
}

func (m *mockPluginManager) ReloadPlugin(_ context.Context, name string) error {
	if err, ok := m.reloadErr[name]; ok {
		return err
	}
	if _, ok := m.plugins[name]; !ok {
		return errors.New("plugin not found: " + name)
	}
	return nil
}

func (m *mockPluginManager) ActivatePlugin(_ context.Context, _ string, _ string) error {
	return nil
}

func (m *mockPluginManager) DeactivatePlugin(_ context.Context, _ string) error {
	return nil
}

func (m *mockPluginManager) GetPlugin(name string) *plugin.PluginInstance {
	return m.plugins[name]
}

func (m *mockPluginManager) ListPlugins() []*plugin.PluginInstance {
	result := make([]*plugin.PluginInstance, 0, len(m.plugins))
	for _, inst := range m.plugins {
		result = append(result, inst)
	}
	return result
}

func (m *mockPluginManager) GetPluginState(name string) (plugin.PluginState, bool) {
	inst, ok := m.plugins[name]
	if !ok {
		return 0, false
	}
	return inst.State, true
}

func (m *mockPluginManager) PluginHealth() plugin.PluginHealthStatus {
	return plugin.PluginHealthStatus{Healthy: true, TotalPlugins: len(m.plugins)}
}

func (m *mockPluginManager) ListDiscovered() []string {
	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}

func (m *mockPluginManager) SyncCapabilities(_ context.Context, name string, _ string) error {
	if _, ok := m.plugins[name]; !ok {
		return errors.New("plugin not found: " + name)
	}
	return nil
}

func (m *mockPluginManager) ListAllPipelines() (*[]db.Pipeline, error) {
	return m.pipelines, nil
}

func (m *mockPluginManager) DryRunPipeline(_, _ string) []plugin.DryRunResult {
	return m.dryRunResults
}

func (m *mockPluginManager) DryRunAllPipelines() []plugin.DryRunResult {
	return m.dryRunResults
}

func (m *mockPluginManager) ListOrphanedTables(_ context.Context) ([]string, error) {
	return m.orphanedTables, nil
}

func (m *mockPluginManager) DropOrphanedTables(_ context.Context, _ []string) ([]string, error) {
	return m.droppedTables, nil
}

func (m *mockPluginManager) Bridge() *plugin.HTTPBridge {
	return m.bridge
}

func (m *mockPluginManager) HookEngine() *plugin.HookEngine {
	return m.hookEngine
}

func (m *mockPluginManager) RequestEngine() *plugin.RequestEngine {
	return m.requestEngine
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPluginService_List_NoPlugins(t *testing.T) {
	mgr := newMockPluginManager()
	svc := service.NewPluginService(mgr)

	summaries, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestPluginService_List_WithPlugins(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.plugins["test-plugin"] = &plugin.PluginInstance{
		Info: plugin.PluginInfo{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "A test plugin",
		},
		State: plugin.StateRunning,
	}
	mgr.plugins["other-plugin"] = &plugin.PluginInstance{
		Info: plugin.PluginInfo{
			Name:        "other-plugin",
			Version:     "2.0.0",
			Description: "Another plugin",
		},
		State: plugin.StateStopped,
	}
	svc := service.NewPluginService(mgr)

	summaries, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}

	// Verify that each plugin is present (order may vary due to map iteration).
	found := map[string]bool{}
	for _, s := range summaries {
		found[s.Name] = true
	}
	if !found["test-plugin"] || !found["other-plugin"] {
		t.Fatalf("expected both plugins in summaries, got: %v", summaries)
	}
}

func TestPluginService_Get_Existing(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.plugins["my-plugin"] = &plugin.PluginInstance{
		Info: plugin.PluginInfo{
			Name:         "my-plugin",
			Version:      "1.2.3",
			Description:  "My cool plugin",
			Author:       "Test Author",
			License:      "MIT",
			Dependencies: []string{"dep-a", "dep-b"},
		},
		State:        plugin.StateRunning,
		FailedReason: "",
	}
	svc := service.NewPluginService(mgr)

	detail, err := svc.Get(context.Background(), "my-plugin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Name != "my-plugin" {
		t.Errorf("expected name 'my-plugin', got %q", detail.Name)
	}
	if detail.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", detail.Version)
	}
	if detail.Author != "Test Author" {
		t.Errorf("expected author 'Test Author', got %q", detail.Author)
	}
	if detail.License != "MIT" {
		t.Errorf("expected license 'MIT', got %q", detail.License)
	}
	if detail.State != "running" {
		t.Errorf("expected state 'running', got %q", detail.State)
	}
	if len(detail.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(detail.Dependencies))
	}
	if detail.SchemaDrift {
		t.Error("expected no schema drift")
	}
}

func TestPluginService_Get_NonExistent(t *testing.T) {
	mgr := newMockPluginManager()
	svc := service.NewPluginService(mgr)

	_, err := svc.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsNotFound(err) {
		t.Fatalf("expected NotFoundError, got: %v", err)
	}
}

func TestPluginService_Enable_Delegates(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.plugins["test"] = &plugin.PluginInstance{
		Info:  plugin.PluginInfo{Name: "test"},
		State: plugin.StateStopped,
	}
	svc := service.NewPluginService(mgr)

	err := svc.Enable(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPluginService_Enable_NonExistent(t *testing.T) {
	mgr := newMockPluginManager()
	svc := service.NewPluginService(mgr)

	err := svc.Enable(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsNotFound(err) {
		t.Fatalf("expected NotFoundError, got: %v", err)
	}
}

func TestPluginService_Enable_AlreadyEnabled(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.plugins["test"] = &plugin.PluginInstance{
		Info:  plugin.PluginInfo{Name: "test"},
		State: plugin.StateRunning,
	}
	mgr.enableErr["test"] = errors.New("already enabled")
	svc := service.NewPluginService(mgr)

	err := svc.Enable(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsConflict(err) {
		t.Fatalf("expected ConflictError, got: %v", err)
	}
}

func TestPluginService_Disable_Delegates(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.plugins["test"] = &plugin.PluginInstance{
		Info:  plugin.PluginInfo{Name: "test"},
		State: plugin.StateRunning,
	}
	svc := service.NewPluginService(mgr)

	err := svc.Disable(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPluginService_Reload_Delegates(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.plugins["test"] = &plugin.PluginInstance{
		Info:  plugin.PluginInfo{Name: "test"},
		State: plugin.StateRunning,
	}
	svc := service.NewPluginService(mgr)

	err := svc.Reload(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPluginService_CleanupDryRun(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.orphanedTables = []string{"plugin_old_table_1", "plugin_old_table_2"}
	svc := service.NewPluginService(mgr)

	tables, err := svc.CleanupDryRun(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(tables))
	}
}

func TestPluginService_CleanupDrop(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.droppedTables = []string{"plugin_old_table_1"}
	svc := service.NewPluginService(mgr)

	dropped, err := svc.CleanupDrop(context.Background(), []string{"plugin_old_table_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dropped) != 1 {
		t.Fatalf("expected 1 dropped table, got %d", len(dropped))
	}
}

func TestPluginService_CleanupDrop_EmptyTables(t *testing.T) {
	mgr := newMockPluginManager()
	svc := service.NewPluginService(mgr)

	_, err := svc.CleanupDrop(context.Background(), []string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Nil Manager Tests
// ---------------------------------------------------------------------------

func TestPluginService_NilManager_List(t *testing.T) {
	svc := service.NewPluginService(nil)

	summaries, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestPluginService_NilManager_Get(t *testing.T) {
	svc := service.NewPluginService(nil)

	_, err := svc.Get(context.Background(), "anything")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsNotFound(err) {
		t.Fatalf("expected NotFoundError, got: %v", err)
	}
}

func TestPluginService_NilManager_Enable(t *testing.T) {
	svc := service.NewPluginService(nil)

	err := svc.Enable(context.Background(), "anything")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_NilManager_Disable(t *testing.T) {
	svc := service.NewPluginService(nil)

	err := svc.Disable(context.Background(), "anything")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_NilManager_Reload(t *testing.T) {
	svc := service.NewPluginService(nil)

	err := svc.Reload(context.Background(), "anything")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_NilManager_Install(t *testing.T) {
	svc := service.NewPluginService(nil)

	_, err := svc.Install(context.Background(), "anything")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_NilManager_Health(t *testing.T) {
	svc := service.NewPluginService(nil)

	health, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !health.Healthy {
		t.Error("expected healthy status when plugins disabled")
	}
}

func TestPluginService_NilManager_CleanupDryRun(t *testing.T) {
	svc := service.NewPluginService(nil)

	tables, err := svc.CleanupDryRun(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tables) != 0 {
		t.Fatalf("expected 0 tables, got %d", len(tables))
	}
}

func TestPluginService_NilManager_CleanupDrop(t *testing.T) {
	svc := service.NewPluginService(nil)

	_, err := svc.CleanupDrop(context.Background(), []string{"some_table"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_NilManager_ListRoutes(t *testing.T) {
	svc := service.NewPluginService(nil)

	routes, err := svc.ListRoutes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(routes) != 0 {
		t.Fatalf("expected 0 routes, got %d", len(routes))
	}
}

func TestPluginService_NilManager_ApproveRoutes(t *testing.T) {
	svc := service.NewPluginService(nil)

	err := svc.ApproveRoutes(context.Background(), []service.RouteApprovalInput{{Plugin: "p", Method: "GET", Path: "/"}}, "admin")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_NilManager_ListHooks(t *testing.T) {
	svc := service.NewPluginService(nil)

	hooks, err := svc.ListHooks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hooks) != 0 {
		t.Fatalf("expected 0 hooks, got %d", len(hooks))
	}
}

func TestPluginService_NilManager_ApproveHooks(t *testing.T) {
	svc := service.NewPluginService(nil)

	err := svc.ApproveHooks(context.Background(), []service.HookApprovalInput{{PluginName: "p", Event: "e", Table: "t"}}, "admin")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_NilManager_ListRequests(t *testing.T) {
	svc := service.NewPluginService(nil)

	requests, err := svc.ListRequests(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("expected 0 requests, got %d", len(requests))
	}
}

func TestPluginService_NilManager_ApproveRequests(t *testing.T) {
	svc := service.NewPluginService(nil)

	err := svc.ApproveRequests(context.Background(), []service.RequestApprovalInput{{PluginName: "p", Domain: "d"}}, "admin")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_NilManager_ListPipelines(t *testing.T) {
	svc := service.NewPluginService(nil)

	pipelines, err := svc.ListPipelines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipelines) != 0 {
		t.Fatalf("expected 0 pipelines, got %d", len(pipelines))
	}
}

func TestPluginService_NilManager_DryRunPipelines(t *testing.T) {
	svc := service.NewPluginService(nil)

	results, err := svc.DryRunPipelines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestPluginService_Enable_EmptyName(t *testing.T) {
	mgr := newMockPluginManager()
	svc := service.NewPluginService(mgr)

	err := svc.Enable(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_Disable_EmptyName(t *testing.T) {
	mgr := newMockPluginManager()
	svc := service.NewPluginService(mgr)

	err := svc.Disable(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_Reload_EmptyName(t *testing.T) {
	mgr := newMockPluginManager()
	svc := service.NewPluginService(mgr)

	err := svc.Reload(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Fatalf("expected ValidationError, got: %v", err)
	}
}

func TestPluginService_Get_SchemaDrift(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.plugins["drifty"] = &plugin.PluginInstance{
		Info: plugin.PluginInfo{
			Name:    "drifty",
			Version: "1.0.0",
		},
		State: plugin.StateRunning,
		SchemaDrift: []plugin.DriftEntry{
			{Table: "plugin_drifty_data", Kind: "missing", Column: "new_col"},
		},
	}
	svc := service.NewPluginService(mgr)

	detail, err := svc.Get(context.Background(), "drifty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !detail.SchemaDrift {
		t.Error("expected schema drift to be true")
	}
}

func TestPluginService_Health_WithPlugins(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.plugins["p1"] = &plugin.PluginInstance{
		Info:  plugin.PluginInfo{Name: "p1"},
		State: plugin.StateRunning,
	}
	svc := service.NewPluginService(mgr)

	health, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.TotalPlugins != 1 {
		t.Errorf("expected 1 total plugin, got %d", health.TotalPlugins)
	}
}

func TestPluginService_ListPipelines_WithData(t *testing.T) {
	mgr := newMockPluginManager()
	pipelines := []db.Pipeline{{PluginName: "p1"}}
	mgr.pipelines = &pipelines
	svc := service.NewPluginService(mgr)

	result, err := svc.ListPipelines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 pipeline, got %d", len(result))
	}
}

func TestPluginService_DryRunPipelines_WithData(t *testing.T) {
	mgr := newMockPluginManager()
	mgr.dryRunResults = []plugin.DryRunResult{
		{Table: "content", Operation: "create", Phase: "before"},
	}
	svc := service.NewPluginService(mgr)

	results, err := svc.DryRunPipelines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
