package plugin

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"
)

// RegisterJSONAPI creates a "json" global Lua table with encode and decode functions.
// This module is available to all plugins and is frozen after registration to prevent
// modification by plugin code.
//
// Functions:
//   - json.encode(table) -> string: Marshals a Lua table to a JSON string.
//   - json.decode(string) -> table: Unmarshals a JSON string to a Lua table.
func RegisterJSONAPI(L *lua.LState) {
	jsonTable := L.NewTable()
	jsonTable.RawSetString("encode", L.NewFunction(jsonEncode))
	jsonTable.RawSetString("decode", L.NewFunction(jsonDecode))
	L.SetGlobal("json", jsonTable)
}

// jsonEncode converts a Lua value to a JSON string.
// Accepts tables (maps/arrays), strings, numbers, booleans, and nil.
func jsonEncode(L *lua.LState) int {
	val := L.Get(1)
	if val == lua.LNil {
		L.Push(lua.LString("null"))
		return 1
	}

	goVal := LuaValueToGo(val)
	b, err := json.Marshal(goVal)
	if err != nil {
		L.RaiseError("json.encode: %s", err.Error())
		return 0
	}

	L.Push(lua.LString(string(b)))
	return 1
}

// jsonDecode parses a JSON string into a Lua value.
// Objects become Lua tables with string keys. Arrays become Lua sequence tables.
// Strings, numbers, booleans, and null are converted to their Lua equivalents.
func jsonDecode(L *lua.LState) int {
	str := L.CheckString(1)

	var goVal any
	if err := json.Unmarshal([]byte(str), &goVal); err != nil {
		L.RaiseError("json.decode: %s", err.Error())
		return 0
	}

	L.Push(GoValueToLua(L, goVal))
	return 1
}
