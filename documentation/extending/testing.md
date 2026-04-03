# Testing Plugins

Run automated tests for your Lua plugins without starting the CMS server. The test framework loads your plugin into an isolated in-memory database, discovers test files, and reports results.

## Quick Start

Create a `test/` directory inside your plugin with files ending in `.test.lua`:

```
plugins/my-plugin/
  init.lua
  lib/
  test/
    routes.test.lua
    hooks.test.lua
```

Each test file contains global functions starting with `test_`:

```lua
function test_hello()
    test.assert(true, "this passes")
    test.assert_eq(1 + 1, 2, "math works")
end

function test_my_route()
    local resp = test.request("GET", "/api/v1/plugins/my-plugin/items")
    test.assert_eq(200, resp.status)
    test.assert_not_nil(resp.json)
end
```

Run them:

```bash
modula plugin test ./plugins/my-plugin
```

```
my-plugin plugin tests

  routes.test.lua
    > test_hello (0ms)
    > test_my_route (1ms)

2 passed, 0 failed (1ms)
```

## Command Flags

```bash
modula plugin test <path> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-v`, `--verbose` | `false` | Print all assertions, not just failures |
| `--filter <str>` | | Run only test functions whose name contains this substring |
| `--json` | `false` | Output as NDJSON (one JSON object per line per test) |
| `--timeout <sec>` | `5` | Per-test timeout in seconds |

Exit codes: `0` all tests passed, `1` one or more failed, `2` plugin load error.

### Filtering

`--filter` matches against function names, not file names. Only functions containing the substring run:

```bash
modula plugin test ./plugins/my-plugin --filter settings
# Runs test_settings_default, test_settings_update, etc.
# Skips test_create_item, test_delete_item, etc.
```

### JSON Output

`--json` outputs one JSON object per line (NDJSON), one per test function:

```json
{"file":"routes.test.lua","test":"test_hello","passed":true,"duration_ms":0,"assertions":2,"failures":[]}
{"file":"routes.test.lua","test":"test_my_route","passed":false,"duration_ms":1,"assertions":1,"failures":[{"message":"expected 200, got 500","line":12}]}
```

## Test Isolation

Each `test_*` function runs inside a SQLite SAVEPOINT. After the function completes (pass or fail), the database rolls back to the savepoint. This means:

- Tests do not depend on execution order.
- A failing test cannot leave dirty state for the next test.
- Your plugin's `on_init` seed data is always the baseline.

The entire test harness runs in a fresh in-memory SQLite database with all CMS tables bootstrapped. Your plugin's `on_init` runs once, creating its tables and seed data. Then each test function gets a clean snapshot of that state.

## Assertions

All assertions are soft failures. They record the failure and continue executing the test function, giving you a complete picture of every failure per test.

### test.assert(condition, message?)

Fails if `condition` is `nil` or `false`.

```lua
test.assert(result ~= nil, "result should exist")
test.assert(count > 0)
```

### test.assert_eq(expected, actual, message?)

Deep equality check. Works with numbers, strings, booleans, nil, and nested tables.

```lua
test.assert_eq(200, resp.status, "status code")
test.assert_eq({a = 1}, {a = 1}, "tables match")
```

### test.assert_neq(a, b, message?)

Fails if the two values are equal.

```lua
test.assert_neq(old_id, new_id, "IDs should differ")
```

### test.assert_nil(value, message?)

Fails if `value` is not `nil`.

```lua
test.assert_nil(err, "should not error")
```

### test.assert_not_nil(value, message?)

Fails if `value` is `nil`.

```lua
test.assert_not_nil(resp.json.id, "should return an ID")
```

### test.assert_error(fn, pattern?)

Asserts that `fn()` raises a Lua error. If `pattern` is given, the error message must contain it (plain string match, not regex).

```lua
test.assert_error(function()
    db.insert("nonexistent_table", {})
end, "no such table")
```

### test.assert_contains(haystack, needle, message?)

Fails if `haystack` does not contain `needle` (string match).

```lua
test.assert_contains(resp.body, "payment required")
```

## Testing Routes

`test.request(method, path, opts?)` sends an HTTP request through your plugin's route handlers and returns the response.

```lua
local resp = test.request("GET", "/api/v1/plugins/my-plugin/items")
```

The response table contains:

| Field | Type | Description |
|-------|------|-------------|
| `status` | number | HTTP status code |
| `body` | string | Raw response body |
| `json` | table/nil | Parsed JSON body (if Content-Type is `application/json`) |
| `headers` | table | Response headers (lowercase keys) |

### Request Options

The optional third argument is a table:

```lua
local resp = test.request("POST", "/api/v1/plugins/my-plugin/items", {
    body = '{"title": "New Item"}',
    headers = { ["x-custom"] = "value" },
    auth = "admin",
})
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `body` | string | | Request body (sets Content-Type to `application/json` automatically) |
| `headers` | table | | Custom request headers |
| `auth` | string | `"admin"` | Authentication context (see below) |

### Authentication

By default, every test request runs as a synthetic admin user with all permissions. You can change this per-request with the `auth` option:

| Value | Behavior |
|-------|----------|
| `"admin"` (default) | Full admin with all permissions |
| `"viewer"` | Read-only user with viewer permissions |
| `"none"` | No authentication (anonymous request) |

```lua
-- Test that unauthenticated users get 401
local resp = test.request("GET", "/api/v1/plugins/my-plugin/admin-only", {
    auth = "none",
})
test.assert_eq(401, resp.status)
```

## Testing Hooks

`test.fire_hook(event, table, data, state?)` fires a content lifecycle hook and returns the result. The return values depend on the hook type.

### Before-read hooks

Returns three values: `response_or_nil`, `state_or_nil`, `error_or_nil`.

```lua
local resp, state, err = test.fire_hook("before_read", "content_data", {
    slug = "/gated-content",
    headers = {},
})

if resp then
    test.assert_eq(402, resp.status, "should require payment")
end
```

### After-read hooks

Requires a fourth argument (`state` from a prior `before_read`):

```lua
local headers, err = test.fire_hook("after_read", "content_data", {
    slug = "/gated-content",
}, state)
```

### Before-mutation hooks

Returns one value: `error_or_nil`.

```lua
local err = test.fire_hook("before_create", "content_data", {
    title = "New Post",
})
test.assert_nil(err, "hook should allow creation")
```

### After-mutation hooks

Fire-and-forget. Always returns `nil`.

```lua
test.fire_hook("after_create", "content_data", { title = "New Post" })
-- Verify side effects (e.g., check a plugin table was updated)
```

### Supported events

`before_read`, `after_read`, `before_create`, `after_create`, `before_update`, `after_update`, `before_delete`, `after_delete`, `before_publish`, `after_publish`.

## Mocking Outbound HTTP

If your plugin makes outbound HTTP requests (via the `request` module), use `test.mock_request` to register canned responses.

```lua
test.mock_request("POST", "api.example.com/verify", {
    status = 200,
    json = { valid = true, receipt = "abc123" },
})
```

Matching ignores the URL scheme and query parameters. Only the method and `host + path` are compared:

```lua
-- This mock:
test.mock_request("GET", "api.stripe.com/v1/charges", { status = 200, json = {} })

-- Matches this plugin call:
request.get("https://api.stripe.com/v1/charges?limit=10", { parse_json = true })
```

If no mock matches an outbound request, a Lua error is raised. You can catch this with `test.assert_error`:

```lua
test.assert_error(function()
    request.get("https://unmocked.example.com/path")
end, "no rule matched")
```

Mock rules are cleared between test functions automatically (SAVEPOINT rollback handles DB state; mock rules are cleared in-memory).

### Mock response format

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | number | No (default 200) | HTTP status code |
| `json` | table | No | Response body as a Lua table (serialized to JSON) |
| `body` | string | No | Raw response body (use instead of `json` for non-JSON) |

## Setup and Teardown

Register functions that run before and after each `test_*` function in the file:

```lua
test.setup(function()
    -- Seed test data
    db.insert("items", { id = db.ulid(), title = "Seed Item" })
end)

test.teardown(function()
    -- Non-DB cleanup (DB changes are rolled back automatically)
end)
```

Execution order per test: SAVEPOINT, setup, test function, teardown, ROLLBACK.

Because the SAVEPOINT rolls back all database changes, teardown is primarily useful for non-DB cleanup like resetting Lua module state or clearing mock rules manually.

## Using Plugin APIs in Tests

Test files have access to every API your plugin has in production:

| Module | Available |
|--------|-----------|
| `db` | Yes -- queries run against the in-memory database |
| `core` | Yes -- reads CMS tables if declared in `core_access` |
| `log` | Yes -- log output is suppressed during tests |
| `request` | Yes -- outbound calls go through mock rules |
| `json` | Yes |
| `hooks` | No -- use `test.fire_hook` instead |
| `http` | No -- use `test.request` instead |

Your `lib/` modules are available via `require()`:

```lua
local utils = require("utils")
local payment = require("payment")

function test_build_requirements()
    local reqs = payment.build_requirements({
        { price = "0.01", currency = "USDC", network = "base" },
    })
    test.assert_not_nil(reqs)
    test.assert_eq(1, #reqs)
end
```

## Complete Example

A test file for a payment gating plugin:

```lua
-- test/payment.test.lua

local payment = require("payment")

test.setup(function()
    -- Configure the plugin
    db.insert("settings", {
        id = db.ulid(),
        key = "wallet_address",
        value = "0xTestWallet",
    })
end)

function test_ungated_content()
    local resp = test.request("GET", "/api/v1/plugins/x402/check/no-rules", {
        auth = "none",
    })
    test.assert_eq(200, resp.status)
    test.assert_eq(false, resp.json.gated)
end

function test_create_rule_requires_fields()
    local resp = test.request("POST", "/api/v1/plugins/x402/rules", {
        body = "{}",
    })
    test.assert_eq(400, resp.status)
end

function test_settings_roundtrip()
    local get = test.request("GET", "/api/v1/plugins/x402/settings")
    test.assert_eq(200, get.status)

    local put = test.request("PUT", "/api/v1/plugins/x402/settings", {
        body = '{"wallet_address": "0xNewWallet"}',
    })
    test.assert_eq(200, put.status)
end

function test_mock_facilitator()
    test.mock_request("POST", "x402.org/facilitator/verify", {
        status = 200,
        json = { valid = true },
    })

    -- Now any plugin code calling request.post("https://x402.org/facilitator/verify", ...)
    -- gets this canned response instead of making a real HTTP call.
end

function test_unauthenticated_admin_route()
    local resp = test.request("GET", "/api/v1/plugins/x402/rules", {
        auth = "none",
    })
    test.assert_eq(401, resp.status)
end
```

Run it:

```bash
modula plugin test ./plugins/x402
```

## Tips

- **Name tests descriptively.** Function names appear in the output. `test_delete_nonexistent_returns_404` is clearer than `test_delete_2`.
- **One assertion per concern.** Soft failures let you see all problems at once, but grouping unrelated assertions makes the failure message less useful.
- **Use `--filter` during development.** Run just the test you're working on: `modula plugin test ./plugins/x402 --filter test_create`.
- **Test unhappy paths.** Verify that missing fields return 400, unauthorized requests return 401, and not-found resources return 404.
- **Mock outbound calls.** If your plugin contacts external services, mock them. Tests should not depend on network access.

## Next Steps

- [Lua API Reference](lua-api.md) for the full list of available functions
- [Examples](examples.md) for complete working plugins
- [Approval Workflow](approval.md) for route and hook approval
