package plugin

import (
	lua "github.com/yuin/gopher-lua"
)

// RegisterTUIModule registers the "tui" Lua module with pure helper
// constructors for building UI primitives. These are convenience functions
// that produce the exact same tables as hand-built Lua tables — no side
// effects, no I/O, no defaulting logic.
//
// Usage in Lua:
//
//	local tui = require("tui")
//	coroutine.yield(tui.grid({
//	    tui.column(3, { tui.cell("List", 1.0, tui.list(items, cursor)) }),
//	    tui.column(9, { tui.cell("Detail", 0.6, tui.detail(fields)) }),
//	}))
func RegisterTUIModule(L *lua.LState) {
	mod := L.NewTable()

	L.SetField(mod, "grid", L.NewFunction(tuiGrid))
	L.SetField(mod, "column", L.NewFunction(tuiColumn))
	L.SetField(mod, "cell", L.NewFunction(tuiCell))
	L.SetField(mod, "list", L.NewFunction(tuiList))
	L.SetField(mod, "detail", L.NewFunction(tuiDetail))
	L.SetField(mod, "text", L.NewFunction(tuiText))
	L.SetField(mod, "table", L.NewFunction(tuiTable))
	L.SetField(mod, "input", L.NewFunction(tuiInput))
	L.SetField(mod, "select_field", L.NewFunction(tuiSelect))
	L.SetField(mod, "tree", L.NewFunction(tuiTree))
	L.SetField(mod, "progress", L.NewFunction(tuiProgress))

	// Set global and freeze the module to prevent modification.
	L.SetGlobal("tui", mod)
	FreezeModule(L, "tui")
}

// tui.grid(columns, hints?) → { type = "grid", columns = columns, hints = hints }
func tuiGrid(L *lua.LState) int {
	columns := L.CheckTable(1)
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString("grid"))
	tbl.RawSetString("columns", columns)
	if L.GetTop() >= 2 {
		hints := L.CheckTable(2)
		tbl.RawSetString("hints", hints)
	}
	L.Push(tbl)
	return 1
}

// tui.column(span, cells) → { span = span, cells = cells }
func tuiColumn(L *lua.LState) int {
	span := L.CheckInt(1)
	cells := L.CheckTable(2)
	tbl := L.NewTable()
	tbl.RawSetString("span", lua.LNumber(span))
	tbl.RawSetString("cells", cells)
	L.Push(tbl)
	return 1
}

// tui.cell(title, height, content) → { title = title, height = height, content = content }
func tuiCell(L *lua.LState) int {
	title := L.CheckString(1)
	height := L.CheckNumber(2)
	content := L.CheckTable(3)
	tbl := L.NewTable()
	tbl.RawSetString("title", lua.LString(title))
	tbl.RawSetString("height", height)
	tbl.RawSetString("content", content)
	L.Push(tbl)
	return 1
}

// tui.list(items, cursor) → { type = "list", items = items, cursor = cursor }
func tuiList(L *lua.LState) int {
	items := L.CheckTable(1)
	cursor := L.OptInt(2, 0)
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString("list"))
	tbl.RawSetString("items", items)
	tbl.RawSetString("cursor", lua.LNumber(cursor))
	L.Push(tbl)
	return 1
}

// tui.detail(fields) → { type = "detail", fields = fields }
func tuiDetail(L *lua.LState) int {
	fields := L.CheckTable(1)
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString("detail"))
	tbl.RawSetString("fields", fields)
	L.Push(tbl)
	return 1
}

// tui.text(lines) → { type = "text", lines = lines }
func tuiText(L *lua.LState) int {
	lines := L.CheckTable(1)
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString("text"))
	tbl.RawSetString("lines", lines)
	L.Push(tbl)
	return 1
}

// tui.table(headers, rows, cursor) → { type = "table", headers = headers, rows = rows, cursor = cursor }
func tuiTable(L *lua.LState) int {
	headers := L.CheckTable(1)
	rows := L.CheckTable(2)
	cursor := L.OptInt(3, 0)
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString("table"))
	tbl.RawSetString("headers", headers)
	tbl.RawSetString("rows", rows)
	tbl.RawSetString("cursor", lua.LNumber(cursor))
	L.Push(tbl)
	return 1
}

// tui.input(id, value, placeholder) → { type = "input", id = id, value = value, placeholder = placeholder }
func tuiInput(L *lua.LState) int {
	id := L.CheckString(1)
	value := L.OptString(2, "")
	placeholder := L.OptString(3, "")
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString("input"))
	tbl.RawSetString("id", lua.LString(id))
	tbl.RawSetString("value", lua.LString(value))
	tbl.RawSetString("placeholder", lua.LString(placeholder))
	L.Push(tbl)
	return 1
}

// tui.select_field(id, options, selected) → { type = "select", id = id, options = options, selected = selected }
// Named select_field because "select" is a Lua reserved word.
func tuiSelect(L *lua.LState) int {
	id := L.CheckString(1)
	options := L.CheckTable(2)
	selected := L.OptInt(3, 0)
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString("select"))
	tbl.RawSetString("id", lua.LString(id))
	tbl.RawSetString("options", options)
	tbl.RawSetString("selected", lua.LNumber(selected))
	L.Push(tbl)
	return 1
}

// tui.tree(nodes, cursor) → { type = "tree", nodes = nodes, cursor = cursor }
func tuiTree(L *lua.LState) int {
	nodes := L.CheckTable(1)
	cursor := L.OptInt(2, 0)
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString("tree"))
	tbl.RawSetString("nodes", nodes)
	tbl.RawSetString("cursor", lua.LNumber(cursor))
	L.Push(tbl)
	return 1
}

// tui.progress(value, label) → { type = "progress", value = value, label = label }
func tuiProgress(L *lua.LState) int {
	value := L.CheckNumber(1)
	label := L.OptString(2, "")
	tbl := L.NewTable()
	tbl.RawSetString("type", lua.LString("progress"))
	tbl.RawSetString("value", value)
	tbl.RawSetString("label", lua.LString(label))
	L.Push(tbl)
	return 1
}
