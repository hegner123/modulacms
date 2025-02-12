package cli

func (m model) View() string {
	m.header = m.StatusTable()
	var ui string
	switch m.page.Index {
	case homePage.Index:
		ui = m.PageHome()
	case databasePage.Index:
		ui = m.PageDatabase()
	case cmsPage.Index:
		ui = m.PageCMS()
	case createPage.Index:
		if m.table == "" {
			ui = m.SelectTableUI("create row")
		} else {
			ui = m.PageCreate()
		}
	case readPage.Index:
		if m.table == "" {
			ui = m.SelectTableUI("read")
		} else {
			ui = m.PageRead()
		}
	case updatePage.Index:
		if m.table == "" {
			ui = m.SelectTableUI("update row")
		} else {
			ui = m.PageUpdate()
		}
	case deletePage.Index:
		if m.table == "" {
			ui = m.SelectTableUI("delete row")
		} else {
			ui = m.PageDelete()
		}
	}
	return ui
}
