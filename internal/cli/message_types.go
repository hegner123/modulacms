package cli

import (
	"database/sql"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

type ClearScreen struct{}

type ReadyTrue struct{}
type ReadyFalse struct{}

type TitleFontNext struct{}
type TitleFontPrevious struct{}
type TablesFetch struct{}
type TablesSet struct {
	Tables []string
}

type LoadingTrue struct{}
type LoadingFalse struct{}
type CursorUp struct{}
type CursorDown struct{}
type CursorReset struct{}
type CursorSet struct {
	Index int
}
type PageModNext struct{}
type PageModPrevious struct{}
type PageSet struct {
	Page Page
}
type UpdatePagination struct{}
type TableSet struct {
	Table string
}
type SetPageContent struct {
	Content string
}

type SetViewportContent struct {
	Content string
}
type FocusSet struct {
	Focus FocusKey
}
type FormCreate struct {
	FormType FormIndex
}

type FormSet struct {
	Form   huh.Form
	Values []*string
}
type FormValuesSet struct {
	Values []*string
}

type FormAborted struct {
	Action DatabaseAction
	Table  string
}
type FormSubmitMsg struct{}
type FormActionMsg struct {
	Action  DatabaseAction
	Table   string
	Columns []string
	Values  []*string
}
type FormCancelMsg struct{}
type HistoryPop struct{}
type HistoryPush struct {
	Page PageHistory
}
type NavigateToPage struct {
	Page Page
	Menu []*Page
}

type NavigateToDatabaseCreate struct{}
type SelectTable struct {
	Table string
}

type DatabaseDeleteEntry struct {
	Id    int
	Table string
}
type DatabaseInsertEntry struct {
	Table   db.DBTable
	Columns []string
	Values  []*string
}
type DatabaseUpdateEntry struct {
	Table  db.DBTable
	Values []*string
}
type DatabaseGetMsg struct {
	Source FetchSource
	Table  db.DBTable
	ID     int64
}

type DatabaseListFilteredMsg struct {
	Source      FetchSource
	Table       db.DBTable
	Columns     []string
	WhereColumn string
	Value       any
}
type DatabaseListMsg struct {
	Source FetchSource
	Table  db.DBTable
}
type DatabaseGetRowMsg struct {
	Source FetchSource
	Table  db.DBTable
	Rows   any
}
type DatabaseListFilteredRowsMsg struct {
	Source FetchSource
	Table  db.DBTable
	Rows   any
}
type DatabaseListRowsMsg struct {
	Source FetchSource
	Table  db.DBTable
	Rows   any
}
type ColumnsFetched struct {
	Columns     *[]string
	ColumnTypes *[]*sql.ColumnType
}
type ColumnsSet struct {
	Columns *[]string
}
type ColumnTypesSet struct {
	ColumnTypes *[]*sql.ColumnType
}
type HeadersSet struct {
	Headers []string
}
type RowsSet struct {
	Rows [][]string
}
type CursorMaxSet struct {
	CursorMax int
}
type PaginatorUpdate struct {
	PerPage    int
	TotalPages int
}
type FormLenSet struct {
	FormLen int
}
type ErrorSet struct {
	Err error
}
type StatusSet struct {
	Status ApplicationState
}
type DialogSet struct {
	Dialog *DialogModel
}
type DialogActiveSet struct {
	DialogActive bool
}
type RootSet struct {
	Root model.Root
}
type DatatypeMenuSet struct {
	DatatypeMenu []string
}
type PageMenuSet struct {
	PageMenu []*Page
}
type DialogReadyOKSet struct {
	Ready bool
}

type DbResMsg struct {
	Result sql.Result
	Table  string
}

type DbErrMsg struct {
	Error error
}

type ReadMsg struct {
	Result *sql.Rows
	Error  error
	RType  any
}
type DatatypesFetchMsg struct{}
type DatatypesFetchResultsMsg struct {
	Data []db.Datatypes
}
type DataFetchErrorMsg struct {
	Error error
}
type LogMsg struct {
	Message string
}

type FetchHeadersRows struct {
	Config config.Config
	Table  string
}

type TableHeadersRowsFetchedMsg struct {
	Headers []string
	Rows    [][]string
}
type GetColumns struct {
	Config config.Config
	Table  string
}
