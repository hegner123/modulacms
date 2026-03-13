# Media Replace-on-Delete

## Problem

When a user deletes a media asset, content fields referencing that media silently break. The existing `clean_refs=true` API parameter blanks field values, but there is no UI for it, no usage count shown, and no option to replace rather than nullify.

## Goal

When deleting media via admin panel or TUI:
1. Show how many content fields reference the asset (across both `content_fields` and `admin_content_fields`)
2. Let the user pick a replacement media asset OR confirm deletion with null
3. Update all referencing field values to either the replacement media ID or empty string
4. Then delete the original media

## Existing Infrastructure

| Component | File | What it does today |
|-----------|------|--------------------|
| Reference scan API | `internal/router/media.go:242-300` | `GET /api/v1/media/references?q=<media_id>` — scans `content_fields` only |
| Reference cleaning | `internal/router/media.go:302-350` | `cleanMediaReferences()` — blanks field values containing the media ID/URL |
| API delete | `internal/router/media.go:164-202` | `?clean_refs=true` blanks then deletes |
| Service delete | `internal/service/media.go:255-290` | Deletes S3 objects + DB record |
| Admin delete handler | `internal/admin/handlers/media.go:240-277` | HTMX DELETE — no ref scan, no replacement |
| TUI delete dialog | `internal/tui/update_dialog_delete.go:149-190` | Simple confirm dialog — no ref scan |
| TUI delete command | `internal/tui/commands.go:521-547` | Calls `d.DeleteMedia()` directly — no ref cleanup |
| Media validation | `internal/validation/type_validators.go:89-94` | Media field values are ULID MediaIDs |

### Key Constraints

- Media field values are stored as 26-char ULID strings in `field_value` (TEXT column)
- Both `content_fields` AND `admin_content_fields` can reference media
- Reference scan currently loads ALL content fields and does substring match — works but doesn't scale. Acceptable for now (greenfield, no active users)
- Block editor may embed media IDs in JSON field values (richtext/json fields)

## Implementation Plan

### Phase 1: Service Layer — Reference Scanning + Replace/Null

**Files:** `internal/service/media.go`

1. **Add `ScanMediaReferences` method** to `MediaService`
   - Scans both `content_fields` AND `admin_content_fields` for media ID and URL
   - Returns structured result with counts per table and field details
   - Moves logic from `router/media.go` handler helper into service layer

```go
type MediaReference struct {
    ContentFieldID types.ContentFieldID
    ContentDataID  types.NullableContentID
    FieldID        types.NullableFieldID
    IsAdmin        bool   // true = admin_content_fields, false = content_fields
}

type MediaReferenceScanResult struct {
    MediaID        types.MediaID
    References     []MediaReference
    ReferenceCount int
    PublicCount    int  // content_fields
    AdminCount     int  // admin_content_fields
}
```

2. **Add `ReplaceMediaReferences` method** to `MediaService`
   - Takes `mediaID` (to find), optional `replacementID` (to replace with, nil = empty string)
   - Validates replacement media exists (if provided)
   - Updates all matching field values in both tables
   - Audits each change
   - Returns count of updated fields

```go
type ReplaceMediaReferencesParams struct {
    MediaID       types.MediaID
    ReplacementID *types.MediaID // nil = set to empty string
}

type ReplaceMediaReferencesResult struct {
    UpdatedCount int
    PublicCount  int
    AdminCount   int
}
```

3. **Add `DeleteMediaWithReferences` method** to `MediaService`
   - Orchestrates: scan → replace/null → delete S3 → delete DB
   - Single entry point for admin panel and TUI

```go
type DeleteMediaWithReferencesParams struct {
    MediaID       types.MediaID
    ReplacementID *types.MediaID // nil = null out references
}

type DeleteMediaWithReferencesResult struct {
    ReferencesUpdated int
    S3ObjectsDeleted  int
}
```

### Phase 2: API Layer

**Files:** `internal/router/media.go`

1. **Update `apiDeleteMedia`** to accept optional `replacement_id` query parameter
   - `DELETE /api/v1/media/{id}?clean_refs=true&replacement_id=<ulid>` — replace then delete
   - `DELETE /api/v1/media/{id}?clean_refs=true` — null then delete (existing behavior, now via service)
   - `DELETE /api/v1/media/{id}` — delete without ref cleanup (existing behavior)
   - Call `svc.Media.DeleteMediaWithReferences()` when `clean_refs=true`

2. **Update `MediaReferencesHandler`** to also scan `admin_content_fields`
   - Return `public_count` and `admin_count` in response
   - Delegate to `svc.Media.ScanMediaReferences()` instead of inline logic

3. **Update response types** to include the new fields

### Phase 3: Admin Panel

**Files:** `internal/admin/handlers/media.go`, `internal/admin/pages/media_detail.templ`, `internal/admin/static/`

1. **Add reference count badge** to `MediaDetailHandler`
   - Call `svc.Media.ScanMediaReferences()` when rendering detail page
   - Pass count to template

2. **Replace simple delete button** with a multi-step HTMX flow:
   - Click "Delete" → HTMX GET fetches reference count
   - If 0 references: standard confirm dialog ("Delete this media?")
   - If N > 0 references: show "Used in N places" dialog with options:
     - **Replace with...** — opens media picker, then confirms replacement + delete
     - **Delete anyway** — nullifies all references then deletes
     - **Cancel**

3. **Add HTMX endpoint** for the reference-aware delete dialog:
   - `GET /admin/media/{id}/delete-check` — returns dialog partial with reference info
   - `DELETE /admin/media/{id}` — updated to accept `replacement_id` form value

4. **Wire media picker into delete dialog**
   - Reuse existing `?picker=true` media list for replacement selection
   - Selected replacement shown as preview before confirming

### Phase 4: TUI

**Files:** `internal/tui/update_dialog_delete.go`, `internal/tui/commands.go`, `internal/tui/screen_media.go`

1. **Add `MediaRefCountMsg`** and `FetchMediaRefCountCmd`**
   - Before showing delete dialog, fetch reference count asynchronously
   - Transition: user presses delete → loading → ref count returned → dialog shown

2. **Enhance `DeleteMediaContext`** with reference info:

```go
type DeleteMediaContext struct {
    MediaID       types.MediaID
    Label         string
    RefCount      int
    ReplacementID *types.MediaID  // set if user picks replacement
    Phase         int             // 0=ref scan, 1=choose action, 2=pick replacement
}
```

3. **Multi-phase delete dialog flow:**
   - **Phase 0:** Async fetch reference count (loading spinner)
   - **Phase 1:** Show dialog:
     - If 0 refs: "Delete media 'X'? This cannot be undone." [Delete] [Cancel]
     - If N refs: "Media 'X' is used in N places." [Replace & Delete] [Delete Anyway] [Cancel]
   - **Phase 2:** (if Replace chosen) Navigate to media list in picker mode, user selects replacement → returns to confirm → execute

4. **Update `HandleDeleteMedia`** to call `svc.Media.DeleteMediaWithReferences()`
   - TUI currently calls `d.DeleteMedia()` directly — needs service layer access
   - For remote mode: needs new SDK method or uses existing API with new params

### Phase 5: Go/TypeScript/Swift SDK Updates

**Files:** `sdks/go/`, `sdks/typescript/modulacms-admin-sdk/`, `sdks/swift/`

1. **Go SDK** — Update `DeleteMedia` to accept optional `ReplacementID`
2. **TypeScript Admin SDK** — Update `deleteMedia` with optional `replacementId` parameter
3. **Swift SDK** — Update `deleteMedia` with optional `replacementID` parameter
4. **All SDKs** — Update `MediaReferenceScanResponse` type with `public_count`/`admin_count`

### Phase 6: MCP Tools

**Files:** `internal/mcp/`

1. **Update `delete_media` MCP tool** to accept `replacement_id` parameter
2. **Add `scan_media_references` MCP tool** or update existing tools

## Execution Order

1. Phase 1 (service layer) — foundation, no UI changes
2. Phase 2 (API layer) — backward compatible additions
3. Phase 5 (SDKs) — update types to match new API
4. Phase 3 (admin panel) — user-facing UI
5. Phase 4 (TUI) — user-facing UI
6. Phase 6 (MCP) — tooling

Phases 3, 4, and 6 are independent of each other and can run in parallel after 1+2+5.

## Edge Cases

- **Replacement media is itself deleted** — Not a concern. The replacement is a different media asset that continues to exist.
- **Circular replacement** — Can't happen. You're replacing references to A with B, then deleting A. B isn't being deleted.
- **Media referenced in JSON/richtext fields** — The existing substring search handles this (finds media ID anywhere in field_value). Replacement should also use string replacement within the JSON value, not full field replacement.
- **Media referenced in published snapshots** — Snapshots are immutable. References in published content remain as the old media ID. This is correct — published content is a point-in-time capture.
- **Concurrent deletion** — Not a concern at this stage (greenfield, no active users).
- **Admin content fields** — Must scan both `content_fields` AND `admin_content_fields`. Current code only scans public.

## Non-Goals

- Automatic media deduplication
- Media versioning (upload new version of same asset)
- Bulk media deletion with replacement
- Reference tracking as a persistent index (materialized view)
