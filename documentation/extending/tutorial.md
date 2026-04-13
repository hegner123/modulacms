# Plugin Tutorial: Build a Bookmarks Plugin

Build a complete plugin with database tables, REST endpoints, a content hook, middleware, and a helper module.

> **Good to know**: This tutorial requires a running ModulaCMS instance with `plugin_enabled: true` in `modula.config.json`. See [configuration](/docs/extending/configuration) for setup.

## 1. Scaffold the Plugin

```bash
modulacms plugin init bookmarks \
  --version 1.0.0 \
  --description "Save and organize bookmarks" \
  --author "Your Name" \
  --license MIT
```

This creates:

```
plugins/
  bookmarks/
    init.lua
    lib/
```

The generated `init.lua` contains a skeleton manifest and empty lifecycle functions.

## 2. Write the Manifest

Open `plugins/bookmarks/init.lua` and replace the contents with:

```lua
plugin_info = {
    name        = "bookmarks",
    version     = "1.0.0",
    description = "Save and organize bookmarks",
    author      = "Your Name",
    license     = "MIT",
}
```

The `name` field must match the directory name. It becomes the database table prefix (`plugin_bookmarks_`) and the route prefix (`/api/v1/plugins/bookmarks/`).

**Version matters.** When you change the version string later, all route and hook approvals are revoked and require re-approval. This prevents stale approvals from applying to updated code.

## 3. Define Tables

Add an `on_init()` function to create your database tables. This runs once after all VMs are created.

```lua
function on_init()
    db.define_table("collections", {
        columns = {
            { name = "name",        type = "text",    not_null = true },
            { name = "description", type = "text" },
            { name = "sort_order",  type = "integer", not_null = true, default = 0 },
            { name = "is_public",   type = "boolean", not_null = true, default = 0 },
        },
        indexes = {
            { columns = {"name"} },
            { columns = {"is_public", "sort_order"} },
        },
    })

    db.define_table("bookmarks", {
        columns = {
            { name = "collection_id", type = "text",      not_null = true },
            { name = "url",           type = "text",      not_null = true },
            { name = "title",         type = "text",      not_null = true },
            { name = "rating",        type = "real" },
            { name = "notes",         type = "text" },
            { name = "visited_at",    type = "timestamp" },
        },
        indexes = {
            { columns = {"collection_id"} },
            { columns = {"url"} },
        },
        foreign_keys = {
            {
                column     = "collection_id",
                ref_table  = "plugin_bookmarks_collections",
                ref_column = "id",
                on_delete  = "CASCADE",
            },
        },
    })

    log.info("bookmarks plugin initialized")
end
```

Three columns are auto-injected on every table (`id`, `created_at`, `updated_at`) -- do not include them in your definition.

Foreign keys must reference tables owned by the same plugin. Use the full prefixed name `plugin_bookmarks_collections` in `ref_table`.

## 4. Register CRUD Routes

Add route registrations at module scope (top-level code), **not** inside `on_init()`. Register routes and hooks at the top level.

```lua
-- List all bookmarks (with optional collection filter)
http.handle("GET", "/bookmarks", function(req)
    local opts = { order_by = "created_at DESC", limit = 50 }
    if req.query.collection_id then
        opts.where = { collection_id = req.query.collection_id }
    end
    local bookmarks = db.query("bookmarks", opts)
    local total = db.count("bookmarks", opts.where and { where = opts.where } or {})
    return {
        status = 200,
        json = { bookmarks = bookmarks, total = total },
    }
end)

-- Create a bookmark
http.handle("POST", "/bookmarks", function(req)
    if not req.json then
        return { status = 400, json = { error = "JSON body required" } }
    end
    local b = req.json
    if not b.url or b.url == "" then
        return { status = 400, json = { error = "url is required" } }
    end
    if not b.title or b.title == "" then
        return { status = 400, json = { error = "title is required" } }
    end
    if not b.collection_id or b.collection_id == "" then
        return { status = 400, json = { error = "collection_id is required" } }
    end
    if not db.exists("collections", { where = { id = b.collection_id } }) then
        return { status = 400, json = { error = "collection not found" } }
    end

    local id = db.ulid()
    db.insert("bookmarks", {
        id            = id,
        collection_id = b.collection_id,
        url           = b.url,
        title         = b.title,
        rating        = b.rating,
        notes         = b.notes,
    })

    local created = db.query_one("bookmarks", { where = { id = id } })
    return { status = 201, json = created }
end)

-- Get a single bookmark
http.handle("GET", "/bookmarks/{id}", function(req)
    local bookmark = db.query_one("bookmarks", { where = { id = req.params.id } })
    if not bookmark then
        return { status = 404, json = { error = "not found" } }
    end
    return { status = 200, json = bookmark }
end)

-- Update a bookmark
http.handle("PUT", "/bookmarks/{id}", function(req)
    if not req.json then
        return { status = 400, json = { error = "JSON body required" } }
    end
    if not db.exists("bookmarks", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end

    local updates = {}
    local b = req.json
    if b.url then updates.url = b.url end
    if b.title then updates.title = b.title end
    if b.rating then updates.rating = b.rating end
    if b.notes then updates.notes = b.notes end

    if next(updates) == nil then
        return { status = 400, json = { error = "no fields to update" } }
    end

    db.update("bookmarks", {
        set   = updates,
        where = { id = req.params.id },
    })

    local updated = db.query_one("bookmarks", { where = { id = req.params.id } })
    return { status = 200, json = updated }
end)

-- Delete a bookmark
http.handle("DELETE", "/bookmarks/{id}", function(req)
    if not db.exists("bookmarks", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end
    db.delete("bookmarks", { where = { id = req.params.id } })
    return { status = 204 }
end)
```

Each route becomes `/api/v1/plugins/bookmarks/<path>`. By default, routes require CMS authentication. Pass `{ public = true }` as the fourth argument to make a route publicly accessible.

## 5. Add a Content Hook

Register a hook at module scope to react when new CMS content is created:

```lua
hooks.on("after_create", "content_data", function(data)
    db.insert("activity", {
        action     = "content_created",
        content_id = data.id,
        title      = data.title or "(untitled)",
    })
    log.info("Logged new content creation", { content_id = data.id })
end)
```

This hook runs asynchronously after the CMS transaction commits. It has full `db.*` access because it is an after-hook. Before-hooks block `db.*` calls to prevent transaction deadlocks.

Add the `activity` table to `on_init()`:

```lua
db.define_table("activity", {
    columns = {
        { name = "action",     type = "text", not_null = true },
        { name = "content_id", type = "text" },
        { name = "title",      type = "text" },
    },
    indexes = {
        { columns = {"action"} },
    },
})
```

## 6. Add Middleware

Middleware runs before all route handlers. Return a response table to short-circuit, or `nil` to continue to the route handler.

```lua
http.use(function(req)
    -- Require API key on all public endpoints
    if not req.headers["x-api-key"] then
        return nil  -- authenticated routes handle their own auth
    end
    -- If an API key is provided, validate it
    if req.headers["x-api-key"] ~= "expected-key-here" then
        return { status = 401, json = { error = "invalid api key" } }
    end
    return nil
end)
```

## 7. Extract Helpers

Create `plugins/bookmarks/lib/validators.lua`:

```lua
local M = {}

function M.is_valid_url(url)
    if type(url) ~= "string" then return false end
    if url == "" then return false end
    return url:sub(1, 7) == "http://" or url:sub(1, 8) == "https://"
end

function M.not_empty(s)
    return type(s) == "string" and #s > 0
end

function M.trim(s)
    if type(s) ~= "string" then return s end
    return s:match("^%s*(.-)%s*$")
end

return M
```

Use it in `init.lua`:

```lua
local validators = require("validators")
-- Resolves to: plugins/bookmarks/lib/validators.lua
```

Only files under the plugin's own `lib/` directory are loadable. Path traversal (`..`, `/`, `\`) is rejected. Modules are cached after first load.

## 8. Validate

Before deploying, validate the plugin manifest and structure:

```bash
modulacms plugin validate ./plugins/bookmarks
```

This checks:
- `plugin_info` table exists and has required fields
- Name matches directory name (lowercase alphanumeric plus underscores, max 32 chars)
- Version string is present
- `init.lua` parses without syntax errors

Validation does not execute the plugin code or verify database operations.

## 9. Deploy and Approve

Copy the plugin directory to the server's `plugin_directory` path (default `./plugins/`). If the server is running:

- With `plugin_hot_reload: true`, the watcher detects the new directory and loads the plugin automatically.
- Without hot reload, restart the server to pick up the plugin.

After the plugin loads, approve its routes and hooks:

```bash
# Approve all routes and hooks at once
modulacms plugin approve bookmarks --all-routes --all-hooks --yes

# Or approve individually
modulacms plugin approve bookmarks --route "GET /bookmarks"
modulacms plugin approve bookmarks --route "POST /bookmarks"
modulacms plugin approve bookmarks --route "GET /bookmarks/{id}"
modulacms plugin approve bookmarks --route "PUT /bookmarks/{id}"
modulacms plugin approve bookmarks --route "DELETE /bookmarks/{id}"
modulacms plugin approve bookmarks --hook "after_create:content_data"
```

See [approval workflow](/docs/extending/approval) for CLI, API, and TUI approval details.

## 10. Test with curl

```bash
BASE="http://localhost:8080/api/v1/plugins/bookmarks"

# Create a collection first (requires auth cookie)
curl -X POST "$BASE/collections" \
  -H "Cookie: session=YOUR_SESSION" \
  -H "Content-Type: application/json" \
  -d '{"name": "Dev Resources", "is_public": 1}'

# Create a bookmark
curl -X POST "$BASE/bookmarks" \
  -H "Cookie: session=YOUR_SESSION" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "COLLECTION_ID_HERE",
    "url": "https://go.dev/doc/",
    "title": "Go Documentation",
    "rating": 4.5
  }'

# List bookmarks
curl "$BASE/bookmarks" -H "Cookie: session=YOUR_SESSION"

# Get a single bookmark
curl "$BASE/bookmarks/BOOKMARK_ID" -H "Cookie: session=YOUR_SESSION"

# Update
curl -X PUT "$BASE/bookmarks/BOOKMARK_ID" \
  -H "Cookie: session=YOUR_SESSION" \
  -H "Content-Type: application/json" \
  -d '{"rating": 5.0}'

# Delete
curl -X DELETE "$BASE/bookmarks/BOOKMARK_ID" \
  -H "Cookie: session=YOUR_SESSION"
```

You need to add collection CRUD routes following the same pattern as bookmarks to make this fully functional.

## 11. Hot Reload for Development

Enable hot reload in `modula.config.json`:

```json
{
  "plugin_hot_reload": true
}
```

The watcher polls every 2 seconds for `.lua` file changes. When it detects changes:

1. A 1-second debounce window waits for file writes to settle.
2. A new plugin instance loads alongside the old one (blue-green).
3. If the new instance succeeds, it replaces the old one atomically.
4. If it fails, the old instance keeps running.

A 10-second cooldown prevents reload storms during rapid iteration. After 3 consecutive slow reloads (>10s each), the watcher pauses for that plugin.

Manual reload is always available:

```bash
modulacms plugin reload bookmarks
```

## Complete init.lua

For reference, here is the complete plugin file combining all sections:

```lua
local validators = require("validators")

plugin_info = {
    name        = "bookmarks",
    version     = "1.0.0",
    description = "Save and organize bookmarks",
    author      = "Your Name",
    license     = "MIT",
}

-- Middleware
http.use(function(req)
    if req.headers["x-api-key"] and req.headers["x-api-key"] ~= "expected-key-here" then
        return { status = 401, json = { error = "invalid api key" } }
    end
    return nil
end)

-- Routes
http.handle("GET", "/bookmarks", function(req)
    local opts = { order_by = "created_at DESC", limit = 50 }
    if req.query.collection_id then
        opts.where = { collection_id = req.query.collection_id }
    end
    local bookmarks = db.query("bookmarks", opts)
    local total = db.count("bookmarks", opts.where and { where = opts.where } or {})
    return { status = 200, json = { bookmarks = bookmarks, total = total } }
end)

http.handle("POST", "/bookmarks", function(req)
    if not req.json then
        return { status = 400, json = { error = "JSON body required" } }
    end
    local b = req.json
    if not validators.not_empty(b.url) then
        return { status = 400, json = { error = "url is required" } }
    end
    if not validators.is_valid_url(b.url) then
        return { status = 400, json = { error = "url must be http:// or https://" } }
    end
    if not validators.not_empty(b.title) then
        return { status = 400, json = { error = "title is required" } }
    end
    if not validators.not_empty(b.collection_id) then
        return { status = 400, json = { error = "collection_id is required" } }
    end
    if not db.exists("collections", { where = { id = b.collection_id } }) then
        return { status = 400, json = { error = "collection not found" } }
    end

    local id = db.ulid()
    db.insert("bookmarks", {
        id            = id,
        collection_id = b.collection_id,
        url           = b.url,
        title         = validators.trim(b.title),
        rating        = b.rating,
        notes         = b.notes,
    })

    local created = db.query_one("bookmarks", { where = { id = id } })
    return { status = 201, json = created }
end)

http.handle("GET", "/bookmarks/{id}", function(req)
    local bookmark = db.query_one("bookmarks", { where = { id = req.params.id } })
    if not bookmark then
        return { status = 404, json = { error = "not found" } }
    end
    return { status = 200, json = bookmark }
end)

http.handle("PUT", "/bookmarks/{id}", function(req)
    if not req.json then
        return { status = 400, json = { error = "JSON body required" } }
    end
    if not db.exists("bookmarks", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end
    local updates = {}
    if req.json.url then updates.url = req.json.url end
    if req.json.title then updates.title = req.json.title end
    if req.json.rating then updates.rating = req.json.rating end
    if req.json.notes then updates.notes = req.json.notes end
    if next(updates) == nil then
        return { status = 400, json = { error = "no fields to update" } }
    end
    db.update("bookmarks", { set = updates, where = { id = req.params.id } })
    local updated = db.query_one("bookmarks", { where = { id = req.params.id } })
    return { status = 200, json = updated }
end)

http.handle("DELETE", "/bookmarks/{id}", function(req)
    if not db.exists("bookmarks", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end
    db.delete("bookmarks", { where = { id = req.params.id } })
    return { status = 204 }
end)

-- Hooks
hooks.on("after_create", "content_data", function(data)
    db.insert("activity", {
        action     = "content_created",
        content_id = data.id,
        title      = data.title or "(untitled)",
    })
    log.info("Logged new content creation", { content_id = data.id })
end)

-- Lifecycle
function on_init()
    db.define_table("collections", {
        columns = {
            { name = "name",        type = "text",    not_null = true },
            { name = "description", type = "text" },
            { name = "sort_order",  type = "integer", not_null = true, default = 0 },
            { name = "is_public",   type = "boolean", not_null = true, default = 0 },
        },
        indexes = {
            { columns = {"name"} },
            { columns = {"is_public", "sort_order"} },
        },
    })

    db.define_table("bookmarks", {
        columns = {
            { name = "collection_id", type = "text",      not_null = true },
            { name = "url",           type = "text",      not_null = true },
            { name = "title",         type = "text",      not_null = true },
            { name = "rating",        type = "real" },
            { name = "notes",         type = "text" },
            { name = "visited_at",    type = "timestamp" },
        },
        indexes = {
            { columns = {"collection_id"} },
            { columns = {"url"} },
        },
        foreign_keys = {
            {
                column     = "collection_id",
                ref_table  = "plugin_bookmarks_collections",
                ref_column = "id",
                on_delete  = "CASCADE",
            },
        },
    })

    db.define_table("activity", {
        columns = {
            { name = "action",     type = "text", not_null = true },
            { name = "content_id", type = "text" },
            { name = "title",      type = "text" },
        },
        indexes = {
            { columns = {"action"} },
        },
    })

    log.info("bookmarks plugin initialized")
end

function on_shutdown()
    log.info("bookmarks plugin shutting down")
end
```

## 12. Write Tests

Create a `test/` directory with test files:

```bash
mkdir plugins/bookmarks/test
```

Create `plugins/bookmarks/test/bookmarks.test.lua`:

```lua
function test_list_empty()
    local resp = test.request("GET", "/api/v1/plugins/bookmarks/bookmarks")
    test.assert_eq(200, resp.status)
    test.assert_not_nil(resp.json)
end

function test_create_bookmark()
    local resp = test.request("POST", "/api/v1/plugins/bookmarks/bookmarks", {
        body = '{"url": "https://example.com", "title": "Example"}',
    })
    test.assert_eq(201, resp.status)
    test.assert_not_nil(resp.json.id)
end

function test_create_requires_url()
    local resp = test.request("POST", "/api/v1/plugins/bookmarks/bookmarks", {
        body = '{"title": "No URL"}',
    })
    test.assert_eq(400, resp.status)
end

function test_unauthenticated_returns_401()
    local resp = test.request("GET", "/api/v1/plugins/bookmarks/bookmarks", {
        auth = "none",
    })
    test.assert_eq(401, resp.status)
end
```

Run the tests:

```bash
modula plugin test ./plugins/bookmarks
```

No running server needed. Tests run in an isolated in-memory database with all CMS tables and your plugin's tables bootstrapped.

See [Testing Plugins](/docs/extending/testing) for the full test API, including assertions, hook testing, and outbound HTTP mocking.
