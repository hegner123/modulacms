---@meta

--- Plugin manifest. Every plugin MUST define this global table.
---@class PluginInfo
---@field name string Unique identifier. Lowercase alphanumeric + underscores, max 32 chars.
---@field version string Semantic version (e.g., `"1.0.0"`). Route approvals reset on change.
---@field description string Short human-readable description.
---@field author? string Plugin author.
---@field license? string License identifier (e.g., `"MIT"`).
---@field min_cms_version? string Minimum ModulaCMS version required.
---@field dependencies? string[] Plugin names this plugin depends on (loaded first).

---@type PluginInfo
plugin_info = {}

--- Called once after the plugin is loaded and all VMs are created.
--- Use for schema setup (`db.define_table`), data seeding, and validation.
---
--- Available: `db.*`, `core.*`, `log.*`
--- NOT available: `http.handle()`, `hooks.on()`, `request.send()`
---
--- If this function calls `error()`, the plugin is marked as "failed".
function on_init() end

--- Called when the plugin is being stopped (server shutdown, admin disable,
--- or hot reload). Use for cleanup and logging final state.
---
--- If this function calls `error()`, the error is logged but shutdown continues.
--- Timeout: 10 seconds.
function on_shutdown() end

--- Sandboxed require. Loads modules from `<plugin_dir>/lib/<name>.lua`.
--- Path traversal (`..`, `/`, `\`) is rejected. Modules are cached after first load.
---@param modname string Module name (no path separators, no `.lua` extension).
---@return any module The value returned by the module file.
function require(modname) end
