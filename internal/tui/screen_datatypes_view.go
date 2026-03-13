package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
)

// ---------------------------------------------------------------------------
// Key Hints
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) KeyHints(km config.KeyMap) []KeyHint {
	if s.Phase == DatatypesPhaseFields {
		return []KeyHint{
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
			{km.HintString(config.ActionReorderUp) + "/" + km.HintString(config.ActionReorderDown), "reorder"},
			{km.HintString(config.ActionSelect), "edit"},
			{km.HintString(config.ActionNew), "new"},
			{km.HintString(config.ActionEdit), "edit all"},
			{km.HintString(config.ActionDelete), "del"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "back"},
		}
	}
	return []KeyHint{
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionSelect), "select"},
		{"/", "search"},
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionEdit), "edit"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionReorderUp) + "/" + km.HintString(config.ActionReorderDown), "reorder"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) View(ctx AppContext) string {
	if s.Phase == DatatypesPhaseFields {
		return s.viewFields(ctx)
	}
	return s.viewBrowse(ctx)
}

func (s *DatatypesScreen) viewBrowse(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderDatatypeTree(), TotalLines: len(s.FlatDTList), ScrollOffset: ClampScroll(s.Cursor, len(s.FlatDTList), CellInnerHeight(ctx.Height, s.Grid, 0))},
		{Content: s.renderDatatypeDetails()},
		{Content: s.renderFieldPreview()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *DatatypesScreen) viewFields(ctx AppContext) string {
	fLen := s.fieldsLen()
	cells := []CellContent{
		{Content: s.renderFieldList(), TotalLines: fLen, ScrollOffset: ClampScroll(s.Cursor, fLen, CellInnerHeight(ctx.Height, s.Grid, 0))},
		{Content: s.renderFieldProperties()},
		{Content: s.renderContext()},
	}
	return s.RenderGrid(ctx, cells)
}

// ---------------------------------------------------------------------------
// Phase 1 renderers
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) renderDatatypeTree() string {
	var lines []string

	// Search bar
	if s.Searching {
		lines = append(lines, " "+s.SearchInput.View())
		lines = append(lines, "")
	} else if s.SearchQuery != "" {
		faint := lipgloss.NewStyle().Faint(true)
		lines = append(lines, faint.Render(fmt.Sprintf(" filter: %s", s.SearchQuery)))
		lines = append(lines, "")
	}

	if len(s.FlatDTList) == 0 {
		lines = append(lines, " (no datatypes)")
		return strings.Join(lines, "\n")
	}

	for i, node := range s.FlatDTList {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		indent := strings.Repeat("  ", node.Depth)

		expandIcon := " "
		if node.Kind == DatatypeNodeGroup {
			if node.Expand {
				expandIcon = "v"
			} else {
				expandIcon = ">"
			}
		}

		typeBadge := ""
		if node.Datatype != nil {
			typeBadge = fmt.Sprintf(" [%s]", node.Datatype.Type)
		} else if node.AdminDT != nil {
			typeBadge = fmt.Sprintf(" [%s]", node.AdminDT.Type)
		}

		lines = append(lines, fmt.Sprintf(" %s %s%s %s%s", cursor, indent, expandIcon, node.Label, typeBadge))
	}

	return strings.Join(lines, "\n")
}

func (s *DatatypesScreen) renderDatatypeDetails() string {
	if len(s.FlatDTList) == 0 || s.Cursor >= len(s.FlatDTList) {
		return " No datatype selected"
	}

	node := s.FlatDTList[s.Cursor]
	faint := lipgloss.NewStyle().Faint(true)

	var lines []string

	if node.Datatype != nil {
		dt := node.Datatype
		lines = append(lines,
			fmt.Sprintf(" Label   %s", dt.Label),
			fmt.Sprintf(" Name    %s", dt.Name),
			fmt.Sprintf(" Type    %s", dt.Type),
		)
		if dt.ParentID.Valid {
			lines = append(lines, fmt.Sprintf(" Parent  %s", s.parentLabel(string(dt.ParentID.ID))))
		}
		lines = append(lines,
			"",
			faint.Render(fmt.Sprintf(" ID      %s", dt.DatatypeID)),
			faint.Render(fmt.Sprintf(" Author  %s", dt.AuthorID)),
			faint.Render(fmt.Sprintf(" Fields  %d", len(s.Fields))),
		)
	} else if node.AdminDT != nil {
		dt := node.AdminDT
		lines = append(lines,
			fmt.Sprintf(" Label   %s", dt.Label),
			fmt.Sprintf(" Name    %s", dt.Name),
			fmt.Sprintf(" Type    %s", dt.Type),
		)
		if dt.ParentID.Valid {
			lines = append(lines, fmt.Sprintf(" Parent  %s", s.adminParentLabel(string(dt.ParentID.ID))))
		}
		lines = append(lines,
			"",
			faint.Render(fmt.Sprintf(" ID      %s", dt.AdminDatatypeID)),
			faint.Render(fmt.Sprintf(" Author  %s", dt.AuthorID)),
			faint.Render(fmt.Sprintf(" Fields  %d", len(s.AdminFields))),
		)
	}

	return strings.Join(lines, "\n")
}

func (s *DatatypesScreen) renderFieldPreview() string {
	if len(s.FlatDTList) == 0 || s.Cursor >= len(s.FlatDTList) {
		return " No datatype selected"
	}

	if s.AdminMode {
		return s.renderAdminFieldPreview()
	}
	return s.renderRegularFieldPreview()
}

func (s *DatatypesScreen) renderRegularFieldPreview() string {
	if len(s.Fields) == 0 {
		return " (no fields)"
	}

	lines := make([]string, 0, len(s.Fields))
	for i, field := range s.Fields {
		lines = append(lines, fmt.Sprintf(" %d. %s [%s]", i+1, field.Label, field.Type))
	}
	return strings.Join(lines, "\n")
}

func (s *DatatypesScreen) renderAdminFieldPreview() string {
	if len(s.AdminFields) == 0 {
		return " (no fields)"
	}

	lines := make([]string, 0, len(s.AdminFields))
	for i, field := range s.AdminFields {
		lines = append(lines, fmt.Sprintf(" %d. %s [%s]", i+1, field.Label, field.Type))
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Phase 2 renderers
// ---------------------------------------------------------------------------

func (s *DatatypesScreen) renderFieldList() string {
	if s.AdminMode {
		return s.renderAdminFieldList()
	}
	return s.renderRegularFieldList()
}

func (s *DatatypesScreen) renderRegularFieldList() string {
	if len(s.Fields) == 0 {
		lines := []string{" (no fields)", "", " Press 'n' to add a field"}
		return strings.Join(lines, "\n")
	}

	lines := make([]string, 0, len(s.Fields))
	for i, field := range s.Fields {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %d. %s [%s]", cursor, i+1, field.Label, field.Type))
	}
	return strings.Join(lines, "\n")
}

func (s *DatatypesScreen) renderAdminFieldList() string {
	if len(s.AdminFields) == 0 {
		lines := []string{" (no fields)", "", " Press 'n' to add a field"}
		return strings.Join(lines, "\n")
	}

	lines := make([]string, 0, len(s.AdminFields))
	for i, field := range s.AdminFields {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %d. %s [%s]", cursor, i+1, field.Label, field.Type))
	}
	return strings.Join(lines, "\n")
}

func (s *DatatypesScreen) renderFieldProperties() string {
	if len(s.Properties) == 0 {
		return " Select a field to view properties"
	}

	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	faint := lipgloss.NewStyle().Faint(true)

	lines := make([]string, 0, len(s.Properties))
	for _, prop := range s.Properties {
		label := fmt.Sprintf(" %-14s", prop.Key)
		if prop.Editable {
			lines = append(lines, fmt.Sprintf("%s %s", accent.Render(label), prop.Value))
		} else {
			lines = append(lines, fmt.Sprintf("%s %s", faint.Render(label), faint.Render(prop.Value)))
		}
	}
	return strings.Join(lines, "\n")
}

func (s *DatatypesScreen) renderContext() string {
	var lines []string

	// Breadcrumb
	if s.SelectedDTNode != nil {
		breadcrumb := s.buildBreadcrumb()
		accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
		lines = append(lines, " "+accent.Render(breadcrumb))
		lines = append(lines, "")
	}

	lines = append(lines, fmt.Sprintf(" Fields: %d", s.fieldsLen()))

	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parentLabel finds the label of a regular datatype by ID.
func (s *DatatypesScreen) parentLabel(id string) string {
	for _, dt := range s.Datatypes {
		if string(dt.DatatypeID) == id {
			return dt.Label
		}
	}
	return id
}

// adminParentLabel finds the label of an admin datatype by ID.
func (s *DatatypesScreen) adminParentLabel(id string) string {
	for _, dt := range s.AdminDatatypes {
		if string(dt.AdminDatatypeID) == id {
			return dt.Label
		}
	}
	return id
}

// buildBreadcrumb builds a parent > child breadcrumb for the selected datatype.
func (s *DatatypesScreen) buildBreadcrumb() string {
	if s.SelectedDTNode == nil {
		return ""
	}

	if s.AdminMode && s.SelectedDTNode.AdminDT != nil {
		dt := s.SelectedDTNode.AdminDT
		if dt.ParentID.Valid {
			parentLabel := s.adminParentLabel(string(dt.ParentID.ID))
			return parentLabel + " > " + dt.Label
		}
		return dt.Label
	}

	if !s.AdminMode && s.SelectedDTNode.Datatype != nil {
		dt := s.SelectedDTNode.Datatype
		if dt.ParentID.Valid {
			parentLabel := s.parentLabel(string(dt.ParentID.ID))
			return parentLabel + " > " + dt.Label
		}
		return dt.Label
	}

	return s.SelectedDTNode.Label
}

// CellInnerHeight estimates the inner height of a grid cell for scroll calculations.
func CellInnerHeight(totalHeight int, grid Grid, cellIndex int) int {
	idx := 0
	for _, col := range grid.Columns {
		for _, cell := range col.Cells {
			if idx == cellIndex {
				h := int(float64(totalHeight) * cell.Height)
				if h < 3 {
					return 1
				}
				return h - 2 // borders
			}
			idx++
		}
	}
	return totalHeight - 2
}
