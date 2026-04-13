package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// TableBackend
// ---------------------------------------------------------------------------

type sdkTableBackend struct {
	client *modula.Client
}

func (b *sdkTableBackend) ListTables(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Tables.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTableBackend) GetTable(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Tables.Get(ctx, modula.TableID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTableBackend) CreateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateTableParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create table params: %w", err)
	}
	result, err := b.client.Tables.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTableBackend) UpdateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateTableParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update table params: %w", err)
	}
	result, err := b.client.Tables.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkTableBackend) DeleteTable(ctx context.Context, id string) error {
	return b.client.Tables.Delete(ctx, modula.TableID(id))
}

// ---------------------------------------------------------------------------
// PluginBackend
// ---------------------------------------------------------------------------

type sdkPluginBackend struct {
	client *modula.Client
}

func (b *sdkPluginBackend) ListPlugins(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Plugins.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) GetPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.client.Plugins.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) ReloadPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.client.Plugins.Reload(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) EnablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.client.Plugins.Enable(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) DisablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	result, err := b.client.Plugins.Disable(ctx, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) PluginCleanupDryRun(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Plugins.CleanupDryRun(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) PluginCleanupDrop(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CleanupDropParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal cleanup drop params: %w", err)
	}
	result, err := b.client.Plugins.CleanupDrop(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) ListPluginRoutes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.PluginRoutes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) ApprovePluginRoutes(ctx context.Context, params json.RawMessage) error {
	var items []modula.RouteApprovalItem
	if err := json.Unmarshal(params, &items); err != nil {
		return fmt.Errorf("unmarshal route approval items: %w", err)
	}
	return b.client.PluginRoutes.Approve(ctx, items)
}

func (b *sdkPluginBackend) RevokePluginRoutes(ctx context.Context, params json.RawMessage) error {
	var items []modula.RouteApprovalItem
	if err := json.Unmarshal(params, &items); err != nil {
		return fmt.Errorf("unmarshal route revocation items: %w", err)
	}
	return b.client.PluginRoutes.Revoke(ctx, items)
}

func (b *sdkPluginBackend) ListPluginHooks(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.PluginHooks.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkPluginBackend) ApprovePluginHooks(ctx context.Context, params json.RawMessage) error {
	var items []modula.HookApprovalItem
	if err := json.Unmarshal(params, &items); err != nil {
		return fmt.Errorf("unmarshal hook approval items: %w", err)
	}
	return b.client.PluginHooks.Approve(ctx, items)
}

func (b *sdkPluginBackend) RevokePluginHooks(ctx context.Context, params json.RawMessage) error {
	var items []modula.HookApprovalItem
	if err := json.Unmarshal(params, &items); err != nil {
		return fmt.Errorf("unmarshal hook revocation items: %w", err)
	}
	return b.client.PluginHooks.Revoke(ctx, items)
}

// ---------------------------------------------------------------------------
// ConfigBackend
// ---------------------------------------------------------------------------

type sdkConfigBackend struct {
	client *modula.Client
}

func (b *sdkConfigBackend) GetConfig(ctx context.Context, category string) (json.RawMessage, error) {
	result, err := b.client.Config.Get(ctx, category)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkConfigBackend) GetConfigMeta(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Config.Meta(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkConfigBackend) UpdateConfig(ctx context.Context, updates map[string]any) (json.RawMessage, error) {
	result, err := b.client.Config.Update(ctx, updates)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// ImportBackend
// ---------------------------------------------------------------------------

type sdkImportBackend struct {
	client *modula.Client
}

func (b *sdkImportBackend) ImportContent(ctx context.Context, format string, data any) (json.RawMessage, error) {
	switch format {
	case "contentful":
		return b.client.Import.Contentful(ctx, data)
	case "sanity":
		return b.client.Import.Sanity(ctx, data)
	case "strapi":
		return b.client.Import.Strapi(ctx, data)
	case "wordpress":
		return b.client.Import.WordPress(ctx, data)
	case "clean":
		return b.client.Import.Clean(ctx, data)
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}
}

func (b *sdkImportBackend) ImportBulk(ctx context.Context, format string, data any) (json.RawMessage, error) {
	return b.client.Import.Bulk(ctx, format, data)
}

// ---------------------------------------------------------------------------
// DeployBackend
// ---------------------------------------------------------------------------

type sdkDeployBackend struct {
	client *modula.Client
}

func (b *sdkDeployBackend) DeployHealth(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Deploy.Health(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkDeployBackend) DeployExport(ctx context.Context, tables []string) (json.RawMessage, error) {
	return b.client.Deploy.Export(ctx, tables)
}

func (b *sdkDeployBackend) DeployImport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	result, err := b.client.Deploy.Import(ctx, payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkDeployBackend) DeployDryRun(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	result, err := b.client.Deploy.DryRunImport(ctx, payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ---------------------------------------------------------------------------
// HealthBackend
// ---------------------------------------------------------------------------

type sdkHealthBackend struct {
	client *modula.Client
}

func (b *sdkHealthBackend) Health(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Health.Check(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkHealthBackend) GetMetrics(ctx context.Context) (json.RawMessage, error) {
	return b.client.Metrics.Get(ctx)
}

func (b *sdkHealthBackend) GetEnvironment(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Environment.Get(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
