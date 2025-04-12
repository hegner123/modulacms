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
	page := fmt.Sprintf("Page: %s  Index %d\n", m.page.Label, m.page.Index)
	cursor := fmt.Sprintf("Cursor: %d\n", m.cursor)
	menu := fmt.Sprintf("Menu: %v\nMenu Len:%d\n", getMenuLabels(m.pageMenu), len(m.pageMenu))
	if len(m.pageMenu) > 0 {
		selected = fmt.Sprintf("Selected: %v\n", m.pageMenu[m.cursor].Label)
	} else {
		selected = "Selected: nil\n"
	}
	controller := fmt.Sprintf("Controller\n%v\n", m.controller)
	//tables := fmt.Sprintf("Tables\n%v\n", m.tables)
	table := fmt.Sprintf("Table\n%s\n", m.table)
	history := fmt.Sprintf("History\nLength:\n %v", len(m.history))
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
				fmt.Sprint("Width:  ", m.width),
				fmt.Sprint("Height: ", m.height),
				table,
				history,
				m.err.Error(),
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
	percent := math.Max(minPercent, math.Min(maxPercent, m.viewport.ScrollPercent()))
	scrollPercent := fmt.Sprintf(" %3.f%% ", percent*percentToStringMagnitude)

	// Note
	var note string
	note = truncate.StringWithTail(" "+note+" ", uint(max(0, m.width-ansi.PrintableRuneWidth(scrollPercent))), ellipsis)
	note = statusBarNoteStyle(note)

	// Empty space
	padding := max(0,
		m.width-
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
