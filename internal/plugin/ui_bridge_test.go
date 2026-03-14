package plugin

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// newTestVM creates a sandboxed Lua VM with coroutines enabled for testing.
func newTestVM() *lua.LState {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	ApplySandbox(L, SandboxConfig{AllowCoroutine: true})
	return L
}

func TestCoroutineBridge_StartAndResume(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	// A screen function that yields a list primitive, then quits.
	err := L.DoString(`
		function test_screen(ctx)
			local event = coroutine.yield({
				type = "list",
				items = {
					{ label = "Item 1", id = "1" },
					{ label = "Item 2", id = "2" },
				},
				cursor = 0,
			})
			if event.type == "key" and event.key == "q" then
				return
			end
			coroutine.yield({
				type = "text",
				lines = { "still running" },
			})
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	// Start with init event.
	initEvent := BuildInitEvent(L, 80, 24, nil)
	yv, err := bridge.Start(initEvent)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if yv.IsAction {
		t.Fatal("expected layout yield, got action")
	}
	if yv.Primitive == nil {
		t.Fatal("expected primitive yield")
	}
	list, ok := yv.Primitive.(*ListPrimitive)
	if !ok {
		t.Fatalf("expected ListPrimitive, got %T", yv.Primitive)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list.Items))
	}
	if list.Items[0].Label != "Item 1" {
		t.Errorf("expected 'Item 1', got %q", list.Items[0].Label)
	}

	// Resume with quit key.
	quitEvent := BuildKeyEvent(L, "q")
	yv, err = bridge.Resume(quitEvent)
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if !yv.IsAction || yv.Action.Name != "quit" {
		t.Fatalf("expected quit action, got %+v", yv)
	}
	if !bridge.Done() {
		t.Fatal("expected bridge to be done")
	}
}

func TestCoroutineBridge_ActionYield(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_screen(ctx)
			coroutine.yield({ action = "toast", message = "Hello!", level = "success" })
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	initEvent := BuildInitEvent(L, 80, 24, nil)
	yv, err := bridge.Start(initEvent)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !yv.IsAction {
		t.Fatal("expected action yield")
	}
	if yv.Action.Name != "toast" {
		t.Fatalf("expected toast action, got %q", yv.Action.Name)
	}
	if yv.Action.Params["message"] != "Hello!" {
		t.Errorf("expected message 'Hello!', got %v", yv.Action.Params["message"])
	}
}

func TestCoroutineBridge_GridLayout(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_screen(ctx)
			coroutine.yield({
				type = "grid",
				columns = {
					{ span = 3, cells = {
						{ title = "List", height = 1.0, content = {
							type = "list", items = {{ label = "A", id = "a" }}, cursor = 0,
						}},
					}},
					{ span = 9, cells = {
						{ title = "Detail", height = 0.6, content = {
							type = "detail", fields = {{ label = "Name", value = "Test" }},
						}},
						{ title = "Info", height = 0.4, content = {
							type = "text", lines = { "Info text" },
						}},
					}},
				},
				hints = {
					{ key = "n", label = "new" },
					{ key = "q", label = "quit" },
				},
			})
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	initEvent := BuildInitEvent(L, 120, 40, nil)
	yv, err := bridge.Start(initEvent)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if yv.Layout == nil {
		t.Fatal("expected layout yield")
	}
	if len(yv.Layout.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(yv.Layout.Columns))
	}
	if yv.Layout.Columns[0].Span != 3 {
		t.Errorf("expected span 3, got %d", yv.Layout.Columns[0].Span)
	}
	if yv.Layout.Columns[1].Span != 9 {
		t.Errorf("expected span 9, got %d", yv.Layout.Columns[1].Span)
	}
	if len(yv.Layout.Columns[1].Cells) != 2 {
		t.Fatalf("expected 2 cells in column 2, got %d", len(yv.Layout.Columns[1].Cells))
	}
	if len(yv.Layout.Hints) != 2 {
		t.Fatalf("expected 2 hints, got %d", len(yv.Layout.Hints))
	}
	if yv.Layout.Hints[0].Key != "n" {
		t.Errorf("expected hint key 'n', got %q", yv.Layout.Hints[0].Key)
	}
}

func TestCoroutineBridge_CommitAction(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_interface(ctx)
			coroutine.yield({ action = "commit", value = "#ff0000" })
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_interface").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	initEvent := BuildFieldInitEvent(L, 60, 3, "#000000", nil)
	yv, err := bridge.Start(initEvent)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !yv.IsAction {
		t.Fatal("expected action yield")
	}
	if yv.Action.Name != "commit" {
		t.Fatalf("expected commit action, got %q", yv.Action.Name)
	}
	if yv.Action.Params["value"] != "#ff0000" {
		t.Errorf("expected value '#ff0000', got %v", yv.Action.Params["value"])
	}
}

func TestCoroutineBridge_FetchAction(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_screen(ctx)
			coroutine.yield({
				action = "fetch",
				id = "load_items",
				query = "tasks",
				params = { where = { status = "active" } },
			})
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	initEvent := BuildInitEvent(L, 80, 24, nil)
	yv, err := bridge.Start(initEvent)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !yv.IsAction || yv.Action.Name != "fetch" {
		t.Fatalf("expected fetch action, got %+v", yv)
	}
	if yv.Action.Params["id"] != "load_items" {
		t.Errorf("expected id 'load_items', got %v", yv.Action.Params["id"])
	}
}

func TestCoroutineBridge_DoubleStart(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_screen(ctx)
			coroutine.yield({ type = "text", lines = { "hello" } })
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	initEvent := BuildInitEvent(L, 80, 24, nil)
	_, err = bridge.Start(initEvent)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	_, err = bridge.Start(initEvent)
	if err == nil {
		t.Fatal("expected error on double start")
	}
}

func TestCoroutineBridge_ResumeBeforeStart(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_screen(ctx)
			coroutine.yield({ type = "text", lines = { "hello" } })
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	event := BuildKeyEvent(L, "j")
	_, err = bridge.Resume(event)
	if err == nil {
		t.Fatal("expected error on resume before start")
	}
}

func TestCoroutineBridge_UnknownAction(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_screen(ctx)
			coroutine.yield({ action = "explode" })
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	initEvent := BuildInitEvent(L, 80, 24, nil)
	_, err = bridge.Start(initEvent)
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestCoroutineBridge_UnknownPrimitiveType(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_screen(ctx)
			coroutine.yield({ type = "spaceship" })
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	initEvent := BuildInitEvent(L, 80, 24, nil)
	_, err = bridge.Start(initEvent)
	if err == nil {
		t.Fatal("expected error for unknown primitive type")
	}
}

func TestParsePrimitive_AllTypes(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	tests := []struct {
		name     string
		lua      string
		wantType string
	}{
		{"list", `{ type = "list", items = {{ label = "A", id = "1" }}, cursor = 0 }`, "list"},
		{"detail", `{ type = "detail", title = "D", fields = {{ label = "K", value = "V" }} }`, "detail"},
		{"text", `{ type = "text", lines = { "hello", { text = "bold", bold = true } } }`, "text"},
		{"table", `{ type = "table", headers = {"A","B"}, rows = {{"1","2"}}, cursor = 0 }`, "table"},
		{"input", `{ type = "input", id = "search", value = "", placeholder = "Search..." }`, "input"},
		{"select", `{ type = "select", id = "s", options = {{ label = "A", value = "a" }}, selected = 0 }`, "select"},
		{"tree", `{ type = "tree", nodes = {{ label = "R", id = "r", children = {{ label = "C", id = "c" }} }}, cursor = 0 }`, "tree"},
		{"progress", `{ type = "progress", value = 0.75, label = "Loading" }`, "progress"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := L.DoString("_test_val = " + tt.lua)
			if err != nil {
				t.Fatalf("DoString: %v", err)
			}
			tbl := L.GetGlobal("_test_val").(*lua.LTable)
			prim, err := ParsePrimitive(tbl)
			if err != nil {
				t.Fatalf("ParsePrimitive: %v", err)
			}
			if prim.PrimitiveType() != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, prim.PrimitiveType())
			}
		})
	}
}

func TestBuildEvent_Types(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	// Key event.
	keyEvt := BuildKeyEvent(L, "enter")
	if s := keyEvt.RawGetString("type").(lua.LString); string(s) != "key" {
		t.Errorf("expected type 'key', got %q", string(s))
	}
	if s := keyEvt.RawGetString("key").(lua.LString); string(s) != "enter" {
		t.Errorf("expected key 'enter', got %q", string(s))
	}

	// Resize event.
	resizeEvt := BuildResizeEvent(L, 120, 40)
	if s := resizeEvt.RawGetString("type").(lua.LString); string(s) != "resize" {
		t.Errorf("expected type 'resize', got %q", string(s))
	}
	if n := resizeEvt.RawGetString("width").(lua.LNumber); int(n) != 120 {
		t.Errorf("expected width 120, got %d", int(n))
	}

	// Data event (success).
	dataEvt := BuildDataEvent(L, "load", true, lua.LString("result"), "")
	if s := dataEvt.RawGetString("id").(lua.LString); string(s) != "load" {
		t.Errorf("expected id 'load', got %q", string(s))
	}
	if b := dataEvt.RawGetString("ok").(lua.LBool); !bool(b) {
		t.Error("expected ok=true")
	}

	// Data event (error).
	errEvt := BuildDataEvent(L, "load", false, nil, "timeout")
	if b := errEvt.RawGetString("ok").(lua.LBool); bool(b) {
		t.Error("expected ok=false")
	}
	if s := errEvt.RawGetString("error").(lua.LString); string(s) != "timeout" {
		t.Errorf("expected error 'timeout', got %q", string(s))
	}

	// Dialog event.
	dialogEvt := BuildDialogEvent(L, true)
	if s := dialogEvt.RawGetString("type").(lua.LString); string(s) != "dialog" {
		t.Errorf("expected type 'dialog', got %q", string(s))
	}

	// Init event with params.
	initEvt := BuildInitEvent(L, 80, 24, map[string]string{"id": "123"})
	if s := initEvt.RawGetString("type").(lua.LString); string(s) != "init" {
		t.Errorf("expected type 'init', got %q", string(s))
	}
	paramsTbl := initEvt.RawGetString("params").(*lua.LTable)
	if s := paramsTbl.RawGetString("id").(lua.LString); string(s) != "123" {
		t.Errorf("expected param id '123', got %q", string(s))
	}

	// Field init event.
	fieldEvt := BuildFieldInitEvent(L, 60, 3, "#ff0000", nil)
	if s := fieldEvt.RawGetString("value").(lua.LString); string(s) != "#ff0000" {
		t.Errorf("expected value '#ff0000', got %q", string(s))
	}
}

func TestCoroutineBridge_ImmediateReturn(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_screen(ctx)
			-- returns immediately without yielding
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	initEvent := BuildInitEvent(L, 80, 24, nil)
	yv, err := bridge.Start(initEvent)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !yv.IsAction || yv.Action.Name != "quit" {
		t.Fatalf("expected quit action for immediate return, got %+v", yv)
	}
	if !bridge.Done() {
		t.Fatal("expected bridge to be done")
	}
}

func TestCoroutineBridge_NavigateAction(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	err := L.DoString(`
		function test_screen(ctx)
			coroutine.yield({
				action = "navigate",
				plugin = "other_plugin",
				screen = "detail",
				params = { id = "abc" },
			})
		end
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	fn := L.GetGlobal("test_screen").(*lua.LFunction)
	bridge := NewCoroutineBridge(nil, L, fn)

	initEvent := BuildInitEvent(L, 80, 24, nil)
	yv, err := bridge.Start(initEvent)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !yv.IsAction || yv.Action.Name != "navigate" {
		t.Fatalf("expected navigate action, got %+v", yv)
	}
	if yv.Action.Params["plugin"] != "other_plugin" {
		t.Errorf("expected plugin 'other_plugin', got %v", yv.Action.Params["plugin"])
	}
	if yv.Action.Params["screen"] != "detail" {
		t.Errorf("expected screen 'detail', got %v", yv.Action.Params["screen"])
	}
}

func TestParseLayout_SpanValidation(t *testing.T) {
	L := newTestVM()
	defer L.Close()

	// Column with zero span defaults to 1.
	err := L.DoString(`
		_test_layout = {
			type = "grid",
			columns = {
				{ span = 0, cells = {
					{ title = "A", height = 1.0, content = { type = "text", lines = {"x"} } },
				}},
			},
		}
	`)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	tbl := L.GetGlobal("_test_layout").(*lua.LTable)
	layout, err := ParseLayout(tbl)
	if err != nil {
		t.Fatalf("ParseLayout: %v", err)
	}
	if layout.Columns[0].Span != 1 {
		t.Errorf("expected span 1 (default), got %d", layout.Columns[0].Span)
	}
}
