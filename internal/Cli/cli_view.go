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
	case tablePage.Index:
		ui = m.PageTable()
	case createPage.Index:
		ui = m.PageCreate()
	case readPage.Index:
		ui = m.PageRead()
	case updatePage.Index:
		ui = m.PageUpdate()
	case deletePage.Index:
		ui = m.PageDelete()
	}
	return ui
}
