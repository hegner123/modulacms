package plugin

import (
	"strings"
	"testing"

	db "github.com/hegner123/modulacms/internal/db"
	lua "github.com/yuin/gopher-lua"
)

// newTestLState creates a minimal Lua state for testing.
func newTestLState() *lua.LState {
	return lua.NewState(lua.Options{SkipOpenLibs: true})
}

// luaDoString is a test helper that executes Lua code and returns the result table from the stack.
func luaExecGetTable(t *testing.T, L *lua.LState, code string) *lua.LTable {
	t.Helper()
	if err := L.DoString(code); err != nil {
		t.Fatalf("lua exec: %v", err)
	}
	val := L.Get(-1)
	L.Pop(1)
	tbl, ok := val.(*lua.LTable)
	if !ok {
		t.Fatalf("expected table, got %s", val.Type())
	}
	return tbl
}

// ===== parseConditionFromLua =====

func TestParseCondition_SimpleCompare(t *testing.T) {
	tests := []struct {
		name   string
		lua    string
		wantOp db.CompareOp
		wantV  any
	}{
		{"eq", `return {"status", "=", "active"}`, db.OpEq, "active"},
		{"neq <>", `return {"status", "<>", "deleted"}`, db.OpNeq, "deleted"},
		{"neq != normalized", `return {"status", "!=", "deleted"}`, db.OpNeq, "deleted"},
		{"gt", `return {"priority", ">", 5}`, db.OpGt, float64(5)},
		{"lt", `return {"priority", "<", 10}`, db.OpLt, float64(10)},
		{"gte", `return {"priority", ">=", 3}`, db.OpGte, float64(3)},
		{"lte", `return {"priority", "<=", 7}`, db.OpLte, float64(7)},
		{"like", `return {"name", "LIKE", "%test%"}`, db.OpLike, "%test%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := newTestLState()
			defer L.Close()

			tbl := luaExecGetTable(t, L, tt.lua)
			cond, err := parseConditionFromLua(L, tbl)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			c, ok := cond.(db.Compare)
			if !ok {
				t.Fatalf("expected Compare, got %T", cond)
			}
			if c.Column != "status" && c.Column != "priority" && c.Column != "name" {
				t.Errorf("unexpected column %q", c.Column)
			}
			if c.Op != tt.wantOp {
				t.Errorf("op = %q, want %q", c.Op, tt.wantOp)
			}
			if c.Value != tt.wantV {
				t.Errorf("value = %v (%T), want %v (%T)", c.Value, c.Value, tt.wantV, tt.wantV)
			}
		})
	}
}

func TestParseCondition_IsNull(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"description", "IS NULL"}`)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := cond.(db.IsNullCondition)
	if !ok {
		t.Fatalf("expected IsNullCondition, got %T", cond)
	}
	if c.Column != "description" {
		t.Errorf("column = %q, want description", c.Column)
	}
}

func TestParseCondition_IsNotNull(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"description", "IS NOT NULL"}`)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := cond.(db.IsNotNullCondition)
	if !ok {
		t.Fatalf("expected IsNotNullCondition, got %T", cond)
	}
	if c.Column != "description" {
		t.Errorf("column = %q, want description", c.Column)
	}
}

func TestParseCondition_In(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"status", "IN", {"active", "draft", "review"}}`)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := cond.(db.InCondition)
	if !ok {
		t.Fatalf("expected InCondition, got %T", cond)
	}
	if c.Column != "status" {
		t.Errorf("column = %q", c.Column)
	}
	if len(c.Values) != 3 {
		t.Fatalf("values len = %d, want 3", len(c.Values))
	}
	if c.Values[0] != "active" || c.Values[1] != "draft" || c.Values[2] != "review" {
		t.Errorf("values = %v", c.Values)
	}
}

func TestParseCondition_Between(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"priority", "BETWEEN", {1, 10}}`)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := cond.(db.BetweenCondition)
	if !ok {
		t.Fatalf("expected BetweenCondition, got %T", cond)
	}
	if c.Column != "priority" {
		t.Errorf("column = %q", c.Column)
	}
	if c.Low != float64(1) || c.High != float64(10) {
		t.Errorf("low = %v, high = %v", c.Low, c.High)
	}
}

func TestParseCondition_ExplicitAnd(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"AND", {{"status", "=", "active"}, {"priority", ">", 5}}}`)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a, ok := cond.(db.And)
	if !ok {
		t.Fatalf("expected And, got %T", cond)
	}
	if len(a.Conditions) != 2 {
		t.Fatalf("children len = %d, want 2", len(a.Conditions))
	}
}

func TestParseCondition_ExplicitOr(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"OR", {{"status", "=", "active"}, {"status", "=", "draft"}}}`)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	o, ok := cond.(db.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", cond)
	}
	if len(o.Conditions) != 2 {
		t.Fatalf("children len = %d, want 2", len(o.Conditions))
	}
}

func TestParseCondition_Not(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"NOT", {"status", "=", "deleted"}}`)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n, ok := cond.(db.Not)
	if !ok {
		t.Fatalf("expected Not, got %T", cond)
	}
	_, ok = n.Condition.(db.Compare)
	if !ok {
		t.Fatalf("expected Not wrapping Compare, got %T", n.Condition)
	}
}

func TestParseCondition_ImplicitAnd(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {{"status", "=", "active"}, {"priority", ">", 3}}`)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a, ok := cond.(db.And)
	if !ok {
		t.Fatalf("expected And (implicit), got %T", cond)
	}
	if len(a.Conditions) != 2 {
		t.Fatalf("children len = %d, want 2", len(a.Conditions))
	}
}

func TestParseCondition_NestedOrOfAnds(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	code := `return {"OR", {
		{"AND", {{"status", "=", "active"}, {"priority", ">", 5}}},
		{"AND", {{"status", "=", "urgent"}, {"priority", ">", 0}}},
	}}`
	tbl := luaExecGetTable(t, L, code)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	o, ok := cond.(db.Or)
	if !ok {
		t.Fatalf("expected Or, got %T", cond)
	}
	if len(o.Conditions) != 2 {
		t.Fatalf("OR children len = %d, want 2", len(o.Conditions))
	}
	for i, child := range o.Conditions {
		a, ok := child.(db.And)
		if !ok {
			t.Fatalf("OR child %d: expected And, got %T", i, child)
		}
		if len(a.Conditions) != 2 {
			t.Fatalf("AND child %d has %d conditions", i, len(a.Conditions))
		}
	}
}

func TestParseCondition_InvalidOperator(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"status", "INVALID_OP", "x"}`)
	_, err := parseConditionFromLua(L, tbl)
	if err == nil {
		t.Fatal("expected error for invalid operator")
	}
	if !strings.Contains(err.Error(), "unsupported operator") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestParseCondition_EmptyTable(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {}`)
	_, err := parseConditionFromLua(L, tbl)
	if err == nil {
		t.Fatal("expected error for empty table")
	}
}

func TestParseCondition_NilCompareValue(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"status", "="}`)
	_, err := parseConditionFromLua(L, tbl)
	if err == nil {
		t.Fatal("expected error for missing compare value")
	}
	if !strings.Contains(err.Error(), "non-nil") {
		t.Errorf("error = %q", err.Error())
	}
}

// ===== parseWhereExtended =====

func TestParseWhereExtended_OldStyleMap(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	// Create opts table with string-key where: {status = "active"}
	if err := L.DoString(`_opts = {where = {status = "active"}}`); err != nil {
		t.Fatalf("lua: %v", err)
	}
	optsTbl := L.GetGlobal("_opts").(*lua.LTable)

	whereMap, cond, err := parseWhereExtended(L, optsTbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cond != nil {
		t.Error("expected nil condition for old-style where")
	}
	if whereMap == nil {
		t.Fatal("expected non-nil whereMap")
	}
	if whereMap["status"] != "active" {
		t.Errorf("status = %v", whereMap["status"])
	}
}

func TestParseWhereExtended_NewStyleCondition(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	// Create opts table with sequence-style where: {{"status", "=", "active"}}
	if err := L.DoString(`_opts = {where = {{"status", "=", "active"}}}`); err != nil {
		t.Fatalf("lua: %v", err)
	}
	optsTbl := L.GetGlobal("_opts").(*lua.LTable)

	whereMap, cond, err := parseWhereExtended(L, optsTbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if whereMap != nil {
		t.Error("expected nil whereMap for new-style where")
	}
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}
	// Should be implicit AND wrapping a single Compare
	a, ok := cond.(db.And)
	if !ok {
		t.Fatalf("expected And, got %T", cond)
	}
	if len(a.Conditions) != 1 {
		t.Fatalf("children len = %d, want 1", len(a.Conditions))
	}
}

func TestParseWhereExtended_MixedTableError(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	// Mixed table: both string keys and integer keys
	if err := L.DoString(`_opts = {where = {status = "active", {"priority", ">", 5}}}`); err != nil {
		t.Fatalf("lua: %v", err)
	}
	optsTbl := L.GetGlobal("_opts").(*lua.LTable)

	_, _, err := parseWhereExtended(L, optsTbl)
	if err == nil {
		t.Fatal("expected error for mixed table")
	}
	if !strings.Contains(err.Error(), "cannot mix") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestParseWhereExtended_NilWhere(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	if err := L.DoString(`_opts = {limit = 10}`); err != nil {
		t.Fatalf("lua: %v", err)
	}
	optsTbl := L.GetGlobal("_opts").(*lua.LTable)

	whereMap, cond, err := parseWhereExtended(L, optsTbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if whereMap != nil || cond != nil {
		t.Error("expected both nil for absent where")
	}
}

func TestParseWhereExtended_EmptyTable(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	if err := L.DoString(`_opts = {where = {}}`); err != nil {
		t.Fatalf("lua: %v", err)
	}
	optsTbl := L.GetGlobal("_opts").(*lua.LTable)

	whereMap, cond, err := parseWhereExtended(L, optsTbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if whereMap != nil || cond != nil {
		t.Error("expected both nil for empty where table")
	}
}

func TestParseWhereExtended_CompoundNewStyle(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	code := `_opts = {where = {
		{"status", "=", "active"},
		{"priority", ">", 3},
	}}`
	if err := L.DoString(code); err != nil {
		t.Fatalf("lua: %v", err)
	}
	optsTbl := L.GetGlobal("_opts").(*lua.LTable)

	whereMap, cond, err := parseWhereExtended(L, optsTbl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if whereMap != nil {
		t.Error("expected nil whereMap")
	}
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}

	a, ok := cond.(db.And)
	if !ok {
		t.Fatalf("expected And, got %T", cond)
	}
	if len(a.Conditions) != 2 {
		t.Fatalf("children len = %d, want 2", len(a.Conditions))
	}
}

// ===== SQL generation from parsed Lua conditions =====

func TestParseCondition_BuildsCorrectSQL(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	code := `return {"AND", {
		{"status", "=", "active"},
		{"priority", "IN", {1, 2, 3}},
		{"name", "LIKE", "%test%"},
	}}`
	tbl := luaExecGetTable(t, L, code)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	ctx := db.NewBuildContext()
	sql, args, _, err := cond.Build(ctx, db.DialectPostgres, 1)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	want := `("status" = $1 AND "priority" IN ($2, $3, $4) AND "name" LIKE $5)`
	if sql != want {
		t.Errorf("sql = %q, want %q", sql, want)
	}
	if len(args) != 5 {
		t.Fatalf("args len = %d, want 5", len(args))
	}
}

func TestParseCondition_NeqNormalizedInSQL(t *testing.T) {
	L := newTestLState()
	defer L.Close()

	tbl := luaExecGetTable(t, L, `return {"status", "!=", "deleted"}`)
	cond, err := parseConditionFromLua(L, tbl)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	ctx := db.NewBuildContext()
	sql, _, _, err := cond.Build(ctx, db.DialectSQLite, 1)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	// != should be normalized to <> in the SQL output
	if !strings.Contains(sql, "<>") {
		t.Errorf("expected <> in SQL, got %q", sql)
	}
	if strings.Contains(sql, "!=") {
		t.Errorf("SQL should not contain !=, got %q", sql)
	}
}
