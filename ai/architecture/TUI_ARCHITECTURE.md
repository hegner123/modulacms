# TUI_ARCHITECTURE.md

Deep dive into ModulaCMS's Terminal User Interface architecture using the Elm Architecture pattern.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TUI_ARCHITECTURE.md`
**Related Code:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/`
**Framework:** [Charmbracelet Bubbletea](https://github.com/charmbracelet/bubbletea)

---

## Overview

ModulaCMS's TUI is built using the **Elm Architecture**, a functional reactive programming pattern that provides predictable state management through immutable data structures and pure functions.

**Why This Matters:**
- Fundamentally different from traditional imperative UI programming
- State is never mutated in place—always returns new state
- Side effects are isolated in Commands
- Message-driven architecture enables complex async operations
- Easier to reason about, test, and debug than stateful callbacks

**Key Principle:** The application is a pure function of state. Given the same state, you always get the same view.

---

## What is the Elm Architecture?

The Elm Architecture is a pattern for building interactive applications. It originated in the Elm programming language but has been adopted in various forms across ecosystems (Redux, Flux, etc.).

### Core Principles

**1. Single Source of Truth**
- All application state lives in one Model struct
- No hidden state, no global variables
- State is explicit and visible

**2. Immutable State**
- State never changes in place
- Every update creates a new state
- Previous state remains unchanged
- Enables time-travel debugging, undo/redo

**3. Pure Functions**
- Update function is deterministic
- Same input always produces same output
- No side effects in Update function
- Side effects isolated in Commands

**4. Unidirectional Data Flow**
```
User Input → Message → Update → New State → View → Render → Wait for Input...
```

### Why Elm Architecture for TUI?

**Advantages over imperative approaches:**

| Aspect | Imperative | Elm Architecture |
|--------|-----------|------------------|
| State management | Scattered across objects | Single Model struct |
| Updates | Mutate state directly | Return new state |
| Side effects | Mixed with logic | Isolated in Commands |
| Async operations | Callbacks, promises | Message-driven |
| Debugging | Hard to trace state | Clear message log |
| Testing | Mock dependencies | Pure functions |
| Race conditions | Common | Rare (message queue) |

---

## The Three Core Components

### 1. Model: Application State

The Model is a struct containing ALL application state.

**Example from ModulaCMS:**
```go
type Model struct {
    // Database connection
    db db.DbDriver

    // Content tree state
    contentTree  *TreeRoot
    selectedNode *TreeNode
    cursor       int
    scrollOffset int

    // UI components
    viewport viewport.Model
    spinner  spinner.Model
    form     *huh.Form

    // Application mode
    mode AppMode

    // UI state
    width  int
    height int
    ready  bool

    // User feedback
    errorMessage   string
    successMessage string

    // Async state
    loading bool
}
```

**Key characteristics:**
- Everything needed to render UI
- No hidden state
- Self-contained
- Can be serialized (in theory)

### 2. Update: State Transitions

The Update function receives messages and returns new state plus optional commands.

**Function signature:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
```

**Key points:**
- Takes old state (m) and message
- Returns new state and optional command
- Does NOT mutate m
- Pure function (except command execution)

**Example:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "j", "down":
            // Create new state (not mutate)
            newModel := m
            newModel.cursor++
            if newModel.cursor >= len(m.visibleNodes) {
                newModel.cursor = len(m.visibleNodes) - 1
            }
            return newModel, nil

        case "enter":
            // Return command for side effect
            m.mode = ModeEdit
            return m, m.loadNodeFields()
        }

    case ContentLoadedMsg:
        // Handle async result
        m.contentTree = msg.Tree
        m.loading = false
        return m, nil
    }

    return m, nil
}
```

### 3. View: Rendering

The View function takes state and returns a string representation.

**Function signature:**
```go
func (m Model) View() string
```

**Key points:**
- Pure function of state
- Same state always produces same output
- No side effects
- Returns ANSI-formatted string

**Example:**
```go
func (m Model) View() string {
    if !m.ready {
        return "Loading..."
    }

    // Build UI from current state
    header := m.renderHeader()
    content := m.renderContent()
    footer := m.renderFooter()

    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        content,
        footer,
    )
}
```

---

## The Message Loop

The heart of the Elm Architecture is the message loop.

### Message Flow Diagram

```
┌─────────────────────────────────────────────────────┐
│                  Bubbletea Runtime                  │
└─────────────────────────────────────────────────────┘
                         │
                         ↓
            ┌────────────────────────┐
            │   User Input Event     │
            │  (keyboard, mouse,     │
            │   window resize, etc.) │
            └────────────────────────┘
                         │
                         ↓
            ┌────────────────────────┐
            │   Create Message       │
            │   (tea.KeyMsg, etc.)   │
            └────────────────────────┘
                         │
                         ↓
            ┌────────────────────────┐
            │   Update(msg)          │
            │   - Receive message    │
            │   - Calculate new state│
            │   - Return (Model, Cmd)│
            └────────────────────────┘
                    │         │
                    │         └────────────────┐
                    ↓                          ↓
        ┌───────────────────┐      ┌──────────────────┐
        │   New Model       │      │   Command (Cmd)  │
        └───────────────────┘      │   (if any)       │
                    │              └──────────────────┘
                    │                       │
                    ↓                       ↓
        ┌───────────────────┐      ┌──────────────────┐
        │   View()          │      │  Execute Command │
        │   - Render state  │      │  (goroutine)     │
        │   - Return string │      └──────────────────┘
        └───────────────────┘               │
                    │                       │
                    ↓                       ↓
        ┌───────────────────┐      ┌──────────────────┐
        │   Display to      │      │  Command sends   │
        │   Terminal        │      │  new message     │
        └───────────────────┘      └──────────────────┘
                                            │
                    ┌───────────────────────┘
                    ↓
        ┌────────────────────────┐
        │   Back to Update(msg)  │
        │   with result message  │
        └────────────────────────┘
```

### Step-by-Step Execution

**1. Initial State**
```go
func main() {
    // Create initial model
    initialModel := InitialModel(db, routeID)

    // Start Bubbletea program
    program := tea.NewProgram(initialModel)
    program.Run()
}
```

**2. Runtime Starts**
- Bubbletea calls `initialModel.Init()`
- Init returns initial commands to execute
- Runtime enters message loop

**3. User Presses Key**
- OS sends keyboard event
- Bubbletea receives event
- Creates `tea.KeyMsg` with key info
- Sends message to Update function

**4. Update Processes Message**
```go
case tea.KeyMsg:
    if msg.String() == "r" {
        m.loading = true
        return m, m.loadContentTree()
    }
```
- Receives message
- Updates state (m.loading = true)
- Returns command (loadContentTree)

**5. View Renders New State**
```go
func (m Model) View() string {
    if m.loading {
        return m.spinner.View() + " Loading..."
    }
    return m.renderContent()
}
```
- Bubbletea calls View() with new model
- View returns spinner because loading=true
- Terminal displays spinner

**6. Command Executes**
```go
func (m *Model) loadContentTree() tea.Cmd {
    return func() tea.Msg {
        // This runs in goroutine
        data := fetchFromDatabase()
        return ContentLoadedMsg{Data: data}
    }
}
```
- Command runs in separate goroutine
- Doesn't block UI
- Returns message when complete

**7. Command Result Arrives**
```go
case ContentLoadedMsg:
    m.contentTree = msg.Data
    m.loading = false
    return m, nil
```
- Update receives ContentLoadedMsg
- Updates state with data
- Sets loading=false
- View re-renders with content

**8. Loop Continues**
- Wait for next user input or message
- Repeat forever until quit

---

## Commands: Handling Side Effects

Commands are the ONLY way to perform side effects in Elm Architecture.

### What Are Commands?

**Definition:**
```go
type Cmd func() Msg
```

A Command is a function that:
1. Executes some side effect (database query, HTTP request, file I/O)
2. Returns a message with the result
3. Runs in a goroutine (non-blocking)

### Command Patterns

**Pattern 1: Simple Command (No Async)**
```go
func showSuccessMsg(text string) tea.Cmd {
    return func() tea.Msg {
        return SuccessMsg{Text: text}
    }
}

// Usage
return m, showSuccessMsg("Content saved")
```

**Pattern 2: Async Database Query**
```go
func (m *Model) loadContent() tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()

        // This runs in goroutine - doesn't block UI
        data, err := m.db.GetContentTree(ctx, m.routeID)

        return ContentLoadedMsg{
            Data:  data,
            Error: err,
        }
    }
}

// Usage
case "r":  // Refresh key
    m.loading = true
    return m, m.loadContent()
```

**Pattern 3: Chaining Commands**
```go
// Save content, then reload tree
case "ctrl+s":
    return m, m.saveContent()

case ContentSavedMsg:
    if msg.Error != nil {
        m.errorMessage = msg.Error.Error()
        return m, nil
    }
    // Chain: after save, reload tree
    return m, m.loadContentTree()

case ContentLoadedMsg:
    m.contentTree = msg.Tree
    return m, nil
```

**Pattern 4: Batch Commands**
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
Executes multiple commands concurrently.

**Pattern 5: Delayed Command**
```go
func (m *Model) clearErrorAfterDelay() tea.Cmd {
    return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
        return ClearErrorMsg{}
    })
}

// Usage
case ErrorOccurred:
    m.errorMessage = "Something went wrong"
    return m, m.clearErrorAfterDelay()

case ClearErrorMsg:
    m.errorMessage = ""
    return m, nil
```

### Why Commands Instead of Direct Calls?

**Bad (Imperative):**
```go
case "r":
    m.loading = true
    data := m.db.GetContentTree()  // Blocks UI!
    m.contentTree = data
    m.loading = false
    return m, nil
```

**Problems:**
- Blocks UI during database query
- Can't handle errors properly
- State mutations scattered
- Hard to test

**Good (Elm Architecture):**
```go
case "r":
    m.loading = true
    return m, m.loadContent()  // Returns command

case ContentLoadedMsg:
    m.loading = false
    if msg.Error != nil {
        m.errorMessage = msg.Error.Error()
        return m, nil
    }
    m.contentTree = msg.Data
    return m, nil
```

**Benefits:**
- UI stays responsive
- Clear error handling
- State updates in one place
- Easy to test (mock messages)
- Can retry, timeout, etc.

---

## Messages: Communication Protocol

Messages are the communication protocol between async operations and the Update function.

### Message Type Design

**Good message design:**
```go
// Include all necessary data
type ContentLoadedMsg struct {
    Tree  *TreeRoot
    Stats *LoadStats
    Error error        // Always include error
}

// Include context
type NodeDeletedMsg struct {
    NodeID      int64
    ParentID    int64  // Useful for re-navigation
    WasSelected bool   // Useful for cursor management
    Error       error
}

// Specific, not generic
type PublishSuccessMsg struct {
    ContentID int64
}
type ArchiveSuccessMsg struct {
    ContentID int64
}
// Better than generic "ActionCompletedMsg"
```

**Bad message design:**
```go
// Too generic
type DataMsg struct {
    Data interface{}  // What data?
}

// Missing error
type LoadedMsg struct {
    Data []byte  // What if load failed?
}

// Not enough context
type SavedMsg struct{}  // What was saved? Where?
```

### Message Flow Examples

**Example 1: Publishing Content**

```go
// 1. User presses 'p'
case tea.KeyMsg:
    if msg.String() == "p" {
        return m, m.publishContent(m.selectedNode.ID)
    }

// 2. Command executes
func (m *Model) publishContent(id int64) tea.Cmd {
    return func() tea.Msg {
        err := m.db.UpdateStatus(ctx, id, StatusPublished)
        return ContentPublishedMsg{
            ContentID: id,
            Error:     err,
        }
    }
}

// 3. Update handles result
case ContentPublishedMsg:
    if msg.Error != nil {
        m.errorMessage = msg.Error.Error()
        return m, m.clearErrorAfterDelay()
    }

    // Update local state
    node := m.contentTree.NodeIndex[msg.ContentID]
    if node != nil {
        node.Instance.Status = StatusPublished
    }

    m.successMessage = "Content published"
    return m, m.clearSuccessAfterDelay()
```

**Example 2: Form Submission**

```go
// 1. Form completed
case huh.StateCompletedMsg:
    values := m.form.GetValues()
    return m, m.saveFormValues(values)

// 2. Command saves to DB
func (m *Model) saveFormValues(values map[string]string) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()

        // Create content
        contentID, err := m.db.CreateContent(ctx, ...)
        if err != nil {
            return FormSavedMsg{Error: err}
        }

        // Create field values
        for fieldID, value := range values {
            err := m.db.CreateFieldValue(ctx, contentID, fieldID, value)
            if err != nil {
                return FormSavedMsg{Error: err}
            }
        }

        return FormSavedMsg{
            ContentID: contentID,
            Error:     nil,
        }
    }
}

// 3. Update handles result
case FormSavedMsg:
    if msg.Error != nil {
        m.errorMessage = "Save failed: " + msg.Error.Error()
        return m, nil  // Stay in form mode
    }

    m.mode = ModeBrowse
    m.successMessage = "Content created"
    return m, tea.Batch(
        m.loadContentTree(),  // Reload tree
        m.clearSuccessAfterDelay(),
    )
```

---

## State Management Patterns

### Pattern 1: Mode-Based State

Use enum to represent different application states.

```go
type AppMode int

const (
    ModeBrowse AppMode = iota
    ModeEdit
    ModeForm
    ModeConfirm
    ModeHelp
)

type Model struct {
    mode AppMode
    // ... other fields
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Global handlers first
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    }

    // Delegate to mode-specific handler
    switch m.mode {
    case ModeBrowse:
        return m.updateBrowseMode(msg)
    case ModeEdit:
        return m.updateEditMode(msg)
    // ...
    }

    return m, nil
}
```

**Benefits:**
- Clear state separation
- Different key bindings per mode
- Easy to add new modes
- Type-safe mode checking

### Pattern 2: Loading State

Track async operations with boolean flags.

```go
type Model struct {
    loading          bool
    loadingContent   bool
    loadingFields    bool
    savingContent    bool
}

// Show loading indicator
func (m Model) View() string {
    if m.loading {
        return m.spinner.View() + " Loading..."
    }
    // ... normal view
}

// Prevent actions while loading
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.KeyMsg:
        if m.loading {
            return m, nil  // Ignore input while loading
        }
        // ... handle input
}
```

### Pattern 3: Validation State

Track validation before state transitions.

```go
type Model struct {
    formValues    map[string]string
    formErrors    map[string]string
    formValid     bool
}

func (m Model) validateForm() Model {
    m.formErrors = make(map[string]string)

    if m.formValues["title"] == "" {
        m.formErrors["title"] = "Title is required"
    }

    if len(m.formValues["title"]) > 200 {
        m.formErrors["title"] = "Title too long"
    }

    m.formValid = len(m.formErrors) == 0
    return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.KeyMsg:
        if msg.String() == "ctrl+s" {
            m = m.validateForm()
            if !m.formValid {
                return m, nil  // Don't save if invalid
            }
            return m, m.saveContent()
        }
}
```

### Pattern 4: Cursor and Selection

Track current selection with index.

```go
type Model struct {
    items         []Item
    cursor        int
    selectedIndex int
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.KeyMsg:
        switch msg.String() {
        case "j", "down":
            m.cursor++
            if m.cursor >= len(m.items) {
                m.cursor = len(m.items) - 1
            }

        case "k", "up":
            m.cursor--
            if m.cursor < 0 {
                m.cursor = 0
            }

        case "enter":
            m.selectedIndex = m.cursor
        }

    return m, nil
}

func (m Model) View() string {
    for i, item := range m.items {
        line := item.Title

        // Highlight cursor
        if i == m.cursor {
            line = selectedStyle.Render("> " + line)
        }

        // Mark selected
        if i == m.selectedIndex {
            line = line + " ✓"
        }
    }
}
```

---

## Advanced Patterns

### Pattern 1: Subscriptions

Subscriptions are commands that continuously send messages.

```go
// Tick every second
func tickSubscription() tea.Cmd {
    return tea.Every(time.Second, func(t time.Time) tea.Msg {
        return TickMsg{Time: t}
    })
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.loadContent(),
        tickSubscription(),  // Continuous ticks
    )
}

// Handle ticks
case TickMsg:
    m.currentTime = msg.Time
    return m, nil
```

### Pattern 2: Debouncing

Delay action until user stops typing.

```go
type Model struct {
    searchInput   string
    searchTimer   *time.Timer
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.KeyMsg:
        m.searchInput += msg.String()

        // Cancel previous timer
        if m.searchTimer != nil {
            m.searchTimer.Stop()
        }

        // Start new timer
        return m, tea.Tick(time.Millisecond*300, func(t time.Time) tea.Msg {
            return SearchDelayedMsg{}
        })

    case SearchDelayedMsg:
        // User stopped typing - execute search
        return m, m.search(m.searchInput)
}
```

### Pattern 3: Optimistic Updates

Update UI immediately, revert on error.

```go
case "p":  // Publish
    // Optimistically update UI
    node := m.selectedNode
    previousStatus := node.Status
    node.Status = StatusPublished

    return m, m.publishContent(node.ID, previousStatus)

case ContentPublishedMsg:
    if msg.Error != nil {
        // Revert on error
        node := m.contentTree.NodeIndex[msg.ContentID]
        node.Status = msg.PreviousStatus
        m.errorMessage = "Publish failed"
        return m, nil
    }

    m.successMessage = "Published"
    return m, nil
```

### Pattern 4: Undo/Redo

Store previous states for undo.

```go
type Model struct {
    current  State
    history  []State
    future   []State
}

func (m Model) performAction(newState State) (Model, tea.Cmd) {
    // Save current state to history
    m.history = append(m.history, m.current)
    m.current = newState
    m.future = nil  // Clear redo stack
    return m, nil
}

func (m Model) undo() Model {
    if len(m.history) == 0 {
        return m
    }

    // Move current to future
    m.future = append(m.future, m.current)

    // Pop from history
    m.current = m.history[len(m.history)-1]
    m.history = m.history[:len(m.history)-1]

    return m
}
```

---

## Testing TUI Code

The Elm Architecture makes testing easier because functions are pure.

### Testing Update Function

```go
func TestCursorMovement(t *testing.T) {
    // Setup initial state
    m := Model{
        cursor: 0,
        items:  []Item{{}, {}, {}},  // 3 items
    }

    // Simulate down arrow
    newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})

    // Assert state changed
    if newModel.(Model).cursor != 1 {
        t.Errorf("Expected cursor 1, got %d", newModel.(Model).cursor)
    }

    // Simulate another down
    newModel2, _ := newModel.Update(tea.KeyMsg{Type: tea.KeyDown})

    if newModel2.(Model).cursor != 2 {
        t.Errorf("Expected cursor 2, got %d", newModel2.(Model).cursor)
    }
}
```

### Testing Commands

```go
func TestLoadContent(t *testing.T) {
    // Create mock database
    mockDB := &MockDB{
        content: []Content{{ID: 1}, {ID: 2}},
    }

    m := Model{db: mockDB}

    // Get command
    cmd := m.loadContent()

    // Execute command (get message)
    msg := cmd()

    // Assert message type and content
    loadedMsg, ok := msg.(ContentLoadedMsg)
    if !ok {
        t.Fatal("Expected ContentLoadedMsg")
    }

    if len(loadedMsg.Content) != 2 {
        t.Errorf("Expected 2 items, got %d", len(loadedMsg.Content))
    }
}
```

### Testing View Function

```go
func TestViewShowsLoading(t *testing.T) {
    m := Model{
        loading: true,
        ready:   true,
    }

    output := m.View()

    if !strings.Contains(output, "Loading") {
        t.Error("Expected 'Loading' in output")
    }
}

func TestViewShowsContent(t *testing.T) {
    m := Model{
        loading: false,
        ready:   true,
        items:   []Item{{Title: "Test Item"}},
    }

    output := m.View()

    if !strings.Contains(output, "Test Item") {
        t.Error("Expected 'Test Item' in output")
    }
}
```

---

## Performance Considerations

### 1. View Rendering Performance

**Problem:** View called on every update, even if state didn't change.

**Solution:** Cache rendered strings.

```go
type Model struct {
    contentCache     string
    contentCacheHash uint64
}

func (m Model) View() string {
    // Calculate hash of relevant state
    hash := m.calculateStateHash()

    if hash == m.contentCacheHash {
        return m.contentCache  // Return cached
    }

    // Re-render
    content := m.renderContent()
    m.contentCache = content
    m.contentCacheHash = hash

    return content
}
```

### 2. Large Lists

**Problem:** Rendering thousands of items is slow.

**Solution:** Virtualization - only render visible items.

```go
func (m Model) renderList() string {
    visibleStart := m.scrollOffset
    visibleEnd := m.scrollOffset + m.viewport.Height

    if visibleEnd > len(m.items) {
        visibleEnd = len(m.items)
    }

    // Only render visible items
    lines := []string{}
    for i := visibleStart; i < visibleEnd; i++ {
        lines = append(lines, m.renderItem(m.items[i]))
    }

    return strings.Join(lines, "\n")
}
```

### 3. Expensive State Calculations

**Problem:** Computing derived state on every render.

**Solution:** Memoization.

```go
type Model struct {
    items        []Item
    filteredCache []Item
    filterCacheKey string
}

func (m Model) getFilteredItems() []Item {
    cacheKey := m.filterString + strconv.Itoa(len(m.items))

    if cacheKey == m.filterCacheKey {
        return m.filteredCache
    }

    // Recompute
    filtered := []Item{}
    for _, item := range m.items {
        if strings.Contains(item.Title, m.filterString) {
            filtered = append(filtered, item)
        }
    }

    m.filteredCache = filtered
    m.filterCacheKey = cacheKey

    return filtered
}
```

### 4. Message Queue Flooding

**Problem:** Too many messages overwhelm the system.

**Solution:** Throttling.

```go
type Model struct {
    lastUpdateTime time.Time
    updateThrottle time.Duration
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Throttle rapid updates
    case HighFrequencyMsg:
        now := time.Now()
        if now.Sub(m.lastUpdateTime) < m.updateThrottle {
            return m, nil  // Ignore
        }
        m.lastUpdateTime = now
        // ... process message
}
```

---

## Common Pitfalls and Solutions

### Pitfall 1: Mutating Model in Place

**Bad:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.cursor++  // Mutates original
    return m, nil
}
```

**Why bad:** Creates confusion about state ownership.

**Good:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    newModel := m
    newModel.cursor++
    return newModel, nil
}
```

### Pitfall 2: Side Effects in Update

**Bad:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case "r":
        m.content = m.db.LoadContent()  // Side effect!
        return m, nil
}
```

**Why bad:** Blocks UI, not testable, can't handle errors.

**Good:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case "r":
        m.loading = true
        return m, m.loadContent()  // Returns command
}
```

### Pitfall 3: Forgetting to Return Model

**Bad:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.cursor++
    // Forgot to return!
}
```

**Compiler error:** Function must return values.

**Good:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.cursor++
    return m, nil  // Always return
}
```

### Pitfall 4: Not Handling All Cases

**Bad:**
```go
case ContentLoadedMsg:
    m.content = msg.Content
    return m, nil
    // Forgot to check msg.Error!
```

**Why bad:** Errors ignored, state inconsistent.

**Good:**
```go
case ContentLoadedMsg:
    if msg.Error != nil {
        m.errorMessage = msg.Error.Error()
        m.loading = false
        return m, nil
    }
    m.content = msg.Content
    m.loading = false
    return m, nil
```

### Pitfall 5: Race Conditions in Commands

**Bad:**
```go
func (m *Model) badCommand() tea.Cmd {
    return func() tea.Msg {
        m.data = fetchData()  // Race condition!
        return nil
    }
}
```

**Why bad:** Mutates model from goroutine.

**Good:**
```go
func (m *Model) goodCommand() tea.Cmd {
    return func() tea.Msg {
        data := fetchData()
        return DataLoadedMsg{Data: data}  // Send via message
    }
}
```

---

## Comparison to Other Patterns

### vs MVC (Model-View-Controller)

| Aspect | MVC | Elm Architecture |
|--------|-----|------------------|
| State | Scattered across models | Single Model struct |
| Updates | Controller mutates models | Pure function returns new state |
| View | Observes model changes | Pure function of state |
| Flow | Bidirectional | Unidirectional |
| Side effects | Anywhere | Isolated in Commands |

### vs Redux/Flux

Very similar! Elm Architecture inspired Redux.

| Aspect | Redux | Elm Architecture |
|--------|-------|------------------|
| State | Store | Model |
| Updates | Reducers | Update function |
| Side effects | Middleware | Commands |
| Actions | Action objects | Messages (any type) |

Main difference: Elm is more strongly typed.

### vs Traditional Event Handlers

| Aspect | Event Handlers | Elm Architecture |
|--------|---------------|------------------|
| State | Global or scattered | Centralized Model |
| Updates | Direct mutation | Return new state |
| Async | Callbacks | Commands → Messages |
| Testing | Hard (mocking) | Easy (pure functions) |
| Debugging | Hard to trace | Clear message log |

---

## Related Documentation

**Implementation:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` - Practical TUI implementation
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/CREATING_TUI_SCREENS.md` - Step-by-step screen creation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - Tree data structure
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Domain model

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` - Feature development
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/DEBUGGING.md` - Debugging strategies

**External Resources:**
- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [Elm Architecture Guide](https://guide.elm-lang.org/architecture/)
- [Redux Documentation](https://redux.js.org/) - Similar pattern

---

## Quick Reference

### Core Pattern

```go
type Model struct { /* state */ }

func (m Model) Init() tea.Cmd {
    return initialCommands()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle messages, return new state + command
    return m, cmd
}

func (m Model) View() string {
    // Pure function: state → string
    return renderedUI
}
```

### Message Flow

```
Input → Message → Update → (New State, Command) → View → Display
                                     ↓
                                 Execute Command
                                     ↓
                                 New Message
                                     ↓
                                 Back to Update
```

### Command Pattern

```go
func (m *Model) asyncOperation() tea.Cmd {
    return func() tea.Msg {
        result := doWork()
        return ResultMsg{Data: result}
    }
}
```

### Key Principles

1. **Immutable State** - Never mutate, always return new state
2. **Pure Functions** - Same input → same output
3. **Isolated Side Effects** - Only in Commands
4. **Unidirectional Flow** - Clear data flow
5. **Message-Driven** - All communication via messages

### Benefits

✓ Predictable state management
✓ Easy to test (pure functions)
✓ Clear debugging (message log)
✓ No race conditions (message queue)
✓ Time-travel debugging possible
✓ Easier to reason about

### Trade-offs

✗ Steeper learning curve
✗ More verbose than imperative
✗ Can't mutate state directly
✗ Async operations are indirect
✗ Performance overhead (copying state)
