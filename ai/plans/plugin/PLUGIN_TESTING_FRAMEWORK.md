# Plugin Test Framework Plan

## Context

Plugin developers have no way to test their Lua plugins without starting the full CMS server. The internal Go test suite has comprehensive helpers (`newTestDB`, `writePluginFile`, `loadPluginForTest`) but these are unexported. The `plugin validate` command only checks syntax and manifest. A test framework would let the x402 and forms plugins (and all future plugins) run automated tests in an isolated in-memory environment.

## Design

A `modulacms plugin test <path>` CLI command that:

1. Loads the plugin into an isolated in-memory SQLite environment
2. Discovers `test/*.test.lua` files
3. Runs all `test_*` functions with assertion support
4. Reports pass/fail per function
5. Exit code: 0 all pass, 1 failures, 2 load error

## Lua `test` Module API

Available only in test files, not production plugin code.

| Function | Signature | Behavior |
|----------|-----------|----------|
| `test.assert` | `(condition, message?)` | Soft fail if falsy |
| `test.assert_eq` | `(expected, actual, message?)` | Deep equality |
| `test.assert_neq` | `(a, b, message?)` | Fail if equal |
| `test.assert_nil` | `(value, message?)` | Fail if not nil |
| `test.assert_not_nil` | `(value, message?)` | Fail if nil |
| `test.assert_error` | `(fn, pattern?)` | Assert `fn()` raises error; if pattern given, use Lua `string.find` (not regex) |
| `test.assert_contains` | `(haystack, needle, message?)` | String containment |
| `test.request` | `(method, path, opts?)` | HTTP request through plugin routes, returns `{status, headers, json, body}` |
| `test.fire_hook` | `(event, table, data, state?)` | Fire a hook; return shape varies by category (see [Hook testing](#hook-testing)) |
| `test.mock_request` | `(method, url_pattern, response)` | Register mock for outbound `request.*` calls |
| `test.setup` | `(fn)` | Register per-test setup (runs inside SAVEPOINT, before `test_*` fn) |
| `test.teardown` | `(fn)` | Register per-test teardown (runs inside SAVEPOINT, after `test_*` fn) |

Assertions are soft failures -- they record the failure and continue executing the test function (not `error()` which would halt). This gives a full picture of all failures per test, not just the first one.

## Test file convention

```
plugins/x402/
  init.lua
  lib/...
  test/
    rules.test.lua
    payment.test.lua
```

Test functions: all globals starting with `test_`. Discovered via Lua global table scan, sorted alphabetically.

## New files

| File | Purpose |
|------|---------|
| `internal/plugin/testing/harness.go` | Core orchestration: load plugin, discover tests, run them |
| `internal/plugin/testing/lua_test_api.go` | `test` Lua module implementation |
| `internal/plugin/testing/mock_request.go` | Mock `OutboundExecutor` for outbound HTTP |
| `internal/plugin/testing/results.go` | Report types and output formatting |
| `cmd/plugin_test_cmd.go` | Cobra subcommand |

## Modified files

| File | Change |
|------|--------|
| `internal/plugin/request_api.go` | Export `outboundExecutor` -> `OutboundExecutor`, `setVMPhase` -> `SetVMPhase` |
| `internal/plugin/manager.go` | Add `outboundOverride` field + setter, use it in VM factory only |
| `cmd/plugin.go` | Register `pluginTestCmd` in `init()` |
| `internal/middleware/authorization.go` | Add exported `SetPermissions` and `SetIsAdmin` context setters |
| All internal callers of `setVMPhase`/`vmPhase`/`outboundExecutor` | Update to exported names |

## Architecture

### Harness

```go
// internal/plugin/testing/harness.go
type Harness struct {
    pluginDir  string
    pluginName string
    db         *sql.DB
    mgr        *plugin.Manager
    bridge     *plugin.HTTPBridge
    hookEngine *plugin.HookEngine
    mux        *http.ServeMux
    mockReq    *MockRequestEngine
}

type HarnessOpts struct {
    Verbose bool          // print all assertions, not just failures
    Timeout time.Duration // per-test timeout (default 5s if zero)
}

func NewHarness(pluginDir string, opts HarnessOpts) (*Harness, error)
func (h *Harness) DiscoverTests() ([]string, error)
func (h *Harness) RunAll(ctx context.Context, files []string) *Report
func (h *Harness) Close()
```

`Close()` must shut down in this order:
1. `mgr.Shutdown(ctx)` -- shuts down plugin instances, drains VM pools, closes the plugin DB pool. Note: `Manager.Shutdown` does NOT close the `RequestEngine`.
2. `mgr.RequestEngine().Close()` -- stops the `RequestEngine`'s background `cleanupRateLimiters` goroutine by closing its `cleanupDone` channel and closing idle HTTP connections. This call is required because `Manager.Shutdown` does not do it.
3. Close the in-memory SQLite connection.

Failing to call `RequestEngine.Close()` leaks a goroutine per test run (the 5-minute cleanup ticker). `bridge.Close(ctx)` is handled by `Manager.Shutdown` which calls `shutdownPlugin` on each plugin.

#### NewHarness steps

1. Validate plugin via `ValidatePlugin`
2. Open shared in-memory SQLite (`file:plugintest?mode=memory&cache=shared`)
3. Bootstrap the CMS schema on the in-memory DB. Construct a `db.Database` wrapper and call `CreateAllTables()` then `CreateBootstrapData(adminHash)`:

   ```go
   d := db.Database{
       Connection: conn,          // the shared in-memory *sql.DB from step 2
       Context:    ctx,
       Config:     config.Config{Node_ID: types.NewNodeID().String()},
   }
   if err := d.CreateAllTables(); err != nil { return nil, fmt.Errorf("bootstrap tables: %w", err) }
   if err := d.CreateBootstrapData(placeholderHash); err != nil { return nil, fmt.Errorf("bootstrap data: %w", err) }
   ```

   `Config.Node_ID` is required because `CreateBootstrapData` passes it to `audited.Ctx()` for change event recording. Use a freshly generated ULID via `types.NewNodeID()`. Use a placeholder argon2 hash for the admin user (e.g., `"$argon2id$v=19$m=65536,t=1,p=4$test$test"`) since no real authentication occurs in the test harness.

   This creates all 29 core CMS tables and inserts minimal seed data (RBAC roles/permissions, system user, default route, `page` datatype, `_reference` system datatype, field types). Required because:
   - `RunBeforeReadHooks` allows `db.query()` and `core.query()` calls that access CMS tables
   - `CoreTableAPI` (`core.*` Lua module) queries whitelisted CMS tables directly
   - Plugin route handlers may read/write content via the `db` or `core` APIs
   - Without these tables, any hook or route test touching CMS data will fail with "no such table"
4. Create temp directory, symlink plugin into it (so Manager discovers only this plugin)
5. Create Manager with `driver=nil` and `tableReg=nil` (filesystem-only discovery). The hookEngine and requestEngine are created internally by `NewManager` -- access them via `mgr.HookEngine()` and `mgr.RequestEngine()`. The HTTPBridge is NOT created by the Manager. Create it externally via `NewHTTPBridge(mgr, db, db.DialectSQLite)` and call `mgr.SetBridge(bridge)`. Set `outboundOverride` to `MockRequestEngine` on the Manager.
6. Create all plugin infrastructure tables on the in-memory DB. These three tables MUST be created before calling `LoadAll`, because the plugin loading process writes route/hook/request registrations into them. Plugin-defined tables (via `db.create_table` in `on_init`) are created automatically by `LoadAll` and do not need manual setup.
   - `bridge.CreatePluginRoutesTable(ctx)`
   - `mgr.HookEngine().CreatePluginHooksTable(ctx)`
   - `mgr.RequestEngine().CreatePluginRequestsTable(ctx)`
7. `LoadAll` with `driver=nil` follows the filesystem-only discovery path: it skips the plugins DB read and discovers plugins from the directory. Plugin `on_init` runs, plugin-defined tables are created, routes/hooks/requests are registered in memory.
8. Auto-approve all registered routes, hooks, and request domains:
   - For each route in `bridge.ListRoutes()`: call `bridge.ApproveRoute(ctx, route.PluginName, route.Method, route.Path, "test-harness")`
   - For each hook in `mgr.HookEngine().ListHooks()`: call `mgr.HookEngine().ApproveHook(ctx, hook.PluginName, hook.Event, hook.Table, "test-harness")`
   - For each req in `mgr.RequestEngine().ListRequests(ctx)`: call `mgr.RequestEngine().ApproveRequest(ctx, req.PluginName, req.Domain, "test-harness")`
   - Note: `ListRequests(ctx)` takes context and returns `([]RequestRegistration, error)`, unlike `ListRoutes`/`ListHooks` which take no context and return no error.
   - If any approval call fails, return the error from `NewHarness` (do not silently continue).
9. Populate `ApprovedAccess` on the `PluginInstance` from the manifest's `core_access` field. This enables `CoreTableAPI` in test VMs. For each declared `core_access` entry, set `inst.ApprovedAccess[table] = permissions` (e.g., `content_data = {"read"}`).
10. Mount bridge on the harness `http.ServeMux` via `bridge.MountOn(mux)`. This registers all approved routes on the mux so that Go's `ServeMux` populates `r.Pattern` on incoming requests. The bridge's `ServeHTTP` looks up routes via `r.Pattern` (not the URL path), so all test requests MUST be dispatched through `h.mux.ServeHTTP(recorder, req)`, never through `h.bridge.ServeHTTP(recorder, req)` directly. Dispatching directly to the bridge skips the ServeMux pattern matching, leaving `r.Pattern` empty, and every route lookup fails with 404.

### Test VM

Each test file gets its own Lua VM with:

- Full production sandbox (`db`, `log`, `http`, `hooks`, `core`, `request`, `json`) -- frozen
- `CoreTableAPI` wired with the plugin's `ApprovedAccess` (populated from manifest `core_access`)
- Plugin's `lib/` modules available via `require()`
- `test` module -- frozen, test-only
- VM phase set to `"runtime"` so `db.*` works
- Same in-memory SQLite as the plugin's `on_init` tables

### Test isolation within a file

Each `test_*` function runs inside a SQLite SAVEPOINT. After the function completes (pass or fail), the harness rolls back to the savepoint. This ensures:

- Tests do not depend on execution order
- A failed test cannot leave dirty state for subsequent tests
- `on_init` seed data is always the baseline

Implementation: before calling each test function, execute `SAVEPOINT sp_N` where N is a 0-based integer incremented per test function within the file (reset to 0 for each file). Do not use the function name in the savepoint name (it may contain SQL-unsafe characters). After the function returns (or errors), execute `ROLLBACK TO sp_N` then `RELEASE sp_N`.

`test.setup(fn)` and `test.teardown(fn)` register functions that run inside the SAVEPOINT bracket for each `test_*` function. Execution order: `SAVEPOINT` -> `setup()` -> `test_*()` -> `teardown()` -> `ROLLBACK`. Because the SAVEPOINT rollback undoes all DB changes, teardown is primarily useful for non-DB cleanup (resetting Lua module state, clearing mock rules, etc.).

### Lua runtime error handling

Before calling each `test_*` function, create a timeout context and set it on the VM:

```go
execCtx, cancel := context.WithTimeout(ctx, opts.Timeout) // default 5s
defer cancel()
L.SetContext(execCtx)
```

This enforces wall-clock timeout at the Lua VM level, matching production behavior (see `http_bridge.go:641` and `pool.go:136`). Without `L.SetContext`, an infinite loop in Lua will hang the test run indefinitely because `L.CallByParam` with `Protect: true` only catches `error()` calls, not CPU-bound loops.

Each `test_*` function is called via `L.CallByParam` with `Protect: true`. If the Lua code raises a runtime error (e.g., indexing nil after a soft assertion failure, or a context deadline exceeded from timeout), the error is caught by the Go side, recorded as an additional failure on the test result with the error message and stack trace, and execution proceeds to the next test function. The harness never panics from Lua errors.

### Mock outbound HTTP

```go
// internal/plugin/testing/mock_request.go
type MockRequestEngine struct {
    rules []MockRule
}

func (m *MockRequestEngine) Execute(ctx context.Context, pluginName, method, urlStr string, opts plugin.OutboundRequestOpts) (map[string]any, error)
func (m *MockRequestEngine) AddRule(method, urlPattern string, response map[string]any)
func (m *MockRequestEngine) ClearRules()
```

**URL matching semantics:** `Execute` receives a full URL string (e.g., `"https://x402.org/facilitator/verify?nonce=abc"`). The mock matches rules by comparing the request's `method` (exact, case-insensitive) and parsing the URL to extract `host+path` (ignoring scheme and query parameters). The `urlPattern` in `AddRule` is matched against `host+path`. Rules are checked in registration order; the first match wins. If no rule matches, `Execute` returns an error: `fmt.Errorf("mock: no rule matched %s %s", method, urlStr)`. This error surfaces as a Lua error in the plugin code, which the test can catch with `test.assert_error`.

**Mock response shape:** The response table must contain `status` (integer, required). Optionally: `json` (table, serialized to response body and sets `content_type` to `application/json`) or `body` (string, raw body). If `status` is missing, default to 200. The Go-side `Execute` returns `map[string]any{"status": statusCode, "body": bodyString, "headers": headerMap, "json": parsedJSON}` matching the real `RequestEngine`'s return shape.

From Lua:

```lua
test.mock_request("POST", "x402.org/facilitator/verify", { status = 200, json = { valid = true } })
-- matches request.post("https://x402.org/facilitator/verify?nonce=abc", ...) because scheme and query are ignored
```

### HTTP route testing

`test.request("GET", "/rules")` dispatches through the real `HTTPBridge` mounted on the harness mux via `httptest.NewRequest` + `httptest.NewRecorder`. The request MUST go through `h.mux.ServeHTTP(recorder, req)` so that Go's `ServeMux` populates `r.Pattern`, which the bridge uses for route lookup. Returns the response as a Lua table.

The harness injects a synthetic admin user context into every test request so that admin-only plugin routes receive a valid authenticated context. Auth is enforced at the Go level by the bridge's `ServeHTTP` method (via `middleware.AuthenticatedUser(r.Context())`), not by a Lua API function. The Lua handler never explicitly checks auth -- the bridge does it before dispatch.

The synthetic user has admin role with all permissions. This is appropriate because the test framework tests plugin logic, not CMS authorization. If a test needs to verify behavior for unauthenticated or non-admin users, the `opts` table accepts an optional `auth` field with exactly three values:

- absent or `"admin"` (default): inject synthetic admin user with all permissions
- `"none"`: no auth context (anonymous request)
- `"viewer"`: inject synthetic viewer user with viewer-role permissions
- Any other value is a Lua error: `error("test.request: auth must be 'admin', 'viewer', or 'none', got: " .. tostring(auth))`

**Auth context injection implementation:** The bridge's `ServeHTTP` reads `middleware.AuthenticatedUser(r.Context())` for non-public routes. The Lua handler also gets auth info from the request context. Three context values must be set:

1. `middleware.SetAuthenticatedUser(ctx, &syntheticUser)` -- already exported
2. Permissions via `permissionsKey` -- currently unexported in `internal/middleware/authorization.go`
3. Admin flag via `isAdminKey` -- currently unexported in `internal/middleware/authorization.go`

**Required change to `internal/middleware/authorization.go`:** Add two exported setter functions (parallel to the existing `SetAuthenticatedUser`):

```go
func SetPermissions(ctx context.Context, ps PermissionSet) context.Context {
    return context.WithValue(ctx, permissionsKey, ps)
}

func SetIsAdmin(ctx context.Context, isAdmin bool) context.Context {
    return context.WithValue(ctx, isAdminKey, isAdmin)
}
```

The harness then injects auth per request:
- `"admin"`: `SetAuthenticatedUser` with synthetic admin user + `SetPermissions` with all permissions + `SetIsAdmin(ctx, true)`
- `"viewer"`: `SetAuthenticatedUser` with synthetic viewer user + `SetPermissions` with the 5 viewer permissions + `SetIsAdmin(ctx, false)`
- `"none"`: no context injection

The viewer permission set can be hardcoded to the 5 bootstrap viewer permissions (`content:read`, `media:read`, `routes:read`, `datatypes:read`, `fields:read`) or loaded from the bootstrapped DB via `PermissionCache`.

`test.request` always returns the HTTP response table regardless of status code. Non-2xx responses are not errors at the test framework level -- the test author asserts on the status field. Example: `test.assert_eq(404, resp.status)`

### Hook testing

`test.fire_hook(event, table, data, state?)` dispatches to the appropriate hook engine method based on event prefix and returns a Lua-typed result. The return contract varies by hook category:

**Before-read hooks** (`event = "before_read"`):
Calls `hookEngine.RunBeforeReadHooks(ctx, table, data)`.
Returns three values to Lua: `(response_table_or_nil, state_map_or_nil, error_string_or_nil)`.
`response_table` contains `{status, headers, body}` if the hook produced a response (e.g., 402).
`state_map` contains any key-value pairs the hook added to the read state.

**Before-mutation hooks** (`event = "before_create"`, `"before_update"`, `"before_delete"`):
Calls `hookEngine.RunBeforeHooks(ctx, audited.HookEvent(event), table, data)`.
Returns one value to Lua: `(error_string_or_nil)`.

**After-mutation hooks** (`event = "after_create"`, `"after_update"`, `"after_delete"`):
Calls `hookEngine.RunAfterHooks(ctx, audited.HookEvent(event), table, data)`.
`RunAfterHooks` has no return value (void). Returns nil to Lua always.
These are fire-and-forget in production. In tests, any Lua error inside the hook will still be caught by the hook engine's internal pcall and logged, but there is no error return to surface.

**After-read hooks** (`event = "after_read"`):
Uses a DIFFERENT engine method than after-mutation hooks.
Calls `hookEngine.RunAfterReadHooks(ctx, table, data, state)`.
Requires a fourth parameter: `state` (`map[string]any`) from a prior `before_read` call.
Returns two values to Lua: `(headers_map_or_nil, error_string_or_nil)`.
`headers_map` is `map[string]string` of response headers the hook wants to set.

The `state` parameter is required for `after_read` events and ignored for all others. If `state` is nil/missing for an `after_read` event, return an error string.

**Examples:**

```lua
-- before_read
local resp, state, err = test.fire_hook("before_read", "content_data", { slug = "/test", headers = {} })

-- after_read (requires state from before_read)
local headers, err2 = test.fire_hook("after_read", "content_data", { slug = "/test" }, state)

-- before_mutation
local err3 = test.fire_hook("before_create", "content_data", { title = "New Post" })
```

### Report output

Default (human-readable):

```
x402 plugin tests

  rules.test.lua
    ✓ test_create_rule (2ms)
    ✓ test_list_rules (1ms)
    ✗ test_delete_nonexistent (3ms)
        FAIL: expected 404, got 500 (rules.test.lua:45)

  payment.test.lua
    ✓ test_build_requirements (1ms)

3 passed, 1 failed (7ms)
```

With `--json`: NDJSON, one JSON object per line per test result:

```json
{"file":"rules.test.lua","test":"test_create_rule","passed":true,"duration_ms":2,"assertions":3,"failures":[]}
{"file":"rules.test.lua","test":"test_delete_nonexistent","passed":false,"duration_ms":3,"assertions":2,"failures":[{"message":"expected 404, got 500","line":45}]}
```

Fields: `file` (string), `test` (string), `passed` (bool), `duration_ms` (int), `assertions` (int, total assertion count), `failures` (array of `{message, line}` objects). If the test hit a Lua runtime error, include it as an additional failure with `"error": true`.

## CLI command

```
modulacms plugin test <path> [flags]
```

Flags:

```
-v, --verbose   Print all assertions, not just failures
--filter <str>  Run only test functions whose name contains this exact substring (file names are NOT matched)
--json          Output as NDJSON
--timeout <sec> Per-test timeout (default 5)
```

## Exports needed from `internal/plugin`

The test harness package (`internal/plugin/testing`) imports `internal/plugin`. Several unexported symbols need to be exported:

- `outboundExecutor` -> `OutboundExecutor` (interface, used by `MockRequestEngine`)
- `setVMPhase` -> `SetVMPhase` (called in test VM setup)
- `vmPhase` -> `VMPhase` (if test needs to read phase)

Use `checkfor` to discover ALL occurrences of `setVMPhase`, `vmPhase`, and `outboundExecutor` in `internal/plugin/`. The counts below are approximate guides, not exhaustive -- `checkfor` output is the source of truth. Do not skip any occurrences that `checkfor` finds.

Approximate callers:

- `setVMPhase`: `manager.go` (5), `request_api_test.go` (~30), `hooks_api_test.go` (1), `http_api_test.go` (3)
- `vmPhase`: `hooks_api.go` (1), `http_api.go` (2), `request_api_test.go` (5). Note: `request_api.go` contains the definition of `vmPhase`, not a call site.
- `outboundExecutor`: `request_api.go` (3 -- interface + field + parameter), `request_api_test.go` (1)

Verify with: `checkfor` to find all occurrences, `repfor` with `dry_run`, then `repfor` to apply, then `just check`.

### `outboundOverride` injection

Manager needs a new field `outboundOverride OutboundExecutor` with a setter. This field is checked ONLY in the VM factory (where `RegisterRequestAPI` is called), not elsewhere. The existing `m.requestEngine` field remains intact for table creation, upserts, cleanup, and approval loading.

In `manager.go`'s VM factory (around the `RegisterRequestAPI` call):

```go
var reqExecutor OutboundExecutor = m.requestEngine
if m.outboundOverride != nil {
    reqExecutor = m.outboundOverride
}
reqAPI := RegisterRequestAPI(L, pluginName, reqExecutor)
```

Do NOT replace `m.requestEngine` itself -- only override the executor passed to Lua VMs.

## Implementation order

1. Export symbols in `request_api.go`, update all callers
2. Add `outboundOverride` to Manager
3. `internal/plugin/testing/results.go` -- types, formatting
4. `internal/plugin/testing/mock_request.go`
5. `internal/plugin/testing/lua_test_api.go`
6. `internal/plugin/testing/harness.go`
7. `cmd/plugin_test_cmd.go` + register in `cmd/plugin.go`
8. Test fixture + Go-level harness tests
9. Write example tests for x402 plugin in `plugins/x402/test/`

## Verification

1. `just check` -- compiles
2. `go test ./internal/plugin/testing/...` -- harness self-tests pass
3. `modulacms plugin test ./plugins/x402` -- discovers and runs x402 tests
4. `modulacms plugin test ./plugins/x402 --json` -- NDJSON output
5. `modulacms plugin test ./plugins/x402 --filter test_create` -- filters
6. Existing tests still pass: `go test ./internal/plugin/...`
