package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// HandleCreateRole processes role creation.
func (m Model) HandleCreateRole(msg CreateRoleFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "configuration not loaded"}
		}
	}

	return func() tea.Msg {
		if msg.Label == "" {
			return ActionResultMsg{Title: "Validation Error", Message: "Role name is required"}
		}

		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		result, err := d.CreateRole(ctx, ac, db.CreateRoleParams{
			Label:           msg.Label,
			SystemProtected: false,
		})
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to create role: %v", err)}
		}

		return RoleCreatedFromDialogMsg{RoleID: result.RoleID, Label: result.Label}
	}
}

// HandleUpdateRole processes role update.
func (m Model) HandleUpdateRole(msg UpdateRoleFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "configuration not loaded"}
		}
	}

	return func() tea.Msg {
		if msg.Label == "" {
			return ActionResultMsg{Title: "Validation Error", Message: "Role name is required"}
		}

		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		roleID := types.RoleID(msg.RoleID)

		// Check system protection
		existing, err := d.GetRole(roleID)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to get role: %v", err)}
		}
		if existing.SystemProtected && msg.Label != existing.Label {
			return ActionResultMsg{Title: "Forbidden", Message: "Cannot rename system-protected role"}
		}

		if _, err := d.UpdateRole(ctx, ac, db.UpdateRoleParams{
			RoleID:          roleID,
			Label:           msg.Label,
			SystemProtected: existing.SystemProtected,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to update role: %v", err)}
		}

		return RoleUpdatedFromDialogMsg{RoleID: roleID, Label: msg.Label}
	}
}

// HandleDeleteRole processes role deletion.
func (m Model) HandleDeleteRole(msg DeleteRoleRequestMsg) tea.Cmd {
	logger := m.Logger
	if logger == nil {
		logger = utility.DefaultLogger
	}
	cfg := m.Config
	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		// Check system protection
		existing, err := d.GetRole(msg.RoleID)
		if err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to get role: %v", err)}
		}
		if existing.SystemProtected {
			return ActionResultMsg{Title: "Forbidden", Message: "Cannot delete system-protected role"}
		}

		logger.Finfo(fmt.Sprintf("Deleting role: %s (%s)", msg.RoleID, existing.Label))

		if err := d.DeleteRole(ctx, ac, msg.RoleID); err != nil {
			logger.Ferror("failed to delete role", err)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to delete role: %v", err)}
		}

		logger.Finfo(fmt.Sprintf("Role deleted: %s", msg.RoleID))
		return RoleDeletedMsg{RoleID: msg.RoleID}
	}
}
