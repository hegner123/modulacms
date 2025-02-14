package cli

import tea "github.com/charmbracelet/bubbletea"

/*
Key Pressed
backspace
Key Pressed
shift+tab
Key Pressed
delete
*/

func (m model) PageControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return m, tea.Quit

	//Navigation
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < option-1 {
			m.cursor++
		}
	case "h", "shift+tab", "backspace":
		m.cursor = 0
		m.page = *m.PopHistory()
		m.controller = m.page.Controller
		m.menu = m.page.Children

	//Action
	case "enter", "l":
		m.PushHistory(m.page)
		m.page = *m.menu[m.cursor]
		m.cursor = 0
		m.controller = m.page.Controller
		m.menu = m.page.Children
	}
	return m, nil
}

func (m model) TableControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return m, tea.Quit

	//Navigation
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tables)-1 {
			m.cursor++
		}
	case "h", "shift+tab", "backspace":
		m.cursor = 0
		m.page = *m.PopHistory()
		m.controller = m.page.Controller
		m.menu = m.page.Children

	//Action
	case "enter","l":
		m.PushHistory(m.page)
		m.table = m.tables[m.cursor]
		m.cursor = 0
		m.page = *m.page.Next
		m.controller = m.page.Controller
		m.menu = m.page.Children
	}
	return m, nil
}
