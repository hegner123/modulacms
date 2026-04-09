package tui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
)

// DatatypeFieldReorderedMsg signals that field sort_order was swapped on the datatypes screen.
type DatatypeFieldReorderedMsg struct {
	Direction string
}

// AdminDatatypeFieldReorderedMsg signals that admin field sort_order was swapped.
type AdminDatatypeFieldReorderedMsg struct {
	Direction string
}

// ReorderFieldCmd swaps sort_order between two adjacent fields.
// Uses the cfg+userID pattern to build audit context (matches existing datatype reorder).
func ReorderFieldCmd(
	fieldA types.FieldID, fieldB types.FieldID,
	orderA int64, orderB int64,
	direction string,
) tea.Cmd {
	return func() tea.Msg {
		return ReorderFieldRequestMsg{
			AID: fieldA, BID: fieldB,
			AOrder: orderA, BOrder: orderB,
			Direction: direction,
		}
	}
}

// ReorderFieldRequestMsg requests a field sort_order swap.
type ReorderFieldRequestMsg struct {
	AID       types.FieldID
	BID       types.FieldID
	AOrder    int64
	BOrder    int64
	Direction string
}

// HandleReorderFieldStandalone swaps sort_order between two fields using provided driver.
func HandleReorderFieldStandalone(cfg config.Config, userID types.UserID, driver db.DbDriver, msg ReorderFieldRequestMsg) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(cfg, userID)

		if err := driver.UpdateFieldSortOrder(ctx, ac, db.UpdateFieldSortOrderParams{
			FieldID:   msg.AID,
			SortOrder: msg.BOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to reorder field: %v", err)}
		}
		if err := driver.UpdateFieldSortOrder(ctx, ac, db.UpdateFieldSortOrderParams{
			FieldID:   msg.BID,
			SortOrder: msg.AOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to reorder field: %v", err)}
		}

		return DatatypeFieldReorderedMsg{Direction: msg.Direction}
	}
}

// ReorderAdminFieldCmd swaps sort_order between two adjacent admin fields.
func ReorderAdminFieldCmd(
	fieldA types.AdminFieldID, fieldB types.AdminFieldID,
	orderA int64, orderB int64,
	direction string,
) tea.Cmd {
	return func() tea.Msg {
		return ReorderAdminFieldRequestMsg{
			AID: fieldA, BID: fieldB,
			AOrder: orderA, BOrder: orderB,
			Direction: direction,
		}
	}
}

// ReorderAdminFieldRequestMsg requests an admin field sort_order swap.
type ReorderAdminFieldRequestMsg struct {
	AID       types.AdminFieldID
	BID       types.AdminFieldID
	AOrder    int64
	BOrder    int64
	Direction string
}

// HandleReorderAdminFieldStandalone swaps sort_order between two admin fields using provided driver.
func HandleReorderAdminFieldStandalone(cfg config.Config, userID types.UserID, driver db.DbDriver, msg ReorderAdminFieldRequestMsg) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(cfg, userID)

		if err := driver.UpdateAdminFieldSortOrder(ctx, ac, db.UpdateAdminFieldSortOrderParams{
			AdminFieldID: msg.AID,
			SortOrder:    msg.BOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to reorder admin field: %v", err)}
		}
		if err := driver.UpdateAdminFieldSortOrder(ctx, ac, db.UpdateAdminFieldSortOrderParams{
			AdminFieldID: msg.BID,
			SortOrder:    msg.AOrder,
		}); err != nil {
			return ActionResultMsg{Title: "Error", Message: fmt.Sprintf("failed to reorder admin field: %v", err)}
		}

		return AdminDatatypeFieldReorderedMsg{Direction: msg.Direction}
	}
}
