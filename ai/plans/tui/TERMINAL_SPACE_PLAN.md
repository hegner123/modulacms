# Terminal Space Utilization Plan

Inspired by lazygit and lazydocker layout systems. Adapted for Bubbletea + Lipgloss (no gocui).

---

## Current State

```
┌──────────────────────────────────────────────────────┐
│  Header (ASCII art title)                            │  ~5 lines fixed
├───────────┬──────────────┬───────────────────────────┤
│  Left 25% │  Center 50%  │  Right 25%                │  bodyH = Height - header - status
│  (Tree)   │  (Content)   │  (Route/Fields)           │  ALWAYS 3 panels, ALWAYS 25/50/25
├───────────┴──────────────┴───────────────────────────┤
│  StatusBar (badges + keybinds)                       │  2 lines fixed
└──────────────────────────────────────────────────────┘
```

**Problems:**
- Fixed 25/50/25 split wastes space — left/right panels are often too wide or too narrow
- No way to focus a single panel full-screen (reading long content, editing forms)
- Three panels always shown even when a page only uses one or two
- Header consumes 5+ lines of vertical space for ASCII art
- No responsive behavior — 80-column terminal gets the same layout as 240-column
- Panel content truncates without scroll indicators
- No accordion/collapse — all panels compete for the same rigid space

---

## Design Principles (Borrowed from lazygit/lazydocker)

| Principle | lazygit | lazydocker | ModulaCMS adaptation |
|-----------|---------|------------|---------------------|
| **Screen modes** | Normal / Half / Full cycling | SCREEN_NORMAL / HALF / FULL | 3 layout modes cycled via `+`/`-` |
| **Accordion** | Focused side panel expands, others collapse | N/A | Focused panel gets 60%+, others shrink |
| **Panel maximize** | `Enter` on side panel opens full main | Main panel grows per mode | `f` toggles focused panel to 100% width |
| **Context-aware layout** | Different contexts show different panels | Detail view replaces main | Pages declare their panel requirements |
| **Responsive sizing** | Width-aware truncation | Tainted render optimization | Breakpoints at 80/120/160 columns |
| **View stacking** | Multiple views per window with tabs | Tabs in main panel | Tab bar within panels |

---

## Phase 1 — Screen Modes

Add three layout modes that cycle with `+` / `-` keys (lazydocker pattern).

### 1.1 Mode Definitions

```go
type ScreenMode int

const (
    ScreenNormal ScreenMode = iota // 3 panels: 25/50/25
    ScreenWide                     // 2 panels: side collapses to icon-width (4 cols), main gets rest
    ScreenFull                     // 1 panel: focused panel takes 100%
)
```

### 1.2 Layout Calculation

Replace the hardcoded split in `panel_view.go:25-39`:

```go
func (m Model) panelWidths() (left, center, right int) {
    switch m.ScreenMode {
    case ScreenNormal:
        left = m.Width / 4
        center = m.Width / 2
        right = m.Width - left - center
    case ScreenWide:
        // Collapse non-focused panels to 4-char gutter (icon + border)
        const gutter = 4
        switch m.PanelFocus {
        case TreePanel:
            left = m.Width - 2*gutter
            center = gutter
            right = gutter
        case ContentPanel:
            left = gutter
            center = m.Width - 2*gutter
            right = gutter
        case RoutePanel:
            left = gutter
            center = gutter
            right = m.Width - 2*gutter
        }
    case ScreenFull:
        // Only focused panel rendered
        switch m.PanelFocus {
        case TreePanel:
            left = m.Width
            center = 0
            right = 0
        case ContentPanel:
            left = 0
            center = m.Width
            right = 0
        case RoutePanel:
            left = 0
            center = 0
            right = m.Width
        }
    }
    return
}
```

### 1.3 Keybindings

| Key | Action |
|-----|--------|
| `+` | Next screen mode (Normal → Wide → Full → Normal) |
| `-` | Previous screen mode |
| `f` | Toggle between ScreenNormal and ScreenFull for current panel |
| `=` | Reset to ScreenNormal |

### 1.4 Collapsed Gutter Rendering

When a panel is collapsed to 4 chars in `ScreenWide` mode, render a vertical icon strip instead of content:

```
┌──┐
│📁│   ← Tree panel collapsed
│  │
│  │
└──┘
```

Use the first character of the panel title or a Unicode icon. Clicking/focusing the gutter expands it.

### 1.5 Files Changed

| File | Change |
|------|--------|
| `model.go` | Add `ScreenMode ScreenMode` field |
| `panel.go` | Add `ScreenMode` type + constants |
| `panel_view.go` | Replace hardcoded widths with `panelWidths()` call, conditional panel render |
| `panel_model_view.go` | Same width refactor for `PanelModel` |
| `update_tea.go` | Handle `+`/`-`/`f`/`=` key messages |
| `statusbar.go` | Show current mode indicator: `[Normal]` / `[Wide]` / `[Full]` |

---

## Phase 2 — Compact Header

The ASCII art header consumes 5+ lines. Lazygit and lazydocker use a single-line header.

### 2.1 Header Modes

```go
type HeaderMode int

const (
    HeaderFull    HeaderMode = iota // ASCII art (current, for tall terminals)
    HeaderCompact                   // Single line: "ModulaCMS v1.2.3 │ Content │ 12 items"
    HeaderHidden                    // No header (maximum content space)
)
```

### 2.2 Automatic Selection

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

### 2.3 Breadcrumb Info Bar (All Header Modes)

The breadcrumb info items — page name, item count, selected item preview, db badge — appear in **every** header mode, not just compact:

| Header Mode | Breadcrumb Rendering |
|-------------|---------------------|
| **HeaderFull** | Second line below ASCII art: ` Datatypes │ 12 items │ Field: title (text)  [sqlite]` |
| **HeaderCompact** | Inline after app name: `ModulaCMS v1.2.3 │ Datatypes │ 12 items │ Field: title (text)  [sqlite]` |
| **HeaderHidden** | Not displayed (header is fully hidden to maximize space) |

Info items sourced from Model state:

```go
type BreadcrumbInfo struct {
    PageName     string // from Page.Title or PageIndex display name
    ItemCount    int    // from current list/table length
    SelectedItem string // from cursor row preview (e.g. "Field: title (text)")
    DBBadge      string // from config db_driver: "sqlite", "mysql", "psql"
}
```

The full-height ASCII art header currently shows only the logo. Adding the breadcrumb bar below it gives users the same at-a-glance context regardless of terminal height, and eliminates the need to scan the panel titles or statusbar for orientation.

### 2.4 Compact Header Layout

```
 ModulaCMS v1.2.3 │ Datatypes │ 12 items │ Field: title (text)       [sqlite]
```

Single line: app name + version merged with breadcrumb info. Reclaims 4+ lines of vertical space.

### 2.5 Files Changed

| File | Change |
|------|--------|
| `model.go` | Add `HeaderMode` type, `BreadcrumbInfo` struct, auto-select in WindowSizeMsg handler |
| `header.go` | Add `renderBreadcrumbBar()` shared function, `renderCompactHeader()`, update `renderCMSHeader()` to append breadcrumb bar below ASCII art |
| `panel_view.go` | Branch on `headerMode()` in `renderCMSPanelLayout`, populate `BreadcrumbInfo` from current page state |

---

## Phase 3 — Responsive Breakpoints

Different terminal widths get different default layouts (lazygit pattern).

### 3.1 Breakpoint Tiers

| Terminal Width | Default Mode | Panel Split | Header |
|---------------|-------------|-------------|--------|
| < 80 cols | `ScreenFull` | Single panel only | Hidden |
| 80–119 cols | `ScreenWide` | Focused panel + gutters | Compact |
| 120–159 cols | `ScreenNormal` | 20/55/25 | Compact |
| 160+ cols | `ScreenNormal` | 25/50/25 | Full |

### 3.2 Adaptive Ratios

Instead of fixed 25/50/25, use proportional ratios that adapt:

```go
type PanelRatios struct {
    Left, Center, Right float64
}

var defaultRatios = map[ScreenMode]map[string]PanelRatios{
    ScreenNormal: {
        "narrow":  {0.20, 0.55, 0.25},  // 80-119
        "medium":  {0.22, 0.53, 0.25},  // 120-159
        "wide":    {0.25, 0.50, 0.25},  // 160+
    },
}
```

### 3.3 Page-Specific Overrides

Not all pages need three panels equally. Let pages declare preferred ratios:

| Page | Left | Center | Right | Notes |
|------|------|--------|-------|-------|
| Home/Quickstart | 0 | 100 | 0 | Single panel dashboard |
| Content Browser | 25 | 50 | 25 | Standard 3-panel |
| Datatype Editor | 20 | 40 | 40 | Right panel needs more space for field details |
| Media Gallery | 0 | 70 | 30 | No tree, wide preview |
| Config/Database | 30 | 70 | 0 | Two-panel: list + detail |
| Plugin Detail | 0 | 100 | 0 | Full-width detail view |
| Actions | 0 | 100 | 0 | Full-width log stream |

```go
type PageLayout struct {
    Panels    int            // 1, 2, or 3
    Ratios    PanelRatios    // proportional widths
    LeftTitle string
    CenterTitle string
    RightTitle string
}

var pageLayouts = map[PageIndex]PageLayout{
    HomeIndex:        {1, PanelRatios{0, 1, 0}, "", "Home", ""},
    ContentIndex:     {3, PanelRatios{0.25, 0.50, 0.25}, "Tree", "Content", "Fields"},
    DatatypesIndex:   {3, PanelRatios{0.20, 0.40, 0.40}, "Types", "Fields", "Detail"},
    MediaIndex:       {2, PanelRatios{0, 0.70, 0.30}, "", "Media", "Info"},
    ConfigIndex:      {2, PanelRatios{0.30, 0.70, 0}, "Categories", "Settings", ""},
    PluginDetailIndex:{1, PanelRatios{0, 1, 0}, "", "Plugin", ""},
    ActionsIndex:     {1, PanelRatios{0, 1, 0}, "", "Actions", ""},
}
```

### 3.4 Files Changed

| File | Change |
|------|--------|
| `panel.go` | Add `PageLayout`, `PanelRatios` types |
| `pages.go` | Add `pageLayouts` map |
| `panel_view.go` | Look up `pageLayouts[m.Page.Index]` before calculating widths |
| `update_tea.go` | Set default `ScreenMode` based on width breakpoint on resize |

---

## Phase 4 — Accordion Focus (lazygit-inspired)

When a panel receives focus, it grows; siblings shrink.

### 4.1 Accordion Ratios

In `ScreenNormal` mode, focused panel gets a bonus:

```go
func (m Model) accordionWidths() (left, center, right int) {
    if !m.AccordionEnabled {
        return m.panelWidths() // fall through to standard
    }
    // Focused panel: 50%, others: 25% each
    // (vs default 25/50/25 where center always gets 50%)
    const focusRatio = 0.50
    const shrinkRatio = 0.25
    
    switch m.PanelFocus {
    case TreePanel:
        left = int(float64(m.Width) * focusRatio)
        center = int(float64(m.Width) * shrinkRatio)
        right = m.Width - left - center
    case ContentPanel:
        // Same as default — center already has 50%
        return m.panelWidths()
    case RoutePanel:
        left = int(float64(m.Width) * shrinkRatio)
        center = int(float64(m.Width) * shrinkRatio)
        right = m.Width - left - center
    }
    return
}
```

### 4.2 Smooth Transition (Optional Enhancement)

Use a `tea.Tick` command to animate the width transition over 3-4 frames (~50ms total). Store `targetWidths` and `currentWidths`, interpolating each tick.

### 4.3 Toggle

| Key | Action |
|-----|--------|
| `a` | Toggle accordion mode on/off |

### 4.4 Files Changed

| File | Change |
|------|--------|
| `model.go` | Add `AccordionEnabled bool` |
| `panel_view.go` | Call `accordionWidths()` instead of `panelWidths()` when enabled |
| `update_tea.go` | Handle `a` key toggle |
| `statusbar.go` | Show `[Accordion]` indicator |

---

## Phase 5 — Panel Tabs (lazygit Window pattern)

Multiple views share the same panel space via tabs.

### 5.1 Concept

```
┌── Content ──────────────────────────────────────────┐
│ [List] [Preview] [JSON]                              │  ← Tab bar
│                                                      │
│  Title         Status      Updated                   │
│  Homepage      published   2024-01-15                │
│  About Us      draft       2024-01-14                │
│                                                      │
└──────────────────────────────────────────────────────┘
```

Tabs within a panel allow switching between different views of the same data without consuming additional panels.

### 5.2 Tab Definitions Per Page

```go
type PanelTab struct {
    Label   string
    Render  func(m Model, w, h int) string
}

// Content page: center panel tabs
contentCenterTabs := []PanelTab{
    {"List", renderContentList},
    {"Preview", renderContentPreview},
    {"JSON", renderContentJSON},
}

// Datatype page: right panel tabs
datatypeRightTabs := []PanelTab{
    {"Fields", renderDatatypeFields},
    {"Validators", renderDatatypeValidators},
    {"Schema", renderDatatypeSchema},
}
```

### 5.3 Tab Navigation

| Key | Action |
|-----|--------|
| `[` | Previous tab in focused panel |
| `]` | Next tab in focused panel |
| `1-9` | Jump to tab N in focused panel |

### 5.4 Tab Bar Rendering

```go
func renderTabBar(tabs []PanelTab, active int, width int) string {
    var parts []string
    for i, tab := range tabs {
        if i == active {
            parts = append(parts, accentStyle.Render("["+tab.Label+"]"))
        } else {
            parts = append(parts, dimStyle.Render(" "+tab.Label+" "))
        }
    }
    bar := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
    return lipgloss.NewStyle().Width(width).Render(bar)
}
```

### 5.5 Files Changed

| File | Change |
|------|--------|
| `panel.go` | Add `Tabs []PanelTab`, `ActiveTab int` to `Panel` struct |
| `panel.go` | Render tab bar inside `Panel.Render()` above content |
| `model.go` | Add per-panel `ActiveTab` tracking |
| `update_tea.go` | Handle `[`/`]` tab switching |
| `panel_view.go` | Supply tabs when constructing panels per page |

---

## Phase 6 — Smart Statusbar

Replace the static 2-line statusbar with a context-aware single line (lazydocker pattern).

### 6.1 Adaptive Statusbar

```
Height ≥ 30:  2-line statusbar (current behavior)
Height < 30:  1-line statusbar (condensed keybinds)
Height < 20:  No statusbar (keys shown in header)
```

### 6.2 Condensed Format

Current (2 lines):
```
 content:published  [Content] Tree  Content  Route    en-US
 n:new  e:edit  d:delete  /:search  ?:help  q:quit
```

Condensed (1 line):
```
 n:new e:edit d:del /:search tab:panel +:mode  [Content] [Normal] [en-US]
```

### 6.3 Contextual Key Grouping

Show only the 5-6 most relevant keys for the current focus + panel, not all possible keys. Group by importance:

1. **Primary action** (n/Enter) — always shown
2. **CRUD** (e/d) — shown when item selected
3. **Navigation** (tab/arrows) — shown when multiple panels
4. **Mode** (+/-/f/a) — shown as compact indicators
5. **Global** (q/?) — always shown, right-aligned

### 6.4 Files Changed

| File | Change |
|------|--------|
| `statusbar.go` | Add `renderCompactStatusBar()`, adaptive height selection |
| `panel_view.go` | Use adaptive statusbar in `renderCMSPanelLayout` |
| `model.go` | Statusbar height auto-determined from terminal height |

---

## Phase 7 — Panel Scroll Indicators

Show scroll position when panel content overflows.

### 7.1 Scroll Indicator

```
┌── Content ──────────────────────────── ▲ 1/47 ──┐
│  Homepage        published   2024-01             │
│  About Us        draft       2024-01             │
│  Contact         published   2024-01             ▓
│  Blog            published   2024-01             ░
│  Products        draft       2024-01             ░
│  Services        published   2024-01             │
└──────────────────────────────────────── ▼ ───────┘
```

### 7.2 Implementation

```go
func (p Panel) Render() string {
    // ... existing border render ...
    
    if p.TotalLines > p.innerHeight() {
        // Calculate scrollbar position
        thumbPos := int(float64(p.ScrollOffset) / float64(p.TotalLines) * float64(p.innerHeight()))
        thumbSize := max(1, int(float64(p.innerHeight()) / float64(p.TotalLines) * float64(p.innerHeight())))
        
        // Render right-edge scrollbar track
        for y := 0; y < p.innerHeight(); y++ {
            if y >= thumbPos && y < thumbPos+thumbSize {
                lines[y] += "▓"
            } else {
                lines[y] += "░"
            }
        }
        
        // Title suffix: position indicator
        p.Title += fmt.Sprintf(" %d/%d", p.ScrollOffset+1, p.TotalLines)
    }
}
```

### 7.3 Files Changed

| File | Change |
|------|--------|
| `panel.go` | Add `TotalLines`, `ScrollOffset` to `Panel`; render scrollbar in `Render()` |
| `panel_view.go` | Pass scroll state when constructing panels |

---

## Implementation Order

```
Phase 1 (Screen Modes)          ← Foundation, highest impact
  │
  ├─ Phase 2 (Compact Header)   ← Quick win, reclaims vertical space
  │
  ├─ Phase 3 (Responsive)       ← Builds on Phase 1 mode system
  │    │
  │    └─ Phase 3.3 (Page-specific layouts)  ← Builds on responsive + modes
  │
  ├─ Phase 4 (Accordion)        ← Builds on Phase 1 width calculation
  │
  ├─ Phase 6 (Smart Statusbar)  ← Independent, reclaims vertical space
  │
  └─ Phase 7 (Scroll Indicators)← Independent, UX polish
      │
      Phase 5 (Panel Tabs)      ← Most complex, do last
```

### Estimated Scope

| Phase | New/Changed Files | Complexity | Impact |
|-------|-------------------|------------|--------|
| 1 — Screen Modes | 6 files | Medium | High — unlocks panel maximize |
| 2 — Compact Header | 3 files | Low | Medium — reclaims 4+ lines |
| 3 — Responsive | 4 files | Medium | High — small terminals usable |
| 4 — Accordion | 4 files | Low | Medium — better focus UX |
| 5 — Panel Tabs | 5 files | High | High — multiplies panel utility |
| 6 — Smart Statusbar | 3 files | Low | Medium — reclaims 1 line |
| 7 — Scroll Indicators | 2 files | Low | Low — polish |

---

## Keybinding Summary

| Key | Action | Phase |
|-----|--------|-------|
| `+` | Next screen mode | 1 |
| `-` | Previous screen mode | 1 |
| `f` | Toggle fullscreen for focused panel | 1 |
| `=` | Reset to normal mode | 1 |
| `a` | Toggle accordion mode | 4 |
| `[` | Previous tab in panel | 5 |
| `]` | Next tab in panel | 5 |

These keys are chosen to avoid conflicts with existing bindings (`n`/`e`/`d`/`s`/`tab`/`?`/`q`/`/`).

---

## Relationship to Major Refactor Plan

This plan is **compatible with and independent of** `ai/plans/tui/major_refactor.md`:

- **Before refactor**: All changes go into existing `panel_view.go`, `model.go`, `panel.go` — same monolithic structure, just better layout logic
- **During refactor**: When Phase 2-4 of the refactor introduces the `Screen` interface, each screen carries its own `PageLayout` config. The width calculation functions move from global helpers to `AppContext` methods
- **After refactor**: Each `PanelScreen` inherits layout behavior from the base struct, overriding `PageLayout` as needed. Screen modes become a field on `AppContext`, not `Model`

The layout system proposed here becomes the **rendering foundation** that the refactored screens plug into.

---

## Comparison to lazygit/lazydocker

| Feature | lazygit | lazydocker | This Plan |
|---------|---------|------------|-----------|
| Layout engine | gocui coordinates | gocui coordinates | Lipgloss proportional |
| Screen modes | Normal + Accordion | Normal/Half/Full | Normal/Wide/Full |
| Panel maximize | Via screen mode | Via screen mode | `f` toggle |
| Resize handling | gocui `layout()` | gocui `layout()` | `tea.WindowSizeMsg` → recalculate |
| Focus navigation | Context stack (LIFO) | View stack | `FocusPanel` + `FocusKey` (existing) |
| Panel tabs | Window views | Main panel tabs | Per-panel tabs |
| Responsive | Width-aware truncation | Tainted render | Breakpoint tiers |
| Scrollbar | gocui native | gocui native | Custom render in `Panel.Render()` |
| Header | Single line | Single line | Auto: Full/Compact/Hidden |
| Statusbar | Context-sensitive | Context-sensitive | Adaptive 2/1/0 lines |

The key adaptation: lazygit/lazydocker use gocui's absolute coordinate system where the library handles view positioning. We use Bubbletea + Lipgloss where layout is string concatenation — so all spatial logic must be calculated before rendering, not declared as coordinates.
