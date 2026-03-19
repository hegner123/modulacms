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

// HandleCreateAdminMediaFolder creates an admin media folder in the database.
func (m Model) HandleCreateAdminMediaFolder(msg CreateAdminMediaFolderRequestMsg) tea.Cmd {
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
				Message: "Cannot create admin folder: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Validate folder name
		if err := d.ValidateAdminMediaFolderName(msg.Name, msg.ParentID); err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Invalid folder name: %v", err),
			}
		}

		params := db.CreateAdminMediaFolderParams{
			Name:         msg.Name,
			ParentID:     msg.ParentID,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		result, err := d.CreateAdminMediaFolder(ctx, ac, params)
		if err != nil {
			logger.Ferror("Failed to create admin media folder", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create folder: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Admin media folder created: %s (%s)", result.Name, result.AdminFolderID))
		return AdminMediaFolderCreatedMsg{
			FolderID: result.AdminFolderID,
			Name:     result.Name,
		}
	}
}

// HandleRenameAdminMediaFolder renames an admin media folder in the database.
func (m Model) HandleRenameAdminMediaFolder(msg RenameAdminMediaFolderRequestMsg) tea.Cmd {
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
				Message: "Cannot rename admin folder: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Fetch existing folder to preserve all fields
		existing, err := d.GetAdminMediaFolder(msg.FolderID)
		if err != nil {
			logger.Ferror("Failed to get admin folder for rename", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get folder for rename: %v", err),
			}
		}

		// Validate new name within the same parent
		if err := d.ValidateAdminMediaFolderName(msg.NewName, existing.ParentID); err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Invalid folder name: %v", err),
			}
		}

		params := db.UpdateAdminMediaFolderParams{
			AdminFolderID: msg.FolderID,
			Name:          msg.NewName,
			ParentID:      existing.ParentID, // PRESERVED
			DateModified:  types.TimestampNow(),
		}

		_, err = d.UpdateAdminMediaFolder(ctx, ac, params)
		if err != nil {
			logger.Ferror("Failed to rename admin media folder", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to rename folder: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Admin media folder renamed: %s -> %s", msg.FolderID, msg.NewName))
		return AdminMediaFolderRenamedMsg{
			FolderID: msg.FolderID,
			NewName:  msg.NewName,
		}
	}
}

// HandleDeleteAdminMediaFolder deletes an admin media folder from the database.
// Fails if the folder contains any media items or subfolders.
func (m Model) HandleDeleteAdminMediaFolder(msg DeleteAdminMediaFolderRequestMsg) tea.Cmd {
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
				Message: "Cannot delete admin folder: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Check for child folders
		children, err := d.ListAdminMediaFoldersByParent(msg.FolderID)
		if err != nil {
			logger.Ferror("Failed to check admin folder children", err)
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
		mediaCount, err := d.CountAdminMediaByFolder(types.NullableAdminMediaFolderID{ID: msg.FolderID, Valid: true})
		if err != nil {
			logger.Ferror("Failed to check admin folder media", err)
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

		if err := d.DeleteAdminMediaFolder(ctx, ac, msg.FolderID); err != nil {
			logger.Ferror("Failed to delete admin media folder", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete folder: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Admin media folder deleted: %s", msg.FolderID))
		return AdminMediaFolderDeletedMsg{FolderID: msg.FolderID}
	}
}

// HandleMoveAdminMediaToFolder moves an admin media item to a folder.
func (m Model) HandleMoveAdminMediaToFolder(msg MoveAdminMediaToFolderRequestMsg) tea.Cmd {
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
				Message: "Cannot move admin media: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		params := db.MoveAdminMediaToFolderParams{
			AdminMediaID: msg.AdminMediaID,
			FolderID:     msg.FolderID,
			DateModified: types.TimestampNow(),
		}

		if err := d.MoveAdminMediaToFolder(ctx, ac, params); err != nil {
			logger.Ferror("Failed to move admin media to folder", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to move media: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Admin media %s moved to folder %s", msg.AdminMediaID, msg.FolderID))
		return AdminMediaMovedToFolderMsg{
			AdminMediaID: msg.AdminMediaID,
			FolderID:     msg.FolderID,
		}
	}
}

// HandleDeleteAdminMedia deletes an admin media item from the database.
func (m Model) HandleDeleteAdminMedia(msg DeleteAdminMediaRequestMsg) tea.Cmd {
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
				Message: "Cannot delete admin media: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		if err := d.DeleteAdminMedia(ctx, ac, msg.AdminMediaID); err != nil {
			logger.Ferror("Failed to delete admin media", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete admin media: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Admin media deleted: %s", msg.AdminMediaID))
		return AdminMediaDeletedMsg{AdminMediaID: msg.AdminMediaID}
	}
}
