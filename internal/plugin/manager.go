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

	return &Manager{
		cfg:     cfg,
		db:      pool,
		dialect: dialect,
		plugins: make(map[string]*PluginInstance),
	}
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

	// Create VM factory -- produces fully sandboxed VMs with db/log APIs.
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
		// 5. FreezeModule("db")
		// 6. FreezeModule("log")
		ApplySandbox(L, SandboxConfig{AllowCoroutine: true, ExecTimeout: timeout})
		RegisterPluginRequire(L, inst.Dir)

		dbAPI := NewDatabaseAPI(m.db, pluginName, m.dialect, m.cfg.MaxOpsPerExec)
		RegisterDBAPI(L, dbAPI)
		RegisterLogAPI(L, pluginName)
		FreezeModule(L, "db")
		FreezeModule(L, "log")

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

		// Store the dbAPI mapping on the instance for op count reset.
		inst.dbAPIs[L] = dbAPI

		return L
	}

	// Create the VM pool.
	pool := NewVMPool(m.cfg.MaxVMsPerPlugin, factory, inst.InitPath, pluginName)
	inst.Pool = pool

	// Check out one VM to run on_init().
	initCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	L, getErr := pool.Get(initCtx)
	if getErr != nil {
		m.failPlugin(inst, fmt.Sprintf("failed to get VM from pool: %s", getErr.Error()))
		return
	}

	// Reset op count before plugin code executes.
	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}

	// Set context with timeout for on_init execution.
	L.SetContext(initCtx)

	// Call on_init() if defined.
	onInit := L.GetGlobal("on_init")
	if fn, ok := onInit.(*lua.LFunction); ok {
		if callErr := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}); callErr != nil {
			pool.Put(L)
			m.failPlugin(inst, fmt.Sprintf("on_init failed: %s", callErr.Error()))
			return
		}
	}

	// Take global snapshot after on_init (captures any globals defined during init).
	pool.SnapshotGlobals(L)

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
	if dbAPI, ok := inst.dbAPIs[L]; ok {
		dbAPI.ResetOpCount()
	}

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

// sortStringSlice sorts a string slice in place for deterministic ordering.
func sortStringSlice(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
