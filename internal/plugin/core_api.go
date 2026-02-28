package plugin

import (
	"fmt"

	db "github.com/hegner123/modulacms/internal/db"
	lua "github.com/yuin/gopher-lua"
)

// CoreTablePolicy defines what operations are allowed on a core CMS table.
type CoreTablePolicy struct {
	Read           bool
	Write          bool
	BlockedColumns []string // columns stripped from query results / rejected pre-query
}

// coreTablePolicy is the hardcoded whitelist of core CMS tables accessible to plugins.
// Tables not listed here are blocked entirely (e.g., tokens, sessions).
var coreTablePolicy = map[string]CoreTablePolicy{
	// Content tables (read + write)
	"content_data":   {Read: true, Write: true},
	"content_fields": {Read: true, Write: true},

	// Schema tables (read-only -- plugins should not alter CMS schema)
	"datatypes": {Read: true, Write: false},
	"fields":    {Read: true, Write: false},
	"tables":    {Read: true, Write: false},

	// Media (read + write)
	"media": {Read: true, Write: true},

	// Routes (read-only -- route changes need admin oversight)
	"routes": {Read: true, Write: false},

	// Users (read-only, blocked columns)
	"users": {Read: true, Write: false, BlockedColumns: []string{"hash"}},

	// Roles + Permissions (read-only)
	"roles":            {Read: true, Write: false},
	"permissions":      {Read: true, Write: false},
	"role_permissions": {Read: true, Write: false},

	// Change events (read-only -- audit trail)
	"change_events": {Read: true, Write: false},
}

// CoreTableAPI provides the sandboxed core.* Lua module for reading/writing
// core CMS tables with per-plugin permission checks.
//
// INVARIANT: each CoreTableAPI instance is bound to exactly one LState,
// just like DatabaseAPI. Never share across VMs.
type CoreTableAPI struct {
	dbAPI          *DatabaseAPI // shared op counter + executor + conn
	pluginName     string
	dialect        db.Dialect
	approvedAccess PluginCoreAccess // loaded from plugins.approved_access DB column
	maxRows        int              // default 100
	maxWriteRows   int              // default 100
}

// NewCoreTableAPI creates a new CoreTableAPI bound to the given DatabaseAPI.
// The DatabaseAPI provides the shared op counter, executor, and connection.
// approvedAccess may be nil (no core table access granted).
func NewCoreTableAPI(dbAPI *DatabaseAPI, pluginName string, dialect db.Dialect, approvedAccess PluginCoreAccess) *CoreTableAPI {
	return &CoreTableAPI{
		dbAPI:          dbAPI,
		pluginName:     pluginName,
		dialect:        dialect,
		approvedAccess: approvedAccess,
		maxRows:        100,
		maxWriteRows:   100,
	}
}

// checkAccess verifies the plugin has permission for the given table and operation.
// Returns nil if allowed, error if denied.
// opType is "read" or "write".
func (api *CoreTableAPI) checkAccess(table, opType string) error {
	// 1. Whitelist check
	policy, ok := coreTablePolicy[table]
	if !ok {
		return fmt.Errorf("core.%s: table %q is not accessible to plugins", opType, table)
	}

	// 2. Policy check
	if opType == "read" && !policy.Read {
		return fmt.Errorf("core.%s: table %q does not allow read access", opType, table)
	}
	if opType == "write" && !policy.Write {
		return fmt.Errorf("core.%s: table %q does not allow write access", opType, table)
	}

	// 3. Approved access check
	if api.approvedAccess == nil {
		return fmt.Errorf("core.%s: plugin %q has no approved access to table %q", opType, api.pluginName, table)
	}
	ops, hasTable := api.approvedAccess[table]
	if !hasTable {
		return fmt.Errorf("core.%s: plugin %q has no approved access to table %q", opType, api.pluginName, table)
	}
	for _, op := range ops {
		if op == opType {
			return nil
		}
	}
	return fmt.Errorf("core.%s: plugin %q is not approved for %q on table %q", opType, api.pluginName, opType, table)
}

// validateColumns checks that explicitly requested columns do not include blocked columns.
// Called before query execution to prevent blocked data from being read into Go memory.
func validateColumns(requested []string, blocked []string) error {
	if len(blocked) == 0 || len(requested) == 0 {
		return nil
	}
	blockedSet := make(map[string]bool, len(blocked))
	for _, col := range blocked {
		blockedSet[col] = true
	}
	for _, col := range requested {
		if blockedSet[col] {
			return fmt.Errorf("column %q is blocked and cannot be accessed by plugins", col)
		}
	}
	return nil
}

// stripBlockedColumns removes sensitive columns from SELECT * result rows.
// Only needed when no explicit columns are requested (wildcard select).
func stripBlockedColumns(rows []map[string]any, blocked []string) {
	if len(blocked) == 0 {
		return
	}
	blockedSet := make(map[string]bool, len(blocked))
	for _, col := range blocked {
		blockedSet[col] = true
	}
	for _, row := range rows {
		for col := range row {
			if blockedSet[col] {
				delete(row, col)
			}
		}
	}
}

// getBlockedColumns returns the blocked columns for a table, or nil if none.
func getBlockedColumns(table string) []string {
	policy, ok := coreTablePolicy[table]
	if !ok {
		return nil
	}
	return policy.BlockedColumns
}

// ===== OPTS PARSING =====

// parsedCoreOpts collects all parsed fields from a Lua opts table for core SELECT queries.
// Deliberately simpler than parsedSelectOpts: no joins, group_by, having, distinct.
type parsedCoreOpts struct {
	columns []string
	where   map[string]any
	whereOr []map[string]any
	orderBy string
	desc    bool
	orders  []db.OrderByClause
	limit   int64
	offset  int64
}

// parseCoreSelectOpts extracts SELECT-related fields from a Lua opts table.
// No prefix logic, no joins/group_by/having/distinct.
func parseCoreSelectOpts(L *lua.LState, optsTbl *lua.LTable) (parsedCoreOpts, error) {
	var opts parsedCoreOpts

	where, err := parseWhereWithConditions(L, optsTbl)
	if err != nil {
		return opts, err
	}
	opts.where = where

	whereOr, err := parseWhereOrFromLua(L, optsTbl)
	if err != nil {
		return opts, err
	}
	opts.whereOr = whereOr

	cols, err := parseCoreColumnsFromLua(L, optsTbl)
	if err != nil {
		return opts, err
	}
	opts.columns = cols

	orderBy, desc, orders, err := parseCoreOrderByFromLua(L, optsTbl)
	if err != nil {
		return opts, err
	}
	opts.orderBy = orderBy
	opts.desc = desc
	opts.orders = orders

	if limitVal := L.GetField(optsTbl, "limit"); limitVal != lua.LNil {
		if n, ok := limitVal.(lua.LNumber); ok {
			opts.limit = int64(n)
		}
	}
	if offsetVal := L.GetField(optsTbl, "offset"); offsetVal != lua.LNil {
		if n, ok := offsetVal.(lua.LNumber); ok {
			opts.offset = int64(n)
		}
	}

	return opts, nil
}

// parseCoreColumnsFromLua extracts the "columns" field without prefix logic.
func parseCoreColumnsFromLua(L *lua.LState, optsTbl *lua.LTable) ([]string, error) {
	colsVal := L.GetField(optsTbl, "columns")
	if colsVal == lua.LNil {
		return nil, nil
	}
	colsTbl, ok := colsVal.(*lua.LTable)
	if !ok {
		return nil, nil
	}

	var cols []string
	var parseErr error
	colsTbl.ForEach(func(key, value lua.LValue) {
		if parseErr != nil {
			return
		}
		if _, ok := key.(lua.LNumber); !ok {
			return
		}
		if s, ok := value.(lua.LString); ok {
			if err := db.ValidColumnName(string(s)); err != nil {
				parseErr = fmt.Errorf("columns: %w", err)
				return
			}
			cols = append(cols, string(s))
		}
	})
	if parseErr != nil {
		return nil, parseErr
	}
	return cols, nil
}

// parseCoreOrderByFromLua extracts the "order_by" field without prefix logic.
func parseCoreOrderByFromLua(L *lua.LState, optsTbl *lua.LTable) (string, bool, []db.OrderByClause, error) {
	obVal := L.GetField(optsTbl, "order_by")
	if obVal == lua.LNil {
		return "", false, nil, nil
	}

	// Legacy: plain string
	if s, ok := obVal.(lua.LString); ok {
		if err := db.ValidColumnName(string(s)); err != nil {
			return "", false, nil, fmt.Errorf("order_by: %w", err)
		}
		desc := parseBoolField(L, optsTbl, "desc")
		return string(s), desc, nil, nil
	}

	// New: table-of-tables [{column = "x", desc = true}, ...]
	obTbl, ok := obVal.(*lua.LTable)
	if !ok {
		return "", false, nil, nil
	}

	var orders []db.OrderByClause
	var parseErr error
	obTbl.ForEach(func(key, value lua.LValue) {
		if parseErr != nil {
			return
		}
		if _, ok := key.(lua.LNumber); !ok {
			return
		}
		entryTbl, ok := value.(*lua.LTable)
		if !ok {
			return
		}
		colVal := L.GetField(entryTbl, "column")
		colStr, ok := colVal.(lua.LString)
		if !ok || string(colStr) == "" {
			parseErr = fmt.Errorf("order_by entry missing 'column' string field")
			return
		}
		if err := db.ValidColumnName(string(colStr)); err != nil {
			parseErr = fmt.Errorf("order_by: %w", err)
			return
		}
		descVal := L.GetField(entryTbl, "desc")
		isDesc := false
		if b, ok := descVal.(lua.LBool); ok {
			isDesc = bool(b)
		}
		orders = append(orders, db.OrderByClause{
			Column: string(colStr),
			Desc:   isDesc,
		})
	})
	if parseErr != nil {
		return "", false, nil, parseErr
	}
	return "", false, orders, nil
}

// ===== LUA HANDLERS =====

// luaCoreQuery implements core.query(table, opts) -> Lua sequence table of row tables.
// Returns empty table on no matches, nil+errmsg on error.
func (api *CoreTableAPI) luaCoreQuery(L *lua.LState) int {
	if err := api.dbAPI.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)

	if err := api.checkAccess(tableName, "read"); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	if err := db.ValidTableName(tableName); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var opts parsedCoreOpts
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			parsed, err := parseCoreSelectOpts(L, optsTbl)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			opts = parsed
		}
	}

	// Pre-query column validation: reject blocked columns in explicit request.
	blocked := getBlockedColumns(tableName)
	if len(opts.columns) > 0 {
		if err := validateColumns(opts.columns, blocked); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
	}

	// Apply default limit if none specified.
	if opts.limit <= 0 {
		opts.limit = int64(api.maxRows)
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("core.query: no context set on Lua state (VM setup error)")
		return 0
	}

	rows, err := db.QSelect(ctx, api.dbAPI.currentExec, api.dialect, db.SelectParams{
		Table:   tableName,
		Columns: opts.columns,
		Where:   opts.where,
		WhereOr: opts.whereOr,
		OrderBy: opts.orderBy,
		Desc:    opts.desc,
		Orders:  opts.orders,
		Limit:   opts.limit,
		Offset:  opts.offset,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Post-query strip blocked columns for SELECT * (no explicit columns).
	result := make([]map[string]any, len(rows))
	for i, r := range rows {
		result[i] = map[string]any(r)
	}
	if len(opts.columns) == 0 {
		stripBlockedColumns(result, blocked)
	}

	L.Push(RowsToLuaTable(L, result))
	return 1
}

// luaCoreQueryOne implements core.query_one(table, opts) -> single row table or nil.
func (api *CoreTableAPI) luaCoreQueryOne(L *lua.LState) int {
	if err := api.dbAPI.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)

	if err := api.checkAccess(tableName, "read"); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	if err := db.ValidTableName(tableName); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var opts parsedCoreOpts
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			parsed, err := parseCoreSelectOpts(L, optsTbl)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			opts = parsed
		}
	}

	// Pre-query column validation.
	blocked := getBlockedColumns(tableName)
	if len(opts.columns) > 0 {
		if err := validateColumns(opts.columns, blocked); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("core.query_one: no context set on Lua state (VM setup error)")
		return 0
	}

	row, err := db.QSelectOne(ctx, api.dbAPI.currentExec, api.dialect, db.SelectParams{
		Table:   tableName,
		Columns: opts.columns,
		Where:   opts.where,
		WhereOr: opts.whereOr,
		OrderBy: opts.orderBy,
		Desc:    opts.desc,
		Orders:  opts.orders,
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

	rowMap := map[string]any(row)

	// Post-query strip blocked columns for SELECT *.
	if len(opts.columns) == 0 {
		stripBlockedColumns([]map[string]any{rowMap}, blocked)
	}

	L.Push(MapToLuaTable(L, rowMap))
	return 1
}

// luaCoreCount implements core.count(table, opts) -> integer.
func (api *CoreTableAPI) luaCoreCount(L *lua.LState) int {
	if err := api.dbAPI.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)

	if err := api.checkAccess(tableName, "read"); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	if err := db.ValidTableName(tableName); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var where map[string]any
	var err error
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			where, err = parseWhereWithConditions(L, optsTbl)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("core.count: no context set on Lua state (VM setup error)")
		return 0
	}

	count, err := db.QCount(ctx, api.dbAPI.currentExec, api.dialect, tableName, where)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LNumber(count))
	return 1
}

// luaCoreExists implements core.exists(table, opts) -> boolean.
func (api *CoreTableAPI) luaCoreExists(L *lua.LState) int {
	if err := api.dbAPI.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)

	if err := api.checkAccess(tableName, "read"); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	if err := db.ValidTableName(tableName); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var where map[string]any
	var err error
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			where, err = parseWhereWithConditions(L, optsTbl)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("core.exists: no context set on Lua state (VM setup error)")
		return 0
	}

	exists, err := db.QExists(ctx, api.dbAPI.currentExec, api.dialect, tableName, where)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LBool(exists))
	return 1
}

// luaCoreInsert implements core.insert(table, data).
// Does NOT auto-set id, created_at, or updated_at -- all values must be explicit.
// Returns nothing on success, nil+errmsg on error.
func (api *CoreTableAPI) luaCoreInsert(L *lua.LState) int {
	if err := api.dbAPI.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)
	valuesTbl := L.CheckTable(2)

	if err := api.checkAccess(tableName, "write"); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	if err := db.ValidTableName(tableName); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	values := LuaTableToMap(L, valuesTbl)

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("core.insert: no context set on Lua state (VM setup error)")
		return 0
	}

	_, err := db.QInsert(ctx, api.dbAPI.currentExec, api.dialect, db.InsertParams{
		Table:  tableName,
		Values: values,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// luaCoreUpdate implements core.update(table, opts).
// Requires non-empty 'set' and at least one of 'where' or 'where_or'.
// Does NOT auto-set updated_at.
// Enforces maxWriteRows safety limit via SELECT COUNT(*) pre-check.
// Returns nothing on success, nil+errmsg on error.
func (api *CoreTableAPI) luaCoreUpdate(L *lua.LState) int {
	if err := api.dbAPI.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)
	optsTbl := L.CheckTable(2)

	if err := api.checkAccess(tableName, "write"); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	if err := db.ValidTableName(tableName); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Parse set (required).
	setVal := L.GetField(optsTbl, "set")
	setTbl, ok := setVal.(*lua.LTable)
	if !ok || setVal == lua.LNil {
		L.ArgError(2, "core.update requires a 'set' table")
		return 0
	}
	setMap := LuaTableToMap(L, setTbl)
	if len(setMap) == 0 {
		L.ArgError(2, "core.update 'set' cannot be empty")
		return 0
	}

	// Parse where with conditions.
	where, err := parseWhereWithConditions(L, optsTbl)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Parse where_or.
	whereOr, err := parseWhereOrFromLua(L, optsTbl)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Safety: require at least one of where or where_or to be non-empty.
	if len(where) == 0 && len(whereOr) == 0 {
		L.RaiseError("core.update requires non-empty where or where_or (safety: prevents full-table update)")
		return 0
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("core.update: no context set on Lua state (VM setup error)")
		return 0
	}

	// Write safety limit: check row count before executing.
	count, err := db.QCount(ctx, api.dbAPI.currentExec, api.dialect, tableName, where)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("core.update: row count check failed: %s", err.Error())))
		return 2
	}
	if count > int64(api.maxWriteRows) {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("core.update: would affect %d rows, exceeding safety limit of %d", count, api.maxWriteRows)))
		return 2
	}

	_, err = db.QUpdate(ctx, api.dbAPI.currentExec, api.dialect, db.UpdateParams{
		Table:   tableName,
		Set:     setMap,
		Where:   where,
		WhereOr: whereOr,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// luaCoreDelete implements core.delete(table, opts).
// Requires at least one of 'where' or 'where_or' to be non-empty.
// Enforces maxWriteRows safety limit via SELECT COUNT(*) pre-check.
// Returns nothing on success, nil+errmsg on error.
func (api *CoreTableAPI) luaCoreDelete(L *lua.LState) int {
	if err := api.dbAPI.checkOpLimit(); err != nil {
		L.RaiseError("%s", err.Error())
		return 0
	}

	tableName := L.CheckString(1)
	optsTbl := L.CheckTable(2)

	if err := api.checkAccess(tableName, "write"); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	if err := db.ValidTableName(tableName); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Parse where with conditions.
	where, err := parseWhereWithConditions(L, optsTbl)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Parse where_or.
	whereOr, err := parseWhereOrFromLua(L, optsTbl)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Safety: require at least one of where or where_or to be non-empty.
	if len(where) == 0 && len(whereOr) == 0 {
		L.RaiseError("core.delete requires non-empty where or where_or (safety: prevents full-table delete)")
		return 0
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("core.delete: no context set on Lua state (VM setup error)")
		return 0
	}

	// Write safety limit: check row count before executing.
	count, err := db.QCount(ctx, api.dbAPI.currentExec, api.dialect, tableName, where)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("core.delete: row count check failed: %s", err.Error())))
		return 2
	}
	if count > int64(api.maxWriteRows) {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("core.delete: would affect %d rows, exceeding safety limit of %d", count, api.maxWriteRows)))
		return 2
	}

	_, err = db.QDelete(ctx, api.dbAPI.currentExec, api.dialect, db.DeleteParams{
		Table:   tableName,
		Where:   where,
		WhereOr: whereOr,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// ===== REGISTRATION =====

// RegisterCoreAPI creates a "core" Lua table with all core table access functions
// and sets it as a global. The provided CoreTableAPI instance must be bound to
// exactly one LState and must not be shared.
//
// After calling RegisterCoreAPI, the caller should call FreezeModule(L, "core")
// to make the module read-only.
func RegisterCoreAPI(L *lua.LState, api *CoreTableAPI) {
	coreTable := L.NewTable()

	coreTable.RawSetString("query", L.NewFunction(api.luaCoreQuery))
	coreTable.RawSetString("query_one", L.NewFunction(api.luaCoreQueryOne))
	coreTable.RawSetString("count", L.NewFunction(api.luaCoreCount))
	coreTable.RawSetString("exists", L.NewFunction(api.luaCoreExists))
	coreTable.RawSetString("insert", L.NewFunction(api.luaCoreInsert))
	coreTable.RawSetString("update", L.NewFunction(api.luaCoreUpdate))
	coreTable.RawSetString("delete", L.NewFunction(api.luaCoreDelete))

	L.SetGlobal("core", coreTable)
}
