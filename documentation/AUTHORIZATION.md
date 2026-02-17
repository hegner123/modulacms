# Authorization System

ModulaCMS uses role-based access control (RBAC) with granular `resource:operation` permissions. Every admin API endpoint requires a specific permission. The system is fail-closed: if authorization state is missing or incomplete, access is denied.

## Overview

```
Request
  -> DefaultMiddlewareChain
       -> HTTPAuthenticationMiddleware   (resolves session cookie to user)
       -> PermissionInjector             (resolves user's role to PermissionSet in context)
  -> ServeMux route match
       -> RequirePermission / RequireResourcePermission   (checks PermissionSet)
       -> Handler
```

Authentication identifies *who* the user is. Authorization determines *what* they can do. These are separate middleware layers:

1. `HTTPAuthenticationMiddleware` reads the session cookie, looks up the user, and stores a `*Users` value in context. Unauthenticated requests pass through with no user in context.
2. `PermissionInjector` reads the authenticated user's role, looks up the role's permissions from `PermissionCache`, and stores a `PermissionSet` and `isAdmin` boolean in context. Unauthenticated requests pass through with no `PermissionSet`.
3. Per-route permission guards (`RequirePermission`, `RequireResourcePermission`, etc.) check the `PermissionSet` in context. If the set is nil or lacks the required permission, the request is rejected with 403.

## Permission Labels

Permissions follow the format `resource:operation`:

- **Resource**: lowercase alphanumeric and underscores (`[a-z0-9_]`). Examples: `content`, `admin_tree`, `ssh_keys`.
- **Operation**: lowercase alphabetic (`[a-z]`). Examples: `read`, `create`, `update`, `delete`, `admin`.
- The literal string `*` is rejected.

Labels are validated by `middleware.ValidatePermissionLabel()` before storage. The validation is character-by-character (no regex).

## Bootstrap Permissions

The system ships with 47 permissions created by `CreateBootstrapData`. All are system-protected (cannot be deleted or renamed via API).

| Resource | Operations |
|----------|-----------|
| `content` | read, create, update, delete, admin |
| `datatypes` | read, create, update, delete, admin |
| `fields` | read, create, update, delete, admin |
| `media` | read, create, update, delete, admin |
| `routes` | read, create, update, delete, admin |
| `users` | read, create, update, delete, admin |
| `roles` | read, create, update, delete, admin |
| `permissions` | read, create, update, delete, admin |
| `admin_tree` | read, create, update, delete, admin |
| `sessions` | read, delete, admin |
| `ssh_keys` | read, create, delete, admin |
| `config` | read, update, admin |

Additional permissions can be created at runtime via the `/api/v1/permissions` endpoint.

## Bootstrap Roles

Three system-protected roles are created at startup:

### admin (all 47 permissions)

Full access to everything. Admin users bypass all permission checks entirely via the `ContextIsAdmin()` fast path -- their `PermissionSet` is never consulted. The admin bypass is based on a boolean flag in context, not a wildcard entry in the permission set.

### editor (28 permissions)

Content management access:

| Capability | Permissions |
|-----------|------------|
| Content CRUD | content:read, content:create, content:update, content:delete |
| Datatypes CRUD | datatypes:read, datatypes:create, datatypes:update, datatypes:delete |
| Fields CRUD | fields:read, fields:create, fields:update, fields:delete |
| Media CRUD | media:read, media:create, media:update, media:delete |
| Routes CRUD | routes:read, routes:create, routes:update, routes:delete |
| Admin Tree CRUD | admin_tree:read, admin_tree:create, admin_tree:update, admin_tree:delete |
| Read-only | users:read, sessions:read, ssh_keys:read, config:read |

Editors cannot manage roles, permissions, tokens, or system configuration. They cannot access `:admin` or `:delete` operations on users, sessions, or SSH keys.

### viewer (3 permissions)

Read-only access: `content:read`, `media:read`, `routes:read`.

This is the default role assigned to new users via `/api/v1/auth/register`.

## Database Schema

### Tables

**roles** -- Role definitions.

| Column | Type | Notes |
|--------|------|-------|
| role_id | RoleID (ULID) | Primary key |
| label | TEXT | Unique name (e.g., "admin", "editor") |
| system_protected | INTEGER/BOOLEAN | If true, cannot be deleted or renamed |

**permissions** -- Permission definitions.

| Column | Type | Notes |
|--------|------|-------|
| permission_id | PermissionID (ULID) | Primary key |
| label | TEXT | UNIQUE. The `resource:operation` string |
| system_protected | INTEGER/BOOLEAN | If true, cannot be deleted or renamed |

**role_permissions** -- Junction table mapping roles to permissions.

| Column | Type | Notes |
|--------|------|-------|
| id | RolePermissionID (ULID) | Primary key |
| role_id | RoleID | FK to roles |
| permission_id | PermissionID | FK to permissions |

Schema location: `sql/schema/26_role_permissions/`.

### Key Queries

- `ListPermissionLabelsByRoleID(roleID)` -- JOIN query that returns `[]string` of permission labels for a role. Used by `PermissionCache.Load()`.
- `GetRoleByLabel(label)` -- Fetch a role by label string. Used for viewer role assignment during registration.
- `GetPermissionByLabel(label)` -- Fetch a permission by label string.

## PermissionCache

`PermissionCache` (`internal/middleware/authorization.go`) is an in-memory cache that maps `RoleID` to `PermissionSet`. It is the only component that touches the database for authorization decisions.

### Lifecycle

1. **Startup** (`cmd/serve.go`): `pc.Load(driver)` is called. If this fails, the server does not start.
2. **Periodic refresh**: `pc.StartPeriodicRefresh(ctx, driver, 60*time.Second)` runs a background goroutine that reloads the cache every 60 seconds. Consecutive failures are logged with dampening (first 3, then every 30th).
3. **On-demand refresh**: When roles, permissions, or role_permissions are created, updated, or deleted via API handlers, the handler triggers `go pc.Load(driver)` asynchronously.
4. **Shutdown**: The context passed to `StartPeriodicRefresh` is cancelled, stopping the goroutine.

### Concurrency Model

The cache uses build-then-swap:

1. `Load()` builds a new map without holding any lock.
2. A write lock is held only for the pointer swap (nanoseconds).
3. Readers use `RLock` and are never blocked during database queries.

### Data Structure

```
PermissionCache
  cache:       map[RoleID]PermissionSet    // role -> set of permission labels
  isAdmin:     map[RoleID]bool             // true if role.Label == "admin"
  adminRoleID: RoleID                      // cached admin role ID
  lastLoaded:  time.Time                   // timestamp of last successful load
```

`PermissionSet` is `map[string]struct{}` with `Has()`, `HasAny()`, and `HasAll()` methods.

## Middleware Functions

All middleware functions live in `internal/middleware/authorization.go`.

### PermissionInjector(pc)

Runs in the `DefaultMiddlewareChain` on every request. If a user is authenticated, it:
1. Reads `user.Role` (a string containing the RoleID).
2. Looks up the `PermissionSet` from `pc.PermissionsForRole(roleID)`.
3. Stores the `PermissionSet` and `isAdmin` boolean in request context.

Unauthenticated requests pass through with no `PermissionSet` in context.

### RequirePermission(permission)

Checks for a single permission string. Used when the endpoint has a fixed permission regardless of HTTP method.

```
RequirePermission("config:read")
RequirePermission("plugins:admin")
RequirePermission("import:create")
```

### RequireResourcePermission(resource)

Maps the HTTP method to an operation automatically:

| HTTP Method | Operation | Permission Checked |
|-------------|-----------|-------------------|
| GET | read | `resource:read` |
| POST | create | `resource:create` |
| PUT | update | `resource:update` |
| PATCH | update | `resource:update` |
| DELETE | delete | `resource:delete` |

Unmapped methods (HEAD, OPTIONS, TRACE, etc.) are denied with 403.

```
RequireResourcePermission("content")   // GET -> content:read, POST -> content:create, etc.
RequireResourcePermission("roles")     // GET -> roles:read, DELETE -> roles:delete, etc.
```

### RequireAnyPermission(permissions...)

OR logic. The request is allowed if the user has at least one of the listed permissions.

### RequireAllPermissions(permissions...)

AND logic. The request is allowed only if the user has all listed permissions.

### Admin Bypass

All `Require*` functions check `ContextIsAdmin()` first. If true, the handler chain continues immediately without consulting the `PermissionSet`. This means the admin role never needs explicit permission entries to access any endpoint.

### Fail-Closed Behavior

If `ContextPermissions()` returns nil (no `PermissionSet` in context), all `Require*` functions return 403. This covers:

- Unauthenticated requests (no session cookie).
- Users whose role is not in the cache (e.g., deleted role, cache not yet refreshed).
- Missing `PermissionInjector` in the middleware chain (programming error).

### 403 Response Format

```json
{"error": "forbidden"}
```

The response deliberately omits which permission was required. Permission names are logged server-side but never exposed to the client.

## Route-to-Permission Mapping

Every admin endpoint in `internal/router/mux.go` is wrapped with a permission guard. The mapping:

| Endpoint | Guard | Permission(s) |
|----------|-------|---------------|
| `/api/v1/admin/tree/` | RequireResourcePermission | admin_tree:read/create/update/delete |
| `/api/v1/admincontentdatas` | RequireResourcePermission | content:read/create/update/delete |
| `/api/v1/admincontentfields` | RequireResourcePermission | content:read/create/update/delete |
| `/api/v1/admindatatypes` | RequireResourcePermission | datatypes:read/create/update/delete |
| `/api/v1/adminfields` | RequireResourcePermission | fields:read/create/update/delete |
| `/api/v1/admindatatypefields` | RequireResourcePermission | fields:read/create/update/delete |
| `/api/v1/adminroutes` | RequireResourcePermission | routes:read/create/update/delete |
| `/api/v1/contentdata` | RequireResourcePermission | content:read/create/update/delete |
| `/api/v1/contentfields` | RequireResourcePermission | content:read/create/update/delete |
| `/api/v1/content/batch` | RequirePermission | content:update |
| `/api/v1/datatype` | RequireResourcePermission | datatypes:read/create/update/delete |
| `/api/v1/datatype/full` | RequirePermission | datatypes:read |
| `/api/v1/datatypefields` | RequireResourcePermission | fields:read/create/update/delete |
| `/api/v1/fields` | RequireResourcePermission | fields:read/create/update/delete |
| `/api/v1/media` | RequireResourcePermission | media:read/create/update/delete |
| `/api/v1/media/health` | RequirePermission | media:admin |
| `/api/v1/media/cleanup` | RequirePermission | media:admin |
| `/api/v1/mediadimensions` | RequireResourcePermission | media:read/create/update/delete |
| `/api/v1/routes` | RequireResourcePermission | routes:read/create/update/delete |
| `/api/v1/roles` | RequireResourcePermission | roles:read/create/update/delete |
| `/api/v1/permissions` | RequireResourcePermission | permissions:read/create/update/delete |
| `/api/v1/role-permissions` | RequireResourcePermission | roles:read/create/update/delete |
| `/api/v1/role-permissions/role/` | RequirePermission | roles:read |
| `/api/v1/sessions` | RequirePermission | sessions:read |
| `/api/v1/sessions/{id}` | RequireResourcePermission | sessions:read/create/update/delete |
| `/api/v1/tables` | RequireResourcePermission | datatypes:read/create/update/delete |
| `/api/v1/tokens` | RequireResourcePermission | tokens:read/create/update/delete |
| `/api/v1/users` | RequireResourcePermission | users:read/create/update/delete |
| `/api/v1/users/full` | RequirePermission | users:read |
| `/api/v1/usersoauth` | RequireResourcePermission | users:read/create/update/delete |
| `/api/v1/ssh-keys` (POST) | RequirePermission | ssh_keys:create |
| `/api/v1/ssh-keys` (GET) | RequirePermission | ssh_keys:read |
| `/api/v1/ssh-keys/{id}` (DELETE) | RequirePermission | ssh_keys:delete |
| `/api/v1/import/*` | RequirePermission | import:create |
| `/api/v1/admin/config` (GET) | RequirePermission | config:read |
| `/api/v1/admin/config` (PATCH) | RequirePermission | config:update |
| `/api/v1/admin/config/meta` | RequirePermission | config:read |
| `/api/v1/admin/plugins/routes` (GET) | RequirePermission | plugins:read |
| `/api/v1/admin/plugins/routes/approve` | RequirePermission | plugins:admin |
| `/api/v1/admin/plugins/routes/revoke` | RequirePermission | plugins:admin |
| Plugin admin endpoints (read) | RequirePermission | plugins:read |
| Plugin admin endpoints (mutations) | RequirePermission | plugins:admin |

### Unprotected Endpoints

These endpoints have no permission guards:

| Endpoint | Reason |
|----------|--------|
| `POST /api/v1/auth/login` | Must be accessible before authentication |
| `POST /api/v1/auth/logout` | Must be accessible to end sessions |
| `GET /api/v1/auth/me` | Returns current user info (or 401) |
| `POST /api/v1/auth/register` | Self-registration (always assigns viewer role) |
| `POST /api/v1/auth/reset` | Password reset |
| `GET /api/v1/auth/oauth/*` | OAuth flow initiation and callback |
| `GET /` (SlugHandler) | Public content delivery |
| `/favicon.ico` | Static asset |

Auth endpoints use rate limiting (10 requests/minute) instead of permission checks.

## Handler-Level Guards

Beyond route-level middleware, several handlers enforce additional authorization rules:

### System-Protected Records

Roles and permissions with `system_protected = true` have restricted mutations:

- **Delete**: Returns 403 if the record is system-protected.
- **Rename**: Returns 403 if the label of a system-protected record would change (the label field is checked against the existing record).
- Other fields on system-protected records can be updated normally.

This applies to handlers in `roles.go`, `permissions.go`, and `role_permissions.go`.

### Role Assignment

- **User creation** (`ApiCreateUser`): Non-admin users cannot set a role. If no role is provided, the viewer role is assigned by default.
- **User update** (`ApiUpdateUser`): Non-admin users cannot change their own or others' roles. Attempting to change a role without admin status returns 403.
- **Registration** (`RegisterHandler`): Always assigns the viewer role regardless of request body content. The role field is not accepted in the registration request.

### Cache Invalidation

Handlers that modify roles, permissions, or role_permissions trigger an asynchronous cache refresh:

```go
go pc.Load(driver)
```

This runs in a separate goroutine to avoid blocking the HTTP response. The next request after the goroutine completes will see the updated permissions. There is a brief window (typically milliseconds) where the old permissions are still in effect.

## Source Files

| File | Purpose |
|------|---------|
| `internal/middleware/authorization.go` | PermissionCache, PermissionSet, all Require* middleware, ValidatePermissionLabel |
| `internal/middleware/authorization_test.go` | 15 tests covering all middleware functions |
| `internal/middleware/http_chain.go` | DefaultMiddlewareChain (includes PermissionInjector) |
| `internal/db/role_permission.go` | RolePermissions wrapper, all 3 driver implementations, audited commands |
| `internal/db/db.go` | DbDriver interface (RolePermissions section), CreateBootstrapData |
| `internal/router/mux.go` | Route registration with permission wrappers |
| `internal/router/roles.go` | Role CRUD handlers with system-protected guards |
| `internal/router/permissions.go` | Permission CRUD handlers with label validation |
| `internal/router/role_permissions.go` | Role-permission junction CRUD handlers |
| `internal/router/users.go` | User handlers with role assignment guards |
| `internal/router/auth.go` | RegisterHandler with forced viewer role |
| `cmd/serve.go` | PermissionCache initialization and lifecycle |
| `sql/schema/26_role_permissions/` | SQL schema and queries for all 3 databases |
| `sql/schema/1_permissions/` | Permissions schema (includes system_protected, UNIQUE label) |
| `sql/schema/2_roles/` | Roles schema (includes system_protected) |

## Adding New Permissions

1. Add the permission label to the `rbacPermissionLabels` slice in `CreateBootstrapData` (`internal/db/db.go`).
2. Assign the permission to the appropriate roles in the same function (admin gets all automatically via the loop; add to `editorPermLabels` or `viewerPermLabels` as needed).
3. Wrap the new endpoint in `mux.go` with `RequirePermission("resource:operation")` or `RequireResourcePermission("resource")`.
4. Run `just test` to verify.

For runtime-created permissions (not part of bootstrap), use the `POST /api/v1/permissions` endpoint and `POST /api/v1/role-permissions` to assign them to roles. These are not system-protected and can be deleted later.
