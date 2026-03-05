Phase 3: MediaService + RouteService                                                                                                                                                   │
     │                                                                                                                                                                                        │
     │ Context                                                                                                                                                                                │
     │                                                                                                                                                                                        │
     │ Phase 0 established the internal/service/ package with a Registry struct, error types, and audit helpers. Phases 1-2 (schema, content) are planned but not yet implemented. Phase 3    │
     │ extracts MediaService and RouteService — two independent domains that can be built in parallel.                                                                                        │
     │                                                                                                                                                                                        │
     │ The goal: admin panel, API, and (eventually) MCP all call the same service methods. Business logic (validation, S3 lifecycle, slug uniqueness) lives in one place.                     │
     │                                                                                                                                                                                        │
     │ ---                                                                                                                                                                                    │
     │ Key Design Decisions                                                                                                                                                                   │
     │                                                                                                                                                                                        │
     │ MediaService wraps internal/media/, does not replace it. The existing ProcessMediaUpload, HandleMediaUpload, and OptimizeUpload are well-structured with rollback semantics. The       │
     │ service constructs the S3 callback closures and delegates to them.                                                                                                                     │
     │                                                                                                                                                                                        │
     │ HandleMediaUpload config coupling — accepted as tech debt. HandleMediaUpload in internal/media/media_upload.go calls db.ConfigDB(c) internally to get its own DB handle, and calls     │
     │ UpdateMedia directly. This is architecturally incoherent with the service layer's injected DbDriver, but it works because ConfigDB returns a cached singleton. Refactoring             │
     │ HandleMediaUpload to accept a DbDriver parameter is out of scope for Phase 3 — it would change the signature used by ProcessMediaUpload and the closure chain. Accepted as tech debt   │
     │ for a future media package cleanup.                                                                                                                                                    │
     │                                                                                                                                                                                        │
     │ Admin upload gains S3 integration — S3 is now required for uploads. Currently admin creates DB-only placeholder records (URL /uploads/filename) with no S3 upload. After Phase 3, all  │
     │ uploads go through ProcessMediaUpload which requires S3 closures. This is an intentional policy change: if S3 is not configured, Upload() returns a ValidationError with field "s3"    │
     │ and message "S3 storage must be configured for media uploads". Existing placeholder records with /uploads/ URLs remain in the DB unchanged — they are not migrated.                    │
     │                                                                                                                                                                                        │
     │ Admin delete gains S3 cleanup. Currently admin does DB-only delete. The service always cleans up S3 objects + DB record.                                                               │
     │                                                                                                                                                                                        │
     │ DeleteMedia follows best-effort S3 cleanup. Matches current API behavior: S3 deletion failures are logged as warnings but do not block DB record deletion. This can create S3 orphans, │
     │  which MediaHealth/MediaCleanup detect and remove. Rationale: failing the entire delete because one S3 variant failed leaves the user unable to remove the media record at all.        │
     │                                                                                                                                                                                        │
     │ Route "rename propagation" is a non-issue. Content references routes via RouteID (ULID), not slug. Slug renames don't break relations. The service just validates slug uniqueness on   │
     │ rename.                                                                                                                                                                                │
     │                                                                                                                                                                                        │
     │ Route slug has a UNIQUE constraint on all three backends (SQLite, MySQL, PostgreSQL). The service's pre-check via GetRouteBySlug is a courtesy that provides a clean ConflictError     │
     │ message. The DB constraint is the actual safety net — if a TOCTOU race slips past the check, the DB rejects the duplicate and the service maps the constraint violation to             │
     │ ConflictError.                                                                                                                                                                         │
     │                                                                                                                                                                                        │
     │ MCP stays as-is. It calls the Go SDK over HTTP. Rewiring to call services directly is Phase 7.                                                                                         │
     │                                                                                                                                                                                        │
     │ Services are methods on *Registry, not separate structs. Matches Phase 0 conventions.                                                                                                  │
     │                                                                                                                                                                                        │
     │ Service methods return db.* types directly. No service-layer DTOs for Phase 3. This is acceptable because admin handlers and API handlers both already work with db.* types. Phase 7   │
     │ (MCP direct calls) may introduce service-layer types if needed.                                                                                                                        │
     │                                                                                                                                                                                        │
     │ API handlers adopt closure-factory pattern. Currently API handlers are bare functions func(w, r, config.Config). After Phase 3, media and route API handlers become closure factories  │
     │ func Foo(svc *service.Registry) http.HandlerFunc — matching the admin handler convention. This eliminates the db.ConfigDB(c) singleton usage in these handlers.                        │
     │                                                                                                                                                                                        │
     │ Two handler patterns coexist during transition. After Phase 3, media and route handlers (both admin and API) use *service.Registry. All other handlers remain unchanged (config.Config │
     │  for API, driver/mgr for admin). This is expected during incremental migration across phases.                                                                                          │
     │                                                                                                                                                                                        │
     │ ---                                                                                                                                                                                    │
     │ MediaService — internal/service/media.go                                                                                                                                               │
     │                                                                                                                                                                                        │
     │ Param Types                                                                                                                                                                            │
     │                                                                                                                                                                                        │
     │ UploadMediaParams     — File, Header, Path, Alt, Caption, Description, DisplayName                                                                                                     │
     │ UpdateMediaParams     — MediaID, DisplayName, Alt, Caption, Description, FocalX, FocalY                                                                                                │
     │ MediaListParams       — Limit, Offset                                                                                                                                                  │
     │ OrphanScanResult      — TotalObjects, TrackedKeys, OrphanedKeys                                                                                                                        │
     │ OrphanCleanupResult   — Deleted, Failed                                                                                                                                                │
     │                                                                                                                                                                                        │
     │ Methods                                                                                                                                                                                │
     │                                                                                                                                                                                        │
     │ ┌─────────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┐                                                                                                                                                                                   │
     │ │                 Method                  │                                                            Logic Beyond DbDriver                                                           │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ Upload(ctx, ac, UploadMediaParams)      │ Validate S3 config, sanitize path, build S3 closures, delegate to media.ProcessMediaUpload, map error types. Must be synchronous —         │
     │    │                                                                                                                                                                                   │
     │ │                                         │ UploadMediaParams.File is a multipart.File valid only during the request.                                                                  │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ GetMedia(ctx, MediaID)                  │ NotFoundError mapping                                                                                                                      │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ ListMedia(ctx)                          │ Passthrough                                                                                                                                │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ ListMediaPaginated(ctx,                 │ Returns items + count                                                                                                                      │
     │    │                                                                                                                                                                                   │
     │ │ MediaListParams)                        │                                                                                                                                            │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ UpdateMediaMetadata(ctx, ac,            │ Fetch existing, overlay non-empty fields, validate focal point range [0,1], set DateModified                                               │
     │    │                                                                                                                                                                                   │
     │ │ UpdateMediaParams)                      │                                                                                                                                            │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ DeleteMedia(ctx, ac, MediaID)           │ Fetch record, extract S3 keys from URL+srcset, delete S3 objects, delete DB record                                                         │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ MediaHealth(ctx)                        │ S3 orphan scan (move findOrphanedMediaKeys from router)                                                                                    │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ MediaCleanup(ctx)                       │ S3 orphan deletion                                                                                                                         │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ ListMediaDimensions(ctx)                │ Passthrough                                                                                                                                │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ GetMediaDimension(ctx, id)              │ NotFoundError mapping                                                                                                                      │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ CreateMediaDimension(ctx, ac, params)   │ Validate width/height positive, label non-empty                                                                                            │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ UpdateMediaDimension(ctx, ac, params)   │ Validate width/height positive                                                                                                             │
     │    │                                                                                                                                                                                   │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┤                                                                                                                                                                                   │
     │ │ DeleteMediaDimension(ctx, ac, id)       │ NotFoundError mapping                                                                                                                      │
     │    │                                                                                                                                                                                   │
     │ └─────────────────────────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ───┘                                                                                                                                                                                   │
     │                                                                                                                                                                                        │
     │ Private Helpers                                                                                                                                                                        │
     │                                                                                                                                                                                        │
     │ - newS3Session() — creates S3 client from current config                                                                                                                               │
     │ - extractMediaS3Keys(record, cfg) — parses URL + srcset JSON to S3 keys                                                                                                                │
     │ - findOrphanedKeys(driver, s3Session, cfg) — moved from internal/router/media.go                                                                                                       │
     │                                                                                                                                                                                        │
     │ Integration with internal/media/                                                                                                                                                       │
     │                                                                                                                                                                                        │
     │ Handler → Registry.Upload() → media.ProcessMediaUpload()                                                                                                                               │
     │                                 → uploadOriginal closure (built by service from config)                                                                                                │
     │                                 → rollbackS3 closure (built by service)                                                                                                                │
     │                                 → media.HandleMediaUpload (pipeline closure)                                                                                                           │
     │                                                                                                                                                                                        │
     │ The three closures are constructed inside Upload() using S3 session and config — identical logic to current apiCreateMedia in internal/router/media.go lines 148-186, extracted once.  │
     │                                                                                                                                                                                        │
     │ ---                                                                                                                                                                                    │
     │ RouteService — internal/service/routes.go                                                                                                                                              │
     │                                                                                                                                                                                        │
     │ Param Types                                                                                                                                                                            │
     │                                                                                                                                                                                        │
     │ CreateRouteInput  — Slug, Title, Status, AuthorID                                                                                                                                      │
     │ UpdateRouteInput  — RouteID, Slug, Title, Status, AuthorID                                                                                                                             │
     │                                                                                                                                                                                        │
     │ Note on Slug_2: The audited UpdateRouteCmd uses a Slug_2 field (the old slug) as the WHERE clause identifier and for GetID() in the audit trail. The service's UpdateRoute fetches the │
     │  existing route by ID, extracts its current slug, and populates Slug_2 in db.UpdateRouteParams from the fetched record. The caller never provides Slug_2 — it is internal to the       │
     │ service.                                                                                                                                                                               │
     │                                                                                                                                                                                        │
     │ Methods                                                                                                                                                                                │
     │                                                                                                                                                                                        │
     │ ┌─────────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ┐                                                                                                                                                                                      │
     │ │                 Method                  │                                                           Logic Beyond DbDriver                                                            │
     │ │                                                                                                                                                                                      │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ┤                                                                                                                                                                                      │
     │ │ CreateRoute(ctx, ac, CreateRouteInput)  │ Validate slug format + title required, slug uniqueness check, auto-set timestamps                                                          │
     │ │                                                                                                                                                                                      │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ┤                                                                                                                                                                                      │
     │ │ UpdateRoute(ctx, ac, UpdateRouteInput)  │ Validate slug + title, fetch existing by ID, slug uniqueness if changed (excluding self), preserve DateCreated/AuthorID, set DateModified  │
     │ │                                                                                                                                                                                      │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ┤                                                                                                                                                                                      │
     │ │ DeleteRoute(ctx, ac, RouteID)           │ Check no content references this route → ConflictError if content exists                                                                   │
     │ │                                                                                                                                                                                      │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ┤                                                                                                                                                                                      │
     │ │ GetRoute(ctx, RouteID)                  │ NotFoundError mapping                                                                                                                      │
     │ │                                                                                                                                                                                      │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ┤                                                                                                                                                                                      │
     │ │ ListRoutes(ctx)                         │ Passthrough                                                                                                                                │
     │ │                                                                                                                                                                                      │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ┤                                                                                                                                                                                      │
     │ │ ListRoutesPaginated(ctx, limit, offset) │ Returns items + count                                                                                                                      │
     │ │                                                                                                                                                                                      │
     │ ├─────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ┤                                                                                                                                                                                      │
     │ │ Admin variants                          │ Same pattern with AdminRouteID types                                                                                                       │
     │ │                                                                                                                                                                                      │
     │ └─────────────────────────────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── │
     │ ┘                                                                                                                                                                                      │
     │                                                                                                                                                                                        │
     │ Validation Fixes (currently missing)                                                                                                                                                   │
     │                                                                                                                                                                                        │
     │ 1. Slug uniqueness — neither admin nor API currently check. Service calls driver.GetRouteBySlug(slug) before create/update. Error handling: sql.ErrNoRows means slug is available. Any │
     │  other error (transient DB failure) is returned as InternalError — never treated as "available." The UNIQUE constraint is the real safety net; the service check provides a            │
     │ user-friendly ConflictError message.                                                                                                                                                   │
     │ 2. Cascade check on delete — currently nothing prevents deleting a route with content attached. Service calls driver.ListContentDataByRoute() and checks len(*result) > 0. No          │
     │ CountContentDataByRoute exists in DbDriver. This loads full rows, which is wasteful for routes with many content items. Accepted for Phase 3 — add a count query in a future DB        │
     │ maintenance pass if profiling shows it matters.                                                                                                                                        │
     │ 3. Consistent validation — admin validates slug format but API validates nothing. Service always validates.                                                                            │
     │ 4. Consistent timestamps — admin uses time.Now(), API uses time.Now().UTC(). Service standardizes on UTC.                                                                              │
     │                                                                                                                                                                                        │
     │ Private Helper                                                                                                                                                                         │
     │                                                                                                                                                                                        │
     │ - validateRouteInput(slug, title string) *ValidationError — shared by create and update                                                                                                │
     │                                                                                                                                                                                        │
     │ ---                                                                                                                                                                                    │
     │ Handler Rewiring                                                                                                                                                                       │
     │                                                                                                                                                                                        │
     │ Admin Handlers                                                                                                                                                                         │
     │                                                                                                                                                                                        │
     │ internal/admin/handlers/media.go — all signatures change from (driver, mgr) to (svc *service.Registry):                                                                                │
     │                                                                                                                                                                                        │
     │ ┌────────────────────┬─────────────────────────────────────────────────┬───────────────────────────────────────────────────────┐                                                       │
     │ │      Handler       │                     Before                      │                         After                         │                                                       │
     │ ├────────────────────┼─────────────────────────────────────────────────┼───────────────────────────────────────────────────────┤                                                       │
     │ │ MediaUploadHandler │ Parse form, MIME default, create DB placeholder │ Parse form → svc.Upload(ctx, ac, params)              │                                                       │
     │ ├────────────────────┼─────────────────────────────────────────────────┼───────────────────────────────────────────────────────┤                                                       │
     │ │ MediaUpdateHandler │ Parse form, manual field overlay, DB update     │ Parse form → svc.UpdateMediaMetadata(ctx, ac, params) │                                                       │
     │ ├────────────────────┼─────────────────────────────────────────────────┼───────────────────────────────────────────────────────┤                                                       │
     │ │ MediaDeleteHandler │ DB-only delete                                  │ svc.DeleteMedia(ctx, ac, id) (gains S3 cleanup)       │                                                       │
     │ ├────────────────────┼─────────────────────────────────────────────────┼───────────────────────────────────────────────────────┤                                                       │
     │ │ MediaListHandler   │ driver.ListMediaPaginated()                     │ svc.ListMediaPaginated(ctx, params)                   │                                                       │
     │ ├────────────────────┼─────────────────────────────────────────────────┼───────────────────────────────────────────────────────┤                                                       │
     │ │ MediaDetailHandler │ driver.GetMedia()                               │ svc.GetMedia(ctx, id)                                 │                                                       │
     │ └────────────────────┴─────────────────────────────────────────────────┴───────────────────────────────────────────────────────┘                                                       │
     │                                                                                                                                                                                        │
     │ internal/admin/handlers/routes.go — all signatures change from (driver, mgr) to (svc *service.Registry):                                                                               │
     │                                                                                                                                                                                        │
     │ ┌────────────────────┬─────────────────────────────────────────────┬─────────────────────────────────────────────────────┐                                                             │
     │ │      Handler       │                   Before                    │                        After                        │                                                             │
     │ ├────────────────────┼─────────────────────────────────────────────┼─────────────────────────────────────────────────────┤                                                             │
     │ │ RouteCreateHandler │ Manual slug validation, no uniqueness check │ Parse form → svc.CreateRoute(ctx, ac, input)        │                                                             │
     │ ├────────────────────┼─────────────────────────────────────────────┼─────────────────────────────────────────────────────┤                                                             │
     │ │ RouteUpdateHandler │ Manual validation, no uniqueness check      │ Parse form → svc.UpdateRoute(ctx, ac, input)        │                                                             │
     │ ├────────────────────┼─────────────────────────────────────────────┼─────────────────────────────────────────────────────┤                                                             │
     │ │ RouteDeleteHandler │ No cascade check                            │ svc.DeleteRoute(ctx, ac, id) (gains cascade safety) │                                                             │
     │ ├────────────────────┼─────────────────────────────────────────────┼─────────────────────────────────────────────────────┤                                                             │
     │ │ RoutesListHandler  │ driver.ListRoutesPaginated()                │ svc.ListRoutesPaginated(ctx, limit, offset)         │                                                             │
     │ └────────────────────┴─────────────────────────────────────────────┴─────────────────────────────────────────────────────┘                                                             │
     │                                                                                                                                                                                        │
     │ Service errors map to HTMX responses:                                                                                                                                                  │
     │ - IsValidation → render form with field errors                                                                                                                                         │
     │ - IsConflict → toast "slug already exists" or "route has content"                                                                                                                      │
     │ - IsNotFound → toast "not found"                                                                                                                                                       │
     │                                                                                                                                                                                        │
     │ API Handlers                                                                                                                                                                           │
     │                                                                                                                                                                                        │
     │ internal/router/media.go — dispatcher signatures change from config.Config to *service.Registry:                                                                                       │
     │                                                                                                                                                                                        │
     │ ┌──────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────┐                                                 │
     │ │                     Handler                      │                                   Key Change                                    │                                                 │
     │ ├──────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────┤                                                 │
     │ │ apiCreateMedia                                   │ Remove ~80 lines of S3 setup/closure construction → svc.Upload(ctx, ac, params) │                                                 │
     │ ├──────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────┤                                                 │
     │ │ apiUpdateMedia                                   │ svc.UpdateMediaMetadata(ctx, ac, params)                                        │                                                 │
     │ ├──────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────┤                                                 │
     │ │ apiDeleteMedia                                   │ Remove S3 key reconstruction + deletion → svc.DeleteMedia(ctx, ac, id)          │                                                 │
     │ ├──────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────┤                                                 │
     │ │ MediaHealthHandler                               │ Remove orphan logic → svc.MediaHealth(ctx)                                      │                                                 │
     │ ├──────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────┤                                                 │
     │ │ MediaCleanupHandler                              │ Remove orphan logic → svc.MediaCleanup(ctx)                                     │                                                 │
     │ ├──────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────────┤                                                 │
     │ │ apiListMedia, apiGetMedia, apiListMediaPaginated │ Use service read methods                                                        │                                                 │
     │ └──────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────────────────────────┘                                                 │
     │                                                                                                                                                                                        │
     │ internal/router/mediaDimensions.go — change to use service dimension methods.                                                                                                          │
     │                                                                                                                                                                                        │
     │ internal/router/routes.go — change from config.Config to *service.Registry. Gains validation that API currently lacks.                                                                 │
     │                                                                                                                                                                                        │
     │ internal/router/adminRoutes.go — change to use service admin route methods.                                                                                                            │
     │                                                                                                                                                                                        │
     │ Mux Wiring (internal/router/mux.go)                                                                                                                                                    │
     │                                                                                                                                                                                        │
     │ Update handler registrations in both NewModulacmsMux (API routes) and registerAdminRoutes (admin routes) to pass svc instead of individual driver/mgr/config dependencies to media and │
     │  route handlers.                                                                                                                                                                       │
     │                                                                                                                                                                                        │
     │ ---                                                                                                                                                                                    │
     │ Files Changed                                                                                                                                                                          │
     │                                                                                                                                                                                        │
     │ ┌────────────────────────────────────┬──────────┬────────────────────────────────────────────────────────────────┐                                                                     │
     │ │                File                │   Type   │                             Scope                              │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/service/media.go          │ Rewrite  │ Replace placeholder with full MediaService (~300 lines)        │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/service/routes.go         │ Rewrite  │ Replace placeholder with full RouteService (~200 lines)        │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/admin/handlers/media.go   │ Major    │ Change signatures, rewrite upload/update/delete to use service │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/admin/handlers/routes.go  │ Major    │ Change signatures, rewrite create/update/delete to use service │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/router/media.go           │ Major    │ Remove S3 logic + orphan helpers, replace with service calls   │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/router/mediaDimensions.go │ Moderate │ Change signatures to use service                               │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/router/routes.go          │ Moderate │ Change signatures, gains validation via service                │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/router/adminRoutes.go     │ Moderate │ Change signatures to use service                               │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/router/mux.go             │ Moderate │ Update registrations to pass svc to media/route handlers       │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/service/media_test.go     │ New      │ Service-level tests                                            │                                                                     │
     │ ├────────────────────────────────────┼──────────┼────────────────────────────────────────────────────────────────┤                                                                     │
     │ │ internal/service/routes_test.go    │ New      │ Service-level tests                                            │                                                                     │
     │ └────────────────────────────────────┴──────────┴────────────────────────────────────────────────────────────────┘                                                                     │
     │                                                                                                                                                                                        │
     │ Not changed: internal/media/*.go (reused as-is), internal/bucket/*.go (reused), mcp/ (Phase 7).                                                                                        │
     │                                                                                                                                                                                        │
     │ ---                                                                                                                                                                                    │
     │ Testing                                                                                                                                                                                │
     │                                                                                                                                                                                        │
     │ Service Tests (SQLite, no S3 needed)                                                                                                                                                   │
     │                                                                                                                                                                                        │
     │ internal/service/routes_test.go:                                                                                                                                                       │
     │ - Create with valid input succeeds                                                                                                                                                     │
     │ - Create with empty slug → ValidationError                                                                                                                                             │
     │ - Create with invalid slug format → ValidationError                                                                                                                                    │
     │ - Create with duplicate slug → ConflictError                                                                                                                                           │
     │ - Update rename to existing slug → ConflictError                                                                                                                                       │
     │ - Update same slug (no rename) passes                                                                                                                                                  │
     │ - Delete route with content → ConflictError                                                                                                                                            │
     │ - Delete route without content succeeds                                                                                                                                                │
     │                                                                                                                                                                                        │
     │ internal/service/media_test.go:                                                                                                                                                        │
     │ - Upload with no S3 config → ValidationError                                                                                                                                           │
     │ - Upload with invalid path → ValidationError                                                                                                                                           │
     │ - UpdateMetadata with focal point > 1.0 → ValidationError                                                                                                                              │
     │ - UpdateMetadata overlays only non-empty fields                                                                                                                                        │
     │ - Delete non-existent → NotFoundError                                                                                                                                                  │
     │ - Dimension create with zero width → ValidationError                                                                                                                                   │
     │                                                                                                                                                                                        │
     │ Integration Tests (MinIO required for media upload)                                                                                                                                    │
     │                                                                                                                                                                                        │
     │ - Full upload via service → verify S3 object exists + DB record correct                                                                                                                │
     │ - Admin upload path (via service) → verify S3 upload + DB record (was previously DB-only placeholder)                                                                                  │
     │ - Upload with no S3 config → verify ValidationError returned                                                                                                                           │
     │ - Delete → verify S3 objects removed + DB record gone                                                                                                                                  │
     │ - Delete with S3 failure → verify DB record still deleted (best-effort), warning logged                                                                                                │
     │ - Health → verify orphan detection                                                                                                                                                     │
     │ - Cleanup → verify orphan removal                                                                                                                                                      │
     │                                                                                                                                                                                        │
     │ Verification                                                                                                                                                                           │
     │                                                                                                                                                                                        │
     │ go build ./...              # compile check                                                                                                                                            │
     │ just test                   # unit tests pass                                                                                                                                          │
     │ just test-minio && just test-integration  # media S3 integration                                                                                                                       │
     │                                                                                                                                                                                        │
     │ ---                                                                                                                                                                                    │
     │ Implementation Order                                                                                                                                                                   │
     │                                                                                                                                                                                        │
     │ These two services are independent and can be built by parallel agents:                                                                                                                │
     │                                                                                                                                                                                        │
     │ Stream A — RouteService (simpler, ~1 session):                                                                                                                                         │
     │ 1. Implement internal/service/routes.go                                                                                                                                                │
     │ 2. Write internal/service/routes_test.go                                                                                                                                               │
     │ 3. Rewire internal/admin/handlers/routes.go                                                                                                                                            │
     │ 4. Rewire internal/router/routes.go + adminRoutes.go                                                                                                                                   │
     │ 5. Update mux registrations for route handlers                                                                                                                                         │
     │                                                                                                                                                                                        │
     │ Stream B — MediaService (more complex, ~1-2 sessions):                                                                                                                                 │
     │ 1. Implement internal/service/media.go                                                                                                                                                 │
     │ 2. Write internal/service/media_test.go                                                                                                                                                │
     │ 3. Rewire internal/admin/handlers/media.go                                                                                                                                             │
     │ 4. Rewire internal/router/media.go + mediaDimensions.go                                                                                                                                │
     │ 5. Update mux registrations for media handlers                                                                                                                                         │
     │                                                                                                                                                                                        │
     │ Final: Build + test both together to verify no conflicts in mux wiring.                                                                                                                │
     ╰────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
