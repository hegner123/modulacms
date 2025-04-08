package cli

import (

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	db "github.com/hegner123/modulacms/internal/db"
	utility "github.com/hegner123/modulacms/internal/utility"
)

func (m *model) CMSCreateDatatypeFormControls(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.focus = FORMFOCUS

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

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
			utility.DefaultLogger.Ferror("", err)
		}
		m.page = m.pages[READPAGE]
		m.controller = m.page.Controller
		m.headers, m.rows, _ = GetColumnsRows(m.table)
	}
	return m, tea.Batch(cmds...)
}

func (m *model) CMSUpdateDatatypeFormControls(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.focus = FORMFOCUS

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

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
			utility.DefaultLogger.Ferror("", err)
		}
		m.page = m.pages[READPAGE]
		m.controller = m.page.Controller
		m.headers, m.rows, _ = GetColumnsRows(m.table)
	}
	return m, tea.Batch(cmds...)
}
