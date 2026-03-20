---@meta

--- Plugin database module. Provides full CRUD for plugin-owned tables.
--- Table names are automatically prefixed with `plugin_<name>_`.
---@class db
db = {}

---@class db.QueryOpts
---@field columns? string[] Column names to select (default: all).
---@field where? table<string, any> AND conditions as `{col = value}`.
---@field where_or? table<string, any>[] OR conditions as array of where-maps.
---@field filter? db.Condition New-style condition from `db.eq()`, `db.gt()`, etc.
---@field order_by? string|db.OrderBy[] Column name or array of `{column, desc}`.
---@field desc? boolean Legacy direction for string `order_by`.
---@field limit? integer Max rows returned (default 100, max 10000).
---@field offset? integer Pagination offset.
---@field group_by? string Column to GROUP BY.
---@field having? table<string, db.Condition> HAVING clause conditions.
---@field distinct? boolean SELECT DISTINCT.
---@field joins? db.Join[] Array of join definitions.
---@field prefix? string Table alias.

---@class db.OrderBy
---@field column string
---@field desc? boolean

---@class db.Join
---@field table string Table name to join.
---@field on string Join condition expression.

---@class db.UpdateOpts
---@field set table<string, any> Fields to update (required).
---@field where table<string, any> AND conditions (required, cannot be empty).
---@field where_or? table<string, any>[] OR conditions.

---@class db.DeleteOpts
---@field where table<string, any> AND conditions (required, cannot be empty).
---@field where_or? table<string, any>[] OR conditions.

---@class db.ColumnDef
---@field name string Column name.
---@field type "text"|"integer"|"boolean"|"real"|"timestamp"|"json"|"blob" SQL type.
---@field not_null? boolean
---@field unique? boolean
---@field default? any Default value.
---@field primary_key? boolean

---@class db.IndexDef
---@field columns string[] Columns to index.
---@field unique? boolean

---@class db.TableDef
---@field columns db.ColumnDef[]
---@field indexes? db.IndexDef[]

--- Opaque condition object returned by `db.eq()`, `db.gt()`, etc.
---@class db.Condition

--- Query rows from a plugin-owned table.
--- Always returns a table (empty if no results, never nil).
---@param table_name string Short table name (auto-prefixed).
---@param opts? db.QueryOpts Query options.
---@return table[] rows Array of row tables with column-name keys.
function db.query(table_name, opts) end

--- Query a single row from a plugin-owned table.
---@param table_name string Short table name (auto-prefixed).
---@param opts? db.QueryOpts Query options.
---@return table|nil row Single row table, or nil if not found.
function db.query_one(table_name, opts) end

--- Count rows in a plugin-owned table.
---@param table_name string Short table name (auto-prefixed).
---@param opts? db.QueryOpts Query options (where/filter apply).
---@return integer count
function db.count(table_name, opts) end

--- Check if any rows exist matching the query.
---@param table_name string Short table name (auto-prefixed).
---@param opts? db.QueryOpts Query options (where/filter apply).
---@return boolean exists
function db.exists(table_name, opts) end

--- Insert a row into a plugin-owned table.
--- Columns `id`, `created_at`, `updated_at` are auto-injected if not provided.
---@param table_name string Short table name (auto-prefixed).
---@param values table<string, any> Column values to insert.
---@return string id The auto-generated or provided row ID.
function db.insert(table_name, values) end

--- Insert multiple rows in a single operation.
---@param table_name string Short table name (auto-prefixed).
---@param rows table<string, any>[] Array of row value tables.
function db.insert_many(table_name, rows) end

--- Insert or update a row based on key columns.
---@param table_name string Short table name (auto-prefixed).
---@param values table<string, any> Column values.
---@param key_columns string[] Columns that form the unique key for conflict detection.
function db.upsert(table_name, values, key_columns) end

--- Update rows in a plugin-owned table.
--- Requires non-empty `where` (safety: prevents full-table updates).
---@param table_name string Short table name (auto-prefixed).
---@param opts db.UpdateOpts Must include `set` and `where`.
function db.update(table_name, opts) end

--- Delete rows from a plugin-owned table.
--- Requires non-empty `where` (safety: prevents full-table deletes).
---@param table_name string Short table name (auto-prefixed).
---@param opts db.DeleteOpts Must include `where`.
function db.delete(table_name, opts) end

--- Create an index on a plugin-owned table.
---@param table_name string Short table name (auto-prefixed).
---@param index_name string Index identifier.
---@param columns string[] Columns to index.
---@param unique? boolean Whether the index enforces uniqueness.
function db.create_index(table_name, index_name, columns, unique) end

--- Execute a function inside a database transaction.
--- If the callback calls `error()`, the transaction is rolled back.
---@param fn fun() Callback executed atomically.
---@return boolean ok True if committed successfully.
---@return string|nil err Error message if rolled back.
function db.transaction(fn) end

--- Generate a new 26-character ULID string (time-sortable, unique).
---@return string ulid
function db.ulid() end

--- Return the current UTC time as an RFC 3339 / ISO 8601 string.
---@return string timestamp
function db.timestamp() end

--- Define (CREATE IF NOT EXISTS) a plugin-owned table.
--- Columns `id` (TEXT PK), `created_at` (TEXT), and `updated_at` (TEXT) are
--- auto-injected. Do NOT include them in `columns` — doing so raises an error.
---@param table_name string Short table name (auto-prefixed).
---@param def db.TableDef Table definition with columns and optional indexes.
function db.define_table(table_name, def) end

-- Condition constructors -------------------------------------------------------

--- Equals condition.
---@param value any
---@return db.Condition
function db.eq(value) end

--- Not-equals condition.
---@param value any
---@return db.Condition
function db.neq(value) end

--- Greater-than condition.
---@param value number|string
---@return db.Condition
function db.gt(value) end

--- Greater-than-or-equal condition.
---@param value number|string
---@return db.Condition
function db.gte(value) end

--- Less-than condition.
---@param value number|string
---@return db.Condition
function db.lt(value) end

--- Less-than-or-equal condition.
---@param value number|string
---@return db.Condition
function db.lte(value) end

--- SQL LIKE pattern condition.
---@param pattern string LIKE pattern (e.g., `"%search%"`).
---@return db.Condition
function db.like(pattern) end

--- Negated SQL LIKE pattern condition.
---@param pattern string
---@return db.Condition
function db.not_like(pattern) end

--- SQL IN condition.
---@param ... any Values for the IN list.
---@return db.Condition
function db.in_list(...) end

--- SQL NOT IN condition.
---@param ... any Values for the NOT IN list.
---@return db.Condition
function db.not_in(...) end

--- SQL BETWEEN condition.
---@param low number|string Lower bound (inclusive).
---@param high number|string Upper bound (inclusive).
---@return db.Condition
function db.between(low, high) end

--- IS NULL condition.
---@return db.Condition
function db.is_null() end

--- IS NOT NULL condition.
---@return db.Condition
function db.is_not_null() end

-- Aggregate functions (for use in group_by queries) ----------------------------

--- COUNT aggregate.
---@param column? string Column name (omit for COUNT(*)).
---@return db.Condition
function db.agg_count(column) end

--- SUM aggregate.
---@param column string
---@return db.Condition
function db.sum(column) end

--- AVG aggregate.
---@param column string
---@return db.Condition
function db.avg(column) end

--- MIN aggregate.
---@param column string
---@return db.Condition
function db.min(column) end

--- MAX aggregate.
---@param column string
---@return db.Condition
function db.max(column) end
