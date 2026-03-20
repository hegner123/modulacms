---@meta

--- Event hooks module. Registers handlers for CMS database mutation events.
--- Must be called at MODULE SCOPE (not inside `on_init()`).
---@class hooks
hooks = {}

---@alias hooks.Event
---| "before_create"   # Fires before a row is inserted. Can abort via `error()`.
---| "after_create"    # Fires after a row is inserted. Fire-and-forget.
---| "before_update"   # Fires before a row is updated. Can abort via `error()`.
---| "after_update"    # Fires after a row is updated. Fire-and-forget.
---| "before_delete"   # Fires before a row is deleted. Can abort via `error()`.
---| "after_delete"    # Fires after a row is deleted. Fire-and-forget.
---| "before_publish"  # Fires before content is published. Can abort via `error()`.
---| "after_publish"   # Fires after content is published. Fire-and-forget.

---@class hooks.EventData
---@field _table string The CMS table name that triggered the event.
---@field _event string The event name.
---@field [string] any Entity fields from the CMS record (varies by table schema).

---@class hooks.Opts
---@field priority? integer Execution order (1-1000, lower runs first, default 100).

--- Register a hook handler for a CMS mutation event.
--- Max 50 hooks per plugin.
---
--- **before-hooks:** Run synchronously inside the CMS transaction.
--- Calling `error()` aborts the mutation. `db.*` calls are BLOCKED (deadlock prevention).
--- Timeout: 2000ms per hook, 5000ms per event chain.
---
--- **after-hooks:** Run asynchronously (fire-and-forget) with reduced operation
--- budget (100 ops). Cannot abort. `db.*` calls are allowed.
---
--- Circuit breaker: 10 consecutive errors auto-disables the hook until plugin reload.
---@param event hooks.Event
---@param table_name string CMS table name (e.g., `"content_data"`) or `"*"` for wildcard.
---@param handler fun(data: hooks.EventData)
---@param opts? hooks.Opts
function hooks.on(event, table_name, handler, opts) end
