# MCP RBAC Authentication Plan

## Problem

The MCP endpoint (`/mcp`) bypasses all RBAC. The current flow:

1. `cmd/serve.go:432` creates a synthetic audit context: `audited.Ctx(types.NewNodeID(), types.UserID("mcp-direct"), "", "127.0.0.1")`
2. The full HTTP middleware chain runs on `/mcp` requests (the MCP handler is registered on the same `mux` that `DefaultMiddlewareChain` wraps at line 427). This means `RequestIDMiddleware`, `ClientIPMiddleware`, `HTTPAuthenticationMiddleware`, and `PermissionInjector` all execute and populate the request context with the authenticated user, permissions, request ID, and client IP. However, `HTTPPublicEndpointMiddleware` does NOT enforce authentication on `/mcp` because the path does not start with `/api` (see `http_middleware.go:112`). The middleware chain populates auth context if valid credentials are provided, but does not reject unauthenticated requests at the HTTP layer. Authentication enforcement for MCP is deferred entirely to the tool middleware (Phase 2, Step 2.2).
3. `mcp.APIKeyAuth` (`internal/mcp/serve.go:105`) ignores the middleware's auth results and does its own constant-time comparison of the Bearer token against the static `cfg.MCP_API_Key` value. If the static key matches, the request proceeds regardless of the user's identity or role.
4. Service backends call `svc.Content.Create(ctx, b.ac, p)` using a static `audited.AuditContext` created at startup, not the authenticated user from context.
5. The service layer trusts its caller unconditionally.

The middleware chain resolves the real user, but the MCP layer ignores it. Anyone with the `mcp_api_key` has unrestricted admin access to every operation. Audit trails show `"mcp-direct"` instead of the actual user.

## Existing Infrastructure

These components already exist and can be reused directly:

| Component | Location | What It Does |
|-----------|----------|-------------|
| `middleware.APIKeyAuth` | `internal/middleware/middleware.go:49` | Resolves Bearer token to `*db.Users` via `GetTokenByTokenValue`, checks type/revoked/expired |
| `PermissionCache.PermissionsForRole` | `internal/middleware/authorization.go:119` | Returns `PermissionSet` with `Has(perm)` for O(1) checks |
| `PermissionCache.IsAdmin` | `internal/middleware/authorization.go:134` | Returns whether a role is the admin role |
| `middleware.AuthenticatedUser` | `internal/middleware/http_middleware.go:138` | Extracts `*db.Users` from context (set by `HTTPAuthenticationMiddleware`) |
| `middleware.ContextPermissions` | `internal/middleware/authorization.go:175` | Extracts `PermissionSet` from context (set by `PermissionInjector`) |
| `middleware.ContextIsAdmin` | `internal/middleware/authorization.go:181` | Extracts admin boolean from context (set by `PermissionInjector`) |
| `middleware.RequestIDFromContext` | `internal/middleware/request_id.go` | Extracts request ID string from context |
| `middleware.ClientIPFromContext` | `internal/middleware/client_ip.go` | Extracts client IP string from context |
| `mcp-go` `WithHTTPContextFunc` | `vendor/.../server/streamable_http.go:68` | `StreamableHTTPOption`: injects context values from HTTP request into MCP tool handler context. Starts from `r.Context()`. |
| `mcp-go` `WithToolHandlerMiddleware` | `vendor/.../server/server.go:207` | `ServerOption`: wraps every tool call with middleware that sees `(ctx, CallToolRequest)`. Must be passed to `NewMCPServer`, not to transport constructors. |
| Permission labels | `internal/router/mux.go` | Every router endpoint already has its required `resource:operation` label |
| `audited.AuditContext` | `internal/db/audited/context.go:9` | Carries `NodeID`, `UserID`, `RequestID`, `IP`, `HookRunner` for audit trail. `Ctx()` constructor leaves `HookRunner` nil (no plugin hooks). MCP tool mutations will not trigger plugin hooks until `HookRunner` is wired in separately. |

## Design

### Architecture

Two layers work together:

1. **HTTP layer** (`WithHTTPContextFunc`): The `DefaultMiddlewareChain` already runs on `/mcp` requests and populates the context with the authenticated user (`middleware.AuthenticatedUser`), permission set (`middleware.ContextPermissions`), admin status (`middleware.ContextIsAdmin`), request ID, and client IP. The `WithHTTPContextFunc` callback is a no-op pass-through for HTTP mode because all auth data is already in context. It exists only as the extension point where stdio-mode fallback values could be injected in the future.

2. **Tool middleware** (`WithToolHandlerMiddleware`): Before each tool call, looks up the tool name in the permission map, reads the user's `PermissionSet` from context via `middleware.ContextPermissions(ctx)` and `middleware.ContextIsAdmin(ctx)`, and returns a 403-equivalent error if denied. Admin bypass works identically to the router: if the user's role is admin, the check is skipped. No MCP-specific context keys are needed for HTTP mode.

The `AuditContext` is no longer static. Instead of a single `audited.AuditContext` baked into the `Backends`, each service backend extracts the authenticated user from context via `middleware.AuthenticatedUser(ctx)` and builds the `AuditContext` per-call, including the real request ID and client IP from the middleware chain.

### Permission Map

A static `map[string]string` mapping each MCP tool name to its required `resource:operation` permission label. The mapping mirrors what the HTTP router uses for the equivalent API endpoint.

Tools that have no router equivalent (connection management: `list_projects`, `switch_project`, `get_connection`) are excluded from permission checks since they are registry-mode only (stdio) and never run in direct mode.

### Backward Compatibility

The `mcp_api_key` config field is renamed to `mcp_proxy_token` and its purpose is narrowed:

- **Before:** Static shared secret used both as server-side MCP auth and client-side proxy credential
- **After:** Renamed to `mcp_proxy_token`. Used exclusively by proxy mode (`connection.go`) to authenticate the outbound SDK client against a remote CMS. The server-side MCP authentication function (`mcp.APIKeyAuth`) is deleted. MCP endpoint authentication is now handled by the existing HTTP middleware chain (API tokens resolved via `GetTokenByTokenValue`).

Since ModulaCMS is greenfield with no active users, the rename has no migration concern.

## Implementation

### Phase 1: Auth Context Injection

**Goal:** Replace static API key check with real user resolution.

#### Step 1.1: Create `internal/mcp/auth.go`

New file with a single context key type for the stdio fallback path, plus the `AuditContextFromMCP` helper (detailed in Phase 3):

```go
// mcpAuditKey is used only in stdio mode to inject a default AuditContext
// when no HTTP middleware chain is available. In HTTP mode, the middleware
// chain populates the context with the authenticated user, permissions,
// and admin status directly. The tool middleware reads those values using
// the existing middleware extractors:
//   - middleware.AuthenticatedUser(ctx)
//   - middleware.ContextPermissions(ctx)
//   - middleware.ContextIsAdmin(ctx)
//   - middleware.RequestIDFromContext(ctx)
//   - middleware.ClientIPFromContext(ctx)
//
// No MCP-specific duplicates of these extractors are created.
type mcpAuditKey struct{}
```

The MCP package does NOT define its own context keys for user, permissions, or admin status. For HTTP mode, the `DefaultMiddlewareChain` already populates these values and the middleware package exports the extractors needed to read them. The `mcpAuditKey` is used only by `injectAuditContextMiddleware` (stdio mode) to store a default `audited.AuditContext` for offline/local operation.

#### Step 1.2: Create `internal/mcp/auth_middleware.go`

New file containing three functions (the latter two are defined in later steps):

1. `PassthroughHTTPContextFunc` (this step):

```go
func PassthroughHTTPContextFunc() server.HTTPContextFunc {
    return func(ctx context.Context, r *http.Request) context.Context {
        return ctx
    }
}
```

No-op: the `DefaultMiddlewareChain` already populates all auth context values before this runs. See Step 1.1 for the full list of middleware extractors.

2. `PermissionMiddleware` (defined in Step 2.2)
3. `injectAuditContextMiddleware` (defined in Step 3.4)

#### Step 1.3: Update `newServer` and `DirectHandler` signatures

**File:** `internal/mcp/serve.go`

`WithToolHandlerMiddleware` is a `ServerOption` (type `func(*MCPServer)`), not a `StreamableHTTPOption`. It must be passed when creating the `MCPServer`, not the `StreamableHTTPServer`. Update `newServer` to accept variadic server options:

```go
// Before:
func newServer(backends *Backends, cm *ConnectionManager) *server.MCPServer

// After:
func newServer(backends *Backends, cm *ConnectionManager, opts ...server.ServerOption) *server.MCPServer
```

Inside `newServer`, after calling `server.NewMCPServer("modula", utility.Version)`, apply the variadic opts. The permission middleware is passed here.

Change `DirectHandler` to accept only `*service.Registry` (no `audited.AuditContext`, no `*config.Config`, no `*middleware.PermissionCache`):

```go
// Before:
func DirectHandler(svc *service.Registry, ac audited.AuditContext) http.Handler

// After:
func DirectHandler(svc *service.Registry) http.Handler
```

The function creates the `MCPServer` with the permission middleware, then creates the `StreamableHTTPServer` with a passthrough `HTTPContextFunc`:

```go
func DirectHandler(svc *service.Registry) http.Handler {
    backends := NewServiceBackends(svc)
    srv := newServer(backends, nil,
        server.WithToolHandlerMiddleware(PermissionMiddleware()),
    )
    return server.NewStreamableHTTPServer(srv,
        server.WithEndpointPath("/mcp"),
        server.WithHTTPContextFunc(PassthroughHTTPContextFunc()),
    )
}
```

`WithHTTPContextFunc` goes on `NewStreamableHTTPServer` (it is a `StreamableHTTPOption`).
`WithToolHandlerMiddleware` goes on `newServer`/`NewMCPServer` (it is a `ServerOption`).
These are different option types and must not be mixed.

#### Step 1.4: Update `cmd/serve.go`

**File:** `cmd/serve.go`

Replace the static audit context setup:

```go
// Before:
mcpAC := audited.Ctx(types.NewNodeID(), types.UserID("mcp-direct"), "", "127.0.0.1")
mcpHandler := mcpserver.DirectHandler(svc, mcpAC)
mux.Handle("/mcp", mcpserver.APIKeyAuth(cfg.MCP_API_Key, mcpHandler))

// After:
mcpHandler := mcpserver.DirectHandler(svc)
mux.Handle("/mcp", mcpHandler)
```

The `APIKeyAuth` wrapper is removed. Authentication is handled by the existing `DefaultMiddlewareChain` which already wraps the mux and populates the request context with the authenticated user, permissions, and admin status. The MCP tool middleware reads these values directly from context.

### Phase 2: Per-Tool Permission Enforcement

**Goal:** Each tool call checks the authenticated user's permissions before executing.

#### Step 2.1: Create permission map in `internal/mcp/permissions.go`

New file containing a `var toolPermissions map[string]string` with every tool name mapped to its `resource:operation` label. The complete map (derived from the 246 `srv.AddTool` calls in `tools_*.go` and the corresponding router permission guards in `mux.go`).

**Drift prevention:** The completeness test in Step 4.4 catches missing entries. To catch incorrect mappings, the test also verifies that every permission string in the map passes `middleware.ValidatePermissionLabel`. Future consideration: generate this map from tool registration metadata rather than maintaining it by hand.

The full map:

**Content tools:**
| Tool | Permission |
|------|-----------|
| `list_content` | `content:read` |
| `get_content` | `content:read` |
| `create_content` | `content:create` |
| `update_content` | `content:update` |
| `delete_content` | `content:delete` |
| `get_page` | `content:read` |
| `get_content_tree` | `content:read` |
| `list_content_fields` | `content:read` |
| `get_content_field` | `content:read` |
| `create_content_field` | `content:create` |
| `update_content_field` | `content:update` |
| `delete_content_field` | `content:delete` |
| `reorder_content` | `content:update` |
| `move_content` | `content:update` |
| `save_content_tree` | `content:update` |
| `heal_content` | `content:update` |
| `batch_update_content` | `content:update` |
| `query_content` | `content:read` |
| `get_globals` | `content:read` |
| `get_content_full` | `content:read` |
| `get_content_by_route` | `content:read` |
| `create_content_composite` | `content:create` |

**Admin content tools:**
| Tool | Permission |
|------|-----------|
| `admin_list_content` | `content:read` |
| `admin_get_content` | `content:read` |
| `admin_create_content` | `content:create` |
| `admin_update_content` | `content:update` |
| `admin_delete_content` | `content:delete` |
| `admin_reorder_content` | `content:update` |
| `admin_move_content` | `content:update` |
| `admin_get_content_full` | `content:read` |
| `get_admin_tree` | `admin_tree:read` |
| `admin_list_content_fields` | `content:read` |
| `admin_get_content_field` | `content:read` |
| `admin_create_content_field` | `content:create` |
| `admin_update_content_field` | `content:update` |
| `admin_delete_content_field` | `content:delete` |

**Schema tools:**
| Tool | Permission |
|------|-----------|
| `list_datatypes` | `datatypes:read` |
| `get_datatype` | `datatypes:read` |
| `create_datatype` | `datatypes:create` |
| `update_datatype` | `datatypes:update` |
| `delete_datatype` | `datatypes:delete` |
| `list_fields` | `fields:read` |
| `get_field` | `fields:read` |
| `create_field` | `fields:create` |
| `update_field` | `fields:update` |
| `delete_field` | `fields:delete` |
| `get_datatype_full` | `datatypes:read` |
| `list_datatypes_full` | `datatypes:read` |
| `get_datatype_max_sort_order` | `datatypes:read` |
| `update_datatype_sort_order` | `datatypes:update` |
| `get_field_max_sort_order` | `fields:read` |
| `update_field_sort_order` | `fields:update` |
| `list_field_types` | `field_types:read` |
| `get_field_type` | `field_types:read` |
| `create_field_type` | `field_types:create` |
| `update_field_type` | `field_types:update` |
| `delete_field_type` | `field_types:delete` |

**Admin schema tools:**
| Tool | Permission |
|------|-----------|
| `admin_list_datatypes` | `datatypes:read` |
| `admin_get_datatype` | `datatypes:read` |
| `admin_create_datatype` | `datatypes:create` |
| `admin_update_datatype` | `datatypes:update` |
| `admin_delete_datatype` | `datatypes:delete` |
| `admin_get_datatype_max_sort_order` | `datatypes:read` |
| `admin_update_datatype_sort_order` | `datatypes:update` |
| `admin_list_fields` | `fields:read` |
| `admin_get_field` | `fields:read` |
| `admin_create_field` | `fields:create` |
| `admin_update_field` | `fields:update` |
| `admin_delete_field` | `fields:delete` |

**Media tools:**
| Tool | Permission |
|------|-----------|
| `list_media` | `media:read` |
| `get_media` | `media:read` |
| `update_media` | `media:update` |
| `delete_media` | `media:delete` |
| `upload_media` | `media:create` |
| `media_health` | `media:read` |
| `media_cleanup_check` | `media:read` |
| `media_cleanup_apply` | `media:delete` |
| `list_media_dimensions` | `media:read` |
| `get_media_dimension` | `media:read` |
| `create_media_dimension` | `media:create` |
| `update_media_dimension` | `media:update` |
| `delete_media_dimension` | `media:delete` |
| `download_media` | `media:read` |
| `get_media_full` | `media:read` |
| `get_media_references` | `media:read` |
| `reprocess_media` | `media:update` |

**Media folder tools:**
| Tool | Permission |
|------|-----------|
| `list_media_folders` | `media:read` |
| `get_media_folder` | `media:read` |
| `create_media_folder` | `media:create` |
| `update_media_folder` | `media:update` |
| `delete_media_folder` | `media:delete` |
| `move_media_to_folder` | `media:update` |
| `get_media_folder_tree` | `media:read` |
| `list_media_in_folder` | `media:read` |

**Admin media tools:**
| Tool | Permission |
|------|-----------|
| `admin_list_media` | `media:read` |
| `admin_get_media` | `media:read` |
| `admin_update_media` | `media:update` |
| `admin_delete_media` | `media:delete` |
| `admin_upload_media` | `media:create` |

**Admin media folder tools:**
| Tool | Permission |
|------|-----------|
| `admin_list_media_folders` | `media:read` |
| `admin_get_media_folder` | `media:read` |
| `admin_create_media_folder` | `media:create` |
| `admin_update_media_folder` | `media:update` |
| `admin_delete_media_folder` | `media:delete` |
| `admin_move_media_to_folder` | `media:update` |
| `admin_get_media_folder_tree` | `media:read` |
| `admin_list_media_in_folder` | `media:read` |

**Route tools:**
| Tool | Permission |
|------|-----------|
| `list_routes` | `routes:read` |
| `get_route` | `routes:read` |
| `list_routes_full` | `routes:read` |
| `create_route` | `routes:create` |
| `update_route` | `routes:update` |
| `delete_route` | `routes:delete` |

**Admin route tools:**
| Tool | Permission |
|------|-----------|
| `admin_list_routes` | `routes:read` |
| `admin_get_route_by_slug` | `routes:read` |
| `admin_create_route` | `routes:create` |
| `admin_update_route` | `routes:update` |
| `admin_delete_route` | `routes:delete` |
| `admin_list_field_types` | `field_types:read` |
| `admin_get_field_type` | `field_types:read` |
| `admin_create_field_type` | `field_types:create` |
| `admin_update_field_type` | `field_types:update` |
| `admin_delete_field_type` | `field_types:delete` |

**User tools:**
| Tool | Permission |
|------|-----------|
| `whoami` | `users:read` |
| `list_users` | `users:read` |
| `get_user` | `users:read` |
| `create_user` | `users:create` |
| `update_user` | `users:update` |
| `delete_user` | `users:delete` |
| `list_users_full` | `users:read` |
| `get_user_full` | `users:read` |
| `reassign_and_delete_user` | `users:delete` |
| `list_user_sessions` | `users:read` |

**RBAC tools:**
| Tool | Permission |
|------|-----------|
| `list_roles` | `roles:read` |
| `get_role` | `roles:read` |
| `create_role` | `roles:create` |
| `update_role` | `roles:update` |
| `delete_role` | `roles:delete` |
| `list_permissions` | `permissions:read` |
| `get_permission` | `permissions:read` |
| `create_permission` | `permissions:create` |
| `update_permission` | `permissions:update` |
| `delete_permission` | `permissions:delete` |
| `assign_role_permission` | `role_permissions:create` |
| `remove_role_permission` | `role_permissions:delete` |
| `list_role_permissions` | `role_permissions:read` |
| `get_role_permission` | `role_permissions:read` |
| `list_role_permissions_by_role` | `role_permissions:read` |

**Session tools:**
| Tool | Permission |
|------|-----------|
| `list_sessions` | `sessions:read` |
| `get_session` | `sessions:read` |
| `update_session` | `sessions:update` |
| `delete_session` | `sessions:delete` |

**Token tools:**
| Tool | Permission |
|------|-----------|
| `list_tokens` | `tokens:read` |
| `get_token` | `tokens:read` |
| `create_token` | `tokens:create` |
| `delete_token` | `tokens:delete` |

**SSH key tools:**
| Tool | Permission |
|------|-----------|
| `list_ssh_keys` | `ssh_keys:read` |
| `create_ssh_key` | `ssh_keys:create` |
| `delete_ssh_key` | `ssh_keys:delete` |

**OAuth tools:**
| Tool | Permission |
|------|-----------|
| `list_users_oauth` | `oauth:read` |
| `get_user_oauth` | `oauth:read` |
| `create_user_oauth` | `oauth:create` |
| `update_user_oauth` | `oauth:update` |
| `delete_user_oauth` | `oauth:delete` |

**Table tools:**
| Tool | Permission |
|------|-----------|
| `list_tables` | `tables:read` |
| `get_table` | `tables:read` |
| `create_table` | `tables:create` |
| `update_table` | `tables:update` |
| `delete_table` | `tables:delete` |

**Plugin tools:**
| Tool | Permission |
|------|-----------|
| `list_plugins` | `plugins:read` |
| `get_plugin` | `plugins:read` |
| `reload_plugin` | `plugins:update` |
| `enable_plugin` | `plugins:update` |
| `disable_plugin` | `plugins:update` |
| `plugin_cleanup_dry_run` | `plugins:read` |
| `plugin_cleanup_drop` | `plugins:delete` |
| `list_plugin_routes` | `plugins:read` |
| `approve_plugin_routes` | `plugins:update` |
| `revoke_plugin_routes` | `plugins:update` |
| `list_plugin_hooks` | `plugins:read` |
| `approve_plugin_hooks` | `plugins:update` |
| `revoke_plugin_hooks` | `plugins:update` |

**Config tools:**
| Tool | Permission |
|------|-----------|
| `get_config` | `config:read` |
| `get_config_meta` | `config:read` |
| `update_config` | `config:update` |

**Publishing tools:**
| Tool | Permission |
|------|-----------|
| `publish_content` | `publishing:create` |
| `unpublish_content` | `publishing:delete` |
| `schedule_content` | `publishing:create` |
| `admin_publish_content` | `publishing:create` |
| `admin_unpublish_content` | `publishing:delete` |
| `admin_schedule_content` | `publishing:create` |

**Version tools:**
| Tool | Permission |
|------|-----------|
| `list_content_versions` | `versions:read` |
| `get_content_version` | `versions:read` |
| `create_content_version` | `versions:create` |
| `delete_content_version` | `versions:delete` |
| `restore_content_version` | `versions:update` |
| `admin_list_content_versions` | `versions:read` |
| `admin_get_content_version` | `versions:read` |
| `admin_create_content_version` | `versions:create` |
| `admin_delete_content_version` | `versions:delete` |
| `admin_restore_content_version` | `versions:update` |

**Webhook tools:**
| Tool | Permission |
|------|-----------|
| `list_webhooks` | `webhooks:read` |
| `get_webhook` | `webhooks:read` |
| `create_webhook` | `webhooks:create` |
| `update_webhook` | `webhooks:update` |
| `delete_webhook` | `webhooks:delete` |
| `test_webhook` | `webhooks:update` |
| `list_webhook_deliveries` | `webhooks:read` |
| `retry_webhook_delivery` | `webhooks:update` |

**Locale tools:**
| Tool | Permission |
|------|-----------|
| `list_locales` | `locales:read` |
| `list_admin_locales` | `locales:read` |
| `get_locale` | `locales:read` |
| `create_locale` | `locales:create` |
| `update_locale` | `locales:update` |
| `delete_locale` | `locales:delete` |
| `create_translation` | `locales:create` |
| `admin_create_translation` | `locales:create` |

**Validation tools:**
| Tool | Permission |
|------|-----------|
| `list_validations` | `validations:read` |
| `get_validation` | `validations:read` |
| `create_validation` | `validations:create` |
| `update_validation` | `validations:update` |
| `delete_validation` | `validations:delete` |
| `search_validations` | `validations:read` |
| `admin_list_validations` | `validations:read` |
| `admin_get_validation` | `validations:read` |
| `admin_create_validation` | `validations:create` |
| `admin_update_validation` | `validations:update` |
| `admin_delete_validation` | `validations:delete` |
| `admin_search_validations` | `validations:read` |

**Search tools:**
| Tool | Permission |
|------|-----------|
| `search_content` | `search:read` |
| `rebuild_search_index` | `search:update` |

**Health/activity tools:**
| Tool | Permission |
|------|-----------|
| `health` | `health:read` |
| `get_metrics` | `health:read` |
| `get_environment` | `health:read` |
| `list_recent_activity` | `activity:read` |

**Import tools:**
| Tool | Permission |
|------|-----------|
| `import_content` | `import:create` |
| `import_bulk` | `import:create` |

**Deploy tools:**
| Tool | Permission |
|------|-----------|
| `sync_health` | `deploy:read` |
| `sync_export` | `deploy:read` |
| `sync_import` | `deploy:create` |
| `sync_preview` | `deploy:read` |

**Public tools (no permission required, in `publicTools` set):**

```go
var publicTools = map[string]bool{
    "register_user":          true,
    "request_password_reset": true,
}
```

Connection tools (`list_projects`, `switch_project`, `get_connection`) are NOT in `publicTools`. They are conditionally registered only when a `ConnectionManager` is provided (`cm != nil` in `newServer`). The completeness test in Step 4.4 creates a server with `cm=nil`, so these tools will not appear in the registered tool list and do not need an entry in either map.

#### Step 2.2: Create `PermissionMiddleware` in `internal/mcp/auth_middleware.go`

```go
// PermissionMiddleware returns a ToolHandlerMiddleware that checks the
// authenticated user's permissions before allowing a tool call.
// It reads auth state from context using the middleware package's
// exported extractors (not MCP-specific keys).
//
// Flow:
// 1. Look up request.Params.Name in toolPermissions map
// 2. If not in map (public or connection tools), allow unconditionally
// 3. If middleware.AuthenticatedUser(ctx) == nil, deny with
//    errAuthRequired (IsError: true). This gate runs BEFORE the admin
//    bypass check to avoid calling ContextIsAdmin on a context where
//    PermissionInjector may not have run.
// 4. If middleware.ContextIsAdmin(ctx) is true, allow (admin bypass)
// 5. ps := middleware.ContextPermissions(ctx); if ps is nil, deny with
//    errAuthRequired (IsError: true) -- defensive, should not happen
//    after step 3
// 6. If !ps.Has(permission), deny with fmt.Sprintf(errForbidden,
//    permission) (IsError: true)
// 7. Otherwise, call next(ctx, request)
//
// Error constants (exact strings, do not modify):
//   const errAuthRequired = "authentication required"
//   const errForbidden    = "forbidden: requires permission '%s'"
func PermissionMiddleware() server.ToolHandlerMiddleware
```

Denied tool calls return a `*mcp.CallToolResult` with `IsError: true` and a text content message like `"forbidden: requires permission 'content:create'"`. This matches MCP protocol semantics (tool errors are not transport errors).

### Phase 3: Per-Call Audit Context

**Goal:** Audit trails reflect the actual authenticated user, not "mcp-direct".

#### Step 3.1: Replace static `audited.AuditContext` in service backends

Every `svc*Backend` struct currently holds a single `ac audited.AuditContext` set at server creation time. This needs to become per-call.

**Approach:** Add a helper function to `internal/mcp/auth.go`:

```go
// AuditContextFromMCP builds an AuditContext from the authenticated user
// in the MCP context. Uses the real request ID and client IP from the
// middleware chain when available. Falls back to a "mcp-anonymous" context
// if no user is present (should not happen after permission middleware,
// but defensive).
func AuditContextFromMCP(ctx context.Context) audited.AuditContext
```

Implementation:
1. Extract `middleware.AuthenticatedUser(ctx)` to get the `*db.Users`
2. Extract `middleware.RequestIDFromContext(ctx)` for the request ID (defaults to `""` if absent, which happens in stdio mode)
3. Extract `middleware.ClientIPFromContext(ctx)` for the client IP (defaults to `"127.0.0.1"` if absent, which happens in stdio mode)
4. If user is non-nil, build `audited.Ctx(types.NewNodeID(), user.UserID, requestID, clientIP)`
5. If user is nil, check for a stdio-mode fallback `audited.AuditContext` stored under `mcpAuditKey{}` in context. If present, use it. Otherwise, return `audited.Ctx(types.NewNodeID(), types.UserID("mcp-anonymous"), requestID, clientIP)`

#### Step 3.2: Update service backend structs

Remove the `ac audited.AuditContext` field from every `svc*Backend` struct. Replace every `b.ac` reference with `AuditContextFromMCP(ctx)`. `AuditContextFromMCP` reads the authenticated user via `middleware.AuthenticatedUser(ctx)` and the request ID/client IP via `middleware.RequestIDFromContext(ctx)` and `middleware.ClientIPFromContext(ctx)`. In stdio mode, it falls back to the `mcpAuditKey{}` context value injected by `injectAuditContextMiddleware`.

**Affected files:**
- `backend_service_content.go` (3 structs: `svcContentBackend`, `svcAdminContentBackend`, `svcVersionBackend`)
- `backend_service_schema.go` (2 structs: `svcSchemaBackend`, `svcAdminSchemaBackend`)
- `backend_service_media.go` (4 structs: `svcMediaBackend`, `svcMediaFolderBackend`, `svcAdminMediaBackend`, `svcAdminMediaFolderBackend`)
- `backend_service_routes.go` (2 structs: `svcRouteBackend`, `svcAdminRouteBackend`)
- `backend_service_users.go` (6 structs: `svcUserBackend`, `svcRBACBackend`, `svcSessionBackend`, `svcTokenBackend`, `svcSSHKeyBackend`, `svcOAuthBackend`)
- `backend_service_webhooks.go` (1 struct: `svcWebhookBackend`)
- `backend_service_locales.go` (1 struct: `svcLocaleBackend`)
- `backend_service_publishing.go` (1 struct: `svcPublishingBackend`)
- `backend_service_validations.go` (1 struct: `svcValidationBackend`)
- `backend_service_infra.go` (2 structs: `svcImportBackend`, `svcTableBackend`)
- `backend_service_auth.go` (1 struct: `svcAuthBackend`)

#### Step 3.3: Update `NewServiceBackends`

**File:** `internal/mcp/backend_service.go`

Remove the `ac audited.AuditContext` parameter. Each struct is created without `ac`:

```go
// Before:
func NewServiceBackends(svc *service.Registry, ac audited.AuditContext) *Backends

// After:
func NewServiceBackends(svc *service.Registry) *Backends
```

#### Step 3.4: Update `ServeDirect`

**File:** `internal/mcp/serve.go`

`ServeDirect` (used for stdio mode) still needs to work without HTTP auth. For stdio mode, the audit context should default to "mcp-local" since stdio is a trusted local pipe. This function keeps its `ac` parameter for local identity but now also needs to pass it through context.

**Important type constraint:** `WithToolHandlerMiddleware` returns `ServerOption` (type `func(*MCPServer)`), not `StdioOption` (type `func(*StdioServer)`). These are different types. The tool handler middleware must be registered on the `MCPServer` via `newServer`, not passed to `server.ServeStdio`.

```go
func ServeDirect(svc *service.Registry, ac audited.AuditContext) error {
    backends := NewServiceBackends(svc)
    // injectAuditContextMiddleware is a ServerOption that registers a
    // ToolHandlerMiddleware on the MCPServer. It stores the provided
    // AuditContext under mcpAuditKey{} in the tool call's context so
    // that AuditContextFromMCP has a fallback when no HTTP middleware
    // chain is present.
    // ServeDirect uses ONLY injectAuditContextMiddleware.
    // Do NOT add PermissionMiddleware here. Stdio mode is a trusted
    // local pipe with no HTTP auth context. Permission enforcement is
    // HTTP-only (DirectHandler, Step 1.3).
    srv := newServer(backends, nil,
        server.WithToolHandlerMiddleware(injectAuditContextMiddleware(ac)),
    )
    return server.ServeStdio(srv)
}
```

The `injectAuditContextMiddleware` is a `server.ToolHandlerMiddleware` (not a `StdioOption`). It wraps the tool handler to inject the default `audited.AuditContext` into the tool call's `context.Context` under the `mcpAuditKey{}` key. `AuditContextFromMCP` checks for this key as a fallback when `middleware.AuthenticatedUser(ctx)` returns nil (stdio mode).

### Phase 4: Cleanup

#### Step 4.1: Remove `mcp.APIKeyAuth`

Delete `APIKeyAuth` from `internal/mcp/serve.go:105-119`. This function is no longer needed since authentication is handled by the existing `DefaultMiddlewareChain` and the per-tool `PermissionMiddleware`.

#### Step 4.2: Rename `MCP_API_Key` to `MCP_Proxy_Token`

**File:** `internal/config/config.go`

Rename the `MCP_API_Key` field to `MCP_Proxy_Token` with JSON tag `mcp_proxy_token`. Update `cmd/serve.go` to check `cfg.MCP_Enabled` only (not `cfg.MCP_API_Key != ""`).

Update `internal/config/validate.go` to rename the `mcp_api_key` getter case to `mcp_proxy_token`.

**Additional files that reference `MCP_API_Key` / `mcp_api_key` and must be updated:**

| File | Line | Usage | Action |
|------|------|-------|--------|
| `internal/mcp/connection.go` | 53-55 | Reads `cfg.MCP_API_Key` to authenticate proxy SDK client against remote server via `modula.WithAPIKey(cfg.MCP_API_Key)` | Rename to `cfg.MCP_Proxy_Token` |
| `internal/admin/handlers/settings.go` | 242 | Processes `mcp_api_key` from admin panel form submissions | Rename form field to `mcp_proxy_token` |
| `internal/admin/pages/settings_templ.go` | 1355 | Renders `mcp_api_key` password field in admin settings page | Update source `.templ` file: rename field to `mcp_proxy_token`, change label to "Proxy Token", update description to "API token for connecting to a remote CMS instance in proxy mode". Then run `just admin generate`. |
| `internal/config/field_meta.go` | 265 | Field metadata entry for config introspection | Rename metadata entry from `mcp_api_key` to `mcp_proxy_token` |
| `internal/config/HELP_TEXT.md` | 672 | Documentation for the config field | Rename and update description to clarify this is a client-side proxy credential, not a server authentication key |
| `cmd/mcp.go` | 26 | Help text references `mcp_api_key` | Rename to `mcp_proxy_token` |

**Config field rename:** `MCP_API_Key` is renamed to `MCP_Proxy_Token` (JSON: `mcp_proxy_token`). This field serves a single purpose after the refactor: authenticating the local MCP server's outbound SDK client against a remote CMS instance in proxy/registry mode (`connection.go:53`). It is NOT the server-side authentication mechanism (which is now API tokens resolved by the HTTP middleware chain). The old server-side use (`mcp.APIKeyAuth` comparing Bearer tokens against a static key) is deleted.

#### Step 4.3: Add init-time validation

In `permissions.go`, add an `init()` function that validates every permission label in the map using `middleware.ValidatePermissionLabel`. This catches typos at startup rather than at request time.

#### Step 4.4: Add test for permission map completeness

New test in `internal/mcp/permissions_test.go` that:

1. Collects all tool names from the MCP server (via `newServer` with mock backends)
2. Checks that every tool name either exists in `toolPermissions` or is in an explicit `publicTools` set
3. Fails if any tool has no mapping and is not public

This prevents new tools from being added without a permission mapping.

## File Changes Summary

| File | Action | Description |
|------|--------|-------------|
| `internal/mcp/auth.go` | **Create** | `mcpAuditKey` context key (stdio fallback only), `AuditContextFromMCP` helper |
| `internal/mcp/auth_middleware.go` | **Create** | `PassthroughHTTPContextFunc`, `PermissionMiddleware`, `injectAuditContextMiddleware` |
| `internal/mcp/permissions.go` | **Create** | `toolPermissions` map, `publicTools` set, `init()` validation |
| `internal/mcp/permissions_test.go` | **Create** | Completeness test: every registered tool has a permission or is public |
| `internal/mcp/serve.go` | **Modify** | Remove `APIKeyAuth`, update `newServer` to accept `...server.ServerOption`, update `DirectHandler` and `ServeDirect` signatures |
| `internal/mcp/backend_service.go` | **Modify** | Remove `ac` parameter from `NewServiceBackends` |
| `internal/mcp/backend_service_*.go` (11 files) | **Modify** | For every `svc*Backend` struct that has an `ac audited.AuditContext` field: remove the field and replace every `b.ac` reference with `AuditContextFromMCP(ctx)`. Structs without an `ac` field (e.g., `svcPluginBackend`, `svcConfigBackend`, `svcDeployBackend`, `svcHealthBackend`, `svcSearchBackend`, `svcActivityBackend`) are unchanged. See Step 3.2 for the specific struct names per file. |
| `cmd/serve.go` | **Modify** | Remove static audit context, remove `APIKeyAuth` wrapper, call `DirectHandler(svc)` |
| `internal/config/config.go` | **Modify** | Rename `MCP_API_Key` to `MCP_Proxy_Token` (JSON: `mcp_proxy_token`) |
| `internal/config/validate.go` | **Modify** | Rename `mcp_api_key` getter case to `mcp_proxy_token` |
| `internal/mcp/connection.go` | **Modify** | Rename `cfg.MCP_API_Key` to `cfg.MCP_Proxy_Token` |
| `internal/admin/handlers/settings.go` | **Modify** | Rename `mcp_api_key` form field to `mcp_proxy_token` |
| `internal/admin/pages/settings.templ` | **Modify** | Rename field to `mcp_proxy_token`, label to "Proxy Token", then run `just admin generate` |
| `internal/config/field_meta.go` | **Modify** | Rename `mcp_api_key` metadata entry to `mcp_proxy_token` |
| `internal/config/HELP_TEXT.md` | **Modify** | Rename and update description for `mcp_proxy_token` |
| `cmd/mcp.go` | **Modify** | Rename `mcp_api_key` to `mcp_proxy_token` in help text |

## Execution Order

1. Phases 1-3 (Steps 1.1 through 3.4): Implement as a single pass. These phases are co-dependent and will not compile independently (`DirectHandler` signature change in Phase 1 requires `NewServiceBackends` signature change in Phase 3). Run `just check` once after completing all of Step 3.4.
2. Phase 4 (Steps 4.1-4.4): Cleanup and tests. Run `just check` after Step 4.2.
3. Run `just test` after all phases are complete.

## Testing Strategy

1. **Unit test:** `permissions_test.go` validates every registered tool has a permission mapping
2. **Unit test:** `auth_middleware_test.go` tests the permission middleware with mock contexts (admin bypass, allowed, denied, unauthenticated)
3. **Existing tests:** `tools_test.go` and `error_handling_test.go` must still pass. These test tool handler behavior which is unchanged. The service backends will need their test helpers updated to provide context with user/permissions instead of static audit contexts.

## Risks

1. **Proxy mode (`NewProxyBackends`)** is unaffected. It delegates to the SDK client which authenticates via its own Bearer token against the remote server. The remote server enforces RBAC on its end. The permission middleware in direct mode is complementary, not redundant.

2. **Stdio mode (`ServeWithRegistry`, `ServeDirect`)** has no HTTP request to extract auth from. `ServeDirect` is only used for local/embedded scenarios (the MCP is trusted). The `injectAuditContextMiddleware` provides a default identity. `ServeWithRegistry` uses proxy backends which authenticate against the remote.

3. **Breaking change for `mcp_api_key` users:** Anyone currently using the static `mcp_api_key` config field must create an API token via the admin panel and use that instead. Since ModulaCMS is greenfield with no active users, this is acceptable.

4. **No rate limiting on MCP auth failures:** The `/mcp` endpoint is not covered by `AuthEndpointChain` rate limiting. An attacker can attempt token brute-force against `/mcp` without throttling. Mitigated by API tokens being hashed ULIDs (high entropy). If rate limiting becomes necessary, add `/mcp` to the rate-limited endpoint set in a follow-up.

5. **SSE session token revocation:** `StreamableHTTPServer` supports SSE for streaming responses. `WithHTTPContextFunc` runs on the initial POST. If a token is revoked mid-SSE-session, the session continues until the next HTTP request. For a greenfield system this is acceptable. If long-lived SSE sessions become common, add periodic revalidation in a follow-up.
