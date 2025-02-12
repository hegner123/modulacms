package cli

import tea "github.com/charmbracelet/bubbletea"

var (
	createInterface CliInterface = "CreateInterface"
	readInterface   CliInterface = "ReadInterface"
	updateInterface CliInterface = "UpdateInterface"
	deleteInterface CliInterface = "DeleteInterface"
	tableInterface  CliInterface = "TableInterface"
	pageInterface   CliInterface = "PageInterface"
)

func (m model) ResetInterface() model {
	m.cursor = 0
	m.menu = []*CliPage{}
	return m
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {

	switch m.controller {
	case createInterface:
	case readInterface:
	case updateInterface:
	case deleteInterface:
	case pageInterface:
		return m.UpdatePageSelect(message)
	case tableInterface:
		return m.UpdateTableSelect(message)
	}
	return m, nil
}
func (m model) UpdateTableSelect(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.tables)-1 {
				m.cursor++
			}
		case "enter":
			m.table = m.tables[m.cursor]
			m.cursor = 0
			m.controller = pageInterface
		}
	}
	return m, nil
}

func (m model) UpdatePageSelect(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menu)-1 {
				m.cursor++
			}
		case "enter":
            m.PushHistory(m.page)
			m.page = *m.menu[m.cursor]
			m.menu = m.page.Children
			m.controller = m.page.Controller
			m.cursor = 0
		}
	}
	return m, nil
}
func (m model) UpdateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) UpdateCreateInterface(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menu)-1 {
				m.cursor++
			}
		case "enter":
			m.cursor = 0
		}
	}
	return m, nil

}
