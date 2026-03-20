# TUI Features

## Dialog Text Wrapping

Dialog messages automatically word-wrap to fit the dialog's inner width via `lipgloss.Style.Width(innerW)` in `dialog.go` `Render()`. Callers should use natural paragraph breaks (`\n\n`) for intentional spacing but do not need to insert manual line breaks for wrapping — lipgloss handles it.

## Connection Modes

Three distinct connection modes exposed via `Model.IsSSH` and `Model.IsRemote`:

| Mode | IsSSH | IsRemote | Source | File Picker |
|------|-------|----------|--------|-------------|
| local | false | false | `cmd/tui.go` | yes |
| SSH | true | false | `middleware.go` | no |
| remote | false | true | `cmd/connect.go` | no |

The home screen Site panel displays the current mode. Media upload (`n`) shows a help dialog when file picker is unavailable (SSH sessions). In remote mode, media uploads go through the SDK with progress tracking. Admin media upload uses the FILEPICKER_ADMINMEDIA purpose and follows the same connection mode constraints.

### Remote Mode Guards

The status bar shows `[remote: <url>]` (cyan) or `[remote: disconnected]` (red) via `panel_view.go`.

| Screen | Remote Support | Notes |
|--------|---------------|-------|
| Home | full | Displays mode, remote connection status |
| Content | full | CRUD via RemoteDriver |
| Media | full | Upload via SDK with progress; file picker opens locally |
| Admin Media | full | Upload via SDK with progress; file picker opens locally |
| Routes | full | CRUD via RemoteDriver |
| Datatypes | full | CRUD via RemoteDriver |
| Field Types | full | CRUD via RemoteDriver |
| Users | full | CRUD via RemoteDriver |
| Plugins | full | List/enable/disable via RemoteDriver |
| Pipelines | full | Read-only listing |
| Webhooks | full | CRUD via RemoteDriver |
| Config | full | Read/update via RemoteDriver |
| Deploy | full | Sync operations target remote environments |
| Quickstart | full | Read-only guide |
| Actions | limited | Only: Check for Updates, Validate Config, Generate API Token |
| Database | hidden | Removed from menu; `GetConnection()` returns `ErrRemoteMode` |

Guard locations:
- Menu filtering: `menus.go` (`HomepageMenuInit` excludes `DATABASEPAGE`)
- Actions filtering: `actions.go` (`ActionsMenuForMode`)
- Status bar: `panel_view.go` (remote indicator with connection health)
- Media upload: `commands_media.go` (SDK path when `isRemote`)
- Error sentinel: `internal/remote/errors.go` (`ErrRemoteMode`)

## Grid Layout System

Screens use a 12-column grid (`grid.go`, `grid_screen.go`) with proportional cell heights. `GridScreen` base struct provides focus cycling and rendering. Migrated screens: Home, Actions, Content, Admin Content, Media, Admin Media, Field Types, Validations, Users, Quickstart, CMS Menu, Deploy, Plugins, Plugin Detail, Plugin TUI, Webhooks, Config, Database, Datatypes, Pipelines, Pipeline Detail, Routes, Tokens, Sessions, Media Dimensions, Import, Roles, Audit, Search.

## Media Tree

Media items are grouped into a URL-path-derived folder tree (`media_tree.go`). Folders are collapsible, and an inline search (`/`) live-filters items by name, display name, mimetype, or URL path. Tree is rebuilt from filtered results so folder structure reflects only matches.

Folder management is supported via dedicated dialogs: folders can be created, renamed, and deleted, and media items can be moved between folders. `MediaFolderNameDialogModel` handles folder creation and renaming via a text input, while `MoveMediaFolderDialogModel` presents a folder selection list for relocating media.

## Admin Media

Admin media is a parallel media system for the admin panel (distinct from public media). The AdminMediaScreen (`screen_admin_media.go`) mirrors the MediaScreen with its own folder tree, upload flow (FILEPICKER_ADMINMEDIA purpose), and folder dialogs (`AdminMediaFolderNameDialogModel`, `MoveAdminMediaFolderDialogModel`). Admin media is accessible when AdminMode is enabled and uses admin-prefixed tables and API endpoints.

## Keybinding: ActionSearch

`config.ActionSearch` (default `/`) activates inline search on screens that support it. Currently used by Media screen.
