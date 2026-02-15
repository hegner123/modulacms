package plugin

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	lua "github.com/yuin/gopher-lua"
)

// ErrOpLimitExceeded is returned when a plugin's per-checkout operation budget
// is exhausted. Callers can check with errors.Is(). The budget resets on the
// next VMPool.Get() via DatabaseAPI.ResetOpCount().
var ErrOpLimitExceeded = errors.New("operation limit exceeded")

// DatabaseAPI provides the sandboxed db.* Lua module for a single plugin.
//
// INVARIANT: each DatabaseAPI instance is bound to exactly one LState.
// Never share across VMs. The 1:1 binding means no concurrent access to
// currentExec, inTx, opCount, or inBeforeHook. If Phase 4 hot reload or
// recovery ever reuses a DatabaseAPI across VMs, a data race will occur.
type DatabaseAPI struct {
	conn          *sql.DB
	currentExec   db.Executor // normally conn; swapped to *sql.Tx inside transaction callbacks
	pluginName    string
	dialect       db.Dialect
	maxRows       int  // default 100, hard cap 10000 (enforced by db.QSelect)
	maxTxOps      int  // default 10
	inTx          bool // prevents nested db.transaction() calls
	opCount       int  // incremented on every db.* call, reset on VMPool.Get()
	maxOpsPerExec int  // default 1000, configurable via ManagerConfig

	// inBeforeHook is set to true by HookEngine.executeBefore before invoking the
	// Lua handler, and cleared on defer. When true, all db.* calls are blocked
	// (M1: SQLite deadlock prevention -- before-hooks run inside a transaction on
	// the main pool, and plugin db.* uses a separate pool that would deadlock).
	inBeforeHook bool
}

// NewDatabaseAPI creates a new DatabaseAPI bound to the given connection and plugin.
//
// INVARIANT: each DatabaseAPI instance is bound to exactly one LState.
// Never share across VMs. See struct comment for details.
func NewDatabaseAPI(conn *sql.DB, pluginName string, dialect db.Dialect, maxOpsPerExec int) *DatabaseAPI {
	if maxOpsPerExec <= 0 {
		maxOpsPerExec = 1000
	}
	return &DatabaseAPI{
		conn:          conn,
		currentExec:   conn,
		pluginName:    pluginName,
		dialect:       dialect,
		maxRows:       100,
		maxTxOps:      10,
		inTx:          false,
		opCount:       0,
		maxOpsPerExec: maxOpsPerExec,
	}
}

// ResetOpCount resets the per-checkout operation counter to 0.
// Called by the Manager after VMPool.Get() returns, before plugin code executes.
//
// If limitOverride is provided (one int), the max ops per execution is set to
// that value for this checkout. This is used by after-hook execution (M10) to
// apply a reduced budget. Zero args uses the original maxOpsPerExec.
func (api *DatabaseAPI) ResetOpCount(limitOverride ...int) {
	api.opCount = 0
	if len(limitOverride) > 0 && limitOverride[0] > 0 {
		api.maxOpsPerExec = limitOverride[0]
	}
}

// checkOpLimit increments the operation counter and returns an error if the
// per-checkout budget is exceeded. Called at the start of every db.* operation.
//
// M1: When inBeforeHook is true, all db.* calls are blocked. Before-hooks run
// inside a transaction on the main pool; the plugin's separate DB pool would
// deadlock on SQLite's file-level lock.
func (api *DatabaseAPI) checkOpLimit() error {
	if api.inBeforeHook {
		return fmt.Errorf("plugin %q: db.* calls are not allowed inside before-hooks", api.pluginName)
	}
	api.opCount++
	if api.opCount > api.maxOpsPerExec {
		return fmt.Errorf("plugin %q exceeded maximum operations per execution (%d): %w",
			api.pluginName, api.maxOpsPerExec, ErrOpLimitExceeded)
	}
	return nil
}

// prefixTable prepends the plugin namespace prefix and validates the result.
// Lua passes relative names (e.g., "tasks"); Go enforces the prefix
// (e.g., "plugin_task_tracker_tasks") via db.ValidTableName().
func prefixTable(pluginName, tableName string) (string, error) {
	prefixed := tablePrefix(pluginName) + tableName
	if err := db.ValidTableName(prefixed); err != nil {
		return "", fmt.Errorf("invalid table name %q: %w", tableName, err)
	}
	return prefixed, nil
}

// RegisterDBAPI creates a "db" Lua table with all database operation functions
// and sets it as a global. The provided DatabaseAPI instance must be bound to
// exactly one LState and must not be shared.
//
// After calling RegisterDBAPI, the caller should call FreezeModule(L, "db")
// to make the module read-only.
func RegisterDBAPI(L *lua.LState, api *DatabaseAPI) {
	dbTable := L.NewTable()

	dbTable.RawSetString("query", L.NewFunction(api.luaQuery))
	dbTable.RawSetString("query_one", L.NewFunction(api.luaQueryOne))
	dbTable.RawSetString("count", L.NewFunction(api.luaCount))
	dbTable.RawSetString("exists", L.NewFunction(api.luaExists))
	dbTable.RawSetString("insert", L.NewFunction(api.luaInsert))
	dbTable.RawSetString("update", L.NewFunction(api.luaUpdate))
	dbTable.RawSetString("delete", L.NewFunction(api.luaDelete))
	dbTable.RawSetString("transaction", L.NewFunction(api.luaTransaction))
	dbTable.RawSetString("ulid", L.NewFunction(luaULID))
	dbTable.RawSetString("timestamp", L.NewFunction(luaTimestamp))
	dbTable.RawSetString("define_table", L.NewFunction(
		luaDefineTable(L, api.pluginName, api.conn, api.dialect),
	))

	L.SetGlobal("db", dbTable)
}

// luaQuery implements db.query(table, opts) -> Lua sequence table of row tables.
// Returns empty table on no matches, nil+errmsg on error.
// Empty/nil where is allowed (returns all rows up to limit).
func (api *DatabaseAPI) luaQuery(L *lua.LState) int {
	if err := api.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)

	prefixed, err := prefixTable(api.pluginName, tableName)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Parse opts (optional second argument).
	var where map[string]any
	var orderBy string
	var limit int64
	var offset int64

	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			where = parseWhereFromLua(L, optsTbl)
			orderBy = luaStringField(L, optsTbl, "order_by")
			if limitVal := L.GetField(optsTbl, "limit"); limitVal != lua.LNil {
				if n, ok := limitVal.(lua.LNumber); ok {
					limit = int64(n)
				}
			}
			if offsetVal := L.GetField(optsTbl, "offset"); offsetVal != lua.LNil {
				if n, ok := offsetVal.(lua.LNumber); ok {
					offset = int64(n)
				}
			}
		}
	}

	// Apply default limit if none specified.
	if limit <= 0 {
		limit = int64(api.maxRows)
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.query: no context set on Lua state (VM setup error)")
		return 0
	}

	rows, err := db.QSelect(ctx, api.currentExec, api.dialect, db.SelectParams{
		Table:   prefixed,
		Where:   where,
		OrderBy: orderBy,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Always return a table (never nil), even for empty results.
	// Convert []db.Row to []map[string]any for RowsToLuaTable.
	// db.Row is a named type for map[string]any, so explicit conversion is needed.
	result := make([]map[string]any, len(rows))
	for i, r := range rows {
		result[i] = map[string]any(r)
	}
	L.Push(RowsToLuaTable(L, result))
	return 1
}

// luaQueryOne implements db.query_one(table, opts) -> single row table or nil.
// Returns nil on no match, nil+errmsg on error.
// Empty/nil where is allowed (returns arbitrary row).
func (api *DatabaseAPI) luaQueryOne(L *lua.LState) int {
	if err := api.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)

	prefixed, err := prefixTable(api.pluginName, tableName)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var where map[string]any
	var orderBy string

	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			where = parseWhereFromLua(L, optsTbl)
			orderBy = luaStringField(L, optsTbl, "order_by")
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.query_one: no context set on Lua state (VM setup error)")
		return 0
	}

	row, err := db.QSelectOne(ctx, api.currentExec, api.dialect, db.SelectParams{
		Table:   prefixed,
		Where:   where,
		OrderBy: orderBy,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	if row == nil {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(MapToLuaTable(L, map[string]any(row)))
	return 1
}

// luaCount implements db.count(table, opts) -> integer.
// Empty/nil where is allowed (returns total row count).
func (api *DatabaseAPI) luaCount(L *lua.LState) int {
	if err := api.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)

	prefixed, err := prefixTable(api.pluginName, tableName)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var where map[string]any
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			where = parseWhereFromLua(L, optsTbl)
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.count: no context set on Lua state (VM setup error)")
		return 0
	}

	count, err := db.QCount(ctx, api.currentExec, api.dialect, prefixed, where)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LNumber(count))
	return 1
}

// luaExists implements db.exists(table, opts) -> boolean.
// Empty/nil where is allowed (returns true if table has any rows).
func (api *DatabaseAPI) luaExists(L *lua.LState) int {
	if err := api.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)

	prefixed, err := prefixTable(api.pluginName, tableName)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var where map[string]any
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			where = parseWhereFromLua(L, optsTbl)
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.exists: no context set on Lua state (VM setup error)")
		return 0
	}

	exists, err := db.QExists(ctx, api.currentExec, api.dialect, prefixed, where)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LBool(exists))
	return 1
}

// luaInsert implements db.insert(table, values).
// Auto-sets id (ULID), created_at, and updated_at if not provided.
// Explicit values are never overridden.
// Returns nothing on success, nil+errmsg on error.
func (api *DatabaseAPI) luaInsert(L *lua.LState) int {
	if err := api.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)
	valuesTbl := L.CheckTable(2)

	prefixed, err := prefixTable(api.pluginName, tableName)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	values := LuaTableToMap(L, valuesTbl)

	// Auto-set id if not provided.
	if _, exists := values["id"]; !exists {
		values["id"] = types.NewULID().String()
	}

	// Auto-set created_at and updated_at if not provided.
	now := time.Now().UTC().Format(time.RFC3339)
	if _, exists := values["created_at"]; !exists {
		values["created_at"] = now
	}
	if _, exists := values["updated_at"]; !exists {
		values["updated_at"] = now
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.insert: no context set on Lua state (VM setup error)")
		return 0
	}

	_, err = db.QInsert(ctx, api.currentExec, api.dialect, db.InsertParams{
		Table:  prefixed,
		Values: values,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// luaUpdate implements db.update(table, opts).
// Requires non-empty 'set' and non-empty 'where' (empty where raises error()).
// Auto-sets updated_at in the set map if not provided.
// Returns nothing on success, nil+errmsg on error.
func (api *DatabaseAPI) luaUpdate(L *lua.LState) int {
	if err := api.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)
	optsTbl := L.CheckTable(2)

	prefixed, err := prefixTable(api.pluginName, tableName)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Parse set (required).
	setVal := L.GetField(optsTbl, "set")
	setTbl, ok := setVal.(*lua.LTable)
	if !ok || setVal == lua.LNil {
		L.ArgError(2, "update requires a 'set' table")
		return 0
	}
	setMap := LuaTableToMap(L, setTbl)
	if len(setMap) == 0 {
		L.ArgError(2, "update 'set' cannot be empty")
		return 0
	}

	// Parse where (required, must be non-empty for safety).
	where := parseWhereFromLua(L, optsTbl)
	if len(where) == 0 {
		L.RaiseError("db.update requires non-empty where (safety: prevents full-table update)")
		return 0
	}

	// Auto-set updated_at if not explicitly provided.
	if _, exists := setMap["updated_at"]; !exists {
		setMap["updated_at"] = time.Now().UTC().Format(time.RFC3339)
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.update: no context set on Lua state (VM setup error)")
		return 0
	}

	_, err = db.QUpdate(ctx, api.currentExec, api.dialect, db.UpdateParams{
		Table: prefixed,
		Set:   setMap,
		Where: where,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// luaDelete implements db.delete(table, opts).
// Requires non-empty 'where' (empty where raises error()).
// Returns nothing on success, nil+errmsg on error.
func (api *DatabaseAPI) luaDelete(L *lua.LState) int {
	if err := api.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)
	optsTbl := L.CheckTable(2)

	prefixed, err := prefixTable(api.pluginName, tableName)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	where := parseWhereFromLua(L, optsTbl)
	if len(where) == 0 {
		L.RaiseError("db.delete requires non-empty where (safety: prevents full-table delete)")
		return 0
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.delete: no context set on Lua state (VM setup error)")
		return 0
	}

	_, err = db.QDelete(ctx, api.currentExec, api.dialect, db.DeleteParams{
		Table: prefixed,
		Where: where,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// luaTransaction implements db.transaction(fn) -> true,nil | false,errmsg.
// Nested transactions are rejected with error(). Inside the callback, all
// db.* calls automatically route through the *sql.Tx via executor swap.
func (api *DatabaseAPI) luaTransaction(L *lua.LState) int {
	if api.inTx {
		L.ArgError(1, "nested transactions are not allowed")
		return 0
	}

	fn := L.CheckFunction(1)

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.transaction: no context set on Lua state (VM setup error)")
		return 0
	}

	err := types.WithTransaction(ctx, api.conn, func(tx *sql.Tx) error {
		// Swap executor -- all db.* calls now go through the transaction.
		prevExec := api.currentExec
		api.currentExec = tx
		api.inTx = true
		defer func() {
			api.currentExec = prevExec
			api.inTx = false
		}()

		// Call Lua function; Protect: true means error() returns err instead of panic.
		if callErr := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}); callErr != nil {
			return callErr // triggers rollback via WithTransaction
		}
		return nil
	})

	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

// luaULID implements db.ulid() -> string.
// Generates a new ULID using the thread-safe types.NewULID() function.
func luaULID(L *lua.LState) int {
	L.Push(lua.LString(types.NewULID().String()))
	return 1
}

// luaTimestamp implements db.timestamp() -> string.
// Returns the current time as an RFC3339 UTC string.
// This replaces os.date() which is sandboxed out.
func luaTimestamp(L *lua.LState) int {
	L.Push(lua.LString(time.Now().UTC().Format(time.RFC3339)))
	return 1
}

// parseWhereFromLua extracts the "where" field from a Lua opts table and
// converts it to a Go map[string]any. Returns nil if the field is absent,
// nil, or not a table.
func parseWhereFromLua(L *lua.LState, optsTbl *lua.LTable) map[string]any {
	whereVal := L.GetField(optsTbl, "where")
	if whereVal == lua.LNil {
		return nil
	}
	whereTbl, ok := whereVal.(*lua.LTable)
	if !ok {
		return nil
	}
	return LuaTableToMap(L, whereTbl)
}

// luaStringField reads an optional string field from a Lua table.
// Returns empty string if the field is absent, nil, or not a string.
func luaStringField(L *lua.LState, tbl *lua.LTable, field string) string {
	val := L.GetField(tbl, field)
	if s, ok := val.(lua.LString); ok {
		return string(s)
	}
	return ""
}
