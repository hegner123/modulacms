# Plan: Extend Query Builder for Plugin SQL Support

## Context

The query builder (`internal/db/query_builder.go`) currently supports only equality-based WHERE
clauses (`column = value` and `IS NULL`), single ORDER BY, and no JOINs. Third-party Lua plugins
need richer query capabilities to build real applications on plugin-scoped tables.

This plan is split into two tiers based on risk and demonstrated need:

- **Tier 1** (this plan): Single-table WHERE extensions -- comparison operators, OR/AND nesting,
  IN, BETWEEN, IS NULL/IS NOT NULL, LIKE, DISTINCT, multiple ORDER BY, bulk INSERT, and
  standalone index creation. These do not cross the table-boundary security line.

- **Tier 2** (deferred): JOINs, GROUP BY/HAVING, aggregates, table-qualified columns. These
  introduce cross-table access through the plugin boundary and require additional security
  validation. Tier 2 ships after Tier 1 is in production and plugin authors have requested it.

## Design Decisions

### Separate functions, not dual-path fields

Existing `SelectParams`, `UpdateParams`, `DeleteParams` and their functions (`QSelect`, `QUpdate`,
`QDelete`, `QCount`, `QExists`) are **frozen**. No fields are added to existing structs.

New structs and functions are added alongside:

```go
// FilteredSelectParams is the extended SELECT params using the Condition system.
// Existing QSelect/SelectParams remain untouched for the simple equality path.
type FilteredSelectParams struct {
    Table       string
    Columns     []string        // nil = SELECT *
    Filter      Condition       // required (replaces Where map)
    OrderByCols []OrderByColumn // empty = no ORDER BY
    Distinct    bool
    Limit       int64           // 0 = default (maxLimit); negative = no limit
    Offset      int64
}

type FilteredUpdateParams struct {
    Table  string
    Set    map[string]any // must be non-empty
    Filter Condition      // required (replaces Where map)
}

type FilteredDeleteParams struct {
    Table  string
    Filter Condition // required (replaces Where map)
}
```

New functions:

- `QSelectFiltered(ctx, exec, dialect, FilteredSelectParams) ([]Row, error)`
- `QSelectOneFiltered(ctx, exec, dialect, FilteredSelectParams) (Row, error)`
- `QUpdateFiltered(ctx, exec, dialect, FilteredUpdateParams) (sql.Result, error)`
- `QDeleteFiltered(ctx, exec, dialect, FilteredDeleteParams) (sql.Result, error)`
- `QCountFiltered(ctx, exec, dialect, table, Condition) (int64, error)`
- `QExistsFiltered(ctx, exec, dialect, table, Condition) (bool, error)`
- `QBulkInsert(ctx, exec, dialect, BulkInsertParams) (sql.Result, error)`

The existing code paths stay frozen. Zero risk of breaking existing behavior. Zero dual-path
confusion for future contributors.

### LIKE dialect behavior

LIKE behaves differently across databases:
- SQLite: case-insensitive for ASCII by default
- MySQL: case-insensitive with default collation
- PostgreSQL: case-sensitive (`ILIKE` is case-insensitive)

This plan does NOT normalize behavior. The Compare condition supports `LIKE` as-is per dialect.
Plugin documentation will prominently note the dialect difference. Rationale: normalizing would
require rewriting the operator per dialect (LIKE vs ILIKE) which adds hidden complexity. Plugin
authors who need case-insensitive matching should use `LOWER(col) LIKE LOWER(?)` patterns via
the existing equality path or accept dialect variance.

LIKE pattern safety: patterns are parameterized (the pattern itself is a `?` bind value), so
SQL injection is not possible. However, pathological patterns (`%a%b%c%d%e%f%`) can cause slow
scans. The query builder does NOT validate LIKE patterns -- this is the plugin author's
responsibility, and the existing operation budget + query timeout provide backpressure.

### BETWEEN type coercion

`BETWEEN` has type-sensitive behavior across dialects. `BETWEEN '2' AND '10'` does string
comparison, not numeric. If a plugin passes Lua numbers for a TEXT column:
- SQLite: permissive (type affinity handles it)
- MySQL: implicit conversion (usually works, sometimes surprising)
- PostgreSQL: strict type mismatch error

Plugin documentation should note that BETWEEN values must match the column's storage type.
No normalization is performed -- type mismatch errors are surfaced to the plugin as error
returns, which is the correct behavior.

---

## Phase 1: Condition System

**New file:** `internal/db/query_conditions.go`

### Condition Interface

```go
// Condition represents a WHERE clause expression.
// Security invariant: every Build implementation MUST:
// 1. Validate all column names via ValidColumnName
// 2. Quote all identifiers via quoteIdent
// 3. Use placeholder() for all user-provided values
// 4. Never embed raw string values into the SQL output
//
// Build accepts a BuildContext that tracks depth and total node count
// to prevent resource exhaustion from deeply nested or excessively wide
// condition trees.
type Condition interface {
    Build(ctx BuildContext, d Dialect, argOffset int) (sql string, args []any, nextOffset int, err error)
}

// BuildContext is NOT goroutine-safe. It must not be shared across concurrent
// calls to Build(). Each top-level query function creates its own BuildContext.
// This is safe in practice because each Lua VM is single-threaded (1:1 LState
// binding) and direct Go callers should create a fresh BuildContext per query.
type BuildContext struct {
    CurrentDepth int
    NodeCount    *int // pointer: shared across recursive calls within one tree
}
```

### Concrete Types

| Type             | SQL Output                            | Tier |
|------------------|---------------------------------------|------|
| Compare          | `"col" > ?` (=, <>, >, <, >=, <=, LIKE) | 1 |
| InCondition      | `"col" IN (?, ?, ?)`                  | 1    |
| BetweenCondition | `"col" BETWEEN ? AND ?`               | 1    |
| IsNull           | `"col" IS NULL`                       | 1    |
| IsNotNull        | `"col" IS NOT NULL`                   | 1    |
| And              | `(c1 AND c2 AND ...)`                 | 1    |
| Or               | `(c1 OR c2 OR ...)`                   | 1    |
| Not              | `NOT (condition)`                     | 1    |

### Not Struct

`Not` wraps a single `Condition`. To negate multiple conditions, wrap them in `And` or `Or` first.

```go
type Not struct {
    Condition Condition // single condition; wrap in And/Or for multiple
}
```

**Deferred to Tier 2:**
- `ColCompare` (`"t1"."col1" = "t2"."col2"` for JOIN ON clauses)
- Any type involving table-qualified column references

### Safety Limits

```go
const (
    MaxConditionDepth  = 10  // max nesting of And/Or/Not
    MaxConditionNodes  = 50  // total Condition nodes in a single tree
    MaxInValues        = 500 // max elements in a single IN clause
    MaxOrderByCols     = 8   // max columns in ORDER BY
    MaxBulkInsertRows  = 1000 // absolute max rows per QBulkInsert call
)
```

Every `Build` call:
1. Increments `*NodeCount` and checks against `MaxConditionNodes`
2. Checks `CurrentDepth` against `MaxConditionDepth` (And/Or/Not increment before recursing)
3. InCondition checks `len(Values)` against `MaxInValues` and rejects `len(Values) == 0`
   (`IN ()` is a syntax error in all three dialects)
4. Compare rejects nil values for all operators (nil comparisons are meaningless; use
   IsNull/IsNotNull instead). `WHERE "col" LIKE NULL` and `WHERE "col" > NULL` both
   produce `UNKNOWN` which silently matches zero rows.

### CompareOp Allowlist

```go
type CompareOp string

const (
    OpEq   CompareOp = "="
    OpNeq  CompareOp = "<>" // standard SQL; Lua API accepts both "!=" and "<>", Build emits "<>"
    OpGt   CompareOp = ">"
    OpLt   CompareOp = "<"
    OpGte  CompareOp = ">="
    OpLte  CompareOp = "<="
    OpLike CompareOp = "LIKE"
)

var validCompareOps = map[CompareOp]bool{
    OpEq: true, OpNeq: true, OpGt: true, OpLt: true,
    OpGte: true, OpLte: true, OpLike: true,
}

// The Lua condition parser normalizes "!=" to "<>" before constructing Compare.
```

### Validation Helper

```go
// ValidateCondition walks the condition tree and returns an error if any
// safety limit is exceeded, any column name is invalid, or any operator
// is not in the allowlist. This is the authoritative pre-flight check
// before Build(). Build() also enforces limits defensively (belt-and-
// suspenders), but callers rely on ValidateCondition for user-facing
// error messages. Both must agree on limits.
func ValidateCondition(c Condition) error

// HasValueBinding returns true if the condition tree contains at least one
// leaf node that binds a parameterized value (Compare, InCondition, or
// BetweenCondition). Used by QUpdateFiltered and QDeleteFiltered to reject
// structurally valid but semantically vacuous conditions like
// IsNotNull{Column: "id"} which would match all rows.
func HasValueBinding(c Condition) bool
```

**New file:** `internal/db/query_conditions_test.go`

Table-driven tests for every condition type across all 3 dialects:
- SQL output correctness
- Placeholder numbering (especially PostgreSQL `$N` sequences)
- Args ordering matches placeholder positions
- Validation errors (invalid column, invalid op, bad nesting)
- Depth limit enforcement (depth 11 fails)
- Node count limit enforcement (51 nodes fails)
- IN size limit enforcement (501 values fails)
- IN with empty values (error -- `IN ()` is invalid SQL)
- Compare with nil value (error for all operators)
- Empty And/Or children (error)
- Single-child And/Or (valid, still wrap parens for consistency)
- `HasValueBinding` returns true for trees with Compare/In/Between leaves
- `HasValueBinding` returns false for trees with only IsNull/IsNotNull leaves
- `OpNeq` emits `<>` not `!=` in SQL output

---

## Phase 2: Filtered Query Functions

**New file:** `internal/db/query_filtered.go`

### New Param Structs

```go
type OrderByColumn struct {
    Column string
    Desc   bool
}

type FilteredSelectParams struct {
    Table       string
    Columns     []string        // nil = SELECT *
    Filter      Condition       // required
    OrderByCols []OrderByColumn // empty = no ORDER BY
    Distinct    bool
    Limit       int64           // 0 = default (maxLimit); negative = no limit
    Offset      int64
}

type FilteredUpdateParams struct {
    Table  string
    Set    map[string]any // must be non-empty
    Filter Condition      // required
}

type FilteredDeleteParams struct {
    Table  string
    Filter Condition // required
}
```

### New Functions

```go
func QSelectFiltered(ctx context.Context, exec Executor, d Dialect, p FilteredSelectParams) ([]Row, error)
func QSelectOneFiltered(ctx context.Context, exec Executor, d Dialect, p FilteredSelectParams) (Row, error)
func QUpdateFiltered(ctx context.Context, exec Executor, d Dialect, p FilteredUpdateParams) (sql.Result, error)
func QDeleteFiltered(ctx context.Context, exec Executor, d Dialect, p FilteredDeleteParams) (sql.Result, error)
func QCountFiltered(ctx context.Context, exec Executor, d Dialect, table string, filter Condition) (int64, error)
func QExistsFiltered(ctx context.Context, exec Executor, d Dialect, table string, filter Condition) (bool, error)
```

`QSelectOneFiltered` forces `p.Limit = 1` internally before executing, matching the behavior of
existing `QSelectOne`. Callers do not need to set Limit themselves.

Each function:
1. Validates table name via `ValidTableName`
2. Calls `ValidateCondition(filter)` as pre-flight
3. Calls `filter.Build(...)` to produce the WHERE clause
4. Assembles the full query using the same `quoteIdent`/`placeholder` helpers as existing code
5. Executes via `exec.QueryContext` / `exec.ExecContext`

`QSelectFiltered` validates `len(OrderByCols) <= MaxOrderByCols`.

`QExistsFiltered` uses `SELECT 1 ... LIMIT 1` (same pattern as existing `QExists`), not
`COUNT(*)`.

`QUpdateFiltered` and `QDeleteFiltered` require:
- Non-nil Filter (same safety as existing "non-empty where" rule)
- `HasValueBinding(filter)` must return true -- rejects structurally valid but semantically
  vacuous conditions (e.g., `IsNotNull{Column: "id"}` matches all rows on a NOT NULL column).
  This prevents accidental full-table mutations from plugin bugs.

### PostgreSQL Placeholder Offset Threading

`QUpdateFiltered` must thread the `argOffset` through the SET clause before passing it to
`filter.Build()`. The SET clause consumes `$1..$N` and the condition must start at `$N+1`:

```
-- 2 SET columns, 2 WHERE conditions:
UPDATE "plugin_tracker_tasks" SET "status" = $1, "priority" = $2
WHERE ("name" = $3 AND "active" = $4)
```

Implementation: `argOffset` starts at 1, SET clause increments it to `len(setArgs) + 1`,
then that value is passed as `argOffset` to `filter.Build()`. Test plan must include a
PostgreSQL-dialect test case for `QUpdateFiltered` with multiple SET columns AND a compound
condition to verify `$N` numbering is contiguous with no gaps or overlaps.

### Bulk Insert

```go
type BulkInsertParams struct {
    Table   string
    Columns []string
    Rows    [][]any // each inner slice has len(Columns) values; nil elements = SQL NULL
}

func QBulkInsert(ctx context.Context, exec Executor, d Dialect, p BulkInsertParams) (sql.Result, error)
```

**Dynamic batch sizing per dialect:**

The per-statement row limit is calculated dynamically based on column count to respect
placeholder/variable constraints:

- **SQLite:** `min(floor(999 / len(Columns)), 100)` -- SQLite's default
  `SQLITE_MAX_VARIABLE_NUMBER` is 999. A 10-column table allows 99 rows/batch; a 64-column
  table allows only 15 rows/batch. The fixed cap of 100 prevents overshooting even with
  few columns.
- **MySQL:** `min(500, dialect cap)` -- conservative against default 4MB `max_allowed_packet`.
  Column count is less of a constraint here, but 500 is a reasonable ceiling.
- **PostgreSQL:** `min(1000, dialect cap)` -- no hard variable limit, but `$N` numbering
  and query planning degrade with huge statements. 1000 rows is the ceiling.

**Transaction wrapping:**

When `len(Rows)` exceeds the per-statement batch size, `QBulkInsert` executes multiple INSERT
statements. To prevent partial writes:

- If `exec` is a `*sql.DB` (not already in a transaction): `QBulkInsert` wraps all batches
  in a single transaction. If any batch fails, the entire operation rolls back.
- If `exec` is a `*sql.Tx` (already in a transaction): batches execute within the existing
  transaction. Caller controls commit/rollback.

Detection uses a type assertion: `if _, ok := exec.(*sql.Tx); ok { ... }`.

**Fallback for unknown Executor types:** If `exec` is neither `*sql.Tx` nor `*sql.DB` (e.g., a
mock in unit tests) and the insert requires multiple batches, return an error:
`"QBulkInsert with multiple batches requires *sql.DB or *sql.Tx, got %T"`. Single-batch
operations do not need transaction wrapping and work with any `Executor`.

`RowsAffected` is accumulated across all batches and returned in a composite `sql.Result`.

**NULL handling:** nil values in `Rows[][]any` are passed through to the database driver as
SQL NULL, consistent with single-row `QInsert`. This is the standard `database/sql` behavior.

**Validation:**
- `len(Columns)` must be > 0
- Every row must have exactly `len(Columns)` values
- All column names validated via `ValidColumnName`
- `len(Rows)` must be > 0
- `len(Rows)` must be <= `MaxBulkInsertRows` (1000) -- hard cap regardless of dialect

**New file:** `internal/db/query_filtered_test.go`

Tests for all new functions:
- `QSelectFiltered` with Compare, In, Between, IsNull, IsNotNull, And, Or, Not
- `QSelectFiltered` with DISTINCT
- `QSelectFiltered` with multiple OrderByCols
- `QSelectFiltered` rejects `len(OrderByCols) > MaxOrderByCols`
- `QUpdateFiltered` with compound conditions
- `QUpdateFiltered` PostgreSQL dialect: verify `$N` numbering is contiguous across SET + WHERE
- `QUpdateFiltered` rejects vacuous condition (IsNotNull-only, no value binding)
- `QDeleteFiltered` with compound conditions
- `QDeleteFiltered` rejects vacuous condition
- `QCountFiltered` and `QExistsFiltered` with conditions
- `QExistsFiltered` generates `SELECT 1 ... LIMIT 1` (not COUNT)
- `QBulkInsert` single batch, multi-batch (exceed limit), validation errors
- `QBulkInsert` dynamic batch sizing: 64-column table on SQLite gets batches of 15 (floor(999/64))
- `QBulkInsert` with nil values in rows (SQL NULL)
- `QBulkInsert` transaction wrapping: multi-batch with second batch failing rolls back first
- `QBulkInsert` rejects `len(Rows) > MaxBulkInsertRows`
- `QBulkInsert` accumulates RowsAffected across batches
- Backward compat: existing `QSelect`/`QUpdate`/`QDelete` tests still pass (no changes to them)

---

## Phase 3: Lua Condition Parser

**New file:** `internal/plugin/db_conditions.go`

### parseConditionFromLua

`parseConditionFromLua(L *lua.LState, tbl *lua.LTable) (db.Condition, error)`

Recursive parser that converts Lua tables into `db.Condition` types:

```lua
-- Simple comparison (3-element sequence)
{"priority", ">", 5}

-- AND (implicit: sequence of conditions)
{{"status", "=", "active"}, {"priority", ">", 3}}

-- OR / AND / NOT (explicit keyword as first element)
{"OR", {{"status", "=", "active"}, {"status", "=", "draft"}}}
{"NOT", {"status", "=", "deleted"}}

-- IN (value is a sequence)
{"status", "IN", {"active", "draft", "review"}}

-- BETWEEN (value is a 2-element sequence)
{"priority", "BETWEEN", {1, 10}}

-- IS NULL / IS NOT NULL (2-element, no value)
{"description", "IS NULL"}
{"description", "IS NOT NULL"}

-- Nested
{"OR", {
    {"AND", {{"status", "=", "active"}, {"priority", ">", 5}}},
    {"AND", {{"status", "=", "urgent"}, {"priority", ">", 0}}},
}}
```

### Detection Logic

**Migration from existing `parseWhereFromLua`:**

The existing `parseWhereFromLua` in `db_api.go` returns a single `map[string]any` with no error.
It is called at six sites: `luaQuery`, `luaQueryOne`, `luaCount`, `luaExists`, `luaUpdate`,
`luaDelete`.

Create a new function `parseWhereExtended` in `db_conditions.go`:

```go
// parseWhereExtended replaces parseWhereFromLua with condition-aware detection.
// Returns exactly one of (whereMap, condition) as non-nil, or both nil for no filter.
func parseWhereExtended(L *lua.LState, optsTbl *lua.LTable) (map[string]any, db.Condition, error)
```

The existing `parseWhereFromLua` in `db_api.go` is deleted. All six call sites must be updated
to call `parseWhereExtended` and handle the error return:

```go
// Before (all six sites):
where = parseWhereFromLua(L, optsTbl)

// After (all six sites):
whereMap, condition, err := parseWhereExtended(L, optsTbl)
if err != nil {
    L.Push(lua.LNil)
    L.Push(lua.LString(err.Error()))
    return 2
}
```

Detection rules inside `parseWhereExtended`:

1. Get `"where"` field from opts table
2. If nil/absent: return `(nil, nil, nil)` -- no filter
3. If it's a table with **only string keys**: old path, return `(map, nil, nil)`
4. If it's a table that is a **pure sequence** (consecutive integer keys 1..N with no string
   keys): new path, parse as conditions, return `(nil, condition, nil)`
5. If it's a **mixed table** (both string and integer keys): return error with a helpful
   message: `"where table cannot mix string keys (equality format) and sequence entries`
   `(condition format); use either {status = 'active'} or {{'status', '=', 'active'}}"`

This uses the same detection logic as the `luaTableToGoValue` function in `lua_helpers.go`
(check `hasStringKey`, `intKeyCount`, `maxIntKey`).

### Additional Parser

```go
// OrderByResult holds the parsed order_by configuration.
// Exactly one of Column or Columns is non-zero.
type OrderByResult struct {
    Column  string             // old-style single column (empty = not set)
    Desc    bool               // old-style direction (only meaningful when Column is set)
    Columns []db.OrderByColumn // new-style multi-column (nil = not set)
}

// parseOrderByFromLua handles both old-style string and new-style table:
//   order_by = "name"                         -- old: single column ASC
//   order_by = "name", desc = true            -- old: single column DESC
//   order_by = { {column="name", desc=true} } -- new: multiple columns
func parseOrderByFromLua(L *lua.LState, optsTbl *lua.LTable) (OrderByResult, error)
```

Caller-side branching in `luaQuery`:

```go
ob, err := parseOrderByFromLua(L, optsTbl)
if err != nil { /* return error */ }

if ob.Columns != nil {
    // new-style: must use QSelectFiltered path
    // set FilteredSelectParams.OrderByCols = ob.Columns
} else if ob.Column != "" {
    // old-style: set SelectParams.OrderBy = ob.Column, SelectParams.Desc = ob.Desc
}
```

**New file:** `internal/plugin/db_conditions_test.go`

Tests using a real gopher-lua VM:
- Every operator through Lua tables
- Nested AND/OR/NOT
- `parseWhereExtended` detection: string-key table -> old path returns `(map, nil, nil)`
- `parseWhereExtended` detection: sequence table -> new path returns `(nil, condition, nil)`
- `parseWhereExtended` detection: mixed table -> error with message containing both format examples
- `"!="` in Lua normalized to `<>` in built condition
- Invalid operators rejected
- Column validation through the parser
- `parseOrderByFromLua` old-style: `order_by = "name"` returns `OrderByResult{Column: "name"}`
- `parseOrderByFromLua` old-style with desc: returns `OrderByResult{Column: "name", Desc: true}`
- `parseOrderByFromLua` new-style: returns `OrderByResult{Columns: [...]}`

---

## Phase 4: Extend Existing Lua API + New Functions

**Modify:** `internal/plugin/db_api.go`

### Changes to Existing Functions

`luaQuery`, `luaQueryOne`, `luaUpdate`, `luaDelete`, `luaCount`, `luaExists`:

1. Call `parseWhereExtended` (replaces old `parseWhereFromLua`) which returns `(map, condition, err)`
2. Handle the error return (see Phase 3 for the error pattern)
3. If `condition != nil`: call the new `QSelectFiltered` / `QUpdateFiltered` / etc.
4. If `map != nil`: call existing `QSelect` / `QUpdate` / etc. (unchanged path)
5. If both nil: existing behavior (no filter)

`luaQuery` and `luaQueryOne` additionally:
- Call `parseOrderByFromLua` for the `order_by` field
- If new-style returned (`ob.Columns != nil`): use `QSelectFiltered` with `OrderByCols`
- If old-style returned (`ob.Column != ""`): pass both `OrderBy` and `Desc` to `SelectParams`.
  The existing `luaQuery` never passed `Desc` to `SelectParams`; this is fixed as part of Phase 4.
  Old-style Lua syntax: `order_by = "name"` with optional `desc = true` field in the opts table.

`luaQuery` and `luaQueryOne` also parse these fields for the filtered path:
- `columns`: Lua sequence of column name strings, e.g., `columns = {"name", "status", "priority"}`.
  Maps to `FilteredSelectParams.Columns`. If absent or nil, defaults to `SELECT *`.
- `distinct`: Lua boolean, e.g., `distinct = true`. Maps to `FilteredSelectParams.Distinct`.
  If absent, defaults to false. Only valid on the filtered path (condition-style where);
  if `distinct = true` is passed with old-style string-key where, it is silently ignored
  (the old `QSelect` path does not support it).

`luaUpdate` auto-injects `updated_at` into the set map if not already provided by the plugin,
regardless of whether the old map-based or new condition-based path is used. This preserves the
existing behavior from the old path.

`luaUpdate` and `luaDelete` vacuous-condition rejection (when `HasValueBinding(condition)` returns
false) returns as `nil, errmsg` (not `L.RaiseError`), consistent with the db-layer error pattern.
Error message: `"update/delete requires a where condition with at least one value binding (e.g., {{'status', '=', 'active'}})"`.

This means **existing Lua code using string-key where tables continues to work unchanged**.
New condition syntax is opt-in.

### New Lua Functions

**New file:** `internal/plugin/db_api_bulk.go`

```go
// luaInsertMany implements db.insert_many(table, columns, rows)
// Auto-injects id, created_at, updated_at per row.
func (api *DatabaseAPI) luaInsertMany(L *lua.LState) int
```

Behavior:
- Parses columns from Lua sequence of strings
- Parses rows from Lua sequence of sequences
- Auto-prepends "id", "created_at", "updated_at" to columns list
- Auto-prepends `types.NewULID().String()`, `now`, `now` to each row's values
- Rejects `len(rows) > MaxBulkInsertRows` (1000) before calling `db.QBulkInsert`
- Calls `db.QBulkInsert`
- Operation budget: counts as `ceil(len(rows) / batchSize)` operations against `checkOpLimit`,
  where `batchSize` is the dialect-specific batch size. This ensures that bulk inserts consume
  op budget proportional to the actual database work performed. A 10-row insert costs 1 op;
  a 1000-row insert on SQLite with 10-column tables costs 11 ops (1000/99 = ~11 batches).
  Budget enforcement strategy: call `checkOpLimit()` once for the initial validation, then set
  `api.opCount += ceil(rows/batchSize) - 1` after the initial check succeeds but **before**
  executing any batches. This ensures the entire bulk insert is rejected upfront if it would
  exceed the budget, rather than failing partway through with partial data committed.

```go
// luaCreateIndex implements db.create_index(table, opts)
// Exposes existing DDLCreateIndex.
func (api *DatabaseAPI) luaCreateIndex(L *lua.LState) int
```

### Registration

In `RegisterDBAPI`:
```go
dbTable.RawSetString("insert_many", L.NewFunction(api.luaInsertMany))
dbTable.RawSetString("create_index", L.NewFunction(api.luaCreateIndex))
```

No new `db.query_advanced` function in Tier 1. The existing `db.query` and `db.query_one`
automatically support the new condition syntax when plugins pass sequence-style where tables.

---

## Phase 5: Full Test Suite

**New file:** `internal/plugin/db_api_filtered_test.go`

Integration tests using real gopher-lua VMs + SQLite:

- All WHERE operators through Lua: `=`, `!=` (normalized to `<>`), `>`, `<`, `>=`, `<=`,
  `LIKE`, `IN`, `BETWEEN`, `IS NULL`, `IS NOT NULL`
- OR/AND/NOT nesting through `db.query`
- Multiple ORDER BY through `db.query`
- DISTINCT through `db.query` with `distinct = true` in opts
- Column selection through `db.query` with `columns = {"name", "status"}` in opts
- `db.update` with condition-style where rejects vacuous conditions (returns `nil, errmsg`)
- `db.delete` with condition-style where rejects vacuous conditions (returns `nil, errmsg`)
- `db.update` auto-injects `updated_at` on filtered path (same as old path)
- `db.insert_many` with auto-injected columns
- `db.insert_many` row count validation (rejects > 1000)
- `db.insert_many` with nil values in rows
- `db.insert_many` op budget: costs ceil(rows/batchSize) ops, not 1
- `db.create_index` standalone
- Backward compatibility: `db.query` with old string-key where format still works
- Depth limit enforcement through Lua
- Node count limit enforcement through Lua
- IN size limit enforcement through Lua
- IN with empty values (error)
- Mixed where table format (error with helpful message)

---

## Dependency Graph

```
Phase 1 (Condition types)
    │
    v
Phase 2 (Filtered query functions + bulk insert)
    │
    v
Phase 3 (Lua condition parser)
    │
    v
Phase 4 (Extend existing Lua API + new functions)
    │
    v
Phase 5 (Full test suite)
```

Phases 1-2 are Go-only (no Lua). Phases 3-5 are Lua integration.

---

## Files Summary

| File | Action | Purpose |
|------|--------|---------|
| `internal/db/query_conditions.go` | Create | Condition interface, concrete types, BuildContext, ValidateCondition |
| `internal/db/query_conditions_test.go` | Create | Condition unit tests across all 3 dialects |
| `internal/db/query_filtered.go` | Create | FilteredSelectParams, QSelectFiltered, QBulkInsert, etc. |
| `internal/db/query_filtered_test.go` | Create | Filtered query function tests |
| `internal/plugin/db_conditions.go` | Create | parseWhereExtended, parseConditionFromLua, parseOrderByFromLua, OrderByResult |
| `internal/plugin/db_conditions_test.go` | Create | Lua parser tests |
| `internal/plugin/db_api_bulk.go` | Create | insert_many, create_index Lua functions |
| `internal/plugin/db_api.go` | Modify | Delete parseWhereFromLua, update 6 call sites to parseWhereExtended, add error handling, integrate parseOrderByFromLua, pass Desc on old path, parse columns/distinct, registration |
| `internal/plugin/db_api_filtered_test.go` | Create | Full Lua-to-SQL integration tests |

**Not modified:** `internal/db/query_builder.go` -- existing code stays frozen.

---

## Verification

1. `go test ./internal/db/ -run TestCondition` -- condition types produce correct SQL per dialect
2. `go test ./internal/db/ -run TestFiltered` -- filtered query functions work correctly
3. `go test ./internal/db/ -run TestBulkInsert` -- multi-row insert with auto-batching
4. `go test ./internal/plugin/ -run TestParseCondition` -- Lua parser produces correct conditions
5. `go test ./internal/plugin/ -run TestLuaFiltered` -- full Lua API integration
6. `go test ./internal/plugin/` -- all existing tests still pass (backward compat)
7. `go test ./internal/db/` -- all existing tests still pass
8. `just check` -- compile check passes

---

## Tier 2 (Deferred)

The following features are explicitly deferred until Tier 1 is production-tested and plugin
authors have demonstrated need:

- **JOINs** -- `JoinClause`, `ColCompare`, table-qualified columns, cross-table prefix validation
- **GROUP BY / HAVING** -- aggregate functions (COUNT, SUM, AVG, MIN, MAX), `SelectColumn` with
  aggregate/alias, HAVING clause
- **db.query_advanced** -- the full-featured Lua function with JOINs, GROUP BY, aggregates
- **Table aliases** -- alias management and validation against plugin prefix

These features require:
1. Security specification for table alias validation (can aliases shadow core tables?)
2. Cross-table prefix enforcement in JOIN ON clauses
3. Aggregate column validation (are aggregate expressions user-controlled?)
4. Query complexity budget (a single query_advanced with 5 JOINs can be far more expensive
   than 100 simple queries -- should it count differently against the operation budget?)
5. Testing against real MySQL and PostgreSQL instances (not just SQLite)
6. EXPLAIN-based guardrails: ability to reject plugin queries with estimated full-table scans
   on tables above a configurable row threshold (advisory, not blocking Tier 2 itself)
