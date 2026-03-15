# Aggregate Expressions in Query Builder

## Goal

Support aggregate expressions (`COUNT(*)`, `SUM(col)`, `AVG(col)`, `MIN(col)`, `MAX(col)`) in
SELECT columns for both the Go query builder and the plugin Lua API.

**Target query:** `SELECT COUNT(*), status FROM tasks GROUP BY status`

## Current State

- `SelectParams.Columns []string` — validated by `ValidColumnName()` (alphanumeric + underscore only)
- `FilteredSelectParams.Columns []string` — same validation, unqualified only
- `parseColumnsFromLua` — expects Lua string sequence, auto-prefixes qualified columns
- `FilteredSelectParams` has **no GroupBy or Having** — those only exist in legacy `SelectParams`
- `DB_QUERY_DONE.md` deferred aggregates to "Tier 2"

## Design

### New type: `AggregateColumn`

```go
// AggregateColumn represents an aggregate function call in a SELECT clause.
type AggregateColumn struct {
    Func  string // COUNT, SUM, AVG, MIN, MAX (validated against allowlist)
    Arg   string // "*" or a valid column name
    Alias string // optional AS alias (validated as identifier)
}
```

Allowlist: `COUNT`, `SUM`, `AVG`, `MIN`, `MAX` — hardcoded, case-insensitive on input, stored uppercase.

### Additive field on both param structs

```go
// SelectParams — add field:
Aggregates []AggregateColumn // aggregate expressions; appended after Columns in SELECT

// FilteredSelectParams — add fields:
Aggregates []AggregateColumn // aggregate expressions; appended after Columns in SELECT
GroupBy    []string          // GROUP BY column names (validated)
Having     Condition         // HAVING condition (uses Condition system, not map)
```

**Column ordering:** Plain columns first, then aggregates. This is deterministic and predictable.
If the caller wants aggregate-only (`SELECT COUNT(*) FROM ...`), they leave `Columns` nil — the
builder uses only `Aggregates` (no `*` fallback when aggregates are present).

**Non-breaking guarantee:** The zero value of `[]AggregateColumn` is nil. Existing callers that
do not set `Aggregates` get exactly the current behavior — `len(Columns) == 0 && len(Aggregates) == 0`
falls through to `SELECT *`. No existing code path changes.

**Column count cap:** `len(Columns) + len(Aggregates)` must not exceed 64. This is a new
constraint for SELECT queries (the existing `maxColumns = 64` constant in `query_builder.go` is
only used for CREATE TABLE DDL). Add `maxSelectColumns = 64` as a new constant. Do not reuse
`maxColumns`, which is scoped to DDL. Reject with an error if exceeded.

### SELECT clause build logic (both builders)

```
if len(Columns) == 0 && len(Aggregates) == 0 → SELECT *
if len(Columns) > 0 && len(Aggregates) == 0  → SELECT col1, col2 (current behavior)
if len(Columns) == 0 && len(Aggregates) > 0  → SELECT COUNT(*), SUM(amount)
if len(Columns) > 0 && len(Aggregates) > 0   → SELECT col1, col2, COUNT(*), SUM(amount)
```

### Aggregate validation (new function)

```go
func ValidateAggregate(a AggregateColumn) error
```

- `Func` must be in allowlist `{COUNT, SUM, AVG, MIN, MAX}`
- `Arg` must be `"*"` (only valid with COUNT) or pass `ValidColumnName()`
- `Alias` must be empty or pass `ValidColumnName()`
- `Arg == "*"` with `Func != "COUNT"` → error

### Aggregate SQL rendering (new function)

```go
func buildAggregateExpr(d Dialect, a AggregateColumn) (string, error)
```

Returns e.g. `COUNT(*)`, `SUM("amount")`, `AVG("price") AS "avg_price"`.
No parameterized values — aggregates are pure structural SQL. Unlike Condition `Build()` methods,
this function does not take `argOffset`, does not return args or nextOffset. It produces a SQL
fragment string only. Use `quoteIdent(d, a.Arg)` for column name args. Emit `*` as a literal
(no quoting). `AggregateColumn.Arg` contains only simple column names after `ValidColumnName`
validation — qualified args (with dots) are rejected by `ValidColumnName`, so
`quoteQualifiedIdent` is never needed here.

### HAVING in FilteredSelectParams

Uses the existing `Condition` interface (not `map[string]any`). This is cleaner than the legacy
Having and supports the full condition tree. However, HAVING often references aggregate results.

Two approaches for HAVING on aggregates:

1. **By alias:** `HAVING "total" > 5` — requires the aggregate to have an alias. The `Compare`
   condition with `Column: "total"` would reference the alias. This works in SQLite and MySQL but
   not always in PostgreSQL (which requires repeating the aggregate expression).

2. **AggregateCondition type:** A new `Condition` implementation that embeds an `AggregateColumn`
   and an operator/value. Builds to `HAVING COUNT(*) > ?`.

**Decision:** Implement `AggregateCondition` as a new `Condition` type. This is portable across
all three backends and doesn't require aliases.

```go
type AggregateCondition struct {
    Agg   AggregateColumn
    Op    CompareOp // reuses existing CompareOp type (OpEq, OpNeq/`<>`, OpGt, OpGte, OpLt, OpLte)
    Value any       // must be non-nil (reject with error, same as Compare)
}
```

`AggregateCondition.Op` reuses the existing `CompareOp` type and its `validCompareOps` allowlist.
This avoids a parallel validation path. Note: `OpLike` is technically in `CompareOp` but is
nonsensical for aggregates — `AggregateCondition.Build()` rejects `OpLike` when used with an
aggregate. There is no `OpNotLike` in `CompareOp` (NOT LIKE only exists as a `ColumnOp` in the
legacy builder). The system uses `<>` for not-equal (`OpNeq`), not `!=`.

Builds to: `COUNT(*) > ?` or `SUM("amount") >= ?`

### Lua API

**Aggregate constructors** (registered on the `db` module table):

`db.count` is already registered as the `luaCount` function (`SELECT COUNT(*) FROM table`).
To avoid overwriting it, aggregate constructors use an `agg_` prefix:

```lua
db.agg_count("*")                    -- {__agg="COUNT", __arg="*", __alias=""}
db.agg_count("*", "total")           -- {__agg="COUNT", __arg="*", __alias="total"}
db.agg_sum("amount")                 -- {__agg="SUM", __arg="amount", __alias=""}
db.agg_sum("amount", "total_amount") -- {__agg="SUM", __arg="amount", __alias="total_amount"}
db.agg_avg("price", "avg_price")     -- {__agg="AVG", ...}
db.agg_min("score")                  -- {__agg="MIN", ...}
db.agg_max("score")                  -- {__agg="MAX", ...}
```

Signature: `db.agg_<func>(arg [, alias])` — arg is required, alias is optional.

**Usage in columns:**

```lua
db.query("tasks", {
    columns = {"status", db.agg_count("*", "total")},
    group_by = {"status"},
    having = {db.agg_count("*"), ">", 5}
})
```

`parseColumnsFromLua` is updated to detect aggregate sentinel tables (check for `__agg` key)
alongside plain strings. Plain strings → `opts.columns`, sentinel tables → `opts.aggregates`.

**HAVING in Lua (new condition-style):**

```lua
having = {db.agg_count("*"), ">", 5}
-- or
having = {"AND", {
    {db.agg_count("*"), ">", 5},
    {db.agg_sum("amount"), "<", 1000}
}}
```

This reuses the condition parsing infrastructure with aggregate sentinels in the column position.

**Concrete parse dispatch for aggregate sentinels in `parseConditionFromLua`:**

The existing parser checks element 1 of a Lua sequence to determine the parse path:
- Element 1 is a string → parse as keyword (`"AND"`, `"OR"`, `"NOT"`) or column name for Compare
- Element 1 is a table (Lua sequence) → parse as implicit AND over sub-conditions

Aggregate sentinels are Lua tables with an `__agg` key. The parser adds a check **before** the
implicit-AND fallthrough: if element 1 is a `*lua.LTable`, call
`first.(*lua.LTable).RawGetString("__agg")`. If the result is not `lua.LNil`, parse as
`AggregateCondition` (extract `__agg`, `__arg`, `__alias` via `RawGetString`, then read element 2
as operator string and element 3 as value). If `__agg` is `lua.LNil`, fall through to implicit-AND
as before.

This is unambiguous because implicit-AND sub-conditions are Lua sequences (integer-keyed tables)
and will never have a string key `__agg`.

**Note:** This modification to `parseConditionFromLua` means `AggregateCondition` is technically
expressible in WHERE clauses too (`WHERE COUNT(*) > 5`). The builder does not reject this — the
database will return a SQL error for aggregates in WHERE without subqueries. This is consistent
with the plan's approach of "the builder does not reject semantically nonsensical combinations;
it is the caller's responsibility."

### Nil Filter relaxation in FilteredSelectParams

Currently `buildFilteredSelectQuery` rejects `Filter == nil` unconditionally (line 293-294 of
`query_filtered.go`). GROUP BY without WHERE is standard SQL and is this plan's target query.

**Rule:** `Filter` is required unless `len(Aggregates) > 0 || len(GroupBy) > 0`. When Filter is
nil and aggregates/GroupBy are present, the builder omits the WHERE clause entirely. When Filter
is nil and no aggregates or GroupBy are present, the builder returns an error (preserving current
behavior for non-aggregate queries).

### DISTINCT + aggregates behavior

`SELECT DISTINCT` with `GROUP BY` is valid SQL but redundant — GROUP BY already deduplicates the
grouped columns. `SELECT DISTINCT` with aggregates but no GROUP BY is semantically nonsensical in
most cases. The builder does not reject this combination; it is the caller's responsibility.
`DISTINCT` applies to the entire row, not individual aggregate expressions. `COUNT(DISTINCT col)`
is a separate feature listed under "Not In Scope."

### GroupBy in FilteredSelectParams

Add `GroupBy []string` to `FilteredSelectParams`. Validated same as columns (`ValidColumnName`
each). The `buildFilteredSelectQuery` appends `GROUP BY col1, col2` after WHERE.

The Lua filtered path (`opts.filter != nil`) currently silently drops `opts.groupBy`. After this
change, it flows through.

### PostgreSQL placeholder offset tracking

The current `buildFilteredSelectQuery` discards the `nextOffset` return from `Filter.Build()`:
```go
whereSQL, args, _, err := p.Filter.Build(bctx, d, 1)
```

HAVING needs the correct next offset for PostgreSQL `$N` placeholders. Change to:
```go
whereSQL, args, nextOffset, err := p.Filter.Build(bctx, d, 1)
```

When Filter is nil (aggregate-only query with no WHERE), `nextOffset` starts at 1. Pass
`nextOffset` into `Having.Build()` to continue placeholder numbering. The args from HAVING are
appended to the WHERE args slice. Final arg order: WHERE args, then HAVING args.

### db.count() and db.exists() do not support GROUP BY

The existing `QCountFiltered` returns a single `int64` (total row count). The existing
`QExistsFiltered` returns a single `bool`. These functions are not compatible with GROUP BY
semantics (which produce per-group results). `db.count()` and `db.exists()` in the Lua API
continue to use their existing implementations unchanged. Plugin authors who need per-group
counts use `db.query()` with `db.agg_count("*")` in the columns list and `group_by`.

## Implementation Steps

### Step 1: Core types and validation (`internal/db/query_builder.go`)

- [ ] Add `AggregateColumn` struct
- [ ] Add `ValidateAggregate(AggregateColumn) error` function
- [ ] Add `buildAggregateExpr(Dialect, AggregateColumn) (string, error)` function
- [ ] Add `validAggregateFuncs` allowlist var
- [ ] Add `Aggregates []AggregateColumn` field to `SelectParams`

### Step 2: Legacy builder support (`internal/db/query_builder.go`)

- [ ] Update `buildSelectQuery` column section to handle `Aggregates`
- [ ] Enforce `len(Columns) + len(Aggregates) <= 64` — return error if exceeded
- [ ] When `len(Columns) == 0 && len(Aggregates) > 0`, don't default to `*`
- [ ] Append validated aggregate expressions after plain columns

### Step 3: Filtered builder support (`internal/db/query_filtered.go`)

- [ ] Add `Aggregates []AggregateColumn` field to `FilteredSelectParams`
- [ ] Add `GroupBy []string` field to `FilteredSelectParams`
- [ ] Add `Having Condition` field to `FilteredSelectParams`
- [ ] Relax nil-Filter check: allow `Filter == nil` when `len(Aggregates) > 0 || len(GroupBy) > 0`.
      When Filter is nil, omit the WHERE clause and set `nextOffset = 1`
- [ ] Enforce `len(Columns) + len(Aggregates) <= 64` — return error if exceeded
- [ ] Update `buildFilteredSelectQuery` to handle aggregates in SELECT
- [ ] Update `buildFilteredSelectQuery` to append GROUP BY clause (after WHERE, before HAVING)
- [ ] Reject `Having != nil` when `len(GroupBy) == 0` — HAVING without GROUP BY is rejected
      (same constraint as the legacy builder at line 861 of `query_builder.go`)
- [ ] Capture `nextOffset` from `Filter.Build()` (change `_` to `nextOffset`). Pass `nextOffset`
      into `Having.Build()` for correct PostgreSQL `$N` placeholder numbering. Append HAVING args
      after WHERE args

### Step 4: AggregateCondition (`internal/db/query_conditions.go`)

- [ ] Add `AggregateCondition` struct implementing `Condition` (uses `CompareOp` for Op field)
- [ ] In `AggregateCondition.Build()`: call `ctx.incrementNode()` as the first operation (same
      pattern as all other Condition implementations). Then validate the aggregate via
      `ValidateAggregate`, validate Op against `validCompareOps` (reject `OpLike`), reject nil
      Value. Render `FUNC(arg) op ?` (or `FUNC(arg) op $N` for PostgreSQL using `argOffset`),
      return the value as a single-element args slice. `ValidateCondition` requires no changes —
      it validates by calling `Build()`, which naturally includes `AggregateCondition` in node
      counting
- [ ] Update `HasValueBinding` type switch to return `true` for `AggregateCondition` (it always
      binds a parameterized value)

### Step 5: Tests for Go query builder (`internal/db/`)

- [ ] `TestBuildSelectQuery_Aggregates` — COUNT(*) only, with alias, mixed columns+aggregates
- [ ] `TestBuildSelectQuery_AggregateValidation` — bad func name, bad arg, star with SUM
- [ ] `TestBuildSelectQuery_AggregatesWithGroupByAndHaving` — verify aggregates work alongside
      existing GroupBy/Having `map[string]any` in the legacy builder
- [ ] `TestBuildFilteredSelectQuery_Aggregates` — same cases for filtered path
- [ ] `TestBuildFilteredSelectQuery_GroupBy` — GROUP BY validation and SQL output
- [ ] `TestBuildFilteredSelectQuery_Having` — HAVING with AggregateCondition
- [ ] `TestBuildFilteredSelectQuery_NilFilter_WithAggregates` — verify nil Filter is accepted when
      Aggregates or GroupBy are present, and rejected when neither is present
- [ ] `TestBuildFilteredSelectQuery_ColumnCap` — verify `len(Columns) + len(Aggregates) > 64` is
      rejected
- [ ] `TestBuildFilteredSelectQuery_Having_PostgresPlaceholders` — verify HAVING args get correct
      `$N` numbering after WHERE args (use `DialectPostgres`)
- [ ] `TestAggregateCondition_Build` — SQL output, arg placement, validation errors, `OpLike`
      rejection (only `OpLike` — there is no `OpNotLike` in `CompareOp`)
- [ ] `TestAggregateCondition_HasValueBinding` — verify `HasValueBinding` returns true
- [ ] Integration: `QSelectFiltered` with GROUP BY + aggregate on real SQLite

### Step 6: Lua aggregate constructors (`internal/plugin/db_api.go`)

- [ ] Add `registerAggregateConstructors(L, mod)` — registers `db.agg_count`, `db.agg_sum`,
      `db.agg_avg`, `db.agg_min`, `db.agg_max` as Lua functions returning sentinel tables
      `{__agg, __arg, __alias}`. Uses `agg_` prefix to avoid overwriting the existing
      `db.count` (`luaCount`) function
- [ ] Call from `RegisterDBAPI`

### Step 7: Lua column parsing (`internal/plugin/db_api.go`)

- [ ] Update `parsedSelectOpts` to add `aggregates []db.AggregateColumn` and
      `havingCondition db.Condition` (the existing `having map[string]any` field stays for legacy
      path backward compat; the new field carries the parsed `Condition` for the filtered path)
- [ ] Update `parseColumnsFromLua` to detect aggregate sentinels (`__agg` key) alongside strings
- [ ] Plain strings → `opts.columns`, sentinel tables → `opts.aggregates`
- [ ] When `parseColumnsFromLua` encounters a sentinel table, extract `__arg`. If `__arg`
      contains a dot and is not `*`, run it through `prefixQualifiedColumn`. Construct the
      `db.AggregateColumn` with the prefixed arg. Unqualified column args pass through unchanged.
      `*` is never prefixed.

### Step 8: Lua HAVING parsing (`internal/plugin/db_conditions.go`)

- [ ] Add `parseHavingCondition(L, optsTbl)` that parses HAVING using condition syntax, returning
      `db.Condition` (not `map[string]any`)
- [ ] Extend `parseConditionFromLua`: when element 1 is a Lua table, check for `__agg` string key
      before falling through to implicit-AND. If `__agg` is present, parse as `AggregateCondition`
      (read `__agg`, `__arg`, `__alias` from the table, element 2 as `CompareOp`, element 3 as
      value). If `__agg` is absent, fall through to implicit-AND as before
- [ ] Update `parseSelectOpts` HAVING dispatch: after `parseWhereExtended` returns, check
      `opts.filter != nil`. If true (new condition path), call `parseHavingCondition` to populate
      `opts.havingCondition` and skip `parseHavingFromLua`. If false (legacy path), call
      `parseHavingFromLua` as before to populate `opts.having`. Never populate both

### Step 9: Lua query execution wiring (`internal/plugin/db_api.go`)

- [ ] Change `luaQuery` dispatch condition from `opts.filter != nil` to
      `opts.filter != nil || opts.havingCondition != nil || len(opts.aggregates) > 0`. This ensures
      aggregate queries without WHERE clauses route through `FilteredSelectParams` (which supports
      nil-Filter relaxation). When dispatching to the filtered path with `opts.filter == nil`, pass
      `nil` as the `Filter` field
- [ ] Update `luaQuery` filtered path to pass `opts.aggregates`, `opts.groupBy`, and
      `opts.havingCondition` into `FilteredSelectParams`
- [ ] Update `luaQuery` legacy path to pass `opts.aggregates` into `SelectParams` (this path is
      only reached when no aggregates or condition-style HAVING are present)
- [ ] Update `luaQueryOne` dispatch condition and field wiring the same way

### Step 10: Lua API tests (`internal/plugin/`)

- [ ] Test aggregate constructors return correct sentinel tables
- [ ] Test `parseColumnsFromLua` with mixed strings and aggregate sentinels
- [ ] Test full round-trip: `db.query("tasks", {columns = {"status", db.agg_count("*", "total")}, group_by = {"status"}})` returns grouped rows
- [ ] Test GROUP BY without WHERE (no `where` key in opts) — verifies nil-Filter relaxation
      works through the Lua path
- [ ] Test HAVING with aggregate condition filters results
- [ ] Test HAVING parse dispatch: `{db.agg_count("*"), ">", 5}` produces `AggregateCondition`,
      not implicit-AND
- [ ] Test validation: bad function name, star with non-COUNT, missing arg, OpLike in HAVING
- [ ] Test aggregate arg auto-prefixing (qualified with dot → prefixed, unqualified → unchanged,
      `*` → unchanged)
- [ ] Test that existing `db.count()` and `db.exists()` are not overwritten by aggregate
      constructors and continue to work unchanged

## Files Modified

| File | Changes |
|------|---------|
| `internal/db/query_builder.go` | `AggregateColumn`, `ValidateAggregate`, `buildAggregateExpr`, update `SelectParams`, update `buildSelectQuery` |
| `internal/db/query_filtered.go` | Update `FilteredSelectParams` (Aggregates, GroupBy, Having), update `buildFilteredSelectQuery` |
| `internal/db/query_conditions.go` | Add `AggregateCondition` implementing `Condition` |
| `internal/db/query_builder_test.go` | Aggregate tests for legacy builder |
| `internal/db/query_filtered_test.go` | Aggregate + GroupBy + Having tests for filtered builder |
| `internal/db/query_conditions_test.go` | `AggregateCondition` tests |
| `internal/plugin/db_api.go` | Aggregate constructors, update `parsedSelectOpts`, update `parseColumnsFromLua`, update `luaQuery`/`luaQueryOne` |
| `internal/plugin/db_conditions.go` | Extend condition parsing for aggregate sentinels, HAVING condition parsing |
| `internal/plugin/db_api_test.go` | Full round-trip Lua tests |

## Security Considerations

- **No raw SQL:** Aggregate function names are validated against a hardcoded allowlist. Arguments
  are validated as identifiers (or `*`). Aliases are validated as identifiers. No user-controlled
  string reaches the SQL query unvalidated.
- **Plugin table prefix:** Aggregate arguments that are column names don't need table prefixing
  because they refer to columns in the query's own table. If qualified (`table.col`), the table
  part is prefixed like existing columns.
- **Operation budget:** Aggregate queries count as one operation against the per-checkout limit
  (same as any other query). No special cost accounting needed — the database handles GROUP BY
  efficiently.

## Not In Scope

- **Window functions** (`ROW_NUMBER() OVER (...)`) — significantly more complex, different syntax
- **Subqueries** — would require recursive query building
- **DISTINCT inside aggregates** (`COUNT(DISTINCT col)`) — could be added later as a bool field
  on `AggregateColumn`
- **Multiple HAVING conditions with mixed aggregate/column refs** — the `Condition` tree handles
  this naturally via `And{AggregateCondition, Compare}`, but Lua syntax ergonomics for complex
  HAVING are deferred
