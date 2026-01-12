# CREATING_TUI_SCREENS.md

Step-by-step guide for creating TUI (Terminal User Interface) screens in ModulaCMS using Charmbracelet Bubbletea.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/CREATING_TUI_SCREENS.md`
**Purpose:** Complete guide to building TUI screens from scratch
**Last Updated:** 2026-01-12

---

## Overview

ModulaCMS uses **Charmbracelet Bubbletea** for its TUI, implementing the **Elm Architecture** pattern. This guide walks through creating new TUI screens step-by-step, from concept to working implementation.

**What You'll Learn:**
1. Understanding the Elm Architecture in ModulaCMS
2. Creating new page types
3. Implementing Model, Init, Update, and View functions
4. Handling user input and navigation
5. Working with forms and dialogs
6. Loading data asynchronously
7. Styling with Lipgloss
8. Real examples from the codebase

---

## Prerequisites

**Before creating TUI screens, you should understand:**

- **Elm Architecture** - Model-Update-View cycle (see [TUI_ARCHITECTURE.md](../architecture/TUI_ARCHITECTURE.md))
- **Bubbletea basics** - Messages, commands, tea.Model interface
- **Go channels and concurrency** - For async operations
- **Lipgloss styling** - Terminal UI styling library

**External Resources:**
- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Huh Forms Documentation](https://github.com/charmbracelet/huh)
- [Elm Architecture Guide](https://guide.elm-lang.org/architecture/)

**Related Documentation:**
- [TUI_ARCHITECTURE.md](../architecture/TUI_ARCHITECTURE.md) - Deep dive into Elm Architecture
- [CLI_PACKAGE.md](../packages/CLI_PACKAGE.md) - Complete CLI package reference
- [ADDING_FEATURES.md](ADDING_FEATURES.md) - Full feature workflow (Step 5: TUI Interface)

---

## The Elm Architecture in ModulaCMS

### Core Concepts

ModulaCMS implements a **single Model** architecture where all application state lives in one `Model` struct. The TUI cycle works like this:

```
User Input → Message → Update → New Model → View → Render
                ↑                                      ↓
                └──────── Commands (async) ────────────┘
```

**Key principles:**
1. **Immutable State** - Model is never mutated, always return new Model
2. **Pure Functions** - Update functions have no side effects
3. **Messages** - All state changes happen through messages
4. **Commands** - Async operations return messages

### The Model Struct

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/model.go`

```go
type Model struct {
    Config       *config.Config
    Status       ApplicationState
    Width        int
    Height       int
    Page         Page              // Current page
    PageMenu     []Page            // Current page's menu
    PageMap      map[PageIndex]Page
    Cursor       int               // Current selection
    CursorMax    int               // Maximum cursor position
    Focus        FocusKey          // Which component has focus
    Loading      bool              // Loading indicator
    Spinner      spinner.Model     // Loading spinner
    Viewport     viewport.Model    // Scrollable content
    Paginator    paginator.Model   // Pagination
    Form         *huh.Form         // Current form
    FormFields   []huh.Field       // Form fields
    FormValues   []*string         // Form input values
    Dialog       *DialogModel      // Active dialog
    DialogActive bool              // Dialog visibility
    Headers      []string          // Table headers
    Rows         [][]string        // Table data
    Root         TreeRoot          // Content tree
    History      []PageHistory     // Navigation history
    Err          error             // Error state
    // ... more fields ...
}
```

### The Three Core Functions

#### 1. Init() - Initialization

```go
func (m Model) Init() tea.Cmd {
    if m.Loading {
        return m.Spinner.Tick
    }
    return nil
}
```

**Purpose:** Return initial commands when program starts.

#### 2. Update() - State Transitions

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Message handlers cascade through specialized update functions
    if m, cmd := m.UpdateLog(msg); cmd != nil {
        return m, cmd
    }
    if m, cmd := m.UpdateState(msg); cmd != nil {
        return m, cmd
    }
    if m, cmd := m.UpdateNavigation(msg); cmd != nil {
        return m, cmd
    }
    // ... more handlers ...

    return m.PageSpecificMsgHandlers(nil, msg)
}
```

**Purpose:** Handle messages and return new model + optional commands.

#### 3. View() - Rendering

```go
func (m Model) View() string {
    switch m.Page.Index {
    case HOMEPAGE:
        p := NewMenuPage()
        p.AddTitle(m.Titles[m.TitleFont])
        p.AddHeader("Home")
        p.AddMenu(menu)
        p.AddStatus(m.RenderStatusBar())
        return p.Render(m)
    case READPAGE:
        p := NewTablePage()
        p.AddTitle(m.Titles[m.TitleFont])
        p.AddHeaders(m.Headers)
        p.AddRows(m.Rows)
        return p.Render(m)
    // ... more pages ...
    }
}
```

**Purpose:** Render current model state to string.

---

## Step 1: Define Your Page Type

First, add a new page constant and initialize it.

### 1.1 Add Page Constant

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/pages.go`

```go
const (
    HOMEPAGE PageIndex = iota
    CMSPAGE
    DATABASEPAGE
    // ... existing pages ...
    COMMENTPAGE  // ← Add your new page here
)
```

### 1.2 Initialize Page in PageMap

In the same file, add to `InitPages()`:

```go
func InitPages() *map[PageIndex]Page {
    homePage := NewPage(HOMEPAGE, "Home")
    // ... existing pages ...
    commentPage := NewPage(COMMENTPAGE, "Comments")  // ← Add here

    p := make(map[PageIndex]Page, 0)
    p[HOMEPAGE] = homePage
    // ... existing mappings ...
    p[COMMENTPAGE] = commentPage  // ← Map here

    return &p
}
```

### 1.3 Create Page Constructor (Optional)

For dynamic pages with parameters:

```go
func NewCommentPage(label string) Page {
    return Page{
        Index: COMMENTPAGE,
        Label: label,
    }
}
```

---

## Step 2: Create Message Types

Define messages for your page's state changes and async operations.

### 2.1 Add Message Types

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/message_types.go`

```go
// Comment-specific messages
type CommentListFetchMsg struct{}

type CommentListFetchedMsg struct {
    Comments []db.Comment
}

type CommentCreateMsg struct {
    ContentDataID int64
    AuthorID      int64
    Text          string
}

type CommentCreatedMsg struct {
    Comment db.Comment
}

type CommentApproveMsg struct {
    CommentID int64
}

type CommentDeleteMsg struct {
    CommentID int64
}

type CommentErrorMsg struct {
    Error error
}
```

**Message naming conventions:**
- `{Action}Msg` - Trigger an action (e.g., `CommentListFetchMsg`)
- `{Action}edMsg` - Action completed (e.g., `CommentListFetchedMsg`)
- `{Entity}ErrorMsg` - Error occurred (e.g., `CommentErrorMsg`)

---

## Step 3: Create Command Functions

Commands perform async operations and return messages.

### 3.1 Create Command File

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/commands_comment.go`

```go
package cli

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/hegner123/modulacms/internal/config"
    "github.com/hegner123/modulacms/internal/db"
)

// FetchCommentListCmd loads all comments from database
func FetchCommentListCmd(c *config.Config) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)
        comments, err := d.ListComments()
        if err != nil {
            return CommentErrorMsg{Error: err}
        }
        return CommentListFetchedMsg{Comments: comments}
    }
}

// FetchCommentsByContentCmd loads comments for specific content
func FetchCommentsByContentCmd(c *config.Config, contentID int64) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)
        comments, err := d.ListCommentsByContent(contentID)
        if err != nil {
            return CommentErrorMsg{Error: err}
        }
        return CommentListFetchedMsg{Comments: comments}
    }
}

// CreateCommentCmd creates a new comment
func CreateCommentCmd(c *config.Config, params db.CreateCommentParams) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)
        comment := d.CreateComment(params)
        if comment.CommentID == 0 {
            return CommentErrorMsg{Error: fmt.Errorf("failed to create comment")}
        }
        return CommentCreatedMsg{Comment: comment}
    }
}

// ApproveCommentCmd approves a comment
func ApproveCommentCmd(c *config.Config, commentID int64) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)
        err := d.ApproveComment(commentID)
        if err != nil {
            return CommentErrorMsg{Error: err}
        }
        return CommentListFetchMsg{} // Refresh list
    }
}

// DeleteCommentCmd deletes a comment
func DeleteCommentCmd(c *config.Config, commentID int64) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)
        err := d.DeleteComment(commentID)
        if err != nil {
            return CommentErrorMsg{Error: err}
        }
        return CommentListFetchMsg{} // Refresh list
    }
}
```

**Command patterns:**
- Return `tea.Cmd` (which is `func() tea.Msg`)
- Perform database/network operations inside the returned function
- Always return a message (success or error)
- Use config to get database connection

---

## Step 4: Implement Update Handler

Handle messages specific to your page.

### 4.1 Create Update File

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/update_comment.go`

```go
package cli

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/hegner123/modulacms/internal/db"
)

func (m Model) UpdateComment(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {

    case CommentListFetchMsg:
        m.Loading = true
        return m, tea.Batch(
            m.Spinner.Tick,
            FetchCommentListCmd(m.Config),
        )

    case CommentListFetchedMsg:
        m.Loading = false
        // Convert comments to table rows
        m.Headers = []string{"ID", "Content", "Author", "Text", "Status", "Created"}
        m.Rows = make([][]string, len(msg.Comments))
        for i, comment := range msg.Comments {
            m.Rows[i] = []string{
                fmt.Sprintf("%d", comment.CommentID),
                fmt.Sprintf("%d", comment.ContentDataID),
                fmt.Sprintf("%d", comment.AuthorID),
                comment.CommentText,
                comment.Status,
                comment.DateCreated,
            }
        }
        m.CursorMax = len(m.Rows) - 1
        m.Cursor = 0
        return m, nil

    case CommentApproveMsg:
        m.Loading = true
        return m, tea.Batch(
            m.Spinner.Tick,
            ApproveCommentCmd(m.Config, msg.CommentID),
        )

    case CommentDeleteMsg:
        // Show confirmation dialog
        dialog := NewDialog(
            "Delete Comment",
            "Are you sure you want to delete this comment?",
            true, // show cancel
            DIALOGDELETE,
        )
        m.Dialog = &dialog
        m.DialogActive = true
        return m, nil

    case CommentCreatedMsg:
        m.Loading = false
        // Refresh comment list
        return m, FetchCommentListCmd(m.Config)

    case CommentErrorMsg:
        m.Loading = false
        m.Err = msg.Error
        m.Status = ERROR
        return m, nil
    }

    return m, nil
}
```

### 4.2 Hook Into Main Update

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/update.go`

Add your update handler to the cascade:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ... existing handlers ...
    if m, cmd := m.UpdateComment(msg); cmd != nil {
        return m, cmd
    }
    // ... rest of handlers ...
}
```

---

## Step 5: Implement View Rendering

Create the visual representation of your page.

### 5.1 Add View Case

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/view.go`

```go
func (m Model) View() string {
    // ... existing cases ...

    case COMMENTPAGE:
        if m.Loading {
            return fmt.Sprintf("\n\n   %s Loading comments...\n\n", m.Spinner.View())
        }

        // Build menu for actions
        menu := []string{
            "View All Comments",
            "Approve Selected",
            "Delete Selected",
            "Back",
        }

        p := NewTablePage()
        p.AddTitle(m.Titles[m.TitleFont])
        p.AddHeader("Comments Management")
        p.AddHeaders(m.Headers)
        p.AddRows(m.Rows)
        p.AddMenu(menu)
        p.AddStatus(m.RenderStatusBar())

        ui := p.Render(m)

        // Overlay dialog if active
        if m.DialogActive && m.Dialog != nil {
            ui = DialogOverlay(ui, *m.Dialog, m.Width, m.Height)
        }

        return ui

    // ... rest of cases ...
}
```

### 5.2 Using Page Builders

ModulaCMS provides page builder helpers:

**MenuPage** - List of menu items:
```go
p := NewMenuPage()
p.AddTitle(title)
p.AddHeader("Page Title")
p.AddMenu([]string{"Option 1", "Option 2", "Option 3"})
p.AddStatus(statusBar)
return p.Render(m)
```

**TablePage** - Display data in table format:
```go
p := NewTablePage()
p.AddTitle(title)
p.AddHeader("Table View")
p.AddHeaders([]string{"ID", "Name", "Status"})
p.AddRows([][]string{
    {"1", "Item 1", "Active"},
    {"2", "Item 2", "Pending"},
})
p.AddStatus(statusBar)
return p.Render(m)
```

**FormPage** - Display forms:
```go
p := NewFormPage()
p.AddTitle(title)
p.AddHeader("Create Item")
p.AddStatus(statusBar)
return p.Render(m)
```

**StaticPage** - Static content:
```go
p := NewStaticPage()
p.AddTitle(title)
p.AddHeader("Details")
p.AddBody(content)
p.AddStatus(statusBar)
return p.Render(m)
```

**CMSPage** - Content management specific:
```go
p := NewCMSPage()
p.AddTitle(title)
p.AddHeader("Content")
p.AddControls("q quit | ↑↓ navigate | enter select")
p.AddStatus(statusBar)
return p.Render(m)
```

---

## Step 6: Handle Keyboard Input

Implement keyboard controls for your page.

### 6.1 Page-Specific Key Handlers

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/update_controls.go` (or create `update_comment_keys.go`)

```go
func (m Model) HandleCommentPageKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
    // Only handle keys when on comment page
    if m.Page.Index != COMMENTPAGE {
        return m, nil
    }

    switch msg.String() {
    case "a": // Approve comment
        if len(m.Rows) > 0 {
            commentID := m.GetCurrentRowID() // Helper to get ID from current row
            return m, func() tea.Msg {
                return CommentApproveMsg{CommentID: commentID}
            }
        }

    case "d": // Delete comment
        if len(m.Rows) > 0 {
            commentID := m.GetCurrentRowID()
            return m, func() tea.Msg {
                return CommentDeleteMsg{CommentID: commentID}
            }
        }

    case "r": // Refresh
        return m, func() tea.Msg {
            return CommentListFetchMsg{}
        }

    case "enter": // View details
        if len(m.Rows) > 0 {
            // Navigate to detail page
            return m, func() tea.Msg {
                return NavigateToPage{
                    Page: NewPage(COMMENTSINGLEPAGE, "Comment Details"),
                }
            }
        }

    case "up", "k":
        if m.Cursor > 0 {
            m.Cursor--
        }
        return m, nil

    case "down", "j":
        if m.Cursor < m.CursorMax {
            m.Cursor++
        }
        return m, nil

    case "q", "esc": // Go back
        return m, func() tea.Msg {
            return NavigateToPage{
                Page: m.PageMap[HOMEPAGE],
            }
        }
    }

    return m, nil
}
```

### 6.2 Hook Into Control Updates

Add to the main keyboard handler cascade:

```go
func (m Model) UpdateControls(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Try comment page keys
        if newM, cmd := m.HandleCommentPageKeys(msg); cmd != nil {
            return newM, cmd
        }

        // ... other key handlers ...
    }
    return m, nil
}
```

---

## Step 7: Add Navigation

Enable navigation to/from your page.

### 7.1 Add to Menu

Add your page to the appropriate menu:

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/menus.go`

```go
func (m Model) HomepageMenuInit() []Page {
    return []Page{
        m.PageMap[CMSPAGE],
        m.PageMap[DATABASEPAGE],
        m.PageMap[COMMENTPAGE],  // ← Add here
        m.PageMap[CONFIGPAGE],
    }
}

// Or create a submenu
func (m Model) CmsMenuInit() []Page {
    return []Page{
        m.PageMap[CONTENT],
        m.PageMap[MEDIA],
        m.PageMap[COMMENTPAGE],  // ← Or add to CMS submenu
        m.PageMap[DATATYPES],
    }
}
```

### 7.2 Navigation Message

Navigate using the `NavigateToPage` message:

```go
// From anywhere in the code
return m, func() tea.Msg {
    return NavigateToPage{
        Page: m.PageMap[COMMENTPAGE],
        Menu: nil, // Optional: set page menu
    }
}
```

### 7.3 Handle Navigation Message

Navigation is handled in `UpdateNavigation`:

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/update_navigation.go`

```go
func (m Model) UpdateNavigation(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case NavigateToPage:
        // Save current page to history
        m.History = append(m.History, PageHistory{
            Page:   m.Page,
            Cursor: m.Cursor,
        })

        // Set new page
        m.Page = msg.Page
        m.Cursor = 0

        // Set menu if provided
        if msg.Menu != nil {
            m.PageMenu = *msg.Menu
        }

        // Page-specific initialization
        if msg.Page.Index == COMMENTPAGE {
            return m, FetchCommentListCmd(m.Config)
        }

        return m, nil
    }
    return m, nil
}
```

---

## Step 8: Working with Forms

Use Charmbracelet Huh for form creation.

### 8.1 Create Form Message

```go
type CommentFormMsg struct {
    ContentDataID int64
}

type CommentFormSubmitMsg struct {
    Text   string
    Status string
}
```

### 8.2 Build Form

```go
func BuildCommentForm(contentDataID int64) *huh.Form {
    var text, status string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Comment Text").
                Value(&text).
                Placeholder("Enter your comment...").
                Validate(func(s string) error {
                    if len(s) < 10 {
                        return fmt.Errorf("comment must be at least 10 characters")
                    }
                    return nil
                }),

            huh.NewSelect[string]().
                Title("Status").
                Options(
                    huh.NewOption("Pending", "pending"),
                    huh.NewOption("Approved", "approved"),
                    huh.NewOption("Rejected", "rejected"),
                ).
                Value(&status),
        ),
    )

    return form
}
```

### 8.3 Handle Form in Update

```go
func (m Model) UpdateComment(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {

    case CommentFormMsg:
        m.Form = BuildCommentForm(msg.ContentDataID)
        m.Focus = FORMFOCUS
        return m, m.Form.Init()

    case tea.KeyMsg:
        if m.Focus == FORMFOCUS && m.Form != nil {
            form, cmd := m.Form.Update(msg)
            m.Form = form.(*huh.Form)

            if m.Form.State == huh.StateCompleted {
                // Extract values and create comment
                // ... form submission logic ...
                return m, CreateCommentCmd(m.Config, params)
            }

            return m, cmd
        }
    }
    return m, nil
}
```

### 8.4 Render Form

```go
func (m Model) View() string {
    case COMMENTFORMPAGE:
        var content string
        if m.Form != nil {
            content = m.Form.View()
        }

        p := NewFormPage()
        p.AddTitle(m.Titles[m.TitleFont])
        p.AddHeader("Create Comment")
        p.AddBody(content)
        p.AddStatus(m.RenderStatusBar())
        return p.Render(m)
}
```

---

## Step 9: Working with Dialogs

Use dialogs for confirmations and alerts.

### 9.1 Show Dialog

```go
case CommentDeleteMsg:
    dialog := NewDialog(
        "Delete Comment",
        fmt.Sprintf("Delete comment #%d? This cannot be undone.", msg.CommentID),
        true, // show cancel button
        DIALOGDELETE,
    )
    m.Dialog = &dialog
    m.DialogActive = true
    return m, nil
```

### 9.2 Handle Dialog Response

```go
case DialogAcceptMsg:
    m.DialogActive = false

    switch msg.Action {
    case DIALOGDELETE:
        commentID := m.GetCurrentRowID()
        return m, DeleteCommentCmd(m.Config, commentID)
    }
    return m, nil

case DialogCancelMsg:
    m.DialogActive = false
    m.Dialog = nil
    return m, nil
```

### 9.3 Render Dialog Overlay

In your View:

```go
ui := p.Render(m)

// Overlay dialog if active
if m.DialogActive && m.Dialog != nil {
    ui = DialogOverlay(ui, *m.Dialog, m.Width, m.Height)
}

return ui
```

---

## Step 10: Styling with Lipgloss

Apply consistent styling to your UI elements.

### 10.1 Using Existing Styles

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/style.go`

ModulaCMS defines styles in the config:

```go
// Access theme colors
config.DefaultStyle.Primary    // Main text color
config.DefaultStyle.Secondary  // Secondary text color
config.DefaultStyle.Accent     // Highlight color
config.DefaultStyle.AccentBG   // Highlight background
config.DefaultStyle.Warn       // Warning color
```

### 10.2 Create Custom Styles

```go
var (
    commentTitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(config.DefaultStyle.Accent).
        Padding(0, 1)

    commentTextStyle = lipgloss.NewStyle().
        Foreground(config.DefaultStyle.Primary).
        PaddingLeft(2)

    commentStatusStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(config.DefaultStyle.Secondary).
        Border(lipgloss.RoundedBorder()).
        Padding(0, 1)

    approvedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00FF00")).
        Background(lipgloss.Color("#003300"))

    pendingStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FFFF00")).
        Background(lipgloss.Color("#333300"))
)
```

### 10.3 Apply Styles

```go
func FormatCommentStatus(status string) string {
    switch status {
    case "approved":
        return approvedStyle.Render("✓ APPROVED")
    case "pending":
        return pendingStyle.Render("⏳ PENDING")
    case "rejected":
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color("#FF0000")).
            Render("✗ REJECTED")
    default:
        return status
    }
}

// Use in view
for i, comment := range msg.Comments {
    m.Rows[i] = []string{
        fmt.Sprintf("%d", comment.CommentID),
        comment.CommentText,
        FormatCommentStatus(comment.Status),
        comment.DateCreated,
    }
}
```

### 10.4 Layout with Lipgloss

```go
// Join horizontally
row := lipgloss.JoinHorizontal(
    lipgloss.Left,
    leftColumn,
    lipgloss.NewStyle().Padding(0, 2).Render("│"),
    rightColumn,
)

// Join vertically
page := lipgloss.JoinVertical(
    lipgloss.Left,
    titleSection,
    headerSection,
    bodySection,
    statusBar,
)

// Center content
centered := lipgloss.Place(
    width,
    height,
    lipgloss.Center,
    lipgloss.Center,
    content,
)
```

---

## Complete Example: Comment Management Screen

Here's a complete working example combining all steps.

### File Structure

```
internal/cli/
├── pages.go              # Add COMMENTPAGE constant
├── message_types.go      # Add comment messages
├── commands_comment.go   # Comment commands
├── update_comment.go     # Comment update handler
├── view.go               # Add comment view case
└── menus.go              # Add to menu
```

### Full Implementation

**pages.go:**
```go
const (
    // ... existing pages ...
    COMMENTPAGE PageIndex = iota
    COMMENTSINGLEPAGE
)

func InitPages() *map[PageIndex]Page {
    // ... existing pages ...
    commentPage := NewPage(COMMENTPAGE, "Comments")
    commentSinglePage := NewPage(COMMENTSINGLEPAGE, "Comment Detail")

    p := make(map[PageIndex]Page, 0)
    // ... existing mappings ...
    p[COMMENTPAGE] = commentPage
    p[COMMENTSINGLEPAGE] = commentSinglePage

    return &p
}
```

**message_types.go:**
```go
type CommentListFetchMsg struct {
    ContentDataID int64 // 0 for all comments
}

type CommentListFetchedMsg struct {
    Comments []db.Comment
}

type CommentApproveMsg struct {
    CommentID int64
}

type CommentDeleteMsg struct {
    CommentID int64
}

type CommentCreateFormMsg struct {
    ContentDataID int64
}

type CommentFormSubmitMsg struct {
    Params db.CreateCommentParams
}

type CommentErrorMsg struct {
    Error error
}
```

**commands_comment.go:**
```go
package cli

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/hegner123/modulacms/internal/config"
    "github.com/hegner123/modulacms/internal/db"
)

func FetchCommentsCmd(c *config.Config, contentID int64) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)

        var comments []db.Comment
        var err error

        if contentID == 0 {
            comments, err = d.ListComments()
        } else {
            comments, err = d.ListCommentsByContent(contentID)
        }

        if err != nil {
            return CommentErrorMsg{Error: err}
        }

        return CommentListFetchedMsg{Comments: comments}
    }
}

func ApproveCommentCmd(c *config.Config, commentID int64) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)
        err := d.ApproveComment(commentID)
        if err != nil {
            return CommentErrorMsg{Error: err}
        }
        // Trigger refresh
        return CommentListFetchMsg{ContentDataID: 0}
    }
}

func DeleteCommentCmd(c *config.Config, commentID int64) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)
        err := d.DeleteComment(commentID)
        if err != nil {
            return CommentErrorMsg{Error: err}
        }
        return CommentListFetchMsg{ContentDataID: 0}
    }
}

func CreateCommentCmd(c *config.Config, params db.CreateCommentParams) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)
        comment := d.CreateComment(params)
        if comment.CommentID == 0 {
            return CommentErrorMsg{Error: fmt.Errorf("failed to create comment")}
        }
        return CommentListFetchMsg{ContentDataID: params.ContentDataID}
    }
}
```

**update_comment.go:**
```go
package cli

import (
    "fmt"
    "strconv"
    tea "github.com/charmbracelet/bubbletea"
)

func (m Model) UpdateComment(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {

    case CommentListFetchMsg:
        m.Loading = true
        return m, tea.Batch(
            m.Spinner.Tick,
            FetchCommentsCmd(m.Config, msg.ContentDataID),
        )

    case CommentListFetchedMsg:
        m.Loading = false
        m.Headers = []string{"ID", "Content ID", "Text", "Status", "Created"}
        m.Rows = make([][]string, len(msg.Comments))

        for i, comment := range msg.Comments {
            m.Rows[i] = []string{
                fmt.Sprintf("%d", comment.CommentID),
                fmt.Sprintf("%d", comment.ContentDataID),
                truncateText(comment.CommentText, 50),
                FormatCommentStatus(comment.Status),
                comment.DateCreated,
            }
        }

        m.CursorMax = len(m.Rows) - 1
        if m.Cursor > m.CursorMax {
            m.Cursor = 0
        }
        return m, nil

    case CommentApproveMsg:
        m.Loading = true
        return m, tea.Batch(
            m.Spinner.Tick,
            ApproveCommentCmd(m.Config, msg.CommentID),
        )

    case CommentDeleteMsg:
        dialog := NewDialog(
            "Delete Comment",
            fmt.Sprintf("Delete comment #%d? This cannot be undone.", msg.CommentID),
            true,
            DIALOGDELETE,
        )
        m.Dialog = &dialog
        m.DialogActive = true
        return m, nil

    case CommentErrorMsg:
        m.Loading = false
        m.Err = msg.Error
        m.Status = ERROR
        return m, nil

    case tea.KeyMsg:
        // Handle keys when on comment page
        if m.Page.Index == COMMENTPAGE {
            return m.HandleCommentKeys(msg)
        }
    }

    return m, nil
}

func (m Model) HandleCommentKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
    switch msg.String() {
    case "a":
        if len(m.Rows) > 0 {
            commentID, _ := strconv.ParseInt(m.Rows[m.Cursor][0], 10, 64)
            return m, func() tea.Msg {
                return CommentApproveMsg{CommentID: commentID}
            }
        }

    case "d":
        if len(m.Rows) > 0 {
            commentID, _ := strconv.ParseInt(m.Rows[m.Cursor][0], 10, 64)
            return m, func() tea.Msg {
                return CommentDeleteMsg{CommentID: commentID}
            }
        }

    case "r":
        return m, func() tea.Msg {
            return CommentListFetchMsg{ContentDataID: 0}
        }

    case "up", "k":
        if m.Cursor > 0 {
            m.Cursor--
        }

    case "down", "j":
        if m.Cursor < m.CursorMax {
            m.Cursor++
        }

    case "q", "esc":
        return m, func() tea.Msg {
            return NavigateToPage{Page: m.PageMap[HOMEPAGE]}
        }
    }

    return m, nil
}

func truncateText(text string, maxLen int) string {
    if len(text) <= maxLen {
        return text
    }
    return text[:maxLen-3] + "..."
}

func FormatCommentStatus(status string) string {
    switch status {
    case "approved":
        return "✓ Approved"
    case "pending":
        return "⏳ Pending"
    case "rejected":
        return "✗ Rejected"
    default:
        return status
    }
}
```

**view.go:**
```go
func (m Model) View() string {
    // ... existing cases ...

    case COMMENTPAGE:
        if m.Loading {
            return fmt.Sprintf("\n\n   %s Loading comments...\n\n", m.Spinner.View())
        }

        p := NewTablePage()
        p.AddTitle(m.Titles[m.TitleFont])
        p.AddHeader("Comment Management")
        p.AddHeaders(m.Headers)
        p.AddRows(m.Rows)

        controls := "a:approve | d:delete | r:refresh | q:back"
        p.AddControls(controls)
        p.AddStatus(m.RenderStatusBar())

        ui := p.Render(m)

        if m.DialogActive && m.Dialog != nil {
            ui = DialogOverlay(ui, *m.Dialog, m.Width, m.Height)
        }

        return ui

    // ... rest of cases ...
}
```

**update.go:**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ... existing handlers ...

    if m, cmd := m.UpdateComment(msg); cmd != nil {
        return m, cmd
    }

    // ... rest of handlers ...
}
```

**menus.go:**
```go
func (m Model) CmsMenuInit() []Page {
    return []Page{
        m.PageMap[CONTENT],
        m.PageMap[MEDIA],
        m.PageMap[COMMENTPAGE],  // ← Add here
        m.PageMap[DATATYPES],
    }
}
```

---

## Common Patterns

### Pattern 1: List → Detail → Edit

```go
// List page
case COMMENTPAGE:
    // Show table of comments
    // Enter key → navigate to detail

// Detail page
case COMMENTSINGLEPAGE:
    // Show single comment details
    // 'e' key → navigate to edit form

// Edit page
case COMMENTEDITPAGE:
    // Show edit form
    // Submit → save and return to list
```

### Pattern 2: Async Loading with Spinner

```go
// Trigger load
case SomeActionMsg:
    m.Loading = true
    return m, tea.Batch(
        m.Spinner.Tick,
        FetchDataCmd(m.Config),
    )

// Handle result
case DataFetchedMsg:
    m.Loading = false
    m.Data = msg.Data
    return m, nil

// View
if m.Loading {
    return fmt.Sprintf("%s Loading...", m.Spinner.View())
}
```

### Pattern 3: Confirmation Dialog

```go
// Show dialog
case DeleteActionMsg:
    dialog := NewDialog("Confirm", "Are you sure?", true, DIALOGDELETE)
    m.Dialog = &dialog
    m.DialogActive = true
    return m, nil

// Handle response
case DialogAcceptMsg:
    m.DialogActive = false
    if msg.Action == DIALOGDELETE {
        return m, DeleteCmd(m.Config, itemID)
    }
```

### Pattern 4: Form with Validation

```go
form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Title("Email").
            Value(&email).
            Validate(func(s string) error {
                if !strings.Contains(s, "@") {
                    return fmt.Errorf("invalid email")
                }
                return nil
            }),
    ),
)
```

### Pattern 5: Pagination

```go
// Update paginator
m.Paginator.SetTotalPages(totalPages)
m.Paginator.PerPage = 10

// Handle page changes
case tea.KeyMsg:
    switch msg.String() {
    case "pgup":
        m.Paginator.PrevPage()
    case "pgdown":
        m.Paginator.NextPage()
    }

// Render
paginatorView := m.Paginator.View()
```

---

## Debugging TUI Issues

### Debug Logging

Use the logger to debug without disrupting the TUI:

```go
import "github.com/hegner123/modulacms/internal/utility"

// In your update function
utility.DefaultLogger.Info("Comment page loaded",
    "cursor", m.Cursor,
    "rows", len(m.Rows),
)

// Check debug.log file
tail -f debug.log
```

### Common Issues

**Issue 1: Screen not updating**
- **Cause:** Forgot to return command or message not handled
- **Fix:** Ensure Update returns `tea.Cmd` when state changes

**Issue 2: Cursor out of bounds**
- **Cause:** CursorMax not updated when Rows change
- **Fix:** Always set `m.CursorMax = len(m.Rows) - 1` after updating Rows

**Issue 3: Keys not working**
- **Cause:** Key handler not in cascade or wrong Focus
- **Fix:** Check Focus state and handler order in Update

**Issue 4: Dialog not showing**
- **Cause:** DialogActive not set or overlay not called
- **Fix:** Set `m.DialogActive = true` and call `DialogOverlay` in View

**Issue 5: Form not submitting**
- **Cause:** Form state not checked or Init not called
- **Fix:** Call `m.Form.Init()` and check `m.Form.State == huh.StateCompleted`

---

## Testing TUI Components

### Unit Testing Update Functions

```go
func TestCommentUpdate(t *testing.T) {
    m := Model{
        Page:      NewPage(COMMENTPAGE, "Comments"),
        Cursor:    0,
        CursorMax: 0,
    }

    msg := CommentListFetchedMsg{
        Comments: []db.Comment{
            {CommentID: 1, CommentText: "Test", Status: "pending"},
        },
    }

    newM, _ := m.UpdateComment(msg)

    if len(newM.Rows) != 1 {
        t.Errorf("Expected 1 row, got %d", len(newM.Rows))
    }

    if newM.Loading {
        t.Error("Loading should be false after fetch")
    }
}
```

### Integration Testing

```go
func TestCommentPageFlow(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    config := &config.Config{/* ... */}
    m, _ := InitialModel(nil, config)

    // Navigate to comment page
    m, _ = m.Update(NavigateToPage{Page: m.PageMap[COMMENTPAGE]})

    // Should trigger fetch
    // ... test flow ...
}
```

---

## Performance Considerations

### 1. Lazy Loading

Load data only when needed:

```go
case NavigateToPage:
    m.Page = msg.Page

    // Only fetch data when entering page
    if msg.Page.Index == COMMENTPAGE && len(m.Rows) == 0 {
        return m, FetchCommentsCmd(m.Config)
    }
```

### 2. Debouncing

Prevent rapid-fire commands:

```go
var lastFetch time.Time

func (m Model) HandleRefresh() (Model, tea.Cmd) {
    if time.Since(lastFetch) < 500*time.Millisecond {
        return m, nil // Skip if too soon
    }
    lastFetch = time.Now()
    return m, FetchCommentsCmd(m.Config)
}
```

### 3. Virtual Scrolling

For large lists, render only visible rows:

```go
visibleStart := m.Cursor - 10
if visibleStart < 0 {
    visibleStart = 0
}
visibleEnd := m.Cursor + 10
if visibleEnd > len(m.Rows) {
    visibleEnd = len(m.Rows)
}

visibleRows := m.Rows[visibleStart:visibleEnd]
```

---

## Checklist for New TUI Screen

Use this checklist when creating a new screen:

- [ ] Define page constant in `pages.go`
- [ ] Initialize page in `InitPages()`
- [ ] Create message types in `message_types.go`
- [ ] Implement command functions (async operations)
- [ ] Create update handler file
- [ ] Hook update handler into main Update cascade
- [ ] Add view case to View() function
- [ ] Implement keyboard controls
- [ ] Add to appropriate menu
- [ ] Test navigation to/from page
- [ ] Add loading states
- [ ] Implement error handling
- [ ] Add dialogs for confirmations
- [ ] Style with Lipgloss
- [ ] Add debug logging
- [ ] Test with real data
- [ ] Document keyboard shortcuts
- [ ] Add to page builder if reusable pattern

---

## Related Documentation

**Essential Reading:**
- [TUI_ARCHITECTURE.md](../architecture/TUI_ARCHITECTURE.md) - Deep dive into Elm Architecture
- [CLI_PACKAGE.md](../packages/CLI_PACKAGE.md) - Complete CLI package reference
- [ADDING_FEATURES.md](ADDING_FEATURES.md) - Full feature workflow

**Related Workflows:**
- [ADDING_TABLES.md](ADDING_TABLES.md) - Adding database tables for your TUI
- [TESTING.md](TESTING.md) - Testing strategies (once available)
- [DEBUGGING.md](DEBUGGING.md) - Debugging guide (once available)

**External Resources:**
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)
- [Huh Forms Guide](https://github.com/charmbracelet/huh#usage)

---

## Quick Reference

### Essential Types

```go
type Model struct { /* application state */ }
func (m Model) Init() tea.Cmd { /* initialization */ }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* state transitions */ }
func (m Model) View() string { /* rendering */ }
```

### Message Pattern

```go
// Action trigger
type ActionMsg struct { ID int64 }

// Action result
type ActionCompletedMsg struct { Data SomeData }

// Action error
type ActionErrorMsg struct { Error error }
```

### Command Pattern

```go
func SomeActionCmd(config *config.Config, params Params) tea.Cmd {
    return func() tea.Msg {
        // Perform async operation
        result, err := doSomething(params)
        if err != nil {
            return ActionErrorMsg{Error: err}
        }
        return ActionCompletedMsg{Data: result}
    }
}
```

### Navigation

```go
return m, func() tea.Msg {
    return NavigateToPage{
        Page: m.PageMap[SOMEPAGE],
        Menu: &[]Page{...}, // optional
    }
}
```

### Loading Pattern

```go
// Start
m.Loading = true
return m, tea.Batch(m.Spinner.Tick, FetchCmd())

// End
m.Loading = false
return m, nil

// View
if m.Loading {
    return fmt.Sprintf("%s Loading...", m.Spinner.View())
}
```

---

**Last Updated:** 2026-01-12
**Status:** Complete
**Part of:** Phase 2 High Priority Documentation
