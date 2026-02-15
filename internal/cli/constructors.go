package cli

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/model"
)

// LogModelCMD creates a command to log the model state with optional filters.
func LogModelCMD(include *[]string, exclude *[]string) tea.Cmd {
	return func() tea.Msg {
		return LogModelMsg{
			Include: include,
			Exclude: exclude,
		}
	}
}

// ClearScreenCmd creates a command to clear the screen.
func ClearScreenCmd() tea.Cmd { return func() tea.Msg { return ClearScreen{} } }

// TitleFontNextCmd creates a command to cycle to the next title font.
func TitleFontNextCmd() tea.Cmd { return func() tea.Msg { return TitleFontNext{} } }

// TitleFontPreviousCmd creates a command to cycle to the previous title font.
func TitleFontPreviousCmd() tea.Cmd { return func() tea.Msg { return TitleFontPrevious{} } }

// LogMessageCmd creates a command to log a message.
func LogMessageCmd(msg string) tea.Cmd { return func() tea.Msg { return LogMsg{Message: msg} } }

// ReadyTrueCmd creates a command to set application ready state to true.
func ReadyTrueCmd() tea.Cmd { return func() tea.Msg { return ReadyTrue{} } }

// ReadyFalseCmd creates a command to set application ready state to false.
func ReadyFalseCmd() tea.Cmd { return func() tea.Msg { return ReadyFalse{} } }

// TablesFetchCmd creates a command to fetch the list of database tables.
func TablesFetchCmd() tea.Cmd { return func() tea.Msg { return TablesFetch{} } }

// TablesSetCmd creates a command to set the tables list.
func TablesSetCmd(tables []string) tea.Cmd {
	return func() tea.Msg { return TablesSet{Tables: tables} }
}

// LoadingStartCmd creates a command to start the loading state.
func LoadingStartCmd() tea.Cmd { return func() tea.Msg { return LoadingTrue{} } }

// LoadingStopCmd creates a command to stop the loading state.
func LoadingStopCmd() tea.Cmd { return func() tea.Msg { return LoadingFalse{} } }

// CursorUpCmd creates a command to move the cursor up.
func CursorUpCmd() tea.Cmd { return func() tea.Msg { return CursorUp{} } }

// CursorDownCmd creates a command to move the cursor down.
func CursorDownCmd() tea.Cmd { return func() tea.Msg { return CursorDown{} } }

// CursorResetCmd creates a command to reset the cursor position.
func CursorResetCmd() tea.Cmd { return func() tea.Msg { return CursorReset{} } }

// CursorSetCmd creates a command to set the cursor to a specific index.
func CursorSetCmd(index int) tea.Cmd { return func() tea.Msg { return CursorSet{Index: index} } }

// CursorMaxSetCmd creates a command to set the maximum cursor value.
func CursorMaxSetCmd(cursorMax int) tea.Cmd {
	return func() tea.Msg { return CursorMaxSet{CursorMax: cursorMax} }
}

// PageModNextCmd creates a command to navigate to the next page.
func PageModNextCmd() tea.Cmd { return func() tea.Msg { return PageModNext{} } }

// PageModPreviousCmd creates a command to navigate to the previous page.
func PageModPreviousCmd() tea.Cmd { return func() tea.Msg { return PageModPrevious{} } }

// PaginationUpdateCmd creates a command to update pagination state.
func PaginationUpdateCmd() tea.Cmd { return func() tea.Msg { return UpdatePagination{} } }

// PageSetCmd creates a command to set the current page.
func PageSetCmd(page Page) tea.Cmd { return func() tea.Msg { return PageSet{Page: page} } }

// TableSetCmd creates a command to set the selected table.
func TableSetCmd(table string) tea.Cmd { return func() tea.Msg { return TableSet{Table: table} } }

// NavigateToPageCmd creates a command to navigate to a specific page.
func NavigateToPageCmd(page Page) tea.Cmd {
	return func() tea.Msg { return NavigateToPage{Page: page} }
}

// SelectTableCmd creates a command to select a table for viewing.
func SelectTableCmd(table string) tea.Cmd { return func() tea.Msg { return SelectTable{Table: table} } }

// SetPageContentCmd creates a command to set the page content.
func SetPageContentCmd(content string) tea.Cmd {
	return func() tea.Msg { return SetPageContent{Content: content} }
}

// SetViewportContentCmd creates a command to set the viewport content.
func SetViewportContentCmd(content string) tea.Cmd {
	return func() tea.Msg { return SetViewportContent{Content: content} }
}

// FocusSetCmd creates a command to set focus to a specific element.
func FocusSetCmd(focus FocusKey) tea.Cmd { return func() tea.Msg { return FocusSet{Focus: focus} } }

// PanelFocusResetCmd creates a command to reset panel focus.
func PanelFocusResetCmd() tea.Cmd { return func() tea.Msg { return PanelFocusReset{} } }

// FormNewCmd creates a command to create a new form of the specified type.
func FormNewCmd(f FormIndex) tea.Cmd { return func() tea.Msg { return FormCreate{FormType: f} } }

// FormSetCmd creates a command to set the active form.
func FormSetCmd(form huh.Form) tea.Cmd { return func() tea.Msg { return FormSet{Form: form} } }

// FormSetValuesCmd creates a command to set form values.
func FormSetValuesCmd(v []*string) tea.Cmd { return func() tea.Msg { return FormValuesSet{Values: v} } }

// FormLenSetCmd creates a command to set the form length.
func FormLenSetCmd(formLen int) tea.Cmd {
	return func() tea.Msg { return FormLenSet{FormLen: formLen} }
}

// FormMapSetCmd creates a command to set the form field mapping.
func FormMapSetCmd(formMap []string) tea.Cmd {
	return func() tea.Msg { return FormMapSet{FormMap: formMap} }
}

// FormSubmitCmd creates a command to submit the current form.
func FormSubmitCmd() tea.Cmd { return func() tea.Msg { return FormSubmitMsg{} } }

// FormCancelCmd creates a command to cancel the current form.
func FormCancelCmd() tea.Cmd { return func() tea.Msg { return FormCancelMsg{} } }

// FormActionCmd creates a command to perform a database action from form data.
func FormActionCmd(action DatabaseCMD, table string, columns []string, values []*string) tea.Cmd {
	return func() tea.Msg {
		return FormActionMsg{
			Action:  action,
			Table:   table,
			Columns: columns,
			Values:  values,
		}
	}
}
// FormAbortCmd creates a command to abort the current form operation.
func FormAbortCmd(action DatabaseCMD, table string) tea.Cmd {
	return func() tea.Msg {
		return FormAborted{
			Action: action,
			Table:  table,
		}
	}
}

// HistoryPopCmd creates a command to pop the last page from history.
func HistoryPopCmd() tea.Cmd { return func() tea.Msg { return HistoryPop{} } }

// HistoryPushCmd creates a command to push a page onto the history stack.
func HistoryPushCmd(page PageHistory) tea.Cmd {
	return func() tea.Msg { return HistoryPush{Page: page} }
}

// FormCompletedCmd creates a command to signal form completion with optional destination.
func FormCompletedCmd(destinationPage *Page) tea.Cmd {
	return func() tea.Msg { return FormCompletedMsg{DestinationPage: destinationPage} }
}

// DatatypesFetchCmd creates a command to fetch all datatypes.
func DatatypesFetchCmd() tea.Cmd { return func() tea.Msg { return DatatypesFetchMsg{} } }

// DatatypesFetchResultCmd creates a command to return datatype fetch results.
func DatatypesFetchResultCmd(data []db.Datatypes) tea.Cmd {
	return func() tea.Msg { return DatatypesFetchResultsMsg{Data: data} }
}

// DatabaseGetCmd creates a command to fetch a single database record.
func DatabaseGetCmd(source FetchSource, table db.DBTable, id int64) tea.Cmd {
	return func() tea.Msg {
		return DatabaseGetMsg{
			Source: source,
			ID:     id,
			Table:  table,
		}
	}
}
// DatabaseListCmd creates a command to list all records from a table.
func DatabaseListCmd(source FetchSource, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListMsg{
			Source: source,
			Table:  table,
		}
	}
}
// DatabaseListFilteredCmd creates a command to list filtered database records.
func DatabaseListFilteredCmd(source FetchSource, table db.DBTable, columns []string, whereColumn string, value any) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListFilteredMsg{
			Source:      source,
			Table:       table,
			Columns:     columns,
			WhereColumn: whereColumn,
			Value:       value,
		}
	}
}
// DatabaseGetRowsCmd creates a command to return single row results.
func DatabaseGetRowsCmd(source FetchSource, rows any, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseGetRowMsg{
			Source: source,
			Table:  table,
			Rows:   rows,
		}
	}
}
// DatabaseListRowsCmd creates a command to return list row results.
func DatabaseListRowsCmd(source FetchSource, rows any, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListRowsMsg{
			Source: source,
			Table:  table,
			Rows:   rows,
		}
	}
}
// DatabaseListFilteredRowsCmd creates a command to return filtered row results.
func DatabaseListFilteredRowsCmd(source FetchSource, rows any, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListFilteredRowsMsg{
			Source: source,
			Table:  table,
			Rows:   rows,
		}
	}
}
// DatabaseDeleteEntryCmd creates a command to delete a database entry.
func DatabaseDeleteEntryCmd(id int, table string) tea.Cmd {
	return func() tea.Msg { return DatabaseDeleteEntry{Id: id, Table: table} }
}

// DatabaseInsertCmd creates a command to insert a new database record.
func DatabaseInsertCmd(table db.DBTable, columns []string, values []*string) tea.Cmd {
	return func() tea.Msg {
		return DatabaseInsertEntry{
			Table:   table,
			Columns: columns,
			Values:  values,
		}
	}
}
// DatabaseUpdateEntryCmd creates a command to update a database record.
func DatabaseUpdateEntryCmd(table db.DBTable, rowID string, values map[string]string) tea.Cmd {
	return func() tea.Msg {
		return DatabaseUpdateEntry{
			Table:  table,
			RowID:  rowID,
			Values: values,
		}
	}
}

// ColumnsFetchedCmd creates a command to return fetched columns and types.
func ColumnsFetchedCmd(columns *[]string, columnTypes *[]*sql.ColumnType) tea.Cmd {
	return func() tea.Msg { return ColumnsFetched{Columns: columns, ColumnTypes: columnTypes} }
}
// ColumnsSetCmd creates a command to set table columns.
func ColumnsSetCmd(columns *[]string) tea.Cmd {
	return func() tea.Msg { return ColumnsSet{Columns: columns} }
}

// ColumnTypesSetCmd creates a command to set column types.
func ColumnTypesSetCmd(columnTypes *[]*sql.ColumnType) tea.Cmd {
	return func() tea.Msg { return ColumnTypesSet{ColumnTypes: columnTypes} }
}

// HeadersSetCmd creates a command to set table headers.
func HeadersSetCmd(columns []string) tea.Cmd {
	return func() tea.Msg { return HeadersSet{Headers: columns} }
}

// RowsSetCmd creates a command to set table rows.
func RowsSetCmd(rows [][]string) tea.Cmd {
	return func() tea.Msg { return RowsSet{Rows: rows} }
}

// PaginatorUpdateCmd creates a command to update the paginator configuration.
func PaginatorUpdateCmd(perPage, totalPages int) tea.Cmd {
	return func() tea.Msg { return PaginatorUpdate{PerPage: perPage, TotalPages: totalPages} }
}

// ErrorSetCmd creates a command to set an error state.
func ErrorSetCmd(err error) tea.Cmd { return func() tea.Msg { return ErrorSet{Err: err} } }

// StatusSetCmd creates a command to set the application status.
func StatusSetCmd(status ApplicationState) tea.Cmd {
	return func() tea.Msg { return StatusSet{Status: status} }
}

// DialogSetCmd creates a command to set the active dialog.
func DialogSetCmd(dialog *DialogModel) tea.Cmd {
	return func() tea.Msg { return DialogSet{Dialog: dialog} }
}

// DialogActiveSetCmd creates a command to set dialog active state.
func DialogActiveSetCmd(active bool) tea.Cmd {
	return func() tea.Msg { return DialogActiveSet{DialogActive: active} }
}

// DialogReadyOKSetCmd creates a command to set dialog OK button ready state.
func DialogReadyOKSetCmd(ready bool) tea.Cmd {
	return func() tea.Msg { return DialogReadyOKSet{Ready: ready} }
}

// RootSetCmd creates a command to set the root model.
func RootSetCmd(root model.Root) tea.Cmd { return func() tea.Msg { return RootSet{Root: root} } }

// DatatypeMenuSetCmd creates a command to set the datatype menu.
func DatatypeMenuSetCmd(menu []string) tea.Cmd {
	return func() tea.Msg { return DatatypeMenuSet{DatatypeMenu: menu} }
}

// PageMenuSetCmd creates a command to set the page menu.
func PageMenuSetCmd(menu []Page) tea.Cmd {
	return func() tea.Msg { return PageMenuSet{PageMenu: menu} }
}

// ShowDialogCmd creates a command to show a dialog with the specified parameters.
func ShowDialogCmd(title, message string, showCancel bool, action DialogAction) tea.Cmd {
	dialog := NewDialog(title, message, showCancel, action)
	return func() tea.Msg {
		return tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)()
	}
}

// HideDialogCmd creates a command to hide the current dialog.
func HideDialogCmd() tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			DialogActiveSetCmd(false),
			FocusSetCmd(PAGEFOCUS),
		)()
	}
}

// SetErrorWithStatusCmd creates a command to set both error and status atomically.
func SetErrorWithStatusCmd(err error, status ApplicationState) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			ErrorSetCmd(err),
			StatusSetCmd(status),
			LoadingStopCmd(),
		)()
	}
}

// SetFormDataCmd creates a command to set complete form data in one batch.
func SetFormDataCmd(form huh.Form, formLen int, values []*string, formMap []string) tea.Cmd {
	return func() tea.Msg {
		cmds := []tea.Cmd{
			FormSetCmd(form),
			FormLenSetCmd(formLen),
			FormSetValuesCmd(values),
		}

		// Only set FormMap if provided
		if len(formMap) > 0 {
			cmds = append(cmds, FormMapSetCmd(formMap))
		}

		cmds = append(cmds, LoadingStopCmd())

		return tea.Batch(cmds...)()
	}
}

// SetTableDataCmd creates a command to set complete table data in one batch.
func SetTableDataCmd(headers []string, rows [][]string, maxRows int) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			HeadersSetCmd(headers),
			RowsSetCmd(rows),
			PaginatorUpdateCmd(maxRows, len(rows)),
			LoadingStopCmd(),
		)()
	}
}

// FetchTableHeadersRowsCmd creates a command to fetch table headers and rows.
func FetchTableHeadersRowsCmd(c config.Config, t string, page *Page) tea.Cmd {
	return func() tea.Msg {
		return FetchHeadersRows{
			Config: c,
			Table:  t,
			Page:   page,
		}
	}
}

// TableHeadersRowsFetchedCmd creates a command to return fetched table headers and rows.
func TableHeadersRowsFetchedCmd(headers []string, rows [][]string, page *Page) tea.Cmd {
	return func() tea.Msg {
		return TableHeadersRowsFetchedMsg{
			Headers: headers,
			Rows:    rows,
			Page:    page,
		}
	}
}

// GetColumnsCmd creates a command to fetch column metadata for a table.
func GetColumnsCmd(c config.Config, t string) tea.Cmd {
	return func() tea.Msg {
		return GetColumns{
			Config: c,
			Table:  t,
		}
	}
}

// DbResultCmd creates a command to return a database operation result.
func DbResultCmd(res sql.Result, table string) tea.Cmd {
	return func() tea.Msg {
		return DbResMsg{
			Result: res,
			Table:  table,
		}
	}
}

// GetFullTreeResCMD creates a command to return full tree query results.
func GetFullTreeResCMD(rows []db.GetRouteTreeByRouteIDRow) tea.Cmd {
	return func() tea.Msg {
		return GetFullTreeResMsg{
			Rows: rows,
		}
	}
}
// BuildTreeFromRowsCmd creates a command to build a tree from database rows.
func BuildTreeFromRowsCmd(rows []db.GetRouteTreeByRouteIDRow) tea.Cmd {
	return func() tea.Msg {
		return BuildTreeFromRows{
			Rows: rows,
		}
	}
}

// DatabaseTreeCMD creates a command to request database tree view.
func DatabaseTreeCMD() tea.Cmd {
	return func() tea.Msg {
		return DatabaseTreeMsg{}
	}
}

// UpdateMaxCursorCmd creates a command to update the maximum cursor based on current page view.
func (m Model) UpdateMaxCursorCmd() tea.Cmd {
	return func() tea.Msg {
		start, end := m.Paginator.GetSliceBounds(len(m.TableState.Rows))
		currentView := m.TableState.Rows[start:end]
		return UpdateMaxCursorMsg{CursorMax: len(currentView)}
	}
}

// CmsDefineDatatypeLoadCmd creates a command to load the datatype definition form.
func CmsDefineDatatypeLoadCmd() tea.Cmd {
	return func() tea.Msg {
		return CmsDefineDatatypeLoadMsg{}
	}
}

// CmsDefineDatatypeReadyCmd creates a command to signal datatype definition ready.
func CmsDefineDatatypeReadyCmd() tea.Cmd {
	return func() tea.Msg {
		return CmsDefineDatatypeReadyMsg{}
	}
}
// CmsBuildDefineDatatypeFormCmd creates a command to build the datatype definition form.
func CmsBuildDefineDatatypeFormCmd() tea.Cmd {
	return func() tea.Msg {
		return CmsBuildDefineDatatypeFormMsg{}
	}
}
// CmsEditDatatypeLoadCmd creates a command to load a datatype for editing.
func CmsEditDatatypeLoadCmd(dt db.Datatypes) tea.Cmd {
	return func() tea.Msg { return CmsEditDatatypeLoadMsg{Datatype: dt} }
}

// CmsEditDatatypeFormCmd creates a command to show the datatype edit form.
func CmsEditDatatypeFormCmd(dt db.Datatypes) tea.Cmd {
	return func() tea.Msg { return CmsEditDatatypeFormMsg{Datatype: dt} }
}

// CmsDefineDatatypeFormCmd creates a command to show the datatype definition form.
func CmsDefineDatatypeFormCmd() tea.Cmd {
	return func() tea.Msg {
		return CmsDefineDatatypeFormMsg{}
	}
}
// FormOptionsSetCmd creates a command to set form options.
func FormOptionsSetCmd(options *FormOptionsMap) tea.Cmd {
	return func() tea.Msg {
		return FormOptionsSet{
			Options: options,
		}
	}
}

// FormInitOptionsCmd creates a command to initialize form options.
func FormInitOptionsCmd(form, table string) tea.Cmd {
	return func() tea.Msg {
		return FormInitOptionsMsg{
			Form:  form,
			Table: table,
		}
	}
}
// CmsAddNewContentDataCmd creates a command to add new content data.
func CmsAddNewContentDataCmd(id types.DatatypeID) tea.Cmd {
	return func() tea.Msg {
		return CmsAddNewContentDataMsg{
			Datatype: id,
		}
	}
}
// CmsAddNewContentFieldsCmd creates a command to add new content fields.
func CmsAddNewContentFieldsCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		return CmsAddNewContentFieldsMsg{
			Datatype: id,
		}
	}
}

// RestoreBackupFromPathCmd creates a command to restore a backup from a file path.
func RestoreBackupFromPathCmd(path string) tea.Cmd {
	return func() tea.Msg {
		return RestoreBackupFromPathMsg{Path: path}
	}
}

// MediaFetchCmd creates a command to fetch all media items.
func MediaFetchCmd() tea.Cmd { return func() tea.Msg { return MediaFetchMsg{} } }

// MediaListSetCmd creates a command to set the media list.
func MediaListSetCmd(media []db.Media) tea.Cmd {
	return func() tea.Msg { return MediaListSet{MediaList: media} }
}

// MediaUploadCmd creates a command to upload a media file.
func MediaUploadCmd(filePath string) tea.Cmd {
	return func() tea.Msg {
		return MediaUploadStartMsg{FilePath: filePath}
	}
}

// UsersFetchCmd creates a command to fetch all users.
func UsersFetchCmd() tea.Cmd { return func() tea.Msg { return UsersFetchMsg{} } }

// UsersListSetCmd creates a command to set the users list.
func UsersListSetCmd(users []db.Users) tea.Cmd {
	return func() tea.Msg { return UsersListSet{UsersList: users} }
}

// RootContentSummaryFetchCmd creates a command to fetch root content summary.
func RootContentSummaryFetchCmd() tea.Cmd {
	return func() tea.Msg { return RootContentSummaryFetchMsg{} }
}
// RootContentSummarySetCmd creates a command to set root content summary data.
func RootContentSummarySetCmd(summary []db.RootContentSummary) tea.Cmd {
	return func() tea.Msg { return RootContentSummarySet{RootContentSummary: summary} }
}

// RoutesFetchCmd creates a command to fetch all routes.
func RoutesFetchCmd() tea.Cmd { return func() tea.Msg { return RoutesFetchMsg{} } }

// RoutesFetchResultCmd creates a command to return route fetch results.
func RoutesFetchResultCmd(data []db.Routes) tea.Cmd {
	return func() tea.Msg { return RoutesFetchResultsMsg{Data: data} }
}

// RouteSelectedCmd creates a command to signal route selection.
func RouteSelectedCmd(route db.Routes) tea.Cmd {
	return func() tea.Msg { return RouteSelectedMsg{Route: route} }
}

// RoutesSetCmd creates a command to set the routes list.
func RoutesSetCmd(routes []db.Routes) tea.Cmd {
	return func() tea.Msg { return RoutesSet{Routes: routes} }
}

// RootDatatypesFetchCmd creates a command to fetch root-level datatypes.
func RootDatatypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return RootDatatypesFetchMsg{} }
}
// RootDatatypesSetCmd creates a command to set root datatypes list.
func RootDatatypesSetCmd(datatypes []db.Datatypes) tea.Cmd {
	return func() tea.Msg { return RootDatatypesSet{RootDatatypes: datatypes} }
}

// AllDatatypesFetchCmd creates a command to fetch all datatypes.
func AllDatatypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AllDatatypesFetchMsg{} }
}
// AllDatatypesSetCmd creates a command to set all datatypes list.
func AllDatatypesSetCmd(datatypes []db.Datatypes) tea.Cmd {
	return func() tea.Msg { return AllDatatypesSet{AllDatatypes: datatypes} }
}

// DatatypeFieldsFetchCmd creates a command to fetch fields for a datatype.
func DatatypeFieldsFetchCmd(datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg { return DatatypeFieldsFetchMsg{DatatypeID: datatypeID} }
}
// DatatypeFieldsSetCmd creates a command to set datatype fields list.
func DatatypeFieldsSetCmd(fields []db.Fields) tea.Cmd {
	return func() tea.Msg { return DatatypeFieldsSet{Fields: fields} }
}

// RoutesByDatatypeFetchCmd creates a command to fetch routes for a specific datatype.
func RoutesByDatatypeFetchCmd(datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg { return RoutesByDatatypeFetchMsg{DatatypeID: datatypeID} }
}
// SelectedDatatypeSetCmd creates a command to set the selected datatype.
func SelectedDatatypeSetCmd(datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg { return SelectedDatatypeSet{DatatypeID: datatypeID} }
}

// BuildContentFormCmd creates a command to build a content creation form.
func BuildContentFormCmd(datatypeID types.DatatypeID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return BuildContentFormMsg{
			DatatypeID: datatypeID,
			RouteID:    routeID,
		}
	}
}

// CreateContentWithFieldsCmd creates a command to create content with field values.
func CreateContentWithFieldsCmd(
	config *config.Config,
	datatypeID types.DatatypeID,
	routeID types.RouteID,
	authorID types.UserID,
	fieldValues map[types.FieldID]string,
) tea.Cmd {
	return func() tea.Msg {
		m := Model{Config: config}
		return m.CreateContentWithFields(config, datatypeID, routeID, authorID, fieldValues)()
	}
}

// ReloadContentTreeCmd creates a command to reload the content tree for a route.
func ReloadContentTreeCmd(config *config.Config, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		m := Model{Config: config}
		return m.ReloadContentTree(config, routeID)()
	}
}

// BuildTreeFromRouteCMD creates a command to build a tree from a route ID.
func BuildTreeFromRouteCMD(id int64) tea.Cmd {
	return func() tea.Msg {
		return BuildTreeFromRouteMsg{
			RouteID: id,
		}
	}
}
