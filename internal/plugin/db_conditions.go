package plugin

import (
	"fmt"
	"strings"

	db "github.com/hegner123/modulacms/internal/db"
	lua "github.com/yuin/gopher-lua"
)

// parseWhereExtended replaces parseWhereFromLua with condition-aware detection.
// Returns exactly one of (whereMap, condition) as non-nil, or both nil for no filter.
//
// Detection rules:
//  1. If "where" field is nil/absent → (nil, nil, nil)
//  2. If table with only string keys → old path: map[string]any
//  3. If pure sequence (integer keys 1..N) → new path: db.Condition
//  4. If mixed (both string and integer keys) → error
func parseWhereExtended(L *lua.LState, optsTbl *lua.LTable) (map[string]any, db.Condition, error) {
	whereVal := L.GetField(optsTbl, "where")
	if whereVal == lua.LNil {
		return nil, nil, nil
	}
	whereTbl, ok := whereVal.(*lua.LTable)
	if !ok {
		return nil, nil, nil
	}

	// Detect table shape: string keys, integer keys, or mixed.
	hasStringKey := false
	maxIntKey := 0
	intKeyCount := 0

	whereTbl.ForEach(func(key, _ lua.LValue) {
		switch k := key.(type) {
		case lua.LNumber:
			intVal := int(k)
			if float64(intVal) == float64(k) && intVal > 0 {
				intKeyCount++
				if intVal > maxIntKey {
					maxIntKey = intVal
				}
			} else {
				hasStringKey = true
			}
		default:
			hasStringKey = true
		}
	})

	// Empty table → no filter.
	if !hasStringKey && intKeyCount == 0 {
		return nil, nil, nil
	}

	// Mixed table → error with helpful message.
	if hasStringKey && intKeyCount > 0 {
		return nil, nil, fmt.Errorf(
			"where table cannot mix string keys (equality format) and sequence entries " +
				"(condition format); use either {status = 'active'} or {{'status', '=', 'active'}}",
		)
	}

	// Old path: string keys only → map-based where.
	if hasStringKey {
		m := LuaTableToMap(L, whereTbl)
		resolved, err := resolveConditions(m)
		if err != nil {
			return nil, nil, err
		}
		return resolved, nil, nil
	}

	// New path: pure sequence → condition-based where.
	if intKeyCount > 0 && maxIntKey == intKeyCount {
		cond, err := parseConditionFromLua(L, whereTbl)
		if err != nil {
			return nil, nil, fmt.Errorf("where condition: %w", err)
		}
		return nil, cond, nil
	}

	return nil, nil, fmt.Errorf("where table has non-contiguous integer keys")
}

// parseConditionFromLua recursively converts a Lua table into a db.Condition.
//
// Supported formats:
//
//	{"col", "=", value}                     → Compare
//	{"col", "IS NULL"}                      → IsNullCondition
//	{"col", "IS NOT NULL"}                  → IsNotNullCondition
//	{"col", "IN", {v1, v2, ...}}            → InCondition
//	{"col", "BETWEEN", {low, high}}         → BetweenCondition
//	{"AND", {{cond1}, {cond2}, ...}}        → And
//	{"OR", {{cond1}, {cond2}, ...}}         → Or
//	{"NOT", {cond}}                         → Not
//	{{cond1}, {cond2}, ...}                 → implicit And
func parseConditionFromLua(L *lua.LState, tbl *lua.LTable) (db.Condition, error) {
	first := tbl.RawGetInt(1)
	if first == lua.LNil {
		return nil, fmt.Errorf("empty condition table")
	}

	// If first element is a string, it's either a keyword (AND/OR/NOT) or a column name.
	if s, ok := first.(lua.LString); ok {
		keyword := strings.ToUpper(string(s))

		switch keyword {
		case "AND", "OR":
			return parseLogicalFromLua(L, tbl, keyword)
		case "NOT":
			return parseNotFromLua(L, tbl)
		default:
			// Column name — parse as leaf condition.
			return parseLeafFromLua(L, tbl, string(s))
		}
	}

	// If first element is a table, it's an implicit AND of conditions.
	if _, ok := first.(*lua.LTable); ok {
		return parseImplicitAndFromLua(L, tbl)
	}

	return nil, fmt.Errorf("condition element 1 must be a string or table, got %s", first.Type())
}

// parseLogicalFromLua parses {"AND"|"OR", {{cond1}, {cond2}, ...}}.
func parseLogicalFromLua(L *lua.LState, tbl *lua.LTable, keyword string) (db.Condition, error) {
	childrenVal := tbl.RawGetInt(2)
	childrenTbl, ok := childrenVal.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("%s requires a table of child conditions as element 2", keyword)
	}

	children, err := parseConditionSequence(L, childrenTbl)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", keyword, err)
	}

	if keyword == "AND" {
		return db.And{Conditions: children}, nil
	}
	return db.Or{Conditions: children}, nil
}

// parseNotFromLua parses {"NOT", {condition}}.
func parseNotFromLua(L *lua.LState, tbl *lua.LTable) (db.Condition, error) {
	childVal := tbl.RawGetInt(2)
	childTbl, ok := childVal.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("NOT requires a condition table as element 2")
	}

	child, err := parseConditionFromLua(L, childTbl)
	if err != nil {
		return nil, fmt.Errorf("NOT: %w", err)
	}
	return db.Not{Condition: child}, nil
}

// parseLeafFromLua parses a leaf condition with a column name at element 1.
// Formats: {"col", "op", value}, {"col", "IS NULL"}, {"col", "IS NOT NULL"},
// {"col", "IN", {vals}}, {"col", "BETWEEN", {low, high}}.
func parseLeafFromLua(L *lua.LState, tbl *lua.LTable, column string) (db.Condition, error) {
	second := tbl.RawGetInt(2)
	opStr, ok := second.(lua.LString)
	if !ok {
		return nil, fmt.Errorf("condition element 2 must be an operator string, got %s", second.Type())
	}

	op := strings.ToUpper(string(opStr))

	switch op {
	case "IS NULL":
		return db.IsNullCondition{Column: column}, nil
	case "IS NOT NULL":
		return db.IsNotNullCondition{Column: column}, nil
	case "IN":
		return parseInFromLua(L, tbl, column)
	case "BETWEEN":
		return parseBetweenFromLua(L, tbl, column)
	default:
		return parseCompareFromLua(L, tbl, column, op)
	}
}

// parseCompareFromLua parses {"col", "op", value} into a Compare condition.
func parseCompareFromLua(L *lua.LState, tbl *lua.LTable, column, op string) (db.Condition, error) {
	// Normalize != to <>
	if op == "!=" {
		op = "<>"
	}

	compareOp := db.CompareOp(op)
	if _, valid := map[db.CompareOp]bool{
		db.OpEq: true, db.OpNeq: true, db.OpGt: true, db.OpLt: true,
		db.OpGte: true, db.OpLte: true, db.OpLike: true,
	}[compareOp]; !valid {
		return nil, fmt.Errorf("unsupported operator %q", op)
	}

	third := tbl.RawGetInt(3)
	if third == lua.LNil {
		return nil, fmt.Errorf("compare with %q requires a non-nil value (element 3)", op)
	}

	value := LuaValueToGo(third)
	return db.Compare{Column: column, Op: compareOp, Value: value}, nil
}

// parseInFromLua parses {"col", "IN", {v1, v2, ...}} into an InCondition.
func parseInFromLua(L *lua.LState, tbl *lua.LTable, column string) (db.Condition, error) {
	third := tbl.RawGetInt(3)
	valsTbl, ok := third.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("IN requires a table of values as element 3")
	}

	var vals []any
	valsTbl.ForEach(func(key, value lua.LValue) {
		if _, ok := key.(lua.LNumber); ok {
			vals = append(vals, LuaValueToGo(value))
		}
	})

	if len(vals) == 0 {
		return nil, fmt.Errorf("IN requires at least one value")
	}

	return db.InCondition{Column: column, Values: vals}, nil
}

// parseBetweenFromLua parses {"col", "BETWEEN", {low, high}} into a BetweenCondition.
func parseBetweenFromLua(L *lua.LState, tbl *lua.LTable, column string) (db.Condition, error) {
	third := tbl.RawGetInt(3)
	valsTbl, ok := third.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("BETWEEN requires a table of 2 values as element 3")
	}

	low := LuaValueToGo(valsTbl.RawGetInt(1))
	high := LuaValueToGo(valsTbl.RawGetInt(2))

	if low == nil || high == nil {
		return nil, fmt.Errorf("BETWEEN requires exactly 2 non-nil values")
	}

	return db.BetweenCondition{Column: column, Low: low, High: high}, nil
}

// parseImplicitAndFromLua parses a sequence of condition tables as implicit AND.
func parseImplicitAndFromLua(L *lua.LState, tbl *lua.LTable) (db.Condition, error) {
	children, err := parseConditionSequence(L, tbl)
	if err != nil {
		return nil, fmt.Errorf("implicit AND: %w", err)
	}
	return db.And{Conditions: children}, nil
}

// parseConditionSequence parses a Lua sequence table where each element is a condition table.
func parseConditionSequence(L *lua.LState, tbl *lua.LTable) ([]db.Condition, error) {
	var conditions []db.Condition
	var parseErr error

	tbl.ForEach(func(key, value lua.LValue) {
		if parseErr != nil {
			return
		}
		if _, ok := key.(lua.LNumber); !ok {
			return
		}
		childTbl, ok := value.(*lua.LTable)
		if !ok {
			parseErr = fmt.Errorf("child condition must be a table, got %s", value.Type())
			return
		}
		cond, err := parseConditionFromLua(L, childTbl)
		if err != nil {
			parseErr = err
			return
		}
		conditions = append(conditions, cond)
	})

	if parseErr != nil {
		return nil, parseErr
	}
	if len(conditions) == 0 {
		return nil, fmt.Errorf("condition sequence cannot be empty")
	}
	return conditions, nil
}
