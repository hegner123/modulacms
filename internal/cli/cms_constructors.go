package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

func BuildTreeFromRouteCMD(id int64) tea.Cmd {
	return func() tea.Msg {
		return BuildTreeFromRouteMsg{
			RouteID: id,
		}
	}
}

