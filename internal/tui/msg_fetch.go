package tui

import (
	"database/sql"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// FetchErrMsg carries an error from a failed fetch operation.
type FetchErrMsg struct {
	Error error
}

// TablesFetch requests fetching the list of database tables.
type TablesFetch struct{}

// TablesSet sets the list of available database tables.
type TablesSet struct {
	Tables []string
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

// ColumnsFetched returns database column names and types after a fetch operation.
type ColumnsFetched struct {
	Columns     *[]string
	ColumnTypes *[]*sql.ColumnType
}

// ColumnInfoSetMsg sets column names and types together.
type ColumnInfoSetMsg struct {
	Columns     *[]string
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

// DatatypesFetchMsg requests fetching all datatypes.
type DatatypesFetchMsg struct{}

// DatatypesFetchResultsMsg returns fetched datatypes.
type DatatypesFetchResultsMsg struct {
	Data []db.Datatypes
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

// RoutesFetchMsg requests fetching all routes.
type RoutesFetchMsg struct{}

// RoutesFetchResultsMsg returns fetched routes.
type RoutesFetchResultsMsg struct {
	Data []db.Routes
}

// RoutesSet sets the routes list.
type RoutesSet struct {
	Routes []db.Routes
}

// RouteSelectedMsg signals a route has been selected.
type RouteSelectedMsg struct {
	Route db.Routes
}

// MediaFetchMsg requests fetching all media items.
type MediaFetchMsg struct{}

// MediaFetchResultsMsg returns fetched media items and folders.
type MediaFetchResultsMsg struct {
	Data    []db.Media
	Folders []db.MediaFolder
}

// MediaListSet sets the media items list and folders.
type MediaListSet struct {
	MediaList  []db.Media
	FolderList []db.MediaFolder
}

// RootContentSummaryFetchMsg requests fetching root content summary.
type RootContentSummaryFetchMsg struct{}

// RootContentSummaryFetchResultsMsg returns fetched root content summary.
type RootContentSummaryFetchResultsMsg struct {
	Data     []db.ContentDataTopLevel
	TitleMap map[string]string // ContentDataID → title from _title field
}

// RootContentSummarySet sets the root content summary data.
type RootContentSummarySet struct {
	RootContentSummary []db.ContentDataTopLevel
}

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

// --- Pipeline management messages ---

// PipelinesFetchMsg requests fetching the pipeline chains from the registry.
type PipelinesFetchMsg struct{}

// PipelinesFetchResultsMsg returns the fetched pipeline display list.
type PipelinesFetchResultsMsg struct {
	Data []PipelineDisplay
}

// PipelinesListSet sets the pipeline display list on the model.
type PipelinesListSet struct {
	PipelinesList []PipelineDisplay
}

// PipelineEntriesFetchMsg requests fetching entries for a specific pipeline chain.
type PipelineEntriesFetchMsg struct {
	Key string
}

// PipelineEntriesFetchResultsMsg returns entries for a pipeline chain.
type PipelineEntriesFetchResultsMsg struct {
	Entries []PipelineEntryDisplay
}

// PipelineEntriesSet sets the pipeline entries on the model.
type PipelineEntriesSet struct {
	PipelineEntries []PipelineEntryDisplay
}

// --- Webhook management messages ---

// WebhooksFetchMsg requests fetching all webhooks.
type WebhooksFetchMsg struct{}

// WebhooksFetchResultsMsg returns fetched webhooks.
type WebhooksFetchResultsMsg struct {
	Data []db.Webhook
}

// WebhooksListSet sets the webhooks list on the model.
type WebhooksListSet struct {
	WebhooksList []db.Webhook
}

// FieldTypesFetchMsg requests fetching all field types.
type FieldTypesFetchMsg struct{}

// FieldTypesFetchResultsMsg returns fetched field types.
type FieldTypesFetchResultsMsg struct {
	Data []db.FieldTypes
}

// FieldTypesSet sets the field types list.
type FieldTypesSet struct {
	FieldTypes []db.FieldTypes
}

// --- Validation fetch messages ---

// ValidationsFetchMsg requests fetching all validations.
type ValidationsFetchMsg struct{}

// ValidationsFetchResultsMsg returns fetched validations.
type ValidationsFetchResultsMsg struct {
	Data []db.Validation
}

// ValidationsSet sets the validations list.
type ValidationsSet struct {
	Validations []db.Validation
}

// --- i18n locale messages ---

// LocaleListMsg carries the list of enabled locales for the locale picker.
type LocaleListMsg struct {
	Locales []db.Locale
	Err     error
}

// LocaleSwitchMsg indicates the user selected a new locale.
type LocaleSwitchMsg struct {
	Locale string
}
