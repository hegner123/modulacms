package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) RenderStatusTable() string {
	doc := strings.Builder{}
	var selected string
	page := fmt.Sprintf("Page\n%s\nIndex %d\n", m.page.Label, m.page.Index)
	i := fmt.Sprintf("Interface\nCursor: %d\n", m.cursor)
	menu := fmt.Sprintf("Menu\nMenu Length: %d\nMenu: %v\n", len(m.menu), getMenuLabels(m.menu))
	if len(m.menu) > 0 {
		selected = fmt.Sprintf("Selected\nSelected: %v\n", m.menu[m.cursor].Label)
	} else {
		selected = "Selected\nSelected: nil\n"
	}
	controller := fmt.Sprintf("Controller\n%v\n", m.controller)
	tables := fmt.Sprintf("Tables\n%v\n", m.tables)
	table := fmt.Sprintf("Table\n%s\n", m.table)
	var history string
	h, haspage := m.Peek()
	if haspage {
		history = fmt.Sprintf("History\nPrev:\n %v", h.Label)
	} else {
		history = "History\nPrev: No History\n"
	}
	doc.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Top,
		RenderBlock(
			lipgloss.JoinVertical(
				lipgloss.Left,
				page,
				i,
				menu,
				selected,
			)),
		RenderBlock(
			lipgloss.JoinVertical(
				lipgloss.Left,
				controller,
			)),
		RenderBlock(
			lipgloss.JoinVertical(
				lipgloss.Left,
				tables,
				table,
				history,
			)),
	))

	return doc.String()
}

func getMenuLabels(m []*CliPage) string {
	var labels string
	if m != nil {
		for _, item := range m {
			labels = labels + fmt.Sprintf("\n%v %v\n", item.Index, item.Label)

		}
	} else {
		labels = "\nMenu is nil\n"
	}
	return labels

}
