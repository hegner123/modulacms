package cli

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/tui"
)

// isCMSPanelPage returns true for pages that use the 3-panel CMS layout.
func isCMSPanelPage(idx PageIndex) bool {
	switch idx {
	case CMSPAGE, ADMINCMSPAGE, CONTENT, MEDIA, USERSADMIN, ROUTES, DATATYPES:
		return true
	default:
		return false
	}
}

// renderCMSPanelLayout renders the full panel layout: header, 3 bordered panels, status bar.
func renderCMSPanelLayout(m Model) string {
	header := renderCMSHeader(m)
	statusBar := renderCMSPanelStatusBar(m)

	headerH := lipgloss.Height(header)
	statusH := lipgloss.Height(statusBar)

	bodyH := m.Height - headerH - statusH
	if bodyH < 3 {
		bodyH = 3
	}

	leftW := m.Width / 4
	centerW := m.Width / 2
	rightW := m.Width - leftW - centerW

	left, center, right := cmsPanelContent(m)

	treePanel := tui.Panel{
		Title:   "Tree",
		Width:   leftW,
		Height:  bodyH,
		Content: left,
		Focused: m.PanelFocus == tui.TreePanel,
	}

	contentPanel := tui.Panel{
		Title:   "Content",
		Width:   centerW,
		Height:  bodyH,
		Content: center,
		Focused: m.PanelFocus == tui.ContentPanel,
	}

	routePanel := tui.Panel{
		Title:   "Route",
		Width:   rightW,
		Height:  bodyH,
		Content: right,
		Focused: m.PanelFocus == tui.RoutePanel,
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		treePanel.Render(),
		contentPanel.Render(),
		routePanel.Render(),
	)

	ui := lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar)

	if m.DialogActive && m.Dialog != nil {
		return DialogOverlay(ui, *m.Dialog, m.Width, m.Height)
	}

	if m.FormDialogActive && m.FormDialog != nil {
		return FormDialogOverlay(ui, *m.FormDialog, m.Width, m.Height)
	}

	return ui
}

// cmsPanelContent returns the left, center, and right panel content strings
// based on the current page.
func cmsPanelContent(m Model) (left, center, right string) {
	switch m.Page.Index {
	case CMSPAGE, ADMINCMSPAGE:
		left = renderCMSMenuContent(m)
		center = "Select an item"
		right = "Route\n\n  (none)"

	case CONTENT:
		if m.PageRouteId.IsZero() {
			// Content selection flow: ROOT types -> routes
			left = renderRootDatatypesList(m)
			if m.SelectedDatatype.IsZero() {
				center = "Select a root type"
				right = fmt.Sprintf("Root Types: %d", len(m.RootDatatypes))
			} else {
				center = renderRoutesList(m)
				right = renderRouteActions(m)
			}
		} else {
			// Content browsing: tree view
			cms := CMSPage{}
			left = cms.ProcessTreeDatatypes(m)
			center = cms.ProcessContentPreview(m)
			right = cms.ProcessFields(m)
		}

	case ROUTES:
		left = renderRoutesList(m)
		center = renderRouteDetail(m)
		right = renderRouteActions(m)

	case MEDIA:
		left = renderMediaList(m)
		center = renderMediaDetail(m)
		right = renderMediaInfo(m)

	case USERSADMIN:
		left = "Users"
		center = "Select a user"
		right = "Permissions"

	case DATATYPES:
		left = renderDatatypesList(m)
		center = renderDatatypeDetail(m)
		right = renderDatatypeActions(m)

	default:
		left = ""
		center = ""
		right = ""
	}
	return left, center, right
}

// renderCMSMenuContent renders the menu list for the left panel on CMSPAGE/ADMINCMSPAGE.
func renderCMSMenuContent(m Model) string {
	if len(m.PageMenu) == 0 {
		return "(no items)"
	}

	lines := make([]string, 0, len(m.PageMenu))
	for i, item := range m.PageMenu {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, item.Label))
	}
	return strings.Join(lines, "\n")
}

// renderCMSHeader renders the top bar with the app title and action buttons.
func renderCMSHeader(m Model) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Accent).
		PaddingRight(2)

	actions := []string{"New", "Save", "Copy", "Duplicate", "Export"}
	buttons := make([]string, len(actions))
	for i, a := range actions {
		buttons[i] = RenderButton(a)
	}

	title := titleStyle.Render(RenderTitle(m.Titles[m.TitleFont]))
	buttonBar := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)

	row := lipgloss.JoinHorizontal(lipgloss.Center, title, buttonBar)

	container := lipgloss.NewStyle().
		Width(m.Width).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(config.DefaultStyle.Tertiary).
		Padding(0, 1).
		Render(row)

	return container
}

// renderCMSPanelStatusBar renders the bottom status bar with status badge,
// panel focus indicator, and key hints.
func renderCMSPanelStatusBar(m Model) string {
	barStyle := lipgloss.NewStyle().
		Background(config.DefaultStyle.Status2BG).
		Foreground(config.DefaultStyle.Status1)

	// Left: status badge
	statusBadge := m.GetStatus()

	// Center: panel focus indicator
	panels := []tui.FocusPanel{tui.TreePanel, tui.ContentPanel, tui.RoutePanel}
	focusParts := make([]string, len(panels))
	for i, p := range panels {
		label := p.String()
		if m.PanelFocus == p {
			focusParts[i] = barStyle.Bold(true).Padding(0, 1).Render("[" + label + "]")
		} else {
			focusParts[i] = barStyle.Padding(0, 1).Render(" " + label + " ")
		}
	}
	focusIndicator := lipgloss.JoinHorizontal(lipgloss.Center, focusParts...)

	// Right: key hints
	hints := barStyle.Padding(0, 1).Render("tab:panel  h:back  q:quit")

	// Calculate spacing
	statusW := lipgloss.Width(statusBadge)
	focusW := lipgloss.Width(focusIndicator)
	hintsW := lipgloss.Width(hints)

	leftGap := (m.Width - statusW - focusW - hintsW) / 2
	if leftGap < 1 {
		leftGap = 1
	}
	rightGap := m.Width - statusW - leftGap - focusW - hintsW
	if rightGap < 0 {
		rightGap = 0
	}

	leftSpacer := barStyle.Render(strings.Repeat(" ", leftGap))
	rightSpacer := barStyle.Render(strings.Repeat(" ", rightGap))

	return statusBadge + leftSpacer + focusIndicator + rightSpacer + hints
}

// renderRootDatatypesList renders ROOT datatypes for the left panel on the CONTENT page.
func renderRootDatatypesList(m Model) string {
	if len(m.RootDatatypes) == 0 {
		return "(no root types)"
	}

	lines := make([]string, 0, len(m.RootDatatypes))
	for i, dt := range m.RootDatatypes {
		cursor := "   "
		if m.SelectedDatatype.IsZero() && m.Cursor == i {
			cursor = " ->"
		}
		active := ""
		if dt.DatatypeID == m.SelectedDatatype {
			active = " *"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, dt.Label, active))
	}
	return strings.Join(lines, "\n")
}

// renderRoutesList renders the route list with cursor and active route indicator.
func renderRoutesList(m Model) string {
	if len(m.Routes) == 0 {
		return "(no routes)"
	}

	lines := make([]string, 0, len(m.Routes))
	for i, route := range m.Routes {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		active := ""
		if route.RouteID == m.PageRouteId {
			active = " *"
		}
		lines = append(lines, fmt.Sprintf("%s %s /%s%s", cursor, route.Title, route.Slug, active))
	}
	return strings.Join(lines, "\n")
}

// renderRouteDetail renders the selected route details for the center panel.
func renderRouteDetail(m Model) string {
	if len(m.Routes) == 0 || m.Cursor >= len(m.Routes) {
		return "No route selected"
	}

	route := m.Routes[m.Cursor]
	lines := []string{
		fmt.Sprintf("Title:    %s", route.Title),
		fmt.Sprintf("Slug:     /%s", route.Slug),
		fmt.Sprintf("Status:   %d", route.Status),
		fmt.Sprintf("Author:   %s", route.AuthorID.String()),
		fmt.Sprintf("Created:  %s", route.DateCreated.String()),
		fmt.Sprintf("Modified: %s", route.DateModified.String()),
	}

	if route.RouteID == m.PageRouteId {
		lines = append(lines, "", "  (active route)")
	}

	return strings.Join(lines, "\n")
}

// renderRouteActions renders available actions for the right panel.
func renderRouteActions(m Model) string {
	lines := []string{
		"Actions",
		"",
		"  n: New",
		"  e: Edit",
		"  d: Delete",
		"",
		fmt.Sprintf("Routes: %d", len(m.Routes)),
	}

	if !m.PageRouteId.IsZero() {
		lines = append(lines, fmt.Sprintf("Active:  %s", m.PageRouteId))
	}

	return strings.Join(lines, "\n")
}

// renderMediaList renders the media list for the left panel.
func renderMediaList(m Model) string {
	if len(m.MediaList) == 0 {
		return "(no media)"
	}

	lines := make([]string, 0, len(m.MediaList))
	for i, media := range m.MediaList {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		name := media.MediaID.String()
		if media.DisplayName.Valid && media.DisplayName.String != "" {
			name = media.DisplayName.String
		} else if media.Name.Valid && media.Name.String != "" {
			name = media.Name.String
		}
		mime := ""
		if media.Mimetype.Valid && media.Mimetype.String != "" {
			mime = " [" + media.Mimetype.String + "]"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, name, mime))
	}
	return strings.Join(lines, "\n")
}

// renderMediaDetail renders the selected media details for the center panel.
func renderMediaDetail(m Model) string {
	if len(m.MediaList) == 0 || m.Cursor >= len(m.MediaList) {
		return "No media selected"
	}

	media := m.MediaList[m.Cursor]

	nullStr := func(ns sql.NullString) string {
		if ns.Valid {
			return ns.String
		}
		return "(none)"
	}

	lines := []string{
		fmt.Sprintf("Name:        %s", nullStr(media.Name)),
		fmt.Sprintf("Display:     %s", nullStr(media.DisplayName)),
		fmt.Sprintf("Alt:         %s", nullStr(media.Alt)),
		fmt.Sprintf("Caption:     %s", nullStr(media.Caption)),
		fmt.Sprintf("Description: %s", nullStr(media.Description)),
		"",
		fmt.Sprintf("Mimetype:    %s", nullStr(media.Mimetype)),
		fmt.Sprintf("Dimensions:  %s", nullStr(media.Dimensions)),
		fmt.Sprintf("URL:         %s", media.URL),
		"",
		fmt.Sprintf("Created:     %s", media.DateCreated.String()),
		fmt.Sprintf("Modified:    %s", media.DateModified.String()),
	}

	return strings.Join(lines, "\n")
}

// renderMediaInfo renders the media info summary for the right panel.
func renderMediaInfo(m Model) string {
	lines := []string{
		"Media Library",
		"",
		fmt.Sprintf("  Total: %d", len(m.MediaList)),
	}

	if len(m.MediaList) > 0 && m.Cursor < len(m.MediaList) {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  ID: %s", m.MediaList[m.Cursor].MediaID))
		if m.MediaList[m.Cursor].Class.Valid && m.MediaList[m.Cursor].Class.String != "" {
			lines = append(lines, fmt.Sprintf("  Class: %s", m.MediaList[m.Cursor].Class.String))
		}
	}

	return strings.Join(lines, "\n")
}

// renderDatatypesList renders all datatypes for the left panel on the DATATYPES page.
func renderDatatypesList(m Model) string {
	if len(m.AllDatatypes) == 0 {
		return "(no datatypes)"
	}

	lines := make([]string, 0, len(m.AllDatatypes))
	for i, dt := range m.AllDatatypes {
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

// renderDatatypeDetail renders the fields for the selected datatype in the center panel.
// Shows cursor when ContentPanel has focus.
func renderDatatypeDetail(m Model) string {
	if len(m.AllDatatypes) == 0 || m.Cursor >= len(m.AllDatatypes) {
		return "No datatype selected"
	}

	dt := m.AllDatatypes[m.Cursor]
	lines := []string{
		fmt.Sprintf("Fields for: %s", dt.Label),
		"",
	}

	if len(m.SelectedDatatypeFields) == 0 {
		// Show (empty) with cursor if focused
		if m.PanelFocus == tui.ContentPanel {
			lines = append(lines, " -> (empty)")
		} else {
			lines = append(lines, "    (empty)")
		}
		lines = append(lines, "")
		lines = append(lines, "Press 'n' to add a field")
	} else {
		for i, field := range m.SelectedDatatypeFields {
			cursor := "   "
			if m.PanelFocus == tui.ContentPanel && m.FieldCursor == i {
				cursor = " ->"
			}
			lines = append(lines, fmt.Sprintf("%s %d. %s [%s]", cursor, i+1, field.Label, field.Type))
		}
	}

	return strings.Join(lines, "\n")
}

// renderDatatypeActions renders available actions for the right panel on DATATYPES page.
// Shows context-sensitive hints based on which panel is focused.
func renderDatatypeActions(m Model) string {
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
	lines = append(lines, fmt.Sprintf("Datatypes: %d", len(m.AllDatatypes)))
	if len(m.AllDatatypes) > 0 && m.Cursor < len(m.AllDatatypes) {
		lines = append(lines, fmt.Sprintf("Fields: %d", len(m.SelectedDatatypeFields)))
	}

	return strings.Join(lines, "\n")
}
