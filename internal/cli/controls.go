package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
)

func (m *Model) PageControls(msg tea.KeyMsg, option int) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return *m, tea.Quit

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
		if len(m.history) > 0 {
			entry := *m.PopHistory()
			m.page = entry.Page
			m.cursor = entry.Cursor
			m.controller = m.page.Controller
			m.pageMenu = m.page.Children
			return *m, nil
		}

	//Action
	case "enter", "l":
		m.PushHistory(PageHistory{Page: m.page, Cursor: m.cursor})
		cmd := m.PageRouter()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	default:
		var scmd tea.Cmd
		m.spinner, scmd = m.spinner.Update(msg)
		cmds = append(cmds, scmd)
	}
	cmds = append(cmds, tea.WindowSize())
	return *m, tea.Batch(cmds...)

}

func (m *Model) TableSelectControls(msg tea.KeyMsg, option int) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return *m, tea.Quit
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
		if len(m.history) > 0 {
			entry := *m.PopHistory()
			m.cursor = entry.Cursor
			m.page = entry.Page
			m.controller = m.page.Controller
			m.pageMenu = m.page.Children
			return *m, nil
		}

	//Action
	case "enter", "l":
		m.PushHistory(PageHistory{Page: m.page, Cursor: m.cursor})
		m.table = m.tables[m.cursor]
		m.cursor = 0
		m.page = *m.page.Next
		m.controller = m.page.Controller
		m.pageMenu = m.page.Children
	}
	return *m, tea.Batch(cmds...)
}

func (m *Model) DatabaseCreateControls(msg tea.Msg) (Model, tea.Cmd) {
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
		err := m.CLICreate(m.config, db.DBTable(m.table))
		cmds = append(cmds, err)
		m.page = m.pages[READPAGE]
		m.controller = m.page.Controller
		cmd := FetchHeadersRows(m.config, m.table)
		cmds = append(cmds, cmd)
	}
	var scmd tea.Cmd
	m.spinner, scmd = m.spinner.Update(msg)
	cmds = append(cmds, scmd)

	return *m, tea.Batch(cmds...)
}

func (m *Model) DatabaseReadControls(msg tea.KeyMsg, option int) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return *m, tea.Quit
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
		if m.cursor < m.cursorMax-1 {
			m.cursor++
		}
	case "h", "shift+tab", "backspace":
		entry := *m.PopHistory()
		m.cursor = entry.Cursor
		m.page = entry.Page
		m.controller = m.page.Controller
		m.pageMenu = m.page.Children
	case "left":
		if m.pageMod > 0 {
			m.pageMod--
		}
	case "right":
		if m.pageMod < len(m.rows)/m.maxRows {
			m.pageMod++
		}

	//Action
	case "enter", "l":
		m.PushHistory(PageHistory{Page: m.page, Cursor: m.cursor})

		recordIndex := (m.pageMod * m.maxRows) + m.cursor
		if recordIndex < len(m.rows) {
			m.cursor = recordIndex
		}
		m.page = m.pages[READSINGLEPAGE]
		m.controller = m.page.Controller
	default:

		var scmd tea.Cmd
		m.spinner, scmd = m.spinner.Update(msg)
		cmds = append(cmds, scmd)
	}
	var pcmd tea.Cmd
	m.paginator, pcmd = m.paginator.Update(msg)
	cmds = append(cmds, pcmd)
	cmd := m.UpdateMaxCursor()
	cmds = append(cmds, cmd)
	return *m, tea.Batch(cmds...)
}

func (m *Model) DatabaseReadSingleControls(msg tea.KeyMsg, option int) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return *m, tea.Quit
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
	case "h", "shift+tab", "backspace":
		if len(m.history) > 0 {
			entry := *m.PopHistory()
			m.cursor = entry.Cursor
			m.page = entry.Page
			m.controller = m.page.Controller
			m.pageMenu = m.page.Children
		}

	default:

		var scmd tea.Cmd
		m.spinner, scmd = m.spinner.Update(msg)
		cmds = append(cmds, scmd)

	}
	return *m, tea.Batch(cmds...)
}

func (m *Model) DatabaseUpdateControls(msg tea.KeyMsg, option int) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var rows [][]string
	m.focus = FORMFOCUS
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return *m, tea.Quit
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
		if m.cursor < m.cursorMax-1 {
			m.cursor++
		}
	case "h", "shift+tab", "backspace":
		if len(m.history) > 0 {
			entry := *m.PopHistory()
			m.page = entry.Page
			m.cursor = entry.Cursor
			m.controller = m.page.Controller
			m.pageMenu = m.page.Children
		}
	case "left":
		if m.pageMod > 0 {
			m.pageMod--
		}
	case "right":
		if m.pageMod < len(m.rows)/m.maxRows {
			m.pageMod++
		}

	//Action
	case "enter", "l":
		rows = m.rows
		m.PushHistory(PageHistory{Page: m.page, Cursor: m.cursor})
		recordIndex := (m.pageMod * m.maxRows) + m.cursor
		// Only update if the calculated index is valid
		if recordIndex < len(m.rows) {
			m.cursor = recordIndex
		}
		m.row = &rows[recordIndex]
		m.PageRouter()
		m.cursor = 0
		m.page = m.pages[UPDATEFORMPAGE]
		m.controller = m.page.Controller
		m.pageMenu = m.page.Children
	default:

		var scmd tea.Cmd
		m.spinner, scmd = m.spinner.Update(msg)
		cmds = append(cmds, scmd)
	}
	var pcmd tea.Cmd
	m.paginator, pcmd = m.paginator.Update(msg)
	cmds = append(cmds, pcmd)
	return *m, tea.Batch(cmds...)
}

func (m *Model) DatabaseUpdateFormControls(msg tea.Msg, option int) (Model, tea.Cmd) {
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
		cmd := m.CLIUpdate(m.config, db.DBTable(m.table))
		cmds = append(cmds, cmd)
	}
	var scmd tea.Cmd
	m.spinner, scmd = m.spinner.Update(msg)
	cmds = append(cmds, scmd)

	return *m, tea.Batch(cmds...)
}

func (m *Model) DatabaseDeleteControls(msg tea.KeyMsg, option int) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return *m, tea.Quit
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
		if m.cursor < m.cursorMax-1 {
			m.cursor++
		}
	case "h", "shift+tab", "backspace":
		if len(m.history) > 0 {
			entry := *m.PopHistory()
			m.cursor = entry.Cursor
			m.page = entry.Page
			m.controller = m.page.Controller
			m.pageMenu = m.page.Children
		}

	//Action
	case "enter", "l":
		err := m.CLIDelete(m.config, db.StringDBTable(m.table))
		if err != nil {
			return *m, nil
		}
		if m.cursor > 0 {
			m.cursor--
		}
		cmd := FetchHeadersRows(m.config, m.table)
		cmds = append(cmds, cmd)
	default:
		var scmd tea.Cmd
		m.spinner, scmd = m.spinner.Update(msg)
		cmds = append(cmds, scmd)
	}
	var pcmd tea.Cmd
	m.paginator, pcmd = m.paginator.Update(msg)
	cmds = append(cmds, pcmd)
	return *m, tea.Batch(cmds...)
}

func (m *Model) ContentControls(message tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		//Exit
		case "q", "esc", "ctrl+c":
			return *m, tea.Quit
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
			if m.cursor < len(m.pageMenu) {
				m.cursor++
			}
		case "h", "left", "shift+tab", "backspace":
			if len(m.history) > 0 {
				entry := *m.PopHistory()
				m.cursor = entry.Cursor
				m.page = entry.Page
				m.controller = m.page.Controller
				m.pageMenu = m.page.Children
			}

		//Action
		case "enter", "l":
			m.PushHistory(PageHistory{Page: m.page, Cursor: m.cursor})
			m.PageRouter()
		default:
			var scmd tea.Cmd
			m.spinner, scmd = m.spinner.Update(msg)
			cmds = append(cmds, scmd)

		}
	}
	return *m, tea.Batch(cmds...)
}

func (m *Model) ConfigControls(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Exit
		case "q", "esc", "ctrl+c":
			return *m, tea.Quit
		case "h", "shift+tab", "backspace":
			if len(m.history) > 0 {
				entry := *m.PopHistory()
				m.cursor = entry.Cursor
				m.page = entry.Page
				m.controller = m.page.Controller
				m.pageMenu = m.page.Children
				return *m, nil
			}
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return *m, tea.Batch(cmds...)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
