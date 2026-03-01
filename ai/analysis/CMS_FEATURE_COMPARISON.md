# CMS Feature Comparison: Publishing, Versioning, and i18n

**Updated:** 2026-02-28
**Status:** Phases 1-2 COMPLETE | Phases 3-6 PLANNED

Competitive analysis of ModulaCMS against Strapi v5, Contentful, Payload CMS, Sanity, Directus, WordPress, Wagtail, and DatoCMS in the publishing/versioning/i18n domain.

---

## Publishing & Versioning (Phase 1 -- DONE)

| Feature | ModulaCMS | Strapi v5 | Contentful | Payload | Sanity | WordPress |
|---------|-----------|-----------|------------|---------|--------|-----------|
| Draft/published separation | Snapshot-based | Row duplication | Version numbers | Row duplication | Two documents | 8 statuses |
| Scheduled publishing | Background ticker | Enterprise only | Scheduled actions | Custom | Scheduled publishing | Built-in (`future` status) |
| Version history | Snapshot per publish + manual | Enterprise only | Auto on publish | Draft autosave | Revision history | Post revisions |
| Version restore | Field-values only (safe) | Enterprise only | Revert to version | Restore draft | Revert | Restore revision |
| Optimistic locking | Revision counter + 409 | `updatedAt` check | `sys.version` | None | None | None |
| Retention policy | Configurable cap (default 50) | None | Unlimited | None | None | `wp_revisions_to_keep` |
| Preview mode | `?preview=true` + auth | Draft preview | Preview API | Draft preview | Preview pane | Preview button |
| Permission-gated publish | `content:publish` | Roles & stages | Roles | Access control | Custom | `publish_posts` capability |

### Where ModulaCMS leads

**Optimistic locking.** Revision counter with HTTP 409 on conflict is more reliable than Strapi's `updatedAt` timestamp comparison (second-resolution granularity means two saves within the same second can collide silently). Contentful uses `sys.version` similarly but it's proprietary. Payload and Sanity have no built-in conflict detection.

**Snapshot-based versioning.** Avoids the tree duplication problem that row-based approaches (Strapi, Payload) hit. ModulaCMS content uses sibling-pointer trees -- duplicating the tree for draft/published copies would double the complexity of every tree operation. Snapshots sidestep this entirely. The snapshot feeds through the same `BuildTree()` pipeline, so all six output formats work unchanged.

**Version restore safety.** Restore only applies field values, never tree structure. Sibling-pointer chains are too fragile to reconstruct from a snapshot (nodes may have been added/deleted/reordered since). This is safer than Strapi/Contentful which restore the full document state and can create orphaned references.

**Configurable retention.** WordPress has `wp_revisions_to_keep` but it's a global constant. ModulaCMS's per-content cap with smart pruning (never deletes the published snapshot or user-labeled versions) is more nuanced.

### Where competitors lead

**Strapi Enterprise** has review workflow stages (To do, In progress, Ready to review, Reviewed) integrated with publishing. ModulaCMS has permission-gated publishing only -- review workflows are Phase 4.

**Contentful** has webhooks deeply integrated with publishing events. ModulaCMS has no webhooks yet -- Phase 3.

**Sanity** has real-time collaborative editing with presence indicators. ModulaCMS uses traditional save/update -- real-time collaboration is out of scope (separate plan).

---

## Internationalization (Phase 2 -- DONE)

| Feature | ModulaCMS | Strapi v5 | Contentful | Directus | Sanity | WordPress |
|---------|-----------|-----------|------------|----------|--------|-----------|
| Approach | Locale column on fields | Row per locale | JSON locale maps | Junction table | Document per locale | Plugin (WPML/Polylang) |
| Shared structure | Tree shared, fields per locale | Separate rows | Fields contain all locales | Separate translation rows | Separate documents | Separate posts |
| Per-field translatable flag | Yes | Yes | Yes | Yes | Custom | Plugin-dependent |
| Fallback chains | 2-hop max, cycle-validated | Fallback locale | Fallback locale | Fallback locale | None built-in | Plugin-dependent |
| Per-locale publishing | Independent snapshots | Independent | Independent | N/A | Independent | Plugin-dependent |
| `locale=*` all-locales response | Yes | No | Yes | No | No | No |
| Non-translatable field handling | Single row, `locale=''`, auto-included in all snapshots | Duplicated per locale | Single field, all locales see it | Duplicated per locale | Per-document | Plugin-dependent |

### Where ModulaCMS leads

**Shared tree, per-field locale.** Strapi v5 duplicates the entire content row per locale -- for a site with 5 locales, that's 5 copies of every content entry with 5x the pointer-maintenance complexity for tree operations. ModulaCMS shares the tree structure across all locales and only varies field values. This is architecturally cleaner and avoids the "reorder blocks in one locale, must repeat in all others" problem.

**Non-translatable fields.** Fields like price, boolean toggles, and media references have a single row with `locale=''` that is automatically included in every locale's published snapshot. Strapi and Directus duplicate these values per locale, requiring sync mechanisms when the source value changes.

**`locale=*` multi-locale response.** Returns all published locales in one API call -- critical for SSG/ISR builds that need to generate all language variants. Only Contentful offers this among the competitors surveyed. Strapi, Sanity, and Directus require N separate requests.

**Fallback chain validation.** Chains are capped at 2 hops and validated on save to prevent cycles. Most competitors allow arbitrary fallback depth (harder to debug) or don't validate chains at all.

### Where competitors lead

**Contentful** stores all locale values in a single JSON field per entry, making cross-locale comparison trivial (everything is in one object). ModulaCMS requires joining across content_field rows.

**Sanity** treats each locale as a separate document, which makes per-locale permissions natural (different teams own different locales). ModulaCMS shares the tree, so locale-specific access control would need a separate permission layer.

**WordPress + WPML** has translation management workflows (assign translators, track translation status, integrate with translation services). ModulaCMS has no translation workflow -- just the data layer.

---

## Planned Features (Phases 3-6) vs Competitors

| Feature | ModulaCMS Status | Strapi v5 | Contentful | Payload | Sanity | DatoCMS |
|---------|-----------------|-----------|------------|---------|--------|---------|
| Webhooks | Phase 3 (planned) | Built-in | Built-in | Built-in | Built-in (GROQ-powered) | Built-in |
| Review workflows | Phase 4 (planned) | Enterprise (4-stage) | Roles only | Access control | Custom | Built-in (2-stage) |
| Preview tokens (shareable) | Phase 5 (planned) | Preview button | Preview API + tokens | Preview URL | `sanity.io/preview` | Preview links |
| Bulk publish | Phase 6 (planned) | Bulk actions | Bulk actions API | No | No | Bulk publish |
| Dependency warnings | Phase 6 (planned) | No | Reference validation | No | Reference checks | Built-in |
| HMAC-signed webhook payloads | Phase 3 (planned) | Yes | Yes (shared secret) | No | Yes | Yes |
| Webhook retry with backoff | Phase 3 (planned) | No built-in retry | Auto-retry | No | Auto-retry | Auto-retry |

### Phase 3 (Webhooks) will close the biggest gap

Every headless CMS ships webhooks. Without them, published content changes can't trigger CDN invalidation, SSG rebuilds, or external notifications. ModulaCMS's planned implementation (HMAC-SHA256 signing, exponential backoff retries, delivery audit log, manual retry) will match or exceed most competitors. Strapi notably lacks built-in retry logic.

### Phase 4 (Review) takes the pragmatic approach

ModulaCMS plans a single optional approval gate per datatype, not Strapi's 4-stage workflow. This matches DatoCMS's approach and avoids the overengineering that teams consistently regret. Users with `content:publish` can bypass the gate entirely -- a feature that Strapi Enterprise lacks (all content must go through all stages).

### Phase 5 (Preview tokens) matches Contentful's model

Database-backed tokens with revocation, locale scoping, and configurable expiry. Most competitors use session-based preview (requires CMS account) or signed URLs (no revocation). The token-per-locale scoping is unique.

### Phase 6 (Bulk ops + dependency tracking) matches DatoCMS

DatoCMS is the only major competitor with built-in dependency warnings on unpublish. ModulaCMS's query-based resolver (scanning relation/content_tree_ref field values rather than maintaining a junction table) is a lighter-weight approach that avoids data duplication.

---

## Overall Position Summary

### ModulaCMS strengths (unique or best-in-class)

1. **Tri-database support** (SQLite/MySQL/PostgreSQL) -- no competitor offers this
2. **Snapshot-based publishing** that avoids tree duplication -- architecturally sound for sibling-pointer trees
3. **Optimistic locking** with revision counter -- more reliable than timestamp-based approaches
4. **Shared tree i18n** -- cleaner than row-per-locale duplication
5. **Six output format transformers** -- serve content in Contentful/Sanity/Strapi/WordPress/clean/raw format from a single API
6. **`locale=*`** all-locales-in-one-response for SSG/ISR
7. **Single binary deployment** -- no competitor matches the operational simplicity
8. **TUI over SSH** -- unique management interface
9. **Admin panel is also headless** -- dual content model

### ModulaCMS gaps (vs enterprise tiers)

| Gap | Impact | Plan |
|-----|--------|------|
| No webhooks | Can't notify external systems on publish | Phase 3 |
| No review workflows | Large teams can't gate content | Phase 4 |
| No preview sharing | Can't share drafts with non-CMS users | Phase 5 |
| No bulk operations | Tedious to publish many items at once | Phase 6 |
| No real-time collaboration | Can't co-edit like Sanity/Notion | Separate plan |
| No translation workflows | Can't assign/track translator work | Not planned |
| No visual page builder | Content teams need technical skills | Not planned (by design) |
| No managed hosting | Self-hosted only | Future consideration |

### Competitive tier placement

**Open-source tier:** ModulaCMS now matches Strapi v5 open-source and Directus in publishing/versioning/i18n capabilities. Exceeds both in optimistic locking, version retention policy, and i18n architecture.

**Enterprise tier:** After Phases 3-4 ship (webhooks + review), ModulaCMS will match Strapi Enterprise's publishing feature set. The remaining enterprise-only gaps (real-time collaboration, translation management, managed hosting) are architectural choices, not missing features.

**Unique position:** No other headless CMS combines tri-database support, snapshot-based publishing, shared-tree i18n, six output format transformers, SSH TUI, and single-binary deployment. The closest architectural peer is Directus (SQL-based, self-hosted), but Directus lacks version history, optimistic locking, and the output format system.
