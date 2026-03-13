# Service Layer — Phase 0: Foundation

**Parent:** [SERVICE_LAYER_ROADMAP.md](SERVICE_LAYER_ROADMAP.md)
**Goal:** Create the `internal/service/` package skeleton, establish conventions, and wire injection into the existing startup path — without changing any handler behavior.

---

## Current State

Three injection patterns coexist:

| Consumer | Pattern | Gets DbDriver via |
|----------|---------|-------------------|
| Admin handlers | Closure factories: `func Foo(driver db.DbDriver, ...) http.HandlerFunc` | Direct parameter |
| API handlers | Bare functions: `func Foo(w, r, config.Config)` | `db.ConfigDB(c)` singleton lookup inside handler |
| MCP server | External binary using Go SDK over HTTP | N/A (calls REST API) |

Audit context is constructed inline in every mutating handler:
```go
ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, middleware.RequestIDFromContext(r.Context()), ip)
```

Context values available from middleware (all 4 chains: Default, Authenticated, AuthEndpoint, PublicAPI):
- **Request ID** — `middleware.RequestIDFromContext(ctx)` (position 1)
- **Client IP** — `middleware.ClientIPFromContext(ctx)` (position 2, resolves X-Forwarded-For → X-Real-IP → RemoteAddr)
- **User Agent** — `middleware.UserAgentFromContext(ctx)` (position 3)
- **User/Session** — `middleware.UserFromContext(ctx)` (session middleware, authenticated chains only)

There is one existing service-layer precedent: `internal/media/media_service.go` defines a `MediaStore` consumer interface. This validates the approach but isn't wired through a central registry.

---

## Deliverables

### 1. Package Structure

Create `internal/service/` as a flat package (one file per domain). No sub-packages — keeps imports simple and avoids circular dependency risk.

```
internal/service/
├── service.go          # Registry struct, constructor, shared types
├── errors.go           # Service-layer error types
├── audit.go            # AuditContext builder helper
├── schema.go           # SchemaService interface (placeholder, Phase 1 fills in)
├── content.go          # ContentService interface (placeholder, Phase 2 fills in)
├── media.go            # MediaService interface (placeholder)
├── routes.go           # RouteService interface (placeholder)
├── users.go            # UserService interface (placeholder)
├── rbac.go             # RBACService interface (placeholder)
├── plugins.go          # PluginService interface (placeholder)
├── webhooks.go         # WebhookService interface (placeholder)
├── locales.go          # LocaleService interface (placeholder)
├── sessions.go         # SessionService interface (placeholder)
├── tokens.go           # TokenService interface (placeholder)
├── ssh_keys.go         # SSHKeyService interface (placeholder)
├── oauth.go            # OAuthService interface (placeholder)
├── tables.go           # TableService interface (placeholder)
├── config.go           # ConfigService interface (placeholder)
├── import.go           # ImportService interface (placeholder)
├── deploy.go           # DeployService interface (placeholder)
├── audit_log.go        # AuditLogService interface (placeholder)
└── backup.go           # BackupService interface (placeholder)
```

### 2. Registry Struct (`service.go`)

A single struct that holds all service instances. Constructed at startup, passed into mux creation and admin route registration.

```go
type Registry struct {
    driver     db.DbDriver
    mgr        *config.Manager
    pc         *middleware.PermissionCache
    emailSvc   *email.Service
    dispatcher publishing.WebhookDispatcher

    // Service instances — added as phases are implemented.
    // Phase 1:
    // Schema *schema.Impl
    // Phase 2:
    // Content *content.Impl
    // etc.
}

func NewRegistry(
    driver db.DbDriver,
    mgr *config.Manager,
    pc *middleware.PermissionCache,
    emailSvc *email.Service,
    dispatcher publishing.WebhookDispatcher,
) *Registry
```

The Registry also exposes its raw dependencies via getters so handlers that haven't been migrated yet can still pull what they need during the incremental migration:

```go
func (r *Registry) Driver() db.DbDriver
func (r *Registry) Config() (*config.Config, error)
func (r *Registry) Manager() *config.Manager
func (r *Registry) PermissionCache() *middleware.PermissionCache
```

### 3. Service-Layer Error Types (`errors.go`)

Standardized errors that all services return. Handlers map these to HTTP status codes or templ error pages. Keeps HTTP semantics out of the service layer.

```go
type NotFoundError struct{ Resource, ID string }
type ValidationError struct {
    Errors []FieldError
}
type FieldError struct{ Field, Message string }
type ConflictError struct{ Resource, ID, Detail string }
type ForbiddenError struct{ Message string }
type InternalError struct{ Err error }
```

`ValidationError` holds a slice of `FieldError` to support batch validation from day one. Services that validate multiple fields (datatype creation, content creation, user creation) return all errors at once rather than failing on the first.

Constructor helpers:

```go
func NewValidationError(field, message string) *ValidationError          // single field convenience
func NewValidationErrors(errs ...FieldError) *ValidationError            // multiple fields
func (e *ValidationError) Add(field, message string) *ValidationError    // builder pattern
```

Plus type-check helpers:

```go
func IsNotFound(err error) bool      // errors.As check
func IsValidation(err error) bool
func IsConflict(err error) bool
func IsForbidden(err error) bool
```

### 4. Audit Helper (`audit.go`)

Extract the repeated `audited.Ctx(...)` construction into a shared helper. Every mutating handler currently does this identically:

```go
cfg, _ := mgr.Config()
user := middleware.UserFromContext(r.Context())
ip := r.RemoteAddr  // now: middleware.ClientIPFromContext(r.Context())
requestID := middleware.RequestIDFromContext(r.Context())
ac := audited.Ctx(types.NodeID(cfg.Node_ID), user.UserID, requestID, ip)
```

The service layer provides two constructors:

**HTTP-origin requests** (admin panel, API handlers):
```go
func (r *Registry) AuditCtx(ctx context.Context) (audited.AuditContext, error)
```
Reads user ID, request ID, and client IP from context (placed by `ClientIPMiddleware`, `RequestIDMiddleware`, and session middleware). Reads node ID from config. Eliminates ~4 lines of boilerplate from every mutating handler.

**Non-HTTP callers** (TUI, scheduler, CLI commands):
```go
func (r *Registry) SystemAuditCtx(userID types.UserID, reason string) (audited.AuditContext, error)
```
For callers where there is no HTTP request context. Uses a synthetic request ID (`reason` parameter, e.g. `"scheduled-publish"`, `"tui-edit"`, `"cli-import"`) and `"system"` as the IP. Reads node ID from config. Matches the existing pattern in `internal/router/scheduler.go`.

### 5. Placeholder Interfaces (one per domain file)

Each domain file contains an empty interface with a doc comment describing the scope. This establishes the file layout and makes later phases additive (append methods, add implementation) rather than creating new files.

Example `schema.go`:
```go
// SchemaService manages datatypes, fields, field types, and
// datatype-field associations. Implemented in Phase 1.
type SchemaService interface{}
```

No implementations yet — just the interface declarations. Each file has a package declaration and doc comment only (no empty `interface{}` declarations that would trigger lint warnings). The interface is added when the phase that implements it begins.

### 6. Wiring Changes

Modify the startup path to construct and pass `*service.Registry`:

**`cmd/serve.go`** (or wherever `NewModulacmsMux` is called):
- After `db.InitDB`, `config.NewManager`, `middleware.NewPermissionCache`, `email.NewService`, and `publishing.NewDispatcher` are created
- Construct `svc := service.NewRegistry(driver, mgr, pc, emailSvc, dispatcher)`
- Pass `svc` to `NewModulacmsMux` alongside existing parameters

**`internal/router/mux.go`**:
- Add `svc *service.Registry` parameter to `NewModulacmsMux` signature
- Add `svc *service.Registry` parameter to `registerAdminRoutes` signature
- No handler changes — the Registry is threaded through but unused until Phase 1

This is the only behavioral change in Phase 0, and it's purely additive (existing parameters stay, one new one appears).

---

## What Phase 0 Does NOT Do

- Does not change any handler behavior
- Does not implement any service methods
- Does not migrate any handlers to use services
- Does not touch the MCP server
- Does not modify DbDriver or any database code
- Does not add tests (no behavior to test yet; Phase 1 adds the first testable service)

---

## Files Changed

| File | Change |
|------|--------|
| `internal/service/service.go` | **New** — Registry struct + constructor |
| `internal/service/errors.go` | **New** — Error types + helpers |
| `internal/service/audit.go` | **New** — AuditCtx builder |
| `internal/service/{domain}.go` (x19) | **New** — Package declaration + doc comment (interface added when phase begins) |
| `cmd/serve.go` | **Modified** — Construct Registry, pass to mux |
| `internal/router/mux.go` | **Modified** — Accept `*service.Registry` param (2 function signatures) |
| `go.mod` / vendor | No changes (no new deps) |

**Total: 22 new files, 2 modified files. Zero behavioral changes.**

---

## Conventions Established

These conventions apply to all subsequent phases:

1. **Services accept `context.Context` as first parameter** for all methods that do I/O.
2. **Mutating methods accept `audited.AuditContext`** as second parameter. HTTP handlers use `Registry.AuditCtx(ctx)` to build it; non-HTTP callers (TUI, scheduler, CLI) use `Registry.SystemAuditCtx(userID, reason)`.
3. **Services return service-layer error types**, not HTTP status codes or raw DB errors. Wrap DB errors with context before returning.
4. **Services do not import `net/http`** — they have no knowledge of transport.
5. **Services use `db.DbDriver`** directly, not the Go SDK or config-based lookup.
6. **Handlers remain responsible for**: parsing request input, calling the service, formatting the response (JSON, templ, MCP result).
7. **Registry is the injection root** — services access other services through the Registry, not direct cross-imports.
8. **One file per domain** in `internal/service/`, not sub-packages.
9. **Consumer-defined interfaces** where a service needs only a subset of DbDriver (following the `MediaStore` precedent in `internal/media/media_service.go`). However, the Registry holds the full `DbDriver` — narrowing happens at the individual service level.
10. **Incremental migration** — handlers can be migrated one at a time. Unmigrated handlers continue to access `driver`/`config` through Registry getters.
11. **Transactions** — single DbDriver calls are inherently atomic (each wrapper method runs one SQL statement). Multi-step operations that must be atomic (content tree save, batch updates, import) will use a `WithTx` pattern: the service obtains a transactional DbDriver from the Registry and passes it to the sequence of calls. The exact `WithTx` API will be designed in Phase 2 (ContentService) when the first multi-step atomic operation is implemented. Until then, services that only make single DbDriver calls per method do not need transaction handling.
12. **Batch validation** — services validate all fields before performing any mutations and return `*ValidationError` with all `FieldError` entries collected. Do not fail-fast on the first invalid field.

---

## Validation Criteria

Phase 0 is complete when:

- [ ] `internal/service/` package exists with all files listed above
- [ ] `go build ./...` succeeds with no errors
- [ ] `just test` passes (no behavioral changes means no test regressions)
- [ ] `NewModulacmsMux` and `registerAdminRoutes` accept `*service.Registry`
- [ ] `cmd/serve.go` constructs the Registry and passes it through
- [ ] No handler code is changed beyond the function signatures accepting the new parameter
