package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

func (m *model) PageControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return m, tea.Quit

	case "shift+left":
		if m.titleFont > 0 {
			m.titleFont--
		}
	case "shift+right":
		if m.titleFont < len(m.titles)-1 {
			m.titleFont++
		}

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

func (m *model) TableSelectControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "shift+left":
		if m.titleFont > 0 {
			m.titleFont--
		}
	case "shift+right":
		if m.titleFont < len(m.titles)-1 {
			m.titleFont++
		}

	//Navigation
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tables)-1 {
			m.cursor++
		}
	case "h", "left", "shift+tab", "backspace":
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

func (m *model) DatabaseCreateControls(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.focus = FORMFOCUS

	logFile, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Update form with the message
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	// Handle form state changes
	if m.form.State == huh.StateAborted {
		_ = tea.ClearScreen()
		m.focus = PAGEFOCUS
		m.page = m.pages[READPAGE]
		m.controller = m.page.Controller
	}

	if m.form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.focus = PAGEFOCUS
		err := m.CLICreate(db.DBTable(m.table))
		if err != nil {
			fmt.Fprintln(logFile, err.Error())
		}
		m.page = m.pages[READPAGE]
		m.controller = m.page.Controller
		m.headers, m.rows, _ = GetColumnsRows(m.table)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) DatabaseReadControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "shift+left":
		if m.titleFont > 0 {
			m.titleFont--
		}
	case "shift+right":
		if m.titleFont < len(m.titles)-1 {
			m.titleFont++
		}

	//Navigation
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < option-1 {
			m.cursor++
		}
	case "h", "left", "shift+tab", "backspace":
		m.cursor = 0
		m.page = *m.PopHistory()
		m.controller = m.page.Controller
		m.menu = m.page.Children

	//Action
	case "enter", "l":
		m.PushHistory(m.page)
		m.page = m.pages[READSINGLEPAGE]
		m.controller = m.page.Controller
	}
	return m, nil
}

func (m *model) DatabaseReadSingleControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "shift+left":
		if m.titleFont > 0 {
			m.titleFont--
		}
	case "shift+right":
		if m.titleFont < len(m.titles)-1 {
			m.titleFont++
		}

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

func (m *model) DatabaseUpdateControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	var rows [][]string
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "shift+left":
		if m.titleFont > 0 {
			m.titleFont--
		}
	case "shift+right":
		if m.titleFont < len(m.titles)-1 {
			m.titleFont++
		}

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
		rows = m.rows
		m.PushHistory(m.page)
		m.row = &rows[m.cursor]
		m.PageRouter()
		m.cursor = 0
		m.page = m.pages[UPDATEFORMPAGE]
		m.controller = m.page.Controller
		m.menu = m.page.Children
	}
	return m, nil
}

func (m *model) DatabaseUpdateFormControls(msg tea.Msg, option int) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.focus = FORMFOCUS

	logFile, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Update form with the message
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	// Handle form state changes
	if m.form.State == huh.StateAborted {
		_ = tea.ClearScreen()
		m.focus = PAGEFOCUS
		m.page = m.pages[UPDATEPAGE]
		m.controller = m.page.Controller
	}

	if m.form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.focus = PAGEFOCUS
		m.page = m.pages[UPDATEPAGE]
		m.controller = m.page.Controller
		err := m.CLIUpdate(db.DBTable(m.table))
		if err != nil {
			utility.DefaultLogger.Ferror(logFile, "", err)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) DatabaseDeleteControls(msg tea.KeyMsg, option int) (tea.Model, tea.Cmd) {
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "shift+left":
		if m.titleFont > 0 {
			m.titleFont--
		}
	case "shift+right":
		if m.titleFont < len(m.titles)-1 {
			m.titleFont++
		}

	//Navigation
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.rows)-1 {
			m.cursor++
		}
	case "h", "left", "shift+tab", "backspace":
		m.cursor = 0
		m.page = *m.PopHistory()
		m.controller = m.page.Controller
		m.menu = m.page.Children

	//Action
	case "enter", "l":
		err := m.CLIDelete(db.StringDBTable(m.table))
		if err != nil {
			return m, nil
		}
		m.headers, m.rows, _ = GetColumnsRows(m.table)
	}
	return m, nil
}

func (m *model) ContentControls(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		//Exit
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "shift+left":
			if m.titleFont > 0 {
				m.titleFont--
			}
		case "shift+right":
			if m.titleFont < len(m.titles)-1 {
				m.titleFont++
			}

		//Navigation
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menu) {
				m.cursor++
			}
		case "h", "left", "shift+tab", "backspace":
			m.cursor = 0
			m.page = *m.PopHistory()
			m.controller = m.page.Controller
			m.menu = m.page.Children

		//Action
		case "enter", "l":
			m.PushHistory(m.page)
			m.PageRouter()
		}
	}
	return m, nil
}
