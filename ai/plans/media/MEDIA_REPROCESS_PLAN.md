# Media Reprocess Plan

Fixes two problems: (1) existing images don't get variants when new dimensions are added, (2) the crop algorithm makes focal points nearly useless.

## Phase 1: Fix the crop algorithm

**Problem:** `OptimizeUpload` extracts a pixel-exact crop region (`width x height` pixels from the source). A "320x240" dimension extracts a 320x240 pixel region from a 4000x3000 image -- a tiny sliver. The focal point controls which sliver, but the result is never what a user expects ("give me a 320x240 version of this image").

**Fix:** Change to aspect-ratio crop + scale (industry standard):

```
1. targetAR := float64(dim.Width.Int64) / float64(dim.Height.Int64)
2. sourceAR := float64(srcWidth) / float64(srcHeight)
3. If sourceAR > targetAR (source is wider than target):
     cropHeight = srcHeight
     cropWidth  = int(float64(srcHeight) * targetAR)
   Else (source is taller than target):
     cropWidth  = srcWidth
     cropHeight = int(float64(srcWidth) / targetAR)
4. Center crop rectangle on focal point (or image center if no focal point), clamp to bounds
5. Scale cropped region to dim.Width.Int64 x dim.Height.Int64 using draw.BiLinear
```

Note: `dim.Width` and `dim.Height` are `types.NullableInt64`. Use explicit `float64()` casts for the aspect ratio division to avoid integer truncation.

This means the focal point controls which part of the image survives the aspect ratio crop, which is its actual purpose. Every dimension preset produces a full-resolution output at the target size.

**Upscaling skip logic:** The current check (`cropWidth > srcWidth || cropHeight > srcHeight`) must change. With aspect-ratio crop + scale, the condition becomes: skip if target output width > source width OR target output height > source height. A 320x240 preset from a 400x300 source is fine (downscale). A 1920x1080 preset from a 800x600 source is skipped (would upscale).

**Files:**
- `internal/media/media_optimize.go` -- rewrite the `// Crop and scale images.` loop in `OptimizeUpload` (the `for _, dim := range *dimensions` block, currently lines 103-145). Replace the entire loop body with the aspect-ratio crop + scale algorithm, including the upscaling skip condition.
- `internal/media/media_optimize_test.go` -- update/add tests for new algorithm, including edge cases around the skip threshold

**Behavior change:** All future uploads and reprocesses will use the new algorithm. Existing variants are unaffected until reprocessed (Phase 2 handles that).

**Blast radius:** `OptimizeUpload` is called from three locations: `internal/media/media_upload.go` (public media upload pipeline), `internal/media/admin_media_service.go:203` (admin media upload pipeline), and `internal/service/media.go` (ReprocessMediaVariants). The algorithm change affects all three automatically since they all call the same function. No additional wiring needed, but verify admin media uploads still work after the change.

## Phase 2: Bulk reprocess on dimension changes

**Problem:** Creating, updating, or deleting a dimension preset has no effect on existing images. Their srcset stays frozen from upload time.

### Service layer

Add to `MediaService` in `internal/service/media.go`:

```
ReprocessAllMediaVariants(ctx context.Context, ac audited.AuditContext) error
```

- Queries all media records using `ListMedia()`, then filters in Go by checking `media.IsImageMIME(record.Mimetype.String)` for each record. A dedicated SQL query is unnecessary for v1 since media tables are typically small (thousands, not millions).
- Calls `ReprocessMediaVariants` for each record
- Runs as a background goroutine -- does not block the HTTP response
- Tracks progress via an in-memory status struct:

```go
type ReprocessStatus struct {
    Running   bool
    Total     int
    Completed int
    Failed    int
    StartedAt time.Time
}
```

- Only one reprocess job runs at a time. If a new dimension change comes in while reprocessing, it queues a restart after the current run finishes (so the new dimension is picked up).

### Concurrency design

This is the first background job in the service layer. The entire `internal/service/` package is synchronous request-scoped code today. The following decisions are fixed:

**Context ownership:** The background goroutine must NOT use the HTTP request context (it is canceled when the response is sent). `MediaService` receives a service-scoped `context.Context` at construction time (derived from the server's root context, canceled on graceful shutdown). The background goroutine uses this context. This requires adding a `context.Context` parameter to `NewMediaService` (currently `func NewMediaService(driver db.DbDriver, mgr *config.Manager)` in `internal/service/media.go:40`), which in turn requires updating `NewRegistry` (`internal/service/service.go:62`) and its call sites: `cmd/serve.go` and `internal/router/routes_test.go`.

**Shared state protection:** `ReprocessStatus` is guarded by a `sync.RWMutex` on the `MediaService` struct. HTTP handlers that read status acquire `mu.RLock()`/`mu.RUnlock()`. The background goroutine acquires `mu.Lock()`/`mu.Unlock()` when updating counters. Fields: `mu sync.RWMutex`, `reprocessStatus ReprocessStatus`.

**Restart queuing:** A single `restartQueued bool` field under the same mutex. When a dimension CRUD triggers reprocess and `Running == true`, set `restartQueued = true` and return. At the end of the background goroutine's run loop, check `restartQueued` under lock; if true, reset it and start a new run. This is a two-state flag, not a channel or state machine.

**Partial failure behavior:** Skip-and-continue. When `ReprocessMediaVariants` returns an error for a single record, log the media ID and error at `Warn` level, increment `Failed`, and continue to the next record. Do not halt the run. Do not retry. The operator can re-trigger manually after investigating.

**Graceful shutdown:** The background goroutine checks `ctx.Done()` between each record. If the context is canceled mid-run, it stops processing, sets `Running = false`, and returns. Partially reprocessed runs are safe because each `ReprocessMediaVariants` call is atomic per-record (either the record's srcset is fully updated or rolled back).

**Audit context:** The background goroutine runs outside any HTTP request, so there is no authenticated user. Construct a system-level `AuditContext` with `UserID` set to `"system"`, `IPAddress` set to `"127.0.0.1"`, and `UserAgent` set to `"media-reprocess-worker"`. Construct this once at the start of `ReprocessAllMediaVariants` and reuse it for every `ReprocessMediaVariants` call in the run.

**S3 throughput:** Process records sequentially (one at a time). No bounded concurrency in v1. This keeps the implementation simple and avoids S3 rate limit concerns. If reprocess speed becomes an issue later, add a configurable worker pool.

### Trigger points

After dimension CRUD succeeds, kick off background reprocess:

- `internal/admin/handlers/media_dimensions.go` -- MediaDimensionCreateHandler, MediaDimensionUpdateHandler, MediaDimensionDeleteHandler
- `internal/router/mediaDimensions.go` -- apiCreateMediaDimension, apiUpdateMediaDimension, apiDeleteMediaDimension

### New API endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/v1/media/reprocess/status` | Returns current reprocess job status |
| `POST` | `/api/v1/media/reprocess` | Manually trigger bulk reprocess (admin only) |

Both endpoints require `media:update` permission. Register them in `mux.go` wrapped with `RequirePermission("media:update")`.

### API response

For non-HTMX API callers (`apiCreateMediaDimension`, etc.), include `"reprocess_started": true` in the JSON response body when a background reprocess is kicked off. This tells API consumers to poll `/api/v1/media/reprocess/status` if they care about completion.

### Admin UI

- After dimension CRUD, toast: "Dimension saved. Regenerating image variants in background..."
- Status indicator on media dimensions page: an HTMX `div` with `hx-get="/admin/media/dimensions/reprocess-status"` and `hx-trigger="every 3s"`. The partial returns a progress bar (Completed/Total with percentage) when `Running == true`, and an empty `div` (no `hx-trigger` attribute) when `Running == false` to stop polling. The polling div is included in the initial page load from `media_dimensions_list.templ` and also injected via OOB swap after dimension CRUD responses.

### Files

- `internal/service/media.go` -- ReprocessAllMediaVariants, ReprocessStatus, background runner
- `internal/admin/handlers/media_dimensions.go` -- trigger after CRUD
- `internal/router/mediaDimensions.go` -- trigger after CRUD
- `internal/router/media.go` -- new status/trigger endpoints
- `internal/admin/pages/media_dimensions_list.templ` -- progress indicator
- `internal/admin/partials/media_dimensions_table_rows.templ` -- progress partial for HTMX polling

## Phase 3: Fix admin media reprocess stub

**Problem:** `internal/admin/handlers/admin_media.go:450` logs "focal point changed, reprocessing admin media variants" but never actually reprocesses. It's a stub.

**Fix:** Add a separate `ReprocessAdminMediaVariants` method to `MediaService`. Do not parameterize the existing public method. Admin media uses different DB accessors (`GetAdminMedia`, `UpdateAdminMedia`), a different bucket (`AdminBucketMedia()` with fallback to shared bucket), and different table structures. A separate method is clearer than a generic function with callbacks.

Signature: `ReprocessAdminMediaVariants(ctx context.Context, ac audited.AuditContext, adminMediaID types.AdminMediaID) error`. Uses `GetAdminMedia`, `UpdateAdminMedia`, and admin bucket config methods. Otherwise mirrors the structure of `ReprocessMediaVariants` line-by-line: download original from S3 (admin bucket), generate variants with `OptimizeUpload`, delete old variants, upload new variants, update admin media DB record.

**Files:**
- `internal/service/media.go` -- add `ReprocessAdminMediaVariants` method
- `internal/admin/handlers/admin_media.go` -- replace stub at line 450 with actual call to `ReprocessAdminMediaVariants`

## Phase 4: Focal point on upload (optional)

**Problem:** Focal point can only be set after upload (on the edit page). Initial crops always use image center. The user must upload, then edit, triggering a full reprocess round-trip.

**Fix:** Add `focal_x` and `focal_y` fields to the upload form. Pass them through the upload pipeline so the initial crop uses them.

**Threading approach:** Today, `ProcessMediaUpload` creates the DB record (without focal point) before running the pipeline. `HandleMediaUpload` then reads the focal point from the DB record (line 43 of `media_upload.go`). For Phase 4, populate the existing `FocalX` and `FocalY` fields of `CreateMediaParams` at construction time in `ProcessMediaUpload`. These fields already exist in the sqlc-generated struct but are currently left at their zero value (null). No schema or sqlc changes are needed. The pipeline's existing DB read then picks up the focal point values. This avoids changing the `HandleMediaUpload` function signature.

**Files:**
- `internal/admin/pages/media_list.templ` -- the upload form lives here (there is no separate media_upload.templ); add focal point picker to the upload section
- `internal/admin/handlers/media.go` -- parse focal fields from upload form
- `internal/service/media.go` -- thread focal point through Upload path, include focal_x/focal_y in the `CreateMediaParams` construction inside `Upload()`
- `internal/media/media_service.go` -- add focal point fields to `ProcessMediaUpload`'s params so they can be passed to the DB create call (no signature change to `HandleMediaUpload` needed)
- `internal/media/media_upload.go` -- no changes needed (reads focal point from DB, which now has values at upload time)

This is optional polish. The reprocess-on-focal-change flow already works; this eliminates one S3 round-trip for users who know the focal point at upload time.

## Execution order

1. **Phase 1** first -- fixes the algorithm that everything else depends on
2. **Phase 2** next -- enables bulk reprocess (makes Phase 1 retroactive for existing images)
3. **Phase 3** -- bug fix, small scope
4. **Phase 4** -- optional, implement only if explicitly requested. Do not implement unless the operator confirms.
