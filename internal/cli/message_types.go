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

// LogModelMsg requests model state logging with optional include/exclude filters.
type LogModelMsg struct {
	Include *[]string
	Exclude *[]string
}

// ClearScreen requests clearing the terminal screen.
type ClearScreen struct{}

// ReadyTrue sets the application ready state to true.
type ReadyTrue struct{}

// ReadyFalse sets the application ready state to false.
type ReadyFalse struct{}

// TitleFontNext cycles to the next title font.
type TitleFontNext struct{}

// TitleFontPrevious cycles to the previous title font.
type TitleFontPrevious struct{}

// TablesFetch requests fetching the list of database tables.
type TablesFetch struct{}

// TablesSet sets the list of available database tables.
type TablesSet struct {
	Tables []string
}

// LoadingTrue sets the loading state to true.
type LoadingTrue struct{}

// LoadingFalse sets the loading state to false.
type LoadingFalse struct{}

// CursorUp moves the cursor up one position.
type CursorUp struct{}

// CursorDown moves the cursor down one position.
type CursorDown struct{}

// CursorReset resets the cursor to the initial position.
type CursorReset struct{}

// CursorSet sets the cursor to a specific index.
type CursorSet struct {
	Index int
}

// UpdateMaxCursorMsg updates the maximum cursor position.
type UpdateMaxCursorMsg struct {
	CursorMax int
}

// PageModNext navigates to the next page in pagination.
type PageModNext struct{}

// PageModPrevious navigates to the previous page in pagination.
type PageModPrevious struct{}

// PageSet sets the current page.
type PageSet struct {
	Page Page
}

// UpdatePagination triggers recalculation of pagination state.
type UpdatePagination struct{}

// TableSet sets the currently selected database table.
type TableSet struct {
	Table string
}

// SetPageContent sets the content to display on the current page.
type SetPageContent struct {
	Content string
}

// SetViewportContent sets the content for the viewport display.
type SetViewportContent struct {
	Content string
}

// FocusSet sets the focus to a specific UI element.
type FocusSet struct {
	Focus FocusKey
}

// FormCreate requests creation of a new form of the specified type.
type FormCreate struct {
	FormType FormIndex
}

// FormSet sets the active form and its initial values.
type FormSet struct {
	Form   huh.Form
	Values []*string
}

// FormValuesSet updates the form values.
type FormValuesSet struct {
	Values []*string
}

// FormAborted signals that a form operation was aborted.
type FormAborted struct {
	Action DatabaseCMD
	Table  string
}

// FormSubmitMsg requests form submission.
type FormSubmitMsg struct{}

// FormCompletedMsg signals form completion with optional destination page.
type FormCompletedMsg struct {
	DestinationPage *Page // Optional - if nil, will try history pop, then home
}

// FormActionMsg requests a database action based on form data.
type FormActionMsg struct {
	Action  DatabaseCMD
	Table   string
	Columns []string
	Values  []*string
}

// FormCancelMsg signals form cancellation.
type FormCancelMsg struct{}

// FormOptionsSet sets the options map for form select fields.
type FormOptionsSet struct {
	Options *FormOptionsMap
}

// FormInitOptionsMsg requests initialization of form options for a specific form and table.
type FormInitOptionsMsg struct {
	Form  string
	Table string
}

// HistoryPop pops the last page from navigation history.
type HistoryPop struct{}

// HistoryPush pushes a page onto the navigation history stack.
type HistoryPush struct {
	Page PageHistory
}

// NavigateToPage requests navigation to a specific page with optional menu.
type NavigateToPage struct {
	Page Page
	Menu []*Page
}

// NavigateToDatabaseCreate requests navigation to the database create page.
type NavigateToDatabaseCreate struct{}

// SelectTable selects a table for viewing or editing.
type SelectTable struct {
	Table string
}

// DatabaseDeleteEntry requests deletion of a database entry.
type DatabaseDeleteEntry struct {
	Id    int
	Table string
}

// DatabaseInsertEntry requests insertion of a new database entry.
type DatabaseInsertEntry struct {
	Table   db.DBTable
	Columns []string
	Values  []*string
}

// DatabaseUpdateEntry requests update of an existing database entry.
type DatabaseUpdateEntry struct {
	Table  db.DBTable
	RowID  string
	Values map[string]string
}

// DatabaseGetMsg requests fetching a single database record by ID.
type DatabaseGetMsg struct {
	Source FetchSource
	Table  db.DBTable
	ID     int64
}

// DatabaseListFilteredMsg requests listing database records with a WHERE filter.
type DatabaseListFilteredMsg struct {
	Source      FetchSource
	Table       db.DBTable
	Columns     []string
	WhereColumn string
	Value       any
}

// DatabaseListMsg requests listing all records from a table.
type DatabaseListMsg struct {
	Source FetchSource
	Table  db.DBTable
}

// DatabaseGetRowMsg returns a single database row result.
type DatabaseGetRowMsg struct {
	Source FetchSource
	Table  db.DBTable
	Rows   any
}

// DatabaseListFilteredRowsMsg returns filtered database rows.
type DatabaseListFilteredRowsMsg struct {
	Source FetchSource
	Table  db.DBTable
	Rows   any
}

// DatabaseListRowsMsg returns all database rows from a table.
type DatabaseListRowsMsg struct {
	Source FetchSource
	Table  db.DBTable
	Rows   any
}
// ColumnsFetched returns database column names and types after a fetch operation.
type ColumnsFetched struct {
	Columns     *[]string
	ColumnTypes *[]*sql.ColumnType
}

// ColumnsSet sets the column names for table display.
type ColumnsSet struct {
	Columns *[]string
}

// ColumnTypesSet sets the column types for table operations.
type ColumnTypesSet struct {
	ColumnTypes *[]*sql.ColumnType
}

// HeadersSet sets the table headers for display.
type HeadersSet struct {
	Headers []string
}

// RowsSet sets the table rows for display.
type RowsSet struct {
	Rows [][]string
}

// CursorMaxSet sets the maximum cursor value for the current view.
type CursorMaxSet struct {
	CursorMax int
}

// PaginatorUpdate updates paginator configuration.
type PaginatorUpdate struct {
	PerPage    int
	TotalPages int
}

// FormLenSet sets the length of the current form.
type FormLenSet struct {
	FormLen int
}

// FormMapSet sets the form field mapping.
type FormMapSet struct {
	FormMap []string
}

// ErrorSet sets the current error state.
type ErrorSet struct {
	Err error
}

// StatusSet sets the application status.
type StatusSet struct {
	Status ApplicationState
}

// DialogSet sets the active dialog model.
type DialogSet struct {
	Dialog *DialogModel
}

// DialogActiveSet sets whether a dialog is currently active.
type DialogActiveSet struct {
	DialogActive bool
}

// RootSet sets the root model state.
type RootSet struct {
	Root model.Root
}

// DatatypeMenuSet sets the datatype menu options.
type DatatypeMenuSet struct {
	DatatypeMenu []string
}

// PageMenuSet sets the page menu items.
type PageMenuSet struct {
	PageMenu []Page
}

// DialogReadyOKSet sets whether the dialog OK button is ready.
type DialogReadyOKSet struct {
	Ready bool
}

// DbResMsg returns a database operation result.
type DbResMsg struct {
	Result sql.Result
	Table  string
}

// DbErrMsg reports a database operation error.
type DbErrMsg struct {
	Error error
}

// ReadMsg returns database read results with optional error.
type ReadMsg struct {
	Result *sql.Rows
	Error  error
	RType  any
}

// DatatypesFetchMsg requests fetching all datatypes.
type DatatypesFetchMsg struct{}

// DatatypesFetchResultsMsg returns fetched datatypes.
type DatatypesFetchResultsMsg struct {
	Data []db.Datatypes
}

// DataFetchErrorMsg reports a data fetch error.
type DataFetchErrorMsg struct {
	Error error
}

// LogMsg requests logging a message.
type LogMsg struct {
	Message string
}

// FetchHeadersRows requests fetching table headers and rows for display.
type FetchHeadersRows struct {
	Config config.Config
	Table  string
	Page   *Page
}

// TableHeadersRowsFetchedMsg returns fetched table headers and rows.
type TableHeadersRowsFetchedMsg struct {
	Headers []string
	Rows    [][]string
	Page    *Page
}

// GetColumns requests fetching column metadata for a table.
type GetColumns struct {
	Config config.Config
	Table  string
}

// BuildTreeFromRows requests building a content tree from database rows.
type BuildTreeFromRows struct {
	Rows []db.GetRouteTreeByRouteIDRow
}

// GetFullTreeResMsg returns the full tree query result rows.
type GetFullTreeResMsg struct {
	Rows []db.GetRouteTreeByRouteIDRow
}

// DatabaseTreeMsg requests building the database tree view.
type DatabaseTreeMsg struct{}

// CmsDefineDatatypeLoadMsg requests loading the datatype definition form.
type CmsDefineDatatypeLoadMsg struct{}

// CmsDefineDatatypeReadyMsg signals that datatype definition is ready.
type CmsDefineDatatypeReadyMsg struct{}

// CmsBuildDefineDatatypeFormMsg requests building the datatype definition form.
type CmsBuildDefineDatatypeFormMsg struct{}

// CmsDefineDatatypeFormMsg signals that the datatype definition form is ready.
type CmsDefineDatatypeFormMsg struct{}

// CmsEditDatatypeLoadMsg requests loading an existing datatype for editing.
type CmsEditDatatypeLoadMsg struct {
	Datatype db.Datatypes
}

// CmsEditDatatypeFormMsg signals that the datatype edit form is ready.
type CmsEditDatatypeFormMsg struct {
	Datatype db.Datatypes
}

// DatatypeUpdateSaveMsg requests saving datatype updates.
type DatatypeUpdateSaveMsg struct {
	DatatypeID types.DatatypeID
	Parent     string
	Label      string
	Type       string
}

// DatatypeUpdatedMsg signals successful datatype update.
type DatatypeUpdatedMsg struct {
	DatatypeID types.DatatypeID
	Label      string
}

// DatatypeUpdateFailedMsg signals datatype update failure.
type DatatypeUpdateFailedMsg struct {
	Error error
}

// CmsGetDatatypeParentOptionsMsg requests fetching parent datatype options.
type CmsGetDatatypeParentOptionsMsg struct {
	Admin bool
}

// CmsAddNewContentDataMsg requests adding new content data for a datatype.
type CmsAddNewContentDataMsg struct {
	Datatype types.DatatypeID
}

// CmsAddNewContentFieldsMsg requests adding new content fields.
type CmsAddNewContentFieldsMsg struct {
	Datatype int64
}

// ContentCreatedMsg signals successful content creation.
type ContentCreatedMsg struct {
	ContentDataID types.ContentID
	RouteID       types.RouteID
	FieldCount    int
}

// ContentCreatedWithErrorsMsg signals content creation with partial field failures.
type ContentCreatedWithErrorsMsg struct {
	ContentDataID types.ContentID
	RouteID       types.RouteID
	CreatedFields int
	FailedFields  []types.FieldID
}

// TreeLoadedMsg signals successful content tree loading.
type TreeLoadedMsg struct {
	RouteID  types.RouteID
	Stats    *tree.LoadStats
	RootNode *tree.Root
}

// BuildContentFormMsg requests building a content creation form.
type BuildContentFormMsg struct {
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
}

// RoutesFetchMsg requests fetching all routes.
type RoutesFetchMsg struct{}

// RoutesFetchResultsMsg returns fetched routes.
type RoutesFetchResultsMsg struct {
	Data []db.Routes
}

// RouteSelectedMsg signals a route has been selected.
type RouteSelectedMsg struct {
	Route db.Routes
}

// RoutesSet sets the routes list.
type RoutesSet struct {
	Routes []db.Routes
}

// RootDatatypesFetchMsg requests fetching root-level datatypes.
type RootDatatypesFetchMsg struct{}

// RootDatatypesFetchResultsMsg returns fetched root datatypes.
type RootDatatypesFetchResultsMsg struct {
	Data []db.Datatypes
}

// RootDatatypesSet sets the root datatypes list.
type RootDatatypesSet struct {
	RootDatatypes []db.Datatypes
}

// AllDatatypesFetchMsg requests fetching all datatypes.
type AllDatatypesFetchMsg struct{}

// AllDatatypesFetchResultsMsg returns all fetched datatypes.
type AllDatatypesFetchResultsMsg struct {
	Data []db.Datatypes
}

// AllDatatypesSet sets the complete datatypes list.
type AllDatatypesSet struct {
	AllDatatypes []db.Datatypes
}

// DatatypeFieldsFetchMsg requests fetching fields for a specific datatype.
type DatatypeFieldsFetchMsg struct {
	DatatypeID types.DatatypeID
}

// DatatypeFieldsFetchResultsMsg returns fetched datatype fields.
type DatatypeFieldsFetchResultsMsg struct {
	Fields []db.Fields
}

// DatatypeFieldsSet sets the fields list for the current datatype.
type DatatypeFieldsSet struct {
	Fields []db.Fields
}

// RoutesByDatatypeFetchMsg requests fetching routes filtered by datatype.
type RoutesByDatatypeFetchMsg struct {
	DatatypeID types.DatatypeID
}

// SelectedDatatypeSet sets the currently selected datatype.
type SelectedDatatypeSet struct {
	DatatypeID types.DatatypeID
}

// MediaFetchMsg requests fetching all media items.
type MediaFetchMsg struct{}

// MediaFetchResultsMsg returns fetched media items.
type MediaFetchResultsMsg struct {
	Data []db.Media
}

// MediaListSet sets the media items list.
type MediaListSet struct {
	MediaList []db.Media
}

// RootContentSummaryFetchMsg requests fetching root content summary.
type RootContentSummaryFetchMsg struct{}

// RootContentSummaryFetchResultsMsg returns fetched root content summary.
type RootContentSummaryFetchResultsMsg struct {
	Data []db.RootContentSummary
}

// RootContentSummarySet sets the root content summary data.
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

// ReorderSiblingRequestMsg requests reordering content siblings.
type ReorderSiblingRequestMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	Direction string // "up" or "down"
}

// ContentReorderedMsg signals successful content reordering.
type ContentReorderedMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	Direction string
}

// CopyContentRequestMsg requests copying content to a new instance.
type CopyContentRequestMsg struct {
	SourceContentID types.ContentID
	RouteID         types.RouteID
}

// ContentCopiedMsg signals successful content copying.
type ContentCopiedMsg struct {
	SourceContentID types.ContentID
	NewContentID    types.ContentID
	RouteID         types.RouteID
	FieldCount      int
}

// TogglePublishRequestMsg requests toggling content publish status.
type TogglePublishRequestMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
}

// ContentPublishToggledMsg signals successful publish status toggle.
type ContentPublishToggledMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	NewStatus types.ContentStatus
}

// ArchiveContentRequestMsg requests archiving or unarchiving content.
type ArchiveContentRequestMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
}

// ContentArchivedMsg signals successful archive status change.
type ContentArchivedMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	NewStatus types.ContentStatus
}

// PanelFocusReset resets panel focus to the default state.
type PanelFocusReset struct{}

// UsersFetchMsg requests fetching all users.
type UsersFetchMsg struct{}

// UsersFetchResultsMsg returns fetched users.
type UsersFetchResultsMsg struct {
	Data []db.UserWithRoleLabelRow
}

// UsersListSet sets the users list.
type UsersListSet struct {
	UsersList []db.UserWithRoleLabelRow
}

// RolesFetchMsg requests fetching all roles.
type RolesFetchMsg struct{}

// RolesFetchResultsMsg returns fetched roles.
type RolesFetchResultsMsg struct {
	Data []db.Roles
}

// RolesListSet sets the roles list.
type RolesListSet struct {
	RolesList []db.Roles
}

// OpenFilePickerForRestoreMsg requests opening the file picker for backup restoration.
type OpenFilePickerForRestoreMsg struct{}

// RestoreBackupFromPathMsg requests restoring a backup from a file path.
type RestoreBackupFromPathMsg struct{ Path string }

// BackupRestoreCompleteMsg signals successful backup restoration.
type BackupRestoreCompleteMsg struct{ Path string }

// BuildTreeFromRouteMsg requests building a content tree from a route ID.
type BuildTreeFromRouteMsg struct {
	RouteID int64
}

// PluginsFetchMsg requests fetching plugin list from the manager.
type PluginsFetchMsg struct{}

// PluginsFetchResultsMsg returns the fetched plugin display list.
type PluginsFetchResultsMsg struct {
	Data []PluginDisplay
}

// PluginsListSet sets the plugin display list on the model.
type PluginsListSet struct {
	PluginsList []PluginDisplay
}

// PluginActionResultMsg signals the result of a plugin action (enable/disable/reload/approve).
type PluginActionResultMsg struct {
	Title   string
	Message string
}

// PluginEnableRequestMsg requests enabling a plugin.
type PluginEnableRequestMsg struct {
	Name string
}

// PluginDisableRequestMsg requests disabling a plugin.
type PluginDisableRequestMsg struct {
	Name string
}

// PluginReloadRequestMsg requests reloading a plugin.
type PluginReloadRequestMsg struct {
	Name string
}

// PluginApproveAllRoutesRequestMsg requests approving all unapproved routes for a plugin.
type PluginApproveAllRoutesRequestMsg struct {
	Name string
}

// PluginApproveAllHooksRequestMsg requests approving all unapproved hooks for a plugin.
type PluginApproveAllHooksRequestMsg struct {
	Name string
}

// PluginEnabledMsg signals that a plugin was successfully enabled.
type PluginEnabledMsg struct {
	Name string
}

// PluginDisabledMsg signals that a plugin was successfully disabled.
type PluginDisabledMsg struct {
	Name string
}

// PluginReloadedMsg signals that a plugin was successfully reloaded.
type PluginReloadedMsg struct {
	Name string
}

// PluginRoutesApprovedMsg signals that all plugin routes were approved.
type PluginRoutesApprovedMsg struct {
	Name  string
	Count int
}

// PluginHooksApprovedMsg signals that all plugin hooks were approved.
type PluginHooksApprovedMsg struct {
	Name  string
	Count int
}

// ShowApproveAllRoutesDialogMsg triggers the route approval confirmation dialog.
type ShowApproveAllRoutesDialogMsg struct {
	PluginName    string
	PendingRoutes []string // human-readable list for display
}

// ShowApproveAllHooksDialogMsg triggers the hook approval confirmation dialog.
type ShowApproveAllHooksDialogMsg struct {
	PluginName   string
	PendingHooks []string // human-readable list for display
}

// ConfigCategorySelectMsg navigates to a config category detail view.
type ConfigCategorySelectMsg struct {
	Category string
}

// ConfigFieldUpdateMsg requests updating a config field value.
type ConfigFieldUpdateMsg struct {
	Key   string
	Value string
}

// ConfigUpdateResultMsg carries the result of a config update.
type ConfigUpdateResultMsg struct {
	RestartRequired []string
	Err             error
}

// ShowDialogMsg is defined in dialog.go
