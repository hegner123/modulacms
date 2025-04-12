package cli

import tea "github.com/charmbracelet/bubbletea"

type CMSInput interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}
