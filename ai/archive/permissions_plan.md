Authorization Middleware Plan

Context

ModulaCMS currently has binary authorization: authenticated or not, admin or not. The adminOnly closure in mux.go compares user.Role against a pre-resolved admin role ID. There is no granular permission checking -- every authenticated user can access every non-admin endpoint. The existing permissions table and roles.permissions TEXT field are largely unused.

This plan adds proper RBAC (Role-Based Access Control) with resource:operation granularity, a role_permissions junction table, an in-memory permission cache, and per-route middleware wrappers.

No migration strategy is needed -- there are no existing deployments. All databases are created fresh via CreateBootstrapData.

---
Phase 1: Database Layer

1a. New typed ID

Add RolePermissionID to internal/db/types/types_ids.go following the existing pattern (String, Value, Scan, MarshalJSON, UnmarshalJSON, Validate, IsZero, NewRolePermissionID).

1b. Schema change: drop UNIQUE constraint on roles.permissions (SQLite only)

The permissions column differs across dialects:
- SQLite: `permissions TEXT NOT NULL UNIQUE` -- drop the UNIQUE constraint
- PostgreSQL: `permissions jsonb` (nullable, no UNIQUE) -- no change needed
- MySQL: `permissions JSON NULL` (nullable, no UNIQUE) -- no change needed

Only the SQLite schema file needs modification:
- sql/schema/2_roles/schema.sql -- drop UNIQUE on permissions column

The SQLite query file also contains DDL that must match:
- sql/schema/2_roles/queries.sql -- update CreateRoleTable to drop UNIQUE on permissions column

The PostgreSQL and MySQL schemas already lack this constraint, so no changes to their schema or query files are needed for this step.

The column stays populated for backward compatibility. Its value will be set to a descriptive string during bootstrap (e.g. `{"role": "editor"}`) but is not used by the new permission system.

1c. New schema: sql/schema/26_role_permissions/

Six files (schema.sql, schema_mysql.sql, schema_psql.sql, queries.sql, queries_mysql.sql, queries_psql.sql).

Table structure:
CREATE TABLE IF NOT EXISTS role_permissions (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    role_id TEXT NOT NULL REFERENCES roles ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);

Key query -- join through junction to get permission labels for a role:
-- name: ListPermissionLabelsByRoleID :many
SELECT p.label FROM role_permissions rp
JOIN permissions p ON rp.permission_id = p.permission_id
WHERE rp.role_id = ? ORDER BY p.label;

Plus standard CRUD queries: Create, Delete, List, ListByRoleID, ListByPermissionID, DeleteByRoleID, Count.

1d. sqlc config updates

Add to all three engine sections in sql/sqlc.yml:
- Rename: role_permission -> RolePermissions
- Column overrides for id (RolePermissionID), role_id (RoleID), permission_id (PermissionID)

Run just sqlc to regenerate.

1e. DbDriver interface + wrapper implementations

New file: internal/db/role_permission.go

Application types:
- RolePermissions struct (ID, RoleID, PermissionID)
- CreateRolePermissionParams (RoleID, PermissionID)

DbDriver additions:
- CreateRolePermissionsTable() error
- CreateRolePermission(ctx, auditCtx, params) (*RolePermissions, error)
- DeleteRolePermission(ctx, auditCtx, id) error
- DeleteRolePermissionsByRoleID(ctx, auditCtx, roleID) error
- ListRolePermissions() (*[]RolePermissions, error)
- ListRolePermissionsByRoleID(roleID) (*[]RolePermissions, error)
- ListPermissionLabelsByRoleID(roleID) (*[]string, error)
- CountRolePermissions() (*int64, error)

Implement on all three wrappers (Database, MysqlDatabase, PsqlDatabase) with audited commands. Add CreateRolePermissionsTable() to CreateAllTables() (after roles, before users).

1f. Add system_protected column to roles and permissions schemas

Add a `system_protected` column to both the `roles` and `permissions` tables. The column type varies by dialect to match existing conventions:

Schema change for roles:
- SQLite (schema.sql, queries.sql): `system_protected INTEGER NOT NULL DEFAULT 0`
- PostgreSQL (schema_psql.sql, queries_psql.sql): `system_protected BOOLEAN NOT NULL DEFAULT FALSE`
- MySQL (schema_mysql.sql, queries_mysql.sql): `system_protected BOOLEAN NOT NULL DEFAULT FALSE`

Schema change for permissions:
- SQLite (schema.sql, queries.sql): `system_protected INTEGER NOT NULL DEFAULT 0`
- PostgreSQL (schema_psql.sql, queries_psql.sql): `system_protected BOOLEAN NOT NULL DEFAULT FALSE`
- MySQL (schema_mysql.sql, queries_mysql.sql): `system_protected BOOLEAN NOT NULL DEFAULT FALSE`

Both the schema files AND the query files must be updated for each table. The query files contain DDL blocks (CreateRoleTable, CreatePermissionTable) that must match their corresponding schema files. The INSERT/UPDATE/SELECT queries in the query files must also include the new column.

Update sqlc config to include the new column in generated types. Update the Role and Permission application-level structs in internal/db/ to include SystemProtected bool.

1g. Add UNIQUE constraint on permissions.label

None of the three dialects currently have a UNIQUE constraint on `permissions.label`. Since the RBAC system uses label strings as the canonical permission identifiers (cached in PermissionSet, referenced by RequirePermission middleware), duplicate labels would cause junction table ambiguity: two different permission_id rows with the same label, and no way to know which one to link to a role.

Add a UNIQUE constraint on `label` in all six permissions files (schema + queries):

- SQLite (schema.sql, queries.sql): `label TEXT NOT NULL UNIQUE`
- PostgreSQL (schema_psql.sql, queries_psql.sql): `label TEXT NOT NULL UNIQUE`
- MySQL (schema_mysql.sql, queries_mysql.sql): `label VARCHAR(255) NOT NULL, CONSTRAINT label_unique UNIQUE (label)`

This is a new constraint on an existing table, but since there are no existing deployments, all databases are created fresh and no migration is needed.

1h. Update CreateBootstrapData

Create discrete permission rows in the permissions table using label as the resource:operation string and table_id as the resource name. The mode column is set to 0 (unused). All bootstrap permissions are created with system_protected = 1.

Permission label validation: labels must match the pattern `resource:operation` where resource is lowercase alphanumeric/underscores and operation is lowercase alphabetic. The literal string `*` is reserved and must never be stored as a permission label. Validation is enforced in the CreatePermission and UpdatePermission handler code (see Phase 3i).

Permission strings (47 total):

┌─────────────┬──────────────────────────────────────────┐
│  Resource   │               Operations                 │
├─────────────┼──────────────────────────────────────────┤
│ datatypes   │ create, read, update, delete             │
├─────────────┼──────────────────────────────────────────┤
│ fields      │ create, read, update, delete             │
├─────────────┼──────────────────────────────────────────┤
│ content     │ create, read, update, delete             │
├─────────────┼──────────────────────────────────────────┤
│ media       │ create, read, update, delete, admin      │
├─────────────┼──────────────────────────────────────────┤
│ users       │ create, read, update, delete             │
├─────────────┼──────────────────────────────────────────┤
│ roles       │ create, read, update, delete             │
├─────────────┼──────────────────────────────────────────┤
│ permissions │ create, read, update, delete             │
├─────────────┼──────────────────────────────────────────┤
│ sessions    │ read, delete                             │
├─────────────┼──────────────────────────────────────────┤
│ ssh_keys    │ create, read, delete                     │
├─────────────┼──────────────────────────────────────────┤
│ tokens      │ create, read, delete                     │
├─────────────┼──────────────────────────────────────────┤
│ routes      │ create, read, update, delete             │
├─────────────┼──────────────────────────────────────────┤
│ config      │ read, update                             │
├─────────────┼──────────────────────────────────────────┤
│ plugins     │ read, admin                              │
├─────────────┼──────────────────────────────────────────┤
│ import      │ create                                   │
├─────────────┼──────────────────────────────────────────┤
│ admin_tree  │ read                                     │
└─────────────┴──────────────────────────────────────────┘

Changes from original plan:
- permissions: expanded from read to full CRUD (create, read, update, delete) to match the existing PermissionsHandler/PermissionHandler which support POST, PUT, DELETE.
- media: added "admin" operation for media:admin (health check, cleanup).
- Removed "list" operation from all resources (see "Removed list permissions" design decision in Phase 2a).

Default roles and their permissions:

- admin: All 47 permissions (bypassed via IsAdmin in middleware, but junction rows created for visibility).
  roles.permissions TEXT value: `{"admin": true}` (unchanged from current bootstrap)
  system_protected: 1
  Junction rows: 47

- editor: Full CRUD on content, datatypes, fields, media, routes; read on users, roles, permissions, sessions; admin_tree:read.
  roles.permissions TEXT value: `{"editor": true}` (ignored by new system)
  system_protected: 1
  Junction rows: 25 (content:4 + datatypes:4 + fields:4 + media:4 + routes:4 + users:read + roles:read + permissions:read + sessions:read + admin_tree:read)

- viewer: Read on content, datatypes, fields, media, routes.
  roles.permissions TEXT value: `{"read": true}` (unchanged from current bootstrap)
  system_protected: 1
  Junction rows: 5

Bootstrap creates 77 total role_permission junction rows (47 + 25 + 5). The current CreateBootstrapData does not create an editor role -- it must be added alongside admin and viewer. All bootstrap inserts (permissions + junction rows) are executed sequentially within CreateBootstrapData. If any insert fails, the function returns an error and the server fails to start (fresh install). Partial bootstrap state is acceptable because databases are created fresh -- a failed bootstrap means the user fixes the issue and restarts, which recreates all tables.

The existing roles.permissions TEXT field stays populated for backward compatibility but is not used by the new system. The UNIQUE constraint is dropped (step 1b) so future custom roles do not need unique JSON values.

Files modified (Phase 1):

- internal/db/types/types_ids.go -- add RolePermissionID
- sql/schema/2_roles/schema.sql -- drop UNIQUE on permissions column, add system_protected column
- sql/schema/2_roles/schema_psql.sql -- add system_protected column (no UNIQUE to drop)
- sql/schema/2_roles/schema_mysql.sql -- add system_protected column (no UNIQUE to drop)
- sql/schema/2_roles/queries.sql -- update CreateRoleTable DDL (drop UNIQUE on permissions, add system_protected), update CreateRole/UpdateRole/GetRole to include system_protected
- sql/schema/2_roles/queries_psql.sql -- update CreateRoleTable DDL (add system_protected), update CreateRole/UpdateRole/GetRole
- sql/schema/2_roles/queries_mysql.sql -- update CreateRoleTable DDL (add system_protected), update CreateRole/UpdateRole/GetRole
- sql/schema/1_permissions/schema.sql, schema_psql.sql, schema_mysql.sql -- add system_protected column, add UNIQUE constraint on label
- sql/schema/1_permissions/queries.sql -- update CreatePermissionTable DDL (add system_protected, add UNIQUE on label), update CreatePermission/UpdatePermission/GetPermission
- sql/schema/1_permissions/queries_psql.sql -- update CreatePermissionTable DDL, update queries
- sql/schema/1_permissions/queries_mysql.sql -- update CreatePermissionTable DDL, update queries
- sql/schema/26_role_permissions/ -- 6 new files
- sql/sqlc.yml -- add overrides for all 3 engines, add system_protected column mappings
- internal/db/role_permission.go -- new file
- internal/db/db.go -- DbDriver interface additions, CreateAllTables update, CreateBootstrapData update (all 3 wrapper structs, including new editor role and 77 junction rows)

SDK impact: the system_protected field is added to roles and permissions DB schemas. The Role and Permission application-level structs gain a SystemProtected bool field, which will appear in API JSON responses. SDK types (Go, TypeScript, Swift) should be updated to include this field. However, since all three SDKs use permissive JSON decoding (unknown fields are ignored), the SDKs will continue to work without updates -- the field will simply be absent from SDK type definitions until updated. SDK updates can be done as a follow-up task.

---
Phase 2: Middleware

2a. New file: internal/middleware/authorization.go

Types:

// PermissionSet is a set of permission strings for O(1) lookup.
type PermissionSet map[string]struct{}

func (ps PermissionSet) Has(perm string) bool
func (ps PermissionSet) HasAny(perms ...string) bool
func (ps PermissionSet) HasAll(perms ...string) bool

// PermissionCache holds role-to-permissions mappings in memory.
// Safe for concurrent reads. Refreshed via Load() using build-then-swap.
type PermissionCache struct {
    mu          sync.RWMutex
    cache       map[types.RoleID]PermissionSet  // role_id -> permissions
    adminRoleID types.RoleID
    isAdmin     map[types.RoleID]bool           // role_id -> true if admin
    lastLoaded  time.Time                 // timestamp of last successful Load
}

func NewPermissionCache() *PermissionCache
func (pc *PermissionCache) Load(driver db.DbDriver) error  // Populate from DB
func (pc *PermissionCache) PermissionsForRole(roleID types.RoleID) PermissionSet
func (pc *PermissionCache) IsAdmin(roleID types.RoleID) bool

Type safety: PermissionsForRole and IsAdmin accept types.RoleID (not plain string) to maintain the typed-ID convention used throughout the codebase. This prevents accidental lookups with wrong ID types (e.g., passing a UserID where a RoleID is expected). Internally, the cache maps use string(roleID) as keys, but the public API enforces type safety at the call site. PermissionInjector reads the user's RoleID from the authenticated user struct and passes it directly -- no string conversion needed at the caller.

Admin bypass implementation:

Admin bypass is determined by pc.IsAdmin(roleID), which checks the isAdmin map (populated from the role label "admin" during Load). The admin's PermissionSet in the cache contains its actual junction table permissions (for audit visibility), NOT a wildcard "*" entry. The RequirePermission middleware checks pc.IsAdmin first, and if true, bypasses the PermissionSet check entirely.

This avoids the security risk of storing "*" in the PermissionSet map, where it could collide with user-created permission labels.

Admin role identity invariant: the admin role is identified by `role.Label == "admin"` during cache Load(). The roles table has a UNIQUE constraint on the `label` column, which prevents creating a second role with label "admin". This UNIQUE constraint is a security dependency -- if it were ever dropped, multiple roles could claim admin status. The constraint must not be removed. Additionally, mutation of the label field on system_protected roles is blocked (see Phase 3h) to prevent renaming the admin role and breaking the IsAdmin detection, which would lock all users out of admin operations.

Load() locking strategy -- build-then-swap:

Load() builds the entire new cache map in a local variable with NO lock held. It queries all roles, then for each role calls ListPermissionLabelsByRoleID to build a PermissionSet. Only after the new map is fully built does it acquire a write lock, swap pc.cache, pc.adminRoleID, pc.isAdmin, and pc.lastLoaded, and release the lock. This means:
- Readers are NEVER blocked during the DB queries (which may take milliseconds).
- The write lock is held only for the pointer swap (nanoseconds).
- If the DB queries fail, the old cache remains untouched.

func (pc *PermissionCache) Load(driver db.DbDriver) error {
    // Build new cache (no lock)
    newCache := make(map[types.RoleID]PermissionSet)
    var newAdminRoleID types.RoleID
    newIsAdmin := make(map[types.RoleID]bool)
    roles, err := driver.ListRoles()
    if err != nil {
        return fmt.Errorf("loading permission cache: %w", err)
    }
    if len(*roles) > 1000 {
        return fmt.Errorf("refusing to load permission cache: %d roles exceeds safety limit of 1000", len(*roles))
    }
    for _, role := range *roles {
        if role.Label == "admin" {
            newAdminRoleID = role.RoleID
            newIsAdmin[role.RoleID] = true
        }
        labels, err := driver.ListPermissionLabelsByRoleID(role.RoleID)
        if err != nil {
            return fmt.Errorf("loading permissions for role %s: %w", role.RoleID, err)
        }
        ps := make(PermissionSet, len(*labels))
        for _, label := range *labels {
            ps[label] = struct{}{}
        }
        newCache[role.RoleID] = ps
    }

N+1 query note: Load() calls ListPermissionLabelsByRoleID once per role. With 3 bootstrap roles this is trivially fast. If the role count grows significantly (50+), consider replacing the per-role queries with a single grouped JOIN query that returns all (role_id, label) pairs and building the PermissionSets from the grouped result. This is an optimization that can be done later without changing the Load() contract.
    // Swap under write lock (nanoseconds)
    pc.mu.Lock()
    pc.cache = newCache
    pc.adminRoleID = newAdminRoleID
    pc.isAdmin = newIsAdmin
    pc.lastLoaded = time.Now()
    pc.mu.Unlock()
    return nil
}

Periodic cache refresh:

To provide bounded staleness guarantees, start a background goroutine that refreshes the cache every 60 seconds. This ensures that even if an event-driven refresh fails, permissions converge to the correct state within one minute.

Maximum staleness guarantee: Under normal operation (database reachable), the cache is at most 60 seconds stale after any permission mutation whose synchronous refresh failed. If the database becomes unreachable, the periodic refresh logs errors and the cache remains frozen at its last successful state. This is a deliberate design choice -- serving stale permissions during a database outage is preferable to failing all requests. When the database recovers, the next periodic tick refreshes the cache. There is no unbounded staleness in the "eventually consistent" sense; staleness is bounded by database availability.

func (pc *PermissionCache) StartPeriodicRefresh(ctx context.Context, driver db.DbDriver, interval time.Duration) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        consecutiveFailures := 0
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                if err := pc.Load(driver); err != nil {
                    consecutiveFailures++
                    if consecutiveFailures <= 3 || consecutiveFailures%30 == 0 {
                        // Log first 3 failures immediately, then every 30th failure (~30 min at 60s interval)
                        utility.DefaultLogger.Error("periodic permission cache refresh failed",
                            "error", err,
                            "consecutive_failures", consecutiveFailures,
                        )
                    }
                } else {
                    if consecutiveFailures > 0 {
                        utility.DefaultLogger.Info("periodic permission cache refresh recovered",
                            "after_failures", consecutiveFailures,
                        )
                    }
                    consecutiveFailures = 0
                }
            }
        }
    }()
}

Log dampening strategy: the first 3 consecutive failures are logged immediately (fast feedback during transient issues). After that, only every 30th failure is logged (~30 minutes at 60s interval). On recovery, a single info-level message reports how many failures preceded it. This prevents log flooding during extended outages (a 24-hour outage produces ~51 log entries instead of 1,440) while still providing visibility. The ticker interval itself is not modified -- the cache is retried every 60 seconds regardless of failure count, so recovery is not delayed by backoff.

Called in cmd/serve.go after initial Load(), using the server's shutdown context so the goroutine stops on graceful shutdown.

Context injection middleware:

// PermissionInjector resolves the user's role to a PermissionSet and stores
// it in context. Must run after HTTPAuthenticationMiddleware.
// Short-circuits for unauthenticated requests: if no user is in context,
// the handler chain continues immediately with no PermissionSet in context.
// This means public endpoints pay zero overhead beyond a nil check.
func PermissionInjector(pc *PermissionCache) func(http.Handler) http.Handler

Implementation detail: PermissionInjector checks middleware.AuthenticatedUser(ctx). If nil, it calls next.ServeHTTP immediately without touching context. If non-nil, it reads the user's role, looks up the PermissionSet from the cache (read lock), and stores it in context via context.WithValue. It also stores the isAdmin boolean in context for the RequirePermission middleware to use.

Permission check middlewares:

// RequirePermission checks for a single permission. Admin bypass is checked
// via pc.IsAdmin (from context), not via a "*" key in the PermissionSet.
func RequirePermission(permission string) func(http.Handler) http.Handler

// RequireAnyPermission checks for at least one permission (OR logic).
func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler

// RequireAllPermissions checks for all permissions (AND logic).
func RequireAllPermissions(permissions ...string) func(http.Handler) http.Handler

// RequireResourcePermission maps HTTP method to resource:operation automatically.
// GET -> resource:read
// POST -> resource:create
// PUT -> resource:update
// PATCH -> resource:update
// DELETE -> resource:delete
// Any other method -> 403 (unmapped methods are denied by default)
func RequireResourcePermission(resource string) func(http.Handler) http.Handler

Fail-closed invariant: RequirePermission, RequireAnyPermission, RequireAllPermissions, and RequireResourcePermission all return 403 if no PermissionSet is found in context. This is the final safety net against misconfigured middleware chains. If PermissionInjector was skipped or the middleware chain was mis-ordered, requests are denied rather than allowed. Implementation: each middleware calls ContextPermissions(ctx); if the result is nil and ContextIsAdmin(ctx) is false, write 403 JSON and return immediately without calling next.ServeHTTP.

CORS/OPTIONS safety: The CorsMiddleware runs at position 3 in both DefaultMiddlewareChain and AuthenticatedChain, before authentication (position 4) and before PermissionInjector (added after authentication). The CorsMiddleware short-circuits OPTIONS requests: after setting CORS headers and writing 204 No Content, it returns without calling next.ServeHTTP (see internal/middleware/cors.go lines 17-19). This means OPTIONS preflight requests NEVER reach the permission middleware. This is a security-relevant invariant: if CorsMiddleware is ever refactored to not short-circuit OPTIONS, RequireResourcePermission will break cross-origin requests by returning 403 on preflight.

Context helpers:
func ContextPermissions(ctx context.Context) PermissionSet
func ContextIsAdmin(ctx context.Context) bool

Error format: JSON 403 response:
{"error": "forbidden"}

The 403 response does NOT include the specific permission that was checked. Revealing the permission name (e.g. "missing permission: datatypes:create") would leak the internal authorization taxonomy to attackers. The specific permission failure is logged server-side with structured fields (user_id, role_id, required_permission, path, method) for debugging and intrusion detection.

Authorization failure logging:

Every 403 from the RBAC middleware is logged via slog with these structured fields:
- user_id (from context, or "anonymous")
- role_id (from context, or empty)
- required_permission (the permission string that was checked)
- path (request URL path)
- method (HTTP method)
- remote_addr (client IP)

This is essential for detecting privilege escalation attempts and compromised accounts.

Design decision -- removed list permissions:

The original plan defined separate "read" and "list" permissions per resource. This was removed because: (a) RequireResourcePermission maps GET to a single operation and distinguishing single-item GETs from collection GETs in middleware requires fragile path inspection; (b) shipping unused permissions into the database creates a semantic trap where a future developer assigns "content:list" to a role and it does nothing. Dead permissions cause confusion that becomes production incidents.

Resolution: Only "read" exists. RequireResourcePermission maps GET -> resource:read for both collection and single-item endpoints. If read/list distinction is needed in the future, add "list" permission strings and wire individual handlers to use RequirePermission("resource:list") explicitly. This is an additive change that does not break existing roles.

2b. Update chains in internal/middleware/http_chain.go

- DefaultMiddlewareChain(mgr, pc) -- add PermissionInjector(pc) after HTTPAuthenticationMiddleware
- AuthenticatedChain signature unchanged -- does NOT add PermissionInjector
- DefaultMiddlewareChain signature gains *PermissionCache parameter

PermissionInjector is added ONLY to DefaultMiddlewareChain, not to AuthenticatedChain. Since DefaultMiddlewareChain wraps the entire mux (including routes that use authChain internally), all requests pass through PermissionInjector exactly once. Adding it to AuthenticatedChain as well would cause double injection: the first by DefaultMiddlewareChain and the second by AuthenticatedChain, resulting in two redundant cache lookups per authenticated request. The second WithValue overwrites the first (no correctness issue), but it wastes a read-lock acquisition and map lookup on every authenticated request.

PermissionInjector short-circuits for unauthenticated requests (see 2a), so DefaultMiddlewareChain adds negligible overhead to public endpoints (one nil check on the user context value).

2c. Tests: internal/middleware/authorization_test.go

Unit tests for:
- PermissionSet.Has, HasAny, HasAll
- PermissionInjector with authenticated user (PermissionSet injected)
- PermissionInjector with unauthenticated request (no PermissionSet, handler called)
- RequirePermission with matching permission (200)
- RequirePermission with missing permission (403 JSON, no detail field in response body)
- RequireAnyPermission (OR logic)
- RequireAllPermissions (AND logic)
- RequireResourcePermission method mapping (GET->read, POST->create, PUT->update, PATCH->update, DELETE->delete)
- RequireResourcePermission with unmapped method (returns 403)
- Fail-closed: RequirePermission with no PermissionSet in context (returns 403, not 200)
- Fail-closed: RequireResourcePermission with no PermissionSet in context (returns 403)
- Admin bypass via ContextIsAdmin (does not rely on "*" in PermissionSet)
- Load() build-then-swap: concurrent reads during Load() do not block
- Load() safety limit: returns error if role count exceeds 1000
- 403 response body does not contain permission names
- ValidatePermissionLabel: valid labels pass, "*" rejected, empty rejected, missing colon rejected, uppercase rejected

Files modified (Phase 2):

- internal/middleware/authorization.go -- new file
- internal/middleware/http_chain.go -- update DefaultMiddlewareChain, AuthenticatedChain signatures
- internal/middleware/authorization_test.go -- new file

---
Phase 3: Router Wiring

3a. Update NewModulacmsMux signature

func NewModulacmsMux(mgr *config.Manager, bridge *plugin.HTTPBridge, driver db.DbDriver, pc *middleware.PermissionCache) *http.ServeMux

3b. Remove adminOnly and resolveAdminRoleID

Delete the adminOnly closure (lines 246-255) and resolveAdminRoleID function (lines 286-303) from mux.go. The PermissionCache replaces both.

3c. Remove /api/v1/users from PublicEndpoints + fix 401 response information leak

In internal/middleware/http_middleware.go, remove "/api/v1/users" from the PublicEndpoints slice. This endpoint requires authentication and RBAC permission checks. Its presence in PublicEndpoints is a pre-existing security issue that bypasses authentication for user listing and creation.

The PublicEndpoints list should only contain:
- /api/v1/auth/login
- /api/v1/auth/register
- /api/v1/auth/logout
- /api/v1/auth/reset
- /api/v1/auth/me
- /api/v1/auth/oauth/login
- /api/v1/auth/oauth/callback
- /favicon.ico

Additionally, fix the 401 response in HTTPPublicEndpointMiddleware to not leak the request path. The current response uses `fmt.Sprintf("Unauthorized request to %s", r.URL.Path)` which reveals internal URL structure to unauthenticated clients. Change to a generic `{"error": "unauthorized"}` JSON response, consistent with the 403 format. The specific path is already available in server logs via the audit middleware.

3d. Replace admin-only routes with permission checks

authChain := middleware.AuthenticatedChain(mgr, pc)

// Config management
mux.Handle("GET /api/v1/admin/config", authChain(middleware.RequirePermission("config:read")(ConfigGetHandler(mgr))))
mux.Handle("PATCH /api/v1/admin/config", authChain(middleware.RequirePermission("config:update")(ConfigUpdateHandler(mgr))))
mux.Handle("GET /api/v1/admin/config/meta", authChain(middleware.RequirePermission("config:read")(ConfigMetaHandler())))

// Plugin admin (routes registered directly in mux.go)
mux.Handle("GET /api/v1/admin/plugins/routes", authChain(middleware.RequirePermission("plugins:read")(pluginRoutesListHandler(bridge))))
mux.Handle("POST /api/v1/admin/plugins/routes/approve", authChain(middleware.RequirePermission("plugins:admin")(pluginRoutesApproveHandler(bridge))))
mux.Handle("POST /api/v1/admin/plugins/routes/revoke", authChain(middleware.RequirePermission("plugins:admin")(pluginRoutesRevokeHandler(bridge))))

3e. Add permission checks to all resource routes

Wrap existing HandleFunc registrations with RequireResourcePermission. The DefaultMiddlewareChain already includes PermissionInjector, so per-route wrappers just read from context.

Conversion pattern: existing routes use mux.HandleFunc with closures. Adding permission middleware requires converting to mux.Handle with the closure wrapped in http.HandlerFunc and the permission middleware applied outside. Concrete before/after for a typical route:

// Before (no permission check)
mux.HandleFunc("/api/v1/roles", func(w http.ResponseWriter, r *http.Request) {
    RolesHandler(w, r, *c)
})

// After (with RequireResourcePermission + pc parameter for cache invalidation)
mux.Handle("/api/v1/roles", middleware.RequireResourcePermission("roles")(
    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        RolesHandler(w, r, *c, pc)
    }),
))

// After (with RequireResourcePermission, handler signature unchanged)
mux.Handle("/api/v1/contentdata", middleware.RequireResourcePermission("content")(
    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ContentDataHandler(w, r, *c)
    }),
))

Only the 6 handler functions listed in Phase 3k (RolesHandler, RoleHandler, PermissionsHandler, PermissionHandler, RolePermissionsHandler, RolePermissionHandler) gain the `pc` parameter. All other handlers keep their existing signature -- only the mux registration changes from HandleFunc to Handle with the permission wrapper.

For routes using RequirePermission (not RequireResourcePermission), the pattern is the same but with an explicit permission string:

mux.Handle("GET /api/v1/admin/tree/", middleware.RequirePermission("admin_tree:read")(
    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        AdminTreeHandler(w, r, *c)
    }),
))

Complete route-to-permission mapping (every route in mux.go and MountAdminEndpoints):

┌──────────────────────────────────────────────┬────────────────┬──────────────────────────────────────────┐
│ Route                                        │ Method(s)      │ Permission                               │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /favicon.ico                                 │ ALL            │ public (no permission)                    │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ POST /api/v1/auth/login                      │ POST           │ public (no permission)                    │
│ POST /api/v1/auth/logout                     │ POST           │ public (no permission)                    │
│ GET /api/v1/auth/me                          │ GET            │ public (no permission)                    │
│ POST /api/v1/auth/register                   │ POST           │ public (no permission)                    │
│ POST /api/v1/auth/reset                      │ POST           │ public (no permission)                    │
│ GET /api/v1/auth/oauth/login                 │ GET            │ public (no permission)                    │
│ GET /api/v1/auth/oauth/callback              │ GET            │ public (no permission)                    │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/admin/tree/                          │ GET            │ admin_tree:read                           │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/admincontentdatas                    │ GET/POST       │ RequireResourcePermission("content")      │
│ /api/v1/admincontentdatas/                   │ GET/PUT/DELETE │ RequireResourcePermission("content")      │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/admincontentfields                   │ GET/POST       │ RequireResourcePermission("content")      │
│ /api/v1/admincontentfields/                  │ GET/PUT/DELETE │ RequireResourcePermission("content")      │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/admindatatypes                       │ GET/POST       │ RequireResourcePermission("datatypes")    │
│ /api/v1/admindatatypes/                      │ GET/PUT/DELETE │ RequireResourcePermission("datatypes")    │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/adminfields                          │ GET/POST       │ RequireResourcePermission("fields")       │
│ /api/v1/adminfields/                         │ GET/PUT/DELETE │ RequireResourcePermission("fields")       │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/admindatatypefields                  │ GET/POST       │ RequireResourcePermission("fields")       │
│ /api/v1/admindatatypefields/                 │ GET/PUT/DELETE │ RequireResourcePermission("fields")       │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/adminroutes                          │ GET/POST       │ RequireResourcePermission("routes")       │
│ /api/v1/adminroutes/                         │ GET/PUT/DELETE │ RequireResourcePermission("routes")       │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/contentdata                          │ GET/POST       │ RequireResourcePermission("content")      │
│ /api/v1/contentdata/                         │ GET/PUT/DELETE │ RequireResourcePermission("content")      │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/contentfields                        │ GET/POST       │ RequireResourcePermission("content")      │
│ /api/v1/contentfields/                       │ GET/PUT/DELETE │ RequireResourcePermission("content")      │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ POST /api/v1/content/batch                   │ POST           │ content:update                            │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/datatype                             │ GET/POST/DELETE│ RequireResourcePermission("datatypes")    │
│ GET /api/v1/datatype/full                    │ GET            │ datatypes:read                            │
│ /api/v1/datatype/                            │ GET/PUT/DELETE │ RequireResourcePermission("datatypes")    │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/datatypefields                       │ GET/POST       │ RequireResourcePermission("fields")       │
│ /api/v1/datatypefields/                      │ GET/PUT/DELETE │ RequireResourcePermission("fields")       │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/fields                               │ GET/POST       │ RequireResourcePermission("fields")       │
│ /api/v1/fields/                              │ GET/PUT/DELETE │ RequireResourcePermission("fields")       │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/media                                │ GET/POST       │ RequireResourcePermission("media")        │
│ GET /api/v1/media/health                     │ GET            │ media:admin                               │
│ DELETE /api/v1/media/cleanup                 │ DELETE         │ media:admin                               │
│ /api/v1/media/                               │ GET/PUT/DELETE │ RequireResourcePermission("media")        │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/mediadimensions                      │ GET/POST       │ RequireResourcePermission("media")        │
│ /api/v1/mediadimensions/                     │ GET/PUT/DELETE │ RequireResourcePermission("media")        │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/routes                               │ GET/POST       │ RequireResourcePermission("routes")       │
│ /api/v1/routes/                              │ GET/PUT/DELETE │ RequireResourcePermission("routes")       │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/roles                                │ GET/POST       │ RequireResourcePermission("roles")        │
│ /api/v1/roles/                               │ GET/PUT/DELETE │ RequireResourcePermission("roles")        │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/permissions                          │ GET/POST       │ RequireResourcePermission("permissions")  │
│ /api/v1/permissions/                         │ GET/PUT/DELETE │ RequireResourcePermission("permissions")  │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/sessions                             │ GET            │ sessions:read                             │
│ /api/v1/sessions/                            │ GET/DELETE     │ RequireResourcePermission("sessions")     │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/tables                               │ GET/POST       │ RequireResourcePermission("datatypes")    │
│ /api/v1/tables/                              │ GET/PUT/DELETE │ RequireResourcePermission("datatypes")    │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/tokens                               │ GET/POST       │ RequireResourcePermission("tokens")       │
│ /api/v1/tokens/                              │ GET/DELETE     │ RequireResourcePermission("tokens")       │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/usersoauth                           │ GET/POST       │ RequireResourcePermission("users")        │
│ /api/v1/usersoauth/                          │ GET/PUT/DELETE │ RequireResourcePermission("users")        │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/users                                │ GET/POST       │ RequireResourcePermission("users")        │
│ GET /api/v1/users/full                       │ GET            │ users:read                                │
│ GET /api/v1/users/full/                      │ GET            │ users:read                                │
│ /api/v1/users/                               │ GET/PUT/DELETE │ RequireResourcePermission("users")        │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ POST /api/v1/ssh-keys                        │ POST           │ ssh_keys:create                           │
│ GET /api/v1/ssh-keys                         │ GET            │ ssh_keys:read                             │
│ DELETE /api/v1/ssh-keys/                     │ DELETE         │ ssh_keys:delete                           │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ /api/v1/import/contentful                    │ POST           │ import:create                             │
│ /api/v1/import/sanity                        │ POST           │ import:create                             │
│ /api/v1/import/strapi                        │ POST           │ import:create                             │
│ /api/v1/import/wordpress                     │ POST           │ import:create                             │
│ /api/v1/import/clean                         │ POST           │ import:create                             │
│ /api/v1/import                               │ POST           │ import:create                             │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ GET /api/v1/admin/config                     │ GET            │ config:read                               │
│ PATCH /api/v1/admin/config                   │ PATCH          │ config:update                             │
│ GET /api/v1/admin/config/meta                │ GET            │ config:read                               │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ GET /api/v1/admin/plugins/routes             │ GET            │ plugins:read                              │
│ POST /api/v1/admin/plugins/routes/approve    │ POST           │ plugins:admin                             │
│ POST /api/v1/admin/plugins/routes/revoke     │ POST           │ plugins:admin                             │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ MountAdminEndpoints (plugin/http_bridge.go): │                │                                          │
│ GET /api/v1/admin/plugins                    │ GET            │ plugins:read                              │
│ GET /api/v1/admin/plugins/cleanup            │ GET            │ plugins:admin                             │
│ POST /api/v1/admin/plugins/cleanup           │ POST           │ plugins:admin                             │
│ GET /api/v1/admin/plugins/hooks              │ GET            │ plugins:read                              │
│ POST /api/v1/admin/plugins/hooks/approve     │ POST           │ plugins:admin                             │
│ POST /api/v1/admin/plugins/hooks/revoke      │ POST           │ plugins:admin                             │
│ GET /api/v1/admin/plugins/{name}             │ GET            │ plugins:read                              │
│ POST /api/v1/admin/plugins/{name}/reload     │ POST           │ plugins:admin                             │
│ POST /api/v1/admin/plugins/{name}/enable     │ POST           │ plugins:admin                             │
│ POST /api/v1/admin/plugins/{name}/disable    │ POST           │ plugins:admin                             │
├──────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────┤
│ / (slug handler)                             │ GET            │ public (no permission)                    │
└──────────────────────────────────────────────┴────────────────┴──────────────────────────────────────────┘

Route-to-resource mapping rationale for non-obvious cases:
- /api/v1/tables -> datatypes: tables are the underlying storage for datatypes, same authorization scope
- /api/v1/usersoauth -> users: OAuth connections are a sub-resource of users
- /api/v1/mediadimensions -> media: dimensions are a sub-resource of media
- /api/v1/contentfields, /api/v1/admincontentfields -> content: content field values are part of content records
- /api/v1/datatypefields, /api/v1/admindatatypefields -> fields: these are field definitions scoped to a datatype
- /api/v1/media/health, /api/v1/media/cleanup -> media:admin: operational/maintenance operations, not regular CRUD
- /api/v1/datatype accepts DELETE on collection endpoint (handler supports it) -- mapped via RequireResourcePermission

MountAdminEndpoints permission changes:
- GET /api/v1/admin/plugins (plugin list) -- currently auth-only, adding plugins:read. Plugin metadata (names, config, hooks) should not be visible to all authenticated users.
- GET /api/v1/admin/plugins/hooks (hooks list) -- currently auth-only, adding plugins:read.
- GET /api/v1/admin/plugins/{name} (plugin info) -- currently auth-only, adding plugins:read.
- All mutation endpoints already use adminOnlyFn, which becomes RequirePermission("plugins:admin").

3f. Update MountAdminEndpoints to use permission checks

The current MountAdminEndpoints signature is:

func (b *HTTPBridge) MountAdminEndpoints(
    mux *http.ServeMux,
    authChain func(http.Handler) http.Handler,
    adminOnlyFn func(http.Handler) http.Handler,
)

The `adminOnlyFn` parameter is replaced with a `permFn` parameter that constructs permission-checked middleware. The three read-only endpoints (plugin list, hooks list, plugin info) that currently use only `authChain` gain a `plugins:read` permission check. The mutation endpoints that currently use `adminOnlyFn` switch to `plugins:admin`.

Updated signature:

func (b *HTTPBridge) MountAdminEndpoints(
    mux *http.ServeMux,
    authChain func(http.Handler) http.Handler,
    readPerm func(http.Handler) http.Handler,   // middleware.RequirePermission("plugins:read")
    adminPerm func(http.Handler) http.Handler,  // middleware.RequirePermission("plugins:admin")
)

Updated endpoint registrations:

mux.Handle("GET /api/v1/admin/plugins",
    authChain(readPerm(PluginListHandler(mgr))))
mux.Handle("GET /api/v1/admin/plugins/hooks",
    authChain(readPerm(PluginHooksListHandler(mgr))))
mux.Handle("GET /api/v1/admin/plugins/{name}",
    authChain(readPerm(PluginInfoHandler(mgr))))

All mutation endpoints use adminPerm instead of adminOnlyFn:

mux.Handle("GET /api/v1/admin/plugins/cleanup",
    authChain(adminPerm(PluginCleanupListHandler(mgr))))
// ... etc for all other mutation endpoints

Note: This tightens access on the three read-only endpoints. Previously any authenticated user could access them; now `plugins:read` is required. The editor role does not have `plugins:read` by default, so editors lose access to plugin listing. This is intentional -- plugin metadata (names, config, hooks) should not be visible to all authenticated users.

The caller in mux.go passes the concrete middleware:

bridge.MountAdminEndpoints(mux, authChain,
    middleware.RequirePermission("plugins:read"),
    middleware.RequirePermission("plugins:admin"),
)

3g. Update cmd/serve.go

- Create PermissionCache after loading config/driver
- Call pc.Load(driver) AFTER install.CheckInstall (which runs CreateBootstrapData on fresh installs). The ordering in serve.go must be: loadConfigAndDB -> install.CheckInstall -> pc.Load(driver). If Load() runs before CreateBootstrapData, the cache is populated with zero roles and zero permissions. With the fail-closed invariant (see Phase 2a), this would deny all requests until the periodic refresh fires 60 seconds later. By loading after bootstrap completes, the cache is guaranteed to have the 3 system roles and 77 junction rows on first load.
- Call pc.Load(driver) to populate cache. If Load() returns an error, log it and exit with non-zero status. The server MUST NOT start with an empty permission cache -- this would either deny all requests (if the middleware treats missing cache as unauthorized) or allow all requests (if it treats missing cache as permissive). Both are unacceptable. A startup Load() failure indicates a database or schema problem that must be resolved before the server can serve traffic.
- Call pc.StartPeriodicRefresh(ctx, driver, 60*time.Second) using the server's shutdown context
- Pass pc to NewModulacmsMux and chain constructors
- Update DefaultMiddlewareChain(mgr) call to DefaultMiddlewareChain(mgr, pc)

3h. System-protected deletion and mutation guards

In the role and permission delete handlers (internal/router/roles.go apiDeleteRole, internal/router/permissions.go apiDeletePermission), add a check before the DB call:

1. Fetch the role/permission by ID
2. If system_protected == 1, return 403 JSON: {"error": "forbidden", "detail": "cannot delete system-protected record"}
3. Otherwise proceed with deletion

In the role and permission update handlers (internal/router/roles.go apiUpdateRole, internal/router/permissions.go apiUpdatePermission), add a check:

1. Fetch the role/permission by ID
2. If system_protected == 1 AND the request changes the `label` field, return 403 JSON: {"error": "forbidden", "detail": "cannot rename system-protected record"}
3. Other field updates on system_protected records are allowed (e.g. updating description)

The label mutation guard is critical for the admin role: the PermissionCache identifies the admin role by `role.Label == "admin"` (see Phase 2a). If an admin user renamed the admin role, the cache would rebuild with no admin detected, locking all users out of admin-only operations. Similarly, renaming a bootstrap permission label would break any RequirePermission() middleware referencing the old label string.

This prevents deletion of the admin, editor, and viewer roles, and all 47 bootstrap permissions via the API. Custom roles and permissions created after bootstrap have system_protected = 0 and can be freely deleted and renamed.

3i. Permission label validation

In the permission create and update handlers (internal/router/permissions.go apiCreatePermission, apiUpdatePermission), validate that the label field:
1. Is not the literal string "*"
2. Matches the format: lowercase alphanumeric/underscores, colon, lowercase alphabetic (e.g. "datatypes:create", "media:admin")
3. Is not empty

Reject with 400 JSON if validation fails: {"error": "invalid permission label"}

Implementation: use a ValidatePermissionLabel(label string) error function with character-by-character validation (no regex). The function splits on ":" to get exactly two parts, then validates the resource part contains only [a-z0-9_] and the operation part contains only [a-z], using byte-range comparisons in a loop. This is consistent with the project's no-regex-for-modifications rule and is straightforward to test.

This prevents creating permissions that could collide with the (now removed) wildcard bypass, and ensures consistent permission naming.

3j. Role assignment validation in user handlers

In internal/router/users.go, add field-level authorization to ApiCreateUser and ApiUpdateUser:

When the request body contains a `role` field:
1. If the caller is admin (ContextIsAdmin), allow any role assignment.
2. If the caller is not admin and the role value differs from the user's current role, reject with 403: {"error": "forbidden", "detail": "only administrators can assign roles"}

This prevents privilege escalation where an editor with users:create or users:update assigns the admin role to a user. The RBAC middleware gates access to the endpoint, but without this check, the role field within the request body is unprotected.

Implementation for ApiUpdateUser: after decoding the request body and before calling the DB, fetch the existing user and compare role values. This allows standard REST clients to send the full object (including the unchanged role field) without triggering a false 403:

existing, err := db.ConfigDB(c).GetUser(r.Context(), userID)
if err != nil {
    // handle error
}
if req.Role != "" && req.Role != string(existing.RoleID) && !middleware.ContextIsAdmin(r.Context()) {
    // Non-admin attempting to change a role
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusForbidden)
    json.NewEncoder(w).Encode(map[string]string{
        "error":  "forbidden",
        "detail": "only administrators can assign roles",
    })
    return nil
}

Implementation for ApiCreateUser: there is no existing user to compare against, so any non-empty role field from a non-admin caller is rejected. Non-admin callers who need to create users must omit the role field, which assigns the default role (viewer). This is acceptable because user creation by non-admins is rare (most CMS deployments restrict users:create to admins):

if req.Role != "" && !middleware.ContextIsAdmin(r.Context()) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusForbidden)
    json.NewEncoder(w).Encode(map[string]string{
        "error":  "forbidden",
        "detail": "only administrators can assign roles",
    })
    return nil
}
// If role is empty and caller is not admin, assign viewer role
if req.Role == "" {
    viewerRole, err := db.ConfigDB(c).GetRoleByLabel("viewer")
    if err != nil {
        // handle error -- viewer role must exist (system-protected bootstrap)
    }
    req.Role = string(viewerRole.RoleID)
}

Default role assignment: when role is omitted (empty string) on user creation, the viewer role is assigned. This applies to both admin and non-admin callers. Admins can override by providing an explicit role value. The viewer role is system-protected and always exists after bootstrap.

Note: the GetRoleByLabel method does not currently exist on the DbDriver interface. It must be added as part of this phase:
- GetRoleByLabel(label string) (*Role, error) -- queries roles WHERE label = ?
- Add to DbDriver interface and implement on all three wrappers
- The query already has UNIQUE on label, so this returns at most one row

Registration flow (public /api/v1/auth/register): the existing registration handler must also assign the viewer role by default. Registration is unauthenticated, so the role field in the registration request body must be ignored entirely -- the handler hardcodes viewer assignment regardless of what the client sends. This prevents privilege escalation via self-registration.

3k. Cache invalidation

After role/permission mutation handlers, call pc.Load(driver) to refresh the in-memory cache. Since Load() uses build-then-swap (see Phase 2), this does not block in-flight requests. The periodic refresh (every 60 seconds) provides a bounded staleness guarantee as a safety net.

Mutations that trigger cache refresh:
- CreateRole, UpdateRole, DeleteRole (role changes)
- CreateRolePermission, DeleteRolePermission, DeleteRolePermissionsByRoleID (junction changes)
- CreatePermission, UpdatePermission, DeletePermission (permission label changes affect cached strings)

Dependency injection approach: `pc *middleware.PermissionCache` and `driver db.DbDriver` are both available in the NewModulacmsMux scope (see 3a signature). The handlers that perform these mutations need to trigger a cache refresh after a successful DB mutation -- not after every request (a 400 validation error should not trigger a cache refresh).

Cache refresh is asynchronous (fire-and-forget in a goroutine) to avoid blocking the HTTP response on a full cache rebuild. Since Load() uses build-then-swap, a concurrent refresh cannot corrupt the cache. If the async refresh fails, the periodic refresh (every 60 seconds) provides a bounded staleness safety net.

Add `pc *middleware.PermissionCache` as a parameter to the specific handler functions that perform role/permission mutations:

- RolesHandler / RoleHandler -> add pc parameter, trigger async refresh after successful create/update/delete
- PermissionsHandler / PermissionHandler -> add pc parameter, trigger async refresh after successful create/update/delete
- RolePermissionsHandler / RolePermissionHandler -> new handlers (see 3l), accept pc parameter

The driver is already available inside these handlers via db.ConfigDB(c). The mux closure passes pc:

    mux.HandleFunc("POST /api/v1/roles", func(w http.ResponseWriter, r *http.Request) {
        RolesHandler(w, r, *c, pc)
    })

Inside the handler, after a successful mutation:

    go func() {
        if err := pc.Load(db.ConfigDB(c)); err != nil {
            utility.DefaultLogger.Error("permission cache refresh failed", "error", err)
        }
    }()

Only 3 handler pairs (6 functions total) need the pc parameter. Other handlers are unaffected.

If Load() fails, the error is logged but the mutation response is unaffected -- the cache is stale but the mutation succeeded. The periodic refresh will fix it within 60 seconds.

3l. Role permissions API endpoints

Add CRUD endpoints for managing the role_permissions junction table. Without these endpoints, there is no way to assign or revoke permissions to/from roles via the API.

New file: internal/router/role_permissions.go

Handlers:
- RolePermissionsHandler (collection): GET (list all), POST (create)
- RolePermissionHandler (single): GET (read), DELETE (delete)
- RolePermissionsByRoleHandler: GET /api/v1/role-permissions/role/{id} (list by role)

Route registrations in mux.go:

┌────────────────────────────────────────────────┬────────────────┬──────────────────────────────────────────────┐
│ Route                                          │ Method(s)      │ Permission                                    │
├────────────────────────────────────────────────┼────────────────┼──────────────────────────────────────────────┤
│ /api/v1/role-permissions                       │ GET/POST       │ RequireResourcePermission("roles")             │
│ /api/v1/role-permissions/                      │ GET/DELETE     │ RequireResourcePermission("roles")             │
│ /api/v1/role-permissions/role/                 │ GET            │ roles:read                                     │
└────────────────────────────────────────────────┴────────────────┴──────────────────────────────────────────────┘

Role permission management is scoped under the "roles" resource permission because assigning permissions to a role is an operation on the role itself. An editor with roles:read can view role-permission assignments but cannot modify them (requires roles:create for POST, roles:delete for DELETE via RequireResourcePermission).

The POST handler accepts `{role_id, permission_id}` and creates a junction row. The DELETE handler removes a junction row by its ID. Both trigger cache invalidation (see 3k).

System-protected junction row guard: the DELETE handler and DeleteRolePermissionsByRoleID must check whether the junction row belongs to a system-protected role before deleting. Fetch the role by the junction row's role_id; if system_protected is true, return 403 JSON: {"error": "forbidden", "detail": "cannot modify permissions on system-protected role"}. This prevents an admin from accidentally stripping all permissions from the admin, editor, or viewer roles via the API. The POST handler does NOT need this guard -- adding extra permissions to system-protected roles is safe and sometimes useful (e.g., granting the editor role a new custom permission).

Handler modification summary (Phase 3):

This table lists every handler function modified in Phase 3 and which sub-phases touch it, to prevent implementing agents from producing code that gets overwritten by a later sub-phase. Implement all changes to a given handler in a single pass.

┌──────────────────────────────────┬──────────────────────┬──────────────────┬─────────────────────────────┬──────────────────────────────┐
│ Handler function                 │ 3h: system_protected │ 3i: label valid. │ 3j: role assign valid.      │ 3k: pc param + cache refresh │
│                                  │ guard                │                  │                             │                              │
├──────────────────────────────────┼──────────────────────┼──────────────────┼─────────────────────────────┼──────────────────────────────┤
│ apiDeleteRole (roles.go)         │ Yes (block delete)   │ --               │ --                          │ Yes                          │
│ apiUpdateRole (roles.go)         │ Yes (block label)    │ --               │ --                          │ Yes                          │
│ apiCreateRole (roles.go)         │ --                   │ --               │ --                          │ Yes                          │
│ apiDeletePermission (perms.go)   │ Yes (block delete)   │ --               │ --                          │ Yes                          │
│ apiUpdatePermission (perms.go)   │ Yes (block label)    │ Yes              │ --                          │ Yes                          │
│ apiCreatePermission (perms.go)   │ --                   │ Yes              │ --                          │ Yes                          │
│ ApiCreateUser (users.go)         │ --                   │ --               │ Yes (reject + default role) │ --                           │
│ ApiUpdateUser (users.go)         │ --                   │ --               │ Yes (compare existing role) │ --                           │
│ RolePermissionsHandler (new)     │ --                   │ --               │ --                          │ Yes                          │
│ RolePermissionHandler (new)      │ junction row guard   │ --               │ --                          │ Yes                          │
│ registration handler (auth.go)   │ --                   │ --               │ Yes (ignore role field)     │ --                           │
└──────────────────────────────────┴──────────────────────┴──────────────────┴─────────────────────────────┴──────────────────────────────┘

Files modified (Phase 3):

- internal/middleware/http_middleware.go -- remove "/api/v1/users" from PublicEndpoints, fix 401 response path leak
- internal/router/mux.go -- signature change, remove adminOnly/resolveAdminRoleID, add RequirePermission wrappers to every route (HandleFunc -> Handle conversion), add role-permissions routes
- internal/plugin/http_bridge.go -- update MountAdminEndpoints signature (replace adminOnlyFn with readPerm + adminPerm), add RequirePermission("plugins:read") to list/info endpoints
- internal/router/roles.go -- add system_protected deletion and label mutation guards, add pc parameter for cache invalidation
- internal/router/permissions.go -- add system_protected deletion and label mutation guards, add label validation, add pc parameter for cache invalidation
- internal/router/role_permissions.go -- new file, CRUD handlers with junction row system-protected guard and cache invalidation
- internal/router/users.go -- add role assignment validation to create/update handlers (compare against existing role for updates, reject non-empty role from non-admins for creates, default to viewer role)
- internal/router/auth.go -- ignore role field in registration request body, hardcode viewer role assignment for self-registration
- internal/db/db.go -- add GetRoleByLabel(label string) (*Role, error) to DbDriver interface, implement on all three wrappers
- cmd/serve.go -- create PermissionCache, call StartPeriodicRefresh (fatal on failure), pass to mux and middleware

---
Phase 4 (future): TUI Authorization

The SSH/TUI is out of scope for this plan. Currently the TUI is effectively admin-only: it requires SSH key authentication, and the actions.go admin check (line 410-417) gates sensitive operations by checking for the admin role.

For now, the TUI remains unchanged. RBAC applies only to the HTTP API surface. The TUI will be addressed in a separate plan that integrates PermissionCache lookups into the Bubbletea model, gating TUI actions based on the authenticated SSH user's role permissions.

This is acceptable because:
- SSH access already requires key-based authentication (higher trust boundary)
- The existing admin role check in actions.go prevents non-admin users from performing destructive operations
- The TUI is used by CMS administrators, not end users

---
Phase 5 (future): Object-Level Authorization (IDOR Prevention)

This plan provides resource-level authorization (can this user access the "users" resource?) but intentionally defers object-level authorization (can this user access THIS specific user record?). This is a known gap documented here for future work.

Affected resources where object-level checks matter:
- users: an editor with users:update can currently update any user, including admins
- sessions: a user with sessions:delete can delete any user's sessions
- ssh_keys: a user with ssh_keys:delete can delete any user's SSH keys
- tokens: a user with tokens:delete can delete any user's tokens

The role assignment validation in Phase 3j mitigates the most dangerous IDOR scenario (privilege escalation via role assignment). Full object-level authorization will be addressed in a separate plan with handler-level ownership checks.

---
Verification

1. just sqlc -- regenerate code (no errors)
2. just check -- compile check
3. just test -- all existing tests pass (bootstrap creates junction rows, test DB is created fresh)
4. Manual test: start server, login as admin -- all routes work (admin bypass)
5. Manual test: create editor user, login -- CRUD on content/datatypes/fields/media/routes works, write to config/plugins returns 403
6. Manual test: create viewer user, login -- only GET on content/datatypes/fields/media/routes returns 200, POST/PUT/DELETE returns 403
7. Manual test: modify role permissions via /api/v1/role-permissions, verify cache refreshes without blocking other requests
8. Manual test: hit /api/v1/permissions with POST/PUT/DELETE as editor -- returns 403 (permissions CRUD requires permissions:create/update/delete which editor does not have)
9. Manual test: hit /api/v1/media/health and /api/v1/media/cleanup as editor -- returns 403 (requires media:admin)
10. Manual test: as editor, POST /api/v1/users with role set to admin role ID -- returns 403 (role assignment blocked)
11. Manual test: as editor, PUT /api/v1/users/ to change own role to admin -- returns 403 (role assignment blocked)
12. Manual test: DELETE /api/v1/roles/ with admin role ID -- returns 403 (system-protected)
13. Manual test: DELETE /api/v1/permissions/ with bootstrap permission ID -- returns 403 (system-protected)
14. Manual test: PUT /api/v1/roles/ to rename admin role -- returns 403 (label mutation blocked on system-protected)
15. Manual test: POST /api/v1/permissions with label "*" -- returns 400 (invalid label)
16. Manual test: verify 403 response body contains {"error": "forbidden"} with no permission detail
17. Manual test: verify 401 response body contains {"error": "unauthorized"} with no path detail
18. Manual test: verify /api/v1/users without auth returns 401 (no longer in PublicEndpoints)
19. Manual test: verify /api/v1/admin/plugins without plugins:read permission returns 403
20. Manual test: stop cache refresh, mutate permissions, wait 60s, verify periodic refresh catches up
21. Manual test: send PATCH to a RequireResourcePermission route -- maps to update permission
22. Manual test: kill database, verify server continues serving with stale cache and logs errors (first 3, then dampened)
23. Manual test: restart database, verify next periodic refresh restores fresh cache and logs recovery message
24. Manual test: DELETE /api/v1/role-permissions/{id} where junction row belongs to admin role -- returns 403 (system-protected junction guard)
25. Manual test: POST /api/v1/role-permissions with system-protected role_id -- succeeds (adding permissions to system roles is allowed)
26. Manual test: POST /api/v1/auth/register with role field set to admin role ID -- user created with viewer role (role field ignored)
27. Manual test: POST /api/v1/users as non-admin with no role field -- user created with viewer role (default assignment)
28. Manual test: PUT /api/v1/users/{id} as non-admin, echoing back unchanged role in request body -- succeeds (no false 403)
29. Manual test: fresh install -- verify pc.Load runs after CreateBootstrapData (cache has 3 roles, 77 junction rows)

Security audit checklist:
- [ ] No endpoint in mux.go or MountAdminEndpoints is missing a permission check
- [ ] PublicEndpoints list contains only auth and favicon endpoints
- [ ] 403 responses do not leak permission names to clients
- [ ] 401 responses do not leak request paths to clients
- [ ] System roles and permissions cannot be deleted via API
- [ ] System role and permission labels cannot be renamed via API
- [ ] System-protected role junction rows cannot be deleted via API (POST allowed, DELETE blocked)
- [ ] Permission labels are validated on create/update (no "*", must match format via character validation)
- [ ] Role assignment in user create/update is admin-only (compare against existing role for updates)
- [ ] User creation defaults to viewer role when role field is omitted
- [ ] Self-registration ignores role field entirely, assigns viewer
- [ ] Admin bypass uses IsAdmin boolean, not "*" in PermissionSet
- [ ] Admin role identified by label "admin" with UNIQUE constraint on roles.label as security dependency
- [ ] Fail-closed: all Require* middlewares return 403 when PermissionSet is missing from context
- [ ] PermissionInjector added to DefaultMiddlewareChain only (not duplicated in AuthenticatedChain)
- [ ] Cache API uses types.RoleID (not plain string) for type safety
- [ ] Cache refresh has bounded staleness via periodic goroutine with log dampening
- [ ] Startup Load() runs after CreateBootstrapData, not before
- [ ] Startup Load() failure is fatal (server does not start)
- [ ] Authorization failures are logged with structured fields
- [ ] Role permissions manageable via /api/v1/role-permissions endpoints
- [ ] SDK types tolerate new system_protected field (unknown field ignored)
