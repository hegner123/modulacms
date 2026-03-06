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

The home screen Site panel displays the current mode. Media upload (`n`) shows a help dialog when file picker is unavailable (SSH sessions).

## Grid Layout System

Screens use a 12-column grid (`grid.go`, `grid_screen.go`) with proportional cell heights. `GridScreen` base struct provides focus cycling and rendering. Migrated screens: Home, Actions, Content, Admin Content, Media.

## Media Tree

Media items are grouped into a URL-path-derived folder tree (`media_tree.go`). Folders are collapsible, and an inline search (`/`) live-filters items by name, display name, mimetype, or URL path. Tree is rebuilt from filtered results so folder structure reflects only matches.

## Keybinding: ActionSearch

`config.ActionSearch` (default `/`) activates inline search on screens that support it. Currently used by Media screen.
