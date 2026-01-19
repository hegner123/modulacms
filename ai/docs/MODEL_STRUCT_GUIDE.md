# Model Struct Field Guide

**File:** `internal/cli/model.go`
**Purpose:** Central state container for the ModulaCMS Terminal UI (TUI)
**Last Updated:** 2026-01-15

## Overview

The `Model` struct is the heart of the CLI/TUI application, following the Elm Architecture pattern from Charmbracelet Bubbletea. It contains 39 fields (after Phase 1 & 2 refactoring) organized into logical groups representing different aspects of the application state.

**Evolution:**
- **Original:** 54 fields (mixed concerns)
- **After Phase 1 (FormModel):** 46 fields (8 extracted)
- **After Phase 2 (TableModel):** 39 fields (15 extracted total)
- **Current State:** 39 fields organized into 6 responsibility groups

---

## Field Groups

### 1. Configuration & System (3 fields)

Fields that configure the application and store system-level information.

#### `Config *config.Config`
- **Type:** Pointer to config.Config
- **Purpose:** Application configuration (database, OAuth, S3, ports, domains)
- **Initialization:** Passed to `InitialModel()` from main.go
- **Usage:** Accessed throughout application for database connections, API credentials, and deployment settings
- **Nil Safety:** Never nil after initialization
- **Related Files:** `internal/config/`, `cmd/main.go`

#### `Verbose bool`
- **Type:** Boolean flag
- **Purpose:** Enables verbose logging output
- **Initialization:** Set from command-line flag via `InitialModel(v *bool, ...)`
- **Usage:** Controls debug output verbosity
- **Default:** `false`
- **Related Files:** `internal/utility/logger.go`

#### `Term string`
- **Type:** String
- **Purpose:** Terminal type identifier (e.g., "xterm", "screen")
- **Initialization:** Zero value (empty string)
- **Usage:** Terminal compatibility checks
- **Related Files:** SSH session handling

---

### 2. Display & Styling (9 fields)

Fields that control visual appearance, layout, and UI components.

#### `Width int`
- **Type:** Integer (terminal columns)
- **Purpose:** Current terminal width in characters
- **Initialization:** Zero value, set by Bubbletea WindowSizeMsg
- **Usage:** Layout calculations, text wrapping, responsive rendering
- **Typical Range:** 80-200+ columns
- **Related Messages:** `tea.WindowSizeMsg`

#### `Height int`
- **Type:** Integer (terminal rows)
- **Purpose:** Current terminal height in characters
- **Initialization:** Zero value, set by Bubbletea WindowSizeMsg
- **Usage:** Viewport sizing, pagination calculations
- **Typical Range:** 24-60+ rows
- **Related Messages:** `tea.WindowSizeMsg`

#### `TitleFont int`
- **Type:** Integer (font index)
- **Purpose:** Index into Titles array for ASCII art font selection
- **Initialization:** `0` (first font)
- **Usage:** Selects which ASCII art title to display
- **Range:** 0 to len(Titles)-1
- **Related Files:** `titles/*.txt` (embedded ASCII art)

#### `Titles []string`
- **Type:** String slice
- **Purpose:** Pre-loaded ASCII art titles in different fonts
- **Initialization:** Loaded from embedded `titles/*.txt` files via `LoadTitles()`
- **Usage:** Display application title/logo in different styles
- **Immutable:** Set once during initialization
- **Related Functions:** `ParseTitles()`, `LoadTitles()`

#### `Bg string`
- **Type:** String (color code or name)
- **Purpose:** Background color for terminal UI
- **Initialization:** Zero value (empty string)
- **Usage:** Terminal background color customization
- **Related:** Lipgloss styling system

#### `TxtStyle lipgloss.Style`
- **Type:** Lipgloss style object
- **Purpose:** Text styling (colors, formatting)
- **Initialization:** Zero value (default style)
- **Usage:** Consistent text rendering across UI
- **Related:** `QuitStyle` for quit message styling

#### `QuitStyle lipgloss.Style`
- **Type:** Lipgloss style object
- **Purpose:** Styling for quit/exit messages
- **Initialization:** Zero value (default style)
- **Usage:** Styled "Are you sure you want to quit?" messages
- **Related:** `TxtStyle` for general text styling

#### `Spinner spinner.Model`
- **Type:** Charmbracelet bubbles Spinner
- **Purpose:** Animated loading indicator
- **Initialization:** Created with `spinner.New()`, style set to Dot spinner
- **Usage:** Display during async operations (DB queries, file operations)
- **Start/Stop:** Via `Spinner.Tick()` commands
- **Related:** `Loading` flag

#### `Viewport viewport.Model`
- **Type:** Charmbracelet bubbles Viewport
- **Purpose:** Scrollable content area for long text
- **Initialization:** Zero value `viewport.Model{}`
- **Usage:** Display long content (logs, documentation, query results)
- **Methods:** `SetContent()`, `ViewUp()`, `ViewDown()`
- **Related:** `Content` field for viewport text

---

### 3. Navigation & Pagination (12 fields)

Fields that manage navigation, cursor position, pagination, and page routing.

#### `Cursor int`
- **Type:** Integer (zero-indexed position)
- **Purpose:** Current cursor/selection position in lists, menus, and tables
- **Initialization:** Zero value `0`
- **Usage:** Track which item is selected in any list view
- **Range:** 0 to CursorMax-1
- **Related:** `CursorMax`, navigation key handlers
- **Files:** `internal/cli/update_controls.go`, `internal/cli/update_navigation.go`

#### `CursorMax int`
- **Type:** Integer (maximum position)
- **Purpose:** Maximum valid cursor position (length of current list)
- **Initialization:** `0`
- **Usage:** Bounds checking for cursor movement
- **Calculation:** Set to `len(menu)`, `len(tables)`, etc.
- **Related:** `Cursor`

#### `FocusIndex int`
- **Type:** Integer (zero-indexed)
- **Purpose:** Index of focused element within a component
- **Initialization:** `0`
- **Usage:** Multi-element focus management (e.g., dialog buttons)
- **Related:** `Focus` (component-level focus)

#### `Focus FocusKey`
- **Type:** Enum (PAGEFOCUS, TABLEFOCUS, FORMFOCUS, DIALOGFOCUS)
- **Purpose:** Which major UI component currently has focus
- **Initialization:** `PAGEFOCUS`
- **Usage:** Route keyboard input to correct component
- **Values:**
  - `PAGEFOCUS`: Main page/menu navigation
  - `TABLEFOCUS`: Table/list viewing
  - `FORMFOCUS`: Form input mode
  - `DIALOGFOCUS`: Modal dialog interaction
- **Related:** `FocusIndex`, Update() message routing

#### `Paginator paginator.Model`
- **Type:** Charmbracelet bubbles Paginator
- **Purpose:** Pagination component for long lists
- **Initialization:** Created with `paginator.New()`, configured with dot style
- **Usage:** Visual pagination indicators (•••)
- **Configuration:**
  - Type: `paginator.Dots`
  - ActiveDot: Styled "•"
  - InactiveDot: Styled "•"
- **Methods:** `SetTotalPages()`, `GetSliceBounds()`
- **Related:** `PageMod`, `MaxRows`

#### `PageMod int`
- **Type:** Integer (zero-indexed page number)
- **Purpose:** Current page number for pagination (offset)
- **Initialization:** `0` (first page)
- **Usage:** Calculate which rows/items to display
- **Calculation:** Used with MaxRows to determine slice bounds
- **Related:** `MaxRows`, `Paginator`
- **Files:** `internal/cli/update_controls.go` (pagination key handlers)

#### `MaxRows int`
- **Type:** Integer (rows per page)
- **Purpose:** Maximum number of rows to display per page
- **Initialization:** `10`
- **Usage:** Pagination calculations, determine page bounds
- **Configurable:** Could be changed based on terminal height
- **Related:** `PageMod`, `Paginator`

#### `Page Page`
- **Type:** Page struct (Index + Label)
- **Purpose:** Current active page/screen
- **Initialization:** `NewPage(HOMEPAGE, "Home")`
- **Usage:** Determine which view to render
- **Fields:**
  - `Index` (PageIndex enum)
  - `Label` (string)
- **Related:** `PageMenu`, `PageMap`, `History`
- **Files:** `internal/cli/pages.go`, `internal/cli/view.go`

#### `PageMenu []Page`
- **Type:** Slice of Page structs
- **Purpose:** Current menu options available to user
- **Initialization:** `m.HomepageMenuInit()`
- **Usage:** Render menu items, cursor navigation bounds
- **Dynamic:** Changes based on current page context
- **Related:** `Cursor`, `CursorMax`, `Page`
- **Files:** `internal/cli/view.go`, menu initialization functions

#### `Pages []Page`
- **Type:** Slice of Page structs
- **Purpose:** All available pages in the application
- **Initialization:** Zero value (empty slice)
- **Usage:** Store complete list of pages (less commonly used than PageMap)
- **Related:** `PageMap` (preferred for page lookup)

#### `PageMap map[PageIndex]Page`
- **Type:** Map from PageIndex enum to Page struct
- **Purpose:** Fast lookup of pages by index
- **Initialization:** `*InitPages()` - creates map of all pages
- **Usage:** Navigate to specific page by index
- **Keys:** PageIndex enum (HOMEPAGE, CMSPAGE, DATABASEPAGE, etc.)
- **Related:** `Page`, `Pages`
- **Files:** `internal/cli/pages.go` (`InitPages()`)

#### `History []PageHistory`
- **Type:** Slice of PageHistory structs
- **Purpose:** Navigation history stack (back button functionality)
- **Initialization:** Empty slice `[]PageHistory{}`
- **Usage:** Store previous page states for back navigation
- **Operations:**
  - `PushHistory()`: Add current state before navigation
  - `PopHistory()`: Return to previous page
- **Fields (PageHistory):**
  - `Cursor` (int): Previous cursor position
  - `Page` (Page): Previous page
  - `Menu` ([]Page): Previous menu state
- **Files:** `internal/cli/history.go`

---

### 4. Database & CMS (5 fields)

Fields specific to database operations and CMS content management.

#### `Tables []string`
- **Type:** String slice
- **Purpose:** List of all database table names
- **Initialization:** Fetched via `GetTablesCMD()` in InitialModel
- **Usage:** Populate database menu, table selection
- **Dynamic:** Loaded from database schema at startup
- **Related:** `TableState.Table` (currently selected table)
- **Files:** `internal/cli/commands.go`, database menu views

#### `DatatypeMenu []string`
- **Type:** String slice
- **Purpose:** Menu options for CMS datatype management
- **Initialization:** Zero value (empty slice)
- **Usage:** Display datatype-related actions
- **CMS Specific:** Part of content type management
- **Related:** CMS page navigation

#### `PageRouteId int64`
- **Type:** 64-bit integer (database ID)
- **Purpose:** CMS page route identifier
- **Initialization:** Zero value `0`
- **Usage:** Track which CMS page/route is being edited
- **CMS Specific:** Maps to content_routes table
- **Related:** `Root` (content tree)

#### `QueryResults []sql.Row`
- **Type:** Slice of database Row objects
- **Purpose:** Store results from custom SQL queries
- **Initialization:** Zero value (nil slice)
- **Usage:** Display query results, debugging
- **Related:** Database debugging features
- **Files:** Query execution commands

#### `Root TreeRoot`
- **Type:** TreeRoot struct (CMS content tree)
- **Purpose:** CMS content hierarchy (tree structure)
- **Initialization:** Zero value
- **Usage:** Navigate content tree, manage parent-child relationships
- **Fields (TreeRoot):**
  - `Root` (*TreeNode): Root node of tree
  - `NodeIndex` (map[int64]*TreeNode): Fast node lookup
  - `Orphans` (map[int64]*TreeNode): Unresolved nodes
  - `MaxRetry` (int): Orphan resolution retry limit
- **Related:** `PageRouteId`, CMS content management
- **Files:** `internal/cli/cms_struct.go`, tree loading functions

---

### 5. UI Components (Sub-Models) (4 fields)

Extracted sub-models that encapsulate specific UI responsibilities.

#### `FormState *FormModel`
- **Type:** Pointer to FormModel struct
- **Purpose:** All form-related state (Phase 1 extraction)
- **Initialization:** `NewFormModel()` in InitialModel
- **Nil Safety:** Never nil after initialization
- **Fields (FormModel):**
  - `Form` (*huh.Form): Charmbracelet Huh form instance
  - `FormLen` (int): Number of fields in form
  - `FormMap` ([]string): Map of field names
  - `FormValues` ([]*string): Pointers to field values
  - `FormSubmit` (bool): Form submission status
  - `FormGroups` ([]huh.Group): Form field groups
  - `FormFields` ([]huh.Field): Individual form fields
  - `FormOptions` (*FormOptionsMap): Dropdown/select options
- **Usage:** Create, display, submit forms for DB/CMS operations
- **Related Files:** `internal/cli/form_model.go`, `internal/cli/form.go`
- **Extracted:** Phase 1 refactoring (54 → 46 fields)

#### `TableState *TableModel`
- **Type:** Pointer to TableModel struct
- **Purpose:** All table-related state (Phase 2 extraction)
- **Initialization:** `NewTableModel()` in InitialModel
- **Nil Safety:** Never nil after initialization
- **Fields (TableModel):**
  - `Table` (string): Current table name
  - `Headers` ([]string): Column headers for display
  - `Rows` ([][]string): Table data rows
  - `Columns` (*[]string): Column names from database
  - `ColumnTypes` (*[]*sql.ColumnType): Column metadata
  - `Selected` (map[int]struct{}): Multi-select state
  - `Row` (*[]string): Currently selected row data
- **Usage:** Display database tables, navigate rows, populate forms
- **Related Files:** `internal/cli/table_model.go`, `internal/cli/view.go`
- **Extracted:** Phase 2 refactoring (46 → 39 fields)

#### `Dialog *DialogModel`
- **Type:** Pointer to DialogModel struct
- **Purpose:** Modal dialog state
- **Initialization:** Created on-demand via `ShowDialog()`
- **Nil Safety:** Check `DialogActive` flag before accessing
- **Fields (DialogModel):**
  - `Title` (string): Dialog title text
  - `Message` (string): Dialog body message
  - `Width`, `Height` (int): Dialog dimensions
  - `OkText`, `CancelText` (string): Button labels
  - `ShowCancel` (bool): Show cancel button
  - `ReadyOK` (bool): OK button selected
  - `Action` (DialogAction): Dialog purpose (generic/delete)
  - `focusIndex` (int): Internal button focus
- **Usage:** Confirmations, warnings, delete confirmations
- **Related:** `DialogActive` flag
- **Files:** `internal/cli/dialog.go`, `internal/cli/update_dialog.go`

#### `DialogActive bool`
- **Type:** Boolean flag
- **Purpose:** Indicates if a dialog is currently displayed
- **Initialization:** `false`
- **Usage:** Route input to dialog, render dialog overlay
- **Pairing:** Used with `Dialog` pointer
- **Pattern:** `if m.DialogActive && m.Dialog != nil { ... }`
- **Related:** `Focus` (DIALOGFOCUS), dialog message handlers

---

### 6. Application State (6 fields)

Fields that track application status, errors, and timing.

#### `Status ApplicationState`
- **Type:** Enum (OK, EDITING, DELETING, WARN, ERROR)
- **Purpose:** Current application operational state
- **Initialization:** `OK`
- **Usage:** Display status indicator, conditional rendering
- **Values:**
  - `OK`: Normal operation (green indicator)
  - `EDITING`: Edit mode active (yellow indicator)
  - `DELETING`: Delete operation pending (red, blinking)
  - `WARN`: Warning state (yellow indicator)
  - `ERROR`: Error state (red, blinking)
- **Rendering:** `GetStatus()` method returns styled status string
- **Files:** `internal/cli/status.go`

#### `Profile string`
- **Type:** String identifier
- **Purpose:** User profile or session identifier
- **Initialization:** Zero value (empty string)
- **Usage:** Track user sessions (SSH), profile-specific settings
- **Related:** SSH authentication, multi-user support

#### `Loading bool`
- **Type:** Boolean flag
- **Purpose:** Indicates async operation in progress
- **Initialization:** `false`
- **Usage:** Show spinner, disable input during operations
- **Related:** `Spinner` component
- **Pattern:** Set `Loading = true`, return spinner.Tick() command

#### `Ready bool`
- **Type:** Boolean flag
- **Purpose:** Application initialization complete
- **Initialization:** `false`
- **Usage:** Prevent rendering until initialization done
- **Pattern:** Set after initial data loads (tables, config)

#### `Content string`
- **Type:** String (potentially long text)
- **Purpose:** Content to display in viewport
- **Initialization:** Zero value (empty string)
- **Usage:** Logs, documentation, long-form text display
- **Related:** `Viewport` (scrollable content area)
- **Files:** Content viewing pages

#### `Err error`
- **Type:** Error interface
- **Purpose:** Last error encountered
- **Initialization:** `nil`
- **Usage:** Display error messages, error handling
- **Nil Safety:** Always check `if m.Err != nil`
- **Related:** `Status` (ERROR state), error views
- **Pattern:** Set error and change Status to ERROR

#### `Time time.Time`
- **Type:** Go time.Time struct
- **Purpose:** Timestamp for operations or last update
- **Initialization:** Zero value (0001-01-01 00:00:00 UTC)
- **Usage:** Track operation timing, display timestamps
- **Related:** Performance monitoring, logging

---

## Key Methods

### Initialization

#### `InitialModel(v *bool, c *config.Config) (Model, tea.Cmd)`
- **Purpose:** Create initial model state
- **Parameters:**
  - `v`: Verbose flag pointer (optional)
  - `c`: Application configuration
- **Returns:** Initialized Model and initial command(s)
- **Process:**
  1. Load ASCII art titles from embedded files
  2. Configure paginator (dots style)
  3. Configure spinner (dot style)
  4. Initialize all fields with defaults
  5. Create FormState and TableState sub-models
  6. Initialize homepage menu
  7. Return GetTablesCMD to fetch database tables
- **File:** model.go:107

#### `ModelPostInit(m Model) tea.Cmd`
- **Purpose:** Post-initialization commands
- **Returns:** Batch of commands to run after initial setup
- **Usage:** Additional async initialization steps
- **File:** model.go:159

### Status Management

#### `GetStatus() string`
- **Purpose:** Render styled status indicator
- **Returns:** Lipgloss-styled status string
- **Behavior:**
  - OK → Green "  OK  "
  - EDITING → Yellow " EDIT " (bold)
  - DELETING → Red "DELETE" (bold, blinking)
  - WARN → Yellow " WARN " (bold)
  - ERROR → Red "ERROR " (bold, blinking)
- **File:** model.go:195

### ModelInterface Implementation

#### `GetConfig() *config.Config`
- **Purpose:** Retrieve application configuration
- **Returns:** Config pointer
- **Interface:** Implements cms.ModelInterface
- **File:** model.go:216

#### Additional ModelInterface methods (implied):
- `GetRoot() model.Root`
- `SetRoot(root model.Root)`
- `SetError(err error)`

---

## Navigation & History

### History Management

#### `PushHistory(entry PageHistory)`
- **Purpose:** Save current page state before navigation
- **Parameters:** PageHistory (Cursor, Page, Menu)
- **Usage:** Call before navigating to new page
- **File:** internal/cli/history.go:9

#### `PopHistory() *PageHistory`
- **Purpose:** Return to previous page state
- **Returns:** Previous PageHistory or nil if history empty
- **Usage:** Back button functionality
- **File:** internal/cli/history.go:13

---

## Field Relationship Map

### Groupings by Common Usage

**Navigation Flow:**
```
Cursor ↔ CursorMax ↔ PageMenu → Page → History
         ↓
    PageMod ↔ MaxRows ↔ Paginator
```

**Table Viewing:**
```
Tables → TableState.Table → TableState.Columns → TableState.Rows
                ↓                    ↓
         TableState.Row      FormState.Form
```

**Form Workflow:**
```
TableState.Columns → FormState.Form → FormState.FormValues → FormState.FormSubmit
                               ↓
                      FormState.FormOptions
```

**Dialog Flow:**
```
DialogActive ← ShowDialog() → Dialog.Update() → DialogAction
```

**Display Pipeline:**
```
Width, Height → Viewport → Content
                      ↓
              Paginator.GetSliceBounds()
```

---

## Common Access Patterns

### Safe Pointer Access

```go
// FormState (always initialized, never nil)
if m.FormState.Form != nil {
    // Safe to access form
}

// TableState (always initialized, never nil)
if m.TableState.Columns != nil {
    columns := *m.TableState.Columns
    // Safe to dereference
}

// Dialog (conditionally initialized)
if m.DialogActive && m.Dialog != nil {
    // Safe to access dialog
}
```

### Navigation Pattern

```go
// Before navigating away
m.PushHistory(PageHistory{
    Cursor: m.Cursor,
    Page:   m.Page,
    Menu:   m.PageMenu,
})

// Navigate to new page
m.Page = m.PageMap[DATABASEPAGE]
m.PageMenu = m.DatabaseMenuInit()
m.Cursor = 0
m.CursorMax = len(m.PageMenu)
```

### Pagination Pattern

```go
// Update pagination
m.Paginator.SetTotalPages(len(m.TableState.Rows))
start, end := m.Paginator.GetSliceBounds(len(m.TableState.Rows))
visibleRows := m.TableState.Rows[start:end]
```

### Status Update Pattern

```go
// Set status with error
m.Status = ERROR
m.Err = fmt.Errorf("database connection failed")

// Clear error
m.Status = OK
m.Err = nil
```

---

## Refactoring History

### Phase 1: FormModel Extraction (2026-01-12)
**Extracted Fields (8):**
- Form, FormLen, FormMap, FormValues, FormSubmit, FormGroups, FormFields, FormOptions

**Result:** 54 → 46 fields (15% reduction)

**Commit:** b0b4ff3

### Phase 2: TableModel Extraction (2026-01-13)
**Extracted Fields (7):**
- Table, Headers, Rows, Columns, ColumnTypes, Selected, Row

**Result:** 46 → 39 fields (15% reduction)

**Commit:** 008850b

### Future Phases (Planned)

**Phase 3: NavigationModel** (candidates):
- Cursor, CursorMax, Page, PageMenu, PageMod, History, Focus, FocusIndex

**Phase 4: DisplayModel** (candidates):
- Width, Height, TitleFont, Titles, Spinner, Viewport, Loading

**End Goal:** ~20-25 core fields + 4-5 sub-models

---

## Best Practices

### Adding New Fields

**Before adding a field, consider:**
1. Does it belong to an existing sub-model (FormState, TableState, Dialog)?
2. Is it navigation-related (future NavigationModel)?
3. Is it display-related (future DisplayModel)?
4. Or is it truly core Model state?

**If adding to Model:**
- Add to appropriate group in struct definition
- Document in this guide
- Initialize in InitialModel if needed
- Add to debug.go for debugging visibility

### Accessing Sub-Models

**Always use nil checks for pointer fields within sub-models:**
```go
// WRONG
columns := *m.TableState.Columns // Can panic if nil

// RIGHT
if m.TableState.Columns != nil {
    columns := *m.TableState.Columns
}
```

**Sub-models themselves are always initialized:**
```go
// Safe (FormState/TableState never nil after InitialModel)
m.FormState.FormSubmit = true
m.TableState.Table = "users"

// Need nil check (Dialog created on-demand)
if m.Dialog != nil {
    m.Dialog.ReadyOK = true
}
```

---

## Related Documentation

- **[CLI_PACKAGE.md](CLI_PACKAGE.md)** - TUI architecture and Elm pattern
- **[MODEL_PACKAGE.md](MODEL_PACKAGE.md)** - Business logic model package
- **[FORM_REFACTOR_PLAN.md](../FORM_REFACTOR_PLAN.md)** - Phase 1 extraction details
- **[TABLE_REFACTOR_PLAN.md](../TABLE_REFACTOR_PLAN.md)** - Phase 2 extraction details
- **[TUI_ARCHITECTURE.md](../architecture/TUI_ARCHITECTURE.md)** - Overall TUI design

---

**Last Updated:** 2026-01-15
**Model Version:** Post-Phase 2 (39 fields)
