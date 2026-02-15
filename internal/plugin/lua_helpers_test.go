package plugin

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// newTestState creates a fresh Lua state for testing. Caller must defer L.Close().
func newTestState() *lua.LState {
	return lua.NewState(lua.Options{SkipOpenLibs: true})
}

func TestGoValueToLua(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tests := []struct {
		name     string
		input    any
		checkFn  func(t *testing.T, v lua.LValue)
	}{
		{
			name:  "string",
			input: "hello",
			checkFn: func(t *testing.T, v lua.LValue) {
				s, ok := v.(lua.LString)
				if !ok {
					t.Fatalf("expected LString, got %T", v)
				}
				if string(s) != "hello" {
					t.Fatalf("expected %q, got %q", "hello", string(s))
				}
			},
		},
		{
			name:  "int64",
			input: int64(42),
			checkFn: func(t *testing.T, v lua.LValue) {
				n, ok := v.(lua.LNumber)
				if !ok {
					t.Fatalf("expected LNumber, got %T", v)
				}
				if float64(n) != 42 {
					t.Fatalf("expected 42, got %v", n)
				}
			},
		},
		{
			name:  "float64",
			input: 3.14,
			checkFn: func(t *testing.T, v lua.LValue) {
				n, ok := v.(lua.LNumber)
				if !ok {
					t.Fatalf("expected LNumber, got %T", v)
				}
				if float64(n) != 3.14 {
					t.Fatalf("expected 3.14, got %v", n)
				}
			},
		},
		{
			name:  "int",
			input: 99,
			checkFn: func(t *testing.T, v lua.LValue) {
				n, ok := v.(lua.LNumber)
				if !ok {
					t.Fatalf("expected LNumber, got %T", v)
				}
				if float64(n) != 99 {
					t.Fatalf("expected 99, got %v", n)
				}
			},
		},
		{
			name:  "int32",
			input: int32(7),
			checkFn: func(t *testing.T, v lua.LValue) {
				n, ok := v.(lua.LNumber)
				if !ok {
					t.Fatalf("expected LNumber, got %T", v)
				}
				if float64(n) != 7 {
					t.Fatalf("expected 7, got %v", n)
				}
			},
		},
		{
			name:  "bool true",
			input: true,
			checkFn: func(t *testing.T, v lua.LValue) {
				b, ok := v.(lua.LBool)
				if !ok {
					t.Fatalf("expected LBool, got %T", v)
				}
				if !bool(b) {
					t.Fatal("expected true")
				}
			},
		},
		{
			name:  "bool false",
			input: false,
			checkFn: func(t *testing.T, v lua.LValue) {
				b, ok := v.(lua.LBool)
				if !ok {
					t.Fatalf("expected LBool, got %T", v)
				}
				if bool(b) {
					t.Fatal("expected false")
				}
			},
		},
		{
			name:  "nil",
			input: nil,
			checkFn: func(t *testing.T, v lua.LValue) {
				if v != lua.LNil {
					t.Fatalf("expected LNil, got %T(%v)", v, v)
				}
			},
		},
		{
			name:  "byte slice",
			input: []byte("binary data"),
			checkFn: func(t *testing.T, v lua.LValue) {
				s, ok := v.(lua.LString)
				if !ok {
					t.Fatalf("expected LString, got %T", v)
				}
				if string(s) != "binary data" {
					t.Fatalf("expected %q, got %q", "binary data", string(s))
				}
			},
		},
		{
			name:  "unsupported type returns LNil",
			input: struct{ X int }{X: 1},
			checkFn: func(t *testing.T, v lua.LValue) {
				if v != lua.LNil {
					t.Fatalf("expected LNil for unsupported type, got %T(%v)", v, v)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GoValueToLua(L, tt.input)
			tt.checkFn(t, result)
		})
	}
}

func TestGoValueToLua_Map(t *testing.T) {
	L := newTestState()
	defer L.Close()

	input := map[string]any{
		"name":  "alice",
		"age":   int64(30),
		"admin": true,
	}

	result := GoValueToLua(L, input)
	tbl, ok := result.(*lua.LTable)
	if !ok {
		t.Fatalf("expected *LTable, got %T", result)
	}

	name := L.GetField(tbl, "name")
	if name.String() != "alice" {
		t.Fatalf("expected name=alice, got %v", name)
	}

	age := L.GetField(tbl, "age")
	if n, ok := age.(lua.LNumber); !ok || float64(n) != 30 {
		t.Fatalf("expected age=30, got %v", age)
	}

	admin := L.GetField(tbl, "admin")
	if b, ok := admin.(lua.LBool); !ok || !bool(b) {
		t.Fatalf("expected admin=true, got %v", admin)
	}
}

func TestGoValueToLua_Slice(t *testing.T) {
	L := newTestState()
	defer L.Close()

	input := []any{"a", "b", "c"}
	result := GoValueToLua(L, input)
	tbl, ok := result.(*lua.LTable)
	if !ok {
		t.Fatalf("expected *LTable, got %T", result)
	}

	// Lua sequences are 1-indexed
	if tbl.RawGetInt(1).String() != "a" {
		t.Fatalf("expected [1]=a, got %v", tbl.RawGetInt(1))
	}
	if tbl.RawGetInt(2).String() != "b" {
		t.Fatalf("expected [2]=b, got %v", tbl.RawGetInt(2))
	}
	if tbl.RawGetInt(3).String() != "c" {
		t.Fatalf("expected [3]=c, got %v", tbl.RawGetInt(3))
	}

	// Verify Lua length operator equivalent
	if tbl.Len() != 3 {
		t.Fatalf("expected table length 3, got %d", tbl.Len())
	}
}

func TestGoValueToLua_SliceOfMaps(t *testing.T) {
	L := newTestState()
	defer L.Close()

	input := []map[string]any{
		{"id": "1", "name": "alice"},
		{"id": "2", "name": "bob"},
	}

	result := GoValueToLua(L, input)
	tbl, ok := result.(*lua.LTable)
	if !ok {
		t.Fatalf("expected *LTable, got %T", result)
	}

	if tbl.Len() != 2 {
		t.Fatalf("expected 2 rows, got %d", tbl.Len())
	}

	row1 := tbl.RawGetInt(1)
	row1Tbl, ok := row1.(*lua.LTable)
	if !ok {
		t.Fatalf("expected row 1 to be *LTable, got %T", row1)
	}
	if L.GetField(row1Tbl, "name").String() != "alice" {
		t.Fatalf("expected row1.name=alice, got %v", L.GetField(row1Tbl, "name"))
	}

	row2 := tbl.RawGetInt(2)
	row2Tbl, ok := row2.(*lua.LTable)
	if !ok {
		t.Fatalf("expected row 2 to be *LTable, got %T", row2)
	}
	if L.GetField(row2Tbl, "name").String() != "bob" {
		t.Fatalf("expected row2.name=bob, got %v", L.GetField(row2Tbl, "name"))
	}
}

func TestGoValueToLua_NestedMap(t *testing.T) {
	L := newTestState()
	defer L.Close()

	input := map[string]any{
		"user": map[string]any{
			"name": "alice",
			"meta": map[string]any{
				"role": "admin",
			},
		},
	}

	result := GoValueToLua(L, input)
	tbl, ok := result.(*lua.LTable)
	if !ok {
		t.Fatalf("expected *LTable, got %T", result)
	}

	userTbl := L.GetField(tbl, "user")
	if userTbl == lua.LNil {
		t.Fatal("expected user field to be present")
	}

	userT, ok := userTbl.(*lua.LTable)
	if !ok {
		t.Fatalf("expected user to be *LTable, got %T", userTbl)
	}

	if L.GetField(userT, "name").String() != "alice" {
		t.Fatalf("expected user.name=alice, got %v", L.GetField(userT, "name"))
	}

	metaTbl := L.GetField(userT, "meta")
	metaT, ok := metaTbl.(*lua.LTable)
	if !ok {
		t.Fatalf("expected meta to be *LTable, got %T", metaTbl)
	}

	if L.GetField(metaT, "role").String() != "admin" {
		t.Fatalf("expected user.meta.role=admin, got %v", L.GetField(metaT, "role"))
	}
}

func TestLuaValueToGo_Primitives(t *testing.T) {
	tests := []struct {
		name     string
		input    lua.LValue
		expected any
	}{
		{"string", lua.LString("hello"), "hello"},
		{"number", lua.LNumber(42.5), float64(42.5)},
		{"integer number", lua.LNumber(10), float64(10)},
		{"bool true", lua.LTrue, true},
		{"bool false", lua.LFalse, false},
		{"nil", lua.LNil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LuaValueToGo(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestLuaValueToGo_NilInput(t *testing.T) {
	// Passing a Go nil (not lua.LNil) should return nil without panic
	result := LuaValueToGo(nil)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestLuaValueToGo_TableAsMap(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	L.SetField(tbl, "name", lua.LString("alice"))
	L.SetField(tbl, "age", lua.LNumber(30))
	L.SetField(tbl, "active", lua.LTrue)

	result := LuaValueToGo(tbl)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}

	if m["name"] != "alice" {
		t.Fatalf("expected name=alice, got %v", m["name"])
	}
	if m["age"] != float64(30) {
		t.Fatalf("expected age=30, got %v", m["age"])
	}
	if m["active"] != true {
		t.Fatalf("expected active=true, got %v", m["active"])
	}
}

func TestLuaValueToGo_TableAsSequence(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.Append(lua.LString("a"))
	tbl.Append(lua.LString("b"))
	tbl.Append(lua.LNumber(3))

	result := LuaValueToGo(tbl)
	sl, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(sl) != 3 {
		t.Fatalf("expected length 3, got %d", len(sl))
	}
	if sl[0] != "a" {
		t.Fatalf("expected [0]=a, got %v", sl[0])
	}
	if sl[1] != "b" {
		t.Fatalf("expected [1]=b, got %v", sl[1])
	}
	if sl[2] != float64(3) {
		t.Fatalf("expected [2]=3, got %v", sl[2])
	}
}

func TestLuaValueToGo_EmptyTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	result := LuaValueToGo(tbl)

	// Empty table should be a map (consistent with dictionary default)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any for empty table, got %T", result)
	}
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %v", m)
	}
}

func TestLuaValueToGo_NestedTable(t *testing.T) {
	L := newTestState()
	defer L.Close()

	inner := L.NewTable()
	L.SetField(inner, "x", lua.LNumber(1))
	L.SetField(inner, "y", lua.LNumber(2))

	outer := L.NewTable()
	L.SetField(outer, "pos", inner)

	result := LuaValueToGo(outer)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}

	posMap, ok := m["pos"].(map[string]any)
	if !ok {
		t.Fatalf("expected pos to be map[string]any, got %T", m["pos"])
	}
	if posMap["x"] != float64(1) {
		t.Fatalf("expected pos.x=1, got %v", posMap["x"])
	}
	if posMap["y"] != float64(2) {
		t.Fatalf("expected pos.y=2, got %v", posMap["y"])
	}
}

func TestLuaValueToGo_SequenceOfTables(t *testing.T) {
	L := newTestState()
	defer L.Close()

	row1 := L.NewTable()
	L.SetField(row1, "id", lua.LString("1"))
	L.SetField(row1, "name", lua.LString("alice"))

	row2 := L.NewTable()
	L.SetField(row2, "id", lua.LString("2"))
	L.SetField(row2, "name", lua.LString("bob"))

	outer := L.NewTable()
	outer.Append(row1)
	outer.Append(row2)

	result := LuaValueToGo(outer)
	sl, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(sl) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(sl))
	}

	r1, ok := sl[0].(map[string]any)
	if !ok {
		t.Fatalf("expected row 0 to be map[string]any, got %T", sl[0])
	}
	if r1["name"] != "alice" {
		t.Fatalf("expected row0.name=alice, got %v", r1["name"])
	}
}

func TestRoundTrip_GoMapToLuaAndBack(t *testing.T) {
	L := newTestState()
	defer L.Close()

	original := map[string]any{
		"name":    "alice",
		"age":     float64(30),
		"active":  true,
		"score":   float64(95.5),
		"nothing": nil,
	}

	luaTbl := MapToLuaTable(L, original)
	roundTripped := LuaTableToMap(L, luaTbl)

	// Check each field individually since nil values won't be in the round-tripped map
	// (Lua nil values are not iterated by ForEach)
	if roundTripped["name"] != "alice" {
		t.Fatalf("name: expected alice, got %v", roundTripped["name"])
	}
	if roundTripped["age"] != float64(30) {
		t.Fatalf("age: expected 30, got %v", roundTripped["age"])
	}
	if roundTripped["active"] != true {
		t.Fatalf("active: expected true, got %v", roundTripped["active"])
	}
	if roundTripped["score"] != float64(95.5) {
		t.Fatalf("score: expected 95.5, got %v", roundTripped["score"])
	}

	// nil values in Go maps become lua.LNil in the table, which ForEach skips.
	// So "nothing" should NOT appear in the round-tripped map.
	if _, exists := roundTripped["nothing"]; exists {
		t.Fatalf("nil values should not survive round-trip (Lua nil is not iterable)")
	}
}

func TestRoundTrip_TypePreservation(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tests := []struct {
		name     string
		input    map[string]any
		key      string
		expected any
	}{
		{
			name:     "string preserves",
			input:    map[string]any{"val": "hello"},
			key:      "val",
			expected: "hello",
		},
		{
			name:     "float64 preserves",
			input:    map[string]any{"val": float64(42)},
			key:      "val",
			expected: float64(42),
		},
		{
			name:     "int64 becomes float64 through Lua",
			input:    map[string]any{"val": int64(42)},
			key:      "val",
			expected: float64(42),
		},
		{
			name:     "int becomes float64 through Lua",
			input:    map[string]any{"val": 42},
			key:      "val",
			expected: float64(42),
		},
		{
			name:     "bool true preserves",
			input:    map[string]any{"val": true},
			key:      "val",
			expected: true,
		},
		{
			name:     "bool false preserves",
			input:    map[string]any{"val": false},
			key:      "val",
			expected: false,
		},
		{
			name:     "byte slice becomes string",
			input:    map[string]any{"val": []byte("bytes")},
			key:      "val",
			expected: "bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			luaTbl := MapToLuaTable(L, tt.input)
			result := LuaTableToMap(L, luaTbl)
			if result[tt.key] != tt.expected {
				t.Fatalf("expected %v (%T), got %v (%T)",
					tt.expected, tt.expected, result[tt.key], result[tt.key])
			}
		})
	}
}

func TestRoundTrip_NestedMap(t *testing.T) {
	L := newTestState()
	defer L.Close()

	original := map[string]any{
		"outer": map[string]any{
			"inner": "value",
			"num":   float64(7),
		},
	}

	luaTbl := MapToLuaTable(L, original)
	result := LuaTableToMap(L, luaTbl)

	outerMap, ok := result["outer"].(map[string]any)
	if !ok {
		t.Fatalf("expected outer to be map[string]any, got %T", result["outer"])
	}
	if outerMap["inner"] != "value" {
		t.Fatalf("expected inner=value, got %v", outerMap["inner"])
	}
	if outerMap["num"] != float64(7) {
		t.Fatalf("expected num=7, got %v", outerMap["num"])
	}
}

func TestRowsToLuaTable_EmptySlice(t *testing.T) {
	L := newTestState()
	defer L.Close()

	result := RowsToLuaTable(L, []map[string]any{})

	if result == nil {
		t.Fatal("RowsToLuaTable must return empty table, never nil")
	}
	if result.Len() != 0 {
		t.Fatalf("expected empty table, got length %d", result.Len())
	}
}

func TestRowsToLuaTable_NilSlice(t *testing.T) {
	L := newTestState()
	defer L.Close()

	result := RowsToLuaTable(L, nil)

	if result == nil {
		t.Fatal("RowsToLuaTable must return empty table, never nil")
	}
	if result.Len() != 0 {
		t.Fatalf("expected empty table, got length %d", result.Len())
	}
}

func TestRowsToLuaTable_MultipleRows(t *testing.T) {
	L := newTestState()
	defer L.Close()

	rows := []map[string]any{
		{"id": "1", "name": "alice", "score": float64(95)},
		{"id": "2", "name": "bob", "score": float64(87)},
		{"id": "3", "name": "carol", "score": float64(91)},
	}

	result := RowsToLuaTable(L, rows)

	if result.Len() != 3 {
		t.Fatalf("expected 3 rows, got %d", result.Len())
	}

	// Verify order preservation (1-indexed Lua sequence)
	for i, expected := range rows {
		luaIdx := i + 1
		rowVal := result.RawGetInt(luaIdx)
		rowTbl, ok := rowVal.(*lua.LTable)
		if !ok {
			t.Fatalf("row %d: expected *LTable, got %T", luaIdx, rowVal)
		}

		name := L.GetField(rowTbl, "name")
		if name.String() != expected["name"] {
			t.Fatalf("row %d: expected name=%v, got %v", luaIdx, expected["name"], name)
		}

		id := L.GetField(rowTbl, "id")
		if id.String() != expected["id"] {
			t.Fatalf("row %d: expected id=%v, got %v", luaIdx, expected["id"], id)
		}

		score := L.GetField(rowTbl, "score")
		if scoreNum, ok := score.(lua.LNumber); !ok || float64(scoreNum) != expected["score"] {
			t.Fatalf("row %d: expected score=%v, got %v", luaIdx, expected["score"], score)
		}
	}
}

func TestRowsToLuaTable_RowWithNilValue(t *testing.T) {
	L := newTestState()
	defer L.Close()

	rows := []map[string]any{
		{"id": "1", "name": "alice", "optional": nil},
	}

	result := RowsToLuaTable(L, rows)
	if result.Len() != 1 {
		t.Fatalf("expected 1 row, got %d", result.Len())
	}

	rowTbl := result.RawGetInt(1).(*lua.LTable)
	optVal := L.GetField(rowTbl, "optional")
	if optVal != lua.LNil {
		t.Fatalf("expected optional=LNil, got %T(%v)", optVal, optVal)
	}
}

func TestRowsToLuaTable_RowWithByteSlice(t *testing.T) {
	L := newTestState()
	defer L.Close()

	rows := []map[string]any{
		{"id": "1", "data": []byte("binary content")},
	}

	result := RowsToLuaTable(L, rows)
	rowTbl := result.RawGetInt(1).(*lua.LTable)

	data := L.GetField(rowTbl, "data")
	if s, ok := data.(lua.LString); !ok || string(s) != "binary content" {
		t.Fatalf("expected data=binary content, got %v", data)
	}
}

func TestLuaTableToMap_SkipsNonStringKeys(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	L.SetField(tbl, "name", lua.LString("alice"))
	// Add an integer key -- should be skipped by LuaTableToMap
	tbl.Append(lua.LString("indexed_value"))

	result := LuaTableToMap(L, tbl)

	if result["name"] != "alice" {
		t.Fatalf("expected name=alice, got %v", result["name"])
	}

	// Integer-keyed value should not appear (LuaTableToMap only handles string keys)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry (string key only), got %d entries: %v", len(result), result)
	}
}

func TestMapToLuaTable_EmptyMap(t *testing.T) {
	L := newTestState()
	defer L.Close()

	result := MapToLuaTable(L, map[string]any{})
	if result == nil {
		t.Fatal("MapToLuaTable should return a table, not nil")
	}
	if result.Len() != 0 {
		t.Fatalf("expected empty table, got length %d", result.Len())
	}
}

func TestMapToLuaTable_NilMap(t *testing.T) {
	L := newTestState()
	defer L.Close()

	result := MapToLuaTable(L, nil)
	if result == nil {
		t.Fatal("MapToLuaTable should return a table, not nil")
	}
}

func TestLuaValueToGo_MixedTable(t *testing.T) {
	// A Lua table with both string and integer keys should be treated as a map
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	L.SetField(tbl, "name", lua.LString("alice"))
	tbl.Append(lua.LString("indexed"))

	result := LuaValueToGo(tbl)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("mixed table should be map, got %T", result)
	}

	if m["name"] != "alice" {
		t.Fatalf("expected name=alice, got %v", m["name"])
	}
}

func TestLuaValueToGo_SingleElementSequence(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.Append(lua.LString("only"))

	result := LuaValueToGo(tbl)
	sl, ok := result.([]any)
	if !ok {
		t.Fatalf("single-element sequence should be []any, got %T", result)
	}
	if len(sl) != 1 || sl[0] != "only" {
		t.Fatalf("expected [only], got %v", sl)
	}
}
