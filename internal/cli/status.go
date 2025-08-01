package cli

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/truncate"
)

var ellipsis string = "..."
var statusBarNoteStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#ffffff")).
	Background(lipgloss.Color("#000000")).
	Render

func (m Model) RenderStatusTable() string {
	doc := strings.Builder{}
	var selected string
	page := fmt.Sprintf("Page: %s  Index %d\n", m.Page.Label, m.Page.Index)
	cursor := fmt.Sprintf("Cursor: %d\n", m.Cursor)
	menu := fmt.Sprintf("Menu: %v\nMenu Len:%d\n", getMenuLabels(m.PageMenu), len(m.PageMenu))
	if len(m.PageMenu) > 0 {
		selected = fmt.Sprintf("Selected: %v\n", m.PageMenu[m.Cursor].Label)
	} else {
		selected = "Selected: nil\n"
	}
	controller := fmt.Sprintf("Controller\n%v\n", m.Controller)
	//tables := fmt.Sprintf("Tables\n%v\n", m.Tables)
	table := fmt.Sprintf("Table\n%s\n", m.Table)
	history := fmt.Sprintf("History\nLength:\n %v", len(m.History))
	
	// Add Root info if available
	rootInfo := "Root: Empty"
	if m.Root.Node != nil {
		rootInfo = fmt.Sprintf("Root: Node with %d children", len(m.Root.Node.Nodes))
	}
	
	doc.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Top,
		RenderBorderBlock(
			lipgloss.JoinVertical(
				lipgloss.Left,
				page,
				cursor,
				menu,
				selected,
				controller,
			)),
		RenderBorderBlock(
			lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprint("Width:  ", m.Width),
				fmt.Sprint("Height: ", m.Height),
				table,
				history,
				rootInfo,
				m.Err.Error(),
			)),
	))

	return doc.String()
}

func getMenuLabels(m []*Page) string {
	var labels string
	if m != nil {
		for _, item := range m {
			labels = labels + fmt.Sprintf("%v %v\n", item.Index, item.Label)

		}
	} else {
		labels = "\nMenu is nil\n"
	}
	return labels

}

func (m Model) StatusBarView(b *strings.Builder) {
	const (
		minPercent               float64 = 0.0
		maxPercent               float64 = 1.0
		percentToStringMagnitude float64 = 100.0
	)

	// Scroll percent
	percent := math.Max(minPercent, math.Min(maxPercent, m.Viewport.ScrollPercent()))
	scrollPercent := fmt.Sprintf(" %3.f%% ", percent*percentToStringMagnitude)

	// Note
	var note string
	note = truncate.StringWithTail(" "+note+" ", uint(max(0, m.Width-ansi.PrintableRuneWidth(scrollPercent))), ellipsis)
	note = statusBarNoteStyle(note)

	// Empty space
	padding := max(0,
		m.Width-
			ansi.PrintableRuneWidth(note)-
			ansi.PrintableRuneWidth(scrollPercent),
	)
	emptySpace := strings.Repeat(" ", padding)
	emptySpace = statusBarNoteStyle(emptySpace)

	fmt.Fprintf(b, "%s%s%s",
		note,
		emptySpace,
		scrollPercent,
	)
}
