package cli

import (
	db "github.com/hegner123/modulacms/internal/Db"
)

func (m *model) PageRouter() {
	var err error
	switch m.page.Index {
	case TABLEPAGE:
		switch m.menu[m.cursor] {
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
			m.page = *m.menu[m.cursor]
			m.controller = m.page.Controller
		case readPage:
			m.headers, m.rows, err = GetColumnsRows(m.table)
			if err != nil {
				m.body = err.Error()
			}
			m.page = *m.menu[m.cursor]
			m.controller = m.page.Controller
		case deletePage:
			m.headers, m.rows, err = GetColumnsRows(m.table)
			if err != nil {
				m.body = err.Error()
			}
			m.page = *m.menu[m.cursor]
			m.controller = m.page.Controller
		}
	case UPDATEPAGE:
		m.form, m.formLen = m.BuildUpdateDBForm(db.StringDBTable(m.table))
		m.form.Init()
		m.focus = FORMFOCUS
		m.formFields[0].Focus()

		m.headers, m.rows, err = GetColumnsRows(m.table)
		if err != nil {
			m.body = err.Error()
		}
		m.page = m.pages[UPDATEFORMPAGE]
		m.controller = m.page.Controller
	default:
		m.page = *m.menu[m.cursor]
		m.controller = m.page.Controller

	}

	m.cursor = 0
	m.menu = m.page.Children
}
