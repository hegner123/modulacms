Plan 4: ContentScreen Implementation

Context

All prior screens follow the Screen interface pattern (screen.go:17-31). Dual-mode screens
(DatatypesScreen, RoutesScreen, FieldTypesScreen) use an AdminMode bool. The ContentScreen is
the most complex screen due to its multi-phase UI (route list → tree browse → field edit) and
the volume of operations (CRUD, move, copy, reorder, publish, version, field edit).

Prerequisites: Plan 1 (tree), Plan 2 (DB), Plan 3 (messages/commands)

NOTE: Admin tree browsing, field editing, publishing, versioning, move, copy, and reorder
are NEW functionality. Only flat-list navigation with delete exists today for admin content
(admin_controls.go:448-501). The regular mode is a port of existing code. Effort estimates
and sub-phase scope reflect this asymmetry.

---

State Ownership Rules

ContentScreen OWNS all content-specific state after construction. Model fields for the same
data exist only for legacy compatibility during the coexistence period. Once ActiveScreen is
set, legacy code paths that read Model.Root, Model.RootContentSummary, etc. do not execute
(they are guarded by `if m.ActiveScreen != nil` returning before the legacy chain runs).

Ownership table:

| Field                       | Owner during ActiveScreen | Model field kept? | Reason                        |
|-----------------------------|---------------------------|-------------------|-------------------------------|
| Root (tree.Root)            | ContentScreen             | Yes (read-only)   | Legacy cleanup in Plan 5      |
| RootContentSummary          | ContentScreen             | Yes (read-only)   | Legacy cleanup in Plan 5      |
| AdminRootContentSummary     | ContentScreen             | Yes (read-only)   | Legacy cleanup in Plan 5      |
| SelectedContentFields       | ContentScreen             | Yes (read-only)   | Legacy cleanup in Plan 5      |
| AdminSelectedContentFields  | ContentScreen             | Yes (read-only)   | Legacy cleanup in Plan 5      |
| Versions / ShowVersionList  | ContentScreen             | Yes (read-only)   | Legacy cleanup in Plan 5      |
| PendingCursorContentID      | ContentScreen             | Yes (read-only)   | Legacy cleanup in Plan 5      |
| PageRouteId                 | ContentScreen             | Yes (read-only)   | Legacy cleanup in Plan 5      |
| FieldCursor                 | ContentScreen             | No                | Owned by screen only          |

Rule: Messages that update content state (TreeLoadedMsg, ContentDataSet, etc.) update ONLY
the ContentScreen via Screen.Update(). The coexistence protocol in update.go does NOT also
update Model fields for these messages. Model fields retain their last value from before
ActiveScreen was set and are not read during the ActiveScreen lifecycle.

Exception: Model.Root is updated in the coexistence protocol alongside the screen forwarding
ONLY if tree.Root is used by non-screen code paths (overlay rendering, dialog context). If
no such dependency exists after Plan 3 implementation, skip the Model-level update entirely.

---

ContentScreen Struct

File: internal/tui/screen_content.go (new)

type ContentScreen struct {
    AdminMode bool

    // Cursors and focus
    Cursor      int
    FieldCursor int
    PanelFocus  FocusPanel

    // Route list phase
    RootContentSummary      []db.ContentDataTopLevel      // regular mode
    AdminRootContentSummary []db.AdminContentDataTopLevel  // admin mode
    RootDatatypes           []db.Datatypes                 // regular mode root datatypes
    AdminRootDatatypes      []db.AdminDatatypes            // admin mode root datatypes

    // Tree browsing phase (entered after selecting a route)
    PageRouteId            types.RouteID        // regular: set when user selects a route
    AdminPageRouteId       types.AdminRouteID   // admin: set when user selects an admin route
    Root                   tree.Root            // tree state (admin IDs converted to ContentID by core)
    SelectedDatatypeID     types.DatatypeID
    SelectedContentFields  []ContentFieldDisplay       // regular
    AdminSelectedFields    []AdminContentFieldDisplay   // admin
    PendingCursorContentID types.ContentID

    // Version state
    ShowVersionList          bool
    Versions                 []db.ContentVersion
    AdminVersions            []db.AdminContentVersion
    VersionCursor            int
    VersionContentID         types.ContentID
    AdminVersionContentID    types.AdminContentID
    VersionRouteID           types.RouteID
    AdminVersionRouteID      types.AdminRouteID

    // Error state
    LastError    error  // last async operation error (nil = no error)
    ErrorContext string // human-readable context for the error (e.g., "loading tree")
}

---

Error Handling

Every async operation result message has a corresponding error variant. The screen handles
both success and error paths.

Error display: When LastError is non-nil, the right panel (RoutePanel) renders the error
context and message at the top, styled with the error style. The error clears on the next
successful operation or when the user presses any key.

Error message types from Plan 3 that ContentScreen must handle:

| Success Message              | Error Variant                     | Screen Behavior              |
|------------------------------|-----------------------------------|------------------------------|
| TreeLoadedMsg                | TreeLoadFailedMsg { Err error }   | Set LastError, keep old tree |
| AdminTreeLoadedMsg           | AdminTreeLoadFailedMsg            | Set LastError, keep old tree |
| ContentCreatedMsg            | ContentCreateFailedMsg            | Set LastError, stay in tree  |
| AdminContentCreatedMsg       | AdminContentCreateFailedMsg       | Set LastError, stay in tree  |
| ContentCopiedMsg             | ContentCopyFailedMsg              | Set LastError                |
| AdminContentCopiedMsg        | AdminContentCopyFailedMsg         | Set LastError                |
| ContentReorderedMsg          | ContentReorderFailedMsg           | Set LastError                |
| AdminContentReorderedMsg     | AdminContentReorderFailedMsg      | Set LastError                |
| PublishCompletedMsg          | PublishFailedMsg                  | Set LastError                |
| AdminPublishCompletedMsg     | AdminPublishFailedMsg             | Set LastError                |
| UnpublishCompletedMsg        | UnpublishFailedMsg                | Set LastError                |
| AdminUnpublishCompletedMsg   | AdminUnpublishFailedMsg           | Set LastError                |
| VersionRestoredMsg           | VersionRestoreFailedMsg           | Set LastError                |
| AdminVersionRestoredMsg      | AdminVersionRestoreFailedMsg      | Set LastError                |
| VersionsListedMsg            | VersionsListFailedMsg             | Set LastError                |
| AdminVersionsListedMsg       | AdminVersionsListFailedMsg        | Set LastError                |

If Plan 3 does not define these error variants, they must be added to msg_crud.go / msg_admin.go
BEFORE implementing Plan 4. Each error message struct contains a single `Err error` field.

The screen handles all error variants with a single helper:

func (s *ContentScreen) setError(context string, err error) {
    s.LastError = err
    s.ErrorContext = context
}

func (s *ContentScreen) clearError() {
    s.LastError = nil
    s.ErrorContext = ""
}

---

AppContext Extension

File: internal/tui/app_context.go

Add to AppContext struct:
  ActiveLocale     string   // current locale for i18n
  AccordionEnabled bool     // accordion panel mode

Update AppCtx() method in model.go to populate these from Model state:
  ActiveLocale:     m.ActiveLocale,
  AccordionEnabled: m.AccordionEnabled,

Locale plumbing: Cmd constructors that need locale (LoadAdminContentFieldsCmd,
LoadContentFieldsCmd, ReloadAdminContentTreeCmd) receive it via AppContext.ActiveLocale
passed as a parameter from the screen's Update() method. The screen does NOT store locale
itself — it reads ctx.ActiveLocale on each Update() call.

---

Constructor

Pass individual fields, NOT *Model. Matches the pattern of every other screen constructor
(DatatypesScreen, RoutesScreen, FieldTypesScreen).

func NewContentScreen(
    adminMode bool,
    rootContentSummary []db.ContentDataTopLevel,
    adminRootContentSummary []db.AdminContentDataTopLevel,
    rootDatatypes []db.Datatypes,
    adminRootDatatypes []db.AdminDatatypes,
    pageRouteId types.RouteID,
) *ContentScreen {
    return &ContentScreen{
        AdminMode:               adminMode,
        Cursor:                  0,
        FieldCursor:             0,
        PanelFocus:              TreePanel,
        RootContentSummary:      rootContentSummary,
        AdminRootContentSummary: adminRootContentSummary,
        RootDatatypes:           rootDatatypes,
        AdminRootDatatypes:      adminRootDatatypes,
        PageRouteId:             pageRouteId,
    }
}

Wired in screenForPage():
  case CONTENT:
      return NewContentScreen(false,
          m.RootContentSummary, nil,
          m.RootDatatypes, nil,
          m.PageRouteId,
      )
  case ADMINCONTENT:
      return NewContentScreen(true,
          nil, m.AdminRootContentSummary,
          nil, m.AdminRootDatatypes,
          "",  // regular route ID unused in admin mode
      )

AdminPageRouteId: Not passed via constructor because it does not exist on Model. It is set
by the screen's handleRouteListKeys() when the user selects an admin route from the list.
This is local screen state, not Model state.

---

PageIndex()

Returns ADMINCONTENT if AdminMode, CONTENT otherwise.

---

Panel Focus Cycle

ContentPanel (center) is a preview panel. It is SKIPPED in the focus cycle:
- Tab from TreePanel → RoutePanel (skip ContentPanel)
- Shift-Tab from RoutePanel → TreePanel (skip ContentPanel)

This avoids the UX confusion of landing on a non-interactive panel. The center panel always
renders the preview for the currently selected tree node.

In handleTreeKeys, right arrow ("l", "right") moves focus to RoutePanel directly.

---

File Split

Given the expected size (1400+ lines), split into two files:

screen_content.go (~600-700 lines):
  - Struct, constructor, PageIndex, error helpers
  - Update() main dispatcher
  - handleRouteListKeys()
  - handleTreeKeys()
  - handleFieldPanelKeys()
  - handleVersionListKeys()
  - handleBackKey()
  - Message handlers for data refresh and operation results (success + error)

screen_content_view.go (~600-700 lines):
  - View() main layout dispatcher
  - renderRouteList() — ported from renderRootContentSummaryList
  - renderRouteDetail() — ported from renderContentSummaryDetail
  - renderRouteActions() — ported from renderContentActions
  - renderTree() — ported from CMSPage.ProcessTreeDatatypes
  - renderContentPreview() — ported from CMSPage.ProcessContentPreview
  - renderFields() — ported from CMSPage.ProcessFields
  - renderVersionList() — ported from renderVersionList
  - renderAdminList() — ported from renderAdminContentList
  - renderAdminDetail() — ported from renderAdminContentDetail
  - renderError() — renders LastError in right panel header
  - Tree traversal helpers (traverseTree, traverseTreeWithDepth, FormatTreeRow)
  - DecideNodeName stays package-level (used by update_controls.go:749)

---

Update() Method

Handles tea.KeyMsg by dispatching to helper methods based on state:

1. Clear LastError on any key press (user acknowledges error)
2. If ShowVersionList → handleVersionListKeys
3. If in tree phase (Root.Root != nil):
   a. PanelFocus == TreePanel → handleTreeKeys
   b. PanelFocus == RoutePanel → handleFieldPanelKeys
   (ContentPanel is skipped in focus cycle — no case needed)
4. If in route list phase → handleRouteListKeys
5. ActionBack/ActionDismiss always → handleBackKey

Handles data messages (non-key):

Tree loading:
- TreeLoadedMsg → update Root, reset cursor, load fields for first node, clearError
- AdminTreeLoadedMsg → same for admin, clearError
- TreeLoadFailedMsg → setError("loading tree", msg.Err)
- AdminTreeLoadFailedMsg → setError("loading admin tree", msg.Err)

Field loading:
- LoadContentFieldsMsg → update SelectedContentFields
- AdminLoadContentFieldsMsg → update AdminSelectedFields

Route/content data refresh:
- RootContentSummarySet → update RootContentSummary
- AdminContentDataSet → update AdminRootContentSummary
- RootDatatypesSet → update RootDatatypes
- AdminRootDatatypesFetchResultsMsg → update AdminRootDatatypes

Operation results — success (update state + trigger tree reload):
- ContentCreatedMsg / AdminContentCreatedMsg → reload tree, clearError
- ContentUpdatedFromDialogMsg / AdminContentUpdatedFromDialogMsg → reload tree, clearError
- ContentMovedMsg / AdminContentMovedMsg → reload tree, clearError
- ContentReorderedMsg / AdminContentReorderedMsg → reload tree, set PendingCursorContentID, clearError
- ContentCopiedMsg / AdminContentCopiedMsg → reload tree, set PendingCursorContentID, clearError
- PublishCompletedMsg / AdminPublishCompletedMsg → reload tree, clearError
- UnpublishCompletedMsg / AdminUnpublishCompletedMsg → reload tree, clearError
- VersionsListedMsg / AdminVersionsListedMsg → set ShowVersionList, Versions, clearError
- VersionRestoredMsg / AdminVersionRestoredMsg → hide version list, reload tree, clearError
- ContentFieldUpdatedMsg / AdminContentFieldUpdatedMsg → reload fields, clearError
- ContentFieldAddedMsg / AdminContentFieldAddedMsg → reload fields, clearError
- ContentFieldDeletedMsg / AdminContentFieldDeletedMsg → reload fields, clearError

Operation results — errors (all error variants):
- All *FailedMsg types → setError with appropriate context string

Each handler branches on AdminMode to use the appropriate ID types and commands.

NOTE on ContentUpdatedFromDialogMsg: This message type does not exist in the current codebase.
It is a regular-mode equivalent to AdminContentUpdatedFromDialogMsg. If it is not created in
Plan 3, handle the existing update result path instead (check update_cms.go for the actual
message type emitted after content update).

---

Dialog Flow Specification

Dialogs are triggered by the screen emitting Cmd functions that return dialog-opening messages.
These messages are handled by the existing dialog infrastructure:

1. Screen emits Cmd → returns a dialog-triggering message (e.g., ShowDeleteContentDialogMsg)
2. update.go coexistence block matches the dialog message → routes to UpdateDialog()
3. UpdateDialog creates the ModalOverlay, sets Model.ActiveOverlay
4. Overlay intercepts all key input (update.go:257-264) until user confirms/cancels
5. Overlay emits a result message (e.g., DeleteContentRequestMsg)
6. Result message flows through the legacy command handler chain (UpdateCms/AdminUpdateCms)
7. Command handler performs DB operation, emits completion message (e.g., ContentDeletedMsg)
8. Completion message falls through to Screen.Update() via update.go:291-295

Messages that need explicit update.go coexistence cases (route to UpdateDialog):

| Message Type                           | Handler          | Source                    |
|----------------------------------------|------------------|---------------------------|
| ShowDialogMsg                          | UpdateDialog     | Already handled (ln 111)  |
| ShowDatabaseFormDialogMsg              | UpdateDialog     | Already handled (ln 109)  |
| TogglePublishRequestMsg                | UpdateDialog     | Plan 3 wires this         |
| AdminTogglePublishRequestMsg           | UpdateDialog     | Plan 3 wires this         |
| ListVersionsRequestMsg                 | UpdateCms        | Plan 3 wires this         |
| AdminListVersionsRequestMsg            | AdminUpdateCms   | Plan 3 wires this         |
| RestoreVersionRequestMsg               | UpdateDialog     | Plan 3 wires this         |
| AdminRestoreVersionRequestMsg          | UpdateDialog     | Plan 3 wires this         |
| BuildContentFormMsg                    | UpdateDialog     | Existing                  |
| AdminBuildContentFormMsg               | UpdateDialog     | Plan 3 wires this         |
| ShowEditAdminContentFormDialogMsg      | UpdateDialog     | Plan 3 wires this         |
| DeleteAdminContentRequestMsg           | AdminUpdateCms   | Existing                  |
| ReorderSiblingRequestMsg               | UpdateCms        | Existing                  |
| AdminReorderSiblingRequestMsg          | AdminUpdateCms   | Plan 3 wires this         |
| CopyContentRequestMsg                  | UpdateCms        | Existing                  |
| AdminCopyContentRequestMsg             | AdminUpdateCms   | Plan 3 wires this         |
| MoveContentRequestMsg                  | UpdateCms        | Existing                  |
| AdminMoveContentRequestMsg             | AdminUpdateCms   | Plan 3 wires this         |

Messages that fall through to Screen.Update() with no explicit case needed:

All completion/result messages (*CreatedMsg, *UpdatedMsg, *DeletedMsg, *MovedMsg,
*ReorderedMsg, *CopiedMsg, *PublishCompletedMsg, *UnpublishCompletedMsg, *VersionsListedMsg,
*VersionRestoredMsg, *FieldUpdatedMsg, *FieldAddedMsg, *FieldDeletedMsg, *FailedMsg).

These are NOT matched by any explicit case in the update.go ActiveScreen block, so they
fall through to line 291-295 where Screen.Update() handles them.

VERIFICATION STEP: During implementation, grep update.go, update_cms.go, admin_update_cms.go,
update_dialog.go, and admin_update_dialog.go for each message type to confirm it is either:
(a) explicitly forwarded in update.go's ActiveScreen block, or
(b) handled by a legacy handler that runs BEFORE the ActiveScreen block (UpdateCms/etc.), or
(c) not matched anywhere else and falls through to Screen.Update().

If a message is caught by a legacy handler that does NOT run when ActiveScreen is non-nil,
it needs an explicit case in the ActiveScreen block to route it to the correct handler.

---

View() Method

Layout depends on state and mode:

Regular + route list:
  Panel 1: renderRouteList (route names with content counts)
  Panel 2: renderRouteDetail (selected route summary)
  Panel 3: renderRouteActions (available operations) + renderError if LastError

Regular + tree browse:
  Panel 1: renderTree (indented tree with expand/collapse indicators)
  Panel 2: renderContentPreview (selected node metadata)
  Panel 3: renderFields or renderVersionList + renderError if LastError

Admin + route list:
  Panel 1: renderAdminList (flat content list for selected route)
  Panel 2: renderAdminDetail (selected content detail)
  Panel 3: empty or renderRouteActions + renderError if LastError

Admin + tree browse:
  Panel 1: renderTree (same tree rendering, admin IDs converted)
  Panel 2: renderContentPreview (same, data is in ContentData shape after conversion)
  Panel 3: renderFields or renderVersionList + renderError if LastError

Panel titles are dynamic based on state (matching current cmsPanelTitles logic).
Context controls (status bar hints) generated based on state and panel focus.

---

Coexistence Protocol (update.go)

Add cases to the ActiveScreen dispatch block (update.go lines 28-296).

Messages that need Model-level mutation BEFORE screen forwarding:
- TreeLoadedMsg / AdminTreeLoadedMsg → update m.Root (ONLY if overlays read Model.Root;
  verify during implementation — if no overlay reads it, skip Model update entirely)
  then forward to screen.
- RootContentSummarySet / AdminContentDataSet → forward to screen only (no Model update;
  the screen owns this data while active).
- NavigateToPage messages → handled by existing navigation dispatch.

Messages routed to legacy handlers (NOT to screen):
- TogglePublishRequestMsg / AdminTogglePublishRequestMsg → UpdateDialog (creates confirm dialog)
- BuildContentFormMsg / AdminBuildContentFormMsg → UpdateDialog (creates content form)
- RestoreVersionRequestMsg / AdminRestoreVersionRequestMsg → UpdateDialog (creates confirm dialog)
- ReorderSiblingRequestMsg / AdminReorderSiblingRequestMsg → UpdateCms / AdminUpdateCms
- CopyContentRequestMsg / AdminCopyContentRequestMsg → UpdateCms / AdminUpdateCms
- MoveContentRequestMsg / AdminMoveContentRequestMsg → UpdateCms / AdminUpdateCms
- DeleteAdminContentRequestMsg → AdminUpdateCms
- ListVersionsRequestMsg / AdminListVersionsRequestMsg → UpdateCms / AdminUpdateCms

Messages that are pure screen forwarding (no Model mutation needed):
- LoadContentFieldsMsg / AdminLoadContentFieldsMsg
- All operation result messages (*CreatedMsg, *UpdatedMsg, *MovedMsg, etc.)
- All version result messages
- All field operation result messages
- All error messages (*FailedMsg)

Pattern for Model+screen dual update (TreeLoadedMsg example):
  case TreeLoadedMsg:
      m.Root = *msg.RootNode // only if overlays need Model.Root
      ctx := m.AppCtx()
      screen, cmd := m.ActiveScreen.Update(ctx, msg)
      m.ActiveScreen = screen
      return m, cmd

Pattern for legacy handler routing:
  case TogglePublishRequestMsg:
      return m.UpdateDialog(msg)
  case AdminTogglePublishRequestMsg:
      return m.UpdateDialog(msg)
  case ReorderSiblingRequestMsg:
      return m.UpdateCms(msg)
  case AdminReorderSiblingRequestMsg:
      return m.UpdateAdminCms(msg)
  // ... etc for each operation request message

Pattern for pure screen forwarding (falls through to line 291-295 automatically):
  // No explicit case needed — the default at line 291 handles it.

---

Implementation Sub-Phases

Break into 4 independently shippable sub-phases. Each compiles and runs. Each can be tested
before proceeding to the next.

Phase 4a: Route List (regular + admin)
  Files: screen_content.go, screen_content_view.go
  Scope:
  - ContentScreen struct (route list fields only — no tree/version/field state yet)
  - Constructor with route list data
  - PageIndex()
  - Update(): handleRouteListKeys (cursor up/down, select route → load tree)
  - Update(): RootContentSummarySet, AdminContentDataSet, RootDatatypesSet,
    AdminRootDatatypesFetchResultsMsg
  - View(): renderRouteList, renderRouteDetail, renderRouteActions
  - View(): renderAdminList, renderAdminDetail
  - Wire into screenForPage (screen.go)
  - AppContext extension (ActiveLocale, AccordionEnabled)
  - Verify: just check
  - Test: navigate to Content/AdminContent, see route list, cursor works
  Size: ~400-500 lines

Phase 4b: Tree Browsing (regular + admin)
  Files: screen_content.go, screen_content_view.go (extend)
  Scope:
  - Add tree fields to struct (Root, PendingCursorContentID, SelectedDatatypeID)
  - handleTreeKeys (cursor navigation, expand/collapse, select node)
  - handleBackKey (tree → route list, route list → history pop)
  - TreeLoadedMsg / AdminTreeLoadedMsg handling
  - TreeLoadFailedMsg / AdminTreeLoadFailedMsg handling
  - renderTree, renderContentPreview
  - Coexistence protocol entries in update.go for TreeLoadedMsg
  - Verify: just check
  - Test: select route, see tree, navigate nodes, back to route list
  Size: ~300-400 lines

Phase 4c: Field Editing + CRUD (regular + admin)
  Files: screen_content.go, screen_content_view.go (extend)
  Scope:
  - Add field state (SelectedContentFields, AdminSelectedFields, FieldCursor)
  - handleFieldPanelKeys (cursor, edit, add, delete field)
  - LoadContentFieldsMsg / AdminLoadContentFieldsMsg handling
  - ContentCreatedMsg, ContentCopiedMsg, ContentMovedMsg, ContentReorderedMsg (+ admin)
  - All *FailedMsg error handling for CRUD operations
  - renderFields, renderError
  - Coexistence protocol entries for operation request messages → legacy handlers
  - Verify: just check
  - Test: select node, see fields, edit field, create/copy/move/reorder content
  Size: ~300-400 lines

Phase 4d: Versioning + Publishing (regular + admin)
  Files: screen_content.go, screen_content_view.go (extend)
  Scope:
  - Add version state (ShowVersionList, Versions, AdminVersions, VersionCursor, etc.)
  - handleVersionListKeys (cursor, select version to restore)
  - VersionsListedMsg / AdminVersionsListedMsg handling
  - VersionRestoredMsg / AdminVersionRestoredMsg handling
  - PublishCompletedMsg / AdminPublishCompletedMsg handling
  - UnpublishCompletedMsg / AdminUnpublishCompletedMsg handling
  - All *FailedMsg error handling for version/publish operations
  - renderVersionList
  - Verify: just check
  - Test: list versions, restore version, publish/unpublish content
  Size: ~200-300 lines

---

Steps (ordered by sub-phase)

Phase 4a:
1. Extend AppContext with ActiveLocale, AccordionEnabled
2. Update AppCtx() in model.go to populate new fields
3. Create screen_content.go with struct (route list fields), constructor, PageIndex
4. Create screen_content_view.go with View and route list render methods
5. Implement Update() for route list keys and data refresh messages
6. Wire into screenForPage (screen.go)
7. Verify: just check

Phase 4b:
8. Add tree fields to ContentScreen struct
9. Implement handleTreeKeys, handleBackKey
10. Implement TreeLoadedMsg / AdminTreeLoadedMsg / *FailedMsg handling in Update()
11. Add renderTree, renderContentPreview to view file
12. Add coexistence protocol entries in update.go for tree messages
13. Verify: just check

Phase 4c:
14. Add field state to ContentScreen struct
15. Implement handleFieldPanelKeys
16. Implement field/CRUD message handling in Update() (success + error)
17. Add renderFields, renderError to view file
18. Add coexistence protocol entries in update.go for operation request messages
19. Verify: just check

Phase 4d:
20. Add version state to ContentScreen struct
21. Implement handleVersionListKeys
22. Implement version/publish message handling in Update() (success + error)
23. Add renderVersionList to view file
24. Verify: just check

Post-implementation:
25. Manual walkthrough: navigate to Content page, browse routes, enter tree, edit fields
26. Manual walkthrough: same for AdminContent page
27. Verify all message types from Plan 3 checklist have cases in Update()

---

Size estimate: ~1300-1500 lines across 2 files (unchanged, but spread across 4 sub-phases).

---

Risk assessment

Medium-high risk due to volume. Key risks:
- State synchronization between ContentScreen and Model (MITIGATED: ownership table above
  establishes ContentScreen as sole owner; Model fields are not read during ActiveScreen lifecycle)
- Missing message case in Update() → operation completes but screen doesn't update
  (MITIGATED: dialog flow table above enumerates every message and its routing; verification
  step in Phase 4c confirms each message reaches the screen)
- Rendering regressions from ported view code (MITIGATED: sub-phases allow incremental
  visual comparison against current behavior)
- Dialog-opening messages caught by wrong handler (MITIGATED: dialog flow specification above
  traces the complete path from screen Cmd → dialog → result → screen)
- Async operation errors silently swallowed (MITIGATED: error handling section above defines
  error variants for every operation and a consistent display pattern)
- Admin tree browsing is new functionality, not a port (MITIGATED: acknowledged in context
  section; effort is front-loaded in Phase 4b where tree browsing is implemented for both modes)

Mitigation:
- Keep a checklist of all message types from Plan 3 and verify each has a case in Update()
- Test each sub-phase individually before proceeding to the next
- Compare rendered output against current behavior before removing legacy code (Plan 5)
- Grep update.go for each dialog/operation message type to confirm correct routing
