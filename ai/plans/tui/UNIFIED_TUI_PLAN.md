# Unified TUI Refactor + Terminal Space Plan

Combines the component isolation refactor (`major_refactor.md`) with the terminal space
utilization plan (`TERMINAL_SPACE_PLAN.md`) into one dependency-ordered sequence. Each
phase is a standalone deliverable that compiles and works before the next begins.

---

## Current State (2026-03-04)

| Aspect | Status | Location |
|--------|--------|----------|
| Model struct | ~40 infra/nav/chrome fields (Phase 13 done) | `model.go` |
| DialogContext (Phase 0.1) | COMPLETE | `model.go` |
| ModalOverlay interface | EXISTS | `modal_overlay.go` |
| Overlay compositing | EXISTS | `layer.go` |
| OverlaySetMsg / OverlayClearMsg | EXISTS | `modal_overlay.go` |
| ActiveOverlay routing | EXISTS | `update.go` (ActiveScreen block) |
| Panel struct | Extended (7 fields: +TotalLines, +ScrollOffset) | `panel.go` |
| FocusPanel type | EXISTS | `panel.go` |
| PageLayout type | COMPLETE (Phase 5) | `panel.go`, `pages.go` |
| Panel width split | Per-page ratios via `pageLayouts` | `pages.go`, `panel_view.go` |
| isCMSPanelPage | DELETED (Phase 14) | — |
| ASCII header | Adaptive: Full/Compact/Hidden | `panel_view.go` |
| CMS status bar | Adaptive: 2-line/1-line/hidden (Phase 9) | `panel_view.go` |
| Old status bar | DELETED (Phase 0.3) | — |
| Screen interface | COMPLETE (Phase 4) | `screen.go` |
| AppContext | COMPLETE (Phase 4) | `app_context.go` |
| ScreenMode | COMPLETE (Phase 2) | `panel.go`, `model.go` |
| Responsive breakpoints | COMPLETE (Phase 5) | `update_tea.go` |
| Scroll indicators | COMPLETE (Phase 11) | `panel.go`, all `screen_*.go` |
| Panel tabs | COMPLETE (Phase 12) | `panel.go`, `screen_panel.go`, `keybindings.go` |
| msg_*.go domain files | COMPLETE (Phase 0.2) | `msg_*.go` (10 files) |
| screen.go, app_context.go | COMPLETE (Phase 4) | `screen.go`, `app_context.go` |
| Screen migrations (pilot) | COMPLETE (Phase 6) | `screen_actions.go`, `screen_quickstart.go`, `screen_plugin_detail.go` |
| KeyBindings | 30 actions defined | `keybindings.go` |
| `f`/`F` keys | BOUND to screen mode toggle/next (Phase 2) | `keybindings.go` |
| `a` key | BOUND to accordion (Phase 9) | `keybindings.go` |
| `[` / `]` keys | BOUND to tab_prev/tab_next (Phase 12) | `keybindings.go` |
| KeyHinter interface | COMPLETE (Phase 9) | `screen.go` |
| KeyHint on all screens | COMPLETE (Phase 9) | `screen_*.go` (15 screens) |

### Phase Progress

| Phase | Description | Status |
|-------|-------------|--------|
| 0.1 | DialogContext | DONE (2026-03-02) |
| 0.2 | Message consolidation | DONE (2026-03-03) |
| 0.3 | Dead code removal | DONE (2026-03-03) |
| 1 | Dialog unification | DONE (2026-03-03) |
| 2 | Screen modes | DONE (2026-03-03) |
| 3 | Compact header + breadcrumbs | DONE (2026-03-03) |
| 4 | Screen interface + AppContext | DONE (2026-03-04) |
| 5 | Responsive breakpoints + page layouts | DONE (2026-03-04) |
| 6 | Pilot screen migrations | DONE (2026-03-04) |
| 7 | Accordion focus | DONE (2026-03-04) |
| 8 | PanelScreen base + CMS migrations | DONE (2026-03-04) |
| 9 | Smart statusbar | DONE (2026-03-04) |
| 10 | Fetch migration into screens | DONE (2026-03-04) |
| 11 | Scroll indicators | DONE (2026-03-04) |
| 12 | Panel tabs | DONE (2026-03-04) |
| 13 | Slim down Model | DONE (2026-03-04) |
| 14 | Cleanup | DONE (2026-03-05) |

### Key Corrections from Original Plans

1. **Overlay system already exists.** The major refactor plan proposed creating `overlay.go`
   with an `Overlay` interface. In reality, `modal_overlay.go` already defines `ModalOverlay`
   with `OverlayUpdate`/`OverlayView`, plus `OverlaySetMsg`/`OverlayClearMsg`. Phase 1 of the
   original refactor plan (dialog unification) is already ~80% complete. What remains is
   confirming all 6 dialog types implement `ModalOverlay` and cleaning up any residual
   per-type Set/ActiveSet messages.

2. **Key binding conflicts.** The terminal space plan proposed `+`/`-` for screen mode cycling
   and `=` for mode reset. These CONFLICT with existing `ActionExpand` (`+`/`=`) and
   `ActionCollapse` (`-`/`_`), which are used for content tree node expand/collapse. This plan
   uses new `Action` constants registered in `keybindings.go` with non-conflicting defaults.

3. **PanelModel is dead code.** `panel_model.go` / `panel_model_view.go` / `panel_middleware.go`
   / `panel_init.go` / `panel_update.go` / `statusbar.go` / `header.go` define a separate
   `PanelModel` with its own 25/50/25 split. Confirmed unused: none of these files are
   referenced from the production `Model`, `InitialModel`, `Update`, or `View` paths. Safe
   to delete.

4. **Two status bar systems coexist.** `renderCMSPanelStatusBar()` (panel_view.go:324) is the
   active one. `RenderStatusBar()` (style.go:167) is only reachable via `renderFallback()`,
   which only runs for unrecognized page indices — effectively dead code since
   `isCMSPanelPage` returns true for all 24 pages.

---

## Phase 0: Prerequisites

### 0.1 Move package-level context vars into Model [DONE]

Completed 2026-03-02. `DCtx DialogContext` on Model replaces 24 package-level vars.

### 0.2 Consolidate message types [DONE]

Execute the plan at `ai/plans/tui/message_types.md`:
- Boolean toggle pairs (LoadingTrue/False, ReadyTrue/False) → single msg with bool
- Cursor messages → `CursorMsg{Action, Index}`
- Direction pairs → enums
- Plugin request/result consolidations
- Reorganize into domain-grouped `msg_*.go` files

Files: `message_types.go`, `admin_message_types.go`, `constructors.go`, `admin_constructors.go`,
`update_state.go`, `update_cms.go`, plus new `msg_*.go` files.

Verify: `just check`, `just test`

### 0.3 Audit and remove dead code [DONE]

Pre-audit findings (2026-03-03):
- `PanelModel` is used only within its own 6 files (`panel_model.go`, `panel_model_view.go`,
  `panel_update.go`, `panel_init.go`, `panel_middleware.go`, `statusbar.go`). None are
  referenced from the production `Model`/`InitialModel`/`Update`/`View` paths. Safe to delete.
- `header.go` contains `renderHeader()` used only by `PanelModel.View()`. Deletes with it.
- `RenderStatusBar()` in `style.go:167-231` is reachable only via `renderFallback()` in
  `view.go:30-43`, which only runs when `isCMSPanelPage` returns false. Since
  `isCMSPanelPage` returns true for all 24 active pages, both are dead code.

Steps:
1. Delete: `panel_model.go`, `panel_model_view.go`, `panel_update.go`, `panel_init.go`,
   `panel_middleware.go`, `statusbar.go`, `header.go`
2. Delete `RenderStatusBar()` method from `style.go`
3. Delete `renderFallback()` from `view.go`. Replace the fallback branch with a log line:
   `m.Logger.Ferror("unrecognized page index", fmt.Errorf("page %d", m.Page.Index))`
   followed by returning an empty string.

Files: delete 7 files, clean `style.go`, `view.go`
Verify: `just check`, `just test`

---

## Phase 1: Complete Dialog Unification [DONE]

The `ModalOverlay` interface and routing already exist. This phase finishes the job.

### 1.1 Verify all dialog types implement ModalOverlay

As of 2026-03-03, all 6 types implement `ModalOverlay`. Verify this still holds by
confirming each has `OverlayUpdate(tea.KeyMsg) (ModalOverlay, tea.Cmd)` and
`OverlayView(width, height int) string`:
- `DialogModel`
- `FormDialogModel`
- `ContentFormDialogModel`
- `UserFormDialogModel`
- `DatabaseFormDialogModel`
- `UIConfigFormDialogModel`

If any regressed, add the two methods. `OverlayUpdate` delegates to the type's internal
key handler and returns `(self, cmd)`. `OverlayView` delegates to the type's internal
render method with the given width/height.

### 1.2 Remove residual per-type Set/ActiveSet messages [LIKELY DONE]

As of 2026-03-03, grep found zero matches for `FormDialogSetMsg`,
`FormDialogActiveSetMsg`, or `ContentFormDialogSetMsg` in any `.go` file. These names
appear only in documentation (`cli.md`). Verify with:
```
grep -r 'DialogSetMsg\|DialogActiveSetMsg' internal/tui/*.go
```
If zero matches, mark this step DONE. If matches found, replace all with
`OverlaySetMsg` / `OverlayClearMsg` and update handlers in `update_state.go`,
`update_dialog.go`, `admin_update_dialog.go`.

### 1.3 Remove residual per-type pointer+bool pairs from Model

If `model.go` still has any of these, remove them:
- `Dialog *DialogModel` / `DialogActive bool`
- `FormDialog *FormDialogModel` / `FormDialogActive bool`
- etc.

`ActiveOverlay ModalOverlay` (already at `model.go:137`) is the sole field.

Files: dialog files, `model.go`, `update_state.go`, `update_dialog.go`,
`admin_update_dialog.go`, `constructors.go`, `admin_constructors.go`
Verify: `just check`, `just test`, manual test of each dialog type

---

## Phase 2: Screen Modes (Terminal Space Phase 1) [DONE]

Foundation for all layout improvements. Introduces cycling layout modes.

### 2.1 Add ScreenMode type and layout actions

New constants in `panel.go`:

```go
type ScreenMode int

const (
    ScreenNormal ScreenMode = iota // 3 panels: proportional split
    ScreenWide                     // 2 panels: focused + gutters
    ScreenFull                     // 1 panel: focused takes 100%
)
```

New actions in `keybindings.go`:

```go
ActionScreenNext   Action = "screen_next"    // default: "F"  (shift+f)
ActionScreenPrev   Action = "screen_prev"    // default: "ctrl+shift+f" or leave unbound
ActionScreenToggle Action = "screen_toggle"  // default: "f"  (toggle Normal ↔ Full)
ActionScreenReset  Action = "screen_reset"   // default: unbound (can use ActionScreenToggle)
```

Chosen keys: `f` (toggle fullscreen) and `F` (cycle modes) are both free.
`+`/`-`/`=` are NOT used — they remain bound to tree expand/collapse.

### 2.2 Add ScreenMode fields to Model

```go
ScreenMode       ScreenMode // default ScreenNormal
ScreenModeManual bool       // true when user explicitly set mode; disables auto-breakpoint
```

### 2.3 Replace hardcoded widths with panelWidths()

Extract `panel_view.go:37-39` into a method:

```go
func (m Model) panelWidths() (left, center, right int) {
    switch m.ScreenMode {
    case ScreenNormal:
        left = m.Width / 4
        center = m.Width / 2
        right = m.Width - left - center
    case ScreenWide:
        const gutter = 4
        switch m.PanelFocus {
        case TreePanel:
            left = m.Width - 2*gutter
            center, right = gutter, gutter
        case ContentPanel:
            left, right = gutter, gutter
            center = m.Width - 2*gutter
        case RoutePanel:
            left, center = gutter, gutter
            right = m.Width - 2*gutter
        }
    case ScreenFull:
        switch m.PanelFocus {
        case TreePanel:
            left, center, right = m.Width, 0, 0
        case ContentPanel:
            left, center, right = 0, m.Width, 0
        case RoutePanel:
            left, center, right = 0, 0, m.Width
        }
    }
    return
}
```

Update `renderCMSPanelLayout` to call `m.panelWidths()` and skip rendering panels with
width 0. In `ScreenWide` mode, panels with `width == gutter` render a collapsed strip
(panel title's first character, vertically).

### 2.4 Handle screen mode keys

In each page's controls handler (or a shared helper called at the top of
`PageSpecificMsgHandlers`), check for `ActionScreenToggle` and `ActionScreenNext`:

```go
if km.Matches(key, config.ActionScreenToggle) {
    if m.ScreenMode == ScreenFull {
        m.ScreenMode = ScreenNormal
    } else {
        m.ScreenMode = ScreenFull
    }
    m.ScreenModeManual = true
    return m, nil
}
if km.Matches(key, config.ActionScreenNext) {
    m.ScreenMode = (m.ScreenMode + 1) % 3
    m.ScreenModeManual = true
    return m, nil
}
if km.Matches(key, config.ActionScreenReset) {
    m.ScreenModeManual = false // re-enable auto breakpoints
    return m, nil
}
```

### 2.5 Status bar mode indicator

Add `[Normal]` / `[Wide]` / `[Full]` badge to `renderCMSPanelStatusBar` line 1.

Files: `panel.go`, `keybindings.go`, `model.go`, `panel_view.go`, `update_controls.go`
Verify: `just check`, navigate to CMS pages, press `f`/`F`, verify mode cycling

---

## Phase 3: Compact Header + Breadcrumbs (Terminal Space Phase 2) [DONE]

### 3.1 Add HeaderMode type

```go
type HeaderMode int

const (
    HeaderFull    HeaderMode = iota // ASCII art (tall terminals)
    HeaderCompact                   // Single line
    HeaderHidden                    // No header
)
```

### 3.2 Auto-select based on terminal height

```go
func (m Model) headerMode() HeaderMode {
    switch {
    case m.Height < 24:
        return HeaderHidden
    case m.Height < 40:
        return HeaderCompact
    default:
        return HeaderFull
    }
}
```

### 3.3 Breadcrumb info bar

```go
type BreadcrumbInfo struct {
    PageName     string
    ItemCount    int
    SelectedItem string
    DBBadge      string
}
```

Populate from Model state per page. Render as:
- **HeaderFull**: second line below ASCII art
- **HeaderCompact**: `ModulaCMS v1.2.3 | PageName | N items | selected  [sqlite]`
- **HeaderHidden**: not shown

### 3.4 Update renderCMSHeader

Branch on `m.headerMode()`. The ASCII art path remains as-is with breadcrumb appended.
The compact path renders a single styled line. Hidden returns empty string.

Files: `panel_view.go` (modify `renderCMSHeader`, add breadcrumb helpers)
Verify: `just check`, resize terminal vertically to trigger mode switches

---

## Phase 4: Screen Interface and AppContext (Refactor Phase 2) [DONE]

### 4.1 Define AppContext

New file `app_context.go`:

```go
type AppContext struct {
    DB             db.DbDriver
    Config         *config.Config
    Logger         Logger
    UserID         types.UserID
    Width          int
    Height         int
    ScreenMode     ScreenMode
    PanelFocus     FocusPanel
    PluginManager  *plugin.Manager
    ConfigManager  *config.Manager
    IsRemote       bool
    SSHFingerprint string
    SSHKeyType     string
    SSHPublicKey   string
}
```

Add `func (m Model) AppCtx() AppContext` to `model.go`.

### 4.2 Define the Screen interface

New file `screen.go`:

```go
type Screen interface {
    Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd)
    View(ctx AppContext) string
    PageIndex() PageIndex
}
```

### 4.3 Add ActiveScreen to Model

```go
ActiveScreen Screen // nil = legacy path
```

### 4.4 Wire Screen into update chain

**Placement in update.go:** The `ActiveScreen` dispatch goes AFTER `UpdateTea` (which
handles `WindowSizeMsg` and must always set `m.Width`/`m.Height`) but BEFORE
`UpdateState` and all subsequent legacy handlers. When `ActiveScreen` is non-nil, it
replaces `UpdateState` through `PageSpecificMsgHandlers` for that page.

In `update.go`, insert after the `UpdateTea` call (line 21) and before `UpdateState`
(line 23):

```go
// Screen-based pages: ActiveScreen handles all messages except WindowSizeMsg
// (which UpdateTea already processed above) and provisioning.
if m.ActiveScreen != nil {
    // Overlay intercepts key input even for Screen-based pages
    if m.ActiveOverlay != nil {
        if keyMsg, ok := msg.(tea.KeyMsg); ok {
            overlay, cmd := m.ActiveOverlay.OverlayUpdate(keyMsg)
            m.ActiveOverlay = overlay
            return m, cmd
        }
    }
    // File picker intercepts key input
    if m.FilePickerActive {
        if keyMsg, ok := msg.(tea.KeyMsg); ok {
            // ... same file picker logic as existing code ...
        }
    }
    // Delegate everything else to the screen
    ctx := m.AppCtx()
    screen, cmd := m.ActiveScreen.Update(ctx, msg)
    m.ActiveScreen = screen
    return m, cmd
}
```

The legacy handlers (`UpdateState`, `UpdateFetch`, `UpdateDialog`, `UpdateCms`, etc.)
only run when `ActiveScreen` is nil (legacy pages). This is the coexistence mechanism:
Screen-based and legacy pages never share the handler chain.

**In view.go**, inside `renderCMSPanelLayout`, before the existing panel construction:

```go
if m.ActiveScreen != nil {
    ui := m.ActiveScreen.View(m.AppCtx())
    if m.FilePickerActive {
        return FilePickerOverlay(ui, m.FilePicker, m.Width, m.Height)
    }
    if m.ActiveOverlay != nil {
        return RenderOverlay(ui, m.ActiveOverlay, m.Width, m.Height)
    }
    return ui
}
```

### 4.6 Coexistence Protocol (Phases 4 through 13)

During the migration period, Screen-based pages and legacy pages coexist. The rules:

**Message routing:**
- `UpdateTea` (WindowSizeMsg) runs for ALL pages, always.
- `UpdateProvisioning` and `UpdateLog` run for ALL pages, always.
- When `ActiveScreen != nil`: overlay, file picker, then Screen.Update handle everything.
  Legacy handlers (`UpdateState`, `UpdateFetch`, `UpdateDialog`, `UpdateCms`, etc.) are
  skipped entirely.
- When `ActiveScreen == nil`: the full legacy chain runs as before.

**How Screens trigger overlays:**
Screens return `OverlaySetCmd(overlay)` as a `tea.Cmd`. This produces an `OverlaySetMsg`
in the next `Update()` cycle. In the next cycle, `UpdateTea` runs (no-op for non-resize),
then the `ActiveScreen` block runs. The overlay check at the top of the `ActiveScreen`
block catches the `OverlaySetMsg` — wait, no. `OverlaySetMsg` is handled by `UpdateState`.
Since `ActiveScreen` blocks skip `UpdateState`, we need explicit handling.

Add to the `ActiveScreen` block, before delegating to the screen:

```go
// Handle overlay set/clear for Screen-based pages
switch typedMsg := msg.(type) {
case OverlaySetMsg:
    m.ActiveOverlay = typedMsg.Overlay
    return m, nil
case OverlayClearMsg:
    m.ActiveOverlay = nil
    return m, nil
}
```

**How Screens change PanelFocus:**
`PanelFocus` lives on Model, not on Screen. Screens handle `ActionNextPanel` /
`ActionPrevPanel` in their own `Update()` and return a `tea.Cmd` that emits a message:

```go
type SetPanelFocusMsg struct{ Panel FocusPanel }
```

This message is handled in the `ActiveScreen` block (same as OverlaySetMsg above):

```go
case SetPanelFocusMsg:
    m.PanelFocus = typedMsg.Panel
    return m, nil
```

Alternatively, since `AppContext` is passed by value, the Screen can track its own
`PanelFocus` internally and ignore the Model field. This is simpler and avoids the
round-trip message. **Preferred approach:** Screens embed `PanelScreen` which owns
`PanelFocus`. The root Model's `PanelFocus` becomes unused after all screens migrate
(removed in Phase 13).

**How Screens trigger data fetches:**
Screens return `tea.Cmd` functions that perform DB queries and return result messages.
The Screen's own `Update()` handles those result messages. No legacy fetch handler
involvement. Example:

```go
func (s *RoutesScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
    switch msg := msg.(type) {
    case RoutesFetchResultsMsg:
        s.Routes = msg.Routes
        s.Loading = false
        return s, nil
    }
    // ...
}
```

**How Screens open the file picker:**
The file picker stays on Model (it is shared state). Screens emit `OpenFilePickerMsg`:

```go
type OpenFilePickerMsg struct{ Purpose FilePickerPurpose }
```

Handled in the `ActiveScreen` block:

```go
case OpenFilePickerMsg:
    m.FilePickerActive = true
    m.FilePickerPurpose = typedMsg.Purpose
    return m, nil
```

### 4.7 Add screenForPage dispatch

In navigation handler, when changing pages:

```go
m.ActiveScreen = m.screenForPage(msg.Page)
```

`screenForPage` starts returning `nil` for everything. Screens get wired as they're built.

Files: `app_context.go` (new), `screen.go` (new), `model.go`, `update.go`, `view.go`,
`update_navigation.go`
Verify: `just check`. No behavior change — `ActiveScreen` is always nil.

---

## Phase 5: Responsive Breakpoints + Page Layouts (Terminal Space Phase 3) [DONE]

### 5.1 Page layout declarations

```go
type PageLayout struct {
    Panels      int         // 1, 2, or 3
    Ratios      [3]float64  // {left, center, right} proportions
    Titles      [3]string   // panel titles
}

var pageLayouts = map[PageIndex]PageLayout{
    HOMEPAGE:          {1, [3]float64{0, 1, 0}, [3]string{"", "Home", ""}},
    QUICKSTARTPAGE:    {1, [3]float64{0, 1, 0}, [3]string{"", "Quickstart", ""}},
    ACTIONSPAGE:       {1, [3]float64{0, 1, 0}, [3]string{"", "Actions", ""}},
    PLUGINDETAILPAGE:  {1, [3]float64{0, 1, 0}, [3]string{"", "Plugin", ""}},
    CONTENT:           {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Tree", "Content", "Fields"}},
    DATATYPES:         {3, [3]float64{0.20, 0.40, 0.40}, [3]string{"Types", "Fields", "Detail"}},
    MEDIA:             {2, [3]float64{0, 0.70, 0.30}, [3]string{"", "Media", "Info"}},
    CONFIGPAGE:        {2, [3]float64{0.30, 0.70, 0}, [3]string{"Categories", "Settings", ""}},
    DATABASEPAGE:      {2, [3]float64{0.30, 0.70, 0}, [3]string{"Tables", "Data", ""}},
    ROUTES:            {3, [3]float64{0.25, 0.50, 0.25}, [3]string{"Routes", "Details", "Actions"}},
    // ... remaining pages use default 3-panel 25/50/25
}
```

Note: `cmsPanelTitles()` already returns per-page titles. That function is replaced by
`pageLayouts` lookups, with a fallback to the existing `cmsPanelTitles()` for pages not
yet in the map.

### 5.2 Responsive default mode on resize

Add a `ScreenModeManual bool` field to Model. When the user explicitly changes the
screen mode via keybinding (Phase 2.4), set `ScreenModeManual = true`. Auto-selection
only runs when `ScreenModeManual` is false.

In `UpdateTea`, after setting `m.Width`:

```go
if !m.ScreenModeManual {
    switch {
    case m.Width < 80:
        m.ScreenMode = ScreenFull
    case m.Width < 120:
        m.ScreenMode = ScreenWide
    default:
        m.ScreenMode = ScreenNormal
    }
}
```

This prevents the auto-breakpoint from overriding a user's explicit mode choice on
every terminal resize. The user can return to auto mode by pressing `ActionScreenReset`
(which sets `ScreenModeManual = false` and triggers a recalculation).

### 5.3 Integrate page layouts into panelWidths

`panelWidths()` looks up `pageLayouts[m.Page.Index]` and uses its ratios in `ScreenNormal`
mode. `ScreenWide` and `ScreenFull` override ratios as before.

Files: `panel.go` (add types), `pages.go` (add `pageLayouts` map), `panel_view.go`
(integrate), `update_tea.go` (breakpoint logic)
Verify: `just check`, test various terminal widths

---

## Phase 6: Pilot Screen Migrations (Refactor Phase 3) [DONE]

Migrate 3 simple screens to validate the Screen interface.

### 6.1 Plugin Detail screen

New file `screen_plugin_detail.go`:

```go
type PluginDetailScreen struct {
    SelectedPlugin string
    Cursor         int
    CursorMax      int // fixed at 5 (6 action items: Enable, Disable, Reload, Approve Routes, Approve Hooks, Sync)
}
```

**Model fields read:** `SelectedPlugin` (→ screen struct), `Cursor` (→ screen struct),
`Config.KeyBindings` (→ from AppContext), `History` (→ back nav via HistoryPopCmd),
`TitleFont`/`Titles` (→ from AppContext or left on Model as chrome),
`PluginManager` (→ from AppContext, used by FetchPendingRoutesForApprovalCmd/FetchPendingHooksForApprovalCmd).

**Model fields written:** `PanelFocus` (→ screen owns via PanelScreen embed), `Cursor` (→ screen struct).

**Messages emitted:** `PluginEnableRequestMsg`, `PluginDisableRequestMsg`,
`PluginReloadRequestMsg`, `PluginSyncCapabilitiesRequestMsg`, `HistoryPopCmd`,
`CursorUpCmd`, `CursorDownCmd` (screen handles cursor internally instead),
`OverlaySetCmd` (for approval dialogs).

Move code from `PluginDetailControls` (update_controls.go:1711-1791) → `Update()`.
Move code from `PLUGINDETAILPAGE` view case in `cmsPanelContent` → `View()`.
`View()` renders header + single panel + status bar using `pageLayouts`.

### 6.2 Actions screen

New file `screen_actions.go`:

```go
type ActionsScreen struct {
    Cursor    int
    CursorMax int
    IsRemote  bool
}
```

**Model fields read:** `IsRemote` (→ screen struct, set at construction), `Cursor` (→ screen),
`Config` (→ AppContext), `UserID`/`SSHFingerprint`/`SSHKeyType`/`SSHPublicKey` (→ AppContext;
add SSH fields to AppContext or pass via ActionParams at construction), `History` (→ HistoryPopCmd).

**Model fields written:** `PanelFocus` (→ screen), `Cursor` (→ screen).

**Messages emitted:** `ActionConfirmMsg`, `LoadingStartCmd`, `RunActionCmd`, `HistoryPopCmd`.

Move from `ActionsControls` (update_controls.go:1454-1511) and the ACTIONSPAGE view path.

### 6.3 Quickstart screen

New file `screen_quickstart.go`:

```go
type QuickstartScreen struct {
    Cursor    int
    CursorMax int
}
```

**Model fields read:** `Cursor` (→ screen), `Config.KeyBindings` (→ AppContext),
`History` (→ HistoryPopCmd).

**Model fields written:** `PanelFocus` (→ screen), `Cursor` (→ screen).

**Messages emitted:** `QuickstartConfirmMsg`, `HistoryPopCmd`.

Move from `QuickstartControls` (update_controls.go:1879-1922) and QUICKSTARTPAGE view.

### 6.4 Common patterns for pilot screens

All three pilot screens share the same boilerplate: quit, panel nav, back/dismiss,
cursor up/down, title font cycling. Extract a shared helper in `screen.go`:

```go
func HandleCommonKeys(key string, km config.KeyMap, cursor, cursorMax int) (newCursor int, cmd tea.Cmd, handled bool)
```

This handles ActionQuit, ActionBack, ActionUp, ActionDown. Each screen calls it first
in its `Update()`. Panel focus is handled by the embedded `PanelScreen` (Phase 8), but
for Phase 6 pilot screens that do not yet embed `PanelScreen`, handle it inline.

For each: update `screenForPage` to return the screen. Remove the page's case from
`PageSpecificMsgHandlers` and `cmsPanelContent`/`cmsPanelTitles`.

Files per screen: one new `screen_*.go`, removals from `update_controls.go`, `panel_view.go`
Verify per screen: `just check`, navigate to screen, test all controls

---

## Phase 7: Accordion Focus (Terminal Space Phase 4) [DONE]

### 7.1 Add accordion toggle

New action in `keybindings.go`:

```go
ActionAccordion Action = "accordion" // default: "a"
```

New field in `model.go`:

```go
AccordionEnabled bool
```

### 7.2 Accordion width calculation

When `AccordionEnabled` is true and `ScreenMode == ScreenNormal`, the focused panel gets
a bonus. Integrated into `panelWidths()`:

```go
if m.AccordionEnabled && m.ScreenMode == ScreenNormal {
    layout, ok := pageLayouts[m.Page.Index]
    if !ok {
        layout = defaultPageLayout // 3-panel 25/50/25
    }

    const focusFraction = 0.60
    remaining := 1.0 - focusFraction
    widths := [3]int{}

    // Sum ratios of non-focused, non-zero panels
    var otherSum float64
    for i, r := range layout.Ratios {
        if FocusPanel(i) != m.PanelFocus && r > 0 {
            otherSum += r
        }
    }

    // Assign widths
    used := 0
    for i := range layout.Ratios {
        if layout.Ratios[i] == 0 {
            widths[i] = 0 // panel not shown on this page
        } else if FocusPanel(i) == m.PanelFocus {
            widths[i] = int(focusFraction * float64(m.Width))
        } else {
            widths[i] = int(remaining * (layout.Ratios[i] / otherSum) * float64(m.Width))
        }
        used += widths[i]
    }
    // Assign rounding remainder to focused panel
    widths[m.PanelFocus] += m.Width - used

    return widths[0], widths[1], widths[2]
}
```

For pages with only 1 or 2 panels (e.g., Media has ratios `{0, 0.70, 0.30}`), the
zero-ratio panel stays at width 0 and accordion only redistributes between the active
panels.

### 7.3 Status bar indicator

Show `[Accordion]` in status bar line 1 when enabled.

Files: `keybindings.go`, `model.go`, `panel_view.go`
Verify: `just check`, toggle with `a`, verify panel resizing

---

## Phase 8: PanelScreen Base + CMS Migrations (Refactor Phase 4)

### 8.1 PanelScreen base struct

New file `screen_panel.go`:

```go
type PanelScreen struct {
    Layout     PageLayout
    PanelFocus FocusPanel
    Cursor     int
    CursorMax  int
}

func (p *PanelScreen) RenderPanels(ctx AppContext, left, center, right string) string
func (p *PanelScreen) HandlePanelNav(key string, km config.KeyMap) (handled bool)
```

`RenderPanels` uses `ctx.ScreenMode`, `ctx.Width`, the screen's `Layout`, and accordion
state to calculate widths and render the panel triplet. This replaces the per-screen
duplication of `renderCMSPanelLayout` logic.

### 8.2 CMS screen migrations

Each new screen embeds `PanelScreen` and implements `Screen`. Screens 1-11 may be
developed in parallel on separate branches after 8.1 is merged. Content (last row)
MUST be done after all others are merged.

| Screen | File | Moves From |
|--------|------|------------|
| Plugins | `screen_plugins.go` | `PluginsControls`, plugins panel rendering |
| Routes (merged) | `screen_routes.go` | `RoutesControls`, `AdminRoutesControls` |
| Media | `screen_media.go` | `MediaControls`, media panel rendering |
| Users | `screen_users.go` | `UsersAdminControls`, users panel rendering |
| Field Types (merged) | `screen_field_types.go` | `FieldTypesControls`, `AdminFieldTypesControls` |
| Datatypes (merged) | `screen_datatypes.go` | `DatatypesControls`, `AdminDatatypesControls` |
| Config | `screen_config.go` | `ConfigPanelControls`, config panel rendering |
| Database | `screen_database.go` | `DatabasePanelControls`, `DatabaseDataPanelControls` |
| Deploy | `screen_deploy.go` | `DeployControls`, deploy panel rendering |
| Pipelines | `screen_pipelines.go` | `PipelinesControls`, `PipelineDetailControls2` |
| Webhooks | `screen_webhooks.go` | `WebhooksControls` |
| **Content (LAST)** | `screen_content.go` | `ContentBrowserControls`, `AdminContentBrowserControls`, content tree rendering |

**Merged admin/regular screens (Routes, Field Types, Datatypes):** These screens have an
`AdminMode bool` field. `screenForPage` maps both page indices to the same screen type:

```go
func (m Model) screenForPage(page PageIndex) Screen {
    switch page {
    case ROUTES:
        return NewRoutesScreen(false) // AdminMode = false
    case ADMINROUTES:
        return NewRoutesScreen(true)  // AdminMode = true
    // ...
    }
}
```

The screen's `PageIndex()` method returns `ROUTES` or `ADMINROUTES` based on `AdminMode`.
Fetch commands and view rendering branch on `AdminMode` internally.

**Parallel merge conflict handling:** Each screen migration removes its own `case` branch
from `update_controls.go:PageSpecificMsgHandlers` and its own rendering path from
`cmsPanelContent` in `panel_view.go`. When developing in parallel, merge conflicts in
these files are expected and are purely subtractive (each branch deletes different `case`
blocks). Resolve by accepting all deletions.

File picker stays on root Model. Media screen emits `OpenFilePickerMsg` (see Phase 4.6
Coexistence Protocol) for root to handle.

Files: `screen_panel.go` (new), one `screen_*.go` per migration, removals from
`update_controls.go`, `panel_view.go`, `admin_panel_view.go`
Verify per screen: `just check`, full manual walkthrough

---

## Phase 9: Smart Statusbar (Terminal Space Phase 6)

### 9.1 Adaptive height

```
Height >= 30:  2-line statusbar (current behavior)
Height < 30:   1-line condensed statusbar
Height < 20:   No statusbar
```

### 9.2 Condensed format

Show exactly 6 key hints, selected by this algorithm:

| Slot | Source | Condition | Example |
|------|--------|-----------|---------|
| 1 | Primary action | Always | `enter:select` or `n:new` |
| 2 | CRUD 1 | Item selected (Cursor > -1) | `e:edit` |
| 3 | CRUD 2 | Item selected | `d:delete` |
| 4 | Navigation | Always | `tab:panel` |
| 5 | (spacer) | — | Right-aligned from here |
| 6 | Global quit | Always | `q:quit` |

If no item is selected (slots 2-3 empty), fill with mode hints (`f:focus`, `a:accord`).
If fewer than 6 hints are available, leave gaps — do not pad with irrelevant keys.

Priority groups for the `KeyHinter` interface (used by migrated screens):
1. **Primary action** (1 hint) — the most important thing the user can do right now
2. **CRUD** (0-2 hints) — only when an item is selected
3. **Navigation** (1 hint) — panel switching or back
4. **Mode** (0-2 hints) — screen mode, accordion
5. **Global** (1 hint, right-aligned) — quit/help

### 9.3 Context from Screen interface

Migrated screens can declare their own key hints via an optional method:

```go
type KeyHinter interface {
    KeyHints(km config.KeyMap) []KeyHint
}
```

### 9.4 Integration

`renderCMSPanelStatusBar` checks terminal height and renders the appropriate variant.
For screens implementing `KeyHinter`, use their hints. Fallback to `getContextControls()`
for legacy pages.

Files: `panel_view.go` (modify `renderCMSPanelStatusBar`), `screen.go` (optional interface)
Verify: `just check`, resize terminal vertically

---

## Phase 10: Fetch Migration Into Screens (Refactor Phase 5)

For each migrated screen:
1. Screen's `Update()` handles its fetch request messages (e.g., `RoutesFetchMsg`) by
   creating the async DB command using `ctx.DB`
2. Screen's `Update()` handles its fetch result messages by storing data locally
3. Remove corresponding cases from `UpdateFetch` / `UpdateAdminFetch`

After all screens migrate, `UpdateFetch` and `UpdateAdminFetch` should be empty. Delete
them and remove from the update chain in `update.go`.

Files: each `screen_*.go`, `update_fetch.go`, `admin_update_fetch.go`, `update.go`
Verify: `just check` after each screen's fetch migration

---

## Phase 11: Scroll Indicators (Terminal Space Phase 7)

### 11.1 Extend Panel struct

```go
type Panel struct {
    Title        string
    Width        int
    Height       int
    Content      string
    Focused      bool
    TotalLines   int // 0 = no scrollbar
    ScrollOffset int
}
```

### 11.2 Scrollbar rendering

In `Panel.Render()`, when `TotalLines > innerHeight`:
- Calculate thumb position and size
- Render `▓` (thumb) / `░` (track) in rightmost column
- Append ` N/M` to title

### 11.3 Pass scroll state

`PanelScreen.RenderPanels` passes `TotalLines` and `ScrollOffset` from each sub-view's
cursor/list state.

Files: `panel.go` (extend struct + render), `screen_panel.go` (pass state)
Verify: `just check`, scroll in lists with more items than panel height

---

## Phase 12: Panel Tabs (Terminal Space Phase 5)

### 12.1 Tab types

```go
type PanelTab struct {
    Label  string
    Render func(ctx AppContext, w, h int) string
}
```

The `Render` function is a closure that captures the screen's state. Each screen creates
its tabs in its constructor or `Update()`, binding `Render` to methods on the screen
struct. This gives the tab render function access to cursor position, selected items,
loaded data, etc. Example:

```go
func NewContentScreen() *ContentScreen {
    s := &ContentScreen{}
    s.TabSets[1] = []PanelTab{
        {"List", s.renderContentList},     // method on *ContentScreen
        {"Preview", s.renderContentPreview},
        {"JSON", s.renderContentJSON},
    }
    return s
}
```

### 12.2 Tab state in PanelScreen

```go
type PanelScreen struct {
    Layout     PageLayout
    PanelFocus FocusPanel
    Cursor     int
    CursorMax  int
    TabSets    [3][]PanelTab  // tabs per panel (left, center, right)
    ActiveTabs [3]int         // active tab index per panel
}
```

### 12.3 Tab navigation

New actions in `keybindings.go`:

```go
ActionTabPrev Action = "tab_prev" // default: "["
ActionTabNext Action = "tab_next" // default: "]"
```

`[` and `]` cycle tabs within the focused panel.

### 12.4 Tab bar rendering

`Panel.Render()` shows a tab bar below the title when `len(tabs) > 1`, consuming 1 line
of inner height.

### 12.5 Per-screen tab definitions

Each screen that wants tabs declares them in its constructor:

```go
// Content screen example
s.TabSets[ContentPanel] = []PanelTab{
    {"List", renderContentList},
    {"Preview", renderContentPreview},
    {"JSON", renderContentJSON},
}
```

Files: `panel.go`, `screen_panel.go`, `keybindings.go`, individual screen files
Verify: per-screen, `just check`, test `[`/`]` tab cycling

---

## Phase 13: Slim Down Model (Refactor Phase 6)

Remove fields from Model that now live in screen components:

**CMS panel fields to remove:**
Routes, AllDatatypes, SelectedDatatype, SelectedDatatypeFields, FieldCursor,
SelectedContentFields, Root, RootDatatypes, RootContentSummary, MediaList, UsersList,
RolesList, PageRouteId, PendingCursorContentID

**Admin fields to remove:**
AdminRoutes, AdminAllDatatypes, AdminSelectedDatatypeFields, AdminRootContentSummary,
AdminSelectedContentFields, AdminFieldCursor

**Plugin fields to remove:**
PluginsList, SelectedPlugin

**Config fields to remove:**
ConfigCategory, ConfigCategoryFields, ConfigFieldCursor

**Pipeline fields to remove:**
PipelinesList, PipelineEntries, SelectedPipelineKey

**Deploy fields to remove:**
DeployEnvironments, DeployLastResult, DeployLastHealth, DeployStatusMessage,
DeployOperationActive

**Webhook/version/field type fields to remove:**
WebhooksList, FieldTypesList, AdminFieldTypesList, Versions, ShowVersionList,
VersionContentID, VersionRouteID, VersionCursor

**Legacy nav to remove:**
Cursor, CursorMax (screens own these now)

**Target Model size:** ~25-30 fields (infra, terminal, navigation, active components,
chrome, provisioning).

Delete emptied handler files (`update_fetch.go`, `admin_update_fetch.go`, `update_cms.go`,
`admin_update_cms.go`, `update_controls.go`) as they become empty.

Files: `model.go`, delete emptied `update_*.go` files, `update.go`
Verify: `just check`, `just test`, grep for removed field names

---

## Phase 14: Cleanup (Refactor Phase 7)

1. Delete `message_types.go` and `admin_message_types.go` if all types moved to `msg_*.go`
   or screen files
2. Delete `panel_view.go` and `admin_panel_view.go` if all rendering moved to
   `screen_panel.go` and screen files
3. Delete emptied constructors if all moved to screen files
4. Remove `isCMSPanelPage` (all pages use Screen interface now)
5. Remove `cmsPanelContent` and `cmsPanelTitles` (replaced by pageLayouts + Screen.View)
6. Update `ai/workflows/CREATING_TUI_SCREENS.md` to document the Screen interface pattern
7. Remove dead code flagged by compiler

Files: various deletions, workflow doc update
Verify: `just check`, `just test`, full manual walkthrough of every page

---

## Dependency Graph

### Explicit Dependencies

| Phase | Depends On | Reason |
|-------|-----------|--------|
| 0.2 | 0.1 | Needs DialogContext done |
| 0.3 | 0.1 | Needs DialogContext done |
| 1 | 0.2, 0.3 | Clean message types and dead code first |
| 2 | 0.2, 0.3 | Needs clean codebase |
| 3 | 0.2, 0.3 | Needs clean codebase |
| 4 | 2 | Uses `ScreenMode` type in `AppContext` |
| 5 | 2, 4 | Uses `ScreenMode` + `Screen` interface for page layouts |
| 6 | 4 | Needs `Screen` interface |
| 7 | 2, 5 | Uses `panelWidths()` and `pageLayouts` |
| 8 | 6 | Needs pilot screens proven first |
| 9 | 8 | Needs `KeyHinter` interface from migrated screens |
| 10 | 8 | Moves fetch logic into screens |
| 11 | 8 | Needs `PanelScreen.RenderPanels` |
| 12 | 8 | Builds tabs into `PanelScreen` |
| 13 | 10, 12 | All screens migrated, all fetches moved |
| 14 | 13 | Final cleanup |

Note: Phases 1 and 3 do NOT depend on Phase 2. Phase 4 does NOT depend on Phases 1 or 3.

### Visual

```
0.1 [DONE]
  │
  ├── 0.2 (msgs) ──┐
  └── 0.3 (dead) ──┤   ← parallel
                    │
        ┌───────────┼───────────┐
        │           │           │
     1 (dialog)  2 (modes)  3 (header)   ← all three parallel
        │           │           │
        │           ▼           │
        │     4 (Screen iface)  │
        │        │    │         │
        │        ▼    ▼         │
        │    5 (responsive)     │
        │        │    │         │
        │        ▼    ▼         │
        │     6 (pilots)        │
        │        │              │
        │        ▼              │
        │   7 (accordion)       │
        │        │              │
        │        ▼              │
        │   8 (CMS screens) ────┘
        │        │
        │   ┌────┼────┐
        │   ▼    ▼    ▼
        │  9   10    11,12   ← 9,11 parallel; 10,12 parallel
        │        │    │
        │        ▼    ▼
        │      13 (slim)
        │        │
        └────► 14 (cleanup)
```

### Parallelism Opportunities

- **0.2 and 0.3** can run in parallel
- **1, 2, 3** can run in parallel (no dependencies between them after 0.2/0.3)
- **5 and 6** can run in parallel after Phase 4
- **Within Phase 8**, all screens except Content can be developed in parallel after 8.1
- **9 and 11** are independent of each other (both need Phase 8 done)
- **10 and 12** are independent of each other (both need Phase 8 done)
- **Within Phase 10**, fetch migrations can run per-screen in parallel

---

## Verification Strategy

### Compile and Test

After every step:
1. `just check` — compile verification (primary safety net)
2. `just test` — after each phase completion

**Note on TUI test coverage:** As of 2026-03-03, there are zero test files in
`internal/tui/`. `just test` verifies non-TUI packages only. For TUI changes,
`just check` (compilation) is the automated safety net. Manual testing is the
functional safety net. Do NOT rely on `just test` passing as proof of TUI correctness.

### Manual Testing

After each screen migration:
3. Manual walkthrough via SSH TUI: navigate to screen, test all keyboard controls,
   test dialog open/close/accept/cancel, test back navigation, test data fetching

After all phases:
4. Full walkthrough of every page
5. `just test` (full suite)
6. Grep for any removed type/field names to catch stale references

### Commit Strategy

Each phase should produce one commit (or one commit per sub-step for Phase 8 screen
migrations). If `just check` fails at any point, revert the commit and debug before
proceeding. Never leave the codebase in a state where `just check` fails.

For Phase 8 parallel development: each screen migration is one commit on its own branch.
Merge branches one at a time, resolving subtractive conflicts in `update_controls.go`
and `panel_view.go` by accepting all deletions.

---

## New Keybinding Summary

| Key | Action | Phase | Notes |
|-----|--------|-------|-------|
| `f` | `screen_toggle` (Normal ↔ Full) | 2 | Was free |
| `F` | `screen_next` (cycle Normal → Wide → Full) | 2 | Was free |
| `a` | `accordion` (toggle accordion mode) | 7 | Was free |
| `[` | `tab_prev` (previous tab in panel) | 12 | Was free |
| `]` | `tab_next` (next tab in panel) | 12 | Was free |

Existing bindings preserved without conflict. `+`/`-`/`=` remain tree expand/collapse.
