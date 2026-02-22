package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// ADMIN ROUTES CONSTRUCTORS
// =============================================================================

// AdminRoutesFetchCmd creates a command to fetch all admin routes.
func AdminRoutesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminRoutesFetchMsg{} }
}

// AdminRoutesSetCmd creates a command to set the admin routes list.
func AdminRoutesSetCmd(routes []db.AdminRoutes) tea.Cmd {
	return func() tea.Msg { return AdminRoutesSet{AdminRoutes: routes} }
}

// CreateAdminRouteFromDialogCmd creates a command to create an admin route from dialog input.
func CreateAdminRouteFromDialogCmd(title, slug string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminRouteFromDialogRequestMsg{Title: title, Slug: slug}
	}
}
// UpdateAdminRouteFromDialogCmd creates a command to update an admin route from dialog input.
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
// DeleteAdminRouteCmd creates a command to delete an admin route.
func DeleteAdminRouteCmd(adminRouteID types.AdminRouteID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminRouteRequestMsg{AdminRouteID: adminRouteID} }
}

// ShowDeleteAdminRouteDialogCmd creates a command to show the delete admin route dialog.
func ShowDeleteAdminRouteDialogCmd(adminRouteID types.AdminRouteID, title string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminRouteDialogMsg{AdminRouteID: adminRouteID, Title: title}
	}
}

// =============================================================================
// ADMIN DATATYPES CONSTRUCTORS
// =============================================================================

// AdminAllDatatypesFetchCmd creates a command to fetch all admin datatypes.
func AdminAllDatatypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminAllDatatypesFetchMsg{} }
}

// AdminAllDatatypesSetCmd creates a command to set the admin datatypes list.
func AdminAllDatatypesSetCmd(datatypes []db.AdminDatatypes) tea.Cmd {
	return func() tea.Msg { return AdminAllDatatypesSet{AdminAllDatatypes: datatypes} }
}

// AdminDatatypeFieldsFetchCmd creates a command to fetch fields for an admin datatype.
func AdminDatatypeFieldsFetchCmd(adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg { return AdminDatatypeFieldsFetchMsg{AdminDatatypeID: adminDatatypeID} }
}

// AdminDatatypeFieldsSetCmd creates a command to set the admin datatype fields list.
func AdminDatatypeFieldsSetCmd(fields []db.AdminFields) tea.Cmd {
	return func() tea.Msg { return AdminDatatypeFieldsSet{Fields: fields} }
}

// CreateAdminDatatypeFromDialogCmd creates a command to create an admin datatype from dialog input.
func CreateAdminDatatypeFromDialogCmd(label, dtype, parentID string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminDatatypeFromDialogRequestMsg{Label: label, Type: dtype, ParentID: parentID}
	}
}
// UpdateAdminDatatypeFromDialogCmd creates a command to update an admin datatype from dialog input.
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
// DeleteAdminDatatypeCmd creates a command to delete an admin datatype.
func DeleteAdminDatatypeCmd(adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminDatatypeRequestMsg{AdminDatatypeID: adminDatatypeID} }
}

// ShowDeleteAdminDatatypeDialogCmd creates a command to show the delete admin datatype dialog.
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

// CreateAdminFieldFromDialogCmd creates a command to create an admin field from dialog input.
func CreateAdminFieldFromDialogCmd(label, fieldType string, adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminFieldFromDialogRequestMsg{
			Label:           label,
			Type:            fieldType,
			AdminDatatypeID: adminDatatypeID,
		}
	}
}
// UpdateAdminFieldFromDialogCmd creates a command to update an admin field from dialog input.
func UpdateAdminFieldFromDialogCmd(adminFieldID, label, fieldType string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminFieldFromDialogRequestMsg{
			AdminFieldID: adminFieldID,
			Label:        label,
			Type:         fieldType,
		}
	}
}
// DeleteAdminFieldCmd creates a command to delete an admin field.
func DeleteAdminFieldCmd(adminFieldID types.AdminFieldID, adminDatatypeID types.AdminDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return DeleteAdminFieldRequestMsg{AdminFieldID: adminFieldID}
	}
}

// ShowDeleteAdminFieldDialogCmd creates a command to show the delete admin field dialog.
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

// AdminContentDataFetchCmd creates a command to fetch all admin content data.
func AdminContentDataFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminContentDataFetchMsg{} }
}

// AdminContentDataSetCmd creates a command to set the admin content data list.
func AdminContentDataSetCmd(data []db.AdminContentData) tea.Cmd {
	return func() tea.Msg { return AdminContentDataSet{AdminContentData: data} }
}

// DeleteAdminContentCmd creates a command to delete admin content.
func DeleteAdminContentCmd(adminContentID types.AdminContentID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminContentRequestMsg{AdminContentID: adminContentID} }
}

// =============================================================================
// FIELD TYPES CONSTRUCTORS
// =============================================================================

// CreateFieldTypeFromDialogCmd creates a command to create a field type from dialog input.
func CreateFieldTypeFromDialogCmd(fieldType, label string) tea.Cmd {
	return func() tea.Msg {
		return CreateFieldTypeFromDialogRequestMsg{Type: fieldType, Label: label}
	}
}

// UpdateFieldTypeFromDialogCmd creates a command to update a field type from dialog input.
func UpdateFieldTypeFromDialogCmd(fieldTypeID, fieldType, label string) tea.Cmd {
	return func() tea.Msg {
		return UpdateFieldTypeFromDialogRequestMsg{
			FieldTypeID: fieldTypeID,
			Type:        fieldType,
			Label:       label,
		}
	}
}

// CreateAdminFieldTypeFromDialogCmd creates a command to create an admin field type from dialog input.
func CreateAdminFieldTypeFromDialogCmd(fieldType, label string) tea.Cmd {
	return func() tea.Msg {
		return CreateAdminFieldTypeFromDialogRequestMsg{Type: fieldType, Label: label}
	}
}

// UpdateAdminFieldTypeFromDialogCmd creates a command to update an admin field type from dialog input.
func UpdateAdminFieldTypeFromDialogCmd(adminFieldTypeID, fieldType, label string) tea.Cmd {
	return func() tea.Msg {
		return UpdateAdminFieldTypeFromDialogRequestMsg{
			AdminFieldTypeID: adminFieldTypeID,
			Type:             fieldType,
			Label:            label,
		}
	}
}

// FieldTypesFetchCmd creates a command to fetch all field types.
func FieldTypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return FieldTypesFetchMsg{} }
}

// FieldTypesSetCmd creates a command to set the field types list.
func FieldTypesSetCmd(fieldTypes []db.FieldTypes) tea.Cmd {
	return func() tea.Msg { return FieldTypesSet{FieldTypes: fieldTypes} }
}

// DeleteFieldTypeCmd creates a command to delete a field type.
func DeleteFieldTypeCmd(fieldTypeID types.FieldTypeID) tea.Cmd {
	return func() tea.Msg { return DeleteFieldTypeRequestMsg{FieldTypeID: fieldTypeID} }
}

// ShowDeleteFieldTypeDialogCmd creates a command to show the delete field type dialog.
func ShowDeleteFieldTypeDialogCmd(fieldTypeID types.FieldTypeID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteFieldTypeDialogMsg{FieldTypeID: fieldTypeID, Label: label}
	}
}

// =============================================================================
// ADMIN FIELD TYPES CONSTRUCTORS
// =============================================================================

// AdminFieldTypesFetchCmd creates a command to fetch all admin field types.
func AdminFieldTypesFetchCmd() tea.Cmd {
	return func() tea.Msg { return AdminFieldTypesFetchMsg{} }
}

// AdminFieldTypesSetCmd creates a command to set the admin field types list.
func AdminFieldTypesSetCmd(adminFieldTypes []db.AdminFieldTypes) tea.Cmd {
	return func() tea.Msg { return AdminFieldTypesSet{AdminFieldTypes: adminFieldTypes} }
}

// DeleteAdminFieldTypeCmd creates a command to delete an admin field type.
func DeleteAdminFieldTypeCmd(adminFieldTypeID types.AdminFieldTypeID) tea.Cmd {
	return func() tea.Msg { return DeleteAdminFieldTypeRequestMsg{AdminFieldTypeID: adminFieldTypeID} }
}

// ShowDeleteAdminFieldTypeDialogCmd creates a command to show the delete admin field type dialog.
func ShowDeleteAdminFieldTypeDialogCmd(adminFieldTypeID types.AdminFieldTypeID, label string) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteAdminFieldTypeDialogMsg{AdminFieldTypeID: adminFieldTypeID, Label: label}
	}
}
