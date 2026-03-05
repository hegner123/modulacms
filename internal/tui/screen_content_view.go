package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// View renders the ContentScreen content based on current state.
func (s *ContentScreen) View(ctx AppContext) string {
	var left, center, right string
	var leftTotal, leftCursor int
	var rightTotal, rightCursor int

	if s.inTreePhase() {
		// Tree browsing phase
		left = s.renderTree()
		center = s.renderContentPreview(ctx)
		leftTotal = s.visibleNodeCount()
		leftCursor = s.Cursor
		if s.ShowVersionList {
			right = s.renderVersionList()
			rightTotal = s.versionListLen()
			rightCursor = s.VersionCursor
		} else {
			right = s.renderFields()
			rightTotal = s.fieldsLen()
			rightCursor = s.FieldCursor
		}
	} else if s.AdminMode {
		// Admin route list phase
		left = s.renderAdminContentList()
		center = s.renderAdminContentDetail()
		right = s.renderRouteActions()
		leftTotal = len(s.AdminRootContentSummary)
		leftCursor = s.Cursor
	} else {
		// Regular route list phase
		left = s.renderRouteList()
		center = s.renderRouteDetail()
		right = s.renderRouteActions()
		leftTotal = len(s.RootContentSummary)
		leftCursor = s.Cursor
	}

	// Error display in right panel
	if s.LastError != nil {
		right = s.renderError() + "\n\n" + right
	}

	layout := layoutForPage(s.PageIndex())
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: s.panelTitle(0, layout), Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel, TotalLines: leftTotal, ScrollOffset: ClampScroll(leftCursor, leftTotal, innerH)}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: s.panelTitle(1, layout), Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: s.panelTitle(2, layout), Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel, TotalLines: rightTotal, ScrollOffset: ClampScroll(rightCursor, rightTotal, innerH)}.Render())
	}

	return strings.Join(panels, "")
}

// panelTitle returns a dynamic panel title based on screen state.
func (s *ContentScreen) panelTitle(panel int, layout PageLayout) string {
	if s.inTreePhase() {
		switch panel {
		case 0:
			return "Tree"
		case 1:
			return "Content"
		case 2:
			if s.ShowVersionList {
				return "Versions"
			}
			return "Fields"
		}
	}
	return layout.Titles[panel]
}

// =============================================================================
// ROUTE LIST RENDERING (regular mode)
// =============================================================================

func (s *ContentScreen) renderRouteList() string {
	if len(s.RootContentSummary) == 0 {
		return "(no content)"
	}

	lines := make([]string, 0, len(s.RootContentSummary))
	for i, content := range s.RootContentSummary {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s [%s] %s", cursor, content.DatatypeLabel, content.RouteSlug))
	}
	return strings.Join(lines, "\n")
}

func (s *ContentScreen) renderRouteDetail() string {
	if len(s.RootContentSummary) == 0 || s.Cursor >= len(s.RootContentSummary) {
		return "No content selected"
	}

	content := s.RootContentSummary[s.Cursor]
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

func (s *ContentScreen) renderRouteActions() string {
	if s.inTreePhase() {
		return s.renderTreeActions()
	}

	if s.AdminMode {
		lines := []string{
			"Actions",
			"",
			"  enter: View content tree",
			"  d: Delete",
			"",
			fmt.Sprintf("Content: %d", len(s.AdminRootContentSummary)),
		}
		return strings.Join(lines, "\n")
	}

	lines := []string{
		"Actions",
		"",
		"  enter: View content tree",
		"  n: New content",
		"  e: Edit",
		"  d: Delete",
		"",
		fmt.Sprintf("Content: %d", len(s.RootContentSummary)),
	}
	return strings.Join(lines, "\n")
}

// =============================================================================
// ADMIN CONTENT LIST RENDERING
// =============================================================================

func (s *ContentScreen) renderAdminContentList() string {
	if len(s.AdminRootContentSummary) == 0 {
		return "(no admin content)"
	}

	lines := make([]string, 0, len(s.AdminRootContentSummary))
	for i, content := range s.AdminRootContentSummary {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		label := string(content.AdminContentDataID)
		if content.DatatypeLabel != "" {
			label = fmt.Sprintf("[%s] %s", content.DatatypeLabel, content.RouteSlug)
		} else if content.AdminDatatypeID.Valid {
			label = fmt.Sprintf("[%s] %s", content.AdminDatatypeID.ID, label)
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, label))
	}
	return strings.Join(lines, "\n")
}

func (s *ContentScreen) renderAdminContentDetail() string {
	if len(s.AdminRootContentSummary) == 0 || s.Cursor >= len(s.AdminRootContentSummary) {
		return "No admin content selected"
	}

	content := s.AdminRootContentSummary[s.Cursor]
	lines := []string{
		fmt.Sprintf("ID:        %s", content.AdminContentDataID),
		fmt.Sprintf("Route:     %s", content.RouteSlug),
		fmt.Sprintf("Title:     %s", content.RouteTitle),
		fmt.Sprintf("Status:    %s", content.Status),
		fmt.Sprintf("Author:    %s", content.AuthorName),
		"",
		fmt.Sprintf("Created:   %s", content.DateCreated.String()),
		fmt.Sprintf("Modified:  %s", content.DateModified.String()),
	}

	if content.DatatypeLabel != "" {
		lines = append([]string{fmt.Sprintf("Datatype:  %s", content.DatatypeLabel)}, lines...)
	} else if content.AdminDatatypeID.Valid {
		lines = append([]string{fmt.Sprintf("Datatype:  %s", content.AdminDatatypeID.ID)}, lines...)
	}

	return strings.Join(lines, "\n")
}

// =============================================================================
// TREE RENDERING
// =============================================================================

func (s *ContentScreen) renderTree() string {
	if s.Root.Root == nil {
		return "No content loaded"
	}
	display := make([]string, 0)
	currentIndex := 0
	s.traverseTreeWithDepth(s.Root.Root, &display, s.Cursor, &currentIndex, 0)
	if len(display) == 0 {
		return "Empty content tree"
	}
	return lipgloss.JoinVertical(lipgloss.Top, display...)
}

func (s *ContentScreen) traverseTreeWithDepth(node *tree.Node, display *[]string, cursor int, currentIndex *int, depth int) {
	if node == nil {
		return
	}

	row := s.formatTreeRow(node, *currentIndex == cursor, depth)
	*display = append(*display, row)
	*currentIndex++

	if node.Expand && node.FirstChild != nil {
		s.traverseTreeWithDepth(node.FirstChild, display, cursor, currentIndex, depth+1)
	}
	if node.NextSibling != nil {
		s.traverseTreeWithDepth(node.NextSibling, display, cursor, currentIndex, depth)
	}
}

func (s *ContentScreen) formatTreeRow(node *tree.Node, isSelected bool, depth int) string {
	indent := strings.Repeat("  ", depth)

	icon := "├─"
	if node.FirstChild != nil {
		if node.Expand {
			icon = "▼"
		} else {
			icon = "▶"
		}
	}

	name := DecideNodeName(*node)

	cursorMark := "  "
	if isSelected {
		cursorMark = "->"
	}

	statusMark := ""
	if node.Instance != nil {
		if node.Instance.Status == types.ContentStatusPublished {
			statusMark = lipgloss.NewStyle().Foreground(lipgloss.Color("#16a34a")).Render("● ")
		} else {
			statusMark = lipgloss.NewStyle().Foreground(lipgloss.Color("#ca8a04")).Render("○ ")
		}
	}

	return cursorMark + indent + icon + " " + statusMark + name
}

// =============================================================================
// CONTENT PREVIEW RENDERING
// =============================================================================

func (s *ContentScreen) renderContentPreview(ctx AppContext) string {
	node := s.Root.NodeAtIndex(s.Cursor)
	if node == nil {
		return "No content selected"
	}

	preview := []string{}

	title := DecideNodeName(*node)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent2)
	preview = append(preview, titleStyle.Render(title))
	preview = append(preview, "")

	metaParts := []string{node.Datatype.Label}
	metaParts = append(metaParts, string(node.Instance.Status))
	if !node.Instance.AuthorID.IsZero() {
		metaParts = append(metaParts, string(node.Instance.AuthorID))
	}
	dimStyle := lipgloss.NewStyle().Faint(true)
	preview = append(preview, dimStyle.Render(strings.Join(metaParts, " | ")))
	preview = append(preview, "")

	// Field values preview
	fields := s.currentFields()
	if len(fields) > 0 {
		labelStyle := lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent)
		for _, cf := range fields {
			preview = append(preview, labelStyle.Render(cf.label))
			if cf.value == "" {
				preview = append(preview, dimStyle.Render("  (empty)"))
			} else {
				lines := strings.Split(cf.value, "\n")
				for _, line := range lines {
					preview = append(preview, fmt.Sprintf("  %s", line))
				}
			}
			preview = append(preview, "")
		}
	}

	// Children summary
	if node.FirstChild != nil {
		preview = append(preview, dimStyle.Render("Children:"))
		child := node.FirstChild
		for child != nil {
			childName := DecideNodeName(*child)
			preview = append(preview, dimStyle.Render(fmt.Sprintf("  %s (%s)", childName, child.Datatype.Label)))
			child = child.NextSibling
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, preview...)
}

// fieldInfo is a mode-agnostic wrapper for rendering field data.
type fieldInfo struct {
	label string
	value string
}

// currentFields returns field info for the current mode.
func (s *ContentScreen) currentFields() []fieldInfo {
	if s.AdminMode {
		result := make([]fieldInfo, len(s.AdminSelectedFields))
		for i, f := range s.AdminSelectedFields {
			result[i] = fieldInfo{label: f.Label, value: f.Value}
		}
		return result
	}
	result := make([]fieldInfo, len(s.SelectedContentFields))
	for i, f := range s.SelectedContentFields {
		result[i] = fieldInfo{label: f.Label, value: f.Value}
	}
	return result
}

// =============================================================================
// FIELDS RENDERING (right panel)
// =============================================================================

func (s *ContentScreen) renderFields() string {
	if s.AdminMode {
		return s.renderAdminFields()
	}
	return s.renderRegularFields()
}

func (s *ContentScreen) renderRegularFields() string {
	if len(s.SelectedContentFields) == 0 {
		return "No fields"
	}

	fields := []string{}
	for i, cf := range s.SelectedContentFields {
		cursor := "   "
		if s.PanelFocus == RoutePanel && s.FieldCursor == i {
			cursor = " > "
		}

		value := cf.Value
		if value == "" {
			value = "(empty)"
		} else if len(value) > 40 {
			value = value[:37] + "..."
		}

		fields = append(fields, fmt.Sprintf("%s%s: %s", cursor, cf.Label, value))
	}

	return lipgloss.JoinVertical(lipgloss.Left, fields...)
}

func (s *ContentScreen) renderAdminFields() string {
	if len(s.AdminSelectedFields) == 0 {
		return "No fields"
	}

	fields := []string{}
	for i, cf := range s.AdminSelectedFields {
		cursor := "   "
		if s.PanelFocus == RoutePanel && s.FieldCursor == i {
			cursor = " > "
		}

		value := cf.Value
		if value == "" {
			value = "(empty)"
		} else if len(value) > 40 {
			value = value[:37] + "..."
		}

		fields = append(fields, fmt.Sprintf("%s%s: %s", cursor, cf.Label, value))
	}

	return lipgloss.JoinVertical(lipgloss.Left, fields...)
}

// =============================================================================
// TREE ACTIONS (right panel when in tree phase and no version list)
// =============================================================================

func (s *ContentScreen) renderTreeActions() string {
	lines := []string{
		"Actions",
		"",
	}

	switch s.PanelFocus {
	case TreePanel:
		lines = append(lines,
			"Tree Panel",
			"",
			"  n: New content",
			"  e: Edit",
			"  d: Delete",
			"  c: Copy",
			"  m: Move",
			"  +/-: Reorder",
			"  p: Publish",
			"  v: Versions",
			"",
			"  enter: Expand/collapse",
			"  tab: Switch to fields",
			"  esc: Back to list",
		)
	case RoutePanel:
		lines = append(lines,
			"Fields Panel",
			"",
			"  e: Edit field",
			"  n: Add field",
			"  d: Delete field",
			"",
			"  esc: Back to tree",
			"  tab: Switch to tree",
		)
	default:
		lines = append(lines,
			"  tab: Switch panel",
		)
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Nodes: %d", s.visibleNodeCount()))
	lines = append(lines, fmt.Sprintf("Fields: %d", s.fieldsLen()))

	return strings.Join(lines, "\n")
}

// =============================================================================
// VERSION LIST RENDERING
// =============================================================================

func (s *ContentScreen) renderVersionList() string {
	if s.AdminMode {
		return s.renderAdminVersionList()
	}
	return s.renderRegularVersionList()
}

func (s *ContentScreen) renderRegularVersionList() string {
	if len(s.Versions) == 0 {
		return "(no versions)\n\nPress esc to close"
	}

	publishedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#16a34a")).Bold(true)
	triggerStyle := lipgloss.NewStyle().Faint(true)
	labelStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	lines := make([]string, 0, len(s.Versions)*3+2)
	for i, v := range s.Versions {
		cursor := "   "
		if s.VersionCursor == i {
			cursor = " > "
		}

		numStr := fmt.Sprintf("#%d", v.VersionNumber)
		if v.Published {
			numStr = publishedStyle.Render(numStr + " [pub]")
		}

		trigger := triggerStyle.Render(v.Trigger)

		label := ""
		if v.Label != "" {
			label = " " + labelStyle.Render(v.Label)
		}

		date := triggerStyle.Render(v.DateCreated.String())

		lines = append(lines, fmt.Sprintf("%s%s %s%s", cursor, numStr, trigger, label))
		lines = append(lines, fmt.Sprintf("     %s", date))
	}

	lines = append(lines, "")
	lines = append(lines, triggerStyle.Render("  enter:restore | esc:close"))

	return strings.Join(lines, "\n")
}

func (s *ContentScreen) renderAdminVersionList() string {
	if len(s.AdminVersions) == 0 {
		return "(no versions)\n\nPress esc to close"
	}

	publishedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#16a34a")).Bold(true)
	triggerStyle := lipgloss.NewStyle().Faint(true)
	labelStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	lines := make([]string, 0, len(s.AdminVersions)*3+2)
	for i, v := range s.AdminVersions {
		cursor := "   "
		if s.VersionCursor == i {
			cursor = " > "
		}

		numStr := fmt.Sprintf("#%d", v.VersionNumber)
		if v.Published {
			numStr = publishedStyle.Render(numStr + " [pub]")
		}

		trigger := triggerStyle.Render(v.Trigger)

		label := ""
		if v.Label != "" {
			label = " " + labelStyle.Render(v.Label)
		}

		date := triggerStyle.Render(v.DateCreated.String())

		lines = append(lines, fmt.Sprintf("%s%s %s%s", cursor, numStr, trigger, label))
		lines = append(lines, fmt.Sprintf("     %s", date))
	}

	lines = append(lines, "")
	lines = append(lines, triggerStyle.Render("  enter:restore | esc:close"))

	return strings.Join(lines, "\n")
}

// =============================================================================
// ERROR RENDERING
// =============================================================================

func (s *ContentScreen) renderError() string {
	errStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ef4444")).
		Bold(true)

	return errStyle.Render(fmt.Sprintf("Error (%s): %v", s.ErrorContext, s.LastError))
}
