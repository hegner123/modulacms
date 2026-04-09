package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// HandleDeleteSession processes the session deletion (revocation) request.
func (m Model) HandleDeleteSession(msg DeleteSessionRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Revoking session: %s", msg.SessionID))

		if err := d.DeleteSession(ctx, ac, msg.SessionID); err != nil {
			logger.Ferror("failed to revoke session", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("failed to revoke session: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("session revoked successfully: %s", msg.SessionID))
		return SessionDeletedMsg{SessionID: msg.SessionID}
	}
}
