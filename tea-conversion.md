# Functions to Convert to tea.Cmd and tea.Msg

## effects.go
- `GetTablesCMD()` (already a tea.Cmd)
- `GetColumns()` (already a tea.Cmd)
- `GetColumnsRows()` (needs conversion to tea.Cmd)
- `GetSuggestionsString()` (needs conversion to tea.Cmd)

## update.go
- All Update* functions need to properly return tea.Cmd:
  - `UpdateTableSelect()`
  - `UpdatePageSelect()`
  - `UpdateDatabaseCreate()`
  - `UpdateDatabaseRead()`
  - `UpdateDatabaseReadSingle()`
  - `UpdateDatabaseUpdate()`
  - `UpdateDatabaseFormUpdate()`
  - `UpdateDatabaseDelete()`
  - `UpdateContent()`
  - `UpdateConfig()`

## model.go
- `InitialModel()` (already returns tea.Cmd)
- `GetIDRow()` (needs conversion to tea.Cmd)
- `ParseTitleFonts()` (needs conversion to tea.Cmd)
- `LoadTitles()` (needs conversion to tea.Cmd)
- `GetStatus()` (needs conversion to tea.Cmd)

## form.go
- `BuildCreateDBForm()` (needs proper tea.Cmd implementation)
- `BuildUpdateDBForm()` (needs proper tea.Cmd implementation)
- `BuildCMSForm()` (needs proper tea.Cmd implementation)

## controls.go
- Control functions that handle user input

## router.go
- Routing logic that changes application state

## controls_v2.go
- New control functions that need tea.Cmd implementations

## inputs_file.go
- File input handling functions

## cms_controls.go
- CMS-specific control handlers

## inputs.go
- Input interface implementation

## install_form.go
- Installation form tea.Cmd implementations

## Message Types Needed
- `errMsg` (already defined)
- `tableFetchedMsg` (already defined)
- `columnFetchedMsg` (already defined)
- `formCompletedMsg` (already defined)
- `formCancelledMsg` (already defined)

### Additional Message Types to Create
- Navigation/routing messages
- State update messages
- UI update messages
- Database operation result messages