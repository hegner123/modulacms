package tui

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// ADMIN MEDIA FOLDER MESSAGES
// =============================================================================

// --- Create folder ---

// ShowCreateAdminMediaFolderDialogMsg triggers showing a create admin media folder dialog.
type ShowCreateAdminMediaFolderDialogMsg struct {
	ParentID types.NullableAdminMediaFolderID // parent folder (Valid=false for root)
}

// CreateAdminMediaFolderRequestMsg triggers folder creation in the DB.
type CreateAdminMediaFolderRequestMsg struct {
	Name     string
	ParentID types.NullableAdminMediaFolderID
}

// AdminMediaFolderCreatedMsg is sent after a folder is successfully created.
type AdminMediaFolderCreatedMsg struct {
	FolderID types.AdminMediaFolderID
	Name     string
}

// --- Rename folder ---

// ShowRenameAdminMediaFolderDialogMsg triggers showing a rename admin media folder dialog.
type ShowRenameAdminMediaFolderDialogMsg struct {
	FolderID    types.AdminMediaFolderID
	CurrentName string
}

// RenameAdminMediaFolderRequestMsg triggers folder rename in the DB.
type RenameAdminMediaFolderRequestMsg struct {
	FolderID types.AdminMediaFolderID
	NewName  string
}

// AdminMediaFolderRenamedMsg is sent after a folder is successfully renamed.
type AdminMediaFolderRenamedMsg struct {
	FolderID types.AdminMediaFolderID
	NewName  string
}

// --- Delete folder ---

// DeleteAdminMediaFolderContext stores context for an admin media folder deletion operation.
type DeleteAdminMediaFolderContext struct {
	FolderID types.AdminMediaFolderID
	Name     string
}

// ShowDeleteAdminMediaFolderDialogMsg triggers showing a delete admin media folder confirmation dialog.
type ShowDeleteAdminMediaFolderDialogMsg struct {
	FolderID types.AdminMediaFolderID
	Name     string
}

// DeleteAdminMediaFolderRequestMsg triggers folder deletion in the DB.
type DeleteAdminMediaFolderRequestMsg struct {
	FolderID types.AdminMediaFolderID
}

// AdminMediaFolderDeletedMsg is sent after a folder is successfully deleted.
type AdminMediaFolderDeletedMsg struct {
	FolderID types.AdminMediaFolderID
}

// --- Delete admin media ---

// DeleteAdminMediaContext stores context for an admin media deletion operation.
type DeleteAdminMediaContext struct {
	AdminMediaID types.AdminMediaID
	Label        string
}

// ShowDeleteAdminMediaDialogMsg triggers showing a delete admin media confirmation dialog.
type ShowDeleteAdminMediaDialogMsg struct {
	AdminMediaID types.AdminMediaID
	Label        string
}

// DeleteAdminMediaRequestMsg triggers admin media deletion.
type DeleteAdminMediaRequestMsg struct {
	AdminMediaID types.AdminMediaID
}

// AdminMediaDeletedMsg is sent after an admin media item is successfully deleted.
type AdminMediaDeletedMsg struct {
	AdminMediaID types.AdminMediaID
}

// --- Move admin media to folder ---

// ShowMoveAdminMediaToFolderDialogMsg triggers showing a move-to-folder picker dialog.
type ShowMoveAdminMediaToFolderDialogMsg struct {
	AdminMediaID types.AdminMediaID
	Label        string
}

// MoveAdminMediaToFolderRequestMsg triggers moving admin media to a folder in the DB.
type MoveAdminMediaToFolderRequestMsg struct {
	AdminMediaID types.AdminMediaID
	FolderID     types.NullableAdminMediaFolderID // Valid=false to move to root (unfiled)
}

// AdminMediaMovedToFolderMsg is sent after admin media is successfully moved to a folder.
type AdminMediaMovedToFolderMsg struct {
	AdminMediaID types.AdminMediaID
	FolderID     types.NullableAdminMediaFolderID
}

// ShowMoveAdminMediaToFolderPickerMsg carries the fetched folder list to the move dialog constructor.
type ShowMoveAdminMediaToFolderPickerMsg struct {
	AdminMediaID types.AdminMediaID
	Label        string
	Folders      []db.AdminMediaFolder
}
