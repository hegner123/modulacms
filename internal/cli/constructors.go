package cli

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

// Basic action constructors
func ClearScreenCmd() tea.Cmd       { return func() tea.Msg { return ClearScreen{} } }
func TitleFontNextCmd() tea.Cmd     { return func() tea.Msg { return TitleFontNext{} } }
func TitleFontPreviousCmd() tea.Cmd { return func() tea.Msg { return TitleFontPrevious{} } }
func LogMessage(msg string) tea.Cmd { return func() tea.Msg { return LogMsg{Message: msg} } }

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

// Page and navigation constructors
func PageSetCmd(page Page) tea.Cmd     { return func() tea.Msg { return PageSet{Page: page} } }
func TableSetCmd(table string) tea.Cmd { return func() tea.Msg { return TableSet{Table: table} } }
func NavigateToPageCmd(page Page) tea.Cmd {
	return func() tea.Msg { return NavigateToPage{Page: page} }
}
func NavigateToDatabaseCreateCmd() tea.Cmd {
	return func() tea.Msg { return NavigateToDatabaseCreate{} }
}

// Page routing constructors with history preservation
func NavigateToTableCreatePageCmd(currentPage Page, cursor int, table string, config *config.Config) tea.Cmd {
	return func() tea.Msg {
		return NavigateToTableCreatePage{
			CurrentPage: currentPage,
			Cursor:      cursor,
			Table:       table,
			Config:      config,
		}
	}
}

func NavigateToTableUpdatePageCmd(currentPage Page, cursor int, table string, config *config.Config, targetPage *Page) tea.Cmd {
	return func() tea.Msg {
		return NavigateToTableUpdatePage{
			CurrentPage: currentPage,
			Cursor:      cursor,
			Table:       table,
			Config:      config,
			TargetPage:  targetPage,
		}
	}
}

func NavigateToTableReadPageCmd(currentPage Page, cursor int, table string, config *config.Config, targetPage *Page) tea.Cmd {
	return func() tea.Msg {
		return NavigateToTableReadPage{
			CurrentPage: currentPage,
			Cursor:      cursor,
			Table:       table,
			Config:      config,
			TargetPage:  targetPage,
		}
	}
}

func NavigateToTableDeletePageCmd(currentPage Page, cursor int, table string, config *config.Config, targetPage *Page) tea.Cmd {
	return func() tea.Msg {
		return NavigateToTableDeletePage{
			CurrentPage: currentPage,
			Cursor:      cursor,
			Table:       table,
			Config:      config,
			TargetPage:  targetPage,
		}
	}
}

func NavigateToUpdateFormPageCmd(currentPage Page, cursor int, table string, config *config.Config) tea.Cmd {
	return func() tea.Msg {
		return NavigateToUpdateFormPage{
			CurrentPage: currentPage,
			Cursor:      cursor,
			Table:       table,
			Config:      config,
		}
	}
}

func NavigateToReadSinglePageCmd(currentPage Page, cursor int) tea.Cmd {
	return func() tea.Msg {
		return NavigateToReadSinglePage{
			CurrentPage: currentPage,
			Cursor:      cursor,
		}
	}
}

func NavigateToConfigPageCmd(currentPage Page, cursor int, config *config.Config, pageMenu []*Page) tea.Cmd {
	return func() tea.Msg {
		return NavigateToConfigPage{
			CurrentPage: currentPage,
			Cursor:      cursor,
			Config:      config,
			PageMenu:    pageMenu,
		}
	}
}

func NavigateWithDefaultRouterCmd(currentPage Page, cursor int, config *config.Config, pageMenu []*Page, pages []Page) tea.Cmd {
	return func() tea.Msg {
		return NavigateWithDefaultRouter{
			CurrentPage: currentPage,
			Cursor:      cursor,
			Config:      config,
			PageMenu:    pageMenu,
			Pages:       pages,
		}
	}
}
func SelectTableCmd(table string) tea.Cmd { return func() tea.Msg { return SelectTable{Table: table} } }

// Focus control constructors
func FocusSetCmd(focus FocusKey) tea.Cmd { return func() tea.Msg { return FocusSet{Focus: focus} } }

// Form control constructors
func FormSetCmd(form huh.Form) tea.Cmd { return func() tea.Msg { return FormSet{Form: form} } }
func FormAbortedCmd() tea.Cmd          { return func() tea.Msg { return FormAborted{} } }
func FormLenSetCmd(formLen int) tea.Cmd {
	return func() tea.Msg { return FormLenSet{FormLen: formLen} }
}

// History management constructors
func HistoryPopCmd() tea.Cmd { return func() tea.Msg { return HistoryPop{} } }
func HistoryPushCmd(page PageHistory) tea.Cmd {
	return func() tea.Msg { return HistoryPush{Page: page} }
}

// Database operation constructors
func DatabaseDeleteEntryCmd(id int, table string) tea.Cmd {
	return func() tea.Msg { return DatabaseDeleteEntry{Id: id, Table: table} }
}
func DatabaseCreateEntryCmd(table db.DBTable, err error) tea.Cmd {
	return func() tea.Msg { return DatabaseCreateEntry{Table: table, Err: err} }
}
func DatabaseUpdateEntryCmd(table db.DBTable, err error) tea.Cmd {
	return func() tea.Msg { return DatabaseUpdateEntry{Table: table, Err: err} }
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
func HeadersSetCmd(headers []string) tea.Cmd {
	return func() tea.Msg { return HeadersSet{Headers: headers} }
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
func ShowDialogCmd(title, message string, showCancel bool) tea.Cmd {
	dialog := NewDialog(title, message, showCancel)
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

func SetFormDataCmd(form huh.Form, formLen int) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			FormSetCmd(form),
			FormLenSetCmd(formLen),
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

func NavigateWithHistoryCmd(targetPage Page, currentPage Page, cursor int) tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			PageSetCmd(targetPage),
			HistoryPushCmd(PageHistory{Page: currentPage, Cursor: cursor}),
			CursorResetCmd(),
		)()
	}
}

