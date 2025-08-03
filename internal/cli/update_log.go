package cli

import tea "github.com/charmbracelet/bubbletea"

func (m Model) UpdateLog(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DbErrMsg:
		return m, LogMessageCmd(msg.Error.Error())
	}

	return m, nil
}
