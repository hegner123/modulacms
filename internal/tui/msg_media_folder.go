package tui

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// MEDIA FOLDER MESSAGES
// =============================================================================

// --- Create folder ---

// ShowCreateMediaFolderDialogMsg triggers showing a create media folder dialog.
type ShowCreateMediaFolderDialogMsg struct {
	ParentID types.NullableMediaFolderID // parent folder (Valid=false for root)
}

// CreateMediaFolderRequestMsg triggers folder creation in the DB.
type CreateMediaFolderRequestMsg struct {
	Name     string
	ParentID types.NullableMediaFolderID
}

// MediaFolderCreatedMsg is sent after a folder is successfully created.
type MediaFolderCreatedMsg struct {
	FolderID types.MediaFolderID
	Name     string
}

// --- Rename folder ---

// ShowRenameMediaFolderDialogMsg triggers showing a rename media folder dialog.
type ShowRenameMediaFolderDialogMsg struct {
	FolderID    types.MediaFolderID
	CurrentName string
}

// RenameMediaFolderRequestMsg triggers folder rename in the DB.
type RenameMediaFolderRequestMsg struct {
	FolderID types.MediaFolderID
	NewName  string
}

// MediaFolderRenamedMsg is sent after a folder is successfully renamed.
type MediaFolderRenamedMsg struct {
	FolderID types.MediaFolderID
	NewName  string
}

// --- Delete folder ---

// DeleteMediaFolderContext stores context for a media folder deletion operation.
type DeleteMediaFolderContext struct {
	FolderID types.MediaFolderID
	Name     string
}

// ShowDeleteMediaFolderDialogMsg triggers showing a delete media folder confirmation dialog.
type ShowDeleteMediaFolderDialogMsg struct {
	FolderID types.MediaFolderID
	Name     string
}

// DeleteMediaFolderRequestMsg triggers folder deletion in the DB.
type DeleteMediaFolderRequestMsg struct {
	FolderID types.MediaFolderID
}

// MediaFolderDeletedMsg is sent after a folder is successfully deleted.
type MediaFolderDeletedMsg struct {
	FolderID types.MediaFolderID
}

// --- Move media to folder ---

// ShowMoveMediaToFolderDialogMsg triggers showing a move-to-folder picker dialog.
type ShowMoveMediaToFolderDialogMsg struct {
	MediaID types.MediaID
	Label   string
}

// MoveMediaToFolderRequestMsg triggers moving media to a folder in the DB.
type MoveMediaToFolderRequestMsg struct {
	MediaID  types.MediaID
	FolderID types.NullableMediaFolderID // Valid=false to move to root (unfiled)
}

// MediaMovedToFolderMsg is sent after media is successfully moved to a folder.
type MediaMovedToFolderMsg struct {
	MediaID  types.MediaID
	FolderID types.NullableMediaFolderID
}

// ShowMoveMediaToFolderPickerMsg carries the fetched folder list to the move dialog constructor.
// This is an intermediate message: ShowMoveMediaToFolderDialogMsg triggers a DB fetch,
// which returns this message with the folder data to build the picker UI.
type ShowMoveMediaToFolderPickerMsg struct {
	MediaID types.MediaID
	Label   string
	Folders []db.MediaFolder
}
