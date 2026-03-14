package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// =============================================================================
// CREATE WEBHOOK HANDLER
// =============================================================================

// HandleCreateWebhook processes the webhook creation request.
func (m Model) HandleCreateWebhook(msg CreateWebhookFromDialogRequestMsg) tea.Cmd {
	authorID := m.UserID
	cfg := m.Config

	if authorID.IsZero() {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create webhook: no user is logged in",
			}
		}
	}

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot create webhook: configuration not loaded",
			}
		}
	}

	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, authorID)

		// Parse comma-separated events
		events := parseCommaSeparated(msg.Events)

		// Validate required fields
		if msg.Name == "" {
			return ActionResultMsg{
				Title:   "Validation Error",
				Message: "Webhook name is required",
			}
		}
		if msg.URL == "" {
			return ActionResultMsg{
				Title:   "Validation Error",
				Message: "Webhook URL is required",
			}
		}

		params := db.CreateWebhookParams{
			Name:         msg.Name,
			URL:          msg.URL,
			Secret:       msg.Secret,
			Events:       events,
			IsActive:     msg.IsActive,
			Headers:      make(map[string]string),
			AuthorID:     authorID,
			DateCreated:  types.TimestampNow(),
			DateModified: types.TimestampNow(),
		}

		webhook, err := d.CreateWebhook(ctx, ac, params)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to create webhook: %v", err),
			}
		}
		if webhook == nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Failed to create webhook in database",
			}
		}

		return WebhookCreatedMsg{
			WebhookID: webhook.WebhookID,
			Name:      webhook.Name,
		}
	}
}

// =============================================================================
// UPDATE WEBHOOK HANDLER
// =============================================================================

// HandleUpdateWebhook processes the webhook update request.
func (m Model) HandleUpdateWebhook(msg UpdateWebhookFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{
				Title:   "Error",
				Message: "Cannot update webhook: configuration not loaded",
			}
		}
	}

	userID := m.UserID
	return func() tea.Msg {
		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		webhookID := types.WebhookID(msg.WebhookID)

		// Fetch existing webhook to preserve unchanged fields
		existing, err := d.GetWebhook(webhookID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to get webhook for update: %v", err),
			}
		}

		// Parse comma-separated events
		events := parseCommaSeparated(msg.Events)

		// Validate required fields
		if msg.Name == "" {
			return ActionResultMsg{
				Title:   "Validation Error",
				Message: "Webhook name is required",
			}
		}
		if msg.URL == "" {
			return ActionResultMsg{
				Title:   "Validation Error",
				Message: "Webhook URL is required",
			}
		}

		params := db.UpdateWebhookParams{
			WebhookID:    webhookID,
			Name:         msg.Name,
			URL:          msg.URL,
			Secret:       msg.Secret,
			Events:       events,
			IsActive:     msg.IsActive,
			Headers:      existing.Headers,     // PRESERVED
			DateCreated:  existing.DateCreated, // PRESERVED
			DateModified: types.TimestampNow(),
		}

		if err := d.UpdateWebhook(ctx, ac, params); err != nil {
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to update webhook: %v", err),
			}
		}

		return WebhookUpdatedMsg{
			WebhookID: webhookID,
			Name:      msg.Name,
		}
	}
}

// =============================================================================
// DELETE WEBHOOK HANDLER
// =============================================================================

// HandleDeleteWebhook processes the webhook deletion request.
func (m Model) HandleDeleteWebhook(msg DeleteWebhookRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Deleting webhook: %s", msg.WebhookID))

		if err := d.DeleteWebhook(ctx, ac, msg.WebhookID); err != nil {
			logger.Ferror("Failed to delete webhook", err)
			return ActionResultMsg{
				Title:   "Error",
				Message: fmt.Sprintf("Failed to delete webhook: %v", err),
			}
		}

		logger.Finfo(fmt.Sprintf("Webhook deleted successfully: %s", msg.WebhookID))
		return WebhookDeletedMsg{WebhookID: msg.WebhookID}
	}
}

// =============================================================================
// HELPERS
// =============================================================================

// parseCommaSeparated splits a comma-separated string into trimmed, non-empty parts.
func parseCommaSeparated(s string) []string {
	raw := strings.Split(s, ",")
	result := make([]string, 0, len(raw))
	for _, part := range raw {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
