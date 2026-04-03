package plugin

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

const (
	// MaxDomainsPerPlugin is the maximum number of external domains a plugin can register.
	MaxDomainsPerPlugin = 50
)

// PendingRequest represents a domain registration discovered during init.lua execution.
type PendingRequest struct {
	Domain      string
	Description string
}

// requestAPIState holds per-VM state for the request API.
// Each instance is bound to exactly one LState (1:1 invariant), same as DatabaseAPI.
// INVARIANT: never share across VMs.
type requestAPIState struct {
	engine       OutboundExecutor // nil until RequestEngine is wired
	pluginName   string
	inBeforeHook bool
}

// OutboundExecutor executes outbound HTTP requests.
// Implemented by RequestEngine (request_engine.go).
type OutboundExecutor interface {
	Execute(ctx context.Context, pluginName, method, urlStr string, opts OutboundRequestOpts) (map[string]any, error)
}

// OutboundRequestOpts contains options for an outbound HTTP request.
type OutboundRequestOpts struct {
	Headers   map[string]string
	Body      string // raw body (mutually exclusive with JSONBody)
	JSONBody  any    // value to JSON-marshal (mutually exclusive with Body)
	Timeout   int    // seconds, 0 means use engine default
	ParseJSON bool   // opt-in: parse response body as JSON into resp.json
}

// RegisterRequestAPI creates a "request" Lua table with all outbound request functions
// and sets it as a global. Returns the requestAPIState for before-hook guard wiring.
//
// After calling RegisterRequestAPI, the caller should call FreezeModule(L, "request")
// to make the module read-only.
func RegisterRequestAPI(L *lua.LState, pluginName string, engine OutboundExecutor) *requestAPIState {
	state := &requestAPIState{
		engine:     engine,
		pluginName: pluginName,
	}

	// Create hidden pending table for request.register() calls.
	pending := L.NewTable()
	L.SetGlobal("__request_pending", pending)

	reqTable := L.NewTable()
	reqTable.RawSetString("register", L.NewFunction(requestRegisterFn(pluginName, pending)))
	reqTable.RawSetString("send", L.NewFunction(state.luaSend))
	reqTable.RawSetString("get", L.NewFunction(state.luaMethodShorthand("GET")))
	reqTable.RawSetString("post", L.NewFunction(state.luaMethodShorthand("POST")))
	reqTable.RawSetString("put", L.NewFunction(state.luaMethodShorthand("PUT")))
	reqTable.RawSetString("delete", L.NewFunction(state.luaMethodShorthand("DELETE")))
	reqTable.RawSetString("patch", L.NewFunction(state.luaMethodShorthand("PATCH")))

	L.SetGlobal("request", reqTable)

	return state
}

// requestRegisterFn returns the Go-bound function for request.register(domain [, opts]).
// Only callable at module scope (__vm_phase == "module_scope").
func requestRegisterFn(pluginName string, pending *lua.LTable) lua.LGFunction {
	return func(L *lua.LState) int {
		domain := L.CheckString(1)

		// Phase guard: request.register() can only be called at module scope.
		registryTbl := L.Get(lua.RegistryIndex)
		if regTbl, ok := registryTbl.(*lua.LTable); ok {
			phase := L.GetField(regTbl, "__vm_phase")
			if ps, ok := phase.(lua.LString); ok && string(ps) != "module_scope" {
				L.RaiseError("request.register() must be called at module scope, not inside on_init() or at runtime")
				return 0
			}
		}

		// Validate domain.
		if err := validateDomain(domain); err != nil {
			L.ArgError(1, err.Error())
			return 0
		}

		// Count limit.
		count := 0
		pending.ForEach(func(_, _ lua.LValue) {
			count++
		})
		if count >= MaxDomainsPerPlugin {
			L.RaiseError("plugin %q exceeded maximum domain registration limit (%d)", pluginName, MaxDomainsPerPlugin)
			return 0
		}

		// Duplicate detection (idempotent: silently ignore re-registration).
		normalizedDomain := strings.ToLower(domain)
		existing := L.GetField(pending, normalizedDomain)
		if existing != lua.LNil {
			return 0
		}

		// Parse optional description from opts table.
		description := ""
		if L.GetTop() >= 2 {
			optVal := L.Get(2)
			if optTbl, ok := optVal.(*lua.LTable); ok {
				descVal := L.GetField(optTbl, "description")
				if s, ok := descVal.(lua.LString); ok {
					description = string(s)
				}
			}
		}

		// Store in pending table keyed by normalized domain.
		entry := L.NewTable()
		entry.RawSetString("domain", lua.LString(normalizedDomain))
		entry.RawSetString("description", lua.LString(description))
		L.SetField(pending, normalizedDomain, entry)

		return 0
	}
}

// luaSend implements request.send(method, url [, opts]).
func (state *requestAPIState) luaSend(L *lua.LState) int {
	method := strings.ToUpper(L.CheckString(1))
	urlStr := L.CheckString(2)

	// Parse opts from arg 3 if present.
	var optsTbl *lua.LTable
	if L.GetTop() >= 3 {
		if tbl, ok := L.Get(3).(*lua.LTable); ok {
			optsTbl = tbl
		}
	}

	return state.doSend(L, method, urlStr, optsTbl)
}

// luaMethodShorthand returns a convenience wrapper for request.get/post/put/delete/patch.
// Each calls the same send logic with a fixed HTTP method.
func (state *requestAPIState) luaMethodShorthand(method string) lua.LGFunction {
	return func(L *lua.LState) int {
		urlStr := L.CheckString(1)

		var optsTbl *lua.LTable
		if L.GetTop() >= 2 {
			if tbl, ok := L.Get(2).(*lua.LTable); ok {
				optsTbl = tbl
			}
		}

		return state.doSend(L, method, urlStr, optsTbl)
	}
}

// doSend is the core send implementation shared by request.send and convenience methods.
func (state *requestAPIState) doSend(L *lua.LState, method, urlStr string, optsTbl *lua.LTable) int {
	// 1. Phase guard: only allowed during "runtime".
	registryTbl := L.Get(lua.RegistryIndex)
	if regTbl, ok := registryTbl.(*lua.LTable); ok {
		phase := L.GetField(regTbl, "__vm_phase")
		if ps, ok := phase.(lua.LString); ok {
			switch string(ps) {
			case "runtime":
				// allowed
			case "module_scope":
				L.RaiseError("request.send() cannot be called at module scope")
				return 0
			case "init":
				L.RaiseError("request.send() cannot be called inside on_init()")
				return 0
			case "shutdown":
				L.RaiseError("request.send() cannot be called inside on_shutdown()")
				return 0
			default:
				L.RaiseError("request.send() cannot be called in current VM phase %q", string(ps))
				return 0
			}
		}
	}

	// 2. Before-hook guard.
	if state.inBeforeHook {
		L.RaiseError("request.send() cannot be called inside a before-hook handler (would hold DB transaction locks)")
		return 0
	}

	// 3. Validate HTTP method.
	switch method {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
		// valid
	default:
		L.RaiseError("request.send(): invalid HTTP method %q", method)
		return 0
	}

	// 4. Parse and validate URL.
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		L.RaiseError("request.send(): invalid URL: %s", err.Error())
		return 0
	}
	if parsedURL.User != nil {
		L.RaiseError("request.send(): URLs with userinfo (user:pass@host) are not allowed")
		return 0
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		L.RaiseError("request.send(): URL must include scheme and host (e.g., https://example.com)")
		return 0
	}

	// 5. Parse options.
	opts := OutboundRequestOpts{}
	hasBody := false
	hasJSON := false

	if optsTbl != nil {
		// Headers.
		headersVal := L.GetField(optsTbl, "headers")
		if headersTbl, ok := headersVal.(*lua.LTable); ok {
			opts.Headers = make(map[string]string)
			headersTbl.ForEach(func(k, v lua.LValue) {
				if ks, ok := k.(lua.LString); ok {
					if vs, ok := v.(lua.LString); ok {
						opts.Headers[string(ks)] = string(vs)
					}
				}
			})
		}

		// Body (string).
		bodyVal := L.GetField(optsTbl, "body")
		if bs, ok := bodyVal.(lua.LString); ok {
			opts.Body = string(bs)
			hasBody = true
		}

		// JSON body.
		jsonVal := L.GetField(optsTbl, "json")
		if jsonVal != lua.LNil {
			opts.JSONBody = LuaValueToGo(jsonVal)
			hasJSON = true
		}

		// Timeout.
		timeoutVal := L.GetField(optsTbl, "timeout")
		if num, ok := timeoutVal.(lua.LNumber); ok {
			opts.Timeout = int(num)
		}

		// Parse JSON response.
		parseJSONVal := L.GetField(optsTbl, "parse_json")
		if b, ok := parseJSONVal.(lua.LBool); ok {
			opts.ParseJSON = bool(b)
		}
	}

	// 6. Validate body/json exclusivity.
	if hasBody && hasJSON {
		L.RaiseError("request.send(): 'body' and 'json' options are mutually exclusive")
		return 0
	}

	// 7. Check engine availability.
	if state.engine == nil {
		errTbl := L.NewTable()
		errTbl.RawSetString("error", lua.LString("outbound requests not configured"))
		L.Push(errTbl)
		return 1
	}

	// 8. Execute via engine.
	ctx := L.Context()
	if ctx == nil {
		L.RaiseError("request.send(): no context set on Lua state (VM setup error)")
		return 0
	}

	result, execErr := state.engine.Execute(ctx, state.pluginName, method, urlStr, opts)
	if execErr != nil {
		errTbl := L.NewTable()
		errTbl.RawSetString("error", lua.LString(execErr.Error()))
		L.Push(errTbl)
		return 1
	}

	// 9. Convert result map to Lua table.
	respTbl := L.NewTable()
	if status, ok := result["status"]; ok {
		if statusInt, ok := status.(int); ok {
			respTbl.RawSetString("status", lua.LNumber(statusInt))
		}
	}
	if body, ok := result["body"]; ok {
		if bodyStr, ok := body.(string); ok {
			respTbl.RawSetString("body", lua.LString(bodyStr))
		}
	}
	if headers, ok := result["headers"]; ok {
		if headerMap, ok := headers.(map[string]string); ok {
			headersTbl := L.NewTable()
			for k, v := range headerMap {
				headersTbl.RawSetString(k, lua.LString(v))
			}
			respTbl.RawSetString("headers", headersTbl)
		}
	}
	if jsonData, ok := result["json"]; ok {
		respTbl.RawSetString("json", GoValueToLua(L, jsonData))
	}
	if errMsg, ok := result["error"]; ok {
		if errStr, ok := errMsg.(string); ok {
			respTbl.RawSetString("error", lua.LString(errStr))
		}
	}

	L.Push(respTbl)
	return 1
}

// ReadPendingRequests extracts pending domain registrations from a VM's
// __request_pending table. Called after init.lua execution to discover
// which domains the plugin wants to access.
func ReadPendingRequests(L *lua.LState) []PendingRequest {
	pendingVal := L.GetGlobal("__request_pending")
	pendingTbl, ok := pendingVal.(*lua.LTable)
	if !ok || pendingVal == lua.LNil {
		return nil
	}

	var requests []PendingRequest
	pendingTbl.ForEach(func(key, value lua.LValue) {
		entry, ok := value.(*lua.LTable)
		if !ok {
			return
		}

		domainVal := L.GetField(entry, "domain")
		domain, ok := domainVal.(lua.LString)
		if !ok {
			return
		}

		descVal := L.GetField(entry, "description")
		description := ""
		if s, ok := descVal.(lua.LString); ok {
			description = string(s)
		}

		requests = append(requests, PendingRequest{
			Domain:      string(domain),
			Description: description,
		})
	})

	return requests
}

// validateDomain checks that a domain string is well-formed for outbound request registration.
// Rules: no scheme, no path, no port, no wildcards, [a-zA-Z0-9.-] only, max 253 chars.
func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	if len(domain) > 253 {
		return fmt.Errorf("domain %q exceeds 253 characters", domain)
	}

	// Reject scheme prefixes.
	if strings.Contains(domain, "://") {
		return fmt.Errorf("domain must not include scheme (e.g., use 'example.com' not 'https://example.com')")
	}

	// Reject paths and ports.
	if strings.Contains(domain, "/") {
		return fmt.Errorf("domain must not include path")
	}
	if strings.Contains(domain, ":") {
		return fmt.Errorf("domain must not include port")
	}

	// Reject wildcards.
	if strings.Contains(domain, "*") {
		return fmt.Errorf("domain must not contain wildcards")
	}

	// Character validation: a-zA-Z0-9.- only.
	for _, r := range domain {
		if !isValidDomainChar(r) {
			return fmt.Errorf("domain contains invalid character %q (allowed: a-zA-Z0-9.-)", string(r))
		}
	}

	// Must contain at least one dot (reject bare hostnames).
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("domain must contain at least one dot (e.g., 'example.com')")
	}

	return nil
}

// isValidDomainChar returns true if the rune is allowed in a domain name.
func isValidDomainChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '.' || r == '-'
}

// VMPhase reads the __vm_phase value from the LState registry.
// Returns empty string if not set.
func VMPhase(L *lua.LState) string {
	registryTbl := L.Get(lua.RegistryIndex)
	regTbl, ok := registryTbl.(*lua.LTable)
	if !ok {
		return ""
	}
	phase := L.GetField(regTbl, "__vm_phase")
	if ps, ok := phase.(lua.LString); ok {
		return string(ps)
	}
	return ""
}

// SetVMPhase sets the __vm_phase value in the LState registry.
func SetVMPhase(L *lua.LState, phase string) {
	registryTbl := L.Get(lua.RegistryIndex)
	if regTbl, ok := registryTbl.(*lua.LTable); ok {
		L.SetField(regTbl, "__vm_phase", lua.LString(phase))
	}
}
