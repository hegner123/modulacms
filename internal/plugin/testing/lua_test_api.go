package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/plugin"
	lua "github.com/yuin/gopher-lua"
)

// testAPIState tracks assertion state for a single test function execution.
type testAPIState struct {
	assertions int
	failures   []Failure
	setupFn    *lua.LFunction
	teardownFn *lua.LFunction

	// injected dependencies
	harness *Harness
	mockReq *MockRequestEngine
}

// registerTestModule creates the "test" Lua module and registers it on L as a
// frozen global. Returns the state so the harness can read assertion results.
func registerTestModule(L *lua.LState, h *Harness) *testAPIState {
	state := &testAPIState{
		harness: h,
		mockReq: h.mockReq,
	}

	mod := L.NewTable()

	L.SetField(mod, "assert", L.NewFunction(state.luaAssert))
	L.SetField(mod, "assert_eq", L.NewFunction(state.luaAssertEq))
	L.SetField(mod, "assert_neq", L.NewFunction(state.luaAssertNeq))
	L.SetField(mod, "assert_nil", L.NewFunction(state.luaAssertNil))
	L.SetField(mod, "assert_not_nil", L.NewFunction(state.luaAssertNotNil))
	L.SetField(mod, "assert_error", L.NewFunction(state.luaAssertError))
	L.SetField(mod, "assert_contains", L.NewFunction(state.luaAssertContains))
	L.SetField(mod, "request", L.NewFunction(state.luaRequest))
	L.SetField(mod, "fire_hook", L.NewFunction(state.luaFireHook))
	L.SetField(mod, "mock_request", L.NewFunction(state.luaMockRequest))
	L.SetField(mod, "setup", L.NewFunction(state.luaSetup))
	L.SetField(mod, "teardown", L.NewFunction(state.luaTeardown))

	L.SetGlobal("test", mod)
	plugin.FreezeModule(L, "test")

	return state
}

// reset clears per-test state between test functions.
func (s *testAPIState) reset() {
	s.assertions = 0
	s.failures = []Failure{} // empty slice, not nil, so JSON encodes as [] not null
}

// currentLine returns the Lua source line of the caller.
func currentLine(L *lua.LState) int {
	dbg, ok := L.GetStack(1)
	if !ok {
		return 0
	}
	L.GetInfo("l", dbg, lua.LNil)
	return dbg.CurrentLine
}

// recordFailure adds a soft failure.
func (s *testAPIState) recordFailure(line int, msg string) {
	s.failures = append(s.failures, Failure{
		Message: msg,
		Line:    line,
	})
}

// --- Assertions ---

func (s *testAPIState) luaAssert(L *lua.LState) int {
	s.assertions++
	cond := L.Get(1)
	msg := L.OptString(2, "assertion failed")

	if cond == lua.LNil || cond == lua.LFalse {
		s.recordFailure(currentLine(L), msg)
	}
	return 0
}

func (s *testAPIState) luaAssertEq(L *lua.LState) int {
	s.assertions++
	expected := L.Get(1)
	actual := L.Get(2)
	msg := L.OptString(3, "")

	if !luaDeepEqual(expected, actual) {
		detail := fmt.Sprintf("expected %s, got %s", luaFormat(expected), luaFormat(actual))
		if msg != "" {
			detail = msg + ": " + detail
		}
		s.recordFailure(currentLine(L), detail)
	}
	return 0
}

func (s *testAPIState) luaAssertNeq(L *lua.LState) int {
	s.assertions++
	a := L.Get(1)
	b := L.Get(2)
	msg := L.OptString(3, "")

	if luaDeepEqual(a, b) {
		detail := fmt.Sprintf("expected values to differ, both are %s", luaFormat(a))
		if msg != "" {
			detail = msg + ": " + detail
		}
		s.recordFailure(currentLine(L), detail)
	}
	return 0
}

func (s *testAPIState) luaAssertNil(L *lua.LState) int {
	s.assertions++
	val := L.Get(1)
	msg := L.OptString(2, "")

	if val != lua.LNil {
		detail := fmt.Sprintf("expected nil, got %s", luaFormat(val))
		if msg != "" {
			detail = msg + ": " + detail
		}
		s.recordFailure(currentLine(L), detail)
	}
	return 0
}

func (s *testAPIState) luaAssertNotNil(L *lua.LState) int {
	s.assertions++
	val := L.Get(1)
	msg := L.OptString(2, "")

	if val == lua.LNil {
		detail := "expected non-nil value, got nil"
		if msg != "" {
			detail = msg + ": " + detail
		}
		s.recordFailure(currentLine(L), detail)
	}
	return 0
}

func (s *testAPIState) luaAssertError(L *lua.LState) int {
	s.assertions++
	fn := L.CheckFunction(1)
	pattern := L.OptString(2, "")

	err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	})

	if err == nil {
		s.recordFailure(currentLine(L), "expected function to raise an error, but it did not")
		return 0
	}

	if pattern != "" {
		errMsg := err.Error()
		if !strings.Contains(errMsg, pattern) {
			s.recordFailure(currentLine(L), fmt.Sprintf("error %q does not contain %q", errMsg, pattern))
		}
	}
	return 0
}

func (s *testAPIState) luaAssertContains(L *lua.LState) int {
	s.assertions++
	haystack := L.CheckString(1)
	needle := L.CheckString(2)
	msg := L.OptString(3, "")

	if !strings.Contains(haystack, needle) {
		detail := fmt.Sprintf("string %q does not contain %q", haystack, needle)
		if msg != "" {
			detail = msg + ": " + detail
		}
		s.recordFailure(currentLine(L), detail)
	}
	return 0
}

// --- test.setup / test.teardown ---

func (s *testAPIState) luaSetup(L *lua.LState) int {
	s.setupFn = L.CheckFunction(1)
	return 0
}

func (s *testAPIState) luaTeardown(L *lua.LState) int {
	s.teardownFn = L.CheckFunction(1)
	return 0
}

// --- test.request ---

func (s *testAPIState) luaRequest(L *lua.LState) int {
	method := L.CheckString(1)
	path := L.CheckString(2)
	var optsTbl *lua.LTable
	if L.GetTop() >= 3 {
		optsTbl = L.OptTable(3, nil)
	}

	// Build auth context
	auth := "admin"
	if optsTbl != nil {
		if v := optsTbl.RawGetString("auth"); v != lua.LNil {
			auth = v.String()
		}
	}

	// Build request body
	var bodyStr string
	if optsTbl != nil {
		if bodyVal := optsTbl.RawGetString("body"); bodyVal != lua.LNil {
			bodyStr = bodyVal.String()
		}
	}

	var body *strings.Reader
	if bodyStr != "" {
		body = strings.NewReader(bodyStr)
	} else {
		body = strings.NewReader("")
	}

	req := httptest.NewRequest(method, path, body)
	if bodyStr != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers from opts
	if optsTbl != nil {
		if headersTbl, ok := optsTbl.RawGetString("headers").(*lua.LTable); ok {
			headersTbl.ForEach(func(k, v lua.LValue) {
				req.Header.Set(k.String(), v.String())
			})
		}
	}

	ctx := req.Context()
	switch auth {
	case "admin":
		ctx = s.injectAdminAuth(ctx)
	case "viewer":
		ctx = s.injectViewerAuth(ctx)
	case "none":
		// no auth context
	default:
		L.RaiseError("test.request: auth must be 'admin', 'viewer', or 'none', got: %s", auth)
		return 0
	}
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	s.harness.mux.ServeHTTP(recorder, req)

	// Build response table
	resp := L.NewTable()
	L.SetField(resp, "status", lua.LNumber(recorder.Code))

	respBody := recorder.Body.String()
	L.SetField(resp, "body", lua.LString(respBody))

	// Parse JSON if content type indicates it
	ct := recorder.Header().Get("Content-Type")
	if strings.Contains(ct, "application/json") && respBody != "" {
		jsonVal := plugin.GoValueToLua(L, parseJSONSafe(respBody))
		L.SetField(resp, "json", jsonVal)
	}

	// Headers
	headersTbl := L.NewTable()
	for k, vals := range recorder.Header() {
		if len(vals) > 0 {
			L.SetField(headersTbl, strings.ToLower(k), lua.LString(vals[0]))
		}
	}
	L.SetField(resp, "headers", headersTbl)

	L.Push(resp)
	return 1
}

// --- test.fire_hook ---

func (s *testAPIState) luaFireHook(L *lua.LState) int {
	event := L.CheckString(1)
	table := L.CheckString(2)
	dataTbl := L.CheckTable(3)

	data := luaTableToMap(L, dataTbl)
	ctx := s.injectAdminAuth(context.Background())
	hookEngine := s.harness.mgr.HookEngine()

	switch event {
	case "before_read":
		resp, state, err := hookEngine.RunBeforeReadHooks(ctx, table, data)
		// Return (response_table_or_nil, state_map_or_nil, error_string_or_nil)
		if resp != nil {
			respTbl := L.NewTable()
			L.SetField(respTbl, "status", lua.LNumber(resp.Status))
			L.SetField(respTbl, "headers", plugin.GoValueToLua(L, resp.Headers))
			L.SetField(respTbl, "body", plugin.GoValueToLua(L, resp.Body))
			L.Push(respTbl)
		} else {
			L.Push(lua.LNil)
		}
		if state != nil {
			L.Push(plugin.GoValueToLua(L, state))
		} else {
			L.Push(lua.LNil)
		}
		if err != nil {
			L.Push(lua.LString(err.Error()))
		} else {
			L.Push(lua.LNil)
		}
		return 3

	case "after_read":
		// Requires state parameter
		var state map[string]any
		if L.GetTop() >= 4 {
			stateTbl := L.OptTable(4, nil)
			if stateTbl != nil {
				state = luaTableToMap(L, stateTbl)
			}
		}
		if state == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("after_read requires a state parameter (4th argument)"))
			return 2
		}
		headers, err := hookEngine.RunAfterReadHooks(ctx, table, data, state)
		if headers != nil {
			L.Push(plugin.GoValueToLua(L, headers))
		} else {
			L.Push(lua.LNil)
		}
		if err != nil {
			L.Push(lua.LString(err.Error()))
		} else {
			L.Push(lua.LNil)
		}
		return 2

	case "before_create", "before_update", "before_delete", "before_publish":
		err := hookEngine.RunBeforeHooks(ctx, audited.HookEvent(event), table, data)
		if err != nil {
			L.Push(lua.LString(err.Error()))
		} else {
			L.Push(lua.LNil)
		}
		return 1

	case "after_create", "after_update", "after_delete", "after_publish":
		hookEngine.RunAfterHooks(ctx, audited.HookEvent(event), table, data)
		L.Push(lua.LNil)
		return 1

	default:
		L.RaiseError("test.fire_hook: unknown event %q", event)
		return 0
	}
}

// --- test.mock_request ---

func (s *testAPIState) luaMockRequest(L *lua.LState) int {
	method := L.CheckString(1)
	urlPattern := L.CheckString(2)
	respTbl := L.CheckTable(3)

	response := luaTableToMap(L, respTbl)
	s.mockReq.AddRule(method, urlPattern, response)
	return 0
}

// --- auth helpers ---

func (s *testAPIState) injectAdminAuth(ctx context.Context) context.Context {
	user := &db.Users{
		UserID:   types.NewUserID(),
		Username: "test-admin",
		Name:     "Test Admin",
		Email:    types.Email("admin@test.local"),
		Role:     "admin",
	}
	ctx = middleware.SetAuthenticatedUser(ctx, user)
	ctx = middleware.SetIsAdmin(ctx, true)

	// All permissions
	ps := middleware.PermissionSet{}
	for _, p := range allPermissions() {
		ps[p] = struct{}{}
	}
	ctx = middleware.SetPermissions(ctx, ps)
	return ctx
}

func (s *testAPIState) injectViewerAuth(ctx context.Context) context.Context {
	user := &db.Users{
		UserID:   types.NewUserID(),
		Username: "test-viewer",
		Name:     "Test Viewer",
		Email:    types.Email("viewer@test.local"),
		Role:     "viewer",
	}
	ctx = middleware.SetAuthenticatedUser(ctx, user)
	ctx = middleware.SetIsAdmin(ctx, false)

	ps := middleware.PermissionSet{}
	for _, p := range viewerPermissions() {
		ps[p] = struct{}{}
	}
	ctx = middleware.SetPermissions(ctx, ps)
	return ctx
}

func allPermissions() []string {
	resources := []string{
		"content", "datatypes", "fields", "media", "routes",
		"users", "roles", "permissions", "plugins", "settings",
		"backups", "webhooks", "locales", "validations",
		"admin_content", "admin_datatypes", "admin_fields", "admin_media", "admin_routes",
	}
	ops := []string{"create", "read", "update", "delete"}
	var perms []string
	for _, r := range resources {
		for _, o := range ops {
			perms = append(perms, r+":"+o)
		}
	}
	return perms
}

func viewerPermissions() []string {
	return []string{
		"content:read",
		"media:read",
		"routes:read",
		"datatypes:read",
		"fields:read",
	}
}

// --- Lua helpers ---

// luaTableToMap converts a Lua table to a Go map. Nested tables are recursively
// converted. Non-table values use plugin.LuaValueToGo.
func luaTableToMap(L *lua.LState, tbl *lua.LTable) map[string]any {
	m := make(map[string]any)
	tbl.ForEach(func(k, v lua.LValue) {
		key := k.String()
		if vt, ok := v.(*lua.LTable); ok {
			m[key] = luaTableToMap(L, vt)
		} else {
			m[key] = plugin.LuaValueToGo(v)
		}
	})
	return m
}

// luaDeepEqual compares two Lua values for deep equality.
func luaDeepEqual(a, b lua.LValue) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch at := a.(type) {
	case *lua.LTable:
		bt, ok := b.(*lua.LTable)
		if !ok {
			return false
		}
		return luaTablesEqual(at, bt)
	default:
		return a.String() == b.String()
	}
}

// luaTablesEqual does a deep comparison of two Lua tables.
func luaTablesEqual(a, b *lua.LTable) bool {
	countA := 0
	equal := true
	a.ForEach(func(k, v lua.LValue) {
		countA++
		bv := b.RawGet(k)
		if !luaDeepEqual(v, bv) {
			equal = false
		}
	})
	if !equal {
		return false
	}
	countB := 0
	b.ForEach(func(k, v lua.LValue) {
		countB++
	})
	return countA == countB
}

// luaFormat returns a human-readable representation of a Lua value.
func luaFormat(v lua.LValue) string {
	if v == nil || v == lua.LNil {
		return "nil"
	}
	switch v.Type() {
	case lua.LTString:
		return fmt.Sprintf("%q", v.String())
	default:
		return v.String()
	}
}

// parseJSONSafe parses a JSON string into a generic value, returning nil on error.
func parseJSONSafe(s string) any {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil
	}
	return v
}
