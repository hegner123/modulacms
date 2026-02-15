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

	db "github.com/hegner123/modulacms/internal/db"
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

// PluginInfo holds the metadata extracted from a plugin's plugin_info global.
type PluginInfo struct {
	Name          string
	Version       string
	Description   string
	Author        string
	License       string
	MinCMSVersion string
	Dependencies  []string
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

	// mu protects dbAPIs from concurrent access. Required because the onReplace
	// callback (called from Put on a VM-returning goroutine) deletes entries while
	// executeBefore/executeAfter read entries on different goroutines (S5).
	mu sync.Mutex

	// dbAPIs maps each VM (*lua.LState) to its bound DatabaseAPI for op count reset.
	// Each DatabaseAPI is bound to exactly one LState (1:1 invariant).
	dbAPIs map[*lua.LState]*DatabaseAPI
}

// ManagerConfig configures the plugin manager runtime behavior.
type ManagerConfig struct {
	Enabled         bool
	Directory       string
	MaxVMsPerPlugin int // default 4
	ExecTimeoutSec  int // default 5
	MaxOpsPerExec   int // default 1000, per VM checkout

	// Hook engine configuration (Phase 3).
	HookReserveVMs          int // VMs reserved for hook execution; default 1
	HookMaxConsecutiveAborts int // circuit breaker threshold; default 10
	HookMaxOps              int // reduced op budget for after-hooks; default 100
	HookMaxConcurrentAfter  int // max concurrent after-hook goroutines; default 10
	HookTimeoutMs           int // per-hook timeout in before-hooks (ms); default 2000
	HookEventTimeoutMs      int // per-event total timeout for before-hook chain (ms); default 5000

	// Phase 4: Production hardening configuration.
	HotReload     bool          // default false (zero value) -- production opt-in only (S10)
	MaxFailures   int           // circuit breaker threshold; default 5
	ResetInterval time.Duration // circuit breaker reset interval; default 60s
}

// Manager is the central coordinator for plugin discovery, loading, lifecycle, and shutdown.
//
// Context handling: Manager does not store a context.Context field. Stored contexts on
// long-lived structs create ambiguity about which context governs a given operation.
// Instead, all Manager methods that perform I/O accept ctx context.Context as their
// first parameter. The caller (cmd/serve.go) passes the application lifecycle context.
type Manager struct {
	cfg     ManagerConfig
	db      *sql.DB // separate plugin pool via db.OpenPool()
	dialect db.Dialect
	plugins map[string]*PluginInstance
	mu      sync.RWMutex

	// loadOrder preserves the topologically sorted order of successfully loaded plugins.
	// Used for reverse-order shutdown.
	loadOrder []string

	// bridge is the HTTP bridge for plugin route registration. Set via SetBridge()
	// before LoadAll(). May be nil if HTTP integration is not enabled.
	bridge *HTTPBridge

	// hookEngine is the content lifecycle hook engine (Phase 3). Created during
	// NewManager, always non-nil when plugins are enabled. Implements audited.HookRunner.
	hookEngine *HookEngine

	// watcher is the file-polling hot reload watcher (Phase 4). Nil when hot
	// reload is disabled. Started via StartWatcher() after LoadAll().
	watcher      *Watcher
	watcherCancel context.CancelFunc

	// shutdownOnce ensures Shutdown is idempotent -- calling it multiple times
	// (e.g., signal handler + deferred call) does not double-close resources.
	shutdownOnce sync.Once
}

// NewManager creates a new plugin Manager with the given configuration.
// Zero-value config fields are replaced with defaults.
// The db pool must be a separate *sql.DB opened via db.OpenPool() for isolation.
func NewManager(cfg ManagerConfig, pool *sql.DB, dialect db.Dialect) *Manager {
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
		cfg:     cfg,
		db:      pool,
		dialect: dialect,
		plugins: make(map[string]*PluginInstance),
	}

	// Create hook engine unconditionally. The engine is inert when no hooks
	// are registered (HasHooks returns false via hasAnyHook fast-path).
	mgr.hookEngine = NewHookEngine(mgr, pool, dialect, HookEngineConfig{
		HookTimeoutMs:        cfg.HookTimeoutMs,
		EventTimeoutMs:       cfg.HookEventTimeoutMs,
		MaxConsecutiveAborts: cfg.HookMaxConsecutiveAborts,
		MaxConcurrentAfter:  cfg.HookMaxConcurrentAfter,
		HookMaxOps:          cfg.HookMaxOps,
		ExecTimeoutMs:       cfg.ExecTimeoutSec * 1000, // reuse exec timeout for after-hooks
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
		info, extractErr := extractManifest(initPath)
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
			Info:     *info,
			State:    StateDiscovered,
			Dir:      pluginDir,
			InitPath: initPath,
			dbAPIs:   make(map[*lua.LState]*DatabaseAPI),
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

		discoveredNames := make([]string, 0, len(discovered))
		for name := range discovered {
			discoveredNames = append(discoveredNames, name)
		}
		if cleanErr := m.hookEngine.CleanupOrphanedHooks(ctx, discoveredNames); cleanErr != nil {
			return fmt.Errorf("cleaning orphaned hooks: %w", cleanErr)
		}
	}

	// Step 4: Load each plugin in dependency order.
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, inst := range sorted {
		m.plugins[inst.Info.Name] = inst
		m.loadPlugin(ctx, inst)
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
		// 2. RegisterPluginRequire
		// 3. RegisterDBAPI
		// 4. RegisterLogAPI
		// 5. RegisterHTTPAPI (Phase 2)
		// 6. RegisterHooksAPI (Phase 3)
		// 7. FreezeModule("db")
		// 8. FreezeModule("log")
		// 9. FreezeModule("http")
		// 10. FreezeModule("hooks")
		ApplySandbox(L, SandboxConfig{AllowCoroutine: true, ExecTimeout: timeout})
		RegisterPluginRequire(L, inst.Dir)

		dbAPI := NewDatabaseAPI(m.db, pluginName, m.dialect, m.cfg.MaxOpsPerExec)
		RegisterDBAPI(L, dbAPI)
		RegisterLogAPI(L, pluginName)
		RegisterHTTPAPI(L, pluginName)
		RegisterHooksAPI(L, pluginName)
		FreezeModule(L, "db")
		FreezeModule(L, "log")
		FreezeModule(L, "http")
		FreezeModule(L, "hooks")

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

		// Store the dbAPI mapping on the instance for op count reset (S5: protected by inst.mu).
		inst.mu.Lock()
		inst.dbAPIs[L] = dbAPI
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

	// Set phase flag so http.handle() rejects calls during on_init.
	// http.handle() must only be called at module scope (during init.lua execution
	// by the factory), not inside on_init(). The flag is stored in the LState's
	// registry table, which is inaccessible from Lua code.
	registryTbl := L.Get(lua.RegistryIndex)
	if regTbl, ok := registryTbl.(*lua.LTable); ok {
		L.SetField(regTbl, "in_init", lua.LTrue)
	}

	// Call on_init() if defined.
	onInit := L.GetGlobal("on_init")
	if fn, ok := onInit.(*lua.LFunction); ok {
		if callErr := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}); callErr != nil {
			// Clear phase flag before returning VM to pool.
			if regTbl, ok := registryTbl.(*lua.LTable); ok {
				L.SetField(regTbl, "in_init", lua.LNil)
			}
			pool.Put(L)
			m.failPlugin(inst, fmt.Sprintf("on_init failed: %s", callErr.Error()))
			return
		}
	}

	// Clear phase flag after on_init completes.
	if regTbl, ok := registryTbl.(*lua.LTable); ok {
		L.SetField(regTbl, "in_init", lua.LNil)
	}

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

// extractManifest creates a temporary sandboxed VM (no db/log APIs), executes
// init.lua, reads the plugin_info global, and validates the manifest fields.
// The temp VM is discarded after extraction.
func extractManifest(initPath string) (*PluginInfo, error) {
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

	// Validate manifest.
	if err := validateManifest(info); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return info, nil
}

// validateManifest checks that all required fields are present and valid.
func validateManifest(info *PluginInfo) error {
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
// side effects, allowing extractManifest to read plugin_info even when the
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

	for _, name := range []string{"db", "log", "http", "hooks"} {
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
	newInfo, extractErr := extractManifest(oldInst.InitPath)
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
		Info:     *newInfo,
		State:    StateDiscovered,
		Dir:      oldInst.Dir,
		InitPath: oldInst.InitPath,
		dbAPIs:   make(map[*lua.LState]*DatabaseAPI),
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

// DisablePlugin stops a plugin and trips its circuit breaker.
// The plugin can be re-enabled via EnablePlugin.
func (m *Manager) DisablePlugin(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %q not found", name)
	}

	if inst.State == StateStopped {
		return fmt.Errorf("plugin %q is already stopped", name)
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

	return nil
}

// EnablePlugin resets the circuit breaker and reloads the plugin.
// S5: Emits slog.Warn audit event with admin user, plugin name, prior CB state,
// and failure count before reset.
func (m *Manager) EnablePlugin(ctx context.Context, name string, adminUser string) error {
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
	return m.ReloadPlugin(ctx, name)
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
		"plugin_routes": true,
		"plugin_hooks":  true,
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
