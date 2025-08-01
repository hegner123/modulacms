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
		if m.TitleFont > 0 {
			m.TitleFont--
		}
	case "shift+right":
		if m.TitleFont < len(m.Titles)-1 {
			m.TitleFont++
		}

	//Navigation
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < option-1 {
			m.Cursor++
		}
	case "h", "shift+tab", "backspace":
		if len(m.History) > 0 {
			entry := *m.PopHistory()
			m.Page = entry.Page
			m.Cursor = entry.Cursor
			m.Controller = m.Page.Controller
			m.PageMenu = m.Page.Children
			return *m, nil
		}

	//Action
	case "enter", "l":
		// Only proceed if we have menu items
		if len(m.PageMenu) > 0 {
			m.PushHistory(PageHistory{Page: m.Page, Cursor: m.Cursor})
			cmd := m.PageRouter()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	default:
		var scmd tea.Cmd
		m.Spinner, scmd = m.Spinner.Update(msg)
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
		if m.TitleFont > 0 {
			m.TitleFont--
		}
	case "shift+right":
		if m.TitleFont < len(m.Titles)-1 {
			m.TitleFont++
		}

	//Navigation
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Tables)-1 {
			m.Cursor++
		}
	case "h", "left", "shift+tab", "backspace":
		if len(m.History) > 0 {
			entry := *m.PopHistory()
			m.Cursor = entry.Cursor
			m.Page = entry.Page
			m.Controller = m.Page.Controller
			m.PageMenu = m.Page.Children
			return *m, nil
		}

	//Action
	case "enter", "l":
		m.PushHistory(PageHistory{Page: m.Page, Cursor: m.Cursor})
		m.Table = m.Tables[m.Cursor]
		m.Cursor = 0
		// Check if Next is nil before dereferencing
		if m.Page.Next != nil {
			m.Page = *m.Page.Next
			m.Controller = m.Page.Controller
			m.PageMenu = m.Page.Children
		} else {
			// If Next is nil, use tableActionsPage instead
			m.Page = *tableActionsPage
			m.Controller = m.Page.Controller
			m.PageMenu = m.Page.Children
		}
	}
	return *m, tea.Batch(cmds...)
}

func (m *Model) DatabaseCreateControls(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.Focus = FORMFOCUS

	logFile, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Printf("Error closing log file: %v\n", err)
		}
	}()

	// Update form with the message
	form, cmd := m.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.Form = f
		cmds = append(cmds, cmd)
	}

	// Handle form state changes
	if m.Form.State == huh.StateAborted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[READPAGE]
		m.Controller = m.Page.Controller
	}

	if m.Form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		err := m.CLICreate(m.Config, db.DBTable(m.Table))
		cmds = append(cmds, err)
		m.Page = m.Pages[READPAGE]
		m.Controller = m.Page.Controller
		cmd := FetchHeadersRows(m.Config, m.Table)
		cmds = append(cmds, cmd)
	}
	var scmd tea.Cmd
	m.Spinner, scmd = m.Spinner.Update(msg)
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
		if m.TitleFont > 0 {
			m.TitleFont--
		}
	case "shift+right":
		if m.TitleFont < len(m.Titles)-1 {
			m.TitleFont++
		}

	//Navigation
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < m.CursorMax-1 {
			m.Cursor++
		}
	case "h", "shift+tab", "backspace":
		entry := *m.PopHistory()
		m.Cursor = entry.Cursor
		m.Page = entry.Page
		m.Controller = m.Page.Controller
		m.PageMenu = m.Page.Children
	case "left":
		if m.PageMod > 0 {
			m.PageMod--
		}
	case "right":
		if m.PageMod < len(m.Rows)/m.MaxRows {
			m.PageMod++
		}

	//Action
	case "enter", "l":
		m.PushHistory(PageHistory{Page: m.Page, Cursor: m.Cursor})

		recordIndex := (m.PageMod * m.MaxRows) + m.Cursor
		if recordIndex < len(m.Rows) {
			m.Cursor = recordIndex
		}
		m.Page = m.Pages[READSINGLEPAGE]
		m.Controller = m.Page.Controller
	default:

		var scmd tea.Cmd
		m.Spinner, scmd = m.Spinner.Update(msg)
		cmds = append(cmds, scmd)
	}
	var pcmd tea.Cmd
	m.Paginator, pcmd = m.Paginator.Update(msg)
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
		if m.TitleFont > 0 {
			m.TitleFont--
		}
	case "shift+right":
		if m.TitleFont < len(m.Titles)-1 {
			m.TitleFont++
		}

	//Navigation
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Rows)-1 {
			m.Cursor++
		}
	case "h", "shift+tab", "backspace":
		if len(m.History) > 0 {
			entry := *m.PopHistory()
			m.Cursor = entry.Cursor
			m.Page = entry.Page
			m.Controller = m.Page.Controller
			m.PageMenu = m.Page.Children
		}

	default:
		// Do nothing for other keys
	}
	return *m, tea.Batch(cmds...)
}

func (m *Model) DatabaseUpdateControls(msg tea.KeyMsg, option int) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var rows [][]string
	m.Focus = FORMFOCUS
	switch msg.String() {
	//Exit
	case "q", "esc", "ctrl+c":
		return *m, tea.Quit
	case "shift+left":
		if m.TitleFont > 0 {
			m.TitleFont--
		}
	case "shift+right":
		if m.TitleFont < len(m.Titles)-1 {
			m.TitleFont++
		}

	//Navigation
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < m.CursorMax-1 {
			m.Cursor++
		}
	case "h", "shift+tab", "backspace":
		if len(m.History) > 0 {
			entry := *m.PopHistory()
			m.Page = entry.Page
			m.Cursor = entry.Cursor
			m.Controller = m.Page.Controller
			m.PageMenu = m.Page.Children
		}
	case "left":
		if m.PageMod > 0 {
			m.PageMod--
		}
	case "right":
		if m.PageMod < len(m.Rows)/m.MaxRows {
			m.PageMod++
		}

	//Action
	case "enter", "l":
		rows = m.Rows
		m.PushHistory(PageHistory{Page: m.Page, Cursor: m.Cursor})
		recordIndex := (m.PageMod * m.MaxRows) + m.Cursor
		// Only update if the calculated index is valid
		if recordIndex < len(m.Rows) {
			m.Cursor = recordIndex
		}
		m.Row = &rows[recordIndex]
		m.PageRouter()
		m.Cursor = 0
		m.Page = m.Pages[UPDATEFORMPAGE]
		m.Controller = m.Page.Controller
		m.PageMenu = m.Page.Children
	default:

		var scmd tea.Cmd
		m.Spinner, scmd = m.Spinner.Update(msg)
		cmds = append(cmds, scmd)
	}
	var pcmd tea.Cmd
	m.Paginator, pcmd = m.Paginator.Update(msg)
	cmds = append(cmds, pcmd)
	return *m, tea.Batch(cmds...)
}

func (m *Model) DatabaseUpdateFormControls(msg tea.Msg, option int) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.Focus = FORMFOCUS

	logFile, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Printf("Error closing log file: %v\n", err)
		}
	}()

	// Update form with the message
	form, cmd := m.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.Form = f
		cmds = append(cmds, cmd)
	}

	// Handle form state changes
	if m.Form.State == huh.StateAborted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[UPDATEPAGE]
		m.Controller = m.Page.Controller
	}

	if m.Form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[UPDATEPAGE]
		m.Controller = m.Page.Controller
		cmd := m.CLIUpdate(m.Config, db.DBTable(m.Table))
		cmds = append(cmds, cmd)
	}
	var scmd tea.Cmd
	m.Spinner, scmd = m.Spinner.Update(msg)
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
		if m.TitleFont > 0 {
			m.TitleFont--
		}
	case "shift+right":
		if m.TitleFont < len(m.Titles)-1 {
			m.TitleFont++
		}

	//Navigation
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < m.CursorMax-1 {
			m.Cursor++
		}
	case "h", "shift+tab", "backspace":
		if len(m.History) > 0 {
			entry := *m.PopHistory()
			m.Cursor = entry.Cursor
			m.Page = entry.Page
			m.Controller = m.Page.Controller
			m.PageMenu = m.Page.Children
		}

	//Action
	case "enter", "l":
		err := m.CLIDelete(m.Config, db.StringDBTable(m.Table))
		if err != nil {
			return *m, nil
		}
		if m.Cursor > 0 {
			m.Cursor--
		}
		cmd := FetchHeadersRows(m.Config, m.Table)
		cmds = append(cmds, cmd)
	default:
		var scmd tea.Cmd
		m.Spinner, scmd = m.Spinner.Update(msg)
		cmds = append(cmds, scmd)
	}
	var pcmd tea.Cmd
	m.Paginator, pcmd = m.Paginator.Update(msg)
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
			if m.TitleFont > 0 {
				m.TitleFont--
			}
		case "shift+right":
			if m.TitleFont < len(m.Titles)-1 {
				m.TitleFont++
			}

		//Navigation
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.PageMenu) {
				m.Cursor++
			}
		case "h", "left", "shift+tab", "backspace":
			if len(m.History) > 0 {
				entry := *m.PopHistory()
				m.Cursor = entry.Cursor
				m.Page = entry.Page
				m.Controller = m.Page.Controller
				m.PageMenu = m.Page.Children
			}

		//Action
		case "enter", "l":
			m.PushHistory(PageHistory{Page: m.Page, Cursor: m.Cursor})
			m.PageRouter()
		default:
			// Only update spinner if we're in a loading state
			if m.Loading {
				var scmd tea.Cmd
				m.Spinner, scmd = m.Spinner.Update(msg)
				cmds = append(cmds, scmd)
			}

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
			if len(m.History) > 0 {
				entry := *m.PopHistory()
				m.Cursor = entry.Cursor
				m.Page = entry.Page
				m.Controller = m.Page.Controller
				m.PageMenu = m.Page.Children
				return *m, nil
			}
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return *m, tea.Batch(cmds...)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
