# MCP vs API Gap Analysis

Systematic comparison of every REST API endpoint in `internal/router/mux.go` against the MCP tool surface in `internal/mcp/tools_*.go`. Identifies API capabilities with no MCP equivalent.

**Source:** `internal/router/mux.go` lines 1-730 (API routes only, excluding HTMX admin panel routes).

---

## Missing from MCP: Entire feature domains

### 1. Authentication (6 endpoints, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `POST /api/v1/auth/login` | Login with credentials |
| `POST /api/v1/auth/logout` | Logout / invalidate session |
| `GET /api/v1/auth/me` | Get current user profile |
| `POST /api/v1/auth/register` | Register new user |
| `POST /api/v1/auth/reset` | Reset password |
| `POST /api/v1/auth/request-password-reset` | Request password reset email |
| `POST /api/v1/auth/confirm-password-reset` | Confirm password reset with token |

The MCP server uses API key auth, so login/logout aren't needed. But `auth/register` and the password reset flow are user management capabilities the MCP cannot perform. The `whoami` MCP tool partially covers `auth/me`.

**Impact:** An LLM cannot register users through the public registration flow (which may have different behavior than `create_user`, e.g. auto-assigning the `viewer` role). Password reset flows are completely inaccessible.

### 2. OAuth flows (2 endpoints, 0 MCP tools for the flow itself)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/auth/oauth/login` | Initiate OAuth login (redirect) |
| `GET /api/v1/auth/oauth/callback` | OAuth callback handler |

The MCP has CRUD tools for OAuth connection records (`list_users_oauth`, etc.) but cannot initiate or complete OAuth flows. This is expected since OAuth requires browser redirects, but the gap means an LLM cannot link a user's OAuth account.

**Impact:** Low. OAuth flows are inherently interactive.

### 3. Content Publishing (6 endpoints, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `POST /api/v1/content/publish` | Publish content |
| `POST /api/v1/content/unpublish` | Unpublish content |
| `POST /api/v1/content/schedule` | Schedule content for future publication |
| `POST /api/v1/admin/content/publish` | Publish admin content |
| `POST /api/v1/admin/content/unpublish` | Unpublish admin content |
| `POST /api/v1/admin/content/schedule` | Schedule admin content publication |

**Impact:** Critical. An LLM can create and edit content but cannot publish it. The entire publish/unpublish/schedule lifecycle is missing. This means content created through MCP stays in draft state forever unless someone uses the API or admin panel to publish it. This is the single largest functional gap.

### 4. Content Versions (10 endpoints, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/contentversions` | List versions filtered by content_id |
| `GET /api/v1/content/versions` | List all versions |
| `GET /api/v1/content/versions/{id}` | Get a specific version |
| `POST /api/v1/content/versions` | Create a manual version snapshot |
| `DELETE /api/v1/content/versions/{id}` | Delete a version |
| `POST /api/v1/content/restore` | Restore content from a version |
| `GET /api/v1/admin/content/versions` | List admin content versions |
| `GET /api/v1/admin/content/versions/{id}` | Get a specific admin version |
| `POST /api/v1/admin/content/versions` | Create admin manual version |
| `DELETE /api/v1/admin/content/versions/{id}` | Delete admin version |
| `POST /api/v1/admin/content/restore` | Restore admin content from version |

**Impact:** High. No version history, no rollback capability. An LLM cannot inspect previous states of content, create snapshots before edits, or restore to a known good state if a batch operation goes wrong.

### 5. Locales (5 endpoints, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/locales` | List enabled locales (public, no auth) |
| `GET /api/v1/admin/locales` | List all locales (admin) |
| `POST /api/v1/admin/locales` | Create locale |
| `GET /api/v1/admin/locales/{id}` | Get locale |
| `PUT /api/v1/admin/locales/{id}` | Update locale |
| `DELETE /api/v1/admin/locales/{id}` | Delete locale |

**Impact:** High for multilingual sites. An LLM cannot manage locales, create translations, or query locale-specific content.

### 6. Content Translations (2 endpoints, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `POST /api/v1/admin/contentdata/{id}/translations` | Create translation of public content |
| `POST /api/v1/admin/admincontentdata/{id}/translations` | Create translation of admin content |

**Impact:** High for multilingual sites. Even if locales were exposed, the translation creation mechanism is missing.

### 7. Webhooks (8 endpoints, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/admin/webhooks` | List webhooks |
| `POST /api/v1/admin/webhooks` | Create webhook |
| `GET /api/v1/admin/webhooks/{id}` | Get webhook |
| `PUT /api/v1/admin/webhooks/{id}` | Update webhook |
| `DELETE /api/v1/admin/webhooks/{id}` | Delete webhook |
| `POST /api/v1/admin/webhooks/{id}/test` | Send test webhook |
| `GET /api/v1/admin/webhooks/{id}/deliveries` | List delivery history |
| `POST /api/v1/admin/webhooks/deliveries/{id}/retry` | Retry failed delivery |

**Impact:** Medium-high. An LLM cannot set up webhook integrations, test them, or debug delivery failures.

### 8. Validations (12 endpoints, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/validations` | List public validations |
| `POST /api/v1/validations` | Create public validation |
| `GET /api/v1/validations/search` | Search validations |
| `GET /api/v1/validations/{id}` | Get validation |
| `PUT /api/v1/validations/{id}` | Update validation |
| `DELETE /api/v1/validations/{id}` | Delete validation |
| `GET /api/v1/admin/validations` | List admin validations |
| `POST /api/v1/admin/validations` | Create admin validation |
| `GET /api/v1/admin/validations/search` | Search admin validations |
| `GET /api/v1/admin/validations/{id}` | Get admin validation |
| `PUT /api/v1/admin/validations/{id}` | Update admin validation |
| `DELETE /api/v1/admin/validations/{id}` | Delete admin validation |

**Impact:** Medium. An LLM cannot configure field validation rules. Content created via MCP may bypass validations that the admin panel enforces.

### 9. Search (2 endpoints, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/search` | Full-text search over published content (public, no auth) |
| `POST /api/v1/admin/search/rebuild` | Rebuild search index |

**Impact:** Medium. An LLM cannot search published content by keyword or rebuild the search index after bulk operations.

### 10. Content Query by Datatype (1 endpoint, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/query/{datatype}` | Query published content filtered by datatype with sorting/pagination |

**Impact:** Medium-high. This is the primary way frontends fetch content collections (e.g. "all blog posts"). An LLM cannot test content queries or verify what the frontend would see.

### 11. Globals (1 endpoint, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/globals` | Get global content (site name, footer, etc.) |

**Impact:** Low-medium. An LLM cannot read or verify global content delivery.

### 12. Metrics (1 endpoint, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/admin/metrics` | Get runtime metrics snapshot |

**Impact:** Low-medium. An LLM cannot monitor server performance or diagnose issues via metrics.

### 13. Activity Feed (1 endpoint, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/activity/recent` | Get recent audit activity |

**Impact:** Medium. An LLM cannot review the audit trail to understand what changed and when. Useful for debugging content issues.

### 14. Environment (1 endpoint, 0 MCP tools)

| API Endpoint | Purpose |
|-------------|---------|
| `GET /api/v1/environment` | Get environment info (public, no auth) |

**Impact:** Low. Could help an LLM understand which environment it's connected to.

---

## Missing from MCP: Individual capabilities within covered domains

### 15. Content: composite create and "full" views

| API Endpoint | MCP Gap |
|-------------|---------|
| `POST /api/v1/content/create` | Composite content create (content + fields in one request). MCP has `create_content` + `create_content_field` as separate calls. |
| `GET /api/v1/contentdata/full` | Content data with joined metadata. MCP `list_content` returns structural metadata only. |
| `GET /api/v1/contentdata/by-route` | List content data filtered by route ID. MCP has no equivalent filter. |
| `GET /api/v1/admincontentdatas/full` | Admin content data with joined metadata. |

**Impact:** Medium. The composite create means an LLM must make N+1 calls (1 create_content + N create_content_field) instead of one. The "full" views would give richer list data without per-item fetches.

### 16. Content tree by route ID

| API Endpoint | MCP Gap |
|-------------|---------|
| `GET /api/v1/content/tree/{routeID}` | Get content tree by route ID. MCP `get_content_tree` takes a slug, not a route ID. |

**Impact:** Low. MCP can get content tree by slug. Route-ID-based lookup is a convenience.

### 17. Datatype sort ordering

| API Endpoint | MCP Gap |
|-------------|---------|
| `GET /api/v1/datatype/max-sort-order` | Get max sort order for datatypes |
| `PUT /api/v1/datatype/{id}/sort-order` | Update datatype sort order |
| `GET /api/v1/datatype/full/list` | List all datatypes full (separate endpoint from single get) |
| `GET /api/v1/admindatatypes/max-sort-order` | Admin datatype max sort order |
| `PUT /api/v1/admindatatypes/{id}/sort-order` | Admin datatype sort order update |
| `GET /api/v1/admindatatypes/full` | Admin datatypes full list |

**Impact:** Medium. An LLM cannot reorder datatypes in the admin panel. The max-sort-order + sort-order update pattern is how the TUI and admin panel manage display ordering.

### 18. Field sort ordering

| API Endpoint | MCP Gap |
|-------------|---------|
| `PUT /api/v1/fields/{id}/sort-order` | Update field sort order |
| `GET /api/v1/fields/max-sort-order` | Get max field sort order |

**Impact:** Medium. An LLM cannot control the display order of fields within a datatype.

### 19. Media: download, references, full view, reprocess

| API Endpoint | MCP Gap |
|-------------|---------|
| `GET /api/v1/media/{id}/download` | Download media file binary |
| `GET /api/v1/media/full` | Media with full joined metadata |
| `GET /api/v1/media/references` | Find content referencing a media asset |
| `GET /api/v1/media/reprocess/status` | Check media reprocessing status |
| `POST /api/v1/media/reprocess` | Trigger media reprocessing (regenerate sizes) |
| `GET /api/v1/adminmedia/{id}/download` | Download admin media file binary |

**Impact:** Medium-high. `media/references` is important for safe deletion (know what breaks). `media/reprocess` is needed after changing dimension presets. Download is needed for LLMs that want to inspect media files.

### 20. Media folder: tree view and folder contents

| API Endpoint | MCP Gap |
|-------------|---------|
| `GET /api/v1/media-folders/tree` | Get full folder tree hierarchy |
| `GET /api/v1/media-folders/{id}/media` | List media within a specific folder |
| `GET /api/v1/adminmedia-folders/tree` | Admin folder tree |
| `GET /api/v1/adminmedia-folders/{id}/media` | Admin folder media list |

**Impact:** Medium. MCP `list_media_folders` only returns root or children of a parent. No way to get the full tree in one call. No way to list media filtered to a specific folder.

### 21. Routes: full view

| API Endpoint | MCP Gap |
|-------------|---------|
| `GET /api/v1/routes/full` | Routes with full joined metadata |

**Impact:** Low. MCP `list_routes` returns route data, "full" adds joined content/metadata.

### 22. User: reassign-delete and sessions

| API Endpoint | MCP Gap |
|-------------|---------|
| `POST /api/v1/users/reassign-delete` | Reassign content to another user then delete the user |
| `GET /api/v1/users/sessions` | List sessions for a specific user |

**Impact:** Medium. `reassign-delete` is the safe way to remove users without orphaning content. MCP `delete_user` may leave orphaned content. `users/sessions` shows which users are currently active.

### 23. Config: search index

| API Endpoint | MCP Gap |
|-------------|---------|
| `GET /api/v1/admin/config/search-index` | Get config field search index for command palette |

**Impact:** Low. This is a UI convenience for the admin panel command palette.

### 24. Deploy sync: dry run

| API Endpoint | MCP Status |
|-------------|-----------|
| `POST /api/v1/deploy/import` | MCP has `deploy_import` (covered) |
| `POST /api/v1/deploy/export` | MCP has `deploy_export` (covered) |

Note: The API has no explicit dry-run endpoint. MCP `deploy_dry_run` appears to be MCP-only functionality (the SDK may implement it client-side or the endpoint may be undocumented). Needs verification.

---

## Summary: Gap severity

### Critical (blocks core workflows)
1. **Publishing** -- content lifecycle is incomplete without publish/unpublish/schedule
2. **Content versions** -- no rollback capability after destructive edits

### High (significant capability loss)
3. **Locales + translations** -- multilingual sites cannot be managed
4. **Webhooks** -- integration setup/debugging impossible
5. **Content query by datatype** -- cannot verify frontend content delivery
6. **Media references** -- cannot safely audit media before deletion

### Medium (functionality gaps)
7. **Validations** -- field validation rules unmanageable
8. **Search** -- cannot query published content or rebuild index
9. **Activity feed** -- no audit trail access
10. **Datatype/field sort ordering** -- cannot control display order
11. **Composite content create** -- N+1 calls instead of one
12. **Content data "full" views** -- missing rich list endpoints
13. **Media folder tree + folder contents** -- partial folder navigation
14. **User reassign-delete** -- unsafe user deletion
15. **Media reprocess** -- cannot regenerate image sizes
16. **Metrics** -- no performance monitoring

### Low (minor gaps)
17. **Environment info** -- cosmetic
18. **Globals delivery** -- read-only convenience
19. **Config search index** -- UI-specific
20. **Routes full view** -- convenience
21. **OAuth flows** -- inherently interactive

---

## Endpoint count

| Category | API endpoints | MCP tools | Coverage |
|----------|--------------|-----------|----------|
| Auth | 7 | 0 | 0% |
| OAuth flows | 2 | 0 | 0% |
| OAuth CRUD | 5 | 5 | 100% |
| Health | 1 | 1 | 100% |
| Environment | 1 | 0 | 0% |
| Content data CRUD | 5 | 5 | 100% |
| Content data extras | 3 | 0 | 0% |
| Content fields CRUD | 4 | 5 | 100%+ |
| Content composite | 2 | 2 | 100% |
| Content tree | 2 | 2 | 100% |
| Content heal | 1 | 1 | 100% |
| Content reorder/move | 2 | 2 | 100% |
| Content publishing | 6 | 0 | 0% |
| Content versions | 11 | 0 | 0% |
| Content delivery (slug) | 1 | 1 | 100% |
| Content query | 1 | 0 | 0% |
| Globals | 1 | 0 | 0% |
| Admin content CRUD | 5 | 5 | 100% |
| Admin content extras | 1 | 0 | 0% |
| Admin content fields | 4 | 5 | 100%+ |
| Admin content reorder/move | 2 | 2 | 100% |
| Admin content publishing | 3 | 0 | 0% |
| Admin content versions | 5 | 0 | 0% |
| Admin tree | 1 | 0 | 0% |
| Datatypes CRUD | 3 | 5 | 100%+ |
| Datatypes extras | 4 | 1 | 25% |
| Admin datatypes | 5 | 5 | 100% |
| Admin datatypes extras | 3 | 0 | 0% |
| Fields CRUD | 3 | 5 | 100%+ |
| Fields extras | 2 | 0 | 0% |
| Admin fields | 3 | 5 | 100%+ |
| Field types | 4 | 5 | 100%+ |
| Admin field types | 4 | 5 | 100%+ |
| Routes | 3 | 5 | 100%+ |
| Routes extras | 1 | 0 | 0% |
| Admin routes | 3 | 5 | 100%+ |
| Media CRUD | 4 | 5 | 100%+ |
| Media extras | 7 | 2 | 29% |
| Media folders | 7 | 6 | 86% |
| Media dimensions | 4 | 5 | 100%+ |
| Admin media | 4 | 5 | 100%+ |
| Admin media extras | 1 | 0 | 0% |
| Admin media folders | 7 | 6 | 86% |
| Roles CRUD | 4 | 5 | 100%+ |
| Permissions CRUD | 4 | 5 | 100%+ |
| Role-permissions | 3 | 5 | 100%+ |
| Sessions | 4 | 4 | 100% |
| Tables | 4 | 5 | 100%+ |
| Tokens | 4 | 4 | 100% |
| Users CRUD | 5 | 6 | 100%+ |
| Users extras | 2 | 2 | 100% |
| User management extras | 2 | 0 | 0% |
| SSH keys | 3 | 3 | 100% |
| Import | 6 | 2 | 33% |
| Deploy sync | 3 | 4 | 100%+ |
| Config | 4 | 3 | 75% |
| Locales | 6 | 0 | 0% |
| Translations | 2 | 0 | 0% |
| Webhooks | 8 | 0 | 0% |
| Validations | 12 | 0 | 0% |
| Search | 2 | 0 | 0% |
| Activity | 1 | 0 | 0% |
| Metrics | 1 | 0 | 0% |
| Plugins | 3 | 12 | 100%+ |

**Total API endpoints (non-admin-panel):** ~198
**MCP tools:** 170
**API endpoints with no MCP equivalent:** ~68 (34%)
