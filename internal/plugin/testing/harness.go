package testing

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/plugin"
	"github.com/hegner123/modulacms/internal/utility"
	lua "github.com/yuin/gopher-lua"

	_ "github.com/mattn/go-sqlite3"
)

// HarnessOpts configures the test harness behavior.
type HarnessOpts struct {
	Verbose bool
	Timeout time.Duration // per-test timeout (default 5s if zero)
	Filter  string        // run only test functions containing this substring
}

// Harness is the core orchestrator for plugin test execution.
type Harness struct {
	pluginDir  string
	pluginName string
	conn       *sql.DB
	mgr        *plugin.Manager
	bridge     *plugin.HTTPBridge
	mux        *http.ServeMux
	mockReq    *MockRequestEngine
	opts       HarnessOpts
}

// NewHarness creates an isolated test environment for a single plugin.
// The plugin is loaded into an in-memory SQLite database with all CMS tables
// bootstrapped. All routes, hooks, and request domains are auto-approved.
func NewHarness(pluginDir string, opts HarnessOpts) (*Harness, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}

	ctx := context.Background()

	// Suppress plugin loading logs during test runs.
	utility.DefaultLogger.SetLevel(utility.ERROR)

	// Step 1: Validate plugin
	info, results, err := plugin.ValidatePlugin(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("validate plugin: %w", err)
	}
	for _, r := range results {
		if r.Severity == plugin.SeverityError {
			return nil, fmt.Errorf("plugin validation error: %s: %s", r.Field, r.Message)
		}
	}

	// Step 2: Open shared in-memory SQLite
	conn, err := sql.Open("sqlite3", "file:plugintest?mode=memory&cache=shared")
	if err != nil {
		return nil, fmt.Errorf("open in-memory db: %w", err)
	}
	// Keep at least one connection open so the in-memory DB persists.
	conn.SetMaxIdleConns(2)
	conn.SetConnMaxLifetime(0)

	// Step 3: Bootstrap CMS schema
	d := db.Database{
		Connection: conn,
		Context:    ctx,
		Config:     config.Config{Node_ID: types.NewNodeID().String()},
	}
	if err := d.CreateAllTables(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("bootstrap tables: %w", err)
	}
	placeholderHash := "$argon2id$v=19$m=65536,t=1,p=4$dGVzdA$dGVzdA"
	if err := d.CreateBootstrapData(placeholderHash); err != nil {
		conn.Close()
		return nil, fmt.Errorf("bootstrap data: %w", err)
	}

	// Step 4: Create temp directory, symlink plugin into it
	tmpDir, err := os.MkdirTemp("", "plugintest-*")
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	absPluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		conn.Close()
		return nil, fmt.Errorf("abs plugin dir: %w", err)
	}

	linkPath := filepath.Join(tmpDir, info.Name)
	if err := os.Symlink(absPluginDir, linkPath); err != nil {
		os.RemoveAll(tmpDir)
		conn.Close()
		return nil, fmt.Errorf("symlink plugin: %w", err)
	}

	// Step 5: Create Manager
	mockReq := &MockRequestEngine{}
	mgr := plugin.NewManager(plugin.ManagerConfig{
		Enabled:   true,
		Directory: tmpDir,
	}, conn, db.DialectSQLite, nil, nil)

	mgr.SetOutboundOverride(mockReq)

	// Step 6: Create HTTPBridge and hook/request infrastructure tables
	bridge := plugin.NewHTTPBridge(mgr, conn, db.DialectSQLite)
	mgr.SetBridge(bridge)

	if err := bridge.CreatePluginRoutesTable(ctx); err != nil {
		cleanupHarness(mgr, conn, tmpDir)
		return nil, fmt.Errorf("create plugin_routes table: %w", err)
	}
	if err := mgr.HookEngine().CreatePluginHooksTable(ctx); err != nil {
		cleanupHarness(mgr, conn, tmpDir)
		return nil, fmt.Errorf("create plugin_hooks table: %w", err)
	}
	if err := mgr.RequestEngine().CreatePluginRequestsTable(ctx); err != nil {
		cleanupHarness(mgr, conn, tmpDir)
		return nil, fmt.Errorf("create plugin_requests table: %w", err)
	}

	// Step 7: LoadAll (filesystem-only discovery, driver=nil)
	if err := mgr.LoadAll(ctx); err != nil {
		cleanupHarness(mgr, conn, tmpDir)
		return nil, fmt.Errorf("load plugin: %w", err)
	}

	// Verify the plugin loaded successfully.
	loadedInst := mgr.GetPlugin(info.Name)
	if loadedInst == nil {
		// Check all loaded plugins for diagnostics.
		var names []string
		for _, p := range mgr.ListPlugins() {
			names = append(names, fmt.Sprintf("%s(%s)", p.Info.Name, p.State))
		}
		cleanupHarness(mgr, conn, tmpDir)
		return nil, fmt.Errorf("plugin %q not found after LoadAll; loaded: %v", info.Name, names)
	}
	if loadedInst.State == plugin.StateFailed {
		cleanupHarness(mgr, conn, tmpDir)
		return nil, fmt.Errorf("plugin %q failed to load: %s", info.Name, loadedInst.FailedReason)
	}

	// Step 8: Auto-approve all registered routes, hooks, and request domains
	for _, route := range bridge.ListRoutes() {
		if err := bridge.ApproveRoute(ctx, route.PluginName, route.Method, route.Path, "test-harness"); err != nil {
			cleanupHarness(mgr, conn, tmpDir)
			return nil, fmt.Errorf("approve route %s %s: %w", route.Method, route.Path, err)
		}
	}
	for _, hook := range mgr.HookEngine().ListHooks() {
		if err := mgr.HookEngine().ApproveHook(ctx, hook.PluginName, hook.Event, hook.Table, "test-harness"); err != nil {
			cleanupHarness(mgr, conn, tmpDir)
			return nil, fmt.Errorf("approve hook %s/%s: %w", hook.Event, hook.Table, err)
		}
	}
	reqs, err := mgr.RequestEngine().ListRequests(ctx)
	if err != nil {
		cleanupHarness(mgr, conn, tmpDir)
		return nil, fmt.Errorf("list requests: %w", err)
	}
	for _, req := range reqs {
		if err := mgr.RequestEngine().ApproveRequest(ctx, req.PluginName, req.Domain, "test-harness"); err != nil {
			cleanupHarness(mgr, conn, tmpDir)
			return nil, fmt.Errorf("approve request domain %s: %w", req.Domain, err)
		}
	}

	// Step 9: Populate ApprovedAccess from manifest core_access
	inst := mgr.GetPlugin(info.Name)
	if inst != nil && info.CoreAccess != nil {
		if inst.ApprovedAccess == nil {
			inst.ApprovedAccess = make(plugin.PluginCoreAccess)
		}
		for table, perms := range info.CoreAccess {
			inst.ApprovedAccess[table] = perms
		}
	}

	// Step 10: Mount bridge on mux
	mux := http.NewServeMux()
	bridge.MountOn(mux)

	h := &Harness{
		pluginDir:  pluginDir,
		pluginName: info.Name,
		conn:       conn,
		mgr:        mgr,
		bridge:     bridge,
		mux:        mux,
		mockReq:    mockReq,
		opts:       opts,
	}

	return h, nil
}

// DiscoverTests finds all test/*.test.lua files in the plugin directory.
func (h *Harness) DiscoverTests() ([]string, error) {
	testDir := filepath.Join(h.pluginDir, "test")
	entries, err := os.ReadDir(testDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no test/ directory found in %s", h.pluginDir)
		}
		return nil, fmt.Errorf("read test dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".test.lua") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		return nil, fmt.Errorf("no *.test.lua files found in %s/test/", h.pluginDir)
	}
	return files, nil
}

// RunAll executes all test functions in the given files and returns a report.
func (h *Harness) RunAll(ctx context.Context, files []string) *Report {
	start := time.Now()
	report := &Report{
		PluginName: h.pluginName,
	}

	for _, file := range files {
		results := h.runFile(ctx, file)
		report.Results = append(report.Results, results...)
	}

	report.TotalTime = time.Since(start)
	return report
}

// runFile executes all test_* functions in a single test file.
func (h *Harness) runFile(ctx context.Context, filename string) []TestResult {
	testPath := filepath.Join(h.pluginDir, "test", filename)

	// Create a fresh VM for this file
	L := lua.NewState(lua.Options{
		SkipOpenLibs: false,
	})
	defer L.Close()

	// Set up the VM with the production sandbox environment
	timeout := time.Duration(5) * time.Second
	plugin.ApplySandbox(L, plugin.SandboxConfig{AllowCoroutine: true, ExecTimeout: timeout})
	plugin.SetVMPhase(L, "runtime")

	// Register plugin's lib/ for require()
	plugin.RegisterPluginRequire(L, h.pluginDir)

	// Register the production APIs (db, log, http, hooks, core, request, json)
	inst := h.mgr.GetPlugin(h.pluginName)
	if inst == nil {
		return []TestResult{{
			File:   filename,
			Test:   "(load)",
			Passed: false,
			Failures: []Failure{{
				Message: fmt.Sprintf("plugin %q not found in manager", h.pluginName),
				Line:    0,
				IsError: true,
			}},
		}}
	}

	dbAPI := plugin.NewDatabaseAPI(h.conn, h.pluginName, db.DialectSQLite, 1000, nil)
	plugin.RegisterDBAPI(L, dbAPI)
	plugin.RegisterLogAPI(L, h.pluginName)
	plugin.RegisterHTTPAPI(L, h.pluginName)
	plugin.RegisterHooksAPI(L, h.pluginName)
	coreAPI := plugin.NewCoreTableAPI(dbAPI, h.pluginName, db.DialectSQLite, inst.ApprovedAccess)
	plugin.RegisterCoreAPI(L, coreAPI)
	plugin.RegisterRequestAPI(L, h.pluginName, h.mockReq)
	plugin.RegisterJSONAPI(L)

	// Freeze production modules
	for _, mod := range []string{"db", "log", "http", "hooks", "core", "request", "json"} {
		plugin.FreezeModule(L, mod)
	}

	// Register the test module
	testState := registerTestModule(L, h)

	// Load the test file
	if err := L.DoFile(testPath); err != nil {
		return []TestResult{{
			File:   filename,
			Test:   "(load)",
			Passed: false,
			Failures: []Failure{{
				Message: fmt.Sprintf("failed to load test file: %s", err.Error()),
				Line:    0,
				IsError: true,
			}},
		}}
	}

	// Discover test_* functions from globals
	var testNames []string
	L.G.Global.ForEach(func(k, v lua.LValue) {
		name := k.String()
		if strings.HasPrefix(name, "test_") {
			if _, ok := v.(*lua.LFunction); ok {
				if h.opts.Filter == "" || strings.Contains(name, h.opts.Filter) {
					testNames = append(testNames, name)
				}
			}
		}
	})
	sort.Strings(testNames)

	var results []TestResult
	for i, name := range testNames {
		result := h.runTest(ctx, L, testState, filename, name, i)
		results = append(results, result)
	}

	return results
}

// runTest executes a single test_* function within a SAVEPOINT.
func (h *Harness) runTest(ctx context.Context, L *lua.LState, testState *testAPIState, filename, testName string, index int) TestResult {
	testState.reset()
	h.mockReq.ClearRules()
	start := time.Now()

	spName := fmt.Sprintf("sp_%d", index)

	// Create savepoint
	if _, err := h.conn.ExecContext(ctx, "SAVEPOINT "+spName); err != nil {
		return TestResult{
			File:   filename,
			Test:   testName,
			Passed: false,
			Failures: []Failure{{
				Message: fmt.Sprintf("SAVEPOINT failed: %s", err.Error()),
				IsError: true,
			}},
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	// Ensure rollback happens regardless of outcome
	defer func() {
		h.conn.ExecContext(ctx, "ROLLBACK TO "+spName)
		h.conn.ExecContext(ctx, "RELEASE "+spName)
	}()

	// Run setup if registered
	if testState.setupFn != nil {
		if err := L.CallByParam(lua.P{
			Fn:      testState.setupFn,
			NRet:    0,
			Protect: true,
		}); err != nil {
			return TestResult{
				File:   filename,
				Test:   testName,
				Passed: false,
				Failures: []Failure{{
					Message: fmt.Sprintf("setup error: %s", err.Error()),
					IsError: true,
				}},
				Assertions: testState.assertions,
				DurationMs: time.Since(start).Milliseconds(),
			}
		}
	}

	// Run test function with timeout
	execCtx, cancel := context.WithTimeout(ctx, h.opts.Timeout)
	defer cancel()
	L.SetContext(execCtx)

	fn := L.GetGlobal(testName)
	luaFn, ok := fn.(*lua.LFunction)
	if !ok {
		return TestResult{
			File:   filename,
			Test:   testName,
			Passed: false,
			Failures: []Failure{{
				Message: fmt.Sprintf("%s is not a function", testName),
				IsError: true,
			}},
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	if err := L.CallByParam(lua.P{
		Fn:      luaFn,
		NRet:    0,
		Protect: true,
	}); err != nil {
		testState.failures = append(testState.failures, Failure{
			Message: err.Error(),
			IsError: true,
		})
	}

	// Reset context so teardown and subsequent operations work
	L.SetContext(ctx)

	// Run teardown if registered
	if testState.teardownFn != nil {
		if err := L.CallByParam(lua.P{
			Fn:      testState.teardownFn,
			NRet:    0,
			Protect: true,
		}); err != nil {
			testState.failures = append(testState.failures, Failure{
				Message: fmt.Sprintf("teardown error: %s", err.Error()),
				IsError: true,
			})
		}
	}

	passed := len(testState.failures) == 0

	return TestResult{
		File:       filename,
		Test:       testName,
		Passed:     passed,
		DurationMs: time.Since(start).Milliseconds(),
		Assertions: testState.assertions,
		Failures:   testState.failures,
	}
}

// Close shuts down the harness in the correct order.
func (h *Harness) Close() {
	ctx := context.Background()

	// 1. Shutdown manager (drains VM pools, closes plugin DB pool)
	h.mgr.Shutdown(ctx)

	// 2. Close RequestEngine background goroutine
	h.mgr.RequestEngine().Close()

	// 3. Close the bridge's background goroutine
	h.bridge.Close(ctx)

	// 4. Close in-memory SQLite
	h.conn.Close()
}

// cleanupHarness handles cleanup when NewHarness fails partway through.
func cleanupHarness(mgr *plugin.Manager, conn *sql.DB, tmpDir string) {
	if mgr != nil {
		ctx := context.Background()
		mgr.Shutdown(ctx)
		mgr.RequestEngine().Close()
	}
	if conn != nil {
		conn.Close()
	}
	if tmpDir != "" {
		os.RemoveAll(tmpDir)
	}
}
