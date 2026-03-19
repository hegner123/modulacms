package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// DatatypeUpdateSaveMsg requests saving datatype updates.
type DatatypeUpdateSaveMsg struct {
	DatatypeID types.DatatypeID
	Parent     string
	Name       string
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

// CmsDefineDatatypeLoadMsg requests loading the datatype definition form.
type CmsDefineDatatypeLoadMsg struct{}

// CmsDefineDatatypeReadyMsg signals that datatype definition is ready.
type CmsDefineDatatypeReadyMsg struct{}

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

// CmsGetDatatypeParentOptionsMsg requests fetching parent datatype options.
type CmsGetDatatypeParentOptionsMsg struct {
	Admin bool
}

// ContentCreatedMsg signals successful content creation.
// AdminMode selects admin vs regular reload/display paths.
type ContentCreatedMsg struct {
	ContentID  types.ContentID
	RouteID    types.RouteID
	FieldCount int
	AdminMode  bool
}

// ContentCreatedWithErrorsMsg signals content creation with partial field failures.
type ContentCreatedWithErrorsMsg struct {
	ContentDataID types.ContentID
	RouteID       types.RouteID
	CreatedFields int
	FailedFields  []types.FieldID
}

// BuildContentFormMsg requests building a content creation form.
type BuildContentFormMsg struct {
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
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
	AdminMode bool
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
	AdminMode       bool
}

// TogglePublishRequestMsg requests toggling content publish status.
// Now triggers a confirmation dialog instead of direct toggle.
type TogglePublishRequestMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
}

// ConfirmedPublishMsg signals user confirmed the publish action.
type ConfirmedPublishMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
}

// ConfirmedUnpublishMsg signals user confirmed the unpublish action.
type ConfirmedUnpublishMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
}

// PublishCompletedMsg signals successful snapshot-based publish.
type PublishCompletedMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	AdminMode bool
}

// UnpublishCompletedMsg signals successful unpublish.
type UnpublishCompletedMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	AdminMode bool
}

// ContentPublishToggledMsg signals successful publish status toggle.
// Kept for backward compatibility; new flow uses PublishCompletedMsg / UnpublishCompletedMsg.
type ContentPublishToggledMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	NewStatus types.ContentStatus
}

// ListVersionsRequestMsg requests listing versions for a content item.
type ListVersionsRequestMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
}

// VersionsListedMsg delivers the version list to the model.
type VersionsListedMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	Versions  []db.ContentVersion
}

// RestoreVersionRequestMsg requests restoring content from a specific version.
type RestoreVersionRequestMsg struct {
	ContentID types.ContentID
	VersionID types.ContentVersionID
	RouteID   types.RouteID
}

// ConfirmedRestoreVersionMsg signals user confirmed the restore action.
type ConfirmedRestoreVersionMsg struct {
	ContentID types.ContentID
	VersionID types.ContentVersionID
	RouteID   types.RouteID
}

// VersionRestoredMsg signals successful version restore.
type VersionRestoredMsg struct {
	ContentID      types.ContentID
	RouteID        types.RouteID
	FieldsRestored int
	AdminMode      bool
}

// TreeLoadedMsg signals successful content tree loading.
type TreeLoadedMsg struct {
	RouteID  types.RouteID
	Stats    *tree.LoadStats
	RootNode *tree.Root
}

// MediaUploadStartMsg triggers the async upload pipeline
type MediaUploadStartMsg struct {
	FilePath string
}

// MediaUploadedMsg signals upload completed successfully
type MediaUploadedMsg struct {
	Name string
}

// MediaUploadProgressMsg carries upload progress for the status display.
type MediaUploadProgressMsg struct {
	BytesSent  int64
	Total      int64
	ProgressCh <-chan tea.Msg // channel for the next progress message
}

// =============================================================================
// ADMIN ROUTE CRUD MESSAGES
// =============================================================================

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

// =============================================================================
// ADMIN DATATYPE CRUD MESSAGES
// =============================================================================

// CreateAdminDatatypeFromDialogRequestMsg requests creating an admin datatype from dialog input.
type CreateAdminDatatypeFromDialogRequestMsg struct {
	Name     string
	Label    string
	Type     string
	ParentID string
}

// UpdateAdminDatatypeFromDialogRequestMsg requests updating an admin datatype from dialog input.
type UpdateAdminDatatypeFromDialogRequestMsg struct {
	AdminDatatypeID string
	Name            string
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

// =============================================================================
// ADMIN FIELD CRUD MESSAGES
// =============================================================================

// CreateAdminFieldFromDialogRequestMsg requests creating an admin field from dialog input.
type CreateAdminFieldFromDialogRequestMsg struct {
	Name            string
	Label           string
	Type            string
	AdminDatatypeID types.AdminDatatypeID
}

// UpdateAdminFieldFromDialogRequestMsg requests updating an admin field from dialog input.
type UpdateAdminFieldFromDialogRequestMsg struct {
	AdminFieldID string
	Name         string
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

// DeleteAdminContentRequestMsg requests deleting admin content.
type DeleteAdminContentRequestMsg struct {
	AdminContentID types.AdminContentID
	AdminRouteID   types.AdminRouteID
}

// AdminContentFieldDisplay represents an admin content field for right panel display.
type AdminContentFieldDisplay struct {
	AdminContentFieldID types.AdminContentFieldID
	AdminFieldID        types.AdminFieldID
	Label               string
	Type                string
	Value               string
	ValidationJSON      string
	DataJSON            string
}

// =============================================================================
// FIELD TYPE CRUD MESSAGES
// =============================================================================

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

// =============================================================================
// ADMIN FIELD TYPE CRUD MESSAGES
// =============================================================================

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

// =============================================================================
// VALIDATION CRUD MESSAGES
// =============================================================================

// CreateValidationFromDialogRequestMsg requests creating a validation from dialog input.
type CreateValidationFromDialogRequestMsg struct {
	Name        string
	Description string
}

// UpdateValidationFromDialogRequestMsg requests updating a validation from dialog input.
type UpdateValidationFromDialogRequestMsg struct {
	ValidationID string
	Name         string
	Description  string
}

// DeleteValidationRequestMsg requests deleting a validation.
type DeleteValidationRequestMsg struct {
	ValidationID types.ValidationID
}

// ValidationCreatedFromDialogMsg signals successful validation creation.
type ValidationCreatedFromDialogMsg struct {
	ValidationID types.ValidationID
	Name         string
}

// ValidationUpdatedFromDialogMsg signals successful validation update.
type ValidationUpdatedFromDialogMsg struct {
	ValidationID types.ValidationID
	Name         string
}

// ValidationDeletedMsg signals successful validation deletion.
type ValidationDeletedMsg struct {
	ValidationID types.ValidationID
}

// =============================================================================
// ADMIN VALIDATION CRUD MESSAGES
// =============================================================================

// CreateAdminValidationFromDialogRequestMsg requests creating an admin validation from dialog input.
type CreateAdminValidationFromDialogRequestMsg struct {
	Name        string
	Description string
}

// UpdateAdminValidationFromDialogRequestMsg requests updating an admin validation from dialog input.
type UpdateAdminValidationFromDialogRequestMsg struct {
	AdminValidationID string
	Name              string
	Description       string
}

// DeleteAdminValidationRequestMsg requests deleting an admin validation.
type DeleteAdminValidationRequestMsg struct {
	AdminValidationID types.AdminValidationID
}

// AdminValidationCreatedFromDialogMsg signals successful admin validation creation.
type AdminValidationCreatedFromDialogMsg struct {
	AdminValidationID types.AdminValidationID
	Name              string
}

// AdminValidationUpdatedFromDialogMsg signals successful admin validation update.
type AdminValidationUpdatedFromDialogMsg struct {
	AdminValidationID types.AdminValidationID
	Name              string
}

// AdminValidationDeletedMsg signals successful admin validation deletion.
type AdminValidationDeletedMsg struct {
	AdminValidationID types.AdminValidationID
}
