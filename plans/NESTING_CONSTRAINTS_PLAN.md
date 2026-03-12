# Nesting Constraints Plan

**Status:** Draft
**Created:** 2026-03-10
**Scope:** Add optional `allowed_child_types` column to datatypes, enforce at content create/move time

---

## Design Decision

**Column:** `allowed_child_types TEXT DEFAULT NULL` on both `datatypes` and `admin_datatypes` tables.

**Semantics:**
- `NULL` — No restriction. Current behavior preserved (hierarchy-based filtering via `filterChildDatatypes()` / `fetchDatatypesGrouped()`). This is the default for all existing datatypes.
- `[]` (empty JSON array) — Leaf-only datatype. No children allowed.
- `["row", "column", "hero"]` — Only datatypes with these **names** may be children.

**Why names, not IDs:**
1. The `name` column has a `UNIQUE INDEX` on `datatypes` (machine-readable identifier).
2. `DatatypeDef` presets reference datatypes by name, not ID — names are stable across environments.
3. Import/export works without ID remapping.
4. Human-readable in API responses and config.
5. IDs are environment-specific ULIDs that change between installs.

**Why not `type`:**
The `type` column is a category (`_root`, `_global`, `_collection`, `page`, `post`), not an identifier. Multiple datatypes can share the same type string.

**Scope limit:** This plan adds only `allowed_child_types` (parent → allowed children). The inverse (`allowed_parent_types` — child → allowed parents) can be added later if needed. One direction is sufficient for v1.

**Interaction with existing hierarchy:**
- The existing `parent_id` on datatypes defines **schema hierarchy** (how datatypes are organized/categorized).
- `allowed_child_types` defines **content nesting rules** (which datatypes can be children in the content tree).
- These are independent concerns. Schema hierarchy determines what appears in picker categories. `allowed_child_types` is the authoritative constraint on what is actually allowed.
- When `allowed_child_types` is set, picker filtering intersects with it (only show types that are both in the hierarchy AND in the allowed list).

---

## Blast Radius

| System | Files Affected | Nature of Change |
|--------|---------------|-----------------|
| SQL schema (3 backends × 2 tables) | 12 | ALTER TABLE + query updates |
| sqlc generated code | 3 dirs | Regenerated (never hand-edit) |
| Go wrapper types + mappers | 4 | New field + mapper lines |
| Audited command types | 4 | New field in params |
| DbDriver interface | 0 | No new methods needed |
| ~~Migration (ensure.go)~~ | 0 | Skipped — greenfield, no existing databases |
| Validation logic | 1 new | `internal/service/nesting.go` |
| ContentService | 2 | Inject validation call |
| Tree ops (move.go) | 0 | Validation at service layer, not ops |
| Router handlers (datatypes) | 2 | Parse new field on create/update |
| Router handlers (content) | 0 | Service returns errors, handlers propagate |
| Admin panel handlers | 1 | Parse allowed_child_types from form |
| Admin panel templ pages | 2 | Add UI for configuring constraint |
| Admin panel templ partials | 2 | Form fields for constraint |
| Block editor JS (cache.js) | 1 | Filter by allowed_child_types |
| TUI screens (datatypes) | 1 | Edit dialog for constraint |
| TUI filtering (update_fetch.go) | 1 | Intersect with allowed_child_types |
| TUI dialogs | 1 | Display constraint in datatype form |
| Definitions (definition.go) | 1 | New field on DatatypeDef |
| Definition presets (def_*.go) | 5 | Add constraints to presets |
| Definition install | 1 | Persist new field |
| Deploy export/import | 0 | Auto-handled via struct JSON tags |
| SDKs (Go, TypeScript, Swift) | 5 | New field on Datatype types |
| MCP tools | 0 | Auto-exposed via struct |
| Tests | ~10 | New + updated |
| Documentation | ~3 | CLAUDE.md, user docs |

**Total: ~60 files across 11 phases**

---

## Phase 1: SQL Schema Changes

**Goal:** Add `allowed_child_types` column to `datatypes` and `admin_datatypes` tables across all three backends. Update all Create and Update queries to include the new column.

**Prerequisite:** None. This is the foundation phase.

### 1.1 datatypes table — SQLite

**File:** `sql/schema/7_datatypes/schema.sql`

Add column after `type`:
```sql
allowed_child_types TEXT DEFAULT NULL,
```

Position: after the `type TEXT NOT NULL,` line, before `author_id TEXT NOT NULL`.

The full column list after change:
```
datatype_id, parent_id, sort_order, name, label, type, allowed_child_types, author_id, date_created, date_modified
```

**File:** `sql/schema/7_datatypes/queries.sql`

Update these queries to include `allowed_child_types`:

1. **CreateDatatype** — Add `allowed_child_types` to INSERT column list and VALUES placeholder list. Add `@allowed_child_types` parameter.

   Current INSERT columns: `datatype_id, parent_id, sort_order, name, label, type, author_id, date_created, date_modified`

   New INSERT columns: `datatype_id, parent_id, sort_order, name, label, type, allowed_child_types, author_id, date_created, date_modified`

   Current VALUES: `?, ?, ?, ?, ?, ?, ?, ?, ?` (9 placeholders)
   New VALUES: `?, ?, ?, ?, ?, ?, ?, ?, ?, ?` (10 placeholders)

2. **UpdateDatatype** — Add `allowed_child_types = ?` to SET clause.

   Current SET: `parent_id = ?, sort_order = ?, name = ?, label = ?, type = ?, author_id = ?, date_created = ?, date_modified = ?`

   New SET: `parent_id = ?, sort_order = ?, name = ?, label = ?, type = ?, allowed_child_types = ?, author_id = ?, date_created = ?, date_modified = ?`

No changes to SELECT queries — `SELECT *` already returns all columns including new ones.

### 1.2 datatypes table — MySQL

**File:** `sql/schema/7_datatypes/schema_mysql.sql`

Add column after `type TEXT NOT NULL,`:
```sql
allowed_child_types TEXT NULL,
```

Note: MySQL uses `NULL` not `DEFAULT NULL` in the column definition (though both work; match existing style in this file).

**File:** `sql/schema/7_datatypes/queries_mysql.sql`

Same query changes as SQLite (1.1 step 2). Placeholder format is identical (`?`).

MySQL's `CreateDatatype` is `:exec` (no RETURNING), so the change is just adding the column to INSERT and VALUES.

### 1.3 datatypes table — PostgreSQL

**File:** `sql/schema/7_datatypes/schema_psql.sql`

Add column after `type TEXT NOT NULL,`:
```sql
allowed_child_types TEXT DEFAULT NULL,
```

**File:** `sql/schema/7_datatypes/queries_psql.sql`

Same query changes as SQLite but with PostgreSQL positional placeholders (`$1, $2, ...`). The parameter position of every placeholder after `allowed_child_types` shifts by +1.

Current CreateDatatype: `$1, $2, $3, $4, $5, $6, $7, $8, $9` (9 params)
New CreateDatatype: `$1, $2, $3, $4, $5, $6, $7, $8, $9, $10` (10 params)

Current UpdateDatatype SET uses `$1` through `$8` for SET columns and `$9` for WHERE.
New: `$1` through `$9` for SET columns, `$10` for WHERE.

**Critical: Verify the exact parameter ordering** by reading the current queries_psql.sql file. The parameter numbers must match the `@param_name` annotations in order.

### 1.4 admin_datatypes table — SQLite

**File:** `sql/schema/9_admin_datatypes/schema.sql`

Add column after `type TEXT NOT NULL,`:
```sql
allowed_child_types TEXT DEFAULT NULL,
```

**File:** `sql/schema/9_admin_datatypes/queries.sql`

Same pattern as datatypes (1.1). Update CreateAdminDatatype and UpdateAdminDatatype to include `allowed_child_types`.

### 1.5 admin_datatypes table — MySQL

**File:** `sql/schema/9_admin_datatypes/schema_mysql.sql`

Add `allowed_child_types TEXT NULL,` after `type`.

**File:** `sql/schema/9_admin_datatypes/queries_mysql.sql`

Same query changes.

### 1.6 admin_datatypes table — PostgreSQL

**File:** `sql/schema/9_admin_datatypes/schema_psql.sql`

Add `allowed_child_types TEXT DEFAULT NULL,` after `type`.

**File:** `sql/schema/9_admin_datatypes/queries_psql.sql`

Same query changes with positional placeholder renumbering.

### 1.7 Run sqlc

```bash
just sqlc
```

This regenerates:
- `internal/db-sqlite/` (package `mdb`)
- `internal/db-mysql/` (package `mdbm`)
- `internal/db-psql/` (package `mdbp`)

**Verify:** The generated `models.go` in each package now has `AllowedChildTypes` field on both `Datatypes` and `AdminDatatypes` structs. The field type will be `sql.NullString` (SQLite/PostgreSQL) or `sql.NullString` (MySQL).

### 1.8 Verify compilation

```bash
just check
```

This will fail because the wrapper types and mappers in `internal/db/` don't have the new field yet. That's expected — Phase 2 fixes it.

---

## Phase 2: Go Wrapper Types and Mappers

**Goal:** Add `AllowedChildTypes` to application-level wrapper types, Create/Update params, and mapper methods. These files are generated by `just dbgen` from entity definitions, but some are hand-maintained.

**Prerequisite:** Phase 1 complete (`just sqlc` run successfully).

### 2.1 Determine which files are generated vs hand-maintained

Check the file headers:
- `internal/db/datatype_gen.go` — **Generated by dbgen.** Do NOT hand-edit. Update the entity definition that dbgen reads.
- `internal/db/admin_datatype_gen.go` — **Generated by dbgen.** Same.
- `internal/db/datatype_custom.go` — **Hand-maintained.** Contains custom wrapper methods.
- `internal/db/admin_datatype_custom.go` — **Hand-maintained.**
- `internal/db/datatype_sort_order.go` — **Hand-maintained.** Sort order specific.

**Action:** Run `just dbgen-entity Datatypes` and `just dbgen-entity AdminDatatypes` to regenerate the `_gen.go` files. If dbgen doesn't automatically pick up the new sqlc field, the entity definition file that dbgen reads from must be updated first.

To understand what dbgen needs, check:
1. Read `internal/db/db.go` for entity definitions or the dbgen source at `tools/dbgen/`
2. The dbgen tool reads the sqlc-generated models and produces wrapper types

**If dbgen auto-derives from sqlc models:** Just run `just dbgen` and verify the new field appears.

**If dbgen uses a separate entity definition:** Update that definition to include `AllowedChildTypes`.

### 2.2 Expected wrapper type changes

After regeneration, these structs should have a new field:

```go
// In datatype_gen.go
type Datatypes struct {
    DatatypeID        types.DatatypeID         `json:"datatype_id"`
    ParentID          types.NullableDatatypeID `json:"parent_id"`
    SortOrder         int64                    `json:"sort_order"`
    Name              string                   `json:"name"`
    Label             string                   `json:"label"`
    Type              string                   `json:"type"`
    AllowedChildTypes types.NullableString      `json:"allowed_child_types"`  // NEW
    AuthorID          types.UserID             `json:"author_id"`
    DateCreated       types.Timestamp          `json:"date_created"`
    DateModified      types.Timestamp          `json:"date_modified"`
}
```

Same for `CreateDatatypeParams`, `UpdateDatatypeParams`, `AdminDatatypes`, `CreateAdminDatatypeParams`, `UpdateAdminDatatypeParams`.

### 2.3 Verify mapper methods

The mapper methods (`MapDatatype`, `MapCreateDatatypeParams`, `MapUpdateDatatypeParams`) must map the new field between sqlc types and wrapper types. If dbgen generates these, they should update automatically.

If mappers are in `_custom.go` files, update them manually to include `AllowedChildTypes`.

### 2.4 Verify audited command types

The audited command constructors (`NewDatatypeCmd`, `UpdateDatatypeCmd`) pass params to sqlc. If dbgen generates these, they update automatically. If they're in `_gen.go` files, `just dbgen` handles it.

### 2.5 String and JSON helper types

Check if `StringDatatypes` and `DatatypeJSON` structs exist in the generated files. If so, add `AllowedChildTypes string` field to both.

### 2.6 Run drivergen for custom wrappers

```bash
just drivergen
```

This replicates any custom SQLite methods to MySQL/PostgreSQL variants.

### 2.7 Compile check

```bash
just check
```

This must pass. If it fails, the error output identifies which mapper or constructor is missing the new field.

### 2.8 Run tests

```bash
just test
```

Existing tests should still pass since `allowed_child_types` defaults to NULL (no behavior change for existing data).

---

## ~~Phase 3: Migration~~ — SKIPPED

Greenfield project with no active users. No migration infrastructure needed. The column is part of the DDL from Phase 1 — fresh installs get it automatically via `CreateAllTables()`.

---

## Phase 4: Validation Logic

**Goal:** Create nesting validation functions and integrate into content create and move operations.

**Prerequisite:** Phase 2 complete (wrapper types compile).

### 4.1 Create validation module

**File:** `internal/service/nesting.go` (new file)

```go
package service

// ValidateNesting checks whether a child content node with the given
// childDatatypeName is allowed under a parent whose datatype has the given
// allowedChildTypes constraint.
//
// Rules:
//   - If parentAllowedChildTypes is empty string or "null" → no restriction (allow)
//   - If parentAllowedChildTypes is "[]" → no children allowed (reject)
//   - If parentAllowedChildTypes is a JSON array of names → child must be in list
//
// Returns nil if allowed, NestingError if rejected.
func ValidateNesting(parentDatatypeName string, parentAllowedChildTypes string, childDatatypeName string) error {
    // ...
}
```

**NestingError type:**

```go
// NestingError is returned when a content node violates its parent's
// allowed_child_types constraint.
type NestingError struct {
    ParentDatatypeName     string
    ChildDatatypeName      string
    AllowedChildTypes      []string
}

func (e *NestingError) Error() string {
    return fmt.Sprintf("datatype %q is not allowed as a child of %q (allowed: %v)",
        e.ChildDatatypeName, e.ParentDatatypeName, e.AllowedChildTypes)
}
```

**Implementation details:**

1. Parse `parentAllowedChildTypes` as JSON `[]string`. If parse fails, log warning and allow (fail-open for data integrity).
2. Check if `childDatatypeName` is in the parsed list.
3. Return `nil` (allowed) or `*NestingError` (rejected).

### 4.2 Helper: resolve datatype for content node

The validation needs to look up the parent content node's datatype to get its `allowed_child_types`. Add a helper:

```go
// resolveParentConstraint fetches the parent content node's datatype and
// returns its allowed_child_types value. Returns ("", nil) if the parent
// has no constraint or doesn't exist.
func (s *ContentService) resolveParentConstraint(ctx context.Context, parentID types.ContentID) (string, string, error) {
    // 1. Get parent content node → get its datatype_id
    // 2. Get the datatype → get allowed_child_types and name
    // 3. Return (datatypeName, allowedChildTypes, nil)
}
```

Same for `AdminContentService`.

### 4.3 Integrate into ContentService.Create()

**File:** `internal/service/content.go`

Insert validation **after** parent root_id resolution (line ~142) and **before** the database insert (line ~144):

```go
func (s *ContentService) Create(ctx context.Context, ac audited.AuditContext, params db.CreateContentDataParams) (*db.ContentData, error) {
    // Resolve root_id (existing code, lines 133-142)
    ...

    // NEW: Validate nesting constraint
    if params.ParentID.Valid {
        parentName, constraint, err := s.resolveParentConstraint(ctx, params.ParentID.ID)
        if err != nil {
            return nil, fmt.Errorf("create: resolve parent constraint: %w", err)
        }
        if constraint != "" {
            childDT, err := s.driver.GetDatatype(params.DatatypeID.ID)
            if err != nil {
                return nil, fmt.Errorf("create: get child datatype: %w", err)
            }
            if err := ValidateNesting(parentName, constraint, childDT.Name); err != nil {
                return nil, err
            }
        }
    }

    // Create content data (existing code, line 144)
    cd, err := s.driver.CreateContentData(ctx, ac, params)
    ...
}
```

### 4.4 Integrate into AdminContentService.Create()

**File:** `internal/service/content_admin.go`

Same pattern as 4.3, using admin types:

```go
func (s *AdminContentService) Create(ctx context.Context, ac audited.AuditContext, params db.CreateAdminContentDataParams) (*db.AdminContentData, error) {
    // Resolve root_id (existing code, lines 132-142)
    ...

    // NEW: Validate nesting constraint
    if params.ParentID.Valid {
        parentName, constraint, err := s.resolveParentConstraint(ctx, params.ParentID.ID)
        if err != nil {
            return nil, fmt.Errorf("admin create: resolve parent constraint: %w", err)
        }
        if constraint != "" {
            childDT, err := s.driver.GetAdminDatatypeById(params.AdminDatatypeID.ID)
            if err != nil {
                return nil, fmt.Errorf("admin create: get child datatype: %w", err)
            }
            if err := ValidateNesting(parentName, constraint, childDT.Name); err != nil {
                return nil, err
            }
        }
    }

    // Create admin content data (existing code, line 144)
    cd, err := s.driver.CreateAdminContentData(ctx, ac, params)
    ...
}
```

### 4.5 Integrate into Move operations

**File:** `internal/service/content.go` — Add validation in the Move method.

The Move method delegates to `ops.Move()`. Validation should happen **before** the move, at the service layer:

```go
func (s *ContentService) Move(ctx context.Context, ac audited.AuditContext, params ops.MoveParams[types.ContentID]) (*ops.MoveResult[types.ContentID], error) {
    // NEW: Validate nesting constraint at new parent
    if params.NewParentID.Valid {
        parentName, constraint, err := s.resolveParentConstraint(ctx, types.ContentID(params.NewParentID.Value))
        if err != nil {
            return nil, fmt.Errorf("move: resolve parent constraint: %w", err)
        }
        if constraint != "" {
            node, err := s.driver.GetContentData(params.NodeID)
            if err != nil {
                return nil, fmt.Errorf("move: get node: %w", err)
            }
            childDT, err := s.driver.GetDatatype(node.DatatypeID.ID)
            if err != nil {
                return nil, fmt.Errorf("move: get child datatype: %w", err)
            }
            if err := ValidateNesting(parentName, constraint, childDT.Name); err != nil {
                return nil, err
            }
        }
    }

    return ops.Move(ctx, ac, s, params)
}
```

Same pattern for `AdminContentService.Move()`.

**Important:** The `tree/ops/move.go` file is NOT modified. Validation stays at the service layer. The ops package has no dependency on db types and should remain generic.

### 4.6 Create nesting_test.go

**File:** `internal/service/nesting_test.go` (new file)

Test cases:
1. NULL constraint → allow any child
2. Empty string constraint → allow any child
3. `[]` constraint → reject all children
4. `["row", "column"]` constraint → allow row, allow column, reject hero
5. Invalid JSON constraint → allow (fail-open, log warning)
6. Case sensitivity: names are case-sensitive (match exact)

---

## Phase 5: API Layer

**Goal:** Update router handlers to accept `allowed_child_types` on datatype create/update, and return nesting errors with appropriate HTTP status codes.

**Prerequisite:** Phase 4 complete.

### 5.1 Update datatype create/update handlers

**File:** `internal/router/adminDatatypes.go`

The admin datatype create handler parses the request body into `CreateAdminDatatypeParams`. Since the struct now has `AllowedChildTypes`, JSON deserialization handles it automatically if the API accepts JSON bodies.

**Verify:** Read the current handler to confirm it uses `json.Decode` into the params struct. If so, no handler changes needed — the new field is automatically parsed from the JSON body.

**If the handler uses form parsing or manual field extraction:** Add parsing for the new field.

**File:** `internal/router/datatypes.go`

Same verification and changes for public datatypes.

### 5.2 Error response for nesting violations

The service layer returns `*NestingError`. Router handlers should detect this and return HTTP 422 (Unprocessable Entity) with a descriptive error message.

**Check existing error handling pattern:** Read the content create handler to see how service errors are currently mapped to HTTP responses. Follow the same pattern for `NestingError`.

If the handler uses a generic error-to-status mapper, register `*NestingError` → 422.

If the handler checks error types manually:
```go
var nestingErr *service.NestingError
if errors.As(err, &nestingErr) {
    http.Error(w, nestingErr.Error(), http.StatusUnprocessableEntity)
    return
}
```

### 5.3 Validate allowed_child_types format on datatype create/update

When a user sets `allowed_child_types`, validate that:
1. It's valid JSON
2. It's an array of strings
3. Each string references an existing datatype name (optional — soft validation with warning)

Add validation in the datatype create/update service methods or handlers:

```go
func validateAllowedChildTypesFormat(value string) error {
    if value == "" {
        return nil // NULL equivalent
    }
    var names []string
    if err := json.Unmarshal([]byte(value), &names); err != nil {
        return fmt.Errorf("allowed_child_types must be a JSON array of strings: %w", err)
    }
    return nil
}
```

Place this validation in the router handler or service layer for datatype create/update.

---

## Phase 6: Admin Panel UI

**Goal:** Add UI controls for configuring `allowed_child_types` on datatype create and edit forms. Update the block editor picker to filter by constraint.

**Prerequisite:** Phase 5 complete (API accepts the new field).

### 6.1 Update datatype detail page

**File:** `internal/admin/pages/datatype_detail.templ`

Add a new form section below the existing Type field for configuring allowed child types. The UI should be a multi-select or tag input showing available datatype names.

**Design:** Use a `<textarea>` with JSON array content (matching the existing pattern for `data`, `validation`, and `ui_config` fields). The user enters a JSON array of datatype names.

**Alternative (better UX):** Use a multi-select `<select multiple>` populated with all datatype names. On form submit, serialize selected values to JSON array.

**Implementation:**
1. The handler (`DatatypeDetailHandler`) already fetches the datatype. Pass `allowed_child_types` to the template.
2. The handler also needs to fetch all datatype names for the picker. Add `ListDatatypes()` call and extract names.
3. Template renders a multi-select with all available names, pre-selecting those in `allowed_child_types`.
4. On form submit, the handler receives selected names as `allowed_child_types[]` form array.
5. Handler serializes to JSON array string and passes to `UpdateDatatypeParams.AllowedChildTypes`.

### 6.2 Update datatype create dialog partial

**File:** `internal/admin/partials/datatype_form.templ`

Add `allowed_child_types` field to the create dialog. For create, this can be a simple textarea (JSON array) since the create dialog is lightweight. Or omit from create and only allow configuration on the detail/edit page.

**Recommendation:** Omit from create dialog. Users create the datatype first, then configure constraints on the detail page. This keeps the create dialog simple and matches the pattern for other advanced fields (validation, ui_config).

### 6.3 Update datatype edit form partial

**File:** `internal/admin/partials/datatype_edit_form.templ`

Add the allowed_child_types field. This partial is swapped in by HTMX on the detail page. Include the multi-select or textarea control.

### 6.4 Update datatype handlers

**File:** `internal/admin/handlers/datatypes.go`

**DatatypeDetailHandler:**
- Add `allDatatypeNames []string` to template data
- Fetch via `ListDatatypes()` → extract names → pass to template

**DatatypeUpdateHandler:**
- Parse `allowed_child_types[]` from form (multi-select) or `allowed_child_types` from textarea
- Serialize to JSON array string
- Set on `UpdateDatatypeParams.AllowedChildTypes`

### 6.5 Update block editor to filter by constraint

**File:** `internal/admin/static/js/block-editor-src/cache.js`

The `fetchDatatypes()` function returns `{ id, parentId, name, label, type }` for each datatype. Add `allowedChildTypes` to the cached shape:

```javascript
return {
    id: dt.datatype_id,
    parentId: dt.parent_id || null,
    name: dt.name,
    label: dt.label,
    type: dt.type,
    allowedChildTypes: dt.allowed_child_types || null  // NEW
};
```

The `fetchDatatypesGrouped()` function builds categories. After building categories, add a filtering step:

```javascript
// If the parent block's datatype has allowedChildTypes, filter results
if (parentDatatypeAllowedChildTypes) {
    var allowed = JSON.parse(parentDatatypeAllowedChildTypes);
    categories = categories.map(function(cat) {
        return {
            name: cat.name,
            items: cat.items.filter(function(item) {
                return allowed.indexOf(item.name) !== -1;
            })
        };
    }).filter(function(cat) { return cat.items.length > 0; });
}
```

**Note:** `fetchDatatypesGrouped` currently takes `rootDatatypeId` as its only parameter. It needs an additional parameter: `parentAllowedChildTypes` (the constraint from the parent block's datatype). This requires updating the caller in `picker.js` to pass the parent's constraint.

**File:** `internal/admin/static/js/block-editor-src/picker.js`

Update `_openPicker()` to pass the parent block's `allowedChildTypes` to `fetchDatatypesGrouped()`:

```javascript
_openPicker: function(insertTargetId, position) {
    // Determine parent block's datatype and its allowed_child_types
    var parentBlock = this._state.blocks[insertTargetId];
    var parentDt = parentBlock ? this._findDatatypeById(parentBlock.datatypeId) : null;
    var parentConstraint = parentDt ? parentDt.allowedChildTypes : null;

    fetchDatatypesGrouped(this._rootDatatypeId, parentConstraint).then(...)
}
```

### 6.6 Update admin API endpoint for datatypes

The admin panel fetches datatypes from `/admin/api/datatypes`. Verify this endpoint returns `allowed_child_types` in the response. Since it serializes `db.Datatypes` which now has the field, it should work automatically via JSON tags.

### 6.7 Rebuild block editor bundle

```bash
just admin bundle
```

---

## Phase 7: TUI

**Goal:** Update the Bubbletea TUI to display and edit `allowed_child_types` on datatypes, and respect the constraint when filtering child datatypes for content creation.

**Prerequisite:** Phase 4 complete. Independent of Phase 6.

### 7.1 Update filterChildDatatypes to respect constraint

**File:** `internal/tui/update_fetch.go`

The `filterChildDatatypes()` function currently filters by datatype hierarchy (3-category logic). Add a constraint intersection step:

```go
func filterChildDatatypes(all []db.Datatypes, rootDatatypeID types.DatatypeID, parentAllowedChildTypes string) []db.Datatypes {
    // Existing 3-category filtering (lines 35-96)
    filtered := existingFilterLogic(all, rootDatatypeID)

    // NEW: Intersect with parent's allowed_child_types if set
    if parentAllowedChildTypes == "" {
        return filtered
    }

    var allowed []string
    if err := json.Unmarshal([]byte(parentAllowedChildTypes), &allowed); err != nil {
        return filtered // fail-open on parse error
    }

    allowedSet := make(map[string]bool, len(allowed))
    for _, name := range allowed {
        allowedSet[name] = true
    }

    var result []db.Datatypes
    for _, dt := range filtered {
        if allowedSet[dt.Name] {
            result = append(result, dt)
        }
    }
    return result
}
```

**Signature change:** Add `parentAllowedChildTypes string` parameter. Update all callers.

### 7.2 Update callers of filterChildDatatypes

**File:** `internal/tui/screen_content.go` (or wherever the child datatype fetch command is)

The `ShowChildDatatypeDialogCmd` fetches all datatypes and calls `filterChildDatatypes()`. It needs to also fetch the parent content node's datatype to get its `allowed_child_types`.

Current flow:
1. Fetch all datatypes
2. Call `filterChildDatatypes(all, rootDatatypeID)`

New flow:
1. Fetch all datatypes
2. Fetch parent content node → get its datatype_id
3. Fetch parent's datatype → get `allowed_child_types`
4. Call `filterChildDatatypes(all, rootDatatypeID, parentAllowedChildTypes)`

### 7.3 Update admin variant

Same changes for admin mode: `filterAdminChildDatatypes` (if it exists) or the admin branch of the filtering logic.

**Note:** The research showed there is no separate `filterAdminChildDatatypes` function. The admin mode likely uses a parallel code path or the same function with admin types. Verify by reading the admin content creation flow in `screen_content.go`.

### 7.4 Update datatype edit dialog

**File:** `internal/tui/screen_datatypes.go` or `internal/tui/form_dialog_constructors.go`

The datatype edit dialog should show `allowed_child_types` as a form field. Options:

**Option A (simple):** Add a text input field for the JSON array string. The user types `["row", "column"]` directly.

**Option B (better UX):** Add a multi-select carousel showing all available datatype names. Selected names form the constraint.

**Recommendation for TUI:** Option A (text input). The TUI is a power-user interface and JSON input is acceptable. Option B requires significant new TUI widget code.

### 7.5 Update message types

If the child datatype fetch messages carry the constraint, update:

**File:** `internal/tui/msg_fetch.go` or `internal/tui/msg_admin.go`

Add `ParentAllowedChildTypes string` to `FetchChildDatatypesMsg` (or equivalent message type).

---

## Phase 8: Definitions and Deploy

**Goal:** Update the schema definition system to support `allowed_child_types` on presets, and ensure deploy import/export handles the new column.

**Prerequisite:** Phase 2 complete (wrapper types have the new field).

### 8.1 Update DatatypeDef struct

**File:** `internal/definitions/definition.go`

Add field:
```go
type DatatypeDef struct {
    Name              string
    Label             string
    Type              types.NullableString
    ParentRef         string
    FieldRefs         []FieldDef
    AllowedChildTypes []string  // NEW: datatype names allowed as children (nil = no constraint)
}
```

### 8.2 Update definition install logic

**File:** `internal/definitions/install.go`

The install function creates datatypes. After creating a datatype, if `AllowedChildTypes` is non-nil, set the value:

```go
// During datatype creation:
allowedJSON := ""
if len(dtDef.AllowedChildTypes) > 0 {
    b, _ := json.Marshal(dtDef.AllowedChildTypes)
    allowedJSON = string(b)
}
// Pass to CreateDatatypeParams.AllowedChildTypes
```

**Important:** The `AllowedChildTypes` references other datatypes by name. During install, some referenced datatypes may not exist yet (created in a later iteration of the install loop). The install already handles circular parent references via iterative resolution. Apply the same pattern:

1. First pass: Create all datatypes with `AllowedChildTypes = NULL`
2. Second pass: After all datatypes exist, set `AllowedChildTypes` on datatypes that have it defined

This avoids name-resolution failures during the first pass. Alternatively, since `allowed_child_types` is just a JSON string of names (not an FK), it can be set on creation without the names needing to exist in the database. The names are validated at content-creation time, not at datatype-creation time. **So the second pass is unnecessary — set on first creation.**

### 8.3 Update preset definitions (optional for this phase)

**Files:** `internal/definitions/def_modulacms.go`, `def_wordpress.go`, `def_strapi.go`, `def_sanity.go`, `def_contentful.go`

Add `AllowedChildTypes` to datatypes where appropriate. Examples from the modula-default preset:

```go
// A "page" root type might allow: row, hero, section, cta, media_block
"page": DatatypeDef{
    Name:              "page",
    Label:             "Page",
    Type:              types.NullableString{String: "page", Valid: true},
    AllowedChildTypes: []string{"row", "hero", "section", "cta", "media_block"},
    FieldRefs:         []FieldDef{...},
},

// A "row" layout type might allow: column
"row": DatatypeDef{
    Name:              "row",
    Label:             "Row",
    Type:              types.NullableString{String: "layout", Valid: true},
    ParentRef:         "page",
    AllowedChildTypes: []string{"column"},
    FieldRefs:         []FieldDef{...},
},
```

**Note:** This is optional for the initial implementation. Existing presets can use `nil` (no constraint) initially, with constraints added in a follow-up task. The feature works correctly without preset constraints — they're enforced at content creation time, not at schema definition time.

### 8.4 Deploy import/export

**No changes needed.** The deploy system serializes/deserializes via struct JSON tags. Since the wrapper types now have `AllowedChildTypes` with a `json:"allowed_child_types"` tag, the export serializer includes it automatically and the import deserializer reads it automatically.

**Verify by reading:**
- `internal/deploy/export.go` — Uses `structSliceToTableData()` which reflects on struct fields
- `internal/deploy/import.go` — Uses bulk insert which maps columns to struct fields

---

## Phase 9: SDKs

**Goal:** Add `allowed_child_types` field to Datatype types in all three SDKs.

**Prerequisite:** Phase 2 complete (API returns the new field).

### 9.1 TypeScript SDK — Types package

**File:** `sdks/typescript/types/src/entities/schema.ts`

Update `Datatype` type:
```typescript
export type Datatype = {
  datatype_id: DatatypeID
  parent_id: ContentID | null
  sort_order: number
  name: string
  label: string
  type: string
  allowed_child_types: string[] | null  // NEW
  author_id: UserID | null
  date_created: string
  date_modified: string
}
```

### 9.2 TypeScript SDK — Admin SDK params

**File:** `sdks/typescript/modulacms-admin-sdk/src/types/schema.ts`

Update `CreateDatatypeParams` and `UpdateDatatypeParams`:
```typescript
export type CreateDatatypeParams = {
  // ... existing fields ...
  allowed_child_types?: string[] | null  // NEW (optional on create)
}

export type UpdateDatatypeParams = {
  // ... existing fields ...
  allowed_child_types?: string[] | null  // NEW (optional on update)
}
```

### 9.3 Go SDK

**File:** `sdks/go/types.go`

Update `Datatype` struct:
```go
type Datatype struct {
    DatatypeID        DatatypeID  `json:"datatype_id"`
    ParentID          *DatatypeID `json:"parent_id"`
    Name              string      `json:"name"`
    Label             string      `json:"label"`
    Type              string      `json:"type"`
    AllowedChildTypes []string    `json:"allowed_child_types"`  // NEW
    AuthorID          *UserID     `json:"author_id"`
    DateCreated       Timestamp   `json:"date_created"`
    DateModified      Timestamp   `json:"date_modified"`
}
```

Update `CreateDatatypeParams` and `UpdateDatatypeParams`:
```go
type CreateDatatypeParams struct {
    // ... existing fields ...
    AllowedChildTypes []string `json:"allowed_child_types,omitempty"`  // NEW
}

type UpdateDatatypeParams struct {
    // ... existing fields ...
    AllowedChildTypes []string `json:"allowed_child_types,omitempty"`  // NEW
}
```

### 9.4 Swift SDK

**File:** `sdks/swift/Sources/Modula/Types.swift`

Update `Datatype` struct:
```swift
public struct Datatype: Codable, Sendable {
    // ... existing fields ...
    public let allowedChildTypes: [String]?  // NEW

    enum CodingKeys: String, CodingKey {
        // ... existing keys ...
        case allowedChildTypes = "allowed_child_types"  // NEW
    }
}
```

Update `CreateDatatypeParams` and `UpdateDatatypeParams`:
```swift
public struct CreateDatatypeParams: Encodable, Sendable {
    // ... existing fields ...
    public let allowedChildTypes: [String]?  // NEW

    public init(
        // ... existing params ...
        allowedChildTypes: [String]? = nil  // NEW (optional)
    ) { ... }

    enum CodingKeys: String, CodingKey {
        // ... existing keys ...
        case allowedChildTypes = "allowed_child_types"  // NEW
    }
}
```

### 9.5 Build and test SDKs

```bash
just sdk ts build
just sdk ts typecheck
just sdk go vet
just sdk swift build
```

---

## Phase 10: Tests

**Goal:** Add tests covering nesting constraint behavior across all layers.

**Prerequisite:** Phases 1-9 complete.

### 10.1 Unit tests — validation logic

**File:** `internal/service/nesting_test.go`

Table-driven tests:

| Scenario | Parent Constraint | Child Name | Expected |
|----------|------------------|------------|----------|
| NULL constraint | `""` | `"anything"` | allow |
| Empty array | `"[]"` | `"anything"` | reject |
| Child in list | `'["row","col"]'` | `"row"` | allow |
| Child not in list | `'["row","col"]'` | `"hero"` | reject |
| Invalid JSON | `"not json"` | `"anything"` | allow (fail-open) |
| Single item | `'["row"]'` | `"row"` | allow |
| Single item no match | `'["row"]'` | `"col"` | reject |
| Case sensitive | `'["Row"]'` | `"row"` | reject |

### 10.2 Integration tests — content create with constraint

**File:** `internal/service/content_test.go` (new or existing)

1. Create datatype "page" with `allowed_child_types = ["row"]`
2. Create datatype "row" and "hero"
3. Create content node with datatype "page"
4. Create child content with datatype "row" → expect success
5. Create child content with datatype "hero" → expect NestingError
6. Create child content with datatype "row" under NULL-constraint parent → expect success

### 10.3 Integration tests — move with constraint

1. Create tree: page (allows ["row"]) → row → hero
2. Move "hero" to be direct child of "page" → expect NestingError
3. Move "row" to be child of "page" → expect success

### 10.4 TUI tests — filterChildDatatypes

**File:** `internal/tui/update_fetch_test.go` (new or existing)

1. With constraint `["row", "column"]`: only row and column appear
2. With constraint `[]`: nothing appears
3. With constraint `""` (NULL): current behavior (all hierarchy-based filtering)

### ~~10.5 Migration test~~ — SKIPPED (greenfield)

---

## Phase 11: Documentation

**Goal:** Update project documentation and CLAUDE.md.

**Prerequisite:** All implementation phases complete.

### 11.1 CLAUDE.md

Add to the "Content Tree Structure" section:

```
### Nesting Constraints

Datatypes can optionally define `allowed_child_types` — a JSON array of
datatype names that are permitted as children in the content tree. When
set, content creation and move operations validate against this constraint.
NULL means no restriction (current behavior).
```

### 11.2 User documentation

**File:** `documentation/guides/content-modeling.md`

Add section on configuring nesting constraints.

### 11.3 API reference

If API documentation exists, add `allowed_child_types` to the datatype endpoints.

---

## Implementation Order and Dependencies

```
Phase 1 (SQL)
    ↓
Phase 2 (Go types) ← depends on Phase 1
    ↓
Phase 3 — SKIPPED (greenfield)
    ↓
Phase 4 (Validation) ← depends on Phase 2
    ↓
Phase 5 (API) ← depends on Phase 4
    ↓
┌──────────────────┬──────────────────┬──────────────────┐
Phase 6 (Admin UI)  Phase 7 (TUI)     Phase 8 (Defs)     ← independent of each other
└──────────────────┴──────────────────┴──────────────────┘
    ↓                   ↓                   ↓
Phase 9 (SDKs) ← depends on Phase 2
    ↓
Phase 10 (Tests) ← depends on all
    ↓
Phase 11 (Docs)
```

Phases 6, 7, 8, and 9 can be implemented in parallel by separate agents.

---

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| Name collision if datatype renamed | Log warning at validation time; names are stable identifiers |
| NULL-safety across three DBs | Use `types.NullableString` which handles NULL consistently |
| Existing content violates new constraint | Constraint only applies to new creates/moves; existing data is grandfathered |
| Performance: extra DB queries on create/move | One additional GetDatatype query per create/move — negligible |
| Block editor JS changes break existing behavior | NULL constraint preserves current behavior; only set constraints change filtering |
| ~~SQLite ALTER TABLE limitations~~ | N/A — greenfield, column is in DDL from day one |
| Import/export with constraints referencing non-existent names | Names are just strings; validation is at content-create time, not import time |

---

## Open Questions

1. **admin_datatypes name uniqueness:** The `datatypes` table has `UNIQUE INDEX` on `name`, but `admin_datatypes` does not. Should we add one? This affects whether `allowed_child_types` names can reliably reference admin datatypes. **Recommendation:** Add UNIQUE INDEX to admin_datatypes.name as part of Phase 1.

2. **Recursive constraint checking:** If Page allows [Row], and Row allows [Column], should creating a Column directly under Page be rejected (yes — Page doesn't list Column)? **Answer:** Yes. Each parent independently enforces its own constraint. The constraint is non-recursive.

3. **Constraint on root-level content:** Should `allowed_child_types` apply to root content nodes (no parent)? **Answer:** No — root nodes have no parent, so there's no parent constraint to check. Routes/trees can have any root datatype.

4. **Wildcard or negation patterns:** Should `allowed_child_types` support "all except X"? **Answer:** No. Keep it simple — explicit allowlist only. Use NULL for "allow all".

5. **Should the picker show disallowed types (greyed out) or hide them?** **Recommendation:** Hide them. Greyed-out items add visual noise without value.
