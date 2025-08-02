package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
)


func (m *Model) PageRouter() tea.Cmd {
	var cmds []tea.Cmd

	// Safety check to ensure PageMenu is not empty and cursor is within bounds
	if len(m.PageMenu) == 0 {
		// No menu items, return nil
		return nil
	}

	// Ensure cursor is within bounds
	if m.Cursor >= len(m.PageMenu) {
		m.Cursor = len(m.PageMenu) - 1
	}

	switch m.Page.Index {
	case TABLEPAGE:
		switch m.PageMenu[m.Cursor] {
		case createPage:
			cmd := FetchHeadersRows(m.Config, m.Table)
			formCmd := m.BuildCreateDBForm(db.StringDBTable(m.Table))
			cmds = append(cmds, formCmd)
			cmds = append(cmds, cmd)
			m.Form.Init()
			m.Focus = FORMFOCUS
			m.Page = m.Pages[CREATEPAGE]
			m.Status = EDITING
			return tea.Batch(cmds...)
		case updatePage:
			cmd := FetchHeadersRows(m.Config, m.Table)
			cmds = append(cmds, cmd)
			m.Page = *m.PageMenu[m.Cursor]
			return tea.Batch(cmds...)
		case readPage:
			cmd := FetchHeadersRows(m.Config, m.Table)
			m.Page = *m.PageMenu[m.Cursor]
			cmds = append(cmds, cmd)
			return tea.Batch(cmds...)
		case deletePage:
			cmd := FetchHeadersRows(m.Config, m.Table)
			m.Page = *m.PageMenu[m.Cursor]
			m.Status = DELETING
			cmds = append(cmds, cmd)
			return tea.Batch(cmds...)
		}
	case UPDATEPAGE:

		formCmd := m.BuildUpdateDBForm(db.StringDBTable(m.Table))
		cmds = append(cmds, formCmd)
		cmd := FetchHeadersRows(m.Config, m.Table)
		m.Page = m.Pages[UPDATEFORMPAGE]
		m.Status = EDITING
		cmds = append(cmds, cmd)
		return tea.Batch(cmds...)
	case READPAGE:

		m.Page = m.Pages[READSINGLEPAGE]
	case CONFIGPAGE:
		formatted, err := formatJSON(m.Config)
		if err == nil {
			m.Content = formatted
		}
		if len(m.PageMenu) > 0 && m.Cursor < len(m.PageMenu) {
			m.Page = *m.PageMenu[m.Cursor]
		}
	case CMSPAGE:
	default:
		form, err := formatJSON(m.Config)
		if err == nil {
			m.Content = form
		}
		m.Viewport.SetContent(m.Content)
		m.Ready = true

		// Check if PageMenu has elements and cursor is within bounds
		if len(m.PageMenu) > 0 && m.Cursor < len(m.PageMenu) {
			m.Page = *m.PageMenu[m.Cursor]
			r := m.Page.PageInit(*m)
			cmds = append(cmds, r)
		}

		m.Status = OK
		return tea.Batch(cmds...)

	}

	m.PageMenu = m.Page.Children
	return nil
}
