package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// ADMIN ROUTES CONSTRUCTORS
// =============================================================================

func AdminRoutesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminRoutesFetchMsg{} }
}
func AdminRoutesSetCmd(routes []db.AdminRoutes) tea.Cmd {
	return func() tea.Msg { return AdminRoutesSet{AdminRoutes: routes} }
}
func CreateAdminRouteFromDialogCmd(title, slug string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminRouteFromDialogRequestMsg{Title: title, Slug: slug}
	}
}
func UpdateAdminRouteFromDialogCmd(routeID, title, slug, originalSlug string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminRouteFromDialogRequestMsg{
			RouteID:      routeID,
			Title:        title,
			Slug:         slug,
			OriginalSlug: originalSlug,
		}
	}
}
func DeleteAdminRouteCmd(adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminRouteRequestMsg{AdminRouteID: adminRouteID} }
}
func ShowDeleteAdminRouteDialogCmd(adminRouteID types.AdminRouteID, title string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminRouteDialogMsg{AdminRouteID: adminRouteID, Title: title}
	}
}

// =============================================================================
// ADMIN DATATYPES CONSTRUCTORS
// =============================================================================

func AdminAllDatatypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminAllDatatypesFetchMsg{} }
}
func AdminAllDatatypesSetCmd(datatypes []db.AdminDatatypes) tea.Cmd {
	return func() tea.Msg { return AdminAllDatatypesSet{AdminAllDatatypes: datatypes} }
}
func AdminDatatypeFieldsFetchCmd(adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg { return AdminDatatypeFieldsFetchMsg{AdminDatatypeID: adminDatatypeID} }
}
func AdminDatatypeFieldsSetCmd(fields []db.AdminFields) tea.Cmd {
	return func() tea.Msg { return AdminDatatypeFieldsSet{Fields: fields} }
}
func CreateAdminDatatypeFromDialogCmd(label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminDatatypeFromDialogRequestMsg{Label: label, Type: dtype, ParentID: parentID}
	}
}
func UpdateAdminDatatypeFromDialogCmd(adminDatatypeID, label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminDatatypeFromDialogRequestMsg{
			AdminDatatypeID: adminDatatypeID,
			Label:           label,
			Type:            dtype,
			ParentID:        parentID,
		}
	}
}
func DeleteAdminDatatypeCmd(adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminDatatypeRequestMsg{AdminDatatypeID: adminDatatypeID} }
}
func ShowDeleteAdminDatatypeDialogCmd(adminDatatypeID types.AdminDatatypeID, label string, hasChildren bool) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminDatatypeDialogMsg{
			AdminDatatypeID: adminDatatypeID,
			Label:           label,
			HasChildren:     hasChildren,
		}
	}
}

// =============================================================================
// ADMIN FIELDS CONSTRUCTORS
// =============================================================================

func CreateAdminFieldFromDialogCmd(label, fieldType string, adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminFieldFromDialogRequestMsg{
			Label:           label,
			Type:            fieldType,
			AdminDatatypeID: adminDatatypeID,
		}
	}
}
func UpdateAdminFieldFromDialogCmd(adminFieldID, label, fieldType string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminFieldFromDialogRequestMsg{
			AdminFieldID: adminFieldID,
			Label:        label,
			Type:         fieldType,
		}
	}
}
func DeleteAdminFieldCmd(adminFieldID types.AdminFieldID, adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return DeleteAdminFieldRequestMsg{AdminFieldID: adminFieldID}
	}
}
func ShowDeleteAdminFieldDialogCmd(adminFieldID types.AdminFieldID, adminDatatypeID types.AdminDatatypeID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminFieldDialogMsg{
			AdminFieldID:    adminFieldID,
			AdminDatatypeID: adminDatatypeID,
			Label:           label,
		}
	}
}

// =============================================================================
// ADMIN CONTENT CONSTRUCTORS
// =============================================================================

func AdminContentDataFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminContentDataFetchMsg{} }
}
func AdminContentDataSetCmd(data []db.AdminContentData) tea.Cmd {
	return func() tea.Msg { return AdminContentDataSet{AdminContentData: data} }
}
func DeleteAdminContentCmd(adminContentID types.AdminContentID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminContentRequestMsg{AdminContentID: adminContentID} }
}
