package cli

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// ADMIN ROUTES MESSAGES
// =============================================================================

// AdminRoutesFetchMsg requests fetching all admin routes.
type AdminRoutesFetchMsg struct{}

// AdminRoutesFetchResultsMsg returns fetched admin routes.
type AdminRoutesFetchResultsMsg struct {
	Data []db.AdminRoutes
}

// AdminRoutesSet sets the admin routes list.
type AdminRoutesSet struct {
	AdminRoutes []db.AdminRoutes
}

// CreateAdminRouteFromDialogRequestMsg requests creating an admin route from dialog input.
type CreateAdminRouteFromDialogRequestMsg struct {
	Title string
	Slug  string
}

// UpdateAdminRouteFromDialogRequestMsg requests updating an admin route from dialog input.
type UpdateAdminRouteFromDialogRequestMsg struct {
	RouteID      string
	Title        string
	Slug         string
	OriginalSlug string
}

// DeleteAdminRouteRequestMsg requests deleting an admin route.
type DeleteAdminRouteRequestMsg struct {
	AdminRouteID types.AdminRouteID
}

// AdminRouteCreatedFromDialogMsg signals successful admin route creation.
type AdminRouteCreatedFromDialogMsg struct {
	AdminRouteID types.AdminRouteID
	Title        string
	Slug         string
}

// AdminRouteUpdatedFromDialogMsg signals successful admin route update.
type AdminRouteUpdatedFromDialogMsg struct {
	AdminRouteID types.AdminRouteID
	Title        string
	Slug         string
}

// AdminRouteDeletedMsg signals successful admin route deletion.
type AdminRouteDeletedMsg struct {
	AdminRouteID types.AdminRouteID
}

// ShowDeleteAdminRouteDialogMsg requests showing the delete admin route confirmation dialog.
type ShowDeleteAdminRouteDialogMsg struct {
	AdminRouteID types.AdminRouteID
	Title        string
}

// ShowEditAdminRouteDialogMsg requests showing the edit admin route dialog.
type ShowEditAdminRouteDialogMsg struct {
	Route db.AdminRoutes
}

// =============================================================================
// ADMIN DATATYPES MESSAGES
// =============================================================================

// AdminAllDatatypesFetchMsg requests fetching all admin datatypes.
type AdminAllDatatypesFetchMsg struct{}

// AdminAllDatatypesFetchResultsMsg returns fetched admin datatypes.
type AdminAllDatatypesFetchResultsMsg struct {
	Data []db.AdminDatatypes
}

// AdminAllDatatypesSet sets the admin datatypes list.
type AdminAllDatatypesSet struct {
	AdminAllDatatypes []db.AdminDatatypes
}

// AdminDatatypeFieldsFetchMsg requests fetching fields for a specific admin datatype.
type AdminDatatypeFieldsFetchMsg struct {
	AdminDatatypeID types.AdminDatatypeID
}

// AdminDatatypeFieldsFetchResultsMsg returns fetched admin datatype fields.
type AdminDatatypeFieldsFetchResultsMsg struct {
	Fields []db.AdminFields
}

// AdminDatatypeFieldsSet sets the admin datatype fields list.
type AdminDatatypeFieldsSet struct {
	Fields []db.AdminFields
}

// CreateAdminDatatypeFromDialogRequestMsg requests creating an admin datatype from dialog input.
type CreateAdminDatatypeFromDialogRequestMsg struct {
	Label    string
	Type     string
	ParentID string
}

// UpdateAdminDatatypeFromDialogRequestMsg requests updating an admin datatype from dialog input.
type UpdateAdminDatatypeFromDialogRequestMsg struct {
	AdminDatatypeID string
	Label           string
	Type            string
	ParentID        string
}

// DeleteAdminDatatypeRequestMsg requests deleting an admin datatype.
type DeleteAdminDatatypeRequestMsg struct {
	AdminDatatypeID types.AdminDatatypeID
}

// AdminDatatypeCreatedFromDialogMsg signals successful admin datatype creation.
type AdminDatatypeCreatedFromDialogMsg struct {
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}

// AdminDatatypeUpdatedFromDialogMsg signals successful admin datatype update.
type AdminDatatypeUpdatedFromDialogMsg struct {
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}

// AdminDatatypeDeletedMsg signals successful admin datatype deletion.
type AdminDatatypeDeletedMsg struct {
	AdminDatatypeID types.AdminDatatypeID
}

// ShowDeleteAdminDatatypeDialogMsg requests showing the delete admin datatype confirmation dialog.
type ShowDeleteAdminDatatypeDialogMsg struct {
	AdminDatatypeID types.AdminDatatypeID
	Label           string
	HasChildren     bool
}

// ShowAdminFormDialogMsg requests showing an admin form dialog.
type ShowAdminFormDialogMsg struct {
	Action  FormDialogAction
	Title   string
	Parents []db.AdminDatatypes
}

// ShowEditAdminDatatypeDialogMsg requests showing the edit admin datatype dialog.
type ShowEditAdminDatatypeDialogMsg struct {
	Datatype db.AdminDatatypes
	Parents  []db.AdminDatatypes
}

// =============================================================================
// ADMIN FIELDS MESSAGES
// =============================================================================

// CreateAdminFieldFromDialogRequestMsg requests creating an admin field from dialog input.
type CreateAdminFieldFromDialogRequestMsg struct {
	Label           string
	Type            string
	AdminDatatypeID types.AdminDatatypeID
}

// UpdateAdminFieldFromDialogRequestMsg requests updating an admin field from dialog input.
type UpdateAdminFieldFromDialogRequestMsg struct {
	AdminFieldID string
	Label        string
	Type         string
}

// DeleteAdminFieldRequestMsg requests deleting an admin field.
type DeleteAdminFieldRequestMsg struct {
	AdminFieldID types.AdminFieldID
}

// AdminFieldCreatedFromDialogMsg signals successful admin field creation.
type AdminFieldCreatedFromDialogMsg struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}

// AdminFieldUpdatedFromDialogMsg signals successful admin field update.
type AdminFieldUpdatedFromDialogMsg struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}

// AdminFieldDeletedMsg signals successful admin field deletion.
type AdminFieldDeletedMsg struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
}

// ShowDeleteAdminFieldDialogMsg requests showing the delete admin field confirmation dialog.
type ShowDeleteAdminFieldDialogMsg struct {
	AdminFieldID    types.AdminFieldID
	AdminDatatypeID types.AdminDatatypeID
	Label           string
}

// ShowEditAdminFieldDialogMsg requests showing the edit admin field dialog.
type ShowEditAdminFieldDialogMsg struct {
	Field db.AdminFields
}

// =============================================================================
// ADMIN CONTENT MESSAGES
// =============================================================================

// AdminContentDataFetchMsg requests fetching all admin content data.
type AdminContentDataFetchMsg struct{}

// AdminContentDataFetchResultsMsg returns fetched admin content data.
type AdminContentDataFetchResultsMsg struct {
	Data []db.AdminContentData
}

// AdminContentDataSet sets the admin content data list.
type AdminContentDataSet struct {
	AdminContentData []db.AdminContentData
}

// AdminContentCreatedMsg signals successful admin content creation.
type AdminContentCreatedMsg struct {
	AdminContentID types.AdminContentID
}

// AdminContentDeletedMsg signals successful admin content deletion.
type AdminContentDeletedMsg struct {
	AdminContentID types.AdminContentID
}

// DeleteAdminContentRequestMsg requests deleting admin content.
type DeleteAdminContentRequestMsg struct {
	AdminContentID types.AdminContentID
}

// AdminLoadContentFieldsMsg requests loading admin content fields for display.
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

// =============================================================================
// FIELD TYPES MESSAGES
// =============================================================================

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

// CreateFieldTypeFromDialogRequestMsg requests creating a field type from dialog input.
type CreateFieldTypeFromDialogRequestMsg struct {
	Type  string
	Label string
}

// UpdateFieldTypeFromDialogRequestMsg requests updating a field type from dialog input.
type UpdateFieldTypeFromDialogRequestMsg struct {
	FieldTypeID string
	Type        string
	Label       string
}

// DeleteFieldTypeRequestMsg requests deleting a field type.
type DeleteFieldTypeRequestMsg struct {
	FieldTypeID types.FieldTypeID
}

// FieldTypeCreatedFromDialogMsg signals successful field type creation.
type FieldTypeCreatedFromDialogMsg struct {
	FieldTypeID types.FieldTypeID
	Type        string
	Label       string
}

// FieldTypeUpdatedFromDialogMsg signals successful field type update.
type FieldTypeUpdatedFromDialogMsg struct {
	FieldTypeID types.FieldTypeID
	Type        string
	Label       string
}

// FieldTypeDeletedMsg signals successful field type deletion.
type FieldTypeDeletedMsg struct {
	FieldTypeID types.FieldTypeID
}

// ShowDeleteFieldTypeDialogMsg requests showing the delete field type confirmation dialog.
type ShowDeleteFieldTypeDialogMsg struct {
	FieldTypeID types.FieldTypeID
	Label       string
}

// ShowEditFieldTypeDialogMsg requests showing the edit field type dialog.
type ShowEditFieldTypeDialogMsg struct {
	FieldType db.FieldTypes
}

// =============================================================================
// ADMIN FIELD TYPES MESSAGES
// =============================================================================

// AdminFieldTypesFetchMsg requests fetching all admin field types.
type AdminFieldTypesFetchMsg struct{}

// AdminFieldTypesFetchResultsMsg returns fetched admin field types.
type AdminFieldTypesFetchResultsMsg struct {
	Data []db.AdminFieldTypes
}

// AdminFieldTypesSet sets the admin field types list.
type AdminFieldTypesSet struct {
	AdminFieldTypes []db.AdminFieldTypes
}

// CreateAdminFieldTypeFromDialogRequestMsg requests creating an admin field type from dialog input.
type CreateAdminFieldTypeFromDialogRequestMsg struct {
	Type  string
	Label string
}

// UpdateAdminFieldTypeFromDialogRequestMsg requests updating an admin field type from dialog input.
type UpdateAdminFieldTypeFromDialogRequestMsg struct {
	AdminFieldTypeID string
	Type             string
	Label            string
}

// DeleteAdminFieldTypeRequestMsg requests deleting an admin field type.
type DeleteAdminFieldTypeRequestMsg struct {
	AdminFieldTypeID types.AdminFieldTypeID
}

// AdminFieldTypeCreatedFromDialogMsg signals successful admin field type creation.
type AdminFieldTypeCreatedFromDialogMsg struct {
	AdminFieldTypeID types.AdminFieldTypeID
	Type             string
	Label            string
}

// AdminFieldTypeUpdatedFromDialogMsg signals successful admin field type update.
type AdminFieldTypeUpdatedFromDialogMsg struct {
	AdminFieldTypeID types.AdminFieldTypeID
	Type             string
	Label            string
}

// AdminFieldTypeDeletedMsg signals successful admin field type deletion.
type AdminFieldTypeDeletedMsg struct {
	AdminFieldTypeID types.AdminFieldTypeID
}

// ShowDeleteAdminFieldTypeDialogMsg requests showing the delete admin field type confirmation dialog.
type ShowDeleteAdminFieldTypeDialogMsg struct {
	AdminFieldTypeID types.AdminFieldTypeID
	Label            string
}

// ShowEditAdminFieldTypeDialogMsg requests showing the edit admin field type dialog.
type ShowEditAdminFieldTypeDialogMsg struct {
	AdminFieldType db.AdminFieldTypes
}
