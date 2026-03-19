package service

import (
	"context"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/plugin"
)

// PluginManager is a consumer-defined interface accepting only the methods
// PluginService actually calls. The plugin.Manager satisfies it implicitly.
//
// Bridge(), HookEngine(), and RequestEngine() return concrete pointer types
// from the plugin package. This is a pragmatic tradeoff -- see the Phase 5 plan
// for rationale.
type PluginManager interface {
	// Lifecycle
	InstallPlugin(ctx context.Context, name string) (*db.Plugin, error)
	EnablePlugin(ctx context.Context, name string) error
	DisablePlugin(ctx context.Context, name string) error
	ReloadPlugin(ctx context.Context, name string) error
	ActivatePlugin(ctx context.Context, name string, adminUser string) error
	DeactivatePlugin(ctx context.Context, name string) error

	// Queries
	GetPlugin(name string) *plugin.PluginInstance
	ListPlugins() []*plugin.PluginInstance
	GetPluginState(name string) (plugin.PluginState, bool)
	PluginHealth() plugin.PluginHealthStatus

	// Discovery & Registry
	ListDiscovered() []string
	SyncCapabilities(ctx context.Context, name string, adminUser string) error
	ListAllPipelines() (*[]db.Pipeline, error)
	DryRunPipeline(table, op string) []plugin.DryRunResult
	DryRunAllPipelines() []plugin.DryRunResult

	// Cleanup
	ListOrphanedTables(ctx context.Context) ([]string, error)
	DropOrphanedTables(ctx context.Context, requestedTables []string) ([]string, error)

	// Sub-components
	Bridge() *plugin.HTTPBridge
	HookEngine() *plugin.HookEngine
	RequestEngine() *plugin.RequestEngine
}

// PluginService wraps plugin.Manager as a thin facade, providing service-layer
// error mapping and a unified interface for all consumers (admin panel, API, MCP).
type PluginService struct {
	mgr PluginManager
}

// NewPluginService creates a PluginService. mgr may be nil if plugins are not enabled;
// read methods return empty results, mutations return a ValidationError.
func NewPluginService(mgr PluginManager) *PluginService {
	return &PluginService{mgr: mgr}
}

// --- Service-Level Types ---

// PluginSummary is a lightweight view of a plugin for list rendering.
type PluginSummary struct {
	Name                string
	Version             string
	Description         string
	State               string
	CircuitBreakerState string
}

// PluginDetail is a full view of a plugin for detail rendering.
type PluginDetail struct {
	PluginSummary
	Author       string
	License      string
	FailedReason string
	VMsAvailable int
	VMsTotal     int
	Dependencies []string
	SchemaDrift  bool
}

// PluginRoute is the service-level view of a registered plugin route.
type PluginRoute struct {
	Method     string
	Path       string
	PluginName string
	FullPath   string
	Public     bool
	Approved   bool
}

// RouteApprovalInput identifies a plugin route for approval or revocation.
type RouteApprovalInput struct {
	Plugin string
	Method string
	Path   string
}

// HookApprovalInput identifies a plugin hook for approval or revocation.
type HookApprovalInput struct {
	PluginName string
	Event      string
	Table      string
}

// HookInfo describes a registered hook with approval status.
type HookInfo struct {
	PluginName string
	Event      string
	Table      string
	Priority   int
	Approved   bool
	IsWildcard bool
}

// PipelineChain is the service-level view of a dry-run pipeline chain.
type PipelineChain struct {
	Key       string // "table.phase_operation"
	Table     string
	Operation string // base operation: "create", "update", "delete"
	Phase     string // "before" or "after"
	Entries   []PipelineEntryInfo
}

// PipelineEntryInfo describes a single entry in a pipeline chain.
type PipelineEntryInfo struct {
	PipelineID string
	PluginName string
	Handler    string
	Priority   int
	Enabled    bool
}

// RequestApprovalInput identifies a plugin request domain for approval or revocation.
type RequestApprovalInput struct {
	PluginName string
	Domain     string
}

// RequestInfo describes a registered request domain with approval status.
type RequestInfo struct {
	PluginName string
	Domain     string
	Approved   bool
}

// --- Error Helpers ---

// pluginsDisabledError returns a ValidationError indicating the plugin system is not enabled.
func pluginsDisabledError() *ValidationError {
	return NewValidationError("plugins", "plugin system is not enabled")
}

// mapPluginError converts a plugin manager error to a service-layer error.
func mapPluginError(err error, name string) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "not found") || strings.Contains(lower, "not installed") {
		return &NotFoundError{Resource: "plugin", ID: name}
	}
	if strings.Contains(lower, "already enabled") || strings.Contains(lower, "already disabled") {
		return &ConflictError{Resource: "plugin", Detail: msg}
	}
	return &InternalError{Err: err}
}

// --- Methods ---

// List returns summaries of all loaded plugins. If the plugin system is not
// enabled (nil manager), returns an empty slice and no error.
func (s *PluginService) List(ctx context.Context) ([]PluginSummary, error) {
	if s.mgr == nil {
		return []PluginSummary{}, nil
	}

	instances := s.mgr.ListPlugins()
	summaries := make([]PluginSummary, 0, len(instances))
	for _, inst := range instances {
		cbState := ""
		if inst.CB != nil {
			cbState = inst.CB.State().String()
		}
		summaries = append(summaries, PluginSummary{
			Name:                inst.Info.Name,
			Version:             inst.Info.Version,
			Description:         inst.Info.Description,
			State:               inst.State.String(),
			CircuitBreakerState: cbState,
		})
	}
	return summaries, nil
}

// Get returns the full detail of a single plugin by name.
func (s *PluginService) Get(ctx context.Context, name string) (*PluginDetail, error) {
	if s.mgr == nil {
		return nil, &NotFoundError{Resource: "plugin", ID: name}
	}

	inst := s.mgr.GetPlugin(name)
	if inst == nil {
		return nil, &NotFoundError{Resource: "plugin", ID: name}
	}

	cbState := ""
	if inst.CB != nil {
		cbState = inst.CB.State().String()
	}

	vmsAvailable := 0
	vmsTotal := 0
	if inst.Pool != nil {
		vmsAvailable = inst.Pool.AvailableCount()
		vmsTotal = inst.Pool.PoolSize()
	}

	deps := inst.Info.Dependencies
	if deps == nil {
		deps = []string{}
	}

	hasDrift := len(inst.SchemaDrift) > 0

	detail := &PluginDetail{
		PluginSummary: PluginSummary{
			Name:                inst.Info.Name,
			Version:             inst.Info.Version,
			Description:         inst.Info.Description,
			State:               inst.State.String(),
			CircuitBreakerState: cbState,
		},
		Author:       inst.Info.Author,
		License:      inst.Info.License,
		FailedReason: inst.FailedReason,
		VMsAvailable: vmsAvailable,
		VMsTotal:     vmsTotal,
		Dependencies: deps,
		SchemaDrift:  hasDrift,
	}
	return detail, nil
}

// Enable enables a plugin by name.
func (s *PluginService) Enable(ctx context.Context, name string) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	if name == "" {
		return NewValidationError("name", "plugin name is required")
	}
	return mapPluginError(s.mgr.EnablePlugin(ctx, name), name)
}

// Disable disables a plugin by name.
func (s *PluginService) Disable(ctx context.Context, name string) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	if name == "" {
		return NewValidationError("name", "plugin name is required")
	}
	return mapPluginError(s.mgr.DisablePlugin(ctx, name), name)
}

// Reload reloads a plugin by name.
func (s *PluginService) Reload(ctx context.Context, name string) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	if name == "" {
		return NewValidationError("name", "plugin name is required")
	}
	return mapPluginError(s.mgr.ReloadPlugin(ctx, name), name)
}

// Install installs a discovered plugin by name.
func (s *PluginService) Install(ctx context.Context, name string) (*db.Plugin, error) {
	if s.mgr == nil {
		return nil, pluginsDisabledError()
	}
	p, err := s.mgr.InstallPlugin(ctx, name)
	if err != nil {
		return nil, mapPluginError(err, name)
	}
	return p, nil
}

// SyncCapabilities synchronizes a plugin's capabilities with the database.
func (s *PluginService) SyncCapabilities(ctx context.Context, name, adminUser string) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	return mapPluginError(s.mgr.SyncCapabilities(ctx, name, adminUser), name)
}

// Health returns the aggregate health status of the plugin system.
func (s *PluginService) Health(ctx context.Context) (*plugin.PluginHealthStatus, error) {
	if s.mgr == nil {
		return &plugin.PluginHealthStatus{Healthy: true}, nil
	}
	h := s.mgr.PluginHealth()
	return &h, nil
}

// CleanupDryRun returns a list of orphaned plugin tables without dropping them.
func (s *PluginService) CleanupDryRun(ctx context.Context) ([]string, error) {
	if s.mgr == nil {
		return []string{}, nil
	}
	tables, err := s.mgr.ListOrphanedTables(ctx)
	if err != nil {
		return nil, &InternalError{Err: err}
	}
	if tables == nil {
		return []string{}, nil
	}
	return tables, nil
}

// CleanupDrop drops the specified orphaned plugin tables.
func (s *PluginService) CleanupDrop(ctx context.Context, tables []string) ([]string, error) {
	if s.mgr == nil {
		return nil, pluginsDisabledError()
	}
	if len(tables) == 0 {
		return nil, NewValidationError("tables", "at least one table name is required")
	}
	dropped, err := s.mgr.DropOrphanedTables(ctx, tables)
	if err != nil {
		return nil, &InternalError{Err: err}
	}
	if dropped == nil {
		return []string{}, nil
	}
	return dropped, nil
}

// ListRoutes returns all registered plugin routes with approval status.
func (s *PluginService) ListRoutes(ctx context.Context) ([]PluginRoute, error) {
	if s.mgr == nil {
		return []PluginRoute{}, nil
	}
	bridge := s.mgr.Bridge()
	if bridge == nil {
		return []PluginRoute{}, nil
	}
	regs := bridge.ListRoutes()
	routes := make([]PluginRoute, 0, len(regs))
	for _, r := range regs {
		routes = append(routes, PluginRoute{
			Method:     r.Method,
			Path:       r.Path,
			PluginName: r.PluginName,
			FullPath:   r.FullPath,
			Public:     r.Public,
			Approved:   r.Approved,
		})
	}
	return routes, nil
}

// ApproveRoutes approves one or more plugin routes.
func (s *PluginService) ApproveRoutes(ctx context.Context, routes []RouteApprovalInput, approvedBy string) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	bridge := s.mgr.Bridge()
	if bridge == nil {
		return pluginsDisabledError()
	}
	for _, r := range routes {
		if err := bridge.ApproveRoute(ctx, r.Plugin, r.Method, r.Path, approvedBy); err != nil {
			return mapPluginError(err, r.Plugin)
		}
	}
	return nil
}

// RevokeRoutes revokes approval for one or more plugin routes.
func (s *PluginService) RevokeRoutes(ctx context.Context, routes []RouteApprovalInput) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	bridge := s.mgr.Bridge()
	if bridge == nil {
		return pluginsDisabledError()
	}
	for _, r := range routes {
		if err := bridge.RevokeRoute(ctx, r.Plugin, r.Method, r.Path); err != nil {
			return mapPluginError(err, r.Plugin)
		}
	}
	return nil
}

// ListHooks returns all registered plugin hooks with approval status.
func (s *PluginService) ListHooks(ctx context.Context) ([]HookInfo, error) {
	if s.mgr == nil {
		return []HookInfo{}, nil
	}
	engine := s.mgr.HookEngine()
	if engine == nil {
		return []HookInfo{}, nil
	}
	regs := engine.ListHooks()
	hooks := make([]HookInfo, 0, len(regs))
	for _, r := range regs {
		hooks = append(hooks, HookInfo{
			PluginName: r.PluginName,
			Event:      r.Event,
			Table:      r.Table,
			Priority:   r.Priority,
			Approved:   r.Approved,
			IsWildcard: r.IsWildcard,
		})
	}
	return hooks, nil
}

// ApproveHooks approves one or more plugin hooks.
func (s *PluginService) ApproveHooks(ctx context.Context, hooks []HookApprovalInput, approvedBy string) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	engine := s.mgr.HookEngine()
	if engine == nil {
		return pluginsDisabledError()
	}
	for _, h := range hooks {
		if err := engine.ApproveHook(ctx, h.PluginName, h.Event, h.Table, approvedBy); err != nil {
			return mapPluginError(err, h.PluginName)
		}
	}
	return nil
}

// RevokeHooks revokes approval for one or more plugin hooks.
func (s *PluginService) RevokeHooks(ctx context.Context, hooks []HookApprovalInput) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	engine := s.mgr.HookEngine()
	if engine == nil {
		return pluginsDisabledError()
	}
	for _, h := range hooks {
		if err := engine.RevokeHook(ctx, h.PluginName, h.Event, h.Table); err != nil {
			return mapPluginError(err, h.PluginName)
		}
	}
	return nil
}

// ListRequests returns all registered plugin request domains with approval status.
func (s *PluginService) ListRequests(ctx context.Context) ([]RequestInfo, error) {
	if s.mgr == nil {
		return []RequestInfo{}, nil
	}
	engine := s.mgr.RequestEngine()
	if engine == nil {
		return []RequestInfo{}, nil
	}
	regs, err := engine.ListRequests(ctx)
	if err != nil {
		return nil, &InternalError{Err: err}
	}
	infos := make([]RequestInfo, 0, len(regs))
	for _, r := range regs {
		infos = append(infos, RequestInfo{
			PluginName: r.PluginName,
			Domain:     r.Domain,
			Approved:   r.Approved,
		})
	}
	return infos, nil
}

// ApproveRequests approves one or more plugin request domains.
func (s *PluginService) ApproveRequests(ctx context.Context, requests []RequestApprovalInput, approvedBy string) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	engine := s.mgr.RequestEngine()
	if engine == nil {
		return pluginsDisabledError()
	}
	for _, r := range requests {
		if err := engine.ApproveRequest(ctx, r.PluginName, r.Domain, approvedBy); err != nil {
			return mapPluginError(err, r.PluginName)
		}
	}
	return nil
}

// RevokeRequests revokes approval for one or more plugin request domains.
func (s *PluginService) RevokeRequests(ctx context.Context, requests []RequestApprovalInput) error {
	if s.mgr == nil {
		return pluginsDisabledError()
	}
	engine := s.mgr.RequestEngine()
	if engine == nil {
		return pluginsDisabledError()
	}
	for _, r := range requests {
		if err := engine.RevokeRequest(ctx, r.PluginName, r.Domain); err != nil {
			return mapPluginError(err, r.PluginName)
		}
	}
	return nil
}

// ListPipelines returns all registered pipelines.
func (s *PluginService) ListPipelines(ctx context.Context) ([]db.Pipeline, error) {
	if s.mgr == nil {
		return []db.Pipeline{}, nil
	}
	pipelines, err := s.mgr.ListAllPipelines()
	if err != nil {
		return nil, &InternalError{Err: err}
	}
	if pipelines == nil {
		return []db.Pipeline{}, nil
	}
	return *pipelines, nil
}

// DryRunPipelines returns dry-run results for all registered pipelines.
func (s *PluginService) DryRunPipelines(ctx context.Context) ([]PipelineChain, error) {
	if s.mgr == nil {
		return []PipelineChain{}, nil
	}
	results := s.mgr.DryRunAllPipelines()
	if results == nil {
		return []PipelineChain{}, nil
	}
	chains := make([]PipelineChain, 0, len(results))
	for _, r := range results {
		entries := make([]PipelineEntryInfo, 0, len(r.Entries))
		for _, e := range r.Entries {
			entries = append(entries, PipelineEntryInfo{
				PipelineID: e.PipelineID,
				PluginName: e.PluginName,
				Handler:    e.Handler,
				Priority:   e.Priority,
				Enabled:    e.Enabled,
			})
		}
		chains = append(chains, PipelineChain{
			Key:       r.Table + "." + r.Phase + "_" + r.Operation,
			Table:     r.Table,
			Operation: r.Operation,
			Phase:     r.Phase,
			Entries:   entries,
		})
	}
	return chains, nil
}
