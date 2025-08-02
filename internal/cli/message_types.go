package cli

import (
	"database/sql"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)


type ClearScreen struct{}

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
type PageSet struct {
	Page Page
}
type TableSet struct {
	Table string
}

type FocusSet struct {
	Focus FocusKey
}

type FormSet struct {
	Form huh.Form
}

type FormAborted struct {
}
type HistoryPop struct{}
type HistoryPush struct {
	Page PageHistory
}
type NavigateToPage struct {
	Page Page
}

type NavigateToDatabaseCreate struct{}
type SelectTable struct {
	Table string
}

type DatabaseDeleteEntry struct {
	Id    int
	Table string
}
type DatabaseCreateEntry struct {
	Table db.DBTable
	Err   error
}
type DatabaseUpdateEntry struct {
	Table db.DBTable
	Err   error
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

type DbErrMsg struct {
	Error error
}

type ReadMsg struct {
	Result *sql.Rows
	Error  error
	RType  any
}

type DatatypesFetchedMsg struct {
	data []db.Datatypes
}
type DataFetchErrorMsg struct {
	Error error
}

