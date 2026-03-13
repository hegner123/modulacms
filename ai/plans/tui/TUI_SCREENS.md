# TUI Screens Inventory

All screens live in `internal/tui/`. Each implements the `Screen` interface (`Update()` + `View()`) and extends `GridScreen` for 12-column responsive layout.

## Page Constants (`pages.go`)

| Constant | Value | Status |
|----------|-------|--------|
| HOMEPAGE | 0 | Active |
| CMSPAGE | 1 | Active |
| ADMINCMSPAGE | 2 | Active |
| DATABASEPAGE | 3 | Active |
| CONFIGPAGE | 4 | Active |
| READPAGE | 7 | Deprecated (redirects to DATABASEPAGE) |
| DATATYPES | 13 | Active |
| USERSADMIN | 18 | Active |
| MEDIA | 19 | Active |
| CONTENT | 20 | Active |
| ACTIONSPAGE | 23 | Active |
| ROUTES | 24 | Active |
| ADMINROUTES | 25 | Active |
| ADMINDATATYPES | 26 | Active |
| ADMINCONTENT | 27 | Active |
| PLUGINSPAGE | 28 | Active |
| PLUGINDETAILPAGE | 29 | Active |
| QUICKSTARTPAGE | 31 | Active |
| FIELDTYPES | 32 | Active |
| ADMINFIELDTYPES | 33 | Active |
| DEPLOYPAGE | 34 | Active |
| PIPELINESPAGE | 35 | Active |
| PIPELINEDETAILPAGE | 36 | Active |
| WEBHOOKSPAGE | 37 | Active |

## Global Controls

These apply on every screen:

- `ctrl+c` = Force quit immediately
- `ctrl+a` = Toggle admin/client CMS mode
- `q` = Quit
- `b` = Back to previous page/state
- `esc` = Dismiss current view
- `tab` = Next panel
- `shift+tab` = Previous panel
- `up` / `down` = Cursor up/down in focused list

## Dialog Controls

When a dialog is open:

- `enter` = Confirm / dismiss
- `esc` = Cancel / dismiss
- `tab` / `shift+tab` / `h` / `l` / `left` / `right` = Navigate between buttons
- `up` / `down` / `j` / `k` = Navigate between buttons or list items

## Active Screens (17)

### 1. HomeScreen

- **Page:** HOMEPAGE
- **Files:** `screen_home.go`
- **Layout:** 3-column grid (Nav/Site Config | Plugins/Connections | Activity/Backups)
- **Purpose:** Dashboard with quick stats, navigation menu, plugins, backups, recent activity. Menu selection, title font cycling, connection status display.
- **Controls:**
  - `up` / `down` = Navigate menu items
  - `enter` / `right` = Select highlighted menu item
  - Title prev/next = Cycle title font style

### 2. CMSMenuScreen

- **Page:** CMSPAGE, ADMINCMSPAGE
- **Files:** `screen_cms_menu.go`
- **Layout:** 3/9 grid (Navigation menu | Details/Info)
- **Purpose:** Navigation hub for CMS operations (Content, Datatypes, Routes, Media, Users, FieldTypes). Regular CMS vs Admin CMS variants.
- **Controls:**
  - `up` / `down` = Navigate menu items
  - `enter` = Navigate to selected page
  - Title prev/next = Cycle title font style

### 3. ContentScreen

- **Page:** CONTENT, ADMINCONTENT
- **Files:** `screen_content.go`, `screen_content_view.go`
- **Layout:** 2-column grid (Content tree/list | Details/Stats)
- **Purpose:** Tree-based content browser. Two phases: route selection then tree node browsing. Field editing, version history, batch field loading.
- **Controls (Select Phase - Route List):**
  - `up` / `down` = Navigate content routes
  - `+` = Expand group/node
  - `-` = Collapse group/node
  - `enter` = Enter tree browsing for selected route
  - `n` = Create new content route
  - `e` = Edit selected content
  - `d` = Delete selected content
  - `p` = Publish/unpublish content
- **Controls (Tree Phase - Node Browsing):**
  - `up` / `down` = Navigate visible tree nodes
  - `enter` = Toggle expand/collapse node
  - `+` = Expand selected node
  - `-` = Collapse selected node
  - `^` = Jump to parent node
  - `v` = Jump to first child node
  - `n` = Create new child content
  - `e` = Edit selected content node
  - `d` = Delete selected content node
  - `c` = Copy/duplicate selected content
  - `m` = Move content to different parent
  - `<` = Reorder up among siblings
  - `>` = Reorder down among siblings
  - `p` = Publish/unpublish content
  - `v` = View version history
  - `l` = Switch locale (if i18n enabled)
- **Controls (Version List):**
  - `up` / `down` = Navigate versions
  - `enter` = Restore selected version
  - `b` = Close version list

### 4. DatatypesScreen

- **Page:** DATATYPES, ADMINDATATYPES
- **Files:** `screen_datatypes.go`, `screen_datatypes_view.go`
- **Layout:** Phase-dependent (4/8 or 3/9 grid)
- **Purpose:** Schema definition browser. Two phases: datatype tree browsing then field editing. Search/filter, field reordering, field property preview, CRUD.
- **Controls (Phase 1 - Datatype Browse):**
  - `up` / `down` = Navigate datatypes in tree
  - `+` / `-` = Expand/collapse parent datatypes
  - `/` = Enter search mode (type to filter, enter to apply, esc to clear)
  - `enter` = Enter field selection (Phase 2)
  - `n` = Create new datatype
  - `e` = Edit selected datatype
  - `d` = Delete selected datatype
- **Controls (Phase 2 - Field Selection):**
  - `up` / `down` = Navigate fields list
  - `b` = Return to Phase 1

### 5. DatabaseScreen

- **Page:** DATABASEPAGE
- **Files:** `screen_database.go`
- **Layout:** 3/9 grid (Tables | Rows/Detail)
- **Purpose:** Raw database table browser for power users (local only, not remote). Table selection, row pagination, row detail/column viewing, create/edit/delete.
- **Controls (Tables panel):**
  - `up` / `down` = Navigate table list
  - `enter` = Select table, move to Rows panel
- **Controls (Rows panel):**
  - `up` / `down` = Navigate rows
  - `[` = Previous page
  - `]` = Next page
  - `enter` = Focus Detail panel
  - `n` = Insert new row
  - `e` = Edit selected row
  - `d` = Delete selected row
  - `b` = Back to Tables panel
- **Controls (Detail panel):**
  - `up` / `down` = Scroll through row columns
  - `n` = Insert new row
  - `e` = Edit selected row
  - `d` = Delete selected row
  - `b` = Back to Rows panel

### 6. MediaScreen

- **Page:** MEDIA
- **Files:** `screen_media.go`, `screen_media_view.go`
- **Layout:** 3/9 grid (Media tree | Summary/Metadata)
- **Purpose:** Media library with file tree and metadata management. Search/filter, media summary, metadata editing, focal point management.
- **Controls:**
  - `up` / `down` = Navigate media tree
  - `/` = Enter search mode (type to filter, enter to apply, esc to clear)
  - `enter` = Toggle expand/collapse folders
  - `n` = Upload new media (opens file picker, unavailable over SSH)
  - `d` = Delete selected media file

### 7. RoutesScreen

- **Page:** ROUTES, ADMINROUTES
- **Files:** `screen_routes.go`
- **Layout:** 3/6/3 grid (Routes | Details/Info | Actions/Stats)
- **Purpose:** URL route management. Route listing, CRUD, activation/deactivation. Regular vs Admin variants.
- **Controls:**
  - `up` / `down` = Navigate routes list
  - `enter` = Select route
  - `n` = Create new route
  - `e` = Edit selected route
  - `d` = Delete selected route

### 8. UsersScreen

- **Page:** USERSADMIN
- **Files:** `screen_users.go`
- **Layout:** 3/9 grid (User list | Details/Permissions)
- **Purpose:** User and role management. User listing with roles, permission viewing, user CRUD, role assignment.
- **Controls:**
  - `up` / `down` = Navigate user list
  - `n` = Create new user
  - `e` = Edit selected user
  - `d` = Delete selected user

### 9. FieldTypesScreen

- **Page:** FIELDTYPES, ADMINFIELDTYPES
- **Files:** `screen_field_types.go`
- **Layout:** 3/9 grid (Field types list | Details/Info)
- **Purpose:** Field type definition browser. Regular vs Admin variants. Field type listing, details, property viewing.
- **Controls:**
  - `up` / `down` = Navigate field types
  - `n` = Create new field type
  - `e` = Edit selected field type
  - `d` = Delete selected field type

### 10. PluginsScreen

- **Page:** PLUGINSPAGE
- **Files:** `screen_plugins.go`
- **Layout:** 3/9 grid (Plugin list | Details/Info)
- **Purpose:** Plugin listing and status overview. Status display (active/inactive), navigation to detail, enable/disable.
- **Controls:**
  - `up` / `down` = Navigate plugins list
  - `enter` = Navigate to plugin detail page

### 11. PluginDetailScreen

- **Page:** PLUGINDETAILPAGE
- **Files:** `screen_plugin_detail.go`
- **Layout:** 3/9 grid (Plugin info | Actions/Info)
- **Purpose:** Detailed plugin information and management. Manifest and capability display, enable/disable/reload, drift detection.
- **Controls (Plugin Info panel):**
  - `b` = Back to plugins list
- **Controls (Actions panel):**
  - `up` / `down` = Navigate plugin actions (enable, disable, reload)
  - `enter` = Execute selected action
  - `b` = Back to plugins list
- **Controls (Action Info panel):**
  - `enter` = Execute highlighted action
  - `b` = Return to Actions panel

### 12. ConfigScreen

- **Page:** CONFIGPAGE
- **Files:** `screen_config.go`
- **Layout:** 3/9 grid (Categories | Fields/Detail)
- **Purpose:** Configuration management by category. Field editing, raw JSON view mode with viewport scrolling, hot-reload support.
- **Controls (Categories panel):**
  - `up` / `down` = Navigate config categories
  - `enter` = Enter selected category
- **Controls (Fields panel):**
  - `up` / `down` = Navigate fields in category
  - `e` / `enter` = Edit selected field
  - `b` = Return to Categories
- **Controls (Raw JSON view):**
  - `up` / `down` / `pgup` / `pgdn` = Scroll JSON content
  - `b` = Return to Categories

### 13. DeployScreen

- **Page:** DEPLOYPAGE
- **Files:** `screen_deploy.go`
- **Layout:** 3/9 grid (Environments | Details/Actions)
- **Purpose:** Multi-environment deployment management. Environment listing, pull/push sync, health checks, last result tracking.
- **Controls:**
  - `up` / `down` = Navigate environments
  - `t` = Test connection to selected environment
  - `p` = Pull dry-run from selected environment
  - `P` = Pull commit from selected environment
  - `s` = Push dry-run to selected environment
  - `S` = Push commit to selected environment
  - All action keys blocked while an operation is active

### 14. PipelinesScreen

- **Page:** PIPELINESPAGE
- **Files:** `screen_pipelines.go`
- **Layout:** 3/6/3 grid (Chains | Chain Info/Help | Registry/By Table)
- **Purpose:** Plugin pipeline chain listing. Chain browsing, navigation to detail, registry view by plugin, view by table.
- **Controls:**
  - `up` / `down` = Navigate pipeline chains
  - `enter` = Navigate to pipeline detail page

### 15. PipelineDetailScreen

- **Page:** PIPELINEDETAILPAGE
- **Files:** (within pipelines implementation)
- **Purpose:** Individual pipeline chain detail view.
- **Controls:**
  - `up` / `down` = Navigate chain steps
  - `b` = Back to pipelines list

### 16. ActionsScreen

- **Page:** ACTIONSPAGE
- **Files:** `screen_actions.go`
- **Layout:** 3/6/3 grid (Actions | Details/Help | System/Updates)
- **Purpose:** System actions and utilities. Backup/restore operations, version/update checking, system info display.
- **Controls:**
  - `up` / `down` = Navigate actions list
  - `enter` = Run selected action (some require confirmation)

### 17. QuickstartScreen

- **Page:** QUICKSTARTPAGE
- **Files:** `screen_quickstart.go`
- **Layout:** 3/9 grid (Schema list | Details/Info)
- **Purpose:** Pre-configured schema templates for bootstrapping. Template listing, details, import.
- **Controls:**
  - `up` / `down` = Navigate schema definitions
  - `enter` = Install selected schema

### 18. WebhooksScreen

- **Page:** WEBHOOKSPAGE
- **Files:** `screen_webhooks.go`
- **Layout:** 3/9 grid (Webhooks | Details/Info)
- **Purpose:** Webhook management. Webhook listing, status display, CRUD operations.
- **Controls:**
  - `up` / `down` = Navigate webhooks list
  - Display-only, no action keys

## Admin Mode Toggle

Ctrl+A toggles admin mode globally. `AdminPageIndex()` in `menus.go` maps:

| Client Page | Admin Page |
|-------------|------------|
| CONTENT | ADMINCONTENT |
| DATATYPES | ADMINDATATYPES |
| ROUTES | ADMINROUTES |
| FIELDTYPES | ADMINFIELDTYPES |

Pages without admin variants (Media, Users, etc.) remain unchanged.

## Navigation Structure

| Category | Screens |
|----------|---------|
| Daily Workflow | Content, Media, Routes |
| Schema/Structure | Datatypes, FieldTypes, Users |
| System | Plugins, Pipelines, Webhooks, Config, Deploy |
| Power User | Actions, Database (local only), Quickstart |

## Infrastructure Files

| File | Purpose |
|------|---------|
| `screen.go` | Screen interface and `screenForPage()` factory |
| `pages.go` | Page constants and registry |
| `menus.go` | Menu initialization and AdminPageIndex |
| `model.go` | Root Model struct, ActiveScreen field |
| `grid.go` | Grid layout model and calculations |
| `grid_screen.go` | GridScreen base (12-column layout, focus, cursor, scroll) |
| `update_navigation.go` | Navigation message routing |
| `page_builders.go` | Centralized grid rendering |
