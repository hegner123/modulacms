package cli

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) UpdateLog(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LogModelMsg:
		out := make([]string, 0)
		stringModel := m.Stringify()
		src := strings.Split(stringModel, "\n")
		if msg.Include != nil {
			iRef := *msg.Include
			for _, item := range iRef {
				line := GetLine(src, item)
				out = append(out, line)
			}
		} else if msg.Exclude != nil {
			eRef := *msg.Exclude
			FilterLines(&src, eRef)
			out = src
		} else {
			out = src
		}
		return m, LogMessageCmd(lipgloss.JoinVertical(lipgloss.Top, out...))
	case DbErrMsg:
		return m, LogMessageCmd(msg.Error.Error())
	}

	return m, nil
}
