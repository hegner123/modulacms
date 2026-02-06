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

func LogModelCMD(include *[]string, exclude *[]string) tea.Cmd {
	return func() tea.Msg {
		return LogModelMsg{
			Include: include,
			Exclude: exclude,
		}
	}
}

// Basic action constructors
func ClearScreenCmd() tea.Cmd          { return func() tea.Msg { return ClearScreen{} } }
func TitleFontNextCmd() tea.Cmd        { return func() tea.Msg { return TitleFontNext{} } }
func TitleFontPreviousCmd() tea.Cmd    { return func() tea.Msg { return TitleFontPrevious{} } }
func LogMessageCmd(msg string) tea.Cmd { return func() tea.Msg { return LogMsg{Message: msg} } }

func ReadyTrueCmd() tea.Cmd  { return func() tea.Msg { return ReadyTrue{} } }
func ReadyFalseCmd() tea.Cmd { return func() tea.Msg { return ReadyFalse{} } }

// Data fetching constructors
func TablesFetchCmd() tea.Cmd { return func() tea.Msg { return TablesFetch{} } }
func TablesSetCmd(tables []string) tea.Cmd {
	return func() tea.Msg { return TablesSet{Tables: tables} }
}

// Loading state constructors
func LoadingStartCmd() tea.Cmd { return func() tea.Msg { return LoadingTrue{} } }
func LoadingStopCmd() tea.Cmd  { return func() tea.Msg { return LoadingFalse{} } }

// Cursor control constructors
func CursorUpCmd() tea.Cmd           { return func() tea.Msg { return CursorUp{} } }
func CursorDownCmd() tea.Cmd         { return func() tea.Msg { return CursorDown{} } }
func CursorResetCmd() tea.Cmd        { return func() tea.Msg { return CursorReset{} } }
func CursorSetCmd(index int) tea.Cmd { return func() tea.Msg { return CursorSet{Index: index} } }
func CursorMaxSetCmd(cursorMax int) tea.Cmd {
	return func() tea.Msg { return CursorMaxSet{CursorMax: cursorMax} }
}
func PageModNextCmd() tea.Cmd      { return func() tea.Msg { return PageModNext{} } }
func PageModPreviousCmd() tea.Cmd  { return func() tea.Msg { return PageModPrevious{} } }
func PaginationUpdateCmd() tea.Cmd { return func() tea.Msg { return UpdatePagination{} } }

// Page and navigation constructors
func PageSetCmd(page Page) tea.Cmd     { return func() tea.Msg { return PageSet{Page: page} } }
func TableSetCmd(table string) tea.Cmd { return func() tea.Msg { return TableSet{Table: table} } }
func NavigateToPageCmd(page Page) tea.Cmd {
	return func() tea.Msg { return NavigateToPage{Page: page} }
}

func SelectTableCmd(table string) tea.Cmd { return func() tea.Msg { return SelectTable{Table: table} } }

func SetPageContentCmd(content string) tea.Cmd {
	return func() tea.Msg { return SetPageContent{Content: content} }
}
func SetViewportContentCmd(content string) tea.Cmd {
	return func() tea.Msg { return SetViewportContent{Content: content} }
}

// Focus control constructors
func FocusSetCmd(focus FocusKey) tea.Cmd    { return func() tea.Msg { return FocusSet{Focus: focus} } }
func PanelFocusResetCmd() tea.Cmd           { return func() tea.Msg { return PanelFocusReset{} } }

// Form control constructors

func FormNewCmd(f FormIndex) tea.Cmd       { return func() tea.Msg { return FormCreate{FormType: f} } }
func FormSetCmd(form huh.Form) tea.Cmd     { return func() tea.Msg { return FormSet{Form: form} } }
func FormSetValuesCmd(v []*string) tea.Cmd { return func() tea.Msg { return FormValuesSet{Values: v} } }
func FormLenSetCmd(formLen int) tea.Cmd {
	return func() tea.Msg { return FormLenSet{FormLen: formLen} }
}

func FormMapSetCmd(formMap []string) tea.Cmd {
	return func() tea.Msg { return FormMapSet{FormMap: formMap} }
}
func FormSubmitCmd() tea.Cmd { return func() tea.Msg { return FormSubmitMsg{} } }
func FormCancelCmd() tea.Cmd { return func() tea.Msg { return FormCancelMsg{} } }
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
func FormAbortCmd(action DatabaseCMD, table string) tea.Cmd {
	return func() tea.Msg {
		return FormAborted{
			Action: action,
			Table:  table,
		}
	}
}

// History management constructors
func HistoryPopCmd() tea.Cmd { return func() tea.Msg { return HistoryPop{} } }
func HistoryPushCmd(page PageHistory) tea.Cmd {
	return func() tea.Msg { return HistoryPush{Page: page} }
}

// Form completion constructor
func FormCompletedCmd(destinationPage *Page) tea.Cmd {
	return func() tea.Msg { return FormCompletedMsg{DestinationPage: destinationPage} }
}

// Cms constructors
func DatatypesFetchCmd() tea.Cmd { return func() tea.Msg { return DatatypesFetchMsg{} } }
func DatatypesFetchResultCmd(data []db.Datatypes) tea.Cmd {
	return func() tea.Msg { return DatatypesFetchResultsMsg{Data: data} }
}

// Database operation constructors
func DatabaseGetCmd(source FetchSource, table db.DBTable, id int64) tea.Cmd {
	return func() tea.Msg {
		return DatabaseGetMsg{
			Source: source,
			ID:     id,
			Table:  table,
		}
	}
}
func DatabaseListCmd(source FetchSource, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListMsg{
			Source: source,
			Table:  table,
		}
	}
}
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
func DatabaseGetRowsCmd(source FetchSource, rows any, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseGetRowMsg{
			Source: source,
			Table:  table,
			Rows:   rows,
		}
	}
}
func DatabaseListRowsCmd(source FetchSource, rows any, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListRowsMsg{
			Source: source,
			Table:  table,
			Rows:   rows,
		}
	}
}
func DatabaseListFilteredRowsCmd(source FetchSource, rows any, table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListFilteredRowsMsg{
			Source: source,
			Table:  table,
			Rows:   rows,
		}
	}
}
func DatabaseDeleteEntryCmd(id int, table string) tea.Cmd {
	return func() tea.Msg { return DatabaseDeleteEntry{Id: id, Table: table} }
}
func DatabaseInsertCmd(table db.DBTable, columns []string, values []*string) tea.Cmd {
	return func() tea.Msg {
		return DatabaseInsertEntry{
			Table:   table,
			Columns: columns,
			Values:  values,
		}
	}
}
func DatabaseUpdateEntryCmd(table db.DBTable, err error) tea.Cmd {
	return func() tea.Msg { return DatabaseUpdateEntry{Table: table} }
}

// Data management constructors
func ColumnsFetchedCmd(columns *[]string, columnTypes *[]*sql.ColumnType) tea.Cmd {
	return func() tea.Msg { return ColumnsFetched{Columns: columns, ColumnTypes: columnTypes} }
}
func ColumnsSetCmd(columns *[]string) tea.Cmd {
	return func() tea.Msg { return ColumnsSet{Columns: columns} }
}
func ColumnTypesSetCmd(columnTypes *[]*sql.ColumnType) tea.Cmd {
	return func() tea.Msg { return ColumnTypesSet{ColumnTypes: columnTypes} }
}
func HeadersSetCmd(columns []string) tea.Cmd {
	return func() tea.Msg { return HeadersSet{Headers: columns} }
}
func RowsSetCmd(rows [][]string) tea.Cmd {
	return func() tea.Msg { return RowsSet{Rows: rows} }
}
func PaginatorUpdateCmd(perPage, totalPages int) tea.Cmd {
	return func() tea.Msg { return PaginatorUpdate{PerPage: perPage, TotalPages: totalPages} }
}

// Error and status constructors
func ErrorSetCmd(err error) tea.Cmd { return func() tea.Msg { return ErrorSet{Err: err} } }
func StatusSetCmd(status ApplicationState) tea.Cmd {
	return func() tea.Msg { return StatusSet{Status: status} }
}

// Dialog constructors
func DialogSetCmd(dialog *DialogModel) tea.Cmd {
	return func() tea.Msg { return DialogSet{Dialog: dialog} }
}
func DialogActiveSetCmd(active bool) tea.Cmd {
	return func() tea.Msg { return DialogActiveSet{DialogActive: active} }
}
func DialogReadyOKSetCmd(ready bool) tea.Cmd {
	return func() tea.Msg { return DialogReadyOKSet{Ready: ready} }
}

// Root and menu constructors
func RootSetCmd(root model.Root) tea.Cmd { return func() tea.Msg { return RootSet{Root: root} } }
func DatatypeMenuSetCmd(menu []string) tea.Cmd {
	return func() tea.Msg { return DatatypeMenuSet{DatatypeMenu: menu} }
}
func PageMenuSetCmd(menu []Page) tea.Cmd {
	return func() tea.Msg { return PageMenuSet{PageMenu: menu} }
}

// Composite constructors for common operations
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

func HideDialogCmd() tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			DialogActiveSetCmd(false),
			FocusSetCmd(PAGEFOCUS),
		)()
	}
}

func SetErrorWithStatusCmd(err error, status ApplicationState) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			ErrorSetCmd(err),
			StatusSetCmd(status),
			LoadingStopCmd(),
		)()
	}
}

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

func FetchTableHeadersRowsCmd(c config.Config, t string, page *Page) tea.Cmd {
	return func() tea.Msg {
		return FetchHeadersRows{
			Config: c,
			Table:  t,
			Page:   page,
		}
	}
}

func TableHeadersRowsFetchedCmd(headers []string, rows [][]string, page *Page) tea.Cmd {
	return func() tea.Msg {
		return TableHeadersRowsFetchedMsg{
			Headers: headers,
			Rows:    rows,
			Page:    page,
		}
	}
}

func GetColumnsCmd(c config.Config, t string) tea.Cmd {
	return func() tea.Msg {
		return GetColumns{
			Config: c,
			Table:  t,
		}
	}
}

func DbResultCmd(res sql.Result, table string) tea.Cmd {
	return func() tea.Msg {
		return DbResMsg{
			Result: res,
			Table:  table,
		}
	}
}

func GetFullTreeResCMD(s string, rows []db.GetRouteTreeByRouteIDRow) tea.Cmd {
	return func() tea.Msg {
		return GetFullTreeResMsg{
			Rows:    rows,
			Content: s,
		}
	}

}
func BuildTreeFromRowsCmd(rows []db.GetRouteTreeByRouteIDRow) tea.Cmd {
	return func() tea.Msg {
		return BuildTreeFromRows{
			Rows: rows,
		}
	}
}

func DatabaseTreeCMD() tea.Cmd {
	return func() tea.Msg {
		return DatabaseTreeMsg{}
	}
}

func (m Model) UpdateMaxCursorCmd() tea.Cmd {
	return func() tea.Msg {
		start, end := m.Paginator.GetSliceBounds(len(m.TableState.Rows))
		currentView := m.TableState.Rows[start:end]
		return UpdateMaxCursorMsg{CursorMax: len(currentView)}
	}
}

func CmsDefineDatatypeLoadCmd() tea.Cmd {
	return func() tea.Msg {
		return CmsDefineDatatypeLoadMsg{}
	}
}

func CmsDefineDatatypeReadyCmd() tea.Cmd {
	return func() tea.Msg {
		return CmsDefineDatatypeReadyMsg{}
	}
}
func CmsBuildDefineDatatypeFormCmd() tea.Cmd {
	return func() tea.Msg {
		return CmsBuildDefineDatatypeFormMsg{}
	}
}
func CmsEditDatatypeLoadCmd(dt db.Datatypes) tea.Cmd {
	return func() tea.Msg { return CmsEditDatatypeLoadMsg{Datatype: dt} }
}
func CmsEditDatatypeFormCmd(dt db.Datatypes) tea.Cmd {
	return func() tea.Msg { return CmsEditDatatypeFormMsg{Datatype: dt} }
}
func CmsDefineDatatypeFormCmd() tea.Cmd {
	return func() tea.Msg {
		return CmsDefineDatatypeFormMsg{}
	}
}
func FormOptionsSetCmd(options *FormOptionsMap) tea.Cmd {
	return func() tea.Msg {
		return FormOptionsSet{
			Options: options,
		}
	}
}

func FormInitOptionsCmd(form, table string) tea.Cmd {
	return func() tea.Msg {
		return FormInitOptionsMsg{
			Form:  form,
			Table: table,
		}
	}
}
func CmsAddNewContentDataCmd(id types.DatatypeID) tea.Cmd {
	return func() tea.Msg {
		return CmsAddNewContentDataMsg{
			Datatype: id,
		}
	}
}
func CmsAddNewContentFieldsCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		return CmsAddNewContentFieldsMsg{
			Datatype: id,
		}
	}
}

// Media constructors
func MediaFetchCmd() tea.Cmd       { return func() tea.Msg { return MediaFetchMsg{} } }
func MediaListSetCmd(media []db.Media) tea.Cmd {
	return func() tea.Msg { return MediaListSet{MediaList: media} }
}

// Users constructors
func UsersFetchCmd() tea.Cmd       { return func() tea.Msg { return UsersFetchMsg{} } }
func UsersListSetCmd(users []db.Users) tea.Cmd {
	return func() tea.Msg { return UsersListSet{UsersList: users} }
}

// RootContentSummary constructors
func RootContentSummaryFetchCmd() tea.Cmd {
	return func() tea.Msg { return RootContentSummaryFetchMsg{} }
}
func RootContentSummarySetCmd(summary []db.RootContentSummary) tea.Cmd {
	return func() tea.Msg { return RootContentSummarySet{RootContentSummary: summary} }
}

// Route constructors
func RoutesFetchCmd() tea.Cmd { return func() tea.Msg { return RoutesFetchMsg{} } }
func RoutesFetchResultCmd(data []db.Routes) tea.Cmd {
	return func() tea.Msg { return RoutesFetchResultsMsg{Data: data} }
}
func RouteSelectedCmd(route db.Routes) tea.Cmd {
	return func() tea.Msg { return RouteSelectedMsg{Route: route} }
}
func RoutesSetCmd(routes []db.Routes) tea.Cmd {
	return func() tea.Msg { return RoutesSet{Routes: routes} }
}

// Root datatypes constructors
func RootDatatypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return RootDatatypesFetchMsg{} }
}
func RootDatatypesSetCmd(datatypes []db.Datatypes) tea.Cmd {
	return func() tea.Msg { return RootDatatypesSet{RootDatatypes: datatypes} }
}

// All datatypes constructors
func AllDatatypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AllDatatypesFetchMsg{} }
}
func AllDatatypesSetCmd(datatypes []db.Datatypes) tea.Cmd {
	return func() tea.Msg { return AllDatatypesSet{AllDatatypes: datatypes} }
}

// Datatype fields constructors
func DatatypeFieldsFetchCmd(datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg { return DatatypeFieldsFetchMsg{DatatypeID: datatypeID} }
}
func DatatypeFieldsSetCmd(fields []db.Fields) tea.Cmd {
	return func() tea.Msg { return DatatypeFieldsSet{Fields: fields} }
}

// Routes by datatype constructors
func RoutesByDatatypeFetchCmd(datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg { return RoutesByDatatypeFetchMsg{DatatypeID: datatypeID} }
}
func SelectedDatatypeSetCmd(datatypeID types.DatatypeID) tea.Cmd {
	return func() tea.Msg { return SelectedDatatypeSet{DatatypeID: datatypeID} }
}

func BuildContentFormCmd(datatypeID types.DatatypeID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return BuildContentFormMsg{
			DatatypeID: datatypeID,
			RouteID:    routeID,
		}
	}
}

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

func ReloadContentTreeCmd(config *config.Config, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		m := Model{Config: config}
		return m.ReloadContentTree(config, routeID)()
	}
}

// Constructors
func BuildTreeFromRouteCMD(id int64) tea.Cmd {
	return func() tea.Msg {
		return BuildTreeFromRouteMsg{
			RouteID: id,
		}
	}
}
