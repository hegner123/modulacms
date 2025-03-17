package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	db "github.com/hegner123/modulacms/internal/Db"
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
		m.PageRouter()
	}
	return m, nil
}

func (m model) TableSelectControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
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
func (m model) DatabaseCreateControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
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
		if m.cursor < m.formLen-1 {
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
		err := m.CLICreate(db.StringDBTable(m.table))
		if err != nil {
			return m, nil
		}
		m.cursor = 0
		m.page = *m.page.Next
		m.controller = m.page.Controller
		m.menu = m.page.Children
	}
	return m, nil
}
func (m model) DatabaseReadControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
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
		m.cursor = 0
		m.page = m.pages[ReadSingle]
		m.controller = m.page.Controller
	}
	return m, nil
}

func (m model) DatabaseReadSingleControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
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
		m.cursor = 0
		m.page = *m.page.Next
		m.controller = m.page.Controller
		m.menu = m.page.Children
	}
	return m, nil
}

func (m model) DatabaseUpdateControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	var rows [][]string
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
		rows = *m.rows
		m.PushHistory(m.page)
		m.row = &rows[m.cursor]
		m.PageRouter()
		m.cursor = 0
		m.page = m.pages[UpdateForm]
		m.controller = m.page.Controller
		m.menu = m.page.Children
	}
	return m, nil
}

func (m model) DatabaseUpdateFormControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	if m.focus == PAGEFOCUS {
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
			if m.cursor < len(m.formActions)-1 {
				m.cursor++
			}
		case "h", "shift+tab", "backspace":
			m.cursor = 0
			m.page = *m.PopHistory()
			m.controller = m.page.Controller
			m.menu = m.page.Children
		case "enter", "l", "right":
			switch m.formActions[m.cursor] {
			case edit:
				m.focus = FORMFOCUS
			case cancel:
				m.cursor = 0
				m.page = *m.PopHistory()
				m.controller = m.page.Controller
				m.menu = m.page.Children
			case reset:
			case submit:
				err := m.CLIUpdate(db.StringDBTable(m.table))
				if err != nil {
					return m, nil
				}
				m.focus = DIALOGFOCUS

			}

		}
	}
	if m.focus == FORMFOCUS {
		switch msg.String() {
		case "ctrl+c":

			return m, tea.Quit
		case "esc", "q":
			return m, tea.Quit
		case "enter", "tab", "right", "down":
			m.form.NextField()
		case "shift+tab", "left", "up":
			m.form.PrevField()
		}

		// Process the form
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			cmds = append(cmds, cmd)
		}

		if m.form.State == huh.StateCompleted {
			// Quit when the form is done.
			cmds = append(cmds, tea.Quit)
		}
	}
	if m.focus == DIALOGFOCUS {
		switch msg.String() {
		case "ctrl+c":

			return m, tea.Quit
		case "esc", "q":
			return m, tea.Quit
		case "enter":
			m.PageRouter()
		}

	}

	return m, tea.Batch(cmds...)
}

func (m model) DatabaseDeleteControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
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
		err := m.CLIDelete(db.StringDBTable(m.table))
		if err != nil {
			return m, nil
		}
		m.cursor = 0
		m.page = m.pages[Database]
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
