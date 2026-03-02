package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/definitions"
	"github.com/hegner123/modulacms/internal/utility"
)

// isCMSPanelPage returns true for pages that use the 3-panel CMS layout.
func isCMSPanelPage(idx PageIndex) bool {
	switch idx {
	case CMSPAGE, ADMINCMSPAGE, CONTENT, MEDIA, USERSADMIN, ROUTES, DATATYPES, ADMINROUTES, ADMINDATATYPES, ADMINCONTENT, PLUGINSPAGE, FIELDTYPES, ADMINFIELDTYPES, DEPLOYPAGE, PIPELINESPAGE, PIPELINEDETAILPAGE, WEBHOOKSPAGE, HOMEPAGE, ACTIONSPAGE, QUICKSTARTPAGE, PLUGINDETAILPAGE, DATABASEPAGE, CONFIGPAGE, READPAGE:
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
	leftTitle, centerTitle, rightTitle := cmsPanelTitles(m)

	treePanel := Panel{
		Title:   leftTitle,
		Width:   leftW,
		Height:  bodyH,
		Content: left,
		Focused: m.PanelFocus == TreePanel,
	}

	contentPanel := Panel{
		Title:   centerTitle,
		Width:   centerW,
		Height:  bodyH,
		Content: center,
		Focused: m.PanelFocus == ContentPanel,
	}

	routePanel := Panel{
		Title:   rightTitle,
		Width:   rightW,
		Height:  bodyH,
		Content: right,
		Focused: m.PanelFocus == RoutePanel,
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		treePanel.Render(),
		contentPanel.Render(),
		routePanel.Render(),
	)

	ui := lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar)

	if m.FilePickerActive {
		return FilePickerOverlay(ui, m.FilePicker, m.Width, m.Height)
	}

	if m.ActiveOverlay != nil {
		return RenderOverlay(ui, m.ActiveOverlay, m.Width, m.Height)
	}

	return ui
}

// cmsPanelTitles returns the panel titles based on the current page context.
func cmsPanelTitles(m Model) (left, center, right string) {
	switch m.Page.Index {
	case CONTENT:
		if m.PageRouteId.IsZero() {
			return "Content", "Details", "Actions"
		}
		if m.ShowVersionList {
			return "Tree", "Content", "Versions"
		}
		return "Tree", "Content", "Fields"
	case ROUTES:
		return "Routes", "Details", "Actions"
	case MEDIA:
		return "Media", "Details", "Info"
	case DATATYPES:
		return "Datatypes", "Fields", "Actions"
	case USERSADMIN:
		return "Users", "Details", "Permissions"
	case ADMINROUTES:
		return "Admin Routes", "Details", "Actions"
	case ADMINDATATYPES:
		return "Admin Datatypes", "Fields", "Actions"
	case ADMINCONTENT:
		return "Admin Content", "Details", "Info"
	case PLUGINSPAGE:
		return "Plugins", "Details", "Info"
	case FIELDTYPES:
		return "Field Types", "Details", "Actions"
	case ADMINFIELDTYPES:
		return "Admin Field Types", "Details", "Actions"
	case DEPLOYPAGE:
		return "Environments", "Details", "Actions"
	case PIPELINESPAGE:
		return "Pipelines", "Entries", "Info"
	case PIPELINEDETAILPAGE:
		return "Pipelines", "Configuration", "Status"
	case WEBHOOKSPAGE:
		return "Webhooks", "Details", "Info"
	case HOMEPAGE:
		return "System", "Navigation", "Info"
	case ACTIONSPAGE:
		return "Actions", "Details", "Status"
	case QUICKSTARTPAGE:
		return "Schemas", "Details", "Status"
	case PLUGINDETAILPAGE:
		return "Plugin", "Actions", "Info"
	case DATABASEPAGE:
		return "Tables", "Actions", "Info"
	case CONFIGPAGE:
		return "Categories", "Fields", "Detail"
	case READPAGE:
		modeNames := []string{"Read", "Update", "Delete"}
		mode := "Read"
		if int(m.DatabaseMode) < len(modeNames) {
			mode = modeNames[m.DatabaseMode]
		}
		return "Mode", fmt.Sprintf("%s: %s", mode, m.TableState.Table), "Detail"
	default:
		return "Tree", "Content", "Route"
	}
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
			// Content selection flow: show content instances with slug and datatype
			left = renderRootContentSummaryList(m)
			center = renderContentSummaryDetail(m)
			right = renderContentActions(m)
		} else {
			// Content browsing: tree view
			cms := CMSPage{}
			left = cms.ProcessTreeDatatypes(m)
			center = cms.ProcessContentPreview(m)
			if m.ShowVersionList {
				right = renderVersionList(m)
			} else {
				right = cms.ProcessFields(m)
			}
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
		left = renderUsersList(m)
		center = renderUserDetail(m)
		right = renderUserPermissions(m)

	case DATATYPES:
		left = renderDatatypesList(m)
		center = renderDatatypeDetail(m)
		right = renderDatatypeActions(m)

	case ADMINROUTES:
		left = renderAdminRoutesList(m)
		center = renderAdminRouteDetail(m)
		right = renderAdminRouteActions(m)

	case ADMINDATATYPES:
		left = renderAdminDatatypesList(m)
		center = renderAdminDatatypeFields(m)
		right = renderAdminDatatypeActions(m)

	case ADMINCONTENT:
		left = renderAdminContentList(m)
		center = renderAdminContentDetail(m)
		right = ""

	case PLUGINSPAGE:
		left = renderPluginsList(m)
		center = renderPluginDetail(m)
		right = renderPluginInfo(m)

	case FIELDTYPES:
		left = renderFieldTypesList(m)
		center = renderFieldTypeDetail(m)
		right = renderFieldTypeActions(m)

	case ADMINFIELDTYPES:
		left = renderAdminFieldTypesList(m)
		center = renderAdminFieldTypeDetail(m)
		right = renderAdminFieldTypeActions(m)

	case DEPLOYPAGE:
		left = renderDeployEnvsList(m)
		center = renderDeployDetail(m)
		right = renderDeployActions(m)

	case PIPELINESPAGE:
		left = renderPipelinesList(m)
		center = renderPipelineDetail(m)
		right = renderPipelineInfo(m)

	case WEBHOOKSPAGE:
		left = renderWebhooksList(m)
		center = renderWebhookDetail(m)
		right = renderWebhookInfo(m)

	case HOMEPAGE:
		left = renderHomeSystem(m)
		center = renderHomeNavigation(m)
		right = renderHomeInfo(m)

	case ACTIONSPAGE:
		left = renderActionsMenu(m)
		center = renderActionsDetail(m)
		right = renderActionsStatus(m)

	case QUICKSTARTPAGE:
		left = renderQuickstartSchemas(m)
		center = renderQuickstartDetail(m)
		right = renderQuickstartStatus(m)

	case PLUGINDETAILPAGE:
		left = renderPluginDetailInfo(m)
		center = renderPluginDetailActions(m)
		right = renderPluginDetailActionInfo(m)

	case DATABASEPAGE:
		left = renderDatabaseTables(m)
		center = renderDatabaseActions(m)
		right = renderDatabaseInfo(m)

	case CONFIGPAGE:
		left = renderConfigCategories(m)
		center = renderConfigFields(m)
		right = renderConfigFieldDetail(m)

	case READPAGE:
		left = renderDatabaseMode(m)
		center = renderDatabaseTable(m)
		right = renderDatabaseRowDetail(m)

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

// renderCMSHeader renders the top bar with the app title.
func renderCMSHeader(m Model) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Accent)

	title := titleStyle.Render(RenderTitle(m.Titles[m.TitleFont]))

	container := lipgloss.NewStyle().
		Width(m.Width).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(config.DefaultStyle.Tertiary).
		Padding(0, 1).
		Render(title)

	return container
}

// renderCMSPanelStatusBar renders a two-line status bar for the CMS
// three-panel layout. Line 1: status badge, panel focus, and locale.
// Line 2: context-sensitive key hints.
func renderCMSPanelStatusBar(m Model) string {
	// High-contrast: white on black.
	barFG := config.DefaultStyle.Primary
	barBG := config.DefaultStyle.PrimaryBG

	barStyle := lipgloss.NewStyle().
		Foreground(barFG).
		Background(barBG)

	keyStyle := barStyle.Bold(true)

	// padLine fills a line to exactly m.Width so the background
	// color spans the full terminal width.
	padLine := func(content string) string {
		w := lipgloss.Width(content)
		if w >= m.Width {
			return content
		}
		return content + barStyle.Render(strings.Repeat(" ", m.Width-w))
	}

	// --- line 1: status badge + panel focus + locale ---
	statusBadge := m.GetStatus()

	// Panel tabs: highlight the focused panel with accent color.
	panels := []FocusPanel{TreePanel, ContentPanel, RoutePanel}
	focusParts := make([]string, len(panels))
	for i, p := range panels {
		label := p.String()
		if m.PanelFocus == p {
			focusParts[i] = lipgloss.NewStyle().
				Bold(true).
				Foreground(config.DefaultStyle.Accent).
				Background(barBG).
				Padding(0, 1).
				Render("[" + label + "]")
		} else {
			focusParts[i] = barStyle.Padding(0, 1).Render(" " + label + " ")
		}
	}
	focusIndicator := lipgloss.JoinHorizontal(lipgloss.Center, focusParts...)

	// Optional locale badge
	var localeBadge string
	if m.Config.I18nEnabled() && m.ActiveLocale != "" {
		localeStyle := lipgloss.NewStyle().
			Foreground(config.DefaultStyle.Accent).
			Background(barBG).
			Bold(true).
			Padding(0, 1)
		localeBadge = barStyle.Render("  ") + localeStyle.Render(strings.ToUpper(m.ActiveLocale))
	}

	line1 := statusBadge + barStyle.Render(" ") + focusIndicator + localeBadge

	// --- line 2: context-sensitive key hints, split with separators ---

	// key renders a single  key:label  pair.
	key := func(k, label string) string {
		return keyStyle.Render(k) + barStyle.Render(":"+label)
	}
	sep := barStyle.Render(" │ ")

	// Build hints from getContextControls, re-rendered with proper
	// styling by splitting the raw string on " │ ".
	rawHints := getContextControls(m)
	parts := strings.Split(rawHints, " │ ")
	styledParts := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Each part looks like "k/k:label" or "k:label".
		colonIdx := strings.Index(part, ":")
		if colonIdx > 0 {
			styledParts = append(styledParts, key(part[:colonIdx], part[colonIdx+1:]))
		} else {
			styledParts = append(styledParts, barStyle.Render(part))
		}
	}

	line2 := barStyle.Render(" ") + strings.Join(styledParts, sep)

	return padLine(line1) + "\n" + padLine(line2)
}

// getContextControls returns context-sensitive keybinding hints based on
// current page and panel focus.
func getContextControls(m Model) string {
	km := m.Config.KeyBindings
	nav := km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown) + ":nav"
	common := km.HintString(config.ActionNextPanel) + ":panel │ " +
		km.HintString(config.ActionBack) + ":back │ " +
		km.HintString(config.ActionQuit) + ":quit"

	switch m.Page.Index {
	case CONTENT:
		if m.PageRouteId.IsZero() {
			return nav + " │ enter:view │ " + km.HintString(config.ActionNew) + ":new │ " +
				km.HintString(config.ActionEdit) + ":edit │ " + common
		}
		if m.ShowVersionList && m.PanelFocus == RoutePanel {
			return nav + " │ enter:restore │ esc:close │ " + common
		}
		if m.PanelFocus == RoutePanel {
			return nav + " │ " +
				km.HintString(config.ActionEdit) + ":edit │ " +
				km.HintString(config.ActionNew) + ":add │ " +
				km.HintString(config.ActionDelete) + ":delete │ " +
				km.HintString(config.ActionReorderUp) + "/" + km.HintString(config.ActionReorderDown) + ":reorder │ " + common
		}
		localeHint := ""
		if m.Config.I18nEnabled() {
			localeHint = km.HintString(config.ActionLocale) + ":locale │ "
		}
		return nav + " │ " + km.HintString(config.ActionExpand) + "/" + km.HintString(config.ActionCollapse) + ":expand │ " +
			km.HintString(config.ActionGoParent) + "/" + km.HintString(config.ActionGoChild) + ":parent/child │ " +
			km.HintString(config.ActionNew) + ":new │ " +
			km.HintString(config.ActionEdit) + ":edit │ " +
			km.HintString(config.ActionDelete) + ":delete │ " +
			km.HintString(config.ActionReorderUp) + "/" + km.HintString(config.ActionReorderDown) + ":reorder │ " +
			km.HintString(config.ActionCopy) + ":copy │ " +
			km.HintString(config.ActionPublish) + ":publish │ " +
			km.HintString(config.ActionVersions) + ":versions │ " +
			localeHint + common

	case ROUTES:
		return nav + " │ enter:select │ " + km.HintString(config.ActionNew) + ":new │ " +
			km.HintString(config.ActionEdit) + ":edit │ " +
			km.HintString(config.ActionDelete) + ":delete │ " + common

	case DATATYPES:
		switch m.PanelFocus {
		case TreePanel:
			return nav + " │ " + km.HintString(config.ActionNew) + ":new │ " +
				km.HintString(config.ActionEdit) + ":edit │ " +
				km.HintString(config.ActionDelete) + ":delete │ " + common
		case ContentPanel:
			return nav + " │ " + km.HintString(config.ActionNew) + ":new field │ " +
				km.HintString(config.ActionEdit) + ":edit │ " +
				km.HintString(config.ActionDelete) + ":delete │ " +
				"u:ui config │ " + common
		default:
			return nav + " │ " + common
		}

	case MEDIA:
		return nav + " │ " + km.HintString(config.ActionNew) + ":upload │ " +
			km.HintString(config.ActionDelete) + ":delete │ " + common

	case USERSADMIN:
		return nav + " │ " + km.HintString(config.ActionNew) + ":new │ " +
			km.HintString(config.ActionEdit) + ":edit │ " +
			km.HintString(config.ActionDelete) + ":delete │ " + common

	case CMSPAGE, ADMINCMSPAGE:
		return nav + " │ enter:select │ " + common

	case ADMINROUTES:
		return nav + " │ " + km.HintString(config.ActionNew) + ":new │ " +
			km.HintString(config.ActionEdit) + ":edit │ " +
			km.HintString(config.ActionDelete) + ":delete │ " + common

	case ADMINDATATYPES:
		switch m.PanelFocus {
		case TreePanel:
			return nav + " │ " + km.HintString(config.ActionNew) + ":new │ " +
				km.HintString(config.ActionEdit) + ":edit │ " +
				km.HintString(config.ActionDelete) + ":delete │ " + common
		case ContentPanel:
			return nav + " │ " + km.HintString(config.ActionNew) + ":new field │ " +
				km.HintString(config.ActionEdit) + ":edit │ " +
				km.HintString(config.ActionDelete) + ":delete │ " + common
		default:
			return nav + " │ " + common
		}

	case ADMINCONTENT:
		return nav + " │ " + km.HintString(config.ActionDelete) + ":delete │ " + common

	case PLUGINSPAGE:
		return nav + " │ enter:view │ " + common

	case FIELDTYPES:
		return nav + " │ " + km.HintString(config.ActionNew) + ":new │ " +
			km.HintString(config.ActionEdit) + ":edit │ " +
			km.HintString(config.ActionDelete) + ":delete │ " + common

	case ADMINFIELDTYPES:
		return nav + " │ " + km.HintString(config.ActionNew) + ":new │ " +
			km.HintString(config.ActionEdit) + ":edit │ " +
			km.HintString(config.ActionDelete) + ":delete │ " + common

	case DEPLOYPAGE:
		return nav + " │ t:test │ p:pull │ s:push │ P:dry pull │ S:dry push │ " + common

	case PIPELINESPAGE:
		return nav + " │ enter:view │ " + common

	case HOMEPAGE:
		return nav + " │ enter:select │ " + common

	case ACTIONSPAGE:
		return nav + " │ enter:run │ " + common

	case QUICKSTARTPAGE:
		return nav + " │ enter:install │ " + common

	case PLUGINDETAILPAGE:
		return nav + " │ enter:run │ " + common

	case DATABASEPAGE:
		switch m.PanelFocus {
		case ContentPanel:
			return nav + " │ enter:run │ " + common
		default:
			return nav + " │ enter:select │ " + common
		}

	case CONFIGPAGE:
		switch m.PanelFocus {
		case ContentPanel:
			if m.ConfigCategory == "raw_json" {
				return "↑↓/pgup/pgdn:scroll │ " + common
			}
			return nav + " │ " + km.HintString(config.ActionEdit) + ":edit │ " + common
		default:
			return nav + " │ enter:select │ " + common
		}

	case READPAGE:
		switch m.PanelFocus {
		case TreePanel:
			return nav + " │ enter:select │ " + common
		case ContentPanel:
			pageHint := km.HintString(config.ActionPagePrev) + "/" + km.HintString(config.ActionPageNext) + ":page"
			switch m.DatabaseMode {
			case DBModeUpdate:
				return nav + " │ enter:edit │ " + pageHint + " │ " + common
			case DBModeDelete:
				return nav + " │ enter:delete │ " + pageHint + " │ " + common
			default:
				return nav + " │ " + pageHint + " │ " + common
			}
		default:
			return common
		}

	default:
		return common
	}
}

// renderRootDatatypesList renders _root datatypes for the left panel on the CONTENT page.
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
		lines = append(lines, fmt.Sprintf("%s %s %s%s", cursor, route.Title, route.Slug, active))
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
		fmt.Sprintf("Slug:     %s", route.Slug),
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

	nullStr := func(ns db.NullString) string {
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
		if m.PanelFocus == ContentPanel {
			lines = append(lines, " -> (empty)")
		} else {
			lines = append(lines, "    (empty)")
		}
		lines = append(lines, "")
		lines = append(lines, "Press 'n' to add a field")
	} else {
		for i, field := range m.SelectedDatatypeFields {
			cursor := "   "
			if m.PanelFocus == ContentPanel && m.FieldCursor == i {
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
	case TreePanel:
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
	case ContentPanel:
		lines = append(lines,
			"Fields Panel",
			"",
			"  n: New field",
			"  e: Edit field",
			"  d: Delete field",
			"  u: UI Config",
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

// renderRootContentSummaryList renders content data instances with slug and _root datatype label for the left panel.
func renderRootContentSummaryList(m Model) string {
	if len(m.RootContentSummary) == 0 {
		return "(no content)"
	}

	lines := make([]string, 0, len(m.RootContentSummary))
	for i, content := range m.RootContentSummary {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s [%s] %s", cursor, content.DatatypeLabel, content.RouteSlug))
	}
	return strings.Join(lines, "\n")
}

// renderContentSummaryDetail renders details for the selected content summary in the center panel.
func renderContentSummaryDetail(m Model) string {
	if len(m.RootContentSummary) == 0 || m.Cursor >= len(m.RootContentSummary) {
		return "No content selected"
	}

	content := m.RootContentSummary[m.Cursor]
	lines := []string{
		fmt.Sprintf("Route:    %s", content.RouteSlug),
		fmt.Sprintf("Title:    %s", content.RouteTitle),
		fmt.Sprintf("Datatype: %s", content.DatatypeLabel),
		"",
		fmt.Sprintf("ID:       %s", content.ContentDataID),
		fmt.Sprintf("Route ID: %s", content.RouteID.String()),
		"",
		fmt.Sprintf("Created:  %s", content.DateCreated.String()),
		fmt.Sprintf("Modified: %s", content.DateModified.String()),
	}

	return strings.Join(lines, "\n")
}

// renderContentActions renders available actions for the right panel on the CONTENT page.
func renderContentActions(m Model) string {
	lines := []string{
		"Actions",
		"",
		"  enter: View content tree",
		"  n: New content",
		"  e: Edit",
		"  d: Delete",
		"",
		fmt.Sprintf("Content: %d", len(m.RootContentSummary)),
	}

	return strings.Join(lines, "\n")
}

// renderUsersList renders the list of users for the left panel.
func renderUsersList(m Model) string {
	if len(m.UsersList) == 0 {
		return "No users found"
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Accent).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary)

	var lines []string
	for i, user := range m.UsersList {
		prefix := "  "
		style := normalStyle
		if i == m.Cursor {
			prefix = "> "
			style = selectedStyle
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s%s (%s)", prefix, user.Username, user.RoleLabel)))
	}

	return strings.Join(lines, "\n")
}

// renderUserDetail renders the details of the selected user for the center panel.
func renderUserDetail(m Model) string {
	if len(m.UsersList) == 0 || m.Cursor >= len(m.UsersList) {
		return "Select a user"
	}

	user := m.UsersList[m.Cursor]
	labelStyle := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Accent).
		Bold(true)

	lines := []string{
		labelStyle.Render("User Details"),
		"",
		fmt.Sprintf("  Username:  %s", user.Username),
		fmt.Sprintf("  Name:      %s", user.Name),
		fmt.Sprintf("  Email:     %s", user.Email),
		fmt.Sprintf("  Role:      %s", user.RoleLabel),
		"",
		fmt.Sprintf("  ID:        %s", user.UserID),
		fmt.Sprintf("  Created:   %s", user.DateCreated),
		fmt.Sprintf("  Modified:  %s", user.DateModified),
	}

	return strings.Join(lines, "\n")
}

// renderUserPermissions renders the permissions panel for the selected user.
func renderUserPermissions(m Model) string {
	if len(m.UsersList) == 0 || m.Cursor >= len(m.UsersList) {
		return "Permissions\n\n  (none)"
	}

	user := m.UsersList[m.Cursor]
	lines := []string{
		"Permissions",
		"",
		fmt.Sprintf("  Role: %s", user.RoleLabel),
		"",
		fmt.Sprintf("  Users: %d", len(m.UsersList)),
	}

	return strings.Join(lines, "\n")
}

// renderPluginsList renders the plugin list for the left panel on the PLUGINSPAGE.
func renderPluginsList(m Model) string {
	if len(m.PluginsList) == 0 {
		return "(no plugins)"
	}

	lines := make([]string, 0, len(m.PluginsList))
	for i, p := range m.PluginsList {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		stateIndicator := p.State
		if p.CBState == "open" {
			stateIndicator = "tripped"
		}
		drift := ""
		if p.ManifestDrift || p.CapabilityDrifts > 0 || p.SchemaDrifts > 0 {
			drift = " [drift]"
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]%s", cursor, p.Name, stateIndicator, drift))
	}
	return strings.Join(lines, "\n")
}

// renderPluginDetail renders the selected plugin details for the center panel.
func renderPluginDetail(m Model) string {
	if len(m.PluginsList) == 0 || m.Cursor >= len(m.PluginsList) {
		return "No plugin selected"
	}

	p := m.PluginsList[m.Cursor]
	lines := []string{
		fmt.Sprintf("Name:        %s", p.Name),
		fmt.Sprintf("Version:     %s", p.Version),
		fmt.Sprintf("State:       %s", p.State),
		fmt.Sprintf("Circuit:     %s", p.CBState),
		"",
	}
	if p.Description != "" {
		lines = append(lines, fmt.Sprintf("Description: %s", p.Description))
	}

	// Drift warnings
	if p.ManifestDrift || p.CapabilityDrifts > 0 || p.SchemaDrifts > 0 {
		lines = append(lines, "", "--- Drift Detected ---")
		if p.ManifestDrift {
			lines = append(lines, "  Manifest changed (hash differs from installed)")
		}
		if p.CapabilityDrifts > 0 {
			lines = append(lines, fmt.Sprintf("  %d capability change(s) (run sync to update)", p.CapabilityDrifts))
		}
		if p.SchemaDrifts > 0 {
			lines = append(lines, fmt.Sprintf("  %d schema drift entr(ies)", p.SchemaDrifts))
		}
	}

	return strings.Join(lines, "\n")
}

// renderPluginInfo renders the plugin info summary for the right panel.
func renderPluginInfo(m Model) string {
	lines := []string{
		"Plugin Manager",
		"",
		fmt.Sprintf("  Total: %d", len(m.PluginsList)),
	}

	// Count by state
	running := 0
	failed := 0
	stopped := 0
	for _, p := range m.PluginsList {
		switch p.State {
		case "running":
			running++
		case "failed":
			failed++
		case "stopped":
			stopped++
		}
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Running: %d", running))
	if failed > 0 {
		lines = append(lines, fmt.Sprintf("  Failed:  %d", failed))
	}
	if stopped > 0 {
		lines = append(lines, fmt.Sprintf("  Stopped: %d", stopped))
	}

	return strings.Join(lines, "\n")
}

// renderPipelinesList renders the pipeline chain list for the left panel on PIPELINESPAGE.
func renderPipelinesList(m Model) string {
	if len(m.PipelinesList) == 0 {
		return "(no pipeline chains)"
	}

	lines := make([]string, 0, len(m.PipelinesList))
	for i, p := range m.PipelinesList {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s.%s_%s (%d)", cursor, p.Table, p.Phase, p.Operation, p.Count))
	}
	return strings.Join(lines, "\n")
}

// renderPipelineDetail renders the entries for the selected pipeline chain (center panel).
func renderPipelineDetail(m Model) string {
	if len(m.PipelinesList) == 0 || m.Cursor >= len(m.PipelinesList) {
		return "No pipeline selected"
	}

	if len(m.PipelineEntries) == 0 {
		return "No entries (select a chain to view)"
	}

	lines := make([]string, 0, len(m.PipelineEntries)+2)
	lines = append(lines, "Pipeline entries (priority order):")
	lines = append(lines, "")
	for i, e := range m.PipelineEntries {
		enabled := "on"
		if !e.Enabled {
			enabled = "off"
		}
		lines = append(lines, fmt.Sprintf("  %d. %s -> %s (pri:%d, %s)", i+1, e.PluginName, e.Handler, e.Priority, enabled))
	}
	return strings.Join(lines, "\n")
}

// renderVersionList renders the version history list for the right panel on CONTENT page.
func renderVersionList(m Model) string {
	if len(m.Versions) == 0 {
		return "(no versions)\n\nPress esc to close"
	}

	publishedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#16a34a")).Bold(true)
	triggerStyle := lipgloss.NewStyle().Faint(true)
	labelStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	lines := make([]string, 0, len(m.Versions)*3+2)
	for i, v := range m.Versions {
		cursor := "   "
		if m.PanelFocus == RoutePanel && m.VersionCursor == i {
			cursor = " > "
		}

		// Version number + published badge
		numStr := fmt.Sprintf("#%d", v.VersionNumber)
		if v.Published {
			numStr = publishedStyle.Render(numStr + " [pub]")
		}

		// Trigger badge
		trigger := triggerStyle.Render(v.Trigger)

		// Label (if present)
		label := ""
		if v.Label != "" {
			label = " " + labelStyle.Render(v.Label)
		}

		// Date
		date := triggerStyle.Render(v.DateCreated.String())

		lines = append(lines, fmt.Sprintf("%s%s %s%s", cursor, numStr, trigger, label))
		lines = append(lines, fmt.Sprintf("     %s", date))
	}

	lines = append(lines, "")
	lines = append(lines, triggerStyle.Render("  enter:restore │ esc:close"))

	return strings.Join(lines, "\n")
}

// renderPipelineInfo renders the pipeline summary for the right panel on PIPELINESPAGE.
func renderPipelineInfo(m Model) string {
	totalEntries := 0
	for _, p := range m.PipelinesList {
		totalEntries += p.Count
	}

	lines := []string{
		"Pipeline Registry",
		"",
		fmt.Sprintf("  Chains: %d", len(m.PipelinesList)),
		fmt.Sprintf("  Total entries: %d", totalEntries),
	}

	// Count by phase
	before := 0
	after := 0
	for _, p := range m.PipelinesList {
		switch p.Phase {
		case "before":
			before++
		case "after":
			after++
		}
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Before chains: %d", before))
	lines = append(lines, fmt.Sprintf("  After chains:  %d", after))

	return strings.Join(lines, "\n")
}

// renderWebhooksList renders the webhook list for the left panel on the WEBHOOKSPAGE.
func renderWebhooksList(m Model) string {
	if len(m.WebhooksList) == 0 {
		return "(no webhooks)"
	}

	lines := make([]string, 0, len(m.WebhooksList))
	for i, wh := range m.WebhooksList {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		status := "off"
		if wh.IsActive {
			status = "on"
		}
		lines = append(lines, fmt.Sprintf("%s %s [%s]", cursor, wh.Name, status))
	}
	return strings.Join(lines, "\n")
}

// renderWebhookDetail renders the selected webhook details for the center panel.
func renderWebhookDetail(m Model) string {
	if len(m.WebhooksList) == 0 || m.Cursor >= len(m.WebhooksList) {
		return "No webhook selected"
	}

	wh := m.WebhooksList[m.Cursor]
	active := "No"
	if wh.IsActive {
		active = "Yes"
	}
	lines := []string{
		fmt.Sprintf("Name:     %s", wh.Name),
		fmt.Sprintf("URL:      %s", wh.URL),
		fmt.Sprintf("Active:   %s", active),
		fmt.Sprintf("Events:   %s", strings.Join(wh.Events, ", ")),
		"",
		fmt.Sprintf("Created:  %s", wh.DateCreated.String()),
		fmt.Sprintf("Modified: %s", wh.DateModified.String()),
	}
	return strings.Join(lines, "\n")
}

// renderWebhookInfo renders the webhook summary for the right panel.
func renderWebhookInfo(m Model) string {
	active := 0
	for _, wh := range m.WebhooksList {
		if wh.IsActive {
			active++
		}
	}
	lines := []string{
		"Webhook Manager",
		"",
		fmt.Sprintf("  Total:  %d", len(m.WebhooksList)),
		fmt.Sprintf("  Active: %d", active),
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// HOMEPAGE panel renders
// ---------------------------------------------------------------------------

// renderHomeSystem renders system info for the left panel on HOMEPAGE.
func renderHomeSystem(m Model) string {
	lines := []string{
		"System Info",
		"",
		fmt.Sprintf("  Version:  %s", utility.Version),
		fmt.Sprintf("  Database: %s", m.Config.Db_Driver),
		fmt.Sprintf("  User:     %s", m.AdminUsername),
	}
	return strings.Join(lines, "\n")
}

// renderHomeNavigation renders the main menu list for the center panel on HOMEPAGE.
func renderHomeNavigation(m Model) string {
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

// renderHomeInfo renders keyboard hints for the right panel on HOMEPAGE.
func renderHomeInfo(m Model) string {
	lines := []string{
		"Keyboard",
		"",
		"  up/down   Navigate",
		"  enter     Select",
		"  tab       Panel",
		"  q         Quit",
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// ACTIONSPAGE panel renders
// ---------------------------------------------------------------------------

// renderActionsMenu renders the action items list for the left panel on ACTIONSPAGE.
func renderActionsMenu(m Model) string {
	actions := ActionsMenu()
	if len(actions) == 0 {
		return "(no actions)"
	}
	warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
	lines := make([]string, 0, len(actions))
	for i, action := range actions {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		label := action.Label
		if action.Destructive {
			label = warnStyle.Render(label)
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, label))
	}
	return strings.Join(lines, "\n")
}

// renderActionsDetail renders the description of the hovered action for the center panel.
func renderActionsDetail(m Model) string {
	actions := ActionsMenu()
	if len(actions) == 0 || m.Cursor >= len(actions) {
		return "No action selected"
	}
	action := actions[m.Cursor]
	return fmt.Sprintf("%s\n\n%s", action.Label, action.Description)
}

// renderActionsStatus renders action count and destructive warning for the right panel.
func renderActionsStatus(m Model) string {
	actions := ActionsMenu()
	lines := []string{
		"Actions",
		"",
		fmt.Sprintf("  Total: %d", len(actions)),
	}
	if m.Cursor < len(actions) && actions[m.Cursor].Destructive {
		lines = append(lines, "")
		warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
		lines = append(lines, warnStyle.Render("  !! Destructive"))
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// QUICKSTARTPAGE panel renders
// ---------------------------------------------------------------------------

// renderQuickstartSchemas renders the definition list for the left panel on QUICKSTARTPAGE.
func renderQuickstartSchemas(m Model) string {
	labels := QuickstartMenuLabels()
	if len(labels) == 0 {
		return "(no schemas)"
	}
	lines := make([]string, 0, len(labels))
	for i, label := range labels {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, label))
	}
	return strings.Join(lines, "\n")
}

// renderQuickstartDetail renders definition description and metadata for the center panel.
func renderQuickstartDetail(m Model) string {
	defs := definitions.List()
	if len(defs) == 0 || m.Cursor >= len(defs) {
		return "No schema selected"
	}
	def := defs[m.Cursor]
	lines := []string{
		fmt.Sprintf("Name: %s", def.Label),
		fmt.Sprintf("Slug: %s", def.Name),
		"",
		def.Description,
	}
	return strings.Join(lines, "\n")
}

// renderQuickstartStatus renders help text for the right panel on QUICKSTARTPAGE.
func renderQuickstartStatus(m Model) string {
	lines := []string{
		"Quickstart",
		"",
		"  Install a predefined",
		"  schema definition to",
		"  quickly set up your",
		"  content structure.",
		"",
		"  Press enter to install",
		"  the selected schema.",
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// PLUGINDETAILPAGE panel renders
// ---------------------------------------------------------------------------

// renderPluginDetailInfo renders plugin info for the left panel on PLUGINDETAILPAGE.
func renderPluginDetailInfo(m Model) string {
	if m.SelectedPlugin == "" {
		return "No plugin selected"
	}
	var found *PluginDisplay
	for i := range m.PluginsList {
		if m.PluginsList[i].Name == m.SelectedPlugin {
			found = &m.PluginsList[i]
			break
		}
	}
	if found == nil {
		return fmt.Sprintf("Plugin: %s\n\n  (not found)", m.SelectedPlugin)
	}
	lines := []string{
		fmt.Sprintf("Name:    %s", found.Name),
		fmt.Sprintf("Version: %s", found.Version),
		fmt.Sprintf("State:   %s", found.State),
		fmt.Sprintf("Circuit: %s", found.CBState),
	}
	if found.ManifestDrift || found.CapabilityDrifts > 0 || found.SchemaDrifts > 0 {
		lines = append(lines, "", "Drift detected")
	}
	return strings.Join(lines, "\n")
}

// renderPluginDetailActions renders the 6-item action menu for the center panel on PLUGINDETAILPAGE.
func renderPluginDetailActions(m Model) string {
	actions := []string{
		"Enable Plugin",
		"Disable Plugin",
		"Reload Plugin",
		"Approve Routes",
		"Approve Hooks",
		"Sync Capabilities",
	}
	lines := make([]string, 0, len(actions))
	for i, action := range actions {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, action))
	}
	return strings.Join(lines, "\n")
}

// renderPluginDetailActionInfo renders a description per hovered action for the right panel.
func renderPluginDetailActionInfo(m Model) string {
	descriptions := []string{
		"Enable this plugin to activate its features",
		"Disable this plugin to deactivate its features",
		"Reload the plugin to apply configuration changes",
		"Review and approve registered routes",
		"Review and approve registered hooks",
		"Sync capabilities with the plugin manifest",
	}
	if m.Cursor < 0 || m.Cursor >= len(descriptions) {
		return ""
	}
	return descriptions[m.Cursor]
}

// ---------------------------------------------------------------------------
// DATABASEPAGE panel renders (merged with former TABLEPAGE)
// ---------------------------------------------------------------------------

// renderDatabaseTables renders the table list for the left panel on DATABASEPAGE.
func renderDatabaseTables(m Model) string {
	if len(m.Tables) == 0 {
		return "(no tables)"
	}
	lines := make([]string, 0, len(m.Tables))
	for i, tbl := range m.Tables {
		cursor := "   "
		if m.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, tbl))
	}
	return strings.Join(lines, "\n")
}

// renderDatabaseActions renders the CRUD action menu for the center panel on DATABASEPAGE.
func renderDatabaseActions(m Model) string {
	if m.TableState.Table == "" {
		return "Select a table"
	}
	actions := []string{"Create", "Read", "Update", "Delete"}
	lines := []string{
		fmt.Sprintf("Table: %s", m.TableState.Table),
		"",
	}
	for i, action := range actions {
		cursor := "   "
		if m.PanelFocus == ContentPanel && m.FieldCursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, action))
	}
	return strings.Join(lines, "\n")
}

// renderDatabaseInfo renders column metadata for the right panel on DATABASEPAGE.
func renderDatabaseInfo(m Model) string {
	if m.TableState.Table == "" {
		return "Database\n\n  Select a table to\n  view its columns."
	}
	lines := []string{
		fmt.Sprintf("Table: %s", m.TableState.Table),
		"",
	}
	if len(m.TableState.Headers) > 0 {
		lines = append(lines, "Columns:")
		lines = append(lines, "")
		for i, h := range m.TableState.Headers {
			lines = append(lines, fmt.Sprintf("  %d. %s", i+1, h))
		}
	} else {
		lines = append(lines, "  (no column info)")
	}
	return strings.Join(lines, "\n")
}

// --- CONFIGPAGE panel render functions ---

// renderConfigCategories renders the category list for the left panel on CONFIGPAGE.
func renderConfigCategories(m Model) string {
	categories := config.AllCategories()
	items := make([]string, 0, len(categories)+1)
	for _, cat := range categories {
		items = append(items, config.CategoryLabel(cat))
	}
	items = append(items, "View Raw JSON")

	lines := make([]string, 0, len(items))
	for i, label := range items {
		cursor := "   "
		if m.PanelFocus == TreePanel && m.Cursor == i {
			cursor = " ->"
		}
		// Highlight active category
		active := ""
		if i < len(categories) && categories[i] == m.ConfigCategory {
			active = " *"
		}
		if i == len(items)-1 && m.ConfigCategory == "raw_json" {
			active = " *"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, label, active))
	}
	return strings.Join(lines, "\n")
}

// renderConfigFields renders the config fields for the center panel on CONFIGPAGE.
func renderConfigFields(m Model) string {
	if m.ConfigCategory == "" {
		return "Select a category"
	}

	if m.ConfigCategory == "raw_json" {
		return m.Viewport.View()
	}

	if len(m.ConfigCategoryFields) == 0 {
		return "(no fields)"
	}

	title := config.CategoryLabel(m.ConfigCategory)
	labelStyle := lipgloss.NewStyle().Bold(true)
	cursorStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	lines := []string{labelStyle.Render(title), ""}

	for i, field := range m.ConfigCategoryFields {
		value := config.ConfigFieldString(*m.Config, field.JSONKey)
		if field.Sensitive && value != "" {
			value = "********"
		}

		restartMark := ""
		if !field.HotReloadable {
			restartMark = lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn).Render(" [restart]")
		}

		if m.PanelFocus == ContentPanel && i == m.ConfigFieldCursor {
			lines = append(lines, cursorStyle.Render("> ")+labelStyle.Render(field.Label)+restartMark)
			lines = append(lines, fmt.Sprintf("    %s", value))
		} else {
			lines = append(lines, fmt.Sprintf("  %s%s", field.Label, restartMark))
			lines = append(lines, fmt.Sprintf("    %s", value))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderConfigFieldDetail renders the detail view for the right panel on CONFIGPAGE.
func renderConfigFieldDetail(m Model) string {
	if m.ConfigCategory == "" {
		return "Config\n\n  Select a category to\n  view its fields."
	}

	if m.ConfigCategory == "raw_json" {
		return "Raw JSON\n\n  Scroll with ↑↓\n  or pgup/pgdn."
	}

	if len(m.ConfigCategoryFields) == 0 || m.ConfigFieldCursor >= len(m.ConfigCategoryFields) {
		return ""
	}

	field := m.ConfigCategoryFields[m.ConfigFieldCursor]
	value := config.ConfigFieldString(*m.Config, field.JSONKey)
	if field.Sensitive && value != "" {
		value = "********"
	}

	lines := []string{
		fmt.Sprintf("Field: %s", field.Label),
		fmt.Sprintf("Key:   %s", field.JSONKey),
		"",
		fmt.Sprintf("Value: %s", value),
		"",
	}

	if field.Sensitive {
		lines = append(lines, "  (sensitive)")
	}
	if field.HotReloadable {
		lines = append(lines, "  Hot-reloadable")
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn).Render("  Requires restart"))
	}

	lines = append(lines, "")
	lines = append(lines, "  Press e to edit")

	return strings.Join(lines, "\n")
}

// renderDatabaseMode renders the mode selector for the left panel on READPAGE.
func renderDatabaseMode(m Model) string {
	modes := []struct {
		label string
		mode  DatabaseMode
	}{
		{"Read", DBModeRead},
		{"Update", DBModeUpdate},
		{"Delete", DBModeDelete},
	}

	lines := make([]string, 0, len(modes)+2)
	for i, mode := range modes {
		cursor := "   "
		if m.PanelFocus == TreePanel && m.FieldCursor == i {
			cursor = " ->"
		}
		active := ""
		if m.DatabaseMode == mode.mode {
			active = " *"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, mode.label, active))
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Rows: %d", len(m.TableState.Rows)))
	if len(m.TableState.Rows) > m.MaxRows {
		totalPages := (len(m.TableState.Rows)-1)/m.MaxRows + 1
		lines = append(lines, fmt.Sprintf("  Page: %d/%d", m.PageMod+1, totalPages))
	}

	return strings.Join(lines, "\n")
}

// renderDatabaseTable renders paginated table rows for the center panel on READPAGE.
func renderDatabaseTable(m Model) string {
	if m.Loading {
		return fmt.Sprintf("\n   %s Loading...\n", m.Spinner.View())
	}
	if len(m.TableState.Headers) == 0 {
		return "No data loaded"
	}

	// Calculate page bounds
	start := m.PageMod * m.MaxRows
	end := start + m.MaxRows
	if end > len(m.TableState.Rows) {
		end = len(m.TableState.Rows)
	}
	if start >= len(m.TableState.Rows) {
		return "No rows on this page"
	}

	currentView := m.TableState.Rows[start:end]

	// Build simple text table with cursor
	lines := make([]string, 0, len(currentView)+2)

	// Header row (truncate each header to fit)
	headerLine := "   "
	for _, h := range m.TableState.Headers {
		if len(h) > 15 {
			h = h[:12] + "..."
		}
		headerLine += fmt.Sprintf("%-16s", h)
	}
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(headerLine))
	lines = append(lines, "")

	for i, row := range currentView {
		cursor := "   "
		if m.PanelFocus == ContentPanel && m.Cursor == i {
			cursor = " ->"
		}
		rowLine := cursor
		for _, cell := range row {
			if len(cell) > 15 {
				cell = cell[:12] + "..."
			}
			rowLine += fmt.Sprintf("%-16s", cell)
		}
		lines = append(lines, rowLine)
	}

	// Pagination indicator
	if len(m.TableState.Rows) > m.MaxRows {
		totalPages := (len(m.TableState.Rows)-1)/m.MaxRows + 1
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  Page %d of %d", m.PageMod+1, totalPages))
	}

	return strings.Join(lines, "\n")
}

// renderDatabaseRowDetail renders key-value detail of the selected row for the right panel on READPAGE.
func renderDatabaseRowDetail(m Model) string {
	if len(m.TableState.Rows) == 0 || len(m.TableState.Headers) == 0 {
		return "No row selected"
	}

	// Calculate actual row index from page offset + cursor
	rowIndex := (m.PageMod * m.MaxRows) + m.Cursor
	if rowIndex >= len(m.TableState.Rows) {
		return "No row selected"
	}

	row := m.TableState.Rows[rowIndex]
	lines := make([]string, 0, len(m.TableState.Headers)+4)
	lines = append(lines, fmt.Sprintf("Row %d", rowIndex))
	lines = append(lines, "")

	for i, header := range m.TableState.Headers {
		value := ""
		if i < len(row) {
			value = row[i]
		}
		lines = append(lines, fmt.Sprintf("%s:", header))
		if len(value) > 40 {
			// Wrap long values
			for len(value) > 40 {
				lines = append(lines, fmt.Sprintf("  %s", value[:40]))
				value = value[40:]
			}
			if len(value) > 0 {
				lines = append(lines, fmt.Sprintf("  %s", value))
			}
		} else {
			lines = append(lines, fmt.Sprintf("  %s", value))
		}
	}

	// Mode-specific hint
	lines = append(lines, "")
	switch m.DatabaseMode {
	case DBModeUpdate:
		lines = append(lines, "  enter: Edit this row")
	case DBModeDelete:
		warnStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
		lines = append(lines, warnStyle.Render("  enter: Delete this row"))
	}

	return strings.Join(lines, "\n")
}
