Plan 3: Messages, Commands & Dialog Infrastructure

Context

The regular content editor has 13 command handlers, ~15 Cmd constructors, and ~20 message types
driving tree browsing, CRUD, field editing, publishing, versioning, copy, move, and reorder.
The admin content editor currently has only DeleteAdminContentCmd (admin_constructors.go:174-177)
and AdminContentBrowserControls (admin_controls.go:448-501) with flat-list navigation.

This plan creates all the admin content message types, Cmd constructors, command handlers, and
DialogContext entries needed for feature parity. It does NOT create the Screen — that is Plan 4.

Prerequisites: Plan 1 (tree.Root.LoadFromAdminData — DONE), Plan 2 (missing DB queries)

---

DB Method Availability

All async Cmd constructors and handlers reference DbDriver methods. These have been verified
to exist on the interface (internal/db/db.go):

  Present today (no Plan 2 dependency):
  - GetAdminContentData(types.AdminContentID) (*AdminContentData, error)
  - CreateAdminContentData(ctx, ac, CreateAdminContentDataParams) (*AdminContentData, error)
  - UpdateAdminContentData(ctx, ac, UpdateAdminContentDataParams) (*string, error)
  - DeleteAdminContentData(ctx, ac, types.AdminContentID) error
  - ListAdminContentDataWithDatatypeByRoute(types.NullableAdminRouteID) (*[]AdminContentDataWithDatatypeRow, error)
  - ListAdminContentFieldsByContentDataAndLocale(types.NullableAdminContentID, string) (*[]AdminContentFields, error)
  - ListAdminContentFieldsByRouteAndLocale(types.NullableAdminRouteID, string) (*[]AdminContentFields, error)
  - ListAdminFieldsByDatatypeID(types.NullableAdminDatatypeID) (*[]AdminFields, error)
  - CreateAdminContentField(ctx, ac, CreateAdminContentFieldParams) (*AdminContentFields, error)
  - UpdateAdminContentField(ctx, ac, UpdateAdminContentFieldParams) (*string, error)
  - DeleteAdminContentField(ctx, ac, types.AdminContentFieldID) error
  - ListAdminContentVersionsByContent(types.AdminContentID) (*[]AdminContentVersion, error)
  - ListAdminContentFieldsWithFieldByRoute(types.NullableAdminRouteID) (*[]AdminContentFieldsWithFieldRow, error)
  - ListAdminDatatypes() (*[]AdminDatatypes, error)

  From Plan 2 (required before implementation):
  - ListAdminContentFieldsWithFieldByContentData — needed by LoadAdminContentFieldsCmd
  - ListAdminContentFieldsByContentDataIDs — needed by HandleCopyAdminContent

  If Plan 2 is incomplete, Parts D and E can still be implemented for all handlers that do NOT
  use the missing methods. The following handlers are blocked until Plan 2 is done:
  - HandleCopyAdminContent (handler #7)

  All other handlers use only methods that exist today.

Publishing function signatures (internal/publishing/):
  - PublishAdminContent(ctx, d, rootID, locale, userID, ac, retentionCap, dispatcher) error
  - UnpublishAdminContent(ctx, d, rootID, locale, userID, ac, dispatcher) error
  - RestoreAdminContent(ctx, d, adminContentDataID, versionID, userID, ac) (*RestoreResult, error)

---

Part A: Message Types

File: internal/tui/msg_admin.go (append to existing)

Add the following message types. Each mirrors its regular equivalent but uses admin ID types.

Tree loading:
- AdminTreeLoadedMsg { RootNode *tree.Root; Stats *tree.LoadStats }

Content CRUD results (update existing types in msg_crud.go — see Part A-1):
- AdminContentCreatedMsg — add AdminRouteID field to existing type (msg_crud.go:347-350)
- AdminContentDeletedMsg — add AdminRouteID field to existing type (msg_crud.go:352-355)
- AdminContentUpdatedFromDialogMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }

Form construction (add to msg_admin.go):
- AdminBuildContentFormMsg { AdminDatatypeID types.AdminDatatypeID; AdminRouteID types.AdminRouteID; Fields []db.AdminFields }
- AdminFetchContentForEditMsg { AdminContentID types.AdminContentID; AdminDatatypeID types.AdminDatatypeID; AdminRouteID types.AdminRouteID; Title string }
- ShowEditAdminContentFormDialogMsg { AdminContentID types.AdminContentID; AdminDatatypeID types.AdminDatatypeID; AdminRouteID types.AdminRouteID; Fields []ExistingAdminContentField }

Tree operations:
- AdminReorderSiblingRequestMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID; Direction string }
- AdminContentReorderedMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID; Direction string }
- AdminCopyContentRequestMsg { SourceID types.AdminContentID; AdminRouteID types.AdminRouteID }
- AdminContentCopiedMsg { NewID types.AdminContentID; AdminRouteID types.AdminRouteID }
- AdminMoveContentRequestMsg { SourceID types.AdminContentID; TargetID types.AdminContentID; AdminRouteID types.AdminRouteID }
- AdminContentMovedMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }

Publishing:
- AdminTogglePublishRequestMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }
- AdminPublishCompletedMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }
- AdminUnpublishCompletedMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }

Versioning:
- AdminListVersionsRequestMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }
- AdminVersionsListedMsg { Versions []db.AdminContentVersion; AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }
- AdminRestoreVersionRequestMsg { AdminContentID types.AdminContentID; VersionID types.AdminContentVersionID; AdminRouteID types.AdminRouteID }
- AdminVersionRestoredMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID; FieldsRestored int }

Field operations:
- AdminContentFieldUpdatedMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }
- AdminContentFieldAddedMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }
- AdminContentFieldDeletedMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }

Dialog show messages (for confirmation dialogs):
- ShowDeleteAdminContentDialogMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID; ContentName string; HasChildren bool }
- ShowPublishAdminContentDialogMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID; Name string; IsPublished bool }
- ShowRestoreAdminVersionDialogMsg { AdminContentID types.AdminContentID; VersionID types.AdminContentVersionID; AdminRouteID types.AdminRouteID; VersionNumber int64 }
- ShowDeleteAdminContentFieldDialogMsg { AdminContentFieldID types.AdminContentFieldID; AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID; AdminDatatypeID types.NullableAdminDatatypeID; Label string }
- ShowMoveAdminContentDialogMsg { SourceNode *tree.Node; AdminRouteID types.AdminRouteID; Targets []tree.MoveTarget }

Confirmed action messages (emitted after dialog accept, consumed by handlers):
- ConfirmedDeleteAdminContentMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }
- ConfirmedPublishAdminContentMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }
- ConfirmedUnpublishAdminContentMsg { AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID }
- ConfirmedRestoreAdminVersionMsg { AdminContentID types.AdminContentID; VersionID types.AdminContentVersionID; AdminRouteID types.AdminRouteID }
- ConfirmedDeleteAdminContentFieldMsg { AdminContentFieldID types.AdminContentFieldID; AdminContentID types.AdminContentID; AdminRouteID types.AdminRouteID; AdminDatatypeID types.NullableAdminDatatypeID }

Also add ExistingAdminContentField struct (mirrors ExistingContentField):
  type ExistingAdminContentField struct {
      AdminContentFieldID types.AdminContentFieldID
      AdminFieldID        types.AdminFieldID
      Label               string
      Type                string
      Value               string
      ValidationJSON      string
      DataJSON            string
  }

Also update AdminContentFieldDisplay (msg_crud.go:362-369) to add:
  ValidationJSON string
  DataJSON       string

---

Part A-1: Existing Message Type Updates

File: internal/tui/msg_crud.go

The following types already exist but need AdminRouteID added for tree reload after mutations:

  AdminContentCreatedMsg (line 347-350):
    ADD: AdminRouteID types.AdminRouteID

  AdminContentDeletedMsg (line 352-355):
    ADD: AdminRouteID types.AdminRouteID

  DeleteAdminContentRequestMsg (line 357-360):
    ADD: AdminRouteID types.AdminRouteID

These updates require fixing all existing call sites:
  - admin_constructors.go:174-177 — DeleteAdminContentCmd: add adminRouteID parameter
  - admin_update_cms.go:124-131 — AdminContentDeletedMsg handler: use msg.AdminRouteID for tree reload

---

Part B: DialogContext Additions

File: internal/tui/model.go (DialogContext struct, lines 284-313)

Add fields:
  DeleteAdminContent       *DeleteAdminContentContext
  PublishAdminContent      *PublishAdminContentContext
  RestoreAdminVersion      *RestoreAdminVersionContext
  MoveAdminContent         *MoveAdminContentContext
  DeleteAdminContentField  *DeleteAdminContentFieldContext
  EditAdminSingleField     *editAdminSingleFieldCtx
  AddAdminContentField     *addAdminContentFieldCtx

Define context structs in msg_admin.go (keep with other admin message types):

  type DeleteAdminContentContext struct {
      AdminContentID types.AdminContentID
      AdminRouteID   types.AdminRouteID
      Name           string
      HasChildren    bool
  }

  type PublishAdminContentContext struct {
      AdminContentID types.AdminContentID
      AdminRouteID   types.AdminRouteID
      Name           string
      IsPublished    bool
  }

  type RestoreAdminVersionContext struct {
      AdminContentID types.AdminContentID
      VersionID      types.AdminContentVersionID
      AdminRouteID   types.AdminRouteID
      VersionNumber  int64
  }

  type MoveAdminContentContext struct {
      SourceNode     *tree.Node
      AdminRouteID   types.AdminRouteID
  }

  type DeleteAdminContentFieldContext struct {
      AdminContentFieldID types.AdminContentFieldID
      AdminContentID      types.AdminContentID
      AdminRouteID        types.AdminRouteID
      AdminDatatypeID     types.NullableAdminDatatypeID
      Label               string
  }

  type editAdminSingleFieldCtx struct {
      AdminContentFieldID types.AdminContentFieldID
      AdminContentID      types.AdminContentID
      AdminFieldID        types.AdminFieldID
      AdminRouteID        types.AdminRouteID
      AdminDatatypeID     types.NullableAdminDatatypeID
      Label               string
      Type                string
      Value               string
  }

  type addAdminContentFieldCtx struct {
      AdminContentID  types.AdminContentID
      AdminRouteID    types.AdminRouteID
      AdminDatatypeID types.NullableAdminDatatypeID
  }

---

Part C: AdminRootDatatypes Fetch

File: internal/tui/msg_admin.go

Add message:
  AdminRootDatatypesFetchResultsMsg { RootDatatypes []db.AdminDatatypes }

File: internal/tui/admin_constructors.go

Add:
  FetchAdminRootDatatypesCmd(config *config.Config) tea.Cmd
    - Calls d.ListAdminDatatypes()
    - Filters in Go: keep only entries where ParentID is null/empty (root-level datatypes)
    - Returns AdminRootDatatypesFetchResultsMsg

  Rationale: No dedicated "root-only" DB query exists. ListAdminDatatypes returns all datatypes.
  The set is small (admin datatypes are schema definitions, not content), so in-memory filtering
  is appropriate. If the set ever grows large, add a ListAdminRootDatatypes query in a future plan.

File: internal/tui/model.go

Add field (if not already present):
  AdminRootDatatypes []db.AdminDatatypes

The fetch is triggered when navigating to admin content page (same as AdminAllDatatypesFetchMsg
but filtered for root types only).

---

Part D: Cmd Constructors

File: internal/tui/admin_constructors.go (append to existing)

Add Cmd functions that wrap message emission (simple Cmds):
- AdminReorderSiblingCmd(adminContentID, adminRouteID, direction) → AdminReorderSiblingRequestMsg
- AdminCopyContentCmd(sourceID, adminRouteID) → AdminCopyContentRequestMsg
- AdminTogglePublishCmd(adminContentID, adminRouteID) → AdminTogglePublishRequestMsg
- AdminListVersionsCmd(adminContentID, adminRouteID) → AdminListVersionsRequestMsg
- AdminMoveContentCmd(sourceID, targetID, adminRouteID) → AdminMoveContentRequestMsg

Add Cmd functions that perform async DB work:
- ReloadAdminContentTreeCmd(config *config.Config, adminRouteID types.AdminRouteID) tea.Cmd
    1. Get driver: d := db.ConfigDB(*config)
    2. Call d.ListAdminContentDataWithDatatypeByRoute(adminRouteID)
    3. Split result into []AdminContentData and []AdminDatatypes slices
    4. Call d.ListAdminContentFieldsWithFieldByRoute(adminRouteID)
    5. Split result into []AdminContentFields and []AdminFields slices
    6. Build tree: root := tree.NewRoot(); stats, err := root.LoadFromAdminData(cd, dt, cf, df)
    7. Return AdminTreeLoadedMsg{RootNode: root, Stats: stats}

- LoadAdminContentFieldsCmd(cfg *config.Config, adminContentDataID types.AdminContentID, adminDatatypeID types.AdminDatatypeID, locale string) tea.Cmd
    1. Get driver: d := db.ConfigDB(*cfg)
    2. Call d.ListAdminContentFieldsByContentDataAndLocale(adminContentDataID, locale)
    3. Call d.ListAdminFieldsByDatatypeID(adminDatatypeID)
    4. Build map of existing values, iterate canonical field list
    5. Return AdminLoadContentFieldsMsg{Fields: []AdminContentFieldDisplay}

  Locale is passed explicitly from m.ActiveLocale at the call site (same pattern as
  LoadContentFieldsForLocaleCmd in the regular content path).

- FetchAdminContentForEditCmd(cfg *config.Config, adminContentID types.AdminContentID, adminDatatypeID types.AdminDatatypeID, adminRouteID types.AdminRouteID, title string, locale string) tea.Cmd
    1. Get driver: d := db.ConfigDB(*cfg)
    2. Call d.ListAdminContentFieldsByContentDataAndLocale(adminContentID, locale)
    3. Call d.ListAdminFieldsByDatatypeID(adminDatatypeID)
    4. Build ExistingAdminContentField slice with values from map
    5. Return ShowEditAdminContentFormDialogMsg

  Locale is passed explicitly from m.ActiveLocale at the call site.

Dialog-triggering Cmds (emit Show*Msg, handled in update_dialog.go):
- ShowAdminChildDatatypeDialogCmd(rootAdminDatatypeID, adminRouteID)
- ShowCreateAdminRouteWithContentDialogCmd(adminRootDatatypes)
- ShowDeleteAdminContentDialogCmd(adminContentID, adminRouteID, name, hasChildren)
- ShowPublishAdminContentDialogCmd(adminContentID, adminRouteID, name, isPublished)
- ShowRestoreAdminVersionDialogCmd(adminContentID, versionID, adminRouteID, versionNumber)
- ShowDeleteAdminContentFieldDialogCmd(adminContentFieldID, adminContentID, adminRouteID, adminDatatypeID, label)
- ShowMoveAdminContentDialogCmd(node, adminRouteID, targets)

---

Part E: Command Handlers

File: internal/tui/admin_commands.go (new)

Port the following handlers from commands.go. Each uses the admin DB methods and
middleware.AuditContextFromCLI for audit trail. All follow the same pattern:
func (m Model) HandleXxx(msg XxxMsg) tea.Cmd

Each handler accesses these from Model:
  cfg := m.Config           // *config.Config
  userID := m.UserID        // types.UserID
  locale := m.ActiveLocale  // string (for locale-dependent operations)
  // For publishing handlers additionally:
  dispatcher := m.Dispatcher  // publishing.WebhookDispatcher (or nil if not configured)

1. HandleCreateAdminContentFromDialog (port of commands.go:378-489)
   - Use d.CreateAdminContentData, d.ListAdminFieldsByDatatypeID, d.CreateAdminContentField
   - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
   - Return AdminContentCreatedMsg{AdminContentID, AdminRouteID}

2. HandleUpdateAdminContentFromDialog (port of commands.go:574-669)
   - Use d.ListAdminContentFieldsByContentDataAndLocale(id, m.ActiveLocale)
   - Use d.UpdateAdminContentField, d.CreateAdminContentField
   - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
   - Return AdminContentUpdatedFromDialogMsg{AdminContentID, AdminRouteID}

3. HandleDeleteAdminContent (port of commands.go:672-794)
   REPLACES the existing flat-delete handler in admin_update_dialog.go:884-913.
   The existing handler does not manage tree pointers — it calls DeleteAdminContentData
   directly. The new tree-aware version:
   - Fetches content via d.GetAdminContentData to check for children
   - If HasChildren: returns ActionResultMsg "Cannot delete — delete children first"
   - Rewires prev_sibling → next_sibling pointer
   - Rewires next_sibling → prev_sibling pointer
   - If first child of parent: updates parent's first_child_id to next sibling
   - Calls d.DeleteAdminContentData
   - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
   - Return AdminContentDeletedMsg{AdminContentID, AdminRouteID}

   Migration: Delete the existing HandleDeleteAdminContent from admin_update_dialog.go:884-913.
   The dispatch in admin_update_cms.go:118-119 changes from:
     case DeleteAdminContentRequestMsg: return m, m.HandleDeleteAdminContent(msg)
   to:
     case ConfirmedDeleteAdminContentMsg: return m, m.HandleDeleteAdminContent(msg)
   (The raw DeleteAdminContentRequestMsg is no longer used directly — deletion now goes
   through the confirmation dialog flow.)

   Known limitation: Tree pointer rewiring is not transactional. If the handler fails after
   detaching from the source position but before completing all pointer updates, the tree is
   left in an inconsistent state. This is the same limitation as the regular content delete
   handler. A future improvement would wrap the operation in a database transaction.

4. HandleMoveAdminContent (port of commands.go:799-1009)
   - Use admin update methods for pointer rewiring (detach from source + attach at target)
   - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
   - Return AdminContentMovedMsg{AdminContentID, AdminRouteID}

   Same non-transactional limitation as delete.

5. HandleAdminReorderSibling (port of commands.go:1773-2046)
   - Use admin update methods for sibling swap
   - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
   - Return AdminContentReorderedMsg{AdminContentID, AdminRouteID, Direction}

   Same non-transactional limitation as delete.

6. HandleCopyAdminContent (port of commands.go:2057-2237)
   - Use admin create + field copy methods
   - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
   - Return AdminContentCopiedMsg{NewID, AdminRouteID}
   - BLOCKED on Plan 2: requires ListAdminContentFieldsByContentDataIDs

7. HandleAdminConfirmedPublish (port of commands.go:2281-2311)
   - Call publishing.PublishAdminContent(ctx, d, msg.AdminContentID, m.ActiveLocale,
     m.UserID, ac, cfg.VersionRetentionCap, m.Dispatcher)
   - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
   - Return AdminPublishCompletedMsg{AdminContentID, AdminRouteID}

   Parameters sourced from Model:
     locale        = m.ActiveLocale
     retentionCap  = cfg.VersionRetentionCap (int from config.Config)
     dispatcher    = m.Dispatcher (publishing.WebhookDispatcher, may be nil)

8. HandleAdminConfirmedUnpublish (port of commands.go:2314-2343)
   - Call publishing.UnpublishAdminContent(ctx, d, msg.AdminContentID, m.ActiveLocale,
     m.UserID, ac, m.Dispatcher)
   - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
   - Return AdminUnpublishCompletedMsg{AdminContentID, AdminRouteID}

   Parameters sourced from Model:
     locale     = m.ActiveLocale
     dispatcher = m.Dispatcher

9. HandleAdminListVersions (port of commands.go:2370-2394)
   - Use d.ListAdminContentVersionsByContent(msg.AdminContentID)
   - No audit context (read-only)
   - Return AdminVersionsListedMsg{Versions, AdminContentID, AdminRouteID}

10. HandleAdminConfirmedRestoreVersion (port of commands.go:2397-2425)
    - Call publishing.RestoreAdminContent(ctx, d, msg.AdminContentID, msg.VersionID,
      m.UserID, ac)
    - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
    - Return AdminVersionRestoredMsg{AdminContentID, AdminRouteID, FieldsRestored}

11. HandleDeleteAdminContentField (port of commands.go:1363-1382)
    - Use d.DeleteAdminContentField(ctx, ac, msg.AdminContentFieldID)
    - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
    - Return AdminContentFieldDeletedMsg{AdminContentID, AdminRouteID}

12. HandleAddAdminContentField (port of commands.go:1384-1412)
    - Use d.CreateAdminContentField(ctx, ac, params)
    - Audit context: middleware.AuditContextFromCLI(*cfg, userID)
    - Return AdminContentFieldAddedMsg{AdminContentID, AdminRouteID}

Handler count: 12 (was 13 — FetchAdminContentForEdit collapsed into Part D as a Cmd constructor,
since it is read-only with no Model state mutation needed).

---

Part F: Admin Content Form Dialogs

This part adds admin content dialog handling. It covers two categories:
(1) Confirmation dialogs (simple OK/Cancel with context stored in DialogContext)
(2) Form dialogs (multi-field input forms using the FormDialog system)

Part F-1: New DialogAction Constants

File: internal/tui/dialog.go

Add to the DialogAction const block:
  DIALOGPUBLISHADMINCONTENT       DialogAction = "publish_admin_content"
  DIALOGUNPUBLISHADMINCONTENT     DialogAction = "unpublish_admin_content"
  DIALOGRESTOREADMINVERSION       DialogAction = "restore_admin_version"
  DIALOGDELETEADMINCONTENTFIELD   DialogAction = "delete_admin_content_field"

Note: DIALOGDELETEADMINCONTENT already exists (dialog.go:34). No new constant needed for delete.

Add all new constants to the ToggleControls case list in DialogModel.Update (dialog.go:115-121).

Part F-2: New FormDialogAction Constants

File: internal/tui/form_dialog.go

Add to the FormDialogAction const block:
  FORMDIALOGCREATEADMINCONTENT      FormDialogAction = "create_admin_content"
  FORMDIALOGEDITADMINCONTENT        FormDialogAction = "edit_admin_content"
  FORMDIALOGMOVEADMINCONTENT        FormDialogAction = "move_admin_content"
  FORMDIALOGCHILDADMINDATATYPE      FormDialogAction = "child_admin_datatype"
  FORMDIALOGADDADMINCONTENTFIELD    FormDialogAction = "add_admin_content_field"
  FORMDIALOGEDITADMINSINGLEFIELD    FormDialogAction = "edit_admin_single_field"

Part F-3: Confirmation Dialog Show Handlers

File: internal/tui/admin_update_dialog.go (append to existing dialog show handling)

Each Show*Msg stores context in DCtx, creates a DialogModel, and sets overlay+focus.

ShowDeleteAdminContentDialogMsg:
  - If msg.HasChildren: show non-cancellable DIALOGGENERIC "Cannot delete — has children"
  - Else: show DIALOGDELETEADMINCONTENT dialog "Delete '<name>'? All field values will be deleted."
  - Store: m.DCtx.DeleteAdminContent = &DeleteAdminContentContext{AdminContentID, AdminRouteID, Name, HasChildren}
  - Dialog buttons: "Delete" / "Cancel"

ShowPublishAdminContentDialogMsg:
  - If msg.IsPublished: show DIALOGUNPUBLISHADMINCONTENT "Unpublish '<name>'?"
  - Else: show DIALOGPUBLISHADMINCONTENT "Publish '<name>'?"
  - Store: m.DCtx.PublishAdminContent = &PublishAdminContentContext{AdminContentID, AdminRouteID, Name, IsPublished}
  - Dialog buttons: "Confirm" / "Cancel"

ShowRestoreAdminVersionDialogMsg:
  - Show DIALOGRESTOREADMINVERSION "Restore version <N>? Current field values will be overwritten."
  - Store: m.DCtx.RestoreAdminVersion = &RestoreAdminVersionContext{AdminContentID, VersionID, AdminRouteID, VersionNumber}
  - Dialog buttons: "Restore" / "Cancel"

ShowDeleteAdminContentFieldDialogMsg:
  - Show DIALOGDELETEADMINCONTENTFIELD "Delete field '<label>'?"
  - Store: m.DCtx.DeleteAdminContentField = &DeleteAdminContentFieldContext{...all fields from msg}
  - Dialog buttons: "Delete" / "Cancel"

ShowMoveAdminContentDialogMsg:
  - If len(msg.Targets) == 0: show DIALOGGENERIC "No valid move targets available."
  - Else: build a form dialog with FORMDIALOGMOVEADMINCONTENT, populate with target names
  - Store: m.DCtx.MoveAdminContent = &MoveAdminContentContext{SourceNode, AdminRouteID}

Part F-4: Confirmation Dialog Accept Handlers

File: internal/tui/admin_update_dialog.go (in the DialogAcceptMsg switch)

case DIALOGDELETEADMINCONTENT:
  if m.DCtx.DeleteAdminContent != nil {
      ctx := m.DCtx.DeleteAdminContent
      m.DCtx.DeleteAdminContent = nil
      return m, tea.Batch(
          OverlayClearCmd(),
          FocusSetCmd(PAGEFOCUS),
          LoadingStartCmd(),
          func() tea.Msg { return ConfirmedDeleteAdminContentMsg{ctx.AdminContentID, ctx.AdminRouteID} },
      )
  }

case DIALOGPUBLISHADMINCONTENT:
  if m.DCtx.PublishAdminContent != nil {
      ctx := m.DCtx.PublishAdminContent
      m.DCtx.PublishAdminContent = nil
      return m, tea.Batch(
          OverlayClearCmd(), FocusSetCmd(PAGEFOCUS), LoadingStartCmd(),
          func() tea.Msg { return ConfirmedPublishAdminContentMsg{ctx.AdminContentID, ctx.AdminRouteID} },
      )
  }

case DIALOGUNPUBLISHADMINCONTENT:
  if m.DCtx.PublishAdminContent != nil {
      ctx := m.DCtx.PublishAdminContent
      m.DCtx.PublishAdminContent = nil
      return m, tea.Batch(
          OverlayClearCmd(), FocusSetCmd(PAGEFOCUS), LoadingStartCmd(),
          func() tea.Msg { return ConfirmedUnpublishAdminContentMsg{ctx.AdminContentID, ctx.AdminRouteID} },
      )
  }

case DIALOGRESTOREADMINVERSION:
  if m.DCtx.RestoreAdminVersion != nil {
      ctx := m.DCtx.RestoreAdminVersion
      m.DCtx.RestoreAdminVersion = nil
      return m, tea.Batch(
          OverlayClearCmd(), FocusSetCmd(PAGEFOCUS), LoadingStartCmd(),
          func() tea.Msg { return ConfirmedRestoreAdminVersionMsg{ctx.AdminContentID, ctx.VersionID, ctx.AdminRouteID} },
      )
  }

case DIALOGDELETEADMINCONTENTFIELD:
  if m.DCtx.DeleteAdminContentField != nil {
      ctx := m.DCtx.DeleteAdminContentField
      m.DCtx.DeleteAdminContentField = nil
      return m, tea.Batch(
          OverlayClearCmd(), FocusSetCmd(PAGEFOCUS), LoadingStartCmd(),
          func() tea.Msg { return ConfirmedDeleteAdminContentFieldMsg{ctx.AdminContentFieldID, ctx.AdminContentID, ctx.AdminRouteID, ctx.AdminDatatypeID} },
      )
  }

Part F-5: Form Dialog Submission Handlers

File: internal/tui/admin_form_dialog.go (new)

This file handles FormDialogAction submissions for admin content form dialogs.
Each handler extracts form field values and dispatches the appropriate request message.

FORMDIALOGCREATEADMINCONTENT:
  Flow: User fills title + field values → extract from form → emit CreateAdminContentFromDialogRequestMsg
  Fields: Title (text input), one text input per field in the datatype's field list
  The form is pre-populated with the datatype's fields via AdminBuildContentFormMsg.
  Submission: extract title + field values, dispatch to HandleCreateAdminContentFromDialog.

FORMDIALOGEDITADMINCONTENT:
  Flow: Existing field values pre-loaded → user edits → emit UpdateAdminContentFromDialogRequestMsg
  Fields: Title (text input, pre-filled), one text input per existing field (pre-filled with current values)
  The form is built from ShowEditAdminContentFormDialogMsg which carries ExistingAdminContentField[].
  Submission: extract updated values, dispatch to HandleUpdateAdminContentFromDialog.

FORMDIALOGCHILDADMINDATATYPE:
  Flow: User picks which child datatype to create → emit AdminBuildContentFormMsg
  Fields: Single select list of available child datatypes (from AdminRootDatatypes or filtered children)
  Submission: selected datatype ID → build content form for that datatype.

FORMDIALOGMOVEADMINCONTENT:
  Flow: User picks target node from tree → emit AdminMoveContentRequestMsg
  Fields: Single select list of valid move targets (from MoveAdminContentContext + targets)
  Submission: selected target ID → dispatch to HandleMoveAdminContent.

FORMDIALOGADDADMINCONTENTFIELD:
  Flow: User picks field to add → fills value → emit AddAdminContentFieldRequestMsg
  Fields: Field picker (select from available fields not yet on this content), value input
  Context from: m.DCtx.AddAdminContentField
  Submission: dispatch to HandleAddAdminContentField.

FORMDIALOGEDITADMINSINGLEFIELD:
  Flow: User edits single field value inline → emit UpdateAdminContentFieldRequestMsg
  Fields: Single text/textarea input pre-filled with current value
  Context from: m.DCtx.EditAdminSingleField
  Submission: dispatch to HandleUpdateAdminContentField (reuses HandleUpdateAdminContentFromDialog logic
  for a single field).

---

Part G: Message Wiring

File: internal/tui/admin_update_cms.go (extend existing)

Wire the following message types into the admin Update switch:

Request messages → dispatch to handlers:
  case ConfirmedDeleteAdminContentMsg:    return m, m.HandleDeleteAdminContent(msg)
  case AdminReorderSiblingRequestMsg:     return m, m.HandleAdminReorderSibling(msg)
  case AdminCopyContentRequestMsg:        return m, m.HandleCopyAdminContent(msg)
  case AdminMoveContentRequestMsg:        return m, m.HandleMoveAdminContent(msg)
  case ConfirmedPublishAdminContentMsg:   return m, m.HandleAdminConfirmedPublish(msg)
  case ConfirmedUnpublishAdminContentMsg: return m, m.HandleAdminConfirmedUnpublish(msg)
  case AdminListVersionsRequestMsg:       return m, m.HandleAdminListVersions(msg)
  case ConfirmedRestoreAdminVersionMsg:   return m, m.HandleAdminConfirmedRestoreVersion(msg)
  case ConfirmedDeleteAdminContentFieldMsg: return m, m.HandleDeleteAdminContentField(msg)

Result messages → update model + re-fetch:
  case AdminTreeLoadedMsg:
      newModel := m
      newModel.AdminContentTree = msg.RootNode
      return newModel, LoadingStopCmd()

  case AdminContentCreatedMsg:
      return m, tea.Batch(LoadingStopCmd(), LogMessageCmd(...), ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID))

  case AdminContentDeletedMsg:
      newModel := m; newModel.Cursor = 0
      return newModel, tea.Batch(LoadingStopCmd(), LogMessageCmd(...), ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID))

  case AdminContentUpdatedFromDialogMsg:
      return m, tea.Batch(LoadingStopCmd(), LogMessageCmd(...), ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID))

  case AdminContentReorderedMsg:
      return m, tea.Batch(LoadingStopCmd(), ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID))

  case AdminContentCopiedMsg:
      return m, tea.Batch(LoadingStopCmd(), LogMessageCmd(...), ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID))

  case AdminContentMovedMsg:
      return m, tea.Batch(LoadingStopCmd(), LogMessageCmd(...), ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID))

  case AdminPublishCompletedMsg:
      return m, tea.Batch(LoadingStopCmd(), LogMessageCmd(...), ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID))

  case AdminUnpublishCompletedMsg:
      return m, tea.Batch(LoadingStopCmd(), LogMessageCmd(...), ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID))

  case AdminVersionsListedMsg:
      newModel := m; newModel.AdminContentVersions = msg.Versions
      return newModel, LoadingStopCmd()

  case AdminVersionRestoredMsg:
      return m, tea.Batch(LoadingStopCmd(), LogMessageCmd(...), ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID))

  case AdminContentFieldAddedMsg:
      return m, tea.Batch(LoadingStopCmd(), LoadAdminContentFieldsCmd(m.Config, msg.AdminContentID, ..., m.ActiveLocale))

  case AdminContentFieldDeletedMsg:
      return m, tea.Batch(LoadingStopCmd(), LoadAdminContentFieldsCmd(m.Config, msg.AdminContentID, ..., m.ActiveLocale))

  case AdminContentFieldUpdatedMsg:
      return m, tea.Batch(LoadingStopCmd(), LoadAdminContentFieldsCmd(m.Config, msg.AdminContentID, ..., m.ActiveLocale))

  case AdminLoadContentFieldsMsg:
      newModel := m; newModel.AdminContentFields = msg.Fields
      return newModel, nil

  case AdminRootDatatypesFetchResultsMsg:
      newModel := m; newModel.AdminRootDatatypes = msg.RootDatatypes
      return newModel, nil

Update existing handler:
  Remove: case DeleteAdminContentRequestMsg: return m, m.HandleDeleteAdminContent(msg)
  (Replaced by ConfirmedDeleteAdminContentMsg dispatch above. The old DeleteAdminContentRequestMsg
  bypassed the confirmation dialog.)

---

Steps

1. Define all new message types in msg_admin.go (Part A)
2. Add ExistingAdminContentField struct to msg_admin.go (Part A)
3. Update existing AdminContentCreatedMsg, AdminContentDeletedMsg, DeleteAdminContentRequestMsg
   in msg_crud.go to add AdminRouteID (Part A-1)
4. Update AdminContentFieldDisplay in msg_crud.go with ValidationJSON, DataJSON (Part A)
5. Fix all call sites broken by Part A-1 changes (admin_constructors.go, admin_update_cms.go)
6. Add DialogContext fields and context structs (Part B)
7. Add AdminRootDatatypes field to Model if not present (Part C)
8. Add AdminRootDatatypes fetch message, Cmd constructor, and Go-side filtering (Part C)
9. Create async Cmd constructors in admin_constructors.go with locale parameters (Part D)
10. Create dialog-triggering Cmd constructors in admin_constructors.go (Part D)
11. Add DialogAction constants to dialog.go, update ToggleControls case list (Part F-1)
12. Add FormDialogAction constants to form_dialog.go (Part F-2)
13. Delete existing HandleDeleteAdminContent from admin_update_dialog.go:884-913
14. Create admin_commands.go with all 12 ported handlers (Part E)
15. Add dialog show handlers to admin_update_dialog.go (Part F-3)
16. Add dialog accept handlers to admin_update_dialog.go (Part F-4)
17. Create admin_form_dialog.go with form submission handlers (Part F-5)
18. Wire all message handling in admin_update_cms.go (Part G)
19. Verify: just check

---

Size estimate: ~1400-1700 lines new code across 5-6 files.

---

Risk assessment

Medium risk. The handlers are mechanical ports, but the volume is high. Key risks:

1. Missing an audit context on a mutation handler (silent data changes with no trail)
   Mitigation: After porting each handler, diff against the original to verify all DB calls
   and audit contexts are preserved. The checklist in Part E marks each handler's audit status.

2. Admin field type conversion errors in form dialog population
   Mitigation: ExistingAdminContentField carries all needed fields including ValidationJSON
   and DataJSON. Type mismatches are caught at compile time by the typed ID system.

3. Tree pointer rewiring bugs in move/reorder/delete (same logic, different types)
   Mitigation: These operations are NOT transactional. A failure midway leaves the tree
   inconsistent. This is a known limitation inherited from the regular content handlers.
   The type system prevents cross-contamination (admin vs regular IDs), but logic bugs in
   pointer manipulation would need to be caught by testing.

4. Locale threading — async Cmds that query locale-specific data must receive locale explicitly
   Mitigation: All locale-dependent Cmds now take locale as an explicit parameter (Part D).
   Call sites pass m.ActiveLocale. This matches the regular content pattern.

5. Publishing parameter mismatch — PublishAdminContent takes 8 parameters
   Mitigation: All parameter sources are now documented in Part E handlers 7-8.
   Parameters come from Model fields (ActiveLocale, UserID, Dispatcher) and Config
   (VersionRetentionCap).

Testing

Unit tests should cover the tree-modifying handlers at minimum. Create:

File: internal/tui/admin_commands_test.go

Test cases for HandleDeleteAdminContent:
  - Delete leaf node (no children): verify sibling pointers updated, content deleted
  - Delete node with children: verify ActionResultMsg returned, no deletion
  - Delete first child: verify parent's first_child_id updated to next sibling
  - Delete middle sibling: verify prev.next → our.next, next.prev → our.prev
  - Delete only child: verify parent's first_child_id cleared

Test cases for HandleAdminReorderSibling:
  - Move up: verify sibling swap
  - Move down: verify sibling swap
  - Move up when first: verify no-op or error
  - Move down when last: verify no-op or error

Test cases for HandleMoveAdminContent:
  - Move to different parent: verify detach + attach pointer updates
  - Move to same parent (no-op): verify graceful handling

These tests use mock DbDriver implementations to verify the sequence and parameters of
DB calls without requiring a real database. Pattern: create a mock that records calls,
execute handler, assert call sequence and parameters.

Test cases for HandleAdminConfirmedPublish / HandleAdminConfirmedUnpublish:
  - Verify correct parameters passed to publishing.PublishAdminContent / UnpublishAdminContent
  - Verify audit context created with correct userID

Integration testing (requires database, deferred to Plan 4):
  - Full round-trip: create → edit → publish → unpublish → version restore → delete
