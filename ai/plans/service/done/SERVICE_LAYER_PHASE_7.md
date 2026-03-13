# Phase 7: MCP Direct Integration + Auth Extraction + Straggler Cleanup

## Context

Phases 0-6 established the service layer and migrated all domain services. After Phase 6:
- **19 services** on the Registry: Schema, Content, AdminContent, Media, Routes, Users, RBAC, Plugins, Webhooks, Locales, Sessions, Tokens, SSHKeys, OAuth, Tables, ConfigSvc, Import, Deploy, AuditLog, Backup
- **~150 API handlers** call services via `*service.Registry`
- **~159 MCP tools** still go through the Go SDK over HTTP (`MCP tool → Go SDK → HTTP API → handler → service → DbDriver`)
- **6 straggler handlers** still take `config.Config` directly: auth.go (9 functions), globals.go, contentComposite.go, query.go, adminTree.go, health.go

**Goal:** After Phase 7:
1. MCP tools call services directly, eliminating the HTTP round-trip
2. All straggler handlers use `*service.Registry`
3. An AuthService encapsulates login, registration, OAuth flows, and password reset
4. Zero remaining `config.Config` dispatcher parameters in handler function signatures

---

## Current MCP Architecture

### How It Works Today

```
┌───────────┐     HTTP      ┌──────────┐    service    ┌─────────┐
│  MCP Tool │──────────────▶│ API      │─────────────▶│ Service │──▶ DB
│ (Go SDK)  │  localhost    │ Handler  │              │  Layer  │
└───────────┘               └──────────┘              └─────────┘
```

The MCP server (`internal/mcp/`) uses the Go SDK (`sdks/go/`) as an HTTP client:

```go
// serve.go — creates SDK client pointing at CMS HTTP server
func Serve(url, apiKey string) error {
    client, _ := modula.NewClient(modula.ClientConfig{BaseURL: url, APIKey: apiKey})
    return server.ServeStdio(newServer(client))
}
```

**Two deployment modes exist:**
1. **Stdio mode** (`modula mcp` CLI) — standalone process connecting to remote CMS over HTTP
2. **In-process mode** (`mux.go:562-570`) — MCP handler mounted at `/mcp` on the CMS HTTP server, but still creates an SDK client pointing at `http://localhost:<port>` — looping back through HTTP to itself

**159 MCP tools** follow a uniform pattern:
```go
func handleCreateUser(client *modula.Client) server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        username, _ := req.RequireString("username")
        // ... extract params
        result, err := client.Users.Create(ctx, modula.CreateUserParams{...})
        if err != nil { return errResult(err), nil }
        return jsonResult(result)
    }
}
```

### Problems with Current Architecture

1. **In-process HTTP loopback** — The most common deployment (embedded `/mcp` endpoint) creates an HTTP client that calls itself. Every MCP tool invocation serializes to JSON, makes an HTTP request to localhost, passes through the full middleware chain (CORS, rate limiting, auth, audit), deserializes, calls the service, serializes the response, sends it back over HTTP, and deserializes again.

2. **Double serialization** — SDK types are serialized to JSON for the HTTP request, then the response JSON is deserialized into SDK types, then re-serialized to JSON for the MCP response. Three JSON round-trips where one would suffice.

3. **Error fidelity loss** — Service errors (`NotFoundError`, `ValidationError`) are mapped to HTTP status codes by handlers, then the SDK maps them back to generic `ApiError` types. The MCP error handler (`errResult`) tries to reconstruct error details from the HTTP response body.

4. **Auth overhead** — The in-process MCP handler authenticates via API key on every request, even though it's running in the same process with full access to the service layer.

5. **Maintenance burden** — Every new service method requires: service method + handler + SDK method + MCP tool handler. Direct integration cuts this to: service method + MCP tool handler.

### After Phase 7

```
┌───────────┐    direct     ┌─────────┐
│  MCP Tool │─────────────▶│ Service │──▶ DB
│ (service) │              │  Layer  │
└───────────┘              └─────────┘
```

---

## Key Design Decisions

**Dual-mode MCP: direct + remote.** The MCP server must support both modes:
- **Direct mode** (in-process): accepts `*service.Registry`, tools call services directly. Used by the embedded `/mcp` endpoint and can be used by the stdio server when run as part of `modula serve`.
- **Remote mode** (stdio): accepts Go SDK client, tools call via HTTP. Used by `modula mcp --url <remote>` for connecting to a remote CMS. This mode is unchanged from today.

This means each tool handler needs a thin abstraction that dispatches to either the service or the SDK. The simplest approach: define a `Backend` interface per domain that both the service and SDK can satisfy.

**Backend interface, not adapter per tool.** Rather than 159 individual if/else branches, define ~18 domain-level interfaces (one per service) that abstract the common operations. Example:

```go
// internal/mcp/backend.go
type ContentBackend interface {
    List(ctx context.Context, limit, offset int64) (*PaginatedResponse[ContentData], error)
    Get(ctx context.Context, id string) (*ContentData, error)
    Create(ctx context.Context, params CreateContentParams) (*ContentData, error)
    // ...
}
```

Both the SDK client and a thin service adapter implement this interface. Tool handlers accept the interface. Registration becomes `registerContentTools(srv, contentBackend)`.

**Alternative considered: pass `*service.Registry` directly into tool handlers.** This is simpler (no interface layer) but means every tool handler must handle two different type systems (SDK types vs db types) and the remote mode would require a completely separate code path. The interface approach keeps tool handlers clean and type-safe.

**Alternative considered: generate adapter code.** The SDK and service method signatures are similar enough that code generation could produce the adapters. Worth investigating but not required — the 18 adapters are small and hand-written is fine for the initial implementation.

**JSON shapes are already compatible — no DTO layer needed.** All `types.Nullable*ID` types (NullableContentID, NullableRouteID, etc.) have custom `MarshalJSON()` methods that produce `"abc123"` or `null` — identical to the SDK's `*ContentID` pointer marshaling. `types.Timestamp` marshals to RFC3339 strings, same as the SDK. `db.NullString` marshals to `"value"` or `null`, same as `*string`. The reviewer's concern about `{"id":"...","valid":true}` was unfounded — the custom marshalers handle this.

The service adapter strategy is: `json.Marshal(dbResult)` for most types. The one exception is `db.Users` which includes a `Hash` field that the SDK's `User` type omits. The HTTP handler currently returns it but the SDK ignores it on deserialization. In direct mode, `json.Marshal(db.Users{})` would expose password hashes in the MCP JSON response. Fix: the service adapter for Users must zero the Hash field before marshaling (`user.Hash = ""`). A helper function `sanitizeUser(*db.Users)` handles this. No other types have sensitive fields that differ between db and SDK representations.

**MediaBackend.Upload accepts `io.Reader`, not multipart types.** The current MCP upload tool opens a local file (`os.Open(filePath)`) and passes the `*os.File` to the SDK's `Upload(ctx, io.Reader, filename, options)`. The service's `MediaService.Upload` takes `UploadMediaParams{File multipart.File, Header *multipart.FileHeader, ...}` — designed for HTTP multipart form data. The MediaBackend interface should accept `(ctx, reader io.Reader, filename string, path string)`, which both modes satisfy naturally:
- **SDK adapter:** passes `io.Reader` directly to `client.MediaUpload.Upload(ctx, reader, filename, nil)` — same as today.
- **Service adapter:** writes `io.Reader` to a temp file, then calls `MediaService.Upload` with a `multipart.File`-compatible wrapper, or better — refactor `UploadMediaParams` to accept `io.Reader` instead of `multipart.File` (the service only calls `io.Copy` and `file.Seek` on it; a temp file satisfies both). The MCP tool handler code (`os.Open` → pass reader) is unchanged.

**MCP response types use `json.RawMessage`.** Today, tools call `jsonResult(sdkType)` which marshals SDK types. With direct mode, the backend methods return `json.RawMessage` (pre-marshaled). The service adapter calls `json.Marshal(dbResult)` — which produces identical JSON to the SDK path thanks to the custom marshalers described above.

**AuthService is Phase 7A, MCP migration is Phase 7B.** Auth extraction is a prerequisite for MCP direct mode because MCP tools that create sessions or manage users need the auth logic in a service. Also, auth is the last straggler handler set — cleaning it up completes the `config.Config` elimination.

**Straggler handlers (globals, query, contentComposite, adminTree, health) are Phase 7C.** These are trivial — most only use `db.ConfigDB(c)` which becomes `svc.Driver()`. Grouped as a quick cleanup pass after the bigger work.

**Vestigial `config.Config` parameters also cleaned up in 7C.** 13 handlers from Phases 1-2 (adminDatatypes, adminFields, admin_field_types, datatypes, fields, field_types) take `c config.Config` as a parameter but never use it — the parameter is passed from the mux closure but the handler bodies only use `svc`. Phase 7C removes these dead parameters and updates their mux registrations. This is a mechanical change: remove `c config.Config` from the signature, remove `*c` from the mux call site.

**Fallback for MCP migration.** Any domain can remain on the SDK adapter indefinitely if the service adapter proves intractable. The system works correctly with a mix of SDK and service backends — this is the architectural benefit of the Backend interface pattern. The SDK adapter is the default; service adapters are opt-in per domain.

---

## Phase 7A: AuthService Extraction

### Scope

Extract auth logic from `internal/router/auth.go` (9 functions, ~650 lines) into `internal/service/auth.go`. This is the last handler file that takes `config.Config` as a primary parameter.

### AuthService Specification

**Dependencies:** `db.DbDriver`, `*config.Manager`, `*email.Service`

**Types:**
```go
type LoginInput struct {
    Username string
    Password string
}

type LoginResult struct {
    User    *db.Users
    Session *db.Sessions
}

type RegisterInput struct {
    Username string
    Name     string
    Email    string
    Password string
}

type PasswordResetRequestInput struct {
    Email string
}

type PasswordResetConfirmInput struct {
    Token       string
    NewPassword string
}

type OAuthProviderConfig struct {
    ClientID       string
    ClientSecret   string
    Scopes         string
    RedirectURL    string
    Endpoint       map[string]string
    ProviderName   string
    SuccessRedirect string
}

type OAuthCallbackInput struct {
    Code         string
    State        string
    StoredState  string   // from session
    CodeVerifier string   // PKCE
}

type OAuthCallbackResult struct {
    User    *db.Users
    Session *db.Sessions
    IsNew   bool
}
```

**Methods:**
```
// Credential authentication
Login(ctx, ac, input LoginInput) (*LoginResult, error)
Logout(ctx, ac, sessionID string) error
GetAuthenticatedUser(ctx, sessionID string) (*db.Users, error)

// Registration
Register(ctx, ac, input RegisterInput) (*db.Users, error)

// Password reset
RequestPasswordReset(ctx, input PasswordResetRequestInput) error  // sends email
ConfirmPasswordReset(ctx, ac, input PasswordResetConfirmInput) error

// OAuth
GetOAuthConfig(ctx) (*OAuthProviderConfig, error)  // reads config fields
HandleOAuthCallback(ctx, ac, input OAuthCallbackInput) (*OAuthCallbackResult, error)
```

**Business logic moved from handlers:**
- `Login`: validate credentials → verify password hash → create session → return user + session
- `Register`: validate input → hash password → lookup viewer role → create user with default role → audit
- `RequestPasswordReset`: find user by email → generate reset token → store token → send email with reset link
- `ConfirmPasswordReset`: validate token → hash new password → update user → delete used tokens
- `HandleOAuthCallback`: validate state → exchange code for token (PKCE) → fetch user info → find-or-create user + OAuth link → create session

**What stays in handlers:**
- Cookie management (`http.SetCookie`, cookie reading) — HTTP-specific, belongs in handler
- OAuth redirect URL construction — HTTP-specific
- PKCE state/verifier generation — can stay in handler or move to service (stateless crypto)
- OAuth initiate redirect — HTTP-specific (302 redirect)

### Handler Changes

After AuthService extraction, `auth.go` handlers become thin:

```go
func LoginHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
    // decode request → call svc.Auth.Login() → set cookie → write response
}

func RegisterHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
    // decode request → call svc.Auth.Register() → write response
}
```

All `config.Config` parameters are replaced with `svc *service.Registry`. The service reads config fields internally via `s.mgr.Config()`.

### Mux Changes

~8 registrations in mux.go change from `*c` / `driver` / `emailSvc` to `svc`. The OAuth closures change from `OauthCallbackHandler(*c)` to `OauthCallbackHandler(svc)`.

### Files Changed

| File | Change | Notes |
|------|--------|-------|
| `service/auth.go` | New | ~300 lines; auth business logic |
| `service/service.go` | Edit | Add `Auth *AuthService` field + constructor |
| `router/auth.go` | Rewrite | 9 functions: `config.Config` → `svc` |
| `router/mux.go` | Edit | ~8 registrations |

---

## Phase 7B: MCP Direct Integration

### Architecture

#### Backend Interface Pattern

Each domain gets a backend interface in `internal/mcp/backend.go`:

```go
package mcp

import "context"

// ContentBackend abstracts content operations for MCP tools.
type ContentBackend interface {
    ListContent(ctx context.Context, limit, offset int64) (json.RawMessage, error)
    GetContent(ctx context.Context, id string) (json.RawMessage, error)
    CreateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
    UpdateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
    DeleteContent(ctx context.Context, id string) error
    // ... remaining content methods
}

// SchemaBackend abstracts schema operations for MCP tools.
type SchemaBackend interface { ... }

// MediaBackend abstracts media operations for MCP tools.
type MediaBackend interface { ... }

// ... one per domain (18 total)
```

Key design choice: Backend methods accept and return `json.RawMessage` for complex objects. This avoids the MCP package importing either SDK types or db types — both adapters handle their own marshaling. For simple parameters (IDs, strings, numbers), use native Go types.

#### Two Adapter Implementations

**SDK adapter** (`internal/mcp/backend_sdk.go`):
```go
type sdkContentBackend struct {
    client *modula.Client
}

func (b *sdkContentBackend) ListContent(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
    result, err := b.client.ContentData.ListPaginated(ctx, modula.PaginationParams{Limit: limit, Offset: offset})
    if err != nil { return nil, err }
    return json.Marshal(result)
}
```

**Service adapter** (`internal/mcp/backend_service.go`):
```go
type serviceContentBackend struct {
    svc *service.Registry
}

func (b *serviceContentBackend) ListContent(ctx context.Context, limit, offset int64) (json.RawMessage, error) {
    result, err := b.svc.Content.ListContent(ctx, int32(limit), int32(offset))
    if err != nil { return nil, err }
    return json.Marshal(result)
}
```

#### Backends Struct

All backends collected in one struct:

```go
// Backends holds all domain backends for MCP tool registration.
type Backends struct {
    Content      ContentBackend
    AdminContent AdminContentBackend
    Schema       SchemaBackend
    AdminSchema  AdminSchemaBackend
    Media        MediaBackend
    Routes       RouteBackend
    AdminRoutes  AdminRouteBackend
    Users        UserBackend
    RBAC         RBACBackend
    Sessions     SessionBackend
    Tokens       TokenBackend
    SSHKeys      SSHKeyBackend
    OAuth        OAuthBackend
    Tables       TableBackend
    Plugins      PluginBackend
    Config       ConfigBackend
    Import       ImportBackend
    Deploy       DeployBackend
    Health       HealthBackend
}
```

#### Construction

```go
// NewSDKBackends creates backends that call the CMS via HTTP (remote mode).
func NewSDKBackends(client *modula.Client) *Backends { ... }

// NewServiceBackends creates backends that call services directly (in-process mode).
func NewServiceBackends(svc *service.Registry) *Backends { ... }
```

#### Updated Server Creation

```go
func newServer(backends *Backends) *server.MCPServer {
    srv := server.NewMCPServer("modula", utility.Version)
    registerContentTools(srv, backends.Content)
    registerSchemaTools(srv, backends.Schema)
    // ...
    return srv
}

func Serve(url, apiKey string) error {
    client, _ := modula.NewClient(modula.ClientConfig{BaseURL: url, APIKey: apiKey})
    return server.ServeStdio(newServer(NewSDKBackends(client)))
}

func ServeDirect(svc *service.Registry) error {
    return server.ServeStdio(newServer(NewServiceBackends(svc)))
}

func Handler(baseURL, apiKey string) (http.Handler, error) {
    // Remote mode (legacy)
    client, _ := modula.NewClient(modula.ClientConfig{BaseURL: baseURL, APIKey: apiKey})
    return newHTTPServer(NewSDKBackends(client)), nil
}

func DirectHandler(svc *service.Registry) http.Handler {
    // Direct mode — no HTTP round-trip
    return newHTTPServer(NewServiceBackends(svc))
}
```

#### Mux Integration

```go
// mux.go — replace HTTP loopback with direct handler
if c.MCP_Enabled && c.MCP_API_Key != "" {
    mcpHandler := mcpserver.DirectHandler(svc)
    mux.Handle("/mcp", mcpAPIKeyAuth(c.MCP_API_Key, mcpHandler))
}
```

### Tool Handler Migration

Each tool file changes from:
```go
func registerContentTools(srv *server.MCPServer, client *modula.Client) {
    srv.AddTool(mcp.NewTool("list_content", ...), handleListContent(client))
}

func handleListContent(client *modula.Client) server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        limit := int64(req.GetFloat("limit", 20))
        offset := int64(req.GetFloat("offset", 0))
        result, err := client.ContentData.ListPaginated(ctx, modula.PaginationParams{...})
        // ...
    }
}
```

To:
```go
func registerContentTools(srv *server.MCPServer, backend ContentBackend) {
    srv.AddTool(mcp.NewTool("list_content", ...), handleListContent(backend))
}

func handleListContent(backend ContentBackend) server.ToolHandlerFunc {
    return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        limit := int64(req.GetFloat("limit", 20))
        offset := int64(req.GetFloat("offset", 0))
        data, err := backend.ListContent(ctx, limit, offset)
        if err != nil { return errResult(err), nil }
        return rawJSONResult(data), nil
    }
}
```

The tool definition (parameter schema) is unchanged. Only the handler body changes from `client.Resource.Method()` to `backend.Method()`.

### Error Handling: `errResult` Update

The current `errResult` (helpers.go) only understands `*modula.ApiError` from the SDK. In direct mode, service adapters return service-layer error types (`*NotFoundError`, `*ValidationError`, `*ConflictError`, `*ForbiddenError`, `*InternalError`). The `errResult` function must be updated to handle both:

```go
func errResult(err error) *mcp.CallToolResult {
    // Service error types (direct mode)
    if service.IsNotFound(err) {
        return mcp.NewToolResultError(formatMCPError(404, err.Error()))
    }
    if service.IsValidation(err) {
        return mcp.NewToolResultError(formatMCPError(422, err.Error()))
    }
    if service.IsConflict(err) {
        return mcp.NewToolResultError(formatMCPError(409, err.Error()))
    }
    if service.IsForbidden(err) {
        return mcp.NewToolResultError(formatMCPError(403, err.Error()))
    }

    // SDK error types (remote mode)
    var apiErr *modula.ApiError
    if errors.As(err, &apiErr) {
        return mcp.NewToolResultError(formatMCPError(apiErr.StatusCode, apiErr.Message))
    }

    // Unknown errors
    return mcp.NewToolResultError(formatMCPError(500, err.Error()))
}

func formatMCPError(status int, message string) string {
    b, _ := json.Marshal(map[string]any{"status": status, "message": message})
    return string(b)
}
```

This produces identical JSON error format regardless of which adapter is in use. The `service.Is*` checks use `errors.As` under the hood, so wrapped errors are handled correctly. The import of `internal/service` is already required for the service adapter — no new dependency.

### Audit Context for MCP

Backend interfaces do NOT take `audited.AuditContext` as a parameter. Each adapter manages audit context internally:

- **SDK adapter:** No audit context needed — the HTTP handler on the receiving end creates it from the HTTP request (IP, user agent, authenticated user). This is unchanged from today.
- **Service adapter:** Creates a synthetic `AuditContext` for each mutating call. The audit context is constructed once at adapter creation time and reused:

```go
type serviceContentBackend struct {
    svc *service.Registry
    ac  audited.AuditContext  // set once at construction
}
```

`NewServiceBackends` resolves the API key to a user at construction time:

```go
func NewServiceBackends(svc *service.Registry, apiKey string) (*Backends, error) {
    // Resolve API key → token → user
    token, err := svc.Tokens.GetByKey(ctx, apiKey)
    if err != nil { return nil, fmt.Errorf("resolve MCP API key: %w", err) }
    user, err := svc.Users.GetUser(ctx, token.UserID)
    if err != nil { return nil, fmt.Errorf("resolve MCP user: %w", err) }

    ac := audited.AuditContext{
        UserID:    string(user.ID),
        Username:  user.Username,
        IP:        "127.0.0.1",
        UserAgent: "modula-mcp-direct",
    }

    return &Backends{
        Content: &serviceContentBackend{svc: svc, ac: ac},
        // ... all other backends share the same ac
    }, nil
}
```

If the API key is invalid or the user doesn't exist, `NewServiceBackends` returns an error and the MCP server refuses to start. This is fail-fast: a misconfigured API key is caught at startup, not on the first tool call.

For read-only methods (List, Get), the audit context is not used — but the service adapter still holds it for consistency. The service methods simply ignore the `ac` parameter for reads (they don't take one).

The Backend interface stays clean — only `context.Context` plus domain parameters:
```go
type ContentBackend interface {
    CreateContent(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
    // no audited.AuditContext in signature
}
```

This means the SDK adapter doesn't need to fabricate an unused `AuditContext`, and the service adapter handles it transparently.

### Implementation Order

Phase 7B has significant surface area (159 tools across 18 domains). Split into sub-steps:

**Step B0: Infrastructure**
- Define all 18 backend interfaces in `backend.go`
- Create `Backends` struct, `NewSDKBackends`, `NewServiceBackends` constructors
- Create `backend_sdk.go` and `backend_service.go` with all adapter implementations
- Update `serve.go` with `ServeDirect`, `DirectHandler`
- Update `mux.go` to use `DirectHandler`
- Verify: `just check` passes, in-process MCP still works (via SDK adapter initially)

**Step B1: Migrate tool files (batch 1 — thin CRUD)**
Migrate 7 tool files with simple CRUD patterns:
- `tools_sessions.go` (4 tools)
- `tools_tokens.go` (4 tools)
- `tools_ssh_keys.go` (3 tools)
- `tools_oauth.go` (4 tools → but see note below)
- `tools_tables.go` (5 tools)
- `tools_config.go` (3 tools)
- `tools_health.go` (1 tool)

Change from `client *modula.Client` to backend interface. ~24 tools total.

**Step B2: Migrate tool files (batch 2 — moderate)**
- `tools_users.go` (8 tools)
- `tools_rbac.go` (15 tools)
- `tools_routes.go` (5 tools)
- `tools_import.go` (2 tools)
- `tools_deploy.go` (4 tools)
- `tools_plugins.go` (13 tools)

~47 tools total.

**Step B3: Migrate tool files (batch 3 — complex)**
- `tools_content.go` (17 tools)
- `tools_schema.go` (21 tools)
- `tools_media.go` (12 tools)
- `tools_admin_content.go` (12 tools)
- `tools_admin_schema.go` (20 tools)
- `tools_admin_routes.go` (10 tools)

~92 tools total. Largest batch but most tools follow identical patterns.

**Step B4: Wire direct mode**
- Update `mux.go` to call `mcpserver.DirectHandler(svc)` instead of `mcpserver.Handler(baseURL, apiKey)`
- Update `cmd/mcp.go` to support a `--direct` flag (or auto-detect when run as part of `modula serve`)
- Remove Go SDK import from `internal/mcp/` (SDK adapter moves to `internal/mcp/sdk/` sub-package or stays but becomes optional)
- Verify: `just check`, `just test`

### Backend Interface Inventory

| Backend | Methods | Corresponding Service | Tools |
|---------|---------|----------------------|-------|
| `ContentBackend` | ~12 | `Content` + content field + tree + batch + heal | 17 |
| `AdminContentBackend` | ~10 | `AdminContent` | 12 |
| `SchemaBackend` | ~16 | `Schema` (datatypes + fields + field types + linking) | 21 |
| `AdminSchemaBackend` | ~15 | `Schema` (admin variants) | 20 |
| `MediaBackend` | ~8 | `Media` + dimensions + cleanup + health | 12 |
| `RouteBackend` | ~5 | `Routes` | 5 |
| `AdminRouteBackend` | ~5 | `Routes` (admin variants) | 10 |
| `UserBackend` | ~6 | `Users` | 8 |
| `RBACBackend` | ~10 | `RBAC` (roles + permissions + associations) | 15 |
| `SessionBackend` | ~4 | `Sessions` | 4 |
| `TokenBackend` | ~4 | `Tokens` | 4 |
| `SSHKeyBackend` | ~3 | `SSHKeys` | 3 |
| `OAuthBackend` | ~4 | `OAuth` | 5 |
| `TableBackend` | ~4 | `Tables` | 5 |
| `PluginBackend` | ~10 | `Plugins` | 13 |
| `ConfigBackend` | ~3 | `ConfigSvc` | 3 |
| `ImportBackend` | ~2 | `Import` | 2 |
| `DeployBackend` | ~4 | `Deploy` | 4 |
| `HealthBackend` | ~1 | (infrastructure) | 1 |

**Total: ~126 interface methods, 18 backends, 159 tools**

---

## Phase 7C: Straggler Handler Cleanup

### Scope

Migrate the 5 remaining non-auth handlers that still use `config.Config`:

| Handler | Functions | Config Usage | Difficulty |
|---------|-----------|-------------|------------|
| `globals.go` | 2 | `db.ConfigDB(c)` only | Trivial |
| `query.go` | 2 | `db.ConfigDB(c)` only | Trivial |
| `contentComposite.go` | 2 | `db.ConfigDB(c)` + audit | Trivial |
| `adminTree.go` | 2 | `db.ConfigDB(c)` + `Output_Format`, `Client_Site`, `Space_ID` | Easy |
| `health.go` | 3 | `db.ConfigDB(c)` + `Bucket_*` fields | Easy |

### Changes

**globals.go, query.go, contentComposite.go** — Replace `c config.Config` with `svc *service.Registry`. Replace `db.ConfigDB(c)` with `svc.Driver()`. Replace `middleware.AuditContextFromRequest(r, c)` with `cfg, _ := svc.Config(); ac := middleware.AuditContextFromRequest(r, *cfg)`. Update mux.go registrations.

**adminTree.go** — Same as above, plus replace `c.Output_Format`, `c.Client_Site`, `c.Space_ID` with reads from `cfg, _ := svc.Config()`.

**health.go** — Replace `c config.Config` with `svc *service.Registry`. Replace `db.ConfigDB(c)` with `svc.Driver()`. Replace `c.Bucket_*` reads with `cfg, _ := svc.Config()`. The `pluginHealthFn` parameter is unchanged.

### Files Changed

| File | Change | Lines |
|------|--------|-------|
| `router/globals.go` | Signature + bridge | ~10 |
| `router/query.go` | Signature + bridge | ~10 |
| `router/contentComposite.go` | Signature + bridge | ~15 |
| `router/adminTree.go` | Signature + bridge | ~20 |
| `router/health.go` | Signature + bridge | ~20 |
| `router/mux.go` | ~10 registrations | ~10 |

Estimated total: ~85 lines changed. Can be done in a single pass.

---

## Implementation Order

```
Phase 7A: AuthService Extraction          (prerequisite — last config.Config handler)
    │
    ├── Phase 7B: MCP Direct Integration  (largest piece — 159 tool migrations)
    │       Step B0: Infrastructure (interfaces, adapters, constructors)
    │       Step B1: Thin CRUD tools (24 tools)
    │       Step B2: Moderate tools (47 tools)
    │       Step B3: Complex tools (92 tools)
    │       Step B4: Wire direct mode
    │
    └── Phase 7C: Straggler Cleanup       (5 handlers, trivial)
    │
Step 8: Final Verification               (grep for config.Config, test all modes)
```

Phase 7A is prerequisite — AuthService must exist before MCP direct mode can handle user creation and session management.

Phase 7B and 7C are independent of each other and can run in parallel.

---

## Estimated Effort

| Sub-phase | Scope | Est. Sessions | Parallelizable |
|-----------|-------|:---:|:---:|
| 7A: AuthService | 1 service, ~9 handlers, ~650 lines | 1-2 | Yes (service + handlers) |
| 7B0: MCP infrastructure | Interfaces + adapters + constructors | 1 | No |
| 7B1: Thin CRUD tools | 24 tools across 7 files | 1 | Yes (per file) |
| 7B2: Moderate tools | 47 tools across 6 files | 1-2 | Yes (per file) |
| 7B3: Complex tools | 92 tools across 6 files | 1-2 | Yes (per file) |
| 7B4: Wire direct mode | mux.go + cmd/mcp.go | 0.5 | No |
| 7C: Stragglers | 5 handlers + mux.go | 0.5 | Yes (per handler) |

**Total: 5-8 sessions** (with parallelization on the lower end).

---

## Testing Strategy

**AuthService tests** (`service/auth_test.go`):
- Login with correct/incorrect credentials
- Registration with duplicate username/email
- Password reset token generation and consumption
- OAuth callback user provisioning (new user, existing user)

**MCP backend tests** (`internal/mcp/backend_service_test.go`):
- Verify service adapter produces correct JSON shape matching SDK contract
- Round-trip test: create via service adapter → get via service adapter → compare
- Error mapping: verify service errors produce correct MCP error format

**MCP integration tests** (update existing `internal/mcp/tools_test.go`):
- Existing tests use SDK backend — keep for regression
- Add parallel tests using service backend to verify identical behavior
- Eventually: test matrix running each tool through both backends

**Straggler handler tests** — existing handler tests continue to work.

**Compile check after each step:** `just check` + `just test`.

---

## Risk & Mitigations

**Backend interface explosion.** 18 interfaces with ~126 total methods is substantial. Mitigation: interfaces are generated from the tool registration code — each tool maps to exactly one interface method. Keep interfaces domain-scoped (not one giant interface).

**JSON shape mismatch between SDK and service types.** The SDK uses `modulacms.ContentData` with specific JSON tags. The db package uses `db.ContentData` with potentially different field names. Mitigation: the service adapter explicitly marshals db types to match the SDK JSON contract. Start with one domain (sessions — simplest), verify JSON shape equality, then proceed.

**Audit context in direct mode.** Without an HTTP request, there's no source for user ID, IP, or user agent. Mitigation: resolve the API key to a user at MCP server startup, create a synthetic audit context. If the API key is invalid or revoked, fail fast at startup.

**Import cycle risk.** `internal/mcp/` imports `internal/service/` (for service adapter). `internal/router/mux.go` imports `internal/mcp/` (for DirectHandler). `internal/service/` must NOT import `internal/mcp/`. This is a one-directional chain: `router → mcp → service → db`. Verify with `go vet`.

**Go SDK becomes optional dependency.** After Phase 7B, the MCP package imports the Go SDK only for remote mode. If we want to make it optional, the SDK adapter can move to a sub-package (`internal/mcp/sdk/`). Not required immediately — the SDK is a local module with zero external deps.

**OAuth config reading.** AuthService needs access to 7+ OAuth config fields. It reads them via `s.mgr.Config()` at each invocation (correct for hot-reload). No new config.Manager methods needed — existing `Config()` returns the full struct.

---

## Files Changed (Full Inventory)

### Phase 7A: AuthService
| File | Change | Est. Lines |
|------|--------|-----------|
| `service/auth.go` | New | ~300 |
| `service/service.go` | Edit | +5 |
| `router/auth.go` | Rewrite | ~650 (same LOC, different signatures) |
| `router/mux.go` | Edit | ~8 registrations |

### Phase 7B: MCP Direct Integration
| File | Change | Est. Lines |
|------|--------|-----------|
| `mcp/backend.go` | New | ~400 (18 interfaces) |
| `mcp/backend_sdk.go` | New | ~600 (SDK adapter implementations) |
| `mcp/backend_service.go` | New | ~600 (service adapter implementations) |
| `mcp/serve.go` | Rewrite | ~80 (add ServeDirect, DirectHandler) |
| `mcp/tools_content.go` | Rewrite | ~576 (client → backend) |
| `mcp/tools_schema.go` | Rewrite | ~500 |
| `mcp/tools_media.go` | Rewrite | ~400 |
| `mcp/tools_routes.go` | Rewrite | ~200 |
| `mcp/tools_users.go` | Rewrite | ~300 |
| `mcp/tools_rbac.go` | Rewrite | ~400 |
| `mcp/tools_sessions.go` | Rewrite | ~150 |
| `mcp/tools_tokens.go` | Rewrite | ~150 |
| `mcp/tools_ssh_keys.go` | Rewrite | ~100 |
| `mcp/tools_oauth.go` | Rewrite | ~200 |
| `mcp/tools_tables.go` | Rewrite | ~150 |
| `mcp/tools_plugins.go` | Rewrite | ~350 |
| `mcp/tools_config.go` | Rewrite | ~100 |
| `mcp/tools_import.go` | Rewrite | ~100 |
| `mcp/tools_deploy.go` | Rewrite | ~150 |
| `mcp/tools_health.go` | Rewrite | ~30 |
| `mcp/tools_admin_content.go` | Rewrite | ~500 |
| `mcp/tools_admin_schema.go` | Rewrite | ~500 |
| `mcp/tools_admin_routes.go` | Rewrite | ~300 |
| `mcp/helpers.go` | Edit | ~20 (add rawJSONResult) |
| `router/mux.go` | Edit | ~5 (DirectHandler) |
| `cmd/mcp.go` | Edit | ~15 (direct mode flag) |

### Phase 7C: Straggler Cleanup
| File | Change | Est. Lines |
|------|--------|-----------|
| `router/globals.go` | Edit | ~10 |
| `router/query.go` | Edit | ~10 |
| `router/contentComposite.go` | Edit | ~15 |
| `router/adminTree.go` | Edit | ~20 |
| `router/health.go` | Edit | ~20 |
| `router/mux.go` | Edit | ~10 |

---

## Success Criteria

1. `grep -r "config\.Config" internal/router/ | grep -v "_test.go"` returns zero results (no handler takes config.Config)
2. In-process MCP endpoint (`/mcp`) calls services directly — no HTTP loopback
3. Standalone `modula mcp --url <remote>` still works via SDK adapter
4. All existing MCP tools produce identical JSON responses in both modes
5. `just check` + `just test` pass
6. No circular imports (`go vet ./...` clean)
