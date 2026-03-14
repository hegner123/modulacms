package plugin

import (
	"fmt"
	"strings"
)

// RenderPrimitive renders a PluginPrimitive to a string within the given
// dimensions. The output is suitable for embedding in a CellContent.Content
// or for inline display in a field bubble.
func RenderPrimitive(prim PluginPrimitive, width, height int, focused bool) string {
	switch p := prim.(type) {
	case *ListPrimitive:
		return renderList(p, width, focused)
	case *DetailPrimitive:
		return renderDetail(p, width)
	case *TextPrimitive:
		return renderText(p, width)
	case *TablePrimitive:
		return renderTable(p, width, focused)
	case *InputPrimitive:
		return renderInput(p, width)
	case *SelectPrimitive:
		return renderSelect(p, width)
	case *TreePrimitive:
		return renderTree(p, width, focused)
	case *ProgressPrimitive:
		return renderProgress(p, width)
	default:
		return fmt.Sprintf(" (unknown primitive: %T)", prim)
	}
}

func renderList(p *ListPrimitive, width int, focused bool) string {
	if len(p.Items) == 0 {
		text := p.EmptyText
		if text == "" {
			text = "(empty)"
		}
		return " " + text
	}

	lines := make([]string, 0, len(p.Items))
	for i, item := range p.Items {
		cursor := "   "
		if i == p.Cursor {
			cursor = " ->"
		}

		label := item.Label
		if item.Faint {
			label = dimText(label)
		}
		if item.Bold {
			label = boldText(label)
		}

		line := fmt.Sprintf("%s %s", cursor, label)
		if width > 0 && len(line) > width {
			line = line[:width]
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func renderDetail(p *DetailPrimitive, width int) string {
	if len(p.Fields) == 0 {
		return " (no details)"
	}

	// Find max label width for alignment.
	maxLabel := 0
	for _, f := range p.Fields {
		if len(f.Label) > maxLabel {
			maxLabel = len(f.Label)
		}
	}

	lines := make([]string, 0, len(p.Fields))
	for _, f := range p.Fields {
		value := f.Value
		if f.Faint {
			value = dimText(value)
		}
		line := fmt.Sprintf(" %-*s  %s", maxLabel, f.Label, value)
		if width > 0 && len(line) > width {
			line = line[:width]
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func renderText(p *TextPrimitive, width int) string {
	if len(p.Lines) == 0 {
		return ""
	}

	lines := make([]string, 0, len(p.Lines))
	for _, tl := range p.Lines {
		text := tl.Text
		if tl.Bold {
			text = boldText(text)
		}
		if tl.Faint {
			text = dimText(text)
		}
		// Accent and style colors are handled by the TUI host via lipgloss
		// since we produce plain strings here. The accent flag is preserved
		// in the primitive for the host to use.
		line := " " + text
		if width > 0 && len(line) > width {
			line = line[:width]
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func renderTable(p *TablePrimitive, width int, focused bool) string {
	if len(p.Headers) == 0 && len(p.Rows) == 0 {
		return " (empty table)"
	}

	// Calculate column widths based on headers and data.
	colCount := len(p.Headers)
	for _, row := range p.Rows {
		if len(row) > colCount {
			colCount = len(row)
		}
	}
	if colCount == 0 {
		return " (empty table)"
	}

	colWidths := make([]int, colCount)
	for i, h := range p.Headers {
		if len(h) > colWidths[i] {
			colWidths[i] = len(h)
		}
	}
	for _, row := range p.Rows {
		for i, cell := range row {
			if i < colCount && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Ensure minimum column width of 3.
	for i := range colWidths {
		if colWidths[i] < 3 {
			colWidths[i] = 3
		}
	}

	var lines []string

	// Header row.
	if len(p.Headers) > 0 {
		headerParts := make([]string, colCount)
		for i := range colCount {
			h := ""
			if i < len(p.Headers) {
				h = p.Headers[i]
			}
			headerParts[i] = fmt.Sprintf("%-*s", colWidths[i], h)
		}
		lines = append(lines, "  "+strings.Join(headerParts, "  "))
		// Separator.
		sepParts := make([]string, colCount)
		for i := range colCount {
			sepParts[i] = strings.Repeat("-", colWidths[i])
		}
		lines = append(lines, "  "+strings.Join(sepParts, "  "))
	}

	// Data rows.
	for ri, row := range p.Rows {
		cursor := "  "
		if ri == p.Cursor {
			cursor = "->"
		}
		cellParts := make([]string, colCount)
		for i := range colCount {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			cellParts[i] = fmt.Sprintf("%-*s", colWidths[i], cell)
		}
		line := cursor + strings.Join(cellParts, "  ")
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func renderInput(p *InputPrimitive, width int) string {
	display := p.Value
	if display == "" && p.Placeholder != "" {
		display = dimText(p.Placeholder)
	}

	prefix := "  "
	if p.Focused {
		prefix = "> "
	}

	// Simple representation — the TUI host may wrap this with a textinput bubble.
	line := prefix + "[" + display + "]"
	if width > 0 && len(line) > width {
		line = line[:width]
	}
	return line
}

func renderSelect(p *SelectPrimitive, width int) string {
	if len(p.Options) == 0 {
		return " (no options)"
	}

	selectedLabel := ""
	if p.Selected >= 0 && p.Selected < len(p.Options) {
		selectedLabel = p.Options[p.Selected].Label
	}

	prefix := "  "
	if p.Focused {
		prefix = "> "
	}
	return prefix + "< " + selectedLabel + " >"
}

func renderTree(p *TreePrimitive, width int, focused bool) string {
	if len(p.Nodes) == 0 {
		return " (empty tree)"
	}

	var lines []string
	flatIdx := 0
	renderTreeNodes(p.Nodes, 0, p.Cursor, &lines, &flatIdx)
	return strings.Join(lines, "\n")
}

func renderTreeNodes(nodes []TreeNode, depth, cursor int, lines *[]string, flatIdx *int) {
	for _, node := range nodes {
		indent := strings.Repeat("  ", depth)
		prefix := "   "
		if *flatIdx == cursor {
			prefix = " ->"
		}

		expandIcon := " "
		if len(node.Children) > 0 {
			if node.Expanded {
				expandIcon = "▼"
			} else {
				expandIcon = "▶"
			}
		}

		*lines = append(*lines, fmt.Sprintf("%s%s%s %s", prefix, indent, expandIcon, node.Label))
		*flatIdx++

		if node.Expanded && len(node.Children) > 0 {
			renderTreeNodes(node.Children, depth+1, cursor, lines, flatIdx)
		}
	}
}

func renderProgress(p *ProgressPrimitive, width int) string {
	barWidth := width - 4 // padding + borders
	if barWidth < 5 {
		barWidth = 5
	}

	filled := int(float64(barWidth) * p.Value)
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	line := " " + bar
	if p.Label != "" {
		line += " " + p.Label
	}

	pct := fmt.Sprintf(" %d%%", int(p.Value*100))
	line += pct

	return line
}

// LayoutToCells converts a PluginLayout into the data needed for grid rendering.
// Returns cell contents (rendered strings) and the Grid definition. The caller
// uses these with GridScreen.RenderGrid().
func LayoutToCells(layout *PluginLayout, width, height, focusIndex int) ([]CellRenderData, GridDef) {
	var cells []CellRenderData

	gridDef := GridDef{
		Columns: make([]GridDefColumn, 0, len(layout.Columns)),
	}

	for _, col := range layout.Columns {
		defCol := GridDefColumn{
			Span:  col.Span,
			Cells: make([]GridDefCell, 0, len(col.Cells)),
		}
		for _, cell := range col.Cells {
			content := ""
			totalLines := 0
			if cell.Content != nil {
				content = RenderPrimitive(cell.Content, width, height, false)
				totalLines = strings.Count(content, "\n") + 1
			}
			cells = append(cells, CellRenderData{
				Content:    content,
				TotalLines: totalLines,
			})
			defCol.Cells = append(defCol.Cells, GridDefCell{
				Title:  cell.Title,
				Height: cell.Height,
			})
		}
		gridDef.Columns = append(gridDef.Columns, defCol)
	}

	return cells, gridDef
}

// CellRenderData holds rendered content for a single grid cell.
type CellRenderData struct {
	Content    string
	TotalLines int
}

// GridDef is a serializable grid definition for passing layout info
// from the plugin package to the tui package without creating an import cycle.
type GridDef struct {
	Columns []GridDefColumn
}

// GridDefColumn is a column in the grid definition.
type GridDefColumn struct {
	Span  int
	Cells []GridDefCell
}

// GridDefCell is a cell in the grid definition.
type GridDefCell struct {
	Title  string
	Height float64
}

// Simple text styling helpers. These produce plain-text markers that the TUI
// host can interpret, or we can use ANSI codes directly since lipgloss will
// handle final rendering.

func dimText(s string) string {
	// Use ANSI dim (2) directly — lipgloss preserves ANSI in content strings.
	return "\033[2m" + s + "\033[22m"
}

func boldText(s string) string {
	return "\033[1m" + s + "\033[22m"
}
