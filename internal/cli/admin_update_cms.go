package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// UpdateAdminCms routes admin CMS create/update/delete messages.
func (m Model) UpdateAdminCms(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	// =========================================================================
	// ADMIN ROUTE REQUEST MESSAGES → dispatch to handlers
	// =========================================================================
	case CreateAdminRouteFromDialogRequestMsg:
		return m, m.HandleCreateAdminRouteFromDialog(msg)
	case UpdateAdminRouteFromDialogRequestMsg:
		return m, m.HandleUpdateAdminRouteFromDialog(msg)
	case DeleteAdminRouteRequestMsg:
		return m, m.HandleDeleteAdminRoute(msg)

	// =========================================================================
	// ADMIN ROUTE RESULT MESSAGES → re-fetch data
	// =========================================================================
	case AdminRouteCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin route created: %s", msg.Title)),
			AdminRoutesFetchCmd(),
		)
	case AdminRouteUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin route updated: %s", msg.Title)),
			AdminRoutesFetchCmd(),
		)
	case AdminRouteDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin route deleted: %s", msg.AdminRouteID)),
			AdminRoutesFetchCmd(),
		)

	// =========================================================================
	// ADMIN DATATYPE REQUEST MESSAGES → dispatch to handlers
	// =========================================================================
	case CreateAdminDatatypeFromDialogRequestMsg:
		return m, m.HandleCreateAdminDatatypeFromDialog(msg)
	case UpdateAdminDatatypeFromDialogRequestMsg:
		return m, m.HandleUpdateAdminDatatypeFromDialog(msg)
	case DeleteAdminDatatypeRequestMsg:
		return m, m.HandleDeleteAdminDatatype(msg)

	// =========================================================================
	// ADMIN DATATYPE RESULT MESSAGES → re-fetch data
	// =========================================================================
	case AdminDatatypeCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin datatype created: %s", msg.Label)),
			AdminAllDatatypesFetchCmd(),
		)
	case AdminDatatypeUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin datatype updated: %s", msg.Label)),
			AdminAllDatatypesFetchCmd(),
		)
	case AdminDatatypeDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		newModel.AdminFieldCursor = 0
		newModel.AdminSelectedDatatypeFields = nil
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin datatype deleted: %s", msg.AdminDatatypeID)),
			AdminAllDatatypesFetchCmd(),
		)

	// =========================================================================
	// ADMIN FIELD REQUEST MESSAGES → dispatch to handlers
	// =========================================================================
	case CreateAdminFieldFromDialogRequestMsg:
		return m, m.HandleCreateAdminFieldFromDialog(msg)
	case UpdateAdminFieldFromDialogRequestMsg:
		return m, m.HandleUpdateAdminFieldFromDialog(msg)
	case DeleteAdminFieldRequestMsg:
		return m, m.HandleDeleteAdminField(msg)

	// =========================================================================
	// ADMIN FIELD RESULT MESSAGES → re-fetch data
	// =========================================================================
	case AdminFieldCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin field created: %s", msg.Label)),
			AdminDatatypeFieldsFetchCmd(msg.AdminDatatypeID),
		)
	case AdminFieldUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin field updated: %s", msg.Label)),
			AdminDatatypeFieldsFetchCmd(msg.AdminDatatypeID),
		)
	case AdminFieldDeletedMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin field deleted: %s", msg.AdminFieldID)),
			AdminDatatypeFieldsFetchCmd(msg.AdminDatatypeID),
		)

	// =========================================================================
	// ADMIN CONTENT REQUEST MESSAGES → dispatch to handlers
	// =========================================================================
	case DeleteAdminContentRequestMsg:
		return m, m.HandleDeleteAdminContent(msg)

	// =========================================================================
	// ADMIN CONTENT RESULT MESSAGES → re-fetch data
	// =========================================================================
	case AdminContentDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin content deleted: %s", msg.AdminContentID)),
			AdminContentDataFetchCmd(),
		)

	default:
		return m, nil
	}
}
