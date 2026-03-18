# Admin Media: Separate Bucket per Content System

## Summary

Complete the dual content system pattern by adding `admin_media` and `admin_media_folders` tables with an independent S3 bucket configuration. Public content media goes to `bucket_media`, admin content media goes to `bucket_admin_media`. When `bucket_admin_media` is not configured, falls back to `bucket_media` for zero-config simplicity.

## Motivation

Every entity in ModulaCMS follows the dual-table pattern — `content_data`/`admin_content_data`, `fields`/`admin_fields`, `routes`/`admin_routes`. Media is the one entity that doesn't. This creates a coupling between the two content systems that breaks the isolation guarantee.

With planned commercial admin panels (pre-built layouts + content models), admin media isolation becomes critical:
- Swapping/resetting an admin panel won't risk touching client content media
- Admin panel assets get their own access policies and CDN configuration
- Storage costs separate naturally between infrastructure and content
- Importing a pre-built panel is a clean, self-contained operation

## Implementation

### Phase 1: Schema & Code Generation

**1.1 — SQL schema files**

Create `sql/schema/38_admin_media/` with all 6 files (schema + queries for SQLite, MySQL, PostgreSQL). Mirror `sql/schema/14_media/` with these changes:
- Table name: `admin_media`
- Primary key: `admin_media_id`
- Foreign key: `folder_id` references `admin_media_folders(admin_folder_id)` instead of `media_folders`
- Index names prefixed with `idx_admin_media_`

Create `sql/schema/39_admin_media_folders/` with all 6 files. Mirror `sql/schema/37_media_folders/` with:
- Table name: `admin_media_folders`
- Primary key: `admin_folder_id`
- Self-referencing FK: `parent_id` references `admin_media_folders(admin_folder_id)`
- Index names prefixed with `idx_admin_media_folders_`

**1.2 — Run code generation**

- `just sqlc` to generate Go code in `internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/`
- `just dbgen` to generate wrapper code

**1.3 — Typed IDs**

Add to `internal/db/types/`:
- `AdminMediaID` (mirrors `MediaID`)
- `AdminMediaFolderID` (mirrors `MediaFolderID`)
- `NullableAdminMediaID`
- `NullableAdminMediaFolderID`

### Phase 2: Database Layer

**2.1 — DbDriver interface**

Add two new embedded interfaces to `internal/db/db.go`:
- `AdminMediaRepository` — mirrors `MediaRepository` with admin types
- `AdminMediaFolderRepository` — mirrors `MediaFolderRepository` with admin types

**2.2 — Wrapper methods**

Create `internal/db/admin_media.go` (generated via `just dbgen`).

Create `internal/db/admin_media_custom.go` with custom methods mirroring `media_custom.go`:
- `ListAdminMediaByFolder`, `ListAdminMediaByFolderPaginated`
- `ListAdminMediaUnfiled`, `ListAdminMediaUnfiledPaginated`
- `CountAdminMediaByFolder`, `CountAdminMediaUnfiled`
- `MoveAdminMediaToFolder`

Create `internal/db/admin_media_folder_custom.go` mirroring `media_folder_custom.go`:
- `GetAdminMediaFolderBreadcrumb`
- `ValidateAdminMediaFolderName`
- `ValidateAdminMediaFolderMove`

Run `just drivergen` after writing SQLite methods.

**2.3 — Consts and generic list**

Update `internal/db/consts.go`:
- Add `DBTableAdminMedia` and `DBTableAdminMediaFolders` constants
- Add `allTables`, `TableStructMap`, `CastToTypedSlice` entries

Update `internal/db/generic_list.go`:
- Add `GenericHeaders` and `GenericList` cases for both new tables

### Phase 3: Config & Media Service

**3.1 — Config fields**

Add to `internal/config/config.go`:
- `Bucket_Admin_Media string` — bucket name for admin media
- `Bucket_Admin_Endpoint string` — optional separate endpoint
- `Bucket_Admin_Access_Key string` — optional separate credentials
- `Bucket_Admin_Secret_Key string` — optional separate credentials
- `Bucket_Admin_Public_URL string` — optional separate public URL

Fallback logic: if `Bucket_Admin_Media` is empty, use `Bucket_Media` and shared credentials. This means zero config change for users who don't need separation.

**3.2 — Admin media service**

Create `internal/media/admin_media_service.go` that mirrors `media_service.go` but:
- Accepts an `AdminMediaStore` interface (admin DB methods)
- Uses the admin bucket client (or shared client via fallback)
- Writes to admin_media table

The upload pipeline (image optimization, dimension variants, srcset generation) is shared — extract common logic into internal helpers if not already factored.

### Phase 4: API Endpoints

**4.1 — Router registration**

Add admin media routes to `internal/router/mux.go` mirroring public media routes:

```
POST   /api/v1/adminmedia                    (media:create)
GET    /api/v1/adminmedia                     (media:read)
GET    /api/v1/adminmedia/{id}                (media:read)
PUT    /api/v1/adminmedia/{id}                (media:update)
DELETE /api/v1/adminmedia/{id}                (media:delete)
GET    /api/v1/adminmedia/{id}/download       (media:read)
GET    /api/v1/adminmedia/references          (media:read)
POST   /api/v1/adminmedia/move                (media:update)

GET    /api/v1/adminmedia-folders             (media:read)
POST   /api/v1/adminmedia-folders             (media:create)
GET    /api/v1/adminmedia-folders/tree         (media:read)
GET    /api/v1/adminmedia-folders/{id}         (media:read)
PUT    /api/v1/adminmedia-folders/{id}         (media:update)
DELETE /api/v1/adminmedia-folders/{id}         (media:delete)
GET    /api/v1/adminmedia-folders/{id}/media   (media:read)
```

**4.2 — Route handlers**

Create `internal/router/admin_media.go` and `internal/router/admin_media_folders.go` mirroring existing media handlers but using admin media service and admin DB methods.

### Phase 5: Admin Panel & TUI

**5.1 — Admin panel pages**

Add admin media management pages to `internal/admin/`:
- `pages/admin_media_list.templ` — list/grid view with folder navigation
- `pages/admin_media_detail.templ` — single media detail/edit
- `partials/admin_media_table_rows.templ` — HTMX table rows
- `handlers/admin_media.go` — admin panel handlers

Register admin panel routes in `mux.go` under `/admin/admin-media/`.

**5.2 — TUI screen**

Add admin media screen to `internal/tui/` following the Screen interface pattern. Mirror the existing media screen but targeting admin_media tables.

### Phase 6: SDKs

**6.1 — TypeScript SDK**

Add to `@modulacms/types`:
- `AdminMedia`, `AdminMediaFolder` types
- `AdminMediaID`, `AdminMediaFolderID` branded IDs
- `CreateAdminMediaParams`, `UpdateAdminMediaParams`

Add to `@modulacms/admin-sdk`:
- `adminMedia` resource with full CRUD + upload
- `adminMediaFolders` resource with CRUD + tree

**6.2 — Go SDK**

Add to `sdks/go/`:
- `admin_media.go` — AdminMedia type + resource
- `admin_media_folders.go` — AdminMediaFolder type + resource
- `AdminMediaID`, `AdminMediaFolderID` in `ids.go`

**6.3 — Swift SDK**

Add to `sdks/swift/Sources/ModulaCMS/`:
- `AdminMediaResource.swift`
- `AdminMediaFoldersResource.swift`
- Types in `Types.swift`, IDs in `IDs.swift`

### Phase 7: MCP Server

Add admin media tools to `internal/mcp/`:
- `admin_upload_media`, `admin_list_media`, `admin_get_media`, `admin_update_media`, `admin_delete_media`
- `admin_create_media_folder`, `admin_list_media_folders`, `admin_get_media_folder`, etc.
- `admin_move_media_to_folder`

## Testing

- Unit tests for admin media service (duplicate naming, folder validation, depth limits)
- Unit tests for config fallback logic (no admin bucket → uses shared bucket)
- Integration tests for admin media upload pipeline
- SDK type tests for new admin media types
- `TestGenericHeaders_MatchesStringStructFields` must pass with new entities

## Config Documentation

New `modula.config.json` fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `bucket_admin_media` | string | falls back to `bucket_media` | Bucket name for admin media assets |
| `bucket_admin_endpoint` | string | falls back to `bucket_endpoint` | S3 endpoint for admin media |
| `bucket_admin_access_key` | string | falls back to `bucket_access_key` | Access key for admin media bucket |
| `bucket_admin_secret_key` | string | falls back to `bucket_secret_key` | Secret key for admin media bucket |
| `bucket_admin_public_url` | string | falls back to `bucket_public_url` | Public URL for admin media links |

Zero-config: omit all `bucket_admin_*` fields and admin media shares the public media bucket.
