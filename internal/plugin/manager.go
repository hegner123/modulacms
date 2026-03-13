package plugin

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"encoding/json"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
	lua "github.com/yuin/gopher-lua"
)

// PluginState represents the lifecycle state of a plugin.
// Uses iota enum (not raw string) per roadmap review round 4.
type PluginState int

const (
	StateDiscovered PluginState = iota
	StateLoading
	StateRunning
	StateFailed
	StateStopped
)

// String returns the human-readable name of the plugin state.
func (s PluginState) String() string {
	switch s {
	case StateDiscovered:
		return "discovered"
	case StateLoading:
		return "loading"
	case StateRunning:
		return "running"
	case StateFailed:
		return "failed"
	case StateStopped:
		return "stopped"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// PluginCapability represents a single declared pipeline capability from the manifest.
type PluginCapability struct {
	Table    string `json:"table"`
	Op       string `json:"op"`
	Handler  string `json:"handler"`
	Priority int    `json:"priority"`
}

// CapabilityDriftEntry represents a single difference between filesystem and DB capabilities.
type CapabilityDriftEntry struct {
	Kind     string            `json:"kind"`               // "added", "removed", "changed"
	Current  *PluginCapability `json:"current,omitempty"`  // from filesystem (nil if removed)
	Previous *PluginCapability `json:"previous,omitempty"` // from DB (nil if added)
}

// PluginCoreAccess maps table names to allowed operations.
type PluginCoreAccess map[string][]string

// PluginInfo holds the metadata extracted from a plugin's plugin_info global.
type PluginInfo struct {
	Name          string
	Version       string
	Description   string
	Author        string
	License       string
	MinCMSVersion string
	Dependencies  []string

	// Capabilities lists the pipeline hooks this plugin declares.
	// Parsed from plugin_info.capabilities during ExtractManifest.
	Capabilities []PluginCapability

	// CoreAccess maps core CMS table names to allowed operations.
	// Parsed from plugin_info.core_access during ExtractManifest.
	CoreAccess PluginCoreAccess

	// HasOnInit indicates whether the plugin defines an on_init() function.
	// Set during ExtractManifest() by checking the global before the VM is closed.
	HasOnInit bool
}

// PluginInstance represents a loaded plugin with its state and resources.
type PluginInstance struct {
	Info         PluginInfo
	State        PluginState
	Dir          string // absolute path to plugin directory
	InitPath     string // absolute path to init.lua
	FailedReason string // human-readable failure message; empty when State != StateFailed
	Pool         *VMPool

	// CB is the plugin-level circuit breaker (Phase 4). Tracks consecutive
	// failures from HTTP handler execution and manager operations (reload, init).
	// Hook failures are tracked by the separate hook-level CB in hook_engine.go.
	CB *CircuitBreaker

	// SchemaDrift holds advisory drift entries detected during define_table.
	// Surfaced via PluginInfoHandler admin response (S7).
	SchemaDrift []DriftEntry

	// ManifestDrift is true when the filesystem Lua hash differs from the
	// stored manifest_hash at load time. Surfaced via plugin info/list.
	ManifestDrift bool

	// CapabilityDrift lists the differences between filesystem capabilities
	// and DB-stored capabilities detected during LoadAll. Empty means no drift.
	CapabilityDrift []CapabilityDriftEntry

	// ApprovedAccess is the set of core table operations approved for this plugin.
	// Loaded from the plugins.approved_access DB column during LoadAll().
	// Nil means no core table access (discovered-only or legacy plugins).
	ApprovedAccess PluginCoreAccess

	// mu protects dbAPIs and requestAPIs from concurrent access. Required because the
	// onReplace callback (called from Put on a VM-returning goroutine) deletes entries
	// while executeBefore/executeAfter read entries on different goroutines (S5).
	mu sync.Mutex

	// dbAPIs maps each VM (*lua.LState) to its bound DatabaseAPI for op count reset.
	// Each DatabaseAPI is bound to exactly one LState (1:1 invariant).
	dbAPIs map[*lua.LState]*DatabaseAPI

	// requestAPIs maps each VM (*lua.LState) to its bound requestAPIState for
	// before-hook guard wiring. Each requestAPIState is bound to exactly one
	// LState (1:1 invariant).
	requestAPIs map[*lua.LState]*requestAPIState
}

// ManagerConfig configures the plugin manager runtime behavior.
type ManagerConfig struct {
	Enabled         bool
	Directory       string
	NodeID          string // config node_id for audited operations
	MaxVMsPerPlugin int    // default 4
	ExecTimeoutSec  int    // default 5
	MaxOpsPerExec   int    // default 1000, per VM checkout

	// Hook engine configuration (Phase 3).
	HookReserveVMs           int // VMs reserved for hook execution; default 1
	HookMaxConsecutiveAborts int // circuit breaker threshold; default 10
	HookMaxOps               int // reduced op budget for after-hooks; default 100
	HookMaxConcurrentAfter   int // max concurrent after-hook goroutines; default 10
	HookTimeoutMs            int // per-hook timeout in before-hooks (ms); default 2000
	HookEventTimeoutMs       int // per-event total timeout for before-hook chain (ms); default 5000

	// Phase 4: Production hardening configuration.
	HotReload     bool          // default false (zero value) -- production opt-in only (S10)
	MaxFailures   int           // circuit breaker threshold; default 5
	ResetInterval time.Duration // circuit breaker reset interval; default 60s

	// Outbound request engine configuration.
	RequestTimeoutSec       int
	RequestMaxResponseBytes int64
	RequestMaxRequestBody   int64
	RequestMaxPerMin        int
	RequestGlobalMaxPerMin  int
	RequestCBMaxFailures    int
	RequestCBResetInterval  int
	RequestAllowLocalhost   bool
}

// TableRegistrar registers table names in the CMS tables registry.
// Implementations must be idempotent — registering an already-registered
// table is a no-op (no error).
type TableRegistrar interface {
	RegisterTable(ctx context.Context, label string) error
}

// Manager is the central coordinator for plugin discovery, loading, lifecycle, and shutdown.
//
// Context handling: Manager does not store a context.Context field. Stored contexts on
// long-lived structs create ambiguity about which context governs a given operation.
// Instead, all Manager methods that perform I/O accept ctx context.Context as their
// first parameter. The caller (cmd/serve.go) passes the application lifecycle context.
type Manager struct {
	cfg      ManagerConfig
	db       *sql.DB // separate plugin pool via db.OpenPool()
	dialect  db.Dialect
	driver   db.DbDriver    // for reading plugins/pipelines tables; may be nil for tests
	tableReg TableRegistrar // registers plugin tables in the CMS tables registry; may be nil
	plugins  map[string]*PluginInstance
	mu       sync.RWMutex

	// registry is the in-memory pipeline cache loaded from the pipelines table.
	// Always non-nil when the Manager is non-nil.
	registry *PipelineRegistry

	// loadOrder preserves the topologically sorted order of successfully loaded plugins.
	// Used for reverse-order shutdown.
	loadOrder []string

	// bridge is the HTTP bridge for plugin route registration. Set via SetBridge()
	// before LoadAll(). May be nil if HTTP integration is not enabled.
	bridge *HTTPBridge

	// hookEngine is the content lifecycle hook engine (Phase 3). Created during
	// NewManager, always non-nil when plugins are enabled. Implements audited.HookRunner.
	hookEngine *HookEngine

	// requestEngine is the outbound HTTP request engine. Created during
	// NewManager, always non-nil when plugins are enabled.
	requestEngine *RequestEngine

	// watcher is the file-polling hot reload watcher (Phase 4). Nil when hot
	// reload is disabled. Started via StartWatcher() after LoadAll().
	watcher       *Watcher
	watcherCancel context.CancelFunc

	// coordinator is the DB state polling coordinator for multi-instance sync.
	// Nil when disabled (Plugin_Sync_Interval == "0" or driver == nil).
	coordinator       *Coordinator
	coordinatorCancel context.CancelFunc

	// shutdownOnce ensures Shutdown is idempotent -- calling it multiple times
	// (e.g., signal handler + deferred call) does not double-close resources.
	shutdownOnce sync.Once
}

// NewManager creates a new plugin Manager with the given configuration.
// Zero-value config fields are replaced with defaults.
// The db pool must be a separate *sql.DB opened via db.OpenPool() for isolation.
// driver may be nil for tests; when nil, lifecycle methods that require DB access return errors.
// tableReg may be nil; when nil, plugin tables are not registered in the CMS tables registry.
func NewManager(cfg ManagerConfig, pool *sql.DB, dialect db.Dialect, driver db.DbDriver, tableReg TableRegistrar) *Manager {
	if cfg.MaxVMsPerPlugin <= 0 {
		cfg.MaxVMsPerPlugin = 4
	}
	if cfg.ExecTimeoutSec <= 0 {
		cfg.ExecTimeoutSec = 5
	}
	if cfg.MaxOpsPerExec <= 0 {
		cfg.MaxOpsPerExec = 1000
	}
	if cfg.HookReserveVMs <= 0 {
		cfg.HookReserveVMs = 1
	}
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = 5
	}
	if cfg.ResetInterval <= 0 {
		cfg.ResetInterval = 60 * time.Second
	}

	mgr := &Manager{
		cfg:      cfg,
		db:       pool,
		dialect:  dialect,
		driver:   driver,
		tableReg: tableReg,
		plugins:  make(map[string]*PluginInstance),
		registry: NewPipelineRegistry(),
	}

	// Create hook engine unconditionally. The engine is inert when no hooks
	// are registered (HasHooks returns false via hasAnyHook fast-path).
	mgr.hookEngine = NewHookEngine(mgr, pool, dialect, HookEngineConfig{
		HookTimeoutMs:        cfg.HookTimeoutMs,
		EventTimeoutMs:       cfg.HookEventTimeoutMs,
		MaxConsecutiveAborts: cfg.HookMaxConsecutiveAborts,
		MaxConcurrentAfter:   cfg.HookMaxConcurrentAfter,
		HookMaxOps:           cfg.HookMaxOps,
		ExecTimeoutMs:        cfg.ExecTimeoutSec * 1000, // reuse exec timeout for after-hooks
	})

	// Create outbound request engine unconditionally. The engine is inert
	// when no plugins register request domains.
	mgr.requestEngine = NewRequestEngine(pool, dialect, RequestEngineConfig{
		DefaultTimeoutSec:       cfg.RequestTimeoutSec,
		MaxResponseBytes:        cfg.RequestMaxResponseBytes,
		MaxRequestBodyBytes:     cfg.RequestMaxRequestBody,
		MaxRequestsPerMin:       cfg.RequestMaxPerMin,
		GlobalMaxRequestsPerMin: cfg.RequestGlobalMaxPerMin,
		CBMaxFailures:           cfg.RequestCBMaxFailures,
		CBResetIntervalSec:      cfg.RequestCBResetInterval,
		AllowLocalhost:          cfg.RequestAllowLocalhost,
	})

	return mgr
}

// SetBridge stores the HTTP bridge on the manager. Must be called before
// LoadAll() so that plugin loading can register routes on the bridge.
func (m *Manager) SetBridge(bridge *HTTPBridge) {
	m.bridge = bridge
}

// Bridge returns the HTTP bridge, or nil if not set.
func (m *Manager) Bridge() *HTTPBridge {
	return m.bridge
}

// HookEngine returns the content lifecycle hook engine.
// Always non-nil when the Manager is non-nil.
func (m *Manager) HookEngine() *HookEngine {
	return m.hookEngine
}

// RequestEngine returns the outbound request engine.
// Always non-nil when the Manager is non-nil.
func (m *Manager) RequestEngine() *RequestEngine {
	return m.requestEngine
}

// Registry returns the in-memory pipeline registry.
// Always non-nil when the Manager is non-nil.
func (m *Manager) Registry() *PipelineRegistry {
	return m.registry
}

// Driver returns the DbDriver, or nil if not set.
func (m *Manager) Driver() db.DbDriver {
	return m.driver
}

// LoadRegistry queries enabled pipelines from the database and rebuilds the
// in-memory PipelineRegistry. Safe to call multiple times (build-then-swap).
// Returns nil if the driver is not set (test mode).
func (m *Manager) LoadRegistry(ctx context.Context) error {
	if m.driver == nil {
		return nil
	}

	pipelines, err := m.driver.ListEnabledPipelines()
	if err != nil {
		return fmt.Errorf("loading enabled pipelines: %w", err)
	}

	if pipelines == nil {
		m.registry.Build(nil)
		return nil
	}

	rows := make([]PipelineRow, 0, len(*pipelines))
	for _, p := range *pipelines {
		configStr := ""
		if p.Config.Valid {
			if s, ok := p.Config.Data.(string); ok {
				configStr = s
			} else {
				// Config.Data may already be a parsed type; marshal it back.
				if b, mErr := json.Marshal(p.Config.Data); mErr == nil {
					configStr = string(b)
				}
			}
		}

		rows = append(rows, PipelineRow{
			PipelineID: p.PipelineID.String(),
			PluginID:   p.PluginID.String(),
			TableName:  p.TableName,
			Operation:  p.Operation,
			PluginName: p.PluginName,
			Handler:    p.Handler,
			Priority:   p.Priority,
			Enabled:    p.Enabled,
			Config:     configStr,
		})
	}

	m.registry.Build(rows)
	return nil
}

// LoadAll discovers plugins in the configured directory, validates manifests,
// resolves dependencies via topological sort, and loads each plugin in
// dependency order. Failed plugins are marked StateFailed and do not prevent
// other plugins from loading.
//
// Loading sequence per plugin:
//  1. Scan cfg.Directory for subdirectories containing init.lua
//  2. Create temp sandboxed VM (no db/log APIs), execute init.lua, extract/validate plugin_info
//  3. Topologically sort by dependencies (detect cycles)
//  4. For each plugin in dependency order: create VMPool, run on_init, snapshot globals
func (m *Manager) LoadAll(ctx context.Context) error {
	if m.cfg.Directory == "" {
		return fmt.Errorf("plugin directory not configured")
	}

	// Step 1: Scan for plugin directories containing init.lua.
	entries, err := os.ReadDir(m.cfg.Directory)
	if err != nil {
		return fmt.Errorf("reading plugin directory %q: %w", m.cfg.Directory, err)
	}

	discovered := make(map[string]*PluginInstance)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginDir := filepath.Join(m.cfg.Directory, entry.Name())
		initPath := filepath.Join(pluginDir, "init.lua")

		if _, statErr := os.Stat(initPath); statErr != nil {
			// No init.lua -- skip this directory silently.
			continue
		}

		// Step 2: Extract and validate manifest from temp VM.
		info, extractErr := ExtractManifest(initPath)
		if extractErr != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: manifest extraction failed: %s", entry.Name(), extractErr.Error()),
				nil,
			)
			continue
		}

		// Check for duplicate plugin names (first-discovered wins).
		if _, exists := discovered[info.Name]; exists {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: duplicate name %q (already discovered), skipping",
					entry.Name(), info.Name),
				nil,
			)
			continue
		}

		discovered[info.Name] = &PluginInstance{
			Info:        *info,
			State:       StateDiscovered,
			Dir:         pluginDir,
			InitPath:    initPath,
			dbAPIs:      make(map[*lua.LState]*DatabaseAPI),
			requestAPIs: make(map[*lua.LState]*requestAPIState),
		}
	}

	// Step 2.5: Load approved_access from the plugins DB table and detect drift.
	// This enriches discovered PluginInstances with their DB-stored permissions
	// and compares filesystem state against stored DB records.
	if m.driver != nil {
		dbPlugins, err := m.driver.ListPlugins()
		if err != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("failed to query plugins table for approved_access: %s", err.Error()),
				nil,
			)
		} else if dbPlugins != nil {
			dbByName := make(map[string]*db.Plugin, len(*dbPlugins))
			for i := range *dbPlugins {
				dbByName[(*dbPlugins)[i].Name] = &(*dbPlugins)[i]
			}
			for name, inst := range discovered {
				dbRow, ok := dbByName[name]
				if !ok {
					continue
				}

				// Enrich with approved_access.
				if dbRow.ApprovedAccess.Valid {
					var access PluginCoreAccess
					if jsonErr := json.Unmarshal([]byte(dbRow.ApprovedAccess.String()), &access); jsonErr != nil {
						utility.DefaultLogger.Warn(
							fmt.Sprintf("plugin %q: invalid approved_access JSON: %s", name, jsonErr.Error()),
							nil,
						)
					} else {
						inst.ApprovedAccess = access
					}
				}

				// Step 2.6: Manifest hash drift (skip if no baseline stored).
				if dbRow.ManifestHash != "" {
					fsHash, hashErr := computeChecksum(inst.Dir)
					if hashErr != nil {
						utility.DefaultLogger.Warn(
							fmt.Sprintf("plugin %q: checksum error during drift detection: %s", name, hashErr.Error()),
							nil,
						)
					} else if fsHash != dbRow.ManifestHash {
						inst.ManifestDrift = true
					}
				}

				// Capability drift: compare filesystem capabilities against DB capabilities.
				var storedCaps []PluginCapability
				capStr := dbRow.Capabilities.String()
				if capStr != "" && capStr != "null" {
					if unmarshalErr := json.Unmarshal([]byte(capStr), &storedCaps); unmarshalErr != nil {
						utility.DefaultLogger.Warn(
							fmt.Sprintf("plugin %q: invalid capabilities JSON in DB: %s", name, unmarshalErr.Error()),
							nil,
						)
					}
				}
				inst.CapabilityDrift = compareCapabilities(inst.Info.Capabilities, storedCaps)
			}
		}
	}

	// Step 3: Topologically sort by dependencies.
	sorted, sortErr := topologicalSort(discovered)
	if sortErr != nil {
		return fmt.Errorf("dependency resolution: %w", sortErr)
	}

	// Phase 2: Create plugin_routes table and clean up orphaned routes before
	// loading plugins. These calls are safe to skip when no bridge is set
	// (Phase 1 operation without HTTP integration).
	if m.bridge != nil {
		if tblErr := m.bridge.CreatePluginRoutesTable(ctx); tblErr != nil {
			return fmt.Errorf("creating plugin_routes table: %w", tblErr)
		}

		if m.tableReg != nil {
			if regErr := m.tableReg.RegisterTable(ctx, "plugin_routes"); regErr != nil {
				utility.DefaultLogger.Warn(
					fmt.Sprintf("failed to register plugin_routes in tables registry: %s", regErr.Error()),
					nil,
				)
			}
		}

		discoveredNames := make([]string, 0, len(discovered))
		for name := range discovered {
			discoveredNames = append(discoveredNames, name)
		}
		if cleanErr := m.bridge.CleanupOrphanedRoutes(ctx, discoveredNames); cleanErr != nil {
			return fmt.Errorf("cleaning orphaned routes: %w", cleanErr)
		}
	}

	// Phase 3: Create plugin_hooks table and clean up orphaned hooks.
	if m.hookEngine != nil {
		if tblErr := m.hookEngine.CreatePluginHooksTable(ctx); tblErr != nil {
			return fmt.Errorf("creating plugin_hooks table: %w", tblErr)
		}

		if m.tableReg != nil {
			if regErr := m.tableReg.RegisterTable(ctx, "plugin_hooks"); regErr != nil {
				utility.DefaultLogger.Warn(
					fmt.Sprintf("failed to register plugin_hooks in tables registry: %s", regErr.Error()),
					nil,
				)
			}
		}

		discoveredNames := make([]string, 0, len(discovered))
		for name := range discovered {
			discoveredNames = append(discoveredNames, name)
		}
		if cleanErr := m.hookEngine.CleanupOrphanedHooks(ctx, discoveredNames); cleanErr != nil {
			return fmt.Errorf("cleaning orphaned hooks: %w", cleanErr)
		}
	}

	// Create plugin_requests table and clean up orphaned request registrations.
	if m.requestEngine != nil {
		if tblErr := m.requestEngine.CreatePluginRequestsTable(ctx); tblErr != nil {
			return fmt.Errorf("creating plugin_requests table: %w", tblErr)
		}

		if m.tableReg != nil {
			if regErr := m.tableReg.RegisterTable(ctx, "plugin_requests"); regErr != nil {
				utility.DefaultLogger.Warn(
					fmt.Sprintf("failed to register plugin_requests in tables registry: %s", regErr.Error()),
					nil,
				)
			}
		}

		discoveredNames := make([]string, 0, len(discovered))
		for name := range discovered {
			discoveredNames = append(discoveredNames, name)
		}
		if cleanErr := m.requestEngine.CleanupOrphanedRequests(ctx, discoveredNames); cleanErr != nil {
			return fmt.Errorf("cleaning orphaned requests: %w", cleanErr)
		}

		if loadErr := m.requestEngine.LoadApprovals(ctx); loadErr != nil {
			return fmt.Errorf("loading request approvals: %w", loadErr)
		}
	}

	// Step 4: Load each plugin in dependency order.
	m.mu.Lock()
	for _, inst := range sorted {
		m.plugins[inst.Info.Name] = inst
		m.loadPlugin(ctx, inst)
	}
	m.mu.Unlock()

	// Step 5: Load the pipeline registry from the database.
	if regErr := m.LoadRegistry(ctx); regErr != nil {
		utility.DefaultLogger.Warn(
			fmt.Sprintf("failed to load pipeline registry: %s", regErr.Error()),
			nil,
		)
	}

	return nil
}

// loadPlugin initializes a single plugin: validates dependencies, creates the
// VM pool, runs on_init, and takes the global snapshot. On any error, the
// plugin is marked StateFailed with a reason and loading continues.
func (m *Manager) loadPlugin(ctx context.Context, inst *PluginInstance) {
	inst.State = StateLoading
	pluginName := inst.Info.Name

	// Validate that all dependencies are running.
	for _, dep := range inst.Info.Dependencies {
		depInst, exists := m.plugins[dep]
		if !exists {
			m.failPlugin(inst, fmt.Sprintf("dependency %q not found", dep))
			return
		}
		if depInst.State != StateRunning {
			m.failPlugin(inst, fmt.Sprintf("dependency %q is in state %s, not running", dep, depInst.State))
			return
		}
	}

	timeout := time.Duration(m.cfg.ExecTimeoutSec) * time.Second

	// Create VM factory -- produces fully sandboxed VMs with db/log/http/hooks APIs.
	factory := func() *lua.LState {
		L := lua.NewState(lua.Options{
			SkipOpenLibs:  true,
			CallStackSize: 256,
			RegistrySize:  5120,
		})

		// VM factory call sequence (roadmap-specified order):
		// 1. ApplySandbox
		// 2. Set __vm_phase = "module_scope"
		// 3. RegisterPluginRequire
		// 4. RegisterDBAPI
		// 5. RegisterLogAPI
		// 6. RegisterHTTPAPI (Phase 2)
		// 7. RegisterHooksAPI (Phase 3)
		// 8. RegisterRequestAPI
		// 9. FreezeModule (all modules)
		ApplySandbox(L, SandboxConfig{AllowCoroutine: true, ExecTimeout: timeout})
		setVMPhase(L, "module_scope")
		RegisterPluginRequire(L, inst.Dir)

		dbAPI := NewDatabaseAPI(m.db, pluginName, m.dialect, m.cfg.MaxOpsPerExec, m.tableReg)
		RegisterDBAPI(L, dbAPI)
		RegisterLogAPI(L, pluginName)
		RegisterHTTPAPI(L, pluginName)
		RegisterHooksAPI(L, pluginName)
		coreAPI := NewCoreTableAPI(dbAPI, pluginName, m.dialect, inst.ApprovedAccess)
		RegisterCoreAPI(L, coreAPI)
		reqAPI := RegisterRequestAPI(L, pluginName, m.requestEngine)
		FreezeModule(L, "db")
		FreezeModule(L, "log")
		FreezeModule(L, "http")
		FreezeModule(L, "hooks")
		FreezeModule(L, "core")
		FreezeModule(L, "request")

		// Execute init.lua to define globals (plugin_info, on_init, on_shutdown).
		// Use a context with timeout for the execution.
		initCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		L.SetContext(initCtx)

		if err := L.DoFile(inst.InitPath); err != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: init.lua execution error in factory: %s", pluginName, err.Error()),
				nil,
			)
			// The VM is still usable -- init.lua may have partially executed.
			// The health check on Put() will catch critical corruption.
		}

		// Clear context after factory init -- the caller will set their own via Get().
		L.SetContext(nil)

		// Store the dbAPI and requestAPI mappings on the instance (S5: protected by inst.mu).
		inst.mu.Lock()
		inst.dbAPIs[L] = dbAPI
		inst.requestAPIs[L] = reqAPI
		inst.mu.Unlock()

		return L
	}

	// Create the VM pool with reserved VMs for hooks (M2).
	// onReplace callback cleans up dbAPIs entries when unhealthy VMs are replaced (S5).
	//
	// Reserve VMs are only allocated when the pool is large enough (>= 2 VMs).
	// A pool of 1 VM cannot afford a reserve -- all VMs go to general.
	reserveSize := m.cfg.HookReserveVMs
	if m.cfg.MaxVMsPerPlugin <= 1 {
		reserveSize = 0
	} else if reserveSize >= m.cfg.MaxVMsPerPlugin {
		reserveSize = m.cfg.MaxVMsPerPlugin - 1
	}
	pool := NewVMPool(VMPoolConfig{
		Size:        m.cfg.MaxVMsPerPlugin,
		ReserveSize: reserveSize,
		Factory:     factory,
		InitPath:    inst.InitPath,
		PluginName:  pluginName,
		OnReplace: func(oldL *lua.LState) {
			inst.mu.Lock()
			delete(inst.dbAPIs, oldL)
			delete(inst.requestAPIs, oldL)
			inst.mu.Unlock()
		},
	})
	inst.Pool = pool

	// Phase 4: Create circuit breaker for this plugin instance.
	inst.CB = NewCircuitBreaker(pluginName, m.cfg.MaxFailures, m.cfg.ResetInterval)

	// Check out one VM to run on_init().
	initCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	L, getErr := pool.Get(initCtx)
	if getErr != nil {
		m.failPlugin(inst, fmt.Sprintf("failed to get VM from pool: %s", getErr.Error()))
		return
	}

	// Reset op count before plugin code executes.
	inst.mu.Lock()
	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}
	inst.mu.Unlock()

	// Set context with timeout for on_init execution.
	L.SetContext(initCtx)

	// Set VM phase to "init" so http.handle()/hooks.on()/request.register() reject
	// calls during on_init. These must only be called at module scope (during init.lua
	// execution by the factory). The phase is stored in the LState's registry table,
	// which is inaccessible from Lua code.
	setVMPhase(L, "init")

	// Call on_init() if defined.
	onInit := L.GetGlobal("on_init")
	if fn, ok := onInit.(*lua.LFunction); ok {
		if callErr := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}); callErr != nil {
			// Set phase to runtime before returning VM to pool.
			setVMPhase(L, "runtime")
			pool.Put(L)
			m.failPlugin(inst, fmt.Sprintf("on_init failed: %s", callErr.Error()))
			return
		}
	}

	// Set phase to runtime after on_init completes.
	setVMPhase(L, "runtime")

	// Take global snapshot after on_init (captures any globals defined during init).
	pool.SnapshotGlobals(L)

	// Phase 2: Register HTTP routes discovered in this VM's __http_handlers table.
	// This must happen after SnapshotGlobals (so route handler globals are captured)
	// and before Put (so the VM is still checked out and safe to read from).
	if m.bridge != nil {
		if regErr := m.bridge.RegisterRoutes(ctx, pluginName, inst.Info.Version, L); regErr != nil {
			pool.Put(L)
			m.failPlugin(inst, fmt.Sprintf("route registration failed: %s", regErr.Error()))
			return
		}
	}

	// Phase 3: Register content lifecycle hooks from this VM's __hook_pending table.
	// Reads pending hooks from Lua state, registers them with the HookEngine,
	// and upserts approval records into the plugin_hooks DB table.
	if m.hookEngine != nil {
		pendingHooks := ReadPendingHooks(L)
		if len(pendingHooks) > 0 {
			m.hookEngine.RegisterHooks(pluginName, pendingHooks)

			if upsertErr := m.hookEngine.UpsertHookRegistrations(ctx, pluginName, inst.Info.Version, pendingHooks); upsertErr != nil {
				pool.Put(L)
				m.failPlugin(inst, fmt.Sprintf("hook registration failed: %s", upsertErr.Error()))
				return
			}

			utility.DefaultLogger.Info(
				fmt.Sprintf("plugin %q: registered %d content hooks", pluginName, len(pendingHooks)),
			)
		}
	}

	// Read pending request registrations and upsert into the DB.
	pendingRequests := ReadPendingRequests(L)
	if len(pendingRequests) > 0 {
		utility.DefaultLogger.Info(
			fmt.Sprintf("plugin %q: discovered %d outbound request domain registrations", pluginName, len(pendingRequests)),
		)
		if m.requestEngine != nil {
			if upsertErr := m.requestEngine.UpsertRequestRegistrations(ctx, pluginName, inst.Info.Version, pendingRequests); upsertErr != nil {
				utility.DefaultLogger.Error(
					fmt.Sprintf("plugin %q: failed to upsert request registrations: %s", pluginName, upsertErr.Error()),
					nil,
				)
			}
		}
	}

	// Phase 4: Read schema drift entries from the VM (set by luaDefineTable).
	driftVal := L.GetGlobal("__schema_drift")
	if driftTbl, ok := driftVal.(*lua.LTable); ok && driftVal != lua.LNil {
		driftTbl.ForEach(func(_, value lua.LValue) {
			entry, ok := value.(*lua.LTable)
			if !ok {
				return
			}
			tblField := L.GetField(entry, "table")
			kindField := L.GetField(entry, "kind")
			colField := L.GetField(entry, "column")

			tblStr, _ := tblField.(lua.LString)
			kindStr, _ := kindField.(lua.LString)
			colStr, _ := colField.(lua.LString)

			inst.SchemaDrift = append(inst.SchemaDrift, DriftEntry{
				Table:  string(tblStr),
				Kind:   string(kindStr),
				Column: string(colStr),
			})
		})
	}

	// Return VM to pool.
	pool.Put(L)

	// Mark plugin as running.
	inst.State = StateRunning
	m.loadOrder = append(m.loadOrder, pluginName)

	utility.DefaultLogger.Info(
		fmt.Sprintf("plugin %q loaded (state: running)", pluginName),
	)
}

// failPlugin sets a plugin to StateFailed with the given reason and logs the error.
func (m *Manager) failPlugin(inst *PluginInstance, reason string) {
	inst.State = StateFailed
	inst.FailedReason = reason
	utility.DefaultLogger.Error(
		fmt.Sprintf("plugin %q failed: %s", inst.Info.Name, reason),
		nil,
	)
}

// Shutdown gracefully shuts down all plugins in reverse dependency order.
// For each plugin: checks out a VM, calls on_shutdown if defined, returns VM.
// After all on_shutdown calls: closes all VM pools and the plugin DB pool.
func (m *Manager) Shutdown(ctx context.Context) {
	m.shutdownOnce.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		// Reverse dependency order: dependents shut down before their dependencies.
		for i := len(m.loadOrder) - 1; i >= 0; i-- {
			name := m.loadOrder[i]
			inst, exists := m.plugins[name]
			if !exists || inst.State != StateRunning {
				continue
			}

			m.shutdownPlugin(ctx, inst)
		}

		// Close all VM pools (including failed plugins that may have pools).
		for _, inst := range m.plugins {
			if inst.Pool != nil {
				inst.Pool.Close()
			}
		}

		// Close the plugin DB pool.
		if m.db != nil {
			if err := m.db.Close(); err != nil {
				utility.DefaultLogger.Warn(
					fmt.Sprintf("plugin db pool close error: %s", err.Error()),
					nil,
				)
			}
		}
	})
}

// shutdownPlugin runs on_shutdown for a single plugin if defined.
func (m *Manager) shutdownPlugin(ctx context.Context, inst *PluginInstance) {
	// Use a generous shutdown timeout (10s or the remaining context deadline).
	shutdownTimeout := 10 * time.Second
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	L, getErr := inst.Pool.Get(shutdownCtx)
	if getErr != nil {
		utility.DefaultLogger.Warn(
			fmt.Sprintf("plugin %q: could not get VM for shutdown (skipping on_shutdown): %s",
				inst.Info.Name, getErr.Error()),
			nil,
		)
		inst.State = StateStopped
		return
	}

	// Reset op count for shutdown execution.
	inst.mu.Lock()
	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}
	inst.mu.Unlock()

	L.SetContext(shutdownCtx)

	// Set VM phase to "shutdown" so request.send() and registration APIs
	// (http.handle, hooks.on) are blocked during on_shutdown().
	setVMPhase(L, "shutdown")

	onShutdown := L.GetGlobal("on_shutdown")
	if fn, ok := onShutdown.(*lua.LFunction); ok {
		if callErr := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}); callErr != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: on_shutdown error: %s", inst.Info.Name, callErr.Error()),
				nil,
			)
		}
	}

	inst.Pool.Put(L)
	inst.State = StateStopped
}

// GetPlugin returns the plugin instance by name, or nil if not found.
// Thread-safe via read lock.
func (m *Manager) GetPlugin(name string) *PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[name]
}

// ListPlugins returns all loaded plugin instances.
// Thread-safe via read lock.
func (m *Manager) ListPlugins() []*PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PluginInstance, 0, len(m.plugins))
	for _, inst := range m.plugins {
		result = append(result, inst)
	}
	return result
}

// ExtractManifest creates a temporary sandboxed VM (no db/log APIs), executes
// init.lua, reads the plugin_info global, and validates the manifest fields.
// The temp VM is discarded after extraction.
func ExtractManifest(initPath string) (*PluginInfo, error) {
	L := lua.NewState(lua.Options{
		SkipOpenLibs:  true,
		CallStackSize: 256,
		RegistrySize:  5120,
	})
	defer L.Close()

	// Apply sandbox and register the sandboxed require loader so that plugins
	// using require() at file scope (outside on_init) can have their manifest
	// extracted. No db/log APIs are registered for manifest extraction -- those
	// are only available during full plugin loading.
	ApplySandbox(L, SandboxConfig{})
	RegisterPluginRequire(L, filepath.Dir(initPath))

	// Register no-op stub globals for db, log, http, and hooks so that
	// module-scope API calls (hooks.on, http.handle, etc.) don't crash
	// during manifest extraction. The stubs absorb calls without side effects.
	registerManifestStubs(L)

	// Set a short timeout for manifest extraction (2 seconds).
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	L.SetContext(ctx)

	if err := L.DoFile(initPath); err != nil {
		return nil, fmt.Errorf("executing init.lua: %w", err)
	}

	// Read plugin_info global.
	infoVal := L.GetGlobal("plugin_info")
	infoTbl, ok := infoVal.(*lua.LTable)
	if !ok || infoVal == lua.LNil {
		return nil, fmt.Errorf("missing required plugin_info global")
	}

	info := &PluginInfo{}

	// Extract required fields.
	info.Name = luaTableString(L, infoTbl, "name")
	info.Version = luaTableString(L, infoTbl, "version")
	info.Description = luaTableString(L, infoTbl, "description")

	// Extract optional fields.
	info.Author = luaTableString(L, infoTbl, "author")
	info.License = luaTableString(L, infoTbl, "license")
	info.MinCMSVersion = luaTableString(L, infoTbl, "min_cms_version")

	// Extract dependencies (optional list of strings).
	depsVal := L.GetField(infoTbl, "dependencies")
	if depsTbl, ok := depsVal.(*lua.LTable); ok {
		depsTbl.ForEach(func(_, v lua.LValue) {
			if s, ok := v.(lua.LString); ok {
				info.Dependencies = append(info.Dependencies, string(s))
			}
		})
	}

	// Extract capabilities (optional list of {table, op, handler, priority}).
	capsVal := L.GetField(infoTbl, "capabilities")
	if capsTbl, ok := capsVal.(*lua.LTable); ok {
		capsTbl.ForEach(func(_, v lua.LValue) {
			entry, ok := v.(*lua.LTable)
			if !ok {
				return
			}
			cap := PluginCapability{
				Table:   luaTableString(L, entry, "table"),
				Op:      luaTableString(L, entry, "op"),
				Handler: luaTableString(L, entry, "handler"),
			}
			priVal := L.GetField(entry, "priority")
			if num, ok := priVal.(lua.LNumber); ok {
				cap.Priority = int(num)
			}
			info.Capabilities = append(info.Capabilities, cap)
		})
	}

	// Extract core_access (optional table of string arrays).
	accessVal := L.GetField(infoTbl, "core_access")
	if accessTbl, ok := accessVal.(*lua.LTable); ok {
		info.CoreAccess = make(PluginCoreAccess)
		accessTbl.ForEach(func(key, val lua.LValue) {
			tableName, ok := key.(lua.LString)
			if !ok {
				return
			}
			opsTbl, ok := val.(*lua.LTable)
			if !ok {
				return
			}
			var ops []string
			opsTbl.ForEach(func(_, opVal lua.LValue) {
				if s, ok := opVal.(lua.LString); ok {
					ops = append(ops, string(s))
				}
			})
			info.CoreAccess[string(tableName)] = ops
		})
	}

	// Check whether on_init() is defined before the deferred L.Close() runs.
	if L.GetGlobal("on_init").Type() == lua.LTFunction {
		info.HasOnInit = true
	}

	// Validate manifest.
	if err := ValidateManifest(info); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return info, nil
}

// ValidateManifest checks that all required fields are present and valid.
func ValidateManifest(info *PluginInfo) error {
	if info.Name == "" {
		return fmt.Errorf("name is required")
	}
	if info.Version == "" {
		return fmt.Errorf("version is required")
	}
	if info.Description == "" {
		return fmt.Errorf("description is required")
	}

	// Validate name: [a-z0-9_], max 32 chars, no trailing underscore.
	if len(info.Name) > 32 {
		return fmt.Errorf("name %q exceeds 32 characters", info.Name)
	}
	for _, r := range info.Name {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("name %q contains invalid character %q (must be [a-z0-9_])", info.Name, string(r))
		}
	}
	// Reject trailing underscore to prevent table prefix collisions
	// (e.g., plugin "a_" would produce prefix "plugin_a__" which collides
	// with plugin "a" prefix "plugin_a_" in ambiguous ways).
	if strings.HasSuffix(info.Name, "_") {
		return fmt.Errorf("name %q has trailing underscore (collision prevention)", info.Name)
	}

	return nil
}

// luaTableString reads a string field from a Lua table.
// Returns empty string if the field is absent or not a string.
func luaTableString(L *lua.LState, tbl *lua.LTable, field string) string {
	val := L.GetField(tbl, field)
	if s, ok := val.(lua.LString); ok {
		return string(s)
	}
	return ""
}

// topologicalSort orders plugins by their dependency relationships using
// Kahn's algorithm. Returns an error if a circular dependency is detected.
func topologicalSort(plugins map[string]*PluginInstance) ([]*PluginInstance, error) {
	if len(plugins) == 0 {
		return nil, nil
	}

	// Build adjacency list and in-degree count.
	// An edge from A -> B means "A depends on B" (B must load before A).
	inDegree := make(map[string]int)
	dependents := make(map[string][]string) // B -> [A] means A depends on B

	for name := range plugins {
		inDegree[name] = 0
	}

	for name, inst := range plugins {
		for _, dep := range inst.Info.Dependencies {
			if _, exists := plugins[dep]; !exists {
				return nil, fmt.Errorf("plugin %q depends on %q which was not discovered", name, dep)
			}
			dependents[dep] = append(dependents[dep], name)
			inDegree[name]++
		}
	}

	// Start with nodes that have no dependencies (in-degree 0).
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Sort the initial queue for deterministic ordering.
	sortStringSlice(queue)

	var result []*PluginInstance
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		result = append(result, plugins[current])

		// Process dependents: reduce their in-degree.
		deps := dependents[current]
		sortStringSlice(deps)
		for _, dependent := range deps {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// If we did not process all plugins, there is a cycle.
	if len(result) != len(plugins) {
		// Find the cycle participants for a useful error message.
		var cycled []string
		for name, degree := range inDegree {
			if degree > 0 {
				cycled = append(cycled, name)
			}
		}
		sortStringSlice(cycled)
		return nil, fmt.Errorf("circular dependency detected among plugins: %s", strings.Join(cycled, ", "))
	}

	return result, nil
}

// registerManifestStubs sets up no-op global tables for db, log, http, and hooks
// in a temporary VM used only for manifest extraction. These stubs absorb
// module-scope API calls (hooks.on, http.handle, db.define_table, etc.) without
// side effects, allowing ExtractManifest to read plugin_info even when the
// plugin's init.lua calls API functions at file scope.
//
// The approach uses a metatable with __index returning a no-op function for any
// key access, so all method calls like hooks.on(...) resolve to a function that
// accepts any arguments and returns nothing.
func registerManifestStubs(L *lua.LState) {
	noop := L.NewFunction(func(L *lua.LState) int {
		return 0
	})

	// Create a metatable that returns the noop function for any key lookup.
	// This handles both known methods (hooks.on, http.handle) and any future
	// additions without needing to enumerate every method name.
	mt := L.NewTable()
	mt.RawSetString("__index", L.NewFunction(func(L *lua.LState) int {
		L.Push(noop)
		return 1
	}))

	for _, name := range []string{"db", "log", "http", "hooks", "core", "request"} {
		stub := L.NewTable()
		L.SetMetatable(stub, mt)
		L.SetGlobal(name, stub)
	}
}

// sortStringSlice sorts a string slice in place for deterministic ordering.
func sortStringSlice(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// compareCapabilities detects differences between filesystem capabilities (current)
// and DB-stored capabilities (stored). Returns nil if they are identical or both empty.
func compareCapabilities(current, stored []PluginCapability) []CapabilityDriftEntry {
	type capKey struct {
		Table string
		Op    string
	}

	currentMap := make(map[capKey]PluginCapability, len(current))
	for _, c := range current {
		currentMap[capKey{c.Table, c.Op}] = c
	}

	storedMap := make(map[capKey]PluginCapability, len(stored))
	for _, s := range stored {
		storedMap[capKey{s.Table, s.Op}] = s
	}

	var drifts []CapabilityDriftEntry

	// Check for added or changed capabilities.
	for key, cur := range currentMap {
		prev, exists := storedMap[key]
		if !exists {
			c := cur
			drifts = append(drifts, CapabilityDriftEntry{
				Kind:    "added",
				Current: &c,
			})
			continue
		}
		if cur.Handler != prev.Handler || cur.Priority != prev.Priority {
			c := cur
			p := prev
			drifts = append(drifts, CapabilityDriftEntry{
				Kind:     "changed",
				Current:  &c,
				Previous: &p,
			})
		}
	}

	// Check for removed capabilities.
	for key, prev := range storedMap {
		if _, exists := currentMap[key]; !exists {
			p := prev
			drifts = append(drifts, CapabilityDriftEntry{
				Kind:     "removed",
				Previous: &p,
			})
		}
	}

	if len(drifts) == 0 {
		return nil
	}
	return drifts
}

// StartWatcher creates and starts the file-polling hot reload watcher.
// The watcher runs in a goroutine until ctx is cancelled or StopWatcher is called.
// Must be called after LoadAll.
func (m *Manager) StartWatcher(ctx context.Context) {
	watchCtx, cancel := context.WithCancel(ctx)
	m.watcher = NewWatcher(m, 2*time.Second)
	m.watcherCancel = cancel
	m.watcher.InitialChecksums()

	go m.watcher.Run(watchCtx)

	utility.DefaultLogger.Info("Plugin hot reload watcher started",
		"poll_interval", "2s",
		"debounce_delay", "1s",
	)
}

// StopWatcher stops the file-polling watcher if running. Idempotent.
// S11: Must be called before bridge.Close() in shutdown sequence to prevent
// reload racing with shutdown.
func (m *Manager) StopWatcher() {
	if m.watcherCancel != nil {
		m.watcherCancel()
		m.watcherCancel = nil
	}
}

// ReloadPlugin performs a blue-green reload of a single plugin (A3).
// Creates a new instance first; only drains the old after the new is confirmed working.
// If the new instance fails, the old keeps running with a logged warning.
func (m *Manager) ReloadPlugin(ctx context.Context, name string) error {
	m.mu.RLock()
	oldInst, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %q not found", name)
	}

	// Re-extract manifest (version may have changed).
	newInfo, extractErr := ExtractManifest(oldInst.InitPath)
	if extractErr != nil {
		return fmt.Errorf("re-extracting manifest for %q: %w", name, extractErr)
	}

	// Check dependency validity for the new manifest.
	m.mu.RLock()
	for _, dep := range newInfo.Dependencies {
		depInst, depExists := m.plugins[dep]
		if !depExists || depInst.State != StateRunning {
			m.mu.RUnlock()
			return fmt.Errorf("dependency %q for plugin %q is not running (aborting reload)", dep, name)
		}
	}
	m.mu.RUnlock()

	// Create new plugin instance independently (shares nothing with old).
	newInst := &PluginInstance{
		Info:        *newInfo,
		State:       StateDiscovered,
		Dir:         oldInst.Dir,
		InitPath:    oldInst.InitPath,
		dbAPIs:      make(map[*lua.LState]*DatabaseAPI),
		requestAPIs: make(map[*lua.LState]*requestAPIState),
	}

	// Load the new instance (creates pool, runs on_init, registers routes/hooks).
	m.mu.Lock()
	m.plugins[name] = newInst // temporarily set for loadPlugin to see dependencies
	m.loadPlugin(ctx, newInst)
	m.mu.Unlock()

	if newInst.State == StateFailed {
		// Rollback: restore old instance.
		m.mu.Lock()
		m.plugins[name] = oldInst
		m.mu.Unlock()

		if newInst.Pool != nil {
			newInst.Pool.Close()
		}

		RecordReload(name, "error")
		return fmt.Errorf("plugin %q reload failed: %s (old version still running)",
			name, newInst.FailedReason)
	}

	// New instance loaded successfully.
	oldInst.State = StateLoading // prevents bridge/hooks from using old instance

	// Unregister old instance's routes and hooks from in-memory indexes.
	if m.bridge != nil {
		m.bridge.UnregisterPlugin(name)
	}
	if m.hookEngine != nil {
		m.hookEngine.UnregisterPlugin(name)
	}
	if m.requestEngine != nil {
		m.requestEngine.UnregisterPlugin(name)
	}

	// The swap already happened in the loadPlugin block above.
	// Now drain the old pool.
	if oldInst.Pool != nil {
		drained := oldInst.Pool.Drain(10 * time.Second)
		if !drained {
			// S6: Drain timeout -- trip the new instance's circuit breaker.
			if newInst.CB != nil {
				newInst.CB.Trip("drain timeout during reload (stuck old handlers)")
			}
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: old pool drain timed out during reload, new instance CB tripped", name),
				nil,
			)
		}
	}

	RecordReload(name, "success")
	utility.DefaultLogger.Info(
		fmt.Sprintf("plugin %q reloaded successfully (version: %s -> %s)",
			name, oldInst.Info.Version, newInst.Info.Version),
	)

	return nil
}

// DeactivatePlugin stops a plugin and trips its circuit breaker.
// The plugin can be re-activated via ActivatePlugin.
// Idempotent: returns nil if the plugin is already stopped.
func (m *Manager) DeactivatePlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %q not found", name)
	}

	if inst.State == StateStopped {
		return nil // idempotent: already stopped
	}

	// Run on_shutdown if the plugin was running.
	if inst.State == StateRunning {
		m.shutdownPlugin(ctx, inst)
	}

	inst.State = StateStopped
	if inst.CB != nil {
		inst.CB.Trip("disabled by admin")
	}

	utility.DefaultLogger.Info(
		fmt.Sprintf("plugin %q disabled by admin", name),
	)

	// Update DB status to installed so other instances see the change.
	if m.driver != nil {
		record, getErr := m.driver.GetPluginByName(name)
		if getErr != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: failed to look up DB record for status update: %s", name, getErr.Error()),
				nil,
			)
		} else {
			ac := audited.Ctx(types.NodeID(m.cfg.NodeID), types.UserID(""), "plugin-deactivate", "127.0.0.1")
			if updateErr := m.driver.UpdatePluginStatus(ctx, ac, record.PluginID, types.PluginStatusInstalled); updateErr != nil {
				utility.DefaultLogger.Warn(
					fmt.Sprintf("plugin %q: failed to update DB status to installed: %s", name, updateErr.Error()),
					nil,
				)
			}
		}
	}

	return nil
}

// ActivatePlugin resets the circuit breaker and reloads the plugin.
// S5: Emits slog.Warn audit event with admin user, plugin name, prior CB state,
// and failure count before reset.
func (m *Manager) ActivatePlugin(ctx context.Context, name string, adminUser string) error {
	m.mu.RLock()
	inst, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %q not found", name)
	}

	// Reset circuit breaker (S5: emits audit event).
	if inst.CB != nil {
		inst.CB.Reset(adminUser)
	}

	// Reload the plugin to get it running again.
	if err := m.ReloadPlugin(ctx, name); err != nil {
		return err
	}

	// Update DB status to enabled so other instances see the change.
	if m.driver != nil {
		record, getErr := m.driver.GetPluginByName(name)
		if getErr != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: failed to look up DB record for status update: %s", name, getErr.Error()),
				nil,
			)
		} else {
			ac := audited.Ctx(types.NodeID(m.cfg.NodeID), types.UserID(adminUser), "plugin-activate", "127.0.0.1")
			if updateErr := m.driver.UpdatePluginStatus(ctx, ac, record.PluginID, types.PluginStatusEnabled); updateErr != nil {
				utility.DefaultLogger.Warn(
					fmt.Sprintf("plugin %q: failed to update DB status to enabled: %s", name, updateErr.Error()),
					nil,
				)
			}
		}
	}

	return nil
}

// EvictPlugin stops a running plugin, marks it as failed in-memory, and updates
// the DB status to installed. Called when plugin files disappear from disk or
// when a plugin is removed from the database by another instance.
// Idempotent: returns nil if the plugin is not found.
func (m *Manager) EvictPlugin(ctx context.Context, name string, reason string) error {
	m.mu.Lock()

	inst, exists := m.plugins[name]
	if !exists {
		m.mu.Unlock()
		return nil // idempotent
	}

	// If running, shut it down.
	if inst.State == StateRunning {
		m.shutdownPlugin(ctx, inst)
	}

	inst.State = StateFailed
	inst.FailedReason = reason
	if inst.CB != nil {
		inst.CB.Trip(reason)
	}

	m.mu.Unlock()

	utility.DefaultLogger.Warn(
		fmt.Sprintf("plugin %q evicted: %s", name, reason),
		nil,
	)

	// Update DB status to installed.
	if m.driver != nil {
		record, getErr := m.driver.GetPluginByName(name)
		if getErr != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: failed to look up DB record for eviction: %s", name, getErr.Error()),
				nil,
			)
		} else {
			ac := audited.Ctx(types.NodeID(m.cfg.NodeID), types.UserID(""), "plugin-eviction", "127.0.0.1")
			if updateErr := m.driver.UpdatePluginStatus(ctx, ac, record.PluginID, types.PluginStatusInstalled); updateErr != nil {
				utility.DefaultLogger.Warn(
					fmt.Sprintf("plugin %q: failed to update DB status after eviction: %s", name, updateErr.Error()),
					nil,
				)
			}
		}
	}

	return nil
}

// GetPluginState returns the current state of a plugin and whether it exists.
// Thread-safe via read lock.
func (m *Manager) GetPluginState(name string) (PluginState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inst, exists := m.plugins[name]
	if !exists {
		return 0, false
	}
	return inst.State, true
}

// PluginHealthStatus reports the aggregate health of the plugin subsystem.
type PluginHealthStatus struct {
	Healthy             bool
	TotalPlugins        int
	RunningPlugins      int
	FailedPlugins       []string
	StoppedPlugins      []string
	OpenCircuitBreakers []string
}

// PluginHealth returns the aggregate health of all loaded plugins.
// Healthy is true when no enabled plugins are failed and no circuit breakers are open.
func (m *Manager) PluginHealth() PluginHealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := PluginHealthStatus{
		Healthy: true,
	}

	for name, inst := range m.plugins {
		status.TotalPlugins++

		switch inst.State {
		case StateRunning:
			status.RunningPlugins++
		case StateFailed:
			status.FailedPlugins = append(status.FailedPlugins, name)
			status.Healthy = false
		case StateStopped:
			status.StoppedPlugins = append(status.StoppedPlugins, name)
		}

		if inst.CB != nil && inst.CB.State() == CircuitOpen {
			status.OpenCircuitBreakers = append(status.OpenCircuitBreakers, name)
			status.Healthy = false
		}
	}

	return status
}

// StartCoordinator starts the DB state polling coordinator for multi-instance sync.
// Only starts if the driver is non-nil. If seed fails (DB unreachable), logs a warning
// and returns without starting the poll loop.
func (m *Manager) StartCoordinator(ctx context.Context, syncInterval time.Duration) {
	if m.driver == nil {
		utility.DefaultLogger.Warn("coordinator not started: no database driver", nil)
		return
	}

	coord := NewCoordinator(m, syncInterval)

	if err := coord.seed(ctx); err != nil {
		utility.DefaultLogger.Warn(
			fmt.Sprintf("coordinator seed failed, multi-instance sync disabled: %s", err.Error()),
			nil,
		)
		return
	}

	coordCtx, cancel := context.WithCancel(ctx)
	m.coordinator = coord
	m.coordinatorCancel = cancel

	go coord.Run(coordCtx)

	utility.DefaultLogger.Info("Plugin coordinator started",
		"poll_interval", syncInterval.String(),
	)
}

// StopCoordinator stops the DB state polling coordinator if running. Idempotent.
func (m *Manager) StopCoordinator() {
	if m.coordinatorCancel != nil {
		m.coordinatorCancel()
		m.coordinatorCancel = nil
	}
}

// InstallPlugin extracts the manifest for a discovered filesystem plugin, computes
// the manifest hash, and creates the plugin row and pipeline entries in a single
// database transaction. Returns the created db.Plugin record.
func (m *Manager) InstallPlugin(ctx context.Context, name string) (*db.Plugin, error) {
	if m.driver == nil {
		return nil, fmt.Errorf("plugin lifecycle requires a database driver")
	}

	// Find the plugin directory on the filesystem.
	pluginDir := filepath.Join(m.cfg.Directory, name)
	initPath := filepath.Join(pluginDir, "init.lua")

	if _, err := os.Stat(initPath); err != nil {
		return nil, fmt.Errorf("plugin %q not found on filesystem (no init.lua at %s)", name, initPath)
	}

	// Extract manifest.
	info, err := ExtractManifest(initPath)
	if err != nil {
		return nil, fmt.Errorf("extracting manifest for %q: %w", name, err)
	}

	// Check if already installed.
	existing, _ := m.driver.GetPluginByName(info.Name)
	if existing != nil {
		return nil, fmt.Errorf("plugin %q is already installed (id: %s)", info.Name, existing.PluginID)
	}

	// Compute manifest hash.
	hash, err := computeChecksum(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("computing manifest hash for %q: %w", name, err)
	}

	// Serialize capabilities and core access.
	capJSON, err := json.Marshal(info.Capabilities)
	if err != nil {
		return nil, fmt.Errorf("marshaling capabilities: %w", err)
	}
	accessJSON, err := json.Marshal(info.CoreAccess)
	if err != nil {
		return nil, fmt.Errorf("marshaling core access: %w", err)
	}

	// Create the plugin record via driver (audited).
	ac := audited.Ctx(types.NodeID(m.cfg.NodeID), types.UserID(""), "install", "cli")
	pluginRecord, err := m.driver.CreatePlugin(ctx, ac, db.CreatePluginParams{
		Name:           info.Name,
		Version:        info.Version,
		Description:    info.Description,
		Author:         info.Author,
		Status:         types.PluginStatusInstalled,
		Capabilities:   string(capJSON),
		ApprovedAccess: string(accessJSON),
		ManifestHash:   hash,
	})
	if err != nil {
		return nil, fmt.Errorf("creating plugin record for %q: %w", name, err)
	}

	// Create pipeline entries from the manifest capabilities.
	for _, cap := range info.Capabilities {
		_, pipeErr := m.driver.CreatePipeline(ctx, ac, db.CreatePipelineParams{
			PluginID:   pluginRecord.PluginID,
			TableName:  cap.Table,
			Operation:  cap.Op,
			PluginName: info.Name,
			Handler:    cap.Handler,
			Priority:   cap.Priority,
			Enabled:    true,
			Config:     types.NewJSONData("{}"),
		})
		if pipeErr != nil {
			// Best-effort cleanup: delete the plugin record if pipeline creation fails.
			_ = m.driver.DeletePlugin(ctx, ac, pluginRecord.PluginID)
			return nil, fmt.Errorf("creating pipeline entry for %q (%s.%s): %w",
				name, cap.Table, cap.Op, pipeErr)
		}
	}

	utility.DefaultLogger.Info(
		fmt.Sprintf("plugin %q installed (version: %s, capabilities: %d)",
			info.Name, info.Version, len(info.Capabilities)),
	)

	return pluginRecord, nil
}

// EnablePlugin updates the database status to "enabled" and loads the plugin into
// the runtime. If the plugin is not installed in the database, returns an error.
func (m *Manager) EnablePlugin(ctx context.Context, name string) error {
	if m.driver == nil {
		return fmt.Errorf("plugin lifecycle requires a database driver")
	}

	record, err := m.driver.GetPluginByName(name)
	if err != nil {
		return fmt.Errorf("plugin %q not found in database: %w", name, err)
	}

	if record.Status == types.PluginStatusEnabled {
		return fmt.Errorf("plugin %q is already enabled", name)
	}

	// Update DB status to enabled.
	ac := audited.Ctx(types.NodeID(m.cfg.NodeID), types.UserID(""), "enable", "cli")
	if err := m.driver.UpdatePluginStatus(ctx, ac, record.PluginID, types.PluginStatusEnabled); err != nil {
		return fmt.Errorf("updating plugin status for %q: %w", name, err)
	}

	// Load the plugin into the runtime.
	pluginDir := filepath.Join(m.cfg.Directory, name)
	initPath := filepath.Join(pluginDir, "init.lua")

	info, err := ExtractManifest(initPath)
	if err != nil {
		return fmt.Errorf("extracting manifest for %q: %w", name, err)
	}

	inst := &PluginInstance{
		Info:        *info,
		State:       StateDiscovered,
		Dir:         pluginDir,
		InitPath:    initPath,
		dbAPIs:      make(map[*lua.LState]*DatabaseAPI),
		requestAPIs: make(map[*lua.LState]*requestAPIState),
	}

	m.mu.Lock()
	m.plugins[info.Name] = inst
	m.loadPlugin(ctx, inst)
	m.mu.Unlock()

	// Reload the pipeline registry to pick up the newly enabled pipelines.
	if regErr := m.LoadRegistry(ctx); regErr != nil {
		utility.DefaultLogger.Warn(
			fmt.Sprintf("failed to reload pipeline registry after enabling %q: %s", name, regErr.Error()),
			nil,
		)
	}

	if inst.State == StateFailed {
		return fmt.Errorf("plugin %q enabled in DB but failed to load: %s", name, inst.FailedReason)
	}

	utility.DefaultLogger.Info(
		fmt.Sprintf("plugin %q enabled and loaded (version: %s)", name, info.Version),
	)

	return nil
}

// DisablePlugin updates the database status to "installed" and unloads the plugin
// from the runtime. The plugin's pipelines are deactivated.
func (m *Manager) DisablePlugin(ctx context.Context, name string) error {
	if m.driver == nil {
		return fmt.Errorf("plugin lifecycle requires a database driver")
	}

	record, err := m.driver.GetPluginByName(name)
	if err != nil {
		return fmt.Errorf("plugin %q not found in database: %w", name, err)
	}

	if record.Status == types.PluginStatusInstalled {
		return fmt.Errorf("plugin %q is already disabled (status: installed)", name)
	}

	// Update DB status to installed.
	ac := audited.Ctx(types.NodeID(m.cfg.NodeID), types.UserID(""), "disable", "cli")
	if err := m.driver.UpdatePluginStatus(ctx, ac, record.PluginID, types.PluginStatusInstalled); err != nil {
		return fmt.Errorf("updating plugin status for %q: %w", name, err)
	}

	// Unload the plugin from the runtime.
	m.mu.Lock()
	inst, exists := m.plugins[name]
	if exists {
		if inst.State == StateRunning {
			m.shutdownPlugin(ctx, inst)
		}
		inst.State = StateStopped
		if inst.CB != nil {
			inst.CB.Trip("disabled via lifecycle")
		}
		delete(m.plugins, name)
	}
	m.mu.Unlock()

	// Reload the pipeline registry to remove the disabled plugin's pipelines.
	if regErr := m.LoadRegistry(ctx); regErr != nil {
		utility.DefaultLogger.Warn(
			fmt.Sprintf("failed to reload pipeline registry after disabling %q: %s", name, regErr.Error()),
			nil,
		)
	}

	utility.DefaultLogger.Info(
		fmt.Sprintf("plugin %q disabled and unloaded", name),
	)

	return nil
}

// ListDiscovered returns the names of filesystem plugins that are not yet
// installed in the database. Returns nil if the driver is not set.
func (m *Manager) ListDiscovered() []string {
	if m.driver == nil || m.cfg.Directory == "" {
		return nil
	}

	entries, err := os.ReadDir(m.cfg.Directory)
	if err != nil {
		return nil
	}

	var discovered []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		initPath := filepath.Join(m.cfg.Directory, entry.Name(), "init.lua")
		if _, statErr := os.Stat(initPath); statErr != nil {
			continue
		}

		// Check if this plugin is already in the database.
		existing, _ := m.driver.GetPluginByName(entry.Name())
		if existing == nil {
			discovered = append(discovered, entry.Name())
		}
	}

	return discovered
}

// ListOrphanedTables returns table names that have the plugin_ prefix but do
// not belong to any known plugin (regardless of state). System tables
// (plugin_routes, plugin_hooks) are excluded.
//
// S8: Uses known-plugin prefix matching (tablePrefix function), not delimiter parsing.
func (m *Manager) ListOrphanedTables(ctx context.Context) ([]string, error) {
	// Query all plugin_ tables from the database.
	var query string
	switch m.dialect {
	case db.DialectSQLite:
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'plugin_%'"
	case db.DialectMySQL:
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name LIKE 'plugin_%'"
	case db.DialectPostgres:
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'plugin_%'"
	default:
		return nil, fmt.Errorf("unsupported dialect for orphaned table listing: %d", m.dialect)
	}

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying plugin tables: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("ListOrphanedTables rows.Close error: %s", cerr.Error()),
				nil,
			)
		}
	}()

	// System tables that are not orphaned.
	systemTables := map[string]bool{
		"plugin_routes":   true,
		"plugin_hooks":    true,
		"plugin_requests": true,
	}

	// Build the set of known plugin prefixes from all plugins (all states).
	m.mu.RLock()
	knownPrefixes := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		knownPrefixes = append(knownPrefixes, tablePrefix(name))
	}
	m.mu.RUnlock()

	var orphaned []string
	for rows.Next() {
		var tableName string
		if scanErr := rows.Scan(&tableName); scanErr != nil {
			return nil, fmt.Errorf("scanning table name: %w", scanErr)
		}

		// Skip system tables.
		if systemTables[tableName] {
			continue
		}

		// Check if any known plugin claims this table.
		claimed := false
		for _, prefix := range knownPrefixes {
			if strings.HasPrefix(tableName, prefix) {
				claimed = true
				break
			}
		}
		if !claimed {
			orphaned = append(orphaned, tableName)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating table names: %w", err)
	}

	return orphaned, nil
}

// DropOrphanedTables safely drops the requested tables after re-verifying
// orphan status. Closes the TOCTOU window between the dry-run GET and the
// destructive POST (S1).
//
// Safety: does NOT trust the caller's requestedTables list directly. Instead:
//  1. Internally re-lists orphaned tables.
//  2. Intersects with requestedTables.
//  3. Validates each via db.ValidTableName.
//  4. Re-verifies orphan status at execution time.
func (m *Manager) DropOrphanedTables(ctx context.Context, requestedTables []string) ([]string, error) {
	// Step 1: Get current orphaned set.
	currentOrphans, err := m.ListOrphanedTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("re-listing orphaned tables: %w", err)
	}

	orphanSet := make(map[string]bool, len(currentOrphans))
	for _, name := range currentOrphans {
		orphanSet[name] = true
	}

	// Step 2: Intersect with requested tables.
	var toDrop []string
	for _, name := range requestedTables {
		if !orphanSet[name] {
			continue // not currently orphaned -- skip
		}
		// Step 3: Validate table name.
		if validErr := db.ValidTableName(name); validErr != nil {
			utility.DefaultLogger.Warn(
				fmt.Sprintf("DropOrphanedTables: invalid table name %q skipped: %s", name, validErr.Error()),
				nil,
			)
			continue
		}
		toDrop = append(toDrop, name)
	}

	// Step 4: Drop each table.
	var dropped []string
	for _, name := range toDrop {
		utility.DefaultLogger.Warn(
			fmt.Sprintf("dropping orphaned plugin table: %s", name),
			nil,
		)

		// Table names are validated via ValidTableName above. SQL parameterized
		// queries cannot bind table names, so we construct the DDL with the
		// validated identifier.
		dropQuery := "DROP TABLE IF EXISTS " + name
		if _, execErr := m.db.ExecContext(ctx, dropQuery); execErr != nil {
			utility.DefaultLogger.Error(
				fmt.Sprintf("failed to drop orphaned table %q: %s", name, execErr.Error()),
				nil,
			)
			continue
		}
		dropped = append(dropped, name)
	}

	return dropped, nil
}

// SyncCapabilities re-reads the current filesystem capabilities and updates the
// DB record and pipeline entries for the named plugin. Clears the drift indicators
// on the in-memory instance after a successful sync.
//
// Split-lock strategy: DB I/O runs without holding m.mu; in-memory state mutation
// uses a brief write lock. Steps 5-7 are not transactional (same limitation as
// InstallPlugin); LoadRegistry self-heals on the next successful call.
func (m *Manager) SyncCapabilities(ctx context.Context, name string, adminUser string) error {
	if m.driver == nil {
		return fmt.Errorf("sync capabilities requires a database driver")
	}

	// Phase A: Read in-memory state under read lock.
	m.mu.RLock()
	inst, exists := m.plugins[name]
	if !exists {
		m.mu.RUnlock()
		return fmt.Errorf("plugin %q not found in memory", name)
	}
	caps := inst.Info.Capabilities
	dir := inst.Dir
	m.mu.RUnlock()

	// Get DB record (must be installed).
	record, err := m.driver.GetPluginByName(name)
	if err != nil {
		return fmt.Errorf("plugin %q not found in database: %w", name, err)
	}

	// Compute fresh manifest hash.
	hash, err := computeChecksum(dir)
	if err != nil {
		return fmt.Errorf("computing checksum for %q: %w", name, err)
	}

	// Marshal current capabilities.
	capJSON, err := json.Marshal(caps)
	if err != nil {
		return fmt.Errorf("marshaling capabilities for %q: %w", name, err)
	}

	// Construct audit context.
	ac := audited.Ctx(types.NodeID(m.cfg.NodeID), types.UserID(adminUser), "sync-capabilities", "api")

	// Update the plugin record with new hash and capabilities.
	if updateErr := m.driver.UpdatePlugin(ctx, ac, db.UpdatePluginParams{
		PluginID:       record.PluginID,
		Version:        record.Version,
		Description:    record.Description,
		Author:         record.Author,
		Status:         record.Status,
		Capabilities:   string(capJSON),
		ApprovedAccess: record.ApprovedAccess.String(),
		ManifestHash:   hash,
	}); updateErr != nil {
		return fmt.Errorf("updating plugin record for %q: %w", name, updateErr)
	}

	// Delete old pipeline entries and create new ones.
	if delErr := m.driver.DeletePipelinesByPluginID(ctx, ac, record.PluginID); delErr != nil {
		return fmt.Errorf("deleting old pipelines for %q: %w", name, delErr)
	}

	for _, cap := range caps {
		if _, createErr := m.driver.CreatePipeline(ctx, ac, db.CreatePipelineParams{
			PluginID:   record.PluginID,
			TableName:  cap.Table,
			Operation:  cap.Op,
			PluginName: name,
			Handler:    cap.Handler,
			Priority:   cap.Priority,
			Enabled:    true,
			Config:     types.NewJSONData("{}"),
		}); createErr != nil {
			return fmt.Errorf("creating pipeline for %q (%s.%s): %w", name, cap.Table, cap.Op, createErr)
		}
	}

	// Reload the pipeline registry.
	if regErr := m.LoadRegistry(ctx); regErr != nil {
		utility.DefaultLogger.Warn(
			fmt.Sprintf("failed to reload pipeline registry after sync for %q: %s", name, regErr.Error()),
			nil,
		)
	}

	// Phase B: Clear drift indicators under write lock.
	m.mu.Lock()
	if inst, ok := m.plugins[name]; ok {
		inst.ManifestDrift = false
		inst.CapabilityDrift = nil
	}
	m.mu.Unlock()

	// Notify coordinator to prevent redundant reload on local instance.
	if m.coordinator != nil {
		refreshed, getErr := m.driver.GetPluginByName(name)
		if getErr == nil {
			m.NotifyCoordinatorSync(name, refreshed.DateModified.String())
		}
	}

	utility.DefaultLogger.Info(
		fmt.Sprintf("plugin %q capabilities synced (hash: %s, capabilities: %d)", name, hash[:8], len(caps)),
	)

	return nil
}

// NotifyCoordinatorSync updates the coordinator's lastSeen timestamp for a plugin
// to prevent redundant registry reloads after a local SyncCapabilities call.
func (m *Manager) NotifyCoordinatorSync(name string, dateModified string) {
	if m.coordinator == nil {
		return
	}
	prev, ok := m.coordinator.lastSeen[name]
	if ok {
		prev.DateModified = dateModified
		m.coordinator.lastSeen[name] = prev
	}
}

// DryRunPipeline returns the ordered pipeline entries that would execute for a
// given table and operation. Thin delegation to the registry.
func (m *Manager) DryRunPipeline(table, op string) []DryRunResult {
	return m.registry.DryRun(table, op)
}

// DryRunAllPipelines returns all pipeline chains from the in-memory registry.
func (m *Manager) DryRunAllPipelines() []DryRunResult {
	return m.registry.DryRunAll()
}

// ListAllPipelines returns all pipeline records from the database (includes disabled).
// Returns nil if the driver is not set.
func (m *Manager) ListAllPipelines() (*[]db.Pipeline, error) {
	if m.driver == nil {
		return nil, nil
	}
	return m.driver.ListPipelines()
}

// Config returns the manager configuration. Used by the watcher and other
// internal components that need read access to config values.
func (m *Manager) Config() ManagerConfig {
	return m.cfg
}

// DB returns the plugin database pool. Used by internal components (e.g., watcher)
// that need to create new instances with the same connection pool.
func (m *Manager) DB() *sql.DB {
	return m.db
}

// Dialect returns the SQL dialect in use. Used by internal components that need
// to construct dialect-specific queries.
func (m *Manager) Dialect() db.Dialect {
	return m.dialect
}
