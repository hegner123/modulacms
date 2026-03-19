package tui

import (
	"context"
	"fmt"
	"strconv"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// HandleCreateMediaDimension processes creation of a media dimension preset.
func (m Model) HandleCreateMediaDimension(msg CreateMediaDimensionFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	return func() tea.Msg {
		if msg.Label == "" {
			return ActionResultMsg{Title: "Validation Error", Message: "Label is required"}
		}
		width, err := strconv.ParseInt(msg.Width, 10, 64)
		if err != nil || width <= 0 {
			return ActionResultMsg{Title: "Validation Error", Message: "Width must be a positive integer"}
		}
		height, err := strconv.ParseInt(msg.Height, 10, 64)
		if err != nil || height <= 0 {
			return ActionResultMsg{Title: "Validation Error", Message: "Height must be a positive integer"}
		}

		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		params := db.CreateMediaDimensionParams{
			Label:       db.NewNullString(msg.Label),
			Width:       types.NewNullableInt64(width),
			Height:      types.NewNullableInt64(height),
			AspectRatio: db.NewNullString(msg.AspectRatio),
		}
		if msg.AspectRatio == "" {
			params.AspectRatio = db.NullString{}
		}

		result, createErr := d.CreateMediaDimension(ctx, ac, params)
		if createErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to create dimension: %v", createErr)}
		}

		label := msg.Label
		if result != nil && result.Label.Valid {
			label = result.Label.String
		}
		return MediaDimensionCreatedMsg{MdID: result.MdID, Label: label}
	}
}

// HandleUpdateMediaDimension processes update of a media dimension preset.
func (m Model) HandleUpdateMediaDimension(msg UpdateMediaDimensionFromDialogRequestMsg) tea.Cmd {
	cfg := m.Config
	userID := m.UserID

	if cfg == nil {
		return func() tea.Msg {
			return ActionResultMsg{Title: "Error", Message: "Configuration not loaded"}
		}
	}

	return func() tea.Msg {
		width, err := strconv.ParseInt(msg.Width, 10, 64)
		if err != nil || width <= 0 {
			return ActionResultMsg{Title: "Validation Error", Message: "Width must be a positive integer"}
		}
		height, err := strconv.ParseInt(msg.Height, 10, 64)
		if err != nil || height <= 0 {
			return ActionResultMsg{Title: "Validation Error", Message: "Height must be a positive integer"}
		}

		d := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		params := db.UpdateMediaDimensionParams{
			MdID:        msg.MdID,
			Label:       db.NewNullString(msg.Label),
			Width:       types.NewNullableInt64(width),
			Height:      types.NewNullableInt64(height),
			AspectRatio: db.NewNullString(msg.AspectRatio),
		}
		if msg.AspectRatio == "" {
			params.AspectRatio = db.NullString{}
		}

		if _, updateErr := d.UpdateMediaDimension(ctx, ac, params); updateErr != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to update dimension: %v", updateErr)}
		}

		return MediaDimensionUpdatedMsg{MdID: msg.MdID, Label: msg.Label}
	}
}

// HandleDeleteMediaDimension processes deletion of a media dimension preset.
func (m Model) HandleDeleteMediaDimension(msg DeleteMediaDimensionRequestMsg) tea.Cmd {
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

		logger.Finfo(fmt.Sprintf("Deleting media dimension: %s", msg.MdID))

		if err := d.DeleteMediaDimension(ctx, ac, msg.MdID); err != nil {
			logger.Ferror("Failed to delete media dimension", err)
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("Failed to delete dimension: %v", err)}
		}

		logger.Finfo(fmt.Sprintf("Media dimension deleted: %s", msg.MdID))
		return MediaDimensionDeletedMsg{MdID: msg.MdID}
	}
}
