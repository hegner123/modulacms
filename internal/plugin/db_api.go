package plugin

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	lua "github.com/yuin/gopher-lua"
)

// ErrOpLimitExceeded is returned when a plugin's per-checkout operation budget
// is exhausted. Callers can check with errors.Is(). The budget resets on the
// next VMPool.Get() via DatabaseAPI.ResetOpCount().
var ErrOpLimitExceeded = errors.New("operation limit exceeded")

// Sentinel keys used in Lua condition tables returned by db.gt(), db.like(), etc.
const (
	conditionSentinelKey = "__op"
	conditionValueKey    = "__val"
)

// validConditionOps is the whitelist of allowed operator names in Lua condition sentinels.
var validConditionOps = map[string]bool{
	"eq": true, "neq": true, "gt": true, "gte": true,
	"lt": true, "lte": true, "like": true, "not_like": true,
	"in": true, "not_in": true, "between": true,
	"is_null": true, "is_not_null": true,
}

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

	// tableReg registers plugin-defined tables in the CMS tables registry.
	// May be nil; when nil, tables are not registered in the registry.
	tableReg TableRegistrar
}

// NewDatabaseAPI creates a new DatabaseAPI bound to the given connection and plugin.
// tableReg may be nil; when nil, plugin-defined tables are not registered in the CMS registry.
//
// INVARIANT: each DatabaseAPI instance is bound to exactly one LState.
// Never share across VMs. See struct comment for details.
func NewDatabaseAPI(conn *sql.DB, pluginName string, dialect db.Dialect, maxOpsPerExec int, tableReg TableRegistrar) *DatabaseAPI {
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
		tableReg:      tableReg,
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

// prefixQualifiedColumn prefixes the table part of a qualified column reference.
// Unqualified columns ("col") pass through unchanged.
// Qualified columns ("table.col") have the table part prefixed with the plugin namespace.
func prefixQualifiedColumn(pluginName, col string) (string, error) {
	parts := strings.SplitN(col, ".", 2)
	if len(parts) == 1 {
		return col, nil
	}
	prefixed, err := prefixTable(pluginName, parts[0])
	if err != nil {
		return "", err
	}
	return prefixed + "." + parts[1], nil
}

// validateJoinTable ensures a prefixed table name belongs to the same plugin.
// Cross-plugin and core table access is blocked.
func validateJoinTable(pluginName, prefixedTable string) error {
	expectedPrefix := tablePrefix(pluginName)
	if !strings.HasPrefix(prefixedTable, expectedPrefix) {
		return fmt.Errorf("join table %q does not belong to plugin %q (cross-plugin access denied)", prefixedTable, pluginName)
	}
	return nil
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
	dbTable.RawSetString("insert_many", L.NewFunction(api.luaInsertMany))
	dbTable.RawSetString("upsert", L.NewFunction(api.luaUpsert))
	dbTable.RawSetString("update", L.NewFunction(api.luaUpdate))
	dbTable.RawSetString("delete", L.NewFunction(api.luaDelete))
	dbTable.RawSetString("create_index", L.NewFunction(api.luaCreateIndex))
	dbTable.RawSetString("transaction", L.NewFunction(api.luaTransaction))
	dbTable.RawSetString("ulid", L.NewFunction(luaULID))
	dbTable.RawSetString("timestamp", L.NewFunction(luaTimestamp))
	dbTable.RawSetString("define_table", L.NewFunction(
		luaDefineTable(L, api.pluginName, api.conn, api.dialect, api.tableReg),
	))

	registerConditionConstructors(L, dbTable)
	registerAggregateConstructors(L, dbTable)

	L.SetGlobal("db", dbTable)
}

// ===== CONDITION CONSTRUCTORS =====

// registerConditionConstructors adds db.eq, db.neq, db.gt, etc. to the db Lua table.
// Each returns a sentinel table {__op = "<op>", __val = <value>} that the Go-side
// condition resolver recognizes and converts to db.ColumnOp.
func registerConditionConstructors(L *lua.LState, dbTable *lua.LTable) {
	// Unary constructors: db.eq(v), db.neq(v), etc.
	for _, entry := range []struct {
		name string
		op   string
	}{
		{"eq", "eq"},
		{"neq", "neq"},
		{"gt", "gt"},
		{"gte", "gte"},
		{"lt", "lt"},
		{"lte", "lte"},
		{"like", "like"},
		{"not_like", "not_like"},
	} {
		op := entry.op // capture for closure
		dbTable.RawSetString(entry.name, L.NewFunction(func(L *lua.LState) int {
			val := L.Get(1)
			if val == lua.LNil {
				L.ArgError(1, "condition value cannot be nil")
				return 0
			}
			tbl := L.NewTable()
			tbl.RawSetString(conditionSentinelKey, lua.LString(op))
			tbl.RawSetString(conditionValueKey, val)
			L.Push(tbl)
			return 1
		}))
	}

	// Variadic constructors: db.in_list(v1, v2, ...), db.not_in(v1, v2, ...)
	for _, entry := range []struct {
		name string
		op   string
	}{
		{"in_list", "in"},
		{"not_in", "not_in"},
	} {
		op := entry.op
		dbTable.RawSetString(entry.name, L.NewFunction(func(L *lua.LState) int {
			n := L.GetTop()
			if n == 0 {
				L.ArgError(1, op+" requires at least one value")
				return 0
			}
			valsTbl := L.NewTable()
			for i := 1; i <= n; i++ {
				valsTbl.Append(L.Get(i))
			}
			tbl := L.NewTable()
			tbl.RawSetString(conditionSentinelKey, lua.LString(op))
			tbl.RawSetString(conditionValueKey, valsTbl)
			L.Push(tbl)
			return 1
		}))
	}

	// Binary constructor: db.between(low, high)
	dbTable.RawSetString("between", L.NewFunction(func(L *lua.LState) int {
		low := L.Get(1)
		high := L.Get(2)
		if low == lua.LNil || high == lua.LNil {
			L.ArgError(1, "between requires two non-nil values")
			return 0
		}
		valsTbl := L.NewTable()
		valsTbl.Append(low)
		valsTbl.Append(high)
		tbl := L.NewTable()
		tbl.RawSetString(conditionSentinelKey, lua.LString("between"))
		tbl.RawSetString(conditionValueKey, valsTbl)
		L.Push(tbl)
		return 1
	}))

	// Nullary constructors: db.is_null(), db.is_not_null()
	dbTable.RawSetString("is_null", L.NewFunction(func(L *lua.LState) int {
		tbl := L.NewTable()
		tbl.RawSetString(conditionSentinelKey, lua.LString("is_null"))
		L.Push(tbl)
		return 1
	}))

	dbTable.RawSetString("is_not_null", L.NewFunction(func(L *lua.LState) int {
		tbl := L.NewTable()
		tbl.RawSetString(conditionSentinelKey, lua.LString("is_not_null"))
		L.Push(tbl)
		return 1
	}))
}

// ===== AGGREGATE CONSTRUCTORS =====

// Sentinel keys for aggregate constructor tables.
const (
	aggregateSentinelKey = "__agg"
	aggregateArgKey      = "__arg"
	aggregateAliasKey    = "__alias"
)

// registerAggregateConstructors adds db.agg_count, db.agg_sum, db.agg_avg,
// db.agg_min, db.agg_max to the db Lua table.
// Each returns a sentinel table {__agg = "FUNC", __arg = arg, __alias = alias}
// that the Go-side column parser recognizes and converts to db.AggregateColumn.
// Uses the agg_ prefix to avoid overwriting the existing db.count (luaCount) function.
func registerAggregateConstructors(L *lua.LState, dbTable *lua.LTable) {
	for _, entry := range []struct {
		name string
		fn   string
	}{
		{"agg_count", "COUNT"},
		{"agg_sum", "SUM"},
		{"agg_avg", "AVG"},
		{"agg_min", "MIN"},
		{"agg_max", "MAX"},
	} {
		fn := entry.fn // capture for closure
		dbTable.RawSetString(entry.name, L.NewFunction(func(L *lua.LState) int {
			arg := L.CheckString(1)
			alias := ""
			if L.GetTop() >= 2 {
				if s, ok := L.Get(2).(lua.LString); ok {
					alias = string(s)
				}
			}
			tbl := L.NewTable()
			tbl.RawSetString(aggregateSentinelKey, lua.LString(fn))
			tbl.RawSetString(aggregateArgKey, lua.LString(arg))
			tbl.RawSetString(aggregateAliasKey, lua.LString(alias))
			L.Push(tbl)
			return 1
		}))
	}
}

// ===== CONDITION RESOLUTION =====

// resolveConditions walks a where map and converts any sentinel tables
// (from Lua condition constructors) into db.ColumnOp values.
// Plain values and nil pass through unchanged.
func resolveConditions(where map[string]any) (map[string]any, error) {
	if where == nil {
		return nil, nil
	}
	result := make(map[string]any, len(where))
	for k, v := range where {
		resolved, err := resolveOneValue(v)
		if err != nil {
			return nil, fmt.Errorf("column %q: %w", k, err)
		}
		result[k] = resolved
	}
	return result, nil
}

// resolveOneValue checks if a value is a condition sentinel map and converts it.
func resolveOneValue(v any) (any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return v, nil
	}
	opRaw, hasOp := m[conditionSentinelKey]
	if !hasOp {
		return v, nil
	}
	opStr, ok := opRaw.(string)
	if !ok {
		return nil, fmt.Errorf("invalid condition: __op must be a string")
	}
	if !validConditionOps[opStr] {
		return nil, fmt.Errorf("invalid condition operator %q", opStr)
	}
	val := m[conditionValueKey] // may be nil for is_null/is_not_null
	return luaOpToCondition(opStr, val)
}

// luaOpToCondition converts a Lua operator name and value into a db.ColumnOp.
func luaOpToCondition(op string, val any) (db.ColumnOp, error) {
	switch op {
	case "eq":
		return db.Eq(val), nil
	case "neq":
		return db.Neq(val), nil
	case "gt":
		return db.Gt(val), nil
	case "gte":
		return db.Gte(val), nil
	case "lt":
		return db.Lt(val), nil
	case "lte":
		return db.Lte(val), nil
	case "like":
		s, ok := val.(string)
		if !ok {
			return db.ColumnOp{}, fmt.Errorf("like requires a string value")
		}
		return db.Like(s), nil
	case "not_like":
		s, ok := val.(string)
		if !ok {
			return db.ColumnOp{}, fmt.Errorf("not_like requires a string value")
		}
		return db.NotLike(s), nil
	case "in":
		vals, err := toAnySlice(val)
		if err != nil {
			return db.ColumnOp{}, fmt.Errorf("in: %w", err)
		}
		if len(vals) == 0 {
			return db.ColumnOp{}, fmt.Errorf("in requires at least one value")
		}
		return db.In(vals...), nil
	case "not_in":
		vals, err := toAnySlice(val)
		if err != nil {
			return db.ColumnOp{}, fmt.Errorf("not_in: %w", err)
		}
		if len(vals) == 0 {
			return db.ColumnOp{}, fmt.Errorf("not_in requires at least one value")
		}
		return db.NotIn(vals...), nil
	case "between":
		vals, err := toAnySlice(val)
		if err != nil {
			return db.ColumnOp{}, fmt.Errorf("between: %w", err)
		}
		if len(vals) != 2 {
			return db.ColumnOp{}, fmt.Errorf("between requires exactly 2 values, got %d", len(vals))
		}
		return db.Between(vals[0], vals[1]), nil
	case "is_null":
		return db.IsNull(), nil
	case "is_not_null":
		return db.IsNotNull(), nil
	default:
		return db.ColumnOp{}, fmt.Errorf("unsupported operator %q", op)
	}
}

// toAnySlice converts a value to []any. Accepts []any directly or extracts from
// a map-represented Lua sequence (integer-keyed maps from LuaValueToGo).
func toAnySlice(val any) ([]any, error) {
	if val == nil {
		return nil, fmt.Errorf("expected a list value, got nil")
	}
	if s, ok := val.([]any); ok {
		return s, nil
	}
	return nil, fmt.Errorf("expected a list value, got %T", val)
}

// ===== PARSING FUNCTIONS =====

// parsedSelectOpts collects all parsed fields from a Lua opts table for SELECT queries.
type parsedSelectOpts struct {
	columns          []string
	aggregates       []db.AggregateColumn
	where            map[string]any
	filter           db.Condition // non-nil when new-style condition syntax is used
	whereOr          []map[string]any
	joins            []db.JoinClause
	orderBy          string
	desc             bool
	orders           []db.OrderByClause
	groupBy          []string
	having           map[string]any
	havingCondition  db.Condition // non-nil when condition-style HAVING is used (filtered path)
	distinct         bool
	limit            int64
	offset           int64
}

// parseSelectOpts extracts all SELECT-related fields from a Lua opts table.
// pluginName is needed for JOIN table prefixing.
func parseSelectOpts(L *lua.LState, optsTbl *lua.LTable, pluginName string) (parsedSelectOpts, error) {
	var opts parsedSelectOpts

	whereMap, filter, err := parseWhereExtended(L, optsTbl)
	if err != nil {
		return opts, err
	}
	opts.where = whereMap
	opts.filter = filter

	whereOr, err := parseWhereOrFromLua(L, optsTbl)
	if err != nil {
		return opts, err
	}
	opts.whereOr = whereOr

	cols, aggs, err := parseColumnsFromLua(L, optsTbl, pluginName)
	if err != nil {
		return opts, err
	}
	opts.columns = cols
	opts.aggregates = aggs

	orderBy, desc, orders, err := parseOrderByFromLua(L, optsTbl, pluginName)
	if err != nil {
		return opts, err
	}
	opts.orderBy = orderBy
	opts.desc = desc
	opts.orders = orders

	groupBy, err := parseGroupByFromLua(L, optsTbl, pluginName)
	if err != nil {
		return opts, err
	}
	opts.groupBy = groupBy

	// HAVING dispatch: condition-style for filtered path, map-style for legacy.
	if opts.filter != nil {
		havingCond, err := parseHavingCondition(L, optsTbl)
		if err != nil {
			return opts, err
		}
		opts.havingCondition = havingCond
	} else {
		having, err := parseHavingFromLua(L, optsTbl)
		if err != nil {
			return opts, err
		}
		opts.having = having
	}

	opts.distinct = parseBoolField(L, optsTbl, "distinct")

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

	joins, err := parseJoinsFromLua(L, optsTbl, pluginName)
	if err != nil {
		return opts, err
	}
	opts.joins = joins

	return opts, nil
}

// parseWhereOrFromLua extracts the "where_or" field: a Lua sequence of where-maps.
// Each map is resolved for condition sentinels.
func parseWhereOrFromLua(L *lua.LState, optsTbl *lua.LTable) ([]map[string]any, error) {
	orVal := L.GetField(optsTbl, "where_or")
	if orVal == lua.LNil {
		return nil, nil
	}
	orTbl, ok := orVal.(*lua.LTable)
	if !ok {
		return nil, nil
	}

	var groups []map[string]any
	orTbl.ForEach(func(key, value lua.LValue) {
		if _, ok := key.(lua.LNumber); !ok {
			return
		}
		groupTbl, ok := value.(*lua.LTable)
		if !ok {
			return
		}
		m := LuaTableToMap(L, groupTbl)
		groups = append(groups, m)
	})

	// Resolve conditions in each group.
	resolved := make([]map[string]any, 0, len(groups))
	for _, g := range groups {
		r, err := resolveConditions(g)
		if err != nil {
			return nil, fmt.Errorf("where_or group: %w", err)
		}
		resolved = append(resolved, r)
	}
	return resolved, nil
}

// parseColumnsFromLua extracts the "columns" field: a Lua sequence of column name strings
// and/or aggregate sentinel tables ({__agg, __arg, __alias}).
// Qualified column names (e.g., "tasks.title") have the table part auto-prefixed.
// Returns plain columns and aggregate columns separately.
func parseColumnsFromLua(L *lua.LState, optsTbl *lua.LTable, pluginName string) ([]string, []db.AggregateColumn, error) {
	colsVal := L.GetField(optsTbl, "columns")
	if colsVal == lua.LNil {
		return nil, nil, nil
	}
	colsTbl, ok := colsVal.(*lua.LTable)
	if !ok {
		return nil, nil, nil
	}

	var cols []string
	var aggs []db.AggregateColumn
	var parseErr error
	colsTbl.ForEach(func(key, value lua.LValue) {
		if parseErr != nil {
			return
		}
		if _, ok := key.(lua.LNumber); !ok {
			return
		}

		// Plain string column.
		if s, ok := value.(lua.LString); ok {
			prefixed, err := prefixQualifiedColumn(pluginName, string(s))
			if err != nil {
				parseErr = fmt.Errorf("columns: %w", err)
				return
			}
			cols = append(cols, prefixed)
			return
		}

		// Aggregate sentinel table.
		if tbl, ok := value.(*lua.LTable); ok {
			aggVal := tbl.RawGetString(aggregateSentinelKey)
			if aggVal == lua.LNil {
				parseErr = fmt.Errorf("columns: table element is not an aggregate sentinel (missing %s key)", aggregateSentinelKey)
				return
			}
			fn, ok := aggVal.(lua.LString)
			if !ok {
				parseErr = fmt.Errorf("columns: %s must be a string", aggregateSentinelKey)
				return
			}

			argVal := tbl.RawGetString(aggregateArgKey)
			arg := ""
			if s, ok := argVal.(lua.LString); ok {
				arg = string(s)
			}

			aliasVal := tbl.RawGetString(aggregateAliasKey)
			alias := ""
			if s, ok := aliasVal.(lua.LString); ok {
				alias = string(s)
			}

			// Prefix qualified aggregate args (contain dot, not "*").
			if arg != "*" && strings.Contains(arg, ".") {
				prefixed, err := prefixQualifiedColumn(pluginName, arg)
				if err != nil {
					parseErr = fmt.Errorf("columns: aggregate arg: %w", err)
					return
				}
				arg = prefixed
			}

			aggs = append(aggs, db.AggregateColumn{
				Func:  string(fn),
				Arg:   arg,
				Alias: alias,
			})
			return
		}
	})
	if parseErr != nil {
		return nil, nil, parseErr
	}
	return cols, aggs, nil
}

// parseOrderByFromLua extracts the "order_by" field.
// Supports both legacy string format and new table-of-tables format.
// Qualified column names in table format have the table part auto-prefixed.
// Returns (orderBy, desc, orders, error).
func parseOrderByFromLua(L *lua.LState, optsTbl *lua.LTable, pluginName string) (string, bool, []db.OrderByClause, error) {
	obVal := L.GetField(optsTbl, "order_by")
	if obVal == lua.LNil {
		return "", false, nil, nil
	}

	// Legacy: plain string (not prefixed -- unqualified column name)
	if s, ok := obVal.(lua.LString); ok {
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
		prefixed, err := prefixQualifiedColumn(pluginName, string(colStr))
		if err != nil {
			parseErr = fmt.Errorf("order_by: %w", err)
			return
		}
		descVal := L.GetField(entryTbl, "desc")
		isDesc := false
		if b, ok := descVal.(lua.LBool); ok {
			isDesc = bool(b)
		}
		orders = append(orders, db.OrderByClause{
			Column: prefixed,
			Desc:   isDesc,
		})
	})
	if parseErr != nil {
		return "", false, nil, parseErr
	}
	return "", false, orders, nil
}

// parseGroupByFromLua extracts the "group_by" field: a Lua sequence of column name strings.
// Qualified column names have the table part auto-prefixed.
func parseGroupByFromLua(L *lua.LState, optsTbl *lua.LTable, pluginName string) ([]string, error) {
	gbVal := L.GetField(optsTbl, "group_by")
	if gbVal == lua.LNil {
		return nil, nil
	}
	gbTbl, ok := gbVal.(*lua.LTable)
	if !ok {
		return nil, nil
	}

	var cols []string
	var parseErr error
	gbTbl.ForEach(func(key, value lua.LValue) {
		if parseErr != nil {
			return
		}
		if _, ok := key.(lua.LNumber); !ok {
			return
		}
		if s, ok := value.(lua.LString); ok {
			prefixed, err := prefixQualifiedColumn(pluginName, string(s))
			if err != nil {
				parseErr = fmt.Errorf("group_by: %w", err)
				return
			}
			cols = append(cols, prefixed)
		}
	})
	if parseErr != nil {
		return nil, parseErr
	}
	return cols, nil
}

// parseHavingFromLua extracts the "having" field and resolves condition sentinels.
func parseHavingFromLua(L *lua.LState, optsTbl *lua.LTable) (map[string]any, error) {
	hvVal := L.GetField(optsTbl, "having")
	if hvVal == lua.LNil {
		return nil, nil
	}
	hvTbl, ok := hvVal.(*lua.LTable)
	if !ok {
		return nil, nil
	}
	raw := LuaTableToMap(L, hvTbl)
	return resolveConditions(raw)
}

// parseBoolField extracts an optional boolean field from a Lua table.
func parseBoolField(L *lua.LState, optsTbl *lua.LTable, field string) bool {
	val := L.GetField(optsTbl, field)
	if b, ok := val.(lua.LBool); ok {
		return bool(b)
	}
	return false
}

// parseJoinsFromLua extracts the "joins" field: a Lua sequence of join definition tables.
// Each join table has: type (string), table (string), local_col (string), foreign_col (string).
// Tables and qualified columns are auto-prefixed with the plugin namespace.
func parseJoinsFromLua(L *lua.LState, optsTbl *lua.LTable, pluginName string) ([]db.JoinClause, error) {
	jVal := L.GetField(optsTbl, "joins")
	if jVal == lua.LNil {
		return nil, nil
	}
	jTbl, ok := jVal.(*lua.LTable)
	if !ok {
		return nil, nil
	}

	var joins []db.JoinClause
	var parseErr error
	jTbl.ForEach(func(key, value lua.LValue) {
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

		// Type
		joinTypeStr := luaStringField(L, entryTbl, "type")
		var joinType db.JoinType
		switch strings.ToLower(joinTypeStr) {
		case "inner", "":
			joinType = db.InnerJoin
		case "left":
			joinType = db.LeftJoin
		case "right":
			joinType = db.RightJoin
		default:
			parseErr = fmt.Errorf("invalid join type %q (must be inner, left, or right)", joinTypeStr)
			return
		}

		// Table (unprefixed from Lua)
		tableName := luaStringField(L, entryTbl, "table")
		if tableName == "" {
			parseErr = fmt.Errorf("join entry missing 'table' field")
			return
		}
		prefixedTable, err := prefixTable(pluginName, tableName)
		if err != nil {
			parseErr = fmt.Errorf("join table: %w", err)
			return
		}
		if err := validateJoinTable(pluginName, prefixedTable); err != nil {
			parseErr = err
			return
		}

		// Local column (may be qualified)
		localCol := luaStringField(L, entryTbl, "local_col")
		if localCol == "" {
			parseErr = fmt.Errorf("join entry missing 'local_col' field")
			return
		}
		localCol, err = prefixQualifiedColumn(pluginName, localCol)
		if err != nil {
			parseErr = fmt.Errorf("join local_col: %w", err)
			return
		}

		// Foreign column (may be qualified)
		foreignCol := luaStringField(L, entryTbl, "foreign_col")
		if foreignCol == "" {
			parseErr = fmt.Errorf("join entry missing 'foreign_col' field")
			return
		}
		foreignCol, err = prefixQualifiedColumn(pluginName, foreignCol)
		if err != nil {
			parseErr = fmt.Errorf("join foreign_col: %w", err)
			return
		}

		joins = append(joins, db.JoinClause{
			Type:       joinType,
			Table:      prefixedTable,
			LocalCol:   localCol,
			ForeignCol: foreignCol,
		})
	})
	if parseErr != nil {
		return nil, parseErr
	}
	return joins, nil
}

// ===== HANDLER FUNCTIONS =====

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

	var opts parsedSelectOpts
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			parsed, err := parseSelectOpts(L, optsTbl, api.pluginName)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			opts = parsed
		}
	}

	// Apply default limit if none specified.
	if opts.limit <= 0 {
		opts.limit = int64(api.maxRows)
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.query: no context set on Lua state (VM setup error)")
		return 0
	}

	var rows []db.Row
	useFilteredPath := opts.filter != nil || opts.havingCondition != nil || len(opts.aggregates) > 0
	if useFilteredPath {
		// New condition-based path (also used for aggregate queries without WHERE).
		var orderByCols []db.OrderByColumn
		for _, o := range opts.orders {
			orderByCols = append(orderByCols, db.OrderByColumn{Column: o.Column, Desc: o.Desc})
		}
		if opts.orderBy != "" && len(orderByCols) == 0 {
			orderByCols = []db.OrderByColumn{{Column: opts.orderBy, Desc: opts.desc}}
		}
		rows, err = db.QSelectFiltered(ctx, api.currentExec, api.dialect, db.FilteredSelectParams{
			Table:       prefixed,
			Columns:     opts.columns,
			Aggregates:  opts.aggregates,
			Filter:      opts.filter,
			GroupBy:     opts.groupBy,
			Having:      opts.havingCondition,
			OrderByCols: orderByCols,
			Distinct:    opts.distinct,
			Limit:       opts.limit,
			Offset:      opts.offset,
		})
	} else {
		// Old map-based path.
		rows, err = db.QSelect(ctx, api.currentExec, api.dialect, db.SelectParams{
			Table:      prefixed,
			Columns:    opts.columns,
			Aggregates: opts.aggregates,
			Where:      opts.where,
			WhereOr:    opts.whereOr,
			Joins:      opts.joins,
			OrderBy:    opts.orderBy,
			Desc:       opts.desc,
			Orders:     opts.orders,
			GroupBy:    opts.groupBy,
			Having:     opts.having,
			Distinct:   opts.distinct,
			Limit:      opts.limit,
			Offset:     opts.offset,
		})
	}
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Always return a table (never nil), even for empty results.
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

	var opts parsedSelectOpts
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			parsed, err := parseSelectOpts(L, optsTbl, api.pluginName)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			opts = parsed
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.query_one: no context set on Lua state (VM setup error)")
		return 0
	}

	var row db.Row
	useFilteredPath := opts.filter != nil || opts.havingCondition != nil || len(opts.aggregates) > 0
	if useFilteredPath {
		row, err = db.QSelectOneFiltered(ctx, api.currentExec, api.dialect, db.FilteredSelectParams{
			Table:      prefixed,
			Columns:    opts.columns,
			Aggregates: opts.aggregates,
			Filter:     opts.filter,
			GroupBy:    opts.groupBy,
			Having:     opts.havingCondition,
			Distinct:   opts.distinct,
		})
	} else {
		row, err = db.QSelectOne(ctx, api.currentExec, api.dialect, db.SelectParams{
			Table:      prefixed,
			Columns:    opts.columns,
			Aggregates: opts.aggregates,
			Where:      opts.where,
			WhereOr:    opts.whereOr,
			Joins:      opts.joins,
			OrderBy:    opts.orderBy,
			Desc:       opts.desc,
			Orders:     opts.orders,
			GroupBy:    opts.groupBy,
			Having:     opts.having,
			Distinct:   opts.distinct,
		})
	}
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

	var whereMap map[string]any
	var filter db.Condition
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			whereMap, filter, err = parseWhereExtended(L, optsTbl)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.count: no context set on Lua state (VM setup error)")
		return 0
	}

	var count int64
	if filter != nil {
		count, err = db.QCountFiltered(ctx, api.currentExec, api.dialect, prefixed, filter)
	} else {
		count, err = db.QCount(ctx, api.currentExec, api.dialect, prefixed, whereMap)
	}
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

	var whereMap map[string]any
	var filter db.Condition
	if L.GetTop() >= 2 {
		optsVal := L.Get(2)
		if optsTbl, ok := optsVal.(*lua.LTable); ok {
			whereMap, filter, err = parseWhereExtended(L, optsTbl)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.exists: no context set on Lua state (VM setup error)")
		return 0
	}

	var exists bool
	if filter != nil {
		exists, err = db.QExistsFiltered(ctx, api.currentExec, api.dialect, prefixed, filter)
	} else {
		exists, err = db.QExists(ctx, api.currentExec, api.dialect, prefixed, whereMap)
	}
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

// luaUpsert implements db.upsert(table, opts).
// Follows luaUpdate pattern for argument parsing (CheckString(1) + CheckTable(2) with named fields).
// Auto-sets updated_at in values if missing. Does NOT auto-set id or created_at.
// Returns nothing on success, nil+errmsg on error. checkOpLimit and nil context use L.RaiseError.
func (api *DatabaseAPI) luaUpsert(L *lua.LState) int {
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

	// Parse values (required).
	valuesVal := L.GetField(optsTbl, "values")
	valuesTbl, ok := valuesVal.(*lua.LTable)
	if !ok || valuesVal == lua.LNil {
		L.ArgError(2, "upsert requires a 'values' table")
		return 0
	}
	values := LuaTableToMap(L, valuesTbl)
	if len(values) == 0 {
		L.ArgError(2, "upsert 'values' cannot be empty")
		return 0
	}

	// Parse conflict_columns (required).
	ccVal := L.GetField(optsTbl, "conflict_columns")
	ccTbl, ok := ccVal.(*lua.LTable)
	if !ok || ccVal == lua.LNil {
		L.ArgError(2, "upsert requires a 'conflict_columns' table")
		return 0
	}
	var conflictColumns []string
	ccTbl.ForEach(func(k, v lua.LValue) {
		if _, isInt := k.(lua.LNumber); isInt {
			s, isStr := v.(lua.LString)
			if !isStr {
				L.ArgError(2, "conflict_columns values must be strings")
				return
			}
			conflictColumns = append(conflictColumns, string(s))
		}
	})
	if len(conflictColumns) == 0 {
		L.ArgError(2, "conflict_columns cannot be empty")
		return 0
	}

	// Parse update (optional).
	var updateMap map[string]any
	updateVal := L.GetField(optsTbl, "update")
	if updateVal != lua.LNil {
		if updateTbl, ok := updateVal.(*lua.LTable); ok {
			updateMap = LuaTableToMap(L, updateTbl)
		}
	}

	// Parse do_nothing (optional).
	doNothing := false
	dnVal := L.GetField(optsTbl, "do_nothing")
	if dnVal != lua.LNil {
		doNothing = lua.LVAsBool(dnVal)
	}

	// Auto-set updated_at in values if missing.
	now := time.Now().UTC().Format(time.RFC3339)
	if _, exists := values["updated_at"]; !exists {
		values["updated_at"] = now
	}

	// Auto-set updated_at in explicit update if present and not DoNothing.
	if updateMap != nil && !doNothing {
		if _, exists := updateMap["updated_at"]; !exists {
			updateMap["updated_at"] = now
		}
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.upsert: no context set on Lua state (VM setup error)")
		return 0
	}

	_, err = db.QUpsert(ctx, api.currentExec, api.dialect, db.UpsertParams{
		Table:           prefixed,
		Values:          values,
		ConflictColumns: conflictColumns,
		Update:          updateMap,
		DoNothing:       doNothing,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// luaUpdate implements db.update(table, opts).
// Requires non-empty 'set' and at least one of 'where' or 'where_or' non-empty.
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

	// Parse where with extended detection (old map or new condition).
	whereMap, filter, err := parseWhereExtended(L, optsTbl)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
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

	if filter != nil {
		// New condition path.
		_, err = db.QUpdateFiltered(ctx, api.currentExec, api.dialect, db.FilteredUpdateParams{
			Table:  prefixed,
			Set:    setMap,
			Filter: filter,
		})
	} else {
		// Old map path — parse where_or and enforce safety.
		whereOr, woErr := parseWhereOrFromLua(L, optsTbl)
		if woErr != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(woErr.Error()))
			return 2
		}
		if len(whereMap) == 0 && len(whereOr) == 0 {
			L.RaiseError("db.update requires non-empty where or where_or (safety: prevents full-table update)")
			return 0
		}
		_, err = db.QUpdate(ctx, api.currentExec, api.dialect, db.UpdateParams{
			Table:   prefixed,
			Set:     setMap,
			Where:   whereMap,
			WhereOr: whereOr,
		})
	}
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	return 0
}

// luaDelete implements db.delete(table, opts).
// Requires at least one of 'where' or 'where_or' to be non-empty.
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

	// Parse where with extended detection (old map or new condition).
	whereMap, filter, err := parseWhereExtended(L, optsTbl)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("db.delete: no context set on Lua state (VM setup error)")
		return 0
	}

	if filter != nil {
		// New condition path.
		_, err = db.QDeleteFiltered(ctx, api.currentExec, api.dialect, db.FilteredDeleteParams{
			Table:  prefixed,
			Filter: filter,
		})
	} else {
		// Old map path — parse where_or and enforce safety.
		whereOr, woErr := parseWhereOrFromLua(L, optsTbl)
		if woErr != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(woErr.Error()))
			return 2
		}
		if len(whereMap) == 0 && len(whereOr) == 0 {
			L.RaiseError("db.delete requires non-empty where or where_or (safety: prevents full-table delete)")
			return 0
		}
		_, err = db.QDelete(ctx, api.currentExec, api.dialect, db.DeleteParams{
			Table:   prefixed,
			Where:   whereMap,
			WhereOr: whereOr,
		})
	}
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

// luaStringField reads an optional string field from a Lua table.
// Returns empty string if the field is absent, nil, or not a string.
func luaStringField(L *lua.LState, tbl *lua.LTable, field string) string {
	val := L.GetField(tbl, field)
	if s, ok := val.(lua.LString); ok {
		return string(s)
	}
	return ""
}
