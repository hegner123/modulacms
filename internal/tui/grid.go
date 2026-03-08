package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Grid defines a 12-column grid layout for screen rendering.
// The sum of all column Span values must equal 12.
type Grid struct {
	Columns []GridColumn
}

// GridColumn is a vertical column in the grid.
type GridColumn struct {
	Span  int        // width units out of 12
	Cells []GridCell // stacked top to bottom
}

// GridCell is a single renderable cell within a column.
type GridCell struct {
	Height float64 // proportional height within column (normalized against siblings)
	Title  string
}

// CellContent holds rendered content and optional scroll state for a grid cell.
type CellContent struct {
	Content      string
	TotalLines   int // 0 disables scrollbar
	ScrollOffset int
}

// CellCount returns the total number of cells across all columns.
func (g Grid) CellCount() int {
	n := 0
	for _, col := range g.Columns {
		n += len(col.Cells)
	}
	return n
}

// CellTitle returns the title of the cell at the given flat index.
func (g Grid) CellTitle(idx int) string {
	n := 0
	for _, col := range g.Columns {
		for _, cell := range col.Cells {
			if n == idx {
				return cell.Title
			}
			n++
		}
	}
	return ""
}

// Render draws the grid layout. Cells are indexed in column-major order
// (top-to-bottom in each column, left-to-right across columns).
// focusIdx highlights the cell at that index. accent overrides the
// default panel accent color when non-zero.
func (g Grid) Render(cells []CellContent, width, height, focusIdx int, accent ...lipgloss.CompleteAdaptiveColor) string {
	if len(g.Columns) == 0 {
		return ""
	}

	var accentColor lipgloss.CompleteAdaptiveColor
	if len(accent) > 0 {
		accentColor = accent[0]
	}

	colWidths := g.columnWidths(width)

	colStrings := make([]string, len(g.Columns))
	cellIdx := 0
	for ci, col := range g.Columns {
		cellHeights := g.cellHeights(ci, height)

		rendered := make([]string, 0, len(col.Cells))
		for ri, cell := range col.Cells {
			var cc CellContent
			if cellIdx < len(cells) {
				cc = cells[cellIdx]
			}

			pan := Panel{
				Title:        cell.Title,
				Width:        colWidths[ci],
				Height:       cellHeights[ri],
				Content:      cc.Content,
				Focused:      cellIdx == focusIdx,
				TotalLines:   cc.TotalLines,
				ScrollOffset: cc.ScrollOffset,
				Accent:       accentColor,
			}
			rendered = append(rendered, pan.Render())
			cellIdx++
		}
		colStrings[ci] = lipgloss.JoinVertical(lipgloss.Left, rendered...)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, colStrings...)
}

// columnWidths converts spans to pixel widths. Remainder goes to the last column.
func (g Grid) columnWidths(width int) []int {
	widths := make([]int, len(g.Columns))
	used := 0
	for i, col := range g.Columns {
		widths[i] = width * col.Span / 12
		used += widths[i]
	}
	if len(widths) > 0 {
		widths[len(widths)-1] += width - used
	}
	return widths
}

// CellInnerHeight returns the usable content height for the cell at the
// given flat index (column-major order), accounting for panel borders and title.
func (g Grid) CellInnerHeight(cellIdx, totalHeight int) int {
	n := 0
	for ci, col := range g.Columns {
		heights := g.cellHeights(ci, totalHeight)
		for ri := range col.Cells {
			if n == cellIdx {
				return PanelInnerHeight(heights[ri])
			}
			n++
		}
	}
	return PanelInnerHeight(totalHeight)
}

// cellHeights converts height ratios to pixel heights within a column.
// Remainder goes to the last cell. Minimum cell height is 3 (border + title).
func (g Grid) cellHeights(colIdx, height int) []int {
	col := g.Columns[colIdx]
	if len(col.Cells) == 0 {
		return nil
	}

	var total float64
	for _, cell := range col.Cells {
		total += cell.Height
	}
	if total == 0 {
		total = float64(len(col.Cells))
	}

	heights := make([]int, len(col.Cells))
	used := 0
	for i, cell := range col.Cells {
		h := int(float64(height) * (cell.Height / total))
		if h < 3 {
			h = 3
		}
		heights[i] = h
		used += h
	}
	heights[len(heights)-1] += height - used

	return heights
}
