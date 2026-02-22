package cli

import (
	"fmt"
	"strings"

	"github.com/hegner123/modulacms/internal/tui"
)

// renderAdminRoutesList renders the admin routes list for the left panel.
func renderAdminRoutesList(m Model) string {
	if len(m.AdminRoutes) == 0 {
		return "(no admin routes)"
	}

	lines := make([]string, 0, len(m.AdminRoutes))
	for i, route := range m.AdminRoutes {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s %s", cursor, route.Title, route.Slug))
	}
	return strings.Join(lines, "\n")
}

// renderAdminRouteDetail renders the selected admin route details for the center panel.
func renderAdminRouteDetail(m Model) string {
	if len(m.AdminRoutes) == 0 || m.Cursor >= len(m.AdminRoutes) {
		return "No admin route selected"
	}

	route := m.AdminRoutes[m.Cursor]
	lines := []string{
		fmt.Sprintf("Title:    %s", route.Title),
		fmt.Sprintf("Slug:     %s", route.Slug),
		fmt.Sprintf("Status:   %d", route.Status),
		fmt.Sprintf("Author:   %s", route.AuthorID.String()),
		fmt.Sprintf("Created:  %s", route.DateCreated.String()),
		fmt.Sprintf("Modified: %s", route.DateModified.String()),
	}

	return strings.Join(lines, "\n")
}

// renderAdminRouteActions renders available actions for the right panel on admin routes.
func renderAdminRouteActions(m Model) string {
	lines := []string{
		"Actions",
		"",
		"  n: New",
		"  e: Edit",
		"  d: Delete",
		"",
		fmt.Sprintf("Routes: %d", len(m.AdminRoutes)),
	}

	return strings.Join(lines, "\n")
}

// renderAdminDatatypesList renders all admin datatypes for the left panel.
func renderAdminDatatypesList(m Model) string {
	if len(m.AdminAllDatatypes) == 0 {
		return "(no admin datatypes)"
	}

	lines := make([]string, 0, len(m.AdminAllDatatypes))
	for i, dt := range m.AdminAllDatatypes {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		parent := ""
		if dt.ParentID.Valid {
			parent = fmt.Sprintf(" (child of %s)", dt.ParentID.ID)
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]%s", cursor, dt.Label, dt.Type, parent))
	}
	return strings.Join(lines, "\n")
}

// renderAdminDatatypeFields renders the fields for the selected admin datatype in the center panel.
// Shows cursor when ContentPanel has focus.
func renderAdminDatatypeFields(m Model) string {
	if len(m.AdminAllDatatypes) == 0 || m.Cursor >= len(m.AdminAllDatatypes) {
		return "No admin datatype selected"
	}

	dt := m.AdminAllDatatypes[m.Cursor]
	lines := []string{
		fmt.Sprintf("Fields for: %s", dt.Label),
		"",
	}

	if len(m.AdminSelectedDatatypeFields) == 0 {
		if m.PanelFocus == tui.ContentPanel {
			lines = append(lines, " -> (empty)")
		} else {
			lines = append(lines, "    (empty)")
		}
		lines = append(lines, "")
		lines = append(lines, "Press 'n' to add a field")
	} else {
		for i, field := range m.AdminSelectedDatatypeFields {
			cursor := "   "
			if m.PanelFocus == tui.ContentPanel && m.AdminFieldCursor == i {
				cursor = " ->"
			}
			lines = append(lines, fmt.Sprintf("%s %d. %s [%s]", cursor, i+1, field.Label, field.Type))
		}
	}

	return strings.Join(lines, "\n")
}

// renderAdminDatatypeActions renders available actions for the right panel on admin datatypes.
// Shows context-sensitive hints based on which panel is focused.
func renderAdminDatatypeActions(m Model) string {
	lines := []string{
		"Actions",
		"",
	}

	switch m.PanelFocus {
	case tui.TreePanel:
		lines = append(lines,
			"Datatypes Panel",
			"",
			"  n: New datatype",
			"  e: Edit datatype",
			"  d: Delete datatype",
			"",
			"  enter: Select",
			"  tab: Switch panel",
		)
	case tui.ContentPanel:
		lines = append(lines,
			"Fields Panel",
			"",
			"  n: New field",
			"  e: Edit field",
			"  d: Delete field",
			"",
			"  esc/h: Back to datatypes",
			"  tab: Switch panel",
		)
	default:
		lines = append(lines,
			"  n: New",
			"  e: Edit",
			"  d: Delete",
		)
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Datatypes: %d", len(m.AdminAllDatatypes)))
	if len(m.AdminAllDatatypes) > 0 && m.Cursor < len(m.AdminAllDatatypes) {
		lines = append(lines, fmt.Sprintf("Fields: %d", len(m.AdminSelectedDatatypeFields)))
	}

	return strings.Join(lines, "\n")
}

// =============================================================================
// FIELD TYPES RENDER FUNCTIONS
// =============================================================================

// renderFieldTypesList renders the field types list for the left panel.
func renderFieldTypesList(m Model) string {
	if len(m.FieldTypesList) == 0 {
		return "(no field types)"
	}

	lines := make([]string, 0, len(m.FieldTypesList))
	for i, ft := range m.FieldTypesList {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]", cursor, ft.Label, ft.Type))
	}
	return strings.Join(lines, "\n")
}

// renderFieldTypeDetail renders the selected field type details for the center panel.
func renderFieldTypeDetail(m Model) string {
	if len(m.FieldTypesList) == 0 || m.Cursor >= len(m.FieldTypesList) {
		return "No field type selected"
	}

	ft := m.FieldTypesList[m.Cursor]
	lines := []string{
		fmt.Sprintf("Type:  %s", ft.Type),
		fmt.Sprintf("Label: %s", ft.Label),
		"",
		fmt.Sprintf("ID:    %s", ft.FieldTypeID),
	}

	return strings.Join(lines, "\n")
}

// renderFieldTypeActions renders available actions for the right panel on field types.
func renderFieldTypeActions(m Model) string {
	lines := []string{
		"Actions",
		"",
		"  n: New",
		"  e: Edit",
		"  d: Delete",
		"",
		fmt.Sprintf("Field Types: %d", len(m.FieldTypesList)),
	}

	return strings.Join(lines, "\n")
}

// =============================================================================
// ADMIN FIELD TYPES RENDER FUNCTIONS
// =============================================================================

// renderAdminFieldTypesList renders the admin field types list for the left panel.
func renderAdminFieldTypesList(m Model) string {
	if len(m.AdminFieldTypesList) == 0 {
		return "(no admin field types)"
	}

	lines := make([]string, 0, len(m.AdminFieldTypesList))
	for i, ft := range m.AdminFieldTypesList {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]", cursor, ft.Label, ft.Type))
	}
	return strings.Join(lines, "\n")
}

// renderAdminFieldTypeDetail renders the selected admin field type details for the center panel.
func renderAdminFieldTypeDetail(m Model) string {
	if len(m.AdminFieldTypesList) == 0 || m.Cursor >= len(m.AdminFieldTypesList) {
		return "No admin field type selected"
	}

	ft := m.AdminFieldTypesList[m.Cursor]
	lines := []string{
		fmt.Sprintf("Type:  %s", ft.Type),
		fmt.Sprintf("Label: %s", ft.Label),
		"",
		fmt.Sprintf("ID:    %s", ft.AdminFieldTypeID),
	}

	return strings.Join(lines, "\n")
}

// renderAdminFieldTypeActions renders available actions for the right panel on admin field types.
func renderAdminFieldTypeActions(m Model) string {
	lines := []string{
		"Actions",
		"",
		"  n: New",
		"  e: Edit",
		"  d: Delete",
		"",
		fmt.Sprintf("Admin Field Types: %d", len(m.AdminFieldTypesList)),
	}

	return strings.Join(lines, "\n")
}

// renderAdminContentList renders a flat list of admin content data for the left panel.
func renderAdminContentList(m Model) string {
	if len(m.AdminRootContentSummary) == 0 {
		return "(no admin content)"
	}

	lines := make([]string, 0, len(m.AdminRootContentSummary))
	for i, content := range m.AdminRootContentSummary {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		label := string(content.AdminContentDataID)
		if content.AdminDatatypeID.Valid {
			label = fmt.Sprintf("[%s] %s", content.AdminDatatypeID.ID, label)
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, label))
	}
	return strings.Join(lines, "\n")
}

// renderAdminContentDetail renders details for the selected admin content in the center panel.
func renderAdminContentDetail(m Model) string {
	if len(m.AdminRootContentSummary) == 0 || m.Cursor >= len(m.AdminRootContentSummary) {
		return "No admin content selected"
	}

	content := m.AdminRootContentSummary[m.Cursor]
	lines := []string{
		fmt.Sprintf("ID:        %s", content.AdminContentDataID),
		fmt.Sprintf("Route:     %s", content.AdminRouteID),
		fmt.Sprintf("Status:    %s", content.Status),
		fmt.Sprintf("Author:    %s", content.AuthorID.String()),
		"",
		fmt.Sprintf("Created:   %s", content.DateCreated.String()),
		fmt.Sprintf("Modified:  %s", content.DateModified.String()),
	}

	if content.AdminDatatypeID.Valid {
		lines = append([]string{fmt.Sprintf("Datatype:  %s", content.AdminDatatypeID.ID)}, lines...)
	}

	return strings.Join(lines, "\n")
}
