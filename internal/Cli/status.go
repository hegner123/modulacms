package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) RenderStatusTable() string {
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
	var history string
	h, haspage := m.Peek()
	if haspage {
		history = fmt.Sprintf("History\nPrev:\n %v", h.Label)
	} else {
		history = "History\nPrev: No History\n"
	}
	var formMapStatus []string
	for i, v := range m.formValues {
		s := fmt.Sprintf("%s: %v\n", m.headers[i], *v)
		formMapStatus = append(formMapStatus, s)

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
				table,
				history,
			)),
		RenderBorderBlock(
			lipgloss.JoinVertical(
				lipgloss.Left,
				formMapStatus...,
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
