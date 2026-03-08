package tui

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// =============================================================================
// ADMIN ROUTES FETCH MESSAGES
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

// =============================================================================
// ADMIN DATATYPES FETCH MESSAGES
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

// =============================================================================
// ADMIN CONTENT FETCH MESSAGES
// =============================================================================

// AdminContentDataFetchMsg requests fetching all admin content data.
type AdminContentDataFetchMsg struct{}

// AdminContentDataFetchResultsMsg returns fetched admin content data.
type AdminContentDataFetchResultsMsg struct {
	Data []db.AdminContentDataTopLevel
}

// AdminContentDataSet sets the admin content data list.
type AdminContentDataSet struct {
	AdminContentData []db.AdminContentDataTopLevel
}

// AdminLoadContentFieldsMsg requests loading admin content fields for display.
type AdminLoadContentFieldsMsg struct {
	Fields []AdminContentFieldDisplay
}

// =============================================================================
// ADMIN FIELD TYPES FETCH MESSAGES
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

// =============================================================================
// ADMIN DIALOG MESSAGES
// =============================================================================

// ShowDeleteAdminRouteDialogMsg requests showing the delete admin route confirmation dialog.
type ShowDeleteAdminRouteDialogMsg struct {
	AdminRouteID types.AdminRouteID
	Title        string
}

// ShowEditAdminRouteDialogMsg requests showing the edit admin route dialog.
type ShowEditAdminRouteDialogMsg struct {
	Route db.AdminRoutes
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

// ShowDeleteFieldTypeDialogMsg requests showing the delete field type confirmation dialog.
type ShowDeleteFieldTypeDialogMsg struct {
	FieldTypeID types.FieldTypeID
	Label       string
}

// ShowEditFieldTypeDialogMsg requests showing the edit field type dialog.
type ShowEditFieldTypeDialogMsg struct {
	FieldType db.FieldTypes
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

// =============================================================================
// ADMIN CONTENT TREE & OPERATIONS MESSAGES
// =============================================================================

// AdminTreeLoadedMsg signals successful admin content tree loading.
type AdminTreeLoadedMsg struct {
	RootNode *tree.Root
	Stats    *tree.LoadStats
}

// AdminContentUpdatedFromDialogMsg signals successful admin content update from dialog.
type AdminContentUpdatedFromDialogMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
}

// AdminRootDatatypesFetchResultsMsg returns fetched root-level admin datatypes.
type AdminRootDatatypesFetchResultsMsg struct {
	RootDatatypes []db.AdminDatatypes
}

// ExistingAdminContentField represents an existing admin content field for edit forms.
type ExistingAdminContentField struct {
	AdminContentFieldID types.AdminContentFieldID
	AdminFieldID        types.AdminFieldID
	Label               string
	Type                string
	Value               string
	ValidationJSON      string
	DataJSON            string
}

// =============================================================================
// ADMIN CONTENT FORM MESSAGES
// =============================================================================

// AdminBuildContentFormMsg requests building an admin content creation form.
type AdminBuildContentFormMsg struct {
	AdminDatatypeID types.AdminDatatypeID
	AdminRouteID    types.AdminRouteID
	Fields          []db.AdminFields
}

// AdminFetchContentForEditMsg requests fetching admin content fields for editing.
type AdminFetchContentForEditMsg struct {
	AdminContentID  types.AdminContentID
	AdminDatatypeID types.AdminDatatypeID
	AdminRouteID    types.AdminRouteID
	Title           string
}

// ShowEditAdminContentFormDialogMsg requests showing the edit admin content form dialog.
type ShowEditAdminContentFormDialogMsg struct {
	AdminContentID  types.AdminContentID
	AdminDatatypeID types.AdminDatatypeID
	AdminRouteID    types.AdminRouteID
	Fields          []ExistingAdminContentField
}

// =============================================================================
// ADMIN TREE OPERATION MESSAGES
// =============================================================================

// AdminReorderSiblingRequestMsg requests reordering admin content siblings.
type AdminReorderSiblingRequestMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
	Direction      string
}

// AdminCopyContentRequestMsg requests copying admin content.
type AdminCopyContentRequestMsg struct {
	SourceID     types.AdminContentID
	AdminRouteID types.AdminRouteID
}

// AdminMoveContentRequestMsg requests moving admin content to a new parent.
type AdminMoveContentRequestMsg struct {
	SourceID     types.AdminContentID
	TargetID     types.AdminContentID
	AdminRouteID types.AdminRouteID
}

// =============================================================================
// ADMIN PUBLISHING MESSAGES
// =============================================================================

// AdminTogglePublishRequestMsg requests toggling admin content publish status.
type AdminTogglePublishRequestMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
}

// =============================================================================
// ADMIN VERSIONING MESSAGES
// =============================================================================

// AdminListVersionsRequestMsg requests listing versions for admin content.
type AdminListVersionsRequestMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
}

// AdminVersionsListedMsg delivers the admin version list.
type AdminVersionsListedMsg struct {
	Versions       []db.AdminContentVersion
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
}

// AdminRestoreVersionRequestMsg requests restoring admin content from a version.
type AdminRestoreVersionRequestMsg struct {
	AdminContentID types.AdminContentID
	VersionID      types.AdminContentVersionID
	AdminRouteID   types.AdminRouteID
}

// =============================================================================
// ADMIN CONTENT FIELD OPERATION MESSAGES (consolidated into unified types)
// =============================================================================
// ContentFieldUpdatedMsg, ContentFieldAddedMsg, ContentFieldDeletedMsg
// in commands_content.go now carry AdminMode bool.

// =============================================================================
// ADMIN CONTENT DIALOG SHOW MESSAGES
// =============================================================================

// ShowDeleteAdminContentDialogMsg requests showing the delete admin content dialog.
type ShowDeleteAdminContentDialogMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
	ContentName    string
	HasChildren    bool
}

// ShowPublishAdminContentDialogMsg requests showing the publish/unpublish dialog.
type ShowPublishAdminContentDialogMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
	Name           string
	IsPublished    bool
}

// ShowRestoreAdminVersionDialogMsg requests showing the restore version dialog.
type ShowRestoreAdminVersionDialogMsg struct {
	AdminContentID types.AdminContentID
	VersionID      types.AdminContentVersionID
	AdminRouteID   types.AdminRouteID
	VersionNumber  int64
}

// ShowDeleteAdminContentFieldDialogMsg requests showing the delete admin content field dialog.
type ShowDeleteAdminContentFieldDialogMsg struct {
	AdminContentFieldID types.AdminContentFieldID
	AdminContentID      types.AdminContentID
	AdminRouteID        types.AdminRouteID
	AdminDatatypeID     types.NullableAdminDatatypeID
	Label               string
}

// ShowMoveAdminContentDialogMsg requests showing the move admin content dialog.
type ShowMoveAdminContentDialogMsg struct {
	SourceNode   *tree.Node
	AdminRouteID types.AdminRouteID
	Targets      []ParentOption
}

// =============================================================================
// ADMIN CONTENT CONFIRMED ACTION MESSAGES
// =============================================================================

// ConfirmedDeleteAdminContentMsg signals user confirmed admin content deletion.
type ConfirmedDeleteAdminContentMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
}

// ConfirmedPublishAdminContentMsg signals user confirmed admin content publishing.
type ConfirmedPublishAdminContentMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
}

// ConfirmedUnpublishAdminContentMsg signals user confirmed admin content unpublishing.
type ConfirmedUnpublishAdminContentMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
}

// ConfirmedRestoreAdminVersionMsg signals user confirmed admin version restore.
type ConfirmedRestoreAdminVersionMsg struct {
	AdminContentID types.AdminContentID
	VersionID      types.AdminContentVersionID
	AdminRouteID   types.AdminRouteID
}

// ConfirmedDeleteAdminContentFieldMsg signals user confirmed admin content field deletion.
type ConfirmedDeleteAdminContentFieldMsg struct {
	AdminContentFieldID types.AdminContentFieldID
	AdminContentID      types.AdminContentID
	AdminRouteID        types.AdminRouteID
	AdminDatatypeID     types.NullableAdminDatatypeID
}

// =============================================================================
// ADMIN DIALOG CONTEXT STRUCTS
// =============================================================================
// DeleteAdminContentContext, PublishAdminContentContext, RestoreAdminVersionContext,
// and DeleteAdminContentFieldContext are now consolidated into their regular
// counterparts with AdminMode bool (see update_dialog_helpers.go and
// form_dialog_constructors.go).

// MoveAdminContentContext stores context for the move admin content dialog.
type MoveAdminContentContext struct {
	SourceNode   *tree.Node
	AdminRouteID types.AdminRouteID
}

// editAdminSingleFieldCtx stores context for editing a single admin content field.
type editAdminSingleFieldCtx struct {
	AdminContentFieldID types.AdminContentFieldID
	AdminContentID      types.AdminContentID
	AdminFieldID        types.AdminFieldID
	AdminRouteID        types.AdminRouteID
	AdminDatatypeID     types.NullableAdminDatatypeID
	Label               string
	Type                string
	Value               string
}

// addAdminContentFieldCtx stores context for adding an admin content field.
type addAdminContentFieldCtx struct {
	AdminContentID  types.AdminContentID
	AdminRouteID    types.AdminRouteID
	AdminDatatypeID types.NullableAdminDatatypeID
}

// =============================================================================
// ADMIN CONTENT FORM DIALOG ACCEPT/CANCEL MESSAGES
// =============================================================================

// AdminContentFormDialogAcceptMsg is sent when an admin content form dialog is accepted.
type AdminContentFormDialogAcceptMsg struct {
	Action      FormDialogAction
	DatatypeID  types.AdminDatatypeID
	RouteID     types.AdminRouteID
	ContentID   types.AdminContentID         // For edit mode
	ParentID    types.NullableAdminContentID // For child creation
	FieldValues map[types.AdminFieldID]string
}

// AdminContentFormDialogCancelMsg is sent when an admin content form dialog is cancelled.
type AdminContentFormDialogCancelMsg struct{}

// CreateAdminRouteWithContentRequestMsg requests creating an admin route with initial content.
type CreateAdminRouteWithContentRequestMsg struct {
	Title           string
	Slug            string
	AdminDatatypeID string
}

// AdminRouteWithContentCreatedMsg signals that an admin route and content were created.
type AdminRouteWithContentCreatedMsg struct {
	AdminRouteID       types.AdminRouteID
	AdminContentDataID types.AdminContentID
	AdminDatatypeID    types.AdminDatatypeID
	Title              string
	Slug               string
}
