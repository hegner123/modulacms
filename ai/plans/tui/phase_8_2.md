Phase 8.2: Content Screen Migration + Admin Content Feature Parity

 Context

 Phase 8.2 is the final screen migration in the UNIFIED_TUI_PLAN. It moves the Content and Admin Content pages from the legacy PageSpecificMsgHandlers / cmsPanelContent pattern into a
 self-contained ContentScreen implementing the Screen interface.

 The regular content editor (CONTENT) is a full-featured content tree browser with CRUD, field editing, tree operations, publishing, versioning, and i18n. The admin content editor
 (ADMINCONTENT) is a read-only flat list with delete-only capability. This migration must bring the admin editor to feature parity with the regular editor.

 This plan decomposes into 4 sub-plans that should be executed sequentially.

 ---
 Sub-Plan A: DB Layer — Admin Content Tree Queries

 Goal: Add missing SQL queries and DB wrapper methods so admin content supports tree navigation and field-by-content operations.

 Missing queries to add

 ┌──────────────────────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────────────────────┐
 │                    Query                     │                      Purpose                       │                        Reference                         │
 ├──────────────────────────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ GetAdminContentDataDescendants               │ Recursive CTE to load content tree                 │ sql/schema/16_content_data/queries.sql:120-128           │
 ├──────────────────────────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ ListAdminContentFieldsByContentData          │ Fields for a specific admin content item           │ ListContentFieldsByContentData in content_fields queries │
 ├──────────────────────────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ ListAdminContentFieldsWithFieldByContentData │ Fields joined with field definitions by content ID │ ListContentFieldsWithFieldByContentData in db.go:270     │
 ├──────────────────────────────────────────────┼────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ ListAdminContentFieldsByContentDataIDs       │ Batch-load fields for multiple content IDs         │ ListContentFieldsByContentDataIDs in db.go:275           │
 └──────────────────────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────────────────────┘

 Files to modify

 ┌──────────────────────────────────────────────────────┬────────────────────────────────────────────────┐
 │                         File                         │                     Change                     │
 ├──────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ sql/schema/18_admin_content_data/queries.sql         │ Add GetAdminContentDataDescendants CTE         │
 ├──────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ sql/schema/18_admin_content_data/queries_mysql.sql   │ MySQL equivalent                               │
 ├──────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ sql/schema/18_admin_content_data/queries_psql.sql    │ PostgreSQL equivalent                          │
 ├──────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ sql/schema/19_admin_content_fields/queries.sql       │ Add 3 new field queries                        │
 ├──────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ sql/schema/19_admin_content_fields/queries_mysql.sql │ MySQL equivalents                              │
 ├──────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ sql/schema/19_admin_content_fields/queries_psql.sql  │ PostgreSQL equivalents                         │
 ├──────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ internal/db/db.go                                    │ Add 4 new methods to DbDriver interface        │
 ├──────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ internal/db/admin_content_data.go                    │ Add wrapper for GetAdminContentDataDescendants │
 ├──────────────────────────────────────────────────────┼────────────────────────────────────────────────┤
 │ internal/db/admin_content_field.go                   │ Add 3 field wrapper methods                    │
 └──────────────────────────────────────────────────────┴────────────────────────────────────────────────┘

 Steps

 1. Write GetAdminContentDataDescendants for all 3 dialects (copy from 16_content_data CTE, adapt table/column names)
 2. Write 3 field queries for all 3 dialects (copy from 17_content_fields, adapt names)
 3. Run just sqlc to regenerate
 4. Add 4 methods to DbDriver interface
 5. Implement wrappers on Database, MysqlDatabase, PsqlDatabase
 6. Verify: just check

 ---
 Sub-Plan B: TUI Commands & Messages — Admin Content Operations

 Goal: Create admin content equivalents of all regular content TUI commands so the screen can issue them.

 New message types needed (in msg_admin.go or msg_crud.go)

 ┌──────────────────────────────────┬─────────────────────────────────────────────┬─────────────────────────────┐
 │             Message              │                   Purpose                   │     Regular equivalent      │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminBuildContentFormMsg         │ Request form construction for admin content │ BuildContentFormMsg         │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminFetchContentForEditMsg      │ Fetch admin content fields for edit dialog  │ FetchContentForEditMsg      │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminContentCreatedMsg           │ Admin content created successfully          │ ContentCreatedMsg           │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminContentUpdatedFromDialogMsg │ Admin content updated                       │ ContentUpdatedFromDialogMsg │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminTreeLoadedMsg               │ Admin content tree loaded from DB           │ TreeLoadedMsg               │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminLoadContentFieldsMsg        │ Admin content fields loaded                 │ LoadContentFieldsMsg        │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminReorderSiblingRequestMsg    │ Reorder admin sibling                       │ ReorderSiblingRequestMsg    │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminContentReorderedMsg         │ Reorder complete                            │ ContentReorderedMsg         │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminCopyContentRequestMsg       │ Copy admin content                          │ CopyContentRequestMsg       │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminContentCopiedMsg            │ Copy complete                               │ ContentCopiedMsg            │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminMoveContentRequestMsg       │ Move admin content                          │ N/A (uses existing move)    │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminContentMovedMsg             │ Move complete                               │ ContentMovedMsg             │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminTogglePublishRequestMsg     │ Publish/unpublish admin content             │ TogglePublishRequestMsg     │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminPublishCompletedMsg         │ Publish complete                            │ PublishCompletedMsg         │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminUnpublishCompletedMsg       │ Unpublish complete                          │ UnpublishCompletedMsg       │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminListVersionsRequestMsg      │ List admin versions                         │ ListVersionsRequestMsg      │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminVersionsListedMsg           │ Versions loaded                             │ VersionsListedMsg           │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminRestoreVersionRequestMsg    │ Restore admin version                       │ RestoreVersionRequestMsg    │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminVersionRestoredMsg          │ Restore complete                            │ VersionRestoredMsg          │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminContentFieldUpdatedMsg      │ Admin field updated                         │ ContentFieldUpdatedMsg      │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminContentFieldAddedMsg        │ Admin field added                           │ ContentFieldAddedMsg        │
 ├──────────────────────────────────┼─────────────────────────────────────────────┼─────────────────────────────┤
 │ AdminContentFieldDeletedMsg      │ Admin field deleted                         │ ContentFieldDeletedMsg      │
 └──────────────────────────────────┴─────────────────────────────────────────────┴─────────────────────────────┘

 New commands needed (in admin_constructors.go or constructors.go)

 Each message above needs a corresponding Cmd function. Most are direct copies of the regular content commands adapted for admin types:

 - ReloadAdminContentTreeCmd(config, adminRouteID)
 - LoadAdminContentFieldsCmd(config, adminContentID, adminDatatypeID)
 - AdminReorderSiblingCmd(adminContentID, adminRouteID, direction)
 - AdminCopyContentCmd(adminContentID, adminRouteID)
 - AdminTogglePublishCmd(adminContentID, adminRouteID)
 - AdminListVersionsCmd(adminContentID, adminRouteID)
 - FetchAdminContentForEditCmd(adminContentID, adminDatatypeID, adminRouteID, title)
 - ShowAdminChildDatatypeDialogCmd(rootAdminDatatypeID, adminRouteID)
 - ShowCreateAdminRouteWithContentDialogCmd(adminRootDatatypes)
 - ShowDeleteAdminContentDialogCmd(id, name, hasChildren)
 - ShowMoveAdminContentDialogCmd(node, adminRouteID, targets)

 New command handlers (in commands.go or new admin_commands.go)

 Port the following handlers from commands.go for admin content, adapting DB calls to use admin equivalents:

 ┌───────────────────────────────┬──────────────────────┬──────────────────────────────────────────────────────────────────┐
 │            Handler            │ Lines in commands.go │                         Admin adaptation                         │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleCreateContentFromDialog │ 378-489              │ Use CreateAdminContentData, CreateAdminContentField              │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleFetchContentForEdit     │ 492-571              │ Use GetAdminContentData, admin field queries                     │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleUpdateContentFromDialog │ 574-669              │ Use UpdateAdminContentData, UpdateAdminContentField              │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleDeleteContent           │ 672+                 │ Use DeleteAdminContentData (already exists)                      │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleMoveContent             │ 799-1012             │ Use admin update methods for pointer rewiring                    │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleReorderSibling          │ 1772-2050            │ Use admin update methods                                         │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleCopyContent             │ 2057-2236            │ Use admin create + field methods                                 │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleConfirmedPublish        │ 2281-2310            │ Use UpdateAdminContentDataPublishMeta, CreateAdminContentVersion │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleConfirmedUnpublish      │ 2312-2343            │ Use UpdateAdminContentDataPublishMeta                            │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleListVersions            │ 2370-2394            │ Use ListAdminContentVersionsByContent                            │
 ├───────────────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────────┤
 │ HandleConfirmedRestoreVersion │ 2396-2425            │ Use GetAdminContentVersion, admin update methods                 │
 └───────────────────────────────┴──────────────────────┴──────────────────────────────────────────────────────────────────┘

 Files to create/modify

 ┌──────────────────────────────────────────────────────────────────┬──────────────────────────────────────────────────────────────────────┐
 │                               File                               │                                Change                                │
 ├──────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────┤
 │ internal/tui/msg_admin.go                                        │ Add ~20 new admin content message types                              │
 ├──────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────┤
 │ internal/tui/admin_constructors.go                               │ Add ~15 new admin content Cmd functions                              │
 ├──────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────┤
 │ internal/tui/admin_commands.go (new)                             │ Port ~11 command handlers for admin content                          │
 ├──────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────────┤
 │ internal/tui/admin_form_dialog.go (new or extend form_dialog.go) │ Admin content form dialogs (create, edit, child picker, move picker) │
 └──────────────────────────────────────────────────────────────────┴──────────────────────────────────────────────────────────────────────┘

 Steps

 1. Define all message types in msg_admin.go
 2. Create Cmd functions in admin_constructors.go
 3. Create admin_commands.go with ported handlers
 4. Create admin content form dialogs (adapt from regular content dialogs)
 5. Wire admin message handling in update_cms.go / admin_update_cms.go
 6. Verify: just check

 ---
 Sub-Plan C: Screen Migration — ContentScreen Implementation

 Goal: Create screen_content.go implementing Screen for both CONTENT and ADMINCONTENT with AdminMode bool.

 ContentScreen struct

 type ContentScreen struct {
     AdminMode bool

     // Cursors and focus
     Cursor      int
     FieldCursor int
     PanelFocus  FocusPanel

     // Route list phase (both modes)
     RootContentSummary      []db.ContentDataTopLevel      // regular mode
     AdminRootContentSummary []db.AdminContentDataTopLevel  // admin mode
     RootDatatypes           []db.Datatypes                 // regular mode root datatypes
     AdminRootDatatypes      []db.AdminDatatypes            // admin mode root datatypes

     // Tree browsing phase (both modes)
     PageRouteId            types.RouteID      // regular
     AdminPageRouteId       types.AdminRouteID // admin
     Root                   tree.Root          // shared tree structure
     SelectedDatatype       types.DatatypeID
     AdminSelectedDatatype  types.AdminDatatypeID
     SelectedContentFields  []ContentFieldDisplay      // regular
     AdminSelectedFields    []AdminContentFieldDisplay  // admin
     PendingCursorContentID types.ContentID

     // Version state
     ShowVersionList  bool
     Versions         []db.ContentVersion      // or AdminContentVersion
     AdminVersions    []db.AdminContentVersion
     VersionCursor    int
     VersionContentID types.ContentID
     VersionRouteID   types.RouteID
 }

 AppContext extension

 Add ActiveLocale string and AccordionEnabled bool to AppContext struct in app_context.go.

 Screen methods

 PageIndex() — Returns CONTENT or ADMINCONTENT based on AdminMode.

 View(ctx AppContext) — Determines layout based on state:
 - Admin + route list: flat list | detail | empty
 - Admin + tree: tree | preview | fields
 - Regular + route list: route list | detail | actions
 - Regular + tree: tree | preview | fields/versions

 Panel titles are dynamic based on state (matching current cmsPanelTitles logic).
 Context controls (status bar hints) are generated based on state and panel focus.

 Update(ctx AppContext, msg tea.Msg) — Handles:
 1. tea.KeyMsg — Dispatches to helper methods based on AdminMode and state phase
 2. Data refresh messages — Updates screen-owned slices
 3. Operation result messages — Updates state after CRUD/tree operations

 Helper methods for Update

 ┌───────────────────────┬──────────────────────────────────────────────────────────────────────────────┐
 │        Method         │                                   Handles                                    │
 ├───────────────────────┼──────────────────────────────────────────────────────────────────────────────┤
 │ handleRouteListKeys   │ Route list navigation, select, new, edit                                     │
 ├───────────────────────┼──────────────────────────────────────────────────────────────────────────────┤
 │ handleTreeKeys        │ Tree expand/collapse, navigate, CRUD, move, copy, reorder, publish, versions │
 ├───────────────────────┼──────────────────────────────────────────────────────────────────────────────┤
 │ handleFieldPanelKeys  │ Field edit/add/delete/reorder in right panel                                 │
 ├───────────────────────┼──────────────────────────────────────────────────────────────────────────────┤
 │ handleVersionListKeys │ Version navigation and restore confirmation                                  │
 ├───────────────────────┼──────────────────────────────────────────────────────────────────────────────┤
 │ handleBackKey         │ Layered dismiss: version list → tree → route list → history                  │
 └───────────────────────┴──────────────────────────────────────────────────────────────────────────────┘

 Each method branches on AdminMode to use the appropriate commands and data.

 Rendering methods (moved from panel_view.go and page_builders.go)

 ┌──────────────────────┬──────────────────────────────────────────────────┐
 │        Method        │                      Source                      │
 ├──────────────────────┼──────────────────────────────────────────────────┤
 │ renderRouteList      │ renderRootContentSummaryList (panel_view.go)     │
 ├──────────────────────┼──────────────────────────────────────────────────┤
 │ renderRouteDetail    │ renderContentSummaryDetail (panel_view.go)       │
 ├──────────────────────┼──────────────────────────────────────────────────┤
 │ renderRouteActions   │ renderContentActions (panel_view.go)             │
 ├──────────────────────┼──────────────────────────────────────────────────┤
 │ renderTree           │ CMSPage.ProcessTreeDatatypes (page_builders.go)  │
 ├──────────────────────┼──────────────────────────────────────────────────┤
 │ renderContentPreview │ CMSPage.ProcessContentPreview (page_builders.go) │
 ├──────────────────────┼──────────────────────────────────────────────────┤
 │ renderFields         │ CMSPage.ProcessFields (page_builders.go)         │
 ├──────────────────────┼──────────────────────────────────────────────────┤
 │ renderVersionList    │ renderVersionList (panel_view.go)                │
 ├──────────────────────┼──────────────────────────────────────────────────┤
 │ renderAdminList      │ renderAdminContentList (admin_panel_view.go)     │
 ├──────────────────────┼──────────────────────────────────────────────────┤
 │ renderAdminDetail    │ renderAdminContentDetail (admin_panel_view.go)   │
 └──────────────────────┴──────────────────────────────────────────────────┘

 Tree traversal helpers (moved from page_builders.go)

 traverseTree, traverseTreeWithDepth, FormatTreeRow, DecideNodeName — move to screen_content.go as methods or package-level functions. DecideNodeName is also called from
 update_controls.go:749 — keep it package-level.

 Coexistence protocol (update.go)

 Add cases to the ActiveScreen dispatch block for content-specific messages. The screen handles data refresh messages directly; operation result messages (from UpdateCms) are forwarded to
 the screen after root Model state is updated.

 Files to create/modify

 ┌──────────────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────┐
 │                 File                 │                                      Change                                      │
 ├──────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tui/screen_content.go (new) │ ~800-1000 lines: struct, constructor, Update, View, render methods, tree helpers │
 ├──────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tui/screen.go               │ Add CONTENT/ADMINCONTENT to screenForPage                                        │
 ├──────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tui/app_context.go          │ Add ActiveLocale, AccordionEnabled fields                                        │
 ├──────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tui/update.go               │ Add ~20 message cases to ActiveScreen block                                      │
 ├──────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ internal/tui/model.go                │ Add AdminRootDatatypes field if not present                                      │
 └──────────────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────┘

 Steps

 1. Extend AppContext
 2. Create ContentScreen struct and NewContentScreen constructor
 3. Implement PageIndex()
 4. Implement View() by porting all render functions
 5. Implement Update() keyboard handling for regular mode (port ContentBrowserControls)
 6. Implement Update() keyboard handling for admin mode (port + extend AdminContentBrowserControls)
 7. Implement Update() message handling for data refresh
 8. Wire into screenForPage
 9. Add coexistence protocol entries in update.go
 10. Verify: just check, manual walkthrough

 ---
 Sub-Plan D: Legacy Cleanup — Remove Migrated Code

 Goal: Remove all content-related code from legacy handlers and rendering.

 Removals from update_controls.go

 - case CONTENT: and case ADMINCONTENT: from PageSpecificMsgHandlers switch (lines 66-69)
 - ContentBrowserControls method (~370 lines, 517-886)
 - contentBrowserCursorUpCmd helper (888-896)
 - contentBrowserCursorDownCmd helper (898-906)
 - contentFieldEdit method (908+)
 - contentFieldAdd method
 - contentFieldDelete method
 - contentFieldReorderUp method
 - contentFieldReorderDown method

 Removals from admin_controls.go

 - AdminContentBrowserControls method (~50 lines, 450-500)

 Removals from panel_view.go

 - case CONTENT: and case ADMINCONTENT: from cmsPanelContent (lines 344-365)
 - case CONTENT: from cmsPanelTitles (lines 313-320)
 - case CONTENT: and case ADMINCONTENT: from getContextControls (lines 595-678)
 - renderRootContentSummaryList function
 - renderContentSummaryDetail function
 - renderContentActions function
 - renderVersionList function

 Removals from admin_panel_view.go

 - renderAdminContentList function (268-289)
 - renderAdminContentDetail function (291-316)

 Removals from page_builders.go

 - CMSPage struct and all methods (ProcessTreeDatatypes, traverseTree, traverseTreeWithDepth, FormatTreeRow, ProcessContentPreview, ProcessFields)
 - Keep DecideNodeName as package-level if used elsewhere, or move to screen_content.go

 Post-cleanup verification

 After removal, PageSpecificMsgHandlers should only contain:
 case HOMEPAGE:
     return m.BasicControls(msg)
 case CMSPAGE:
     return m.BasicCMSControls(msg)
 case ADMINCMSPAGE:
     return m.BasicCMSControls(msg)

 Steps

 1. Remove legacy control handlers
 2. Remove legacy rendering functions
 3. Remove page_builders.go content code
 4. Verify: just check, just test
 5. Manual walkthrough of all content operations

 ---
 Execution Order

 Sub-Plan A (DB Layer)
     |
     v
 Sub-Plan B (Commands & Messages)
     |
     v
 Sub-Plan C (Screen Implementation)
     |
     v
 Sub-Plan D (Legacy Cleanup)

 Each sub-plan is independently verifiable with just check. Sub-Plans A and B are prerequisites for C. Sub-Plan D depends on C being complete and tested.

 ---
 Verification

 After all 4 sub-plans:

 1. just check — compilation
 2. just test — existing tests pass
 3. Manual TUI walkthrough — both CONTENT and ADMINCONTENT pages:
   - Route list navigation, search/select
   - Tree browsing: expand, collapse, navigate parent/child
   - Content CRUD: create, edit, delete (with children check)
   - Field operations: edit, add, delete, reorder
   - Tree operations: move, copy, reorder siblings
   - Publishing/unpublishing
   - Version history, restore
   - Locale switching (i18n)
   - Back navigation through all state layers
 4. Verify PageSpecificMsgHandlers only has HOMEPAGE/CMSPAGE/ADMINCMSPAGE
 5. Verify cmsPanelContent has no CONTENT/ADMINCONTENT cases
