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


