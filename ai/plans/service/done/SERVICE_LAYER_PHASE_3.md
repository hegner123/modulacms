# Phase 3: MediaService + RouteService

## Context

Phase 0 established the internal/service/ package with a Registry struct, error types, and audit helpers. Phase 2 implemented ContentService and AdminContentService — fully migrating content CRUD, fields, tree operations, publishing, versioning, batch, and heal to the service layer. Phase 3 extracts MediaService and RouteService — two independent domains that can be built in parallel.

The goal: admin panel, API, and (eventually) MCP all call the same service methods. Business logic (validation, S3 lifecycle, slug uniqueness) lives in one place.

### Current State (as of Phase 2 completion)

**Registry struct** (`internal/service/service.go`):
- Fields: `driver`, `mgr`, `pc`, `emailSvc`, `dispatcher`
- Domain services: `Schema *SchemaService`, `Content *ContentService`, `AdminContent *AdminContentService`
- Getter methods: `Driver()`, `Config()`, `Manager()`, `PermissionCache()`, `EmailService()`, `Dispatcher()` for unmigrated handlers

**Migrated API handlers** (Phase 2 pattern — dispatcher functions taking `*service.Registry`):
- `ContentDatasHandler(w, r, svc *service.Registry)` — contentData.go
- `ContentDataHandler(w, r, svc *service.Registry)` — contentData.go
- `AdminContentDatasHandler(w, r, svc *service.Registry)` — adminContentData.go
- `ContentFieldsHandler(w, r, svc *service.Registry)` — contentFields.go
- `ContentBatchHandler(w, r, svc *service.Registry)` — contentBatch.go
- `PublishHandler(w, r, svc *service.Registry)` — publish.go
- `RestoreVersionHandler(w, r, svc *service.Registry)` — restore.go
- Plus move, reorder, versions handlers

**Not yet migrated** (still use `config.Config` dispatcher pattern):
- Media: `MediasHandler(w, r, c config.Config)`, `MediaHandler(w, r, c config.Config)`
- MediaDimensions: `MediaDimensionsHandler(w, r, c config.Config)`, `MediaDimensionHandler(w, r, c config.Config)`
- Routes: `RoutesHandler(w, r, c config.Config)`, `RouteHandler(w, r, c config.Config)`
- AdminRoutes: `AdminRoutesHandler(w, r, c config.Config)`, `AdminRouteHandler(w, r, c config.Config)`
- Auth, OAuth, health, admin tree, users, roles, permissions, config, plugins, deploy, SSH keys, sessions

**Not yet migrated admin handlers** (still use `(driver, mgr)` closure pattern):
- `MediaListHandler(driver)`, `MediaDetailHandler(driver)`, `MediaUploadHandler(driver, mgr)`, `MediaUpdateHandler(driver, mgr)`, `MediaDeleteHandler(driver, mgr)`
- `RoutesListHandler(driver)`, `AdminRoutesListHandler(driver)`, `RouteCreateHandler(driver, mgr)`, `RouteUpdateHandler(driver, mgr)`, `RouteDeleteHandler(driver, mgr)`

**Mux registration** (`internal/router/mux.go`):
- Migrated content endpoints pass `svc`: `ContentDatasHandler(w, r, svc)`
- Media/route API endpoints pass `*c`: `MediasHandler(w, r, *c)`, `RoutesHandler(w, r, *c)`
- Media/route admin endpoints pass `driver`/`mgr`: `adminhandlers.MediaUploadHandler(driver, mgr)`

**Placeholder stubs** (5 lines each, comment only):
- `internal/service/media.go`
- `internal/service/routes.go`

**Error infrastructure** (fully built in Phase 2):
- `HandleServiceError(w, r, err)` in errors_http.go — maps service errors to JSON or HTMX toast responses
- Checks: `IsNotFound`, `IsValidation`, `IsConflict`, `IsForbidden`
- HTMX detection via `HX-Request` header

---

## Key Design Decisions

MediaService wraps internal/media/, does not replace it. The existing ProcessMediaUpload, HandleMediaUpload, and OptimizeUpload are well-structured with rollback semantics. The service constructs the S3 callback closures and delegates to them.

HandleMediaUpload config coupling — accepted as tech debt. HandleMediaUpload in internal/media/media_upload.go calls db.ConfigDB(c) internally to get its own DB handle, and calls UpdateMedia directly. This is architecturally incoherent with the service layer's injected DbDriver, but it works because ConfigDB returns a cached singleton. Refactoring HandleMediaUpload to accept a DbDriver parameter is out of scope for Phase 3 — it would change the signature used by ProcessMediaUpload and the closure chain. Accepted as tech debt for a future media package cleanup.

Admin upload gains S3 integration — S3 is now required for uploads. Currently admin creates DB-only placeholder records (URL /uploads/filename) with no S3 upload. After Phase 3, all uploads go through ProcessMediaUpload which requires S3 closures. This is an intentional policy change: if S3 is not configured, Upload() returns a ValidationError with field "s3" and message "S3 storage must be configured for media uploads". Existing placeholder records with /uploads/ URLs remain in the DB unchanged — they are not migrated.

Admin delete gains S3 cleanup. Currently admin does DB-only delete. The service always cleans up S3 objects + DB record.

DeleteMedia follows best-effort S3 cleanup. Matches current API behavior: S3 deletion failures are logged as warnings but do not block DB record deletion. This can create S3 orphans, which MediaHealth/MediaCleanup detect and remove. Rationale: failing the entire delete because one S3 variant failed leaves the user unable to remove the media record at all.

Route "rename propagation" is a non-issue. Content references routes via RouteID (ULID), not slug. Slug renames don't break relations. The service just validates slug uniqueness on rename.

Route slug has a UNIQUE constraint on all three backends (SQLite, MySQL, PostgreSQL). The service's pre-check via `driver.GetRouteID(slug)` is a courtesy that provides a clean ConflictError message. The DB constraint is the actual safety net — if a TOCTOU race slips past the check, the DB rejects the duplicate and the service maps the constraint violation to ConflictError.

MCP stays as-is. It calls the Go SDK over HTTP. Rewiring to call services directly is Phase 7.

Services are separate structs attached as exported fields on Registry (matching ContentService/AdminContentService pattern from Phase 2). MediaService and RouteService each get their own struct with injected dependencies, initialized in NewRegistry.

Service methods return db.* types directly. No service-layer DTOs for Phase 3. This is acceptable because admin handlers and API handlers both already work with db.* types. Phase 7 (MCP direct calls) may introduce service-layer types if needed.

API handlers adopt the Phase 2 dispatcher pattern: `func FooHandler(w, r, svc *service.Registry)` — matching ContentDatasHandler, not closure factories. This eliminates the db.ConfigDB(c) singleton usage in these handlers.

Admin handlers adopt closure-factory pattern with svc: `func FooHandler(svc *service.Registry) http.HandlerFunc` — matching the existing admin handler convention but replacing `(driver, mgr)` with `(svc)`.

Two handler patterns coexist during transition. After Phase 3, media and route handlers (both admin and API) use *service.Registry. All other handlers remain unchanged (config.Config for API, driver/mgr for admin). This is expected during incremental migration across phases.

Audit context pattern (established in Phase 2): handlers construct `audited.AuditContext` via `middleware.AuditContextFromRequest(r, *cfg)` where `cfg` comes from `svc.Config()`. Service methods receive `ac audited.AuditContext` as a parameter — they never construct it themselves. Handlers that need config for audit context or other purposes call `svc.Config()` (returns `(*config.Config, error)`).

MediaReferencesHandler and clean_refs stay as handler-level concerns. `MediaReferencesHandler` gets its signature rewired to `(w, r, svc)` and calls `svc.Driver()` for the reference scan (unmigrated helper pattern). The `clean_refs=true` option on delete stays in the API handler: the handler calls `svc.Driver()` for the reference scan, then `svc.Media.DeleteMedia()` for the S3 + DB cleanup. The service's `DeleteMedia` does not absorb reference cleaning — it only handles S3 + DB.

---

## MediaService — internal/service/media.go

### Struct

```go
type MediaService struct {
    driver db.DbDriver
    mgr    *config.Manager
}

func NewMediaService(driver db.DbDriver, mgr *config.Manager) *MediaService
```

### Param Types

- UploadMediaParams — File, Header, Path, Alt, Caption, Description, DisplayName
- UpdateMediaParams — MediaID, DisplayName, Alt, Caption, Description, FocalX, FocalY
- MediaListParams — Limit, Offset
- OrphanScanResult — TotalObjects, TrackedKeys, OrphanedKeys
- OrphanCleanupResult — Deleted, Failed

### Methods

| Method | Logic Beyond DbDriver |
|--------|----------------------|
| Upload(ctx, ac, UploadMediaParams) | Call `m.mgr.Config()` to get config snapshot (error -> InternalError). Validate S3 config present (missing -> ValidationError field "s3"). Sanitize path, build S3 closures from config snapshot. Delegate to `media.ProcessMediaUpload(m.driver, ...)` — `db.DbDriver` satisfies the `media.MediaStore` interface, so `m.driver` passes directly. Pipeline closure calls `media.HandleMediaUpload(srcFile, dstPath, *cfg)` with dereferenced config. Map media errors: `media.DuplicateMediaError` -> `ConflictError{Resource: "media"}`, `media.FileTooLargeError` -> `ValidationError` field "file". Must be synchronous — UploadMediaParams.File is a multipart.File valid only during the request. |
| GetMedia(ctx, MediaID) | NotFoundError mapping |
| ListMedia(ctx) | Passthrough |
| ListMediaPaginated(ctx, MediaListParams) | Returns items + count |
| UpdateMediaMetadata(ctx, ac, UpdateMediaParams) | Fetch existing, overlay non-empty fields, validate focal point range [0,1], set DateModified |
| DeleteMedia(ctx, ac, MediaID) | Fetch record, extract S3 keys from URL+srcset, delete S3 objects (best-effort), delete DB record. Does NOT handle reference cleanup (clean_refs) — that stays in the handler layer. |
| MediaHealth(ctx) | S3 orphan scan (move findOrphanedMediaKeys from router) |
| MediaCleanup(ctx) | S3 orphan deletion |
| ListMediaDimensions(ctx) | Passthrough |
| GetMediaDimension(ctx, id) | NotFoundError mapping |
| CreateMediaDimension(ctx, ac, params) | Validate width/height positive, label non-empty |
| UpdateMediaDimension(ctx, ac, params) | Validate width/height positive |
| DeleteMediaDimension(ctx, ac, id) | NotFoundError mapping |

### Private Helpers

- newS3Session() — creates S3 client from current config
- extractMediaS3Keys(record, cfg) — parses URL + srcset JSON to S3 keys
- findOrphanedKeys(driver, s3Session, cfg) — moved from internal/router/media.go

### Integration with internal/media/

Handler -> Registry.Media.Upload() -> media.ProcessMediaUpload(m.driver, ...) -> uploadOriginal closure (built by service from config) -> rollbackS3 closure (built by service) -> media.HandleMediaUpload(srcFile, dstPath, *cfg)

Key details:
- `Upload()` calls `m.mgr.Config()` once at the top to get a `*config.Config` snapshot. All S3 credentials, bucket settings, max upload size, media directory, and public URL come from this snapshot. Error from `Config()` returns as `InternalError`.
- `db.DbDriver` satisfies the `media.MediaStore` interface, so `m.driver` passes directly to `ProcessMediaUpload` with no adapter.
- `HandleMediaUpload` takes `config.Config` by value (not pointer): pass `*cfg` where `cfg` is the `*config.Config` from `m.mgr.Config()`.
- The three closures are constructed inside Upload() using S3 session and config — identical logic to current apiCreateMedia in internal/router/media.go lines 55-186, extracted once.

---

## RouteService — internal/service/routes.go

### Struct

```go
type RouteService struct {
    driver db.DbDriver
    mgr    *config.Manager
}

func NewRouteService(driver db.DbDriver, mgr *config.Manager) *RouteService
```

### Param Types

- CreateRouteInput — Slug, Title, Status, AuthorID
- UpdateRouteInput — RouteID, Slug, Title, Status, AuthorID

Note on Slug_2: The audited UpdateRouteCmd uses a Slug_2 field (the old slug) as the WHERE clause identifier and for GetID() in the audit trail. The service's UpdateRoute fetches the existing route by ID, extracts its current slug, and populates Slug_2 in db.UpdateRouteParams from the fetched record. The caller never provides Slug_2 — it is internal to the service. Race condition: if another request changes the slug between the fetch and the update, the UPDATE WHERE clause matches 0 rows. The audited command returns this as an error (affected rows = 0). The service should return NotFoundError in this case — the caller can retry.

### Methods

| Method | Logic Beyond DbDriver |
|--------|----------------------|
| CreateRoute(ctx, ac, CreateRouteInput) | Validate slug format + title required, slug uniqueness check, auto-set timestamps |
| UpdateRoute(ctx, ac, UpdateRouteInput) | Validate slug + title, fetch existing by ID, slug uniqueness if changed (excluding self), preserve DateCreated/AuthorID, set DateModified |
| DeleteRoute(ctx, ac, RouteID) | Check no content references this route -> ConflictError if content exists |
| GetRoute(ctx, RouteID) | NotFoundError mapping |
| ListRoutes(ctx) | Passthrough |
| ListRoutesPaginated(ctx, limit, offset) | Returns items + count |
| CreateAdminRoute(ctx, ac, CreateRouteInput) | Same validation as CreateRoute but uses AdminRouteID types and admin route DB methods |
| UpdateAdminRoute(ctx, ac, UpdateRouteInput) | Same as UpdateRoute but with admin route types. Slug_2 pattern applies identically. |
| DeleteAdminRoute(ctx, ac, AdminRouteID) | Cascade check against admin content, then delete |
| GetAdminRoute(ctx, AdminRouteID) | NotFoundError mapping |
| GetAdminRouteBySlug(ctx, slug) | Lookup by slug (admin routes use slug-based GET), NotFoundError mapping |
| ListAdminRoutes(ctx) | Passthrough |
| ListAdminRoutesPaginated(ctx, limit, offset) | Returns items + count |
| ListOrderedAdminRoutes(ctx) | Calls ListAdminRoutes then sorts into tree order in-memory. The sorting logic (parent-child ordering) moves from apiListOrderedAdminRoutes in adminRoutes.go into this service method. |

### Validation Fixes (currently missing)

1. **Slug uniqueness** — neither admin nor API currently check. Service calls `driver.GetRouteID(string(slug))` before create/update. This returns `(*types.RouteID, error)`. If `err` wraps `sql.ErrNoRows`, the slug is available. If `err` is nil, a route with that slug already exists — return ConflictError. Any other error (transient DB failure) is returned as InternalError — never treated as "available." On update, exclude self: if the returned RouteID matches the route being updated, the slug is unchanged and no conflict. The UNIQUE constraint is the real safety net; the service check provides a user-friendly ConflictError message.
2. **Cascade check on delete** — currently nothing prevents deleting a route with content attached. Service calls `driver.ListContentDataByRoute(types.NullableRouteID{ID: id, Valid: true})` and checks `len(*result) > 0`. Note: `ListContentDataByRoute` takes `types.NullableRouteID`, not `types.RouteID` — the ID must be wrapped. No CountContentDataByRoute exists in DbDriver. This loads full rows, which is wasteful for routes with many content items. Accepted for Phase 3 — add a count query in a future DB maintenance pass if profiling shows it matters.
3. **Consistent validation** — admin validates slug format (via `types.Slug(slug).Validate()`) but API validates nothing. Service always validates.
4. **Consistent timestamps** — admin uses time.Now(), API uses time.Now().UTC(). Service standardizes on UTC.

### Private Helper

- validateRouteInput(slug, title string) *ValidationError — shared by create and update

---

## Handler Rewiring

### Admin Handlers

internal/admin/handlers/media.go — all signatures change from `(driver db.DbDriver)` / `(driver db.DbDriver, mgr *config.Manager)` to `(svc *service.Registry)`:

| Handler | Before | After |
|---------|--------|-------|
| MediaUploadHandler | Parse form, MIME default, create DB placeholder (no S3) | Parse form -> svc.Media.Upload(ctx, ac, params) |
| MediaUpdateHandler | Parse form, manual field overlay, DB update | Parse form -> svc.Media.UpdateMediaMetadata(ctx, ac, params) |
| MediaDeleteHandler | DB-only delete (no S3 cleanup) | svc.Media.DeleteMedia(ctx, ac, id) (gains S3 cleanup) |
| MediaListHandler | driver.ListMediaPaginated() | svc.Media.ListMediaPaginated(ctx, params) |
| MediaDetailHandler | driver.GetMedia() | svc.Media.GetMedia(ctx, id) |

internal/admin/handlers/routes.go — all signatures change from `(driver db.DbDriver)` / `(driver db.DbDriver, mgr *config.Manager)` to `(svc *service.Registry)`:

| Handler | Before | After |
|---------|--------|-------|
| RouteCreateHandler | Manual slug validation via types.Slug, no uniqueness check | Parse form -> svc.Routes.CreateRoute(ctx, ac, input) |
| RouteUpdateHandler | Manual validation, no uniqueness check | Parse form -> svc.Routes.UpdateRoute(ctx, ac, input) |
| RouteDeleteHandler | No cascade check | svc.Routes.DeleteRoute(ctx, ac, id) (gains cascade safety) |
| RoutesListHandler | driver.ListRoutesPaginated() | svc.Routes.ListRoutesPaginated(ctx, limit, offset) |
| AdminRoutesListHandler | driver.ListAdminRoutes() | svc.Routes.ListAdminRoutes(ctx) |

Service errors map to HTMX responses via existing HandleServiceError:
- IsValidation -> render form with field errors
- IsConflict -> toast "slug already exists" or "route has content"
- IsNotFound -> toast "not found"

### API Handlers

internal/router/media.go — change from `(w, r, c config.Config)` dispatcher to `(w, r, svc *service.Registry)` dispatcher (matching Phase 2 content pattern):

| Handler | Key Change |
|---------|-----------|
| MediasHandler(w, r, svc) | Signature change, dispatch to service methods |
| MediaHandler(w, r, svc) | Signature change, dispatch to service methods |
| apiCreateMedia | Remove ~80 lines of S3 setup/closure construction -> svc.Media.Upload(ctx, ac, params) |
| apiUpdateMedia | svc.Media.UpdateMediaMetadata(ctx, ac, params) |
| apiDeleteMedia | Remove S3 key reconstruction + deletion -> svc.Media.DeleteMedia(ctx, ac, id). Retain clean_refs logic in handler: if `clean_refs=true`, call `svc.Driver()` for reference scan before calling DeleteMedia. |
| MediaHealthHandler | Remove orphan logic -> svc.Media.MediaHealth(ctx) |
| MediaCleanupHandler | Remove orphan logic -> svc.Media.MediaCleanup(ctx) |
| MediaReferencesHandler | Signature rewires to (w, r, svc). Calls `svc.Driver()` for the reference scan (unmigrated helper pattern). |
| apiListMedia, apiGetMedia, apiListMediaPaginated | Use service read methods |

internal/router/mediaDimensions.go — change from `(w, r, c config.Config)` to `(w, r, svc *service.Registry)`.

internal/router/routes.go — change from `(w, r, c config.Config)` to `(w, r, svc *service.Registry)`. Gains validation that API currently lacks.

internal/router/adminRoutes.go — change from `(w, r, c config.Config)` to `(w, r, svc *service.Registry)`. Contains 9 functions: `AdminRoutesHandler`, `AdminRouteHandler`, `apiGetAdminRoute`, `apiListAdminRoutes`, `apiListOrderedAdminRoutes`, `apiCreateAdminRoute`, `apiUpdateAdminRoute`, `apiDeleteAdminRoute`, `apiListAdminRoutesPaginated`. The `apiListOrderedAdminRoutes` in-memory tree sorting logic moves into `svc.Routes.ListOrderedAdminRoutes()`. Admin route types differ from public routes: uses `db.AdminRoutes`, `types.AdminRouteID`, and slug-based GET lookup.

### Mux Wiring (internal/router/mux.go)

Update handler registrations to pass `svc` instead of `*c` or `(driver, mgr)`:

**API routes** (currently pass `*c`, change to `svc`):
```go
// Before
MediasHandler(w, r, *c)
MediaHandler(w, r, *c)
MediaDimensionsHandler(w, r, *c)
MediaDimensionHandler(w, r, *c)
RoutesHandler(w, r, *c)
RouteHandler(w, r, *c)
AdminRoutesHandler(w, r, *c)
AdminRouteHandler(w, r, *c)
MediaHealthHandler(w, r, *c)
MediaCleanupHandler(w, r, *c)
MediaReferencesHandler(w, r, *c)

// After
MediasHandler(w, r, svc)
MediaHandler(w, r, svc)
// ... etc
```

**Admin routes** (currently pass `driver`/`mgr`, change to `svc`):
```go
// Before
adminhandlers.MediaListHandler(driver)
adminhandlers.MediaUploadHandler(driver, mgr)
adminhandlers.RouteCreateHandler(driver, mgr)

// After
adminhandlers.MediaListHandler(svc)
adminhandlers.MediaUploadHandler(svc)
adminhandlers.RouteCreateHandler(svc)
```

---

## Registry Changes

Add Media and Routes fields to Registry struct and initialize in NewRegistry:

```go
type Registry struct {
    // ... existing fields ...

    Schema       *SchemaService
    Content      *ContentService
    AdminContent *AdminContentService
    Media        *MediaService        // Phase 3
    Routes       *RouteService        // Phase 3
}

func NewRegistry(...) *Registry {
    reg := &Registry{...}
    reg.Schema = NewSchemaService(driver, driver)
    reg.Content = NewContentService(driver, mgr, dispatcher)
    reg.AdminContent = NewAdminContentService(driver, mgr, dispatcher)
    reg.Media = NewMediaService(driver, mgr)     // Phase 3
    reg.Routes = NewRouteService(driver, mgr)    // Phase 3
    return reg
}
```

---

## Files Changed

| File | Type | Scope |
|------|------|-------|
| internal/service/service.go | Moderate | Add Media + Routes fields to Registry, initialize in NewRegistry |
| internal/service/media.go | Rewrite | Replace 5-line stub with full MediaService struct (~300 lines) |
| internal/service/routes.go | Rewrite | Replace 5-line stub with full RouteService struct (~200 lines) |
| internal/admin/handlers/media.go | Major | Change all signatures from (driver, mgr) to (svc), rewrite upload/update/delete |
| internal/admin/handlers/routes.go | Major | Change all signatures from (driver, mgr) to (svc), rewrite create/update/delete |
| internal/router/media.go | Major | Change from config.Config to svc, remove S3 logic + orphan helpers |
| internal/router/mediaDimensions.go | Moderate | Change from config.Config to svc |
| internal/router/routes.go | Moderate | Change from config.Config to svc, gains validation |
| internal/router/adminRoutes.go | Major | Change from config.Config to svc, 9 functions rewired, tree sorting logic moves to service |
| internal/router/mux.go | Moderate | Update ~15 handler registrations to pass svc instead of *c / (driver, mgr) |
| internal/service/media_test.go | New | Service-level tests |
| internal/service/routes_test.go | New | Service-level tests |

Not changed: internal/media/*.go (reused as-is), internal/bucket/*.go (reused), mcp/ (Phase 7).

---

## Testing

### Service Tests (SQLite, no S3 needed)

internal/service/routes_test.go:
- Create with valid input succeeds
- Create with empty slug -> ValidationError
- Create with invalid slug format -> ValidationError
- Create with duplicate slug -> ConflictError
- Update rename to existing slug -> ConflictError
- Update same slug (no rename) passes
- Delete route with content -> ConflictError
- Delete route without content succeeds

internal/service/media_test.go:
- Upload with no S3 config -> ValidationError
- Upload with invalid path -> ValidationError
- UpdateMetadata with focal point > 1.0 -> ValidationError
- UpdateMetadata overlays only non-empty fields
- Delete non-existent -> NotFoundError
- Dimension create with zero width -> ValidationError

### Integration Tests (MinIO required for media upload)

- Full upload via service -> verify S3 object exists + DB record correct
- Admin upload path (via service) -> verify S3 upload + DB record (was previously DB-only placeholder)
- Upload with no S3 config -> verify ValidationError returned
- Delete -> verify S3 objects removed + DB record gone
- Delete with S3 failure -> verify DB record still deleted (best-effort), warning logged
- Health -> verify orphan detection
- Cleanup -> verify orphan removal

### Verification

```
go build ./...              # compile check
just test                   # unit tests pass
just test-minio && just test-integration  # media S3 integration
```

---

## Implementation Order

These two services are independent and can be built by parallel agents after a shared prerequisite.

### Step 0 — Registry Wiring (prerequisite, do first):

Add both `Media` and `Routes` fields to the Registry struct in `internal/service/service.go` and update `NewRegistry` to initialize them. This must be done once before parallel streams start, to avoid merge conflicts on the same file. Both `NewMediaService` and `NewRouteService` can initially be stubs that return empty structs — the parallel streams will fill in the implementations.

### Stream A — RouteService (simpler, ~1 session):
1. Implement internal/service/routes.go (RouteService struct + all methods including admin route variants)
2. Write internal/service/routes_test.go
3. Rewire internal/admin/handlers/routes.go signatures to (svc)
4. Rewire internal/router/routes.go + adminRoutes.go to (w, r, svc) pattern
5. Update mux.go registrations for route handlers

### Stream B — MediaService (more complex, ~1-2 sessions):
1. Implement internal/service/media.go (MediaService struct + methods)
2. Write internal/service/media_test.go
3. Rewire internal/admin/handlers/media.go signatures to (svc)
4. Rewire internal/router/media.go + mediaDimensions.go to (w, r, svc) pattern
5. Update mux.go registrations for media handlers

Final: Build + test both together to verify no conflicts in mux wiring.
