# MCP Tool Naming Audit

Audit of all 170 ModulaCMS MCP server tools for naming problems that cause LLMs to select the wrong tool, pass wrong parameters, or misinterpret results.

**Source files audited:** Every `tools_*.go` file in `internal/mcp/`.

**Methodology:** Read every tool registration (name, description, parameters, handler) and identified cases where the name alone (what an LLM reads first) would lead to a different action than what the handler implements.

---

## Category 1: Name-identical patterns that behave differently

### 1.1 `get_route` (ID-based) vs `admin_get_route` (slug-based)

**Files:** `tools_routes.go:19-23`, `tools_admin_routes.go:13`

`get_route` takes an `id` parameter (ULID). `admin_get_route` takes a `slug` parameter. The `admin_` prefix does not signal a lookup-mechanism change. An LLM that successfully uses `get_route(id=X)` will call `admin_get_route(id=X)` and fail.

**Fix:** Rename to `admin_get_route_by_slug`. Or change the admin API to also accept ID lookups for consistency.

### 1.2 `update_*` full-replace vs partial-merge inconsistency

**Across all `tools_*.go` files**

Some `update_*` tools do full replacement (omitted fields set to null/empty), others do partial merge (omitted fields unchanged). Both use the identical `update_` naming pattern. An LLM that learns `update_media` is safe to call with partial fields will then call `update_content` with partial fields and null out every omitted pointer field.

Full replace (dangerous with omitted fields):
- `update_content` -- "full replacement -- all fields are sent. Omitted pointer fields will be set to null"
- `update_datatype` -- "full replacement"
- `update_field` -- "full replacement"
- `update_user` -- "full replacement -- provide all fields you want to keep"
- `update_field_type`
- `admin_update_content`, `admin_update_datatype`, `admin_update_field`

Partial merge (safe with omitted fields):
- `update_media` -- "Only provided fields are changed; omitted fields remain unchanged"
- `update_media_folder` -- "Only provided fields are changed"
- `admin_update_media`, `admin_update_media_folder`

Unclear/undocumented (description does not state which strategy):
- `update_route`, `admin_update_route`
- `update_session`
- `update_role`, `update_permission`
- `update_table`
- `update_media_dimension`

**Fix options:**
- A) Rename full-replace tools to `replace_*` (e.g. `replace_content`, `replace_datatype`). Keep `update_*` for partial merge.
- B) Add `[FULL REPLACE]` or `[PARTIAL]` prefix to every update tool description.
- C) Standardize all updates to partial merge behavior.

### 1.3 `create_content` author_id NOT auto-populated vs `batch_update_content` author_id IS auto-populated

**Files:** `tools_content.go:33-41`, `tools_content.go:184-200`

`create_content` description warns: "author_id is NOT auto-populated; use the whoami tool to get your user ID." But `batch_update_content` says "author_id is auto-populated from the API token on the batch endpoint." Same entity, opposite behavior on the same field. An LLM that learns to omit author_id from batch operations will also omit it from create operations, producing authorless content.

**Fix:** Auto-populate author_id on create_content from the API token, matching batch behavior. Or add author_id auto-population to all mutation tools consistently.

### 1.4 `create_field` enum `_id` vs `admin_create_field` enum `relation`

**Files:** `tools_schema.go:84`, `tools_admin_schema.go:21`

Public `create_field` field_type enum: `text, textarea, number, date, datetime, boolean, select, media, _id, json, richtext, slug, email, url`.
Admin `admin_create_field` field_type enum: `text, textarea, number, date, datetime, boolean, select, media, relation, json, richtext, slug, email, url`.

`_id` exists only in public. `relation` exists only in admin. An LLM cannot know this without inspecting the enum. If it tries `create_field(field_type="relation")` it gets an opaque enum validation error.

**Fix:** Document the difference in descriptions. Or unify the enums (support both `_id` and `relation` in both, mapping as needed).

---

## Category 2: Names that mislead about what the tool returns

### 2.1 `list_content` returns metadata, not content

**File:** `tools_content.go:15-22`

Description says "Returns structural metadata (IDs, status, timestamps, tree pointers) without field values." A user saying "show me my blog posts" triggers `list_content`. The LLM gets IDs and timestamps with zero text. Getting actual content requires: `list_routes` -> find slug -> `get_page(slug)` per item. Three tools for what the name promises in one.

**Fix:** Rename to `list_content_metadata` or `list_content_entries`. Or add a `fields=true` parameter that inlines field values (like `list_datatypes` has `full=true`).

### 2.2 `list_content_fields` returns ALL fields globally, not filterable by content item

**File:** `tools_content.go:88-95`

Description says "Returns ALL content fields across all content items (cannot filter by content item)." An LLM trying to get fields for a specific blog post calls `list_content_fields` expecting a `content_data_id` filter. No such filter exists. It gets a paginated dump of every field value in the system.

**Fix:** Add a `content_data_id` filter parameter. Or rename to `list_all_content_fields` to signal the global scope.

### 2.3 `get_content_tree` is content delivery, not tree structure

**File:** `tools_content.go:77-84`

Takes a `slug` and returns assembled content with field values in a format (contentful/sanity/etc.). An LLM needing the tree structure (parent/child/sibling pointers) calls this expecting structural data. Gets formatted content delivery output instead. The name says "tree" but the behavior is "assembled page."

**Fix:** Rename to `get_content_by_slug` or `get_assembled_content`. Or create a separate tool that actually returns tree structure (node IDs, parent/child/sibling pointers).

### 2.4 `list_tables` returns CMS metadata records, not database tables

**File:** `tools_tables.go:11-16`

Description: "List all CMS metadata tables." These are ULID-identified label records, not actual database tables. An LLM debugging schema issues calls `list_tables` expecting column definitions and gets opaque metadata.

**Fix:** Rename to `list_table_records` or `list_cms_tables`. Or add a description that explicitly says "These are NOT database tables. They are CMS-internal metadata records."

### 2.5 `get_datatype_full` doubles as a list tool

**File:** `tools_schema.go:120-126`

Description: "Get a single datatype with its linked fields joined. If id is omitted, returns all datatypes with fields." A tool named `get_*` that returns a list when you omit a parameter violates the get/list convention. An LLM calling `get_datatype_full()` without an ID expects one record and gets an array.

**Fix:** Split into `get_datatype_full(id)` (required ID, returns one) and `list_datatypes_full()` (returns all). Or make the existing `list_datatypes(full=true)` the canonical way to list all with fields and require `id` on `get_datatype_full`.

### 2.6 `list_media` description contradicts `upload_media` existence

**File:** `tools_media.go:15-21`

`list_media` description says "This MCP server can view and update media metadata but cannot upload new files." But `upload_media` exists on line 56 and works. The description is stale.

**Fix:** Remove the incorrect sentence from `list_media` description.

---

## Category 3: Dangerous name/behavior mismatches

### 3.1 `media_cleanup` permanently deletes without dry_run or confirmation

**File:** `tools_media.go:75-79`

Named like a maintenance/optimization tool. Actually "Removes media records without backing files." No dry_run parameter, no confirmation flag. Compare with the safety patterns used elsewhere:
- `heal_content` has a `dry_run` parameter
- `plugin_cleanup_dry_run` is a separate read-only tool
- `plugin_cleanup_drop` requires both `confirm: true` AND an explicit table list

An LLM exploring/maintaining the CMS calls `media_cleanup` thinking it's safe.

**Fix:** Split into `media_cleanup_dry_run` and `media_cleanup_apply` (matching plugin pattern). Or add `dry_run` parameter defaulting to true. Or add `confirm: true` required parameter.

### 3.2 `deploy_import` vs `import_content` vs `import_bulk` -- three "import" tools

**Files:** `tools_deploy.go:27-32`, `tools_import.go:12-18`, `tools_import.go:20-27`

- `import_content`: Converts external CMS data (Contentful, Sanity, etc.) into ModulaCMS entities.
- `import_bulk`: Same parameters as `import_content` (format + data). Description says "Bulk import data. Accepts a raw JSON object that is posted directly to the bulk import endpoint." Indistinguishable from `import_content` based on name and params.
- `deploy_import`: Applies a sync payload between ModulaCMS environments.

A user saying "import my Contentful data" could trigger any of these. `import_content` and `import_bulk` appear identical. The `deploy_` prefix is insufficient disambiguation.

**Fix:**
- Rename deploy tools: `sync_export`, `sync_import`, `sync_dry_run`, `sync_health`.
- Clarify `import_content` vs `import_bulk`: if they differ, explain how in the descriptions. If they don't, remove one.

---

## Category 4: Parameter-level confusion

### 4.1 `admin_` parameter prefix inconsistency

**Files:** `tools_admin_content.go`, `tools_admin_schema.go`, `tools_admin_routes.go`

Some admin tools prefix parameters with `admin_`, others don't. The pattern is inconsistent even within the same tool:

| Tool | Prefixed params | Unprefixed params |
|------|----------------|-------------------|
| `admin_create_content` | `admin_route_id`, `admin_datatype_id` | `parent_id`, `author_id`, `status` |
| `admin_create_content_field` | `admin_content_data_id`, `admin_field_id`, `admin_field_value`, `admin_route_id` | `author_id` |
| `admin_create_datatype` | (none) | `name`, `label`, `type`, `parent_id`, `author_id` |
| `admin_create_field` | (none) | `name`, `label`, `field_type`, `parent_id`, etc. |
| `admin_create_route` | (none) | `slug`, `title`, `status`, `author_id` |

The implicit rule: prefix when the referenced entity is admin-specific (admin_route_id -> admin route). Don't prefix shared entities (author_id -> user). This is invisible to an LLM. It will guess wrong.

**Fix:** Drop `admin_` prefix from all parameters. The tool name already signals the domain. Use `content_data_id`, `field_id`, `field_value` in both public and admin tools.

### 4.2 `field_type` parameter vs `field_types` entity

**Files:** `tools_schema.go:84`, `tools_schema.go:130-170`

`create_field` has a parameter `field_type` that accepts a hardcoded enum (text, number, etc.). Separately, `field_types` is a user-extensible entity with CRUD tools (`create_field_type`, `list_field_types`). An LLM creates `create_field_type(type="color", label="Color Picker")` then tries `create_field(field_type="color")` and gets an enum validation error. The hardcoded enum is not driven by the field_types table.

**Fix:** Either remove the hardcoded enum on `create_field` and validate against the field_types table at runtime. Or document in `create_field_type` description that creating a field type does not automatically make it available as a field_type parameter value.

### 4.3 Route `status` (number) vs content `status` (string enum)

**Files:** `tools_routes.go:31`, `tools_content.go:35`

`create_route` takes `status` as a number ("Route status (positive integer)"). `create_content` takes `status` as a string enum ("draft, published, archived, pending"). Both parameters are called `status` on closely related entities.

**Fix:** Rename route status to `route_status` or `status_code`. Or convert to a string enum (e.g. "active", "inactive") with server-side mapping.

### 4.4 `remove_role_permission` takes junction ID, not role + permission

**File:** `tools_rbac.go:34-39`

The name suggests you'd pass a role and permission to unlink them. Actually takes `id` (the role_permission junction record ULID). Contrast with the inverse: `assign_role_permission` takes `role_id` + `permission_id` directly. An LLM that just assigned a permission has those two IDs in context but not the junction ID.

**Fix:** Accept `role_id` + `permission_id` as an alternative to `id`. Or rename to `delete_role_permission` with `id` to make clear it's a record deletion.

---

## Category 5: Confusable tool pairs

### 5.1 `list_plugin_routes` vs `list_routes`

**Files:** `tools_plugins.go:68-72`, `tools_routes.go:11-15`

`list_routes` returns CMS content routes (URL slugs mapped to content). `list_plugin_routes` returns HTTP routes registered by Lua plugins with approval status. Both are "routes." An LLM managing URL routing calls the wrong one.

**Fix:** Rename `list_plugin_routes` to `list_plugin_http_endpoints` or `list_plugin_api_routes`.

### 5.2 `health` vs `media_health` vs `deploy_health`

**Files:** `tools_health.go`, `tools_media.go:67-72`, `tools_deploy.go:12-16`

Three health tools with no naming convention showing scope. An LLM asked "is the CMS healthy?" may call any of them. Only `health` gives the overall picture.

**Fix:** Rename to `server_health`, `media_storage_health`, `deploy_sync_health` for unambiguous scope.

### 5.3 `save_content_tree` vs `batch_update_content`

**Files:** `tools_content.go:163-168`, `tools_content.go:183-200`

Both do atomic multi-entity writes on content. `save_content_tree` handles tree structure (creates, deletes, pointer updates). `batch_update_content` handles field value changes on one item. An LLM updating a post with new field values AND moving it under a new parent has no clear single tool. Both names sound like "save my changes."

**Fix:** Rename `save_content_tree` to `apply_tree_structure_changes` or `save_tree_operations`. Makes it clear this is for structural tree edits, not content value edits.

### 5.4 `admin_list_media_dimensions` returns global data despite `admin_` prefix

**File:** `tools_admin_media.go:65-73`

Source code comment on line 65-66: "admin_list_media_dimensions reuses the public media dimensions -- they are shared." The `admin_` prefix implies admin-scoped data. An LLM calls both `list_media_dimensions` and `admin_list_media_dimensions` expecting different results, gets identical data.

**Fix:** Remove `admin_list_media_dimensions`. Document that `list_media_dimensions` applies to both content systems.

---

## Category 6: Description-level issues

### 6.1 Lowercase description starts on update tools

All `update_*` descriptions start with lowercase "update" instead of "Update":
- "update an existing content data entry"
- "update media asset metadata"
- "update a route by ID"

Every other verb (List, Get, Create, Delete, Check, Run, Scan, etc.) is capitalized. Minor, but inconsistent.

**Fix:** Capitalize all description starts.

### 6.2 Inconsistent maintenance verb vocabulary

| Tool | Verb | Safety mechanism |
|------|------|-----------------|
| `heal_content` | heal | `dry_run` parameter (defaults false) |
| `media_cleanup` | cleanup | none |
| `media_health` | health | read-only |
| `plugin_cleanup_dry_run` | cleanup | separate tool (read-only) |
| `plugin_cleanup_drop` | cleanup | `confirm: true` + explicit table list |
| `health` | health | read-only |
| `deploy_health` | health | read-only |

Three verbs (heal, cleanup, health) with three different safety patterns. "Cleanup" is destructive in media but read-only+destructive split in plugins.

**Fix:** Standardize on the plugin pattern: separate `_check` (read-only) and `_apply` (destructive with confirmation) tools for every maintenance operation.

---

## Priority order for fixes

1. **`list_media` stale description** (#2.6) -- factually wrong, one-line fix
2. **Lowercase description starts** (#6.1) -- trivial consistency fix
3. **`media_cleanup` safety** (#3.1) -- data loss risk
4. **`update_*` replace vs merge clarity** (#1.2) -- data loss risk
5. **`deploy_import`/`import_content`/`import_bulk` disambiguation** (#3.2) -- wrong tool selection
6. **`admin_get_route` slug vs ID** (#1.1) -- wrong parameter
7. **`list_content` returns no content** (#2.1) -- wrong expectations
8. **`list_content_fields` not filterable** (#2.2) -- wrong expectations
9. **`admin_` parameter prefix inconsistency** (#4.1) -- wrong parameters
10. **`get_datatype_full` list/get dual behavior** (#2.5) -- wrong expectations
11. **`field_type` param vs `field_types` entity** (#4.2) -- wrong expectations
12. **Route status number vs content status string** (#4.3) -- wrong type
13. **`remove_role_permission` junction ID** (#4.4) -- requires extra lookup
14. **`get_content_tree` name** (#2.3) -- misleading
15. **`list_tables` name** (#2.4) -- misleading
16. **`list_plugin_routes` vs `list_routes`** (#5.1) -- confusable
17. **Health tool scoping** (#5.2) -- confusable
18. **`save_content_tree` vs `batch_update_content`** (#5.3) -- confusable
19. **`admin_list_media_dimensions` duplicate** (#5.4) -- redundant
20. **`create_content` vs `batch_update_content` author_id** (#1.3) -- inconsistent
21. **`create_field` vs `admin_create_field` enum mismatch** (#1.4) -- silent failure
22. **Maintenance verb inconsistency** (#6.2) -- pattern debt
