package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
)

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
	}

	if m.Form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		err := m.CLICreate(m.Config, db.DBTable(m.Table))
		cmds = append(cmds, err)
		m.Page = m.Pages[READPAGE]
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

