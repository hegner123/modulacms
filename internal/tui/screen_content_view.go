package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// View renders the ContentScreen using the grid layout system.
func (s *ContentScreen) View(ctx AppContext) string {
	if s.inTreePhase() {
		return s.viewTreePhase(ctx)
	}
	return s.viewSelectPhase(ctx)
}

// viewSelectPhase renders the content select grid (2 columns: content list + details/stats).
func (s *ContentScreen) viewSelectPhase(ctx AppContext) string {
	innerH := PanelInnerHeight(ctx.Height)
	listTotal := len(s.FlatSelectList)

	cells := []CellContent{
		{
			Content:      s.renderSelectList(),
			TotalLines:   listTotal,
			ScrollOffset: ClampScroll(s.Cursor, listTotal, innerH),
		},
		{Content: s.renderSelectDetail()},
		{Content: s.renderSelectStats()},
	}

	if s.LastError != nil {
		cells[1].Content = s.renderError() + "\n\n" + cells[1].Content
	}

	return s.RenderGrid(ctx, cells)
}

// viewTreePhase renders the tree browsing grid (2 columns: tree + preview).
func (s *ContentScreen) viewTreePhase(ctx AppContext) string {
	innerH := PanelInnerHeight(ctx.Height)
	treeTotal := s.visibleNodeCount()

	treeCell := CellContent{
		Content:      s.renderTree(),
		TotalLines:   treeTotal,
		ScrollOffset: ClampScroll(s.Cursor, treeTotal, innerH),
	}
	preview := s.renderDocumentPreview(ctx)
	previewCell := CellContent{
		Content:      preview.content,
		TotalLines:   preview.totalLines,
		ScrollOffset: ClampScroll(preview.selectedLine, preview.totalLines, innerH),
	}

	if !s.ShowVersionList {
		cells := []CellContent{treeCell, previewCell}
		if s.LastError != nil {
			cells[1].Content = s.renderError() + "\n\n" + cells[1].Content
		}
		return s.RenderGrid(ctx, cells)
	}

	// Version list: split height between grid and footer
	gridHeight := int(float64(ctx.Height) * 0.65)
	footerHeight := ctx.Height - gridHeight

	gridCtx := ctx
	gridCtx.Height = gridHeight
	cells := []CellContent{treeCell, previewCell}
	gridStr := s.RenderGrid(gridCtx, cells)

	versionPanel := Panel{
		Title:        "Versions",
		Width:        ctx.Width,
		Height:       footerHeight,
		Content:      s.renderVersionList(),
		Focused:      true,
		TotalLines:   s.versionListLen(),
		ScrollOffset: ClampScroll(s.VersionCursor, s.versionListLen(), footerHeight-2),
		Accent:       ctx.ActiveAccent,
	}

	return lipgloss.JoinVertical(lipgloss.Left, gridStr, versionPanel.Render())
}

// =============================================================================
// SELECT PHASE RENDERING (slug tree)
// =============================================================================

func (s *ContentScreen) renderSelectList() string {
	if len(s.FlatSelectList) == 0 && len(s.SelectTree) == 0 {
		return "(no content)"
	}

	var lines []string
	flatIdx := 0

	for _, node := range s.SelectTree {
		if node.Kind == NodeSection {
			sectionStyle := lipgloss.NewStyle().Bold(true).Faint(true)
			lines = append(lines, sectionStyle.Render("-- "+node.Label+" --"))
			continue
		}
		s.renderSelectNode(node, &lines, &flatIdx, 0)
	}

	if len(lines) == 0 {
		return "(no content)"
	}
	return strings.Join(lines, "\n")
}

func (s *ContentScreen) renderSelectNode(node *ContentSelectNode, lines *[]string, flatIdx *int, depth int) {
	indent := strings.Repeat("  ", depth)
	cursorMark := "  "
	if s.Cursor == *flatIdx {
		cursorMark = "->"
	}

	switch node.Kind {
	case NodeGroup:
		expandIcon := "▶"
		if node.Expand {
			expandIcon = "▼"
		}
		// A group that is also content (e.g., /about with children)
		if node.Content != nil || node.AdminContent != nil {
			status := s.selectNodeStatus(node)
			dtLabel := s.selectNodeDatatypeLabel(node)
			*lines = append(*lines, fmt.Sprintf("%s%s%s %s %s [%s]", cursorMark, indent, expandIcon, node.Label, status, dtLabel))
		} else {
			*lines = append(*lines, fmt.Sprintf("%s%s%s %s", cursorMark, indent, expandIcon, node.Label))
		}
		*flatIdx++
		if node.Expand {
			child := node.FirstChild
			for child != nil {
				s.renderSelectNode(child, lines, flatIdx, depth+1)
				child = child.NextSibling
			}
		}

	case NodeContent:
		status := s.selectNodeStatus(node)
		dtLabel := s.selectNodeDatatypeLabel(node)
		*lines = append(*lines, fmt.Sprintf("%s%s  %s %s [%s]", cursorMark, indent, node.Label, status, dtLabel))
		*flatIdx++
	}
}

func (s *ContentScreen) selectNodeStatus(node *ContentSelectNode) string {
	publishedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#16a34a"))
	draftStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ca8a04"))

	var status types.ContentStatus
	if node.Content != nil {
		status = node.Content.Status
	} else if node.AdminContent != nil {
		status = node.AdminContent.Status
	}

	if status == types.ContentStatusPublished {
		return publishedStyle.Render("●")
	}
	return draftStyle.Render("○")
}

func (s *ContentScreen) selectNodeDatatypeLabel(node *ContentSelectNode) string {
	if node.Content != nil {
		return node.Content.DatatypeLabel
	}
	if node.AdminContent != nil {
		return node.AdminContent.DatatypeLabel
	}
	return ""
}

func (s *ContentScreen) renderSelectDetail() string {
	if s.Cursor >= len(s.FlatSelectList) || len(s.FlatSelectList) == 0 {
		return "No content selected"
	}

	node := s.FlatSelectList[s.Cursor]

	if node.Content != nil {
		c := node.Content
		lines := []string{
			fmt.Sprintf("Route:    %s", c.RouteSlug),
			fmt.Sprintf("Title:    %s", c.RouteTitle),
			fmt.Sprintf("Datatype: %s", c.DatatypeLabel),
			fmt.Sprintf("Status:   %s", c.Status),
			"",
			fmt.Sprintf("Author:   %s", c.AuthorName),
			fmt.Sprintf("ID:       %s", c.ContentDataID),
			"",
			fmt.Sprintf("Created:  %s", c.DateCreated.String()),
			fmt.Sprintf("Modified: %s", c.DateModified.String()),
		}
		return strings.Join(lines, "\n")
	}

	if node.AdminContent != nil {
		c := node.AdminContent
		lines := []string{
			fmt.Sprintf("Route:    %s", c.RouteSlug),
			fmt.Sprintf("Title:    %s", c.RouteTitle),
			fmt.Sprintf("Datatype: %s", c.DatatypeLabel),
			fmt.Sprintf("Status:   %s", c.Status),
			"",
			fmt.Sprintf("Author:   %s", c.AuthorName),
			fmt.Sprintf("ID:       %s", c.AdminContentDataID),
			"",
			fmt.Sprintf("Created:  %s", c.DateCreated.String()),
			fmt.Sprintf("Modified: %s", c.DateModified.String()),
		}
		return strings.Join(lines, "\n")
	}

	// Group node without content
	return fmt.Sprintf("Group: %s", node.Label)
}

func (s *ContentScreen) renderSelectStats() string {
	routedCount := 0
	standaloneCount := 0
	if s.AdminMode {
		for _, item := range s.AdminRootContentSummary {
			if item.AdminRouteID.Valid {
				routedCount++
			} else {
				standaloneCount++
			}
		}
	} else {
		for _, item := range s.RootContentSummary {
			if item.RouteID.Valid {
				routedCount++
			} else {
				standaloneCount++
			}
		}
	}

	lines := []string{
		fmt.Sprintf("Total:      %d", routedCount+standaloneCount),
		fmt.Sprintf("Pages:      %d", routedCount),
		fmt.Sprintf("Standalone: %d", standaloneCount),
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
// DOCUMENT PREVIEW RENDERING (tree phase)
// =============================================================================

// previewResult holds the rendered preview content and scroll metadata.
type previewResult struct {
	content      string
	totalLines   int
	selectedLine int // line index where the selected node's section starts
}

// renderDocumentPreview renders all tree nodes depth-first with their field values.
// The selected node's section is highlighted; others are dimmed.
func (s *ContentScreen) renderDocumentPreview(ctx AppContext) previewResult {
	if s.Root.Root == nil {
		return previewResult{content: "No content loaded"}
	}

	selectedNode := s.Root.NodeAtIndex(s.Cursor)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent2)
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent)
	dimStyle := lipgloss.NewStyle().Faint(true)
	labelStyle := lipgloss.NewStyle().Bold(true)

	var preview []string
	currentIndex := 0
	selectedLine := 0
	s.renderPreviewNode(s.Root.Root, selectedNode, &preview, &currentIndex, &selectedLine, 0, titleStyle, selectedStyle, dimStyle, labelStyle)

	if len(preview) == 0 {
		return previewResult{content: "Empty content tree"}
	}
	return previewResult{
		content:      strings.Join(preview, "\n"),
		totalLines:   len(preview),
		selectedLine: selectedLine,
	}
}

func (s *ContentScreen) renderPreviewNode(
	node *tree.Node,
	selectedNode *tree.Node,
	preview *[]string,
	currentIndex *int,
	selectedLine *int,
	depth int,
	titleStyle, selectedStyle, dimStyle, labelStyle lipgloss.Style,
) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	name := DecideNodeName(*node)
	isSelected := selectedNode != nil && node.Instance != nil && selectedNode.Instance != nil &&
		node.Instance.ContentDataID == selectedNode.Instance.ContentDataID

	// Record line position of the selected node
	if isSelected {
		*selectedLine = len(*preview)
	}

	// Section title
	if isSelected {
		*preview = append(*preview, selectedStyle.Render(indent+name+" ════════"))
	} else {
		*preview = append(*preview, dimStyle.Render(indent+name))
	}

	// Field values from batch-loaded data
	fields := s.fieldsForNode(node)
	fieldIndent := indent + "  "
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		if isSelected {
			*preview = append(*preview, labelStyle.Render(fieldIndent+f.label)+": "+f.value)
		} else {
			*preview = append(*preview, dimStyle.Render(fieldIndent+f.label+": "+f.value))
		}
	}

	if len(fields) > 0 || isSelected {
		*preview = append(*preview, "")
	}

	*currentIndex++

	// Recurse into children (always show in preview, regardless of expand state)
	if node.FirstChild != nil {
		s.renderPreviewNode(node.FirstChild, selectedNode, preview, currentIndex, selectedLine, depth+1, titleStyle, selectedStyle, dimStyle, labelStyle)
	}
	if node.NextSibling != nil {
		s.renderPreviewNode(node.NextSibling, selectedNode, preview, currentIndex, selectedLine, depth, titleStyle, selectedStyle, dimStyle, labelStyle)
	}
}

// fieldInfo is a mode-agnostic wrapper for rendering field data.
type fieldInfo struct {
	label string
	value string
}

// fieldsForNode returns field info for a tree node from batch-loaded data.
func (s *ContentScreen) fieldsForNode(node *tree.Node) []fieldInfo {
	if node == nil || node.Instance == nil {
		return nil
	}

	if s.AdminMode {
		adminID := types.AdminContentID(node.Instance.ContentDataID)
		if fields, ok := s.AllAdminFields[adminID]; ok {
			result := make([]fieldInfo, len(fields))
			for i, f := range fields {
				result[i] = fieldInfo{label: f.Label, value: f.Value}
			}
			return result
		}
		return nil
	}

	if fields, ok := s.AllFields[node.Instance.ContentDataID]; ok {
		result := make([]fieldInfo, len(fields))
		for i, f := range fields {
			result[i] = fieldInfo{label: f.Label, value: f.Value}
		}
		return result
	}
	return nil
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
