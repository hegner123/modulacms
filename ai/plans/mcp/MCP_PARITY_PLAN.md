# MCP Parity Plan

Two-track plan: fix existing tools for LLM reliability (Track A), then extend to 100% API coverage (Track B). Tracks can run in parallel since Track A modifies existing files and Track B creates new ones.

**Prerequisites:** Both audit documents in `ai/plans/mcp/`:
- `MCP_TOOL_NAMING_AUDIT.md` -- 22 findings across existing tools
- `MCP_API_GAP_ANALYSIS.md` -- 68 missing API endpoints

**Architecture note:** The MCP server uses a backend interface pattern. Each domain has a `*Backend` interface in `backend.go`, implemented by three backend layers: SDK wrappers (`backend_sdk.go`), service wrappers (`backend_service.go`), and proxy wrappers (`backend_proxy.go`). Phase 0 splits these into domain-grouped files (`backend_sdk_*.go`, `backend_service_*.go`, `backend_proxy_*.go`). Adding a new tool requires: (1) add method to backend interface, (2) implement in all three backends, (3) register tool + handler in `tools_*.go`. The Go SDK in `sdks/go/` must have the corresponding HTTP method, or it needs to be added there first.

---

## Phase 0: Backend file split (prerequisite)

Split the three monolithic backend implementation files into entity-grouped files to match the existing `tools_*.go` organization. This is a move-only refactor (no logic changes) that prevents the files from growing to 3000+ lines during Track B implementation and reduces merge conflict density.

**Current state:**
| File | Lines | Structs |
|------|-------|---------|
| `backend_service.go` | 2,353 | 22 svc*Backend structs |
| `backend_sdk.go` | 1,691 | 22 sdk*Backend structs |
| `backend_proxy.go` | 1,344 | 22 proxy*Backend structs |

**Split strategy:** Each file keeps its constructor (`NewSDKBackends`, `NewServiceBackends`, `NewProxyBackends`) and shared infrastructure (imports, `proxyBackends` struct, `errNoConnection`). Every `*Backend` struct and its methods move to a domain-specific file. Group related domains to avoid excessive file count:

| New file suffix | Contains | Approx lines (sdk/svc/proxy) |
|----------------|----------|------------------------------|
| `_content.go` | `ContentBackend` + `AdminContentBackend` | 200/350/230 |
| `_schema.go` | `SchemaBackend` + `AdminSchemaBackend` | 200/260/200 |
| `_media.go` | `MediaBackend` + `AdminMediaBackend` + `MediaFolderBackend` + `AdminMediaFolderBackend` | 250/400/250 |
| `_routes.go` | `RouteBackend` + `AdminRouteBackend` | 150/200/180 |
| `_users.go` | `UserBackend` + `RBACBackend` + `SessionBackend` + `TokenBackend` + `SSHKeyBackend` + `OAuthBackend` | 350/400/350 |
| `_system.go` | `TableBackend` + `PluginBackend` + `ConfigBackend` + `ImportBackend` + `DeployBackend` + `HealthBackend` | 250/350/250 |

Result: 3 monolithic files become 3 base files (constructors only, ~60 lines each) + 18 domain files. Example filenames:
- `backend_sdk.go` (constructor only)
- `backend_sdk_content.go`, `backend_sdk_schema.go`, `backend_sdk_media.go`, `backend_sdk_routes.go`, `backend_sdk_users.go`, `backend_sdk_system.go`
- Same pattern for `backend_service_*.go` and `backend_proxy_*.go`

`backend.go` (interfaces, 299 lines) stays as-is. It is small enough and serves as a single reference for all interface contracts.

**Process:**
1. For each of the 3 backend files: extract struct + method blocks into domain files. Leave constructor and shared types in the base file.
2. Run `just check` after each file to verify compilation.
3. Run `just test` once at the end to verify no behavioral changes.

**Struct naming convention:** SDK structs are `sdk<Domain>Backend` (e.g., `sdkContentBackend`), service structs are `svc<Domain>Backend`, proxy structs are `proxy<Domain>Backend`. Match each struct to its domain file by the interface it implements in `backend.go`.

**No logic changes. No renames. No signature modifications.** This is purely moving code between files within the same package.

---

## Track A: Fix existing tools (naming, descriptions, safety)

### A1. Stale/wrong descriptions (10 min)

**File:** `tools_media.go`

Remove the stale sentence from `list_media` description (line 16):
```
"Media files must be uploaded through the CMS web interface or API directly. This MCP server can view and update media metadata but cannot upload new files."
```
This contradicts `upload_media` which exists on line 57.

**File:** All `tools_*.go` files

Capitalize all description starts. Every `update_*` tool has a lowercase "update" description start. Change to "Update". Also lowercase on `export` (tools_deploy.go:21), `import` (tools_deploy.go:29, tools_import.go:13).

### A2. Fix update handlers to use partial-update semantics (30 min)

**Problem:** Six update handlers use `req.GetString("field", "")` for optional fields, which sends empty strings when the caller omits a field. This silently overwrites existing data with empty values. Other update handlers correctly use `optionalStrPtr()` which returns nil for omitted fields.

**Fix:** Replace `req.GetString("field", "")` with `optionalStrPtr(req, "field")` in all affected handlers so omitted optional fields are not included in the update payload.

**Affected handlers:**

| Handler | File | Dangerous fields |
|---------|------|-----------------|
| `handleUpdateUser` | `tools_users.go` | username, name, email, password, role (ALL 5 optional fields) |
| `handleUpdateSession` | `tools_sessions.go` | expires_at (only field using `GetString`; the other 5 fields already use `optionalStrPtr`, do not change them) |
| `handleUpdateDatatype` | `tools_schema.go` | name |
| `handleAdminUpdateDatatype` | `tools_admin_schema.go` | name |
| `handleUpdateField` | `tools_schema.go` | name, data, validation, ui_config |
| `handleAdminUpdateField` | `tools_admin_schema.go` | name, data, validation, ui_config |

**Description update:** All update tool descriptions should say "Update X." (capitalized, no semantic tags). The `[FULL REPLACE]`/`[PARTIAL]` tagging concept is removed. Note: `handleUpdateDatatype`, `handleAdminUpdateDatatype`, `handleUpdateField`, and `handleAdminUpdateField` still require `label` and `type`/`field_type` on every call (these use `RequireString`). Switching `name` and other optional fields to `optionalStrPtr` makes those fields partial-update, but the required fields remain full-replace. Do not change `GetString` in create handlers, where empty-string defaults are intentional (the server derives a name from the label when name is empty).

### A3. Safety: `media_cleanup` (15 min)

**File:** `tools_media.go`

Split `media_cleanup` into two tools:
1. `media_cleanup_check` -- read-only, lists orphaned records without deleting. New backend method `MediaCleanupCheck`.
2. `media_cleanup_apply` -- destructive, requires `confirm: true` parameter. Reuses existing `MediaCleanup` backend method.

Backend changes:
- `backend.go`: Add `MediaCleanupCheck(ctx context.Context) (json.RawMessage, error)` to `MediaBackend`
- `backend_sdk_media.go`: Implement `MediaCleanupCheck` by delegating to the existing `MediaHealth` method, which returns `OrphanScanResult` containing `orphaned_keys []string` (the actual S3 keys). The cleanup API endpoint (`DELETE /api/v1/media/cleanup`) does not support a dry-run parameter.
- `backend_service_media.go`: Implement `MediaCleanupCheck` by calling `svc.Media.MediaHealth(ctx)` and marshaling the result.
- `backend_proxy_media.go`: Delegate `MediaCleanupCheck` to the underlying SDK backend.

`media_cleanup_check` description: "List orphaned S3 objects that have no corresponding database record. Returns total_objects, tracked_keys, and orphaned_keys. Use media_cleanup_apply to delete the listed orphans."

Remove the old `media_cleanup` tool.

### A4. Deploy/import disambiguation (15 min)

**File:** `tools_deploy.go`

Rename all deploy tool registration strings (the first argument to `mcp.NewTool`). Do NOT rename handler function names (`handleDeploy*`), backend interface method names (`Deploy*`), or backend struct field names. The tool name is the LLM-facing surface; the Go identifiers are internal.
- `deploy_health` -> `sync_health`
- `deploy_export` -> `sync_export`
- `deploy_import` -> `sync_import`
- `deploy_dry_run` -> `sync_preview`

Update descriptions to use "sync" language consistently. Add to each description: "This is for synchronizing data between ModulaCMS environments, not for importing from external CMS platforms."

**File:** `tools_import.go`

Add to `import_content` description: "For importing from external CMS platforms (Contentful, Sanity, etc.). To sync between ModulaCMS environments, use sync_import instead."

`import_bulk` and `import_content` are distinct: `import_content` calls format-specific endpoints (e.g., `/api/v1/import/contentful`), while `import_bulk` calls `/api/v1/import` with format as a parameter. In direct mode (service backend), `import_bulk` currently delegates to `import_content` as a placeholder. Keep both tools. Update `import_bulk` description to: "Bulk import data via the generic import endpoint. Accepts format and data parameters. For format-specific imports (Contentful, Sanity, etc.), use import_content instead."

### A5. `admin_get_route` slug clarity (5 min)

**File:** `tools_admin_routes.go`

Rename `admin_get_route` to `admin_get_route_by_slug`. Update description to: "Get a single admin route by URL slug. Unlike get_route which uses ID-based lookup, admin routes use slug-based lookup."

### ~~A6. `admin_` parameter prefix cleanup~~ REMOVED

The `admin_` prefixed parameters (`admin_route_id`, `admin_content_data_id`, etc.) are **not** redundant with the tool name. They reference distinct admin-specific entity types (`AdminRouteID`, `AdminContentID`, etc.) with their own SDK struct JSON tags. Renaming them to match public parameter names (`route_id`, `content_data_id`) would incorrectly suggest they accept public entity IDs. The prefixes stay.

### A7. `get_datatype_full` split (10 min)

**File:** `tools_schema.go`

Make `id` required on `get_datatype_full`. Add a new tool `list_datatypes_full` that calls the same backend method with empty string ID (which triggers list behavior). This separates the get-one vs list-all semantics.

### A8. `admin_list_media_dimensions` removal (5 min)

**File:** `tools_admin_media.go`

Remove `admin_list_media_dimensions` (lines 68-73). Update `list_media_dimensions` description to: "List all media dimension presets. Dimensions are shared across both public and admin media systems."

### A9. Description consistency pass (15 min)

**All `tools_*.go` files**

- Add `heal_content` default: Change description to clarify `dry_run` defaults to false (currently ambiguous).
- Add route `status` type clarity: `create_route` status description should say "Route status code (positive integer, e.g. 1=active, 0=inactive)" to distinguish from content status string enum.
- Add `remove_role_permission` clarity: Change description to "Remove a permission from a role by association ID. Use list_role_permissions or list_role_permissions_by_role to find the association ID."
- `list_content_fields`: Already says "Use get_content_tree or get_page to see fields for a specific content item." No change needed.
- `create_content` author_id: Tool description already says "author_id is NOT auto-populated; use the whoami tool to get your user ID." No change to tool description. If the `author_id` parameter description (the `mcp.Description` on the parameter, not the tool) lacks this note, add it there.

---

## Track B: New tools for 100% API coverage

### B1. Publishing (critical) -- 6 tools

**New file:** `tools_publishing.go`

**New backend interface:** `PublishingBackend`

```
PublishContent(ctx, params) (json.RawMessage, error)
UnpublishContent(ctx, params) (json.RawMessage, error)
ScheduleContent(ctx, params) (json.RawMessage, error)
AdminPublishContent(ctx, params) (json.RawMessage, error)
AdminUnpublishContent(ctx, params) (json.RawMessage, error)
AdminScheduleContent(ctx, params) (json.RawMessage, error)
```

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `publish_content` | `content_id` (required), `route_id` | `POST /api/v1/content/publish` |
| `unpublish_content` | `content_id` (required), `route_id` | `POST /api/v1/content/unpublish` |
| `schedule_content` | `content_id` (required), `publish_at` (required, RFC3339) | `POST /api/v1/content/schedule` |
| `admin_publish_content` | `content_id` (required), `route_id` | `POST /api/v1/admin/content/publish` |
| `admin_unpublish_content` | `content_id` (required), `route_id` | `POST /api/v1/admin/content/unpublish` |
| `admin_schedule_content` | `content_id` (required), `publish_at` (required) | `POST /api/v1/admin/content/schedule` |

**Go SDK:** `sdks/go/publishing.go` already has all 6 methods: `Publish`, `AdminPublish`, `Unpublish`, `AdminUnpublish`, `Schedule`, `AdminSchedule`. No SDK changes needed for B1.

**Backends struct:** Add `Publishing PublishingBackend` field.

### B2. Content versions (critical) -- 10 tools

**New file:** `tools_versions.go`

**New backend interface:** `VersionBackend`

```
ListVersions(ctx, contentID string) (json.RawMessage, error)
GetVersion(ctx, versionID string) (json.RawMessage, error)
CreateVersion(ctx, params) (json.RawMessage, error)
DeleteVersion(ctx, versionID string) error
RestoreVersion(ctx, params) (json.RawMessage, error)
AdminListVersions(ctx, contentID string) (json.RawMessage, error)
AdminGetVersion(ctx, versionID string) (json.RawMessage, error)
AdminCreateVersion(ctx, params) (json.RawMessage, error)
AdminDeleteVersion(ctx, versionID string) error
AdminRestoreVersion(ctx, params) (json.RawMessage, error)
```

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `list_content_versions` | `content_id` (required) | `GET /api/v1/content/versions?content_id=` |
| `get_content_version` | `id` (required) | `GET /api/v1/content/versions/{id}` |
| `create_content_version` | `content_id` (required), `label` | `POST /api/v1/content/versions` |
| `delete_content_version` | `id` (required) | `DELETE /api/v1/content/versions/{id}` |
| `restore_content_version` | `version_id` (required), `content_id` (required) | `POST /api/v1/content/restore` |
| `admin_list_content_versions` | `content_id` (required) | `GET /api/v1/admin/content/versions` |
| `admin_get_content_version` | `id` (required) | `GET /api/v1/admin/content/versions/{id}` |
| `admin_create_content_version` | `content_id` (required), `label` | `POST /api/v1/admin/content/versions` |
| `admin_delete_content_version` | `id` (required) | `DELETE /api/v1/admin/content/versions/{id}` |
| `admin_restore_content_version` | `version_id` (required), `content_id` (required) | `POST /api/v1/admin/content/restore` |

### B3. Webhooks -- 8 tools

**New file:** `tools_webhooks.go`

**New backend interface:** `WebhookBackend`

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `list_webhooks` | | `GET /api/v1/admin/webhooks` |
| `get_webhook` | `id` | `GET /api/v1/admin/webhooks/{id}` |
| `create_webhook` | `url`, `events`, `secret`, `active` | `POST /api/v1/admin/webhooks` |
| `update_webhook` | `id`, `url`, `events`, `secret`, `active` | `PUT /api/v1/admin/webhooks/{id}` |
| `delete_webhook` | `id` | `DELETE /api/v1/admin/webhooks/{id}` |
| `test_webhook` | `id` | `POST /api/v1/admin/webhooks/{id}/test` |
| `list_webhook_deliveries` | `webhook_id` | `GET /api/v1/admin/webhooks/{id}/deliveries` |
| `retry_webhook_delivery` | `delivery_id` | `POST /api/v1/admin/webhooks/deliveries/{id}/retry` |

### B4. Locales -- 6 tools

**New file:** `tools_locales.go`

**New backend interface:** `LocaleBackend`

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `list_locales` | | `GET /api/v1/locales` (public) |
| `list_admin_locales` | | `GET /api/v1/admin/locales` |
| `get_locale` | `id` | `GET /api/v1/admin/locales/{id}` |
| `create_locale` | `code`, `label`, `enabled` | `POST /api/v1/admin/locales` |
| `update_locale` | `id`, `code`, `label`, `enabled` | `PUT /api/v1/admin/locales/{id}` |
| `delete_locale` | `id` | `DELETE /api/v1/admin/locales/{id}` |

### B5. Translations -- 2 tools

**New file:** `tools_translations.go` (or add to `tools_locales.go`)

**New backend interface:** `TranslationBackend`

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `create_translation` | `content_id`, `locale_id`, fields... | `POST /api/v1/admin/contentdata/{id}/translations` |
| `admin_create_translation` | `content_id`, `locale_id`, fields... | `POST /api/v1/admin/admincontentdata/{id}/translations` |

### B6. Validations -- 12 tools

**New file:** `tools_validations.go`

**New backend interface:** `ValidationBackend`

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `list_validations` | | `GET /api/v1/validations` |
| `get_validation` | `id` | `GET /api/v1/validations/{id}` |
| `create_validation` | `field_id`, `rules`... | `POST /api/v1/validations` |
| `update_validation` | `id`, `rules`... | `PUT /api/v1/validations/{id}` |
| `delete_validation` | `id` | `DELETE /api/v1/validations/{id}` |
| `search_validations` | `query` | `GET /api/v1/validations/search` |
| `admin_list_validations` | | `GET /api/v1/admin/validations` |
| `admin_get_validation` | `id` | `GET /api/v1/admin/validations/{id}` |
| `admin_create_validation` | `field_id`, `rules`... | `POST /api/v1/admin/validations` |
| `admin_update_validation` | `id`, `rules`... | `PUT /api/v1/admin/validations/{id}` |
| `admin_delete_validation` | `id` | `DELETE /api/v1/admin/validations/{id}` |
| `admin_search_validations` | `query` | `GET /api/v1/admin/validations/search` |

### B7. Search -- 2 tools

**New file:** `tools_search.go`

**New backend interface:** `SearchBackend`

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `search_content` | `q` (required), `limit`, `offset` | `GET /api/v1/search` |
| `rebuild_search_index` | | `POST /api/v1/admin/search/rebuild` |

### B8. Content query -- 1 tool

**Add to:** `tools_content.go`

**Add to:** `ContentBackend` interface

Tool:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `query_content` | `datatype` (required), `filter`, `sort`, `limit`, `offset` | `GET /api/v1/query/{datatype}` |

### B9. Activity feed -- 1 tool

**New file:** `tools_activity.go`

**New backend interface:** `ActivityBackend`

Tool:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `list_recent_activity` | `limit` | `GET /api/v1/activity/recent` |

### B10. Metrics -- 1 tool

**Add to:** `tools_health.go`

**Add to:** `HealthBackend` interface

Tool:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `get_metrics` | | `GET /api/v1/admin/metrics` |

### B11. Environment -- 1 tool

**Add to:** `tools_health.go`

**Add to:** `HealthBackend` interface

Tool:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `get_environment` | | `GET /api/v1/environment` |

### B12. Globals -- 1 tool

**Add to:** `tools_content.go`

**Add to:** `ContentBackend` interface

Tool:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `get_globals` | `format` | `GET /api/v1/globals` |

### B13. Content extras -- 3 tools

**Add to:** `tools_content.go`

**Add to:** `ContentBackend` interface

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `get_content_full` | `limit`, `offset` | `GET /api/v1/contentdata/full` |
| `get_content_by_route` | `route_id` (required) | `GET /api/v1/contentdata/by-route` |
| `create_content_composite` | `content`, `fields` | `POST /api/v1/content/create` |

`create_content_composite` creates a content entry with field values in a single request, eliminating the N+1 call pattern.

### B14. Admin content extras -- 1 tool

**Add to:** `tools_admin_content.go`

**Add to:** `AdminContentBackend` interface

Tool:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `admin_get_content_full` | `limit`, `offset` | `GET /api/v1/admincontentdatas/full` |

### B15. Datatype sort ordering -- 4 tools

**Add to:** `tools_schema.go` and `tools_admin_schema.go`

**Add to:** `SchemaBackend` and `AdminSchemaBackend` interfaces

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `get_datatype_max_sort_order` | | `GET /api/v1/datatype/max-sort-order` |
| `update_datatype_sort_order` | `id`, `sort_order` | `PUT /api/v1/datatype/{id}/sort-order` |
| `admin_get_datatype_max_sort_order` | | `GET /api/v1/admindatatypes/max-sort-order` |
| `admin_update_datatype_sort_order` | `id`, `sort_order` | `PUT /api/v1/admindatatypes/{id}/sort-order` |

### B16. Field sort ordering -- 2 tools

**Add to:** `tools_schema.go`

**Add to:** `SchemaBackend` interface

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `get_field_max_sort_order` | | `GET /api/v1/fields/max-sort-order` |
| `update_field_sort_order` | `id`, `sort_order` | `PUT /api/v1/fields/{id}/sort-order` |

### B17. Media extras -- 4 tools

**Add to:** `tools_media.go`

**Add to:** `MediaBackend` interface

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `download_media` | `id` (required) | `GET /api/v1/media/{id}/download` |
| `get_media_full` | `limit`, `offset` | `GET /api/v1/media/full` |
| `get_media_references` | `id` (required) | `GET /api/v1/media/references` |
| `reprocess_media` | | `POST /api/v1/media/reprocess` |

Note: `download_media` returns binary data. The MCP tool should save to a temp file and return the path, since MCP cannot stream binary.

### B18. Media folder extras -- 4 tools

**Add to:** `tools_media_folders.go` and `tools_admin_media_folders.go`

**Add to:** `MediaFolderBackend` and `AdminMediaFolderBackend` interfaces

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `get_media_folder_tree` | | `GET /api/v1/media-folders/tree` |
| `list_media_in_folder` | `folder_id` (required), `limit`, `offset` | `GET /api/v1/media-folders/{id}/media` |
| `admin_get_media_folder_tree` | | `GET /api/v1/adminmedia-folders/tree` |
| `admin_list_media_in_folder` | `folder_id` (required), `limit`, `offset` | `GET /api/v1/adminmedia-folders/{id}/media` |

### B19. Routes full view -- 1 tool

**Add to:** `tools_routes.go`

**Add to:** `RouteBackend` interface

Tool:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `list_routes_full` | | `GET /api/v1/routes/full` |

### B20. User extras -- 2 tools

**Add to:** `tools_users.go`

**Add to:** `UserBackend` interface

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `reassign_and_delete_user` | `user_id` (required), `reassign_to` (required) | `POST /api/v1/users/reassign-delete` |
| `list_user_sessions` | | `GET /api/v1/users/sessions` |

### B21. Auth (selective) -- 2 tools

**New file:** `tools_auth.go`

**New backend interface:** `AuthBackend`

Only the non-interactive auth operations that make sense for MCP:

Tools:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `register_user` | `username`, `name`, `email`, `password` | `POST /api/v1/auth/register` |
| `request_password_reset` | `email` | `POST /api/v1/auth/request-password-reset` |

Exclude: `login`, `logout` (MCP uses API key), `confirm-password-reset` (requires token from email), `reset` (legacy), OAuth flows (require browser).

### B22. Admin tree -- 1 tool

**Add to:** `tools_admin_content.go`

**Add to:** `AdminContentBackend` interface

Tool:
| Tool | Params | API Endpoint |
|------|--------|-------------|
| `get_admin_tree` | `slug` | `GET /api/v1/admin/tree/{slug}` |

---

## Intentionally excluded (not viable via MCP)

| API Endpoint | Reason |
|-------------|--------|
| `GET /api/v1/auth/oauth/login` | Requires browser redirect |
| `GET /api/v1/auth/oauth/callback` | Requires browser redirect |
| `POST /api/v1/auth/login` | MCP uses API key auth |
| `POST /api/v1/auth/logout` | MCP uses API key auth |
| `GET /api/v1/auth/me` | Covered by `whoami` |
| `POST /api/v1/auth/confirm-password-reset` | Requires email token |
| `POST /api/v1/auth/reset` | Legacy, superseded by request+confirm flow |
| `GET /api/v1/admin/config/search-index` | Internal to admin command palette UI |
| `GET /api/v1/media/reprocess/status` | Polling endpoint, add only if reprocess is added |

---

## Breaking changes (no external consumers)

This project has no active users or external MCP consumers. The following renames and removals are breaking changes that require no migration path:
- A4: `deploy_*` renamed to `sync_*`
- A5: `admin_get_route` renamed to `admin_get_route_by_slug`
- A3: `media_cleanup` removed and split into `media_cleanup_check` + `media_cleanup_apply`
- A7: `get_datatype_full` becomes required-ID; callers using it for list behavior must switch to `list_datatypes_full`
- A8: `admin_list_media_dimensions` removed

## Dependents to Update

When adding new backend interface methods (A3 `MediaCleanupCheck`, all Track B interfaces):
- `backend.go` (interface definition)
- All three backend implementations: `backend_sdk_*.go`, `backend_service_*.go`, `backend_proxy_*.go`
- `NewSDKBackends`, `NewServiceBackends`, `NewProxyBackends` constructors (when adding new `*Backend` fields to `Backends` struct)
- `serve.go` -- only when adding a new `register*Tools` call in `newServer` (needed for Track B new tool files like `tools_publishing.go`; not needed for A3's `MediaCleanupCheck` which extends the existing `MediaBackend` interface)

---

## Implementation order

### Phase 0: Backend file split (prerequisite)
1. Split `backend_sdk.go` into base + 6 domain files
2. Split `backend_service.go` into base + 6 domain files
3. Split `backend_proxy.go` into base + 6 domain files

**Deliverable:** 3 monolithic backend files replaced by 3 base files + 18 domain files. Zero logic changes.

**Test:** Run `just check` after each file split. Run `just test` once at the end.

### Phase 1: Reliability (Track A only)
1. A1: Stale descriptions + capitalization
2. A2: Update handler partial-update semantics
3. A3: media_cleanup safety split
4. A4: Deploy/import rename
5. A5: admin_get_route rename
6. A7: get_datatype_full split
7. A8: admin_list_media_dimensions removal
8. A9: Description consistency pass

**Deliverable:** All 170 existing tools have clear, unambiguous names and descriptions. No tool changes behavior, only names/descriptions change (except media_cleanup split and A2 update-handler fix).

**Test:** Run `just check` after each step. Run `just test` after A2 and A3 (handler changes).

### Phase 2: Critical gaps (Track B, high priority)
1. B1: Publishing (6 tools) -- unblocks content lifecycle
2. B2: Content versions (10 tools) -- enables rollback safety
3. B8: Content query (1 tool) -- enables frontend verification
4. B12: Globals (1 tool) -- quick win
5. B13: Content extras (3 tools) -- composite create, full views

**Deliverable:** Content can be created, published, versioned, queried, and restored entirely through MCP.

### Phase 3: Integration gaps (Track B, medium priority)
1. B3: Webhooks (8 tools)
2. B4: Locales (6 tools)
3. B5: Translations (2 tools)
4. B6: Validations (12 tools)
5. B7: Search (2 tools)

**Deliverable:** Full CMS administration capability via MCP.

### Phase 4: Completeness (Track B, fill remaining)
1. B9: Activity feed (1 tool)
2. B10: Metrics (1 tool)
3. B11: Environment (1 tool)
4. B14: Admin content extras (1 tool)
5. B15: Datatype sort ordering (4 tools)
6. B16: Field sort ordering (2 tools)
7. B17: Media extras (4 tools)
8. B18: Media folder extras (4 tools)
9. B19: Routes full view (1 tool)
10. B20: User extras (2 tools)
11. B21: Auth selective (2 tools)
12. B22: Admin tree (1 tool)

**Deliverable:** 100% API parity minus intentional exclusions.

---

## New tool count

| Phase | New tools | Running total |
|-------|-----------|---------------|
| Current | 0 | 170 |
| Phase 1 (Track A) | +1 (media_cleanup split), -2 (media_cleanup removed, admin_list_media_dimensions removed) | 169 |
| Phase 2 | +21 | 190 |
| Phase 3 | +30 | 220 |
| Phase 4 | +23 | 243 |

**Final: ~243 tools covering 100% of the non-interactive API surface.**

---

## Go SDK gaps to check

Before implementing each Track B step, verify the Go SDK (`sdks/go/`) has the corresponding HTTP method. Known SDK files to check:

| Feature | Expected SDK file | Methods needed |
|---------|------------------|----------------|
| Publishing | `publishing.go` | Publish, Unpublish, Schedule (public + admin) |
| Versions | `publishing.go` (not `content_versions.go`) | All exist except `AdminDeleteVersion`. `content_versions.go` has only `ListByContent` (duplicate of `ListVersions`). |
| Webhooks | `webhooks.go` | CRUD + Test + Deliveries + Retry |
| Locales | `locales.go` | CRUD + public list |
| Translations | (new or in locales) | CreateTranslation (public + admin) |
| Validations | `validations.go` | CRUD + Search (public + admin) |
| Search | `search.go` | Search, RebuildIndex |
| Query | `query.go` | QueryByDatatype |
| Activity | (new) | ListRecentActivity |
| Metrics | (new or in health) | GetMetrics |
| Environment | (new or in health) | GetEnvironment |
| Globals | `globals.go` | GetGlobals |
| Content extras | `content_composite.go` has `CreateWithFields` (not `CreateComposite`). `content_data_full.go` has `GetFull`, `ListByRoute`, `AdminGetFull`. |
| Sort ordering | `datatypes_extra.go`, `fields_extra.go` | MaxSortOrder, UpdateSortOrder |
| Media extras | various | Download, ListFull, References, Reprocess |
| Media folder extras | `media_folders.go` | Tree, ListMedia |
| Routes full | (in routes or new) | ListFull |
| User extras | `user_composite.go` | ReassignDelete, ListSessions |
| Auth | `auth.go` | Register, RequestPasswordReset |

If an SDK method is missing, it must be added before the MCP tool. The SDK change is: add a method to the client that calls the HTTP endpoint and returns `json.RawMessage`.

---

## Files modified per phase

### Phase 1
- `internal/mcp/tools_media.go` (A1, A3)
- `internal/mcp/tools_content.go` (A2, A7, A9)
- `internal/mcp/tools_schema.go` (A2, A7)
- `internal/mcp/tools_users.go` (A2)
- `internal/mcp/tools_routes.go` (A2)
- `internal/mcp/tools_deploy.go` (A2, A4)
- `internal/mcp/tools_import.go` (A4)
- `internal/mcp/tools_admin_routes.go` (A2, A5)
- `internal/mcp/tools_admin_content.go` (A2)
- `internal/mcp/tools_admin_schema.go` (A2)
- `internal/mcp/tools_admin_media.go` (A2, A8)
- `internal/mcp/tools_admin_media_folders.go` (A2)
- `internal/mcp/tools_media_folders.go` (A2)
- `internal/mcp/tools_rbac.go` (A9)
- `internal/mcp/tools_sessions.go` (A2)
- `internal/mcp/tools_tables.go` (A2)
- `internal/mcp/tools_tokens.go` (A2)
- `internal/mcp/tools_oauth.go` (A2)
- `internal/mcp/backend.go` (A3)
- `internal/mcp/backend_sdk_media.go` (A3)
- `internal/mcp/backend_service_media.go` (A3)
- `internal/mcp/backend_proxy_media.go` (A3)

### Phase 0
- `internal/mcp/backend_sdk.go` -> base + `backend_sdk_content.go`, `backend_sdk_schema.go`, `backend_sdk_media.go`, `backend_sdk_routes.go`, `backend_sdk_users.go`, `backend_sdk_system.go`
- `internal/mcp/backend_service.go` -> base + `backend_service_content.go`, `backend_service_schema.go`, `backend_service_media.go`, `backend_service_routes.go`, `backend_service_users.go`, `backend_service_system.go`
- `internal/mcp/backend_proxy.go` -> base + `backend_proxy_content.go`, `backend_proxy_schema.go`, `backend_proxy_media.go`, `backend_proxy_routes.go`, `backend_proxy_users.go`, `backend_proxy_system.go`

### Phase 2-4
- `internal/mcp/backend.go` (new interfaces)
- `internal/mcp/backend_sdk_*.go`, `backend_service_*.go`, `backend_proxy_*.go` (new implementations in domain-specific files)
- `internal/mcp/serve.go` (wire new backends)
- New files: `tools_publishing.go`, `tools_versions.go`, `tools_webhooks.go`, `tools_locales.go`, `tools_translations.go`, `tools_validations.go`, `tools_search.go`, `tools_activity.go`, `tools_auth.go`
- Extended files: `tools_content.go`, `tools_admin_content.go`, `tools_schema.go`, `tools_admin_schema.go`, `tools_media.go`, `tools_media_folders.go`, `tools_admin_media_folders.go`, `tools_routes.go`, `tools_users.go`, `tools_health.go`
- SDK files in `sdks/go/` (as needed per gap check)
