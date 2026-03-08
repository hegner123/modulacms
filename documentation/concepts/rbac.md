# Role-Based Access Control

ModulaCMS uses role-based access control (RBAC) with granular permissions to protect every API endpoint. Users are assigned roles, roles are linked to permissions, and middleware enforces permission checks on each request. The system is fail-closed: missing permissions result in a 403 Forbidden response.

## Bootstrap Roles

The CMS ships with three built-in roles created during installation:

| Role | Permissions | Behavior |
|------|-------------|----------|
| `admin` | All 47 permissions | Bypasses permission checks entirely via `ContextIsAdmin` |
| `editor` | 28 permissions | Content management, media, routes, datatypes, fields -- no user or role management |
| `viewer` | 3 permissions | `content:read`, `media:read`, `routes:read` |

These roles are system-protected. They cannot be deleted or renamed via the API.

## Permissions

Permissions follow the `resource:operation` label format. The label is a colon-separated pair where the left side identifies the resource and the right side identifies the operation.

Examples:

| Label | Grants |
|-------|--------|
| `content:read` | Read content data and content fields |
| `content:create` | Create content data nodes and content fields |
| `content:update` | Update content data, move nodes, reorder siblings, publish |
| `content:delete` | Delete content data nodes |
| `media:create` | Upload media files |
| `users:read` | List and view user accounts |
| `roles:update` | Modify role definitions |
| `permissions:read` | List permissions |

Permission labels are validated with `ValidatePermissionLabel`, which checks the `resource:operation` format character-by-character (no regex). Both the resource and operation segments must be non-empty and contain only lowercase letters, numbers, underscores, and hyphens.

System-protected permissions cannot be deleted or renamed.

## Custom Roles

You can create custom roles with any subset of the available permissions:

```bash
# Create a role
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"label": "contributor"}'

# Grant permissions to the role
curl -X POST http://localhost:8080/api/v1/rolepermissions \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"role_id": "01ABC...", "permission_id": "01DEF..."}'
```

The `RolePermission` junction table maps roles to permissions. Each row grants one permission to one role.

## Permission Cache

The `PermissionCache` is an in-memory map of role IDs to permission sets. It eliminates database queries on every request.

### Loading

At startup, the cache loads all roles, permissions, and role-permission mappings from the database into memory. The cache builds a `map[RoleID]PermissionSet` where `PermissionSet` is a `map[string]struct{}` of permission labels.

### Refresh

The cache refreshes automatically every 60 seconds via `StartPeriodicRefresh`. The refresh uses a build-then-swap strategy: a new map is built from the database, then atomically swapped in. This provides lock-free reads during normal operation -- readers access the current map without acquiring a lock, while the refresh goroutine only holds the write lock during the pointer swap.

### Manual Refresh

Route handlers that modify roles or permissions trigger an async `pc.Load(driver)` to refresh the cache immediately, so permission changes take effect without waiting for the next periodic refresh.

## Request Flow

Permission enforcement happens in the middleware chain:

1. **Authentication middleware** identifies the user from session cookie or API key.
2. **`PermissionInjector`** resolves the user's role to a `PermissionSet` and an admin boolean, then stores both in the request context.
3. **Permission guard middleware** checks the context for required permissions before the handler executes.

### Permission Guard Functions

| Function | Behavior |
|----------|----------|
| `RequirePermission("content:read")` | Requires exactly this permission. Admin bypasses. |
| `RequireResourcePermission("content")` | Auto-maps HTTP method to operation: GET to `content:read`, POST to `content:create`, PUT/PATCH to `content:update`, DELETE to `content:delete`. |
| `RequireAnyPermission("media:read", "media:create")` | Requires at least one of the listed permissions (OR logic). |
| `RequireAllPermissions("content:read", "media:read")` | Requires all listed permissions (AND logic). |

### Admin Bypass

Admin users bypass all permission checks. The bypass is implemented via a `ContextIsAdmin` boolean stored in the request context -- not via a wildcard permission in the `PermissionSet`. This is a deliberate design choice: checking `ContextIsAdmin` is a single boolean test, and the admin's `PermissionSet` does not need to enumerate every possible permission.

### Fail-Closed

If the `PermissionInjector` middleware cannot find a `PermissionSet` in the context (because the user is not authenticated or the role lookup failed), all permission guard middleware returns 403 Forbidden. There is no fallback to a default role.

## Registration

New user registration always assigns the `viewer` role. Non-admin users cannot set or change their own role or any other user's role via the API.

## Data Model

```go
type Role struct {
    RoleID RoleID `json:"role_id"`
    Label  string `json:"label"`
}

type Permission struct {
    PermissionID PermissionID `json:"permission_id"`
    Label        string       `json:"label"`
}

type RolePermission struct {
    ID           RolePermissionID `json:"id"`
    RoleID       RoleID           `json:"role_id"`
    PermissionID PermissionID     `json:"permission_id"`
}
```

The `role_permissions` junction table in `sql/schema/26_role_permissions/` maps roles to permissions. Bootstrap data (the three default roles and their permission assignments) is seeded by `CreateBootstrapData` during installation.

## API Endpoints

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/roles` | `roles:read` | List all roles |
| POST | `/api/v1/roles` | `roles:create` | Create a role |
| PUT | `/api/v1/roles/` | `roles:update` | Update a role |
| DELETE | `/api/v1/roles/` | `roles:delete` | Delete a role |
| GET | `/api/v1/permissions` | `permissions:read` | List all permissions |
| POST | `/api/v1/permissions` | `permissions:create` | Create a permission |
| PUT | `/api/v1/permissions/` | `permissions:update` | Update a permission |
| DELETE | `/api/v1/permissions/` | `permissions:delete` | Delete a permission |
| GET | `/api/v1/rolepermissions` | `roles:read` | List role-permission mappings |
| POST | `/api/v1/rolepermissions` | `roles:update` | Grant a permission to a role |
| DELETE | `/api/v1/rolepermissions/` | `roles:update` | Revoke a permission from a role |
