package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	utility "github.com/hegner123/modulacms/internal/utility"
)

// HandleCreateMediaFolder creates a media folder in the database.
func (m Model) HandleCreateMediaFolder(msg CreateMediaFolderRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create folder: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Validate folder name
		if err := d.ValidateMediaFolderName(msg.Name, msg.ParentID); err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Invalid folder name: %v", err),
			}
		}

		params := db.CreateMediaFolderParams{
			Name:         msg.Name,
			ParentID:     msg.ParentID,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		result, err := d.CreateMediaFolder(ctx, ac, params)
		if err != nil {
			logger.Ferror("Failed to create media folder", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create folder: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Media folder created: %s (%s)", result.Name, result.FolderID))
		return MediaFolderCreatedMsg{
			FolderID: result.FolderID,
			Name:     result.Name,
		}
	}
}

// HandleRenameMediaFolder renames a media folder in the database.
func (m Model) HandleRenameMediaFolder(msg RenameMediaFolderRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot rename folder: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Fetch existing folder to preserve all fields
		existing, err := d.GetMediaFolder(msg.FolderID)
		if err != nil {
			logger.Ferror("Failed to get folder for rename", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get folder for rename: %v", err),
			}
		}

		// Validate new name within the same parent
		if err := d.ValidateMediaFolderName(msg.NewName, existing.ParentID); err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Invalid folder name: %v", err),
			}
		}

		params := db.UpdateMediaFolderParams{
			FolderID:     msg.FolderID,
			Name:         msg.NewName,
			ParentID:     existing.ParentID,     // PRESERVED
			DateModified: types.TimestampNow(),
		}

		_, err = d.UpdateMediaFolder(ctx, ac, params)
		if err != nil {
			logger.Ferror("Failed to rename media folder", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to rename folder: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Media folder renamed: %s -> %s", msg.FolderID, msg.NewName))
		return MediaFolderRenamedMsg{
			FolderID: msg.FolderID,
			NewName:  msg.NewName,
		}
	}
}

// HandleDeleteMediaFolder deletes a media folder from the database.
// Fails if the folder contains any media items or subfolders.
func (m Model) HandleDeleteMediaFolder(msg DeleteMediaFolderRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot delete folder: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Check for child folders
		children, err := d.ListMediaFoldersByParent(msg.FolderID)
		if err != nil {
			logger.Ferror("Failed to check folder children", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to check folder contents: %v", err),
			}
		}
		if children != nil && len(*children) > 0 {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot delete folder: it contains subfolders. Delete or move them first.",
			}
		}

		// Check for media in folder
		mediaCount, err := d.CountMediaByFolder(types.NullableMediaFolderID{ID: msg.FolderID, Valid: true})
		if err != nil {
			logger.Ferror("Failed to check folder media", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to check folder contents: %v", err),
			}
		}
		if mediaCount != nil && *mediaCount > 0 {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot delete folder: it contains media items. Delete or move them first.",
			}
		}

		if err := d.DeleteMediaFolder(ctx, ac, msg.FolderID); err != nil {
			logger.Ferror("Failed to delete media folder", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete folder: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Media folder deleted: %s", msg.FolderID))
		return MediaFolderDeletedMsg{FolderID: msg.FolderID}
	}
}

// HandleMoveMediaToFolder moves a media item to a folder.
func (m Model) HandleMoveMediaToFolder(msg MoveMediaToFolderRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot move media: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		params := db.MoveMediaToFolderParams{
			MediaID:      msg.MediaID,
			FolderID:     msg.FolderID,
			DateModified: types.TimestampNow(),
		}

		if err := d.MoveMediaToFolder(ctx, ac, params); err != nil {
			logger.Ferror("Failed to move media to folder", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to move media: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Media %s moved to folder %s", msg.MediaID, msg.FolderID))
		return MediaMovedToFolderMsg{
			MediaID:  msg.MediaID,
			FolderID: msg.FolderID,
		}
	}
}
