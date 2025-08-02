package cli

import "strings"

func ViewPageMenus(m Model) string {
	out := strings.Builder{}
	for _, item := range m.PageMenu {
		out.WriteString(" " + item.Label + " ")

	}
    return out.String()
}
