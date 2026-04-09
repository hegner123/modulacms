package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// HandleUnlinkOauth processes the OAuth connection unlink request.
func (m Model) HandleUnlinkOauth(msg UnlinkOauthRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Unlinking OAuth connection: %s", msg.UserOauthID))

		if err := d.DeleteUserOauth(ctx, ac, msg.UserOauthID); err != nil {
			logger.Ferror("failed to unlink OAuth connection", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("failed to unlink OAuth: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("OAuth connection unlinked: %s", msg.UserOauthID))
		return UserOauthDeletedMsg{UserID: msg.UserID, UserOauthID: msg.UserOauthID}
	}
}
