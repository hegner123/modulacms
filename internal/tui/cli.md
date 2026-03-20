# cli

The tui package implements the Terminal User Interface for ModulaCMS using Charmbracelet Bubbletea v2 following the Elm Architecture pattern. It provides interactive management of the CMS through SSH or direct terminal access.

## Overview

This package implements a full-featured TUI for ModulaCMS content management. The architecture uses a root Model with self-contained Screen implementations for each page. Screens receive an AppContext snapshot (read-only shared state) and return commands -- they never mutate the root Model directly.

## Core Model

### Model

Main application state struct implementing tea.Model. Contains database driver, configuration, logger, plugin manager, UI state (dimensions, focus, loading), active screen, dialog/form states, and user session data.

### Screen

Interface for self-contained page implementations. Each screen owns its state, update logic, and rendering:

```go
type Screen interface {
    Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd)
    View(ctx AppContext) string
    PageIndex() PageIndex
}
```

### AppContext

Read-only snapshot of shared application state passed to screens. Contains DB, Config, Logger, UserID, Width, Height, ScreenMode, PluginManager, ConfigManager, IsRemote, IsSSH, SSHFingerprint, SSHKeyType, SSHPublicKey, ActiveLocale, AccordionEnabled, AdminMode, and ActiveAccent. Built via Model.AppCtx().

### FocusKey

Type for focus tracking. Values: PAGEFOCUS, TABLEFOCUS, FORMFOCUS, DIALOGFOCUS.

### ApplicationState

Type for application status. Values: OK, EDITING, DELETING, WARN, ERROR.

### FilePickerPurpose

Purpose enumeration for file picker. Values: FILEPICKER_MEDIA for media upload, FILEPICKER_RESTORE for backup restore, FILEPICKER_IMPORT for data import, FILEPICKER_ADMINMEDIA for admin media upload.

## Initialization and Lifecycle

### CliRun

Runs the CLI program with Bubbletea tea.Program. Takes model pointer, creates program, runs it, and returns continuation flag CliContinue.

### InitialModel

Creates initial Model with provided verbose flag, config, database driver, logger, and plugin manager. Sets up paginator, spinner, viewport, page map, form model, focus state, and panel focus. Returns Model and initial command.

### Init

Implements Bubbletea Init method. Returns batch of initial commands including table fetch and optional dashboard/spinner commands.

### Update

Main update function. Routes messages through specialized update handlers in order: UpdateProvisioning, UpdateLog, UpdateTea, UpdateState, UpdateNavigation, UpdateFetch, UpdateForm, UpdateDialog, UpdateCms, and admin variants. Dispatches to ActiveScreen.Update() for screen-specific messages.

### View

```go
func (m Model) View() tea.View
```

Returns tea.View (Bubbletea v2). Shows user provisioning form if needed. Otherwise renders panel layout with header, active screen content, and statusbar. Overlays file picker, dialogs, and form dialogs when active.

## Screen Implementations

Screens are defined in screen_*.go files. Each implements the Screen interface and typically embeds GridScreen for multi-panel layouts. Screens include: HomeScreen, ContentScreen, DatatypesScreen, MediaScreen, AdminMediaScreen, RoutesScreen, UsersScreen, PluginsScreen, PluginDetailScreen, PluginTUIScreen, WebhooksScreen, ConfigScreen, DatabaseScreen, DeployScreen, ActionsScreen, FieldTypesScreen, ValidationsScreen, PipelinesScreen, PipelineDetailScreen, QuickstartScreen, CMSMenuScreen, TokensScreen, SessionsScreen, MediaDimensionsScreen, ImportScreen, RolesScreen, AuditScreen, and SearchScreen. View helper files (screen_content_view.go, screen_datatypes_view.go, screen_media_view.go, screen_admin_media_view.go) provide rendering methods for their parent screens.

## Page Management

### PageIndex

Enumeration of page types: HOMEPAGE, CMSPAGE, ADMINCMSPAGE, DATABASEPAGE, CONFIGPAGE, READPAGE, DATATYPES, USERSADMIN, MEDIA, CONTENT, ACTIONSPAGE, ROUTES, ADMINROUTES, ADMINDATATYPES, ADMINCONTENT, PLUGINSPAGE, PLUGINDETAILPAGE, QUICKSTARTPAGE, FIELDTYPES, ADMINFIELDTYPES, DEPLOYPAGE, PIPELINESPAGE, PIPELINEDETAILPAGE, WEBHOOKSPAGE, PLUGINTUIPAGE, VALIDATIONS, ADMINVALIDATIONS, TOKENSPAGE, SESSIONSPAGE, MEDIADIMENSIONSPAGE, IMPORTPAGE, ROLESPAGE, AUDITPAGE, SEARCHPAGE, ADMINMEDIA.

### Page

Represents a navigable page with index and label.

## Dialogs

- **DialogModel** — Modal confirmation with OK/Cancel buttons
- **FormDialogModel** — Form dialog with text inputs and selector carousels
- **ContentFormDialogModel** — Dynamic content form from datatype field definitions
- **UserFormDialogModel** — User CRUD form
- **WebhookFormDialogModel** — Webhook CRUD form
- **TokenFormDialogModel** — API token CRUD form
- **RoleFormDialogModel** — Role CRUD form
- **MediaFolderNameDialogModel** — Text input dialog for creating/renaming media folders
- **MoveMediaFolderDialogModel** — Selection dialog for moving media to a different folder
- **AdminMediaFolderNameDialogModel** — Text input dialog for creating/renaming admin media folders
- **MoveAdminMediaFolderDialogModel** — Selection dialog for moving admin media to a different folder
- **MediaDimensionFormDialogModel** — Media dimension CRUD form
- **UIConfigFormDialogModel** — UI configuration form
- **DatabaseFormDialogModel** — Database insert/update form

## Command Constructors

The package exports command constructor functions following the pattern FunctionNameCmd that return tea.Cmd. Categories include logging, state management, pagination, focus control, form management, history management, database operations, dialog management, CMS operations, route operations, datatype operations, field operations, media operations, user operations, backup operations, webhook operations, plugin operations, and content operations.

## Message Types

Over 150 message types for the Bubbletea message-passing pattern. Key categories: CRUD result messages (Created/Updated/Deleted), fetch trigger/result messages, dialog show/accept/cancel messages, navigation messages, and state update messages.
