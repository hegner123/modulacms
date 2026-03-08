# TUI Message Flow

How messages travel through the Bubbletea update cycle in ModulaCMS. Each example traces one user action from keypress to screen update, showing only the immediate next call at each step.

## How to Read These Traces

Each step shows:
1. **Where** — the file and function executing
2. **What happens** — the logic at this step
3. **Next** — what gets returned or emitted (a `tea.Cmd`, a `tea.Msg`, or a state mutation)

The Bubbletea runtime drives the cycle: `Update()` returns a `tea.Cmd` → runtime executes it → the resulting `tea.Msg` arrives back at `Update()`.

---

## Example 1: Esc → Quit Confirmation → Confirm → Exit

User presses `esc` on any screen. A "Are you sure you want to quit?" dialog appears. User presses `enter` on "Quit" to confirm. The program exits.

### Step 1: Keypress arrives at `Model.Update()`

**File:** `update.go:14`

The Bubbletea runtime calls `Model.Update(msg)` with `tea.KeyMsg{String: "esc"}`.

Before any dispatch, `Update()` checks for `ctrl+c` and `esc` at the top:

```go
case "esc":
    if d, ok := m.ActiveOverlay.(*DialogModel); ok && d != nil && d.Action == DIALOGQUITCONFIRM {
        return m, tea.Quit
    }
    m.ActiveOverlay = nil
    return m.UpdateDialog(ShowQuitConfirmDialogMsg{})
```

No quit dialog is showing yet, so `ActiveOverlay` is nil (or not a quit confirm). The guard fails.

**Next:** Calls `m.UpdateDialog(ShowQuitConfirmDialogMsg{})`.

---

### Step 2: `UpdateDialog` creates the dialog

**File:** `update_dialog.go:46`

```go
case ShowQuitConfirmDialogMsg:
    dialog := NewDialog("Quit", "Are you sure you want to quit?", true, DIALOGQUITCONFIRM)
    dialog.SetButtons("Quit", "Cancel")
    return m, tea.Batch(
        OverlaySetCmd(&dialog),
        FocusSetCmd(DIALOGFOCUS),
    )
```

Creates a `DialogModel` with two buttons ("Quit" at index 0, "Cancel" at index 1). Returns two commands batched together.

**Next:** Bubbletea runtime executes both commands. They produce two messages: `OverlaySetMsg` and `SetPanelFocusMsg`.

---

### Step 3a: `OverlaySetMsg` sets the overlay

**File:** `update.go:47`

```go
case OverlaySetMsg:
    m.ActiveOverlay = typedMsg.Overlay
    return m, nil
```

Stores the `DialogModel` pointer on `m.ActiveOverlay`. The dialog is now visible in `View()`.

**Next:** Returns nil — no further command.

---

### Step 3b: `SetPanelFocusMsg` sets focus to dialog

**File:** `update.go:53`

```go
case SetPanelFocusMsg:
    m.PanelFocus = typedMsg.Panel
    return m, nil
```

Sets `m.PanelFocus = DIALOGFOCUS`. This tells the `View()` renderer to style the dialog as focused and tells `Update()` to route key input to the overlay.

**Next:** Returns nil — no further command. The dialog is now rendered and receiving input.

---

### Step 4: User presses `enter` on "Quit" button

**File:** `update.go:491`

A new `tea.KeyMsg{String: "enter"}` arrives at `Model.Update()`. The `esc`/`ctrl+c` guard at the top doesn't match `enter`, so execution continues to the overlay intercept:

```go
if m.ActiveOverlay != nil {
    if keyMsg, ok := msg.(tea.KeyMsg); ok {
        overlay, cmd := m.ActiveOverlay.OverlayUpdate(keyMsg)
        m.ActiveOverlay = overlay
        return m, cmd
    }
}
```

`ActiveOverlay` is set (the quit dialog), so the key goes to `OverlayUpdate`.

**Next:** Calls `m.ActiveOverlay.OverlayUpdate(keyMsg)`.

---

### Step 5: `DialogModel.OverlayUpdate` delegates to `Update`

**File:** `dialog.go:146`

```go
func (d *DialogModel) OverlayUpdate(msg tea.KeyMsg) (ModalOverlay, tea.Cmd) {
    updated, cmd := d.Update(msg)
    return &updated, cmd
}
```

Thin wrapper. Delegates to `DialogModel.Update()`.

**Next:** Calls `d.Update(msg)`.

---

### Step 6: `DialogModel.Update` routes to `ToggleControls`

**File:** `dialog.go:117`

`DIALOGQUITCONFIRM` matches the list of toggle-control dialog actions:

```go
case DIALOGDELETE, DIALOGACTIONCONFIRM, DIALOGINITCONTENT, DIALOGQUITCONFIRM, ...:
    return d.ToggleControls(msg)
```

**Next:** Calls `d.ToggleControls(msg)`.

---

### Step 7: `ToggleControls` handles `enter`

**File:** `dialog.go:157`

```go
case "enter":
    if d.focusIndex == 0 {
        return *d, func() tea.Msg { return DialogAcceptMsg{Action: d.Action} }
    }
    return *d, func() tea.Msg { return DialogCancelMsg{} }
```

`focusIndex` is 0 (the "Quit" button is focused by default). Returns a command that produces `DialogAcceptMsg{Action: DIALOGQUITCONFIRM}`.

If the user had tabbed to "Cancel" (focusIndex 1), it would produce `DialogCancelMsg{}` instead.

**Next:** Bubbletea runtime executes the command. `DialogAcceptMsg{Action: DIALOGQUITCONFIRM}` arrives at `Model.Update()`.

---

### Step 8: `DialogAcceptMsg` arrives at `Model.Update()`

**File:** `update.go:337`

The message is not a `tea.KeyMsg`, so the `esc`/`ctrl+c` guard is skipped. The type switch routes it:

```go
case DialogAcceptMsg:
    return m.UpdateDialog(msg)
```

**Next:** Calls `m.UpdateDialog(msg)`.

---

### Step 9: `UpdateDialog` handles the accept

**File:** `update_dialog.go:842`

```go
case DialogAcceptMsg:
    switch msg.Action {
    case DIALOGQUITCONFIRM:
        return m, tea.Quit
```

**Next:** Returns `tea.Quit`. The Bubbletea runtime shuts down the program.

---

## Alternative Path: User Cancels

If the user presses `esc` or `enter` on "Cancel" while the dialog is showing:

### Cancel via `esc` in dialog

At **Step 7**, `ToggleControls` handles `esc`:

```go
case "esc":
    return *d, func() tea.Msg { return DialogCancelMsg{} }
```

### Cancel via `enter` on "Cancel" button

At **Step 7**, if `focusIndex == 1`:

```go
return *d, func() tea.Msg { return DialogCancelMsg{} }
```

### `DialogCancelMsg` handling

**File:** `update_dialog.go:1350`

```go
case DialogCancelMsg:
    return m, tea.Batch(
        OverlayClearCmd(),
        FocusSetCmd(PAGEFOCUS),
    )
```

Clears the overlay and returns focus to the page. The dialog disappears.

## Alternative Path: Double-Esc Force Quit

If the user presses `esc` while the quit confirmation dialog is already showing:

At **Step 1**, the guard succeeds this time:

```go
case "esc":
    if d, ok := m.ActiveOverlay.(*DialogModel); ok && d != nil && d.Action == DIALOGQUITCONFIRM {
        return m, tea.Quit
    }
```

`ActiveOverlay` is a `*DialogModel` with `Action == DIALOGQUITCONFIRM`. Returns `tea.Quit` immediately — no second confirmation.

---

## File Reference

| File | Role in this flow |
|------|-------------------|
| `update.go:14` | Entry point — `Model.Update()`, top-level `esc`/`ctrl+c` guard, message type dispatch |
| `update.go:491` | Overlay key intercept — routes keystrokes to `ActiveOverlay.OverlayUpdate()` when overlay is active |
| `update_dialog.go:46` | `ShowQuitConfirmDialogMsg` handler — creates dialog, sets overlay + focus |
| `update_dialog.go:842` | `DialogAcceptMsg` handler — dispatches on `Action` to perform confirmed operation |
| `update_dialog.go:1350` | `DialogCancelMsg` handler — clears overlay, restores page focus |
| `dialog.go:117` | `DialogModel.Update()` — routes to control handler based on dialog action type |
| `dialog.go:146` | `DialogModel.OverlayUpdate()` — `ModalOverlay` adapter, delegates to `Update()` |
| `dialog.go:157` | `ToggleControls()` — handles tab/enter/esc for two-button dialogs |

---

## Example 2: New Datatype → Form Dialog → DB Create → List Refresh

User presses the `new` key on the Datatypes screen in browse phase. A form dialog appears with name, label, type, and parent fields. User fills the form and presses enter on "Confirm". The datatype is created in the database and the list refreshes.

### Step 1: Keypress arrives at `Model.Update()`

**File:** `update.go:14`

The Bubbletea runtime calls `Model.Update(msg)` with `tea.KeyMsg` matching the `new` action. The `esc`/`ctrl+c` guard doesn't match, so execution falls through.

No overlay is active, so the overlay intercept at `update.go:491` is skipped. No typed message matches the `switch` in `update.go:46`, so execution reaches the default delegate:

```go
// Delegate everything else to the screen
ctx := m.AppCtx()
screen, cmd := m.ActiveScreen.Update(ctx, msg)
m.ActiveScreen = screen
return m, cmd
```

**Next:** Calls `m.ActiveScreen.Update(ctx, msg)` — the `DatatypesScreen`.

---

### Step 2: `DatatypesScreen.Update` routes to `updateBrowse`

**File:** `screen_datatypes.go:213`

```go
case tea.KeyMsg:
    if s.Phase == DatatypesPhaseBrowse {
        return s.updateBrowse(ctx, msg)
    }
```

The screen is in browse phase (viewing the datatype tree, not editing fields).

**Next:** Calls `s.updateBrowse(ctx, msg)`.

---

### Step 3: `updateBrowse` matches `ActionNew`

**File:** `screen_datatypes.go:357`

```go
if km.Matches(key, config.ActionNew) {
    if s.AdminMode {
        return s, ShowAdminFormDialogCmd(FORMDIALOGCREATEADMINDATATYPE, "New Admin Datatype", s.AdminDatatypes)
    }
    return s, ShowFormDialogCmd(FORMDIALOGCREATEDATATYPE, "New Datatype", s.Datatypes)
}
```

Not in admin mode, so returns a command that will produce `ShowFormDialogMsg`. The current `s.Datatypes` slice is passed as the parent options for the form.

**Next:** Bubbletea runtime executes the command. `ShowFormDialogMsg{Action: FORMDIALOGCREATEDATATYPE, Title: "New Datatype", Parents: s.Datatypes}` arrives at `Model.Update()`.

---

### Step 4: `ShowFormDialogMsg` routes through `Model.Update()` to `UpdateDialog`

**File:** `update.go:239`

The message is not a `tea.KeyMsg`, so the `esc`/`ctrl+c` guard is skipped. The type switch matches:

```go
case ShowFormDialogMsg:
    return m.UpdateDialog(msg)
```

**Next:** Calls `m.UpdateDialog(msg)`.

---

### Step 5: `UpdateDialog` creates the form dialog

**File:** `update_dialog.go:1363`

```go
case ShowFormDialogMsg:
    dialog := NewFormDialog(msg.Title, msg.Action, msg.Parents)
    return m, tea.Batch(
        OverlaySetCmd(&dialog),
        FocusSetCmd(DIALOGFOCUS),
    )
```

Creates a `FormDialogModel` with text inputs for name, label, type selector, and parent selector populated from the passed datatypes. Returns two batched commands.

**Next:** Bubbletea runtime executes both commands. `OverlaySetMsg` stores the form dialog on `m.ActiveOverlay`. `SetPanelFocusMsg` sets `m.PanelFocus = DIALOGFOCUS`. The form dialog is now rendered and receiving input.

---

### Step 6: User fills the form and presses `enter` on "Confirm"

**File:** `update.go:491`

Key messages arrive at `Model.Update()`. The overlay intercept catches them:

```go
if m.ActiveOverlay != nil {
    if keyMsg, ok := msg.(tea.KeyMsg); ok {
        overlay, cmd := m.ActiveOverlay.OverlayUpdate(keyMsg)
        m.ActiveOverlay = overlay
        return m, cmd
    }
}
```

Each keystroke (typing name, label, tabbing between fields) goes to `FormDialogModel.OverlayUpdate()` which delegates to `FormDialogModel.Update()`. The form updates its internal text input state and returns nil commands.

When the user tabs to "Confirm" and presses `enter`:

**Next:** Calls `m.ActiveOverlay.OverlayUpdate(keyMsg)`.

---

### Step 7: `FormDialogModel.OverlayUpdate` delegates to `Update`

**File:** `form_dialog.go:519`

```go
func (d *FormDialogModel) OverlayUpdate(msg tea.KeyMsg) (ModalOverlay, tea.Cmd) {
    updated, cmd := d.Update(msg)
    return &updated, cmd
}
```

**Next:** Calls `d.Update(msg)`.

---

### Step 8: `FormDialogModel.Update` handles `enter` on Confirm

**File:** `form_dialog.go:371`

```go
case "enter":
    if d.focusIndex == FormDialogButtonConfirm {
        parentID := ""
        if d.HasParentSelector() && d.ParentIndex < len(d.ParentOptions) {
            parentID = d.ParentOptions[d.ParentIndex].Value
        }
        typeValue := d.TypeInput.Value()
        if d.HasTypeSelector() && d.TypeIndex < len(d.TypeOptions) {
            typeValue = d.TypeOptions[d.TypeIndex].Value
        }
        return *d, func() tea.Msg {
            return FormDialogAcceptMsg{
                Action:   d.Action,
                Name:     d.NameInput.Value(),
                Label:    d.LabelInput.Value(),
                Type:     typeValue,
                ParentID: parentID,
            }
        }
    }
```

Collects all field values from the form inputs and returns a command producing `FormDialogAcceptMsg` with action `FORMDIALOGCREATEDATATYPE`.

**Next:** Bubbletea runtime executes the command. `FormDialogAcceptMsg{Action: FORMDIALOGCREATEDATATYPE, Name: "...", Label: "...", Type: "...", ParentID: "..."}` arrives at `Model.Update()`.

---

### Step 9: `FormDialogAcceptMsg` routes to `UpdateDialog`

**File:** `update.go:383`

```go
case FormDialogAcceptMsg:
    return m.UpdateDialog(msg)
```

**Next:** Calls `m.UpdateDialog(msg)`.

---

### Step 10: `UpdateDialog` dispatches the create command

**File:** `update_dialog.go:1524`

```go
case FormDialogAcceptMsg:
    switch msg.Action {
    case FORMDIALOGCREATEDATATYPE:
        return m, tea.Batch(
            OverlayClearCmd(),
            FocusSetCmd(PAGEFOCUS),
            LoadingStartCmd(),
            CreateDatatypeFromDialogCmd(msg.Name, msg.Label, msg.Type, msg.ParentID),
        )
```

Returns four batched commands:
1. `OverlayClearCmd()` — clears the form dialog overlay
2. `FocusSetCmd(PAGEFOCUS)` — returns focus to the page
3. `LoadingStartCmd()` — shows loading indicator
4. `CreateDatatypeFromDialogCmd(...)` — produces `CreateDatatypeFromDialogRequestMsg`

**Next:** Bubbletea runtime executes all four. The first three update Model state (clear overlay, set focus, set loading). The fourth produces `CreateDatatypeFromDialogRequestMsg`.

---

### Step 11: `CreateDatatypeFromDialogRequestMsg` routes to `UpdateCms`

**File:** `update.go:263`

```go
case CreateDatatypeFromDialogRequestMsg:
    return m.UpdateCms(msg)
```

**Next:** Calls `m.UpdateCms(msg)`.

---

### Step 12: `UpdateCms` delegates to the handler

**File:** `update_cms.go:28`

```go
case CreateDatatypeFromDialogRequestMsg:
    return m, m.HandleCreateDatatypeFromDialog(msg)
```

**Next:** Calls `m.HandleCreateDatatypeFromDialog(msg)`, which returns a `tea.Cmd`.

---

### Step 13: `HandleCreateDatatypeFromDialog` performs the DB operation

**File:** `update_dialog.go:2275`

This is the async boundary. The method validates `authorID` and `cfg`, then returns a closure that runs on a background goroutine:

```go
return func() tea.Msg {
    d := db.ConfigDB(*cfg)
    ctx := context.Background()
    ac := middleware.AuditContextFromCLI(*cfg, authorID)

    // Determine next sort order
    maxSort, sortErr := d.GetMaxDatatypeSortOrder(parentID)

    // Create the datatype
    params := db.CreateDatatypeParams{
        DatatypeID:   types.NewDatatypeID(),
        Name:         msg.Name,
        Label:        msg.Label,
        Type:         dtype,
        ParentID:     parentID,
        SortOrder:    maxSort + 1,
        AuthorID:     authorID,
        DateCreated:  types.TimestampNow(),
        DateModified: types.TimestampNow(),
    }

    dt, err := d.CreateDatatype(ctx, ac, params)
    if err != nil {
        return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to create datatype: %v", err)}
    }
    return DatatypeCreatedFromDialogMsg{DatatypeID: dt.DatatypeID, Label: dt.Label}
}
```

Creates a ULID, calls `d.CreateDatatype()` through the `DbDriver` interface (which records an audit event). On success, returns `DatatypeCreatedFromDialogMsg`. On failure, returns `ActionResultMsg` with the error.

**Next:** Bubbletea runtime executes the closure on a goroutine. `DatatypeCreatedFromDialogMsg{DatatypeID: "...", Label: "..."}` arrives at `Model.Update()`.

---

### Step 14: `DatatypeCreatedFromDialogMsg` routes to `UpdateDialog`

**File:** `update.go:289`

```go
case DatatypeCreatedFromDialogMsg:
    return m.UpdateDialog(msg)
```

**Next:** Calls `m.UpdateDialog(msg)`.

---

### Step 15: `UpdateDialog` triggers a re-fetch

**File:** `update_dialog.go:1967`

```go
case DatatypeCreatedFromDialogMsg:
    return m, tea.Batch(
        LoadingStopCmd(),
        LogMessageCmd(fmt.Sprintf("Datatype created: %s", msg.Label)),
        AllDatatypesFetchCmd(),
    )
```

Returns three batched commands:
1. `LoadingStopCmd()` — hides loading indicator
2. `LogMessageCmd(...)` — writes to the log
3. `AllDatatypesFetchCmd()` — produces `AllDatatypesFetchMsg{}`

**Next:** Bubbletea runtime executes all three. `AllDatatypesFetchMsg{}` arrives at `Model.Update()`.

---

### Step 16: `AllDatatypesFetchMsg` delegates to the screen

**File:** `update.go:562`

`AllDatatypesFetchMsg` doesn't match any case in the Model-level type switch, so it falls through to the default screen delegate:

```go
ctx := m.AppCtx()
screen, cmd := m.ActiveScreen.Update(ctx, msg)
m.ActiveScreen = screen
return m, cmd
```

**Next:** Calls `DatatypesScreen.Update(ctx, msg)`.

---

### Step 17: `DatatypesScreen.Update` handles the fetch request

**File:** `screen_datatypes.go:222`

```go
case AllDatatypesFetchMsg:
    return s.handleAllDatatypesFetch(ctx)
```

**Next:** Calls `s.handleAllDatatypesFetch(ctx)`.

---

### Step 18: `handleAllDatatypesFetch` queries the database

**File:** `screen_datatypes.go:708`

```go
func (s *DatatypesScreen) handleAllDatatypesFetch(ctx AppContext) (Screen, tea.Cmd) {
    d := ctx.DB
    return s, func() tea.Msg {
        datatypes, err := d.ListDatatypes()
        if err != nil {
            return FetchErrMsg{Error: err}
        }
        return AllDatatypesFetchResultsMsg{Data: *datatypes}
    }
}
```

Returns a closure that queries all datatypes from the database. This is the second async DB call in the flow.

**Next:** Bubbletea runtime executes the closure. `AllDatatypesFetchResultsMsg{Data: [...]}` arrives at `Model.Update()`.

---

### Step 19: `AllDatatypesFetchResultsMsg` delegates to the screen

**File:** `update.go:562`

Same default delegate as Step 16 — falls through to `ActiveScreen.Update()`.

**Next:** Calls `DatatypesScreen.Update(ctx, msg)`.

---

### Step 20: `DatatypesScreen.Update` stores the results

**File:** `screen_datatypes.go:224`

```go
case AllDatatypesFetchResultsMsg:
    return s.handleAllDatatypesFetchResults(msg)
```

**Next:** Calls `s.handleAllDatatypesFetchResults(msg)`.

---

### Step 21: `handleAllDatatypesFetchResults` updates screen state

**File:** `screen_datatypes.go:725`

```go
func (s *DatatypesScreen) handleAllDatatypesFetchResults(msg AllDatatypesFetchResultsMsg) (Screen, tea.Cmd) {
    s.Datatypes = msg.Data
    s.rebuildTree()
    cmds := []tea.Cmd{LoadingStopCmd()}
    if s.Phase == DatatypesPhaseBrowse && len(s.FlatDTList) > 0 {
        cmds = append(cmds, s.fetchFieldsForCurrentDT())
    }
    return s, tea.Batch(cmds...)
}
```

Stores the new datatypes list, rebuilds the tree structure for rendering, stops loading, and fetches fields for the currently selected datatype.

**Next:** The screen's `View()` method renders the updated tree with the new datatype included. The field fetch continues as a separate message cycle.

---

## Message Flow Diagram

```
User presses "new"
    |
    v
Model.Update() --[KeyMsg]--> ActiveScreen.Update()
    |
    v
DatatypesScreen.updateBrowse() --[ActionNew]--> ShowFormDialogCmd()
    |
    v
Model.Update() --[ShowFormDialogMsg]--> UpdateDialog()
    |                                       |
    |              creates FormDialogModel, sets overlay + focus
    |
    v
[User fills form, presses Confirm]
    |
    v
Model.Update() --[KeyMsg]--> ActiveOverlay.OverlayUpdate()
    |
    v
FormDialogModel.Update() --[enter on Confirm]--> FormDialogAcceptMsg
    |
    v
Model.Update() --[FormDialogAcceptMsg]--> UpdateDialog()
    |                                         |
    |              clears overlay, starts loading, emits CreateDatatypeFromDialogCmd
    |
    v
Model.Update() --[CreateDatatypeFromDialogRequestMsg]--> UpdateCms()
    |
    v
HandleCreateDatatypeFromDialog() --[async DB call]--> d.CreateDatatype()
    |
    v
Model.Update() --[DatatypeCreatedFromDialogMsg]--> UpdateDialog()
    |                                                   |
    |              stops loading, emits AllDatatypesFetchCmd
    |
    v
Model.Update() --[AllDatatypesFetchMsg]--> ActiveScreen.Update()
    |
    v
DatatypesScreen.handleAllDatatypesFetch() --[async DB call]--> d.ListDatatypes()
    |
    v
Model.Update() --[AllDatatypesFetchResultsMsg]--> ActiveScreen.Update()
    |
    v
DatatypesScreen.handleAllDatatypesFetchResults()
    |
    s.Datatypes = msg.Data
    s.rebuildTree()
    View() renders updated list
```

## File Reference

| File | Role in this flow |
|------|-------------------|
| `update.go:14` | Entry point — `Model.Update()`, top-level guards, message type dispatch |
| `update.go:383` | `FormDialogAcceptMsg` routing to `UpdateDialog` |
| `update.go:263` | `CreateDatatypeFromDialogRequestMsg` routing to `UpdateCms` |
| `update.go:289` | `DatatypeCreatedFromDialogMsg` routing to `UpdateDialog` |
| `update.go:562` | Default delegate — unmatched messages forwarded to `ActiveScreen.Update()` |
| `screen_datatypes.go:213` | `DatatypesScreen.Update()` — routes `KeyMsg` to `updateBrowse` |
| `screen_datatypes.go:357` | `updateBrowse()` — matches `ActionNew`, emits `ShowFormDialogCmd` |
| `screen_datatypes.go:708` | `handleAllDatatypesFetch()` — async DB query for all datatypes |
| `screen_datatypes.go:725` | `handleAllDatatypesFetchResults()` — stores data, rebuilds tree |
| `update_dialog.go:1363` | `ShowFormDialogMsg` handler — creates `FormDialogModel`, sets overlay |
| `update_dialog.go:1524` | `FormDialogAcceptMsg` handler — clears dialog, emits create command |
| `update_dialog.go:1967` | `DatatypeCreatedFromDialogMsg` handler — stops loading, emits re-fetch |
| `update_dialog.go:2275` | `HandleCreateDatatypeFromDialog()` — async DB create with audit context |
| `update_cms.go:28` | `UpdateCms` routing — delegates to handler method |
| `form_dialog.go:371` | `FormDialogModel.Update()` — enter on Confirm collects values, emits `FormDialogAcceptMsg` |
| `form_dialog.go:519` | `FormDialogModel.OverlayUpdate()` — `ModalOverlay` adapter |
| `constructors.go:603` | `AllDatatypesFetchCmd()` — produces `AllDatatypesFetchMsg` |

---

## Example 3: New Content Node (Tree Phase) — With Bug Discovery

User presses `new` while viewing a content tree on the Content screen. The flow involves a multi-step dialog chain: select child datatype → fetch fields → build content form → create content. Tracing this flow uncovered two routing bugs that were fixed during documentation.

This example covers the **tree phase** "new" action (creating a child content node under the selected node). The **select phase** "new" (creating root content with a route) follows a different path through `ShowCreateRouteWithContentDialogCmd`.

### Step 1: Keypress arrives at `Model.Update()`

**File:** `update.go:14`

`tea.KeyMsg` matching the `new` action. No overlay is active. Falls through to the default screen delegate at `update.go:562`:

```go
ctx := m.AppCtx()
screen, cmd := m.ActiveScreen.Update(ctx, msg)
m.ActiveScreen = screen
return m, cmd
```

**Next:** Calls `ContentScreen.Update(ctx, msg)`.

---

### Step 2: `ContentScreen.Update` routes to tree key handler

**File:** `screen_content.go:215` (approximately)

```go
case tea.KeyMsg:
    // ... phase routing ...
    return s.updateTreePhase(ctx, msg)
```

The screen is in tree phase (viewing the content tree for a selected route).

**Next:** Calls `s.updateTreePhase(ctx, msg)`.

---

### Step 3: `updateTreePhase` matches `ActionNew`

**File:** `screen_content.go:780`

```go
if km.Matches(key, config.ActionNew) {
    if s.AdminMode {
        return s.handleAdminTreeNew(ctx)
    }
    return s.handleRegularTreeNew(ctx)
}
```

**Next:** Calls `s.handleRegularTreeNew(ctx)`.

---

### Step 4: `handleRegularTreeNew` emits fetch command

**File:** `screen_content.go:1010`

```go
func (s *ContentScreen) handleRegularTreeNew(ctx AppContext) (Screen, tea.Cmd) {
    node := s.Root.NodeAtIndex(s.Cursor)
    if node == nil {
        return s, ShowDialog("Error", "Please select a content node first", false)
    }
    rootDatatypeID := node.Datatype.DatatypeID
    if s.Root.Root != nil {
        rootDatatypeID = s.Root.Root.Datatype.DatatypeID
    }
    return s, ShowChildDatatypeDialogCmd(rootDatatypeID, s.PageRouteId)
}
```

Identifies the root datatype (always uses the tree root's datatype, not the selected node's). Returns a command producing `FetchChildDatatypesMsg`.

Note: `ShowChildDatatypeDialogCmd` is named as if it shows a dialog directly, but it actually produces `FetchChildDatatypesMsg` — the fetch happens first, then the dialog is shown with the results.

**Next:** Bubbletea runtime executes the command. `FetchChildDatatypesMsg{ParentDatatypeID: "...", RouteID: "..."}` arrives at `Model.Update()`.

---

### Step 5: `FetchChildDatatypesMsg` delegates to the screen

**File:** `update.go:570` (default delegate)

`FetchChildDatatypesMsg` doesn't match any case in the Model-level type switch, so it falls through to:

```go
ctx := m.AppCtx()
screen, cmd := m.ActiveScreen.Update(ctx, msg)
m.ActiveScreen = screen
return m, cmd
```

**Next:** Calls `ContentScreen.Update(ctx, msg)`.

---

### Step 6: `ContentScreen` handles the fetch asynchronously

**File:** `screen_content.go:270`

```go
case FetchChildDatatypesMsg:
    d := ctx.DB
    routeID := msg.RouteID
    rootDatatypeID := msg.ParentDatatypeID
    return s, func() tea.Msg {
        all, err := d.ListDatatypes()
        // ... error handling ...
        filtered := filterChildDatatypes(*all, rootDatatypeID)
        if len(filtered) == 0 {
            return ActionResultMsg{Title: "No Datatypes", Message: "..."}
        }
        return ShowChildDatatypeDialogMsg{
            ChildDatatypes: filtered,
            RouteID:        string(routeID),
        }
    }
```

Queries all datatypes, filters to eligible children of the root type. Returns `ShowChildDatatypeDialogMsg` with the filtered list.

**Next:** Bubbletea runtime executes the async closure. `ShowChildDatatypeDialogMsg{ChildDatatypes: [...], RouteID: "..."}` arrives at `Model.Update()`.

---

### Bug #1 Found Here (Fixed)

`ShowChildDatatypeDialogMsg` was **not routed** in the `ActiveScreen` type switch in `update.go`. It fell through to the screen (which doesn't handle it as a `tea.KeyMsg`), silently doing nothing. The child datatype selection dialog would never appear.

**Fix:** Added `case ShowChildDatatypeDialogMsg: return m.UpdateDialog(msg)` at `update.go:463`.

---

### Step 7: `ShowChildDatatypeDialogMsg` routes to `UpdateDialog`

**File:** `update.go:463`

```go
case ShowChildDatatypeDialogMsg:
    return m.UpdateDialog(msg)
```

**Next:** Calls `m.UpdateDialog(msg)`.

---

### Step 8: `UpdateDialog` creates the child datatype selection dialog

**File:** `update_dialog.go:1507`

```go
case ShowChildDatatypeDialogMsg:
    dialog := NewChildDatatypeDialog("Select Child Type", msg.ChildDatatypes, string(msg.RouteID))
    return m, tea.Batch(
        OverlaySetCmd(&dialog),
        FocusSetCmd(DIALOGFOCUS),
    )
```

Creates a `FormDialogModel` with `Action: FORMDIALOGCHILDDATATYPE`. The child datatypes are presented as a vertical selection list. The `EntityID` on the dialog stores the route ID.

**Next:** Dialog is rendered. User selects a child datatype and presses enter.

---

### Step 9: User selects a child datatype

**File:** `update.go:491` → overlay intercept → `FormDialogModel.OverlayUpdate` → `FormDialogModel.Update`

The user navigates the list and presses enter. The `FORMDIALOGCHILDDATATYPE` case in `FormDialogModel.Update` enters the parent-selection controls flow (since this dialog uses `ParentOptions` for the list):

**File:** `form_dialog.go:433`

```go
case "enter":
    if len(d.ParentOptions) > 0 && d.ParentIndex < len(d.ParentOptions) {
        parentID := d.ParentOptions[d.ParentIndex].Value
        return *d, func() tea.Msg {
            return FormDialogAcceptMsg{
                Action:   d.Action,
                EntityID: d.EntityID,  // route ID
                ParentID: parentID,    // selected datatype ID
            }
        }
    }
```

**Next:** `FormDialogAcceptMsg{Action: FORMDIALOGCHILDDATATYPE, EntityID: routeID, ParentID: datatypeID}` arrives at `Model.Update()`.

---

### Step 10: `FormDialogAcceptMsg` routes to `UpdateDialog`

**File:** `update.go:383`

```go
case FormDialogAcceptMsg:
    return m.UpdateDialog(msg)
```

**Next:** Calls `m.UpdateDialog(msg)`.

---

### Step 11: `UpdateDialog` dispatches `ChildDatatypeSelectedCmd`

**File:** `update_dialog.go:1594`

```go
case FORMDIALOGCHILDDATATYPE:
    if msg.ParentID != "" {
        return m, tea.Batch(
            OverlayClearCmd(),
            FocusSetCmd(PAGEFOCUS),
            ChildDatatypeSelectedCmd(types.DatatypeID(msg.ParentID), types.RouteID(msg.EntityID)),
        )
    }
```

Clears the dialog, returns focus to page, and emits `ChildDatatypeSelectedMsg` with the selected datatype ID and route ID.

**Next:** `ChildDatatypeSelectedMsg{DatatypeID: "...", RouteID: "..."}` arrives at `Model.Update()`.

---

### Bug #2 Found Here (Fixed)

`ChildDatatypeSelectedMsg` was routed to `UpdateDialog` at `update.go:465`, but `UpdateDialog` had **no handler** for it. The message was silently swallowed. Even after Bug #1 was fixed, the flow would still break here — the user would see the child datatype dialog but nothing would happen after selecting one.

**Fix:** Changed the routing to delegate to `ActiveScreen.Update()` instead of `UpdateDialog`, since the `ContentScreen` needs its cursor/tree context to determine the parent content ID. Added a handler in `ContentScreen.Update` that extracts the parent content ID from the current tree node and calls `FetchContentFieldsCmd`.

---

### Step 12: `ChildDatatypeSelectedMsg` delegates to the screen

**File:** `update.go:465`

```go
case ChildDatatypeSelectedMsg:
    if m.ActiveScreen != nil {
        ctx := m.AppCtx()
        screen, cmd := m.ActiveScreen.Update(ctx, msg)
        m.ActiveScreen = screen
        return m, cmd
    }
```

**Next:** Calls `ContentScreen.Update(ctx, msg)`.

---

### Step 13: `ContentScreen` fetches fields for the content form

**File:** `screen_content.go:300`

```go
case ChildDatatypeSelectedMsg:
    var parentID types.NullableContentID
    node := s.Root.NodeAtIndex(s.Cursor)
    if node != nil && node.Instance != nil {
        parentID = types.NullableContentID{ID: node.Instance.ContentDataID, Valid: true}
    }
    return s, FetchContentFieldsCmd(
        msg.DatatypeID,
        msg.RouteID,
        parentID,
        "New Content",
    )
```

Gets the parent content ID from the currently selected tree node (the node under which the new content will be created). Returns `FetchContentFieldsCmd`.

**Next:** `FetchContentFieldsMsg{DatatypeID, RouteID, ParentID, Title}` arrives at `Model.Update()`.

---

### Step 14: `FetchContentFieldsMsg` delegates to the screen

**File:** `update.go:570` (default delegate)

Falls through to `ContentScreen.Update()`.

---

### Step 15: `ContentScreen` queries fields and builds form message

**File:** `screen_content.go:314`

```go
case FetchContentFieldsMsg:
    d := ctx.DB
    return s, func() tea.Msg {
        fieldList, err := d.ListFieldsByDatatypeID(...)
        // ... error handling ...
        return ShowContentFormDialogMsg{
            Action:     FORMDIALOGCREATECONTENT,
            Title:      title,
            DatatypeID: datatypeID,
            RouteID:    routeID,
            ParentID:   parentID,
            Fields:     fields,
        }
    }
```

Async DB call fetches all fields for the selected datatype. Returns `ShowContentFormDialogMsg` with the field definitions, which `UpdateDialog` uses to build a dynamic content form.

**Next:** `ShowContentFormDialogMsg` arrives at `Model.Update()` → routed to `UpdateDialog` at `update.go:467` → creates a `ContentFormDialogModel` with text inputs for each field → overlay + focus set → user fills in the form → `ContentFormDialogAcceptMsg` → `CreateContentFromDialogRequestMsg` → async DB create → `ContentCreatedFromDialogMsg` → re-fetch tree.

The remaining steps (form display → submit → DB create → tree refresh) follow the same pattern as Example 2.

---

## Message Flow Diagram

```
User presses "new" on tree node
    |
    v
Model.Update() --[KeyMsg]--> ContentScreen.updateTreePhase()
    |
    v
handleRegularTreeNew() ---> ShowChildDatatypeDialogCmd()
    |
    v
Model.Update() --[FetchChildDatatypesMsg]--> ContentScreen.Update()
    |
    v
ContentScreen --[async DB: ListDatatypes + filter]--> ShowChildDatatypeDialogMsg
    |
    v
Model.Update() --[ShowChildDatatypeDialogMsg]--> UpdateDialog()   <-- BUG #1 was here
    |
    v
UpdateDialog() creates child datatype selection dialog
    |
    v
[User selects a child datatype, presses enter]
    |
    v
FormDialogModel --[enter]--> FormDialogAcceptMsg{FORMDIALOGCHILDDATATYPE}
    |
    v
Model.Update() --[FormDialogAcceptMsg]--> UpdateDialog()
    |
    v
UpdateDialog() --[FORMDIALOGCHILDDATATYPE]--> ChildDatatypeSelectedCmd()
    |
    v
Model.Update() --[ChildDatatypeSelectedMsg]--> ContentScreen.Update()   <-- BUG #2 was here
    |
    v
ContentScreen extracts parent content ID from cursor, emits FetchContentFieldsCmd()
    |
    v
Model.Update() --[FetchContentFieldsMsg]--> ContentScreen.Update()
    |
    v
ContentScreen --[async DB: ListFieldsByDatatypeID]--> ShowContentFormDialogMsg
    |
    v
Model.Update() --[ShowContentFormDialogMsg]--> UpdateDialog()
    |
    v
[Content form dialog → user fills fields → submit → DB create → tree refresh]
```

## Bugs Fixed During This Trace

| Bug | Symptom | Root Cause | Fix |
|-----|---------|------------|-----|
| #1: `ShowChildDatatypeDialogMsg` not routed | Child datatype selection dialog never appeared after pressing "new" on a tree node | Message fell through to default screen delegate, which didn't handle it | Added `case ShowChildDatatypeDialogMsg: return m.UpdateDialog(msg)` in `update.go:463` |
| #2: `ChildDatatypeSelectedMsg` swallowed | After selecting a child datatype, nothing happened — no content form appeared | Routed to `UpdateDialog` which had no handler for it | Changed routing to `ActiveScreen.Update()` and added handler in `ContentScreen` that extracts parent content ID and calls `FetchContentFieldsCmd` |

## File Reference

| File | Role in this flow |
|------|-------------------|
| `update.go:14` | Entry point — `Model.Update()`, message type dispatch |
| `update.go:463` | `ShowChildDatatypeDialogMsg` routing to `UpdateDialog` (Bug #1 fix) |
| `update.go:465` | `ChildDatatypeSelectedMsg` routing to `ActiveScreen` (Bug #2 fix) |
| `update.go:383` | `FormDialogAcceptMsg` routing to `UpdateDialog` |
| `update.go:467` | `ShowContentFormDialogMsg` routing to `UpdateDialog` |
| `update.go:570` | Default delegate — unmatched messages forwarded to `ActiveScreen.Update()` |
| `screen_content.go:780` | `updateTreePhase()` — matches `ActionNew` |
| `screen_content.go:1010` | `handleRegularTreeNew()` — identifies root datatype, emits fetch |
| `screen_content.go:270` | `FetchChildDatatypesMsg` handler — async DB query, filters children |
| `screen_content.go:300` | `ChildDatatypeSelectedMsg` handler — extracts parent ID, fetches fields |
| `screen_content.go:314` | `FetchContentFieldsMsg` handler — async DB query, builds form message |
| `update_dialog.go:1507` | `ShowChildDatatypeDialogMsg` handler — creates selection dialog |
| `update_dialog.go:1594` | `FORMDIALOGCHILDDATATYPE` accept — emits `ChildDatatypeSelectedCmd` |
| `form_dialog.go:1073` | `ShowChildDatatypeDialogCmd()` — produces `FetchChildDatatypesMsg` |
| `form_dialog.go:1682` | `FetchContentFieldsCmd()` — produces `FetchContentFieldsMsg` |
| `form_dialog.go:433` | `FormDialogModel.Update()` — enter on parent selection emits `FormDialogAcceptMsg` |
