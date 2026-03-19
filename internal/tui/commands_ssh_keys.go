package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// HandleDeleteSshKey processes the SSH key deletion request.
func (m Model) HandleDeleteSshKey(msg DeleteSshKeyRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Removing SSH key: %s", msg.SshKeyID))

		if err := d.DeleteUserSshKey(ctx, ac, msg.SshKeyID); err != nil {
			logger.Ferror("Failed to remove SSH key", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to remove SSH key: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("SSH key removed: %s", msg.SshKeyID))
		return UserSshKeyDeletedMsg{UserID: msg.UserID, SshKeyID: msg.SshKeyID}
	}
}
