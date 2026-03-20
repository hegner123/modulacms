---@meta

--- Core CMS table API. Provides gated read/write access to CMS-owned tables.
--- Access is controlled by a hardcoded whitelist, per-table read/write policy,
--- and the plugin's `approved_access` column in the database.
---@class core
core = {}

---@alias core.TableName
---| "content_data"     # Read + write
---| "content_fields"   # Read + write
---| "datatypes"        # Read-only
---| "fields"           # Read-only
---| "tables"           # Read-only
---| "media"            # Read + write
---| "routes"           # Read-only
---| "users"            # Read-only (hash column blocked)
---| "roles"            # Read-only
---| "permissions"      # Read-only
---| "role_permissions"  # Read-only
---| "change_events"    # Read-only

---@class core.QueryOpts
---@field columns? string[] Column names to select (default: all allowed).
---@field where? table<string, any> AND conditions.
---@field where_or? table<string, any>[] OR conditions.
---@field filter? db.Condition New-style condition from `db.eq()`, etc.
---@field order_by? string|table[] Column or `{column, desc}` array.
---@field desc? boolean Legacy direction for string `order_by`.
---@field limit? integer Max rows (default 100, max 10000).
---@field offset? integer Pagination offset.

---@class core.UpdateOpts
---@field set table<string, any> Fields to update (required).
---@field where table<string, any> AND conditions (required).
---@field where_or? table<string, any>[] OR conditions.

---@class core.DeleteOpts
---@field where table<string, any> AND conditions (required).
---@field where_or? table<string, any>[] OR conditions.

--- Query rows from a core CMS table.
--- Blocked columns (e.g., `users.hash`) are silently excluded.
---@param table_name core.TableName
---@param opts? core.QueryOpts
---@return table[] rows
function core.query(table_name, opts) end

--- Query a single row from a core CMS table.
---@param table_name core.TableName
---@param opts? core.QueryOpts
---@return table|nil row
function core.query_one(table_name, opts) end

--- Count rows in a core CMS table.
---@param table_name core.TableName
---@param opts? core.QueryOpts
---@return integer count
function core.count(table_name, opts) end

--- Check if any rows exist in a core CMS table.
---@param table_name core.TableName
---@param opts? core.QueryOpts
---@return boolean exists
function core.exists(table_name, opts) end

--- Insert a row into a writable core CMS table.
--- Only allowed on tables with write access: `content_data`, `content_fields`, `media`.
---@param table_name core.TableName
---@param values table<string, any>
function core.insert(table_name, values) end

--- Update rows in a writable core CMS table.
--- Only allowed on tables with write access.
---@param table_name core.TableName
---@param opts core.UpdateOpts
function core.update(table_name, opts) end

--- Delete rows from a writable core CMS table.
--- Only allowed on tables with write access.
---@param table_name core.TableName
---@param opts core.DeleteOpts
function core.delete(table_name, opts) end
