# TUI QA Phase 3: Schema Management

Covers: Datatypes (browse + fields phases, search, reorder), Fields (CRUD + reorder), Field Types, Routes.

## Prerequisites

- DB Wipe & Redeploy + Modula Default schema installed

## 3.1 Datatypes Screen — Browse Phase

### 3.1.1 Datatype list renders
- Navigate to Datatypes, wait for stable
- Verify list shows datatypes in tree hierarchy (parent → children indented)
- Verify root types visible: page, post, case_study, documentation, menu, footer
- Verify child types visible under parents: row, columns, cta, etc. under page
- Verify Details panel shows: name, label, type, parent, field count for selected
- Snapshot: `goldens/3.1.1_datatypes_browse.txt`

### 3.1.2 Cursor movement updates detail
- Move cursor through datatypes
- Verify Details panel updates for each selection
- Verify field preview updates (shows field names/types for selected datatype)

### 3.1.3 Search/filter
- Press `/` → verify search input appears
- Type "cta" → verify list filters to show only matching datatypes
- Verify filter is case-insensitive substring match
- Press `escape` → verify search cleared, full list restored
- Press `/`, type something with no match → verify empty list or "no results"

### 3.1.4 Create datatype
- Press `n` → verify create datatype form dialog
- Verify fields: Name, Label, Type
- Fill: name="test_block", label="Test Block", type="content"
- Submit → verify new datatype appears in list
- Verify Details panel shows correct data for new type

### 3.1.5 Edit datatype
- Select a datatype, press `e`
- Verify edit form pre-populated with current values
- Modify label, confirm
- Verify updated label in list and Details panel

### 3.1.6 Delete leaf datatype
- Select a datatype with NO children, press `d`
- Verify confirmation dialog
- Confirm → verify removed from list

### 3.1.7 Delete parent datatype (blocked)
- Select a datatype WITH children (e.g., "page"), press `d`
- Verify error: "Cannot delete because it has child datatypes"
- Verify datatype preserved

### 3.1.8 Expand/Collapse datatype tree
- Select a parent datatype with children
- Press `-` or `left`/`h` → verify children hidden
- Press `+` or `right`/`l` → verify children shown
- Note: expand/collapse uses hardcoded left/right/h/l keys (not KeyMap)

### 3.1.9 Reorder datatypes
- Select a datatype, press `shift+down` → verify sort order changed
- Press `shift+up` → verify restored
- Test boundary (first/last) → verify no crash

## 3.2 Datatypes Screen — Fields Phase

### 3.2.1 Enter fields phase
- Select a datatype (e.g., "cta"), press `enter`
- Verify transition to Phase 2 (fields list)
- Verify breadcrumb shows parent path
- Verify fields listed: Heading, Subheading, Button Text, Button URL (for CTA)
- Verify field details show: label, type, field ID

### 3.2.2 Create field
- Press `n` → verify create field form
- Verify fields: Label, Type (selector from field type registry)
- Navigate Type selector → verify available types (text, textarea, select, boolean, etc.)
- Fill: label="Test Field", type="text"
- Submit → verify field appears in list

### 3.2.3 Edit field
- Select a field, press `e`
- Verify edit form pre-populated
- Modify label, confirm
- Verify updated label in list

### 3.2.4 Delete field
- Select a field, press `d`
- Verify confirmation dialog
- Confirm → verify field removed

### 3.2.5 Reorder fields
- With 2+ fields visible
- Select a field, press `shift+down` → verify order changed
- Press `shift+up` → verify restored
- Boundary test: first/last → no crash

### 3.2.6 Back to browse phase
- Press `h` or `backspace`
- Verify return to datatype browse (Phase 1)

### 3.2.7 Field detail properties
- Select different field types and verify properties panel shows relevant info
- Text field → shows type "text"
- Select field → shows options from data JSON
- Boolean field → shows type "boolean"

## 3.3 Datatypes — Admin Mode

### 3.3.1 Admin datatypes
- Toggle admin mode (ctrl+a from Home)
- Navigate to Datatypes
- Verify admin datatypes shown (separate from public)
- Verify all CRUD operations work in admin mode
- Toggle back to verify public datatypes restored

## 3.4 Field Types Screen

### 3.4.1 Field type list renders
- Navigate to Field Types, wait for stable
- Verify list of registered field types with type badges
- Verify Details panel shows: label, type, ID
- Snapshot: `goldens/3.4.1_field_types.txt`

### 3.4.2 Create field type
- Press `n` → verify create form
- Fill fields, submit
- Verify new type in list

### 3.4.3 Edit field type
- Select field type, press `e`
- Modify, confirm
- Verify updated

### 3.4.4 Delete field type
- Select field type, press `d`
- Confirm → verify removed

### 3.4.5 Admin field types
- Toggle admin mode
- Verify admin field types screen works independently

## 3.5 Routes Screen

### 3.5.1 Route list renders
- Navigate to Routes, wait for stable
- Verify route list shows bootstrap routes (at minimum `/`)
- Verify Details panel shows: slug, title, status, author, dates
- Snapshot: `goldens/3.5.1_routes.txt`

### 3.5.2 Create route
- Press `n` → verify route form dialog
- Verify fields: Slug, Title
- Fill: slug="/blog", title="Blog"
- Submit → verify new route in list

### 3.5.3 Create route with content initialization
- Create a new route
- Verify content is automatically created for the route (or option to initialize)
- Document actual behavior

### 3.5.4 Edit route
- Select route, press `e`
- Verify edit form pre-populated with current slug/title
- Modify title, confirm
- Verify updated in list

### 3.5.5 Delete route
- Select a non-bootstrap route, press `d`
- Verify confirmation dialog
- Confirm → verify removed from list
- Attempt to delete bootstrap `/` route → verify if blocked or allowed

### 3.5.6 Admin routes
- Toggle admin mode
- Verify admin routes screen works
- CRUD operations functional
