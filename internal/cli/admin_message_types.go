package cli

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// ADMIN ROUTES MESSAGES
// =============================================================================

type AdminRoutesFetchMsg struct{}
type AdminRoutesFetchResultsMsg struct {
	Data []db.AdminRoutes
}
type AdminRoutesSet struct {
	AdminRoutes []db.AdminRoutes
}

type CreateAdminRouteFromDialogRequestMsg struct {
	Title string
	Slug  string
}
type UpdateAdminRouteFromDialogRequestMsg struct {
	RouteID      string
	Title        string
	Slug         string
	OriginalSlug string
}
type DeleteAdminRouteRequestMsg struct {
	AdminRouteID types.AdminRouteID
}

type AdminRouteCreatedFromDialogMsg struct {
	AdminRouteID types.AdminRouteID
	Title        string
	Slug         string
}
type AdminRouteUpdatedFromDialogMsg struct {
	AdminRouteID types.AdminRouteID
	Title        string
	Slug         string
}
type AdminRouteDeletedMsg struct {
	AdminRouteID types.AdminRouteID
}

type ShowDeleteAdminRouteDialogMsg struct {
	AdminRouteID types.AdminRouteID
	Title        string
}
type ShowEditAdminRouteDialogMsg struct {
	Route db.AdminRoutes
}

// =============================================================================
// ADMIN DATATYPES MESSAGES
// =============================================================================

type AdminAllDatatypesFetchMsg struct{}
type AdminAllDatatypesFetchResultsMsg struct {
	Data []db.AdminDatatypes
}
type AdminAllDatatypesSet struct {
	AdminAllDatatypes []db.AdminDatatypes
}

type AdminDatatypeFieldsFetchMsg struct {
	AdminDatatypeID types.AdminDatatypeID
}
type AdminDatatypeFieldsFetchResultsMsg struct {
	Fields []db.AdminFields
}
type AdminDatatypeFieldsSet struct {
	Fields []db.AdminFields
}

type CreateAdminDatatypeFromDialogRequestMsg struct {
	Label    string
	Type     string
	ParentID string
}
type UpdateAdminDatatypeFromDialogRequestMsg struct {
	AdminDatatypeID string
	Label           string
	Type            string
	ParentID        string
}
type DeleteAdminDatatypeRequestMsg struct {
	AdminDatatypeID types.AdminDatatypeID
}

type AdminDatatypeCreatedFromDialogMsg struct {
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}
type AdminDatatypeUpdatedFromDialogMsg struct {
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}
type AdminDatatypeDeletedMsg struct {
	AdminDatatypeID types.AdminDatatypeID
}

type ShowDeleteAdminDatatypeDialogMsg struct {
	AdminDatatypeID types.AdminDatatypeID
	Label           string
	HasChildren     bool
}
type ShowAdminFormDialogMsg struct {
	Action  FormDialogAction
	Title   string
	Parents []db.AdminDatatypes
}
type ShowEditAdminDatatypeDialogMsg struct {
	Datatype db.AdminDatatypes
	Parents  []db.AdminDatatypes
}

// =============================================================================
// ADMIN FIELDS MESSAGES
// =============================================================================

type CreateAdminFieldFromDialogRequestMsg struct {
	Label           string
	Type            string
	AdminDatatypeID types.AdminDatatypeID
}
type UpdateAdminFieldFromDialogRequestMsg struct {
	AdminFieldID string
	Label        string
	Type         string
}
type DeleteAdminFieldRequestMsg struct {
	AdminFieldID types.AdminFieldID
}

type AdminFieldCreatedFromDialogMsg struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}
type AdminFieldUpdatedFromDialogMsg struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}
type AdminFieldDeletedMsg struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
}

type ShowDeleteAdminFieldDialogMsg struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}
type ShowEditAdminFieldDialogMsg struct {
	Field db.AdminFields
}

// =============================================================================
// ADMIN CONTENT MESSAGES
// =============================================================================

type AdminContentDataFetchMsg struct{}
type AdminContentDataFetchResultsMsg struct {
	Data []db.AdminContentData
}
type AdminContentDataSet struct {
	AdminContentData []db.AdminContentData
}

type AdminContentCreatedMsg struct {
	AdminContentID types.AdminContentID
}
type AdminContentDeletedMsg struct {
	AdminContentID types.AdminContentID
}
type DeleteAdminContentRequestMsg struct {
	AdminContentID types.AdminContentID
}

type AdminLoadContentFieldsMsg struct {
	Fields []AdminContentFieldDisplay
}

// =============================================================================
// ADMIN DISPLAY TYPES
// =============================================================================

// AdminContentFieldDisplay represents an admin content field for right panel display.
type AdminContentFieldDisplay struct {
	AdminContentFieldID types.AdminContentFieldID
	AdminFieldID        types.AdminFieldID
	Label               string
	Type                string
	Value               string
}
