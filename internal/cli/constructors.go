package cli

import (
	"database/sql"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

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
func FocusSetCmd(focus FocusKey) tea.Cmd { return func() tea.Msg { return FocusSet{Focus: focus} } }

// Form control constructors

func FormNewCmd(f FormIndex) tea.Cmd       { return func() tea.Msg { return FormCreate{FormType: f} } }
func FormSetCmd(form huh.Form) tea.Cmd     { return func() tea.Msg { return FormSet{Form: form} } }
func FormSetValuesCmd(v []*string) tea.Cmd { return func() tea.Msg { return FormValuesSet{Values: v} } }
func FormLenSetCmd(formLen int) tea.Cmd {
	return func() tea.Msg { return FormLenSet{FormLen: formLen} }
}
func FormSubmitCmd() tea.Cmd { return func() tea.Msg { return FormSubmitMsg{} } }
func FormCancelCmd() tea.Cmd { return func() tea.Msg { return FormCancelMsg{} } }
func FormActionCmd(action DatabaseAction, table string, columns []string, values []*string) tea.Cmd {
	return func() tea.Msg {
		// Debug log to trace FormActionCmd execution
		return tea.Batch(
			LogMessageCmd(fmt.Sprintf("FormActionCmd executed: %s action on table %s", action, table)),
			func() tea.Msg {
				return FormActionMsg{
					Action:  action,
					Table:   table,
					Columns: columns,
					Values:  values,
				}
			},
		)()
	}
}
func FormAbortCmd(action DatabaseAction, table string) tea.Cmd {
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

// Cms constructors
func DatatypesFetchCmd() tea.Cmd { return func() tea.Msg { return DatatypesFetchMsg{} } }
func DatatypesFetchResultCmd(data []db.Datatypes) tea.Cmd {
	return func() tea.Msg { return DatatypesFetchResultsMsg{Data: data} }
}

// Database operation constructors
func DatabaseGetCmd(table db.DBTable, id int) tea.Cmd {
	return func() tea.Msg {
		return DatabaseGetMsg{
			Id:    id,
			Table: table,
		}
	}
}
func DatabaseListCmd(table db.DBTable) tea.Cmd {
	return func() tea.Msg {
		return DatabaseListMsg{
			Table: table,
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
func PageMenuSetCmd(menu []*Page) tea.Cmd {
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

func SetFormDataCmd(form huh.Form, formLen int, values []*string) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			FormSetCmd(form),
			FormLenSetCmd(formLen),
			FormSetValuesCmd(values),
			LoadingStopCmd(),
		)()
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

func FetchTableHeadersRowsCmd(c config.Config, t string) tea.Cmd {
	return func() tea.Msg {
		return FetchHeadersRows{
			Config: c,
			Table:  t,
		}
	}
}

func TableHeadersRowsFetchedCmd(headers []string, rows [][]string) tea.Cmd {
	return func() tea.Msg {
		return TableHeadersRowsFetchedMsg{
			Headers: headers,
			Rows:    rows,
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
