# Plugin UI via Coroutine Bridge

## Overview

Plugins render UIs using gopher-lua coroutines bridged to Bubbletea's Update/View cycle. Go defines a fixed set of UI primitives. Lua plugins compose these primitives into layouts, yield them back to Go for rendering, and receive events on resume. The plugin never touches Bubbletea directly — it speaks in layout tables and receives structured events.

There are two kinds of plugin UI:

1. **Screens** — Standalone pages in the TUI sidebar. Full-screen, long-lived. Used for plugin-specific workflows (task manager, analytics dashboard, custom admin).
2. **Field Interfaces** — Editors for content fields of type `plugin`. Embedded inline in the content form dialog or opened as an overlay. Used to extend the CMS's field type system (color picker, location selector, markdown preview, widget configurator).

Both use the same coroutine bridge, the same primitives, and the same event/action protocol. They differ in lifecycle and hosting context.

```
Bubbletea                              │ gopher-lua coroutine
───────────────────────────────────────┼──────────────────────────────────────
tea.Msg arrives                        │ Go calls L.Resume(thread, eventTable)
Update() runs                          │ Lua processes event, updates state
Update() returns model                 │ Lua calls coroutine.yield(layoutTable)
View() renders                         │ Go reads yielded layout → Bubbletea
Next tea.Msg                           │ Next L.Resume()
```

The yield/resume cycle is in-process stack manipulation — microseconds, no IPC, no serialization. Yield is literally "swap a pointer and return -1."

## Design Principles

1. **Primitives, not widgets** — Go defines atomic UI building blocks. Plugins compose them. Go controls rendering.
2. **Yield is the only exit** — Plugin code runs synchronously within Resume. The only way to return control to Go is `coroutine.yield()`. No callbacks, no event listeners.
3. **Events are values, not handlers** — Go converts `tea.Msg` into a Lua table and passes it on Resume. The plugin processes it in a loop.
4. **Actions are yield payloads** — Navigation, dialogs, toasts, async requests, and value commits are special yield values distinguished by an `action` field. Layout yields have no `action` field.
5. **Single goroutine** — The plugin coroutine runs on the Bubbletea update goroutine. No LState thread safety issues. Async work uses the glua-async pattern.

## Plugin Manifest

```lua
plugin_info = {
    name = "my_plugin",
    version = "1.0.0",
    description = "Example plugin",

    -- Standalone TUI pages
    screens = {
        { name = "main", label = "My Plugin", icon = "list" },
        { name = "edit", label = "Edit Item", hidden = true },
    },

    -- Field-level editors (used by fields of type "plugin")
    interfaces = {
        { name = "picker", label = "Color Picker", mode = "overlay" },
        { name = "swatch", label = "Color Swatch", mode = "inline" },
    },
}
```

### Screen Entries

- `name` — Identifier, maps to `screens/<name>.lua`
- `label` — Display name in TUI sidebar
- `hidden` — If true, not shown in nav (reachable only via `navigate` action)

### Interface Entries

- `name` — Identifier, maps to `interfaces/<name>.lua`
- `label` — Display name shown in field type selector
- `mode` — `"inline"` or `"overlay"`
  - **inline**: Renders within the field's row in the content form dialog. Constrained to field width and 1-5 lines height. Receives keys only when field is focused.
  - **overlay**: Renders as a full-screen modal opened by pressing enter on the field. Full terminal dimensions. Captures all keys until dismissed.

### Plugin Directory Structure

```
my_plugin/
  init.lua                    # Manifest + hooks + routes
  screens/
    main.lua                  # function screen(ctx) ... end
    edit.lua
  interfaces/
    picker.lua                # function interface(ctx) ... end
    swatch.lua
  lib/
    helpers.lua               # Shared modules
```

## Field Type: `plugin`

A new field type `plugin` is added to `FieldType` enum. When a CMS operator creates a field with type `plugin`, a companion row in the `field_plugin_config` extension table binds it to a specific plugin and interface.

### Extension Table: `field_plugin_config`

```sql
-- sql/schema/XX_field_plugin_config/schema.sql
CREATE TABLE field_plugin_config (
    field_id         TEXT PRIMARY KEY NOT NULL
        REFERENCES fields ON DELETE CASCADE,
    plugin_name      TEXT NOT NULL,
    plugin_interface TEXT NOT NULL,
    plugin_version   TEXT NOT NULL DEFAULT '',
    date_created     TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Dual content schema: admin fields get their own table
CREATE TABLE admin_field_plugin_config (
    field_id         TEXT PRIMARY KEY NOT NULL
        REFERENCES admin_fields ON DELETE CASCADE,
    plugin_name      TEXT NOT NULL,
    plugin_interface TEXT NOT NULL,
    plugin_version   TEXT NOT NULL DEFAULT '',
    date_created     TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified    TEXT DEFAULT CURRENT_TIMESTAMP
);
```

This follows the same extension-table pattern as `plugin_routes`, `plugin_hooks`, and `plugin_requests` — separate table with approval/config data, FK back to the parent entity, CASCADE on delete.

**Why not the `data` JSON column?** The `data` TEXT column on `fields` stores field-type-specific config as opaque JSON (e.g., min/max for numbers, options for selects). Using it for plugin identity has three problems: (1) JSON extraction syntax differs across SQLite/MySQL/PostgreSQL, making queries unreliable in the tri-database pattern; (2) `plugin_name` and `plugin_interface` are identity — they define what the field IS, not how it's configured; (3) a dedicated table is queryable with `WHERE plugin_name = ?` for listing all fields bound to a plugin.

**`plugin_version`**: Stores the plugin version at the time the field was created. Used for drift detection — if the plugin upgrades and removes the interface, we can surface a warning instead of silently breaking.

The `field_value` in `content_fields` stores whatever string the plugin interface produces — a hex color, a JSON blob, a ULID, whatever the plugin decides. The CMS treats it as an opaque string, same as all other field types.

### Validation

When creating/updating a field with type `plugin`:
1. Read the `field_plugin_config` row (or create it if new)
2. Verify the named plugin exists and is enabled
3. Verify the plugin declares the named interface in its manifest
4. Verify the interface mode matches any constraints (inline height, etc.)
5. Store the field definition and extension row atomically

When the plugin is disabled or uninstalled, fields referencing it still exist but their editors show a "Plugin unavailable" message instead of the interface. The `field_plugin_config` row is preserved — re-enabling the plugin restores functionality without reconfiguration.

## Primitive Inventory

These are the Lua table shapes that Go recognizes in yielded values. Each maps to an existing TUI component.

### Layout Container (Screens + Overlay Interfaces)

The top-level yield for screens and overlay interfaces is a **grid**: an array of columns with cells.

```lua
{
    type = "grid",
    columns = {
        { span = 3, cells = {
            { title = "Items", height = 1.0, content = list_primitive },
        }},
        { span = 9, cells = {
            { title = "Detail", height = 0.6, content = detail_primitive },
            { title = "Info",   height = 0.4, content = text_primitive },
        }},
    },
    hints = {
        { key = "n", label = "new" },
        { key = "q", label = "back" },
    },
}
```

Maps to: `Grid`, `GridColumn`, `GridCell`, `CellContent` in `internal/tui/grid.go`.

### Inline Container (Inline Interfaces)

Inline interfaces yield a single primitive directly — no grid wrapper. The primitive renders within the field bubble's allocated space.

```lua
-- Inline interface yields a single primitive
coroutine.yield({
    type = "text",
    lines = {
        { text = "#ff0000", accent = true },
        { text = "██████████", style = { fg = "#ff0000" } },
    },
})
```

### Primitives

#### `list` — Vertical item list with cursor

```lua
{
    type = "list",
    items = {
        { label = "First Item", id = "item_1" },
        { label = "Second Item", id = "item_2", faint = true },
        { label = "Third Item", id = "item_3", bold = true },
    },
    cursor = 0,
    empty_text = "(none)",
}
```

#### `detail` — Key-value pair display

```lua
{
    type = "detail",
    title = "Item Details",
    fields = {
        { label = "Name",     value = "My Item" },
        { label = "Status",   value = "active" },
        { label = "ID",       value = "01ARZ...", faint = true },
    },
}
```

#### `text` — Styled text block

```lua
{
    type = "text",
    lines = {
        "Plain text line.",
        "",
        { text = "Bold heading", bold = true },
        { text = "Faint note", faint = true },
        { text = "Highlighted", accent = true },
    },
}
```

#### `table` — Headers + rows with cursor

```lua
{
    type = "table",
    headers = { "Name", "Status", "Date" },
    rows = {
        { "Item A", "active",  "2026-03-01" },
        { "Item B", "draft",   "2026-03-05" },
    },
    cursor = 0,
}
```

#### `input` — Text input field

```lua
{
    type = "input",
    id = "search",
    placeholder = "Search...",
    value = "",
    focused = true,
    char_limit = 128,
}
```

#### `select` — Option selector

```lua
{
    type = "select",
    id = "status_filter",
    options = {
        { label = "All", value = "" },
        { label = "Active", value = "active" },
    },
    selected = 0,
    focused = false,
}
```

#### `tree` — Hierarchical expandable tree

```lua
{
    type = "tree",
    nodes = {
        { label = "Root", id = "r1", expanded = true, children = {
            { label = "Child A", id = "c1" },
            { label = "Child B", id = "c2", children = {
                { label = "Grandchild", id = "g1" },
            }},
        }},
    },
    cursor = 0,
}
```

### Actions (Special Yields)

When the plugin yields a table with an `action` field, Go executes the action instead of rendering.

```lua
-- Navigate to a CMS page
coroutine.yield({ action = "navigate", page = "content" })

-- Navigate to a plugin screen
coroutine.yield({ action = "navigate", plugin = "my_plugin", screen = "edit", params = { id = "123" } })

-- Show a confirmation dialog (result delivered as "dialog" event)
coroutine.yield({ action = "confirm", title = "Delete?", message = "This cannot be undone." })

-- Show a toast notification (non-blocking)
coroutine.yield({ action = "toast", message = "Item saved", level = "success" })

-- Start async data fetch (result delivered as "data" event)
coroutine.yield({ action = "fetch", id = "load_items", query = "tasks", params = { where = { status = "active" } } })

-- Start async mutation (result delivered as "data" event)
coroutine.yield({ action = "fetch", id = "save", mutation = "update", query = "tasks", params = { set = { status = "done" }, where = { id = "123" } } })

-- Start async HTTP request (result delivered as "data" event)
coroutine.yield({ action = "request", id = "api_call", method = "GET", url = "https://example.com/api" })

-- Commit a value and close (field interfaces only)
coroutine.yield({ action = "commit", value = "#ff0000" })

-- Cancel without changing the value (field interfaces only)
coroutine.yield({ action = "cancel" })

-- Exit the screen (screens only)
coroutine.yield({ action = "quit" })
```

### Events (Go → Lua on Resume)

```lua
-- Initialization
{ type = "init", width = 120, height = 40, params = { id = "123" } }

-- Field interface initialization (includes current field value)
{ type = "init", width = 60, height = 3, value = "#ff0000", config = { ... } }

-- Key press
{ type = "key", key = "j" }
{ type = "key", key = "enter" }
{ type = "key", key = "ctrl+c" }

-- Terminal resized
{ type = "resize", width = 120, height = 40 }

-- Async data response
{ type = "data", id = "load_items", ok = true, result = { ... } }
{ type = "data", id = "load_items", ok = false, error = "timeout" }

-- Dialog response
{ type = "dialog", accepted = true }
{ type = "dialog", accepted = false }

-- Focus changed (panels in grid)
{ type = "focus", panel = 1 }
```

## How Field Interfaces Work

### In the TUI

When the content form dialog (`ContentFormDialogModel`) encounters a field of type `plugin`:

1. Query `field_plugin_config` (or `admin_field_plugin_config`) by `field_id` → get `plugin_name` and `plugin_interface`
2. Look up the plugin via `PluginManager`, verify enabled
3. Look up the interface definition, read its `mode`
4. Create a `PluginFieldBubble` that wraps a `CoroutineBridge`

#### Inline Mode

The `PluginFieldBubble` implements `FieldBubble`. It renders inline within the content form dialog's field list, just like a TextInputBubble or SelectBubble.

```
┌─ Create Content ──────────────────────────────────────────┐
│                                                            │
│  Title       [My Blog Post                           ]     │
│  Slug        [my-blog-post                           ]     │
│  Status      < draft >                                     │
│  Color    -> ██████ #ff0000        ← inline plugin field   │
│  Body        [Click to edit...                       ]     │
│                                                            │
│                            [ Cancel ]  [ Create ]          │
└────────────────────────────────────────────────────────────┘
```

- The coroutine receives key events only when the field is focused (same as any FieldBubble)
- The content form dialog intercepts tab/shift-tab/esc before forwarding keys to the bubble
- The coroutine yields a single primitive (not a grid) — it renders within the field's allocated width and a constrained height (1-5 lines, configurable in interface manifest)
- When the coroutine yields `{ action = "commit", value = "..." }`, the bubble's internal value is updated. The content form dialog reads it via `Value()` on submit.

```go
// PluginFieldBubble implements FieldBubble using a CoroutineBridge.
type PluginFieldBubble struct {
    bridge    *CoroutineBridge
    value     string           // current committed value
    primitive PluginPrimitive  // last yielded primitive for View()
    width     int
    height    int              // max lines (from interface manifest)
    focused   bool
    mode      string           // "inline" or "overlay"
    err       error            // non-nil if coroutine errored
}

func (b *PluginFieldBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
    // Convert msg → Lua event, Resume coroutine
    // If commit action → update b.value
    // If layout yield → update b.primitive
}

func (b *PluginFieldBubble) View() string {
    // Render b.primitive within b.width x b.height
}

func (b *PluginFieldBubble) Value() string { return b.value }
func (b *PluginFieldBubble) SetValue(v string) { b.value = v }
```

#### Overlay Mode

The `PluginFieldBubble` renders as a read-only display of the current value plus a "press enter to edit" hint. On enter, it opens a `PluginFieldOverlay` (implements `ModalOverlay`) with the full coroutine interface.

```
Field shows:  Color    -> #ff0000 [enter to edit]

On enter, overlay opens:

┌─ Color Picker ─────────────────────────────────────────────┐
│                                                             │
│  ┌── Palette ──────────┐ ┌── Preview ────────────────────┐ │
│  │ -> Red     #ff0000  │ │  ████████████████████████████  │ │
│  │    Green   #00ff00  │ │                                │ │
│  │    Blue    #0000ff  │ │  Current: #ff0000              │ │
│  │    Yellow  #ffff00  │ │  RGB: 255, 0, 0                │ │
│  │    Custom...        │ │                                │ │
│  └─────────────────────┘ └────────────────────────────────┘ │
│                                                             │
│  enter: select  q: cancel  tab: panel                       │
└─────────────────────────────────────────────────────────────┘
```

- The overlay captures all keys (standard ModalOverlay behavior)
- The coroutine yields full grid layouts
- `{ action = "commit", value = "..." }` closes the overlay and updates the field value
- `{ action = "cancel" }` closes the overlay without changing the value
- If the coroutine returns (function exits), that's treated as cancel

```go
// PluginFieldOverlay implements ModalOverlay for overlay-mode field interfaces.
type PluginFieldOverlay struct {
    bridge  *CoroutineBridge
    layout  *PluginLayout
    grid    GridScreen
    title   string
    value   string  // committed value (set on commit action)
    done    bool
    committed bool
}

func (o *PluginFieldOverlay) OverlayUpdate(msg tea.KeyMsg) (ModalOverlay, tea.Cmd)
func (o *PluginFieldOverlay) OverlayView(width, height int) string
```

When the overlay closes, the `PluginFieldBubble` reads the committed value and updates its internal state.

### In the Admin Panel

When `mcms-field-renderer` encounters type `plugin`:

1. Read the field's `data-plugin-name` and `data-plugin-interface` attributes (set by the handler from the `field_plugin_config` / `admin_field_plugin_config` row)
2. Check if the plugin is enabled via a lightweight admin API call
3. Render based on mode:

**Inline mode**: Load a web component served by the plugin's approved HTTP route at `/api/v1/plugins/<plugin>/interface/<interface>`. The plugin registers this route like any other HTTP route. The web component must dispatch `field-change` custom events (same contract as all admin field renderers).

**Overlay mode**: Show a button with the current value display. On click, open a modal that loads the plugin's web UI. The modal communicates the committed value back via a custom event.

This is an extension point in the existing `mcms-field-renderer.js` `connectedCallback()` switch statement — one new case for `plugin`.

### Interface Function Signature

```lua
-- interfaces/picker.lua
function interface(ctx)
    -- ctx fields (from init event):
    --   ctx.value    = current field value (string, may be empty)
    --   ctx.config   = field data config (the JSON from field definition)
    --   ctx.width    = available width (inline: field width, overlay: terminal width)
    --   ctx.height   = available height (inline: max lines, overlay: terminal height)

    local color = ctx.value or "#000000"

    while true do
        -- Inline: yield single primitive
        -- Overlay: yield grid layout
        local event = coroutine.yield({
            type = "text",
            lines = {
                { text = color, accent = true },
            },
        })

        if event.type == "key" then
            if event.key == "enter" then
                -- Commit and close
                coroutine.yield({ action = "commit", value = color })
                return -- after commit, coroutine exits
            elseif event.key == "esc" then
                coroutine.yield({ action = "cancel" })
                return
            end
            -- ... handle other keys to change color
        end
    end
end
```

### Interface vs Screen: Key Differences

| Aspect | Screen | Interface |
|--------|--------|-----------|
| Entry point | `function screen(ctx)` | `function interface(ctx)` |
| File location | `screens/<name>.lua` | `interfaces/<name>.lua` |
| Init event | `{ type = "init", params = {...} }` | `{ type = "init", value = "...", config = {...} }` |
| Yield format | Grid layout | Grid (overlay) or single primitive (inline) |
| Exit | `return` or `{ action = "quit" }` | `{ action = "commit", value = "..." }` or `{ action = "cancel" }` |
| Lifecycle | Long-lived, user navigates to it | Opened when editing a field, produces a value |
| Hosting | Full-screen `PluginTUIScreen` | `PluginFieldBubble` (inline) or `PluginFieldOverlay` (overlay) |
| Navigation | Appears in sidebar | Never in sidebar; triggered by field editing |

## Architecture

### New Files

| File | Purpose |
|------|---------|
| `internal/plugin/ui_bridge.go` | CoroutineBridge: create, resume, read yields, convert events |
| `internal/plugin/ui_primitives.go` | Lua table → Go primitive conversion (layout parser) |
| `internal/plugin/ui_renderer.go` | Go primitives → styled string rendering |
| `internal/plugin/ui_api.go` | Register `tui` Lua module with helper constructors |
| `internal/tui/screen_plugin_tui.go` | `PluginTUIScreen` — Screen impl for standalone screens |
| `internal/tui/bubble_plugin.go` | `PluginFieldBubble` — FieldBubble impl for plugin fields |
| `internal/tui/overlay_plugin.go` | `PluginFieldOverlay` — ModalOverlay impl for overlay-mode fields |
| `internal/tui/commands_plugin_tui.go` | tea.Cmd functions for plugin UI lifecycle |
| `sql/schema/XX_field_plugin_config/` | Schema + queries for `field_plugin_config` and `admin_field_plugin_config` (all 3 dialects) |
| `internal/db/field_plugin_config.go` | DbDriver wrapper methods for extension table CRUD |

### CoroutineBridge

Shared by screens, inline interfaces, and overlay interfaces. Manages one coroutine on one checked-out LState.

```go
type CoroutineBridge struct {
    parentL  *lua.LState      // checked out from VMPool
    thread   *lua.LState      // child coroutine
    entryFn  *lua.LFunction   // screen(ctx) or interface(ctx)
    plugin   *PluginInstance
    started  bool
    done     bool
}

func NewCoroutineBridge(plugin *PluginInstance, L *lua.LState, fn *lua.LFunction) *CoroutineBridge
func (cb *CoroutineBridge) Start(initEvent *lua.LTable) (YieldValue, error)
func (cb *CoroutineBridge) Resume(event *lua.LTable) (YieldValue, error)
func (cb *CoroutineBridge) Close()  // returns LState to pool
func (cb *CoroutineBridge) Done() bool
```

### YieldValue

```go
type YieldValue struct {
    IsAction  bool
    Layout    *PluginLayout    // non-nil for grid yields (screens, overlay interfaces)
    Primitive PluginPrimitive  // non-nil for single primitive yields (inline interfaces)
    Action    *PluginAction    // non-nil for action yields
}

type PluginAction struct {
    Name   string            // "navigate", "confirm", "toast", "fetch", "request", "commit", "cancel", "quit"
    Params map[string]any    // action-specific parameters
}
```

### PluginTUIScreen (Screens)

```go
type PluginTUIScreen struct {
    GridScreen
    bridge     *CoroutineBridge
    layout     *PluginLayout
    pluginName string
    screenName string
    params     map[string]string
}

func (s *PluginTUIScreen) PageIndex() PageIndex { return PLUGINTUIPAGE }
func (s *PluginTUIScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd)
func (s *PluginTUIScreen) View(ctx AppContext) string
func (s *PluginTUIScreen) KeyHints(km config.KeyMap) []KeyHint
```

### Component Flow Diagram

```
                     ┌───────────────────────────────────────┐
                     │  Bubbletea Update Loop                │
                     │  (single goroutine)                   │
                     │                                       │
  ┌──────────────────┼───────────────────────────────────┐   │
  │  Screen Mode     │                                   │   │
  │                  ▼                                   │   │
  │  PluginTUIScreen (implements Screen)                 │   │
  │    → Resume coroutine with key/data events           │   │
  │    → Render yielded grid layout                      │   │
  │    → Handle actions (navigate, fetch, quit)           │   │
  └──────────────────────────────────────────────────────┘   │
                                                             │
  ┌──────────────────────────────────────────────────────┐   │
  │  Field Mode (inline)                                 │   │
  │                                                      │   │
  │  ContentFormDialog                                   │   │
  │    └── PluginFieldBubble (implements FieldBubble)     │   │
  │          → Resume coroutine with key events           │   │
  │          → Render yielded single primitive             │   │
  │          → Handle commit/cancel actions               │   │
  └──────────────────────────────────────────────────────┘   │
                                                             │
  ┌──────────────────────────────────────────────────────┐   │
  │  Field Mode (overlay)                                │   │
  │                                                      │   │
  │  ContentFormDialog                                   │   │
  │    └── PluginFieldBubble (enter → open overlay)      │   │
  │          └── PluginFieldOverlay (ModalOverlay)        │   │
  │                → Resume coroutine with all keys       │   │
  │                → Render yielded grid layout            │   │
  │                → Handle commit/cancel actions          │   │
  └──────────────────────────────────────────────────────┘   │
                     │                                       │
                     │  All three share:                     │
                     │                                       │
                     │  CoroutineBridge                      │
                     │    → L.Resume(thread, fn, event)      │
                     │    → Read yielded layout/action       │
                     │    → Return LState to pool on close   │
                     └───────────────────────────────────────┘
```

## Full Examples

### Standalone Screen: Task Tracker

```lua
-- screens/tasks.lua
function screen(ctx)
    coroutine.yield({ action = "fetch", id = "tasks", query = "tasks", params = {} })

    local items = {}
    local cursor = 0
    local selected = nil

    while true do
        local list_items = {}
        for i, task in ipairs(items) do
            list_items[i] = { label = task.title, id = task.id, faint = task.status == "done" }
        end

        local detail_fields = {}
        if selected then
            detail_fields = {
                { label = "Title",   value = selected.title },
                { label = "Status",  value = selected.status },
                { label = "Created", value = selected.created_at },
            }
        end

        local event = coroutine.yield({
            type = "grid",
            columns = {
                { span = 4, cells = {
                    { title = "Tasks", height = 1.0, content = {
                        type = "list", items = list_items, cursor = cursor,
                    }},
                }},
                { span = 8, cells = {
                    { title = "Detail", height = 0.6, content = {
                        type = "detail", fields = detail_fields,
                    }},
                    { title = "Help", height = 0.4, content = {
                        type = "text", lines = {{ text = "n:new d:del enter:edit q:back", faint = true }},
                    }},
                }},
            },
            hints = { { key = "n", label = "new" }, { key = "q", label = "back" } },
        })

        if event.type == "data" and event.id == "tasks" then
            if event.ok then items = event.result or {} end
            if #items > 0 then selected = items[1]; cursor = 0 end

        elseif event.type == "key" then
            if event.key == "j" and cursor < #items - 1 then
                cursor = cursor + 1; selected = items[cursor + 1]
            elseif event.key == "k" and cursor > 0 then
                cursor = cursor - 1; selected = items[cursor + 1]
            elseif event.key == "n" then
                coroutine.yield({ action = "navigate", plugin = "task_tracker", screen = "edit" })
            elseif event.key == "q" then
                return
            end
        end
    end
end
```

### Inline Field Interface: Color Swatch

```lua
-- interfaces/swatch.lua
function interface(ctx)
    local color = ctx.value or "#000000"

    while true do
        local event = coroutine.yield({
            type = "text",
            lines = {
                { text = "██ " .. color, accent = true },
            },
        })

        if event.type == "key" then
            if event.key == "r" then color = "#ff0000"
            elseif event.key == "g" then color = "#00ff00"
            elseif event.key == "b" then color = "#0000ff"
            elseif event.key == "enter" then
                coroutine.yield({ action = "commit", value = color })
                return
            end
        end
    end
end
```

### Overlay Field Interface: Color Picker

```lua
-- interfaces/picker.lua
function interface(ctx)
    local colors = {
        { label = "Red",    value = "#ff0000" },
        { label = "Green",  value = "#00ff00" },
        { label = "Blue",   value = "#0000ff" },
        { label = "Yellow", value = "#ffff00" },
        { label = "White",  value = "#ffffff" },
        { label = "Black",  value = "#000000" },
    }

    local cursor = 0
    local current = ctx.value or "#000000"

    -- Find initial cursor position
    for i, c in ipairs(colors) do
        if c.value == current then cursor = i - 1; break end
    end

    while true do
        local list_items = {}
        for i, c in ipairs(colors) do
            list_items[i] = { label = c.label .. "  " .. c.value, id = c.value }
        end

        local event = coroutine.yield({
            type = "grid",
            columns = {
                { span = 4, cells = {
                    { title = "Colors", height = 1.0, content = {
                        type = "list", items = list_items, cursor = cursor,
                    }},
                }},
                { span = 8, cells = {
                    { title = "Preview", height = 1.0, content = {
                        type = "text", lines = {
                            "",
                            { text = "  ████████████████████████", accent = true },
                            "",
                            "  Selected: " .. current,
                        },
                    }},
                }},
            },
            hints = { { key = "enter", label = "select" }, { key = "esc", label = "cancel" } },
        })

        if event.type == "key" then
            if event.key == "j" and cursor < #colors - 1 then
                cursor = cursor + 1; current = colors[cursor + 1].value
            elseif event.key == "k" and cursor > 0 then
                cursor = cursor - 1; current = colors[cursor + 1].value
            elseif event.key == "enter" then
                coroutine.yield({ action = "commit", value = current })
                return
            elseif event.key == "esc" or event.key == "q" then
                coroutine.yield({ action = "cancel" })
                return
            end
        end
    end
end
```

## TUI Integration Points

### Type Registry

Register `plugin` in `type_registry.go`:

```go
func init() {
    RegisterFieldInput(FieldInputEntry{
        Key:         "plugin",
        Label:       "Plugin",
        Description: "Plugin-provided field editor",
        NewBubble:   func() FieldBubble { return NewPluginFieldBubble() },
    })
}
```

### Content Form Dialog

In `form_dialog_content.go`, the `resolveFieldInput` function already resolves the FieldBubble by type. For type `plugin`, the `PluginFieldBubble` is created and initialized:

```go
case "plugin":
    bubble := NewPluginFieldBubble()
    // Look up plugin binding from extension table
    pluginCfg, err := driver.GetFieldPluginConfig(ctx, field.FieldID) // or GetAdminFieldPluginConfig for admin fields
    if err != nil {
        bubble.SetError(fmt.Errorf("plugin config not found for field %s", field.FieldID))
    } else {
        bubble.Configure(ctx.PluginManager, pluginCfg.PluginName, pluginCfg.PluginInterface)
    }
    bubble.SetValue(existingValue)
```

The `Configure` call:
1. Looks up the plugin instance
2. Checks out an LState from the pool
3. Loads `interfaces/<name>.lua`
4. Extracts the `interface` function
5. Creates a CoroutineBridge
6. Starts the coroutine with init event containing the current value

### Navigation

Visible plugin screens appear in the TUI sidebar under a "Plugins" group. The page system gets a new `PLUGINTUIPAGE` page index. Selecting a plugin screen entry creates a `PluginTUIScreen` and pushes it onto the navigation history.

## Admin Panel Integration Points

### mcms-field-renderer Extension

In `mcms-field-renderer.js`, add a new case in `connectedCallback()`:

```javascript
case 'plugin':
    this._buildPluginField(wrapper, name, value);
    break;
```

**Inline mode**: Load the plugin's web component from its approved HTTP route. The plugin registers an HTTP handler that serves the web component JS. The component receives the current value and dispatches `field-change` events.

**Overlay mode**: Render a button showing the current value. On click, open a modal (`<dialog>`) that loads the plugin's web UI. The modal captures the committed value via a custom event and dispatches `field-change`.

The plugin's HTTP route for serving field interfaces:
```
GET /api/v1/plugins/<plugin>/interface/<interface>/component.js
```

This route must be approved via the existing plugin route approval system.

## Sandbox Changes

### Enable Coroutines

When a plugin declares `screens` or `interfaces` in its manifest, `AllowCoroutine` is set to `true` in the sandbox config for that plugin's VMs. Plugins without UI declarations keep coroutines disabled.

### TUI Module

A new frozen module `tui` provides helper constructors:

```lua
tui.grid(columns)
tui.column(span, cells)
tui.cell(title, height, content)
tui.list(items, cursor)
tui.detail(fields)
tui.text(lines)
tui.table(headers, rows, cursor)
tui.input(id, value, placeholder)
tui.select(id, options, selected)
tui.tree(nodes, cursor)
```

These are pure functions that build tables — no side effects, no I/O. Plugins can also build tables by hand.

## VM Pool Impact

A plugin with UI declarations needs additional VMs:

- Standard pool: 4 general + 1 reserved = 5 VMs
- With UI: +1 per concurrent screen/interface viewer

Each active plugin screen or field interface holds one VM for its lifetime. For screens, that's until the user navigates away. For field interfaces, that's until the content form dialog closes.

```go
// ManagerConfig addition
MaxUIVMs int // Additional VMs for screens/interfaces (default 2)
```

Pool exhaustion shows "Plugin busy" instead of blocking.

## Approval Model

**Screens**: No approval required. Same sandbox, user explicitly navigates to them.

**Field interfaces**: No approval required for the coroutine itself. However, the admin panel integration requires an approved HTTP route to serve the web component JS. The TUI integration uses the coroutine bridge directly — no HTTP involved.

**Plugin must be enabled** for its screens to appear in nav and its field interfaces to function.

## Validation

Layout yields are validated before rendering:

- Column spans sum to 12 (grid only)
- Cell heights are positive
- Primitive types are recognized
- List/table row counts capped at 10,000
- String values capped at 10KB
- Tree depth capped at 20 levels
- Inline interface yields must be a single primitive (not a grid)

Invalid yields show an error message instead of the plugin UI.

## Execution Limits

- **Per-resume timeout**: Same as `ExecTimeout` (default 5s). Coroutine killed on timeout.
- **No operation budget**: UI code is pure computation (no direct DB calls). Timeout prevents infinite loops.
- **Async I/O only**: All data access goes through `fetch`/`request` actions. Direct `db.query()` calls inside screen/interface functions are blocked.

## Implementation Phases

### Phase 1: Core Bridge

**Files:** `internal/plugin/ui_bridge.go`, `internal/plugin/ui_primitives.go`

1. `CoroutineBridge` — Create, Start, Resume, Close lifecycle
2. Layout table parser — Lua table → `PluginLayout` / `PluginPrimitive` with validation
3. Event builder — Go values → Lua event table
4. Action parser — Lua action table → `PluginAction` struct
5. Enable coroutine library when plugin has screens or interfaces
6. Unit tests with mock Lua functions

### Phase 2: Renderer

**Files:** `internal/plugin/ui_renderer.go`

1. Each primitive type's `Render(width, height, focused, accent)` method
2. `PluginLayout` → `Grid` + `[]CellContent` conversion
3. Single primitive rendering for inline mode
4. Snapshot tests for each primitive

### Phase 3: Standalone Screens

**Files:** `internal/tui/screen_plugin_tui.go`, `internal/tui/commands_plugin_tui.go`

1. `PluginTUIScreen` implementing `Screen`
2. `tea.Msg` → Lua event conversion
3. Action handling (navigate, confirm, toast, quit)
4. Screen discovery from manifest
5. Sidebar navigation entries for visible screens
6. Screen file loading from `screens/<name>.lua`

### Phase 4: Async Data

**Files:** `internal/tui/commands_plugin_tui.go`

1. `fetch` action → goroutine → `PluginDataMsg` → resume with data event
2. `mutation` actions → same pattern
3. `request` action → RequestEngine → data event
4. Error handling for timeout, pool exhaustion, DB errors

### Phase 5: Field Interfaces

**Files:** `internal/tui/bubble_plugin.go`, `internal/tui/overlay_plugin.go`

1. `PluginFieldBubble` implementing `FieldBubble`
2. Inline mode: single primitive rendering within field row
3. Overlay mode: `PluginFieldOverlay` implementing `ModalOverlay`
4. `commit`/`cancel` action handling
5. Coroutine bridge lifecycle tied to content form dialog
6. Register in type_registry as `plugin` field type
7. Wire into `resolveFieldInput` in `uiconfig_form_dialog.go`

### Phase 6: Plugin Field Type + Extension Table

**Files:** `sql/schema/XX_field_plugin_config/`, `internal/db/types/types_enums.go`, `internal/db/field_plugin_config.go`, `internal/validation/type_validators.go`

1. Create `field_plugin_config` and `admin_field_plugin_config` schemas and queries for all three dialects (SQLite, MySQL, PostgreSQL)
2. Run `just sqlc` to generate type-safe Go code
3. Add `FieldPluginConfigID` typed ID to `internal/db/types/`
4. Add `FieldTypePlugin` to `FieldType` enum
5. Add `DbDriver` interface methods: `GetFieldPluginConfig`, `CreateFieldPluginConfig`, `UpdateFieldPluginConfig`, `DeleteFieldPluginConfig` (and admin variants)
6. Implement wrapper methods in `internal/db/field_plugin_config.go` (SQLite source), run `just drivergen`
7. Validation: query extension table, verify plugin/interface exist and are enabled
8. Content field value stored as opaque string (whatever plugin produces)
9. Handle "plugin unavailable" gracefully when plugin disabled/missing
10. Add `'plugin'` to the `type` CHECK constraint in `fields` and `admin_fields` schemas

### Phase 7: TUI Module + Screen/Interface Discovery

**Files:** `internal/plugin/ui_api.go`, `internal/plugin/manager.go`

1. Register `tui` Lua module with constructors
2. Parse `screens` and `interfaces` from manifest
3. Store definitions on `PluginInstance`
4. `Manager.PluginScreens()`, `Manager.PluginInterfaces()`, `Manager.PluginInterface(plugin, name)`
5. VM pool sizing with `MaxUIVMs`

### Phase 8: Admin Panel Integration

**Files:** `internal/admin/static/js/components/mcms-field-renderer.js`

1. Add `plugin` case to field renderer
2. Inline mode: dynamic web component loading from plugin HTTP route
3. Overlay mode: modal with plugin web UI
4. `field-change` event dispatch for value propagation

### Phase 9: Plugin-to-Plugin Navigation

1. `navigate` action with `plugin` field → look up target's screen
2. Create new `PluginTUIScreen`, push to history
3. Pass `params` to target screen's init event

## Execution Order

```
Phase 1 (bridge) → Phase 2 (renderer) → Phase 3 (screens)
                                            ↓
                                       Phase 4 (async)
                                            ↓
                              ┌─────── Phase 5 (field interfaces)
                              │              ↓
                              │         Phase 6 (field type)
                              │
                              ├─────── Phase 7 (discovery + tui module)
                              │
                              ├─────── Phase 8 (admin panel)
                              │
                              └─────── Phase 9 (plugin-to-plugin nav)
```

Phases 5-9 are independent of each other after Phase 4.

## Edge Cases

- **Plugin disabled while field interface is open** — Interface continues until form closes. "Plugin unavailable" shown on next open.
- **Plugin hot-reloaded while screen/interface active** — Old coroutine keeps running on old VM. New instances use new VMs.
- **Field references a nonexistent plugin** — Shows "Plugin 'X' not found" as the field value display. Field value preserved but not editable.
- **Interface coroutine returns without committing** — Treated as cancel. Field value unchanged.
- **Inline interface yields a grid instead of primitive** — Validation error, shows error message in field row.
- **Overlay interface yields a primitive instead of grid** — Allowed. Rendered centered in the overlay area.
- **Multiple plugin fields in same content form** — Each gets its own CoroutineBridge and VM. Independent lifecycles.
- **Pool exhausted when opening field interface** — Field shows "Plugin busy" instead of editor. Value display remains visible.
- **Content form has both inline and overlay plugin fields** — Inline fields render in-place. Overlay fields show buttons. Only one overlay can be open at a time (standard ModalOverlay behavior).

## Non-Goals

- Plugin-defined custom rendering (raw strings, ANSI) — all rendering goes through primitives
- Plugin access to other users' UI state — each session is isolated
- Plugin-defined key bindings that override CMS bindings
- Animated/streaming renders — one layout per Update cycle
- Plugin fields storing structured metadata alongside the value (the value IS the metadata)
- Admin panel plugin interfaces using the coroutine bridge (admin uses HTTP-served web components)
