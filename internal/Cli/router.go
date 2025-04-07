package cli

import (
	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
)

func (m *model) PageRouter() {
	var err error
	switch m.page.Index {
	case TABLEPAGE:
		switch m.pageMenu[m.cursor] {
		case createPage:
			m.form, m.formLen = m.BuildCreateDBForm(db.StringDBTable(m.table))
			m.headers, m.rows, err = GetColumnsRows(m.table)
			if err != nil {
				m.body = err.Error()
			}
			m.form.Init()
			m.focus = FORMFOCUS
			m.page = m.pages[CREATEPAGE]
			m.controller = m.page.Controller
		case updatePage:
			m.headers, m.rows, err = GetColumnsRows(m.table)
			if err != nil {
				m.body = err.Error()
			}
			m.paginator.PerPage = m.maxRows
			m.paginator.SetTotalPages(len(m.rows))
			m.page = *m.pageMenu[m.cursor]
			m.controller = m.page.Controller
		case readPage:
			m.headers, m.rows, err = GetColumnsRows(m.table)
			if err != nil {
				m.body = err.Error()
			}
			m.paginator.PerPage = m.maxRows
			m.paginator.SetTotalPages(len(m.rows))
			m.page = *m.pageMenu[m.cursor]
			m.controller = m.page.Controller
		case deletePage:
			m.headers, m.rows, err = GetColumnsRows(m.table)
			if err != nil {
				m.body = err.Error()
			}
			m.paginator.PerPage = m.maxRows
			m.paginator.SetTotalPages(len(m.rows))
			m.page = *m.pageMenu[m.cursor]
			m.controller = m.page.Controller
		}
	case UPDATEPAGE:
		m.form, m.formLen = m.BuildUpdateDBForm(db.StringDBTable(m.table))
		if m.form == nil {
			m.page = m.pages[UPDATEPAGE]
			m.controller = m.page.Controller
			m.err = "Form build failed"
		}
		m.form.Init()

		m.headers, m.rows, err = GetColumnsRows(m.table)
		if err != nil {
			m.body = err.Error()
		}
		m.page = m.pages[UPDATEFORMPAGE]
		m.controller = m.page.Controller
	case READPAGE:

		m.page = m.pages[READSINGLEPAGE]
		m.controller = m.page.Controller
	case CONFIGPAGE:
		formatted, err := formatJSON(config.Env)
		if err == nil {
			m.content = formatted
		}
		m.page = *m.pageMenu[m.cursor]
		m.controller = m.page.Controller
	default:
		form, err := formatJSON(config.Env)
		if err == nil {
			m.content = form
		}
		m.viewport.SetContent(m.content)
		m.ready = true
		m.page = *m.pageMenu[m.cursor]
		m.controller = m.page.Controller

	}

	m.pageMenu = m.page.Children
}
