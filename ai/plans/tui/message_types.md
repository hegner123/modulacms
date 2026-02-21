Ready to code?

 Here is Claude's plan:
╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
 Plan: Consolidate TUI Message Types (Conservative)

 Context

 The TUI defines 200+ message types scattered across 13+ files (message_types.go, admin_message_types.go, form_dialog.go, update_dialog.go, dialog.go, commands.go, actions.go,
 uiconfig_form_dialog.go, database_form_dialog.go, etc.). Many are structurally identical one-line structs differing only by name - particularly boolean toggle pairs, direction pairs, and
 same-payload plugin messages.

 Why generics won't work: Go does not support type switching on generic instantiations (case FetchResultMsg[[]db.Routes] is a compile error). Since the entire Bubbletea update chain relies
 on msg.(type) switches, FetchResultMsg[T any] is not viable.

 Goal: Reduce type count by ~25 through consolidating structurally identical groups, introduce small enums for discriminated dispatch, and reorganize message files by domain. No behavioral
 changes.

 Consolidations

 1. Boolean Toggle Pairs → Single Message with Bool (save 2 types)

 ┌───────────────────────────┬─────────────────────────────┐
 │          Remove           │        Replace With         │
 ├───────────────────────────┼─────────────────────────────┤
 │ LoadingTrue, LoadingFalse │ SetLoadingMsg{Loading bool} │
 ├───────────────────────────┼─────────────────────────────┤
 │ ReadyTrue, ReadyFalse     │ SetReadyMsg{Ready bool}     │
 └───────────────────────────┴─────────────────────────────┘

 Files touched: message_types.go, constructors.go, update_state.go

 Constructors become:
 - LoadingStartCmd() → returns SetLoadingMsg{Loading: true}
 - LoadingStopCmd() → returns SetLoadingMsg{Loading: false}
 - Same pattern for Ready

 2. Cursor Messages → Single Message with Action Enum (save 3 types)

 ┌──────────────────────────────────────────────┬───────────────────────────────────────────┐
 │                    Remove                    │               Replace With                │
 ├──────────────────────────────────────────────┼───────────────────────────────────────────┤
 │ CursorUp, CursorDown, CursorReset, CursorSet │ CursorMsg{Action CursorAction, Index int} │
 └──────────────────────────────────────────────┴───────────────────────────────────────────┘

 New enum in message_types.go:
 type CursorAction int
 const (
     CursorMoveUp CursorAction = iota
     CursorMoveDown
     CursorMoveReset
     CursorMoveSet
 )

 Files touched: message_types.go, constructors.go, update_state.go

 Handler in update_state.go becomes:
 case CursorMsg:
     newModel := m
     switch msg.Action {
     case CursorMoveUp:
         newModel.Cursor = m.Cursor - 1
     case CursorMoveDown:
         newModel.Cursor = m.Cursor + 1
     case CursorMoveReset:
         newModel.Cursor = 0
     case CursorMoveSet:
         newModel.Cursor = msg.Index
     }
     return newModel, NewStateUpdate()

 3. Direction Pairs → Single Message with Bool (save 2 types)

 ┌──────────────────────────────────┬────────────────────────────┐
 │              Remove              │        Replace With        │
 ├──────────────────────────────────┼────────────────────────────┤
 │ PageModNext, PageModPrevious     │ PageModMsg{Forward bool}   │
 ├──────────────────────────────────┼────────────────────────────┤
 │ TitleFontNext, TitleFontPrevious │ TitleFontMsg{Forward bool} │
 └──────────────────────────────────┴────────────────────────────┘

 Files touched: message_types.go, constructors.go, update_state.go

 4. Plugin Request Messages → Single Tagged Message (save 4 types)

 Remove: PluginEnableRequestMsg, PluginDisableRequestMsg, PluginReloadRequestMsg, PluginApproveAllRoutesRequestMsg, PluginApproveAllHooksRequestMsg
 Replace With: PluginActionRequestMsg{Name string, Action PluginAction}

 New enum in message_types.go:
 type PluginAction int
 const (
     PluginActionEnable PluginAction = iota
     PluginActionDisable
     PluginActionReload
     PluginActionApproveRoutes
     PluginActionApproveHooks
 )

 Files touched: message_types.go, constructors.go, update_cms.go (where plugin request cases are handled)

 5. Plugin Simple Result Messages → Single Tagged Message (save 2 types)

 ┌────────────────────────────────────────────────────────┬───────────────────────────────────────────────────────────┐
 │                         Remove                         │                       Replace With                        │
 ├────────────────────────────────────────────────────────┼───────────────────────────────────────────────────────────┤
 │ PluginEnabledMsg, PluginDisabledMsg, PluginReloadedMsg │ PluginActionCompleteMsg{Name string, Action PluginAction} │
 └────────────────────────────────────────────────────────┴───────────────────────────────────────────────────────────┘

 Keep PluginRoutesApprovedMsg and PluginHooksApprovedMsg as-is (they carry an additional Count field).

 Files touched: message_types.go, constructors.go, update_cms.go

 6. Redundant Error Types → Audit and Deduplicate (save ~3 types)

 Current error-only types: ErrorSet{Err}, DbErrMsg{Error}, DataFetchErrorMsg{Error}, DatatypeUpdateFailedMsg{Error}, FetchErrMsg{Error}

 Step 1: Check usage of DbErrMsg and DataFetchErrorMsg - if they are only returned but never handled (no case in any switch), they are dead code to remove.

 Step 2: For remaining types, consolidate those handled by the same handler into one type with context:
 type AppErrorMsg struct {
     Context string // "fetch", "update", "db", etc.
     Err     error
 }

 Keep ErrorSet separate (it sets model state in UpdateState) and FetchErrMsg separate (it triggers dialog in UpdateFetch). Only merge types that share the same handler path.

 Files touched: message_types.go, commands.go, relevant update_*.go files

 7. Column Info Pair → Single Message (save 1 type)

 ┌────────────────────────────┬─────────────────────────────────────────────────────────────────────┐
 │           Remove           │                            Replace With                             │
 ├────────────────────────────┼─────────────────────────────────────────────────────────────────────┤
 │ ColumnsSet, ColumnTypesSet │ ColumnInfoSetMsg{Columns *[]string, ColumnTypes *[]*sql.ColumnType} │
 └────────────────────────────┴─────────────────────────────────────────────────────────────────────┘

 These are always set together (see ColumnsFetched handler in update_fetch.go:93). The separate messages are unnecessary.

 Files touched: message_types.go, constructors.go, update_state.go, update_fetch.go

 File Reorganization

 After consolidations, reorganize by moving all message type definitions out of implementation files and into domain-grouped files. Messages currently scattered across form_dialog.go (32
 types), update_dialog.go (28 types), dialog.go (5 types), commands.go (5 types), etc. should be colocated by domain:

 ┌─────────────────┬────────────────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────┐
 │    New File     │                          Content                           │                              Types Moved From                               │
 ├─────────────────┼────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
 │ msg_state.go    │ Cursor, loading, ready, focus, pagination, display setters │ message_types.go                                                            │
 ├─────────────────┼────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
 │ msg_fetch.go    │ Fetch request/result messages + fetch error                │ message_types.go, commands.go                                               │
 ├─────────────────┼────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
 │ msg_crud.go     │ Content/Route/Datatype/Field/User CRUD request/result      │ update_dialog.go, form_dialog.go, message_types.go                          │
 ├─────────────────┼────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
 │ msg_dialog.go   │ Dialog show/accept/cancel + form dialog types              │ dialog.go, form_dialog.go, uiconfig_form_dialog.go, database_form_dialog.go │
 ├─────────────────┼────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
 │ msg_admin.go    │ Admin entity messages (routes, datatypes, fields, content) │ admin_message_types.go                                                      │
 ├─────────────────┼────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
 │ msg_plugin.go   │ Plugin action/result messages                              │ message_types.go                                                            │
 ├─────────────────┼────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
 │ msg_nav.go      │ Navigation, history, page messages                         │ message_types.go                                                            │
 ├─────────────────┼────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
 │ msg_form.go     │ Form lifecycle messages                                    │ message_types.go                                                            │
 ├─────────────────┼────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────────┤
 │ msg_database.go │ Generic database CRUD messages                             │ message_types.go                                                            │
 └─────────────────┴────────────────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────────────────────┘

 Delete message_types.go and admin_message_types.go after migration. The types in implementation files (e.g., form_dialog.go) move to the appropriate msg_*.go file.

 Summary

 ┌──────────────────────┬─────────────┐
 │        Change        │ Types Saved │
 ├──────────────────────┼─────────────┤
 │ Boolean toggles      │ 2           │
 ├──────────────────────┼─────────────┤
 │ Cursor consolidation │ 3           │
 ├──────────────────────┼─────────────┤
 │ Direction pairs      │ 2           │
 ├──────────────────────┼─────────────┤
 │ Plugin requests      │ 4           │
 ├──────────────────────┼─────────────┤
 │ Plugin results       │ 2           │
 ├──────────────────────┼─────────────┤
 │ Error dedup          │ ~3          │
 ├──────────────────────┼─────────────┤
 │ Column info pair     │ 1           │
 ├──────────────────────┼─────────────┤
 │ Total                │ ~17         │
 └──────────────────────┴─────────────┘

 Plus file reorganization makes the remaining ~185 types navigable by domain instead of scattered across 13 files.

 Implementation Order

 1. Enums first - Add CursorAction and PluginAction enums to message_types.go
 2. One consolidation at a time - Do each group (boolean, cursor, direction, plugin, error, column) as a separate pass: update type, update constructor, update all handlers, compile-check
 3. File reorganization last - Move types after all consolidations are stable
 4. Each step: just check to compile-verify before moving to next

 Verification

 1. just check after each consolidation pass (compile-check)
 2. just test after all consolidations complete
 3. Grep for removed type names to confirm no stale references
 4. just check after file reorganization

 Critical Files

 - internal/cli/message_types.go - Main message type definitions (780 lines)
 - internal/cli/admin_message_types.go - Admin message types (267 lines)
 - internal/cli/constructors.go - Message constructors (689 lines)
 - internal/cli/admin_constructors.go - Admin constructors (168 lines)
 - internal/cli/update_state.go - State setter handler (300 lines)
 - internal/cli/update_fetch.go - Fetch handler (407 lines)
 - internal/cli/update_cms.go - CMS operation handler (plugin cases)
 - internal/cli/update_dialog.go - Dialog handler (28+ inline message types)
 - internal/cli/form_dialog.go - Form dialog (32+ inline message types)
 - internal/cli/commands.go - Command constructors (5+ inline message types)
