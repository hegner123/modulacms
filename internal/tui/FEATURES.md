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

The home screen Site panel displays the current mode. Media upload (`n`) shows a help dialog when file picker is unavailable (SSH sessions). In remote mode, media uploads go through the SDK with progress tracking.

### Remote Mode Guards

The status bar shows `[remote: <url>]` (cyan) or `[remote: disconnected]` (red) via `panel_view.go`.

| Screen | Remote Support | Notes |
|--------|---------------|-------|
| Home | full | Displays mode, remote connection status |
| Content | full | CRUD via RemoteDriver |
| Media | full | Upload via SDK with progress; file picker opens locally |
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

Screens use a 12-column grid (`grid.go`, `grid_screen.go`) with proportional cell heights. `GridScreen` base struct provides focus cycling and rendering. Migrated screens: Home, Actions, Content, Admin Content, Media, Field Types, Users, Quickstart, CMS Menu, Deploy, Plugins, Plugin Detail, Webhooks, Config, Database, Datatypes.

## Media Tree

Media items are grouped into a URL-path-derived folder tree (`media_tree.go`). Folders are collapsible, and an inline search (`/`) live-filters items by name, display name, mimetype, or URL path. Tree is rebuilt from filtered results so folder structure reflects only matches.

## Keybinding: ActionSearch

`config.ActionSearch` (default `/`) activates inline search on screens that support it. Currently used by Media screen.
