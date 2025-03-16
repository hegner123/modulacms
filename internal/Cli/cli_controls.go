package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

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
	case "enter", "l":
		m.PushHistory(m.page)
		m.table = m.tables[m.cursor]
		m.cursor = 0
		m.page = *m.page.Next
		m.controller = m.page.Controller
		m.menu = m.page.Children
	}
	return m, nil
}



func (m model) FormInputControl(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	return m, cmd
}

// Update handles messages (key presses, etc.).
func (m model) InputControls(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Quit on ctrl+c or esc.
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Cycle focus on tab, shift+tab, enter, up, or down.
		case "tab", "shift+tab", "enter", "up", "down":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			// Wrap the focus index.
			if m.focusIndex >= len(m.textInputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.textInputs) - 1
			}

			// Update focus for each input.
			for i := 0; i < len(m.textInputs); i++ {
				if i == m.focusIndex {
					m.textInputs[i].Focus()
				} else {
					m.textInputs[i].Blur()
				}
			}
			return m, nil
		}
	}

	// Let each text input handle the message.
	var cmd tea.Cmd
	for i := range m.textInputs {
		m.textInputs[i], cmd = m.textInputs[i].Update(msg)
	}

	return m, cmd
}
