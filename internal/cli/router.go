package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	 "github.com/hegner123/modulacms/internal/db"
)

func (m *Model) PageRouter() tea.Cmd {
	var cmds []tea.Cmd
	switch m.page.Index {
	case TABLEPAGE:
		switch m.pageMenu[m.cursor] {
		case createPage:
			cmd := FetchHeadersRows(m.config, m.table)
			formCmd := m.BuildCreateDBForm(db.StringDBTable(m.table))
			cmds = append(cmds, formCmd)
			cmds = append(cmds, cmd)
			m.form.Init()
			m.focus = FORMFOCUS
			m.page = m.pages[CREATEPAGE]
			m.controller = m.page.Controller
			m.status = EDITING
			return tea.Batch(cmds...)
		case updatePage:
			cmd := FetchHeadersRows(m.config, m.table)
			cmds = append(cmds, cmd)
			m.page = *m.pageMenu[m.cursor]
			m.controller = m.page.Controller
			return tea.Batch(cmds...)
		case readPage:
			cmd := FetchHeadersRows(m.config, m.table)
			m.page = *m.pageMenu[m.cursor]
			m.controller = m.page.Controller
			cmds = append(cmds, cmd)
			return tea.Batch(cmds...)
		case deletePage:
			cmd := FetchHeadersRows(m.config, m.table)
			m.page = *m.pageMenu[m.cursor]
			m.controller = m.page.Controller
			m.status = DELETING
			cmds = append(cmds, cmd)
			return tea.Batch(cmds...)
		}
	case UPDATEPAGE:

		formCmd := m.BuildUpdateDBForm(db.StringDBTable(m.table))
		cmds = append(cmds, formCmd)
		cmd := FetchHeadersRows(m.config, m.table)
		m.page = m.pages[UPDATEFORMPAGE]
		m.controller = m.page.Controller
		m.status = EDITING
		cmds = append(cmds, cmd)
		return tea.Batch(cmds...)
	case READPAGE:

		m.page = m.pages[READSINGLEPAGE]
		m.controller = m.page.Controller
	case CONFIGPAGE:
		formatted, err := formatJSON(m.config)
		if err == nil {
			m.content = formatted
		}
		m.page = *m.pageMenu[m.cursor]
		m.controller = m.page.Controller
	default:
		form, err := formatJSON(m.config)
		if err == nil {
			m.content = form
		}
		m.viewport.SetContent(m.content)
		m.ready = true
		m.page = *m.pageMenu[m.cursor]
		m.controller = m.page.Controller
		m.status = OK
		cmds = append(cmds, GetTablesCMD(m.config))
		return tea.Batch(cmds...)

	}

	m.pageMenu = m.page.Children
	return nil
}
