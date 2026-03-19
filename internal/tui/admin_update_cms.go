package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
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
	case CreateAdminRouteWithContentRequestMsg:
		return m, m.HandleCreateAdminRouteWithContent(msg)

	// =========================================================================
	// ADMIN ROUTE RESULT MESSAGES → re-fetch data
	// =========================================================================
	case AdminRouteWithContentCreatedMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin route created with content: %s (ContentID: %s)", msg.Title, msg.AdminContentDataID)),
			AdminRoutesFetchCmd(),
		)
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
	case ReorderAdminDatatypeRequestMsg:
		return m, m.HandleReorderAdminDatatype(msg)
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
	case ConfirmedDeleteAdminContentMsg:
		return m, m.HandleDeleteAdminContent(msg)
	case AdminReorderSiblingRequestMsg:
		return m, m.HandleAdminReorderSibling(msg)
	case AdminCopyContentRequestMsg:
		return m, m.HandleCopyAdminContent(msg)
	case AdminMoveContentRequestMsg:
		return m, m.HandleMoveAdminContent(msg)
	case ConfirmedPublishAdminContentMsg:
		return m, m.HandleAdminConfirmedPublish(msg)
	case ConfirmedUnpublishAdminContentMsg:
		return m, m.HandleAdminConfirmedUnpublish(msg)
	case AdminListVersionsRequestMsg:
		return m, m.HandleAdminListVersions(msg)
	case ConfirmedRestoreAdminVersionMsg:
		return m, m.HandleAdminConfirmedRestoreVersion(msg)
	case ConfirmedDeleteAdminContentFieldMsg:
		return m, m.HandleDeleteAdminContentField(msg)

	case AdminFetchContentForEditMsg:
		locale := m.ActiveLocale
		return m, FetchAdminContentForEditCmd(m.Config, msg.AdminContentID, msg.AdminDatatypeID, msg.AdminRouteID, msg.Title, locale)

	// =========================================================================
	// ADMIN CONTENT RESULT MESSAGES
	// =========================================================================
	// Most admin result messages are now unified types (ContentCreatedMsg,
	// ContentDeletedMsg, etc. with AdminMode=true) and handled in UpdateCms
	// or fall through to the screen. Only admin-specific messages remain here.

	case AdminTreeLoadedMsg:
		// Screen handles tree state; Model-level just stops loading.
		return m, LoadingStopCmd()

	case AdminContentUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin content updated: %s", msg.AdminContentID)),
			ReloadAdminContentTreeCmd(m.Config, msg.AdminRouteID),
		)
	case AdminVersionsListedMsg:
		// Screen handles version list state; Model-level just stops loading.
		return m, LoadingStopCmd()
	case AdminLoadContentFieldsMsg:
		// Screen handles field display state directly.
		return m, nil
	case AdminRootDatatypesFetchResultsMsg:
		// Screen handles root datatypes state directly.
		return m, nil

	// =========================================================================
	// FIELD TYPE REQUEST MESSAGES -> dispatch to handlers
	// =========================================================================
	case CreateFieldTypeFromDialogRequestMsg:
		return m, m.HandleCreateFieldTypeFromDialog(msg)
	case UpdateFieldTypeFromDialogRequestMsg:
		return m, m.HandleUpdateFieldTypeFromDialog(msg)
	case DeleteFieldTypeRequestMsg:
		return m, m.HandleDeleteFieldType(msg)

	// =========================================================================
	// FIELD TYPE RESULT MESSAGES -> re-fetch data
	// =========================================================================
	case FieldTypeCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Field type created: %s", msg.Label)),
			FieldTypesFetchCmd(),
		)
	case FieldTypeUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Field type updated: %s", msg.Label)),
			FieldTypesFetchCmd(),
		)
	case FieldTypeDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Field type deleted: %s", msg.FieldTypeID)),
			FieldTypesFetchCmd(),
		)

	// =========================================================================
	// ADMIN FIELD TYPE REQUEST MESSAGES -> dispatch to handlers
	// =========================================================================
	case CreateAdminFieldTypeFromDialogRequestMsg:
		return m, m.HandleCreateAdminFieldTypeFromDialog(msg)
	case UpdateAdminFieldTypeFromDialogRequestMsg:
		return m, m.HandleUpdateAdminFieldTypeFromDialog(msg)
	case DeleteAdminFieldTypeRequestMsg:
		return m, m.HandleDeleteAdminFieldType(msg)

	// =========================================================================
	// ADMIN FIELD TYPE RESULT MESSAGES -> re-fetch data
	// =========================================================================
	case AdminFieldTypeCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin field type created: %s", msg.Label)),
			AdminFieldTypesFetchCmd(),
		)
	case AdminFieldTypeUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin field type updated: %s", msg.Label)),
			AdminFieldTypesFetchCmd(),
		)
	case AdminFieldTypeDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin field type deleted: %s", msg.AdminFieldTypeID)),
			AdminFieldTypesFetchCmd(),
		)

	// =========================================================================
	// VALIDATION REQUEST MESSAGES -> dispatch to handlers
	// =========================================================================
	case CreateValidationFromDialogRequestMsg:
		return m, m.HandleCreateValidationFromDialog(msg)
	case UpdateValidationFromDialogRequestMsg:
		return m, m.HandleUpdateValidationFromDialog(msg)
	case DeleteValidationRequestMsg:
		return m, m.HandleDeleteValidation(msg)

	// =========================================================================
	// VALIDATION RESULT MESSAGES -> re-fetch data
	// =========================================================================
	case ValidationCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Validation created: %s", msg.Name)),
			ValidationsFetchCmd(),
		)
	case ValidationUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Validation updated: %s", msg.Name)),
			ValidationsFetchCmd(),
		)
	case ValidationDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Validation deleted: %s", msg.ValidationID)),
			ValidationsFetchCmd(),
		)

	// =========================================================================
	// ADMIN VALIDATION REQUEST MESSAGES -> dispatch to handlers
	// =========================================================================
	case CreateAdminValidationFromDialogRequestMsg:
		return m, m.HandleCreateAdminValidationFromDialog(msg)
	case UpdateAdminValidationFromDialogRequestMsg:
		return m, m.HandleUpdateAdminValidationFromDialog(msg)
	case DeleteAdminValidationRequestMsg:
		return m, m.HandleDeleteAdminValidation(msg)

	// =========================================================================
	// ADMIN VALIDATION RESULT MESSAGES -> re-fetch data
	// =========================================================================
	case AdminValidationCreatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin validation created: %s", msg.Name)),
			AdminValidationsFetchCmd(),
		)
	case AdminValidationUpdatedFromDialogMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin validation updated: %s", msg.Name)),
			AdminValidationsFetchCmd(),
		)
	case AdminValidationDeletedMsg:
		newModel := m
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Admin validation deleted: %s", msg.AdminValidationID)),
			AdminValidationsFetchCmd(),
		)

	default:
		return m, nil
	}
}
