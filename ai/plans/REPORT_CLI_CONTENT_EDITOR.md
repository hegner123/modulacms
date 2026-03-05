# CLI Content Editor: Fields & Content Fields Resolution Report

## Executive Summary

The CLI content editor has **two separate field display systems** that serve different purposes, and **both have gaps** in the admin CMS context. The public CMS content editor (tree browser) works correctly for field resolution. The admin CMS content editor has scaffolded types and state storage but **no fetch command, no rendering, and no edit trigger**.

---

## Architecture: Two Separate Systems

### System 1: Right Panel Field Display (Browse Mode)

When a user navigates the content tree and selects a node, the right panel shows that node's field values.

**Flow:**
```
User moves cursor (j/k keys)
  -> contentBrowserCursorUpCmd / contentBrowserCursorDownCmd
  -> LoadContentFieldsCmd(cfg, contentDataID, datatypeID)
  -> ListContentFieldsByContentData()           // get content_field values
  -> ListDatatypeFieldByDatatypeID()            // get junction table entries
  -> GetField() for each junction entry         // N+1 to resolve labels/types
  -> LoadContentFieldsMsg{Fields: []ContentFieldDisplay}
  -> update_state.go sets m.SelectedContentFields
  -> page_builders.go CMSPage.ProcessFields() renders them
```

**Status: WORKING** for public CMS content trees. Fields resolve correctly with label, type, and value.

### System 2: Edit Dialog (Popup Form)

When a user presses `e` on a content node, a form dialog opens with editable field inputs.

**Flow:**
```
User presses 'e'
  -> FetchContentForEditCmd(contentID, datatypeID, routeID, title)
  -> HandleFetchContentForEdit()
     -> ListContentFieldsByContentData()        // existing values
     -> ListDatatypeFieldByDatatypeID()         // junction entries
     -> GetField() per entry                    // N+1 for full definitions
     -> ParseUIConfig() for widget overrides
     -> Build ExistingContentField array
  -> ShowEditContentFormDialogMsg
  -> NewEditContentFormDialog()
     -> FieldBubbleForType() creates TUI widgets (TextInput, NumberInput, etc.)
     -> SetValue() pre-populates each field
  -> Dialog renders and user can edit
```

**Status: WORKING** for public CMS content. The dialog shows fields with labels, correct widget types, and existing values.

---

## The Problem: Admin CMS Content

The admin CMS (tables prefixed `admin_`) has a **parallel but incomplete** implementation:

### What Exists

| Component | File | Status |
|-----------|------|--------|
| `AdminContentFieldDisplay` type | `admin_message_types.go:263` | Defined |
| `AdminLoadContentFieldsMsg` type | `admin_message_types.go:254` | Defined |
| `AdminSelectedContentFields` model field | `model.go:154` | Defined |
| State handler for `AdminLoadContentFieldsMsg` | `update_state.go:276` | Wired |
| `AdminContentFieldsWithFieldRow` DB type | `getTree.go:124` | Defined |
| `ListAdminContentFieldsWithFieldByRoute()` | `db.go:116` | Available |
| `MapAdminContentFieldsWithFieldRow()` | `getTree.go` | Implemented |

### What Does NOT Exist

| Component | Expected Location | Status |
|-----------|-------------------|--------|
| `AdminLoadContentFieldsCmd` (fetch command) | `admin_constructors.go` | **MISSING** |
| Admin content field rendering in view | `admin_panel_view.go` | **MISSING** |
| Admin edit trigger (e key handler) | `update_controls.go` | **MISSING** |
| Admin edit dialog creation | `form_dialog.go` or `update_dialog.go` | **MISSING** |
| Admin content edit fetch handler | `commands.go` | **MISSING** |

### The Gap

The admin CMS content screens can display the content tree (left panel) and content details (center panel), but:

1. **No command ever produces `AdminLoadContentFieldsMsg`** - the message type is defined and the state handler exists, but nothing triggers the fetch.
2. **`admin_panel_view.go` never reads `AdminSelectedContentFields`** - even if fields were fetched, they wouldn't render.
3. **No admin-specific edit flow exists** - pressing `e` in admin content context either does nothing or falls through to the public CMS edit path (which uses public `content_fields` tables, not `admin_content_fields`).

---

## Database Layer Comparison

### Public CMS (Working)

```
content_data  <->  content_fields  <->  fields
              via content_data_id       via field_id

datatypes  <->  datatypes_fields  <->  fields
           via datatype_id             via field_id (junction)
```

**Key queries used by CLI:**
- `ListContentFieldsByContentData(contentDataID)` - raw field values
- `ListDatatypeFieldByDatatypeID(datatypeID)` - junction entries for ordering
- `GetField(fieldID)` - individual field definition (N+1 pattern)
- `ListContentFieldsWithFieldByContentData(contentDataID)` - single JOIN query (available but **not used by CLI**)

### Admin CMS (Not Wired)

```
admin_content_data  <->  admin_content_fields  <->  admin_fields
                    via admin_content_data_id       via admin_field_id

admin_datatypes  <->  admin_datatypes_fields  <->  admin_fields
                 via admin_datatype_id              via admin_field_id (junction)
```

**Available queries not used:**
- `ListAdminContentFieldsWithFieldByRoute(routeID)` - single JOIN query
- `ListAdminContentFields()` / `ListAdminContentFieldsByRoute(routeID)` - raw values
- Individual admin field/content_field CRUD methods

---

## Efficiency Note: N+1 Query Pattern

Both the right-panel display (`LoadContentFieldsCmd`) and the edit dialog (`HandleFetchContentForEdit`) use an N+1 pattern:

```go
// Step 1: Get junction entries
dtFields, err := d.ListDatatypeFieldByDatatypeID(datatypeID)

// Step 2: For each entry, fetch the full field (N queries)
for _, dtf := range *dtFields {
    field, err := d.GetField(dtf.FieldID)  // 1 query per field
}
```

The database layer already has JOIN queries that would eliminate this:
- `ListContentFieldsWithFieldByContentData()` - returns values + definitions in one query
- `ListFieldsWithSortOrderByDatatypeID()` - returns definitions + sort order in one query
- `ListAdminContentFieldsWithFieldByRoute()` - admin equivalent

The admin panel (HTMX) uses the JOIN query correctly:
```go
// internal/admin/handlers/content.go:308
fields, fieldsErr := driver.ListContentFieldsWithFieldByContentData(contentDataID)
```

---

## What Needs to Be Built

### To fix admin CMS content field display (right panel):

1. Create `AdminLoadContentFieldsCmd(cfg, adminContentDataID, adminDatatypeID)` in `admin_constructors.go`
   - Should call `ListAdminContentFieldsWithFieldByRoute()` or equivalent
   - Map results to `[]AdminContentFieldDisplay`
   - Return `AdminLoadContentFieldsMsg`

2. Wire the command into cursor movement in the admin content tree context

3. Add rendering in `admin_panel_view.go` that reads `m.AdminSelectedContentFields`

### To fix admin CMS content field editing (dialog):

4. Create admin-specific edit trigger in `update_controls.go` for admin context
5. Create `AdminFetchContentForEditCmd` / `HandleAdminFetchContentForEdit`
6. Create `AdminNewEditContentFormDialog` or adapt existing to work with admin field types
7. Create admin content field update handler

### Optional efficiency improvement (both systems):

8. Replace N+1 pattern in `LoadContentFieldsCmd` with `ListFieldsWithSortOrderByDatatypeID()`
9. Replace N+1 pattern in `HandleFetchContentForEdit` with `ListContentFieldsWithFieldByContentData()` + `ListFieldsWithSortOrderByDatatypeID()`

---

## Complexity Assessment

This is a **single context window** task for implementing the fix. The patterns are already established in the public CMS code - the admin CMS just needs the same wiring with admin-prefixed types. Estimated ~200-300 lines of new code across 4-5 files, following existing patterns exactly.

No multi-agent plan needed - the work is sequential (constructors -> commands -> wiring -> views) and all patterns are already documented in the public CMS equivalent.

---

## Key Files Reference

| File | Lines | Purpose |
|------|-------|---------|
| `model.go:134,154` | State fields | `SelectedContentFields`, `AdminSelectedContentFields` |
| `commands.go:448-534` | Edit fetch | `HandleFetchContentForEdit` (public CMS) |
| `commands.go:1136-1215` | Panel fetch | `LoadContentFieldsCmd` (public CMS) |
| `form_dialog.go:1608-1710` | Edit dialog | `ExistingContentField`, `NewEditContentFormDialog` |
| `update_controls.go:710-736` | Edit trigger | 'e' key handler |
| `update_dialog.go:1353-1367` | Dialog show | `ShowEditContentFormDialogMsg` handler |
| `update_cms.go:127-132` | Message routing | `FetchContentForEditMsg` routing |
| `update_state.go:276-279` | State set | `AdminLoadContentFieldsMsg` handler |
| `admin_message_types.go:254-269` | Types | `AdminLoadContentFieldsMsg`, `AdminContentFieldDisplay` |
| `admin_panel_view.go` | Rendering | Admin views (fields NOT rendered here) |
| `page_builders.go:591-617,622-` | Rendering | Public CMS field rendering |
| `getTree.go:124-142` | DB types | `AdminContentFieldsWithFieldRow` |
| `db.go:106-117` | Interface | Admin content field methods |
