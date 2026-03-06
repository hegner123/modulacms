# Media Folders Plan

Database-backed virtual folders for the media library. Folders are first-class entities stored in a `media_folders` table. Media items reference folders via an optional `folder_id` FK. S3 keys remain flat (ULID-based) -- folder structure is purely a database/UI concept.

## Current State

- No folder concept in the database
- TUI builds a virtual tree by parsing URL path segments from S3 URLs (`media_tree.go`)
- No way to create empty folders or organize media before uploading
- Admin panel shows a flat grid with pagination

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Folder storage | `media_folders` table | Decouples organization from S3 keys entirely |
| Nesting | Self-referential `parent_id` FK | Matches existing tree patterns (content tree), infinite depth |
| Media linkage | Nullable `folder_id` FK on `media` table | Unfiled media lives at root, no sentinel row needed |
| S3 keys | No change (stay flat/ULID-based) | Folder renames are O(1) DB ops, no S3 key mutations |
| Folder ID type | `MediaFolderID` (ULID) | Consistent with all other entity IDs |
| Ordering | Alphabetical by `name` | Simple, predictable, no reorder query complexity |
| Folder deletion | `ON DELETE RESTRICT` on `parent_id` | Refuse delete of non-empty folders; matches file manager behavior |
| Media orphaning | `ON DELETE SET NULL` on `media.folder_id` | Deleting an empty folder never deletes media files |
| Name uniqueness | Application-level enforcement | `UNIQUE(parent_id, name)` breaks on NULL across databases; enforce before insert |
| Breadcrumbs | App-side parent walk with depth guard | Avoids recursive CTE portability issues; practical depths are <10 |
| Audit trail | All folder mutations audited | Consistent with every other entity in the system |
| Permissions | Reuse `media:*` | Folders are organizational, not a separate resource; avoids permission bloat |
| S3 prefix alignment | No | Keeps folder renames free; nobody should browse the bucket directly |

## Schema

### `media_folders` table (new)

```
media_folders
  folder_id     TEXT/VARCHAR(26) PK NOT NULL  -- ULID
  name          TEXT NOT NULL                 -- display name
  slug          TEXT NOT NULL                 -- URL-safe, auto-generated from name, used for breadcrumb paths
  parent_id     TEXT/VARCHAR(26) NULL         -- FK self-ref, NULL = root, ON DELETE RESTRICT
  date_created  TIMESTAMP
  date_modified TIMESTAMP
```

**Constraints:**
- `parent_id` FK references `media_folders(folder_id)` with `ON DELETE RESTRICT` -- must empty a folder before deleting it
- Name uniqueness within a parent is enforced at the application layer before insert/update (not via DB constraint, due to NULL parent_id inconsistency across SQLite/MySQL/PostgreSQL)
- `slug` is auto-generated from `name` via `strings.ReplaceAll(strings.ToLower(name), " ", "-")` (or similar sanitizer)
- Max nesting depth enforced at application layer: 10 levels (prevents UI issues and runaway parent walks)

### `media` table (alter)

Add one column:

```
folder_id TEXT/VARCHAR(26) NULL REFERENCES media_folders(folder_id) ON DELETE SET NULL
```

`ON DELETE SET NULL` means deleting an (empty) folder orphans its media to root rather than deleting the files.

---

## Phases

### Phase 1: Database Layer -- `media_folders` table

**Schema directory:** `sql/schema/36_media_folders/`

1. Create 6 SQL files (schema + queries for SQLite, MySQL, PostgreSQL)
   - Standard CRUD: Create, Get, List, Update, Delete, Count
   - Extra queries:
     - `ListMediaFoldersByParent(parent_id)` -- children of a folder (NULL = root)
     - `GetMediaFolderBySlugAndParent(slug, parent_id)` -- for uniqueness check before insert
2. Add `MediaFolderID` typed ID to `internal/db/types/types_ids.go`
3. Add `NullableMediaFolderID` for the nullable FK on media
4. Add sqlc overrides in `tools/sqlcgen/definitions.go`
5. Add dbgen entity definition in `tools/dbgen/definitions.go`
   - All CRUD methods use `(context.Context, audited.AuditContext, Params)` signatures
6. Run `just sqlc` and `just dbgen`
7. Add methods to `DbDriver` interface in `internal/db/db.go`
8. Add to `CreateAllTables` / `DropAllTables`:
   - **CreateAllTables:** `CreateMediaFolderTable()` must run **before** `CreateMediaTable()` (FK dependency)
   - **DropAllTables:** `DropMediaTable()` must run **before** `DropMediaFolderTable()` (reverse order)
9. Add breadcrumb helper in `internal/db/media_folder_custom.go`:
   - `GetMediaFolderBreadcrumb(folder_id) ([]MediaFolder, error)` -- app-side walk: loop `GetMediaFolder(parent_id)` up to root, with depth guard (max 10 iterations, return error on circular reference)
   - `ValidateMediaFolderName(name, parent_id) error` -- check uniqueness within parent before insert/update
10. Write tests

### Phase 2: Add `folder_id` to `media` table

1. Add `folder_id` column to all 3 media schemas (`sql/schema/14_media/`)
2. Update media queries:
   - `ListMediaByFolder(folder_id)` -- media in a specific folder
   - `ListMediaUnfiled()` -- media with NULL folder_id (root level)
   - `MoveMediaToFolder(media_id, folder_id)` -- reassign folder (single item, audited)
3. No bulk move query -- use app-side loop within a transaction for batch moves (avoids array syntax portability issues across SQLite/MySQL/PostgreSQL)
4. Update sqlc overrides for `media.folder_id` -> `types.NullableMediaFolderID`
5. Update `CreateMediaParams` / `UpdateMediaParams` to include `folder_id`
6. Run `just sqlc` and `just dbgen`
7. Update `DbDriver` interface with new media queries
8. Add `MoveMediaBatch` in `internal/db/media_custom.go`:
   - Wraps individual `MoveMediaToFolder` calls in a transaction
   - Returns on first error, rolls back entire batch

### Phase 3: REST API

**New endpoints for folder CRUD:**

```
GET    /api/v1/media-folders              -- list root folders (or ?parent_id= for children)
GET    /api/v1/media-folders/{id}         -- get folder details
POST   /api/v1/media-folders              -- create folder (audited)
PUT    /api/v1/media-folders/{id}         -- rename/move folder (audited)
DELETE /api/v1/media-folders/{id}         -- delete folder (audited, fails if non-empty)
GET    /api/v1/media-folders/{id}/media   -- list media in folder
GET    /api/v1/media-folders/tree         -- full folder tree (app-side assembly)
```

**Update existing media endpoints:**

- `POST /api/v1/media` -- accept optional `folder_id` in upload
- `PUT /api/v1/media/{id}` -- accept `folder_id` for move
- `POST /api/v1/media/move` -- batch move media to folder (transactional)

**Public content delivery:**

- `GET /api/v1/delivery/media?folder_id=` -- folder-based filtering for galleries/collections

**Permissions:** Reuse existing `media:create`, `media:update`, `media:delete` for folder operations. Folders are organizational metadata, not a separate resource.

**Handler file:** `internal/router/media_folders.go`

**Route registration:** Add to `internal/router/mux.go`

### Phase 4: TUI

Replace URL-derived virtual tree with DB-driven folder tree.

1. **Update `media_tree.go`:**
   - New `MediaTreeNode` source: DB folders instead of URL parsing
   - `BuildMediaTree` takes `(folders []db.MediaFolder, items []db.Media)` instead of just items
   - Folders from DB become `MediaNodeFolder` nodes with a `FolderID` field
   - Media items become `MediaNodeFile` children under their folder
   - Unfiled media (NULL folder_id) appears at root level
   - Keep `splitMediaPath` / URL-parsing code as a fallback for the import tool (Phase migration)

2. **Update `screen_media.go`:**
   - Fetch folders alongside media on `MediaFetchMsg`
   - Store `FolderList []db.MediaFolder` on screen state
   - New keybindings:
     - `n` on a folder node -> create subfolder dialog
     - `n` at root -> create root folder dialog
     - `d` on a folder node -> delete folder (with confirmation, fails if non-empty)
     - `r` on a folder node -> rename folder dialog
     - `m` on a file node -> move to folder (folder picker)
   - Upload respects current folder context (selected folder becomes default destination)

3. **New messages:**
   - `MediaFolderCreateMsg`, `MediaFolderDeleteMsg`, `MediaFolderRenameMsg`
   - `MediaMoveToFolderMsg`
   - Update `MediaFetchResultsMsg` to include folders

4. **Search/filter:**
   - Filter still works on media name/mimetype
   - Folders that contain matching media stay visible (ancestor preservation)

### Phase 5: Admin Panel (HTMX)

1. **Sidebar folder tree** on media list page:
   - Left panel: collapsible folder tree (similar to file manager)
   - Click folder to filter grid to that folder's contents
   - Context menu (action buttons): create subfolder, rename, delete
   - Drag-and-drop media into folders: requires SortableJS (or similar JS library) for drag events, HTMX request on drop. This is a non-trivial UI component -- implement as a `mcms-media-tree` web component.

2. **Folder breadcrumb** on media list:
   - Shows current path: `All Media / Products / Hero Images`
   - Clickable segments to navigate up
   - Uses `GetMediaFolderBreadcrumb()` for ancestor chain

3. **Upload dialog** -- add folder selector:
   - Dropdown or tree picker to choose destination folder
   - Default to currently viewed folder

4. **Media detail page** -- show/edit folder:
   - Display current folder path
   - Dropdown to move to different folder

5. **New templates:**
   - `internal/admin/pages/media_folders.templ` (or extend `media_list.templ`)
   - `internal/admin/partials/media_folder_tree.templ`

6. **New admin routes:**
   ```
   POST   /admin/media/folders           -- create folder
   PUT    /admin/media/folders/{id}       -- rename/move
   DELETE /admin/media/folders/{id}       -- delete (fails if non-empty)
   POST   /admin/media/{id}/move         -- move media to folder
   ```

### Phase 6: MCP Tools

New file `mcp/tools_media_folders.go`:

- `list_media_folders` -- list folders (optionally filtered by parent)
- `get_media_folder` -- get folder details + children
- `create_media_folder` -- create with name + optional parent
- `update_media_folder` -- rename or move
- `delete_media_folder` -- delete (fails if non-empty, returns error with child count)
- `move_media_to_folder` -- move media item(s) to a folder
- Update `upload_media` tool to accept optional `folder_id`

### Phase 7: SDKs

**TypeScript (`sdks/typescript/`):**
- Add `MediaFolder` type to `@modulacms/types`
- Add `MediaFolderID` branded type
- Add `folder_id?: MediaFolderID` to `Media` type
- Add `mediaFolders` resource to admin SDK
- Add `moveMedia(id, folderId)` method
- Add `folder_id` filter param to content delivery SDK's media listing

**Go (`sdks/go/`):**
- Add `MediaFolder` struct to `types.go`
- Add `MediaFolderID` to `ids.go`
- Add `FolderID` to `Media` struct
- Add `MediaFolders` resource to client
- Add `MoveMedia` method

**Swift (`sdks/swift/`):**
- Add `MediaFolder` struct to `Types.swift`
- Add `MediaFolderID` to `IDs.swift`
- Add `folderId` to `Media` struct
- Add `MediaFoldersResource` to client
- Add `moveMedia` method

---

## Migration Strategy

### Fresh installs

Handled by `CreateAllTables`. `CreateMediaFolderTable()` runs before `CreateMediaTable()`. No special action needed.

### Existing installations

On startup, the auto-setup check detects missing tables and columns:

1. `CREATE TABLE IF NOT EXISTS media_folders (...)` -- idempotent, safe to run on existing DBs
2. `ALTER TABLE media ADD COLUMN folder_id ...` -- only if column does not exist
3. Create index on `media.folder_id` for query performance

All existing media starts unfiled (`folder_id = NULL`, appears at root).

### Optional: Import folders from URL structure

A one-time CLI command or TUI action that:
1. Scans all media URLs for common path prefixes
2. Creates DB folders matching the prefix structure
3. Assigns each media item to its matching folder
4. Reports results (folders created, media assigned, unmatched items)

This is a convenience tool, not a migration requirement. Users can also organize manually.

No data loss. No S3 changes. Fully backward compatible.

## Backup & Restore

- Add `media_folders` to the backup export pipeline (`internal/backup/`)
- Add `media_folders` to the backup import/restore pipeline
- Restore order: `media_folders` before `media` (FK dependency)
- Deploy sync (`internal/deploy/`): include `media_folders` in export/import/push/pull operations

## Files Changed (estimated)

| Area | Files | New/Modified |
|------|-------|-------------|
| SQL schema | 6 | New: `sql/schema/36_media_folders/` |
| SQL schema | 6 | Modified: `sql/schema/14_media/` (add folder_id) |
| Types | 1 | Modified: `internal/db/types/types_ids.go` |
| Codegen defs | 2 | Modified: `tools/sqlcgen/definitions.go`, `tools/dbgen/definitions.go` |
| Generated | ~8 | Regenerated: `internal/db-*/`, `internal/db/media_gen.go` |
| DbDriver | 2 | Modified: `internal/db/db.go`, `internal/db/wipe.go` |
| Custom queries | 1 | New: `internal/db/media_folder_custom.go` |
| Batch move | 1 | New or modified: `internal/db/media_custom.go` |
| API handlers | 1 | New: `internal/router/media_folders.go` |
| API handlers | 1 | Modified: `internal/router/media.go` |
| Routes | 1 | Modified: `internal/router/mux.go` |
| RemoteDriver | 1 | Modified: `internal/remote/driver.go` (stub implementations for new DbDriver methods) |
| TUI | 3 | Modified: `screen_media.go`, `media_tree.go`, `screen_media_view.go` |
| Admin templates | 2-3 | Modified/New in `internal/admin/pages/`, `partials/` |
| Admin handlers | 1 | New: `internal/admin/handlers/media_folders.go` |
| Admin JS | 1 | New: `internal/admin/static/js/mcms-media-tree.js` (drag-and-drop web component) |
| MCP | 1 | New: `mcp/tools_media_folders.go` |
| Backup | 2 | Modified: `internal/backup/` (export + import) |
| Deploy | 1 | Modified: `internal/deploy/` (sync operations) |
| SDKs | ~6 | Modified across TypeScript, Go, Swift |
| Tests | 2-3 | New test files |
