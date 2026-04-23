# Separate Admin Resource Permissions

## Context

Admin-prefixed resources (admin content, admin datatypes, admin fields, admin media, admin routes) share permissions with their public counterparts. A client using MCP or the API to manage their website should not be able to modify the admin panel structure, which is managed by the agency. Five admin resources need separate permission namespaces. Three already have them (admin_tree, admin_field_types, admin_validations).

**Editor role rationale:** Editors receive full CRUD on admin resources because editors are agency staff who build and maintain the admin panel. Viewers (client staff) get no admin resource permissions. The permission separation isolates API key and MCP access, not agency staff from each other.

## Dependency

None. This plan is independent of the MCP phantom permission plan. Either can land first.

## Changes

### 1. Bootstrap data: internal/db/db.go

**This file contains three identical copies of `CreateBootstrapData` for the tri-database pattern. All three MUST be updated identically:**

- `func (d Database) CreateBootstrapData` (SQLite, line 381)
- `func (d MysqlDatabase) CreateBootstrapData` (MySQL, line 1556)
- `func (d PsqlDatabase) CreateBootstrapData` (PostgreSQL, line 2691)

In each copy, apply the following changes:

**a) Add 26 new permissions to `rbacPermissionLabels`** (5 resources x 5 actions + admin_content:publish):

```
"admin_content:read", "admin_content:create", "admin_content:update", "admin_content:delete", "admin_content:publish", "admin_content:admin",
"admin_datatypes:read", "admin_datatypes:create", "admin_datatypes:update", "admin_datatypes:delete", "admin_datatypes:admin",
"admin_fields:read", "admin_fields:create", "admin_fields:update", "admin_fields:delete", "admin_fields:admin",
"admin_media:read", "admin_media:create", "admin_media:update", "admin_media:delete", "admin_media:admin",
"admin_routes:read", "admin_routes:create", "admin_routes:update", "admin_routes:delete", "admin_routes:admin",
```

Place after the existing `admin_validations:*` line in each copy.

**b) Add to `editorPermLabels`** in each copy (CRUD + publish for admin_content, CRUD for the other four):

```
"admin_content:read", "admin_content:create", "admin_content:update", "admin_content:delete", "admin_content:publish",  // publish is unique to admin_content
"admin_datatypes:read", "admin_datatypes:create", "admin_datatypes:update", "admin_datatypes:delete",
"admin_fields:read", "admin_fields:create", "admin_fields:update", "admin_fields:delete",
"admin_media:read", "admin_media:create", "admin_media:update", "admin_media:delete",
"admin_routes:read", "admin_routes:create", "admin_routes:update", "admin_routes:delete",
```

**c) Viewer role:** No admin resource permissions. Do not add any `admin_content:*`, `admin_datatypes:*`, `admin_fields:*`, `admin_media:*`, or `admin_routes:*` labels to `viewerPermLabels` in any copy.

**d) Do NOT add `content:publish` to `editorPermLabels`.** Editors intentionally lack public content publish permission. Only `admin_content:publish` is added. Do not modify the existing editor permissions; only append the new admin resource permissions listed above.

### 2. Ensure function for existing installs: internal/db/ensure.go

New function `EnsureAdminResourcePermissions` following the `EnsureLocalePermissions` pattern (line 239). Creates the 26 permissions if missing and assigns them to admin and editor roles. Idempotent, no-op on fresh installs where `CreateBootstrapData` already includes them.

**Duplicate handling:** The function checks if each permission label exists before creating it. If the permission already exists, the entire block (permission creation + role-permission assignment) is skipped. Role-permission assignment only runs for newly created permissions, so the UNIQUE constraint on `(role_id, permission_id)` in `role_permissions` is never hit. This matches the existing `EnsureLocalePermissions` and `EnsureWebhookPermissions` patterns.

Permission-to-role assignments:

| Permission | admin | editor |
|---|---|---|
| admin_content:read | yes | yes |
| admin_content:create | yes | yes |
| admin_content:update | yes | yes |
| admin_content:delete | yes | yes |
| admin_content:publish | yes | yes |
| admin_content:admin | yes | no |
| admin_datatypes:read | yes | yes |
| admin_datatypes:create | yes | yes |
| admin_datatypes:update | yes | yes |
| admin_datatypes:delete | yes | yes |
| admin_datatypes:admin | yes | no |
| admin_fields:read | yes | yes |
| admin_fields:create | yes | yes |
| admin_fields:update | yes | yes |
| admin_fields:delete | yes | yes |
| admin_fields:admin | yes | no |
| admin_media:read | yes | yes |
| admin_media:create | yes | yes |
| admin_media:update | yes | yes |
| admin_media:delete | yes | yes |
| admin_media:admin | yes | no |
| admin_routes:read | yes | yes |
| admin_routes:create | yes | yes |
| admin_routes:update | yes | yes |
| admin_routes:delete | yes | yes |
| admin_routes:admin | yes | no |

### 3. Boot call: cmd/serve.go

Add after the closing brace of the `EnsureWebhookPermissions` block (line 198), before the `cfg, err := mgr.Config()` call (line 200):

```go
// Ensure admin resource permissions exist (backfill for upgrades).
if ensureErr := db.EnsureAdminResourcePermissions(rootCtx, driver); ensureErr != nil {
    utility.DefaultLogger.Warn("ensureAdminResourcePermissions failed", ensureErr)
}
```

### 4. API router guards: internal/router/mux.go

Change permission labels on admin-prefixed API endpoints. The `admin/content/heal` endpoint (line 268) stays on `content:update` because it operates on the public content tree despite the admin URL prefix.

**Admin content data** (lines 99, 102):
- `"content"` -> `"admin_content"` (RequireResourcePermission)

**Admin content data read** (line 105):
- `"content:read"` -> `"admin_content:read"` (RequirePermission)

**Admin content fields** (lines 110-114):
- `"content"` -> `"admin_content"` (RequireResourcePermission)

**Admin content data reorder** (line 263):
- `"content:update"` -> `"admin_content:update"` (RequirePermission)

**Admin content data move** (line 278):
- `"content:update"` -> `"admin_content:update"` (RequirePermission)

**Admin content translations** (line 691):
- `"content:create"` -> `"admin_content:create"` (RequirePermission)

Line 688 (`POST /api/v1/admin/contentdata/{id}/translations`) stays as `"content:create"` because it operates on public content.

**Admin datatypes** (lines 118-131):
- `"datatypes"` -> `"admin_datatypes"` (RequireResourcePermission)
- `"datatypes:read"` -> `"admin_datatypes:read"` (RequirePermission)
- `"datatypes:update"` -> `"admin_datatypes:update"` (RequirePermission)

**Admin fields** (lines 135-139):
- `"fields"` -> `"admin_fields"` (RequireResourcePermission)

**Admin routes** (lines 143-147):
- `"routes"` -> `"admin_routes"` (RequireResourcePermission)

**Admin publish/unpublish/schedule** (lines 228-234):
- `"content:publish"` -> `"admin_content:publish"` (RequirePermission) on all three endpoints

**Admin versions** (lines 239-253):
- `"content:read"` -> `"admin_content:read"` on GET endpoints (lines 239, 242)
- `"content:update"` -> `"admin_content:update"` on POST endpoints (lines 245, 253)
- `"content:delete"` -> `"admin_content:delete"` on DELETE endpoint (line 248)

**Admin media** (lines 401-410):
- `"media:read"` -> `"admin_media:read"` (RequirePermission)
- `"media:update"` -> `"admin_media:update"` (RequirePermission)
- `"media"` -> `"admin_media"` (RequireResourcePermission)

**Admin media folders** (lines 414-435):
- All `"media:read"` -> `"admin_media:read"`
- All `"media:create"` -> `"admin_media:create"`
- All `"media:update"` -> `"admin_media:update"`
- All `"media:delete"` -> `"admin_media:delete"`

### 5. Admin panel (HTMX) route guards: internal/router/mux.go

These are in `registerAdminRoutes`. Only admin-prefixed panel routes change. Public panel routes (`/admin/content`, `/admin/datatypes`, etc.) stay as-is because those manage public resources.

**Admin media panel routes** (lines 940-951):
- All `"media:*"` -> `"admin_media:*"` for `/admin/admin-media*` and `/admin/admin-media-folders*` paths

**Admin datatypes panel routes** (lines 969-976):
- `viewing("datatypes", ...)` -> `viewing("admin_datatypes", ...)`
- `mutating("datatypes:*", ...)` -> `mutating("admin_datatypes:*", ...)`
- `mutating("fields:create", ...)` -> `mutating("admin_fields:create", ...)` (line 975, admin datatype field creation)
- `mutating("fields:update", ...)` -> `mutating("admin_fields:update", ...)` (line 976, admin datatype field reorder)

**Admin fields panel routes** (lines 979-981):
- `viewing("fields", ...)` -> `viewing("admin_fields", ...)`
- `mutating("fields:*", ...)` -> `mutating("admin_fields:*", ...)`

**Admin routes list page** (line 962):
- `viewing("routes", adminhandlers.AdminRoutesListHandler(...))` -> `viewing("admin_routes", adminhandlers.AdminRoutesListHandler(...))`

This is the `GET /admin/routes/admin` page that lists admin routes via `AdminRoutesListHandler`. It must use `admin_routes:read` because it displays admin route data. Lines 961, 963-966 manage public routes (despite appearing in the same section). Do NOT change their permission labels.

**Admin routes panel routes** (lines 998-1001):
- `viewing("routes", ...)` -> `viewing("admin_routes", ...)`
- `mutating("routes:*", ...)` -> `mutating("admin_routes:*", ...)`

**Admin content panel routes** (lines 1004-1011):
- `viewing("content", ...)` -> `viewing("admin_content", ...)`
- `mutating("content:*", ...)` -> `mutating("admin_content:*", ...)`
- `mutating("content:publish", ...)` -> `mutating("admin_content:publish", ...)` (lines 1009-1010)

### 6. Admin panel sidebar navigation: internal/admin/components/nav.go

The sidebar `NavItems` slice (lines 37-42) contains hardcoded permission strings that control which navigation links are visible. Update the admin panel section items:

| Line | Label | Current Permission | New Permission |
|---|---|---|---|
| 37 | Admin Content | `content:read` | `admin_content:read` |
| 38 | Admin Datatypes | `datatypes:read` | `admin_datatypes:read` |
| 39 | Admin Field Types | `field_types:read` | `admin_field_types:read` (opportunistic fix: `admin_field_types:read` already exists in bootstrap but the nav item was never updated) |
| 41 | Admin Routes | `routes:read` | `admin_routes:read` |
| 42 | Admin Media | `media:read` | `admin_media:read` |

Line 40 (Admin Validations) already uses `admin_validations:read`. Do not change it.

Without this change, the sidebar would show admin panel links to users who have public resource permissions (e.g., `content:read`) but not the corresponding admin permissions, and hide them from users who have admin permissions but not public ones.

### 7. MCP permission map: internal/mcp/permissions.go

Update all admin-prefixed tool mappings. Each tool name below must be changed to the new permission label. The `get_admin_tree` tool stays as `admin_tree:read` (already correct).

**Admin content tools** (lines 47-60):

| Tool name | Current | New |
|---|---|---|
| `admin_list_content` | `content:read` | `admin_content:read` |
| `admin_get_content` | `content:read` | `admin_content:read` |
| `admin_create_content` | `content:create` | `admin_content:create` |
| `admin_update_content` | `content:update` | `admin_content:update` |
| `admin_delete_content` | `content:delete` | `admin_content:delete` |
| `admin_reorder_content` | `content:update` | `admin_content:update` |
| `admin_move_content` | `content:update` | `admin_content:update` |
| `admin_get_content_full` | `content:read` | `admin_content:read` |
| `admin_list_content_fields` | `content:read` | `admin_content:read` |
| `admin_get_content_field` | `content:read` | `admin_content:read` |
| `admin_create_content_field` | `content:create` | `admin_content:create` |
| `admin_update_content_field` | `content:update` | `admin_content:update` |
| `admin_delete_content_field` | `content:delete` | `admin_content:delete` |

**Admin schema tools** (lines 86-97):

| Tool name | Current | New |
|---|---|---|
| `admin_list_datatypes` | `datatypes:read` | `admin_datatypes:read` |
| `admin_get_datatype` | `datatypes:read` | `admin_datatypes:read` |
| `admin_create_datatype` | `datatypes:create` | `admin_datatypes:create` |
| `admin_update_datatype` | `datatypes:update` | `admin_datatypes:update` |
| `admin_delete_datatype` | `datatypes:delete` | `admin_datatypes:delete` |
| `admin_get_datatype_max_sort_order` | `datatypes:read` | `admin_datatypes:read` |
| `admin_update_datatype_sort_order` | `datatypes:update` | `admin_datatypes:update` |
| `admin_list_fields` | `fields:read` | `admin_fields:read` |
| `admin_get_field` | `fields:read` | `admin_fields:read` |
| `admin_create_field` | `fields:create` | `admin_fields:create` |
| `admin_update_field` | `fields:update` | `admin_fields:update` |
| `admin_delete_field` | `fields:delete` | `admin_fields:delete` |

**Admin media tools** (lines 129-133):

| Tool name | Current | New |
|---|---|---|
| `admin_list_media` | `media:read` | `admin_media:read` |
| `admin_get_media` | `media:read` | `admin_media:read` |
| `admin_update_media` | `media:update` | `admin_media:update` |
| `admin_delete_media` | `media:delete` | `admin_media:delete` |
| `admin_upload_media` | `media:create` | `admin_media:create` |

**Admin media folder tools** (lines 136-143):

| Tool name | Current | New |
|---|---|---|
| `admin_list_media_folders` | `media:read` | `admin_media:read` |
| `admin_get_media_folder` | `media:read` | `admin_media:read` |
| `admin_create_media_folder` | `media:create` | `admin_media:create` |
| `admin_update_media_folder` | `media:update` | `admin_media:update` |
| `admin_delete_media_folder` | `media:delete` | `admin_media:delete` |
| `admin_move_media_to_folder` | `media:update` | `admin_media:update` |
| `admin_get_media_folder_tree` | `media:read` | `admin_media:read` |
| `admin_list_media_in_folder` | `media:read` | `admin_media:read` |

**Admin route tools** (lines 154-158):

| Tool name | Current | New |
|---|---|---|
| `admin_list_routes` | `routes:read` | `admin_routes:read` |
| `admin_get_route_by_slug` | `routes:read` | `admin_routes:read` |
| `admin_create_route` | `routes:create` | `admin_routes:create` |
| `admin_update_route` | `routes:update` | `admin_routes:update` |
| `admin_delete_route` | `routes:delete` | `admin_routes:delete` |

**Admin field_types tools** (lines 159-163, pre-existing inconsistency fix):

| Tool name | Current | New |
|---|---|---|
| `admin_list_field_types` | `field_types:read` | `admin_field_types:read` |
| `admin_get_field_type` | `field_types:read` | `admin_field_types:read` |
| `admin_create_field_type` | `field_types:create` | `admin_field_types:create` |
| `admin_update_field_type` | `field_types:update` | `admin_field_types:update` |
| `admin_delete_field_type` | `field_types:delete` | `admin_field_types:delete` |

**Admin publishing tools** (lines 249-251):

| Tool name | Current (phantom) | New |
|---|---|---|
| `admin_publish_content` | `publishing:create` | `admin_content:publish` |
| `admin_unpublish_content` | `publishing:delete` | `admin_content:publish` |
| `admin_schedule_content` | `publishing:create` | `admin_content:publish` |

**Admin version tools** (lines 259-263):

| Tool name | Current (phantom) | New |
|---|---|---|
| `admin_list_content_versions` | `versions:read` | `admin_content:read` |
| `admin_get_content_version` | `versions:read` | `admin_content:read` |
| `admin_create_content_version` | `versions:create` | `admin_content:update` |
| `admin_delete_content_version` | `versions:delete` | `admin_content:delete` |
| `admin_restore_content_version` | `versions:update` | `admin_content:update` |

`admin_create_content_version` maps to `admin_content:update` (not `admin_content:create`) because the router uses `content:update` for `POST /api/v1/admin/content/versions` (line 245). Creating a version snapshot is an update operation.

`admin_restore_content_version` maps to `admin_content:update` because the router uses `content:update` for `POST /api/v1/admin/content/restore` (line 253).

**Admin translation tool** (line 283):

| Tool name | Current | New |
|---|---|---|
| `admin_create_translation` | `locales:create` | `admin_content:create` |

`admin_create_translation` maps to `admin_content:create` (not `locale:create`) because the corresponding router endpoint (`POST /api/v1/admin/admincontentdata/{id}/translations`, line 691) uses `admin_content:create`. This tool creates locale translations on admin content data, so the permission aligns with the resource being modified (admin content), not the feature domain (locales).

**Admin validation tools** (lines 292-297, pre-existing inconsistency fix):

| Tool name | Current | New |
|---|---|---|
| `admin_list_validations` | `validations:read` | `admin_validations:read` |
| `admin_get_validation` | `validations:read` | `admin_validations:read` |
| `admin_create_validation` | `validations:create` | `admin_validations:create` |
| `admin_update_validation` | `validations:update` | `admin_validations:update` |
| `admin_delete_validation` | `validations:delete` | `admin_validations:delete` |
| `admin_search_validations` | `validations:read` | `admin_validations:read` |

### 8. Documentation

**documentation/reference/mcp-authentication.md:** Update the "Admin-prefixed tools" sentence and all admin permission reference tables to show the new `admin_content:*`, `admin_datatypes:*`, `admin_fields:*`, `admin_media:*`, `admin_routes:*` labels.

**documentation/custom-admin/authentication.md:** Update permission counts and the "All available permissions" table to include the 26 new permissions.

### 9. Verification

1. `just check` after steps 1-7.
2. `just test` after all steps.
3. `go test -run TestPermissionMapCompleteness ./internal/mcp/` -- verify MCP map still covers all tools.
4. `go test -run TestPermissionLabelsValid ./internal/mcp/` -- verify all new labels pass validation.

## Endpoints that stay unchanged

These endpoints have admin-like URLs but operate on public resources. Do NOT change their permission labels:

- `POST /api/v1/admin/content/heal` (line 268) -- heals the public content tree, stays on `content:update`
- `POST /api/v1/admin/contentdata/{id}/translations` (line 688) -- creates translations on public content, stays on `content:create`

## Line number note

Line numbers reference the current state of files at planning time. Use the string values (tool names, permission labels, URL patterns) for matching, not line numbers. Adding 26 new strings to `rbacPermissionLabels` in step 1 will shift all subsequent line numbers in `db.go`.
