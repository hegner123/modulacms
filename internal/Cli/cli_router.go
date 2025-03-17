package cli

import (
	"fmt"

	db "github.com/hegner123/modulacms/internal/Db"
)

func (m *model) PageRouter() {
	var err error
	switch m.page.Index {
	case Table:
		switch m.menu[m.cursor] {
		case createPage:
			m.form, m.formLen = m.BuildForm(db.StringDBTable(m.table))
			m.page = *m.menu[m.cursor]
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
	case Update:
		fmt.Println("updateform")
		m.form, m.formLen = m.BuildForm(db.StringDBTable(m.table))
		m.form.Init()
        
		m.headers, m.rows, err = GetColumnsRows(m.table)
		if err != nil {
			m.body = err.Error()
		}
		m.page = m.pages[UpdateForm]
		m.controller = m.page.Controller
	default:
		m.page = *m.menu[m.cursor]
		m.controller = m.page.Controller

	}

	m.cursor = 0
	m.menu = m.page.Children
}
