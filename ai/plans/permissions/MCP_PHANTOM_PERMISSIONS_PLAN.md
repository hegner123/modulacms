# Fix MCP Phantom Permission Labels

## Context

The MCP permission map (`internal/mcp/permissions.go` `toolPermissions`) maps tool names to `resource:operation` permission labels. Several entries reference labels that do not exist in the bootstrap data (`internal/db/db.go` `rbacPermissionLabels`). The HTTP router uses the correct labels for the same operations. Non-admin users cannot use these MCP tools because the permission check always fails: the label in the map has no matching row in the `permissions` table, so no role can ever include it.

Seven categories of phantom labels exist:

1. **Publishing tools** use `publishing:create` / `publishing:delete`. Bootstrap has `content:publish`. Router uses `content:publish`.
2. **Version tools** use `versions:read` / `versions:create` / `versions:update` / `versions:delete`. No `versions:*` permissions exist in bootstrap. Router uses `content:read`, `content:update`, `content:delete`.
3. **Webhook tools** use plural `webhooks:*`. Bootstrap uses singular `webhook:*`. Router uses singular `webhook:*`.
4. **Locale tools** use plural `locales:*`. Bootstrap uses singular `locale:*`.
5. **OAuth tools** use `oauth:read` / `oauth:create` / `oauth:update` / `oauth:delete`. No `oauth:*` permissions exist in bootstrap. Router uses `RequireResourcePermission("users")` for OAuth endpoints.
6. **Role permission tools** use `role_permissions:read` / `role_permissions:create` / `role_permissions:delete`. No `role_permissions:*` permissions exist in bootstrap. Router uses `RequireResourcePermission("roles")`.
7. **Activity tool** uses `activity:read`. Bootstrap has `audit:read`. Router uses `RequirePermission("audit:read")`.

This plan fixes only the public tool entries. Admin-prefixed tool entries (`admin_publish_content`, `admin_list_content_versions`, etc.) are addressed by the admin resource permission separation plan.

## Dependency

None. This plan has no dependency on the admin separation plan. Either can land first.

## Line number note

Line numbers reference the current state of `internal/mcp/permissions.go` at planning time. Use the tool name strings for matching, not line numbers.

## Changes

### 1. Fix publishing tool labels: internal/mcp/permissions.go (lines 246-248)

| Tool name | Current (phantom) | Correct (matches router + bootstrap) |
|---|---|---|
| `publish_content` | `publishing:create` | `content:publish` |
| `unpublish_content` | `publishing:delete` | `content:publish` |
| `schedule_content` | `publishing:create` | `content:publish` |

Do NOT change lines 249-251 (`admin_publish_content`, `admin_unpublish_content`, `admin_schedule_content`). Those are addressed by the admin separation plan.

### 2. Fix version tool labels: internal/mcp/permissions.go (lines 254-258)

| Tool name | Current (phantom) | Correct (matches router + bootstrap) |
|---|---|---|
| `list_content_versions` | `versions:read` | `content:read` |
| `get_content_version` | `versions:read` | `content:read` |
| `create_content_version` | `versions:create` | `content:update` |
| `delete_content_version` | `versions:delete` | `content:delete` |
| `restore_content_version` | `versions:update` | `content:update` |

`create_content_version` maps to `content:update` (not `content:create`) because the router uses `content:update` for `POST /api/v1/content/versions` (line 215 of mux.go). Creating a version snapshot is an update operation on the content, not a content creation.

`restore_content_version` maps to `content:update` because the router uses `content:update` for `POST /api/v1/content/restore` (line 223 of mux.go).

Do NOT change lines 259-263 (`admin_list_content_versions`, etc.). Those are addressed by the admin separation plan.

### 3. Fix webhook tool labels: internal/mcp/permissions.go (lines 266-273)

| Tool name | Current (phantom, plural) | Correct (matches router + bootstrap, singular) |
|---|---|---|
| `list_webhooks` | `webhooks:read` | `webhook:read` |
| `get_webhook` | `webhooks:read` | `webhook:read` |
| `create_webhook` | `webhooks:create` | `webhook:create` |
| `update_webhook` | `webhooks:update` | `webhook:update` |
| `delete_webhook` | `webhooks:delete` | `webhook:delete` |
| `test_webhook` | `webhooks:update` | `webhook:update` |
| `list_webhook_deliveries` | `webhooks:read` | `webhook:read` |
| `retry_webhook_delivery` | `webhooks:update` | `webhook:update` |

### 4. Fix locale tool labels: internal/mcp/permissions.go (lines 276-283)

| Tool name | Current (phantom, plural) | Correct (matches router + bootstrap, singular) |
|---|---|---|
| `list_locales` | `locales:read` | `locale:read` |
| `list_admin_locales` | `locales:read` | `locale:read` | (router: `RequireResourcePermission("locale")` for `/api/v1/admin/locales`, line 614 of mux.go)
| `get_locale` | `locales:read` | `locale:read` |
| `create_locale` | `locales:create` | `locale:create` |
| `update_locale` | `locales:update` | `locale:update` |
| `delete_locale` | `locales:delete` | `locale:delete` |
| `create_translation` | `locales:create` | `content:create` |

`create_translation` maps to `content:create` (not `locale:create`) because the router uses `RequirePermission("content:create")` for `POST /api/v1/admin/contentdata/{id}/translations` (line 688 of mux.go). This tool creates a locale translation on a content item, so the permission aligns with the resource being modified (content), not the feature domain (locales).

Do NOT change `admin_create_translation` (line 283). It is an admin-prefixed tool addressed by the admin resource permission separation plan.

### 5. Fix OAuth tool labels: internal/mcp/permissions.go (lines 212-216)

| Tool name | Current (phantom) | Correct (matches router + bootstrap) |
|---|---|---|
| `list_users_oauth` | `oauth:read` | `users:read` |
| `get_user_oauth` | `oauth:read` | `users:read` |
| `create_user_oauth` | `oauth:create` | `users:create` |
| `update_user_oauth` | `oauth:update` | `users:update` |
| `delete_user_oauth` | `oauth:delete` | `users:delete` |

The router uses `RequireResourcePermission("users")` for `/api/v1/usersoauth` and `/api/v1/usersoauth/` (lines 489, 492 of mux.go). OAuth connections are a sub-resource of users, so the permission aligns with the parent resource.

### 6. Fix role permission tool labels: internal/mcp/permissions.go (lines 188-192)

| Tool name | Current (phantom) | Correct (matches router + bootstrap) |
|---|---|---|
| `assign_role_permission` | `role_permissions:create` | `roles:create` |
| `remove_role_permission` | `role_permissions:delete` | `roles:delete` |
| `list_role_permissions` | `role_permissions:read` | `roles:read` |
| `get_role_permission` | `role_permissions:read` | `roles:read` |
| `list_role_permissions_by_role` | `role_permissions:read` | `roles:read` |

The router uses `RequireResourcePermission("roles")` for `/api/v1/role-permissions` and `/api/v1/role-permissions/` (lines 555, 558 of mux.go), and `RequirePermission("roles:read")` for `/api/v1/role-permissions/role/` (line 561). Role-permission management is a sub-resource of roles.

### 7. Fix activity tool label: internal/mcp/permissions.go (line 307)

| Tool name | Current (phantom) | Correct (matches router + bootstrap) |
|---|---|---|
| `list_recent_activity` | `activity:read` | `audit:read` |

The router uses `RequirePermission("audit:read")` for `GET /api/v1/activity/recent` (line 530 of mux.go). Bootstrap has `audit:read` (line 422 of db.go).

### 8. Update section comments: internal/mcp/permissions.go

Change the `// Publishing tools` comment to `// Publishing tools (content:publish, matching router)`.

Change the `// Version tools` comment to `// Version tools (content:read/update/delete, matching router)`.

Do NOT modify section comments for steps 3-7 (Webhooks, Locales, OAuth, Role permissions, Activity). Only update comments for Publishing and Versions.

### 9. Documentation

Update `documentation/reference/mcp-authentication.md` permission reference tables. Phantom labels appear in multiple sections across the file:

- Role permissions section (lines 359-363): `role_permissions:*` to `roles:*`
- OAuth section (lines 395-399): `oauth:*` to `users:*`
- Publishing section (lines 441-446): `publishing:*` to `content:publish`
- Versions section (lines 452-461): `versions:*` to `content:read/update/delete`
- Webhooks section (lines 467-474): `webhooks:*` to `webhook:*`
- Locales section (lines 480-487): `locales:*` to `locale:*`
- Activity entry (line 520): `activity:read` to `audit:read`

Apply the same tool-to-permission mappings from steps 1-7 to the corresponding documentation tables. Each documentation row must match its counterpart in the updated `toolPermissions` map. Do not add a new cross-reference test in this plan; that is tracked separately.

### 10. Verification

1. `just check` after steps 1-8.
2. `go test -run TestPermissionMapCompleteness ./internal/mcp/` -- verify all tools still covered.
3. `go test -run TestPermissionLabelsValid ./internal/mcp/` -- verify all labels pass format validation. Note: this test only validates syntactic format (`resource:operation` shape). It does NOT verify that labels exist in bootstrap data. The test passing means no syntax errors were introduced, not that the labels are correct.
4. `just test` -- full suite.

## Known issues out of scope

- `plugins:update` and `plugins:delete` are used in the MCP map but only `plugins:read` and `plugins:admin` exist in bootstrap. This is likely intentional (plugins are admin-only operations). Track separately if granular plugin permissions are needed.
- `health:read` is used by 3 MCP tools (`health`, `get_metrics`, `get_environment`) but no `health:*` permission exists in bootstrap. The HTTP health endpoint (`GET /api/v1/health`) has no permission guard (publicly accessible). Decision needed: should these MCP tools be added to `publicTools`, or should `health:read` be added to bootstrap? Track separately.
- `sessions:update` is used by the `update_session` MCP tool but bootstrap only has `sessions:read` and `sessions:delete`. The router also uses `RequireResourcePermission("sessions")` which auto-maps PUT to `sessions:update`, so this is a broader RBAC bootstrap gap affecting both router and MCP. Fixing it requires adding `sessions:update` to bootstrap and the Ensure function. Track separately.
- Admin-prefixed tool phantom permissions (`admin_publish_content` using `publishing:create`, etc.) are addressed by the admin resource permission separation plan.
