# TUI QA Phase 1: Foundation

Covers: Home screen, navigation system, Actions screen (12 actions), Quickstart schema installation.

## Setup

Every test begins with:
```
tui_start command=modula-x86 args=["tui"] cols=120 rows=40
tui_wait stable_ms=2000 timeout_ms=8000
```

## 1.1 Home Screen

### 1.1.1 Dashboard renders with real data
- Wait for `Home  [postgres]`
- Verify screen text contains: `Nav`, `Plugins`, `Activity`, `Connections`, `Backups`, `Site`
- Verify nav items: Content, Media, Routes, Datatypes, Field Types, Users, Plugins, Pipelines, Webhooks, Config, Deploy, Actions, Database, Quickstart (14 items)
- Verify Site panel shows: Mode, HTTP, SSL, SSH, CORS
- Verify Connections panel shows: DB postgres OK, S3 status, Remote status
- Snapshot: `goldens/1.1.1_home_default.txt`

### 1.1.2 Dashboard counts reflect DB state
- After DB Wipe & Redeploy (no schema): verify Stats show minimal counts
- After Quickstart install: navigate away and back, verify counts increase

### 1.1.3 Activity panel shows change events
- After DB Wipe & Redeploy: verify Activity panel shows recent INSERT events
- Events should show timestamp + operation + table name

### 1.1.4 Backups panel
- Fresh DB: verify "No backups" or empty state
- After creating a backup (Suite 2.8): verify backup appears

## 1.2 Navigation

### 1.2.1 Nav cursor movement — down
- From Content (top), send `down` 13 times
- After each press, `tui_screen` and verify `->` indicator moves to next item
- Verify cursor stops at Quickstart (bottom) — does NOT wrap

### 1.2.2 Nav cursor movement — up
- From Quickstart (bottom), send `up` 13 times
- Verify cursor stops at Content (top) — does NOT wrap

### 1.2.3 Enter each screen and verify header
For each of the 14 nav items:
- Navigate to item, press `enter`
- `tui_wait` for screen name in header (e.g., "Content  [postgres]", "Actions  [postgres]")
- `tui_screen` to capture
- Return to Home (press `h` — may trigger quit dialog on top-level screens; verify behavior)

### 1.2.4 Panel focus cycling
- On Home screen, press `tab`
- Verify focus changes (border style change or content highlight)
- Press `shift+tab` — verify reverse

### 1.2.5 Quit dialog
- Press `q` on Home screen
- Verify "Are you sure you want to quit?" dialog
- Verify Cancel is focused by default
- Press `enter` on Cancel → verify dialog dismissed, Home screen restored
- Press `q` again, navigate to Quit, press `enter` → verify TUI exits (tui_wait for exit)

### 1.2.6 Back navigation from sub-screens
- Enter Actions screen, press `h` → verify returns to Home (not quit dialog)
- Enter Content screen, press `h` → verify behavior (may show quit dialog since Content is top-level navigated from Home)
- Document actual back behavior for each screen

## 1.3 Actions Screen

### 1.3.1 Actions list renders
- Navigate to Actions, wait for stable
- Verify all 12 actions visible in list
- Verify System panel shows: Version, Commit, Built, Go, OS/Arch, DB, Env
- Verify Help panel shows keybinding hints
- Snapshot: `goldens/1.3.1_actions_default.txt`

### 1.3.2 Action detail updates on cursor movement
- Move cursor through each action (12 down presses)
- After each, verify Details panel shows correct description:
  - DB Init: "Create database tables and bootstrap data"
  - DB Wipe: mentions destructive/warning
  - DB Wipe & Redeploy: "Drop all tables, recreate schema, and bootstrap data"
  - DB Reset: mentions SQLite file deletion
  - DB Export: mentions SQL dump
  - Generate Certs: mentions self-signed/Let's Encrypt
  - Check for Updates: mentions version check
  - Validate Config: mentions configuration validation
  - Generate API Token: mentions API token
  - Register SSH Key: mentions SSH key registration
  - Create Backup: mentions backup creation
  - Restore Backup: mentions backup restore

### 1.3.3 DB Init
- Select DB Init, press `enter`
- Verify: NO confirmation dialog (non-destructive)
- Wait for result dialog
- Verify result shows success or "tables already exist" error

### 1.3.4 DB Wipe (confirmation + cancel)
- Select DB Wipe, press `enter`
- Verify confirmation dialog with WARNING text
- Navigate to Cancel, press `enter`
- Verify dialog dismissed, back on Actions — no wipe occurred

### 1.3.5 DB Wipe (confirmation + OK)
- Select DB Wipe, press `enter`
- Verify confirmation dialog
- Press `enter` on OK
- Wait for result dialog showing success
- Dismiss result

### 1.3.6 DB Wipe & Redeploy
- Select "DB Wipe & Redeploy", press `enter`
- Confirm OK
- Wait for result: verify shows "Redeploy Complete" with admin credentials
- Verify temporary password is displayed
- Dismiss result

### 1.3.7 DB Reset (SQLite-specific)
- Skip if not SQLite backend
- Select DB Reset, confirm
- Verify result dialog

### 1.3.8 DB Export
- Select DB Export, press `enter`
- Verify result (success with file path or error)
- No confirmation needed

### 1.3.9 Generate Certs
- Select "Generate Certs", press `enter`
- Verify result shows cert/key paths or error

### 1.3.10 Check for Updates
- Select "Check for Updates", press `enter`
- Wait for result (may take network time)
- Verify result shows version info

### 1.3.11 Validate Config
- Select "Validate Config", press `enter`
- Verify result shows validation pass/fail with details

### 1.3.12 Generate API Token
- Select "Generate API Token", press `enter`
- Verify result shows generated token string

### 1.3.13 Register SSH Key
- Select "Register SSH Key", press `enter`
- Verify result or error (may need SSH context)

### 1.3.14 Create Backup
- Select "Create Backup", press `enter`
- Verify result shows backup path and size

### 1.3.15 Restore Backup (with file picker)
- Select "Restore Backup", press `enter`
- Verify file picker opens or error if no backups exist
- If file picker: select a .zip, verify confirmation dialog
- Confirm → verify restore result

## 1.4 Quickstart

### 1.4.1 Schema list renders
- Navigate to Quickstart, wait for stable
- Verify 5 schemas listed: Modula Default, Contentful Starter, Sanity Starter, Strapi Starter, WordPress Blog
- Verify Details panel shows label, format, slug, description for selected
- Verify Info panel shows "Schemas: 5"
- Snapshot: `goldens/1.4.1_quickstart_default.txt`

### 1.4.2 Schema detail updates on cursor
- Move through all 5 schemas
- Verify Details panel updates for each (label, format change)

### 1.4.3 Install Modula Default
- Select Modula Default (first item), press `enter`
- Verify confirmation dialog: "Install Modula Default (modula-default)?"
- Press `enter` (Install focused)
- Wait for "Install Complete" text
- Verify: "Datatypes: 35" and "Fields: 121" (35 because bootstrap page reused)
- Dismiss result

### 1.4.4 Install cancel
- Select any schema, press `enter`
- Navigate to Cancel in confirmation dialog
- Press `enter` → verify dialog dismissed, no install

### 1.4.5 Bootstrap + schema dedup (regression for page datatype fix)
- After DB Wipe & Redeploy + Modula Default install
- Query DB: verify exactly ONE datatype with name="page" and type="_root"
- Verify child datatypes (Row, CTA, Rich Text, etc.) have parent_id pointing to that single page datatype
- This can be verified by entering Content → tree → pressing `n` → verifying Row, CTA, etc. appear in child type selector

### 1.4.6 Install other schemas
- DB Wipe & Redeploy
- Install Contentful Starter → verify success with datatype/field counts
- DB Wipe & Redeploy
- Install Sanity Starter → verify success
- DB Wipe & Redeploy
- Install Strapi Starter → verify success
- DB Wipe & Redeploy
- Install WordPress Blog → verify success
