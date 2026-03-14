# tui

The tui package implements a Bubbletea v2-based terminal user interface for ModulaCMS. It provides an SSH-accessible TUI with a Screen-based architecture for managing content, datatypes, media, routes, users, plugins, webhooks, and system configuration. The package follows the Elm Architecture pattern (Model-Update-View) with typed message flows and database abstraction.

## Overview

This package integrates with the Charmbracelet ecosystem (charm.land/bubbletea/v2, lipgloss/v2, bubbles/v2, huh/v2) to provide a full-featured TUI over SSH or direct terminal access. The architecture uses a root Model that dispatches messages to self-contained Screen implementations via the Screen interface.

Key architectural concepts:
- **Screen interface** — each page is a self-contained struct that owns its state, update logic, and rendering
- **AppContext** — read-only snapshot of shared state passed to screens (DB, Config, UserID, dimensions, etc.)
- **PanelScreen** — base type for multi-panel layouts with grid-based positioning
- **Modal overlays** — dialogs, form dialogs, content forms rendered as composited layers

## Types

### Model

Model is the top-level Bubbletea model implementing tea.Model. Contains all application state including database driver, configuration, UI state, dialog models, active screen, and user session data.

Key fields: DB (db.DbDriver), Config (*config.Config), ActiveScreen (Screen), Page (Page), Width/Height (terminal dimensions), PanelFocus (int), Focus (FocusKey), various dialog/form states.

### Screen

Interface for self-contained TUI pages. Defined in screen.go:

```go
type Screen interface {
    Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd)
    View(ctx AppContext) string
    PageIndex() PageIndex
}
```

### AppContext

Read-only snapshot of shared application state passed to Screen implementations. Contains DB, Config, Logger, UserID, Width, Height, ScreenMode, PluginManager, IsRemote, IsSSH, and other shared state. Built via Model.AppCtx().

### FocusKey

Identifies which UI component has keyboard focus: PAGEFOCUS, TABLEFOCUS, FORMFOCUS, DIALOGFOCUS.

### ApplicationState

Represents operational state: OK, EDITING, DELETING, WARN, ERROR.

### KeyHint / KeyHinter

KeyHint holds a key:label pair for the statusbar. Screens can implement the KeyHinter interface to provide context-aware statusbar hints.

## Functions

### CliRun

```go
func CliRun(m *Model) (*tea.Program, bool)
```

Creates and runs a tea.Program. Returns the CliContinue flag.

### InitialModel

Creates the initial Model with provided verbose flag, config, database driver, logger, and plugin manager. Sets up paginator, spinner, viewport, page map, form model, focus state, and panel focus. Returns Model and initial command.

### Init

Implements tea.Model Init. Returns batch of initial commands including table fetch and optional dashboard/spinner commands.

### Update

Main update function. Routes messages through specialized update handlers: UpdateProvisioning, UpdateLog, UpdateTea, UpdateState, UpdateNavigation, UpdateFetch, UpdateForm, UpdateDialog, UpdateCms, and admin variants. Dispatches to ActiveScreen.Update() for screen-specific logic.

### View

```go
func (m Model) View() tea.View
```

Main view function. Returns tea.View (Bubbletea v2). Shows provisioning form if needed, otherwise renders the panel layout via ActiveScreen.View() with header, statusbar, and modal overlays.

## Screen Implementations

Each screen_*.go file implements the Screen interface:

| File | Screen | Purpose |
|------|--------|---------|
| screen_home.go | HomeScreen | Dashboard with stats |
| screen_content.go | ContentScreen | Content tree + field editing |
| screen_content_view.go | ContentViewScreen | Content detail view |
| screen_datatypes.go | DatatypesScreen | Datatype management |
| screen_datatypes_view.go | DatatypesViewScreen | Datatype detail view |
| screen_media.go | MediaScreen | Media library |
| screen_media_view.go | MediaViewScreen | Media detail view |
| screen_routes.go | RoutesScreen | Route management |
| screen_users.go | UsersScreen | User management |
| screen_plugins.go | PluginsScreen | Plugin listing |
| screen_plugin_detail.go | PluginDetailScreen | Plugin detail/config |
| screen_webhooks.go | WebhooksScreen | Webhook management |
| screen_config.go | ConfigScreen | Configuration |
| screen_database.go | DatabaseScreen | Direct DB operations |
| screen_deploy.go | DeployScreen | Deploy/sync operations |
| screen_actions.go | ActionsScreen | System actions menu |
| screen_cms_menu.go | CmsMenuScreen | CMS navigation menu |
| screen_field_types.go | FieldTypesScreen | Field type management |
| screen_pipelines.go | PipelinesScreen | Pipeline management |
| screen_quickstart.go | QuickstartScreen | First-run guide |
| screen_plugin_tui.go | PluginTUIScreen | Standalone plugin UI via coroutine bridge |

## Dialogs

### DialogModel

Modal confirmation dialog with title, message, OK/Cancel buttons. Created via NewDialog().

### FormDialogModel

Form dialog with text inputs and selector carousels for datatypes, fields, routes.

### ContentFormDialogModel

Dynamic content form with fields generated from datatype definitions.

### UserFormDialogModel

User CRUD form with username, name, email, role fields.

## Overlay Compositing

Dialogs render as overlays composited on top of the base screen content using Composite(base, overlay). The modal_overlay.go handles centering and layer composition.

## Static Variables

### CliContinue

Global bool flag indicating whether CLI should continue running after exit.

### TitleFile

Embedded filesystem containing ASCII art title files.
