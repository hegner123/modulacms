# Plan: Publishing, Versioning, and i18n

**Status:** Phases 1-2 COMPLETE | Phases 3-6 PLANNED
**Date:** 2026-02-28

---

## Table of Contents

1. [Current State](#current-state)
2. [Design Decisions](#design-decisions)
3. [Publishing and Versioning (Unified)](#publishing-and-versioning)
4. [Internationalization (i18n)](#internationalization)
5. [Webhooks & Content Lifecycle Events (Phase 3)](#webhooks--content-lifecycle-events)
6. [Review Workflows (Phase 4)](#review-workflows)
7. [Preview Environments & Draft Sharing (Phase 5)](#preview-environments--draft-sharing)
8. [Bulk Operations & Content Dependencies (Phase 6)](#bulk-operations--content-dependencies)
9. [Implementation Phases](#implementation-phases)
10. [Schema Changes](#schema-changes)

---

## Current State

| Area | State |
|------|-------|
| Content statuses | `draft` and `published` — enforced via snapshot publishing (Phase 1 COMPLETE) |
| Public delivery | Snapshot-based — only published content is publicly accessible (Phase 1 COMPLETE) |
| Publish permission | `content:publish` gates snapshot creation and unpublish (Phase 1 COMPLETE) |
| Version history | Snapshot-based versioning with restore, retention policy, optimistic locking (Phase 1 COMPLETE) |
| Scheduled publishing | Background goroutine with configurable interval (Phase 1 COMPLETE) |
| Plugin hooks | `before_publish`/`after_publish` fire on status transitions |
| Audit trail | `change_events` records old/new JSON for every mutation |
| Locale support | Per-field locale column, shared tree, fallback chains, per-locale publishing (Phase 2 COMPLETE) |
| Content structure | Tree with sibling pointers on `content_data`; field values in separate `content_fields` table (EAV) |
| Webhooks | Not yet (Phase 3 PLANNED) |
| Review workflows | Not yet (Phase 4 PLANNED) |
| Preview sharing | Not yet (Phase 5 PLANNED) |
| Bulk operations | Not yet (Phase 6 PLANNED) |

---

## Design Decisions

These decisions are based on research into how Strapi v5, Contentful, Payload CMS, Sanity, Directus, WordPress, Wagtail, and DatoCMS implement these features, filtered through ModulaCMS's specific architecture (sibling-pointer trees, EAV field storage, tri-database, self-hosted single binary).

### Publishing: Snapshot-based two-copy

**Why not row-duplication (Strapi v5 style)?** Strapi maintains two database rows per document (draft row + published row). ModulaCMS's content is a tree with sibling pointers. Duplicating the tree — maintaining parallel sets of parent_id, first_child_id, next_sibling_id, prev_sibling_id for draft and published — would double the complexity of every tree operation across all three database backends.

**The approach:** The editing tables (`content_data` + `content_fields`) are always the working draft. Publishing freezes the current state into a JSON snapshot stored in `content_versions`. The delivery API reads exclusively from published snapshots, never from the live editing tables. This gives true draft/published separation without duplicating the tree.

**Why this works well for ModulaCMS:** The tree structure with sibling pointers is already complex. Snapshotting it as JSON avoids duplicating it entirely. The snapshot stores the raw pieces (tree nodes + field values), not a pre-composed final output. Delivery deserializes the snapshot back into the same intermediate types (`ContentDataJSON`, `DatatypeJSON`, `ContentFieldsJSON`, `FieldsJSON`) and feeds them through the existing `model.BuildTree()` → `transform.TransformAndWrite()` pipeline. The `?format=` parameter and all six output format transformers (raw, contentful, sanity, strapi, wordpress, clean) continue to work unchanged. Tree composition (`ComposeTrees` for `content_tree_ref` fields) also still happens at delivery time — referenced subtrees are resolved by looking up their published snapshots.

### Versioning: Unified with publishing

**Why not a separate versioning system?** If publishing creates a snapshot, version history is just the series of snapshots. The `content_versions` table serves both purposes — a snapshot marked `published=true` is the live version, all other snapshots are version history. No need to build and maintain two parallel systems.

### Statuses: Two only

**Why drop `pending` and `archived`?** The research shows the industry has converged on two core statuses: `draft` and `published`. WordPress's 8-status system is widely considered overcomplicated. "Scheduled" should be a timestamp field (the document is still a draft until publication actually happens). "Pending review" belongs in a separate optional workflow layer that doesn't leak into the delivery API.

`archived` can be achieved by unpublishing (removing the published snapshot) and is an editorial UI concept, not a delivery API concept. `pending` can be reintroduced later as part of an optional review workflow layer — a separate field orthogonal to status, as Strapi v5 does.

For now, content is either `draft` (not publicly accessible) or `published` (a published snapshot exists). The `status` field on `content_data` becomes an editorial indicator. The delivery API checks for the existence of a published snapshot, not the status field.

### Webhooks: Async delivery with retries

**Why now?** Webhooks are table stakes for headless CMS. Every major platform (Strapi, Contentful, Contentstack, Hygraph) ships them alongside publishing. Without webhooks, published content changes can't notify external systems — no CDN invalidation, no SSG rebuild triggers, no Slack notifications.

**Why a dedicated table, not plugin hooks?** The existing `before_publish`/`after_publish` Lua hooks fire in-process. Webhooks need async delivery with retries, audit trails, and admin-managed configuration. These are complementary — Lua hooks for in-process logic, webhooks for external notification.

**Retry strategy:** Exponential backoff with 3 retries (1min, 5min, 30min). After final failure, mark as failed — no infinite retry loops. Admin can manually retry from the UI.

**Security:** HMAC-SHA256 signature in `X-ModulaCMS-Signature` header using a per-webhook secret. Payload is JSON. HTTPS-only enforced on save (with dev bypass config flag).

### Review workflows: Single approval gate

**Why not multi-stage (Strapi style)?** Strapi's 4-stage workflow (To do → In progress → Ready to review → Reviewed) is overcomplicated for most teams. Every CMS that shipped complex workflows before knowing if anyone would use them regretted it. Start with the simplest useful model.

**The approach:** A single optional approval gate between draft and published. Datatypes can opt in via a `requires_review` flag. When enabled, content must pass through `pending_review` status before it can be published. Users with `content:review` permission can approve or reject.

**Why per-datatype, not global?** Blog posts may need review; footer links may not. Making it per-datatype lets teams apply review only where it matters.

**Why not a separate workflow engine?** A dedicated workflow table with stages, transitions, and assignees adds significant complexity. The single-gate model uses the existing status field and permission system. If multi-stage is needed later, it can be built on top of this foundation.

### Preview environments: Token-based draft sharing

**Why not OAuth/session-based?** The whole point is sharing with people who don't have CMS accounts — stakeholders, clients, external reviewers. A simple token in the URL is the right UX.

**Why database-backed tokens, not JWTs?** Tokens need revocation (stakeholder leaves project, content is published and preview no longer needed). Database lookup is simple and revocation is immediate. JWTs would require a denylist, which is database-backed anyway.

**Scope:** Tokens are scoped to a specific content item + locale. A token for the English version doesn't grant access to the Spanish draft. Generate separate tokens per locale if needed.

### Bulk operations: Dependency tracking via query, not table

**Why not a dedicated dependencies table?** Content references are already stored in `content_fields` as relation and `content_tree_ref` field values. Querying these fields to find "what references content X" is a read operation, not a new data structure. A junction table would duplicate data and require sync on every field update.

**The approach:** A query-based dependency resolver that scans relation/content_tree_ref fields. On unpublish, query dependents and warn. On bulk publish, use the same resolver to find and include referenced content.

### i18n: Locale column on content_fields, shared tree

**Why not row-per-locale on content_data (Strapi style)?** It would duplicate the entire sibling-pointer tree per locale. 5 locales = 5 copies of every tree with 5x the pointer-maintenance complexity. Reordering blocks on one locale would require N separate tree operations.

**Why not JSON locale maps in field_value (Contentful style)?** Breaks the existing string field_value contract, every handler, and all three SDKs.

**The approach:** Add a `locale` column to `content_fields`. The tree structure (`content_data`) is shared across all locales — structure is language-independent. Only field values differ per locale. Translatable fields have N rows (one per locale). Non-translatable fields have one row with `locale=''`. This is closest to the Directus junction-table model, adapted to the existing EAV structure.

**Why this is the best fit:**
- Tree structure stays untouched (no duplication, no per-locale pointer maintenance)
- Non-translatable fields have one row — zero duplication, no sync mechanism needed
- Translatable fields are simply additional content_field rows filtered by locale
- Existing queries add one WHERE clause (`AND locale = ?`)
- Per-locale publishing works through the snapshot system (separate snapshots per locale)
- Per-locale version history comes free (each locale's snapshots are independent)

**Tradeoff accepted:** All locales share the same page structure. You cannot have a different block layout per locale. This matches the Wagtail model (synchronized structure, translated values) and is correct for the vast majority of CMS use cases.

### Tree-level publishing

**The root content entry is the unit of publishing.** Publishing snapshots the entire tree (root node + all descendants + all field values for the target locale) as one JSON blob. No per-node publishing. This prevents "half-published tree" problems where a page loads but sections are missing because individual blocks are still in draft.

### Fallback chains

**Capped at 2 hops.** `fr-CA → fr → en` is standard and sufficient. Unlimited chains are harder to debug and can silently serve wrong-language content through long chains. The API response includes `fallback_used: true` and `resolved_locale` so the frontend can decide whether to show fallback content or a "not available" message.

### Admin table parity

ModulaCMS has a complete parallel set of admin tables (`admin_content_data`, `admin_content_fields`, `admin_routes`, `admin_datatypes`, `admin_fields`, etc.) that are structurally identical to their public counterparts. These share the same sibling-pointer tree structure and the same EAV field storage pattern. The admin tree pipeline already uses `MapAdminContentDataJSON` to adapt admin types into the common `ContentDataJSON` shape consumed by `model.BuildTree()` and the transform layer.

**All three features (publishing, versioning, i18n) apply to admin tables in parallel.** Every schema change, new table, and new column described in this plan has an admin counterpart:

| Public | Admin |
|--------|-------|
| `content_versions` | `admin_content_versions` |
| `content_data.published_at` | `admin_content_data.published_at` |
| `content_data.revision` | `admin_content_data.revision` |
| `content_fields.locale` | `admin_content_fields.locale` |
| `fields.translatable` | `admin_fields.translatable` |

The DbDriver interface additions, sqlc queries, and wrapper implementations are duplicated per the existing pattern (public + admin). The publishing/versioning/i18n logic is shared — the same `BuildSnapshot`, `RestoreVersion`, and locale resolution code operates on both table families via the `ContentDataJSON` / `FieldsJSON` adapter layer.

---

## Publishing and Versioning

These are a single system. Publishing creates a versioned snapshot. Version history is the series of snapshots. The published version is a specific snapshot.

### 3.1 content_versions Table

Schema number: `27_content_versions`

```sql
CREATE TABLE IF NOT EXISTS content_versions (
    version_id       TEXT PRIMARY KEY CHECK (length(version_id) = 26),
    content_data_id  TEXT NOT NULL REFERENCES content_data(content_data_id) ON DELETE CASCADE,
    version_number   INTEGER NOT NULL,
    locale           TEXT NOT NULL DEFAULT '',
    snapshot         TEXT NOT NULL,             -- JSON: full tree + field values (MySQL: use MEDIUMTEXT for >64KB snapshots)
    published        INTEGER NOT NULL DEFAULT 0,-- 1 = this is the currently live version
    trigger          TEXT NOT NULL,             -- 'publish' | 'manual' | 'autosave' | 'pre_restore'
    label            TEXT,                      -- optional user label ("launch copy v2")
    created_by       TEXT NOT NULL REFERENCES users(user_id),
    created_at       TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (content_data_id, locale, version_number)
);

CREATE INDEX idx_cv_content_locale ON content_versions(content_data_id, locale);
CREATE INDEX idx_cv_published ON content_versions(content_data_id, locale, published)
    WHERE published = 1;
```

The `snapshot` JSON contains the complete tree state for delivery:

```json
{
  "tree": [
    {
      "content_data_id": "...",
      "parent_id": "...",
      "first_child_id": "...",
      "next_sibling_id": "...",
      "prev_sibling_id": "...",
      "datatype_id": "...",
      "datatype_label": "..."
    }
  ],
  "fields": [
    {
      "content_data_id": "...",
      "field_id": "...",
      "field_label": "...",
      "field_type": "...",
      "field_value": "..."
    }
  ],
  "route": {
    "route_id": "...",
    "slug": "...",
    "title": "..."
  },
  "schema_version": 1
}
```

The snapshot stores the same data shapes that `model.BuildTree()` consumes today. At delivery time, the snapshot is deserialized back into `[]ContentDataJSON`, `[]DatatypeJSON`, `[]ContentFieldsJSON`, and `[]FieldsJSON`, then fed through `BuildTree()` → `TransformAndWrite()`. This means:
- All six output format transformers (`?format=raw|contentful|sanity|strapi|wordpress|clean`) continue to work unchanged
- Tree composition (`ComposeTrees`) resolves `content_tree_ref` references by looking up other published snapshots at delivery time — composed subtrees are not baked into the snapshot
- The denormalized labels (datatype_label, field_label, etc.) avoid post-deserialization lookups against the datatypes/fields tables
- `schema_version` enables forward-compatible snapshot format evolution

### 3.2 New Columns on content_data (and admin_content_data)

```sql
published_at   TEXT,              -- timestamp of most recent publish (any locale)
published_by   TEXT REFERENCES users(user_id),
publish_at     TEXT,              -- scheduled future publish time (NULL = not scheduled)
revision       INTEGER NOT NULL DEFAULT 0  -- monotonic counter for optimistic locking
```

These are editorial metadata — the actual published state is determined by whether a published snapshot exists in content_versions. The `revision` column is incremented on every write and used for conflict detection (see 3.11).

### 3.3 Status Simplification

Reduce `ContentStatus` to two values:

```go
ContentStatusDraft     ContentStatus = "draft"
ContentStatusPublished ContentStatus = "published"
```

Remove `archived` and `pending` from the enum. Existing content with those statuses migrates to `draft`.

The `status` field on `content_data` indicates:
- `draft`: working copy, may or may not have a published snapshot
- `published`: has been published at least once (a published snapshot exists)

When content is edited after publishing, `status` stays `published` — the working copy differs from the published snapshot. The admin UI can derive a "modified" indicator by comparing the working copy's `date_modified` against the published snapshot's `created_at`. No third status needed.

### 3.4 New Permissions

| Permission | admin | editor | viewer |
|------------|-------|--------|--------|
| `content:publish` | yes | no | no |

Editors can create and edit content (draft operations). Only users with `content:publish` can create published snapshots. Unpublishing (removing the published flag) also requires `content:publish`.

### 3.5 Publish Flow

Publishing root content_data_id X for locale L:

1. Validate: user has `content:publish` permission
2. Load tree: `GetContentDataDescendants(X)` — root + all child nodes
3. Load fields: for each node, `ListContentFieldsByContentData` filtered to `locale = L` plus `locale = ''` (non-translatable)
4. Determine next version number: `SELECT COALESCE(MAX(version_number), 0) + 1 FROM content_versions WHERE content_data_id = ? AND locale = ?`
5. Build snapshot JSON (tree nodes + fields + route metadata + denormalized labels) — **outside any transaction** (read-only assembly)
6. **Begin transaction:**
   - Clear previous published flag: `UPDATE content_versions SET published = 0 WHERE content_data_id = ? AND locale = ? AND published = 1`
   - Insert new snapshot: `INSERT INTO content_versions (... published = 1, trigger = 'publish' ...)`
   - Update content_data: set `published_at`, `published_by`, `status = 'published'`, increment `revision`
   - Record in `change_events` with `action = 'publish'`
7. **Commit transaction**
8. Fire `after_publish` plugin hooks (async)
9. Prune old versions if count exceeds retention cap (async)

Snapshot construction (step 5) is separated from the write transaction (step 6) to minimize lock duration. On SQLite, the write transaction holds a database-level lock — keeping it short prevents blocking other editors. The snapshot JSON can be megabytes for large trees; serializing it inside the transaction would hold the lock unnecessarily.

### 3.6 Unpublish Flow

1. Validate: user has `content:publish` permission
2. Clear published flag: `UPDATE content_versions SET published = 0 WHERE content_data_id = ? AND locale = ? AND published = 1`
3. Update content_data: set `status = 'draft'`
4. Record in `change_events`
5. Fire hooks

The snapshot is not deleted — it becomes part of version history. The content is just no longer served by the delivery API.

### 3.7 Delivery API

`GET /api/v1/content/{slug}` changes fundamentally:

**Current flow:** slug → route_id → query content_data + content_fields with joins → build tree → transform → serve

**New flow:** slug → route_id → find published snapshot (`SELECT snapshot FROM content_versions WHERE content_data_id = (SELECT content_data_id FROM content_data WHERE route_id = ?) AND locale = ? AND published = 1`) → deserialize JSON into intermediate types → `model.BuildTree()` → optionally `ComposeTrees()` (resolving `content_tree_ref` via other published snapshots) → `transform.TransformAndWrite()` with `?format=` → serve

One query replaces the multi-table join. The existing transform pipeline and all six output formats are unchanged — the snapshot is an input cache for `BuildTree()`, not a replacement for the transform layer.

**Preview mode:** Authenticated requests with `content:read` permission and `?preview=true` use the old flow — read directly from the editing tables. This shows the current working draft. Preview responses include `X-Robots-Tag: noindex` header.

**When no published snapshot exists:** Return 404 (not a fallback to draft content).

### 3.8 Scheduled Publishing

A background goroutine started in `cmd/serve.go`:

```go
func startPublishScheduler(ctx context.Context, driver db.DbDriver, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    // Catch-up pass on startup
    publishDueContent(ctx, driver)

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            publishDueContent(ctx, driver)
        }
    }
}
```

`publishDueContent` queries for `publish_at <= now() AND status = 'draft'`, runs the publish flow for each match. Same code path as manual publishing — hooks, audit events, snapshot creation all fire.

Config:

```go
Publish_Schedule_Interval int `json:"publish_schedule_interval"` // seconds, default 60
```

### 3.9 Restore Flow

Restoring version N for content_data_id X, locale L:

1. Snapshot current state first (trigger = `pre_restore`, label = "Auto-saved before restoring version N") — safety net
2. Read version N's snapshot JSON
3. Parse field values from snapshot
4. For each field in snapshot:
   - If field_id still exists in current datatype schema: update or create the content_field row
   - If field_id no longer exists: skip, but include in the response as "unmapped fields" so the UI can show them
5. For fields in current content not in snapshot: leave as-is (don't delete — the editor can manually clean up)
6. Set `status = 'draft'` (restored content always returns to draft — requires explicit re-publish)
7. Record in `change_events`
8. Fire `after_update` plugin hooks

**Restore is field-values-only.** Tree structure (sibling pointers, parent/child relationships) is never restored. The sibling-pointer doubly-linked list is too fragile to reconstruct from a snapshot — nodes may have been added, deleted, or reordered since the snapshot was taken, and restoring partial pointer chains would create inconsistent trees. If the tree structure has changed since the snapshot, the editor keeps the current structure and gets the old field values applied to it.

The restore response includes:

```json
{
  "restored_version": 3,
  "fields_restored": 12,
  "unmapped_fields": [
    {"field_id": "...", "field_label": "subtitle", "value": "Old subtitle text"}
  ],
  "status": "draft"
}
```

### 3.10 Retention Policy

```go
Version_Max_Per_Content int `json:"version_max_per_content"` // 0 = unlimited, default 50
```

On every new version creation, if the count for that (content_data_id, locale) exceeds the cap, delete the oldest versions that are:
- Not the currently published snapshot
- Not labeled by a user

A periodic cleanup goroutine (piggybacking on the scheduled publish ticker) handles any edge cases the per-save pruning misses.

### 3.11 Optimistic Locking

New column on `content_data` (and `admin_content_data`):

```sql
revision INTEGER NOT NULL DEFAULT 0
```

Every update to content_data includes the expected `revision` value. The UPDATE statement uses `WHERE content_data_id = ? AND revision = ?`. If no rows are affected (someone else incremented the revision), return HTTP 409 Conflict. The editor must re-fetch and retry. The revision is incremented on every successful write.

`date_modified` remains as-is for display and querying purposes. The `revision` counter is the authoritative conflict detection mechanism — it is a monotonic integer with no granularity issues (unlike `date_modified` which has second resolution and cannot distinguish two saves within the same second).

This prevents the race condition where Editor A and Editor B both start from the same state, and B's save silently overwrites A's changes. Simpler than document locking, much simpler than CRDTs.

### 3.12 API Endpoints

**Publishing actions:**

```
POST /api/v1/contentdata/{id}/publish          content:publish
POST /api/v1/contentdata/{id}/unpublish        content:publish
POST /api/v1/contentdata/{id}/schedule         content:publish  (body: {"publish_at": "..."})
```

**Version management:**

```
GET    /api/v1/contentdata/{id}/versions                content:read
GET    /api/v1/contentdata/{id}/versions/{vid}          content:read
POST   /api/v1/contentdata/{id}/versions                content:update  (manual snapshot)
POST   /api/v1/contentdata/{id}/versions/{vid}/restore  content:update
DELETE /api/v1/contentdata/{id}/versions/{vid}           content:delete
```

**Delivery (public):**

```
GET /api/v1/content/{slug}                     no auth (published snapshot only)
GET /api/v1/content/{slug}?preview=true        content:read (live editing tables)
```

### 3.13 Admin UI Changes

**Content list page:**
- Status badge: gray=draft, green=published, blue dot=modified (published but working copy differs)
- Filter: All / Draft / Published / Modified / Scheduled
- Scheduled indicator with date/time

**Content edit page:**
- Action buttons based on state + permissions:
  - Draft, never published → "Publish" (if permitted)
  - Draft, has published snapshot → "Publish Changes" / "Discard Changes" (revert to published snapshot)
  - Published, unmodified → "Unpublish"
  - Any state → "Schedule" (date/time picker)
- Version history panel (collapsible sidebar):
  - List: `#N — trigger — timestamp — author — [label]`
  - Currently published version highlighted
  - "Restore" button → confirmation → restore flow
  - "Save Version" button → manual snapshot with optional label
  - "Compare" → side-by-side field value comparison between two versions
- Show `published_at`, `published_by` metadata

### 3.14 TUI Changes

- Content list: status column with colored indicators
- Content edit: publish/unpublish actions in action menu
- Version list subcommand per content item
- Restore action with confirmation

### 3.15 SDK Changes

All three SDKs (Go, TypeScript, Swift):

Types:
- Add `published_at`, `published_by`, `publish_at` to `ContentData`
- Add `ContentVersion` type (version_id, version_number, locale, trigger, label, created_by, created_at, published)
- Add `ContentVersionID` branded type

Methods:
- `Publish(contentID)`, `Unpublish(contentID)`, `Schedule(contentID, publishAt)`
- `ListVersions(contentID)`, `GetVersion(contentID, versionID)`, `CreateVersion(contentID)`, `RestoreVersion(contentID, versionID)`, `DeleteVersion(contentID, versionID)`

### 3.16 Typed IDs

Add to `internal/db/types/`:
- `ContentVersionID` — ULID, same pattern as existing typed IDs

---

## Internationalization

Builds on the publishing/versioning system. Each locale's field values live in the same `content_fields` table with a `locale` column. Per-locale publishing creates separate snapshots per locale.

### 4.1 Locale Column on content_fields

```sql
ALTER TABLE content_fields ADD COLUMN locale TEXT NOT NULL DEFAULT '';
```

Existing rows get `locale = ''` (empty string = unlocalized / default). Existing queries continue to work since `WHERE locale = ''` matches the default.

New unique constraint:

```sql
CREATE UNIQUE INDEX idx_cf_unique_locale
    ON content_fields(content_data_id, field_id, locale);
```

This prevents duplicate field values for the same (node, field, locale) combination.

### 4.2 Translatable Flag on fields

```sql
ALTER TABLE fields ADD COLUMN translatable INTEGER NOT NULL DEFAULT 0;
```

When `translatable = 0` (default): one content_field row with `locale = ''`. The value is shared across all locales.

When `translatable = 1`: one content_field row per enabled locale. Each locale gets its own value.

Default translatability by field type:
- Translatable: `text`, `textarea`, `richtext`, `slug`
- Not translatable: `number`, `boolean`, `date`, `datetime`, `media`, `json`, `select`, `email`, `url`, `relation`, `content_tree_ref`

These are defaults — the user can override per field. A price field stays the same across locales. A title field gets translated.

### 4.3 Locales Table

Schema number: `28_locales`

```sql
CREATE TABLE IF NOT EXISTS locales (
    locale_id     TEXT PRIMARY KEY CHECK (length(locale_id) = 26),
    code          TEXT NOT NULL UNIQUE,         -- BCP 47: 'en', 'es', 'fr-CA'
    label         TEXT NOT NULL,                -- "English", "Español", "Français (Canada)"
    is_default    INTEGER NOT NULL DEFAULT 0,   -- exactly one row has is_default = 1
    is_enabled    INTEGER NOT NULL DEFAULT 1,
    fallback_code TEXT,                         -- locale code to fall back to (max 2 hops)
    sort_order    INTEGER NOT NULL DEFAULT 0,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP
);
```

Constraints:
- Exactly one default locale (`is_default = 1`)
- `code` uses BCP 47 with hyphens (`en-US`, not `en_US`), normalized on input
- `fallback_code` references another locale's code. Chains capped at 2 hops and validated on save to prevent cycles
- The default locale cannot be deleted (must reassign default first)

### 4.4 Config Changes

```go
I18n_Enabled        bool   `json:"i18n_enabled"`         // default false
I18n_Default_Locale string `json:"i18n_default_locale"`  // default "en"
```

When `i18n_enabled = false`:
- No locale UI in admin/TUI
- No locale filtering on queries
- No locale resolution in delivery
- content_fields.locale stays `''` for everything
- The system behaves identically to today

When `i18n_enabled = true`:
- Locale management UI appears in admin settings
- Content edit pages show locale tabs
- Delivery API supports `?locale=` parameter
- Publishing creates locale-specific snapshots

### 4.5 Translation Creation Workflow

1. User opens content edit page, clicks "Add Translation" → locale picker
2. System creates new content_field rows for the target locale:
   - For each translatable field: copy the default locale's value as a starting point (editor translates from there)
   - Non-translatable fields (`locale = ''`) are shared — no new rows needed
3. New content_field rows have `locale = 'es'` (or whatever was selected)
4. The translation starts in draft state (no published snapshot for this locale yet)
5. Editor modifies field values in the target language
6. Editor publishes the locale independently — creates a snapshot for that locale

**No new content_data rows are created.** The tree structure is shared. Only field values (content_fields) differ per locale.

### 4.6 Delivery API with Locale Resolution

`GET /api/v1/content/{slug}` resolution order:

1. `?locale=es` query param — highest priority
2. `Accept-Language` header — parsed, matched against enabled locales
3. Default locale — final fallback

Query:

```sql
SELECT snapshot FROM content_versions
WHERE content_data_id = ?
  AND locale = ?
  AND published = 1
LIMIT 1
```

If no published snapshot for requested locale, walk the fallback chain (max 2 hops). If still nothing, try the default locale. If still nothing, 404.

Response includes locale metadata:

```json
{
  "locale": "es",
  "resolved_locale": "es",
  "fallback_used": false,
  "available_locales": ["en", "es", "fr"],
  "content": { ... }
}
```

When fallback was used:

```json
{
  "locale": "fr-CA",
  "resolved_locale": "fr",
  "fallback_used": true,
  "available_locales": ["en", "fr"],
  "content": { ... }
}
```

The `locale=*` parameter returns all available locales in one response (for SSG/ISR):

```json
{
  "locales": {
    "en": { ... },
    "es": { ... }
  },
  "available_locales": ["en", "es"]
}
```

### 4.7 Non-Translatable Field Handling

Because non-translatable fields have `locale = ''`, they are automatically included in every locale's published snapshot (the publish flow collects `locale = L` fields + `locale = ''` fields). One source of truth, no sync mechanism, no duplication.

If a non-translatable field is updated, all future publishes for any locale will pick up the new value. Existing published snapshots retain the old value (they're immutable). Re-publishing a locale creates a fresh snapshot with the current non-translatable values.

### 4.8 Content Relations Across Locales

When content references other content (via `content_tree_ref` or `relation` field types), the relation points to a content_data_id. Since the tree is shared across locales, the reference is locale-independent. The referenced content's field values resolve to the requested locale when the frontend fetches them.

### 4.9 DbDriver Interface Additions

```go
// Locales
CreateLocale(context.Context, audited.AuditContext, CreateLocaleParams) (*Locale, error)
GetLocale(types.LocaleID) (*Locale, error)
GetLocaleByCode(string) (*Locale, error)
GetDefaultLocale() (*Locale, error)
ListLocales() (*[]Locale, error)
ListEnabledLocales() (*[]Locale, error)
UpdateLocale(context.Context, audited.AuditContext, UpdateLocaleParams) error
DeleteLocale(context.Context, audited.AuditContext, types.LocaleID) error
CreateLocaleTable() error

// Locale-aware field queries
ListContentFieldsByContentDataAndLocale(types.ContentID, string) (*[]ContentFields, error)
ListContentFieldsWithFieldByContentDataAndLocale(types.ContentID, string) (*[]ContentFieldWithFieldRow, error)
```

### 4.10 Admin UI Changes

**Content edit page (when i18n enabled):**
- Locale tab bar: `[EN] [ES] [FR] [+ Add Translation]`
- Switching tabs reloads field values for the selected locale (tree stays the same)
- Non-translatable fields show once, not per locale (with a lock icon indicating "shared across languages")
- Translation completeness: "8/12 fields translated" per locale tab
- Publish button is per-locale: "Publish (English)" / "Publish (Spanish)"

**Content list page (when i18n enabled):**
- Locale badges showing which locales have published snapshots
- Filter by locale

**Settings page:**
- Locale management: add/remove/reorder locales, set default, configure fallback chains
- Per-field translatable toggle in datatype/field management

### 4.11 API Endpoints

```
GET    /api/v1/locales                  no auth (list enabled locales)
POST   /api/v1/locales                  locale:create
GET    /api/v1/locales/{id}             locale:read
PUT    /api/v1/locales/{id}             locale:update
DELETE /api/v1/locales/{id}             locale:delete
```

Publishing endpoints gain a locale parameter:

```
POST /api/v1/contentdata/{id}/publish?locale=es     content:publish
POST /api/v1/contentdata/{id}/unpublish?locale=es   content:publish
```

### 4.12 Typed IDs

Add to `internal/db/types/`:
- `LocaleID` — ULID for the locales table

No `LocaleGroupID` needed — the tree is shared, locale is a column on content_fields, not a grouping mechanism on content_data.

### 4.13 SDK Changes

All three SDKs:

Types:
- Add `locale` field to `ContentField`
- Add `translatable` field to `Field`
- Add `Locale` type (locale_id, code, label, is_default, is_enabled, fallback_code)
- Add `LocaleID` branded type
- Delivery response types gain `locale`, `resolved_locale`, `fallback_used`, `available_locales`

Methods:
- Locale CRUD: `ListLocales()`, `CreateLocale()`, `UpdateLocale()`, `DeleteLocale()`
- `?locale=` parameter on content delivery methods
- `?locale=` parameter on publish/unpublish

---

## Webhooks & Content Lifecycle Events

Phase 3. Depends on Phase 1 (publishing). Other phases fire webhook events through this system.

### 5.1 webhooks Table

Schema number: `34_webhooks`

```sql
CREATE TABLE IF NOT EXISTS webhooks (
    webhook_id    TEXT PRIMARY KEY CHECK (length(webhook_id) = 26),
    name          TEXT NOT NULL,
    url           TEXT NOT NULL,
    secret        TEXT NOT NULL,                    -- HMAC-SHA256 signing key
    events        TEXT NOT NULL DEFAULT '[]',       -- JSON array: ["content.published", "content.unpublished"]
    is_active     INTEGER NOT NULL DEFAULT 1,
    headers       TEXT NOT NULL DEFAULT '{}',       -- JSON object: custom headers to include
    author_id     TEXT NOT NULL REFERENCES users(user_id),
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

No admin counterpart — webhooks are system-wide, not per-table-family.

### 5.2 webhook_deliveries Table

Schema number: `35_webhook_deliveries`

```sql
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    delivery_id   TEXT PRIMARY KEY CHECK (length(delivery_id) = 26),
    webhook_id    TEXT NOT NULL REFERENCES webhooks(webhook_id) ON DELETE CASCADE,
    event         TEXT NOT NULL,                    -- "content.published"
    payload       TEXT NOT NULL,                    -- JSON request body sent
    status        TEXT NOT NULL DEFAULT 'pending',  -- pending | success | failed | retrying
    attempts      INTEGER NOT NULL DEFAULT 0,
    last_status_code INTEGER,                       -- HTTP response code from target
    last_error    TEXT,                             -- error message if failed
    next_retry_at TEXT,                             -- NULL if not retrying
    created_at    TEXT DEFAULT CURRENT_TIMESTAMP,
    completed_at  TEXT
);

CREATE INDEX idx_wd_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX idx_wd_retry ON webhook_deliveries(status, next_retry_at)
    WHERE status = 'retrying';
```

### 5.3 Events

| Event | Fires when | Payload includes |
|-------|-----------|-----------------|
| `content.published` | Snapshot created and marked published | content_data_id, route slug, locale, version_number |
| `content.unpublished` | Published flag cleared | content_data_id, route slug, locale |
| `content.updated` | Draft content_data modified | content_data_id, route slug, changed fields list |
| `content.scheduled` | publish_at timestamp set | content_data_id, route slug, locale, publish_at |
| `content.deleted` | Content data deleted | content_data_id, route slug |
| `locale.published` | Locale-specific snapshot published | content_data_id, locale, version_number |
| `version.created` | Manual or auto version snapshot | content_data_id, locale, version_number, trigger |

Admin-side events mirror with `admin.` prefix (e.g., `admin.content.published`).

### 5.4 Webhook Delivery Engine

New package: `internal/webhooks/`

- `Dispatch(event string, payload any)` — called from publish/unpublish/update handlers
- Queries active webhooks matching the event
- Inserts delivery rows with `status = 'pending'`
- Sends to an async worker (buffered channel, configurable pool size)
- Worker signs payload with HMAC-SHA256, sends HTTP POST, records result
- On failure: schedules retry with exponential backoff
- Background goroutine in `cmd/serve.go` processes retry queue (piggybacks on publish scheduler ticker)

### 5.5 Config

```go
Webhook_Enabled      bool `json:"webhook_enabled"`       // default false
Webhook_Timeout      int  `json:"webhook_timeout"`       // seconds, default 10
Webhook_Max_Retries  int  `json:"webhook_max_retries"`   // default 3
Webhook_Workers      int  `json:"webhook_workers"`       // concurrent delivery workers, default 4
Webhook_Allow_HTTP   bool `json:"webhook_allow_http"`    // default false (dev only)
```

### 5.6 New Permissions

| Permission | admin | editor | viewer |
|------------|-------|--------|--------|
| `webhook:create` | yes | no | no |
| `webhook:read` | yes | no | no |
| `webhook:update` | yes | no | no |
| `webhook:delete` | yes | no | no |

### 5.7 API Endpoints

```
GET    /api/v1/admin/webhooks                    webhook:read
POST   /api/v1/admin/webhooks                    webhook:create
GET    /api/v1/admin/webhooks/{id}               webhook:read
PUT    /api/v1/admin/webhooks/{id}               webhook:update
DELETE /api/v1/admin/webhooks/{id}               webhook:delete
POST   /api/v1/admin/webhooks/{id}/test          webhook:update   (send test event)
GET    /api/v1/admin/webhooks/{id}/deliveries    webhook:read     (delivery log)
POST   /api/v1/admin/webhooks/deliveries/{id}/retry  webhook:update  (manual retry)
```

### 5.8 DbDriver Interface Additions

```go
// Webhooks
CreateWebhook(context.Context, audited.AuditContext, CreateWebhookParams) (*Webhook, error)
GetWebhook(types.WebhookID) (*Webhook, error)
ListWebhooks() (*[]Webhook, error)
ListActiveWebhooksByEvent(string) (*[]Webhook, error)
UpdateWebhook(context.Context, audited.AuditContext, UpdateWebhookParams) (*string, error)
DeleteWebhook(context.Context, audited.AuditContext, types.WebhookID) error
CreateWebhookTable() error

// Webhook Deliveries
CreateWebhookDelivery(context.Context, CreateWebhookDeliveryParams) (*WebhookDelivery, error)
UpdateWebhookDeliveryStatus(context.Context, UpdateWebhookDeliveryStatusParams) error
ListWebhookDeliveries(types.WebhookID) (*[]WebhookDelivery, error)
ListPendingRetries(types.Timestamp) (*[]WebhookDelivery, error)
CreateWebhookDeliveryTable() error
```

### 5.9 Typed IDs

- `WebhookID` — ULID
- `WebhookDeliveryID` — ULID

### 5.10 Admin UI

**Settings > Webhooks page:**
- List of configured webhooks with active/inactive toggle
- Create/edit form: name, URL, secret (auto-generated), event checkboxes, custom headers
- "Test" button sends a test event and shows response
- Delivery log per webhook: timestamp, event, status, response code, retry count
- "Retry" button on failed deliveries

### 5.11 TUI

- Webhook list/create/edit in settings menu
- Delivery log view per webhook

### 5.12 SDK Changes

All three SDKs:
- `Webhook` type, `WebhookDelivery` type, `WebhookID`/`WebhookDeliveryID` branded types
- Admin SDK: `ListWebhooks()`, `CreateWebhook()`, `UpdateWebhook()`, `DeleteWebhook()`, `TestWebhook()`, `ListDeliveries()`, `RetryDelivery()`

---

## Review Workflows

Phase 4. Builds on Phase 1 (publishing statuses) and optionally Phase 3 (webhook notifications for review events).

### 6.1 New Column on datatypes/admin_datatypes

```sql
ALTER TABLE datatypes ADD COLUMN requires_review INTEGER NOT NULL DEFAULT 0;
ALTER TABLE admin_datatypes ADD COLUMN requires_review INTEGER NOT NULL DEFAULT 0;
```

### 6.2 New Status Value

Extend `ContentStatus` to three values:

```go
ContentStatusDraft         ContentStatus = "draft"
ContentStatusPendingReview ContentStatus = "pending_review"
ContentStatusPublished     ContentStatus = "published"
```

### 6.3 content_reviews / admin_content_reviews Tables

Schema number: `36_content_reviews`

```sql
CREATE TABLE IF NOT EXISTS content_reviews (
    review_id       TEXT PRIMARY KEY CHECK (length(review_id) = 26),
    content_data_id TEXT NOT NULL REFERENCES content_data(content_data_id) ON DELETE CASCADE,
    locale          TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending | approved | rejected
    submitted_by    TEXT NOT NULL REFERENCES users(user_id),
    reviewed_by     TEXT REFERENCES users(user_id),
    comment         TEXT,                             -- reviewer's comment on approve/reject
    date_submitted  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_reviewed   TEXT,

    UNIQUE (content_data_id, locale, status)          -- one pending review per locale
);

CREATE INDEX idx_cr_pending ON content_reviews(status) WHERE status = 'pending';
```

Admin counterpart: `admin_content_reviews` with identical structure referencing `admin_content_data`.

### 6.4 Review Flow

**Submit for review:**
1. Editor clicks "Submit for Review" (requires `content:update`)
2. Validate: datatype has `requires_review = 1`
3. Set `status = 'pending_review'` on content_data
4. Insert `content_reviews` row with `status = 'pending'`
5. Fire webhook: `content.review_submitted`
6. Record in `change_events`

**Approve:**
1. Reviewer clicks "Approve" (requires `content:review`)
2. Update review row: `status = 'approved'`, set `reviewed_by`, `comment`, `date_reviewed`
3. Content is now eligible for publishing (status stays `pending_review` until published)
4. Fire webhook: `content.review_approved`
5. If reviewer also has `content:publish`, offer "Approve & Publish" combo action

**Reject:**
1. Reviewer clicks "Reject" with comment (requires `content:review`)
2. Update review row: `status = 'rejected'`
3. Set content_data `status = 'draft'`
4. Fire webhook: `content.review_rejected`
5. Editor sees rejection with comment, can edit and resubmit

**Bypass:** Users with `content:publish` can always publish directly, even if `requires_review = 1`. The review gate is for editors without publish permission.

### 6.5 New Permissions

| Permission | admin | editor | viewer |
|------------|-------|--------|--------|
| `content:review` | yes | no | no |

### 6.6 API Endpoints

```
POST /api/v1/contentdata/{id}/submit-review      content:update
POST /api/v1/contentdata/{id}/approve             content:review
POST /api/v1/contentdata/{id}/reject              content:review  (body: {"comment": "..."})
GET  /api/v1/reviews/pending                      content:review  (review queue)
```

Admin counterparts with `/admin/` prefix.

### 6.7 DbDriver Interface Additions

```go
// Content Reviews
CreateContentReview(context.Context, audited.AuditContext, CreateContentReviewParams) (*ContentReview, error)
GetContentReview(types.ContentReviewID) (*ContentReview, error)
GetPendingReview(types.ContentID, string) (*ContentReview, error)  // by content_data_id + locale
ListPendingReviews() (*[]ContentReview, error)
UpdateContentReviewStatus(context.Context, UpdateContentReviewStatusParams) error
CreateContentReviewTable() error

// Admin Content Reviews (parallel)
CreateAdminContentReview(context.Context, audited.AuditContext, CreateAdminContentReviewParams) (*AdminContentReview, error)
GetAdminContentReview(types.AdminContentReviewID) (*AdminContentReview, error)
GetAdminPendingReview(types.AdminContentID, string) (*AdminContentReview, error)
ListAdminPendingReviews() (*[]AdminContentReview, error)
UpdateAdminContentReviewStatus(context.Context, UpdateAdminContentReviewStatusParams) error
CreateAdminContentReviewTable() error
```

### 6.8 Typed IDs

- `ContentReviewID` — ULID
- `AdminContentReviewID` — ULID

### 6.9 Admin UI

**Content edit page (when datatype requires_review):**
- Editor without `content:publish`: "Submit for Review" button replaces "Publish"
- Reviewer with `content:review`: "Approve" / "Reject" buttons with comment textarea
- "Approve & Publish" combo button if reviewer also has `content:publish`
- Review status badge: yellow=pending, green=approved, red=rejected
- Rejection comment displayed to editor

**Review queue page (`/admin/reviews`):**
- List of pending reviews across all content
- Filter by datatype, locale, submitter
- Quick approve/reject from the list

**Datatype settings:**
- "Requires review before publishing" toggle per datatype

### 6.10 TUI

- Review queue screen
- Approve/reject actions with comment input

### 6.11 SDK Changes

All three SDKs:
- `ContentReview` type, `ContentReviewID` branded type
- `SubmitForReview(contentID)`, `ApproveReview(contentID)`, `RejectReview(contentID, comment)`, `ListPendingReviews()`

---

## Preview Environments & Draft Sharing

Phase 5. Standalone — builds on Phase 1 (delivery API with published snapshots) and Phase 2 (locale support).

### 7.1 preview_tokens Table

Schema number: `37_preview_tokens`

```sql
CREATE TABLE IF NOT EXISTS preview_tokens (
    token_id        TEXT PRIMARY KEY CHECK (length(token_id) = 26),
    token           TEXT NOT NULL UNIQUE,              -- random 32-byte hex string
    content_data_id TEXT NOT NULL REFERENCES content_data(content_data_id) ON DELETE CASCADE,
    locale          TEXT NOT NULL DEFAULT '',
    label           TEXT,                              -- "For client review"
    created_by      TEXT NOT NULL REFERENCES users(user_id),
    expires_at      TEXT NOT NULL,
    revoked         INTEGER NOT NULL DEFAULT 0,
    date_created    TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pt_token ON preview_tokens(token) WHERE revoked = 0;
CREATE INDEX idx_pt_content ON preview_tokens(content_data_id);
```

Admin counterpart: `admin_preview_tokens` with identical structure referencing `admin_content_data`.

### 7.2 Preview Delivery

New public endpoint (no auth required):

```
GET /api/v1/preview/{token}
```

Flow:
1. Look up token in `preview_tokens` where `revoked = 0`
2. Validate: not expired (`expires_at > now`)
3. Read live content from editing tables (same as `?preview=true` authenticated path)
4. Resolve locale from token's locale field
5. Feed through `BuildTree()` → `TransformAndWrite()` with `?format=` support
6. Response includes `X-Robots-Tag: noindex` and `X-Preview: true` headers
7. Response body includes `"preview": true, "expires_at": "..."` metadata

### 7.3 Config

```go
Preview_Enabled           bool `json:"preview_enabled"`           // default true
Preview_Token_Expiry_Days int  `json:"preview_token_expiry_days"` // default 7
Preview_Max_Per_Content   int  `json:"preview_max_per_content"`   // default 10
```

### 7.4 API Endpoints

```
POST   /api/v1/contentdata/{id}/preview-tokens          content:update  (create token)
GET    /api/v1/contentdata/{id}/preview-tokens           content:read    (list active tokens)
DELETE /api/v1/contentdata/{id}/preview-tokens/{tid}     content:update  (revoke token)
GET    /api/v1/preview/{token}                           no auth         (preview delivery)
```

Admin counterparts for token management.

### 7.5 DbDriver Interface Additions

```go
// Preview Tokens
CreatePreviewToken(context.Context, CreatePreviewTokenParams) (*PreviewToken, error)
GetPreviewTokenByToken(string) (*PreviewToken, error)
ListPreviewTokensByContent(types.ContentID) (*[]PreviewToken, error)
RevokePreviewToken(context.Context, types.PreviewTokenID) error
CleanExpiredPreviewTokens(types.Timestamp) error
CreatePreviewTokenTable() error

// Admin Preview Tokens (parallel)
CreateAdminPreviewToken(context.Context, CreateAdminPreviewTokenParams) (*AdminPreviewToken, error)
GetAdminPreviewTokenByToken(string) (*AdminPreviewToken, error)
ListAdminPreviewTokensByContent(types.AdminContentID) (*[]AdminPreviewToken, error)
RevokeAdminPreviewToken(context.Context, types.AdminPreviewTokenID) error
CleanExpiredAdminPreviewTokens(types.Timestamp) error
CreateAdminPreviewTokenTable() error
```

### 7.6 Typed IDs

- `PreviewTokenID` — ULID
- `AdminPreviewTokenID` — ULID

### 7.7 Admin UI

**Content edit page:**
- "Share Preview" button → dialog with locale selector + optional label + expiry picker
- Generated URL displayed with copy button
- List of active preview tokens with revoke buttons
- Token expiry countdown display

### 7.8 No TUI Changes

Preview token management is admin-facing only. TUI users can use the REST API directly.

### 7.9 SDK Changes

All three SDKs:
- `PreviewToken` type, `PreviewTokenID` branded type
- `CreatePreviewToken(contentID, locale, label)`, `ListPreviewTokens(contentID)`, `RevokePreviewToken(contentID, tokenID)`
- Content delivery SDK: `GetPreview(token)` method

---

## Bulk Operations & Content Dependencies

Phase 6. Standalone — enhanced by Phase 3 webhooks (fires events on bulk operations) but does not depend on it.

### 8.1 Dependency Resolution

New functions in `internal/publishing/`:

```go
// FindDependents returns content items that reference the given content_data_id
// via relation or content_tree_ref fields.
func FindDependents(ctx context.Context, driver db.DbDriver, contentDataID types.ContentID) ([]DependencyRef, error)

// FindDependencies returns content items that the given content references.
func FindDependencies(ctx context.Context, driver db.DbDriver, contentDataID types.ContentID) ([]DependencyRef, error)

type DependencyRef struct {
    ContentDataID types.ContentID
    RouteSlug     string
    FieldID       types.FieldID
    FieldLabel    string
    Direction     string  // "references" or "referenced_by"
}
```

### 8.2 Unpublish Warning

Updated unpublish flow (modifies Phase 1 unpublish):

1. Before unpublishing, call `FindDependents()`
2. If dependents exist AND any are published:
   - Return 409 with dependent list (force = false, default)
   - OR proceed if `?force=true` query param is set
3. Response includes list of dependent content that may show broken references

### 8.3 Bulk Publish API

```
POST /api/v1/contentdata/bulk-publish    content:publish
```

Body:
```json
{
  "items": [
    {"content_data_id": "...", "locale": "en"},
    {"content_data_id": "...", "locale": "es"}
  ],
  "include_dependencies": false
}
```

When `include_dependencies: true`, the resolver finds all referenced content and adds them to the publish set.

Response:
```json
{
  "published": 5,
  "failed": 1,
  "results": [
    {"content_data_id": "...", "locale": "en", "status": "published", "version_number": 3},
    {"content_data_id": "...", "locale": "en", "status": "failed", "error": "no content fields for locale"}
  ]
}
```

### 8.4 Publish All Locales

```
POST /api/v1/contentdata/{id}/publish-all-locales    content:publish
```

Publishes the content item in every enabled locale that has content_field rows. Returns per-locale results.

### 8.5 Bulk Schedule

```
POST /api/v1/contentdata/bulk-schedule    content:publish
```

Body:
```json
{
  "items": [
    {"content_data_id": "...", "locale": "en"},
    {"content_data_id": "...", "locale": "es"}
  ],
  "publish_at": "2026-03-15T09:00:00Z"
}
```

### 8.6 DbDriver Interface Additions

```go
// Dependency queries
ListContentFieldsByFieldValue(string) (*[]ContentFields, error)  // find fields referencing a content_data_id
ListPublishedContentByIDs([]types.ContentID) (*[]ContentData, error)

// Bulk operations
BulkUpdateContentDataSchedule(context.Context, []UpdateContentDataScheduleParams) error

// Admin parallels
ListAdminContentFieldsByFieldValue(string) (*[]AdminContentFields, error)
ListAdminPublishedContentByIDs([]types.AdminContentID) (*[]AdminContentData, error)
BulkUpdateAdminContentDataSchedule(context.Context, []UpdateAdminContentDataScheduleParams) error
```

### 8.7 API Endpoints

```
GET    /api/v1/contentdata/{id}/dependencies     content:read     (what this references)
GET    /api/v1/contentdata/{id}/dependents        content:read     (what references this)
POST   /api/v1/contentdata/bulk-publish           content:publish
POST   /api/v1/contentdata/bulk-schedule          content:publish
POST   /api/v1/contentdata/{id}/publish-all-locales  content:publish
```

Admin counterparts with `/admin/` prefix.

### 8.8 Admin UI

**Content list page:**
- Multi-select checkboxes on content rows
- Bulk action dropdown: "Publish Selected", "Schedule Selected", "Unpublish Selected"

**Content edit page:**
- "Dependencies" tab showing referenced and referencing content
- Unpublish confirmation dialog lists published dependents with warning
- "Publish All Locales" button (when i18n enabled, content has multiple locale translations)

### 8.9 TUI

- Multi-select in content list (space to toggle, enter to act)
- Bulk publish/unpublish/schedule actions

### 8.10 SDK Changes

All three SDKs:
- `DependencyRef` type
- `ListDependencies(contentID)`, `ListDependents(contentID)`
- `BulkPublish(items)`, `BulkSchedule(items, publishAt)`, `PublishAllLocales(contentID)`

---

## Implementation Phases

### Phase 1: Snapshot Publishing (foundation)

Everything else depends on this. The delivery API must stop serving draft content.

```
1a  Reduce ContentStatus to draft/published
    - Update types_enums.go
    - Migration: existing 'archived'/'pending' rows → 'draft' (both content_data and admin_content_data)
    - Update all handlers that reference old statuses

1b  Schema: content_versions + admin_content_versions tables, content_data + admin_content_data columns
    - SQL for all 3 databases × both public and admin tables
    - sqlc queries for both table families
    - Typed IDs: ContentVersionID, AdminContentVersionID

1c  DbDriver interface + 3 implementations
    - Public: CreateContentVersion, GetContentVersion, ListContentVersionsByContent,
      GetPublishedSnapshot, ClearPublishedFlag, PruneContentVersions
    - Admin: same methods for admin_content_versions
    - Shared snapshot builder logic via ContentDataJSON adapter layer

1d  Publish/unpublish logic
    - Snapshot builder: tree + fields → JSON (shared for public and admin via ContentDataJSON/FieldsJSON)
    - Publish flow (read tree + fields outside transaction; write transaction: insert version, clear old flag, update metadata)
    - Unpublish flow
    - ValidateTransition function (shared by all interfaces)
    - Fire existing plugin hooks

1e  New permission: content:publish
    - Add to bootstrap RBAC data
    - Wire into middleware for both public and admin endpoints

1f  Delivery API: switch to snapshot-based
    - GET /api/v1/content/{slug} reads from content_versions WHERE published = 1
    - Admin tree delivery reads from admin_content_versions WHERE published = 1
    - ?preview=true falls back to live tables (requires auth + content:read)
    - X-Robots-Tag: noindex on preview responses

1g  REST API endpoints
    - Public: POST /api/v1/contentdata/{id}/publish, unpublish, schedule
    - Admin: POST /api/v1/admincontentdatas/{id}/publish, unpublish, schedule

1h  Admin UI
    - Publish/unpublish action buttons (state + permission aware)
    - Status badges on list page
    - Status filter dropdown
    - Published metadata display

1i  Version history UI
    - Version list panel in content edit page
    - Manual "Save Version" button
    - Field-by-field comparison between versions

1j  Restore flow (field values only — no tree structure restore)
    - Pre-restore safety snapshot
    - Field restoration with schema mismatch handling
    - Unmapped field warnings in response

1k  Scheduled publishing
    - Background goroutine in cmd/serve.go
    - Catch-up pass on startup
    - Config: publish_schedule_interval
    - Handles both content_data and admin_content_data

1l  Optimistic locking
    - revision counter on content_data and admin_content_data
    - 409 Conflict response on revision mismatch

1m  Retention policy
    - Config: version_max_per_content
    - Per-save pruning + periodic cleanup (both table families)

1n  TUI updates
    - Status display + publish/unpublish actions
    - Version list + restore

1o  SDK updates
    - New types + methods across Go, TypeScript, Swift
```

### Phase 2: Internationalization (builds on Phase 1)

```
2a  Schema: locales table + locale columns + translatable columns
    - locales table (shared, system-wide): SQL for all 3 databases
    - content_fields.locale + admin_content_fields.locale: SQL for all 3 databases
    - fields.translatable + admin_fields.translatable: SQL for all 3 databases
    - sqlc queries for both table families
    - Typed ID: LocaleID
    - Migration: existing content_fields and admin_content_fields get locale = ''
    - Unique index on (content_data_id, field_id, locale) for both tables

2b  Config: i18n_enabled, i18n_default_locale
    - Config struct additions
    - Feature gate: all i18n code checks i18n_enabled

2c  DbDriver interface + 3 implementations
    - Locale CRUD methods (shared locales table)
    - Locale-aware field queries for both public and admin content_fields

2d  Locale management
    - REST API endpoints for locale CRUD
    - Admin settings page for locale management
    - Fallback chain validation (no cycles, max 2 hops)

2e  Translation creation workflow
    - "Add Translation" action: create content_field rows for target locale
    - Works for both public and admin content
    - Copy default locale values as starting point
    - Per-field translatable flag enforcement

2f  Locale-aware publishing
    - Publish flow gains locale parameter (both public and admin)
    - Snapshot includes locale-specific fields + non-translatable fields
    - Separate published snapshot per locale

2g  Delivery API locale resolution
    - ?locale= query param (both public and admin delivery)
    - Accept-Language header parsing
    - Fallback chain walking
    - Response metadata: locale, resolved_locale, fallback_used, available_locales
    - ?locale=* for all-locales response

2h  Admin UI
    - Locale tab bar on content edit page
    - Non-translatable field indicators
    - Translation completeness tracking
    - Per-locale publish buttons
    - Locale filter on content list

2i  TUI updates
    - Locale selector in content editing

2j  SDK updates
    - New types + methods across Go, TypeScript, Swift
```

### Phase 3: Webhooks & Content Lifecycle Events (builds on Phase 1)

```
3a  Schema: webhooks + webhook_deliveries tables
    - SQL for all 3 databases (no admin counterpart — system-wide)
    - sqlc queries
    - Typed IDs: WebhookID, WebhookDeliveryID

3b  DbDriver interface + 3 implementations
    - Webhook CRUD + delivery tracking methods

3c  Webhook delivery engine (internal/webhooks/)
    - Dispatch function, async worker pool, HMAC signing
    - Retry queue processing
    - Integration: call Dispatch from publish/unpublish/update handlers

3d  Background retry goroutine
    - Piggyback on publish scheduler ticker
    - Process pending retries

3e  Config additions
    - webhook_enabled, webhook_timeout, webhook_max_retries, etc.

3f  Permissions: webhook:create/read/update/delete
    - Bootstrap RBAC data update

3g  REST API endpoints
    - CRUD + test + delivery log + manual retry

3h  Admin UI
    - Webhook management page
    - Delivery log with retry

3i  TUI updates
    - Webhook management screens

3j  SDK updates
    - Types + methods across Go, TypeScript, Swift
```

### Phase 4: Review Workflows (builds on Phase 1, optionally Phase 3)

```
4a  Schema: requires_review column on datatypes/admin_datatypes
    - SQL for all 3 databases

4b  Status: add pending_review to ContentStatus enum
    - Update types_enums.go
    - Update status validation

4c  Schema: content_reviews + admin_content_reviews tables
    - SQL for all 3 databases
    - sqlc queries
    - Typed IDs: ContentReviewID, AdminContentReviewID

4d  DbDriver interface + 3 implementations
    - Review CRUD methods for both table families

4e  Review flow logic
    - Submit, approve, reject handlers
    - Publish flow updated: check requires_review + approved status
    - Bypass for users with content:publish

4f  Permissions: content:review
    - Bootstrap RBAC data update

4g  Webhook events
    - content.review_submitted, content.review_approved, content.review_rejected

4h  REST API endpoints
    - Submit, approve, reject, review queue

4i  Admin UI
    - Review queue page
    - Content edit page review state
    - Datatype requires_review toggle

4j  TUI updates
    - Review queue + approve/reject

4k  SDK updates
    - Types + methods across Go, TypeScript, Swift
```

### Phase 5: Preview Environments & Draft Sharing (standalone)

```
5a  Schema: preview_tokens + admin_preview_tokens tables
    - SQL for all 3 databases
    - sqlc queries
    - Typed IDs: PreviewTokenID, AdminPreviewTokenID

5b  DbDriver interface + 3 implementations
    - Token CRUD + lookup + cleanup methods

5c  Preview delivery endpoint
    - Token validation, live content read, format support
    - X-Robots-Tag and preview metadata in response

5d  Token management endpoints
    - Create, list, revoke

5e  Expired token cleanup
    - Piggyback on publish scheduler ticker

5f  Config additions
    - preview_enabled, preview_token_expiry_days, preview_max_per_content

5g  Admin UI
    - Share Preview dialog on content edit page
    - Token list with revoke

5h  SDK updates
    - Types + methods across Go, TypeScript, Swift
```

### Phase 6: Bulk Operations & Content Dependencies (standalone)

```
6a  Dependency resolver (internal/publishing/)
    - FindDependents, FindDependencies functions
    - Scans relation/content_tree_ref field values

6b  Unpublish warning
    - Modified unpublish flow: check dependents, return 409 or proceed with force
    - Works for both public and admin

6c  Bulk publish endpoint + logic
    - Iterate items, publish each, collect results
    - Optional dependency inclusion

6d  Publish all locales endpoint
    - Query enabled locales with content_field rows, publish each

6e  Bulk schedule endpoint
    - Set publish_at on multiple items

6f  Dependency API endpoints
    - GET dependencies/dependents

6g  Admin UI
    - Multi-select on content list
    - Bulk action dropdown
    - Dependencies tab on content edit
    - Unpublish warning dialog
    - Publish All Locales button

6h  TUI updates
    - Multi-select + bulk actions

6i  SDK updates
    - Types + methods across Go, TypeScript, Swift
```

### Scope per phase

| Phase | New tables | New typed IDs | New permissions | Estimated files |
|-------|-----------|---------------|-----------------|----------------|
| 1 | content_versions, admin_content_versions | ContentVersionID, AdminContentVersionID | content:publish | ~45 |
| 2 | locales | LocaleID | locale:create/read/update/delete | ~35 |
| 3 | webhooks, webhook_deliveries | WebhookID, WebhookDeliveryID | webhook:create/read/update/delete | ~40 |
| 4 | content_reviews, admin_content_reviews | ContentReviewID, AdminContentReviewID | content:review | ~35 |
| 5 | preview_tokens, admin_preview_tokens | PreviewTokenID, AdminPreviewTokenID | (none, uses content:update) | ~30 |
| 6 | (none) | (none) | (none) | ~25 |

### Dependency Chain

```
Phase 1 (DONE) ──> Phase 3 (Webhooks)
Phase 2 (DONE) ──> Phase 4 (Review) ──> uses Phase 3 webhooks for notifications
                ──> Phase 5 (Preview) ──> standalone
                ──> Phase 6 (Bulk Ops) ──> standalone, enhanced by Phase 3 webhooks
```

Phase 3 should ship first (other phases fire webhook events). Phases 4, 5, 6 are independent and can ship in any order.

---

## Schema Changes

### New Tables

| # | Table | Phase |
|---|-------|-------|
| 27 | `content_versions` | 1 |
| 27a | `admin_content_versions` | 1 |
| 28 | `locales` | 2 |
| 34 | `webhooks` | 3 |
| 35 | `webhook_deliveries` | 3 |
| 36 | `content_reviews` | 4 |
| 36a | `admin_content_reviews` | 4 |
| 37 | `preview_tokens` | 5 |
| 37a | `admin_preview_tokens` | 5 |

The `locales` and `webhooks` tables are shared between public and admin content — definitions are system-wide.

### Altered Tables

| Table | Change | Phase |
|-------|--------|-------|
| `content_data` | Add `published_at`, `published_by`, `publish_at`, `revision` | 1 |
| `admin_content_data` | Add `published_at`, `published_by`, `publish_at`, `revision` | 1 |
| `content_fields` | Add `locale` column + unique index | 2 |
| `admin_content_fields` | Add `locale` column + unique index | 2 |
| `fields` | Add `translatable` column | 2 |
| `admin_fields` | Add `translatable` column | 2 |
| `datatypes` | Add `requires_review` column | 4 |
| `admin_datatypes` | Add `requires_review` column | 4 |

### New Typed IDs

| Type | Phase |
|------|-------|
| `ContentVersionID` | 1 |
| `AdminContentVersionID` | 1 |
| `LocaleID` | 2 |
| `WebhookID` | 3 |
| `WebhookDeliveryID` | 3 |
| `ContentReviewID` | 4 |
| `AdminContentReviewID` | 4 |
| `PreviewTokenID` | 5 |
| `AdminPreviewTokenID` | 5 |

### New Permissions

| Permission | Phase |
|------------|-------|
| `content:publish` | 1 |
| `locale:create` | 2 |
| `locale:read` | 2 |
| `locale:update` | 2 |
| `locale:delete` | 2 |
| `webhook:create` | 3 |
| `webhook:read` | 3 |
| `webhook:update` | 3 |
| `webhook:delete` | 3 |
| `content:review` | 4 |

### Migration Notes

There are no active deployments requiring a phased migration strategy. Schema changes and code changes deploy together.

**Phase 1 migration:**
- `UPDATE content_data SET status = 'draft' WHERE status IN ('archived', 'pending')` — runs in schema migration
- Same for `admin_content_data`
- `ContentStatus.Validate()` drops `archived` and `pending` in the same deploy
- `ALTER TABLE content_data ADD COLUMN revision INTEGER NOT NULL DEFAULT 0` — all rows start at 0
- Same for `admin_content_data`
- No published snapshots exist yet — all existing content is effectively draft
- If the site has live traffic, existing content should be bulk-published after migration (one-time script that calls the publish flow for each root content item, bypassing hook execution for performance)

**Phase 2 migration:**
- Existing content_fields rows get `locale = ''` (the DEFAULT handles this)
- Same for admin_content_fields
- Existing fields get `translatable = 0`
- Same for admin_fields
- No locales table rows until admin creates them
- `i18n_enabled` defaults to `false` — zero behavior change until explicitly enabled

**Phase 3 migration:**
- New tables only — no existing data affected
- `webhook_enabled` defaults to `false` — zero behavior change until explicitly enabled
- New bootstrap permissions added for admin role

**Phase 4 migration:**
- `ALTER TABLE datatypes ADD COLUMN requires_review INTEGER NOT NULL DEFAULT 0` — all existing datatypes default to no review
- Same for `admin_datatypes`
- `ContentStatus` gains `pending_review` value — existing content is unaffected (only triggered by explicit submit-for-review action)
- New `content:review` permission added to admin role bootstrap

**Phase 5 migration:**
- New tables only — no existing data affected
- `preview_enabled` defaults to `true` — feature is available immediately but no tokens exist until created

**Phase 6 migration:**
- No schema changes — all functionality is query-based and uses existing tables
- New API endpoints are additive

---

## Verification

After implementing each phase:

1. `go build ./...` passes cleanly
2. `go test ./internal/db/` passes (includes new table CRUD)
3. `go test ./internal/publishing/` passes (webhook dispatch, review flow, preview, bulk ops)
4. `just sqlc` regenerates cleanly
5. `just admin-verify` confirms templ files are up-to-date
6. `just sdk-build && just sdk-test` passes for TypeScript SDKs
7. `just sdk-go-test` passes for Go SDK
8. `just sdk-swift-build` passes for Swift SDK
9. Manual test: create webhook, publish content, verify delivery received
10. Manual test: enable review workflow, submit/approve/reject cycle
11. Manual test: generate preview token, access preview URL without auth
12. Manual test: bulk publish multiple items, verify all snapshots created
