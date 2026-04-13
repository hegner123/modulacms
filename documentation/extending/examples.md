# Example Plugins

Learn common plugin patterns through complete, working examples with full `init.lua` source code.

For a step-by-step walkthrough of building a plugin from scratch, see the [tutorial](/docs/extending/tutorial).

## Task Tracker

A CRUD API with categories and tasks, demonstrating foreign keys, transactions, and pagination.

**Key patterns:** Multi-table relationships, transaction-based seeding, paginated list endpoints.

```lua
plugin_info = {
    name        = "task_tracker",
    version     = "1.0.0",
    description = "Task tracking with categories",
    author      = "Example",
}

-- List categories
http.handle("GET", "/categories", function(req)
    local categories = db.query("categories", {
        order_by = "name",
        limit    = 100,
    })
    return { status = 200, json = categories }
end)

-- Create category
http.handle("POST", "/categories", function(req)
    if not req.json or not req.json.name or req.json.name == "" then
        return { status = 400, json = { error = "name is required" } }
    end
    local id = db.ulid()
    db.insert("categories", {
        id   = id,
        name = req.json.name,
    })
    local created = db.query_one("categories", { where = { id = id } })
    return { status = 201, json = created }
end)

-- List tasks with pagination
http.handle("GET", "/tasks", function(req)
    local page = tonumber(req.query.page) or 1
    local per_page = tonumber(req.query.per_page) or 20
    if per_page > 100 then per_page = 100 end
    local offset = (page - 1) * per_page

    local opts = {
        order_by = "created_at DESC",
        limit    = per_page,
        offset   = offset,
    }

    -- Optional filters
    if req.query.status then
        opts.where = { status = req.query.status }
    end
    if req.query.category_id then
        opts.where = opts.where or {}
        opts.where.category_id = req.query.category_id
    end

    local tasks = db.query("tasks", opts)
    local total = db.count("tasks", opts.where and { where = opts.where } or {})

    return {
        status = 200,
        json = {
            tasks    = tasks,
            total    = total,
            page     = page,
            per_page = per_page,
            pages    = math.ceil(total / per_page),
        },
    }
end)

-- Create task
http.handle("POST", "/tasks", function(req)
    if not req.json then
        return { status = 400, json = { error = "JSON body required" } }
    end
    local t = req.json
    if not t.title or t.title == "" then
        return { status = 400, json = { error = "title is required" } }
    end
    if not t.category_id or t.category_id == "" then
        return { status = 400, json = { error = "category_id is required" } }
    end
    if not db.exists("categories", { where = { id = t.category_id } }) then
        return { status = 400, json = { error = "category not found" } }
    end

    local id = db.ulid()
    db.insert("tasks", {
        id          = id,
        title       = t.title,
        description = t.description or "",
        status      = "todo",
        priority    = t.priority or 0,
        category_id = t.category_id,
    })

    local created = db.query_one("tasks", { where = { id = id } })
    return { status = 201, json = created }
end)

-- Get task
http.handle("GET", "/tasks/{id}", function(req)
    local task = db.query_one("tasks", { where = { id = req.params.id } })
    if not task then
        return { status = 404, json = { error = "not found" } }
    end
    return { status = 200, json = task }
end)

-- Update task
http.handle("PUT", "/tasks/{id}", function(req)
    if not req.json then
        return { status = 400, json = { error = "JSON body required" } }
    end
    if not db.exists("tasks", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end

    local updates = {}
    local t = req.json
    if t.title then updates.title = t.title end
    if t.description then updates.description = t.description end
    if t.status then updates.status = t.status end
    if t.priority then updates.priority = t.priority end

    if next(updates) == nil then
        return { status = 400, json = { error = "no fields to update" } }
    end

    db.update("tasks", { set = updates, where = { id = req.params.id } })
    local updated = db.query_one("tasks", { where = { id = req.params.id } })
    return { status = 200, json = updated }
end)

-- Delete task
http.handle("DELETE", "/tasks/{id}", function(req)
    if not db.exists("tasks", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end
    db.delete("tasks", { where = { id = req.params.id } })
    return { status = 204 }
end)

function on_init()
    db.define_table("categories", {
        columns = {
            { name = "name", type = "text", not_null = true },
        },
        indexes = {
            { columns = {"name"}, unique = true },
        },
    })

    db.define_table("tasks", {
        columns = {
            { name = "title",       type = "text",    not_null = true },
            { name = "description", type = "text",    default = "" },
            { name = "status",      type = "text",    not_null = true, default = "todo" },
            { name = "priority",    type = "integer", default = 0 },
            { name = "category_id", type = "text",    not_null = true },
        },
        indexes = {
            { columns = {"status"} },
            { columns = {"category_id"} },
            { columns = {"status", "priority"} },
        },
        foreign_keys = {
            {
                column     = "category_id",
                ref_table  = "plugin_task_tracker_categories",
                ref_column = "id",
                on_delete  = "CASCADE",
            },
        },
    })

    -- Seed default categories using a transaction
    if db.count("categories", {}) == 0 then
        local ok, err = db.transaction(function()
            db.insert("categories", { name = "Bug" })
            db.insert("categories", { name = "Feature" })
            db.insert("categories", { name = "Improvement" })
            db.insert("categories", { name = "Documentation" })
        end)
        if not ok then
            log.error("Failed to seed categories", { err = err })
        end
    end

    log.info("task_tracker initialized")
end
```

---

## Content Validator

Before-hooks that enforce business rules on CMS content. Demonstrates validation logic, conditional abort, and priority ordering.

**Key patterns:** Before-hooks that call `error()` to abort transactions, multiple hooks with priority ordering, no `db.*` calls in before-hooks.

```lua
plugin_info = {
    name        = "content_validator",
    version     = "1.0.0",
    description = "Enforce content business rules",
    author      = "Example",
}

-- Validate required fields on content creation (runs first)
hooks.on("before_create", "content_data", function(data)
    if not data.title or data.title == "" then
        error("title is required")
    end
    if data.title and string.len(data.title) > 200 then
        error("title must be 200 characters or fewer")
    end
end, { priority = 10 })

-- Validate slug format on content creation (runs second)
hooks.on("before_create", "content_data", function(data)
    if data.slug then
        -- Only allow lowercase letters, numbers, and hyphens
        for i = 1, #data.slug do
            local c = data.slug:sub(i, i)
            local b = string.byte(c)
            local valid = (b >= 97 and b <= 122)  -- a-z
                       or (b >= 48 and b <= 57)   -- 0-9
                       or b == 45                  -- hyphen
            if not valid then
                error("slug contains invalid character: " .. c)
            end
        end
    end
end, { priority = 20 })

-- Validate content before publishing
hooks.on("before_publish", "content_data", function(data)
    if not data.title or data.title == "" then
        error("cannot publish content without a title")
    end
end, { priority = 10 })

-- Log validation passes (after-hook, can use db.*)
hooks.on("after_create", "content_data", function(data)
    log.info("Content passed validation", {
        id    = data.id,
        title = data.title,
    })
end)

function on_init()
    log.info("content_validator initialized")
end
```

The `before_create` hooks run at priorities 10 and 20, ensuring the required-fields check runs before the slug format check. Both hooks run inside the CMS transaction. If either calls `error()`, the create operation aborts and the client receives HTTP 422 with the error message.

No `db.*` calls appear in the before-hooks. Before-hooks block all database operations to prevent transaction deadlocks.

---

## Webhook Relay

After-hooks that queue notifications when content changes. Demonstrates public routes for receiving webhook registrations and async event logging via after-hooks.

**Key patterns:** After-hooks with database logging, public routes for inbound webhooks, structured error logging.

```lua
plugin_info = {
    name        = "webhook_relay",
    version     = "1.0.0",
    description = "Relay content events to external services",
    author      = "Example",
}

-- Public endpoint for external services to register webhooks
http.handle("POST", "/register", function(req)
    if not req.json then
        return { status = 400, json = { error = "JSON body required" } }
    end
    local w = req.json
    if not w.url or w.url == "" then
        return { status = 400, json = { error = "url is required" } }
    end
    if not w.events or type(w.events) ~= "string" then
        return { status = 400, json = { error = "events is required (comma-separated)" } }
    end

    local id = db.ulid()
    db.insert("endpoints", {
        id     = id,
        url    = w.url,
        events = w.events,
        active = 1,
        secret = w.secret or "",
    })

    return { status = 201, json = { id = id, url = w.url, events = w.events } }
end, { public = true })

-- List registered endpoints (authenticated)
http.handle("GET", "/endpoints", function(req)
    local endpoints = db.query("endpoints", {
        where = { active = 1 },
        order_by = "created_at DESC",
    })
    return { status = 200, json = endpoints }
end)

-- Delete an endpoint
http.handle("DELETE", "/endpoints/{id}", function(req)
    if not db.exists("endpoints", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end
    db.delete("endpoints", { where = { id = req.params.id } })
    return { status = 204 }
end)

-- List delivery attempts (authenticated)
http.handle("GET", "/deliveries", function(req)
    local deliveries = db.query("deliveries", {
        order_by = "created_at DESC",
        limit    = 50,
    })
    return { status = 200, json = deliveries }
end)

-- After-hooks: log content events for delivery
-- Plugins cannot make outbound HTTP requests directly.
-- This pattern logs events to a plugin table. An external worker
-- (cron job, queue consumer) polls the deliveries table and
-- performs the actual HTTP POST to registered endpoints.

hooks.on("after_create", "content_data", function(data)
    local endpoints = db.query("endpoints", { where = { active = 1 } })
    for _, ep in ipairs(endpoints) do
        if ep.events:find("create") then
            db.insert("deliveries", {
                endpoint_id = ep.id,
                event       = "content.created",
                payload     = '{"id":"' .. (data.id or "") .. '","title":"' .. (data.title or "") .. '"}',
                status      = "pending",
            })
        end
    end
    log.info("Queued webhook deliveries for content.created", { content_id = data.id })
end)

hooks.on("after_update", "content_data", function(data)
    local endpoints = db.query("endpoints", { where = { active = 1 } })
    for _, ep in ipairs(endpoints) do
        if ep.events:find("update") then
            db.insert("deliveries", {
                endpoint_id = ep.id,
                event       = "content.updated",
                payload     = '{"id":"' .. (data.id or "") .. '","title":"' .. (data.title or "") .. '"}',
                status      = "pending",
            })
        end
    end
end)

hooks.on("after_delete", "content_data", function(data)
    local endpoints = db.query("endpoints", { where = { active = 1 } })
    for _, ep in ipairs(endpoints) do
        if ep.events:find("delete") then
            db.insert("deliveries", {
                endpoint_id = ep.id,
                event       = "content.deleted",
                payload     = '{"id":"' .. (data.id or "") .. '"}',
                status      = "pending",
            })
        end
    end
end)

function on_init()
    db.define_table("endpoints", {
        columns = {
            { name = "url",    type = "text",    not_null = true },
            { name = "events", type = "text",    not_null = true },
            { name = "secret", type = "text",    default = "" },
            { name = "active", type = "boolean", not_null = true, default = 1 },
        },
        indexes = {
            { columns = {"active"} },
        },
    })

    db.define_table("deliveries", {
        columns = {
            { name = "endpoint_id", type = "text",    not_null = true },
            { name = "event",       type = "text",    not_null = true },
            { name = "payload",     type = "text",    not_null = true },
            { name = "status",      type = "text",    not_null = true, default = "pending" },
            { name = "attempts",    type = "integer", default = 0 },
            { name = "last_error",  type = "text" },
        },
        indexes = {
            { columns = {"status"} },
            { columns = {"endpoint_id"} },
            { columns = {"event"} },
        },
        foreign_keys = {
            {
                column     = "endpoint_id",
                ref_table  = "plugin_webhook_relay_endpoints",
                ref_column = "id",
                on_delete  = "CASCADE",
            },
        },
    })

    log.info("webhook_relay initialized")
end
```

> **Good to know**: Plugins cannot make outbound HTTP requests. This plugin logs delivery records to a database table. An external worker (cron job, queue consumer, or separate service) polls the `deliveries` table and performs the actual HTTP POST to registered endpoints.

---

## Analytics Logger

Wildcard after-hooks that log all content mutations across all tables. Demonstrates wildcard hooks, structured logging, and high-volume event tracking.

**Key patterns:** Wildcard hooks (`"*"` table), structured log context, activity feed with time-based queries.

```lua
plugin_info = {
    name        = "analytics_logger",
    version     = "1.0.0",
    description = "Log all content mutations for analytics",
    author      = "Example",
}

-- API: query activity feed
http.handle("GET", "/activity", function(req)
    local opts = {
        order_by = "created_at DESC",
        limit    = tonumber(req.query.limit) or 50,
        offset   = tonumber(req.query.offset) or 0,
    }

    if req.query.action then
        opts.where = { action = req.query.action }
    end
    if req.query.table_name then
        opts.where = opts.where or {}
        opts.where.table_name = req.query.table_name
    end

    local events = db.query("events", opts)
    local total = db.count("events", opts.where and { where = opts.where } or {})

    return {
        status = 200,
        json = { events = events, total = total },
    }
end)

-- API: get summary counts
http.handle("GET", "/summary", function(req)
    local creates = db.count("events", { where = { action = "created" } })
    local updates = db.count("events", { where = { action = "updated" } })
    local deletes = db.count("events", { where = { action = "deleted" } })
    local publishes = db.count("events", { where = { action = "published" } })
    local archives = db.count("events", { where = { action = "archived" } })

    return {
        status = 200,
        json = {
            created   = creates,
            updated   = updates,
            deleted   = deletes,
            published = publishes,
            archived  = archives,
            total     = creates + updates + deletes + publishes + archives,
        },
    }
end)

-- Wildcard after-hooks: log every mutation on every table
hooks.on("after_create", "*", function(data)
    db.insert("events", {
        action     = "created",
        table_name = data._table,
        entity_id  = data.id or "",
    })
    log.debug("Logged create event", { table_name = data._table, id = data.id })
end)

hooks.on("after_update", "*", function(data)
    db.insert("events", {
        action     = "updated",
        table_name = data._table,
        entity_id  = data.id or "",
    })
end)

hooks.on("after_delete", "*", function(data)
    db.insert("events", {
        action     = "deleted",
        table_name = data._table,
        entity_id  = data.id or "",
    })
    log.info("Entity deleted", { table_name = data._table, id = data.id })
end)

hooks.on("after_publish", "*", function(data)
    db.insert("events", {
        action     = "published",
        table_name = data._table,
        entity_id  = data.id or "",
    })
end)

hooks.on("after_archive", "*", function(data)
    db.insert("events", {
        action     = "archived",
        table_name = data._table,
        entity_id  = data.id or "",
    })
end)

function on_init()
    db.define_table("events", {
        columns = {
            { name = "action",     type = "text", not_null = true },
            { name = "table_name", type = "text", not_null = true },
            { name = "entity_id",  type = "text", not_null = true },
        },
        indexes = {
            { columns = {"action"} },
            { columns = {"table_name"} },
            { columns = {"action", "table_name"} },
        },
    })

    log.info("analytics_logger initialized")
end
```

Wildcard hooks (table `"*"`) fire for all CMS tables. At equal priority, table-specific hooks from other plugins run before wildcard hooks. This plugin uses only after-hooks, so it has full `db.*` access and runs asynchronously without blocking CMS operations.

---

## API Gateway

Middleware-based authentication with API key validation on public routes. Demonstrates `http.use` middleware, public routes, and key management.

**Key patterns:** `http.use` middleware for cross-cutting concerns, public routes with custom auth, API key CRUD.

```lua
local validators = require("validators")

plugin_info = {
    name        = "api_gateway",
    version     = "1.0.0",
    description = "API key management and validation",
    author      = "Example",
}

-- Middleware: validate API key on requests that include one
http.use(function(req)
    local key = req.headers["x-api-key"]
    if not key then
        return nil  -- no key provided, fall through to default auth
    end

    -- Check if key exists and is active
    local api_key = db.query_one("api_keys", {
        where = { key_value = key, active = 1 },
    })
    if not api_key then
        return { status = 401, json = { error = "invalid or inactive api key" } }
    end

    -- Update last-used timestamp
    db.update("api_keys", {
        set   = { last_used_at = db.timestamp(), use_count = api_key.use_count + 1 },
        where = { id = api_key.id },
    })

    return nil  -- key valid, continue to handler
end)

-- Public: validate a key without needing a CMS session
http.handle("POST", "/validate", function(req)
    if not req.json or not req.json.key then
        return { status = 400, json = { error = "key is required" } }
    end
    local api_key = db.query_one("api_keys", {
        where = { key_value = req.json.key, active = 1 },
    })
    if not api_key then
        return { status = 401, json = { error = "invalid" } }
    end
    return {
        status = 200,
        json = {
            valid      = true,
            name       = api_key.name,
            created_at = api_key.created_at,
        },
    }
end, { public = true })

-- Authenticated: list all API keys
http.handle("GET", "/keys", function(req)
    local keys = db.query("api_keys", {
        order_by = "created_at DESC",
        limit    = 100,
    })
    -- Strip the actual key values from the response
    for _, k in ipairs(keys) do
        k.key_value = k.key_value:sub(1, 8) .. "..."
    end
    return { status = 200, json = keys }
end)

-- Authenticated: create a new API key
http.handle("POST", "/keys", function(req)
    if not req.json or not req.json.name then
        return { status = 400, json = { error = "name is required" } }
    end

    -- Generate a random-looking key using ULIDs
    local key_value = "mck_" .. db.ulid() .. db.ulid()

    local id = db.ulid()
    db.insert("api_keys", {
        id           = id,
        name         = req.json.name,
        key_value    = key_value,
        active       = 1,
        use_count    = 0,
        last_used_at = "",
    })

    -- Return the full key value only on creation
    local created = db.query_one("api_keys", { where = { id = id } })
    return { status = 201, json = created }
end)

-- Authenticated: revoke a key
http.handle("DELETE", "/keys/{id}", function(req)
    if not db.exists("api_keys", { where = { id = req.params.id } }) then
        return { status = 404, json = { error = "not found" } }
    end
    db.update("api_keys", {
        set   = { active = 0 },
        where = { id = req.params.id },
    })
    return { status = 200, json = { revoked = true } }
end)

function on_init()
    db.define_table("api_keys", {
        columns = {
            { name = "name",         type = "text",    not_null = true },
            { name = "key_value",    type = "text",    not_null = true },
            { name = "active",       type = "boolean", not_null = true, default = 1 },
            { name = "use_count",    type = "integer", default = 0 },
            { name = "last_used_at", type = "text",    default = "" },
        },
        indexes = {
            { columns = {"key_value"}, unique = true },
            { columns = {"active"} },
        },
    })

    log.info("api_gateway initialized")
end
```

The `lib/validators.lua` module for this plugin:

```lua
-- lib/validators.lua
local M = {}

function M.not_empty(s)
    return type(s) == "string" and #s > 0
end

return M
```

The middleware runs before every route handler. It checks for an `x-api-key` header and validates it against the `api_keys` table. Routes without the header fall through to the CMS session authentication (for authenticated routes) or proceed directly (for public routes).

The `/validate` endpoint is public, allowing external services to verify API keys without a CMS session. The key management endpoints (`/keys`) require CMS authentication. The full key value is only returned on creation -- list endpoints mask it.
