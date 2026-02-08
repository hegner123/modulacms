package plugin

import (
	lua "github.com/yuin/gopher-lua"
)

// LuaTableToMap converts a Lua table with string keys to a Go map.
// Non-string keys are skipped. Nested tables are recursively converted
// via LuaValueToGo, which decides between map and slice representation.
func LuaTableToMap(L *lua.LState, tbl *lua.LTable) map[string]any {
	result := make(map[string]any)
	tbl.ForEach(func(key, value lua.LValue) {
		if s, ok := key.(lua.LString); ok {
			result[string(s)] = LuaValueToGo(value)
		}
	})
	return result
}

// MapToLuaTable converts a Go map[string]any to a Lua table.
// Values are converted via GoValueToLua.
func MapToLuaTable(L *lua.LState, m map[string]any) *lua.LTable {
	tbl := L.NewTable()
	for k, v := range m {
		L.SetField(tbl, k, GoValueToLua(L, v))
	}
	return tbl
}

// GoValueToLua converts a Go value to the corresponding Lua value.
// Supported types:
//   - string        -> lua.LString
//   - int64         -> lua.LNumber
//   - float64       -> lua.LNumber
//   - int           -> lua.LNumber
//   - int32         -> lua.LNumber
//   - bool          -> lua.LBool
//   - nil           -> lua.LNil
//   - []byte        -> lua.LString
//   - map[string]any -> recursive MapToLuaTable
//   - []any         -> Lua sequence table (1-indexed)
//   - []map[string]any -> Lua sequence table of tables (1-indexed)
//
// Unsupported types return lua.LNil. This is intentional: plugin code
// should never see Go-internal types, and silently converting to nil
// is safer than panicking in production.
func GoValueToLua(L *lua.LState, v any) lua.LValue {
	if v == nil {
		return lua.LNil
	}

	switch val := v.(type) {
	case string:
		return lua.LString(val)
	case int64:
		return lua.LNumber(val)
	case float64:
		return lua.LNumber(val)
	case int:
		return lua.LNumber(val)
	case int32:
		return lua.LNumber(val)
	case bool:
		return lua.LBool(val)
	case []byte:
		return lua.LString(string(val))
	case map[string]any:
		return MapToLuaTable(L, val)
	case []any:
		tbl := L.NewTable()
		for _, item := range val {
			tbl.Append(GoValueToLua(L, item))
		}
		return tbl
	case []map[string]any:
		tbl := L.NewTable()
		for _, item := range val {
			tbl.Append(MapToLuaTable(L, item))
		}
		return tbl
	default:
		return lua.LNil
	}
}

// LuaValueToGo converts a Lua value to the corresponding Go value.
// Conversion rules:
//   - lua.LString  -> string
//   - lua.LNumber  -> float64
//   - lua.LBool    -> bool
//   - lua.LNil     -> nil
//   - *lua.LTable  -> map[string]any (if any string key) or []any (if pure sequence)
//
// For tables: if the table has any string keys, it is converted to map[string]any.
// Integer-keyed entries in a mixed table are also included with string-converted keys.
// If the table has only consecutive integer keys starting at 1 (a Lua sequence),
// it is converted to []any.
func LuaValueToGo(v lua.LValue) any {
	if v == nil || v == lua.LNil {
		return nil
	}

	switch val := v.(type) {
	case lua.LBool:
		return bool(val)
	case lua.LNumber:
		return float64(val)
	case *lua.LNilType:
		return nil
	case lua.LString:
		return string(val)
	case *lua.LTable:
		return luaTableToGoValue(val)
	default:
		// Functions, userdata, etc. are not convertible to Go values.
		return nil
	}
}

// luaTableToGoValue determines whether a Lua table is a sequence ([]any)
// or a dictionary (map[string]any) and converts accordingly.
//
// Detection strategy: iterate all keys. If all keys are consecutive integers
// starting at 1 with no string keys, treat as sequence. Otherwise treat as map.
// This matches the Lua convention where # operator and ipairs work on sequences.
func luaTableToGoValue(tbl *lua.LTable) any {
	hasStringKey := false
	maxIntKey := 0
	intKeyCount := 0

	tbl.ForEach(func(key, _ lua.LValue) {
		switch k := key.(type) {
		case lua.LNumber:
			intVal := int(k)
			if float64(intVal) == float64(k) && intVal > 0 {
				intKeyCount++
				if intVal > maxIntKey {
					maxIntKey = intVal
				}
			} else {
				// Non-positive or fractional number key -- treat as mixed
				hasStringKey = true
			}
		default:
			hasStringKey = true
		}
	})

	// Pure sequence: all keys are positive integers 1..N with no gaps
	if !hasStringKey && intKeyCount > 0 && maxIntKey == intKeyCount {
		result := make([]any, intKeyCount)
		for i := 1; i <= intKeyCount; i++ {
			result[i-1] = LuaValueToGo(tbl.RawGetInt(i))
		}
		return result
	}

	// Dictionary (or empty table -- empty map is fine)
	result := make(map[string]any)
	tbl.ForEach(func(key, value lua.LValue) {
		switch k := key.(type) {
		case lua.LString:
			result[string(k)] = LuaValueToGo(value)
		case lua.LNumber:
			// Include integer-keyed values with string-converted keys.
			// This handles mixed tables where both string and integer keys exist.
			result[k.String()] = LuaValueToGo(value)
		}
	})
	return result
}

// RowsToLuaTable converts a slice of db.Row results ([]map[string]any) to a
// Lua sequence table. Each row becomes a Lua table. Empty slices return an
// empty table (never nil). This is the query result contract: plugin authors
// can always use #results, ipairs(results), and results[1] safely.
//
// Design decision: The roadmap specifies SQLRowsToLuaTable(L, *sql.Rows) but
// the query builder already returns []Row ([]map[string]any) from QSelect.
// Converting from the already-scanned Go representation avoids coupling this
// helper to *sql.Rows lifecycle management, which belongs in db_api.go.
func RowsToLuaTable(L *lua.LState, rows []map[string]any) *lua.LTable {
	tbl := L.NewTable()
	for _, row := range rows {
		tbl.Append(MapToLuaTable(L, row))
	}
	return tbl
}
