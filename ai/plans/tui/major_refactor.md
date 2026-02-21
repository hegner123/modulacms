───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 Ready to code?

 Here is Claude's plan:
╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
 TUI Component Isolation Refactor

 Context

 The TUI (internal/cli/) is a ~20,000 LOC Bubbletea app built during learning. It has a single monolithic Model with 68+ fields holding ALL state for 46 pages, ~292 message types scattered
 across 13+ files, 12 chained sub-handlers using early-return dispatch, and zero test coverage. The goal is to introduce component isolation: each major screen becomes a self-contained
 struct with its own Update/View, the root Model shrinks to global state + active screen, and messages become domain-scoped.

 Phase 0: Prerequisites

 0.1 Move package-level context vars into Model

 19 package-level var pointers in update_dialog.go and admin_update_dialog.go store state between "show dialog" and "accept dialog" messages. These are unsafe for concurrent SSH sessions.

 Add to model.go:
 type DialogContext struct {
     DeleteContent       *DeleteContentContext
     DeleteDatatype      *DeleteDatatypeContext
     DeleteField         *DeleteFieldContext
     DeleteRoute         *DeleteRouteContext
     DeleteMedia         *DeleteMediaContext
     DeleteUser          *DeleteUserContext
     DeleteAdminRoute    *DeleteAdminRouteContext
     DeleteAdminDatatype *DeleteAdminDatatypeContext
     DeleteAdminField    *DeleteAdminFieldContext
     DeleteContentField  *DeleteContentFieldContext
     EditSingleField     *editSingleFieldCtx
     AddContentField     *addContentFieldCtx
     InitRouteContent    *InitializeRouteContentContext
     RestoreBackup       *RestoreBackupContext
     ApprovePlugin       *ApprovePluginContext
     RestoreRequiresQuit bool
 }

 Add DCtx DialogContext field to Model. Change every deleteContentContext = &DeleteContentContext{...} to m.DCtx.DeleteContent = &DeleteContentContext{...}, and similarly for reads and
 clears.

 Files: model.go, update_dialog.go, admin_update_dialog.go
 Verify: just check

 0.2 Consolidate message types (existing plan)

 Execute the consolidation plan at ai/plans/tui/message_types.md with the fixes identified in the skeptic review:
 - Resolve ReadyTrue/ReadyFalse (dead code audit, not consolidation)
 - Skip the error dedup (leave error types as-is)
 - Do boolean toggles, cursor, direction pairs, plugin request/result consolidations
 - Do file reorganization into domain-grouped msg_*.go files

 Files: message_types.go, admin_message_types.go, constructors.go, admin_constructors.go, update_state.go, update_cms.go, plus new msg_*.go files
 Verify: just check, just test

 ---
 Phase 1: Unify the Dialog System

 Replace 6 dialog pointer+bool pairs with a single interface.

 1.1 Define the Overlay interface

 New file overlay.go:
 type Overlay interface {
     Update(msg tea.KeyMsg) (Overlay, tea.Cmd)
     View(width, height int) string
 }

 1.2 Adapt each dialog type to satisfy Overlay

 For each of DialogModel, FormDialogModel, ContentFormDialogModel, UserFormDialogModel, DatabaseFormDialogModel, UIConfigFormDialogModel:
 - Ensure Update accepts tea.KeyMsg, returns (Overlay, tea.Cmd)
 - Ensure View/Render accepts (width, height int), returns string

 Do one dialog type at a time. Compile-check between each.

 1.3 Replace 6 pointer+bool pairs with single field

 In model.go, replace:
 Dialog *DialogModel / DialogActive bool
 FormDialog *FormDialogModel / FormDialogActive bool
 ContentFormDialog *ContentFormDialogModel / ContentFormDialogActive bool
 UserFormDialog *UserFormDialogModel / UserFormDialogActive bool
 DatabaseFormDialog *DatabaseFormDialogModel / DatabaseFormDialogActive bool
 UIConfigFormDialog *UIConfigFormDialogModel / UIConfigFormDialogActive bool

 With:
 ActiveOverlay Overlay  // nil = no dialog active

 1.4 Collapse overlay handling in update.go

 The 6 inline dialog-capture blocks (lines 113-165) become:
 if m.ActiveOverlay != nil {
     if keyMsg, ok := msg.(tea.KeyMsg); ok {
         overlay, cmd := m.ActiveOverlay.Update(keyMsg)
         m.ActiveOverlay = overlay
         return m, cmd
     }
 }

 1.5 Collapse overlay rendering in view.go and panel_view.go

 The 6 overlay checks in view.go (lines 218-244) and panel_view.go become one check.

 1.6 Update all dialog open/close code

 Every m.FormDialog = &dialog; m.FormDialogActive = true becomes m.ActiveOverlay = &dialog.
 Every m.FormDialogActive = false; m.FormDialog = nil becomes m.ActiveOverlay = nil.

 Replace the 12 Set/ActiveSet message types (FormDialogSetMsg, FormDialogActiveSetMsg, etc.) with:
 type OverlaySetMsg struct{ Overlay Overlay }
 type OverlayClearMsg struct{}

 Files: model.go, overlay.go (new), update.go, view.go, panel_view.go, update_state.go, update_dialog.go, admin_update_dialog.go, constructors.go, admin_constructors.go, each dialog file
 Verify: just check. Manual test of each dialog type.

 ---
 Phase 2: Screen Interface and AppContext

 2.1 Define AppContext

 New file app_context.go:
 type AppContext struct {
     DB            db.DbDriver
     Config        *config.Config
     Logger        Logger
     UserID        types.UserID
     Width         int
     Height        int
     PluginManager *plugin.Manager
     ConfigManager *config.Manager
 }

 Add func (m Model) AppCtx() AppContext to model.go.

 2.2 Define the Screen interface

 New file screen.go:
 type Screen interface {
     Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd)
     View(ctx AppContext) string
     PageIndex() PageIndex
 }

 2.3 Add ActiveScreen to Model

 ActiveScreen Screen  // nil = legacy path

 2.4 Wire Screen into update chain and view

 In update.go, after overlay handling, before legacy sub-handlers:
 if m.ActiveScreen != nil {
     ctx := m.AppCtx()
     screen, cmd := m.ActiveScreen.Update(ctx, msg)
     m.ActiveScreen = screen
     if cmd != nil {
         return m, cmd
     }
 }

 In view.go, before the page switch:
 if m.ActiveScreen != nil {
     ui := m.ActiveScreen.View(m.AppCtx())
     if m.ActiveOverlay != nil {
         return renderOverlay(ui, m.ActiveOverlay.View(m.Width, m.Height), m.Width, m.Height)
     }
     return ui
 }

 2.5 Add screenForPage dispatch

 In update_navigation.go, when handling NavigateToPage:
 m.ActiveScreen = m.screenForPage(msg.Page)
 // nil = legacy path runs as before

 screenForPage starts returning nil for everything. Screens get wired in as they're built.

 Files: app_context.go (new), screen.go (new), model.go, update.go, view.go, update_navigation.go
 Verify: just check. No behavior change -- ActiveScreen is always nil.

 ---
 Phase 3: Pilot Screen Migrations

 Migrate 3 simple screens to validate the pattern.

 3.1 Plugin Detail screen

 New file screen_plugin_detail.go:
 type PluginDetailScreen struct {
     SelectedPlugin string
     Cursor         int
     CursorMax      int
 }

 Move code from:
 - PluginDetailControls in update_controls.go -> Update()
 - PLUGINDETAILPAGE case in view.go -> View()

 Update screenForPage to return NewPluginDetailScreen(...) for PLUGINDETAILPAGE.
 Remove the PLUGINDETAILPAGE case from PageSpecificMsgHandlers and view.go.

 3.2 Config Category screen

 New file screen_config_category.go:
 type ConfigCategoryScreen struct {
     Category      config.FieldCategory
     Fields        []config.FieldMeta
     FieldCursor   int
 }

 Move from ConfigCategoryControls and renderConfigCategoryPage.

 3.3 Actions screen

 New file screen_actions.go:
 type ActionsScreen struct {
     Cursor    int
     CursorMax int
 }

 Move from ActionsControls and ACTIONSPAGE view case.

 Files per screen: one new screen_*.go. Remove corresponding code from update_controls.go and view.go.
 Verify per screen: just check. Navigate to screen, test all controls, test dialogs, test back nav.

 ---
 Phase 4: Panel Screen Base and CMS Migrations

 4.1 PanelScreen base

 New file screen_panel.go:
 type PanelScreen struct {
     PanelFocus tui.FocusPanel
     Cursor     int
     CursorMax  int
 }

 func (p *PanelScreen) RenderPanels(ctx AppContext, titles [3]string, left, center, right string, statusBar string) string
 func (p *PanelScreen) HandlePanelNav(key string, km config.KeyBindings) (handled bool, cmd tea.Cmd)

 Extract from renderCMSPanelLayout, cmsPanelContent, cmsPanelTitles in panel_view.go.

 4.2 Plugins screen (CMS panel)

 New file screen_plugins.go:
 type PluginsScreen struct {
     PanelScreen
     PluginsList    []PluginDisplay
     SelectedPlugin string
     Loading        bool
 }

 Move from PluginsControls and plugins panel rendering.

 4.3 Routes screen (merged admin+regular)

 New file screen_routes.go:
 type RoutesScreen struct {
     PanelScreen
     Routes    []db.Routes       // regular mode
     AdminRoutes []db.AdminRoutes // admin mode
     AdminMode bool
     Loading   bool
 }

 Move from RoutesControls, AdminRoutesControls, and routes panel rendering.
 When AdminMode is true, fetch/display uses admin types. PageIndex() returns ROUTES or ADMINROUTES.

 4.4 Media screen

 New file screen_media.go:
 type MediaScreen struct {
     PanelScreen
     MediaList []db.Media
     Loading   bool
 }

 File picker stays on root Model for now. Media screen emits OpenFilePickerMsg for the root to handle.

 4.5 Users screen

 New file screen_users.go:
 type UsersScreen struct {
     PanelScreen
     UsersList []db.UserWithRoleLabelRow
     RolesList []db.Roles
     Loading   bool
 }

 4.6 Datatypes screen (merged admin+regular)

 New file screen_datatypes.go:
 type DatatypesScreen struct {
     PanelScreen
     AllDatatypes     []db.Datatypes       // regular
     AdminDatatypes   []db.AdminDatatypes  // admin
     AdminMode        bool
     SelectedFields   []db.Fields          // regular
     AdminFields      []db.AdminFields     // admin
     FieldCursor      int
     Loading          bool
 }

 4.7 Content screen (merged admin+regular, most complex)

 New file screen_content.go:
 type ContentScreen struct {
     PanelScreen
     // Regular mode
     Root                   tree.Root
     RootContentSummary     []db.RootContentSummary
     SelectedContentFields  []ContentFieldDisplay
     RootDatatypes          []db.Datatypes
     // Admin mode
     AdminContentData       []db.AdminContentData
     AdminContentFields     []AdminContentFieldDisplay
     // Shared
     AdminMode              bool
     PageRouteId            types.RouteID
     SelectedDatatype       types.DatatypeID
     FieldCursor            int
     PendingCursorContentID types.ContentID
     UsersList              []db.UserWithRoleLabelRow
     Loading                bool
 }

 This is the largest migration. Move from ContentBrowserControls, AdminContentBrowserControls, content panel rendering, and content-specific messages from update_cms.go /
 admin_update_cms.go.

 Migrate content last after all other panel screens are proven.

 Files: screen_panel.go (new), one new screen_*.go per screen. Remove from update_controls.go, panel_view.go, admin_panel_view.go, update_cms.go, admin_update_cms.go.
 Verify per screen: just check. Full manual walkthrough.

 ---
 Phase 5: Migrate Fetch Logic Into Screens

 Move fetch request/result handling from UpdateFetch/UpdateAdminFetch into each screen's Update.

 For each migrated screen:
 1. The screen's Update handles its fetch request messages (e.g., RoutesFetchMsg) by creating the async DB command using ctx.DB
 2. The screen's Update handles its fetch result messages (e.g., RoutesFetchResultsMsg) by storing the data locally
 3. Remove the corresponding cases from UpdateFetch / UpdateAdminFetch

 After all screens migrate their fetches, UpdateFetch and UpdateAdminFetch should be empty or near-empty. Delete them and remove from the update chain.

 Files: each screen_*.go, update_fetch.go, admin_update_fetch.go, update.go
 Verify: just check after each screen's fetch migration.

 ---
 Phase 6: Slim Down Model

 Remove fields from Model that now live in screen components:

 Fields to remove (grouped):
 - CMS panel: Routes, AllDatatypes, SelectedDatatype, SelectedDatatypeFields, FieldCursor, SelectedContentFields, Root, RootDatatypes, RootContentSummary, MediaList, UsersList, RolesList,
 PanelFocus, PageRouteId, PendingCursorContentID
 - Admin: AdminRoutes, AdminAllDatatypes, AdminSelectedDatatypeFields, AdminRootContentSummary, AdminSelectedContentFields, AdminFieldCursor
 - Plugin: PluginsList, SelectedPlugin
 - Config: ConfigCategory, ConfigCategoryFields, ConfigFieldCursor
 - Legacy nav: remove old cursor fields that screens now own

 Target Model size: ~25-30 fields (infra, terminal, navigation, active components, chrome, provisioning).

 Remove emptied handler files (update_fetch.go, update_admin_fetch.go, update_cms.go, admin_update_cms.go, update_controls.go) as they become empty.

 Files: model.go, delete emptied update_*.go files, update.go (remove dead handler calls)
 Verify: just check, just test, grep for removed field names.

 ---
 Phase 7: Cleanup

 1. Delete message_types.go and admin_message_types.go if all types have moved to msg_*.go or screen files
 2. Delete panel_view.go and admin_panel_view.go if all rendering moved to screen_panel.go and screen files
 3. Delete emptied constructors if all moved to screen files
 4. Update ai/workflows/CREATING_TUI_SCREENS.md to document the Screen interface pattern
 5. Remove dead code flagged by compiler

 Files: various deletions, workflow doc update
 Verify: just check, just test

 ---
 Implementation Order and Parallelism

 Phase 0.1 (context vars) ──┐
 Phase 0.2 (msg consolidate)─┼── Phase 1 (dialog unify) ── Phase 2 (Screen interface)
                              │
                              ├── Phase 3.1 (plugin detail) ─┐
                              ├── Phase 3.2 (config category) ├── Phase 4.1 (PanelScreen)
                              └── Phase 3.3 (actions)        ─┘         │
                                                                        ├── 4.2 plugins
                                                                        ├── 4.3 routes
                                                                        ├── 4.4 media     ─── all parallel
                                                                        ├── 4.5 users
                                                                        ├── 4.6 datatypes
                                                                        └── 4.7 content (last)
                                                                               │
                                                               Phase 5 (fetch migration, per-screen parallel)
                                                                               │
                                                               Phase 6 (slim Model) ── Phase 7 (cleanup)

 Phases 0.1 and 0.2 can run in parallel. Within Phase 3, all 3 screens can be done in parallel. Within Phase 4, screens 4.2-4.6 can be done in parallel after 4.1; content (4.7) should be
 last. Phase 5 fetch migrations can run per-screen in parallel.

 Verification Strategy

 After every step:
 1. just check -- compile verification (primary safety net)
 2. just test -- after each phase completion

 After each screen migration:
 3. Manual walkthrough via SSH TUI: navigate to screen, test all keyboard controls, test dialog open/close/accept/cancel, test back navigation, test data fetching

 After all phases:
 4. Full walkthrough of every page
 5. just test (full test suite)
 6. Grep for any removed type/field names to catch stale references

 Critical Files

 ┌────────────────────┬──────────────────────────────────────┬──────────────────────────────────┐
 │        File        │                 Role                 │          Primary Phase           │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ model.go           │ Root Model struct                    │ 0.1, 1.3, 2.3, 6                 │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ update.go          │ Update dispatch chain                │ 1.4, 2.4                         │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ view.go            │ View dispatch                        │ 1.5, 2.4                         │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ update_dialog.go   │ Dialog handling (2,893 lines)        │ 0.1, 1.6                         │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ update_controls.go │ Page keyboard handlers (1,800 lines) │ 3, 4 (decompose into screens)    │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ panel_view.go      │ 3-panel CMS layout (851 lines)       │ 4.1 (extract to screen_panel.go) │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ update_fetch.go    │ Fetch handlers                       │ 5 (move into screens)            │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ update_cms.go      │ CMS handlers                         │ 4.7 (move into content screen)   │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ overlay.go         │ New: Overlay interface               │ 1.1                              │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ screen.go          │ New: Screen interface                │ 2.2                              │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ app_context.go     │ New: AppContext struct               │ 2.1                              │
 ├────────────────────┼──────────────────────────────────────┼──────────────────────────────────┤
 │ screen_panel.go    │ New: PanelScreen base                │ 4.1                              │
 └────────────────────┴──────────────────────────────────────┴──────────────────────────────────┘
