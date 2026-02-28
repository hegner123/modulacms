# CMS Publishing, Versioning, and i18n — Research Reference

**Purpose:** Decision-making reference for ModulaCMS. Covers best practices, implementation mechanisms, common footguns, and proven solutions drawn from how Strapi v5, Contentful, Sanity, Payload CMS, Directus, WordPress, Wagtail, and DatoCMS actually work.

---

## Table of Contents

1. [Publishing Workflows](#1-publishing-workflows)
2. [Content Versioning and Rollback](#2-content-versioning-and-rollback)
3. [Internationalization (i18n)](#3-internationalization-i18n)
4. [Cross-Cutting Concerns](#4-cross-cutting-concerns)

---

## 1. Publishing Workflows

### Best Practices

**Keep status values minimal.** The industry has converged on two core statuses: `draft` and `published`. Everything else is either derived or belongs in a separate layer.

| CMS | Core Statuses | Notes |
|-----|--------------|-------|
| Strapi v5 | `draft`, `published` | "Modified" is a UI indicator (draft differs from published copy), not a stored status |
| Payload CMS | `draft`, `published` | Stored as `_status` field |
| Sanity | Draft doc exists, published doc exists | Two separate documents — if both exist, content is "modified" |
| Contentful | Derived from version numbers | `sys.version` vs `sys.publishedVersion` comparison |
| WordPress | 8 statuses | `auto-draft`, `draft`, `pending`, `future`, `publish`, `private`, `trash`, `inherit` — widely considered overcomplicated |

**"Scheduled" should be a timestamp field, not a status.** WordPress makes `future` a status, which means a scheduled post is neither draft nor published — it's its own thing that breaks queries. Payload CMS, Sanity, and Contentful all keep scheduled content as `draft` with a `publish_at` timestamp. A background worker transitions it when the time arrives. The document's status always reflects what it *actually is right now*.

**Review/approval is a separate layer from publication status.** Strapi v5 Enterprise has review workflow stages ("To do" → "In progress" → "Ready to review" → "Reviewed") that are *orthogonal* to draft/published. A document can be in review stage "Approved" but still in `draft` status. This prevents internal workflow concerns from leaking into the delivery API. If you couple them (making "pending review" a content status), your delivery API must understand review stages.

**The simplest review workflow that works:** Permission-gated publishing. Editors can save drafts. Only users with a `publish` permission can transition to published. This is what Payload CMS does. No workflow stages, no assignment system — just access control on the publish action. Add stages later only if teams actually need them.

**What teams regret about review workflows:**
- Too many stages (6-8 stages that mirror their Jira board — editors hate it)
- Mandatory review for all content types (applying the same workflow to blog posts and banner text changes)
- No admin bypass for hotfixes
- Building review stages before knowing whether anyone will use them

### Mechanisms

**The two-copy pattern (draft/published separation).** This is the dominant approach in modern headless CMSs. Two representations exist simultaneously:

- A mutable **draft** working copy (editors freely modify this)
- An immutable **published** frozen copy (the delivery API serves this)
- Publishing atomically copies draft → published
- Edits never touch the published copy directly

Strapi v5 implements this as two database rows per document linked by `documentId`. Payload CMS stores the published state in the main collection table and drafts in a versions table. Contentful uses version numbers (published version is a frozen snapshot at `sys.publishedVersion`).

**Why two-copy won over single-record-with-status:** With a single record, saving a draft of already-published content either (a) overwrites the live version (unacceptable) or (b) requires bolting on a separate draft mechanism (which reinvents two-copy anyway). The two-copy pattern means editors can work on changes over days/weeks without affecting the live site.

**Scheduled publishing implementation.** Background worker polling every 60 seconds for `publish_at <= now() AND status = 'draft'`. On startup, immediately run a catch-up pass for anything that was due while the server was down. The publish action must be the same code path as manual publishing (hooks, webhooks, audit events all fire). The publish operation must be idempotent — publishing an already-published document should be a no-op.

**Published-only delivery.** Three patterns exist:

| Pattern | Who | How |
|---------|-----|-----|
| Separate API hosts | Contentful | `cdn.contentful.com` (published only) vs `preview.contentful.com` (drafts) with separate tokens |
| Query parameter | Strapi, Payload, Sanity | `?status=published` (default for unauthenticated) vs `?status=draft` (requires auth) |
| Access control only | WordPress | Anonymous = published only; authenticated with permission = can see drafts |

The query parameter approach is most practical for a self-hosted CMS. Default to published-only for unauthenticated requests. Require authentication + permission for draft access.

**Preview mode** standard flow: CMS generates a URL with an HMAC-signed short-lived token → frontend validates token, sets a preview cookie → while cookie is active, frontend fetches from draft API → cookie has short TTL (5-15 minutes). Preview pages must include `<meta name="robots" content="noindex, nofollow">`.

### Footguns and Solutions

**Footgun: Adding published-only filtering to an existing system that served everything.**
When you add a status filter, existing content must be classified. Default to `published` (treating existing content as live) — defaulting to `draft` makes the public API return nothing.

**Footgun: Webhook consumers break.**
Before publishing existed, consumers subscribed to `entry.update`. Now they receive events for draft saves they should ignore. You need distinct event types: `entry.published`, `entry.unpublished`, `entry.draft_saved`.

**Footgun: Race condition between editors and publishers.**
Editor A is making changes. Publisher B publishes the current version. Editor A saves, overwriting the just-published version. The two-copy pattern prevents this entirely — Editor A's saves only touch the draft copy, which is separate from the published copy.

**Footgun: Preview URLs indexed by search engines.**
Preview URLs with long-lived or guessable tokens get crawled. Fix: short-lived tokens (5-15 min), `noindex` meta tags, serve previews from a different subdomain blocked in `robots.txt`.

**Footgun: Scheduled content fails validation at publish time.**
Content scheduled for publication, then the schema changes (field removed, validation added). When the scheduler fires, publish fails. Fix: validate at schedule time AND at publish time. On failure, mark the job as failed and notify the author — don't silently skip it.

**Footgun: Publishing a node in a content tree while children are in draft.**
Parent is published, children are still draft. In the published view, the page loads but sections are missing. Fix: the root content entry is the unit of publishing — publish snapshots the entire tree atomically. No per-node publishing. This is what WordPress Gutenberg does (blocks are embedded in the post, not separate publishable entities).

**Footgun: Unpublishing a parent leaves children orphaned.**
Children are published and reachable by direct URL, but the parent (their navigation entry point) is gone. Fix: unpublishing a parent must cascade-warn ("This will also unpublish 5 child items") or block if children are published.

---

## 2. Content Versioning and Rollback

### Best Practices

**Use full document snapshots, not event sourcing.** Every major CMS uses snapshots. No CMS uses pure event sourcing for content versioning. Event sourcing adds complexity (replaying events, handling schema evolution in event streams) without benefit for content use cases. Store the complete state at a point in time.

**Adopt the two-copy (draft/published) pattern** as the foundation for versioning. This naturally creates the concept of "the current draft" and "the current published version." Version history is a series of snapshots of the draft copy. Publishing freezes the current draft as the new published version.

| CMS | Snapshot Trigger | What's Stored |
|-----|-----------------|---------------|
| WordPress | Every save (auto + manual) | Full post content (title, body, excerpt, author). Custom fields are NOT versioned — major limitation |
| Wagtail | Every `save_revision()` | Full JSON-serialized page state |
| Payload CMS | Every save; optionally on interval (autosave) | Full document snapshot in `{slug}_versions` table |
| Contentful | On publish | Snapshot at `sys.publishedVersion` |
| Sanity | Every keystroke/patch | Patches stored as transactions (closest to event sourcing) |
| Strapi v5 | Every save (Content History, paid feature) | Full snapshot |
| Directus | Every mutation | Full snapshot + delta of changes |

**Version history should be append-only.** Restoring an old version creates a *new* version with the old data (copy-forward). You never lose intermediate history. Every CMS does this — none use pointer-swap.

**Restore should create a draft, not go live.** Safer: the old version becomes the current draft, giving editors a chance to review before publishing. WordPress is the outlier — it restores directly to the live post. Strapi, Payload, Wagtail, and Directus all restore to draft.

**Use sequential integers for display, timestamps for querying.** No CMS uses semantic versioning for content. Version 1, 2, 3... is the human-readable label. The `created_at` timestamp is how you query ("show me what this looked like last Tuesday").

**Set a retention cap.** Unbounded version storage causes real problems:
- WordPress case study: 13,779 revisions deleted reduced database from 297MB to 65MB
- Payload CMS: `maxPerDoc` config option (cap per document)
- Wagtail: `purge_revisions` management command with `--days` flag
- Directus: `REVISIONS_RETENTION` env var (e.g., `30d`)
- Sanity: Time-based by plan (Free: 3 days, Growth: 90 days, Enterprise: 365 days)
- Reasonable default: **keep 20-50 versions per document**, always preserve current published + current draft, optionally prune by age

### Mechanisms

**Snapshot storage.** A separate table (e.g., `content_versions`) with:
- Reference to the parent content item
- Version number (sequential integer)
- Full field data as JSON
- Metadata: created_by, created_at, trigger (manual/publish/autosave/restore)
- Optional label (user-provided, e.g., "launch copy v2")

Payload CMS creates `{slug}_versions` tables automatically. Directus uses `directus_revisions`. WordPress reuses `wp_posts` with `post_type = 'revision'`.

**Restore flow** (the standard across all CMSs):
1. Snapshot current state as a new version (safety net)
2. Read the target version's snapshot
3. Update the document with the snapshot data
4. Set status to `draft` (require explicit re-publish)
5. Record in audit trail

**Schema change handling on restore.** The hardest problem. When a field has been added/removed since the snapshot was taken:

| CMS | Approach |
|-----|----------|
| Strapi v5 | Shows orphaned fields under "Unknown fields" section — editor can see old data that no longer maps |
| DatoCMS | Retroactively modifies ALL existing versions when schema changes (loses historical field data) |
| Payload CMS | Unknown fields ignored on restore; missing required fields may fail validation |
| WordPress | Not a problem because custom fields aren't versioned (but that's its own problem) |

Best approach: **store snapshots as immutable JSON. Never retroactively modify them.** On restore, map old fields to current schema with best effort. Show a clear UI for unmapped fields ("This version contained a field 'subtitle' that no longer exists. Its value was: '...'"). Let the editor decide what to do.

**Diff and comparison.** Every CMS that offers comparison does field-by-field, side-by-side. Rich text diffing is the hard part:
- Render both to HTML, diff at word level (Wagtail approach — acknowledged as "far from ideal")
- Compare underlying AST (Contentful's rich text versioning app)
- Plain text fallback (strip formatting, diff as text)

Start with field-by-field comparison showing old value / new value. Rich text diff is a stretch goal — even mature CMSs struggle with it.

### Footguns and Solutions

**Footgun: Snapshots don't capture referenced content.**
The snapshot stores `author_id: "abc"`, not a copy of the author record. If the author is deleted between snapshot and restore, the reference is broken. DatoCMS acknowledges this: "When restoring a record containing a reference to a deleted upload, the image field will be empty."
*Fix:* Validate all references on restore. Warn about broken references. Optionally store denormalized display data (author name, image URL) in the snapshot for the diff UI, even though the canonical reference is an ID.

**Footgun: Restoring a version that references deleted content/media.**
The restored content has a broken image, a missing related article, a dangling category reference.
*Fix:* On restore, check every foreign-key reference. For each that no longer exists: null out the field and show a warning ("Referenced image was deleted"). Optionally block restore and list what's missing.

**Footgun: Concurrent editing creates confusing version history.**
Two editors edit the same document. Editor A saves (version 5). Editor B, working from version 4, saves (version 6, overwriting A's changes).
*Fix:* Optimistic locking — require the current version number with every update. If someone else saved first, reject with HTTP 409 Conflict. The editor must re-fetch and retry. Contentful does this via `X-Contentful-Version` header. Simpler than document locking, much simpler than CRDTs.

**Footgun: Version storage grows unbounded.**
*Fix:* `maxPerDoc` cap (Payload approach). On every new version creation: if count exceeds cap, delete the oldest versions that (a) are not the current published version, (b) are not the current draft, and (c) do not have a user-provided label.

**Footgun: Schema evolution breaks old snapshots.**
Field `subtitle` renamed to `tagline`, field `legacy_flag` removed. Old snapshots still contain the old field names.
*Fix:* Never retroactively modify snapshots (that destroys history). Handle mismatches at restore time with clear UI. Store a schema version or field-ID-based references (not field names) in snapshots so renames don't orphan data.

---

## 3. Internationalization (i18n)

### Best Practices

**The model choice is the most consequential decision.** Three real options:

| Model | How It Works | Best For |
|-------|-------------|----------|
| **Field-level** | Each translatable field stores `{locale: value}` map. One record per content item. | Identical structure across locales. Single source of truth. |
| **Row-per-locale** | Each locale is a separate record linked by a group ID. | Independent publishing per locale. Different structures per locale. |
| **Junction table** | Shared fields on parent table, translated fields in a separate translations table. | Relational databases. Zero duplication of non-translatable fields. |

Who uses what:

| CMS | Model |
|-----|-------|
| Contentful | Field-level |
| Payload CMS | Field-level |
| DatoCMS | Field-level |
| Strapi v5 | Row-per-locale |
| WordPress/WPML | Row-per-locale |
| Wagtail | Row-per-locale (tree-per-locale) |
| Sanity | Document-per-locale (plugin) |
| Directus | Junction table |

**Per-locale publishing should be a first-class feature.** Contentful launched without it, treating publish as all-locales-at-once. They later had to build "Locale-Based Publishing" as a premium add-on because customers demanded it. Every modern CMS now supports or defaults to independent publishing per locale:
- Strapi v5: Each locale row has its own status. Publishing one locale does not affect others.
- DatoCMS: "Publishing changes in one locale doesn't affect the status of that item in any other locale."
- Row-per-locale models get this for free (each row is independent).
- Field-level models must add it as a separate feature.

**Not every field should be translated.** Prices, dates, booleans, sort orders, external URLs — these are language-independent. Translating them is nonsensical and creates maintenance burden.

| CMS | How non-translatable fields work |
|-----|--------------------------------|
| Payload CMS | Only fields with `localized: true` are translated. Others remain scalar. |
| Contentful | Per-field "Enable localization" checkbox. |
| Wagtail | `SynchronizedField("price")` — auto-propagated from source to all translations. |
| Directus | Non-translatable fields live on the parent table (not the translations junction). Zero duplication. |
| Strapi v5 (row-per-locale) | **Duplicated across rows.** Updating a shared field requires updating N locale rows. No automatic sync. |

The junction table model (Directus) handles non-translatable fields most cleanly. Row-per-locale models (Strapi, WordPress) are the worst for this — data duplication is unavoidable without additional sync logic.

**Use BCP 47 locale codes with hyphens.** `en`, `en-US`, `fr-CA` — not underscores (`en_US`). BCP 47 aligns with web standards (HTML `lang`, HTTP `Accept-Language`, `hreflang` tags). Normalize on input. Do not allow both formats to coexist. Storyblok uses underscores and it causes interoperability problems.

**Fallback chains should be configurable and visible.** `fr-CA → fr → en` is a common pattern. But fallback is dangerous when invisible — a French-Canadian user silently receiving France-French content with different pricing, terminology, or legal language is worse than showing "not available in your language."

The API should signal when fallback was used. No major CMS does this well today — it's an opportunity.

### Mechanisms

**Locale resolution for content delivery** (headless API):

1. **Explicit query param** `?locale=es` — highest priority. This is the universal standard across headless CMSs.
2. **Accept-Language header** — secondary. Parsed and matched against enabled locales.
3. **Default locale** — fallback when no match found.

URL-path-based resolution (`/es/about`) is a **frontend** concern, not a CMS API concern. The CMS serves content via API; the frontend framework maps URL paths to API locale parameters.

**Response shape** — single locale per response is the standard:
```json
{
  "id": "abc123",
  "locale": "es",
  "title": "Sobre Nosotros",
  "body": "...",
  "meta": {
    "available_locales": ["en", "es", "fr"],
    "fallback_used": false
  }
}
```

Support `?locale=*` for returning all locales in one response (Payload CMS supports this). Essential for SSG/ISR scenarios where the frontend needs to know which locales exist.

**Translation creation workflow:**
- Most CMSs copy source content as a starting point (not blank fields)
- Strapi v5: "Fill in from another locale" action
- Wagtail/wagtail-localize: Automatically creates translation pages with source content, tracks translation completeness per-segment
- Completeness tracking ("12/15 fields translated") is valuable but can be a stretch goal

**Content relations across locales:**
If a blog post in Spanish references a category, the API must return the Spanish category, not the English one. Field-level models (Payload, Contentful) handle this automatically — the relation resolves within the requested locale context. Row-per-locale models must explicitly link to the correct locale variant or resolve at query time.

### Footguns and Solutions

**Footgun: Serving broken pages from untranslated required fields.**
A required field in the default locale is empty in a secondary locale. The CMS publishes it. The frontend renders a broken page.
*Fix:* Per-locale validation before publishing. Block publishing a locale variant that has empty required fields. Or: serve fallback content for empty fields and flag it in the API response.

**Footgun: Fallback chains silently serve wrong-language content.**
User requests `fr-CA`, gets `fr` (France-French) silently. Legal language, pricing, and terminology may differ.
*Fix:* Include `fallback_used: true` and `resolved_locale: "fr"` in API responses. Let the frontend decide whether to show the fallback content or display a "not available in your language" message.

**Footgun: Structural divergence between locale variants.**
In row-per-locale models, the English page has 5 blocks, Spanish has 3. Frontend code that assumes structural consistency breaks.
*Fix:* For models with content trees/blocks — make structure a synchronized (non-translatable) concern. Wagtail's `SynchronizedField` propagates structural changes from source to all translations. Or: accept divergence as a feature (some markets genuinely need different page layouts).

**Footgun: Non-translatable fields get out of sync in row-per-locale models.**
Price updated on English row but not Spanish row. Now two locales show different prices.
*Fix (if row-per-locale):* Build a sync mechanism. When a non-translatable field is updated on any locale row, propagate the change to all rows in the same locale group. Wagtail's synchronized fields do this automatically.
*Fix (if junction table):* Non-translatable fields live on the parent table. One source of truth. No sync needed.

**Footgun: Content relations point to wrong locale.**
Spanish blog post references English category because the relation stores the default-locale entity ID.
*Fix:* Relations should point to the locale group (not a specific locale row). At query time, resolve the relation to the requested locale. If the target content doesn't exist in the requested locale, fall back through the chain.

**Footgun: Deleting the default locale variant orphans translations.**
*Fix:* Prevent deletion of the default locale variant. This is what Contentful does ("The first locale you create is your primary locale — you can't later make a secondary locale your primary locale"). Allow reassigning the default locale before deletion, but never allow a group with no default.

**Footgun: Converting a field from non-localized to localized (or vice versa) is destructive.**
Payload CMS documents this: "When converting an existing field to or from `localized: true` the data structure in the document will change for this field and so existing data for this field will be lost."
*Fix:* Make the localized/non-localized decision at field definition time and treat changing it as a migration with explicit data handling (copy default locale values to all locale rows, or collapse locale rows to a single value).

**Footgun: SEO problems from incomplete i18n.**
Missing or non-reciprocal hreflang tags. Duplicate content across locales. Geo-redirects blocking crawlers.
*Fix:* This is a frontend concern, but the CMS API should provide the data needed: available locales per content item, canonical locale, and published status per locale. The frontend generates hreflang tags, canonical URLs, and localized sitemaps from this data.

**Footgun: Performance impact of locale resolution on every request.**
WPML adds 20+ database queries per page for string translation. Polylang is lighter (~5% overhead).
*Fix:* Row-per-locale with a simple WHERE clause is fast. Field-level with JSON parsing is moderate. Junction table JOINs are moderate. Cache locale lookups aggressively — the list of enabled locales changes infrequently.

---

## 4. Cross-Cutting Concerns

### Webhooks / Event Integration

Publish events need to reach external systems:

| Consumer | Why | When |
|----------|-----|------|
| CDN / edge cache | Purge cached pages | On publish/unpublish |
| SSG / ISR | Rebuild or revalidate specific paths | On publish (can debounce) |
| Search index | Update/remove entries | On publish/unpublish |
| Translation service | Queue new content | On first publish |
| Email / notifications | Notify subscribers | On first publish only |

Standard event types (following Strapi v5): `entry.create`, `entry.update`, `entry.delete`, `entry.publish`, `entry.unpublish`. Webhook consumers must be idempotent — Contentful retries up to 3x with 30-second delays.

ModulaCMS already has plugin hooks (`before_publish`, `after_publish`, etc.). Webhooks would be a separate system — HTTP POST to registered URLs with JSON payloads. The existing hook infrastructure can trigger webhook dispatch.

### Optimistic Locking

Multiple editors working on the same content need conflict detection. The standard approach: every update request includes the expected current version number. If the actual version differs (someone else saved), return HTTP 409 Conflict. The editor must re-fetch and retry.

Contentful implements this via `X-Contentful-Version` header. This is simpler than document locking (which requires heartbeats, timeout handling, and lock-break mechanisms) and much simpler than CRDTs/OT (which require a real-time collaboration layer).

### Bulk Operations

Real CMSs support bulk publish/unpublish from list views. Partial failure handling: execute each operation independently, collect results, return summary. Do not roll back successful operations — publishing has side effects (webhooks, CDN purges) that cannot be rolled back.

```json
{
  "total": 10,
  "published": 8,
  "failed": 2,
  "errors": [
    {"id": "abc", "error": "required field 'title' is empty"},
    {"id": "def", "error": "version conflict"}
  ]
}
```

### How These Three Features Interact

Publishing + Versioning: Publishing should auto-create a version snapshot. The published copy is always a specific version. "Restore version N" always restores to draft (requires re-publish).

Publishing + i18n: Each locale variant has its own publication status. Publishing English does not publish Spanish. The delivery API filters by `status = published AND locale = ?`.

Versioning + i18n: Each locale variant has its own version history. Restoring version 3 of the Spanish variant does not affect the English variant. Version snapshots include the locale they belong to.

All three together: A content item has N locale variants. Each variant has its own draft/published status and its own version history. Publishing a locale variant creates a version snapshot for that variant. The delivery API serves the published version of the requested locale, with fallback chain.
