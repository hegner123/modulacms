package cli

import (
	"database/sql"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/tree"
)

type LogModelMsg struct {
	Include *[]string
	Exclude *[]string
}

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
type UpdateMaxCursorMsg struct {
	CursorMax int
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
	Action DatabaseCMD
	Table  string
}
type FormSubmitMsg struct{}
type FormCompletedMsg struct {
	DestinationPage *Page // Optional - if nil, will try history pop, then home
}
type FormActionMsg struct {
	Action  DatabaseCMD
	Table   string
	Columns []string
	Values  []*string
}
type FormCancelMsg struct{}
type FormOptionsSet struct {
	Options *FormOptionsMap
}
type FormInitOptionsMsg struct {
	Form  string
	Table string
}

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
	RowID  string
	Values map[string]string
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

type FormMapSet struct {
	FormMap []string
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
	PageMenu []Page
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
	Page   *Page
}

type TableHeadersRowsFetchedMsg struct {
	Headers []string
	Rows    [][]string
	Page    *Page
}
type GetColumns struct {
	Config config.Config
	Table  string
}
type BuildTreeFromRows struct {
	Rows []db.GetRouteTreeByRouteIDRow
}
type GetFullTreeResMsg struct {
	Rows []db.GetRouteTreeByRouteIDRow
}
type DatabaseTreeMsg struct{}

type CmsDefineDatatypeLoadMsg struct{}
type CmsDefineDatatypeReadyMsg struct{}

type CmsBuildDefineDatatypeFormMsg struct{}
type CmsDefineDatatypeFormMsg struct{}

type CmsEditDatatypeLoadMsg struct {
	Datatype db.Datatypes
}
type CmsEditDatatypeFormMsg struct {
	Datatype db.Datatypes
}
type DatatypeUpdateSaveMsg struct {
	DatatypeID types.DatatypeID
	Parent     string
	Label      string
	Type       string
}
type DatatypeUpdatedMsg struct {
	DatatypeID types.DatatypeID
	Label      string
}
type DatatypeUpdateFailedMsg struct {
	Error error
}
type CmsGetDatatypeParentOptionsMsg struct {
	Admin bool
}

type CmsAddNewContentDataMsg struct {
	Datatype types.DatatypeID
}
type CmsAddNewContentFieldsMsg struct {
	Datatype int64
}

type ContentCreatedMsg struct {
	ContentDataID types.ContentID
	RouteID       types.RouteID
	FieldCount    int
}

type ContentCreatedWithErrorsMsg struct {
	ContentDataID types.ContentID
	RouteID       types.RouteID
	CreatedFields int
	FailedFields  []types.FieldID
}

type TreeLoadedMsg struct {
	RouteID  types.RouteID
	Stats    *tree.LoadStats
	RootNode *tree.Root
}

type BuildContentFormMsg struct {
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
}

type RoutesFetchMsg struct{}
type RoutesFetchResultsMsg struct {
	Data []db.Routes
}
type RouteSelectedMsg struct {
	Route db.Routes
}
type RoutesSet struct {
	Routes []db.Routes
}

type RootDatatypesFetchMsg struct{}
type RootDatatypesFetchResultsMsg struct {
	Data []db.Datatypes
}
type RootDatatypesSet struct {
	RootDatatypes []db.Datatypes
}

type AllDatatypesFetchMsg struct{}
type AllDatatypesFetchResultsMsg struct {
	Data []db.Datatypes
}
type AllDatatypesSet struct {
	AllDatatypes []db.Datatypes
}

type DatatypeFieldsFetchMsg struct {
	DatatypeID types.DatatypeID
}
type DatatypeFieldsFetchResultsMsg struct {
	Fields []db.Fields
}
type DatatypeFieldsSet struct {
	Fields []db.Fields
}

type RoutesByDatatypeFetchMsg struct {
	DatatypeID types.DatatypeID
}
type SelectedDatatypeSet struct {
	DatatypeID types.DatatypeID
}

type MediaFetchMsg struct{}
type MediaFetchResultsMsg struct {
	Data []db.Media
}
type MediaListSet struct {
	MediaList []db.Media
}
type RootContentSummaryFetchMsg struct{}
type RootContentSummaryFetchResultsMsg struct {
	Data []db.RootContentSummary
}
type RootContentSummarySet struct {
	RootContentSummary []db.RootContentSummary
}

// MediaUploadStartMsg triggers the async upload pipeline
type MediaUploadStartMsg struct {
	FilePath string
}

// MediaUploadedMsg signals upload completed successfully
type MediaUploadedMsg struct {
	Name string
}

// Reorder siblings
type ReorderSiblingRequestMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	Direction string // "up" or "down"
}
type ContentReorderedMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	Direction string
}

// Copy content
type CopyContentRequestMsg struct {
	SourceContentID types.ContentID
	RouteID         types.RouteID
}
type ContentCopiedMsg struct {
	SourceContentID types.ContentID
	NewContentID    types.ContentID
	RouteID         types.RouteID
	FieldCount      int
}

// Publish/Unpublish
type TogglePublishRequestMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
}
type ContentPublishToggledMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	NewStatus types.ContentStatus
}

// Archive/Unarchive
type ArchiveContentRequestMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
}
type ContentArchivedMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	NewStatus types.ContentStatus
}

type PanelFocusReset struct{}

// Users fetch messages
type UsersFetchMsg struct{}
type UsersFetchResultsMsg struct {
	Data []db.Users
}
type UsersListSet struct {
	UsersList []db.Users
}

// Backup/Restore messages
type OpenFilePickerForRestoreMsg struct{}
type RestoreBackupFromPathMsg struct{ Path string }
type BackupRestoreCompleteMsg struct{ Path string }

// Message types
type BuildTreeFromRouteMsg struct {
	RouteID int64
}
