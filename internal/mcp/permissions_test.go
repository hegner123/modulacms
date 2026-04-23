package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestPermissionMapCompleteness verifies that every tool registered on the
// MCP server either has an entry in toolPermissions or is in the publicTools
// set. Connection tools (list_projects, switch_project, get_connection) are
// excluded because they are only registered when cm != nil, and this test
// creates the server with cm=nil.
func TestPermissionMapCompleteness(t *testing.T) {
	backends := newNilBackends()
	srv := newServer(backends, nil)

	// Send a tools/list JSON-RPC request to collect all registered tool names.
	listReq := mcp.JSONRPCRequest{
		JSONRPC: mcp.JSONRPC_VERSION,
		ID:      mcp.NewRequestId(1),
		Request: mcp.Request{Method: "tools/list"},
	}
	reqBytes, err := json.Marshal(listReq)
	if err != nil {
		t.Fatalf("marshal tools/list request: %v", err)
	}

	resp := srv.HandleMessage(context.Background(), reqBytes)
	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var rpcResp struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		t.Fatalf("unmarshal tools/list response: %v", err)
	}

	if len(rpcResp.Result.Tools) == 0 {
		t.Fatal("no tools registered on server")
	}

	var missing []string
	for _, tool := range rpcResp.Result.Tools {
		if _, inPerms := toolPermissions[tool.Name]; inPerms {
			continue
		}
		if publicTools[tool.Name] {
			continue
		}
		missing = append(missing, tool.Name)
	}

	if len(missing) > 0 {
		t.Errorf("tools registered without permission mapping or public designation: %v", missing)
	}
}

// TestPermissionLabelsValid verifies that every entry in toolPermissions
// passes middleware.ValidatePermissionLabel. The init() function in
// permissions.go already does this at startup (panics on invalid labels),
// so this test is a belt-and-suspenders check.
func TestPermissionLabelsValid(t *testing.T) {
	// If we get here without a panic from init(), all labels are valid.
	if len(toolPermissions) == 0 {
		t.Fatal("toolPermissions is empty")
	}
}

// TestPermissionLabelsExistInBootstrap verifies that every permission label
// in toolPermissions corresponds to a label that exists in the bootstrap
// permission set (rbacPermissionLabels in internal/db/db.go). This catches
// phantom labels: syntactically valid labels that pass format validation
// but have no matching row in the permissions table, so no role can ever
// include them.
//
// The bootstrap set is duplicated here as a second source of truth. When
// bootstrap labels change, this test catches the drift.
func TestPermissionLabelsExistInBootstrap(t *testing.T) {
	// Canonical copy of rbacPermissionLabels from internal/db/db.go.
	// Keep this in sync with CreateBootstrapData.
	bootstrapLabels := map[string]struct{}{
		"content:read": {}, "content:create": {}, "content:update": {}, "content:delete": {}, "content:publish": {}, "content:admin": {},
		"datatypes:read": {}, "datatypes:create": {}, "datatypes:update": {}, "datatypes:delete": {}, "datatypes:admin": {},
		"fields:read": {}, "fields:create": {}, "fields:update": {}, "fields:delete": {}, "fields:admin": {},
		"media:read": {}, "media:create": {}, "media:update": {}, "media:delete": {}, "media:admin": {},
		"routes:read": {}, "routes:create": {}, "routes:update": {}, "routes:delete": {}, "routes:admin": {},
		"users:read": {}, "users:create": {}, "users:update": {}, "users:delete": {}, "users:admin": {},
		"roles:read": {}, "roles:create": {}, "roles:update": {}, "roles:delete": {}, "roles:admin": {},
		"permissions:read": {}, "permissions:create": {}, "permissions:update": {}, "permissions:delete": {}, "permissions:admin": {},
		"sessions:read": {}, "sessions:delete": {}, "sessions:admin": {},
		"ssh_keys:read": {}, "ssh_keys:create": {}, "ssh_keys:delete": {}, "ssh_keys:admin": {},
		"config:read": {}, "config:update": {}, "config:admin": {},
		"admin_tree:read": {}, "admin_tree:create": {}, "admin_tree:update": {}, "admin_tree:delete": {}, "admin_tree:admin": {},
		"field_types:read": {}, "field_types:create": {}, "field_types:update": {}, "field_types:delete": {}, "field_types:admin": {},
		"admin_field_types:read": {}, "admin_field_types:create": {}, "admin_field_types:update": {}, "admin_field_types:delete": {}, "admin_field_types:admin": {},
		"deploy:read": {}, "deploy:create": {},
		"validations:read": {}, "validations:create": {}, "validations:update": {}, "validations:delete": {}, "validations:admin": {},
		"admin_validations:read": {}, "admin_validations:create": {}, "admin_validations:update": {}, "admin_validations:delete": {}, "admin_validations:admin": {},
		"webhook:read": {}, "webhook:create": {}, "webhook:update": {}, "webhook:delete": {}, "webhook:admin": {},
		"plugins:read": {}, "plugins:admin": {},
		"tables:read": {}, "tables:create": {}, "tables:update": {}, "tables:delete": {}, "tables:admin": {},
		"import:read": {}, "import:create": {}, "import:admin": {},
		"tokens:read": {}, "tokens:create": {}, "tokens:delete": {}, "tokens:admin": {},
		"locale:read": {}, "locale:create": {}, "locale:update": {}, "locale:delete": {}, "locale:admin": {},
		"audit:read": {}, "audit:admin": {},
		"backup:read": {}, "backup:create": {}, "backup:update": {}, "backup:delete": {}, "backup:admin": {},
		"search:read": {}, "search:update": {}, "search:admin": {},
	}

	// Known phantom labels documented as out of scope in the current plan.
	// Each entry maps a tool name to its phantom label along with the reason
	// for the exception. Remove entries as they are fixed.
	//
	// Admin-prefixed tools: addressed by admin resource permission separation plan.
	// health:read: decision needed on publicTools vs bootstrap addition.
	// sessions:update: broader RBAC bootstrap gap affecting both router and MCP.
	// plugins:update/delete: plugins are admin-only; granular perms tracked separately.
	allowedPhantoms := map[string]string{
		// Admin publishing/version/translation (admin separation plan)
		"admin_publish_content":          "publishing:create",
		"admin_unpublish_content":        "publishing:delete",
		"admin_schedule_content":         "publishing:create",
		"admin_list_content_versions":    "versions:read",
		"admin_get_content_version":      "versions:read",
		"admin_create_content_version":   "versions:create",
		"admin_delete_content_version":   "versions:delete",
		"admin_restore_content_version":  "versions:update",
		"admin_create_translation":       "locales:create",
		// Health tools (no health:* in bootstrap)
		"health":          "health:read",
		"get_metrics":     "health:read",
		"get_environment": "health:read",
		// Sessions gap
		"update_session": "sessions:update",
		// Plugin granular permissions gap
		"reload_plugin":         "plugins:update",
		"enable_plugin":         "plugins:update",
		"disable_plugin":        "plugins:update",
		"plugin_cleanup_drop":   "plugins:delete",
		"approve_plugin_routes": "plugins:update",
		"revoke_plugin_routes":  "plugins:update",
		"approve_plugin_hooks":  "plugins:update",
		"revoke_plugin_hooks":   "plugins:update",
	}

	var phantoms []string
	for tool, label := range toolPermissions {
		if _, exists := bootstrapLabels[label]; exists {
			continue
		}
		// Check if this is a known allowed phantom.
		if expected, ok := allowedPhantoms[tool]; ok && expected == label {
			continue
		}
		phantoms = append(phantoms, tool+"->"+label)
	}

	if len(phantoms) > 0 {
		t.Errorf("tools reference permission labels that do not exist in bootstrap data (phantom labels): %v", phantoms)
	}
}

// TestPermissionMapNoDuplicateValues is not adversarial per se, but catches
// a class of copy-paste error: two tools that should have different
// permissions accidentally sharing the same label. This test verifies that
// the inverse map (permission -> tools) contains only expected groupings.
// If a permission label maps to more than 30 tools, something is wrong.
func TestPermissionMapNoDuplicateValues(t *testing.T) {
	inverse := make(map[string]int32)
	for _, label := range toolPermissions {
		inverse[label]++
	}

	const maxToolsPerLabel int32 = 30
	for label, count := range inverse {
		if count > maxToolsPerLabel {
			t.Errorf("permission %q is used by %d tools (max %d), possible copy-paste error", label, count, maxToolsPerLabel)
		}
	}
}

// newNilBackends creates a Backends with all fields set to nil-valued
// interface pointers. Tool handlers will panic if called, but the server
// can still register tools and list them.
func newNilBackends() *Backends {
	return &Backends{}
}
