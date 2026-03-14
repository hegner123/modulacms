# TUI QA Phase 6: Edge Cases & Regression

Covers: Terminal resize, rapid input, empty states, error states, admin mode parity, dialog system stress, form validation, dead keybindings, snapshot regression.

## 6.1 Terminal Resize

### 6.1.1 Resize on home screen
- Start at 120x40
- `tui_resize` to 80x24
- `tui_screen` → verify layout adapts (panels may truncate but no crash)
- `tui_resize` back to 120x40
- Verify layout restored

### 6.1.2 Resize during dialog
- Open a form dialog (e.g., new content)
- `tui_resize` to 80x24
- Verify dialog still visible and functional
- Type text, submit → verify works
- Resize back

### 6.1.3 Resize during tree view
- Enter content tree with several nodes
- Resize to very small (60x20)
- Verify no crash, content truncates gracefully
- Resize to very large (200x60)
- Verify layout expands

### 6.1.4 Minimum viable terminal size
- Resize progressively smaller: 120→100→80→60→40
- At each size, `tui_screen` and verify no panic/crash
- Document minimum size where TUI remains usable

## 6.2 Rapid Input

### 6.2.1 Rapid navigation
- Send 50 `down` keys in rapid succession
- `tui_screen` → verify cursor at a valid position, no crash, no hang
- Send 50 `up` keys → verify cursor at top

### 6.2.2 Rapid key during data fetch
- Press `enter` to enter a screen that fetches data
- Immediately send 10 `down` keys before data loads
- Verify no crash; keys either queued or dropped gracefully

### 6.2.3 Rapid dialog open/close
- Open dialog (press `n`), immediately press `escape` (cancel)
- Repeat 10 times rapidly
- Verify no crash, no orphaned dialogs

### 6.2.4 Rapid text input
- Open a form dialog
- Send a long string (200+ characters) in one `tui_send text=` call
- Verify text appears in input field (may truncate display)
- Submit → verify full text was captured

## 6.3 Empty States

### 6.3.1 Fresh database (no schema)
- DB Wipe & Redeploy (no quickstart install)
- Navigate to Content → verify single bootstrap page or empty message
- Navigate to Datatypes → verify only bootstrap page type + _reference
- Navigate to Media → verify empty state
- Navigate to Routes → verify only bootstrap `/` route
- Navigate to Users → verify only system admin
- Navigate to Plugins → verify "No plugins installed"
- Navigate to Webhooks → verify empty
- Navigate to Pipelines → verify empty

### 6.3.2 All content deleted
- After quickstart install, delete all content
- Navigate to Content → verify empty state within route group
- Verify `n` (new) still works from empty state

### 6.3.3 All routes deleted
- Delete all routes (including bootstrap if allowed)
- Navigate to Content → verify appropriate state
- Navigate to Routes → verify empty list

## 6.4 Error States

### 6.4.1 Form validation — required field empty
- Open content creation form
- Leave required fields empty
- Tab to Confirm, press enter
- Verify validation errors displayed below fields
- Verify form stays open (not submitted)

### 6.4.2 Form validation — invalid URL
- Create a CTA with invalid Button URL (e.g., "not-a-url")
- Verify URL validation catches it (or document if no URL validation)

### 6.4.3 Form validation — field type mismatch
- In Database screen, try to insert a row with invalid type (e.g., text in number field)
- Verify error handling

### 6.4.4 Delete with foreign key references
- In Database screen, try to delete a datatype that has content referencing it
- Verify error message about foreign key constraint

### 6.4.5 Validation error clear on edit
- Get a validation error on a field
- Edit that field (type new text)
- Verify error clears for that specific field
- Verify other field errors remain

## 6.5 Admin Mode Parity

### 6.5.1 Admin content CRUD
- Toggle to admin mode (ctrl+a)
- Create admin content → verify success
- Edit admin content → verify success
- Delete admin content → verify success
- Publish admin content → verify success
- Toggle back → verify public content unchanged

### 6.5.2 Admin datatypes CRUD
- In admin mode: navigate to Datatypes
- Create admin datatype → verify success
- Create admin field on it → verify success
- Edit/delete → verify
- Toggle back → verify public datatypes unchanged

### 6.5.3 Admin routes CRUD
- In admin mode: navigate to Routes
- Create admin route → verify success
- Edit/delete → verify
- Toggle back → verify public routes unchanged

### 6.5.4 Admin field types CRUD
- In admin mode: navigate to Field Types
- Full CRUD cycle → verify works
- Toggle back → verify public field types unchanged

### 6.5.5 Mode indicator consistency
- Toggle admin mode on various screens
- Verify `[Admin]`/`[Client]` indicator always visible in status bar
- Verify mode persists across screen navigation

## 6.6 Dialog System Stress

### 6.6.1 Dialog dismiss via escape
- Open every dialog type (form, confirmation, generic)
- Press `escape` → verify dismissed without side effects

### 6.6.2 Dialog cancel button
- Open confirmation dialogs
- Navigate to Cancel (tab/arrow), press enter
- Verify action NOT executed

### 6.6.3 Form field navigation
- Open a multi-field form (e.g., CTA with 4 fields)
- Tab through all fields → verify focus moves forward sequentially
- Shift+tab → verify focus moves backward
- Verify text input preserved when moving between fields

### 6.6.4 Form with select field
- Open a form with a select field (e.g., Button variant)
- Navigate to select field
- Verify up/down changes selected option
- Verify selected value shown in field
- Submit → verify selected value stored

### 6.6.5 Form with boolean field
- Open a form with boolean (e.g., Row full_width)
- Navigate to boolean field
- Toggle value (left/right or space)
- Verify visual indicator changes
- Submit → verify value stored

### 6.6.6 Form with textarea field
- Open a form with textarea (e.g., CTA subheading)
- Type multi-word text
- Verify text wraps in textarea display
- Submit → verify full text stored

### 6.6.7 Nested dialogs don't occur
- While a dialog is open, press `n` or `d`
- Verify new dialog does NOT open on top of existing one
- Existing dialog handles key or ignores it

## 6.7 Dead Keybindings (Audit Verification)

### 6.7.1 Verify unhandled keys don't crash
For each dead keybinding, press the key on every screen and verify no crash:
- `shift+left` (ActionTitlePrev) — press on Home, Content, Datatypes, etc.
- `shift+right` (ActionTitleNext) — same
- `F` (ActionScreenNext) — same
- `f` (ActionScreenToggle) — same
- `ctrl+x` (ActionAccordion) — same
- `[` (ActionTabPrev) — same
- `]` (ActionTabNext) — same

### 6.7.2 Document dead keys behavior
- For each: record whether key is silently ignored, triggers unexpected action, or crashes
- Recommend: either implement or remove from keybinding definitions

## 6.8 Snapshot Regression

Golden file snapshots for visual regression testing. First run creates goldens, subsequent runs diff.

### 6.8.1 Home screen golden
- Fresh state after wipe + quickstart install
- Snapshot: `testdata/tui_goldens/home.txt`

### 6.8.2 Content list golden
- With bootstrap + About Us page
- Snapshot: `testdata/tui_goldens/content_list.txt`

### 6.8.3 Content tree golden
- Home page tree with Row → CTA (filled fields)
- Snapshot: `testdata/tui_goldens/content_tree.txt`

### 6.8.4 Actions screen golden
- Default state, cursor on DB Init
- Snapshot: `testdata/tui_goldens/actions.txt`

### 6.8.5 Datatypes browse golden
- After Modula Default install
- Snapshot: `testdata/tui_goldens/datatypes_browse.txt`

### 6.8.6 Datatypes fields golden
- CTA datatype fields phase
- Snapshot: `testdata/tui_goldens/datatypes_fields.txt`

### 6.8.7 Routes golden
- Snapshot: `testdata/tui_goldens/routes.txt`

### 6.8.8 Users golden
- Snapshot: `testdata/tui_goldens/users.txt`

### 6.8.9 Config golden
- Snapshot: `testdata/tui_goldens/config.txt`

### 6.8.10 Database golden
- Datatypes table selected, showing rows
- Snapshot: `testdata/tui_goldens/database.txt`

### 6.8.11 Quickstart golden
- Snapshot: `testdata/tui_goldens/quickstart.txt`

## 6.9 Performance

### 6.9.1 Large content tree
- Create 20+ content nodes in a single tree (nested Rows with blocks)
- Verify tree renders without lag
- Verify cursor navigation responsive
- Verify Preview panel updates promptly

### 6.9.2 Large datatype count
- After Modula Default install (35+ datatypes)
- Verify datatype list renders fully
- Verify search/filter responsive

### 6.9.3 Database screen with many rows
- Select a table with 100+ rows
- Verify pagination works
- Verify page transitions are prompt
