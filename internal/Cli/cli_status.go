package cli

import (
	"fmt"

	utility "github.com/hegner123/modulacms/internal/Utility"
)

func (m model) StatusTable() string {
	headerColor := utility.BLUEB
	r := utility.RESET
	var selected string
	page := fmt.Sprintf("%vPage%v\nPage: %s\nPage Index %d\n", headerColor, r, m.page.Label, m.page.Index)
	i := fmt.Sprintf("%vInterface%v\nCursor: %d\n", headerColor, r, m.cursor)
	menu := fmt.Sprintf("%vMenu%v\nMenu Length: %d\nMenu:\n%v\n", headerColor, r, len(m.menu), getMenuLabels(m.menu))
	if len(m.menu) > 0 {
		selected = fmt.Sprintf("%vSelected%v\nSelected: %v\n", headerColor, r, m.menu[m.cursor].Label)
	} else {
		selected = fmt.Sprintf("%vSelected%v\nSelected: nil\n", headerColor, r)
	}
	controller := fmt.Sprintf("%vController%v\n%v\n", headerColor, r, m.controller)
	tables := fmt.Sprintf("%vTables%v\n%v\n", headerColor, r, m.tables)
	table := fmt.Sprintf("%vTable%v\n%s\n", headerColor, r, m.table)
	var history string
	h, haspage := m.Peek()
	if haspage {
		history = fmt.Sprintf("%vHistory%v\nPrev:\n %v", headerColor, r, h.Label)
	} else {
		history = fmt.Sprintf("%vHistory%v\nPrev: No History", headerColor, r)
	}
	return page + i + menu + selected + controller + tables + table + history + "\n\n"
}

func getMenuLabels(m []*CliPage) string {
	var labels string
	if m != nil {
		for _, item := range m {
			labels = labels + fmt.Sprintf("%v %v\n", item.Index, item.Label)

		}
	} else {
		labels = "Menu is nil"
	}
	return labels

}
