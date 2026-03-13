# Phase 4: UserService + RBACService ✅ COMPLETE (2026-03-08)

## Context

Phase 0 established the internal/service/ package with a Registry struct, error types, and audit helpers. Phase 1 extracted SchemaService. Phase 2 extracted ContentService and AdminContentService. Phase 3 extracted MediaService and RouteService. Phase 4 extracts UserService and RBACService — two domains with shared concerns (role assignment touches both, permission cache refresh bridges them) but distinct enough to implement in parallel streams.

The goal: admin panel, API, and (eventually) MCP all call the same service methods. Business logic (password hashing, system-protected guards, permission cache refresh, email/username uniqueness, role assignment validation) lives in one place.

### Current State (as of Phase 3 completion)

**Registry struct** (`internal/service/service.go`):
- Fields: `driver`, `mgr`, `pc`, `emailSvc`, `dispatcher`
- Domain services: `Schema *SchemaService`, `Content *ContentService`, `AdminContent *AdminContentService`, `Media *MediaService`, `Routes *RouteService`
- Getter methods: `Driver()`, `Config()`, `Manager()`, `PermissionCache()`, `EmailService()`, `Dispatcher()` for unmigrated handlers

**Migrated API handlers** (Phase 2/3 — dispatcher functions taking `*service.Registry`):
- Content: `ContentDatasHandler(w, r, svc)`, `AdminContentDatasHandler(w, r, svc)`, etc.
- Media: `MediasHandler(w, r, svc)`, `MediaHandler(w, r, svc)`, etc.
- Routes: `RoutesHandler(w, r, svc)`, `RouteHandler(w, r, svc)`, etc.

**Not yet migrated** (still use `config.Config` dispatcher pattern):
- Users: `UsersHandler(w, r, c)`, `UserHandler(w, r, c)`, `UsersFullHandler(w, r, c)`, `UserFullHandler(w, r, c)`, `UserReassignDeleteHandler(w, r, c)`
- Roles: `RolesHandler(w, r, c, pc)`, `RoleHandler(w, r, c, pc)`
- Permissions: `PermissionsHandler(w, r, c, pc)`, `PermissionHandler(w, r, c, pc)`
- Role-Permissions: `RolePermissionsHandler(w, r, c, pc)`, `RolePermissionHandler(w, r, c, pc)`, `RolePermissionsByRoleHandler(w, r, c)`

**Not yet migrated admin handlers** (still use `(driver, mgr)` or `(driver, pc, mgr)` closure pattern):
- Users: `UsersListHandler(driver)`, `UserDetailHandler(driver)`, `UserCreateHandler(driver, mgr)`, `UserUpdateHandler(driver, mgr)`, `UserDeleteHandler(driver, mgr)`
- Roles: `RolesListHandler(driver, pc)`, `RoleDetailHandler(driver, pc)`, `RoleNewFormHandler(driver, pc)`, `RoleCreateHandler(driver, pc, mgr)`, `RoleUpdateHandler(driver, pc, mgr)`, `RoleDeleteHandler(driver, pc, mgr)`

**Placeholder stubs** (comment only):
- `internal/service/users.go`
- `internal/service/rbac.go`

**Error infrastructure** (fully built in Phase 2):
- `HandleServiceError(w, r, err)` in errors_http.go — maps service errors to JSON or HTMX toast responses

---

## Key Design Decisions

**UserService does not absorb authentication.** LoginHandler, RegisterHandler, ResetPasswordHandler, MeHandler, and all OAuth handlers remain unchanged. They have their own authentication flow (session creation, cookie management, PKCE) that is distinct from user CRUD. UserService handles user lifecycle (create, update, delete, list, get) — not login/logout. Phase 6 may revisit if auth handlers share enough logic to warrant extraction.

**UserService accepts `isAdmin bool` for role assignment gating.** The handler extracts `middleware.ContextIsAdmin(r.Context())` and passes it to the service. The service never reads the HTTP context directly. This keeps the admin-check as a handler concern while letting the service enforce the policy: "only admins can assign roles." The alternative — passing the full context and having the service call `ContextIsAdmin` — would couple the service to the middleware package's context keys.

**Password hashing stays in the service layer, not the handler.** Currently both admin and API handlers call `auth.HashPassword` directly. Moving it into UserService ensures consistent password policy enforcement (length validation, bcrypt cost) regardless of consumer. Services accept plaintext passwords in create/update params and return the hashed user (hash field redacted in the returned struct — same as current behavior where hash is omitted from JSON via `json:"-"`).

**RBACService triggers permission cache refresh synchronously, not in a goroutine.** Currently all 8 API handler trigger points spawn `go pc.Load(driver)`. This fire-and-forget pattern means the response is sent before the cache is updated — a subsequent request within milliseconds could still see stale permissions. The service method calls `pc.Load(s.driver)` synchronously before returning. The caller (handler) has already sent the response body via `json.NewEncoder(w).Encode(...)` before calling the service method's trailing cache refresh. Wait — that's not right. The service method does the mutation AND the cache refresh, then returns. The handler sends the response AFTER the service returns. So the cache is refreshed before the response. This is a behavior change: responses will be ~10-30ms slower (one DB query for cache rebuild) but permissions are immediately consistent. Accepted as the correct trade-off.

Actually, re-examining the current code: the handlers write the JSON response first (`json.NewEncoder(w).Encode(...)`, `w.WriteHeader(...)`) and THEN spawn the goroutine. The response is already flushed. The goroutine runs after. With the service pattern, the flow changes: handler calls service method → service does mutation + cache refresh → service returns → handler writes response. The cache refresh now happens BEFORE the response. This adds latency to the response. To preserve the current behavior (fast response, eventual consistency), the service method should return first, and the handler should spawn the refresh goroutine. But this defeats the purpose of centralizing the logic.

**Resolution:** RBACService methods that mutate roles/permissions/role_permissions call `s.pc.Load(s.driver)` synchronously as the last step. This ensures cache consistency. The ~10-30ms added latency for admin RBAC mutations is acceptable — these are infrequent operations. The handler does NOT spawn any goroutine — the service handles it.

**System-protected guards are service-layer concerns.** Currently duplicated across API handlers and admin handlers. The service enforces system-protected checks and returns `ForbiddenError`. This fixes admin handlers that currently lack some guards that the API has.

**Admin role update's bulk-sync pattern moves into RBACService.** The admin handler currently does: delete all role-permission links, re-create from form. This becomes `RBACService.SyncRolePermissions(ctx, ac, roleID, []types.PermissionID)` — the service deletes existing links and creates new ones atomically. The API's individual create/delete role-permission endpoints remain as separate service methods.

**Username uniqueness check uses the query builder, not O(n) list scan.** The admin handler currently calls `ListUsers()` and scans the entire list to check username uniqueness. No `GetUserByUsername` exists in DbDriver, and adding one requires sqlc queries across all three backends — out of scope for a service extraction phase. Instead, the service uses `db.QExists(ctx, conn, dialect, "users", map[string]any{"username": username})` which produces `SELECT 1 FROM "users" WHERE "username" = ? LIMIT 1` — a single indexed lookup, no full table scan, no password hashes in memory. The service obtains `conn` via `s.driver.GetConnection()` and `dialect` via `db.DialectFromString(string(cfg.Db_Driver))` from the config snapshot.

**Email uniqueness check uses `db.QExists` with compound WHERE.** On create: `db.QExists(ctx, conn, dialect, "users", map[string]any{"email": string(email)})`. On update (exclude self): `WHERE email = ? AND user_id != ?` via `db.QExists` with a `ColumnOp` for the exclusion. Returns a boolean — no full user record fetch needed. Same pattern as username uniqueness.

**db.AssembleUserFullView stays in the db package.** It's a view-assembly helper that composes multiple DbDriver calls into a UserFullView. The service calls it via `db.AssembleUserFullView(s.driver, userID)`. No reason to duplicate or move this logic — it's already well-structured with bounded sequential queries.

**UserReassignDelete becomes a service method.** Currently in `userComposite.go` with ~90 lines of handler-level orchestration (count, reassign, delete). This is pure business logic that belongs in the service. The handler becomes ~15 lines: parse JSON, validate IDs, call service, write response.

**MCP stays as-is.** It calls the Go SDK over HTTP. Rewiring to call services directly is Phase 7.

**Services are separate structs attached as exported fields on Registry** (matching ContentService/MediaService pattern from Phases 2-3).

**Service methods return db.* types directly.** No service-layer DTOs for Phase 4.

**API handlers adopt the Phase 2 dispatcher pattern:** `func FooHandler(w, r, svc *service.Registry)`. This eliminates the `db.ConfigDB(c)` singleton usage and the `*middleware.PermissionCache` parameter (services access pc internally via Registry).

**Admin handlers adopt closure-factory pattern with svc:** `func FooHandler(svc *service.Registry) http.HandlerFunc`. Replaces `(driver, mgr)` and `(driver, pc, mgr)` signatures.

**Audit context pattern** (established in Phase 2): handlers construct `audited.AuditContext` via `middleware.AuditContextFromRequest(r, *cfg)` where `cfg` comes from `svc.Config()`. Service methods receive `ac audited.AuditContext` as a parameter. **Admin handlers must switch to `AuditContextFromRequest`.** Admin handlers currently construct audit context manually via `audited.Ctx()` with inline `net.SplitHostPort(r.RemoteAddr)`. This is NOT equivalent to `AuditContextFromRequest`, which uses `ClientIPFromContext` (middleware-provided, handles X-Forwarded-For) and injects `HookRunner` (Phase 3 plugin hook support). The manual construction is missing HookRunner — a Phase 3 regression. All 6 admin handler call sites (3 in users.go, 3 in roles.go) switch to `middleware.AuditContextFromRequest(r, *cfg)`. This fixes the HookRunner gap and deletes ~6 lines of duplicated `net.SplitHostPort` per handler. Note: admin handler files still need the `audited` import after this change — `AuditContextFromRequest` returns `audited.AuditContext` and the service methods accept that type. Do not remove the `audited` import during cleanup.

**Pre-existing bug: `writeHTMXError` JSON injection.** The `errors_http.go:writeHTMXError` function string-concatenates the error message into a JSON HX-Trigger header without escaping. If a service error message contains a double quote (e.g., `user "foo" not found`), the header produces invalid JSON. Phase 4 routes more errors through this path. Fix during Phase 4 implementation: use `json.Marshal` for the message string before embedding it in the HX-Trigger template. This is a one-line fix in `writeHTMXError`.

---

## UserService — internal/service/users.go

### Struct

```go
type UserService struct {
    driver db.DbDriver
    mgr    *config.Manager
}

func NewUserService(driver db.DbDriver, mgr *config.Manager) *UserService
```

### Param Types

```go
type CreateUserInput struct {
    Username string
    Name     string
    Email    types.Email
    Password string          // plaintext — service hashes it; must be >= 8 chars
    Role     types.RoleID    // typed role ID; zero value = default to viewer
    IsAdmin  bool            // whether the caller is an admin (for role assignment gating)
}

type UpdateUserInput struct {
    UserID   types.UserID
    Username string
    Name     string
    Email    types.Email
    Password string          // plaintext; empty = keep existing hash; if set, must be >= 8 chars
    Role     types.RoleID    // typed role ID; zero value = keep existing role
    IsAdmin  bool            // whether the caller is an admin (for role change gating)
}

type ReassignDeleteInput struct {
    UserID     types.UserID
    ReassignTo types.UserID    // empty/zero = default to SystemUserID
}

type ReassignDeleteResult struct {
    DeletedUserID              types.UserID
    ReassignedTo               types.UserID
    ContentDataReassigned      int64
    DatatypesReassigned        int64
    AdminContentDataReassigned int64
}
```

### Methods

| Method | Logic Beyond DbDriver |
|--------|----------------------|
| CreateUser(ctx, ac, CreateUserInput) (*db.Users, error) | Validate required fields (username, name, email, password). Password must be >= 8 chars and <= 72 bytes (bcrypt limit). Check email uniqueness via `db.QExists` with `WHERE email = ?`. Check username uniqueness via `db.QExists` with `WHERE username = ?`. Role gating: if `input.Role` is non-zero and `!input.IsAdmin`, return ForbiddenError. Default to viewer role if zero (fetch via `GetRoleByLabel("viewer")`). Hash password via `auth.HashPassword`. Set timestamps to `time.Now().UTC()`. |
| UpdateUser(ctx, ac, UpdateUserInput) (*db.Users, error) | Fetch existing user by ID (NotFoundError if missing). Validate required fields (username, name, email). System user role protection: if `UserID == SystemUserID` and role differs from existing, return ForbiddenError. Role change gating: if role changed and `!input.IsAdmin`, return ForbiddenError. Password: if non-empty, validate length >= 8 and hash; if empty, keep existing hash. Email uniqueness: if changed, check via `db.QExists` with `WHERE email = ? AND user_id != ?` (exclude self). Username uniqueness: if changed, check via `db.QExists` with `WHERE username = ? AND user_id != ?` (exclude self). Preserve `DateCreated`, set `DateModified` to `time.Now().UTC()`. |
| DeleteUser(ctx, ac, UserID) error | System user protection: if `UserID == SystemUserID`, return ForbiddenError. Fetch user (NotFoundError if missing). Delete via driver. |
| GetUser(ctx, UserID) (*db.Users, error) | NotFoundError mapping. |
| GetUserByEmail(ctx, Email) (*db.Users, error) | NotFoundError mapping. |
| ListUsers(ctx) (*[]db.Users, error) | Passthrough. |
| ListUsersWithRoleLabel(ctx) (*[]db.UserWithRoleLabelRow, error) | Passthrough. |
| GetUserFull(ctx, UserID) (*db.UserFullView, error) | Delegates to `db.AssembleUserFullView(s.driver, userID)`. NotFoundError mapping. |
| ReassignDelete(ctx, ac, ReassignDeleteInput) (*ReassignDeleteResult, error) | Validate `UserID` (not zero, not SystemUserID). Resolve `ReassignTo` (default to SystemUserID if zero/empty). Validate `ReassignTo` (not same as UserID). Verify both users exist via `s.driver.GetUser()`. Obtain `conn` via `s.driver.GetConnection()` and `dialect` via `db.DialectFromString`. **Phase 1: counts + reassigns inside `db.WithTransactionResult[ReassignDeleteResult]`** — the transaction function receives `*sql.Tx` which satisfies `db.Executor`, enabling `db.QCount` and `db.QUpdate` inside the tx. Three `QCount` calls (`content_data`, `datatypes`, `admin_content_data` with `WHERE author_id = ?`) capture counts for the result. Three `QUpdate` calls (same tables, `SET author_id = ? WHERE author_id = ?`) reassign authorship. All six operations are atomic — if any fails, the tx rolls back and no reassignment is partial. These operations are un-audited (bulk ownership transfer, no per-row change events). **Phase 2: audited delete after tx commits** — the `ac` parameter passed to `ReassignDelete` is forwarded directly to `s.driver.DeleteUser(ctx, ac, input.UserID)`. No new AuditContext is constructed inside the service method. `DeleteUser` uses `d.Connection` (cannot participate in the `*sql.Tx` above) and records a `change_event` via the audited command pattern. **Failure mode:** if the tx commits but DeleteUser fails, the reassigns are committed and no content is orphaned — the user simply still exists and the operation can be retried. This is strictly safer than the current non-transactional approach where a failure mid-reassign can leave some content un-reassigned. Note: DbDriver reassign methods (`ReassignContentDataAuthor`, etc.) hardcode `mdb.New(d.Connection)` and cannot accept a `*sql.Tx` — this is why the service uses `QUpdate` directly instead of the driver methods. |

### Validation Fixes (currently missing or inconsistent)

1. **API CreateUser lacks input validation** — no checks for empty username, name, or email. Only password emptiness is checked. Service adds full field validation for all consumers.
2. **API CreateUser accepts client-provided timestamps** — `DateCreated` and `DateModified` come from the JSON body. Service ignores client timestamps and sets both to `time.Now().UTC()`. This prevents timestamp manipulation.
3. **API UpdateUser accepts client-provided DateCreated** — service preserves existing `DateCreated` from the DB record.
4. **Username uniqueness** — admin handler checks it via O(n) `ListUsers` scan, API handler does not check at all. Service always checks via `db.QExists` (single indexed lookup).
5. **Password minimum length** — admin enforces 8 chars, API only checks non-empty. Service enforces **8 chars minimum** on both create and update (when provided), matching `internal/install/validation.go:ValidatePassword()` and admin handlers. Upper bound is 72 bytes (bcrypt limit, enforced by `auth.HashPassword`). Bcrypt cost is 12.
6. **Role default** — admin defaults to `"viewer"` string literal, API fetches viewer role by label. Service always fetches by label for correctness (role ID, not label, is stored).

### Private Helpers

- `validateCreateUserInput(input CreateUserInput) *ValidationError` — shared validation for required fields
- `validateUpdateUserInput(input UpdateUserInput, existing *db.Users) *ValidationError` — validate + diff-check
- `checkEmailUniqueness(ctx context.Context, email types.Email, excludeUserID types.UserID) error` — uses `db.QExists` against "users" table. On create (excludeUserID zero): `WHERE email = ?`. On update: `db.QExists(ctx, conn, dialect, "users", map[string]any{"email": string(email), "user_id": db.Neq(excludeUserID)})` — compound map with `db.Neq` ColumnOp produces `WHERE email = ? AND user_id != ?`. Returns ConflictError if exists.
- `checkUsernameUniqueness(ctx context.Context, username string, excludeUserID types.UserID) error` — uses `db.QExists` against "users" table. On create (excludeUserID zero): `WHERE username = ?`. On update: `db.QExists(ctx, conn, dialect, "users", map[string]any{"username": username, "user_id": db.Neq(excludeUserID)})` — same compound map pattern as email. Returns ConflictError if exists.
- `resolveDefaultRole(roleInput types.RoleID) (types.RoleID, error)` — if zero value (empty string), fetch viewer role ID via `GetRoleByLabel("viewer")`. If non-zero, returns the input unchanged. Only fires when the caller omits the role — NOT when they explicitly pass the viewer role ID.

---

## RBACService — internal/service/rbac.go

### Struct

```go
type RBACService struct {
    driver db.DbDriver
    mgr    *config.Manager
    pc     *middleware.PermissionCache
}

func NewRBACService(driver db.DbDriver, mgr *config.Manager, pc *middleware.PermissionCache) *RBACService
```

### Param Types

```go
type CreateRoleInput struct {
    Label string
}

type UpdateRoleInput struct {
    RoleID types.RoleID
    Label  string
}

type CreatePermissionInput struct {
    Label       string
    Description string
}

type UpdatePermissionInput struct {
    PermissionID types.PermissionID
    Label        string
    Description  string
}
```

### Methods — Roles

| Method | Logic Beyond DbDriver |
|--------|----------------------|
| CreateRole(ctx, ac, CreateRoleInput) (*db.Roles, error) | Validate label non-empty. Create with `SystemProtected: false`. Refresh permission cache synchronously via `s.pc.Load(s.driver)`. |
| UpdateRole(ctx, ac, UpdateRoleInput) (*db.Roles, error) | Fetch existing by ID (NotFoundError). System-protected label mutation guard: if `existing.SystemProtected && input.Label != existing.Label`, return ForbiddenError "cannot rename system-protected role". Update via driver. Re-fetch and return updated role. Refresh permission cache. |
| DeleteRole(ctx, ac, RoleID) error | Fetch existing by ID (NotFoundError). System-protected guard: if `existing.SystemProtected`, return ForbiddenError "cannot delete system-protected role". Check for users assigned to this role: use `db.QCount(ctx, conn, dialect, "users", map[string]any{"role": string(roleID)})`. If count > 0, return ConflictError "role has N user(s) assigned" — this prevents the cross-backend inconsistency where SQLite/PostgreSQL silently NULL the user's role (violating the NOT NULL intent) while MySQL rejects the delete via ON DELETE RESTRICT. Delete role-permission links via `s.driver.DeleteRolePermissionsByRoleID(ctx, ac, roleID)` first (prevents orphaned junction rows — `ac` is forwarded from the method parameter). Delete role via driver. Refresh permission cache. |
| GetRole(ctx, RoleID) (*db.Roles, error) | NotFoundError mapping. |
| GetRoleByLabel(ctx, string) (*db.Roles, error) | NotFoundError mapping. |
| ListRoles(ctx) (*[]db.Roles, error) | Passthrough. |

### Methods — Permissions

| Method | Logic Beyond DbDriver |
|--------|----------------------|
| CreatePermission(ctx, ac, CreatePermissionInput) (*db.Permissions, error) | Validate label format via `middleware.ValidatePermissionLabel`. Create via driver. Refresh permission cache. |
| UpdatePermission(ctx, ac, UpdatePermissionInput) (*db.Permissions, error) | Fetch existing by ID (NotFoundError). System-protected label mutation guard. Validate label format. Update via driver. Re-fetch and return updated. Refresh permission cache. |
| DeletePermission(ctx, ac, PermissionID) error | Fetch existing by ID (NotFoundError). System-protected guard. Delete via driver. Refresh permission cache. |
| GetPermission(ctx, PermissionID) (*db.Permissions, error) | NotFoundError mapping. |
| ListPermissions(ctx) (*[]db.Permissions, error) | Passthrough. |

### Methods — Role-Permissions

| Method | Logic Beyond DbDriver |
|--------|----------------------|
| CreateRolePermission(ctx, ac, db.CreateRolePermissionParams) (*db.RolePermissions, error) | Create via driver. Refresh permission cache. |
| DeleteRolePermission(ctx, ac, RolePermissionID) error | Fetch link (NotFoundError). Fetch role (NotFoundError). System-protected junction guard: if role is system-protected, return ForbiddenError "cannot modify permissions on system-protected role". Delete via driver. Refresh permission cache. |
| GetRolePermission(ctx, RolePermissionID) (*db.RolePermissions, error) | NotFoundError mapping. |
| ListRolePermissions(ctx) (*[]db.RolePermissions, error) | Passthrough. |
| ListRolePermissionsByRoleID(ctx, RoleID) (*[]db.RolePermissions, error) | Validate RoleID. Passthrough. |
| SyncRolePermissions(ctx, ac, RoleID, []types.PermissionID) ([]types.PermissionID, error) | Wrap the entire delete+create sequence in `db.WithTransactionResult[[]types.PermissionID]`. Inside the transaction: delete all existing links for role via `s.driver.DeleteRolePermissionsByRoleID(ctx, ac, roleID)`, then create new links from the provided slice via `s.driver.CreateRolePermission(ctx, ac, params)` for each. **Error handling inside the tx loop:** if an individual `CreateRolePermission` fails with a not-found or constraint violation (the permission ID does not exist), collect the PermissionID into the failed slice and continue. For any other error (connection failure, context cancellation), abort the loop immediately and return the error — the tx rolls back and no links are modified. If the loop completes, the tx commits. The failed PermissionIDs (if any) are returned as the first return value; nil if all succeeded. The caller (admin handler) can use the failed list to display which permissions were not assigned. Refresh permission cache once at the end, outside the transaction. |

### Permission Label Validation

The service delegates to `middleware.ValidatePermissionLabel(label)` which validates the `resource:operation` format character-by-character (no regex). This function already exists and is well-tested. The service wraps validation failure as `ValidationError{Field: "label", Message: "invalid permission label format"}`.

### Permission Cache Refresh

Every mutating method on RBACService calls `s.refreshCache()` as its final step before returning. This is 8 call sites consolidated from 8 separate `go pc.Load(db.ConfigDB(c))` goroutine spawns scattered across 3 handler files (`roles.go` has 3, `permissions.go` has 3, `role_permissions.go` has 2). Note: `s.driver` is `db.DbDriver` which embeds `db.RBACRepository` — this satisfies `pc.Load`'s parameter type.

Error from `pc.Load` is logged but does not fail the operation — the mutation itself succeeded and the periodic background refresh (60s interval) will pick up the change eventually.

### Private Helpers

- `refreshCache()` — calls `s.pc.Load(s.driver)`. If `Load` returns an error, logs it via `utility.DefaultLogger.Error` and discards it. Returns nothing. The error is intentionally swallowed: the mutation succeeded and the 60-second periodic refresh (`StartPeriodicRefresh`) will catch up. No return value — callers do not need to handle or propagate cache refresh failures.

---

## Handler Rewiring

### Admin Handlers — Users

`internal/admin/handlers/users.go` — all signatures change from `(driver db.DbDriver)` / `(driver db.DbDriver, mgr *config.Manager)` to `(svc *service.Registry)`:

| Handler | Before | After |
|---------|--------|-------|
| UsersListHandler | `(driver db.DbDriver)` | `(svc *service.Registry)` — calls `svc.Users.ListUsersWithRoleLabel()` and `svc.RBAC.ListRoles()` |
| UserDetailHandler | `(driver db.DbDriver)` | `(svc *service.Registry)` — calls `svc.Users.GetUser()` and `svc.RBAC.ListRoles()` |
| UserCreateHandler | `(driver db.DbDriver, mgr *config.Manager)` | `(svc *service.Registry)` — parse form, build `CreateUserInput`, call `svc.Users.CreateUser()`. Admin callers always set `IsAdmin: middleware.ContextIsAdmin(r.Context())`. Removes manual hash call, uniqueness checks, role defaulting — all in service now. |
| UserUpdateHandler | `(driver db.DbDriver, mgr *config.Manager)` | `(svc *service.Registry)` — parse form, build `UpdateUserInput`, call `svc.Users.UpdateUser()`. Removes manual hash call, email uniqueness check — all in service now. |
| UserDeleteHandler | `(driver db.DbDriver, mgr *config.Manager)` | `(svc *service.Registry)` — calls `svc.Users.DeleteUser()`. System-user guard now in service. |

Service errors map to HTMX responses via `HandleServiceError`:
- IsValidation → re-render form partial with field errors
- IsForbidden → toast "cannot modify system user" or "only administrators can assign roles"
- IsNotFound → toast "user not found"
- IsConflict → toast "email already exists" or "username already exists"

**Note on admin form error rendering:** The admin handlers currently re-render form partials with inline error maps (`map[string]string`). With service errors, the handler must translate `*ValidationError` field errors back into this map format for the templ partials. A helper `validationErrorToMap(err error) map[string]string` in the handlers package converts `FieldError` entries: iterate `ve.Errors`, set `result[fe.Field] = fe.Message` for each entry. If multiple errors exist for the same field, the last one wins. This matches the existing admin form error map pattern where each field key maps to a single error string. If `err` is not a `*ValidationError`, return nil.

### Admin Handlers — Roles

`internal/admin/handlers/roles.go` — all signatures change from `(driver db.DbDriver, pc *middleware.PermissionCache)` / `(driver db.DbDriver, pc *middleware.PermissionCache, mgr *config.Manager)` to `(svc *service.Registry)`:

| Handler | Before | After |
|---------|--------|-------|
| RolesListHandler | `(driver, pc)` | `(svc)` — calls `svc.RBAC.ListRoles()`, `svc.RBAC.ListPermissions()`, `svc.RBAC.ListRolePermissions()` |
| RoleDetailHandler | `(driver, pc)` | `(svc)` — calls `svc.RBAC.GetRole()`, `svc.RBAC.ListPermissions()`, `svc.RBAC.ListRolePermissions()` |
| RoleNewFormHandler | `(driver, pc)` | `(svc)` — calls `svc.RBAC.ListPermissions()` |
| RoleCreateHandler | `(driver, pc, mgr)` | `(svc)` — calls `svc.RBAC.CreateRole()`. Cache refresh now handled by service. |
| RoleUpdateHandler | `(driver, pc, mgr)` | `(svc)` — calls `svc.RBAC.UpdateRole()` then `svc.RBAC.SyncRolePermissions()`. Cache refresh now handled by service. |
| RoleDeleteHandler | `(driver, pc, mgr)` | `(svc)` — calls `svc.RBAC.DeleteRole()`. Pre-delete permission link cleanup handled by service (add to DeleteRole or call SyncRolePermissions with empty slice first). Cache refresh handled by service. |

**RoleDeleteHandler cleanup detail:** Currently the handler calls `DeleteRolePermissionsByRoleID` before `DeleteRole`. The service's `DeleteRole` should incorporate this cleanup: delete all role-permission links for the role, then delete the role itself. This avoids orphaned junction rows. The service already has `SyncRolePermissions` which deletes by role ID — but calling sync with an empty slice and then delete is two operations. Better: `DeleteRole` internally calls `driver.DeleteRolePermissionsByRoleID` first.

**buildRolePermMap stays as a handler-level helper.** It transforms `[]db.RolePermissions` into a `map[RoleID]map[PermissionID]bool` for the templ permission matrix UI. This is a view concern, not business logic.

### API Handlers — Users

`internal/router/users.go` — change from `(w, r, c config.Config)` to `(w, r, svc *service.Registry)`:

| Handler | Key Change |
|---------|-----------|
| UsersHandler(w, r, svc) | Signature change. Dispatch to service methods. |
| UserHandler(w, r, svc) | Signature change. |
| UsersFullHandler(w, r, svc) | Signature change. Calls `svc.Users.ListUsersWithRoleLabel()`. |
| UserFullHandler(w, r, svc) | Signature change. Calls `svc.Users.GetUserFull()`. |
| ApiCreateUser | Remove ~50 lines (hash, role lookup, role gating). Build `CreateUserInput{..., IsAdmin: middleware.ContextIsAdmin(r.Context())}`, call `svc.Users.CreateUser()`. Map service errors via `HandleServiceError`. |
| ApiUpdateUser | Remove ~60 lines (existing fetch, hash, role gating, system user check). Build `UpdateUserInput`, call `svc.Users.UpdateUser()`. |
| ApiDeleteUser | Remove system user guard. Call `svc.Users.DeleteUser()`. |
| ApiGetUser | Call `svc.Users.GetUser()`. |
| ApiListUsers | Email param branch: `svc.Users.GetUserByEmail()`. Default: `svc.Users.ListUsers()`. |
| apiListUsersWithRoleLabel | Call `svc.Users.ListUsersWithRoleLabel()`. |
| apiGetUserFull | Call `svc.Users.GetUserFull()`. |

`internal/router/userComposite.go` — change from `(w, r, c config.Config)` to `(w, r, svc *service.Registry)`:

| Handler | Key Change |
|---------|-----------|
| UserReassignDeleteHandler | Remove ~90 lines of orchestration. Parse JSON body into existing `UserReassignDeleteRequest` struct (stays in `userComposite.go` — transport-layer concern with json tags). Map request fields to `ReassignDeleteInput` (service-layer type, no json tags, in `users.go`). Call `svc.Users.ReassignDelete()`. Map `ReassignDeleteResult` fields to existing `UserReassignDeleteResponse` struct (stays in `userComposite.go` — transport-layer concern with json tags). Write response as JSON. The two-type layering (request/response in router, input/result in service) keeps the service package decoupled from HTTP serialization. |

### API Handlers — Roles

`internal/router/roles.go` — change from `(w, r, c config.Config, pc *middleware.PermissionCache)` to `(w, r, svc *service.Registry)`:

| Handler | Key Change |
|---------|-----------|
| RolesHandler(w, r, svc) | Signature drops `pc` — service has it internally. |
| RoleHandler(w, r, svc) | Signature drops `pc`. |
| apiCreateRole | Remove goroutine cache refresh. Call `svc.RBAC.CreateRole()`. |
| apiUpdateRole | Remove system-protected check + goroutine refresh. Call `svc.RBAC.UpdateRole()`. |
| apiDeleteRole | Remove system-protected check + goroutine refresh. Call `svc.RBAC.DeleteRole()`. |
| apiGetRole | Call `svc.RBAC.GetRole()`. |
| apiListRoles | Call `svc.RBAC.ListRoles()`. |

### API Handlers — Permissions

`internal/router/permissions.go` — change from `(w, r, c config.Config, pc *middleware.PermissionCache)` to `(w, r, svc *service.Registry)`:

| Handler | Key Change |
|---------|-----------|
| PermissionsHandler(w, r, svc) | Signature drops `pc`. |
| PermissionHandler(w, r, svc) | Signature drops `pc`. |
| apiCreatePermission | Remove label validation + goroutine refresh. Call `svc.RBAC.CreatePermission()`. |
| apiUpdatePermission | Remove system-protected check + label validation + goroutine refresh. Call `svc.RBAC.UpdatePermission()`. |
| apiDeletePermission | Remove system-protected check + goroutine refresh. Call `svc.RBAC.DeletePermission()`. |
| apiGetPermission | Call `svc.RBAC.GetPermission()`. |
| apiListPermissions | Call `svc.RBAC.ListPermissions()`. |

### API Handlers — Role-Permissions

`internal/router/role_permissions.go` — change from `(w, r, c config.Config, pc *middleware.PermissionCache)` to `(w, r, svc *service.Registry)`:

| Handler | Key Change |
|---------|-----------|
| RolePermissionsHandler(w, r, svc) | Signature drops `pc`. |
| RolePermissionHandler(w, r, svc) | Signature drops `pc`. |
| RolePermissionsByRoleHandler(w, r, svc) | Signature drops `c`, gains `svc`. |
| apiCreateRolePermission | Remove goroutine refresh. Call `svc.RBAC.CreateRolePermission()`. |
| apiDeleteRolePermission | Remove system-protected junction guard + goroutine refresh. Call `svc.RBAC.DeleteRolePermission()`. |
| apiGetRolePermission | Call `svc.RBAC.GetRolePermission()`. |
| apiListRolePermissions | Call `svc.RBAC.ListRolePermissions()`. |
| apiListRolePermissionsByRole | Call `svc.RBAC.ListRolePermissionsByRoleID()`. |

### Mux Wiring (internal/router/mux.go)

Update handler registrations to pass `svc` instead of `*c` / `(driver, pc, mgr)` / `(driver, mgr)`:

**API routes** (currently pass `*c` and/or `pc`, change to `svc`):
```go
// Before
UsersHandler(w, r, *c)
UserHandler(w, r, *c)
UsersFullHandler(w, r, *c)
UserFullHandler(w, r, *c)
UserReassignDeleteHandler(w, r, *c)
RolesHandler(w, r, *c, pc)
RoleHandler(w, r, *c, pc)
PermissionsHandler(w, r, *c, pc)
PermissionHandler(w, r, *c, pc)
RolePermissionsHandler(w, r, *c, pc)
RolePermissionHandler(w, r, *c, pc)
RolePermissionsByRoleHandler(w, r, *c)

// After
UsersHandler(w, r, svc)
UserHandler(w, r, svc)
UsersFullHandler(w, r, svc)
UserFullHandler(w, r, svc)
UserReassignDeleteHandler(w, r, svc)
RolesHandler(w, r, svc)
RoleHandler(w, r, svc)
PermissionsHandler(w, r, svc)
PermissionHandler(w, r, svc)
RolePermissionsHandler(w, r, svc)
RolePermissionHandler(w, r, svc)
RolePermissionsByRoleHandler(w, r, svc)
```

**Admin routes** (currently pass `driver` / `(driver, pc)` / `(driver, mgr)` / `(driver, pc, mgr)`, change to `svc`):
```go
// Before
adminhandlers.UsersListHandler(driver)
adminhandlers.UserDetailHandler(driver)
adminhandlers.UserCreateHandler(driver, mgr)
adminhandlers.UserUpdateHandler(driver, mgr)
adminhandlers.UserDeleteHandler(driver, mgr)
adminhandlers.RolesListHandler(driver, pc)
adminhandlers.RoleDetailHandler(driver, pc)
adminhandlers.RoleNewFormHandler(driver, pc)
adminhandlers.RoleCreateHandler(driver, pc, mgr)
adminhandlers.RoleUpdateHandler(driver, pc, mgr)
adminhandlers.RoleDeleteHandler(driver, pc, mgr)

// After
adminhandlers.UsersListHandler(svc)
adminhandlers.UserDetailHandler(svc)
adminhandlers.UserCreateHandler(svc)
adminhandlers.UserUpdateHandler(svc)
adminhandlers.UserDeleteHandler(svc)
adminhandlers.RolesListHandler(svc)
adminhandlers.RoleDetailHandler(svc)
adminhandlers.RoleNewFormHandler(svc)
adminhandlers.RoleCreateHandler(svc)
adminhandlers.RoleUpdateHandler(svc)
adminhandlers.RoleDeleteHandler(svc)
```

---

## Registry Changes

Add Users and RBAC fields to Registry struct and initialize in NewRegistry:

```go
type Registry struct {
    // ... existing fields ...

    Schema       *SchemaService
    Content      *ContentService
    AdminContent *AdminContentService
    Media        *MediaService
    Routes       *RouteService
    Users        *UserService         // Phase 4
    RBAC         *RBACService         // Phase 4
}

func NewRegistry(...) *Registry {
    reg := &Registry{...}
    // ... existing initializations ...
    reg.Users = NewUserService(driver, mgr)           // Phase 4
    reg.RBAC = NewRBACService(driver, mgr, pc)        // Phase 4
    return reg
}
```

---

## Files Changed

| File | Type | Scope |
|------|------|-------|
| internal/service/service.go | Moderate | Add Users + RBAC fields to Registry, initialize in NewRegistry |
| internal/service/users.go | Rewrite | Replace 2-line stub with full UserService struct (~350 lines) |
| internal/service/rbac.go | Rewrite | Replace 3-line stub with full RBACService struct (~400 lines) |
| internal/admin/handlers/users.go | Major | Change all 5 signatures from (driver, mgr) to (svc), rewrite create/update/delete to use service |
| internal/admin/handlers/roles.go | Major | Change all 7 signatures from (driver, pc, mgr) to (svc), rewrite create/update/delete to use service |
| internal/router/users.go | Major | Change 11 functions from (w, r, c) to (w, r, svc), remove business logic |
| internal/router/userComposite.go | Major | Change from (w, r, c) to (w, r, svc), remove ~90 lines of orchestration |
| internal/router/roles.go | Major | Change 7 functions from (w, r, c, pc) to (w, r, svc), remove guards + goroutines |
| internal/router/permissions.go | Major | Change 7 functions from (w, r, c, pc) to (w, r, svc), remove guards + goroutines |
| internal/router/role_permissions.go | Major | Change 8 functions from (w, r, c, pc) to (w, r, svc), remove guards + goroutines |
| internal/router/mux.go | Moderate | Update ~23 handler registrations to pass svc instead of *c/pc/(driver, pc, mgr) |
| internal/service/users_test.go | New | Service-level tests |
| internal/service/rbac_test.go | New | Service-level tests |

Not changed: internal/auth/auth.go (reused as-is), internal/middleware/authorization.go (reused as-is), internal/db/assemble.go (reused as-is), mcp/ (Phase 7).

---

## Testing

### UserService Tests (SQLite)

internal/service/users_test.go:
- Create with valid input succeeds, returns hashed password (hash != plaintext)
- Create with empty username → ValidationError field "username"
- Create with empty name → ValidationError field "name"
- Create with empty email → ValidationError field "email"
- Create with empty password → ValidationError field "password"
- Create with password < 8 chars → ValidationError field "password"
- Create with duplicate email → ConflictError
- Create with duplicate username → ConflictError
- Create with role and IsAdmin=false → ForbiddenError
- Create with role and IsAdmin=true succeeds
- Create with empty role defaults to viewer (verify stored role matches viewer role ID)
- Update with new password hashes correctly
- Update with empty password preserves existing hash
- Update changing email to existing email → ConflictError
- Update changing email to own email passes (no false conflict)
- Update changing username to existing username → ConflictError
- Update system user role → ForbiddenError
- Update role change with IsAdmin=false → ForbiddenError
- Delete system user → ForbiddenError
- Delete non-existent → NotFoundError
- Delete valid user succeeds
- GetUser non-existent → NotFoundError
- GetUserFull returns composed view with role label
- ReassignDelete with SystemUserID → ForbiddenError
- ReassignDelete with same user as target → ValidationError
- ReassignDelete succeeds, returns correct counts

### RBACService Tests (SQLite)

internal/service/rbac_test.go:
- Create role with valid label succeeds
- Create role with empty label → ValidationError
- Update role label succeeds (non-protected)
- Update system-protected role label → ForbiddenError
- Update system-protected role with same label succeeds (no rename)
- Delete non-protected role succeeds
- Delete system-protected role → ForbiddenError
- Delete non-existent role → NotFoundError
- Create permission with valid label ("test:read") succeeds
- Create permission with invalid label ("bad") → ValidationError
- Update system-protected permission label → ForbiddenError
- Delete system-protected permission → ForbiddenError
- Create role-permission link succeeds
- Delete role-permission link on system-protected role → ForbiddenError
- Delete role with users assigned → ConflictError "role has N user(s) assigned"
- Delete role without users assigned succeeds (also cleans up junction rows)
- SyncRolePermissions replaces all links (verify count before/after)
- SyncRolePermissions with empty slice deletes all links
- SyncRolePermissions with invalid permission ID returns failed IDs in first return value (non-nil)
- SyncRolePermissions with all valid IDs returns nil failed slice
- ListRolePermissionsByRoleID with valid role returns expected links

### Verification

```
just check              # compile check
just test               # unit tests pass
```

---

## Implementation Order

These two services have a shared concern (UserService.CreateUser needs to call `GetRoleByLabel("viewer")` which is a DbDriver method, not an RBACService method). They do NOT depend on each other at the service level. They can be built in parallel after a shared prerequisite.

### Step 0 — Prerequisites (do first, before parallel streams):

**0a. Registry wiring:** Add both `Users` and `RBAC` fields to the Registry struct in `internal/service/service.go` and update `NewRegistry` to initialize them. This must be done once before parallel streams start, to avoid merge conflicts on the same file. Both `NewUserService` and `NewRBACService` can initially be minimal constructors — the parallel streams will fill in the implementations.

**0b. Fix `writeHTMXError` JSON injection bug:** In `internal/service/errors_http.go`, the `writeHTMXError` function string-concatenates the error message into a JSON HX-Trigger header without escaping. Fix: use `json.Marshal` on the message string before embedding it in the header template. This is a one-line fix, independent of both streams, and must be done before Phase 4 routes more error paths through `HandleServiceError`.

### Stream A — UserService (~1-2 sessions):
1. Implement internal/service/users.go (UserService struct + all methods)
2. Write internal/service/users_test.go
3. Rewire internal/admin/handlers/users.go signatures to (svc)
4. Rewire internal/router/users.go + userComposite.go to (w, r, svc) pattern
5. Do NOT touch mux.go — deferred to Step 2

### Stream B — RBACService (~1-2 sessions):
1. Implement internal/service/rbac.go (RBACService struct + all methods)
2. Write internal/service/rbac_test.go
3. Rewire internal/admin/handlers/roles.go signatures to (svc)
4. Rewire internal/router/roles.go + permissions.go + role_permissions.go to (w, r, svc) pattern
5. Do NOT touch mux.go — deferred to Step 2

### Step 2 — Mux Wiring + Integration Verification (do after BOTH streams complete):

All mux.go changes happen in a single step after both streams are done. This avoids merge conflicts — mux.go is the single most-modified file across every phase and both streams touch it. Update all ~23 handler registrations (user + RBAC routes, both API and admin) to pass `svc` instead of `*c`/`pc`/`(driver, pc, mgr)`.

```
just check
just test
```
