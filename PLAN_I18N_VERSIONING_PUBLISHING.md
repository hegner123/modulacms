# Plan: Publishing, Versioning, and i18n

**Status:** Reviewed — must-fix items addressed
**Date:** 2026-02-28

---

## Table of Contents

1. [Current State](#current-state)
2. [Design Decisions](#design-decisions)
3. [Publishing and Versioning (Unified)](#publishing-and-versioning)
4. [Internationalization (i18n)](#internationalization)
5. [Implementation Phases](#implementation-phases)
6. [Schema Changes](#schema-changes)

---

## Current State

| Area | State |
|------|-------|
| Content statuses | `draft`, `published`, `archived`, `pending` defined — no transition enforcement |
| Public delivery | Serves all statuses — **draft content is publicly accessible** |
| Publish permission | None — publishing is just `content:update` |
| Plugin hooks | `before_publish`/`after_publish` already fire on status transitions |
| Audit trail | `change_events` records old/new JSON for every mutation |
| Locale support | None |
| Content structure | Tree with sibling pointers on `content_data`; field values in separate `content_fields` table (EAV) |

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

### Review workflows: Not yet

**Why defer?** Every CMS team that built complex review workflows before knowing whether anyone would use them regretted it. Common regrets: too many stages, mandatory review for all content types, no admin bypass. The simplest approach that works: permission-gated publishing. Only users with `content:publish` can create published snapshots. Add workflow stages later when a real use case demands it.

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

### Scope per phase

| Phase | New tables | New typed IDs | Estimated files touched |
|-------|-----------|---------------|------------------------|
| 1 | content_versions, admin_content_versions | ContentVersionID, AdminContentVersionID | ~45 (schema × 3 DBs × 2 table families, sqlc, db interface, 3 wrappers × 2, router, handlers, admin pages/partials, TUI, config, 3 SDKs) |
| 2 | locales | LocaleID | ~35 (schema × 3 DBs × 2 table families, sqlc, db interface, 3 wrappers × 2, router, handlers, admin pages/partials, config, 3 SDKs) |

---

## Schema Changes

### New Tables

| # | Table | Phase |
|---|-------|-------|
| 27 | `content_versions` | 1 |
| 27a | `admin_content_versions` | 1 |
| 28 | `locales` | 2 |

The `locales` table is shared between public and admin content — locale definitions are system-wide.

### Altered Tables

| Table | Change | Phase |
|-------|--------|-------|
| `content_data` | Add `published_at`, `published_by`, `publish_at`, `revision` | 1 |
| `admin_content_data` | Add `published_at`, `published_by`, `publish_at`, `revision` | 1 |
| `content_fields` | Add `locale` column + unique index | 2 |
| `admin_content_fields` | Add `locale` column + unique index | 2 |
| `fields` | Add `translatable` column | 2 |
| `admin_fields` | Add `translatable` column | 2 |

### New Typed IDs

| Type | Phase |
|------|-------|
| `ContentVersionID` | 1 |
| `AdminContentVersionID` | 1 |
| `LocaleID` | 2 |

### New Permissions

| Permission | Phase |
|------------|-------|
| `content:publish` | 1 |
| `locale:create` | 2 |
| `locale:read` | 2 |
| `locale:update` | 2 |
| `locale:delete` | 2 |

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
