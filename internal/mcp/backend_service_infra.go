package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// ---------------------------------------------------------------------------
// TableBackend
// ---------------------------------------------------------------------------

type svcTableBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcTableBackend) ListTables(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Tables.ListTables(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTableBackend) GetTable(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Tables.GetTable(ctx, id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTableBackend) CreateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	// TableService does not have a CreateTable method. Call the driver directly.
	var p db.CreateTableParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create table params: %w", err)
	}
	result, err := b.svc.Driver().CreateTable(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTableBackend) UpdateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateTableParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update table params: %w", err)
	}
	result, err := b.svc.Tables.UpdateTable(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcTableBackend) DeleteTable(ctx context.Context, id string) error {
	return b.svc.Tables.DeleteTable(ctx, b.ac, id)
}

// ---------------------------------------------------------------------------
// PluginBackend
// ---------------------------------------------------------------------------

type svcPluginBackend struct {
	svc *service.Registry
}

func (b *svcPluginBackend) ListPlugins(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Plugins.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcPluginBackend) GetPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.svc.Plugins.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcPluginBackend) ReloadPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	if err := b.svc.Plugins.Reload(ctx, name); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "reloaded", "name": name})
}

func (b *svcPluginBackend) EnablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	if err := b.svc.Plugins.Enable(ctx, name); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "enabled", "name": name})
}

func (b *svcPluginBackend) DisablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	if err := b.svc.Plugins.Disable(ctx, name); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{"status": "disabled", "name": name})
}

func (b *svcPluginBackend) PluginCleanupDryRun(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Plugins.CleanupDryRun(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"orphaned_tables": result})
}

func (b *svcPluginBackend) PluginCleanupDrop(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Tables []string `json:"tables"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, fmt.Errorf("unmarshal plugin cleanup drop params: %w", err)
	}
	result, err := b.svc.Plugins.CleanupDrop(ctx, input.Tables)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"dropped": result})
}

func (b *svcPluginBackend) ListPluginRoutes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Plugins.ListRoutes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcPluginBackend) ApprovePluginRoutes(ctx context.Context, params json.RawMessage) error {
	var input struct {
		Routes     []service.RouteApprovalInput `json:"routes"`
		ApprovedBy string                       `json:"approved_by"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return fmt.Errorf("unmarshal approve plugin routes params: %w", err)
	}
	return b.svc.Plugins.ApproveRoutes(ctx, input.Routes, input.ApprovedBy)
}

func (b *svcPluginBackend) RevokePluginRoutes(ctx context.Context, params json.RawMessage) error {
	var input struct {
		Routes []service.RouteApprovalInput `json:"routes"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return fmt.Errorf("unmarshal revoke plugin routes params: %w", err)
	}
	return b.svc.Plugins.RevokeRoutes(ctx, input.Routes)
}

func (b *svcPluginBackend) ListPluginHooks(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Plugins.ListHooks(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcPluginBackend) ApprovePluginHooks(ctx context.Context, params json.RawMessage) error {
	var input struct {
		Hooks      []service.HookApprovalInput `json:"hooks"`
		ApprovedBy string                      `json:"approved_by"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return fmt.Errorf("unmarshal approve plugin hooks params: %w", err)
	}
	return b.svc.Plugins.ApproveHooks(ctx, input.Hooks, input.ApprovedBy)
}

func (b *svcPluginBackend) RevokePluginHooks(ctx context.Context, params json.RawMessage) error {
	var input struct {
		Hooks []service.HookApprovalInput `json:"hooks"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return fmt.Errorf("unmarshal revoke plugin hooks params: %w", err)
	}
	return b.svc.Plugins.RevokeHooks(ctx, input.Hooks)
}

// ---------------------------------------------------------------------------
// ConfigBackend
// ---------------------------------------------------------------------------

type svcConfigBackend struct {
	svc *service.Registry
}

func (b *svcConfigBackend) GetConfig(ctx context.Context, category string) (json.RawMessage, error) {
	if category != "" {
		result, err := b.svc.ConfigSvc.GetConfigByCategory(category)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)
	}
	data, err := b.svc.ConfigSvc.GetConfig()
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func (b *svcConfigBackend) GetConfigMeta(ctx context.Context) (json.RawMessage, error) {
	fields, categories := b.svc.ConfigSvc.GetFieldMetadata()
	return json.Marshal(map[string]any{
		"fields":     fields,
		"categories": categories,
	})
}

func (b *svcConfigBackend) UpdateConfig(ctx context.Context, updates map[string]any) (json.RawMessage, error) {
	result, err := b.svc.ConfigSvc.UpdateConfig(updates)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// ImportBackend
// ---------------------------------------------------------------------------

type svcImportBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcImportBackend) ImportContent(ctx context.Context, format string, data any) (json.RawMessage, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal import data: %w", err)
	}
	result, err := b.svc.Import.ImportContent(ctx, b.ac, service.ImportContentInput{
		Format: config.OutputFormat(format),
		Body:   body,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcImportBackend) ImportBulk(ctx context.Context, format string, data any) (json.RawMessage, error) {
	// ImportBulk is not yet implemented in the ImportService.
	// Delegate to ImportContent as a reasonable fallback.
	return b.ImportContent(ctx, format, data)
}

// ---------------------------------------------------------------------------
// DeployBackend
// ---------------------------------------------------------------------------

type svcDeployBackend struct {
	svc *service.Registry
}

func (b *svcDeployBackend) DeployHealth(ctx context.Context) (json.RawMessage, error) {
	// DeployService is a placeholder with no methods. Return a status indicator.
	return json.Marshal(map[string]string{"status": "not implemented in direct mode"})
}

func (b *svcDeployBackend) DeployExport(ctx context.Context, tables []string) (json.RawMessage, error) {
	return nil, fmt.Errorf("deploy export is not supported in direct mode")
}

func (b *svcDeployBackend) DeployImport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("deploy import is not supported in direct mode")
}

func (b *svcDeployBackend) DeployDryRun(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	return nil, fmt.Errorf("deploy dry run is not supported in direct mode")
}

// ---------------------------------------------------------------------------
// HealthBackend
// ---------------------------------------------------------------------------

type svcHealthBackend struct {
	svc *service.Registry
}

func (b *svcHealthBackend) Health(ctx context.Context) (json.RawMessage, error) {
	resp := map[string]any{
		"status": "ok",
		"checks": map[string]bool{},
	}

	checks := resp["checks"].(map[string]bool)

	// Database check via driver Ping.
	driver := b.svc.Driver()
	if driver != nil {
		if err := driver.Ping(); err != nil {
			checks["database"] = false
			resp["status"] = "degraded"
		} else {
			checks["database"] = true
		}
	} else {
		checks["database"] = false
		resp["status"] = "degraded"
	}

	// Plugin health check if available.
	if b.svc.Plugins != nil {
		pluginHealth, err := b.svc.Plugins.Health(ctx)
		if err != nil {
			checks["plugins"] = false
		} else {
			checks["plugins"] = pluginHealth.Healthy
			if !pluginHealth.Healthy {
				resp["status"] = "degraded"
			}
		}
	}

	return json.Marshal(resp)
}

func (b *svcHealthBackend) GetMetrics(ctx context.Context) (json.RawMessage, error) {
	snapshot := utility.GlobalMetrics.GetSnapshot()
	return json.Marshal(snapshot)
}

func (b *svcHealthBackend) GetEnvironment(ctx context.Context) (json.RawMessage, error) {
	cfg, err := b.svc.Config()
	if err != nil {
		return nil, fmt.Errorf("configuration unavailable: %w", err)
	}
	return json.Marshal(map[string]string{
		"environment": string(cfg.Environment),
		"stage":       cfg.Environment.Stage(),
	})
}
