 Plan: INSERT ... ON CONFLICT DO UPDATE (Upsert) for Query Builder + Plugin Lua API

 Context

 Plugin authors currently work around missing upsert with select-then-insert/update, which is not atomic. The codebase already has 3 hand-rolled per-dialect upserts
 (request_engine.go:507-543, hook_engine.go:705-738, http_bridge.go:436-514), confirming the pattern is needed. This adds a single generic QUpsert to the query builder so plugins (and future
  internal code) can do atomic upserts without dialect switch blocks. Note: http_bridge.go uses CASE expressions in its SET clauses (conditional logic based on comparing old vs excluded values) — QUpsert cannot express this and that upsert remains hand-rolled. The request_engine.go and hook_engine.go upserts are simple enough to be replaced by QUpsert in a follow-up.

 Design Decisions

 - Update columns: nil Update = auto-derive from Values minus ConflictColumns using excluded/VALUES() references (zero extra params). Explicit Update = parameterized SET with caller-provided
  values.
 - DO NOTHING: Supported via DoNothing bool. SQLite/Postgres: ON CONFLICT ... DO NOTHING. MySQL: ON DUPLICATE KEY UPDATE with no-op assignment (first ConflictColumn = first ConflictColumn). INSERT IGNORE is not used because it swallows all errors (data truncation, FK violations, NOT NULL), not just duplicate key conflicts.
 - ConflictColumns always required. MySQL ignores them in SQL but validates them — forces callers to think about which conflict they're resolving.
 - Single-row only. No bulk upsert needed yet.
 - File placement: QUpsert in query_builder.go alongside QInsert (same value-map pattern).

 Files to Modify

 ┌───────────────────────────────────┬────────────────────────────────────────────────────────────────────┐
 │               File                │                               Change                               │
 ├───────────────────────────────────┼────────────────────────────────────────────────────────────────────┤
 │ internal/db/query_builder.go      │ UpsertParams struct, buildUpsertQuery(), QUpsert()                 │
 ├───────────────────────────────────┼────────────────────────────────────────────────────────────────────┤
 │ internal/db/query_builder_test.go │ TestQUpsert — functional + validation + SQL generation per dialect │
 ├───────────────────────────────────┼────────────────────────────────────────────────────────────────────┤
 │ internal/plugin/db_api.go         │ luaUpsert method, register db.upsert in RegisterDBAPI              │
 ├───────────────────────────────────┼────────────────────────────────────────────────────────────────────┤
 │ internal/plugin/db_api_test.go    │ Lua-level upsert tests                                             │
 └───────────────────────────────────┴────────────────────────────────────────────────────────────────────┘

 Step 1: UpsertParams + QUpsert in query_builder.go

 Struct

 // UpsertParams configures an INSERT ... ON CONFLICT DO UPDATE query.
 type UpsertParams struct {
     Table           string
     Values          map[string]any // columns and values to insert (required, non-empty)
     ConflictColumns []string       // conflict target columns (required, non-empty, must all exist in Values)
     Update          map[string]any // SET on conflict; nil = auto-derive from Values minus ConflictColumns
     DoNothing       bool           // ON CONFLICT DO NOTHING; mutually exclusive with Update
 }

 buildUpsertQuery(d Dialect, p UpsertParams) (string, []any, error)  // unexported, matching buildSelectQuery. Only QUpsert is exported.

 Validation:
 1. ValidTableName(p.Table)
 2. p.Values non-empty, all keys ValidColumnName
 3. p.ConflictColumns non-empty, all ValidColumnName, all must be keys in Values
 4. p.DoNothing && p.Update != nil = error (mutually exclusive)
 5. p.Update non-nil: must be non-empty, all keys ValidColumnName, no key may appear in ConflictColumns (updating a conflict target column produces surprising results)

 SQL generation:
 1. Sort Values keys deterministically via sortedKeys()
 2. Build INSERT portion: INSERT INTO "table" ("c1", "c2") VALUES (?, ?)
 3. Append conflict/update clause per dialect:

 DoNothing:
 - SQLite/Postgres: ON CONFLICT ("c1") DO NOTHING
 - MySQL: ON DUPLICATE KEY UPDATE "c1" = "c1" (no-op assignment using first ConflictColumn; preserves error semantics unlike INSERT IGNORE which swallows all errors)

 Update nil (auto-derive):
 - Compute update columns: sorted Values keys minus ConflictColumns
 - If no update columns remain, error (all insert cols are conflict cols — use DoNothing)
 - SQLite: ON CONFLICT ("c1") DO UPDATE SET "c2" = excluded."c2", "c3" = excluded."c3"
 - Postgres: ON CONFLICT ("c1") DO UPDATE SET "c2" = EXCLUDED."c2", "c3" = EXCLUDED."c3" (uppercase EXCLUDED — matches codebase convention in request_engine.go, hook_engine.go, http_bridge.go)
 - MySQL: ON DUPLICATE KEY UPDATE "c2" = VALUES("c2"), "c3" = VALUES("c3")
 - No additional args needed

 Update explicit:
 - Sort Update keys deterministically
 - SQLite: ON CONFLICT ("c1") DO UPDATE SET "c2" = ?, "c3" = ? — new args appended after INSERT args
 - Postgres: ON CONFLICT ("c1") DO UPDATE SET "c2" = ?, "c3" = ? — new args appended, placeholder index continues ($N+1, $N+2, ...)
 - MySQL: ON DUPLICATE KEY UPDATE "c2" = ?, "c3" = ? — new args appended

 NULL values: In the INSERT portion, nil values in Values are passed as parameterized args (same as QInsert — the driver handles NULL binding). In the explicit Update map, nil values generate SET "col" = NULL without a placeholder (same as QUpdate). Auto-derived update columns always use excluded/VALUES() references and do not produce placeholders, so nil handling is not applicable there.

 QUpsert(ctx, exec, d, p) (sql.Result, error)

 Calls buildUpsertQuery, then exec.ExecContext.

 Step 2: Tests in query_builder_test.go

 Create a setupUpsertTestDB helper that creates a table with both a PRIMARY KEY and a separate composite UNIQUE constraint (e.g., UNIQUE(name, status)) for the multi_column_conflict test case. Single-column conflict tests use the PRIMARY KEY.

 Functional (SQLite in-memory via setupUpsertTestDB):
 - insert_new_row — no conflict, row inserted
 - update_on_conflict — conflict, row updated via auto-derive
 - explicit_update — only specified columns change
 - do_nothing_on_conflict — existing row unchanged
 - do_nothing_new_row — no conflict, row inserted
 - nil_value_in_update — update column to NULL
 - multi_column_conflict — composite unique key

 Validation errors:
 - empty_values, empty_conflict_columns, conflict_not_in_values, invalid_table, invalid_column, do_nothing_with_update, empty_explicit_update, update_key_in_conflict_columns

 SQL generation (string assertion, no DB — use dialect-appropriate quoting: double quotes for SQLite/Postgres, backticks for MySQL, matching quoteIdent output):
 - sqlite_auto_derive, postgres_auto_derive (verify $N placeholders + uppercase EXCLUDED), postgres_explicit (verify $N continues), mysql_auto_derive (VALUES()), mysql_do_nothing (verify ON DUPLICATE KEY UPDATE with no-op assignment, not INSERT IGNORE)

 Step 3: luaUpsert in db_api.go

 Lua API

 db.upsert("tasks", {
     values = {id = "abc", title = "Hello", status = "active"},
     conflict_columns = {"id"},
     update = {title = "Hello"},     -- optional
     do_nothing = false,              -- optional
 })

 Implementation (follows luaUpdate pattern for argument parsing: CheckString(1) + CheckTable(2) with named fields; follows luaInsert pattern for timestamp auto-setting)

 1. checkOpLimit()
 2. Parse table string, prefixTable(api.pluginName, tableName)
 3. Parse opts table:
   - values (required) via LuaTableToMap()
   - conflict_columns (required): L.GetField(optsTbl, "conflict_columns"). If nil or not a table, L.ArgError(2, "upsert requires a 'conflict_columns' table"). Iterate as Lua sequence (integer keys 1..N). Each value must be lua.LString; if not, L.ArgError(2, "conflict_columns values must be strings"). If resulting slice is empty, L.ArgError(2, "conflict_columns cannot be empty").
   - update (optional) via LuaTableToMap() if present
   - do_nothing (optional): L.GetField(optsTbl, "do_nothing"); if not lua.LNil, use lua.LVAsBool to coerce to Go bool; default false
 4. Auto-set in values: updated_at if missing. Do NOT auto-set id (caller must provide it — auto-generating a ULID would never conflict, making the update path unreachable). Do NOT auto-set created_at (auto-derive would include it in the SET clause, overwriting the original creation timestamp on conflict).
 5. Auto-set in update (when non-nil, not DoNothing): updated_at if missing
 6. Call db.QUpsert(ctx, api.currentExec, api.dialect, db.UpsertParams{...})
 7. Error handling: follow luaInsert pattern exactly. checkOpLimit() failure and nil L.Context() use L.RaiseError (hard abort). All other errors (prefixTable, LuaTableToMap validation, QUpsert) return nil, errmsg (two values, recoverable).

 Registration

 In RegisterDBAPI, add after insert_many:
 dbTable.RawSetString("upsert", L.NewFunction(api.luaUpsert))

 Step 4: Tests in db_api_test.go

 - inserts_new_row, updates_on_conflict, auto_sets_timestamps, auto_sets_updated_at_in_update, do_nothing, requires_conflict_columns, requires_values, namespace_prefixed

 Step 5: Verify

 just check                          # compile
 just test ./internal/db/...         # query builder tests
 just test ./internal/plugin/...     # plugin API tests

 Notes

 - MySQL VALUES(col) is deprecated in 8.0.20+ but still works in all versions and MariaDB. Comment this in code for future reference.
 - Postgres uses uppercase EXCLUDED (matching existing codebase convention in request_engine.go, hook_engine.go, http_bridge.go). SQLite uses lowercase excluded.
 - Plugin tables need a UNIQUE constraint for upsert to work — this is a database-level concern, not something the query builder validates.
 - QUpsert returns sql.Result but luaUpsert discards it. RowsAffected semantics differ across dialects for DoNothing (0 on conflict skip, 1 on insert) and are not exposed to Lua.
