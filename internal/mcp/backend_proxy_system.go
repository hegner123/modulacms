package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// Tables
// ---------------------------------------------------------------------------

type proxyTableBackend struct{ p *proxyBackends }

func (b *proxyTableBackend) ListTables(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tables.ListTables(ctx)
}
func (b *proxyTableBackend) GetTable(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tables.GetTable(ctx, id)
}
func (b *proxyTableBackend) CreateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tables.CreateTable(ctx, params)
}
func (b *proxyTableBackend) UpdateTable(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Tables.UpdateTable(ctx, params)
}
func (b *proxyTableBackend) DeleteTable(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Tables.DeleteTable(ctx, id)
}

// ---------------------------------------------------------------------------
// Plugins
// ---------------------------------------------------------------------------

type proxyPluginBackend struct{ p *proxyBackends }

func (b *proxyPluginBackend) ListPlugins(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.ListPlugins(ctx)
}
func (b *proxyPluginBackend) GetPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.GetPlugin(ctx, name)
}
func (b *proxyPluginBackend) ReloadPlugin(ctx context.Context, name string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.ReloadPlugin(ctx, name)
}
func (b *proxyPluginBackend) EnablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.EnablePlugin(ctx, name)
}
func (b *proxyPluginBackend) DisablePlugin(ctx context.Context, name string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.DisablePlugin(ctx, name)
}
func (b *proxyPluginBackend) PluginCleanupDryRun(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.PluginCleanupDryRun(ctx)
}
func (b *proxyPluginBackend) PluginCleanupDrop(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.PluginCleanupDrop(ctx, params)
}
func (b *proxyPluginBackend) ListPluginRoutes(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.ListPluginRoutes(ctx)
}
func (b *proxyPluginBackend) ApprovePluginRoutes(ctx context.Context, params json.RawMessage) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Plugins.ApprovePluginRoutes(ctx, params)
}
func (b *proxyPluginBackend) RevokePluginRoutes(ctx context.Context, params json.RawMessage) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Plugins.RevokePluginRoutes(ctx, params)
}
func (b *proxyPluginBackend) ListPluginHooks(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Plugins.ListPluginHooks(ctx)
}
func (b *proxyPluginBackend) ApprovePluginHooks(ctx context.Context, params json.RawMessage) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Plugins.ApprovePluginHooks(ctx, params)
}
func (b *proxyPluginBackend) RevokePluginHooks(ctx context.Context, params json.RawMessage) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Plugins.RevokePluginHooks(ctx, params)
}

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

type proxyConfigBackend struct{ p *proxyBackends }

func (b *proxyConfigBackend) GetConfig(ctx context.Context, category string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Config.GetConfig(ctx, category)
}
func (b *proxyConfigBackend) GetConfigMeta(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Config.GetConfigMeta(ctx)
}
func (b *proxyConfigBackend) UpdateConfig(ctx context.Context, updates map[string]any) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Config.UpdateConfig(ctx, updates)
}

// ---------------------------------------------------------------------------
// Import
// ---------------------------------------------------------------------------

type proxyImportBackend struct{ p *proxyBackends }

func (b *proxyImportBackend) ImportContent(ctx context.Context, format string, data any) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Import.ImportContent(ctx, format, data)
}
func (b *proxyImportBackend) ImportBulk(ctx context.Context, format string, data any) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Import.ImportBulk(ctx, format, data)
}

// ---------------------------------------------------------------------------
// Deploy
// ---------------------------------------------------------------------------

type proxyDeployBackend struct{ p *proxyBackends }

func (b *proxyDeployBackend) DeployHealth(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Deploy.DeployHealth(ctx)
}
func (b *proxyDeployBackend) DeployExport(ctx context.Context, tables []string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Deploy.DeployExport(ctx, tables)
}
func (b *proxyDeployBackend) DeployImport(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Deploy.DeployImport(ctx, payload)
}
func (b *proxyDeployBackend) DeployDryRun(ctx context.Context, payload json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Deploy.DeployDryRun(ctx, payload)
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

type proxyHealthBackend struct{ p *proxyBackends }

func (b *proxyHealthBackend) Health(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Health.Health(ctx)
}
func (b *proxyHealthBackend) GetMetrics(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Health.GetMetrics(ctx)
}
func (b *proxyHealthBackend) GetEnvironment(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Health.GetEnvironment(ctx)
}
