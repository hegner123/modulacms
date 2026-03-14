# TUI QA Phase 4: Administration

Covers: Users screen (CRUD + permissions), Config screen (browse + edit), Database screen (generic CRUD + pagination), Media screen (upload + delete + search).

## Prerequisites

- DB Wipe & Redeploy + Modula Default schema installed

## 4.1 Users Screen

### 4.1.1 User list renders
- Navigate to Users, wait for stable
- Verify at minimum "System Administrator" with admin role
- Verify Details panel shows: username, name, email, role, dates
- Verify permissions panel shows grouped permissions for selected user's role
- Snapshot: `goldens/4.1.1_users.txt`

### 4.1.2 Permission display per role
- Select system admin user
- Verify permissions panel shows all admin permissions grouped by resource
- Verify lazy-loaded permissions update when cursor moves between users of different roles

### 4.1.3 Create user
- Press `n` → verify user creation form
- Verify fields: Username, Name, Email, Password, Role (selector)
- Verify Role selector shows: admin, editor, viewer
- Fill all fields, select "editor" role, submit
- Verify new user appears in list with "editor" role

### 4.1.4 Edit user
- Select a user, press `e`
- Verify edit form pre-populated (password field may be empty for security)
- Modify name, confirm
- Verify updated name in list

### 4.1.5 Delete user
- Select a non-system user, press `d`
- Verify confirmation dialog
- Confirm → verify user removed
- Attempt to delete system admin → verify if blocked (system-protected)

### 4.1.6 Role-based permission differences
- Create users with admin, editor, viewer roles
- Navigate between them
- Verify permission counts differ:
  - Admin: all permissions
  - Editor: ~28 permissions (CRUD on content/datatypes/fields/media/routes)
  - Viewer: ~5 permissions (read-only)

## 4.2 Config Screen

### 4.2.1 Config categories render
- Navigate to Config, wait for stable
- Verify category list (left panel): General, Database, SSL, SSH, S3, CORS, OAuth, Plugins, Observability, etc. + "View Raw JSON"
- Verify first category auto-selected with fields visible
- Snapshot: `goldens/4.2.1_config.txt`

### 4.2.2 Browse categories
- Move cursor through all categories
- Verify fields panel updates with relevant config fields for each category
- Verify field values match modula.config.json

### 4.2.3 Field detail
- Select a specific field within a category
- Verify detail panel shows: description, example value, hot-reload status

### 4.2.4 Edit config field
- Select an editable field, press `e` or `enter`
- Verify edit dialog with current value
- Modify value, confirm
- Verify updated value displayed
- Verify if hot-reloadable fields take effect without restart

### 4.2.5 Sensitive field masking
- Navigate to fields that contain passwords/tokens/secrets
- Verify values are masked (not displayed in plain text)

### 4.2.6 View Raw JSON
- Navigate to "View Raw JSON" option, press `enter`
- Verify raw JSON viewport opens
- Verify JSON content matches config
- Verify scrolling works (up/down or viewport controls)
- Press `h` or `backspace` to return to category list

## 4.3 Database Screen

### 4.3.1 Table list renders
- Navigate to Database, wait for stable
- Verify left panel shows all database tables
- Verify at minimum: users, roles, permissions, datatypes, fields, routes, content_data, content_field, etc.
- Snapshot: `goldens/4.3.1_database.txt`

### 4.3.2 Select table and view rows
- Select "datatypes" table, press `enter`
- Verify column headers render (datatype_id, name, label, type, etc.)
- Verify rows render with actual data
- Verify pagination indicator (page X of Y)

### 4.3.3 Pagination
- Select a table with many rows (e.g., fields — 121+ after quickstart install)
- Verify page navigation: next page / prev page keys
- Verify page indicator updates
- Verify boundary: can't go past last page or before first page

### 4.3.4 Row detail view
- Select a row, press `enter`
- Verify detail view shows all columns as key-value pairs
- Verify scrolling for tables with many columns

### 4.3.5 Insert row
- Press `n` → verify insert form with fields for all table columns
- Fill fields with valid data
- Submit → verify row appears in table
- Test with invalid data → verify error handling

### 4.3.6 Edit row
- Select a row, press `e`
- Verify edit form pre-populated with row data
- Modify a field, confirm
- Verify updated value in table

### 4.3.7 Delete row
- Select a row, press `d`
- Verify confirmation dialog
- Confirm → verify row removed
- Test with row that has foreign key references → verify error or cascade behavior

### 4.3.8 Back to table list
- From row view, press `h` → verify return to table list

## 4.4 Media Screen

### 4.4.1 Media list renders
- Navigate to Media, wait for stable
- Verify tree structure or "No media" empty state
- Verify metadata panel (right side) shows info when media selected
- Snapshot: `goldens/4.4.1_media.txt`

### 4.4.2 Empty state
- After fresh DB wipe (no uploads): verify appropriate empty message

### 4.4.3 Search/Filter
- Press `/` → verify search input activates
- Type filter text → verify list filters in real time (case-insensitive substring)
- Press `escape` → verify filter cleared

### 4.4.4 Upload media (if SSH context available)
- Press `n` → verify file picker opens or SSH-not-available warning
- If file picker: select a file
- Verify upload progress/result
- Verify new media appears in list with metadata

### 4.4.5 Delete media
- Select media item, press `d`
- Verify confirmation dialog
- Confirm → verify media removed from list

### 4.4.6 Media metadata display
- Select uploaded media
- Verify metadata panel shows: name, mimetype, URL, srcset, alt text, caption, focal points, author, dates
- Note: NO edit operation for metadata in TUI (known gap)

### 4.4.7 Folder tree navigation
- If media organized in folders: expand/collapse folders
- Verify folder grouping works
- Verify navigation into/out of folders
