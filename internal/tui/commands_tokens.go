package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// =============================================================================
// CREATE TOKEN HANDLER
// =============================================================================

// HandleCreateToken processes the token creation request.
// Generates a random API token via the service layer and returns the
// one-time raw token value that must be shown to the user immediately.
func (m Model) HandleCreateToken(msg CreateTokenFromDialogRequestMsg) tea.Cmd {
	authorID := m.UserID
	cfg := m.Config

	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create token: no user is logged in",
			}
		}
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create token: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		tokenType := msg.TokenType
		if tokenType == "" {
			tokenType = types.TokenTypeAPIKey
		}

		input := service.CreateTokenInput{
			UserID:    types.NullableUserID{ID: authorID, Valid: true},
			TokenType: tokenType,
			Expiry:    types.NewTimestamp(types.TimestampNow().UTC().AddDate(1, 0, 0)),
		}

		tokenSvc := service.NewTokenService(d)
		result, err := tokenSvc.CreateToken(ctx, ac, input)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create token: %v", err),
			}
		}

		return TokenCreatedFromDialogMsg{
			TokenID:  result.Token.ID,
			RawToken: result.RawToken,
		}
	}
}

// =============================================================================
// DELETE TOKEN HANDLER
// =============================================================================

// HandleDeleteToken processes the token deletion request.
func (m Model) HandleDeleteToken(msg DeleteTokenRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Deleting token: %s", msg.TokenID))

		if err := d.DeleteToken(ctx, ac, msg.TokenID); err != nil {
			logger.Ferror("Failed to delete token", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete token: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Token deleted successfully: %s", msg.TokenID))
		return TokenDeletedMsg{TokenID: msg.TokenID}
	}
}
