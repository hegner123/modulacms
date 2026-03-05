package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// HandleAdminContentFormDialogAccept dispatches admin content form dialog
// submissions based on the form action.
func (m Model) HandleAdminContentFormDialogAccept(msg AdminContentFormDialogAcceptMsg) (Model, tea.Cmd) {
	switch msg.Action {
	case FORMDIALOGCREATEADMINCONTENT:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			m.HandleCreateAdminContentFromDialog(msg.DatatypeID, msg.RouteID, msg.ParentID, msg.FieldValues),
		)
	case FORMDIALOGEDITADMINCONTENT:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
			LoadingStartCmd(),
			m.HandleUpdateAdminContentFromDialog(msg.ContentID, msg.RouteID, msg.FieldValues),
		)
	case FORMDIALOGEDITADMINSINGLEFIELD:
		// Single-field edit: use stored context for AdminContentFieldID
		if m.DCtx.EditAdminSingleField != nil {
			ctx := m.DCtx.EditAdminSingleField
			m.DCtx.EditAdminSingleField = nil
			var newValue string
			for _, val := range msg.FieldValues {
				newValue = val
				break
			}
			return m, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
				m.HandleEditAdminSingleField(
					ctx.AdminContentFieldID,
					ctx.AdminContentID,
					ctx.AdminFieldID,
					newValue,
					ctx.AdminRouteID,
					ctx.AdminDatatypeID,
				),
			)
		}
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	default:
		return m, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	}
}
