# CLI_PACKAGE.md

Comprehensive documentation on ModulaCMS's TUI (Terminal User Interface) implementation using Charmbracelet Bubbletea.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md`
**Related Code:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/`
**Framework:** [Charmbracelet Bubbletea](https://github.com/charmbracelet/bubbletea)

---

## Overview

The `internal/cli/` package implements ModulaCMS's SSH-accessible Terminal User Interface (TUI) for content management. It provides a rich, interactive interface for managing content trees, editing fields, and performing administrative tasks—all from the command line.

**Why This Matters:**
- The TUI is the primary interface for content editors and administrators
- Uses Elm Architecture, which is fundamentally different from typical imperative UI code
- Most complex package in the codebase (~3,000+ lines)
- Understanding this is essential for adding new features or screens

**Key Technologies:**
- **Bubbletea** - TUI framework based on Elm Architecture
- **Lipgloss** - Styling and layout
- **Huh** - Form components
- **Bubbles** - Additional UI components (viewport, spinner, etc.)

---

## The Elm Architecture

ModulaCMS's TUI is built on the **Elm Architecture** pattern, which is different from traditional imperative UI programming.

### Imperative vs Elm Architecture

**Traditional Imperative (Not used in ModulaCMS):**
```go
// Imperative - directly mutate UI
func handleKeyPress(key string) {
    if key == "enter" {
        selectedItem.edit()      // Direct action
        ui.refresh()             // Direct UI update
    }
}
```

**Elm Architecture (Used in ModulaCMS):**
```go
// Declarative - return new state and commands
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            return m, m.editSelectedItem()  // Return command
        }
    }
    return m, nil
}
```

### Three Core Components

**1. Model** - Application state
```go
type Model struct {
    contentTree  *TreeRoot
    selectedNode *TreeNode
    cursor       int
    viewport     viewport.Model
}
```

**2. Update** - State transitions
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle messages, return new state
}
```

**3. View** - Render UI
```go
func (m Model) View() string {
    // Return string to display
}
```

### The Message Loop

```
User Input (keyboard, etc.)
    ↓
Message created (tea.KeyMsg, etc.)
    ↓
Update function receives message
    ↓
Update returns (new Model, Command)
    ↓
Command executed (async operations)
    ↓
Command sends new message when done
    ↓
Update receives result message
    ↓
Update returns new Model
    ↓
View function renders new Model
    ↓
Display updated to screen
    ↓
Wait for next user input...
```

**Key insight:** State never mutates in place. Every update creates a new state.

---

## Model: Application State

The Model struct holds all application state.

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/model.go`

### Core Model Structure

```go
type Model struct {
    // Database
    db db.DbDriver

    // Content tree
    contentTree  *TreeRoot
    selectedNode *TreeNode
    cursor       int
    scrollOffset int

    // UI components
    viewport     viewport.Model
    spinner      spinner.Model
    help         help.Model

    // Application state
    mode         AppMode  // BROWSE, EDIT, etc.
    width        int
    height       int
    ready        bool

    // User feedback
    errorMessage   string
    successMessage string
    statusMessage  string

    // Async operation state
    loading bool

    // Route/session
    currentRoute   db.Routes
    currentUser    db.Users
}
```

### AppMode Enum

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/modes.go`

```go
type AppMode int

const (
    ModeBrowse AppMode = iota  // Browsing content tree
    ModeEdit                    // Editing content
    ModeForm                    // Filling out form
    ModeConfirm                 // Confirmation dialog
    ModeHelp                    // Help screen
)
```

**Usage:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.mode {
    case ModeBrowse:
        return m.handleBrowseMode(msg)
    case ModeEdit:
        return m.handleEditMode(msg)
    // ...
    }
}
```

### State Initialization

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/init.go`

```go
func InitialModel(db db.DbDriver, routeID int64) Model {
    return Model{
        db:           db,
        mode:         ModeBrowse,
        cursor:       0,
        scrollOffset: 0,
        loading:      false,
        spinner:      spinner.New(),
        viewport:     viewport.New(80, 24),
        help:         help.New(),
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.spinner.Tick,
        m.loadContentTree(),
        m.loadCurrentRoute(),
    )
}
```

**Key points:**
- `InitialModel` creates default state
- `Init()` returns commands to run on startup
- Commands load data asynchronously

---

## Update: Message Handling

The Update function is the heart of the application—it receives messages and returns new state.

**Function signature:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
```

**Returns:**
- `tea.Model` - New state (modified copy of m)
- `tea.Cmd` - Command to execute (or nil)

### Message Types

**Built-in Bubbletea messages:**
- `tea.KeyMsg` - Keyboard input
- `tea.MouseMsg` - Mouse input (if enabled)
- `tea.WindowSizeMsg` - Terminal resize
- `tea.QuitMsg` - Application quit

**Custom messages (defined in project):**
- `ContentLoadedMsg` - Content tree loaded
- `ContentSavedMsg` - Content saved to database
- `ErrorMsg` - Error occurred
- `NavigationMsg` - Navigate to node
- `StatusChangedMsg` - Status updated

### Update Function Structure

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/model.go`

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle global messages first
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height - 5
        return m, nil

    case tea.KeyMsg:
        // Global key handlers
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        }
    }

    // Delegate to mode-specific handler
    switch m.mode {
    case ModeBrowse:
        return m.updateBrowseMode(msg)
    case ModeEdit:
        return m.updateEditMode(msg)
    case ModeForm:
        return m.updateFormMode(msg)
    }

    return m, nil
}
```

### Mode-Specific Update Handlers

**Browse mode (tree navigation):**

```go
func (m Model) updateBrowseMode(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            m.cursor--
            if m.cursor < 0 {
                m.cursor = 0
            }
            return m, nil

        case "down", "j":
            m.cursor++
            if m.cursor >= len(m.visibleNodes) {
                m.cursor = len(m.visibleNodes) - 1
            }
            return m, nil

        case "enter":
            // Edit selected node
            m.mode = ModeEdit
            return m, m.loadNodeFields(m.selectedNode)

        case "n":
            // Create new node
            return m, m.showCreateForm()

        case "d":
            // Delete node (show confirmation)
            m.mode = ModeConfirm
            m.confirmAction = "delete"
            return m, nil
        }

    case ContentLoadedMsg:
        m.contentTree = msg.Tree
        m.loading = false
        if msg.Error != nil {
            m.errorMessage = msg.Error.Error()
        }
        return m, nil
    }

    return m, nil
}
```

**Edit mode (content editing):**

```go
func (m Model) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "esc":
            // Cancel editing
            m.mode = ModeBrowse
            return m, nil

        case "ctrl+s":
            // Save changes
            return m, m.saveContent()
        }

    case ContentSavedMsg:
        m.mode = ModeBrowse
        m.successMessage = "Content saved"
        if msg.Error != nil {
            m.errorMessage = msg.Error.Error()
        }
        return m, m.loadContentTree()  // Reload tree
    }

    // Delegate to form component
    updatedForm, cmd := m.form.Update(msg)
    m.form = updatedForm
    return m, cmd
}
```

---

## Commands: Async Operations

Commands are functions that return messages. They enable async operations without blocking the UI.

**Command signature:**
```go
type Cmd func() Msg
```

### Creating Commands

**Simple command (no async):**
```go
func showErrorMsg(err error) tea.Cmd {
    return func() tea.Msg {
        return ErrorMsg{Error: err}
    }
}
```

**Async command (database query):**
```go
func (m *Model) loadContentTree() tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()

        // This runs in a goroutine - doesn't block UI
        rows, err := m.db.GetContentTreeByRoute(ctx, m.currentRoute.RouteID)
        if err != nil {
            return ContentLoadedMsg{Error: err}
        }

        // Build tree
        tree := NewTreeRoot()
        stats, err := tree.LoadFromRows(rows)
        if err != nil {
            return ContentLoadedMsg{Error: err}
        }

        return ContentLoadedMsg{
            Tree:  tree,
            Stats: stats,
            Error: nil,
        }
    }
}
```

**Using the command:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.KeyMsg:
        if msg.String() == "r" {  // Refresh
            m.loading = true
            return m, m.loadContentTree()  // Returns command
        }
}
```

### Batching Commands

Execute multiple commands at once:

```go
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.loadContentTree(),
        m.loadDatatypes(),
        m.loadFields(),
        m.spinner.Tick,
    )
}
```

### Sequential Commands

Commands can trigger other commands via messages:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case ContentSavedMsg:
        // After save, reload tree
        return m, m.loadContentTree()

    case ContentLoadedMsg:
        // After load, select first node
        m.cursor = 0
        return m, nil
}
```

---

## View: Rendering the UI

The View function renders the current state to a string.

**Function signature:**
```go
func (m Model) View() string
```

**Returns:** String containing ANSI escape codes for terminal display.

### Main View Structure

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/view.go`

```go
func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }

    // Render based on current mode
    switch m.mode {
    case ModeBrowse:
        return m.renderBrowseView()
    case ModeEdit:
        return m.renderEditView()
    case ModeForm:
        return m.renderFormView()
    case ModeHelp:
        return m.renderHelpView()
    }

    return ""
}
```

### Three-Column Layout

ModulaCMS uses a three-column layout:

```
┌─────────────────────────────────────────────────┐
│ Header (Title, Route, User)                    │
├──────────────┬──────────────────┬───────────────┤
│              │                  │               │
│  Tree View   │  Content Fields  │  Details      │
│  (Column 1)  │  (Column 2)      │  (Column 3)   │
│              │                  │               │
├──────────────┴──────────────────┴───────────────┤
│ Status Bar (Messages, Help)                    │
└─────────────────────────────────────────────────┘
```

**Implementation:**

```go
func (m Model) renderBrowseView() string {
    // Calculate column widths
    col1Width := m.width / 3
    col2Width := m.width / 3
    col3Width := m.width - col1Width - col2Width

    // Render columns
    col1 := m.renderTreeColumn(col1Width)
    col2 := m.renderFieldsColumn(col2Width)
    col3 := m.renderDetailsColumn(col3Width)

    // Join horizontally
    mainContent := lipgloss.JoinHorizontal(
        lipgloss.Top,
        col1,
        col2,
        col3,
    )

    // Add header and footer
    header := m.renderHeader()
    footer := m.renderFooter()

    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        mainContent,
        footer,
    )
}
```

### Styling with Lipgloss

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/styles.go`

```go
var (
    // Color palette
    primaryColor   = lipgloss.Color("86")   // Cyan
    secondaryColor = lipgloss.Color("212")  // Purple
    errorColor     = lipgloss.Color("9")    // Red
    successColor   = lipgloss.Color("10")   // Green

    // Styles
    headerStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("230")).
        Background(primaryColor).
        Bold(true).
        Padding(0, 1)

    selectedStyle = lipgloss.NewStyle().
        Foreground(primaryColor).
        Bold(true)

    errorStyle = lipgloss.NewStyle().
        Foreground(errorColor).
        Bold(true)

    borderStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(primaryColor).
        Padding(1, 2)
)
```

**Using styles:**

```go
func (m Model) renderTreeItem(node *TreeNode, selected bool) string {
    // Build item string
    item := fmt.Sprintf("%s %s", node.Icon, node.Datatype.Label)

    // Apply style
    if selected {
        return selectedStyle.Render(item)
    }
    return item
}
```

### Tree Rendering

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/view_tree.go`

```go
func (m Model) renderTreeColumn(width int) string {
    if m.contentTree == nil {
        return "Loading..."
    }

    lines := []string{}
    m.visibleNodes = []*TreeNode{}

    // Render tree recursively
    m.renderTreeNode(m.contentTree.Root, 0, &lines)

    // Apply viewport scrolling
    start := m.scrollOffset
    end := start + m.viewport.Height
    if end > len(lines) {
        end = len(lines)
    }

    visibleLines := lines[start:end]
    content := strings.Join(visibleLines, "\n")

    // Wrap in border
    return borderStyle.
        Width(width - 4).
        Height(m.viewport.Height).
        Render(content)
}

func (m *Model) renderTreeNode(node *TreeNode, depth int, lines *[]string) {
    if node == nil {
        return
    }

    // Calculate indent
    indent := strings.Repeat("  ", depth)

    // Get icon
    icon := "▸"
    if node.Expand {
        icon = "▾"
    }

    // Build line
    line := fmt.Sprintf("%s%s %s", indent, icon, node.Datatype.Label)

    // Apply selection
    if m.cursor == len(m.visibleNodes) {
        line = selectedStyle.Render(line)
        m.selectedNode = node
    }

    *lines = append(*lines, line)
    m.visibleNodes = append(m.visibleNodes, node)

    // Render children if expanded
    if node.Expand {
        child := node.FirstChild
        for child != nil {
            m.renderTreeNode(child, depth+1, lines)
            child = child.NextSibling
        }
    }
}
```

---

## Message Types

Custom messages enable communication between async operations and the Update function.

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/messages.go`

### Defining Messages

```go
// ContentLoadedMsg is sent when content tree finishes loading
type ContentLoadedMsg struct {
    Tree  *TreeRoot
    Stats *LoadStats
    Error error
}

// ContentSavedMsg is sent when content is saved
type ContentSavedMsg struct {
    ContentDataID int64
    Error         error
}

// NavigationMsg is sent to navigate to a specific node
type NavigationMsg struct {
    NodeID int64
}

// StatusChangedMsg is sent when content status changes
type StatusChangedMsg struct {
    ContentDataID int64
    NewStatus     int32
    Error         error
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
    Error error
}
```

### Message Flow Example

**Saving content:**

1. User presses `ctrl+s`
2. Update creates save command
3. Command runs async (database INSERT/UPDATE)
4. Command returns `ContentSavedMsg`
5. Update receives `ContentSavedMsg`
6. Update changes mode back to Browse
7. Update creates command to reload tree
8. View renders success message

```go
// Step 1-2: User input creates command
case tea.KeyMsg:
    if msg.String() == "ctrl+s" {
        return m, m.saveContent()  // Returns command
    }

// Step 3-4: Command executes and returns message
func (m *Model) saveContent() tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        err := m.db.UpdateContent(ctx, m.editedContent)
        return ContentSavedMsg{
            ContentDataID: m.editedContent.ContentDataID,
            Error:         err,
        }
    }
}

// Step 5-7: Update handles result message
case ContentSavedMsg:
    if msg.Error != nil {
        m.errorMessage = msg.Error.Error()
        return m, nil
    }
    m.mode = ModeBrowse
    m.successMessage = "Content saved"
    return m, m.loadContentTree()  // Reload tree
```

---

## Form Handling with Huh

ModulaCMS uses [Huh](https://github.com/charmbracelet/huh) for forms.

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/forms.go`

### Creating Forms

```go
import "github.com/charmbracelet/huh"

func (m *Model) createContentForm(datatype db.Datatypes, fields []db.Fields) tea.Cmd {
    // Create form groups
    groups := []*huh.Group{}

    // Add field for each datatype field
    for _, field := range fields {
        var input huh.Field

        switch field.Type {
        case "text":
            input = huh.NewInput().
                Key(fmt.Sprintf("field_%d", field.FieldID)).
                Title(field.Label).
                Placeholder("Enter " + field.Label)

        case "richtext":
            input = huh.NewText().
                Key(fmt.Sprintf("field_%d", field.FieldID)).
                Title(field.Label).
                Lines(5)

        case "number":
            input = huh.NewInput().
                Key(fmt.Sprintf("field_%d", field.FieldID)).
                Title(field.Label).
                Validate(func(s string) error {
                    _, err := strconv.ParseFloat(s, 64)
                    return err
                })

        case "boolean":
            input = huh.NewConfirm().
                Key(fmt.Sprintf("field_%d", field.FieldID)).
                Title(field.Label).
                Affirmative("Yes").
                Negative("No")
        }

        groups = append(groups, huh.NewGroup(input))
    }

    // Create form
    form := huh.NewForm(groups...)
    m.form = form
    m.mode = ModeForm

    return form.Init()
}
```

### Handling Form Updates

```go
func (m Model) updateFormMode(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Delegate to form
    form, cmd := m.form.Update(msg)
    m.form = form.(*huh.Form)

    // Check if form completed
    if m.form.State == huh.StateCompleted {
        // Extract values
        values := m.form.GetValues()

        // Save to database
        return m, m.saveFormValues(values)
    }

    // Check if form canceled
    if m.form.State == huh.StateAborted {
        m.mode = ModeBrowse
        return m, nil
    }

    return m, cmd
}
```

---

## External Editor Integration

ModulaCMS supports opening external editors (vim, nano, etc.) for rich text fields.

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/editor.go`

### Opening External Editor

```go
import "os"
import "os/exec"

// OpenEditor opens external editor for field content
func (m *Model) openEditor(fieldValue string) tea.Cmd {
    return func() tea.Msg {
        // Create temp file
        tmpfile, err := os.CreateTemp("", "modulacms-*.md")
        if err != nil {
            return ErrorMsg{Error: err}
        }
        defer os.Remove(tmpfile.Name())

        // Write current content
        if _, err := tmpfile.Write([]byte(fieldValue)); err != nil {
            return ErrorMsg{Error: err}
        }
        tmpfile.Close()

        // Get editor from environment
        editor := os.Getenv("EDITOR")
        if editor == "" {
            editor = "vim"
        }

        // Open editor (blocking)
        cmd := exec.Command(editor, tmpfile.Name())
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr

        if err := cmd.Run(); err != nil {
            return ErrorMsg{Error: err}
        }

        // Read edited content
        content, err := os.ReadFile(tmpfile.Name())
        if err != nil {
            return ErrorMsg{Error: err}
        }

        return EditorClosedMsg{
            Content: string(content),
        }
    }
}
```

**Usage:**
```go
case tea.KeyMsg:
    if msg.String() == "e" {
        // Open editor for current field
        return m, m.openEditor(m.currentFieldValue)
    }

case EditorClosedMsg:
    // Update field with edited content
    m.currentFieldValue = msg.Content
    return m, nil
```

---

## Tree Navigation with Lazy Loading

ModulaCMS implements lazy loading to handle large content trees.

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/tree_navigation.go`

### Expand/Collapse Nodes

```go
func (m Model) handleTreeExpansion(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    node := m.selectedNode
    if node == nil {
        return m, nil
    }

    switch msg.String() {
    case "right", "l":
        // Expand node
        if !node.Expand {
            if node.FirstChild == nil {
                // Children not loaded - load now
                return m, m.loadNodeChildren(node)
            } else {
                // Children already loaded - just expand
                node.Expand = true
                return m, nil
            }
        }

    case "left", "h":
        // Collapse node
        if node.Expand {
            node.Expand = false
            return m, nil
        } else if node.Parent != nil {
            // Navigate to parent
            m.cursor = m.findNodeIndex(node.Parent)
            return m, nil
        }
    }

    return m, nil
}
```

### Loading Children On-Demand

```go
func (m *Model) loadNodeChildren(node *TreeNode) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()

        // Query children from database
        rows, err := m.db.GetContentChildren(ctx, node.Instance.ContentDataID)
        if err != nil {
            return ErrorMsg{Error: err}
        }

        // Build child nodes
        children := []*TreeNode{}
        for _, row := range rows {
            child := NewTreeNodeFromRow(row)
            children = append(children, child)
        }

        return ChildrenLoadedMsg{
            ParentID: node.Instance.ContentDataID,
            Children: children,
            Error:    nil,
        }
    }
}

// Handle loaded children
case ChildrenLoadedMsg:
    parent := m.contentTree.NodeIndex[msg.ParentID]
    if parent != nil {
        // Attach children to parent
        for i, child := range msg.Children {
            if i == 0 {
                parent.FirstChild = child
                child.Parent = parent
            } else {
                prev := msg.Children[i-1]
                prev.NextSibling = child
                child.PrevSibling = prev
                child.Parent = parent
            }
            m.contentTree.NodeIndex[child.Instance.ContentDataID] = child
        }
        parent.Expand = true
    }
    return m, nil
```

---

## Common Patterns

### Pattern 1: Loading Data Asynchronously

**Problem:** Need to fetch data from database without blocking UI.

**Solution:** Create command that returns message with data.

```go
// 1. Define message type
type DataLoadedMsg struct {
    Data  []YourType
    Error error
}

// 2. Create command function
func (m *Model) loadData() tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        data, err := m.db.GetYourData(ctx)
        return DataLoadedMsg{Data: data, Error: err}
    }
}

// 3. Trigger command in Update
case tea.KeyMsg:
    if msg.String() == "r" {
        m.loading = true
        return m, m.loadData()
    }

// 4. Handle result message
case DataLoadedMsg:
    m.loading = false
    if msg.Error != nil {
        m.errorMessage = msg.Error.Error()
        return m, nil
    }
    m.data = msg.Data
    return m, nil
```

### Pattern 2: Mode-Based Navigation

**Problem:** Different screens need different key bindings.

**Solution:** Use mode enum and delegate to mode-specific handlers.

```go
// 1. Define modes
type AppMode int
const (
    ModeList AppMode = iota
    ModeEdit
)

// 2. Add mode to model
type Model struct {
    mode AppMode
}

// 3. Delegate in Update
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.mode {
    case ModeList:
        return m.updateListMode(msg)
    case ModeEdit:
        return m.updateEditMode(msg)
    }
    return m, nil
}

// 4. Mode-specific handlers
func (m Model) updateListMode(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            m.mode = ModeEdit
            return m, nil
        }
    }
    return m, nil
}
```

### Pattern 3: Confirmation Dialogs

**Problem:** Need user confirmation before destructive action.

**Solution:** Create confirmation mode with action callback.

```go
// 1. Add confirmation state
type Model struct {
    confirmAction string
    confirmData   any
}

// 2. Request confirmation
case "d":  // Delete
    m.mode = ModeConfirm
    m.confirmAction = "delete"
    m.confirmData = m.selectedNode
    return m, nil

// 3. Handle confirmation mode
func (m Model) updateConfirmMode(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "y":
            // Execute action
            switch m.confirmAction {
            case "delete":
                node := m.confirmData.(*TreeNode)
                m.mode = ModeBrowse
                return m, m.deleteNode(node)
            }
        case "n", "esc":
            // Cancel
            m.mode = ModeBrowse
            return m, nil
        }
    }
    return m, nil
}
```

### Pattern 4: Error Handling

**Problem:** Show errors to user without crashing.

**Solution:** Store error in model, display in view, clear after delay.

```go
// 1. Add error fields to model
type Model struct {
    errorMessage string
    errorTimer   *time.Timer
}

// 2. Set error when received
case ErrorMsg:
    m.errorMessage = msg.Error.Error()
    return m, m.clearErrorAfterDelay()

case ContentLoadedMsg:
    if msg.Error != nil {
        m.errorMessage = msg.Error.Error()
        return m, m.clearErrorAfterDelay()
    }
    // ... success handling

// 3. Clear error after delay
func (m *Model) clearErrorAfterDelay() tea.Cmd {
    return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
        return ClearErrorMsg{}
    })
}

case ClearErrorMsg:
    m.errorMessage = ""
    return m, nil

// 4. Display in view
func (m Model) renderFooter() string {
    if m.errorMessage != "" {
        return errorStyle.Render("Error: " + m.errorMessage)
    }
    return m.renderHelpText()
}
```

---

## Adding a New Screen

Step-by-step guide to adding a new TUI screen.

### Step 1: Define Mode

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/modes.go`

```go
const (
    ModeBrowse AppMode = iota
    ModeEdit
    ModeMediaBrowser  // NEW MODE
)
```

### Step 2: Add State to Model

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/model.go`

```go
type Model struct {
    // ... existing fields ...

    // Media browser state
    mediaItems      []db.Media
    mediaCursor     int
    mediaFilter     string
}
```

### Step 3: Create Update Handler

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/update_media.go` (new file)

```go
func (m Model) updateMediaBrowserMode(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "esc":
            m.mode = ModeBrowse
            return m, nil

        case "up", "k":
            m.mediaCursor--
            if m.mediaCursor < 0 {
                m.mediaCursor = 0
            }
            return m, nil

        case "down", "j":
            m.mediaCursor++
            if m.mediaCursor >= len(m.mediaItems) {
                m.mediaCursor = len(m.mediaItems) - 1
            }
            return m, nil

        case "enter":
            // Select media item
            selected := m.mediaItems[m.mediaCursor]
            return m, m.selectMedia(selected)

        case "/":
            // Start filter input
            return m, m.startMediaFilter()
        }

    case MediaLoadedMsg:
        m.mediaItems = msg.Items
        if msg.Error != nil {
            m.errorMessage = msg.Error.Error()
        }
        return m, nil
    }

    return m, nil
}
```

### Step 4: Create View Renderer

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/view_media.go` (new file)

```go
func (m Model) renderMediaBrowserView() string {
    // Header
    header := headerStyle.Render("Media Browser")

    // Media list
    items := []string{}
    for i, media := range m.mediaItems {
        item := fmt.Sprintf("%s (%s)", media.Filename, media.Size)
        if i == m.mediaCursor {
            item = selectedStyle.Render("> " + item)
        } else {
            item = "  " + item
        }
        items = append(items, item)
    }

    // Filter bar
    filterBar := ""
    if m.mediaFilter != "" {
        filterBar = fmt.Sprintf("Filter: %s", m.mediaFilter)
    }

    // Combine
    content := strings.Join(items, "\n")
    footer := m.renderHelpText("↑/↓: Navigate | enter: Select | /: Filter | esc: Back")

    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        filterBar,
        content,
        footer,
    )
}
```

### Step 5: Add to Main Update and View

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/model.go`

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ... existing code ...

    switch m.mode {
    case ModeBrowse:
        return m.updateBrowseMode(msg)
    case ModeEdit:
        return m.updateEditMode(msg)
    case ModeMediaBrowser:  // ADD THIS
        return m.updateMediaBrowserMode(msg)
    }

    return m, nil
}

func (m Model) View() string {
    switch m.mode {
    case ModeBrowse:
        return m.renderBrowseView()
    case ModeEdit:
        return m.renderEditView()
    case ModeMediaBrowser:  // ADD THIS
        return m.renderMediaBrowserView()
    }

    return ""
}
```

### Step 6: Add Entry Point

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/update_browse.go`

```go
func (m Model) updateBrowseMode(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        // ... existing keys ...

        case "m":  // NEW KEYBINDING
            m.mode = ModeMediaBrowser
            return m, m.loadMediaItems()
        }
    }
    return m, nil
}
```

---

## Debugging TUI Issues

### Enable Debug Logging

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/model.go`

```go
import "github.com/modula/modulacms/internal/utility"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Log all messages
    utility.DefaultLogger.Debug("TUI message received",
        "type", fmt.Sprintf("%T", msg),
        "value", msg)

    // ... rest of update ...
}
```

### Run with Debug Output

```bash
# Run TUI with debug log to file
./modulacms-x86 --cli 2>debug.log

# In another terminal, tail the log
tail -f debug.log
```

### Common Issues

**Issue 1: Screen not updating**
- **Cause:** Update function not returning new model
- **Fix:** Always return `m` (even if unchanged): `return m, nil`

**Issue 2: Keys not responding**
- **Cause:** Message not handled in current mode
- **Fix:** Check mode-specific update handler includes key

**Issue 3: Async operation not completing**
- **Cause:** Command returns nil instead of message
- **Fix:** Command must return message (even if error)

```go
// BAD - returns nil
func (m *Model) loadData() tea.Cmd {
    return func() tea.Msg {
        data, err := m.db.GetData()
        if err != nil {
            return nil  // BAD - should return ErrorMsg
        }
        return DataLoadedMsg{Data: data}
    }
}

// GOOD - always returns message
func (m *Model) loadData() tea.Cmd {
    return func() tea.Msg {
        data, err := m.db.GetData()
        if err != nil {
            return ErrorMsg{Error: err}  // GOOD
        }
        return DataLoadedMsg{Data: data, Error: nil}
    }
}
```

**Issue 4: Race conditions**
- **Cause:** Multiple goroutines accessing model
- **Fix:** Never mutate model outside Update function

```go
// BAD - mutates model in command
func (m *Model) badCommand() tea.Cmd {
    return func() tea.Msg {
        m.data = fetchData()  // BAD - race condition
        return nil
    }
}

// GOOD - returns data in message
func (m *Model) goodCommand() tea.Cmd {
    return func() tea.Msg {
        data := fetchData()
        return DataMsg{Data: data}  // GOOD
    }
}
```

---

## State Management Best Practices

### 1. Never Mutate Model Outside Update

**Bad:**
```go
func (m *Model) loadData() tea.Cmd {
    return func() tea.Msg {
        m.data = fetchData()  // Mutates model in goroutine
        return nil
    }
}
```

**Good:**
```go
func (m *Model) loadData() tea.Cmd {
    return func() tea.Msg {
        data := fetchData()
        return DataLoadedMsg{Data: data}  // Returns data
    }
}

case DataLoadedMsg:
    m.data = msg.Data  // Mutate only in Update
    return m, nil
```

### 2. Always Return Model and Command

**Bad:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Forgot to return
    m.cursor++
}
```

**Good:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.cursor++
    return m, nil  // Always return
}
```

### 3. Use Batch for Multiple Commands

**Bad:**
```go
func (m Model) Init() tea.Cmd {
    m.loadData()  // Only executes last one
    m.loadUsers()
    return m.loadSettings()
}
```

**Good:**
```go
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.loadData(),
        m.loadUsers(),
        m.loadSettings(),
    )
}
```

### 4. Handle All Error Cases

**Bad:**
```go
case ContentLoadedMsg:
    m.content = msg.Content
    return m, nil
```

**Good:**
```go
case ContentLoadedMsg:
    if msg.Error != nil {
        m.errorMessage = msg.Error.Error()
        return m, m.clearErrorAfterDelay()
    }
    m.content = msg.Content
    return m, nil
```

### 5. Validate State Transitions

**Bad:**
```go
case "e":  // Edit
    m.mode = ModeEdit
    return m, nil
```

**Good:**
```go
case "e":  // Edit
    if m.selectedNode == nil {
        m.errorMessage = "No node selected"
        return m, nil
    }
    m.mode = ModeEdit
    return m, m.loadNodeFields(m.selectedNode)
```

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TUI_ARCHITECTURE.md` - Elm Architecture deep dive
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - Tree data structure
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Content domain model

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/CREATING_TUI_SCREENS.md` - Detailed screen creation
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` - Feature development flow
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/DEBUGGING.md` - Debugging strategies

**Packages:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/MODEL_PACKAGE.md` - Business logic layer

**General:**
- `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md` - Development guidelines
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/FILE_TREE.md` - Project structure

**External Resources:**
- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Huh Documentation](https://github.com/charmbracelet/huh)
- [Elm Architecture Guide](https://guide.elm-lang.org/architecture/)

---

## Quick Reference

### Elm Architecture Components

```go
type Model struct { /* state */ }
func (m Model) Init() tea.Cmd { /* initialize */ }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* handle messages */ }
func (m Model) View() string { /* render UI */ }
```

### Common Message Types

```go
tea.KeyMsg           // Keyboard input
tea.MouseMsg         // Mouse events
tea.WindowSizeMsg    // Terminal resize
tea.QuitMsg          // Quit signal
CustomMsg            // Your custom messages
```

### Command Patterns

```go
// Simple command
func simpleCmd() tea.Msg { return msg }

// Async command
func asyncCmd() tea.Cmd {
    return func() tea.Msg {
        // Do work
        return ResultMsg{}
    }
}

// Batch commands
tea.Batch(cmd1, cmd2, cmd3)
```

### Key Files

```go
internal/cli/
├── model.go          // Model definition
├── init.go           // Initialization
├── update.go         // Update function (main)
├── update_browse.go  // Browse mode updates
├── update_edit.go    // Edit mode updates
├── view.go           // View function (main)
├── view_tree.go      // Tree rendering
├── view_fields.go    // Fields rendering
├── messages.go       // Message types
├── commands.go       // Command functions
├── styles.go         // Lipgloss styles
└── forms.go          // Huh forms
```

### Debugging Commands

```bash
# Run with debug log
./modulacms-x86 --cli 2>debug.log

# Watch debug log
tail -f debug.log

# Run in CLI mode (no SSH)
./modulacms-x86 --cli
```
