package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	db "github.com/hegner123/modulacms/internal/db"
)

func (m *Model) CMSCreateDatatypeFormControls(msg tea.Msg) (Model, tea.Cmd) {
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
		createCmd := m.CLICreate(m.config, db.DBTable(m.table))
		cmds = append(cmds, createCmd)
		m.page = m.pages[READPAGE]
		m.controller = m.page.Controller
		cmds = append(cmds, FetchHeadersRows(m.config, m.table))
	}
	return *m, tea.Batch(cmds...)
}

func (m *Model) CMSUpdateDatatypeFormControls(msg tea.Msg) (Model, tea.Cmd) {
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
		createCmd := m.CLICreate(m.config, db.DBTable(m.table))
		cmds = append(cmds, createCmd)
		m.page = m.pages[READPAGE]
		m.controller = m.page.Controller
		cmds = append(cmds, FetchHeadersRows(m.config, m.table))
	}
	return *m, tea.Batch(cmds...)
}
